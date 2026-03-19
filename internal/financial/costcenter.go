package financial

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/opsnerve/fireline/internal/tenant"
	"github.com/opsnerve/fireline/pkg/database"
)

// IngredientCost represents cost attribution for a single ingredient.
type IngredientCost struct {
	IngredientID   string  `json:"ingredient_id"`
	IngredientName string  `json:"ingredient_name"`
	TotalCost      int64   `json:"total_cost"`
	UnitCost       int64   `json:"unit_cost"`
	QuantityUsed   float64 `json:"quantity_used"`
	Unit           string  `json:"unit"`
	CostPct        float64 `json:"cost_pct"`
}

// CostCenter aggregates COGS by ingredient category.
type CostCenter struct {
	Category        string           `json:"category"`
	COGS            int64            `json:"cogs"`
	COGSPct         float64          `json:"cogs_pct"`
	RevenuePct      float64          `json:"revenue_pct"`
	IngredientCount int              `json:"ingredient_count"`
	TopIngredients  []IngredientCost `json:"top_ingredients"`
}

// GetCostCenterBreakdown returns COGS grouped by ingredient category with top ingredient detail.
func (s *Service) GetCostCenterBreakdown(ctx context.Context, orgID, locationID string, from, to time.Time) ([]CostCenter, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var centers []CostCenter

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		// Get total net revenue for the period to compute revenue percentages
		var totalRevenue int64
		err := tx.QueryRow(tenantCtx,
			`SELECT COALESCE(SUM(subtotal - discount), 0)
			 FROM checks
			 WHERE location_id = $1
			   AND closed_at >= $2 AND closed_at < $3
			   AND status = 'closed'`,
			locationID, from, to,
		).Scan(&totalRevenue)
		if err != nil {
			return fmt.Errorf("total revenue: %w", err)
		}

		// Aggregate COGS by ingredient category
		rows, err := tx.Query(tenantCtx,
			`SELECT
				COALESCE(i.category, 'uncategorized') AS category,
				CAST(SUM(ci.quantity * re.quantity_per_unit * i.cost_per_unit) AS BIGINT) AS total_cogs,
				COUNT(DISTINCT i.ingredient_id) AS ingredient_count
			 FROM check_items ci
			 JOIN checks c ON c.check_id = ci.check_id
			 JOIN menu_items mi ON mi.menu_item_id = ci.menu_item_id
			 JOIN recipe_explosion re ON re.menu_item_id = mi.menu_item_id
			 JOIN ingredients i ON i.ingredient_id = re.ingredient_id
			 WHERE c.location_id = $1
			   AND c.closed_at >= $2 AND c.closed_at < $3
			   AND c.status = 'closed'
			   AND ci.voided_at IS NULL
			 GROUP BY COALESCE(i.category, 'uncategorized')
			 ORDER BY total_cogs DESC`,
			locationID, from, to,
		)
		if err != nil {
			return fmt.Errorf("cost center breakdown: %w", err)
		}
		defer rows.Close()

		var totalCOGS int64
		type centerRow struct {
			category        string
			cogs            int64
			ingredientCount int
		}
		var rawCenters []centerRow

		for rows.Next() {
			var cr centerRow
			if err := rows.Scan(&cr.category, &cr.cogs, &cr.ingredientCount); err != nil {
				return err
			}
			totalCOGS += cr.cogs
			rawCenters = append(rawCenters, cr)
		}
		if err := rows.Err(); err != nil {
			return err
		}

		// For each category, get top 5 ingredients
		for _, cr := range rawCenters {
			cc := CostCenter{
				Category:        cr.category,
				COGS:            cr.cogs,
				IngredientCount: cr.ingredientCount,
			}
			if totalCOGS > 0 {
				cc.COGSPct = float64(cr.cogs) / float64(totalCOGS) * 100
			}
			if totalRevenue > 0 {
				cc.RevenuePct = float64(cr.cogs) / float64(totalRevenue) * 100
			}

			// Top 5 ingredients in this category
			ingRows, err := tx.Query(tenantCtx,
				`SELECT
					i.ingredient_id,
					i.name,
					CAST(SUM(ci.quantity * re.quantity_per_unit * i.cost_per_unit) AS BIGINT) AS total_cost,
					i.cost_per_unit,
					SUM(ci.quantity * re.quantity_per_unit) AS qty_used,
					i.unit
				 FROM check_items ci
				 JOIN checks c ON c.check_id = ci.check_id
				 JOIN menu_items mi ON mi.menu_item_id = ci.menu_item_id
				 JOIN recipe_explosion re ON re.menu_item_id = mi.menu_item_id
				 JOIN ingredients i ON i.ingredient_id = re.ingredient_id
				 WHERE c.location_id = $1
				   AND c.closed_at >= $2 AND c.closed_at < $3
				   AND c.status = 'closed'
				   AND ci.voided_at IS NULL
				   AND COALESCE(i.category, 'uncategorized') = $4
				 GROUP BY i.ingredient_id, i.name, i.cost_per_unit, i.unit
				 ORDER BY total_cost DESC
				 LIMIT 5`,
				locationID, from, to, cr.category,
			)
			if err != nil {
				return fmt.Errorf("top ingredients for %s: %w", cr.category, err)
			}

			for ingRows.Next() {
				var ic IngredientCost
				if err := ingRows.Scan(
					&ic.IngredientID, &ic.IngredientName,
					&ic.TotalCost, &ic.UnitCost, &ic.QuantityUsed, &ic.Unit,
				); err != nil {
					ingRows.Close()
					return err
				}
				if cr.cogs > 0 {
					ic.CostPct = float64(ic.TotalCost) / float64(cr.cogs) * 100
				}
				cc.TopIngredients = append(cc.TopIngredients, ic)
			}
			ingRows.Close()
			if err := ingRows.Err(); err != nil {
				return err
			}

			centers = append(centers, cc)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return centers, nil
}
