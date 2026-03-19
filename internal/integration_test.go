package integration_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/opsnerve/fireline/internal/adapter"
	"github.com/opsnerve/fireline/internal/adapter/toast"
	"github.com/opsnerve/fireline/internal/alerting"
	"github.com/opsnerve/fireline/internal/event"
	"github.com/opsnerve/fireline/internal/financial"
	"github.com/opsnerve/fireline/internal/inventory"
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
	t.Cleanup(func() { superPool.Close(); appPool.Close() })
	return superPool, appPool
}

type integrationFixtures struct {
	orgID        string
	locationID   string
	menuItemID   string
	ingredientID string
}

func setupIntegrationFixtures(t *testing.T, superPool *pgxpool.Pool) integrationFixtures {
	t.Helper()
	ctx := context.Background()
	f := integrationFixtures{}

	superPool.QueryRow(ctx,
		`INSERT INTO organizations (name, slug) VALUES ($1, $2) RETURNING org_id`,
		"E2E Test "+t.Name(), "e2e-"+t.Name(),
	).Scan(&f.orgID)

	superPool.QueryRow(ctx,
		`INSERT INTO locations (org_id, name) VALUES ($1, 'E2E Location') RETURNING location_id`,
		f.orgID,
	).Scan(&f.locationID)

	// Create menu item + ingredient + recipe + explosion for COGS calculation
	superPool.QueryRow(ctx,
		`INSERT INTO menu_items (org_id, location_id, external_id, name, category, price, source)
		 VALUES ($1, $2, 'toast-mi-1', 'Cheeseburger', 'Burgers', 1495, 'toast') RETURNING menu_item_id`,
		f.orgID, f.locationID,
	).Scan(&f.menuItemID)

	superPool.QueryRow(ctx,
		`INSERT INTO ingredients (org_id, name, category, unit, cost_per_unit)
		 VALUES ($1, 'Ground Beef', 'Protein', 'oz', 25) RETURNING ingredient_id`,
		f.orgID,
	).Scan(&f.ingredientID)

	var recipeID string
	superPool.QueryRow(ctx,
		`INSERT INTO recipes (org_id, menu_item_id, name) VALUES ($1, $2, 'Burger Recipe') RETURNING recipe_id`,
		f.orgID, f.menuItemID,
	).Scan(&recipeID)

	superPool.Exec(ctx,
		`INSERT INTO recipe_ingredients (org_id, recipe_id, ingredient_id, quantity, unit)
		 VALUES ($1, $2, $3, 8.0, 'oz')`,
		f.orgID, recipeID, f.ingredientID,
	)

	// Set PAR levels
	superPool.Exec(ctx,
		`INSERT INTO ingredient_location_configs (org_id, ingredient_id, location_id, par_level, reorder_point)
		 VALUES ($1, $2, $3, 200.00, 50.00)`,
		f.orgID, f.ingredientID, f.locationID,
	)

	t.Cleanup(func() {
		superPool.Exec(ctx, "DELETE FROM recipe_explosion WHERE org_id = $1", f.orgID)
		superPool.Exec(ctx, "DELETE FROM recipe_ingredients WHERE org_id = $1", f.orgID)
		superPool.Exec(ctx, "DELETE FROM recipes WHERE org_id = $1", f.orgID)
		superPool.Exec(ctx, "DELETE FROM check_items WHERE org_id = $1", f.orgID)
		superPool.Exec(ctx, "DELETE FROM checks WHERE org_id = $1", f.orgID)
		superPool.Exec(ctx, "DELETE FROM ingredient_location_configs WHERE org_id = $1", f.orgID)
		superPool.Exec(ctx, "DELETE FROM ingredients WHERE org_id = $1", f.orgID)
		superPool.Exec(ctx, "DELETE FROM menu_items WHERE org_id = $1", f.orgID)
		superPool.Exec(ctx, "DELETE FROM locations WHERE org_id = $1", f.orgID)
		superPool.Exec(ctx, "DELETE FROM organizations WHERE org_id = $1", f.orgID)
	})

	return f
}

