package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/opsnerve/fireline/internal/api"
	"github.com/opsnerve/fireline/internal/auth"
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

	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to create database pool", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

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
		// Load from file — implement later
		slog.Error("JWT key file loading not yet implemented")
		os.Exit(1)
	}

	issuer := auth.NewTokenIssuer(privKey, &privKey.PublicKey, 15*time.Minute)
	authService := auth.NewService(pool.Raw(), issuer)
	authHandler := auth.NewHandler(authService, issuer)

	mux := http.NewServeMux()

	// Health (no auth)
	mux.HandleFunc("GET /health/live", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"status":"ok"}`)
	})
	mux.HandleFunc("GET /health/ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := pool.Ping(r.Context()); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintln(w, `{"status":"not_ready","error":"database"}`)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"status":"ready"}`)
	})

	// Auth routes
	authHandler.RegisterRoutes(mux)

	// Middleware chain
	handler := api.CorrelationID(api.RequestLogger(api.Recovery(mux)))

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
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown error", "error", err)
	}
	slog.Info("server stopped")
}
