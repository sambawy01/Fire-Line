package portfolio

import (
	"context"
	"math"
	"sort"
	"time"
)

// LocationBenchmark holds computed percentile ranks for a location in a period.
type LocationBenchmark struct {
	BenchmarkID         string    `json:"benchmark_id"`
	OrgID               string    `json:"org_id"`
	LocationID          string    `json:"location_id"`
	LocationName        string    `json:"location_name"`
	PeriodStart         time.Time `json:"period_start"`
	PeriodEnd           time.Time `json:"period_end"`
	Revenue             int64     `json:"revenue"`
	FoodCostPct         float64   `json:"food_cost_pct"`
	LaborCostPct        float64   `json:"labor_cost_pct"`
	AvgCheckCents       int64     `json:"avg_check_cents"`
	CheckCount          int       `json:"check_count"`
	RevenuePercentile   float64   `json:"revenue_percentile"`
	FoodCostPercentile  float64   `json:"food_cost_percentile"`
	LaborCostPercentile float64   `json:"labor_cost_percentile"`
	AvgCheckPercentile  float64   `json:"avg_check_percentile"`
	ComputedAt          time.Time `json:"computed_at"`
}

// Outlier represents a location that deviates significantly from the peer group.
type Outlier struct {
	LocationID   string  `json:"location_id"`
	LocationName string  `json:"location_name"`
	Metric       string  `json:"metric"`
	Value        float64 `json:"value"`
	Median       float64 `json:"median"`
	IQR          float64 `json:"iqr"`
	Direction    string  `json:"direction"` // "above" or "below"
}

// CalculateBenchmarks computes per-location metrics and percentile ranks,
// then upserts into location_benchmarks.
func (s *Service) CalculateBenchmarks(ctx context.Context, orgID string, from, to time.Time) error {
	// Get all org locations
	rows, err := s.pool.Query(ctx, `
		SELECT location_id FROM locations WHERE org_id = $1
	`, orgID)
	if err != nil {
		return err
	}
	defer rows.Close()

	var locationIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return err
		}
		locationIDs = append(locationIDs, id)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if len(locationIDs) == 0 {
		return nil
	}

	metrics, err := s.GetLocationComparison(ctx, orgID, locationIDs, from, to)
	if err != nil {
		return err
	}

	// Extract slices for percentile computation
	revenues := make([]float64, len(metrics))
	foodCosts := make([]float64, len(metrics))
	laborCosts := make([]float64, len(metrics))
	avgChecks := make([]float64, len(metrics))

	for i, m := range metrics {
		revenues[i] = float64(m.Revenue)
		foodCosts[i] = m.FoodCostPct
		laborCosts[i] = m.LaborCostPct
		avgChecks[i] = float64(m.AvgCheckCents)
	}

	for i, m := range metrics {
		revPct := PercentileRank(revenues, revenues[i])
		// For cost metrics, lower is better — invert so high percentile = good
		fcPct := 100 - PercentileRank(foodCosts, foodCosts[i])
		lcPct := 100 - PercentileRank(laborCosts, laborCosts[i])
		acPct := PercentileRank(avgChecks, avgChecks[i])

		_, err := s.pool.Exec(ctx, `
			INSERT INTO location_benchmarks
				(org_id, location_id, period_start, period_end,
				 revenue, food_cost_pct, labor_cost_pct, avg_check_cents, check_count,
				 revenue_percentile, food_cost_percentile, labor_cost_percentile, avg_check_percentile, computed_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, now())
			ON CONFLICT (org_id, location_id, period_start, period_end) DO UPDATE
			SET revenue = EXCLUDED.revenue,
				food_cost_pct = EXCLUDED.food_cost_pct,
				labor_cost_pct = EXCLUDED.labor_cost_pct,
				avg_check_cents = EXCLUDED.avg_check_cents,
				check_count = EXCLUDED.check_count,
				revenue_percentile = EXCLUDED.revenue_percentile,
				food_cost_percentile = EXCLUDED.food_cost_percentile,
				labor_cost_percentile = EXCLUDED.labor_cost_percentile,
				avg_check_percentile = EXCLUDED.avg_check_percentile,
				computed_at = now()
		`, orgID, m.LocationID, from, to,
			m.Revenue, m.FoodCostPct, m.LaborCostPct, m.AvgCheckCents, m.CheckCount,
			revPct, fcPct, lcPct, acPct)
		if err != nil {
			return err
		}
		_ = i
	}

	return nil
}

