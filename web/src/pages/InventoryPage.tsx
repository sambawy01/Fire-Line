import { useState } from 'react';
import { useLocationStore } from '../stores/location';
import { useUsage, usePARStatus, useVariances } from '../hooks/useInventory';
import { usePARBreaches } from '../hooks/usePurchaseOrders';
import DataTable from '../components/ui/DataTable';
import type { Column } from '../components/ui/DataTable';
import StatusBadge from '../components/ui/StatusBadge';
import ErrorBanner from '../components/ui/ErrorBanner';
import LoadingSpinner from '../components/ui/LoadingSpinner';
import type { TheoreticalUsage, PARStatus, CountVariance } from '../lib/api';

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

// ─── Tab types ────────────────────────────────────────────────────────────────

type Tab = 'usage' | 'par' | 'variances';

const TABS: { id: Tab; label: string }[] = [
  { id: 'usage', label: 'Usage' },
  { id: 'par', label: 'PAR Status' },
  { id: 'variances', label: 'Variances' },
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

  if (!locationId) return <LoadingSpinner fullPage />;

  // Sort variances by variance $ descending as default display order
  const sortedVariances = varianceData?.variances
    ? [...varianceData.variances].sort((a, b) => b.variance_cents - a.variance_cents)
    : [];

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
    </div>
  );
}
