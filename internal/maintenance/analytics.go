package maintenance

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/opsnerve/fireline/internal/tenant"
	"github.com/opsnerve/fireline/pkg/database"
)

// MaintenanceStats holds aggregate maintenance analytics.
type MaintenanceStats struct {
	OpenTickets       int     `json:"open_tickets"`
	InProgressTickets int     `json:"in_progress_tickets"`
	OverdueCount      int     `json:"overdue_count"`
	TotalCostThisMonth int    `json:"total_cost_this_month"`
	AvgResolutionHours float64 `json:"avg_resolution_hours"`
	TotalEquipment    int     `json:"total_equipment"`
	OperationalCount  int     `json:"operational_count"`
	NeedsMaintenanceCount int `json:"needs_maintenance_count"`
	OutOfServiceCount int     `json:"out_of_service_count"`
	AvgHealthScore    float64 `json:"avg_health_score"`
	TicketsByType     []TypeCount     `json:"tickets_by_type"`
	HealthDistribution []HealthBucket `json:"health_distribution"`
}

// TypeCount represents ticket count by type.
type TypeCount struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
}

// HealthBucket represents equipment count by health score range.
type HealthBucket struct {
	Range string `json:"range"`
	Count int    `json:"count"`
}

// GetMaintenanceStats returns aggregate maintenance analytics.
func (s *Service) GetMaintenanceStats(ctx context.Context, orgID, locationID string) (*MaintenanceStats, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	stats := &MaintenanceStats{}

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		locFilter := ""
		args := []any{}
		argIdx := 1
		if locationID != "" {
			locFilter = fmt.Sprintf(" AND location_id = $%d", argIdx)
			args = append(args, locationID)
			argIdx++
		}

		// Open and in-progress tickets
		if err := tx.QueryRow(tenantCtx,
			fmt.Sprintf("SELECT COUNT(*) FILTER (WHERE status = 'open'), COUNT(*) FILTER (WHERE status = 'in_progress') FROM maintenance_tickets WHERE 1=1%s", locFilter),
			args...,
		).Scan(&stats.OpenTickets, &stats.InProgressTickets); err != nil {
			return err
		}

		// Overdue equipment count
		if err := tx.QueryRow(tenantCtx,
			fmt.Sprintf("SELECT COUNT(*) FROM equipment WHERE next_maintenance < CURRENT_DATE%s", locFilter),
			args...,
		).Scan(&stats.OverdueCount); err != nil {
			return err
		}

		// Total cost this month
		if err := tx.QueryRow(tenantCtx,
			fmt.Sprintf("SELECT COALESCE(SUM(actual_cost), 0) FROM maintenance_tickets WHERE completed_at >= date_trunc('month', CURRENT_DATE)%s", locFilter),
			args...,
		).Scan(&stats.TotalCostThisMonth); err != nil {
			return err
		}

		// Average resolution time in hours
		if err := tx.QueryRow(tenantCtx,
			fmt.Sprintf(`SELECT COALESCE(AVG(EXTRACT(EPOCH FROM (completed_at - started_at)) / 3600), 0)
				FROM maintenance_tickets WHERE status = 'completed' AND started_at IS NOT NULL AND completed_at IS NOT NULL%s`, locFilter),
			args...,
		).Scan(&stats.AvgResolutionHours); err != nil {
			return err
		}

		// Equipment counts
		if err := tx.QueryRow(tenantCtx,
			fmt.Sprintf(`SELECT
				COUNT(*),
				COUNT(*) FILTER (WHERE status = 'operational'),
				COUNT(*) FILTER (WHERE status = 'needs_maintenance'),
				COUNT(*) FILTER (WHERE status IN ('out_of_service', 'under_repair')),
				COALESCE(AVG(health_score), 0)
			 FROM equipment WHERE 1=1%s`, locFilter),
			args...,
		).Scan(&stats.TotalEquipment, &stats.OperationalCount, &stats.NeedsMaintenanceCount, &stats.OutOfServiceCount, &stats.AvgHealthScore); err != nil {
			return err
		}

		// Tickets by type
		rows, err := tx.Query(tenantCtx,
			fmt.Sprintf("SELECT type, COUNT(*) FROM maintenance_tickets WHERE 1=1%s GROUP BY type ORDER BY COUNT(*) DESC", locFilter),
			args...,
		)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var tc TypeCount
			if err := rows.Scan(&tc.Type, &tc.Count); err != nil {
				return err
			}
			stats.TicketsByType = append(stats.TicketsByType, tc)
		}
		if err := rows.Err(); err != nil {
			return err
		}

		// Health distribution
		rows2, err := tx.Query(tenantCtx,
			fmt.Sprintf(`SELECT
				CASE
					WHEN health_score >= 80 THEN 'Good (80-100)'
					WHEN health_score >= 50 THEN 'Fair (50-79)'
					ELSE 'Poor (0-49)'
				END AS range,
				COUNT(*)
			 FROM equipment WHERE 1=1%s
			 GROUP BY range ORDER BY range`, locFilter),
			args...,
		)
		if err != nil {
			return err
		}
		defer rows2.Close()
		for rows2.Next() {
			var hb HealthBucket
			if err := rows2.Scan(&hb.Range, &hb.Count); err != nil {
				return err
			}
			stats.HealthDistribution = append(stats.HealthDistribution, hb)
		}
		return rows2.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("get maintenance stats: %w", err)
	}
	return stats, nil
}
