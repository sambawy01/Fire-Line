package auth_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/opsnerve/fireline/internal/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_Signup_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	svc := testService(t)
	superPool := getTestPool(t)
	issuer := svc.Issuer()
	handler := auth.NewHandler(svc, issuer)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	slug := fmt.Sprintf("handler-test-%d", time.Now().UnixNano())
	email := fmt.Sprintf("handler-%d@test.com", time.Now().UnixNano())

	body, _ := json.Marshal(map[string]string{
		"org_name":     "Handler Test Org",
		"org_slug":     slug,
		"email":        email,
		"password":     "SecureP@ss123!",
		"display_name": "Handler Test",
	})

	req := httptest.NewRequest("POST", "/api/v1/auth/signup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.NotEmpty(t, resp["org_id"])
	assert.NotEmpty(t, resp["user_id"])
	assert.NotEmpty(t, resp["access_token"])
	assert.NotEmpty(t, resp["refresh_token"])

	// Cleanup
	if orgID, ok := resp["org_id"].(string); ok {
		superPool.Exec(context.Background(), "DELETE FROM refresh_tokens WHERE org_id = $1", orgID)
		superPool.Exec(context.Background(), "DELETE FROM users WHERE org_id = $1", orgID)
		superPool.Exec(context.Background(), "DELETE FROM organizations WHERE org_id = $1", orgID)
	}
}

func TestHandler_Signup_MissingFields(t *testing.T) {
	if testing.Short() {
		t.Skip("requires database")
	}
	svc := testService(t)
	handler := auth.NewHandler(svc, svc.Issuer())

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	body, _ := json.Marshal(map[string]string{
		"email": "only@email.com",
	})

	req := httptest.NewRequest("POST", "/api/v1/auth/signup", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
