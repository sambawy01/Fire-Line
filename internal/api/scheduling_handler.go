package api

import (
	"log/slog"
	"encoding/json"
	"net/http"
	"time"

	"github.com/opsnerve/fireline/internal/labor"
	"github.com/opsnerve/fireline/internal/tenant"
)

// RegisterRoutes registers all scheduling, forecast, and swap routes on the
// existing LaborHandler. Specific paths are registered before parameterized
// ones to avoid ambiguity.
func (h *LaborHandler) RegisterSchedulingRoutes(mux *http.ServeMux, authMW func(http.Handler) http.Handler) {
	// Forecast
	mux.Handle("POST /api/v1/labor/forecast", authMW(http.HandlerFunc(h.GenerateForecast)))
	mux.Handle("GET /api/v1/labor/forecast", authMW(http.HandlerFunc(h.GetForecast)))

	// Schedules — specific/static paths before parameterized ones
	mux.Handle("POST /api/v1/labor/schedules/generate", authMW(http.HandlerFunc(h.GenerateScheduleDraft)))
	mux.Handle("GET /api/v1/labor/schedules/employee/{id}", authMW(http.HandlerFunc(h.GetEmployeeSchedule)))
	mux.Handle("GET /api/v1/labor/schedules/cost", authMW(http.HandlerFunc(h.ProjectLaborCost)))
	mux.Handle("GET /api/v1/labor/schedules", authMW(http.HandlerFunc(h.GetSchedule)))
	mux.Handle("PUT /api/v1/labor/schedules/{id}", authMW(http.HandlerFunc(h.UpdateSchedule)))
	mux.Handle("POST /api/v1/labor/schedules/{id}/publish", authMW(http.HandlerFunc(h.PublishSchedule)))

	// Overtime risk
	mux.Handle("GET /api/v1/labor/overtime-risk", authMW(http.HandlerFunc(h.CheckOvertimeRisk)))

	// Swaps
	mux.Handle("POST /api/v1/labor/swaps", authMW(http.HandlerFunc(h.RequestSwap)))
	mux.Handle("GET /api/v1/labor/swaps", authMW(http.HandlerFunc(h.ListSwaps)))
	mux.Handle("PUT /api/v1/labor/swaps/{id}", authMW(http.HandlerFunc(h.ReviewSwap)))
}

