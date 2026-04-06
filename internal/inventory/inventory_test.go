package inventory_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/opsnerve/fireline/internal/event"
	"github.com/opsnerve/fireline/internal/inventory"
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

type testFixtures struct {
	orgID        string
	locationID   string
	menuItemID   string
	ingredientID string
	recipeID     string
}

func setupInventoryFixtures(t *testing.T, superPool *pgxpool.Pool) testFixtures {
	t.Helper()
	ctx := context.Background()
	f := testFixtures{}

	superPool.QueryRow(ctx,
		`INSERT INTO organizations (name, slug) VALUES ($1, $2) RETURNING org_id`,
		"Inv Test "+t.Name(), "inv-test-"+t.Name(),
	).Scan(&f.orgID)

	superPool.QueryRow(ctx,
		`INSERT INTO locations (org_id, name) VALUES ($1, 'Test Loc') RETURNING location_id`,
		f.orgID,
	).Scan(&f.locationID)

	superPool.QueryRow(ctx,
		`INSERT INTO menu_items (org_id, location_id, external_id, name, category, price, source)
		 VALUES ($1, $2, 'ext-burger', 'Cheeseburger', 'Burgers', 1495, 'toast') RETURNING menu_item_id`,
		f.orgID, f.locationID,
	).Scan(&f.menuItemID)

	superPool.QueryRow(ctx,
		`INSERT INTO ingredients (org_id, name, category, unit, cost_per_unit, prep_yield_factor)
		 VALUES ($1, 'Ground Beef', 'Protein', 'oz', 25, 0.9500) RETURNING ingredient_id`,
		f.orgID,
	).Scan(&f.ingredientID)

	superPool.QueryRow(ctx,
		`INSERT INTO recipes (org_id, menu_item_id, name, yield_quantity, yield_unit)
		 VALUES ($1, $2, 'Cheeseburger Recipe', 1.00, 'ea') RETURNING recipe_id`,
		f.orgID, f.menuItemID,
	).Scan(&f.recipeID)

	superPool.Exec(ctx,
		`INSERT INTO recipe_ingredients (org_id, recipe_id, ingredient_id, quantity, unit)
		 VALUES ($1, $2, $3, 8.0000, 'oz')`,
		f.orgID, f.recipeID, f.ingredientID,
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

func TestMaterializeRecipeExplosion(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	superPool, appPool := getTestPools(t)
	f := setupInventoryFixtures(t, superPool)

	bus := event.New()
	svc := inventory.New(appPool, bus)

	ctx := tenant.WithOrgID(context.Background(), f.orgID)
	err := svc.MaterializeRecipeExplosion(ctx, f.orgID, f.menuItemID)
	require.NoError(t, err)

	// Verify explosion was created
	var qty float64
	err = database.TenantTx(ctx, appPool, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			`SELECT quantity_per_unit FROM recipe_explosion
			 WHERE menu_item_id = $1 AND ingredient_id = $2`,
			f.menuItemID, f.ingredientID,
		).Scan(&qty)
	})
	require.NoError(t, err)
	assert.InDelta(t, 8.0, qty, 0.001, "should be 8oz of ground beef per burger")
}

func TestCalculateTheoreticalUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	superPool, appPool := getTestPools(t)
	f := setupInventoryFixtures(t, superPool)

	bus := event.New()
	svc := inventory.New(appPool, bus)

	// Materialize recipe explosion first
	ctx := tenant.WithOrgID(context.Background(), f.orgID)
	err := svc.MaterializeRecipeExplosion(ctx, f.orgID, f.menuItemID)
	require.NoError(t, err)

	// Create a closed check with 3 burgers
	closedAt := time.Now()
	var checkID string
	superPool.QueryRow(ctx,
		`INSERT INTO checks (org_id, location_id, external_id, order_number, status, channel,
		                     subtotal, tax, total, opened_at, closed_at, source)
		 VALUES ($1, $2, 'order-1', '1001', 'closed', 'dine_in', 4485, 359, 4844, $3, $4, 'toast')
		 RETURNING check_id`,
		f.orgID, f.locationID, closedAt.Add(-30*time.Minute), closedAt,
	).Scan(&checkID)

	superPool.Exec(ctx,
		`INSERT INTO check_items (org_id, check_id, menu_item_id, external_id, name, quantity, unit_price)
		 VALUES ($1, $2, $3, 'item-1', 'Cheeseburger', 3, 1495)`,
		f.orgID, checkID, f.menuItemID,
	)

	// Calculate theoretical usage for the last hour
	usage, err := svc.CalculateTheoreticalUsage(
		ctx, f.orgID, f.locationID,
		time.Now().Add(-1*time.Hour), time.Now().Add(1*time.Hour),
	)
	require.NoError(t, err)
	require.Len(t, usage, 1)
	assert.Equal(t, "Ground Beef", usage[0].IngredientName)
	assert.InDelta(t, 24.0, usage[0].TotalUsed, 0.001, "3 burgers × 8oz = 24oz ground beef")
	assert.Equal(t, int64(600), usage[0].TotalCost, "24oz × $0.25/oz = $6.00")
}

