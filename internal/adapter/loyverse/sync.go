package loyverse

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/opsnerve/fireline/internal/adapter"
	"github.com/opsnerve/fireline/internal/event"
)

// Syncer orchestrates polling and sync operations for the Loyverse adapter.
type Syncer struct {
	client  *Client
	bus     *event.Bus
	cfg     adapter.Config
	storeID string
}

// newSyncer creates a Syncer using the adapter's client and config.
func newSyncer(client *Client, bus *event.Bus, cfg adapter.Config, storeID string) *Syncer {
	return &Syncer{
		client:  client,
		bus:     bus,
		cfg:     cfg,
		storeID: storeID,
	}
}

// SyncMenu fetches all Loyverse items and categories, maps them, and publishes
// a loyverse.menu.synced event on the bus.
func (s *Syncer) SyncMenu(ctx context.Context) ([]adapter.NormalizedMenuItem, error) {
	// Build category lookup map.
	catResp, err := s.client.GetCategories()
	if err != nil {
		return nil, fmt.Errorf("loyverse SyncMenu: fetch categories: %w", err)
	}
	categories := make(map[string]string, len(catResp.Categories))
	for _, c := range catResp.Categories {
		categories[c.ID] = c.Name
	}

	// Paginate through all items.
	var allItems []adapter.NormalizedMenuItem
	cursor := ""
	for {
		page, err := s.client.GetItems(cursor)
		if err != nil {
			return nil, fmt.Errorf("loyverse SyncMenu: fetch items (cursor=%q): %w", cursor, err)
		}
		for _, item := range page.Items {
			catName := categories[item.CategoryID]
			mapped := MapItem(item, catName)
			mapped.OrgID = s.cfg.OrgID
			mapped.LocationID = s.cfg.LocationID
			allItems = append(allItems, mapped)
		}
		if page.Cursor == "" {
			break
		}
		cursor = page.Cursor
	}

	slog.Info("loyverse: menu sync complete", "item_count", len(allItems), "location_id", s.cfg.LocationID)

	if s.bus != nil {
		s.bus.Publish(ctx, event.Envelope{
			EventType:  "adapter.menu.synced",
			OrgID:      s.cfg.OrgID,
			LocationID: s.cfg.LocationID,
			Source:     "loyverse",
			Payload:    allItems,
		})
	}

	return allItems, nil
}

// SyncOrders fetches receipts since `since`, maps them, and publishes events.
func (s *Syncer) SyncOrders(ctx context.Context, since time.Time) ([]adapter.NormalizedOrder, error) {
	var allOrders []adapter.NormalizedOrder
	cursor := ""
	for {
		page, err := s.client.GetReceipts(since, cursor)
		if err != nil {
			return nil, fmt.Errorf("loyverse SyncOrders: fetch receipts: %w", err)
		}
		for _, r := range page.Receipts {
			mapped := MapReceipt(r)
			mapped.OrgID = s.cfg.OrgID
			mapped.LocationID = s.cfg.LocationID
			allOrders = append(allOrders, mapped)
		}
		if page.Cursor == "" {
			break
		}
		cursor = page.Cursor
	}

	slog.Info("loyverse: orders sync complete", "order_count", len(allOrders), "since", since, "location_id", s.cfg.LocationID)

	if s.bus != nil {
		s.bus.Publish(ctx, event.Envelope{
			EventType:  "adapter.orders.synced",
			OrgID:      s.cfg.OrgID,
			LocationID: s.cfg.LocationID,
			Source:     "loyverse",
			Payload:    allOrders,
		})
	}

	return allOrders, nil
}

