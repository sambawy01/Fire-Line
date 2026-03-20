import { useState, useMemo } from 'react';
import {
  ChevronLeft,
  ChevronRight,
  Clock,
  DollarSign,
  TrendingUp,
  Users,
  ChevronDown,
  ChevronUp,
  AlertTriangle,
  ArrowLeftRight,
} from 'lucide-react';
import { useLocationStore } from '../stores/location';
import {
  useSchedule,
  useGenerateSchedule,
  usePublishSchedule,
  useLaborCost,
  useForecast,
  useSwaps,
  useOvertimeRisk,
  useReviewSwap,
} from '../hooks/useScheduling';
import StatusBadge from '../components/ui/StatusBadge';
import KPICard from '../components/ui/KPICard';
import LoadingSpinner from '../components/ui/LoadingSpinner';
import ErrorBanner from '../components/ui/ErrorBanner';
import type { ScheduledShift } from '../lib/api';

// Station color map
const STATION_COLORS: Record<string, string> = {
  grill: 'bg-red-500',
  fryer: 'bg-orange-500',
  prep: 'bg-green-500',
  expo: 'bg-blue-500',
  register: 'bg-violet-500',
  dish: 'bg-gray-500',
};

const STATION_TEXT: Record<string, string> = {
  grill: 'text-red-700 bg-red-50 border-red-200',
  fryer: 'text-orange-700 bg-orange-50 border-orange-200',
  prep: 'text-green-700 bg-green-50 border-green-200',
  expo: 'text-blue-700 bg-blue-50 border-blue-200',
  register: 'text-violet-700 bg-violet-50 border-violet-200',
  dish: 'text-slate-300 bg-white/5 border-white/10',
};

function getMonday(date: Date): Date {
  const d = new Date(date);
  const day = d.getDay();
  const diff = (day === 0 ? -6 : 1 - day);
  d.setDate(d.getDate() + diff);
  d.setHours(0, 0, 0, 0);
  return d;
}

function formatDate(date: Date): string {
  return date.toISOString().split('T')[0];
}

function addDays(date: Date, n: number): Date {
  const d = new Date(date);
  d.setDate(d.getDate() + n);
  return d;
}

function formatWeekLabel(start: Date): string {
  const end = addDays(start, 6);
  const opts: Intl.DateTimeFormatOptions = { month: 'short', day: 'numeric', year: 'numeric' };
  return `${start.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })} – ${end.toLocaleDateString('en-US', opts)}`;
}

const DAY_LABELS = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'];

function scheduleStatusVariant(status: string): 'neutral' | 'info' | 'success' {
  if (status === 'published') return 'success';
  if (status === 'draft') return 'info';
  return 'neutral';
}

function overtimeSeverityVariant(severity: string): 'warning' | 'critical' | 'info' {
  if (severity === 'critical') return 'critical';
  if (severity === 'warning') return 'warning';
  return 'info';
}

function costStatusVariant(over_under: string): 'success' | 'critical' | 'info' {
  if (over_under === 'on_track') return 'success';
  if (over_under === 'over') return 'critical';
  return 'info';
}

