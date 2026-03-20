package marketing

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/opsnerve/fireline/internal/tenant"
	"github.com/opsnerve/fireline/pkg/database"
)

// LoyaltyMember represents a loyalty program member.
type LoyaltyMember struct {
	MemberID       string    `json:"member_id"`
	GuestID        string    `json:"guest_id"`
	GuestName      string    `json:"guest_name,omitempty"`
	PointsBalance  float64   `json:"points_balance"`
	LifetimePoints float64   `json:"lifetime_points"`
	Tier           string    `json:"tier"`
	JoinedAt       time.Time `json:"joined_at"`
}

// LoyaltyTransaction represents a loyalty points transaction.
type LoyaltyTransaction struct {
	TransactionID string    `json:"transaction_id"`
	Type          string    `json:"type"`
	Points        float64   `json:"points"`
	Description   string    `json:"description"`
	CreatedAt     time.Time `json:"created_at"`
}

// ErrInsufficientPoints is returned when a member lacks enough points to redeem.
var ErrInsufficientPoints = errors.New("insufficient points balance")

// calculateTier returns the loyalty tier for a given lifetime points total.
func calculateTier(lifetimePoints float64) string {
	switch {
	case lifetimePoints >= 5000:
		return "platinum"
	case lifetimePoints >= 2000:
		return "gold"
	case lifetimePoints >= 500:
		return "silver"
	default:
		return "bronze"
	}
}

// EnrollMember enrolls a guest in the loyalty program. Silently succeeds if already enrolled.
func (s *Service) EnrollMember(ctx context.Context, orgID, guestID string) (*LoyaltyMember, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var m LoyaltyMember
	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		row := tx.QueryRow(tenantCtx,
			`INSERT INTO loyalty_members (org_id, guest_id)
			VALUES ($1, $2)
			ON CONFLICT (org_id, guest_id) DO UPDATE SET org_id = EXCLUDED.org_id
			RETURNING member_id, guest_id, points_balance, lifetime_points, tier, joined_at`,
			orgID, guestID,
		)
		return row.Scan(&m.MemberID, &m.GuestID, &m.PointsBalance, &m.LifetimePoints, &m.Tier, &m.JoinedAt)
	})
	if err != nil {
		return nil, fmt.Errorf("enroll member: %w", err)
	}
	return &m, nil
}

// EarnPoints adds points for a guest and recalculates their tier.
func (s *Service) EarnPoints(ctx context.Context, orgID, guestID string, points float64, description string, checkID *string) (*LoyaltyMember, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var m LoyaltyMember
	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		// Find member.
		var memberID string
		var lifetime float64
		err := tx.QueryRow(tenantCtx,
			`SELECT member_id, lifetime_points FROM loyalty_members WHERE org_id = $1 AND guest_id = $2`,
			orgID, guestID,
		).Scan(&memberID, &lifetime)
		if err != nil {
			return fmt.Errorf("find loyalty member: %w", err)
		}

		// Insert transaction.
		_, err = tx.Exec(tenantCtx,
			`INSERT INTO loyalty_transactions (org_id, member_id, type, points, description, check_id)
			VALUES ($1, $2, 'earn', $3, $4, $5)`,
			orgID, memberID, points, description, checkID,
		)
		if err != nil {
			return fmt.Errorf("insert earn transaction: %w", err)
		}

		// Update balance + lifetime + tier.
		newLifetime := lifetime + points
		newTier := calculateTier(newLifetime)
		row := tx.QueryRow(tenantCtx,
			`UPDATE loyalty_members SET
				points_balance = points_balance + $3,
				lifetime_points = lifetime_points + $3,
				tier = $4,
				joined_at = joined_at
			WHERE org_id = $1 AND member_id = $2
			RETURNING member_id, guest_id, points_balance, lifetime_points, tier, joined_at`,
			orgID, memberID, points, newTier,
		)
		return row.Scan(&m.MemberID, &m.GuestID, &m.PointsBalance, &m.LifetimePoints, &m.Tier, &m.JoinedAt)
	})
	if err != nil {
		return nil, fmt.Errorf("earn points: %w", err)
	}
	return &m, nil
}

