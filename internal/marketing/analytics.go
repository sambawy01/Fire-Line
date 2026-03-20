package marketing

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/opsnerve/fireline/internal/tenant"
	"github.com/opsnerve/fireline/pkg/database"
)

// CampaignMetrics holds aggregate campaign analytics for an org.
type CampaignMetrics struct {
	ActiveCampaigns   int     `json:"active_campaigns"`
	TotalRedemptions  int     `json:"total_redemptions"`
	RevenueAttributed int64   `json:"revenue_attributed"`
	AvgRedemptionRate float64 `json:"avg_redemption_rate"`
}

// LoyaltyMetrics holds aggregate loyalty program analytics for an org.
type LoyaltyMetrics struct {
	TotalMembers  int     `json:"total_members"`
	BronzeCount   int     `json:"bronze_count"`
	SilverCount   int     `json:"silver_count"`
	GoldCount     int     `json:"gold_count"`
	PlatinumCount int     `json:"platinum_count"`
	AvgBalance    float64 `json:"avg_balance"`
	TotalIssued   float64 `json:"total_issued"`
	TotalRedeemed float64 `json:"total_redeemed"`
}

// GetCampaignMetrics returns aggregate campaign analytics for an org.
func (s *Service) GetCampaignMetrics(ctx context.Context, orgID string) (*CampaignMetrics, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var m CampaignMetrics
	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		return tx.QueryRow(tenantCtx,
			`SELECT
				COUNT(*) FILTER (WHERE status = 'active') AS active_campaigns,
				COALESCE(SUM(redemptions), 0) AS total_redemptions,
				COALESCE(SUM(revenue_attributed), 0) AS revenue_attributed,
				COALESCE(
					AVG(CASE WHEN redemptions > 0 THEN redemptions::FLOAT ELSE NULL END),
					0
				) AS avg_redemption_rate
			FROM campaigns WHERE org_id = $1`,
			orgID,
		).Scan(&m.ActiveCampaigns, &m.TotalRedemptions, &m.RevenueAttributed, &m.AvgRedemptionRate)
	})
	if err != nil {
		return nil, fmt.Errorf("get campaign metrics: %w", err)
	}
	return &m, nil
}

// GetLoyaltyMetrics returns aggregate loyalty program analytics for an org.
func (s *Service) GetLoyaltyMetrics(ctx context.Context, orgID string) (*LoyaltyMetrics, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var m LoyaltyMetrics
	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		// Member tier distribution and balance.
		err := tx.QueryRow(tenantCtx,
			`SELECT
				COUNT(*) AS total_members,
				COUNT(*) FILTER (WHERE tier = 'bronze') AS bronze_count,
				COUNT(*) FILTER (WHERE tier = 'silver') AS silver_count,
				COUNT(*) FILTER (WHERE tier = 'gold') AS gold_count,
				COUNT(*) FILTER (WHERE tier = 'platinum') AS platinum_count,
				COALESCE(AVG(points_balance), 0) AS avg_balance
			FROM loyalty_members WHERE org_id = $1`,
			orgID,
		).Scan(&m.TotalMembers, &m.BronzeCount, &m.SilverCount, &m.GoldCount, &m.PlatinumCount, &m.AvgBalance)
		if err != nil {
			return fmt.Errorf("query member metrics: %w", err)
		}

		// Points issued and redeemed from transactions.
		return tx.QueryRow(tenantCtx,
			`SELECT
				COALESCE(SUM(points) FILTER (WHERE type = 'earn'), 0) AS total_issued,
				COALESCE(SUM(points) FILTER (WHERE type = 'redeem'), 0) AS total_redeemed
			FROM loyalty_transactions WHERE org_id = $1`,
			orgID,
		).Scan(&m.TotalIssued, &m.TotalRedeemed)
	})
	if err != nil {
		return nil, fmt.Errorf("get loyalty metrics: %w", err)
	}
	return &m, nil
}
