package marketing

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/opsnerve/fireline/internal/event"
	"github.com/opsnerve/fireline/internal/tenant"
	"github.com/opsnerve/fireline/pkg/database"
)

// Campaign represents a marketing campaign.
type Campaign struct {
	CampaignID        string     `json:"campaign_id"`
	LocationID        *string    `json:"location_id,omitempty"`
	Name              string     `json:"name"`
	CampaignType      string     `json:"campaign_type"`
	Status            string     `json:"status"`
	TargetSegment     *string    `json:"target_segment,omitempty"`
	Channel           *string    `json:"channel,omitempty"`
	DiscountType      *string    `json:"discount_type,omitempty"`
	DiscountValue     *float64   `json:"discount_value,omitempty"`
	MinPurchase       int64      `json:"min_purchase"`
	StartAt           *time.Time `json:"start_at,omitempty"`
	EndAt             *time.Time `json:"end_at,omitempty"`
	Recurring         bool       `json:"recurring"`
	Redemptions       int        `json:"redemptions"`
	RevenueAttributed int64      `json:"revenue_attributed"`
	CostOfPromotion   int64      `json:"cost_of_promotion"`
	CreatedAt         time.Time  `json:"created_at"`
}

// CampaignInput holds fields for creating or updating a campaign.
type CampaignInput struct {
	LocationID    *string    `json:"location_id"`
	Name          string     `json:"name"`
	CampaignType  string     `json:"campaign_type"`
	TargetSegment *string    `json:"target_segment"`
	Channel       *string    `json:"channel"`
	DiscountType  *string    `json:"discount_type"`
	DiscountValue *float64   `json:"discount_value"`
	MinPurchase   int64      `json:"min_purchase"`
	StartAt       *time.Time `json:"start_at"`
	EndAt         *time.Time `json:"end_at"`
	Recurring     bool       `json:"recurring"`
	CreatedBy     string     `json:"created_by"`
}

// SimulationResult holds the projected outcome of a campaign simulation.
type SimulationResult struct {
	SegmentSize           int     `json:"segment_size"`
	ProjectedRedemptions  float64 `json:"projected_redemptions"`
	ProjectedRevenue      float64 `json:"projected_revenue"`
	AvgCheck              float64 `json:"avg_check"`
	ResponseRatePct       float64 `json:"response_rate_pct"`
}

// CreateCampaign inserts a new campaign and returns it.
func (s *Service) CreateCampaign(ctx context.Context, orgID string, input CampaignInput) (*Campaign, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var c Campaign
	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		row := tx.QueryRow(tenantCtx,
			`INSERT INTO campaigns (
				org_id, location_id, name, campaign_type, status,
				target_segment, channel, discount_type, discount_value,
				min_purchase, start_at, end_at, recurring, created_by
			) VALUES (
				$1, $2, $3, $4, 'draft',
				$5, $6, $7, $8,
				$9, $10, $11, $12, $13
			)
			RETURNING campaign_id, location_id, name, campaign_type, status,
				target_segment, channel, discount_type, discount_value,
				min_purchase, start_at, end_at, recurring,
				redemptions, revenue_attributed, cost_of_promotion, created_at`,
			orgID, input.LocationID, input.Name, input.CampaignType,
			input.TargetSegment, input.Channel, input.DiscountType, input.DiscountValue,
			input.MinPurchase, input.StartAt, input.EndAt, input.Recurring,
			nullableString(input.CreatedBy),
		)
		return scanCampaign(row, &c)
	})
	if err != nil {
		return nil, fmt.Errorf("create campaign: %w", err)
	}
	return &c, nil
}

