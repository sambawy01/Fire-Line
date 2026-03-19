# SP3: Labor & Workforce Intelligence Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build employee roster display, shift tracking, and labor cost dashboard with labor cost % of revenue — the #1 metric every restaurant owner watches.

**Architecture:** New migration adds `shifts` table. New Go service `internal/labor/` queries employees + shifts + checks via TenantTx. HTTP handler exposes summary and employee list endpoints. React frontend adds a Labor page with KPI cards and employee DataTable. Demo seed script adds employees + shifts.

**Tech Stack:** Go 1.22+ (pgx/v5, TenantTx), PostgreSQL 16, React 19, TypeScript, Tailwind CSS 4, TanStack React Query, Lucide icons.

**Spec:** `docs/superpowers/specs/2026-03-19-sp3-labor-intelligence-design.md`

---

### Task 1: Database Migration — Shifts Table

**Files:**
- Create: `migrations/004_shifts.sql`

- [ ] **Step 1: Create the migration file**

```sql
-- Shift tracking for labor cost intelligence

CREATE TABLE shifts (
    shift_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(org_id),
    location_id UUID NOT NULL REFERENCES locations(location_id),
    employee_id UUID NOT NULL REFERENCES employees(employee_id),
    role TEXT NOT NULL DEFAULT 'staff',
    clock_in TIMESTAMPTZ NOT NULL,
    clock_out TIMESTAMPTZ,
    hourly_rate INT NOT NULL DEFAULT 0 CHECK (hourly_rate >= 0),
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'completed', 'no_show')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (clock_out IS NULL OR clock_out > clock_in)
);

CREATE INDEX idx_shifts_org ON shifts(org_id);
CREATE INDEX idx_shifts_location ON shifts(org_id, location_id);
CREATE INDEX idx_shifts_employee ON shifts(employee_id);
CREATE INDEX idx_shifts_clock ON shifts(org_id, location_id, clock_in DESC);

ALTER TABLE shifts ENABLE ROW LEVEL SECURITY;
ALTER TABLE shifts FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON shifts
    USING (org_id = current_setting('app.current_org_id')::UUID);
GRANT SELECT, INSERT, UPDATE, DELETE ON shifts TO fireline_app;
```

- [ ] **Step 2: Regenerate atlas hash**

Run: `cd /Users/bistrocloud/Documents/AI_Restaurant_System/fireline && atlas migrate hash --dir file://migrations`

If `atlas` is not installed, manually update the hash file — or run the migration directly via docker:

```bash
docker exec -i fireline-postgres-1 psql -U fireline -d fireline < migrations/004_shifts.sql
```

- [ ] **Step 3: Apply migration**

```bash
docker exec -i fireline-postgres-1 psql -U fireline -d fireline < migrations/004_shifts.sql
```

Verify:
```bash
docker exec fireline-postgres-1 psql -U fireline -d fireline -c "\d shifts"
```
Expected: Table with all columns including CHECK constraints.

- [ ] **Step 4: Commit**

```bash
git add migrations/004_shifts.sql
git commit -m "feat: add shifts table migration for labor intelligence"
```

---

### Task 2: Demo Seed — Employees + Shifts

**Files:**
- Create: `scripts/seed_labor.sh`

- [ ] **Step 1: Create the labor seed script**

