import { useState } from 'react';
import {
  Activity,
  AlertTriangle,
  CheckCircle,
  Clock,
  DollarSign,
  Package,
  Truck,
  Users,
  Zap,
  ChevronDown,
  ChevronUp,
  X,
} from 'lucide-react';
import { useLocationStore } from '../stores/location';
import {
  useHealth,
  useOverload,
  usePriorities,
  useRealtimeHorizon,
  useShiftHorizon,
  useDailyHorizon,
  useWeeklyHorizon,
  useStrategicHorizon,
} from '../hooks/useOperations';
import KPICard from '../components/ui/KPICard';
import StatusBadge from '../components/ui/StatusBadge';
import DataTable from '../components/ui/DataTable';
import type { Column } from '../components/ui/DataTable';
import LoadingSpinner from '../components/ui/LoadingSpinner';
import ErrorBanner from '../components/ui/ErrorBanner';
import type { TicketPriority } from '../lib/api';

// ─── helpers ────────────────────────────────────────────────────────────────

function dollars(cents: number): string {
  return `EGP ${(cents / 100).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
}

function healthColor(score: number): string {
  if (score >= 80) return '#10B981'; // emerald
  if (score >= 60) return '#F59E0B'; // amber
  return '#EF4444'; // red
}

function healthLabel(status: string): string {
  const map: Record<string, string> = {
    healthy: 'Healthy',
    degraded: 'Degraded',
    critical: 'Critical',
    good: 'Good',
    warning: 'Warning',
  };
  return map[status] ?? status;
}

function severityVariant(severity: string): 'success' | 'warning' | 'critical' | 'neutral' {
  if (severity === 'critical') return 'critical';
  if (severity === 'elevated' || severity === 'warning') return 'warning';
  if (severity === 'normal') return 'success';
  return 'neutral';
}

function urgencyVariant(urgency: string): 'neutral' | 'warning' | 'critical' {
  if (urgency === 'critical') return 'critical';
  if (urgency === 'urgent') return 'warning';
  return 'neutral';
}

const CHANNEL_LABELS: Record<string, string> = {
  dine_in: 'Dine-in',
  takeout: 'Takeout',
  delivery: 'Delivery',
  drive_thru: 'Drive-Thru',
  catering: 'Catering',
  online: 'Online',
};

// ─── Circular Health Gauge ───────────────────────────────────────────────────

function HealthGauge({ score, status }: { score: number; status: string }) {
  const r = 44;
  const circ = 2 * Math.PI * r;
  const fill = circ * (score / 100);
  const color = healthColor(score);

  return (
    <div className="flex flex-col items-center gap-1">
      <div className="relative w-28 h-28">
        <svg viewBox="0 0 100 100" className="w-full h-full -rotate-90">
          <circle cx="50" cy="50" r={r} fill="none" stroke="#E5E7EB" strokeWidth="8" />
          <circle
            cx="50"
            cy="50"
            r={r}
            fill="none"
            stroke={color}
            strokeWidth="8"
            strokeDasharray={`${fill} ${circ}`}
            strokeLinecap="round"
          />
        </svg>
        <div className="absolute inset-0 flex flex-col items-center justify-center">
          <span className="text-2xl font-bold text-white">{Math.round(score)}</span>
          <span className="text-xs text-slate-300">/ 100</span>
        </div>
      </div>
      <span className="text-sm font-semibold" style={{ color }}>{healthLabel(status)}</span>
      <span className="text-xs text-slate-300 uppercase tracking-wide">Ops Health</span>
    </div>
  );
}

// ─── Sub-score pills ─────────────────────────────────────────────────────────

function SubScorePill({ label, score }: { label: string; score: number }) {
  const color = healthColor(score);
  return (
    <div className="flex flex-col items-center gap-0.5">
      <span className="text-xs text-slate-400">{label}</span>
      <span className="text-sm font-bold" style={{ color }}>{Math.round(score)}</span>
    </div>
  );
}

// ─── Overload indicator dot ───────────────────────────────────────────────────

function OverloadDot({ severity }: { severity: string }) {
  if (severity === 'normal') {
    return <span className="inline-block w-3 h-3 rounded-full bg-emerald-500" />;
  }
  const color = severity === 'critical' ? 'bg-red-500' : 'bg-amber-400';
  return (
    <span className="relative inline-flex">
      <span className={`animate-ping absolute inline-flex h-3 w-3 rounded-full ${color} opacity-75`} />
      <span className={`relative inline-flex rounded-full h-3 w-3 ${color}`} />
    </span>
  );
}

// ─── Priority table columns ───────────────────────────────────────────────────

const priorityColumns: Column<TicketPriority>[] = [
  {
    key: 'order_number',
    header: 'Order',
    render: (r) => <span className="font-mono font-semibold text-slate-200">{r.order_number}</span>,
  },
  {
    key: 'channel',
    header: 'Channel',
    render: (r) => (
      <StatusBadge variant="info">{CHANNEL_LABELS[r.channel] ?? r.channel}</StatusBadge>
    ),
  },
  {
    key: 'priority_score',
    header: 'Priority',
    align: 'right',
    render: (r) => (
      <div className="flex items-center gap-2 justify-end">
        <div className="w-20 h-2 bg-white/10 rounded-full overflow-hidden">
          <div
            className="h-full rounded-full"
            style={{
              width: `${Math.min(r.priority_score, 100)}%`,
              background: r.priority_score >= 80 ? '#EF4444' : r.priority_score >= 50 ? '#F59E0B' : '#10B981',
            }}
          />
        </div>
        <span className="text-xs text-slate-300 w-6 text-right">{r.priority_score}</span>
      </div>
    ),
  },
  {
    key: 'elapsed_minutes',
    header: 'Elapsed',
    align: 'right',
    render: (r) => {
      const remaining = r.sla_minutes - r.elapsed_minutes;
      const pct = Math.min((r.elapsed_minutes / r.sla_minutes) * 100, 100);
      const timeColor = pct >= 90 ? 'text-red-600' : pct >= 70 ? 'text-amber-600' : 'text-slate-300';
      return (
        <span className={`text-sm font-mono ${timeColor}`}>
          {r.elapsed_minutes.toFixed(0)}m / {r.sla_minutes}m
          {remaining < 0 ? ` (+${Math.abs(remaining).toFixed(0)}m)` : ''}
        </span>
      );
    },
  },
  {
    key: 'urgency',
    header: 'Urgency',
    render: (r) => (
      <StatusBadge variant={urgencyVariant(r.urgency)}>
        {r.urgency.charAt(0).toUpperCase() + r.urgency.slice(1)}
      </StatusBadge>
    ),
  },
];

// ─── Planning Horizon Tabs ────────────────────────────────────────────────────

type HorizonTab = 'shift' | 'daily' | 'weekly' | 'strategic';

const HORIZON_TABS: { id: HorizonTab; label: string }[] = [
  { id: 'shift', label: 'Shift (4hr)' },
  { id: 'daily', label: 'Daily' },
  { id: 'weekly', label: 'Weekly' },
  { id: 'strategic', label: 'Strategic (30d)' },
];

// ─── Main Page ────────────────────────────────────────────────────────────────

export default function OperationsPage() {
  const locationId = useLocationStore((s) => s.selectedLocationId);
  const [dismissedActions, setDismissedActions] = useState<Set<string>>(new Set());
  const [horizonsOpen, setHorizonsOpen] = useState(true);
  const [activeTab, setActiveTab] = useState<HorizonTab>('shift');

  const { data: health, isLoading: healthLoading, error: healthError } = useHealth(locationId);
  const { data: overload, isLoading: overloadLoading } = useOverload(locationId);
  const { data: priorityData, isLoading: prioritiesLoading } = usePriorities(locationId);
  const { data: realtime, isLoading: realtimeLoading } = useRealtimeHorizon(locationId);
  const { data: shift, isLoading: shiftLoading } = useShiftHorizon(locationId);
  const { data: daily, isLoading: dailyLoading } = useDailyHorizon(locationId);
  const { data: weekly, isLoading: weeklyLoading } = useWeeklyHorizon(locationId);
  const { data: strategic, isLoading: strategicLoading } = useStrategicHorizon(locationId);

  if (!locationId) return <LoadingSpinner fullPage />;

  const errorMessage = healthError instanceof Error ? healthError.message : 'Failed to load operations data';

  const visibleActions = (overload?.suggested_actions ?? []).filter(
    (_, i) => !dismissedActions.has(String(i))
  );

  const priorities = priorityData?.priorities ?? [];

  return (
    <div className="space-y-8">
      {/* Page header */}
      <div>
        <h1 className="text-2xl font-bold text-white">Operations Command Center</h1>
        <p className="text-sm text-slate-400 mt-1">
          Live health score, capacity status, ticket priorities, and planning horizons
        </p>
      </div>

      {healthError && <ErrorBanner message={errorMessage} />}

      {/* ── Top Bar: Health + Overload ── */}
      <div className="bg-white/5 rounded-xl border border-white/10 p-6 shadow-sm">
        {healthLoading || overloadLoading ? (
          <div className="flex justify-center py-6">
            <LoadingSpinner />
          </div>
        ) : (
          <div className="flex flex-wrap items-center gap-8">
            {/* Circular gauge */}
            {health && <HealthGauge score={Math.round(health.overall_score ?? 0)} status={health.status ?? 'unknown'} />}

            {/* Sub-scores */}
            {health && (
              <div className="flex flex-wrap gap-6">
                <SubScorePill label="Kitchen" score={Math.round(health.kitchen_score ?? 0)} />
                <SubScorePill label="Tickets" score={Math.round(health.ticket_score ?? 0)} />
                <SubScorePill label="Staff" score={Math.round(health.staff_score ?? 0)} />
                <SubScorePill label="Financial" score={Math.round(health.financial_score ?? 0)} />
                <SubScorePill label="Inventory" score={Math.round(health.inventory_score ?? 0)} />
              </div>
            )}

            {/* Separator */}
            <div className="hidden sm:block h-16 w-px bg-white/10" />

            {/* Overload indicator */}
            {overload && (
              <div className="flex flex-col gap-1">
                <div className="flex items-center gap-2">
                  <OverloadDot severity={overload.severity} />
                  <span className="text-sm font-semibold text-slate-200">
                    {overload.is_overloaded ? 'Overloaded' : 'Normal capacity'}
                  </span>
                  <StatusBadge variant={severityVariant(overload.severity)}>
                    {(overload.capacity_pct ?? 0).toFixed(0)}%
                  </StatusBadge>
                </div>
                <span className="text-xs text-slate-300 ml-5">
                  {overload.severity.charAt(0).toUpperCase() + overload.severity.slice(1)} severity
                </span>
              </div>
            )}
          </div>
        )}
      </div>

      {/* ── Suggested actions (dismissible) ── */}
      {visibleActions.length > 0 && (
        <div className="space-y-2">
          <h2 className="text-sm font-semibold text-slate-400 uppercase tracking-wide">Suggested Actions</h2>
          {(overload?.suggested_actions ?? []).map((action, i) => {
            if (dismissedActions.has(String(i))) return null;
            return (
              <div
                key={i}
                className="flex items-start justify-between gap-4 bg-amber-50 border border-amber-200 rounded-lg px-4 py-3"
              >
                <div className="flex items-start gap-3">
                  <AlertTriangle className="h-4 w-4 text-amber-600 mt-0.5 shrink-0" />
                  <div>
                    <p className="text-sm font-semibold text-amber-800">{action.action_type.replace(/_/g, ' ')}</p>
                    <p className="text-sm text-amber-700">{action.description}</p>
                    {action.impact && (
                      <p className="text-xs text-amber-600 mt-0.5">Impact: {action.impact}</p>
                    )}
                  </div>
                </div>
                <button
                  onClick={() => setDismissedActions((prev) => new Set([...prev, String(i)]))}
                  className="text-amber-400 hover:text-amber-600 transition-colors shrink-0"
                >
                  <X className="h-4 w-4" />
                </button>
              </div>
            );
          })}
        </div>
      )}

      {/* ── Section 1: Real-Time ── */}
      <div className="space-y-4">
        <h2 className="text-lg font-semibold text-white">Real-Time Status</h2>

        {realtimeLoading ? (
          <div className="flex justify-center py-6"><LoadingSpinner /></div>
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
            <KPICard
              label="Active Tickets"
              value={realtime ? String(realtime.active_tickets) : '—'}
              icon={Activity}
              iconColor="text-orange-600"
              bgTint="bg-orange-50"
            />
            <KPICard
              label="Avg Ticket Time"
              value={realtime ? `${(realtime.avg_ticket_time ?? 0).toFixed(1)} min` : '—'}
              icon={Clock}
              iconColor="text-purple-600"
              bgTint="bg-purple-50"
            />
            <KPICard
              label="Capacity"
              value={overload ? `${(overload.capacity_pct ?? 0).toFixed(0)}%` : '—'}
              icon={Zap}
              iconColor={
                overload && (overload.capacity_pct ?? 0) >= 90
                  ? 'text-red-600'
                  : overload && (overload.capacity_pct ?? 0) >= 70
                  ? 'text-amber-600'
                  : 'text-emerald-600'
              }
              bgTint={
                overload && (overload.capacity_pct ?? 0) >= 90
                  ? 'bg-red-50'
                  : overload && (overload.capacity_pct ?? 0) >= 70
                  ? 'bg-amber-50'
                  : 'bg-emerald-50'
              }
            />
          </div>
        )}

        {/* Ticket Priority Table */}
        <div>
          <h3 className="text-sm font-semibold text-slate-400 uppercase tracking-wide mb-2">Ticket Priority Queue</h3>
          <DataTable
            columns={priorityColumns}
            data={priorities}
            keyExtractor={(r) => r.ticket_id}
            isLoading={prioritiesLoading}
            emptyTitle="No active tickets"
            emptyDescription="All tickets are cleared or no data available."
          />
        </div>
      </div>

      {/* ── Section 2: Planning Horizons ── */}
      <div className="bg-white/5 rounded-xl border border-white/10 shadow-sm overflow-hidden">
        {/* Collapsible header */}
        <button
          onClick={() => setHorizonsOpen((v) => !v)}
          className="w-full flex items-center justify-between px-6 py-4 hover:bg-white/5 transition-colors"
        >
          <h2 className="text-lg font-semibold text-white">Planning Horizons</h2>
          {horizonsOpen ? (
            <ChevronUp className="h-5 w-5 text-slate-300" />
          ) : (
            <ChevronDown className="h-5 w-5 text-slate-300" />
          )}
        </button>

        {horizonsOpen && (
          <div className="border-t border-white/5">
            {/* Tabs */}
            <div className="flex border-b border-white/5 px-6 gap-1 pt-2">
              {HORIZON_TABS.map((tab) => (
                <button
                  key={tab.id}
                  onClick={() => setActiveTab(tab.id)}
                  className={`px-4 py-2 text-sm font-medium rounded-t transition-colors ${
                    activeTab === tab.id
                      ? 'text-orange-600 border-b-2 border-orange-500 bg-orange-50'
                      : 'text-slate-400 hover:text-slate-200'
                  }`}
                >
                  {tab.label}
                </button>
              ))}
            </div>

            <div className="p-6">
              {/* ── Shift Tab ── */}
              {activeTab === 'shift' && (
                shiftLoading ? (
                  <div className="flex justify-center py-8"><LoadingSpinner /></div>
                ) : shift ? (
                  <div className="space-y-4">
                    <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
                      <KPICard
                        label="Forecasted Covers"
                        value={String(shift.forecasted_covers)}
                        icon={Users}
                        iconColor="text-blue-600"
                        bgTint="bg-blue-50"
                      />
                      <KPICard
                        label="Scheduled Staff"
                        value={String(shift.scheduled_staff)}
                        icon={Users}
                        iconColor="text-emerald-600"
                        bgTint="bg-emerald-50"
                      />
                      <KPICard
                        label="Required Staff"
                        value={String(shift.required_staff)}
                        icon={Users}
                        iconColor="text-purple-600"
                        bgTint="bg-purple-50"
                      />
                      <KPICard
                        label="Expected Revenue"
                        value={dollars(shift.expected_revenue)}
                        icon={DollarSign}
                        iconColor="text-green-600"
                        bgTint="bg-green-50"
                      />
                    </div>

                    {/* Staff gap indicator */}
                    <div
                      className={`flex items-center gap-3 px-4 py-3 rounded-lg border ${
                        shift.staff_gap <= 0
                          ? 'bg-emerald-50 border-emerald-200'
                          : 'bg-red-50 border-red-200'
                      }`}
                    >
                      {shift.staff_gap <= 0 ? (
                        <CheckCircle className="h-5 w-5 text-emerald-600" />
                      ) : (
                        <AlertTriangle className="h-5 w-5 text-red-600" />
                      )}
                      <span
                        className={`text-sm font-semibold ${
                          shift.staff_gap <= 0 ? 'text-emerald-700' : 'text-red-700'
                        }`}
                      >
                        {shift.staff_gap <= 0
                          ? `Staffing covered (+${Math.abs(shift.staff_gap)} buffer)`
                          : `Staff gap: ${shift.staff_gap} needed`}
                      </span>
                    </div>
                  </div>
                ) : (
                  <p className="text-center text-slate-300 py-8 text-sm">No shift data available</p>
                )
              )}

              {/* ── Daily Tab ── */}
              {activeTab === 'daily' && (
                dailyLoading ? (
                  <div className="flex justify-center py-8"><LoadingSpinner /></div>
                ) : daily ? (
                  <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
                    <KPICard
                      label="Prep Items"
                      value={String(daily.prep_items)}
                      icon={Package}
                      iconColor="text-amber-600"
                      bgTint="bg-amber-50"
                    />
                    <KPICard
                      label="Expected Deliveries"
                      value={String(daily.expected_deliveries)}
                      icon={Truck}
                      iconColor="text-blue-600"
                      bgTint="bg-blue-50"
                    />
                    <KPICard
                      label="Scheduled Shifts"
                      value={String(daily.scheduled_shifts)}
                      icon={Users}
                      iconColor="text-purple-600"
                      bgTint="bg-purple-50"
                    />
                    <KPICard
                      label="Forecasted Revenue"
                      value={dollars(daily.forecasted_revenue)}
                      icon={DollarSign}
                      iconColor="text-green-600"
                      bgTint="bg-green-50"
                    />
                  </div>
                ) : (
                  <p className="text-center text-slate-300 py-8 text-sm">No daily data available</p>
                )
              )}

              {/* ── Weekly Tab ── */}
              {activeTab === 'weekly' && (
                weeklyLoading ? (
                  <div className="flex justify-center py-8"><LoadingSpinner /></div>
                ) : weekly ? (
                  <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
                    <KPICard
                      label="Total Scheduled Hours"
                      value={`${(weekly.total_hours ?? 0).toFixed(0)}h`}
                      icon={Clock}
                      iconColor="text-blue-600"
                      bgTint="bg-blue-50"
                    />
                    <KPICard
                      label="Pending POs"
                      value={String(weekly.pending_pos)}
                      icon={Package}
                      iconColor="text-amber-600"
                      bgTint="bg-amber-50"
                    />
                    <KPICard
                      label="Projected Labor Cost"
                      value={dollars(weekly.projected_labor_cost)}
                      icon={Users}
                      iconColor="text-red-600"
                      bgTint="bg-red-50"
                    />
                    <KPICard
                      label="Projected Revenue"
                      value={dollars(weekly.projected_revenue)}
                      icon={DollarSign}
                      iconColor="text-green-600"
                      bgTint="bg-green-50"
                    />
                  </div>
                ) : (
                  <p className="text-center text-slate-300 py-8 text-sm">No weekly data available</p>
                )
              )}

              {/* ── Strategic Tab ── */}
              {activeTab === 'strategic' && (
                strategicLoading ? (
                  <div className="flex justify-center py-8"><LoadingSpinner /></div>
                ) : strategic ? (
                  <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
                    {/* Revenue card with delta */}
                    <div className="bg-white/5 rounded-xl border border-white/10 p-5 shadow-sm flex items-start gap-4">
                      <div className="bg-green-50 p-3 rounded-lg">
                        <DollarSign className="h-6 w-6 text-green-600" />
                      </div>
                      <div>
                        <p className="text-sm text-slate-400">30-Day Revenue</p>
                        <p className="text-2xl font-bold text-white mt-0.5">{dollars(strategic.revenue_30d)}</p>
                        {(strategic.revenue_delta_pct ?? 0) !== 0 && (
                          <span
                            className={`text-xs font-semibold mt-1 inline-block ${
                              (strategic.revenue_delta_pct ?? 0) >= 0 ? 'text-emerald-600' : 'text-red-600'
                            }`}
                          >
                            {(strategic.revenue_delta_pct ?? 0) >= 0 ? '+' : ''}
                            {(strategic.revenue_delta_pct ?? 0).toFixed(1)}% vs prior period
                          </span>
                        )}
                      </div>
                    </div>

                    <KPICard
                      label="30-Day COGS"
                      value={dollars(strategic.cogs_30d)}
                      icon={Package}
                      iconColor="text-amber-600"
                      bgTint="bg-amber-50"
                    />

                    {/* Labor cost % with trend */}
                    <div className="bg-white/5 rounded-xl border border-white/10 p-5 shadow-sm flex items-start gap-4">
                      <div className="bg-red-50 p-3 rounded-lg">
                        <Users className="h-6 w-6 text-red-600" />
                      </div>
                      <div>
                        <p className="text-sm text-slate-400">Labor Cost %</p>
                        <p className="text-2xl font-bold text-white mt-0.5">
                          {(strategic.labor_cost_pct ?? 0).toFixed(1)}%
                        </p>
                        {strategic.labor_trend && (
                          <span className="text-xs text-slate-300 mt-1 inline-block capitalize">
                            Trend: {strategic.labor_trend}
                          </span>
                        )}
                      </div>
                    </div>
                  </div>
                ) : (
                  <p className="text-center text-slate-300 py-8 text-sm">No strategic data available</p>
                )
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
