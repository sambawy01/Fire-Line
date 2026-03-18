package database

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/opsnerve/fireline/internal/tenant"
)

type Pool struct {
	pool *pgxpool.Pool
}

func NewPool(ctx context.Context, databaseURL string) (*Pool, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database URL: %w", err)
	}

	// BeforeAcquire: validate tenant context exists but do NOT set the GUC here.
	// TenantTx is the sole authority for setting SET LOCAL within a transaction.
	// This hook is a safety net — it logs warnings for connections acquired without context.
	config.BeforeAcquire = func(ctx context.Context, conn *pgx.Conn) bool {
		_, err := tenant.OrgIDFrom(ctx)
		if err != nil {
			slog.Warn("connection acquired without tenant context — TenantTx will enforce", "error", err)
		}
		return true
	}

	// AfterRelease: explicitly clear any session-level GUC to prevent leakage
	// between pool checkouts. Setting to empty string causes UUID cast failure
	// in RLS policies if TenantTx SET LOCAL is somehow bypassed (fail-closed).
	config.AfterRelease = func(conn *pgx.Conn) bool {
		_, err := conn.Exec(context.Background(), "SELECT set_config('app.current_org_id', '', false)")
		if err != nil {
			slog.Error("failed to clear tenant context on release", "error", err)
			return false // destroy connection
		}
		return true
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	return &Pool{pool: pool}, nil
}

func (p *Pool) Close() {
	p.pool.Close()
}

func (p *Pool) Ping(ctx context.Context) error {
	return p.pool.Ping(ctx)
}

// Raw returns the underlying pgxpool.Pool. Use ONLY for migrations, admin tasks,
// and test setup/teardown. Application queries MUST use TenantTx.
func (p *Pool) Raw() *pgxpool.Pool {
	return p.pool
}
