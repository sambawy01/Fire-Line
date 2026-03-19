# SP2: Menu Intelligence Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a menu engineering page with contribution margin analysis, popularity scoring, health classification, and channel-aware margin breakdowns.

**Architecture:** New Go service `internal/menu/` queries existing tables (menu_items, checks, check_items, recipe_explosion, ingredients, ingredient_location_configs) via TenantTx. HTTP handler exposes two endpoints. React frontend adds a Menu page with KPI cards, scatter plot, filterable DataTable, and channel detail modal.

**Tech Stack:** Go 1.22+ (pgx/v5, TenantTx), React 19, TypeScript, Tailwind CSS 4, TanStack React Query, Recharts, Zustand, Lucide icons.

**Spec:** `docs/superpowers/specs/2026-03-19-sp2-menu-intelligence-design.md`

---

### Task 1: Backend — Menu Service Types + Constructor

**Files:**
- Create: `internal/menu/menu.go`

- [ ] **Step 1: Create the service file with types and constructor**

```go
package menu

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/opsnerve/fireline/internal/event"
	"github.com/opsnerve/fireline/internal/tenant"
	"github.com/opsnerve/fireline/pkg/database"
)

const deliveryCommissionRate = 0.30

type Service struct {
	pool *pgxpool.Pool
	bus  *event.Bus
}

func New(pool *pgxpool.Pool, bus *event.Bus) *Service {
	return &Service{pool: pool, bus: bus}
}

type MenuItemAnalysis struct {
	MenuItemID       string          `json:"menu_item_id"`
	Name             string          `json:"name"`
	Category         string          `json:"category"`
	Price            int64           `json:"price"`
	FoodCost         int64           `json:"food_cost"`
	UnitsSold        int             `json:"units_sold"`
	ContribMargin    int64           `json:"contrib_margin"`
	ContribMarginPct float64         `json:"contrib_margin_pct"`
	PopularityPct    float64         `json:"popularity_pct"`
	HealthScore      float64         `json:"health_score"`
	Classification   string          `json:"classification"`
	ByChannel        []ChannelMargin `json:"by_channel"`
}

type ChannelMargin struct {
	Channel    string  `json:"channel"`
	Revenue    int64   `json:"revenue"`
	Commission int64   `json:"commission"`
	FoodCost   int64   `json:"food_cost"`
	Margin     int64   `json:"margin"`
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
	TopItem      string  `json:"top_item"`
}
```

- [ ] **Step 2: Verify it compiles**

Run: `go build ./internal/menu/`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add internal/menu/menu.go
git commit -m "feat: add menu intelligence service types and constructor"
```

---

### Task 2: Backend — AnalyzeMenuItems Method

**Files:**
- Modify: `internal/menu/menu.go`

- [ ] **Step 1: Add the AnalyzeMenuItems method**

Append to `internal/menu/menu.go`:

```go
// internal raw data fetched from DB before computing derived metrics
type rawItem struct {
	MenuItemID string
	Name       string
	Category   string
	Price      int64
	FoodCost   int64
	// channel -> units sold
	ChannelUnits map[string]int
}

