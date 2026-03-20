package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/opsnerve/fireline/internal/tenant"
)

// ScoreMenuItems handles POST /api/v1/menu/score
// Computes and persists 5-dimension scores for all active items at a location.
func (h *MenuHandler) ScoreMenuItems(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	var req struct {
		LocationID string `json:"location_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.LocationID == "" {
		WriteError(w, http.StatusBadRequest, "MENU_MISSING_LOCATION", "location_id is required")
		return
	}
	scores, err := h.svc.ScoreMenuItems(r.Context(), orgID, req.LocationID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "MENU_SCORE_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"scores": scores})
}

// GetMenuScores handles GET /api/v1/menu/scores?location_id=<uuid>
// Returns the current stored scores for all active items at a location.
func (h *MenuHandler) GetMenuScores(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	locationID := r.URL.Query().Get("location_id")
	if locationID == "" {
		WriteError(w, http.StatusBadRequest, "MENU_MISSING_LOCATION", "location_id is required")
		return
	}
	scores, err := h.svc.GetMenuItemScores(r.Context(), orgID, locationID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "MENU_SCORES_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"scores": scores})
}

// SetStrategicScore handles PUT /api/v1/menu/scores/{id}/strategic
// Sets a manual strategic override score for a menu item.
func (h *MenuHandler) SetStrategicScore(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	menuItemID := r.PathValue("id")
	if menuItemID == "" {
		WriteError(w, http.StatusBadRequest, "MENU_MISSING_ID", "menu item id is required")
		return
	}
	var req struct {
		Score float64 `json:"score"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "MENU_INVALID_BODY", "invalid request body")
		return
	}
	if err := h.svc.SetStrategicScore(r.Context(), orgID, menuItemID, req.Score); err != nil {
		WriteError(w, http.StatusInternalServerError, "MENU_STRATEGIC_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// SimulatePriceChange handles POST /api/v1/menu/simulate/price
func (h *MenuHandler) SimulatePriceChange(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	var req struct {
		LocationID string `json:"location_id"`
		MenuItemID string `json:"menu_item_id"`
		NewPrice   int64  `json:"new_price"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "MENU_INVALID_BODY", "invalid request body")
		return
	}
	if req.LocationID == "" || req.MenuItemID == "" {
		WriteError(w, http.StatusBadRequest, "MENU_MISSING_PARAMS", "location_id and menu_item_id are required")
		return
	}
	result, err := h.svc.SimulatePriceChange(r.Context(), orgID, req.LocationID, req.MenuItemID, req.NewPrice)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "MENU_SIMULATE_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, result)
}

// SimulateItemRemoval handles POST /api/v1/menu/simulate/removal
func (h *MenuHandler) SimulateItemRemoval(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	var req struct {
		LocationID string `json:"location_id"`
		MenuItemID string `json:"menu_item_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "MENU_INVALID_BODY", "invalid request body")
		return
	}
	if req.LocationID == "" || req.MenuItemID == "" {
		WriteError(w, http.StatusBadRequest, "MENU_MISSING_PARAMS", "location_id and menu_item_id are required")
		return
	}
	result, err := h.svc.SimulateItemRemoval(r.Context(), orgID, req.LocationID, req.MenuItemID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "MENU_SIMULATE_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, result)
}

// SimulateIngredientCost handles POST /api/v1/menu/simulate/ingredient-cost
func (h *MenuHandler) SimulateIngredientCost(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	var req struct {
		LocationID   string `json:"location_id"`
		IngredientID string `json:"ingredient_id"`
		NewCostPerUnit int  `json:"new_cost_per_unit"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "MENU_INVALID_BODY", "invalid request body")
		return
	}
	if req.LocationID == "" || req.IngredientID == "" {
		WriteError(w, http.StatusBadRequest, "MENU_MISSING_PARAMS", "location_id and ingredient_id are required")
		return
	}
	result, err := h.svc.SimulateIngredientPriceChange(r.Context(), orgID, req.LocationID, req.IngredientID, req.NewCostPerUnit)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "MENU_SIMULATE_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, result)
}

// GetDependencies handles GET /api/v1/menu/dependencies?location_id=<uuid>
func (h *MenuHandler) GetDependencies(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	locationID := r.URL.Query().Get("location_id")
	if locationID == "" {
		WriteError(w, http.StatusBadRequest, "MENU_MISSING_LOCATION", "location_id is required")
		return
	}
	deps, err := h.svc.GetIngredientDependencies(r.Context(), orgID, locationID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "MENU_DEPS_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"dependencies": deps})
}

// GetCrossSell handles GET /api/v1/menu/cross-sell?location_id=<uuid>&limit=<n>
func (h *MenuHandler) GetCrossSell(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	locationID := r.URL.Query().Get("location_id")
	if locationID == "" {
		WriteError(w, http.StatusBadRequest, "MENU_MISSING_LOCATION", "location_id is required")
		return
	}
	limit := 20
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if n, err := strconv.Atoi(lStr); err == nil && n > 0 {
			limit = n
		}
	}
	pairs, err := h.svc.GetCrossSellAffinity(r.Context(), orgID, locationID, limit)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "MENU_CROSSSELL_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"pairs": pairs})
}
