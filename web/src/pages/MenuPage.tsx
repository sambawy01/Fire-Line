import { useState } from 'react';
import {
  RadarChart,
  Radar,
  PolarGrid,
  PolarAngleAxis,
  ResponsiveContainer,
} from 'recharts';
import { useLocationStore } from '../stores/location';
import {
  useMenuScores,
  useDependencies,
  useCrossSell,
  useScoreMenu,
  useSimulatePrice,
  useSimulateRemoval,
  useSimulateIngredientCost,
} from '../hooks/useMenuScoring';
import DataTable from '../components/ui/DataTable';
import type { Column } from '../components/ui/DataTable';
import StatusBadge from '../components/ui/StatusBadge';
import LoadingSpinner from '../components/ui/LoadingSpinner';
import ErrorBanner from '../components/ui/ErrorBanner';
import type { MenuItemScore, IngredientDependency, CrossSellPair, SimulationResult } from '../lib/api';

// ─── helpers ────────────────────────────────────────────────────────────────

function dollars(cents: number): string {
  return `EGP ${(cents / 100).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
}

function deltaCls(v: number) {
  return v >= 0 ? 'text-emerald-600 font-semibold' : 'text-red-600 font-semibold';
}

function deltaLabel(v: number, isCents = true): string {
  const formatted = isCents ? dollars(Math.abs(v)) : `EGP Math.abs(v / 100).toFixed(2)}`;
  return `${v >= 0 ? '+' : '-'}${formatted}`;
}

// ─── classification config ──────────────────────────────────────────────────

const CLASSIFICATION_COLOR: Record<MenuItemScore['classification'], string> = {
  powerhouse:       '#22c55e',
  hidden_gem:       '#8b5cf6',
  crowd_pleaser:    '#3b82f6',
  workhorse:        '#6b7280',
  complex_star:     '#f59e0b',
  declining_star:   '#ef4444',
  underperformer:   '#dc2626',
  strategic_anchor: '#06b6d4',
};

const CLASSIFICATION_LABEL: Record<MenuItemScore['classification'], string> = {
  powerhouse:       'Powerhouse',
  hidden_gem:       'Hidden Gem',
  crowd_pleaser:    'Crowd Pleaser',
  workhorse:        'Workhorse',
  complex_star:     'Complex Star',
  declining_star:   'Declining Star',
  underperformer:   'Underperformer',
  strategic_anchor: 'Strategic Anchor',
};

function ClassificationBadge({ cls }: { cls: MenuItemScore['classification'] }) {
  const color = CLASSIFICATION_COLOR[cls];
  const label = CLASSIFICATION_LABEL[cls];
  return (
    <span
      className="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium text-white"
      style={{ backgroundColor: color }}
    >
      {label}
    </span>
  );
}

// ─── radar chart for a single item ──────────────────────────────────────────

function ItemRadar({ item }: { item: MenuItemScore }) {
  const data = [
    { axis: 'Margin',       value: item.margin_score },
    { axis: 'Velocity',     value: item.velocity_score },
    { axis: 'Complexity',   value: item.complexity_score },
    { axis: 'Satisfaction', value: item.satisfaction_score },
    { axis: 'Strategic',    value: item.strategic_score },
  ];
  return (
    <ResponsiveContainer width="100%" height={220}>
      <RadarChart data={data}>
        <PolarGrid />
        <PolarAngleAxis dataKey="axis" tick={{ fontSize: 11 }} />
        <Radar
          name={item.name}
          dataKey="value"
          stroke={CLASSIFICATION_COLOR[item.classification]}
          fill={CLASSIFICATION_COLOR[item.classification]}
          fillOpacity={0.3}
        />
      </RadarChart>
    </ResponsiveContainer>
  );
}

// ─── simulation delta card ───────────────────────────────────────────────────

function SimDeltaCard({ result }: { result: SimulationResult }) {
  return (
    <div className="mt-4 rounded-lg border border-white/10 bg-gray-50 p-4 space-y-3">
      <h4 className="text-sm font-semibold text-slate-200">Simulation Results</h4>
      <div className="grid grid-cols-2 gap-3 text-sm">
        <div>
          <p className="text-slate-400 text-xs">Current Revenue</p>
          <p className="font-medium">{dollars(result.current_revenue)}</p>
        </div>
        <div>
          <p className="text-slate-400 text-xs">Projected Revenue</p>
          <p className="font-medium">{dollars(result.projected_revenue)}</p>
        </div>
        <div>
          <p className="text-slate-400 text-xs">Revenue Delta</p>
          <p className={deltaCls(result.revenue_delta)}>{deltaLabel(result.revenue_delta)}</p>
        </div>
        <div>
          <p className="text-slate-400 text-xs">Profit Delta</p>
          <p className={deltaCls(result.profit_delta)}>{deltaLabel(result.profit_delta)}</p>
        </div>
      </div>
    </div>
  );
}

// ─── tabs ────────────────────────────────────────────────────────────────────

const TABS = ['Menu Matrix', 'Simulation', 'Dependencies', 'Cross-Sell'] as const;
type Tab = typeof TABS[number];

// ─── main page ───────────────────────────────────────────────────────────────

export default function MenuPage() {
  const locationId = useLocationStore((s) => s.selectedLocationId);
  const [activeTab, setActiveTab] = useState<Tab>('Menu Matrix');

  // Menu Matrix state
  const [classFilter, setClassFilter] = useState<string>('all');
  const [expandedRow, setExpandedRow] = useState<string | null>(null);

  // Simulation state
  const [priceItemId, setPriceItemId] = useState('');
  const [newPrice, setNewPrice] = useState('');
  const [priceResult, setPriceResult] = useState<SimulationResult | null>(null);

  const [removeItemId, setRemoveItemId] = useState('');
  const [removeResult, setRemoveResult] = useState<SimulationResult | null>(null);

  const [costIngredientId, setCostIngredientId] = useState('');
  const [newCost, setNewCost] = useState('');
  const [costResult, setCostResult] = useState<SimulationResult | null>(null);

  // Queries
  const { data: scoresData, isLoading: scoresLoading, error: scoresError, refetch: refetchScores } = useMenuScores(locationId);
  const { data: depsData,   isLoading: depsLoading,   error: depsError }   = useDependencies(locationId);
  const { data: crossData,  isLoading: crossLoading,  error: crossError }  = useCrossSell(locationId);

  // Mutations
  const scoreMutation = useScoreMenu();
  const priceMutation  = useSimulatePrice();
  const removeMutation = useSimulateRemoval();
  const costMutation   = useSimulateIngredientCost();

  if (!locationId) return <LoadingSpinner fullPage />;

  const allItems   = scoresData?.items ?? [];
  const allDeps    = depsData?.dependencies ?? [];
  const allPairs   = crossData?.pairs ?? [];

  // classifications for filter
  const classifications = Array.from(new Set(allItems.map((i) => i.classification))).sort();

  const filteredItems =
    classFilter === 'all' ? allItems : allItems.filter((i) => i.classification === classFilter);

  // Score menu handler
  async function handleRecalculate() {
    await scoreMutation.mutateAsync(locationId);
    refetchScores();
  }

  // ── Tab 1: Menu Matrix columns ────────────────────────────────────────────

  const scoreColumns: Column<MenuItemScore>[] = [
    { key: 'name',     header: 'Item',     sortable: true },
    { key: 'category', header: 'Category', sortable: true },
    {
      key: 'price',
      header: 'Price',
      align: 'right',
      sortable: true,
      render: (r) => `EGP ${(r.price / 100).toFixed(2)}`,
    },
    {
      key: 'classification',
      header: 'Classification',
      align: 'center',
      render: (r) => <ClassificationBadge cls={r.classification} />,
    },
    {
      key: 'margin_score',
      header: 'Margin',
      align: 'right',
      sortable: true,
      render: (r) => r.margin_score.toFixed(1),
    },
    {
      key: 'velocity_score',
      header: 'Velocity',
      align: 'right',
      sortable: true,
      render: (r) => r.velocity_score.toFixed(1),
    },
    {
      key: 'complexity_score',
      header: 'Complexity',
      align: 'right',
      sortable: true,
      render: (r) => r.complexity_score.toFixed(1),
    },
    {
      key: 'satisfaction_score',
      header: 'Satisfaction',
      align: 'right',
      sortable: true,
      render: (r) => r.satisfaction_score.toFixed(1),
    },
    {
      key: 'strategic_score',
      header: 'Strategic',
      align: 'right',
      sortable: true,
      render: (r) => r.strategic_score.toFixed(1),
    },
    {
      key: 'expand',
      header: '',
      align: 'center',
      render: (r) => (
        <button
          onClick={() => setExpandedRow(expandedRow === r.menu_item_id ? null : r.menu_item_id)}
          className="text-xs font-medium text-[#F97316] hover:underline focus:outline-none"
        >
          {expandedRow === r.menu_item_id ? 'Close' : 'Detail'}
        </button>
      ),
    },
  ];

  // ── Tab 3: Dependencies columns ───────────────────────────────────────────

  const depColumns: Column<IngredientDependency>[] = [
    {
      key: 'ingredient_name',
      header: 'Ingredient',
      sortable: true,
      render: (r) => (
        <span className={r.menu_item_count === 1 ? 'text-red-600 font-semibold' : ''}>
          {r.ingredient_name}
          {r.menu_item_count === 1 && (
            <span className="ml-2 text-[10px] bg-red-100 text-red-700 rounded px-1 py-0.5">SPOF</span>
          )}
        </span>
      ),
    },
    {
      key: 'menu_item_count',
      header: '# Items',
      align: 'right',
      sortable: true,
    },
    {
      key: 'menu_items',
      header: 'Used By',
      render: (r) => (
        <span className="text-sm text-slate-300">{r.menu_items.join(', ')}</span>
      ),
    },
  ];

  // ── Tab 4: Cross-Sell columns ─────────────────────────────────────────────

  const maxAffinity = allPairs.length > 0 ? Math.max(...allPairs.map((p) => p.affinity)) : 1;

  const crossColumns: Column<CrossSellPair>[] = [
    { key: 'item_a_name', header: 'Item A', sortable: true },
    { key: 'item_b_name', header: 'Item B', sortable: true },
    {
      key: 'co_occurrences',
      header: 'Co-orders',
      align: 'right',
      sortable: true,
    },
    {
      key: 'affinity',
      header: 'Affinity',
      render: (r) => {
        const pct = maxAffinity > 0 ? (r.affinity / maxAffinity) * 100 : 0;
        return (
          <div className="flex items-center gap-2">
            <div className="w-28 h-2 rounded-full bg-gray-100 overflow-hidden">
              <div
                className="h-full rounded-full bg-[#F97316]"
                style={{ width: `${pct.toFixed(1)}%` }}
              />
            </div>
            <span className="text-xs text-slate-300 whitespace-nowrap">
              {(r.affinity * 100).toFixed(1)}%
            </span>
          </div>
        );
      },
    },
  ];

  // ── sorted deps ───────────────────────────────────────────────────────────
  const sortedDeps = [...allDeps].sort((a, b) => b.menu_item_count - a.menu_item_count);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold text-white">Menu Intelligence</h1>
        <p className="text-sm text-slate-400 mt-1">
          5-dimension scoring, simulation sandbox, and cross-sell analysis
        </p>
      </div>

      {/* Tab bar */}
      <div className="flex gap-1 border-b border-white/10">
        {TABS.map((tab) => (
          <button
            key={tab}
            onClick={() => setActiveTab(tab)}
            className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
              activeTab === tab
                ? 'border-[#F97316] text-[#F97316]'
                : 'border-transparent text-slate-400 hover:text-slate-200'
            }`}
          >
            {tab}
          </button>
        ))}
      </div>

      {/* ── Tab 1: Menu Matrix ── */}
      {activeTab === 'Menu Matrix' && (
        <div className="space-y-4">
          {scoresError && (
            <ErrorBanner
              message={scoresError instanceof Error ? scoresError.message : 'Failed to load scores'}
              retry={() => refetchScores()}
            />
          )}

          {/* Controls */}
          <div className="flex flex-wrap items-center gap-3">
            <button
              onClick={handleRecalculate}
              disabled={scoreMutation.isPending}
              className="px-4 py-2 text-sm font-medium rounded-lg bg-[#F97316] text-white hover:bg-orange-600 disabled:opacity-50 transition-colors"
            >
              {scoreMutation.isPending ? 'Recalculating…' : 'Recalculate Scores'}
            </button>
            {scoreMutation.isSuccess && (
              <span className="text-xs text-emerald-600 font-medium">Scores updated.</span>
            )}
            {classifications.length > 0 && (
              <select
                value={classFilter}
                onChange={(e) => setClassFilter(e.target.value)}
                className="text-sm border border-white/20 rounded-md px-3 py-1.5 bg-white/10 text-white focus:outline-none focus:ring-2 focus:ring-[#F97316]"
              >
                <option value="all">All Classifications</option>
                {classifications.map((c) => (
                  <option key={c} value={c}>
                    {CLASSIFICATION_LABEL[c]}
                  </option>
                ))}
              </select>
            )}
          </div>

          {/* Table with expandable rows */}
          {scoresLoading ? (
            <div className="flex justify-center py-12"><LoadingSpinner /></div>
          ) : (
            <div className="bg-white/5 rounded-xl border border-white/10 shadow-sm overflow-hidden">
              <DataTable
                columns={scoreColumns}
                data={filteredItems}
                keyExtractor={(r) => r.menu_item_id}
                emptyTitle="No scores available"
                emptyDescription="Click 'Recalculate Scores' to generate 5-dimension scores for this location."
              />
              {/* Expanded radar rows — rendered below table via overlay approach */}
              {expandedRow && (() => {
                const item = filteredItems.find((i) => i.menu_item_id === expandedRow);
                if (!item) return null;
                return (
                  <div className="border-t border-orange-500/20 bg-orange-500/10 px-6 py-4">
                    <div className="flex flex-col md:flex-row gap-6 items-start">
                      <div className="w-full md:w-72 shrink-0">
                        <ItemRadar item={item} />
                      </div>
                      <div className="flex-1 space-y-2 text-sm">
                        <div className="flex items-center gap-2 mb-2">
                          <ClassificationBadge cls={item.classification} />
                          <span className="text-slate-400">{item.category}</span>
                        </div>
                        <div className="grid grid-cols-2 sm:grid-cols-3 gap-3">
                          {[
                            { label: 'Margin Score',       val: item.margin_score },
                            { label: 'Velocity Score',     val: item.velocity_score },
                            { label: 'Complexity Score',   val: item.complexity_score },
                            { label: 'Satisfaction Score', val: item.satisfaction_score },
                            { label: 'Strategic Score',    val: item.strategic_score },
                          ].map(({ label, val }) => (
                            <div key={label} className="bg-white/5 rounded-lg border border-white/10 p-3">
                              <p className="text-[10px] text-slate-400 uppercase tracking-wide">{label}</p>
                              <p className="text-xl font-bold text-white mt-0.5">{val.toFixed(1)}</p>
                            </div>
                          ))}
                          <div className="bg-white/5 rounded-lg border border-white/10 p-3">
                            <p className="text-[10px] text-slate-400 uppercase tracking-wide">Units Sold</p>
                            <p className="text-xl font-bold text-white mt-0.5">{item.units_sold.toLocaleString()}</p>
                          </div>
                          <div className="bg-white/5 rounded-lg border border-white/10 p-3">
                            <p className="text-[10px] text-slate-400 uppercase tracking-wide">Contrib. Margin</p>
                            <p className="text-xl font-bold text-white mt-0.5">{dollars(item.contribution_margin)}</p>
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                );
              })()}
            </div>
          )}
        </div>
      )}

      {/* ── Tab 2: Simulation Sandbox ── */}
      {activeTab === 'Simulation' && (
        <div className="grid grid-cols-1 md:grid-cols-3 gap-5">

          {/* Card 1: Price Change */}
          <div className="bg-white/5 rounded-xl border border-white/10 shadow-sm p-5 space-y-4">
            <div>
              <h3 className="text-base font-semibold text-white">Price Change</h3>
              <p className="text-xs text-slate-400 mt-0.5">Simulate a new price point for a menu item.</p>
            </div>
            <div className="space-y-3">
              <div>
                <label className="block text-xs font-medium text-slate-300 mb-1">Menu Item</label>
                <select
                  value={priceItemId}
                  onChange={(e) => setPriceItemId(e.target.value)}
                  className="w-full text-sm border border-white/20 rounded-md px-3 py-1.5 bg-white/10 text-white focus:outline-none focus:ring-2 focus:ring-[#F97316]"
                >
                  <option value="">Select item…</option>
                  {allItems.map((i) => (
                    <option key={i.menu_item_id} value={i.menu_item_id}>{i.name}</option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-xs font-medium text-slate-300 mb-1">New Price ($)</label>
                <input
                  type="number"
                  min="0"
                  step="0.01"
                  value={newPrice}
                  onChange={(e) => setNewPrice(e.target.value)}
                  placeholder="e.g. 14.99"
                  className="w-full text-sm border border-white/20 rounded-md px-3 py-1.5 bg-white/10 text-white focus:outline-none focus:ring-2 focus:ring-[#F97316]"
                />
              </div>
              <button
                disabled={!priceItemId || !newPrice || priceMutation.isPending}
                onClick={async () => {
                  const result = await priceMutation.mutateAsync({
                    locationId,
                    menuItemId: priceItemId,
                    newPrice: Math.round(parseFloat(newPrice) * 100),
                  });
                  setPriceResult(result);
                }}
                className="w-full px-4 py-2 text-sm font-medium rounded-lg bg-[#F97316] text-white hover:bg-orange-600 disabled:opacity-50 transition-colors"
              >
                {priceMutation.isPending ? 'Simulating…' : 'Simulate'}
              </button>
              {priceMutation.isError && (
                <p className="text-xs text-red-600">{(priceMutation.error as Error).message}</p>
              )}
            </div>
            {priceResult && <SimDeltaCard result={priceResult} />}
          </div>

          {/* Card 2: Item Removal */}
          <div className="bg-white/5 rounded-xl border border-white/10 shadow-sm p-5 space-y-4">
            <div>
              <h3 className="text-base font-semibold text-white">Item Removal (86)</h3>
              <p className="text-xs text-slate-400 mt-0.5">Forecast impact of removing an item entirely.</p>
            </div>
            <div className="space-y-3">
              <div>
                <label className="block text-xs font-medium text-slate-300 mb-1">Menu Item</label>
                <select
                  value={removeItemId}
                  onChange={(e) => setRemoveItemId(e.target.value)}
                  className="w-full text-sm border border-white/20 rounded-md px-3 py-1.5 bg-white/10 text-white focus:outline-none focus:ring-2 focus:ring-[#F97316]"
                >
                  <option value="">Select item…</option>
                  {allItems.map((i) => (
                    <option key={i.menu_item_id} value={i.menu_item_id}>{i.name}</option>
                  ))}
                </select>
              </div>
              <button
                disabled={!removeItemId || removeMutation.isPending}
                onClick={async () => {
                  const result = await removeMutation.mutateAsync({
                    locationId,
                    menuItemId: removeItemId,
                  });
                  setRemoveResult(result);
                }}
                className="w-full px-4 py-2 text-sm font-medium rounded-lg bg-[#F97316] text-white hover:bg-orange-600 disabled:opacity-50 transition-colors"
              >
                {removeMutation.isPending ? 'Simulating…' : 'Simulate'}
              </button>
              {removeMutation.isError && (
                <p className="text-xs text-red-600">{(removeMutation.error as Error).message}</p>
              )}
            </div>
            {removeResult && (
              <>
                <SimDeltaCard result={removeResult} />
                {removeResult.affected_items && removeResult.affected_items.length > 0 && (
                  <div className="mt-2 space-y-1">
                    <p className="text-xs font-semibold text-slate-300">Affected Ingredients</p>
                    {removeResult.affected_items.map((ai) => (
                      <div key={ai.menu_item_id} className="flex items-center justify-between text-xs">
                        <span className="text-slate-200">{ai.name}</span>
                        <StatusBadge variant={ai.shared ? 'info' : 'warning'}>
                          {ai.shared ? 'Shared' : 'Exclusive'}
                        </StatusBadge>
                      </div>
                    ))}
                  </div>
                )}
              </>
            )}
          </div>

          {/* Card 3: Ingredient Cost Change */}
          <div className="bg-white/5 rounded-xl border border-white/10 shadow-sm p-5 space-y-4">
            <div>
              <h3 className="text-base font-semibold text-white">Ingredient Cost Change</h3>
              <p className="text-xs text-slate-400 mt-0.5">Model the margin impact of a supplier price change.</p>
            </div>
            <div className="space-y-3">
              <div>
                <label className="block text-xs font-medium text-slate-300 mb-1">Ingredient</label>
                <select
                  value={costIngredientId}
                  onChange={(e) => setCostIngredientId(e.target.value)}
                  className="w-full text-sm border border-white/20 rounded-md px-3 py-1.5 bg-white/10 text-white focus:outline-none focus:ring-2 focus:ring-[#F97316]"
                >
                  <option value="">Select ingredient…</option>
                  {allDeps.map((d) => (
                    <option key={d.ingredient_id} value={d.ingredient_id}>{d.ingredient_name}</option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-xs font-medium text-slate-300 mb-1">New Cost per Unit ($)</label>
                <input
                  type="number"
                  min="0"
                  step="0.01"
                  value={newCost}
                  onChange={(e) => setNewCost(e.target.value)}
                  placeholder="e.g. 2.50"
                  className="w-full text-sm border border-white/20 rounded-md px-3 py-1.5 bg-white/10 text-white focus:outline-none focus:ring-2 focus:ring-[#F97316]"
                />
              </div>
              <button
                disabled={!costIngredientId || !newCost || costMutation.isPending}
                onClick={async () => {
                  const result = await costMutation.mutateAsync({
                    locationId,
                    ingredientId: costIngredientId,
                    newCostPerUnit: parseFloat(newCost),
                  });
                  setCostResult(result);
                }}
                className="w-full px-4 py-2 text-sm font-medium rounded-lg bg-[#F97316] text-white hover:bg-orange-600 disabled:opacity-50 transition-colors"
              >
                {costMutation.isPending ? 'Simulating…' : 'Simulate'}
              </button>
              {costMutation.isError && (
                <p className="text-xs text-red-600">{(costMutation.error as Error).message}</p>
              )}
            </div>
            {costResult && (
              <>
                <SimDeltaCard result={costResult} />
                {costResult.affected_items && costResult.affected_items.length > 0 && (
                  <div className="mt-2 space-y-1">
                    <p className="text-xs font-semibold text-slate-300">Affected Items</p>
                    {costResult.affected_items.map((ai) => (
                      <div key={ai.menu_item_id} className="flex items-center justify-between text-xs">
                        <span className="text-slate-200">{ai.name}</span>
                        {ai.margin_delta !== undefined && (
                          <span className={deltaCls(ai.margin_delta)}>
                            {deltaLabel(ai.margin_delta)}
                          </span>
                        )}
                      </div>
                    ))}
                  </div>
                )}
              </>
            )}
          </div>
        </div>
      )}

      {/* ── Tab 3: Dependencies ── */}
      {activeTab === 'Dependencies' && (
        <div className="space-y-4">
          {depsError && (
            <ErrorBanner
              message={depsError instanceof Error ? depsError.message : 'Failed to load dependencies'}
            />
          )}
          <div className="flex items-center gap-3">
            <p className="text-sm text-slate-400">
              Ingredients sorted by menu item usage. Items highlighted in red are single points of failure (SPOF).
            </p>
          </div>
          {depsLoading ? (
            <div className="flex justify-center py-12"><LoadingSpinner /></div>
          ) : (
            <DataTable
              columns={depColumns}
              data={sortedDeps}
              keyExtractor={(r) => r.ingredient_id}
              emptyTitle="No dependency data"
              emptyDescription="No ingredient dependency data is available for this location."
            />
          )}
        </div>
      )}

      {/* ── Tab 4: Cross-Sell ── */}
      {activeTab === 'Cross-Sell' && (
        <div className="space-y-4">
          {crossError && (
            <ErrorBanner
              message={crossError instanceof Error ? crossError.message : 'Failed to load cross-sell data'}
            />
          )}
          <p className="text-sm text-slate-400">
            Top item pairs most frequently ordered together. Affinity bar is normalized to the highest pair.
          </p>
          {crossLoading ? (
            <div className="flex justify-center py-12"><LoadingSpinner /></div>
          ) : (
            <DataTable
              columns={crossColumns}
              data={allPairs}
              keyExtractor={(r) => `${r.item_a_name}__${r.item_b_name}`}
              emptyTitle="No cross-sell data"
              emptyDescription="No co-occurrence data found for this location."
            />
          )}
        </div>
      )}
    </div>
  );
}
