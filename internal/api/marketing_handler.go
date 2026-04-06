package api

import (
	"log/slog"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/opsnerve/fireline/internal/marketing"
	"github.com/opsnerve/fireline/internal/tenant"
)

// MarketingHandler handles HTTP requests for marketing campaign and loyalty endpoints.
type MarketingHandler struct {
	svc *marketing.Service
}

// NewMarketingHandler creates a new MarketingHandler.
func NewMarketingHandler(svc *marketing.Service) *MarketingHandler {
	return &MarketingHandler{svc: svc}
}

// RegisterRoutes registers all marketing API routes. Specific paths are registered before
// parameterized ones to avoid ambiguity with Go's net/http mux.
func (h *MarketingHandler) RegisterRoutes(mux *http.ServeMux, authMW func(http.Handler) http.Handler) {
	// Campaigns — specific before parameterized
	mux.Handle("POST /api/v1/marketing/campaigns/simulate", authMW(http.HandlerFunc(h.SimulateCampaign)))
	mux.Handle("POST /api/v1/marketing/campaigns", authMW(http.HandlerFunc(h.CreateCampaign)))
	mux.Handle("GET /api/v1/marketing/campaigns", authMW(http.HandlerFunc(h.ListCampaigns)))
	mux.Handle("GET /api/v1/marketing/campaigns/{id}", authMW(http.HandlerFunc(h.GetCampaign)))
	mux.Handle("PUT /api/v1/marketing/campaigns/{id}", authMW(http.HandlerFunc(h.UpdateCampaign)))
	mux.Handle("POST /api/v1/marketing/campaigns/{id}/activate", authMW(http.HandlerFunc(h.ActivateCampaign)))
	mux.Handle("POST /api/v1/marketing/campaigns/{id}/pause", authMW(http.HandlerFunc(h.PauseCampaign)))

	// Loyalty
	mux.Handle("POST /api/v1/marketing/loyalty/enroll", authMW(http.HandlerFunc(h.EnrollLoyalty)))
	mux.Handle("POST /api/v1/marketing/loyalty/earn", authMW(http.HandlerFunc(h.EarnPoints)))
	mux.Handle("POST /api/v1/marketing/loyalty/redeem", authMW(http.HandlerFunc(h.RedeemPoints)))
	mux.Handle("GET /api/v1/marketing/loyalty/members", authMW(http.HandlerFunc(h.ListLoyaltyMembers)))
	mux.Handle("GET /api/v1/marketing/loyalty/member/{guest_id}", authMW(http.HandlerFunc(h.GetLoyaltyMember)))

	// Analytics
	mux.Handle("GET /api/v1/marketing/analytics/campaigns", authMW(http.HandlerFunc(h.GetCampaignMetrics)))
	mux.Handle("GET /api/v1/marketing/analytics/loyalty", authMW(http.HandlerFunc(h.GetLoyaltyMetrics)))
}

// CreateCampaign handles POST /api/v1/marketing/campaigns
func (h *MarketingHandler) CreateCampaign(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	var input marketing.CampaignInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}
	if input.Name == "" || input.CampaignType == "" {
		WriteError(w, http.StatusBadRequest, "MARKETING_MISSING_FIELDS", "name and campaign_type are required")
		return
	}
	c, err := h.svc.CreateCampaign(r.Context(), orgID, input)
	if err != nil {
		slog.Error("marketing campaign create error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "MARKETING_CAMPAIGN_CREATE_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusCreated, c)
}

