import { useEffect, useState, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  AreaChart,
  Area,
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
              <MapPin className="h-3 w-3 text-slate-500" />
              <span className="text-xs text-slate-400">{city}</span>
            </div>
          </div>
          <HealthCircle score={isLoading ? 0 : healthScore} size={60} />
        </div>

        {/* KPI Row */}
        <div className="grid grid-cols-3 gap-2">
          <div className="bg-white/5 rounded-xl p-2.5 text-center">
            <p className="text-[10px] text-slate-500 uppercase tracking-wider mb-1">Revenue</p>
            {isLoading ? (
              <div className="h-4 bg-white/10 rounded animate-pulse mx-1" />
            ) : (
              <p className="text-sm font-bold text-white leading-none">
                EGP {animatedRevenue.toLocaleString()}
              </p>
            )}
          </div>
          <div className="bg-white/5 rounded-xl p-2.5 text-center">
            <p className="text-[10px] text-slate-500 uppercase tracking-wider mb-1">Orders</p>
            {isLoading ? (
              <div className="h-4 bg-white/10 rounded animate-pulse mx-1" />
            ) : (
              <p className="text-sm font-bold text-white leading-none">{animatedOrders.toLocaleString()}</p>
            )}
          </div>
          <div className="bg-white/5 rounded-xl p-2.5 text-center">
            <p className="text-[10px] text-slate-500 uppercase tracking-wider mb-1">Margin</p>
            {isLoading ? (
              <div className="h-4 bg-white/10 rounded animate-pulse mx-1" />
            ) : (
              <p className="text-sm font-bold text-white leading-none">{fmtPct(margin)}</p>
            )}
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
              <span className="text-xs text-slate-500">No alerts</span>
            )}
          </div>
          {staffOnShift > 0 && (
            <div className="flex items-center gap-1.5 text-slate-400">
              <Users className="h-3.5 w-3.5" />
              <span className="text-xs">{staffOnShift} on shift</span>
            </div>
          )}
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
          <p className="text-xs text-slate-500 mt-1">{item.label}</p>
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
      <span className="text-sm text-slate-500">— Chain Health {score}/100</span>
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

  return (
    <div className="overflow-x-auto">
      <div className="min-w-[640px]">
        <div className="flex items-center gap-3 mb-4">
          <div className="h-px flex-1 bg-white/5" />
          <span className="text-xs font-semibold uppercase tracking-widest text-slate-500">
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
              <span className="text-xs font-semibold text-slate-500 uppercase tracking-wider">Metric</span>
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
                  {b.revenue > 0 ? Math.round(b.revenue / 10000).toLocaleString() + 'K' : '—'}
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
                  <span className={`text-sm font-bold ${count > 0 ? (i === cols.length - 1 ? 'text-slate-300' : 'text-red-400') : 'text-slate-500'}`}>
                    {count}
                  </span>
                </div>
              );
            })}
          </div>
        </div>
      </div>
    </div>
  );
}

// ── CEO Executive Briefing ────────────────────────────────────────────────────

interface BriefingItem {
  icon: string;
  title: string;
  impact: string;
  branch: string;
}

const ATTENTION_ITEMS: BriefingItem[] = [
  {
    icon: '🔴',
    title: 'Food Cost Breach — New Cairo',
    impact: 'Trending 34.2% vs 32% target. Protein costs drove +EGP 18,000 excess this month.',
    branch: 'New Cairo · Action: Review Sea Bass sourcing',
  },
  {
    icon: '🔴',
    title: 'VIP Customer Churn Risk',
    impact: '5 high-CLV guests (avg EGP 2,400/mo) haven\'t visited in 21+ days. Projected loss: EGP 12,000/mo.',
    branch: 'El Gouna & Sheikh Zayed',
  },
  {
    icon: '🟡',
    title: 'North Coast Ticket Time Degradation',
    impact: 'Average ticket time increased from 12→18 min this week. Ceviche Bar is the bottleneck.',
    branch: 'North Coast · Guest complaints likely to follow',
  },
  {
    icon: '🟡',
    title: 'Labor Overtime at 3 Branches',
    impact: '7 staff exceeded 40hr/week. Projected overtime cost: EGP 8,500.',
    branch: 'New Cairo, Zayed, North Coast',
  },
  {
    icon: '🟡',
    title: 'Vendor Reliability Drop — Metro Market',
    impact: 'OTIF rate fell to 72%. 3 short deliveries this month affecting produce quality.',
    branch: 'North Coast',
  },
];

const HIGHLIGHT_ITEMS: BriefingItem[] = [
  {
    icon: '✅',
    title: 'Chain Revenue On Track',
    impact: 'EGP 387K today across 4 branches (target: 400K). Trending +5% vs last week.',
    branch: 'All Branches',
  },
  {
    icon: '✅',
    title: 'Sheikh Zayed Best Performer',
    impact: '12% below chain average on food cost. Recommend propagating their portioning practices.',
    branch: 'Sheikh Zayed',
  },
  {
    icon: '✅',
    title: 'Pisco Hour Campaign Success',
    impact: '45 redemptions, EGP 2,250 attributed revenue. Consider expanding to all branches.',
    branch: 'El Gouna',
  },
  {
    icon: '✅',
    title: 'New Menu Classification',
    impact: 'Empanadas reclassified to \'crowd_pleaser\' after 22% velocity increase post portion adjustment.',
    branch: 'All Branches',
  },
];