// ListCampaigns returns campaigns for the org, optionally filtered by locationID and status.
func (s *Service) ListCampaigns(ctx context.Context, orgID, locationID, status string) ([]Campaign, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var campaigns []Campaign
	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		args := []any{orgID}
		where := "org_id = $1"
		idx := 2
		if locationID != "" {
			where += fmt.Sprintf(" AND location_id = $%d", idx)
			args = append(args, locationID)
			idx++
		}
		if status != "" {
			where += fmt.Sprintf(" AND status = $%d", idx)
			args = append(args, status)
		}
		rows, err := tx.Query(tenantCtx,
			fmt.Sprintf(`SELECT campaign_id, location_id, name, campaign_type, status,
				target_segment, channel, discount_type, discount_value,
				min_purchase, start_at, end_at, recurring,
				redemptions, revenue_attributed, cost_of_promotion, created_at
			FROM campaigns WHERE %s ORDER BY created_at DESC`, where),
			args...,
		)
		if err != nil {
			return fmt.Errorf("query campaigns: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			var c Campaign
			if err := rows.Scan(
				&c.CampaignID, &c.LocationID, &c.Name, &c.CampaignType, &c.Status,
				&c.TargetSegment, &c.Channel, &c.DiscountType, &c.DiscountValue,
				&c.MinPurchase, &c.StartAt, &c.EndAt, &c.Recurring,
				&c.Redemptions, &c.RevenueAttributed, &c.CostOfPromotion, &c.CreatedAt,
			); err != nil {
				return fmt.Errorf("scan campaign: %w", err)
			}
			campaigns = append(campaigns, c)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	if campaigns == nil {
		campaigns = []Campaign{}
	}
	return campaigns, nil
}

// GetCampaign returns a single campaign by ID.
func (s *Service) GetCampaign(ctx context.Context, orgID, campaignID string) (*Campaign, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var c Campaign
	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		row := tx.QueryRow(tenantCtx,
			`SELECT campaign_id, location_id, name, campaign_type, status,
				target_segment, channel, discount_type, discount_value,
				min_purchase, start_at, end_at, recurring,
				redemptions, revenue_attributed, cost_of_promotion, created_at
			FROM campaigns WHERE org_id = $1 AND campaign_id = $2`,
			orgID, campaignID,
		)
		return scanCampaign(row, &c)
	})
	if err != nil {
		return nil, fmt.Errorf("get campaign: %w", err)
	}
	return &c, nil
}

// UpdateCampaign updates a campaign that is in draft or scheduled status.
func (s *Service) UpdateCampaign(ctx context.Context, orgID, campaignID string, input CampaignInput) (*Campaign, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var c Campaign
	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		row := tx.QueryRow(tenantCtx,
			`UPDATE campaigns SET
				location_id = $3, name = $4, campaign_type = $5,
				target_segment = $6, channel = $7, discount_type = $8, discount_value = $9,
				min_purchase = $10, start_at = $11, end_at = $12, recurring = $13,
				updated_at = now()
			WHERE org_id = $1 AND campaign_id = $2 AND status IN ('draft', 'scheduled')
			RETURNING campaign_id, location_id, name, campaign_type, status,
				target_segment, channel, discount_type, discount_value,
				min_purchase, start_at, end_at, recurring,
				redemptions, revenue_attributed, cost_of_promotion, created_at`,
			orgID, campaignID,
			input.LocationID, input.Name, input.CampaignType,
			input.TargetSegment, input.Channel, input.DiscountType, input.DiscountValue,
			input.MinPurchase, input.StartAt, input.EndAt, input.Recurring,
		)
		return scanCampaign(row, &c)
	})
	if err != nil {
		return nil, fmt.Errorf("update campaign: %w", err)
	}
	return &c, nil
}

// ActivateCampaign sets a campaign to active and emits an event.
func (s *Service) ActivateCampaign(ctx context.Context, orgID, campaignID string) (*Campaign, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var c Campaign
	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		row := tx.QueryRow(tenantCtx,
			`UPDATE campaigns SET status = 'active', updated_at = now()
			WHERE org_id = $1 AND campaign_id = $2
			RETURNING campaign_id, location_id, name, campaign_type, status,
				target_segment, channel, discount_type, discount_value,
				min_purchase, start_at, end_at, recurring,
				redemptions, revenue_attributed, cost_of_promotion, created_at`,
			orgID, campaignID,
		)
		return scanCampaign(row, &c)
	})
	if err != nil {
		return nil, fmt.Errorf("activate campaign: %w", err)
	}
	s.bus.Publish(ctx, event.Envelope{
		EventType: "marketing.campaign.activated",
		OrgID:     orgID,
		Source:    "marketing",
		Payload:   map[string]string{"campaign_id": campaignID},
	})
	return &c, nil
}

