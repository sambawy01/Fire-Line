package api

import (
	"encoding/json"
	"net/http"

	"github.com/opsnerve/fireline/internal/auth"
	"github.com/opsnerve/fireline/internal/maintenance"
	"github.com/opsnerve/fireline/internal/tenant"
)

// MaintenanceHandler handles maintenance API requests.
type MaintenanceHandler struct {
	svc *maintenance.Service
}

// NewMaintenanceHandler creates a new MaintenanceHandler.
func NewMaintenanceHandler(svc *maintenance.Service) *MaintenanceHandler {
	return &MaintenanceHandler{svc: svc}
}

// requireRole returns a middleware that restricts access to the given roles.
func requireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role := auth.RoleFrom(r.Context())
			if !allowed[role] {
				WriteError(w, http.StatusForbidden, "AUTH_FORBIDDEN", "insufficient permissions")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// chain applies middleware in order.
func chain(h http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// RegisterRoutes registers maintenance API routes on the given mux.
func (h *MaintenanceHandler) RegisterRoutes(mux *http.ServeMux, authMW func(http.Handler) http.Handler) {
	readRoles := requireRole("shift_manager", "gm", "owner")
	writeRoles := requireRole("gm", "owner")

	// Equipment
	mux.Handle("GET /api/v1/maintenance/equipment", chain(http.HandlerFunc(h.ListEquipment), authMW, readRoles))
	mux.Handle("POST /api/v1/maintenance/equipment", chain(http.HandlerFunc(h.CreateEquipment), authMW, writeRoles))
	mux.Handle("GET /api/v1/maintenance/equipment/{id}/readings", chain(http.HandlerFunc(h.GetEquipmentReadings), authMW, readRoles))
	mux.Handle("GET /api/v1/maintenance/equipment/{id}", chain(http.HandlerFunc(h.GetEquipment), authMW, readRoles))
	mux.Handle("PUT /api/v1/maintenance/equipment/{id}", chain(http.HandlerFunc(h.UpdateEquipment), authMW, writeRoles))

	// Tickets
	mux.Handle("GET /api/v1/maintenance/tickets", chain(http.HandlerFunc(h.ListTickets), authMW, readRoles))
	mux.Handle("POST /api/v1/maintenance/tickets", chain(http.HandlerFunc(h.CreateTicket), authMW, writeRoles))
	mux.Handle("GET /api/v1/maintenance/tickets/{id}", chain(http.HandlerFunc(h.GetTicket), authMW, readRoles))
	mux.Handle("PUT /api/v1/maintenance/tickets/{id}", chain(http.HandlerFunc(h.UpdateTicket), authMW, writeRoles))
	mux.Handle("POST /api/v1/maintenance/tickets/{id}/complete", chain(http.HandlerFunc(h.CompleteTicket), authMW, writeRoles))
	mux.Handle("POST /api/v1/maintenance/tickets/{id}/log", chain(http.HandlerFunc(h.AddLog), authMW, writeRoles))

	// Analytics
	mux.Handle("GET /api/v1/maintenance/overdue", chain(http.HandlerFunc(h.GetOverdue), authMW, readRoles))
	mux.Handle("GET /api/v1/maintenance/stats", chain(http.HandlerFunc(h.GetStats), authMW, readRoles))
}

// ─── Equipment Handlers ─────────────────────────────────────────────────────

func (h *MaintenanceHandler) ListEquipment(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	locationID := r.URL.Query().Get("location_id")
	status := r.URL.Query().Get("status")
	category := r.URL.Query().Get("category")

	equipment, err := h.svc.ListEquipment(r.Context(), orgID, locationID, status, category)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "MAINT_LIST_EQUIPMENT_ERROR", err.Error())
		return
	}
	WriteList(w, http.StatusOK, "equipment", equipment)
}

func (h *MaintenanceHandler) CreateEquipment(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	var input maintenance.EquipmentInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}
	eq, err := h.svc.CreateEquipment(r.Context(), orgID, input)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "MAINT_CREATE_EQUIPMENT_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, eq)
}

