package maintenance

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/opsnerve/fireline/internal/event"
	"github.com/opsnerve/fireline/internal/tenant"
	"github.com/opsnerve/fireline/pkg/database"
)

// MaintenanceTicket represents a maintenance work order.
type MaintenanceTicket struct {
	TicketID      string  `json:"ticket_id"`
	OrgID         string  `json:"org_id"`
	LocationID    string  `json:"location_id"`
	EquipmentID   string  `json:"equipment_id"`
	EquipmentName string  `json:"equipment_name"`
	TicketNumber  string  `json:"ticket_number"`
	Type          string  `json:"type"`
	Priority      string  `json:"priority"`
	Status        string  `json:"status"`
	Title         string  `json:"title"`
	Description   *string `json:"description"`
	AssignedTo    *string `json:"assigned_to"`
	EstimatedCost int     `json:"estimated_cost"`
	ActualCost    int     `json:"actual_cost"`
	ScheduledDate *string `json:"scheduled_date"`
	StartedAt     *string `json:"started_at"`
	CompletedAt   *string `json:"completed_at"`
	Resolution    *string `json:"resolution"`
	CreatedBy     *string `json:"created_by"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
	Logs          []MaintenanceLog `json:"logs,omitempty"`
}

// TicketInput represents the input for creating a ticket.
type TicketInput struct {
	LocationID    string  `json:"location_id"`
	EquipmentID   string  `json:"equipment_id"`
	Type          string  `json:"type"`
	Priority      string  `json:"priority"`
	Title         string  `json:"title"`
	Description   *string `json:"description"`
	AssignedTo    *string `json:"assigned_to"`
	EstimatedCost int     `json:"estimated_cost"`
	ScheduledDate *string `json:"scheduled_date"`
}

// MaintenanceLog represents a log entry for a maintenance activity.
type MaintenanceLog struct {
	LogID       string  `json:"log_id"`
	OrgID       string  `json:"org_id"`
	TicketID    *string `json:"ticket_id"`
	EquipmentID string  `json:"equipment_id"`
	Action      string  `json:"action"`
	Notes       *string `json:"notes"`
	Cost        int     `json:"cost"`
	PerformedBy *string `json:"performed_by"`
	PerformedAt string  `json:"performed_at"`
}

// CreateTicket inserts a new maintenance ticket and emits an event.
func (s *Service) CreateTicket(ctx context.Context, orgID string, input TicketInput) (*MaintenanceTicket, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var ticket MaintenanceTicket

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		// Generate ticket number from sequence (concurrency-safe, no duplicates)
		var ticketNumber int
		if err := tx.QueryRow(tenantCtx, "SELECT nextval('maintenance_ticket_seq')").Scan(&ticketNumber); err != nil {
			return fmt.Errorf("generate ticket number: %w", err)
		}
		ticketNum := fmt.Sprintf("MT-%04d", ticketNumber)

		return tx.QueryRow(tenantCtx,
			`INSERT INTO maintenance_tickets (org_id, location_id, equipment_id, ticket_number, type, priority, status,
				title, description, assigned_to, estimated_cost, scheduled_date)
			 VALUES ($1, $2, $3, $4, $5, $6, 'open', $7, $8, $9, $10, $11::DATE)
			 RETURNING ticket_id, org_id, location_id, equipment_id, ticket_number, type, priority, status,
				title, description, assigned_to, estimated_cost, actual_cost, scheduled_date::TEXT,
				started_at::TEXT, completed_at::TEXT, resolution, created_by::TEXT, created_at::TEXT, updated_at::TEXT`,
			orgID, input.LocationID, input.EquipmentID, ticketNum, input.Type, input.Priority,
			input.Title, input.Description, input.AssignedTo, input.EstimatedCost, input.ScheduledDate,
		).Scan(
			&ticket.TicketID, &ticket.OrgID, &ticket.LocationID, &ticket.EquipmentID,
			&ticket.TicketNumber, &ticket.Type, &ticket.Priority, &ticket.Status,
			&ticket.Title, &ticket.Description, &ticket.AssignedTo, &ticket.EstimatedCost, &ticket.ActualCost,
			&ticket.ScheduledDate, &ticket.StartedAt, &ticket.CompletedAt, &ticket.Resolution,
			&ticket.CreatedBy, &ticket.CreatedAt, &ticket.UpdatedAt,
		)
	})
	if err != nil {
		return nil, fmt.Errorf("create ticket: %w", err)
	}

	// Emit event
	s.bus.Publish(ctx, event.Envelope{
		EventType:  "maintenance.ticket.created",
		OrgID:      orgID,
		LocationID: input.LocationID,
		Source:     "maintenance",
		Payload:    ticket,
	})

	return &ticket, nil
}

// ListTickets returns tickets filtered by location, status, and priority, with equipment name.
func (s *Service) ListTickets(ctx context.Context, orgID, locationID, status, priority string) ([]MaintenanceTicket, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var results []MaintenanceTicket

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		query := `SELECT t.ticket_id, t.org_id, t.location_id, t.equipment_id, e.name,
				t.ticket_number, t.type, t.priority, t.status,
				t.title, t.description, t.assigned_to, t.estimated_cost, t.actual_cost,
				t.scheduled_date::TEXT, t.started_at::TEXT, t.completed_at::TEXT, t.resolution,
				t.created_by::TEXT, t.created_at::TEXT, t.updated_at::TEXT
			FROM maintenance_tickets t
			JOIN equipment e ON e.equipment_id = t.equipment_id
			WHERE 1=1`
		args := []any{}
		argIdx := 1

		if locationID != "" {
			query += fmt.Sprintf(" AND t.location_id = $%d", argIdx)
			args = append(args, locationID)
			argIdx++
		}
		if status != "" {
			query += fmt.Sprintf(" AND t.status = $%d", argIdx)
			args = append(args, status)
			argIdx++
		}
		if priority != "" {
			query += fmt.Sprintf(" AND t.priority = $%d", argIdx)
			args = append(args, priority)
			argIdx++
		}
		query += " ORDER BY t.created_at DESC"

		rows, err := tx.Query(tenantCtx, query, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var t MaintenanceTicket
			if err := rows.Scan(
				&t.TicketID, &t.OrgID, &t.LocationID, &t.EquipmentID, &t.EquipmentName,
				&t.TicketNumber, &t.Type, &t.Priority, &t.Status,
				&t.Title, &t.Description, &t.AssignedTo, &t.EstimatedCost, &t.ActualCost,
				&t.ScheduledDate, &t.StartedAt, &t.CompletedAt, &t.Resolution,
				&t.CreatedBy, &t.CreatedAt, &t.UpdatedAt,
			); err != nil {
				return err
			}
			results = append(results, t)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("list tickets: %w", err)
	}
	return results, nil
}

// GetTicket returns a single ticket with its logs.
func (s *Service) GetTicket(ctx context.Context, orgID, ticketID string) (*MaintenanceTicket, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var ticket MaintenanceTicket

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		if err := tx.QueryRow(tenantCtx,
			`SELECT t.ticket_id, t.org_id, t.location_id, t.equipment_id, e.name,
				t.ticket_number, t.type, t.priority, t.status,
				t.title, t.description, t.assigned_to, t.estimated_cost, t.actual_cost,
				t.scheduled_date::TEXT, t.started_at::TEXT, t.completed_at::TEXT, t.resolution,
				t.created_by::TEXT, t.created_at::TEXT, t.updated_at::TEXT
			 FROM maintenance_tickets t
			 JOIN equipment e ON e.equipment_id = t.equipment_id
			 WHERE t.ticket_id = $1`,
			ticketID,
		).Scan(
			&ticket.TicketID, &ticket.OrgID, &ticket.LocationID, &ticket.EquipmentID, &ticket.EquipmentName,
			&ticket.TicketNumber, &ticket.Type, &ticket.Priority, &ticket.Status,
			&ticket.Title, &ticket.Description, &ticket.AssignedTo, &ticket.EstimatedCost, &ticket.ActualCost,
			&ticket.ScheduledDate, &ticket.StartedAt, &ticket.CompletedAt, &ticket.Resolution,
			&ticket.CreatedBy, &ticket.CreatedAt, &ticket.UpdatedAt,
		); err != nil {
			return err
		}

		// Get logs
		rows, err := tx.Query(tenantCtx,
			`SELECT log_id, org_id, ticket_id::TEXT, equipment_id, action, notes, cost, performed_by, performed_at::TEXT
			 FROM maintenance_logs WHERE ticket_id = $1 ORDER BY performed_at ASC`,
			ticketID,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var log MaintenanceLog
			if err := rows.Scan(
				&log.LogID, &log.OrgID, &log.TicketID, &log.EquipmentID,
				&log.Action, &log.Notes, &log.Cost, &log.PerformedBy, &log.PerformedAt,
			); err != nil {
				return err
			}
			ticket.Logs = append(ticket.Logs, log)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("get ticket: %w", err)
	}
	return &ticket, nil
}

// UpdateTicket updates ticket fields.
func (s *Service) UpdateTicket(ctx context.Context, orgID, ticketID string, updates map[string]any) error {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	return database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		// Build dynamic update
		setClauses := "updated_at = now()"
		args := []any{ticketID}
		argIdx := 2

		for key, val := range updates {
			switch key {
			case "status":
				setClauses += fmt.Sprintf(", status = $%d", argIdx)
				args = append(args, val)
				argIdx++
				if val == "in_progress" {
					setClauses += ", started_at = COALESCE(started_at, now())"
				}
			case "priority":
				setClauses += fmt.Sprintf(", priority = $%d", argIdx)
				args = append(args, val)
				argIdx++
			case "assigned_to":
				setClauses += fmt.Sprintf(", assigned_to = $%d", argIdx)
				args = append(args, val)
				argIdx++
			case "description":
				setClauses += fmt.Sprintf(", description = $%d", argIdx)
				args = append(args, val)
				argIdx++
			case "estimated_cost":
				setClauses += fmt.Sprintf(", estimated_cost = $%d", argIdx)
				args = append(args, val)
				argIdx++
			case "scheduled_date":
				setClauses += fmt.Sprintf(", scheduled_date = $%d::DATE", argIdx)
				args = append(args, val)
				argIdx++
			}
		}

		query := fmt.Sprintf("UPDATE maintenance_tickets SET %s WHERE ticket_id = $1", setClauses)
		_, err := tx.Exec(tenantCtx, query, args...)
		return err
	})
}

// CompleteTicket marks a ticket as completed and updates the equipment's maintenance dates.
func (s *Service) CompleteTicket(ctx context.Context, orgID, ticketID, resolution string, actualCost int) error {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	return database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		// Complete the ticket
		var equipmentID string
		if err := tx.QueryRow(tenantCtx,
			`UPDATE maintenance_tickets SET
				status = 'completed', completed_at = now(), resolution = $2, actual_cost = $3, updated_at = now()
			 WHERE ticket_id = $1
			 RETURNING equipment_id`,
			ticketID, resolution, actualCost,
		).Scan(&equipmentID); err != nil {
			return fmt.Errorf("complete ticket: %w", err)
		}

		// Update equipment maintenance dates
		today := time.Now().Format("2006-01-02")
		_, err := tx.Exec(tenantCtx,
			`UPDATE equipment SET
				last_maintenance = $2::DATE,
				next_maintenance = ($2::DATE + (maintenance_interval_days || ' days')::INTERVAL)::DATE,
				updated_at = now()
			 WHERE equipment_id = $1`,
			equipmentID, today,
		)
		return err
	})
}

// AddLog adds a maintenance log entry.
func (s *Service) AddLog(ctx context.Context, orgID, ticketID, equipmentID, action string, notes *string, cost int, performedBy *string) (*MaintenanceLog, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var log MaintenanceLog

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		// If ticketID is provided, get equipmentID from ticket
		if equipmentID == "" && ticketID != "" {
			if err := tx.QueryRow(tenantCtx,
				"SELECT equipment_id FROM maintenance_tickets WHERE ticket_id = $1",
				ticketID,
			).Scan(&equipmentID); err != nil {
				return fmt.Errorf("lookup equipment for ticket: %w", err)
			}
		}

		var ticketPtr *string
		if ticketID != "" {
			ticketPtr = &ticketID
		}

		return tx.QueryRow(tenantCtx,
			`INSERT INTO maintenance_logs (org_id, ticket_id, equipment_id, action, notes, cost, performed_by)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)
			 RETURNING log_id, org_id, ticket_id::TEXT, equipment_id, action, notes, cost, performed_by, performed_at::TEXT`,
			orgID, ticketPtr, equipmentID, action, notes, cost, performedBy,
		).Scan(
			&log.LogID, &log.OrgID, &log.TicketID, &log.EquipmentID,
			&log.Action, &log.Notes, &log.Cost, &log.PerformedBy, &log.PerformedAt,
		)
	})
	if err != nil {
		return nil, fmt.Errorf("add log: %w", err)
	}
	return &log, nil
}
