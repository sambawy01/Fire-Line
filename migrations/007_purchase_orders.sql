-- Purchase orders and delivery receiving

CREATE TABLE purchase_orders (
    purchase_order_id  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id             UUID NOT NULL REFERENCES organizations(org_id),
    location_id        UUID NOT NULL REFERENCES locations(location_id),
    vendor_name        TEXT NOT NULL,
    status             TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'approved', 'received', 'cancelled')),
    source             TEXT NOT NULL DEFAULT 'manual' CHECK (source IN ('manual', 'system_recommended')),
    suggested_at       TIMESTAMPTZ,
    approved_by        UUID REFERENCES users(user_id),
    approved_at        TIMESTAMPTZ,
    received_by        UUID REFERENCES users(user_id),
    received_at        TIMESTAMPTZ,
    total_estimated    BIGINT NOT NULL DEFAULT 0,
    total_actual       BIGINT NOT NULL DEFAULT 0,
    notes              TEXT,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE purchase_order_lines (
    po_line_id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id              UUID NOT NULL REFERENCES organizations(org_id),
    purchase_order_id   UUID NOT NULL REFERENCES purchase_orders(purchase_order_id),
    ingredient_id       UUID NOT NULL REFERENCES ingredients(ingredient_id),
    ordered_qty         NUMERIC(12,4) NOT NULL,
    ordered_unit        TEXT NOT NULL,
    estimated_unit_cost INT NOT NULL DEFAULT 0,
    received_qty        NUMERIC(12,4),
    received_unit_cost  INT,
    variance_qty        NUMERIC(12,4),
    variance_flag       TEXT CHECK (variance_flag IN ('exact', 'short', 'over', 'not_received')),
    received_at         TIMESTAMPTZ,
    note                TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Add lead time and usage tracking to ingredient configs
ALTER TABLE ingredient_location_configs
    ADD COLUMN lead_time_days INT NOT NULL DEFAULT 1,
    ADD COLUMN avg_daily_usage NUMERIC(12,4) NOT NULL DEFAULT 0;

-- Indexes
CREATE INDEX idx_po_location ON purchase_orders(org_id, location_id, created_at DESC);
CREATE INDEX idx_po_status ON purchase_orders(org_id, status);
CREATE INDEX idx_po_vendor ON purchase_orders(org_id, vendor_name);
CREATE INDEX idx_po_lines_po ON purchase_order_lines(purchase_order_id);
CREATE INDEX idx_po_lines_ingredient ON purchase_order_lines(org_id, ingredient_id);

-- RLS: purchase_orders
ALTER TABLE purchase_orders ENABLE ROW LEVEL SECURITY;
ALTER TABLE purchase_orders FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON purchase_orders
    USING (org_id = current_setting('app.current_org_id')::UUID);
GRANT SELECT, INSERT, UPDATE, DELETE ON purchase_orders TO fireline_app;

-- RLS: purchase_order_lines
ALTER TABLE purchase_order_lines ENABLE ROW LEVEL SECURITY;
ALTER TABLE purchase_order_lines FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON purchase_order_lines
    USING (org_id = current_setting('app.current_org_id')::UUID);
GRANT SELECT, INSERT, UPDATE, DELETE ON purchase_order_lines TO fireline_app;
