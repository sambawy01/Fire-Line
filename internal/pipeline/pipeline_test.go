package pipeline_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/opsnerve/fireline/internal/adapter"
	"github.com/opsnerve/fireline/internal/adapter/toast"
	"github.com/opsnerve/fireline/internal/event"
	"github.com/opsnerve/fireline/internal/pipeline"
	"github.com/opsnerve/fireline/internal/tenant"
	"github.com/opsnerve/fireline/pkg/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getTestPools(t *testing.T) (*pgxpool.Pool, *pgxpool.Pool) {
	t.Helper()
	superURL := os.Getenv("TEST_DATABASE_URL")
	if superURL == "" {
		superURL = "postgres://fireline:fireline@localhost:5432/fireline?sslmode=disable"
	}
	appURL := os.Getenv("TEST_APP_DATABASE_URL")
	if appURL == "" {
		appURL = "postgres://fireline_app:fireline_app@localhost:5432/fireline?sslmode=disable"
	}
	superPool, err := pgxpool.New(context.Background(), superURL)
	require.NoError(t, err)
	appPool, err := pgxpool.New(context.Background(), appURL)
	require.NoError(t, err)
	t.Cleanup(func() {
		superPool.Close()
		appPool.Close()
	})
	return superPool, appPool
}

// setupTestOrg creates a test org+location via superuser, returns orgID and locationID.
func setupTestOrg(t *testing.T, superPool *pgxpool.Pool) (string, string) {
	t.Helper()
	ctx := context.Background()
	var orgID, locID string
	err := superPool.QueryRow(ctx,
		`INSERT INTO organizations (name, slug) VALUES ($1, $2) RETURNING org_id`,
		"Pipeline Test Org "+t.Name(), "pipeline-test-"+t.Name(),
	).Scan(&orgID)
	require.NoError(t, err)

	err = superPool.QueryRow(ctx,
		`INSERT INTO locations (org_id, name) VALUES ($1, $2) RETURNING location_id`,
		orgID, "Test Location",
	).Scan(&locID)
	require.NoError(t, err)

	t.Cleanup(func() {
		superPool.Exec(ctx, "DELETE FROM check_items WHERE org_id = $1", orgID)
		superPool.Exec(ctx, "DELETE FROM checks WHERE org_id = $1", orgID)
		superPool.Exec(ctx, "DELETE FROM item_id_mappings WHERE org_id = $1", orgID)
		superPool.Exec(ctx, "DELETE FROM menu_items WHERE org_id = $1", orgID)
		superPool.Exec(ctx, "DELETE FROM locations WHERE org_id = $1", orgID)
		superPool.Exec(ctx, "DELETE FROM organizations WHERE org_id = $1", orgID)
	})

	return orgID, locID
}

func TestPipeline_OrdersSync(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	superPool, appPool := getTestPools(t)
	orgID, locID := setupTestOrg(t, superPool)

	bus := event.New()
	p := pipeline.New(appPool, bus)
	p.RegisterHandlers()

	// Create Toast adapter and read orders
	ta := toast.New()
	err := ta.Initialize(context.Background(), adapter.Config{
		AdapterType: "toast",
		OrgID:       orgID,
		LocationID:  locID,
	})
	require.NoError(t, err)

	// Sync orders through the pipeline
	cfg := adapter.Config{OrgID: orgID, LocationID: locID, CreatedAt: time.Now().Add(-1 * time.Hour)}
	err = p.SyncOrders(context.Background(), ta, ta.(*toast.ToastAdapter), cfg)
	require.NoError(t, err)

	// Verify checks were written
	ctx := tenant.WithOrgID(context.Background(), orgID)
	var count int
	err = database.TenantTx(ctx, appPool, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			"SELECT COUNT(*) FROM checks WHERE org_id = $1 AND location_id = $2",
			orgID, locID,
		).Scan(&count)
	})
	require.NoError(t, err)
	assert.Greater(t, count, 0, "should have inserted checks")

	// Verify check items
	var itemCount int
	err = database.TenantTx(ctx, appPool, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			"SELECT COUNT(*) FROM check_items WHERE org_id = $1",
			orgID,
		).Scan(&itemCount)
	})
	require.NoError(t, err)
	assert.Greater(t, itemCount, 0, "should have inserted check items")
}

func TestPipeline_MenuSync(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	superPool, appPool := getTestPools(t)
	orgID, locID := setupTestOrg(t, superPool)

	bus := event.New()
	p := pipeline.New(appPool, bus)
	p.RegisterHandlers()

	ta := toast.New()
	err := ta.Initialize(context.Background(), adapter.Config{
		AdapterType: "toast",
		OrgID:       orgID,
		LocationID:  locID,
	})
	require.NoError(t, err)

	cfg := adapter.Config{OrgID: orgID, LocationID: locID}
	err = p.SyncMenu(context.Background(), ta, ta.(*toast.ToastAdapter), cfg)
	require.NoError(t, err)

	// Verify menu items were written
	ctx := tenant.WithOrgID(context.Background(), orgID)
	var count int
	err = database.TenantTx(ctx, appPool, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			"SELECT COUNT(*) FROM menu_items WHERE org_id = $1 AND location_id = $2",
			orgID, locID,
		).Scan(&count)
	})
	require.NoError(t, err)
	assert.Equal(t, 8, count, "should have 8 menu items from Toast mock")
}

func TestPipeline_DownstreamEvents(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	superPool, appPool := getTestPools(t)
	orgID, locID := setupTestOrg(t, superPool)

	bus := event.New()

	// Track downstream events
	var orderProcessed, menuProcessed bool
	bus.Subscribe("pipeline.orders.processed", func(ctx context.Context, env event.Envelope) error {
		orderProcessed = true
		return nil
	})
	bus.Subscribe("pipeline.menu.processed", func(ctx context.Context, env event.Envelope) error {
		menuProcessed = true
		return nil
	})

	p := pipeline.New(appPool, bus)
	p.RegisterHandlers()

	ta := toast.New()
	ta.Initialize(context.Background(), adapter.Config{
		AdapterType: "toast", OrgID: orgID, LocationID: locID,
	})

	cfg := adapter.Config{OrgID: orgID, LocationID: locID, CreatedAt: time.Now().Add(-1 * time.Hour)}
	p.SyncOrders(context.Background(), ta, ta.(*toast.ToastAdapter), cfg)
	p.SyncMenu(context.Background(), ta, ta.(*toast.ToastAdapter), cfg)

	assert.True(t, orderProcessed, "should have fired pipeline.orders.processed")
	assert.True(t, menuProcessed, "should have fired pipeline.menu.processed")
}
