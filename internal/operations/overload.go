package operations

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/opsnerve/fireline/internal/event"
	"github.com/opsnerve/fireline/internal/tenant"
	"github.com/opsnerve/fireline/pkg/database"
)

// OverloadStatus describes the current overload state of a location.
type OverloadStatus struct {
	IsOverloaded     bool              `json:"is_overloaded"`
	CapacityPct      float64           `json:"capacity_pct"`
	Severity         string            `json:"severity"`
	ActiveResponses  []OverloadResponse `json:"active_responses"`
	SuggestedActions []SuggestedAction  `json:"suggested_actions"`
}

// OverloadResponse describes an applied or available overload response tier.
type OverloadResponse struct {
	Tier        int    `json:"tier"`
	Action      string `json:"action"`
	Description string `json:"description"`
	AutoApplied bool   `json:"auto_applied"`
}

// SuggestedAction describes a recommended action during overload.
type SuggestedAction struct {
	ActionType  string `json:"action_type"`
	Description string `json:"description"`
	Impact      string `json:"impact"`
	ItemID      string `json:"item_id,omitempty"`
}

// classifyOverload returns the overload severity based on capacity percentage.
// "normal" < 85, "elevated" 85-95, "critical" > 95.
func classifyOverload(capacityPct float64) string {
	switch {
	case capacityPct > 95:
		return "critical"
	case capacityPct >= 85:
		return "elevated"
	default:
		return "normal"
	}
}

// GetOverloadStatus computes the current overload state for a location.
func (s *Service) GetOverloadStatus(ctx context.Context, orgID, locationID string) (*OverloadStatus, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	capacity, err := s.CalculateCapacity(ctx, orgID, locationID)
	if err != nil {
		return nil, fmt.Errorf("calculate capacity: %w", err)
	}

	pct := capacity.OverallLoadPct
	severity := classifyOverload(pct)

	status := &OverloadStatus{
		IsOverloaded:     severity != "normal",
		CapacityPct:      pct,
		Severity:         severity,
		ActiveResponses:  []OverloadResponse{},
		SuggestedActions: []SuggestedAction{},
	}

	if severity == "normal" {
		return status, nil
	}

	// Elevated: suggest extending delivery times, flag complex items.
	if severity == "elevated" {
		status.SuggestedActions = append(status.SuggestedActions,
			SuggestedAction{
				ActionType:  "extend_delivery_time",
				Description: "Extend quoted delivery times by 10-15 minutes",
				Impact:      "Reduces customer expectation mismatch during peak load",
			},
		)

		// Flag complex items (high complexity_score) for potential 86.
		complexItems, err := s.queryHighComplexityItems(tenantCtx, orgID, locationID)
		if err == nil {
			for _, item := range complexItems {
				status.SuggestedActions = append(status.SuggestedActions, SuggestedAction{
					ActionType:  "flag_complex_item",
					Description: fmt.Sprintf("Consider 86ing '%s' to reduce kitchen load", item.name),
					Impact:      "Frees station capacity; complexity score " + fmt.Sprintf("%.0f", item.complexityScore),
					ItemID:      item.menuItemID,
				})
			}
		}
	}

	// Critical: suggest 86 of low-complexity items, call staff, close delivery.
	if severity == "critical" {
		status.ActiveResponses = append(status.ActiveResponses,
			OverloadResponse{
				Tier:        1,
				Action:      "extend_delivery_time",
				Description: "Delivery times automatically extended by 20 minutes",
				AutoApplied: true,
			},
		)

		lowComplexItems, err := s.queryLowComplexityItems(tenantCtx, orgID, locationID)
		if err == nil {
			for _, item := range lowComplexItems {
				status.SuggestedActions = append(status.SuggestedActions, SuggestedAction{
					ActionType:  "eighty_six_item",
					Description: fmt.Sprintf("86 '%s' immediately (low complexity, easy to restore)", item.name),
					Impact:      "Reduces ticket complexity; complexity score " + fmt.Sprintf("%.0f", item.complexityScore),
					ItemID:      item.menuItemID,
				})
			}
		}

		status.SuggestedActions = append(status.SuggestedActions,
			SuggestedAction{
				ActionType:  "call_staff",
				Description: "Call in additional kitchen staff immediately",
				Impact:      "Adds human capacity; reduces ticket back-log",
			},
			SuggestedAction{
				ActionType:  "close_delivery",
				Description: "Temporarily close delivery channel to protect dine-in service",
				Impact:      "Reduces incoming order volume by estimated 30-40%",
			},
		)

		// Query near-SLA tickets.
		nearSLA, err := s.queryNearSLATickets(tenantCtx, orgID, locationID)
		if err == nil && nearSLA > 0 {
			status.SuggestedActions = append(status.SuggestedActions, SuggestedAction{
				ActionType:  "prioritize_near_sla",
				Description: fmt.Sprintf("%d ticket(s) approaching SLA — prioritize immediately", nearSLA),
				Impact:      "Prevents SLA breach and customer dissatisfaction",
			})
		}
	}

	return status, nil
}

