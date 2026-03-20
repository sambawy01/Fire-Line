# SP11: Phase 2 Hardening — Integration Tests, Error Handling & Data Freshness

**Date:** 2026-03-19
**Status:** Approved
**Scope:** End-to-end integration tests, graceful degradation, data provenance tagging, cross-module event flow verification
**Maps to:** Build Plan Sprint 21 (Phase 2 Integration Testing, Hardening & MVP Alpha Gate)

---

## 1. Integration Test Suite

New file: `internal/integration_test.go` (extend existing)

### Test Scenarios

1. **Full Lifecycle Test**: signup → create org → seed locations → seed ingredients → seed orders → P&L calculation → inventory usage → variance detection
2. **Count → PO Flow**: create count → enter quantities → submit count → variance calculated → PO auto-generated for breaching ingredients → approve PO → receive delivery → discrepancies detected
3. **Financial Drill-Down Chain**: P&L → cost centers → category drill-down → menu item → ingredient → vendor history
4. **Cross-Module Event Flow**: order event → pipeline processes → inventory depleted → financial P&L updated → alert generated
5. **Budget vs Actual**: create budget → generate orders → calculate variance → verify on_track/over/under status
6. **Error Handling**: invalid org_id → proper error response, missing location_id → 400, expired JWT → 401

### Test Infrastructure

Use the existing test database (Postgres in Docker). Tests run with `go test ./internal/ -tags=integration -count=1`.

---

## 2. Data Freshness Tagging

Add `data_freshness` metadata to API responses that depend on POS sync or calculations.

### Freshness Levels
- `live` — real-time data from current POS sync (< 5 minutes old)
- `recent` — data from last hour
- `stale` — data older than 1 hour
- `estimated` — calculated/projected, not directly measured

Add to P&L, inventory, and alert responses as a top-level field:
```json
{
  "data_freshness": "recent",
  "last_sync_at": "2026-03-19T21:30:00Z",
  ...existing fields...
}
```

---

## 3. Error Handling Improvements

### Graceful Degradation
- If POS adapter disconnected: return cached data tagged as `stale`
- If calculation fails mid-stream: return partial results with error field
- Standardize all error responses across handlers

### Health Check Enhancement
Extend `GET /health/ready` to report module-level status:
```json
{
  "status": "ready",
  "modules": {
    "database": "ok",
    "event_bus": "ok",
    "adapters": {"toast": "connected"}
  },
  "last_event_at": "2026-03-19T21:30:00Z"
}
```

---

## 4. API Response Consistency Audit

Ensure all API handlers follow the same patterns:
- Error responses use `api.WriteError` with consistent code format
- Success responses use `api.WriteJSON`
- All list endpoints return empty arrays (not null) when no data
- All endpoints validate required query params

---

## 5. Testing Strategy

- Integration tests with real database
- Each test creates its own org/location for isolation (no cross-test pollution)
- Tests cover the happy path + key error paths
- Run with: `go test ./internal/ -run TestIntegration -v -count=1`
