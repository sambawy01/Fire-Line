import { useState } from 'react';
import {
  Megaphone,
  Plus,
  Play,
  Pause,
  Users,
  DollarSign,
  Tag,
  TrendingUp,
  Star,
  BarChart2,
  Award,
} from 'lucide-react';
import {
  PieChart,
  Pie,
  Cell,
  Tooltip,
  ResponsiveContainer,
} from 'recharts';
import { useLocationStore } from '../stores/location';
import {
  useCampaigns,
  useCampaignMetrics,
  useLoyaltyMembers,
  useLoyaltyMetrics,
  useActivateCampaign,
  usePauseCampaign,
  useCreateCampaign,
} from '../hooks/useMarketing';
import KPICard from '../components/ui/KPICard';
import DataTable from '../components/ui/DataTable';
import type { Column } from '../components/ui/DataTable';
import StatusBadge from '../components/ui/StatusBadge';
import LoadingSpinner from '../components/ui/LoadingSpinner';
import ErrorBanner from '../components/ui/ErrorBanner';
import Modal from '../components/ui/Modal';
import type { Campaign, LoyaltyMember } from '../lib/api';

// ─── Helpers ──────────────────────────────────────────────────────────────────

function dollars(cents: number): string {
  return `$${(cents / 100).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
}

function fmtDate(iso: string): string {
  return new Date(iso).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
}

// ─── Badge helpers ────────────────────────────────────────────────────────────

type BadgeVariant = 'success' | 'info' | 'warning' | 'critical' | 'neutral';

function statusVariant(status: Campaign['status']): BadgeVariant {
  switch (status) {
    case 'active': return 'success';
    case 'scheduled': return 'info';
    case 'paused': return 'warning';
    case 'completed': return 'neutral';
    case 'cancelled': return 'critical';
    default: return 'neutral'; // draft
  }
}

const TYPE_COLORS: Record<Campaign['type'], string> = {
  discount: 'bg-emerald-100 text-emerald-700',
  bogo: 'bg-violet-100 text-violet-700',
  happy_hour: 'bg-amber-100 text-amber-700',
  bundle: 'bg-blue-100 text-blue-700',
  loyalty_reward: 'bg-pink-100 text-pink-700',
  custom: 'bg-gray-100 text-gray-700',
};

const TYPE_LABELS: Record<Campaign['type'], string> = {
  discount: 'Discount',
  bogo: 'BOGO',
  happy_hour: 'Happy Hour',
  bundle: 'Bundle',
  loyalty_reward: 'Loyalty Reward',
  custom: 'Custom',
};

function TypeBadge({ type }: { type: Campaign['type'] }) {
  return (
    <span className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-semibold ${TYPE_COLORS[type] ?? 'bg-gray-100 text-gray-700'}`}>
      {TYPE_LABELS[type] ?? type}
    </span>
  );
}

const TIER_COLORS: Record<string, string> = {
  bronze: '#cd7f32',
  silver: '#c0c0c0',
  gold: '#ffd700',
  platinum: '#e5e4e2',
};

function tierVariant(tier: string): BadgeVariant {
  switch (tier) {
    case 'gold': return 'warning';
    case 'platinum': return 'neutral';
    case 'silver': return 'info';
    default: return 'neutral'; // bronze
  }
}

// ─── Tab types ────────────────────────────────────────────────────────────────

type Tab = 'campaigns' | 'loyalty' | 'analytics';

const TABS: { id: Tab; label: string }[] = [
  { id: 'campaigns', label: 'Campaigns' },
  { id: 'loyalty', label: 'Loyalty' },
  { id: 'analytics', label: 'Analytics' },
];

// ─── Create Campaign Modal ────────────────────────────────────────────────────

interface CreateCampaignModalProps {
  open: boolean;
  onClose: () => void;
  locationId: string;
}

const CAMPAIGN_TYPES: Campaign['type'][] = [
  'discount', 'bogo', 'happy_hour', 'bundle', 'loyalty_reward', 'custom',
];

const CHANNELS = ['email', 'sms', 'push', 'in_app', 'social', 'all'];

