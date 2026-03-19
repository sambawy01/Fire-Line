# SP4: Vendor Intelligence Design Spec

**Date:** 2026-03-19
**Status:** Approved
**Scope:** Vendor directory with spend analysis, item coverage, and simple scoring — all derived from existing data, no new tables

---

## 1. Backend — Vendor Intelligence Service

New Go service at `internal/vendor/` following existing patterns. No new migrations — all data derived from existing tables.

### Data Sources

- `ingredient_location_configs` (ilc) — vendor_name, local_cost_per_unit, ingredient_id, location_id
- `ingredients` (i) — name, cost_per_unit (fallback when no local config)
- `recipe_explosion` (re) — quantity_per_unit per menu item per ingredient
- `check_items` (ci) + `checks` (c) — actual units sold (for real spend calculation)

### Calculations

**1. Vendor List**
Distinct `vendor_name` from `ingredient_location_configs` WHERE `location_id = $1` AND `vendor_name IS NOT NULL AND vendor_name != ''`.

**2. Per-Vendor Spend**
Real spend = ingredients consumed via orders:
```sql
SELECT ilc.vendor_name,
       COUNT(DISTINCT ilc.ingredient_id) AS items_supplied,
       COALESCE(SUM(ci_qty.total_qty * re.quantity_per_unit * COALESCE(ilc.local_cost_per_unit, i.cost_per_unit))::BIGINT, 0) AS total_spend
FROM ingredient_location_configs ilc
JOIN ingredients i ON i.ingredient_id = ilc.ingredient_id AND i.org_id = ilc.org_id
LEFT JOIN recipe_explosion re ON re.ingredient_id = ilc.ingredient_id AND re.org_id = ilc.org_id
LEFT JOIN (
    SELECT ci.menu_item_id, SUM(ci.quantity) AS total_qty
    FROM check_items ci
    JOIN checks c ON c.check_id = ci.check_id AND c.org_id = ci.org_id
    WHERE c.location_id = $1 AND c.status = 'closed' AND ci.voided_at IS NULL
      AND c.closed_at >= $2 AND c.closed_at < $3 AND ci.menu_item_id IS NOT NULL
    GROUP BY ci.menu_item_id
) ci_qty ON ci_qty.menu_item_id = re.menu_item_id
WHERE ilc.location_id = $1 AND ilc.vendor_name IS NOT NULL AND ilc.vendor_name != ''
GROUP BY ilc.vendor_name
ORDER BY total_spend DESC
```

**3. Spend Percentage**
`vendor_spend / total_all_vendor_spend * 100`

**4. Simple Score (0-100)**
Two components, equally weighted:
- **Price competitiveness (50%):** For each ingredient this vendor supplies, compare their `local_cost_per_unit` to the base `ingredients.cost_per_unit`. Score = `100 - ABS(deviation_pct)`. Averaged across items. Vendors at or below base cost score higher.
- **Coverage breadth (50%):** `items_supplied / max_items_any_vendor * 100`. Vendors supplying more items score higher (operational simplicity of fewer vendors).

Final score capped at 0-100, rounded to nearest integer.

### Types

```go
type VendorAnalysis struct {
    VendorName     string  `json:"vendor_name"`
    ItemsSupplied  int     `json:"items_supplied"`
    TotalSpend     int64   `json:"total_spend"`      // cents
    SpendPct       float64 `json:"spend_pct"`         // % of total location spend
    AvgCostPerItem int64   `json:"avg_cost_per_item"` // cents (total_spend / total units from this vendor)
    Score          int     `json:"score"`             // 0-100
}

type VendorSummary struct {
    TotalVendors    int    `json:"total_vendors"`
    TotalSpend      int64  `json:"total_spend"`       // cents
    TopVendorName   string `json:"top_vendor_name"`
    TopVendorPct    float64 `json:"top_vendor_pct"`
    AvgItemsPerVendor float64 `json:"avg_items_per_vendor"`
}
```

### API Endpoints

**`GET /api/v1/vendors?location_id=X`**
- Requires JWT auth + location_id
- Optional: `from`, `to` (defaults to last 30 days, like menu)
- Returns: `{ vendors: VendorAnalysis[] }`

