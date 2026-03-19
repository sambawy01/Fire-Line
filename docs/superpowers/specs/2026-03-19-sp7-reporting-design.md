# SP7: Reporting & Analytics Design Spec

**Date:** 2026-03-19
**Status:** Approved
**Scope:** Daily summary report aggregating all modules, operational health score, PDF export, critical issues, staff summary, zero-sales items, category breakdown

---

## 1. Backend — Reporting Service

New Go service at `internal/reporting/` that aggregates data from existing services/queries.

### Dependencies

- `go-pdf/fpdf` — lightweight PDF generation (add to go.mod: `github.com/go-pdf/fpdf`)

### Daily Report Data Structure

```go
type DailyReport struct {
    LocationID       string              `json:"location_id"`
    LocationName     string              `json:"location_name"`
    ReportDate       string              `json:"report_date"`
    HealthScore      int                 `json:"health_score"`        // 0-100 composite

    // KPIs
    NetRevenue       int64               `json:"net_revenue"`         // cents
    GrossMarginPct   float64             `json:"gross_margin_pct"`
    LaborCostPct     float64             `json:"labor_cost_pct"`
    OrdersToday      int                 `json:"orders_today"`
    AvgTicketTime    float64             `json:"avg_ticket_time"`     // minutes
    ActiveAlerts     int                 `json:"active_alerts"`
    CriticalCount    int                 `json:"critical_count"`

    // Critical Issues
    CriticalIssues   []CriticalIssue     `json:"critical_issues"`

    // Channel Breakdown
    Channels         []ReportChannel     `json:"channels"`

    // Menu Performance
    TopItems         []ReportMenuItem    `json:"top_items"`          // top 3 by units
    WorstItem        *ReportMenuItem     `json:"worst_item"`         // lowest performer
    ZeroSalesItems   []string            `json:"zero_sales_items"`   // names of items with 0 sales
    CategoryRevenue  []CategoryRevenue   `json:"category_revenue"`

    // Labor
    StaffSummary     []StaffEntry        `json:"staff_summary"`
    TotalLaborCost   int64               `json:"total_labor_cost"`   // cents
    TotalHoursWorked float64             `json:"total_hours_worked"`
    OvertimeFlags    []string            `json:"overtime_flags"`     // employee names > 10hrs

    // Inventory
    ReorderNeeded    []ReorderItem       `json:"reorder_needed"`
}

type CriticalIssue struct {
    Title     string `json:"title"`
    Module    string `json:"module"`
    CreatedAt string `json:"created_at"`
}

type ReportChannel struct {
    Channel       string  `json:"channel"`
    Orders        int     `json:"orders"`
    Revenue       int64   `json:"revenue"`        // cents
    PctOfTotal    float64 `json:"pct_of_total"`
    AvgTicketTime float64 `json:"avg_ticket_time"` // minutes
}

type ReportMenuItem struct {
    Name           string  `json:"name"`
    Category       string  `json:"category"`
    UnitsSold      int     `json:"units_sold"`
    Revenue        int64   `json:"revenue"`         // cents
    MarginPct      float64 `json:"margin_pct"`
}

type CategoryRevenue struct {
    Category  string `json:"category"`
    Revenue   int64  `json:"revenue"`   // cents
    PctOfTotal float64 `json:"pct_of_total"`
    ItemCount  int    `json:"item_count"`
}

type StaffEntry struct {
    Name        string  `json:"name"`
    Role        string  `json:"role"`
    HoursWorked float64 `json:"hours_worked"`
    LaborCost   int64   `json:"labor_cost"`   // cents
    IsOvertime  bool    `json:"is_overtime"`   // > 10 hours
}

type ReorderItem struct {
    Name         string  `json:"name"`
    CurrentLevel float64 `json:"current_level"`
    PARLevel     float64 `json:"par_level"`
    Unit         string  `json:"unit"`
}
```

### Operational Health Score (0-100)

