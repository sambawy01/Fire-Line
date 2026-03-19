package inventory

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/opsnerve/fireline/internal/event"
	"github.com/opsnerve/fireline/internal/tenant"
	"github.com/opsnerve/fireline/pkg/database"
)

// Service provides inventory intelligence capabilities.
type Service struct {
	pool *pgxpool.Pool
	bus  *event.Bus
}

// New creates a new inventory intelligence service.
func New(pool *pgxpool.Pool, bus *event.Bus) *Service {
	return &Service{pool: pool, bus: bus}
}

// RegisterHandlers subscribes to pipeline events.
func (s *Service) RegisterHandlers() {
	s.bus.Subscribe("pipeline.orders.processed", s.handleOrdersProcessed)
}

// TheoreticalUsage represents the calculated ingredient usage for a period.
type TheoreticalUsage struct {
	IngredientID   string  `json:"ingredient_id"`
	IngredientName string  `json:"ingredient_name"`
	TotalUsed      float64 `json:"total_used"`
	Unit           string  `json:"unit"`
	CostPerUnit    int64   `json:"cost_per_unit"` // cents
	TotalCost      int64   `json:"total_cost"`    // cents
}

// Variance represents the difference between theoretical and actual usage.
type Variance struct {
	IngredientID     string  `json:"ingredient_id"`
	IngredientName   string  `json:"ingredient_name"`
	TheoreticalUsage float64 `json:"theoretical_usage"`
	ActualUsage      float64 `json:"actual_usage"`
	VarianceAmount   float64 `json:"variance_amount"` // actual - theoretical
	VariancePercent  float64 `json:"variance_percent"`
	Unit             string  `json:"unit"`
	CostImpact       int64   `json:"cost_impact"` // cents
}

// PARStatus represents the current PAR level status for an ingredient.
type PARStatus struct {
	IngredientID  string  `json:"ingredient_id"`
	IngredientName string `json:"ingredient_name"`
	CurrentLevel  float64 `json:"current_level"`
	PARLevel      float64 `json:"par_level"`
	ReorderPoint  float64 `json:"reorder_point"`
	Unit          string  `json:"unit"`
	NeedsReorder  bool    `json:"needs_reorder"`
	SuggestedQty  float64 `json:"suggested_qty"` // how much to order
}

// CalculateTheoreticalUsage computes ingredient usage from sold checks in a date range.
func (s *Service) CalculateTheoreticalUsage(ctx context.Context, orgID, locationID string, from, to time.Time) ([]TheoreticalUsage, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var results []TheoreticalUsage

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(tenantCtx,
			`SELECT
				re.ingredient_id,
				i.name,
				SUM(ci.quantity * re.quantity_per_unit) AS total_used,
				re.unit,
				i.cost_per_unit,
				CAST(SUM(ci.quantity * re.quantity_per_unit * i.cost_per_unit) AS BIGINT) AS total_cost
			 FROM check_items ci
			 JOIN checks c ON c.check_id = ci.check_id
			 JOIN menu_items mi ON mi.menu_item_id = ci.menu_item_id
			 JOIN recipe_explosion re ON re.menu_item_id = mi.menu_item_id
			 JOIN ingredients i ON i.ingredient_id = re.ingredient_id
			 WHERE c.location_id = $1
			   AND c.closed_at >= $2 AND c.closed_at < $3
			   AND c.status = 'closed'
			   AND ci.voided_at IS NULL
			 GROUP BY re.ingredient_id, i.name, re.unit, i.cost_per_unit
			 ORDER BY total_cost DESC`,
			locationID, from, to,
		)
		if err != nil {
			return fmt.Errorf("query theoretical usage: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var u TheoreticalUsage
			if err := rows.Scan(&u.IngredientID, &u.IngredientName, &u.TotalUsed, &u.Unit, &u.CostPerUnit, &u.TotalCost); err != nil {
				return fmt.Errorf("scan usage: %w", err)
			}
			results = append(results, u)
		}
		return rows.Err()
	})

	return results, err
}

// CalculateVariance computes the difference between theoretical and actual inventory usage.
// actualCounts maps ingredient_id -> actual quantity used (from physical counts).
func (s *Service) CalculateVariance(theoretical []TheoreticalUsage, actualCounts map[string]float64) []Variance {
	var variances []Variance
	for _, t := range theoretical {
		actual, ok := actualCounts[t.IngredientID]
		if !ok {
			continue
		}
		diff := actual - t.TotalUsed
		pct := 0.0
		if t.TotalUsed > 0 {
			pct = (diff / t.TotalUsed) * 100
		}
		costImpact := int64(diff * float64(t.CostPerUnit))

		variances = append(variances, Variance{
			IngredientID:     t.IngredientID,
			IngredientName:   t.IngredientName,
			TheoreticalUsage: t.TotalUsed,
			ActualUsage:      actual,
			VarianceAmount:   diff,
			VariancePercent:  pct,
			Unit:             t.Unit,
			CostImpact:       costImpact,
		})
	}
	return variances
}

