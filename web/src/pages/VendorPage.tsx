import { useLocationStore } from '../stores/location';
import { useVendors, useVendorSummary } from '../hooks/useVendor';
import KPICard from '../components/ui/KPICard';
import DataTable from '../components/ui/DataTable';
import type { Column } from '../components/ui/DataTable';
import StatusBadge from '../components/ui/StatusBadge';
import LoadingSpinner from '../components/ui/LoadingSpinner';
import ErrorBanner from '../components/ui/ErrorBanner';
import type { VendorAnalysis } from '../lib/api';
import { Truck, DollarSign, Star, Package } from 'lucide-react';

function cents(v: number): string {
  return `$${(v / 100).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
}

function scoreVariant(score: number): 'success' | 'warning' | 'critical' {
  if (score >= 70) return 'success';
  if (score >= 40) return 'warning';
  return 'critical';
}

const vendorColumns: Column<VendorAnalysis>[] = [
  {
    key: 'vendor_name',
    header: 'Vendor',
    sortable: true,
    render: (r) => <span className="font-semibold text-gray-800">{r.vendor_name}</span>,
  },
  {
    key: 'items_supplied',
    header: 'Items',
    align: 'right',
    sortable: true,
  },
  {
    key: 'total_spend',
    header: 'Spend ($)',
    align: 'right',
    sortable: true,
    render: (r) => cents(r.total_spend),
  },
  {
    key: 'spend_pct',
    header: '% of Spend',
    align: 'right',
    sortable: true,
    render: (r) => `${r.spend_pct.toFixed(1)}%`,
  },
  {
    key: 'avg_cost_per_item',
    header: 'Avg Cost/Item',
    align: 'right',
    sortable: true,
    render: (r) => cents(r.avg_cost_per_item),
  },
  {
    key: 'score',
    header: 'Score',
    align: 'center',
    sortable: true,
    render: (r) => (
      <StatusBadge variant={scoreVariant(r.score)}>
        {r.score}
      </StatusBadge>
    ),
  },
];

export default function VendorPage() {
  const locationId = useLocationStore((s) => s.selectedLocationId);

  const {
    data: summary,
    isLoading: summaryLoading,
    error: summaryError,
    refetch: refetchSummary,
  } = useVendorSummary(locationId);

  const {
    data: vendorsData,
    isLoading: vendorsLoading,
    error: vendorsError,
    refetch: refetchVendors,
  } = useVendors(locationId);

  if (!locationId) return <LoadingSpinner fullPage />;

  const vendors = vendorsData?.vendors ?? [];

  const error = summaryError ?? vendorsError;
  const errorMessage = error instanceof Error ? error.message : 'Failed to load vendor data';

  return (
    <div className="space-y-8">
      {/* Page header */}
      <div>
        <h1 className="text-2xl font-bold text-gray-800">Vendor Intelligence</h1>
        <p className="text-sm text-gray-500 mt-1">
          Supplier spend analysis, concentration, and performance scoring
        </p>
      </div>

      {error && (
        <ErrorBanner
          message={errorMessage}
          retry={() => {
            void refetchSummary();
            void refetchVendors();
          }}
        />
      )}

      {/* KPI Cards */}
      {summaryLoading ? (
        <div className="flex justify-center py-8">
          <LoadingSpinner />
        </div>
      ) : (
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
          <KPICard
            label="Total Vendors"
            value={summary ? String(summary.total_vendors) : '—'}
            icon={Truck}
            iconColor="text-gray-600"
            bgTint="bg-gray-100"
          />
          <KPICard
            label="Total Spend"
            value={summary ? cents(summary.total_spend) : '$—'}
            icon={DollarSign}
            iconColor="text-red-600"
            bgTint="bg-red-50"
          />
          <KPICard
            label="Top Vendor"
            value={summary ? `${summary.top_vendor_name} (${summary.top_vendor_pct.toFixed(1)}%)` : '—'}
            icon={Star}
            iconColor="text-blue-600"
            bgTint="bg-blue-50"
          />
          <KPICard
            label="Avg Items/Vendor"
            value={summary ? summary.avg_items_per_vendor.toFixed(1) : '—'}
            icon={Package}
            iconColor="text-purple-600"
            bgTint="bg-purple-50"
          />
        </div>
      )}

      {/* Vendor table */}
      <div>
        <h2 className="text-lg font-semibold text-gray-800 mb-3">Vendor Detail</h2>
        <DataTable
          columns={vendorColumns}
          data={vendors}
          keyExtractor={(r) => r.vendor_name}
          isLoading={vendorsLoading}
          emptyTitle="No vendors found"
          emptyDescription="No vendor data is available for this location and period."
        />
      </div>
    </div>
  );
}
