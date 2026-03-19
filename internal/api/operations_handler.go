package api

import (
	"net/http"

	"github.com/opsnerve/fireline/internal/operations"
	"github.com/opsnerve/fireline/internal/tenant"
)

// OperationsHandler handles operations intelligence API requests.
type OperationsHandler struct {
	svc *operations.Service
}

// NewOperationsHandler creates a new OperationsHandler.
func NewOperationsHandler(svc *operations.Service) *OperationsHandler {
	return &OperationsHandler{svc: svc}
}

// RegisterRoutes registers operations API routes on the given mux.
func (h *OperationsHandler) RegisterRoutes(mux *http.ServeMux, authMW func(http.Handler) http.Handler) {
	mux.Handle("GET /api/v1/operations/summary", authMW(http.HandlerFunc(h.GetSummary)))
	mux.Handle("GET /api/v1/operations/hourly", authMW(http.HandlerFunc(h.GetHourly)))
}

// GetSummary returns operational KPIs for a location.
func (h *OperationsHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	locationID := r.URL.Query().Get("location_id")
	if locationID == "" {
		WriteError(w, http.StatusBadRequest, "OPS_MISSING_LOCATION", "location_id is required")
		return
	}
	from, to := parseDateRange(r)
	summary, err := h.svc.GetSummary(r.Context(), orgID, locationID, from, to)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "OPS_SUMMARY_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, summary)
}

// GetHourly returns hourly order and revenue breakdown for a location.
func (h *OperationsHandler) GetHourly(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	locationID := r.URL.Query().Get("location_id")
	if locationID == "" {
		WriteError(w, http.StatusBadRequest, "OPS_MISSING_LOCATION", "location_id is required")
		return
	}
	from, to := parseDateRange(r)
	hourly, err := h.svc.GetHourly(r.Context(), orgID, locationID, from, to)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "OPS_HOURLY_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"hourly": hourly})
}
