package marketing

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/opsnerve/fireline/internal/event"
)

// Service provides marketing campaign and loyalty program capabilities.
type Service struct {
	pool *pgxpool.Pool
	bus  *event.Bus
}

// New creates a new marketing service.
func New(pool *pgxpool.Pool, bus *event.Bus) *Service {
	return &Service{pool: pool, bus: bus}
}
