# SP8: Inventory Counting, Waste Logging & Variance Analysis Design Spec

**Date:** 2026-03-19
**Status:** Approved
**Scope:** React Native (Expo) tablet app scaffold, inventory counting UI, waste logging, variance categorization engine, manager variance dashboard
**Maps to:** Build Plan Sprints 12 (tablet shell) + 16 (inventory actual usage, variance & counting)

---

## 1. Tablet App — Expo Project

New Expo project at `fireline/tablet/` using Expo SDK 52 + Expo Router (file-based routing).

### Directory Structure

```
tablet/
  app/
    _layout.tsx                  — root layout (auth gate)
    (auth)/
      login.tsx                  — manager email/password login
      pin.tsx                    — staff PIN entry keypad
    (tabs)/
      _layout.tsx                — tab navigator (5 tabs)
      count.tsx                  — inventory counting (this sprint)
      waste.tsx                  — waste logging (this sprint)
      tasks.tsx                  — placeholder
      kds.tsx                    — placeholder
      clock.tsx                  — placeholder
  components/
    PinPad.tsx                   — numeric PIN entry keypad (6-digit)
    CountRow.tsx                 — ingredient count input row
    WasteForm.tsx                — waste entry form with reason picker
    CategoryGroup.tsx            — collapsible ingredient category group
    ProgressBar.tsx              — count progress indicator
  lib/
    api.ts                       — fetch client with JWT, base URL config
    auth.ts                      — secure storage for manager JWT, staff context
    offline.ts                   — AsyncStorage queue for offline sync
  stores/
    auth.ts                      — zustand: manager JWT, active staff, location
    count.ts                     — zustand: current count session, lines
    waste.ts                     — zustand: waste form state
```

### Authentication Flow

1. **Manager Login** — email/password via `POST /api/v1/auth/login`, JWT stored in expo-secure-store
2. **Location Binding** — manager selects location after login, stored in auth store
3. **Staff PIN Entry** — 6-digit PIN verified via `POST /api/v1/auth/pin-verify`, returns staff display name + user ID
4. **Session Model** — 2-minute inactivity timeout returns to PIN screen; manager JWT persists until explicit logout
5. **PIN Lockout** — 5 failed attempts locks PIN entry for 5 minutes (client-enforced + server-enforced)

### Tab Navigation Shell

| Tab | Icon | Sprint | Status |
|-----|------|--------|--------|
| Count | clipboard-list | SP8 | Active |
| Waste | trash | SP8 | Active |
| Tasks | list-checks | Future | Placeholder |
| KDS | monitor | Future | Placeholder |
| Clock | clock | Future | Placeholder |

Placeholder tabs show centered text: "Coming Soon" with the feature name.

### UX Constraints

- Touch targets minimum 48px for gloved hands
- High-contrast theme support (dark mode for back-of-house)
- Large font sizes (16px minimum body, 20px+ for count inputs)
- Category-based navigation for counting (walk-in cooler, dry storage, etc.)

---

## 2. Backend — New Migration (006)

New migration file: `migrations/006_inventory_counting.sql`

### Tables

