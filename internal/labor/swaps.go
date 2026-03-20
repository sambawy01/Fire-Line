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

// SwapRequest holds a shift-swap request record.
type SwapRequest struct {
	SwapID           string     `json:"swap_id"`
	RequesterShiftID string     `json:"requester_shift_id"`
	RequesterName    string     `json:"requester_name,omitempty"`
	TargetEmployeeID *string    `json:"target_employee_id,omitempty"`
	TargetName       string     `json:"target_name,omitempty"`
	Status           string     `json:"status"`
	Reason           string     `json:"reason"`
	ReviewedBy       *string    `json:"reviewed_by,omitempty"`
	ReviewedAt       *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
}

// RequestSwap creates a shift-swap request.
// targetEmployeeID is optional (pass empty string for open/any swap).
func (s *Service) RequestSwap(ctx context.Context, orgID, requesterShiftID, targetEmployeeID, reason string) (*SwapRequest, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	var req SwapRequest

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		// Verify the requester shift exists and belongs to this org.
		var shiftExists bool
		if err := tx.QueryRow(tenantCtx,
			`SELECT EXISTS(SELECT 1 FROM scheduled_shifts WHERE scheduled_shift_id = $1)`,
			requesterShiftID,
		).Scan(&shiftExists); err != nil {
			return fmt.Errorf("verify shift: %w", err)
		}
		if !shiftExists {
			return fmt.Errorf("shift not found: %s", requesterShiftID)
		}

		var targetEmpID *string
		if targetEmployeeID != "" {
			targetEmpID = &targetEmployeeID
		}

		err := tx.QueryRow(tenantCtx,
			`INSERT INTO shift_swap_requests
			    (org_id, requester_shift_id, target_employee_id, status, reason)
			 VALUES ($1, $2, $3, 'pending', $4)
			 RETURNING swap_id::TEXT, requester_shift_id::TEXT,
			           target_employee_id::TEXT, status,
			           COALESCE(reason, ''), created_at`,
			orgID, requesterShiftID, targetEmpID, reason,
		).Scan(
			&req.SwapID,
			&req.RequesterShiftID,
			&req.TargetEmployeeID,
			&req.Status,
			&req.Reason,
			&req.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("insert swap request: %w", err)
		}

		// Mark the requester shift as swap_requested.
		if _, err := tx.Exec(tenantCtx,
			`UPDATE scheduled_shifts SET status = 'swap_requested'
			 WHERE scheduled_shift_id = $1`,
			requesterShiftID,
		); err != nil {
			return fmt.Errorf("update shift status: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	s.bus.Publish(ctx, event.Envelope{
		EventType: "labor.swap.requested",
		OrgID:     orgID,
		Source:    "labor",
		Payload: map[string]any{
			"swap_id":            req.SwapID,
			"requester_shift_id": req.RequesterShiftID,
		},
	})

	return &req, nil
}

