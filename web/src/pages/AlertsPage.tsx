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
    <div className="min-h-screen">
      <div className="mx-auto max-w-4xl px-4 py-8 sm:px-6 lg:px-8">
        {/* Header */}
        <div className="mb-6 flex items-center gap-3">
          <Shield className="h-7 w-7 text-white" />
          <h1 className="text-2xl font-bold text-white">Priority Action Queue</h1>
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
          <Filter className="h-4 w-4 text-slate-400" />
          {filters.map((f) => (
            <button
              key={f.key}
              onClick={() => setActiveFilter(f.key)}
              className={`rounded-lg px-4 py-1.5 text-sm font-medium transition-colors ${
                activeFilter === f.key
                  ? 'bg-[#F97316] text-white'
                  : 'bg-white/5 text-slate-300 ring-1 ring-white/10 hover:bg-white/10'
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
          <div className="rounded-xl border border-dashed border-white/15 bg-white/5 py-16 text-center">
            <CheckCircle className="mx-auto mb-3 h-10 w-10 text-green-400" />
            <p className="text-lg font-medium text-slate-200">No alerts match this filter</p>
            <p className="mt-1 text-sm text-slate-400">All clear — nothing needs your attention right now.</p>
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
                  className={`rounded-xl border border-white/10 bg-white/5 transition-opacity ${isResolved ? 'opacity-50' : ''}`}
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
                      <span className="ml-auto flex items-center gap-1 text-xs text-slate-300">
                        <Clock className="h-3.5 w-3.5" />
                        {formatTimestamp(alert.created_at)}
                      </span>
                    </div>

                    <h3 className="text-base font-semibold text-white">{alert.title}</h3>
                    <p className="mt-1 text-sm leading-relaxed text-slate-300">{alert.description}</p>

                    <div className="mt-4 flex items-center gap-3">
                      <button
                        disabled={isAcked || ackMutation.isPending}
                        onClick={() => ackMutation.mutate(alert.alert_id)}
                        className={`rounded-lg px-4 py-1.5 text-sm font-medium transition-colors ${
                          isAcked
                            ? 'cursor-default bg-white/5 text-slate-300'
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
                            ? 'cursor-default bg-white/5 text-slate-300'
                            : 'bg-white/10 text-slate-200 ring-1 ring-white/10 hover:bg-white/15'
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
