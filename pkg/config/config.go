package config

import (
	"fmt"
	"os"
)

type Config struct {
	Env               string // "development" or "production"
	Port              string
	DatabaseURL       string
	AdminDatabaseURL  string // superuser — for signup/login (pre-tenant, bypasses RLS)
	RedisURL          string
	LogLevel          string
	JWTPrivateKeyPath string
	JWTPublicKeyPath  string
	AllowedOrigins    string // comma-separated list of allowed CORS origins
}

func Load() (*Config, error) {
	cfg := &Config{
		Env:               getEnv("ENV", "development"),
		Port:              getEnv("PORT", "8080"),
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://fireline_app:fireline_app@localhost:5432/fireline?sslmode=disable"),
		AdminDatabaseURL:  getEnv("ADMIN_DATABASE_URL", "postgres://fireline:fireline@localhost:5432/fireline?sslmode=disable"),
		RedisURL:          getEnv("REDIS_URL", "redis://localhost:6379/0"),
		LogLevel:          getEnv("LOG_LEVEL", "info"),
		JWTPrivateKeyPath: getEnv("JWT_PRIVATE_KEY_PATH", ""),
		JWTPublicKeyPath:  getEnv("JWT_PUBLIC_KEY_PATH", ""),
		AllowedOrigins:    getEnv("ALLOWED_ORIGINS", "http://localhost:5173,http://localhost:5174"),
	}

	if cfg.Env == "production" {
		if cfg.DatabaseURL == "" || cfg.DatabaseURL == "postgres://fireline_app:fireline_app@localhost:5432/fireline?sslmode=disable" {
			return nil, fmt.Errorf("DATABASE_URL is required in production")
		}
		if cfg.AdminDatabaseURL == "" || cfg.AdminDatabaseURL == "postgres://fireline:fireline@localhost:5432/fireline?sslmode=disable" {
			return nil, fmt.Errorf("ADMIN_DATABASE_URL must not use default localhost value in production")
		}
		if cfg.JWTPrivateKeyPath == "" {
			return nil, fmt.Errorf("JWT_PRIVATE_KEY_PATH must be set in production")
		}
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
