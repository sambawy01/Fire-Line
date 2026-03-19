# SP1: Foundation + Wire-Up Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build reusable UI component library, add location context, and wire all existing dashboard pages to real backend APIs.

**Architecture:** React Query hooks fetch from existing Go API endpoints. Zustand location store provides location context. Pure UI components in `src/components/ui/` are composed by pages. New `GET /api/v1/locations` backend endpoint returns user's accessible locations.

**Tech Stack:** React 19, TypeScript, Tailwind CSS 4, TanStack React Query, Zustand, Recharts, Lucide icons, Go 1.22+ backend with pgx/v5.

**Spec:** `docs/superpowers/specs/2026-03-19-sp1-foundation-wire-up-design.md`

---

### Task 1: Backend — GET /api/v1/locations Endpoint

**Files:**
- Create: `internal/api/location_handler.go`
- Create: `internal/api/location_handler_test.go`
- Modify: `cmd/fireline/main.go` (register route)

- [ ] **Step 1: Write the handler test**

Create `internal/api/location_handler_test.go`:

```go
package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/opsnerve/fireline/internal/api"
	"github.com/opsnerve/fireline/internal/auth"
	"github.com/opsnerve/fireline/internal/tenant"
)

func TestGetLocations_MissingTenant(t *testing.T) {
	handler := api.NewLocationHandler(nil)
	req := httptest.NewRequest("GET", "/api/v1/locations", nil)
	w := httptest.NewRecorder()

	handler.GetLocations(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestGetLocations_ReturnsLocations(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	pool := testPool(t) // helper that connects to test DB
	handler := api.NewLocationHandler(pool)

	// Create org + user + location via direct SQL for test setup
	ctx := context.Background()
	var orgID, userID, locID string
	err := pool.QueryRow(ctx,
		"INSERT INTO organizations (name, slug) VALUES ('Test Org', 'test-org') RETURNING org_id").Scan(&orgID)
	if err != nil {
		t.Fatal(err)
	}
	err = pool.QueryRow(ctx,
		"INSERT INTO users (org_id, email, password_hash, display_name, role) VALUES ($1, 'test@test.com', 'hash', 'Test', 'owner') RETURNING user_id", orgID).Scan(&userID)
	if err != nil {
		t.Fatal(err)
	}
	err = pool.QueryRow(ctx,
		"INSERT INTO locations (org_id, name) VALUES ($1, 'Downtown') RETURNING location_id", orgID).Scan(&locID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = pool.Exec(ctx,
		"INSERT INTO user_location_access (user_id, location_id, org_id) VALUES ($1, $2, $3)", userID, locID, orgID)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("GET", "/api/v1/locations", nil)
	ctx = tenant.WithOrgID(req.Context(), orgID)
	ctx = auth.WithUserID(ctx, userID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.GetLocations(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Locations []api.LocationResponse `json:"locations"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Locations) != 1 {
		t.Fatalf("expected 1 location, got %d", len(resp.Locations))
	}
	if resp.Locations[0].Name != "Downtown" {
		t.Fatalf("expected 'Downtown', got %q", resp.Locations[0].Name)
	}
}
```

- [ ] **Step 2: Write the handler**

Create `internal/api/location_handler.go`:

```go
package api

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/opsnerve/fireline/internal/auth"
	"github.com/opsnerve/fireline/internal/tenant"
)

type LocationResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	OrgID string `json:"org_id"`
}

type LocationHandler struct {
	pool *pgxpool.Pool
}

func NewLocationHandler(pool *pgxpool.Pool) *LocationHandler {
	return &LocationHandler{pool: pool}
}

func (h *LocationHandler) RegisterRoutes(mux *http.ServeMux, authMW func(http.Handler) http.Handler) {
	mux.Handle("GET /api/v1/locations", authMW(http.HandlerFunc(h.GetLocations)))
}

func (h *LocationHandler) GetLocations(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	userID := auth.UserIDFrom(r.Context())
	if userID == "" {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_USER", "no user context")
		return
	}

	rows, err := h.pool.Query(r.Context(), `
		SELECT l.location_id, l.name, l.org_id
		FROM locations l
		JOIN user_location_access ula ON ula.location_id = l.location_id
		WHERE ula.user_id = $1 AND ula.org_id = $2 AND l.status = 'active'
		ORDER BY l.name
	`, userID, orgID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "LOCATION_QUERY_ERROR", err.Error())
		return
	}
	defer rows.Close()

	locations := []LocationResponse{}
	for rows.Next() {
		var loc LocationResponse
		if err := rows.Scan(&loc.ID, &loc.Name, &loc.OrgID); err != nil {
			WriteError(w, http.StatusInternalServerError, "LOCATION_SCAN_ERROR", err.Error())
			return
		}
		locations = append(locations, loc)
	}

	WriteJSON(w, http.StatusOK, map[string]any{"locations": locations})
}
```

- [ ] **Step 3: Register route in main.go**

In `cmd/fireline/main.go`, after the alertHandler registration, add:

```go
locHandler := api.NewLocationHandler(pool.Raw())
locHandler.RegisterRoutes(mux, authMW)
```

- [ ] **Step 4: Run tests**

Run: `go test ./internal/api/ -run TestGetLocations_MissingTenant -v`
Expected: PASS (unit test, no DB needed)

- [ ] **Step 5: Verify server starts**

Run: `go build -o fireline ./cmd/fireline && echo "build ok"`
Expected: "build ok"

- [ ] **Step 6: Commit**

```bash
git add internal/api/location_handler.go internal/api/location_handler_test.go cmd/fireline/main.go
git commit -m "feat: add GET /api/v1/locations endpoint for user's accessible locations"
```

---

### Task 2: UI Component — LoadingSpinner

**Files:**
- Create: `web/src/components/ui/LoadingSpinner.tsx`

- [ ] **Step 1: Create LoadingSpinner**

```tsx
import { Loader2 } from 'lucide-react';

const sizes = { sm: 'h-4 w-4', md: 'h-6 w-6', lg: 'h-10 w-10' } as const;

interface LoadingSpinnerProps {
  size?: keyof typeof sizes;
  fullPage?: boolean;
}

