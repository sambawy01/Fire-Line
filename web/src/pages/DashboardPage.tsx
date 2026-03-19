import {
  DollarSign,
  TrendingDown,
  Percent,
  AlertTriangle,
  AlertCircle,
  Info,
} from 'lucide-react';

const kpiCards = [
  {
    label: 'Revenue (Today)',
    value: '$4,250',
    icon: DollarSign,
    iconColor: 'text-emerald-600',
    bgTint: 'bg-emerald-50',
  },
  {
    label: 'COGS',
    value: '$1,275',
    icon: TrendingDown,
    iconColor: 'text-red-600',
    bgTint: 'bg-red-50',
  },
  {
    label: 'Gross Margin %',
    value: '70%',
    icon: Percent,
    iconColor: 'text-blue-600',
    bgTint: 'bg-blue-50',
  },
  {
    label: 'Active Alerts',
    value: '3',
    icon: AlertTriangle,
    iconColor: 'text-orange-600',
    bgTint: 'bg-orange-50',
  },
];

type Severity = 'critical' | 'warning' | 'info';

const priorityAlerts: { title: string; message: string; severity: Severity }[] = [
  {
    title: 'Ground Beef Below PAR',
    message: 'Current stock is 12 lbs — PAR level is 30 lbs. Reorder immediately.',
    severity: 'critical',
  },
  {
    title: 'Delivery COGS Spike',
    message: 'Delivery channel COGS rose 8% over the last 7 days. Review vendor pricing.',
    severity: 'warning',
  },
  {
    title: 'New POS Sync Available',
    message: 'Toast adapter v2.4 is available with improved ticket parsing.',
    severity: 'info',
  },
];

const severityConfig: Record<Severity, { badge: string; border: string; icon: typeof AlertCircle }> = {
  critical: { badge: 'bg-red-100 text-red-700', border: 'border-l-red-500', icon: AlertCircle },
  warning: { badge: 'bg-yellow-100 text-yellow-700', border: 'border-l-yellow-500', icon: AlertTriangle },
  info: { badge: 'bg-blue-100 text-blue-700', border: 'border-l-blue-500', icon: Info },
};

export default function DashboardPage() {
  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold text-gray-800">Dashboard</h1>
        <p className="text-sm text-gray-500 mt-1">
          Today's operational snapshot
        </p>
      </div>

      {/* KPI Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-5">
        {kpiCards.map(({ label, value, icon: Icon, iconColor, bgTint }) => (
          <div
            key={label}
            className="bg-white rounded-xl border border-gray-200 p-5 flex items-start gap-4 shadow-sm"
          >
            <div className={`${bgTint} p-3 rounded-lg`}>
              <Icon className={`h-6 w-6 ${iconColor}`} />
            </div>
            <div>
              <p className="text-sm text-gray-500">{label}</p>
              <p className="text-2xl font-bold text-gray-800 mt-0.5">{value}</p>
            </div>
          </div>
        ))}
      </div>

      {/* Priority Action Queue */}
      <div>
        <h2 className="text-lg font-semibold text-gray-800 mb-4">
          Priority Action Queue
        </h2>
        <div className="space-y-3">
          {priorityAlerts.map(({ title, message, severity }) => {
            const config = severityConfig[severity];
            const SevIcon = config.icon;
            return (
              <div
                key={title}
                className={`bg-white rounded-lg border border-gray-200 border-l-4 ${config.border} p-4 flex items-start gap-3 shadow-sm`}
              >
                <SevIcon className={`h-5 w-5 mt-0.5 shrink-0 ${config.badge.split(' ')[1]}`} />
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 mb-1">
                    <p className="font-medium text-gray-800">{title}</p>
                    <span
                      className={`text-xs font-medium px-2 py-0.5 rounded-full ${config.badge}`}
                    >
                      {severity}
                    </span>
                  </div>
                  <p className="text-sm text-gray-500">{message}</p>
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}
