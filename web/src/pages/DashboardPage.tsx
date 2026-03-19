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