// PauseCampaign sets a campaign to paused.
func (s *Service) PauseCampaign(ctx context.Context, orgID, campaignID string) (*Campaign, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var c Campaign
	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		row := tx.QueryRow(tenantCtx,
			`UPDATE campaigns SET status = 'paused', updated_at = now()
			WHERE org_id = $1 AND campaign_id = $2
			RETURNING campaign_id, location_id, name, campaign_type, status,
				target_segment, channel, discount_type, discount_value,
				min_purchase, start_at, end_at, recurring,
				redemptions, revenue_attributed, cost_of_promotion, created_at`,
			orgID, campaignID,
		)
		return scanCampaign(row, &c)
	})
	if err != nil {
		return nil, fmt.Errorf("pause campaign: %w", err)
	}
	return &c, nil
}

// CompleteCampaign sets a campaign to completed.
func (s *Service) CompleteCampaign(ctx context.Context, orgID, campaignID string) (*Campaign, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var c Campaign
	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		row := tx.QueryRow(tenantCtx,
			`UPDATE campaigns SET status = 'completed', updated_at = now()
			WHERE org_id = $1 AND campaign_id = $2
			RETURNING campaign_id, location_id, name, campaign_type, status,
				target_segment, channel, discount_type, discount_value,
				min_purchase, start_at, end_at, recurring,
				redemptions, revenue_attributed, cost_of_promotion, created_at`,
			orgID, campaignID,
		)
		return scanCampaign(row, &c)
	})
	if err != nil {
		return nil, fmt.Errorf("complete campaign: %w", err)
	}
	return &c, nil
}

// TrackRedemption increments a campaign's redemption count and revenue_attributed.
func (s *Service) TrackRedemption(ctx context.Context, orgID, campaignID string, amount int64) error {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	return database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(tenantCtx,
			`UPDATE campaigns SET
				redemptions = redemptions + 1,
				revenue_attributed = revenue_attributed + $3,
				updated_at = now()
			WHERE org_id = $1 AND campaign_id = $2`,
			orgID, campaignID, amount,
		)
		return err
	})
}

// SimulateCampaign estimates campaign performance based on segment size and avg check.
func (s *Service) SimulateCampaign(ctx context.Context, orgID string, input CampaignInput) (*SimulationResult, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var segmentSize int
	var avgCheck float64
	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		// Count target segment from guest_profiles.
		segQ := `SELECT COUNT(*) FROM guest_profiles WHERE org_id = $1`
		args := []any{orgID}
		if input.TargetSegment != nil && *input.TargetSegment != "" {
			segQ += ` AND segment = $2`
			args = append(args, *input.TargetSegment)
		}
		if err := tx.QueryRow(tenantCtx, segQ, args...).Scan(&segmentSize); err != nil {
			return fmt.Errorf("query segment size: %w", err)
		}

		// Compute average closed check value for the org.
		err := tx.QueryRow(tenantCtx,
			`SELECT COALESCE(AVG(total_amount), 0) FROM checks WHERE org_id = $1 AND status = 'closed'`,
			orgID,
		).Scan(&avgCheck)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("simulate campaign: %w", err)
	}

	const responseRate = 0.05
	projectedRedemptions := float64(segmentSize) * responseRate
	projectedRevenue := projectedRedemptions * avgCheck

	return &SimulationResult{
		SegmentSize:          segmentSize,
		ProjectedRedemptions: projectedRedemptions,
		ProjectedRevenue:     projectedRevenue,
		AvgCheck:             avgCheck,
		ResponseRatePct:      responseRate * 100,
	}, nil
}

// scanCampaign scans a single campaign row.
func scanCampaign(row pgx.Row, c *Campaign) error {
	return row.Scan(
		&c.CampaignID, &c.LocationID, &c.Name, &c.CampaignType, &c.Status,
		&c.TargetSegment, &c.Channel, &c.DiscountType, &c.DiscountValue,
		&c.MinPurchase, &c.StartAt, &c.EndAt, &c.Recurring,
		&c.Redemptions, &c.RevenueAttributed, &c.CostOfPromotion, &c.CreatedAt,
	)
}

// nullableString returns nil if s is empty, otherwise a pointer to s.
func nullableString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
