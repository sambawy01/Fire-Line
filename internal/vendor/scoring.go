package vendor

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

// VendorScore holds computed reliability and performance metrics for a single vendor.
type VendorScore struct {
	ScoreID       string  `json:"score_id"`
	VendorName    string  `json:"vendor_name"`
	OverallScore  float64 `json:"overall_score"`
	PriceScore    float64 `json:"price_score"`
	DeliveryScore float64 `json:"delivery_score"`
	QualityScore  float64 `json:"quality_score"`
	AccuracyScore float64 `json:"accuracy_score"`
	TotalOrders   int     `json:"total_orders"`
	OTIFRate      float64 `json:"otif_rate"`
	OnTimeRate    float64 `json:"on_time_rate"`
	InFullRate    float64 `json:"in_full_rate"`
	AvgLeadDays   float64 `json:"avg_lead_days"`
}

// VendorComparison holds a per-ingredient vendor comparison with a recommendation.
type VendorComparison struct {
	IngredientName string        `json:"ingredient_name"`
	Vendors        []VendorScore `json:"vendors"`
	Recommended    string        `json:"recommended"`
}

// calculateOverallScore computes a weighted overall score from four sub-scores.
// Weights: price 30%, delivery 25%, quality 25%, accuracy 20%.
func calculateOverallScore(price, delivery, quality, accuracy float64) float64 {
	raw := price*0.30 + delivery*0.25 + quality*0.25 + accuracy*0.20
	// Round to 2 decimal places.
	return math.Round(raw*100) / 100
}