```sql
CREATE TABLE inventory_counts (
    count_id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID NOT NULL REFERENCES organizations(org_id),
    location_id    UUID NOT NULL REFERENCES locations(location_id),
    counted_by     UUID NOT NULL REFERENCES employees(employee_id),
    count_type     TEXT NOT NULL CHECK (count_type IN ('full', 'spot_check')),
    status         TEXT NOT NULL DEFAULT 'in_progress' CHECK (status IN ('in_progress', 'submitted', 'approved')),
    started_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    submitted_at   TIMESTAMPTZ,
    approved_by    UUID REFERENCES users(user_id),  -- manager approves via user account
    approved_at    TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE inventory_count_lines (
    count_line_id    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id           UUID NOT NULL REFERENCES organizations(org_id),
    count_id         UUID NOT NULL REFERENCES inventory_counts(count_id),
    location_id      UUID NOT NULL REFERENCES locations(location_id),
    ingredient_id    UUID NOT NULL REFERENCES ingredients(ingredient_id),
    expected_qty     NUMERIC(12,4),
    counted_qty      NUMERIC(12,4),
    unit             TEXT NOT NULL,
    note             TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE waste_logs (
    waste_id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID NOT NULL REFERENCES organizations(org_id),
    location_id    UUID NOT NULL REFERENCES locations(location_id),
    ingredient_id  UUID NOT NULL REFERENCES ingredients(ingredient_id),
    quantity       NUMERIC(12,4) NOT NULL,
    unit           TEXT NOT NULL,
    reason         TEXT NOT NULL CHECK (reason IN ('expired', 'dropped', 'overcooked', 'contaminated', 'overproduction', 'other')),
    logged_by      UUID NOT NULL REFERENCES employees(employee_id),
    logged_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    note           TEXT
);

CREATE TABLE inventory_variances (
    variance_id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id              UUID NOT NULL REFERENCES organizations(org_id),
    location_id         UUID NOT NULL REFERENCES locations(location_id),
    ingredient_id       UUID NOT NULL REFERENCES ingredients(ingredient_id),
    count_id            UUID NOT NULL REFERENCES inventory_counts(count_id),
    period_start        TIMESTAMPTZ NOT NULL,
    period_end          TIMESTAMPTZ NOT NULL,
    theoretical_usage   NUMERIC(12,4) NOT NULL,
    actual_usage        NUMERIC(12,4) NOT NULL,
    variance_qty        NUMERIC(12,4) NOT NULL,
    variance_cents      INT NOT NULL,      -- monetary impact in cents (consistent with all other money fields)
    cause_probabilities JSONB NOT NULL DEFAULT '{}',
    severity            TEXT NOT NULL DEFAULT 'info' CHECK (severity IN ('info', 'warning', 'critical')),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

### RLS Policies

All four tables get standard RLS policies following the established pattern:
```sql
ALTER TABLE <table> ENABLE ROW LEVEL SECURITY;
ALTER TABLE <table> FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON <table>
    USING (org_id = current_setting('app.current_org_id')::UUID);