function CreateCampaignModal({ open, onClose, locationId }: CreateCampaignModalProps) {
  const createCampaign = useCreateCampaign();
  const [form, setForm] = useState({
    name: '',
    type: 'discount' as Campaign['type'],
    target_segment: '',
    channel: 'email',
    discount_type: 'percentage',
    discount_value: '',
    start_date: '',
    end_date: '',
  });

  function handleChange(field: string, value: string) {
    setForm((prev) => ({ ...prev, [field]: value }));
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    await createCampaign.mutateAsync({
      ...form,
      location_id: locationId,
      discount_value: Number(form.discount_value),
    });
    onClose();
    setForm({
      name: '',
      type: 'discount',
      target_segment: '',
      channel: 'email',
      discount_type: 'percentage',
      discount_value: '',
      start_date: '',
      end_date: '',
    });
  }

  const inputClass =
    'w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-1 focus:ring-[#F97316]';
  const labelClass = 'block text-xs font-medium text-gray-600 mb-1';

  return (
    <Modal
      open={open}
      onClose={onClose}
      title="Create Campaign"
      footer={
        <>
          <button
            type="button"
            onClick={onClose}
            className="px-4 py-2 text-sm font-medium text-gray-600 border border-gray-300 rounded-md hover:bg-gray-50"
          >
            Cancel
          </button>
          <button
            form="create-campaign-form"
            type="submit"
            disabled={createCampaign.isPending}
            className="px-4 py-2 text-sm font-semibold text-white bg-[#F97316] rounded-md hover:bg-orange-600 disabled:opacity-60 disabled:cursor-not-allowed"
          >
            {createCampaign.isPending ? 'Creating…' : 'Create Campaign'}
          </button>
        </>
      }
    >
      <form id="create-campaign-form" onSubmit={(e) => void handleSubmit(e)} className="space-y-4">
        <div>
          <label className={labelClass}>Campaign Name</label>
          <input
            required
            className={inputClass}
            placeholder="Summer Discount"
            value={form.name}
            onChange={(e) => handleChange('name', e.target.value)}
          />
        </div>
        <div className="grid grid-cols-2 gap-3">
          <div>
            <label className={labelClass}>Type</label>
            <select
              className={inputClass}
              value={form.type}
              onChange={(e) => handleChange('type', e.target.value)}
            >
              {CAMPAIGN_TYPES.map((t) => (
                <option key={t} value={t}>{TYPE_LABELS[t]}</option>
              ))}
            </select>
          </div>
          <div>
            <label className={labelClass}>Channel</label>
            <select
              className={inputClass}
              value={form.channel}
              onChange={(e) => handleChange('channel', e.target.value)}
            >
              {CHANNELS.map((c) => (
                <option key={c} value={c}>{c.charAt(0).toUpperCase() + c.slice(1)}</option>
              ))}
            </select>
          </div>
        </div>
        <div>
          <label className={labelClass}>Target Segment</label>
          <input
            className={inputClass}
            placeholder="e.g. vip, lapsed, all"
            value={form.target_segment}
            onChange={(e) => handleChange('target_segment', e.target.value)}
          />
        </div>
        <div className="grid grid-cols-2 gap-3">
          <div>
            <label className={labelClass}>Discount Type</label>
            <select
              className={inputClass}
              value={form.discount_type}
              onChange={(e) => handleChange('discount_type', e.target.value)}
            >
              <option value="percentage">Percentage (%)</option>
              <option value="fixed">Fixed ($)</option>
              <option value="bogo">BOGO</option>
              <option value="none">None</option>
            </select>
          </div>
          <div>
            <label className={labelClass}>Discount Value</label>
            <input
              type="number"
              min="0"
              step="0.01"
              className={inputClass}
              placeholder="10"
              value={form.discount_value}
              onChange={(e) => handleChange('discount_value', e.target.value)}
            />
          </div>
        </div>
        <div className="grid grid-cols-2 gap-3">
          <div>
            <label className={labelClass}>Start Date</label>
            <input
              required
              type="date"
              className={inputClass}
              value={form.start_date}
              onChange={(e) => handleChange('start_date', e.target.value)}
            />
          </div>
          <div>
            <label className={labelClass}>End Date</label>
            <input
              required
              type="date"
              className={inputClass}
              value={form.end_date}
              onChange={(e) => handleChange('end_date', e.target.value)}
            />
          </div>
        </div>
        {createCampaign.error && (
          <p className="text-sm text-red-600">
            {createCampaign.error instanceof Error ? createCampaign.error.message : 'Failed to create campaign'}
          </p>
        )}
      </form>
    </Modal>
  );
}

// ─── Campaign Card ────────────────────────────────────────────────────────────

