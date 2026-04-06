package api

import (
	"log/slog"
	"encoding/json"
	"net/http"
	"time"

	"github.com/opsnerve/fireline/internal/financial"
	"github.com/opsnerve/fireline/internal/tenant"
)

// CreateBudget handles POST /api/v1/financial/budgets
func (h *FinancialHandler) CreateBudget(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	var b financial.Budget
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}
	if b.LocationID == "" {
		WriteError(w, http.StatusBadRequest, "FINANCIAL_MISSING_LOCATION", "location_id is required")
		return
	}
	if b.PeriodType == "" {
		WriteError(w, http.StatusBadRequest, "FINANCIAL_MISSING_PERIOD_TYPE", "period_type is required")
		return
	}
	if b.PeriodStart == "" || b.PeriodEnd == "" {
		WriteError(w, http.StatusBadRequest, "FINANCIAL_MISSING_PERIOD", "period_start and period_end are required")
		return
	}

	created, err := h.svc.CreateBudget(r.Context(), orgID, b)
	if err != nil {
		slog.Error("financial budget error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "FINANCIAL_BUDGET_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusCreated, created)
}

// ListBudgets handles GET /api/v1/financial/budgets
func (h *FinancialHandler) ListBudgets(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	locationID := r.URL.Query().Get("location_id")
	if locationID == "" {
		WriteError(w, http.StatusBadRequest, "FINANCIAL_MISSING_LOCATION", "location_id is required")
		return
	}
	periodType := r.URL.Query().Get("period_type")

	budgets, err := h.svc.ListBudgets(r.Context(), orgID, locationID, periodType)
	if err != nil {
		slog.Error("financial budget error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "FINANCIAL_BUDGET_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"budgets": budgets})
}

// GetBudgetVariance handles GET /api/v1/financial/budget-variance
func (h *FinancialHandler) GetBudgetVariance(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	locationID := r.URL.Query().Get("location_id")
	if locationID == "" {
		WriteError(w, http.StatusBadRequest, "FINANCIAL_MISSING_LOCATION", "location_id is required")
		return
	}

	dateStr := r.URL.Query().Get("date")
	date := time.Now()
	if dateStr != "" {
		if parsed, err := time.Parse("2006-01-02", dateStr); err == nil {
			date = parsed
		}
	}

	variance, err := h.svc.CalculateBudgetVariance(r.Context(), orgID, locationID, date)
	if err != nil {
		slog.Error("financial variance error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "FINANCIAL_VARIANCE_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, variance)
}

// GetCostCenters handles GET /api/v1/financial/cost-centers
func (h *FinancialHandler) GetCostCenters(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	locationID := r.URL.Query().Get("location_id")
	if locationID == "" {
		WriteError(w, http.StatusBadRequest, "FINANCIAL_MISSING_LOCATION", "location_id is required")
		return
	}

	from, to := parseDateRange(r)
	centers, err := h.svc.GetCostCenterBreakdown(r.Context(), orgID, locationID, from, to)
	if err != nil {
		slog.Error("financial cost center error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "FINANCIAL_COST_CENTER_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"cost_centers": centers, "period_start": from, "period_end": to})
}

// GetTransactionAnomalies handles GET /api/v1/financial/transaction-anomalies
func (h *FinancialHandler) GetTransactionAnomalies(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	locationID := r.URL.Query().Get("location_id")
	if locationID == "" {
		WriteError(w, http.StatusBadRequest, "FINANCIAL_MISSING_LOCATION", "location_id is required")
		return
	}

	dateStr := r.URL.Query().Get("date")
	day := time.Now()
	if dateStr != "" {
		if parsed, err := time.Parse("2006-01-02", dateStr); err == nil {
			day = parsed
		}
	}

	anomalies, err := h.svc.DetectTransactionAnomalies(r.Context(), orgID, locationID, day)
	if err != nil {
		slog.Error("financial tx anomaly error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "FINANCIAL_TX_ANOMALY_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"anomalies": anomalies})
}

// GetDrilldownItems handles GET /api/v1/financial/drilldown/items
func (h *FinancialHandler) GetDrilldownItems(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	locationID := r.URL.Query().Get("location_id")
	if locationID == "" {
		WriteError(w, http.StatusBadRequest, "FINANCIAL_MISSING_LOCATION", "location_id is required")
		return
	}
	category := r.URL.Query().Get("category")

	from, to := parseDateRange(r)
	items, err := h.svc.GetItemCostBreakdown(r.Context(), orgID, locationID, category, from, to)
	if err != nil {
		slog.Error("financial drilldown error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "FINANCIAL_DRILLDOWN_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"items": items, "period_start": from, "period_end": to})
}

// GetDrilldownIngredients handles GET /api/v1/financial/drilldown/ingredients
func (h *FinancialHandler) GetDrilldownIngredients(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	locationID := r.URL.Query().Get("location_id")
	if locationID == "" {
		WriteError(w, http.StatusBadRequest, "FINANCIAL_MISSING_LOCATION", "location_id is required")
		return
	}
	menuItemID := r.URL.Query().Get("menu_item_id")
	if menuItemID == "" {
		WriteError(w, http.StatusBadRequest, "FINANCIAL_MISSING_MENU_ITEM", "menu_item_id is required")
		return
	}

	from, to := parseDateRange(r)
	breakdown, err := h.svc.GetIngredientCostBreakdown(r.Context(), orgID, locationID, menuItemID, from, to)
	if err != nil {
		slog.Error("financial drilldown error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "FINANCIAL_DRILLDOWN_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"ingredients": breakdown, "period_start": from, "period_end": to})
}

// GetDrilldownVendor handles GET /api/v1/financial/drilldown/vendor
func (h *FinancialHandler) GetDrilldownVendor(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	locationID := r.URL.Query().Get("location_id")
	if locationID == "" {
		WriteError(w, http.StatusBadRequest, "FINANCIAL_MISSING_LOCATION", "location_id is required")
		return
	}
	ingredientID := r.URL.Query().Get("ingredient_id")
	if ingredientID == "" {
		WriteError(w, http.StatusBadRequest, "FINANCIAL_MISSING_INGREDIENT", "ingredient_id is required")
		return
	}

	history, err := h.svc.GetIngredientVendorHistory(r.Context(), orgID, locationID, ingredientID)
	if err != nil {
		slog.Error("financial vendor history error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "FINANCIAL_VENDOR_HISTORY_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"vendor_history": history})
}

// GetPeriodComparison handles GET /api/v1/financial/period-comparison
func (h *FinancialHandler) GetPeriodComparison(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	locationID := r.URL.Query().Get("location_id")
	if locationID == "" {
		WriteError(w, http.StatusBadRequest, "FINANCIAL_MISSING_LOCATION", "location_id is required")
		return
	}

	from, to := parseDateRange(r)
	comparison, err := h.svc.CalculatePeriodComparison(r.Context(), orgID, locationID, from, to)
	if err != nil {
		slog.Error("financial comparison error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "FINANCIAL_COMPARISON_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, comparison)
}
