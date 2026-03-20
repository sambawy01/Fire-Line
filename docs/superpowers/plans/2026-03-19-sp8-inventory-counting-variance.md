# SP8: Inventory Counting, Waste Logging & Variance Analysis — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the inventory counting/waste logging tablet app (Expo), backend APIs, variance categorization engine, and manager variance dashboard.

**Architecture:** New Expo React Native tablet app at `tablet/` with PIN auth, connecting to existing Go backend. Four new DB tables (migration 006) for counts, count lines, waste logs, and variances. Variance engine runs on count submission, categorizes causes probabilistically, emits alerts via event bus. Web dashboard extended with variance tab.

**Tech Stack:** Go 1.22, PostgreSQL 16, Expo SDK 52, Expo Router, Zustand, React Query, Tailwind CSS

---

## File Map

### Backend — New Files
| File | Responsibility |
|------|---------------|
| `migrations/006_inventory_counting.sql` | 4 new tables + RLS + indexes |
| `internal/inventory/counting.go` | Count session CRUD, count line upsert |
| `internal/inventory/waste.go` | Waste log CRUD |
| `internal/inventory/variance.go` | CountVariance type, cause categorization engine, severity thresholds |
| `internal/inventory/counting_test.go` | Unit tests for counting service |
| `internal/inventory/variance_test.go` | Unit tests for variance engine |
| `internal/api/counting_handler.go` | HTTP handlers for count + waste + variance endpoints |
| `scripts/seed_pins.sh` | Seed employee PINs for dev testing |

### Backend — Modified Files
| File | Change |
|------|--------|
| `internal/auth/service.go:229-271` | Fix PINLogin to query ALL employees (not QueryRow) |
| `internal/auth/handler.go:17-22` | Add `POST /api/v1/auth/pin-verify` route |
| `internal/auth/rbac.go` | Add `inventory:count`, `inventory:approve`, `inventory:waste` permissions |
| `internal/api/handlers.go:30-34` | Add new counting routes to InventoryHandler.RegisterRoutes |
| `cmd/fireline/main.go` | Wire counting handler routes |

### Tablet — New Files (entire project)
| File | Responsibility |
|------|---------------|
| `tablet/app/_layout.tsx` | Root layout with auth gate |
| `tablet/app/(auth)/login.tsx` | Manager email/password login |
| `tablet/app/(auth)/pin.tsx` | Staff PIN entry screen |
| `tablet/app/(tabs)/_layout.tsx` | 5-tab navigator |
| `tablet/app/(tabs)/count.tsx` | Inventory counting screen |
| `tablet/app/(tabs)/waste.tsx` | Waste logging screen |
| `tablet/app/(tabs)/tasks.tsx` | Placeholder |
| `tablet/app/(tabs)/kds.tsx` | Placeholder |
| `tablet/app/(tabs)/clock.tsx` | Placeholder |
| `tablet/components/PinPad.tsx` | 6-digit PIN keypad |
| `tablet/components/CountRow.tsx` | Ingredient count input row |
| `tablet/components/WasteForm.tsx` | Waste entry form |
| `tablet/components/CategoryGroup.tsx` | Collapsible ingredient group |
| `tablet/components/ProgressBar.tsx` | Count progress indicator |
| `tablet/lib/api.ts` | Fetch client with JWT |
| `tablet/lib/auth.ts` | Secure storage helpers |
| `tablet/lib/offline.ts` | AsyncStorage offline queue |
| `tablet/stores/auth.ts` | Zustand auth store |
| `tablet/stores/count.ts` | Zustand count session store |
| `tablet/stores/waste.ts` | Zustand waste form store |

### Web Dashboard — Modified Files
| File | Change |
|------|--------|
| `web/src/pages/InventoryPage.tsx` | Add Variance tab with table, cause bars, drill-down |
| `web/src/hooks/useInventory.ts` | Add `useVariances`, `useCounts`, `useWasteLogs` hooks |
| `web/src/lib/api.ts` | Add `CountVariance`, `WasteLog`, `InventoryCount` types + API methods |

---

## Task 1: Migration 006 — Counting Tables

**Files:**
- Create: `migrations/006_inventory_counting.sql`

- [ ] **Step 1: Write migration file**

```sql
-- Inventory counting, waste logging, and variance analysis

CREATE TABLE inventory_counts (
    count_id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID NOT NULL REFERENCES organizations(org_id),
    location_id    UUID NOT NULL REFERENCES locations(location_id),
    counted_by     UUID NOT NULL REFERENCES employees(employee_id),
    count_type     TEXT NOT NULL CHECK (count_type IN ('full', 'spot_check')),
    status         TEXT NOT NULL DEFAULT 'in_progress' CHECK (status IN ('in_progress', 'submitted', 'approved')),
    started_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    submitted_at   TIMESTAMPTZ,
    approved_by    UUID REFERENCES users(user_id),
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
    variance_cents      INT NOT NULL,
    cause_probabilities JSONB NOT NULL DEFAULT '{}',
    severity            TEXT NOT NULL DEFAULT 'info' CHECK (severity IN ('info', 'warning', 'critical')),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Indexes
CREATE INDEX idx_inventory_counts_location ON inventory_counts(org_id, location_id, started_at DESC);
CREATE INDEX idx_count_lines_count ON inventory_count_lines(count_id);
CREATE INDEX idx_count_lines_ingredient ON inventory_count_lines(org_id, ingredient_id);
CREATE INDEX idx_waste_logs_location ON waste_logs(org_id, location_id, logged_at DESC);
CREATE INDEX idx_variances_location ON inventory_variances(org_id, location_id, created_at DESC);

-- RLS: inventory_counts
ALTER TABLE inventory_counts ENABLE ROW LEVEL SECURITY;
ALTER TABLE inventory_counts FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON inventory_counts
    USING (org_id = current_setting('app.current_org_id')::UUID);
GRANT SELECT, INSERT, UPDATE, DELETE ON inventory_counts TO fireline_app;

-- RLS: inventory_count_lines
ALTER TABLE inventory_count_lines ENABLE ROW LEVEL SECURITY;
ALTER TABLE inventory_count_lines FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON inventory_count_lines
    USING (org_id = current_setting('app.current_org_id')::UUID);
GRANT SELECT, INSERT, UPDATE, DELETE ON inventory_count_lines TO fireline_app;

-- RLS: waste_logs
ALTER TABLE waste_logs ENABLE ROW LEVEL SECURITY;
ALTER TABLE waste_logs FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON waste_logs
    USING (org_id = current_setting('app.current_org_id')::UUID);
GRANT SELECT, INSERT, UPDATE, DELETE ON waste_logs TO fireline_app;

-- RLS: inventory_variances
ALTER TABLE inventory_variances ENABLE ROW LEVEL SECURITY;
ALTER TABLE inventory_variances FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON inventory_variances
    USING (org_id = current_setting('app.current_org_id')::UUID);
GRANT SELECT, INSERT, UPDATE, DELETE ON inventory_variances TO fireline_app;
```

