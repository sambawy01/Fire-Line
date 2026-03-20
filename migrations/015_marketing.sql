-- Marketing campaigns, loyalty program

CREATE TABLE campaigns (
    campaign_id    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID NOT NULL REFERENCES organizations(org_id),
    location_id    UUID REFERENCES locations(location_id),
    name           TEXT NOT NULL,
    campaign_type  TEXT NOT NULL CHECK (campaign_type IN ('discount', 'bogo', 'happy_hour', 'bundle', 'loyalty_reward', 'custom')),
    status         TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'scheduled', 'active', 'paused', 'completed', 'cancelled')),
    target_segment TEXT,
    channel        TEXT CHECK (channel IN ('email', 'sms', 'push', 'in_app', 'all')),
    discount_type  TEXT CHECK (discount_type IN ('percentage', 'dollar_off', 'bogo', 'bundle')),
    discount_value NUMERIC(10,2),
    min_purchase   BIGINT DEFAULT 0,
    start_at       TIMESTAMPTZ,
    end_at         TIMESTAMPTZ,
    recurring      BOOLEAN NOT NULL DEFAULT false,
    recurrence_rule TEXT,
    redemptions    INT NOT NULL DEFAULT 0,
    revenue_attributed BIGINT NOT NULL DEFAULT 0,
    cost_of_promotion BIGINT NOT NULL DEFAULT 0,
    created_by     UUID REFERENCES users(user_id),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE loyalty_members (
    member_id      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID NOT NULL REFERENCES organizations(org_id),
    guest_id       UUID NOT NULL REFERENCES guest_profiles(guest_id),
    points_balance NUMERIC(12,2) NOT NULL DEFAULT 0,
    lifetime_points NUMERIC(12,2) NOT NULL DEFAULT 0,
    tier           TEXT NOT NULL DEFAULT 'bronze' CHECK (tier IN ('bronze', 'silver', 'gold', 'platinum')),
    joined_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(org_id, guest_id)
);

CREATE TABLE loyalty_transactions (
    transaction_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID NOT NULL REFERENCES organizations(org_id),
    member_id      UUID NOT NULL REFERENCES loyalty_members(member_id),
    type           TEXT NOT NULL CHECK (type IN ('earn', 'redeem', 'adjustment', 'expire')),
    points         NUMERIC(12,2) NOT NULL,
    description    TEXT,
    check_id       UUID REFERENCES checks(check_id),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_campaigns_org ON campaigns(org_id, location_id, status);
CREATE INDEX idx_campaigns_status ON campaigns(org_id, status);
CREATE INDEX idx_loyalty_members_org ON loyalty_members(org_id, tier);
CREATE INDEX idx_loyalty_members_guest ON loyalty_members(org_id, guest_id);
CREATE INDEX idx_loyalty_tx_member ON loyalty_transactions(member_id, created_at DESC);

ALTER TABLE campaigns ENABLE ROW LEVEL SECURITY;
ALTER TABLE campaigns FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON campaigns USING (org_id = current_setting('app.current_org_id')::UUID);
GRANT SELECT, INSERT, UPDATE, DELETE ON campaigns TO fireline_app;

ALTER TABLE loyalty_members ENABLE ROW LEVEL SECURITY;
ALTER TABLE loyalty_members FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON loyalty_members USING (org_id = current_setting('app.current_org_id')::UUID);
GRANT SELECT, INSERT, UPDATE, DELETE ON loyalty_members TO fireline_app;

ALTER TABLE loyalty_transactions ENABLE ROW LEVEL SECURITY;
ALTER TABLE loyalty_transactions FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON loyalty_transactions USING (org_id = current_setting('app.current_org_id')::UUID);
GRANT SELECT, INSERT, UPDATE, DELETE ON loyalty_transactions TO fireline_app;
