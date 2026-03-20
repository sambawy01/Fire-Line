package alerting

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/opsnerve/fireline/internal/event"
)

// Severity levels for alerts.
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityCritical Severity = "critical"
)

// Alert represents a triggered alert in the priority action queue.
type Alert struct {
	AlertID     string    `json:"alert_id"`
	OrgID       string    `json:"org_id"`
	LocationID  string    `json:"location_id"`
	RuleID      string    `json:"rule_id"`
	Severity    Severity  `json:"severity"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Module      string    `json:"module"` // "inventory", "financial", "adapter"
	Status      string    `json:"status"` // "active", "acknowledged", "resolved"
	CreatedAt   time.Time `json:"created_at"`
	AckedAt     *time.Time `json:"acked_at"`
	ResolvedAt  *time.Time `json:"resolved_at"`
	Metadata    map[string]any `json:"metadata"`
}

// Rule defines a condition that triggers an alert.
type Rule struct {
	RuleID      string   `json:"rule_id"`
	Name        string   `json:"name"`
	Module      string   `json:"module"`      // which module's events to watch
	EventType   string   `json:"event_type"`  // event bus subject to match
	Severity    Severity `json:"severity"`
	Enabled     bool     `json:"enabled"`
	Evaluate    RuleEvaluator `json:"-"` // function that evaluates the condition
}

// RuleEvaluator checks an event and returns an alert if the condition is met.
type RuleEvaluator func(ctx context.Context, env event.Envelope) *Alert

// Service manages alert rules and the priority action queue.
type Service struct {
	bus   *event.Bus
	mu    sync.RWMutex
	rules []Rule
	queue []Alert // priority action queue (in-memory for now)
	seq   int64
}

// New creates a new alerting service.
func New(bus *event.Bus) *Service {
	return &Service{bus: bus}
}

// AddRule registers an alert rule.
func (s *Service) AddRule(rule Rule) {
	s.mu.Lock()
	s.rules = append(s.rules, rule)
	s.mu.Unlock()

	// Subscribe to the event type
	s.bus.Subscribe(rule.EventType, func(ctx context.Context, env event.Envelope) error {
		if !rule.Enabled {
			return nil
		}
		alert := rule.Evaluate(ctx, env)
		if alert != nil {
			alert.RuleID = rule.RuleID
			alert.Severity = rule.Severity
			alert.Module = rule.Module
			s.enqueue(*alert)
		}
		return nil
	})
}

// RegisterDefaultRules sets up standard alert rules for inventory and financial modules.
func (s *Service) RegisterDefaultRules() {
	s.AddRule(Rule{
		RuleID:    "inv-usage-updated",
		Name:      "Inventory Usage Updated",
		Module:    "inventory",
		EventType: "inventory.usage.updated",
		Severity:  SeverityInfo,
		Enabled:   true,
		Evaluate: func(ctx context.Context, env event.Envelope) *Alert {
			return &Alert{
				OrgID:       env.OrgID,
				LocationID:  env.LocationID,
				Title:       "Inventory usage recalculated",
				Description: "Theoretical usage updated after new orders processed",
				Status:      "active",
				Metadata:    map[string]any{"event_id": env.EventID},
			}
		},
	})

	s.AddRule(Rule{
		RuleID:    "fin-metrics-updated",
		Name:      "Financial Metrics Updated",
		Module:    "financial",
		EventType: "financial.metrics.updated",
		Severity:  SeverityInfo,
		Enabled:   true,
		Evaluate: func(ctx context.Context, env event.Envelope) *Alert {
			return &Alert{
				OrgID:       env.OrgID,
				LocationID:  env.LocationID,
				Title:       "Financial metrics recalculated",
				Description: "P&L and margins updated after new orders processed",
				Status:      "active",
				Metadata:    map[string]any{"event_id": env.EventID},
			}
		},
	})

	s.AddRule(Rule{
		RuleID:    "adapter-error",
		Name:      "Adapter Error",
		Module:    "adapter",
		EventType: "adapter.error",
		Severity:  SeverityCritical,
		Enabled:   true,
		Evaluate: func(ctx context.Context, env event.Envelope) *Alert {
			return &Alert{
				OrgID:       env.OrgID,
				LocationID:  env.LocationID,
				Title:       "POS adapter error",
				Description: "Adapter encountered an error and may need attention",
				Status:      "active",
				Metadata:    map[string]any{"source": env.Source},
			}
		},
	})
}

// enqueue adds an alert to the priority action queue.
func (s *Service) enqueue(alert Alert) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seq++
	alert.AlertID = formatAlertID(s.seq)
	alert.CreatedAt = time.Now()
	s.queue = append(s.queue, alert)

	slog.Info("alert enqueued",
		"alert_id", alert.AlertID,
		"severity", alert.Severity,
		"title", alert.Title,
		"org_id", alert.OrgID,
		"location_id", alert.LocationID,
	)

	// Publish alert event for real-time notification
	s.bus.Publish(context.Background(), event.Envelope{
		EventID:    alert.AlertID,
		EventType:  "alerting.alert.created",
		OrgID:      alert.OrgID,
		LocationID: alert.LocationID,
		Source:     "alerting",
		Payload:    alert,
	})
}

// GetQueue returns active alerts, ordered by severity (critical first).
func (s *Service) GetQueue(orgID, locationID string) []Alert {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []Alert
	for _, a := range s.queue {
		if a.OrgID == orgID && a.Status == "active" {
			if locationID == "" || a.LocationID == locationID {
				results = append(results, a)
			}
		}
	}

	// Sort: critical > warning > info
	sortAlerts(results)
	return results
}

// Acknowledge marks an alert as acknowledged.
func (s *Service) Acknowledge(alertID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, a := range s.queue {
		if a.AlertID == alertID && a.Status == "active" {
			now := time.Now()
			s.queue[i].Status = "acknowledged"
			s.queue[i].AckedAt = &now
			return true
		}
	}
	return false
}

// Resolve marks an alert as resolved.
func (s *Service) Resolve(alertID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, a := range s.queue {
		if a.AlertID == alertID && (a.Status == "active" || a.Status == "acknowledged") {
			now := time.Now()
			s.queue[i].Status = "resolved"
			s.queue[i].ResolvedAt = &now
			return true
		}
	}
	return false
}

// ActiveCount returns the number of active alerts for an org.
func (s *Service) ActiveCount(orgID string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	for _, a := range s.queue {
		if a.OrgID == orgID && a.Status == "active" {
			count++
		}
	}
	return count
}

// SeedAlerts injects demo alerts for development/testing purposes.
// Expects locations in order: [El Gouna, New Cairo, Sheikh Zayed, North Coast]
func (s *Service) SeedAlerts(orgID string, locations []string) {
	now := time.Now()

	// Helper to safely get location ID (falls back to first location)
	loc := func(idx int) string {
		if idx < len(locations) {
			return locations[idx]
		}
		return locations[0]
	}

	demos := []Alert{
		// ── Predictive Maintenance (3) ──
		{
			OrgID: orgID, LocationID: loc(0), Severity: SeverityWarning, Module: "operations",
			Title:       "Grill station temperature sensor trending 8°C above baseline",
			Description: "El Gouna grill station #1 has been running 8°C above the 3-month average for the past 48 hours. Maintenance recommended within 48 hours to prevent equipment failure during peak service.",
			Status: "active", CreatedAt: now.Add(-2 * time.Hour),
			Metadata: map[string]any{"station": "Grill Station", "temp_delta_c": 8, "location": "El Gouna"},
		},
		{
			OrgID: orgID, LocationID: loc(1), Severity: SeverityWarning, Module: "operations",
			Title:       "Walk-in cooler compressor runtime increased 23% this week",
			Description: "New Cairo walk-in cooler compressor is running 23% longer per cycle compared to last week. This pattern preceded a failure at El Gouna 3 months ago. Schedule inspection before Friday dinner service.",
			Status: "active", CreatedAt: now.Add(-4 * time.Hour),
			Metadata: map[string]any{"equipment": "Walk-in Cooler", "runtime_increase_pct": 23, "location": "New Cairo"},
		},
		{
			OrgID: orgID, LocationID: loc(2), Severity: SeverityInfo, Module: "operations",
			Title:       "Dishwasher water pressure dropped 15% — filter replacement due",
			Description: "Sheikh Zayed dishwasher water pressure has dropped 15% over 2 weeks. Filter replacement typically resolves this. Last filter change was 45 days ago (recommended: 30 days).",
			Status: "active", CreatedAt: now.Add(-6 * time.Hour),
			Metadata: map[string]any{"equipment": "Dishwasher", "pressure_drop_pct": 15, "days_since_filter": 45},
		},

		// ── Inventory Intelligence (3) ──
		{
			OrgID: orgID, LocationID: loc(0), Severity: SeverityCritical, Module: "inventory",
			Title:       "Sea Bass wholesale price up 18% this month",
			Description: "Sea Bass cost from Sysco Egypt increased from 38 EGP/lb to 48 EGP/lb over 6 months. Ceviche Clasico food cost now at 36% (target: 32%). Consider menu price adjustment from 285 to 310 EGP or substitute with local catch.",
			Status: "active", CreatedAt: now.Add(-30 * time.Minute),
			Metadata: map[string]any{"ingredient": "Sea Bass Fillet", "price_increase_pct": 18, "current_cost": 4800, "item_affected": "Ceviche Clasico"},
		},
		{
			OrgID: orgID, LocationID: loc(1), Severity: SeverityWarning, Module: "inventory",
			Title:       "Aji Amarillo Paste: 3-week supply disruption predicted",
			Description: "Specialty Imports has flagged a potential 3-week delay on aji amarillo paste shipments. Current stock covers 12 days. Auto-switching 30% of orders to Metro Market backup supply at 15% higher cost.",
			Status: "active", CreatedAt: now.Add(-1 * time.Hour),
			Metadata: map[string]any{"ingredient": "Aji Amarillo Paste", "days_stock_remaining": 12, "disruption_weeks": 3},
		},
		{
			OrgID: orgID, LocationID: loc(0), Severity: SeverityWarning, Module: "inventory",
			Title:       "Beef Tenderloin usage variance +12% vs theoretical at El Gouna",
			Description: "Actual beef tenderloin consumption is 12% above recipe-calculated usage over the past 14 days. Possible portioning issue — 2.3 kg/day excess translates to ~1,500 EGP/day in waste. Review Lomo Saltado and Churrasco prep procedures.",
			Status: "active", CreatedAt: now.Add(-3 * time.Hour),
			Metadata: map[string]any{"ingredient": "Beef Tenderloin", "variance_pct": 12, "daily_excess_egp": 1500},
		},

		// ── Financial Intelligence (3) ──
		{
			OrgID: orgID, LocationID: loc(1), Severity: SeverityCritical, Module: "financial",
			Title:       "New Cairo food cost trending 34.2% (target: 32%)",
			Description: "Food cost at New Cairo has exceeded the 32% target for 8 consecutive days, driven by a 18% increase in sea bass and 15% in shrimp costs. Protein costs account for 68% of the overage. Recommend menu price review or supplier renegotiation.",
			Status: "active", CreatedAt: now.Add(-45 * time.Minute),
			Metadata: map[string]any{"food_cost_pct": 34.2, "target_pct": 32.0, "consecutive_days": 8},
		},
		{
			OrgID: orgID, LocationID: loc(3), Severity: SeverityWarning, Module: "financial",
			Title:       "North Coast labor cost at 29% — 3% above budget",
			Description: "Labor cost at North Coast reached 29% of revenue this week, 3 percentage points above the 26% target. Overtime detected for 4 staff members (avg 12 extra hours each). Review scheduling or hire 2 additional part-time staff for weekend coverage.",
			Status: "active", CreatedAt: now.Add(-2 * time.Hour),
			Metadata: map[string]any{"labor_cost_pct": 29.0, "budget_pct": 26.0, "overtime_staff_count": 4},
		},
		{
			OrgID: orgID, LocationID: loc(2), Severity: SeverityWarning, Module: "financial",
			Title:       "Delivery channel margin at Zayed dropped to 18%",
			Description: "Sheikh Zayed delivery channel gross margin fell to 18% (from 24% last month). Platform commission increased from 22% to 28% on the top delivery app. Consider renegotiating terms, adjusting delivery menu prices, or shifting volume to direct ordering.",
			Status: "active", CreatedAt: now.Add(-5 * time.Hour),
			Metadata: map[string]any{"channel": "delivery", "margin_pct": 18.0, "prev_margin_pct": 24.0, "commission_pct": 28.0},
		},

		// ── Customer Intelligence (2) ──
		{
			OrgID: orgID, LocationID: loc(0), Severity: SeverityCritical, Module: "customer",
			Title:       "5 high-CLV guests haven't visited in 21+ days — churn risk: HIGH",
			Description: "5 guests with average monthly spend of 2,400 EGP (total CLV: 12,000 EGP/month) have not visited any Chicha location in 21+ days. Their visit frequency was previously bi-weekly. Recommend personalized re-engagement: complimentary Pisco Sour or Ceviche tasting invitation.",
			Status: "active", CreatedAt: now.Add(-1 * time.Hour),
			Metadata: map[string]any{"guest_count": 5, "avg_monthly_spend_egp": 2400, "days_since_visit": 21, "segment": "champion"},
		},
		{
			OrgID: orgID, LocationID: loc(0), Severity: SeverityInfo, Module: "customer",
			Title:       "Weekend dinner covers down 8% vs last month at El Gouna",
			Description: "Friday and Saturday dinner covers at El Gouna decreased 8% compared to the same period last month (avg 85 vs 92 covers). This deviates from the seasonal uptick pattern. Possible causes: competitor opening nearby, weather pattern change, or pricing sensitivity.",
			Status: "active", CreatedAt: now.Add(-8 * time.Hour),
			Metadata: map[string]any{"cover_change_pct": -8, "current_avg": 85, "previous_avg": 92},
		},

		// ── Operations Intelligence (2) ──
		{
			OrgID: orgID, LocationID: loc(1), Severity: SeverityWarning, Module: "operations",
			Title:       "Kitchen capacity at New Cairo projected to exceed 90% during Friday dinner",
			Description: "Based on reservation data and historical patterns, New Cairo kitchen capacity is projected to hit 94% between 7-9 PM on Friday. Current scheduled staff can handle 88% capacity. Recommend increasing prep cook coverage by 2 staff and pre-prepping Causa Limena and Anticuchos.",
			Status: "active", CreatedAt: now.Add(-3 * time.Hour),
			Metadata: map[string]any{"projected_capacity_pct": 94, "peak_window": "19:00-21:00", "day": "Friday"},
		},
		{
			OrgID: orgID, LocationID: loc(3), Severity: SeverityCritical, Module: "operations",
			Title:       "Average ticket time at North Coast increased from 12 to 18 minutes",
			Description: "Average kitchen ticket time at North Coast has increased 50% over the past week (12 min → 18 min). Analysis shows Ceviche Bar is the bottleneck — 73% of delayed tickets include ceviche items. Root cause: new cook on ceviche station with 2.5 ELU rating (team avg: 4.0).",
			Status: "active", CreatedAt: now.Add(-90 * time.Minute),
			Metadata: map[string]any{"avg_ticket_min": 18, "prev_avg_ticket_min": 12, "bottleneck_station": "Ceviche Bar"},
		},

		// ── Menu Intelligence (2) ──
		{
			OrgID: orgID, LocationID: loc(2), Severity: SeverityInfo, Module: "menu",
			Title:       "Churrasco Chimichurri classified as 'complex_star'",
			Description: "Churrasco Chimichurri has the highest margin (68%) in the menu but its 8-minute grill time limits throughput to 6 units/hour at Sheikh Zayed. During peak hours, it blocks 25% of grill capacity. Consider adding a dedicated churrasco grill or offering a pre-seared option.",
			Status: "active", CreatedAt: now.Add(-4 * time.Hour),
			Metadata: map[string]any{"item": "Churrasco Chimichurri", "margin_pct": 68, "grill_time_min": 8, "classification": "complex_star"},
		},
		{
			OrgID: orgID, LocationID: loc(1), Severity: SeverityInfo, Module: "menu",
			Title:       "Empanadas reclassified from 'workhorse' to 'crowd_pleaser'",
			Description: "Empanadas (3 pcs) sales velocity at New Cairo increased 22% after the portion adjustment last month. Now selling 45 units/day (was 37). Classification upgraded from workhorse to crowd_pleaser. Margin remains healthy at 62%. Consider featuring as a starter special.",
			Status: "active", CreatedAt: now.Add(-7 * time.Hour),
			Metadata: map[string]any{"item": "Empanadas (3 pcs)", "velocity_increase_pct": 22, "daily_units": 45, "new_classification": "crowd_pleaser"},
		},

		// ── Vendor Intelligence (2) ──
		{
			OrgID: orgID, LocationID: loc(3), Severity: SeverityWarning, Module: "vendor",
			Title:       "Metro Market OTIF rate dropped to 72% — recommend shifting orders",
			Description: "Metro Market's on-time-in-full rate at North Coast has dropped from 85% to 72% over the past month. 3 of the last 5 deliveries arrived late or short. Recommend shifting 30% of produce orders to Seoudi Fresh (OTIF: 90%) while monitoring Metro Market performance.",
			Status: "active", CreatedAt: now.Add(-5 * time.Hour),
			Metadata: map[string]any{"vendor": "Metro Market", "otif_rate": 72.0, "prev_otif_rate": 85.0, "recommended_vendor": "Seoudi Fresh"},
		},
		{
			OrgID: orgID, LocationID: loc(0), Severity: SeverityInfo, Module: "vendor",
			Title:       "Specialty Imports: Pisco price forecast shows 15% increase in Q2",
			Description: "Based on 6-month trend analysis, Pisco cost from Specialty Imports is projected to increase 15% in Q2 2026 (from 200 to 230 piasters/oz). At current Pisco Sour volume (35 units/day across all branches), this adds ~52,500 EGP/month in costs. Consider forward purchasing 3-month supply or sourcing alternative brands.",
			Status: "active", CreatedAt: now.Add(-10 * time.Hour),
			Metadata: map[string]any{"vendor": "Specialty Imports", "ingredient": "Pisco", "forecast_increase_pct": 15, "monthly_impact_egp": 52500},
		},

		// ── Additional Financial/Ops (3) ──
		{
			OrgID: orgID, LocationID: loc(2), Severity: SeverityInfo, Module: "financial",
			Title:       "Sheikh Zayed outperforming chain average by 12% on food cost",
			Description: "Sheikh Zayed is running a 30.8% food cost vs the chain average of 32.6%. Key driver: demand-based ceviche prep scheduling reduces fish waste by 22%. This best practice has been flagged for rollout to other locations.",
			Status: "active", CreatedAt: now.Add(-12 * time.Hour),
			Metadata: map[string]any{"food_cost_pct": 30.8, "chain_avg_pct": 32.6, "best_practice": "ceviche_prep_scheduling"},
		},
		{
			OrgID: orgID, LocationID: loc(0), Severity: SeverityWarning, Module: "inventory",
			Title:       "Avocado waste rate at El Gouna is 2.3x chain average",
			Description: "El Gouna discarded 23 avocados in the past 14 days (chain average: 10). Primary cause: overripe at delivery (65%) and over-prepping for Causa Limena (35%). Recommend switching to 3x/week delivery for avocados and implementing FIFO labels.",
			Status: "active", CreatedAt: now.Add(-6 * time.Hour),
			Metadata: map[string]any{"ingredient": "Avocado", "waste_count": 23, "chain_avg": 10, "multiplier": 2.3},
		},
		{
			OrgID: orgID, LocationID: loc(3), Severity: SeverityInfo, Module: "financial",
			Title:       "Pisco Sour is the highest-margin beverage across all locations",
			Description: "Pisco Sour generates 72% margin at 185 EGP and sells 35+ units/day chain-wide. The upcoming 'Pisco Hour' campaign at El Gouna saw 85 redemptions in 2 weeks. Recommend expanding to all 4 locations during the 5-7 PM window.",
			Status: "active", CreatedAt: now.Add(-9 * time.Hour),
			Metadata: map[string]any{"item": "Pisco Sour", "margin_pct": 72, "daily_chain_units": 35, "campaign": "Pisco Hour"},
		},
	}

	for _, a := range demos {
		s.enqueue(a)
	}
	slog.Info("demo alerts seeded", "count", len(demos), "org_id", orgID)
}

func formatAlertID(seq int64) string {
	return "alert-" + time.Now().Format("20060102") + "-" + itoa(seq)
}

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}

func sortAlerts(alerts []Alert) {
	sevOrder := map[Severity]int{
		SeverityCritical: 0,
		SeverityWarning:  1,
		SeverityInfo:     2,
	}
	for i := 1; i < len(alerts); i++ {
		for j := i; j > 0 && sevOrder[alerts[j].Severity] < sevOrder[alerts[j-1].Severity]; j-- {
			alerts[j], alerts[j-1] = alerts[j-1], alerts[j]
		}
	}
}
