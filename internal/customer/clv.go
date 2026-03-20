package customer

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/opsnerve/fireline/internal/tenant"
	"github.com/opsnerve/fireline/pkg/database"
)

// CalculateCLV computes customer lifetime value in dollars using a simplified
// RFM-weighted lifespan model. avgCheck is in cents.
func CalculateCLV(avgCheck int64, totalVisits int, firstVisitAt time.Time, churnRisk string) float64 {
	if totalVisits == 0 {
		return 0
	}

	monthsActive := int(time.Since(firstVisitAt).Hours() / 730)
	if monthsActive < 1 {
		monthsActive = 1
	}

	visitFreq := float64(totalVisits) / float64(monthsActive)

	lifespan := 24.0 // default months
	switch churnRisk {
	case "critical":
		lifespan = 3
	case "high":
		lifespan = 9
	case "medium":
		lifespan = 18
	}

	return float64(avgCheck) / 100.0 * visitFreq * lifespan * 0.65
}

// RecalculateAllCLV performs a batch CLV recalculation for every guest in the
// org, reading first-visit date from guest_visits and churn_risk from the
// profile, then writing the updated clv_score back.
func (s *Service) RecalculateAllCLV(ctx context.Context, orgID string) (int, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	type guestRow struct {
		guestID     string
		avgCheck    int64
		totalVisits int
		firstVisit  time.Time
		churnRisk   string
	}

	var guests []guestRow

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(tenantCtx,
			`SELECT gp.guest_id, gp.avg_check, gp.total_visits,
			        COALESCE(MIN(gv.visited_at), gp.created_at) AS first_visit,
			        COALESCE(gp.churn_risk, 'low') AS churn_risk
			 FROM guest_profiles gp
			 LEFT JOIN guest_visits gv ON gv.guest_id = gp.guest_id
			 WHERE gp.org_id = $1
			 GROUP BY gp.guest_id`,
			orgID,
		)
		if err != nil {
			return fmt.Errorf("query guests for CLV: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var g guestRow
			if err := rows.Scan(&g.guestID, &g.avgCheck, &g.totalVisits, &g.firstVisit, &g.churnRisk); err != nil {
				return fmt.Errorf("scan guest: %w", err)
			}
			guests = append(guests, g)
		}
		return rows.Err()
	})
	if err != nil {
		return 0, err
	}

	updated := 0
	for _, g := range guests {
		clv := CalculateCLV(g.avgCheck, g.totalVisits, g.firstVisit, g.churnRisk)

		err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
			_, err := tx.Exec(tenantCtx,
				`UPDATE guest_profiles SET clv_score = $1, updated_at = now() WHERE guest_id = $2`,
				clv, g.guestID,
			)
			return err
		})
		if err != nil {
			continue
		}
		updated++
	}

	return updated, nil
}