func (s *Service) AnalyzeMenuItems(ctx context.Context, orgID, locationID string, from, to time.Time) ([]MenuItemAnalysis, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	itemMap := make(map[string]*rawItem)

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		// Step 1: Get all active menu items with food cost
		rows, err := tx.Query(tenantCtx, `
			SELECT mi.menu_item_id, mi.name, mi.category, mi.price,
			       COALESCE(SUM(re.quantity_per_unit * COALESCE(ilc.local_cost_per_unit, i.cost_per_unit)), 0)::BIGINT AS food_cost
			FROM menu_items mi
			LEFT JOIN recipe_explosion re ON re.menu_item_id = mi.menu_item_id AND re.org_id = mi.org_id
			LEFT JOIN ingredients i ON i.ingredient_id = re.ingredient_id AND i.org_id = mi.org_id
			LEFT JOIN ingredient_location_configs ilc
			    ON ilc.ingredient_id = i.ingredient_id AND ilc.location_id = mi.location_id AND ilc.org_id = mi.org_id
			WHERE mi.location_id = $1 AND mi.available = true
			GROUP BY mi.menu_item_id, mi.name, mi.category, mi.price
			ORDER BY mi.name
		`, locationID)
		if err != nil {
			return fmt.Errorf("query menu items: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var r rawItem
			if err := rows.Scan(&r.MenuItemID, &r.Name, &r.Category, &r.Price, &r.FoodCost); err != nil {
				return fmt.Errorf("scan menu item: %w", err)
			}
			r.ChannelUnits = make(map[string]int)
			itemMap[r.MenuItemID] = &r
		}
		if err := rows.Err(); err != nil {
			return err
		}

		// Step 2: Get sales volume per item per channel
		salesRows, err := tx.Query(tenantCtx, `
			SELECT ci.menu_item_id, c.channel, SUM(ci.quantity)::INT AS units_sold
			FROM check_items ci
			JOIN checks c ON c.check_id = ci.check_id AND c.org_id = ci.org_id
			WHERE c.location_id = $1
			  AND c.status = 'closed'
			  AND ci.voided_at IS NULL
			  AND c.closed_at >= $2 AND c.closed_at < $3
			  AND ci.menu_item_id IS NOT NULL
			GROUP BY ci.menu_item_id, c.channel
		`, locationID, from, to)
		if err != nil {
			return fmt.Errorf("query sales: %w", err)
		}
		defer salesRows.Close()

		for salesRows.Next() {
			var itemID, channel string
			var units int
			if err := salesRows.Scan(&itemID, &channel, &units); err != nil {
				return fmt.Errorf("scan sales: %w", err)
			}
			if item, ok := itemMap[itemID]; ok {
				item.ChannelUnits[channel] = units
			}
		}
		return salesRows.Err()
	})
	if err != nil {
		return nil, err
	}

	// Build items slice from itemMap (pointer data is authoritative)
	items := make([]rawItem, 0, len(itemMap))
	for _, v := range itemMap {
		items = append(items, *v)
	}

	// Compute derived metrics
	return s.computeAnalysis(items), nil
}

func totalUnits(r rawItem) int {
	total := 0
	for _, u := range r.ChannelUnits {
		total += u
	}
	return total
}

func (s *Service) computeAnalysis(items []rawItem) []MenuItemAnalysis {
	// Build category totals for popularity
	catTotals := make(map[string]int)    // category -> total units
	catCounts := make(map[string]int)    // category -> number of items
	for _, r := range items {
		catTotals[r.Category] += totalUnits(r)
		catCounts[r.Category]++
	}

	// Compute per-item analysis
	results := make([]MenuItemAnalysis, 0, len(items))
	for _, r := range items {
		units := totalUnits(r)

		// Channel margins
		channels := buildChannelMargins(r)

		// Blended contribution margin (weighted by channel units)
		var blendedMargin int64
		if units > 0 {
			var totalMarginCents int64
			for _, ch := range channels {
				totalMarginCents += ch.Margin * int64(ch.UnitsSold)
			}
			blendedMargin = totalMarginCents / int64(units)
		} else if r.Price > 0 {
			// No sales — use dine-in margin as default
			blendedMargin = r.Price - r.FoodCost
		}

		marginPct := 0.0
		if r.Price > 0 {
			marginPct = float64(blendedMargin) / float64(r.Price) * 100
		}

		// Popularity
		popPct := 0.0
		if catTotals[r.Category] > 0 {
			popPct = float64(units) / float64(catTotals[r.Category]) * 100
		}

		results = append(results, MenuItemAnalysis{
			MenuItemID:       r.MenuItemID,
			Name:             r.Name,
			Category:         r.Category,
			Price:            r.Price,
			FoodCost:         r.FoodCost,
			UnitsSold:        units,
			ContribMargin:    blendedMargin,
			ContribMarginPct: math.Round(marginPct*10) / 10,
			PopularityPct:    math.Round(popPct*10) / 10,
			ByChannel:        channels,
		})
	}

	// Classify items
	classifyItems(results, catCounts)

	return results
}

