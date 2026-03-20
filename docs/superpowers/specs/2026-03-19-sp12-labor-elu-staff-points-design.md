# SP12: Labor ELU Ratings & Staff Point System

**Date:** 2026-03-19
**Status:** Approved
**Scope:** ELU rating model, staff point system, employee profile extensions, manager staff oversight, tablet point display
**Maps to:** Build Plan Sprint 22 (Labor — Employee Profiles, ELU & Staff Point System)

---

## 1. Database — Migration 009

New migration: `migrations/009_elu_staff_points.sql`

### Schema Changes

```sql
-- ELU ratings per employee per station
ALTER TABLE employees
    ADD COLUMN elu_ratings JSONB NOT NULL DEFAULT '{}',
    ADD COLUMN staff_points NUMERIC(10,2) NOT NULL DEFAULT 0,
    ADD COLUMN certifications TEXT[] NOT NULL DEFAULT '{}',
    ADD COLUMN availability JSONB NOT NULL DEFAULT '{}';

-- Point history for audit trail
CREATE TABLE staff_point_events (
    event_id      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id        UUID NOT NULL REFERENCES organizations(org_id),
    employee_id   UUID NOT NULL REFERENCES employees(employee_id),
    points        NUMERIC(10,2) NOT NULL,
    reason        TEXT NOT NULL CHECK (reason IN (
        'task_completion', 'speed_bonus', 'accuracy_bonus',
        'attendance', 'peer_nominated', 'late', 'no_show',
        'incomplete_task', 'manager_adjustment'
    )),
    description   TEXT,
    shift_id      UUID REFERENCES shifts(shift_id),
    awarded_by    UUID REFERENCES users(user_id),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_point_events_employee ON staff_point_events(org_id, employee_id, created_at DESC);
CREATE INDEX idx_point_events_shift ON staff_point_events(shift_id);

ALTER TABLE staff_point_events ENABLE ROW LEVEL SECURITY;
ALTER TABLE staff_point_events FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON staff_point_events
    USING (org_id = current_setting('app.current_org_id')::UUID);
GRANT SELECT, INSERT, UPDATE, DELETE ON staff_point_events TO fireline_app;
```

### ELU Ratings JSON Structure

```json
{
  "grill": 1.5,
  "fryer": 1.0,
  "prep": 0.8,
  "expo": 1.2,
  "register": 1.0,
  "dish": 0.5
}
```

Scale: 0.0 (untrained) → 1.0 (fully trained) → 2.0 (expert/trainer)

### Availability JSON Structure

```json
{
  "monday": {"start": "06:00", "end": "14:00"},
  "tuesday": {"start": "06:00", "end": "14:00"},
  "wednesday": null,
  "thursday": {"start": "14:00", "end": "22:00"},
  "friday": {"start": "14:00", "end": "22:00"},
  "saturday": {"start": "10:00", "end": "18:00"},
  "sunday": null
}
```

---

## 2. ELU Rating Service

New file: `internal/labor/elu.go`

### Types

```go
type ELURating struct {
    Station string  `json:"station"`
    Score   float64 `json:"score"`   // 0.0-2.0
}

type EmployeeProfile struct {
    EmployeeID     string            `json:"employee_id"`
    DisplayName    string            `json:"display_name"`
    Role           string            `json:"role"`
    Status         string            `json:"status"`
    ELURatings     map[string]float64 `json:"elu_ratings"`
    StaffPoints    float64           `json:"staff_points"`
    PointsTrend    string            `json:"points_trend"`   // "up", "down", "stable"
    Certifications []string          `json:"certifications"`
    Availability   map[string]any    `json:"availability"`
}
```

### Methods on `*Service`

- `GetEmployeeProfile(ctx, orgID, employeeID)` — full profile with ELU + points + trend
- `ListEmployeeProfiles(ctx, orgID, locationID)` — all employees with profiles (for manager view)
- `UpdateELURatings(ctx, orgID, employeeID, ratings map[string]float64)` — manager sets station ratings
- `UpdateAvailability(ctx, orgID, employeeID, availability)` — employee/manager sets availability
- `UpdateCertifications(ctx, orgID, employeeID, certs []string)` — manager sets certifications

