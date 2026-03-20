package operations

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/opsnerve/fireline/internal/tenant"
	"github.com/opsnerve/fireline/pkg/database"
)

// RealTimeHorizon holds a live snapshot of the current operational state.
type RealTimeHorizon struct {
	Health        *OperationalHealth `json:"health"`
	Overload      *OverloadStatus    `json:"overload"`
	ActiveTickets int                `json:"active_tickets"`
	AvgTicketTime int                `json:"avg_ticket_time_secs"`
	StationLoads  []any              `json:"station_loads"`
}

// ShiftHorizon holds the operational view for the next 4 hours.
type ShiftHorizon struct {
	ForecastedCovers int   `json:"forecasted_covers"`
	ScheduledStaff   int   `json:"scheduled_staff"`
	RequiredStaff    int   `json:"required_staff"`
	StaffGap         int   `json:"staff_gap"`
	ExpectedRevenue  int64 `json:"expected_revenue"`
}

// DailyHorizon holds today's operational overview.
type DailyHorizon struct {
	PrepItems          int   `json:"prep_items"`
	ExpectedDeliveries int   `json:"expected_deliveries"`
	ScheduledShifts    int   `json:"scheduled_shifts"`
	ForecastedRevenue  int64 `json:"forecasted_revenue"`
}

// WeeklyHorizon holds this week's operational plan.
type WeeklyHorizon struct {
	TotalScheduledHours float64 `json:"total_scheduled_hours"`
	PendingPOs          int     `json:"pending_pos"`
	ProjectedLaborCost  int64   `json:"projected_labor_cost"`
	ProjectedRevenue    int64   `json:"projected_revenue"`
}

// StrategicHorizon holds 30-day trailing trends.
type StrategicHorizon struct {
	RevenueTrailing30d      int64   `json:"revenue_trailing_30d"`
	COGSTrailing30d         int64   `json:"cogs_trailing_30d"`
	LaborCostPctTrend       float64 `json:"labor_cost_pct_trend"`
	TopClassificationShifts []any   `json:"top_classification_shifts"`
}