func buildChannelMargins(r rawItem) []ChannelMargin {
	var channels []ChannelMargin
	for _, ch := range []string{"dine_in", "takeout", "delivery", "drive_thru"} {
		units, ok := r.ChannelUnits[ch]
		if !ok || units == 0 {
			continue
		}
		commission := int64(0)
		if ch == "delivery" {
			commission = int64(float64(r.Price) * deliveryCommissionRate)
		}
		margin := r.Price - commission - r.FoodCost
		marginPct := 0.0
		if r.Price > 0 {
			marginPct = float64(margin) / float64(r.Price) * 100
		}
		channels = append(channels, ChannelMargin{
			Channel:    ch,
			Revenue:    r.Price,
			Commission: commission,
			FoodCost:   r.FoodCost,
			Margin:     margin,
			MarginPct:  math.Round(marginPct*10) / 10,
			UnitsSold:  units,
		})
	}
	return channels
}

func classifyItems(items []MenuItemAnalysis, catCounts map[string]int) {
	// Group by category for median calculation
	catItems := make(map[string][]int) // category -> indices
	for i, item := range items {
		catItems[item.Category] = append(catItems[item.Category], i)
	}

	for cat, indices := range catItems {
		// Compute median margin for category
		margins := make([]float64, len(indices))
		for j, idx := range indices {
			margins[j] = items[idx].ContribMarginPct
		}
		sort.Float64s(margins)
		medianMargin := margins[len(margins)/2]

		// Popularity threshold (Kasavana-Smith)
		n := catCounts[cat]
		popThreshold := 0.0
		if n > 0 {
			popThreshold = (1.0 / float64(n)) * 0.7 * 100
		}

		for _, idx := range indices {
			highMargin := items[idx].ContribMarginPct >= medianMargin
			highPop := items[idx].PopularityPct >= popThreshold

			switch {
			case highMargin && highPop:
				items[idx].Classification = "powerhouse"
			case highMargin && !highPop:
				items[idx].Classification = "hidden_gem"
			case !highMargin && highPop:
				items[idx].Classification = "crowd_pleaser"
			default:
				items[idx].Classification = "underperformer"
			}
		}
	}

	// Health score: location-wide percentile ranking (per spec)
	allMargins := make([]float64, len(items))
	allPops := make([]float64, len(items))
	for i, item := range items {
		allMargins[i] = item.ContribMarginPct
		allPops[i] = item.PopularityPct
	}
	sort.Float64s(allMargins)
	sort.Float64s(allPops)
	for i := range items {
		marginRank := percentileRank(items[i].ContribMarginPct, allMargins)
		popRank := percentileRank(items[i].PopularityPct, allPops)
		items[i].HealthScore = math.Round((marginRank*0.5+popRank*0.5)*10) / 10
	}
}

func percentileRank(value float64, sorted []float64) float64 {
	if len(sorted) <= 1 {
		return 50.0
	}
	count := 0
	for _, v := range sorted {
		if v < value {
			count++
		}
	}
	return float64(count) / float64(len(sorted)-1) * 100
}
```

- [ ] **Step 2: Verify it compiles**

Run: `go build ./internal/menu/`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add internal/menu/menu.go
git commit -m "feat: add AnalyzeMenuItems with margin, popularity, and classification"
```

---

### Task 3: Backend — GetSummary Method

**Files:**
- Modify: `internal/menu/menu.go`

- [ ] **Step 1: Add GetSummary method**

Append to `internal/menu/menu.go`:

```go
func (s *Service) GetSummary(ctx context.Context, orgID, locationID string, from, to time.Time) (*MenuSummary, error) {
	items, err := s.AnalyzeMenuItems(ctx, orgID, locationID, from, to)
	if err != nil {
		return nil, err
	}

	summary := &MenuSummary{
		TotalItems: len(items),
	}

	if len(items) == 0 {
		return summary, nil
	}

	var totalMarginPct float64
	catMap := make(map[string]*CategorySummary)

	for _, item := range items {
		totalMarginPct += item.ContribMarginPct

		switch item.Classification {
		case "powerhouse":
			summary.PowerhouseCount++
		case "underperformer":
			summary.UnderperformCount++
		}

		cs, ok := catMap[item.Category]
		if !ok {
			cs = &CategorySummary{Category: item.Category}
			catMap[item.Category] = cs
		}
		cs.ItemCount++
		cs.AvgMarginPct += item.ContribMarginPct
	}

	summary.AvgMarginPct = math.Round(totalMarginPct/float64(len(items))*10) / 10

	// Finalize category summaries
	for cat, cs := range catMap {
		cs.AvgMarginPct = math.Round(cs.AvgMarginPct/float64(cs.ItemCount)*10) / 10
		// Find top item in category
		var topMargin float64
		for _, item := range items {
			if item.Category == cat && item.ContribMarginPct > topMargin {
				topMargin = item.ContribMarginPct
				cs.TopItem = item.Name
			}
		}
		summary.Categories = append(summary.Categories, *cs)
	}

	sort.Slice(summary.Categories, func(i, j int) bool {
		return summary.Categories[i].Category < summary.Categories[j].Category
	})

	return summary, nil
}
```

