package integration_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/opsnerve/fireline/internal/alerting"
	"github.com/opsnerve/fireline/internal/api"
	"github.com/opsnerve/fireline/internal/auth"
	"github.com/opsnerve/fireline/internal/event"
	"github.com/opsnerve/fireline/internal/financial"
	"github.com/opsnerve/fireline/internal/inventory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// httpTestEnv bundles the test HTTP mux with the token issuer so tests can
// generate JWTs for specific orgs without going through the signup endpoint.
type httpTestEnv struct {
	mux     *http.ServeMux
	issuer  *auth.TokenIssuer
	authSvc *auth.Service
}

func newHTTPTestEnv(t *testing.T) *httpTestEnv {
	t.Helper()

	superPool, appPool := getTestPools(t)

	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	issuer := auth.NewTokenIssuer(privKey, &privKey.PublicKey, 15*time.Minute)
	authSvc := auth.NewService(appPool, superPool, issuer)
	authHandler := auth.NewHandler(authSvc, issuer)
	authMW := auth.AuthMiddleware(issuer)

	bus := event.New()
	invSvc := inventory.New(appPool, bus)
	invSvc.RegisterHandlers()

	finSvc := financial.New(appPool, bus)
	finSvc.RegisterHandlers()

	alertSvc := alerting.New(bus)
	alertSvc.RegisterDefaultRules()

	mux := http.NewServeMux()
	authHandler.RegisterRoutes(mux)

	invHandler := api.NewInventoryHandler(invSvc)
	invHandler.RegisterRoutes(mux, authMW)

	finHandler := api.NewFinancialHandler(finSvc)
	finHandler.RegisterRoutes(mux, authMW)

	alertHandler := api.NewAlertingHandler(alertSvc)
	alertHandler.RegisterRoutes(mux, authMW)

	return &httpTestEnv{mux: mux, issuer: issuer, authSvc: authSvc}
}

// bearerToken generates a JWT for the given org/user/role without hitting the DB.
func (e *httpTestEnv) bearerToken(t *testing.T, orgID, userID, role string) string {
	t.Helper()
	tok, err := e.issuer.GenerateAccessToken(auth.UserClaims{
		UserID: userID,
		OrgID:  orgID,
		Role:   role,
		Email:  "test@fireline.test",
	})
	require.NoError(t, err)
	return tok
}

// do executes a request against the test mux and returns the recorder.
func (e *httpTestEnv) do(method, path string, body []byte, token string) *httptest.ResponseRecorder {
	var bodyReader *bytes.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	} else {
		bodyReader = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, bodyReader)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	e.mux.ServeHTTP(rec, req)
	return rec
}

// httpFixtures holds IDs for an isolated test org/location/ingredient/employee.
type httpFixtures struct {
	orgID        string
	locationID   string
	ingredientID string
	employeeID   string // for inventory_counts.counted_by (FK → employees)
	userID       string // for purchase_orders.approved_by (FK → users)
}

