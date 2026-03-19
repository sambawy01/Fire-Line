import { useState } from 'react';
import {
  ScatterChart,
  Scatter,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  ReferenceLine,
  Cell,
} from 'recharts';
import { useLocationStore } from '../stores/location';
import { useMenuItems, useMenuSummary } from '../hooks/useMenu';
import KPICard from '../components/ui/KPICard';
import DataTable from '../components/ui/DataTable';
import type { Column } from '../components/ui/DataTable';
import StatusBadge from '../components/ui/StatusBadge';
import LoadingSpinner from '../components/ui/LoadingSpinner';
import ErrorBanner from '../components/ui/ErrorBanner';
import Modal from '../components/ui/Modal';
import type { MenuItemAnalysis } from '../lib/api';
import { LayoutList, Star, TrendingDown, Percent } from 'lucide-react';

function cents(v: number): string {
  return `$${(v / 100).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
}

const CHANNEL_LABELS: Record<string, string> = {
  dine_in: 'Dine-in',
  takeout: 'Takeout',
  delivery: 'Delivery',
  drive_thru: 'Drive-Thru',
};

const CLASSIFICATION_COLOR: Record<MenuItemAnalysis['classification'], string> = {
  powerhouse: '#10B981',   // emerald-500
  hidden_gem: '#3B82F6',   // blue-500
  crowd_pleaser: '#F59E0B', // amber-500
  underperformer: '#EF4444', // red-500
};

const CLASSIFICATION_BADGE: Record<
  MenuItemAnalysis['classification'],
  { variant: 'success' | 'info' | 'warning' | 'critical'; label: string }
> = {
  powerhouse:    { variant: 'success',  label: 'Powerhouse' },
  hidden_gem:    { variant: 'info',     label: 'Hidden Gem' },
  crowd_pleaser: { variant: 'warning',  label: 'Crowd Pleaser' },
  underperformer:{ variant: 'critical', label: 'Underperformer' },
};

const LEGEND_ITEMS: Array<{ classification: MenuItemAnalysis['classification']; label: string }> = [
  { classification: 'powerhouse',    label: 'Powerhouse' },
  { classification: 'hidden_gem',    label: 'Hidden Gem' },
  { classification: 'crowd_pleaser', label: 'Crowd Pleaser' },
  { classification: 'underperformer',label: 'Underperformer' },
];

export default function MenuPage() {
  const locationId = useLocationStore((s) => s.selectedLocationId);
  const [categoryFilter, setCategoryFilter] = useState<string>('all');
  const [selectedItem, setSelectedItem] = useState<MenuItemAnalysis | null>(null);

  const {
    data: itemsData,
    isLoading: itemsLoading,
    error: itemsError,
    refetch: refetchItems,
  } = useMenuItems(locationId);

  const {
    data: summary,
    isLoading: summaryLoading,
  } = useMenuSummary(locationId);

  if (!locationId) return <LoadingSpinner fullPage />;

  const allItems = itemsData?.items ?? [];

  // Collect unique categories for the filter dropdown
  const categories = Array.from(new Set(allItems.map((i) => i.category))).sort();

  const filteredItems =
    categoryFilter === 'all'
      ? allItems
      : allItems.filter((i) => i.category === categoryFilter);

  // Scatter plot data
  const scatterData = filteredItems.map((item) => ({
    x: item.popularity_pct,
    y: item.contrib_margin_pct,
    classification: item.classification,
    name: item.name,
  }));

  const medianMargin =
    summary?.avg_margin_pct ?? (allItems.length > 0
      ? allItems.reduce((acc, i) => acc + i.contrib_margin_pct, 0) / allItems.length
      : 0);

  // Channel breakdown columns (used inside modal)
  const channelColumns: Column<MenuItemAnalysis['by_channel'][number]>[] = [
    {
      key: 'channel',
      header: 'Channel',
      render: (r) => CHANNEL_LABELS[r.channel] ?? r.channel,
    },
    {
      key: 'units_sold',
      header: 'Units',
      align: 'right',
      sortable: true,
    },
    {
      key: 'revenue',
      header: 'Revenue',
      align: 'right',
      sortable: true,
      render: (r) => cents(r.revenue),
    },
    {
      key: 'commission',
      header: 'Commission',
      align: 'right',
      render: (r) => cents(r.commission),
    },
    {
      key: 'food_cost',
      header: 'Food Cost',
      align: 'right',
      render: (r) => cents(r.food_cost),
    },
    {
      key: 'margin',
      header: 'Margin ($)',
      align: 'right',
      sortable: true,
      render: (r) => cents(r.margin),
    },
    {
      key: 'margin_pct',
      header: 'Margin %',
      align: 'right',
      sortable: true,
      render: (r) => `${r.margin_pct.toFixed(1)}%`,
    },
  ];

  // Main item table columns — must be inside component to reference setSelectedItem
  const itemColumns: Column<MenuItemAnalysis>[] = [
    { key: 'name', header: 'Item', sortable: true },
    { key: 'category', header: 'Category', sortable: true },
    {
      key: 'price',
      header: 'Price',
      align: 'right',
      sortable: true,
      render: (r) => cents(r.price),
    },
    {
      key: 'units_sold',
      header: 'Units Sold',
      align: 'right',
      sortable: true,
    },
    {
      key: 'food_cost',
      header: 'Food Cost',
      align: 'right',
      render: (r) => cents(r.food_cost),
    },
    {
      key: 'contrib_margin',
      header: 'Margin ($)',
      align: 'right',
      sortable: true,
      render: (r) => cents(r.contrib_margin),
    },
    {
      key: 'contrib_margin_pct',
      header: 'Margin %',
      align: 'right',
      sortable: true,
      render: (r) => `${r.contrib_margin_pct.toFixed(1)}%`,
    },
    {
      key: 'popularity_pct',
      header: 'Popularity',
      align: 'right',
      sortable: true,
      render: (r) => `${r.popularity_pct.toFixed(1)}%`,
    },
    {
      key: 'classification',
      header: 'Class',
      align: 'center',
      render: (r) => {
        const cfg = CLASSIFICATION_BADGE[r.classification];
        return <StatusBadge variant={cfg.variant}>{cfg.label}</StatusBadge>;
      },
    },
    {
      key: 'detail',
      header: '',
      align: 'center',
      render: (r) => (
        <button
          onClick={() => setSelectedItem(r)}
          className="text-xs font-medium text-[#F97316] hover:underline focus:outline-none"
        >
          Detail
        </button>
      ),
    },
  ];

  return (
    <div className="space-y-8">
      {/* Page header */}
      <div>
        <h1 className="text-2xl font-bold text-gray-800">Menu Intelligence</h1>
        <p className="text-sm text-gray-500 mt-1">
          Item performance, margin analysis, and channel breakdown
        </p>
      </div>

      {itemsError && (
        <ErrorBanner
          message={
            itemsError instanceof Error
              ? itemsError.message
              : 'Failed to load menu data'
          }
          retry={() => refetchItems()}
        />
      )}

      {/* KPI Cards */}
      {summaryLoading || itemsLoading ? (
        <div className="flex justify-center py-8">
          <LoadingSpinner />
        </div>
      ) : (
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
          <KPICard
            label="Menu Items"
            value={String(summary?.total_items ?? allItems.length)}
            icon={LayoutList}
            iconColor="text-blue-600"
            bgTint="bg-blue-50"
          />
          <KPICard
            label="Avg Margin %"
            value={`${(summary?.avg_margin_pct ?? 0).toFixed(1)}%`}
            icon={Percent}
            iconColor="text-purple-600"
            bgTint="bg-purple-50"
          />
          <KPICard
            label="Powerhouses"
            value={String(summary?.powerhouse_count ?? 0)}
            icon={Star}
            iconColor="text-emerald-600"
            bgTint="bg-emerald-50"
          />
          <KPICard
            label="Underperformers"
            value={String(summary?.underperform_count ?? 0)}
            icon={TrendingDown}
            iconColor="text-red-600"
            bgTint="bg-red-50"
          />
        </div>
      )}

      {/* Scatter plot */}
      {filteredItems.length > 0 && (
        <div className="bg-white rounded-xl border border-gray-200 p-6 shadow-sm">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-semibold text-gray-800">
              Popularity vs. Margin
            </h2>
            {/* Legend */}
            <div className="flex flex-wrap gap-4">
              {LEGEND_ITEMS.map(({ classification, label }) => (
                <span key={classification} className="flex items-center gap-1.5 text-xs text-gray-600">
                  <span
                    className="inline-block h-3 w-3 rounded-full"
                    style={{ backgroundColor: CLASSIFICATION_COLOR[classification] }}
                  />
                  {label}
                </span>
              ))}
            </div>
          </div>
          <div className="h-72">
            <ResponsiveContainer width="100%" height="100%">
              <ScatterChart margin={{ top: 10, right: 20, left: 0, bottom: 10 }}>
                <CartesianGrid strokeDasharray="3 3" stroke="#E5E7EB" />
                <XAxis
                  type="number"
                  dataKey="x"
                  name="Popularity"
                  tick={{ fontSize: 12 }}
                  label={{ value: 'Popularity %', position: 'insideBottom', offset: -4, fontSize: 12 }}
                />
                <YAxis
                  type="number"
                  dataKey="y"
                  name="Margin %"
                  tick={{ fontSize: 12 }}
                  label={{ value: 'Margin %', angle: -90, position: 'insideLeft', offset: 10, fontSize: 12 }}
                />
                <Tooltip
                  cursor={{ strokeDasharray: '3 3' }}
                  content={({ payload }) => {
                    if (!payload?.length) return null;
                    const d = payload[0].payload as typeof scatterData[number];
                    return (
                      <div className="bg-white border border-gray-200 rounded-lg shadow-md px-3 py-2 text-xs">
                        <p className="font-semibold text-gray-800 mb-1">{d.name}</p>
                        <p className="text-gray-500">Popularity: {d.x.toFixed(1)}%</p>
                        <p className="text-gray-500">Margin: {d.y.toFixed(1)}%</p>
                      </div>
                    );
                  }}
                />
                <ReferenceLine
                  y={medianMargin}
                  stroke="#94A3B8"
                  strokeDasharray="4 4"
                  label={{ value: 'Avg margin', position: 'insideTopRight', fontSize: 11, fill: '#94A3B8' }}
                />
                <Scatter data={scatterData} isAnimationActive={false}>
                  {scatterData.map((entry, index) => (
                    <Cell
                      key={`cell-${index}`}
                      fill={CLASSIFICATION_COLOR[entry.classification]}
                    />
                  ))}
                </Scatter>
              </ScatterChart>
            </ResponsiveContainer>
          </div>
        </div>
      )}

      {/* Category filter + table */}
      <div>
        <div className="flex items-center justify-between mb-3">
          <h2 className="text-lg font-semibold text-gray-800">Item Detail</h2>
          {categories.length > 0 && (
            <select
              value={categoryFilter}
              onChange={(e) => setCategoryFilter(e.target.value)}
              className="text-sm border border-gray-300 rounded-md px-3 py-1.5 bg-white text-gray-700 focus:outline-none focus:ring-2 focus:ring-[#F97316]"
            >
              <option value="all">All Categories</option>
              {categories.map((cat) => (
                <option key={cat} value={cat}>
                  {cat}
                </option>
              ))}
            </select>
          )}
        </div>

        <DataTable
          columns={itemColumns}
          data={filteredItems}
          keyExtractor={(r) => r.menu_item_id}
          isLoading={itemsLoading}
          emptyTitle="No menu items"
          emptyDescription="No items match the selected category filter."
        />
      </div>

      {/* Channel detail modal */}
      <Modal
        open={selectedItem !== null}
        onClose={() => setSelectedItem(null)}
        title={selectedItem ? `${selectedItem.name} — Channel Breakdown` : ''}
      >
        {selectedItem && (
          <div className="space-y-4">
            <div className="flex items-center gap-3">
              <StatusBadge variant={CLASSIFICATION_BADGE[selectedItem.classification].variant}>
                {CLASSIFICATION_BADGE[selectedItem.classification].label}
              </StatusBadge>
              <span className="text-sm text-gray-500">
                {selectedItem.category} &bull; {cents(selectedItem.price)}
              </span>
            </div>
            <DataTable
              columns={channelColumns}
              data={selectedItem.by_channel}
              keyExtractor={(r) => r.channel}
              emptyTitle="No channel data"
            />
          </div>
        )}
      </Modal>
    </div>
  );
}
