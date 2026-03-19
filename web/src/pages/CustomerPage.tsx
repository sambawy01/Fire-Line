import { useState } from 'react';
import { UserCheck, DollarSign, Star, AlertTriangle, Sparkles } from 'lucide-react';
import { useLocationStore } from '../stores/location';
import { useCustomers, useCustomerSummary, useAnalyzeCustomers } from '../hooks/useCustomers';
import KPICard from '../components/ui/KPICard';
import DataTable from '../components/ui/DataTable';
import type { Column } from '../components/ui/DataTable';
import StatusBadge from '../components/ui/StatusBadge';
import LoadingSpinner from '../components/ui/LoadingSpinner';
import ErrorBanner from '../components/ui/ErrorBanner';
import type { CustomerDetail } from '../lib/api';

function cents(v: number): string {
  return `$${(v / 100).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
}

const SEGMENT_LABELS: Record<string, string> = {
  new: 'New',
  regular: 'Regular',
  vip: 'VIP',
  at_risk: 'At Risk',
  lapsed: 'Lapsed',
};

type SegmentVariant = 'success' | 'info' | 'neutral' | 'warning' | 'critical';

function segmentVariant(segment: CustomerDetail['segment']): SegmentVariant {
  switch (segment) {
    case 'vip': return 'success';
    case 'regular': return 'info';
    case 'new': return 'neutral';
    case 'at_risk': return 'warning';
    case 'lapsed': return 'critical';
  }
}

const customerColumns: Column<CustomerDetail>[] = [
  {
    key: 'name',
    header: 'Name',
    sortable: true,
    render: (r) => <span className="font-semibold text-gray-800">{r.name}</span>,
  },
  {
    key: 'segment',
    header: 'Segment',
    sortable: true,
    render: (r) => (
      <StatusBadge variant={segmentVariant(r.segment)}>
        {SEGMENT_LABELS[r.segment] ?? r.segment}
      </StatusBadge>
    ),
  },
  {
    key: 'total_visits',
    header: 'Visits',
    align: 'right',
    sortable: true,
  },
  {
    key: 'total_spend',
    header: 'Total Spend',
    align: 'right',
    sortable: true,
    render: (r) => cents(r.total_spend),
  },
  {
    key: 'avg_check',
    header: 'Avg Check',
    align: 'right',
    sortable: true,
    render: (r) => cents(r.avg_check),
  },
  {
    key: 'last_visit',
    header: 'Last Visit',
    align: 'right',
    sortable: true,
    render: (r) =>
      r.last_visit
        ? new Date(r.last_visit).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
        : '—',
  },
  {
    key: 'ai_summary',
    header: 'AI Insight',
    render: (r) =>
      r.ai_summary ? (
        <span
          className="text-gray-700 text-sm"
          title={r.ai_summary}
        >
          {r.ai_summary.length > 60 ? `${r.ai_summary.slice(0, 60)}…` : r.ai_summary}
        </span>
      ) : (
        <span className="text-gray-400 text-sm italic">No AI summary</span>
      ),
  },
];

const SEGMENT_OPTIONS = [
  { value: '', label: 'All Segments' },
  { value: 'new', label: 'New' },
  { value: 'regular', label: 'Regular' },
  { value: 'vip', label: 'VIP' },
  { value: 'at_risk', label: 'At Risk' },
  { value: 'lapsed', label: 'Lapsed' },
];

export default function CustomerPage() {
  const locationId = useLocationStore((s) => s.selectedLocationId);
  const [segmentFilter, setSegmentFilter] = useState('');
  const [analyzeMessage, setAnalyzeMessage] = useState<string | null>(null);

  const {
    data: summary,
    isLoading: summaryLoading,
    error: summaryError,
    refetch: refetchSummary,
  } = useCustomerSummary(locationId);

  const {
    data: customersData,
    isLoading: customersLoading,
    error: customersError,
    refetch: refetchCustomers,
  } = useCustomers(locationId);

  const analyze = useAnalyzeCustomers();

  if (!locationId) return <LoadingSpinner fullPage />;

  const allCustomers = customersData?.customers ?? [];
  const customers = segmentFilter
    ? allCustomers.filter((c) => c.segment === segmentFilter)
    : allCustomers;

  const error = summaryError ?? customersError;
  const errorMessage = error instanceof Error ? error.message : 'Failed to load customer data';

  function handleAnalyze() {
    if (!locationId) return;
    setAnalyzeMessage(null);
    analyze.mutate(locationId, {
      onSuccess: (result) => {
        setAnalyzeMessage(result.message);
        setTimeout(() => setAnalyzeMessage(null), 5000);
      },
    });
  }

  return (
    <div className="space-y-8">
      {/* Page header */}
      <div>
        <h1 className="text-2xl font-bold text-gray-800">Customer Intelligence</h1>
        <p className="text-sm text-gray-500 mt-1">
          Guest segmentation, lifetime value, and AI-powered customer insights
        </p>
      </div>

      {error && (
        <ErrorBanner
          message={errorMessage}
          retry={() => {
            void refetchSummary();
            void refetchCustomers();
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
            label="Total Customers"
            value={summary ? String(summary.total_customers) : '—'}
            icon={UserCheck}
            iconColor="text-gray-600"
            bgTint="bg-gray-100"
          />
          <KPICard
            label="Avg Lifetime Value"
            value={summary ? cents(summary.avg_lifetime_value) : '$—'}
            icon={DollarSign}
            iconColor="text-blue-600"
            bgTint="bg-blue-50"
          />
          <KPICard
            label="VIP Customers"
            value={summary ? String(summary.vip_count) : '—'}
            icon={Star}
            iconColor="text-emerald-600"
            bgTint="bg-emerald-50"
          />
          <KPICard
            label="At Risk"
            value={summary ? String(summary.at_risk_count) : '—'}
            icon={AlertTriangle}
            iconColor="text-red-600"
            bgTint="bg-red-50"
          />
        </div>
      )}

      {/* Controls row */}
      <div className="flex flex-wrap items-center gap-3">
        <button
          onClick={handleAnalyze}
          disabled={analyze.isPending}
          className="flex items-center gap-2 px-4 py-2 rounded-md text-sm font-semibold text-white bg-[#F97316] hover:bg-orange-600 disabled:opacity-60 disabled:cursor-not-allowed transition-colors"
        >
          {analyze.isPending ? (
            <LoadingSpinner size="sm" />
          ) : (
            <Sparkles className="h-4 w-4" />
          )}
          Analyze with AI
        </button>

        <select
          value={segmentFilter}
          onChange={(e) => setSegmentFilter(e.target.value)}
          className="rounded-md border border-gray-300 bg-white text-sm px-3 py-2 text-gray-700 focus:outline-none focus:ring-2 focus:ring-[#F97316]"
        >
          {SEGMENT_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>
              {opt.label}
            </option>
          ))}
        </select>

        {analyzeMessage && (
          <span className="text-sm text-emerald-700 font-medium">{analyzeMessage}</span>
        )}
      </div>

      {/* Customer table */}
      <div>
        <h2 className="text-lg font-semibold text-gray-800 mb-3">Customer Detail</h2>
        <DataTable
          columns={customerColumns}
          data={customers}
          keyExtractor={(r) => r.customer_id}
          isLoading={customersLoading}
          emptyTitle="No customers found"
          emptyDescription="No customer data is available for this location."
        />
      </div>
    </div>
  );
}
