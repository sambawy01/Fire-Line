package menu

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/opsnerve/fireline/internal/event"
	"github.com/opsnerve/fireline/internal/tenant"
	"github.com/opsnerve/fireline/pkg/database"
)

// MenuItemScore holds the 5-dimension score and classification for a menu item.
type MenuItemScore struct {
	MenuItemID         string  `json:"menu_item_id"`
	Name               string  `json:"name"`
	Category           string  `json:"category"`
	Price              int64   `json:"price"`
	MarginScore        float64 `json:"margin_score"`
	VelocityScore      float64 `json:"velocity_score"`
	ComplexityScore    float64 `json:"complexity_score"`
	SatisfactionScore  float64 `json:"satisfaction_score"`
	StrategicScore     float64 `json:"strategic_score"`
	Classification     string  `json:"classification"`
	ContributionMargin int64   `json:"contribution_margin"`
	UnitsSold          int     `json:"units_sold"`
}

// classifyItem returns one of 8 classification strings based on 5-dimension scores.
//
// Thresholds (derived from spec test cases):
//   - highMargin:     margin >= 50
//   - highVelocity:   velocity >= 60  (clearly above midpoint — avoids ambiguity at 50)
//   - highComplexity: complexity >= 70 (complexity is a burden marker)
//   - midComplexity:  complexity >= 50 (used to split declining_star from complex_star)
//   - highSat:        satisfaction > 50 (strict — 50 is neutral, not high)
//   - highStrategic:  strategic >= 70
func classifyItem(margin, velocity, complexity, satisfaction, strategic float64) string {
	highMargin := margin >= 50
	highVelocity := velocity >= 60
	highComplexity := complexity >= 70
	midComplexity := complexity >= 50
	highSatisfaction := satisfaction > 50
	highStrategic := strategic >= 70

	switch {
	case highStrategic && !highMargin && !highVelocity:
		// Kept by management intent despite weak economics.
		return "strategic_anchor"
	case highMargin && highVelocity && highComplexity:
		// High margin + popular but operationally demanding.
		return "workhorse"
	case highMargin && highVelocity:
		// Best-in-class: high margin, popular, manageable complexity.
		return "powerhouse"
	case !highMargin && highVelocity:
		// Popular but thin margin — drives volume, not profit.
		return "crowd_pleaser"
	case highMargin && !highVelocity && highSatisfaction:
		// High margin, low velocity, but guests love it when they order it.
		return "hidden_gem"
	case highMargin && !highVelocity && midComplexity:
		// High margin + low velocity + medium-to-high complexity → fading star.
		return "declining_star"
	case highMargin && !highVelocity:
		// High margin, low velocity, low complexity, neutral satisfaction.
		return "complex_star"
	default:
		return "underperformer"
	}
}

// normalizeScore scales value relative to maxValue into [0,100], capped at 100.
// Returns 0 if maxValue is 0 to avoid division by zero.
func normalizeScore(value, maxValue float64) float64 {
	if maxValue == 0 {
		return 0
	}
	score := value / maxValue * 100
	if score > 100 {
		return 100
	}
	return score
}

// scoringRow holds raw metrics gathered for a single menu item before normalizing.
type scoringRow struct {
	menuItemID         string
	name               string
	category           string
	price              int64
	cogs               int64 // food cost in cents
	unitsSold          int
	taskDurationSecs   int
	voidRate           float64
	strategicScore     float64 // persisted manual override, kept as-is
	prevClassification string
}

