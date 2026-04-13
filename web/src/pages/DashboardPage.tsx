import { useState, useMemo, useEffect } from 'react';
import {
  LineChart,
  Line,
  PieChart,
  Pie,
  Cell,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts';
import {
  DollarSign,
  ShoppingBag,
  Target,
  Heart,
  AlertTriangle,
  ChefHat,
  TrendingUp,
  Users,
  Clock,
  CheckCircle,
} from 'lucide-react';
import { useLocationStore } from '../stores/location';
import { usePnL, usePeriodComparison } from '../hooks/useFinancial';
import { useAlertQueue, useAlertCount, useAcknowledgeAlert } from '../hooks/useAlerts';
import { useHealth, useOperationsHourly } from '../hooks/useOperations';
import { useLaborSummary, useProfiles } from '../hooks/useLabor';
import { useCapacity } from '../hooks/useKitchen';
import { useMenuScores } from '../hooks/useMenuScoring';

// ── Helpers ───────────────────────────────────────────────────────────────────

function fmtEGP(piasters: number): string {
  return `EGP ${(piasters / 100).toLocaleString('en-US', {
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  })}`;
}

function fmtEGPShort(piasters: number): string {
  const val = piasters / 100;
  if (val >= 1_000_000) return `EGP ${(val / 1_000_000).toFixed(1)}M`;
  if (val >= 1_000) return `EGP ${(val / 1_000).toFixed(1)}K`;
  return `EGP ${val.toFixed(0)}`;
}

function timeAgo(iso: string): string {
  const diff = Date.now() - new Date(iso).getTime();
  const mins = Math.floor(diff / 60_000);
  if (mins < 1) return 'just now';
  if (mins < 60) return `${mins}m ago`;
  const hrs = Math.floor(mins / 60);
  if (hrs < 24) return `${hrs}h ago`;
  return `${Math.floor(hrs / 24)}d ago`;
}

function getInitials(name: string): string {
  return name
    .split(' ')
    .map((w) => w[0])
    .join('')
    .slice(0, 2)
    .toUpperCase();
}

// ── Hourly data formatter (from API response) ────────────────────────────────

function formatHourLabel(hour: number): string {
  if (hour < 12) return `${hour}am`;
  if (hour === 12) return '12pm';
  return `${hour - 12}pm`;
}

// ── Classification config ─────────────────────────────────────────────────────

const CLASS_COLORS: Record<string, string> = {
  powerhouse: 'bg-emerald-500/20 text-emerald-300 border border-emerald-500/30',
  hidden_gem: 'bg-violet-500/20 text-violet-300 border border-violet-500/30',
  crowd_pleaser: 'bg-blue-500/20 text-blue-300 border border-blue-500/30',
  workhorse: 'bg-slate-500/20 text-slate-300 border border-slate-500/30',
  question_mark: 'bg-amber-500/20 text-amber-300 border border-amber-500/30',
};

const CLASS_LABEL: Record<string, string> = {
  powerhouse: 'Powerhouse',
  hidden_gem: 'Hidden Gem',
  crowd_pleaser: 'Crowd Pleaser',
  workhorse: 'Workhorse',
  question_mark: 'Question Mark',
};

// ── Channel donut colors ──────────────────────────────────────────────────────

const CHANNEL_COLORS: Record<string, string> = {
  dine_in: '#10b981',
  takeout: '#3b82f6',
  delivery: '#f59e0b',
};

const CHANNEL_LABELS: Record<string, string> = {
  dine_in: 'Dine-in',
  takeout: 'Takeout',
  delivery: 'Delivery',
};

// ── Station load color ────────────────────────────────────────────────────────

function loadColor(pct: number): string {
  if (pct >= 80) return 'bg-red-500';
  if (pct >= 50) return 'bg-amber-500';
  return 'bg-emerald-500';
}

function loadTextColor(pct: number): string {
  if (pct >= 80) return 'text-red-400';
  if (pct >= 50) return 'text-amber-400';
  return 'text-emerald-400';
}

// ── Sub-components ────────────────────────────────────────────────────────────

function DeltaBadge({ delta, label = 'vs last wk' }: { delta: number | null; label?: string }) {
  if (delta == null) return null;
  const positive = delta >= 0;
  return (
    <span
      className={`inline-flex items-center gap-0.5 text-xs font-semibold ${
        positive ? 'text-emerald-400' : 'text-red-400'
      }`}
    >
      {positive ? '↑' : '↓'} {Math.abs(delta).toFixed(1)}%
      <span className="text-slate-300 font-normal ml-1">{label}</span>
    </span>
  );
}

function Card({ children, className = '' }: { children: React.ReactNode; className?: string }) {
  return (
    <div
      className={`bg-white/5 border border-white/10 rounded-2xl p-5 ${className}`}
    >
      {children}
    </div>
  );
}

function SectionLabel({ children }: { children: React.ReactNode }) {
  return (
    <p className="text-xs font-bold text-slate-300 uppercase tracking-wider mb-3">
      {children}
    </p>
  );
}

// ── Custom Recharts tooltip ───────────────────────────────────────────────────

function ChartTooltip({ active, payload, label }: any) {
  if (!active || !payload?.length) return null;
  return (
    <div className="bg-slate-800 border border-white/10 rounded-xl px-3 py-2 text-xs shadow-xl">
      <p className="text-slate-400 mb-1">{label}</p>
      {payload.map((p: any) => (
        <p key={p.name} style={{ color: p.color }}>
          {p.name}: EGP {p.value?.toLocaleString()}
        </p>
      ))}
    </div>
  );
}

// ── Main Page ─────────────────────────────────────────────────────────────────

export default function DashboardPage() {
  const locationId = useLocationStore((s) => s.selectedLocationId);
  const locations = useLocationStore((s) => s.locations);
  const selectedLocation = locations.find((l) => l.id === locationId);

  const { data: pnl, dataUpdatedAt: pnlUpdatedAt } = usePnL(locationId);
  const { data: alertCountData } = useAlertCount(locationId);
  const { data: alerts, dataUpdatedAt: alertsUpdatedAt } = useAlertQueue(locationId, { limit: 10 });
  const { data: health } = useHealth(locationId);
  const { data: labor } = useLaborSummary(locationId);
  const { data: profiles } = useProfiles(locationId);
  const { data: capacity } = useCapacity(locationId);
  const { data: menuScores } = useMenuScores(locationId);
  const { data: periodComp } = usePeriodComparison(locationId);
  const { mutate: acknowledgeAlert } = useAcknowledgeAlert();

  // Hourly data: today
  const today = new Date().toISOString().slice(0, 10);
  const yesterday = new Date(Date.now() - 86_400_000).toISOString().slice(0, 10);
  const { data: hourlyToday } = useOperationsHourly(locationId, today, today);
  const { data: hourlyYesterday } = useOperationsHourly(locationId, yesterday, yesterday);

  const [hoveredAlert, setHoveredAlert] = useState<string | null>(null);

  // ── Last data update indicator ─────────────────────────────────────────────

  const lastUpdatedAt = Math.max(pnlUpdatedAt || 0, alertsUpdatedAt || 0);
  const [, setTick] = useState(0);
  // Re-render every 10s to keep "Updated Xs ago" fresh
  useEffect(() => {
    const id = setInterval(() => setTick((t) => t + 1), 10_000);
    return () => clearInterval(id);
  }, []);
  const updatedAgoText = useMemo(() => {
    if (!lastUpdatedAt) return 'Loading...';
    const seconds = Math.floor((Date.now() - lastUpdatedAt) / 1000);
    if (seconds < 5) return 'Updated just now';
    if (seconds < 60) return `Updated ${seconds}s ago`;
    return `Updated ${Math.floor(seconds / 60)}m ago`;
  }, [lastUpdatedAt]);

  // ── Derived values ─────────────────────────────────────────────────────────

  const revenue = pnl?.net_revenue ?? 0;
  const orderCount = pnl?.check_count ?? 0;
  const avgCheck = pnl?.avg_check_size ?? 0;
  const healthScore = Math.round(health?.overall_score ?? 0);
  const healthStatus = health?.status ?? 'unknown';
  const totalAlerts = alertCountData?.count ?? 0;

  // ── Deltas from period comparison (vs last week) ───────────────────────────

  const deltas = useMemo(() => {
    if (!periodComp?.current || !periodComp?.last_week) {
      return { revenue: null, orders: null, avgCheck: null, health: null };
    }
    const cur = periodComp.current;
    const prev = periodComp.last_week;
    const pctChange = (a: number, b: number) =>
      b !== 0 ? ((a - b) / Math.abs(b)) * 100 : null;
    return {
      revenue: pctChange(cur.net_revenue, prev.net_revenue),
      orders: pctChange(cur.check_count, prev.check_count),
      avgCheck: pctChange(cur.avg_check_size, prev.avg_check_size),
      health: null as number | null, // Health score has no historical comparison
    };
  }, [periodComp]);

  // ── Hourly chart data from API ─────────────────────────────────────────────

  const hourlyData = useMemo(() => {
    const todayHours = hourlyToday?.hourly ?? [];
    const yesterdayHours = hourlyYesterday?.hourly ?? [];

    if (todayHours.length === 0 && yesterdayHours.length === 0) return [];

    // Build a map from hour -> data for both days
    const todayMap = new Map(todayHours.map((h: any) => [h.hour, h.revenue / 100]));
    const yesterdayMap = new Map(yesterdayHours.map((h: any) => [h.hour, h.revenue / 100]));

    // Build cumulative revenue by hour (6am–midnight)
    let todayCum = 0;
    let yestCum = 0;
    const result: { hour: string; today: number; yesterday: number }[] = [];

    for (let h = 6; h <= 23; h++) {
      todayCum += todayMap.get(h) ?? 0;
      yestCum += yesterdayMap.get(h) ?? 0;
      result.push({
        hour: formatHourLabel(h),
        today: Math.round(todayCum),
        yesterday: Math.round(yestCum),
      });
    }

    return result;
  }, [hourlyToday, hourlyYesterday]);

  // ── Activity feed from live API data ───────────────────────────────────────

  const activityFeed = useMemo(() => {
    const items: { id: string; icon: string; text: string; ts: string; border: string }[] = [];

    // Add alerts as activity items
    (alerts ?? []).forEach((alert: any) => {
      const severityIcon = alert.severity === 'critical' ? '!!' : alert.severity === 'warning' ? '!' : 'i';
      const severityBorder =
        alert.severity === 'critical'
          ? 'border-red-500'
          : alert.severity === 'warning'
          ? 'border-amber-500'
          : 'border-blue-500';
      items.push({
        id: `alert-${alert.alert_id}`,
        icon: severityIcon === '!!' ? '!!' : severityIcon === '!' ? '!' : 'i',
        text: `Alert: ${alert.title}`,
        ts: alert.created_at ?? new Date().toISOString(),
        border: severityBorder,
      });
    });

    // Add revenue milestone from PnL
    if (pnl && revenue > 0) {
      items.push({
        id: 'pnl-revenue',
        icon: '$',
        text: `Revenue today: ${fmtEGP(revenue)} across ${orderCount} orders (avg check ${fmtEGP(avgCheck)})`,
        ts: pnl.period_end ?? new Date().toISOString(),
        border: 'border-emerald-500',
      });
      if (pnl.gross_margin > 0) {
        items.push({
          id: 'pnl-margin',
          icon: '%',
          text: `Gross margin at ${pnl.gross_margin.toFixed(1)}% — Gross profit ${fmtEGP(pnl.gross_profit)}`,
          ts: pnl.period_end ?? new Date().toISOString(),
          border: 'border-emerald-500',
        });
      }
    }

    // Add staff on shift from labor data
    if (labor && (labor as any).employee_count > 0) {
      items.push({
        id: 'labor-shift',
        icon: '#',
        text: `${(labor as any).employee_count} employees on shift — labor cost ${((labor as any).labor_cost_pct ?? 0).toFixed(1)}% of revenue`,
        ts: new Date().toISOString(),
        border: 'border-blue-500',
      });
    }

    // Add kitchen capacity from capacity data
    if (capacity && (capacity as any).total_capacity_pct != null) {
      const capPct = (capacity as any).total_capacity_pct;
      items.push({
        id: 'kitchen-cap',
        icon: capPct >= 80 ? '!!' : 'i',
        text: `Kitchen at ${capPct.toFixed(0)}% capacity — ${(capacity as any).active_tickets ?? 0} active tickets`,
        ts: new Date().toISOString(),
        border: capPct >= 80 ? 'border-red-500' : capPct >= 50 ? 'border-amber-500' : 'border-emerald-500',
      });
    }

    // Sort by time descending
    items.sort((a, b) => new Date(b.ts).getTime() - new Date(a.ts).getTime());

    return items;
  }, [alerts, pnl, revenue, orderCount, avgCheck, labor, capacity]);

  const channelData = useMemo(() => {
    const ch = pnl?.by_channel ?? {};
    return Object.entries(ch)
      .filter(([, v]: any) => (v?.net_revenue ?? 0) > 0)
      .map(([key, v]: any) => ({
        name: CHANNEL_LABELS[key] ?? key,
        key,
        value: v?.check_count ?? 0,
        revenue: v?.net_revenue ?? 0,
      }));
  }, [pnl]);

  const totalChannelOrders = channelData.reduce((s, d) => s + d.value, 0);

  const topMenuItems = useMemo(() => {
    const items = (menuScores?.items ?? []) as any[];
    return [...items]
      .sort((a, b) => (b.velocity_score ?? 0) - (a.velocity_score ?? 0))
      .slice(0, 5);
  }, [menuScores]);

  const staffProfiles = (profiles?.profiles ?? []) as any[];

  const laborCostPct = labor?.labor_cost_pct ?? 0;
  const foodCostPct =
    revenue > 0 ? ((pnl?.cogs ?? 0) / revenue) * 100 : 0;
  const primeCostPct = foodCostPct + laborCostPct;

  const stations = (capacity?.stations ?? []) as any[];

  return (
    <div className="min-h-full">

      {/* ── Page Header ───────────────────────────────────────────────── */}
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">
            {selectedLocation?.name ?? 'Branch Dashboard'}
          </h1>
          <p className="text-sm text-slate-300 mt-0.5">
            Live command center · {new Date().toLocaleDateString('en-US', {
              weekday: 'long', month: 'long', day: 'numeric',
            })}
          </p>
        </div>
        <div className="flex items-center gap-2">
          <span className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse" />
          <span className="text-xs text-slate-400">{updatedAgoText}</span>
        </div>
      </div>

      {/* ── Row 1: Hero KPI Strip ─────────────────────────────────────── */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">

        {/* Revenue */}
        <Card>
          <div className="flex items-start justify-between mb-3">
            <div className="p-2 rounded-xl bg-emerald-500/10">
              <DollarSign className="w-5 h-5 text-emerald-400" />
            </div>
            <DeltaBadge delta={deltas.revenue} />
          </div>
          <SectionLabel>Revenue Today</SectionLabel>
          <p className="text-3xl font-bold text-white leading-none">
            {fmtEGPShort(revenue)}
          </p>
        </Card>

        {/* Orders */}
        <Card>
          <div className="flex items-start justify-between mb-3">
            <div className="p-2 rounded-xl bg-blue-500/10">
              <ShoppingBag className="w-5 h-5 text-blue-400" />
            </div>
            <DeltaBadge delta={deltas.orders} />
          </div>
          <SectionLabel>Orders Today</SectionLabel>
          <p className="text-3xl font-bold text-white leading-none">
            {orderCount.toLocaleString()}
          </p>
        </Card>

        {/* Avg Check */}
        <Card>
          <div className="flex items-start justify-between mb-3">
            <div className="p-2 rounded-xl bg-amber-500/10">
              <Target className="w-5 h-5 text-amber-400" />
            </div>
            <DeltaBadge delta={deltas.avgCheck} />
          </div>
          <SectionLabel>Avg Check Size</SectionLabel>
          <p className="text-3xl font-bold text-white leading-none">
            {fmtEGP(avgCheck)}
          </p>
        </Card>

        {/* Health */}
        <Card>
          <div className="flex items-start justify-between mb-3">
            <div className="p-2 rounded-xl bg-pink-500/10">
              <Heart className="w-5 h-5 text-pink-400" />
            </div>
            <DeltaBadge delta={deltas.health} />
          </div>
          <SectionLabel>Health Score</SectionLabel>
          <p className="text-3xl font-bold text-white leading-none">
            {healthScore}
            <span className="text-lg text-slate-300 font-normal">/100</span>
          </p>
          <p className="text-xs text-slate-400 mt-1 capitalize">{healthStatus}</p>
        </Card>
      </div>

      {/* ── Row 2: Charts ─────────────────────────────────────────────── */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-6">

        {/* Revenue by Hour */}
        <Card>
          <div className="flex items-center gap-2 mb-4">
            <TrendingUp className="w-4 h-4 text-emerald-400" />
            <p className="text-sm font-bold text-white uppercase tracking-wider">
              Revenue by Hour
            </p>
          </div>
          {hourlyData.length > 0 ? (
          <ResponsiveContainer width="100%" height={220}>
            <LineChart data={hourlyData} margin={{ top: 4, right: 8, left: -20, bottom: 0 }}>
              <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.05)" />
              <XAxis
                dataKey="hour"
                tick={{ fill: '#64748b', fontSize: 11 }}
                axisLine={false}
                tickLine={false}
                interval={2}
              />
              <YAxis
                tick={{ fill: '#64748b', fontSize: 11 }}
                axisLine={false}
                tickLine={false}
                tickFormatter={(v) => `${(v / 1000).toFixed(0)}K`}
              />
              <Tooltip content={<ChartTooltip />} />
              <Legend
                wrapperStyle={{ fontSize: 12, color: '#94a3b8' }}
                formatter={(v) => v === 'today' ? 'Today' : 'Yesterday'}
              />
              <Line
                type="monotone"
                dataKey="today"
                stroke="#10b981"
                strokeWidth={2}
                dot={false}
                name="today"
              />
              <Line
                type="monotone"
                dataKey="yesterday"
                stroke="#475569"
                strokeWidth={1.5}
                strokeDasharray="4 3"
                dot={false}
                name="yesterday"
              />
            </LineChart>
          </ResponsiveContainer>
          ) : (
            <div className="flex items-center justify-center h-[220px] text-slate-400 text-sm">
              No hourly data available yet
            </div>
          )}
        </Card>

        {/* Channel Mix Donut */}
        <Card>
          <div className="flex items-center gap-2 mb-4">
            <ShoppingBag className="w-4 h-4 text-blue-400" />
            <p className="text-sm font-bold text-white uppercase tracking-wider">
              Channel Mix
            </p>
          </div>
          {channelData.length > 0 ? (
            <div className="flex items-center gap-4">
              <div className="flex-1">
                <ResponsiveContainer width="100%" height={180}>
                  <PieChart>
                    <Pie
                      data={channelData}
                      cx="50%"
                      cy="50%"
                      innerRadius={55}
                      outerRadius={80}
                      paddingAngle={3}
                      dataKey="value"
                    >
                      {channelData.map((entry, idx) => (
                        <Cell
                          key={entry.key}
                          fill={CHANNEL_COLORS[entry.key] ?? `hsl(${idx * 120}, 60%, 55%)`}
                        />
                      ))}
                    </Pie>
                    <Tooltip
                      formatter={(v: any) => [`${v} orders`, '']}
                      contentStyle={{
                        background: '#1e293b',
                        border: '1px solid rgba(255,255,255,0.1)',
                        borderRadius: 8,
                        color: '#fff',
                        fontSize: 12,
                      }}
                    />
                  </PieChart>
                </ResponsiveContainer>
                {/* Center label via absolute positioning trick */}
              </div>
              <div className="shrink-0 space-y-3 pr-2">
                <p className="text-xs text-slate-300 uppercase tracking-wider mb-1">Breakdown</p>
                {channelData.map((d) => {
                  const pct = totalChannelOrders > 0
                    ? ((d.value / totalChannelOrders) * 100).toFixed(1)
                    : '0.0';
                  return (
                    <div key={d.key} className="flex items-center gap-2">
                      <span
                        className="w-2.5 h-2.5 rounded-full shrink-0"
                        style={{ background: CHANNEL_COLORS[d.key] ?? '#888' }}
                      />
                      <div>
                        <p className="text-xs text-white font-medium">{d.name}</p>
                        <p className="text-xs text-slate-400">{d.value} orders · {pct}%</p>
                      </div>
                    </div>
                  );
                })}
                <div className="pt-2 border-t border-white/10">
                  <p className="text-xs text-slate-300">Total Orders</p>
                  <p className="text-lg font-bold text-white">{totalChannelOrders}</p>
                </div>
              </div>
            </div>
          ) : (
            <div className="flex items-center justify-center h-40 text-slate-400 text-sm">
              No channel data yet
            </div>
          )}
        </Card>
      </div>

      {/* ── Row 3: Alerts / Kitchen / Top Sellers ──────────────────────── */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 mb-6">

        {/* AI Alerts */}
        <Card>
          <div className="flex items-center justify-between mb-4">
            <div className="flex items-center gap-2">
              <AlertTriangle className="w-4 h-4 text-red-400" />
              <p className="text-sm font-bold text-white uppercase tracking-wider">AI Alerts</p>
              {totalAlerts > 0 && (
                <span className="w-2 h-2 rounded-full bg-red-500 animate-pulse" />
              )}
            </div>
            <span className="text-xs font-bold text-red-400 bg-red-500/10 px-2 py-0.5 rounded-full">
              {totalAlerts}
            </span>
          </div>

          <div className="space-y-2">
            {(alerts ?? []).length > 0 ? (
              (alerts ?? []).map((alert: any) => (
                <div
                  key={alert.alert_id}
                  className="group relative flex items-start gap-2.5 p-2.5 rounded-xl bg-white/3 hover:bg-white/8 transition-colors cursor-default"
                  onMouseEnter={() => setHoveredAlert(alert.alert_id)}
                  onMouseLeave={() => setHoveredAlert(null)}
                >
                  <span className="shrink-0 text-sm leading-none mt-0.5">
                    {alert.severity === 'critical' ? '🔴' : alert.severity === 'warning' ? '🟡' : '🔵'}
                  </span>
                  <div className="flex-1 min-w-0">
                    <p className="text-xs font-medium text-white truncate">{alert.title}</p>
                    <p className="text-xs text-slate-300 mt-0.5">
                      {timeAgo(alert.created_at ?? new Date().toISOString())}
                    </p>
                  </div>
                  {hoveredAlert === alert.alert_id && (
                    <button
                      onClick={() => acknowledgeAlert(alert.alert_id)}
                      className="shrink-0 flex items-center gap-1 text-xs text-emerald-400 bg-emerald-500/10 hover:bg-emerald-500/20 px-2 py-1 rounded-lg transition-colors"
                    >
                      <CheckCircle className="w-3 h-3" />
                      Ack
                    </button>
                  )}
                </div>
              ))
            ) : (
              <div className="flex items-center justify-center py-8 text-slate-400 text-sm">
                All clear — no alerts
              </div>
            )}
          </div>

          {totalAlerts > 5 && (
            <p className="mt-3 text-xs text-slate-300 text-right">
              View all {totalAlerts} alerts →
            </p>
          )}
        </Card>

        {/* Kitchen Pulse */}
        <Card>
          <div className="flex items-center justify-between mb-4">
            <div className="flex items-center gap-2">
              <ChefHat className="w-4 h-4 text-amber-400" />
              <p className="text-sm font-bold text-white uppercase tracking-wider">Kitchen Status</p>
            </div>
            {capacity?.total_capacity_pct != null && (
              <span
                className={`text-xs font-bold px-2 py-0.5 rounded-full ${
                  (capacity.total_capacity_pct ?? 0) >= 80
                    ? 'bg-red-500/10 text-red-400'
                    : (capacity.total_capacity_pct ?? 0) >= 50
                    ? 'bg-amber-500/10 text-amber-400'
                    : 'bg-emerald-500/10 text-emerald-400'
                }`}
              >
                {capacity.total_capacity_pct?.toFixed(0) ?? 0}% cap
              </span>
            )}
          </div>

          {stations.length > 0 ? (
            <div className="space-y-3">
              {stations.map((st: any, i: number) => {
                const pct = Math.min(st.load_pct ?? 0, 100);
                return (
                  <div key={st.station_id ?? i}>
                    <div className="flex items-center justify-between mb-1">
                      <p className="text-xs font-medium text-white">{st.name ?? `Station ${i + 1}`}</p>
                      <p className={`text-xs font-bold ${loadTextColor(pct)}`}>{pct.toFixed(0)}%</p>
                    </div>
                    <div className="h-1.5 bg-white/10 rounded-full overflow-hidden">
                      <div
                        className={`h-full rounded-full transition-all duration-700 ${loadColor(pct)}`}
                        style={{ width: `${pct}%` }}
                      />
                    </div>
                  </div>
                );
              })}
              {(capacity?.active_tickets ?? 0) > 0 && (
                <div className="flex items-center gap-1.5 mt-3 pt-3 border-t border-white/10">
                  <Clock className="w-3.5 h-3.5 text-slate-400" />
                  <p className="text-xs text-slate-400">
                    {capacity?.active_tickets} active tickets in queue
                  </p>
                </div>
              )}
            </div>
          ) : (
            <div className="flex items-center justify-center py-8 text-slate-400 text-sm">
              No station data available
            </div>
          )}
        </Card>

        {/* Top Sellers */}
        <Card>
          <div className="flex items-center gap-2 mb-4">
            <TrendingUp className="w-4 h-4 text-violet-400" />
            <p className="text-sm font-bold text-white uppercase tracking-wider">Top Menu Items</p>
          </div>

          {topMenuItems.length > 0 ? (
            <div className="space-y-3">
              {topMenuItems.map((item: any, i: number) => {
                const score = item.velocity_score ?? 0;
                const cls = item.classification ?? 'workhorse';
                return (
                  <div key={item.menu_item_id ?? i} className="flex items-center gap-3">
                    <span className="text-xs font-bold text-slate-400 w-4 shrink-0">
                      #{i + 1}
                    </span>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2 mb-1">
                        <p className="text-xs font-medium text-white truncate">{item.name ?? 'Unknown'}</p>
                        <span
                          className={`shrink-0 text-[10px] font-semibold px-1.5 py-0.5 rounded-full ${
                            CLASS_COLORS[cls] ?? CLASS_COLORS.workhorse
                          }`}
                        >
                          {CLASS_LABEL[cls] ?? cls}
                        </span>
                      </div>
                      <div className="h-1 bg-white/10 rounded-full overflow-hidden">
                        <div
                          className="h-full rounded-full bg-violet-500"
                          style={{ width: `${Math.min(score * 10, 100)}%` }}
                        />
                      </div>
                    </div>
                    <span className="text-xs text-slate-400 shrink-0">{score.toFixed(1)}</span>
                  </div>
                );
              })}
            </div>
          ) : (
            <div className="flex items-center justify-center py-8 text-slate-400 text-sm">
              No menu score data yet
            </div>
          )}
        </Card>
      </div>

      {/* ── Row 4: Staff + Financial ──────────────────────────────────── */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-6">

        {/* Staff On Floor */}
        <Card>
          <div className="flex items-center justify-between mb-4">
            <div className="flex items-center gap-2">
              <Users className="w-4 h-4 text-blue-400" />
              <p className="text-sm font-bold text-white uppercase tracking-wider">Team On Shift</p>
            </div>
            <span className="text-xs font-bold text-blue-400 bg-blue-500/10 px-2 py-0.5 rounded-full">
              {staffProfiles.length > 0
                ? staffProfiles.length
                : (labor?.employee_count ?? 0)}{' '}
              staff
            </span>
          </div>

          {staffProfiles.length > 0 ? (
            <>
              <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
                {staffProfiles.slice(0, 8).map((emp: any, i: number) => {
                  const topStation =
                    emp.elu_ratings
                      ? Object.entries(emp.elu_ratings as Record<string, number>).sort(
                          ([, a], [, b]) => (b as number) - (a as number)
                        )[0]?.[0]
                      : null;
                  return (
                    <div
                      key={emp.employee_id ?? i}
                      className="flex flex-col items-center gap-1.5 p-2.5 rounded-xl bg-white/3 hover:bg-white/8 transition-colors"
                    >
                      <div className="w-9 h-9 rounded-full bg-gradient-to-br from-blue-500 to-violet-600 flex items-center justify-center shrink-0">
                        <span className="text-xs font-bold text-white">
                          {getInitials(emp.name ?? 'UN')}
                        </span>
                      </div>
                      <p className="text-xs font-medium text-white text-center leading-tight truncate w-full text-center">
                        {(emp.name ?? 'Unknown').split(' ')[0]}
                      </p>
                      {emp.role && (
                        <span className="text-[10px] text-slate-400 bg-white/5 px-1.5 py-0.5 rounded truncate max-w-full">
                          {emp.role}
                        </span>
                      )}
                      {topStation && (
                        <span className="text-[10px] text-blue-300">⭐ {topStation}</span>
                      )}
                    </div>
                  );
                })}
              </div>
              {staffProfiles.length > 8 && (
                <p className="text-xs text-slate-300 text-right mt-3">
                  +{staffProfiles.length - 8} more on shift
                </p>
              )}
            </>
          ) : (
            <div className="flex flex-col items-center justify-center py-8 gap-2">
              <Users className="w-8 h-8 text-slate-700" />
              <p className="text-slate-300 text-sm">
                {(labor?.employee_count ?? 0) > 0
                  ? `${labor?.employee_count} employees on shift today`
                  : 'No shift data available'}
              </p>
            </div>
          )}
        </Card>

        {/* P&L Snapshot */}
        <Card>
          <div className="flex items-center gap-2 mb-4">
            <DollarSign className="w-4 h-4 text-emerald-400" />
            <p className="text-sm font-bold text-white uppercase tracking-wider">P&L Summary</p>
          </div>

          <div className="space-y-4">
            {/* Food Cost % */}
            {(() => {
              const actual = foodCostPct;
              const target = 32;
              const delta = actual - target;
              const barPct = Math.min((actual / 60) * 100, 100);
              const good = delta <= 0;
              return (
                <div>
                  <div className="flex items-center justify-between mb-1.5">
                    <p className="text-xs font-medium text-white">Food Cost %</p>
                    <div className="flex items-center gap-2">
                      <span className="text-xs text-slate-300">Target 32%</span>
                      <span className={`text-xs font-bold ${good ? 'text-emerald-400' : 'text-red-400'}`}>
                        {actual.toFixed(1)}%
                        <span className="ml-1 font-normal">
                          ({good ? '↓' : '↑'}{Math.abs(delta).toFixed(1)}pp)
                        </span>
                      </span>
                    </div>
                  </div>
                  <div className="relative h-2 bg-white/10 rounded-full overflow-hidden">
                    <div
                      className={`h-full rounded-full ${good ? 'bg-emerald-500' : 'bg-red-500'}`}
                      style={{ width: `${barPct}%` }}
                    />
                    <div
                      className="absolute top-0 h-full w-0.5 bg-white/30"
                      style={{ left: `${(32 / 60) * 100}%` }}
                    />
                  </div>
                </div>
              );
            })()}

            {/* Labor Cost % */}
            {(() => {
              const actual = laborCostPct;
              const target = 28;
              const delta = actual - target;
              const barPct = Math.min((actual / 60) * 100, 100);
              const good = delta <= 0;
              return (
                <div>
                  <div className="flex items-center justify-between mb-1.5">
                    <p className="text-xs font-medium text-white">Labor Cost %</p>
                    <div className="flex items-center gap-2">
                      <span className="text-xs text-slate-300">Target 28%</span>
                      <span className={`text-xs font-bold ${good ? 'text-emerald-400' : 'text-red-400'}`}>
                        {actual.toFixed(1)}%
                        <span className="ml-1 font-normal">
                          ({good ? '↓' : '↑'}{Math.abs(delta).toFixed(1)}pp)
                        </span>
                      </span>
                    </div>
                  </div>
                  <div className="relative h-2 bg-white/10 rounded-full overflow-hidden">
                    <div
                      className={`h-full rounded-full ${good ? 'bg-emerald-500' : 'bg-red-500'}`}
                      style={{ width: `${barPct}%` }}
                    />
                    <div
                      className="absolute top-0 h-full w-0.5 bg-white/30"
                      style={{ left: `${(28 / 60) * 100}%` }}
                    />
                  </div>
                </div>
              );
            })()}

            {/* Prime Cost % */}
            {(() => {
              const actual = primeCostPct;
              const target = 60;
              const delta = actual - target;
              const barPct = Math.min((actual / 80) * 100, 100);
              const good = delta <= 0;
              return (
                <div>
                  <div className="flex items-center justify-between mb-1.5">
                    <p className="text-xs font-medium text-white">Prime Cost %</p>
                    <div className="flex items-center gap-2">
                      <span className="text-xs text-slate-300">Target 60%</span>
                      <span className={`text-xs font-bold ${good ? 'text-emerald-400' : 'text-red-400'}`}>
                        {actual.toFixed(1)}%
                        <span className="ml-1 font-normal">
                          ({good ? '↓' : '↑'}{Math.abs(delta).toFixed(1)}pp)
                        </span>
                      </span>
                    </div>
                  </div>
                  <div className="relative h-2 bg-white/10 rounded-full overflow-hidden">
                    <div
                      className={`h-full rounded-full ${good ? 'bg-emerald-500' : 'bg-amber-500'}`}
                      style={{ width: `${barPct}%` }}
                    />
                    <div
                      className="absolute top-0 h-full w-0.5 bg-white/30"
                      style={{ left: `${(60 / 80) * 100}%` }}
                    />
                  </div>
                </div>
              );
            })()}

            {/* Summary row */}
            <div className="pt-3 border-t border-white/10 grid grid-cols-3 gap-2">
              <div>
                <p className="text-xs text-slate-300 uppercase tracking-wider">COGS</p>
                <p className="text-sm font-bold text-white mt-0.5">{fmtEGPShort(pnl?.cogs ?? 0)}</p>
              </div>
              <div>
                <p className="text-xs text-slate-300 uppercase tracking-wider">Gross Profit</p>
                <p className="text-sm font-bold text-white mt-0.5">{fmtEGPShort(pnl?.gross_profit ?? 0)}</p>
              </div>
              <div>
                <p className="text-xs text-slate-300 uppercase tracking-wider">GM%</p>
                <p className="text-sm font-bold text-emerald-400 mt-0.5">
                  {(pnl?.gross_margin ?? 0).toFixed(1)}%
                </p>
              </div>
            </div>
          </div>
        </Card>
      </div>

      {/* ── Row 5: Activity Feed ──────────────────────────────────────── */}
      <Card>
        <div className="flex items-center gap-2 mb-4">
          <Clock className="w-4 h-4 text-slate-400" />
          <p className="text-sm font-bold text-white uppercase tracking-wider">Recent Activity</p>
        </div>

        {activityFeed.length > 0 ? (
        <div
          className="space-y-1 overflow-y-auto"
          style={{ maxHeight: 320 }}
        >
          {activityFeed.map((event) => {
            const iconMap: Record<string, string> = {
              '!!': '!!', '!': '!', 'i': 'i', '$': '$', '%': '%', '#': '#',
            };
            const iconColorMap: Record<string, string> = {
              '!!': 'bg-red-500/20 text-red-400',
              '!': 'bg-amber-500/20 text-amber-400',
              'i': 'bg-blue-500/20 text-blue-400',
              '$': 'bg-emerald-500/20 text-emerald-400',
              '%': 'bg-emerald-500/20 text-emerald-400',
              '#': 'bg-blue-500/20 text-blue-400',
            };
            return (
            <div
              key={event.id}
              className={`flex items-start gap-3 p-3 rounded-xl hover:bg-white/3 transition-colors border-l-2 ${event.border}`}
            >
              <span className={`shrink-0 mt-0.5 w-6 h-6 rounded-full flex items-center justify-center text-[10px] font-bold ${iconColorMap[event.icon] ?? 'bg-slate-500/20 text-slate-400'}`}>
                {iconMap[event.icon] ?? event.icon}
              </span>
              <div className="flex-1 min-w-0">
                <p className="text-xs text-slate-300 leading-relaxed">{event.text}</p>
              </div>
              <span className="text-xs text-slate-400 shrink-0 whitespace-nowrap">{timeAgo(event.ts)}</span>
            </div>
            );
          })}
        </div>
        ) : (
          <div className="flex items-center justify-center py-8 text-slate-400 text-sm">
            No recent activity
          </div>
        )}

        {/* Bottom fade gradient */}
        {activityFeed.length > 0 && (
        <div className="relative -mb-5 mx-0 h-8 bg-gradient-to-t from-slate-900/90 to-transparent rounded-b-2xl pointer-events-none" />
        )}
      </Card>
    </div>
  );
}