// TestEndToEnd_POS_Pipeline_Intelligence_Alerts verifies the full data flow:
// POS Adapter → Pipeline → Check/Menu Tables → Inventory Intelligence → Financial Intelligence → Alerting
func TestEndToEnd_POS_Pipeline_Intelligence_Alerts(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	superPool, appPool := getTestPools(t)
	f := setupIntegrationFixtures(t, superPool)

	// 1. Set up the full event-driven system
	bus := event.New()
	pipe := pipeline.New(appPool, bus)
	pipe.RegisterHandlers()

	invSvc := inventory.New(appPool, bus)
	invSvc.RegisterHandlers()

	finSvc := financial.New(appPool, bus)
	finSvc.RegisterHandlers()

	alertSvc := alerting.New(bus)
	alertSvc.RegisterDefaultRules()

	// 2. Materialize recipe explosion
	err := invSvc.MaterializeRecipeExplosion(context.Background(), f.orgID, f.menuItemID)
	require.NoError(t, err)

	// 3. Initialize Toast adapter and sync orders
	ta := toast.New()
	err = ta.Initialize(context.Background(), adapter.Config{
		AdapterType: "toast",
		OrgID:       f.orgID,
		LocationID:  f.locationID,
	})
	require.NoError(t, err)

	cfg := adapter.Config{OrgID: f.orgID, LocationID: f.locationID, CreatedAt: time.Now().Add(-1 * time.Hour)}
	err = pipe.SyncOrders(context.Background(), ta, ta.(*toast.ToastAdapter), cfg)
	require.NoError(t, err)

	// 4. Verify checks were written to database
	ctx := tenant.WithOrgID(context.Background(), f.orgID)
	var checkCount int
	err = database.TenantTx(ctx, appPool, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			"SELECT COUNT(*) FROM checks WHERE org_id = $1 AND location_id = $2",
			f.orgID, f.locationID,
		).Scan(&checkCount)
	})
	require.NoError(t, err)
	assert.Greater(t, checkCount, 0, "pipeline should have written checks")

	// 5. Verify financial P&L can be calculated
	pnl, err := finSvc.CalculatePnL(context.Background(), f.orgID, f.locationID,
		time.Now().Add(-2*time.Hour), time.Now().Add(2*time.Hour))
	require.NoError(t, err)
	assert.Greater(t, pnl.GrossRevenue, int64(0), "should have revenue from orders")
	assert.Greater(t, pnl.CheckCount, 0)

	// 6. Verify alerting system received events
	activeAlerts := alertSvc.ActiveCount(f.orgID)
	assert.Greater(t, activeAlerts, 0, "alerting system should have received events")

	t.Logf("E2E results: %d checks, revenue=$%.2f, COGS=$%.2f, margin=%.1f%%, %d alerts",
		pnl.CheckCount, float64(pnl.GrossRevenue)/100, float64(pnl.COGS)/100,
		pnl.GrossMargin, activeAlerts)
}

// TestCrossTenantIsolation_EndToEnd verifies tenant A cannot see tenant B data
// across all modules.
func TestCrossTenantIsolation_EndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	superPool, appPool := getTestPools(t)
	ctx := context.Background()

	// Create two separate orgs
	var orgA, orgB, locA, locB string
	superPool.QueryRow(ctx,
		`INSERT INTO organizations (name, slug) VALUES ('Tenant A', 'tenant-a-e2e') RETURNING org_id`,
	).Scan(&orgA)
	superPool.QueryRow(ctx,
		`INSERT INTO organizations (name, slug) VALUES ('Tenant B', 'tenant-b-e2e') RETURNING org_id`,
	).Scan(&orgB)
	superPool.QueryRow(ctx,
		`INSERT INTO locations (org_id, name) VALUES ($1, 'Loc A') RETURNING location_id`, orgA,
	).Scan(&locA)
	superPool.QueryRow(ctx,
		`INSERT INTO locations (org_id, name) VALUES ($1, 'Loc B') RETURNING location_id`, orgB,
	).Scan(&locB)

	// Insert a check for each org
	superPool.Exec(ctx,
		`INSERT INTO checks (org_id, location_id, external_id, status, channel, subtotal, tax, total, opened_at, closed_at, source)
		 VALUES ($1, $2, 'a-order-1', 'closed', 'dine_in', 5000, 400, 5400, now() - interval '30 min', now(), 'test')`,
		orgA, locA,
	)
	superPool.Exec(ctx,
		`INSERT INTO checks (org_id, location_id, external_id, status, channel, subtotal, tax, total, opened_at, closed_at, source)
		 VALUES ($1, $2, 'b-order-1', 'closed', 'dine_in', 7000, 560, 7560, now() - interval '30 min', now(), 'test')`,
		orgB, locB,
	)

	t.Cleanup(func() {
		superPool.Exec(ctx, "DELETE FROM checks WHERE org_id IN ($1, $2)", orgA, orgB)
		superPool.Exec(ctx, "DELETE FROM locations WHERE org_id IN ($1, $2)", orgA, orgB)
		superPool.Exec(ctx, "DELETE FROM organizations WHERE org_id IN ($1, $2)", orgA, orgB)
	})

	// Tenant A should see only their check
	ctxA := tenant.WithOrgID(ctx, orgA)
	var countA int
	err := database.TenantTx(ctxA, appPool, func(tx pgx.Tx) error {
		return tx.QueryRow(ctxA, "SELECT COUNT(*) FROM checks").Scan(&countA)
	})
	require.NoError(t, err)
	assert.Equal(t, 1, countA, "Tenant A should see exactly 1 check")

	// Tenant B should see only their check
	ctxB := tenant.WithOrgID(ctx, orgB)
	var countB int
	err = database.TenantTx(ctxB, appPool, func(tx pgx.Tx) error {
		return tx.QueryRow(ctxB, "SELECT COUNT(*) FROM checks").Scan(&countB)
	})
	require.NoError(t, err)
	assert.Equal(t, 1, countB, "Tenant B should see exactly 1 check")

	// Financial P&L should be tenant-scoped
	bus := event.New()
	finSvc := financial.New(appPool, bus)

	pnlA, err := finSvc.CalculatePnL(context.Background(), orgA, locA,
		time.Now().Add(-1*time.Hour), time.Now().Add(1*time.Hour))
	require.NoError(t, err)
	assert.Equal(t, int64(5000), pnlA.GrossRevenue, "Tenant A revenue should be $50")

	pnlB, err := finSvc.CalculatePnL(context.Background(), orgB, locB,
		time.Now().Add(-1*time.Hour), time.Now().Add(1*time.Hour))
	require.NoError(t, err)
	assert.Equal(t, int64(7000), pnlB.GrossRevenue, "Tenant B revenue should be $70")
}
