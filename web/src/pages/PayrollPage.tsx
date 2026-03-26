import { useState, useMemo } from 'react';
import { DollarSign, Clock, Users, Download, Briefcase } from 'lucide-react';
import { useLocationStore } from '../stores/location';
import { usePayrollSummary, usePayrollHistory } from '../hooks/usePayroll';
import { payrollApi } from '../lib/api';
import KPICard from '../components/ui/KPICard';
import LoadingSpinner from '../components/ui/LoadingSpinner';
import ErrorBanner from '../components/ui/ErrorBanner';
import type { PayrollEmployee } from '../lib/api';

// ── Helpers ─────────────────────────────────────────────────────────────────

function fmtEGP(cents: number): string {
  return `EGP ${(cents / 100).toLocaleString('en-US', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  })}`;
}

function fmtHours(h: number): string {
  return h.toLocaleString('en-US', { minimumFractionDigits: 1, maximumFractionDigits: 1 });
}

function getMonthRange(): { start: string; end: string } {
  const now = new Date();
  const start = new Date(now.getFullYear(), now.getMonth(), 1);
  const end = new Date(now.getFullYear(), now.getMonth() + 1, 0);
  return {
    start: start.toISOString().slice(0, 10),
    end: end.toISOString().slice(0, 10),
  };
}

// ── Component ───────────────────────────────────────────────────────────────

