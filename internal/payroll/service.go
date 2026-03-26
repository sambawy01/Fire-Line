package payroll

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/opsnerve/fireline/internal/event"
	"github.com/opsnerve/fireline/internal/tenant"
	"github.com/opsnerve/fireline/pkg/database"
)

// Service provides payroll reporting capabilities.
type Service struct {
	pool *pgxpool.Pool
	bus  *event.Bus
}

// New creates a new payroll service.
func New(pool *pgxpool.Pool, bus *event.Bus) *Service {
	return &Service{pool: pool, bus: bus}
}

// ─── Structs ────────────────────────────────────────────────────────────────

// PayrollSummary is the top-level payroll report for a period.
type PayrollSummary struct {
	PeriodStart   string            `json:"period_start"`
	PeriodEnd     string            `json:"period_end"`
	LocationID    string            `json:"location_id"`
	TotalGrossPay int64             `json:"total_gross_pay"`
	TotalOvertime int64             `json:"total_overtime_pay"`
	TotalHours    float64           `json:"total_hours"`
	EmployeeCount int               `json:"employee_count"`
	Employees     []EmployeePayroll `json:"employees"`
}

// EmployeePayroll holds per-employee payroll data for a period.
type EmployeePayroll struct {
	EmployeeID    string  `json:"employee_id"`
	DisplayName   string  `json:"display_name"`
	Role          string  `json:"role"`
	TotalHours    float64 `json:"total_hours"`
	RegularHours  float64 `json:"regular_hours"`
	OvertimeHours float64 `json:"overtime_hours"`
	HourlyRate    int64   `json:"hourly_rate"`
	RegularPay    int64   `json:"regular_pay"`
	OvertimePay   int64   `json:"overtime_pay"`
	GrossPay      int64   `json:"gross_pay"`
	ShiftCount    int     `json:"shift_count"`
}

// PayrollPeriod holds aggregated monthly payroll data.
type PayrollPeriod struct {
	Month         string  `json:"month"`
	TotalGrossPay int64   `json:"total_gross_pay"`
	TotalHours    float64 `json:"total_hours"`
	EmployeeCount int     `json:"employee_count"`
	LaborCostPct  float64 `json:"labor_cost_pct"`
}

// shiftRow is an intermediate row scanned from the shifts+employees join.
type shiftRow struct {
	employeeID  string
	displayName string
	role        string
	clockIn     time.Time
	clockOut    time.Time
	hourlyRate  int64
}

// ─── GetPayrollSummary ──────────────────────────────────────────────────────

// GetPayrollSummary computes per-employee payroll for a date range at a location.
func (s *Service) GetPayrollSummary(ctx context.Context, orgID, locationID, periodStart, periodEnd string) (*PayrollSummary, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	summary := &PayrollSummary{
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		LocationID:  locationID,
		Employees:   []EmployeePayroll{},
	}

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(tenantCtx,
			`SELECT s.employee_id,
				COALESCE(e.display_name, 'Unknown'),
				COALESCE(e.role, s.role),
				s.clock_in,
				s.clock_out,
				s.hourly_rate
			FROM shifts s
			LEFT JOIN employees e ON e.employee_id = s.employee_id
			WHERE s.location_id = $1
				AND s.clock_in >= $2::TIMESTAMPTZ
				AND s.clock_in < $3::TIMESTAMPTZ
				AND s.clock_out IS NOT NULL
				AND s.status = 'completed'
			ORDER BY s.employee_id, s.clock_in`,
			locationID, periodStart, periodEnd,
		)
		if err != nil {
			return fmt.Errorf("shifts query: %w", err)
		}
		defer rows.Close()

		var shifts []shiftRow
		for rows.Next() {
			var sr shiftRow
			if err := rows.Scan(&sr.employeeID, &sr.displayName, &sr.role, &sr.clockIn, &sr.clockOut, &sr.hourlyRate); err != nil {
				return err
			}
			shifts = append(shifts, sr)
		}
		if err := rows.Err(); err != nil {
			return err
		}

		// Group shifts by employee and compute payroll.
		empMap := make(map[string]*employeeAccumulator)
		var empOrder []string
		for _, sr := range shifts {
			acc, ok := empMap[sr.employeeID]
			if !ok {
				acc = &employeeAccumulator{
					employeeID:  sr.employeeID,
					displayName: sr.displayName,
					role:        sr.role,
					hourlyRate:  sr.hourlyRate,
				}
				empMap[sr.employeeID] = acc
				empOrder = append(empOrder, sr.employeeID)
			}
			acc.shifts = append(acc.shifts, sr)
		}

		for _, eid := range empOrder {
			acc := empMap[eid]
			ep := acc.compute(periodStart, periodEnd)
			summary.Employees = append(summary.Employees, ep)
			summary.TotalGrossPay += ep.GrossPay
			summary.TotalOvertime += ep.OvertimePay
			summary.TotalHours += ep.TotalHours
		}
		summary.EmployeeCount = len(empOrder)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("get payroll summary: %w", err)
	}
	return summary, nil
}