Composite weighted score:
- **Gross Margin** (25%): score = margin_pct normalized (60%+ = 100, 30% = 50, 0% = 0)
- **Labor Cost** (20%): score = inverted (25% = 100, 35% = 50, 45%+ = 0) — lower is better
- **Void Rate** (15%): score = inverted (0% = 100, 5% = 50, 10%+ = 0) — lower is better
- **Ticket Time** (15%): score = inverted (15min = 100, 25min = 50, 40min+ = 0) — lower is better
- **Order Volume** (15%): score = orders vs capacity estimate (80 orders = 100 for now, linear scale)
- **Critical Alerts** (10%): score = inverted (0 = 100, 1 = 70, 2 = 40, 3+ = 0)

Clamped to 0-100.

### SQL Queries

All within TenantTx for the given location + date range:

**Financial:** Reuse the same pattern as financial service — SUM(subtotal), checks count, etc.

**Menu top/worst items:**
```sql
SELECT mi.name, mi.category, SUM(ci.quantity) AS units_sold,
       SUM(ci.quantity * ci.unit_price)::BIGINT AS revenue
FROM check_items ci
JOIN checks c ON c.check_id = ci.check_id
JOIN menu_items mi ON mi.menu_item_id = ci.menu_item_id
WHERE c.location_id = $1 AND c.status = 'closed' AND ci.voided_at IS NULL
  AND c.closed_at >= $2 AND c.closed_at < $3
GROUP BY mi.menu_item_id, mi.name, mi.category
ORDER BY units_sold DESC
```
Top 3 = first 3 rows. Worst = last row.

**Zero-sales items:**
```sql
SELECT mi.name FROM menu_items mi
WHERE mi.location_id = $1 AND mi.available = true
  AND mi.menu_item_id NOT IN (
    SELECT DISTINCT ci.menu_item_id FROM check_items ci
    JOIN checks c ON c.check_id = ci.check_id
    WHERE c.location_id = $1 AND c.status = 'closed'
      AND c.closed_at >= $2 AND c.closed_at < $3
      AND ci.menu_item_id IS NOT NULL AND ci.voided_at IS NULL
  )
ORDER BY mi.name
```

**Category revenue:**
```sql
SELECT mi.category, SUM(ci.quantity * ci.unit_price)::BIGINT AS revenue,
       COUNT(DISTINCT mi.menu_item_id) AS item_count
FROM check_items ci
JOIN checks c ON c.check_id = ci.check_id
JOIN menu_items mi ON mi.menu_item_id = ci.menu_item_id
WHERE c.location_id = $1 AND c.status = 'closed' AND ci.voided_at IS NULL
  AND c.closed_at >= $2 AND c.closed_at < $3
GROUP BY mi.category ORDER BY revenue DESC
```

**Staff:** Same query pattern as labor service — employees + shifts for the period.

**Inventory reorder:** Items where `current_level <= reorder_point` from ingredient_location_configs (use PAR data).

**Alerts:** Filter alerting queue for org + location, count by severity.

### PDF Generation

Using `github.com/go-pdf/fpdf`:
- Landscape A4
- Header: "FireLine Daily Report" + location name + date
- Health Score: large number with color (green ≥70, yellow 40-69, red <40)
- KPI row: 7 values in a formatted table
- Critical Issues: red-highlighted rows
- Channel table, top menu items table, staff table
- Footer: "Generated by FireLine by OpsNerve"

### API Endpoints

**`GET /api/v1/reports/daily?location_id=X`**
- Optional: `from`, `to` (defaults to today)
- Returns: `DailyReport` as JSON

**`GET /api/v1/reports/daily/pdf?location_id=X`**
- Same params
- Returns: PDF file (Content-Type: application/pdf, Content-Disposition: attachment)

### Service

```go
func New(pool *pgxpool.Pool, bus *event.Bus, alertSvc *alerting.Service) *Service
func (s *Service) GenerateDaily(ctx context.Context, orgID, locationID string, from, to time.Time) (*DailyReport, error)
func (s *Service) GeneratePDF(report *DailyReport) ([]byte, error)
```

Note: Takes `alertSvc` as dependency since alerts are in-memory on the alerting service.

### Files

- `internal/reporting/reporting.go` — Service with GenerateDaily
- `internal/reporting/pdf.go` — PDF generation with fpdf
- `internal/api/reporting_handler.go` — HTTP handlers
- `cmd/fireline/main.go` — Wire service and routes

