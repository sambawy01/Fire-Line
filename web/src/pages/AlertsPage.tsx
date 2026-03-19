import React, { useState, useMemo } from 'react';
import {
  AlertTriangle,
  AlertCircle,
  Info,
  CheckCircle,
  Shield,
  Clock,
  Filter,
} from 'lucide-react';

type Severity = 'critical' | 'warning' | 'info';
type FilterType = 'all' | Severity;

interface Alert {
  id: string;
  severity: Severity;
  title: string;
  description: string;
  module: string;
  timestamp: string;
  acknowledged: boolean;
  resolved: boolean;
}

const MOCK_ALERTS: Alert[] = [
  {
    id: 'alert-001',
    severity: 'critical',
    title: 'Chicken Breast stock critically low',
    description:
      'Current inventory is 4 lbs remaining — below the 10 lb reorder threshold. Projected to run out before tomorrow lunch service.',
    module: 'inventory',
    timestamp: '2026-03-19T14:32:00Z',
    acknowledged: false,
    resolved: false,
  },
  {
    id: 'alert-002',
    severity: 'critical',
    title: 'Daily revenue variance exceeds threshold',
    description:
      'Actual revenue is 22% below forecast for today. Check for missed transactions or POS sync issues.',
    module: 'financial',
    timestamp: '2026-03-19T13:15:00Z',
    acknowledged: false,
    resolved: false,
  },
  {
    id: 'alert-003',
    severity: 'warning',
    title: 'Toast POS sync delayed',
    description:
      'The Toast adapter has not synced in over 30 minutes. Orders placed after 1:45 PM may not be reflected in reports.',
    module: 'adapter',
    timestamp: '2026-03-19T14:18:00Z',
    acknowledged: false,
    resolved: false,
  },
  {
    id: 'alert-004',
    severity: 'warning',
    title: 'Fryer oil temperature fluctuation',
    description:
      'Fryer #2 reported temperature readings outside the acceptable range twice in the last hour. Schedule maintenance check.',
    module: 'inventory',
    timestamp: '2026-03-19T11:47:00Z',
    acknowledged: false,
    resolved: false,
  },
  {
    id: 'alert-005',
    severity: 'info',
    title: 'Weekly COGS report ready',
    description:
      'The automated cost-of-goods-sold report for the week ending March 15 has been generated and is ready for review.',
    module: 'financial',
    timestamp: '2026-03-19T08:00:00Z',
    acknowledged: false,
    resolved: false,
  },
];

const SEVERITY_CONFIG: Record<
  Severity,
  { label: string; bg: string; text: string; border: string; icon: React.ElementType }
> = {
  critical: {
    label: 'Critical',
    bg: 'bg-red-50',
    text: 'text-red-700',
    border: 'border-red-200',
    icon: AlertCircle,
  },
  warning: {
    label: 'Warning',
    bg: 'bg-amber-50',
    text: 'text-amber-700',
    border: 'border-amber-200',
    icon: AlertTriangle,
  },
  info: {
    label: 'Info',
    bg: 'bg-blue-50',
    text: 'text-blue-700',
    border: 'border-blue-200',
    icon: Info,
  },
};

const MODULE_LABELS: Record<string, string> = {
  inventory: 'Inventory',
  financial: 'Financial',
  adapter: 'Adapter',
};

function formatTimestamp(iso: string): string {
  const date = new Date(iso);
  return date.toLocaleString('en-US', {
    month: 'short',
    day: 'numeric',
    hour: 'numeric',
    minute: '2-digit',
    hour12: true,
  });
}

