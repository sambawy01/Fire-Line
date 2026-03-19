import { useLocationStore } from '../stores/location';
import { useUsage, usePARStatus } from '../hooks/useInventory';
import DataTable from '../components/ui/DataTable';
import type { Column } from '../components/ui/DataTable';
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
