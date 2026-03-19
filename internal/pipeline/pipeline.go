package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/opsnerve/fireline/internal/adapter"
	"github.com/opsnerve/fireline/internal/event"
	"github.com/opsnerve/fireline/internal/tenant"
	"github.com/opsnerve/fireline/pkg/database"
)

// Pipeline processes normalized data from adapters and writes to domain tables.
type Pipeline struct {
	pool *pgxpool.Pool
	bus  *event.Bus
}

// New creates a new data pipeline.
func New(pool *pgxpool.Pool, bus *event.Bus) *Pipeline {
	return &Pipeline{pool: pool, bus: bus}
}

// RegisterHandlers subscribes to adapter events on the event bus.
func (p *Pipeline) RegisterHandlers() {
	p.bus.Subscribe("adapter.orders.synced", p.handleOrdersSync)
	p.bus.Subscribe("adapter.menu.synced", p.handleMenuSync)
}

// handleOrdersSync processes a batch of normalized orders.
func (p *Pipeline) handleOrdersSync(ctx context.Context, env event.Envelope) error {
	orders, ok := env.Payload.([]adapter.NormalizedOrder)
	if !ok {
		return fmt.Errorf("invalid payload type for orders sync")
	}

	tenantCtx := tenant.WithOrgID(ctx, env.OrgID)
	var processed int

	for _, order := range orders {
		if err := p.upsertCheck(tenantCtx, order); err != nil {
			slog.Error("failed to upsert check", "external_id", order.ExternalID, "error", err)
			continue
		}
		processed++
	}

	slog.Info("orders pipeline processed",
		"org_id", env.OrgID,
		"total", len(orders),
		"processed", processed,
	)

	// Publish downstream event for intelligence modules
	p.bus.Publish(ctx, event.Envelope{
		EventID:   env.EventID + ".processed",
		EventType: "pipeline.orders.processed",
		OrgID:     env.OrgID,
		LocationID: env.LocationID,
		Source:    "pipeline",
		Payload: map[string]int{
			"total":     len(orders),
			"processed": processed,
		},
	})

	return nil
}

