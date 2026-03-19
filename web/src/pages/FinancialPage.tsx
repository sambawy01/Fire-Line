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
import DataTable from '../components/ui/DataTable';
import type { Column } from '../components/ui/DataTable';
import StatusBadge from '../components/ui/StatusBadge';
import LoadingSpinner from '../components/ui/LoadingSpinner';
import ErrorBanner from '../components/ui/ErrorBanner';
import {
  DollarSign,
  TrendingDown,
  TrendingUp,
  Percent,
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
