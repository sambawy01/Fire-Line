package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/opsnerve/fireline/internal/portfolio"
	"github.com/opsnerve/fireline/internal/tenant"
)

// PortfolioHandler handles HTTP requests for multi-location portfolio endpoints.
type PortfolioHandler struct {
	svc *portfolio.Service
}

// NewPortfolioHandler creates a new PortfolioHandler.
func NewPortfolioHandler(svc *portfolio.Service) *PortfolioHandler {
	return &PortfolioHandler{svc: svc}
}

// RegisterRoutes registers all portfolio API routes.
func (h *PortfolioHandler) RegisterRoutes(mux *http.ServeMux, authMW func(http.Handler) http.Handler) {
	// Hierarchy
	mux.Handle("POST /api/v1/portfolio/nodes", authMW(http.HandlerFunc(h.CreateNode)))
	mux.Handle("GET /api/v1/portfolio/nodes", authMW(http.HandlerFunc(h.GetHierarchy)))
	mux.Handle("PUT /api/v1/portfolio/nodes/{id}", authMW(http.HandlerFunc(h.UpdateNode)))
	mux.Handle("DELETE /api/v1/portfolio/nodes/{id}", authMW(http.HandlerFunc(h.DeleteNode)))

	// Aggregation
	mux.Handle("GET /api/v1/portfolio/kpis", authMW(http.HandlerFunc(h.AggregateKPIs)))
	mux.Handle("GET /api/v1/portfolio/compare", authMW(http.HandlerFunc(h.CompareLocations)))

	// Benchmarking — specific before parameterized
	mux.Handle("POST /api/v1/portfolio/benchmarks/calculate", authMW(http.HandlerFunc(h.CalculateBenchmarks)))
	mux.Handle("GET /api/v1/portfolio/benchmarks", authMW(http.HandlerFunc(h.GetBenchmarks)))
	mux.Handle("GET /api/v1/portfolio/outliers", authMW(http.HandlerFunc(h.DetectOutliers)))

	// Best practices
	mux.Handle("POST /api/v1/portfolio/best-practices/detect", authMW(http.HandlerFunc(h.DetectBestPractices)))
	mux.Handle("GET /api/v1/portfolio/best-practices", authMW(http.HandlerFunc(h.ListBestPractices)))
	mux.Handle("POST /api/v1/portfolio/best-practices/{id}/adopt", authMW(http.HandlerFunc(h.AdoptPractice)))
	mux.Handle("POST /api/v1/portfolio/best-practices/{id}/dismiss", authMW(http.HandlerFunc(h.DismissPractice)))
}