// GetRealTimeHorizon composes health + overload + ticket data into a live snapshot.
func (s *Service) GetRealTimeHorizon(ctx context.Context, orgID, locationID string) (*RealTimeHorizon, error) {
	health, err := s.GetOperationalHealth(ctx, orgID, locationID)
	if err != nil {
		return nil, fmt.Errorf("health: %w", err)
	}

	overload, err := s.GetOverloadStatus(ctx, orgID, locationID)
	if err != nil {
		return nil, fmt.Errorf("overload: %w", err)
	}

	var activeTickets int
	var avgTicketSecs int
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	err = database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		if err := tx.QueryRow(tenantCtx,
			`SELECT COUNT(*) FROM kds_tickets
			 WHERE org_id = $1 AND location_id = $2
			   AND status NOT IN ('ready', 'delivered', 'cancelled')`,
			orgID, locationID,
		).Scan(&activeTickets); err != nil {
			return fmt.Errorf("count active tickets: %w", err)
		}

		if err := tx.QueryRow(tenantCtx,
			`SELECT COALESCE(
			    AVG(EXTRACT(EPOCH FROM (actual_ready_at - created_at)))::INT, 0
			 )
			 FROM kds_tickets
			 WHERE org_id = $1 AND location_id = $2
			   AND status = 'ready'
			   AND actual_ready_at IS NOT NULL
			   AND created_at >= now() - INTERVAL '1 hour'`,
			orgID, locationID,
		).Scan(&avgTicketSecs); err != nil {
			return fmt.Errorf("avg ticket time: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Station loads from capacity.
	capacity, err := s.CalculateCapacity(ctx, orgID, locationID)
	stationLoads := []any{}
	if err == nil {
		for _, sl := range capacity.StationCapacity {
			stationLoads = append(stationLoads, sl)
		}
	}

	return &RealTimeHorizon{
		Health:        health,
		Overload:      overload,
		ActiveTickets: activeTickets,
		AvgTicketTime: avgTicketSecs,
		StationLoads:  stationLoads,
	}, nil
}

// GetShiftHorizon returns forecast + schedule data for the next 4 hours.
func (s *Service) GetShiftHorizon(ctx context.Context, orgID, locationID string) (*ShiftHorizon, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	var forecastedCovers, scheduledStaff, requiredStaff int
	var expectedRevenue int64

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		now := time.Now()
		windowEnd := now.Add(4 * time.Hour)

		// Forecasted covers for next 4 hours.
		if err := tx.QueryRow(tenantCtx,
			`SELECT COALESCE(SUM(forecasted_covers), 0)
			 FROM demand_forecasts
			 WHERE org_id = $1 AND location_id = $2
			   AND forecast_date = $3
			   AND forecast_hour >= $4 AND forecast_hour < $5`,
			orgID, locationID, now.Format("2006-01-02"),
			now.Hour(), windowEnd.Hour(),
		).Scan(&forecastedCovers); err != nil {
			return fmt.Errorf("forecasted covers: %w", err)
		}

		// Scheduled staff active in next 4 hours.
		if err := tx.QueryRow(tenantCtx,
			`SELECT COUNT(*)
			 FROM shifts
			 WHERE org_id = $1 AND location_id = $2
			   AND shift_date = $3
			   AND status NOT IN ('cancelled')`,
			orgID, locationID, now.Format("2006-01-02"),
		).Scan(&scheduledStaff); err != nil {
			return fmt.Errorf("scheduled staff: %w", err)
		}

		// Required headcount from forecast.
		if err := tx.QueryRow(tenantCtx,
			`SELECT COALESCE(SUM(required_headcount), 0)
			 FROM demand_forecasts
			 WHERE org_id = $1 AND location_id = $2
			   AND forecast_date = $3
			   AND forecast_hour >= $4 AND forecast_hour < $5`,
			orgID, locationID, now.Format("2006-01-02"),
			now.Hour(), windowEnd.Hour(),
		).Scan(&requiredStaff); err != nil {
			return fmt.Errorf("required headcount: %w", err)
		}

		// Expected revenue from last comparable 4-hour window.
		if err := tx.QueryRow(tenantCtx,
			`SELECT COALESCE(SUM(forecasted_revenue), 0)::BIGINT
			 FROM demand_forecasts
			 WHERE org_id = $1 AND location_id = $2
			   AND forecast_date = $3
			   AND forecast_hour >= $4 AND forecast_hour < $5`,
			orgID, locationID, now.Format("2006-01-02"),
			now.Hour(), windowEnd.Hour(),
		).Scan(&expectedRevenue); err != nil {
			return fmt.Errorf("expected revenue: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	staffGap := requiredStaff - scheduledStaff
	if staffGap < 0 {
		staffGap = 0
	}

	return &ShiftHorizon{
		ForecastedCovers: forecastedCovers,
		ScheduledStaff:   scheduledStaff,
		RequiredStaff:    requiredStaff,
		StaffGap:         staffGap,
		ExpectedRevenue:  expectedRevenue,
	}, nil
}

// GetDailyHorizon returns today's operational overview.
func (s *Service) GetDailyHorizon(ctx context.Context, orgID, locationID string) (*DailyHorizon, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	var prepItems, expectedDeliveries, scheduledShifts int
	var forecastedRevenue int64

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		today := time.Now().Format("2006-01-02")

		// Prep items: count prep-type ticket items for today.
		if err := tx.QueryRow(tenantCtx,
			`SELECT COUNT(*)
			 FROM kds_ticket_items kti
			 JOIN kds_tickets kt ON kt.ticket_id = kti.ticket_id
			 WHERE kti.org_id = $1 AND kt.location_id = $2
			   AND kti.station_type = 'prep'
			   AND DATE(kt.created_at) = $3`,
			orgID, locationID, today,
		).Scan(&prepItems); err != nil {
			return fmt.Errorf("prep items: %w", err)
		}

		// Expected deliveries: purchase orders expected today.
		if err := tx.QueryRow(tenantCtx,
			`SELECT COUNT(*)
			 FROM purchase_orders
			 WHERE org_id = $1 AND location_id = $2
			   AND expected_delivery_date = $3
			   AND status NOT IN ('cancelled', 'received')`,
			orgID, locationID, today,
		).Scan(&expectedDeliveries); err != nil {
			return fmt.Errorf("expected deliveries: %w", err)
		}

		// Scheduled shifts today.
		if err := tx.QueryRow(tenantCtx,
			`SELECT COUNT(*)
			 FROM shifts
			 WHERE org_id = $1 AND location_id = $2
			   AND shift_date = $3
			   AND status NOT IN ('cancelled')`,
			orgID, locationID, today,
		).Scan(&scheduledShifts); err != nil {
			return fmt.Errorf("scheduled shifts: %w", err)
		}

		// Forecasted revenue today.
		if err := tx.QueryRow(tenantCtx,
			`SELECT COALESCE(SUM(forecasted_revenue), 0)::BIGINT
			 FROM demand_forecasts
			 WHERE org_id = $1 AND location_id = $2
			   AND forecast_date = $3`,
			orgID, locationID, today,
		).Scan(&forecastedRevenue); err != nil {
			return fmt.Errorf("forecasted revenue: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &DailyHorizon{
		PrepItems:          prepItems,
		ExpectedDeliveries: expectedDeliveries,
		ScheduledShifts:    scheduledShifts,
		ForecastedRevenue:  forecastedRevenue,
	}, nil
}

// GetWeeklyHorizon returns this week's operational plan.
func (s *Service) GetWeeklyHorizon(ctx context.Context, orgID, locationID string) (*WeeklyHorizon, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	var totalHours float64
	var pendingPOs int
	var projectedLaborCost, projectedRevenue int64

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		now := time.Now()
		weekStart := now.AddDate(0, 0, -int(now.Weekday()))
		weekEnd := weekStart.AddDate(0, 0, 7)

		// Total scheduled hours this week.
		if err := tx.QueryRow(tenantCtx,
			`SELECT COALESCE(
			    SUM(EXTRACT(EPOCH FROM (shift_end - shift_start)) / 3600.0), 0
			 )
			 FROM shifts
			 WHERE org_id = $1 AND location_id = $2
			   AND shift_date >= $3 AND shift_date < $4
			   AND status NOT IN ('cancelled')`,
			orgID, locationID,
			weekStart.Format("2006-01-02"), weekEnd.Format("2006-01-02"),
		).Scan(&totalHours); err != nil {
			return fmt.Errorf("total scheduled hours: %w", err)
		}

		// Pending purchase orders.
		if err := tx.QueryRow(tenantCtx,
			`SELECT COUNT(*)
			 FROM purchase_orders
			 WHERE org_id = $1 AND location_id = $2
			   AND status = 'pending'`,
			orgID, locationID,
		).Scan(&pendingPOs); err != nil {
			return fmt.Errorf("pending POs: %w", err)
		}

		// Projected labor cost: sum of (hours * hourly_wage) from shifts this week.
		if err := tx.QueryRow(tenantCtx,
			`SELECT COALESCE(
			    SUM(
			        EXTRACT(EPOCH FROM (shift_end - shift_start)) / 3600.0
			        * COALESCE(hourly_wage, 0)
			    )::BIGINT, 0
			 )
			 FROM shifts
			 WHERE org_id = $1 AND location_id = $2
			   AND shift_date >= $3 AND shift_date < $4
			   AND status NOT IN ('cancelled')`,
			orgID, locationID,
			weekStart.Format("2006-01-02"), weekEnd.Format("2006-01-02"),
		).Scan(&projectedLaborCost); err != nil {
			return fmt.Errorf("projected labor cost: %w", err)
		}

		// Projected revenue from demand forecasts this week.
		if err := tx.QueryRow(tenantCtx,
			`SELECT COALESCE(SUM(forecasted_revenue), 0)::BIGINT
			 FROM demand_forecasts
			 WHERE org_id = $1 AND location_id = $2
			   AND forecast_date >= $3 AND forecast_date < $4`,
			orgID, locationID,
			weekStart.Format("2006-01-02"), weekEnd.Format("2006-01-02"),
		).Scan(&projectedRevenue); err != nil {
			return fmt.Errorf("projected revenue: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &WeeklyHorizon{
		TotalScheduledHours: totalHours,
		PendingPOs:          pendingPOs,
		ProjectedLaborCost:  projectedLaborCost,
		ProjectedRevenue:    projectedRevenue,
	}, nil
}

// GetStrategicHorizon returns 30-day trailing trend data.
func (s *Service) GetStrategicHorizon(ctx context.Context, orgID, locationID string) (*StrategicHorizon, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	var revenue30d, cogs30d int64
	var laborCostPct float64

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		cutoff := time.Now().AddDate(0, 0, -30)

		// Revenue trailing 30 days.
		if err := tx.QueryRow(tenantCtx,
			`SELECT COALESCE(SUM(subtotal), 0)::BIGINT
			 FROM checks
			 WHERE org_id = $1 AND location_id = $2
			   AND status = 'closed'
			   AND closed_at >= $3`,
			orgID, locationID, cutoff,
		).Scan(&revenue30d); err != nil {
			return fmt.Errorf("revenue 30d: %w", err)
		}

		// COGS trailing 30 days from inventory_transactions (type = 'usage').
		if err := tx.QueryRow(tenantCtx,
			`SELECT COALESCE(SUM(unit_cost * quantity), 0)::BIGINT
			 FROM inventory_transactions
			 WHERE org_id = $1 AND location_id = $2
			   AND transaction_type = 'usage'
			   AND created_at >= $3`,
			orgID, locationID, cutoff,
		).Scan(&cogs30d); err != nil {
			return fmt.Errorf("cogs 30d: %w", err)
		}

		// Labor cost % trend: labor cost / revenue * 100.
		var laborCost int64
		if err := tx.QueryRow(tenantCtx,
			`SELECT COALESCE(
			    SUM(
			        EXTRACT(EPOCH FROM (shift_end - shift_start)) / 3600.0
			        * COALESCE(hourly_wage, 0)
			    )::BIGINT, 0
			 )
			 FROM shifts
			 WHERE org_id = $1 AND location_id = $2
			   AND shift_date >= $3
			   AND status NOT IN ('cancelled')`,
			orgID, locationID, cutoff.Format("2006-01-02"),
		).Scan(&laborCost); err != nil {
			return fmt.Errorf("labor cost: %w", err)
		}

		if revenue30d > 0 {
			laborCostPct = float64(laborCost) / float64(revenue30d) * 100
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &StrategicHorizon{
		RevenueTrailing30d:      revenue30d,
		COGSTrailing30d:         cogs30d,
		LaborCostPctTrend:       laborCostPct,
		TopClassificationShifts: []any{},
	}, nil
}
