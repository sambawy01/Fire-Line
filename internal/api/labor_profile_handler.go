package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/opsnerve/fireline/internal/tenant"
)

// ListProfiles returns ELU profiles for all employees at a location.
// GET /api/v1/labor/profiles?location_id=...
func (h *LaborHandler) ListProfiles(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	locationID := r.URL.Query().Get("location_id")
	if locationID == "" {
		WriteError(w, http.StatusBadRequest, "LABOR_MISSING_LOCATION", "location_id is required")
		return
	}
	profiles, err := h.svc.ListEmployeeProfiles(r.Context(), orgID, locationID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "LABOR_PROFILES_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"profiles": profiles})
}

// GetProfile returns the ELU profile for a single employee.
// GET /api/v1/labor/profiles/{id}
func (h *LaborHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	employeeID := r.PathValue("id")
	if employeeID == "" {
		WriteError(w, http.StatusBadRequest, "LABOR_MISSING_EMPLOYEE", "employee id is required")
		return
	}
	profile, err := h.svc.GetEmployeeProfile(r.Context(), orgID, employeeID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "LABOR_PROFILE_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, profile)
}

// UpdateELU overwrites the ELU ratings for an employee.
// PUT /api/v1/labor/profiles/{id}/elu
func (h *LaborHandler) UpdateELU(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	employeeID := r.PathValue("id")
	if employeeID == "" {
		WriteError(w, http.StatusBadRequest, "LABOR_MISSING_EMPLOYEE", "employee id is required")
		return
	}
	var body struct {
		Ratings map[string]float64 `json:"ratings"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}
	if err := h.svc.UpdateELURatings(r.Context(), orgID, employeeID, body.Ratings); err != nil {
		WriteError(w, http.StatusInternalServerError, "LABOR_ELU_UPDATE_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// UpdateAvailability overwrites the availability schedule for an employee.
// PUT /api/v1/labor/profiles/{id}/availability
func (h *LaborHandler) UpdateAvailability(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	employeeID := r.PathValue("id")
	if employeeID == "" {
		WriteError(w, http.StatusBadRequest, "LABOR_MISSING_EMPLOYEE", "employee id is required")
		return
	}
	var availability map[string]any
	if err := json.NewDecoder(r.Body).Decode(&availability); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}
	if err := h.svc.UpdateAvailability(r.Context(), orgID, employeeID, availability); err != nil {
		WriteError(w, http.StatusInternalServerError, "LABOR_AVAILABILITY_UPDATE_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// UpdateCertifications overwrites the certifications list for an employee.
// PUT /api/v1/labor/profiles/{id}/certifications
func (h *LaborHandler) UpdateCertifications(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	employeeID := r.PathValue("id")
	if employeeID == "" {
		WriteError(w, http.StatusBadRequest, "LABOR_MISSING_EMPLOYEE", "employee id is required")
		return
	}
	var body struct {
		Certifications []string `json:"certifications"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}
	if err := h.svc.UpdateCertifications(r.Context(), orgID, employeeID, body.Certifications); err != nil {
		WriteError(w, http.StatusInternalServerError, "LABOR_CERTS_UPDATE_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// AwardPoints awards or deducts staff points for an employee.
// POST /api/v1/labor/points
func (h *LaborHandler) AwardPoints(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	var body struct {
		EmployeeID  string  `json:"employee_id"`
		Points      float64 `json:"points"`
		Reason      string  `json:"reason"`
		Description string  `json:"description"`
		ShiftID     *string `json:"shift_id,omitempty"`
		AwardedBy   *string `json:"awarded_by,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}
	if body.EmployeeID == "" || body.Reason == "" {
		WriteError(w, http.StatusBadRequest, "LABOR_MISSING_FIELDS", "employee_id and reason are required")
		return
	}
	pe, err := h.svc.AwardPoints(r.Context(), orgID, body.EmployeeID, body.Points, body.Reason, body.Description, body.ShiftID, body.AwardedBy)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "LABOR_AWARD_POINTS_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, pe)
}

// GetPointHistory returns the recent point event history for an employee.
// GET /api/v1/labor/points/{employee_id}?limit=50
func (h *LaborHandler) GetPointHistory(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	employeeID := r.PathValue("employee_id")
	if employeeID == "" {
		WriteError(w, http.StatusBadRequest, "LABOR_MISSING_EMPLOYEE", "employee_id is required")
		return
	}
	limit := 50
	if ls := r.URL.Query().Get("limit"); ls != "" {
		if n, err := strconv.Atoi(ls); err == nil && n > 0 {
			limit = n
		}
	}
	events, err := h.svc.GetPointHistory(r.Context(), orgID, employeeID, limit)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "LABOR_POINT_HISTORY_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"events": events})
}

// GetLeaderboard returns employees ranked by staff points.
// GET /api/v1/labor/leaderboard?location_id=...&limit=25
func (h *LaborHandler) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	locationID := r.URL.Query().Get("location_id")
	limit := 25
	if ls := r.URL.Query().Get("limit"); ls != "" {
		if n, err := strconv.Atoi(ls); err == nil && n > 0 {
			limit = n
		}
	}
	entries, err := h.svc.GetLeaderboard(r.Context(), orgID, locationID, limit)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "LABOR_LEADERBOARD_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"leaderboard": entries})
}
