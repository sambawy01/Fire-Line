package financial

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/opsnerve/fireline/internal/tenant"
	"github.com/opsnerve/fireline/pkg/database"
)

// Budget represents a financial budget for a location and period.
type Budget struct {
	BudgetID            string  `json:"budget_id"`
	LocationID          string  `json:"location_id"`
	PeriodType          string  `json:"period_type"`
	PeriodStart         string  `json:"period_start"`
	PeriodEnd           string  `json:"period_end"`
	RevenueTarget       int64   `json:"revenue_target"`
	FoodCostPctTarget   float64 `json:"food_cost_pct_target"`
	LaborCostPctTarget  float64 `json:"labor_cost_pct_target"`
	COGSTarget          int64   `json:"cogs_target"`
}

// BudgetVariance shows actual vs. budgeted performance.
type BudgetVariance struct {
	Budget              Budget  `json:"budget"`
	ActualRevenue       int64   `json:"actual_revenue"`
	ActualCOGS          int64   `json:"actual_cogs"`
	ActualFoodCostPct   float64 `json:"actual_food_cost_pct"`
	RevenueVariance     int64   `json:"revenue_variance"`
	RevenueVariancePct  float64 `json:"revenue_variance_pct"`
	COGSVariance        int64   `json:"cogs_variance"`
	COGSVariancePct     float64 `json:"cogs_variance_pct"`
	FoodCostPctDelta    float64 `json:"food_cost_pct_delta"`
	Status              string  `json:"status"` // on_track, over, under
}

// PeriodComparison compares current period P&L against prior periods.
type PeriodComparison struct {
	Current              *ProfitAndLoss `json:"current"`
	LastWeek             *ProfitAndLoss `json:"last_week"`
	LastMonth            *ProfitAndLoss `json:"last_month"`
	RevenueVsLastWeek    float64        `json:"revenue_vs_last_week_pct"`
	RevenueVsLastMonth   float64        `json:"revenue_vs_last_month_pct"`
	COGSVsLastWeek       float64        `json:"cogs_vs_last_week_pct"`
	COGSVsLastMonth      float64        `json:"cogs_vs_last_month_pct"`
}

// CreateBudget inserts a new budget record, updating on conflict.
func (s *Service) CreateBudget(ctx context.Context, orgID string, b Budget) (*Budget, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var created Budget

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		return tx.QueryRow(tenantCtx,
			`INSERT INTO budgets (
				org_id, location_id, period_type, period_start, period_end,
				revenue_target, food_cost_pct_target, labor_cost_pct_target, cogs_target
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (org_id, location_id, period_type, period_start) DO UPDATE SET
				period_end           = EXCLUDED.period_end,
				revenue_target       = EXCLUDED.revenue_target,
				food_cost_pct_target = EXCLUDED.food_cost_pct_target,
				labor_cost_pct_target = EXCLUDED.labor_cost_pct_target,
				cogs_target          = EXCLUDED.cogs_target,
				updated_at           = now()
			RETURNING budget_id, location_id, period_type,
				period_start::TEXT, period_end::TEXT,
				revenue_target, food_cost_pct_target, labor_cost_pct_target, cogs_target`,
			orgID, b.LocationID, b.PeriodType, b.PeriodStart, b.PeriodEnd,
			b.RevenueTarget, b.FoodCostPctTarget, b.LaborCostPctTarget, b.COGSTarget,
		).Scan(
			&created.BudgetID, &created.LocationID, &created.PeriodType,
			&created.PeriodStart, &created.PeriodEnd,
			&created.RevenueTarget, &created.FoodCostPctTarget,
			&created.LaborCostPctTarget, &created.COGSTarget,
		)
	})
	if err != nil {
		return nil, fmt.Errorf("create budget: %w", err)
	}
	return &created, nil
}

