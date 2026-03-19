package financial

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/opsnerve/fireline/internal/event"
	"github.com/opsnerve/fireline/internal/tenant"
	"github.com/opsnerve/fireline/pkg/database"
)

// Service provides financial intelligence capabilities.
type Service struct {
	pool *pgxpool.Pool
	bus  *event.Bus
}

// New creates a new financial intelligence service.
func New(pool *pgxpool.Pool, bus *event.Bus) *Service {
	return &Service{pool: pool, bus: bus}
}

// RegisterHandlers subscribes to pipeline events.
func (s *Service) RegisterHandlers() {
	s.bus.Subscribe("pipeline.orders.processed", s.handleOrdersProcessed)
}

// ProfitAndLoss represents a P&L summary for a period.
type ProfitAndLoss struct {
	LocationID     string         `json:"location_id"`
	PeriodStart    time.Time      `json:"period_start"`
	PeriodEnd      time.Time      `json:"period_end"`
	GrossRevenue   int64          `json:"gross_revenue"`   // cents
	Discounts      int64          `json:"discounts"`       // cents
	NetRevenue     int64          `json:"net_revenue"`     // cents
	COGS           int64          `json:"cogs"`            // cents (cost of goods sold)
	GrossProfit    int64          `json:"gross_profit"`    // cents
	GrossMargin    float64        `json:"gross_margin"`    // percentage
	TaxCollected   int64          `json:"tax_collected"`   // cents
	Tips           int64          `json:"tips"`            // cents
	CheckCount     int            `json:"check_count"`
	AvgCheckSize   int64          `json:"avg_check_size"`  // cents
	ByChannel      []ChannelBreakdown `json:"by_channel"`
}

// ChannelBreakdown shows revenue and margins per sales channel.
type ChannelBreakdown struct {
	Channel      string  `json:"channel"`
	Revenue      int64   `json:"revenue"`      // cents
	COGS         int64   `json:"cogs"`         // cents
	GrossProfit  int64   `json:"gross_profit"` // cents
	GrossMargin  float64 `json:"gross_margin"` // percentage
	CheckCount   int     `json:"check_count"`
	AvgCheckSize int64   `json:"avg_check_size"` // cents
}

// Anomaly represents a detected financial anomaly using Z-score.
type Anomaly struct {
	MetricName   string    `json:"metric_name"`
	CurrentValue float64   `json:"current_value"`
	Mean         float64   `json:"mean"`
	StdDev       float64   `json:"std_dev"`
	ZScore       float64   `json:"z_score"`
	Severity     string    `json:"severity"` // "warning" (2σ) or "critical" (3σ)
	DetectedAt   time.Time `json:"detected_at"`
}

// CalculatePnL computes a P&L summary for a location and date range.
func (s *Service) CalculatePnL(ctx context.Context, orgID, locationID string, from, to time.Time) (*ProfitAndLoss, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	pnl := &ProfitAndLoss{
		LocationID:  locationID,
		PeriodStart: from,
		PeriodEnd:   to,
	}

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		// Aggregate check-level financials
		err := tx.QueryRow(tenantCtx,
			`SELECT
				COALESCE(SUM(subtotal), 0),
				COALESCE(SUM(discount), 0),
				COALESCE(SUM(subtotal - discount), 0),
				COALESCE(SUM(tax), 0),
				COALESCE(SUM(tip), 0),
				COUNT(*)
			 FROM checks
			 WHERE location_id = $1
			   AND closed_at >= $2 AND closed_at < $3
			   AND status = 'closed'`,
			locationID, from, to,
		).Scan(&pnl.GrossRevenue, &pnl.Discounts, &pnl.NetRevenue,
			&pnl.TaxCollected, &pnl.Tips, &pnl.CheckCount)
		if err != nil {
			return fmt.Errorf("aggregate checks: %w", err)
		}

		if pnl.CheckCount > 0 {
			pnl.AvgCheckSize = pnl.NetRevenue / int64(pnl.CheckCount)
		}

		// Calculate COGS from recipe explosion + check items
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
			locationID, from, to,
		).Scan(&pnl.COGS)
		if err != nil {
			return fmt.Errorf("calculate COGS: %w", err)
		}

		pnl.GrossProfit = pnl.NetRevenue - pnl.COGS
		if pnl.NetRevenue > 0 {
			pnl.GrossMargin = float64(pnl.GrossProfit) / float64(pnl.NetRevenue) * 100
		}

		// Channel breakdown
		rows, err := tx.Query(tenantCtx,
			`SELECT
				c.channel,
				COALESCE(SUM(c.subtotal - c.discount), 0) AS revenue,
				COUNT(*) AS check_count
			 FROM checks c
			 WHERE c.location_id = $1
			   AND c.closed_at >= $2 AND c.closed_at < $3
			   AND c.status = 'closed'
			 GROUP BY c.channel
			 ORDER BY revenue DESC`,
			locationID, from, to,
		)
		if err != nil {
			return fmt.Errorf("channel breakdown: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var cb ChannelBreakdown
			if err := rows.Scan(&cb.Channel, &cb.Revenue, &cb.CheckCount); err != nil {
				return err
			}
			if cb.CheckCount > 0 {
				cb.AvgCheckSize = cb.Revenue / int64(cb.CheckCount)
			}
			pnl.ByChannel = append(pnl.ByChannel, cb)
		}
		return rows.Err()
	})

	return pnl, err
}

