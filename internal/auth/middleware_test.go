package auth_test

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/opsnerve/fireline/internal/auth"
	"github.com/opsnerve/fireline/internal/tenant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthMiddleware_ValidToken(t *testing.T) {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	issuer := auth.NewTokenIssuer(privKey, &privKey.PublicKey, 15*time.Minute)

	claims := auth.UserClaims{
		UserID: "user-abc",
		OrgID:  "org-xyz",
		Role:   "gm",
		Email:  "gm@test.com",
	}
	tokenStr, err := issuer.GenerateAccessToken(claims)
	require.NoError(t, err)

	var capturedOrgID, capturedUserID, capturedRole string
	handler := auth.AuthMiddleware(issuer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedOrgID, _ = tenant.OrgIDFrom(r.Context())
		capturedUserID = auth.UserIDFrom(r.Context())
		capturedRole = auth.RoleFrom(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "org-xyz", capturedOrgID)
	assert.Equal(t, "user-abc", capturedUserID)
	assert.Equal(t, "gm", capturedRole)
}

func TestAuthMiddleware_MissingToken(t *testing.T) {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	issuer := auth.NewTokenIssuer(privKey, &privKey.PublicKey, 15*time.Minute)

	handler := auth.AuthMiddleware(issuer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	issuer := auth.NewTokenIssuer(privKey, &privKey.PublicKey, -1*time.Minute)

	claims := auth.UserClaims{UserID: "u", OrgID: "o", Role: "staff", Email: "s@t.com"}
	tokenStr, err := issuer.GenerateAccessToken(claims)
	require.NoError(t, err)

	handler := auth.AuthMiddleware(issuer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestRequirePermission_Allowed(t *testing.T) {
	handler := auth.RequirePermission("financial:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req = req.WithContext(auth.WithRole(req.Context(), "gm"))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRequirePermission_Denied(t *testing.T) {
	handler := auth.RequirePermission("financial:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req = req.WithContext(auth.WithRole(req.Context(), "staff"))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}
