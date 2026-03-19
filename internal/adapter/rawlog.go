package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RawDataLog stores immutable copies of all inbound data for audit and replay.
type RawDataLog struct {
	pool *pgxpool.Pool
}

// RawDataEntry represents a single raw data log entry.
type RawDataEntry struct {
	LogID       string          `json:"log_id"`
	OrgID       string          `json:"org_id"`
	LocationID  string          `json:"location_id"`
	AdapterType string          `json:"adapter_type"`
	DataType    string          `json:"data_type"` // "order", "menu", "employee", "webhook"
	ExternalID  string          `json:"external_id"`
	Payload     json.RawMessage `json:"payload"`
	ReceivedAt  time.Time       `json:"received_at"`
}

// NewRawDataLog creates a raw data log writer.
func NewRawDataLog(pool *pgxpool.Pool) *RawDataLog {
	return &RawDataLog{pool: pool}
}

// Append writes a raw payload to the immutable log within a transaction.
func (r *RawDataLog) Append(ctx context.Context, tx pgx.Tx, entry RawDataEntry) (string, error) {
	var logID string
	err := tx.QueryRow(ctx,
		`INSERT INTO raw_data_log (org_id, location_id, adapter_type, data_type, external_id, payload)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING log_id`,
		entry.OrgID, entry.LocationID, entry.AdapterType, entry.DataType, entry.ExternalID, entry.Payload,
	).Scan(&logID)
	if err != nil {
		return "", fmt.Errorf("append to raw data log: %w", err)
	}
	return logID, nil
}

// AppendDirect writes a raw payload outside of an existing transaction (uses pool directly).
func (r *RawDataLog) AppendDirect(ctx context.Context, entry RawDataEntry) (string, error) {
	var logID string
	err := r.pool.QueryRow(ctx,
		`INSERT INTO raw_data_log (org_id, location_id, adapter_type, data_type, external_id, payload)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING log_id`,
		entry.OrgID, entry.LocationID, entry.AdapterType, entry.DataType, entry.ExternalID, entry.Payload,
	).Scan(&logID)
	if err != nil {
		return "", fmt.Errorf("append to raw data log: %w", err)
	}
	return logID, nil
}
