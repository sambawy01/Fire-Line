# SP2: Menu Intelligence Design Spec

**Date:** 2026-03-19
**Status:** Approved
**Scope:** Menu engineering page with contribution margins, popularity scoring, classification, and channel-aware analysis

---

## 1. Backend — Menu Intelligence Service

New Go service at `internal/menu/` following existing patterns (inventory, financial).

### Data Sources (all existing tables)

- `menu_items` — name, category, price, location_id (filter: `available = true`)
- `checks` + `check_items` — sales volume per item, per channel (filter: `checks.status = 'closed'`, `check_items.voided_at IS NULL`)
- `recipe_explosion` + `ingredients` — food cost per item
- `ingredient_location_configs` — per-location cost overrides

### Calculations

**1. Food Cost per Item**

Full join path:
```sql
SELECT mi.menu_item_id,
       COALESCE(SUM(re.quantity_per_unit * COALESCE(ilc.local_cost_per_unit, i.cost_per_unit)), 0) AS food_cost
FROM menu_items mi
LEFT JOIN recipe_explosion re ON re.menu_item_id = mi.menu_item_id AND re.org_id = mi.org_id
LEFT JOIN ingredients i ON i.ingredient_id = re.ingredient_id AND i.org_id = mi.org_id
LEFT JOIN ingredient_location_configs ilc
    ON ilc.ingredient_id = i.ingredient_id
    AND ilc.location_id = mi.location_id
    AND ilc.org_id = mi.org_id
WHERE mi.org_id = $orgID AND mi.location_id = $locationID AND mi.available = true
GROUP BY mi.menu_item_id
```
Result in cents.

**2. Units Sold per Item per Channel**

```sql
SELECT ci.menu_item_id, c.channel, SUM(ci.quantity) AS units_sold
FROM check_items ci
JOIN checks c ON c.check_id = ci.check_id AND c.org_id = ci.org_id
WHERE c.org_id = $orgID AND c.location_id = $locationID
  AND c.status = 'closed'
  AND ci.voided_at IS NULL
  AND c.closed_at BETWEEN $from AND $to
GROUP BY ci.menu_item_id, c.channel
```

**3. Contribution Margin (channel-aware)**
- Dine-in / Takeout / Drive-thru: `margin = price - food_cost` (zero commission)
- Delivery: `margin = price - (price * 0.30) - food_cost` (30% platform commission)
- Margin percentage: `margin / price * 100`
- Blended margin: weighted average across channels by units sold

**4. Popularity Score**
Per-item share of category volume over the query period:
```
popularity = (item_units_sold / category_total_units_sold) * 100
```
High popularity threshold: `(1/N) * 0.7 * 100` where N = number of active items in that category (Kasavana-Smith).

**5. Composite Health Score (0-100)**
Weighted blend:
- Margin percentile within location (50%)
- Popularity percentile within location (50%)

Each component normalized to 0-100 scale, then weighted. (Consistency/stddev removed — requires multi-day range to be meaningful; can be added when date range picker ships.)

**6. Classification**
Based on margin and popularity relative to their category medians:

| Classification | Margin | Popularity |
|---|---|---|
| **Powerhouse** | Above median | Above threshold |
| **Hidden Gem** | Above median | Below threshold |
| **Crowd Pleaser** | Below median | Above threshold |
| **Underperformer** | Below median | Below threshold |

### Types

