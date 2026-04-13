-- 024_alert_persistence.sql
-- Persist alerts to PostgreSQL instead of in-memory queue.

CREATE TABLE IF NOT EXISTS alerts (
    alert_id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID NOT NULL,
    location_id    UUID,
    module         TEXT NOT NULL,
    severity       TEXT NOT NULL CHECK (severity IN ('critical', 'warning', 'info')),
    title          TEXT NOT NULL,
    description    TEXT,
    status         TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'acknowledged', 'resolved')),
    acknowledged_by UUID,
    resolved_by    UUID,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- RLS: tenant isolation
ALTER TABLE alerts ENABLE ROW LEVEL SECURITY;

CREATE POLICY alerts_tenant_isolation ON alerts
    USING (org_id = current_setting('app.current_org_id')::uuid);

-- Grant to application role
GRANT SELECT, INSERT, UPDATE, DELETE ON alerts TO fireline_app;

-- Indexes for common query patterns
CREATE INDEX idx_alerts_org_status ON alerts (org_id, status);
CREATE INDEX idx_alerts_org_created ON alerts (org_id, created_at DESC);
