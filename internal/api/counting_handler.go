package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/opsnerve/fireline/internal/auth"
	"github.com/opsnerve/fireline/internal/inventory"
	"github.com/opsnerve/fireline/internal/tenant"
)

// CreateCount handles POST /api/v1/inventory/counts
func (h *InventoryHandler) CreateCount(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	var req struct {
		LocationID string `json:"location_id"`
		CountType  string `json:"count_type"`
		CountedBy  string `json:"counted_by"`
		Category   string `json:"category"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}
	if req.LocationID == "" {
		WriteError(w, http.StatusBadRequest, "COUNT_MISSING_LOCATION", "location_id is required")
		return
	}
	if req.CountType == "" {
		WriteError(w, http.StatusBadRequest, "COUNT_MISSING_TYPE", "count_type is required")
		return
	}
	if req.CountedBy == "" {
		WriteError(w, http.StatusBadRequest, "COUNT_MISSING_COUNTED_BY", "counted_by is required")
		return
	}

	cs, err := h.svc.CreateCount(r.Context(), orgID, req.LocationID, req.CountedBy, req.CountType, req.Category)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "COUNT_CREATE_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, cs)
}

// GetCount handles GET /api/v1/inventory/counts/{id}
func (h *InventoryHandler) GetCount(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	countID := r.PathValue("id")
	if countID == "" {
		WriteError(w, http.StatusBadRequest, "COUNT_MISSING_ID", "count id is required")
		return
	}

	cw, err := h.svc.GetCount(r.Context(), orgID, countID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "COUNT_GET_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, cw)
}

// UpdateCountStatus handles PUT /api/v1/inventory/counts/{id}
// Body: {"status": "submitted"|"approved", "approved_by": "(optional, for approve)"}
func (h *InventoryHandler) UpdateCountStatus(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	countID := r.PathValue("id")
	if countID == "" {
		WriteError(w, http.StatusBadRequest, "COUNT_MISSING_ID", "count id is required")
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}

	switch req.Status {
	case "submitted":
		if err := h.svc.SubmitCount(r.Context(), orgID, countID); err != nil {
			WriteError(w, http.StatusInternalServerError, "COUNT_SUBMIT_ERROR", err.Error())
			return
		}

		// Fetch the count to get locationID and period bounds for variance calculation
		cw, err := h.svc.GetCount(r.Context(), orgID, countID)
		if err != nil {
			// Variance calc is best-effort; count was already submitted
			WriteJSON(w, http.StatusOK, map[string]string{"status": "submitted"})
			return
		}

		periodStart := cw.StartedAt
		periodEnd := time.Now()
		_, _ = h.svc.CalculateCountVariances(r.Context(), orgID, cw.LocationID, countID, periodStart, periodEnd)

		// Auto-generate PO suggestions (best-effort, async)
		locID := cw.LocationID
		go func() {
			if err := h.svc.GenerateSuggestedPOs(context.Background(), orgID, locID, countID); err != nil {
				slog.Error("auto-generate POs failed", "error", err, "count_id", countID)
			}
		}()

		WriteJSON(w, http.StatusOK, map[string]string{"status": "submitted"})

	case "approved":
		approvedBy := auth.UserIDFrom(r.Context())
		if approvedBy == "" {
			WriteError(w, http.StatusBadRequest, "COUNT_MISSING_APPROVER", "could not determine approving user from token")
			return
		}
		if err := h.svc.ApproveCount(r.Context(), orgID, countID, approvedBy); err != nil {
			WriteError(w, http.StatusInternalServerError, "COUNT_APPROVE_ERROR", err.Error())
			return
		}
		WriteJSON(w, http.StatusOK, map[string]string{"status": "approved"})

	default:
		WriteError(w, http.StatusBadRequest, "COUNT_INVALID_STATUS", "status must be 'submitted' or 'approved'")
	}
}

// UpsertCountLines handles POST /api/v1/inventory/counts/{id}/lines
func (h *InventoryHandler) UpsertCountLines(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	countID := r.PathValue("id")
	if countID == "" {
		WriteError(w, http.StatusBadRequest, "COUNT_MISSING_ID", "count id is required")
		return
	}

	var req struct {
		Lines []inventory.CountLineInput `json:"lines"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}
	if len(req.Lines) == 0 {
		WriteError(w, http.StatusBadRequest, "COUNT_MISSING_LINES", "lines array is required and must not be empty")
		return
	}

	updated, err := h.svc.UpsertCountLines(r.Context(), orgID, countID, req.Lines)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "COUNT_LINES_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]int{"updated": updated})
}

// LogWaste handles POST /api/v1/inventory/waste
func (h *InventoryHandler) LogWaste(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	var req struct {
		LocationID string  `json:"location_id"`
		IngredientID string  `json:"ingredient_id"`
		Quantity   float64 `json:"quantity"`
		Unit       string  `json:"unit"`
		Reason     string  `json:"reason"`
		LoggedBy   string  `json:"logged_by"`
		Note       string  `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}
	if req.LocationID == "" {
		WriteError(w, http.StatusBadRequest, "WASTE_MISSING_LOCATION", "location_id is required")
		return
	}

	input := inventory.WasteInput{
		IngredientID: req.IngredientID,
		Quantity:     req.Quantity,
		Unit:         req.Unit,
		Reason:       req.Reason,
		LoggedBy:     req.LoggedBy,
		Note:         req.Note,
	}

	wl, err := h.svc.LogWaste(r.Context(), orgID, req.LocationID, input)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "WASTE_LOG_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, wl)
}

// ListWasteLogs handles GET /api/v1/inventory/waste
func (h *InventoryHandler) ListWasteLogs(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	locationID := r.URL.Query().Get("location_id")
	if locationID == "" {
		WriteError(w, http.StatusBadRequest, "WASTE_MISSING_LOCATION", "location_id is required")
		return
	}

	from, to := parseDateRange(r)
	logs, err := h.svc.ListWasteLogs(r.Context(), orgID, locationID, from, to)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "WASTE_LIST_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"waste_logs": logs})
}

// DeleteWaste handles DELETE /api/v1/inventory/waste/{id} — stub, not yet implemented.
func (h *InventoryHandler) DeleteWaste(w http.ResponseWriter, r *http.Request) {
	WriteError(w, http.StatusNotImplemented, "WASTE_DELETE_NOT_IMPLEMENTED", "waste deletion is not yet implemented")
}

// ListVariances handles GET /api/v1/inventory/variances
func (h *InventoryHandler) ListVariances(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	locationID := r.URL.Query().Get("location_id")
	if locationID == "" {
		WriteError(w, http.StatusBadRequest, "VARIANCE_MISSING_LOCATION", "location_id is required")
		return
	}

	from, to := parseDateRange(r)
	variances, err := h.svc.ListVariances(r.Context(), orgID, locationID, from, to)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "VARIANCE_LIST_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"variances": variances})
}