export default function PayrollPage() {
  const locationId = useLocationStore((s) => s.selectedLocationId);

  const defaultRange = useMemo(() => getMonthRange(), []);
  const [periodStart, setPeriodStart] = useState(defaultRange.start);
  const [periodEnd, setPeriodEnd] = useState(defaultRange.end);
  const [exporting, setExporting] = useState(false);

  const {
    data: summary,
    isLoading: summaryLoading,
    error: summaryError,
    refetch: refetchSummary,
  } = usePayrollSummary(locationId, periodStart, periodEnd);

  const {
    data: historyData,
    isLoading: historyLoading,
  } = usePayrollHistory(locationId);

  const sortedEmployees = useMemo(() => {
    if (!summary?.employees) return [];
    return [...summary.employees].sort((a, b) => b.gross_pay - a.gross_pay);
  }, [summary]);

  const history = historyData?.history ?? [];

  const maxGross = useMemo(() => {
    if (!history.length) return 1;
    return Math.max(...history.map((h) => h.gross_pay));
  }, [history]);

  const handleExport = async () => {
    if (!locationId) return;
    setExporting(true);
    try {
      await payrollApi.exportCsv(locationId, periodStart, periodEnd);
    } catch {
      // silently handle
    } finally {
      setExporting(false);
    }
  };

  if (!locationId) {
    return (
      <div className="flex items-center justify-center min-h-[60vh]">
        <p className="text-slate-400">Select a location to view payroll data.</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div className="flex items-center gap-3">
          <div className="bg-emerald-500/20 p-2.5 rounded-lg">
            <DollarSign className="h-6 w-6 text-emerald-400" />
          </div>
          <div>
            <h1 className="text-2xl font-bold text-white">Payroll</h1>
            <p className="text-sm text-slate-400">Employee compensation overview</p>
          </div>
        </div>

        <div className="flex flex-wrap items-center gap-3">
          <div className="flex items-center gap-2">
            <label className="text-xs text-slate-400" htmlFor="pay-start">From</label>
            <input
              id="pay-start"
              type="date"
              value={periodStart}
              onChange={(e) => setPeriodStart(e.target.value)}
              className="bg-white/5 border border-white/10 rounded-lg px-3 py-1.5 text-sm text-white focus:outline-none focus:ring-1 focus:ring-[#F97316]"
            />
          </div>
          <div className="flex items-center gap-2">
            <label className="text-xs text-slate-400" htmlFor="pay-end">To</label>
            <input
              id="pay-end"
              type="date"
              value={periodEnd}
              onChange={(e) => setPeriodEnd(e.target.value)}
              className="bg-white/5 border border-white/10 rounded-lg px-3 py-1.5 text-sm text-white focus:outline-none focus:ring-1 focus:ring-[#F97316]"
            />
          </div>
          <button
            onClick={handleExport}
            disabled={exporting}
            className="flex items-center gap-2 bg-[#F97316] hover:bg-[#EA580C] text-white text-sm font-medium px-4 py-2 rounded-lg transition-colors disabled:opacity-50"
          >
            <Download className="h-4 w-4" />
            {exporting ? 'Exporting...' : 'Export CSV'}
          </button>
        </div>
      </div>

      {/* Error */}
      {summaryError && (
        <ErrorBanner
          message={(summaryError as Error).message || 'Failed to load payroll data'}
          retry={() => refetchSummary()}
        />
      )}

      {/* Loading */}
      {summaryLoading && <LoadingSpinner fullPage />}

      {/* KPI Cards */}
      {summary && (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          <KPICard
            label="Total Gross Pay"
            value={fmtEGP(summary.total_gross_pay)}
            icon={DollarSign}
            iconColor="text-emerald-400"
            bgTint="bg-emerald-500/20"
          />
          <KPICard
            label="Total Overtime Pay"
            value={fmtEGP(summary.total_overtime_pay)}
            icon={Clock}
            iconColor="text-amber-400"
            bgTint="bg-amber-500/20"
          />
          <KPICard
            label="Total Hours"
            value={fmtHours(summary.total_hours)}
            icon={Briefcase}
            iconColor="text-blue-400"
            bgTint="bg-blue-500/20"
          />
          <KPICard
            label="Employee Count"
            value={String(summary.employee_count)}
            icon={Users}
            iconColor="text-purple-400"
            bgTint="bg-purple-500/20"
          />
        </div>
      )}

      {/* Empty state */}
      {summary && sortedEmployees.length === 0 && (
        <div className="bg-white/5 rounded-xl border border-white/10 p-12 text-center">
          <DollarSign className="mx-auto mb-3 h-10 w-10 text-slate-500" />
          <p className="text-lg font-medium text-slate-300">No payroll data for this period</p>
          <p className="mt-1 text-sm text-slate-500">Try adjusting the date range above.</p>
        </div>
      )}

      {/* Employee Payroll Table */}
      {sortedEmployees.length > 0 && (
        <div className="bg-white/5 rounded-xl border border-white/10 overflow-hidden">
          <div className="px-5 py-4 border-b border-white/10">
            <h2 className="text-lg font-semibold text-white">Employee Payroll</h2>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-sm" role="table">
              <thead>
                <tr className="border-b border-white/10 text-left">
                  <th className="px-5 py-3 text-xs font-medium text-slate-400 uppercase tracking-wider">Employee</th>
                  <th className="px-5 py-3 text-xs font-medium text-slate-400 uppercase tracking-wider">Role</th>
                  <th className="px-5 py-3 text-xs font-medium text-slate-400 uppercase tracking-wider text-right">Shifts</th>
                  <th className="px-5 py-3 text-xs font-medium text-slate-400 uppercase tracking-wider text-right">Regular Hrs</th>
                  <th className="px-5 py-3 text-xs font-medium text-slate-400 uppercase tracking-wider text-right">OT Hrs</th>
                  <th className="px-5 py-3 text-xs font-medium text-slate-400 uppercase tracking-wider text-right">Regular Pay</th>
                  <th className="px-5 py-3 text-xs font-medium text-slate-400 uppercase tracking-wider text-right">OT Pay</th>
                  <th className="px-5 py-3 text-xs font-medium text-slate-400 uppercase tracking-wider text-right">Gross Pay</th>
                </tr>
              </thead>
              <tbody>
                {sortedEmployees.map((emp: PayrollEmployee) => {
                  const hasOT = emp.overtime_hours > 0;
                  return (
                    <tr
                      key={emp.employee_id}
                      className={`border-b border-white/5 hover:bg-white/5 transition-colors ${
                        hasOT ? 'bg-amber-500/5' : ''
                      }`}
                    >
                      <td className="px-5 py-3 text-white font-medium">{emp.display_name}</td>
                      <td className="px-5 py-3 text-slate-400 capitalize">{emp.role}</td>
                      <td className="px-5 py-3 text-slate-300 text-right">{emp.shift_count}</td>
                      <td className="px-5 py-3 text-slate-300 text-right">{fmtHours(emp.regular_hours)}</td>
                      <td className={`px-5 py-3 text-right font-medium ${hasOT ? 'text-amber-400' : 'text-slate-300'}`}>
                        {fmtHours(emp.overtime_hours)}
                      </td>
                      <td className="px-5 py-3 text-slate-300 text-right">{fmtEGP(emp.regular_pay)}</td>
                      <td className={`px-5 py-3 text-right font-medium ${hasOT ? 'text-amber-400' : 'text-slate-300'}`}>
                        {fmtEGP(emp.overtime_pay)}
                      </td>
                      <td className="px-5 py-3 text-white text-right font-semibold">{fmtEGP(emp.gross_pay)}</td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Payroll History Chart */}
      {!historyLoading && history.length > 0 && (
        <div className="bg-white/5 rounded-xl border border-white/10 p-5">
          <h2 className="text-lg font-semibold text-white mb-4">Payroll History (Last 6 Months)</h2>

          <div className="relative" style={{ height: 280 }}>
            {/* Y-axis labels */}
            <div className="absolute left-0 top-0 bottom-8 w-16 flex flex-col justify-between text-xs text-slate-500 text-right pr-2">
              <span>{fmtEGPShort(maxGross)}</span>
              <span>{fmtEGPShort(maxGross * 0.75)}</span>
              <span>{fmtEGPShort(maxGross * 0.5)}</span>
              <span>{fmtEGPShort(maxGross * 0.25)}</span>
              <span>0</span>
            </div>

            {/* Chart area */}
            <div className="ml-16 h-full flex flex-col">
              {/* Bars + line overlay */}
              <div className="flex-1 flex items-end gap-2 relative">
                {/* Grid lines */}
                <div className="absolute inset-0 flex flex-col justify-between pointer-events-none">
                  {[0, 1, 2, 3, 4].map((i) => (
                    <div key={i} className="border-t border-white/5 w-full" />
                  ))}
                </div>

                {history.map((m, i) => {
                  const barH = maxGross > 0 ? (m.gross_pay / maxGross) * 100 : 0;
                  return (
                    <div key={m.month} className="flex-1 flex flex-col items-center justify-end relative z-10">
                      {/* Tooltip on hover */}
                      <div className="group relative w-full flex justify-center">
                        <div
                          className="w-full max-w-[48px] rounded-t-md bg-emerald-500 hover:bg-emerald-400 transition-colors cursor-default"
                          style={{ height: `${barH}%`, minHeight: barH > 0 ? 4 : 0 }}
                          title={`${m.month}: ${fmtEGPShort(m.gross_pay)} | Labor: ${m.labor_cost_pct.toFixed(1)}%`}
                        />
                      </div>
                    </div>
                  );
                })}

                {/* Labor cost % line overlay */}
                <svg
                  className="absolute inset-0 pointer-events-none z-20"
                  viewBox={`0 0 ${history.length * 100} 100`}
                  preserveAspectRatio="none"
                >
                  <polyline
                    fill="none"
                    stroke="#f59e0b"
                    strokeWidth="3"
                    strokeLinejoin="round"
                    strokeLinecap="round"
                    vectorEffect="non-scaling-stroke"
                    points={history
                      .map((m, i) => {
                        const x = (i + 0.5) * (100 / history.length) * history.length;
                        const y = 100 - Math.min(m.labor_cost_pct, 100);
                        return `${x},${y}`;
                      })
                      .join(' ')}
                  />
                  {history.map((m, i) => {
                    const x = (i + 0.5) * (100 / history.length) * history.length;
                    const y = 100 - Math.min(m.labor_cost_pct, 100);
                    return (
                      <circle key={i} cx={x} cy={y} r="4" fill="#f59e0b" vectorEffect="non-scaling-stroke" />
                    );
                  })}
                </svg>
              </div>

              {/* X-axis month labels */}
              <div className="flex gap-2 mt-2">
                {history.map((m) => (
                  <div key={m.month} className="flex-1 text-center text-xs text-slate-400 truncate">
                    {m.month}
                  </div>
                ))}
              </div>
            </div>
          </div>

          {/* Legend */}
          <div className="flex items-center gap-6 mt-4 ml-16">
            <div className="flex items-center gap-2">
              <div className="w-3 h-3 rounded-sm bg-emerald-500" />
              <span className="text-xs text-slate-400">Gross Pay</span>
            </div>
            <div className="flex items-center gap-2">
              <div className="w-3 h-3 rounded-full bg-amber-500" />
              <span className="text-xs text-slate-400">Labor Cost %</span>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

// ── Short currency helper ───────────────────────────────────────────────────

function fmtEGPShort(cents: number): string {
  const val = cents / 100;
  if (val >= 1_000_000) return `EGP ${(val / 1_000_000).toFixed(1)}M`;
  if (val >= 1_000) return `EGP ${(val / 1_000).toFixed(1)}K`;
  return `EGP ${val.toFixed(0)}`;
}
