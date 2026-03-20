-- Vendor reliability scoring, OTIF tracking, and price intelligence

CREATE TABLE vendor_scores (
    score_id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID NOT NULL REFERENCES organizations(org_id),
    location_id    UUID NOT NULL REFERENCES locations(location_id),
    vendor_name    TEXT NOT NULL,
    overall_score  NUMERIC(5,2) NOT NULL DEFAULT 0,
    price_score    NUMERIC(5,2) NOT NULL DEFAULT 50,
    delivery_score NUMERIC(5,2) NOT NULL DEFAULT 50,
    quality_score  NUMERIC(5,2) NOT NULL DEFAULT 50,
    accuracy_score NUMERIC(5,2) NOT NULL DEFAULT 50,
    total_orders   INT NOT NULL DEFAULT 0,
    otif_rate      NUMERIC(5,2) NOT NULL DEFAULT 0,
    on_time_rate   NUMERIC(5,2) NOT NULL DEFAULT 0,
    in_full_rate   NUMERIC(5,2) NOT NULL DEFAULT 0,
    avg_lead_days  NUMERIC(5,2) NOT NULL DEFAULT 0,
    calculated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(org_id, location_id, vendor_name)
);

CREATE TABLE ingredient_price_history (
    price_id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID NOT NULL REFERENCES organizations(org_id),
    ingredient_id  UUID NOT NULL REFERENCES ingredients(ingredient_id),
    vendor_name    TEXT NOT NULL,
    unit_cost      INT NOT NULL,
    quantity       NUMERIC(12,4),
    source         TEXT NOT NULL CHECK (source IN ('po_received', 'manual', 'market')),
    recorded_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_vendor_scores_location ON vendor_scores(org_id, location_id);
CREATE INDEX idx_vendor_scores_name ON vendor_scores(org_id, vendor_name);
CREATE INDEX idx_price_history_ingredient ON ingredient_price_history(org_id, ingredient_id, recorded_at DESC);
CREATE INDEX idx_price_history_vendor ON ingredient_price_history(org_id, vendor_name, recorded_at DESC);

ALTER TABLE vendor_scores ENABLE ROW LEVEL SECURITY;
ALTER TABLE vendor_scores FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON vendor_scores USING (org_id = current_setting('app.current_org_id')::UUID);
GRANT SELECT, INSERT, UPDATE, DELETE ON vendor_scores TO fireline_app;

ALTER TABLE ingredient_price_history ENABLE ROW LEVEL SECURITY;
ALTER TABLE ingredient_price_history FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON ingredient_price_history USING (org_id = current_setting('app.current_org_id')::UUID);
GRANT SELECT, INSERT, UPDATE, DELETE ON ingredient_price_history TO fireline_app;
