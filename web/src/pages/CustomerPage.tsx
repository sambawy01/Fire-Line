import { useState } from 'react';
import {
  UserCheck,
  DollarSign,
  TrendingUp,
  ShoppingCart,
  RefreshCw,
  AlertTriangle,
  ChevronDown,
  ChevronUp,
} from 'lucide-react';
import {
  PieChart,
  Pie,
  Cell,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
} from 'recharts';
import { useLocationStore } from '../stores/location';
import {
  useGuests,
  useSegments,
  useChurnDist,
  useCLVDist,
  useRefreshAnalytics,
} from '../hooks/useCustomers';
import KPICard from '../components/ui/KPICard';
import DataTable from '../components/ui/DataTable';
import type { Column } from '../components/ui/DataTable';
import StatusBadge from '../components/ui/StatusBadge';
import LoadingSpinner from '../components/ui/LoadingSpinner';
import ErrorBanner from '../components/ui/ErrorBanner';
import type { GuestProfile } from '../lib/api';

// ─── Helpers ────────────────────────────────────────────────────────────────

function dollars(cents: number): string {
  return `EGP ${(cents / 100).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
}

function daysSince(isoDate: string | null): number | null {
  if (!isoDate) return null;
  const diff = Date.now() - new Date(isoDate).getTime();
  return Math.floor(diff / (1000 * 60 * 60 * 24));
}

function guestDisplayName(g: GuestProfile, index: number): string {
  return g.first_name ? g.first_name : `Guest #${index + 1}`;
}

// ─── Segment config ──────────────────────────────────────────────────────────

const SEGMENT_COLORS: Record<string, string> = {
  champion: '#22c55e',
  loyal: '#3b82f6',
  potential_loyalist: '#8b5cf6',
  at_risk: '#f59e0b',
  new: '#06b6d4',
  lapsed: '#ef4444',
  regular: '#6b7280',
};

const SEGMENT_LABELS: Record<string, string> = {
  champion: 'Champion',
  loyal: 'Loyal',
  potential_loyalist: 'Potential Loyalist',
  at_risk: 'At Risk',
  new: 'New',
  lapsed: 'Lapsed',
  regular: 'Regular',
};

type BadgeVariant = 'success' | 'info' | 'warning' | 'critical' | 'neutral';

function segmentBadgeVariant(segment: string): BadgeVariant {
  switch (segment) {
    case 'champion': return 'success';
    case 'loyal': return 'info';
    case 'potential_loyalist': return 'info';
    case 'at_risk': return 'warning';
    case 'lapsed': return 'critical';
    case 'new': return 'neutral';
    default: return 'neutral';
  }
}

// ─── Churn config ────────────────────────────────────────────────────────────

const CHURN_COLORS: Record<string, string> = {
  low: '#22c55e',
  medium: '#eab308',
  high: '#f97316',
  critical: '#ef4444',
};

function churnBadgeVariant(risk: string): BadgeVariant {
  switch (risk) {
    case 'low': return 'success';
    case 'medium': return 'warning';
    case 'high': return 'warning';
    case 'critical': return 'critical';
    default: return 'neutral';
  }
}

// ─── Tab definitions ─────────────────────────────────────────────────────────

type Tab = 'guests' | 'analytics' | 'at_risk';

const TABS: { id: Tab; label: string }[] = [
  { id: 'guests', label: 'Guest List' },
  { id: 'analytics', label: 'Analytics' },
  { id: 'at_risk', label: 'At Risk' },
];

// ─── Guest List Tab ──────────────────────────────────────────────────────────