- [ ] **Step 2: Update atlas hash and apply migration**

Run:
```bash
cd /Users/bistrocloud/Documents/AI_Restaurant_System/fireline
atlas migrate hash --dir file://migrations
atlas migrate apply --dir file://migrations --url "postgres://fireline:fireline@localhost:5432/fireline?sslmode=disable"
```
Expected: `Migrating to version 006... ok`

- [ ] **Step 3: Verify tables exist**

Run:
```bash
docker exec fireline-postgres-1 psql -U fireline -d fireline -c "\dt inventory_*" -c "\dt waste_logs"
```
Expected: 4 tables listed

- [ ] **Step 4: Commit**

```bash
git add migrations/006_inventory_counting.sql
git commit -m "feat: add migration 006 — inventory counting, waste logs, variances tables"
```

---

## Task 2: RBAC Permissions + PIN Auth Fix

**Files:**
- Modify: `internal/auth/rbac.go`
- Modify: `internal/auth/service.go:229-271`
- Modify: `internal/auth/handler.go`
- Create: `scripts/seed_pins.sh`
- Test: `internal/auth/rbac_test.go`

- [ ] **Step 1: Write test for new RBAC permissions**

Add to `internal/auth/rbac_test.go`:

```go
// Add these test cases to the existing TestHasPermission table:
{"staff", "inventory:count", true},
{"staff", "inventory:waste", true},
{"staff", "inventory:approve", false},
{"shift_manager", "inventory:count", true},
{"shift_manager", "inventory:waste", true},
{"shift_manager", "inventory:approve", true},
{"gm", "inventory:approve", true},
{"owner", "inventory:approve", true},
{"read_only", "inventory:count", false},
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/auth/ -run TestHasPermission -v`
Expected: FAIL — `inventory:count` not found for staff

- [ ] **Step 3: Add permissions to rbac.go**

In `internal/auth/rbac.go`, add to each role's permission slice:
- `staff`: add `"inventory:count"`, `"inventory:waste"`
- `shift_manager`: add `"inventory:count"`, `"inventory:waste"`, `"inventory:approve"`
- `gm`: add `"inventory:count"`, `"inventory:waste"`, `"inventory:approve"`
- `owner`: add `"inventory:count"`, `"inventory:waste"`, `"inventory:approve"`

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/auth/ -run TestHasPermission -v`
Expected: PASS

- [ ] **Step 5: Fix PINLogin to scan ALL employees**

In `internal/auth/service.go`, replace the `PINLogin` method (lines 229-271). Change `QueryRow` to `Query`, iterate all active employees at the location, compare PIN hash for each:

```go
func (s *Service) PINLogin(ctx context.Context, req PINLoginRequest) (*LoginResult, error) {
    rows, err := s.adminPool.Query(ctx,
        `SELECT e.employee_id, e.org_id, e.role, e.pin_hash, e.user_id, e.display_name
         FROM employees e
         WHERE e.location_id = $1 AND e.status = 'active' AND e.pin_hash IS NOT NULL`,
        req.LocationID,
    )
    if err != nil {
        return nil, fmt.Errorf("query employees: %w", err)
    }
    defer rows.Close()

    for rows.Next() {
        var employeeID, orgID, role, pinHash, displayName string
        var userID *string
        if err := rows.Scan(&employeeID, &orgID, &role, &pinHash, &userID, &displayName); err != nil {
            return nil, fmt.Errorf("scan employee: %w", err)
        }

        ok, err := VerifyPIN(pinHash, req.PIN)
        if err != nil || !ok {
            continue
        }

        // Found matching employee
        uid := employeeID
        if userID != nil {
            uid = *userID
        }

        accessToken, err := s.issuer.GenerateAccessToken(UserClaims{
            UserID: uid,
            OrgID:  orgID,
            Role:   role,
        })
        if err != nil {
            return nil, fmt.Errorf("generate access token: %w", err)
        }

        return &LoginResult{
            UserID:      uid,
            OrgID:       orgID,
            Role:        role,
            DisplayName: displayName,
            AccessToken: accessToken,
        }, nil
    }

    return nil, fmt.Errorf("invalid PIN")
}
```

- [ ] **Step 6: Add PIN verify handler to auth handler**

In `internal/auth/handler.go`, add the route and handler:

```go
// In RegisterRoutes, add:
mux.HandleFunc("POST /api/v1/auth/pin-verify", h.PINVerify)

// PIN lockout tracking (in-memory, keyed by location_id)
var pinAttempts sync.Map // key: locationID -> value: *attemptTracker

type attemptTracker struct {
    mu       sync.Mutex
    failures int
    lockedAt time.Time
}

func checkPINLockout(locationID string) bool {
    val, ok := pinAttempts.Load(locationID)
    if !ok { return false }
    t := val.(*attemptTracker)
    t.mu.Lock()
    defer t.mu.Unlock()
    if t.failures >= 5 && time.Since(t.lockedAt) < 5*time.Minute {
        return true // locked
    }
    if time.Since(t.lockedAt) >= 5*time.Minute {
        t.failures = 0 // reset after 5 min
    }
    return false
}

func recordPINFailure(locationID string) {
    val, _ := pinAttempts.LoadOrStore(locationID, &attemptTracker{})
    t := val.(*attemptTracker)
    t.mu.Lock()
    defer t.mu.Unlock()
    t.failures++
    if t.failures >= 5 { t.lockedAt = time.Now() }
}

func resetPINAttempts(locationID string) {
    pinAttempts.Delete(locationID)
}

