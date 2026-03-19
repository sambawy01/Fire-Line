# SP1: Foundation + Wire-Up Design Spec

**Date:** 2026-03-19
**Status:** Approved
**Scope:** Reusable component library, location context, wire existing 5 pages to real APIs

---

## 1. Reusable Component Library

All components in `src/components/ui/`, built with Tailwind, no external UI library.

### Components

- **DataTable** — Sortable columns, pagination, loading skeleton, empty state. Generic `<DataTable<T>>` with column definitions, data array, optional sort/pagination config.

  Column definition shape:
  ```typescript
  interface Column<T> {
    key: keyof T | string;
    header: string;
    render?: (row: T) => React.ReactNode;  // custom cell renderer
    sortable?: boolean;
    align?: 'left' | 'center' | 'right';
  }
  ```

- **KPICard** — Stat card with label, value, optional icon. No trend/changePercent for now (P&L API returns single-period data; trend comparison is a future enhancement). Props: `label`, `value`, `icon`.
- **StatusBadge** — Colored badge for severity/status/module tags. Props: `variant` (critical/warning/info/success/neutral), `children`.
- **Modal** — Overlay dialog with backdrop click to close, escape key support. Props: `open`, `onClose`, `title`, `children`, `footer`.
- **FormField** — Label + input + error message wrapper. Props: `label`, `error`, `children` (wraps any input element).
- **EmptyState** — Centered icon + message + optional action button. Props: `icon`, `title`, `description`, `action`.
- **LoadingSpinner** — Full-page and inline variants. Props: `size` (sm/md/lg), `fullPage` (boolean).
- **ErrorBanner** — Dismissable error display. Props: `message`, `onDismiss`, `retry` (optional callback).

## 2. State Management & Location Context

### Backend Prerequisite: Location List Endpoint

The backend currently has no endpoint to list a user's accessible locations. SP1 includes adding:

- **`GET /api/v1/locations`** — Returns locations the authenticated user has access to, based on the `user_location_access` table. Requires JWT auth. Response:

  ```json
  { "locations": [{ "id": "uuid", "name": "Downtown", "org_id": "uuid" }] }
  ```

  Backend implementation: new handler in `internal/api/` that queries `locations` joined with `user_location_access` for the authenticated user's ID (from JWT claims).

### Location Type

```typescript
interface Location {
  id: string;
  name: string;
  org_id: string;
}
```

### Location Store (`stores/location.ts`)

Zustand store tracking the user's selected location:

```typescript
interface LocationState {
  selectedLocationId: string | null;
  locations: Location[];
  setLocation: (id: string) => void;
  loadLocations: () => Promise<void>;  // calls GET /api/v1/locations
}
```

- `loadLocations()` calls `GET /api/v1/locations` after login
- Persists `selectedLocationId` to localStorage
- Auto-selects first location if none persisted
- Layout header gets a location switcher dropdown (only shown if user has 2+ locations)

### API Client Updates (`lib/api.ts`)

- All tenant-scoped requests include `location_id` as a **query parameter** (matches existing backend convention — handlers read `r.URL.Query().Get("location_id")`)
- Error interceptor: 401 → clear auth, redirect to `/login`; 403 → show permission error via ErrorBanner
- No additional Zustand stores for domain data — React Query handles caching/loading/error
- Demo mode (`?demo=true`): bypass error interceptor 401 redirect when in demo mode to avoid kicking demo users to login

## 3. React Query Hooks (`src/hooks/`)

Each hook wraps `useQuery`/`useMutation` and returns `{ data, isLoading, error }`.
All location-scoped hooks accept `locationId: string | null` and set `enabled: !!locationId` so queries don't fire until a location is selected.

### `useFinancial.ts`
- `usePnL(locationId)` — `GET /api/v1/financial/pnl?location_id=X` — staleTime 30s
- `useAnomalies(locationId)` — `GET /api/v1/financial/anomalies?location_id=X` — staleTime 60s

