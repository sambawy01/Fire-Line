package api

import (
	"encoding/json"
	"net/http"

	"github.com/opsnerve/fireline/internal/auth"
	"github.com/opsnerve/fireline/internal/inventory"
	"github.com/opsnerve/fireline/internal/tenant"
)

// CreatePO handles POST /api/v1/inventory/po
func (h *InventoryHandler) CreatePO(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	var req struct {
		LocationID string                   `json:"location_id"`
		VendorName string                   `json:"vendor_name"`
		Notes      string                   `json:"notes"`
		Lines      []inventory.POLineInput  `json:"lines"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}
	if req.LocationID == "" {
		WriteError(w, http.StatusBadRequest, "PO_MISSING_LOCATION", "location_id is required")
		return
	}
	if req.VendorName == "" {
		WriteError(w, http.StatusBadRequest, "PO_MISSING_VENDOR", "vendor_name is required")
		return
	}

	po, err := h.svc.CreatePO(r.Context(), orgID, req.LocationID, req.VendorName, req.Notes, req.Lines)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "PO_CREATE_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, po)
}

// ListPOs handles GET /api/v1/inventory/po
func (h *InventoryHandler) ListPOs(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	locationID := r.URL.Query().Get("location_id")
	if locationID == "" {
		WriteError(w, http.StatusBadRequest, "PO_MISSING_LOCATION", "location_id is required")
		return
	}
	status := r.URL.Query().Get("status")

	pos, err := h.svc.ListPOs(r.Context(), orgID, locationID, status)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "PO_LIST_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"purchase_orders": pos})
}

// GetPO handles GET /api/v1/inventory/po/{id}
func (h *InventoryHandler) GetPO(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	poID := r.PathValue("id")
	if poID == "" {
		WriteError(w, http.StatusBadRequest, "PO_MISSING_ID", "purchase order id is required")
		return
	}

	po, err := h.svc.GetPO(r.Context(), orgID, poID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "PO_GET_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, po)
}

// UpdatePO handles PUT /api/v1/inventory/po/{id}
// Routes to UpdatePOStatus if body contains "status", or UpdatePODraft if body contains "lines".
func (h *InventoryHandler) UpdatePO(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	poID := r.PathValue("id")
	if poID == "" {
		WriteError(w, http.StatusBadRequest, "PO_MISSING_ID", "purchase order id is required")
		return
	}

	var body map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}

	if statusRaw, ok := body["status"]; ok {
		var newStatus string
		if err := json.Unmarshal(statusRaw, &newStatus); err != nil {
			WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "status must be a string")
			return
		}
		userID := auth.UserIDFrom(r.Context())
		if userID == "" {
			WriteError(w, http.StatusBadRequest, "AUTH_MISSING_USER", "could not determine user from token")
			return
		}
		if err := h.svc.UpdatePOStatus(r.Context(), orgID, poID, newStatus, userID); err != nil {
			WriteError(w, http.StatusInternalServerError, "PO_STATUS_ERROR", err.Error())
			return
		}
		WriteJSON(w, http.StatusOK, map[string]string{"status": newStatus})
		return
	}

	if linesRaw, ok := body["lines"]; ok {
		var lines []inventory.POLineInput
		if err := json.Unmarshal(linesRaw, &lines); err != nil {
			WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "lines must be an array")
			return
		}
		var notes string
		if notesRaw, ok := body["notes"]; ok {
			_ = json.Unmarshal(notesRaw, &notes)
		}
		if err := h.svc.UpdatePODraft(r.Context(), orgID, poID, notes, lines); err != nil {
			WriteError(w, http.StatusInternalServerError, "PO_DRAFT_UPDATE_ERROR", err.Error())
			return
		}
		WriteJSON(w, http.StatusOK, map[string]string{"status": "updated"})
		return
	}

	WriteError(w, http.StatusBadRequest, "PO_INVALID_UPDATE", "body must contain 'status' or 'lines'")
}

// DeletePO handles DELETE /api/v1/inventory/po/{id}
func (h *InventoryHandler) DeletePO(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	poID := r.PathValue("id")
	if poID == "" {
		WriteError(w, http.StatusBadRequest, "PO_MISSING_ID", "purchase order id is required")
		return
	}

	if err := h.svc.DeletePO(r.Context(), orgID, poID); err != nil {
		WriteError(w, http.StatusInternalServerError, "PO_DELETE_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// ListPendingPOs handles GET /api/v1/inventory/po/pending
func (h *InventoryHandler) ListPendingPOs(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	locationID := r.URL.Query().Get("location_id")

	pos, err := h.svc.ListPendingPOs(r.Context(), orgID, locationID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "PO_PENDING_LIST_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"purchase_orders": pos})
}

// ReceivePO handles POST /api/v1/inventory/po/{id}/receive
func (h *InventoryHandler) ReceivePO(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	poID := r.PathValue("id")
	if poID == "" {
		WriteError(w, http.StatusBadRequest, "PO_MISSING_ID", "purchase order id is required")
		return
	}

	var req struct {
		Lines []inventory.ReceiveLineInput `json:"lines"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}

	receivedBy := auth.UserIDFrom(r.Context())
	if receivedBy == "" {
		WriteError(w, http.StatusBadRequest, "AUTH_MISSING_USER", "could not determine user from token")
		return
	}

	discrepancies, totalActual, err := h.svc.ReceivePO(r.Context(), orgID, poID, receivedBy, req.Lines)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "PO_RECEIVE_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{
		"status":        "received",
		"total_actual":  totalActual,
		"discrepancies": discrepancies,
	})
}

// GetPARBreaches handles GET /api/v1/inventory/par-breaches
func (h *InventoryHandler) GetPARBreaches(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	locationID := r.URL.Query().Get("location_id")
	if locationID == "" {
		WriteError(w, http.StatusBadRequest, "PAR_MISSING_LOCATION", "location_id is required")
		return
	}

	breaches, err := h.svc.GetPARBreaches(r.Context(), orgID, locationID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "PAR_BREACHES_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"par_breaches": breaches})
}