// New handler method:
func (h *Handler) PINVerify(w http.ResponseWriter, r *http.Request) {
    var req struct {
        PIN        string `json:"pin"`
        LocationID string `json:"location_id"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeHandlerError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
        return
    }
    if req.PIN == "" || req.LocationID == "" {
        writeHandlerError(w, http.StatusBadRequest, "MISSING_FIELDS", "pin and location_id required")
        return
    }

    // Check lockout
    if checkPINLockout(req.LocationID) {
        writeHandlerError(w, http.StatusTooManyRequests, "PIN_LOCKED", "too many attempts, try again in 5 minutes")
        return
    }

    result, err := h.service.PINLogin(r.Context(), PINLoginRequest{
        LocationID: req.LocationID,
        PIN:        req.PIN,
    })
    if err != nil {
        recordPINFailure(req.LocationID)
        writeHandlerError(w, http.StatusUnauthorized, "PIN_FAILED", "invalid PIN")
        return
    }

    resetPINAttempts(req.LocationID)
    writeHandlerJSON(w, http.StatusOK, map[string]interface{}{
        "employee_id":  result.UserID,
        "display_name": result.DisplayName,
        "role":         result.Role,
    })
}
```

**Important:** Before this step, modify `LoginResult` in `internal/auth/service.go` (around line 123) to add:
```go
type LoginResult struct {
    UserID       string
    OrgID        string
    Role         string
    DisplayName  string  // ADD THIS FIELD
    AccessToken  string
    RefreshToken string
    MFARequired  bool
}
```

- [ ] **Step 7: Create PIN seed script**

PINs use argon2id (not bcrypt). The seed script must generate hashes via Go.

Create `scripts/seed_pins.sh`:

```bash
#!/usr/bin/env bash
# Seed employee PINs for development testing
# Uses Go to generate argon2id hashes (matching internal/auth/pin.go)
set -euo pipefail

cd /Users/bistrocloud/Documents/AI_Restaurant_System/fireline

# Generate argon2id hash for PIN "123456" using Go
PIN_HASH=$(go run -modfile go.mod - <<'GOEOF'
package main

import (
    "fmt"
    "github.com/opsnerve/fireline/internal/auth"
)

func main() {
    hash, err := auth.HashPIN("123456")
    if err != nil {
        panic(err)
    }
    fmt.Print(hash)
}
GOEOF
)

docker exec -i fireline-postgres-1 psql -U fireline -d fireline <<SQL
UPDATE employees SET pin_hash = '$PIN_HASH' WHERE status = 'active';
SQL

echo "All active employees now have PIN: 123456"
```

- [ ] **Step 8: Run tests**

Run: `go test ./internal/auth/ -v -count=1`
Expected: All auth tests pass

- [ ] **Step 9: Commit**

```bash
git add internal/auth/rbac.go internal/auth/rbac_test.go internal/auth/service.go internal/auth/handler.go scripts/seed_pins.sh
git commit -m "feat: add inventory RBAC permissions, fix PINLogin to scan all employees, add pin-verify endpoint"
```

---

## Task 3: Counting Service — Backend

**Files:**
- Create: `internal/inventory/counting.go`
- Create: `internal/inventory/counting_test.go`

- [ ] **Step 1: Write test for count lifecycle**

Create `internal/inventory/counting_test.go`:

```go
package inventory

import (
    "testing"
)

func TestCountSessionLifecycle(t *testing.T) {
    // Test that a count can be created with valid fields
    cs := CountSession{
        CountType:  "full",
        LocationID: "loc-1",
        CountedBy:  "emp-1",
    }
    if cs.CountType != "full" {
        t.Errorf("expected full, got %s", cs.CountType)
    }
}

func TestCountTypeValidation(t *testing.T) {
    valid := []string{"full", "spot_check"}
    invalid := []string{"partial", "", "FULL"}

    for _, v := range valid {
        if !validCountType(v) {
            t.Errorf("expected %q to be valid", v)
        }
    }
    for _, v := range invalid {
        if validCountType(v) {
            t.Errorf("expected %q to be invalid", v)
        }
    }
}

func TestCountProgress(t *testing.T) {
    lines := []CountLine{
        {IngredientID: "a", CountedQty: ptrFloat(10.0)},
        {IngredientID: "b", CountedQty: nil},
        {IngredientID: "c", CountedQty: ptrFloat(5.0)},
    }
    counted, total := countProgress(lines)
    if counted != 2 || total != 3 {
        t.Errorf("expected 2/3, got %d/%d", counted, total)
    }
}

func ptrFloat(f float64) *float64 { return &f }
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/inventory/ -run TestCount -v`
Expected: FAIL — types not defined

- [ ] **Step 3: Write counting.go**

Create `internal/inventory/counting.go`:

```go
package inventory

import (
    "context"
    "fmt"
    "time"

    "github.com/jackc/pgx/v5"
    "github.com/opsnerve/fireline/internal/tenant"
    "github.com/opsnerve/fireline/pkg/database"
)

type CountSession struct {
    CountID    string     `json:"count_id"`
    OrgID      string     `json:"org_id"`
    LocationID string     `json:"location_id"`
    CountedBy  string     `json:"counted_by"`
    CountType  string     `json:"count_type"`
    Status     string     `json:"status"`
    StartedAt  time.Time  `json:"started_at"`
    SubmittedAt *time.Time `json:"submitted_at,omitempty"`
    ApprovedBy *string    `json:"approved_by,omitempty"`
    ApprovedAt *time.Time `json:"approved_at,omitempty"`
}

type CountLine struct {
    CountLineID  string   `json:"count_line_id"`
    IngredientID string   `json:"ingredient_id"`
    Name         string   `json:"name"`
    Category     string   `json:"category"`
    ExpectedQty  *float64 `json:"expected_qty"`
    CountedQty   *float64 `json:"counted_qty"`
    Unit         string   `json:"unit"`
    Note         string   `json:"note"`
}

type CountWithLines struct {
    CountSession
    Lines    []CountLine `json:"lines"`
    Progress Progress    `json:"progress"`
}

type Progress struct {
    Counted int `json:"counted"`
    Total   int `json:"total"`
}

func validCountType(ct string) bool {
    return ct == "full" || ct == "spot_check"
}

func countProgress(lines []CountLine) (counted, total int) {
    total = len(lines)
    for _, l := range lines {
        if l.CountedQty != nil {
            counted++
        }
    }
    return
}

// CreateCount starts a new inventory count session.
func (s *Service) CreateCount(ctx context.Context, orgID, locationID, countedBy, countType string, category string) (*CountSession, error) {
    if !validCountType(countType) {
        return nil, fmt.Errorf("invalid count_type: %s", countType)
    }

    tenantCtx := tenant.WithOrgID(ctx, orgID)
    var cs CountSession

    err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
        err := tx.QueryRow(tenantCtx,
            `INSERT INTO inventory_counts (org_id, location_id, counted_by, count_type)
             VALUES ($1, $2, $3, $4)
             RETURNING count_id, org_id, location_id, counted_by, count_type, status, started_at`,
            orgID, locationID, countedBy, countType,
        ).Scan(&cs.CountID, &cs.OrgID, &cs.LocationID, &cs.CountedBy, &cs.CountType, &cs.Status, &cs.StartedAt)
        if err != nil {
            return fmt.Errorf("insert count: %w", err)
        }

        // Pre-populate count lines from ingredients at this location
        categoryFilter := ""
        args := []any{orgID, cs.CountID, locationID}
        if category != "" && countType == "spot_check" {
            categoryFilter = " AND i.category = $4"
            args = append(args, category)
        }

        _, err = tx.Exec(tenantCtx,
            `INSERT INTO inventory_count_lines (org_id, count_id, location_id, ingredient_id, unit)
             SELECT i.org_id, $2, $3, i.ingredient_id, i.unit
             FROM ingredients i
             WHERE i.org_id = $1 AND i.status = 'active'`+categoryFilter,
            args...,
        )
        if err != nil {
            return fmt.Errorf("populate lines: %w", err)
        }

        return nil
    })

    return &cs, err
}

// GetCount retrieves a count session with all its lines.
func (s *Service) GetCount(ctx context.Context, orgID, countID string) (*CountWithLines, error) {
    tenantCtx := tenant.WithOrgID(ctx, orgID)
    var result CountWithLines

    err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
        err := tx.QueryRow(tenantCtx,
            `SELECT count_id, org_id, location_id, counted_by, count_type, status, started_at, submitted_at, approved_by, approved_at
             FROM inventory_counts WHERE count_id = $1`,
            countID,
        ).Scan(&result.CountID, &result.OrgID, &result.LocationID, &result.CountedBy,
            &result.CountType, &result.Status, &result.StartedAt, &result.SubmittedAt,
            &result.ApprovedBy, &result.ApprovedAt)
        if err != nil {
            return fmt.Errorf("get count: %w", err)
        }

        rows, err := tx.Query(tenantCtx,
            `SELECT cl.count_line_id, cl.ingredient_id, i.name, i.category,
                    cl.expected_qty, cl.counted_qty, cl.unit, COALESCE(cl.note, '')
             FROM inventory_count_lines cl
             JOIN ingredients i ON i.ingredient_id = cl.ingredient_id
             WHERE cl.count_id = $1
             ORDER BY i.category, i.name`,
            countID,
        )
        if err != nil {
            return fmt.Errorf("get lines: %w", err)
        }
        defer rows.Close()

        for rows.Next() {
            var l CountLine
            if err := rows.Scan(&l.CountLineID, &l.IngredientID, &l.Name, &l.Category,
                &l.ExpectedQty, &l.CountedQty, &l.Unit, &l.Note); err != nil {
                return fmt.Errorf("scan line: %w", err)
            }
            result.Lines = append(result.Lines, l)
        }
        return rows.Err()
    })

    if err == nil {
        c, t := countProgress(result.Lines)
        result.Progress = Progress{Counted: c, Total: t}
    }
    return &result, err
}

// UpsertCountLines batch updates count lines with counted quantities.
func (s *Service) UpsertCountLines(ctx context.Context, orgID, countID string, lines []CountLineInput) (int, error) {
    tenantCtx := tenant.WithOrgID(ctx, orgID)
    var updated int

    err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
        for _, l := range lines {
            tag, err := tx.Exec(tenantCtx,
                `UPDATE inventory_count_lines
                 SET counted_qty = $1, note = $2, updated_at = now()
                 WHERE count_id = $3 AND ingredient_id = $4`,
                l.CountedQty, l.Note, countID, l.IngredientID,
            )
            if err != nil {
                return fmt.Errorf("upsert line %s: %w", l.IngredientID, err)
            }
            updated += int(tag.RowsAffected())
        }
        return nil
    })
    return updated, err
}

type CountLineInput struct {
    IngredientID string   `json:"ingredient_id"`
    CountedQty   float64  `json:"counted_qty"`
    Unit         string   `json:"unit"`
    Note         string   `json:"note"`
}

// SubmitCount marks a count as submitted and triggers variance calculation.
func (s *Service) SubmitCount(ctx context.Context, orgID, countID string) error {
    tenantCtx := tenant.WithOrgID(ctx, orgID)
    return database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
        tag, err := tx.Exec(tenantCtx,
            `UPDATE inventory_counts SET status = 'submitted', submitted_at = now(), updated_at = now()
             WHERE count_id = $1 AND status = 'in_progress'`,
            countID,
        )
        if err != nil {
            return fmt.Errorf("submit count: %w", err)
        }
        if tag.RowsAffected() == 0 {
            return fmt.Errorf("count not found or already submitted")
        }
        return nil
    })
}

// ApproveCount marks a count as approved by a manager.
func (s *Service) ApproveCount(ctx context.Context, orgID, countID, approvedBy string) error {
    tenantCtx := tenant.WithOrgID(ctx, orgID)
    return database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
        tag, err := tx.Exec(tenantCtx,
            `UPDATE inventory_counts SET status = 'approved', approved_by = $2, approved_at = now(), updated_at = now()
             WHERE count_id = $1 AND status = 'submitted'`,
            countID, approvedBy,
        )
        if err != nil {
            return fmt.Errorf("approve count: %w", err)
        }
        if tag.RowsAffected() == 0 {
            return fmt.Errorf("count not found or not submitted")
        }
        return nil
    })
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./internal/inventory/ -run TestCount -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/inventory/counting.go internal/inventory/counting_test.go
git commit -m "feat: add inventory counting service — session CRUD, line upsert, submit/approve"
```

---

## Task 4: Waste Logging Service — Backend

**Files:**
- Create: `internal/inventory/waste.go`

- [ ] **Step 1: Write waste.go**

Create `internal/inventory/waste.go`:

```go
package inventory

import (
    "context"
    "fmt"
    "time"

    "github.com/jackc/pgx/v5"
    "github.com/opsnerve/fireline/internal/tenant"
    "github.com/opsnerve/fireline/pkg/database"
)

type WasteLog struct {
    WasteID        string    `json:"waste_id"`
    IngredientID   string    `json:"ingredient_id"`
    IngredientName string    `json:"ingredient_name"`
    Quantity       float64   `json:"quantity"`
    Unit           string    `json:"unit"`
    Reason         string    `json:"reason"`
    LoggedBy       string    `json:"logged_by"`
    LoggedByName   string    `json:"logged_by_name"`
    LoggedAt       time.Time `json:"logged_at"`
    Note           string    `json:"note"`
}

var validReasons = map[string]bool{
    "expired": true, "dropped": true, "overcooked": true,
    "contaminated": true, "overproduction": true, "other": true,
}

type WasteInput struct {
    LocationID   string  `json:"location_id"`
    IngredientID string  `json:"ingredient_id"`
    Quantity     float64 `json:"quantity"`
    Unit         string  `json:"unit"`
    Reason       string  `json:"reason"`
    LoggedBy     string  `json:"logged_by"`
    Note         string  `json:"note"`
}

// LogWaste creates a waste log entry.
func (s *Service) LogWaste(ctx context.Context, orgID string, input WasteInput) (*WasteLog, error) {
    if !validReasons[input.Reason] {
        return nil, fmt.Errorf("invalid reason: %s", input.Reason)
    }

    tenantCtx := tenant.WithOrgID(ctx, orgID)
    var wl WasteLog

    err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
        return tx.QueryRow(tenantCtx,
            `INSERT INTO waste_logs (org_id, location_id, ingredient_id, quantity, unit, reason, logged_by, note)
             VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
             RETURNING waste_id, ingredient_id, quantity, unit, reason, logged_by, logged_at, COALESCE(note, '')`,
            orgID, input.LocationID, input.IngredientID, input.Quantity, input.Unit, input.Reason, input.LoggedBy, input.Note,
        ).Scan(&wl.WasteID, &wl.IngredientID, &wl.Quantity, &wl.Unit, &wl.Reason, &wl.LoggedBy, &wl.LoggedAt, &wl.Note)
    })
    return &wl, err
}

// ListWasteLogs returns waste logs for a location within a date range.
func (s *Service) ListWasteLogs(ctx context.Context, orgID, locationID string, from, to time.Time) ([]WasteLog, error) {
    tenantCtx := tenant.WithOrgID(ctx, orgID)
    var results []WasteLog

    err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
        rows, err := tx.Query(tenantCtx,
            `SELECT w.waste_id, w.ingredient_id, i.name, w.quantity, w.unit, w.reason,
                    w.logged_by, e.display_name, w.logged_at, COALESCE(w.note, '')
             FROM waste_logs w
             JOIN ingredients i ON i.ingredient_id = w.ingredient_id
             JOIN employees e ON e.employee_id = w.logged_by
             WHERE w.location_id = $1 AND w.logged_at >= $2 AND w.logged_at < $3
             ORDER BY w.logged_at DESC`,
            locationID, from, to,
        )
        if err != nil {
            return fmt.Errorf("query waste logs: %w", err)
        }
        defer rows.Close()

        for rows.Next() {
            var wl WasteLog
            if err := rows.Scan(&wl.WasteID, &wl.IngredientID, &wl.IngredientName, &wl.Quantity,
                &wl.Unit, &wl.Reason, &wl.LoggedBy, &wl.LoggedByName, &wl.LoggedAt, &wl.Note); err != nil {
                return fmt.Errorf("scan waste log: %w", err)
            }
            results = append(results, wl)
        }
        return rows.Err()
    })
    return results, err
}
```

- [ ] **Step 2: Verify build**

Run: `go build ./internal/inventory/`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add internal/inventory/waste.go
git commit -m "feat: add waste logging service — log entry and list by date range"
```

---

## Task 5: Variance Categorization Engine — Backend

**Files:**
- Create: `internal/inventory/variance.go`
- Create: `internal/inventory/variance_test.go`

- [ ] **Step 1: Write variance tests**

Create `internal/inventory/variance_test.go`:

```go
package inventory

import (
    "math"
    "testing"
)

func TestCategorizeCauses_UnrecordedWaste(t *testing.T) {
    signals := VarianceSignals{
        VariancePct:       20.0,
        WasteLogged:       0.0,
        VarianceQty:       5.0,
        IsExpensiveProtein: false,
    }
    probs := categorizeCauses(signals)
    if probs["unrecorded_waste"] < 0.3 {
        t.Errorf("expected unrecorded_waste to be dominant, got %v", probs)
    }
    assertNormalized(t, probs)
}

func TestCategorizeCauses_Portioning(t *testing.T) {
    signals := VarianceSignals{
        VariancePct:       12.0,
        WasteLogged:       1.0,
        VarianceQty:       5.0,
        IsExpensiveProtein: false,
    }
    probs := categorizeCauses(signals)
    if probs["portioning"] < 0.2 {
        t.Errorf("expected portioning to be significant, got %v", probs)
    }
    assertNormalized(t, probs)
}

func TestSeverityThresholds(t *testing.T) {
    tests := []struct {
        variancePct float64
        cents       int
        expected    string
    }{
        {16.0, 6000, "critical"},
        {16.0, 4000, "warning"},    // >15% but <5000 cents
        {11.0, 2000, "warning"},    // >10%
        {6.0, 3000, "warning"},     // >$25
        {6.0, 2000, "info"},        // >5%, <$25
        {4.0, 1000, ""},            // below noise threshold
    }
    for _, tt := range tests {
        got := classifySeverity(tt.variancePct, tt.cents)
        if got != tt.expected {
            t.Errorf("severity(%.1f%%, %d cents) = %q, want %q", tt.variancePct, tt.cents, got, tt.expected)
        }
    }
}

func assertNormalized(t *testing.T, probs map[string]float64) {
    t.Helper()
    sum := 0.0
    for _, v := range probs {
        sum += v
    }
    if math.Abs(sum-1.0) > 0.01 {
        t.Errorf("probabilities sum to %.4f, expected 1.0: %v", sum, probs)
    }
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/inventory/ -run "TestCategorize\|TestSeverity" -v`
Expected: FAIL — functions not defined

- [ ] **Step 3: Write variance.go**

Create `internal/inventory/variance.go`:

```go
package inventory

import (
    "context"
    "fmt"
    "math"
    "time"

    "github.com/jackc/pgx/v5"
    "github.com/opsnerve/fireline/internal/event"
    "github.com/opsnerve/fireline/internal/tenant"
    "github.com/opsnerve/fireline/pkg/database"
)

// CountVariance is the persisted, categorized variance from a physical count.
type CountVariance struct {
    VarianceID       string             `json:"variance_id"`
    IngredientID     string             `json:"ingredient_id"`
    IngredientName   string             `json:"ingredient_name"`
    Category         string             `json:"category"`
    TheoreticalUsage float64            `json:"theoretical_usage"`
    ActualUsage      float64            `json:"actual_usage"`
    VarianceQty      float64            `json:"variance_qty"`
    VariancePct      float64            `json:"variance_pct"`
    VarianceCents    int                `json:"variance_cents"`
    CauseProbabilities map[string]float64 `json:"cause_probabilities"`
    Severity         string             `json:"severity"`
    CreatedAt        time.Time          `json:"created_at"`
}

type VarianceSignals struct {
    VariancePct        float64
    WasteLogged        float64
    VarianceQty        float64
    IsExpensiveProtein bool
}

func categorizeCauses(s VarianceSignals) map[string]float64 {
    raw := make(map[string]float64)

    // No waste logged for this ingredient → unrecorded waste
    if s.WasteLogged == 0 && s.VarianceQty > 0 {
        raw["unrecorded_waste"] = 0.40
    } else if s.WasteLogged > 0 && s.WasteLogged < s.VarianceQty {
        raw["portioning"] = 0.35
        raw["unrecorded_waste"] = 0.15
    }

    // Expensive protein variance → theft signal
    if s.IsExpensiveProtein && s.VariancePct > 10 {
        raw["theft_signal"] = 0.20
    }

    // Default: some measurement error always present
    raw["measurement_error"] = 0.10

    // Fill in recipe_error as remainder
    raw["recipe_error"] = 0.10

    // Normalize to sum to 1.0
    return normalize(raw)
}

func normalize(probs map[string]float64) map[string]float64 {
    sum := 0.0
    for _, v := range probs {
        sum += v
    }
    if sum == 0 {
        return map[string]float64{"unknown": 1.0}
    }
    result := make(map[string]float64, len(probs))
    for k, v := range probs {
        result[k] = math.Round(v/sum*100) / 100
    }
    return result
}

func classifySeverity(variancePct float64, varianceCents int) string {
    absPct := math.Abs(variancePct)
    absCents := varianceCents
    if absCents < 0 {
        absCents = -absCents
    }

    if absPct > 15 && absCents > 5000 {
        return "critical"
    }
    if absPct > 10 || absCents > 2500 {
        return "warning"
    }
    if absPct > 5 {
        return "info"
    }
    return "" // below noise
}

// CalculateCountVariances runs the variance engine after a count is submitted.
func (s *Service) CalculateCountVariances(ctx context.Context, orgID, countID string) ([]CountVariance, error) {
    tenantCtx := tenant.WithOrgID(ctx, orgID)
    var results []CountVariance

    err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
        // Get count details
        var locationID string
        var countEnd time.Time
        err := tx.QueryRow(tenantCtx,
            `SELECT location_id, submitted_at FROM inventory_counts WHERE count_id = $1`,
            countID,
        ).Scan(&locationID, &countEnd)
        if err != nil {
            return fmt.Errorf("get count: %w", err)
        }

        // Find the previous count's submission time as period start
        var periodStart time.Time
        err = tx.QueryRow(tenantCtx,
            `SELECT COALESCE(MAX(submitted_at), $2 - INTERVAL '30 days')
             FROM inventory_counts
             WHERE location_id = $1 AND status IN ('submitted', 'approved')
               AND count_id != $3 AND submitted_at < $2`,
            locationID, countEnd, countID,
        ).Scan(&periodStart)
        if err != nil {
            periodStart = countEnd.AddDate(0, 0, -30)
        }

        // Get counted lines with ingredient cost info
        rows, err := tx.Query(tenantCtx,
            `SELECT cl.ingredient_id, i.name, i.category, cl.counted_qty, i.cost_per_unit, i.unit
             FROM inventory_count_lines cl
             JOIN ingredients i ON i.ingredient_id = cl.ingredient_id
             WHERE cl.count_id = $1 AND cl.counted_qty IS NOT NULL`,
            countID,
        )
        if err != nil {
            return fmt.Errorf("get lines: %w", err)
        }
        defer rows.Close()

        type lineInfo struct {
            ingredientID string
            name         string
            category     string
            countedQty   float64
            costPerUnit  int64
            unit         string
        }
        var lines []lineInfo

        for rows.Next() {
            var li lineInfo
            if err := rows.Scan(&li.ingredientID, &li.name, &li.category, &li.countedQty, &li.costPerUnit, &li.unit); err != nil {
                return fmt.Errorf("scan line: %w", err)
            }
            lines = append(lines, li)
        }
        if err := rows.Err(); err != nil {
            return err
        }

        // Get theoretical usage for the period
        theoretical, err := s.CalculateTheoreticalUsage(ctx, orgID, locationID, periodStart, countEnd)
        if err != nil {
            return fmt.Errorf("theoretical usage: %w", err)
        }
        theoMap := make(map[string]float64)
        for _, t := range theoretical {
            theoMap[t.IngredientID] = t.TotalUsed
        }

        // Get waste logged per ingredient in the period
        wasteRows, err := tx.Query(tenantCtx,
            `SELECT ingredient_id, SUM(quantity) FROM waste_logs
             WHERE location_id = $1 AND logged_at >= $2 AND logged_at < $3
             GROUP BY ingredient_id`,
            locationID, periodStart, countEnd,
        )
        if err != nil {
            return fmt.Errorf("waste query: %w", err)
        }
        defer wasteRows.Close()
        wasteMap := make(map[string]float64)
        for wasteRows.Next() {
            var id string
            var qty float64
            if err := wasteRows.Scan(&id, &qty); err != nil {
                return err
            }
            wasteMap[id] = qty
        }

        // Calculate variances
        for _, li := range lines {
            theo := theoMap[li.ingredientID]
            actualUsage := li.countedQty // simplified: the counted qty represents actual on-hand
            varianceQty := actualUsage - theo
            variancePct := 0.0
            if theo > 0 {
                variancePct = (varianceQty / theo) * 100
            }
            varianceCents := int(varianceQty * float64(li.costPerUnit))

            severity := classifySeverity(variancePct, varianceCents)
            if severity == "" {
                continue // below noise threshold
            }

            isExpensiveProtein := li.category == "protein" && li.costPerUnit > 400

            causes := categorizeCauses(VarianceSignals{
                VariancePct:        variancePct,
                WasteLogged:        wasteMap[li.ingredientID],
                VarianceQty:        math.Abs(varianceQty),
                IsExpensiveProtein: isExpensiveProtein,
            })

            var vid string
            err := tx.QueryRow(tenantCtx,
                `INSERT INTO inventory_variances
                 (org_id, location_id, ingredient_id, count_id, period_start, period_end,
                  theoretical_usage, actual_usage, variance_qty, variance_cents, cause_probabilities, severity)
                 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
                 RETURNING variance_id`,
                orgID, locationID, li.ingredientID, countID,
                periodStart, countEnd, theo, actualUsage,
                varianceQty, varianceCents, causes, severity,
            ).Scan(&vid)
            if err != nil {
                return fmt.Errorf("insert variance: %w", err)
            }

            results = append(results, CountVariance{
                VarianceID:         vid,
                IngredientID:       li.ingredientID,
                IngredientName:     li.name,
                Category:           li.category,
                TheoreticalUsage:   theo,
                ActualUsage:        actualUsage,
                VarianceQty:        varianceQty,
                VariancePct:        variancePct,
                VarianceCents:      varianceCents,
                CauseProbabilities: causes,
                Severity:           severity,
            })
        }

        return nil
    })

    // Emit alert events
    if err == nil {
        for _, v := range results {
            if v.Severity == "critical" {
                s.bus.Publish(ctx, event.Envelope{
                    EventType:  "inventory.variance.critical",
                    OrgID:      orgID,
                    Source:     "inventory",
                    Payload:    v,
                })
            } else if v.Severity == "warning" && (v.VarianceCents > 3000 || v.VarianceCents < -3000) {
                s.bus.Publish(ctx, event.Envelope{
                    EventType:  "inventory.variance.warning",
                    OrgID:      orgID,
                    Source:     "inventory",
                    Payload:    v,
                })
            }
        }
    }

    return results, err
}

// ListVariances returns persisted variances for a location in a date range.
func (s *Service) ListVariances(ctx context.Context, orgID, locationID string, from, to time.Time) ([]CountVariance, error) {
    tenantCtx := tenant.WithOrgID(ctx, orgID)
    var results []CountVariance

    err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
        rows, err := tx.Query(tenantCtx,
            `SELECT v.variance_id, v.ingredient_id, i.name, i.category,
                    v.theoretical_usage, v.actual_usage, v.variance_qty, v.variance_cents,
                    v.cause_probabilities, v.severity, v.created_at
             FROM inventory_variances v
             JOIN ingredients i ON i.ingredient_id = v.ingredient_id
             WHERE v.location_id = $1 AND v.created_at >= $2 AND v.created_at < $3
             ORDER BY ABS(v.variance_cents) DESC`,
            locationID, from, to,
        )
        if err != nil {
            return err
        }
        defer rows.Close()

        for rows.Next() {
            var cv CountVariance
            if err := rows.Scan(&cv.VarianceID, &cv.IngredientID, &cv.IngredientName, &cv.Category,
                &cv.TheoreticalUsage, &cv.ActualUsage, &cv.VarianceQty, &cv.VarianceCents,
                &cv.CauseProbabilities, &cv.Severity, &cv.CreatedAt); err != nil {
                return err
            }
            if cv.TheoreticalUsage > 0 {
                cv.VariancePct = (cv.VarianceQty / cv.TheoreticalUsage) * 100
            }
            results = append(results, cv)
        }
        return rows.Err()
    })
    return results, err
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./internal/inventory/ -run "TestCategorize\|TestSeverity" -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/inventory/variance.go internal/inventory/variance_test.go
git commit -m "feat: add variance categorization engine — cause probabilities, severity classification, alert integration"
```

---

## Task 6: HTTP Handlers — Count, Waste, Variance Endpoints

**Files:**
- Create: `internal/api/counting_handler.go`
- Modify: `cmd/fireline/main.go`

- [ ] **Step 1: Write counting_handler.go**

Create `internal/api/counting_handler.go` with handlers for all 6 endpoints:
- `POST /api/v1/inventory/counts` — CreateCount
- `GET /api/v1/inventory/counts/{id}` — GetCount
- `PUT /api/v1/inventory/counts/{id}` — UpdateCountStatus (submit/approve)
- `POST /api/v1/inventory/counts/{id}/lines` — UpsertLines
- `POST /api/v1/inventory/waste` — LogWaste
- `GET /api/v1/inventory/waste` — ListWasteLogs
- `DELETE /api/v1/inventory/waste/{id}` — DeleteWaste (corrections)
- `GET /api/v1/inventory/variances` — ListVariances

**Important notes for handler implementation:**
- `counted_by` in CreateCount: read from request body (tablet sends the active staff's employee_id)
- `logged_by` in LogWaste: read from request body (tablet sends the active staff's employee_id)
- `category` for spot checks: add optional `category` field to CreateCount request body

Follow the pattern from `vendor_handler.go`: extract orgID from tenant context, locationID from query params, decode JSON body, call service, return JSON response.

Register routes in `InventoryHandler.RegisterRoutes` by adding new mux.Handle lines alongside the existing usage/par/explode routes.

- [ ] **Step 2: Wire in main.go**

No changes needed — the existing `invHandler.RegisterRoutes(mux, authMW)` call in `main.go` already registers all InventoryHandler routes. The new routes are added within `RegisterRoutes`.

- [ ] **Step 3: Build and verify**

Run: `go build ./cmd/fireline/`
Expected: No errors

- [ ] **Step 4: Restart server and test endpoints**

```bash
# Kill old server, rebuild, restart
pkill -f "./fireline" || true
DATABASE_URL="postgres://fireline_app:fireline_app@localhost:5432/fireline?sslmode=disable" \
ADMIN_DATABASE_URL="postgres://fireline:fireline@localhost:5432/fireline?sslmode=disable" \
ENV=development PORT=8080 ./fireline &
sleep 2

