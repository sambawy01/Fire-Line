package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/opsnerve/fireline/pkg/database"
)

const (
	RefreshTokenTTLWeb    = 7 * 24 * time.Hour  // 7 days
	RefreshTokenTTLMobile = 30 * 24 * time.Hour // 30 days
	refreshTokenBytes     = 32
)

type RefreshToken struct {
	TokenID   string
	UserID    string
	OrgID     string
	PlainText string // only set on creation, never stored
	ExpiresAt time.Time
}

func GenerateRefreshToken(ctx context.Context, pool *pgxpool.Pool, userID, orgID string, ttl time.Duration) (*RefreshToken, error) {
	raw := make([]byte, refreshTokenBytes)
	if _, err := rand.Read(raw); err != nil {
		return nil, fmt.Errorf("generate random bytes: %w", err)
	}

	plainText := hex.EncodeToString(raw)
	hash := hashToken(plainText)
	expiresAt := time.Now().Add(ttl)

	var tokenID string
	err := database.TenantTx(ctx, pool, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			`INSERT INTO refresh_tokens (org_id, user_id, token_hash, expires_at)
			 VALUES ($1, $2, $3, $4)
			 RETURNING token_id`,
			orgID, userID, hash, expiresAt,
		).Scan(&tokenID)
	})
	if err != nil {
		return nil, fmt.Errorf("store refresh token: %w", err)
	}

	return &RefreshToken{
		TokenID:   tokenID,
		UserID:    userID,
		OrgID:     orgID,
		PlainText: plainText,
		ExpiresAt: expiresAt,
	}, nil
}

func ValidateRefreshToken(ctx context.Context, pool *pgxpool.Pool, plainText string) (*RefreshToken, error) {
	hash := hashToken(plainText)

	var rt RefreshToken
	var revoked bool
	err := database.TenantTx(ctx, pool, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			`SELECT token_id, user_id, org_id, expires_at, revoked
			 FROM refresh_tokens
			 WHERE token_hash = $1`,
			hash,
		).Scan(&rt.TokenID, &rt.UserID, &rt.OrgID, &rt.ExpiresAt, &revoked)
	})
	if err != nil {
		return nil, fmt.Errorf("validate refresh token: %w", err)
	}

	if revoked {
		// Reuse detection: revoke ALL tokens for this user (possible theft)
		_ = RevokeAllUserTokens(ctx, pool, rt.UserID)
		return nil, fmt.Errorf("refresh token reuse detected — all sessions terminated")
	}

	if time.Now().After(rt.ExpiresAt) {
		return nil, fmt.Errorf("refresh token expired")
	}

	return &rt, nil
}

func RevokeRefreshToken(ctx context.Context, pool *pgxpool.Pool, tokenID string) error {
	return database.TenantTx(ctx, pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx,
			`UPDATE refresh_tokens SET revoked = true, revoked_at = now() WHERE token_id = $1`,
			tokenID,
		)
		return err
	})
}

func RevokeAllUserTokens(ctx context.Context, pool *pgxpool.Pool, userID string) error {
	return database.TenantTx(ctx, pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx,
			`UPDATE refresh_tokens SET revoked = true, revoked_at = now()
			 WHERE user_id = $1 AND revoked = false`,
			userID,
		)
		return err
	})
}

func hashToken(plainText string) string {
	h := sha256.Sum256([]byte(plainText))
	return hex.EncodeToString(h[:])
}
