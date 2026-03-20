package onboarding

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/opsnerve/fireline/internal/tenant"
	"github.com/opsnerve/fireline/pkg/database"
)

// OnboardingSession represents a wizard session for an org.
type OnboardingSession struct {
	SessionID     string         `json:"session_id"`
	CurrentStep   string         `json:"current_step"`
	ProfileData   map[string]any `json:"profile_data"`
	ConceptType   *string        `json:"concept_type"`
	Priorities    []string       `json:"priorities"`
	ActiveModules []string       `json:"active_modules"`
	InsightsData  map[string]any `json:"insights_data"`
	CompletedAt   *time.Time     `json:"completed_at"`
}

// ChecklistItem is a single personalised onboarding todo.
type ChecklistItem struct {
	ItemID      string     `json:"item_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Category    string     `json:"category"`
	Priority    int        `json:"priority"`
	Completed   bool       `json:"completed"`
	CompletedAt *time.Time `json:"completed_at"`
}

// FirstInsights holds KPIs derived from the first data import.
type FirstInsights struct {
	DailyRevenueAvg int64    `json:"daily_revenue_avg"`
	TopSellers      []string `json:"top_sellers"`
	PeakHour        int      `json:"peak_hour"`
	AvgCheck        int64    `json:"avg_check"`
	VoidRate        float64  `json:"void_rate"`
	StaffCount      int      `json:"staff_count"`
	CheckCount      int      `json:"check_count"`
}

// ─── DB methods ─────────────────────────────────────────────────────────────

// StartOnboarding inserts a new onboarding session for the org/user.
func (s *Service) StartOnboarding(ctx context.Context, orgID, userID string) (*OnboardingSession, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var sess OnboardingSession
	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, `
			INSERT INTO onboarding_sessions (org_id, user_id)
			VALUES ($1, $2)
			RETURNING session_id, current_step, profile_data, concept_type,
			          priorities, active_modules, insights_data, completed_at`,
			orgID, userID)
		return scanSession(row, &sess)
	})
	if err != nil {
		return nil, fmt.Errorf("start onboarding: %w", err)
	}
	return &sess, nil
}

// GetSession returns the most recent onboarding session for the org.
func (s *Service) GetSession(ctx context.Context, orgID string) (*OnboardingSession, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var sess OnboardingSession
	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, `
			SELECT session_id, current_step, profile_data, concept_type,
			       priorities, active_modules, insights_data, completed_at
			FROM onboarding_sessions
			WHERE org_id = $1
			ORDER BY created_at DESC
			LIMIT 1`,
			orgID)
		return scanSession(row, &sess)
	})
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}
	return &sess, nil
}

// UpdateStep advances the wizard step and merges step data into the session.
func (s *Service) UpdateStep(ctx context.Context, orgID, sessionID, step string, data map[string]any) (*OnboardingSession, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshal step data: %w", err)
	}

	var sess OnboardingSession
	err = database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		// Merge data into the appropriate column based on step.
		var q string
		switch step {
		case "profile":
			q = `UPDATE onboarding_sessions
				 SET current_step = $3, profile_data = profile_data || $4::jsonb, updated_at = now()
				 WHERE session_id = $1 AND org_id = $2
				 RETURNING session_id, current_step, profile_data, concept_type,
				           priorities, active_modules, insights_data, completed_at`
		case "concept_type":
			q = `UPDATE onboarding_sessions
				 SET current_step = $3, concept_type = ($4::jsonb->>'concept_type'), updated_at = now()
				 WHERE session_id = $1 AND org_id = $2
				 RETURNING session_id, current_step, profile_data, concept_type,
				           priorities, active_modules, insights_data, completed_at`
		case "priorities":
			q = `UPDATE onboarding_sessions
				 SET current_step = $3,
				     priorities = ARRAY(SELECT jsonb_array_elements_text($4::jsonb->'priorities')),
				     updated_at = now()
				 WHERE session_id = $1 AND org_id = $2
				 RETURNING session_id, current_step, profile_data, concept_type,
				           priorities, active_modules, insights_data, completed_at`
		case "modules":
			q = `UPDATE onboarding_sessions
				 SET current_step = $3,
				     active_modules = ARRAY(SELECT jsonb_array_elements_text($4::jsonb->'modules')),
				     updated_at = now()
				 WHERE session_id = $1 AND org_id = $2
				 RETURNING session_id, current_step, profile_data, concept_type,
				           priorities, active_modules, insights_data, completed_at`
		case "first_insights":
			q = `UPDATE onboarding_sessions
				 SET current_step = $3, insights_data = $4::jsonb, updated_at = now()
				 WHERE session_id = $1 AND org_id = $2
				 RETURNING session_id, current_step, profile_data, concept_type,
				           priorities, active_modules, insights_data, completed_at`
		case "complete":
			q = `UPDATE onboarding_sessions
				 SET current_step = $3, completed_at = now(), updated_at = now()
				 WHERE session_id = $1 AND org_id = $2
				 RETURNING session_id, current_step, profile_data, concept_type,
				           priorities, active_modules, insights_data, completed_at`
		default:
			q = `UPDATE onboarding_sessions
				 SET current_step = $3, updated_at = now()
				 WHERE session_id = $1 AND org_id = $2
				 RETURNING session_id, current_step, profile_data, concept_type,
				           priorities, active_modules, insights_data, completed_at`
		}
		row := tx.QueryRow(ctx, q, sessionID, orgID, step, string(dataJSON))
		return scanSession(row, &sess)
	})
	if err != nil {
		return nil, fmt.Errorf("update step: %w", err)
	}
	return &sess, nil
}

// InferConceptType queries avg_check from checks and classifies the concept.
func (s *Service) InferConceptType(ctx context.Context, orgID, locationID string) (string, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var avgCheck int64
	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `
			SELECT COALESCE(AVG(total_cents), 0)::BIGINT
			FROM checks
			WHERE org_id = $1 AND location_id = $2
			  AND voided = false`,
			orgID, locationID).Scan(&avgCheck)
	})
	if err != nil {
		// Fallback to casual if no data yet
		return "casual_dining", nil
	}
	return inferConceptFromAvgCheck(avgCheck), nil
}

// GenerateFirstInsights queries checks for KPI data and returns insights.
func (s *Service) GenerateFirstInsights(ctx context.Context, orgID, locationID string) (*FirstInsights, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	insights := &FirstInsights{}
	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		// Daily revenue avg and avg check
		err := tx.QueryRow(ctx, `
			SELECT
				COALESCE(AVG(daily_rev), 0)::BIGINT,
				COALESCE(AVG(avg_check), 0)::BIGINT,
				COALESCE(SUM(cnt), 0)::INT
			FROM (
				SELECT
					DATE(opened_at) AS day,
					SUM(total_cents) AS daily_rev,
					AVG(total_cents) AS avg_check,
					COUNT(*) AS cnt
				FROM checks
				WHERE org_id = $1 AND location_id = $2 AND voided = false
				GROUP BY DATE(opened_at)
			) d`,
			orgID, locationID,
		).Scan(&insights.DailyRevenueAvg, &insights.AvgCheck, &insights.CheckCount)
		if err != nil {
			return err
		}

		// Peak hour
		err = tx.QueryRow(ctx, `
			SELECT COALESCE(EXTRACT(HOUR FROM opened_at)::INT, 12)
			FROM checks
			WHERE org_id = $1 AND location_id = $2 AND voided = false
			GROUP BY EXTRACT(HOUR FROM opened_at)
			ORDER BY COUNT(*) DESC
			LIMIT 1`,
			orgID, locationID,
		).Scan(&insights.PeakHour)
		if err != nil {
			insights.PeakHour = 12 // default noon
		}

		// Void rate
		var totalChecks, voidedChecks int64
		err = tx.QueryRow(ctx, `
			SELECT
				COUNT(*),
				COUNT(*) FILTER (WHERE voided = true)
			FROM checks
			WHERE org_id = $1 AND location_id = $2`,
			orgID, locationID,
		).Scan(&totalChecks, &voidedChecks)
		if err == nil && totalChecks > 0 {
			insights.VoidRate = float64(voidedChecks) / float64(totalChecks) * 100
		}

		// Top sellers (menu item names by total quantity)
		rows, err := tx.Query(ctx, `
			SELECT mi.name
			FROM check_items ci
			JOIN menu_items mi ON mi.menu_item_id = ci.menu_item_id
			JOIN checks c ON c.check_id = ci.check_id
			WHERE c.org_id = $1 AND c.location_id = $2 AND c.voided = false
			GROUP BY mi.name
			ORDER BY SUM(ci.quantity) DESC
			LIMIT 5`,
			orgID, locationID)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var name string
				if scanErr := rows.Scan(&name); scanErr == nil {
					insights.TopSellers = append(insights.TopSellers, name)
				}
			}
		}

		// Staff count (distinct employees with shifts)
		tx.QueryRow(ctx, `
			SELECT COUNT(DISTINCT employee_id)
			FROM shifts
			WHERE org_id = $1 AND location_id = $2`,
			orgID, locationID,
		).Scan(&insights.StaffCount)

		return nil
	})
	if err != nil {
		// Return partial insights with defaults on query failure
		return insights, nil
	}
	if insights.TopSellers == nil {
		insights.TopSellers = []string{}
	}
	return insights, nil
}

// RecommendModules maps selected priorities to module identifiers.
func (s *Service) RecommendModules(priorities []string) []string {
	return recommendModules(priorities)
}

// GenerateChecklist inserts personalised checklist items for the org.
func (s *Service) GenerateChecklist(ctx context.Context, orgID, conceptType string, modules []string) ([]ChecklistItem, error) {
	items := generateChecklistItems(conceptType, modules)
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		for _, item := range items {
			_, err := tx.Exec(ctx, `
				INSERT INTO onboarding_checklist_items
				    (org_id, title, description, category, priority)
				VALUES ($1, $2, $3, $4, $5)`,
				orgID, item.Title, item.Description, item.Category, item.Priority)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("generate checklist: %w", err)
	}
	return s.GetChecklist(ctx, orgID)
}

// GetChecklist returns all checklist items for the org ordered by priority.
func (s *Service) GetChecklist(ctx context.Context, orgID string) ([]ChecklistItem, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var items []ChecklistItem
	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, `
			SELECT item_id, title, COALESCE(description,''), category, priority, completed, completed_at
			FROM onboarding_checklist_items
			WHERE org_id = $1
			ORDER BY priority ASC, created_at ASC`,
			orgID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var it ChecklistItem
			if err := rows.Scan(&it.ItemID, &it.Title, &it.Description,
				&it.Category, &it.Priority, &it.Completed, &it.CompletedAt); err != nil {
				return err
			}
			items = append(items, it)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("get checklist: %w", err)
	}
	if items == nil {
		items = []ChecklistItem{}
	}
	return items, nil
}

// CompleteChecklistItem marks a checklist item as done.
func (s *Service) CompleteChecklistItem(ctx context.Context, orgID, itemID string) error {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	return database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `
			UPDATE onboarding_checklist_items
			SET completed = true, completed_at = now()
			WHERE item_id = $1 AND org_id = $2`,
			itemID, orgID)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return fmt.Errorf("item not found")
		}
		return nil
	})
}

// ─── Pure functions ──────────────────────────────────────────────────────────

// inferConceptFromAvgCheck classifies concept type based on average check size.
func inferConceptFromAvgCheck(avgCheckCents int64) string {
	switch {
	case avgCheckCents < 1200: // < $12
		return "quick_service"
	case avgCheckCents < 2500: // < $25
		return "fast_casual"
	case avgCheckCents < 5000: // < $50
		return "casual_dining"
	case avgCheckCents < 10000: // < $100
		return "upscale_casual"
	default:
		return "fine_dining"
	}
}

// recommendModules maps priority identifiers to FireLine module names.
func recommendModules(priorities []string) []string {
	mapping := map[string][]string{
		"reduce_waste":    {"inventory", "menu_scoring"},
		"boost_revenue":   {"financial", "marketing", "menu_scoring"},
		"labor_efficiency": {"labor", "scheduling"},
		"food_cost_control": {"inventory", "financial", "vendor"},
		"guest_experience": {"customers", "operations"},
		"growth_insights":  {"reporting", "portfolio"},
	}
	seen := map[string]bool{}
	var result []string
	for _, p := range priorities {
		for _, m := range mapping[p] {
			if !seen[m] {
				seen[m] = true
				result = append(result, m)
			}
		}
	}
	if result == nil {
		result = []string{}
	}
	return result
}

// checklistTemplate defines a checklist item template.
type checklistTemplate struct {
	Title       string
	Description string
	Category    string
	Priority    int
}

// generateChecklistItems returns a personalised set of checklist items.
func generateChecklistItems(conceptType string, modules []string) []ChecklistItem {
	// Base items every restaurant gets
	base := []checklistTemplate{
		{"Complete your restaurant profile", "Add logo, address, and contact info", "setup", 0},
		{"Invite your management team", "Add team members and assign roles", "setup", 1},
		{"Configure your locations", "Set up each location with its details", "setup", 2},
		{"Connect your POS system", "Sync your point-of-sale for live data", "integration", 3},
		{"Import your menu", "Upload or sync your full menu catalog", "menu", 4},
	}

	// Concept-specific items
	conceptItems := map[string][]checklistTemplate{
		"quick_service": {
			{"Set up combo/value meal pricing", "Configure bundled item pricing for speed", "menu", 10},
			{"Configure drive-through timing alerts", "Track service speed KPIs", "operations", 11},
		},
		"fast_casual": {
			{"Set up build-your-own menu items", "Configure customizable items", "menu", 10},
			{"Enable online ordering integration", "Connect your online order channel", "integration", 11},
		},
		"casual_dining": {
			{"Configure table management", "Set up your floor plan and table sections", "operations", 10},
			{"Set up reservation tracking", "Enable guest reservation management", "guests", 11},
		},
		"upscale_casual": {
			{"Set up tasting menu pricing tiers", "Configure premium menu categories", "menu", 10},
			{"Enable VIP guest profiles", "Track and reward high-value guests", "guests", 11},
		},
		"fine_dining": {
			{"Configure wine list and pairings", "Set up beverage program with pairing suggestions", "menu", 10},
			{"Enable chef's tasting menu management", "Configure prix-fixe and tasting menus", "menu", 11},
			{"Set up per-diner check reporting", "Track per-cover averages", "financial", 12},
		},
	}

	// Module-specific items
	moduleItems := map[string][]checklistTemplate{
		"inventory": {
			{"Set PAR levels for key ingredients", "Define min/max stock thresholds", "inventory", 20},
			{"Run your first inventory count", "Establish baseline stock levels", "inventory", 21},
		},
		"financial": {
			{"Create your first budget", "Set revenue and cost targets for the period", "financial", 20},
			{"Review your P&L baseline", "Understand your starting financial position", "financial", 21},
		},
		"labor": {
			{"Import your staff roster", "Add employees and their roles", "labor", 20},
			{"Set your labor cost targets", "Configure labor percentage goals by role", "labor", 21},
		},
		"scheduling": {
			{"Build your first schedule", "Create the upcoming week's schedule", "labor", 22},
		},
		"marketing": {
			{"Launch your first campaign", "Create a targeted guest promotion", "marketing", 20},
			{"Set up loyalty tier thresholds", "Define visit/spend goals for loyalty levels", "marketing", 21},
		},
		"vendor": {
			{"Add your top vendors", "Enter vendor contacts and terms", "vendor", 20},
			{"Create your first purchase order", "Order from your primary food vendor", "vendor", 21},
		},
		"customers": {
			{"Review top guest profiles", "Explore your most frequent guests", "guests", 20},
		},
		"menu_scoring": {
			{"Review your menu engineering report", "See which items are stars vs. dogs", "menu", 20},
		},
		"reporting": {
			{"Schedule your first weekly report", "Set up automated performance summaries", "reporting", 20},
		},
		"portfolio": {
			{"Compare location performance", "Review KPIs across all your locations", "reporting", 20},
		},
		"operations": {
			{"Configure kitchen display routing", "Set up KDS ticket routing rules", "operations", 20},
		},
	}

	var templates []checklistTemplate
	templates = append(templates, base...)
	if ct, ok := conceptItems[conceptType]; ok {
		templates = append(templates, ct...)
	}
	moduleSet := map[string]bool{}
	for _, m := range modules {
		moduleSet[m] = true
	}
	for _, m := range modules {
		if mt, ok := moduleItems[m]; ok {
			for _, t := range mt {
				if !moduleSet[t.Category+"_"+t.Title] {
					templates = append(templates, t)
				}
			}
		}
	}

	items := make([]ChecklistItem, len(templates))
	for i, t := range templates {
		items[i] = ChecklistItem{
			Title:       t.Title,
			Description: t.Description,
			Category:    t.Category,
			Priority:    t.Priority,
		}
	}
	return items
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

type rowScanner interface {
	Scan(dest ...any) error
}

func scanSession(row rowScanner, s *OnboardingSession) error {
	var profileRaw, insightsRaw []byte
	var priorities, modules []string
	err := row.Scan(
		&s.SessionID,
		&s.CurrentStep,
		&profileRaw,
		&s.ConceptType,
		&priorities,
		&modules,
		&insightsRaw,
		&s.CompletedAt,
	)
	if err != nil {
		return err
	}
	if len(profileRaw) > 0 {
		_ = json.Unmarshal(profileRaw, &s.ProfileData)
	}
	if s.ProfileData == nil {
		s.ProfileData = map[string]any{}
	}
	if len(insightsRaw) > 0 {
		_ = json.Unmarshal(insightsRaw, &s.InsightsData)
	}
	if s.InsightsData == nil {
		s.InsightsData = map[string]any{}
	}
	s.Priorities = priorities
	if s.Priorities == nil {
		s.Priorities = []string{}
	}
	s.ActiveModules = modules
	if s.ActiveModules == nil {
		s.ActiveModules = []string{}
	}
	return nil
}