function CampaignCard({ campaign }: { campaign: Campaign }) {
  const activate = useActivateCampaign();
  const pause = usePauseCampaign();

  const canActivate = campaign.status === 'draft' || campaign.status === 'paused' || campaign.status === 'scheduled';
  const canPause = campaign.status === 'active';

  return (
    <div className="bg-white rounded-xl border border-gray-200 shadow-sm p-5 space-y-3">
      <div className="flex items-start justify-between gap-2">
        <div className="space-y-1 min-w-0">
          <h3 className="text-base font-semibold text-gray-800 truncate">{campaign.name}</h3>
          <div className="flex items-center gap-2 flex-wrap">
            <TypeBadge type={campaign.type} />
            <StatusBadge variant={statusVariant(campaign.status)}>
              {campaign.status.charAt(0).toUpperCase() + campaign.status.slice(1)}
            </StatusBadge>
          </div>
        </div>
        <div className="flex items-center gap-2 shrink-0">
          {canActivate && (
            <button
              onClick={() => activate.mutate(campaign.campaign_id)}
              disabled={activate.isPending}
              title="Activate"
              className="flex items-center gap-1 px-3 py-1.5 text-xs font-semibold text-white bg-emerald-500 hover:bg-emerald-600 disabled:opacity-60 rounded-md transition-colors"
            >
              <Play className="h-3.5 w-3.5" />
              Activate
            </button>
          )}
          {canPause && (
            <button
              onClick={() => pause.mutate(campaign.campaign_id)}
              disabled={pause.isPending}
              title="Pause"
              className="flex items-center gap-1 px-3 py-1.5 text-xs font-semibold text-white bg-amber-500 hover:bg-amber-600 disabled:opacity-60 rounded-md transition-colors"
            >
              <Pause className="h-3.5 w-3.5" />
              Pause
            </button>
          )}
        </div>
      </div>

      <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 text-sm">
        <div>
          <p className="text-xs text-gray-500">Target Segment</p>
          <p className="font-medium text-gray-800 capitalize">{campaign.target_segment || '—'}</p>
        </div>
        <div>
          <p className="text-xs text-gray-500">Channel</p>
          <p className="font-medium text-gray-800 capitalize">{campaign.channel || '—'}</p>
        </div>
        <div>
          <p className="text-xs text-gray-500">Redemptions</p>
          <p className="font-medium text-gray-800">{campaign.redemptions?.toLocaleString() ?? 0}</p>
        </div>
        <div>
          <p className="text-xs text-gray-500">Revenue Attributed</p>
          <p className="font-medium text-gray-800">{dollars(campaign.revenue_attributed ?? 0)}</p>
        </div>
      </div>

      <div className="flex items-center gap-1 text-xs text-gray-400">
        <span>{fmtDate(campaign.start_date)}</span>
        <span>→</span>
        <span>{fmtDate(campaign.end_date)}</span>
      </div>
    </div>
  );
}

// ─── Campaigns Tab ────────────────────────────────────────────────────────────

function CampaignsTab({ locationId }: { locationId: string }) {
  const [showCreate, setShowCreate] = useState(false);
  const [statusFilter, setStatusFilter] = useState('');
  const { data, isLoading, error, refetch } = useCampaigns(locationId, statusFilter || undefined);

  const campaigns = data?.campaigns ?? [];

  if (error) {
    const msg = error instanceof Error ? error.message : 'Failed to load campaigns';
    return <ErrorBanner message={msg} retry={() => void refetch()} />;
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between gap-3 flex-wrap">
        <div className="flex items-center gap-2">
          <label className="text-sm text-gray-600">Status:</label>
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className="border border-gray-300 rounded-md px-3 py-1.5 text-sm focus:outline-none focus:ring-1 focus:ring-[#F97316]"
          >
            <option value="">All</option>
            <option value="draft">Draft</option>
            <option value="scheduled">Scheduled</option>
            <option value="active">Active</option>
            <option value="paused">Paused</option>
            <option value="completed">Completed</option>
            <option value="cancelled">Cancelled</option>
          </select>
        </div>
        <button
          onClick={() => setShowCreate(true)}
          className="flex items-center gap-2 px-4 py-2 text-sm font-semibold text-white bg-[#F97316] hover:bg-orange-600 rounded-md transition-colors"
        >
          <Plus className="h-4 w-4" />
          Create Campaign
        </button>
      </div>

      {isLoading ? (
        <div className="flex justify-center py-12"><LoadingSpinner size="lg" /></div>
      ) : campaigns.length === 0 ? (
        <div className="bg-white rounded-xl border border-gray-200 shadow-sm p-12 text-center">
          <Megaphone className="h-10 w-10 text-gray-300 mx-auto mb-3" />
          <p className="text-gray-500 font-medium">No campaigns found</p>
          <p className="text-sm text-gray-400 mt-1">Create your first campaign to get started.</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
          {campaigns.map((c) => (
            <CampaignCard key={c.campaign_id} campaign={c} />
          ))}
        </div>
      )}

      <CreateCampaignModal
        open={showCreate}
        onClose={() => setShowCreate(false)}
        locationId={locationId}
      />
    </div>
  );
}

