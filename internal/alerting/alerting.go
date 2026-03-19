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
func (s *Service) SeedAlerts(orgID string, locations []string) {
	now := time.Now()
	demos := []Alert{
		{
			OrgID:       orgID,
			LocationID:  locations[0],
			Severity:    SeverityCritical,
			Title:       "Ground Beef stock critically low",
			Description: "Current inventory is 8 lbs — PAR level is 50 lbs. Projected to run out before tomorrow lunch service. Reorder immediately from US Foods.",
			Module:      "inventory",
			Status:      "active",
			CreatedAt:   now.Add(-25 * time.Minute),
			Metadata:    map[string]any{"ingredient": "Ground Beef (80/20)", "current_level": 8, "par_level": 50},
		},
		{
			OrgID:       orgID,
			LocationID:  locations[0],
			Severity:    SeverityCritical,
			Title:       "Delivery COGS margin below 30%",
			Description: "Delivery channel gross margin dropped to 27.3% over the last 7 days — below the 30% minimum threshold. Review delivery pricing or renegotiate platform commission.",
			Module:      "financial",
			Status:      "active",
			CreatedAt:   now.Add(-2 * time.Hour),
			Metadata:    map[string]any{"channel": "delivery", "margin_pct": 27.3, "threshold": 30.0},
		},
		{
			OrgID:       orgID,
			LocationID:  locations[0],
			Severity:    SeverityWarning,
			Title:       "Romaine Lettuce approaching reorder point",
			Description: "Current stock is 12 heads — reorder point is 10 heads. Place order within 24 hours to avoid stockout.",
			Module:      "inventory",
			Status:      "active",
			CreatedAt:   now.Add(-1 * time.Hour),
			Metadata:    map[string]any{"ingredient": "Romaine Lettuce", "current_level": 12, "reorder_point": 10},
		},
		{
			OrgID:       orgID,
			LocationID:  locations[0],
			Severity:    SeverityWarning,
			Title:       "Classic Burger food cost up 12% this week",
			Description: "Ground beef cost per unit increased from $4.40 to $4.80/lb. Contribution margin on Classic Burger dropped from 78% to 72%. Consider supplier alternatives or price adjustment.",
			Module:      "financial",
			Status:      "active",
			CreatedAt:   now.Add(-3 * time.Hour),
			Metadata:    map[string]any{"item": "Classic Burger", "old_cost": 440, "new_cost": 480},
		},
		{
			OrgID:       orgID,
			LocationID:  locations[0],
			Severity:    SeverityInfo,
			Title:       "Weekly COGS report ready",
			Description: "The automated cost-of-goods-sold report for the week ending March 16 has been generated and is ready for review.",
			Module:      "financial",
			Status:      "active",
			CreatedAt:   now.Add(-5 * time.Hour),
			Metadata:    map[string]any{"report_type": "weekly_cogs", "week_ending": "2026-03-16"},
		},
		{
			OrgID:       orgID,
			LocationID:  locations[1],
			Severity:    SeverityCritical,
			Title:       "Toast POS sync failed",
			Description: "The Toast adapter for Airport Terminal 4 has not synced in 45 minutes. Orders placed after 2:15 PM may not be reflected in reports. Check network connectivity.",
			Module:      "adapter",
			Status:      "active",
			CreatedAt:   now.Add(-45 * time.Minute),
			Metadata:    map[string]any{"adapter": "toast", "last_sync": now.Add(-45 * time.Minute).Format(time.RFC3339)},
		},
		{
			OrgID:       orgID,
			LocationID:  locations[1],
			Severity:    SeverityWarning,
			Title:       "Bacon Avocado Burger underperforming at Airport",
			Description: "Sales volume dropped 35% over the last 14 days at Airport Terminal 4. Only 3 units sold today vs 7-day average of 8 units. Consider repositioning on menu or running a promotion.",
			Module:      "financial",
			Status:      "active",
			CreatedAt:   now.Add(-4 * time.Hour),
			Metadata:    map[string]any{"item": "Bacon Avocado Burger", "current_daily": 3, "avg_daily": 8},
		},
		{
			OrgID:       orgID,
			LocationID:  locations[1],
			Severity:    SeverityInfo,
			Title:       "New menu item performance check",
			Description: "Loaded Fries was added to Airport Terminal 4 menu 14 days ago. Trending at 6 units/day — 15% above projected volume. Consider permanent menu placement.",
			Module:      "financial",
			Status:      "active",
			CreatedAt:   now.Add(-6 * time.Hour),
			Metadata:    map[string]any{"item": "Loaded Fries", "days_since_added": 14, "daily_units": 6},
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