// ListBudgets returns budgets for a location, optionally filtered by period type.
func (s *Service) ListBudgets(ctx context.Context, orgID, locationID, periodType string) ([]Budget, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var budgets []Budget

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		var rows pgx.Rows
		var err error
		if periodType != "" {
			rows, err = tx.Query(tenantCtx,
				`SELECT budget_id, location_id, period_type,
					period_start::TEXT, period_end::TEXT,
					revenue_target, food_cost_pct_target, labor_cost_pct_target, cogs_target
				 FROM budgets
				 WHERE org_id = $1 AND location_id = $2 AND period_type = $3
				 ORDER BY period_start DESC`,
				orgID, locationID, periodType,
			)
		} else {
			rows, err = tx.Query(tenantCtx,
				`SELECT budget_id, location_id, period_type,
					period_start::TEXT, period_end::TEXT,
					revenue_target, food_cost_pct_target, labor_cost_pct_target, cogs_target
				 FROM budgets
				 WHERE org_id = $1 AND location_id = $2
				 ORDER BY period_start DESC`,
				orgID, locationID,
			)
		}
		if err != nil {
			return fmt.Errorf("list budgets: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var b Budget
			if err := rows.Scan(
				&b.BudgetID, &b.LocationID, &b.PeriodType,
				&b.PeriodStart, &b.PeriodEnd,
				&b.RevenueTarget, &b.FoodCostPctTarget,
				&b.LaborCostPctTarget, &b.COGSTarget,
			); err != nil {
				return err
			}
			budgets = append(budgets, b)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	return budgets, nil
}

// CalculateBudgetVariance finds the budget covering the given date and computes variance.
func (s *Service) CalculateBudgetVariance(ctx context.Context, orgID, locationID string, date time.Time) (*BudgetVariance, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var bv BudgetVariance

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		// Find the most specific budget covering this date (prefer daily > weekly > monthly)
		err := tx.QueryRow(tenantCtx,
			`SELECT budget_id, location_id, period_type,
				period_start::TEXT, period_end::TEXT,
				revenue_target, food_cost_pct_target, labor_cost_pct_target, cogs_target
			 FROM budgets
			 WHERE org_id = $1
			   AND location_id = $2
			   AND period_start <= $3::DATE
			   AND period_end >= $3::DATE
			 ORDER BY CASE period_type
				WHEN 'daily'   THEN 1
				WHEN 'weekly'  THEN 2
				WHEN 'monthly' THEN 3
			 END
			 LIMIT 1`,
			orgID, locationID, date.Format("2006-01-02"),
		).Scan(
			&bv.Budget.BudgetID, &bv.Budget.LocationID, &bv.Budget.PeriodType,
			&bv.Budget.PeriodStart, &bv.Budget.PeriodEnd,
			&bv.Budget.RevenueTarget, &bv.Budget.FoodCostPctTarget,
			&bv.Budget.LaborCostPctTarget, &bv.Budget.COGSTarget,
		)
		if err != nil {
			return fmt.Errorf("find budget: %w", err)
		}

		// Get actual revenue and COGS for the budget period
		periodStart, _ := time.Parse("2006-01-02", bv.Budget.PeriodStart)
		periodEnd, _ := time.Parse("2006-01-02", bv.Budget.PeriodEnd)
		periodEnd = periodEnd.Add(24 * time.Hour) // exclusive upper bound

		err = tx.QueryRow(tenantCtx,
			`SELECT
				COALESCE(SUM(subtotal - discount), 0) AS net_revenue
			 FROM checks
			 WHERE location_id = $1
			   AND closed_at >= $2 AND closed_at < $3
			   AND status = 'closed'`,
			locationID, periodStart, periodEnd,
		).Scan(&bv.ActualRevenue)
		if err != nil {
			return fmt.Errorf("actual revenue: %w", err)
		}

		err = tx.QueryRow(tenantCtx,
			`SELECT COALESCE(CAST(SUM(ci.quantity * re.quantity_per_unit * i.cost_per_unit) AS BIGINT), 0)
			 FROM check_items ci
			 JOIN checks c ON c.check_id = ci.check_id
			 JOIN menu_items mi ON mi.menu_item_id = ci.menu_item_id
			 JOIN recipe_explosion re ON re.menu_item_id = mi.menu_item_id
			 JOIN ingredients i ON i.ingredient_id = re.ingredient_id
			 WHERE c.location_id = $1
			   AND c.closed_at >= $2 AND c.closed_at < $3
			   AND c.status = 'closed'
			   AND ci.voided_at IS NULL`,
			locationID, periodStart, periodEnd,
		).Scan(&bv.ActualCOGS)
		if err != nil {
			return fmt.Errorf("actual COGS: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Compute variance metrics
	bv.RevenueVariance = bv.ActualRevenue - bv.Budget.RevenueTarget
	if bv.Budget.RevenueTarget > 0 {
		bv.RevenueVariancePct = float64(bv.RevenueVariance) / float64(bv.Budget.RevenueTarget) * 100
	}

	bv.COGSVariance = bv.ActualCOGS - bv.Budget.COGSTarget
	if bv.Budget.COGSTarget > 0 {
		bv.COGSVariancePct = float64(bv.COGSVariance) / float64(bv.Budget.COGSTarget) * 100
	}

	if bv.ActualRevenue > 0 {
		bv.ActualFoodCostPct = float64(bv.ActualCOGS) / float64(bv.ActualRevenue) * 100
	}
	bv.FoodCostPctDelta = bv.ActualFoodCostPct - bv.Budget.FoodCostPctTarget

	// Determine status
	switch {
	case bv.RevenueVariancePct >= -5:
		bv.Status = "on_track"
	case bv.RevenueVariancePct < -5:
		bv.Status = "under"
	default:
		bv.Status = "over"
	}
	if bv.ActualRevenue > bv.Budget.RevenueTarget {
		bv.Status = "over"
	}

	return &bv, nil
}

// CalculatePeriodComparison returns P&L for current period vs. equivalent prior periods.
func (s *Service) CalculatePeriodComparison(ctx context.Context, orgID, locationID string, from, to time.Time) (*PeriodComparison, error) {
	duration := to.Sub(from)

	current, err := s.CalculatePnL(ctx, orgID, locationID, from, to)
	if err != nil {
		return nil, fmt.Errorf("current period: %w", err)
	}

	lastWeekFrom := from.AddDate(0, 0, -7)
	lastWeekTo := to.AddDate(0, 0, -7)
	lastWeek, err := s.CalculatePnL(ctx, orgID, locationID, lastWeekFrom, lastWeekTo)
	if err != nil {
		return nil, fmt.Errorf("last week period: %w", err)
	}

	lastMonthFrom := from.AddDate(0, -1, 0)
	lastMonthTo := lastMonthFrom.Add(duration)
	lastMonth, err := s.CalculatePnL(ctx, orgID, locationID, lastMonthFrom, lastMonthTo)
	if err != nil {
		return nil, fmt.Errorf("last month period: %w", err)
	}

	pc := &PeriodComparison{
		Current:   current,
		LastWeek:  lastWeek,
		LastMonth: lastMonth,
	}

	if lastWeek.NetRevenue > 0 {
		pc.RevenueVsLastWeek = float64(current.NetRevenue-lastWeek.NetRevenue) / float64(lastWeek.NetRevenue) * 100
	}
	if lastMonth.NetRevenue > 0 {
		pc.RevenueVsLastMonth = float64(current.NetRevenue-lastMonth.NetRevenue) / float64(lastMonth.NetRevenue) * 100
	}
	if lastWeek.COGS > 0 {
		pc.COGSVsLastWeek = float64(current.COGS-lastWeek.COGS) / float64(lastWeek.COGS) * 100
	}
	if lastMonth.COGS > 0 {
		pc.COGSVsLastMonth = float64(current.COGS-lastMonth.COGS) / float64(lastMonth.COGS) * 100
	}

	return pc, nil
}
