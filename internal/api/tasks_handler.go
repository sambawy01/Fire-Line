package api

import (
	"log/slog"
	"encoding/json"
	"net/http"

	"github.com/opsnerve/fireline/internal/auth"
	"github.com/opsnerve/fireline/internal/tasks"
	"github.com/opsnerve/fireline/internal/tenant"
)

// TasksHandler handles task management API requests.
type TasksHandler struct {
	svc *tasks.Service
}

// NewTasksHandler creates a new TasksHandler.
func NewTasksHandler(svc *tasks.Service) *TasksHandler {
	return &TasksHandler{svc: svc}
}

// RegisterRoutes registers task management API routes on the given mux.
func (h *TasksHandler) RegisterRoutes(mux *http.ServeMux, authMW func(http.Handler) http.Handler) {
	managerRoles := requireRole("gm", "ops_director", "owner")
	shiftRoles := requireRole("shift_manager", "gm", "ops_director", "owner")
	allRoles := requireRole("staff", "shift_manager", "gm", "ops_director", "owner")

	// Task Templates
	mux.Handle("POST /api/v1/task-templates", chain(http.HandlerFunc(h.CreateTemplate), authMW, managerRoles))
	mux.Handle("GET /api/v1/task-templates", chain(http.HandlerFunc(h.ListTemplates), authMW, shiftRoles))
	mux.Handle("POST /api/v1/task-templates/{id}/instantiate", chain(http.HandlerFunc(h.InstantiateTemplate), authMW, shiftRoles))

	// Tasks — specific paths before parameterized ones
	mux.Handle("POST /api/v1/tasks", chain(http.HandlerFunc(h.CreateTask), authMW, shiftRoles))
	mux.Handle("GET /api/v1/tasks/my", chain(http.HandlerFunc(h.GetMyTasks), authMW, allRoles))
	mux.Handle("GET /api/v1/tasks", chain(http.HandlerFunc(h.ListTasks), authMW, shiftRoles))
	mux.Handle("PUT /api/v1/tasks/{id}/status", chain(http.HandlerFunc(h.UpdateTaskStatus), authMW, allRoles))
	mux.Handle("PUT /api/v1/tasks/{id}/complete", chain(http.HandlerFunc(h.CompleteTask), authMW, allRoles))

	// Announcements
	mux.Handle("POST /api/v1/announcements", chain(http.HandlerFunc(h.CreateAnnouncement), authMW, shiftRoles))
	mux.Handle("GET /api/v1/announcements", chain(http.HandlerFunc(h.ListAnnouncements), authMW, allRoles))
}

// ─── Task Template Handlers ─────────────────────────────────────────────────

// CreateTemplate creates a new task template.
// POST /api/v1/task-templates
func (h *TasksHandler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	var input tasks.TemplateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}
	if input.LocationID == "" || input.Name == "" {
		WriteError(w, http.StatusBadRequest, "TASK_MISSING_FIELDS", "location_id and name are required")
		return
	}

	tmpl, err := h.svc.CreateTemplate(r.Context(), orgID, input)
	if err != nil {
		slog.Error("task create template error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "TASK_CREATE_TEMPLATE_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusCreated, tmpl)
}