const OUTLOOK_ITEMS: BriefingItem[] = [
  {
    icon: '📊',
    title: 'Q2 Pisco Price Forecast',
    impact: 'Specialty Imports projects 15% price increase. Forward-purchasing 3-month supply saves ~EGP 25,000.',
    branch: 'Strategic · Recommend action by April 1',
  },
  {
    icon: '📊',
    title: 'Ramadan Prep Required',
    impact: 'Ramadan starts April 2. Historical: +30% dinner volume, −40% lunch. Schedule adjustments needed.',
    branch: 'All Branches',
  },
  {
    icon: '📊',
    title: 'North Coast Seasonal Ramp',
    impact: 'Summer season begins May 1. Last year\'s volume was 2.3× winter. Hiring pipeline should start now.',
    branch: 'North Coast',
  },
  {
    icon: '📊',
    title: 'Expansion Opportunity',
    impact: 'Zayed consistently exceeds capacity on weekends (92% kitchen utilization). Consider satellite prep kitchen.',
    branch: 'Sheikh Zayed',
  },
];

function BriefingCard({
  title,
  items,
  accentColor,
  borderColor,
  bgColor,
  headerColor,
}: {
  title: string;
  items: BriefingItem[];
  accentColor: string;
  borderColor: string;
  bgColor: string;
  headerColor: string;
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
        {items.map((item, idx) => (
          <div
            key={idx}
            className="group px-5 py-3.5 hover:bg-white/5 transition-colors duration-150 cursor-default flex items-start gap-3"
          >
            <span className="text-base shrink-0 mt-0.5">{item.icon}</span>
            <div className="flex-1 min-w-0">
              <div className="flex items-start justify-between gap-2">
                <p className="text-sm font-semibold text-white leading-tight">{item.title}</p>
                <span className="shrink-0 text-slate-600 group-hover:text-slate-400 transition-colors text-xs mt-0.5">→</span>
              </div>
              <p className="text-xs text-slate-400 mt-1 leading-relaxed">{item.impact}</p>
              <p className="text-[10px] text-slate-600 mt-1 font-medium uppercase tracking-wide">{item.branch}</p>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

function CEOBriefing() {
  const [visible, setVisible] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => { if (entry.isIntersecting) setVisible(true); },
      { threshold: 0.1 }
    );
    if (ref.current) observer.observe(ref.current);
    return () => observer.disconnect();
  }, []);

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
          <span className="text-xs text-slate-500 font-medium">· Updated just now</span>
        </div>
        <div className="h-px flex-1 bg-white/5" />
      </div>

      {/* 3-column grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-5">
        <BriefingCard
          title="Requires Attention"
          items={ATTENTION_ITEMS}
          accentColor="red"
          borderColor="border-red-500/40"
          bgColor="bg-red-950/20"
          headerColor="text-red-400"
        />
        <BriefingCard
          title="Performance Highlights"
          items={HIGHLIGHT_ITEMS}
          accentColor="green"
          borderColor="border-green-500/40"
          bgColor="bg-green-950/20"
          headerColor="text-green-400"
        />
        <BriefingCard
          title="Strategic Outlook"
          items={OUTLOOK_ITEMS}
          accentColor="blue"
          borderColor="border-blue-500/40"
          bgColor="bg-blue-950/20"
          headerColor="text-blue-400"
        />
      </div>
    </div>
  );
}

// ── Main Page ─────────────────────────────────────────────────────────────────

const LOCATION_CITIES: Record<string, string> = {
  'El Gouna': 'El Gouna, Red Sea',
  'New Cairo': 'New Cairo, Cairo',
  'Sheikh Zayed': 'Sheikh Zayed, Giza',
  'North Coast': 'North Coast',
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
      shortName: locations[idx]?.name?.split(' ')[0] ?? '',
      revenue: d.pnl.data?.net_revenue ?? 0,
      margin: d.pnl.data?.gross_margin ?? 0,
      health: d.health.data?.overall_score ?? 0,
      alerts: d.alertCount.data?.count ?? 0,
    }));

  return (
    <div className="min-h-screen relative overflow-x-hidden">
      {/* Animated background */}
      <div
        className="fixed inset-0 -z-10"
        style={{
          background: 'linear-gradient(135deg, #0f172a 0%, #1e293b 35%, #0f172a 65%, #1a1f35 100%)',
        }}
      />
      {/* Subtle radial glow */}
      <div
        className="fixed inset-0 -z-10 opacity-40"
        style={{
          background: 'radial-gradient(ellipse 80% 50% at 50% -20%, rgba(249,115,22,0.15) 0%, transparent 60%)',
        }}
      />
      {/* Grid pattern */}
      <div
        className="fixed inset-0 -z-10 opacity-[0.03]"
        style={{
          backgroundImage: 'linear-gradient(rgba(255,255,255,0.5) 1px, transparent 1px), linear-gradient(90deg, rgba(255,255,255,0.5) 1px, transparent 1px)',
          backgroundSize: '40px 40px',
        }}
      />

      <div className="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-10 space-y-10">
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

          <h1 className="text-5xl sm:text-6xl font-black text-white tracking-tight leading-none">
            Chicha Egypt
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
          <span className="text-xs font-semibold uppercase tracking-widest text-slate-500">
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
              <Activity className="h-6 w-6 text-slate-500" />
            </div>
            <p className="text-slate-400">Loading branch data...</p>
          </div>
        )}

        {/* Chain KPI Comparison Table */}
        {branchRows.length > 0 && (
          <ChainComparisonTable branches={branchRows} />
        )}

        {/* CEO Executive Briefing */}
        <CEOBriefing />

        {/* Footer */}
        <div className="text-center pb-6">
          <p className="text-xs text-slate-600">
            FireLine by OpsNerve · Real-time AI operations intelligence · Data refreshes every 30s
          </p>
        </div>
      </div>
    </div>
  );
}
