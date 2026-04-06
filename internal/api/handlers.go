package api

import (
	"log/slog"
	"encoding/json"
	"net/http"
	"time"

	"github.com/opsnerve/fireline/internal/alerting"
	"github.com/opsnerve/fireline/internal/financial"
	"github.com/opsnerve/fireline/internal/inventory"
	"github.com/opsnerve/fireline/internal/tenant"
)

// Handlers holds all module API handlers.
type Handlers struct {
	Inventory *InventoryHandler
	Financial *FinancialHandler
	Alerting  *AlertingHandler
}

// InventoryHandler handles inventory intelligence API requests.
type InventoryHandler struct {
	svc *inventory.Service
}

func NewInventoryHandler(svc *inventory.Service) *InventoryHandler {
	return &InventoryHandler{svc: svc}
}

func (h *InventoryHandler) RegisterRoutes(mux *http.ServeMux, authMW func(http.Handler) http.Handler) {
	mux.Handle("GET /api/v1/inventory/usage", authMW(http.HandlerFunc(h.GetTheoreticalUsage)))
	mux.Handle("GET /api/v1/inventory/par", authMW(http.HandlerFunc(h.GetPARStatus)))
	mux.Handle("POST /api/v1/inventory/explode", authMW(http.HandlerFunc(h.MaterializeRecipeExplosion)))

	// Counting
	mux.Handle("POST /api/v1/inventory/counts", authMW(http.HandlerFunc(h.CreateCount)))
	mux.Handle("GET /api/v1/inventory/counts/{id}", authMW(http.HandlerFunc(h.GetCount)))
	mux.Handle("PUT /api/v1/inventory/counts/{id}", authMW(http.HandlerFunc(h.UpdateCountStatus)))
	mux.Handle("POST /api/v1/inventory/counts/{id}/lines", authMW(http.HandlerFunc(h.UpsertCountLines)))

	// Waste
	mux.Handle("POST /api/v1/inventory/waste", authMW(http.HandlerFunc(h.LogWaste)))
	mux.Handle("GET /api/v1/inventory/waste", authMW(http.HandlerFunc(h.ListWasteLogs)))
	mux.Handle("DELETE /api/v1/inventory/waste/{id}", authMW(http.HandlerFunc(h.DeleteWaste)))

	// Variances
	mux.Handle("GET /api/v1/inventory/variances", authMW(http.HandlerFunc(h.ListVariances)))
	mux.Handle("GET /api/v1/inventory/expiry", authMW(http.HandlerFunc(h.GetExpiryAlerts)))

	// Purchase Orders — specific paths before parameterized ones
	mux.Handle("GET /api/v1/inventory/po/pending", authMW(http.HandlerFunc(h.ListPendingPOs)))
	mux.Handle("GET /api/v1/inventory/par-breaches", authMW(http.HandlerFunc(h.GetPARBreaches)))
	mux.Handle("POST /api/v1/inventory/po", authMW(http.HandlerFunc(h.CreatePO)))
	mux.Handle("GET /api/v1/inventory/po", authMW(http.HandlerFunc(h.ListPOs)))
	mux.Handle("GET /api/v1/inventory/po/{id}", authMW(http.HandlerFunc(h.GetPO)))
	mux.Handle("PUT /api/v1/inventory/po/{id}", authMW(http.HandlerFunc(h.UpdatePO)))
	mux.Handle("DELETE /api/v1/inventory/po/{id}", authMW(http.HandlerFunc(h.DeletePO)))
	mux.Handle("POST /api/v1/inventory/po/{id}/receive", authMW(http.HandlerFunc(h.ReceivePO)))
}

func (h *InventoryHandler) GetTheoreticalUsage(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	locationID := r.URL.Query().Get("location_id")
	if locationID == "" {
		WriteError(w, http.StatusBadRequest, "INVENTORY_MISSING_LOCATION", "location_id is required")
		return
	}

	from, to := parseDateRange(r)
	usage, err := h.svc.CalculateTheoreticalUsage(r.Context(), orgID, locationID, from, to)
	if err != nil {
		slog.Error("inventory usage error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "INVENTORY_USAGE_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"usage": usage, "period_start": from, "period_end": to})
}

func (h *InventoryHandler) GetPARStatus(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	locationID := r.URL.Query().Get("location_id")
	if locationID == "" {
		WriteError(w, http.StatusBadRequest, "INVENTORY_MISSING_LOCATION", "location_id is required")
		return
	}

	// In production, current levels would come from inventory counts
	status, err := h.svc.GetPARStatus(r.Context(), orgID, locationID, map[string]float64{})
	if err != nil {
		slog.Error("inventory par error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "INVENTORY_PAR_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"par_status": status})
}

