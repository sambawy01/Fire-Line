// Package loyverse implements a FireLine adapter for the Loyverse POS system.
// It reads menu items, receipts, employees, and inventory via the Loyverse REST
// API and can write 86 (out-of-stock) status back by zeroing variant stock.
//
// Configuration is sourced from environment variables:
//
//	LOYVERSE_API_TOKEN     — merchant API token (required)
//	LOYVERSE_STORE_ID      — store ID for inventory/multi-store (optional)
//	LOYVERSE_POLL_INTERVAL — polling interval in minutes (default: 5)
package loyverse

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/opsnerve/fireline/internal/adapter"
	"github.com/opsnerve/fireline/internal/event"
)

// LoyverseAdapter implements adapter.Adapter and the optional reader/writer interfaces.
type LoyverseAdapter struct {
	mu        sync.RWMutex
	cfg       adapter.Config
	status    adapter.Status
	freshness map[string]*adapter.DataFreshness

	client  *Client
	syncer  *Syncer
	bus     *event.Bus
	storeID string

	// cancelPoll stops the background polling goroutine.
	cancelPoll context.CancelFunc
}

// compile-time interface assertions
var _ adapter.Adapter = (*LoyverseAdapter)(nil)
var _ adapter.OrderReader = (*LoyverseAdapter)(nil)
var _ adapter.MenuReader = (*LoyverseAdapter)(nil)
var _ adapter.EmployeeReader = (*LoyverseAdapter)(nil)
var _ adapter.StatusWriter = (*LoyverseAdapter)(nil)

// New returns a new LoyverseAdapter. The event bus is optional; pass nil to
// disable event publishing.
func New() adapter.Adapter {
	return &LoyverseAdapter{
		status:    adapter.StatusInitializing,
		freshness: make(map[string]*adapter.DataFreshness),
	}
}

// NewWithBus returns a LoyverseAdapter wired to the given event bus.
// It returns the concrete type so callers can pass it to NewHandler.
func NewWithBus(bus *event.Bus) *LoyverseAdapter {
	return &LoyverseAdapter{
		status:    adapter.StatusInitializing,
		freshness: make(map[string]*adapter.DataFreshness),
		bus:       bus,
	}
}

// --- adapter.Adapter ---

func (a *LoyverseAdapter) Type() string { return "loyverse" }

func (a *LoyverseAdapter) Capabilities() []adapter.Capability {
	return []adapter.Capability{
		adapter.CapReadOrders,
		adapter.CapReadMenu,
		adapter.CapReadEmployees,
		adapter.CapReadInventory,
		adapter.CapWrite86Status,
	}
}

func (a *LoyverseAdapter) HasCapability(cap adapter.Capability) bool {
	for _, c := range a.Capabilities() {
		if c == cap {
			return true
		}
	}
	return false
}

// Initialize sets up the adapter from config + environment variables and
// optionally starts background polling.
func (a *LoyverseAdapter) Initialize(ctx context.Context, cfg adapter.Config) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Pull credentials from config map first, then fall back to environment.
	token := credOrEnv(cfg.Credentials, "api_token", "LOYVERSE_API_TOKEN")
	if token == "" {
		return fmt.Errorf("loyverse: LOYVERSE_API_TOKEN is required")
	}

	a.storeID = credOrEnv(cfg.Credentials, "store_id", "LOYVERSE_STORE_ID")
	a.cfg = cfg
	a.client = NewClient(token)
	a.syncer = newSyncer(a.client, a.bus, cfg, a.storeID)

	// Determine poll interval.
	pollInterval := cfg.PollInterval
	if pollInterval == 0 {
		pollInterval = parsePollInterval()
	}

	// Start background polling.
	pollCtx, cancel := context.WithCancel(context.Background())
	a.cancelPoll = cancel
	a.syncer.StartPolling(pollCtx, pollInterval)

	a.status = adapter.StatusActive
	slog.Info("loyverse adapter initialized",
		"location_id", cfg.LocationID,
		"store_id", a.storeID,
		"poll_interval", pollInterval,
	)
	return nil
}

// Shutdown stops polling and marks the adapter disconnected.
func (a *LoyverseAdapter) Shutdown(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.cancelPoll != nil {
		a.cancelPoll()
	}
	a.status = adapter.StatusDisconnected
	slog.Info("loyverse adapter shut down", "location_id", a.cfg.LocationID)
	return nil
}

// HealthCheck verifies the adapter is active and can reach the Loyverse API.
func (a *LoyverseAdapter) HealthCheck(ctx context.Context) error {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.status != adapter.StatusActive {
		return fmt.Errorf("loyverse adapter not active: %s", a.status)
	}
	// Light probe: fetch store list (small payload).
	if _, err := a.client.GetStores(); err != nil {
		return fmt.Errorf("loyverse health check failed: %w", err)
	}
	return nil
}

func (a *LoyverseAdapter) GetStatus() adapter.Status {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.status
}

