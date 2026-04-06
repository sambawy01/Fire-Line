package api

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/opsnerve/fireline/internal/tenant"
	"github.com/opsnerve/fireline/pkg/database"
)

// GDPRHandler handles GDPR right-to-erasure and right-to-portability requests.
type GDPRHandler struct {
	pool *pgxpool.Pool
}

func NewGDPRHandler(pool *pgxpool.Pool) *GDPRHandler {
	return &GDPRHandler{pool: pool}
}

func (h *GDPRHandler) RegisterRoutes(mux *http.ServeMux, authMW func(http.Handler) http.Handler) {
	mux.Handle("DELETE /api/v1/gdpr/guest/{guestID}", authMW(http.HandlerFunc(h.deleteGuestData)))
	mux.Handle("GET /api/v1/gdpr/guest/{guestID}/export", authMW(http.HandlerFunc(h.exportGuestData)))
}

// deleteGuestData handles right-to-erasure (GDPR Article 17).
// It removes visit history and anonymizes the guest profile, keeping the row
// for aggregate analytics but stripping all personally identifiable information.
func (h *GDPRHandler) deleteGuestData(w http.ResponseWriter, r *http.Request) {
	guestID := r.PathValue("guestID")
	if guestID == "" {
		WriteError(w, http.StatusBadRequest, "MISSING_GUEST_ID", "guest ID is required")
		return
	}

	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	ctx := tenant.WithOrgID(r.Context(), orgID)

	err = database.TenantTx(ctx, h.pool, func(tx pgx.Tx) error {
		// Delete guest visits
		_, err := tx.Exec(ctx, "DELETE FROM guest_visits WHERE guest_id = $1", guestID)
		if err != nil {
			return fmt.Errorf("delete guest visits: %w", err)
		}

		// Anonymize guest profile (keep for aggregate analytics but remove PII)
		_, err = tx.Exec(ctx,
			`UPDATE guest_profiles SET
				name = 'REDACTED',
				email = NULL,
				phone = NULL,
				favorite_items = '[]',
				notes = NULL,
				updated_at = now()
			WHERE guest_id = $1`, guestID)
		if err != nil {
			return fmt.Errorf("anonymize guest profile: %w", err)
		}

		return nil
	})

	if err != nil {
		slog.Error("gdpr delete failed", "error", err, "guest_id", guestID)
		WriteError(w, http.StatusInternalServerError, "GDPR_DELETE_ERROR", "an internal error occurred")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// exportGuestData handles right-to-portability (GDPR Article 20).
// It returns all stored data for a guest in a structured JSON format.
func (h *GDPRHandler) exportGuestData(w http.ResponseWriter, r *http.Request) {
	guestID := r.PathValue("guestID")
	if guestID == "" {
		WriteError(w, http.StatusBadRequest, "MISSING_GUEST_ID", "guest ID is required")
		return
	}

	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	ctx := tenant.WithOrgID(r.Context(), orgID)

	type guestExport struct {
		Profile interface{} `json:"profile"`
		Visits  interface{} `json:"visits"`
	}

	var export guestExport

	err = database.TenantTx(ctx, h.pool, func(tx pgx.Tx) error {
		// Get profile
		rows, err := tx.Query(ctx,
			`SELECT guest_id, name, email, phone, first_seen, last_seen, visit_count,
				total_spend, avg_check, favorite_items, clv_score, churn_risk
			 FROM guest_profiles WHERE guest_id = $1`, guestID)
		if err != nil {
			return fmt.Errorf("query guest profile: %w", err)
		}
		defer rows.Close()

		cols := rows.FieldDescriptions()
		if rows.Next() {
			values, err := rows.Values()
			if err != nil {
				return fmt.Errorf("scan profile: %w", err)
			}
			profile := make(map[string]interface{})
			for i, col := range cols {
				profile[string(col.Name)] = values[i]
			}
			export.Profile = profile
		}
		rows.Close()

		// Get visits
		visitRows, err := tx.Query(ctx,
			`SELECT visit_id, check_id, visited_at, spend, items_ordered
			 FROM guest_visits WHERE guest_id = $1 ORDER BY visited_at DESC`, guestID)
		if err != nil {
			return fmt.Errorf("query visits: %w", err)
		}
		defer visitRows.Close()

		var visits []map[string]interface{}
		visitCols := visitRows.FieldDescriptions()
		for visitRows.Next() {
			values, err := visitRows.Values()
			if err != nil {
				return fmt.Errorf("scan visit: %w", err)
			}
			visit := make(map[string]interface{})
			for i, col := range visitCols {
				visit[string(col.Name)] = values[i]
			}
			visits = append(visits, visit)
		}
		export.Visits = visits

		return nil
	})

	if err != nil {
		slog.Error("gdpr export failed", "error", err, "guest_id", guestID)
		WriteError(w, http.StatusInternalServerError, "GDPR_EXPORT_ERROR", "an internal error occurred")
		return
	}

	WriteJSON(w, http.StatusOK, export)
}