```bash
#!/usr/bin/env bash
# Seed demo employees and shifts for labor intelligence
set -euo pipefail

echo "=== Seeding labor demo data ==="

# Get the org_id for Bistro Cloud
ORG_ID=$(docker exec fireline-postgres-1 psql -U fireline -d fireline -t -c \
  "SELECT org_id FROM organizations WHERE slug = 'bistro-cloud'" | tr -d ' \n')

if [ -z "$ORG_ID" ]; then
  echo "ERROR: Bistro Cloud org not found. Run scripts/seed_demo.sh first."
  exit 1
fi

echo "org_id: $ORG_ID"

docker exec -i fireline-postgres-1 psql -U fireline -d fireline <<SQL

-- Clean existing labor data
DELETE FROM shifts WHERE org_id = '$ORG_ID';
DELETE FROM employees WHERE org_id = '$ORG_ID';

-- === EMPLOYEES ===
-- Downtown Flagship (5 employees)
INSERT INTO employees (employee_id, org_id, location_id, display_name, role, status) VALUES
  ('ee111111-1111-1111-1111-111111111111', '$ORG_ID', 'a1111111-1111-1111-1111-111111111111', 'Maria Santos', 'gm', 'active'),
  ('ee111111-2222-2222-2222-222222222222', '$ORG_ID', 'a1111111-1111-1111-1111-111111111111', 'Jake Thompson', 'shift_manager', 'active'),
  ('ee111111-3333-3333-3333-333333333333', '$ORG_ID', 'a1111111-1111-1111-1111-111111111111', 'Sarah Chen', 'staff', 'active'),
  ('ee111111-4444-4444-4444-444444444444', '$ORG_ID', 'a1111111-1111-1111-1111-111111111111', 'Marcus Brown', 'staff', 'active'),
  ('ee111111-5555-5555-5555-555555555555', '$ORG_ID', 'a1111111-1111-1111-1111-111111111111', 'Lily Nguyen', 'staff', 'active');

-- Airport Terminal 4 (3 employees)
INSERT INTO employees (employee_id, org_id, location_id, display_name, role, status) VALUES
  ('ee222222-1111-1111-1111-111111111111', '$ORG_ID', 'b2222222-2222-2222-2222-222222222222', 'Priya Patel', 'gm', 'active'),
  ('ee222222-2222-2222-2222-222222222222', '$ORG_ID', 'b2222222-2222-2222-2222-222222222222', 'David Kim', 'shift_manager', 'active'),
  ('ee222222-3333-3333-3333-333333333333', '$ORG_ID', 'b2222222-2222-2222-2222-222222222222', 'Emily Zhao', 'staff', 'active');

-- === SHIFTS (today) ===
-- Downtown: busy day, multiple shifts
-- Maria Santos (GM) - full day 8am-5pm, completed
INSERT INTO shifts (org_id, location_id, employee_id, role, clock_in, clock_out, hourly_rate, status) VALUES
  ('$ORG_ID', 'a1111111-1111-1111-1111-111111111111', 'ee111111-1111-1111-1111-111111111111', 'gm',
   now()::date + INTERVAL '8 hours', now()::date + INTERVAL '17 hours', 2800, 'completed');

-- Jake Thompson (Shift Mgr) - morning 6am-2pm, completed
INSERT INTO shifts (org_id, location_id, employee_id, role, clock_in, clock_out, hourly_rate, status) VALUES
  ('$ORG_ID', 'a1111111-1111-1111-1111-111111111111', 'ee111111-2222-2222-2222-222222222222', 'shift_manager',
   now()::date + INTERVAL '6 hours', now()::date + INTERVAL '14 hours', 2200, 'completed');

-- Jake Thompson (Shift Mgr) - dinner 4pm-still active
INSERT INTO shifts (org_id, location_id, employee_id, role, clock_in, clock_out, hourly_rate, status) VALUES
  ('$ORG_ID', 'a1111111-1111-1111-1111-111111111111', 'ee111111-2222-2222-2222-222222222222', 'shift_manager',
   now()::date + INTERVAL '16 hours', NULL, 2200, 'active');

-- Sarah Chen (Line Cook) - morning prep 6am-2pm, completed
INSERT INTO shifts (org_id, location_id, employee_id, role, clock_in, clock_out, hourly_rate, status) VALUES
  ('$ORG_ID', 'a1111111-1111-1111-1111-111111111111', 'ee111111-3333-3333-3333-333333333333', 'staff',
   now()::date + INTERVAL '6 hours', now()::date + INTERVAL '14 hours', 1800, 'completed');

-- Sarah Chen (Line Cook) - dinner 4pm-still active
INSERT INTO shifts (org_id, location_id, employee_id, role, clock_in, clock_out, hourly_rate, status) VALUES
  ('$ORG_ID', 'a1111111-1111-1111-1111-111111111111', 'ee111111-3333-3333-3333-333333333333', 'staff',
   now()::date + INTERVAL '16 hours', NULL, 1800, 'active');

-- Marcus Brown (Server) - lunch 10am-4pm, completed
INSERT INTO shifts (org_id, location_id, employee_id, role, clock_in, clock_out, hourly_rate, status) VALUES
  ('$ORG_ID', 'a1111111-1111-1111-1111-111111111111', 'ee111111-4444-4444-4444-444444444444', 'staff',
   now()::date + INTERVAL '10 hours', now()::date + INTERVAL '16 hours', 1500, 'completed');

-- Marcus Brown (Server) - dinner 5pm-still active
INSERT INTO shifts (org_id, location_id, employee_id, role, clock_in, clock_out, hourly_rate, status) VALUES
  ('$ORG_ID', 'a1111111-1111-1111-1111-111111111111', 'ee111111-4444-4444-4444-444444444444', 'staff',
   now()::date + INTERVAL '17 hours', NULL, 1500, 'active');

-- Lily Nguyen (Server) - lunch+dinner 11am-8pm, completed
INSERT INTO shifts (org_id, location_id, employee_id, role, clock_in, clock_out, hourly_rate, status) VALUES
  ('$ORG_ID', 'a1111111-1111-1111-1111-111111111111', 'ee111111-5555-5555-5555-555555555555', 'staff',
   now()::date + INTERVAL '11 hours', now()::date + INTERVAL '20 hours', 1500, 'completed');

-- Airport: smaller crew
-- Priya Patel (GM) - 7am-4pm, completed
INSERT INTO shifts (org_id, location_id, employee_id, role, clock_in, clock_out, hourly_rate, status) VALUES
  ('$ORG_ID', 'b2222222-2222-2222-2222-222222222222', 'ee222222-1111-1111-1111-111111111111', 'gm',
   now()::date + INTERVAL '7 hours', now()::date + INTERVAL '16 hours', 3000, 'completed');

-- David Kim (Shift Mgr) - 8am-5pm, completed
INSERT INTO shifts (org_id, location_id, employee_id, role, clock_in, clock_out, hourly_rate, status) VALUES
  ('$ORG_ID', 'b2222222-2222-2222-2222-222222222222', 'ee222222-2222-2222-2222-222222222222', 'shift_manager',
   now()::date + INTERVAL '8 hours', now()::date + INTERVAL '17 hours', 2400, 'completed');

-- David Kim (Shift Mgr) - dinner 5pm-still active
INSERT INTO shifts (org_id, location_id, employee_id, role, clock_in, clock_out, hourly_rate, status) VALUES
  ('$ORG_ID', 'b2222222-2222-2222-2222-222222222222', 'ee222222-2222-2222-2222-222222222222', 'shift_manager',
   now()::date + INTERVAL '17 hours', NULL, 2400, 'active');

-- Emily Zhao (Server) - 10am-6pm, completed
INSERT INTO shifts (org_id, location_id, employee_id, role, clock_in, clock_out, hourly_rate, status) VALUES
  ('$ORG_ID', 'b2222222-2222-2222-2222-222222222222', 'ee222222-3333-3333-3333-333333333333', 'staff',
   now()::date + INTERVAL '10 hours', now()::date + INTERVAL '18 hours', 1600, 'completed');

SQL

echo ""
echo "=== Labor seed complete! ==="
echo "Downtown Flagship: 5 employees, 8 shifts (3 still active)"
echo "Airport Terminal 4: 3 employees, 4 shifts (1 still active)"
```

