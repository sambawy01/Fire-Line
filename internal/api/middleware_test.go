package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/opsnerve/fireline/internal/api"
	"github.com/stretchr/testify/assert"
)

func TestCorrelationID_Generated(t *testing.T) {
	handler := api.CorrelationID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cid := api.CorrelationIDFrom(r.Context())
		assert.NotEmpty(t, cid)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NotEmpty(t, rec.Header().Get("X-Request-ID"))
}

func TestCorrelationID_PassedThrough(t *testing.T) {
	handler := api.CorrelationID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cid := api.CorrelationIDFrom(r.Context())
		assert.Equal(t, "existing-id-123", cid)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", "existing-id-123")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, "existing-id-123", rec.Header().Get("X-Request-ID"))
}

func TestRecovery_CatchesPanic(t *testing.T) {
	handler := api.Recovery(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
