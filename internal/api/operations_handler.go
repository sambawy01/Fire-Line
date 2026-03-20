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

	// Kitchen stations — specific paths before parameterized
	mux.Handle("GET /api/v1/operations/stations", authMW(http.HandlerFunc(h.GetStations)))
	mux.Handle("POST /api/v1/operations/stations/setup", authMW(http.HandlerFunc(h.SetupStations)))
	mux.Handle("GET /api/v1/operations/capacity", authMW(http.HandlerFunc(h.GetCapacity)))
	mux.Handle("GET /api/v1/operations/ticket-time-estimate", authMW(http.HandlerFunc(h.EstimateTicketTime)))

	// Resource profiles
	mux.Handle("GET /api/v1/operations/resource-profiles/{menu_item_id}", authMW(http.HandlerFunc(h.GetResourceProfiles)))
	mux.Handle("PUT /api/v1/operations/resource-profiles/{menu_item_id}", authMW(http.HandlerFunc(h.SetResourceProfiles)))

	// KDS — specific paths before parameterized
	mux.Handle("GET /api/v1/operations/kds/metrics", authMW(http.HandlerFunc(h.GetKDSMetrics)))
	mux.Handle("POST /api/v1/operations/kds/tickets", authMW(http.HandlerFunc(h.CreateKDSTicket)))
	mux.Handle("GET /api/v1/operations/kds/tickets", authMW(http.HandlerFunc(h.GetAllKDSTickets)))
	mux.Handle("GET /api/v1/operations/kds/station/{type}", authMW(http.HandlerFunc(h.GetStationKDSTickets)))
	mux.Handle("PUT /api/v1/operations/kds/items/{id}/bump", authMW(http.HandlerFunc(h.BumpKDSItem)))
	mux.Handle("DELETE /api/v1/operations/kds/tickets/{id}", authMW(http.HandlerFunc(h.CancelKDSTicket)))

	// Overload, health, priority — specific paths first
	mux.Handle("GET /api/v1/operations/overload", authMW(http.HandlerFunc(h.GetOverloadStatus)))
	mux.Handle("POST /api/v1/operations/overload/respond", authMW(http.HandlerFunc(h.ApplyOverloadResponse)))
	mux.Handle("GET /api/v1/operations/health", authMW(http.HandlerFunc(h.GetHealth)))
	mux.Handle("GET /api/v1/operations/priority", authMW(http.HandlerFunc(h.GetTicketPriorities)))

	// Planning horizons — specific paths first
	mux.Handle("GET /api/v1/operations/horizon/realtime", authMW(http.HandlerFunc(h.GetRealTimeHorizon)))
	mux.Handle("GET /api/v1/operations/horizon/shift", authMW(http.HandlerFunc(h.GetShiftHorizon)))
	mux.Handle("GET /api/v1/operations/horizon/daily", authMW(http.HandlerFunc(h.GetDailyHorizon)))
	mux.Handle("GET /api/v1/operations/horizon/weekly", authMW(http.HandlerFunc(h.GetWeeklyHorizon)))
	mux.Handle("GET /api/v1/operations/horizon/strategic", authMW(http.HandlerFunc(h.GetStrategicHorizon)))
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
