import { useState } from 'react';
import { useLocationStore } from '../stores/location';
import { useUsage, usePARStatus, useVariances, useExpiry } from '../hooks/useInventory';
import { usePARBreaches } from '../hooks/usePurchaseOrders';
import DataTable from '../components/ui/DataTable';
import type { Column } from '../components/ui/DataTable';
import StatusBadge from '../components/ui/StatusBadge';
import ErrorBanner from '../components/ui/ErrorBanner';
import LoadingSpinner from '../components/ui/LoadingSpinner';
import type { TheoreticalUsage, PARStatus, CountVariance, ExpiryItem } from '../lib/api';

// ─── Helpers ────────────────────────────────────────────────────────────────

function cents(v: number): string {
  return `EGP ${(v / 100).toFixed(2)}`;
}

function varianceDollars(v: number): JSX.Element {
  const formatted = `${v < 0 ? '-' : '+'}EGP ${Math.abs(v / 100).toFixed(2)}`;
  const color = v > 0 ? 'text-red-600' : v < 0 ? 'text-green-600' : 'text-slate-300';
  return <span className={color}>{formatted}</span>;
}

// ─── Cause probability stacked bar ──────────────────────────────────────────

const CAUSE_COLORS: Record<string, string> = {
  unrecorded_waste: '#ef4444',
  portioning: '#f97316',
  measurement_error: '#6b7280',
  recipe_error: '#8b5cf6',
  theft_signal: '#dc2626',
  other: '#9ca3af',
};

function causeLabel(key: string): string {
  return key.replace(/_/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase());
}

function CauseProbabilityBar({ probs }: { probs: Record<string, number> }) {
  const entries = Object.entries(probs).filter(([, v]) => v > 0);
  if (entries.length === 0) return <span className="text-slate-300 text-xs">—</span>;

  return (
    <div className="flex h-4 w-full min-w-[100px] rounded overflow-hidden">
      {entries.map(([key, pct]) => (
        <div
          key={key}
          style={{
            width: `${(pct * 100).toFixed(1)}%`,
            backgroundColor: CAUSE_COLORS[key] ?? CAUSE_COLORS.other,
          }}
          title={`${causeLabel(key)}: ${(pct * 100).toFixed(0)}%`}
        />
      ))}
    </div>
  );
}

// ─── Columns ─────────────────────────────────────────────────────────────────

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

const varianceColumns: Column<CountVariance>[] = [
  { key: 'ingredient_name', header: 'Ingredient', sortable: true },
  { key: 'category', header: 'Category', sortable: true },
  {
    key: 'theoretical_usage',
    header: 'Expected Qty',
    align: 'right',
    sortable: true,
    render: (r) => r.theoretical_usage.toFixed(2),
  },
  {
    key: 'actual_usage',
    header: 'Actual Qty',
    align: 'right',
    sortable: true,
    render: (r) => r.actual_usage.toFixed(2),
  },
  {
    key: 'variance_pct',
    header: 'Variance %',
    align: 'right',
    sortable: true,
    render: (r) => {
      const color = r.variance_pct > 10 ? 'text-red-600' : r.variance_pct > 5 ? 'text-amber-600' : 'text-slate-200';
      return <span className={color}>{r.variance_pct.toFixed(1)}%</span>;
    },
  },
  {
    key: 'variance_cents',
    header: 'Variance $',
    align: 'right',
    sortable: true,
    render: (r) => varianceDollars(r.variance_cents),
  },
  {
    key: 'severity',
    header: 'Severity',
    align: 'center',
    render: (r) => {
      const variant = r.severity === 'critical' ? 'critical' : r.severity === 'warning' ? 'warning' : 'info';
      return <StatusBadge variant={variant}>{r.severity}</StatusBadge>;
    },
  },
  {
    key: 'cause_probabilities',
    header: 'Likely Cause',
    render: (r) => <CauseProbabilityBar probs={r.cause_probabilities} />,
  },
];

// ─── Expiry columns ───────────────────────────────────────────────────────────

function expiryStatusVariant(status: ExpiryItem['status']): 'critical' | 'warning' | 'info' | 'success' {
  if (status === 'expired') return 'critical';
  if (status === 'expires_today') return 'warning';
  if (status === 'expiring_soon') return 'info';
  return 'success';
}

function expiryStatusLabel(status: ExpiryItem['status']): string {
  if (status === 'expired') return 'Expired';
  if (status === 'expires_today') return 'Today';
  if (status === 'expiring_soon') return 'Soon';
  return 'OK';
}

