package menu

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/opsnerve/fireline/internal/tenant"
	"github.com/opsnerve/fireline/pkg/database"
)

// SimulationResult summarises the financial impact of a simulation scenario.
type SimulationResult struct {
	SimulationType   string         `json:"simulation_type"`
	CurrentRevenue   int64          `json:"current_revenue"`
	ProjectedRevenue int64          `json:"projected_revenue"`
	RevenueDelta     int64          `json:"revenue_delta"`
	CurrentProfit    int64          `json:"current_profit"`
	ProjectedProfit  int64          `json:"projected_profit"`
	ProfitDelta      int64          `json:"profit_delta"`
	AffectedItems    []AffectedItem `json:"affected_items"`
}

// AffectedItem holds the per-item margin impact within a simulation.
type AffectedItem struct {
	MenuItemID    string `json:"menu_item_id"`
	Name          string `json:"name"`
	CurrentMargin int64  `json:"current_margin"`
	NewMargin     int64  `json:"new_margin"`
	MarginDelta   int64  `json:"margin_delta"`
}

// IngredientDependency maps an ingredient to all menu items that use it.
type IngredientDependency struct {
	IngredientID    string   `json:"ingredient_id"`
	IngredientName  string   `json:"ingredient_name"`
	MenuItemCount   int      `json:"menu_item_count"`
	MenuItems       []string `json:"menu_items"`
	TotalCostImpact int64    `json:"total_cost_impact"`
}

// CrossSellPair represents two items frequently ordered together.
type CrossSellPair struct {
	ItemAID       string  `json:"item_a_id"`
	ItemAName     string  `json:"item_a_name"`
	ItemBID       string  `json:"item_b_id"`
	ItemBName     string  `json:"item_b_name"`
	CoOccurrences int     `json:"co_occurrences"`
	Affinity      float64 `json:"affinity"`
}