export default function AlertsPage() {
  const [alerts, setAlerts] = useState<Alert[]>(MOCK_ALERTS);
  const [activeFilter, setActiveFilter] = useState<FilterType>('all');

  const filteredAlerts = useMemo(
    () =>
      activeFilter === 'all'
        ? alerts
        : alerts.filter((a) => a.severity === activeFilter),
    [alerts, activeFilter],
  );

  const activeCount = alerts.filter((a) => !a.resolved).length;

  function handleAcknowledge(id: string) {
    console.log(`Acknowledged alert: ${id}`);
    setAlerts((prev) =>
      prev.map((a) => (a.id === id ? { ...a, acknowledged: true } : a)),
    );
  }

  function handleResolve(id: string) {
    console.log(`Resolved alert: ${id}`);
    setAlerts((prev) =>
      prev.map((a) => (a.id === id ? { ...a, resolved: true } : a)),
    );
  }

  const filters: { key: FilterType; label: string }[] = [
    { key: 'all', label: 'All' },
    { key: 'critical', label: 'Critical' },
    { key: 'warning', label: 'Warning' },
    { key: 'info', label: 'Info' },
  ];

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="mx-auto max-w-4xl px-4 py-8 sm:px-6 lg:px-8">
        {/* Header */}
        <div className="mb-6 flex items-center gap-3">
          <Shield className="h-7 w-7 text-gray-900" />
          <h1 className="text-2xl font-bold text-gray-900">
            Priority Action Queue
          </h1>
          <span className="inline-flex items-center rounded-full bg-[#F97316] px-3 py-0.5 text-sm font-semibold text-white">
            {activeCount} active
          </span>
        </div>

        {/* Filters */}
        <div className="mb-6 flex flex-wrap items-center gap-2">
          <Filter className="h-4 w-4 text-gray-500" />
          {filters.map((f) => (
            <button
              key={f.key}
              onClick={() => setActiveFilter(f.key)}
              className={`rounded-lg px-4 py-1.5 text-sm font-medium transition-colors ${
                activeFilter === f.key
                  ? 'bg-gray-900 text-white'
                  : 'bg-white text-gray-600 ring-1 ring-gray-200 hover:bg-gray-100'
              }`}
            >
              {f.label}
            </button>
          ))}
        </div>

        {/* Alert Cards */}
        {filteredAlerts.length === 0 ? (
          <div className="rounded-xl border border-dashed border-gray-300 bg-white py-16 text-center">
            <CheckCircle className="mx-auto mb-3 h-10 w-10 text-green-400" />
            <p className="text-lg font-medium text-gray-700">
              No alerts match this filter
            </p>
            <p className="mt-1 text-sm text-gray-500">
              All clear — nothing needs your attention right now.
            </p>
          </div>
        ) : (
          <ul className="space-y-4">
            {filteredAlerts.map((alert) => {
              const config = SEVERITY_CONFIG[alert.severity];
              const SeverityIcon = config.icon;

              return (
                <li
                  key={alert.id}
                  className={`rounded-xl border bg-white shadow-sm transition-opacity ${
                    alert.resolved ? 'opacity-50' : ''
                  }`}
                >
                  <div className="p-5">
                    {/* Top row: severity badge + module + timestamp */}
                    <div className="mb-3 flex flex-wrap items-center gap-2">
                      <span
                        className={`inline-flex items-center gap-1 rounded-full px-2.5 py-0.5 text-xs font-semibold ${config.bg} ${config.text} ${config.border} border`}
                      >
                        <SeverityIcon className="h-3.5 w-3.5" />
                        {config.label}
                      </span>
                      <span className="rounded-full bg-gray-100 px-2.5 py-0.5 text-xs font-medium text-gray-600">
                        {MODULE_LABELS[alert.module] ?? alert.module}
                      </span>
                      <span className="ml-auto flex items-center gap-1 text-xs text-gray-400">
                        <Clock className="h-3.5 w-3.5" />
                        {formatTimestamp(alert.timestamp)}
                      </span>
                    </div>

                    {/* Title + description */}
                    <h3 className="text-base font-semibold text-gray-900">
                      {alert.title}
                    </h3>
                    <p className="mt-1 text-sm leading-relaxed text-gray-600">
                      {alert.description}
                    </p>

                    {/* Actions */}
                    <div className="mt-4 flex items-center gap-3">
                      <button
                        disabled={alert.acknowledged}
                        onClick={() => handleAcknowledge(alert.id)}
                        className={`rounded-lg px-4 py-1.5 text-sm font-medium transition-colors ${
                          alert.acknowledged
                            ? 'cursor-default bg-gray-100 text-gray-400'
                            : 'bg-[#F97316] text-white hover:bg-[#EA580C]'
                        }`}
                      >
                        {alert.acknowledged ? 'Acknowledged' : 'Acknowledge'}
                      </button>
                      <button
                        disabled={alert.resolved}
                        onClick={() => handleResolve(alert.id)}
                        className={`rounded-lg px-4 py-1.5 text-sm font-medium transition-colors ${
                          alert.resolved
                            ? 'cursor-default bg-gray-100 text-gray-400'
                            : 'bg-white text-gray-700 ring-1 ring-gray-200 hover:bg-gray-50'
                        }`}
                      >
                        {alert.resolved ? 'Resolved' : 'Resolve'}
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