// ─── Loyalty Tab ──────────────────────────────────────────────────────────────

function LoyaltyTab() {
  const [tierFilter, setTierFilter] = useState('');
  const { data: metricsData, isLoading: metricsLoading } = useLoyaltyMetrics();
  const { data: membersData, isLoading: membersLoading, error, refetch } = useLoyaltyMembers(tierFilter || undefined);

  const metrics = metricsData;
  const members = membersData?.members ?? [];

  const tierCards = [
    { key: 'bronze_count', label: 'Bronze', color: '#cd7f32', count: metrics?.bronze_count ?? 0 },
    { key: 'silver_count', label: 'Silver', color: '#c0c0c0', count: metrics?.silver_count ?? 0 },
    { key: 'gold_count', label: 'Gold', color: '#ffd700', count: metrics?.gold_count ?? 0 },
    { key: 'platinum_count', label: 'Platinum', color: '#e5e4e2', count: metrics?.platinum_count ?? 0 },
  ];

  const columns: Column<LoyaltyMember>[] = [
    {
      key: 'guest_name',
      header: 'Name',
      sortable: true,
      render: (r) => <span className="font-semibold text-gray-800">{r.guest_name || '—'}</span>,
    },
    {
      key: 'points_balance',
      header: 'Points Balance',
      align: 'right',
      sortable: true,
      render: (r) => <span className="font-medium text-gray-800">{(r.points_balance ?? 0).toLocaleString()}</span>,
    },
    {
      key: 'lifetime_points',
      header: 'Lifetime Points',
      align: 'right',
      sortable: true,
      render: (r) => (r.lifetime_points ?? 0).toLocaleString(),
    },
    {
      key: 'tier',
      header: 'Tier',
      sortable: true,
      render: (r) => (
        <StatusBadge variant={tierVariant(r.tier ?? '')}>
          <span style={{ color: TIER_COLORS[r.tier ?? ''] ?? undefined }}>
            {(r.tier ?? '').charAt(0).toUpperCase() + (r.tier ?? '').slice(1)}
          </span>
        </StatusBadge>
      ),
    },
    {
      key: 'joined_at',
      header: 'Joined',
      align: 'right',
      sortable: true,
      render: (r) => fmtDate(r.joined_at),
    },
  ];

  if (error) {
    const msg = error instanceof Error ? error.message : 'Failed to load loyalty members';
    return <ErrorBanner message={msg} retry={() => void refetch()} />;
  }

  return (
    <div className="space-y-6">
      {/* KPI cards */}
      {metricsLoading ? (
        <div className="flex justify-center py-8"><LoadingSpinner /></div>
      ) : (
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
          <KPICard
            label="Total Members"
            value={String(metrics?.total_members ?? 0)}
            icon={Users}
            iconColor="text-blue-600"
            bgTint="bg-blue-50"
          />
          <KPICard
            label="Avg Points Balance"
            value={(metrics?.avg_balance ?? 0).toLocaleString(undefined, { maximumFractionDigits: 0 })}
            icon={Star}
            iconColor="text-amber-500"
            bgTint="bg-amber-50"
          />
          <KPICard
            label="Total Issued"
            value={(metrics?.total_issued ?? 0).toLocaleString()}
            icon={Award}
            iconColor="text-emerald-600"
            bgTint="bg-emerald-50"
          />
          <KPICard
            label="Total Redeemed"
            value={(metrics?.total_redeemed ?? 0).toLocaleString()}
            icon={Tag}
            iconColor="text-purple-600"
            bgTint="bg-purple-50"
          />
        </div>
      )}

      {/* Tier breakdown */}
      {!metricsLoading && (
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
          {tierCards.map(({ key, label, color, count }) => (
            <div
              key={key}
              className="rounded-xl border p-5 text-center cursor-pointer transition-all hover:shadow-md"
              style={{ borderColor: color + '60', backgroundColor: color + '18' }}
              onClick={() => setTierFilter(tierFilter === label.toLowerCase() ? '' : label.toLowerCase())}
            >
              <p className="text-sm font-semibold" style={{ color }}>{label}</p>
              <p className="text-3xl font-bold text-gray-800 mt-1">{count}</p>
              <p className="text-xs text-gray-500 mt-0.5">members</p>
            </div>
          ))}
        </div>
      )}

      {/* Filter indicator */}
      {tierFilter && (
        <div className="flex items-center gap-2 text-sm text-gray-600">
          <span>Filtered by tier:</span>
          <span
            className="font-semibold capitalize"
            style={{ color: TIER_COLORS[tierFilter] ?? undefined }}
          >
            {tierFilter}
          </span>
          <button
            onClick={() => setTierFilter('')}
            className="text-xs underline text-gray-400 hover:text-gray-600"
          >
            clear
          </button>
        </div>
      )}

      {/* Member list */}
      <DataTable
        columns={columns}
        data={members}
        keyExtractor={(r) => r.member_id}
        isLoading={membersLoading}
        emptyTitle="No loyalty members found"
        emptyDescription="No members match the current filter."
      />
    </div>
  );
}

