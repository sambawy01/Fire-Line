# SP5: Customer Intelligence Design Spec

**Date:** 2026-03-19
**Status:** Approved
**Scope:** Customer table, visit/spend metrics, Ollama-powered segmentation and AI summaries

---

## 1. Database — New Customers Table

New migration `005_customers.sql`:

```sql
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
    total_spend INT NOT NULL DEFAULT 0,     -- cents
    avg_check INT NOT NULL DEFAULT 0,       -- cents
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
```

## 2. Backend — Customer Intelligence Service

New Go service at `internal/customer/`.

### Ollama Integration

HTTP client calling Ollama's local API:
- **URL:** `http://localhost:11434/api/generate` (configurable via `OLLAMA_URL` env var)
- **Model:** configurable via `OLLAMA_MODEL` env var (default: `llama3.2`)
- **Timeout:** 30 seconds per request

### Ollama Client (`internal/customer/ollama.go`)

```go
type OllamaClient struct {
    baseURL string
    model   string
    client  *http.Client
}

type OllamaRequest struct {
    Model  string `json:"model"`
    Prompt string `json:"prompt"`
    Stream bool   `json:"stream"`
}

type OllamaResponse struct {
    Response string `json:"response"`
}

func NewOllamaClient(baseURL, model string) *OllamaClient
func (c *OllamaClient) Generate(ctx context.Context, prompt string) (string, error)
```

### Prompt Templates

**Segmentation:**
```
You are a restaurant customer analyst. Given this customer data:
- Total visits: {visits}
- Total spend: ${spend}
- Average check: ${avg_check}
- First visit: {first_visit}
- Last visit: {last_visit}
- Days since last visit: {recency}

Classify this customer as exactly ONE of these segments:
- new: visited 1-2 times
- regular: visits consistently, moderate spend
- vip: high frequency (5+ visits) AND high spend (top 20%)
- at_risk: was regular but hasn't visited in 14+ days
- lapsed: hasn't visited in 30+ days

Reply with ONLY the segment label, nothing else.
```

**Summary:**
```
You are a restaurant manager's AI assistant. Write a 1-2 sentence actionable insight about this customer:
- Name: {name}
- Segment: {segment}
- Total visits: {visits}
- Total spend: ${spend}
- Average check: ${avg_check}
- First visit: {first_visit}
- Last visit: {last_visit}
- Days since last visit: {recency}

Be specific, concise, and actionable. Focus on what the manager should do.
```

### Service Methods

```go
func New(pool *pgxpool.Pool, bus *event.Bus, ollama *OllamaClient) *Service

func (s *Service) GetCustomers(ctx context.Context, orgID, locationID string) ([]CustomerDetail, error)
func (s *Service) GetSummary(ctx context.Context, orgID, locationID string) (*CustomerSummary, error)
func (s *Service) AnalyzeAll(ctx context.Context, orgID, locationID string) (*AnalyzeResult, error)
```

**GetCustomers:** Query customers table for the location, ordered by total_spend DESC.

**GetSummary:** Aggregate: total count, avg total_spend (lifetime value), count by segment.

**AnalyzeAll:** For each customer at the location:
1. Call Ollama with segmentation prompt → update `segment`
2. Call Ollama with summary prompt → update `ai_summary` + `ai_summary_updated_at`
3. Return count of customers analyzed + any errors

If Ollama is unavailable, AnalyzeAll returns an error. GetCustomers/GetSummary still work (they just show existing data without AI).

### Types

```go
type CustomerDetail struct {
    CustomerID         string  `json:"customer_id"`
    Name               string  `json:"name"`
    Email              string  `json:"email"`
    Phone              string  `json:"phone"`
    FirstVisit         *time.Time `json:"first_visit"`
    LastVisit          *time.Time `json:"last_visit"`
    TotalVisits        int     `json:"total_visits"`
    TotalSpend         int64   `json:"total_spend"`
    AvgCheck           int64   `json:"avg_check"`
    Segment            string  `json:"segment"`
    AISummary          string  `json:"ai_summary"`
    AISummaryUpdatedAt *time.Time `json:"ai_summary_updated_at"`
}

type CustomerSummary struct {
    TotalCustomers int     `json:"total_customers"`
    AvgLifetimeValue int64 `json:"avg_lifetime_value"`  // cents
    VIPCount       int     `json:"vip_count"`
    AtRiskCount    int     `json:"at_risk_count"`       // at_risk + lapsed
    SegmentCounts  map[string]int `json:"segment_counts"`
}

type AnalyzeResult struct {
    Analyzed int    `json:"analyzed"`
    Errors   int    `json:"errors"`
    Message  string `json:"message"`
}
```