- [ ] **Step 2: Make executable and run**

```bash
chmod +x scripts/seed_labor.sh
bash scripts/seed_labor.sh
```

- [ ] **Step 3: Verify data**

```bash
docker exec fireline-postgres-1 psql -U fireline -d fireline -c "SELECT e.display_name, e.role, COUNT(s.shift_id) as shifts FROM employees e LEFT JOIN shifts s ON s.employee_id = e.employee_id GROUP BY e.employee_id, e.display_name, e.role ORDER BY e.display_name"
```

Expected: 8 employees with shift counts.

- [ ] **Step 4: Commit**

```bash
git add scripts/seed_labor.sh
git commit -m "feat: add labor demo seed script with employees and shifts"
```

---

### Task 3: Backend — Labor Service

**Files:**
- Create: `internal/labor/labor.go`

- [ ] **Step 1: Create the labor service**

The service must:
- Query `employees` table for roster
- Query `shifts` table for hours/cost (cap active shifts at 16h max)
- Query `checks` for net revenue (to compute labor cost %)
- Guard against division by zero on labor_cost_pct

Key patterns to follow:
- `database.TenantTx` for all queries (see `internal/inventory/inventory.go`)
- `tenant.WithOrgID(ctx, orgID)` to set tenant context
- All money in cents (int64)
- `parseDateRange` style defaults (today)

