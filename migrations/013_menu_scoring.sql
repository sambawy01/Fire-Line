-- Menu item scoring, classification, and simulation sandbox

ALTER TABLE menu_items
    ADD COLUMN margin_score NUMERIC(5,2) NOT NULL DEFAULT 0,
    ADD COLUMN velocity_score NUMERIC(5,2) NOT NULL DEFAULT 0,
    ADD COLUMN complexity_score NUMERIC(5,2) NOT NULL DEFAULT 0,
    ADD COLUMN satisfaction_score NUMERIC(5,2) NOT NULL DEFAULT 0,
    ADD COLUMN strategic_score NUMERIC(5,2) NOT NULL DEFAULT 50,
    ADD COLUMN classification TEXT,
    ADD COLUMN classification_changed_at TIMESTAMPTZ;

CREATE TABLE menu_simulations (
    simulation_id  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID NOT NULL REFERENCES organizations(org_id),
    location_id    UUID NOT NULL REFERENCES locations(location_id),
    simulation_type TEXT NOT NULL CHECK (simulation_type IN ('price_change', 'item_removal', 'ingredient_price_change')),
    parameters     JSONB NOT NULL DEFAULT '{}',
    results        JSONB NOT NULL DEFAULT '{}',
    created_by     UUID REFERENCES users(user_id),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_simulations_location ON menu_simulations(org_id, location_id, created_at DESC);

ALTER TABLE menu_simulations ENABLE ROW LEVEL SECURITY;
ALTER TABLE menu_simulations FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON menu_simulations USING (org_id = current_setting('app.current_org_id')::UUID);
GRANT SELECT, INSERT, UPDATE, DELETE ON menu_simulations TO fireline_app;
