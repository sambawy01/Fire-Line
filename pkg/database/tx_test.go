package database_test

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/opsnerve/fireline/internal/tenant"
	"github.com/opsnerve/fireline/pkg/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getTestPool returns a superuser pool for test setup/teardown only.
func getTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://fireline:fireline@localhost:5432/fireline?sslmode=disable"
	}
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		t.Skip("database not available:", err)
	}
	t.Cleanup(func() { pool.Close() })
	return pool
}

// getAppPool returns a pool connecting as fireline_app (non-superuser, RLS enforced).
// This MUST be used for all queries that test RLS behavior.
func getAppPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dbURL := os.Getenv("TEST_APP_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://fireline_app:fireline_app@localhost:5432/fireline?sslmode=disable"
	}
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		t.Skip("database not available:", err)
	}
	t.Cleanup(func() { pool.Close() })
	return pool
}

func TestTenantTx_SetsLocalOrgID(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	pool := getTestPool(t)
	ctx := tenant.WithOrgID(context.Background(), "org-test-123")

	err := database.TenantTx(ctx, pool, func(tx pgx.Tx) error {
		var orgID string
		err := tx.QueryRow(ctx, "SELECT current_setting('app.current_org_id')").Scan(&orgID)
		assert.NoError(t, err)
		assert.Equal(t, "org-test-123", orgID)
		return nil
	})
	require.NoError(t, err)
}

func TestTenantTx_FailsWithoutTenantContext(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	pool := getTestPool(t)
	ctx := context.Background() // no tenant

	err := database.TenantTx(ctx, pool, func(tx pgx.Tx) error {
		return nil
	})
	assert.ErrorIs(t, err, tenant.ErrNoTenant)
}

func TestTenantTx_RollsBackOnError(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	pool := getTestPool(t)
	ctx := tenant.WithOrgID(context.Background(), "org-test-456")

	// Create a temp table for this test
	_, err := pool.Exec(ctx, "CREATE TABLE IF NOT EXISTS _tx_test (id TEXT)")
	require.NoError(t, err)
	t.Cleanup(func() { pool.Exec(context.Background(), "DROP TABLE IF EXISTS _tx_test") })

	// This transaction should roll back
	txErr := database.TenantTx(ctx, pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, "INSERT INTO _tx_test (id) VALUES ('should-not-persist')")
		require.NoError(t, err)
		return assert.AnError // force rollback
	})
	assert.Error(t, txErr)

	// Verify nothing was committed
	var count int
	err = pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM _tx_test").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}
