# SP6: Operations Intelligence Design Spec

**Date:** 2026-03-19
**Status:** Approved
**Scope:** Operations overview dashboard with order flow, ticket times, hourly volume/revenue, channel performance, void rate — all derived from existing checks data

---

## 1. Backend — Operations Intelligence Service

New Go service at `internal/operations/`. No new tables — all derived from `checks` + `check_items`.

### Data Sources

- `checks` — status, channel, opened_at, closed_at, subtotal, for order flow and timing
- `check_items` — voided_at for void tracking

### Calculations

**1. Orders Today** — `COUNT(*) FROM checks WHERE status = 'closed' AND location_id AND date range`

**2. Avg Ticket Time** — `AVG(EXTRACT(EPOCH FROM (closed_at - opened_at)) / 60.0)` in minutes, for closed checks

**3. Avg Ticket Time by Channel** — Same but grouped by channel

**4. Current Orders/Hour** — Closed checks where `closed_at >= now() - INTERVAL '1 hour'`

**5. Active Tickets** — `COUNT(*) FROM checks WHERE status = 'open' AND location_id`

**6. Longest Open Ticket** — `MIN(opened_at) FROM checks WHERE status = 'open'` → compute minutes since

**7. Revenue/Hour Current** — `SUM(subtotal) FROM checks WHERE status = 'closed' AND closed_at >= now() - INTERVAL '1 hour'`

**8. Void Rate** — `voided_count / (closed_count + voided_count) * 100` (guard div by zero)

**9. Hourly Volume** — For each hour in the date range:
```sql
SELECT EXTRACT(HOUR FROM closed_at)::INT AS hour,
       COUNT(*) AS orders,
       COALESCE(SUM(subtotal), 0)::BIGINT AS revenue
FROM checks
WHERE location_id = $1 AND status = 'closed'
  AND closed_at >= $2 AND closed_at < $3
GROUP BY EXTRACT(HOUR FROM closed_at)
ORDER BY hour
```

**10. Channel Performance** — Per channel: order count, avg ticket time, revenue, % of total

### Types

```go
type OperationsSummary struct {
    OrdersToday       int     `json:"orders_today"`
    AvgTicketTime     float64 `json:"avg_ticket_time"`      // minutes
    OrdersPerHour     int     `json:"orders_per_hour"`       // current hour
    ActiveTickets     int     `json:"active_tickets"`
    LongestOpenMin    float64 `json:"longest_open_min"`      // minutes, 0 if none open
    RevenuePerHour    int64   `json:"revenue_per_hour"`      // cents, current hour
    VoidRate          float64 `json:"void_rate"`             // percentage
    ChannelPerformance []ChannelPerf `json:"channel_performance"`
}

type ChannelPerf struct {
    Channel       string  `json:"channel"`
    Orders        int     `json:"orders"`
    PctOfTotal    float64 `json:"pct_of_total"`
    AvgTicketTime float64 `json:"avg_ticket_time"`  // minutes
    Revenue       int64   `json:"revenue"`           // cents
}

type HourlyData struct {
    Hour    int   `json:"hour"`     // 0-23
    Orders  int   `json:"orders"`
    Revenue int64 `json:"revenue"`  // cents
}
```

### API Endpoints

**`GET /api/v1/operations/summary?location_id=X`**
- Optional: `from`, `to` (defaults to today via `parseDateRange`)
- Returns: `OperationsSummary`

**`GET /api/v1/operations/hourly?location_id=X`**
- Same params
- Returns: `{ hourly: HourlyData[] }`

### Service

```go
func New(pool *pgxpool.Pool, bus *event.Bus) *Service
func (s *Service) GetSummary(ctx context.Context, orgID, locationID string, from, to time.Time) (*OperationsSummary, error)
func (s *Service) GetHourly(ctx context.Context, orgID, locationID string, from, to time.Time) ([]HourlyData, error)
```

### Files

- `internal/operations/operations.go` — Service with types and methods
- `internal/api/operations_handler.go` — HTTP handlers
- `cmd/fireline/main.go` — Wire service and routes

## 2. Frontend — Operations Intelligence Page

### New Files

- `web/src/pages/OperationsPage.tsx`
- `web/src/hooks/useOperations.ts`
- Modify: `web/src/lib/api.ts` — Add operationsApi + types
- Modify: `web/src/App.tsx` — Add `/operations` route
- Modify: `web/src/components/Layout.tsx` — Add nav item

### TypeScript Types

```typescript
export interface OperationsSummary {
  orders_today: number;
  avg_ticket_time: number;
  orders_per_hour: number;
  active_tickets: number;
  longest_open_min: number;
  revenue_per_hour: number;
  void_rate: number;
  channel_performance: ChannelPerf[];
}

export interface ChannelPerf {
  channel: string;
  orders: number;
  pct_of_total: number;
  avg_ticket_time: number;
  revenue: number;
}

export interface HourlyData {
  hour: number;
  orders: number;
  revenue: number;
}
```

### API Client

```typescript
export const operationsApi = {
  getSummary(locationId: string, from?: string, to?: string) {
    const params = new URLSearchParams({ location_id: locationId });
    if (from) params.set('from', from);
    if (to) params.set('to', to);
    return request<OperationsSummary>(`/operations/summary?${params}`);
  },
  getHourly(locationId: string, from?: string, to?: string) {
    const params = new URLSearchParams({ location_id: locationId });
    if (from) params.set('from', from);
    if (to) params.set('to', to);
    return request<{ hourly: HourlyData[] }>(`/operations/hourly?${params}`);
  },
};
```

### Hooks

```typescript
export function useOperationsSummary(locationId, from?, to?)  // staleTime 15s, refetchInterval 15s
export function useOperationsHourly(locationId, from?, to?)    // staleTime 30s
```

Note: 15s refetch for summary (ops needs near-real-time for active tickets).

### Page Layout

**Row 1 — KPI Cards (6 cards, 3x2 grid on lg):**
- Orders Today (ShoppingBag, blue)
- Avg Ticket Time (Clock, purple) — X.X min
- Orders/Hour (TrendingUp, emerald)
- Active Tickets (AlertCircle, orange)
- Revenue/Hour (DollarSign, green) — cents→$
- Void Rate (XCircle, red) — X.X%

**Row 2 — Hourly Chart (Recharts ComposedChart):**
- Bar = orders per hour (left Y-axis, label "Orders")
- Line = revenue per hour in dollars (right Y-axis, label "Revenue")
- X-axis = hour labels ("6AM", "7AM", ... "11PM")
- Tooltip shows both values
- Current hour bar highlighted with different color

**Row 3 — Channel Performance DataTable:**
| Column | Key | Sortable | Align | Render |
|--------|-----|----------|-------|--------|
| Channel | channel | yes | left | label map |
| Orders | orders | yes | right | — |
| % of Total | pct_of_total | yes | right | X.X% |
| Avg Ticket | avg_ticket_time | yes | right | X.X min |
| Revenue | revenue | yes | right | cents→$ |

### Navigation

Icon: `Activity` from lucide-react. After Customers:
```typescript
{ to: '/operations', label: 'Operations', icon: Activity }
```

## 3. Conventions

- Use existing `parseDateRange` (defaults to today — appropriate for ops)
- All money in cents
- Ticket time in minutes (float64)
- 15s refetch for summary (near-real-time ops monitoring)
- No new tables
- `event.Bus` in constructor for future extensibility
