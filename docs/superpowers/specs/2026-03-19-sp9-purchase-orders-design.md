# SP9: Purchase Orders, PAR Breach Detection & Delivery Receiving Design Spec

**Date:** 2026-03-19
**Status:** Approved
**Scope:** PO tables, auto-generation engine, PO management web page, tablet delivery receiving, PAR breach alerting
**Maps to:** Build Plan Sprint 17 (Inventory — PAR Levels, Reorder Points & Purchase Orders)

---

## 1. Database — New Migration (007)

New migration file: `migrations/007_purchase_orders.sql`

### New Tables

```sql
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
    received_by        UUID REFERENCES users(user_id),  -- PIN auth resolves to linked user_id
    received_at        TIMESTAMPTZ,
    total_estimated    BIGINT NOT NULL DEFAULT 0,  -- cents, BIGINT for multi-line sums
    total_actual       BIGINT NOT NULL DEFAULT 0,  -- cents, BIGINT for multi-line sums
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
    variance_qty        NUMERIC(12,4),  -- computed by application: received_qty - ordered_qty
    variance_flag       TEXT CHECK (variance_flag IN ('exact', 'short', 'over', 'not_received')),  -- set by application during receive
    received_at         TIMESTAMPTZ,
    note                TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

### Schema Alterations

```sql
ALTER TABLE ingredient_location_configs
    ADD COLUMN lead_time_days INT NOT NULL DEFAULT 1,
    ADD COLUMN avg_daily_usage NUMERIC(12,4) NOT NULL DEFAULT 0;
```

### RLS Policies

Both tables follow established pattern:
```sql
ALTER TABLE <table> ENABLE ROW LEVEL SECURITY;
ALTER TABLE <table> FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON <table>
    USING (org_id = current_setting('app.current_org_id')::UUID);