// GenerateForecast generates a demand forecast for a given location and date.
// POST /api/v1/labor/forecast
// Body: {"location_id": "...", "date": "YYYY-MM-DD"}
func (h *LaborHandler) GenerateForecast(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	var body struct {
		LocationID string `json:"location_id"`
		Date       string `json:"date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}
	if body.LocationID == "" {
		WriteError(w, http.StatusBadRequest, "SCHED_MISSING_LOCATION", "location_id is required")
		return
	}

	targetDate := time.Now().Truncate(24 * time.Hour)
	if body.Date != "" {
		if d, err := time.Parse("2006-01-02", body.Date); err == nil {
			targetDate = d
		}
	}

	blocks, err := h.svc.GenerateForecast(r.Context(), orgID, body.LocationID, targetDate)
	if err != nil {
		slog.Error("sched forecast error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "SCHED_FORECAST_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{
		"location_id": body.LocationID,
		"date":        targetDate.Format("2006-01-02"),
		"forecast":    blocks,
	})
}

// GetForecast retrieves the stored forecast for a location and date.
// GET /api/v1/labor/forecast?location_id=...&date=YYYY-MM-DD
func (h *LaborHandler) GetForecast(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	locationID := r.URL.Query().Get("location_id")
	if locationID == "" {
		WriteError(w, http.StatusBadRequest, "SCHED_MISSING_LOCATION", "location_id is required")
		return
	}

	targetDate := time.Now().Truncate(24 * time.Hour)
	if d := r.URL.Query().Get("date"); d != "" {
		if parsed, err := time.Parse("2006-01-02", d); err == nil {
			targetDate = parsed
		}
	}

	blocks, err := h.svc.GetForecast(r.Context(), orgID, locationID, targetDate)
	if err != nil {
		slog.Error("sched forecast error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "SCHED_FORECAST_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{
		"location_id": locationID,
		"date":        targetDate.Format("2006-01-02"),
		"forecast":    blocks,
	})
}

// GenerateScheduleDraft generates a demand-driven draft schedule for a week.
// POST /api/v1/labor/schedules/generate
// Body: {"location_id": "...", "week_start": "YYYY-MM-DD", "created_by": "..."}
func (h *LaborHandler) GenerateScheduleDraft(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	var body struct {
		LocationID string `json:"location_id"`
		WeekStart  string `json:"week_start"`
		CreatedBy  string `json:"created_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}
	if body.LocationID == "" || body.WeekStart == "" {
		WriteError(w, http.StatusBadRequest, "SCHED_MISSING_FIELDS", "location_id and week_start are required")
		return
	}

	sched, err := h.svc.GenerateScheduleDraft(r.Context(), orgID, body.LocationID, body.WeekStart, body.CreatedBy)
	if err != nil {
		slog.Error("sched generate error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "SCHED_GENERATE_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusCreated, sched)
}

// GetSchedule retrieves a schedule with its shifts.
// GET /api/v1/labor/schedules?location_id=...&week_start=YYYY-MM-DD
func (h *LaborHandler) GetSchedule(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	locationID := r.URL.Query().Get("location_id")
	weekStart := r.URL.Query().Get("week_start")
	if locationID == "" || weekStart == "" {
		WriteError(w, http.StatusBadRequest, "SCHED_MISSING_FIELDS", "location_id and week_start are required")
		return
	}

	sched, err := h.svc.GetSchedule(r.Context(), orgID, locationID, weekStart)
	if err != nil {
		slog.Error("sched get error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "SCHED_GET_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, sched)
}

// UpdateSchedule replaces all shifts on a draft schedule.
// PUT /api/v1/labor/schedules/{id}
// Body: {"shifts": [...]}
func (h *LaborHandler) UpdateSchedule(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	scheduleID := r.PathValue("id")
	if scheduleID == "" {
		WriteError(w, http.StatusBadRequest, "SCHED_MISSING_ID", "schedule id is required")
		return
	}

	var body struct {
		Shifts []labor.ScheduledShift `json:"shifts"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}

	if err := h.svc.UpdateScheduleShifts(r.Context(), orgID, scheduleID, body.Shifts); err != nil {
		slog.Error("sched update error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "SCHED_UPDATE_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// PublishSchedule transitions a draft schedule to published.
// POST /api/v1/labor/schedules/{id}/publish
func (h *LaborHandler) PublishSchedule(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	scheduleID := r.PathValue("id")
	if scheduleID == "" {
		WriteError(w, http.StatusBadRequest, "SCHED_MISSING_ID", "schedule id is required")
		return
	}

	if err := h.svc.PublishSchedule(r.Context(), orgID, scheduleID); err != nil {
		slog.Error("sched publish error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "SCHED_PUBLISH_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "published"})
}

// GetEmployeeSchedule returns shifts for a single employee for a given week.
// GET /api/v1/labor/schedules/employee/{id}?week_start=YYYY-MM-DD
func (h *LaborHandler) GetEmployeeSchedule(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	employeeID := r.PathValue("id")
	if employeeID == "" {
		WriteError(w, http.StatusBadRequest, "SCHED_MISSING_EMPLOYEE", "employee id is required")
		return
	}

	weekStart := r.URL.Query().Get("week_start")
	if weekStart == "" {
		// Default to current week Monday.
		now := time.Now()
		offset := int(time.Monday - now.Weekday())
		if offset > 0 {
			offset -= 7
		}
		weekStart = now.AddDate(0, 0, offset).Format("2006-01-02")
	}

	shifts, err := h.svc.GetEmployeeSchedule(r.Context(), orgID, employeeID, weekStart)
	if err != nil {
		slog.Error("sched employee error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "SCHED_EMPLOYEE_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"shifts": shifts})
}

// ProjectLaborCost returns a labor cost projection for a schedule.
// GET /api/v1/labor/schedules/cost?schedule_id=X
func (h *LaborHandler) ProjectLaborCost(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	scheduleID := r.URL.Query().Get("schedule_id")
	if scheduleID == "" {
		WriteError(w, http.StatusBadRequest, "SCHED_MISSING_ID", "schedule_id is required")
		return
	}

	proj, err := h.svc.ProjectLaborCost(r.Context(), orgID, scheduleID)
	if err != nil {
		slog.Error("sched cost error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "SCHED_COST_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, proj)
}

// CheckOvertimeRisk returns employees at overtime risk for a given week.
// GET /api/v1/labor/overtime-risk?location_id=...&week_start=YYYY-MM-DD
func (h *LaborHandler) CheckOvertimeRisk(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	locationID := r.URL.Query().Get("location_id")
	if locationID == "" {
		WriteError(w, http.StatusBadRequest, "SCHED_MISSING_LOCATION", "location_id is required")
		return
	}

	weekStart := r.URL.Query().Get("week_start")
	if weekStart == "" {
		now := time.Now()
		offset := int(time.Monday - now.Weekday())
		if offset > 0 {
			offset -= 7
		}
		weekStart = now.AddDate(0, 0, offset).Format("2006-01-02")
	}

	risks, err := h.svc.CheckOvertimeRisk(r.Context(), orgID, locationID, weekStart)
	if err != nil {
		slog.Error("sched overtime error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "SCHED_OVERTIME_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"overtime_risks": risks})
}

// RequestSwap creates a shift-swap request.
// POST /api/v1/labor/swaps
// Body: {"requester_shift_id": "...", "target_employee_id": "...", "reason": "..."}
func (h *LaborHandler) RequestSwap(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	var body struct {
		RequesterShiftID string `json:"requester_shift_id"`
		TargetEmployeeID string `json:"target_employee_id"`
		Reason           string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}
	if body.RequesterShiftID == "" {
		WriteError(w, http.StatusBadRequest, "SWAP_MISSING_SHIFT", "requester_shift_id is required")
		return
	}

	req, err := h.svc.RequestSwap(r.Context(), orgID, body.RequesterShiftID, body.TargetEmployeeID, body.Reason)
	if err != nil {
		slog.Error("swap request error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "SWAP_REQUEST_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusCreated, req)
}

// ListSwaps returns swap requests, optionally filtered by location and status.
// GET /api/v1/labor/swaps?location_id=...&status=pending
func (h *LaborHandler) ListSwaps(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	locationID := r.URL.Query().Get("location_id")
	status := r.URL.Query().Get("status")

	reqs, err := h.svc.ListSwapRequests(r.Context(), orgID, locationID, status)
	if err != nil {
		slog.Error("swap list error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "SWAP_LIST_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"swap_requests": reqs})
}

// ReviewSwap approves or denies a swap request.
// PUT /api/v1/labor/swaps/{id}
// Body: {"approved": true, "reviewed_by": "..."}
func (h *LaborHandler) ReviewSwap(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	swapID := r.PathValue("id")
	if swapID == "" {
		WriteError(w, http.StatusBadRequest, "SWAP_MISSING_ID", "swap id is required")
		return
	}

	var body struct {
		Approved   bool   `json:"approved"`
		ReviewedBy string `json:"reviewed_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}
	if body.ReviewedBy == "" {
		WriteError(w, http.StatusBadRequest, "SWAP_MISSING_REVIEWER", "reviewed_by is required")
		return
	}

	if err := h.svc.ReviewSwap(r.Context(), orgID, swapID, body.Approved, body.ReviewedBy); err != nil {
		slog.Error("swap review error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "SWAP_REVIEW_ERROR", "an internal error occurred")
		return
	}

	status := "denied"
	if body.Approved {
		status = "approved"
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": status})
}
