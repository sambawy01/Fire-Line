# SP10: Financial Intelligence Depth — Budget, Cost Centers, Anomaly Detection & Drill-Down

**Date:** 2026-03-19
**Status:** Approved
**Scope:** Budget management with variance analysis, COGS cost center breakdown, transaction anomaly detection (voids/comps/off-hours), P&L drill-down from summary to ingredient level
**Maps to:** Build Plan Sprint 19 (Financial Intelligence — Variance Analysis & Cost Tracking)

---

## 1. Database — New Migration (008)

New migration file: `migrations/008_financial_budgets.sql`

### New Tables

```sql
CREATE TABLE budgets (
    budget_id     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id        UUID NOT NULL REFERENCES organizations(org_id),
    location_id   UUID NOT NULL REFERENCES locations(location_id),
    period_type   TEXT NOT NULL CHECK (period_type IN ('daily', 'weekly', 'monthly')),
    period_start  DATE NOT NULL,
    period_end    DATE NOT NULL,
    revenue_target      BIGINT NOT NULL DEFAULT 0,
    food_cost_pct_target NUMERIC(5,2) NOT NULL DEFAULT 30.00,
    labor_cost_pct_target NUMERIC(5,2) NOT NULL DEFAULT 28.00,
    cogs_target         BIGINT NOT NULL DEFAULT 0,
    created_by    UUID REFERENCES users(user_id),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(org_id, location_id, period_type, period_start)
);
```

### RLS + Indexes

```sql
ALTER TABLE budgets ENABLE ROW LEVEL SECURITY;
ALTER TABLE budgets FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON budgets
    USING (org_id = current_setting('app.current_org_id')::UUID);
GRANT SELECT, INSERT, UPDATE, DELETE ON budgets TO fireline_app;

CREATE INDEX idx_budgets_location ON budgets(org_id, location_id, period_start DESC);
CREATE INDEX idx_budgets_period ON budgets(org_id, location_id, period_type, period_start);
```

No new tables needed for cost centers (derived from existing ingredient categories), transaction anomalies (computed in real-time), or drill-down (queries existing tables).

---

## 2. Budget Management & Variance Analysis

New file: `internal/financial/budget.go`

### Budget CRUD

- `CreateBudget(ctx, orgID, locationID, periodType, periodStart, periodEnd, revenueTarget, foodCostPctTarget, laborCostPctTarget, cogsTarget, createdBy)` — INSERT with UNIQUE conflict handling
- `GetBudget(ctx, orgID, locationID, periodType, date)` — find budget covering a specific date
- `ListBudgets(ctx, orgID, locationID, periodType)` — list all budgets for a location

### Budget vs Actual Variance

- `CalculateBudgetVariance(ctx, orgID, locationID, date)` — returns:
  - Budget values (revenue target, food cost % target, labor cost % target, COGS target)
  - Actual values (from existing CalculatePnL + labor data)
  - Variance: `actual - budget` for each metric
  - Variance %: `(actual - budget) / budget * 100`
  - Status: `on_track` (within 5%), `over` (>5% above), `under` (>5% below)

### Prior Period Comparison

- `CalculatePeriodComparison(ctx, orgID, locationID, from, to)` — returns:
  - Current period P&L
  - Same period last week
  - Same period last month
  - Same period last year (if data exists)
  - % change for each metric between periods

---

## 3. Cost Center Breakdown

New file: `internal/financial/costcenter.go`

### COGS by Category

Query COGS broken down by ingredient category (protein, produce, dairy, bakery, frozen, sauce):

```sql
SELECT i.category,
       SUM(ci.quantity * re.quantity_per_unit * i.cost_per_unit) AS category_cogs,
       COUNT(DISTINCT i.ingredient_id) AS ingredient_count
FROM check_items ci
JOIN checks c ON c.check_id = ci.check_id
JOIN recipe_explosion re ON re.menu_item_id = ci.menu_item_id
JOIN ingredients i ON i.ingredient_id = re.ingredient_id
WHERE c.location_id = $1 AND c.closed_at >= $2 AND c.closed_at < $3
  AND c.status = 'closed' AND ci.voided_at IS NULL
GROUP BY i.category
ORDER BY category_cogs DESC
```

### Cost Center Response

```go
type CostCenter struct {
    Category        string  `json:"category"`
    COGS            int64   `json:"cogs"`             // cents
    COGSPct         float64 `json:"cogs_pct"`         // % of total COGS
    RevenuePct      float64 `json:"revenue_pct"`      // % of revenue
    IngredientCount int     `json:"ingredient_count"`
    TopIngredients  []IngredientCost `json:"top_ingredients"`
}

type IngredientCost struct {
    IngredientID   string  `json:"ingredient_id"`
    IngredientName string  `json:"ingredient_name"`
    TotalCost      int64   `json:"total_cost"`  // cents
    UnitCost       int64   `json:"unit_cost"`   // cents per unit
    QuantityUsed   float64 `json:"quantity_used"`
    Unit           string  `json:"unit"`
    CostPct        float64 `json:"cost_pct"`    // % of category COGS
}
```

---

## 4. Transaction Anomaly Detection

New file: `internal/financial/anomaly.go`

Extends the existing Z-score anomaly detection with transaction-level analysis.

### Anomaly Types

