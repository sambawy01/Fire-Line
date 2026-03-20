import { useState } from 'react';
import {
  FileDown,
  Download,
  DollarSign,
  Percent,
  ShoppingBag,
  Clock,
  Bell,
  AlertCircle,
  FileText,
} from 'lucide-react';
import { useLocationStore } from '../stores/location';
import { useDailyReport } from '../hooks/useReports';
import { reportsApi } from '../lib/api';
import type { ReportChannel, ReportMenuItem, CategoryRevData, StaffEntry, ReorderItem, CriticalIssue } from '../lib/api';
import KPICard from '../components/ui/KPICard';
import DataTable from '../components/ui/DataTable';
import type { Column } from '../components/ui/DataTable';
import StatusBadge from '../components/ui/StatusBadge';
import LoadingSpinner from '../components/ui/LoadingSpinner';
import ErrorBanner from '../components/ui/ErrorBanner';

function cents(v: number): string {
  return `$${(v / 100).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
}

function formatDate(iso: string): string {
  try {
    return new Date(iso).toLocaleDateString('en-US', {
      weekday: 'long',
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    });
  } catch {
    return iso;
  }
}

function formatTimestamp(iso: string): string {
  try {
    return new Date(iso).toLocaleString('en-US', {
      month: 'short',
      day: 'numeric',
      hour: 'numeric',
      minute: '2-digit',
    });
  } catch {
    return iso;
  }
}

const CHANNEL_LABELS: Record<string, string> = {
  dine_in: 'Dine-in',
  takeout: 'Takeout',
  delivery: 'Delivery',
  drive_thru: 'Drive-Thru',
  catering: 'Catering',
  online: 'Online',
};

function healthScoreColor(score: number): string {
  if (score >= 70) return 'text-emerald-600';
  if (score >= 40) return 'text-amber-500';
  return 'text-red-600';
}

function healthScoreBg(score: number): string {
  if (score >= 70) return 'bg-emerald-50 border-emerald-200';
  if (score >= 40) return 'bg-amber-50 border-amber-200';
  return 'bg-red-50 border-red-200';
}

const channelColumns: Column<ReportChannel>[] = [
  {
    key: 'channel',
    header: 'Channel',
    sortable: true,
    render: (r) => CHANNEL_LABELS[r.channel] ?? r.channel,
  },
  {
    key: 'orders',
    header: 'Orders',
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
    key: 'pct_of_total',
    header: '% of Total',
    align: 'right',
    sortable: true,
    render: (r) => `${(r.pct_of_total ?? 0).toFixed(1)}%`,
  },
  {
    key: 'avg_ticket_time',
    header: 'Avg Ticket',
    align: 'right',
    sortable: true,
    render: (r) => `${(r.avg_ticket_time ?? 0).toFixed(1)} min`,
  },
];

const topItemColumns: Column<ReportMenuItem>[] = [
  { key: 'name', header: 'Item', sortable: true },
  { key: 'category', header: 'Category' },
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
    key: 'margin_pct',
    header: 'Margin',
    align: 'right',
    sortable: true,
    render: (r) => `${(r.margin_pct ?? 0).toFixed(1)}%`,
  },
];

const categoryColumns: Column<CategoryRevData>[] = [
  { key: 'category', header: 'Category', sortable: true },
  {
    key: 'revenue',
    header: 'Revenue',
    align: 'right',
    sortable: true,
    render: (r) => cents(r.revenue),
  },
  {
    key: 'pct_of_total',
    header: '% of Total',
    align: 'right',
    sortable: true,
    render: (r) => `${(r.pct_of_total ?? 0).toFixed(1)}%`,
  },
  {
    key: 'item_count',
    header: 'Items',
    align: 'right',
    sortable: true,
  },
];

const staffColumns: Column<StaffEntry>[] = [
  {
    key: 'name',
    header: 'Name',
    sortable: true,
    render: (r) =>
      r.is_overtime ? (
        <span className="text-red-600 font-medium">{r.name}</span>
      ) : (
        <span>{r.name}</span>
      ),
  },
  { key: 'role', header: 'Role' },
  {
    key: 'hours_worked',
    header: 'Hours',
    align: 'right',
    sortable: true,
    render: (r) => (r.hours_worked ?? 0).toFixed(1),
  },
  {
    key: 'labor_cost',
    header: 'Cost',
    align: 'right',
    sortable: true,
    render: (r) => cents(r.labor_cost),
  },
  {
    key: 'is_overtime',
    header: 'Overtime',
    align: 'center',
    render: (r) =>
      r.is_overtime ? (
        <StatusBadge variant="warning">OT</StatusBadge>
      ) : null,
  },
];

const reorderColumns: Column<ReorderItem>[] = [
  { key: 'name', header: 'Item', sortable: true },
  {
    key: 'current_level',
    header: 'Current',
    align: 'right',
    sortable: true,
    render: (r) => `${r.current_level} ${r.unit}`,
  },
  {
    key: 'par_level',
    header: 'PAR Level',
    align: 'right',
    sortable: true,
    render: (r) => `${r.par_level} ${r.unit}`,
  },
  {
    key: 'unit',
    header: 'Unit',
  },
];

export default function ReportsPage() {
  const locationId = useLocationStore((s) => s.selectedLocationId);
  const locations = useLocationStore((s) => s.locations);
  const [pdfLoading, setPdfLoading] = useState(false);
  const [pdfError, setPdfError] = useState<string | null>(null);

  const { data: report, isLoading, error, refetch } = useDailyReport(locationId);

  if (!locationId) return <LoadingSpinner fullPage />;

  const locationName = locations.find((l) => l.id === locationId)?.name ?? report?.location_name ?? '';
  const errorMessage = error instanceof Error ? error.message : 'Failed to load report data';

  async function handleDownloadPdf() {
    if (!locationId) return;
    setPdfError(null);
    setPdfLoading(true);
    try {
      await reportsApi.downloadPdf(locationId);
    } catch (e) {
      setPdfError(e instanceof Error ? e.message : 'PDF download failed');
    } finally {
      setPdfLoading(false);
    }
  }

  function handleExportJson() {
    if (!report) return;
    const blob = new Blob([JSON.stringify(report, null, 2)], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `fireline-daily-report-${report.report_date}.json`;
    a.click();
    URL.revokeObjectURL(url);
  }

  const score = report?.health_score ?? 0;

  return (
    <div className="space-y-8">
      {/* Page Header */}
      <div className="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-gray-800">Daily Report</h1>
          {locationName && (
            <p className="text-sm text-gray-500 mt-0.5">{locationName}</p>
          )}
          {report?.report_date && (
            <p className="text-sm text-gray-400 mt-0.5">{formatDate(report.report_date)}</p>
          )}
        </div>
        <div className="flex items-center gap-3 shrink-0">
          <button
            onClick={handleDownloadPdf}
            disabled={pdfLoading || !report}
            className="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-[#F97316] text-white text-sm font-medium hover:bg-orange-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            <FileDown className="h-4 w-4" />
            {pdfLoading ? 'Downloading…' : 'Download PDF'}
          </button>
          <button
            onClick={handleExportJson}
            disabled={!report}
            className="inline-flex items-center gap-2 px-4 py-2 rounded-lg border border-gray-300 bg-white text-gray-700 text-sm font-medium hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            <Download className="h-4 w-4" />
            Export JSON
          </button>
        </div>
      </div>

      {/* Error banners */}
      {error && <ErrorBanner message={errorMessage} retry={() => void refetch()} />}
      {pdfError && <ErrorBanner message={pdfError} />}

      {isLoading ? (
        <div className="flex justify-center py-16">
          <LoadingSpinner size="lg" />
        </div>
      ) : report ? (
        <>
          {/* Health Score Banner */}
          <div className={`rounded-xl border p-8 flex flex-col items-center text-center ${healthScoreBg(score)}`}>
            <p className="text-sm font-medium text-gray-500 mb-1">Operational Health Score</p>
            <span className={`text-5xl font-bold ${healthScoreColor(score)}`}>
              {score}
            </span>
            <p className="text-xs text-gray-400 mt-2">out of 100</p>
          </div>

          {/* Critical Issues */}
          {(report.critical_count ?? 0) > 0 && (
            <div className="rounded-xl border border-red-200 bg-red-50 p-6">
              <div className="flex items-center gap-2 mb-4">
                <AlertCircle className="h-5 w-5 text-red-600" />
                <h2 className="text-base font-semibold text-red-800">
                  Critical Issues ({report.critical_count ?? 0})
                </h2>
              </div>
              <ul className="space-y-3">
                {(report.critical_issues ?? []).map((issue: CriticalIssue, i: number) => (
                  <li key={i} className="flex items-start justify-between gap-4 text-sm">
                    <span className="text-red-700 font-medium">{issue.title}</span>
                    <div className="flex items-center gap-2 shrink-0">
                      <StatusBadge variant="neutral">{issue.module}</StatusBadge>
                      <span className="text-gray-400 text-xs">{formatTimestamp(issue.created_at)}</span>
                    </div>
                  </li>
                ))}
              </ul>
            </div>
          )}

          {/* KPI Cards */}
          <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-7 gap-4">
            <KPICard
              label="Net Revenue"
              value={cents(report.net_revenue)}
              icon={DollarSign}
              iconColor="text-emerald-600"
              bgTint="bg-emerald-50"
            />
            <KPICard
              label="Gross Margin"
              value={`${(report.gross_margin_pct ?? 0).toFixed(1)}%`}
              icon={Percent}
              iconColor="text-blue-600"
              bgTint="bg-blue-50"
            />
            <KPICard
              label="Labor Cost"
              value={`${(report.labor_cost_pct ?? 0).toFixed(1)}%`}
              icon={Percent}
              iconColor="text-red-600"
              bgTint="bg-red-50"
            />
            <KPICard
              label="Orders Today"
              value={String(report.orders_today)}
              icon={ShoppingBag}
              iconColor="text-blue-600"
              bgTint="bg-blue-50"
            />
            <KPICard
              label="Avg Ticket Time"
              value={`${(report.avg_ticket_time ?? 0).toFixed(1)} min`}
              icon={Clock}
              iconColor="text-purple-600"
              bgTint="bg-purple-50"
            />
            <KPICard
              label="Active Alerts"
              value={String(report.active_alerts)}
              icon={Bell}
              iconColor="text-orange-600"
              bgTint="bg-orange-50"
            />
            <KPICard
              label="Critical Issues"
              value={String(report.critical_count)}
              icon={AlertCircle}
              iconColor="text-red-600"
              bgTint="bg-red-50"
            />
          </div>

          {/* Channel Breakdown */}
          <div>
            <h2 className="text-lg font-semibold text-gray-800 mb-3">Channel Breakdown</h2>
            <DataTable
              columns={channelColumns}
              data={report.channels ?? []}
              keyExtractor={(r) => r.channel}
              emptyTitle="No channel data"
              emptyDescription="No channel data is available for this report."
            />
          </div>

          {/* Menu Performance */}
          <div className="space-y-4">
            <h2 className="text-lg font-semibold text-gray-800">Menu Performance</h2>

            {/* Top Performers */}
            {(report.top_items ?? []).length > 0 && (
              <div>
                <h3 className="text-sm font-semibold text-gray-600 uppercase tracking-wider mb-2">
                  Top Performers
                </h3>
                <DataTable
                  columns={topItemColumns}
                  data={report.top_items ?? []}
                  keyExtractor={(r) => r.name}
                  emptyTitle="No top items"
                />
              </div>
            )}

            {/* Underperformer */}
            {report.worst_item && (
              <div className="rounded-xl border border-amber-200 bg-amber-50 p-5">
                <h3 className="text-sm font-semibold text-amber-700 uppercase tracking-wider mb-2">
                  Underperformer
                </h3>
                <div className="flex flex-wrap items-center gap-6 text-sm">
                  <div>
                    <span className="text-gray-500">Item: </span>
                    <span className="font-medium text-gray-800">{report.worst_item.name}</span>
                  </div>
                  <div>
                    <span className="text-gray-500">Category: </span>
                    <span className="text-gray-700">{report.worst_item.category}</span>
                  </div>
                  <div>
                    <span className="text-gray-500">Units Sold: </span>
                    <span className="text-gray-700">{report.worst_item.units_sold}</span>
                  </div>
                  <div>
                    <span className="text-gray-500">Revenue: </span>
                    <span className="text-gray-700">{cents(report.worst_item.revenue)}</span>
                  </div>
                  <div>
                    <span className="text-gray-500">Margin: </span>
                    <span className="text-amber-700 font-medium">{(report.worst_item.margin_pct ?? 0).toFixed(1)}%</span>
                  </div>
                </div>
              </div>
            )}

            {/* Zero Sales Items */}
            {(report.zero_sales_items ?? []).length > 0 && (
              <div className="rounded-xl border border-amber-200 bg-amber-50 p-5">
                <h3 className="text-sm font-semibold text-amber-700 uppercase tracking-wider mb-3">
                  Zero Sales Items ({(report.zero_sales_items ?? []).length})
                </h3>
                <ul className="flex flex-wrap gap-2">
                  {(report.zero_sales_items ?? []).map((name: string) => (
                    <li key={name}>
                      <StatusBadge variant="warning">{name}</StatusBadge>
                    </li>
                  ))}
                </ul>
              </div>
            )}
          </div>

          {/* Category Revenue */}
          <div>
            <h2 className="text-lg font-semibold text-gray-800 mb-3">Category Revenue</h2>
            <DataTable
              columns={categoryColumns}
              data={report.category_revenue ?? []}
              keyExtractor={(r) => r.category}
              emptyTitle="No category data"
              emptyDescription="No category revenue data is available for this report."
            />
          </div>

          {/* Staff Summary */}
          <div>
            <h2 className="text-lg font-semibold text-gray-800 mb-3">Staff Summary</h2>
            <DataTable
              columns={staffColumns}
              data={report.staff_summary ?? []}
              keyExtractor={(r) => r.name}
              emptyTitle="No staff data"
              emptyDescription="No staff data is available for this report."
            />
            {(report.staff_summary ?? []).length > 0 && (
              <div className="mt-3 flex flex-wrap gap-6 text-sm text-gray-600 px-1">
                <span>
                  Total Hours:{' '}
                  <span className="font-semibold text-gray-800">
                    {(report.total_hours_worked ?? 0).toFixed(1)} hrs
                  </span>
                </span>
                <span>
                  Total Labor Cost:{' '}
                  <span className="font-semibold text-gray-800">{cents(report.total_labor_cost)}</span>
                </span>
                {(report.overtime_flags ?? []).length > 0 && (
                  <span className="text-amber-600">
                    Overtime employees:{' '}
                    <span className="font-semibold">{(report.overtime_flags ?? []).join(', ')}</span>
                  </span>
                )}
              </div>
            )}
          </div>

          {/* Inventory Alerts */}
          {(report.reorder_needed ?? []).length > 0 && (
            <div>
              <div className="flex items-center gap-2 mb-3">
                <h2 className="text-lg font-semibold text-gray-800">Inventory Alerts</h2>
                <StatusBadge variant="warning">
                  {(report.reorder_needed ?? []).length} items need reorder
                </StatusBadge>
              </div>
              <DataTable
                columns={reorderColumns}
                data={report.reorder_needed ?? []}
                keyExtractor={(r) => r.name}
                emptyTitle="No reorder needed"
              />
            </div>
          )}

          {/* Footer metadata */}
          <div className="flex items-center gap-2 text-xs text-gray-400 pt-2 border-t border-gray-100">
            <FileText className="h-3.5 w-3.5" />
            <span>Report generated for {report.report_date} — {locationName}</span>
          </div>
        </>
      ) : null}
    </div>
  );
}
