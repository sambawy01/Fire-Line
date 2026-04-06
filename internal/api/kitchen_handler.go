package api

import (
	"log/slog"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/opsnerve/fireline/internal/operations"
	"github.com/opsnerve/fireline/internal/tenant"
)

// RegisterRoutes registers operations API routes including kitchen and KDS endpoints.
// NOTE: This method is defined in kitchen_handler.go and replaces the stub in operations_handler.go.

// GetStations returns all kitchen stations for a location.
func (h *OperationsHandler) GetStations(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	locationID := r.URL.Query().Get("location_id")
	if locationID == "" {
		WriteError(w, http.StatusBadRequest, "KITCHEN_MISSING_LOCATION", "location_id is required")
		return
	}
	stations, err := h.svc.GetStations(r.Context(), orgID, locationID)
	if err != nil {
		slog.Error("kitchen stations error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "KITCHEN_STATIONS_ERROR", "an internal error occurred")
		return
	}
	WriteList(w, http.StatusOK, "stations", stations)
}

// SetupStations creates the 6 default kitchen stations for a location.
func (h *OperationsHandler) SetupStations(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	var req struct {
		LocationID string `json:"location_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.LocationID == "" {
		WriteError(w, http.StatusBadRequest, "KITCHEN_MISSING_LOCATION", "location_id is required")
		return
	}
	if err := h.svc.SetupDefaultStations(r.Context(), orgID, req.LocationID); err != nil {
		slog.Error("kitchen setup error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "KITCHEN_SETUP_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// GetCapacity returns per-station load and overall capacity for a location.
func (h *OperationsHandler) GetCapacity(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	locationID := r.URL.Query().Get("location_id")
	if locationID == "" {
		WriteError(w, http.StatusBadRequest, "KITCHEN_MISSING_LOCATION", "location_id is required")
		return
	}
	capacity, err := h.svc.CalculateCapacity(r.Context(), orgID, locationID)
	if err != nil {
		slog.Error("kitchen capacity error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "KITCHEN_CAPACITY_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, capacity)
}

// EstimateTicketTime estimates total prep time for a comma-separated list of menu item IDs.
func (h *OperationsHandler) EstimateTicketTime(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	locationID := r.URL.Query().Get("location_id")
	rawIDs := r.URL.Query().Get("menu_item_ids")
	if rawIDs == "" {
		WriteError(w, http.StatusBadRequest, "KITCHEN_MISSING_ITEMS", "menu_item_ids is required")
		return
	}
	menuItemIDs := strings.Split(rawIDs, ",")
	estimate, err := h.svc.EstimateTicketTime(r.Context(), orgID, locationID, menuItemIDs)
	if err != nil {
		slog.Error("kitchen estimate error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "KITCHEN_ESTIMATE_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, estimate)
}

// GetResourceProfiles returns all resource profile task sequences for a menu item.
func (h *OperationsHandler) GetResourceProfiles(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	menuItemID := r.PathValue("menu_item_id")
	if menuItemID == "" {
		WriteError(w, http.StatusBadRequest, "KITCHEN_MISSING_ITEM", "menu_item_id is required")
		return
	}
	profiles, err := h.svc.GetResourceProfiles(r.Context(), orgID, menuItemID)
	if err != nil {
		slog.Error("kitchen profiles error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "KITCHEN_PROFILES_ERROR", "an internal error occurred")
		return
	}
	WriteList(w, http.StatusOK, "profiles", profiles)
}

// SetResourceProfiles replaces all resource profiles for a menu item.
func (h *OperationsHandler) SetResourceProfiles(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	menuItemID := r.PathValue("menu_item_id")
	if menuItemID == "" {
		WriteError(w, http.StatusBadRequest, "KITCHEN_MISSING_ITEM", "menu_item_id is required")
		return
	}
	var req struct {
		Profiles []operations.ResourceProfile `json:"profiles"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}
	if err := h.svc.SetResourceProfile(r.Context(), orgID, menuItemID, req.Profiles); err != nil {
		slog.Error("kitchen profiles set error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "KITCHEN_PROFILES_SET_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// CreateKDSTicket creates a KDS ticket from a check.
func (h *OperationsHandler) CreateKDSTicket(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	var req struct {
		LocationID  string `json:"location_id"`
		CheckID     string `json:"check_id"`
		OrderNumber string `json:"order_number"`
		Channel     string `json:"channel"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}
	if req.LocationID == "" || req.CheckID == "" {
		WriteError(w, http.StatusBadRequest, "KDS_MISSING_FIELDS", "location_id and check_id are required")
		return
	}
	ticket, err := h.svc.CreateTicketFromCheck(r.Context(), orgID, req.LocationID, req.CheckID, req.OrderNumber, req.Channel)
	if err != nil {
		slog.Error("kds create error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "KDS_CREATE_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusCreated, ticket)
}

// GetAllKDSTickets returns all active tickets for expo view.
func (h *OperationsHandler) GetAllKDSTickets(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	locationID := r.URL.Query().Get("location_id")
	if locationID == "" {
		WriteError(w, http.StatusBadRequest, "KDS_MISSING_LOCATION", "location_id is required")
		return
	}
	tickets, err := h.svc.GetAllTickets(r.Context(), orgID, locationID)
	if err != nil {
		slog.Error("kds list error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "KDS_LIST_ERROR", "an internal error occurred")
		return
	}
	WriteList(w, http.StatusOK, "tickets", tickets)
}

// GetStationKDSTickets returns active tickets for a specific station type.
func (h *OperationsHandler) GetStationKDSTickets(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	stationType := r.PathValue("type")
	if stationType == "" {
		WriteError(w, http.StatusBadRequest, "KDS_MISSING_STATION", "station type is required")
		return
	}
	locationID := r.URL.Query().Get("location_id")
	if locationID == "" {
		WriteError(w, http.StatusBadRequest, "KDS_MISSING_LOCATION", "location_id is required")
		return
	}
	tickets, err := h.svc.GetStationTickets(r.Context(), orgID, locationID, stationType)
	if err != nil {
		slog.Error("kds station error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "KDS_STATION_ERROR", "an internal error occurred")
		return
	}
	WriteList(w, http.StatusOK, "tickets", tickets)
}

// BumpKDSItem advances a ticket item to a new status.
func (h *OperationsHandler) BumpKDSItem(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	itemID := r.PathValue("id")
	if itemID == "" {
		WriteError(w, http.StatusBadRequest, "KDS_MISSING_ITEM", "item id is required")
		return
	}
	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Status == "" {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "status is required")
		return
	}
	item, err := h.svc.BumpTicketItem(r.Context(), orgID, itemID, req.Status)
	if err != nil {
		slog.Error("kds bump error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "KDS_BUMP_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, item)
}

// CancelKDSTicket cancels a ticket and all its items.
func (h *OperationsHandler) CancelKDSTicket(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	ticketID := r.PathValue("id")
	if ticketID == "" {
		WriteError(w, http.StatusBadRequest, "KDS_MISSING_TICKET", "ticket id is required")
		return
	}
	if err := h.svc.CancelTicket(r.Context(), orgID, ticketID); err != nil {
		slog.Error("kds cancel error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "KDS_CANCEL_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
}

// GetKDSMetrics returns aggregated KDS performance metrics.
func (h *OperationsHandler) GetKDSMetrics(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	locationID := r.URL.Query().Get("location_id")
	if locationID == "" {
		WriteError(w, http.StatusBadRequest, "KDS_MISSING_LOCATION", "location_id is required")
		return
	}
	from, to := parseDateRange(r)
	metrics, err := h.svc.GetKDSMetrics(r.Context(), orgID, locationID, from, to)
	if err != nil {
		slog.Error("kds metrics error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "KDS_METRICS_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, metrics)
}