func (h *MaintenanceHandler) GetEquipment(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	equipmentID := r.PathValue("id")
	eq, err := h.svc.GetEquipment(r.Context(), orgID, equipmentID)
	if err != nil {
		WriteError(w, http.StatusNotFound, "MAINT_EQUIPMENT_NOT_FOUND", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, eq)
}

func (h *MaintenanceHandler) UpdateEquipment(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	equipmentID := r.PathValue("id")
	var input maintenance.EquipmentInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}
	eq, err := h.svc.UpdateEquipment(r.Context(), orgID, equipmentID, input)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "MAINT_UPDATE_EQUIPMENT_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, eq)
}

// ─── Ticket Handlers ────────────────────────────────────────────────────────

func (h *MaintenanceHandler) ListTickets(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	locationID := r.URL.Query().Get("location_id")
	status := r.URL.Query().Get("status")
	priority := r.URL.Query().Get("priority")

	tickets, err := h.svc.ListTickets(r.Context(), orgID, locationID, status, priority)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "MAINT_LIST_TICKETS_ERROR", err.Error())
		return
	}
	WriteList(w, http.StatusOK, "tickets", tickets)
}

func (h *MaintenanceHandler) CreateTicket(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	var input maintenance.TicketInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}
	ticket, err := h.svc.CreateTicket(r.Context(), orgID, input)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "MAINT_CREATE_TICKET_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, ticket)
}

func (h *MaintenanceHandler) GetTicket(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	ticketID := r.PathValue("id")
	ticket, err := h.svc.GetTicket(r.Context(), orgID, ticketID)
	if err != nil {
		WriteError(w, http.StatusNotFound, "MAINT_TICKET_NOT_FOUND", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, ticket)
}

func (h *MaintenanceHandler) UpdateTicket(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	ticketID := r.PathValue("id")
	var updates map[string]any
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}
	if err := h.svc.UpdateTicket(r.Context(), orgID, ticketID, updates); err != nil {
		WriteError(w, http.StatusInternalServerError, "MAINT_UPDATE_TICKET_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *MaintenanceHandler) CompleteTicket(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	ticketID := r.PathValue("id")
	var req struct {
		Resolution string `json:"resolution"`
		ActualCost int    `json:"actual_cost"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}
	if err := h.svc.CompleteTicket(r.Context(), orgID, ticketID, req.Resolution, req.ActualCost); err != nil {
		WriteError(w, http.StatusInternalServerError, "MAINT_COMPLETE_TICKET_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "completed"})
}

func (h *MaintenanceHandler) AddLog(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	ticketID := r.PathValue("id")
	var req struct {
		EquipmentID string  `json:"equipment_id"`
		Action      string  `json:"action"`
		Notes       *string `json:"notes"`
		Cost        int     `json:"cost"`
		PerformedBy *string `json:"performed_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}
	log, err := h.svc.AddLog(r.Context(), orgID, ticketID, req.EquipmentID, req.Action, req.Notes, req.Cost, req.PerformedBy)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "MAINT_ADD_LOG_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, log)
}

// ─── Analytics Handlers ─────────────────────────────────────────────────────

func (h *MaintenanceHandler) GetOverdue(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	locationID := r.URL.Query().Get("location_id")
	equipment, err := h.svc.GetOverdueMaintenanceEquipment(r.Context(), orgID, locationID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "MAINT_OVERDUE_ERROR", err.Error())
		return
	}
	WriteList(w, http.StatusOK, "equipment", equipment)
}

func (h *MaintenanceHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	locationID := r.URL.Query().Get("location_id")
	stats, err := h.svc.GetMaintenanceStats(r.Context(), orgID, locationID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "MAINT_STATS_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, stats)
}

// GetEquipmentReadings returns sensor readings for equipment.
func (h *MaintenanceHandler) GetEquipmentReadings(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	equipmentID := r.PathValue("id")
	readings, err := h.svc.GetEquipmentReadings(r.Context(), orgID, equipmentID, 50)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "MAINT_READINGS_ERROR", err.Error())
		return
	}
	if readings == nil {
		readings = []maintenance.EquipmentReading{}
	}
	WriteJSON(w, http.StatusOK, map[string]any{"readings": readings})
}
