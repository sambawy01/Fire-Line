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

// PricePoint represents a single price observation for an ingredient+vendor pair.
type PricePoint struct {
	UnitCost   int       `json:"unit_cost"`
	Quantity   float64   `json:"quantity"`
	RecordedAt time.Time `json:"recorded_at"`
	Source     string    `json:"source"`
}

// PriceAnomaly describes a statistically significant price deviation.
type PriceAnomaly struct {
	IngredientID   string  `json:"ingredient_id"`
	IngredientName string  `json:"ingredient_name"`
	VendorName     string  `json:"vendor_name"`
	CurrentPrice   int     `json:"current_price"`
	AvgPrice       float64 `json:"avg_price"`
	ZScore         float64 `json:"z_score"`
	Severity       string  `json:"severity"` // "warning" (2-3σ) or "critical" (>3σ)
}

// VendorRecommendation ranks a vendor for a specific ingredient purchase.
type VendorRecommendation struct {
	VendorName       string  `json:"vendor_name"`
	Score            float64 `json:"score"`
	UnitCost         int     `json:"unit_cost"`
	ReliabilityScore float64 `json:"reliability_score"`
	LeadTimeDays     float64 `json:"lead_time_days"`
	Reasoning        string  `json:"reasoning"`
}

// RecordPrice inserts a price observation into ingredient_price_history.
func (s *Service) RecordPrice(ctx context.Context, orgID, ingredientID, vendorName string, unitCost int, quantity float64, source string) error {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(tenantCtx, `
			INSERT INTO ingredient_price_history
				(org_id, ingredient_id, vendor_name, unit_cost, quantity, source)
			VALUES ($1, $2, $3, $4, $5, $6)`,
			orgID, ingredientID, vendorName, unitCost, quantity, source)
		return err
	})
	if err != nil {
		return fmt.Errorf("record price: %w", err)
	}
	return nil
}