// RedeemPoints deducts points from a member's balance.
func (s *Service) RedeemPoints(ctx context.Context, orgID, guestID string, points float64, description string) (*LoyaltyMember, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var m LoyaltyMember
	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		// Find member and validate balance.
		var memberID string
		var balance float64
		err := tx.QueryRow(tenantCtx,
			`SELECT member_id, points_balance FROM loyalty_members WHERE org_id = $1 AND guest_id = $2`,
			orgID, guestID,
		).Scan(&memberID, &balance)
		if err != nil {
			return fmt.Errorf("find loyalty member: %w", err)
		}
		if balance < points {
			return ErrInsufficientPoints
		}

		// Insert transaction.
		_, err = tx.Exec(tenantCtx,
			`INSERT INTO loyalty_transactions (org_id, member_id, type, points, description)
			VALUES ($1, $2, 'redeem', $3, $4)`,
			orgID, memberID, points, description,
		)
		if err != nil {
			return fmt.Errorf("insert redeem transaction: %w", err)
		}

		// Update balance.
		row := tx.QueryRow(tenantCtx,
			`UPDATE loyalty_members SET
				points_balance = points_balance - $3,
				joined_at = joined_at
			WHERE org_id = $1 AND member_id = $2
			RETURNING member_id, guest_id, points_balance, lifetime_points, tier, joined_at`,
			orgID, memberID, points,
		)
		return row.Scan(&m.MemberID, &m.GuestID, &m.PointsBalance, &m.LifetimePoints, &m.Tier, &m.JoinedAt)
	})
	if err != nil {
		return nil, fmt.Errorf("redeem points: %w", err)
	}
	return &m, nil
}

// GetMember returns a loyalty member's details including their guest name.
func (s *Service) GetMember(ctx context.Context, orgID, guestID string) (*LoyaltyMember, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var m LoyaltyMember
	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		row := tx.QueryRow(tenantCtx,
			`SELECT lm.member_id, lm.guest_id,
				COALESCE(gp.first_name || ' ' || gp.last_name, gp.first_name, '') AS guest_name,
				lm.points_balance, lm.lifetime_points, lm.tier, lm.joined_at
			FROM loyalty_members lm
			LEFT JOIN guest_profiles gp ON gp.guest_id = lm.guest_id AND gp.org_id = lm.org_id
			WHERE lm.org_id = $1 AND lm.guest_id = $2`,
			orgID, guestID,
		)
		return row.Scan(&m.MemberID, &m.GuestID, &m.GuestName, &m.PointsBalance, &m.LifetimePoints, &m.Tier, &m.JoinedAt)
	})
	if err != nil {
		return nil, fmt.Errorf("get loyalty member: %w", err)
	}
	return &m, nil
}

// ListMembers returns loyalty members for an org, optionally filtered by tier.
func (s *Service) ListMembers(ctx context.Context, orgID, tier string, limit int) ([]LoyaltyMember, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	if limit <= 0 {
		limit = 50
	}
	var members []LoyaltyMember
	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		args := []any{orgID}
		where := "lm.org_id = $1"
		idx := 2
		if tier != "" {
			where += fmt.Sprintf(" AND lm.tier = $%d", idx)
			args = append(args, tier)
			idx++
		}
		args = append(args, limit)
		rows, err := tx.Query(tenantCtx,
			fmt.Sprintf(`SELECT lm.member_id, lm.guest_id,
				COALESCE(gp.first_name || ' ' || gp.last_name, gp.first_name, '') AS guest_name,
				lm.points_balance, lm.lifetime_points, lm.tier, lm.joined_at
			FROM loyalty_members lm
			LEFT JOIN guest_profiles gp ON gp.guest_id = lm.guest_id AND gp.org_id = lm.org_id
			WHERE %s ORDER BY lm.lifetime_points DESC LIMIT $%d`, where, idx),
			args...,
		)
		if err != nil {
			return fmt.Errorf("query loyalty members: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			var m LoyaltyMember
			if err := rows.Scan(&m.MemberID, &m.GuestID, &m.GuestName, &m.PointsBalance, &m.LifetimePoints, &m.Tier, &m.JoinedAt); err != nil {
				return fmt.Errorf("scan loyalty member: %w", err)
			}
			members = append(members, m)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	if members == nil {
		members = []LoyaltyMember{}
	}
	return members, nil
}

// GetTransactionHistory returns recent loyalty transactions for a member.
func (s *Service) GetTransactionHistory(ctx context.Context, orgID, memberID string, limit int) ([]LoyaltyTransaction, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	if limit <= 0 {
		limit = 50
	}
	var txns []LoyaltyTransaction
	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(tenantCtx,
			`SELECT transaction_id, type, points, COALESCE(description, ''), created_at
			FROM loyalty_transactions
			WHERE org_id = $1 AND member_id = $2
			ORDER BY created_at DESC LIMIT $3`,
			orgID, memberID, limit,
		)
		if err != nil {
			return fmt.Errorf("query transactions: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			var t LoyaltyTransaction
			if err := rows.Scan(&t.TransactionID, &t.Type, &t.Points, &t.Description, &t.CreatedAt); err != nil {
				return fmt.Errorf("scan transaction: %w", err)
			}
			txns = append(txns, t)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	if txns == nil {
		txns = []LoyaltyTransaction{}
	}
	return txns, nil
}
