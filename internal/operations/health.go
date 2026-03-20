package operations

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/opsnerve/fireline/internal/tenant"
	"github.com/opsnerve/fireline/pkg/database"
)

// OperationalHealth holds the computed health scores for a location.
type OperationalHealth struct {
	OverallScore   float64 `json:"overall_score"`
	KitchenScore   float64 `json:"kitchen_score"`
	TicketScore    float64 `json:"ticket_score"`
	StaffScore     float64 `json:"staff_score"`
	FinancialScore float64 `json:"financial_score"`
	InventoryScore float64 `json:"inventory_score"`
	Status         string  `json:"status"`
}

// classifyHealth returns a health status label based on overall score.
func classifyHealth(score float64) string {
	switch {
	case score > 90:
		return "excellent"
	case score > 75:
		return "good"
	case score > 60:
		return "fair"
	case score > 40:
		return "poor"
	default:
		return "critical"
	}
}

// computeHealthScore computes the weighted overall health score.
// Weights: kitchen 25%, ticket 25%, staff 20%, financial 15%, inventory 15%.
func computeHealthScore(kitchen, ticket, staff, financial, inventory float64) float64 {
	return kitchen*0.25 + ticket*0.25 + staff*0.20 + financial*0.15 + inventory*0.15
}

// GetOperationalHealth computes and returns the operational health for a location.
func (s *Service) GetOperationalHealth(ctx context.Context, orgID, locationID string) (*OperationalHealth, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	// Kitchen score: 100 - overall capacity percentage.
	kitchenScore, err := s.computeKitchenScore(ctx, orgID, locationID)
	if err != nil {
		return nil, fmt.Errorf("kitchen score: %w", err)
	}

	// Ticket score: % of tickets completed within 10-minute SLA in the last hour.
	ticketScore, err := s.computeTicketScore(tenantCtx, orgID, locationID)
	if err != nil {
		return nil, fmt.Errorf("ticket score: %w", err)
	}

	// Staff score: scheduled vs required headcount ratio.
	staffScore, err := s.computeStaffScore(tenantCtx, orgID, locationID)
	if err != nil {
		return nil, fmt.Errorf("staff score: %w", err)
	}

	// Financial score: stubbed at 75 (requires budget variance data).
	financialScore := 75.0

	// Inventory score: 100 - (par_breach_count / total_ingredients * 100).
	inventoryScore, err := s.computeInventoryScore(tenantCtx, orgID, locationID)
	if err != nil {
		return nil, fmt.Errorf("inventory score: %w", err)
	}

	overall := computeHealthScore(kitchenScore, ticketScore, staffScore, financialScore, inventoryScore)

	return &OperationalHealth{
		OverallScore:   overall,
		KitchenScore:   kitchenScore,
		TicketScore:    ticketScore,
		StaffScore:     staffScore,
		FinancialScore: financialScore,
		InventoryScore: inventoryScore,
		Status:         classifyHealth(overall),
	}, nil
}

// computeKitchenScore returns 100 - overall_load_pct.
func (s *Service) computeKitchenScore(ctx context.Context, orgID, locationID string) (float64, error) {
	capacity, err := s.CalculateCapacity(ctx, orgID, locationID)
	if err != nil {
		return 0, err
	}
	score := 100.0 - capacity.OverallLoadPct
	if score < 0 {
		score = 0
	}
	return score, nil
}

// computeTicketScore returns the percentage of tickets completed within 10 minutes in the last hour.
func (s *Service) computeTicketScore(ctx context.Context, orgID, locationID string) (float64, error) {
	var total, withinSLA int
	err := database.TenantTx(ctx, s.pool, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			`SELECT
			    COUNT(*) AS total,
			    COUNT(*) FILTER (
			        WHERE EXTRACT(EPOCH FROM (actual_ready_at - created_at)) <= 600
			    ) AS within_sla
			 FROM kds_tickets
			 WHERE org_id = $1 AND location_id = $2
			   AND status = 'ready'
			   AND actual_ready_at IS NOT NULL
			   AND created_at >= now() - INTERVAL '1 hour'`,
			orgID, locationID,
		).Scan(&total, &withinSLA)
	})
	if err != nil {
		return 0, err
	}
	if total == 0 {
		return 100.0, nil // no tickets = no issues
	}
	return float64(withinSLA) / float64(total) * 100, nil
}

// computeStaffScore computes staff adequacy based on scheduled vs forecast headcount.
func (s *Service) computeStaffScore(ctx context.Context, orgID, locationID string) (float64, error) {
	var scheduled, required int
	err := database.TenantTx(ctx, s.pool, func(tx pgx.Tx) error {
		// Count shifts scheduled for today's date that overlap now.
		if err := tx.QueryRow(ctx,
			`SELECT COUNT(*)
			 FROM shifts
			 WHERE org_id = $1 AND location_id = $2
			   AND DATE(shift_date) = CURRENT_DATE
			   AND status NOT IN ('cancelled')`,
			orgID, locationID,
		).Scan(&scheduled); err != nil {
			return fmt.Errorf("query scheduled shifts: %w", err)
		}

		// Forecast required headcount from demand_forecasts for today.
		if err := tx.QueryRow(ctx,
			`SELECT COALESCE(
			    (SELECT SUM(required_headcount)
			     FROM demand_forecasts
			     WHERE org_id = $1 AND location_id = $2
			       AND forecast_date = CURRENT_DATE),
			    0
			 )`,
			orgID, locationID,
		).Scan(&required); err != nil {
			return fmt.Errorf("query forecast headcount: %w", err)
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	if required == 0 {
		// No forecast data — assume fully staffed.
		return 100.0, nil
	}

	ratio := float64(scheduled) / float64(required)
	if ratio > 1 {
		ratio = 1
	}
	return ratio * 100, nil
}

// computeInventoryScore returns 100 - (par_breach_count / total_ingredients * 100).
func (s *Service) computeInventoryScore(ctx context.Context, orgID, locationID string) (float64, error) {
	var total, breaches int
	err := database.TenantTx(ctx, s.pool, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx,
			`SELECT COUNT(*) FROM ingredients WHERE org_id = $1`,
			orgID,
		).Scan(&total); err != nil {
			return fmt.Errorf("count ingredients: %w", err)
		}

		if err := tx.QueryRow(ctx,
			`SELECT COUNT(DISTINCT ingredient_id)
			 FROM par_levels
			 WHERE org_id = $1 AND location_id = $2
			   AND current_qty < par_qty`,
			orgID, locationID,
		).Scan(&breaches); err != nil {
			return fmt.Errorf("count par breaches: %w", err)
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	if total == 0 {
		return 100.0, nil
	}

	score := 100.0 - (float64(breaches)/float64(total))*100
	if score < 0 {
		score = 0
	}
	return score, nil
}
