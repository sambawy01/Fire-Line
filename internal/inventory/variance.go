package inventory

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/opsnerve/fireline/internal/event"
	"github.com/opsnerve/fireline/internal/tenant"
	"github.com/opsnerve/fireline/pkg/database"
)

// CountVariance represents the analyzed variance for a single ingredient from a count.
type CountVariance struct {
	VarianceID         string             `json:"variance_id"`
	OrgID              string             `json:"org_id"`
	LocationID         string             `json:"location_id"`
	IngredientID       string             `json:"ingredient_id"`
	Name               string             `json:"name"`
	CountID            string             `json:"count_id"`
	PeriodStart        time.Time          `json:"period_start"`
	PeriodEnd          time.Time          `json:"period_end"`
	TheoreticalUsage   float64            `json:"theoretical_usage"`
	ActualUsage        float64            `json:"actual_usage"`
	VarianceQty        float64            `json:"variance_qty"`
	VarianceCents      int                `json:"variance_cents"`
	CauseProbabilities map[string]float64 `json:"cause_probabilities"`
	Severity           string             `json:"severity"`
	CreatedAt          time.Time          `json:"created_at"`
}

// VarianceSignals holds the input signals used by the categorization engine.
type VarianceSignals struct {
	VarianceQty      float64 // actual - theoretical (negative = shortage, positive = surplus)
	TheoreticalUsage float64
	LoggedWasteQty   float64 // total waste logged for this ingredient in the period
	PortioningFlag   bool    // true if this ingredient appears frequently in high-waste checks
}

// categorizeCauses assigns probability scores to possible variance causes.
// Returns a map of cause -> probability (0.0–1.0). Values do not need to sum to 1.
func categorizeCauses(sig VarianceSignals) map[string]float64 {
	causes := map[string]float64{}

	if sig.TheoreticalUsage == 0 {
		causes["unknown"] = 1.0
		return causes
	}

	pct := sig.VarianceQty / sig.TheoreticalUsage // negative = shortage

	// Unrecorded waste: shortage that is largely explained by logged waste
	if sig.VarianceQty < 0 && sig.LoggedWasteQty > 0 {
		wasteExplains := math.Min(math.Abs(sig.VarianceQty), sig.LoggedWasteQty) / math.Abs(sig.VarianceQty)
		causes["unrecorded_waste"] = wasteExplains
	}

	// Portioning: shortage with portioning flag suggests over-portioning
	if sig.VarianceQty < 0 && sig.PortioningFlag {
		causes["over_portioning"] = math.Min(0.8, math.Abs(pct))
	}

	// Theft / unrecorded usage: large shortage with no waste logged
	if sig.VarianceQty < 0 && sig.LoggedWasteQty == 0 && math.Abs(pct) > 0.10 {
		causes["theft_or_unrecorded"] = math.Min(0.9, math.Abs(pct))
	}

	// Measurement error: small variance either direction
	if math.Abs(pct) <= 0.05 {
		causes["measurement_error"] = 0.7
	}

	// Surplus: over-reporting or receiving discrepancy
	if sig.VarianceQty > 0 {
		causes["receiving_discrepancy"] = math.Min(0.6, pct)
	}

	if len(causes) == 0 {
		causes["unknown"] = 1.0
	}

	return causes
}

// normalize converts a raw quantity variance to a percentage of theoretical usage.
func normalize(varianceQty, theoreticalUsage float64) float64 {
	if theoreticalUsage == 0 {
		return 0
	}
	return varianceQty / theoreticalUsage
}

// classifySeverity assigns a severity level based on the variance percentage.
// info: |pct| < 5%, warning: 5–15%, critical: >15%
func classifySeverity(variancePct float64) string {
	abs := math.Abs(variancePct)
	switch {
	case abs >= 0.15:
		return "critical"
	case abs >= 0.05:
		return "warning"
	default:
		return "info"
	}
}

