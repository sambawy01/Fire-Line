-- Core schema for FireLine multi-tenant platform
-- All tables use org_id for Row-Level Security

-- Organizations
CREATE TABLE organizations (
    org_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    slug TEXT NOT NULL UNIQUE,
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'cancelled')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Locations
CREATE TABLE locations (
    location_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(org_id),
    name TEXT NOT NULL,
    address TEXT,
    timezone TEXT NOT NULL DEFAULT 'America/New_York',
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'onboarding')),
    franchise_node_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_locations_org ON locations(org_id);

-- Users (owners, managers)
CREATE TABLE users (
    user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(org_id),
    email TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    display_name TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'staff' CHECK (role IN ('owner', 'gm', 'shift_manager', 'staff', 'read_only')),
    mfa_secret_hash TEXT,
    mfa_enabled BOOLEAN NOT NULL DEFAULT false,
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'locked')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(email)
);
CREATE INDEX idx_users_org ON users(org_id);

-- User-Location access junction (org_id denormalized for RLS performance)
CREATE TABLE user_location_access (
    user_id UUID NOT NULL REFERENCES users(user_id),
    location_id UUID NOT NULL REFERENCES locations(location_id),
    org_id UUID NOT NULL REFERENCES organizations(org_id),
    PRIMARY KEY (user_id, location_id)
);
CREATE INDEX idx_ula_org ON user_location_access(org_id);

-- Employees (staff with PIN auth)
CREATE TABLE employees (
    employee_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(org_id),
    location_id UUID NOT NULL REFERENCES locations(location_id),
    user_id UUID REFERENCES users(user_id),
    display_name TEXT NOT NULL,
    pin_hash TEXT,
    role TEXT NOT NULL DEFAULT 'staff' CHECK (role IN ('owner', 'gm', 'shift_manager', 'staff', 'read_only')),
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'terminated')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_employees_org ON employees(org_id);
CREATE INDEX idx_employees_location ON employees(org_id, location_id);

-- Audit log (NO RLS — separate role with INSERT-only)
CREATE TABLE audit_log (
    log_id BIGINT GENERATED ALWAYS AS IDENTITY,
    org_id UUID NOT NULL,
    user_id UUID,
    action TEXT NOT NULL,
    resource_type TEXT NOT NULL,
    resource_id TEXT,
    detail JSONB,
    ip_address INET,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (log_id, created_at)
) PARTITION BY RANGE (created_at);

CREATE TABLE audit_log_2026_03 PARTITION OF audit_log
    FOR VALUES FROM ('2026-03-01') TO ('2026-04-01');

CREATE TABLE audit_log_2026_04 PARTITION OF audit_log
    FOR VALUES FROM ('2026-04-01') TO ('2026-05-01');

CREATE INDEX idx_audit_org ON audit_log(org_id, created_at DESC);

-- ============================================================
-- ROW-LEVEL SECURITY
-- ============================================================

ALTER TABLE organizations ENABLE ROW LEVEL SECURITY;
ALTER TABLE organizations FORCE ROW LEVEL SECURITY;

ALTER TABLE locations ENABLE ROW LEVEL SECURITY;
ALTER TABLE locations FORCE ROW LEVEL SECURITY;

ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE users FORCE ROW LEVEL SECURITY;

ALTER TABLE user_location_access ENABLE ROW LEVEL SECURITY;
ALTER TABLE user_location_access FORCE ROW LEVEL SECURITY;

ALTER TABLE employees ENABLE ROW LEVEL SECURITY;
ALTER TABLE employees FORCE ROW LEVEL SECURITY;

-- RLS Policies
CREATE POLICY org_isolation ON organizations
    USING (org_id = current_setting('app.current_org_id')::UUID);

CREATE POLICY org_isolation ON locations
    USING (org_id = current_setting('app.current_org_id')::UUID);

CREATE POLICY org_isolation ON users
    USING (org_id = current_setting('app.current_org_id')::UUID);

CREATE POLICY org_isolation ON user_location_access
    USING (org_id = current_setting('app.current_org_id')::UUID);

CREATE POLICY org_isolation ON employees
    USING (org_id = current_setting('app.current_org_id')::UUID);

-- ============================================================
-- APPLICATION ROLE (non-superuser, RLS enforced)
-- ============================================================

DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'fireline_app') THEN
        CREATE ROLE fireline_app LOGIN PASSWORD 'fireline_app';
    END IF;
END
$$;

GRANT USAGE ON SCHEMA public TO fireline_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO fireline_app;
GRANT USAGE ON ALL SEQUENCES IN SCHEMA public TO fireline_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO fireline_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT USAGE ON SEQUENCES TO fireline_app;

-- Audit role: INSERT-only on audit_log
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'fireline_audit') THEN
        CREATE ROLE fireline_audit LOGIN PASSWORD 'fireline_audit';
    END IF;
END
$$;

GRANT USAGE ON SCHEMA public TO fireline_audit;
GRANT INSERT ON audit_log TO fireline_audit;