const expiryColumns: Column<ExpiryItem>[] = [
  {
    key: 'name',
    header: 'Name',
    sortable: true,
    render: (r) => {
      const urgent = r.status === 'expired' || r.status === 'expires_today';
      return <span className={urgent ? 'font-medium text-white' : ''}>{r.name}</span>;
    },
  },
  {
    key: 'category',
    header: 'Category',
    render: (r) => (
      <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-slate-700 text-slate-200">
        {r.category}
      </span>
    ),
  },
  {
    key: 'batch_number',
    header: 'Batch #',
    render: (r) => r.batch_number ? <span className="font-mono text-xs text-slate-300">{r.batch_number}</span> : <span className="text-slate-500">—</span>,
  },
  {
    key: 'expiry_date',
    header: 'Expiry Date',
    render: (r) => r.expiry_date ?? <span className="text-slate-500">—</span>,
  },
  {
    key: 'days_until_expiry',
    header: 'Days Left',
    align: 'right',
    sortable: true,
    render: (r) => {
      if (r.days_until_expiry < 0) {
        return <span className="font-bold text-red-500">{r.days_until_expiry}</span>;
      }
      if (r.days_until_expiry === 0) {
        return <span className="font-bold text-amber-400">0</span>;
      }
      return <span>{r.days_until_expiry}</span>;
    },
  },
  {
    key: 'status',
    header: 'Status',
    align: 'center',
    render: (r) => (
      <StatusBadge variant={expiryStatusVariant(r.status)}>
        {expiryStatusLabel(r.status)}
      </StatusBadge>
    ),
  },
  {
    key: 'vendor_name',
    header: 'Vendor',
    render: (r) => <span className="text-slate-300">{r.vendor_name}</span>,
  },
];

// ─── Tab types ────────────────────────────────────────────────────────────────

type Tab = 'usage' | 'par' | 'variances' | 'expiry';

const TABS: { id: Tab; label: string }[] = [
  { id: 'usage', label: 'Usage' },
  { id: 'par', label: 'PAR Status' },
  { id: 'variances', label: 'Variances' },
  { id: 'expiry', label: 'Expiry' },
];

// ─── Page ─────────────────────────────────────────────────────────────────────