export default function LoadingSpinner({ size = 'md', fullPage = false }: LoadingSpinnerProps) {
  const spinner = <Loader2 className={`${sizes[size]} animate-spin text-gray-400`} />;

  if (fullPage) {
    return (
      <div className="flex items-center justify-center min-h-[60vh]">
        {spinner}
      </div>
    );
  }
  return spinner;
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd web && npx tsc --noEmit`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add web/src/components/ui/LoadingSpinner.tsx
git commit -m "feat: add LoadingSpinner component"
```

---

### Task 3: UI Component — ErrorBanner

**Files:**
- Create: `web/src/components/ui/ErrorBanner.tsx`

- [ ] **Step 1: Create ErrorBanner**

```tsx
import { AlertCircle, X, RefreshCw } from 'lucide-react';

interface ErrorBannerProps {
  message: string;
  onDismiss?: () => void;
  retry?: () => void;
}

export default function ErrorBanner({ message, onDismiss, retry }: ErrorBannerProps) {
  return (
    <div className="rounded-lg border border-red-200 bg-red-50 p-4 flex items-start gap-3">
      <AlertCircle className="h-5 w-5 text-red-500 shrink-0 mt-0.5" />
      <div className="flex-1 min-w-0">
        <p className="text-sm text-red-700">{message}</p>
      </div>
      <div className="flex items-center gap-2 shrink-0">
        {retry && (
          <button onClick={retry} className="text-red-500 hover:text-red-700 transition-colors">
            <RefreshCw className="h-4 w-4" />
          </button>
        )}
        {onDismiss && (
          <button onClick={onDismiss} className="text-red-400 hover:text-red-600 transition-colors">
            <X className="h-4 w-4" />
          </button>
        )}
      </div>
    </div>
  );
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd web && npx tsc --noEmit`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add web/src/components/ui/ErrorBanner.tsx
git commit -m "feat: add ErrorBanner component"
```

---

### Task 4: UI Component — EmptyState

**Files:**
- Create: `web/src/components/ui/EmptyState.tsx`

- [ ] **Step 1: Create EmptyState**

```tsx
import type { LucideIcon } from 'lucide-react';
import { Inbox } from 'lucide-react';

interface EmptyStateProps {
  icon?: LucideIcon;
  title: string;
  description?: string;
  action?: { label: string; onClick: () => void };
}

export default function EmptyState({ icon: Icon = Inbox, title, description, action }: EmptyStateProps) {
  return (
    <div className="rounded-xl border border-dashed border-gray-300 bg-white py-16 text-center">
      <Icon className="mx-auto mb-3 h-10 w-10 text-gray-300" />
      <p className="text-lg font-medium text-gray-700">{title}</p>
      {description && <p className="mt-1 text-sm text-gray-500">{description}</p>}
      {action && (
        <button
          onClick={action.onClick}
          className="mt-4 rounded-lg bg-[#F97316] px-4 py-2 text-sm font-medium text-white hover:bg-[#EA580C] transition-colors"
        >
          {action.label}
        </button>
      )}
    </div>
  );
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd web && npx tsc --noEmit`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add web/src/components/ui/EmptyState.tsx
git commit -m "feat: add EmptyState component"
```

---

### Task 5: UI Component — StatusBadge

**Files:**
- Create: `web/src/components/ui/StatusBadge.tsx`

- [ ] **Step 1: Create StatusBadge**

```tsx
import type { ReactNode } from 'react';

const variants = {
  critical: 'bg-red-50 text-red-700 border-red-200',
  warning: 'bg-amber-50 text-amber-700 border-amber-200',
  info: 'bg-blue-50 text-blue-700 border-blue-200',
  success: 'bg-emerald-50 text-emerald-700 border-emerald-200',
  neutral: 'bg-gray-100 text-gray-600 border-gray-200',
} as const;

interface StatusBadgeProps {
  variant: keyof typeof variants;
  children: ReactNode;
}

export default function StatusBadge({ variant, children }: StatusBadgeProps) {
  return (
    <span className={`inline-flex items-center gap-1 rounded-full border px-2.5 py-0.5 text-xs font-semibold ${variants[variant]}`}>
      {children}
    </span>
  );
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd web && npx tsc --noEmit`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add web/src/components/ui/StatusBadge.tsx
git commit -m "feat: add StatusBadge component"
```

---

### Task 6: UI Component — KPICard

**Files:**
- Create: `web/src/components/ui/KPICard.tsx`

- [ ] **Step 1: Create KPICard**

```tsx
import type { LucideIcon } from 'lucide-react';

interface KPICardProps {
  label: string;
  value: string;
  icon: LucideIcon;
  iconColor?: string;
  bgTint?: string;
}

export default function KPICard({
  label,
  value,
  icon: Icon,
  iconColor = 'text-gray-600',
  bgTint = 'bg-gray-50',
}: KPICardProps) {
  return (
    <div className="bg-white rounded-xl border border-gray-200 p-5 flex items-start gap-4 shadow-sm">
      <div className={`${bgTint} p-3 rounded-lg`}>
        <Icon className={`h-6 w-6 ${iconColor}`} />
      </div>
      <div>
        <p className="text-sm text-gray-500">{label}</p>
        <p className="text-2xl font-bold text-gray-800 mt-0.5">{value}</p>
      </div>
    </div>
  );
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd web && npx tsc --noEmit`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add web/src/components/ui/KPICard.tsx
git commit -m "feat: add KPICard component"
```

---

### Task 7: UI Component — DataTable

**Files:**
- Create: `web/src/components/ui/DataTable.tsx`

- [ ] **Step 1: Create DataTable**

```tsx
import { useState, type ReactNode } from 'react';
import LoadingSpinner from './LoadingSpinner';
import EmptyState from './EmptyState';

export interface Column<T> {
  key: string;
  header: string;
  render?: (row: T) => ReactNode;
  sortable?: boolean;
  align?: 'left' | 'center' | 'right';
}

interface DataTableProps<T> {
  columns: Column<T>[];
  data: T[];
  keyExtractor: (row: T) => string;
  isLoading?: boolean;
  emptyTitle?: string;
  emptyDescription?: string;
}

export default function DataTable<T>({
  columns,
  data,
  keyExtractor,
  isLoading = false,
  emptyTitle = 'No data',
  emptyDescription,
}: DataTableProps<T>) {
  const [sortKey, setSortKey] = useState<string | null>(null);
  const [sortAsc, setSortAsc] = useState(true);

  if (isLoading) {
    return (
      <div className="bg-white rounded-xl border border-gray-200 shadow-sm p-12 flex justify-center">
        <LoadingSpinner size="lg" />
      </div>
    );
  }

  if (data.length === 0) {
    return <EmptyState title={emptyTitle} description={emptyDescription} />;
  }

  const sorted = sortKey
    ? [...data].sort((a, b) => {
        const aVal = (a as Record<string, unknown>)[sortKey];
        const bVal = (b as Record<string, unknown>)[sortKey];
        if (typeof aVal === 'number' && typeof bVal === 'number') {
          return sortAsc ? aVal - bVal : bVal - aVal;
        }
        return sortAsc
          ? String(aVal).localeCompare(String(bVal))
          : String(bVal).localeCompare(String(aVal));
      })
    : data;

  function handleSort(key: string) {
    if (sortKey === key) {
      setSortAsc(!sortAsc);
    } else {
      setSortKey(key);
      setSortAsc(true);
    }
  }

  const alignClass = (align?: string) =>
    align === 'right' ? 'text-right' : align === 'center' ? 'text-center' : 'text-left';

  return (
    <div className="bg-white rounded-xl border border-gray-200 shadow-sm overflow-hidden">
      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="bg-gray-50 text-gray-500 uppercase tracking-wider text-xs">
              {columns.map((col) => (
                <th
                  key={col.key}
                  className={`px-6 py-3 font-medium ${alignClass(col.align)} ${col.sortable ? 'cursor-pointer select-none hover:text-gray-700' : ''}`}
                  onClick={col.sortable ? () => handleSort(col.key) : undefined}
                >
                  {col.header}
                  {col.sortable && sortKey === col.key && (
                    <span className="ml-1">{sortAsc ? '↑' : '↓'}</span>
                  )}
                </th>
              ))}
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-100">
            {sorted.map((row) => (
              <tr key={keyExtractor(row)} className="hover:bg-gray-50 transition-colors">
                {columns.map((col) => (
                  <td key={col.key} className={`px-6 py-3 ${alignClass(col.align)} text-gray-700`}>
                    {col.render
                      ? col.render(row)
                      : String((row as Record<string, unknown>)[col.key] ?? '')}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd web && npx tsc --noEmit`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add web/src/components/ui/DataTable.tsx
git commit -m "feat: add DataTable component with sorting"
```

---

### Task 8: UI Components — Modal + FormField

**Files:**
- Create: `web/src/components/ui/Modal.tsx`
- Create: `web/src/components/ui/FormField.tsx`

- [ ] **Step 1: Create Modal**

```tsx
import { useEffect, useRef, type ReactNode } from 'react';
import { X } from 'lucide-react';

interface ModalProps {
  open: boolean;
  onClose: () => void;
  title: string;
  children: ReactNode;
  footer?: ReactNode;
}

export default function Modal({ open, onClose, title, children, footer }: ModalProps) {
  const overlayRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) return;
    const handleEsc = (e: KeyboardEvent) => { if (e.key === 'Escape') onClose(); };
    document.addEventListener('keydown', handleEsc);
    return () => document.removeEventListener('keydown', handleEsc);
  }, [open, onClose]);

  if (!open) return null;

  return (
    <div
      ref={overlayRef}
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
      onClick={(e) => { if (e.target === overlayRef.current) onClose(); }}
    >
      <div className="bg-white rounded-xl shadow-xl w-full max-w-lg mx-4">
        <div className="flex items-center justify-between px-6 py-4 border-b border-gray-200">
          <h3 className="text-lg font-semibold text-gray-800">{title}</h3>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600 transition-colors">
            <X className="h-5 w-5" />
          </button>
        </div>
        <div className="px-6 py-4">{children}</div>
        {footer && <div className="px-6 py-4 border-t border-gray-200 flex justify-end gap-3">{footer}</div>}
      </div>
    </div>
  );
}
```

- [ ] **Step 2: Create FormField**

```tsx
import type { ReactNode } from 'react';