# Test health
curl -s http://localhost:8080/health/live
```

- [ ] **Step 5: Commit**

```bash
git add internal/api/counting_handler.go internal/api/handlers.go
git commit -m "feat: add HTTP handlers for counting, waste logging, and variance endpoints"
```

---

## Task 7: Expo Tablet App — Scaffold + Auth

**Files:**
- Create: `tablet/` (entire Expo project)

- [ ] **Step 1: Create Expo project**

```bash
cd /Users/bistrocloud/Documents/AI_Restaurant_System/fireline
npx create-expo-app@latest tablet --template blank-typescript
cd tablet
npx expo install expo-router expo-secure-store @react-native-async-storage/async-storage react-native-reanimated
npm install zustand
```

- [ ] **Step 2: Configure Expo Router**

Update `tablet/app.json` to use expo-router scheme. Create `tablet/app/_layout.tsx` as root layout with auth gate.

- [ ] **Step 3: Create lib/api.ts**

Fetch client that reads JWT from expo-secure-store, base URL configurable (default `http://localhost:8080/api/v1`).

- [ ] **Step 4: Create stores/auth.ts**

Zustand store with:
- `managerToken: string | null`
- `activeStaff: { employee_id, display_name, role } | null`
- `locationId: string | null`
- `lastActivity: number`
- `login(email, password)` — calls `/auth/login`, stores JWT
- `pinVerify(pin)` — calls `/auth/pin-verify`, sets active staff
- `checkTimeout()` — if `Date.now() - lastActivity > 120000`, clear active staff
- `logout()` — clear everything

