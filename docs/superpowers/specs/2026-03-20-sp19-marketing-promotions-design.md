# SP19: Marketing — Campaign Engine, Promotions & Loyalty Program

**Date:** 2026-03-20
**Status:** Approved
**Scope:** Campaign CRUD, promotion types, segment targeting, campaign tracking, loyalty program, marketing dashboard
**Maps to:** Build Plan Sprints 39-40 (Marketing & Promotions)

---

## 1. Database — Migration 015

```sql
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
    ab_test        BOOLEAN NOT NULL DEFAULT false,
    ab_variant     TEXT,
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
```

RLS on all 3. Indexes on org+location, status, guest_id, member_id.

---

## 2. Campaign Engine

New file: `internal/marketing/campaign.go`

Types: `Campaign`, `CampaignSummary`

Methods:
- `CreateCampaign(ctx, orgID, campaign)` — INSERT with validation
- `ListCampaigns(ctx, orgID, locationID, status)` — filtered list
- `GetCampaign(ctx, orgID, campaignID)` — single campaign
- `UpdateCampaign(ctx, orgID, campaignID, updates)` — edit draft/scheduled
- `ActivateCampaign(ctx, orgID, campaignID)` — set status=active
- `PauseCampaign(ctx, orgID, campaignID)` — pause active
- `CompleteCampaign(ctx, orgID, campaignID)` — mark completed
- `TrackRedemption(ctx, orgID, campaignID, checkID, amount)` — increment redemptions + revenue
- `SimulateCampaign(ctx, orgID, campaign)` — estimate projected redemptions based on segment size and historical response rates

---

## 3. Loyalty Program

New file: `internal/marketing/loyalty.go`

Types: `LoyaltyMember`, `LoyaltyTransaction`

Tier thresholds: bronze (0), silver (500 lifetime), gold (2000), platinum (5000)

Methods:
- `EnrollMember(ctx, orgID, guestID)` — create loyalty member
- `EarnPoints(ctx, orgID, memberID, points, description, checkID)` — add points, update tier
- `RedeemPoints(ctx, orgID, memberID, points, description)` — deduct, validate balance
- `GetMember(ctx, orgID, guestID)` — member details
- `ListMembers(ctx, orgID, tier, limit)` — list by tier
- `GetTransactionHistory(ctx, orgID, memberID)` — earn/redeem history
- `RecalculateTier(member)` — pure function based on lifetime points

---

## 4. Marketing Analytics

New file: `internal/marketing/analytics.go`

Methods:
- `GetCampaignMetrics(ctx, orgID)` — aggregate: active campaigns, total redemptions, revenue attributed, avg redemption rate
- `GetLoyaltyMetrics(ctx, orgID)` — total members, by tier, avg points balance, total points issued/redeemed

---

## 5. API Endpoints

```
POST   /api/v1/marketing/campaigns              — Create campaign
GET    /api/v1/marketing/campaigns              — List campaigns
GET    /api/v1/marketing/campaigns/{id}         — Get campaign
PUT    /api/v1/marketing/campaigns/{id}         — Update campaign
POST   /api/v1/marketing/campaigns/{id}/activate — Activate
POST   /api/v1/marketing/campaigns/{id}/pause    — Pause
POST   /api/v1/marketing/campaigns/simulate      — Simulate impact

POST   /api/v1/marketing/loyalty/enroll          — Enroll guest
POST   /api/v1/marketing/loyalty/earn            — Earn points
POST   /api/v1/marketing/loyalty/redeem          — Redeem points
GET    /api/v1/marketing/loyalty/member/{guest_id} — Member details
GET    /api/v1/marketing/loyalty/members         — List members

GET    /api/v1/marketing/analytics/campaigns     — Campaign metrics
GET    /api/v1/marketing/analytics/loyalty        — Loyalty metrics
```

---

## 6. Web Dashboard — Marketing Page

New page: `MarketingPage.tsx` (route `/marketing`, nav after Reports)

### Tab 1: Campaigns
- Campaign cards: name, type badge, status badge, segment, redemptions, revenue
- "Create Campaign" button → form modal
- Click campaign → detail view

### Tab 2: Loyalty
- KPI cards: total members, by tier breakdown, avg points
- Member list with search
- Tier distribution pie chart

### Tab 3: Analytics
- Campaign performance: redemption rate, revenue attributed, cost vs revenue
- Loyalty trends: enrollment rate, points velocity

---

## 7. RBAC
- `marketing:read` — view campaigns/loyalty (roles: `shift_manager`, `gm`, `owner`)
- `marketing:write` — create/manage campaigns (roles: `gm`, `owner`)

## 8. Testing
- Campaign lifecycle: draft → scheduled → active → completed
- Loyalty: earn → balance increases, redeem → decreases, tier upgrades at thresholds
- Tier calculation: 499 pts = bronze, 500 = silver, 2000 = gold, 5000 = platinum
