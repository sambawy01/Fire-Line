package customer

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/opsnerve/fireline/internal/event"
	"github.com/opsnerve/fireline/internal/tenant"
	"github.com/opsnerve/fireline/pkg/database"
)

// PredictChurn estimates churn risk for a guest given their historical visit
// dates (sorted oldest to newest). If fewer than 3 visits are available the
// function returns a conservative low-risk estimate.
//
// Risk tiers:
//
//	days_overdue <= 0    → "low"
//	days_overdue <= 0.5x → "medium"
//	days_overdue <= 1.5x → "high"
//	days_overdue >  1.5x → "critical"
func PredictChurn(visitDates []time.Time) (risk string, probability float64) {
	if len(visitDates) < 3 {
		return "low", 0.1
	}

	// Sort ascending.
	sorted := make([]time.Time, len(visitDates))
	copy(sorted, visitDates)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Before(sorted[j]) })

	// Average inter-visit interval in days.
	var totalGap float64
	for i := 1; i < len(sorted); i++ {
		totalGap += sorted[i].Sub(sorted[i-1]).Hours() / 24
	}
	avgInterval := totalGap / float64(len(sorted)-1)
	if avgInterval < 1 {
		avgInterval = 1
	}

	daysSinceLast := time.Since(sorted[len(sorted)-1]).Hours() / 24
	daysOverdue := daysSinceLast - avgInterval

	switch {
	case daysOverdue <= 0:
		return "low", 0.05
	case daysOverdue <= avgInterval*0.5:
		return "medium", 0.25
	case daysOverdue <= avgInterval*1.5:
		return "high", 0.60
	default:
		return "critical", 0.90
	}
}

// RunChurnPrediction loads visit histories for all guests in the org, runs
// PredictChurn, writes results back, and emits bus events for any high-CLV
// guest entering a high or critical tier.
func (s *Service) RunChurnPrediction(ctx context.Context, orgID string) (int, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	// Load all guest visit dates grouped by guest.
	type guestVisitRow struct {
		guestID   string
		clvScore  float64
		visitedAt time.Time
	}

	var rows []guestVisitRow

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		r, err := tx.Query(tenantCtx,
			`SELECT gp.guest_id, gp.clv_score::FLOAT8, gv.visited_at
			 FROM guest_profiles gp
			 JOIN guest_visits gv ON gv.guest_id = gp.guest_id
			 WHERE gp.org_id = $1
			 ORDER BY gp.guest_id, gv.visited_at ASC`,
			orgID,
		)
		if err != nil {
			return fmt.Errorf("query churn visits: %w", err)
		}
		defer r.Close()

		for r.Next() {
			var row guestVisitRow
			if err := r.Scan(&row.guestID, &row.clvScore, &row.visitedAt); err != nil {
				return fmt.Errorf("scan churn row: %w", err)
			}
			rows = append(rows, row)
		}
		return r.Err()
	})
	if err != nil {
		return 0, err
	}

	// Group visit dates by guest.
	type guestData struct {
		clvScore float64
		dates    []time.Time
	}
	guestMap := make(map[string]*guestData)
	for _, row := range rows {
		if guestMap[row.guestID] == nil {
			guestMap[row.guestID] = &guestData{clvScore: row.clvScore}
		}
		guestMap[row.guestID].dates = append(guestMap[row.guestID].dates, row.visitedAt)
	}

	// Compute churn predictions for all guests before opening the write tx.
	type churnResult struct {
		guestID string
		risk    string
		prob    float64
		clv     float64
	}
	var results []churnResult
	for guestID, data := range guestMap {
		risk, prob := PredictChurn(data.dates)
		results = append(results, churnResult{
			guestID: guestID,
			risk:    risk,
			prob:    prob,
			clv:     data.clvScore,
		})
	}

	// Batch all updates into a single transaction to avoid N+1 TenantTx calls.
	updated := 0
	err = database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		for _, r := range results {
			_, execErr := tx.Exec(tenantCtx,
				`UPDATE guest_profiles
				 SET churn_risk = $1, churn_probability = $2, updated_at = now()
				 WHERE guest_id = $3`,
				r.risk, r.prob, r.guestID,
			)
			if execErr != nil {
				return fmt.Errorf("update guest %s churn: %w", r.guestID, execErr)
			}
			updated++
		}
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("batch churn update: %w", err)
	}

	// Emit alert events for high-CLV guests entering high or critical churn risk.
	for _, r := range results {
		if r.clv >= 200 && (r.risk == "high" || r.risk == "critical") {
			s.bus.Publish(ctx, event.Envelope{
				EventType: "customer.churn_alert",
				OrgID:     orgID,
				Source:    "customer.churn",
				Payload: map[string]any{
					"guest_id":   r.guestID,
					"churn_risk": r.risk,
					"clv_score":  r.clv,
				},
			})
		}
	}

	return updated, nil
}
