package api

import (
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/opsnerve/fireline/internal/auth"
	"github.com/opsnerve/fireline/internal/tenant"
)

type LocationResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	OrgID string `json:"org_id"`
}

type LocationHandler struct {
	pool *pgxpool.Pool
}

func NewLocationHandler(pool *pgxpool.Pool) *LocationHandler {
	return &LocationHandler{pool: pool}
}

func (h *LocationHandler) RegisterRoutes(mux *http.ServeMux, authMW func(http.Handler) http.Handler) {
	mux.Handle("GET /api/v1/locations", authMW(http.HandlerFunc(h.GetLocations)))
}

func (h *LocationHandler) GetLocations(w http.ResponseWriter, r *http.Request) {
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

	rows, err := h.pool.Query(r.Context(), `
		SELECT l.location_id, l.name, l.org_id
		FROM locations l
		JOIN user_location_access ula ON ula.location_id = l.location_id
		WHERE ula.user_id = $1 AND ula.org_id = $2 AND l.status = 'active'
		ORDER BY l.name
	`, userID, orgID)
	if err != nil {
		slog.Error("location query error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "LOCATION_QUERY_ERROR", "an internal error occurred")
		return
	}
	defer rows.Close()

	locations := []LocationResponse{}
	for rows.Next() {
		var loc LocationResponse
		if err := rows.Scan(&loc.ID, &loc.Name, &loc.OrgID); err != nil {
			slog.Error("location scan error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
			WriteError(w, http.StatusInternalServerError, "LOCATION_SCAN_ERROR", "an internal error occurred")
			return
		}
		locations = append(locations, loc)
	}

	WriteJSON(w, http.StatusOK, map[string]any{"locations": locations})
}