// SyncEmployees fetches all employees, maps them, and publishes an event.
func (s *Syncer) SyncEmployees(ctx context.Context) ([]adapter.NormalizedEmployee, error) {
	resp, err := s.client.GetEmployees()
	if err != nil {
		return nil, fmt.Errorf("loyverse SyncEmployees: %w", err)
	}

	employees := make([]adapter.NormalizedEmployee, 0, len(resp.Employees))
	for _, e := range resp.Employees {
		mapped := MapEmployee(e)
		mapped.OrgID = s.cfg.OrgID
		mapped.LocationID = s.cfg.LocationID
		employees = append(employees, mapped)
	}

	slog.Info("loyverse: employees sync complete", "employee_count", len(employees), "location_id", s.cfg.LocationID)

	if s.bus != nil {
		s.bus.Publish(ctx, event.Envelope{
			EventType:  "loyverse.employees.synced",
			OrgID:      s.cfg.OrgID,
			LocationID: s.cfg.LocationID,
			Source:     "loyverse",
			Payload: map[string]any{
				"employee_count": len(employees),
				"employees":      employees,
			},
		})
	}

	return employees, nil
}

// SyncInventory fetches inventory levels for the configured store and publishes an event.
func (s *Syncer) SyncInventory(ctx context.Context) ([]InventoryLevel, error) {
	resp, err := s.client.GetInventory(s.storeID)
	if err != nil {
		return nil, fmt.Errorf("loyverse SyncInventory: %w", err)
	}

	slog.Info("loyverse: inventory sync complete",
		"level_count", len(resp.InventoryLevels),
		"store_id", s.storeID,
		"location_id", s.cfg.LocationID,
	)

	if s.bus != nil {
		s.bus.Publish(ctx, event.Envelope{
			EventType:  "loyverse.inventory.synced",
			OrgID:      s.cfg.OrgID,
			LocationID: s.cfg.LocationID,
			Source:     "loyverse",
			Payload: map[string]any{
				"level_count": len(resp.InventoryLevels),
				"levels":      resp.InventoryLevels,
			},
		})
	}

	return resp.InventoryLevels, nil
}

// StartPolling launches a background goroutine that runs an immediate full sync
// followed by periodic syncs at the given interval.
// The goroutine exits when ctx is cancelled.
func (s *Syncer) StartPolling(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = 5 * time.Minute
	}
	go func() {
		// Run an immediate sync on connect so data is available right away.
		slog.Info("loyverse: running initial sync", "location_id", s.cfg.LocationID)
		s.runInitialSync(ctx)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		slog.Info("loyverse: polling started", "interval", interval, "location_id", s.cfg.LocationID)
		for {
			select {
			case <-ctx.Done():
				slog.Info("loyverse: polling stopped", "location_id", s.cfg.LocationID)
				return
			case t := <-ticker.C:
				slog.Info("loyverse: poll tick", "time", t, "location_id", s.cfg.LocationID)
				s.runFullSync(ctx)
			}
		}
	}()
}

// runInitialSync performs a first-time sync with 30 days of historical orders.
func (s *Syncer) runInitialSync(ctx context.Context) {
	since := time.Now().AddDate(0, 0, -30)

	if _, err := s.SyncMenu(ctx); err != nil {
		slog.Error("loyverse: initial menu sync failed", "error", err)
	}
	if _, err := s.SyncOrders(ctx, since); err != nil {
		slog.Error("loyverse: initial orders sync failed", "error", err)
	}
	if _, err := s.SyncEmployees(ctx); err != nil {
		slog.Error("loyverse: initial employees sync failed", "error", err)
	}
	if _, err := s.SyncInventory(ctx); err != nil {
		slog.Error("loyverse: initial inventory sync failed", "error", err)
	}
}

// runFullSync performs a complete sync of all data types.
func (s *Syncer) runFullSync(ctx context.Context) {
	since := time.Now().Truncate(24 * time.Hour) // pull all of today's orders

	if _, err := s.SyncMenu(ctx); err != nil {
		slog.Error("loyverse: menu sync failed", "error", err)
	}
	if _, err := s.SyncOrders(ctx, since); err != nil {
		slog.Error("loyverse: orders sync failed", "error", err)
	}
	if _, err := s.SyncEmployees(ctx); err != nil {
		slog.Error("loyverse: employees sync failed", "error", err)
	}
	if _, err := s.SyncInventory(ctx); err != nil {
		slog.Error("loyverse: inventory sync failed", "error", err)
	}
}
