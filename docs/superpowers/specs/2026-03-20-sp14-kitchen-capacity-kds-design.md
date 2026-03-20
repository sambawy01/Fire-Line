# SP14: Kitchen Capacity Model & KDS (Kitchen Display System)

**Date:** 2026-03-20
**Status:** Approved
**Scope:** Kitchen resource model, capacity calculator, ticket time prediction, KDS ticket routing, station-specific displays, kitchen load visualization
**Maps to:** Build Plan Sprints 24-25 (Operations — Kitchen Capacity + KDS)

---

## 1. Database — Migration 011

New migration: `migrations/011_kitchen_operations.sql`

### New Tables

```sql
-- Kitchen stations configuration
CREATE TABLE kitchen_stations (
    station_id     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID NOT NULL REFERENCES organizations(org_id),
    location_id    UUID NOT NULL REFERENCES locations(location_id),
    name           TEXT NOT NULL,
    station_type   TEXT NOT NULL,
    max_concurrent INT NOT NULL DEFAULT 4,
    status         TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Menu item resource profiles (how items use kitchen resources)
CREATE TABLE menu_item_resource_profiles (
    profile_id     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID NOT NULL REFERENCES organizations(org_id),
    menu_item_id   UUID NOT NULL REFERENCES menu_items(menu_item_id),
    station_type   TEXT NOT NULL,
    task_sequence  INT NOT NULL DEFAULT 1,
    duration_secs  INT NOT NULL DEFAULT 300,
    elu_required   NUMERIC(4,2) NOT NULL DEFAULT 1.0,
    batch_size     INT NOT NULL DEFAULT 1,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(org_id, menu_item_id, station_type, task_sequence)
);

-- KDS tickets (kitchen orders routed to stations)
CREATE TABLE kds_tickets (
    ticket_id      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID NOT NULL REFERENCES organizations(org_id),
    location_id    UUID NOT NULL REFERENCES locations(location_id),
    check_id       UUID REFERENCES checks(check_id),
    order_number   TEXT,
    channel        TEXT,
    status         TEXT NOT NULL DEFAULT 'new' CHECK (status IN ('new', 'in_progress', 'ready', 'delivered', 'cancelled')),
    priority       INT NOT NULL DEFAULT 0,
    estimated_ready_at TIMESTAMPTZ,
    actual_ready_at    TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Individual KDS ticket items (one per item-station combination)
CREATE TABLE kds_ticket_items (
    ticket_item_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID NOT NULL REFERENCES organizations(org_id),
    ticket_id      UUID NOT NULL REFERENCES kds_tickets(ticket_id),
    menu_item_id   UUID NOT NULL REFERENCES menu_items(menu_item_id),
    item_name      TEXT NOT NULL,
    quantity       INT NOT NULL DEFAULT 1,
    station_type   TEXT NOT NULL,
    status         TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'fired', 'cooking', 'ready', 'cancelled')),
    fire_at        TIMESTAMPTZ,
    started_at     TIMESTAMPTZ,
    completed_at   TIMESTAMPTZ,
    duration_secs  INT,
    notes          TEXT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

RLS on all 4 tables. Indexes on location_id, check_id, station_type, status.

---

## 2. Kitchen Resource Model

New file: `internal/operations/kitchen.go`

### Types

```go
type KitchenStation struct {
    StationID     string `json:"station_id"`
    Name          string `json:"name"`
    StationType   string `json:"station_type"`
    MaxConcurrent int    `json:"max_concurrent"`
    CurrentLoad   int    `json:"current_load"`
    LoadPct       float64 `json:"load_pct"`
    Status        string `json:"status"`
}

type ResourceProfile struct {
    MenuItemID   string  `json:"menu_item_id"`
    ItemName     string  `json:"item_name"`
    StationType  string  `json:"station_type"`
    TaskSequence int     `json:"task_sequence"`
    DurationSecs int     `json:"duration_secs"`
    ELURequired  float64 `json:"elu_required"`
    BatchSize    int     `json:"batch_size"`
}

type KitchenCapacity struct {
    Stations        []KitchenStation `json:"stations"`
    TotalCapacityPct float64        `json:"total_capacity_pct"`
    ActiveTickets   int              `json:"active_tickets"`
    AvgTicketTime   int              `json:"avg_ticket_time_secs"`
    EstNextReady    int              `json:"est_next_ready_secs"`
}

