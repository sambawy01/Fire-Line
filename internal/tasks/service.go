package tasks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/opsnerve/fireline/internal/event"
	"github.com/opsnerve/fireline/internal/tenant"
	"github.com/opsnerve/fireline/pkg/database"
)

// Service provides task management capabilities.
type Service struct {
	pool *pgxpool.Pool
	bus  *event.Bus
}

// New creates a new tasks service.
func New(pool *pgxpool.Pool, bus *event.Bus) *Service {
	return &Service{pool: pool, bus: bus}
}

// ─── Structs ────────────────────────────────────────────────────────────────

// Template represents a reusable task template.
type Template struct {
	TemplateID    string          `json:"template_id"`
	OrgID         string          `json:"org_id"`
	LocationID    *string         `json:"location_id"`
	Name          string          `json:"name"`
	Description   *string         `json:"description"`
	TriggerType   string          `json:"trigger_type"`
	TargetRole    *string         `json:"target_role"`
	TargetStation *string         `json:"target_station"`
	Items         json.RawMessage `json:"items"`
	Active        bool            `json:"active"`
	CreatedBy     *string         `json:"created_by"`
	CreatedAt     string          `json:"created_at"`
	UpdatedAt     string          `json:"updated_at"`
}

// TemplateInput is the input for creating a task template.
type TemplateInput struct {
	LocationID    string          `json:"location_id"`
	Name          string          `json:"name"`
	Description   *string         `json:"description"`
	TriggerType   *string         `json:"trigger_type"`
	TargetRole    *string         `json:"target_role"`
	TargetStation *string         `json:"target_station"`
	Items         json.RawMessage `json:"items"`
	Active        *bool           `json:"active"`
	CreatedBy     string          `json:"created_by"`
}

// Task represents a single task assignment.
type Task struct {
	TaskID          string          `json:"task_id"`
	OrgID           string          `json:"org_id"`
	LocationID      string          `json:"location_id"`
	TemplateID      *string         `json:"template_id"`
	Title           string          `json:"title"`
	Description     *string         `json:"description"`
	Type            string          `json:"type"`
	AssignedTo      *string         `json:"assigned_to"`
	AssignedBy      *string         `json:"assigned_by"`
	Priority        string          `json:"priority"`
	DueAt           *string         `json:"due_at"`
	Status          string          `json:"status"`
	DataEntryConfig json.RawMessage `json:"data_entry_config,omitempty"`
	DataEntryValue  json.RawMessage `json:"data_entry_value,omitempty"`
	PhotoURL        *string         `json:"photo_url"`
	CompletedAt     *string         `json:"completed_at"`
	CompletedBy     *string         `json:"completed_by"`
	CreatedAt       string          `json:"created_at"`
	UpdatedAt       string          `json:"updated_at"`
}

// TaskInput is the input for creating a task.
type TaskInput struct {
	LocationID      string          `json:"location_id"`
	TemplateID      *string         `json:"template_id"`
	Title           string          `json:"title"`
	Description     *string         `json:"description"`
	Type            *string         `json:"type"`
	AssignedTo      *string         `json:"assigned_to"`
	AssignedBy      *string         `json:"assigned_by"`
	Priority        *string         `json:"priority"`
	DueAt           *string         `json:"due_at"`
	DataEntryConfig json.RawMessage `json:"data_entry_config,omitempty"`
}

// CompleteInput is the input for completing a task.
type CompleteInput struct {
	DataEntryValue json.RawMessage `json:"data_entry_value,omitempty"`
	PhotoURL       *string         `json:"photo_url"`
	CompletedBy    string          `json:"completed_by"`
}

// Announcement represents a team announcement.
type Announcement struct {
	AnnouncementID string  `json:"announcement_id"`
	OrgID          string  `json:"org_id"`
	LocationID     string  `json:"location_id"`
	Title          string  `json:"title"`
	Body           *string `json:"body"`
	Priority       string  `json:"priority"`
	CreatedBy      *string `json:"created_by"`
	ExpiresAt      *string `json:"expires_at"`
	CreatedAt      string  `json:"created_at"`
}

// AnnouncementInput is the input for creating an announcement.
type AnnouncementInput struct {
	LocationID string  `json:"location_id"`
	Title      string  `json:"title"`
	Body       string  `json:"body"`
	Priority   *string `json:"priority"`
	CreatedBy  string  `json:"created_by"`
	ExpiresAt  *string `json:"expires_at"`
}

// ─── Task Templates ─────────────────────────────────────────────────────────