- [ ] **Step 5: Create app/(auth)/login.tsx**

Manager login screen: email input, password input, "Login" button. On success → location selection → redirect to tabs.

- [ ] **Step 6: Create components/PinPad.tsx**

6-digit PIN keypad: 0-9 buttons in phone layout, display dots for entered digits, backspace button, auto-submit on 6th digit.

- [ ] **Step 7: Create app/(auth)/pin.tsx**

PIN entry screen showing location name, staff name prompt, PinPad component. On success → navigate to tabs. Show error on invalid PIN. Lock after 5 attempts.

- [ ] **Step 8: Create app/(tabs)/_layout.tsx**

Tab navigator with 5 tabs: Count, Waste, Tasks, KDS, Clock. Last 3 show "Coming Soon" placeholder.

- [ ] **Step 9: Create placeholder tabs**

`tasks.tsx`, `kds.tsx`, `clock.tsx` — each renders centered "Coming Soon" text.

- [ ] **Step 10: Verify app builds**

```bash
cd /Users/bistrocloud/Documents/AI_Restaurant_System/fireline/tablet
npx expo start --clear
```
Expected: Expo dev server starts, app renders login screen

- [ ] **Step 11: Commit**

```bash
git add tablet/
git commit -m "feat: scaffold Expo tablet app with auth flow, PIN entry, and 5-tab navigation shell"
```

