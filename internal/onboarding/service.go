package onboarding

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/opsnerve/fireline/internal/event"
)

// Service manages onboarding sessions and wizard state.
type Service struct {
	pool *pgxpool.Pool
	bus  *event.Bus
}

// New creates a new onboarding Service.
func New(pool *pgxpool.Pool, bus *event.Bus) *Service {
	return &Service{pool: pool, bus: bus}
}