// CreateTemplate inserts a new task template.
func (s *Service) CreateTemplate(ctx context.Context, orgID string, input TemplateInput) (*Template, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var t Template

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		return tx.QueryRow(tenantCtx,
			`INSERT INTO task_templates (org_id, location_id, name, description, trigger_type,
				target_role, target_station, items, active, created_by)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, COALESCE($9, true), $10)
			 RETURNING template_id, org_id, location_id, name, COALESCE(description,''), trigger_type,
				target_role, target_station, items, active, created_by, created_at::TEXT, updated_at::TEXT`,
			orgID, input.LocationID, input.Name, input.Description, input.TriggerType,
			input.TargetRole, input.TargetStation, input.Items, input.Active, input.CreatedBy,
		).Scan(
			&t.TemplateID, &t.OrgID, &t.LocationID, &t.Name, &t.Description, &t.TriggerType,
			&t.TargetRole, &t.TargetStation, &t.Items, &t.Active, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt,
		)
	})
	if err != nil {
		return nil, fmt.Errorf("create template: %w", err)
	}
	return &t, nil
}

// ListTemplates returns task templates, optionally filtered by location.
func (s *Service) ListTemplates(ctx context.Context, orgID, locationID string) ([]Template, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var results []Template

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		query := `SELECT template_id, org_id, location_id, name, COALESCE(description,''), trigger_type,
				target_role, target_station, items, active, created_by, created_at::TEXT, updated_at::TEXT
			FROM task_templates WHERE 1=1`
		args := []any{}
		argIdx := 1

		if locationID != "" {
			query += fmt.Sprintf(" AND location_id = $%d", argIdx)
			args = append(args, locationID)
			argIdx++
		}
		query += " ORDER BY name"

		rows, err := tx.Query(tenantCtx, query, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var t Template
			if err := rows.Scan(
				&t.TemplateID, &t.OrgID, &t.LocationID, &t.Name, &t.Description, &t.TriggerType,
				&t.TargetRole, &t.TargetStation, &t.Items, &t.Active, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt,
			); err != nil {
				return err
			}
			results = append(results, t)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("list templates: %w", err)
	}
	return results, nil
}

// templateItem is an internal struct for parsing template items JSON.
type templateItem struct {
	Title           string          `json:"title"`
	Description     string          `json:"description"`
	Type            string          `json:"type"`
	Priority        string          `json:"priority"`
	DueAt           *string         `json:"due_at"`
	DataEntryConfig json.RawMessage `json:"data_entry_config,omitempty"`
}

// InstantiateTemplate reads a template and creates one task per item.
func (s *Service) InstantiateTemplate(ctx context.Context, orgID, templateID, assignedTo, assignedBy string) ([]Task, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var tasks []Task

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		// Read the template.
		var tmpl Template
		err := tx.QueryRow(tenantCtx,
			`SELECT template_id, org_id, location_id, name, COALESCE(description,''), trigger_type,
				target_role, target_station, items, active, created_by, created_at::TEXT, updated_at::TEXT
			 FROM task_templates WHERE template_id = $1`,
			templateID,
		).Scan(
			&tmpl.TemplateID, &tmpl.OrgID, &tmpl.LocationID, &tmpl.Name, &tmpl.Description, &tmpl.TriggerType,
			&tmpl.TargetRole, &tmpl.TargetStation, &tmpl.Items, &tmpl.Active, &tmpl.CreatedBy, &tmpl.CreatedAt, &tmpl.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("template not found: %w", err)
		}

		// Parse items JSON.
		var items []templateItem
		if err := json.Unmarshal(tmpl.Items, &items); err != nil {
			return fmt.Errorf("parse template items: %w", err)
		}

		// Create one task per item.
		for _, item := range items {
			var task Task
			err := tx.QueryRow(tenantCtx,
				`INSERT INTO tasks (org_id, location_id, template_id, title, description, type,
					assigned_to, assigned_by, priority, due_at, status, data_entry_config)
				 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10::TIMESTAMPTZ, 'pending', $11)
				 RETURNING task_id, org_id, location_id, template_id, title, COALESCE(description,''), type,
					assigned_to, assigned_by, priority, due_at::TEXT, status,
					data_entry_config, data_entry_value, photo_url, completed_at::TEXT, completed_by,
					created_at::TEXT, updated_at::TEXT`,
				orgID, tmpl.LocationID, templateID, item.Title, item.Description, item.Type,
				assignedTo, assignedBy, item.Priority, item.DueAt, item.DataEntryConfig,
			).Scan(
				&task.TaskID, &task.OrgID, &task.LocationID, &task.TemplateID, &task.Title, &task.Description, &task.Type,
				&task.AssignedTo, &task.AssignedBy, &task.Priority, &task.DueAt, &task.Status,
				&task.DataEntryConfig, &task.DataEntryValue, &task.PhotoURL, &task.CompletedAt, &task.CompletedBy,
				&task.CreatedAt, &task.UpdatedAt,
			)
			if err != nil {
				return fmt.Errorf("create task from template item: %w", err)
			}
			tasks = append(tasks, task)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("instantiate template: %w", err)
	}
	return tasks, nil
}

