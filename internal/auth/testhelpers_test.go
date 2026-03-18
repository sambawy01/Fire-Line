package auth_test

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func getTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://fireline:fireline@localhost:5432/fireline?sslmode=disable"
	}
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		t.Skip("database not available:", err)
	}
	t.Cleanup(func() { pool.Close() })
	return pool
}

func getAppPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dbURL := os.Getenv("TEST_APP_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://fireline_app:fireline_app@localhost:5432/fireline?sslmode=disable"
	}
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		t.Skip("database not available:", err)
	}
	t.Cleanup(func() { pool.Close() })
	return pool
}

func seedTestUser(t *testing.T, superPool *pgxpool.Pool) (orgID, userID string) {
	t.Helper()
	ctx := context.Background()

	err := superPool.QueryRow(ctx,
		"INSERT INTO organizations (name, slug) VALUES ('Auth Test Org', 'auth-test-'||gen_random_uuid()) RETURNING org_id",
	).Scan(&orgID)
	require.NoError(t, err)

	err = superPool.QueryRow(ctx,
		`INSERT INTO users (org_id, email, password_hash, display_name, role)
		 VALUES ($1, 'test-'||gen_random_uuid()||'@test.com', '$2a$12$placeholder', 'Test User', 'owner')
		 RETURNING user_id`,
		orgID,
	).Scan(&userID)
	require.NoError(t, err)

	t.Cleanup(func() {
		superPool.Exec(ctx, "DELETE FROM refresh_tokens WHERE org_id = $1", orgID)
		superPool.Exec(ctx, "DELETE FROM users WHERE org_id = $1", orgID)
		superPool.Exec(ctx, "DELETE FROM organizations WHERE org_id = $1", orgID)
	})

	return orgID, userID
}