// GetBenchmarks returns stored benchmarks for an org in a period.
func (s *Service) GetBenchmarks(ctx context.Context, orgID string, from, to time.Time) ([]LocationBenchmark, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT lb.benchmark_id, lb.org_id, lb.location_id, l.name,
			   lb.period_start, lb.period_end,
			   lb.revenue, lb.food_cost_pct, lb.labor_cost_pct, lb.avg_check_cents, lb.check_count,
			   COALESCE(lb.revenue_percentile, 0),
			   COALESCE(lb.food_cost_percentile, 0),
			   COALESCE(lb.labor_cost_percentile, 0),
			   COALESCE(lb.avg_check_percentile, 0),
			   lb.computed_at
		FROM location_benchmarks lb
		JOIN locations l ON l.location_id = lb.location_id
		WHERE lb.org_id = $1
		  AND lb.period_start >= $2
		  AND lb.period_end <= $3
		ORDER BY lb.location_id, lb.period_start
	`, orgID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var benchmarks []LocationBenchmark
	for rows.Next() {
		var b LocationBenchmark
		if err := rows.Scan(
			&b.BenchmarkID, &b.OrgID, &b.LocationID, &b.LocationName,
			&b.PeriodStart, &b.PeriodEnd,
			&b.Revenue, &b.FoodCostPct, &b.LaborCostPct, &b.AvgCheckCents, &b.CheckCount,
			&b.RevenuePercentile, &b.FoodCostPercentile, &b.LaborCostPercentile, &b.AvgCheckPercentile,
			&b.ComputedAt,
		); err != nil {
			return nil, err
		}
		benchmarks = append(benchmarks, b)
	}
	return benchmarks, rows.Err()
}

// DetectOutliers returns locations where any metric deviates by more than 1.5 IQR.
func (s *Service) DetectOutliers(ctx context.Context, orgID string, from, to time.Time) ([]Outlier, error) {
	benchmarks, err := s.GetBenchmarks(ctx, orgID, from, to)
	if err != nil {
		return nil, err
	}
	if len(benchmarks) < 4 {
		return nil, nil
	}

	type metricExtractor struct {
		name    string
		extract func(LocationBenchmark) float64
	}

	extractors := []metricExtractor{
		{"revenue", func(b LocationBenchmark) float64 { return float64(b.Revenue) }},
		{"food_cost_pct", func(b LocationBenchmark) float64 { return b.FoodCostPct }},
		{"labor_cost_pct", func(b LocationBenchmark) float64 { return b.LaborCostPct }},
		{"avg_check_cents", func(b LocationBenchmark) float64 { return float64(b.AvgCheckCents) }},
	}

	var outliers []Outlier

	for _, ex := range extractors {
		values := make([]float64, len(benchmarks))
		for i, b := range benchmarks {
			values[i] = ex.extract(b)
		}

		median, q1, q3 := Quartiles(values)
		iqr := q3 - q1
		lower := q1 - 1.5*iqr
		upper := q3 + 1.5*iqr

		for _, b := range benchmarks {
			v := ex.extract(b)
			if v < lower || v > upper {
				dir := "above"
				if v < lower {
					dir = "below"
				}
				outliers = append(outliers, Outlier{
					LocationID:   b.LocationID,
					LocationName: b.LocationName,
					Metric:       ex.name,
					Value:        v,
					Median:       median,
					IQR:          iqr,
					Direction:    dir,
				})
			}
		}
	}

	return outliers, nil
}

// PercentileRank returns the percentile rank of value in the sorted values slice.
// Result is 0..100.
func PercentileRank(values []float64, value float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	below := 0
	for _, v := range sorted {
		if v < value {
			below++
		}
	}
	return math.Round(float64(below)/float64(len(sorted))*100*100) / 100
}

// Quartiles returns the median, Q1, and Q3 of a set of values.
func Quartiles(values []float64) (median, q1, q3 float64) {
	if len(values) == 0 {
		return 0, 0, 0
	}
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	n := len(sorted)
	median = percentile(sorted, 50)
	q1 = percentile(sorted, 25)
	q3 = percentile(sorted, 75)
	_ = n
	return
}

func percentile(sorted []float64, p float64) float64 {
	n := len(sorted)
	if n == 0 {
		return 0
	}
	idx := p / 100 * float64(n-1)
	lo := int(math.Floor(idx))
	hi := int(math.Ceil(idx))
	if lo == hi {
		return sorted[lo]
	}
	frac := idx - float64(lo)
	return sorted[lo]*(1-frac) + sorted[hi]*frac
}