## 2. Frontend — Reports Page

### New Files

- `web/src/pages/ReportsPage.tsx`
- `web/src/hooks/useReports.ts`
- Modify: `web/src/lib/api.ts`
- Modify: `web/src/App.tsx`
- Modify: `web/src/components/Layout.tsx`

### TypeScript Types

```typescript
export interface DailyReport {
  location_id: string;
  location_name: string;
  report_date: string;
  health_score: number;
  net_revenue: number;
  gross_margin_pct: number;
  labor_cost_pct: number;
  orders_today: number;
  avg_ticket_time: number;
  active_alerts: number;
  critical_count: number;
  critical_issues: CriticalIssue[];
  channels: ReportChannel[];
  top_items: ReportMenuItem[];
  worst_item: ReportMenuItem | null;
  zero_sales_items: string[];
  category_revenue: CategoryRevData[];
  staff_summary: StaffEntry[];
  total_labor_cost: number;
  total_hours_worked: number;
  overtime_flags: string[];
  reorder_needed: ReorderItem[];
}

export interface CriticalIssue {
  title: string;
  module: string;
  created_at: string;
}

export interface ReportChannel {
  channel: string;
  orders: number;
  revenue: number;
  pct_of_total: number;
  avg_ticket_time: number;
}

export interface ReportMenuItem {
  name: string;
  category: string;
  units_sold: number;
  revenue: number;
  margin_pct: number;
}

export interface CategoryRevData {
  category: string;
  revenue: number;
  pct_of_total: number;
  item_count: number;
}

export interface StaffEntry {
  name: string;
  role: string;
  hours_worked: number;
  labor_cost: number;
  is_overtime: boolean;
}

export interface ReorderItem {
  name: string;
  current_level: number;
  par_level: number;
  unit: string;
}
```

### API Client

```typescript
export const reportsApi = {
  getDaily(locationId: string, from?: string, to?: string) {
    const params = new URLSearchParams({ location_id: locationId });
    if (from) params.set('from', from);
    if (to) params.set('to', to);
    return request<DailyReport>(`/reports/daily?${params}`);
  },
  getDailyPdfUrl(locationId: string) {
    const token = localStorage.getItem('access_token');
    return `/api/v1/reports/daily/pdf?location_id=${locationId}&token=${token}`;
  },
};
```

Note: PDF endpoint needs auth token passed as query param since `window.open()` can't set Authorization headers. The handler should accept `?token=` as fallback auth.

### Hooks

```typescript
export function useDailyReport(locationId: string | null)  // staleTime 30s
```

### Page Layout

**Header:** "Daily Report" + location name + date + two download buttons:
- "Download PDF" (orange button) — `window.open(reportsApi.getDailyPdfUrl(locationId))`
- "Export JSON" (outline button) — fetch JSON + save as file

**Health Score Banner:**
- Large circular/pill display showing score 0-100
- Color: green ≥70, yellow 40-69, red <40
- Text: "Operational Health Score"

**Critical Issues Section** (only if critical_count > 0):
- Red-tinted card listing each critical issue with title + module + time

**KPI Cards (7 cards, grid):**
- Net Revenue, Gross Margin %, Labor Cost %, Orders Today, Avg Ticket Time, Active Alerts, Critical Issues

**Channel Breakdown Table** (DataTable)

**Menu Performance Section:**
- "Top Performers" — small table of top 3 items
- "Underperformer" — single card showing worst item
- "Zero Sales" — list of items that didn't sell (warning tint)

**Category Revenue Table** (DataTable)

**Staff Summary Section:**
- Staff table with overtime flags highlighted
- Total hours + total cost at bottom

**Inventory Alerts Section:**
- Items needing reorder (below PAR)

### Navigation

Icon: `FileText` from lucide-react. After Operations:
```typescript
{ to: '/reports', label: 'Reports', icon: FileText }
```

## 3. Conventions

- Reporting service depends on `alerting.Service` for alert data (passed via constructor)
- PDF auth via query param `?token=` as fallback (handler checks both header and query)
- Health score computed server-side, consistent across JSON and PDF
- All money in cents
- Use `parseDateRange` (defaults to today)
