package toast

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"sync"
	"time"

	"github.com/opsnerve/fireline/internal/adapter"
)

// ToastAdapter is a mock implementation of a Toast POS adapter.
// In production this would call the real Toast API; here it generates
// realistic mock data for development and integration testing.
type ToastAdapter struct {
	mu        sync.RWMutex
	cfg       adapter.Config
	status    adapter.Status
	freshness map[string]*adapter.DataFreshness
	items86   map[string]bool // item external ID -> available
}

var _ adapter.Adapter = (*ToastAdapter)(nil)
var _ adapter.OrderReader = (*ToastAdapter)(nil)
var _ adapter.MenuReader = (*ToastAdapter)(nil)
var _ adapter.EmployeeReader = (*ToastAdapter)(nil)
var _ adapter.StatusWriter = (*ToastAdapter)(nil)

func New() adapter.Adapter {
	return &ToastAdapter{
		status:    adapter.StatusInitializing,
		freshness: make(map[string]*adapter.DataFreshness),
		items86:   make(map[string]bool),
	}
}

func (t *ToastAdapter) Type() string { return "toast" }

func (t *ToastAdapter) Capabilities() []adapter.Capability {
	return []adapter.Capability{
		adapter.CapReadOrders,
		adapter.CapReadMenu,
		adapter.CapReadEmployees,
		adapter.CapWrite86Status,
	}
}

func (t *ToastAdapter) HasCapability(cap adapter.Capability) bool {
	for _, c := range t.Capabilities() {
		if c == cap {
			return true
		}
	}
	return false
}

func (t *ToastAdapter) Initialize(ctx context.Context, cfg adapter.Config) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.cfg = cfg
	t.status = adapter.StatusActive
	slog.Info("toast adapter initialized", "location_id", cfg.LocationID)
	return nil
}

func (t *ToastAdapter) Shutdown(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.status = adapter.StatusDisconnected
	return nil
}

func (t *ToastAdapter) HealthCheck(ctx context.Context) error {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.status != adapter.StatusActive {
		return fmt.Errorf("toast adapter not active: %s", t.status)
	}
	return nil
}

func (t *ToastAdapter) GetStatus() adapter.Status {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.status
}

func (t *ToastAdapter) GetDataFreshness(dataType string) (*adapter.DataFreshness, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	f, ok := t.freshness[dataType]
	if !ok {
		return nil, fmt.Errorf("no freshness data for %s", dataType)
	}
	return f, nil
}

// ReadOrders generates mock order data simulating Toast POS output.
func (t *ToastAdapter) ReadOrders(ctx context.Context, since time.Time, limit int) ([]adapter.NormalizedOrder, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.status != adapter.StatusActive {
		return nil, fmt.Errorf("adapter not active")
	}

	channels := []string{"dine_in", "takeout", "delivery"}
	menuItems := []struct {
		name  string
		price int64
	}{
		{"Cheeseburger", 1495},
		{"Caesar Salad", 1295},
		{"Fish & Chips", 1795},
		{"Margherita Pizza", 1595},
		{"Chicken Wings", 1395},
		{"Grilled Salmon", 2295},
		{"Pasta Carbonara", 1695},
		{"Club Sandwich", 1295},
	}

	orders := make([]adapter.NormalizedOrder, 0, limit)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < limit; i++ {
		openTime := since.Add(time.Duration(i) * time.Minute)
		closeTime := openTime.Add(time.Duration(20+rng.Intn(40)) * time.Minute)

		numItems := 1 + rng.Intn(4)
		items := make([]adapter.NormalizedOrderItem, numItems)
		var subtotal int64
		for j := 0; j < numItems; j++ {
			mi := menuItems[rng.Intn(len(menuItems))]
			qty := 1 + rng.Intn(2)
			items[j] = adapter.NormalizedOrderItem{
				ExternalID: fmt.Sprintf("toast-item-%d-%d", i, j),
				Name:       mi.name,
				Quantity:   qty,
				UnitPrice:  mi.price,
			}
			subtotal += mi.price * int64(qty)
		}

		tax := subtotal * 8 / 100 // 8% tax
		tip := subtotal * int64(10+rng.Intn(15)) / 100

		orders = append(orders, adapter.NormalizedOrder{
			ExternalID:  fmt.Sprintf("toast-order-%d", 1000+i),
			OrgID:       t.cfg.OrgID,
			LocationID:  t.cfg.LocationID,
			OrderNumber: fmt.Sprintf("%d", 1000+i),
			Status:      "closed",
			Channel:     channels[rng.Intn(len(channels))],
			Items:       items,
			Subtotal:    subtotal,
			Tax:         tax,
			Total:       subtotal + tax,
			Tip:         tip,
			OpenedAt:    openTime,
			ClosedAt:    &closeTime,
			Source:      "toast",
		})
	}

	t.freshness["orders"] = &adapter.DataFreshness{
		DataType:    "orders",
		LastSyncAt:  time.Now(),
		RecordCount: int64(len(orders)),
	}

	return orders, nil
}

