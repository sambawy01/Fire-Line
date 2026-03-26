# Staff App & Intelligence Layer — Design Spec

## Overview

Standalone mobile-first React app (`web-staff/`) for frontline restaurant staff. PIN-based auth, task management with data entry, shift scheduling, points/gamification, and a surveillance intelligence layer for anomaly detection visible to GM, Ops Director, and CEO.

## Role Hierarchy

| Role | Level | Staff App | Management Dashboard |
|---|---|---|---|
| staff | 1 | Own tasks, schedule, points | No access |
| shift_manager | 2 | + Team performance, task completion | No access |
| gm | 3 | Staff app overview + performance metrics | Full location management, anomaly flags |
| ops_director | 4 | Same staff app access as GM | Multi-location anomaly investigation, loss prevention |
| owner (CEO) | 5 | No access | Cross-location benchmarking, fraud dashboard, predictive alerts |

## Staff App — `web-staff/`

Separate Vite + React + Tailwind project. Shares the Go backend API. Deployed independently to Vercel. Dark theme, bottom tab navigation.

### PIN Login

Staff enters 4-6 digit PIN → `POST /api/v1/auth/pin-login` → JWT with user_id, org_id, role, location_id. Role determines tab visibility.

### Tabs

1. **Home** — Command center: shift status, active task count + progress, announcements, today's points, pending swaps. Clock in/out button.
2. **Tasks** — Active checklists (opening/closing/mid-shift/prep), ad-hoc tasks, data entry tasks. Three types:
   - Checklist item (checkbox)
   - Ad-hoc (one-off assignment)
   - Data entry (temperature log, waste count, cash drop — with expected range, unit, photo)
3. **Schedule** — Weekly view of my shifts, swap request flow, break timer.
4. **Points** — Balance, history feed, monthly trend, location leaderboard.
5. **Team** (shift_manager+) — Who's clocked in, per-person task completion, KDS speed.
6. **Performance** (gm+) — Daily/weekly task completion trends, employee rankings, shift comparisons.

### Clock In/Out

- PIN re-confirmation on clock-in
- Auto-assigns checklist templates on clock-in based on role/station
- Break timer with configurable duration
- Soft warning on clock-out if urgent tasks incomplete

### Task Model

Fields: title, description, type (checklist_item/ad_hoc/data_entry), assigned_to, assigned_by, priority (low/normal/urgent), due_at, status (pending/in_progress/completed/skipped), data_entry_config (JSONB: expected_range, unit), data_entry_value (JSONB), photo_url, completed_at, completed_by, org_id, location_id, template_id (nullable).

### Checklist Template Example — "Opening Duties"

- Check walk-in temperature (data entry: °C, expected 1-4)
- Prep station sanitized (checkbox)
- Fryer oil level checked (checkbox)
- Cash drawer counted (data entry: EGP amount)
- Floor swept and mopped (checkbox + optional photo)

## Intelligence Layer

Computed server-side from existing data (checks, shifts, inventory counts, KDS tickets, staff points). No cameras — purely data-driven behavioral analysis.

### Anomaly Types

1. **Void/discount anomalies** — Employee voids/discounts significantly above peer average (z-score)
2. **Cash variance** — Register shortages correlated with employee/shift
3. **Clock irregularities** — Buddy punching signals (clocked in but no POS activity for 30+ min)
4. **Inventory shrinkage** — Shrinkage spikes during specific shifts
5. **Suspicious transaction patterns** — Frequent small voids, rounding patterns

### Visibility

- **Shift Manager**: Team performance + task completion only. No anomaly data.
- **GM**: Location anomaly flags in Alerts page. Can view but not resolve.
- **Ops Director**: All anomaly flags across ALL locations. Investigation view (employee timeline: shifts, POS activity, voids, inventory). Resolution workflow (confirmed/false-positive, assign follow-up, notes). Loss prevention scorecard.
- **CEO**: Fraud/loss summary in Portfolio briefing. Cross-location staff intelligence. Predictive alerts.

### Investigation Workflow (Ops Director)

Click anomaly → full employee activity timeline for the period → resolution: confirm incident / mark false positive / assign follow-up to GM → add investigation notes → close.

## New Backend Components

### Database (1 migration)

