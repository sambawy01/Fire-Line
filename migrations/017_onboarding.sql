CREATE TABLE onboarding_sessions (
    session_id     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID NOT NULL REFERENCES organizations(org_id),
    user_id        UUID NOT NULL REFERENCES users(user_id),
    current_step   TEXT NOT NULL DEFAULT 'profile' CHECK (current_step IN ('profile', 'pos_connect', 'importing', 'first_insights', 'concept_type', 'priorities', 'modules', 'checklist', 'complete')),
    profile_data   JSONB NOT NULL DEFAULT '{}',
    concept_type   TEXT,
    priorities     TEXT[] NOT NULL DEFAULT '{}',
    active_modules TEXT[] NOT NULL DEFAULT '{}',
    insights_data  JSONB NOT NULL DEFAULT '{}',
    completed_at   TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE onboarding_checklist_items (
    item_id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID NOT NULL REFERENCES organizations(org_id),
    title          TEXT NOT NULL,
    description    TEXT,
    category       TEXT NOT NULL,
    priority       INT NOT NULL DEFAULT 0,
    completed      BOOLEAN NOT NULL DEFAULT false,
    completed_at   TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_onboarding_org ON onboarding_sessions(org_id, user_id);
CREATE INDEX idx_checklist_org ON onboarding_checklist_items(org_id, completed);

ALTER TABLE onboarding_sessions ENABLE ROW LEVEL SECURITY;
ALTER TABLE onboarding_sessions FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON onboarding_sessions USING (org_id = current_setting('app.current_org_id')::UUID);
GRANT SELECT, INSERT, UPDATE, DELETE ON onboarding_sessions TO fireline_app;

ALTER TABLE onboarding_checklist_items ENABLE ROW LEVEL SECURITY;
ALTER TABLE onboarding_checklist_items FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON onboarding_checklist_items USING (org_id = current_setting('app.current_org_id')::UUID);
GRANT SELECT, INSERT, UPDATE, DELETE ON onboarding_checklist_items TO fireline_app;