// ─── Tasks ──────────────────────────────────────────────────────────────────

// CreateTask inserts a new task.
func (s *Service) CreateTask(ctx context.Context, orgID string, input TaskInput) (*Task, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var task Task

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		return tx.QueryRow(tenantCtx,
			`INSERT INTO tasks (org_id, location_id, template_id, title, description, type,
				assigned_to, assigned_by, priority, due_at, status, data_entry_config)
			 VALUES ($1, $2, $3, $4, $5, COALESCE($6, 'general'), $7, $8, COALESCE($9, 'normal'),
				$10::TIMESTAMPTZ, 'pending', $11)
			 RETURNING task_id, org_id, location_id, template_id, title, COALESCE(description,''), type,
				assigned_to, assigned_by, priority, due_at::TEXT, status,
				data_entry_config, data_entry_value, photo_url, completed_at::TEXT, completed_by,
				created_at::TEXT, updated_at::TEXT`,
			orgID, input.LocationID, input.TemplateID, input.Title, input.Description, input.Type,
			input.AssignedTo, input.AssignedBy, input.Priority, input.DueAt, input.DataEntryConfig,
		).Scan(
			&task.TaskID, &task.OrgID, &task.LocationID, &task.TemplateID, &task.Title, &task.Description, &task.Type,
			&task.AssignedTo, &task.AssignedBy, &task.Priority, &task.DueAt, &task.Status,
			&task.DataEntryConfig, &task.DataEntryValue, &task.PhotoURL, &task.CompletedAt, &task.CompletedBy,
			&task.CreatedAt, &task.UpdatedAt,
		)
	})
	if err != nil {
		return nil, fmt.Errorf("create task: %w", err)
	}
	return &task, nil
}

// ListTasks returns tasks with optional filters.
func (s *Service) ListTasks(ctx context.Context, orgID, locationID, assignedTo, status string) ([]Task, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var results []Task

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		query := `SELECT task_id, org_id, location_id, template_id, title, COALESCE(description,''), type,
				assigned_to, assigned_by, priority, due_at::TEXT, status,
				data_entry_config, data_entry_value, photo_url, completed_at::TEXT, completed_by,
				created_at::TEXT, updated_at::TEXT
			FROM tasks WHERE 1=1`
		args := []any{}
		argIdx := 1

		if locationID != "" {
			query += fmt.Sprintf(" AND location_id = $%d", argIdx)
			args = append(args, locationID)
			argIdx++
		}
		if assignedTo != "" {
			query += fmt.Sprintf(" AND assigned_to = $%d", argIdx)
			args = append(args, assignedTo)
			argIdx++
		}
		if status != "" {
			query += fmt.Sprintf(" AND status = $%d", argIdx)
			args = append(args, status)
			argIdx++
		}
		query += " ORDER BY created_at DESC"

		rows, err := tx.Query(tenantCtx, query, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var task Task
			if err := rows.Scan(
				&task.TaskID, &task.OrgID, &task.LocationID, &task.TemplateID, &task.Title, &task.Description, &task.Type,
				&task.AssignedTo, &task.AssignedBy, &task.Priority, &task.DueAt, &task.Status,
				&task.DataEntryConfig, &task.DataEntryValue, &task.PhotoURL, &task.CompletedAt, &task.CompletedBy,
				&task.CreatedAt, &task.UpdatedAt,
			); err != nil {
				return err
			}
			results = append(results, task)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	return results, nil
}

// GetMyTasks returns pending and in_progress tasks for a specific employee.
func (s *Service) GetMyTasks(ctx context.Context, orgID, employeeID string) ([]Task, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var results []Task

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(tenantCtx,
			`SELECT task_id, org_id, location_id, template_id, title, COALESCE(description,''), type,
				assigned_to, assigned_by, priority, due_at::TEXT, status,
				data_entry_config, data_entry_value, photo_url, completed_at::TEXT, completed_by,
				created_at::TEXT, updated_at::TEXT
			FROM tasks
			WHERE assigned_to = $1 AND status IN ('pending', 'in_progress')
			ORDER BY
				CASE priority WHEN 'urgent' THEN 0 WHEN 'high' THEN 1 WHEN 'normal' THEN 2 WHEN 'low' THEN 3 ELSE 4 END,
				due_at ASC NULLS LAST,
				created_at DESC`,
			employeeID,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var task Task
			if err := rows.Scan(
				&task.TaskID, &task.OrgID, &task.LocationID, &task.TemplateID, &task.Title, &task.Description, &task.Type,
				&task.AssignedTo, &task.AssignedBy, &task.Priority, &task.DueAt, &task.Status,
				&task.DataEntryConfig, &task.DataEntryValue, &task.PhotoURL, &task.CompletedAt, &task.CompletedBy,
				&task.CreatedAt, &task.UpdatedAt,
			); err != nil {
				return err
			}
			results = append(results, task)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("get my tasks: %w", err)
	}
	return results, nil
}

// UpdateTaskStatus updates the status of a task.
func (s *Service) UpdateTaskStatus(ctx context.Context, orgID, taskID, status string) error {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		tag, err := tx.Exec(tenantCtx,
			`UPDATE tasks SET status = $1, updated_at = NOW() WHERE task_id = $2`,
			status, taskID,
		)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return fmt.Errorf("task not found")
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("update task status: %w", err)
	}
	return nil
}