### API Endpoints

**`GET /api/v1/customers?location_id=X`**
- Returns: `{ customers: CustomerDetail[] }`

**`GET /api/v1/customers/summary?location_id=X`**
- Returns: `CustomerSummary`

**`POST /api/v1/customers/analyze?location_id=X`**
- Triggers Ollama analysis for all customers at location
- Returns: `AnalyzeResult`

### Files

- `internal/customer/ollama.go` — Ollama HTTP client
- `internal/customer/customer.go` — Service with types, GetCustomers, GetSummary, AnalyzeAll
- `internal/api/customer_handler.go` — HTTP handlers
- `cmd/fireline/main.go` — Create service, register routes

## 3. Demo Seed Data

Seed 12 customers across both locations with varied visit patterns:
- 2 VIPs (high spend, frequent visits)
- 4 regulars (moderate)
- 2 new (1-2 visits)
- 2 at-risk (no visit in 14+ days)
- 2 lapsed (no visit in 30+ days)

Pre-populate `segment` field based on seed data. `ai_summary` left NULL until user clicks "Analyze".

## 4. Frontend — Customer Intelligence Page

### New Files

- `web/src/pages/CustomerPage.tsx`
- `web/src/hooks/useCustomers.ts`
- Modify: `web/src/lib/api.ts` — Add customerApi + types
- Modify: `web/src/App.tsx` — Add `/customers` route
- Modify: `web/src/components/Layout.tsx` — Add nav item

### TypeScript Types

```typescript
export interface CustomerDetail {
  customer_id: string;
  name: string;
  email: string;
  phone: string;
  first_visit: string | null;
  last_visit: string | null;
  total_visits: number;
  total_spend: number;
  avg_check: number;
  segment: 'new' | 'regular' | 'vip' | 'lapsed' | 'at_risk';
  ai_summary: string;
  ai_summary_updated_at: string | null;
}

export interface CustomerSummary {
  total_customers: number;
  avg_lifetime_value: number;
  vip_count: number;
  at_risk_count: number;
  segment_counts: Record<string, number>;
}

export interface AnalyzeResult {
  analyzed: number;
  errors: number;
  message: string;
}
```

### API Client

```typescript
export const customerApi = {
  getCustomers(locationId: string) {
    return request<{ customers: CustomerDetail[] }>(`/customers?location_id=${locationId}`);
  },
  getSummary(locationId: string) {
    return request<CustomerSummary>(`/customers/summary?location_id=${locationId}`);
  },
  analyze(locationId: string) {
    return request<AnalyzeResult>(`/customers/analyze?location_id=${locationId}`, { method: 'POST' });
  },
};
```

### Hooks

```typescript
export function useCustomers(locationId: string | null)    // staleTime 30s
export function useCustomerSummary(locationId: string | null) // staleTime 30s
export function useAnalyzeCustomers()                       // useMutation, invalidates customer queries
```

### Page Layout

**Row 1 — KPI Cards (4):**
- Total Customers (UserCheck icon, gray)
- Avg Lifetime Value ($) (DollarSign, blue)
- VIP Customers (Crown/Star, emerald)
- At Risk (AlertTriangle, red)

**Row 2 — Analyze button + Segment filter:**
- "Analyze with AI" button (triggers POST, shows loading spinner, disables during processing)
- Segment dropdown filter: All, New, Regular, VIP, At Risk, Lapsed

**Row 3 — Customer DataTable:**
| Column | Key | Sortable | Align | Render |
|--------|-----|----------|-------|--------|
| Name | name | yes | left | bold |
| Segment | segment | yes | center | StatusBadge |
| Visits | total_visits | yes | right | — |
| Total Spend | total_spend | yes | right | cents→$ |
| Avg Check | avg_check | yes | right | cents→$ |
| Last Visit | last_visit | yes | right | date format |
| AI Insight | ai_summary | no | left | truncated, tooltip for full |

Segment badge: vip→success, regular→info, new→neutral, at_risk→warning, lapsed→critical

### Navigation

Icon: `UserCheck` from lucide-react. After Vendors:
```typescript
{ to: '/customers', label: 'Customers', icon: UserCheck }
```

## 5. Conventions

- Ollama URL/model configurable via env vars (graceful fallback if unavailable)
- AnalyzeAll processes sequentially (not parallel) to avoid overloading local Ollama
- AI summary and segment stored on customer record, not recomputed on every read
- Frontend works without Ollama — just shows "No AI summary" until Analyze is clicked
- All money in cents, standard patterns for hooks/handlers/TenantTx