// CreateNode handles POST /api/v1/portfolio/nodes
func (h *PortfolioHandler) CreateNode(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	var req struct {
		ParentNodeID *string `json:"parent_node_id"`
		Name         string  `json:"name"`
		NodeType     string  `json:"node_type"`
		LocationID   *string `json:"location_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}
	if req.Name == "" || req.NodeType == "" {
		WriteError(w, http.StatusBadRequest, "PORTFOLIO_MISSING_FIELDS", "name and node_type are required")
		return
	}

	node, err := h.svc.CreateNode(r.Context(), orgID, req.ParentNodeID, req.Name, req.NodeType, req.LocationID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "PORTFOLIO_CREATE_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, node)
}

// GetHierarchy handles GET /api/v1/portfolio/nodes
func (h *PortfolioHandler) GetHierarchy(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	nodes, err := h.svc.GetHierarchy(r.Context(), orgID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "PORTFOLIO_HIERARCHY_ERROR", err.Error())
		return
	}
	WriteList(w, http.StatusOK, "nodes", nodes)
}

// UpdateNode handles PUT /api/v1/portfolio/nodes/{id}
func (h *PortfolioHandler) UpdateNode(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	nodeID := r.PathValue("id")
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}
	if err := h.svc.UpdateNode(r.Context(), orgID, nodeID, req.Name); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			WriteError(w, http.StatusNotFound, "PORTFOLIO_NODE_NOT_FOUND", "node not found")
			return
		}
		WriteError(w, http.StatusInternalServerError, "PORTFOLIO_UPDATE_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// DeleteNode handles DELETE /api/v1/portfolio/nodes/{id}
func (h *PortfolioHandler) DeleteNode(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	nodeID := r.PathValue("id")
	if err := h.svc.DeleteNode(r.Context(), orgID, nodeID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			WriteError(w, http.StatusNotFound, "PORTFOLIO_NODE_NOT_FOUND", "node not found")
			return
		}
		WriteError(w, http.StatusInternalServerError, "PORTFOLIO_DELETE_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// AggregateKPIs handles GET /api/v1/portfolio/kpis?node_id=&from=&to=
func (h *PortfolioHandler) AggregateKPIs(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	nodeID := r.URL.Query().Get("node_id")
	if nodeID == "" {
		WriteError(w, http.StatusBadRequest, "PORTFOLIO_MISSING_NODE", "node_id is required")
		return
	}
	from, to := parseDateRange(r)
	kpis, err := h.svc.AggregateKPIs(r.Context(), orgID, nodeID, from, to)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "PORTFOLIO_KPI_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, kpis)
}

// CompareLocations handles GET /api/v1/portfolio/compare?location_ids=a,b,c&from=&to=
func (h *PortfolioHandler) CompareLocations(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	idsParam := r.URL.Query().Get("location_ids")
	if idsParam == "" {
		WriteError(w, http.StatusBadRequest, "PORTFOLIO_MISSING_LOCATIONS", "location_ids is required")
		return
	}
	locationIDs := strings.Split(idsParam, ",")
	from, to := parseDateRange(r)

	metrics, err := h.svc.GetLocationComparison(r.Context(), orgID, locationIDs, from, to)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "PORTFOLIO_COMPARE_ERROR", err.Error())
		return
	}
	WriteList(w, http.StatusOK, "locations", metrics)
}

// CalculateBenchmarks handles POST /api/v1/portfolio/benchmarks/calculate
func (h *PortfolioHandler) CalculateBenchmarks(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	from, to := parseDateRange(r)
	if err := h.svc.CalculateBenchmarks(r.Context(), orgID, from, to); err != nil {
		WriteError(w, http.StatusInternalServerError, "PORTFOLIO_BENCHMARK_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "calculated"})
}

// GetBenchmarks handles GET /api/v1/portfolio/benchmarks?from=&to=
func (h *PortfolioHandler) GetBenchmarks(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	from, to := parseDateRange(r)
	benchmarks, err := h.svc.GetBenchmarks(r.Context(), orgID, from, to)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "PORTFOLIO_BENCHMARK_FETCH_ERROR", err.Error())
		return
	}
	WriteList(w, http.StatusOK, "benchmarks", benchmarks)
}

// DetectOutliers handles GET /api/v1/portfolio/outliers?from=&to=
func (h *PortfolioHandler) DetectOutliers(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	from, to := parseDateRange(r)
	outliers, err := h.svc.DetectOutliers(r.Context(), orgID, from, to)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "PORTFOLIO_OUTLIER_ERROR", err.Error())
		return
	}
	WriteList(w, http.StatusOK, "outliers", outliers)
}

// DetectBestPractices handles POST /api/v1/portfolio/best-practices/detect
func (h *PortfolioHandler) DetectBestPractices(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	practices, err := h.svc.DetectBestPractices(r.Context(), orgID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "PORTFOLIO_DETECT_ERROR", err.Error())
		return
	}
	WriteList(w, http.StatusOK, "best_practices", practices)
}

// ListBestPractices handles GET /api/v1/portfolio/best-practices?status=
func (h *PortfolioHandler) ListBestPractices(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	status := r.URL.Query().Get("status")
	practices, err := h.svc.ListBestPractices(r.Context(), orgID, status)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "PORTFOLIO_PRACTICES_ERROR", err.Error())
		return
	}
	WriteList(w, http.StatusOK, "best_practices", practices)
}

// AdoptPractice handles POST /api/v1/portfolio/best-practices/{id}/adopt
func (h *PortfolioHandler) AdoptPractice(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	practiceID := r.PathValue("id")
	if err := h.svc.AdoptPractice(r.Context(), orgID, practiceID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			WriteError(w, http.StatusNotFound, "PORTFOLIO_PRACTICE_NOT_FOUND", "best practice not found")
			return
		}
		WriteError(w, http.StatusInternalServerError, "PORTFOLIO_ADOPT_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "adopted"})
}

// DismissPractice handles POST /api/v1/portfolio/best-practices/{id}/dismiss
func (h *PortfolioHandler) DismissPractice(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}
	practiceID := r.PathValue("id")
	if err := h.svc.DismissPractice(r.Context(), orgID, practiceID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			WriteError(w, http.StatusNotFound, "PORTFOLIO_PRACTICE_NOT_FOUND", "best practice not found")
			return
		}
		WriteError(w, http.StatusInternalServerError, "PORTFOLIO_DISMISS_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "dismissed"})
}
