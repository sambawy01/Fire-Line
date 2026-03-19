package labor

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

// Service provides labor intelligence capabilities.
type Service struct {
	pool *pgxpool.Pool
	bus  *event.Bus
}

// New creates a new labor intelligence service.
func New(pool *pgxpool.Pool, bus *event.Bus) *Service {
	return &Service{pool: pool, bus: bus}
}

// LaborSummary holds location-wide labor cost KPIs.
type LaborSummary struct {
	TotalLaborCost int64   `json:"total_labor_cost"`
	LaborCostPct   float64 `json:"labor_cost_pct"`
	NetRevenue     int64   `json:"net_revenue"`
	EmployeeCount  int     `json:"employee_count"`
	TotalHours     float64 `json:"total_hours"`
	TotalShifts    int     `json:"total_shifts"`
}

// EmployeeDetail holds per-employee labor cost and shift metrics.
type EmployeeDetail struct {
	EmployeeID       string  `json:"employee_id"`
	DisplayName      string  `json:"display_name"`
	Role             string  `json:"role"`
	Status           string  `json:"status"`
	ShiftCount       int     `json:"shift_count"`
	HoursWorked      float64 `json:"hours_worked"`
	LaborCost        int64   `json:"labor_cost"`
	AvgHoursPerShift float64 `json:"avg_hours_per_shift"`
	HourlyRate       int64   `json:"hourly_rate"`
}

// GetEmployees queries active employees and their shift data for the given
// location and date range. Open shifts are capped at 16 hours from clock_in
// to prevent forgotten clock-outs from inflating labor cost.
func (s *Service) GetEmployees(ctx context.Context, orgID, locationID string, from, to time.Time) ([]EmployeeDetail, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	var employees []EmployeeDetail

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		// LEFT JOIN shifts so employees with zero shifts still appear.
		// Ghost shift cap: clock_out NULLs are bounded to clock_in + 16h to
		// prevent runaway costs from forgotten clock-outs.
		rows, err := tx.Query(tenantCtx,
			`SELECT
			    e.employee_id::TEXT,
			    e.display_name,
			    e.role,
			    e.status,
			    COUNT(s.shift_id)::INT AS shift_count,
			    COALESCE(
			        SUM(
			            EXTRACT(EPOCH FROM (
			                COALESCE(s.clock_out, LEAST(now(), s.clock_in + INTERVAL '16 hours')) - s.clock_in
			            )) / 3600.0
			        ), 0
			    ) AS hours_worked,
			    COALESCE(
			        SUM(
			            EXTRACT(EPOCH FROM (
			                COALESCE(s.clock_out, LEAST(now(), s.clock_in + INTERVAL '16 hours')) - s.clock_in
			            )) / 3600.0
			            * s.hourly_rate
			        )::BIGINT, 0
			    ) AS labor_cost,
			    COALESCE(MAX(s.hourly_rate), 0)::BIGINT AS hourly_rate
			FROM employees e
			LEFT JOIN shifts s
			    ON s.employee_id = e.employee_id
			    AND s.org_id = e.org_id
			    AND s.location_id = $1
			    AND s.status != 'no_show'
			    AND s.clock_in >= $2
			    AND s.clock_in < $3
			WHERE e.location_id = $1
			  AND e.status = 'active'
			GROUP BY e.employee_id, e.display_name, e.role, e.status
			ORDER BY e.display_name`,
			locationID, from, to,
		)
		if err != nil {
			return fmt.Errorf("query employees: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var emp EmployeeDetail
			var shiftCount int
			var hoursWorked float64
			var laborCost int64
			var hourlyRate int64

			if err := rows.Scan(
				&emp.EmployeeID,
				&emp.DisplayName,
				&emp.Role,
				&emp.Status,
				&shiftCount,
				&hoursWorked,
				&laborCost,
				&hourlyRate,
			); err != nil {
				return fmt.Errorf("scan employee row: %w", err)
			}

			emp.ShiftCount = shiftCount
			emp.HoursWorked = hoursWorked
			emp.LaborCost = laborCost
			emp.HourlyRate = hourlyRate

			// Guard div-by-zero: avg hours per shift.
			if shiftCount > 0 {
				emp.AvgHoursPerShift = hoursWorked / float64(shiftCount)
			}

			employees = append(employees, emp)
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("iterate employee rows: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	if employees == nil {
		employees = []EmployeeDetail{}
	}

	return employees, nil
}

// GetSummary aggregates employee labor data and net revenue into location-wide
// KPIs for the given date range.
func (s *Service) GetSummary(ctx context.Context, orgID, locationID string, from, to time.Time) (*LaborSummary, error) {
	employees, err := s.GetEmployees(ctx, orgID, locationID, from, to)
	if err != nil {
		return nil, fmt.Errorf("get summary: %w", err)
	}

	// Aggregate employee totals.
	summary := &LaborSummary{
		EmployeeCount: len(employees),
	}
	for _, emp := range employees {
		summary.TotalLaborCost += emp.LaborCost
		summary.TotalHours += emp.HoursWorked
		summary.TotalShifts += emp.ShiftCount
	}

	// Query net revenue from closed checks in the same period.
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	err = database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		return tx.QueryRow(tenantCtx,
			`SELECT COALESCE(SUM(subtotal), 0)::BIGINT
			 FROM checks
			 WHERE location_id = $1
			   AND status = 'closed'
			   AND closed_at >= $2
			   AND closed_at < $3`,
			locationID, from, to,
		).Scan(&summary.NetRevenue)
	})
	if err != nil {
		return nil, fmt.Errorf("query net revenue: %w", err)
	}

	// Guard division by zero: return 0.0 if no revenue.
	if summary.NetRevenue > 0 {
		summary.LaborCostPct = float64(summary.TotalLaborCost) / float64(summary.NetRevenue) * 100
	}

	return summary, nil
}
