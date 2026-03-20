package operations

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/opsnerve/fireline/internal/tenant"
	"github.com/opsnerve/fireline/pkg/database"
)

// TicketPriority holds the computed priority score for a KDS ticket.
type TicketPriority struct {
	TicketID       string  `json:"ticket_id"`
	OrderNumber    string  `json:"order_number"`
	Channel        string  `json:"channel"`
	PriorityScore  float64 `json:"priority_score"`
	SLAMinutes     int     `json:"sla_minutes"`
	ElapsedMinutes int     `json:"elapsed_minutes"`
	Urgency        string  `json:"urgency"` // "normal", "urgent", "critical"
}

// computePriority computes a weighted priority score in [0, 1].
// Weights: slaProximity 35%, customerValue 25%, channelWeight 20%, complexityInverse 20%.
func computePriority(slaProximity, customerValue, channelWeight, complexityInverse float64) float64 {
	return slaProximity*0.35 + customerValue*0.25 + channelWeight*0.20 + complexityInverse*0.20
}

// classifyUrgency returns an urgency label based on SLA elapsed fraction.
func classifyUrgency(elapsed, sla int) string {
	if sla <= 0 {
		return "normal"
	}
	fraction := float64(elapsed) / float64(sla)
	switch {
	case fraction >= 1.0:
		return "critical"
	case fraction >= 0.75:
		return "urgent"
	default:
		return "normal"
	}
}

// channelWeightFor maps a channel name to its priority weight.
func channelWeightFor(channel string) float64 {
	switch channel {
	case "dine-in", "dine_in":
		return 1.0
	case "takeout", "pickup":
		return 0.85
	case "delivery", "third_party":
		return 0.70
	default:
		return 0.75
	}
}

// GetTicketPriorities returns all active tickets with computed priority scores, sorted descending.
func (s *Service) GetTicketPriorities(ctx context.Context, orgID, locationID string) ([]TicketPriority, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	type rawTicket struct {
		ticketID        string
		orderNumber     string
		channel         string
		createdAt       time.Time
		estimatedReadyAt *time.Time
	}

	var rawTickets []rawTicket

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(tenantCtx,
			`SELECT ticket_id, order_number, channel, created_at, estimated_ready_at
			 FROM kds_tickets
			 WHERE org_id = $1 AND location_id = $2
			   AND status NOT IN ('ready', 'delivered', 'cancelled')
			 ORDER BY created_at ASC`,
			orgID, locationID,
		)
		if err != nil {
			return fmt.Errorf("query active tickets: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var t rawTicket
			if err := rows.Scan(&t.ticketID, &t.orderNumber, &t.channel, &t.createdAt, &t.estimatedReadyAt); err != nil {
				return fmt.Errorf("scan ticket: %w", err)
			}
			rawTickets = append(rawTickets, t)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}

	now := time.Now()
	priorities := make([]TicketPriority, 0, len(rawTickets))

	for _, rt := range rawTickets {
		elapsed := int(now.Sub(rt.createdAt).Minutes())

		// Default SLA: 10 minutes (600 seconds).
		slaMinutes := 10
		if rt.estimatedReadyAt != nil {
			slaMinutes = int(rt.estimatedReadyAt.Sub(rt.createdAt).Minutes())
			if slaMinutes <= 0 {
				slaMinutes = 10
			}
		}

		// SLA proximity: fraction of SLA elapsed, clamped to [0, 1].
		slaProximity := float64(elapsed) / float64(slaMinutes)
		if slaProximity > 1 {
			slaProximity = 1
		}

		// Customer value: stub at 0.5 (requires customer tier data).
		customerValue := 0.5

		// Channel weight.
		cw := channelWeightFor(rt.channel)

		// Complexity inverse: stub at 0.5 (complexity data not joined here).
		complexityInverse := 0.5

		score := computePriority(slaProximity, customerValue, cw, complexityInverse)

		priorities = append(priorities, TicketPriority{
			TicketID:       rt.ticketID,
			OrderNumber:    rt.orderNumber,
			Channel:        rt.channel,
			PriorityScore:  score,
			SLAMinutes:     slaMinutes,
			ElapsedMinutes: elapsed,
			Urgency:        classifyUrgency(elapsed, slaMinutes),
		})
	}

	// Sort descending by priority score.
	sort.Slice(priorities, func(i, j int) bool {
		return priorities[i].PriorityScore > priorities[j].PriorityScore
	})

	return priorities, nil
}
