package inventory

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/opsnerve/fireline/internal/tenant"
	"github.com/opsnerve/fireline/pkg/database"
)

// WasteLog represents a recorded waste event for an ingredient.
type WasteLog struct {
	WasteID      string    `json:"waste_id"`
	OrgID        string    `json:"org_id"`
	LocationID   string    `json:"location_id"`
	IngredientID string    `json:"ingredient_id"`
	Name         string    `json:"name"`
	Quantity     float64   `json:"quantity"`
	Unit         string    `json:"unit"`
	Reason       string    `json:"reason"`
	LoggedBy     string    `json:"logged_by"`
	LoggedAt     time.Time `json:"logged_at"`
	Note         string    `json:"note"`
}

// WasteInput holds the data needed to log a waste event.
type WasteInput struct {
	IngredientID string  `json:"ingredient_id"`
	Quantity     float64 `json:"quantity"`
	Unit         string  `json:"unit"`
	Reason       string  `json:"reason"`
	LoggedBy     string  `json:"logged_by"`
	Note         string  `json:"note"`
}

// validReasons is the set of allowed waste reason codes.
var validReasons = map[string]bool{
	"expired":        true,
	"dropped":        true,
	"overcooked":     true,
	"contaminated":   true,
	"overproduction": true,
	"other":          true,
}

// LogWaste records a waste event for an ingredient.
func (s *Service) LogWaste(ctx context.Context, orgID, locationID string, input WasteInput) (*WasteLog, error) {
	if !validReasons[input.Reason] {
		return nil, fmt.Errorf("invalid reason: %s", input.Reason)
	}
	if input.Quantity <= 0 {
		return nil, fmt.Errorf("quantity must be positive")
	}

	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var wl WasteLog

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		err := tx.QueryRow(tenantCtx,
			`INSERT INTO waste_logs (org_id, location_id, ingredient_id, quantity, unit, reason, logged_by, note)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			 RETURNING waste_id, org_id, location_id, ingredient_id, quantity, unit, reason, logged_by, logged_at, COALESCE(note, '')`,
			orgID, locationID, input.IngredientID, input.Quantity, input.Unit, input.Reason, input.LoggedBy, input.Note,
		).Scan(&wl.WasteID, &wl.OrgID, &wl.LocationID, &wl.IngredientID,
			&wl.Quantity, &wl.Unit, &wl.Reason, &wl.LoggedBy, &wl.LoggedAt, &wl.Note)
		if err != nil {
			return fmt.Errorf("insert waste log: %w", err)
		}

		// Populate name from ingredients table
		err = tx.QueryRow(tenantCtx,
			`SELECT name FROM ingredients WHERE ingredient_id = $1`,
			input.IngredientID,
		).Scan(&wl.Name)
		if err != nil {
			return fmt.Errorf("fetch ingredient name: %w", err)
		}

		return nil
	})

	return &wl, err
}

// ListWasteLogs returns waste logs for a location within a time range.
func (s *Service) ListWasteLogs(ctx context.Context, orgID, locationID string, from, to time.Time) ([]WasteLog, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var results []WasteLog

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(tenantCtx,
			`SELECT wl.waste_id, wl.org_id, wl.location_id, wl.ingredient_id, i.name,
			        wl.quantity, wl.unit, wl.reason, wl.logged_by, wl.logged_at, COALESCE(wl.note, '')
			 FROM waste_logs wl
			 JOIN ingredients i ON i.ingredient_id = wl.ingredient_id
			 WHERE wl.location_id = $1
			   AND wl.logged_at >= $2 AND wl.logged_at < $3
			 ORDER BY wl.logged_at DESC`,
			locationID, from, to,
		)
		if err != nil {
			return fmt.Errorf("query waste logs: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var wl WasteLog
			if err := rows.Scan(&wl.WasteID, &wl.OrgID, &wl.LocationID, &wl.IngredientID, &wl.Name,
				&wl.Quantity, &wl.Unit, &wl.Reason, &wl.LoggedBy, &wl.LoggedAt, &wl.Note); err != nil {
				return fmt.Errorf("scan waste log: %w", err)
			}
			results = append(results, wl)
		}
		return rows.Err()
	})

	return results, err
}