GRANT SELECT, INSERT, UPDATE, DELETE ON <table> TO fireline_app;
```

### Indexes

```sql
CREATE INDEX idx_po_location ON purchase_orders(org_id, location_id, created_at DESC);
CREATE INDEX idx_po_status ON purchase_orders(org_id, status);
CREATE INDEX idx_po_vendor ON purchase_orders(org_id, vendor_name);
CREATE INDEX idx_po_lines_po ON purchase_order_lines(purchase_order_id);
CREATE INDEX idx_po_lines_ingredient ON purchase_order_lines(org_id, ingredient_id);
```

---

## 2. PO Auto-Generation Engine

New file: `internal/inventory/purchasing.go`

### Trigger

Runs after a count is submitted (called from the count submission handler after variance calculation).

### Logic

1. Query all ingredients at the location where latest counted quantity falls below the effective reorder point
   - Effective reorder point = `MAX(ilc.reorder_point, lead_time_days * avg_daily_usage)` — uses the higher of the manually-set reorder point and the dynamically computed one. This means manual overrides are always respected while the system can raise the bar when usage data suggests it.
   - Counted quantity = `counted_qty` from the just-submitted count's lines
2. Group breaching ingredients by `vendor_name` (from `ingredient_location_configs`)
3. For each vendor group:
   - Check if an existing `draft` PO already exists for this vendor + location — if yes, merge new items; if no, create new draft
   - Per ingredient: suggested order qty = `par_level - counted_qty` (bring back to PAR)
   - Estimated line cost = `suggested_qty * local_cost_per_unit`
   - Sum all line estimates into `total_estimated`
4. Mark PO `source = 'system_recommended'`, set `suggested_at = now()`
5. Emit `inventory.po.suggested` event on event bus

### avg_daily_usage Calculation

Updated when a count is submitted:
- `theoretical_usage_since_last_count / days_between_counts`
- Fallback if no prior count: `par_level / 7`
- Written to `ingredient_location_configs.avg_daily_usage`

### Alert Integration

- `inventory.po.suggested` event → alert enqueued
- Severity: `critical` if any ingredient is below 20% of PAR level, `warning` otherwise
- Alert title: "Suggested PO for [Vendor] — [N] items, ~$[total]"
- Alert recommended action: "Approve PO" with PO ID

---

## 3. API Endpoints

All endpoints require auth middleware (JWT). Tenant context via `TenantTx`.

### PO Management (web dashboard)

**POST /api/v1/inventory/po** — Create manual PO
```json
Request: {
  "location_id": "uuid",
  "vendor_name": "US Foods",
  "notes": "Weekly order",
  "lines": [
    { "ingredient_id": "uuid", "ordered_qty": 50.0, "ordered_unit": "lb", "estimated_unit_cost": 440 }
  ]
}
Response: { "purchase_order_id": "uuid", "status": "draft", "total_estimated": 22000 }
```

**GET /api/v1/inventory/po?location_id=X&status=Y** — List POs
```json
Response: {
  "purchase_orders": [{
    "purchase_order_id": "uuid", "vendor_name": "US Foods", "status": "draft",
    "source": "system_recommended", "line_count": 5, "total_estimated": 22000,
    "total_actual": 0, "suggested_at": "...", "approved_at": null, "received_at": null
  }]
}
```

**GET /api/v1/inventory/po/{id}** — Get PO with lines
```json
Response: {
  "purchase_order_id": "uuid", "vendor_name": "...", "status": "...",
  "lines": [{
    "po_line_id": "uuid", "ingredient_id": "uuid", "ingredient_name": "Ground Beef",
    "ordered_qty": 50.0, "ordered_unit": "lb", "estimated_unit_cost": 440,
    "received_qty": null, "received_unit_cost": null,
    "variance_qty": null, "variance_flag": null, "note": ""
  }],
  "total_estimated": 22000, "total_actual": 0
}
```

**PUT /api/v1/inventory/po/{id}** — Update PO
```json
Request: { "status": "approved" }  // or "cancelled"
// For draft edits:
Request: { "notes": "updated", "lines": [...] }
```
On approve: set `approved_by` from JWT, `approved_at = now()`.
On cancel: only allowed for draft or approved POs.

**DELETE /api/v1/inventory/po/{id}** — Delete draft PO only
Returns 400 if PO is not in draft status.

### Tablet Receiving

**GET /api/v1/inventory/po/pending?location_id=X** — List approved POs awaiting delivery
```json
Response: {
  "pending": [{
    "purchase_order_id": "uuid", "vendor_name": "US Foods",
    "line_count": 5, "total_estimated": 22000,
    "approved_at": "...", "days_since_approved": 2
  }]
}
```

**POST /api/v1/inventory/po/{id}/receive** — Submit received quantities
```json
Request: {
  "lines": [
    { "po_line_id": "uuid", "received_qty": 48.0, "received_unit_cost": 440, "note": "one case damaged" }
  ]
}
Response: {
  "status": "received",
  "total_actual": 21120,
  "discrepancies": [
    { "ingredient_name": "Ground Beef", "ordered": 50.0, "received": 48.0, "flag": "short" }
  ]
}
```

`received_by` is extracted from the authenticated session (employee_id from PIN auth context or user_id from JWT), NOT from the request body. This prevents impersonation.

**Lines omitted from request:** Any PO line NOT included in the receive request body is automatically set to `received_qty = 0` and `variance_flag = 'not_received'`. This ensures complete receiving.

On receive:
1. Update each PO line: `received_qty`, `received_unit_cost`, `received_at`, compute `variance_qty` and `variance_flag`
   - `variance_qty` = `received_qty - ordered_qty` (set by application)
   - `variance_flag` computed by application:
     - `exact`: received within <=2% of ordered (inclusive)
     - `short`: received < ordered beyond 2% tolerance
     - `over`: received > ordered beyond 2% tolerance
     - `not_received`: received_qty = 0
2. Set omitted lines to `received_qty = 0`, `variance_flag = 'not_received'`
3. Compute `total_actual` = sum of `received_qty * received_unit_cost` across all lines
4. Set PO `status = 'received'`, `received_by` (from auth context), `received_at`
5. Update `ingredient_location_configs.avg_daily_usage` for each received ingredient
6. If any line is `short` or `not_received`: emit `inventory.delivery.discrepancy` event → alert with vendor name and discrepancy details

### PAR Breaches

**GET /api/v1/inventory/par-breaches?location_id=X** — Ingredients below reorder point
```json
Response: {
  "breaches": [{
    "ingredient_id": "uuid", "ingredient_name": "Ground Beef",
    "current_level": 12.5, "reorder_point": 25.0, "par_level": 50.0,
    "avg_daily_usage": 8.3, "projected_stockout_days": 1.5,
    "vendor_name": "US Foods", "has_pending_po": true
  }]
}
```

`projected_stockout_days` = `current_level / avg_daily_usage` (0 if avg_daily_usage is 0).

---

## 4. Web Dashboard — Purchase Orders Page

New page: `web/src/pages/PurchaseOrdersPage.tsx`
New hooks: `web/src/hooks/usePurchaseOrders.ts`
New sidebar entry in Layout.tsx at route `/purchase-orders`, positioned after Inventory in nav order
New Zustand store for tablet: `tablet/stores/receive.ts` (following `count.ts` pattern — offline persistence, pending state, auto-sync)

### Page Layout

**Section 1: Suggested POs** (only shown if any exist)
- Card per system-recommended draft PO
- Each card: vendor name, item count, estimated total, suggested date
- Two action buttons: "Approve" (green, one-tap) and "Review" (opens detail)
- Approve calls `PUT /po/{id}` with `status: approved`

**Section 2: Active POs**
- DataTable of approved POs (awaiting delivery)
- Columns: vendor, items, est. total, approved by, approved date, days waiting
- Click row → detail modal

**Section 3: PO History**
- DataTable of received + cancelled POs
- Columns: vendor, ordered $, actual $, variance $ (color-coded: green if savings, red if over), status badge, date
- Click row → detail modal with line-by-line receiving data

**PO Detail Modal:**
- Header: vendor, status badge, source badge (manual/system), dates
- Lines table: ingredient, ordered qty, unit, est. cost
- For received POs: additional columns — received qty, actual cost, variance qty, variance flag badge
- For drafts: inline editable quantities, add line button, remove line button, save button

### Inventory Page Extension

Add to PAR Status tab in `InventoryPage.tsx`:
- Red warning banner when PAR breaches exist: "X ingredients below reorder point"
- Link to Purchase Orders page

---

## 5. Tablet — Delivery Receiving Screen

**Replace the "Tasks" placeholder tab** with "Receive" tab (truck icon).

### Pending Deliveries List (`tablet/app/(tabs)/receive.tsx`)
- FlatList of approved POs for this location
- Each card: vendor name, item count, estimated total, days since approved
- Tap card → receiving screen

### Receiving Screen (line-by-line)
- Header: vendor name, PO info
- ScrollView/FlatList of lines, each showing:
  - Ingredient name + ordered qty + unit (read-only reference)
  - Large numeric input: received qty (pre-filled with ordered qty)
  - Unit cost input: received cost (pre-filled with estimated)
  - Live variance indicator:
    - Green checkmark: within 2% of ordered
    - Amber warning: short or over
    - Red X: not received (qty = 0)
  - "Not Received" button per line (sets qty to 0)
  - Note field (expandable)
- ProgressBar: "8 of 12 items verified"
- Any modification marks the line as "verified"

### Review & Submit
- Summary section: total items, discrepancy count
- Discrepancy list: ingredient name, ordered vs received, flag
- "Complete Receiving" button
- On submit: POST `/po/{id}/receive`
- Success: show summary, navigate back to pending list

### UX
- 48px+ touch targets, numeric keyboard
- Offline support: save receiving progress to AsyncStorage, sync on reconnect
- Progress persists if app is backgrounded

---

## 6. RBAC Permissions

Add to `internal/auth/rbac.go` (following existing `domain:action` convention):
- `inventory:purchase` — create and manage POs (roles: `shift_manager`, `gm`, `owner`)
- `inventory:approve` — already exists for count approval; reuse for PO approval (roles: `gm`, `owner`)
- `inventory:receive` — receive deliveries (roles: `staff`, `shift_manager`, `gm`, `owner`)

---

## 7. Dependencies

### Go Backend
- No new Go dependencies

### Tablet App
- No new tablet dependencies (uses existing Expo + Zustand + AsyncStorage)

### Web Dashboard
- No new web dependencies (uses existing React Query + Tailwind + DataTable)

---

## 8. Testing Strategy

### Backend Tests
- PO creation: manual and system-recommended
- PO lifecycle: draft → approved → received
- PO lifecycle: draft → cancelled
- Auto-generation: mock count with ingredients below reorder → draft PO created per vendor
- Receiving: line-by-line variance flag computation (exact/short/over/not_received)
- 2% tolerance boundary: 49.0 ordered, 48.02 received = exact; 48.01 = short
- avg_daily_usage update after receiving
- PAR breach query returns correct projected stockout days
- Alert emitted on PO suggestion and delivery discrepancy
- RBAC: staff cannot create POs, shift_manager cannot approve
- RLS: PO from org A not visible to org B

### Tablet Tests
- Receiving flow: load pending POs → tap → enter quantities → submit
- Pre-fill: received qty defaults to ordered qty
- Variance indicators update live
- Offline: enter receiving data offline → sync on reconnect

### Web Tests
- PO list filters by status
- One-tap approve changes status
- Detail modal shows correct data
- History shows variance calculations