- [ ] **Step 2: Verify it compiles**

Run: `go build ./internal/menu/`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add internal/menu/menu.go
git commit -m "feat: add GetSummary for menu KPI rollup"
```

---

### Task 4: Backend — Menu HTTP Handler

**Files:**
- Create: `internal/api/menu_handler.go`
- Modify: `cmd/fireline/main.go`

- [ ] **Step 1: Create menu handler**

Create `internal/api/menu_handler.go`:

```go
package api

import (
	"net/http"
	"time"

	"github.com/opsnerve/fireline/internal/menu"
	"github.com/opsnerve/fireline/internal/tenant"
)

type MenuHandler struct {
	svc *menu.Service
}

func NewMenuHandler(svc *menu.Service) *MenuHandler {
	return &MenuHandler{svc: svc}
}

func (h *MenuHandler) RegisterRoutes(mux *http.ServeMux, authMW func(http.Handler) http.Handler) {
	mux.Handle("GET /api/v1/menu/items", authMW(http.HandlerFunc(h.GetItems)))
	mux.Handle("GET /api/v1/menu/summary", authMW(http.HandlerFunc(h.GetSummary)))
}

func (h *MenuHandler) GetItems(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	locationID := r.URL.Query().Get("location_id")
	if locationID == "" {
		WriteError(w, http.StatusBadRequest, "MENU_MISSING_LOCATION", "location_id is required")
		return
	}

	from, to := parseMenuDateRange(r)
	items, err := h.svc.AnalyzeMenuItems(r.Context(), orgID, locationID, from, to)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "MENU_ANALYSIS_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *MenuHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	locationID := r.URL.Query().Get("location_id")
	if locationID == "" {
		WriteError(w, http.StatusBadRequest, "MENU_MISSING_LOCATION", "location_id is required")
		return
	}

	from, to := parseMenuDateRange(r)
	summary, err := h.svc.GetSummary(r.Context(), orgID, locationID, from, to)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "MENU_SUMMARY_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, summary)
}

// parseMenuDateRange defaults to last 30 days (unlike parseDateRange which defaults to today).
func parseMenuDateRange(r *http.Request) (time.Time, time.Time) {
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	from, err := time.Parse(time.RFC3339, fromStr)
	if err != nil {
		from = time.Now().AddDate(0, 0, -30).Truncate(24 * time.Hour)
	}
	to, err2 := time.Parse(time.RFC3339, toStr)
	if err2 != nil {
		to = time.Now()
	}
	return from, to
}
```

- [ ] **Step 2: Register in main.go**

In `cmd/fireline/main.go`:

Add import: `"github.com/opsnerve/fireline/internal/menu"`

After the `finSvc` initialization (around line 93), add:
```go
menuSvc := menu.New(pool.Raw(), bus)
```

After the `locHandler` registration (around line 147), add:
```go
menuHandler := api.NewMenuHandler(menuSvc)
menuHandler.RegisterRoutes(mux, authMW)
```

Update the `slog.Info("all modules initialized"` call to include `"menu", "ready"`.

- [ ] **Step 3: Verify build**

Run: `go build -o /dev/null ./cmd/fireline`
Expected: No errors

- [ ] **Step 4: Commit**

```bash
git add internal/api/menu_handler.go cmd/fireline/main.go
git commit -m "feat: add menu intelligence HTTP handler and wire into server"
```

---

### Task 5: Backend — Smoke Test

**Files:** None (verification only)

- [ ] **Step 1: Restart server and test endpoint**

```bash
pkill -f './fireline' 2>/dev/null; sleep 1
go build -o fireline ./cmd/fireline && ./fireline &
sleep 2
```

- [ ] **Step 2: Login and test menu items endpoint**

```bash
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"owner@bistrocloud.com","password":"DemoPassword1234!"}' \
  | python3 -c "import sys,json; print(json.load(sys.stdin)['access_token'])")