func TestCalculateVariance(t *testing.T) {
	bus := event.New()
	svc := inventory.New(nil, bus)

	theoretical := []inventory.TheoreticalUsage{
		{IngredientID: "i1", IngredientName: "Ground Beef", TotalUsed: 24.0, Unit: "oz", CostPerUnit: 25},
		{IngredientID: "i2", IngredientName: "Cheese", TotalUsed: 12.0, Unit: "oz", CostPerUnit: 15},
	}

	// Actual usage shows 28oz beef used (4oz overage) and 11oz cheese (1oz under)
	actual := map[string]float64{
		"i1": 28.0,
		"i2": 11.0,
	}

	variances := svc.CalculateVariance(theoretical, actual)
	require.Len(t, variances, 2)

	// Beef: 28 - 24 = +4oz overage
	assert.Equal(t, "Ground Beef", variances[0].IngredientName)
	assert.InDelta(t, 4.0, variances[0].VarianceAmount, 0.001)
	assert.InDelta(t, 16.67, variances[0].VariancePercent, 0.1)
	assert.Equal(t, int64(100), variances[0].CostImpact) // 4oz × $0.25

	// Cheese: 11 - 12 = -1oz underage
	assert.Equal(t, "Cheese", variances[1].IngredientName)
	assert.InDelta(t, -1.0, variances[1].VarianceAmount, 0.001)
}

func TestGetPARStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	superPool, appPool := getTestPools(t)
	f := setupInventoryFixtures(t, superPool)

	// Set PAR levels
	ctx := context.Background()
	superPool.Exec(ctx,
		`INSERT INTO ingredient_location_configs (org_id, ingredient_id, location_id, par_level, reorder_point)
		 VALUES ($1, $2, $3, 100.00, 25.00)`,
		f.orgID, f.ingredientID, f.locationID,
	)

	bus := event.New()
	svc := inventory.New(appPool, bus)

	// Current level below reorder point
	currentLevels := map[string]float64{f.ingredientID: 20.0}
	tctx := tenant.WithOrgID(context.Background(), f.orgID)
	status, err := svc.GetPARStatus(tctx, f.orgID, f.locationID, currentLevels)
	require.NoError(t, err)
	require.Len(t, status, 1)

	assert.Equal(t, "Ground Beef", status[0].IngredientName)
	assert.True(t, status[0].NeedsReorder)
	assert.InDelta(t, 80.0, status[0].SuggestedQty, 0.001, "need 80oz to reach PAR of 100")
}

func TestInventory_EventHandler(t *testing.T) {
	bus := event.New()

	var inventoryUpdated bool
	bus.Subscribe("inventory.usage.updated", func(ctx context.Context, env event.Envelope) error {
		inventoryUpdated = true
		return nil
	})

	svc := inventory.New(nil, bus)
	svc.RegisterHandlers()

	bus.Publish(context.Background(), event.Envelope{
		EventID:    "test-evt",
		EventType:  "pipeline.orders.processed",
		OrgID:      "org-1",
		LocationID: "loc-1",
		Source:     "pipeline",
	})

	assert.True(t, inventoryUpdated, "should have fired inventory.usage.updated")
}
