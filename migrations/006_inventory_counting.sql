-- Inventory counting, waste logging, and variance analysis

CREATE TABLE inventory_counts (
    count_id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID NOT NULL REFERENCES organizations(org_id),
    location_id    UUID NOT NULL REFERENCES locations(location_id),
    counted_by     UUID NOT NULL REFERENCES employees(employee_id),
    count_type     TEXT NOT NULL CHECK (count_type IN ('full', 'spot_check')),
    status         TEXT NOT NULL DEFAULT 'in_progress' CHECK (status IN ('in_progress', 'submitted', 'approved')),
    started_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    submitted_at   TIMESTAMPTZ,
    approved_by    UUID REFERENCES users(user_id),
    approved_at    TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE inventory_count_lines (
    count_line_id    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id           UUID NOT NULL REFERENCES organizations(org_id),
    count_id         UUID NOT NULL REFERENCES inventory_counts(count_id),
    location_id      UUID NOT NULL REFERENCES locations(location_id),
    ingredient_id    UUID NOT NULL REFERENCES ingredients(ingredient_id),
    expected_qty     NUMERIC(12,4),
    counted_qty      NUMERIC(12,4),
    unit             TEXT NOT NULL,
    note             TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE waste_logs (
    waste_id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID NOT NULL REFERENCES organizations(org_id),
    location_id    UUID NOT NULL REFERENCES locations(location_id),
    ingredient_id  UUID NOT NULL REFERENCES ingredients(ingredient_id),
    quantity       NUMERIC(12,4) NOT NULL,
    unit           TEXT NOT NULL,
    reason         TEXT NOT NULL CHECK (reason IN ('expired', 'dropped', 'overcooked', 'contaminated', 'overproduction', 'other')),
    logged_by      UUID NOT NULL REFERENCES employees(employee_id),
    logged_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    note           TEXT
);

CREATE TABLE inventory_variances (
    variance_id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id              UUID NOT NULL REFERENCES organizations(org_id),
    location_id         UUID NOT NULL REFERENCES locations(location_id),
    ingredient_id       UUID NOT NULL REFERENCES ingredients(ingredient_id),
    count_id            UUID NOT NULL REFERENCES inventory_counts(count_id),
    period_start        TIMESTAMPTZ NOT NULL,
    period_end          TIMESTAMPTZ NOT NULL,
    theoretical_usage   NUMERIC(12,4) NOT NULL,
    actual_usage        NUMERIC(12,4) NOT NULL,
    variance_qty        NUMERIC(12,4) NOT NULL,
    variance_cents      INT NOT NULL,
    cause_probabilities JSONB NOT NULL DEFAULT '{}',
    severity            TEXT NOT NULL DEFAULT 'info' CHECK (severity IN ('info', 'warning', 'critical')),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Indexes
CREATE INDEX idx_inventory_counts_location ON inventory_counts(org_id, location_id, started_at DESC);
CREATE INDEX idx_count_lines_count ON inventory_count_lines(count_id);
CREATE INDEX idx_count_lines_ingredient ON inventory_count_lines(org_id, ingredient_id);
CREATE INDEX idx_waste_logs_location ON waste_logs(org_id, location_id, logged_at DESC);
CREATE INDEX idx_variances_location ON inventory_variances(org_id, location_id, created_at DESC);

-- RLS: inventory_counts
ALTER TABLE inventory_counts ENABLE ROW LEVEL SECURITY;
ALTER TABLE inventory_counts FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON inventory_counts
    USING (org_id = current_setting('app.current_org_id')::UUID);
GRANT SELECT, INSERT, UPDATE, DELETE ON inventory_counts TO fireline_app;

-- RLS: inventory_count_lines
ALTER TABLE inventory_count_lines ENABLE ROW LEVEL SECURITY;
ALTER TABLE inventory_count_lines FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON inventory_count_lines
    USING (org_id = current_setting('app.current_org_id')::UUID);
GRANT SELECT, INSERT, UPDATE, DELETE ON inventory_count_lines TO fireline_app;

-- RLS: waste_logs
ALTER TABLE waste_logs ENABLE ROW LEVEL SECURITY;
ALTER TABLE waste_logs FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON waste_logs
    USING (org_id = current_setting('app.current_org_id')::UUID);
GRANT SELECT, INSERT, UPDATE, DELETE ON waste_logs TO fireline_app;

-- RLS: inventory_variances
ALTER TABLE inventory_variances ENABLE ROW LEVEL SECURITY;
ALTER TABLE inventory_variances FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON inventory_variances
    USING (org_id = current_setting('app.current_org_id')::UUID);
GRANT SELECT, INSERT, UPDATE, DELETE ON inventory_variances TO fireline_app;
