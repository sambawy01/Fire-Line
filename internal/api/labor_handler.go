package api

import (
	"net/http"

	"github.com/opsnerve/fireline/internal/labor"
	"github.com/opsnerve/fireline/internal/tenant"
)

// LaborHandler handles labor intelligence API requests.
type LaborHandler struct {
	svc *labor.Service
}

// NewLaborHandler creates a new LaborHandler.
func NewLaborHandler(svc *labor.Service) *LaborHandler {
	return &LaborHandler{svc: svc}
}

// RegisterRoutes registers the labor API routes on mux, protected by authMW.
func (h *LaborHandler) RegisterRoutes(mux *http.ServeMux, authMW func(http.Handler) http.Handler) {
	mux.Handle("GET /api/v1/labor/summary", authMW(http.HandlerFunc(h.GetSummary)))
	mux.Handle("GET /api/v1/labor/employees", authMW(http.HandlerFunc(h.GetEmployees)))
}

// GetSummary returns location-wide labor cost KPIs for the requested date range.
// GET /api/v1/labor/summary?location_id=...&from=...&to=...
func (h *LaborHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	locationID := r.URL.Query().Get("location_id")
	if locationID == "" {
		WriteError(w, http.StatusBadRequest, "LABOR_MISSING_LOCATION", "location_id is required")
		return
	}
	from, to := parseDateRange(r)
	summary, err := h.svc.GetSummary(r.Context(), orgID, locationID, from, to)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "LABOR_SUMMARY_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, summary)
}

// GetEmployees returns per-employee labor cost and shift detail for the
// requested date range.
// GET /api/v1/labor/employees?location_id=...&from=...&to=...
func (h *LaborHandler) GetEmployees(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	locationID := r.URL.Query().Get("location_id")
	if locationID == "" {
		WriteError(w, http.StatusBadRequest, "LABOR_MISSING_LOCATION", "location_id is required")
		return
	}
	from, to := parseDateRange(r)
	employees, err := h.svc.GetEmployees(r.Context(), orgID, locationID, from, to)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "LABOR_EMPLOYEES_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"employees": employees})
}
