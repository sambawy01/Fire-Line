package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/opsnerve/fireline/internal/api"
)

func TestGetLocations_MissingTenant(t *testing.T) {
	handler := api.NewLocationHandler(nil)
	req := httptest.NewRequest("GET", "/api/v1/locations", nil)
	w := httptest.NewRecorder()

	handler.GetLocations(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}
