package financial

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/opsnerve/fireline/internal/tenant"
	"github.com/opsnerve/fireline/pkg/database"
)

// ItemCost summarizes revenue, COGS, and margin for a menu item.
type ItemCost struct {
	MenuItemID  string  `json:"menu_item_id"`
	Name        string  `json:"name"`
	Category    string  `json:"category"`
	Revenue     int64   `json:"revenue"`
	COGS        int64   `json:"cogs"`
	GrossProfit int64   `json:"gross_profit"`
	GrossMargin float64 `json:"gross_margin"`
	UnitsSold   int     `json:"units_sold"`
}

// IngredientBreakdown shows per-ingredient cost for a menu item.
type IngredientBreakdown struct {
	IngredientID    string  `json:"ingredient_id"`
	IngredientName  string  `json:"ingredient_name"`
	QuantityPerUnit float64 `json:"quantity_per_unit"`
	Unit            string  `json:"unit"`
	CostPerUnit     int64   `json:"cost_per_unit"`
	TotalCost       int64   `json:"total_cost"`
	CostPct         float64 `json:"cost_pct"`
}

// VendorPricePoint is a single PO line for a given ingredient from a vendor.
type VendorPricePoint struct {
	VendorName string    `json:"vendor_name"`
	UnitCost   int       `json:"unit_cost"`
	OrderedQty float64   `json:"ordered_qty"`
	PODate     time.Time `json:"po_date"`
	POStatus   string    `json:"po_status"`
}

