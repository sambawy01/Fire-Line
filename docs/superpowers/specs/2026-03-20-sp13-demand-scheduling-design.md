# SP13: Labor Demand-Based Scheduling Engine

**Date:** 2026-03-20
**Status:** Approved
**Scope:** Labor demand forecast, schedule generation with constraint satisfaction, schedule management UI, shift swap workflow, overtime alerts
**Maps to:** Build Plan Sprint 23 (Labor — Demand-Based Scheduling Engine)

---

## 1. Database — Migration 010

New migration: `migrations/010_scheduling.sql`

### New Tables

```sql
-- Generated schedules
CREATE TABLE schedules (
    schedule_id    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID NOT NULL REFERENCES organizations(org_id),
    location_id    UUID NOT NULL REFERENCES locations(location_id),
    week_start     DATE NOT NULL,
    status         TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'published', 'archived')),
    created_by     UUID REFERENCES users(user_id),
    published_at   TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(org_id, location_id, week_start)
);

-- Individual scheduled shifts within a schedule
CREATE TABLE scheduled_shifts (
    scheduled_shift_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID NOT NULL REFERENCES organizations(org_id),
    schedule_id    UUID NOT NULL REFERENCES schedules(schedule_id),
    employee_id    UUID NOT NULL REFERENCES employees(employee_id),
    shift_date     DATE NOT NULL,
    start_time     TIME NOT NULL,
    end_time       TIME NOT NULL,
    station        TEXT,
    status         TEXT NOT NULL DEFAULT 'scheduled' CHECK (status IN ('scheduled', 'confirmed', 'swap_requested', 'swapped', 'cancelled')),
    notes          TEXT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Shift swap requests
CREATE TABLE shift_swap_requests (
    swap_id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID NOT NULL REFERENCES organizations(org_id),
    requester_shift_id UUID NOT NULL REFERENCES scheduled_shifts(scheduled_shift_id),
    target_shift_id    UUID REFERENCES scheduled_shifts(scheduled_shift_id),
    target_employee_id UUID REFERENCES employees(employee_id),
    status         TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'denied', 'cancelled')),
    reason         TEXT,
    reviewed_by    UUID REFERENCES users(user_id),
    reviewed_at    TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Labor demand forecast per 30-min block
CREATE TABLE labor_demand_forecast (
    forecast_id    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID NOT NULL REFERENCES organizations(org_id),
    location_id    UUID NOT NULL REFERENCES locations(location_id),
    forecast_date  DATE NOT NULL,
    time_block     TIME NOT NULL,
    forecasted_covers  INT NOT NULL DEFAULT 0,
    required_elu       NUMERIC(8,2) NOT NULL DEFAULT 0,
    required_headcount INT NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(org_id, location_id, forecast_date, time_block)
);
```

### RLS + Indexes

Standard org_isolation on all 4 tables. Indexes on schedule_id, employee_id, forecast_date.

---

## 2. Labor Demand Forecast Engine

New file: `internal/labor/forecast.go`

### Tier 0 Forecast (trailing average)

For each 30-minute block of a target day:
1. Query average covers (check count) for the same day-of-week over the last 4 weeks
2. Apply staffing ratio: `required_headcount = ceil(forecasted_covers / covers_per_staff)`
3. Compute required ELU: `required_headcount * 1.0` (assumes fully-trained staff)

Default covers_per_staff = 15 (configurable per location).

### Methods on `*Service`

- `GenerateForecast(ctx, orgID, locationID, targetDate)` — compute and persist forecast for a full day in 30-min blocks
- `GetForecast(ctx, orgID, locationID, date)` — retrieve forecast blocks for a day

---

## 3. Schedule Generation Engine

New file: `internal/labor/scheduler.go`

### Constraint Solver (greedy heuristic)

