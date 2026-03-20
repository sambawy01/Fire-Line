package labor

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/opsnerve/fireline/internal/event"
	"github.com/opsnerve/fireline/internal/tenant"
	"github.com/opsnerve/fireline/pkg/database"
)

// PointEvent records a single staff point award or deduction.
type PointEvent struct {
	EventID     string    `json:"event_id"`
	EmployeeID  string    `json:"employee_id"`
	Points      float64   `json:"points"`
	Reason      string    `json:"reason"`
	Description string    `json:"description"`
	ShiftID     *string   `json:"shift_id,omitempty"`
	AwardedBy   *string   `json:"awarded_by,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// LeaderboardEntry holds a summarised staff-points ranking for one employee.
type LeaderboardEntry struct {
	EmployeeID  string  `json:"employee_id"`
	DisplayName string  `json:"display_name"`
	Role        string  `json:"role"`
	StaffPoints float64 `json:"staff_points"`
	PointsTrend string  `json:"points_trend"`
}

// AwardPoints inserts a point event, updates the employee's running total,
// and publishes a "labor.points.awarded" event on the bus.
func (s *Service) AwardPoints(
	ctx context.Context,
	orgID, employeeID string,
	points float64,
	reason, description string,
	shiftID, awardedBy *string,
) (*PointEvent, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	var pe PointEvent
	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		err := tx.QueryRow(tenantCtx,
			`INSERT INTO staff_point_events
			    (org_id, employee_id, points, reason, description, shift_id, awarded_by)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)
			 RETURNING event_id::TEXT, employee_id::TEXT, points, reason,
			           COALESCE(description, ''), shift_id::TEXT, awarded_by::TEXT, created_at`,
			orgID, employeeID, points, reason, description, shiftID, awardedBy,
		).Scan(
			&pe.EventID,
			&pe.EmployeeID,
			&pe.Points,
			&pe.Reason,
			&pe.Description,
			&pe.ShiftID,
			&pe.AwardedBy,
			&pe.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("insert point event: %w", err)
		}

		_, err = tx.Exec(tenantCtx,
			`UPDATE employees SET staff_points = staff_points + $1 WHERE employee_id = $2`,
			points, employeeID,
		)
		if err != nil {
			return fmt.Errorf("update staff_points: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	s.bus.Publish(ctx, event.Envelope{
		EventType: "labor.points.awarded",
		OrgID:     orgID,
		Source:    "labor",
		Payload: map[string]any{
			"event_id":    pe.EventID,
			"employee_id": pe.EmployeeID,
			"points":      pe.Points,
			"reason":      pe.Reason,
		},
	})

	return &pe, nil
}

// GetPointHistory returns up to limit recent point events for an employee,
// ordered newest first.
func (s *Service) GetPointHistory(ctx context.Context, orgID, employeeID string, limit int) ([]PointEvent, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	if limit <= 0 {
		limit = 50
	}

	var events []PointEvent
	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(tenantCtx,
			`SELECT
			    event_id::TEXT,
			    employee_id::TEXT,
			    points,
			    reason,
			    COALESCE(description, ''),
			    shift_id::TEXT,
			    awarded_by::TEXT,
			    created_at
			FROM staff_point_events
			WHERE employee_id = $1
			ORDER BY created_at DESC
			LIMIT $2`,
			employeeID, limit,
		)
		if err != nil {
			return fmt.Errorf("query point history: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var pe PointEvent
			if err := rows.Scan(
				&pe.EventID,
				&pe.EmployeeID,
				&pe.Points,
				&pe.Reason,
				&pe.Description,
				&pe.ShiftID,
				&pe.AwardedBy,
				&pe.CreatedAt,
			); err != nil {
				return fmt.Errorf("scan point event: %w", err)
			}
			events = append(events, pe)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	if events == nil {
		events = []PointEvent{}
	}
	return events, nil
}

// GetLeaderboard returns up to limit employees ordered by staff_points descending.
// locationID is optional; pass empty string to include all locations in the org.
func (s *Service) GetLeaderboard(ctx context.Context, orgID, locationID string, limit int) ([]LeaderboardEntry, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	if limit <= 0 {
		limit = 25
	}

	var entries []LeaderboardEntry
	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		var (
			rows pgx.Rows
			err  error
		)
		if locationID != "" {
			rows, err = tx.Query(tenantCtx,
				`SELECT
				    e.employee_id::TEXT,
				    e.display_name,
				    e.role,
				    e.staff_points
				FROM employees e
				WHERE e.location_id = $1
				  AND e.status = 'active'
				ORDER BY e.staff_points DESC
				LIMIT $2`,
				locationID, limit,
			)
		} else {
			rows, err = tx.Query(tenantCtx,
				`SELECT
				    e.employee_id::TEXT,
				    e.display_name,
				    e.role,
				    e.staff_points
				FROM employees e
				WHERE e.status = 'active'
				ORDER BY e.staff_points DESC
				LIMIT $1`,
				limit,
			)
		}
		if err != nil {
			return fmt.Errorf("query leaderboard: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var entry LeaderboardEntry
			var staffPoints float64
			if err := rows.Scan(
				&entry.EmployeeID,
				&entry.DisplayName,
				&entry.Role,
				&staffPoints,
			); err != nil {
				return fmt.Errorf("scan leaderboard row: %w", err)
			}
			entry.StaffPoints = staffPoints
			entries = append(entries, entry)
		}
		if err := rows.Err(); err != nil {
			return err
		}
		rows.Close()

		// Compute trends in a second pass (avoids conn busy)
		for i, entry := range entries {
			var sevenDaysAgo float64
			if err := tx.QueryRow(tenantCtx,
				`SELECT COALESCE(SUM(points), 0)
				 FROM staff_point_events
				 WHERE employee_id = $1
				   AND created_at < now() - INTERVAL '7 days'`,
				entry.EmployeeID,
			).Scan(&sevenDaysAgo); err != nil {
				entries[i].PointsTrend = "stable"
				continue
			}
			entries[i].PointsTrend = computePointsTrend(entry.StaffPoints, sevenDaysAgo)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if entries == nil {
		entries = []LeaderboardEntry{}
	}
	return entries, nil
}
