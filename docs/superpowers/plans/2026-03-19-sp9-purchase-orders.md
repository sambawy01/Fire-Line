# SP9: Purchase Orders, PAR Breach Detection & Delivery Receiving — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the purchase order system with auto-generation from PAR breaches, manager approval workflow, line-by-line delivery receiving on the tablet, and a PO management web dashboard.

**Architecture:** New PO tables (migration 007), purchasing engine triggered on count submission, PO CRUD + receiving API endpoints, web PO management page with one-tap approve, tablet receiving screen replacing Tasks placeholder. Feeds discrepancy data back into variance engine and alerts.

**Tech Stack:** Go 1.22, PostgreSQL 16, React/TypeScript/Tailwind (web), Expo/React Native/Zustand (tablet)

---

## File Map

### Backend — New Files
| File | Responsibility |
|------|---------------|
| `migrations/007_purchase_orders.sql` | purchase_orders + purchase_order_lines tables, ALTER ingredient_location_configs, RLS, indexes |
| `internal/inventory/purchasing.go` | PO types, CRUD, auto-generation engine, receiving logic, PAR breach query |
| `internal/inventory/purchasing_test.go` | Unit tests for variance flag computation, 2% tolerance, avg_daily_usage |
| `internal/api/po_handler.go` | HTTP handlers for all PO endpoints |

### Backend — Modified Files
| File | Change |
|------|--------|
| `internal/auth/rbac.go` | Add `inventory:purchase` and `inventory:receive` permissions |
| `internal/auth/rbac_test.go` | Add test cases for new permissions |
| `internal/api/handlers.go:30-34` | Register PO routes in InventoryHandler.RegisterRoutes |
| `internal/api/counting_handler.go` | Call auto-generation engine after count submission + variance calc |

### Web Dashboard — New/Modified Files
| File | Change |
|------|--------|
| `web/src/pages/PurchaseOrdersPage.tsx` | New page: suggested POs, active, history, detail modal |
| `web/src/hooks/usePurchaseOrders.ts` | React Query hooks for PO endpoints |
| `web/src/lib/api.ts` | PO types + API methods |
| `web/src/components/Layout.tsx:24-36` | Add Purchase Orders nav item after Inventory |
| `web/src/App.tsx:40-51` | Add PO route |
| `web/src/pages/InventoryPage.tsx` | Add PAR breach warning banner |

### Tablet — New/Modified Files
| File | Change |
|------|--------|
| `tablet/stores/receive.ts` | Zustand store for receiving flow |
| `tablet/app/(tabs)/receive.tsx` | Receiving screen (replaces tasks.tsx) |
| `tablet/components/ReceiveRow.tsx` | Line-by-line receiving input row |
| `tablet/app/(tabs)/_layout.tsx` | Replace Tasks tab with Receive tab |

---

## Task 1: Migration 007 — Purchase Orders Tables

**Files:**
- Create: `migrations/007_purchase_orders.sql`

- [ ] **Step 1: Write migration file**

