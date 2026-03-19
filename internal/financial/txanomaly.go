package financial

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/opsnerve/fireline/internal/event"
	"github.com/opsnerve/fireline/internal/tenant"
	"github.com/opsnerve/fireline/pkg/database"
)

// TransactionAnomaly represents a detected anomaly at the transaction level.
type TransactionAnomaly struct {
	Type         string    `json:"type"`
	Description  string    `json:"description"`
	CurrentValue float64   `json:"current_value"`
	Baseline     float64   `json:"baseline"`
	ZScore       float64   `json:"z_score"`
	Severity     string    `json:"severity"`
	DetectedAt   time.Time `json:"detected_at"`
}

// DetectTransactionAnomalies checks void count, discount total, off-hours count,
// and discount rate against 30-day baselines using Z-score analysis.
func (s *Service) DetectTransactionAnomalies(ctx context.Context, orgID, locationID string, day time.Time) ([]TransactionAnomaly, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var anomalies []TransactionAnomaly

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		dayStart := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location())
		dayEnd := dayStart.AddDate(0, 0, 1)
		baselineStart := dayStart.AddDate(0, 0, -30)

		// --- Build 30-day daily baselines ---
		baselineRows, err := tx.Query(tenantCtx,
			`SELECT
				DATE(c.closed_at) AS day,
				COUNT(ci.voided_at)::float                                   AS void_count,
				COALESCE(SUM(c.discount), 0)::float                          AS discount_total,
				COUNT(CASE WHEN EXTRACT(HOUR FROM c.opened_at) < 6
				            OR EXTRACT(HOUR FROM c.opened_at) >= 24
				           THEN 1 END)::float                                 AS off_hours_count,
				CASE WHEN SUM(c.subtotal) > 0
					THEN SUM(c.discount)::float / SUM(c.subtotal)::float * 100
					ELSE 0
				END                                                           AS discount_rate
			 FROM checks c
			 LEFT JOIN check_items ci ON ci.check_id = c.check_id
			 WHERE c.location_id = $1
			   AND c.closed_at >= $2 AND c.closed_at < $3
			   AND c.status = 'closed'
			 GROUP BY DATE(c.closed_at)`,
			locationID, baselineStart, dayStart,
		)
		if err != nil {
			return fmt.Errorf("baseline query: %w", err)
		}
		defer baselineRows.Close()

		var voidCounts, discountTotals, offHoursCounts, discountRates []float64
		for baselineRows.Next() {
			var d time.Time
			var voidCount, discountTotal, offHoursCount, discountRate float64
			if err := baselineRows.Scan(&d, &voidCount, &discountTotal, &offHoursCount, &discountRate); err != nil {
				return err
			}
			voidCounts = append(voidCounts, voidCount)
			discountTotals = append(discountTotals, discountTotal)
			offHoursCounts = append(offHoursCounts, offHoursCount)
			discountRates = append(discountRates, discountRate)
		}
		if err := baselineRows.Err(); err != nil {
			return err
		}

		// --- Get current day metrics ---
		var currentVoidCount, currentDiscountTotal, currentOffHoursCount float64
		var currentSubtotal, currentDiscount float64

		err = tx.QueryRow(tenantCtx,
			`SELECT
				COUNT(ci.voided_at)::float                                   AS void_count,
				COALESCE(SUM(c.discount), 0)::float                          AS discount_total,
				COUNT(CASE WHEN EXTRACT(HOUR FROM c.opened_at) < 6
				            OR EXTRACT(HOUR FROM c.opened_at) >= 24
				           THEN 1 END)::float                                 AS off_hours_count,
				COALESCE(SUM(c.subtotal), 0)::float,
				COALESCE(SUM(c.discount), 0)::float
			 FROM checks c
			 LEFT JOIN check_items ci ON ci.check_id = c.check_id
			 WHERE c.location_id = $1
			   AND c.closed_at >= $2 AND c.closed_at < $3
			   AND c.status = 'closed'`,
			locationID, dayStart, dayEnd,
		).Scan(&currentVoidCount, &currentDiscountTotal, &currentOffHoursCount, &currentSubtotal, &currentDiscount)
		if err != nil {
			return fmt.Errorf("current day metrics: %w", err)
		}

		currentDiscountRate := float64(0)
		if currentSubtotal > 0 {
			currentDiscountRate = currentDiscount / currentSubtotal * 100
		}

		// --- Z-score checks ---
		type metricCheck struct {
			metricType  string
			description string
			current     float64
			historical  []float64
		}

		checks := []metricCheck{
			{
				metricType:  "void_count",
				description: fmt.Sprintf("Void count of %.0f deviates from 30-day baseline", currentVoidCount),
				current:     currentVoidCount,
				historical:  voidCounts,
			},
			{
				metricType:  "discount_total",
				description: fmt.Sprintf("Discount total of $%.2f deviates from 30-day baseline", currentDiscountTotal/100),
				current:     currentDiscountTotal,
				historical:  discountTotals,
			},
			{
				metricType:  "off_hours_count",
				description: fmt.Sprintf("Off-hours transaction count of %.0f deviates from 30-day baseline", currentOffHoursCount),
				current:     currentOffHoursCount,
				historical:  offHoursCounts,
			},
			{
				metricType:  "discount_rate",
				description: fmt.Sprintf("Discount rate of %.1f%% deviates from 30-day baseline", currentDiscountRate),
				current:     currentDiscountRate,
				historical:  discountRates,
			},
		}

		for _, c := range checks {
			anomaly := detectZScoreAnomaly(c.metricType, c.current, c.historical)
			if anomaly == nil {
				continue
			}

			ta := TransactionAnomaly{
				Type:         c.metricType,
				Description:  c.description,
				CurrentValue: anomaly.CurrentValue,
				Baseline:     anomaly.Mean,
				ZScore:       anomaly.ZScore,
				Severity:     anomaly.Severity,
				DetectedAt:   anomaly.DetectedAt,
			}
			anomalies = append(anomalies, ta)

			// Emit alert event for critical anomalies
			if anomaly.Severity == "critical" {
				s.bus.Publish(ctx, event.Envelope{
					EventID:    fmt.Sprintf("tx-anomaly-%s-%s", c.metricType, day.Format("20060102")),
					EventType:  "financial.transaction.anomaly",
					OrgID:      orgID,
					LocationID: locationID,
					Source:     "financial",
					Payload: map[string]any{
						"type":          c.metricType,
						"current_value": c.current,
						"baseline":      anomaly.Mean,
						"z_score":       anomaly.ZScore,
						"severity":      anomaly.Severity,
					},
				})
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return anomalies, nil
}