**`GET /api/v1/vendors/summary?location_id=X`**
- Same auth/params
- Returns: `VendorSummary`

### Service Constructor & Methods

```go
func New(pool *pgxpool.Pool, bus *event.Bus) *Service

func (s *Service) GetVendors(ctx context.Context, orgID, locationID string, from, to time.Time) ([]VendorAnalysis, error)
func (s *Service) GetSummary(ctx context.Context, orgID, locationID string, from, to time.Time) (*VendorSummary, error)
```

### Files

- `internal/vendor/vendor.go` — Service with types, GetVendors, GetSummary
- `internal/api/vendor_handler.go` — HTTP handlers (VendorHandler with RegisterRoutes)
- `cmd/fireline/main.go` — Create vendor service, register routes

## 2. Frontend — Vendor Intelligence Page

### New Files

- `web/src/pages/VendorPage.tsx`
- `web/src/hooks/useVendor.ts`
- Modify: `web/src/lib/api.ts` — Add vendorApi + types
- Modify: `web/src/App.tsx` — Add `/vendors` route
- Modify: `web/src/components/Layout.tsx` — Add Vendors nav item

### TypeScript Types (in api.ts)

```typescript
export interface VendorAnalysis {
  vendor_name: string;
  items_supplied: number;
  total_spend: number;        // cents
  spend_pct: number;
  avg_cost_per_item: number;  // cents
  score: number;              // 0-100
}

export interface VendorSummary {
  total_vendors: number;
  total_spend: number;        // cents
  top_vendor_name: string;
  top_vendor_pct: number;
  avg_items_per_vendor: number;
}
```

### API Client

```typescript
export const vendorApi = {
  getVendors(locationId: string, from?: string, to?: string) {
    const params = new URLSearchParams({ location_id: locationId });
    if (from) params.set('from', from);
    if (to) params.set('to', to);
    return request<{ vendors: VendorAnalysis[] }>(`/vendors?${params}`);
  },
  getSummary(locationId: string, from?: string, to?: string) {
    const params = new URLSearchParams({ location_id: locationId });
    if (from) params.set('from', from);
    if (to) params.set('to', to);
    return request<VendorSummary>(`/vendors/summary?${params}`);
  },
};
```

### React Query Hooks (useVendor.ts)

```typescript
export function useVendors(locationId: string | null, from?: string, to?: string)
  // staleTime 30s
export function useVendorSummary(locationId: string | null, from?: string, to?: string)
  // staleTime 30s
```

### Page Layout (VendorPage.tsx)

**Row 1 — KPI Cards (4 cards):**
- Total Vendors — count, Truck icon, gray tint
- Total Spend ($) — cents→$, DollarSign icon, red tint
- Top Vendor — name + X% of spend, Star icon, blue tint
- Avg Items/Vendor — number, Package icon, purple tint

Data from `useVendorSummary`.

**Row 2 — Vendor DataTable:**
| Column | Key | Sortable | Align | Render |
|--------|-----|----------|-------|--------|
| Vendor | vendor_name | yes | left | bold |
| Items | items_supplied | yes | right | — |
| Spend ($) | total_spend | yes | right | cents→$ |
| % of Spend | spend_pct | yes | right | X.X% |
| Avg Cost/Item | avg_cost_per_item | yes | right | cents→$ |
| Score | score | yes | center | StatusBadge (≥70→success, 40-69→warning, <40→critical) |

### Navigation

Add to `Layout.tsx` navItems after Labor:
```typescript
{ to: '/vendors', label: 'Vendors', icon: Truck }
```

Add route in `App.tsx`:
```tsx
<Route path="vendors" element={<VendorPage />} />
```

## 3. Conventions

- Backend follows existing patterns: TenantTx, handler with orgID + locationID extraction
- Use `parseMenuDateRange` pattern (defaults to last 30 days — vendor spend needs history)
- `event.Bus` in constructor for future extensibility
- All financial values in cents
- Frontend follows SP1-3 patterns: hooks with locationId guard, cents→$ helper, loading/error/empty states
- `Truck` icon from lucide-react for nav item