```go
type MenuItemAnalysis struct {
    MenuItemID       string          `json:"menu_item_id"`
    Name             string          `json:"name"`
    Category         string          `json:"category"`
    Price            int64           `json:"price"`              // cents
    FoodCost         int64           `json:"food_cost"`          // cents
    UnitsSold        int             `json:"units_sold"`
    ContribMargin    int64           `json:"contrib_margin"`     // cents (blended)
    ContribMarginPct float64         `json:"contrib_margin_pct"`
    PopularityPct    float64         `json:"popularity_pct"`     // % of category volume
    HealthScore      float64         `json:"health_score"`       // 0-100
    Classification   string          `json:"classification"`     // powerhouse|hidden_gem|crowd_pleaser|underperformer
    ByChannel        []ChannelMargin `json:"by_channel"`
}

type ChannelMargin struct {
    Channel    string  `json:"channel"`
    Revenue    int64   `json:"revenue"`    // cents (item price)
    Commission int64   `json:"commission"` // cents (0 for dine_in/takeout/drive_thru)
    FoodCost   int64   `json:"food_cost"`  // cents
    Margin     int64   `json:"margin"`     // cents
    MarginPct  float64 `json:"margin_pct"`
    UnitsSold  int     `json:"units_sold"`
}

type MenuSummary struct {
    TotalItems        int               `json:"total_items"`
    AvgMarginPct      float64           `json:"avg_margin_pct"`
    PowerhouseCount   int               `json:"powerhouse_count"`
    UnderperformCount int               `json:"underperform_count"`
    Categories        []CategorySummary `json:"categories"`
}

type CategorySummary struct {
    Category     string  `json:"category"`
    ItemCount    int     `json:"item_count"`
    AvgMarginPct float64 `json:"avg_margin_pct"`
    TopItem      string  `json:"top_item"` // name of highest margin item
}
```

### API Endpoints

**`GET /api/v1/menu/items?location_id=X`**
- Requires JWT auth + location_id query param
- Optional: `from`, `to` date range (defaults to last 30 days)
- Returns: `{ items: MenuItemAnalysis[] }`

**`GET /api/v1/menu/summary?location_id=X`**
- Same auth/params
- Returns: `MenuSummary`

### Service Constructor & Methods

```go
func New(pool *pgxpool.Pool, bus *event.Bus) *Service

func (s *Service) AnalyzeMenuItems(ctx context.Context, orgID, locationID string, from, to time.Time) ([]MenuItemAnalysis, error)
func (s *Service) GetSummary(ctx context.Context, orgID, locationID string, from, to time.Time) (*MenuSummary, error)
```

### Files

- `internal/menu/menu.go` — Service with AnalyzeMenuItems, GetSummary
- `internal/menu/menu_test.go` — Unit tests
- `internal/api/menu_handler.go` — HTTP handlers (MenuHandler with RegisterRoutes)
- `cmd/fireline/main.go` — Create menu service, register routes

## 2. Frontend — Menu Intelligence Page

### New Files

- `web/src/pages/MenuPage.tsx` — Full page
- `web/src/hooks/useMenu.ts` — React Query hooks
- Modify: `web/src/lib/api.ts` — Add menuApi + types
- Modify: `web/src/App.tsx` — Add `/menu` route
- Modify: `web/src/components/Layout.tsx` — Add Menu nav item

### TypeScript Types (in api.ts)

```typescript
export interface MenuItemAnalysis {
  menu_item_id: string;
  name: string;
  category: string;
  price: number;           // cents
  food_cost: number;       // cents
  units_sold: number;
  contrib_margin: number;  // cents
  contrib_margin_pct: number;
  popularity_pct: number;
  health_score: number;
  classification: 'powerhouse' | 'hidden_gem' | 'crowd_pleaser' | 'underperformer';
  by_channel: ChannelMarginData[];
}

export interface ChannelMarginData {
  channel: string;
  revenue: number;
  commission: number;
  food_cost: number;
  margin: number;
  margin_pct: number;
  units_sold: number;
}

export interface MenuSummary {
  total_items: number;
  avg_margin_pct: number;
  powerhouse_count: number;
  underperform_count: number;
  categories: CategorySummary[];
}

export interface CategorySummary {
  category: string;
  item_count: number;
  avg_margin_pct: number;
  top_item: string;
}
```

### API Client