```sql
-- Purchase orders and delivery receiving

CREATE TABLE purchase_orders (
    purchase_order_id  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id             UUID NOT NULL REFERENCES organizations(org_id),
    location_id        UUID NOT NULL REFERENCES locations(location_id),
    vendor_name        TEXT NOT NULL,
    status             TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'approved', 'received', 'cancelled')),
    source             TEXT NOT NULL DEFAULT 'manual' CHECK (source IN ('manual', 'system_recommended')),
    suggested_at       TIMESTAMPTZ,
    approved_by        UUID REFERENCES users(user_id),
    approved_at        TIMESTAMPTZ,
    received_by        UUID REFERENCES users(user_id),  -- uses user_id (PIN auth resolves employee to linked user_id)
    received_at        TIMESTAMPTZ,
    total_estimated    BIGINT NOT NULL DEFAULT 0,
    total_actual       BIGINT NOT NULL DEFAULT 0,
    notes              TEXT,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE purchase_order_lines (
    po_line_id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id              UUID NOT NULL REFERENCES organizations(org_id),
    purchase_order_id   UUID NOT NULL REFERENCES purchase_orders(purchase_order_id),
    ingredient_id       UUID NOT NULL REFERENCES ingredients(ingredient_id),
    ordered_qty         NUMERIC(12,4) NOT NULL,
    ordered_unit        TEXT NOT NULL,
    estimated_unit_cost INT NOT NULL DEFAULT 0,
    received_qty        NUMERIC(12,4),
    received_unit_cost  INT,
    variance_qty        NUMERIC(12,4),
    variance_flag       TEXT CHECK (variance_flag IN ('exact', 'short', 'over', 'not_received')),
    received_at         TIMESTAMPTZ,
    note                TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Add lead time and usage tracking to ingredient configs
ALTER TABLE ingredient_location_configs
    ADD COLUMN lead_time_days INT NOT NULL DEFAULT 1,
    ADD COLUMN avg_daily_usage NUMERIC(12,4) NOT NULL DEFAULT 0;

-- Indexes
CREATE INDEX idx_po_location ON purchase_orders(org_id, location_id, created_at DESC);
CREATE INDEX idx_po_status ON purchase_orders(org_id, status);
CREATE INDEX idx_po_vendor ON purchase_orders(org_id, vendor_name);
CREATE INDEX idx_po_lines_po ON purchase_order_lines(purchase_order_id);
CREATE INDEX idx_po_lines_ingredient ON purchase_order_lines(org_id, ingredient_id);

-- RLS: purchase_orders
ALTER TABLE purchase_orders ENABLE ROW LEVEL SECURITY;
ALTER TABLE purchase_orders FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON purchase_orders
    USING (org_id = current_setting('app.current_org_id')::UUID);
GRANT SELECT, INSERT, UPDATE, DELETE ON purchase_orders TO fireline_app;

-- RLS: purchase_order_lines
ALTER TABLE purchase_order_lines ENABLE ROW LEVEL SECURITY;
ALTER TABLE purchase_order_lines FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON purchase_order_lines
    USING (org_id = current_setting('app.current_org_id')::UUID);
GRANT SELECT, INSERT, UPDATE, DELETE ON purchase_order_lines TO fireline_app;
```

- [ ] **Step 2: Update atlas hash and apply migration**

```bash
cd /Users/bistrocloud/Documents/AI_Restaurant_System/fireline
atlas migrate hash --dir file://migrations
atlas migrate apply --dir file://migrations --url "postgres://fireline:fireline@localhost:5432/fireline?sslmode=disable"
```

- [ ] **Step 3: Verify tables and columns**

```bash
docker exec fireline-postgres-1 psql -U fireline -d fireline -c "\dt purchase_*" -c "\d ingredient_location_configs" | grep -E "lead_time|avg_daily"
```

- [ ] **Step 4: Commit**

```bash
git add migrations/007_purchase_orders.sql migrations/atlas.sum
git commit -m "feat: add migration 007 — purchase orders, PO lines, lead time and avg usage columns"
```

---

## Task 2: RBAC Permissions for POs

**Files:**
- Modify: `internal/auth/rbac.go`
- Modify: `internal/auth/rbac_test.go`

- [ ] **Step 1: Add test cases to existing TestRole_HasPermission table**

In `internal/auth/rbac_test.go`, add to the existing test table (match testify/t.Run style):
```go
{"staff", "inventory:purchase", false},
{"staff", "inventory:receive", true},
{"shift_manager", "inventory:purchase", true},
{"shift_manager", "inventory:receive", true},
{"gm", "inventory:purchase", true},
{"gm", "inventory:receive", true},
{"owner", "inventory:purchase", true},
{"owner", "inventory:receive", true},
{"read_only", "inventory:purchase", false},
{"read_only", "inventory:receive", false},
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/auth/ -run TestRole_HasPermission -v`
Expected: FAIL

- [ ] **Step 3: Add permissions to rbac.go**

