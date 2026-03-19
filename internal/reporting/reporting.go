package reporting

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/opsnerve/fireline/internal/alerting"
	"github.com/opsnerve/fireline/internal/event"
	"github.com/opsnerve/fireline/internal/tenant"
	"github.com/opsnerve/fireline/pkg/database"
)

// ─── Report Types ────────────────────────────────────────────────────────────

// DailyReport is the full aggregated daily summary for a location.
type DailyReport struct {
	LocationID       string            `json:"location_id"`
	LocationName     string            `json:"location_name"`
	ReportDate       string            `json:"report_date"`
	HealthScore      int               `json:"health_score"`
	NetRevenue       int64             `json:"net_revenue"`
	GrossMarginPct   float64           `json:"gross_margin_pct"`
	LaborCostPct     float64           `json:"labor_cost_pct"`
	OrdersToday      int               `json:"orders_today"`
	AvgTicketTime    float64           `json:"avg_ticket_time"`
	ActiveAlerts     int               `json:"active_alerts"`
	CriticalCount    int               `json:"critical_count"`
	CriticalIssues   []CriticalIssue   `json:"critical_issues"`
	Channels         []ReportChannel   `json:"channels"`
	TopItems         []ReportMenuItem  `json:"top_items"`
	WorstItem        *ReportMenuItem   `json:"worst_item"`
	ZeroSalesItems   []string          `json:"zero_sales_items"`
	CategoryRevenue  []CategoryRevenue `json:"category_revenue"`
	StaffSummary     []StaffEntry      `json:"staff_summary"`
	TotalLaborCost   int64             `json:"total_labor_cost"`
	TotalHoursWorked float64           `json:"total_hours_worked"`
	OvertimeFlags    []string          `json:"overtime_flags"`
	ReorderNeeded    []ReorderItem     `json:"reorder_needed"`
}