// CalculateVendorScores queries received POs from the last 90 days, computes
// sub-scores for each vendor, and UPSERTs results into vendor_scores.
func (s *Service) CalculateVendorScores(ctx context.Context, orgID, locationID string) ([]VendorScore, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	cutoff := time.Now().AddDate(0, 0, -90)

	// Intermediate accumulator per vendor.
	type vendorStats struct {
		totalOrders   int
		onTimeCount   int
		inFullCount   int
		otifCount     int
		totalLeadDays float64
		// price deviation: abs pct from estimated to actual cost
		totalPriceDev float64
		priceDevCount int
		// quality: inverse of variance (short/not_received treated as quality issues)
		qualityOK int
		// accuracy: exact qty fulfilled
		accuracyOK int
	}
	stats := make(map[string]*vendorStats)

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(tenantCtx, `
			SELECT
				po.vendor_name,
				po.approved_at,
				po.received_at,
				pol.ordered_qty,
				pol.received_qty,
				pol.estimated_unit_cost,
				pol.received_unit_cost,
				pol.variance_flag,
				ilc.lead_time_days
			FROM purchase_orders po
			JOIN purchase_order_lines pol
				ON pol.purchase_order_id = po.purchase_order_id
				AND pol.org_id = po.org_id
			LEFT JOIN ingredient_location_configs ilc
				ON ilc.ingredient_id = pol.ingredient_id
				AND ilc.org_id = po.org_id
				AND ilc.location_id = po.location_id
			WHERE po.org_id = $1
			  AND po.location_id = $2
			  AND po.status = 'received'
			  AND po.received_at >= $3
		`, orgID, locationID, cutoff)
		if err != nil {
			return fmt.Errorf("query po data: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var vendorName string
			var approvedAt, receivedAt *time.Time
			var orderedQty float64
			var receivedQty *float64
			var estimatedCost int
			var receivedCost *int
			var varianceFlag *string
			var leadTimeDays *int

			if err := rows.Scan(
				&vendorName, &approvedAt, &receivedAt,
				&orderedQty, &receivedQty,
				&estimatedCost, &receivedCost,
				&varianceFlag, &leadTimeDays,
			); err != nil {
				return fmt.Errorf("scan po row: %w", err)
			}

			st, ok := stats[vendorName]
			if !ok {
				st = &vendorStats{}
				stats[vendorName] = st
			}
			st.totalOrders++

			// Lead time: days between approved_at and received_at.
			if approvedAt != nil && receivedAt != nil && receivedAt.After(*approvedAt) {
				days := receivedAt.Sub(*approvedAt).Hours() / 24
				st.totalLeadDays += days

				// On-time: received within configured lead_time_days (default 3 if unknown).
				expectedDays := 3.0
				if leadTimeDays != nil && *leadTimeDays > 0 {
					expectedDays = float64(*leadTimeDays)
				}
				if days <= expectedDays {
					st.onTimeCount++
				}
			}

			// In-full: received_qty >= ordered_qty (within 2% tolerance).
			if receivedQty != nil && orderedQty > 0 {
				ratio := *receivedQty / orderedQty
				if ratio >= 0.98 {
					st.inFullCount++
				}
			}

			// OTIF: both on-time and in-full.
			if approvedAt != nil && receivedAt != nil && receivedQty != nil && orderedQty > 0 {
				days := receivedAt.Sub(*approvedAt).Hours() / 24
				expectedDays := 3.0
				if leadTimeDays != nil && *leadTimeDays > 0 {
					expectedDays = float64(*leadTimeDays)
				}
				ratio := *receivedQty / orderedQty
				if days <= expectedDays && ratio >= 0.98 {
					st.otifCount++
				}
			}

			// Price deviation: abs pct diff between estimated and actual cost.
			if receivedCost != nil && estimatedCost > 0 {
				dev := math.Abs(float64(*receivedCost-estimatedCost)) / float64(estimatedCost) * 100
				st.totalPriceDev += dev
				st.priceDevCount++
			}

			// Quality: treat short/not_received as quality issues.
			if varianceFlag != nil {
				switch *varianceFlag {
				case "exact", "over":
					st.qualityOK++
				}
			} else {
				st.qualityOK++ // no flag = no issue
			}

			// Accuracy: exact match.
			if varianceFlag != nil && *varianceFlag == "exact" {
				st.accuracyOK++
			} else if varianceFlag == nil {
				st.accuracyOK++
			}
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("iterate po rows: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	if len(stats) == 0 {
		return []VendorScore{}, nil
	}

	// Compute scores and upsert.
	scores := make([]VendorScore, 0, len(stats))
	for vendorName, st := range stats {
		if st.totalOrders == 0 {
			continue
		}

		n := float64(st.totalOrders)
		onTimeRate := float64(st.onTimeCount) / n * 100
		inFullRate := float64(st.inFullCount) / n * 100
		otifRate := float64(st.otifCount) / n * 100
		avgLeadDays := 0.0
		if st.totalLeadDays > 0 {
			avgLeadDays = st.totalLeadDays / n
		}

		// Price score: 100 - avg_price_deviation, clamped 0-100.
		priceScore := 100.0
		if st.priceDevCount > 0 {
			priceScore = 100.0 - (st.totalPriceDev / float64(st.priceDevCount))
		}
		if priceScore < 0 {
			priceScore = 0
		}

		// Delivery score: OTIF rate.
		deliveryScore := otifRate

		// Quality score: pct of orders with no quality variance.
		qualityScore := float64(st.qualityOK) / n * 100

		// Accuracy score: pct of exact fulfillments.
		accuracyScore := float64(st.accuracyOK) / n * 100

		overallScore := calculateOverallScore(priceScore, deliveryScore, qualityScore, accuracyScore)

		vs := VendorScore{
			VendorName:    vendorName,
			OverallScore:  overallScore,
			PriceScore:    math.Round(priceScore*100) / 100,
			DeliveryScore: math.Round(deliveryScore*100) / 100,
			QualityScore:  math.Round(qualityScore*100) / 100,
			AccuracyScore: math.Round(accuracyScore*100) / 100,
			TotalOrders:   st.totalOrders,
			OTIFRate:      math.Round(otifRate*100) / 100,
			OnTimeRate:    math.Round(onTimeRate*100) / 100,
			InFullRate:    math.Round(inFullRate*100) / 100,
			AvgLeadDays:   math.Round(avgLeadDays*100) / 100,
		}

		// Upsert into vendor_scores.
		upsertCtx := tenant.WithOrgID(ctx, orgID)
		upsertErr := database.TenantTx(upsertCtx, s.pool, func(tx pgx.Tx) error {
			row := tx.QueryRow(upsertCtx, `
				INSERT INTO vendor_scores (
					org_id, location_id, vendor_name,
					overall_score, price_score, delivery_score, quality_score, accuracy_score,
					total_orders, otif_rate, on_time_rate, in_full_rate, avg_lead_days,
					calculated_at, updated_at
				) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,now(),now())
				ON CONFLICT (org_id, location_id, vendor_name) DO UPDATE SET
					overall_score  = EXCLUDED.overall_score,
					price_score    = EXCLUDED.price_score,
					delivery_score = EXCLUDED.delivery_score,
					quality_score  = EXCLUDED.quality_score,
					accuracy_score = EXCLUDED.accuracy_score,
					total_orders   = EXCLUDED.total_orders,
					otif_rate      = EXCLUDED.otif_rate,
					on_time_rate   = EXCLUDED.on_time_rate,
					in_full_rate   = EXCLUDED.in_full_rate,
					avg_lead_days  = EXCLUDED.avg_lead_days,
					calculated_at  = now(),
					updated_at     = now()
				RETURNING score_id`,
				orgID, locationID, vendorName,
				vs.OverallScore, vs.PriceScore, vs.DeliveryScore, vs.QualityScore, vs.AccuracyScore,
				vs.TotalOrders, vs.OTIFRate, vs.OnTimeRate, vs.InFullRate, vs.AvgLeadDays,
			)
			return row.Scan(&vs.ScoreID)
		})
		if upsertErr != nil {
			return nil, fmt.Errorf("upsert vendor score for %q: %w", vendorName, upsertErr)
		}

		scores = append(scores, vs)

		// Emit event.
		s.bus.Publish(ctx, event.Envelope{
			EventType:  "vendor.score.calculated",
			OrgID:      orgID,
			LocationID: locationID,
			Source:     "vendor",
			Payload:    vs,
		})
	}

	return scores, nil
}

// GetVendorScores returns all stored vendor scores for a location.
func (s *Service) GetVendorScores(ctx context.Context, orgID, locationID string) ([]VendorScore, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var scores []VendorScore

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(tenantCtx, `
			SELECT score_id, vendor_name, overall_score, price_score, delivery_score,
				   quality_score, accuracy_score, total_orders, otif_rate, on_time_rate,
				   in_full_rate, avg_lead_days
			FROM vendor_scores
			WHERE org_id = $1 AND location_id = $2
			ORDER BY overall_score DESC`, orgID, locationID)
		if err != nil {
			return fmt.Errorf("query vendor scores: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var vs VendorScore
			if err := rows.Scan(
				&vs.ScoreID, &vs.VendorName, &vs.OverallScore, &vs.PriceScore,
				&vs.DeliveryScore, &vs.QualityScore, &vs.AccuracyScore, &vs.TotalOrders,
				&vs.OTIFRate, &vs.OnTimeRate, &vs.InFullRate, &vs.AvgLeadDays,
			); err != nil {
				return fmt.Errorf("scan vendor score: %w", err)
			}
			scores = append(scores, vs)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	if scores == nil {
		return []VendorScore{}, nil
	}
	return scores, nil
}

// GetVendorScorecard returns the score record for a single named vendor.
func (s *Service) GetVendorScorecard(ctx context.Context, orgID, locationID, vendorName string) (*VendorScore, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var vs VendorScore

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		row := tx.QueryRow(tenantCtx, `
			SELECT score_id, vendor_name, overall_score, price_score, delivery_score,
				   quality_score, accuracy_score, total_orders, otif_rate, on_time_rate,
				   in_full_rate, avg_lead_days
			FROM vendor_scores
			WHERE org_id = $1 AND location_id = $2 AND vendor_name = $3`,
			orgID, locationID, vendorName)
		return row.Scan(
			&vs.ScoreID, &vs.VendorName, &vs.OverallScore, &vs.PriceScore,
			&vs.DeliveryScore, &vs.QualityScore, &vs.AccuracyScore, &vs.TotalOrders,
			&vs.OTIFRate, &vs.OnTimeRate, &vs.InFullRate, &vs.AvgLeadDays,
		)
	})
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, fmt.Errorf("vendor scorecard not found: %s", vendorName)
		}
		return nil, err
	}
	return &vs, nil
}