### `useInventory.ts`
- `useUsage(locationId)` — `GET /api/v1/inventory/usage?location_id=X` — staleTime 30s
- `usePARStatus(locationId)` — `GET /api/v1/inventory/par?location_id=X` — staleTime 60s

### `useAlerts.ts`
- `useAlertQueue(locationId, opts?: { limit?: number })` — `GET /api/v1/alerts?location_id=X` — staleTime 10s. `limit` is applied client-side (slice). Severity filtering is client-side.
- `useAlertCount(locationId)` — `GET /api/v1/alerts/count?location_id=X` — staleTime 10s
- `useAcknowledgeAlert()` — `POST /api/v1/alerts/{id}/acknowledge` — mutation, invalidates queue + count
- `useResolveAlert()` — `POST /api/v1/alerts/{id}/resolve` — mutation, invalidates queue + count

### Anomaly Type

```typescript
interface Anomaly {
  id: string;
  metric: string;
  expected: number;
  actual: number;
  z_score: number;
  severity: 'critical' | 'warning' | 'info';
  detected_at: string;
}
```

### `useAdapters.ts` — DEFERRED

No adapter HTTP endpoints exist in the backend. AdaptersPage keeps its current mock data until a future SP adds adapter management APIs. No hooks built for adapters in SP1.

## 4. Page Wire-Up

Each page follows: `useQuery` hook → LoadingSpinner → ErrorBanner → data render with ui components.

### DashboardPage
- `usePnL(locationId)` → KPICard (revenue, COGS, gross margin)
- `useAlertCount(locationId)` → KPICard (active alerts)
- `useAlertQueue(locationId, { limit: 5 })` → Priority action queue list (top 5)
- Auto-refresh: refetchInterval 30s on P&L

### InventoryPage
- `useUsage(locationId)` → DataTable (theoretical usage)
- `usePARStatus(locationId)` → DataTable (PAR levels)
- EmptyState when no data

### FinancialPage
- `usePnL(locationId)` → KPICard cards + Recharts bar chart + DataTable (channel breakdown)
- `useAnomalies(locationId)` → Anomaly alert list with StatusBadge

### AlertsPage
- `useAlertQueue(locationId)` → Alert cards with StatusBadge
- `useAcknowledgeAlert()` / `useResolveAlert()` → real mutations
- Client-side severity filter
- `useAlertCount(locationId)` → badge in sidebar (via Layout)

### AdaptersPage — NO CHANGES (keeps mock data)

## 5. File Structure

```
src/
├── components/
│   ├── ui/
│   │   ├── DataTable.tsx
│   │   ├── KPICard.tsx
│   │   ├── StatusBadge.tsx
│   │   ├── Modal.tsx
│   │   ├── FormField.tsx
│   │   ├── EmptyState.tsx
│   │   ├── LoadingSpinner.tsx
│   │   └── ErrorBanner.tsx
│   ├── Layout.tsx           # Add location switcher + alert count badge
│   └── ProtectedRoute.tsx   # No changes
├── hooks/
│   ├── useFinancial.ts
│   ├── useInventory.ts
│   └── useAlerts.ts
├── stores/
│   ├── auth.ts              # No changes
│   └── location.ts          # New
├── lib/
│   └── api.ts               # Add location query param, error interceptor, Location types
├── pages/                   # Refactor 4 pages (dashboard, inventory, financial, alerts)
└── App.tsx                  # No changes
```

## 6. Backend Work Included in SP1

- **`GET /api/v1/locations`** — New handler returning user's accessible locations (from `user_location_access` join). Added to `internal/api/` with auth middleware.

## 7. Conventions

- Hooks in `hooks/` wrap React Query — return `{ data, isLoading, error }`
- All location-scoped hooks take `locationId: string | null`, disabled when null
- UI components are pure/presentational — no API calls
- Pages compose hooks + UI components
- All API calls go through `lib/api.ts`
- Location passed as `location_id` query parameter (matches existing backend)
- No external UI component libraries
