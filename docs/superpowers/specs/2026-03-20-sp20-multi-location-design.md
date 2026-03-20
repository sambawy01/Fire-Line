# SP20: Multi-Location Intelligence — Portfolio, Benchmarking & Best Practices

**Date:** 2026-03-20
**Status:** Approved
**Scope:** Portfolio hierarchy, cross-location aggregation, benchmarking, outlier detection, best practice propagation, portfolio dashboard
**Maps to:** Build Plan Sprints 48-49

---

## 1. Database — Migration 016

```sql
CREATE TABLE portfolio_nodes (
    node_id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID NOT NULL REFERENCES organizations(org_id),
    parent_node_id UUID REFERENCES portfolio_nodes(node_id),
    name           TEXT NOT NULL,
    node_type      TEXT NOT NULL CHECK (node_type IN ('portfolio', 'concept', 'region', 'district', 'cluster', 'location')),
    location_id    UUID REFERENCES locations(location_id),
    is_data_boundary BOOLEAN NOT NULL DEFAULT false,
    metadata       JSONB NOT NULL DEFAULT '{}',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE location_benchmarks (
    benchmark_id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID NOT NULL REFERENCES organizations(org_id),
    location_id    UUID NOT NULL REFERENCES locations(location_id),
    period_start   DATE NOT NULL,
    period_end     DATE NOT NULL,
    revenue        BIGINT NOT NULL DEFAULT 0,
    food_cost_pct  NUMERIC(5,2) NOT NULL DEFAULT 0,
    labor_cost_pct NUMERIC(5,2) NOT NULL DEFAULT 0,
    prime_cost_pct NUMERIC(5,2) NOT NULL DEFAULT 0,
    health_score   NUMERIC(5,2) NOT NULL DEFAULT 0,
    check_count    INT NOT NULL DEFAULT 0,
    avg_check      BIGINT NOT NULL DEFAULT 0,
    percentile_rank INT,
    calculated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(org_id, location_id, period_start)
);

CREATE TABLE best_practices (
    practice_id    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID NOT NULL REFERENCES organizations(org_id),
    source_location_id UUID NOT NULL REFERENCES locations(location_id),
    category       TEXT NOT NULL,
    title          TEXT NOT NULL,
    description    TEXT NOT NULL,
    impact_metric  TEXT NOT NULL,
    impact_value   NUMERIC(10,2) NOT NULL,
    status         TEXT NOT NULL DEFAULT 'detected' CHECK (status IN ('detected', 'recommended', 'adopted', 'dismissed')),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

RLS on all 3. Indexes on org_id, parent_node_id, location_id, period_start.

---

## 2. Portfolio Service

New package: `internal/portfolio/`

### `service.go` — Service struct
### `hierarchy.go` — Portfolio node CRUD

Methods:
- `CreateNode(ctx, orgID, parentID, name, nodeType, locationID)` — INSERT
- `GetHierarchy(ctx, orgID)` — recursive query returning tree structure
- `GetNodeChildren(ctx, orgID, nodeID)` — direct children
- `UpdateNode(ctx, orgID, nodeID, name, metadata)` — UPDATE
- `DeleteNode(ctx, orgID, nodeID)` — DELETE (cascade or check for children)

### `aggregation.go` — Cross-location aggregation

Methods:
- `AggregateKPIs(ctx, orgID, nodeID, from, to)` — roll up revenue, food cost %, labor cost %, prime cost %, health score from all descendant locations
- `GetLocationComparison(ctx, orgID, locationIDs []string, from, to)` — side-by-side metrics

### `benchmarking.go` — Benchmarking engine

Methods:
- `CalculateBenchmarks(ctx, orgID, periodStart, periodEnd)` — for each location: compute metrics, rank against peers, persist to location_benchmarks
- `GetBenchmarks(ctx, orgID, from, to)` — list all benchmarks with percentile ranks
- `DetectOutliers(ctx, orgID, from, to)` — locations where any metric is >1.5 IQR above/below median

### `bestpractices.go` — Best practice detection

Methods:
- `DetectBestPractices(ctx, orgID)` — find top-performing locations, identify operational patterns that differ from underperformers
- `ListBestPractices(ctx, orgID, status)` — list practices
- `AdoptPractice(ctx, orgID, practiceID)` — mark as adopted
- `DismissPractice(ctx, orgID, practiceID)` — dismiss

---

## 3. API Endpoints

```
POST   /api/v1/portfolio/nodes               — Create node
GET    /api/v1/portfolio/hierarchy            — Get full tree
PUT    /api/v1/portfolio/nodes/{id}           — Update node
DELETE /api/v1/portfolio/nodes/{id}           — Delete node
GET    /api/v1/portfolio/kpis                 — Aggregated KPIs for node (query: node_id)
GET    /api/v1/portfolio/comparison           — Location comparison (query: location_ids)
POST   /api/v1/portfolio/benchmarks/calculate — Calculate benchmarks for period
GET    /api/v1/portfolio/benchmarks           — Get benchmarks (query: from, to)
GET    /api/v1/portfolio/outliers             — Detect outliers
GET    /api/v1/portfolio/best-practices       — List best practices
POST   /api/v1/portfolio/best-practices/{id}/adopt — Adopt practice
POST   /api/v1/portfolio/best-practices/{id}/dismiss — Dismiss
```

---

## 4. Web Dashboard — Portfolio Page

New page: `PortfolioPage.tsx` (route `/portfolio`, nav item with `Building2` icon)

### Tab 1: Hierarchy
- Tree visualization: expandable nodes showing name, type badge, health score
- Location nodes show key KPIs inline
- Click node → aggregate KPIs for all descendant locations

### Tab 2: Benchmarking
- "Calculate Benchmarks" button
- Heatmap table: locations as rows, metrics as columns, color intensity by percentile
- Outlier badges on extreme performers

### Tab 3: Best Practices
- Cards: title, source location, category, impact metric + value, status badge
- Adopt/Dismiss buttons

### Tab 4: Comparison
- Multi-select locations → side-by-side KPI cards
- Bar charts comparing selected locations

---

## 5. RBAC
- `portfolio:read` — view hierarchy, benchmarks (roles: `gm`, `owner`)
- `portfolio:write` — manage nodes, adopt practices (roles: `owner`)

## 6. Testing
- Hierarchy: create tree, verify recursive query returns correct structure
- Aggregation: known location metrics → correct rollup
- Percentile ranking: known values → correct percentiles
- Outlier detection: extreme value → flagged