---

## Task 8: Tablet — Counting Screen

**Files:**
- Create: `tablet/app/(tabs)/count.tsx`
- Create: `tablet/components/CountRow.tsx`
- Create: `tablet/components/CategoryGroup.tsx`
- Create: `tablet/components/ProgressBar.tsx`
- Create: `tablet/stores/count.ts`
- Create: `tablet/lib/offline.ts`

- [ ] **Step 1: Create count store**

`tablet/stores/count.ts` — Zustand store:
- `activeCount: CountSession | null`
- `lines: CountLine[]`
- `startCount(type, category?)` — POST `/inventory/counts`
- `loadCount(id)` — GET `/inventory/counts/:id`
- `updateLine(ingredientId, qty, note)` — update local, queue for sync
- `submitCount()` — PUT `/inventory/counts/:id` with status=submitted
- `syncLines()` — POST `/inventory/counts/:id/lines` with pending changes

- [ ] **Step 2: Create offline.ts**

AsyncStorage queue: `enqueue(key, data)`, `dequeue(key)`, `getPending(key)`. Used by count store to persist unsaved lines.

- [ ] **Step 3: Create ProgressBar component**

Simple horizontal bar showing `counted/total` with label text.

- [ ] **Step 4: Create CategoryGroup component**

Collapsible section: category name header (tap to expand/collapse), children rendered inside.