// ReadMenu generates mock menu data simulating Toast POS output.
func (t *ToastAdapter) ReadMenu(ctx context.Context) ([]adapter.NormalizedMenuItem, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.status != adapter.StatusActive {
		return nil, fmt.Errorf("adapter not active")
	}

	items := []adapter.NormalizedMenuItem{
		{ExternalID: "toast-mi-1", Name: "Cheeseburger", Category: "Burgers", Price: 1495, Available: true},
		{ExternalID: "toast-mi-2", Name: "Caesar Salad", Category: "Salads", Price: 1295, Available: true},
		{ExternalID: "toast-mi-3", Name: "Fish & Chips", Category: "Entrees", Price: 1795, Available: true},
		{ExternalID: "toast-mi-4", Name: "Margherita Pizza", Category: "Pizza", Price: 1595, Available: true},
		{ExternalID: "toast-mi-5", Name: "Chicken Wings", Category: "Appetizers", Price: 1395, Available: true},
		{ExternalID: "toast-mi-6", Name: "Grilled Salmon", Category: "Entrees", Price: 2295, Available: true},
		{ExternalID: "toast-mi-7", Name: "Pasta Carbonara", Category: "Pasta", Price: 1695, Available: true},
		{ExternalID: "toast-mi-8", Name: "Club Sandwich", Category: "Sandwiches", Price: 1295, Available: true},
	}

	for i := range items {
		items[i].OrgID = t.cfg.OrgID
		items[i].LocationID = t.cfg.LocationID
		items[i].Source = "toast"
		// Apply 86 status
		if avail, ok := t.items86[items[i].ExternalID]; ok {
			items[i].Available = avail
		}
	}

	t.freshness["menu"] = &adapter.DataFreshness{
		DataType:    "menu",
		LastSyncAt:  time.Now(),
		RecordCount: int64(len(items)),
	}

	return items, nil
}

// ReadEmployees generates mock employee data simulating Toast POS output.
func (t *ToastAdapter) ReadEmployees(ctx context.Context) ([]adapter.NormalizedEmployee, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.status != adapter.StatusActive {
		return nil, fmt.Errorf("adapter not active")
	}

	employees := []adapter.NormalizedEmployee{
		{ExternalID: "toast-emp-1", FirstName: "Maria", LastName: "Garcia", Role: "manager", Active: true},
		{ExternalID: "toast-emp-2", FirstName: "James", LastName: "Chen", Role: "line_cook", Active: true},
		{ExternalID: "toast-emp-3", FirstName: "Sarah", LastName: "Johnson", Role: "server", Active: true},
		{ExternalID: "toast-emp-4", FirstName: "David", LastName: "Kim", Role: "bartender", Active: true},
		{ExternalID: "toast-emp-5", FirstName: "Emily", LastName: "Rodriguez", Role: "host", Active: true},
	}

	for i := range employees {
		employees[i].OrgID = t.cfg.OrgID
		employees[i].LocationID = t.cfg.LocationID
		employees[i].Source = "toast"
	}

	return employees, nil
}

// Write86Status marks an item as available or unavailable (86'd).
func (t *ToastAdapter) Write86Status(ctx context.Context, itemID string, available bool) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.status != adapter.StatusActive {
		return fmt.Errorf("adapter not active")
	}

	t.items86[itemID] = available
	slog.Info("toast 86 status updated",
		"item_id", itemID,
		"available", available,
		"location_id", t.cfg.LocationID,
	)
	return nil
}

// ValidateWebhookSignature validates a Toast webhook HMAC-SHA256 signature.
func ValidateWebhookSignature(payload []byte, signature string, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

// WebhookPayload represents a Toast webhook event.
type WebhookPayload struct {
	EventType  string          `json:"event_type"`
	LocationID string          `json:"location_id"`
	Timestamp  time.Time       `json:"timestamp"`
	Data       json.RawMessage `json:"data"`
}