export default function InventoryPage() {
  const locationId = useLocationStore((s) => s.selectedLocationId);
  const [activeTab, setActiveTab] = useState<Tab>('usage');

  const { data: usageData, isLoading: usageLoading, error: usageError, refetch: refetchUsage } = useUsage(locationId);
  const { data: parData, isLoading: parLoading, error: parError, refetch: refetchPar } = usePARStatus(locationId);
  const { data: breachData } = usePARBreaches(locationId);
  const {
    data: varianceData,
    isLoading: varianceLoading,
    error: varianceError,
    refetch: refetchVariances,
  } = useVariances(locationId);
  const { data: expiryData, isLoading: expiryLoading, error: expiryError, refetch: refetchExpiry } = useExpiry(locationId);

  if (!locationId) return <LoadingSpinner fullPage />;

  // Sort variances by variance $ descending as default display order
  const sortedVariances = varianceData?.variances
    ? [...varianceData.variances].sort((a, b) => b.variance_cents - a.variance_cents)
    : [];

  // Sort expiry items by days_until_expiry ascending (most urgent first)
  const sortedExpiryItems = expiryData?.items
    ? [...expiryData.items].sort((a, b) => a.days_until_expiry - b.days_until_expiry)
    : [];

  const expiredCount = expiryData?.expired_count ?? 0;
  const expiringTodayCount = expiryData?.expiring_today_count ?? 0;
  const expiringSoonCount = expiryData?.expiring_soon_count ?? 0;
  const okCount = sortedExpiryItems.filter((i) => i.status === 'ok').length;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold text-white">Inventory Intelligence</h1>
        <p className="text-sm text-slate-400 mt-1">Theoretical usage, PAR status, and variance analysis</p>
      </div>

      {/* Tab bar */}
      <div className="border-b border-white/10">
        <nav className="-mb-px flex gap-6" aria-label="Inventory tabs">
          {TABS.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={[
                'pb-3 text-sm font-medium border-b-2 transition-colors',
                activeTab === tab.id
                  ? 'border-orange-500 text-orange-600'
                  : 'border-transparent text-slate-400 hover:text-slate-200 hover:border-white/15',
              ].join(' ')}
            >
              {tab.label}
            </button>
          ))}
        </nav>
      </div>

      {/* Usage tab */}
      {activeTab === 'usage' && (
        <div>
          {usageError && (
            <ErrorBanner
              message={usageError instanceof Error ? usageError.message : 'Failed to load usage data'}
              retry={() => refetchUsage()}
            />
          )}
          <div className="mb-3">
            <h2 className="text-lg font-semibold text-white">Theoretical Usage</h2>
            <p className="text-xs text-slate-400 mt-0.5">Based on today's sales mix</p>
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
      )}

      {/* PAR Status tab */}
      {activeTab === 'par' && (
        <div>
          {parError && (
            <ErrorBanner
              message={parError instanceof Error ? parError.message : 'Failed to load PAR data'}
              retry={() => refetchPar()}
            />
          )}
          {breachData?.breaches && breachData.breaches.length > 0 && (
            <div className="bg-red-50 border border-red-200 rounded-lg p-3 mb-4 flex items-center justify-between">
              <span className="text-red-800 text-sm font-medium">
                {breachData.breaches.length} ingredient{breachData.breaches.length > 1 ? 's' : ''} below reorder point
              </span>
              <a href="/purchase-orders" className="text-red-600 text-sm underline">View Purchase Orders →</a>
            </div>
          )}
          <div className="mb-3">
            <h2 className="text-lg font-semibold text-white">PAR Status</h2>
            <p className="text-xs text-slate-400 mt-0.5">Current stock vs. target levels</p>
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
      )}

      {/* Variances tab */}
      {activeTab === 'variances' && (
        <div>
          {varianceError && (
            <ErrorBanner
              message={varianceError instanceof Error ? varianceError.message : 'Failed to load variance data'}
              retry={() => refetchVariances()}
            />
          )}
          <div className="mb-3">
            <h2 className="text-lg font-semibold text-white">Inventory Variances</h2>
            <p className="text-xs text-slate-400 mt-0.5">
              Theoretical vs. actual usage — sorted by highest dollar impact. Hover the cause bar for details.
            </p>
          </div>
          <DataTable
            columns={varianceColumns}
            data={sortedVariances}
            keyExtractor={(r) => r.variance_id}
            isLoading={varianceLoading}
            emptyTitle="No variances found"
            emptyDescription="All ingredient counts are within expected range."
          />
        </div>
      )}

      {/* Expiry tab */}
      {activeTab === 'expiry' && (
        <div>
          {expiryError && (
            <ErrorBanner
              message={expiryError instanceof Error ? expiryError.message : 'Failed to load expiry data'}
              retry={() => refetchExpiry()}
            />
          )}

          {/* Alert banner — shown only when there are expired or expiring-today items */}
          {(expiredCount > 0 || expiringTodayCount > 0) && (
            <div className="bg-red-950 border border-red-700 rounded-lg px-4 py-3 mb-4 flex items-center gap-2">
              <span className="text-red-300 text-sm font-medium">
                ⚠️ {expiredCount} item{expiredCount !== 1 ? 's' : ''} expired,{' '}
                {expiringTodayCount} expiring today — immediate action required
              </span>
            </div>
          )}

          {/* KPI cards */}
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 mb-5">
            <div className="bg-red-950/60 border border-red-800 rounded-lg p-3">
              <p className="text-xs text-red-400 font-medium uppercase tracking-wide mb-1">Expired</p>
              <p className="text-2xl font-bold text-red-400">{expiredCount}</p>
            </div>
            <div className="bg-amber-950/60 border border-amber-700 rounded-lg p-3">
              <p className="text-xs text-amber-400 font-medium uppercase tracking-wide mb-1">Expires Today</p>
              <p className="text-2xl font-bold text-amber-400">{expiringTodayCount}</p>
            </div>
            <div className="bg-yellow-950/60 border border-yellow-700 rounded-lg p-3">
              <p className="text-xs text-yellow-400 font-medium uppercase tracking-wide mb-1">Expiring Soon</p>
              <p className="text-2xl font-bold text-yellow-400">{expiringSoonCount}</p>
            </div>
            <div className="bg-green-950/60 border border-green-800 rounded-lg p-3">
              <p className="text-xs text-green-400 font-medium uppercase tracking-wide mb-1">OK</p>
              <p className="text-2xl font-bold text-green-400">{okCount}</p>
            </div>
          </div>

          <div className="mb-3">
            <h2 className="text-lg font-semibold text-white">Expiry Tracking</h2>
            <p className="text-xs text-slate-400 mt-0.5">Sorted by urgency — most critical items first.</p>
          </div>

          <DataTable
            columns={expiryColumns}
            data={sortedExpiryItems}
            keyExtractor={(r) => r.ingredient_id}
            isLoading={expiryLoading}
            emptyTitle="No expiry data"
            emptyDescription="No ingredient expiry dates have been recorded for this location."
            rowClassName={(r) => {
              if (r.status === 'expired') return 'border-l-2 border-red-600 bg-red-950/20';
              if (r.status === 'expires_today') return 'border-l-2 border-amber-500 bg-amber-950/20';
              return '';
            }}
          />
        </div>
      )}
    </div>
  );
}
