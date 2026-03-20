# SP15: Customer Intelligence — Guest Profiles, CLV, Segmentation & Churn Prediction

**Date:** 2026-03-20
**Status:** Approved
**Scope:** Guest profile resolution via payment tokens, CLV scoring, RFM segmentation, churn prediction, win-back alerts, customer analytics dashboard
**Maps to:** Build Plan Sprints 29-30 (Customer Intelligence)

---

## 1. Database — Migration 012

New migration: `migrations/012_guest_profiles.sql`

### New Tables

```sql
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
```

RLS on both tables. Indexes on guest_id, payment_token_hash, segment, churn_risk, last_visit_at.

---

## 2. Guest Profile Resolution Engine

New file: `internal/customer/profiles.go`

### Resolution Logic
- Hash payment method token from checks/payments with SHA-256
- Look up existing guest by `(org_id, payment_token_hash)`
- If found: update visit metrics (total_visits++, total_spend += check total, recalculate avg_check)
- If not found: create new guest profile at privacy tier 1 (behavioral/anonymous)
- Cross-channel: same payment token used across dine-in/takeout/delivery → same guest

### Methods on `*Service`
- `ResolveGuest(ctx, orgID, checkID)` — resolve or create guest from check's payment data, create visit record, update metrics
- `GetGuestProfile(ctx, orgID, guestID)` — full profile
- `ListGuests(ctx, orgID, locationID, sortBy, limit, offset)` — sortable by clv_score, total_spend, last_visit_at, churn_risk
- `EnrichGuest(ctx, orgID, guestID, name, email, phone)` — upgrade to tier 2/3

---

## 3. CLV Scoring Engine

New file: `internal/customer/clv.go`

### CLV Formula (Tier 0)
```
CLV = avg_check * visit_frequency * predicted_lifespan * margin_factor
```
Where:
- `avg_check` = total_spend / total_visits
- `visit_frequency` = total_visits / months_since_first_visit
- `predicted_lifespan` = 24 months (default, adjusted by churn risk)
- `margin_factor` = 0.65 (average gross margin)

### Methods
- `CalculateCLV(profile)` → float64 (pure function)
- `RecalculateAllCLV(ctx, orgID)` — batch update all guest profiles

---

## 4. RFM Segmentation

New file: `internal/customer/segmentation.go`

### RFM Scoring
- **Recency**: days since last visit → quintile 1-5 (5 = most recent)
- **Frequency**: total visits → quintile 1-5 (5 = most frequent)
- **Monetary**: total spend → quintile 1-5 (5 = highest spend)

### Segment Labels
| RFM Score Range | Segment |
|----------------|---------|
| R5 F5 M5 | Champion |
| R4-5 F4-5 M4-5 | Loyal Regular |
| R4-5 F1-3 M4-5 | High-Value New |
| R3-4 F3-4 M3-4 | Promising |
| R1-2 F4-5 M4-5 | At Risk (High-Value) |
| R1-2 F1-3 M1-3 | Lost |
| R4-5 F1-2 M1-3 | New Discoverer |
| Default | Casual |

### Methods
- `SegmentGuest(recency, frequency, monetary int) string` — pure function
- `RunSegmentation(ctx, orgID)` — batch classify all guests with 3+ visits

---

## 5. Churn Prediction

New file: `internal/customer/churn.go`

### Churn Model (Tier 0 — frequency decay)
- Calculate average inter-visit interval for the guest
- Expected next visit = last_visit + avg_interval
- If days_overdue = today - expected_next_visit:
  - days_overdue < 0: low (0.1)
  - days_overdue 0-7: medium (0.3)
  - days_overdue 7-21: high (0.6)
  - days_overdue > 21: critical (0.85)
- For guests with < 3 visits: default to "low"

### Alert Integration
- When high-CLV guest enters "high" or "critical" churn: emit `customer.churn.risk` alert
- Alert: "High-value guest [segment] hasn't visited in [N] days — consider win-back action"

### Methods
- `PredictChurn(visits []time.Time) (risk string, probability float64)` — pure function
- `RunChurnPrediction(ctx, orgID)` — batch update all guest profiles
- `PredictNextVisit(visits []time.Time) time.Time` — pure function

---

## 6. API Endpoints

```
GET    /api/v1/customers/guests              — List guests (query: location_id, sort_by, limit, offset)
GET    /api/v1/customers/guests/{id}         — Guest profile
PUT    /api/v1/customers/guests/{id}/enrich  — Enrich with name/email/phone
POST   /api/v1/customers/resolve             — Resolve guest from check_id
POST   /api/v1/customers/analytics/refresh   — Recalculate CLV, segmentation, churn for all guests
GET    /api/v1/customers/analytics/segments  — Segment distribution counts
GET    /api/v1/customers/analytics/churn     — Churn risk distribution
GET    /api/v1/customers/analytics/clv       — CLV distribution (histogram buckets)
```

---

## 7. Web Dashboard — Enhanced Customer Page

Rewrite `CustomerPage.tsx` with tabs:

### Tab 1: Guest List
- DataTable: name (or "Anonymous #X"), segment badge, CLV $, total visits, last visit, churn risk badge
- Sortable by CLV, visits, spend, churn
- Click row → detail view with visit history timeline

### Tab 2: Analytics
- Segment distribution donut chart
- CLV distribution histogram
- Churn risk breakdown (cards: X low, Y medium, Z high, W critical)
- Key stats: avg CLV, avg visits, avg check

### Tab 3: At Risk
- Filtered list of guests with churn_risk = 'high' or 'critical'
- Sorted by CLV descending (highest value at risk first)
- Action buttons: "Create Win-Back Campaign" (placeholder)

---

## 8. RBAC
- Existing `customer:read` covers all read endpoints
- `customer:write` covers enrich and analytics refresh (roles: `gm`, `owner`)

---

## 9. Testing
- CLV calculation: known inputs → expected CLV value
- RFM segmentation: known quintiles → correct segment label
- Churn prediction: visit intervals → correct risk tier
- Guest resolution: same payment token → same guest profile
- Privacy tier upgrade: enrich with email → tier 3
