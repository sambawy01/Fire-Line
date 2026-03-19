-- Customer intelligence table

CREATE TABLE customers (
    customer_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(org_id),
    location_id UUID NOT NULL REFERENCES locations(location_id),
    name TEXT,
    email TEXT,
    phone TEXT,
    first_visit TIMESTAMPTZ,
    last_visit TIMESTAMPTZ,
    total_visits INT NOT NULL DEFAULT 0,
    total_spend INT NOT NULL DEFAULT 0,
    avg_check INT NOT NULL DEFAULT 0,
    segment TEXT NOT NULL DEFAULT 'new' CHECK (segment IN ('new', 'regular', 'vip', 'lapsed', 'at_risk')),
    ai_summary TEXT,
    ai_summary_updated_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_customers_org ON customers(org_id);
CREATE INDEX idx_customers_location ON customers(org_id, location_id);
CREATE INDEX idx_customers_segment ON customers(org_id, location_id, segment);

ALTER TABLE customers ENABLE ROW LEVEL SECURITY;
ALTER TABLE customers FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON customers
    USING (org_id = current_setting('app.current_org_id')::UUID);
GRANT SELECT, INSERT, UPDATE, DELETE ON customers TO fireline_app;
