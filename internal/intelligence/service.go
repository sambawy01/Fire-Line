package intelligence

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/opsnerve/fireline/internal/event"
	"github.com/opsnerve/fireline/internal/tenant"
	"github.com/opsnerve/fireline/pkg/database"
)

// Service provides surveillance and anomaly detection capabilities.
type Service struct {
	pool *pgxpool.Pool
	bus  *event.Bus
}

// New creates a new intelligence service.
func New(pool *pgxpool.Pool, bus *event.Bus) *Service {
	return &Service{pool: pool, bus: bus}
}

// ─── Structs ────────────────────────────────────────────────────────────────

// Anomaly represents a detected operational anomaly.
type Anomaly struct {
	AnomalyID       string          `json:"anomaly_id"`
	OrgID           string          `json:"org_id"`
	LocationID      string          `json:"location_id"`
	EmployeeID      *string         `json:"employee_id"`
	Type            string          `json:"type"`
	Severity        string          `json:"severity"`
	Title           string          `json:"title"`
	Description     string          `json:"description"`
	Evidence        json.RawMessage `json:"evidence"`
	Status          string          `json:"status"`
	AssignedTo      *string         `json:"assigned_to"`
	ResolvedBy      *string         `json:"resolved_by"`
	ResolvedAt      *string         `json:"resolved_at"`
	ResolutionNotes *string         `json:"resolution_notes"`
	DetectedAt      string          `json:"detected_at"`
	CreatedAt       string          `json:"created_at"`
}

// ResolveInput is the input for resolving an anomaly.
type ResolveInput struct {
	Status          string `json:"status"`           // confirmed, false_positive, resolved
	ResolutionNotes string `json:"resolution_notes"`
	ResolvedBy      string `json:"-"` // set from auth context, not from JSON body
}