curl -s -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/v1/menu/items?location_id=a1111111-1111-1111-1111-111111111111" \
  | python3 -m json.tool | head -30
```

Expected: JSON with `items` array containing menu items with food_cost, contrib_margin, classification fields.

- [ ] **Step 3: Test summary endpoint**

```bash
curl -s -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/v1/menu/summary?location_id=a1111111-1111-1111-1111-111111111111" \
  | python3 -m json.tool
```

Expected: JSON with total_items, avg_margin_pct, powerhouse_count, categories.

- [ ] **Step 4: Commit (no changes — verification only)**

---

### Task 6: Frontend — API Types + Client + Hooks

**Files:**
- Modify: `web/src/lib/api.ts`
- Create: `web/src/hooks/useMenu.ts`

- [ ] **Step 1: Add types and menuApi to api.ts**

Add the following TypeScript types and API client at the end of `web/src/lib/api.ts` (before `export { ApiError }`):

```typescript
// Menu Intelligence
export interface MenuItemAnalysis {
  menu_item_id: string;
  name: string;
  category: string;
  price: number;
  food_cost: number;
  units_sold: number;
  contrib_margin: number;
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
  categories: CategorySummaryData[];
}

export interface CategorySummaryData {
  category: string;
  item_count: number;
  avg_margin_pct: number;
  top_item: string;
}

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

- [ ] **Step 2: Create useMenu.ts hooks**

Create `web/src/hooks/useMenu.ts`:

```typescript
import { useQuery } from '@tanstack/react-query';
import { menuApi } from '../lib/api';

export function useMenuItems(locationId: string | null, from?: string, to?: string) {
  return useQuery({
    queryKey: ['menu', 'items', locationId, from, to],
    queryFn: () => menuApi.getItems(locationId!, from, to),
    enabled: !!locationId,
    staleTime: 30_000,
  });
}

export function useMenuSummary(locationId: string | null, from?: string, to?: string) {
  return useQuery({
    queryKey: ['menu', 'summary', locationId, from, to],
    queryFn: () => menuApi.getSummary(locationId!, from, to),
    enabled: !!locationId,
    staleTime: 30_000,
  });
}
```

- [ ] **Step 3: Verify it compiles**

Run: `cd web && npx tsc --noEmit`
Expected: No errors

- [ ] **Step 4: Commit**

```bash
git add web/src/lib/api.ts web/src/hooks/useMenu.ts
git commit -m "feat: add menu intelligence API types, client, and React Query hooks"
```

---

### Task 7: Frontend — MenuPage

**Files:**
- Create: `web/src/pages/MenuPage.tsx`
- Modify: `web/src/App.tsx`
- Modify: `web/src/components/Layout.tsx`

- [ ] **Step 1: Create MenuPage.tsx**

```tsx
import { useState, useMemo } from 'react';
import {
  ScatterChart,
  Scatter,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  ReferenceLine,
  Cell,
} from 'recharts';
import {
  UtensilsCrossed,
  Star,
  TrendingUp,
  AlertTriangle,
} from 'lucide-react';
import { useLocationStore } from '../stores/location';
import { useMenuItems, useMenuSummary } from '../hooks/useMenu';
import KPICard from '../components/ui/KPICard';
import DataTable from '../components/ui/DataTable';
import type { Column } from '../components/ui/DataTable';
import StatusBadge from '../components/ui/StatusBadge';
import LoadingSpinner from '../components/ui/LoadingSpinner';
import ErrorBanner from '../components/ui/ErrorBanner';
import Modal from '../components/ui/Modal';
import type { MenuItemAnalysis } from '../lib/api';

function cents(v: number): string {
  return `$${(v / 100).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
}

const CHANNEL_LABELS: Record<string, string> = {
  dine_in: 'Dine-in',
  takeout: 'Takeout',
  delivery: 'Delivery',
  drive_thru: 'Drive-Thru',
};

const CLASS_BADGE: Record<string, { variant: 'success' | 'info' | 'warning' | 'critical'; label: string }> = {
  powerhouse: { variant: 'success', label: 'Powerhouse' },
  hidden_gem: { variant: 'info', label: 'Hidden Gem' },
  crowd_pleaser: { variant: 'warning', label: 'Crowd Pleaser' },
  underperformer: { variant: 'critical', label: 'Underperformer' },
};