interface FormFieldProps {
  label: string;
  error?: string;
  children: ReactNode;
}

export default function FormField({ label, error, children }: FormFieldProps) {
  return (
    <div className="space-y-1.5">
      <label className="block text-sm font-medium text-gray-700">{label}</label>
      {children}
      {error && <p className="text-sm text-red-600">{error}</p>}
    </div>
  );
}
```

- [ ] **Step 3: Verify they compile**

Run: `cd web && npx tsc --noEmit`
Expected: No errors

- [ ] **Step 4: Commit**

```bash
git add web/src/components/ui/Modal.tsx web/src/components/ui/FormField.tsx
git commit -m "feat: add Modal and FormField components"
```

---

### Task 9: API Client Updates + Anomaly Type

**Files:**
- Modify: `web/src/lib/api.ts`

- [ ] **Step 1: Add Location type, Anomaly type, location API, and error interceptor**

Add to `web/src/lib/api.ts` — after the `ApiError` class and `request` function, but before authApi:

```typescript
// Location
export interface Location {
  id: string;
  name: string;
  org_id: string;
}

export const locationApi = {
  getLocations() {
    return request<{ locations: Location[] }>('/locations');
  },
};
```

Replace the `any[]` in the anomalies return type. Change:
```typescript
  getAnomalies(locationId: string) {
    return request<{ anomalies: any[] }>(`/financial/anomalies?location_id=${locationId}`);
  },
```
to:
```typescript
  getAnomalies(locationId: string) {
    return request<{ anomalies: Anomaly[] }>(`/financial/anomalies?location_id=${locationId}`);
  },