GRANT SELECT, INSERT, UPDATE, DELETE ON <table> TO fireline_app;
```

### Indexes

```sql
CREATE INDEX idx_inventory_counts_location ON inventory_counts(org_id, location_id, started_at DESC);
CREATE INDEX idx_count_lines_count ON inventory_count_lines(count_id);
CREATE INDEX idx_waste_logs_location ON waste_logs(org_id, location_id, logged_at DESC);
CREATE INDEX idx_variances_location ON inventory_variances(org_id, location_id, created_at DESC);
```

---

## 3. Backend — New API Endpoints

All endpoints require auth middleware (JWT). Tenant context set via `TenantTx`.

### Count Endpoints

**POST /api/v1/inventory/counts** — Start a count session
```json
Request:  { "location_id": "uuid", "count_type": "full" | "spot_check" }
Response: { "count_id": "uuid", "status": "in_progress", "started_at": "..." }
```

**GET /api/v1/inventory/counts/:id** — Get count with all lines
```json
Response: {
  "count_id": "uuid", "status": "...", "count_type": "...",
  "counted_by": "uuid", "started_at": "...",
  "lines": [
    { "ingredient_id": "uuid", "name": "Ground Beef", "category": "protein",
      "expected_qty": 42.5, "counted_qty": 38.0, "unit": "lb", "note": "" }
  ],
  "progress": { "counted": 32, "total": 48 }
}
```

**PUT /api/v1/inventory/counts/:id** — Update status (submit or approve)
```json
Request:  { "status": "submitted" }  // or "approved" (manager only)
Response: { "count_id": "uuid", "status": "submitted", "submitted_at": "..." }
```
On submit: triggers variance calculation.

**POST /api/v1/inventory/counts/:id/lines** — Batch upsert count lines
```json
Request: {
  "lines": [
    { "ingredient_id": "uuid", "counted_qty": 38.0, "unit": "lb", "note": "partial case" }
  ]
}
Response: { "updated": 5 }
```

### Waste Endpoints

**POST /api/v1/inventory/waste** — Log waste entry
```json
Request: {
  "location_id": "uuid", "ingredient_id": "uuid",
  "quantity": 2.5, "unit": "lb", "reason": "expired", "note": "found in back of walk-in"
}
Response: { "waste_id": "uuid", "logged_at": "..." }
```

**GET /api/v1/inventory/waste?location_id=X&from=DATE&to=DATE** — List waste logs

### Variance Endpoint

**GET /api/v1/inventory/variances?location_id=X&from=DATE&to=DATE** — List variances
```json
Response: {
  "variances": [
    {
      "variance_id": "uuid", "ingredient_name": "Ground Beef",
      "category": "protein", "theoretical_usage": 45.0, "actual_usage": 52.3,
      "variance_qty": 7.3, "variance_cents": 3285,
      "cause_probabilities": { "portioning": 0.45, "unrecorded_waste": 0.30, "recipe_error": 0.15, "other": 0.10 },
      "severity": "warning"
    }
  ]
}
```

### PIN Verify Endpoint

**POST /api/v1/auth/pin-verify** — Verify staff PIN within manager session
```json
Request:  { "pin": "123456" }
Response: { "user_id": "uuid", "display_name": "Maria S.", "role": "line_cook" }
```
Requires valid manager JWT. PIN looked up within the same org. 5-attempt lockout.

---

## 4. Backend — Variance Categorization Engine

Rule-based Tier 0 engine in `internal/inventory/variance.go`.

Note: The existing `Variance` struct in `inventory.go` (basic theoretical vs actual) is replaced by a new `CountVariance` struct that adds `cause_probabilities`, `severity`, `count_id`, and `period_start/end`. The existing `DetectVariances` method continues to work for real-time theoretical variance; `CountVariance` is the persisted, categorized version triggered by physical counts.

### Trigger

Runs when a count is submitted. Compares counted quantities against theoretical usage since last count.

### Actual Usage Calculation

```
actual_usage = opening_inventory + purchases_received + transfers_in - closing_count - transfers_out - recorded_waste
```

For this sprint, simplified to:
```
actual_usage = last_count_qty - current_count_qty + recorded_waste_since_last_count
```
(Purchases and transfers deferred to Sprint 17 PO workflow.)

### Cause Probability Rules

| Signal | Primary Cause | Probability Weight |
|--------|--------------|-------------------|
| High variance + no waste logged for ingredient | Unrecorded waste | 0.40 |
| High variance + waste logged but < variance | Portioning drift | 0.35 |
| Variance correlates with specific shifts (deferred to ML tier — requires shift-order join) | Individual portioning | 0.30 |
| Variance on expensive proteins only | Theft signal | 0.20 (flagged, never accused) |
| Vendor delivery discrepancy on file | Vendor spec change | 0.25 |
| Recipe recently changed | Recipe error | 0.30 |
| Variance is uniform across all ingredients | Measurement error | 0.35 |

Probabilities are normalized to sum to 1.0 for each variance record.

### Severity Thresholds

- **Critical:** variance > 15% AND dollar impact > 5000 cents ($50)
- **Warning:** variance > 10% OR dollar impact > 2500 cents ($25)
- **Info:** any variance > 5%
- Below 5% variance: no record created (within noise)

### Alert Integration

Critical and warning variances emit events on the event bus:
- `inventory.variance.critical` → alert enqueued automatically
- `inventory.variance.warning` → alert enqueued if dollar impact > 3000 cents ($30)

---

## 5. Tablet UI — Counting Flow

### Start Count Screen

- "New Full Count" and "New Spot Check" buttons
- Shows any in-progress counts for this location (resume option)
- Staff must be PIN'd in to start a count

### Count Entry Screen

- Ingredients grouped by category (collapsible sections): Protein, Produce, Dairy, Bakery, Frozen, Sauce
- Each row: ingredient name, unit label, large numeric input field
- **Expected quantity is NOT shown during counting** (prevents anchoring bias)
- Running progress bar: "32 of 48 items counted"
- Auto-save: lines saved to local storage on each entry, synced to API in batches
- Search bar at top to find specific ingredients

### Review Screen (before submit)

- Shows all counted items with expected vs actual side-by-side
- Highlights variances > 10% in amber, > 15% in red
- Staff can tap any item to revise count
- "Submit Count" button

### Offline Support

- Count lines stored in AsyncStorage during entry
- Background sync: when connectivity returns, batch POST to `/counts/:id/lines`
- Sync status indicator: green dot (synced), yellow dot (pending), red dot (offline)
- Conflict resolution: last write wins (batch upsert replaces existing line for same count_id + ingredient_id)

---

## 6. Tablet UI — Waste Logging Flow

### Quick Entry Form

- Ingredient picker (searchable dropdown)
- Quantity input (numeric, large touch target)
- Unit display (auto-filled from ingredient)
- Reason picker: expired, dropped, overcooked, contaminated, overproduction, other
- Optional note (text input)
- "Log Waste" submit button

### Today's Waste Feed

- Scrollable list of today's waste entries for this location
- Each entry shows: ingredient, quantity, reason badge, who logged it, time
- Swipe to delete (with confirmation) for corrections

### Photo Capture (stub)

- Camera button on waste form (placeholder in this sprint)
- Shows "Photo upload coming soon" toast
- S3 integration deferred to future sprint

---

## 7. Manager Web Dashboard — Variance View

Extends existing `InventoryPage.tsx` with a new tab/toggle.

### Variance Tab

- **Top Variances Table** — sorted by dollar impact descending
  - Columns: ingredient, category, expected qty, actual qty, variance %, variance $, severity badge
  - Cause probability: horizontal stacked bar per row (portioning, waste, recipe error, etc.)
  - Click row to drill down

### Drill-Down View

- **Count History** — list of counts for this ingredient, who counted, when, status
- **Waste Logs** — waste entries for this ingredient in the period
- **Trend Sparkline** — variance trend over last 30 days for this ingredient

### Count Management

- List of submitted counts awaiting approval
- Manager can review and approve counts
- Approved counts trigger variance recalculation

### Filters

- Date range picker
- Category filter (protein, produce, dairy, etc.)
- Severity filter (critical, warning, info)
- Location selector (uses existing location switcher)

---

## 8. Staff PIN System — Backend

### Existing Infrastructure

The `employees` table already has a `pin_hash TEXT` column (migration 001), and `internal/auth/service.go` already has a `PINLogin` method that queries employees by `location_id` where `pin_hash IS NOT NULL`.

### No Schema Changes Needed

PIN auth uses the existing `employees.pin_hash` column. No new migration for PIN support.

### PIN Verify Endpoint — Wraps Existing PINLogin

**POST /api/v1/auth/pin-verify** — Tablet-specific PIN verification within a manager session.

Request includes `location_id` (derived from manager's bound location in JWT or request body):
```json
Request:  { "pin": "123456", "location_id": "uuid" }
Response: { "employee_id": "uuid", "display_name": "Maria S.", "role": "line_cook" }
```

Implementation in `internal/auth/handler.go`:
1. Require valid manager JWT (auth middleware)
2. Extract `location_id` from request body
3. Call existing `PINLogin` logic — queries `employees` at that location with active status and non-null `pin_hash`
4. Iterate ALL active employees at the location (fix: existing PINLogin uses QueryRow which only checks one — must use Query to scan all and compare hashes)
5. Return employee_id, display_name, role (no JWT issued — tablet uses manager JWT for API calls)
6. Lockout tracking: use in-memory `sync.Map` keyed by `location_id:employee_id`, reset after 5 minutes

### PIN Assignment

For this sprint: staff PINs set via seed script (`bcrypt` hash inserted into `employees.pin_hash`).
PIN management UI deferred to a settings sprint.

### RBAC Permissions

New permissions added to `internal/auth/rbac.go` (additive to existing `inventory:read` and `inventory:write`):
- `inventory:count` — start and submit counts (roles: `staff`, `shift_manager`, `gm`, `owner`)
- `inventory:approve` — approve submitted counts (roles: `shift_manager`, `gm`, `owner`)
- `inventory:waste` — log waste entries (roles: `staff`, `shift_manager`, `gm`, `owner`)

Note: existing `inventory:read` and `inventory:write` remain unchanged for backward compatibility.

Existing `inventory:read` covers viewing variances and count history.

### Spot Check Scope

For spot checks, the staff selects a category filter (e.g., "Proteins only") when starting the count. The system returns only ingredients matching that category for the location. Full counts return all ingredients.

---

## 9. Dependencies & New Packages

### Go Backend
- No new Go dependencies (fpdf already in go.mod, bcrypt already used for passwords)

### Tablet App (new)
- expo (SDK 52)
- expo-router
- expo-secure-store
- @react-native-async-storage/async-storage
- zustand
- react-native-reanimated (for tab transitions)

---

## 10. Testing Strategy

### Backend Tests
- Variance calculation: known inputs → expected variance output
- Cause probability: test each signal rule produces correct attribution
- Severity thresholds: test boundary cases (4.9% vs 5.1%, $24 vs $26)
- PIN verify: correct PIN, wrong PIN, lockout after 5 attempts
- Count lifecycle: create → add lines → submit → variance generated → approve
- RLS: count from org A not visible to org B

### Tablet Tests
- Auth flow: manager login → PIN entry → session timeout → PIN re-entry
- Count flow: start → enter quantities → review → submit
- Offline: enter counts offline → come online → sync succeeds
- Waste: log entry → appears in today's feed