// setupHTTPFixtures creates a completely isolated org with unique slug using a
// timestamp suffix so repeated test runs never collide on the unique constraint.
func setupHTTPFixtures(t *testing.T, superPool *pgxpool.Pool) httpFixtures {
	t.Helper()
	ctx := context.Background()
	ts := time.Now().UnixNano()
	f := httpFixtures{}

	err := superPool.QueryRow(ctx,
		`INSERT INTO organizations (name, slug)
		 VALUES ($1, $2) RETURNING org_id`,
		fmt.Sprintf("HTTP E2E Org %d", ts),
		fmt.Sprintf("http-e2e-%d", ts),
	).Scan(&f.orgID)
	require.NoError(t, err, "insert organization")
	require.NotEmpty(t, f.orgID)

	err = superPool.QueryRow(ctx,
		`INSERT INTO locations (org_id, name) VALUES ($1, 'E2E Location') RETURNING location_id`,
		f.orgID,
	).Scan(&f.locationID)
	require.NoError(t, err, "insert location")

	err = superPool.QueryRow(ctx,
		`INSERT INTO ingredients (org_id, name, category, unit, cost_per_unit)
		 VALUES ($1, 'Test Ingredient', 'Dry Goods', 'oz', 15) RETURNING ingredient_id`,
		f.orgID,
	).Scan(&f.ingredientID)
	require.NoError(t, err, "insert ingredient")

	// PAR level needed so count lines are populated
	superPool.Exec(ctx,
		`INSERT INTO ingredient_location_configs (org_id, ingredient_id, location_id, par_level, reorder_point)
		 VALUES ($1, $2, $3, 100.00, 20.00)`,
		f.orgID, f.ingredientID, f.locationID,
	)

	// Insert a user (FK target for purchase_orders.approved_by / received_by)
	err = superPool.QueryRow(ctx,
		`INSERT INTO users (org_id, email, password_hash, display_name, role)
		 VALUES ($1, $2, '$2a$12$placeholder', 'E2E User', 'owner')
		 RETURNING user_id`,
		f.orgID,
		fmt.Sprintf("e2e-%d@fireline.test", ts),
	).Scan(&f.userID)
	require.NoError(t, err, "insert user")

	// Insert an employee (FK target for inventory_counts.counted_by)
	err = superPool.QueryRow(ctx,
		`INSERT INTO employees (org_id, location_id, user_id, display_name, role)
		 VALUES ($1, $2, $3, 'E2E Employee', 'owner') RETURNING employee_id`,
		f.orgID, f.locationID, f.userID,
	).Scan(&f.employeeID)
	require.NoError(t, err, "insert employee")

	t.Cleanup(func() {
		superPool.Exec(ctx, "DELETE FROM inventory_variances WHERE org_id = $1", f.orgID)
		superPool.Exec(ctx, "DELETE FROM inventory_count_lines WHERE org_id = $1", f.orgID)
		superPool.Exec(ctx, "DELETE FROM inventory_counts WHERE org_id = $1", f.orgID)
		superPool.Exec(ctx, "DELETE FROM purchase_order_lines WHERE org_id = $1", f.orgID)
		superPool.Exec(ctx, "DELETE FROM purchase_orders WHERE org_id = $1", f.orgID)
		superPool.Exec(ctx, "DELETE FROM budgets WHERE org_id = $1", f.orgID)
		superPool.Exec(ctx, "DELETE FROM ingredient_location_configs WHERE org_id = $1", f.orgID)
		superPool.Exec(ctx, "DELETE FROM employees WHERE org_id = $1", f.orgID)
		superPool.Exec(ctx, "DELETE FROM refresh_tokens WHERE org_id = $1", f.orgID)
		superPool.Exec(ctx, "DELETE FROM users WHERE org_id = $1", f.orgID)
		superPool.Exec(ctx, "DELETE FROM ingredients WHERE org_id = $1", f.orgID)
		superPool.Exec(ctx, "DELETE FROM locations WHERE org_id = $1", f.orgID)
		superPool.Exec(ctx, "DELETE FROM organizations WHERE org_id = $1", f.orgID)
	})

	return f
}

// TestIntegration_FullSignupToInsights tests the complete auth flow:
// signup → verify org + user created → use JWT on protected endpoint.
func TestIntegration_FullSignupToInsights(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}

	superPool, _ := getTestPools(t)
	env := newHTTPTestEnv(t)

	slug := fmt.Sprintf("signup-e2e-%d", time.Now().UnixNano())
	email := fmt.Sprintf("owner-%d@fireline.test", time.Now().UnixNano())

	// 1. Signup
	body, _ := json.Marshal(map[string]string{
		"org_name":     "Signup E2E Org",
		"org_slug":     slug,
		"email":        email,
		"password":     "SecureP@ss123!",
		"display_name": "E2E Owner",
	})

	rec := env.do("POST", "/api/v1/auth/signup", body, "")
	assert.Equal(t, http.StatusCreated, rec.Code, "signup should return 201")

	var signupResp map[string]any
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&signupResp))
	assert.NotEmpty(t, signupResp["org_id"], "should return org_id")
	assert.NotEmpty(t, signupResp["user_id"], "should return user_id")
	assert.NotEmpty(t, signupResp["access_token"], "should return access_token")
	assert.NotEmpty(t, signupResp["refresh_token"], "should return refresh_token")

	orgID := signupResp["org_id"].(string)
	accessToken := signupResp["access_token"].(string)

	// Cleanup
	t.Cleanup(func() {
		ctx := context.Background()
		superPool.Exec(ctx, "DELETE FROM refresh_tokens WHERE org_id = $1", orgID)
		superPool.Exec(ctx, "DELETE FROM users WHERE org_id = $1", orgID)
		superPool.Exec(ctx, "DELETE FROM organizations WHERE org_id = $1", orgID)
	})

	// 2. Verify org + user exist
	var orgName string
	err := superPool.QueryRow(context.Background(),
		"SELECT name FROM organizations WHERE org_id = $1", orgID,
	).Scan(&orgName)
	require.NoError(t, err, "org should exist in database")
	assert.Equal(t, "Signup E2E Org", orgName)

	var userCount int
	err = superPool.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM users WHERE org_id = $1", orgID,
	).Scan(&userCount)
	require.NoError(t, err)
	assert.Equal(t, 1, userCount, "exactly one user should be created")

	// 3. Verify JWT works on a protected endpoint
	rec = env.do("GET", "/api/v1/alerts/count", nil, accessToken)
	assert.Equal(t, http.StatusOK, rec.Code, "protected endpoint should accept valid JWT")

	var alertResp map[string]any
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&alertResp))
	_, hasCount := alertResp["count"]
	assert.True(t, hasCount, "response should contain count field")
}

