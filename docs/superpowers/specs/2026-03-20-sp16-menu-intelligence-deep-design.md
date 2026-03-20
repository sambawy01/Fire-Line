# SP16: Menu Intelligence — 5-Dimension Scoring, Classification & Simulation Sandbox

**Date:** 2026-03-20
**Status:** Approved
**Scope:** 5-dimension menu scoring, 8-class classification, radar chart visualization, simulation sandbox (price/86/removal), ingredient dependency graph, cross-sell affinity
**Maps to:** Build Plan Sprints 32-33 (Menu Intelligence)

---

## 1. Database — Migration 013

New migration: `migrations/013_menu_scoring.sql`

```sql
-- Menu item scoring and classification

ALTER TABLE menu_items
    ADD COLUMN margin_score NUMERIC(5,2) NOT NULL DEFAULT 0,
    ADD COLUMN velocity_score NUMERIC(5,2) NOT NULL DEFAULT 0,
    ADD COLUMN complexity_score NUMERIC(5,2) NOT NULL DEFAULT 0,
    ADD COLUMN satisfaction_score NUMERIC(5,2) NOT NULL DEFAULT 0,
    ADD COLUMN strategic_score NUMERIC(5,2) NOT NULL DEFAULT 0,
    ADD COLUMN classification TEXT,
    ADD COLUMN classification_changed_at TIMESTAMPTZ;

CREATE TABLE menu_simulations (
    simulation_id  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID NOT NULL REFERENCES organizations(org_id),
    location_id    UUID NOT NULL REFERENCES locations(location_id),
    simulation_type TEXT NOT NULL CHECK (simulation_type IN ('price_change', 'item_removal', 'item_addition', 'ingredient_price_change')),
    parameters     JSONB NOT NULL DEFAULT '{}',
    results        JSONB NOT NULL DEFAULT '{}',
    created_by     UUID REFERENCES users(user_id),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

RLS on menu_simulations. Index on location_id.

All 5 score columns on menu_items are 0-100 scale. Classification is one of 8 classes.

---

## 2. 5-Dimension Scoring Engine

New file: `internal/menu/scoring.go`

### Dimension Calculations

1. **Margin Score (0-100)**: `contribution_margin / max_margin_in_category * 100`
   - contribution_margin = price - COGS (from recipe explosion)

2. **Velocity Score (0-100)**: `units_sold / max_units_in_category * 100`
   - Units sold in trailing 30 days

3. **Complexity Score (0-100)**: `100 - (total_task_duration / max_duration * 100)`
   - From resource profiles: sum of duration_secs
   - Inverted: lower duration = higher score (simpler = better)
   - Default 50 if no resource profile exists

4. **Satisfaction Score (0-100)**: `100 - (void_rate * 200)`
   - void_rate = voided items / total items for this menu item
   - Capped at 0. No review data yet, so void/comp rate is the proxy.

5. **Strategic Score (0-100)**: Manager-assigned via API (default 50)

### 8-Class Classification

Based on composite scores:

| Class | Criteria |
|-------|---------|
| Powerhouse | margin≥70 AND velocity≥70 AND complexity≥50 |
| Hidden Gem | margin≥70 AND velocity<50 AND satisfaction≥60 |
| Crowd Pleaser | velocity≥70 AND margin<50 |
| Workhorse | velocity≥50 AND margin≥50 AND complexity≥70 |
| Complex Star | margin≥70 AND complexity<30 |
| Declining Star | margin≥60 AND velocity<30 (was higher) |
| Underperformer | margin<40 AND velocity<40 |
| Strategic Anchor | strategic≥70 (overrides other classifications) |

### Methods on menu `*Service`

- `ScoreMenuItems(ctx, orgID, locationID)` — calculate all 5 dimensions + classify all items
- `GetMenuItemScores(ctx, orgID, locationID)` — return scored + classified items
- `SetStrategicScore(ctx, orgID, menuItemID, score)` — manager sets strategic value
- `DetectClassificationChanges(ctx, orgID, locationID)` — compare current vs stored, emit alerts for changes

---

## 3. Simulation Sandbox

New file: `internal/menu/simulation.go`

### Simulation Types

**Price Change**: Given item + new price:
- New margin = new_price - COGS
- Projected velocity change: price_elasticity * price_change_pct (default elasticity = -1.5)
- Projected revenue = new_price * projected_units
- Projected profit = new_margin * projected_units
- Compare vs current

**Item Removal (86)**: Given item to remove:
- Revenue impact = current revenue from this item
- Shared ingredients: which other items use the same ingredients (via recipe_explosion)
- Purchasing impact: reduced ingredient demand
- If no other item uses an ingredient, flag as "ordering reduction"

**Ingredient Price Change**: Given ingredient + new cost:
- Find all menu items using this ingredient (via recipe_explosion)
- Calculate new COGS per item
- Calculate margin impact per item
- Total margin impact across menu

### Types

```go
type SimulationResult struct {
    SimulationID   string `json:"simulation_id"`
    Type           string `json:"simulation_type"`
    CurrentRevenue int64  `json:"current_revenue"`
    ProjectedRevenue int64 `json:"projected_revenue"`
    RevenueDelta   int64  `json:"revenue_delta"`
    CurrentProfit  int64  `json:"current_profit"`
    ProjectedProfit int64 `json:"projected_profit"`
    ProfitDelta    int64  `json:"profit_delta"`
    AffectedItems  []AffectedItem `json:"affected_items"`
}

