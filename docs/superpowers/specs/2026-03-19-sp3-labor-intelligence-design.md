# SP3: Labor & Workforce Intelligence Design Spec

**Date:** 2026-03-19
**Status:** Approved
**Scope:** Employee roster, shift tracking, labor cost dashboard with labor cost % of revenue

---

## 1. Database — New Shifts Table

New migration `004_shifts.sql`:

```sql
CREATE TABLE shifts (
    shift_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(org_id),
    location_id UUID NOT NULL REFERENCES locations(location_id),
    employee_id UUID NOT NULL REFERENCES employees(employee_id),
    role TEXT NOT NULL DEFAULT 'staff',
    clock_in TIMESTAMPTZ NOT NULL,
    clock_out TIMESTAMPTZ,
    hourly_rate INT NOT NULL DEFAULT 0 CHECK (hourly_rate >= 0),  -- cents per hour
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

## 2. Backend — Labor Intelligence Service

New Go service at `internal/labor/` following existing patterns.

### Data Sources

- `employees` — roster per location (existing table: employee_id, org_id, location_id, display_name, role, status)
- `shifts` — clock in/out, hourly rate, status (new table)
- `checks` — net revenue for labor cost % calculation (existing, via financial service pattern)

### Calculations

**1. Labor Cost**
```sql
SELECT SUM(
    CASE WHEN s.clock_out IS NOT NULL
        THEN (EXTRACT(EPOCH FROM (s.clock_out - s.clock_in)) / 3600.0 * s.hourly_rate)::BIGINT
        ELSE (EXTRACT(EPOCH FROM (now() - s.clock_in)) / 3600.0 * s.hourly_rate)::BIGINT
    END
) AS total_labor_cost
FROM shifts s
WHERE s.location_id = $1 AND s.status != 'no_show'
  AND s.clock_in >= $2 AND s.clock_in < $3
```
Active shifts (no clock_out) use `now()` for running cost. Result in cents.

**2. Labor Cost %**
```
labor_cost_pct = net_revenue > 0 ? (total_labor_cost / net_revenue) * 100 : 0.0
```
Net revenue from `checks` table (same query as financial P&L, sum of `subtotal` for closed checks in the period). **Guard against zero revenue** — return 0.0% when no revenue exists (e.g., early morning before orders).

**Ghost shift cap:** For shifts with no `clock_out`, cap the running duration at 16 hours. If `now() - clock_in > 16h`, use 16h instead — this prevents forgotten clock-outs from inflating historical queries.

**3. Hours Worked**
Per employee: `SUM(EXTRACT(EPOCH FROM (COALESCE(clock_out, now()) - clock_in)) / 3600.0)`

**4. Employee Summary**
Per employee for the period: display_name, role, status, shift_count, hours_worked, labor_cost.

### Types

```go
type LaborSummary struct {
    TotalLaborCost   int64   `json:"total_labor_cost"`    // cents
    LaborCostPct     float64 `json:"labor_cost_pct"`      // percentage of net revenue
    NetRevenue       int64   `json:"net_revenue"`         // cents (for context)
    EmployeeCount    int     `json:"employee_count"`      // active employees at location
    TotalHours       float64 `json:"total_hours"`         // total hours worked
    TotalShifts      int     `json:"total_shifts"`
}

type EmployeeDetail struct {
    EmployeeID  string  `json:"employee_id"`
    DisplayName string  `json:"display_name"`
    Role        string  `json:"role"`
    Status      string  `json:"status"`        // active/inactive/terminated
    ShiftCount  int     `json:"shift_count"`
    HoursWorked float64 `json:"hours_worked"`
    LaborCost   int64   `json:"labor_cost"`    // cents
    AvgHoursPerShift float64 `json:"avg_hours_per_shift"`
    HourlyRate  int64   `json:"hourly_rate"`   // cents (latest)
}
```

### API Endpoints

**`GET /api/v1/labor/summary?location_id=X`**
- Requires JWT auth + location_id
- Optional: `from`, `to` (defaults to today)
- Returns: `LaborSummary`

**`GET /api/v1/labor/employees?location_id=X`**
- Same auth/params
- Optional: `from`, `to` (defaults to today)
- Returns: `{ employees: EmployeeDetail[] }`

### Service Constructor & Methods

```go
func New(pool *pgxpool.Pool, bus *event.Bus) *Service

