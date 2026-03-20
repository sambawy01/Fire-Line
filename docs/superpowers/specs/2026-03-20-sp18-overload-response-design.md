# SP18: Operations Overload Response & Planning Horizons

**Date:** 2026-03-20
**Status:** Approved
**Scope:** Overload detection, tiered autonomy response, operational health score, 5 planning horizons, order priority scoring
**Maps to:** Build Plan Sprint 26

---

## 1. Overload Detection & Response Engine

New file: `internal/operations/overload.go`

### Overload Detection
- Query current kitchen capacity from existing `CalculateCapacity`
- If `total_capacity_pct > 85%` (configurable): overload state triggered
- Severity levels: `elevated` (85-95%), `critical` (>95%)

### Tiered Response

**Tier 1 (Auto — elevated overload):**
- Extend quoted delivery times by 15 min
- Emit `operations.overload.elevated` event

**Tier 2 (Recommend — critical overload):**
- Suggest 86 of highest-complexity menu items (complexity_score < 30)
- Suggest ticket resequencing (prioritize near-SLA tickets)
- Emit `operations.overload.critical` alert

**Tier 3 (Manager Decision — sustained critical):**
- Present options: reduce menu scope, call in staff, close delivery channel
- Emit `operations.overload.manager_action_needed` alert

### Types
```go
type OverloadStatus struct {
    IsOverloaded     bool    `json:"is_overloaded"`
    CapacityPct      float64 `json:"capacity_pct"`
    Severity         string  `json:"severity"` // "normal", "elevated", "critical"
    ActiveResponses  []OverloadResponse `json:"active_responses"`
    SuggestedActions []SuggestedAction  `json:"suggested_actions"`
}

type OverloadResponse struct {
    Tier        int    `json:"tier"`
    Action      string `json:"action"`
    Description string `json:"description"`
    AutoApplied bool   `json:"auto_applied"`
}

type SuggestedAction struct {
    ActionType  string `json:"action_type"` // "86_item", "resequence", "extend_times", "call_staff", "close_channel"
    Description string `json:"description"`
    Impact      string `json:"impact"`
    ItemID      string `json:"item_id,omitempty"`
}
```

---

## 2. Operational Health Score

New file: `internal/operations/health.go`

### Composite Health Score (0-100)

Components (weighted):
- **Kitchen Load** (25%): 100 - capacity_pct
- **Ticket Performance** (25%): % of tickets completed within SLA (10 min default)
- **Staff Coverage** (20%): scheduled_headcount / required_headcount * 100
- **Financial Health** (15%): budget variance (on_track = 100, over = 50, under = 75)
- **Inventory Health** (15%): 100 - (par_breach_count / total_ingredients * 100)

### Types
```go
type OperationalHealth struct {
    OverallScore      float64          `json:"overall_score"`
    KitchenScore      float64          `json:"kitchen_score"`
    TicketScore       float64          `json:"ticket_score"`
    StaffScore        float64          `json:"staff_score"`
    FinancialScore    float64          `json:"financial_score"`
    InventoryScore    float64          `json:"inventory_score"`
    Status            string           `json:"status"` // "excellent", "good", "fair", "poor", "critical"
}
```

---

## 3. Order Priority Scoring

New file: `internal/operations/priority.go`

### Priority Formula
```
priority = sla_proximity_score * 0.35
         + customer_value_score * 0.25
         + channel_weight * 0.20
         + complexity_inverse * 0.20
```

- `sla_proximity`: minutes_remaining / sla_minutes (lower = higher priority)
- `customer_value`: guest CLV quintile (if resolved)
- `channel_weight`: dine_in=1.0, takeout=0.8, delivery=0.6
- `complexity_inverse`: 1 - (complexity_score / 100)

---

## 4. Planning Horizons

New file: `internal/operations/horizons.go`

### 5 Horizons

1. **Real-Time** — current capacity, active tickets, overload status, health score
2. **Shift (4hr)** — projected demand from forecast, scheduled staff vs required, upcoming deliveries
3. **Daily (24hr)** — prep list (ingredients needed for forecast), equipment schedule, delivery POs expected
4. **Weekly** — schedule overview, PO plan, projected costs
5. **Strategic** — 30-day trends: revenue, COGS, labor cost %, menu performance shifts

Each returns a structured summary for the dashboard.

---

## 5. API Endpoints

```
GET    /api/v1/operations/overload          — Current overload status (query: location_id)
POST   /api/v1/operations/overload/respond  — Apply overload response (body: {action_type, item_id?})
GET    /api/v1/operations/health            — Operational health score (query: location_id)
GET    /api/v1/operations/priority          — Active ticket priorities (query: location_id)
GET    /api/v1/operations/horizon/realtime  — Real-time horizon
GET    /api/v1/operations/horizon/shift     — Shift horizon (4hr)
GET    /api/v1/operations/horizon/daily     — Daily horizon (24hr)
GET    /api/v1/operations/horizon/weekly    — Weekly horizon
GET    /api/v1/operations/horizon/strategic — Strategic horizon (30-day)
```

---

## 6. Web Dashboard — Operations Command Center

Rewrite `OperationsPage.tsx` as the operational command center:

### Top Bar: Health Score + Overload Status
- Large health score gauge (0-100, color-coded)
- Overload indicator: green/yellow/red
- Active responses shown if overloaded

### Section 1: Real-Time View (default)
- Station load cards (from Kitchen page data)
- Active ticket count + avg ticket time
- Overload response controls

### Section 2: Planning Horizons (tabs or collapsible)
- Shift: demand forecast chart + staffing gaps
- Daily: prep list + delivery schedule
- Weekly: schedule summary + PO plan
- Strategic: trend sparklines for key metrics

---

## 7. Testing
- Overload detection: 86% capacity → elevated, 96% → critical
- Health score: known component scores → correct weighted composite
- Priority: near-SLA dine-in > far-SLA delivery
- Horizon data structures return correctly
