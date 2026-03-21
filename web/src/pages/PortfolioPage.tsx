import { useEffect, useState, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  AreaChart,
  Area,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from 'recharts';
import {
  Activity,
  AlertTriangle,
  AlertCircle,
  Users,
  TrendingUp,
  MapPin,
  Zap,
} from 'lucide-react';
import { useLocationStore } from '../stores/location';
import { usePnL } from '../hooks/useFinancial';
import { useHealth } from '../hooks/useOperations';
import { useAlertQueue, useAlertCount } from '../hooks/useAlerts';
import { useLaborSummary } from '../hooks/useLabor';

// ── CSS Animation for ticker ──────────────────────────────────────────────────

const tickerStyle = `
@keyframes scroll-left {
  0% { transform: translateX(0); }
  100% { transform: translateX(-50%); }
}
.animate-scroll-left {
  animation: scroll-left 60s linear infinite;
}
`;

// ── AI Insights for ticker ────────────────────────────────────────────────────

const AI_INSIGHTS = [
  "📈 Ceviche Clásico sales up 23% this week at Nimbu El Gouna — consider featuring in social media",
  "⚠️ Beef Tenderloin cost spike detected — 3 menu items affected, margin impact: -2.1%",
  "🎯 Pisco Sour is the #1 margin contributor across all branches at 78% gross margin",
  "👥 5 VIP customers at risk of churning — combined monthly value: EGP 12,000",
  "🔄 Auto-PO generated for Sysco Egypt — 8 items below reorder point at Nimbu New Cairo",
  "📊 Nimbu Zayed outperforming chain average by 12% on food cost control",
  "🌡️ Kitchen capacity at Nimbu North Coast hit 92% during Friday dinner — consider overflow prep",
  "💡 Empanadas velocity up 22% after portion adjustment — reclassified to crowd_pleaser",
  "📦 Metro Market OTIF rate dropped to 72% — AI recommends shifting 30% to Seoudi Fresh",
  "🎉 Loyalty program: 15 active members, 24 transactions, EGP 3,200 in points issued",
];

// ── Helpers ──────────────────────────────────────────────────────────────────

function fmtEGP(cents: number) {
  return `EGP ${(cents / 100).toLocaleString('en-US', {
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  })}`;
}

function fmtPct(v: number) {
  return `${v.toFixed(1)}%`;
}

// ── Animated counter hook ────────────────────────────────────────────────────

function useCountUp(target: number, duration = 1200) {
  const [value, setValue] = useState(0);
  const frameRef = useRef<number>(0);
  useEffect(() => {
    if (target === 0) { setValue(0); return; }
    const start = performance.now();
    const tick = (now: number) => {
      const progress = Math.min((now - start) / duration, 1);
      const eased = 1 - Math.pow(1 - progress, 3);
      setValue(Math.round(eased * target));
      if (progress < 1) frameRef.current = requestAnimationFrame(tick);
    };
    frameRef.current = requestAnimationFrame(tick);
    return () => cancelAnimationFrame(frameRef.current);
  }, [target, duration]);
  return value;
}

// ── Health Circle ────────────────────────────────────────────────────────────

function HealthCircle({ score, size = 64 }: { score: number; size?: number }) {
  const [animated, setAnimated] = useState(0);

  useEffect(() => {
    const timer = setTimeout(() => setAnimated(score), 100);
    return () => clearTimeout(timer);
  }, [score]);

  const radius = (size / 2) - 5;
  const circumference = 2 * Math.PI * radius;
  const offset = circumference - (animated / 100) * circumference;

  const color =
    score >= 75 ? '#22c55e' :
    score >= 50 ? '#f59e0b' :
    '#ef4444';

  const trackColor =
    score >= 75 ? 'rgba(34,197,94,0.15)' :
    score >= 50 ? 'rgba(245,158,11,0.15)' :
    'rgba(239,68,68,0.15)';

  return (
    <div className="relative flex items-center justify-center" style={{ width: size, height: size }}>
      <svg width={size} height={size} className="-rotate-90">
        <circle
          cx={size / 2}
          cy={size / 2}
          r={radius}
          fill="none"
          stroke={trackColor}
          strokeWidth="5"
        />
        <circle
          cx={size / 2}
          cy={size / 2}
          r={radius}
          fill="none"
          stroke={color}
          strokeWidth="5"
          strokeLinecap="round"
          strokeDasharray={circumference}
          strokeDashoffset={offset}
          style={{ transition: 'stroke-dashoffset 1.2s cubic-bezier(0.4, 0, 0.2, 1)' }}
        />
      </svg>
      <div className="absolute inset-0 flex items-center justify-center">
        <span className="text-sm font-bold" style={{ color }}>{Math.round(score)}</span>
      </div>
    </div>
  );
}

// ── Sparkline placeholder data (7 days) ─────────────────────────────────────

function generateSparkData(seed: number) {
  return Array.from({ length: 7 }, (_, i) => ({
    day: i,
    value: Math.max(20, Math.round(60 + (Math.sin((seed + i) * 1.3) * 25) + Math.random() * 15)),
  }));
}

// ── Status Dot ───────────────────────────────────────────────────────────────

function StatusDot({ score }: { score: number }) {
  const color =
    score >= 75 ? 'bg-green-400' :
    score >= 50 ? 'bg-amber-400' :
    'bg-red-400';
  return (
    <span className="relative flex h-2.5 w-2.5">
      <span className={`animate-ping absolute inline-flex h-full w-full rounded-full opacity-75 ${color}`} />
      <span className={`relative inline-flex rounded-full h-2.5 w-2.5 ${color}`} />
    </span>
  );
}

// ── Branch Card ──────────────────────────────────────────────────────────────

interface BranchCardProps {
  locationId: string;
  name: string;
  city: string;
  seed: number;
  onClick: () => void;
}

function BranchCard({ locationId, name, city, seed, onClick }: BranchCardProps) {
  const { data: pnl, isLoading: pnlLoading } = usePnL(locationId);
  const { data: health, isLoading: healthLoading } = useHealth(locationId);
  const { data: alertCount } = useAlertCount(locationId);
  const { data: alerts } = useAlertQueue(locationId);
  const { data: labor } = useLaborSummary(locationId);

  const healthScore = Math.round(health?.overall_score ?? 0);
  const revenue = pnl?.net_revenue ?? 0;
  const orders = pnl?.check_count ?? 0;
  const margin = pnl?.gross_margin ?? 0;
  const totalAlerts = alertCount?.count ?? 0;
  const criticalAlerts = (alerts ?? []).filter((a: any) => a.severity === 'critical').length;
  const warningAlerts = (alerts ?? []).filter((a: any) => a.severity === 'warning').length;
  const staffOnShift = (labor as any)?.total_shifts ?? (labor as any)?.employee_count ?? 0;

  const sparkData = generateSparkData(seed);

  const cardGradient =
    healthScore >= 75
      ? 'from-slate-800/95 via-slate-800/90 to-green-950/80'
      : healthScore >= 50
      ? 'from-slate-800/95 via-slate-800/90 to-amber-950/80'
      : 'from-slate-800/95 via-slate-800/90 to-red-950/80';

  const borderColor =
    healthScore >= 75
      ? 'border-green-500/20 hover:border-green-400/40'
      : healthScore >= 50
      ? 'border-amber-500/20 hover:border-amber-400/40'
      : 'border-red-500/20 hover:border-red-400/40';

  const sparkColor =
    healthScore >= 75 ? '#22c55e' :
    healthScore >= 50 ? '#f59e0b' :
    '#ef4444';

  const isLoading = pnlLoading || healthLoading;

  const animatedRevenue = useCountUp(Math.round(revenue / 100));
  const animatedOrders = useCountUp(orders);

  // Revenue progress bar
  const TARGET_PIASTERS = 10_000_000; // 100K EGP
  const revenuePct = Math.min((revenue / TARGET_PIASTERS) * 100, 100);
  const progressColor =
    revenuePct >= 80 ? 'bg-green-500' :
    revenuePct >= 50 ? 'bg-amber-500' :
    'bg-red-500';

  return (
    <div
      onClick={onClick}
      className={`
        relative group cursor-pointer rounded-2xl border backdrop-blur-sm
        bg-gradient-to-br ${cardGradient} ${borderColor}
        transition-all duration-300 ease-out
        hover:scale-[1.025] hover:shadow-2xl hover:shadow-black/40
        overflow-hidden
      `}
    >
      {/* Subtle grid overlay */}
      <div
        className="absolute inset-0 opacity-5"
        style={{
          backgroundImage: 'linear-gradient(rgba(255,255,255,0.1) 1px, transparent 1px), linear-gradient(90deg, rgba(255,255,255,0.1) 1px, transparent 1px)',
          backgroundSize: '20px 20px',
        }}
      />

      {/* Glassmorphism shine */}
      <div className="absolute inset-0 rounded-2xl bg-gradient-to-br from-white/5 to-transparent pointer-events-none" />

      <div className="relative p-5 flex flex-col gap-4">
        {/* Header */}
        <div className="flex items-start justify-between">
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2 mb-1">
              <StatusDot score={healthScore} />
              <span className="text-xs font-medium text-slate-400 uppercase tracking-wider">
                {healthScore >= 75 ? 'Healthy' : healthScore >= 50 ? 'Fair' : 'Critical'}
              </span>
            </div>
            <h3 className="text-lg font-bold text-white truncate leading-tight">{name}</h3>
            <div className="flex items-center gap-1 mt-0.5">
              <MapPin className="h-3 w-3 text-slate-400" />
              <span className="text-xs text-slate-400">{city}</span>
            </div>
          </div>
          <HealthCircle score={isLoading ? 0 : healthScore} size={60} />
        </div>

        {/* KPI Row */}
        <div className="grid grid-cols-3 gap-2">
          <div className="relative rounded-xl p-2.5 text-center overflow-hidden" style={{ background: 'linear-gradient(135deg, rgba(16,185,129,0.15) 0%, rgba(5,150,105,0.08) 100%)' }}>
            <div className="absolute inset-0 border border-emerald-500/30 rounded-xl" />
            <div className="absolute -top-6 -right-6 w-16 h-16 bg-emerald-400/10 rounded-full blur-xl" />
            <p className="text-[10px] text-emerald-400/80 uppercase tracking-wider mb-1 relative">Revenue</p>
            {isLoading ? (
              <div className="h-5 bg-white/10 rounded animate-pulse mx-1" />
            ) : (
              <p className="text-base font-extrabold text-emerald-400 leading-none relative drop-shadow-[0_0_8px_rgba(16,185,129,0.4)]">
                EGP {animatedRevenue.toLocaleString()}
              </p>
            )}
          </div>
          <div className="bg-white/8 rounded-xl p-2.5 text-center border border-white/5">
            <p className="text-[10px] text-slate-400 uppercase tracking-wider mb-1">Orders</p>
            {isLoading ? (
              <div className="h-4 bg-white/10 rounded animate-pulse mx-1" />
            ) : (
              <p className="text-sm font-bold text-white leading-none">{animatedOrders.toLocaleString()}</p>
            )}
          </div>
          <div className="bg-white/8 rounded-xl p-2.5 text-center border border-white/5">
            <p className="text-[10px] text-slate-400 uppercase tracking-wider mb-1">Margin</p>
            {isLoading ? (
              <div className="h-4 bg-white/10 rounded animate-pulse mx-1" />
            ) : (
              <p className="text-sm font-bold text-white leading-none">{fmtPct(margin)}</p>
            )}
          </div>
        </div>

        {/* Feature 1: Revenue vs Target Progress Bar */}
        <div className="mt-1">
          <div className="flex justify-between text-[10px] text-slate-400 mb-1">
            <span>Daily Target</span>
            <span>{Math.round(revenue / 100000)}K / 100K EGP</span>
          </div>
          <div className="h-1.5 bg-white/10 rounded-full overflow-hidden">
            <div
              className={`h-full rounded-full transition-all duration-1000 ${progressColor}`}
              style={{ width: `${revenuePct}%` }}
            />
          </div>
        </div>

        {/* Sparkline */}
        <div className="h-14 -mx-1">
          <ResponsiveContainer width="100%" height="100%">
            <AreaChart data={sparkData} margin={{ top: 2, right: 2, left: 2, bottom: 0 }}>
              <defs>
                <linearGradient id={`sparkGrad-${locationId}`} x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor={sparkColor} stopOpacity={0.3} />
                  <stop offset="95%" stopColor={sparkColor} stopOpacity={0} />
                </linearGradient>
              </defs>
              <Area
                type="monotone"
                dataKey="value"
                stroke={sparkColor}
                strokeWidth={1.5}
                fill={`url(#sparkGrad-${locationId})`}
                dot={false}
                isAnimationActive={true}
              />
            </AreaChart>
          </ResponsiveContainer>
        </div>

        {/* Footer: alerts + staff */}
        <div className="flex items-center justify-between pt-1 border-t border-white/5">
          <div className="flex items-center gap-2">
            {criticalAlerts > 0 && (
              <span className="flex items-center gap-1 px-2 py-0.5 rounded-full bg-red-500/20 border border-red-500/30 text-red-400 text-xs font-semibold">
                <AlertCircle className="h-3 w-3" />
                {criticalAlerts}
              </span>
            )}
            {warningAlerts > 0 && (
              <span className="flex items-center gap-1 px-2 py-0.5 rounded-full bg-amber-500/20 border border-amber-500/30 text-amber-400 text-xs font-semibold">
                <AlertTriangle className="h-3 w-3" />
                {warningAlerts}
              </span>
            )}
            {totalAlerts === 0 && (
              <span className="text-xs text-slate-300">No alerts</span>
            )}
          </div>
          {staffOnShift > 0 && (
            <div className="flex items-center gap-1.5 text-slate-400">
              <Users className="h-3.5 w-3.5" />
              <span className="text-xs">{staffOnShift} on shift</span>
            </div>
          )}
        </div>

        {/* Feature 5: Staff On Floor Summary */}
        <div className="flex items-center gap-1.5 text-xs text-slate-400 -mt-2">
          <Users className="h-3 w-3" />
          <span>{staffOnShift > 0 ? `${staffOnShift} staff on floor` : 'Staff data loading...'}</span>
        </div>

        {/* Click CTA overlay on hover */}
        <div className="absolute inset-0 rounded-2xl flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity duration-200 bg-black/20">
          <div className="flex items-center gap-2 px-4 py-2 bg-[#F97316] rounded-full text-white text-sm font-semibold shadow-lg">
            <TrendingUp className="h-4 w-4" />
            Open Dashboard
          </div>
        </div>
      </div>
    </div>
  );
}

// ── Chain KPI Bar ─────────────────────────────────────────────────────────────

interface ChainKPI {
  totalRevenue: number;
  totalChecks: number;
  avgHealth: number;
  totalAlerts: number;
}

function ChainKPIBar({ kpi }: { kpi: ChainKPI }) {
  const animRevenue = useCountUp(Math.round(kpi.totalRevenue / 100), 1500);
  const animChecks = useCountUp(kpi.totalChecks, 1500);
  const animHealth = useCountUp(kpi.avgHealth, 1500);
  const animAlerts = useCountUp(kpi.totalAlerts, 800);

  const items = [
    {
      label: 'Total Revenue Today',
      value: `EGP ${animRevenue.toLocaleString()}`,
      icon: <span className="text-lg">💰</span>,
      color: 'text-emerald-400',
    },
    {
      label: 'Total Checks',
      value: animChecks.toLocaleString(),
      icon: <span className="text-lg">🧾</span>,
      color: 'text-blue-400',
    },
    {
      label: 'Avg Health Score',
      value: `${animHealth}`,
      icon: <Activity className="h-5 w-5" />,
      color: animHealth >= 75 ? 'text-green-400' : animHealth >= 50 ? 'text-amber-400' : 'text-red-400',
    },
    {
      label: 'Active Alerts',
      value: animAlerts.toString(),
      icon: <AlertTriangle className="h-5 w-5" />,
      color: animAlerts > 0 ? 'text-red-400' : 'text-slate-400',
    },
  ];

  return (
    <div className="grid grid-cols-2 lg:grid-cols-4 gap-3 mx-auto max-w-5xl w-full">
      {items.map((item) => (
        <div
          key={item.label}
          className="bg-white/5 backdrop-blur-sm border border-white/10 rounded-2xl px-5 py-4 text-center"
        >
          <div className={`flex items-center justify-center gap-2 mb-1 ${item.color}`}>
            {item.icon}
          </div>
          <p className={`text-2xl font-bold ${item.color}`}>{item.value}</p>
          <p className="text-xs text-slate-300 mt-1">{item.label}</p>
        </div>
      ))}
    </div>
  );
}

// ── Health Pulse ──────────────────────────────────────────────────────────────

function HealthPulse({ score }: { score: number }) {
  const color =
    score >= 75 ? 'bg-green-400' :
    score >= 50 ? 'bg-amber-400' :
    'bg-red-400';
  const label =
    score >= 75 ? 'All Systems Healthy' :
    score >= 50 ? 'Some Issues Detected' :
    'Critical Attention Needed';

  return (
    <div className="flex items-center justify-center gap-3">
      <span className="relative flex h-4 w-4">
        <span className={`animate-ping absolute inline-flex h-full w-full rounded-full opacity-60 ${color}`} />
        <span className={`relative inline-flex rounded-full h-4 w-4 ${color}`} />
      </span>
      <span className="text-sm font-medium text-slate-300">{label}</span>
      <span className="text-sm text-slate-300">— Chain Health {score}/100</span>
    </div>
  );
}

// ── Per-location data aggregator ──────────────────────────────────────────────

function useBranchData(locationId: string) {
  const pnl = usePnL(locationId);
  const health = useHealth(locationId);
  const alertCount = useAlertCount(locationId);
  return { pnl, health, alertCount };
}

// ── Chain KPI Comparison Table ────────────────────────────────────────────────

interface BranchKPIRow {
  name: string;
  shortName: string;
  revenue: number;    // piasters
  margin: number;     // percent
  health: number;
  alerts: number;
}

function ChainComparisonTable({ branches }: { branches: BranchKPIRow[] }) {
  const validBranches = branches.filter((b) => b.revenue > 0 || b.health > 0);
  const chainAvgRevenue = validBranches.length
    ? validBranches.reduce((s, b) => s + b.revenue, 0) / validBranches.length
    : 0;
  const chainAvgMargin = validBranches.length
    ? validBranches.reduce((s, b) => s + b.margin, 0) / validBranches.length
    : 0;
  const chainAvgHealth = validBranches.length
    ? validBranches.reduce((s, b) => s + b.health, 0) / validBranches.length
    : 0;
  const chainAvgAlerts = validBranches.length
    ? validBranches.reduce((s, b) => s + b.alerts, 0) / validBranches.length
    : 0;

  const cols = [
    ...branches,
    {
      name: 'Chain Avg',
      shortName: 'Avg',
      revenue: chainAvgRevenue,
      margin: chainAvgMargin,
      health: chainAvgHealth,
      alerts: chainAvgAlerts,
    },
  ];

  // Data for bar chart
  const branchChartData = branches.map((b) => ({
    name: b.shortName,
    revenue: Math.round(b.revenue / 100),
    health: Math.round(b.health),
  }));

  return (
    <div className="space-y-6">
      <div className="overflow-x-auto">
        <div className="min-w-0">
          <div className="flex items-center gap-3 mb-4">
            <div className="h-px flex-1 bg-white/5" />
            <span className="text-xs font-semibold uppercase tracking-widest text-slate-300">
              Chain KPI Comparison
            </span>
            <div className="h-px flex-1 bg-white/5" />
          </div>
          <div className="rounded-xl border border-white/8 overflow-hidden">
            {/* Header row */}
            <div
              className="grid text-center"
              style={{ gridTemplateColumns: `160px repeat(${cols.length}, 1fr)` }}
            >
              <div className="bg-white/5 px-4 py-3 text-left">
                <span className="text-xs font-semibold text-slate-300 uppercase tracking-wider">Metric</span>
              </div>
              {cols.map((b, i) => (
                <div
                  key={b.name}
                  className={`px-3 py-3 ${i === cols.length - 1 ? 'bg-white/8 border-l border-white/8' : 'bg-white/5'}`}
                >
                  <span className={`text-xs font-bold uppercase tracking-wider ${i === cols.length - 1 ? 'text-slate-300' : 'text-slate-400'}`}>
                    {b.shortName}
                  </span>
                </div>
              ))}
            </div>

            {/* Revenue row */}
            <div
              className="grid text-center border-t border-white/5"
              style={{ gridTemplateColumns: `160px repeat(${cols.length}, 1fr)` }}
            >
              <div className="bg-white/3 px-4 py-3 text-left flex items-center gap-2">
                <span className="text-xs text-slate-400 font-medium">Revenue (EGP K)</span>
              </div>
              {cols.map((b, i) => (
                <div
                  key={b.name}
                  className={`px-3 py-3 ${i === cols.length - 1 ? 'bg-white/5 border-l border-white/8' : 'bg-white/3'}`}
                >
                  <span className={`text-sm font-bold ${i === cols.length - 1 ? 'text-slate-300' : 'text-emerald-400'}`}>
                    {b.revenue > 0 ? Math.round(b.revenue / 100000).toLocaleString() + 'K' : '—'}
                  </span>
                </div>
              ))}
            </div>

            {/* Margin row */}
            <div
              className="grid text-center border-t border-white/5"
              style={{ gridTemplateColumns: `160px repeat(${cols.length}, 1fr)` }}
            >
              <div className="bg-white/3 px-4 py-3 text-left">
                <span className="text-xs text-slate-400 font-medium">Gross Margin</span>
              </div>
              {cols.map((b, i) => (
                <div
                  key={b.name}
                  className={`px-3 py-3 ${i === cols.length - 1 ? 'bg-white/5 border-l border-white/8' : 'bg-white/3'}`}
                >
                  <span className={`text-sm font-bold ${i === cols.length - 1 ? 'text-slate-300' : 'text-blue-400'}`}>
                    {b.margin > 0 ? b.margin.toFixed(1) + '%' : '—'}
                  </span>
                </div>
              ))}
            </div>

            {/* Health row */}
            <div
              className="grid text-center border-t border-white/5"
              style={{ gridTemplateColumns: `160px repeat(${cols.length}, 1fr)` }}
            >
              <div className="bg-white/3 px-4 py-3 text-left">
                <span className="text-xs text-slate-400 font-medium">Health Score</span>
              </div>
              {cols.map((b, i) => {
                const score = Math.round(b.health);
                const color = score >= 75 ? 'text-green-400' : score >= 50 ? 'text-amber-400' : 'text-red-400';
                return (
                  <div
                    key={b.name}
                    className={`px-3 py-3 ${i === cols.length - 1 ? 'bg-white/5 border-l border-white/8' : 'bg-white/3'}`}
                  >
                    <span className={`text-sm font-bold ${i === cols.length - 1 ? 'text-slate-300' : color}`}>
                      {score > 0 ? score : '—'}
                    </span>
                  </div>
                );
              })}
            </div>

            {/* Alerts row */}
            <div
              className="grid text-center border-t border-white/5"
              style={{ gridTemplateColumns: `160px repeat(${cols.length}, 1fr)` }}
            >
              <div className="bg-white/3 px-4 py-3 text-left">
                <span className="text-xs text-slate-400 font-medium">Active Alerts</span>
              </div>
              {cols.map((b, i) => {
                const count = Math.round(b.alerts);
                return (
                  <div
                    key={b.name}
                    className={`px-3 py-3 ${i === cols.length - 1 ? 'bg-white/5 border-l border-white/8' : 'bg-white/3'}`}
                  >
                    <span className={`text-sm font-bold ${count > 0 ? (i === cols.length - 1 ? 'text-slate-300' : 'text-red-400') : 'text-slate-300'}`}>
                      {count}
                    </span>
                  </div>
                );
              })}
            </div>
          </div>
        </div>
      </div>

      {/* Feature 3: Chain-Wide Comparison Bar Chart */}
      <div className="bg-white/3 border border-white/8 rounded-xl p-5">
        <p className="text-xs font-semibold uppercase tracking-widest text-slate-300 mb-4 text-center">
          Branch Performance Chart
        </p>
        <ResponsiveContainer width="100%" height={200}>
          <BarChart data={branchChartData} margin={{ top: 10, right: 20, left: 20, bottom: 5 }}>
            <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.05)" />
            <XAxis dataKey="name" tick={{ fill: '#94a3b8', fontSize: 11 }} />
            <YAxis tick={{ fill: '#94a3b8', fontSize: 11 }} />
            <Tooltip
              contentStyle={{
                background: '#1e293b',
                border: '1px solid rgba(255,255,255,0.1)',
                borderRadius: '8px',
                color: '#f1f5f9',
              }}
            />
            <Bar dataKey="revenue" name="Revenue (EGP)" fill="#22c55e" radius={[4, 4, 0, 0]} />
            <Bar dataKey="health" name="Health Score" fill="#3b82f6" radius={[4, 4, 0, 0]} />
          </BarChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
}

// ── Feature 2: AI Insights Ticker ────────────────────────────────────────────

function AIInsightsTicker() {
  return (
    <div className="overflow-hidden bg-slate-900/50 border-y border-white/5 py-3">
      <div className="animate-scroll-left flex gap-12 whitespace-nowrap">
        {[...AI_INSIGHTS, ...AI_INSIGHTS].map((insight, i) => (
          <span key={i} className="text-sm text-slate-300">{insight}</span>
        ))}
      </div>
    </div>
  );
}

// ── Feature 4: Today's Revenue Race ──────────────────────────────────────────

function RevenueRace({ branchRows }: { branchRows: BranchKPIRow[] }) {
  const sorted = [...branchRows].sort((a, b) => b.revenue - a.revenue);
  const maxRevenue = Math.max(...sorted.map((r) => r.revenue), 1);

  return (
    <div className="bg-white/3 border border-white/8 rounded-xl p-5 space-y-4">
      <h3 className="text-sm font-bold text-white uppercase tracking-wider">Today's Revenue Race</h3>
      <div className="space-y-3">
        {sorted.map((b, idx) => {
          const pct = (b.revenue / maxRevenue) * 100;
          return (
            <div key={b.name} className="flex items-center gap-3">
              <span className="text-xs text-slate-400 w-16 sm:w-28 md:w-36 truncate">{b.shortName}</span>
              <div className="flex-1 h-6 bg-white/5 rounded-full overflow-hidden">
                <div
                  className={`h-full rounded-full transition-all duration-1000 ${
                    idx === 0
                      ? 'bg-gradient-to-r from-emerald-500 to-emerald-400'
                      : 'bg-gradient-to-r from-slate-600 to-slate-500'
                  }`}
                  style={{ width: `${pct}%` }}
                />
              </div>
              <span className="text-xs font-bold text-white w-16 text-right">
                {Math.round(b.revenue / 100000)}K
              </span>
              {idx === 0 && <span className="text-xs">🏆</span>}
            </div>
          );
        })}
      </div>
    </div>
  );
}

// ── CEO Executive Briefing ────────────────────────────────────────────────────

const TEAM_MEMBERS = [
  'Ahmed Hassan (GM - El Gouna)',
  'Fatma Ali (GM - New Cairo)',
  'Omar Sayed (GM - Zayed)',
  'Layla Ibrahim (GM - North Coast)',
  'Sara Mostafa (Operations Director)',
  'Khaled Nabil (Finance Director)',
  'Dina Rashwan (Legal & Compliance)',
  'Hany Farid (Supply Chain Manager)',
  'Rania Helmy (HR Director)',
  'Marketing Department',
  'Finance Department',
  'Legal Department',
  'You (CEO)',
];

interface ItemState {
  expanded: boolean;
  status: 'open' | 'assigned' | 'in_progress' | 'resolved';
  assignedTo: string | null;
  comments: { text: string; time: string; author: string }[];
  showAssign: boolean;
  showComment: boolean;
  commentDraft: string;
  priority: 'critical' | 'high' | 'medium' | 'low';
  activityLog: { text: string; time: string }[];
}

function defaultItemState(): ItemState {
  return {
    expanded: false,
    status: 'open',
    assignedTo: null,
    comments: [],
    showAssign: false,
    showComment: false,
    commentDraft: '',
    priority: 'high',
    activityLog: [],
  };
}

function StatusBadge({ status }: { status: ItemState['status'] }) {
  if (status === 'open') return null;
  const styles: Record<string, string> = {
    assigned: 'bg-blue-500/20 text-blue-400',
    in_progress: 'bg-amber-500/20 text-amber-400',
    resolved: 'bg-green-500/20 text-green-400',
  };
  const labels: Record<string, string> = {
    assigned: 'Assigned',
    in_progress: 'In Progress',
    resolved: 'Resolved',
  };
  return (
    <span className={`text-[10px] font-semibold uppercase tracking-wider px-2 py-0.5 rounded-full ${styles[status]}`}>
      {labels[status]}
    </span>
  );
}

interface BriefingItem {
  icon: string;
  title: string;
  impact: string;
  branch: string;
  severity?: 'critical' | 'warning' | 'positive' | 'info';
  metric?: string;
  timeframe?: string;
  rootCause?: string;
  factors?: string[];
  impactChain?: string;
  recommendations?: { immediate: string[]; longTerm: string[] };
  aiInsight?: string;
}

const ATTENTION_ITEMS: BriefingItem[] = [
  {
    icon: '🔴',
    title: 'Food Cost Breach — Nimbu New Cairo',
    impact: 'Trending 34.2% vs 32% target. Protein costs drove +EGP 18,000 excess this month.',
    branch: 'Nimbu New Cairo · Action: Review Sea Bass sourcing',
    severity: 'critical',
    metric: '34.2% vs 32% target',
    timeframe: 'Detected 45 minutes ago · Trending for 8 days',
    rootCause: 'Sea Bass wholesale price increased 18% ($38→$48/kg) from Sysco Egypt over the past 6 weeks. Combined with Jumbo Shrimp up 15%, protein costs now account for 68% of the food cost overage.',
    factors: [
      'Sea Bass: +18% cost, used in Ceviche Clásico (285 EGP) and Tiradito Nikkei (320 EGP)',
      'Jumbo Shrimp: +15% cost, used in Arroz con Mariscos (480 EGP)',
      'No menu price adjustment since original launch',
      'Portion sizes not recalibrated after cost increase',
    ],
    impactChain: 'Ingredient cost ↑ → COGS ↑ → Food cost % ↑ → Gross margin ↓ → Profitability risk',
    recommendations: {
      immediate: [
        'Adjust Ceviche Clásico price from 285 → 310 EGP (+8.8%) — projected <3% volume impact',
        'Negotiate with Seoudi Fresh for backup Sea Bass supply at current market rate',
        'Reduce Sea Bass portion in Tiradito from 180g to 160g (invisible to guest)',
      ],
      longTerm: [
        'Set up automated price alert when any protein exceeds +10% from baseline',
        'Enable dynamic menu pricing tied to ingredient cost thresholds',
        'Diversify protein suppliers — single-source dependency on Sysco for Sea Bass is a risk',
      ],
    },
  },
  {
    icon: '🔴',
    title: 'VIP Customer Churn Risk',
    impact: '5 high-CLV guests (avg EGP 2,400/mo) haven\'t visited in 21+ days. Projected loss: EGP 12,000/mo.',
    branch: 'Nimbu El Gouna & Nimbu Zayed',
    severity: 'critical',
    metric: '5 VIPs · EGP 12,000/mo at risk',
    timeframe: 'Detected 2 hours ago · Pattern emerging over 3 weeks',
    rootCause: '5 champion-segment guests with combined monthly spend of EGP 12,000 show visit frequency decay pattern. Average inter-visit interval expanded from 7 days to 24+ days.',
    factors: [
      'Amr Abdel-Rahman — 12 visits, avg EGP 2,400/mo, last visit 23 days ago',
      'Nadia Farouk — 8 visits, avg EGP 1,800/mo, last visit 28 days ago',
      '3 additional guests with similar decay patterns',
      'No specific incident detected — seasonal pattern possible (post-winter slowdown)',
    ],
    impactChain: 'Visit frequency ↓ → Revenue per VIP ↓ → CLV erosion → Potential permanent churn',
    recommendations: {
      immediate: [
        'Send personalized win-back offer: complimentary Pisco Sour on next visit',
        'Have GM personally reach out to Amr Abdel-Rahman and Nadia Farouk',
        'Trigger loyalty points double-earn event this weekend',
      ],
      longTerm: [
        'Build automated VIP churn detection alert at 14-day no-visit threshold',
        'Create a VIP concierge protocol — dedicated table, welcome note on arrival',
        'Analyze historical triggers: what brought these guests back last time?',
      ],
    },
  },
  {
    icon: '🟡',
    title: 'Nimbu North Coast Ticket Time Degradation',
    impact: 'Average ticket time increased from 12→18 min this week. Ceviche Bar is the bottleneck.',
    branch: 'Nimbu North Coast · Guest complaints likely to follow',
    severity: 'warning',
    metric: '18 min avg · Target: 12 min',
    timeframe: 'Detected 6 hours ago · Began March 14',
    rootCause: 'New cook (Youssef) assigned to Ceviche Bar station with ELU rating 0.6 (team avg: 1.2). 73% of delayed tickets include ceviche items. Average ceviche prep time: 8.2 min (target: 4.5 min).',
    factors: [
      'Youssef — ELU rating 0.6 vs team avg 1.2 at Ceviche Bar station',
      '73% of delayed tickets include at least one ceviche item',
      'Average ceviche prep time 8.2 min vs 4.5 min target',
      'Issue began March 14 when previous ceviche cook transferred to Nimbu Zayed',
    ],
    impactChain: 'Station underperformance → Ticket time ↑ → Guest wait ↑ → Satisfaction ↓ → Review risk',
    recommendations: {
      immediate: [
        'Pair Youssef with senior cook for 3-day intensive ceviche station training',
        'Temporarily reduce ceviche menu to 2 items during peak hours to ease station load',
        'Move fastest available prep cook to ceviche support role until resolved',
      ],
      longTerm: [
        'Build station ELU minimum thresholds before assigning new cooks solo',
        'Create ceviche prep SOP video for faster onboarding',
        'Consider cross-training at least 2 staff on ceviche station at each branch',
      ],
    },
  },
  {
    icon: '🟡',
    title: 'Labor Overtime at 3 Branches',
    impact: '7 staff exceeded 40hr/week. Projected overtime cost: EGP 8,500.',
    branch: 'Nimbu New Cairo, Zayed, Nimbu North Coast',
    severity: 'warning',
    metric: '7 staff · EGP 8,500 overtime projected',
    timeframe: 'Detected this morning · Week 3 of 4',
    rootCause: 'Uneven shift scheduling combined with higher-than-forecast Friday covers caused overtime accumulation across 3 branches. Nimbu New Cairo accounts for 58% of the excess hours.',
    factors: [
      'Friday dinner covers were 22% above forecast at Nimbu New Cairo',
      '3 staff at Nimbu Zayed picked up extra shifts to cover illness call-outs',
      'Scheduling system did not flag approaching 40hr limit before shifts were approved',
      'Nimbu North Coast: 1 cook at 47hrs due to Ceviche Bar coverage needs',
    ],
    impactChain: 'Over-scheduling → Overtime threshold breach → Payroll cost spike → Budget overrun',
    recommendations: {
      immediate: [
        'Cap all remaining shifts this week — no staff to exceed current hours',
        'Pull 2 part-time staff from the on-call roster to cover peak weekend hours',
        'Flag Nimbu New Cairo schedule to Sara Mostafa for immediate review',
      ],
      longTerm: [
        'Enable 38hr weekly alert in scheduling software — automated warning before overtime threshold',
        'Build a part-time flex pool — minimum 3 on-call staff per branch',
        'Review Friday forecast model — currently under-predicting by 20%+',
      ],
    },
  },
  {
    icon: '🟡',
    title: 'Vendor Reliability Drop — Metro Market',
    impact: 'OTIF rate fell to 72%. 3 short deliveries this month affecting produce quality.',
    branch: 'Nimbu North Coast',
    severity: 'warning',
    metric: '72% OTIF · Target: 95%',
    timeframe: 'Detected 1 day ago · 3 incidents this month',
    rootCause: 'Metro Market logistics capacity appears strained — 3 of 8 deliveries this month were either late (>2hr) or short-shipped on produce. Avocado and Limes most frequently impacted.',
    factors: [
      'Delivery #4 (March 8): Avocados short 40%, forced menu substitution',
      'Delivery #6 (March 13): 3hr late — cold chain compliance risk',
      'Delivery #8 (March 19): Lime short 60%, purchased emergency stock at retail',
      'Metro Market\'s Borg El Arab hub reportedly under new management since February',
    ],
    impactChain: 'Vendor unreliability → Stock shortages → Menu gaps → Guest experience impact → Emergency spend',
    recommendations: {
      immediate: [
        'Place dual orders for Avocado and Limes — split 50/50 between Metro Market and Seoudi Fresh',
        'Issue formal performance notice to Metro Market account manager',
        'Increase safety stock for top 5 produce items at North Coast by 30%',
      ],
      longTerm: [
        'Qualify Seoudi Fresh as primary produce vendor for North Coast branch',
        'Set OTIF floor at 85% — automatic vendor review if breached over 2 consecutive weeks',
        'Build a dual-vendor policy for all A-category ingredients',
      ],
    },
  },
];

const HIGHLIGHT_ITEMS: BriefingItem[] = [
  {
    icon: '✅',
    title: 'Chain Revenue On Track',
    impact: 'EGP 387K today across 4 branches (target: 400K). Trending +5% vs last week.',
    branch: 'All Branches',
    severity: 'positive',
    metric: 'EGP 387K / 400K target',
    timeframe: 'As of 11:00 PM today',
    aiInsight: 'Revenue trajectory is on pace to hit 400K by end of service. Nimbu El Gouna dinner surge (+18% vs Monday avg) is the primary driver. If current trends hold, chain will exceed monthly target by ~EGP 45,000. No action required — monitor through close.',
    recommendations: {
      immediate: [
        'Ensure all branches are fully staffed for late dinner push',
        'Upsell premium cocktails during last seating to boost per-cover average',
      ],
      longTerm: [
        'Analyze which weekday drivers contributed to the +5% week-over-week lift',
        'Consider raising monthly target to EGP 420K for Q2 planning',
      ],
    },
  },
  {
    icon: '✅',
    title: 'Nimbu Zayed Best Performer',
    impact: '12% below chain average on food cost. Recommend propagating their portioning practices.',
    branch: 'Nimbu Zayed',
    severity: 'positive',
    metric: 'Food cost 28.1% vs 32% chain avg',
    timeframe: 'Consistent for 6 weeks running',
    aiInsight: 'Nimbu Zayed\'s kitchen team, led by Head Chef Mona, has maintained a 28.1% food cost for 6 consecutive weeks — nearly 4 points below chain average. Key differentiators: daily waste log reviews, prep-to-order ratio discipline, and consistent portion calibration using digital scales. Their practices should be documented and rolled out chain-wide.',
    recommendations: {
      immediate: [
        'Schedule a knowledge-sharing session: Chef Mona presents portioning practices to all GMs',
        'Film a short prep video at Nimbu Zayed for training library',
      ],
      longTerm: [
        'Roll out digital portion scale requirement to all branches by end of month',
        'Create a "Best Practice Bounty" incentive — branches that maintain <30% food cost earn a bonus',
      ],
    },
  },
  {
    icon: '✅',
    title: 'Pisco Hour Campaign Success',
    impact: '45 redemptions, EGP 2,250 attributed revenue. Consider expanding to all branches.',
    branch: 'Nimbu El Gouna',
    severity: 'positive',
    metric: '45 redemptions · EGP 2,250 revenue',
    timeframe: 'Campaign ran March 14–21',
    aiInsight: 'The Pisco Hour happy-hour promotion at Nimbu El Gouna exceeded projections by 28%. Average redemption spend was EGP 50 on cocktails, with 62% of guests ordering a food item alongside — creating a halo effect. Profit margin on Pisco Sour is 78%, making this a high-return campaign. Recommend expanding to all branches.',
    recommendations: {
      immediate: [
        'Launch Pisco Hour at Nimbu Zayed and Nimbu New Cairo starting this Friday',
        'Create simple social media assets — short reel featuring the cocktail',
      ],
      longTerm: [
        'Build a rotating happy-hour calendar across all branches to drive weekday traffic',
        'Analyze which food items were most ordered alongside — feature them as a pairing menu',
      ],
    },
  },
  {
    icon: '✅',
    title: 'New Menu Classification',
    impact: 'Empanadas reclassified to \'crowd_pleaser\' after 22% velocity increase post portion adjustment.',
    branch: 'All Branches',
    severity: 'positive',
    metric: '+22% velocity · Reclassified to crowd_pleaser',
    timeframe: 'Reclassified March 18',
    aiInsight: 'Following the portion adjustment that increased Empanadas from 3 to 4 pieces per serving (cost increase: EGP 4.50), velocity jumped 22% and the item moved from "question mark" to "crowd_pleaser" classification. The net effect is positive — higher volume more than compensates for the portion cost increase. This is a textbook example of a price-to-volume optimization working as intended.',
    recommendations: {
      immediate: [
        'Feature Empanadas in the "Staff Recommendation" section of menus this week',
        'Brief all servers on the item reclassification — encourage upsell as a starter',
      ],
      longTerm: [
        'Apply the same portion-sensitivity analysis to 3 other "question mark" items',
        'Consider creating an Empanadas trio sampler for the bar menu',
      ],
    },
  },
];

const OUTLOOK_ITEMS: BriefingItem[] = [
  {
    icon: '📊',
    title: 'Q2 Pisco Price Forecast',
    impact: 'Specialty Imports projects 15% price increase. Forward-purchasing 3-month supply saves ~EGP 25,000.',
    branch: 'Strategic · Recommend action by April 1',
    severity: 'info',
    metric: '+15% price forecast · EGP 25,000 savings opportunity',
    timeframe: 'Price increase effective April 15',
    aiInsight: 'Specialty Imports has communicated a 15% price increase on all Pisco SKUs effective April 15, citing Peruvian harvest conditions. Current consumption across 4 branches is approximately 180 bottles/month. Locking in a 3-month forward purchase (540 bottles) at current pricing saves approximately EGP 25,000 and ensures supply continuity through the Ramadan season when Pisco Sour demand typically remains strong among non-fasting guests.',
    recommendations: {
      immediate: [
        'Authorize forward purchase of 540 Pisco bottles at current price — deadline: April 1',
        'Confirm storage capacity at each branch for approximately 135 bottles each',
        'Request formal quote from Specialty Imports to lock in pricing',
      ],
      longTerm: [
        'Build a commodity forward-purchasing policy for top 5 beverage SKUs',
        'Add price forecast monitoring to the AI system — auto-alert when >10% increase projected',
        'Evaluate a second Pisco supplier to reduce single-source dependency',
      ],
    },
  },
  {
    icon: '📊',
    title: 'Ramadan Prep Required',
    impact: 'Ramadan starts April 2. Historical: +30% dinner volume, −40% lunch. Schedule adjustments needed.',
    branch: 'All Branches',
    severity: 'info',
    metric: '+30% dinner · −40% lunch · April 2 start',
    timeframe: 'Action required by March 28',
    aiInsight: 'Based on last year\'s Ramadan data across all Nimbu branches: dinner covers surged 30–35% while lunch service dropped 40%. Net revenue was +8% vs non-Ramadan weeks. However, labor cost efficiency dropped because schedules were not adjusted until week 2. This year, pre-adjusting schedules by March 28 will capture the full revenue lift without the cost drag.',
    recommendations: {
      immediate: [
        'Shift all branch schedules: reduce lunch staff 40%, add 2 evening staff per branch',
        'Build a Ramadan set menu (Iftar offer) to capitalize on group bookings',
        'Pre-book large tables for family Iftar — open reservations now on all platforms',
      ],
      longTerm: [
        'Create a Ramadan operations playbook to use annually — replicable every year',
        'Evaluate a dedicated Iftar delivery offering through Talabat/Elmenus',
        'Plan a Sohour (late-night) menu for El Gouna and North Coast branches',
      ],
    },
  },
  {
    icon: '📊',
    title: 'Nimbu North Coast Seasonal Ramp',
    impact: 'Summer season begins May 1. Last year\'s volume was 2.3× winter. Hiring pipeline should start now.',
    branch: 'Nimbu North Coast',
    severity: 'info',
    metric: '2.3× winter volume projected · May 1 onset',
    timeframe: 'Hiring pipeline must start by April 1',
    aiInsight: 'North Coast\'s summer peak (May–September) consistently delivers 2.3× winter volume based on 2024 data. To staff up appropriately, the hiring pipeline needs to begin now — recruitment, onboarding, and training for 8 additional seasonal staff takes approximately 4 weeks minimum. Missing the May 1 window risks understaffing during the highest-revenue period of the year.',
    recommendations: {
      immediate: [
        'Post job listings for 8 seasonal staff (4 FOH, 4 BOH) immediately',
        'Contact last year\'s seasonal staff — offer early sign-up bonus for returnees',
        'Book training block for weeks of April 14 and April 21',
      ],
      longTerm: [
        'Build a seasonal staff retention program — returning staff get priority rehire + higher starting rate',
        'Consider a permanent bump in North Coast kitchen capacity ahead of summer 2027',
        'Create a Summer Operations Manual for North Coast — covers staffing ratios, extended hours, beach delivery',
      ],
    },
  },
  {
    icon: '📊',
    title: 'Expansion Opportunity',
    impact: 'Zayed consistently exceeds capacity on weekends (92% kitchen utilization). Consider satellite prep kitchen.',
    branch: 'Nimbu Zayed',
    severity: 'info',
    metric: '92% kitchen utilization on weekends',
    timeframe: 'Trending for 6+ weeks',
    aiInsight: 'Nimbu Zayed\'s kitchen has been running at 92% utilization on Friday and Saturday for 6+ consecutive weeks. This is above the 85% threshold where ticket times begin to degrade. A satellite prep kitchen within 2km could handle cold prep and mise en place, freeing the main kitchen for service. Alternatively, a second Nimbu location in the Zayed corridor could capture demand that is currently being turned away.',
    recommendations: {
      immediate: [
        'Commission a feasibility study on a satellite prep kitchen — budget EGP 15,000 for the study',
        'Identify suitable commercial kitchen rental spaces in the Sheikh Zayed area',
        'Analyze turn-away data — how many covers are being lost on peak nights?',
      ],
      longTerm: [
        'Evaluate a Nimbu Zayed 2 location — full branch vs. delivery-only dark kitchen',
        'Model the ROI of a satellite prep kitchen vs. new branch vs. status quo',
        'Consider a partnership with a shared commercial kitchen facility as interim solution',
      ],
    },
  },
];

// ── Briefing Detail Modal ─────────────────────────────────────────────────────

interface BriefingDetailModalProps {
  item: BriefingItem;
  itemKey: string;
  isAttention: boolean;
  state: ItemState;
  onUpdate: (key: string, patch: Partial<ItemState>) => void;
  onClose: () => void;
}

function BriefingDetailModal({ item, itemKey, isAttention, state, onUpdate, onClose }: BriefingDetailModalProps) {
  const severityBadge = {
    critical: 'bg-red-500/20 text-red-400 border border-red-500/30',
    warning: 'bg-amber-500/20 text-amber-400 border border-amber-500/30',
    positive: 'bg-green-500/20 text-green-400 border border-green-500/30',
    info: 'bg-blue-500/20 text-blue-400 border border-blue-500/30',
  };
  const severityLabel = {
    critical: 'Critical',
    warning: 'Warning',
    positive: 'Positive',
    info: 'Outlook',
  };
  const sev = item.severity ?? 'info';

  // Close on backdrop click
  function handleBackdropClick(e: React.MouseEvent<HTMLDivElement>) {
    if (e.target === e.currentTarget) onClose();
  }

  // Escape key
  useEffect(() => {
    function onKey(e: KeyboardEvent) {
      if (e.key === 'Escape') onClose();
    }
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [onClose]);

  const priorityColors: Record<string, string> = {
    critical: 'bg-red-500/20 text-red-400 border-red-500/30',
    high: 'bg-amber-500/20 text-amber-400 border-amber-500/30',
    medium: 'bg-blue-500/20 text-blue-400 border-blue-500/30',
    low: 'bg-slate-500/20 text-slate-400 border-slate-500/30',
  };

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center p-4"
      style={{ background: 'rgba(0,0,0,0.65)', backdropFilter: 'blur(4px)' }}
      onClick={handleBackdropClick}
    >
      <div
        className="bg-slate-800 border border-white/15 rounded-2xl shadow-2xl max-w-2xl w-full max-h-[85vh] overflow-y-auto"
        style={{ scrollbarWidth: 'thin', scrollbarColor: 'rgba(255,255,255,0.1) transparent' }}
      >
        {/* Modal Header */}
        <div className="sticky top-0 bg-slate-800 border-b border-white/10 px-6 py-4 rounded-t-2xl z-10">
          <div className="flex items-start gap-3">
            <span className="text-2xl shrink-0 mt-0.5">{item.icon}</span>
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-2 flex-wrap mb-1">
                <span className={`text-[10px] font-bold uppercase tracking-wider px-2 py-0.5 rounded-full ${severityBadge[sev]}`}>
                  {severityLabel[sev]}
                </span>
                <span className="text-[10px] bg-white/10 text-slate-400 px-2 py-0.5 rounded-full font-medium">
                  {item.branch.split('·')[0].trim()}
                </span>
                <StatusBadge status={state.status} />
                {state.assignedTo && (
                  <span className="text-[10px] bg-blue-500/20 text-blue-400 px-2 py-0.5 rounded-full font-semibold">
                    {state.assignedTo.split(' (')[0]}
                  </span>
                )}
              </div>
              <h2 className="text-lg font-bold text-white leading-tight">{item.title}</h2>
            </div>
            <button
              onClick={onClose}
              className="shrink-0 w-8 h-8 flex items-center justify-center rounded-lg bg-white/5 hover:bg-white/15 text-slate-400 hover:text-white transition-colors text-lg leading-none"
            >
              ×
            </button>
          </div>
        </div>

        <div className="px-6 py-5 space-y-4">
          {/* Section 1: Impact Summary */}
          <div className="bg-white/5 rounded-xl border border-white/10 p-4 space-y-3">
            <p className="text-xs font-bold uppercase tracking-wider text-slate-400 flex items-center gap-2">
              <span>📊</span> Impact Summary
            </p>
            <p className="text-sm text-slate-300 leading-relaxed">{item.impact}</p>
            {item.metric && (
              <div className="bg-white/5 rounded-lg px-4 py-3 border border-white/8">
                <p className="text-[10px] text-slate-300 uppercase tracking-wider mb-1">Key Metric</p>
                <p className="text-xl font-black text-white">{item.metric}</p>
              </div>
            )}
            {item.timeframe && (
              <p className="text-xs text-slate-300 flex items-center gap-1.5">
                <span>🕐</span> {item.timeframe}
              </p>
            )}
          </div>

          {/* Section 2: AI Root Cause Analysis (for attention items) or AI Insights */}
          {(item.rootCause || item.aiInsight) && (
            <div className="bg-white/5 rounded-xl border border-white/10 p-4 space-y-3">
              <p className="text-xs font-bold uppercase tracking-wider text-slate-400 flex items-center gap-2">
                <span>🔍</span> {item.rootCause ? 'AI Root Cause Analysis' : 'AI Insight'}
              </p>
              <p className="text-sm text-slate-300 leading-relaxed">
                {item.rootCause ?? item.aiInsight}
              </p>
              {item.factors && item.factors.length > 0 && (
                <div className="space-y-1.5 mt-1">
                  <p className="text-[11px] font-semibold text-slate-400 uppercase tracking-wider">Contributing Factors</p>
                  <ul className="space-y-1.5">
                    {item.factors.map((f, i) => (
                      <li key={i} className="flex items-start gap-2 text-xs text-slate-300">
                        <span className="text-indigo-400 shrink-0 mt-0.5">•</span>
                        <span>{f}</span>
                      </li>
                    ))}
                  </ul>
                </div>
              )}
              {item.impactChain && (
                <div className="bg-red-950/30 border border-red-500/20 rounded-lg px-3 py-2.5 mt-1">
                  <p className="text-[10px] font-semibold text-red-400 uppercase tracking-wider mb-1">Impact Chain</p>
                  <p className="text-xs text-slate-300 font-mono">{item.impactChain}</p>
                </div>
              )}
            </div>
          )}

          {/* Section 3: Prevention & Recommendations */}
          {item.recommendations && (
            <div className="bg-white/5 rounded-xl border border-white/10 p-4 space-y-3">
              <p className="text-xs font-bold uppercase tracking-wider text-slate-400 flex items-center gap-2">
                <span>🛡️</span> Prevention & Recommendations
              </p>
              {item.recommendations.immediate.length > 0 && (
                <div className="space-y-2">
                  <p className="text-[11px] font-semibold text-amber-400 uppercase tracking-wider">Immediate Actions</p>
                  <ol className="space-y-1.5">
                    {item.recommendations.immediate.map((r, i) => (
                      <li key={i} className="flex items-start gap-2.5 text-xs text-slate-300">
                        <span className="shrink-0 w-4 h-4 rounded-full bg-amber-500/20 text-amber-400 text-[10px] font-bold flex items-center justify-center mt-0.5">
                          {i + 1}
                        </span>
                        <span>{r}</span>
                      </li>
                    ))}
                  </ol>
                </div>
              )}
              {item.recommendations.longTerm.length > 0 && (
                <div className="space-y-2 pt-1 border-t border-white/8">
                  <p className="text-[11px] font-semibold text-blue-400 uppercase tracking-wider">Long-Term Prevention</p>
                  <ul className="space-y-1.5">
                    {item.recommendations.longTerm.map((r, i) => (
                      <li key={i} className="flex items-start gap-2 text-xs text-slate-300">
                        <span className="text-blue-400 shrink-0 mt-0.5">•</span>
                        <span>{r}</span>
                      </li>
                    ))}
                  </ul>
                </div>
              )}
            </div>
          )}

          {/* Section 4: Actions */}
          <div className="bg-white/5 rounded-xl border border-white/10 p-4 space-y-3">
            <p className="text-xs font-bold uppercase tracking-wider text-slate-400 flex items-center gap-2">
              <span>⚡</span> Actions
            </p>

            {/* Priority selector */}
            <div className="flex items-center gap-2 flex-wrap">
              <span className="text-[11px] text-slate-300 mr-1">Priority:</span>
              {(['critical', 'high', 'medium', 'low'] as const).map((p) => (
                <button
                  key={p}
                  onClick={() => onUpdate(itemKey, { priority: p })}
                  className={`text-[10px] font-bold uppercase tracking-wider px-2.5 py-1 rounded-full border transition-all ${
                    state.priority === p
                      ? priorityColors[p] + ' ring-1 ring-offset-0'
                      : 'bg-white/5 text-slate-300 border-white/10 hover:bg-white/10'
                  }`}
                >
                  {p}
                </button>
              ))}
            </div>

            {/* Action buttons */}
            <div className="flex flex-wrap gap-2">
              <button
                onClick={() => onUpdate(itemKey, { showAssign: !state.showAssign, showComment: false })}
                className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium transition bg-blue-500/20 text-blue-400 hover:bg-blue-500/30"
              >
                👤 Assign
              </button>
              <button
                onClick={() => onUpdate(itemKey, { showComment: !state.showComment, showAssign: false })}
                className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium transition bg-white/10 text-slate-300 hover:bg-white/15"
              >
                💬 Comment
              </button>
              {isAttention && (
                <button
                  onClick={() => {
                    onUpdate(itemKey, {
                      status: 'resolved',
                      activityLog: [...state.activityLog, { text: 'Marked as Resolved', time: 'Just now' }],
                    });
                    onClose();
                  }}
                  className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium transition bg-green-500/20 text-green-400 hover:bg-green-500/30"
                >
                  ✓ Mark Resolved
                </button>
              )}
              <button
                onClick={() => {
                  onUpdate(itemKey, {
                    status: 'in_progress',
                    activityLog: [...state.activityLog, { text: 'Acknowledged', time: 'Just now' }],
                  });
                  onClose();
                }}
                className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium transition bg-white/10 text-slate-400 hover:bg-white/15"
              >
                Acknowledge
              </button>
            </div>

            {/* Assign dropdown */}
            {state.showAssign && (
              <div className="bg-slate-900 border border-white/15 rounded-lg shadow-xl py-1 w-56">
                {TEAM_MEMBERS.map((member) => (
                  <div
                    key={member}
                    className="px-3 py-2 text-sm text-slate-300 hover:bg-white/10 cursor-pointer"
                    onClick={() => onUpdate(itemKey, {
                      assignedTo: member,
                      status: 'assigned',
                      showAssign: false,
                      activityLog: [...state.activityLog, { text: `Assigned to ${member.split(' (')[0]}`, time: 'Just now' }],
                    })}
                  >
                    {member}
                  </div>
                ))}
              </div>
            )}

            {/* Comment input */}
            {state.showComment && (
              <div className="flex gap-2">
                <input
                  type="text"
                  value={state.commentDraft}
                  onChange={(e) => onUpdate(itemKey, { commentDraft: e.target.value })}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter' && state.commentDraft.trim()) {
                      onUpdate(itemKey, {
                        comments: [...state.comments, { text: state.commentDraft.trim(), time: 'Just now', author: 'You' }],
                        activityLog: [...state.activityLog, { text: `Commented: "${state.commentDraft.trim()}"`, time: 'Just now' }],
                        commentDraft: '',
                        showComment: false,
                      });
                    }
                  }}
                  placeholder="Add a comment..."
                  className="bg-white/10 border border-white/15 rounded-lg px-3 py-1.5 text-sm text-white placeholder-slate-500 flex-1 outline-none focus:border-white/30"
                />
                <button
                  onClick={() => {
                    if (state.commentDraft.trim()) {
                      onUpdate(itemKey, {
                        comments: [...state.comments, { text: state.commentDraft.trim(), time: 'Just now', author: 'You' }],
                        activityLog: [...state.activityLog, { text: `Commented: "${state.commentDraft.trim()}"`, time: 'Just now' }],
                        commentDraft: '',
                        showComment: false,
                      });
                    }
                  }}
                  className="bg-blue-600 text-white px-3 py-1.5 rounded-lg text-xs font-medium hover:bg-blue-500 transition"
                >
                  Post
                </button>
              </div>
            )}
          </div>

          {/* Section 5: Activity Log */}
          {(state.comments.length > 0 || state.activityLog.length > 0) && (
            <div className="bg-white/5 rounded-xl border border-white/10 p-4 space-y-3">
              <p className="text-xs font-bold uppercase tracking-wider text-slate-400 flex items-center gap-2">
                <span>📋</span> Activity Log
              </p>
              <div className="space-y-2">
                {state.activityLog.map((entry, i) => (
                  <div key={`log-${i}`} className="flex items-start gap-2.5 text-xs">
                    <span className="shrink-0 w-1.5 h-1.5 rounded-full bg-slate-500 mt-1.5" />
                    <div className="flex-1">
                      <span className="text-slate-400">{entry.text}</span>
                      <span className="text-slate-400 ml-2">{entry.time}</span>
                    </div>
                  </div>
                ))}
                {state.comments.map((c, i) => (
                  <div key={`comment-${i}`} className="bg-white/5 rounded-lg px-3 py-2">
                    <div className="flex items-center gap-2 mb-1">
                      <span className="text-[10px] font-semibold text-blue-400">{c.author}</span>
                      <span className="text-[10px] text-slate-400">{c.time}</span>
                    </div>
                    <p className="text-xs text-slate-300">{c.text}</p>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

function BriefingCard({
  title,
  items,
  accentColor,
  borderColor,
  bgColor,
  headerColor,
  columnKey,
  itemStates,
  onItemAction,
  onOpenModal,
}: {
  title: string;
  items: BriefingItem[];
  accentColor: string;
  borderColor: string;
  bgColor: string;
  headerColor: string;
  columnKey: string;
  itemStates: Record<string, ItemState>;
  onItemAction: (key: string, patch: Partial<ItemState>) => void;
  onOpenModal: (key: string) => void;
}) {
  return (
    <div
      className={`rounded-2xl border ${borderColor} ${bgColor} overflow-hidden flex flex-col`}
      style={{ borderLeftWidth: '3px' }}
    >
      <div className="px-5 py-4 border-b border-white/5">
        <h3 className={`text-sm font-bold uppercase tracking-wider ${headerColor}`}>{title}</h3>
      </div>
      <div className="flex-1 divide-y divide-white/5">
        {items.map((item, idx) => {
          const key = `${columnKey}-${idx}`;
          const state = itemStates[key] ?? defaultItemState();
          return (
            <div
              key={idx}
              className="px-5 py-3.5 group cursor-pointer hover:bg-white/3 transition-colors"
              onClick={() => onOpenModal(key)}
            >
              <div className="flex items-start gap-3">
                <span className="text-base shrink-0 mt-0.5">{item.icon}</span>
                <div className="flex-1 min-w-0">
                  <div className="flex items-start justify-between gap-2">
                    <div className="flex items-center gap-2 flex-wrap">
                      <p className="text-sm font-semibold text-white leading-tight">{item.title}</p>
                      <StatusBadge status={state.status} />
                      {state.assignedTo && (
                        <span className="text-[10px] bg-blue-500/20 text-blue-400 px-2 py-0.5 rounded-full font-semibold">
                          {state.assignedTo.split(' (')[0]}
                        </span>
                      )}
                    </div>
                    <span className="shrink-0 text-slate-400 group-hover:text-slate-400 transition-colors text-xs mt-0.5">↗</span>
                  </div>
                  <p className="text-xs text-slate-400 mt-1 leading-relaxed">{item.impact}</p>
                  <p className="text-[10px] text-slate-400 mt-1 font-medium uppercase tracking-wide">{item.branch}</p>
                  {item.metric && (
                    <span className="inline-block mt-1.5 text-[10px] font-bold bg-white/5 text-slate-300 px-2 py-0.5 rounded-full border border-white/8">
                      {item.metric}
                    </span>
                  )}
                </div>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}

// ── Feature 6: AI Recommendation Cards ───────────────────────────────────────

interface AIRecommendation {
  icon: string;
  title: string;
  desc: string;
  confidence: 'High' | 'Medium' | 'Low';
  projectedImpact: string;
  riskLevel: 'Low' | 'Medium' | 'High';
  implementationSteps: string[];
  explanation: string;
}

const AI_RECOMMENDATIONS: AIRecommendation[] = [
  {
    icon: '🧠',
    title: 'Dynamic Pricing Opportunity',
    desc: 'Increase Churrasco Chimichurri by 8% (EGP 520→562). Model predicts <3% volume drop, +EGP 4,200/month profit.',
    confidence: 'High',
    projectedImpact: '+EGP 4,200/month net profit · <3% volume impact · ROI: immediate',
    riskLevel: 'Low',
    explanation: 'Price elasticity analysis on Churrasco Chimichurri shows demand inelasticity in the EGP 500–600 range. The dish has a 4.8/5 guest satisfaction score and is a "star" menu item with high volume AND high margin. An 8% price increase brings it in line with comparable premium protein dishes at competing restaurants in the same segment. Historical data from similar adjustments on Lomo Saltado showed <2% volume reduction.',
    implementationSteps: [
      'Update POS price for Churrasco Chimichurri from EGP 520 to EGP 562 at all branches',
      'Brief servers — frame as "premium ingredient upgrade" if guests ask',
      'Monitor volume for 7 days — revert if >5% drop detected',
      'Re-evaluate pricing for 3 other star items using same elasticity model',
    ],
  },
  {
    icon: '📅',
    title: 'Schedule Optimization',
    desc: 'Reduce Tuesday lunch shift by 2 staff at Nimbu El Gouna. Historical covers suggest overstaffing. Save EGP 1,800/week.',
    confidence: 'Medium',
    projectedImpact: 'EGP 1,800/week labor savings · Affects 2 staff positions · No service impact projected',
    riskLevel: 'Low',
    explanation: 'Analysis of 8 weeks of Tuesday lunch data at Nimbu El Gouna shows average covers of 18–22 per lunch service. Current staffing of 5 FOH + 3 BOH is calibrated for 35+ covers. Reducing to 3 FOH + 2 BOH matches the actual cover load and keeps staff-to-cover ratio within optimal range (1:7). The 2 affected staff can be shifted to higher-demand weekend slots, improving their own hours without reducing headcount.',
    implementationSteps: [
      'Pull affected staff from Tuesday lunch — reassign to Friday/Saturday evening slots',
      'Run reduced staffing for 2 Tuesdays — monitor service quality scores',
      'Check that kitchen output isn\'t gated by FOH speed before finalizing',
      'Apply same analysis to other low-demand day-parts across the chain',
    ],
  },
  {
    icon: '🎯',
    title: 'Win-Back Campaign',
    desc: 'Target 5 lapsed VIP guests with personalized Pisco Sour offer. Similar campaigns recovered 60% of at-risk guests.',
    confidence: 'High',
    projectedImpact: 'Projected recovery: 3/5 guests (60%) · Recaptured value: EGP 7,200/month',
    riskLevel: 'Low',
    explanation: 'Based on historical win-back campaign data across 3 comparable restaurants in the Nimbu network, personalized outreach with a high-value complimentary item recovered 60% of at-risk VIP guests within 14 days. The 5 identified guests have a combined monthly value of EGP 12,000. A 60% recovery rate = EGP 7,200/month restored, at a campaign cost of approximately EGP 750 (5 complimentary Pisco Sours at cost + personalized note). ROI: 8.6× in first month alone.',
    implementationSteps: [
      'Draft personalized message for each of the 5 guests — reference their last visit and favorite dish',
      'Attach a complimentary Pisco Sour voucher valid for 14 days',
      'Send via WhatsApp (preferred) or email — GM to send from personal number for authenticity',
      'Follow up if no response within 5 days with a second touchpoint',
      'Track redemption and subsequent visit frequency over 30 days',
    ],
  },
  {
    icon: '📦',
    title: 'Forward-Purchase Pisco',
    desc: 'Buy 3-month Pisco supply before Q2 price increase. Save ~EGP 25,000 vs spot buying.',
    confidence: 'High',
    projectedImpact: 'EGP 25,000 savings over 3 months · Requires EGP 180,000 upfront',
    riskLevel: 'Medium',
    explanation: 'Specialty Imports has indicated a 15% Pisco price increase in Q2. At current consumption (35 units/day chain-wide), a 3-month forward purchase locks in current pricing and saves approximately EGP 25,000.',
    implementationSteps: ['Negotiate forward contract with Specialty Imports', 'Approve EGP 180,000 purchase order', 'Arrange storage at Nimbu Zayed warehouse', 'Monitor inventory drawdown weekly'],
  },
  {
    icon: '👥',
    title: 'Cross-Train Ceviche Station',
    desc: 'Train 3 additional cooks on ceviche prep to eliminate single-point-of-failure at Nimbu North Coast.',
    confidence: 'High',
    projectedImpact: 'Reduce ticket time by 33% · Eliminate station bottleneck · Improve resilience',
    riskLevel: 'Low',
    explanation: 'The Nimbu North Coast ticket time spike (12→18 min) is caused by a single undertrained cook on the Ceviche Bar. Cross-training 3 existing cooks eliminates this vulnerability.',
    implementationSteps: ['Identify 3 cooks with highest ELU scores on adjacent stations', 'Schedule 2-hour ceviche training sessions over 1 week', 'Shadow shifts with experienced ceviche cook from El Gouna', 'Validate ELU rating >1.0 before solo assignment'],
  },
  {
    icon: '🎪',
    title: 'Launch Ramadan Menu',
    desc: 'Create a limited Iftar set menu (EGP 450/person). Historical data shows 30% dinner volume increase during Ramadan.',
    confidence: 'Medium',
    projectedImpact: 'Projected +EGP 120,000 revenue during Ramadan · 30-day campaign',
    riskLevel: 'Medium',
    explanation: 'Ramadan starts April 2. Historical data from comparable restaurants shows a 30% increase in dinner covers during Ramadan, with strong demand for set menus and family-style dining.',
    implementationSteps: ['Design 3-tier Iftar menu (Basic EGP 350, Classic EGP 450, Premium EGP 650)', 'Source dates, laban, and traditional Ramadan items', 'Train staff on Iftar service flow and timing', 'Launch social media campaign 1 week before Ramadan'],
  },
  {
    icon: '📊',
    title: 'Renegotiate Delivery Commissions',
    desc: 'Nimbu Zayed delivery margin dropped to 18%. Negotiate platform commission from 28% to 22% or adjust delivery prices.',
    confidence: 'Medium',
    projectedImpact: 'Restore delivery margin to 24% · +EGP 8,400/month at Zayed alone',
    riskLevel: 'Low',
    explanation: 'Delivery app commission at Nimbu Zayed increased from 22% to 28% without notice. This pushed delivery channel margin from 24% to 18%, making delivery nearly unprofitable.',
    implementationSteps: ['Contact platform account manager — request commission review', 'Prepare switching threat (competitor platform offers 20%)', 'If negotiation fails, increase delivery menu prices by 6%', 'Consider launching direct ordering via WhatsApp to bypass commission'],
  },
  {
    icon: '🏆',
    title: 'Propagate Zayed Best Practice',
    desc: 'Nimbu Zayed\'s ceviche prep scheduling reduces fish waste by 22%. Roll out to all branches.',
    confidence: 'High',
    projectedImpact: 'Chain-wide fish waste reduction: 22% · Save ~EGP 15,000/month across 4 branches',
    riskLevel: 'Low',
    explanation: 'Nimbu Zayed runs demand-based ceviche prep scheduling instead of batch prep. This reduces Sea Bass and Shrimp waste by 22% compared to the chain average.',
    implementationSteps: ['Document Zayed ceviche prep SOP in detail', 'Schedule training sessions at El Gouna, New Cairo, North Coast', 'Implement demand-forecast-linked prep scheduling at each branch', 'Track waste reduction weekly for 30 days'],
  },
  {
    icon: '⚡',
    title: 'Install Kitchen Display Screens',
    desc: 'Replace paper tickets with KDS screens at Nimbu North Coast. Reduce ticket errors by 40%.',
    confidence: 'Medium',
    projectedImpact: '-40% ticket errors · -15% food waste from wrong orders · EGP 6,000/month savings',
    riskLevel: 'Low',
    explanation: 'Nimbu North Coast still uses paper tickets. Analysis of void/remake data shows 8% of tickets have errors vs 3% at KDS-equipped branches. Digital ticket routing also improves station coordination.',
    implementationSteps: ['Purchase 3 KDS screens for North Coast kitchen (budget: EGP 15,000)', 'Install and configure with FireLine KDS module', 'Train kitchen staff over 2 days', 'Run parallel paper+digital for 1 week before full cutover'],
  },
];

interface RecState {
  modalOpen: boolean;
  status: 'pending' | 'approved' | 'dismissed';
  assignedTo: string | null;
  showAssign: boolean;
  showModify: boolean;
  modifyDraft: string;
}

function defaultRecState(): RecState {
  return {
    modalOpen: false,
    status: 'pending',
    assignedTo: null,
    showAssign: false,
    showModify: false,
    modifyDraft: '',
  };
}

// ── Recommendation Detail Modal ───────────────────────────────────────────────

function RecommendationModal({
  rec,
  state,
  onUpdate,
  onClose,
}: {
  rec: AIRecommendation;
  state: RecState;
  onUpdate: (patch: Partial<RecState>) => void;
  onClose: () => void;
}) {
  function handleBackdropClick(e: React.MouseEvent<HTMLDivElement>) {
    if (e.target === e.currentTarget) onClose();
  }

  useEffect(() => {
    function onKey(e: KeyboardEvent) {
      if (e.key === 'Escape') onClose();
    }
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [onClose]);

  const riskColor = {
    Low: 'bg-green-500/20 text-green-400 border-green-500/30',
    Medium: 'bg-amber-500/20 text-amber-400 border-amber-500/30',
    High: 'bg-red-500/20 text-red-400 border-red-500/30',
  }[rec.riskLevel];

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center p-4"
      style={{ background: 'rgba(0,0,0,0.65)', backdropFilter: 'blur(4px)' }}
      onClick={handleBackdropClick}
    >
      <div
        className="bg-slate-800 border border-white/15 rounded-2xl shadow-2xl max-w-2xl w-full max-h-[85vh] overflow-y-auto"
        style={{ scrollbarWidth: 'thin', scrollbarColor: 'rgba(255,255,255,0.1) transparent' }}
      >
        {/* Header */}
        <div className="sticky top-0 bg-slate-800 border-b border-white/10 px-6 py-4 rounded-t-2xl z-10">
          <div className="flex items-start gap-3">
            <span className="text-2xl shrink-0 mt-0.5">{rec.icon}</span>
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-2 flex-wrap mb-1">
                <span className={`text-[10px] font-bold uppercase tracking-wider px-2 py-0.5 rounded-full ${
                  rec.confidence === 'High'
                    ? 'bg-green-500/20 text-green-400 border border-green-500/30'
                    : rec.confidence === 'Medium'
                    ? 'bg-amber-500/20 text-amber-400 border border-amber-500/30'
                    : 'bg-slate-500/20 text-slate-400 border border-slate-500/30'
                }`}>
                  {rec.confidence} Confidence
                </span>
                <span className={`text-[10px] font-bold uppercase tracking-wider px-2 py-0.5 rounded-full border ${riskColor}`}>
                  {rec.riskLevel} Risk
                </span>
                {state.status === 'approved' && (
                  <span className="text-[10px] font-bold bg-green-500/20 text-green-400 border border-green-500/30 px-2 py-0.5 rounded-full">
                    Approved ✓
                  </span>
                )}
                {state.assignedTo && (
                  <span className="text-[10px] bg-blue-500/20 text-blue-400 px-2 py-0.5 rounded-full font-semibold">
                    {state.assignedTo.split(' (')[0]}
                  </span>
                )}
              </div>
              <h2 className="text-lg font-bold text-white leading-tight">{rec.title}</h2>
            </div>
            <button
              onClick={onClose}
              className="shrink-0 w-8 h-8 flex items-center justify-center rounded-lg bg-white/5 hover:bg-white/15 text-slate-400 hover:text-white transition-colors text-lg leading-none"
            >
              ×
            </button>
          </div>
        </div>

        <div className="px-6 py-5 space-y-4">
          {/* Summary */}
          <div className="bg-white/5 rounded-xl border border-white/10 p-4 space-y-3">
            <p className="text-xs font-bold uppercase tracking-wider text-slate-400 flex items-center gap-2">
              <span>📊</span> Recommendation
            </p>
            <p className="text-sm text-slate-300 leading-relaxed">{rec.desc}</p>
          </div>

          {/* Projected Impact */}
          <div className="bg-white/5 rounded-xl border border-white/10 p-4 space-y-3">
            <p className="text-xs font-bold uppercase tracking-wider text-slate-400 flex items-center gap-2">
              <span>💹</span> Projected Impact
            </p>
            <div className="bg-white/5 rounded-lg px-4 py-3 border border-white/8">
              <p className="text-sm font-bold text-emerald-400">{rec.projectedImpact}</p>
            </div>
          </div>

          {/* Detailed Explanation */}
          <div className="bg-white/5 rounded-xl border border-white/10 p-4 space-y-3">
            <p className="text-xs font-bold uppercase tracking-wider text-slate-400 flex items-center gap-2">
              <span>🔍</span> AI Analysis
            </p>
            <p className="text-sm text-slate-300 leading-relaxed">{rec.explanation}</p>
          </div>

          {/* Implementation Steps */}
          <div className="bg-white/5 rounded-xl border border-white/10 p-4 space-y-3">
            <p className="text-xs font-bold uppercase tracking-wider text-slate-400 flex items-center gap-2">
              <span>🛡️</span> Implementation Steps
            </p>
            <ol className="space-y-2">
              {rec.implementationSteps.map((step, i) => (
                <li key={i} className="flex items-start gap-2.5 text-xs text-slate-300">
                  <span className="shrink-0 w-4 h-4 rounded-full bg-indigo-500/20 text-indigo-400 text-[10px] font-bold flex items-center justify-center mt-0.5">
                    {i + 1}
                  </span>
                  <span>{step}</span>
                </li>
              ))}
            </ol>
          </div>

          {/* Actions */}
          <div className="bg-white/5 rounded-xl border border-white/10 p-4 space-y-3">
            <p className="text-xs font-bold uppercase tracking-wider text-slate-400 flex items-center gap-2">
              <span>⚡</span> Actions
            </p>
            <div className="flex flex-wrap gap-2">
              <button
                onClick={() => { onUpdate({ status: 'approved' }); onClose(); }}
                className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium transition bg-green-500/20 text-green-400 hover:bg-green-500/30"
              >
                ✓ Approve & Execute
              </button>
              <button
                onClick={() => onUpdate({ showModify: !state.showModify, showAssign: false })}
                className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium transition bg-blue-500/20 text-blue-400 hover:bg-blue-500/30"
              >
                ✏️ Modify
              </button>
              <button
                onClick={() => onUpdate({ showAssign: !state.showAssign, showModify: false })}
                className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium transition bg-white/10 text-slate-300 hover:bg-white/15"
              >
                👤 Assign to Team
              </button>
              <button
                onClick={() => { onUpdate({ status: 'dismissed' }); onClose(); }}
                className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium transition bg-white/10 text-slate-400 hover:bg-white/15"
              >
                Dismiss
              </button>
            </div>

            {state.showAssign && (
              <div className="bg-slate-900 border border-white/15 rounded-lg shadow-xl py-1 w-56">
                {TEAM_MEMBERS.map((member) => (
                  <div
                    key={member}
                    className="px-3 py-2 text-sm text-slate-300 hover:bg-white/10 cursor-pointer"
                    onClick={() => onUpdate({ assignedTo: member, showAssign: false })}
                  >
                    {member}
                  </div>
                ))}
              </div>
            )}

            {state.showModify && (
              <div className="space-y-2">
                <textarea
                  rows={3}
                  value={state.modifyDraft}
                  onChange={(e) => onUpdate({ modifyDraft: e.target.value })}
                  placeholder="Describe how you'd like to adjust this recommendation..."
                  className="w-full bg-white/10 border border-white/15 rounded-lg px-3 py-2 text-sm text-white placeholder-slate-500 outline-none focus:border-white/30 resize-none"
                />
                <button
                  onClick={() => { onUpdate({ showModify: false, status: 'approved' }); onClose(); }}
                  className="bg-blue-600 text-white px-3 py-1.5 rounded-lg text-xs font-medium hover:bg-blue-500 transition"
                >
                  Save & Approve Modified
                </button>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

function AIRecommendations() {
  const [recStates, setRecStates] = useState<Record<number, RecState>>({});

  function updateRec(i: number, patch: Partial<RecState>) {
    setRecStates((prev) => ({
      ...prev,
      [i]: { ...(prev[i] ?? defaultRecState()), ...patch },
    }));
  }

  const openModalIdx = Object.entries(recStates).find(([, s]) => s.modalOpen)?.[0];

  return (
    <div className="space-y-4 mt-6">
      <div className="flex items-center gap-3">
        <div className="h-px flex-1 bg-white/5" />
        <div className="flex items-center gap-2">
          <span className="text-base">🧠</span>
          <span className="text-xs font-bold text-white uppercase tracking-widest">AI Recommendations</span>
          <span className="text-[10px] bg-indigo-500/30 text-indigo-300 px-2 py-0.5 rounded-full font-bold">
            {AI_RECOMMENDATIONS.filter((_, i) => !(recStates[i]?.status === 'approved' || recStates[i]?.status === 'dismissed')).length} pending
          </span>
        </div>
        <div className="h-px flex-1 bg-white/5" />
      </div>
      <div className="overflow-x-auto pb-2 -mx-2 px-2">
        <div className="flex gap-4" style={{ minWidth: 'max-content' }}>
        {AI_RECOMMENDATIONS.map((rec, i) => {
          const state = recStates[i] ?? defaultRecState();
          if (state.status === 'approved' || state.status === 'dismissed') return null;
          return (
            <div
              key={i}
              onClick={() => updateRec(i, { modalOpen: true })}
              className="bg-gradient-to-br from-indigo-950/40 to-purple-950/30 border border-indigo-500/20 hover:border-indigo-400/40 hover:scale-[1.02] rounded-xl p-4 transition-all cursor-pointer group w-60 sm:w-72 flex-shrink-0"
            >
              <div className="flex items-center justify-between gap-2 mb-2">
                <div className="flex items-center gap-2">
                  <span className="text-lg">{rec.icon}</span>
                  <span
                    className={`text-[10px] font-bold uppercase tracking-wider px-2 py-0.5 rounded-full ${
                      rec.confidence === 'High'
                        ? 'bg-green-500/20 text-green-400'
                        : 'bg-amber-500/20 text-amber-400'
                    }`}
                  >
                    {rec.confidence} Confidence
                  </span>
                  {state.status === 'approved' && (
                    <span className="text-[10px] font-bold bg-green-500/20 text-green-400 px-2 py-0.5 rounded-full">
                      Approved ✓
                    </span>
                  )}
                </div>
                <span className="text-slate-400 group-hover:text-slate-400 text-xs transition-colors">↗</span>
              </div>
              <h4 className="text-sm font-bold text-white mb-1">{rec.title}</h4>
              <p className="text-xs text-slate-400 leading-relaxed">{rec.desc}</p>
              {state.assignedTo && (
                <span className="mt-2 inline-block text-[10px] bg-blue-500/20 text-blue-400 px-2 py-0.5 rounded-full font-semibold">
                  {state.assignedTo.split(' (')[0]}
                </span>
              )}
              <p className="text-[10px] text-indigo-400/60 mt-2 font-medium">Click for full analysis →</p>
            </div>
          );
        })}
        </div>
      </div>

      {/* Empty state */}
      {AI_RECOMMENDATIONS.every((_, i) => {
        const s = recStates[i];
        return s && (s.status === 'approved' || s.status === 'dismissed');
      }) && (
        <div className="text-center py-8 text-slate-300 text-sm">
          <span className="text-2xl block mb-2">✨</span>
          All recommendations acted on. New AI insights will appear as patterns are detected.
        </div>
      )}

      {/* Recommendation Modal */}
      {openModalIdx !== undefined && (() => {
        const idx = Number(openModalIdx);
        const state = recStates[idx] ?? defaultRecState();
        return (
          <RecommendationModal
            rec={AI_RECOMMENDATIONS[idx]}
            state={state}
            onUpdate={(patch) => updateRec(idx, patch)}
            onClose={() => updateRec(idx, { modalOpen: false, showAssign: false, showModify: false })}
          />
        );
      })()}
    </div>
  );
}

// Maps a briefing column key + item index to the matching item data
function getBriefingItem(key: string): BriefingItem | null {
  const [col, idxStr] = key.split('-');
  const idx = Number(idxStr);
  if (col === 'attention') return ATTENTION_ITEMS[idx] ?? null;
  if (col === 'highlights') return HIGHLIGHT_ITEMS[idx] ?? null;
  if (col === 'outlook') return OUTLOOK_ITEMS[idx] ?? null;
  return null;
}

function CEOBriefing() {
  const [visible, setVisible] = useState(false);
  const ref = useRef<HTMLDivElement>(null);
  const [itemStates, setItemStates] = useState<Record<string, ItemState>>({});
  const [modalKey, setModalKey] = useState<string | null>(null);

  function handleItemAction(key: string, patch: Partial<ItemState>) {
    setItemStates((prev) => ({
      ...prev,
      [key]: { ...(prev[key] ?? defaultItemState()), ...patch },
    }));
  }

  function handleOpenModal(key: string) {
    setModalKey(key);
  }

  function handleCloseModal() {
    setModalKey(null);
  }

  useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => { if (entry.isIntersecting) setVisible(true); },
      { threshold: 0.1 }
    );
    if (ref.current) observer.observe(ref.current);
    return () => observer.disconnect();
  }, []);

  const activeItem = modalKey ? getBriefingItem(modalKey) : null;
  const activeState = modalKey ? (itemStates[modalKey] ?? defaultItemState()) : null;
  const isAttentionModal = modalKey?.startsWith('attention') ?? false;

  return (
    <div
      ref={ref}
      className="space-y-6"
      style={{
        opacity: visible ? 1 : 0,
        transform: visible ? 'translateY(0)' : 'translateY(24px)',
        transition: 'opacity 0.6s ease, transform 0.6s ease',
      }}
    >
      {/* Section header */}
      <div className="flex items-center gap-4">
        <div className="h-px flex-1 bg-white/5" />
        <div className="flex items-center gap-2.5">
          <Zap className="h-4 w-4 text-amber-400" />
          <span className="text-sm font-bold text-white uppercase tracking-widest">Executive Briefing</span>
          <span className="text-xs text-slate-300 font-medium">· Updated just now</span>
        </div>
        <div className="h-px flex-1 bg-white/5" />
      </div>

      {/* 3-column grid */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
        <BriefingCard
          title="Requires Attention"
          items={ATTENTION_ITEMS}
          accentColor="red"
          borderColor="border-red-500/40"
          bgColor="bg-red-950/20"
          headerColor="text-red-400"
          columnKey="attention"
          itemStates={itemStates}
          onItemAction={handleItemAction}
          onOpenModal={handleOpenModal}
        />
        <BriefingCard
          title="Performance Highlights"
          items={HIGHLIGHT_ITEMS}
          accentColor="green"
          borderColor="border-green-500/40"
          bgColor="bg-green-950/20"
          headerColor="text-green-400"
          columnKey="highlights"
          itemStates={itemStates}
          onItemAction={handleItemAction}
          onOpenModal={handleOpenModal}
        />
        <BriefingCard
          title="Strategic Outlook"
          items={OUTLOOK_ITEMS}
          accentColor="blue"
          borderColor="border-blue-500/40"
          bgColor="bg-blue-950/20"
          headerColor="text-blue-400"
          columnKey="outlook"
          itemStates={itemStates}
          onItemAction={handleItemAction}
          onOpenModal={handleOpenModal}
        />
      </div>

      {/* Briefing Detail Modal */}
      {modalKey && activeItem && activeState && (
        <BriefingDetailModal
          item={activeItem}
          itemKey={modalKey}
          isAttention={isAttentionModal}
          state={activeState}
          onUpdate={handleItemAction}
          onClose={handleCloseModal}
        />
      )}

      {/* Feature 6: AI Recommendation Cards */}
      <AIRecommendations />
    </div>
  );
}

// ── Main Page ─────────────────────────────────────────────────────────────────

const LOCATION_CITIES: Record<string, string> = {
  'Nimbu El Gouna': 'El Gouna, Red Sea',
  'Nimbu New Cairo': 'New Cairo, Cairo',
  'Nimbu Zayed': 'Sheikh Zayed, Giza',
  'Nimbu North Coast': 'North Coast',
};

export default function PortfolioPage() {
  const navigate = useNavigate();
  const { locations, selectedLocationId, setLocation, loadLocations } = useLocationStore();
  const isAuthenticated = true; // already inside ProtectedRoute

  useEffect(() => {
    loadLocations();
  }, [loadLocations]);

  // Fetch data for all locations
  const loc0 = locations[0];
  const loc1 = locations[1];
  const loc2 = locations[2];
  const loc3 = locations[3];

  const data0 = useBranchData(loc0?.id ?? '');
  const data1 = useBranchData(loc1?.id ?? '');
  const data2 = useBranchData(loc2?.id ?? '');
  const data3 = useBranchData(loc3?.id ?? '');

  const allData = [data0, data1, data2, data3].slice(0, locations.length);

  // Compute chain-wide KPIs
  const totalRevenue = allData.reduce((s, d) => s + (d.pnl.data?.net_revenue ?? 0), 0);
  const totalChecks = allData.reduce((s, d) => s + (d.pnl.data?.check_count ?? 0), 0);
  const healthScores = allData.map((d) => d.health.data?.overall_score ?? 0).filter((v) => v > 0);
  const avgHealth = healthScores.length ? Math.round(healthScores.reduce((a, b) => a + b, 0) / healthScores.length) : 0;
  const totalAlerts = allData.reduce((s, d) => s + (d.alertCount.data?.count ?? 0), 0);

  const handleBranchClick = (locationId: string) => {
    setLocation(locationId);
    navigate('/dashboard');
  };

  // Build per-branch rows for comparison table
  const branchRows: BranchKPIRow[] = [data0, data1, data2, data3]
    .slice(0, locations.length)
    .map((d, idx) => ({
      name: locations[idx]?.name ?? '',
      shortName: locations[idx]?.name ?? '',
      revenue: d.pnl.data?.net_revenue ?? 0,
      margin: d.pnl.data?.gross_margin ?? 0,
      health: d.health.data?.overall_score ?? 0,
      alerts: d.alertCount.data?.count ?? 0,
    }));

  return (
    <div className="min-h-full relative overflow-hidden">
      {/* Inject ticker animation CSS */}
      <style>{tickerStyle}</style>

      {/* Animated background */}
      <div
        className="absolute inset-0 -z-10"
        style={{
          background: 'linear-gradient(135deg, #0f172a 0%, #1e293b 35%, #0f172a 65%, #1a1f35 100%)',
        }}
      />
      {/* Subtle radial glow */}
      <div
        className="absolute inset-0 -z-10 opacity-40 pointer-events-none"
        style={{
          background: 'radial-gradient(ellipse 80% 50% at 50% -20%, rgba(249,115,22,0.15) 0%, transparent 60%)',
        }}
      />
      {/* Grid pattern */}
      <div
        className="absolute inset-0 -z-10 opacity-[0.03] pointer-events-none"
        style={{
          backgroundImage: 'linear-gradient(rgba(255,255,255,0.5) 1px, transparent 1px), linear-gradient(90deg, rgba(255,255,255,0.5) 1px, transparent 1px)',
          backgroundSize: '40px 40px',
        }}
      />

      <div className="relative py-6 space-y-8 overflow-hidden">
        {/* Hero Header */}
        <div className="text-center space-y-4 pt-4">
          <div className="flex items-center justify-center gap-3 mb-2">
            <div className="h-px w-16 bg-gradient-to-r from-transparent to-[#F97316]/60" />
            <span className="flex items-center gap-2 text-[#F97316] text-xs font-semibold uppercase tracking-widest">
              <Zap className="h-3.5 w-3.5" />
              FireLine by OpsNerve
            </span>
            <div className="h-px w-16 bg-gradient-to-l from-transparent to-[#F97316]/60" />
          </div>

          <h1 className="text-3xl sm:text-5xl md:text-6xl font-black text-white tracking-tight leading-none">
            Nimbu
          </h1>
          <p className="text-lg text-slate-400 font-medium max-w-xl mx-auto">
            AI-Powered Operations Command Center
          </p>

          {avgHealth > 0 && <HealthPulse score={avgHealth} />}
        </div>

        {/* Chain KPI Summary */}
        {locations.length > 0 && (
          <ChainKPIBar
            kpi={{ totalRevenue, totalChecks, avgHealth, totalAlerts }}
          />
        )}

        {/* Section label */}
        <div className="flex items-center gap-4">
          <div className="h-px flex-1 bg-white/5" />
          <span className="text-xs font-semibold uppercase tracking-widest text-slate-300">
            {locations.length} Branch{locations.length !== 1 ? 'es' : ''} — Select to Open Dashboard
          </span>
          <div className="h-px flex-1 bg-white/5" />
        </div>

        {/* Branch Cards Grid */}
        {locations.length > 0 ? (
          <div className="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-4 gap-5">
            {locations.map((loc, idx) => (
              <BranchCard
                key={loc.id}
                locationId={loc.id}
                name={loc.name}
                city={LOCATION_CITIES[loc.name] ?? loc.name}
                seed={idx * 7 + 3}
                onClick={() => handleBranchClick(loc.id)}
              />
            ))}
          </div>
        ) : (
          <div className="text-center py-20">
            <div className="inline-flex h-12 w-12 items-center justify-center rounded-full bg-white/5 mb-4">
              <Activity className="h-6 w-6 text-slate-300" />
            </div>
            <p className="text-slate-400">Loading branch data...</p>
          </div>
        )}

        {/* Chain KPI Comparison Table + Bar Chart */}
        {branchRows.length > 0 && (
          <ChainComparisonTable branches={branchRows} />
        )}

        {/* Feature 2: AI Insights Ticker */}
        <div className="-mx-4 sm:-mx-6 lg:-mx-8">
          <AIInsightsTicker />
        </div>

        {/* Feature 4: Revenue Race */}
        {branchRows.length > 0 && (
          <RevenueRace branchRows={branchRows} />
        )}

        {/* CEO Executive Briefing */}
        <CEOBriefing />

        {/* Footer */}
        <div className="text-center pb-6">
          <p className="text-xs text-slate-400">
            FireLine by OpsNerve · Real-time AI operations intelligence · Data refreshes every 30s
          </p>
        </div>
      </div>
    </div>
  );
}
