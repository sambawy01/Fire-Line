-- Guest profiles and visit tracking for customer intelligence

CREATE TABLE guest_profiles (
    guest_id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID NOT NULL REFERENCES organizations(org_id),
    payment_token_hash TEXT,
    privacy_tier   INT NOT NULL DEFAULT 1 CHECK (privacy_tier IN (1, 2, 3)),
    first_name     TEXT,
    email          TEXT,
    phone          TEXT,
    total_visits   INT NOT NULL DEFAULT 0,
    total_spend    BIGINT NOT NULL DEFAULT 0,
    avg_check      BIGINT NOT NULL DEFAULT 0,
    preferred_channel TEXT,
    favorite_items JSONB NOT NULL DEFAULT '[]',
    clv_score      NUMERIC(12,2) NOT NULL DEFAULT 0,
    segment        TEXT,
    churn_risk     TEXT CHECK (churn_risk IN ('low', 'medium', 'high', 'critical')),
    churn_probability NUMERIC(5,4),
    next_visit_predicted DATE,
    last_visit_at  TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(org_id, payment_token_hash)
);

CREATE TABLE guest_visits (
    visit_id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID NOT NULL REFERENCES organizations(org_id),
    guest_id       UUID NOT NULL REFERENCES guest_profiles(guest_id),
    location_id    UUID NOT NULL REFERENCES locations(location_id),
    check_id       UUID REFERENCES checks(check_id),
    channel        TEXT,
    spend          BIGINT NOT NULL DEFAULT 0,
    item_count     INT NOT NULL DEFAULT 0,
    party_size     INT,
    visited_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_guests_org ON guest_profiles(org_id, last_visit_at DESC);
CREATE INDEX idx_guests_token ON guest_profiles(org_id, payment_token_hash);
CREATE INDEX idx_guests_segment ON guest_profiles(org_id, segment);
CREATE INDEX idx_guests_churn ON guest_profiles(org_id, churn_risk);
CREATE INDEX idx_guests_clv ON guest_profiles(org_id, clv_score DESC);
CREATE INDEX idx_guest_visits_guest ON guest_visits(guest_id, visited_at DESC);
CREATE INDEX idx_guest_visits_location ON guest_visits(org_id, location_id, visited_at DESC);

ALTER TABLE guest_profiles ENABLE ROW LEVEL SECURITY;
ALTER TABLE guest_profiles FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON guest_profiles USING (org_id = current_setting('app.current_org_id')::UUID);
GRANT SELECT, INSERT, UPDATE, DELETE ON guest_profiles TO fireline_app;

ALTER TABLE guest_visits ENABLE ROW LEVEL SECURITY;
ALTER TABLE guest_visits FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON guest_visits USING (org_id = current_setting('app.current_org_id')::UUID);
GRANT SELECT, INSERT, UPDATE, DELETE ON guest_visits TO fireline_app;