const CLASS_COLORS: Record<string, string> = {
  powerhouse: '#10B981',
  hidden_gem: '#3B82F6',
  crowd_pleaser: '#F59E0B',
  underperformer: '#EF4444',
};

const columns: Column<MenuItemAnalysis>[] = [
  { key: 'name', header: 'Item', sortable: true, render: (r) => <span className="font-medium text-gray-900">{r.name}</span> },
  { key: 'category', header: 'Category', sortable: true },
  { key: 'price', header: 'Price', sortable: true, align: 'right', render: (r) => cents(r.price) },
  { key: 'units_sold', header: 'Units Sold', sortable: true, align: 'right' },
  { key: 'food_cost', header: 'Food Cost', sortable: true, align: 'right', render: (r) => cents(r.food_cost) },
  { key: 'contrib_margin', header: 'Margin ($)', sortable: true, align: 'right', render: (r) => cents(r.contrib_margin) },
  { key: 'contrib_margin_pct', header: 'Margin (%)', sortable: true, align: 'right', render: (r) => `${r.contrib_margin_pct.toFixed(1)}%` },
  { key: 'popularity_pct', header: 'Popularity', sortable: true, align: 'right', render: (r) => `${r.popularity_pct.toFixed(1)}%` },
  {
    key: 'classification',
    header: 'Class',
    sortable: true,
    align: 'center',
    render: (r) => {
      const cls = CLASS_BADGE[r.classification] ?? CLASS_BADGE.underperformer;
      return <StatusBadge variant={cls.variant}>{cls.label}</StatusBadge>;
    },
  },
];

