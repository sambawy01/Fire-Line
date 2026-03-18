package config_test

import (
	"testing"

	"github.com/opsnerve/fireline/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_Defaults(t *testing.T) {
	cfg, err := config.Load()
	require.NoError(t, err)
	assert.Equal(t, "8080", cfg.Port)
	assert.Equal(t, "info", cfg.LogLevel)
}

func TestLoad_FromEnv(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/test?sslmode=disable")
	t.Setenv("REDIS_URL", "redis://localhost:6379/0")
	t.Setenv("LOG_LEVEL", "debug")

	cfg, err := config.Load()
	require.NoError(t, err)
	assert.Equal(t, "9090", cfg.Port)
	assert.Equal(t, "postgres://test:test@localhost:5432/test?sslmode=disable", cfg.DatabaseURL)
	assert.Equal(t, "redis://localhost:6379/0", cfg.RedisURL)
	assert.Equal(t, "debug", cfg.LogLevel)
}

func TestLoad_RequiredFields(t *testing.T) {
	// DATABASE_URL is required in production mode
	t.Setenv("ENV", "production")
	t.Setenv("DATABASE_URL", "")

	_, err := config.Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "DATABASE_URL")
}
