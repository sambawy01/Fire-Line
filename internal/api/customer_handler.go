package api

import (
	"net/http"

	"github.com/opsnerve/fireline/internal/customer"
	"github.com/opsnerve/fireline/internal/tenant"
)

// CustomerHandler exposes the customer intelligence endpoints.
type CustomerHandler struct {
	svc *customer.Service
}

// NewCustomerHandler creates a new CustomerHandler.
func NewCustomerHandler(svc *customer.Service) *CustomerHandler {
	return &CustomerHandler{svc: svc}
}

// RegisterRoutes mounts customer routes onto the provided mux, all behind the
// auth middleware. Specific paths are registered before parameterized ones to
// avoid pattern conflicts.
func (h *CustomerHandler) RegisterRoutes(mux *http.ServeMux, authMW func(http.Handler) http.Handler) {
	// Existing routes.
	mux.Handle("GET /api/v1/customers", authMW(http.HandlerFunc(h.GetCustomers)))
	mux.Handle("GET /api/v1/customers/summary", authMW(http.HandlerFunc(h.GetSummary)))
	mux.Handle("POST /api/v1/customers/analyze", authMW(http.HandlerFunc(h.Analyze)))

	// Guest intelligence — analytics (specific paths before parameterized).
	mux.Handle("POST /api/v1/customers/analytics/refresh", authMW(http.HandlerFunc(h.RefreshAnalytics)))
	mux.Handle("GET /api/v1/customers/analytics/segments", authMW(http.HandlerFunc(h.GetSegmentDist)))
	mux.Handle("GET /api/v1/customers/analytics/churn", authMW(http.HandlerFunc(h.GetChurnDist)))
	mux.Handle("GET /api/v1/customers/analytics/clv", authMW(http.HandlerFunc(h.GetCLVDist)))

	// Guest intelligence — profiles.
	mux.Handle("POST /api/v1/customers/resolve", authMW(http.HandlerFunc(h.ResolveGuest)))
	mux.Handle("GET /api/v1/customers/guests", authMW(http.HandlerFunc(h.ListGuests)))
	mux.Handle("GET /api/v1/customers/guests/{id}", authMW(http.HandlerFunc(h.GetGuest)))
	mux.Handle("PUT /api/v1/customers/guests/{id}/enrich", authMW(http.HandlerFunc(h.EnrichGuest)))
}

// GetCustomers returns all customers for a location ordered by total spend.
func (h *CustomerHandler) GetCustomers(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	locationID := r.URL.Query().Get("location_id")
	if locationID == "" {
		WriteError(w, http.StatusBadRequest, "CUSTOMER_MISSING_LOCATION", "location_id is required")
		return
	}
	customers, err := h.svc.GetCustomers(r.Context(), orgID, locationID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "CUSTOMER_LIST_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"customers": customers})
}

// GetSummary returns location-wide customer KPI rollups.
func (h *CustomerHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	locationID := r.URL.Query().Get("location_id")
	if locationID == "" {
		WriteError(w, http.StatusBadRequest, "CUSTOMER_MISSING_LOCATION", "location_id is required")
		return
	}
	summary, err := h.svc.GetSummary(r.Context(), orgID, locationID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "CUSTOMER_SUMMARY_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, summary)
}

// Analyze triggers a batch AI segmentation and summary run for the location.
func (h *CustomerHandler) Analyze(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	locationID := r.URL.Query().Get("location_id")
	if locationID == "" {
		WriteError(w, http.StatusBadRequest, "CUSTOMER_MISSING_LOCATION", "location_id is required")
		return
	}
	result, err := h.svc.AnalyzeAll(r.Context(), orgID, locationID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "CUSTOMER_ANALYZE_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, result)
}