function GuestListTab({ locationId }: { locationId: string }) {
  const { data, isLoading, error, refetch } = useGuests(locationId, 'clv_score');
  const [expandedId, setExpandedId] = useState<string | null>(null);

  const guests = data?.guests ?? [];

  const columns: Column<GuestProfile & { _index: number }>[] = [
    {
      key: 'first_name',
      header: 'Name',
      sortable: true,
      render: (r) => (
        <span className="font-semibold text-white">{guestDisplayName(r, r._index)}</span>
      ),
    },
    {
      key: 'segment',
      header: 'Segment',
      sortable: true,
      render: (r) => (
        <StatusBadge variant={segmentBadgeVariant(r.segment)}>
          {SEGMENT_LABELS[r.segment] ?? r.segment}
        </StatusBadge>
      ),
    },
    {
      key: 'clv_score',
      header: 'CLV',
      align: 'right',
      sortable: true,
      render: (r) => (
        <span className="font-medium text-white">{dollars(r.clv_score)}</span>
      ),
    },
    {
      key: 'total_visits',
      header: 'Visits',
      align: 'right',
      sortable: true,
    },
    {
      key: 'avg_check',
      header: 'Avg Check',
      align: 'right',
      sortable: true,
      render: (r) => dollars(r.avg_check),
    },
    {
      key: 'last_visit_at',
      header: 'Last Visit',
      align: 'right',
      sortable: true,
      render: (r) =>
        r.last_visit_at
          ? new Date(r.last_visit_at).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
          : '—',
    },
    {
      key: 'churn_risk',
      header: 'Churn Risk',
      align: 'right',
      sortable: true,
      render: (r) => (
        <StatusBadge variant={churnBadgeVariant(r.churn_risk)}>
          {r.churn_risk.charAt(0).toUpperCase() + r.churn_risk.slice(1)}
        </StatusBadge>
      ),
    },
  ];

  if (error) {
    const msg = error instanceof Error ? error.message : 'Failed to load guests';
    return <ErrorBanner message={msg} retry={() => void refetch()} />;
  }

  // Attach _index for display name resolution
  const indexedGuests = guests.map((g, i) => ({ ...g, _index: i }));

  return (
    <div className="space-y-4">
      <DataTable
        columns={columns}
        data={indexedGuests}
        keyExtractor={(r) => r.guest_id}
        isLoading={isLoading}
        emptyTitle="No guests found"
        emptyDescription="No guest data is available for this location."
        onRowClick={(r) => setExpandedId(expandedId === r.guest_id ? null : r.guest_id)}
      />
      {/* Expandable detail placeholder */}
      {expandedId && (() => {
        const guest = indexedGuests.find((g) => g.guest_id === expandedId);
        if (!guest) return null;
        return (
          <div className="bg-white/5 rounded-xl border border-white/10 shadow-sm p-6 space-y-3">
            <div className="flex items-center justify-between">
              <h3 className="text-base font-semibold text-white">
                {guestDisplayName(guest, guest._index)} — Profile Detail
              </h3>
              <button
                onClick={() => setExpandedId(null)}
                className="text-slate-300 hover:text-slate-300 transition-colors"
              >
                <ChevronUp className="h-5 w-5" />
              </button>
            </div>
            <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 text-sm">
              <div>
                <p className="text-slate-400">CLV Score</p>
                <p className="font-semibold text-white">{dollars(guest.clv_score)}</p>
              </div>
              <div>
                <p className="text-slate-400">Total Spend</p>
                <p className="font-semibold text-white">{dollars(guest.total_spend)}</p>
              </div>
              <div>
                <p className="text-slate-400">Churn Probability</p>
                <p className="font-semibold text-white">{(guest.churn_probability * 100).toFixed(1)}%</p>
              </div>
              <div>
                <p className="text-slate-400">Privacy Tier</p>
                <p className="font-semibold text-white capitalize">{guest.privacy_tier}</p>
              </div>
            </div>
            <p className="text-xs text-slate-300 italic">Visit history detail coming soon.</p>
          </div>
        );
      })()}
    </div>
  );
}

// ─── Analytics Tab ───────────────────────────────────────────────────────────