// GetPriceTrend returns price history for an ingredient+vendor pair over N months.
func (s *Service) GetPriceTrend(ctx context.Context, orgID, ingredientID, vendorName string, months int) ([]PricePoint, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	if months <= 0 {
		months = 3
	}
	cutoff := time.Now().AddDate(0, -months, 0)

	var points []PricePoint
	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(tenantCtx, `
			SELECT unit_cost, COALESCE(quantity, 0), recorded_at, source
			FROM ingredient_price_history
			WHERE org_id = $1
			  AND ingredient_id = $2
			  AND vendor_name = $3
			  AND recorded_at >= $4
			ORDER BY recorded_at ASC`,
			orgID, ingredientID, vendorName, cutoff)
		if err != nil {
			return fmt.Errorf("query price trend: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var pp PricePoint
			if err := rows.Scan(&pp.UnitCost, &pp.Quantity, &pp.RecordedAt, &pp.Source); err != nil {
				return fmt.Errorf("scan price point: %w", err)
			}
			points = append(points, pp)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	if points == nil {
		return []PricePoint{}, nil
	}
	return points, nil
}

// DetectPriceAnomalies scans all ingredient+vendor pairs with 5+ price points and
// flags those where the latest price deviates more than 2σ from the trailing average.
func (s *Service) DetectPriceAnomalies(ctx context.Context, orgID, locationID string) ([]PriceAnomaly, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	// Fetch all ingredient+vendor pairs that have 5+ price points.
	type pairKey struct {
		ingredientID   string
		ingredientName string
		vendorName     string
	}

	var pairs []pairKey
	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(tenantCtx, `
			SELECT ph.ingredient_id, i.name, ph.vendor_name
			FROM ingredient_price_history ph
			JOIN ingredients i ON i.ingredient_id = ph.ingredient_id AND i.org_id = ph.org_id
			WHERE ph.org_id = $1
			GROUP BY ph.ingredient_id, i.name, ph.vendor_name
			HAVING COUNT(*) >= 5`,
			orgID)
		if err != nil {
			return fmt.Errorf("query price pairs: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var pk pairKey
			if err := rows.Scan(&pk.ingredientID, &pk.ingredientName, &pk.vendorName); err != nil {
				return fmt.Errorf("scan pair: %w", err)
			}
			pairs = append(pairs, pk)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}

	var anomalies []PriceAnomaly
	for _, pk := range pairs {
		// Fetch all price points for this pair, ordered oldest-first.
		points, err := s.GetPriceTrend(ctx, orgID, pk.ingredientID, pk.vendorName, 24)
		if err != nil || len(points) < 5 {
			continue
		}

		latest := points[len(points)-1]
		trailing := points[:len(points)-1]

		// Compute mean and std dev of the trailing prices.
		var sum float64
		for _, p := range trailing {
			sum += float64(p.UnitCost)
		}
		mean := sum / float64(len(trailing))

		var variance float64
		for _, p := range trailing {
			diff := float64(p.UnitCost) - mean
			variance += diff * diff
		}
		variance /= float64(len(trailing))
		stdDev := math.Sqrt(variance)

		if stdDev == 0 {
			continue
		}

		zScore := (float64(latest.UnitCost) - mean) / stdDev
		absZ := math.Abs(zScore)
		if absZ <= 2.0 {
			continue
		}

		severity := "warning"
		if absZ > 3.0 {
			severity = "critical"
		}

		anomaly := PriceAnomaly{
			IngredientID:   pk.ingredientID,
			IngredientName: pk.ingredientName,
			VendorName:     pk.vendorName,
			CurrentPrice:   latest.UnitCost,
			AvgPrice:       math.Round(mean*100) / 100,
			ZScore:         math.Round(zScore*100) / 100,
			Severity:       severity,
		}
		anomalies = append(anomalies, anomaly)

		// Emit alert event.
		s.bus.Publish(ctx, event.Envelope{
			EventType:  "vendor.price.anomaly",
			OrgID:      orgID,
			LocationID: locationID,
			Source:     "vendor",
			Payload:    anomaly,
		})
	}

	if anomalies == nil {
		return []PriceAnomaly{}, nil
	}
	return anomalies, nil
}

// RecommendVendor ranks vendors for a given ingredient by a composite score:
// price 40%, vendor reliability score 40%, lead time 20%.
// Returns the top recommendation with reasoning.
func (s *Service) RecommendVendor(ctx context.Context, orgID, locationID, ingredientID string) ([]VendorRecommendation, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	type vendorCandidate struct {
		vendorName   string
		latestCost   int
		leadTimeDays float64
	}

	var candidates []vendorCandidate
	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		// Get all vendors supplying this ingredient with their latest price and lead time.
		rows, err := tx.Query(tenantCtx, `
			SELECT
				ilc.vendor_name,
				COALESCE(
					(SELECT unit_cost FROM ingredient_price_history
					 WHERE org_id = $1 AND ingredient_id = $3 AND vendor_name = ilc.vendor_name
					 ORDER BY recorded_at DESC LIMIT 1),
					COALESCE(ilc.local_cost_per_unit, i.cost_per_unit)
				) AS unit_cost,
				COALESCE(ilc.lead_time_days, 3)::FLOAT AS lead_time_days
			FROM ingredient_location_configs ilc
			JOIN ingredients i ON i.ingredient_id = ilc.ingredient_id AND i.org_id = ilc.org_id
			WHERE ilc.org_id = $1
			  AND ilc.location_id = $2
			  AND ilc.ingredient_id = $3
			  AND ilc.vendor_name IS NOT NULL AND ilc.vendor_name != ''`,
			orgID, locationID, ingredientID)
		if err != nil {
			return fmt.Errorf("query vendor candidates: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var c vendorCandidate
			if err := rows.Scan(&c.vendorName, &c.latestCost, &c.leadTimeDays); err != nil {
				return fmt.Errorf("scan candidate: %w", err)
			}
			candidates = append(candidates, c)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}

	if len(candidates) == 0 {
		return []VendorRecommendation{}, nil
	}

	// Find min/max cost and lead time for normalization.
	minCost, maxCost := candidates[0].latestCost, candidates[0].latestCost
	minLead, maxLead := candidates[0].leadTimeDays, candidates[0].leadTimeDays
	for _, c := range candidates[1:] {
		if c.latestCost < minCost {
			minCost = c.latestCost
		}
		if c.latestCost > maxCost {
			maxCost = c.latestCost
		}
		if c.leadTimeDays < minLead {
			minLead = c.leadTimeDays
		}
		if c.leadTimeDays > maxLead {
			maxLead = c.leadTimeDays
		}
	}

	recs := make([]VendorRecommendation, 0, len(candidates))
	for _, c := range candidates {
		// Normalize price score: cheaper = higher score (inverted).
		priceScore := 100.0
		if maxCost > minCost {
			priceScore = (1.0 - float64(c.latestCost-minCost)/float64(maxCost-minCost)) * 100
		}

		// Reliability score from vendor_scores table (default 50 if absent).
		reliabilityScore := 50.0
		vs, err := s.GetVendorScorecard(ctx, orgID, locationID, c.vendorName)
		if err == nil {
			reliabilityScore = vs.OverallScore
		}

		// Lead time score: shorter = higher (inverted).
		leadScore := 100.0
		if maxLead > minLead {
			leadScore = (1.0 - (c.leadTimeDays-minLead)/(maxLead-minLead)) * 100
		}

		composite := priceScore*0.40 + reliabilityScore*0.40 + leadScore*0.20
		composite = math.Round(composite*100) / 100

		reasoning := fmt.Sprintf(
			"Price score %.0f/100 (unit cost %d), reliability %.0f/100, lead time %.1f days (score %.0f/100)",
			priceScore, c.latestCost, reliabilityScore, c.leadTimeDays, leadScore,
		)

		recs = append(recs, VendorRecommendation{
			VendorName:       c.vendorName,
			Score:            composite,
			UnitCost:         c.latestCost,
			ReliabilityScore: reliabilityScore,
			LeadTimeDays:     c.leadTimeDays,
			Reasoning:        reasoning,
		})
	}

	// Sort by composite score descending (simple insertion sort for small slices).
	for i := 1; i < len(recs); i++ {
		for j := i; j > 0 && recs[j].Score > recs[j-1].Score; j-- {
			recs[j], recs[j-1] = recs[j-1], recs[j]
		}
	}

	return recs, nil
}
