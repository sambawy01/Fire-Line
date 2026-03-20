package customer

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/opsnerve/fireline/internal/tenant"
	"github.com/opsnerve/fireline/pkg/database"
)

// GuestProfile is the full customer intelligence record for a guest.
type GuestProfile struct {
	GuestID            string     `json:"guest_id"`
	PrivacyTier        int        `json:"privacy_tier"`
	FirstName          *string    `json:"first_name"`
	Email              *string    `json:"email"`
	Phone              *string    `json:"phone"`
	TotalVisits        int        `json:"total_visits"`
	TotalSpend         int64      `json:"total_spend"`
	AvgCheck           int64      `json:"avg_check"`
	PreferredChannel   *string    `json:"preferred_channel"`
	FavoriteItems      []string   `json:"favorite_items"`
	CLVScore           float64    `json:"clv_score"`
	Segment            *string    `json:"segment"`
	ChurnRisk          *string    `json:"churn_risk"`
	ChurnProbability   *float64   `json:"churn_probability"`
	NextVisitPredicted *string    `json:"next_visit_predicted"`
	LastVisitAt        *time.Time `json:"last_visit_at"`
}

// GuestVisit records a single resolved visit for a guest.
type GuestVisit struct {
	VisitID    string    `json:"visit_id"`
	LocationID string    `json:"location_id"`
	CheckID    *string   `json:"check_id"`
	Channel    string    `json:"channel"`
	Spend      int64     `json:"spend"`
	ItemCount  int       `json:"item_count"`
	VisitedAt  time.Time `json:"visited_at"`
}

// hashToken returns a hex-encoded SHA-256 of the raw token string.
func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", sum)
}

