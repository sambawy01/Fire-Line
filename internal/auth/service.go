package auth

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/opsnerve/fireline/internal/tenant"
	"github.com/opsnerve/fireline/pkg/database"
)

type Service struct {
	pool   *pgxpool.Pool
	issuer *TokenIssuer
}

func NewService(pool *pgxpool.Pool, issuer *TokenIssuer) *Service {
	return &Service{pool: pool, issuer: issuer}
}

func (s *Service) Issuer() *TokenIssuer {
	return s.issuer
}

type SignupRequest struct {
	OrgName     string
	OrgSlug     string
	Email       string
	Password    string
	DisplayName string
}

type SignupResult struct {
	OrgID        string
	UserID       string
	AccessToken  string
	RefreshToken string
}

func (s *Service) Signup(ctx context.Context, req SignupRequest) (*SignupResult, error) {
	if err := ValidatePasswordPolicy(req.Password); err != nil {
		return nil, fmt.Errorf("password policy: %w", err)
	}

	hash, err := HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	var orgID, userID string

	// Signup uses the superuser pool (bypasses RLS) because no tenant exists yet.
	// We create the org first, then set tenant context for the user insert.
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx,
		`INSERT INTO organizations (name, slug) VALUES ($1, $2) RETURNING org_id`,
		req.OrgName, req.OrgSlug,
	).Scan(&orgID)
	if err != nil {
		return nil, fmt.Errorf("create organization: %w", err)
	}

	// Set tenant context for the user insert (used by RLS if pool is non-superuser)
	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_org_id', $1, true)", orgID); err != nil {
		return nil, fmt.Errorf("set tenant context: %w", err)
	}

	err = tx.QueryRow(ctx,
		`INSERT INTO users (org_id, email, password_hash, display_name, role)
		 VALUES ($1, $2, $3, $4, 'owner')
		 RETURNING user_id`,
		orgID, req.Email, hash, req.DisplayName,
	).Scan(&userID)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	// Generate tokens in tenant context
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	accessToken, err := s.issuer.GenerateAccessToken(UserClaims{
		UserID: userID,
		OrgID:  orgID,
		Role:   "owner",
		Email:  req.Email,
	})
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	rt, err := GenerateRefreshToken(tenantCtx, s.pool, userID, orgID, RefreshTokenTTLWeb)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	return &SignupResult{
		OrgID:        orgID,
		UserID:       userID,
		AccessToken:  accessToken,
		RefreshToken: rt.PlainText,
	}, nil
}

type LoginRequest struct {
	Email    string
	Password string
}

type LoginResult struct {
	UserID       string
	OrgID        string
	Role         string
	AccessToken  string
	RefreshToken string
	MFARequired  bool
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*LoginResult, error) {
	// Login is pre-tenant — we look up the user by email across all orgs
	var userID, orgID, role, passwordHash string
	var mfaEnabled bool
	err := s.pool.QueryRow(ctx,
		`SELECT user_id, org_id, role, password_hash, mfa_enabled
		 FROM users WHERE email = $1 AND status = 'active'`,
		req.Email,
	).Scan(&userID, &orgID, &role, &passwordHash, &mfaEnabled)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if !VerifyPassword(passwordHash, req.Password) {
		return nil, fmt.Errorf("invalid credentials")
	}

	if mfaEnabled {
		return &LoginResult{
			UserID:      userID,
			OrgID:       orgID,
			Role:        role,
			MFARequired: true,
		}, nil
	}

	accessToken, err := s.issuer.GenerateAccessToken(UserClaims{
		UserID: userID,
		OrgID:  orgID,
		Role:   role,
		Email:  req.Email,
	})
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	tenantCtx := tenant.WithOrgID(ctx, orgID)
	rt, err := GenerateRefreshToken(tenantCtx, s.pool, userID, orgID, RefreshTokenTTLWeb)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	return &LoginResult{
		UserID:       userID,
		OrgID:        orgID,
		Role:         role,
		AccessToken:  accessToken,
		RefreshToken: rt.PlainText,
	}, nil
}

func (s *Service) RefreshAccessToken(ctx context.Context, refreshTokenPlain string) (string, string, error) {
	rt, err := ValidateRefreshToken(ctx, s.pool, refreshTokenPlain)
	if err != nil {
		return "", "", fmt.Errorf("invalid refresh token: %w", err)
	}

	// Revoke the old token (rotation)
	if err := RevokeRefreshToken(ctx, s.pool, rt.TokenID); err != nil {
		return "", "", fmt.Errorf("revoke old token: %w", err)
	}

	// Look up user for claims
	var role, email string
	err = database.TenantTx(ctx, s.pool, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			`SELECT role, email FROM users WHERE user_id = $1 AND status = 'active'`,
			rt.UserID,
		).Scan(&role, &email)
	})
	if err != nil {
		return "", "", fmt.Errorf("user lookup: %w", err)
	}

	accessToken, err := s.issuer.GenerateAccessToken(UserClaims{
		UserID: rt.UserID,
		OrgID:  rt.OrgID,
		Role:   role,
		Email:  email,
	})
	if err != nil {
		return "", "", fmt.Errorf("generate access token: %w", err)
	}

	newRT, err := GenerateRefreshToken(ctx, s.pool, rt.UserID, rt.OrgID, RefreshTokenTTLWeb)
	if err != nil {
		return "", "", fmt.Errorf("generate new refresh token: %w", err)
	}

	return accessToken, newRT.PlainText, nil
}

type PINLoginRequest struct {
	LocationID string
	PIN        string
}

func (s *Service) PINLogin(ctx context.Context, req PINLoginRequest) (*LoginResult, error) {
	// PIN login is pre-tenant — look up by location
	var employeeID, orgID, role, pinHash string
	var userID *string
	err := s.pool.QueryRow(ctx,
		`SELECT e.employee_id, e.org_id, e.role, e.pin_hash, e.user_id
		 FROM employees e
		 WHERE e.location_id = $1 AND e.status = 'active' AND e.pin_hash IS NOT NULL`,
		req.LocationID,
	).Scan(&employeeID, &orgID, &role, &pinHash, &userID)

	// This is simplified — in reality we'd need to check ALL employees at the location
	// and verify the PIN against each hash. For now, this is a placeholder.
	if err != nil {
		return nil, fmt.Errorf("invalid PIN")
	}

	ok, err := VerifyPIN(pinHash, req.PIN)
	if err != nil || !ok {
		return nil, fmt.Errorf("invalid PIN")
	}

	uid := employeeID
	if userID != nil {
		uid = *userID
	}

	accessToken, err := s.issuer.GenerateAccessToken(UserClaims{
		UserID: uid,
		OrgID:  orgID,
		Role:   role,
	})
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	return &LoginResult{
		UserID:      uid,
		OrgID:       orgID,
		Role:        role,
		AccessToken: accessToken,
	}, nil
}

func (s *Service) Logout(ctx context.Context, refreshTokenPlain string) error {
	rt, err := ValidateRefreshToken(ctx, s.pool, refreshTokenPlain)
	if err != nil {
		// Already invalid/expired — that's fine for logout
		return nil
	}
	return RevokeRefreshToken(ctx, s.pool, rt.TokenID)
}