```typescript
export const menuApi = {
  getItems(locationId: string, from?: string, to?: string) {
    const params = new URLSearchParams({ location_id: locationId });
    if (from) params.set('from', from);
    if (to) params.set('to', to);
    return request<{ items: MenuItemAnalysis[] }>(`/menu/items?${params}`);
  },
  getSummary(locationId: string, from?: string, to?: string) {
    const params = new URLSearchParams({ location_id: locationId });
    if (from) params.set('from', from);
    if (to) params.set('to', to);
    return request<MenuSummary>(`/menu/summary?${params}`);
  },
};
```

### React Query Hooks (useMenu.ts)

```typescript
export function useMenuItems(locationId: string | null, from?: string, to?: string)
  // GET /menu/items, staleTime 30s, enabled: !!locationId

export function useMenuSummary(locationId: string | null, from?: string, to?: string)
  // GET /menu/summary, staleTime 30s, enabled: !!locationId
```

### Page Layout (MenuPage.tsx)

**Row 1 — KPI Cards (4 cards using KPICard component):**
- Total Menu Items (count)
- Avg Contribution Margin (%)
- Powerhouses (count, green tint)
- Underperformers (count, red tint)

Data from `useMenuSummary`.

**Row 2 — Scatter Plot (Recharts ScatterChart):**
- X-axis: Popularity % (share of category volume)
- Y-axis: Contribution Margin %
- Dot color by classification:
  - Powerhouse = emerald
  - Hidden Gem = blue
  - Crowd Pleaser = amber
  - Underperformer = red
- Quadrant reference lines computed client-side from items data (median margin, popularity threshold)
- Tooltip: item name, price (cents→$), margin %, units sold

**Row 3 — Category Filter:**
- Dropdown to filter by category (populated from data). Default: "All Categories"

**Row 4 — Menu Items DataTable:**
Columns:
| Column | Key | Sortable | Align | Render |
|--------|-----|----------|-------|--------|
| Item | name | yes | left | bold text |
| Category | category | yes | left | — |
| Price | price | yes | right | cents→$ |
| Units Sold | units_sold | yes | right | — |
| Food Cost | food_cost | yes | right | cents→$ |
| Margin ($) | contrib_margin | yes | right | cents→$ |
| Margin (%) | contrib_margin_pct | yes | right | X.X% |
| Popularity | popularity_pct | yes | right | X.X% |
| Class | classification | yes | center | StatusBadge |

Classification badge variants:
- powerhouse → success
- hidden_gem → info
- crowd_pleaser → warning
- underperformer → critical

**Row click → Modal (channel margin detail):**
Opens existing `Modal` component showing a table:

| Channel | Revenue | Commission | Food Cost | Margin ($) | Margin (%) | Units |
|---------|---------|------------|-----------|------------|------------|-------|

Data from the clicked item's `by_channel` array. Channel labels: dine_in→"Dine-in", takeout→"Takeout", delivery→"Delivery", drive_thru→"Drive-Thru".

### Navigation

Add to `Layout.tsx` navItems after Financial:
```typescript
{ to: '/menu', label: 'Menu', icon: UtensilsCrossed }
```

Add route in `App.tsx`:
```tsx
<Route path="menu" element={<MenuPage />} />
```

## 3. Conventions

- Backend follows existing patterns: handler extracts orgID from tenant context, locationID from query param, uses `parseDateRange` (defaults to last 30 days for menu), calls service, returns JSON
- Service constructor: `New(pool *pgxpool.Pool, bus *event.Bus)` matching inventory/financial
- Frontend follows SP1 patterns: useQuery hooks with locationId guard, cents→$ helper, LoadingSpinner/ErrorBanner/EmptyState for states
- All financial values in cents (int64 backend, number frontend)
- Delivery commission hardcoded at 30% for now (configurable later)
- Drive-thru treated same as dine-in/takeout (zero commission)
- Only `available = true` menu items included in analysis
- Only `closed` checks with non-voided items count toward sales
- Default date range: last 30 days (not today) for meaningful volume analysis
