import { useLocationStore } from '../stores/location';
import { useLaborSummary, useLaborEmployees } from '../hooks/useLabor';
import KPICard from '../components/ui/KPICard';
import DataTable from '../components/ui/DataTable';
import type { Column } from '../components/ui/DataTable';
import StatusBadge from '../components/ui/StatusBadge';
import LoadingSpinner from '../components/ui/LoadingSpinner';
import ErrorBanner from '../components/ui/ErrorBanner';
import type { EmployeeDetail } from '../lib/api';
import { DollarSign, Percent, Users, Clock } from 'lucide-react';

function cents(v: number): string {
  return `$${(v / 100).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
}

function capitalize(s: string): string {
  return s.charAt(0).toUpperCase() + s.slice(1).toLowerCase();
}

function statusVariant(status: string): 'success' | 'neutral' | 'critical' {
  if (status === 'active') return 'success';
  if (status === 'terminated') return 'critical';
  return 'neutral';
}

const employeeColumns: Column<EmployeeDetail>[] = [
  {
    key: 'display_name',
    header: 'Employee',
    sortable: true,
    render: (r) => <span className="font-semibold text-gray-800">{r.display_name}</span>,
  },
  {
    key: 'role',
    header: 'Role',
    sortable: true,
    render: (r) => capitalize(r.role),
  },
  {
    key: 'status',
    header: 'Status',
    sortable: true,
    render: (r) => (
      <StatusBadge variant={statusVariant(r.status)}>
        {capitalize(r.status)}
      </StatusBadge>
    ),
  },
  {
    key: 'shift_count',
    header: 'Shifts',
    align: 'right',
    sortable: true,
  },
  {
    key: 'hours_worked',
    header: 'Hours',
    align: 'right',
    sortable: true,
    render: (r) => r.hours_worked.toFixed(1),
  },
  {
    key: 'labor_cost',
    header: 'Cost ($)',
    align: 'right',
    sortable: true,
    render: (r) => cents(r.labor_cost),
  },
  {
    key: 'avg_hours_per_shift',
    header: 'Avg Hrs/Shift',
    align: 'right',
    sortable: true,
    render: (r) => r.avg_hours_per_shift.toFixed(1),
  },
  {
    key: 'hourly_rate',
    header: 'Rate ($/hr)',
    align: 'right',
    sortable: true,
    render: (r) => cents(r.hourly_rate),
  },
];

export default function LaborPage() {
  const locationId = useLocationStore((s) => s.selectedLocationId);

  const {
    data: summary,
    isLoading: summaryLoading,
    error: summaryError,
    refetch: refetchSummary,
  } = useLaborSummary(locationId);

  const {
    data: employeesData,
    isLoading: employeesLoading,
    error: employeesError,
    refetch: refetchEmployees,
  } = useLaborEmployees(locationId);

  if (!locationId) return <LoadingSpinner fullPage />;

  const employees = employeesData?.employees ?? [];

  const error = summaryError ?? employeesError;
  const errorMessage = error instanceof Error ? error.message : 'Failed to load labor data';

  return (
    <div className="space-y-8">
      {/* Page header */}
      <div>
        <h1 className="text-2xl font-bold text-gray-800">Labor Intelligence</h1>
        <p className="text-sm text-gray-500 mt-1">
          Workforce costs, hours, and employee performance breakdown
        </p>
      </div>

      {error && (
        <ErrorBanner
          message={errorMessage}
          retry={() => {
            void refetchSummary();
            void refetchEmployees();
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
            label="Labor Cost"
            value={summary ? cents(summary.total_labor_cost) : '$—'}
            icon={DollarSign}
            iconColor="text-red-600"
            bgTint="bg-red-50"
          />
          <KPICard
            label="Labor Cost %"
            value={summary ? `${summary.labor_cost_pct.toFixed(1)}%` : '—'}
            icon={Percent}
            iconColor="text-blue-600"
            bgTint="bg-blue-50"
          />
          <KPICard
            label="Active Employees"
            value={summary ? String(summary.employee_count) : '—'}
            icon={Users}
            iconColor="text-gray-600"
            bgTint="bg-gray-100"
          />
          <KPICard
            label="Total Hours"
            value={summary ? summary.total_hours.toFixed(1) : '—'}
            icon={Clock}
            iconColor="text-purple-600"
            bgTint="bg-purple-50"
          />
        </div>
      )}

      {/* Employee table */}
      <div>
        <h2 className="text-lg font-semibold text-gray-800 mb-3">Employee Detail</h2>
        <DataTable
          columns={employeeColumns}
          data={employees}
          keyExtractor={(r) => r.employee_id}
          isLoading={employeesLoading}
          emptyTitle="No employees found"
          emptyDescription="No employee data is available for this location and period."
        />
      </div>
    </div>
  );
}