// SimulatePriceChange projects the revenue and profit impact of changing the price
// of a single menu item. Uses a price-elasticity factor of -1.5.
func (s *Service) SimulatePriceChange(ctx context.Context, orgID, locationID, menuItemID string, newPrice int64) (*SimulationResult, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	var currentPrice, cogs int64
	var unitsSold int

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		// Current price and COGS.
		err := tx.QueryRow(tenantCtx,
			`SELECT
			    mi.price,
			    COALESCE(SUM(re.quantity_per_unit * COALESCE(ilc.local_cost_per_unit, i.cost_per_unit)), 0)::BIGINT
			 FROM menu_items mi
			 LEFT JOIN recipe_explosion re ON re.menu_item_id = mi.menu_item_id AND re.org_id = mi.org_id
			 LEFT JOIN ingredients i ON i.ingredient_id = re.ingredient_id AND i.org_id = mi.org_id
			 LEFT JOIN ingredient_location_configs ilc
			     ON ilc.ingredient_id = i.ingredient_id
			     AND ilc.location_id = mi.location_id
			     AND ilc.org_id = mi.org_id
			 WHERE mi.menu_item_id = $1 AND mi.location_id = $2
			 GROUP BY mi.price`,
			menuItemID, locationID,
		).Scan(&currentPrice, &cogs)
		if err != nil {
			return fmt.Errorf("query item: %w", err)
		}

		// Units sold in last 30 days.
		err = tx.QueryRow(tenantCtx,
			`SELECT COALESCE(SUM(ci.quantity), 0)::INT
			 FROM check_items ci
			 JOIN checks c ON c.check_id = ci.check_id AND c.org_id = ci.org_id
			 WHERE c.location_id = $1
			   AND ci.menu_item_id = $2
			   AND c.status = 'closed'
			   AND ci.voided_at IS NULL
			   AND c.closed_at >= now() - INTERVAL '30 days'`,
			locationID, menuItemID,
		).Scan(&unitsSold)
		if err != nil {
			return fmt.Errorf("query units sold: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Price elasticity: 1% price increase → 1.5% demand decrease.
	var priceDeltaPct float64
	if currentPrice > 0 {
		priceDeltaPct = float64(newPrice-currentPrice) / float64(currentPrice)
	}
	projectedUnits := int64(float64(unitsSold) * (1 + priceDeltaPct*-1.5))
	if projectedUnits < 0 {
		projectedUnits = 0
	}

	currentRevenue := currentPrice * int64(unitsSold)
	projectedRevenue := newPrice * projectedUnits
	currentProfit := (currentPrice - cogs) * int64(unitsSold)
	projectedProfit := (newPrice - cogs) * projectedUnits

	affected := AffectedItem{
		MenuItemID:    menuItemID,
		CurrentMargin: currentPrice - cogs,
		NewMargin:     newPrice - cogs,
		MarginDelta:   (newPrice - cogs) - (currentPrice - cogs),
	}

	return &SimulationResult{
		SimulationType:   "price_change",
		CurrentRevenue:   currentRevenue,
		ProjectedRevenue: projectedRevenue,
		RevenueDelta:     projectedRevenue - currentRevenue,
		CurrentProfit:    currentProfit,
		ProjectedProfit:  projectedProfit,
		ProfitDelta:      projectedProfit - currentProfit,
		AffectedItems:    []AffectedItem{affected},
	}, nil
}

// SimulateItemRemoval projects the revenue impact of removing a single menu item,
// and identifies ingredients that would no longer be needed.
func (s *Service) SimulateItemRemoval(ctx context.Context, orgID, locationID, menuItemID string) (*SimulationResult, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	type itemInfo struct {
		name         string
		price        int64
		cogs         int64
		unitsSold    int
		ingredientIDs []string
	}

	var info itemInfo

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		// Item price and COGS.
		err := tx.QueryRow(tenantCtx,
			`SELECT
			    mi.name,
			    mi.price,
			    COALESCE(SUM(re.quantity_per_unit * COALESCE(ilc.local_cost_per_unit, i.cost_per_unit)), 0)::BIGINT
			 FROM menu_items mi
			 LEFT JOIN recipe_explosion re ON re.menu_item_id = mi.menu_item_id AND re.org_id = mi.org_id
			 LEFT JOIN ingredients i ON i.ingredient_id = re.ingredient_id AND i.org_id = mi.org_id
			 LEFT JOIN ingredient_location_configs ilc
			     ON ilc.ingredient_id = i.ingredient_id
			     AND ilc.location_id = mi.location_id
			     AND ilc.org_id = mi.org_id
			 WHERE mi.menu_item_id = $1 AND mi.location_id = $2
			 GROUP BY mi.name, mi.price`,
			menuItemID, locationID,
		).Scan(&info.name, &info.price, &info.cogs)
		if err != nil {
			return fmt.Errorf("query item: %w", err)
		}

		// Units sold last 30 days.
		err = tx.QueryRow(tenantCtx,
			`SELECT COALESCE(SUM(ci.quantity), 0)::INT
			 FROM check_items ci
			 JOIN checks c ON c.check_id = ci.check_id AND c.org_id = ci.org_id
			 WHERE c.location_id = $1 AND ci.menu_item_id = $2
			   AND c.status = 'closed' AND ci.voided_at IS NULL
			   AND c.closed_at >= now() - INTERVAL '30 days'`,
			locationID, menuItemID,
		).Scan(&info.unitsSold)
		if err != nil {
			return fmt.Errorf("query units sold: %w", err)
		}

		// Ingredients used exclusively by this item.
		ingRows, err := tx.Query(tenantCtx,
			`SELECT re.ingredient_id
			 FROM recipe_explosion re
			 WHERE re.menu_item_id = $1
			   AND re.org_id = current_setting('app.current_org_id')::UUID
			   AND NOT EXISTS (
			       SELECT 1 FROM recipe_explosion re2
			       WHERE re2.ingredient_id = re.ingredient_id
			         AND re2.menu_item_id <> re.menu_item_id
			         AND re2.org_id = re.org_id
			   )`,
			menuItemID,
		)
		if err != nil {
			return fmt.Errorf("query exclusive ingredients: %w", err)
		}
		defer ingRows.Close()
		for ingRows.Next() {
			var ingID string
			if err := ingRows.Scan(&ingID); err != nil {
				return fmt.Errorf("scan ingredient: %w", err)
			}
			info.ingredientIDs = append(info.ingredientIDs, ingID)
		}
		return ingRows.Err()
	})
	if err != nil {
		return nil, err
	}

	currentRevenue := info.price * int64(info.unitsSold)
	currentProfit := (info.price - info.cogs) * int64(info.unitsSold)

	affected := AffectedItem{
		MenuItemID:    menuItemID,
		Name:          info.name,
		CurrentMargin: info.price - info.cogs,
		NewMargin:     0,
		MarginDelta:   -(info.price - info.cogs),
	}

	return &SimulationResult{
		SimulationType:   "item_removal",
		CurrentRevenue:   currentRevenue,
		ProjectedRevenue: 0,
		RevenueDelta:     -currentRevenue,
		CurrentProfit:    currentProfit,
		ProjectedProfit:  0,
		ProfitDelta:      -currentProfit,
		AffectedItems:    []AffectedItem{affected},
	}, nil
}