1. **Excessive Voids** — void count or void $ above 2σ of trailing 30-day baseline
2. **Excessive Comps** — comp count or comp $ above 2σ
3. **Off-Hours Transactions** — orders placed outside normal business hours (before 6AM or after midnight)
4. **High Discount Rate** — discount % of gross revenue above 2σ
5. **Cash Discrepancy** — cash payment totals vs expected (if cash count data exists — stub for future)

### Detection Logic

- Query checks for the current day at the location
- For voids: count `check_items WHERE voided_at IS NOT NULL`
- For comps: sum `checks.discount WHERE discount > 0`
- For off-hours: count checks where `opened_at` hour is outside 6-24 range
- For discount rate: `SUM(discount) / SUM(subtotal) * 100`
- Compare each metric against its 30-day trailing baseline using Z-score
- Return anomalies with severity (warning at 2σ, critical at 3σ)

### Transaction Anomaly Response

```go
type TransactionAnomaly struct {
    Type          string    `json:"type"`          // "void", "comp", "off_hours", "discount_rate"
    Description   string    `json:"description"`
    CurrentValue  float64   `json:"current_value"`
    Baseline      float64   `json:"baseline"`      // 30-day mean
    ZScore        float64   `json:"z_score"`
    Severity      string    `json:"severity"`
    AffectedChecks []string `json:"affected_checks"` // check IDs for drill-down
    DetectedAt    time.Time `json:"detected_at"`
}
```

### Alert Integration

- `financial.anomaly.void` → alert if void count > 2σ
- `financial.anomaly.off_hours` → alert if any off-hours transactions detected
- `financial.anomaly.discount` → alert if discount rate > 2σ

---

## 5. P&L Drill-Down

New file: `internal/financial/drilldown.go`

### Drill-Down Chain

P&L summary → cost center (category) → menu item → ingredient → vendor/PO

Each level returns the next level of detail:

1. **Category Level** — `GetCostCenterBreakdown` (from Section 3)
2. **Menu Item Level** — `GetItemCostBreakdown(ctx, orgID, locationID, category, from, to)`:
   - Revenue, COGS, margin per menu item within a category
   - Sorted by COGS descending
3. **Ingredient Level** — `GetIngredientCostBreakdown(ctx, orgID, locationID, menuItemID, from, to)`:
   - Cost per ingredient for a specific menu item
   - Via recipe_explosion JOIN
4. **Vendor Level** — `GetIngredientVendorHistory(ctx, orgID, locationID, ingredientID)`:
   - Vendor name, cost per unit, last PO date, price trend (last 5 POs)

---

## 6. API Endpoints

### Budget

```
POST   /api/v1/financial/budgets              — Create/update budget
GET    /api/v1/financial/budgets               — List budgets (query: location_id, period_type)
GET    /api/v1/financial/budget-variance       — Get budget vs actual variance (query: location_id, date)
```

### Cost Centers

```
GET    /api/v1/financial/cost-centers          — COGS by category (query: location_id, from, to)
```

### Transaction Anomalies

```
GET    /api/v1/financial/transaction-anomalies — Detect anomalies (query: location_id)
```

### Drill-Down

```
GET    /api/v1/financial/drilldown/items       — Menu items by category (query: location_id, category, from, to)
GET    /api/v1/financial/drilldown/ingredients  — Ingredients by menu item (query: location_id, menu_item_id, from, to)
GET    /api/v1/financial/drilldown/vendor       — Vendor history for ingredient (query: location_id, ingredient_id)
```

### Period Comparison

```
GET    /api/v1/financial/period-comparison     — Current vs prior periods (query: location_id, from, to)
```

---

## 7. Web Dashboard — Enhanced Financial Page

Extend existing `FinancialPage.tsx` with tabs:

### Tab 1: P&L Overview (existing, enhanced)
- Add budget variance badges next to each KPI card (green=on track, red=over, blue=under)
- Add period comparison: "vs last week: +5.2%", "vs last month: -2.1%"

### Tab 2: Cost Centers
- Donut chart of COGS by category
- DataTable: category, COGS $, COGS %, revenue %, ingredient count
- Click category → expand to show top ingredients with costs
- Click ingredient → show vendor price history

### Tab 3: Anomalies
- Enhanced anomaly cards (existing Z-score anomalies + new transaction anomalies)
- Transaction anomalies: void count, comp $, off-hours, discount rate
- Each card shows severity badge, current vs baseline, affected check count
- Click to drill down to affected checks

### Tab 4: Budget
- Budget entry form: set revenue target, food cost % target, labor cost % target for daily/weekly/monthly periods
- Budget vs actual variance table with color-coded cells
- Sparkline trend showing variance over time

---

## 8. RBAC

- `financial:budget` — create/edit budgets (roles: `gm`, `owner`)
- Existing `financial:read` covers all read endpoints

---

## 9. Testing Strategy

### Backend Tests
- Budget CRUD: create, get, list, unique constraint
- Budget variance: known actuals vs budget → correct variance calculation
- Cost center breakdown: verify COGS sums by category
- Transaction anomaly: mock void/comp data → correct Z-score detection
- Drill-down queries: category → item → ingredient chain
- Period comparison: current vs prior period delta calculation

---

## 10. Dependencies

- No new Go dependencies
- `recharts` already available in web dashboard (for donut chart, use PieChart)
- No tablet changes in this sprint