// GetItemCostBreakdown returns revenue and COGS per menu item for the period.
// If category is empty, all categories are returned.
func (s *Service) GetItemCostBreakdown(ctx context.Context, orgID, locationID, category string, from, to time.Time) ([]ItemCost, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var items []ItemCost

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		var rows pgx.Rows
		var err error

		baseQuery := `SELECT
				mi.menu_item_id,
				mi.name,
				COALESCE(mi.category, 'uncategorized') AS category,
				CAST(SUM(ci.quantity * ci.unit_price) AS BIGINT) AS revenue,
				COALESCE(CAST(SUM(ci.quantity * ingredient_cost.item_cogs) AS BIGINT), 0) AS cogs,
				CAST(SUM(ci.quantity) AS INT) AS units_sold
			 FROM check_items ci
			 JOIN checks c ON c.check_id = ci.check_id
			 JOIN menu_items mi ON mi.menu_item_id = ci.menu_item_id
			 LEFT JOIN LATERAL (
				SELECT SUM(re.quantity_per_unit * i.cost_per_unit) AS item_cogs
				FROM recipe_explosion re
				JOIN ingredients i ON i.ingredient_id = re.ingredient_id
				WHERE re.menu_item_id = mi.menu_item_id
			 ) ingredient_cost ON true
			 WHERE c.location_id = $1
			   AND c.closed_at >= $2 AND c.closed_at < $3
			   AND c.status = 'closed'
			   AND ci.voided_at IS NULL`

		if category != "" {
			rows, err = tx.Query(tenantCtx,
				baseQuery+` AND COALESCE(mi.category, 'uncategorized') = $4
				 GROUP BY mi.menu_item_id, mi.name, mi.category
				 ORDER BY revenue DESC`,
				locationID, from, to, category,
			)
		} else {
			rows, err = tx.Query(tenantCtx,
				baseQuery+`
				 GROUP BY mi.menu_item_id, mi.name, mi.category
				 ORDER BY revenue DESC`,
				locationID, from, to,
			)
		}
		if err != nil {
			return fmt.Errorf("item cost breakdown: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var ic ItemCost
			if err := rows.Scan(
				&ic.MenuItemID, &ic.Name, &ic.Category,
				&ic.Revenue, &ic.COGS, &ic.UnitsSold,
			); err != nil {
				return err
			}
			ic.GrossProfit = ic.Revenue - ic.COGS
			if ic.Revenue > 0 {
				ic.GrossMargin = float64(ic.GrossProfit) / float64(ic.Revenue) * 100
			}
			items = append(items, ic)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	return items, nil
}

// GetIngredientCostBreakdown returns the ingredient recipe cost breakdown for a menu item
// over a period, using units_sold to scale total cost.
func (s *Service) GetIngredientCostBreakdown(ctx context.Context, orgID, locationID, menuItemID string, from, to time.Time) ([]IngredientBreakdown, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var breakdown []IngredientBreakdown

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		// Get units sold for this item in the period
		var unitsSold float64
		err := tx.QueryRow(tenantCtx,
			`SELECT COALESCE(SUM(ci.quantity), 0)
			 FROM check_items ci
			 JOIN checks c ON c.check_id = ci.check_id
			 WHERE c.location_id = $1
			   AND c.closed_at >= $2 AND c.closed_at < $3
			   AND c.status = 'closed'
			   AND ci.voided_at IS NULL
			   AND ci.menu_item_id = $4`,
			locationID, from, to, menuItemID,
		).Scan(&unitsSold)
		if err != nil {
			return fmt.Errorf("units sold: %w", err)
		}

		// Compute total cost per ingredient from recipe explosion
		var totalCOGS float64
		type ingRow struct {
			IngredientID   string
			IngredientName string
			QtyPerUnit     float64
			Unit           string
			CostPerUnit    int64
		}
		var ingRows []ingRow

		rows, err := tx.Query(tenantCtx,
			`SELECT
				i.ingredient_id,
				i.name,
				re.quantity_per_unit,
				i.unit,
				i.cost_per_unit
			 FROM recipe_explosion re
			 JOIN ingredients i ON i.ingredient_id = re.ingredient_id
			 WHERE re.menu_item_id = $1`,
			menuItemID,
		)
		if err != nil {
			return fmt.Errorf("recipe explosion: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var ir ingRow
			if err := rows.Scan(
				&ir.IngredientID, &ir.IngredientName,
				&ir.QtyPerUnit, &ir.Unit, &ir.CostPerUnit,
			); err != nil {
				return err
			}
			ingRows = append(ingRows, ir)
			totalCOGS += ir.QtyPerUnit * float64(ir.CostPerUnit) * unitsSold
		}
		if err := rows.Err(); err != nil {
			return err
		}

		for _, ir := range ingRows {
			lineCost := int64(ir.QtyPerUnit * float64(ir.CostPerUnit) * unitsSold)
			ib := IngredientBreakdown{
				IngredientID:    ir.IngredientID,
				IngredientName:  ir.IngredientName,
				QuantityPerUnit: ir.QtyPerUnit,
				Unit:            ir.Unit,
				CostPerUnit:     ir.CostPerUnit,
				TotalCost:       lineCost,
			}
			if totalCOGS > 0 {
				ib.CostPct = float64(lineCost) / totalCOGS * 100
			}
			breakdown = append(breakdown, ib)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return breakdown, nil
}

// GetIngredientVendorHistory returns the last 10 purchase order lines for an ingredient.
func (s *Service) GetIngredientVendorHistory(ctx context.Context, orgID, locationID, ingredientID string) ([]VendorPricePoint, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var history []VendorPricePoint

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(tenantCtx,
			`SELECT
				po.vendor_name,
				COALESCE(pol.received_unit_cost, pol.estimated_unit_cost) AS unit_cost,
				pol.ordered_qty,
				po.created_at,
				po.status
			 FROM purchase_order_lines pol
			 JOIN purchase_orders po ON po.purchase_order_id = pol.purchase_order_id
			 WHERE pol.org_id = $1
			   AND po.location_id = $2
			   AND pol.ingredient_id = $3
			 ORDER BY po.created_at DESC
			 LIMIT 10`,
			orgID, locationID, ingredientID,
		)
		if err != nil {
			return fmt.Errorf("vendor history: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var vp VendorPricePoint
			if err := rows.Scan(
				&vp.VendorName, &vp.UnitCost, &vp.OrderedQty, &vp.PODate, &vp.POStatus,
			); err != nil {
				return err
			}
			history = append(history, vp)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	return history, nil
}