// employeeAccumulator collects shifts for a single employee and computes payroll.
type employeeAccumulator struct {
	employeeID  string
	displayName string
	role        string
	hourlyRate  int64
	shifts      []shiftRow
}

// compute calculates overtime using both daily (>8h) and weekly (>40h) rules.
func (a *employeeAccumulator) compute(periodStart, periodEnd string) EmployeePayroll {
	ep := EmployeePayroll{
		EmployeeID:  a.employeeID,
		DisplayName: a.displayName,
		Role:        a.role,
		HourlyRate:  a.hourlyRate,
		ShiftCount:  len(a.shifts),
	}

	// Accumulate hours per day and per ISO week.
	type dayHours struct {
		total float64
	}
	type weekHours struct {
		total float64
	}

	dailyHours := make(map[string]float64)  // "2006-01-02" -> hours
	weeklyHours := make(map[string]float64) // "2006-W02" -> hours

	for _, sr := range a.shifts {
		hours := sr.clockOut.Sub(sr.clockIn).Hours()
		if hours <= 0 {
			continue
		}
		ep.TotalHours += hours

		dayKey := sr.clockIn.Format("2006-01-02")
		dailyHours[dayKey] += hours

		yr, wk := sr.clockIn.ISOWeek()
		weekKey := fmt.Sprintf("%d-W%02d", yr, wk)
		weeklyHours[weekKey] += hours
	}

	// Calculate daily overtime: hours beyond 8 per day.
	var dailyOT float64
	for _, h := range dailyHours {
		if h > 8 {
			dailyOT += h - 8
		}
	}

	// Calculate weekly overtime: hours beyond 40 per week.
	var weeklyOT float64
	for _, h := range weeklyHours {
		if h > 40 {
			weeklyOT += h - 40
		}
	}

	// Use the greater of daily or weekly overtime.
	ep.OvertimeHours = math.Max(dailyOT, weeklyOT)
	ep.OvertimeHours = math.Round(ep.OvertimeHours*100) / 100
	ep.RegularHours = ep.TotalHours - ep.OvertimeHours
	if ep.RegularHours < 0 {
		ep.RegularHours = 0
	}
	ep.RegularHours = math.Round(ep.RegularHours*100) / 100
	ep.TotalHours = math.Round(ep.TotalHours*100) / 100

	// Pay calculations (hourly_rate is in cents/piasters).
	ep.RegularPay = int64(math.Round(ep.RegularHours * float64(a.hourlyRate)))
	ep.OvertimePay = int64(math.Round(ep.OvertimeHours * float64(a.hourlyRate) * 1.5))
	ep.GrossPay = ep.RegularPay + ep.OvertimePay

	return ep
}

// ─── GetPayrollHistory ──────────────────────────────────────────────────────