// ReviewSwap approves or denies a swap request.
// If approved, the requester shift is marked 'swapped' and (if a target
// employee is set) their shift statuses are updated accordingly.
func (s *Service) ReviewSwap(ctx context.Context, orgID, swapID string, approved bool, reviewedBy string) error {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	newStatus := "denied"
	if approved {
		newStatus = "approved"
	}

	var requesterShiftID, targetEmployeeID string
	var hasTarget bool

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		// Fetch the swap request.
		var targetEmpIDPtr *string
		err := tx.QueryRow(tenantCtx,
			`UPDATE shift_swap_requests
			 SET status = $1, reviewed_by = $2, reviewed_at = now()
			 WHERE swap_id = $3 AND status = 'pending'
			 RETURNING requester_shift_id::TEXT, target_employee_id::TEXT`,
			newStatus, reviewedBy, swapID,
		).Scan(&requesterShiftID, &targetEmpIDPtr)
		if err != nil {
			return fmt.Errorf("review swap: %w", err)
		}

		if targetEmpIDPtr != nil {
			targetEmployeeID = *targetEmpIDPtr
			hasTarget = true
		}

		if approved {
			// Mark requester shift as swapped.
			if _, err := tx.Exec(tenantCtx,
				`UPDATE scheduled_shifts SET status = 'swapped'
				 WHERE scheduled_shift_id = $1`,
				requesterShiftID,
			); err != nil {
				return fmt.Errorf("mark requester shift swapped: %w", err)
			}

			// If a target employee is named, copy the shift to them.
			if hasTarget {
				if _, err := tx.Exec(tenantCtx,
					`INSERT INTO scheduled_shifts
					    (org_id, schedule_id, employee_id, shift_date,
					     start_time, end_time, station, status, notes)
					 SELECT org_id, schedule_id, $1, shift_date,
					        start_time, end_time, station, 'scheduled',
					        'Swap from ' || scheduled_shift_id::TEXT
					 FROM scheduled_shifts
					 WHERE scheduled_shift_id = $2`,
					targetEmployeeID, requesterShiftID,
				); err != nil {
					return fmt.Errorf("create target shift: %w", err)
				}
			}
		} else {
			// Denied: revert the requester shift back to scheduled.
			if _, err := tx.Exec(tenantCtx,
				`UPDATE scheduled_shifts SET status = 'scheduled'
				 WHERE scheduled_shift_id = $1`,
				requesterShiftID,
			); err != nil {
				return fmt.Errorf("revert requester shift: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	s.bus.Publish(ctx, event.Envelope{
		EventType: "labor.swap.reviewed",
		OrgID:     orgID,
		Source:    "labor",
		Payload: map[string]any{
			"swap_id":    swapID,
			"status":     newStatus,
			"reviewed_by": reviewedBy,
		},
	})

	return nil
}

// ListSwapRequests returns swap requests for an org, optionally filtered by
// locationID and status. Pass empty strings to omit filters.
func (s *Service) ListSwapRequests(ctx context.Context, orgID, locationID, status string) ([]SwapRequest, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	var reqs []SwapRequest

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		// Base query: join to scheduled_shifts to filter by location_id,
		// and join employees to surface requester/target names.
		query := `
			SELECT
			    sr.swap_id::TEXT,
			    sr.requester_shift_id::TEXT,
			    COALESCE(req_emp.display_name, '') AS requester_name,
			    sr.target_employee_id::TEXT,
			    COALESCE(tgt_emp.display_name, '') AS target_name,
			    sr.status,
			    COALESCE(sr.reason, ''),
			    sr.reviewed_by::TEXT,
			    sr.reviewed_at,
			    sr.created_at
			FROM shift_swap_requests sr
			JOIN scheduled_shifts ss ON ss.scheduled_shift_id = sr.requester_shift_id
			JOIN employees req_emp ON req_emp.employee_id = ss.employee_id
			LEFT JOIN employees tgt_emp ON tgt_emp.employee_id = sr.target_employee_id
			WHERE sr.org_id = $1`

		args := []any{orgID}
		argN := 2

		if locationID != "" {
			query += fmt.Sprintf(" AND ss.schedule_id IN (SELECT schedule_id FROM schedules WHERE location_id = $%d)", argN)
			args = append(args, locationID)
			argN++
		}
		if status != "" {
			query += fmt.Sprintf(" AND sr.status = $%d", argN)
			args = append(args, status)
			argN++
		}
		query += " ORDER BY sr.created_at DESC"

		rows, err := tx.Query(tenantCtx, query, args...)
		if err != nil {
			return fmt.Errorf("query swap requests: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var r SwapRequest
			if err := rows.Scan(
				&r.SwapID,
				&r.RequesterShiftID,
				&r.RequesterName,
				&r.TargetEmployeeID,
				&r.TargetName,
				&r.Status,
				&r.Reason,
				&r.ReviewedBy,
				&r.ReviewedAt,
				&r.CreatedAt,
			); err != nil {
				return fmt.Errorf("scan swap row: %w", err)
			}
			reqs = append(reqs, r)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	if reqs == nil {
		reqs = []SwapRequest{}
	}
	return reqs, nil
}
