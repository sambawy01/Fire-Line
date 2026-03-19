package api

import (
	"net/http"
	"time"

	"github.com/opsnerve/fireline/internal/tenant"
	"github.com/opsnerve/fireline/internal/vendor"
)

// VendorHandler handles HTTP requests for vendor intelligence endpoints.
type VendorHandler struct {
	svc *vendor.Service
}

// NewVendorHandler creates a new VendorHandler.
func NewVendorHandler(svc *vendor.Service) *VendorHandler {
	return &VendorHandler{svc: svc}
}

// RegisterRoutes registers the vendor API routes on the provided mux.
func (h *VendorHandler) RegisterRoutes(mux *http.ServeMux, authMW func(http.Handler) http.Handler) {
	mux.Handle("GET /api/v1/vendors", authMW(http.HandlerFunc(h.GetVendors)))
	mux.Handle("GET /api/v1/vendors/summary", authMW(http.HandlerFunc(h.GetSummary)))
}

// GetVendors returns the list of vendors with spend and scoring analytics.
func (h *VendorHandler) GetVendors(w http.ResponseWriter, r *http.Request) {
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
	from, to := parseVendorDateRange(r)
	vendors, err := h.svc.GetVendors(r.Context(), orgID, locationID, from, to)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "VENDOR_LIST_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"vendors": vendors})
}

// GetSummary returns location-wide vendor KPI rollups.
func (h *VendorHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
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
	from, to := parseVendorDateRange(r)
	summary, err := h.svc.GetSummary(r.Context(), orgID, locationID, from, to)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "VENDOR_SUMMARY_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, summary)
}

// parseVendorDateRange defaults to last 30 days when query params are absent or malformed.
func parseVendorDateRange(r *http.Request) (time.Time, time.Time) {
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
