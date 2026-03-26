package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/opsnerve/fireline/internal/auth"
	"github.com/opsnerve/fireline/internal/intelligence"
	"github.com/opsnerve/fireline/internal/tenant"
)

// IntelligenceHandler handles intelligence and surveillance API requests.
type IntelligenceHandler struct {
	svc *intelligence.Service
}

// NewIntelligenceHandler creates a new IntelligenceHandler.
func NewIntelligenceHandler(svc *intelligence.Service) *IntelligenceHandler {
	return &IntelligenceHandler{svc: svc}
}

// RegisterRoutes registers intelligence API routes on the given mux.
func (h *IntelligenceHandler) RegisterRoutes(mux *http.ServeMux, authMW func(http.Handler) http.Handler) {
	readRoles := requireRole("gm", "ops_director", "owner")
	restrictedRoles := requireRole("ops_director", "owner")

	mux.Handle("GET /api/v1/intelligence/anomalies", chain(http.HandlerFunc(h.ListAnomalies), authMW, readRoles))
	mux.Handle("GET /api/v1/intelligence/anomalies/{id}", chain(http.HandlerFunc(h.GetAnomaly), authMW, restrictedRoles))
	mux.Handle("PUT /api/v1/intelligence/anomalies/{id}/resolve", chain(http.HandlerFunc(h.ResolveAnomaly), authMW, restrictedRoles))
	mux.Handle("GET /api/v1/intelligence/investigation/{id}", chain(http.HandlerFunc(h.GetEmployeeTimeline), authMW, restrictedRoles))
}

// ListAnomalies returns anomalies with optional filters.
// GET /api/v1/intelligence/anomalies?location_id=...&status=...&type=...
func (h *IntelligenceHandler) ListAnomalies(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	locationID := r.URL.Query().Get("location_id")
	status := r.URL.Query().Get("status")
	anomalyType := r.URL.Query().Get("type")

	anomalies, err := h.svc.ListAnomalies(r.Context(), orgID, locationID, status, anomalyType)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTEL_LIST_ERROR", err.Error())
		return
	}

	WriteList(w, http.StatusOK, "anomalies", anomalies)
}

// GetAnomaly returns a single anomaly with evidence.
// GET /api/v1/intelligence/anomalies/{id}
func (h *IntelligenceHandler) GetAnomaly(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	anomalyID := r.PathValue("id")
	if anomalyID == "" {
		WriteError(w, http.StatusBadRequest, "INTEL_MISSING_ID", "anomaly id is required")
		return
	}

	anomaly, err := h.svc.GetAnomaly(r.Context(), orgID, anomalyID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTEL_GET_ERROR", err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, anomaly)
}

// ResolveAnomaly updates an anomaly's resolution status.
// PUT /api/v1/intelligence/anomalies/{id}/resolve
func (h *IntelligenceHandler) ResolveAnomaly(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	anomalyID := r.PathValue("id")
	if anomalyID == "" {
		WriteError(w, http.StatusBadRequest, "INTEL_MISSING_ID", "anomaly id is required")
		return
	}

	var input intelligence.ResolveInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "INTEL_INVALID_BODY", "invalid request body")
		return
	}

	// Validate status
	switch input.Status {
	case "confirmed", "false_positive", "resolved":
		// valid
	default:
		WriteError(w, http.StatusBadRequest, "INTEL_INVALID_STATUS", "status must be confirmed, false_positive, or resolved")
		return
	}

	// Set resolved_by from auth context
	input.ResolvedBy = auth.UserIDFrom(r.Context())

	anomaly, err := h.svc.ResolveAnomaly(r.Context(), orgID, anomalyID, input)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTEL_RESOLVE_ERROR", err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, anomaly)
}

// GetEmployeeTimeline returns an aggregated investigative view of an employee.
// GET /api/v1/intelligence/investigation/{id}?days=30
func (h *IntelligenceHandler) GetEmployeeTimeline(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	employeeID := r.PathValue("id")
	if employeeID == "" {
		WriteError(w, http.StatusBadRequest, "INTEL_MISSING_EMPLOYEE", "employee id is required")
		return
	}

	days := 30
	if d := r.URL.Query().Get("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 {
			days = parsed
		}
	}

	timeline, err := h.svc.GetEmployeeTimeline(r.Context(), orgID, employeeID, days)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTEL_TIMELINE_ERROR", err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, timeline)
}