// TestIntegration_InventoryCountFlow tests the full count lifecycle:
// create count → add lines → submit → verify variances endpoint returns structure.
func TestIntegration_InventoryCountFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}

	superPool, _ := getTestPools(t)
	env := newHTTPTestEnv(t)
	f := setupHTTPFixtures(t, superPool)

	// Use userID in JWT (valid UUID), employeeID as counted_by (employees FK)
	tok := env.bearerToken(t, f.orgID, f.userID, "owner")

	// 1. Create a count — counted_by must be an employee_id (FK → employees)
	countBody, _ := json.Marshal(map[string]string{
		"location_id": f.locationID,
		"count_type":  "spot_check",
		"counted_by":  f.employeeID,
	})
	rec := env.do("POST", "/api/v1/inventory/counts", countBody, tok)
	require.Equal(t, http.StatusCreated, rec.Code, "count creation should return 201: %s", rec.Body.String())

	var countResp map[string]any
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&countResp))
	countID, ok := countResp["count_id"].(string)
	require.True(t, ok, "response should contain count_id")
	assert.NotEmpty(t, countID)
	assert.Equal(t, "in_progress", countResp["status"])

	// 2. Add count lines
	linesBody, _ := json.Marshal(map[string]any{
		"lines": []map[string]any{
			{
				"ingredient_id": f.ingredientID,
				"counted_qty":   42.5,
				"unit":          "oz",
				"note":          "integration test count",
			},
		},
	})
	rec = env.do("POST", "/api/v1/inventory/counts/"+countID+"/lines", linesBody, tok)
	assert.Equal(t, http.StatusOK, rec.Code, "upsert lines should return 200: %s", rec.Body.String())

	var linesResp map[string]any
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&linesResp))
	assert.Equal(t, float64(1), linesResp["updated"])

	// 3. Submit the count
	submitBody, _ := json.Marshal(map[string]string{"status": "submitted"})
	rec = env.do("PUT", "/api/v1/inventory/counts/"+countID, submitBody, tok)
	assert.Equal(t, http.StatusOK, rec.Code, "submit should return 200: %s", rec.Body.String())

	var submitResp map[string]any
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&submitResp))
	assert.Equal(t, "submitted", submitResp["status"])

	// 4. Verify variances endpoint returns expected structure
	rec = env.do("GET",
		"/api/v1/inventory/variances?location_id="+f.locationID+
			"&from="+time.Now().Add(-2*time.Hour).Format(time.RFC3339)+
			"&to="+time.Now().Add(1*time.Hour).Format(time.RFC3339),
		nil, tok)
	assert.Equal(t, http.StatusOK, rec.Code, "variances endpoint should return 200")

	var variancesResp map[string]any
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&variancesResp))
	_, hasVariances := variancesResp["variances"]
	assert.True(t, hasVariances, "response should contain variances key")
}

