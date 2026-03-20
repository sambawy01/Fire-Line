package api

import (
	"net/http"
	"strconv"

	"github.com/opsnerve/fireline/internal/tenant"
)

// CalculateVendorScores triggers score calculation for all vendors at a location.
// POST /api/v1/vendors/scores/calculate?location_id=<uuid>
func (h *VendorHandler) CalculateVendorScores(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	locationID := r.URL.Query().Get("location_id")
	if locationID == "" {
		WriteError(w, http.StatusBadRequest, "VENDOR_MISSING_LOCATION", "location_id is required")
		return
	}
	scores, err := h.svc.CalculateVendorScores(r.Context(), orgID, locationID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "VENDOR_SCORE_CALC_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"scores": scores})
}

// GetVendorScores returns stored vendor scores for a location.
// GET /api/v1/vendors/scores?location_id=<uuid>
func (h *VendorHandler) GetVendorScores(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	locationID := r.URL.Query().Get("location_id")
	if locationID == "" {
		WriteError(w, http.StatusBadRequest, "VENDOR_MISSING_LOCATION", "location_id is required")
		return
	}
	scores, err := h.svc.GetVendorScores(r.Context(), orgID, locationID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "VENDOR_SCORES_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"scores": scores})
}

// GetVendorScorecard returns a detailed scorecard for a single vendor.
// GET /api/v1/vendors/scorecard?location_id=<uuid>&vendor_name=<name>
func (h *VendorHandler) GetVendorScorecard(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	locationID := r.URL.Query().Get("location_id")
	if locationID == "" {
		WriteError(w, http.StatusBadRequest, "VENDOR_MISSING_LOCATION", "location_id is required")
		return
	}
	vendorName := r.URL.Query().Get("vendor_name")
	if vendorName == "" {
		WriteError(w, http.StatusBadRequest, "VENDOR_MISSING_NAME", "vendor_name is required")
		return
	}
	scorecard, err := h.svc.GetVendorScorecard(r.Context(), orgID, locationID, vendorName)
	if err != nil {
		WriteError(w, http.StatusNotFound, "VENDOR_SCORECARD_NOT_FOUND", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, scorecard)
}

// CompareVendors compares vendors supplying the same ingredient.
// GET /api/v1/vendors/compare?location_id=<uuid>&ingredient_id=<uuid>
func (h *VendorHandler) CompareVendors(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	locationID := r.URL.Query().Get("location_id")
	if locationID == "" {
		WriteError(w, http.StatusBadRequest, "VENDOR_MISSING_LOCATION", "location_id is required")
		return
	}
	ingredientID := r.URL.Query().Get("ingredient_id")
	if ingredientID == "" {
		WriteError(w, http.StatusBadRequest, "VENDOR_MISSING_INGREDIENT", "ingredient_id is required")
		return
	}
	comparison, err := h.svc.CompareVendors(r.Context(), orgID, locationID, ingredientID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "VENDOR_COMPARE_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, comparison)
}

// GetPriceTrend returns price history for an ingredient+vendor pair.
// GET /api/v1/vendors/price-trend?location_id=<uuid>&ingredient_id=<uuid>&vendor_name=<name>&months=<int>
func (h *VendorHandler) GetPriceTrend(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	ingredientID := r.URL.Query().Get("ingredient_id")
	if ingredientID == "" {
		WriteError(w, http.StatusBadRequest, "VENDOR_MISSING_INGREDIENT", "ingredient_id is required")
		return
	}
	vendorName := r.URL.Query().Get("vendor_name")
	if vendorName == "" {
		WriteError(w, http.StatusBadRequest, "VENDOR_MISSING_NAME", "vendor_name is required")
		return
	}
	months := 3
	if m := r.URL.Query().Get("months"); m != "" {
		if v, err := strconv.Atoi(m); err == nil && v > 0 {
			months = v
		}
	}
	points, err := h.svc.GetPriceTrend(r.Context(), orgID, ingredientID, vendorName, months)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "VENDOR_PRICE_TREND_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"price_trend": points})
}

// DetectPriceAnomalies detects statistically significant price deviations.
// GET /api/v1/vendors/price-anomalies?location_id=<uuid>
func (h *VendorHandler) DetectPriceAnomalies(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	locationID := r.URL.Query().Get("location_id")
	if locationID == "" {
		WriteError(w, http.StatusBadRequest, "VENDOR_MISSING_LOCATION", "location_id is required")
		return
	}
	anomalies, err := h.svc.DetectPriceAnomalies(r.Context(), orgID, locationID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "VENDOR_ANOMALY_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"anomalies": anomalies})
}

// RecommendVendor recommends the best vendor for a given ingredient.
// GET /api/v1/vendors/recommend?location_id=<uuid>&ingredient_id=<uuid>
func (h *VendorHandler) RecommendVendor(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	locationID := r.URL.Query().Get("location_id")
	if locationID == "" {
		WriteError(w, http.StatusBadRequest, "VENDOR_MISSING_LOCATION", "location_id is required")
		return
	}
	ingredientID := r.URL.Query().Get("ingredient_id")
	if ingredientID == "" {
		WriteError(w, http.StatusBadRequest, "VENDOR_MISSING_INGREDIENT", "ingredient_id is required")
		return
	}
	recs, err := h.svc.RecommendVendor(r.Context(), orgID, locationID, ingredientID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "VENDOR_RECOMMEND_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"recommendations": recs})
}