// CompareVendors returns all vendors that supply the given ingredient, with
// their scores and a recommendation for the best option.
func (s *Service) CompareVendors(ctx context.Context, orgID, locationID, ingredientID string) (*VendorComparison, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	var ingredientName string
	var vendorNames []string

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		// Get ingredient name.
		row := tx.QueryRow(tenantCtx,
			`SELECT name FROM ingredients WHERE org_id = $1 AND ingredient_id = $2`,
			orgID, ingredientID)
		if err := row.Scan(&ingredientName); err != nil {
			return fmt.Errorf("get ingredient name: %w", err)
		}

		// Find all vendors that supply this ingredient via ingredient_location_configs.
		rows, err := tx.Query(tenantCtx, `
			SELECT DISTINCT vendor_name
			FROM ingredient_location_configs
			WHERE org_id = $1 AND location_id = $2
			  AND ingredient_id = $3
			  AND vendor_name IS NOT NULL AND vendor_name != ''`,
			orgID, locationID, ingredientID)
		if err != nil {
			return fmt.Errorf("query vendors for ingredient: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var vn string
			if err := rows.Scan(&vn); err != nil {
				return fmt.Errorf("scan vendor name: %w", err)
			}
			vendorNames = append(vendorNames, vn)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}

	// Fetch scores for each vendor.
	scores := make([]VendorScore, 0, len(vendorNames))
	for _, vn := range vendorNames {
		vs, err := s.GetVendorScorecard(ctx, orgID, locationID, vn)
		if err != nil {
			// If no scorecard exists yet, include a zero-score placeholder.
			scores = append(scores, VendorScore{VendorName: vn})
			continue
		}
		scores = append(scores, *vs)
	}

	// Determine recommended vendor (highest overall score).
	recommended := ""
	var bestScore float64
	for _, vs := range scores {
		if vs.OverallScore > bestScore {
			bestScore = vs.OverallScore
			recommended = vs.VendorName
		}
	}

	return &VendorComparison{
		IngredientName: ingredientName,
		Vendors:        scores,
		Recommended:    recommended,
	}, nil
}
