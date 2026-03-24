package loyverse

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/opsnerve/fireline/internal/adapter"
)

// Handler exposes HTTP endpoints for managing the Loyverse adapter connection.
// It is self-contained and does not depend on the broader FireLine API package.
type Handler struct {
	a *LoyverseAdapter
}

// NewHandler creates a Handler backed by the given LoyverseAdapter.
func NewHandler(a *LoyverseAdapter) *Handler {
	return &Handler{a: a}
}

// RegisterRoutes mounts the Loyverse adapter endpoints on mux.
// authMW should be the same JWT middleware used by all other FireLine routes.
func (h *Handler) RegisterRoutes(mux *http.ServeMux, authMW func(http.Handler) http.Handler) {
	mux.Handle("POST /api/v1/adapters/loyverse/connect",
		authMW(http.HandlerFunc(h.Connect)))
	mux.Handle("GET /api/v1/adapters/loyverse/status",
		authMW(http.HandlerFunc(h.Status)))
	mux.Handle("POST /api/v1/adapters/loyverse/sync",
		authMW(http.HandlerFunc(h.TriggerSync)))
}

// connectRequest is the body for POST /api/v1/adapters/loyverse/connect.
type connectRequest struct {
	APIToken string `json:"api_token"`
	StoreID  string `json:"store_id"`
}

// Connect configures and (re-)initializes the Loyverse adapter.
//
// POST /api/v1/adapters/loyverse/connect
// Body: {"api_token": "...", "store_id": "..."}
func (h *Handler) Connect(w http.ResponseWriter, r *http.Request) {
	var req connectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid JSON body",
		})
		return
	}
	if req.APIToken == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "api_token is required",
		})
		return
	}

	// Shut down any existing polling before re-initializing.
	h.a.mu.Lock()
	if h.a.cancelPoll != nil {
		h.a.cancelPoll()
		h.a.cancelPoll = nil
	}
	h.a.status = adapter.StatusInitializing
	h.a.mu.Unlock()

	cfg := adapter.Config{
		AdapterType: "loyverse",
		Credentials: map[string]string{
			"api_token": req.APIToken,
			"store_id":  req.StoreID,
		},
	}

	if err := h.a.Initialize(r.Context(), cfg); err != nil {
		h.a.setStatus(adapter.StatusErrored)
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("failed to initialize loyverse adapter: %v", err),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "connected",
		"adapter": "loyverse",
	})
}

// Status returns the current health and freshness of the Loyverse adapter.
//
// GET /api/v1/adapters/loyverse/status
func (h *Handler) Status(w http.ResponseWriter, r *http.Request) {
	h.a.mu.RLock()
	status := h.a.status
	freshness := make(map[string]any, len(h.a.freshness))
	for k, v := range h.a.freshness {
		freshness[k] = v
	}
	h.a.mu.RUnlock()

	healthy := status == adapter.StatusActive
	httpStatus := http.StatusOK
	if !healthy {
		httpStatus = http.StatusServiceUnavailable
	}

	writeJSON(w, httpStatus, map[string]any{
		"adapter":   "loyverse",
		"status":    string(status),
		"healthy":   healthy,
		"freshness": freshness,
	})
}

// TriggerSync triggers an immediate full sync of all data types.
// The sync runs in the background; the response returns immediately with 202.
//
// POST /api/v1/adapters/loyverse/sync
func (h *Handler) TriggerSync(w http.ResponseWriter, r *http.Request) {
	h.a.mu.RLock()
	if h.a.status != adapter.StatusActive {
		h.a.mu.RUnlock()
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": "loyverse adapter not active",
		})
		return
	}
	syncer := h.a.syncer
	h.a.mu.RUnlock()

	go func() {
		ctx := context.Background()
		since := time.Now().Add(-24 * time.Hour)
		syncer.runFullSync(ctx)
		_ = since
	}()

	writeJSON(w, http.StatusAccepted, map[string]string{
		"status": "sync_triggered",
	})
}

// writeJSON is a minimal JSON response helper scoped to this package.
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data) //nolint:errcheck
}
