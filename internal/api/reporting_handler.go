package api

import (
	"log/slog"
	"fmt"
	"net/http"

	"github.com/opsnerve/fireline/internal/reporting"
	"github.com/opsnerve/fireline/internal/tenant"
)

// ReportingHandler serves daily report endpoints.
type ReportingHandler struct {
	svc *reporting.Service
}

// NewReportingHandler creates a new ReportingHandler.
func NewReportingHandler(svc *reporting.Service) *ReportingHandler {
	return &ReportingHandler{svc: svc}
}

// RegisterRoutes mounts report endpoints behind the provided auth middleware.
func (h *ReportingHandler) RegisterRoutes(mux *http.ServeMux, authMW func(http.Handler) http.Handler) {
	mux.Handle("GET /api/v1/reports/daily", authMW(http.HandlerFunc(h.GetDaily)))
	mux.Handle("GET /api/v1/reports/daily/pdf", authMW(http.HandlerFunc(h.GetDailyPDF)))
}

// GetDaily returns a JSON daily report for the requested location and date range.
func (h *ReportingHandler) GetDaily(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	locationID := r.URL.Query().Get("location_id")
	if locationID == "" {
		WriteError(w, http.StatusBadRequest, "REPORT_MISSING_LOCATION", "location_id is required")
		return
	}

	from, to := parseDateRange(r)

	report, err := h.svc.GenerateDaily(r.Context(), orgID, locationID, from, to)
	if err != nil {
		slog.Error("report daily error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "REPORT_DAILY_ERROR", "an internal error occurred")
		return
	}

	WriteJSON(w, http.StatusOK, report)
}

// GetDailyPDF generates and streams a PDF daily report for the requested location.
// The frontend should use fetch + blob to download, as the endpoint requires a
// standard Authorization header (Bearer token) like all other authenticated routes.
func (h *ReportingHandler) GetDailyPDF(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	locationID := r.URL.Query().Get("location_id")
	if locationID == "" {
		WriteError(w, http.StatusBadRequest, "REPORT_MISSING_LOCATION", "location_id is required")
		return
	}

	from, to := parseDateRange(r)

	report, err := h.svc.GenerateDaily(r.Context(), orgID, locationID, from, to)
	if err != nil {
		slog.Error("report daily error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "REPORT_DAILY_ERROR", "an internal error occurred")
		return
	}

	pdfBytes, err := h.svc.GeneratePDF(report)
	if err != nil {
		slog.Error("report pdf error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "REPORT_PDF_ERROR", "an internal error occurred")
		return
	}

	filename := fmt.Sprintf("fireline-daily-%s.pdf", report.ReportDate)
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(pdfBytes)))
	w.WriteHeader(http.StatusOK)
	w.Write(pdfBytes) //nolint:errcheck
}
