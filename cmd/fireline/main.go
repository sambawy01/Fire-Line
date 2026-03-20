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
	"syscall"
	"time"

	"github.com/opsnerve/fireline/internal/adapter"
	"github.com/opsnerve/fireline/internal/adapter/toast"
	"github.com/opsnerve/fireline/internal/alerting"
	"github.com/opsnerve/fireline/internal/api"
	"github.com/opsnerve/fireline/internal/auth"
	"github.com/opsnerve/fireline/internal/customer"
	"github.com/opsnerve/fireline/internal/event"
	"github.com/opsnerve/fireline/internal/financial"
	"github.com/opsnerve/fireline/internal/inventory"
	"github.com/opsnerve/fireline/internal/labor"
	"github.com/opsnerve/fireline/internal/marketing"
	"github.com/opsnerve/fireline/internal/menu"
	"github.com/opsnerve/fireline/internal/operations"
	"github.com/opsnerve/fireline/internal/pipeline"
	"github.com/opsnerve/fireline/internal/portfolio"
	"github.com/opsnerve/fireline/internal/reporting"
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

	if err := pool.Ping(ctx); err != nil {
		slog.Error("failed to ping database", "error", err)
		os.Exit(1)
	}
	slog.Info("database connected")

	// JWT keys (ephemeral for dev)
	var privKey *rsa.PrivateKey
	if cfg.JWTPrivateKeyPath == "" {
		slog.Warn("no JWT key configured, generating ephemeral key (development only)")
		privKey, err = rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			slog.Error("failed to generate ephemeral RSA key", "error", err)
			os.Exit(1)
		}
	} else {
		slog.Error("JWT key file loading not yet implemented")
		os.Exit(1)
	}

	// ─── Event Bus ───
	bus := event.New()

	// ─── Adapter Registry ───
	registry := adapter.NewRegistry()
	registry.RegisterFactory("toast", func() adapter.Adapter { return toast.New() })

	// ─── Data Pipeline ───
	pipe := pipeline.New(pool.Raw(), bus)
	pipe.RegisterHandlers()

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
	alertSvc := alerting.New(bus)
	alertSvc.RegisterDefaultRules()

	// Seed demo alerts if the demo org exists
	if cfg.Env == "development" {
		var demoOrgID string
		err := adminPool.Raw().QueryRow(ctx, "SELECT org_id FROM organizations WHERE slug = 'bistro-cloud'").Scan(&demoOrgID)
		if err == nil && demoOrgID != "" {
			alertSvc.SeedAlerts(demoOrgID, []string{
				"a1111111-1111-1111-1111-111111111111",
				"b2222222-2222-2222-2222-222222222222",
			})
		}
	}

	// ─── Reporting ───
	reportSvc := reporting.New(pool.Raw(), bus, alertSvc)

	// ─── Marketing ───
	mktSvc := marketing.New(pool.Raw(), bus)

	// ─── Portfolio ───
	portfolioSvc := portfolio.New(pool.Raw(), bus)

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
	)

	// ─── Auth ───
	issuer := auth.NewTokenIssuer(privKey, &privKey.PublicKey, 15*time.Minute)
	authService := auth.NewService(pool.Raw(), adminPool.Raw(), issuer)
	authHandler := auth.NewHandler(authService, issuer)
	authMW := auth.AuthMiddleware(issuer)

	// ─── HTTP Router ───
	mux := http.NewServeMux()

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
				"event_bus": "ok",
				"adapters":  map[string]string{"toast": "registered"},
			},
		})
	})

	// Auth routes (no auth middleware — these create/validate auth)
	authHandler.RegisterRoutes(mux)

	// Module API routes (auth required)
	invHandler := api.NewInventoryHandler(invSvc)
	invHandler.RegisterRoutes(mux, authMW)

	finHandler := api.NewFinancialHandler(finSvc)
	finHandler.RegisterRoutes(mux, authMW)

	alertHandler := api.NewAlertingHandler(alertSvc)
	alertHandler.RegisterRoutes(mux, authMW)

	locHandler := api.NewLocationHandler(adminPool.Raw())
	locHandler.RegisterRoutes(mux, authMW)

	menuHandler := api.NewMenuHandler(menuSvc)
	menuHandler.RegisterRoutes(mux, authMW)

	laborHandler := api.NewLaborHandler(laborSvc)
	laborHandler.RegisterRoutes(mux, authMW)

	vendorHandler := api.NewVendorHandler(vendorSvc)
	vendorHandler.RegisterRoutes(mux, authMW)

	customerHandler := api.NewCustomerHandler(customerSvc)
	customerHandler.RegisterRoutes(mux, authMW)

	opsHandler := api.NewOperationsHandler(opsSvc)
	opsHandler.RegisterRoutes(mux, authMW)

	reportHandler := api.NewReportingHandler(reportSvc)
	reportHandler.RegisterRoutes(mux, authMW)

	mktHandler := api.NewMarketingHandler(mktSvc)
	mktHandler.RegisterRoutes(mux, authMW)

	portfolioHandler := api.NewPortfolioHandler(portfolioSvc)
	portfolioHandler.RegisterRoutes(mux, authMW)

	// CORS for frontend dev
	handler := corsMiddleware(api.CorrelationID(api.RequestLogger(api.Recovery(mux))))

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

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
