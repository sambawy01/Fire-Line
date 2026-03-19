package adapter

import (
	"context"
	"time"
)

// Capability represents a specific integration capability an adapter can provide.
type Capability string

const (
	CapReadOrders     Capability = "READ_ORDERS"
	CapReadMenu       Capability = "READ_MENU"
	CapReadEmployees  Capability = "READ_EMPLOYEES"
	CapReadPayments   Capability = "READ_PAYMENTS"
	CapWrite86Status  Capability = "WRITE_86_STATUS"
	CapWriteMenu      Capability = "WRITE_MENU"
	CapEmitCallbacks  Capability = "EMIT_CALLBACKS"
	CapReadInventory  Capability = "READ_INVENTORY"
	CapReadLabor      Capability = "READ_LABOR"
	CapReadFinancials Capability = "READ_FINANCIALS"
)

// Status represents the lifecycle state of an adapter instance.
type Status string

const (
	StatusInitializing Status = "initializing"
	StatusActive       Status = "active"
	StatusPaused       Status = "paused"
	StatusErrored      Status = "errored"
	StatusDisconnected Status = "disconnected"
)

// SyncMode determines how data is synchronized.
type SyncMode string

const (
	SyncRealtime     SyncMode = "REALTIME"
	SyncNearRealtime SyncMode = "NEAR_REALTIME"
	SyncBatch        SyncMode = "BATCH"
)

// Config holds per-location adapter configuration.
type Config struct {
	AdapterID    string            `json:"adapter_id"`
	AdapterType  string            `json:"adapter_type"` // e.g., "toast", "square"
	OrgID        string            `json:"org_id"`
	LocationID   string            `json:"location_id"`
	Credentials  map[string]string `json:"credentials"` // encrypted at rest
	SyncMode     SyncMode          `json:"sync_mode"`
	PollInterval time.Duration     `json:"poll_interval"`
	Status       Status            `json:"status"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

// DataFreshness tracks the last sync timestamp per data type.
type DataFreshness struct {
	DataType      string    `json:"data_type"`
	LastSyncAt    time.Time `json:"last_sync_at"`
	RecordCount   int64     `json:"record_count"`
	SourceVersion string    `json:"source_version"`
}

// Adapter is the interface all POS/integration adapters must implement.
type Adapter interface {
	// Type returns the adapter type identifier (e.g., "toast", "square").
	Type() string

	// Capabilities returns the set of capabilities this adapter supports.
	Capabilities() []Capability

	// HasCapability checks if the adapter supports a specific capability.
	HasCapability(cap Capability) bool

	// Initialize sets up the adapter with its configuration.
	Initialize(ctx context.Context, cfg Config) error

	// Shutdown gracefully stops the adapter.
	Shutdown(ctx context.Context) error

	// HealthCheck returns nil if the adapter is healthy.
	HealthCheck(ctx context.Context) error

	// Status returns the current adapter status.
	GetStatus() Status

	// GetDataFreshness returns freshness info for a data type.
	GetDataFreshness(dataType string) (*DataFreshness, error)
}

// OrderReader can read orders from a POS.
type OrderReader interface {
	ReadOrders(ctx context.Context, since time.Time, limit int) ([]NormalizedOrder, error)
}

// MenuReader can read menu data from a POS.
type MenuReader interface {
	ReadMenu(ctx context.Context) ([]NormalizedMenuItem, error)
}

// EmployeeReader can read employee data from a POS.
type EmployeeReader interface {
	ReadEmployees(ctx context.Context) ([]NormalizedEmployee, error)
}

// StatusWriter can write 86 (out-of-stock) status to a POS.
type StatusWriter interface {
	Write86Status(ctx context.Context, itemID string, available bool) error
}