// ListCampaigns handles GET /api/v1/marketing/campaigns
func (h *MarketingHandler) ListCampaigns(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	locationID := r.URL.Query().Get("location_id")
	status := r.URL.Query().Get("status")
	campaigns, err := h.svc.ListCampaigns(r.Context(), orgID, locationID, status)
	if err != nil {
		slog.Error("marketing campaign list error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "MARKETING_CAMPAIGN_LIST_ERROR", "an internal error occurred")
		return
	}
	WriteList(w, http.StatusOK, "campaigns", campaigns)
}

// GetCampaign handles GET /api/v1/marketing/campaigns/{id}
func (h *MarketingHandler) GetCampaign(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	campaignID := r.PathValue("id")
	c, err := h.svc.GetCampaign(r.Context(), orgID, campaignID)
	if err != nil {
		slog.Error("marketing campaign not found", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusNotFound, "MARKETING_CAMPAIGN_NOT_FOUND", "resource not found")
		return
	}
	WriteJSON(w, http.StatusOK, c)
}

// UpdateCampaign handles PUT /api/v1/marketing/campaigns/{id}
func (h *MarketingHandler) UpdateCampaign(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	campaignID := r.PathValue("id")
	var input marketing.CampaignInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}
	c, err := h.svc.UpdateCampaign(r.Context(), orgID, campaignID, input)
	if err != nil {
		slog.Error("marketing campaign update error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "MARKETING_CAMPAIGN_UPDATE_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, c)
}

// ActivateCampaign handles POST /api/v1/marketing/campaigns/{id}/activate
func (h *MarketingHandler) ActivateCampaign(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	campaignID := r.PathValue("id")
	c, err := h.svc.ActivateCampaign(r.Context(), orgID, campaignID)
	if err != nil {
		slog.Error("marketing campaign activate error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "MARKETING_CAMPAIGN_ACTIVATE_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, c)
}

// PauseCampaign handles POST /api/v1/marketing/campaigns/{id}/pause
func (h *MarketingHandler) PauseCampaign(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	campaignID := r.PathValue("id")
	c, err := h.svc.PauseCampaign(r.Context(), orgID, campaignID)
	if err != nil {
		slog.Error("marketing campaign pause error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "MARKETING_CAMPAIGN_PAUSE_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, c)
}

// SimulateCampaign handles POST /api/v1/marketing/campaigns/simulate
func (h *MarketingHandler) SimulateCampaign(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	var input marketing.CampaignInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}
	result, err := h.svc.SimulateCampaign(r.Context(), orgID, input)
	if err != nil {
		slog.Error("marketing simulate error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "MARKETING_SIMULATE_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, result)
}

// EnrollLoyalty handles POST /api/v1/marketing/loyalty/enroll
func (h *MarketingHandler) EnrollLoyalty(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	var req struct {
		GuestID string `json:"guest_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}
	if req.GuestID == "" {
		WriteError(w, http.StatusBadRequest, "LOYALTY_MISSING_GUEST", "guest_id is required")
		return
	}
	member, err := h.svc.EnrollMember(r.Context(), orgID, req.GuestID)
	if err != nil {
		slog.Error("loyalty enroll error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "LOYALTY_ENROLL_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusCreated, member)
}

// EarnPoints handles POST /api/v1/marketing/loyalty/earn
func (h *MarketingHandler) EarnPoints(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	var req struct {
		GuestID     string  `json:"guest_id"`
		Points      float64 `json:"points"`
		Description string  `json:"description"`
		CheckID     *string `json:"check_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}
	if req.GuestID == "" || req.Points <= 0 {
		WriteError(w, http.StatusBadRequest, "LOYALTY_MISSING_FIELDS", "guest_id and positive points are required")
		return
	}
	member, err := h.svc.EarnPoints(r.Context(), orgID, req.GuestID, req.Points, req.Description, req.CheckID)
	if err != nil {
		slog.Error("loyalty earn error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "LOYALTY_EARN_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, member)
}

// RedeemPoints handles POST /api/v1/marketing/loyalty/redeem
func (h *MarketingHandler) RedeemPoints(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	var req struct {
		GuestID     string  `json:"guest_id"`
		Points      float64 `json:"points"`
		Description string  `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}
	if req.GuestID == "" || req.Points <= 0 {
		WriteError(w, http.StatusBadRequest, "LOYALTY_MISSING_FIELDS", "guest_id and positive points are required")
		return
	}
	member, err := h.svc.RedeemPoints(r.Context(), orgID, req.GuestID, req.Points, req.Description)
	if err != nil {
		slog.Error("loyalty redeem error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "LOYALTY_REDEEM_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, member)
}

// GetLoyaltyMember handles GET /api/v1/marketing/loyalty/member/{guest_id}
func (h *MarketingHandler) GetLoyaltyMember(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	guestID := r.PathValue("guest_id")
	member, err := h.svc.GetMember(r.Context(), orgID, guestID)
	if err != nil {
		slog.Error("loyalty member not found", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusNotFound, "LOYALTY_MEMBER_NOT_FOUND", "resource not found")
		return
	}
	WriteJSON(w, http.StatusOK, member)
}

// ListLoyaltyMembers handles GET /api/v1/marketing/loyalty/members
func (h *MarketingHandler) ListLoyaltyMembers(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	tier := r.URL.Query().Get("tier")
	limitStr := r.URL.Query().Get("limit")
	limit, _ := strconv.Atoi(limitStr)
	members, err := h.svc.ListMembers(r.Context(), orgID, tier, limit)
	if err != nil {
		slog.Error("loyalty list error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "LOYALTY_LIST_ERROR", "an internal error occurred")
		return
	}
	WriteList(w, http.StatusOK, "members", members)
}

// GetCampaignMetrics handles GET /api/v1/marketing/analytics/campaigns
func (h *MarketingHandler) GetCampaignMetrics(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	metrics, err := h.svc.GetCampaignMetrics(r.Context(), orgID)
	if err != nil {
		slog.Error("marketing metrics error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "MARKETING_METRICS_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, metrics)
}

// GetLoyaltyMetrics handles GET /api/v1/marketing/analytics/loyalty
func (h *MarketingHandler) GetLoyaltyMetrics(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	metrics, err := h.svc.GetLoyaltyMetrics(r.Context(), orgID)
	if err != nil {
		slog.Error("loyalty metrics error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "LOYALTY_METRICS_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, metrics)
}