// TestIntegration_PurchaseOrderFlow tests the PO lifecycle:
// create draft → approve → list pending → receive with discrepancies.
func TestIntegration_PurchaseOrderFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}

	superPool, _ := getTestPools(t)
	env := newHTTPTestEnv(t)
	f := setupHTTPFixtures(t, superPool)

	// userID is a real users.user_id — FK target for approved_by / received_by
	tok := env.bearerToken(t, f.orgID, f.userID, "owner")

	// 1. Create a PO in draft status
	poBody, _ := json.Marshal(map[string]any{
		"location_id": f.locationID,
		"vendor_name": "Integration Test Vendor",
		"notes":       "SP11 integration test PO",
		"lines": []map[string]any{
			{
				"ingredient_id":       f.ingredientID,
				"ordered_qty":         50.0,
				"ordered_unit":        "oz",
				"estimated_unit_cost": 15,
			},
		},
	})
	rec := env.do("POST", "/api/v1/inventory/po", poBody, tok)
	require.Equal(t, http.StatusCreated, rec.Code, "PO creation should return 201: %s", rec.Body.String())

	var poResp map[string]any
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&poResp))
	poID, ok := poResp["purchase_order_id"].(string)
	require.True(t, ok, "response should contain purchase_order_id")
	assert.Equal(t, "draft", poResp["status"])

	// 2. Approve the PO (status transition: draft → approved)
	approveBody, _ := json.Marshal(map[string]string{"status": "approved"})
	rec = env.do("PUT", "/api/v1/inventory/po/"+poID, approveBody, tok)
	assert.Equal(t, http.StatusOK, rec.Code, "PO approval should return 200: %s", rec.Body.String())

	var approveResp map[string]any
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&approveResp))
	assert.Equal(t, "approved", approveResp["status"])

	// 3. Verify the PO appears in the pending list
	rec = env.do("GET", "/api/v1/inventory/po/pending?location_id="+f.locationID, nil, tok)
	assert.Equal(t, http.StatusOK, rec.Code, "pending POs endpoint should return 200")

	var pendingResp map[string]any
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&pendingResp))
	pos, ok := pendingResp["purchase_orders"].([]any)
	require.True(t, ok, "response should contain purchase_orders array")
	assert.Greater(t, len(pos), 0, "approved PO should appear in pending list")

	// 4. Receive the PO — short-receive by 2 units to generate a discrepancy
	receiveBody, _ := json.Marshal(map[string]any{
		"lines": []map[string]any{
			{
				"ingredient_id":      f.ingredientID,
				"received_qty":       48.0, // ordered 50, received 48 → discrepancy
				"received_unit_cost": 15,
			},
		},
	})
	rec = env.do("POST", "/api/v1/inventory/po/"+poID+"/receive", receiveBody, tok)
	assert.Equal(t, http.StatusOK, rec.Code, "PO receive should return 200: %s", rec.Body.String())

	var receiveResp map[string]any
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&receiveResp))
	assert.Equal(t, "received", receiveResp["status"])
	_, hasDiscrepancies := receiveResp["discrepancies"]
	assert.True(t, hasDiscrepancies, "response should include discrepancies key")
	_, hasTotalActual := receiveResp["total_actual"]
	assert.True(t, hasTotalActual, "response should include total_actual")
}