export default function SchedulingPage() {
  const { selectedLocationId } = useLocationStore();

  // Week navigation
  const today = useMemo(() => new Date(), []);
  const [weekStart, setWeekStart] = useState<Date>(() => getMonday(today));
  const weekStartStr = formatDate(weekStart);

  if (!selectedLocationId) return <LoadingSpinner fullPage />;

  // Selected day for forecast (defaults to Monday)
  const [selectedDayIdx, setSelectedDayIdx] = useState(0);
  const selectedDayDate = formatDate(addDays(weekStart, selectedDayIdx));

  // Forecast collapsible
  const [forecastOpen, setForecastOpen] = useState(false);

  // Queries
  const {
    data: schedule,
    isLoading: scheduleLoading,
    error: scheduleError,
    refetch: refetchSchedule,
  } = useSchedule(selectedLocationId, weekStartStr);

  const { data: costData, isLoading: costLoading } = useLaborCost(schedule?.schedule_id ?? null);
  const { data: forecastData, isLoading: forecastLoading } = useForecast(
    selectedLocationId,
    selectedDayDate
  );
  const { data: swapsData } = useSwaps(selectedLocationId);
  const { data: overtimeData } = useOvertimeRisk(selectedLocationId, weekStartStr);

  // Mutations
  const generateMutation = useGenerateSchedule();
  const publishMutation = usePublishSchedule();
  const reviewSwapMutation = useReviewSwap();

  // Build grid data: unique employees x 7 days
  const { employees, shiftMap } = useMemo(() => {
    if (!schedule?.shifts?.length) return { employees: [] as string[], shiftMap: {} as Record<string, Record<string, ScheduledShift[]>> };
    const empSet = new Map<string, string>(); // id -> name
    const map: Record<string, Record<string, ScheduledShift[]>> = {};
    for (const shift of schedule.shifts) {
      if (!empSet.has(shift.employee_id)) empSet.set(shift.employee_id, shift.employee_name);
      if (!map[shift.employee_id]) map[shift.employee_id] = {};
      if (!map[shift.employee_id][shift.shift_date]) map[shift.employee_id][shift.shift_date] = [];
      map[shift.employee_id][shift.shift_date].push(shift);
    }
    return { employees: Array.from(empSet.entries()).map(([id, name]) => ({ id, name })), shiftMap: map };
  }, [schedule]);

  const handleGenerate = () => {
    if (!selectedLocationId) return;
    generateMutation.mutate({ locationId: selectedLocationId, weekStart: weekStartStr });
  };

  const handlePublish = () => {
    if (!schedule?.schedule_id) return;
    publishMutation.mutate(schedule.schedule_id, {
      onSuccess: () => refetchSchedule(),
    });
  };

  const hasSchedule = !!schedule?.schedule_id;
  const isDraft = schedule?.status === 'draft';

  return (
    <div className="space-y-6">
      {/* Page Title */}
      <div>
        <h1 className="text-2xl font-bold text-white">Scheduling</h1>
        <p className="text-sm text-slate-400 mt-1">Weekly shift schedule, demand forecast, and labor cost management</p>
      </div>

      {/* Top Bar: week selector + actions */}
      <div className="bg-white/5 rounded-xl border border-white/10 px-5 py-4 flex flex-wrap items-center gap-4 shadow-sm">
        {/* Week Selector */}
        <div className="flex items-center gap-2">
          <button
            onClick={() => setWeekStart(addDays(weekStart, -7))}
            className="p-1.5 rounded-md hover:bg-white/10 text-slate-300 transition-colors"
            aria-label="Previous week"
          >
            <ChevronLeft className="h-5 w-5" />
          </button>
          <span className="text-sm font-medium text-white min-w-[200px] text-center">
            {formatWeekLabel(weekStart)}
          </span>
          <button
            onClick={() => setWeekStart(addDays(weekStart, 7))}
            className="p-1.5 rounded-md hover:bg-white/10 text-slate-300 transition-colors"
            aria-label="Next week"
          >
            <ChevronRight className="h-5 w-5" />
          </button>
        </div>

        {/* Status badge */}
        {hasSchedule && (
          <StatusBadge variant={scheduleStatusVariant(schedule.status)}>
            {schedule.status.charAt(0).toUpperCase() + schedule.status.slice(1)}
          </StatusBadge>
        )}

        <div className="flex items-center gap-2 ml-auto flex-wrap">
          {/* Generate Draft */}
          <button
            onClick={handleGenerate}
            disabled={hasSchedule || generateMutation.isPending || !selectedLocationId}
            className="px-4 py-2 rounded-lg text-sm font-medium bg-[#1E293B] text-white hover:bg-[#334155] disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
          >
            {generateMutation.isPending ? 'Generating…' : 'Generate Draft'}
          </button>

          {/* Publish */}
          {hasSchedule && isDraft && (
            <button
              onClick={handlePublish}
              disabled={publishMutation.isPending}
              className="px-4 py-2 rounded-lg text-sm font-medium bg-[#F97316] text-white hover:bg-orange-600 disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
            >
              {publishMutation.isPending ? 'Publishing…' : 'Publish'}
            </button>
          )}
        </div>
      </div>

      {/* Error states */}
      {generateMutation.isError && (
        <ErrorBanner message={(generateMutation.error as Error)?.message ?? 'Failed to generate schedule'} />
      )}
      {publishMutation.isError && (
        <ErrorBanner message={(publishMutation.error as Error)?.message ?? 'Failed to publish schedule'} />
      )}

      {/* Schedule Grid */}
      <div className="bg-white/5 rounded-xl border border-white/10 shadow-sm overflow-hidden">
        <div className="px-5 py-4 border-b border-white/5">
          <h2 className="text-base font-semibold text-white">Weekly Schedule</h2>
          <p className="text-xs text-slate-400 mt-0.5">Click a day header to view demand forecast below</p>
        </div>

        {scheduleLoading ? (
          <div className="flex items-center justify-center h-40">
            <LoadingSpinner />
          </div>
        ) : scheduleError && !hasSchedule ? (
          <div className="flex flex-col items-center justify-center h-40 gap-2 text-slate-500">
            <Users className="h-8 w-8" />
            <p className="text-sm">No schedule for this week. Generate a draft to get started.</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-white/5 bg-white/5">
                  <th className="text-left px-4 py-3 font-medium text-slate-300 w-40 sticky left-0 bg-slate-900 z-10">
                    Employee
                  </th>
                  {DAY_LABELS.map((day, idx) => {
                    const dayDate = addDays(weekStart, idx);
                    const isSelected = idx === selectedDayIdx;
                    return (
                      <th
                        key={day}
                        onClick={() => {
                          setSelectedDayIdx(idx);
                          setForecastOpen(true);
                        }}
                        className={`px-3 py-3 text-center font-medium cursor-pointer select-none transition-colors ${
                          isSelected
                            ? 'text-[#F97316] bg-orange-50'
                            : 'text-slate-300 hover:bg-white/10'
                        }`}
                      >
                        <div>{day}</div>
                        <div className="text-xs font-normal mt-0.5">
                          {dayDate.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })}
                        </div>
                      </th>
                    );
                  })}
                </tr>
              </thead>
              <tbody>
                {(employees as { id: string; name: string }[]).length === 0 ? (
                  <tr>
                    <td colSpan={8} className="text-center py-10 text-slate-500 text-sm">
                      No shifts scheduled for this week.
                    </td>
                  </tr>
                ) : (
                  (employees as { id: string; name: string }[]).map(({ id, name }) => (
                    <tr key={id} className="border-t border-white/5 hover:bg-white/5">
                      <td className="px-4 py-2 font-medium text-white sticky left-0 bg-slate-900 z-10 border-r border-white/5">
                        {name}
                      </td>
                      {DAY_LABELS.map((_, idx) => {
                        const dayStr = formatDate(addDays(weekStart, idx));
                        const dayShifts = shiftMap[id]?.[dayStr] ?? [];
                        return (
                          <td key={idx} className="px-2 py-2 align-top min-w-[110px]">
                            {dayShifts.length === 0 ? (
                              <span className="text-slate-500 text-xs">—</span>
                            ) : (
                              <div className="space-y-1">
                                {dayShifts.map((shift) => {
                                  const stationKey = shift.station?.toLowerCase() ?? '';
                                  const stationClass = STATION_TEXT[stationKey] || 'text-slate-300 bg-white/5 border-white/10';
                                  const dotClass = STATION_COLORS[stationKey] || 'bg-gray-400';
                                  return (
                                    <div
                                      key={shift.scheduled_shift_id}
                                      className={`rounded-md border px-2 py-1 text-xs ${stationClass}`}
                                    >
                                      <div className="flex items-center gap-1 mb-0.5">
                                        <span className={`inline-block h-2 w-2 rounded-full ${dotClass} shrink-0`} />
                                        <span className="font-semibold capitalize">{shift.station}</span>
                                      </div>
                                      <div className="text-xs opacity-80">
                                        {shift.start_time}–{shift.end_time}
                                      </div>
                                    </div>
                                  );
                                })}
                              </div>
                            )}
                          </td>
                        );
                      })}
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Demand Forecast (collapsible) */}
      <div className="bg-white/5 rounded-xl border border-white/10 shadow-sm">
        <button
          onClick={() => setForecastOpen((v) => !v)}
          className="w-full flex items-center justify-between px-5 py-4 text-left"
        >
          <div>
            <h2 className="text-base font-semibold text-white">Demand Forecast</h2>
            <p className="text-xs text-slate-400 mt-0.5">
              {DAY_LABELS[selectedDayIdx]}{' '}
              {addDays(weekStart, selectedDayIdx).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })}
            </p>
          </div>
          {forecastOpen ? (
            <ChevronUp className="h-5 w-5 text-slate-500" />
          ) : (
            <ChevronDown className="h-5 w-5 text-slate-500" />
          )}
        </button>

        {forecastOpen && (
          <div className="border-t border-white/5 px-5 py-4">
            {forecastLoading ? (
              <div className="flex justify-center py-6">
                <LoadingSpinner />
              </div>
            ) : !forecastData?.forecast?.length ? (
              <p className="text-sm text-slate-500 text-center py-6">No forecast data available for this day.</p>
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="text-left border-b border-white/5">
                      <th className="pb-2 font-medium text-slate-300 pr-4">Time Block</th>
                      <th className="pb-2 font-medium text-slate-300 pr-4 text-right">Forecasted Covers</th>
                      <th className="pb-2 font-medium text-slate-300 text-right">Required Headcount</th>
                    </tr>
                  </thead>
                  <tbody>
                    {forecastData.forecast.map((block) => (
                      <tr key={block.time_block} className="border-t border-white/5">
                        <td className="py-2 pr-4 text-slate-200 font-medium">{block.time_block}</td>
                        <td className="py-2 pr-4 text-right">
                          <div className="flex items-center justify-end gap-2">
                            <div
                              className="h-2 rounded-full bg-[#F97316] opacity-70"
                              style={{ width: `${Math.min(block.forecasted_covers * 2, 160)}px` }}
                            />
                            <span className="text-slate-200 w-8 text-right">{block.forecasted_covers}</span>
                          </div>
                        </td>
                        <td className="py-2 text-right">
                          <span className="inline-flex items-center justify-center h-6 w-6 rounded-full bg-blue-100 text-blue-700 font-semibold text-xs">
                            {block.required_headcount}
                          </span>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        )}
      </div>

      {/* Labor Cost Projection */}
      {hasSchedule && (
        <div>
          <h2 className="text-base font-semibold text-white mb-3">Labor Cost Projection</h2>
          {costLoading ? (
            <div className="flex justify-center py-6">
              <LoadingSpinner />
            </div>
          ) : costData ? (
            <div className="space-y-3">
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                <KPICard
                  label="Total Hours"
                  value={`${costData.total_hours.toFixed(1)}h`}
                  icon={Clock}
                  iconColor="text-blue-600"
                  bgTint="bg-blue-50"
                />
                <KPICard
                  label="Total Cost"
                  value={`$${costData.total_cost.toLocaleString()}`}
                  icon={DollarSign}
                  iconColor="text-green-600"
                  bgTint="bg-green-50"
                />
                <KPICard
                  label="Labor Cost %"
                  value={`${costData.labor_cost_pct.toFixed(1)}%`}
                  icon={TrendingUp}
                  iconColor="text-orange-600"
                  bgTint="bg-orange-50"
                />
                <KPICard
                  label="Budget Target %"
                  value={`${costData.budget_target_pct.toFixed(1)}%`}
                  icon={TrendingUp}
                  iconColor="text-violet-600"
                  bgTint="bg-violet-50"
                />
              </div>
              <div className="flex items-center gap-2">
                <span className="text-sm text-slate-300">Status:</span>
                <StatusBadge variant={costStatusVariant(costData.over_under)}>
                  {costData.over_under === 'on_track'
                    ? 'On Track'
                    : costData.over_under === 'over'
                    ? 'Over Budget'
                    : 'Under Budget'}
                </StatusBadge>
              </div>
            </div>
          ) : (
            <p className="text-sm text-slate-500">Cost projection not available.</p>
          )}
        </div>
      )}

      {/* Alerts Row: Overtime + Swaps */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Overtime Risks */}
        <div className="bg-white/5 rounded-xl border border-white/10 shadow-sm">
          <div className="px-5 py-4 border-b border-white/5 flex items-center gap-2">
            <AlertTriangle className="h-4 w-4 text-amber-500" />
            <h2 className="text-base font-semibold text-white">Overtime Warnings</h2>
          </div>
          <div className="px-5 py-4 space-y-3">
            {!overtimeData?.risks?.length ? (
              <p className="text-sm text-slate-500 text-center py-4">No overtime risks for this week.</p>
            ) : (
              overtimeData.risks.map((risk) => (
                <div
                  key={risk.employee_id}
                  className="flex items-center justify-between rounded-lg border border-white/5 bg-white/5 px-4 py-3"
                >
                  <div>
                    <p className="font-medium text-white text-sm">{risk.employee_name}</p>
                    <p className="text-xs text-slate-400 mt-0.5">{risk.scheduled_hours}h scheduled</p>
                  </div>
                  <StatusBadge variant={overtimeSeverityVariant(risk.severity)}>
                    {risk.severity.charAt(0).toUpperCase() + risk.severity.slice(1)}
                  </StatusBadge>
                </div>
              ))
            )}
          </div>
        </div>

        {/* Swap Requests */}
        <div className="bg-white/5 rounded-xl border border-white/10 shadow-sm">
          <div className="px-5 py-4 border-b border-white/5 flex items-center gap-2">
            <ArrowLeftRight className="h-4 w-4 text-blue-500" />
            <h2 className="text-base font-semibold text-white">Pending Swap Requests</h2>
          </div>
          <div className="px-5 py-4 space-y-3">
            {!swapsData?.swap_requests?.length ? (
              <p className="text-sm text-slate-500 text-center py-4">No pending swap requests.</p>
            ) : (
              swapsData.swap_requests.map((swap) => (
                <div
                  key={swap.swap_id}
                  className="rounded-lg border border-white/5 bg-white/5 px-4 py-3"
                >
                  <div className="flex items-start justify-between gap-2">
                    <div className="flex-1 min-w-0">
                      <p className="font-medium text-white text-sm truncate">
                        {swap.requester_name}
                        {swap.target_name ? ` → ${swap.target_name}` : ''}
                      </p>
                      {swap.reason && (
                        <p className="text-xs text-slate-400 mt-0.5 line-clamp-2">{swap.reason}</p>
                      )}
                      <p className="text-xs text-slate-500 mt-1">
                        {new Date(swap.created_at).toLocaleDateString('en-US', {
                          month: 'short',
                          day: 'numeric',
                          hour: '2-digit',
                          minute: '2-digit',
                        })}
                      </p>
                    </div>
                    <div className="flex items-center gap-1.5 shrink-0">
                      <button
                        onClick={() => reviewSwapMutation.mutate({ id: swap.swap_id, approved: true })}
                        disabled={reviewSwapMutation.isPending}
                        className="px-3 py-1.5 rounded-md text-xs font-medium bg-emerald-600 text-white hover:bg-emerald-700 disabled:opacity-50 transition-colors"
                      >
                        Approve
                      </button>
                      <button
                        onClick={() => reviewSwapMutation.mutate({ id: swap.swap_id, approved: false })}
                        disabled={reviewSwapMutation.isPending}
                        className="px-3 py-1.5 rounded-md text-xs font-medium bg-red-100 text-red-700 hover:bg-red-200 disabled:opacity-50 transition-colors"
                      >
                        Deny
                      </button>
                    </div>
                  </div>
                </div>
              ))
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
