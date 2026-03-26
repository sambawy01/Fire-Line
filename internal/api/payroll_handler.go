package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/opsnerve/fireline/internal/payroll"
	"github.com/opsnerve/fireline/internal/tenant"
)

// PayrollHandler handles payroll reporting API requests.
type PayrollHandler struct {
	svc *payroll.Service
}

// NewPayrollHandler creates a new PayrollHandler.
func NewPayrollHandler(svc *payroll.Service) *PayrollHandler {
	return &PayrollHandler{svc: svc}
}

// RegisterRoutes registers payroll API routes on the given mux.
func (h *PayrollHandler) RegisterRoutes(mux *http.ServeMux, authMW func(http.Handler) http.Handler) {
	allowed := requireRole("gm", "ops_director", "owner")

	mux.Handle("GET /api/v1/payroll/summary", chain(http.HandlerFunc(h.GetPayrollSummary), authMW, allowed))
	mux.Handle("GET /api/v1/payroll/history", chain(http.HandlerFunc(h.GetPayrollHistory), authMW, allowed))
	mux.Handle("GET /api/v1/payroll/export", chain(http.HandlerFunc(h.ExportPayroll), authMW, allowed))
}

// GetPayrollSummary returns per-employee payroll data for a date range.
// GET /api/v1/payroll/summary?location_id=...&period_start=...&period_end=...
func (h *PayrollHandler) GetPayrollSummary(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	locationID := r.URL.Query().Get("location_id")
	if locationID == "" {
		WriteError(w, http.StatusBadRequest, "PAYROLL_MISSING_LOCATION", "location_id is required")
		return
	}

	periodStart := r.URL.Query().Get("period_start")
	periodEnd := r.URL.Query().Get("period_end")
	if periodStart == "" || periodEnd == "" {
		WriteError(w, http.StatusBadRequest, "PAYROLL_MISSING_PERIOD", "period_start and period_end are required")
		return
	}

	summary, err := h.svc.GetPayrollSummary(r.Context(), orgID, locationID, periodStart, periodEnd)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "PAYROLL_SUMMARY_ERROR", err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, summary)
}

// GetPayrollHistory returns monthly payroll aggregates.
// GET /api/v1/payroll/history?location_id=...&months=6
func (h *PayrollHandler) GetPayrollHistory(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	locationID := r.URL.Query().Get("location_id")
	if locationID == "" {
		WriteError(w, http.StatusBadRequest, "PAYROLL_MISSING_LOCATION", "location_id is required")
		return
	}

	months := 6
	if m := r.URL.Query().Get("months"); m != "" {
		if parsed, err := strconv.Atoi(m); err == nil && parsed > 0 {
			months = parsed
		}
	}

	history, err := h.svc.GetPayrollHistory(r.Context(), orgID, locationID, months)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "PAYROLL_HISTORY_ERROR", err.Error())
		return
	}

	WriteList(w, http.StatusOK, "periods", history)
}

// ExportPayroll streams a CSV payroll report.
// GET /api/v1/payroll/export?location_id=...&period_start=...&period_end=...
func (h *PayrollHandler) ExportPayroll(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	locationID := r.URL.Query().Get("location_id")
	if locationID == "" {
		WriteError(w, http.StatusBadRequest, "PAYROLL_MISSING_LOCATION", "location_id is required")
		return
	}

	periodStart := r.URL.Query().Get("period_start")
	periodEnd := r.URL.Query().Get("period_end")
	if periodStart == "" || periodEnd == "" {
		WriteError(w, http.StatusBadRequest, "PAYROLL_MISSING_PERIOD", "period_start and period_end are required")
		return
	}

	csvBytes, err := h.svc.ExportPayroll(r.Context(), orgID, locationID, periodStart, periodEnd)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "PAYROLL_EXPORT_ERROR", err.Error())
		return
	}

	filename := fmt.Sprintf("payroll-%s-to-%s.csv", periodStart, periodEnd)
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(csvBytes)))
	w.WriteHeader(http.StatusOK)
	w.Write(csvBytes) //nolint:errcheck
}
