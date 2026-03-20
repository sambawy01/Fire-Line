package labor

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

// Schedule holds header metadata for a weekly schedule.
type Schedule struct {
	ScheduleID  string     `json:"schedule_id"`
	LocationID  string     `json:"location_id"`
	WeekStart   string     `json:"week_start"`
	Status      string     `json:"status"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	CreatedBy   *string    `json:"created_by,omitempty"`
	ShiftCount  int        `json:"shift_count,omitempty"`
}

// ScheduledShift holds a single scheduled shift record.
type ScheduledShift struct {
	ShiftID      string `json:"scheduled_shift_id"`
	EmployeeID   string `json:"employee_id"`
	EmployeeName string `json:"employee_name,omitempty"`
	ShiftDate    string `json:"shift_date"`
	StartTime    string `json:"start_time"`
	EndTime      string `json:"end_time"`
	Station      string `json:"station"`
	Status       string `json:"status"`
	Notes        string `json:"notes"`
}

// ScheduleWithShifts bundles a schedule header with its assigned shifts.
type ScheduleWithShifts struct {
	Schedule
	Shifts []ScheduledShift `json:"shifts"`
}

// LaborCostProjection holds projected cost metrics for a schedule.
type LaborCostProjection struct {
	TotalHours   float64 `json:"total_hours"`
	TotalCost    int64   `json:"total_cost"`        // cents
	LaborCostPct float64 `json:"labor_cost_pct"`
	BudgetTarget float64 `json:"budget_target_pct"`
	OverUnder    string  `json:"over_under"` // "on_track", "over", "under"
}

// OvertimeRisk flags an employee who is close to or over 40 scheduled hours.
type OvertimeRisk struct {
	EmployeeID     string  `json:"employee_id"`
	EmployeeName   string  `json:"employee_name"`
	ScheduledHours float64 `json:"scheduled_hours"`
	Severity       string  `json:"severity"` // "warning" (>38), "critical" (>40)
}

// classifyOvertimeRisk returns the severity string for a given scheduled-hours
// value. >40 => "critical", >38 => "warning", otherwise "".
// This is a pure function exposed for unit testing.
func classifyOvertimeRisk(hours float64) string {
	switch {
	case hours > 40:
		return "critical"
	case hours > 38:
		return "warning"
	default:
		return ""
	}
}

// GenerateScheduleDraft creates a draft schedule for the week starting on
// weekStart (YYYY-MM-DD) by running a greedy forecast-driven shift assignment.
//
// Algorithm per day:
//  1. Generate (or retrieve existing) forecast blocks.
//  2. Load active employees for the location.
//  3. Greedily assign employees to contiguous coverage windows, respecting:
//     - max 8 h/shift, max 40 h/week, min 8 h gap between shifts, availability.
//  4. Insert all shifts and return the full schedule.
func (s *Service) GenerateScheduleDraft(ctx context.Context, orgID, locationID, weekStart, createdBy string) (*ScheduleWithShifts, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	weekDate, err := time.Parse("2006-01-02", weekStart)
	if err != nil {
		return nil, fmt.Errorf("invalid week_start %q: %w", weekStart, err)
	}

	var scheduleID string

	// Create the schedule header inside a transaction.
	err = database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		var cb *string
		if createdBy != "" {
			cb = &createdBy
		}
		return tx.QueryRow(tenantCtx,
			`INSERT INTO schedules (org_id, location_id, week_start, status, created_by)
			 VALUES ($1, $2, $3, 'draft', $4)
			 ON CONFLICT (org_id, location_id, week_start)
			 DO UPDATE SET status = 'draft', updated_at = now()
			 RETURNING schedule_id::TEXT`,
			orgID, locationID, weekStart, cb,
		).Scan(&scheduleID)
	})
	if err != nil {
		return nil, fmt.Errorf("create schedule: %w", err)
	}

	// Load employees once for the whole week.
	profiles, err := s.ListEmployeeProfiles(ctx, orgID, locationID)
	if err != nil {
		return nil, fmt.Errorf("load employee profiles: %w", err)
	}

	// Per-employee weekly hour accumulator (employeeID -> hours).
	weeklyHours := make(map[string]float64, len(profiles))
	// Per-employee last shift end datetime (for min-8h-gap enforcement).
	lastShiftEnd := make(map[string]time.Time, len(profiles))

	var allShifts []ScheduledShift

	// Iterate Mon (0) through Sun (6).
	for dayOffset := 0; dayOffset < 7; dayOffset++ {
		shiftDate := weekDate.AddDate(0, 0, dayOffset)
		dateStr := shiftDate.Format("2006-01-02")
		// Day-of-week name for availability lookup.
		dowName := shiftDate.Weekday().String() // "Monday", "Tuesday", …

		// Generate / retrieve forecast blocks for this day.
		blocks, err := s.GenerateForecast(ctx, orgID, locationID, shiftDate)
		if err != nil {
			return nil, fmt.Errorf("forecast for %s: %w", dateStr, err)
		}

		// Greedy assignment: build shifts block by block.
		// We accumulate a "window" of contiguous blocks, flushing when:
		//   - window reaches 8 h (max shift duration), or
		//   - coverage demand drops to 0 (break in demand).
		//
		// employeeID assigned to current window.
		type window struct {
			employeeID string
			startBlock string // "HH:MM"
			endBlock   string // exclusive end block (start of next slot)
			hours      float64
		}
		var openWindows []window

		flushWindow := func(w window) {
			if w.employeeID == "" || w.hours <= 0 {
				return
			}
			// endTime is w.endBlock (the block after the last covered slot,
			// which is startBlock + accumulated 30-min increments).
			shift := ScheduledShift{
				EmployeeID: w.employeeID,
				ShiftDate:  dateStr,
				StartTime:  w.startBlock,
				EndTime:    w.endBlock,
				Station:    "floor",
				Status:     "scheduled",
			}
			allShifts = append(allShifts, shift)
			weeklyHours[w.employeeID] += w.hours
			// Record end datetime for gap enforcement.
			endDT, _ := time.Parse("2006-01-02 15:04", dateStr+" "+w.endBlock)
			lastShiftEnd[w.employeeID] = endDT
		}

		// Assign each block greedily.
		for _, block := range blocks {
			if block.RequiredHeadcount == 0 {
				// Flush any open windows.
				for _, w := range openWindows {
					flushWindow(w)
				}
				openWindows = nil
				continue
			}

			// Convert block time to shift datetime for gap checking.
			blockDT, _ := time.Parse("2006-01-02 15:04", dateStr+" "+block.TimeBlock)
			blockEndDT := blockDT.Add(30 * time.Minute)

			// Count how many open windows cover this block.
			needed := block.RequiredHeadcount - len(openWindows)

			// Flush over-staffed windows that hit 8h.
			var stillOpen []window
			for _, w := range openWindows {
				if w.hours >= 8.0 {
					flushWindow(w)
				} else {
					// Extend the window.
					w.hours += 0.5
					w.endBlock = blockEndDT.Format("15:04")
					stillOpen = append(stillOpen, w)
				}
			}
			openWindows = stillOpen

			needed = block.RequiredHeadcount - len(openWindows)

			// Assign employees to fill remaining need.
			for i := 0; i < needed; i++ {
				emp := s.pickEmployee(profiles, weeklyHours, lastShiftEnd, dowName, blockDT, block.TimeBlock)
				if emp == nil {
					break // no available employee
				}
				openWindows = append(openWindows, window{
					employeeID: emp.EmployeeID,
					startBlock: block.TimeBlock,
					endBlock:   blockEndDT.Format("15:04"),
					hours:      0.5,
				})
			}
		}

		// Flush remaining open windows at end of day.
		for _, w := range openWindows {
			flushWindow(w)
		}
		openWindows = nil
	}

	// Persist all generated shifts in a single transaction.
	if len(allShifts) > 0 {
		err = database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
			for i, sh := range allShifts {
				var shiftID string
				err := tx.QueryRow(tenantCtx,
					`INSERT INTO scheduled_shifts
					    (org_id, schedule_id, employee_id, shift_date,
					     start_time, end_time, station, status, notes)
					 VALUES ($1, $2, $3, $4, $5::TIME, $6::TIME, $7, 'scheduled', '')
					 RETURNING scheduled_shift_id::TEXT`,
					orgID, scheduleID, sh.EmployeeID, sh.ShiftDate,
					sh.StartTime, sh.EndTime, sh.Station,
				).Scan(&shiftID)
				if err != nil {
					return fmt.Errorf("insert shift %d: %w", i, err)
				}
				allShifts[i].ShiftID = shiftID
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("persist shifts: %w", err)
		}
	}

	return s.GetSchedule(ctx, orgID, locationID, weekStart)
}

// pickEmployee selects the best available employee for a block using a greedy
// strategy: prefer employees with fewer hours this week (fairness), filter out
// those who exceed 40 h or violate the min-8h rest gap, and check availability.
func (s *Service) pickEmployee(
	profiles []EmployeeProfile,
	weeklyHours map[string]float64,
	lastShiftEnd map[string]time.Time,
	dowName string,
	blockDT time.Time,
	blockTimeStr string,
) *EmployeeProfile {
	var best *EmployeeProfile
	bestHours := math.MaxFloat64

	for i := range profiles {
		p := &profiles[i]
		if p.Status != "active" {
			continue
		}
		// Max weekly hours guard.
		if weeklyHours[p.EmployeeID]+0.5 > 40 {
			continue
		}
		// Min 8h gap between shifts.
		if last, ok := lastShiftEnd[p.EmployeeID]; ok {
			if blockDT.Sub(last) < 8*time.Hour {
				continue
			}
		}
		// Availability check: availability JSONB is map[day]map[start/end].
		// We accept missing availability as "always available".
		if !employeeAvailable(p.Availability, dowName, blockTimeStr) {
			continue
		}
		// Prefer fewest scheduled hours this week.
		hours := weeklyHours[p.EmployeeID]
		if best == nil || hours < bestHours {
			best = p
			bestHours = hours
		}
	}
	return best
}

// employeeAvailable checks whether the employee's availability map permits
// the given day and time. The availability map is expected to be either:
//
//	{"Monday": {"start": "09:00", "end": "17:00"}, …}  or
//	{"Monday": true, …}  (open all day)
//
// If the day key is absent, the employee is treated as available.
func employeeAvailable(availability map[string]any, dowName, blockTime string) bool {
	val, ok := availability[dowName]
	if !ok {
		return true // no constraint for this day
	}
	switch v := val.(type) {
	case bool:
		return v
	case map[string]any:
		startRaw, hasStart := v["start"]
		endRaw, hasEnd := v["end"]
		if !hasStart || !hasEnd {
			return true
		}
		start, _ := startRaw.(string)
		end, _ := endRaw.(string)
		return blockTime >= start && blockTime < end
	default:
		return true
	}
}

// GetSchedule returns the schedule for a location/week with all shifts joined
// to employee display names.
func (s *Service) GetSchedule(ctx context.Context, orgID, locationID, weekStart string) (*ScheduleWithShifts, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	var sched Schedule

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		err := tx.QueryRow(tenantCtx,
			`SELECT
			    s.schedule_id::TEXT,
			    s.location_id::TEXT,
			    s.week_start::TEXT,
			    s.status,
			    s.published_at,
			    s.created_by::TEXT
			FROM schedules s
			WHERE s.location_id = $1
			  AND s.week_start = $2`,
			locationID, weekStart,
		).Scan(
			&sched.ScheduleID,
			&sched.LocationID,
			&sched.WeekStart,
			&sched.Status,
			&sched.PublishedAt,
			&sched.CreatedBy,
		)
		if err != nil {
			return fmt.Errorf("query schedule: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	shifts, err := s.listShiftsForSchedule(ctx, orgID, sched.ScheduleID)
	if err != nil {
		return nil, err
	}

	sched.ShiftCount = len(shifts)
	return &ScheduleWithShifts{Schedule: sched, Shifts: shifts}, nil
}

// listShiftsForSchedule fetches all shifts for a schedule, joined with employee names.
func (s *Service) listShiftsForSchedule(ctx context.Context, orgID, scheduleID string) ([]ScheduledShift, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	var shifts []ScheduledShift

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(tenantCtx,
			`SELECT
			    ss.scheduled_shift_id::TEXT,
			    ss.employee_id::TEXT,
			    COALESCE(e.display_name, '') AS employee_name,
			    ss.shift_date::TEXT,
			    to_char(ss.start_time, 'HH24:MI'),
			    to_char(ss.end_time,   'HH24:MI'),
			    COALESCE(ss.station, ''),
			    ss.status,
			    COALESCE(ss.notes, '')
			FROM scheduled_shifts ss
			JOIN employees e ON e.employee_id = ss.employee_id
			WHERE ss.schedule_id = $1
			ORDER BY ss.shift_date, ss.start_time, e.display_name`,
			scheduleID,
		)
		if err != nil {
			return fmt.Errorf("query shifts: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var sh ScheduledShift
			if err := rows.Scan(
				&sh.ShiftID,
				&sh.EmployeeID,
				&sh.EmployeeName,
				&sh.ShiftDate,
				&sh.StartTime,
				&sh.EndTime,
				&sh.Station,
				&sh.Status,
				&sh.Notes,
			); err != nil {
				return fmt.Errorf("scan shift row: %w", err)
			}
			shifts = append(shifts, sh)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	if shifts == nil {
		shifts = []ScheduledShift{}
	}
	return shifts, nil
}

// UpdateScheduleShifts replaces all shifts on a draft schedule with the
// provided set. Non-draft schedules are rejected.
func (s *Service) UpdateScheduleShifts(ctx context.Context, orgID, scheduleID string, shifts []ScheduledShift) error {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	return database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		// Verify the schedule exists and is a draft.
		var status string
		err := tx.QueryRow(tenantCtx,
			`SELECT status FROM schedules WHERE schedule_id = $1`,
			scheduleID,
		).Scan(&status)
		if err != nil {
			return fmt.Errorf("lookup schedule: %w", err)
		}
		if status != "draft" {
			return fmt.Errorf("schedule is not a draft (status=%s)", status)
		}

		// Delete existing shifts.
		if _, err := tx.Exec(tenantCtx,
			`DELETE FROM scheduled_shifts WHERE schedule_id = $1`, scheduleID,
		); err != nil {
			return fmt.Errorf("delete old shifts: %w", err)
		}

		// Insert replacement shifts.
		for i, sh := range shifts {
			_, err := tx.Exec(tenantCtx,
				`INSERT INTO scheduled_shifts
				    (org_id, schedule_id, employee_id, shift_date,
				     start_time, end_time, station, status, notes)
				 VALUES ($1, $2, $3, $4, $5::TIME, $6::TIME, $7, $8, $9)`,
				orgID, scheduleID, sh.EmployeeID, sh.ShiftDate,
				sh.StartTime, sh.EndTime, sh.Station, sh.Status, sh.Notes,
			)
			if err != nil {
				return fmt.Errorf("insert replacement shift %d: %w", i, err)
			}
		}

		// Update the schedule updated_at timestamp.
		_, err = tx.Exec(tenantCtx,
			`UPDATE schedules SET updated_at = now() WHERE schedule_id = $1`,
			scheduleID,
		)
		return err
	})
}

// PublishSchedule transitions a draft schedule to published, records the
// timestamp, and emits a labor.schedule.published event.
func (s *Service) PublishSchedule(ctx context.Context, orgID, scheduleID string) error {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	var locationID, weekStart string

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		err := tx.QueryRow(tenantCtx,
			`UPDATE schedules
			 SET status = 'published', published_at = now(), updated_at = now()
			 WHERE schedule_id = $1 AND status = 'draft'
			 RETURNING location_id::TEXT, week_start::TEXT`,
			scheduleID,
		).Scan(&locationID, &weekStart)
		if err != nil {
			return fmt.Errorf("publish schedule: %w", err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	s.bus.Publish(ctx, event.Envelope{
		EventType: "labor.schedule.published",
		OrgID:     orgID,
		LocationID: locationID,
		Source:    "labor",
		Payload: map[string]any{
			"schedule_id": scheduleID,
			"location_id": locationID,
			"week_start":  weekStart,
		},
	})

	return nil
}

// GetEmployeeSchedule returns all scheduled shifts for a single employee
// for the week starting weekStart.
func (s *Service) GetEmployeeSchedule(ctx context.Context, orgID, employeeID, weekStart string) ([]ScheduledShift, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	weekDate, err := time.Parse("2006-01-02", weekStart)
	if err != nil {
		return nil, fmt.Errorf("invalid week_start: %w", err)
	}
	weekEnd := weekDate.AddDate(0, 0, 7)

	var shifts []ScheduledShift

	err = database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(tenantCtx,
			`SELECT
			    ss.scheduled_shift_id::TEXT,
			    ss.employee_id::TEXT,
			    COALESCE(e.display_name, '') AS employee_name,
			    ss.shift_date::TEXT,
			    to_char(ss.start_time, 'HH24:MI'),
			    to_char(ss.end_time,   'HH24:MI'),
			    COALESCE(ss.station, ''),
			    ss.status,
			    COALESCE(ss.notes, '')
			FROM scheduled_shifts ss
			JOIN employees e ON e.employee_id = ss.employee_id
			WHERE ss.employee_id = $1
			  AND ss.shift_date >= $2
			  AND ss.shift_date < $3
			ORDER BY ss.shift_date, ss.start_time`,
			employeeID,
			weekDate.Format("2006-01-02"),
			weekEnd.Format("2006-01-02"),
		)
		if err != nil {
			return fmt.Errorf("query employee schedule: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var sh ScheduledShift
			if err := rows.Scan(
				&sh.ShiftID,
				&sh.EmployeeID,
				&sh.EmployeeName,
				&sh.ShiftDate,
				&sh.StartTime,
				&sh.EndTime,
				&sh.Station,
				&sh.Status,
				&sh.Notes,
			); err != nil {
				return fmt.Errorf("scan employee shift: %w", err)
			}
			shifts = append(shifts, sh)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	if shifts == nil {
		shifts = []ScheduledShift{}
	}
	return shifts, nil
}

// ProjectLaborCost calculates projected labor cost for a schedule by summing
// (shift hours × employee hourly_rate) for all shifts.
// Budget target is hardcoded to 30% for now; a future sprint will pull from
// the financial budgets table.
func (s *Service) ProjectLaborCost(ctx context.Context, orgID, scheduleID string) (*LaborCostProjection, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	var proj LaborCostProjection

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(tenantCtx,
			`SELECT
			    EXTRACT(EPOCH FROM (ss.end_time - ss.start_time)) / 3600.0 AS shift_hours,
			    COALESCE(e.hourly_rate, 0) AS hourly_rate
			FROM scheduled_shifts ss
			JOIN employees e ON e.employee_id = ss.employee_id
			WHERE ss.schedule_id = $1
			  AND ss.status != 'cancelled'`,
			scheduleID,
		)
		if err != nil {
			return fmt.Errorf("query shifts for cost: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var shiftHours float64
			var hourlyRate int64
			if err := rows.Scan(&shiftHours, &hourlyRate); err != nil {
				return fmt.Errorf("scan cost row: %w", err)
			}
			proj.TotalHours += shiftHours
			proj.TotalCost += int64(shiftHours * float64(hourlyRate))
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}

	const budgetTargetPct = 30.0
	proj.BudgetTarget = budgetTargetPct

	// Labor cost % requires projected revenue; approximate from the last full
	// week of actual revenue for the location. Fall back gracefully if absent.
	tenantCtx2 := tenant.WithOrgID(ctx, orgID)
	var locationID, weekStart string
	_ = database.TenantTx(tenantCtx2, s.pool, func(tx pgx.Tx) error {
		return tx.QueryRow(tenantCtx2,
			`SELECT location_id::TEXT, week_start::TEXT FROM schedules WHERE schedule_id = $1`,
			scheduleID,
		).Scan(&locationID, &weekStart)
	})

	if locationID != "" && weekStart != "" {
		weekDate, _ := time.Parse("2006-01-02", weekStart)
		weekEnd := weekDate.AddDate(0, 0, 7)

		var projRevenue int64
		_ = database.TenantTx(tenantCtx2, s.pool, func(tx pgx.Tx) error {
			return tx.QueryRow(tenantCtx2,
				`SELECT COALESCE(SUM(subtotal), 0)::BIGINT
				 FROM checks
				 WHERE location_id = $1
				   AND status = 'closed'
				   AND closed_at >= $2
				   AND closed_at < $3`,
				locationID,
				weekDate.Format("2006-01-02"),
				weekEnd.Format("2006-01-02"),
			).Scan(&projRevenue)
		})

		if projRevenue > 0 {
			proj.LaborCostPct = float64(proj.TotalCost) / float64(projRevenue) * 100
		}
	}

	switch {
	case proj.LaborCostPct > budgetTargetPct:
		proj.OverUnder = "over"
	case proj.LaborCostPct > 0 && proj.LaborCostPct < budgetTargetPct:
		proj.OverUnder = "under"
	default:
		proj.OverUnder = "on_track"
	}

	return &proj, nil
}

// CheckOvertimeRisk returns employees scheduled for more than 38 h in the
// given week, emitting alert events for those over 40 h.
func (s *Service) CheckOvertimeRisk(ctx context.Context, orgID, locationID, weekStart string) ([]OvertimeRisk, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	weekDate, err := time.Parse("2006-01-02", weekStart)
	if err != nil {
		return nil, fmt.Errorf("invalid week_start: %w", err)
	}
	weekEnd := weekDate.AddDate(0, 0, 7)

	var risks []OvertimeRisk

	err = database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(tenantCtx,
			`SELECT
			    e.employee_id::TEXT,
			    e.display_name,
			    SUM(EXTRACT(EPOCH FROM (ss.end_time - ss.start_time)) / 3600.0) AS scheduled_hours
			FROM scheduled_shifts ss
			JOIN employees e ON e.employee_id = ss.employee_id
			WHERE e.location_id = $1
			  AND ss.shift_date >= $2
			  AND ss.shift_date < $3
			  AND ss.status != 'cancelled'
			GROUP BY e.employee_id, e.display_name
			HAVING SUM(EXTRACT(EPOCH FROM (ss.end_time - ss.start_time)) / 3600.0) > 38
			ORDER BY scheduled_hours DESC`,
			locationID,
			weekDate.Format("2006-01-02"),
			weekEnd.Format("2006-01-02"),
		)
		if err != nil {
			return fmt.Errorf("query overtime risk: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var r OvertimeRisk
			if err := rows.Scan(&r.EmployeeID, &r.EmployeeName, &r.ScheduledHours); err != nil {
				return fmt.Errorf("scan overtime row: %w", err)
			}
			r.Severity = classifyOvertimeRisk(r.ScheduledHours)
			risks = append(risks, r)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}

	// Emit alert events for critical (>40h) employees.
	for _, r := range risks {
		if r.Severity == "critical" {
			s.bus.Publish(ctx, event.Envelope{
				EventType:  "labor.overtime.critical",
				OrgID:      orgID,
				LocationID: locationID,
				Source:     "labor",
				Payload: map[string]any{
					"employee_id":     r.EmployeeID,
					"employee_name":   r.EmployeeName,
					"scheduled_hours": r.ScheduledHours,
					"week_start":      weekStart,
				},
			})
		}
	}

	if risks == nil {
		risks = []OvertimeRisk{}
	}
	return risks, nil
}
