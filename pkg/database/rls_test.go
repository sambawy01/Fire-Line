package database_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/opsnerve/fireline/internal/tenant"
	"github.com/opsnerve/fireline/pkg/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRLS_TenantIsolation(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	superPool := getTestPool(t) // superuser for setup/teardown
	appPool := getAppPool(t)    // fireline_app role for RLS-enforced queries
	ctx := context.Background()

	// Setup: create two organizations via superuser (bypasses RLS)
	var orgA, orgB string
	err := superPool.QueryRow(ctx,
		"INSERT INTO organizations (name, slug) VALUES ('Tenant A', 'tenant-a') RETURNING org_id",
	).Scan(&orgA)
	require.NoError(t, err)

	err = superPool.QueryRow(ctx,
		"INSERT INTO organizations (name, slug) VALUES ('Tenant B', 'tenant-b') RETURNING org_id",
	).Scan(&orgB)
	require.NoError(t, err)

	// Insert a location for each org via superuser
	_, err = superPool.Exec(ctx,
		"INSERT INTO locations (org_id, name) VALUES ($1, 'Location A1')", orgA)
	require.NoError(t, err)

	_, err = superPool.Exec(ctx,
		"INSERT INTO locations (org_id, name) VALUES ($1, 'Location B1')", orgB)
	require.NoError(t, err)

	t.Cleanup(func() {
		superPool.Exec(ctx, "DELETE FROM locations WHERE org_id IN ($1, $2)", orgA, orgB)
		superPool.Exec(ctx, "DELETE FROM organizations WHERE org_id IN ($1, $2)", orgA, orgB)
	})

	// Query as Tenant A via fireline_app role: should see ONLY Tenant A's location
	ctxA := tenant.WithOrgID(ctx, orgA)
	err = database.TenantTx(ctxA, appPool, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctxA, "SELECT org_id, name FROM locations")
		require.NoError(t, err)
		defer rows.Close()

		var count int
		for rows.Next() {
			var foundOrgID, name string
			err := rows.Scan(&foundOrgID, &name)
			require.NoError(t, err)
			assert.Equal(t, orgA, foundOrgID, "Tenant A should only see own data")
			count++
		}
		assert.Equal(t, 1, count, "Tenant A should see exactly 1 location")
		return nil
	})
	require.NoError(t, err)

	// Query as Tenant B via fireline_app role: should see ONLY Tenant B's location
	ctxB := tenant.WithOrgID(ctx, orgB)
	err = database.TenantTx(ctxB, appPool, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctxB, "SELECT org_id, name FROM locations")
		require.NoError(t, err)
		defer rows.Close()

		var count int
		for rows.Next() {
			var foundOrgID, name string
			err := rows.Scan(&foundOrgID, &name)
			require.NoError(t, err)
			assert.Equal(t, orgB, foundOrgID, "Tenant B should only see own data")
			count++
		}
		assert.Equal(t, 1, count, "Tenant B should see exactly 1 location")
		return nil
	})
	require.NoError(t, err)
}

func TestRLS_UnsetGUC_FailsClosed(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	appPool := getAppPool(t) // MUST use fireline_app, not superuser

	// Query locations as fireline_app without setting org_id GUC
	// The RLS policy does current_setting('app.current_org_id')::UUID
	// which fails on empty string -> UUID cast, so query should error
	tx, err := appPool.Begin(context.Background())
	require.NoError(t, err)
	defer tx.Rollback(context.Background())

	rows, err := tx.Query(context.Background(), "SELECT * FROM locations")
	if err == nil {
		// pgx may defer errors until row iteration
		for rows.Next() {
			// should not reach here
		}
		err = rows.Err()
		rows.Close()
	}
	assert.Error(t, err, "Query without tenant context on fireline_app should fail (fail-closed)")
}