// DetectAnomalies compares current day metrics against historical data using Z-scores.
// Returns anomalies where metrics deviate more than 2 standard deviations.
func (s *Service) DetectAnomalies(ctx context.Context, orgID, locationID string, currentDay time.Time) ([]Anomaly, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var anomalies []Anomaly

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		// Get last 30 days of daily revenue for baseline
		rows, err := tx.Query(tenantCtx,
			`SELECT
				DATE(closed_at) AS day,
				SUM(subtotal - discount) AS daily_revenue,
				COUNT(*) AS daily_checks
			 FROM checks
			 WHERE location_id = $1
			   AND closed_at >= $2
			   AND closed_at < $3
			   AND status = 'closed'
			 GROUP BY DATE(closed_at)
			 ORDER BY day`,
			locationID,
			currentDay.AddDate(0, 0, -30),
			currentDay,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		var revenues []float64
		var checkCounts []float64
		for rows.Next() {
			var day time.Time
			var rev int64
			var cnt int
			if err := rows.Scan(&day, &rev, &cnt); err != nil {
				return err
			}
			revenues = append(revenues, float64(rev))
			checkCounts = append(checkCounts, float64(cnt))
		}
		if err := rows.Err(); err != nil {
			return err
		}

		// Get current day metrics
		var todayRev int64
		var todayChecks int
		err = tx.QueryRow(tenantCtx,
			`SELECT COALESCE(SUM(subtotal - discount), 0), COUNT(*)
			 FROM checks
			 WHERE location_id = $1
			   AND closed_at >= $2
			   AND closed_at < $3
			   AND status = 'closed'`,
			locationID, currentDay, currentDay.AddDate(0, 0, 1),
		).Scan(&todayRev, &todayChecks)
		if err != nil {
			return err
		}

		// Check revenue anomaly
		if a := detectZScoreAnomaly("daily_revenue", float64(todayRev), revenues); a != nil {
			anomalies = append(anomalies, *a)
		}

		// Check check count anomaly
		if a := detectZScoreAnomaly("daily_check_count", float64(todayChecks), checkCounts); a != nil {
			anomalies = append(anomalies, *a)
		}

		return nil
	})

	return anomalies, err
}

// detectZScoreAnomaly computes Z-score and returns an anomaly if |Z| > 2.
func detectZScoreAnomaly(metricName string, currentValue float64, historical []float64) *Anomaly {
	if len(historical) < 7 { // need at least a week of data
		return nil
	}

	mean, stdDev := meanStdDev(historical)
	if stdDev == 0 {
		return nil
	}

	zScore := (currentValue - mean) / stdDev

	if math.Abs(zScore) < 2.0 {
		return nil
	}

	severity := "warning"
	if math.Abs(zScore) >= 3.0 {
		severity = "critical"
	}

	return &Anomaly{
		MetricName:   metricName,
		CurrentValue: currentValue,
		Mean:         mean,
		StdDev:       stdDev,
		ZScore:       zScore,
		Severity:     severity,
		DetectedAt:   time.Now(),
	}
}

func meanStdDev(data []float64) (float64, float64) {
	n := float64(len(data))
	if n == 0 {
		return 0, 0
	}

	var sum float64
	for _, v := range data {
		sum += v
	}
	mean := sum / n

	var sqDiffSum float64
	for _, v := range data {
		diff := v - mean
		sqDiffSum += diff * diff
	}
	stdDev := math.Sqrt(sqDiffSum / n)

	return mean, stdDev
}

func (s *Service) handleOrdersProcessed(ctx context.Context, env event.Envelope) error {
	slog.Info("financial: orders processed event received",
		"org_id", env.OrgID,
		"location_id", env.LocationID,
	)

	s.bus.Publish(ctx, event.Envelope{
		EventID:    env.EventID + ".financial.updated",
		EventType:  "financial.metrics.updated",
		OrgID:      env.OrgID,
		LocationID: env.LocationID,
		Source:     "financial",
	})

	return nil
}