// menuItemComplexity is an internal helper struct for overload queries.
type menuItemComplexity struct {
	menuItemID      string
	name            string
	complexityScore float64
}

// queryHighComplexityItems returns menu items with complexity_score >= 60.
func (s *Service) queryHighComplexityItems(ctx context.Context, orgID, locationID string) ([]menuItemComplexity, error) {
	var items []menuItemComplexity
	err := database.TenantTx(ctx, s.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx,
			`SELECT menu_item_id, name, COALESCE(complexity_score, 50)
			 FROM menu_items
			 WHERE org_id = $1 AND is_available = true
			   AND COALESCE(complexity_score, 50) >= 60
			 ORDER BY complexity_score DESC
			 LIMIT 5`,
			orgID,
		)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var item menuItemComplexity
			if err := rows.Scan(&item.menuItemID, &item.name, &item.complexityScore); err != nil {
				return err
			}
			items = append(items, item)
		}
		return rows.Err()
	})
	return items, err
}

// queryLowComplexityItems returns menu items with complexity_score < 30.
func (s *Service) queryLowComplexityItems(ctx context.Context, orgID, locationID string) ([]menuItemComplexity, error) {
	var items []menuItemComplexity
	err := database.TenantTx(ctx, s.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx,
			`SELECT menu_item_id, name, COALESCE(complexity_score, 50)
			 FROM menu_items
			 WHERE org_id = $1 AND is_available = true
			   AND COALESCE(complexity_score, 50) < 30
			 ORDER BY complexity_score ASC
			 LIMIT 5`,
			orgID,
		)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var item menuItemComplexity
			if err := rows.Scan(&item.menuItemID, &item.name, &item.complexityScore); err != nil {
				return err
			}
			items = append(items, item)
		}
		return rows.Err()
	})
	return items, err
}

// queryNearSLATickets returns the count of active tickets within 2 minutes of their estimated_ready_at.
func (s *Service) queryNearSLATickets(ctx context.Context, orgID, locationID string) (int, error) {
	var count int
	err := database.TenantTx(ctx, s.pool, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			`SELECT COUNT(*)
			 FROM kds_tickets
			 WHERE org_id = $1 AND location_id = $2
			   AND status NOT IN ('ready', 'delivered', 'cancelled')
			   AND estimated_ready_at IS NOT NULL
			   AND estimated_ready_at <= now() + INTERVAL '2 minutes'`,
			orgID, locationID,
		).Scan(&count)
	})
	return count, err
}

// ApplyOverloadResponse logs an overload response action and emits an event.
func (s *Service) ApplyOverloadResponse(ctx context.Context, orgID, locationID, actionType, itemID string) error {
	s.bus.Publish(ctx, event.Envelope{
		EventType:  "operations.overload.response",
		OrgID:      orgID,
		LocationID: locationID,
		Source:     "operations",
		Payload: map[string]string{
			"action_type": actionType,
			"item_id":     itemID,
		},
	})
	return nil
}
