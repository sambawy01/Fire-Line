package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/opsnerve/fireline/internal/tenant"
)

// TenantTx executes a function within a tenant-scoped transaction.
// This is the ONLY way application code should access the database.
//
// It: begins a transaction, sets SET LOCAL app.current_org_id,
// executes the callback, and commits or rolls back.
func TenantTx(ctx context.Context, pool *pgxpool.Pool, fn func(tx pgx.Tx) error) error {
	orgID, err := tenant.OrgIDFrom(ctx)
	if err != nil {
		return fmt.Errorf("tenant tx: %w", err)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_org_id', $1, true)", orgID); err != nil {
		return fmt.Errorf("set tenant context: %w", err)
	}

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