type TicketTimeEstimate struct {
    MenuItemID    string `json:"menu_item_id"`
    ItemName      string `json:"item_name"`
    EstimatedSecs int    `json:"estimated_secs"`
    Confidence    string `json:"confidence"` // "high", "medium", "low"
}
```

### Methods on operations `*Service`

- `SetupDefaultStations(ctx, orgID, locationID)` — create default stations (grill, fryer, saute, prep, expo, dish)
- `GetStations(ctx, orgID, locationID)` — list stations with current load
- `GetResourceProfiles(ctx, orgID, menuItemID)` — get task sequence for a menu item
- `SetResourceProfile(ctx, orgID, menuItemID, profiles []ResourceProfile)` — set/update profiles
- `CalculateCapacity(ctx, orgID, locationID)` — compute current kitchen capacity
- `EstimateTicketTime(ctx, orgID, locationID, menuItemIDs []string)` — estimate time for items

### Ticket Time Prediction (Tier 0)

- Default times by station type: grill=420s, fryer=300s, saute=360s, prep=180s
- If resource profiles exist: sum task durations for the item
- If historical data exists (completed tickets): use trailing avg of actual ticket times

---

## 3. KDS Service

New file: `internal/operations/kds.go`

### Types

```go
type KDSTicket struct {
    TicketID        string          `json:"ticket_id"`
    OrderNumber     string          `json:"order_number"`
    Channel         string          `json:"channel"`
    Status          string          `json:"status"`
    Priority        int             `json:"priority"`
    EstimatedReady  *time.Time      `json:"estimated_ready_at"`
    ActualReady     *time.Time      `json:"actual_ready_at"`
    ElapsedSecs     int             `json:"elapsed_secs"`
    Items           []KDSTicketItem `json:"items"`
    CreatedAt       time.Time       `json:"created_at"`
}

type KDSTicketItem struct {
    TicketItemID string     `json:"ticket_item_id"`
    ItemName     string     `json:"item_name"`
    Quantity     int        `json:"quantity"`
    StationType  string     `json:"station_type"`
    Status       string     `json:"status"`
    FireAt       *time.Time `json:"fire_at"`
    DurationSecs *int       `json:"duration_secs"`
}
```

### Methods

- `CreateTicketFromCheck(ctx, orgID, locationID, checkID)` — decompose order items into station-specific ticket items using resource profiles, set estimated_ready_at, emit `operations.ticket.created`
- `GetStationTickets(ctx, orgID, locationID, stationType)` — active tickets for a station
- `GetAllTickets(ctx, orgID, locationID)` — all active tickets (expo view)
- `BumpTicketItem(ctx, orgID, ticketItemID, newStatus)` — update item status (pending→fired→cooking→ready), if all items ready: mark ticket as ready
- `CancelTicket(ctx, orgID, ticketID)` — cancel ticket and all items
- `GetKDSMetrics(ctx, orgID, locationID, from, to)` — avg ticket time, items/hr, bump times per station

---

## 4. API Endpoints

### Kitchen Configuration
```
GET    /api/v1/operations/stations              — List stations with load
POST   /api/v1/operations/stations/setup        — Create default stations
GET    /api/v1/operations/capacity              — Current kitchen capacity
GET    /api/v1/operations/ticket-time-estimate  — Estimate time for items
```

### Resource Profiles
```
GET    /api/v1/operations/resource-profiles/{menu_item_id} — Get profiles for item
PUT    /api/v1/operations/resource-profiles/{menu_item_id} — Set profiles for item
```

### KDS
```
POST   /api/v1/operations/kds/tickets           — Create ticket from check
GET    /api/v1/operations/kds/tickets           — All active tickets (expo view)
GET    /api/v1/operations/kds/station/{type}    — Station-specific tickets
PUT    /api/v1/operations/kds/items/{id}/bump   — Bump ticket item status
DELETE /api/v1/operations/kds/tickets/{id}      — Cancel ticket
GET    /api/v1/operations/kds/metrics           — KDS performance metrics
```

---

## 5. Web Dashboard — Kitchen Operations Page

New page: `web/src/pages/KitchenPage.tsx` (route `/kitchen`, nav after Operations)

### Layout

**Section 1: Kitchen Load** — station cards showing load % as progress bars, color-coded (green <50%, yellow 50-80%, red >80%)

**Section 2: Active Tickets** (expo view) — card per ticket showing order number, channel, items with station badges and status, elapsed time with color warning at threshold

**Section 3: KDS Metrics** — avg ticket time, items/hr, tickets completed today, bump time by station

**Section 4: Resource Profiles** — table of menu items with their station assignments and durations, editable

---

## 6. Tablet — KDS Screen

Replace the KDS placeholder tab with a station-specific ticket view.

- On load, select station (based on employee's highest ELU station)
- Show tickets assigned to that station
- Each ticket: order #, item name, quantity, elapsed time, urgency color
- "Start" button → fires item (status: cooking)
- "Done" button → marks ready
- Timer flashes red at SLA threshold (e.g., 10 minutes)

---

## 7. RBAC

- `operations:kitchen` — manage stations, resource profiles (roles: `shift_manager`, `gm`, `owner`)
- `operations:kds` — bump KDS items (roles: `staff`, `shift_manager`, `gm`, `owner`)

---

## 8. Testing

- Capacity calculation: known station loads → correct utilization %
- Ticket time estimate: items with resource profiles → sum of durations
- KDS ticket creation: check with items → decomposed to station-specific ticket items
- Bump lifecycle: pending → fired → cooking → ready → ticket complete
- Default station types seeded correctly