For each day in the week:
1. Load demand forecast (required headcount per 30-min block)
2. Load available employees (from availability JSON + not already scheduled)
3. Load ELU ratings per employee
4. For each time block needing coverage:
   - Sort available employees by: ELU score for needed station (desc), then fairness score (fewer hours this week = higher priority)
   - Assign top candidate to a shift spanning contiguous blocks
   - Respect constraints:
     - Max 8 hours per shift (configurable)
     - Max 40 hours per week
     - At least 8 hours between shifts
     - Employee availability window
5. Return generated schedule

### Methods on `*Service`

- `GenerateScheduleDraft(ctx, orgID, locationID, weekStart)` — create schedule + scheduled shifts
- `GetSchedule(ctx, orgID, locationID, weekStart)` — get schedule with all shifts
- `UpdateSchedule(ctx, orgID, scheduleID, shifts)` — manager edits shifts
- `PublishSchedule(ctx, orgID, scheduleID)` — set status to published, emit event
- `GetEmployeeSchedule(ctx, orgID, employeeID, weekStart)` — shifts for one employee

### Labor Cost Projection

- `ProjectLaborCost(ctx, orgID, scheduleID)` — sum(shift_hours * hourly_rate) for all shifts, compute labor cost % against forecasted revenue

---

## 4. Shift Swap Workflow

New file: `internal/labor/swaps.go`

- `RequestSwap(ctx, orgID, requesterShiftID, targetEmployeeID, reason)` — create swap request
- `ReviewSwap(ctx, orgID, swapID, approved bool, reviewedBy)` — approve/deny, if approved: update both shifts
- `ListSwapRequests(ctx, orgID, locationID, status)` — list pending/all swaps

---

## 5. Overtime Detection

In the scheduler or as a separate check:
- `CheckOvertimeRisk(ctx, orgID, locationID, weekStart)` — for each employee, check if scheduled hours > 38 (warning) or > 40 (critical)
- Emit `labor.overtime.risk` alert

---

## 6. API Endpoints

```
POST   /api/v1/labor/forecast            — Generate forecast for a date
GET    /api/v1/labor/forecast             — Get forecast (query: location_id, date)
POST   /api/v1/labor/schedules/generate   — Generate draft schedule for a week
GET    /api/v1/labor/schedules            — Get schedule (query: location_id, week_start)
PUT    /api/v1/labor/schedules/{id}       — Update schedule shifts
POST   /api/v1/labor/schedules/{id}/publish — Publish schedule
GET    /api/v1/labor/schedules/employee/{id} — Employee's schedule
GET    /api/v1/labor/schedules/{id}/cost   — Labor cost projection
POST   /api/v1/labor/swaps               — Request shift swap
GET    /api/v1/labor/swaps               — List swap requests (query: location_id, status)
PUT    /api/v1/labor/swaps/{id}          — Approve/deny swap
GET    /api/v1/labor/overtime-risk        — Check overtime risks (query: location_id, week_start)
```

---

## 7. Web Dashboard — Scheduling Page

New page: `web/src/pages/SchedulingPage.tsx` (new route `/scheduling`, new nav item after Labor)

### Layout

**Visual Weekly Grid** — 7 columns (Mon-Sun), rows = employees. Each cell shows shift time + station. Color-coded by station.

**Top bar:** week selector (prev/next), "Generate Draft" button, "Publish" button

**Sections:**
1. **Schedule Grid** — the main visual
2. **Demand Overlay** — toggle to show forecasted demand per time block as a heatmap behind the grid
3. **Labor Cost** — projected cost %, compared to budget target
4. **Swap Requests** — pending swaps with approve/deny buttons
5. **Overtime Warnings** — employees approaching 40hr limit

---

## 8. RBAC

- `labor:schedule` — create/edit/publish schedules (roles: `shift_manager`, `gm`, `owner`)
- `labor:swap` — request swaps (roles: `staff`, `shift_manager`, `gm`, `owner`)

---

## 9. Testing

- Forecast: verify trailing average produces expected covers
- Scheduler: verify constraints (max hours, availability, between-shift gap)
- Overtime: verify warning at 38hr, critical at 40hr
- Swap: request → approve → shifts updated
