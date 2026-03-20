package operations

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/opsnerve/fireline/internal/event"
	"github.com/opsnerve/fireline/internal/tenant"
	"github.com/opsnerve/fireline/pkg/database"
)

// KDSTicket represents a kitchen display system ticket.
type KDSTicket struct {
	TicketID         string          `json:"ticket_id"`
	OrgID            string          `json:"org_id"`
	LocationID       string          `json:"location_id"`
	CheckID          *string         `json:"check_id,omitempty"`
	OrderNumber      string          `json:"order_number"`
	Channel          string          `json:"channel"`
	Status           string          `json:"status"`
	Priority         int             `json:"priority"`
	EstimatedReadyAt *time.Time      `json:"estimated_ready_at,omitempty"`
	ActualReadyAt    *time.Time      `json:"actual_ready_at,omitempty"`
	Items            []KDSTicketItem `json:"items"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

// KDSTicketItem represents a single item line on a KDS ticket.
type KDSTicketItem struct {
	TicketItemID string     `json:"ticket_item_id"`
	OrgID        string     `json:"org_id"`
	TicketID     string     `json:"ticket_id"`
	MenuItemID   string     `json:"menu_item_id"`
	ItemName     string     `json:"item_name"`
	Quantity     int        `json:"quantity"`
	StationType  string     `json:"station_type"`
	Status       string     `json:"status"`
	FireAt       *time.Time `json:"fire_at,omitempty"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	DurationSecs *int       `json:"duration_secs,omitempty"`
	Notes        *string    `json:"notes,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

// KDSMetrics holds aggregated KDS performance data.
type KDSMetrics struct {
	LocationID        string           `json:"location_id"`
	From              time.Time        `json:"from"`
	To                time.Time        `json:"to"`
	AvgTicketTimeSecs float64          `json:"avg_ticket_time_secs"`
	ItemsCompleted    int              `json:"items_completed"`
	PerStation        []StationMetrics `json:"per_station"`
}

// StationMetrics holds per-station KDS performance.
type StationMetrics struct {
	StationType      string  `json:"station_type"`
	ItemsCompleted   int     `json:"items_completed"`
	AvgBumpTimeSecs  float64 `json:"avg_bump_time_secs"`
}

// CreateTicketFromCheck creates a KDS ticket from an existing check.
func (s *Service) CreateTicketFromCheck(ctx context.Context, orgID, locationID, checkID, orderNumber, channel string) (*KDSTicket, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	var ticket KDSTicket

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		// Query check items.
		type checkItem struct {
			menuItemID string
			itemName   string
			quantity   int
			notes      *string
		}

		rows, err := tx.Query(tenantCtx,
			`SELECT ci.menu_item_id, mi.name, ci.quantity, ci.notes
			 FROM check_items ci
			 JOIN menu_items mi ON mi.menu_item_id = ci.menu_item_id
			 WHERE ci.check_id = $1`,
			checkID,
		)
		if err != nil {
			return fmt.Errorf("query check items: %w", err)
		}

		var items []checkItem
		for rows.Next() {
			var ci checkItem
			if err := rows.Scan(&ci.menuItemID, &ci.itemName, &ci.quantity, &ci.notes); err != nil {
				rows.Close()
				return fmt.Errorf("scan check item: %w", err)
			}
			items = append(items, ci)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return fmt.Errorf("iterate check items: %w", err)
		}

		// Create the ticket.
		now := time.Now().UTC()
		var ticketID string
		if err := tx.QueryRow(tenantCtx,
			`INSERT INTO kds_tickets (org_id, location_id, check_id, order_number, channel, status, priority)
			 VALUES ($1, $2, $3, $4, $5, 'new', 0)
			 RETURNING ticket_id, org_id, location_id, check_id, order_number, channel, status, priority,
			           estimated_ready_at, actual_ready_at, created_at, updated_at`,
			orgID, locationID, checkID, orderNumber, channel,
		).Scan(
			&ticketID, &ticket.OrgID, &ticket.LocationID, &ticket.CheckID,
			&ticket.OrderNumber, &ticket.Channel, &ticket.Status, &ticket.Priority,
			&ticket.EstimatedReadyAt, &ticket.ActualReadyAt, &ticket.CreatedAt, &ticket.UpdatedAt,
		); err != nil {
			return fmt.Errorf("insert ticket: %w", err)
		}
		ticket.TicketID = ticketID

		defaults := defaultStationTimes()
		maxDuration := 0

		// For each check item, look up resource profiles and insert ticket items.
		for _, ci := range items {
			profiles, err := s.getResourceProfilesTx(tenantCtx, tx, orgID, ci.menuItemID)
			if err != nil {
				return fmt.Errorf("get profiles for %s: %w", ci.menuItemID, err)
			}

			if len(profiles) == 0 {
				// Fallback: create a single prep item using defaults.
				stType := "prep"
				dur := defaults[stType]
				if dur > maxDuration {
					maxDuration = dur
				}
				var itemID string
				if err := tx.QueryRow(tenantCtx,
					`INSERT INTO kds_ticket_items
					 (org_id, ticket_id, menu_item_id, item_name, quantity, station_type, status, notes)
					 VALUES ($1, $2, $3, $4, $5, $6, 'pending', $7)
					 RETURNING ticket_item_id`,
					orgID, ticketID, ci.menuItemID, ci.itemName, ci.quantity, stType, ci.notes,
				).Scan(&itemID); err != nil {
					return fmt.Errorf("insert fallback ticket item: %w", err)
				}
				ticket.Items = append(ticket.Items, KDSTicketItem{
					TicketItemID: itemID,
					OrgID:        orgID,
					TicketID:     ticketID,
					MenuItemID:   ci.menuItemID,
					ItemName:     ci.itemName,
					Quantity:     ci.quantity,
					StationType:  stType,
					Status:       "pending",
					Notes:        ci.notes,
					CreatedAt:    now,
				})
				continue
			}

			for _, p := range profiles {
				if p.DurationSecs > maxDuration {
					maxDuration = p.DurationSecs
				}
				var itemID string
				if err := tx.QueryRow(tenantCtx,
					`INSERT INTO kds_ticket_items
					 (org_id, ticket_id, menu_item_id, item_name, quantity, station_type, status, notes)
					 VALUES ($1, $2, $3, $4, $5, $6, 'pending', $7)
					 RETURNING ticket_item_id`,
					orgID, ticketID, ci.menuItemID, ci.itemName, ci.quantity, p.StationType, ci.notes,
				).Scan(&itemID); err != nil {
					return fmt.Errorf("insert ticket item: %w", err)
				}
				ticket.Items = append(ticket.Items, KDSTicketItem{
					TicketItemID: itemID,
					OrgID:        orgID,
					TicketID:     ticketID,
					MenuItemID:   ci.menuItemID,
					ItemName:     ci.itemName,
					Quantity:     ci.quantity,
					StationType:  p.StationType,
					Status:       "pending",
					Notes:        ci.notes,
					CreatedAt:    now,
				})
			}
		}

		// Set estimated_ready_at.
		estimatedReady := now.Add(time.Duration(maxDuration) * time.Second)
		ticket.EstimatedReadyAt = &estimatedReady
		if _, err := tx.Exec(tenantCtx,
			`UPDATE kds_tickets SET estimated_ready_at = $1 WHERE ticket_id = $2`,
			estimatedReady, ticketID,
		); err != nil {
			return fmt.Errorf("update estimated_ready_at: %w", err)
		}

		if ticket.Items == nil {
			ticket.Items = []KDSTicketItem{}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Emit event.
	s.bus.Publish(ctx, event.Envelope{
		EventType:  "operations.ticket.created",
		OrgID:      orgID,
		LocationID: locationID,
		Source:     "operations",
		Payload:    map[string]string{"ticket_id": ticket.TicketID, "check_id": checkID},
	})

	return &ticket, nil
}

// getResourceProfilesTx fetches resource profiles within an existing transaction.
func (s *Service) getResourceProfilesTx(ctx context.Context, tx pgx.Tx, orgID, menuItemID string) ([]ResourceProfile, error) {
	rows, err := tx.Query(ctx,
		`SELECT profile_id, org_id, menu_item_id, station_type, task_sequence,
		        duration_secs, elu_required, batch_size, created_at
		 FROM menu_item_resource_profiles
		 WHERE org_id = $1 AND menu_item_id = $2
		 ORDER BY task_sequence`,
		orgID, menuItemID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []ResourceProfile
	for rows.Next() {
		var p ResourceProfile
		if err := rows.Scan(
			&p.ProfileID, &p.OrgID, &p.MenuItemID, &p.StationType, &p.TaskSequence,
			&p.DurationSecs, &p.ELURequired, &p.BatchSize, &p.CreatedAt,
		); err != nil {
			return nil, err
		}
		profiles = append(profiles, p)
	}
	return profiles, rows.Err()
}

// GetStationTickets returns active tickets for a specific station.
func (s *Service) GetStationTickets(ctx context.Context, orgID, locationID, stationType string) ([]KDSTicket, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	var tickets []KDSTicket

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(tenantCtx,
			`SELECT DISTINCT kt.ticket_id, kt.org_id, kt.location_id, kt.check_id,
			        kt.order_number, kt.channel, kt.status, kt.priority,
			        kt.estimated_ready_at, kt.actual_ready_at, kt.created_at, kt.updated_at
			 FROM kds_tickets kt
			 JOIN kds_ticket_items kti ON kti.ticket_id = kt.ticket_id
			 WHERE kt.org_id = $1 AND kt.location_id = $2
			   AND kti.station_type = $3
			   AND kt.status NOT IN ('ready', 'delivered', 'cancelled')
			 ORDER BY kt.priority DESC, kt.created_at ASC`,
			orgID, locationID, stationType,
		)
		if err != nil {
			return fmt.Errorf("query station tickets: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var t KDSTicket
			if err := rows.Scan(
				&t.TicketID, &t.OrgID, &t.LocationID, &t.CheckID,
				&t.OrderNumber, &t.Channel, &t.Status, &t.Priority,
				&t.EstimatedReadyAt, &t.ActualReadyAt, &t.CreatedAt, &t.UpdatedAt,
			); err != nil {
				return fmt.Errorf("scan ticket: %w", err)
			}
			tickets = append(tickets, t)
		}
		if err := rows.Err(); err != nil {
			return err
		}

		// Load items for each ticket filtered to this station.
		for i := range tickets {
			items, err := s.loadTicketItemsByStation(tenantCtx, tx, orgID, tickets[i].TicketID, stationType)
			if err != nil {
				return fmt.Errorf("load items for ticket %s: %w", tickets[i].TicketID, err)
			}
			tickets[i].Items = items
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if tickets == nil {
		tickets = []KDSTicket{}
	}
	return tickets, nil
}

// GetAllTickets returns all active tickets with all items (expo view).
func (s *Service) GetAllTickets(ctx context.Context, orgID, locationID string) ([]KDSTicket, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	var tickets []KDSTicket

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(tenantCtx,
			`SELECT ticket_id, org_id, location_id, check_id, order_number, channel,
			        status, priority, estimated_ready_at, actual_ready_at, created_at, updated_at
			 FROM kds_tickets
			 WHERE org_id = $1 AND location_id = $2
			   AND status NOT IN ('delivered', 'cancelled')
			 ORDER BY priority DESC, created_at ASC`,
			orgID, locationID,
		)
		if err != nil {
			return fmt.Errorf("query all tickets: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var t KDSTicket
			if err := rows.Scan(
				&t.TicketID, &t.OrgID, &t.LocationID, &t.CheckID,
				&t.OrderNumber, &t.Channel, &t.Status, &t.Priority,
				&t.EstimatedReadyAt, &t.ActualReadyAt, &t.CreatedAt, &t.UpdatedAt,
			); err != nil {
				return fmt.Errorf("scan ticket: %w", err)
			}
			tickets = append(tickets, t)
		}
		if err := rows.Err(); err != nil {
			return err
		}

		for i := range tickets {
			items, err := s.loadAllTicketItems(tenantCtx, tx, orgID, tickets[i].TicketID)
			if err != nil {
				return fmt.Errorf("load items for ticket %s: %w", tickets[i].TicketID, err)
			}
			tickets[i].Items = items
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if tickets == nil {
		tickets = []KDSTicket{}
	}
	return tickets, nil
}

// BumpTicketItem advances a ticket item's status.
func (s *Service) BumpTicketItem(ctx context.Context, orgID, ticketItemID, newStatus string) (*KDSTicketItem, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	var item KDSTicketItem
	var ticketID string

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		now := time.Now().UTC()

		var query string
		switch newStatus {
		case "cooking":
			query = `UPDATE kds_ticket_items
			          SET status = $1, started_at = $2
			          WHERE org_id = $3 AND ticket_item_id = $4
			          RETURNING ticket_item_id, org_id, ticket_id, menu_item_id, item_name,
			                    quantity, station_type, status, fire_at, started_at,
			                    completed_at, duration_secs, notes, created_at`
			if err := tx.QueryRow(tenantCtx, query,
				newStatus, now, orgID, ticketItemID,
			).Scan(
				&item.TicketItemID, &item.OrgID, &ticketID, &item.MenuItemID, &item.ItemName,
				&item.Quantity, &item.StationType, &item.Status, &item.FireAt, &item.StartedAt,
				&item.CompletedAt, &item.DurationSecs, &item.Notes, &item.CreatedAt,
			); err != nil {
				return fmt.Errorf("bump item to cooking: %w", err)
			}
		case "ready":
			// Fetch started_at to compute duration.
			var startedAt *time.Time
			if err := tx.QueryRow(tenantCtx,
				`SELECT started_at FROM kds_ticket_items WHERE org_id = $1 AND ticket_item_id = $2`,
				orgID, ticketItemID,
			).Scan(&startedAt); err != nil {
				return fmt.Errorf("fetch started_at: %w", err)
			}

			var durSecs *int
			if startedAt != nil {
				d := int(now.Sub(*startedAt).Seconds())
				durSecs = &d
			}

			query = `UPDATE kds_ticket_items
			          SET status = $1, completed_at = $2, duration_secs = $3
			          WHERE org_id = $4 AND ticket_item_id = $5
			          RETURNING ticket_item_id, org_id, ticket_id, menu_item_id, item_name,
			                    quantity, station_type, status, fire_at, started_at,
			                    completed_at, duration_secs, notes, created_at`
			if err := tx.QueryRow(tenantCtx, query,
				newStatus, now, durSecs, orgID, ticketItemID,
			).Scan(
				&item.TicketItemID, &item.OrgID, &ticketID, &item.MenuItemID, &item.ItemName,
				&item.Quantity, &item.StationType, &item.Status, &item.FireAt, &item.StartedAt,
				&item.CompletedAt, &item.DurationSecs, &item.Notes, &item.CreatedAt,
			); err != nil {
				return fmt.Errorf("bump item to ready: %w", err)
			}
		default:
			query = `UPDATE kds_ticket_items
			          SET status = $1
			          WHERE org_id = $2 AND ticket_item_id = $3
			          RETURNING ticket_item_id, org_id, ticket_id, menu_item_id, item_name,
			                    quantity, station_type, status, fire_at, started_at,
			                    completed_at, duration_secs, notes, created_at`
			if err := tx.QueryRow(tenantCtx, query,
				newStatus, orgID, ticketItemID,
			).Scan(
				&item.TicketItemID, &item.OrgID, &ticketID, &item.MenuItemID, &item.ItemName,
				&item.Quantity, &item.StationType, &item.Status, &item.FireAt, &item.StartedAt,
				&item.CompletedAt, &item.DurationSecs, &item.Notes, &item.CreatedAt,
			); err != nil {
				return fmt.Errorf("bump item to %s: %w", newStatus, err)
			}
		}

		item.TicketID = ticketID

		// Check if all non-cancelled items for this ticket are ready.
		if newStatus == "ready" {
			var pendingCount int
			if err := tx.QueryRow(tenantCtx,
				`SELECT COUNT(*) FROM kds_ticket_items
				 WHERE ticket_id = $1 AND status NOT IN ('ready', 'cancelled')`,
				ticketID,
			).Scan(&pendingCount); err != nil {
				return fmt.Errorf("check pending items: %w", err)
			}

			if pendingCount == 0 {
				actualReady := now
				if _, err := tx.Exec(tenantCtx,
					`UPDATE kds_tickets
					 SET status = 'ready', actual_ready_at = $1, updated_at = $2
					 WHERE ticket_id = $3`,
					actualReady, now, ticketID,
				); err != nil {
					return fmt.Errorf("update ticket to ready: %w", err)
				}
			}
		}

		// Update ticket updated_at.
		if _, err := tx.Exec(tenantCtx,
			`UPDATE kds_tickets SET updated_at = $1 WHERE ticket_id = $2`,
			now, ticketID,
		); err != nil {
			return fmt.Errorf("update ticket updated_at: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Emit event.
	s.bus.Publish(ctx, event.Envelope{
		EventType: "operations.ticket.item.bumped",
		OrgID:     orgID,
		Source:    "operations",
		Payload: map[string]string{
			"ticket_item_id": ticketItemID,
			"ticket_id":      ticketID,
			"new_status":     newStatus,
		},
	})

	return &item, nil
}

// CancelTicket cancels a ticket and all its items.
func (s *Service) CancelTicket(ctx context.Context, orgID, ticketID string) error {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	return database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		now := time.Now().UTC()

		if _, err := tx.Exec(tenantCtx,
			`UPDATE kds_ticket_items SET status = 'cancelled'
			 WHERE org_id = $1 AND ticket_id = $2`,
			orgID, ticketID,
		); err != nil {
			return fmt.Errorf("cancel ticket items: %w", err)
		}

		if _, err := tx.Exec(tenantCtx,
			`UPDATE kds_tickets SET status = 'cancelled', updated_at = $1
			 WHERE org_id = $2 AND ticket_id = $3`,
			now, orgID, ticketID,
		); err != nil {
			return fmt.Errorf("cancel ticket: %w", err)
		}

		return nil
	})
}

// GetKDSMetrics returns aggregated KDS performance metrics for a time range.
func (s *Service) GetKDSMetrics(ctx context.Context, orgID, locationID string, from, to time.Time) (*KDSMetrics, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	metrics := &KDSMetrics{
		LocationID: locationID,
		From:       from,
		To:         to,
	}

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		// Average ticket time (actual_ready_at - created_at in seconds).
		if err := tx.QueryRow(tenantCtx,
			`SELECT
			    COALESCE(AVG(EXTRACT(EPOCH FROM (actual_ready_at - created_at))), 0)
			 FROM kds_tickets
			 WHERE org_id = $1 AND location_id = $2
			   AND status = 'ready'
			   AND actual_ready_at IS NOT NULL
			   AND created_at >= $3 AND created_at < $4`,
			orgID, locationID, from, to,
		).Scan(&metrics.AvgTicketTimeSecs); err != nil {
			return fmt.Errorf("avg ticket time: %w", err)
		}

		// Total items completed.
		if err := tx.QueryRow(tenantCtx,
			`SELECT COUNT(*)
			 FROM kds_ticket_items kti
			 JOIN kds_tickets kt ON kt.ticket_id = kti.ticket_id
			 WHERE kti.org_id = $1 AND kt.location_id = $2
			   AND kti.status = 'ready'
			   AND kti.completed_at >= $3 AND kti.completed_at < $4`,
			orgID, locationID, from, to,
		).Scan(&metrics.ItemsCompleted); err != nil {
			return fmt.Errorf("items completed: %w", err)
		}

		// Per-station metrics.
		rows, err := tx.Query(tenantCtx,
			`SELECT kti.station_type,
			        COUNT(*) AS items_completed,
			        COALESCE(AVG(kti.duration_secs), 0) AS avg_bump_time_secs
			 FROM kds_ticket_items kti
			 JOIN kds_tickets kt ON kt.ticket_id = kti.ticket_id
			 WHERE kti.org_id = $1 AND kt.location_id = $2
			   AND kti.status = 'ready'
			   AND kti.completed_at >= $3 AND kti.completed_at < $4
			 GROUP BY kti.station_type
			 ORDER BY kti.station_type`,
			orgID, locationID, from, to,
		)
		if err != nil {
			return fmt.Errorf("per-station metrics: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var sm StationMetrics
			if err := rows.Scan(&sm.StationType, &sm.ItemsCompleted, &sm.AvgBumpTimeSecs); err != nil {
				return fmt.Errorf("scan station metrics: %w", err)
			}
			metrics.PerStation = append(metrics.PerStation, sm)
		}
		if err := rows.Err(); err != nil {
			return err
		}

		if metrics.PerStation == nil {
			metrics.PerStation = []StationMetrics{}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return metrics, nil
}

// loadTicketItemsByStation loads ticket items for a specific station.
func (s *Service) loadTicketItemsByStation(ctx context.Context, tx pgx.Tx, orgID, ticketID, stationType string) ([]KDSTicketItem, error) {
	rows, err := tx.Query(ctx,
		`SELECT ticket_item_id, org_id, ticket_id, menu_item_id, item_name,
		        quantity, station_type, status, fire_at, started_at,
		        completed_at, duration_secs, notes, created_at
		 FROM kds_ticket_items
		 WHERE org_id = $1 AND ticket_id = $2 AND station_type = $3
		 ORDER BY created_at`,
		orgID, ticketID, stationType,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []KDSTicketItem
	for rows.Next() {
		var item KDSTicketItem
		if err := rows.Scan(
			&item.TicketItemID, &item.OrgID, &item.TicketID, &item.MenuItemID, &item.ItemName,
			&item.Quantity, &item.StationType, &item.Status, &item.FireAt, &item.StartedAt,
			&item.CompletedAt, &item.DurationSecs, &item.Notes, &item.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if items == nil {
		items = []KDSTicketItem{}
	}
	return items, rows.Err()
}

// loadAllTicketItems loads all items for a ticket.
func (s *Service) loadAllTicketItems(ctx context.Context, tx pgx.Tx, orgID, ticketID string) ([]KDSTicketItem, error) {
	rows, err := tx.Query(ctx,
		`SELECT ticket_item_id, org_id, ticket_id, menu_item_id, item_name,
		        quantity, station_type, status, fire_at, started_at,
		        completed_at, duration_secs, notes, created_at
		 FROM kds_ticket_items
		 WHERE org_id = $1 AND ticket_id = $2
		 ORDER BY station_type, created_at`,
		orgID, ticketID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []KDSTicketItem
	for rows.Next() {
		var item KDSTicketItem
		if err := rows.Scan(
			&item.TicketItemID, &item.OrgID, &item.TicketID, &item.MenuItemID, &item.ItemName,
			&item.Quantity, &item.StationType, &item.Status, &item.FireAt, &item.StartedAt,
			&item.CompletedAt, &item.DurationSecs, &item.Notes, &item.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if items == nil {
		items = []KDSTicketItem{}
	}
	return items, rows.Err()
}