// SimulateIngredientPriceChange projects the margin impact across all menu items
// that use a given ingredient when its cost per unit changes.
func (s *Service) SimulateIngredientPriceChange(ctx context.Context, orgID, locationID, ingredientID string, newCostPerUnit int) (*SimulationResult, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	type itemImpact struct {
		menuItemID   string
		name         string
		price        int64
		currentCOGS  int64
		newCOGS      int64
		unitsSold    int
	}

	var impacts []itemImpact

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		// Find all items using this ingredient, with current and new COGS.
		rows, err := tx.Query(tenantCtx,
			`SELECT
			    mi.menu_item_id,
			    mi.name,
			    mi.price,
			    -- current total COGS (all ingredients at current cost)
			    COALESCE(SUM(re2.quantity_per_unit * COALESCE(ilc2.local_cost_per_unit, i2.cost_per_unit)), 0)::BIGINT AS current_cogs,
			    -- new COGS: substitute the changed ingredient's cost
			    COALESCE(SUM(
			        CASE
			            WHEN re2.ingredient_id = $3::UUID
			                THEN re2.quantity_per_unit * $4
			            ELSE re2.quantity_per_unit * COALESCE(ilc2.local_cost_per_unit, i2.cost_per_unit)
			        END
			    ), 0)::BIGINT AS new_cogs
			 FROM menu_items mi
			 JOIN recipe_explosion re ON re.menu_item_id = mi.menu_item_id AND re.org_id = mi.org_id
			     AND re.ingredient_id = $3::UUID
			 LEFT JOIN recipe_explosion re2 ON re2.menu_item_id = mi.menu_item_id AND re2.org_id = mi.org_id
			 LEFT JOIN ingredients i2 ON i2.ingredient_id = re2.ingredient_id AND i2.org_id = mi.org_id
			 LEFT JOIN ingredient_location_configs ilc2
			     ON ilc2.ingredient_id = i2.ingredient_id
			     AND ilc2.location_id = mi.location_id
			     AND ilc2.org_id = mi.org_id
			 WHERE mi.location_id = $1 AND mi.available = true
			 GROUP BY mi.menu_item_id, mi.name, mi.price`,
			locationID, locationID, ingredientID, newCostPerUnit,
		)
		if err != nil {
			return fmt.Errorf("query ingredient impact: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var imp itemImpact
			if err := rows.Scan(&imp.menuItemID, &imp.name, &imp.price, &imp.currentCOGS, &imp.newCOGS); err != nil {
				return fmt.Errorf("scan impact row: %w", err)
			}
			impacts = append(impacts, imp)
		}
		if err := rows.Err(); err != nil {
			return err
		}

		// Units sold per affected item.
		for i := range impacts {
			err := tx.QueryRow(tenantCtx,
				`SELECT COALESCE(SUM(ci.quantity), 0)::INT
				 FROM check_items ci
				 JOIN checks c ON c.check_id = ci.check_id AND c.org_id = ci.org_id
				 WHERE c.location_id = $1 AND ci.menu_item_id = $2
				   AND c.status = 'closed' AND ci.voided_at IS NULL
				   AND c.closed_at >= now() - INTERVAL '30 days'`,
				locationID, impacts[i].menuItemID,
			).Scan(&impacts[i].unitsSold)
			if err != nil {
				return fmt.Errorf("query units for %s: %w", impacts[i].menuItemID, err)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	var currentRevenue, projectedRevenue, currentProfit, projectedProfit int64
	affectedItems := make([]AffectedItem, 0, len(impacts))

	for _, imp := range impacts {
		units := int64(imp.unitsSold)
		currentRevenue += imp.price * units
		projectedRevenue += imp.price * units // revenue unchanged (same price)
		currentProfit += (imp.price - imp.currentCOGS) * units
		projectedProfit += (imp.price - imp.newCOGS) * units

		affectedItems = append(affectedItems, AffectedItem{
			MenuItemID:    imp.menuItemID,
			Name:          imp.name,
			CurrentMargin: imp.price - imp.currentCOGS,
			NewMargin:     imp.price - imp.newCOGS,
			MarginDelta:   (imp.price - imp.newCOGS) - (imp.price - imp.currentCOGS),
		})
	}

	return &SimulationResult{
		SimulationType:   "ingredient_price_change",
		CurrentRevenue:   currentRevenue,
		ProjectedRevenue: projectedRevenue,
		RevenueDelta:     0,
		CurrentProfit:    currentProfit,
		ProjectedProfit:  projectedProfit,
		ProfitDelta:      projectedProfit - currentProfit,
		AffectedItems:    affectedItems,
	}, nil
}

// GetIngredientDependencies returns all ingredients grouped with the menu items that use them,
// sorted by the number of dependent menu items descending.
func (s *Service) GetIngredientDependencies(ctx context.Context, orgID, locationID string) ([]IngredientDependency, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	var deps []IngredientDependency

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(tenantCtx,
			`SELECT
			    re.ingredient_id,
			    i.name AS ingredient_name,
			    COUNT(DISTINCT re.menu_item_id)::INT AS menu_item_count,
			    array_agg(DISTINCT mi.name ORDER BY mi.name) AS menu_item_names
			 FROM recipe_explosion re
			 JOIN ingredients i ON i.ingredient_id = re.ingredient_id AND i.org_id = re.org_id
			 JOIN menu_items mi ON mi.menu_item_id = re.menu_item_id AND mi.org_id = re.org_id
			 WHERE mi.location_id = $1 AND mi.available = true
			 GROUP BY re.ingredient_id, i.name
			 ORDER BY menu_item_count DESC, i.name`,
			locationID,
		)
		if err != nil {
			return fmt.Errorf("query dependencies: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var dep IngredientDependency
			if err := rows.Scan(&dep.IngredientID, &dep.IngredientName, &dep.MenuItemCount, &dep.MenuItems); err != nil {
				return fmt.Errorf("scan dependency row: %w", err)
			}
			deps = append(deps, dep)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	if deps == nil {
		deps = []IngredientDependency{}
	}
	return deps, nil
}

// GetCrossSellAffinity returns the top N pairs of menu items most frequently
// ordered together within the same check.
func (s *Service) GetCrossSellAffinity(ctx context.Context, orgID, locationID string, limit int) ([]CrossSellPair, error) {
	if limit <= 0 {
		limit = 20
	}
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	var pairs []CrossSellPair

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		// Total distinct orders at location (denominator for affinity).
		var totalOrders int
		err := tx.QueryRow(tenantCtx,
			`SELECT COUNT(DISTINCT check_id)
			 FROM checks
			 WHERE location_id = $1 AND status = 'closed'`,
			locationID,
		).Scan(&totalOrders)
		if err != nil {
			return fmt.Errorf("query total orders: %w", err)
		}
		if totalOrders == 0 {
			return nil
		}

		rows, err := tx.Query(tenantCtx,
			`SELECT
			    a.menu_item_id AS item_a_id,
			    mia.name AS item_a_name,
			    b.menu_item_id AS item_b_id,
			    mib.name AS item_b_name,
			    COUNT(*)::INT AS co_occurrences
			 FROM check_items a
			 JOIN check_items b ON b.check_id = a.check_id
			     AND b.menu_item_id > a.menu_item_id  -- canonical ordering to avoid double-counting
			     AND b.voided_at IS NULL
			 JOIN checks c ON c.check_id = a.check_id AND c.org_id = a.org_id
			 JOIN menu_items mia ON mia.menu_item_id = a.menu_item_id AND mia.org_id = a.org_id
			 JOIN menu_items mib ON mib.menu_item_id = b.menu_item_id AND mib.org_id = b.org_id
			 WHERE c.location_id = $1
			   AND c.status = 'closed'
			   AND a.voided_at IS NULL
			   AND a.menu_item_id IS NOT NULL
			   AND b.menu_item_id IS NOT NULL
			 GROUP BY a.menu_item_id, mia.name, b.menu_item_id, mib.name
			 ORDER BY co_occurrences DESC
			 LIMIT $2`,
			locationID, limit,
		)
		if err != nil {
			return fmt.Errorf("query cross-sell: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var p CrossSellPair
			if err := rows.Scan(&p.ItemAID, &p.ItemAName, &p.ItemBID, &p.ItemBName, &p.CoOccurrences); err != nil {
				return fmt.Errorf("scan cross-sell row: %w", err)
			}
			p.Affinity = float64(p.CoOccurrences) / float64(totalOrders)
			pairs = append(pairs, p)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	if pairs == nil {
		pairs = []CrossSellPair{}
	}
	return pairs, nil
}