// CalculateCountVariances computes and persists variance records for a submitted/approved count.
// It compares counted quantities against theoretical usage derived from the recipe explosion.
func (s *Service) CalculateCountVariances(ctx context.Context, orgID, locationID, countID string, periodStart, periodEnd time.Time) ([]CountVariance, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var variances []CountVariance

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		// Pull count lines that have been counted
		rows, err := tx.Query(tenantCtx,
			`SELECT cl.ingredient_id, i.name, i.cost_per_unit,
			        COALESCE(cl.counted_qty, 0) AS counted_qty,
			        COALESCE(cl.expected_qty, 0) AS expected_qty,
			        cl.unit
			 FROM inventory_count_lines cl
			 JOIN ingredients i ON i.ingredient_id = cl.ingredient_id
			 WHERE cl.count_id = $1 AND cl.counted_qty IS NOT NULL`,
			countID,
		)
		if err != nil {
			return fmt.Errorf("query count lines: %w", err)
		}
		defer rows.Close()

		type lineData struct {
			ingredientID string
			name         string
			costPerUnit  int64
			countedQty   float64
			expectedQty  float64
			unit         string
		}

		var lines []lineData
		for rows.Next() {
			var ld lineData
			if err := rows.Scan(&ld.ingredientID, &ld.name, &ld.costPerUnit,
				&ld.countedQty, &ld.expectedQty, &ld.unit); err != nil {
				return fmt.Errorf("scan line: %w", err)
			}
			lines = append(lines, ld)
		}
		if err := rows.Err(); err != nil {
			return err
		}

		for _, ld := range lines {
			// Theoretical usage from recipe explosion over the period
			var theoreticalUsage float64
			err := tx.QueryRow(tenantCtx,
				`SELECT COALESCE(SUM(ci.quantity * re.quantity_per_unit), 0)
				 FROM check_items ci
				 JOIN checks c ON c.check_id = ci.check_id
				 JOIN recipe_explosion re ON re.menu_item_id = ci.menu_item_id
				 WHERE c.location_id = $1
				   AND c.closed_at >= $2 AND c.closed_at < $3
				   AND c.status = 'closed'
				   AND ci.voided_at IS NULL
				   AND re.ingredient_id = $4`,
				locationID, periodStart, periodEnd, ld.ingredientID,
			).Scan(&theoreticalUsage)
			if err != nil {
				return fmt.Errorf("query theoretical usage for %s: %w", ld.ingredientID, err)
			}

			// Logged waste over the period
			var loggedWaste float64
			err = tx.QueryRow(tenantCtx,
				`SELECT COALESCE(SUM(quantity), 0)
				 FROM waste_logs
				 WHERE location_id = $1
				   AND ingredient_id = $2
				   AND logged_at >= $3 AND logged_at < $4`,
				locationID, ld.ingredientID, periodStart, periodEnd,
			).Scan(&loggedWaste)
			if err != nil {
				return fmt.Errorf("query waste for %s: %w", ld.ingredientID, err)
			}

			// actual usage = expected (opening) - counted (closing) + theoretical adjustments
			// Simpler: variance = counted_qty - expected_qty (positive means we have more than expected)
			// For the purposes of this engine: varianceQty = counted - expected
			varianceQty := ld.countedQty - ld.expectedQty
			variancePct := normalize(varianceQty, theoreticalUsage)
			severity := classifySeverity(variancePct)

			sig := VarianceSignals{
				VarianceQty:      varianceQty,
				TheoreticalUsage: theoreticalUsage,
				LoggedWasteQty:   loggedWaste,
			}
			causes := categorizeCauses(sig)

			varianceCents := int(varianceQty * float64(ld.costPerUnit))

			var cv CountVariance
			err = tx.QueryRow(tenantCtx,
				`INSERT INTO inventory_variances
				   (org_id, location_id, ingredient_id, count_id, period_start, period_end,
				    theoretical_usage, actual_usage, variance_qty, variance_cents,
				    cause_probabilities, severity)
				 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
				 RETURNING variance_id, created_at`,
				orgID, locationID, ld.ingredientID, countID,
				periodStart, periodEnd,
				theoreticalUsage, ld.countedQty, varianceQty, varianceCents,
				causes, severity,
			).Scan(&cv.VarianceID, &cv.CreatedAt)
			if err != nil {
				return fmt.Errorf("insert variance for %s: %w", ld.ingredientID, err)
			}

			cv.OrgID = orgID
			cv.LocationID = locationID
			cv.IngredientID = ld.ingredientID
			cv.Name = ld.name
			cv.CountID = countID
			cv.PeriodStart = periodStart
			cv.PeriodEnd = periodEnd
			cv.TheoreticalUsage = theoreticalUsage
			cv.ActualUsage = ld.countedQty
			cv.VarianceQty = varianceQty
			cv.VarianceCents = varianceCents
			cv.CauseProbabilities = causes
			cv.Severity = severity

			variances = append(variances, cv)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Publish event so downstream systems can react
	s.bus.Publish(ctx, event.Envelope{
		EventID:    fmt.Sprintf("%s.variances.calculated", countID),
		EventType:  "inventory.variances.calculated",
		OrgID:      orgID,
		LocationID: locationID,
		Source:     "inventory",
		Payload: map[string]any{
			"count_id":       countID,
			"variance_count": len(variances),
			"period_start":   periodStart,
			"period_end":     periodEnd,
		},
	})

	return variances, nil
}

// ListVariances returns stored variance records for a location within a time range.
func (s *Service) ListVariances(ctx context.Context, orgID, locationID string, from, to time.Time) ([]CountVariance, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var results []CountVariance

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(tenantCtx,
			`SELECT iv.variance_id, iv.org_id, iv.location_id, iv.ingredient_id, i.name,
			        iv.count_id, iv.period_start, iv.period_end,
			        iv.theoretical_usage, iv.actual_usage, iv.variance_qty, iv.variance_cents,
			        iv.cause_probabilities, iv.severity, iv.created_at
			 FROM inventory_variances iv
			 JOIN ingredients i ON i.ingredient_id = iv.ingredient_id
			 WHERE iv.location_id = $1
			   AND iv.created_at >= $2 AND iv.created_at < $3
			 ORDER BY iv.created_at DESC`,
			locationID, from, to,
		)
		if err != nil {
			return fmt.Errorf("query variances: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var cv CountVariance
			if err := rows.Scan(
				&cv.VarianceID, &cv.OrgID, &cv.LocationID, &cv.IngredientID, &cv.Name,
				&cv.CountID, &cv.PeriodStart, &cv.PeriodEnd,
				&cv.TheoreticalUsage, &cv.ActualUsage, &cv.VarianceQty, &cv.VarianceCents,
				&cv.CauseProbabilities, &cv.Severity, &cv.CreatedAt,
			); err != nil {
				return fmt.Errorf("scan variance: %w", err)
			}
			results = append(results, cv)
		}
		return rows.Err()
	})

	return results, err
}
