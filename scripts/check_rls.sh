#!/usr/bin/env bash
set -euo pipefail

# Tables that are explicitly exempt from RLS
EXEMPT_TABLES="audit_log|audit_log_.*|_tx_test"

DB_URL="${TEST_DATABASE_URL:-postgres://fireline:fireline@localhost:5432/fireline?sslmode=disable}"

echo "Checking RLS enforcement on all tables..."

# Use docker compose exec locally; CI has psql available natively.
run_psql() {
  if command -v psql &>/dev/null; then
    psql "$DB_URL" -t -A -c "$1"
  else
    docker compose exec -T postgres psql -U fireline -d fireline -t -A -c "$1"
  fi
}

TABLES_WITHOUT_RLS=$(run_psql "
SELECT t.tablename
FROM pg_tables t
JOIN pg_class c ON c.relname = t.tablename AND c.relnamespace = 'public'::regnamespace
WHERE t.schemaname = 'public'
  AND t.tablename !~ '^(${EXEMPT_TABLES})$'
  AND NOT (
    c.relrowsecurity = true
    AND c.relforcerowsecurity = true
    AND EXISTS (
      SELECT 1 FROM pg_catalog.pg_policies p
      WHERE p.schemaname = 'public' AND p.tablename = t.tablename
    )
  )
ORDER BY t.tablename;
")

if [ -n "$TABLES_WITHOUT_RLS" ]; then
  echo "FAIL: The following tables are missing RLS policies:"
  echo "$TABLES_WITHOUT_RLS" | while read -r table; do
    echo "  - $table"
  done
  echo ""
  echo "Every tenant-scoped table MUST have:"
  echo "  ALTER TABLE <name> ENABLE ROW LEVEL SECURITY;"
  echo "  ALTER TABLE <name> FORCE ROW LEVEL SECURITY;"
  echo "  CREATE POLICY org_isolation ON <name> USING (org_id = current_setting('app.current_org_id')::UUID);"
  exit 1
fi

echo "PASS: All tables have RLS enforced."