- `staff`: add `"inventory:receive"`
- `shift_manager`: add `"inventory:purchase"`, `"inventory:receive"`
- `gm`: add `"inventory:purchase"`, `"inventory:receive"`
- `owner`: add `"inventory:purchase"`, `"inventory:receive"`

Note: `inventory:approve` already exists on `shift_manager`, `gm`, `owner` — reused for PO approval.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/auth/ -run TestRole_HasPermission -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/auth/rbac.go internal/auth/rbac_test.go
git commit -m "feat: add inventory:purchase and inventory:receive RBAC permissions"
```

---

## Task 3: Purchasing Service — PO CRUD + Auto-Generation + Receiving

**Files:**
- Create: `internal/inventory/purchasing.go`
- Create: `internal/inventory/purchasing_test.go`

- [ ] **Step 1: Write tests for pure functions**

Create `internal/inventory/purchasing_test.go`:

```go
package inventory

import (
	"testing"
)

func TestComputeVarianceFlag(t *testing.T) {
	tests := []struct {
		ordered  float64
		received float64
		expected string
	}{
		{49.0, 48.02, "exact"},   // exactly 2% tolerance (inclusive)
		{49.0, 48.01, "short"},   // just beyond 2%
		{49.0, 49.98, "exact"},   // 2% over (inclusive)
		{49.0, 49.99, "over"},    // just beyond 2% over
		{49.0, 0.0, "not_received"},
		{10.0, 10.0, "exact"},    // perfect match
		{10.0, 9.8, "exact"},     // exactly 2%
		{10.0, 9.79, "short"},    // beyond 2%
		{10.0, 10.2, "exact"},    // 2% over
		{10.0, 10.21, "over"},    // beyond 2% over
		{0.0, 0.0, "exact"},      // zero ordered, zero received
	}
	for _, tt := range tests {
		got := computeVarianceFlag(tt.ordered, tt.received)
		if got != tt.expected {
			t.Errorf("computeVarianceFlag(%.2f, %.2f) = %q, want %q", tt.ordered, tt.received, got, tt.expected)
		}
	}
}

func TestComputeAvgDailyUsage(t *testing.T) {
	// 100 units used over 10 days = 10/day
	got := computeAvgDailyUsage(100.0, 10)
	if got != 10.0 {
		t.Errorf("expected 10.0, got %.2f", got)
	}

	// Fallback: 0 days → uses par_level/7
	got = computeAvgDailyUsage(0, 0)
	if got != 0.0 {
		t.Errorf("expected 0.0, got %.2f", got)
	}
}

