package portfolio

import (
	"context"
	"time"
)

// AggregatedKPIs holds rolled-up KPIs for a portfolio node.
type AggregatedKPIs struct {
	NodeID        string    `json:"node_id"`
	OrgID         string    `json:"org_id"`
	PeriodStart   time.Time `json:"period_start"`
	PeriodEnd     time.Time `json:"period_end"`
	Revenue       int64     `json:"revenue"`
	FoodCostPct   float64   `json:"food_cost_pct"`
	LaborCostPct  float64   `json:"labor_cost_pct"`
	AvgCheckCents int64     `json:"avg_check_cents"`
	CheckCount    int       `json:"check_count"`
	LocationCount int       `json:"location_count"`
}

// LocationMetrics holds side-by-side KPIs for a single location.
type LocationMetrics struct {
	LocationID    string  `json:"location_id"`
	LocationName  string  `json:"location_name"`
	Revenue       int64   `json:"revenue"`
	FoodCostPct   float64 `json:"food_cost_pct"`
	LaborCostPct  float64 `json:"labor_cost_pct"`
	AvgCheckCents int64   `json:"avg_check_cents"`
	CheckCount    int     `json:"check_count"`
}

// AggregateKPIs finds all descendant location_ids under a portfolio node via
// recursive CTE, then aggregates revenue/costs from checks and shifts.
func (s *Service) AggregateKPIs(ctx context.Context, orgID, nodeID string, from, to time.Time) (*AggregatedKPIs, error) {
	// Step 1: collect all descendant location_ids via recursive CTE
	rows, err := s.pool.Query(ctx, `
		WITH RECURSIVE descendants AS (
			SELECT node_id, location_id FROM portfolio_nodes
			WHERE node_id = $1 AND org_id = $2
			UNION ALL
			SELECT pn.node_id, pn.location_id FROM portfolio_nodes pn
			INNER JOIN descendants d ON pn.parent_node_id = d.node_id
		)
		SELECT DISTINCT location_id FROM descendants WHERE location_id IS NOT NULL
	`, nodeID, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locationIDs []string
	for rows.Next() {
		var locID string
		if err := rows.Scan(&locID); err != nil {
			return nil, err
		}
		locationIDs = append(locationIDs, locID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(locationIDs) == 0 {
		return &AggregatedKPIs{
			NodeID:      nodeID,
			OrgID:       orgID,
			PeriodStart: from,
			PeriodEnd:   to,
		}, nil
	}

	metrics, err := s.GetLocationComparison(ctx, orgID, locationIDs, from, to)
	if err != nil {
		return nil, err
	}

	agg := &AggregatedKPIs{
		NodeID:        nodeID,
		OrgID:         orgID,
		PeriodStart:   from,
		PeriodEnd:     to,
		LocationCount: len(metrics),
	}

	var totalFoodCost, totalLaborCost float64
	for _, m := range metrics {
		agg.Revenue += m.Revenue
		agg.CheckCount += m.CheckCount
		totalFoodCost += m.FoodCostPct
		totalLaborCost += m.LaborCostPct
	}
	if agg.CheckCount > 0 {
		agg.AvgCheckCents = agg.Revenue / int64(agg.CheckCount)
	}
	if len(metrics) > 0 {
		agg.FoodCostPct = totalFoodCost / float64(len(metrics))
		agg.LaborCostPct = totalLaborCost / float64(len(metrics))
	}

	return agg, nil
}

// GetLocationComparison returns per-location KPIs for side-by-side comparison.
func (s *Service) GetLocationComparison(ctx context.Context, orgID string, locationIDs []string, from, to time.Time) ([]LocationMetrics, error) {
	if len(locationIDs) == 0 {
		return nil, nil
	}

	// Build revenue + check count from checks table
	rows, err := s.pool.Query(ctx, `
		SELECT
			c.location_id,
			l.name AS location_name,
			COALESCE(SUM(c.net_total), 0)   AS revenue,
			COUNT(*)                         AS check_count
		FROM checks c
		JOIN locations l ON l.location_id = c.location_id
		WHERE c.org_id = $1
		  AND c.location_id = ANY($2::uuid[])
		  AND c.closed_at >= $3
		  AND c.closed_at < $4
		  AND c.status = 'closed'
		GROUP BY c.location_id, l.name
	`, orgID, locationIDs, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	metricsMap := make(map[string]*LocationMetrics)
	for rows.Next() {
		var m LocationMetrics
		if err := rows.Scan(&m.LocationID, &m.LocationName, &m.Revenue, &m.CheckCount); err != nil {
			return nil, err
		}
		if m.CheckCount > 0 {
			m.AvgCheckCents = m.Revenue / int64(m.CheckCount)
		}
		metricsMap[m.LocationID] = &m
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Compute food cost % from COGS relative to revenue
	fcRows, err := s.pool.Query(ctx, `
		SELECT
			ci.location_id,
			CASE WHEN SUM(c.net_total) > 0
				 THEN ROUND(SUM(ci.total_cost)::numeric / SUM(c.net_total) * 100, 3)
				 ELSE 0 END AS food_cost_pct
		FROM check_items ci
		JOIN checks c ON c.check_id = ci.check_id
		WHERE ci.org_id = $1
		  AND ci.location_id = ANY($2::uuid[])
		  AND c.closed_at >= $3
		  AND c.closed_at < $4
		  AND c.status = 'closed'
		GROUP BY ci.location_id
	`, orgID, locationIDs, from, to)
	if err == nil {
		defer fcRows.Close()
		for fcRows.Next() {
			var locID string
			var pct float64
			if err := fcRows.Scan(&locID, &pct); err == nil {
				if m, ok := metricsMap[locID]; ok {
					m.FoodCostPct = pct
				}
			}
		}
	}

	// Compute labor cost % from shifts
	lcRows, err := s.pool.Query(ctx, `
		SELECT
			sh.location_id,
			CASE WHEN rev.revenue > 0
				 THEN ROUND(SUM((EXTRACT(EPOCH FROM (sh.clock_out - sh.clock_in)) / 3600.0) * sh.hourly_rate_cents)::numeric / rev.revenue * 100, 3)
				 ELSE 0 END AS labor_cost_pct
		FROM shifts sh
		JOIN (
			SELECT location_id, SUM(net_total) AS revenue
			FROM checks
			WHERE org_id = $1 AND location_id = ANY($2::uuid[]) AND closed_at >= $3 AND closed_at < $4 AND status = 'closed'
			GROUP BY location_id
		) rev ON rev.location_id = sh.location_id
		WHERE sh.org_id = $1
		  AND sh.location_id = ANY($2::uuid[])
		  AND sh.clock_in >= $3
		  AND sh.clock_in < $4
		  AND sh.clock_out IS NOT NULL
		GROUP BY sh.location_id, rev.revenue
	`, orgID, locationIDs, from, to)
	if err == nil {
		defer lcRows.Close()
		for lcRows.Next() {
			var locID string
			var pct float64
			if err := lcRows.Scan(&locID, &pct); err == nil {
				if m, ok := metricsMap[locID]; ok {
					m.LaborCostPct = pct
				}
			}
		}
	}

	// Ensure all requested locations are represented
	for _, locID := range locationIDs {
		if _, exists := metricsMap[locID]; !exists {
			metricsMap[locID] = &LocationMetrics{LocationID: locID}
		}
	}

	result := make([]LocationMetrics, 0, len(metricsMap))
	for _, m := range metricsMap {
		result = append(result, *m)
	}
	return result, nil
}