// ─── Analytics Tab ────────────────────────────────────────────────────────────

const CAMPAIGN_TYPE_COLORS: Record<string, string> = {
  discount: '#10b981',
  bogo: '#8b5cf6',
  happy_hour: '#f59e0b',
  bundle: '#3b82f6',
  loyalty_reward: '#ec4899',
  custom: '#6b7280',
};

const TIER_PIE_COLORS: Record<string, string> = {
  bronze: '#cd7f32',
  silver: '#c0c0c0',
  gold: '#ffd700',
  platinum: '#e5e4e2',
};

function AnalyticsTab({ locationId }: { locationId: string }) {
  const { data: campaignMetrics, isLoading: cmLoading } = useCampaignMetrics();
  const { data: loyaltyMetrics, isLoading: lmLoading } = useLoyaltyMetrics();
  const { data: campaignsData, isLoading: cLoading } = useCampaigns(locationId);

  const campaigns = campaignsData?.campaigns ?? [];

  // Campaign type distribution
  const typeCount: Record<string, number> = {};
  for (const c of campaigns) {
    typeCount[c.type] = (typeCount[c.type] ?? 0) + 1;
  }
  const typePieData = Object.entries(typeCount).map(([type, count]) => ({
    name: TYPE_LABELS[type as Campaign['type']] ?? type,
    value: count,
    color: CAMPAIGN_TYPE_COLORS[type] ?? '#6b7280',
  }));

  // Loyalty tier distribution
  const lm = loyaltyMetrics;
  const tierPieData = lm
    ? [
        { name: 'Bronze', value: lm.bronze_count ?? 0, color: TIER_PIE_COLORS.bronze },
        { name: 'Silver', value: lm.silver_count ?? 0, color: TIER_PIE_COLORS.silver },
        { name: 'Gold', value: lm.gold_count ?? 0, color: TIER_PIE_COLORS.gold },
        { name: 'Platinum', value: lm.platinum_count ?? 0, color: TIER_PIE_COLORS.platinum },
      ].filter((d) => d.value > 0)
    : [];

  const anyLoading = cmLoading || lmLoading || cLoading;

  return (
    <div className="space-y-6">
      {/* Campaign KPIs */}
      {cmLoading ? (
        <div className="flex justify-center py-8"><LoadingSpinner /></div>
      ) : (
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
          <KPICard
            label="Active Campaigns"
            value={String(campaignMetrics?.active_campaigns ?? 0)}
            icon={Megaphone}
            iconColor="text-orange-500"
            bgTint="bg-orange-50"
          />
          <KPICard
            label="Total Redemptions"
            value={(campaignMetrics?.total_redemptions ?? 0).toLocaleString()}
            icon={Tag}
            iconColor="text-blue-600"
            bgTint="bg-blue-50"
          />
          <KPICard
            label="Revenue Attributed"
            value={dollars(campaignMetrics?.revenue_attributed ?? 0)}
            icon={DollarSign}
            iconColor="text-emerald-600"
            bgTint="bg-emerald-50"
          />
          <KPICard
            label="Avg Redemption Rate"
            value={`${((campaignMetrics?.avg_redemption_rate ?? 0) * 100).toFixed(1)}%`}
            icon={TrendingUp}
            iconColor="text-purple-600"
            bgTint="bg-purple-50"
          />
        </div>
      )}

      {/* Charts */}
      {!anyLoading && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Campaign type distribution */}
          <div className="bg-white rounded-xl border border-gray-200 shadow-sm p-6">
            <h3 className="text-base font-semibold text-gray-800 mb-4 flex items-center gap-2">
              <BarChart2 className="h-5 w-5 text-gray-400" />
              Campaign Type Distribution
            </h3>
            {typePieData.length === 0 ? (
              <p className="text-sm text-gray-400 text-center py-8">No campaign data</p>
            ) : (
              <div className="flex items-center gap-4">
                <ResponsiveContainer width="55%" height={200}>
                  <PieChart>
                    <Pie
                      data={typePieData}
                      dataKey="value"
                      nameKey="name"
                      cx="50%"
                      cy="50%"
                      outerRadius={80}
                    >
                      {typePieData.map((entry, i) => (
                        <Cell key={i} fill={entry.color} />
                      ))}
                    </Pie>
                    <Tooltip formatter={(v) => [v, 'Campaigns']} />
                  </PieChart>
                </ResponsiveContainer>
                <div className="flex-1 space-y-1.5">
                  {typePieData.map((entry, i) => (
                    <div key={i} className="flex items-center gap-2 text-sm">
                      <span
                        className="inline-block w-3 h-3 rounded-full shrink-0"
                        style={{ backgroundColor: entry.color }}
                      />
                      <span className="text-gray-600 truncate">{entry.name}</span>
                      <span className="ml-auto font-semibold text-gray-800">{entry.value}</span>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>

          {/* Loyalty tier distribution */}
          <div className="bg-white rounded-xl border border-gray-200 shadow-sm p-6">
            <h3 className="text-base font-semibold text-gray-800 mb-4 flex items-center gap-2">
              <Award className="h-5 w-5 text-gray-400" />
              Loyalty Tier Distribution
            </h3>
            {lmLoading ? (
              <div className="flex justify-center py-8"><LoadingSpinner /></div>
            ) : tierPieData.length === 0 ? (
              <p className="text-sm text-gray-400 text-center py-8">No loyalty data</p>
            ) : (
              <div className="flex items-center gap-4">
                <ResponsiveContainer width="55%" height={200}>
                  <PieChart>
                    <Pie
                      data={tierPieData}
                      dataKey="value"
                      nameKey="name"
                      cx="50%"
                      cy="50%"
                      outerRadius={80}
                    >
                      {tierPieData.map((entry, i) => (
                        <Cell key={i} fill={entry.color} />
                      ))}
                    </Pie>
                    <Tooltip formatter={(v) => [v, 'Members']} />
                  </PieChart>
                </ResponsiveContainer>
                <div className="flex-1 space-y-1.5">
                  {tierPieData.map((entry, i) => (
                    <div key={i} className="flex items-center gap-2 text-sm">
                      <span
                        className="inline-block w-3 h-3 rounded-full shrink-0"
                        style={{ backgroundColor: entry.color }}
                      />
                      <span className="text-gray-600">{entry.name}</span>
                      <span className="ml-auto font-semibold text-gray-800">{entry.value}</span>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}

// ─── Main Page ─────────────────────────────────────────────────────────────────

export default function MarketingPage() {
  const locationId = useLocationStore((s) => s.selectedLocationId);
  const [activeTab, setActiveTab] = useState<Tab>('campaigns');

  if (!locationId) return <LoadingSpinner fullPage />;

  return (
    <div className="space-y-6">
      {/* Page header */}
      <div>
        <h1 className="text-2xl font-bold text-gray-800">Marketing</h1>
        <p className="text-sm text-gray-500 mt-1">
          Campaigns, loyalty program management, and marketing analytics
        </p>
      </div>

      {/* Tab bar */}
      <div className="flex gap-1 border-b border-gray-200">
        {TABS.map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            className={`px-4 py-2.5 text-sm font-medium border-b-2 transition-colors ${
              activeTab === tab.id
                ? 'border-[#F97316] text-[#F97316]'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            }`}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {/* Tab content */}
      <div>
        {activeTab === 'campaigns' && <CampaignsTab locationId={locationId} />}
        {activeTab === 'loyalty' && <LoyaltyTab />}
        {activeTab === 'analytics' && <AnalyticsTab locationId={locationId} />}
      </div>
    </div>
  );
}