### Points Trend Calculation

Compare current points to 7 days ago:
- If current > 7_days_ago + 5: "up"
- If current < 7_days_ago - 5: "down"
- Else: "stable"

---

## 3. Staff Point System

New file: `internal/labor/points.go`

### Point Rules

| Event | Points | Direction |
|-------|--------|-----------|
| task_completion | +5 | earn |
| speed_bonus | +3 | earn |
| accuracy_bonus | +3 | earn |
| attendance (on-time clock-in) | +2 | earn |
| peer_nominated | +10 | earn |
| late (clock-in > 5 min after shift start) | -3 | deduct |
| no_show | -10 | deduct |
| incomplete_task | -2 | deduct |
| manager_adjustment | variable | either |

### Methods on `*Service`

- `AwardPoints(ctx, orgID, employeeID, points, reason, description, shiftID, awardedBy)` — INSERT event + UPDATE employee.staff_points
- `GetPointHistory(ctx, orgID, employeeID, limit)` — recent point events
- `GetLeaderboard(ctx, orgID, locationID, limit)` — top employees by points
- `RecalculatePoints(ctx, orgID, employeeID)` — sum all events, update staff_points (for consistency)

---

## 4. API Endpoints

### Employee Profiles

```
GET    /api/v1/labor/profiles              — List profiles for location (query: location_id)
GET    /api/v1/labor/profiles/{id}         — Get single profile
PUT    /api/v1/labor/profiles/{id}/elu     — Update ELU ratings (body: {ratings: {station: score}})
PUT    /api/v1/labor/profiles/{id}/availability — Update availability
PUT    /api/v1/labor/profiles/{id}/certifications — Update certifications
```

### Staff Points

```
POST   /api/v1/labor/points               — Award/deduct points (body: {employee_id, points, reason, description})
GET    /api/v1/labor/points/{employee_id}  — Point history
GET    /api/v1/labor/leaderboard           — Top performers (query: location_id, limit)
```

---

## 5. Web Dashboard — Enhanced Labor Page

Extend existing `LaborPage.tsx` with tabs:

### Tab 1: Overview (existing)
- Existing KPI cards + employee table

### Tab 2: Staff Profiles
- DataTable: name, role, ELU summary (avg score), points, trend arrow (↑↓→), certifications count
- Click row → expand showing:
  - ELU ratings bar chart (one bar per station, color-coded by score)
  - Point history timeline
  - Availability grid

### Tab 3: Leaderboard
- Ranked list of employees by points
- Profile cards: rank #, name, points, trend, top station

### Tab 4: ELU Management
- Manager selects employee
- Station rating sliders (0.0-2.0 in 0.1 increments)
- Save button
- Visual guide: 0.0-0.5 red (training), 0.6-1.0 yellow (competent), 1.1-2.0 green (expert)

---

## 6. Tablet — Staff Point Display

On PIN login success, show the staff member's:
- Current point score (large number)
- Trend arrow (up/down/stable) with color
- Last 3 point events (reason + points)
- Display for 3 seconds, then proceed to tabs

Modify `tablet/app/(auth)/pin.tsx` to show point summary after successful PIN verify.

Add `staff_points` and `points_trend` to the PIN verify response.

---

## 7. RBAC

- `labor:elu` — manage ELU ratings (roles: `shift_manager`, `gm`, `owner`)
- `labor:points` — award/deduct points (roles: `shift_manager`, `gm`, `owner`)
- Existing `staff:read` covers viewing profiles

---

## 8. Testing

- ELU update: set ratings, verify persisted
- Point award: award points, verify employee.staff_points updated
- Point deduction: deduct, verify negative allowed
- Trend calculation: verify up/down/stable thresholds
- Leaderboard: verify ordering by points desc
- Point history: verify chronological order