- [ ] **Step 5: Create CountRow component**

Row showing: ingredient name, unit label, large numeric TextInput (48px+ height), optional note icon.

- [ ] **Step 6: Create count.tsx screen**

Three states:
1. **Start screen**: "New Full Count" / "New Spot Check" buttons + list of in-progress counts
2. **Count entry**: ingredient list grouped by CategoryGroup, CountRow for each, ProgressBar at top, search bar
3. **Review screen**: shows all items with expected vs actual (expected only shown here), highlight variances, "Submit" button

- [ ] **Step 7: Test on Expo Go**

Open Expo Go on device/simulator, verify:
- Can start a new count
- Ingredients load grouped by category
- Can enter quantities
- Progress bar updates
- Submit sends data to backend

- [ ] **Step 8: Commit**

```bash
git add tablet/app/\(tabs\)/count.tsx tablet/components/ tablet/stores/count.ts tablet/lib/offline.ts
git commit -m "feat: add tablet inventory counting screen with category groups, progress tracking, and offline sync"
```

---

## Task 9: Tablet — Waste Logging Screen

**Files:**
- Create: `tablet/app/(tabs)/waste.tsx`
- Create: `tablet/components/WasteForm.tsx`
- Create: `tablet/stores/waste.ts`

- [ ] **Step 1: Create waste store**