func (h *InventoryHandler) MaterializeRecipeExplosion(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	var req struct {
		MenuItemID string `json:"menu_item_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}
	if err := h.svc.MaterializeRecipeExplosion(r.Context(), orgID, req.MenuItemID); err != nil {
		slog.Error("inventory explode error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "INVENTORY_EXPLODE_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// FinancialHandler handles financial intelligence API requests.
type FinancialHandler struct {
	svc *financial.Service
}

func NewFinancialHandler(svc *financial.Service) *FinancialHandler {
	return &FinancialHandler{svc: svc}
}

func (h *FinancialHandler) RegisterRoutes(mux *http.ServeMux, authMW func(http.Handler) http.Handler) {
	mux.Handle("GET /api/v1/financial/pnl", authMW(http.HandlerFunc(h.GetPnL)))
	mux.Handle("GET /api/v1/financial/anomalies", authMW(http.HandlerFunc(h.GetAnomalies)))

	// Budgets
	mux.Handle("POST /api/v1/financial/budgets", authMW(http.HandlerFunc(h.CreateBudget)))
	mux.Handle("GET /api/v1/financial/budgets", authMW(http.HandlerFunc(h.ListBudgets)))

	// Variance & period comparison
	mux.Handle("GET /api/v1/financial/budget-variance", authMW(http.HandlerFunc(h.GetBudgetVariance)))
	mux.Handle("GET /api/v1/financial/period-comparison", authMW(http.HandlerFunc(h.GetPeriodComparison)))

	// Cost centers
	mux.Handle("GET /api/v1/financial/cost-centers", authMW(http.HandlerFunc(h.GetCostCenters)))

	// Transaction anomalies
	mux.Handle("GET /api/v1/financial/transaction-anomalies", authMW(http.HandlerFunc(h.GetTransactionAnomalies)))

	// Drilldown — specific paths before parameterized ones
	mux.Handle("GET /api/v1/financial/drilldown/items", authMW(http.HandlerFunc(h.GetDrilldownItems)))
	mux.Handle("GET /api/v1/financial/drilldown/ingredients", authMW(http.HandlerFunc(h.GetDrilldownIngredients)))
	mux.Handle("GET /api/v1/financial/drilldown/vendor", authMW(http.HandlerFunc(h.GetDrilldownVendor)))
}

func (h *FinancialHandler) GetPnL(w http.ResponseWriter, r *http.Request) {
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
	pnl, err := h.svc.CalculatePnL(r.Context(), orgID, locationID, from, to)
	if err != nil {
		slog.Error("financial pnl error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "FINANCIAL_PNL_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, pnl)
}

func (h *FinancialHandler) GetAnomalies(w http.ResponseWriter, r *http.Request) {
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

	anomalies, err := h.svc.DetectAnomalies(r.Context(), orgID, locationID, time.Now())
	if err != nil {
		slog.Error("financial anomaly error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "FINANCIAL_ANOMALY_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"anomalies": anomalies})
}

// AlertingHandler handles alerting API requests.
type AlertingHandler struct {
	svc *alerting.Service
}

func NewAlertingHandler(svc *alerting.Service) *AlertingHandler {
	return &AlertingHandler{svc: svc}
}

func (h *AlertingHandler) RegisterRoutes(mux *http.ServeMux, authMW func(http.Handler) http.Handler) {
	mux.Handle("GET /api/v1/alerts", authMW(http.HandlerFunc(h.GetQueue)))
	mux.Handle("GET /api/v1/alerts/count", authMW(http.HandlerFunc(h.GetCount)))
	mux.Handle("POST /api/v1/alerts/{id}/acknowledge", authMW(http.HandlerFunc(h.Acknowledge)))
	mux.Handle("POST /api/v1/alerts/{id}/resolve", authMW(http.HandlerFunc(h.Resolve)))
}

func (h *AlertingHandler) GetQueue(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	locationID := r.URL.Query().Get("location_id")
	queue := h.svc.GetQueue(orgID, locationID)
	WriteJSON(w, http.StatusOK, map[string]any{"alerts": queue})
}

func (h *AlertingHandler) GetCount(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	count := h.svc.ActiveCount(orgID)
	WriteJSON(w, http.StatusOK, map[string]int{"count": count})
}

func (h *AlertingHandler) Acknowledge(w http.ResponseWriter, r *http.Request) {
	alertID := r.PathValue("id")
	if !h.svc.Acknowledge(alertID) {
		WriteError(w, http.StatusNotFound, "ALERT_NOT_FOUND", "alert not found or not active")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "acknowledged"})
}

func (h *AlertingHandler) Resolve(w http.ResponseWriter, r *http.Request) {
	alertID := r.PathValue("id")
	if !h.svc.Resolve(alertID) {
		WriteError(w, http.StatusNotFound, "ALERT_NOT_FOUND", "alert not found or not active")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "resolved"})
}

func parseDateRange(r *http.Request) (time.Time, time.Time) {
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	from, err := time.Parse(time.RFC3339, fromStr)
	if err != nil {
		from = time.Now().Truncate(24 * time.Hour) // start of today
	}
	to, err := time.Parse(time.RFC3339, toStr)
	if err != nil {
		to = time.Now()
	}
	return from, to
}