**Types** (exact JSON tags):

```go
type LaborSummary struct {
    TotalLaborCost int64   `json:"total_labor_cost"`
    LaborCostPct   float64 `json:"labor_cost_pct"`
    NetRevenue     int64   `json:"net_revenue"`
    EmployeeCount  int     `json:"employee_count"`
    TotalHours     float64 `json:"total_hours"`
    TotalShifts    int     `json:"total_shifts"`
}

type EmployeeDetail struct {
    EmployeeID       string  `json:"employee_id"`
    DisplayName      string  `json:"display_name"`
    Role             string  `json:"role"`
    Status           string  `json:"status"`
    ShiftCount       int     `json:"shift_count"`
    HoursWorked      float64 `json:"hours_worked"`
    LaborCost        int64   `json:"labor_cost"`
    AvgHoursPerShift float64 `json:"avg_hours_per_shift"`
    HourlyRate       int64   `json:"hourly_rate"`
}
```

**Ghost shift cap:** For shifts with NULL `clock_out`, use `LEAST(now(), clock_in + INTERVAL '16 hours')` instead of raw `now()`.

**SQL for hours/cost per employee:**
```sql
SELECT e.employee_id, e.display_name, e.role, e.status,
       COUNT(s.shift_id) AS shift_count,
       COALESCE(SUM(EXTRACT(EPOCH FROM (
           COALESCE(s.clock_out, LEAST(now(), s.clock_in + INTERVAL '16 hours')) - s.clock_in
       )) / 3600.0), 0) AS hours_worked,
       COALESCE(SUM((EXTRACT(EPOCH FROM (
           COALESCE(s.clock_out, LEAST(now(), s.clock_in + INTERVAL '16 hours')) - s.clock_in
       )) / 3600.0 * s.hourly_rate)::BIGINT), 0) AS labor_cost,
       COALESCE(MAX(s.hourly_rate), 0) AS latest_rate
FROM employees e
LEFT JOIN shifts s ON s.employee_id = e.employee_id
    AND s.status != 'no_show'
    AND s.clock_in >= $2 AND s.clock_in < $3
WHERE e.location_id = $1 AND e.status = 'active'
GROUP BY e.employee_id, e.display_name, e.role, e.status
ORDER BY e.display_name
```

**SQL for net revenue** (reuse financial pattern):
```sql
SELECT COALESCE(SUM(subtotal), 0)::BIGINT AS net_revenue
FROM checks
WHERE location_id = $1 AND status = 'closed'
  AND closed_at >= $2 AND closed_at < $3
```

**Constructor:** `func New(pool *pgxpool.Pool, bus *event.Bus) *Service`

**Methods:**
- `GetSummary(ctx, orgID, locationID, from, to) (*LaborSummary, error)` — calls GetEmployees internally, aggregates
- `GetEmployees(ctx, orgID, locationID, from, to) ([]EmployeeDetail, error)` — returns per-employee detail

- [ ] **Step 2: Verify it compiles**

Run: `go build ./internal/labor/`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add internal/labor/labor.go
git commit -m "feat: add labor intelligence service with cost calculation and ghost shift cap"
```

---

### Task 4: Backend — Labor HTTP Handler + Wiring

**Files:**
- Create: `internal/api/labor_handler.go`
- Modify: `cmd/fireline/main.go`

- [ ] **Step 1: Create labor handler**

Follow the exact pattern from `internal/api/menu_handler.go`:

```go
package api

import (
    "net/http"

    "github.com/opsnerve/fireline/internal/labor"
    "github.com/opsnerve/fireline/internal/tenant"
)

type LaborHandler struct {
    svc *labor.Service
}

func NewLaborHandler(svc *labor.Service) *LaborHandler {
    return &LaborHandler{svc: svc}
}

func (h *LaborHandler) RegisterRoutes(mux *http.ServeMux, authMW func(http.Handler) http.Handler) {
    mux.Handle("GET /api/v1/labor/summary", authMW(http.HandlerFunc(h.GetSummary)))
    mux.Handle("GET /api/v1/labor/employees", authMW(http.HandlerFunc(h.GetEmployees)))
}

