# SP17: Vendor Intelligence — Scoring, OTIF Tracking & Price Intelligence

**Date:** 2026-03-20
**Status:** Approved
**Scope:** Vendor Reliability Score (VRS), OTIF tracking, price trend analysis, price anomaly detection, vendor comparison, enhanced vendor dashboard
**Maps to:** Build Plan Sprints 36-37 (Vendor Intelligence)

---

## 1. Database — Migration 014

New migration: `migrations/014_vendor_scoring.sql`

```sql
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
```

RLS on both. Indexes on vendor_name, ingredient_id, recorded_at.

---

## 2. Vendor Reliability Score (VRS)

New file: `internal/vendor/scoring.go`

### Score Calculation (rolling 90 days)

- **Price Score (0-100)**: 100 - (avg price variance from baseline * 10). Lower prices = higher score.
- **Delivery Score (0-100)**: on_time_rate (% of POs received within lead_time_days of approval)
- **Quality Score (0-100)**: 100 - (short_delivery_rate * 100). Higher in-full rate = higher score.
- **Accuracy Score (0-100)**: % of PO lines where variance_flag = 'exact'

Overall: weighted average — price 30%, delivery 25%, quality 25%, accuracy 20%

### OTIF Tracking

- **On-Time**: PO received_at - approved_at <= lead_time_days
- **In-Full**: all PO lines have variance_flag = 'exact' or 'over'
- **OTIF**: both on-time AND in-full

### Methods

- `CalculateVendorScores(ctx, orgID, locationID)` — compute all scores from PO/receiving data, upsert vendor_scores
- `GetVendorScores(ctx, orgID, locationID)` — list all vendor scores
- `GetVendorScorecard(ctx, orgID, locationID, vendorName)` — detailed scorecard with trend
- `CompareVendors(ctx, orgID, locationID, ingredientID)` — side-by-side for vendors supplying same ingredient

---

## 3. Price Intelligence

New file: `internal/vendor/pricing.go`

### Price Tracking

When a PO is received, record actual unit cost per ingredient per vendor in `ingredient_price_history`.

### Price Trend

- `GetPriceTrend(ctx, orgID, ingredientID, vendorName, months)` — historical unit costs over time
- `DetectPriceAnomalies(ctx, orgID, locationID)` — for each ingredient, compare latest price to 90-day moving average. Flag if > 2σ deviation. Emit `vendor.price.anomaly` alert.

### Vendor Selection Recommendation

- `RecommendVendor(ctx, orgID, locationID, ingredientID)` — rank vendors by composite: price (40%) + reliability score (40%) + lead time (20%). Return top recommendation with reasoning.

### Methods

- `RecordPrice(ctx, orgID, ingredientID, vendorName, unitCost, quantity, source)` — INSERT into price_history
- `GetPriceTrend(ctx, orgID, ingredientID, vendorName, months int)` — SELECT grouped by month
- `DetectPriceAnomalies(ctx, orgID, locationID)` — batch check all ingredients
- `RecommendVendor(ctx, orgID, locationID, ingredientID)` — scored recommendation

---

## 4. API Endpoints

```
POST   /api/v1/vendors/scores/calculate     — Recalculate all vendor scores
GET    /api/v1/vendors/scores               — List vendor scores (query: location_id)
GET    /api/v1/vendors/scorecard            — Vendor scorecard (query: location_id, vendor_name)
GET    /api/v1/vendors/compare              — Compare vendors for ingredient (query: location_id, ingredient_id)
GET    /api/v1/vendors/price-trend          — Price history (query: ingredient_id, vendor_name, months)
GET    /api/v1/vendors/price-anomalies      — Detect price anomalies (query: location_id)
GET    /api/v1/vendors/recommend            — Vendor recommendation (query: location_id, ingredient_id)
```

---

## 5. Web Dashboard — Enhanced Vendor Page

Rewrite `VendorPage.tsx` with tabs:

### Tab 1: Vendor Scorecards
- Cards per vendor: name, overall score (large), sub-score bars, OTIF rate, total orders
- Color-coded: >80 green, 60-80 yellow, <60 red
- "Recalculate" button

### Tab 2: Price Intelligence
- Ingredient selector → line chart showing price over time per vendor
- Price anomaly alerts highlighted
- Vendor recommendation card for selected ingredient

### Tab 3: Vendor Comparison
- Select an ingredient → side-by-side vendor cards
- Each showing: price, reliability, lead time, OTIF rate
- Recommended vendor highlighted

---

## 6. Testing
- VRS: known PO data → correct score computation
- OTIF: on-time + in-full → OTIF true; late → false
- Price anomaly: spike above 2σ → flagged
- Vendor recommendation: highest composite score selected