```

Add the Anomaly interface after the `ChannelBreakdown` interface:

```typescript
export interface Anomaly {
  metric_name: string;
  current_value: number;
  mean: number;
  std_dev: number;
  z_score: number;
  severity: 'warning' | 'critical';
  detected_at: string;
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd web && npx tsc --noEmit`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add web/src/lib/api.ts
git commit -m "feat: add Location/Anomaly types and location API to client"
```

---

### Task 9b: API Client — Error Interceptor + Demo Mode Bypass

**Files:**
- Modify: `web/src/lib/api.ts`

- [ ] **Step 1: Add error interceptor to the `request` function**

Replace the existing `request` function in `web/src/lib/api.ts`:

```typescript
async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const token = localStorage.getItem('access_token');
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
  };

  const res = await fetch(`${API_BASE}${path}`, { ...options, headers });

  if (!res.ok) {
    const body = await res.json().catch(() => ({}));

    // 401: clear auth state and redirect to login (unless in demo mode)
    if (res.status === 401) {
      const isDemo = sessionStorage.getItem('fireline_demo') === 'true';
      if (!isDemo) {
        localStorage.removeItem('access_token');
        localStorage.removeItem('refresh_token');
        localStorage.removeItem('org_id');
        localStorage.removeItem('user_id');
        localStorage.removeItem('role');
        window.location.href = '/login';
      }
    }

    throw new ApiError(res.status, body.error?.code || 'UNKNOWN', body.error?.message || res.statusText);
  }

  return res.json();
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd web && npx tsc --noEmit`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add web/src/lib/api.ts
git commit -m "feat: add 401 error interceptor with demo mode bypass"
```

---

### Task 10: Location Store

**Files:**
- Create: `web/src/stores/location.ts`

- [ ] **Step 1: Create location store**

```typescript
import { create } from 'zustand';
import { locationApi, type Location } from '../lib/api';

interface LocationState {
  selectedLocationId: string | null;
  locations: Location[];
  isLoading: boolean;
  setLocation: (id: string) => void;
  loadLocations: () => Promise<void>;
  clear: () => void;
}

export const useLocationStore = create<LocationState>((set, get) => ({
  selectedLocationId: localStorage.getItem('selected_location_id'),
  locations: [],
  isLoading: false,

  setLocation: (id: string) => {
    localStorage.setItem('selected_location_id', id);
    set({ selectedLocationId: id });
  },

  loadLocations: async () => {
    set({ isLoading: true });
    try {
      const { locations } = await locationApi.getLocations();
      const current = get().selectedLocationId;
      const validSelection = locations.some((l) => l.id === current);
      set({
        locations,
        isLoading: false,
        selectedLocationId: validSelection ? current : locations[0]?.id ?? null,
      });
      // Persist the auto-selected location
      const finalId = validSelection ? current : locations[0]?.id ?? null;
      if (finalId) localStorage.setItem('selected_location_id', finalId);
    } catch {
      set({ isLoading: false });
    }
  },

  clear: () => {
    localStorage.removeItem('selected_location_id');
    set({ selectedLocationId: null, locations: [] });
  },
}));
```

- [ ] **Step 2: Verify it compiles**

Run: `cd web && npx tsc --noEmit`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add web/src/stores/location.ts
git commit -m "feat: add location store with persistence"
```

---

### Task 11: React Query Hooks — Financial

**Files:**
- Create: `web/src/hooks/useFinancial.ts`

- [ ] **Step 1: Create financial hooks**

```typescript
import { useQuery } from '@tanstack/react-query';
import { financialApi } from '../lib/api';

export function usePnL(locationId: string | null) {
  return useQuery({
    queryKey: ['financial', 'pnl', locationId],
    queryFn: () => financialApi.getPnL(locationId!),
    enabled: !!locationId,
    staleTime: 30_000,
    refetchInterval: 30_000,
  });
}

export function useAnomalies(locationId: string | null) {
  return useQuery({
    queryKey: ['financial', 'anomalies', locationId],
    queryFn: () => financialApi.getAnomalies(locationId!),
    enabled: !!locationId,
    staleTime: 60_000,
  });
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd web && npx tsc --noEmit`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add web/src/hooks/useFinancial.ts
git commit -m "feat: add financial React Query hooks"
```

---

### Task 12: React Query Hooks — Inventory

**Files:**
- Create: `web/src/hooks/useInventory.ts`

- [ ] **Step 1: Create inventory hooks**

```typescript
import { useQuery } from '@tanstack/react-query';
import { inventoryApi } from '../lib/api';

export function useUsage(locationId: string | null) {
  return useQuery({
    queryKey: ['inventory', 'usage', locationId],
    queryFn: () => inventoryApi.getUsage(locationId!),
    enabled: !!locationId,
    staleTime: 30_000,
  });
}

export function usePARStatus(locationId: string | null) {
  return useQuery({
    queryKey: ['inventory', 'par', locationId],
    queryFn: () => inventoryApi.getPARStatus(locationId!),
    enabled: !!locationId,
    staleTime: 60_000,
  });
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd web && npx tsc --noEmit`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add web/src/hooks/useInventory.ts
git commit -m "feat: add inventory React Query hooks"
```

---

### Task 13: React Query Hooks — Alerts

**Files:**
- Create: `web/src/hooks/useAlerts.ts`

- [ ] **Step 1: Create alert hooks**

```typescript
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { alertsApi } from '../lib/api';

export function useAlertQueue(locationId: string | null, opts?: { limit?: number }) {
  return useQuery({
    queryKey: ['alerts', 'queue', locationId],
    queryFn: async () => {
      const { alerts } = await alertsApi.getQueue(locationId ?? undefined);
      return opts?.limit ? alerts.slice(0, opts.limit) : alerts;
    },
    enabled: !!locationId,
    staleTime: 10_000,
  });
}

export function useAlertCount(locationId: string | null) {
  return useQuery({
    queryKey: ['alerts', 'count', locationId],
    queryFn: () => alertsApi.getCount(),
    enabled: !!locationId,
    staleTime: 10_000,
  });
}

export function useAcknowledgeAlert() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (alertId: string) => alertsApi.acknowledge(alertId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['alerts'] });
    },
  });
}

export function useResolveAlert() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (alertId: string) => alertsApi.resolve(alertId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['alerts'] });
    },
  });
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd web && npx tsc --noEmit`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add web/src/hooks/useAlerts.ts
git commit -m "feat: add alerts React Query hooks with mutations"
```

---

### Task 14: Layout — Location Switcher + Alert Badge

**Files:**
- Modify: `web/src/components/Layout.tsx`

- [ ] **Step 1: Add location loading on mount and switcher to Layout**

Replace the entire `Layout.tsx` content with:

```tsx
import { useEffect } from 'react';
import { Outlet, NavLink, useNavigate } from 'react-router-dom';
import {
  LayoutDashboard,
  Package,
  DollarSign,
  Bell,
  Plug,
  LogOut,
  Flame,
  User,
  MapPin,
} from 'lucide-react';
import { useAuthStore } from '../stores/auth';
import { useLocationStore } from '../stores/location';
import { useAlertCount } from '../hooks/useAlerts';

const navItems = [
  { to: '/', label: 'Dashboard', icon: LayoutDashboard },
  { to: '/inventory', label: 'Inventory', icon: Package },
  { to: '/financial', label: 'Financial', icon: DollarSign },
  { to: '/alerts', label: 'Alerts', icon: Bell, showBadge: true },
  { to: '/adapters', label: 'Adapters', icon: Plug },
];

export default function Layout() {
  const navigate = useNavigate();
  const logout = useAuthStore((s) => s.logout);
  const role = useAuthStore((s) => s.role);
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);

  const { locations, selectedLocationId, setLocation, loadLocations } = useLocationStore();
  const { data: alertCount } = useAlertCount(selectedLocationId);

  useEffect(() => {
    if (isAuthenticated) {
      loadLocations();
    }
  }, [isAuthenticated, loadLocations]);

  const handleLogout = () => {
    useLocationStore.getState().clear();
    logout();
    navigate('/login');
  };

  return (
    <div className="flex h-screen bg-gray-100">
      {/* Sidebar */}
      <aside className="hidden md:flex md:flex-col md:w-64 md:fixed md:inset-y-0 bg-[#1E293B] text-white">
        {/* Logo */}
        <div className="flex items-center gap-3 px-6 py-5 border-b border-white/10">
          <Flame className="h-8 w-8 text-[#F97316]" />
          <div>
            <h1 className="text-lg font-bold tracking-tight">FireLine</h1>
            <p className="text-xs text-gray-400">by OpsNerve</p>
          </div>
        </div>

        {/* Location Switcher */}
        {locations.length > 1 && (
          <div className="px-3 py-3 border-b border-white/10">
            <div className="flex items-center gap-2 px-3 mb-1.5">
              <MapPin className="h-3.5 w-3.5 text-gray-400" />
              <span className="text-xs text-gray-400 uppercase tracking-wider">Location</span>
            </div>
            <select
              value={selectedLocationId ?? ''}
              onChange={(e) => setLocation(e.target.value)}
              className="w-full bg-white/10 text-white text-sm rounded-md px-3 py-1.5 border border-white/10 focus:outline-none focus:ring-1 focus:ring-[#F97316]"
            >
              {locations.map((loc) => (
                <option key={loc.id} value={loc.id} className="bg-[#1E293B]">
                  {loc.name}
                </option>
              ))}
            </select>
          </div>
        )}

        {/* Navigation */}
        <nav className="flex-1 px-3 py-4 space-y-1 overflow-y-auto">
          {navItems.map(({ to, label, icon: Icon, showBadge }) => (
            <NavLink
              key={to}
              to={to}
              className={({ isActive }) =>
                `flex items-center gap-3 px-3 py-2.5 rounded-md text-sm font-medium transition-colors ${
                  isActive
                    ? 'border-l-[3px] border-[#F97316] text-[#F97316] bg-white/5'
                    : 'border-l-[3px] border-transparent text-gray-300 hover:text-white hover:bg-white/5'
                }`
              }
            >
              <Icon className="h-5 w-5 shrink-0" />
              {label}
              {showBadge && alertCount?.count != null && alertCount.count > 0 && (
                <span className="ml-auto inline-flex items-center justify-center rounded-full bg-[#F97316] px-2 py-0.5 text-xs font-bold text-white min-w-[20px]">
                  {alertCount.count}
                </span>
              )}
            </NavLink>
          ))}
        </nav>

        {/* Logout */}
        <div className="px-3 py-4 border-t border-white/10">
          <button
            onClick={handleLogout}
            className="flex items-center gap-3 w-full px-3 py-2.5 rounded-md text-sm font-medium text-gray-300 hover:text-white hover:bg-white/5 transition-colors"
          >
            <LogOut className="h-5 w-5 shrink-0" />
            Logout
          </button>
        </div>
      </aside>

      {/* Main content */}
      <div className="flex-1 md:ml-64 flex flex-col min-h-screen">
        {/* Top header */}
        <header className="sticky top-0 z-10 bg-white border-b border-gray-200 px-6 py-4 flex items-center justify-between">
          <h2 className="text-xl font-semibold text-gray-800">
            {locations.length === 1 ? locations[0].name : 'FireLine'}
          </h2>
          <div className="flex items-center gap-3">
            <div className="text-right hidden sm:block">
              <p className="text-sm font-medium text-gray-700">
                {role ?? 'Operator'}
              </p>
              <p className="text-xs text-gray-400">Restaurant Manager</p>
            </div>
            <div className="h-9 w-9 rounded-full bg-[#1E293B] flex items-center justify-center">
              <User className="h-5 w-5 text-white" />
            </div>
          </div>
        </header>

        {/* Page content */}
        <main className="flex-1 p-6 overflow-y-auto">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd web && npx tsc --noEmit`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add web/src/components/Layout.tsx
git commit -m "feat: add location switcher and alert count badge to Layout"
```

---

### Task 15: Wire DashboardPage to Real APIs

**Files:**
- Modify: `web/src/pages/DashboardPage.tsx`

- [ ] **Step 1: Rewrite DashboardPage with real data**

Replace entire file:

```tsx
import {
  DollarSign,
  TrendingDown,
  Percent,
  AlertTriangle,
  AlertCircle,
  Info,
} from 'lucide-react';
import { useLocationStore } from '../stores/location';
import { usePnL } from '../hooks/useFinancial';
import { useAlertQueue, useAlertCount } from '../hooks/useAlerts';
import KPICard from '../components/ui/KPICard';
import LoadingSpinner from '../components/ui/LoadingSpinner';
import ErrorBanner from '../components/ui/ErrorBanner';
import StatusBadge from '../components/ui/StatusBadge';

function cents(v: number): string {
  return `$${(v / 100).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
}

const severityIcon: Record<string, typeof AlertCircle> = {
  critical: AlertCircle,
  warning: AlertTriangle,
  info: Info,
};

export default function DashboardPage() {
  const locationId = useLocationStore((s) => s.selectedLocationId);
  const { data: pnl, isLoading: pnlLoading, error: pnlError, refetch: refetchPnl } = usePnL(locationId);
  const { data: alertCount } = useAlertCount(locationId);
  const { data: topAlerts, isLoading: alertsLoading } = useAlertQueue(locationId, { limit: 5 });

  if (!locationId) {
    return <LoadingSpinner fullPage />;
  }

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold text-gray-800">Dashboard</h1>
        <p className="text-sm text-gray-500 mt-1">Today's operational snapshot</p>
      </div>

      {pnlError && (
        <ErrorBanner
          message={pnlError instanceof Error ? pnlError.message : 'Failed to load financial data'}
          retry={() => refetchPnl()}
        />
      )}

      {/* KPI Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-5">
        {pnlLoading ? (
          <div className="col-span-full flex justify-center py-8">
            <LoadingSpinner />
          </div>
        ) : pnl ? (
          <>
            <KPICard label="Revenue (Today)" value={cents(pnl.net_revenue)} icon={DollarSign} iconColor="text-emerald-600" bgTint="bg-emerald-50" />
            <KPICard label="COGS" value={cents(pnl.cogs)} icon={TrendingDown} iconColor="text-red-600" bgTint="bg-red-50" />
            <KPICard label="Gross Margin %" value={`${pnl.gross_margin.toFixed(1)}%`} icon={Percent} iconColor="text-blue-600" bgTint="bg-blue-50" />
            <KPICard label="Active Alerts" value={String(alertCount?.count ?? 0)} icon={AlertTriangle} iconColor="text-orange-600" bgTint="bg-orange-50" />
          </>
        ) : null}
      </div>

      {/* Priority Action Queue */}
      <div>
        <h2 className="text-lg font-semibold text-gray-800 mb-4">Priority Action Queue</h2>
        {alertsLoading ? (
          <div className="flex justify-center py-8"><LoadingSpinner /></div>
        ) : topAlerts && topAlerts.length > 0 ? (
          <div className="space-y-3">
            {topAlerts.map((alert) => {
              const SevIcon = severityIcon[alert.severity] ?? Info;
              return (
                <div
                  key={alert.alert_id}
                  className={`bg-white rounded-lg border border-gray-200 border-l-4 p-4 flex items-start gap-3 shadow-sm ${
                    alert.severity === 'critical' ? 'border-l-red-500' :
                    alert.severity === 'warning' ? 'border-l-yellow-500' : 'border-l-blue-500'
                  }`}
                >
                  <SevIcon className={`h-5 w-5 mt-0.5 shrink-0 ${
                    alert.severity === 'critical' ? 'text-red-700' :
                    alert.severity === 'warning' ? 'text-yellow-700' : 'text-blue-700'
                  }`} />
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-1">
                      <p className="font-medium text-gray-800">{alert.title}</p>
                      <StatusBadge variant={alert.severity}>{alert.severity}</StatusBadge>
                    </div>
                    <p className="text-sm text-gray-500">{alert.description}</p>
                  </div>
                </div>
              );
            })}
          </div>
        ) : (
          <div className="text-center py-8 text-gray-400">No alerts — all clear.</div>
        )}
      </div>
    </div>
  );
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd web && npx tsc --noEmit`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add web/src/pages/DashboardPage.tsx
git commit -m "feat: wire DashboardPage to real P&L and alerts APIs"
```

---

### Task 16: Wire InventoryPage to Real APIs

**Files:**
- Modify: `web/src/pages/InventoryPage.tsx`

- [ ] **Step 1: Rewrite InventoryPage with real data**

Replace entire file:

```tsx
import { useLocationStore } from '../stores/location';
import { useUsage, usePARStatus } from '../hooks/useInventory';
import DataTable, { type Column } from '../components/ui/DataTable';
import StatusBadge from '../components/ui/StatusBadge';
import ErrorBanner from '../components/ui/ErrorBanner';
import LoadingSpinner from '../components/ui/LoadingSpinner';
import type { TheoreticalUsage, PARStatus } from '../lib/api';

function cents(v: number): string {
  return `$${(v / 100).toFixed(2)}`;
}

const usageColumns: Column<TheoreticalUsage>[] = [
  { key: 'ingredient_name', header: 'Ingredient', sortable: true },
  { key: 'total_used', header: 'Qty Used', align: 'right', sortable: true, render: (r) => r.total_used.toFixed(2) },
  { key: 'unit', header: 'Unit' },
  { key: 'cost_per_unit', header: 'Cost/Unit', align: 'right', render: (r) => cents(r.cost_per_unit) },
  { key: 'total_cost', header: 'Total Cost', align: 'right', sortable: true, render: (r) => cents(r.total_cost) },
];

function parStatus(row: PARStatus): 'Critical' | 'Low' | 'OK' {
  if (row.needs_reorder && row.current_level <= row.reorder_point) return 'Critical';
  if (row.needs_reorder) return 'Low';
  return 'OK';
}

const parColumns: Column<PARStatus>[] = [
  { key: 'ingredient_name', header: 'Ingredient', sortable: true },
  { key: 'current_level', header: 'Current', align: 'right', render: (r) => r.current_level.toFixed(1) },
  { key: 'par_level', header: 'PAR Level', align: 'right', render: (r) => r.par_level.toFixed(1) },
  { key: 'reorder_point', header: 'Reorder Point', align: 'right', render: (r) => r.reorder_point.toFixed(1) },
  {
    key: 'status',
    header: 'Status',
    align: 'center',
    render: (r) => {
      const s = parStatus(r);
      const variant = s === 'Critical' ? 'critical' : s === 'Low' ? 'warning' : 'success';
      return <StatusBadge variant={variant}>{s}</StatusBadge>;
    },
  },
];

export default function InventoryPage() {
  const locationId = useLocationStore((s) => s.selectedLocationId);
  const { data: usageData, isLoading: usageLoading, error: usageError, refetch: refetchUsage } = useUsage(locationId);
  const { data: parData, isLoading: parLoading, error: parError, refetch: refetchPar } = usePARStatus(locationId);

  if (!locationId) return <LoadingSpinner fullPage />;

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold text-gray-800">Inventory Intelligence</h1>
        <p className="text-sm text-gray-500 mt-1">Theoretical usage and PAR status overview</p>
      </div>

      {usageError && (
        <ErrorBanner
          message={usageError instanceof Error ? usageError.message : 'Failed to load usage data'}
          retry={() => refetchUsage()}
        />
      )}

      <div>
        <div className="mb-3">
          <h2 className="text-lg font-semibold text-gray-800">Theoretical Usage</h2>
          <p className="text-xs text-gray-500 mt-0.5">Based on today's sales mix</p>
        </div>
        <DataTable
          columns={usageColumns}
          data={usageData?.usage ?? []}
          keyExtractor={(r) => r.ingredient_id}
          isLoading={usageLoading}
          emptyTitle="No usage data"
          emptyDescription="No orders synced yet for this location."
        />
      </div>

      {parError && (
        <ErrorBanner
          message={parError instanceof Error ? parError.message : 'Failed to load PAR data'}
          retry={() => refetchPar()}
        />
      )}

      <div>
        <div className="mb-3">
          <h2 className="text-lg font-semibold text-gray-800">PAR Status</h2>
          <p className="text-xs text-gray-500 mt-0.5">Current stock vs. target levels</p>
        </div>
        <DataTable
          columns={parColumns}
          data={parData?.par_status ?? []}
          keyExtractor={(r) => r.ingredient_id}
          isLoading={parLoading}
          emptyTitle="No PAR data"
          emptyDescription="Configure PAR levels for your ingredients to see status here."
        />
      </div>
    </div>
  );
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd web && npx tsc --noEmit`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add web/src/pages/InventoryPage.tsx
git commit -m "feat: wire InventoryPage to real usage and PAR APIs"
```

---

### Task 17: Wire FinancialPage to Real APIs

**Files:**
- Modify: `web/src/pages/FinancialPage.tsx`

- [ ] **Step 1: Rewrite FinancialPage with real data**

Replace entire file:

```tsx
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from 'recharts';
import { useLocationStore } from '../stores/location';
import { usePnL, useAnomalies } from '../hooks/useFinancial';
import KPICard from '../components/ui/KPICard';
import DataTable, { type Column } from '../components/ui/DataTable';
import StatusBadge from '../components/ui/StatusBadge';
import LoadingSpinner from '../components/ui/LoadingSpinner';
import ErrorBanner from '../components/ui/ErrorBanner';
import {
  DollarSign,
  TrendingDown,
  TrendingUp,
  Percent,
  Receipt,
} from 'lucide-react';
import type { ChannelBreakdown, Anomaly } from '../lib/api';

function cents(v: number): string {
  return `$${(v / 100).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
}

const CHANNEL_LABELS: Record<string, string> = {
  dine_in: 'Dine-in',
  takeout: 'Takeout',
  delivery: 'Delivery',
  drive_thru: 'Drive-thru',
};

const channelColumns: Column<ChannelBreakdown>[] = [
  { key: 'channel', header: 'Channel', render: (r) => CHANNEL_LABELS[r.channel] ?? r.channel },
  { key: 'revenue', header: 'Revenue', align: 'right', sortable: true, render: (r) => cents(r.revenue) },
  { key: 'cogs', header: 'COGS', align: 'right', render: (r) => cents(r.cogs) },
  { key: 'gross_margin', header: 'Margin %', align: 'right', sortable: true, render: (r) => `${r.gross_margin.toFixed(1)}%` },
  { key: 'check_count', header: 'Checks', align: 'right', sortable: true },
  { key: 'avg_check_size', header: 'Avg Check', align: 'right', render: (r) => cents(r.avg_check_size) },
];

export default function FinancialPage() {
  const locationId = useLocationStore((s) => s.selectedLocationId);
  const { data: pnl, isLoading: pnlLoading, error: pnlError, refetch: refetchPnl } = usePnL(locationId);
  const { data: anomalyData, isLoading: anomLoading } = useAnomalies(locationId);

  if (!locationId) return <LoadingSpinner fullPage />;

  const chartData = pnl?.by_channel?.map((ch) => ({
    channel: CHANNEL_LABELS[ch.channel] ?? ch.channel,
    revenue: ch.revenue / 100,
  })) ?? [];

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold text-gray-800">Financial Intelligence</h1>
        <p className="text-sm text-gray-500 mt-1">P&L overview and channel performance</p>
      </div>

      {pnlError && (
        <ErrorBanner
          message={pnlError instanceof Error ? pnlError.message : 'Failed to load financial data'}
          retry={() => refetchPnl()}
        />
      )}

      {/* P&L Summary Cards */}
      {pnlLoading ? (
        <div className="flex justify-center py-8"><LoadingSpinner /></div>
      ) : pnl ? (
        <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-4">
          <KPICard label="Gross Revenue" value={cents(pnl.gross_revenue)} icon={DollarSign} iconColor="text-emerald-600" bgTint="bg-emerald-50" />
          <KPICard label="Net Revenue" value={cents(pnl.net_revenue)} icon={TrendingUp} iconColor="text-blue-600" bgTint="bg-blue-50" />
          <KPICard label="COGS" value={cents(pnl.cogs)} icon={TrendingDown} iconColor="text-red-600" bgTint="bg-red-50" />
          <KPICard label="Gross Profit" value={cents(pnl.gross_profit)} icon={DollarSign} iconColor="text-emerald-600" bgTint="bg-emerald-50" />
          <KPICard label="Margin %" value={`${pnl.gross_margin.toFixed(1)}%`} icon={Percent} iconColor="text-purple-600" bgTint="bg-purple-50" />
        </div>
      ) : null}

      {/* Channel Revenue Chart */}
      {chartData.length > 0 && (
        <div className="bg-white rounded-xl border border-gray-200 p-6 shadow-sm">
          <h2 className="text-lg font-semibold text-gray-800 mb-4">Revenue by Channel</h2>
          <div className="h-72">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={chartData} margin={{ top: 5, right: 20, left: 0, bottom: 5 }}>
                <CartesianGrid strokeDasharray="3 3" stroke="#E5E7EB" />
                <XAxis dataKey="channel" tick={{ fontSize: 13 }} />
                <YAxis tick={{ fontSize: 13 }} tickFormatter={(v: number) => `$${v.toLocaleString()}`} />
                <Tooltip
                  formatter={(value) => [`$${Number(value).toLocaleString()}`, 'Revenue']}
                  contentStyle={{ borderRadius: '8px', border: '1px solid #E5E7EB', fontSize: '13px' }}
                />
                <Bar dataKey="revenue" fill="#F97316" radius={[6, 6, 0, 0]} />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </div>
      )}

      {/* Channel Breakdown Table */}
      <div>
        <h2 className="text-lg font-semibold text-gray-800 mb-3">Channel Breakdown</h2>
        <DataTable
          columns={channelColumns}
          data={pnl?.by_channel ?? []}
          keyExtractor={(r) => r.channel}
          isLoading={pnlLoading}
          emptyTitle="No channel data"
        />
      </div>

      {/* Anomalies */}
      {!anomLoading && anomalyData?.anomalies && anomalyData.anomalies.length > 0 && (
        <div>
          <h2 className="text-lg font-semibold text-gray-800 mb-3">Anomalies Detected</h2>
          <div className="space-y-3">
            {anomalyData.anomalies.map((a: Anomaly, i: number) => (
              <div key={i} className="bg-white rounded-lg border border-gray-200 p-4 flex items-start gap-3 shadow-sm">
                <StatusBadge variant={a.severity === 'critical' ? 'critical' : 'warning'}>
                  {a.severity} (z={a.z_score.toFixed(1)})
                </StatusBadge>
                <div>
                  <p className="font-medium text-gray-800">{a.metric_name.replace(/_/g, ' ')}</p>
                  <p className="text-sm text-gray-500">
                    Expected ~{a.mean.toFixed(0)} ± {a.std_dev.toFixed(0)}, got {a.current_value.toFixed(0)}
                  </p>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd web && npx tsc --noEmit`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add web/src/pages/FinancialPage.tsx
git commit -m "feat: wire FinancialPage to real P&L and anomaly APIs"
```

---

### Task 18: Wire AlertsPage to Real APIs

**Files:**
- Modify: `web/src/pages/AlertsPage.tsx`

- [ ] **Step 1: Rewrite AlertsPage with real data**

Replace entire file:

```tsx
import { useState, useMemo } from 'react';
import {
  AlertTriangle,
  AlertCircle,
  Info,
  CheckCircle,
  Shield,
  Clock,
  Filter,
} from 'lucide-react';
import { useLocationStore } from '../stores/location';
import { useAlertQueue, useAcknowledgeAlert, useResolveAlert } from '../hooks/useAlerts';
import StatusBadge from '../components/ui/StatusBadge';
import LoadingSpinner from '../components/ui/LoadingSpinner';
import ErrorBanner from '../components/ui/ErrorBanner';
import type { Alert } from '../lib/api';

type Severity = 'critical' | 'warning' | 'info';
type FilterType = 'all' | Severity;

const SEVERITY_CONFIG: Record<Severity, { label: string; icon: typeof AlertCircle }> = {
  critical: { label: 'Critical', icon: AlertCircle },
  warning: { label: 'Warning', icon: AlertTriangle },
  info: { label: 'Info', icon: Info },
};

const MODULE_LABELS: Record<string, string> = {
  inventory: 'Inventory',
  financial: 'Financial',
  adapter: 'Adapter',
};

function formatTimestamp(iso: string): string {
  return new Date(iso).toLocaleString('en-US', {
    month: 'short', day: 'numeric', hour: 'numeric', minute: '2-digit', hour12: true,
  });
}

const filters: { key: FilterType; label: string }[] = [
  { key: 'all', label: 'All' },
  { key: 'critical', label: 'Critical' },
  { key: 'warning', label: 'Warning' },
  { key: 'info', label: 'Info' },
];

export default function AlertsPage() {
  const locationId = useLocationStore((s) => s.selectedLocationId);
  const { data: alerts, isLoading, error, refetch } = useAlertQueue(locationId);
  const ackMutation = useAcknowledgeAlert();
  const resolveMutation = useResolveAlert();
  const [activeFilter, setActiveFilter] = useState<FilterType>('all');

  const filteredAlerts = useMemo(
    () => {
      if (!alerts) return [];
      return activeFilter === 'all' ? alerts : alerts.filter((a) => a.severity === activeFilter);
    },
    [alerts, activeFilter],
  );

  const activeCount = alerts?.filter((a) => a.status === 'active').length ?? 0;

  if (!locationId) return <LoadingSpinner fullPage />;

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="mx-auto max-w-4xl px-4 py-8 sm:px-6 lg:px-8">
        {/* Header */}
        <div className="mb-6 flex items-center gap-3">
          <Shield className="h-7 w-7 text-gray-900" />
          <h1 className="text-2xl font-bold text-gray-900">Priority Action Queue</h1>
          <span className="inline-flex items-center rounded-full bg-[#F97316] px-3 py-0.5 text-sm font-semibold text-white">
            {activeCount} active
          </span>
        </div>

        {error && (
          <div className="mb-6">
            <ErrorBanner
              message={error instanceof Error ? error.message : 'Failed to load alerts'}
              retry={() => refetch()}
            />
          </div>
        )}

        {/* Filters */}
        <div className="mb-6 flex flex-wrap items-center gap-2">
          <Filter className="h-4 w-4 text-gray-500" />
          {filters.map((f) => (
            <button
              key={f.key}
              onClick={() => setActiveFilter(f.key)}
              className={`rounded-lg px-4 py-1.5 text-sm font-medium transition-colors ${
                activeFilter === f.key
                  ? 'bg-gray-900 text-white'
                  : 'bg-white text-gray-600 ring-1 ring-gray-200 hover:bg-gray-100'
              }`}
            >
              {f.label}
            </button>
          ))}
        </div>

        {/* Alert Cards */}
        {isLoading ? (
          <LoadingSpinner fullPage />
        ) : filteredAlerts.length === 0 ? (
          <div className="rounded-xl border border-dashed border-gray-300 bg-white py-16 text-center">
            <CheckCircle className="mx-auto mb-3 h-10 w-10 text-green-400" />
            <p className="text-lg font-medium text-gray-700">No alerts match this filter</p>
            <p className="mt-1 text-sm text-gray-500">All clear — nothing needs your attention right now.</p>
          </div>
        ) : (
          <ul className="space-y-4">
            {filteredAlerts.map((alert: Alert) => {
              const config = SEVERITY_CONFIG[alert.severity] ?? SEVERITY_CONFIG.info;
              const SeverityIcon = config.icon;
              const isAcked = alert.status === 'acknowledged' || alert.status === 'resolved';
              const isResolved = alert.status === 'resolved';

              return (
                <li
                  key={alert.alert_id}
                  className={`rounded-xl border bg-white shadow-sm transition-opacity ${isResolved ? 'opacity-50' : ''}`}
                >
                  <div className="p-5">
                    <div className="mb-3 flex flex-wrap items-center gap-2">
                      <StatusBadge variant={alert.severity}>
                        <SeverityIcon className="h-3.5 w-3.5" />
                        {config.label}
                      </StatusBadge>
                      <StatusBadge variant="neutral">
                        {MODULE_LABELS[alert.module] ?? alert.module}
                      </StatusBadge>
                      <span className="ml-auto flex items-center gap-1 text-xs text-gray-400">
                        <Clock className="h-3.5 w-3.5" />
                        {formatTimestamp(alert.created_at)}
                      </span>
                    </div>

                    <h3 className="text-base font-semibold text-gray-900">{alert.title}</h3>
                    <p className="mt-1 text-sm leading-relaxed text-gray-600">{alert.description}</p>

                    <div className="mt-4 flex items-center gap-3">
                      <button
                        disabled={isAcked || ackMutation.isPending}
                        onClick={() => ackMutation.mutate(alert.alert_id)}
                        className={`rounded-lg px-4 py-1.5 text-sm font-medium transition-colors ${
                          isAcked
                            ? 'cursor-default bg-gray-100 text-gray-400'
                            : 'bg-[#F97316] text-white hover:bg-[#EA580C]'
                        }`}
                      >
                        {isAcked ? 'Acknowledged' : 'Acknowledge'}
                      </button>
                      <button
                        disabled={isResolved || resolveMutation.isPending}
                        onClick={() => resolveMutation.mutate(alert.alert_id)}
                        className={`rounded-lg px-4 py-1.5 text-sm font-medium transition-colors ${
                          isResolved
                            ? 'cursor-default bg-gray-100 text-gray-400'
                            : 'bg-white text-gray-700 ring-1 ring-gray-200 hover:bg-gray-50'
                        }`}
                      >
                        {isResolved ? 'Resolved' : 'Resolve'}
                      </button>
                    </div>
                  </div>
                </li>
              );
            })}
          </ul>
        )}
      </div>
    </div>
  );
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd web && npx tsc --noEmit`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add web/src/pages/AlertsPage.tsx
git commit -m "feat: wire AlertsPage to real alerts API with acknowledge/resolve"
```

---

### Task 19: Full Build + Smoke Test

**Files:** None (verification only)

- [ ] **Step 1: Run TypeScript type check**

Run: `cd web && npx tsc --noEmit`
Expected: No errors

- [ ] **Step 2: Run Go tests**

Run: `cd /Users/bistrocloud/Documents/AI_Restaurant_System/fireline && go test ./internal/api/ -run TestGetLocations_MissingTenant -v`
Expected: PASS

- [ ] **Step 3: Build Go server**

Run: `go build -o fireline ./cmd/fireline`
Expected: No errors

- [ ] **Step 4: Build frontend**

Run: `cd web && npm run build`
Expected: Build succeeds, output in `web/dist/`

- [ ] **Step 5: Smoke test — start server and hit health endpoint**

Run: `./fireline &` then `curl http://localhost:8080/health/live`
Expected: `{"status":"ok"}`

- [ ] **Step 6: Smoke test — start frontend dev server**

Run: `cd web && npm run dev &` then `curl -s http://localhost:3000/ | head -5`
Expected: HTML response with React app

- [ ] **Step 7: Final commit**

```bash
git add -A
git commit -m "feat: SP1 Foundation + Wire-Up complete — all pages wired to real APIs"
```