// ResolveGuest looks up or creates a guest profile by hashing the payment
// external_id on the check identified by checkID, then creates a guest_visit
// and refreshes aggregate metrics.
func (s *Service) ResolveGuest(ctx context.Context, orgID, checkID string) (*GuestProfile, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	var profile GuestProfile
	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		// Pull check + payment data to build the visit record.
		var locationID, channel string
		var checkTotal int64
		var paymentToken *string

		err := tx.QueryRow(tenantCtx,
			`SELECT c.location_id, c.channel, c.total,
			        (SELECT p.external_id FROM payments p WHERE p.check_id = c.check_id AND p.status = 'completed' ORDER BY p.created_at LIMIT 1)
			 FROM checks c
			 WHERE c.check_id = $1 AND c.org_id = $2`,
			checkID, orgID,
		).Scan(&locationID, &channel, &checkTotal, &paymentToken)
		if err != nil {
			return fmt.Errorf("resolve check: %w", err)
		}

		// Build the token hash (nil token → use check_id as fallback key).
		var tokenHash string
		if paymentToken != nil && *paymentToken != "" {
			tokenHash = hashToken(*paymentToken)
		} else {
			tokenHash = hashToken("check:" + checkID)
		}

		// Upsert guest_profile.
		var guestID string
		err = tx.QueryRow(tenantCtx,
			`INSERT INTO guest_profiles (org_id, payment_token_hash, preferred_channel, last_visit_at, updated_at)
			 VALUES ($1, $2, $3, now(), now())
			 ON CONFLICT (org_id, payment_token_hash) DO UPDATE
			     SET last_visit_at = now(),
			         updated_at    = now()
			 RETURNING guest_id`,
			orgID, tokenHash, channel,
		).Scan(&guestID)
		if err != nil {
			return fmt.Errorf("upsert guest profile: %w", err)
		}

		// Count items on this check.
		var itemCount int
		if err := tx.QueryRow(tenantCtx,
			`SELECT COALESCE(SUM(quantity), 0) FROM check_items WHERE check_id = $1`,
			checkID,
		).Scan(&itemCount); err != nil {
			itemCount = 0
		}

		// Insert guest_visit (idempotent: skip if this check_id already has a visit).
		if _, err := tx.Exec(tenantCtx,
			`INSERT INTO guest_visits (org_id, guest_id, location_id, check_id, channel, spend, item_count, visited_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, now())
			 ON CONFLICT DO NOTHING`,
			orgID, guestID, locationID, checkID, channel, checkTotal, itemCount,
		); err != nil {
			return fmt.Errorf("insert guest visit: %w", err)
		}

		// Refresh aggregates from all visits.
		if _, err := tx.Exec(tenantCtx,
			`UPDATE guest_profiles gp
			 SET total_visits = v.cnt,
			     total_spend  = v.total,
			     avg_check    = CASE WHEN v.cnt > 0 THEN v.total / v.cnt ELSE 0 END,
			     updated_at   = now()
			 FROM (
			     SELECT COUNT(*)::INT AS cnt, COALESCE(SUM(spend), 0)::BIGINT AS total
			     FROM guest_visits
			     WHERE guest_id = $1
			 ) v
			 WHERE gp.guest_id = $1`,
			guestID,
		); err != nil {
			return fmt.Errorf("update aggregates: %w", err)
		}

		// Read back the profile.
		var favJSON []byte
		var nextVisit *string
		err = tx.QueryRow(tenantCtx,
			`SELECT guest_id, privacy_tier, first_name, email, phone,
			        total_visits, total_spend, avg_check, preferred_channel,
			        favorite_items, clv_score::FLOAT8,
			        segment, churn_risk, churn_probability::FLOAT8,
			        to_char(next_visit_predicted, 'YYYY-MM-DD'), last_visit_at
			 FROM guest_profiles
			 WHERE guest_id = $1`,
			guestID,
		).Scan(
			&profile.GuestID, &profile.PrivacyTier, &profile.FirstName, &profile.Email, &profile.Phone,
			&profile.TotalVisits, &profile.TotalSpend, &profile.AvgCheck, &profile.PreferredChannel,
			&favJSON, &profile.CLVScore,
			&profile.Segment, &profile.ChurnRisk, &profile.ChurnProbability,
			&nextVisit, &profile.LastVisitAt,
		)
		if err != nil {
			return fmt.Errorf("read profile: %w", err)
		}
		profile.NextVisitPredicted = nextVisit

		if favJSON != nil {
			_ = json.Unmarshal(favJSON, &profile.FavoriteItems)
		}
		if profile.FavoriteItems == nil {
			profile.FavoriteItems = []string{}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &profile, nil
}

// GetGuestProfile returns the full profile for a single guest.
func (s *Service) GetGuestProfile(ctx context.Context, orgID, guestID string) (*GuestProfile, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	var profile GuestProfile
	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		var favJSON []byte
		var nextVisit *string
		err := tx.QueryRow(tenantCtx,
			`SELECT guest_id, privacy_tier, first_name, email, phone,
			        total_visits, total_spend, avg_check, preferred_channel,
			        favorite_items, clv_score::FLOAT8,
			        segment, churn_risk, churn_probability::FLOAT8,
			        to_char(next_visit_predicted, 'YYYY-MM-DD'), last_visit_at
			 FROM guest_profiles
			 WHERE guest_id = $1`,
			guestID,
		).Scan(
			&profile.GuestID, &profile.PrivacyTier, &profile.FirstName, &profile.Email, &profile.Phone,
			&profile.TotalVisits, &profile.TotalSpend, &profile.AvgCheck, &profile.PreferredChannel,
			&favJSON, &profile.CLVScore,
			&profile.Segment, &profile.ChurnRisk, &profile.ChurnProbability,
			&nextVisit, &profile.LastVisitAt,
		)
		if err != nil {
			if err == pgx.ErrNoRows {
				return fmt.Errorf("guest not found: %s", guestID)
			}
			return fmt.Errorf("query guest profile: %w", err)
		}
		profile.NextVisitPredicted = nextVisit

		if favJSON != nil {
			_ = json.Unmarshal(favJSON, &profile.FavoriteItems)
		}
		if profile.FavoriteItems == nil {
			profile.FavoriteItems = []string{}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

// ListGuests returns a paginated list of guest profiles for the org, optionally
// filtered by locationID (guests who have visited that location). sortBy may be
// "clv_score", "total_spend", "total_visits", or "last_visit_at".
func (s *Service) ListGuests(ctx context.Context, orgID, locationID, sortBy string, limit, offset int) ([]GuestProfile, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	// Validate / default sort column to prevent SQL injection.
	allowedSort := map[string]string{
		"clv_score":    "clv_score DESC",
		"total_spend":  "total_spend DESC",
		"total_visits": "total_visits DESC",
		"last_visit_at": "last_visit_at DESC NULLS LAST",
		"":             "clv_score DESC",
	}
	orderClause, ok := allowedSort[sortBy]
	if !ok {
		orderClause = "clv_score DESC"
	}

	if limit <= 0 || limit > 200 {
		limit = 50
	}

	var profiles []GuestProfile
	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		var rows pgx.Rows
		var err error

		if locationID != "" {
			rows, err = tx.Query(tenantCtx,
				`SELECT gp.guest_id, gp.privacy_tier, gp.first_name, gp.email, gp.phone,
				        gp.total_visits, gp.total_spend, gp.avg_check, gp.preferred_channel,
				        gp.favorite_items, gp.clv_score::FLOAT8,
				        gp.segment, gp.churn_risk, gp.churn_probability::FLOAT8,
				        to_char(gp.next_visit_predicted, 'YYYY-MM-DD'), gp.last_visit_at
				 FROM guest_profiles gp
				 WHERE gp.org_id = $1
				   AND EXISTS (SELECT 1 FROM guest_visits gv WHERE gv.guest_id = gp.guest_id AND gv.location_id = $2)
				 ORDER BY `+orderClause+`
				 LIMIT $3 OFFSET $4`,
				orgID, locationID, limit, offset,
			)
		} else {
			rows, err = tx.Query(tenantCtx,
				`SELECT guest_id, privacy_tier, first_name, email, phone,
				        total_visits, total_spend, avg_check, preferred_channel,
				        favorite_items, clv_score::FLOAT8,
				        segment, churn_risk, churn_probability::FLOAT8,
				        to_char(next_visit_predicted, 'YYYY-MM-DD'), last_visit_at
				 FROM guest_profiles
				 WHERE org_id = $1
				 ORDER BY `+orderClause+`
				 LIMIT $2 OFFSET $3`,
				orgID, limit, offset,
			)
		}
		if err != nil {
			return fmt.Errorf("query guests: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var p GuestProfile
			var favJSON []byte
			var nextVisit *string
			if err := rows.Scan(
				&p.GuestID, &p.PrivacyTier, &p.FirstName, &p.Email, &p.Phone,
				&p.TotalVisits, &p.TotalSpend, &p.AvgCheck, &p.PreferredChannel,
				&favJSON, &p.CLVScore,
				&p.Segment, &p.ChurnRisk, &p.ChurnProbability,
				&nextVisit, &p.LastVisitAt,
			); err != nil {
				return fmt.Errorf("scan guest row: %w", err)
			}
			p.NextVisitPredicted = nextVisit
			if favJSON != nil {
				_ = json.Unmarshal(favJSON, &p.FavoriteItems)
			}
			if p.FavoriteItems == nil {
				p.FavoriteItems = []string{}
			}
			profiles = append(profiles, p)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	if profiles == nil {
		profiles = []GuestProfile{}
	}
	return profiles, nil
}

// EnrichGuest updates name, email, and/or phone on a guest profile and
// upgrades the privacy tier to 2 (identified) if currently tier 1.
func (s *Service) EnrichGuest(ctx context.Context, orgID, guestID string, name, email, phone *string) (*GuestProfile, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		// Only update fields that are non-nil in the request.
		_, err := tx.Exec(tenantCtx,
			`UPDATE guest_profiles
			 SET first_name   = COALESCE($1, first_name),
			     email        = COALESCE($2, email),
			     phone        = COALESCE($3, phone),
			     privacy_tier = GREATEST(privacy_tier, 2),
			     updated_at   = now()
			 WHERE guest_id = $4 AND org_id = $5`,
			name, email, phone, guestID, orgID,
		)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("enrich guest: %w", err)
	}
	return s.GetGuestProfile(ctx, orgID, guestID)
}
