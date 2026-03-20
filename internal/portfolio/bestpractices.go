package portfolio

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// BestPractice represents a detected operational pattern worth adopting.
type BestPractice struct {
	PracticeID       string    `json:"practice_id"`
	OrgID            string    `json:"org_id"`
	Title            string    `json:"title"`
	Description      string    `json:"description"`
	Metric           string    `json:"metric"`
	SourceLocationID *string   `json:"source_location_id"`
	SourceName       string    `json:"source_name"`
	ImpactPct        float64   `json:"impact_pct"`
	Status           string    `json:"status"`
	DetectedAt       time.Time `json:"detected_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// DetectBestPractices compares top-quartile vs bottom-quartile locations
// and identifies metric gaps worth surfacing as best practices.
func (s *Service) DetectBestPractices(ctx context.Context, orgID string) ([]BestPractice, error) {
	to := time.Now()
	from := to.AddDate(0, -1, 0)

	benchmarks, err := s.GetBenchmarks(ctx, orgID, from, to)
	if err != nil {
		return nil, err
	}
	if len(benchmarks) < 4 {
		return nil, nil
	}

	type metricDef struct {
		name      string
		label     string
		extract   func(LocationBenchmark) float64
		higherBetter bool
	}

	metrics := []metricDef{
		{"revenue", "Revenue", func(b LocationBenchmark) float64 { return float64(b.Revenue) }, true},
		{"food_cost_pct", "Food Cost %", func(b LocationBenchmark) float64 { return b.FoodCostPct }, false},
		{"labor_cost_pct", "Labor Cost %", func(b LocationBenchmark) float64 { return b.LaborCostPct }, false},
		{"avg_check_cents", "Average Check", func(b LocationBenchmark) float64 { return float64(b.AvgCheckCents) }, true},
	}

	var practices []BestPractice

	for _, m := range metrics {
		values := make([]float64, len(benchmarks))
		for i, b := range benchmarks {
			values[i] = m.extract(b)
		}
		_, q1, q3 := Quartiles(values)

		var topLocs, bottomLocs []LocationBenchmark
		for _, b := range benchmarks {
			v := m.extract(b)
			if m.higherBetter {
				if v >= q3 {
					topLocs = append(topLocs, b)
				} else if v <= q1 {
					bottomLocs = append(bottomLocs, b)
				}
			} else {
				// lower is better: top performers have low values
				if v <= q1 {
					topLocs = append(topLocs, b)
				} else if v >= q3 {
					bottomLocs = append(bottomLocs, b)
				}
			}
		}

		if len(topLocs) == 0 || len(bottomLocs) == 0 {
			continue
		}

		// Average top vs average bottom
		var topSum, bottomSum float64
		for _, b := range topLocs {
			topSum += m.extract(b)
		}
		for _, b := range bottomLocs {
			bottomSum += m.extract(b)
		}
		topAvg := topSum / float64(len(topLocs))
		bottomAvg := bottomSum / float64(len(bottomLocs))

		var impactPct float64
		if bottomAvg != 0 {
			if m.higherBetter {
				impactPct = (topAvg - bottomAvg) / bottomAvg * 100
			} else {
				impactPct = (bottomAvg - topAvg) / bottomAvg * 100
			}
		}

		if impactPct < 5 {
			continue // only surface meaningful gaps
		}

		// Best source: the single top performer
		best := topLocs[0]
		for _, b := range topLocs {
			if m.higherBetter && m.extract(b) > m.extract(best) {
				best = b
			} else if !m.higherBetter && m.extract(b) < m.extract(best) {
				best = b
			}
		}

		locID := best.LocationID
		practices = append(practices, BestPractice{
			OrgID:            orgID,
			Title:            fmt.Sprintf("Improve %s", m.label),
			Description:      fmt.Sprintf("%s has %.1f%% better %s than bottom-quartile locations. Closing this gap could yield %.1f%% improvement.", best.LocationName, impactPct, m.label, impactPct),
			Metric:           m.name,
			SourceLocationID: &locID,
			SourceName:       best.LocationName,
			ImpactPct:        impactPct,
			Status:           "suggested",
			DetectedAt:       time.Now(),
			UpdatedAt:        time.Now(),
		})
	}

	// Upsert detected practices (skip if one with same org+metric already suggested)
	var saved []BestPractice
	for _, p := range practices {
		var existing string
		err := s.pool.QueryRow(ctx, `
			SELECT practice_id FROM best_practices
			WHERE org_id = $1 AND metric = $2 AND status = 'suggested'
			LIMIT 1
		`, orgID, p.Metric).Scan(&existing)

		if err == pgx.ErrNoRows {
			// Insert new
			var saved_p BestPractice
			src := p.SourceLocationID
			err = s.pool.QueryRow(ctx, `
				INSERT INTO best_practices (org_id, title, description, metric, source_location_id, impact_pct, status)
				VALUES ($1, $2, $3, $4, $5, $6, 'suggested')
				RETURNING practice_id, org_id, title, description, metric, source_location_id, impact_pct, status, detected_at, updated_at
			`, orgID, p.Title, p.Description, p.Metric, src, p.ImpactPct).Scan(
				&saved_p.PracticeID, &saved_p.OrgID, &saved_p.Title, &saved_p.Description,
				&saved_p.Metric, &saved_p.SourceLocationID, &saved_p.ImpactPct, &saved_p.Status,
				&saved_p.DetectedAt, &saved_p.UpdatedAt,
			)
			if err == nil {
				saved_p.SourceName = p.SourceName
				saved = append(saved, saved_p)
			}
		}
	}

	return saved, nil
}

// ListBestPractices returns stored best practices filtered by optional status.
func (s *Service) ListBestPractices(ctx context.Context, orgID, status string) ([]BestPractice, error) {
	query := `
		SELECT bp.practice_id, bp.org_id, bp.title, bp.description, bp.metric,
			   bp.source_location_id, COALESCE(l.name, ''), bp.impact_pct, bp.status,
			   bp.detected_at, bp.updated_at
		FROM best_practices bp
		LEFT JOIN locations l ON l.location_id = bp.source_location_id
		WHERE bp.org_id = $1
	`
	args := []any{orgID}
	if status != "" {
		query += " AND bp.status = $2"
		args = append(args, status)
	}
	query += " ORDER BY bp.impact_pct DESC, bp.detected_at DESC"

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var practices []BestPractice
	for rows.Next() {
		var p BestPractice
		if err := rows.Scan(
			&p.PracticeID, &p.OrgID, &p.Title, &p.Description, &p.Metric,
			&p.SourceLocationID, &p.SourceName, &p.ImpactPct, &p.Status,
			&p.DetectedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		practices = append(practices, p)
	}
	return practices, rows.Err()
}

// AdoptPractice marks a best practice as adopted.
func (s *Service) AdoptPractice(ctx context.Context, orgID, practiceID string) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE best_practices
		SET status = 'adopted', updated_at = now()
		WHERE practice_id = $1 AND org_id = $2
	`, practiceID, orgID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// DismissPractice marks a best practice as dismissed.
func (s *Service) DismissPractice(ctx context.Context, orgID, practiceID string) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE best_practices
		SET status = 'dismissed', updated_at = now()
		WHERE practice_id = $1 AND org_id = $2
	`, practiceID, orgID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