```sql
-- Task templates (reusable checklists)
CREATE TABLE task_templates (
    template_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(org_id),
    location_id UUID REFERENCES locations(location_id),
    name TEXT NOT NULL,
    description TEXT,
    trigger_type TEXT NOT NULL DEFAULT 'manual', -- manual, shift_start, shift_end, scheduled
    target_role TEXT, -- which role gets this template
    target_station TEXT, -- which station
    items JSONB NOT NULL, -- [{title, type, data_entry_config}]
    active BOOLEAN DEFAULT true,
    created_by UUID REFERENCES users(user_id),
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

-- Task instances
CREATE TABLE tasks (
    task_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(org_id),
    location_id UUID NOT NULL REFERENCES locations(location_id),
    template_id UUID REFERENCES task_templates(template_id),
    title TEXT NOT NULL,
    description TEXT,
    type TEXT NOT NULL DEFAULT 'checklist_item', -- checklist_item, ad_hoc, data_entry
    assigned_to UUID REFERENCES employees(employee_id),
    assigned_by UUID REFERENCES users(user_id),
    priority TEXT NOT NULL DEFAULT 'normal', -- low, normal, urgent
    due_at TIMESTAMPTZ,
    status TEXT NOT NULL DEFAULT 'pending', -- pending, in_progress, completed, skipped
    data_entry_config JSONB, -- {expected_min, expected_max, unit}
    data_entry_value JSONB, -- {value, unit, within_range}
    photo_url TEXT,
    completed_at TIMESTAMPTZ,
    completed_by UUID REFERENCES employees(employee_id),
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

-- Announcements
CREATE TABLE announcements (
    announcement_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(org_id),
    location_id UUID NOT NULL REFERENCES locations(location_id),
    title TEXT NOT NULL,
    body TEXT,
    priority TEXT NOT NULL DEFAULT 'normal',
    created_by UUID REFERENCES users(user_id),
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT now()
);

-- Intelligence anomalies
CREATE TABLE anomalies (
    anomaly_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(org_id),
    location_id UUID NOT NULL REFERENCES locations(location_id),
    employee_id UUID REFERENCES employees(employee_id),
    type TEXT NOT NULL, -- void_pattern, cash_variance, clock_irregularity, shrinkage, transaction_pattern
    severity TEXT NOT NULL DEFAULT 'warning', -- info, warning, critical
    title TEXT NOT NULL,
    description TEXT,
    evidence JSONB, -- supporting data points
    status TEXT NOT NULL DEFAULT 'open', -- open, investigating, confirmed, false_positive, resolved
    assigned_to UUID REFERENCES users(user_id),
    resolved_by UUID REFERENCES users(user_id),
    resolved_at TIMESTAMPTZ,
    resolution_notes TEXT,
    detected_at TIMESTAMPTZ DEFAULT now(),
    created_at TIMESTAMPTZ DEFAULT now()
);
```

### RBAC Update

Add `ops_director` role (level 4) to `internal/auth/rbac.go`:
- All GM permissions
- `intelligence:anomalies` — view anomaly feed
- `intelligence:investigate` — view employee timelines
- `intelligence:resolve` — resolve/close anomalies
- Multi-location read access (not restricted to one location_id)

### New API Endpoints

**Tasks:**
- `POST /api/v1/tasks` — Create task (shift_manager+)
- `GET /api/v1/tasks/my` — My active tasks (staff+, filtered by employee_id)
- `GET /api/v1/tasks` — All tasks for location (shift_manager+)
- `PUT /api/v1/tasks/{id}/status` — Update status (staff+)
- `PUT /api/v1/tasks/{id}/complete` — Complete with data entry/photo (staff+)

**Templates:**
- `POST /api/v1/task-templates` — Create template (gm+)
- `GET /api/v1/task-templates` — List templates (shift_manager+)
- `POST /api/v1/task-templates/{id}/instantiate` — Create tasks from template (shift_manager+)

**Announcements:**
- `POST /api/v1/announcements` — Create (shift_manager+)
- `GET /api/v1/announcements` — List active for location (staff+)

**Intelligence:**
- `GET /api/v1/intelligence/anomalies` — Anomaly feed (gm+)
- `GET /api/v1/intelligence/anomalies/{id}` — Detail with evidence (ops_director+)
- `PUT /api/v1/intelligence/anomalies/{id}/resolve` — Resolve (ops_director+)
- `GET /api/v1/intelligence/investigation/{employee_id}` — Full activity timeline (ops_director+)

### Intelligence Package — `internal/intelligence/`

Subscribes to event bus. Runs anomaly detection on:
- `pipeline.orders.processed` — void/discount analysis per employee
- `shift.completed` — clock pattern analysis
- `inventory.counted` — shrinkage correlation

Uses z-score thresholds (same pattern as existing financial/inventory intelligence). Writes to `anomalies` table and publishes alerts.

## Tech Stack

- **web-staff/**: Vite + React 19 + TypeScript + Tailwind CSS 4
- **Backend**: Go, same `cmd/fireline/` binary, new handlers registered in main.go
- **Database**: PostgreSQL 16 with RLS, new migration 019
- **Deploy**: Vercel (staff app), Railway auto-deploys backend