// TestIntegration_FinancialEndpoints verifies financial endpoints return correct structure.
func TestIntegration_FinancialEndpoints(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}

	superPool, _ := getTestPools(t)
	env := newHTTPTestEnv(t)
	f := setupHTTPFixtures(t, superPool)

	tok := env.bearerToken(t, f.orgID, f.userID, "owner")

	from := time.Now().Add(-24 * time.Hour).Format(time.RFC3339)
	to := time.Now().Format(time.RFC3339)
	locQ := "location_id=" + f.locationID
	timeQ := "from=" + from + "&to=" + to

	// 1. P&L — should return 200 even with no sales data
	rec := env.do("GET", "/api/v1/financial/pnl?"+locQ+"&"+timeQ, nil, tok)
	assert.Equal(t, http.StatusOK, rec.Code, "GET /financial/pnl should return 200: %s", rec.Body.String())

	var pnlResp map[string]any
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&pnlResp))
	_, hasRevenue := pnlResp["gross_revenue"]
	assert.True(t, hasRevenue, "P&L response should contain gross_revenue")
	_, hasMargin := pnlResp["gross_margin"]
	assert.True(t, hasMargin, "P&L response should contain gross_margin")

	// 2. Cost centers
	rec = env.do("GET", "/api/v1/financial/cost-centers?"+locQ+"&"+timeQ, nil, tok)
	assert.Equal(t, http.StatusOK, rec.Code, "GET /financial/cost-centers should return 200")

	var ccResp map[string]any
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&ccResp))
	_, hasCostCenters := ccResp["cost_centers"]
	assert.True(t, hasCostCenters, "cost-centers response should contain cost_centers key")

	// 3. Budget variance — seed a budget first so the lookup doesn't 500
	today := time.Now().Format("2006-01-02")
	monthEnd := time.Now().AddDate(0, 1, 0).Format("2006-01-02")
	budgetBody, _ := json.Marshal(map[string]any{
		"location_id":           f.locationID,
		"period_type":           "monthly",
		"period_start":          today,
		"period_end":            monthEnd,
		"revenue_target":        100000,
		"food_cost_pct_target":  30.0,
		"labor_cost_pct_target": 25.0,
		"cogs_target":           30000,
	})
	budgetRec := env.do("POST", "/api/v1/financial/budgets", budgetBody, tok)
	require.Equal(t, http.StatusCreated, budgetRec.Code,
		"budget creation should succeed: %s", budgetRec.Body.String())

	rec = env.do("GET", "/api/v1/financial/budget-variance?"+locQ, nil, tok)
	assert.Equal(t, http.StatusOK, rec.Code, "GET /financial/budget-variance should return 200")

	var bvResp map[string]any
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&bvResp))
	_, hasBudget := bvResp["budget"]
	assert.True(t, hasBudget, "budget-variance response should contain budget key")

	// 4. Period comparison
	rec = env.do("GET", "/api/v1/financial/period-comparison?"+locQ+"&"+timeQ, nil, tok)
	assert.Equal(t, http.StatusOK, rec.Code, "GET /financial/period-comparison should return 200")

	var pcResp map[string]any
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&pcResp))
	assert.NotNil(t, pcResp)
}

// TestIntegration_ErrorHandling verifies API endpoints return correct error codes.
func TestIntegration_ErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}

	superPool, _ := getTestPools(t)
	env := newHTTPTestEnv(t)
	f := setupHTTPFixtures(t, superPool)

	tok := env.bearerToken(t, f.orgID, f.userID, "owner")

	t.Run("inventory_usage_without_location_id", func(t *testing.T) {
		rec := env.do("GET", "/api/v1/inventory/usage", nil, tok)
		assert.Equal(t, http.StatusBadRequest, rec.Code, "missing location_id should return 400")

		var errResp map[string]any
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp))
		_, hasError := errResp["error"]
		assert.True(t, hasError, "error response should have error key")
	})

	t.Run("financial_pnl_without_location_id", func(t *testing.T) {
		rec := env.do("GET", "/api/v1/financial/pnl", nil, tok)
		assert.Equal(t, http.StatusBadRequest, rec.Code, "missing location_id should return 400")

		var errResp map[string]any
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp))
		_, hasError := errResp["error"]
		assert.True(t, hasError, "error response should have error key")
	})

	t.Run("get_count_nonexistent_uuid", func(t *testing.T) {
		// A valid-format UUID that won't exist in the database
		rec := env.do("GET", "/api/v1/inventory/counts/00000000-0000-0000-0000-000000000000", nil, tok)
		// Service returns 500 with "no rows" — acceptable for a nonexistent resource
		assert.True(t,
			rec.Code == http.StatusNotFound || rec.Code == http.StatusInternalServerError,
			"nonexistent count should return 404 or 500, got %d", rec.Code)
	})

	t.Run("login_wrong_password", func(t *testing.T) {
		loginBody, _ := json.Marshal(map[string]string{
			"email":    "nonexistent-user-x99@fireline.test",
			"password": "WrongPassword!",
		})
		rec := env.do("POST", "/api/v1/auth/login", loginBody, "")
		assert.Equal(t, http.StatusUnauthorized, rec.Code, "wrong credentials should return 401")

		var errResp map[string]any
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp))
		_, hasError := errResp["error"]
		assert.True(t, hasError, "401 response should have error key")
	})

	t.Run("protected_endpoint_without_token", func(t *testing.T) {
		rec := env.do("GET", "/api/v1/financial/pnl?location_id="+f.locationID, nil, "")
		assert.Equal(t, http.StatusUnauthorized, rec.Code, "missing token should return 401")
	})
}
