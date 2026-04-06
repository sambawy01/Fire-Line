package api

import (
	"log/slog"
	"encoding/json"
	"net/http"

	"github.com/opsnerve/fireline/internal/tenant"
)

// GetOverloadStatus returns the current overload state for a location.
func (h *OperationsHandler) GetOverloadStatus(w http.ResponseWriter, r *http.Request) {
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
	status, err := h.svc.GetOverloadStatus(r.Context(), orgID, locationID)
	if err != nil {
		slog.Error("ops overload error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "OPS_OVERLOAD_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, status)
}

// ApplyOverloadResponse logs and emits an overload response action.
func (h *OperationsHandler) ApplyOverloadResponse(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	var req struct {
		LocationID string `json:"location_id"`
		ActionType string `json:"action_type"`
		ItemID     string `json:"item_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}
	if req.LocationID == "" || req.ActionType == "" {
		WriteError(w, http.StatusBadRequest, "OPS_MISSING_FIELDS", "location_id and action_type are required")
		return
	}
	if err := h.svc.ApplyOverloadResponse(r.Context(), orgID, req.LocationID, req.ActionType, req.ItemID); err != nil {
		slog.Error("ops overload apply error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "OPS_OVERLOAD_APPLY_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "applied"})
}

// GetHealth returns the operational health score for a location.
func (h *OperationsHandler) GetHealth(w http.ResponseWriter, r *http.Request) {
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
	health, err := h.svc.GetOperationalHealth(r.Context(), orgID, locationID)
	if err != nil {
		slog.Error("ops health error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "OPS_HEALTH_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, health)
}

// GetTicketPriorities returns active tickets sorted by computed priority score.
func (h *OperationsHandler) GetTicketPriorities(w http.ResponseWriter, r *http.Request) {
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
	priorities, err := h.svc.GetTicketPriorities(r.Context(), orgID, locationID)
	if err != nil {
		slog.Error("ops priority error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "OPS_PRIORITY_ERROR", "an internal error occurred")
		return
	}
	WriteList(w, http.StatusOK, "priorities", priorities)
}

// GetRealTimeHorizon returns a live operational snapshot.
func (h *OperationsHandler) GetRealTimeHorizon(w http.ResponseWriter, r *http.Request) {
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
	horizon, err := h.svc.GetRealTimeHorizon(r.Context(), orgID, locationID)
	if err != nil {
		slog.Error("ops horizon rt error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "OPS_HORIZON_RT_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, horizon)
}

// GetShiftHorizon returns the 4-hour shift forecast.
func (h *OperationsHandler) GetShiftHorizon(w http.ResponseWriter, r *http.Request) {
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
	horizon, err := h.svc.GetShiftHorizon(r.Context(), orgID, locationID)
	if err != nil {
		slog.Error("ops horizon shift error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "OPS_HORIZON_SHIFT_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, horizon)
}

// GetDailyHorizon returns today's operational overview.
func (h *OperationsHandler) GetDailyHorizon(w http.ResponseWriter, r *http.Request) {
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
	horizon, err := h.svc.GetDailyHorizon(r.Context(), orgID, locationID)
	if err != nil {
		slog.Error("ops horizon daily error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "OPS_HORIZON_DAILY_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, horizon)
}

// GetWeeklyHorizon returns this week's operational plan.
func (h *OperationsHandler) GetWeeklyHorizon(w http.ResponseWriter, r *http.Request) {
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
	horizon, err := h.svc.GetWeeklyHorizon(r.Context(), orgID, locationID)
	if err != nil {
		slog.Error("ops horizon weekly error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "OPS_HORIZON_WEEKLY_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, horizon)
}

// GetStrategicHorizon returns 30-day trailing trends.
func (h *OperationsHandler) GetStrategicHorizon(w http.ResponseWriter, r *http.Request) {
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
	horizon, err := h.svc.GetStrategicHorizon(r.Context(), orgID, locationID)
	if err != nil {
		slog.Error("ops horizon strategic error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "OPS_HORIZON_STRATEGIC_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, horizon)
}
