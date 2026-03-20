-- Multi-location portfolio hierarchy, benchmarking, and best practices

-- Portfolio nodes: hierarchical groupings (org, region, district, location)
CREATE TABLE portfolio_nodes (
    node_id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(org_id) ON DELETE CASCADE,
    parent_node_id  UUID REFERENCES portfolio_nodes(node_id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    node_type       TEXT NOT NULL CHECK (node_type IN ('org', 'region', 'district', 'location')),
    location_id     UUID REFERENCES locations(location_id) ON DELETE CASCADE,
    sort_order      INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT portfolio_nodes_location_unique UNIQUE (org_id, location_id)
        DEFERRABLE INITIALLY DEFERRED
);

CREATE INDEX idx_portfolio_nodes_org ON portfolio_nodes(org_id);
CREATE INDEX idx_portfolio_nodes_parent ON portfolio_nodes(parent_node_id);
CREATE INDEX idx_portfolio_nodes_location ON portfolio_nodes(location_id);

ALTER TABLE portfolio_nodes ENABLE ROW LEVEL SECURITY;
CREATE POLICY portfolio_nodes_org ON portfolio_nodes
    USING (org_id = current_setting('app.current_org_id', true)::uuid);

-- Location benchmarks: periodic computed metrics for each location
CREATE TABLE location_benchmarks (
    benchmark_id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id              UUID NOT NULL REFERENCES organizations(org_id) ON DELETE CASCADE,
    location_id         UUID NOT NULL REFERENCES locations(location_id) ON DELETE CASCADE,
    period_start        TIMESTAMPTZ NOT NULL,
    period_end          TIMESTAMPTZ NOT NULL,
    revenue             BIGINT NOT NULL DEFAULT 0,
    food_cost_pct       NUMERIC(6,3) NOT NULL DEFAULT 0,
    labor_cost_pct      NUMERIC(6,3) NOT NULL DEFAULT 0,
    avg_check_cents     BIGINT NOT NULL DEFAULT 0,
    check_count         INT NOT NULL DEFAULT 0,
    revenue_percentile  NUMERIC(6,3),
    food_cost_percentile NUMERIC(6,3),
    labor_cost_percentile NUMERIC(6,3),
    avg_check_percentile NUMERIC(6,3),
    computed_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT location_benchmarks_period_unique UNIQUE (org_id, location_id, period_start, period_end)
);

CREATE INDEX idx_location_benchmarks_org ON location_benchmarks(org_id, period_start, period_end);
CREATE INDEX idx_location_benchmarks_location ON location_benchmarks(location_id, period_start);

ALTER TABLE location_benchmarks ENABLE ROW LEVEL SECURITY;
CREATE POLICY location_benchmarks_org ON location_benchmarks
    USING (org_id = current_setting('app.current_org_id', true)::uuid);

-- Best practices: detected patterns from top-quartile locations
CREATE TABLE best_practices (
    practice_id     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(org_id) ON DELETE CASCADE,
    title           TEXT NOT NULL,
    description     TEXT NOT NULL,
    metric          TEXT NOT NULL,
    source_location_id UUID REFERENCES locations(location_id) ON DELETE SET NULL,
    impact_pct      NUMERIC(6,3) NOT NULL DEFAULT 0,
    status          TEXT NOT NULL DEFAULT 'suggested' CHECK (status IN ('suggested', 'adopted', 'dismissed')),
    detected_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_best_practices_org ON best_practices(org_id, status);
CREATE INDEX idx_best_practices_source ON best_practices(source_location_id);

ALTER TABLE best_practices ENABLE ROW LEVEL SECURITY;
CREATE POLICY best_practices_org ON best_practices
    USING (org_id = current_setting('app.current_org_id', true)::uuid);
