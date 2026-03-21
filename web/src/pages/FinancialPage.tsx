import { useState } from 'react';
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
  Legend,
} from 'recharts';
import { useLocationStore } from '../stores/location';
import {
  usePnL,
  useAnomalies,
  useBudgetVariance,
  useCostCenters,
  useTxAnomalies,
  usePeriodComparison,
  useListBudgets,
  useCreateBudget,
} from '../hooks/useFinancial';
import KPICard from '../components/ui/KPICard';
import DataTable from '../components/ui/DataTable';
import type { Column } from '../components/ui/DataTable';
import StatusBadge from '../components/ui/StatusBadge';
import LoadingSpinner from '../components/ui/LoadingSpinner';
import ErrorBanner from '../components/ui/ErrorBanner';
import {
  DollarSign,
  TrendingDown,
  TrendingUp,
  Percent,
  AlertTriangle,
  Clock,
  Tag,
  Scissors,
  ChevronDown,
  ChevronRight,
} from 'lucide-react';
import type {
  ChannelBreakdown,
  Anomaly,
  CostCenter,
  TransactionAnomaly,
  Budget,
  BudgetVariance,
} from '../lib/api';

// ── helpers ──────────────────────────────────────────────────────────────────

function cents(v: number): string {
  return `EGP ${(v / 100).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
}

function pctArrow(v: number): string {
  return `${v >= 0 ? '+' : ''}${v.toFixed(1)}%`;
}

const CHANNEL_LABELS: Record<string, string> = {
  dine_in: 'Dine-in',
  takeout: 'Takeout',
  delivery: 'Delivery',
  drive_thru: 'Drive-thru',
};

const CATEGORY_COLORS: Record<string, string> = {
  protein: '#ef4444',
  produce: '#22c55e',
  dairy: '#3b82f6',
  bakery: '#f59e0b',
  frozen: '#6366f1',
  sauce: '#ec4899',
  other: '#9ca3af',
};

function categoryColor(cat: string): string {
  return CATEGORY_COLORS[cat.toLowerCase()] ?? '#9ca3af';
}

// ── tab bar ──────────────────────────────────────────────────────────────────

type Tab = 'pnl' | 'cost-centers' | 'anomalies' | 'budget';

const TABS: { id: Tab; label: string }[] = [
  { id: 'pnl', label: 'P&L' },
  { id: 'cost-centers', label: 'Cost Centers' },
  { id: 'anomalies', label: 'Anomalies' },
  { id: 'budget', label: 'Budget' },
];

// ── channel columns ───────────────────────────────────────────────────────────

const channelColumns: Column<ChannelBreakdown>[] = [
  { key: 'channel', header: 'Channel', render: (r) => CHANNEL_LABELS[r.channel] ?? r.channel },
  { key: 'revenue', header: 'Revenue', align: 'right', sortable: true, render: (r) => cents(r.revenue) },
  { key: 'cogs', header: 'COGS', align: 'right', render: (r) => cents(r.cogs) },
  { key: 'gross_margin', header: 'Margin %', align: 'right', sortable: true, render: (r) => `${r.gross_margin.toFixed(1)}%` },
  { key: 'check_count', header: 'Checks', align: 'right', sortable: true },
  { key: 'avg_check_size', header: 'Avg Check', align: 'right', render: (r) => cents(r.avg_check_size) },
];

// ── budget variance badge ─────────────────────────────────────────────────────

function VarianceBadge({ variance }: { variance: BudgetVariance | undefined }) {
  if (!variance) return null;
  const { status, revenue_variance_pct } = variance;
  if (status === 'on_track') return <StatusBadge variant="success">On Track</StatusBadge>;
  if (status === 'over')
    return (
      <StatusBadge variant="critical">
        {Math.abs(revenue_variance_pct).toFixed(1)}% Over
      </StatusBadge>
    );
  return (
    <StatusBadge variant="info">
      {Math.abs(revenue_variance_pct).toFixed(1)}% Under
    </StatusBadge>
  );
}

// ── period comparison row ─────────────────────────────────────────────────────

function PeriodRow({ pct, label }: { pct: number | null | undefined; label: string }) {
  if (pct == null) return null;
  const up = pct >= 0;
  return (
    <span className={`text-sm font-medium ${up ? 'text-emerald-600' : 'text-red-500'}`}>
      {label}: {pctArrow(pct)} {up ? '▲' : '▼'}
    </span>
  );
}

// ── transaction anomaly icon ──────────────────────────────────────────────────

function TxIcon({ type }: { type: string }) {
  const t = type.toLowerCase();
  if (t.includes('void')) return <Scissors className="h-5 w-5 text-red-500" />;
  if (t.includes('comp')) return <Tag className="h-5 w-5 text-amber-500" />;
  if (t.includes('off') || t.includes('hour')) return <Clock className="h-5 w-5 text-blue-500" />;
  if (t.includes('discount')) return <Percent className="h-5 w-5 text-purple-500" />;
  return <AlertTriangle className="h-5 w-5 text-slate-400" />;
}

// ─────────────────────────────────────────────────────────────────────────────
// Tab 1: P&L
// ─────────────────────────────────────────────────────────────────────────────

function PnLTab({ locationId }: { locationId: string }) {
  const { data: pnl, isLoading: pnlLoading, error: pnlError, refetch: refetchPnl } = usePnL(locationId);
  const { data: variance } = useBudgetVariance(locationId);
  const { data: period } = usePeriodComparison(locationId);
  const { data: anomalyData, isLoading: anomLoading } = useAnomalies(locationId);

  const chartData =
    pnl?.by_channel?.map((ch) => ({
      channel: CHANNEL_LABELS[ch.channel] ?? ch.channel,
      revenue: ch.revenue / 100,
    })) ?? [];

  return (
    <div className="space-y-8">
      {pnlError && (
        <ErrorBanner
          message={pnlError instanceof Error ? pnlError.message : 'Failed to load financial data'}
          retry={() => refetchPnl()}
        />
      )}

      {/* KPI cards */}
      {pnlLoading ? (
        <div className="flex justify-center py-8">
          <LoadingSpinner />
        </div>
      ) : pnl ? (
        <>
          <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-4">
            <KPICard label="Gross Revenue" value={cents(pnl.gross_revenue)} icon={DollarSign} iconColor="text-emerald-600" bgTint="bg-emerald-50" />
            <KPICard label="Net Revenue" value={cents(pnl.net_revenue)} icon={TrendingUp} iconColor="text-blue-600" bgTint="bg-blue-50" />
            <KPICard label="COGS" value={cents(pnl.cogs)} icon={TrendingDown} iconColor="text-red-600" bgTint="bg-red-50" />
            <KPICard label="Gross Profit" value={cents(pnl.gross_profit)} icon={DollarSign} iconColor="text-emerald-600" bgTint="bg-emerald-50" />
            <KPICard label="Margin %" value={`${pnl.gross_margin.toFixed(1)}%`} icon={Percent} iconColor="text-purple-600" bgTint="bg-purple-50" />
          </div>

          {/* budget variance badge + period comparison */}
          <div className="flex flex-wrap items-center gap-4">
            <VarianceBadge variance={variance} />
            <PeriodRow pct={period?.revenue_vs_last_week_pct} label="vs last week" />
            <PeriodRow pct={period?.revenue_vs_last_month_pct} label="vs last month" />
          </div>
        </>
      ) : null}

      {/* Channel Revenue Chart */}
      {chartData.length > 0 && (
        <div className="bg-white/5 rounded-xl border border-white/10 p-6 shadow-sm">
          <h2 className="text-lg font-semibold text-white mb-4">Revenue by Channel</h2>
          <div className="h-72">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={chartData} margin={{ top: 5, right: 20, left: 0, bottom: 5 }}>
                <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
                <XAxis dataKey="channel" tick={{ fontSize: 13 }} />
                <YAxis tick={{ fontSize: 13 }} tickFormatter={(v: number) => `EGP v.toLocaleString()}`} />
                <Tooltip
                  formatter={(value) => [`EGP Number(value).toLocaleString()}`, 'Revenue']}
                  contentStyle={{ borderRadius: '8px', border: "1px solid #334155", fontSize: '13px' }}
                />
                <Bar dataKey="revenue" fill="#F97316" radius={[6, 6, 0, 0]} />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </div>
      )}

      {/* Channel Breakdown Table */}
      <div>
        <h2 className="text-lg font-semibold text-white mb-3">Channel Breakdown</h2>
        <DataTable
          columns={channelColumns}
          data={pnl?.by_channel ?? []}
          keyExtractor={(r) => r.channel}
          isLoading={pnlLoading}
          emptyTitle="No channel data"
        />
      </div>

      {/* Z-score Anomalies */}
      {!anomLoading && anomalyData?.anomalies && anomalyData.anomalies.length > 0 && (
        <div>
          <h2 className="text-lg font-semibold text-white mb-3">Statistical Anomalies</h2>
          <div className="space-y-3">
            {anomalyData.anomalies.map((a: Anomaly, i: number) => (
              <div key={i} className="bg-white/5 rounded-lg border border-white/10 p-4 flex items-start gap-3 shadow-sm">
                <StatusBadge variant={a.severity === 'critical' ? 'critical' : 'warning'}>
                  {a.severity} (z={a.z_score.toFixed(1)})
                </StatusBadge>
                <div>
                  <p className="font-medium text-white">{a.metric_name.replace(/_/g, ' ')}</p>
                  <p className="text-sm text-slate-400">
                    Expected ~{a.mean.toFixed(0)} ± {a.std_dev.toFixed(0)}, got {a.current_value.toFixed(0)}
                  </p>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

// ─────────────────────────────────────────────────────────────────────────────
// Tab 2: Cost Centers
// ─────────────────────────────────────────────────────────────────────────────

function CostCentersTab({ locationId }: { locationId: string }) {
  const { data, isLoading, error, refetch } = useCostCenters(locationId);
  const [expanded, setExpanded] = useState<string | null>(null);

  const centers: CostCenter[] = data?.cost_centers ?? [];

  const pieData = centers.map((c) => ({
    name: c.category,
    value: c.cogs,
  }));

  const costCenterColumns: Column<CostCenter>[] = [
    {
      key: 'category',
      header: 'Category',
      render: (r) => (
        <button
          className="flex items-center gap-1 text-left font-medium text-white hover:text-orange-600"
          onClick={() => setExpanded(expanded === r.category ? null : r.category)}
        >
          {expanded === r.category ? (
            <ChevronDown className="h-4 w-4" />
          ) : (
            <ChevronRight className="h-4 w-4" />
          )}
          <span className="capitalize">{r.category}</span>
        </button>
      ),
    },
    { key: 'cogs', header: 'COGS $', align: 'right', sortable: true, render: (r) => cents(r.cogs) },
    { key: 'cogs_pct', header: 'COGS %', align: 'right', sortable: true, render: (r) => `${r.cogs_pct.toFixed(1)}%` },
    { key: 'revenue_pct', header: 'Rev %', align: 'right', render: (r) => `${r.revenue_pct.toFixed(1)}%` },
    { key: 'ingredient_count', header: 'Ingredients', align: 'right', render: (r) => r.ingredient_count.toString() },
  ];

  return (
    <div className="space-y-8">
      {error && (
        <ErrorBanner
          message={error instanceof Error ? error.message : 'Failed to load cost centers'}
          retry={() => refetch()}
        />
      )}

      {isLoading ? (
        <div className="flex justify-center py-8">
          <LoadingSpinner />
        </div>
      ) : (
        <>
          {/* Pie chart */}
          {pieData.length > 0 && (
            <div className="bg-white/5 rounded-xl border border-white/10 p-6 shadow-sm">
              <h2 className="text-lg font-semibold text-white mb-4">COGS by Category</h2>
              <div className="h-72">
                <ResponsiveContainer width="100%" height="100%">
                  <PieChart>
                    <Pie
                      data={pieData}
                      dataKey="value"
                      nameKey="name"
                      cx="50%"
                      cy="50%"
                      outerRadius={100}
                      label={({ name, percent }) =>
                        `${name} ${(percent * 100).toFixed(1)}%`
                      }
                    >
                      {pieData.map((entry) => (
                        <Cell key={entry.name} fill={categoryColor(entry.name)} />
                      ))}
                    </Pie>
                    <Tooltip
                      formatter={(value) => [cents(Number(value)), 'COGS']}
                      contentStyle={{ borderRadius: '8px', border: "1px solid #334155", fontSize: '13px' }}
                    />
                    <Legend />
                  </PieChart>
                </ResponsiveContainer>
              </div>
            </div>
          )}

          {/* Cost center table with expandable rows */}
          <div>
            <h2 className="text-lg font-semibold text-white mb-3">Category Breakdown</h2>
            <DataTable
              columns={costCenterColumns}
              data={centers}
              keyExtractor={(r) => r.category}
              isLoading={isLoading}
              emptyTitle="No cost center data"
            />

            {/* Expanded ingredient rows */}
            {expanded && (() => {
              const center = centers.find((c) => c.category === expanded);
              if (!center || center.top_ingredients.length === 0) return null;
              return (
                <div className="mt-2 ml-4 border-l-2 border-orange-200 pl-4">
                  <p className="text-xs font-semibold text-slate-400 uppercase mb-2">
                    Top Ingredients — {expanded}
                  </p>
                  <div className="space-y-2">
                    {center.top_ingredients.slice(0, 5).map((ing) => (
                      <div
                        key={ing.ingredient_id}
                        className="flex items-center justify-between bg-white/5 rounded-lg px-4 py-2 text-sm"
                      >
                        <span className="font-medium text-slate-200">{ing.ingredient_name}</span>
                        <div className="flex items-center gap-6 text-slate-400">
                          <span>{cents(ing.total_cost)}</span>
                          <span>{ing.cost_pct.toFixed(1)}% of cat.</span>
                          <span>
                            {ing.quantity_used.toFixed(2)} {ing.unit}
                          </span>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              );
            })()}
          </div>
        </>
      )}
    </div>
  );
}

// ─────────────────────────────────────────────────────────────────────────────
// Tab 3: Anomalies
// ─────────────────────────────────────────────────────────────────────────────

function AnomaliesTab({ locationId }: { locationId: string }) {
  const { data: zsData, isLoading: zsLoading } = useAnomalies(locationId);
  const { data: txData, isLoading: txLoading, error: txError, refetch: txRefetch } = useTxAnomalies(locationId);

  const txAnomalies: TransactionAnomaly[] = txData?.anomalies ?? [];

  const severityVariant = (s: string): 'critical' | 'warning' | 'info' => {
    if (s === 'critical') return 'critical';
    if (s === 'warning') return 'warning';
    return 'info';
  };

  return (
    <div className="space-y-8">
      {/* Z-score anomalies */}
      <div>
        <h2 className="text-lg font-semibold text-white mb-3">Statistical Anomalies (Z-score)</h2>
        {zsLoading ? (
          <div className="flex justify-center py-4">
            <LoadingSpinner />
          </div>
        ) : zsData?.anomalies && zsData.anomalies.length > 0 ? (
          <div className="space-y-3">
            {zsData.anomalies.map((a: Anomaly, i: number) => (
              <div
                key={i}
                className="bg-white/5 rounded-lg border border-white/10 p-4 flex items-start gap-3 shadow-sm"
              >
                <StatusBadge variant={a.severity === 'critical' ? 'critical' : 'warning'}>
                  {a.severity} (z={a.z_score.toFixed(1)})
                </StatusBadge>
                <div>
                  <p className="font-medium text-white">{a.metric_name.replace(/_/g, ' ')}</p>
                  <p className="text-sm text-slate-400">
                    Expected ~{a.mean.toFixed(0)} ± {a.std_dev.toFixed(0)}, got{' '}
                    {a.current_value.toFixed(0)}
                  </p>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <p className="text-sm text-slate-300">No statistical anomalies detected.</p>
        )}
      </div>

      {/* Transaction anomalies */}
      <div>
        <h2 className="text-lg font-semibold text-white mb-3">Transaction Anomalies</h2>
        {txError && (
          <ErrorBanner
            message={txError instanceof Error ? txError.message : 'Failed to load transaction anomalies'}
            retry={() => txRefetch()}
          />
        )}
        {txLoading ? (
          <div className="flex justify-center py-4">
            <LoadingSpinner />
          </div>
        ) : txAnomalies.length > 0 ? (
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            {txAnomalies.map((a, i) => (
              <div
                key={i}
                className="bg-white/5 rounded-xl border border-white/10 p-5 shadow-sm flex gap-4"
              >
                <div className="flex-shrink-0 mt-0.5">
                  <TxIcon type={a.type} />
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center justify-between gap-2 mb-1">
                    <p className="font-semibold text-white capitalize text-sm">
                      {a.type.replace(/_/g, ' ')}
                    </p>
                    <StatusBadge variant={severityVariant(a.severity)}>
                      {a.severity}
                    </StatusBadge>
                  </div>
                  <p className="text-sm text-slate-400 mb-2">{a.description}</p>
                  <div className="flex flex-wrap gap-x-4 gap-y-1 text-xs text-slate-400">
                    <span>
                      Current:{' '}
                      <span className="font-medium text-slate-200">
                        {a.current_value.toFixed(2)}
                      </span>
                    </span>
                    <span>
                      Baseline:{' '}
                      <span className="font-medium text-slate-200">{a.baseline.toFixed(2)}</span>
                    </span>
                    <span>
                      Z-score:{' '}
                      <span className="font-medium text-slate-200">{a.z_score.toFixed(2)}</span>
                    </span>
                  </div>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <p className="text-sm text-slate-300">No transaction anomalies detected.</p>
        )}
      </div>
    </div>
  );
}

// ─────────────────────────────────────────────────────────────────────────────
// Tab 4: Budget
// ─────────────────────────────────────────────────────────────────────────────

function BudgetTab({ locationId }: { locationId: string }) {
  const { data: varianceData, isLoading: vLoading, refetch: vRefetch } = useBudgetVariance(locationId);
  const { data: budgetList, isLoading: bLoading } = useListBudgets(locationId);
  const createBudget = useCreateBudget();

  const [periodType, setPeriodType] = useState<'daily' | 'weekly' | 'monthly'>('weekly');
  const [revenueTarget, setRevenueTarget] = useState('');
  const [foodCostPct, setFoodCostPct] = useState('');
  const [cogsTarget, setCogsTarget] = useState('');
  const [saved, setSaved] = useState(false);

  async function handleSave(e: React.FormEvent) {
    e.preventDefault();
    await createBudget.mutateAsync({
      location_id: locationId,
      period_type: periodType,
      revenue_target: Math.round(parseFloat(revenueTarget) * 100),
      food_cost_pct_target: parseFloat(foodCostPct),
      labor_cost_pct_target: 0,
      cogs_target: Math.round(parseFloat(cogsTarget) * 100),
    });
    setSaved(true);
    setTimeout(() => setSaved(false), 3000);
    vRefetch();
  }

  const v = varianceData;

  const varianceRowClass = (status: string) => {
    if (status === 'on_track') return 'text-emerald-700 bg-emerald-50';
    if (status === 'over') return 'text-red-700 bg-red-50';
    return 'text-blue-700 bg-blue-50';
  };

  const budgets: Budget[] = budgetList?.budgets ?? [];

  return (
    <div className="space-y-8">
      {/* Budget form */}
      <div className="bg-white/5 rounded-xl border border-white/10 p-6 shadow-sm max-w-lg">
        <h2 className="text-lg font-semibold text-white mb-4">Create Budget</h2>
        <form onSubmit={handleSave} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-slate-200 mb-1">Period Type</label>
            <select
              className="w-full border border-white/20 rounded-lg px-3 py-2 text-sm bg-white/10 text-white focus:outline-none focus:ring-2 focus:ring-orange-400"
              value={periodType}
              onChange={(e) => setPeriodType(e.target.value as typeof periodType)}
            >
              <option value="daily">Daily</option>
              <option value="weekly">Weekly</option>
              <option value="monthly">Monthly</option>
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-slate-200 mb-1">Revenue Target ($)</label>
            <input
              type="number"
              min="0"
              step="0.01"
              required
              className="w-full border border-white/20 rounded-lg px-3 py-2 text-sm bg-white/10 text-white focus:outline-none focus:ring-2 focus:ring-orange-400"
              placeholder="e.g. 15000.00"
              value={revenueTarget}
              onChange={(e) => setRevenueTarget(e.target.value)}
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-slate-200 mb-1">Food Cost % Target</label>
            <input
              type="number"
              min="0"
              max="100"
              step="0.1"
              required
              className="w-full border border-white/20 rounded-lg px-3 py-2 text-sm bg-white/10 text-white focus:outline-none focus:ring-2 focus:ring-orange-400"
              placeholder="e.g. 28.5"
              value={foodCostPct}
              onChange={(e) => setFoodCostPct(e.target.value)}
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-slate-200 mb-1">COGS Target ($)</label>
            <input
              type="number"
              min="0"
              step="0.01"
              required
              className="w-full border border-white/20 rounded-lg px-3 py-2 text-sm bg-white/10 text-white focus:outline-none focus:ring-2 focus:ring-orange-400"
              placeholder="e.g. 4275.00"
              value={cogsTarget}
              onChange={(e) => setCogsTarget(e.target.value)}
            />
          </div>
          <button
            type="submit"
            disabled={createBudget.isPending}
            className="w-full bg-orange-500 hover:bg-orange-600 disabled:opacity-50 text-white font-semibold rounded-lg px-4 py-2 text-sm transition-colors"
          >
            {createBudget.isPending ? 'Saving…' : 'Save Budget'}
          </button>
          {saved && (
            <p className="text-sm text-emerald-600 font-medium text-center">Budget saved successfully.</p>
          )}
          {createBudget.isError && (
            <p className="text-sm text-red-600 font-medium text-center">
              {createBudget.error instanceof Error ? createBudget.error.message : 'Save failed.'}
            </p>
          )}
        </form>
      </div>

      {/* Current variance */}
      {vLoading ? (
        <div className="flex justify-center py-4">
          <LoadingSpinner />
        </div>
      ) : v ? (
        <div className="bg-white/5 rounded-xl border border-white/10 shadow-sm overflow-hidden">
          <div className="px-6 py-4 border-b border-white/5 flex items-center justify-between">
            <h2 className="text-lg font-semibold text-white">Budget vs Actual</h2>
            <StatusBadge
              variant={
                v.status === 'on_track' ? 'success' : v.status === 'over' ? 'critical' : 'info'
              }
            >
              {v.status === 'on_track' ? 'On Track' : v.status === 'over' ? 'Over Budget' : 'Under Budget'}
            </StatusBadge>
          </div>
          <table className="w-full text-sm">
            <thead className="bg-white/5 text-xs font-semibold text-slate-400 uppercase">
              <tr>
                <th className="px-6 py-3 text-left">Metric</th>
                <th className="px-6 py-3 text-right">Target</th>
                <th className="px-6 py-3 text-right">Actual</th>
                <th className="px-6 py-3 text-right">Variance</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-white/5">
              <tr className={varianceRowClass(v.status)}>
                <td className="px-6 py-3 font-medium">Revenue</td>
                <td className="px-6 py-3 text-right">{cents(v.budget.revenue_target)}</td>
                <td className="px-6 py-3 text-right">{cents(v.actual_revenue)}</td>
                <td className="px-6 py-3 text-right font-semibold">
                  {pctArrow(v.revenue_variance_pct)}
                </td>
              </tr>
              <tr
                className={
                  v.cogs_variance_pct > 0 ? 'text-red-700 bg-red-50' : 'text-emerald-700 bg-emerald-50'
                }
              >
                <td className="px-6 py-3 font-medium">COGS</td>
                <td className="px-6 py-3 text-right">{cents(v.budget.cogs_target)}</td>
                <td className="px-6 py-3 text-right">{cents(v.actual_cogs)}</td>
                <td className="px-6 py-3 text-right font-semibold">
                  {pctArrow(v.cogs_variance_pct)}
                </td>
              </tr>
              <tr
                className={
                  v.food_cost_pct_delta > 0 ? 'text-red-700 bg-red-50' : 'text-emerald-700 bg-emerald-50'
                }
              >
                <td className="px-6 py-3 font-medium">Food Cost %</td>
                <td className="px-6 py-3 text-right">{v.budget.food_cost_pct_target.toFixed(1)}%</td>
                <td className="px-6 py-3 text-right">{v.actual_food_cost_pct.toFixed(1)}%</td>
                <td className="px-6 py-3 text-right font-semibold">
                  {v.food_cost_pct_delta >= 0 ? '+' : ''}
                  {v.food_cost_pct_delta.toFixed(1)} pts
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      ) : null}

      {/* Budget history */}
      {!bLoading && budgets.length > 0 && (
        <div>
          <h2 className="text-lg font-semibold text-white mb-3">Budget History</h2>
          <div className="bg-white/5 rounded-xl border border-white/10 shadow-sm overflow-hidden">
            <table className="w-full text-sm">
              <thead className="bg-white/5 text-xs font-semibold text-slate-400 uppercase">
                <tr>
                  <th className="px-6 py-3 text-left">Period</th>
                  <th className="px-6 py-3 text-left">Type</th>
                  <th className="px-6 py-3 text-right">Revenue Target</th>
                  <th className="px-6 py-3 text-right">Food Cost %</th>
                  <th className="px-6 py-3 text-right">COGS Target</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-white/5">
                {budgets.map((b) => (
                  <tr key={b.budget_id} className="hover:bg-white/5 text-slate-300">
                    <td className="px-6 py-3 text-slate-300">
                      {new Date(b.period_start).toLocaleDateString()} –{' '}
                      {new Date(b.period_end).toLocaleDateString()}
                    </td>
                    <td className="px-6 py-3 capitalize">{b.period_type}</td>
                    <td className="px-6 py-3 text-right">{cents(b.revenue_target)}</td>
                    <td className="px-6 py-3 text-right">{b.food_cost_pct_target.toFixed(1)}%</td>
                    <td className="px-6 py-3 text-right">{cents(b.cogs_target)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  );
}

// ─────────────────────────────────────────────────────────────────────────────
// Root page
// ─────────────────────────────────────────────────────────────────────────────

export default function FinancialPage() {
  const locationId = useLocationStore((s) => s.selectedLocationId);
  const [activeTab, setActiveTab] = useState<Tab>('pnl');

  if (!locationId) return <LoadingSpinner fullPage />;

  return (
    <div className="space-y-6">
      {/* Page header */}
      <div>
        <h1 className="text-2xl font-bold text-white">Financial Intelligence</h1>
        <p className="text-sm text-slate-400 mt-1">P&L, cost centers, anomalies, and budget tracking</p>
      </div>

      {/* Tab bar */}
      <div className="flex gap-1 bg-white/5 rounded-xl p-1 w-fit">
        {TABS.map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
              activeTab === tab.id
                ? 'bg-[#F97316] text-white shadow-sm'
                : 'text-slate-400 hover:text-slate-200'
            }`}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {/* Tab content */}
      {activeTab === 'pnl' && <PnLTab locationId={locationId} />}
      {activeTab === 'cost-centers' && <CostCentersTab locationId={locationId} />}
      {activeTab === 'anomalies' && <AnomaliesTab locationId={locationId} />}
      {activeTab === 'budget' && <BudgetTab locationId={locationId} />}
    </div>
  );
}
