package api

import (
	"log/slog"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/opsnerve/fireline/internal/tenant"
)

// ListGuests handles GET /api/v1/customers/guests
// Query params: location_id (optional), sort_by, limit, offset
func (h *CustomerHandler) ListGuests(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	locationID := r.URL.Query().Get("location_id")
	sortBy := r.URL.Query().Get("sort_by")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	guests, err := h.svc.ListGuests(r.Context(), orgID, locationID, sortBy, limit, offset)
	if err != nil {
		slog.Error("guest list error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "GUEST_LIST_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"guests": guests})
}

// GetGuest handles GET /api/v1/customers/guests/{id}
func (h *CustomerHandler) GetGuest(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	guestID := r.PathValue("id")
	if guestID == "" {
		WriteError(w, http.StatusBadRequest, "GUEST_MISSING_ID", "guest id is required")
		return
	}

	profile, err := h.svc.GetGuestProfile(r.Context(), orgID, guestID)
	if err != nil {
		slog.Error("guest get error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "GUEST_GET_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, profile)
}

// EnrichGuest handles PUT /api/v1/customers/guests/{id}/enrich
// Body: { "first_name": "...", "email": "...", "phone": "..." }  (all optional)
func (h *CustomerHandler) EnrichGuest(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	guestID := r.PathValue("id")
	if guestID == "" {
		WriteError(w, http.StatusBadRequest, "GUEST_MISSING_ID", "guest id is required")
		return
	}

	var body struct {
		FirstName *string `json:"first_name"`
		Email     *string `json:"email"`
		Phone     *string `json:"phone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}

	profile, err := h.svc.EnrichGuest(r.Context(), orgID, guestID, body.FirstName, body.Email, body.Phone)
	if err != nil {
		slog.Error("guest enrich error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "GUEST_ENRICH_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, profile)
}

// ResolveGuest handles POST /api/v1/customers/resolve
// Body: { "check_id": "..." }
func (h *CustomerHandler) ResolveGuest(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	var body struct {
		CheckID string `json:"check_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}
	if body.CheckID == "" {
		WriteError(w, http.StatusBadRequest, "RESOLVE_MISSING_CHECK", "check_id is required")
		return
	}

	profile, err := h.svc.ResolveGuest(r.Context(), orgID, body.CheckID)
	if err != nil {
		slog.Error("resolve error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "RESOLVE_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, profile)
}

// RefreshAnalytics handles POST /api/v1/customers/analytics/refresh
// Runs segmentation, churn prediction, and CLV recalculation in sequence.
func (h *CustomerHandler) RefreshAnalytics(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	segUpdated, err := h.svc.RunSegmentation(r.Context(), orgID)
	if err != nil {
		slog.Error("analytics seg error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "ANALYTICS_SEG_ERROR", "an internal error occurred")
		return
	}

	churnUpdated, err := h.svc.RunChurnPrediction(r.Context(), orgID)
	if err != nil {
		slog.Error("analytics churn error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "ANALYTICS_CHURN_ERROR", "an internal error occurred")
		return
	}

	clvUpdated, err := h.svc.RecalculateAllCLV(r.Context(), orgID)
	if err != nil {
		slog.Error("analytics clv error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "ANALYTICS_CLV_ERROR", "an internal error occurred")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"segmentation_updated": segUpdated,
		"churn_updated":        churnUpdated,
		"clv_updated":          clvUpdated,
	})
}

// GetSegmentDist handles GET /api/v1/customers/analytics/segments
func (h *CustomerHandler) GetSegmentDist(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	buckets, err := h.svc.GetSegmentDistribution(r.Context(), orgID)
	if err != nil {
		slog.Error("analytics seg dist error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "ANALYTICS_SEG_DIST_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"segments": buckets})
}

// GetChurnDist handles GET /api/v1/customers/analytics/churn
func (h *CustomerHandler) GetChurnDist(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	buckets, err := h.svc.GetChurnDistribution(r.Context(), orgID)
	if err != nil {
		slog.Error("analytics churn dist error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "ANALYTICS_CHURN_DIST_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"churn": buckets})
}

// GetCLVDist handles GET /api/v1/customers/analytics/clv
func (h *CustomerHandler) GetCLVDist(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	buckets, err := h.svc.GetCLVDistribution(r.Context(), orgID)
	if err != nil {
		slog.Error("analytics clv dist error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "ANALYTICS_CLV_DIST_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"clv": buckets})
}