type AffectedItem struct {
    MenuItemID   string  `json:"menu_item_id"`
    Name         string  `json:"name"`
    CurrentMargin int64  `json:"current_margin"`
    NewMargin    int64   `json:"new_margin"`
    MarginDelta  int64   `json:"margin_delta"`
}

type IngredientDependency struct {
    IngredientID   string   `json:"ingredient_id"`
    IngredientName string   `json:"ingredient_name"`
    MenuItems      []string `json:"menu_items"`
    TotalCost      int64    `json:"total_cost"`
}

type CrossSellPair struct {
    ItemA     string  `json:"item_a"`
    ItemAName string  `json:"item_a_name"`
    ItemB     string  `json:"item_b"`
    ItemBName string  `json:"item_b_name"`
    CoOccurrences int `json:"co_occurrences"`
    Affinity  float64 `json:"affinity"`
}
```

### Methods

- `SimulatePriceChange(ctx, orgID, locationID, menuItemID, newPrice)` → SimulationResult
- `SimulateItemRemoval(ctx, orgID, locationID, menuItemID)` → SimulationResult
- `SimulateIngredientPriceChange(ctx, orgID, locationID, ingredientID, newCostPerUnit)` → SimulationResult
- `GetIngredientDependencies(ctx, orgID, locationID)` — full dependency graph
- `GetCrossSellAffinity(ctx, orgID, locationID, limit)` — top co-ordered item pairs

---

## 4. API Endpoints

```
POST   /api/v1/menu/score                    — Recalculate all scores + classifications
GET    /api/v1/menu/scores                   — Get all scored items (query: location_id)
PUT    /api/v1/menu/scores/{id}/strategic    — Set strategic score
POST   /api/v1/menu/simulate/price           — Price change simulation
POST   /api/v1/menu/simulate/removal         — Item removal simulation
POST   /api/v1/menu/simulate/ingredient-cost — Ingredient price change simulation
GET    /api/v1/menu/dependencies             — Ingredient dependency graph
GET    /api/v1/menu/cross-sell               — Cross-sell affinity pairs
```

---

## 5. Web Dashboard — Enhanced Menu Page

Rewrite `MenuPage.tsx` with tabs:

### Tab 1: Menu Matrix
- Grid of menu item cards, each showing: name, classification badge (colored), radar chart (tiny), price, margin
- Filter by classification
- Sort by any dimension

### Tab 2: Item Detail (selected from matrix)
- Radar chart (5 axes) showing all dimension scores
- Classification badge with explanation
- Trend indicators per dimension
- Strategic score slider (editable)

### Tab 3: Simulation Sandbox
- Three simulation cards:
  1. **Price Change**: select item, enter new price, "Simulate" → show projected impact
  2. **Item Removal**: select item, "Simulate" → show revenue loss, shared ingredients, purchasing impact
  3. **Ingredient Cost**: select ingredient, enter new cost, "Simulate" → show margin impact across menu
- Results shown as before/after comparison cards

### Tab 4: Dependencies
- Ingredient dependency table: ingredient → list of menu items using it
- Cross-sell affinity: top 10 item pairs with co-occurrence count

---

## 6. RBAC
- Existing `menu:read` covers scores and simulations
- `menu:write` covers scoring, strategic score updates (existing permission)

---

## 7. Testing
- Scoring: known margin/velocity → correct score 0-100
- Classification: known score combos → correct class assignment
- Price simulation: item at $15 with $5 COGS, change to $18 → correct projected impact
- Item removal: item using unique ingredient → flagged as ordering reduction
- Cross-sell: co-occurring items in same checks → correct affinity ranking
