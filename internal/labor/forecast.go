package labor

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/opsnerve/fireline/internal/tenant"
	"github.com/opsnerve/fireline/pkg/database"
)

// ForecastBlock holds demand forecast data for a single 30-minute time block.
type ForecastBlock struct {
	TimeBlock         string  `json:"time_block"`       // "11:00", "11:30", etc.
	ForecastedCovers  int     `json:"forecasted_covers"`
	RequiredELU       float64 `json:"required_elu"`
	RequiredHeadcount int     `json:"required_headcount"`
}

// GenerateForecast computes a demand forecast for targetDate by averaging
// check_count per 30-min block across the same day-of-week over the prior
// 4 weeks. Results are upserted into labor_demand_forecast and returned.
//
// Sizing: required_headcount = ceil(covers / 15), required_elu = headcount * 1.0.
func (s *Service) GenerateForecast(ctx context.Context, orgID, locationID string, targetDate time.Time) ([]ForecastBlock, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	var blocks []ForecastBlock

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		// Compute average covers per 30-min block for the same weekday over
		// the 4 weeks immediately preceding targetDate.
		rows, err := tx.Query(tenantCtx,
			`SELECT
			    to_char(date_trunc('hour', closed_at) +
			        INTERVAL '30 min' * FLOOR(EXTRACT(MINUTE FROM closed_at) / 30),
			        'HH24:MI') AS time_block,
			    ROUND(AVG(cover_count)) AS avg_covers
			FROM checks
			WHERE location_id = $1
			  AND status = 'closed'
			  AND EXTRACT(DOW FROM closed_at) = EXTRACT(DOW FROM $2::DATE)
			  AND closed_at >= $2::DATE - INTERVAL '28 days'
			  AND closed_at < $2::DATE
			GROUP BY time_block
			ORDER BY time_block`,
			locationID, targetDate.Format("2006-01-02"),
		)
		if err != nil {
			return fmt.Errorf("query forecast data: %w", err)
		}
		defer rows.Close()

		type rawBlock struct {
			timeBlock string
			covers    int
		}
		var raws []rawBlock

		for rows.Next() {
			var rb rawBlock
			if err := rows.Scan(&rb.timeBlock, &rb.covers); err != nil {
				return fmt.Errorf("scan forecast row: %w", err)
			}
			raws = append(raws, rb)
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("iterate forecast rows: %w", err)
		}

		// If no historical data, fall back to zero blocks (still upserted so
		// the manager can see a blank forecast rather than an error).
		if len(raws) == 0 {
			// Generate zeroed blocks for business hours 07:00–23:00
			for h := 7; h < 23; h++ {
				for _, m := range []string{"00", "30"} {
					raws = append(raws, rawBlock{
						timeBlock: fmt.Sprintf("%02d:%s", h, m),
						covers:    0,
					})
				}
			}
		}

		dateStr := targetDate.Format("2006-01-02")

		for _, rb := range raws {
			headcount := int(math.Ceil(float64(rb.covers) / 15.0))
			elu := float64(headcount) * 1.0

			fb := ForecastBlock{
				TimeBlock:         rb.timeBlock,
				ForecastedCovers:  rb.covers,
				RequiredELU:       elu,
				RequiredHeadcount: headcount,
			}
			blocks = append(blocks, fb)

			// Upsert into persistent forecast table.
			_, err := tx.Exec(tenantCtx,
				`INSERT INTO labor_demand_forecast
				    (org_id, location_id, forecast_date, time_block,
				     forecasted_covers, required_elu, required_headcount)
				 VALUES ($1, $2, $3, $4::TIME, $5, $6, $7)
				 ON CONFLICT (org_id, location_id, forecast_date, time_block)
				 DO UPDATE SET
				    forecasted_covers = EXCLUDED.forecasted_covers,
				    required_elu      = EXCLUDED.required_elu,
				    required_headcount = EXCLUDED.required_headcount`,
				orgID, locationID, dateStr, rb.timeBlock,
				rb.covers, elu, headcount,
			)
			if err != nil {
				return fmt.Errorf("upsert forecast block %s: %w", rb.timeBlock, err)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	if blocks == nil {
		blocks = []ForecastBlock{}
	}
	return blocks, nil
}

// GetForecast retrieves the stored forecast for a specific date and location,
// ordered by time_block ascending.
func (s *Service) GetForecast(ctx context.Context, orgID, locationID string, date time.Time) ([]ForecastBlock, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	var blocks []ForecastBlock

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(tenantCtx,
			`SELECT
			    to_char(time_block, 'HH24:MI') AS time_block,
			    forecasted_covers,
			    required_elu,
			    required_headcount
			FROM labor_demand_forecast
			WHERE location_id = $1
			  AND forecast_date = $2
			ORDER BY time_block`,
			locationID, date.Format("2006-01-02"),
		)
		if err != nil {
			return fmt.Errorf("query forecast: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var fb ForecastBlock
			if err := rows.Scan(
				&fb.TimeBlock,
				&fb.ForecastedCovers,
				&fb.RequiredELU,
				&fb.RequiredHeadcount,
			); err != nil {
				return fmt.Errorf("scan forecast block: %w", err)
			}
			blocks = append(blocks, fb)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}

	if blocks == nil {
		blocks = []ForecastBlock{}
	}
	return blocks, nil
}
