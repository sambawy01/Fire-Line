# SP10: Financial Intelligence Depth — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add budget management with variance analysis, COGS cost center breakdown, transaction anomaly detection, P&L drill-down, and period comparison to the financial intelligence module.

**Architecture:** New budgets table (migration 008), three new service files (budget.go, costcenter.go, drilldown.go) extending the financial Service, enhanced anomaly detection, new API endpoints, and enhanced web Financial page with 4 tabs.

**Tech Stack:** Go 1.22, PostgreSQL 16, React/TypeScript/Tailwind/Recharts

---

## File Map

### Backend — New Files
| File | Responsibility |
|------|---------------|
| `migrations/008_financial_budgets.sql` | budgets table + RLS + indexes |
| `internal/financial/budget.go` | Budget CRUD, variance analysis, period comparison |
| `internal/financial/costcenter.go` | COGS by category, top ingredients per category |
| `internal/financial/drilldown.go` | Menu item cost breakdown, ingredient cost, vendor history |
| `internal/financial/txanomaly.go` | Transaction-level anomaly detection (voids, comps, off-hours, discounts) |
| `internal/financial/budget_test.go` | Budget variance calculation tests |
| `internal/financial/txanomaly_test.go` | Anomaly detection tests |
| `internal/api/financial_handler.go` | New HTTP handlers for all new endpoints |

### Backend — Modified Files
| File | Change |
|------|--------|
| `internal/auth/rbac.go` | Add `financial:budget` permission |
| `internal/api/handlers.go` | Register new financial routes |

### Web — Modified Files
| File | Change |
|------|--------|
| `web/src/lib/api.ts` | Add budget, cost center, anomaly, drilldown types + API methods |
| `web/src/hooks/useFinancial.ts` | Add hooks for new endpoints |
| `web/src/pages/FinancialPage.tsx` | Rewrite with 4 tabs: P&L, Cost Centers, Anomalies, Budget |

---

## Task 1: Migration 008 — Budgets Table

**Files:** Create `migrations/008_financial_budgets.sql`

The complete SQL from the spec: budgets table with period_type, targets, UNIQUE constraint, RLS, indexes.

After creating: `atlas migrate hash`, apply, verify, commit.

---

## Task 2: RBAC + Budget Service

**Files:**
- Modify: `internal/auth/rbac.go` — add `financial:budget` to `gm` and `owner`
- Create: `internal/financial/budget.go` — Budget types, CRUD, variance calc, period comparison
- Create: `internal/financial/budget_test.go` — variance calculation tests

Budget service methods on `*Service`:
- `CreateBudget` — INSERT with ON CONFLICT UPDATE
- `GetBudget` — find budget covering a date
- `ListBudgets` — list by location + period type
- `CalculateBudgetVariance` — compare actual P&L vs budget, return variance with on_track/over/under status
- `CalculatePeriodComparison` — current vs last week/month/year

---

## Task 3: Cost Center + Drill-Down Services

**Files:**
- Create: `internal/financial/costcenter.go` — COGS by category with top ingredients
- Create: `internal/financial/drilldown.go` — item → ingredient → vendor drill-down chain

Service methods:
- `GetCostCenterBreakdown` — COGS aggregated by ingredient category, top 5 ingredients per category
- `GetItemCostBreakdown` — revenue, COGS, margin per menu item in a category
- `GetIngredientCostBreakdown` — cost per ingredient for a menu item via recipe_explosion
- `GetIngredientVendorHistory` — vendor name, cost history from PO lines

---

## Task 4: Transaction Anomaly Detection

**Files:**
- Create: `internal/financial/txanomaly.go` — void/comp/off-hours/discount detection
- Create: `internal/financial/txanomaly_test.go` — Z-score tests

Service methods:
- `DetectTransactionAnomalies` — query voids, comps, off-hours, discount rates for current day, compare against 30-day baseline via Z-score, emit alerts for critical anomalies

---

## Task 5: HTTP Handlers

**Files:**
- Create: `internal/api/financial_handler.go` — new handler methods on `*FinancialHandler`
- Modify: `internal/api/handlers.go` — register routes

Endpoints:
- `POST /api/v1/financial/budgets`
- `GET /api/v1/financial/budgets`
- `GET /api/v1/financial/budget-variance`
- `GET /api/v1/financial/cost-centers`
- `GET /api/v1/financial/transaction-anomalies`
- `GET /api/v1/financial/drilldown/items`
- `GET /api/v1/financial/drilldown/ingredients`
- `GET /api/v1/financial/drilldown/vendor`
- `GET /api/v1/financial/period-comparison`

---

## Task 6: Web Dashboard — Enhanced Financial Page

**Files:**
- Modify: `web/src/lib/api.ts` — types + API methods
- Modify: `web/src/hooks/useFinancial.ts` — new hooks
- Rewrite: `web/src/pages/FinancialPage.tsx` — 4-tab layout

Tab 1 (P&L): existing + budget variance badges + period comparison
Tab 2 (Cost Centers): donut chart + category table + ingredient expand
Tab 3 (Anomalies): existing Z-score + transaction anomalies
Tab 4 (Budget): entry form + variance table

---

## Task 7: E2E Test

Run all Go tests, rebuild server, test endpoints via curl, build web frontend.
