package tenant_test

import (
	"context"
	"testing"

	"github.com/opsnerve/fireline/internal/tenant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrgIDFrom_Success(t *testing.T) {
	ctx := tenant.WithOrgID(context.Background(), "org-123")
	orgID, err := tenant.OrgIDFrom(ctx)
	require.NoError(t, err)
	assert.Equal(t, "org-123", orgID)
}

func TestOrgIDFrom_Missing(t *testing.T) {
	_, err := tenant.OrgIDFrom(context.Background())
	assert.ErrorIs(t, err, tenant.ErrNoTenant)
}

func TestOrgIDFrom_Empty(t *testing.T) {
	ctx := tenant.WithOrgID(context.Background(), "")
	_, err := tenant.OrgIDFrom(ctx)
	assert.ErrorIs(t, err, tenant.ErrNoTenant)
}