func TestEffectiveReorderPoint(t *testing.T) {
	// Manual reorder point is higher
	got := effectiveReorderPoint(25.0, 1, 8.0)
	if got != 25.0 {
		t.Errorf("expected 25.0 (manual wins), got %.2f", got)
	}

	// Dynamic is higher: 3 days * 10/day = 30
	got = effectiveReorderPoint(25.0, 3, 10.0)
	if got != 30.0 {
		t.Errorf("expected 30.0 (dynamic wins), got %.2f", got)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/inventory/ -run "TestComputeVariance\|TestComputeAvg\|TestEffective" -v`
Expected: FAIL

- [ ] **Step 3: Write purchasing.go**

Create `internal/inventory/purchasing.go` with:

**Types:**
```go
type PurchaseOrder struct {
	PurchaseOrderID string     `json:"purchase_order_id"`
	OrgID           string     `json:"org_id"`
	LocationID      string     `json:"location_id"`
	VendorName      string     `json:"vendor_name"`
	Status          string     `json:"status"`
	Source          string     `json:"source"`
	SuggestedAt     *time.Time `json:"suggested_at,omitempty"`
	ApprovedBy      *string    `json:"approved_by,omitempty"`
	ApprovedAt      *time.Time `json:"approved_at,omitempty"`
	ReceivedBy      *string    `json:"received_by,omitempty"`
	ReceivedAt      *time.Time `json:"received_at,omitempty"`
	TotalEstimated  int64      `json:"total_estimated"`
	TotalActual     int64      `json:"total_actual"`
	Notes           string     `json:"notes"`
	LineCount       int        `json:"line_count,omitempty"`
}

type POLine struct {
	POLineID         string   `json:"po_line_id"`
	IngredientID     string   `json:"ingredient_id"`
	IngredientName   string   `json:"ingredient_name,omitempty"`
	OrderedQty       float64  `json:"ordered_qty"`
	OrderedUnit      string   `json:"ordered_unit"`
	EstimatedUnitCost int     `json:"estimated_unit_cost"`
	ReceivedQty      *float64 `json:"received_qty"`
	ReceivedUnitCost *int     `json:"received_unit_cost"`
	VarianceQty      *float64 `json:"variance_qty"`
	VarianceFlag     *string  `json:"variance_flag"`
	Note             string   `json:"note"`
}

type POWithLines struct {
	PurchaseOrder
	Lines []POLine `json:"lines"`
}

type POLineInput struct {
	IngredientID     string  `json:"ingredient_id"`
	OrderedQty       float64 `json:"ordered_qty"`
	OrderedUnit      string  `json:"ordered_unit"`
	EstimatedUnitCost int    `json:"estimated_unit_cost"`
}

type ReceiveLineInput struct {
	POLineID        string  `json:"po_line_id"`
	ReceivedQty     float64 `json:"received_qty"`
	ReceivedUnitCost int    `json:"received_unit_cost"`
	Note            string  `json:"note"`
}

type Discrepancy struct {
	IngredientName string  `json:"ingredient_name"`
	Ordered        float64 `json:"ordered"`
	Received       float64 `json:"received"`
	Flag           string  `json:"flag"`
}

type PARBreach struct {
	IngredientID        string  `json:"ingredient_id"`
	IngredientName      string  `json:"ingredient_name"`
	CurrentLevel        float64 `json:"current_level"`
	ReorderPoint        float64 `json:"reorder_point"`
	PARLevel            float64 `json:"par_level"`
	AvgDailyUsage       float64 `json:"avg_daily_usage"`
	ProjectedStockoutDays float64 `json:"projected_stockout_days"`
	VendorName          string  `json:"vendor_name"`
	HasPendingPO        bool    `json:"has_pending_po"`
}
```

**Pure functions:**
```go
func computeVarianceFlag(ordered, received float64) string {
	if received == 0 && ordered > 0 {
		return "not_received"
	}
	if ordered == 0 {
		return "exact"
	}
	tolerance := ordered * 0.02
	diff := received - ordered
	if diff >= -tolerance && diff <= tolerance {
		return "exact"
	}
	if received < ordered {
		return "short"
	}
	return "over"
}

func computeAvgDailyUsage(totalUsage float64, days int) float64 {
	if days <= 0 {
		return 0
	}
	return totalUsage / float64(days)
}

func effectiveReorderPoint(manualReorderPoint float64, leadTimeDays int, avgDailyUsage float64) float64 {
	dynamic := float64(leadTimeDays) * avgDailyUsage
	if manualReorderPoint > dynamic {
		return manualReorderPoint
	}
	return dynamic
}
```

**Service methods (all on `*Service`, using `s.pool` and `s.bus`):**

- `CreatePO(ctx, orgID, locationID, vendorName, notes string, lines []POLineInput) (*PurchaseOrder, error)` — INSERT PO + lines, compute total_estimated
- `ListPOs(ctx, orgID, locationID, status string) ([]PurchaseOrder, error)` — query with optional filters, include line_count via subquery
- `GetPO(ctx, orgID, poID string) (*POWithLines, error)` — PO + lines with ingredient names via JOIN
- `UpdatePOStatus(ctx, orgID, poID, status, approvedBy string) error` — approve/cancel with validation
- `UpdatePODraft(ctx, orgID, poID, notes string, lines []POLineInput) error` — edit draft only
- `DeletePO(ctx, orgID, poID string) error` — delete draft only
- `ListPendingPOs(ctx, orgID, locationID string) ([]PurchaseOrder, error)` — approved POs with days_since_approved computed
- `ReceivePO(ctx, orgID, poID, receivedBy string, lines []ReceiveLineInput) ([]Discrepancy, int64, error)` — line-by-line receive, compute flags, set omitted lines to not_received, update avg_daily_usage, emit discrepancy alerts, return discrepancies + total_actual
- `GenerateSuggestedPOs(ctx, orgID, locationID, countID string) error` — auto-generation engine: query breaching ingredients, group by vendor, create/merge draft POs, emit alerts
- `UpdateAvgDailyUsage(ctx, orgID, locationID string, countID string) error` — compute and write avg_daily_usage
- `GetPARBreaches(ctx, orgID, locationID string) ([]PARBreach, error)` — query ingredients below effective reorder point with projected stockout

- [ ] **Step 4: Run tests**

Run: `go test ./internal/inventory/ -run "TestComputeVariance\|TestComputeAvg\|TestEffective" -v`
Expected: PASS

- [ ] **Step 5: Verify build**

Run: `go build ./cmd/fireline/`
Expected: No errors

- [ ] **Step 6: Commit**

```bash
git add internal/inventory/purchasing.go internal/inventory/purchasing_test.go
git commit -m "feat: add purchasing service — PO CRUD, auto-generation engine, receiving, PAR breaches"
```

---

## Task 4: HTTP Handlers — PO Endpoints

**Files:**
- Create: `internal/api/po_handler.go`
- Modify: `internal/api/handlers.go`
- Modify: `internal/api/counting_handler.go`

- [ ] **Step 1: Write po_handler.go**

New file with handler methods on `*InventoryHandler`:

| Method | Endpoint | Body/Query |
|--------|----------|------------|
| `CreatePO` | `POST /api/v1/inventory/po` | JSON: location_id, vendor_name, notes, lines |
| `ListPOs` | `GET /api/v1/inventory/po` | Query: location_id, status |
| `GetPO` | `GET /api/v1/inventory/po/{id}` | — |
| `UpdatePO` | `PUT /api/v1/inventory/po/{id}` | JSON: status or notes+lines for draft edit |
| `DeletePO` | `DELETE /api/v1/inventory/po/{id}` | — |
| `ListPendingPOs` | `GET /api/v1/inventory/po/pending` | Query: location_id |
| `ReceivePO` | `POST /api/v1/inventory/po/{id}/receive` | JSON: lines[] (received_by from auth context) |
| `GetPARBreaches` | `GET /api/v1/inventory/par-breaches` | Query: location_id |

Follow existing `counting_handler.go` pattern:
- Extract orgID via `tenant.OrgIDFrom(r.Context())`
- Extract userID via `auth.UserIDFrom(r.Context())` for approve/receive
- Use `r.PathValue("id")` for path params
- Use `api.WriteJSON` and `api.WriteError`

For `ReceivePO`: extract `received_by` from `auth.UserIDFrom(r.Context())`, NOT from request body. The PINLogin method resolves employees to their linked `user_id`, so this works for both JWT and PIN auth paths. The `received_by` column references `users(user_id)`.

**RBAC enforcement:** The existing auth middleware validates JWT but does NOT check granular permissions. For this sprint, RBAC permissions (`inventory:purchase`, `inventory:approve`, `inventory:receive`) are defined in `rbac.go` for future middleware enforcement. The handlers do not add explicit permission checks — consistent with the existing handler pattern (no handler in the codebase currently checks permissions). Permission enforcement middleware will be added in a hardening sprint.

- [ ] **Step 2: Register routes in handlers.go**

Add to `InventoryHandler.RegisterRoutes`, AFTER existing routes:
```go
// PO routes — must register specific paths BEFORE parameterized ones
mux.Handle("GET /api/v1/inventory/po/pending", authMW(http.HandlerFunc(h.ListPendingPOs)))
mux.Handle("GET /api/v1/inventory/par-breaches", authMW(http.HandlerFunc(h.GetPARBreaches)))
mux.Handle("POST /api/v1/inventory/po", authMW(http.HandlerFunc(h.CreatePO)))
mux.Handle("GET /api/v1/inventory/po", authMW(http.HandlerFunc(h.ListPOs)))
mux.Handle("GET /api/v1/inventory/po/{id}", authMW(http.HandlerFunc(h.GetPO)))
mux.Handle("PUT /api/v1/inventory/po/{id}", authMW(http.HandlerFunc(h.UpdatePO)))
mux.Handle("DELETE /api/v1/inventory/po/{id}", authMW(http.HandlerFunc(h.DeletePO)))
mux.Handle("POST /api/v1/inventory/po/{id}/receive", authMW(http.HandlerFunc(h.ReceivePO)))
```

**Important:** Register `/po/pending` and `/par-breaches` BEFORE `/po/{id}` to prevent the wildcard from matching first.

- [ ] **Step 3: Wire auto-generation into count submission**

In `internal/api/counting_handler.go`, in the `UpdateCountStatus` handler, after the variance calculation call, add a call to `GenerateSuggestedPOs`:

```go
// After variance calculation (best-effort), trigger PO auto-generation
locID := cw.LocationID  // capture before goroutine
go func() {
    if err := h.svc.GenerateSuggestedPOs(context.Background(), orgID, locID, countID); err != nil {
        slog.Error("auto-generate POs failed", "error", err, "count_id", countID)
    }
}()
```

This runs async so it doesn't block the submission response.

- [ ] **Step 4: Build and verify**

Run: `go build ./cmd/fireline/`
Expected: No errors

- [ ] **Step 5: Commit**

```bash
git add internal/api/po_handler.go internal/api/handlers.go internal/api/counting_handler.go
git commit -m "feat: add PO HTTP handlers, wire auto-generation into count submission"
```

---

## Task 5: Web Dashboard — Purchase Orders Page

**Files:**
- Create: `web/src/pages/PurchaseOrdersPage.tsx`
- Create: `web/src/hooks/usePurchaseOrders.ts`
- Modify: `web/src/lib/api.ts`
- Modify: `web/src/components/Layout.tsx`
- Modify: `web/src/App.tsx`
- Modify: `web/src/pages/InventoryPage.tsx`

- [ ] **Step 1: Add types and API methods to api.ts**

Add to `web/src/lib/api.ts`:

Types: `PurchaseOrder`, `POLine`, `POWithLines`, `PARBreach`

API methods:
```typescript
export const poApi = {
  list: (locationId: string, status?: string) =>
    request<{purchase_orders: PurchaseOrder[]}>(`/inventory/po?location_id=${locationId}${status ? `&status=${status}` : ''}`),
  get: (id: string) => request<POWithLines>(`/inventory/po/${id}`),
  create: (data: any) => request<PurchaseOrder>('/inventory/po', { method: 'POST', body: JSON.stringify(data) }),
  update: (id: string, data: any) => request<any>(`/inventory/po/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  delete: (id: string) => request<any>(`/inventory/po/${id}`, { method: 'DELETE' }),
  pending: (locationId: string) => request<{pending: PurchaseOrder[]}>(`/inventory/po/pending?location_id=${locationId}`),
  parBreaches: (locationId: string) => request<{breaches: PARBreach[]}>(`/inventory/par-breaches?location_id=${locationId}`),
};
```

- [ ] **Step 2: Create hooks**

Create `web/src/hooks/usePurchaseOrders.ts` with:
- `usePOs(locationId, status?)` — list POs
- `usePO(poId)` — get single PO with lines
- `usePendingPOs(locationId)` — pending POs
- `usePARBreaches(locationId)` — PAR breaches
- `useApprovePO()` — mutation
- `useCancelPO()` — mutation

- [ ] **Step 3: Add nav item and route**

In `Layout.tsx`:
1. Add `ShoppingCart` to the lucide-react import statement at the top of the file
2. Add to `navItems` array after the Inventory entry:
```typescript
{ to: '/purchase-orders', label: 'Purchase Orders', icon: ShoppingCart },
```

In `App.tsx`:
```typescript
import PurchaseOrdersPage from './pages/PurchaseOrdersPage';
// Add route:
<Route path="purchase-orders" element={<PurchaseOrdersPage />} />
```

- [ ] **Step 4: Create PurchaseOrdersPage.tsx**

Three sections:
1. **Suggested POs**: cards with vendor name, item count, est. total, "Approve" (green) and "Review" buttons. Only shown if draft system_recommended POs exist.
2. **Active POs**: DataTable of approved POs — vendor, items, est. total, approved date, days waiting. Click → detail modal.
3. **PO History**: DataTable of received/cancelled — vendor, ordered $, actual $, variance $ (color-coded), status badge, date.

Detail modal: shows PO header + lines table. For received POs, shows received columns + variance flags.

- [ ] **Step 5: Add PAR breach banner to InventoryPage**

In `InventoryPage.tsx`, import `usePARBreaches`. In the PAR Status tab section, add a warning banner above the table when breaches exist:
```tsx
{breaches.length > 0 && (
  <div className="bg-red-50 border border-red-200 rounded-lg p-3 mb-4 flex items-center justify-between">
    <span className="text-red-800 text-sm font-medium">
      {breaches.length} ingredient{breaches.length > 1 ? 's' : ''} below reorder point
    </span>
    <a href="/purchase-orders" className="text-red-600 text-sm underline">View Purchase Orders →</a>
  </div>
)}
```

- [ ] **Step 6: Build and verify**

```bash
cd /Users/bistrocloud/Documents/AI_Restaurant_System/fireline/web
npx vite build
```
Expected: Build succeeds

- [ ] **Step 7: Commit**

```bash
cd /Users/bistrocloud/Documents/AI_Restaurant_System/fireline
git add web/src/lib/api.ts web/src/hooks/usePurchaseOrders.ts web/src/pages/PurchaseOrdersPage.tsx web/src/components/Layout.tsx web/src/App.tsx web/src/pages/InventoryPage.tsx
git commit -m "feat: add Purchase Orders page with suggested POs, one-tap approve, history, and PAR breach banner"
```

---

## Task 6: Tablet — Delivery Receiving Screen

**Files:**
- Create: `tablet/stores/receive.ts`
- Create: `tablet/components/ReceiveRow.tsx`
- Create: `tablet/app/(tabs)/receive.tsx`
- Modify: `tablet/app/(tabs)/_layout.tsx`
- Delete: `tablet/app/(tabs)/tasks.tsx`

- [ ] **Step 1: Create receive store**

Create `tablet/stores/receive.ts` (Zustand, following `count.ts` pattern):
- `pendingPOs: PurchaseOrder[]`
- `activePO: POWithLines | null`
- `receivedLines: Map<poLineId, {received_qty, received_unit_cost, note}>` — local edits
- `loadPending()` — GET `/inventory/po/pending?location_id=X`
- `startReceiving(poId)` — GET `/inventory/po/{id}`, pre-fill receivedLines with ordered values
- `updateLine(poLineId, receivedQty, receivedUnitCost, note)` — update local + save to AsyncStorage
- `markNotReceived(poLineId)` — set qty to 0
- `submitReceiving()` — POST `/inventory/po/{id}/receive` with receivedLines
- `progress` computed: count of lines where user has made any modification

- [ ] **Step 2: Create ReceiveRow component**

Props: `ingredientName, orderedQty, orderedUnit, estimatedCost, receivedQty, receivedCost, onChangeQty, onChangeCost, onNotReceived, onChangeNote, note`

Layout:
- Ingredient name + "Ordered: 50 lb" reference text
- Large numeric input for received qty (pre-filled with ordered)
- Cost input (pre-filled with estimated, smaller)
- Live variance indicator:
  - Green check (within 2%)
  - Amber warning (short/over)
  - Red X (not received)
- "Not Received" button
- Expandable note field
- 48px+ touch targets

- [ ] **Step 3: Create receive.tsx screen**

Two states:
1. **Pending list**: FlatList of pending POs. Each card: vendor, items, est. total, days waiting. Tap → start receiving.
2. **Receiving**: Header with vendor name. ProgressBar ("8 of 12 verified"). FlatList of ReceiveRow components. "Review & Submit" button at bottom. Review shows discrepancy summary before final submit.

- [ ] **Step 4: Update tab layout**

In `tablet/app/(tabs)/_layout.tsx`:
- Replace the Tasks tab entry with Receive:
```tsx
<Tabs.Screen
  name="receive"
  options={{
    title: 'Receive',
    tabBarIcon: () => <TabIcon label="📦" />,
  }}
/>
```
- Remove the `tasks` tab entry
- Delete `tablet/app/(tabs)/tasks.tsx`

- [ ] **Step 5: Commit**

```bash
cd /Users/bistrocloud/Documents/AI_Restaurant_System/fireline
git add tablet/stores/receive.ts tablet/components/ReceiveRow.tsx tablet/app/\(tabs\)/receive.tsx tablet/app/\(tabs\)/_layout.tsx
git rm tablet/app/\(tabs\)/tasks.tsx
git commit -m "feat: add tablet delivery receiving screen with line-by-line verification and variance indicators"
```

---

## Task 7: End-to-End Test

- [ ] **Step 1: Run all Go tests**

```bash
cd /Users/bistrocloud/Documents/AI_Restaurant_System/fireline
go test ./... -count=1 -p 1
```
Expected: All packages pass

- [ ] **Step 2: Rebuild and restart server**

```bash
pkill -f "./fireline" || true
go build -o fireline ./cmd/fireline/
DATABASE_URL="postgres://fireline_app:fireline_app@localhost:5432/fireline?sslmode=disable" \
ADMIN_DATABASE_URL="postgres://fireline:fireline@localhost:5432/fireline?sslmode=disable" \
ENV=development PORT=8080 ./fireline &
sleep 3
```

- [ ] **Step 3: Test PO lifecycle**

```bash
TOKEN=$(curl -s http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"owner@bistrocloud.com","password":"DemoPassword1234!"}' \
  | python3 -c "import sys,json; print(json.load(sys.stdin)['access_token'])")

# Create manual PO
curl -s -X POST http://localhost:8080/api/v1/inventory/po \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"location_id":"a1111111-1111-1111-1111-111111111111","vendor_name":"US Foods","notes":"Weekly order","lines":[{"ingredient_id":"11111111-aaaa-1111-aaaa-111111111111","ordered_qty":50.0,"ordered_unit":"lb","estimated_unit_cost":440}]}'

# List POs
curl -s "http://localhost:8080/api/v1/inventory/po?location_id=a1111111-1111-1111-1111-111111111111" \
  -H "Authorization: Bearer $TOKEN"

# Approve PO
PO_ID=$(curl -s "http://localhost:8080/api/v1/inventory/po?location_id=a1111111-1111-1111-1111-111111111111" \
  -H "Authorization: Bearer $TOKEN" | python3 -c "import sys,json; print(json.load(sys.stdin)['purchase_orders'][0]['purchase_order_id'])")

curl -s -X PUT "http://localhost:8080/api/v1/inventory/po/$PO_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"status":"approved"}'

# Check pending
curl -s "http://localhost:8080/api/v1/inventory/po/pending?location_id=a1111111-1111-1111-1111-111111111111" \
  -H "Authorization: Bearer $TOKEN"

# Check PAR breaches
curl -s "http://localhost:8080/api/v1/inventory/par-breaches?location_id=a1111111-1111-1111-1111-111111111111" \
  -H "Authorization: Bearer $TOKEN"
```

- [ ] **Step 4: Build web frontend**

```bash
cd /Users/bistrocloud/Documents/AI_Restaurant_System/fireline/web
npx vite build
```
Expected: Build succeeds

- [ ] **Step 5: Final commit**

```bash
cd /Users/bistrocloud/Documents/AI_Restaurant_System/fireline
git add -A
git commit -m "feat: SP9 complete — purchase orders, auto-generation, delivery receiving, PAR breaches"
```
