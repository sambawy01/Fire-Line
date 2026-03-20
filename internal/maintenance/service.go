package maintenance

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/opsnerve/fireline/internal/event"
)

// Service provides equipment and maintenance management capabilities.
type Service struct {
	pool *pgxpool.Pool
	bus  *event.Bus
}

// New creates a new maintenance service.
func New(pool *pgxpool.Pool, bus *event.Bus) *Service {
	return &Service{pool: pool, bus: bus}
}
