package messaging

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/opsnerve/fireline/internal/event"
	"github.com/opsnerve/fireline/internal/tenant"
	"github.com/opsnerve/fireline/pkg/database"
)

// Service provides staff messaging capabilities.
type Service struct {
	pool *pgxpool.Pool
	bus  *event.Bus
}

// New creates a new messaging service.
func New(pool *pgxpool.Pool, bus *event.Bus) *Service {
	return &Service{pool: pool, bus: bus}
}

// ─── Structs ────────────────────────────────────────────────────────────────

// Channel represents a chat channel.
type Channel struct {
	ChannelID  string  `json:"channel_id"`
	OrgID      string  `json:"org_id"`
	LocationID *string `json:"location_id"`
	Name       string  `json:"name"`
	Type       string  `json:"type"`
	CreatedAt  string  `json:"created_at"`
}

// Message represents a chat message.
type Message struct {
	MessageID  string `json:"message_id"`
	OrgID      string `json:"org_id"`
	ChannelID  string `json:"channel_id"`
	SenderID   string `json:"sender_id"`
	SenderName string `json:"sender_name"`
	SenderRole string `json:"sender_role"`
	Body       string `json:"body"`
	Pinned     bool   `json:"pinned"`
	CreatedAt  string `json:"created_at"`
}

// MessageInput is the input for sending a message.
type MessageInput struct {
	ChannelID  string `json:"channel_id"`
	SenderID   string `json:"sender_id"`
	SenderName string `json:"sender_name"`
	SenderRole string `json:"sender_role"`
	Body       string `json:"body"`
}

// ─── Channels ───────────────────────────────────────────────────────────────

// ListChannels returns channels for a location plus broadcast channels.
func (s *Service) ListChannels(ctx context.Context, orgID, locationID string) ([]Channel, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var results []Channel

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		query := `SELECT channel_id, org_id, location_id, name, type, created_at::TEXT
			FROM chat_channels
			WHERE (type = 'broadcast'`
		args := []any{}
		argIdx := 1

		if locationID != "" {
			query += fmt.Sprintf(" OR location_id = $%d", argIdx)
			args = append(args, locationID)
			argIdx++
		}
		query += `) ORDER BY type DESC, name`

		rows, err := tx.Query(tenantCtx, query, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var ch Channel
			if err := rows.Scan(
				&ch.ChannelID, &ch.OrgID, &ch.LocationID, &ch.Name, &ch.Type, &ch.CreatedAt,
			); err != nil {
				return err
			}
			results = append(results, ch)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("list channels: %w", err)
	}
	return results, nil
}

// ─── Messages ───────────────────────────────────────────────────────────────

// ListMessages returns messages for a channel, newest first.
func (s *Service) ListMessages(ctx context.Context, orgID, channelID string, limit int) ([]Message, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var results []Message

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(tenantCtx,
			`SELECT message_id, org_id, channel_id, sender_id, sender_name, sender_role,
				body, pinned, created_at::TEXT
			FROM chat_messages
			WHERE channel_id = $1
			ORDER BY created_at DESC
			LIMIT $2`,
			channelID, limit,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var m Message
			if err := rows.Scan(
				&m.MessageID, &m.OrgID, &m.ChannelID, &m.SenderID, &m.SenderName, &m.SenderRole,
				&m.Body, &m.Pinned, &m.CreatedAt,
			); err != nil {
				return err
			}
			results = append(results, m)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}
	return results, nil
}

// SendMessage inserts a message and publishes an event.
func (s *Service) SendMessage(ctx context.Context, orgID string, input MessageInput) (*Message, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var m Message

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		return tx.QueryRow(tenantCtx,
			`INSERT INTO chat_messages (org_id, channel_id, sender_id, sender_name, sender_role, body)
			 VALUES ($1, $2, $3, $4, $5, $6)
			 RETURNING message_id, org_id, channel_id, sender_id, sender_name, sender_role,
				body, pinned, created_at::TEXT`,
			orgID, input.ChannelID, input.SenderID, input.SenderName, input.SenderRole, input.Body,
		).Scan(
			&m.MessageID, &m.OrgID, &m.ChannelID, &m.SenderID, &m.SenderName, &m.SenderRole,
			&m.Body, &m.Pinned, &m.CreatedAt,
		)
	})
	if err != nil {
		return nil, fmt.Errorf("send message: %w", err)
	}

	// Publish event
	s.bus.Publish(ctx, event.Envelope{
		EventType: "messaging.message.sent",
		OrgID:     orgID,
		Source:    "messaging",
		Payload:   m,
	})

	return &m, nil
}

// PinMessage toggles the pinned status of a message.
func (s *Service) PinMessage(ctx context.Context, orgID, messageID string, pinned bool) error {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		tag, err := tx.Exec(tenantCtx,
			`UPDATE chat_messages SET pinned = $1 WHERE message_id = $2`,
			pinned, messageID,
		)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return fmt.Errorf("message not found")
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("pin message: %w", err)
	}
	return nil
}