func (h *LaborHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
    orgID, err := tenant.OrgIDFrom(r.Context())
    if err != nil {
        WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
        return
    }
    locationID := r.URL.Query().Get("location_id")
    if locationID == "" {
        WriteError(w, http.StatusBadRequest, "LABOR_MISSING_LOCATION", "location_id is required")
        return
    }
    from, to := parseDateRange(r)
    summary, err := h.svc.GetSummary(r.Context(), orgID, locationID, from, to)
    if err != nil {
        WriteError(w, http.StatusInternalServerError, "LABOR_SUMMARY_ERROR", err.Error())
        return
    }
    WriteJSON(w, http.StatusOK, summary)
}

func (h *LaborHandler) GetEmployees(w http.ResponseWriter, r *http.Request) {
    orgID, err := tenant.OrgIDFrom(r.Context())
    if err != nil {
        WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
        return
    }
    locationID := r.URL.Query().Get("location_id")
    if locationID == "" {
        WriteError(w, http.StatusBadRequest, "LABOR_MISSING_LOCATION", "location_id is required")
        return
    }
    from, to := parseDateRange(r)
    employees, err := h.svc.GetEmployees(r.Context(), orgID, locationID, from, to)
    if err != nil {
        WriteError(w, http.StatusInternalServerError, "LABOR_EMPLOYEES_ERROR", err.Error())
        return
    }
    WriteJSON(w, http.StatusOK, map[string]any{"employees": employees})
}
```

Note: Uses existing `parseDateRange` from `handlers.go` (defaults to today).

- [ ] **Step 2: Wire into main.go**

In `cmd/fireline/main.go`:
- Add import: `"github.com/opsnerve/fireline/internal/labor"`
- After `menuSvc := menu.New(pool.Raw(), bus)` (line 96), add:
  ```go
  laborSvc := labor.New(pool.Raw(), bus)
  ```
- After `menuHandler.RegisterRoutes(mux, authMW)` (line 166), add:
  ```go
  laborHandler := api.NewLaborHandler(laborSvc)
  laborHandler.RegisterRoutes(mux, authMW)
  ```
- Update `slog.Info("all modules initialized"` to include `"labor", "ready"`

- [ ] **Step 3: Verify build**

Run: `go build -o /dev/null ./cmd/fireline`
Expected: No errors

- [ ] **Step 4: Commit**

```bash
git add internal/api/labor_handler.go cmd/fireline/main.go
git commit -m "feat: add labor intelligence HTTP handler and wire into server"
```

---

### Task 5: Frontend — API Types + Client + Hooks

**Files:**
- Modify: `web/src/lib/api.ts`
- Create: `web/src/hooks/useLabor.ts`

- [ ] **Step 1: Add types and laborApi to api.ts**

Add before `export { ApiError }`:

```typescript
// Labor Intelligence
export interface LaborSummary {
  total_labor_cost: number;
  labor_cost_pct: number;
  net_revenue: number;
  employee_count: number;
  total_hours: number;
  total_shifts: number;
}

export interface EmployeeDetail {
  employee_id: string;
  display_name: string;
  role: string;
  status: string;
  shift_count: number;
  hours_worked: number;
  labor_cost: number;
  avg_hours_per_shift: number;
  hourly_rate: number;
}

export const laborApi = {
  getSummary(locationId: string, from?: string, to?: string) {
    const params = new URLSearchParams({ location_id: locationId });
    if (from) params.set('from', from);
    if (to) params.set('to', to);
    return request<LaborSummary>(`/labor/summary?${params}`);
  },
  getEmployees(locationId: string, from?: string, to?: string) {
    const params = new URLSearchParams({ location_id: locationId });
    if (from) params.set('from', from);
    if (to) params.set('to', to);
    return request<{ employees: EmployeeDetail[] }>(`/labor/employees?${params}`);
  },
};
```

- [ ] **Step 2: Create useLabor.ts**

```typescript
import { useQuery } from '@tanstack/react-query';
import { laborApi } from '../lib/api';

export function useLaborSummary(locationId: string | null, from?: string, to?: string) {
  return useQuery({
    queryKey: ['labor', 'summary', locationId, from, to],
    queryFn: () => laborApi.getSummary(locationId!, from, to),
    enabled: !!locationId,
    staleTime: 30_000,
    refetchInterval: 30_000,
  });
}

export function useLaborEmployees(locationId: string | null, from?: string, to?: string) {
  return useQuery({
    queryKey: ['labor', 'employees', locationId, from, to],
    queryFn: () => laborApi.getEmployees(locationId!, from, to),
    enabled: !!locationId,
    staleTime: 30_000,
  });
}
```

- [ ] **Step 3: Verify compiles**

Run: `cd web && npx tsc --noEmit`

- [ ] **Step 4: Commit**

```bash
git add web/src/lib/api.ts web/src/hooks/useLabor.ts
git commit -m "feat: add labor intelligence API types, client, and hooks"
```

---

### Task 6: Frontend — LaborPage + Routing + Nav

**Files:**
- Create: `web/src/pages/LaborPage.tsx`
- Modify: `web/src/App.tsx`
- Modify: `web/src/components/Layout.tsx`

- [ ] **Step 1: Create LaborPage.tsx**

Key requirements:
- 4 KPI cards: Labor Cost ($), Labor Cost %, Active Employees, Total Hours
- Employee DataTable with columns: Employee, Role, Status (StatusBadge), Shifts, Hours, Cost ($), Avg Hrs/Shift, Rate ($/hr)
- Status badge: active→success, inactive→neutral, terminated→critical
- Role labels: capitalize first letter
- `verbatimModuleSyntax: true` — use `import type` for type-only imports
- Icons: `DollarSign`, `Percent`, `Users`, `Clock` from lucide-react
- `cents()` helper for money, `hours.toFixed(1)` for time
- Loading/Error/Empty states using existing components

- [ ] **Step 2: Add route in App.tsx**

Import `LaborPage` and add route inside Layout group after `menu`:
```tsx
<Route path="labor" element={<LaborPage />} />
```

- [ ] **Step 3: Add nav item in Layout.tsx**

Add `Users` to the lucide-react import (it's different from the existing `User` singular import).

Add to `navItems` array after Menu:
```typescript
{ to: '/labor', label: 'Labor', icon: Users },
```

- [ ] **Step 4: Verify compiles**

Run: `cd web && npx tsc --noEmit`

- [ ] **Step 5: Commit**

```bash
git add web/src/pages/LaborPage.tsx web/src/App.tsx web/src/components/Layout.tsx
git commit -m "feat: add Labor page with KPI cards and employee table"
```

---

### Task 7: Full Build + Smoke Test

**Files:** None (verification only)

- [ ] **Step 1: TypeScript check**

Run: `cd web && npx tsc --noEmit`

- [ ] **Step 2: Go build**

Run: `cd /Users/bistrocloud/Documents/AI_Restaurant_System/fireline && go build -o /dev/null ./cmd/fireline`

- [ ] **Step 3: Frontend production build**

Run: `cd web && npm run build`

- [ ] **Step 4: Restart server and test labor endpoints**

```bash
pkill -f './fireline'; sleep 1
go build -o fireline ./cmd/fireline && ./fireline &
sleep 2
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login -H "Content-Type: application/json" -d '{"email":"owner@bistrocloud.com","password":"DemoPassword1234!"}' | python3 -c "import sys,json; print(json.load(sys.stdin)['access_token'])")
echo "=== Labor Summary ==="
curl -s -H "Authorization: Bearer $TOKEN" "http://localhost:8080/api/v1/labor/summary?location_id=a1111111-1111-1111-1111-111111111111" | python3 -m json.tool
echo ""
echo "=== Employees ==="
curl -s -H "Authorization: Bearer $TOKEN" "http://localhost:8080/api/v1/labor/employees?location_id=a1111111-1111-1111-1111-111111111111" | python3 -c "import sys,json; [print(f'  {e[\"display_name\"]:20s} {e[\"role\"]:15s} shifts={e[\"shift_count\"]} hrs={e[\"hours_worked\"]:.1f} cost=\${e[\"labor_cost\"]/100:.2f}') for e in json.load(sys.stdin)['employees']]"
```

Expected: Labor summary with cost, percentage, hours. Employee list with shift details.

- [ ] **Step 5: Verify frontend loads**

Restart dev server if needed, open http://localhost:3000/labor