// GetPARStatus returns PAR level status for all ingredients at a location.
func (s *Service) GetPARStatus(ctx context.Context, orgID, locationID string, currentLevels map[string]float64) ([]PARStatus, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var results []PARStatus

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(tenantCtx,
			`SELECT
				i.ingredient_id,
				i.name,
				i.unit,
				COALESCE(ilc.par_level, 0) AS par_level,
				COALESCE(ilc.reorder_point, 0) AS reorder_point
			 FROM ingredients i
			 LEFT JOIN ingredient_location_configs ilc
			   ON ilc.ingredient_id = i.ingredient_id AND ilc.location_id = $1
			 WHERE i.org_id = $2 AND i.status = 'active'
			 ORDER BY i.name`,
			locationID, orgID,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var ps PARStatus
			var parLevel, reorderPoint float64
			if err := rows.Scan(&ps.IngredientID, &ps.IngredientName, &ps.Unit, &parLevel, &reorderPoint); err != nil {
				return err
			}
			ps.PARLevel = parLevel
			ps.ReorderPoint = reorderPoint
			ps.CurrentLevel = currentLevels[ps.IngredientID]
			ps.NeedsReorder = ps.CurrentLevel <= ps.ReorderPoint && ps.ReorderPoint > 0
			if ps.NeedsReorder && ps.PARLevel > 0 {
				ps.SuggestedQty = ps.PARLevel - ps.CurrentLevel
				if ps.SuggestedQty < 0 {
					ps.SuggestedQty = 0
				}
			}
			results = append(results, ps)
		}
		return rows.Err()
	})

	return results, err
}

// MaterializeRecipeExplosion rebuilds the recipe_explosion table for a menu item.
// This flattens the recipe DAG into ingredient quantities per 1 unit sold.
func (s *Service) MaterializeRecipeExplosion(ctx context.Context, orgID, menuItemID string) error {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	return database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		// Delete existing explosions for this menu item
		_, err := tx.Exec(tenantCtx,
			`DELETE FROM recipe_explosion WHERE menu_item_id = $1`, menuItemID,
		)
		if err != nil {
			return fmt.Errorf("delete old explosions: %w", err)
		}

		// Recursive CTE to walk the recipe DAG and compute ingredient quantities
		_, err = tx.Exec(tenantCtx,
			`WITH RECURSIVE recipe_tree AS (
				-- Base: top-level recipe for this menu item
				SELECT r.recipe_id, r.yield_quantity, 1.0::NUMERIC AS multiplier
				FROM recipes r
				WHERE r.menu_item_id = $1 AND r.status = 'active'

				UNION ALL

				-- Recursive: sub-recipes
				SELECT sr.recipe_id, sr.yield_quantity, rt.multiplier / sr.yield_quantity
				FROM recipes sr
				JOIN recipe_tree rt ON sr.parent_recipe_id = rt.recipe_id
				WHERE sr.status = 'active'
			)
			INSERT INTO recipe_explosion (org_id, menu_item_id, ingredient_id, quantity_per_unit, unit)
			SELECT $2, $1, ri.ingredient_id,
			       SUM(ri.quantity * rt.multiplier / rt.yield_quantity),
			       ri.unit
			FROM recipe_tree rt
			JOIN recipe_ingredients ri ON ri.recipe_id = rt.recipe_id
			GROUP BY ri.ingredient_id, ri.unit`,
			menuItemID, orgID,
		)
		if err != nil {
			return fmt.Errorf("materialize recipe explosion: %w", err)
		}

		return nil
	})
}

// handleOrdersProcessed recalculates metrics when orders are processed.
func (s *Service) handleOrdersProcessed(ctx context.Context, env event.Envelope) error {
	slog.Info("inventory: orders processed event received",
		"org_id", env.OrgID,
		"location_id", env.LocationID,
	)

	// Publish an event so the alerting system can check thresholds
	s.bus.Publish(ctx, event.Envelope{
		EventID:    env.EventID + ".inventory.updated",
		EventType:  "inventory.usage.updated",
		OrgID:      env.OrgID,
		LocationID: env.LocationID,
		Source:     "inventory",
	})

	return nil
}
