package api

import (
	"net/http"
	"time"

	"github.com/opsnerve/fireline/internal/menu"
	"github.com/opsnerve/fireline/internal/tenant"
)

type MenuHandler struct {
	svc *menu.Service
}

func NewMenuHandler(svc *menu.Service) *MenuHandler {
	return &MenuHandler{svc: svc}
}

func (h *MenuHandler) RegisterRoutes(mux *http.ServeMux, authMW func(http.Handler) http.Handler) {
	mux.Handle("GET /api/v1/menu/items", authMW(http.HandlerFunc(h.GetItems)))
	mux.Handle("GET /api/v1/menu/summary", authMW(http.HandlerFunc(h.GetSummary)))
}

func (h *MenuHandler) GetItems(w http.ResponseWriter, r *http.Request) {
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
	from, to := parseMenuDateRange(r)
	items, err := h.svc.AnalyzeMenuItems(r.Context(), orgID, locationID, from, to)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "MENU_ANALYSIS_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *MenuHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
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
	from, to := parseMenuDateRange(r)
	summary, err := h.svc.GetSummary(r.Context(), orgID, locationID, from, to)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "MENU_SUMMARY_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, summary)
}

// parseMenuDateRange defaults to last 30 days (menu needs historical data).
func parseMenuDateRange(r *http.Request) (time.Time, time.Time) {
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	from, err := time.Parse(time.RFC3339, fromStr)
	if err != nil {
		from = time.Now().AddDate(0, 0, -30).Truncate(24 * time.Hour)
	}
	to, err2 := time.Parse(time.RFC3339, toStr)
	if err2 != nil {
		to = time.Now()
	}
	return from, to
}
