-- Budget management for financial variance analysis

CREATE TABLE budgets (
    budget_id     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id        UUID NOT NULL REFERENCES organizations(org_id),
    location_id   UUID NOT NULL REFERENCES locations(location_id),
    period_type   TEXT NOT NULL CHECK (period_type IN ('daily', 'weekly', 'monthly')),
    period_start  DATE NOT NULL,
    period_end    DATE NOT NULL,
    revenue_target      BIGINT NOT NULL DEFAULT 0,
    food_cost_pct_target NUMERIC(5,2) NOT NULL DEFAULT 30.00,
    labor_cost_pct_target NUMERIC(5,2) NOT NULL DEFAULT 28.00,
    cogs_target         BIGINT NOT NULL DEFAULT 0,
    created_by    UUID REFERENCES users(user_id),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(org_id, location_id, period_type, period_start)
);

CREATE INDEX idx_budgets_location ON budgets(org_id, location_id, period_start DESC);
CREATE INDEX idx_budgets_period ON budgets(org_id, location_id, period_type, period_start);

ALTER TABLE budgets ENABLE ROW LEVEL SECURITY;
ALTER TABLE budgets FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON budgets
    USING (org_id = current_setting('app.current_org_id')::UUID);
GRANT SELECT, INSERT, UPDATE, DELETE ON budgets TO fireline_app;