func (s *Service) GetSummary(ctx context.Context, orgID, locationID string, from, to time.Time) (*LaborSummary, error)
func (s *Service) GetEmployees(ctx context.Context, orgID, locationID string, from, to time.Time) ([]EmployeeDetail, error)
```

### Files

- `internal/labor/labor.go` — Service with types, GetSummary, GetEmployees
- `internal/api/labor_handler.go` — HTTP handlers (LaborHandler with RegisterRoutes)
- `cmd/fireline/main.go` — Create labor service, register routes

## 3. Frontend — Labor & Workforce Page

### New Files

- `web/src/pages/LaborPage.tsx`
- `web/src/hooks/useLabor.ts`
- Modify: `web/src/lib/api.ts` — Add laborApi + types
- Modify: `web/src/App.tsx` — Add `/labor` route
- Modify: `web/src/components/Layout.tsx` — Add Labor nav item

### TypeScript Types (in api.ts)

```typescript
export interface LaborSummary {
  total_labor_cost: number;   // cents
  labor_cost_pct: number;
  net_revenue: number;        // cents
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
  labor_cost: number;         // cents
  avg_hours_per_shift: number;
  hourly_rate: number;        // cents
}
```

### API Client

```typescript
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

### React Query Hooks (useLabor.ts)

```typescript
export function useLaborSummary(locationId: string | null, from?: string, to?: string)
  // staleTime 30s, refetchInterval 30s (running cost updates for active shifts)
export function useLaborEmployees(locationId: string | null, from?: string, to?: string)
  // staleTime 30s
```

### Page Layout (LaborPage.tsx)

**Row 1 — KPI Cards (4 cards):**
- Labor Cost ($) — cents→$ with DollarSign icon, red tint
- Labor Cost % — percentage with Percent icon, blue tint (industry target: 25-35%)
- Active Employees — count with Users icon, gray tint
- Total Hours — with Clock icon, purple tint

Data from `useLaborSummary`.

**Row 2 — Employee DataTable:**
Columns:
| Column | Key | Sortable | Align | Render |
|--------|-----|----------|-------|--------|
| Employee | display_name | yes | left | bold text |
| Role | role | yes | left | capitalize |
| Status | status | yes | center | StatusBadge (active→success, inactive→neutral, terminated→critical) |
| Shifts | shift_count | yes | right | — |
| Hours | hours_worked | yes | right | X.X |
| Cost ($) | labor_cost | yes | right | cents→$ |
| Avg Hrs/Shift | avg_hours_per_shift | yes | right | X.X |
| Rate ($/hr) | hourly_rate | yes | right | cents→$ |

### Navigation

Add to `Layout.tsx` navItems after Menu:
```typescript
{ to: '/labor', label: 'Labor', icon: Users }
```

Add route in `App.tsx`:
```tsx
<Route path="labor" element={<LaborPage />} />
```

## 4. Migration + Demo Seed Data

### Migration

New file `migrations/004_shifts.sql` with the shifts table, indexes, RLS policy, and grants (as shown in Section 1).

Update `atlas.sum` hash after creating the migration.

### Demo Seed Extension

Extend `scripts/seed_demo.sh` (or create `scripts/seed_labor.sh`) to insert:

**Employees (8 total across both locations):**

Downtown Flagship (5 employees):
- Alex Rivera (owner, $0/hr — salaried)
- Maria Santos (gm, $2800/hr = $28.00)
- Jake Thompson (shift_manager, $2200/hr = $22.00)
- Sarah Chen (staff - line cook, $1800/hr = $18.00)
- Marcus Brown (staff - server, $1500/hr = $15.00)

Airport Terminal 4 (3 employees):
- Priya Patel (gm, $3000/hr = $30.00)
- David Kim (shift_manager, $2400/hr = $24.00)
- Emily Zhao (staff - server, $1600/hr = $16.00)

**Shifts (today, realistic schedule):**
- Morning prep: 6 AM - 2 PM (BOH staff)
- Lunch: 10 AM - 4 PM (FOH + BOH)
- Dinner: 3 PM - 11 PM (FOH + BOH)
- Manager: 8 AM - 5 PM

Each employee gets 1-2 shifts for today, some completed (clock_out set), some still active (clock_out NULL for current shift).

## 5. Conventions

- Backend follows existing patterns: TenantTx for all queries, handler extracts orgID from tenant context + locationID from query param
- Reuse existing `parseDateRange` from `handlers.go` (defaults to today — appropriate for labor)
- `event.Bus` accepted in constructor for future extensibility (no subscriptions in SP3)
- Default date range: today (labor cost is most relevant for current day)
- Employee hourly_rate in `EmployeeDetail` comes from their most recent shift record (employees table has no rate column)
- Run `atlas migrate hash --dir file://migrations` after creating 004_shifts.sql
- All financial values in cents
- Active shifts (no clock_out) calculate running cost using `now()`
- Frontend follows SP1/SP2 patterns: hooks with locationId guard, cents→$ helper, loading/error/empty states
- `Users` icon from lucide-react for nav item