function AnalyticsTab({ locationId }: { locationId: string }) {
  const { data: guestsData, isLoading: guestsLoading } = useGuests(locationId, 'clv_score');
  const { data: segData, isLoading: segLoading } = useSegments();
  const { data: churnData, isLoading: churnLoading } = useChurnDist();
  const { data: clvData, isLoading: clvLoading } = useCLVDist();
  const refresh = useRefreshAnalytics();

  const guests = guestsData?.guests ?? [];
  const segments = segData?.segments ?? [];
  const churnDist = churnData?.distribution ?? [];
  const clvBuckets = clvData?.buckets ?? [];

  const totalGuests = guests.length;
  const avgCLV = totalGuests > 0 ? guests.reduce((s, g) => s + g.clv_score, 0) / totalGuests : 0;
  const avgVisits = totalGuests > 0 ? guests.reduce((s, g) => s + g.total_visits, 0) / totalGuests : 0;
  const avgCheck = totalGuests > 0 ? guests.reduce((s, g) => s + g.avg_check, 0) / totalGuests : 0;

  const totalChurn = churnDist.reduce((s, d) => s + d.count, 0);

  const CHURN_ORDER = ['low', 'medium', 'high', 'critical'];
  const orderedChurn = CHURN_ORDER.map((risk) => {
    const entry = churnDist.find((d) => d.risk === risk);
    return { risk, count: entry?.count ?? 0 };
  });

  const pieData = segments.map((s) => ({
    name: SEGMENT_LABELS[s.segment] ?? s.segment,
    value: s.count,
    color: SEGMENT_COLORS[s.segment] ?? '#6b7280',
  }));

  const anyLoading = guestsLoading || segLoading || churnLoading || clvLoading;

  return (
    <div className="space-y-6">
      {/* Refresh button */}
      <div className="flex justify-end">
        <button
          onClick={() => refresh.mutate()}
          disabled={refresh.isPending}
          className="flex items-center gap-2 px-4 py-2 rounded-md text-sm font-semibold text-white bg-[#F97316] hover:bg-orange-600 disabled:opacity-60 disabled:cursor-not-allowed transition-colors"
        >
          {refresh.isPending ? <LoadingSpinner size="sm" /> : <RefreshCw className="h-4 w-4" />}
          Refresh Analytics
        </button>
      </div>

      {/* KPI Cards */}
      {anyLoading ? (
        <div className="flex justify-center py-8"><LoadingSpinner /></div>
      ) : (
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
          <KPICard
            label="Total Guests"
            value={String(totalGuests)}
            icon={UserCheck}
            iconColor="text-slate-300"
            bgTint="bg-white/10"
          />
          <KPICard
            label="Avg CLV"
            value={dollars(avgCLV)}
            icon={DollarSign}
            iconColor="text-blue-400"
            bgTint="bg-blue-500/10"
          />
          <KPICard
            label="Avg Visits"
            value={avgVisits.toFixed(1)}
            icon={TrendingUp}
            iconColor="text-emerald-400"
            bgTint="bg-emerald-500/10"
          />
          <KPICard
            label="Avg Check"
            value={dollars(avgCheck)}
            icon={ShoppingCart}
            iconColor="text-purple-400"
            bgTint="bg-purple-500/10"
          />
        </div>
      )}

      {/* Charts row */}
      {!anyLoading && (
        <>
          {/* Segment pie + CLV bar */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {/* Segment distribution pie */}
            <div className="bg-white/5 rounded-xl border border-white/10 shadow-sm p-6">
              <h3 className="text-base font-semibold text-white mb-4">Segment Distribution</h3>
              {pieData.length === 0 ? (
                <p className="text-sm text-slate-300 text-center py-8">No segment data</p>
              ) : (
                <div className="flex items-center gap-4">
                  <ResponsiveContainer width="60%" height={200}>
                    <PieChart>
                      <Pie
                        data={pieData}
                        dataKey="value"
                        nameKey="name"
                        cx="50%"
                        cy="50%"
                        outerRadius={80}
                      >
                        {pieData.map((entry, i) => (
                          <Cell key={i} fill={entry.color} />
                        ))}
                      </Pie>
                      <Tooltip formatter={(v) => [v, 'Guests']} />
                    </PieChart>
                  </ResponsiveContainer>
                  <div className="flex-1 space-y-1.5">
                    {pieData.map((entry, i) => (
                      <div key={i} className="flex items-center gap-2 text-sm">
                        <span
                          className="inline-block w-3 h-3 rounded-full flex-shrink-0"
                          style={{ backgroundColor: entry.color }}
                        />
                        <span className="text-slate-300 truncate">{entry.name}</span>
                        <span className="ml-auto font-semibold text-white">{entry.value}</span>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>

            {/* CLV distribution bar */}
            <div className="bg-white/5 rounded-xl border border-white/10 shadow-sm p-6">
              <h3 className="text-base font-semibold text-white mb-4">CLV Distribution</h3>
              {clvBuckets.length === 0 ? (
                <p className="text-sm text-slate-300 text-center py-8">No CLV data</p>
              ) : (
                <ResponsiveContainer width="100%" height={200}>
                  <BarChart data={clvBuckets} margin={{ top: 0, right: 8, left: 0, bottom: 0 }}>
                    <XAxis dataKey="range" tick={{ fontSize: 11, fill: '#94a3b8' }} />
                    <YAxis tick={{ fontSize: 11, fill: '#94a3b8' }} allowDecimals={false} />
                    <Tooltip contentStyle={{ backgroundColor: '#1e293b', border: '1px solid rgba(255,255,255,0.1)', color: '#e2e8f0' }} />
                    <Bar dataKey="count" name="Guests" fill="#3b82f6" radius={[4, 4, 0, 0]} />
                  </BarChart>
                </ResponsiveContainer>
              )}
            </div>
          </div>

          {/* Churn risk breakdown */}
          <div className="bg-white/5 rounded-xl border border-white/10 shadow-sm p-6">
            <h3 className="text-base font-semibold text-white mb-4">Churn Risk Breakdown</h3>
            {churnDist.length === 0 ? (
              <p className="text-sm text-slate-300">No churn data</p>
            ) : (
              <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
                {orderedChurn.map(({ risk, count }) => {
                  const pct = totalChurn > 0 ? ((count / totalChurn) * 100).toFixed(1) : '0.0';
                  const color = CHURN_COLORS[risk] ?? '#6b7280';
                  const label = risk.charAt(0).toUpperCase() + risk.slice(1);
                  return (
                    <div
                      key={risk}
                      className="rounded-lg border p-4 text-center"
                      style={{ borderColor: color + '40', backgroundColor: color + '0d' }}
                    >
                      <p className="text-sm font-medium" style={{ color }}>{label}</p>
                      <p className="text-2xl font-bold text-white mt-1">{count}</p>
                      <p className="text-xs text-slate-400 mt-0.5">{pct}% of total</p>
                    </div>
                  );
                })}
              </div>
            )}
          </div>
        </>
      )}
    </div>
  );
}

// ─── At Risk Tab ─────────────────────────────────────────────────────────────

function AtRiskTab({ locationId }: { locationId: string }) {
  const { data, isLoading, error, refetch } = useGuests(locationId, 'clv_score');

  const allGuests = data?.guests ?? [];
  const atRisk = allGuests
    .filter((g) => g.churn_risk === 'high' || g.churn_risk === 'critical')
    .sort((a, b) => b.clv_score - a.clv_score);

  // Map to indexed form for name rendering
  const guestIndexMap = new Map(allGuests.map((g, i) => [g.guest_id, i]));

  const columns: Column<GuestProfile>[] = [
    {
      key: 'first_name',
      header: 'Name',
      sortable: true,
      render: (r) => (
        <span className="font-semibold text-white">
          {guestDisplayName(r, guestIndexMap.get(r.guest_id) ?? 0)}
        </span>
      ),
    },
    {
      key: 'clv_score',
      header: 'CLV',
      align: 'right',
      sortable: true,
      render: (r) => <span className="font-medium text-white">{dollars(r.clv_score)}</span>,
    },
    {
      key: 'segment',
      header: 'Segment',
      sortable: true,
      render: (r) => (
        <StatusBadge variant={segmentBadgeVariant(r.segment)}>
          {SEGMENT_LABELS[r.segment] ?? r.segment}
        </StatusBadge>
      ),
    },
    {
      key: 'last_visit_at',
      header: 'Days Since Visit',
      align: 'right',
      sortable: true,
      render: (r) => {
        const d = daysSince(r.last_visit_at);
        return d !== null ? (
          <span className={d > 60 ? 'text-red-600 font-semibold' : 'text-slate-200'}>{d}d</span>
        ) : '—';
      },
    },
    {
      key: 'churn_probability',
      header: 'Churn %',
      align: 'right',
      sortable: true,
      render: (r) => (
        <span className="font-semibold text-red-600">
          {(r.churn_probability * 100).toFixed(1)}%
        </span>
      ),
    },
    {
      key: 'churn_risk',
      header: 'Risk',
      align: 'right',
      render: (r) => (
        <StatusBadge variant={churnBadgeVariant(r.churn_risk)}>
          {r.churn_risk.charAt(0).toUpperCase() + r.churn_risk.slice(1)}
        </StatusBadge>
      ),
    },
  ];

  if (error) {
    const msg = error instanceof Error ? error.message : 'Failed to load guests';
    return <ErrorBanner message={msg} retry={() => void refetch()} />;
  }

  return (
    <div className="space-y-4">
      {/* Red banner */}
      {!isLoading && atRisk.length > 0 && (
        <div className="flex items-center gap-3 bg-red-500/10 border border-red-500/30 text-red-400 rounded-lg px-4 py-3">
          <AlertTriangle className="h-5 w-5 flex-shrink-0" />
          <span className="text-sm font-semibold">
            {atRisk.length} high-value guest{atRisk.length !== 1 ? 's' : ''} at risk of churning
          </span>
        </div>
      )}

      <DataTable
        columns={columns}
        data={atRisk}
        keyExtractor={(r) => r.guest_id}
        isLoading={isLoading}
        emptyTitle="No at-risk guests"
        emptyDescription="No guests are currently flagged as high or critical churn risk."
      />
    </div>
  );
}

// ─── Main Page ────────────────────────────────────────────────────────────────

export default function CustomerPage() {
  const locationId = useLocationStore((s) => s.selectedLocationId);
  const [activeTab, setActiveTab] = useState<Tab>('guests');

  if (!locationId) return <LoadingSpinner fullPage />;

  return (
    <div className="space-y-6">
      {/* Page header */}
      <div>
        <h1 className="text-2xl font-bold text-white">Customer Intelligence</h1>
        <p className="text-sm text-slate-400 mt-1">
          Guest profiles, CLV analytics, segmentation, and churn risk management
        </p>
      </div>

      {/* Tab bar */}
      <div className="flex gap-1 border-b border-white/10">
        {TABS.map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            className={`px-4 py-2.5 text-sm font-medium border-b-2 transition-colors ${
              activeTab === tab.id
                ? 'border-[#F97316] text-[#F97316]'
                : 'border-transparent text-slate-400 hover:text-slate-200 hover:border-white/15'
            }`}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {/* Tab content */}
      <div>
        {activeTab === 'guests' && <GuestListTab locationId={locationId} />}
        {activeTab === 'analytics' && <AnalyticsTab locationId={locationId} />}
        {activeTab === 'at_risk' && <AtRiskTab locationId={locationId} />}
      </div>
    </div>
  );
}