// CompleteTask marks a task as completed with optional data entry and photo.
func (s *Service) CompleteTask(ctx context.Context, orgID, taskID string, input CompleteInput) (*Task, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var task Task

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		return tx.QueryRow(tenantCtx,
			`UPDATE tasks
			 SET status = 'completed', data_entry_value = $1, photo_url = $2,
				completed_by = $3, completed_at = NOW(), updated_at = NOW()
			 WHERE task_id = $4
			 RETURNING task_id, org_id, location_id, template_id, title, COALESCE(description,''), type,
				assigned_to, assigned_by, priority, due_at::TEXT, status,
				data_entry_config, data_entry_value, photo_url, completed_at::TEXT, completed_by,
				created_at::TEXT, updated_at::TEXT`,
			input.DataEntryValue, input.PhotoURL, input.CompletedBy, taskID,
		).Scan(
			&task.TaskID, &task.OrgID, &task.LocationID, &task.TemplateID, &task.Title, &task.Description, &task.Type,
			&task.AssignedTo, &task.AssignedBy, &task.Priority, &task.DueAt, &task.Status,
			&task.DataEntryConfig, &task.DataEntryValue, &task.PhotoURL, &task.CompletedAt, &task.CompletedBy,
			&task.CreatedAt, &task.UpdatedAt,
		)
	})
	if err != nil {
		return nil, fmt.Errorf("complete task: %w", err)
	}
	return &task, nil
}

// ─── Announcements ──────────────────────────────────────────────────────────

// CreateAnnouncement inserts a new announcement.
func (s *Service) CreateAnnouncement(ctx context.Context, orgID string, input AnnouncementInput) (*Announcement, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var a Announcement

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		return tx.QueryRow(tenantCtx,
			`INSERT INTO announcements (org_id, location_id, title, body, priority, created_by, expires_at)
			 VALUES ($1, $2, $3, $4, COALESCE($5, 'normal'), $6, $7::TIMESTAMPTZ)
			 RETURNING announcement_id, org_id, location_id, title, body, priority, created_by,
				expires_at::TEXT, created_at::TEXT`,
			orgID, input.LocationID, input.Title, input.Body, input.Priority, input.CreatedBy, input.ExpiresAt,
		).Scan(
			&a.AnnouncementID, &a.OrgID, &a.LocationID, &a.Title, &a.Body, &a.Priority,
			&a.CreatedBy, &a.ExpiresAt, &a.CreatedAt,
		)
	})
	if err != nil {
		return nil, fmt.Errorf("create announcement: %w", err)
	}
	return &a, nil
}

// ListAnnouncements returns active (non-expired) announcements, optionally filtered by location.
func (s *Service) ListAnnouncements(ctx context.Context, orgID, locationID string) ([]Announcement, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var results []Announcement

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		query := `SELECT announcement_id, org_id, location_id, title, body, priority, created_by,
				expires_at::TEXT, created_at::TEXT
			FROM announcements
			WHERE (expires_at IS NULL OR expires_at > NOW())`
		args := []any{}
		argIdx := 1

		if locationID != "" {
			query += fmt.Sprintf(" AND location_id = $%d", argIdx)
			args = append(args, locationID)
			argIdx++
		}
		query += " ORDER BY created_at DESC"

		rows, err := tx.Query(tenantCtx, query, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var a Announcement
			if err := rows.Scan(
				&a.AnnouncementID, &a.OrgID, &a.LocationID, &a.Title, &a.Body, &a.Priority,
				&a.CreatedBy, &a.ExpiresAt, &a.CreatedAt,
			); err != nil {
				return err
			}
			results = append(results, a)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("list announcements: %w", err)
	}
	return results, nil
}
