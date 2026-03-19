package adapter_test

import (
	"context"
	"testing"
	"time"

	"github.com/opsnerve/fireline/internal/adapter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockAdapter implements adapter.Adapter for testing.
type mockAdapter struct {
	adapterType  string
	caps         []adapter.Capability
	status       adapter.Status
	initialized  bool
	shutdownCalled bool
	healthErr    error
	cfg          adapter.Config
}

func (m *mockAdapter) Type() string                     { return m.adapterType }
func (m *mockAdapter) Capabilities() []adapter.Capability { return m.caps }
func (m *mockAdapter) HasCapability(cap adapter.Capability) bool {
	for _, c := range m.caps {
		if c == cap {
			return true
		}
	}
	return false
}
func (m *mockAdapter) Initialize(ctx context.Context, cfg adapter.Config) error {
	m.initialized = true
	m.cfg = cfg
	m.status = adapter.StatusActive
	return nil
}
func (m *mockAdapter) Shutdown(ctx context.Context) error {
	m.shutdownCalled = true
	m.status = adapter.StatusDisconnected
	return nil
}
func (m *mockAdapter) HealthCheck(ctx context.Context) error { return m.healthErr }
func (m *mockAdapter) GetStatus() adapter.Status            { return m.status }
func (m *mockAdapter) GetDataFreshness(dataType string) (*adapter.DataFreshness, error) {
	return &adapter.DataFreshness{
		DataType:   dataType,
		LastSyncAt: time.Now(),
	}, nil
}

func newMockFactory(adapterType string, caps ...adapter.Capability) adapter.Factory {
	return func() adapter.Adapter {
		return &mockAdapter{adapterType: adapterType, caps: caps, status: adapter.StatusInitializing}
	}
}

func TestRegistry_RegisterAndCreate(t *testing.T) {
	reg := adapter.NewRegistry()
	reg.RegisterFactory("toast", newMockFactory("toast", adapter.CapReadOrders, adapter.CapReadMenu))

	cfg := adapter.Config{
		AdapterID:   "a-1",
		AdapterType: "toast",
		OrgID:       "org-1",
		LocationID:  "loc-1",
		SyncMode:    adapter.SyncRealtime,
	}

	a, err := reg.Create(context.Background(), cfg)
	require.NoError(t, err)
	assert.Equal(t, "toast", a.Type())
	assert.Equal(t, adapter.StatusActive, a.GetStatus())
	assert.True(t, a.HasCapability(adapter.CapReadOrders))
	assert.False(t, a.HasCapability(adapter.CapWrite86Status))
}

func TestRegistry_Get(t *testing.T) {
	reg := adapter.NewRegistry()
	reg.RegisterFactory("toast", newMockFactory("toast"))

	cfg := adapter.Config{AdapterID: "a-1", AdapterType: "toast"}
	_, err := reg.Create(context.Background(), cfg)
	require.NoError(t, err)

	a, ok := reg.Get("a-1")
	assert.True(t, ok)
	assert.NotNil(t, a)

	_, ok = reg.Get("nonexistent")
	assert.False(t, ok)
}

func TestRegistry_Remove(t *testing.T) {
	reg := adapter.NewRegistry()
	reg.RegisterFactory("toast", newMockFactory("toast"))

	cfg := adapter.Config{AdapterID: "a-1", AdapterType: "toast"}
	_, err := reg.Create(context.Background(), cfg)
	require.NoError(t, err)

	err = reg.Remove(context.Background(), "a-1")
	require.NoError(t, err)

	_, ok := reg.Get("a-1")
	assert.False(t, ok)
}

func TestRegistry_RemoveNotFound(t *testing.T) {
	reg := adapter.NewRegistry()
	err := reg.Remove(context.Background(), "nonexistent")
	assert.Error(t, err)
}

func TestRegistry_UnknownType(t *testing.T) {
	reg := adapter.NewRegistry()
	cfg := adapter.Config{AdapterID: "a-1", AdapterType: "unknown"}
	_, err := reg.Create(context.Background(), cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown adapter type")
}

func TestRegistry_List(t *testing.T) {
	reg := adapter.NewRegistry()
	reg.RegisterFactory("toast", newMockFactory("toast"))

	reg.Create(context.Background(), adapter.Config{AdapterID: "a-1", AdapterType: "toast"})
	reg.Create(context.Background(), adapter.Config{AdapterID: "a-2", AdapterType: "toast"})

	ids := reg.List()
	assert.Len(t, ids, 2)
	assert.Contains(t, ids, "a-1")
	assert.Contains(t, ids, "a-2")
}

func TestRegistry_ShutdownAll(t *testing.T) {
	reg := adapter.NewRegistry()
	reg.RegisterFactory("toast", newMockFactory("toast"))

	reg.Create(context.Background(), adapter.Config{AdapterID: "a-1", AdapterType: "toast"})
	reg.Create(context.Background(), adapter.Config{AdapterID: "a-2", AdapterType: "toast"})

	reg.ShutdownAll(context.Background())

	assert.Empty(t, reg.List())
}

func TestAdapter_Capabilities(t *testing.T) {
	m := &mockAdapter{
		adapterType: "toast",
		caps: []adapter.Capability{
			adapter.CapReadOrders,
			adapter.CapReadMenu,
			adapter.CapWrite86Status,
		},
	}

	assert.True(t, m.HasCapability(adapter.CapReadOrders))
	assert.True(t, m.HasCapability(adapter.CapReadMenu))
	assert.True(t, m.HasCapability(adapter.CapWrite86Status))
	assert.False(t, m.HasCapability(adapter.CapReadEmployees))
}