// AnomalyInput is the input for creating a new anomaly.
type AnomalyInput struct {
	LocationID  string          `json:"location_id"`
	EmployeeID  *string         `json:"employee_id"`
	Type        string          `json:"type"`
	Severity    string          `json:"severity"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Evidence    json.RawMessage `json:"evidence"`
}

// EmployeeTimeline is an aggregated investigative view of an employee.
type EmployeeTimeline struct {
	EmployeeID     string         `json:"employee_id"`
	DisplayName    string         `json:"display_name"`
	Role           string         `json:"role"`
	RecentShifts   []ShiftSummary `json:"recent_shifts"`
	VoidHistory    []VoidRecord   `json:"void_history"`
	TaskCompletion TaskStats      `json:"task_completion"`
	AnomalyHistory []Anomaly      `json:"anomaly_history"`
}

// ShiftSummary is a condensed shift record for the timeline.
type ShiftSummary struct {
	ShiftID  string  `json:"shift_id"`
	ClockIn  string  `json:"clock_in"`
	ClockOut *string `json:"clock_out"`
	Role     string  `json:"role"`
	Hours    float64 `json:"hours"`
}

// VoidRecord represents a single voided item in the timeline.
type VoidRecord struct {
	CheckID    string  `json:"check_id"`
	ItemName   string  `json:"item_name"`
	Amount     int64   `json:"amount"`
	VoidReason *string `json:"void_reason"`
	VoidedAt   string  `json:"voided_at"`
}

// TaskStats holds completion statistics for an employee.
type TaskStats struct {
	Total     int     `json:"total"`
	Completed int     `json:"completed"`
	Rate      float64 `json:"rate"`
}

// ─── Anomaly CRUD ───────────────────────────────────────────────────────────

// ListAnomalies returns anomalies with optional filters. If locationID is empty,
// anomalies across ALL locations are returned (for ops_director cross-location view).
func (s *Service) ListAnomalies(ctx context.Context, orgID, locationID, status, anomalyType string) ([]Anomaly, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var results []Anomaly

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		query := `SELECT anomaly_id, org_id, location_id, employee_id,
				type, severity, title, description, evidence, status,
				assigned_to, resolved_by, resolved_at::TEXT, resolution_notes,
				detected_at::TEXT, created_at::TEXT
			FROM anomalies WHERE 1=1`
		args := []any{}
		argIdx := 1

		if locationID != "" {
			query += fmt.Sprintf(" AND location_id = $%d", argIdx)
			args = append(args, locationID)
			argIdx++
		}
		if status != "" {
			query += fmt.Sprintf(" AND status = $%d", argIdx)
			args = append(args, status)
			argIdx++
		}
		if anomalyType != "" {
			query += fmt.Sprintf(" AND type = $%d", argIdx)
			args = append(args, anomalyType)
			argIdx++
		}

		query += " ORDER BY detected_at DESC LIMIT 200"

		rows, err := tx.Query(tenantCtx, query, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var a Anomaly
			if err := rows.Scan(
				&a.AnomalyID, &a.OrgID, &a.LocationID, &a.EmployeeID,
				&a.Type, &a.Severity, &a.Title, &a.Description, &a.Evidence, &a.Status,
				&a.AssignedTo, &a.ResolvedBy, &a.ResolvedAt, &a.ResolutionNotes,
				&a.DetectedAt, &a.CreatedAt,
			); err != nil {
				return err
			}
			results = append(results, a)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("list anomalies: %w", err)
	}
	return results, nil
}

// GetAnomaly returns a single anomaly by ID.
func (s *Service) GetAnomaly(ctx context.Context, orgID, anomalyID string) (*Anomaly, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var a Anomaly

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		return tx.QueryRow(tenantCtx,
			`SELECT anomaly_id, org_id, location_id, employee_id,
				type, severity, title, description, evidence, status,
				assigned_to, resolved_by, resolved_at::TEXT, resolution_notes,
				detected_at::TEXT, created_at::TEXT
			FROM anomalies
			WHERE anomaly_id = $1`,
			anomalyID,
		).Scan(
			&a.AnomalyID, &a.OrgID, &a.LocationID, &a.EmployeeID,
			&a.Type, &a.Severity, &a.Title, &a.Description, &a.Evidence, &a.Status,
			&a.AssignedTo, &a.ResolvedBy, &a.ResolvedAt, &a.ResolutionNotes,
			&a.DetectedAt, &a.CreatedAt,
		)
	})
	if err != nil {
		return nil, fmt.Errorf("get anomaly: %w", err)
	}
	return &a, nil
}

// ResolveAnomaly updates the status, resolved_by, resolved_at, and resolution_notes.
func (s *Service) ResolveAnomaly(ctx context.Context, orgID, anomalyID string, input ResolveInput) (*Anomaly, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var a Anomaly

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		return tx.QueryRow(tenantCtx,
			`UPDATE anomalies
			SET status = $1,
				resolved_by = $2,
				resolved_at = NOW(),
				resolution_notes = $3
			WHERE anomaly_id = $4
			RETURNING anomaly_id, org_id, location_id, employee_id,
				type, severity, title, description, evidence, status,
				assigned_to, resolved_by, resolved_at::TEXT, resolution_notes,
				detected_at::TEXT, created_at::TEXT`,
			input.Status, input.ResolvedBy, input.ResolutionNotes, anomalyID,
		).Scan(
			&a.AnomalyID, &a.OrgID, &a.LocationID, &a.EmployeeID,
			&a.Type, &a.Severity, &a.Title, &a.Description, &a.Evidence, &a.Status,
			&a.AssignedTo, &a.ResolvedBy, &a.ResolvedAt, &a.ResolutionNotes,
			&a.DetectedAt, &a.CreatedAt,
		)
	})
	if err != nil {
		return nil, fmt.Errorf("resolve anomaly: %w", err)
	}
	return &a, nil
}

// CreateAnomaly inserts a new detected anomaly.
func (s *Service) CreateAnomaly(ctx context.Context, orgID string, input AnomalyInput) (*Anomaly, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var a Anomaly

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		return tx.QueryRow(tenantCtx,
			`INSERT INTO anomalies (org_id, location_id, employee_id, type, severity,
				title, description, evidence, status, detected_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'open', NOW())
			RETURNING anomaly_id, org_id, location_id, employee_id,
				type, severity, title, description, evidence, status,
				assigned_to, resolved_by, resolved_at::TEXT, resolution_notes,
				detected_at::TEXT, created_at::TEXT`,
			orgID, input.LocationID, input.EmployeeID, input.Type, input.Severity,
			input.Title, input.Description, input.Evidence,
		).Scan(
			&a.AnomalyID, &a.OrgID, &a.LocationID, &a.EmployeeID,
			&a.Type, &a.Severity, &a.Title, &a.Description, &a.Evidence, &a.Status,
			&a.AssignedTo, &a.ResolvedBy, &a.ResolvedAt, &a.ResolutionNotes,
			&a.DetectedAt, &a.CreatedAt,
		)
	})
	if err != nil {
		return nil, fmt.Errorf("create anomaly: %w", err)
	}
	return &a, nil
}

// ─── Investigation ──────────────────────────────────────────────────────────

// GetEmployeeTimeline returns an aggregated investigative view of an employee
// including recent shifts, void/discount history, task completion, and anomaly history.
func (s *Service) GetEmployeeTimeline(ctx context.Context, orgID, employeeID string, days int) (*EmployeeTimeline, error) {
	if days <= 0 {
		days = 30
	}
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	tl := &EmployeeTimeline{
		EmployeeID:   employeeID,
		RecentShifts: []ShiftSummary{},
		VoidHistory:  []VoidRecord{},
		AnomalyHistory: []Anomaly{},
	}

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		// Fetch employee display_name and role
		err := tx.QueryRow(tenantCtx,
			`SELECT COALESCE(display_name, first_name || ' ' || last_name), role
			FROM employees WHERE employee_id = $1`,
			employeeID,
		).Scan(&tl.DisplayName, &tl.Role)
		if err != nil {
			return fmt.Errorf("employee lookup: %w", err)
		}

		// Recent shifts
		shiftRows, err := tx.Query(tenantCtx,
			`SELECT shift_id, clock_in::TEXT, clock_out::TEXT, role,
				COALESCE(EXTRACT(EPOCH FROM (clock_out - clock_in)) / 3600.0, 0)
			FROM shifts
			WHERE employee_id = $1
				AND clock_in >= NOW() - make_interval(days => $2)
			ORDER BY clock_in DESC
			LIMIT 50`,
			employeeID, days,
		)
		if err != nil {
			return fmt.Errorf("shifts query: %w", err)
		}
		defer shiftRows.Close()

		for shiftRows.Next() {
			var ss ShiftSummary
			if err := shiftRows.Scan(&ss.ShiftID, &ss.ClockIn, &ss.ClockOut, &ss.Role, &ss.Hours); err != nil {
				return err
			}
			tl.RecentShifts = append(tl.RecentShifts, ss)
		}
		if err := shiftRows.Err(); err != nil {
			return err
		}

		// Void history: voided check items where the check's server matches
		voidRows, err := tx.Query(tenantCtx,
			`SELECT c.check_id, ci.name, ci.price, ci.void_reason, ci.voided_at::TEXT
			FROM check_items ci
			JOIN checks c ON c.check_id = ci.check_id
			WHERE c.server_id = $1
				AND ci.voided_at IS NOT NULL
				AND ci.voided_at >= NOW() - make_interval(days => $2)
			ORDER BY ci.voided_at DESC
			LIMIT 100`,
			employeeID, days,
		)
		if err != nil {
			return fmt.Errorf("void history query: %w", err)
		}
		defer voidRows.Close()

		for voidRows.Next() {
			var vr VoidRecord
			if err := voidRows.Scan(&vr.CheckID, &vr.ItemName, &vr.Amount, &vr.VoidReason, &vr.VoidedAt); err != nil {
				return err
			}
			tl.VoidHistory = append(tl.VoidHistory, vr)
		}
		if err := voidRows.Err(); err != nil {
			return err
		}

		// Task completion rate
		err = tx.QueryRow(tenantCtx,
			`SELECT
				COUNT(*)::INT,
				COUNT(*) FILTER (WHERE status = 'completed')::INT
			FROM tasks
			WHERE assigned_to = $1
				AND created_at >= NOW() - make_interval(days => $2)`,
			employeeID, days,
		).Scan(&tl.TaskCompletion.Total, &tl.TaskCompletion.Completed)
		if err != nil {
			return fmt.Errorf("task stats query: %w", err)
		}
		if tl.TaskCompletion.Total > 0 {
			tl.TaskCompletion.Rate = float64(tl.TaskCompletion.Completed) / float64(tl.TaskCompletion.Total)
		}

		// Anomaly history
		anomalyRows, err := tx.Query(tenantCtx,
			`SELECT anomaly_id, org_id, location_id, employee_id,
				type, severity, title, description, evidence, status,
				assigned_to, resolved_by, resolved_at::TEXT, resolution_notes,
				detected_at::TEXT, created_at::TEXT
			FROM anomalies
			WHERE employee_id = $1
				AND detected_at >= NOW() - make_interval(days => $2)
			ORDER BY detected_at DESC
			LIMIT 50`,
			employeeID, days,
		)
		if err != nil {
			return fmt.Errorf("anomaly history query: %w", err)
		}
		defer anomalyRows.Close()

		for anomalyRows.Next() {
			var a Anomaly
			if err := anomalyRows.Scan(
				&a.AnomalyID, &a.OrgID, &a.LocationID, &a.EmployeeID,
				&a.Type, &a.Severity, &a.Title, &a.Description, &a.Evidence, &a.Status,
				&a.AssignedTo, &a.ResolvedBy, &a.ResolvedAt, &a.ResolutionNotes,
				&a.DetectedAt, &a.CreatedAt,
			); err != nil {
				return err
			}
			tl.AnomalyHistory = append(tl.AnomalyHistory, a)
		}
		return anomalyRows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("get employee timeline: %w", err)
	}
	return tl, nil
}

// ─── Detection (Event Handlers) ─────────────────────────────────────────────

// RegisterHandlers subscribes to events that trigger anomaly detection.
func (s *Service) RegisterHandlers() {
	s.bus.Subscribe("pipeline.orders.processed", s.detectOrderAnomalies)
}

// ordersProcessedPayload is the expected shape of the pipeline.orders.processed event payload.
type ordersProcessedPayload struct {
	LocationID string `json:"location_id"`
}

// employeeVoidStats holds per-employee void counts for z-score calculation.
type employeeVoidStats struct {
	EmployeeID string
	VoidCount  int
	TotalItems int
	VoidRate   float64
}

// detectOrderAnomalies runs after orders are processed. It checks for void/discount
// anomalies using a simple z-score approach: for each employee with voids, compare
// their void rate to the location average. z > 2.0 = warning, z > 3.0 = critical.
func (s *Service) detectOrderAnomalies(ctx context.Context, env event.Envelope) error {
	// Parse payload to get location_id
	payloadBytes, err := json.Marshal(env.Payload)
	if err != nil {
		slog.Error("intelligence: failed to marshal event payload", "error", err)
		return fmt.Errorf("marshal payload: %w", err)
	}

	var payload ordersProcessedPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		slog.Error("intelligence: failed to unmarshal event payload", "error", err)
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	locationID := payload.LocationID
	if locationID == "" {
		locationID = env.LocationID
	}
	if locationID == "" {
		slog.Warn("intelligence: no location_id in orders.processed event, skipping detection")
		return nil
	}

	orgID := env.OrgID
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	return database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		// Get per-employee void stats for the last 7 days at this location
		rows, err := tx.Query(tenantCtx,
			`SELECT c.server_id,
				COUNT(*) FILTER (WHERE ci.voided_at IS NOT NULL) AS void_count,
				COUNT(*) AS total_items,
				CASE WHEN COUNT(*) > 0
					THEN COUNT(*) FILTER (WHERE ci.voided_at IS NOT NULL)::FLOAT / COUNT(*)::FLOAT
					ELSE 0
				END AS void_rate
			FROM check_items ci
			JOIN checks c ON c.check_id = ci.check_id
			WHERE c.location_id = $1
				AND c.created_at >= NOW() - INTERVAL '7 days'
				AND c.server_id IS NOT NULL
			GROUP BY c.server_id
			HAVING COUNT(*) >= 10`,
			locationID,
		)
		if err != nil {
			return fmt.Errorf("void stats query: %w", err)
		}
		defer rows.Close()

		var stats []employeeVoidStats
		for rows.Next() {
			var es employeeVoidStats
			if err := rows.Scan(&es.EmployeeID, &es.VoidCount, &es.TotalItems, &es.VoidRate); err != nil {
				return err
			}
			stats = append(stats, es)
		}
		if err := rows.Err(); err != nil {
			return err
		}

		if len(stats) < 3 {
			// Not enough data for meaningful z-score analysis
			return nil
		}

		// Calculate mean and standard deviation of void rates
		var sum, sumSq float64
		for _, es := range stats {
			sum += es.VoidRate
			sumSq += es.VoidRate * es.VoidRate
		}
		n := float64(len(stats))
		mean := sum / n
		variance := (sumSq / n) - (mean * mean)
		stdDev := math.Sqrt(variance)

		if stdDev < 0.001 {
			// All employees have nearly identical void rates, no outliers
			return nil
		}

		// Check each employee for anomalous void rates
		for _, es := range stats {
			zScore := (es.VoidRate - mean) / stdDev
			if zScore <= 2.0 {
				continue
			}

			severity := "warning"
			if zScore > 3.0 {
				severity = "critical"
			}

			evidence, _ := json.Marshal(map[string]any{
				"employee_void_rate":  es.VoidRate,
				"location_avg_rate":   mean,
				"z_score":             math.Round(zScore*100) / 100,
				"void_count":          es.VoidCount,
				"total_items":         es.TotalItems,
				"analysis_window":     "7 days",
				"employees_in_sample": len(stats),
			})

			empID := es.EmployeeID
			title := fmt.Sprintf("Elevated void rate detected (z=%.1f)", zScore)
			description := fmt.Sprintf(
				"Employee has a void rate of %.1f%% compared to the location average of %.1f%% over the last 7 days (%d voids out of %d items).",
				es.VoidRate*100, mean*100, es.VoidCount, es.TotalItems,
			)

			_, err := tx.Exec(tenantCtx,
				`INSERT INTO anomalies (org_id, location_id, employee_id, type, severity,
					title, description, evidence, status, detected_at)
				VALUES ($1, $2, $3, 'void_pattern', $4, $5, $6, $7, 'open', NOW())
				ON CONFLICT DO NOTHING`,
				orgID, locationID, empID, severity, title, description, json.RawMessage(evidence),
			)
			if err != nil {
				slog.Error("intelligence: failed to insert void anomaly",
					"employee_id", empID,
					"z_score", strconv.FormatFloat(zScore, 'f', 2, 64),
					"error", err,
				)
				// Continue checking other employees rather than aborting
				continue
			}

			slog.Info("intelligence: anomaly detected",
				"type", "void_pattern",
				"severity", severity,
				"employee_id", empID,
				"z_score", strconv.FormatFloat(zScore, 'f', 2, 64),
			)
		}

		return nil
	})
}