`tablet/stores/waste.ts` — Zustand store:
- `todaysLogs: WasteLog[]`
- `loadLogs()` — GET `/inventory/waste?location_id=X&from=today`
- `logWaste(input)` — POST `/inventory/waste`, add to todaysLogs

- [ ] **Step 2: Create WasteForm component**

Form with: searchable ingredient picker, quantity input (large), unit display (auto), reason picker (6 options as pills/buttons), note input, "Log Waste" button.

- [ ] **Step 3: Create waste.tsx screen**

Split view: WasteForm at top, Today's Waste feed (FlatList) below. Each feed item shows ingredient name, quantity, reason badge (colored), logged by, time.

- [ ] **Step 4: Test**

Verify waste entry creates record, appears in feed, persists on reload.

- [ ] **Step 5: Commit**

```bash
git add tablet/app/\(tabs\)/waste.tsx tablet/components/WasteForm.tsx tablet/stores/waste.ts
git commit -m "feat: add tablet waste logging screen with quick entry form and today's waste feed"
```

---

## Task 10: Web Dashboard — Variance Tab

**Files:**
- Modify: `web/src/lib/api.ts`
- Modify: `web/src/hooks/useInventory.ts`
- Modify: `web/src/pages/InventoryPage.tsx`

- [ ] **Step 1: Add types and API methods to api.ts**

In `web/src/lib/api.ts`, add:
- `CountVariance` type (variance_id, ingredient_name, category, theoretical_usage, actual_usage, variance_qty, variance_pct, variance_cents, cause_probabilities, severity)
- `WasteLog` type
- `InventoryCount` type
- `inventoryApi.getVariances(locationId, from?, to?)`
- `inventoryApi.getWasteLogs(locationId, from?, to?)`
- `inventoryApi.approvCount(countId)`

- [ ] **Step 2: Add hooks to useInventory.ts**

Add `useVariances(locationId)`, `useWasteLogs(locationId)` React Query hooks.

- [ ] **Step 3: Update InventoryPage.tsx**

Add tab bar at top: "Usage" | "PAR Status" | "Variances"

Variances tab shows:
- DataTable with columns: ingredient, category, expected, actual, variance %, variance $, severity badge
- Each row has a horizontal stacked bar for cause_probabilities (colored segments)
- Click row → expandable drill-down showing count history + waste logs for that ingredient

- [ ] **Step 4: Build and verify**

```bash
cd /Users/bistrocloud/Documents/AI_Restaurant_System/fireline/web
npx vite build
```
Expected: Build succeeds

- [ ] **Step 5: Test in browser**

Open http://localhost:5173, navigate to Inventory, click Variances tab. Verify table renders (may be empty until counts are submitted).

- [ ] **Step 6: Commit**

```bash
git add web/src/lib/api.ts web/src/hooks/useInventory.ts web/src/pages/InventoryPage.tsx
git commit -m "feat: add variance tab to inventory dashboard with cause probability bars and drill-down"
```

---

## Task 11: Seed PINs + End-to-End Test

**Files:**
- Modify: `scripts/seed_pins.sh`

- [ ] **Step 1: Run PIN seed script**

```bash
chmod +x scripts/seed_pins.sh
./scripts/seed_pins.sh
```
Expected: "All active employees now have PIN: 123456"

- [ ] **Step 2: Test PIN verify endpoint**

```bash
# First login as manager to get JWT
TOKEN=$(curl -s http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"owner@bistrocloud.com","password":"DemoPassword1234!"}' \
  | python3 -c "import sys,json; print(json.load(sys.stdin)['access_token'])")

# Verify PIN
curl -s http://localhost:8080/api/v1/auth/pin-verify \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"pin":"123456","location_id":"a1111111-1111-1111-1111-111111111111"}'
```
Expected: Returns employee_id and display_name

- [ ] **Step 3: Test count lifecycle**

```bash
# Create count
COUNT=$(curl -s -X POST http://localhost:8080/api/v1/inventory/counts \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"location_id":"a1111111-1111-1111-1111-111111111111","count_type":"full","counted_by":"ee111111-1111-1111-1111-111111111111"}')
echo "$COUNT"

COUNT_ID=$(echo "$COUNT" | python3 -c "import sys,json; print(json.load(sys.stdin)['count_id'])")

# Get count with lines
curl -s http://localhost:8080/api/v1/inventory/counts/$COUNT_ID \
  -H "Authorization: Bearer $TOKEN"

# Submit count
curl -s -X PUT http://localhost:8080/api/v1/inventory/counts/$COUNT_ID \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"status":"submitted"}'
```

- [ ] **Step 4: Test waste logging**

```bash
curl -s -X POST http://localhost:8080/api/v1/inventory/waste \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"location_id":"a1111111-1111-1111-1111-111111111111","ingredient_id":"11111111-aaaa-1111-aaaa-111111111111","quantity":2.5,"unit":"lb","reason":"expired","logged_by":"ee111111-3333-3333-3333-333333333333","note":"found in back of walk-in"}'
```
Expected: Returns waste_id and logged_at

- [ ] **Step 5: Run all Go tests**

```bash
cd /Users/bistrocloud/Documents/AI_Restaurant_System/fireline
go test ./... -count=1 -p 1
```
Expected: All packages pass

- [ ] **Step 6: Final commit**

```bash
git add -A
git commit -m "feat: SP8 complete — inventory counting, waste logging, variance engine, tablet app"
```