export default function MenuPage() {
  const locationId = useLocationStore((s) => s.selectedLocationId);
  const { data: itemsData, isLoading: itemsLoading, error: itemsError, refetch: refetchItems } = useMenuItems(locationId);
  const { data: summary, isLoading: summaryLoading } = useMenuSummary(locationId);

  const [categoryFilter, setCategoryFilter] = useState('all');
  const [selectedItem, setSelectedItem] = useState<MenuItemAnalysis | null>(null);

  const items = itemsData?.items ?? [];

  const categories = useMemo(() => {
    const cats = new Set(items.map((i) => i.category));
    return ['all', ...Array.from(cats).sort()];
  }, [items]);

  const filtered = useMemo(
    () => categoryFilter === 'all' ? items : items.filter((i) => i.category === categoryFilter),
    [items, categoryFilter],
  );

  // Scatter plot data
  const scatterData = useMemo(
    () => filtered.map((item) => ({
      x: item.popularity_pct,
      y: item.contrib_margin_pct,
      name: item.name,
      price: item.price,
      units: item.units_sold,
      classification: item.classification,
      fill: CLASS_COLORS[item.classification] ?? '#9CA3AF',
    })),
    [filtered],
  );

  // Quadrant lines
  const medianMargin = useMemo(() => {
    if (filtered.length === 0) return 0;
    const sorted = [...filtered].sort((a, b) => a.contrib_margin_pct - b.contrib_margin_pct);
    return sorted[Math.floor(sorted.length / 2)].contrib_margin_pct;
  }, [filtered]);

  if (!locationId) return <LoadingSpinner fullPage />;

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold text-gray-800">Menu Intelligence</h1>
        <p className="text-sm text-gray-500 mt-1">Menu engineering analysis — last 30 days</p>
      </div>

      {itemsError && (
        <ErrorBanner
          message={itemsError instanceof Error ? itemsError.message : 'Failed to load menu data'}
          retry={() => refetchItems()}
        />
      )}

      {/* KPI Cards */}
      {summaryLoading ? (
        <div className="flex justify-center py-8"><LoadingSpinner /></div>
      ) : summary ? (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-5">
          <KPICard label="Menu Items" value={String(summary.total_items)} icon={UtensilsCrossed} iconColor="text-gray-600" bgTint="bg-gray-50" />
          <KPICard label="Avg Margin" value={`${summary.avg_margin_pct.toFixed(1)}%`} icon={TrendingUp} iconColor="text-blue-600" bgTint="bg-blue-50" />
          <KPICard label="Powerhouses" value={String(summary.powerhouse_count)} icon={Star} iconColor="text-emerald-600" bgTint="bg-emerald-50" />
          <KPICard label="Underperformers" value={String(summary.underperform_count)} icon={AlertTriangle} iconColor="text-red-600" bgTint="bg-red-50" />
        </div>
      ) : null}

      {/* Scatter Plot */}
      {!itemsLoading && scatterData.length > 0 && (
        <div className="bg-white rounded-xl border border-gray-200 p-6 shadow-sm">
          <h2 className="text-lg font-semibold text-gray-800 mb-4">Menu Engineering Matrix</h2>
          <div className="h-80">
            <ResponsiveContainer width="100%" height="100%">
              <ScatterChart margin={{ top: 10, right: 20, bottom: 20, left: 10 }}>
                <CartesianGrid strokeDasharray="3 3" stroke="#E5E7EB" />
                <XAxis
                  type="number"
                  dataKey="x"
                  name="Popularity"
                  unit="%"
                  tick={{ fontSize: 12 }}
                  label={{ value: 'Popularity %', position: 'bottom', offset: 0, fontSize: 12 }}
                />
                <YAxis
                  type="number"
                  dataKey="y"
                  name="Margin"
                  unit="%"
                  tick={{ fontSize: 12 }}
                  label={{ value: 'Margin %', angle: -90, position: 'insideLeft', fontSize: 12 }}
                />
                <ReferenceLine y={medianMargin} stroke="#9CA3AF" strokeDasharray="4 4" />
                <Tooltip
                  content={({ payload }) => {
                    if (!payload || payload.length === 0) return null;
                    const d = payload[0].payload as (typeof scatterData)[0];
                    return (
                      <div className="bg-white border border-gray-200 rounded-lg p-3 shadow-md text-sm">
                        <p className="font-semibold text-gray-800">{d.name}</p>
                        <p className="text-gray-500">Price: {cents(d.price)}</p>
                        <p className="text-gray-500">Margin: {d.y.toFixed(1)}%</p>
                        <p className="text-gray-500">Popularity: {d.x.toFixed(1)}%</p>
                        <p className="text-gray-500">Units: {d.units}</p>
                      </div>
                    );
                  }}
                />
                <Scatter data={scatterData}>
                  {scatterData.map((entry, idx) => (
                    <Cell key={idx} fill={entry.fill} />
                  ))}
                </Scatter>
              </ScatterChart>
            </ResponsiveContainer>
          </div>
          <div className="mt-3 flex flex-wrap gap-4 justify-center text-xs">
            {Object.entries(CLASS_BADGE).map(([key, { label }]) => (
              <div key={key} className="flex items-center gap-1.5">
                <div className="w-3 h-3 rounded-full" style={{ backgroundColor: CLASS_COLORS[key] }} />
                {label}
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Category Filter */}
      <div className="flex items-center gap-3">
        <label className="text-sm font-medium text-gray-700">Category:</label>
        <select
          value={categoryFilter}
          onChange={(e) => setCategoryFilter(e.target.value)}
          className="rounded-lg border border-gray-200 bg-white px-3 py-1.5 text-sm text-gray-700 focus:outline-none focus:ring-1 focus:ring-[#F97316]"
        >
          {categories.map((cat) => (
            <option key={cat} value={cat}>
              {cat === 'all' ? 'All Categories' : cat.charAt(0).toUpperCase() + cat.slice(1)}
            </option>
          ))}
        </select>
      </div>

      {/* Menu Items Table */}
      <DataTable
        columns={[...columns, {
          key: 'actions',
          header: '',
          align: 'center' as const,
          render: (r: MenuItemAnalysis) => (
            <button
              onClick={() => setSelectedItem(r)}
              className="text-sm text-[#F97316] hover:text-[#EA580C] font-medium"
            >
              Detail
            </button>
          ),
        }]}
        data={filtered}
        keyExtractor={(r) => r.menu_item_id}
        isLoading={itemsLoading}
        emptyTitle="No menu items"
        emptyDescription="No menu items found for this location."
      />

      {/* Channel Detail Modal */}
      <Modal
        open={selectedItem !== null}
        onClose={() => setSelectedItem(null)}
        title={selectedItem?.name ?? ''}
      >
        {selectedItem && (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="bg-gray-50 text-gray-500 uppercase tracking-wider text-xs">
                  <th className="px-4 py-2 font-medium text-left">Channel</th>
                  <th className="px-4 py-2 font-medium text-right">Revenue</th>
                  <th className="px-4 py-2 font-medium text-right">Commission</th>
                  <th className="px-4 py-2 font-medium text-right">Food Cost</th>
                  <th className="px-4 py-2 font-medium text-right">Margin</th>
                  <th className="px-4 py-2 font-medium text-right">Margin %</th>
                  <th className="px-4 py-2 font-medium text-right">Units</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {selectedItem.by_channel.map((ch) => (
                  <tr key={ch.channel} className="hover:bg-gray-50">
                    <td className="px-4 py-2 font-medium text-gray-800">{CHANNEL_LABELS[ch.channel] ?? ch.channel}</td>
                    <td className="px-4 py-2 text-right text-gray-700">{cents(ch.revenue)}</td>
                    <td className="px-4 py-2 text-right text-gray-700">{ch.commission > 0 ? cents(ch.commission) : '—'}</td>
                    <td className="px-4 py-2 text-right text-gray-700">{cents(ch.food_cost)}</td>
                    <td className="px-4 py-2 text-right text-gray-700">{cents(ch.margin)}</td>
                    <td className="px-4 py-2 text-right text-gray-700">{ch.margin_pct.toFixed(1)}%</td>
                    <td className="px-4 py-2 text-right text-gray-700">{ch.units_sold}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </Modal>
    </div>
  );
}
```

**Note:** The DataTable doesn't support row click. Add a "Detail" button column to the columns array to open the modal. Insert this column at the end of the `columns` array (it will reference `setSelectedItem` so it must be defined inside the component). The implementer should move the `columns` definition inside the component body or add the detail column dynamically.

- [ ] **Step 2: Add route in App.tsx**

Add import: `import MenuPage from './pages/MenuPage';`

Add route inside the Layout Route group, after the `financial` route:
```tsx
<Route path="menu" element={<MenuPage />} />
```

- [ ] **Step 3: Add nav item in Layout.tsx**

Add `UtensilsCrossed` to the lucide-react import.

Add to `navItems` array after the Financial entry:
```typescript
{ to: '/menu', label: 'Menu', icon: UtensilsCrossed },
```

- [ ] **Step 4: Verify it compiles**

Run: `cd web && npx tsc --noEmit`
Expected: No errors

- [ ] **Step 5: Commit**

```bash
git add web/src/pages/MenuPage.tsx web/src/App.tsx web/src/components/Layout.tsx
git commit -m "feat: add Menu Intelligence page with scatter plot, table, and channel detail modal"
```

---

### Task 8: Full Build + Smoke Test

**Files:** None (verification only)

- [ ] **Step 1: TypeScript type check**

Run: `cd web && npx tsc --noEmit`
Expected: No errors

- [ ] **Step 2: Go build**

Run: `cd /Users/bistrocloud/Documents/AI_Restaurant_System/fireline && go build -o /dev/null ./cmd/fireline`
Expected: No errors

- [ ] **Step 3: Frontend production build**

Run: `cd web && npm run build`
Expected: Build succeeds

- [ ] **Step 4: Restart server and verify menu endpoint**

```bash
pkill -f './fireline'; sleep 1
cd /Users/bistrocloud/Documents/AI_Restaurant_System/fireline
go build -o fireline ./cmd/fireline && ./fireline &
sleep 2
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login -H "Content-Type: application/json" -d '{"email":"owner@bistrocloud.com","password":"DemoPassword1234!"}' | python3 -c "import sys,json; print(json.load(sys.stdin)['access_token'])")
curl -s -H "Authorization: Bearer $TOKEN" "http://localhost:8080/api/v1/menu/items?location_id=a1111111-1111-1111-1111-111111111111" | python3 -c "import sys,json; d=json.load(sys.stdin); print(f'{len(d[\"items\"])} items'); [print(f'  {i[\"name\"]}: margin={i[\"contrib_margin_pct\"]}% class={i[\"classification\"]}') for i in d['items']]"
```

Expected: 6 menu items with classifications

- [ ] **Step 5: Verify frontend dev server loads menu page**

Restart frontend dev server if needed, open http://localhost:3000/menu

- [ ] **Step 6: Final commit if any uncommitted changes**

```bash
git status --short
# Only commit if there are changes
```
