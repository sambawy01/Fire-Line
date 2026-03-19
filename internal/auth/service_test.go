package auth_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/opsnerve/fireline/internal/auth"
	"github.com/opsnerve/fireline/internal/tenant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testService(t *testing.T) *auth.Service {
	t.Helper()
	appPool := getAppPool(t)   // fireline_app (RLS enforced) for tenant-scoped ops
	adminPool := getTestPool(t) // superuser for signup/login (pre-tenant, bypasses RLS)
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	issuer := auth.NewTokenIssuer(privKey, &privKey.PublicKey, 15*time.Minute)
	return auth.NewService(appPool, adminPool, issuer)
}

func TestService_Signup(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	svc := testService(t)
	superPool := getTestPool(t)

	result, err := svc.Signup(context.Background(), auth.SignupRequest{
		OrgName:     "Test Restaurant",
		OrgSlug:     "test-signup-" + time.Now().Format("150405"),
		Email:       "owner-" + time.Now().Format("150405") + "@test.com",
		Password:    "SecureP@ss123!",
		DisplayName: "Test Owner",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, result.OrgID)
	assert.NotEmpty(t, result.UserID)
	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)

	t.Cleanup(func() {
		superPool.Exec(context.Background(), "DELETE FROM refresh_tokens WHERE org_id = $1", result.OrgID)
		superPool.Exec(context.Background(), "DELETE FROM users WHERE org_id = $1", result.OrgID)
		superPool.Exec(context.Background(), "DELETE FROM organizations WHERE org_id = $1", result.OrgID)
	})
}

func TestService_Login(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	svc := testService(t)
	superPool := getTestPool(t)

	// First signup
	slug := "test-login-" + time.Now().Format("150405")
	email := "login-" + time.Now().Format("150405") + "@test.com"
	signup, err := svc.Signup(context.Background(), auth.SignupRequest{
		OrgName:     "Login Test",
		OrgSlug:     slug,
		Email:       email,
		Password:    "SecureP@ss123!",
		DisplayName: "Login User",
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		superPool.Exec(context.Background(), "DELETE FROM refresh_tokens WHERE org_id = $1", signup.OrgID)
		superPool.Exec(context.Background(), "DELETE FROM users WHERE org_id = $1", signup.OrgID)
		superPool.Exec(context.Background(), "DELETE FROM organizations WHERE org_id = $1", signup.OrgID)
	})

	// Now login
	result, err := svc.Login(context.Background(), auth.LoginRequest{
		Email:    email,
		Password: "SecureP@ss123!",
	})
	require.NoError(t, err)
	assert.Equal(t, signup.UserID, result.UserID)
	assert.Equal(t, signup.OrgID, result.OrgID)
	assert.Equal(t, "owner", result.Role)
	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)
	assert.False(t, result.MFARequired)
}

func TestService_RefreshAccessToken(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	svc := testService(t)
	superPool := getTestPool(t)

	slug := "test-refresh-" + time.Now().Format("150405")
	email := "refresh-" + time.Now().Format("150405") + "@test.com"
	signup, err := svc.Signup(context.Background(), auth.SignupRequest{
		OrgName:     "Refresh Test",
		OrgSlug:     slug,
		Email:       email,
		Password:    "SecureP@ss123!",
		DisplayName: "Refresh User",
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		superPool.Exec(context.Background(), "DELETE FROM refresh_tokens WHERE org_id = $1", signup.OrgID)
		superPool.Exec(context.Background(), "DELETE FROM users WHERE org_id = $1", signup.OrgID)
		superPool.Exec(context.Background(), "DELETE FROM organizations WHERE org_id = $1", signup.OrgID)
	})

	// Refresh using the token from signup
	ctx := tenant.WithOrgID(context.Background(), signup.OrgID)
	newAccess, newRefresh, err := svc.RefreshAccessToken(ctx, signup.RefreshToken)
	require.NoError(t, err)
	assert.NotEmpty(t, newAccess)
	assert.NotEmpty(t, newRefresh)
	assert.NotEqual(t, signup.RefreshToken, newRefresh) // rotated
}