// ScoreMenuItems computes 5-dimension scores for all active menu items at a location,
// persists results to menu_items, and emits classification-change events when the
// classification differs from the previous stored value.
func (s *Service) ScoreMenuItems(ctx context.Context, orgID, locationID string) ([]MenuItemScore, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	cutoff := time.Now().AddDate(0, 0, -30)

	var rows []scoringRow

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		// --- Query 1: base item data + COGS ---
		baseRows, err := tx.Query(tenantCtx,
			`SELECT
			    mi.menu_item_id,
			    mi.name,
			    mi.category,
			    mi.price,
			    COALESCE(SUM(re.quantity_per_unit * COALESCE(ilc.local_cost_per_unit, i.cost_per_unit)), 0)::BIGINT AS cogs,
			    COALESCE(mi.strategic_score, 50)::FLOAT,
			    COALESCE(mi.classification, '') AS prev_classification
			 FROM menu_items mi
			 LEFT JOIN recipe_explosion re ON re.menu_item_id = mi.menu_item_id AND re.org_id = mi.org_id
			 LEFT JOIN ingredients i ON i.ingredient_id = re.ingredient_id AND i.org_id = mi.org_id
			 LEFT JOIN ingredient_location_configs ilc
			     ON ilc.ingredient_id = i.ingredient_id
			     AND ilc.location_id = mi.location_id
			     AND ilc.org_id = mi.org_id
			 WHERE mi.location_id = $1 AND mi.available = true
			 GROUP BY mi.menu_item_id, mi.name, mi.category, mi.price, mi.strategic_score, mi.classification
			 ORDER BY mi.name`,
			locationID,
		)
		if err != nil {
			return fmt.Errorf("query base items: %w", err)
		}
		defer baseRows.Close()

		itemIndex := make(map[string]int)
		for baseRows.Next() {
			var r scoringRow
			if err := baseRows.Scan(
				&r.menuItemID, &r.name, &r.category, &r.price,
				&r.cogs, &r.strategicScore, &r.prevClassification,
			); err != nil {
				return fmt.Errorf("scan base row: %w", err)
			}
			itemIndex[r.menuItemID] = len(rows)
			rows = append(rows, r)
		}
		if err := baseRows.Err(); err != nil {
			return fmt.Errorf("iterate base rows: %w", err)
		}

		// --- Query 2: units sold in last 30 days ---
		salesRows, err := tx.Query(tenantCtx,
			`SELECT ci.menu_item_id, COALESCE(SUM(ci.quantity), 0)::INT AS units_sold
			 FROM check_items ci
			 JOIN checks c ON c.check_id = ci.check_id AND c.org_id = ci.org_id
			 WHERE c.location_id = $1
			   AND c.status = 'closed'
			   AND ci.voided_at IS NULL
			   AND ci.menu_item_id IS NOT NULL
			   AND c.closed_at >= $2
			 GROUP BY ci.menu_item_id`,
			locationID, cutoff,
		)
		if err != nil {
			return fmt.Errorf("query sales: %w", err)
		}
		defer salesRows.Close()

		for salesRows.Next() {
			var menuItemID string
			var units int
			if err := salesRows.Scan(&menuItemID, &units); err != nil {
				return fmt.Errorf("scan sales row: %w", err)
			}
			if idx, ok := itemIndex[menuItemID]; ok {
				rows[idx].unitsSold = units
			}
		}
		if err := salesRows.Err(); err != nil {
			return fmt.Errorf("iterate sales rows: %w", err)
		}

		// --- Query 3: task duration from resource profiles ---
		profileRows, err := tx.Query(tenantCtx,
			`SELECT menu_item_id, COALESCE(SUM(duration_secs), 0)::INT AS total_duration
			 FROM menu_item_resource_profiles
			 WHERE org_id = current_setting('app.current_org_id')::UUID
			 GROUP BY menu_item_id`,
		)
		if err != nil {
			return fmt.Errorf("query resource profiles: %w", err)
		}
		defer profileRows.Close()

		for profileRows.Next() {
			var menuItemID string
			var dur int
			if err := profileRows.Scan(&menuItemID, &dur); err != nil {
				return fmt.Errorf("scan profile row: %w", err)
			}
			if idx, ok := itemIndex[menuItemID]; ok {
				rows[idx].taskDurationSecs = dur
			}
		}
		if err := profileRows.Err(); err != nil {
			return fmt.Errorf("iterate profile rows: %w", err)
		}

		// --- Query 4: void rate per item ---
		voidRows, err := tx.Query(tenantCtx,
			`SELECT
			    ci.menu_item_id,
			    COUNT(*) FILTER (WHERE ci.voided_at IS NOT NULL)::FLOAT /
			        NULLIF(COUNT(*), 0) AS void_rate
			 FROM check_items ci
			 JOIN checks c ON c.check_id = ci.check_id AND c.org_id = ci.org_id
			 WHERE c.location_id = $1
			   AND c.status = 'closed'
			   AND ci.menu_item_id IS NOT NULL
			   AND c.closed_at >= $2
			 GROUP BY ci.menu_item_id`,
			locationID, cutoff,
		)
		if err != nil {
			return fmt.Errorf("query void rate: %w", err)
		}
		defer voidRows.Close()

		for voidRows.Next() {
			var menuItemID string
			var voidRate float64
			if err := voidRows.Scan(&menuItemID, &voidRate); err != nil {
				return fmt.Errorf("scan void row: %w", err)
			}
			if idx, ok := itemIndex[menuItemID]; ok {
				rows[idx].voidRate = voidRate
			}
		}
		if err := voidRows.Err(); err != nil {
			return fmt.Errorf("iterate void rows: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	if len(rows) == 0 {
		return []MenuItemScore{}, nil
	}

	// --- Normalize each dimension relative to max in set ---
	var maxMargin, maxVelocity, maxDuration float64
	for i := range rows {
		if rows[i].taskDurationSecs == 0 {
			rows[i].taskDurationSecs = 300 // default 5 minutes
		}
		cm := float64(rows[i].price - rows[i].cogs)
		if cm > maxMargin {
			maxMargin = cm
		}
		v := float64(rows[i].unitsSold)
		if v > maxVelocity {
			maxVelocity = v
		}
		d := float64(rows[i].taskDurationSecs)
		if d > maxDuration {
			maxDuration = d
		}
	}

	scores := make([]MenuItemScore, len(rows))
	for i, r := range rows {
		cm := float64(r.price - r.cogs)
		marginScore := normalizeScore(cm, maxMargin)
		velocityScore := normalizeScore(float64(r.unitsSold), maxVelocity)
		// complexity: higher duration = higher complexity score
		complexityScore := normalizeScore(float64(r.taskDurationSecs), maxDuration)
		// satisfaction: inverse of void rate (0 voids = 100 satisfaction)
		satisfactionScore := (1.0 - r.voidRate) * 100
		if satisfactionScore < 0 {
			satisfactionScore = 0
		}
		strategicScore := r.strategicScore

		classification := classifyItem(marginScore, velocityScore, complexityScore, satisfactionScore, strategicScore)

		scores[i] = MenuItemScore{
			MenuItemID:         r.menuItemID,
			Name:               r.name,
			Category:           r.category,
			Price:              r.price,
			MarginScore:        marginScore,
			VelocityScore:      velocityScore,
			ComplexityScore:    complexityScore,
			SatisfactionScore:  satisfactionScore,
			StrategicScore:     strategicScore,
			Classification:     classification,
			ContributionMargin: r.price - r.cogs,
			UnitsSold:          r.unitsSold,
		}
	}

	// --- Persist scores and emit classification-change events ---
	err = database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		for i, sc := range scores {
			r := rows[i]
			classificationChanged := r.prevClassification != "" && r.prevClassification != sc.Classification
			var changedAt *time.Time
			if classificationChanged {
				now := time.Now()
				changedAt = &now
			}

			_, err := tx.Exec(tenantCtx,
				`UPDATE menu_items SET
				    margin_score          = $2,
				    velocity_score        = $3,
				    complexity_score      = $4,
				    satisfaction_score    = $5,
				    classification        = $6,
				    classification_changed_at = CASE WHEN $7 THEN now() ELSE classification_changed_at END,
				    updated_at            = now()
				 WHERE menu_item_id = $1`,
				sc.MenuItemID,
				sc.MarginScore,
				sc.VelocityScore,
				sc.ComplexityScore,
				sc.SatisfactionScore,
				sc.Classification,
				changedAt != nil,
			)
			if err != nil {
				return fmt.Errorf("update scores for %s: %w", sc.MenuItemID, err)
			}

			if classificationChanged && s.bus != nil {
				s.bus.Publish(ctx, event.Envelope{
					EventType:  "menu.classification.changed",
					OrgID:      orgID,
					LocationID: locationID,
					Source:     "menu.scoring",
					Payload: map[string]any{
						"menu_item_id":     sc.MenuItemID,
						"name":             sc.Name,
						"old_class":        r.prevClassification,
						"new_class":        sc.Classification,
						"margin_score":     sc.MarginScore,
						"velocity_score":   sc.VelocityScore,
					},
				})
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return scores, nil
}

// GetMenuItemScores returns current scores for all active menu items at a location.
func (s *Service) GetMenuItemScores(ctx context.Context, orgID, locationID string) ([]MenuItemScore, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	var scores []MenuItemScore

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(tenantCtx,
			`SELECT
			    mi.menu_item_id,
			    mi.name,
			    mi.category,
			    mi.price,
			    mi.margin_score,
			    mi.velocity_score,
			    mi.complexity_score,
			    mi.satisfaction_score,
			    mi.strategic_score,
			    COALESCE(mi.classification, '') AS classification
			 FROM menu_items mi
			 WHERE mi.location_id = $1 AND mi.available = true
			 ORDER BY mi.name`,
			locationID,
		)
		if err != nil {
			return fmt.Errorf("query scores: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var sc MenuItemScore
			if err := rows.Scan(
				&sc.MenuItemID, &sc.Name, &sc.Category, &sc.Price,
				&sc.MarginScore, &sc.VelocityScore, &sc.ComplexityScore,
				&sc.SatisfactionScore, &sc.StrategicScore, &sc.Classification,
			); err != nil {
				return fmt.Errorf("scan score row: %w", err)
			}
			scores = append(scores, sc)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	if scores == nil {
		scores = []MenuItemScore{}
	}
	return scores, nil
}

// SetStrategicScore sets a manual strategic override score for a menu item.
// score must be in [0, 100].
func (s *Service) SetStrategicScore(ctx context.Context, orgID, menuItemID string, score float64) error {
	if score < 0 || score > 100 {
		return fmt.Errorf("strategic score must be between 0 and 100, got %.2f", score)
	}
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	return database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		tag, err := tx.Exec(tenantCtx,
			`UPDATE menu_items SET strategic_score = $2, updated_at = now()
			 WHERE menu_item_id = $1`,
			menuItemID, score,
		)
		if err != nil {
			return fmt.Errorf("set strategic score: %w", err)
		}
		if tag.RowsAffected() == 0 {
			return fmt.Errorf("menu item %s not found", menuItemID)
		}
		return nil
	})
}
