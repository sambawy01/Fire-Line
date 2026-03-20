import { useState } from 'react';
import {
  Building2,
  ChevronRight,
  ChevronDown,
  BarChart2,
  Lightbulb,
  GitCompare,
  RefreshCw,
  CheckCircle,
  XCircle,
  AlertTriangle,
} from 'lucide-react';
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from 'recharts';
import {
  useHierarchy,
  useBenchmarks,
  useOutliers,
  useBestPractices,
  useComparison,
  useCalculateBenchmarks,
  useAdoptPractice,
  useDismissPractice,
} from '../hooks/usePortfolio';
import type { PortfolioNode, LocationBenchmark, Outlier } from '../lib/api';
import LoadingSpinner from '../components/ui/LoadingSpinner';
import ErrorBanner from '../components/ui/ErrorBanner';
import EmptyState from '../components/ui/EmptyState';

// ── Helpers ──────────────────────────────────────────────────────────────────

function fmtCents(cents: number) {
  return `$${(cents / 100).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
}

function fmtPct(v: number) {
  return `${v.toFixed(1)}%`;
}

function percentileColor(p: number): string {
  if (p >= 75) return 'bg-green-100 text-green-800';
  if (p >= 25) return 'bg-yellow-100 text-yellow-800';
  return 'bg-red-100 text-red-800';
}

// ── Types ────────────────────────────────────────────────────────────────────

type Tab = 'hierarchy' | 'benchmarking' | 'practices' | 'comparison';

// ── Tree helpers ─────────────────────────────────────────────────────────────

interface TreeNode extends PortfolioNode {
  children: TreeNode[];
}

function buildTree(nodes: PortfolioNode[]): TreeNode[] {
  const map = new Map<string, TreeNode>();
  nodes.forEach((n) => map.set(n.node_id, { ...n, children: [] }));

  const roots: TreeNode[] = [];
  map.forEach((node) => {
    if (node.parent_node_id && map.has(node.parent_node_id)) {
      map.get(node.parent_node_id)!.children.push(node);
    } else {
      roots.push(node);
    }
  });
  return roots;
}

// ── NodeRow ──────────────────────────────────────────────────────────────────

function NodeRow({
  node,
  depth,
  onSelect,
  selected,
}: {
  node: TreeNode;
  depth: number;
  onSelect: (n: TreeNode) => void;
  selected: string | null;
}) {
  const [expanded, setExpanded] = useState(true);
  const hasChildren = node.children.length > 0;

  const typeBadgeColor: Record<string, string> = {
    org: 'bg-purple-100 text-purple-700',
    region: 'bg-blue-100 text-blue-700',
    district: 'bg-indigo-100 text-indigo-700',
    location: 'bg-orange-100 text-orange-700',
  };

  return (
    <>
      <div
        className={`flex items-center gap-2 px-3 py-2 rounded-md cursor-pointer transition-colors ${
          selected === node.node_id
            ? 'bg-[#F97316]/10 border border-[#F97316]/30'
            : 'hover:bg-gray-50'
        }`}
        style={{ paddingLeft: `${12 + depth * 20}px` }}
        onClick={() => onSelect(node)}
      >
        {hasChildren ? (
          <button
            className="shrink-0 text-gray-400 hover:text-gray-600"
            onClick={(e) => {
              e.stopPropagation();
              setExpanded(!expanded);
            }}
          >
            {expanded ? (
              <ChevronDown className="h-4 w-4" />
            ) : (
              <ChevronRight className="h-4 w-4" />
            )}
          </button>
        ) : (
          <span className="w-4 shrink-0" />
        )}

        <Building2 className="h-4 w-4 shrink-0 text-gray-400" />

        <span className="flex-1 text-sm font-medium text-gray-800 truncate">
          {node.name}
        </span>

        <span
          className={`text-xs px-1.5 py-0.5 rounded font-medium shrink-0 ${
            typeBadgeColor[node.node_type] ?? 'bg-gray-100 text-gray-600'
          }`}
        >
          {node.node_type}
        </span>
      </div>

      {expanded &&
        hasChildren &&
        node.children.map((child) => (
          <NodeRow
            key={child.node_id}
            node={child}
            depth={depth + 1}
            onSelect={onSelect}
            selected={selected}
          />
        ))}
    </>
  );
}

// ── Hierarchy Tab ─────────────────────────────────────────────────────────────

function HierarchyTab() {
  const { data, isLoading, error } = useHierarchy();
  const [selected, setSelected] = useState<TreeNode | null>(null);

  const now = new Date();
  const from = new Date(now.getFullYear(), now.getMonth(), 1).toISOString();
  const to = now.toISOString();

  // We can't conditionally fetch KPIs when selected changes without extra hooks.
  // Use a simple inline query triggered by selection.
  const kpiNodeId = selected?.node_id ?? null;
  // Use useComparison for the selected location node directly
  const { data: compData, isLoading: kpiLoading } = useComparison(
    selected?.location_id ? [selected.location_id] : [],
    from,
    to
  );

  if (isLoading) return <LoadingSpinner />;
  if (error) return <ErrorBanner message={(error as Error).message} />;

  const nodes = data?.nodes ?? [];
  const tree = buildTree(nodes);

  if (tree.length === 0) {
    return (
      <EmptyState
        title="No hierarchy configured"
        description="Create portfolio nodes to organize your locations into regions and districts."
      />
    );
  }

  const kpiLoc = compData?.locations?.[0];

  return (
    <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
      {/* Tree */}
      <div className="lg:col-span-2 bg-white rounded-lg border border-gray-200 p-4">
        <h3 className="text-sm font-semibold text-gray-700 mb-3">
          Portfolio Structure
        </h3>
        <div className="space-y-0.5">
          {tree.map((root) => (
            <NodeRow
              key={root.node_id}
              node={root}
              depth={0}
              onSelect={setSelected}
              selected={selected?.node_id ?? null}
            />
          ))}
        </div>
      </div>

      {/* KPI panel */}
      <div className="bg-white rounded-lg border border-gray-200 p-4">
        <h3 className="text-sm font-semibold text-gray-700 mb-3">
          {selected ? selected.name : 'Select a node'}
        </h3>
        {!selected && (
          <p className="text-sm text-gray-400">
            Click a location node to view its current-month KPIs.
          </p>
        )}
        {selected && selected.node_type !== 'location' && (
          <p className="text-sm text-gray-400">
            Select a location node to view KPIs.
          </p>
        )}
        {selected && selected.node_type === 'location' && kpiLoading && (
          <LoadingSpinner />
        )}
        {selected && selected.node_type === 'location' && kpiLoc && (
          <div className="space-y-3">
            <Metric label="Revenue" value={fmtCents(kpiLoc.revenue)} />
            <Metric label="Food Cost %" value={fmtPct(kpiLoc.food_cost_pct)} />
            <Metric label="Labor Cost %" value={fmtPct(kpiLoc.labor_cost_pct)} />
            <Metric label="Avg Check" value={fmtCents(kpiLoc.avg_check_cents)} />
            <Metric label="Check Count" value={kpiLoc.check_count.toLocaleString()} />
          </div>
        )}
        {selected && selected.node_type === 'location' && !kpiLoading && !kpiLoc && (
          <p className="text-sm text-gray-400">No data for current month.</p>
        )}
        {selected && kpiNodeId && (
          <p className="mt-4 text-xs text-gray-400">Node ID: {kpiNodeId}</p>
        )}
      </div>
    </div>
  );
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between py-1.5 border-b border-gray-100 last:border-0">
      <span className="text-sm text-gray-500">{label}</span>
      <span className="text-sm font-semibold text-gray-800">{value}</span>
    </div>
  );
}

// ── Benchmarking Tab ─────────────────────────────────────────────────────────

function BenchmarkingTab() {
  const now = new Date();
  const from = new Date(now.getFullYear(), now.getMonth(), 1).toISOString();
  const to = now.toISOString();

  const { data: bData, isLoading: bLoading } = useBenchmarks(from, to);
  const { data: oData, isLoading: oLoading } = useOutliers(from, to);
  const calcMutation = useCalculateBenchmarks();

  const benchmarks = bData?.benchmarks ?? [];
  const outlierSet = new Set(
    (oData?.outliers ?? []).map((o: Outlier) => `${o.location_id}:${o.metric}`)
  );

  const metrics: { key: keyof LocationBenchmark; label: string; fmt: (v: number) => string }[] = [
    { key: 'revenue_percentile', label: 'Revenue', fmt: fmtPct },
    { key: 'food_cost_percentile', label: 'Food Cost', fmt: fmtPct },
    { key: 'labor_cost_percentile', label: 'Labor Cost', fmt: fmtPct },
    { key: 'avg_check_percentile', label: 'Avg Check', fmt: fmtPct },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-base font-semibold text-gray-800">
            Location Benchmarks — Current Month
          </h3>
          <p className="text-sm text-gray-500 mt-0.5">
            Percentile ranks across all org locations. Green ≥ 75th, Yellow 25–74th, Red &lt; 25th.
          </p>
        </div>
        <button
          onClick={() => calcMutation.mutate({ from, to })}
          disabled={calcMutation.isPending}
          className="flex items-center gap-2 px-4 py-2 bg-[#F97316] text-white rounded-md text-sm font-medium hover:bg-orange-600 disabled:opacity-50 transition-colors"
        >
          <RefreshCw className={`h-4 w-4 ${calcMutation.isPending ? 'animate-spin' : ''}`} />
          {calcMutation.isPending ? 'Calculating…' : 'Calculate'}
        </button>
      </div>

      {(bLoading || oLoading) && <LoadingSpinner />}

      {!bLoading && benchmarks.length === 0 && (
        <EmptyState
          title="No benchmark data"
          description="Click Calculate to compute benchmarks for the current month."
        />
      )}

      {benchmarks.length > 0 && (
        <div className="bg-white rounded-lg border border-gray-200 overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200 text-sm">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left font-semibold text-gray-600">
                  Location
                </th>
                {metrics.map((m) => (
                  <th
                    key={m.key}
                    className="px-4 py-3 text-center font-semibold text-gray-600"
                  >
                    {m.label}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {benchmarks.map((b) => (
                <tr key={b.benchmark_id} className="hover:bg-gray-50">
                  <td className="px-4 py-3 font-medium text-gray-800">
                    {b.location_name}
                  </td>
                  {metrics.map((m) => {
                    const pct = b[m.key] as number;
                    const rawMetric = m.key.replace('_percentile', '');
                    const isOutlier = outlierSet.has(`${b.location_id}:${rawMetric}`);
                    return (
                      <td key={m.key} className="px-4 py-3 text-center">
                        <span
                          className={`inline-block px-2.5 py-1 rounded-full text-xs font-semibold ${percentileColor(pct)}`}
                        >
                          {pct.toFixed(0)}th
                        </span>
                        {isOutlier && (
                          <AlertTriangle className="inline-block ml-1 h-3.5 w-3.5 text-amber-500" />
                        )}
                      </td>
                    );
                  })}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Outliers summary */}
      {(oData?.outliers ?? []).length > 0 && (
        <div className="bg-amber-50 border border-amber-200 rounded-lg p-4">
          <h4 className="text-sm font-semibold text-amber-800 mb-2">
            Outliers Detected ({oData!.outliers.length})
          </h4>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
            {oData!.outliers.map((o, i) => (
              <div
                key={i}
                className="bg-white rounded-md border border-amber-100 px-3 py-2 text-xs"
              >
                <span className="font-semibold text-gray-800">{o.location_name}</span>
                <span className="text-gray-500 ml-1">
                  — {o.metric} is{' '}
                  <span className={o.direction === 'above' ? 'text-green-600' : 'text-red-600'}>
                    {o.direction}
                  </span>{' '}
                  normal range
                </span>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

// ── Best Practices Tab ────────────────────────────────────────────────────────

function BestPracticesTab() {
  const { data, isLoading, error } = useBestPractices();
  const adopt = useAdoptPractice();
  const dismiss = useDismissPractice();

  if (isLoading) return <LoadingSpinner />;
  if (error) return <ErrorBanner message={(error as Error).message} />;

  const practices = data?.best_practices ?? [];

  if (practices.length === 0) {
    return (
      <EmptyState
        title="No best practices detected"
        description="Calculate benchmarks first, then detect practices from the top-quartile locations."
      />
    );
  }

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
      {practices.map((p) => (
        <div
          key={p.practice_id}
          className="bg-white rounded-lg border border-gray-200 p-4 flex flex-col gap-3"
        >
          <div className="flex items-start gap-2">
            <Lightbulb className="h-5 w-5 text-amber-500 shrink-0 mt-0.5" />
            <div className="flex-1 min-w-0">
              <p className="text-sm font-semibold text-gray-800 leading-snug">
                {p.title}
              </p>
              {p.source_name && (
                <p className="text-xs text-gray-400 mt-0.5">
                  Source: {p.source_name}
                </p>
              )}
            </div>
            <span
              className={`text-xs px-2 py-0.5 rounded-full font-medium shrink-0 ${
                p.status === 'adopted'
                  ? 'bg-green-100 text-green-700'
                  : p.status === 'dismissed'
                  ? 'bg-gray-100 text-gray-500'
                  : 'bg-blue-100 text-blue-700'
              }`}
            >
              {p.status}
            </span>
          </div>

          <p className="text-xs text-gray-600 leading-relaxed">{p.description}</p>

          <div className="flex items-center justify-between mt-auto pt-2 border-t border-gray-100">
            <span className="text-xs font-medium text-emerald-600">
              +{p.impact_pct.toFixed(1)}% potential impact
            </span>
            {p.status === 'suggested' && (
              <div className="flex gap-2">
                <button
                  onClick={() => adopt.mutate(p.practice_id)}
                  disabled={adopt.isPending}
                  className="flex items-center gap-1 text-xs px-2 py-1 bg-green-600 text-white rounded hover:bg-green-700 disabled:opacity-50 transition-colors"
                >
                  <CheckCircle className="h-3.5 w-3.5" />
                  Adopt
                </button>
                <button
                  onClick={() => dismiss.mutate(p.practice_id)}
                  disabled={dismiss.isPending}
                  className="flex items-center gap-1 text-xs px-2 py-1 bg-gray-200 text-gray-600 rounded hover:bg-gray-300 disabled:opacity-50 transition-colors"
                >
                  <XCircle className="h-3.5 w-3.5" />
                  Dismiss
                </button>
              </div>
            )}
          </div>
        </div>
      ))}
    </div>
  );
}

// ── Comparison Tab ────────────────────────────────────────────────────────────

function ComparisonTab() {
  const { data: hierarchyData } = useHierarchy();
  const [selected, setSelected] = useState<string[]>([]);

  const nodes = hierarchyData?.nodes ?? [];
  const locationNodes = nodes.filter((n) => n.node_type === 'location' && n.location_id);

  const now = new Date();
  const from = new Date(now.getFullYear(), now.getMonth(), 1).toISOString();
  const to = now.toISOString();

  const { data: compData, isLoading } = useComparison(selected, from, to);
  const locations = compData?.locations ?? [];

  const revenueData = locations.map((l) => ({
    name: l.location_name || l.location_id.slice(0, 8),
    Revenue: Math.round(l.revenue / 100),
  }));

  const costData = locations.map((l) => ({
    name: l.location_name || l.location_id.slice(0, 8),
    'Food Cost %': parseFloat(l.food_cost_pct.toFixed(1)),
    'Labor Cost %': parseFloat(l.labor_cost_pct.toFixed(1)),
  }));

  const toggleLocation = (locId: string) => {
    setSelected((prev) =>
      prev.includes(locId) ? prev.filter((id) => id !== locId) : [...prev, locId]
    );
  };

  return (
    <div className="space-y-6">
      {/* Location selector */}
      <div className="bg-white rounded-lg border border-gray-200 p-4">
        <h3 className="text-sm font-semibold text-gray-700 mb-3">
          Select Locations to Compare
        </h3>
        {locationNodes.length === 0 ? (
          <p className="text-sm text-gray-400">
            No location nodes in hierarchy. Add location-type nodes first.
          </p>
        ) : (
          <div className="flex flex-wrap gap-2">
            {locationNodes.map((n) => {
              const locId = n.location_id!;
              const active = selected.includes(locId);
              return (
                <button
                  key={locId}
                  onClick={() => toggleLocation(locId)}
                  className={`text-sm px-3 py-1.5 rounded-full border font-medium transition-colors ${
                    active
                      ? 'bg-[#F97316] text-white border-[#F97316]'
                      : 'bg-white text-gray-600 border-gray-300 hover:border-[#F97316]'
                  }`}
                >
                  {n.name}
                </button>
              );
            })}
          </div>
        )}
      </div>

      {isLoading && selected.length > 0 && <LoadingSpinner />}

      {!isLoading && selected.length > 0 && locations.length === 0 && (
        <EmptyState title="No data" description="No check data found for the selected period." />
      )}

      {locations.length > 0 && (
        <>
          {/* Revenue chart */}
          <div className="bg-white rounded-lg border border-gray-200 p-4">
            <h4 className="text-sm font-semibold text-gray-700 mb-4">
              Revenue Comparison (Current Month)
            </h4>
            <ResponsiveContainer width="100%" height={260}>
              <BarChart data={revenueData} margin={{ top: 4, right: 16, left: 0, bottom: 4 }}>
                <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
                <XAxis dataKey="name" tick={{ fontSize: 12 }} />
                <YAxis tickFormatter={(v) => `$${(v / 1000).toFixed(0)}k`} tick={{ fontSize: 12 }} />
                <Tooltip formatter={(v) => [`$${Number(v).toLocaleString()}`, 'Revenue']} />
                <Bar dataKey="Revenue" fill="#F97316" radius={[4, 4, 0, 0]} />
              </BarChart>
            </ResponsiveContainer>
          </div>

          {/* Cost % chart */}
          <div className="bg-white rounded-lg border border-gray-200 p-4">
            <h4 className="text-sm font-semibold text-gray-700 mb-4">
              Cost % Comparison
            </h4>
            <ResponsiveContainer width="100%" height={260}>
              <BarChart data={costData} margin={{ top: 4, right: 16, left: 0, bottom: 4 }}>
                <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
                <XAxis dataKey="name" tick={{ fontSize: 12 }} />
                <YAxis tickFormatter={(v) => `${v}%`} tick={{ fontSize: 12 }} />
                <Tooltip formatter={(v) => [`${Number(v).toFixed(1)}%`]} />
                <Legend />
                <Bar dataKey="Food Cost %" fill="#3B82F6" radius={[4, 4, 0, 0]} />
                <Bar dataKey="Labor Cost %" fill="#8B5CF6" radius={[4, 4, 0, 0]} />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </>
      )}
    </div>
  );
}

// ── Main Page ─────────────────────────────────────────────────────────────────

const tabs: { id: Tab; label: string; icon: React.ElementType }[] = [
  { id: 'hierarchy', label: 'Hierarchy', icon: Building2 },
  { id: 'benchmarking', label: 'Benchmarking', icon: BarChart2 },
  { id: 'practices', label: 'Best Practices', icon: Lightbulb },
  { id: 'comparison', label: 'Comparison', icon: GitCompare },
];

export default function PortfolioPage() {
  const [activeTab, setActiveTab] = useState<Tab>('hierarchy');

  return (
    <div className="space-y-6">
      {/* Page header */}
      <div className="flex items-center gap-3">
        <Building2 className="h-7 w-7 text-[#F97316]" />
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Portfolio Intelligence</h1>
          <p className="text-sm text-gray-500">
            Multi-location hierarchy, benchmarking, and best practices
          </p>
        </div>
      </div>

      {/* Tabs */}
      <div className="border-b border-gray-200">
        <nav className="-mb-px flex gap-6">
          {tabs.map(({ id, label, icon: Icon }) => (
            <button
              key={id}
              onClick={() => setActiveTab(id)}
              className={`flex items-center gap-2 pb-3 text-sm font-medium border-b-2 transition-colors ${
                activeTab === id
                  ? 'border-[#F97316] text-[#F97316]'
                  : 'border-transparent text-gray-500 hover:text-gray-700'
              }`}
            >
              <Icon className="h-4 w-4" />
              {label}
            </button>
          ))}
        </nav>
      </div>

      {/* Tab content */}
      {activeTab === 'hierarchy' && <HierarchyTab />}
      {activeTab === 'benchmarking' && <BenchmarkingTab />}
      {activeTab === 'practices' && <BestPracticesTab />}
      {activeTab === 'comparison' && <ComparisonTab />}
    </div>
  );
}
