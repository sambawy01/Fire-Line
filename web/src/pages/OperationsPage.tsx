import {
  ComposedChart,
  Bar,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from 'recharts';
import { ShoppingBag, Clock, TrendingUp, AlertCircle, DollarSign, XCircle } from 'lucide-react';
import { useLocationStore } from '../stores/location';
import { useOperationsSummary, useOperationsHourly } from '../hooks/useOperations';
import KPICard from '../components/ui/KPICard';
import DataTable from '../components/ui/DataTable';
import type { Column } from '../components/ui/DataTable';
import LoadingSpinner from '../components/ui/LoadingSpinner';
import ErrorBanner from '../components/ui/ErrorBanner';
import type { ChannelPerf } from '../lib/api';

function cents(v: number): string {
  return `$${(v / 100).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
}

function formatHour(h: number): string {
  if (h === 0) return '12AM';
  if (h === 12) return '12PM';
  return h < 12 ? `${h}AM` : `${h - 12}PM`;
}

const CHANNEL_LABELS: Record<string, string> = {
  dine_in: 'Dine-in',
  takeout: 'Takeout',
  delivery: 'Delivery',
  drive_thru: 'Drive-Thru',
  catering: 'Catering',
  online: 'Online',
};

const channelColumns: Column<ChannelPerf>[] = [
  {
    key: 'channel',
    header: 'Channel',
    sortable: true,
    render: (r) => CHANNEL_LABELS[r.channel] ?? r.channel,
  },
  {
    key: 'orders',
    header: 'Orders',
    align: 'right',
    sortable: true,
  },
  {
    key: 'pct_of_total',
    header: '% of Total',
    align: 'right',
    sortable: true,
    render: (r) => `${r.pct_of_total.toFixed(1)}%`,
  },
  {
    key: 'avg_ticket_time',
    header: 'Avg Ticket',
    align: 'right',
    sortable: true,
    render: (r) => `${r.avg_ticket_time.toFixed(1)} min`,
  },
  {
    key: 'revenue',
    header: 'Revenue',
    align: 'right',
    sortable: true,
    render: (r) => cents(r.revenue),
  },
];

export default function OperationsPage() {
  const locationId = useLocationStore((s) => s.selectedLocationId);

  const {
    data: summary,
    isLoading: summaryLoading,
    error: summaryError,
    refetch: refetchSummary,
  } = useOperationsSummary(locationId);

  const {
    data: hourlyData,
    isLoading: hourlyLoading,
    error: hourlyError,
    refetch: refetchHourly,
  } = useOperationsHourly(locationId);

  if (!locationId) return <LoadingSpinner fullPage />;

  const error = summaryError ?? hourlyError;
  const errorMessage = error instanceof Error ? error.message : 'Failed to load operations data';

  const hourly = (hourlyData?.hourly ?? []).map((h) => ({
    ...h,
    revenueDollars: h.revenue / 100,
  }));

  const channels = summary?.channel_performance ?? [];

  return (
    <div className="space-y-8">
      {/* Page header */}
      <div>
        <h1 className="text-2xl font-bold text-gray-800">Operations Intelligence</h1>
        <p className="text-sm text-gray-500 mt-1">
          Real-time ticket flow, hourly throughput, and channel performance
        </p>
      </div>

      {error && (
        <ErrorBanner
          message={errorMessage}
          retry={() => {
            void refetchSummary();
            void refetchHourly();
          }}
        />
      )}

      {/* KPI Cards */}
      {summaryLoading ? (
        <div className="flex justify-center py-8">
          <LoadingSpinner />
        </div>
      ) : (
        <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-4">
          <KPICard
            label="Orders Today"
            value={summary ? String(summary.orders_today) : '—'}
            icon={ShoppingBag}
            iconColor="text-blue-600"
            bgTint="bg-blue-50"
          />
          <KPICard
            label="Avg Ticket Time"
            value={summary ? `${summary.avg_ticket_time.toFixed(1)} min` : '—'}
            icon={Clock}
            iconColor="text-purple-600"
            bgTint="bg-purple-50"
          />
          <KPICard
            label="Orders / Hour"
            value={summary ? String(summary.orders_per_hour) : '—'}
            icon={TrendingUp}
            iconColor="text-emerald-600"
            bgTint="bg-emerald-50"
          />
          <KPICard
            label="Active Tickets"
            value={summary ? String(summary.active_tickets) : '—'}
            icon={AlertCircle}
            iconColor="text-orange-600"
            bgTint="bg-orange-50"
          />
          <KPICard
            label="Revenue / Hour"
            value={summary ? cents(summary.revenue_per_hour) : '$—'}
            icon={DollarSign}
            iconColor="text-green-600"
            bgTint="bg-green-50"
          />
          <KPICard
            label="Void Rate"
            value={summary ? `${summary.void_rate.toFixed(1)}%` : '—'}
            icon={XCircle}
            iconColor="text-red-600"
            bgTint="bg-red-50"
          />
        </div>
      )}

      {/* Hourly Chart */}
      <div className="bg-white rounded-xl border border-gray-200 p-6">
        <h2 className="text-lg font-semibold text-gray-800 mb-4">Hourly Throughput</h2>
        {hourlyLoading ? (
          <div className="flex justify-center py-12">
            <LoadingSpinner />
          </div>
        ) : hourly.length === 0 ? (
          <p className="text-center text-gray-400 py-12 text-sm">No hourly data available</p>
        ) : (
          <ResponsiveContainer width="100%" height={280}>
            <ComposedChart data={hourly} margin={{ top: 4, right: 24, left: 0, bottom: 0 }}>
              <CartesianGrid strokeDasharray="3 3" stroke="#F3F4F6" />
              <XAxis
                dataKey="hour"
                tickFormatter={formatHour}
                tick={{ fontSize: 12, fill: '#6B7280' }}
                axisLine={false}
                tickLine={false}
              />
              <YAxis
                yAxisId="left"
                tick={{ fontSize: 12, fill: '#6B7280' }}
                axisLine={false}
                tickLine={false}
                allowDecimals={false}
              />
              <YAxis
                yAxisId="right"
                orientation="right"
                tick={{ fontSize: 12, fill: '#6B7280' }}
                axisLine={false}
                tickLine={false}
                tickFormatter={(v: number) => `$${v}`}
              />
              <Tooltip
                formatter={(value: number, name: string) => {
                  if (name === 'Revenue') return [`$${value.toFixed(2)}`, name];
                  return [value, name];
                }}
                labelFormatter={(label: number) => formatHour(label)}
                contentStyle={{ borderRadius: '8px', border: '1px solid #E5E7EB', fontSize: 13 }}
              />
              <Bar yAxisId="left" dataKey="orders" name="Orders" fill="#F97316" radius={[3, 3, 0, 0]} />
              <Line
                yAxisId="right"
                type="monotone"
                dataKey="revenueDollars"
                name="Revenue"
                stroke="#3B82F6"
                strokeWidth={2}
                dot={false}
              />
            </ComposedChart>
          </ResponsiveContainer>
        )}
      </div>

      {/* Channel Performance */}
      <div>
        <h2 className="text-lg font-semibold text-gray-800 mb-3">Channel Performance</h2>
        <DataTable
          columns={channelColumns}
          data={channels}
          keyExtractor={(r) => r.channel}
          isLoading={summaryLoading}
          emptyTitle="No channel data"
          emptyDescription="No channel performance data is available for this location."
        />
      </div>
    </div>
  );
}
