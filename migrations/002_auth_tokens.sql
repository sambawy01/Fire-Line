-- Refresh tokens for session management
CREATE TABLE refresh_tokens (
    token_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(org_id),
    user_id UUID NOT NULL REFERENCES users(user_id),
    token_hash TEXT NOT NULL, -- SHA-256 hash of opaque token
    expires_at TIMESTAMPTZ NOT NULL,
    revoked BOOLEAN NOT NULL DEFAULT false,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_refresh_tokens_org ON refresh_tokens(org_id);
CREATE INDEX idx_refresh_tokens_user ON refresh_tokens(user_id, revoked);
CREATE INDEX idx_refresh_tokens_hash ON refresh_tokens(token_hash) WHERE revoked = false;

ALTER TABLE refresh_tokens ENABLE ROW LEVEL SECURITY;
ALTER TABLE refresh_tokens FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON refresh_tokens
    USING (org_id = current_setting('app.current_org_id')::UUID);

GRANT SELECT, INSERT, UPDATE ON refresh_tokens TO fireline_app;

-- MFA recovery codes
CREATE TABLE mfa_recovery_codes (
    code_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(org_id),
    user_id UUID NOT NULL REFERENCES users(user_id),
    code_hash TEXT NOT NULL, -- bcrypt hash
    used BOOLEAN NOT NULL DEFAULT false,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_mfa_recovery_org ON mfa_recovery_codes(org_id);

ALTER TABLE mfa_recovery_codes ENABLE ROW LEVEL SECURITY;
ALTER TABLE mfa_recovery_codes FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON mfa_recovery_codes
    USING (org_id = current_setting('app.current_org_id')::UUID);

GRANT SELECT, INSERT, UPDATE ON mfa_recovery_codes TO fireline_app;

-- Brute-force lockout tracking
CREATE TABLE lockout_attempts (
    attempt_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    org_id UUID NOT NULL,
    target_type TEXT NOT NULL CHECK (target_type IN ('email', 'pin', 'tablet')),
    target_id TEXT NOT NULL, -- email address, employee_id, or tablet device_id
    location_id UUID,
    success BOOLEAN NOT NULL DEFAULT false,
    ip_address INET,
    attempted_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_lockout_target ON lockout_attempts(target_type, target_id, attempted_at DESC);
CREATE INDEX idx_lockout_location ON lockout_attempts(location_id, attempted_at DESC) WHERE location_id IS NOT NULL;

-- No RLS on lockout_attempts — brute-force tracking is security-critical
-- and must work even if tenant context is not yet established (pre-auth)
GRANT INSERT, SELECT ON lockout_attempts TO fireline_app;