// ListTemplates returns task templates, optionally filtered by location.
// GET /api/v1/task-templates?location_id=...
func (h *TasksHandler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	locationID := r.URL.Query().Get("location_id")

	templates, err := h.svc.ListTemplates(r.Context(), orgID, locationID)
	if err != nil {
		slog.Error("task list templates error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "TASK_LIST_TEMPLATES_ERROR", "an internal error occurred")
		return
	}
	WriteList(w, http.StatusOK, "templates", templates)
}

// InstantiateTemplate creates tasks from a template.
// POST /api/v1/task-templates/{id}/instantiate
// Body: {"assigned_to": "...", "assigned_by": "..."}
func (h *TasksHandler) InstantiateTemplate(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	templateID := r.PathValue("id")
	if templateID == "" {
		WriteError(w, http.StatusBadRequest, "TASK_MISSING_ID", "template id is required")
		return
	}

	var body struct {
		AssignedTo string `json:"assigned_to"`
		AssignedBy string `json:"assigned_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}

	created, err := h.svc.InstantiateTemplate(r.Context(), orgID, templateID, body.AssignedTo, body.AssignedBy)
	if err != nil {
		slog.Error("task instantiate error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "TASK_INSTANTIATE_ERROR", "an internal error occurred")
		return
	}
	WriteList(w, http.StatusCreated, "tasks", created)
}

// ─── Task Handlers ──────────────────────────────────────────────────────────

// CreateTask creates a new task.
// POST /api/v1/tasks
func (h *TasksHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	var input tasks.TaskInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}
	if input.LocationID == "" || input.Title == "" {
		WriteError(w, http.StatusBadRequest, "TASK_MISSING_FIELDS", "location_id and title are required")
		return
	}

	task, err := h.svc.CreateTask(r.Context(), orgID, input)
	if err != nil {
		slog.Error("task create error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "TASK_CREATE_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusCreated, task)
}

// ListTasks returns tasks with optional filters.
// GET /api/v1/tasks?location_id=...&assigned_to=...&status=...
func (h *TasksHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	locationID := r.URL.Query().Get("location_id")
	assignedTo := r.URL.Query().Get("assigned_to")
	status := r.URL.Query().Get("status")

	taskList, err := h.svc.ListTasks(r.Context(), orgID, locationID, assignedTo, status)
	if err != nil {
		slog.Error("task list error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "TASK_LIST_ERROR", "an internal error occurred")
		return
	}
	WriteList(w, http.StatusOK, "tasks", taskList)
}

// GetMyTasks returns the authenticated user's pending and in-progress tasks.
// GET /api/v1/tasks/my
func (h *TasksHandler) GetMyTasks(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	userID := auth.UserIDFrom(r.Context())
	if userID == "" {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_USER", "no user context")
		return
	}

	myTasks, err := h.svc.GetMyTasks(r.Context(), orgID, userID)
	if err != nil {
		slog.Error("task my error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "TASK_MY_ERROR", "an internal error occurred")
		return
	}
	WriteList(w, http.StatusOK, "tasks", myTasks)
}

// UpdateTaskStatus updates the status of a task.
// PUT /api/v1/tasks/{id}/status
// Body: {"status": "in_progress"}
func (h *TasksHandler) UpdateTaskStatus(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	taskID := r.PathValue("id")
	if taskID == "" {
		WriteError(w, http.StatusBadRequest, "TASK_MISSING_ID", "task id is required")
		return
	}

	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}
	if body.Status == "" {
		WriteError(w, http.StatusBadRequest, "TASK_MISSING_STATUS", "status is required")
		return
	}

	if err := h.svc.UpdateTaskStatus(r.Context(), orgID, taskID, body.Status); err != nil {
		slog.Error("task update status error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "TASK_UPDATE_STATUS_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": body.Status})
}

// CompleteTask marks a task as completed.
// PUT /api/v1/tasks/{id}/complete
// Body: {"data_entry_value": {...}, "photo_url": "...", "completed_by": "..."}
func (h *TasksHandler) CompleteTask(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	taskID := r.PathValue("id")
	if taskID == "" {
		WriteError(w, http.StatusBadRequest, "TASK_MISSING_ID", "task id is required")
		return
	}

	var input tasks.CompleteInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}
	if input.CompletedBy == "" {
		WriteError(w, http.StatusBadRequest, "TASK_MISSING_COMPLETED_BY", "completed_by is required")
		return
	}

	task, err := h.svc.CompleteTask(r.Context(), orgID, taskID, input)
	if err != nil {
		slog.Error("task complete error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "TASK_COMPLETE_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusOK, task)
}

// ─── Announcement Handlers ──────────────────────────────────────────────────

// CreateAnnouncement creates a new team announcement.
// POST /api/v1/announcements
func (h *TasksHandler) CreateAnnouncement(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	var input tasks.AnnouncementInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}
	if input.LocationID == "" || input.Title == "" || input.Body == "" {
		WriteError(w, http.StatusBadRequest, "TASK_MISSING_FIELDS", "location_id, title, and body are required")
		return
	}

	announcement, err := h.svc.CreateAnnouncement(r.Context(), orgID, input)
	if err != nil {
		slog.Error("task create announcement error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "TASK_CREATE_ANNOUNCEMENT_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusCreated, announcement)
}

// ListAnnouncements returns active announcements, optionally filtered by location.
// GET /api/v1/announcements?location_id=...
func (h *TasksHandler) ListAnnouncements(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	locationID := r.URL.Query().Get("location_id")

	announcements, err := h.svc.ListAnnouncements(r.Context(), orgID, locationID)
	if err != nil {
		slog.Error("task list announcements error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "TASK_LIST_ANNOUNCEMENTS_ERROR", "an internal error occurred")
		return
	}
	WriteList(w, http.StatusOK, "announcements", announcements)
}
