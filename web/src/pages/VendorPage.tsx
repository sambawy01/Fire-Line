import { useState } from 'react';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from 'recharts';
import { useLocationStore } from '../stores/location';
import { useVendors, useVendorSummary } from '../hooks/useVendor';
import {
  useVendorScores,
  usePriceAnomalies,
  usePriceTrend,
  useVendorRecommendation,
  useVendorCompare,
  useCalculateScores,
} from '../hooks/useVendorScoring';
import { usePARStatus } from '../hooks/useInventory';
import KPICard from '../components/ui/KPICard';
import StatusBadge from '../components/ui/StatusBadge';
import LoadingSpinner from '../components/ui/LoadingSpinner';
import ErrorBanner from '../components/ui/ErrorBanner';
import { Truck, DollarSign, Star, Package, RefreshCw, AlertTriangle, CheckCircle, TrendingUp } from 'lucide-react';

// ─── helpers ────────────────────────────────────────────────────────────────

function cents(v: number): string {
  return `$${(v / 100).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
}

function scoreColor(score: number): string {
  if (score >= 80) return 'text-green-600';
  if (score >= 60) return 'text-yellow-600';
  return 'text-red-600';
}

function scoreBg(score: number): string {
  if (score >= 80) return 'bg-green-50 border-green-200';
  if (score >= 60) return 'bg-yellow-50 border-yellow-200';
  return 'bg-red-50 border-red-200';
}

function scoreBarColor(score: number): string {
  if (score >= 80) return 'bg-green-500';
  if (score >= 60) return 'bg-yellow-400';
  return 'bg-red-500';
}

function pct(v: number): string {
  return `${(v * 100).toFixed(1)}%`;
}

// ─── Sub-score bar ───────────────────────────────────────────────────────────

function SubScoreBar({ label, value }: { label: string; value: number }) {
  return (
    <div className="space-y-1">
      <div className="flex justify-between text-xs text-slate-400">
        <span>{label}</span>
        <span className={scoreColor(value)}>{value.toFixed(0)}</span>
      </div>
      <div className="h-1.5 w-full bg-white/10 rounded-full overflow-hidden">
        <div
          className={`h-full rounded-full transition-all duration-500 ${scoreBarColor(value)}`}
          style={{ width: `${Math.min(value, 100)}%` }}
        />
      </div>
    </div>
  );
}

// ─── Tabs ────────────────────────────────────────────────────────────────────

type Tab = 'scorecards' | 'price' | 'comparison';

const TABS: { id: Tab; label: string }[] = [
  { id: 'scorecards', label: 'Scorecards' },
  { id: 'price', label: 'Price Intelligence' },
  { id: 'comparison', label: 'Comparison' },
];

// ─── Tab 1: Scorecards ────────────────────────────────────────────────────────

function ScorecardsTab({ locationId }: { locationId: string }) {
  const { data, isLoading, error, refetch } = useVendorScores(locationId);
  const calcMutation = useCalculateScores();

  const scores = data?.vendor_scores ?? [];

  async function handleRecalculate() {
    await calcMutation.mutateAsync(locationId);
    void refetch();
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <p className="text-sm text-slate-400">Reliability scores across all vendors for this location.</p>
        <button
          onClick={() => void handleRecalculate()}
          disabled={calcMutation.isPending}
          className="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-blue-600 text-white text-sm font-medium hover:bg-blue-700 disabled:opacity-60 transition"
        >
          <RefreshCw className={`h-4 w-4 ${calcMutation.isPending ? 'animate-spin' : ''}`} />
          {calcMutation.isPending ? 'Calculating…' : 'Recalculate'}
        </button>
      </div>

      {error && (
        <ErrorBanner
          message={error instanceof Error ? error.message : 'Failed to load scores'}
          retry={() => void refetch()}
        />
      )}

      {isLoading ? (
        <div className="flex justify-center py-12">
          <LoadingSpinner />
        </div>
      ) : scores.length === 0 ? (
        <div className="text-center py-12 text-slate-300 text-sm">
          No scores available. Click Recalculate to generate scores.
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-5">
          {scores.map((vs) => (
            <div
              key={vs.vendor_name}
              className={`bg-white/5 rounded-xl border p-5 space-y-4 ${scoreBg(vs.overall_score)}`}
            >
              {/* Header */}
              <div className="flex items-start justify-between gap-2">
                <h3 className="font-semibold text-white text-base leading-tight">{vs.vendor_name}</h3>
                <span className={`text-3xl font-bold tabular-nums ${scoreColor(vs.overall_score)}`}>
                  {vs.overall_score.toFixed(0)}
                </span>
              </div>

              {/* Sub-score bars */}
              <div className="space-y-2">
                <SubScoreBar label="Price" value={vs.price_score} />
                <SubScoreBar label="Delivery" value={vs.delivery_score} />
                <SubScoreBar label="Quality" value={vs.quality_score} />
                <SubScoreBar label="Accuracy" value={vs.accuracy_score} />
              </div>

              {/* Stats row */}
              <div className="grid grid-cols-3 gap-2 pt-1 border-t border-white/10 text-center">
                <div>
                  <p className="text-xs text-slate-300">OTIF</p>
                  <p className="text-sm font-semibold text-slate-200">{pct(vs.otif_rate)}</p>
                </div>
                <div>
                  <p className="text-xs text-slate-300">Orders</p>
                  <p className="text-sm font-semibold text-slate-200">{vs.total_orders}</p>
                </div>
                <div>
                  <p className="text-xs text-slate-300">Avg Lead</p>
                  <p className="text-sm font-semibold text-slate-200">{vs.avg_lead_days.toFixed(1)}d</p>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

// ─── Tab 2: Price Intelligence ────────────────────────────────────────────────

function PriceIntelligenceTab({ locationId }: { locationId: string }) {
  const [selectedIngredientId, setSelectedIngredientId] = useState<string>('');
  const [selectedVendor, setSelectedVendor] = useState<string>('');

  const { data: inventoryData } = usePARStatus(locationId);
  const ingredients = inventoryData?.par_status ?? [];

  const { data: anomalyData, isLoading: anomalyLoading } = usePriceAnomalies(locationId);
  const anomalies = anomalyData?.anomalies ?? [];

  const { data: trendData, isLoading: trendLoading } = usePriceTrend(
    selectedIngredientId || null,
    selectedVendor || null
  );

  const { data: recommendData } = useVendorRecommendation(
    locationId,
    selectedIngredientId || null
  );

  const { data: scoresData } = useVendorScores(locationId);
  const vendorNames = Array.from(
    new Set((scoresData?.vendor_scores ?? []).map((v) => v.vendor_name))
  );

  const chartData = (trendData?.prices ?? []).map((p) => ({
    date: new Date(p.recorded_at).toLocaleDateString('en-US', { month: 'short', day: 'numeric' }),
    cost: +(p.unit_cost / 100).toFixed(2),
  }));

  return (
    <div className="space-y-8">
      {/* Controls */}
      <div className="flex flex-wrap gap-4 items-end">
        <div>
          <label className="block text-xs font-medium text-slate-400 mb-1">Ingredient</label>
          <select
            value={selectedIngredientId}
            onChange={(e) => setSelectedIngredientId(e.target.value)}
            className="border border-white/10 rounded-lg px-3 py-2 text-sm text-white bg-white focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="">Select ingredient…</option>
            {ingredients.map((ing) => (
              <option key={ing.ingredient_id} value={ing.ingredient_id}>
                {ing.ingredient_name}
              </option>
            ))}
          </select>
        </div>
        <div>
          <label className="block text-xs font-medium text-slate-400 mb-1">Vendor</label>
          <select
            value={selectedVendor}
            onChange={(e) => setSelectedVendor(e.target.value)}
            className="border border-white/10 rounded-lg px-3 py-2 text-sm text-white bg-white focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="">Select vendor…</option>
            {vendorNames.map((v) => (
              <option key={v} value={v}>{v}</option>
            ))}
          </select>
        </div>
      </div>

      {/* Price trend chart */}
      <div className="bg-white/5 rounded-xl border border-white/10 shadow-sm p-5">
        <h3 className="text-sm font-semibold text-slate-200 mb-4">Unit Cost Trend (6 months)</h3>
        {trendLoading ? (
          <div className="flex justify-center py-10"><LoadingSpinner /></div>
        ) : !selectedIngredientId || !selectedVendor ? (
          <div className="flex items-center justify-center h-40 text-slate-300 text-sm">
            Select an ingredient and vendor to view the price trend.
          </div>
        ) : chartData.length === 0 ? (
          <div className="flex items-center justify-center h-40 text-slate-300 text-sm">
            No price history available for this selection.
          </div>
        ) : (
          <ResponsiveContainer width="100%" height={220}>
            <LineChart data={chartData} margin={{ top: 4, right: 16, left: 0, bottom: 4 }}>
              <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
              <XAxis dataKey="date" tick={{ fontSize: 11 }} />
              <YAxis
                tick={{ fontSize: 11 }}
                tickFormatter={(v) => `$${v}`}
                width={52}
              />
              <Tooltip formatter={(v: number) => [`$${v.toFixed(2)}`, 'Unit Cost']} />
              <Line
                type="monotone"
                dataKey="cost"
                stroke="#2563eb"
                strokeWidth={2}
                dot={{ r: 3 }}
                activeDot={{ r: 5 }}
              />
            </LineChart>
          </ResponsiveContainer>
        )}
      </div>

      {/* Vendor recommendation */}
      {recommendData && selectedIngredientId && (
        <div className="bg-blue-50 border border-blue-200 rounded-xl p-5 flex items-start gap-4">
          <CheckCircle className="h-5 w-5 text-blue-600 mt-0.5 shrink-0" />
          <div>
            <p className="text-sm font-semibold text-blue-800">
              Recommended Vendor: {recommendData.vendor_name}
            </p>
            <p className="text-xs text-blue-700 mt-0.5">
              Score {recommendData.score.toFixed(0)} &middot; Unit cost {cents(recommendData.unit_cost)}
            </p>
            <p className="text-xs text-blue-600 mt-1">{recommendData.reasoning}</p>
          </div>
        </div>
      )}

      {/* Price anomaly alerts */}
      <div>
        <h3 className="text-sm font-semibold text-slate-200 mb-3 flex items-center gap-2">
          <AlertTriangle className="h-4 w-4 text-amber-500" />
          Price Anomaly Alerts
        </h3>
        {anomalyLoading ? (
          <div className="flex justify-center py-6"><LoadingSpinner /></div>
        ) : anomalies.length === 0 ? (
          <div className="text-sm text-slate-300 py-4">No price anomalies detected.</div>
        ) : (
          <div className="space-y-3">
            {anomalies.map((a, i) => (
              <div
                key={i}
                className="bg-white/5 rounded-xl border border-white/10 shadow-sm p-4 flex items-start justify-between gap-4"
              >
                <div>
                  <p className="text-sm font-semibold text-white">{a.ingredient_name}</p>
                  <p className="text-xs text-slate-400 mt-0.5">
                    {a.vendor_name} &middot; Current: {cents(a.current_price)} &middot; Avg: {cents(a.avg_price)} &middot; Z-score: {a.z_score.toFixed(2)}
                  </p>
                </div>
                <StatusBadge variant={a.severity === 'critical' ? 'critical' : 'warning'}>
                  {a.severity}
                </StatusBadge>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

// ─── Tab 3: Comparison ────────────────────────────────────────────────────────

function ComparisonTab({ locationId }: { locationId: string }) {
  const [selectedIngredientId, setSelectedIngredientId] = useState<string>('');

  const { data: inventoryData } = usePARStatus(locationId);
  const ingredients = inventoryData?.par_status ?? [];

  const { data: compareData, isLoading } = useVendorCompare(
    locationId,
    selectedIngredientId || null
  );

  const vendors = compareData?.vendors ?? [];
  const recommended = compareData?.recommended ?? '';

  return (
    <div className="space-y-6">
      <div>
        <label className="block text-xs font-medium text-slate-400 mb-1">Ingredient</label>
        <select
          value={selectedIngredientId}
          onChange={(e) => setSelectedIngredientId(e.target.value)}
          className="border border-white/10 rounded-lg px-3 py-2 text-sm text-white bg-white focus:outline-none focus:ring-2 focus:ring-blue-500"
        >
          <option value="">Select ingredient…</option>
          {ingredients.map((ing) => (
            <option key={ing.ingredient_id} value={ing.ingredient_id}>
              {ing.ingredient_name}
            </option>
          ))}
        </select>
      </div>

      {!selectedIngredientId ? (
        <div className="text-center py-16 text-slate-300 text-sm">
          Select an ingredient to compare vendors side by side.
        </div>
      ) : isLoading ? (
        <div className="flex justify-center py-12"><LoadingSpinner /></div>
      ) : vendors.length === 0 ? (
        <div className="text-center py-12 text-slate-300 text-sm">
          No vendor comparison data available for this ingredient.
        </div>
      ) : (
        <>
          {compareData?.ingredient_name && (
            <p className="text-sm text-slate-400">
              Comparing vendors for <span className="font-semibold text-slate-200">{compareData.ingredient_name}</span>
            </p>
          )}
          <div className="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-3 gap-5">
            {vendors.map((v) => {
              const isRecommended = v.vendor_name === recommended;
              return (
                <div
                  key={v.vendor_name}
                  className={`bg-white/5 rounded-xl border-2 p-5 space-y-4 transition ${
                    isRecommended
                      ? 'border-blue-500 ring-2 ring-blue-500/30'
                      : 'border-white/10'
                  }`}
                >
                  <div className="flex items-start justify-between gap-2">
                    <h3 className="font-semibold text-white text-base leading-tight">{v.vendor_name}</h3>
                    {isRecommended && (
                      <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full bg-blue-100 text-blue-700 text-xs font-medium shrink-0">
                        <TrendingUp className="h-3 w-3" />
                        Recommended
                      </span>
                    )}
                  </div>

                  <div className="grid grid-cols-2 gap-3">
                    <div className="bg-white/5 rounded-lg p-3 text-center">
                      <p className="text-xs text-slate-300">Score</p>
                      <p className={`text-xl font-bold tabular-nums ${scoreColor(v.overall_score)}`}>
                        {v.overall_score.toFixed(0)}
                      </p>
                    </div>
                    <div className="bg-white/5 rounded-lg p-3 text-center">
                      <p className="text-xs text-slate-300">Unit Cost</p>
                      <p className="text-xl font-bold text-white">{cents(v.unit_cost)}</p>
                    </div>
                    <div className="bg-white/5 rounded-lg p-3 text-center">
                      <p className="text-xs text-slate-300">OTIF</p>
                      <p className="text-lg font-semibold text-slate-200">{pct(v.otif_rate)}</p>
                    </div>
                    <div className="bg-white/5 rounded-lg p-3 text-center">
                      <p className="text-xs text-slate-300">Lead Time</p>
                      <p className="text-lg font-semibold text-slate-200">{v.avg_lead_days.toFixed(1)}d</p>
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        </>
      )}
    </div>
  );
}

// ─── Main page ────────────────────────────────────────────────────────────────

export default function VendorPage() {
  const locationId = useLocationStore((s) => s.selectedLocationId);
  const [activeTab, setActiveTab] = useState<Tab>('scorecards');

  const { data: scoresData, isLoading: scoresLoading, error: scoresError, refetch: refetchScores } = useVendorScores(locationId);

  if (!locationId) return <LoadingSpinner fullPage />;

  const vendors = scoresData?.vendor_scores ?? [];
  const topVendor = vendors.length > 0 ? vendors.reduce((a, b) => a.overall_score > b.overall_score ? a : b) : null;
  const avgScore = vendors.length > 0 ? (vendors.reduce((s, v) => s + v.overall_score, 0) / vendors.length) : 0;
  const avgOTIF = vendors.length > 0 ? (vendors.reduce((s, v) => s + (v.otif_rate ?? 0), 0) / vendors.length) : 0;

  return (
    <div className="space-y-8">
      {/* Page header */}
      <div>
        <h1 className="text-2xl font-bold text-white">Vendor Intelligence</h1>
        <p className="text-sm text-slate-400 mt-1">
          Reliability scorecards, price intelligence, and vendor comparison
        </p>
      </div>

      {scoresError && (
        <ErrorBanner
          message={scoresError instanceof Error ? scoresError.message : 'Failed to load vendor data'}
          retry={() => refetchScores()}
        />
      )}

      {/* KPI Cards */}
      {scoresLoading ? (
        <div className="flex justify-center py-8">
          <LoadingSpinner />
        </div>
      ) : (
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
          <KPICard
            label="Total Vendors"
            value={String(vendors.length)}
            icon={Truck}
            iconColor="text-slate-300"
            bgTint="bg-white/10"
          />
          <KPICard
            label="Avg Reliability"
            value={`${avgScore.toFixed(0)}/100`}
            icon={Star}
            iconColor="text-blue-600"
            bgTint="bg-blue-50"
          />
          <KPICard
            label="Top Vendor"
            value={topVendor ? `${topVendor.vendor_name} (${topVendor.overall_score.toFixed(0)})` : '—'}
            icon={CheckCircle}
            iconColor="text-emerald-600"
            bgTint="bg-emerald-50"
          />
          <KPICard
            label="Avg OTIF Rate"
            value={`${avgOTIF.toFixed(0)}%`}
            icon={TrendingUp}
            iconColor="text-purple-600"
            bgTint="bg-purple-50"
          />
        </div>
      )}

      {/* Tabs */}
      <div>
        <div className="flex gap-1 border-b border-white/10 mb-6">
          {TABS.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={`px-4 py-2.5 text-sm font-medium rounded-t-lg transition ${
                activeTab === tab.id
                  ? 'bg-white border border-b-white border-white/10 text-blue-600 -mb-px'
                  : 'text-slate-400 hover:text-slate-200'
              }`}
            >
              {tab.label}
            </button>
          ))}
        </div>

        {activeTab === 'scorecards' && <ScorecardsTab locationId={locationId} />}
        {activeTab === 'price' && <PriceIntelligenceTab locationId={locationId} />}
        {activeTab === 'comparison' && <ComparisonTab locationId={locationId} />}
      </div>
    </div>
  );
}
