package maintenance

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/opsnerve/fireline/internal/tenant"
	"github.com/opsnerve/fireline/pkg/database"
)

// Equipment represents a piece of restaurant equipment.
type Equipment struct {
	EquipmentID          string  `json:"equipment_id"`
	OrgID                string  `json:"org_id"`
	LocationID           string  `json:"location_id"`
	Name                 string  `json:"name"`
	Category             string  `json:"category"`
	Make                 *string `json:"make"`
	Model                *string `json:"model"`
	SerialNumber         *string `json:"serial_number"`
	InstallDate          *string `json:"install_date"`
	WarrantyExpiry       *string `json:"warranty_expiry"`
	Status               string  `json:"status"`
	LastMaintenance       *string `json:"last_maintenance"`
	NextMaintenance       *string `json:"next_maintenance"`
	MaintenanceIntervalDays int  `json:"maintenance_interval_days"`
	HealthScore          int     `json:"health_score"`
	Notes                *string `json:"notes"`
	CreatedAt            string  `json:"created_at"`
	UpdatedAt            string  `json:"updated_at"`
}

// EquipmentInput represents the input for creating/updating equipment.
type EquipmentInput struct {
	LocationID           string  `json:"location_id"`
	Name                 string  `json:"name"`
	Category             string  `json:"category"`
	Make                 *string `json:"make"`
	Model                *string `json:"model"`
	SerialNumber         *string `json:"serial_number"`
	InstallDate          *string `json:"install_date"`
	WarrantyExpiry       *string `json:"warranty_expiry"`
	Status               string  `json:"status"`
	MaintenanceIntervalDays int  `json:"maintenance_interval_days"`
	HealthScore          int     `json:"health_score"`
	Notes                *string `json:"notes"`
}

// CreateEquipment inserts a new equipment record.
func (s *Service) CreateEquipment(ctx context.Context, orgID string, input EquipmentInput) (*Equipment, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var eq Equipment

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		return tx.QueryRow(tenantCtx,
			`INSERT INTO equipment (org_id, location_id, name, category, make, model, serial_number,
				install_date, warranty_expiry, status, maintenance_interval_days, health_score, notes)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8::DATE, $9::DATE, $10, $11, $12, $13)
			 RETURNING equipment_id, org_id, location_id, name, category, make, model, serial_number,
				install_date::TEXT, warranty_expiry::TEXT, status, last_maintenance::TEXT, next_maintenance::TEXT,
				maintenance_interval_days, health_score, notes, created_at::TEXT, updated_at::TEXT`,
			orgID, input.LocationID, input.Name, input.Category, input.Make, input.Model, input.SerialNumber,
			input.InstallDate, input.WarrantyExpiry, input.Status, input.MaintenanceIntervalDays, input.HealthScore, input.Notes,
		).Scan(
			&eq.EquipmentID, &eq.OrgID, &eq.LocationID, &eq.Name, &eq.Category,
			&eq.Make, &eq.Model, &eq.SerialNumber, &eq.InstallDate, &eq.WarrantyExpiry,
			&eq.Status, &eq.LastMaintenance, &eq.NextMaintenance,
			&eq.MaintenanceIntervalDays, &eq.HealthScore, &eq.Notes, &eq.CreatedAt, &eq.UpdatedAt,
		)
	})
	if err != nil {
		return nil, fmt.Errorf("create equipment: %w", err)
	}
	return &eq, nil
}

// ListEquipment returns equipment filtered by location, status, and category.
func (s *Service) ListEquipment(ctx context.Context, orgID, locationID, status, category string) ([]Equipment, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var results []Equipment

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		query := `SELECT equipment_id, org_id, location_id, name, category, make, model, serial_number,
				install_date::TEXT, warranty_expiry::TEXT, status, last_maintenance::TEXT, next_maintenance::TEXT,
				maintenance_interval_days, health_score, notes, created_at::TEXT, updated_at::TEXT
			FROM equipment WHERE 1=1`
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
		if category != "" {
			query += fmt.Sprintf(" AND category = $%d", argIdx)
			args = append(args, category)
			argIdx++
		}
		query += " ORDER BY name"

		rows, err := tx.Query(tenantCtx, query, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var eq Equipment
			if err := rows.Scan(
				&eq.EquipmentID, &eq.OrgID, &eq.LocationID, &eq.Name, &eq.Category,
				&eq.Make, &eq.Model, &eq.SerialNumber, &eq.InstallDate, &eq.WarrantyExpiry,
				&eq.Status, &eq.LastMaintenance, &eq.NextMaintenance,
				&eq.MaintenanceIntervalDays, &eq.HealthScore, &eq.Notes, &eq.CreatedAt, &eq.UpdatedAt,
			); err != nil {
				return err
			}
			results = append(results, eq)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("list equipment: %w", err)
	}
	return results, nil
}

