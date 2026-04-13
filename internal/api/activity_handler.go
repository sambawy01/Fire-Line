package api

import (
	"fmt"
	"log/slog"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/opsnerve/fireline/internal/alerting"
	"github.com/opsnerve/fireline/internal/tenant"
)

// ActivityItem represents a single entry in the unified activity feed.
type ActivityItem struct {
	ID        string `json:"id"`
	Type      string `json:"type"`      // "order", "alert", "clock_in"
	Title     string `json:"title"`
	Detail    string `json:"detail"`
	Timestamp string `json:"timestamp"` // RFC3339
}

// ActivityHandler serves the unified activity feed endpoint.
type ActivityHandler struct {
	pool     *pgxpool.Pool
	alertSvc *alerting.Service
}

// NewActivityHandler creates a new ActivityHandler.
func NewActivityHandler(pool *pgxpool.Pool, alertSvc *alerting.Service) *ActivityHandler {
	return &ActivityHandler{pool: pool, alertSvc: alertSvc}
}

// RegisterRoutes registers activity feed routes on the given mux.
func (h *ActivityHandler) RegisterRoutes(mux *http.ServeMux, authMW func(http.Handler) http.Handler) {
	mux.Handle("GET /api/v1/activity/recent", authMW(http.HandlerFunc(h.GetRecent)))
}

// GetRecent returns a unified, time-sorted activity feed combining recent
// orders, alerts, and employee clock-ins for a location.
func (h *ActivityHandler) GetRecent(w http.ResponseWriter, r *http.Request) {
	_, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	locationID := r.URL.Query().Get("location_id")
	if locationID == "" {
		WriteError(w, http.StatusBadRequest, "ACTIVITY_MISSING_LOCATION", "location_id is required")
		return
	}

	limit := 10
	if v := r.URL.Query().Get("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 && parsed <= 50 {
			limit = parsed
		}
	}

	correlationID := r.Header.Get("X-Request-ID")
	ctx := r.Context()
	var items []ActivityItem

	// ── Recent orders ──
	orderRows, err := h.pool.Query(ctx, `
		SELECT check_id, order_number, channel, total, status, closed_at
		FROM checks
		WHERE location_id = $1 AND closed_at IS NOT NULL
		ORDER BY closed_at DESC
		LIMIT 5
	`, locationID)
	if err != nil {
		slog.Error("activity: order query error", "error", err, "correlation_id", correlationID)
	} else {
		defer orderRows.Close()
		for orderRows.Next() {
			var checkID, orderNumber, channel, status string
			var total float64
			var closedAt time.Time
			if err := orderRows.Scan(&checkID, &orderNumber, &channel, &total, &status, &closedAt); err != nil {
				slog.Error("activity: order scan error", "error", err, "correlation_id", correlationID)
				continue
			}
			items = append(items, ActivityItem{
				ID:        checkID,
				Type:      "order",
				Title:     fmt.Sprintf("Order #%s — $%.2f", orderNumber, total),
				Detail:    fmt.Sprintf("%s via %s", status, channel),
				Timestamp: closedAt.Format(time.RFC3339),
			})
		}
	}

	// ── Recent alerts (from alerting service) ──
	orgID, _ := tenant.OrgIDFrom(ctx)
	if h.alertSvc != nil {
		alerts := h.alertSvc.GetQueue(orgID, locationID)
		alertLimit := 3
		if len(alerts) < alertLimit {
			alertLimit = len(alerts)
		}
		for i := 0; i < alertLimit; i++ {
			a := alerts[i]
			items = append(items, ActivityItem{
				ID:        a.AlertID,
				Type:      "alert",
				Title:     a.Title,
				Detail:    fmt.Sprintf("[%s] %s", a.Severity, a.Description),
				Timestamp: a.CreatedAt.Format(time.RFC3339),
			})
		}
	}

	// ── Recent clock-ins ──
	clockRows, err := h.pool.Query(ctx, `
		SELECT e.display_name, s.clock_in
		FROM shifts s
		JOIN employees e ON e.employee_id = s.employee_id
		WHERE s.location_id = $1 AND s.clock_in IS NOT NULL
		ORDER BY s.clock_in DESC
		LIMIT 3
	`, locationID)
	if err != nil {
		slog.Error("activity: clock-in query error", "error", err, "correlation_id", correlationID)
	} else {
		defer clockRows.Close()
		for clockRows.Next() {
			var displayName string
			var clockIn time.Time
			if err := clockRows.Scan(&displayName, &clockIn); err != nil {
				slog.Error("activity: clock-in scan error", "error", err, "correlation_id", correlationID)
				continue
			}
			items = append(items, ActivityItem{
				ID:        fmt.Sprintf("clockin-%s-%d", locationID, clockIn.Unix()),
				Type:      "clock_in",
				Title:     fmt.Sprintf("%s clocked in", displayName),
				Detail:    clockIn.Format("3:04 PM"),
				Timestamp: clockIn.Format(time.RFC3339),
			})
		}
	}

	// ── Sort by timestamp descending and apply limit ──
	sort.Slice(items, func(i, j int) bool {
		return items[i].Timestamp > items[j].Timestamp
	})
	if len(items) > limit {
		items = items[:limit]
	}

	WriteJSON(w, http.StatusOK, map[string]any{"activity": items})
}