// GetPayrollHistory returns monthly payroll aggregates for the last N months.
func (s *Service) GetPayrollHistory(ctx context.Context, orgID, locationID string, months int) ([]PayrollPeriod, error) {
	if months <= 0 {
		months = 6
	}
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var results []PayrollPeriod

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(tenantCtx,
			`SELECT
				TO_CHAR(s.clock_in, 'YYYY-MM') AS month,
				SUM(
					EXTRACT(EPOCH FROM (s.clock_out - s.clock_in)) / 3600.0 * s.hourly_rate
				)::BIGINT AS total_gross_pay,
				SUM(
					EXTRACT(EPOCH FROM (s.clock_out - s.clock_in)) / 3600.0
				) AS total_hours,
				COUNT(DISTINCT s.employee_id)::INT AS employee_count
			FROM shifts s
			WHERE s.location_id = $1
				AND s.clock_out IS NOT NULL
				AND s.status = 'completed'
				AND s.clock_in >= DATE_TRUNC('month', NOW()) - make_interval(months => $2)
			GROUP BY TO_CHAR(s.clock_in, 'YYYY-MM')
			ORDER BY month DESC`,
			locationID, months,
		)
		if err != nil {
			return fmt.Errorf("payroll history query: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var pp PayrollPeriod
			if err := rows.Scan(&pp.Month, &pp.TotalGrossPay, &pp.TotalHours, &pp.EmployeeCount); err != nil {
				return err
			}
			results = append(results, pp)
		}
		if err := rows.Err(); err != nil {
			return err
		}

		// Calculate labor cost percentage vs revenue for each month.
		for i := range results {
			var revenue int64
			err := tx.QueryRow(tenantCtx,
				`SELECT COALESCE(SUM(total), 0)
				FROM checks
				WHERE location_id = $1
					AND TO_CHAR(created_at, 'YYYY-MM') = $2
					AND status IN ('closed', 'paid')`,
				locationID, results[i].Month,
			).Scan(&revenue)
			if err != nil {
				return fmt.Errorf("revenue query for %s: %w", results[i].Month, err)
			}
			if revenue > 0 {
				results[i].LaborCostPct = math.Round(float64(results[i].TotalGrossPay)/float64(revenue)*10000) / 100
			}
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("get payroll history: %w", err)
	}
	if results == nil {
		results = []PayrollPeriod{}
	}
	return results, nil
}

// ─── ExportPayroll ──────────────────────────────────────────────────────────

// ExportPayroll generates a CSV of employee payroll data for the given period.
func (s *Service) ExportPayroll(ctx context.Context, orgID, locationID, periodStart, periodEnd string) ([]byte, error) {
	summary, err := s.GetPayrollSummary(ctx, orgID, locationID, periodStart, periodEnd)
	if err != nil {
		return nil, fmt.Errorf("export payroll: %w", err)
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	// Header row
	if err := w.Write([]string{
		"Employee ID", "Display Name", "Role",
		"Total Hours", "Regular Hours", "Overtime Hours",
		"Hourly Rate (cents)", "Regular Pay (cents)", "Overtime Pay (cents)", "Gross Pay (cents)",
		"Shift Count",
	}); err != nil {
		return nil, fmt.Errorf("csv header: %w", err)
	}

	for _, ep := range summary.Employees {
		if err := w.Write([]string{
			ep.EmployeeID,
			ep.DisplayName,
			ep.Role,
			strconv.FormatFloat(ep.TotalHours, 'f', 2, 64),
			strconv.FormatFloat(ep.RegularHours, 'f', 2, 64),
			strconv.FormatFloat(ep.OvertimeHours, 'f', 2, 64),
			strconv.FormatInt(ep.HourlyRate, 10),
			strconv.FormatInt(ep.RegularPay, 10),
			strconv.FormatInt(ep.OvertimePay, 10),
			strconv.FormatInt(ep.GrossPay, 10),
			strconv.Itoa(ep.ShiftCount),
		}); err != nil {
			return nil, fmt.Errorf("csv row: %w", err)
		}
	}

	// Totals row
	if err := w.Write([]string{
		"", "TOTALS", "",
		strconv.FormatFloat(summary.TotalHours, 'f', 2, 64),
		"", "",
		"",
		"",
		strconv.FormatInt(summary.TotalOvertime, 10),
		strconv.FormatInt(summary.TotalGrossPay, 10),
		strconv.Itoa(summary.EmployeeCount) + " employees",
	}); err != nil {
		return nil, fmt.Errorf("csv totals: %w", err)
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, fmt.Errorf("csv flush: %w", err)
	}

	return buf.Bytes(), nil
}
