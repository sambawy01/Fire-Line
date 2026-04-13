package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"

	"github.com/opsnerve/fireline/internal/adapter"
	"github.com/opsnerve/fireline/internal/adapter/loyverse"
	"github.com/opsnerve/fireline/internal/adapter/toast"
	"github.com/opsnerve/fireline/internal/alerting"
	"github.com/opsnerve/fireline/internal/api"
	"github.com/opsnerve/fireline/internal/auth"
	"github.com/opsnerve/fireline/internal/customer"
	"github.com/opsnerve/fireline/internal/event"
	"github.com/opsnerve/fireline/internal/financial"
	"github.com/opsnerve/fireline/internal/intelligence"
	"github.com/opsnerve/fireline/internal/inventory"
	"github.com/opsnerve/fireline/internal/labor"
	"github.com/opsnerve/fireline/internal/maintenance"
	"github.com/opsnerve/fireline/internal/marketing"
	"github.com/opsnerve/fireline/internal/menu"
	"github.com/opsnerve/fireline/internal/messaging"
	"github.com/opsnerve/fireline/internal/onboarding"
	"github.com/opsnerve/fireline/internal/operations"
	"github.com/opsnerve/fireline/internal/payroll"
	"github.com/opsnerve/fireline/internal/pipeline"
	"github.com/opsnerve/fireline/internal/portfolio"
	"github.com/opsnerve/fireline/internal/reporting"
	"github.com/opsnerve/fireline/internal/tasks"
	"github.com/opsnerve/fireline/internal/vendor"
	"github.com/opsnerve/fireline/pkg/config"
	"github.com/opsnerve/fireline/pkg/database"
	"github.com/opsnerve/fireline/pkg/observability"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger := observability.NewLogger(cfg.LogLevel, nil)
	slog.SetDefault(logger)

	ctx := context.Background()

	// App pool: fireline_app (RLS enforced) — for all tenant-scoped operations
	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to create database pool", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Admin pool: superuser — for pre-tenant operations (signup, login)
	adminPool, err := database.NewPool(ctx, cfg.AdminDatabaseURL)
	if err != nil {
		slog.Error("failed to create admin database pool", "error", err)
		os.Exit(1)
	}
	defer adminPool.Close()

	// ─── Redis ───
	var rdb *redis.Client
	redisOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		slog.Warn("invalid REDIS_URL, falling back to in-memory rate limiting", "error", err)
	} else {
		rdb = redis.NewClient(redisOpts)
		if err := rdb.Ping(ctx).Err(); err != nil {
			slog.Warn("redis unavailable, falling back to in-memory rate limiting", "error", err)
			rdb = nil
		} else {
			slog.Info("redis connected", "addr", redisOpts.Addr)
		}
	}
	defer func() {
		if rdb != nil {
			rdb.Close()
		}
	}()

	if err := pool.Ping(ctx); err != nil {
		slog.Error("failed to ping database", "error", err)
		os.Exit(1)
	}
	slog.Info("database connected")

	// JWT keys (ephemeral for dev, PEM file for production)
	var privKey *rsa.PrivateKey
	var pubKey *rsa.PublicKey
	if cfg.JWTPrivateKeyPath == "" {
		slog.Warn("no JWT key configured, generating ephemeral key (development only)")
		privKey, err = rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			slog.Error("failed to generate ephemeral RSA key", "error", err)
			os.Exit(1)
		}
		pubKey = &privKey.PublicKey
	} else {
		privPEM, err := os.ReadFile(cfg.JWTPrivateKeyPath)
		if err != nil {
			slog.Error("failed to read JWT private key", "path", cfg.JWTPrivateKeyPath, "error", err)
			os.Exit(1)
		}
		privKey, err = jwt.ParseRSAPrivateKeyFromPEM(privPEM)
		if err != nil {
			slog.Error("failed to parse JWT private key", "error", err)
			os.Exit(1)
		}
		pubKey = &privKey.PublicKey
	}

	// ─── Event Bus ───
	bus := event.New()

	// ─── Adapter Registry ───
	registry := adapter.NewRegistry()
	registry.RegisterFactory("toast", func() adapter.Adapter { return toast.New() })
	registry.RegisterFactory("loyverse", func() adapter.Adapter { return loyverse.New() })

	// Loyverse HTTP handler — uses a shared adapter instance wired to the event bus.
	loyverseAdapter := loyverse.NewWithBus(bus)
	loyverseHandler := loyverse.NewHandler(loyverseAdapter)

	// ─── Data Pipeline ───
	// Register pipeline handlers BEFORE auto-connecting adapters so the initial
	// sync events are captured by the pipeline.
	pipe := pipeline.New(pool.Raw(), bus)
	pipe.RegisterHandlers()

	// Auto-connect Loyverse if env vars are set.
	if loyToken := os.Getenv("LOYVERSE_API_TOKEN"); loyToken != "" {
		loyStoreID := os.Getenv("LOYVERSE_STORE_ID")
		loyOrgID := os.Getenv("LOYVERSE_ORG_ID")
		loyLocID := os.Getenv("LOYVERSE_LOCATION_ID")
		if loyOrgID == "" {
			// Fall back: look up the first org in the DB.
			_ = adminPool.Raw().QueryRow(ctx, "SELECT org_id FROM organizations LIMIT 1").Scan(&loyOrgID)
		}
		if err := loyverseAdapter.Initialize(ctx, adapter.Config{
			AdapterType: "loyverse",
			OrgID:       loyOrgID,
			LocationID:  loyLocID,
			Credentials: map[string]string{
				"api_token": loyToken,
				"store_id":  loyStoreID,
			},
		}); err != nil {
			slog.Error("loyverse auto-connect failed", "error", err)
		} else {
			slog.Info("loyverse auto-connected", "store_id", loyStoreID, "location_id", loyLocID)
		}
	}

	// ─── Intelligence Services ───
	invSvc := inventory.New(pool.Raw(), bus)
	invSvc.RegisterHandlers()

	finSvc := financial.New(pool.Raw(), bus)
	finSvc.RegisterHandlers()

	menuSvc := menu.New(pool.Raw(), bus)
	laborSvc := labor.New(pool.Raw(), bus)
	vendorSvc := vendor.New(pool.Raw(), bus)

	ollamaURL := os.Getenv("OLLAMA_URL")
	ollamaModel := os.Getenv("OLLAMA_MODEL")
	ollamaClient := customer.NewOllamaClient(ollamaURL, ollamaModel)
	customerSvc := customer.New(pool.Raw(), bus, ollamaClient)
	opsSvc := operations.New(pool.Raw(), bus)

	// ─── Alerting ───
	alertSvc := alerting.New(bus, adminPool.Raw())
	alertSvc.RegisterDefaultRules()

	// Seed demo alerts if the demo org exists
	{
		var demoOrgID string
		err := adminPool.Raw().QueryRow(ctx, "SELECT org_id FROM organizations LIMIT 1").Scan(&demoOrgID)
		if err == nil && demoOrgID != "" {
			alertSvc.SeedAlerts(demoOrgID, []string{
				"a1111111-1111-1111-1111-111111111111",
				"b2222222-2222-2222-2222-222222222222",
				"c3333333-3333-3333-3333-333333333333",
				"d4444444-4444-4444-4444-444444444444",
			})
		}
	}

	// ─── Reporting ───
	reportSvc := reporting.New(pool.Raw(), bus, alertSvc)

	// ─── Marketing ───
	mktSvc := marketing.New(pool.Raw(), bus)

	// ─── Portfolio ───
	portfolioSvc := portfolio.New(pool.Raw(), bus)

	// ─── Onboarding ───
	onboardingSvc := onboarding.New(pool.Raw(), bus)

	// ─── Tasks ───
	tasksSvc := tasks.New(pool.Raw(), bus)

	// ─── Intelligence ───
	intelSvc := intelligence.New(pool.Raw(), bus)
	intelSvc.RegisterHandlers()

	// ─── Messaging ───
	msgSvc := messaging.New(pool.Raw(), bus)

	// ─── Payroll ───
	payrollSvc := payroll.New(pool.Raw(), bus)

	// ─── Maintenance ───
	maintSvc := maintenance.New(pool.Raw(), bus)

	slog.Info("all modules initialized",
		"event_bus", "ready",
		"pipeline", "ready",
		"inventory", "ready",
		"financial", "ready",
		"menu", "ready",
		"labor", "ready",
		"vendor", "ready",
		"customer", "ready",
		"operations", "ready",
		"alerting", "ready",
		"reporting", "ready",
		"marketing", "ready",
		"portfolio", "ready",
		"onboarding", "ready",
		"maintenance", "ready",
		"tasks", "ready",
		"intelligence", "ready",
		"messaging", "ready",
		"payroll", "ready",
	)

	// ─── Auth ───
	issuer := auth.NewTokenIssuer(privKey, pubKey, 15*time.Minute)
	authService := auth.NewService(pool.Raw(), adminPool.Raw(), issuer)
	authHandler := auth.NewHandler(authService, issuer)
	authMW := auth.AuthMiddleware(issuer)

	// ─── Metrics ───
	metrics := observability.NewMetrics()

	// ─── HTTP Router ───
	mux := http.NewServeMux()

	// Metrics endpoint (no auth)
	mux.HandleFunc("GET /metrics", metrics.Handler())

	// Health (no auth)
	mux.HandleFunc("GET /health/live", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"status":"ok"}`)
	})
	mux.HandleFunc("GET /health/ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		dbStatus := "ok"
		if err := pool.Ping(r.Context()); err != nil {
			dbStatus = "error"
		}

		redisStatus := "not_configured"
		if rdb != nil {
			if err := rdb.Ping(r.Context()).Err(); err != nil {
				redisStatus = "error"
			} else {
				redisStatus = "ok"
			}
		}

		status := "ready"
		if dbStatus != "ok" {
			status = "not_ready"
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		json.NewEncoder(w).Encode(map[string]any{
			"status": status,
			"modules": map[string]any{
				"database":  dbStatus,
				"redis":     redisStatus,
				"event_bus": "ok",
				"adapters":  map[string]string{"toast": "registered", "loyverse": "registered"},
			},
		})
	})

	// ─── Permission Middleware Helpers ───
	requirePerm := auth.RequirePermission

	// authMW + permission composition helper
	authWithPerm := func(perm string) func(http.Handler) http.Handler {
		return func(h http.Handler) http.Handler {
			return authMW(requirePerm(perm)(h))
		}
	}

	// ─── Rate Limiters ───
	authRL := api.NewRateLimiter(rdb, 10, time.Minute)  // 10 req/min for auth endpoints
	apiRL := api.NewRateLimiter(rdb, 100, time.Minute) // 100 req/min for general API

	// Auth routes (no auth middleware — these create/validate auth)
	// Wrap auth routes with rate limiting
	authMux := http.NewServeMux()
	authHandler.RegisterRoutes(authMux)
	mux.Handle("POST /api/v1/auth/", api.RateLimit(authRL)(authMux))

	// Module API routes (auth required + permission checks)

	// Inventory: inventory:read for GET, inventory:write for writes
	invHandler := api.NewInventoryHandler(invSvc)
	invHandler.RegisterRoutes(mux, authWithPerm("inventory:read"))

	// Financial: financial:read for all (handler registers only GET + POST budget)
	finHandler := api.NewFinancialHandler(finSvc)
	finHandler.RegisterRoutes(mux, authWithPerm("financial:read"))

	// Alerting: reporting:read (alerts are read by anyone with reporting access)
	alertHandler := api.NewAlertingHandler(alertSvc)
	alertHandler.RegisterRoutes(mux, authMW)

	// Locations: any authenticated user can see their own locations
	locHandler := api.NewLocationHandler(pool.Raw())
	locHandler.RegisterRoutes(mux, authMW)

	// Menu: menu:read
	menuHandler := api.NewMenuHandler(menuSvc)
	menuHandler.RegisterRoutes(mux, authWithPerm("menu:read"))

	// Labor: labor schedule-related permissions (handler already has role checks internally)
	laborHandler := api.NewLaborHandler(laborSvc)
	laborHandler.RegisterRoutes(mux, authMW)

	// Vendor: vendor:read
	vendorHandler := api.NewVendorHandler(vendorSvc)
	vendorHandler.RegisterRoutes(mux, authWithPerm("vendor:read"))

	// Customer: customer:read
	customerHandler := api.NewCustomerHandler(customerSvc)
	customerHandler.RegisterRoutes(mux, authWithPerm("customer:read"))

	// Operations: operations:kitchen
	opsHandler := api.NewOperationsHandler(opsSvc)
	opsHandler.RegisterRoutes(mux, authWithPerm("operations:kitchen"))

	// Reporting: reporting:read
	reportHandler := api.NewReportingHandler(reportSvc)
	reportHandler.RegisterRoutes(mux, authWithPerm("reporting:read"))

	// Marketing: marketing:read
	mktHandler := api.NewMarketingHandler(mktSvc)
	mktHandler.RegisterRoutes(mux, authWithPerm("marketing:read"))

	// Portfolio: portfolio:read
	portfolioHandler := api.NewPortfolioHandler(portfolioSvc)
	portfolioHandler.RegisterRoutes(mux, authWithPerm("portfolio:read"))

	// Onboarding: system:admin (only owners set up orgs)
	onboardingHandler := api.NewOnboardingHandler(onboardingSvc)
	onboardingHandler.RegisterRoutes(mux, authWithPerm("system:admin"))

	// Maintenance: handler already has internal role checks
	maintHandler := api.NewMaintenanceHandler(maintSvc)
	maintHandler.RegisterRoutes(mux, authMW)

	// Tasks: handler already has internal role checks
	tasksHandler := api.NewTasksHandler(tasksSvc)
	tasksHandler.RegisterRoutes(mux, authMW)

	// Intelligence: handler already has internal role checks
	intelHandler := api.NewIntelligenceHandler(intelSvc)
	intelHandler.RegisterRoutes(mux, authMW)

	// Messaging: handler already has internal role checks
	msgHandler := api.NewMessagingHandler(msgSvc)
	msgHandler.RegisterRoutes(mux, authMW)

	// Payroll: financial:read (handler already has internal role checks)
	payrollHandler := api.NewPayrollHandler(payrollSvc)
	payrollHandler.RegisterRoutes(mux, authMW)

	// Activity feed: any authenticated user
	activityHandler := api.NewActivityHandler(pool.Raw(), alertSvc)
	activityHandler.RegisterRoutes(mux, authMW)

	// GDPR: system:admin (only admins can erase/export guest data)
	gdprHandler := api.NewGDPRHandler(pool.Raw())
	gdprHandler.RegisterRoutes(mux, authWithPerm("system:admin"))

	// Loyverse adapter routes: integrations:manage
	loyverseHandler.RegisterRoutes(mux, authWithPerm("integrations:manage"))

	// ─── Middleware Chain ───
	// Order (outermost first): CORS -> Security Headers -> Body Size -> Rate Limit -> Metrics -> Correlation -> Logger -> Recovery -> mux
	handler := corsMiddleware(cfg,
		api.SecurityHeaders(
			api.MaxBodySize(1<<20)( // 1 MB body limit
				api.RateLimit(apiRL)(
					observability.MetricsMiddleware(metrics)(
						api.CorrelationID(
							api.RequestLogger(
								api.Recovery(mux))))))))

	// Suppress unused variable warnings
	_ = registry
	_ = pipe

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("starting server", "port", cfg.Port, "env", cfg.Env)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server")
	registry.ShutdownAll(ctx)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown error", "error", err)
	}
	slog.Info("server stopped")
}

func corsMiddleware(cfg *config.Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowedOrigins := strings.Split(cfg.AllowedOrigins, ",")
		for _, allowed := range allowedOrigins {
			if strings.TrimSpace(allowed) == origin {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				break
			}
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400")
		w.Header().Set("Vary", "Origin")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