// GetEquipment returns a single equipment item.
func (s *Service) GetEquipment(ctx context.Context, orgID, equipmentID string) (*Equipment, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var eq Equipment

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		return tx.QueryRow(tenantCtx,
			`SELECT equipment_id, org_id, location_id, name, category, make, model, serial_number,
				install_date::TEXT, warranty_expiry::TEXT, status, last_maintenance::TEXT, next_maintenance::TEXT,
				maintenance_interval_days, health_score, notes, created_at::TEXT, updated_at::TEXT
			 FROM equipment WHERE equipment_id = $1`,
			equipmentID,
		).Scan(
			&eq.EquipmentID, &eq.OrgID, &eq.LocationID, &eq.Name, &eq.Category,
			&eq.Make, &eq.Model, &eq.SerialNumber, &eq.InstallDate, &eq.WarrantyExpiry,
			&eq.Status, &eq.LastMaintenance, &eq.NextMaintenance,
			&eq.MaintenanceIntervalDays, &eq.HealthScore, &eq.Notes, &eq.CreatedAt, &eq.UpdatedAt,
		)
	})
	if err != nil {
		return nil, fmt.Errorf("get equipment: %w", err)
	}
	return &eq, nil
}

// UpdateEquipment updates an equipment record.
func (s *Service) UpdateEquipment(ctx context.Context, orgID, equipmentID string, input EquipmentInput) (*Equipment, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var eq Equipment

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		return tx.QueryRow(tenantCtx,
			`UPDATE equipment SET
				name = $2, category = $3, make = $4, model = $5, serial_number = $6,
				install_date = $7::DATE, warranty_expiry = $8::DATE, status = $9,
				maintenance_interval_days = $10, health_score = $11, notes = $12,
				updated_at = now()
			 WHERE equipment_id = $1
			 RETURNING equipment_id, org_id, location_id, name, category, make, model, serial_number,
				install_date::TEXT, warranty_expiry::TEXT, status, last_maintenance::TEXT, next_maintenance::TEXT,
				maintenance_interval_days, health_score, notes, created_at::TEXT, updated_at::TEXT`,
			equipmentID, input.Name, input.Category, input.Make, input.Model, input.SerialNumber,
			input.InstallDate, input.WarrantyExpiry, input.Status,
			input.MaintenanceIntervalDays, input.HealthScore, input.Notes,
		).Scan(
			&eq.EquipmentID, &eq.OrgID, &eq.LocationID, &eq.Name, &eq.Category,
			&eq.Make, &eq.Model, &eq.SerialNumber, &eq.InstallDate, &eq.WarrantyExpiry,
			&eq.Status, &eq.LastMaintenance, &eq.NextMaintenance,
			&eq.MaintenanceIntervalDays, &eq.HealthScore, &eq.Notes, &eq.CreatedAt, &eq.UpdatedAt,
		)
	})
	if err != nil {
		return nil, fmt.Errorf("update equipment: %w", err)
	}
	return &eq, nil
}

// UpdateHealthScore sets the health score and adjusts status based on the score.
func (s *Service) UpdateHealthScore(ctx context.Context, orgID, equipmentID string, score int) error {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	// Determine status based on score
	status := "operational"
	if score < 50 {
		status = "out_of_service"
	} else if score < 80 {
		status = "needs_maintenance"
	}

	return database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(tenantCtx,
			`UPDATE equipment SET health_score = $2, status = $3, updated_at = now()
			 WHERE equipment_id = $1`,
			equipmentID, score, status,
		)
		return err
	})
}

// GetOverdueMaintenanceEquipment returns equipment where next_maintenance < today.
func (s *Service) GetOverdueMaintenanceEquipment(ctx context.Context, orgID, locationID string) ([]Equipment, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var results []Equipment

	today := time.Now().Format("2006-01-02")

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		query := `SELECT equipment_id, org_id, location_id, name, category, make, model, serial_number,
				install_date::TEXT, warranty_expiry::TEXT, status, last_maintenance::TEXT, next_maintenance::TEXT,
				maintenance_interval_days, health_score, notes, created_at::TEXT, updated_at::TEXT
			FROM equipment WHERE next_maintenance < $1::DATE`
		args := []any{today}
		argIdx := 2

		if locationID != "" {
			query += fmt.Sprintf(" AND location_id = $%d", argIdx)
			args = append(args, locationID)
		}
		query += " ORDER BY next_maintenance ASC"

		rows, err := tx.Query(tenantCtx, query, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var eq Equipment
			if err := rows.Scan(
				&eq.EquipmentID, &eq.OrgID, &eq.LocationID, &eq.Name, &eq.Category,
				&eq.Make, &eq.Model, &eq.SerialNumber, &eq.InstallDate, &eq.WarrantyExpiry,
				&eq.Status, &eq.LastMaintenance, &eq.NextMaintenance,
				&eq.MaintenanceIntervalDays, &eq.HealthScore, &eq.Notes, &eq.CreatedAt, &eq.UpdatedAt,
			); err != nil {
				return err
			}
			results = append(results, eq)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("get overdue maintenance: %w", err)
	}
	return results, nil
}