// CriticalIssue summarizes a critical alert for the report.
type CriticalIssue struct {
	AlertID     string `json:"alert_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Module      string `json:"module"`
}

// ReportChannel shows order volume and revenue by sales channel.
type ReportChannel struct {
	Channel   string  `json:"channel"`
	Orders    int     `json:"orders"`
	Revenue   int64   `json:"revenue"`
	AvgTicket float64 `json:"avg_ticket"`
}

// ReportMenuItem represents a menu item's sales performance.
type ReportMenuItem struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Units    int64  `json:"units"`
	Revenue  int64  `json:"revenue"`
}

// CategoryRevenue shows aggregated revenue per menu category.
type CategoryRevenue struct {
	Category string `json:"category"`
	Revenue  int64  `json:"revenue"`
	Units    int64  `json:"units"`
}

// StaffEntry is one employee's labor summary for the period.
type StaffEntry struct {
	Name  string  `json:"name"`
	Role  string  `json:"role"`
	Hours float64 `json:"hours"`
	Cost  int64   `json:"cost"`
}

// ReorderItem flags an ingredient that has PAR configuration (placeholder until real counts arrive).
type ReorderItem struct {
	Name         string  `json:"name"`
	Unit         string  `json:"unit"`
	ParLevel     float64 `json:"par_level"`
	ReorderPoint float64 `json:"reorder_point"`
	CurrentLevel float64 `json:"current_level"` // 0 until live inventory counts are tracked
}

// ─── Service ─────────────────────────────────────────────────────────────────

// Service aggregates multi-module data into reports.
type Service struct {
	pool     *pgxpool.Pool
	bus      *event.Bus
	alertSvc *alerting.Service
}

// New creates a new reporting service.
func New(pool *pgxpool.Pool, bus *event.Bus, alertSvc *alerting.Service) *Service {
	return &Service{pool: pool, bus: bus, alertSvc: alertSvc}
}

// GenerateDaily builds a full DailyReport for the given location and date range.
func (s *Service) GenerateDaily(ctx context.Context, orgID, locationID string, from, to time.Time) (*DailyReport, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	report := &DailyReport{
		LocationID:     locationID,
		ReportDate:     from.Format("2006-01-02"),
		CriticalIssues: []CriticalIssue{},
		Channels:       []ReportChannel{},
		TopItems:       []ReportMenuItem{},
		ZeroSalesItems: []string{},
		CategoryRevenue: []CategoryRevenue{},
		StaffSummary:   []StaffEntry{},
		OvertimeFlags:  []string{},
		ReorderNeeded:  []ReorderItem{},
	}

	var voidCount int64

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		// 1. Location name
		if err := tx.QueryRow(tenantCtx,
			`SELECT name FROM locations WHERE location_id = $1`,
			locationID,
		).Scan(&report.LocationName); err != nil {
			return fmt.Errorf("location name: %w", err)
		}

		// 2. Financial KPIs
		if err := tx.QueryRow(tenantCtx,
			`SELECT COUNT(*) AS order_count,
			        COALESCE(SUM(subtotal), 0)::BIGINT AS net_revenue,
			        COALESCE(AVG(EXTRACT(EPOCH FROM (closed_at - opened_at)) / 60.0), 0) AS avg_ticket
			 FROM checks
			 WHERE location_id = $1 AND status = 'closed'
			   AND closed_at >= $2 AND closed_at < $3`,
			locationID, from, to,
		).Scan(&report.OrdersToday, &report.NetRevenue, &report.AvgTicketTime); err != nil {
			return fmt.Errorf("financial KPIs: %w", err)
		}

		// 3. COGS
		var cogs int64
		if err := tx.QueryRow(tenantCtx,
			`SELECT COALESCE(SUM(ci.quantity * re.quantity_per_unit *
			         COALESCE(ilc.local_cost_per_unit, i.cost_per_unit))::BIGINT, 0) AS cogs
			 FROM check_items ci
			 JOIN checks c ON c.check_id = ci.check_id
			 JOIN recipe_explosion re ON re.menu_item_id = ci.menu_item_id AND re.org_id = ci.org_id
			 JOIN ingredients i ON i.ingredient_id = re.ingredient_id
			 LEFT JOIN ingredient_location_configs ilc
			        ON ilc.ingredient_id = i.ingredient_id
			       AND ilc.location_id = $1
			       AND ilc.org_id = ci.org_id
			 WHERE c.location_id = $1 AND c.status = 'closed' AND ci.voided_at IS NULL
			   AND c.closed_at >= $2 AND c.closed_at < $3`,
			locationID, from, to,
		).Scan(&cogs); err != nil {
			return fmt.Errorf("COGS: %w", err)
		}
		if report.NetRevenue > 0 {
			report.GrossMarginPct = float64(report.NetRevenue-cogs) / float64(report.NetRevenue) * 100
		}

		// 4. Labor
		rows, err := tx.Query(tenantCtx,
			`SELECT e.display_name, e.role,
			        COALESCE(SUM(EXTRACT(EPOCH FROM (
			            COALESCE(s.clock_out, LEAST(now(), s.clock_in + INTERVAL '16 hours')) - s.clock_in
			        )) / 3600.0), 0) AS hours,
			        COALESCE(SUM((EXTRACT(EPOCH FROM (
			            COALESCE(s.clock_out, LEAST(now(), s.clock_in + INTERVAL '16 hours')) - s.clock_in
			        )) / 3600.0 * s.hourly_rate)::BIGINT), 0) AS cost
			 FROM employees e
			 LEFT JOIN shifts s ON s.employee_id = e.employee_id
			        AND s.status != 'no_show'
			        AND s.clock_in >= $2 AND s.clock_in < $3
			 WHERE e.location_id = $1 AND e.status = 'active'
			 GROUP BY e.employee_id, e.display_name, e.role
			 ORDER BY e.display_name`,
			locationID, from, to,
		)
		if err != nil {
			return fmt.Errorf("labor query: %w", err)
		}
		defer rows.Close()

		var totalLaborCost int64
		var totalHours float64
		for rows.Next() {
			var entry StaffEntry
			if err := rows.Scan(&entry.Name, &entry.Role, &entry.Hours, &entry.Cost); err != nil {
				return fmt.Errorf("labor scan: %w", err)
			}
			report.StaffSummary = append(report.StaffSummary, entry)
			totalLaborCost += entry.Cost
			totalHours += entry.Hours
			if entry.Hours > 10 {
				report.OvertimeFlags = append(report.OvertimeFlags, entry.Name)
			}
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("labor rows: %w", err)
		}
		report.TotalLaborCost = totalLaborCost
		report.TotalHoursWorked = totalHours
		if report.NetRevenue > 0 {
			report.LaborCostPct = float64(totalLaborCost) / float64(report.NetRevenue) * 100
		}

		// 5. Channel breakdown
		chRows, err := tx.Query(tenantCtx,
			`SELECT channel, COUNT(*) AS orders,
			        COALESCE(SUM(subtotal), 0)::BIGINT AS revenue,
			        COALESCE(AVG(EXTRACT(EPOCH FROM (closed_at - opened_at)) / 60.0), 0) AS avg_ticket
			 FROM checks
			 WHERE location_id = $1 AND status = 'closed'
			   AND closed_at >= $2 AND closed_at < $3
			 GROUP BY channel ORDER BY orders DESC`,
			locationID, from, to,
		)
		if err != nil {
			return fmt.Errorf("channel query: %w", err)
		}
		defer chRows.Close()

		for chRows.Next() {
			var ch ReportChannel
			if err := chRows.Scan(&ch.Channel, &ch.Orders, &ch.Revenue, &ch.AvgTicket); err != nil {
				return fmt.Errorf("channel scan: %w", err)
			}
			report.Channels = append(report.Channels, ch)
		}
		if err := chRows.Err(); err != nil {
			return fmt.Errorf("channel rows: %w", err)
		}

		// 6. Menu item performance
		miRows, err := tx.Query(tenantCtx,
			`SELECT mi.name, mi.category,
			        SUM(ci.quantity) AS units,
			        SUM(ci.quantity * ci.unit_price)::BIGINT AS revenue
			 FROM check_items ci
			 JOIN checks c ON c.check_id = ci.check_id
			 JOIN menu_items mi ON mi.menu_item_id = ci.menu_item_id
			 WHERE c.location_id = $1 AND c.status = 'closed' AND ci.voided_at IS NULL
			   AND c.closed_at >= $2 AND c.closed_at < $3
			 GROUP BY mi.menu_item_id, mi.name, mi.category
			 ORDER BY units DESC`,
			locationID, from, to,
		)
		if err != nil {
			return fmt.Errorf("menu items query: %w", err)
		}
		defer miRows.Close()

		var allItems []ReportMenuItem
		for miRows.Next() {
			var item ReportMenuItem
			if err := miRows.Scan(&item.Name, &item.Category, &item.Units, &item.Revenue); err != nil {
				return fmt.Errorf("menu item scan: %w", err)
			}
			allItems = append(allItems, item)
		}
		if err := miRows.Err(); err != nil {
			return fmt.Errorf("menu item rows: %w", err)
		}

		// Top 3 items
		if len(allItems) >= 3 {
			report.TopItems = allItems[:3]
		} else {
			report.TopItems = allItems
		}
		// Worst item = last row (lowest units)
		if len(allItems) > 0 {
			worst := allItems[len(allItems)-1]
			report.WorstItem = &worst
		}

		// Category revenue — group allItems by category
		catMap := map[string]*CategoryRevenue{}
		catOrder := []string{}
		for _, item := range allItems {
			if _, ok := catMap[item.Category]; !ok {
				catMap[item.Category] = &CategoryRevenue{Category: item.Category}
				catOrder = append(catOrder, item.Category)
			}
			catMap[item.Category].Revenue += item.Revenue
			catMap[item.Category].Units += item.Units
		}
		for _, cat := range catOrder {
			report.CategoryRevenue = append(report.CategoryRevenue, *catMap[cat])
		}

		// 7. Zero-sales items
		zsRows, err := tx.Query(tenantCtx,
			`SELECT name FROM menu_items
			 WHERE location_id = $1 AND available = true
			   AND menu_item_id NOT IN (
			       SELECT DISTINCT ci.menu_item_id FROM check_items ci
			       JOIN checks c ON c.check_id = ci.check_id
			       WHERE c.location_id = $1 AND c.status = 'closed'
			         AND c.closed_at >= $2 AND c.closed_at < $3
			         AND ci.menu_item_id IS NOT NULL AND ci.voided_at IS NULL
			   )
			 ORDER BY name`,
			locationID, from, to,
		)
		if err != nil {
			return fmt.Errorf("zero-sales query: %w", err)
		}
		defer zsRows.Close()

		for zsRows.Next() {
			var name string
			if err := zsRows.Scan(&name); err != nil {
				return fmt.Errorf("zero-sales scan: %w", err)
			}
			report.ZeroSalesItems = append(report.ZeroSalesItems, name)
		}
		if err := zsRows.Err(); err != nil {
			return fmt.Errorf("zero-sales rows: %w", err)
		}

		// 8. Void count
		if err := tx.QueryRow(tenantCtx,
			`SELECT COUNT(*) FROM checks
			 WHERE location_id = $1 AND status = 'voided'
			   AND opened_at >= $2 AND opened_at < $3`,
			locationID, from, to,
		).Scan(&voidCount); err != nil {
			return fmt.Errorf("void count: %w", err)
		}

		// 9. Reorder needed (items with PAR configuration)
		reorderRows, err := tx.Query(tenantCtx,
			`SELECT i.name, ilc.par_level, ilc.reorder_point, i.unit
			 FROM ingredient_location_configs ilc
			 JOIN ingredients i ON i.ingredient_id = ilc.ingredient_id
			 WHERE ilc.location_id = $1
			   AND ilc.par_level IS NOT NULL AND ilc.reorder_point IS NOT NULL`,
			locationID,
		)
		if err != nil {
			return fmt.Errorf("reorder query: %w", err)
		}
		defer reorderRows.Close()

		for reorderRows.Next() {
			var ri ReorderItem
			if err := reorderRows.Scan(&ri.Name, &ri.ParLevel, &ri.ReorderPoint, &ri.Unit); err != nil {
				return fmt.Errorf("reorder scan: %w", err)
			}
			// CurrentLevel = 0 placeholder until live inventory counts are tracked
			ri.CurrentLevel = 0
			report.ReorderNeeded = append(report.ReorderNeeded, ri)
		}
		if err := reorderRows.Err(); err != nil {
			return fmt.Errorf("reorder rows: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("generate daily report: %w", err)
	}

	// 10. Alerts (outside TenantTx — in-memory, no DB needed)
	alerts := s.alertSvc.GetQueue(orgID, locationID)
	report.ActiveAlerts = s.alertSvc.ActiveCount(orgID)
	for _, a := range alerts {
		if a.Severity == alerting.SeverityCritical {
			report.CriticalCount++
			report.CriticalIssues = append(report.CriticalIssues, CriticalIssue{
				AlertID:     a.AlertID,
				Title:       a.Title,
				Description: a.Description,
				Module:      a.Module,
			})
		}
	}

	// Health score
	var voidRate float64
	if report.OrdersToday+int(voidCount) > 0 {
		voidRate = float64(voidCount) / float64(report.OrdersToday+int(voidCount)) * 100
	}
	report.HealthScore = calculateHealthScore(report, voidRate)

	return report, nil
}

// ─── Health Score ─────────────────────────────────────────────────────────────

func calculateHealthScore(report *DailyReport, voidRate float64) int {
	// Margin: 60%+ = 100, 30% = 50, 0% = 0
	marginScore := clamp(report.GrossMarginPct/60.0*100, 0, 100)
	// Labor: 25% = 100, 35% = 50, 45%+ = 0
	laborScore := clamp((45.0-report.LaborCostPct)/20.0*100, 0, 100)
	// Void: 0% = 100, 5% = 50, 10%+ = 0
	voidScore := clamp((10.0-voidRate)/10.0*100, 0, 100)
	// Ticket: 15min = 100, 25min = 50, 40min+ = 0
	ticketScore := clamp((40.0-report.AvgTicketTime)/25.0*100, 0, 100)
	// Orders: linear, 80 = 100
	orderScore := clamp(float64(report.OrdersToday)/80.0*100, 0, 100)
	// Critical alerts
	critScore := 100.0
	if report.CriticalCount == 1 {
		critScore = 70
	}
	if report.CriticalCount == 2 {
		critScore = 40
	}
	if report.CriticalCount >= 3 {
		critScore = 0
	}

	score := marginScore*0.25 + laborScore*0.20 + voidScore*0.15 + ticketScore*0.15 + orderScore*0.15 + critScore*0.10
	return int(math.Round(score))
}

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
