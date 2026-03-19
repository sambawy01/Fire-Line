package operations

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/opsnerve/fireline/internal/event"
	"github.com/opsnerve/fireline/internal/tenant"
	"github.com/opsnerve/fireline/pkg/database"
)

// Service provides operations intelligence capabilities.
type Service struct {
	pool *pgxpool.Pool
	bus  *event.Bus
}

// New creates a new operations intelligence service.
func New(pool *pgxpool.Pool, bus *event.Bus) *Service {
	return &Service{pool: pool, bus: bus}
}

// OperationsSummary holds the computed operational KPIs for a location.
type OperationsSummary struct {
	OrdersToday        int           `json:"orders_today"`
	AvgTicketTime      float64       `json:"avg_ticket_time"`
	OrdersPerHour      int           `json:"orders_per_hour"`
	ActiveTickets      int           `json:"active_tickets"`
	LongestOpenMin     float64       `json:"longest_open_min"`
	RevenuePerHour     int64         `json:"revenue_per_hour"`
	VoidRate           float64       `json:"void_rate"`
	ChannelPerformance []ChannelPerf `json:"channel_performance"`
}

// ChannelPerf holds per-channel operational metrics.
type ChannelPerf struct {
	Channel       string  `json:"channel"`
	Orders        int     `json:"orders"`
	PctOfTotal    float64 `json:"pct_of_total"`
	AvgTicketTime float64 `json:"avg_ticket_time"`
	Revenue       int64   `json:"revenue"`
}

// HourlyData holds order and revenue data aggregated by hour.
type HourlyData struct {
	Hour    int   `json:"hour"`
	Orders  int   `json:"orders"`
	Revenue int64 `json:"revenue"`
}

// GetSummary returns operational KPIs for the given location and time period.
func (s *Service) GetSummary(ctx context.Context, orgID, locationID string, from, to time.Time) (*OperationsSummary, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	var summary OperationsSummary

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		// --- Query 1: closed orders, avg ticket time, total revenue ---
		var closedCount int
		var avgTicketMin float64
		var totalRevenue int64
		if err := tx.QueryRow(tenantCtx,
			`SELECT COUNT(*) AS order_count,
			        COALESCE(AVG(EXTRACT(EPOCH FROM (closed_at - opened_at)) / 60.0), 0) AS avg_ticket_min,
			        COALESCE(SUM(subtotal), 0)::BIGINT AS total_revenue
			 FROM checks
			 WHERE location_id = $1 AND status = 'closed'
			   AND closed_at >= $2 AND closed_at < $3`,
			locationID, from, to,
		).Scan(&closedCount, &avgTicketMin, &totalRevenue); err != nil {
			return fmt.Errorf("query closed orders: %w", err)
		}

		// --- Query 2: current hour metrics ---
		var ordersPerHour int
		var revenuePerHour int64
		if err := tx.QueryRow(tenantCtx,
			`SELECT COUNT(*) AS orders_per_hour,
			        COALESCE(SUM(subtotal), 0)::BIGINT AS revenue_per_hour
			 FROM checks
			 WHERE location_id = $1 AND status = 'closed'
			   AND closed_at >= now() - INTERVAL '1 hour'`,
			locationID,
		).Scan(&ordersPerHour, &revenuePerHour); err != nil {
			return fmt.Errorf("query current hour metrics: %w", err)
		}

		// --- Query 3: active tickets + longest open ---
		var activeCount int
		var longestMin float64
		if err := tx.QueryRow(tenantCtx,
			`SELECT COUNT(*) AS active,
			        COALESCE(EXTRACT(EPOCH FROM (now() - MIN(opened_at))) / 60.0, 0) AS longest_min
			 FROM checks
			 WHERE location_id = $1 AND status = 'open'`,
			locationID,
		).Scan(&activeCount, &longestMin); err != nil {
			return fmt.Errorf("query active tickets: %w", err)
		}

		// --- Query 4: void count ---
		var voidCount int
		if err := tx.QueryRow(tenantCtx,
			`SELECT COUNT(*) FROM checks
			 WHERE location_id = $1 AND status = 'voided'
			   AND opened_at >= $2 AND opened_at < $3`,
			locationID, from, to,
		).Scan(&voidCount); err != nil {
			return fmt.Errorf("query void count: %w", err)
		}

		// Void rate: guard against division by zero.
		var voidRate float64
		total := closedCount + voidCount
		if total > 0 {
			voidRate = float64(voidCount) / float64(total) * 100
		}

		// --- Query 5: channel performance ---
		rows, err := tx.Query(tenantCtx,
			`SELECT channel, COUNT(*) AS orders,
			        COALESCE(AVG(EXTRACT(EPOCH FROM (closed_at - opened_at)) / 60.0), 0) AS avg_ticket,
			        COALESCE(SUM(subtotal), 0)::BIGINT AS revenue
			 FROM checks
			 WHERE location_id = $1 AND status = 'closed'
			   AND closed_at >= $2 AND closed_at < $3
			 GROUP BY channel ORDER BY orders DESC`,
			locationID, from, to,
		)
		if err != nil {
			return fmt.Errorf("query channel performance: %w", err)
		}
		defer rows.Close()

		var channels []ChannelPerf
		for rows.Next() {
			var cp ChannelPerf
			if err := rows.Scan(&cp.Channel, &cp.Orders, &cp.AvgTicketTime, &cp.Revenue); err != nil {
				return fmt.Errorf("scan channel row: %w", err)
			}
			channels = append(channels, cp)
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("iterate channel rows: %w", err)
		}

		// Compute pct_of_total.
		if closedCount > 0 {
			for i := range channels {
				channels[i].PctOfTotal = float64(channels[i].Orders) / float64(closedCount) * 100
			}
		}
		if channels == nil {
			channels = []ChannelPerf{}
		}

		summary = OperationsSummary{
			OrdersToday:        closedCount,
			AvgTicketTime:      avgTicketMin,
			OrdersPerHour:      ordersPerHour,
			ActiveTickets:      activeCount,
			LongestOpenMin:     longestMin,
			RevenuePerHour:     revenuePerHour,
			VoidRate:           voidRate,
			ChannelPerformance: channels,
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &summary, nil
}

// GetHourly returns order counts and revenue aggregated by hour for the given period.
func (s *Service) GetHourly(ctx context.Context, orgID, locationID string, from, to time.Time) ([]HourlyData, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	var results []HourlyData

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(tenantCtx,
			`SELECT EXTRACT(HOUR FROM closed_at)::INT AS hour,
			        COUNT(*) AS orders,
			        COALESCE(SUM(subtotal), 0)::BIGINT AS revenue
			 FROM checks
			 WHERE location_id = $1 AND status = 'closed'
			   AND closed_at >= $2 AND closed_at < $3
			 GROUP BY EXTRACT(HOUR FROM closed_at)
			 ORDER BY hour`,
			locationID, from, to,
		)
		if err != nil {
			return fmt.Errorf("query hourly data: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var hd HourlyData
			if err := rows.Scan(&hd.Hour, &hd.Orders, &hd.Revenue); err != nil {
				return fmt.Errorf("scan hourly row: %w", err)
			}
			results = append(results, hd)
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("iterate hourly rows: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	if results == nil {
		results = []HourlyData{}
	}

	return results, nil
}