func (a *LoyverseAdapter) GetDataFreshness(dataType string) (*adapter.DataFreshness, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	f, ok := a.freshness[dataType]
	if !ok {
		return nil, fmt.Errorf("loyverse: no freshness data for %s", dataType)
	}
	return f, nil
}

// --- adapter.OrderReader ---

// ReadOrders fetches Loyverse receipts since `since` (up to `limit` items
// per page; Loyverse paginates automatically).
func (a *LoyverseAdapter) ReadOrders(ctx context.Context, since time.Time, limit int) ([]adapter.NormalizedOrder, error) {
	a.mu.RLock()
	if a.status != adapter.StatusActive {
		a.mu.RUnlock()
		return nil, fmt.Errorf("loyverse adapter not active")
	}
	syncer := a.syncer
	a.mu.RUnlock()

	orders, err := syncer.SyncOrders(ctx, since)
	if err != nil {
		a.setStatus(adapter.StatusErrored)
		return nil, err
	}

	// Apply limit (Loyverse paginates; we respect the caller's limit here).
	if limit > 0 && len(orders) > limit {
		orders = orders[:limit]
	}

	a.mu.Lock()
	a.freshness["orders"] = &adapter.DataFreshness{
		DataType:    "orders",
		LastSyncAt:  time.Now(),
		RecordCount: int64(len(orders)),
	}
	a.mu.Unlock()

	return orders, nil
}

// --- adapter.MenuReader ---

// ReadMenu fetches and maps all Loyverse items.
func (a *LoyverseAdapter) ReadMenu(ctx context.Context) ([]adapter.NormalizedMenuItem, error) {
	a.mu.RLock()
	if a.status != adapter.StatusActive {
		a.mu.RUnlock()
		return nil, fmt.Errorf("loyverse adapter not active")
	}
	syncer := a.syncer
	a.mu.RUnlock()

	items, err := syncer.SyncMenu(ctx)
	if err != nil {
		a.setStatus(adapter.StatusErrored)
		return nil, err
	}

	a.mu.Lock()
	a.freshness["menu"] = &adapter.DataFreshness{
		DataType:    "menu",
		LastSyncAt:  time.Now(),
		RecordCount: int64(len(items)),
	}
	a.mu.Unlock()

	return items, nil
}

// --- adapter.EmployeeReader ---

// ReadEmployees fetches and maps all Loyverse employees.
func (a *LoyverseAdapter) ReadEmployees(ctx context.Context) ([]adapter.NormalizedEmployee, error) {
	a.mu.RLock()
	if a.status != adapter.StatusActive {
		a.mu.RUnlock()
		return nil, fmt.Errorf("loyverse adapter not active")
	}
	syncer := a.syncer
	a.mu.RUnlock()

	employees, err := syncer.SyncEmployees(ctx)
	if err != nil {
		a.setStatus(adapter.StatusErrored)
		return nil, err
	}

	a.mu.Lock()
	a.freshness["employees"] = &adapter.DataFreshness{
		DataType:    "employees",
		LastSyncAt:  time.Now(),
		RecordCount: int64(len(employees)),
	}
	a.mu.Unlock()

	return employees, nil
}

// --- adapter.StatusWriter ---

// Write86Status sets a Loyverse item variant's stock to 0 (86'd) or restores
// it to a nominal level (1) when available=true.
//
// The itemID must be a Loyverse variant_id. The storeID is taken from the
// adapter's configured LOYVERSE_STORE_ID.
func (a *LoyverseAdapter) Write86Status(ctx context.Context, itemID string, available bool) error {
	a.mu.RLock()
	if a.status != adapter.StatusActive {
		a.mu.RUnlock()
		return fmt.Errorf("loyverse adapter not active")
	}
	client := a.client
	storeID := a.storeID
	a.mu.RUnlock()

	inStock := float64(0)
	if available {
		inStock = 1 // restore to a nominal positive stock level
	}

	if err := client.UpdateVariantStock(itemID, storeID, inStock); err != nil {
		return fmt.Errorf("loyverse Write86Status: %w", err)
	}

	slog.Info("loyverse 86 status updated",
		"variant_id", itemID,
		"available", available,
		"store_id", storeID,
	)
	return nil
}

// --- helpers ---

func (a *LoyverseAdapter) setStatus(s adapter.Status) {
	a.mu.Lock()
	a.status = s
	a.mu.Unlock()
}

// credOrEnv returns the value from the credentials map under `key`, or falls
// back to reading the environment variable `envVar`.
func credOrEnv(creds map[string]string, key, envVar string) string {
	if v, ok := creds[key]; ok && v != "" {
		return v
	}
	return os.Getenv(envVar)
}

// parsePollInterval reads LOYVERSE_POLL_INTERVAL (minutes) from the environment.
func parsePollInterval() time.Duration {
	raw := os.Getenv("LOYVERSE_POLL_INTERVAL")
	if raw == "" {
		return 5 * time.Minute
	}
	mins, err := strconv.Atoi(raw)
	if err != nil || mins <= 0 {
		return 5 * time.Minute
	}
	return time.Duration(mins) * time.Minute
}
