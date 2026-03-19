package toast_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/opsnerve/fireline/internal/adapter"
	"github.com/opsnerve/fireline/internal/adapter/toast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newInitializedAdapter(t *testing.T) *toast.ToastAdapter {
	t.Helper()
	a := toast.New().(*toast.ToastAdapter)
	err := a.Initialize(context.Background(), adapter.Config{
		AdapterID:   "test-adapter",
		AdapterType: "toast",
		OrgID:       "org-test",
		LocationID:  "loc-test",
		SyncMode:    adapter.SyncRealtime,
	})
	require.NoError(t, err)
	return a
}

func TestToastAdapter_TypeAndCapabilities(t *testing.T) {
	a := toast.New()
	assert.Equal(t, "toast", a.Type())

	caps := a.Capabilities()
	assert.Contains(t, caps, adapter.CapReadOrders)
	assert.Contains(t, caps, adapter.CapReadMenu)
	assert.Contains(t, caps, adapter.CapReadEmployees)
	assert.Contains(t, caps, adapter.CapWrite86Status)
}

func TestToastAdapter_Lifecycle(t *testing.T) {
	a := toast.New()
	assert.Equal(t, adapter.StatusInitializing, a.GetStatus())

	err := a.Initialize(context.Background(), adapter.Config{
		AdapterType: "toast",
		OrgID:       "org-1",
		LocationID:  "loc-1",
	})
	require.NoError(t, err)
	assert.Equal(t, adapter.StatusActive, a.GetStatus())

	err = a.HealthCheck(context.Background())
	require.NoError(t, err)

	err = a.Shutdown(context.Background())
	require.NoError(t, err)
	assert.Equal(t, adapter.StatusDisconnected, a.GetStatus())

	err = a.HealthCheck(context.Background())
	assert.Error(t, err)
}

func TestToastAdapter_ReadOrders(t *testing.T) {
	a := newInitializedAdapter(t)

	orders, err := a.ReadOrders(context.Background(), time.Now().Add(-24*time.Hour), 5)
	require.NoError(t, err)
	assert.Len(t, orders, 5)

	for _, o := range orders {
		assert.Equal(t, "org-test", o.OrgID)
		assert.Equal(t, "loc-test", o.LocationID)
		assert.Equal(t, "toast", o.Source)
		assert.Equal(t, "closed", o.Status)
		assert.NotEmpty(t, o.Items)
		assert.Greater(t, o.Total, int64(0))
		assert.Contains(t, []string{"dine_in", "takeout", "delivery"}, o.Channel)
	}

	f, err := a.GetDataFreshness("orders")
	require.NoError(t, err)
	assert.Equal(t, int64(5), f.RecordCount)
}

func TestToastAdapter_ReadMenu(t *testing.T) {
	a := newInitializedAdapter(t)

	items, err := a.ReadMenu(context.Background())
	require.NoError(t, err)
	assert.Len(t, items, 8)

	for _, item := range items {
		assert.Equal(t, "org-test", item.OrgID)
		assert.Equal(t, "loc-test", item.LocationID)
		assert.Equal(t, "toast", item.Source)
		assert.NotEmpty(t, item.Name)
		assert.Greater(t, item.Price, int64(0))
		assert.True(t, item.Available)
	}
}

func TestToastAdapter_ReadEmployees(t *testing.T) {
	a := newInitializedAdapter(t)

	employees, err := a.ReadEmployees(context.Background())
	require.NoError(t, err)
	assert.Len(t, employees, 5)

	for _, emp := range employees {
		assert.Equal(t, "org-test", emp.OrgID)
		assert.Equal(t, "loc-test", emp.LocationID)
		assert.Equal(t, "toast", emp.Source)
		assert.NotEmpty(t, emp.FirstName)
		assert.True(t, emp.Active)
	}
}

func TestToastAdapter_Write86Status(t *testing.T) {
	a := newInitializedAdapter(t)

	// 86 an item
	err := a.Write86Status(context.Background(), "toast-mi-1", false)
	require.NoError(t, err)

	// Verify it shows as unavailable in menu
	items, err := a.ReadMenu(context.Background())
	require.NoError(t, err)

	for _, item := range items {
		if item.ExternalID == "toast-mi-1" {
			assert.False(t, item.Available, "item should be 86'd")
		}
	}

	// Un-86 the item
	err = a.Write86Status(context.Background(), "toast-mi-1", true)
	require.NoError(t, err)

	items, err = a.ReadMenu(context.Background())
	require.NoError(t, err)
	for _, item := range items {
		if item.ExternalID == "toast-mi-1" {
			assert.True(t, item.Available, "item should be available again")
		}
	}
}

func TestToastAdapter_NotActive(t *testing.T) {
	a := toast.New()
	// Not initialized — should fail on reads

	_, err := a.(*toast.ToastAdapter).ReadOrders(context.Background(), time.Now(), 5)
	assert.Error(t, err)

	_, err = a.(*toast.ToastAdapter).ReadMenu(context.Background())
	assert.Error(t, err)

	err = a.(*toast.ToastAdapter).Write86Status(context.Background(), "x", true)
	assert.Error(t, err)
}

func TestValidateWebhookSignature(t *testing.T) {
	secret := "test-secret-key"
	payload := []byte(`{"event_type":"order.created","location_id":"loc-1"}`)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	validSig := hex.EncodeToString(mac.Sum(nil))

	assert.True(t, toast.ValidateWebhookSignature(payload, validSig, secret))
	assert.False(t, toast.ValidateWebhookSignature(payload, "invalidsig", secret))
	assert.False(t, toast.ValidateWebhookSignature(payload, validSig, "wrong-secret"))
}

func TestToastAdapter_RegistryIntegration(t *testing.T) {
	reg := adapter.NewRegistry()
	reg.RegisterFactory("toast", func() adapter.Adapter { return toast.New() })

	a, err := reg.Create(context.Background(), adapter.Config{
		AdapterID:   "toast-1",
		AdapterType: "toast",
		OrgID:       "org-1",
		LocationID:  "loc-1",
	})
	require.NoError(t, err)
	assert.Equal(t, "toast", a.Type())
	assert.Equal(t, adapter.StatusActive, a.GetStatus())
}
