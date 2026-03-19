package adapter

import "time"

// NormalizedOrder is the canonical order representation regardless of POS source.
type NormalizedOrder struct {
	ExternalID   string               `json:"external_id"`
	OrgID        string               `json:"org_id"`
	LocationID   string               `json:"location_id"`
	OrderNumber  string               `json:"order_number"`
	Status       string               `json:"status"` // open, closed, voided
	Channel      string               `json:"channel"` // dine_in, takeout, delivery, drive_thru
	Items        []NormalizedOrderItem `json:"items"`
	Subtotal     int64                `json:"subtotal"`    // cents
	Tax          int64                `json:"tax"`         // cents
	Total        int64                `json:"total"`       // cents
	Tip          int64                `json:"tip"`         // cents
	Discount     int64                `json:"discount"`    // cents
	OpenedAt     time.Time            `json:"opened_at"`
	ClosedAt     *time.Time           `json:"closed_at"`
	Source       string               `json:"source"` // adapter type that produced this
	RawPayloadID string               `json:"raw_payload_id"` // reference to raw data log
}

// NormalizedOrderItem is a single line item within an order.
type NormalizedOrderItem struct {
	ExternalID  string                      `json:"external_id"`
	MenuItemID  string                      `json:"menu_item_id"` // mapped internal ID
	Name        string                      `json:"name"`
	Quantity    int                         `json:"quantity"`
	UnitPrice   int64                       `json:"unit_price"` // cents
	Modifiers   []NormalizedOrderModifier   `json:"modifiers"`
	VoidedAt    *time.Time                  `json:"voided_at"`
	VoidReason  string                      `json:"void_reason"`
	FiredAt     *time.Time                  `json:"fired_at"` // when sent to kitchen
}

// NormalizedOrderModifier is a modifier on an order item.
type NormalizedOrderModifier struct {
	ExternalID string `json:"external_id"`
	Name       string `json:"name"`
	Price      int64  `json:"price"` // cents (can be negative for removals)
}

// NormalizedMenuItem is the canonical menu item representation.
type NormalizedMenuItem struct {
	ExternalID  string `json:"external_id"`
	OrgID       string `json:"org_id"`
	LocationID  string `json:"location_id"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Price       int64  `json:"price"` // cents
	Available   bool   `json:"available"`
	Description string `json:"description"`
	Source      string `json:"source"`
}

// NormalizedEmployee is the canonical employee representation.
type NormalizedEmployee struct {
	ExternalID string `json:"external_id"`
	OrgID      string `json:"org_id"`
	LocationID string `json:"location_id"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Role       string `json:"role"`
	Active     bool   `json:"active"`
	Source     string `json:"source"`
}
