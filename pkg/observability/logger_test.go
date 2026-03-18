package observability_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/opsnerve/fireline/pkg/observability"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLogger_JSONOutput(t *testing.T) {
	var buf bytes.Buffer
	logger := observability.NewLogger("info", &buf)

	logger.Info("test message", "key", "value")

	var entry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &entry)
	require.NoError(t, err)
	assert.Equal(t, "test message", entry["msg"])
	assert.Equal(t, "value", entry["key"])
	assert.Equal(t, "INFO", entry["level"])
}

func TestNewLogger_WithCorrelationID(t *testing.T) {
	var buf bytes.Buffer
	logger := observability.NewLogger("info", &buf)
	logger = logger.With("correlation_id", "req-123")

	logger.Info("correlated")

	var entry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &entry)
	require.NoError(t, err)
	assert.Equal(t, "req-123", entry["correlation_id"])
}
