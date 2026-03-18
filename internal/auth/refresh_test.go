package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/opsnerve/fireline/internal/auth"
	"github.com/opsnerve/fireline/internal/tenant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRefreshToken_CreateAndValidate(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	superPool := getTestPool(t)
	appPool := getAppPool(t)

	// Create a test org and user via superuser
	orgID, userID := seedTestUser(t, superPool)
	ctx := tenant.WithOrgID(context.Background(), orgID)

	// Generate refresh token
	rt, err := auth.GenerateRefreshToken(ctx, appPool, userID, orgID, 1*time.Hour)
	require.NoError(t, err)
	assert.NotEmpty(t, rt.PlainText)
	assert.NotEmpty(t, rt.TokenID)

	// Validate the token
	validated, err := auth.ValidateRefreshToken(ctx, appPool, rt.PlainText)
	require.NoError(t, err)
	assert.Equal(t, userID, validated.UserID)
	assert.Equal(t, orgID, validated.OrgID)
}

func TestRefreshToken_RevokeAndReject(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	superPool := getTestPool(t)
	appPool := getAppPool(t)

	orgID, userID := seedTestUser(t, superPool)
	ctx := tenant.WithOrgID(context.Background(), orgID)

	rt, err := auth.GenerateRefreshToken(ctx, appPool, userID, orgID, 1*time.Hour)
	require.NoError(t, err)

	// Revoke
	err = auth.RevokeRefreshToken(ctx, appPool, rt.TokenID)
	require.NoError(t, err)

	// Validate should fail AND revoke all user tokens (reuse detection)
	_, err = auth.ValidateRefreshToken(ctx, appPool, rt.PlainText)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "reuse detected")
}
