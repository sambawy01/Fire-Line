package api

import (
	"log/slog"
	"encoding/json"
	"net/http"

	"github.com/opsnerve/fireline/internal/onboarding"
	"github.com/opsnerve/fireline/internal/tenant"
)

// OnboardingHandler handles the onboarding wizard API.
type OnboardingHandler struct {
	svc *onboarding.Service
}

// NewOnboardingHandler creates a new OnboardingHandler.
func NewOnboardingHandler(svc *onboarding.Service) *OnboardingHandler {
	return &OnboardingHandler{svc: svc}
}

// RegisterRoutes registers all onboarding API routes.
func (h *OnboardingHandler) RegisterRoutes(mux *http.ServeMux, authMW func(http.Handler) http.Handler) {
	mux.Handle("POST /api/v1/onboarding/start", authMW(http.HandlerFunc(h.StartOnboarding)))
	mux.Handle("GET /api/v1/onboarding/session", authMW(http.HandlerFunc(h.GetSession)))
	mux.Handle("PUT /api/v1/onboarding/step", authMW(http.HandlerFunc(h.UpdateStep)))
	mux.Handle("GET /api/v1/onboarding/insights", authMW(http.HandlerFunc(h.GetInsights)))
	mux.Handle("GET /api/v1/onboarding/concept", authMW(http.HandlerFunc(h.InferConcept)))
	mux.Handle("GET /api/v1/onboarding/modules", authMW(http.HandlerFunc(h.RecommendModules)))
	mux.Handle("GET /api/v1/onboarding/checklist", authMW(http.HandlerFunc(h.GetChecklist)))
	mux.Handle("POST /api/v1/onboarding/checklist/{id}/complete", authMW(http.HandlerFunc(h.CompleteChecklistItem)))
}

// StartOnboarding POST /api/v1/onboarding/start
// Body: { "user_id": "..." }
func (h *OnboardingHandler) StartOnboarding(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	var req struct {
		UserID string `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.UserID == "" {
		WriteError(w, http.StatusBadRequest, "ONBOARDING_INVALID_REQUEST", "user_id is required")
		return
	}
	sess, err := h.svc.StartOnboarding(r.Context(), orgID, req.UserID)
	if err != nil {
		slog.Error("onboarding start error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "ONBOARDING_START_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusCreated, map[string]any{"session": sess})
}

// GetSession GET /api/v1/onboarding/session
func (h *OnboardingHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	sess, err := h.svc.GetSession(r.Context(), orgID)
	if err != nil {
		WriteError(w, http.StatusNotFound, "ONBOARDING_NOT_FOUND", "no onboarding session found")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"session": sess})
}

// UpdateStep PUT /api/v1/onboarding/step
// Body: { "session_id": "...", "step": "...", "data": {...} }
func (h *OnboardingHandler) UpdateStep(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	var req struct {
		SessionID string         `json:"session_id"`
		Step      string         `json:"step"`
		Data      map[string]any `json:"data"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "ONBOARDING_INVALID_REQUEST", "invalid JSON body")
		return
	}
	if req.SessionID == "" || req.Step == "" {
		WriteError(w, http.StatusBadRequest, "ONBOARDING_INVALID_REQUEST", "session_id and step are required")
		return
	}
	if req.Data == nil {
		req.Data = map[string]any{}
	}
	sess, err := h.svc.UpdateStep(r.Context(), orgID, req.SessionID, req.Step, req.Data)
	if err != nil {
		slog.Error("onboarding update error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "ONBOARDING_UPDATE_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"session": sess})
}

// GetInsights GET /api/v1/onboarding/insights?location_id=...
func (h *OnboardingHandler) GetInsights(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	locationID := r.URL.Query().Get("location_id")
	if locationID == "" {
		WriteError(w, http.StatusBadRequest, "ONBOARDING_MISSING_LOCATION", "location_id is required")
		return
	}
	insights, err := h.svc.GenerateFirstInsights(r.Context(), orgID, locationID)
	if err != nil {
		slog.Error("onboarding insights error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "ONBOARDING_INSIGHTS_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"insights": insights})
}

// InferConcept GET /api/v1/onboarding/concept?location_id=...
func (h *OnboardingHandler) InferConcept(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	locationID := r.URL.Query().Get("location_id")
	if locationID == "" {
		WriteError(w, http.StatusBadRequest, "ONBOARDING_MISSING_LOCATION", "location_id is required")
		return
	}
	conceptType, err := h.svc.InferConceptType(r.Context(), orgID, locationID)
	if err != nil {
		slog.Error("onboarding concept error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "ONBOARDING_CONCEPT_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"concept_type": conceptType})
}

// RecommendModules GET /api/v1/onboarding/modules?priorities=reduce_waste,boost_revenue
func (h *OnboardingHandler) RecommendModules(w http.ResponseWriter, r *http.Request) {
	_, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	// Accept priorities as comma-separated query param or JSON body
	prioritiesParam := r.URL.Query().Get("priorities")
	var priorities []string
	if prioritiesParam != "" {
		// Split comma-separated string
		for _, p := range splitComma(prioritiesParam) {
			if p != "" {
				priorities = append(priorities, p)
			}
		}
	}
	modules := h.svc.RecommendModules(priorities)
	WriteJSON(w, http.StatusOK, map[string]any{"modules": modules})
}

// GetChecklist GET /api/v1/onboarding/checklist
func (h *OnboardingHandler) GetChecklist(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	items, err := h.svc.GetChecklist(r.Context(), orgID)
	if err != nil {
		slog.Error("onboarding checklist error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "ONBOARDING_CHECKLIST_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

// CompleteChecklistItem POST /api/v1/onboarding/checklist/{id}/complete
func (h *OnboardingHandler) CompleteChecklistItem(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	itemID := r.PathValue("id")
	if itemID == "" {
		WriteError(w, http.StatusBadRequest, "ONBOARDING_INVALID_REQUEST", "item id is required")
		return
	}
	if err := h.svc.CompleteChecklistItem(r.Context(), orgID, itemID); err != nil {
		if err.Error() == "item not found" {
			WriteError(w, http.StatusNotFound, "ONBOARDING_ITEM_NOT_FOUND", "checklist item not found")
			return
		}
		slog.Error("onboarding complete error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "ONBOARDING_COMPLETE_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "completed"})
}

// splitComma splits a comma-separated string, trimming spaces.
func splitComma(s string) []string {
	var parts []string
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == ',' {
			part := s[start:i]
			// trim spaces
			for len(part) > 0 && part[0] == ' ' {
				part = part[1:]
			}
			for len(part) > 0 && part[len(part)-1] == ' ' {
				part = part[:len(part)-1]
			}
			parts = append(parts, part)
			start = i + 1
		}
	}
	return parts
}