// upsertCheck inserts or updates a check and its items from a normalized order.
func (p *Pipeline) upsertCheck(ctx context.Context, order adapter.NormalizedOrder) error {
	return database.TenantTx(ctx, p.pool, func(tx pgx.Tx) error {
		var checkID string
		// First try to find existing check
		var checkExists bool
		err := tx.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM checks WHERE location_id = $1 AND external_id = $2)`,
			order.LocationID, order.ExternalID,
		).Scan(&checkExists)
		if err != nil {
			return fmt.Errorf("check existence: %w", err)
		}

		if checkExists {
			err = tx.QueryRow(ctx,
				`UPDATE checks SET status = $3, subtotal = $4, tax = $5, total = $6, tip = $7,
				                   discount = $8, closed_at = $9
				 WHERE location_id = $1 AND external_id = $2
				 RETURNING check_id`,
				order.LocationID, order.ExternalID,
				order.Status, order.Subtotal, order.Tax, order.Total, order.Tip,
				order.Discount, order.ClosedAt,
			).Scan(&checkID)
		} else {
			err = tx.QueryRow(ctx,
				`INSERT INTO checks (org_id, location_id, external_id, order_number, status, channel,
				                     subtotal, tax, total, tip, discount, opened_at, closed_at, source)
				 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
				 RETURNING check_id`,
				order.OrgID, order.LocationID, order.ExternalID, order.OrderNumber,
				order.Status, order.Channel, order.Subtotal, order.Tax, order.Total,
				order.Tip, order.Discount, order.OpenedAt, order.ClosedAt, order.Source,
			).Scan(&checkID)
		}
		if err != nil {
			return fmt.Errorf("upsert check: %w", err)
		}

		for _, item := range order.Items {
			_, err := tx.Exec(ctx,
				`INSERT INTO check_items (org_id, check_id, external_id, name, quantity, unit_price,
				                          voided_at, void_reason, fired_at)
				 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
				 ON CONFLICT DO NOTHING`,
				order.OrgID, checkID, item.ExternalID, item.Name, item.Quantity,
				item.UnitPrice, item.VoidedAt, item.VoidReason, item.FiredAt,
			)
			if err != nil {
				return fmt.Errorf("insert check item: %w", err)
			}
		}

		return nil
	})
}

// handleMenuSync processes a batch of normalized menu items.
func (p *Pipeline) handleMenuSync(ctx context.Context, env event.Envelope) error {
	items, ok := env.Payload.([]adapter.NormalizedMenuItem)
	if !ok {
		return fmt.Errorf("invalid payload type for menu sync")
	}

	tenantCtx := tenant.WithOrgID(ctx, env.OrgID)
	var processed int

	for _, item := range items {
		if err := p.upsertMenuItem(tenantCtx, item); err != nil {
			slog.Error("failed to upsert menu item", "external_id", item.ExternalID, "error", err)
			continue
		}
		processed++
	}

	slog.Info("menu pipeline processed",
		"org_id", env.OrgID,
		"total", len(items),
		"processed", processed,
	)

	p.bus.Publish(ctx, event.Envelope{
		EventID:   env.EventID + ".processed",
		EventType: "pipeline.menu.processed",
		OrgID:     env.OrgID,
		LocationID: env.LocationID,
		Source:    "pipeline",
		Payload: map[string]int{
			"total":     len(items),
			"processed": processed,
		},
	})

	return nil
}

// upsertMenuItem inserts or updates a menu item from normalized data.
func (p *Pipeline) upsertMenuItem(ctx context.Context, item adapter.NormalizedMenuItem) error {
	return database.TenantTx(ctx, p.pool, func(tx pgx.Tx) error {
		var exists bool
		err := tx.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM menu_items WHERE location_id = $1 AND external_id = $2)`,
			item.LocationID, item.ExternalID,
		).Scan(&exists)
		if err != nil {
			return err
		}
		if exists {
			_, err = tx.Exec(ctx,
				`UPDATE menu_items SET name = $3, category = $4, price = $5, available = $6, updated_at = now()
				 WHERE location_id = $1 AND external_id = $2`,
				item.LocationID, item.ExternalID,
				item.Name, item.Category, item.Price, item.Available,
			)
		} else {
			_, err = tx.Exec(ctx,
				`INSERT INTO menu_items (org_id, location_id, external_id, name, category, price, available, source)
				 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
				item.OrgID, item.LocationID, item.ExternalID, item.Name,
				item.Category, item.Price, item.Available, item.Source,
			)
		}
		return err
	})
}

// SyncOrders runs a full order sync from an adapter through the pipeline.
func (p *Pipeline) SyncOrders(ctx context.Context, a adapter.Adapter, reader adapter.OrderReader, cfg adapter.Config) error {
	orders, err := reader.ReadOrders(ctx, cfg.CreatedAt, 100)
	if err != nil {
		return fmt.Errorf("read orders: %w", err)
	}

	payloadJSON, _ := json.Marshal(orders)
	_ = payloadJSON // raw log would go here in production

	p.bus.Publish(ctx, event.Envelope{
		EventID:    fmt.Sprintf("sync-orders-%s", cfg.LocationID),
		EventType:  "adapter.orders.synced",
		OrgID:      cfg.OrgID,
		LocationID: cfg.LocationID,
		Source:     a.Type(),
		Payload:    orders,
	})

	return nil
}

// SyncMenu runs a full menu sync from an adapter through the pipeline.
func (p *Pipeline) SyncMenu(ctx context.Context, a adapter.Adapter, reader adapter.MenuReader, cfg adapter.Config) error {
	items, err := reader.ReadMenu(ctx)
	if err != nil {
		return fmt.Errorf("read menu: %w", err)
	}

	p.bus.Publish(ctx, event.Envelope{
		EventID:    fmt.Sprintf("sync-menu-%s", cfg.LocationID),
		EventType:  "adapter.menu.synced",
		OrgID:      cfg.OrgID,
		LocationID: cfg.LocationID,
		Source:     a.Type(),
		Payload:    items,
	})

	return nil
}
