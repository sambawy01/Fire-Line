import { useState, useEffect, useCallback, useMemo } from 'react';
import {
  ChevronLeft,
  ChevronRight,
  CalendarOff,
  ArrowLeftRight,
  Loader2,
  X,
  Check,
} from 'lucide-react';
import { api } from '../lib/api';
import { getUser } from '../stores/auth';

/* ---------- types ---------- */
interface ScheduleEntry {
  id: string;
  date: string;
  start_time: string;
  end_time: string;
  role?: string;
  station?: string;
  status?: 'confirmed' | 'pending_swap';
}

interface SwapRequest {
  id: string;
  shift_id: string;
  shift_date: string;
  requested_at: string;
  status: 'pending' | 'approved' | 'denied';
  target_employee_name?: string;
  reason?: string;
}

/* ---------- helpers ---------- */
function formatTime(t: string): string {
  const [h, m] = t.split(':').map(Number);
  const ampm = h >= 12 ? 'PM' : 'AM';
  const hour = h % 12 || 12;
  return `${hour}:${String(m).padStart(2, '0')} ${ampm}`;
}

function formatElapsed(seconds: number): string {
  const h = Math.floor(seconds / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  const s = seconds % 60;
  return `${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`;
}

function getMonday(date: Date): Date {
  const d = new Date(date);
  const day = d.getDay();
  const diff = d.getDate() - day + (day === 0 ? -6 : 1);
  d.setDate(diff);
  d.setHours(0, 0, 0, 0);
  return d;
}

function addDays(date: Date, days: number): Date {
  const d = new Date(date);
  d.setDate(d.getDate() + days);
  return d;
}

function isSameDay(a: Date, b: Date): boolean {
  return (
    a.getFullYear() === b.getFullYear() &&
    a.getMonth() === b.getMonth() &&
    a.getDate() === b.getDate()
  );
}

function toDateStr(d: Date): string {
  return d.toISOString().slice(0, 10);
}

const DAY_NAMES = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'];
const MONTH_SHORT = [
  'Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun',
  'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec',
];

function formatDateRange(mon: Date): string {
  const sun = addDays(mon, 6);
  const mMonth = MONTH_SHORT[mon.getMonth()];
  const sMonth = MONTH_SHORT[sun.getMonth()];
  if (mMonth === sMonth) {
    return `${mMonth} ${mon.getDate()} - ${sun.getDate()}, ${mon.getFullYear()}`;
  }
  return `${mMonth} ${mon.getDate()} - ${sMonth} ${sun.getDate()}, ${sun.getFullYear()}`;
}

/* ---------- card wrapper ---------- */
function Card({ children, className = '' }: { children: React.ReactNode; className?: string }) {
  return (
    <div className={`rounded-xl border border-white/10 bg-white/5 p-4 ${className}`}>
      {children}
    </div>
  );
}

/* ---------- status badge ---------- */
function StatusBadge({ status }: { status: 'pending' | 'approved' | 'denied' }) {
  const styles = {
    pending: 'bg-amber-500/20 text-amber-400 border-amber-500/30',
    approved: 'bg-green-500/20 text-green-400 border-green-500/30',
    denied: 'bg-red-500/20 text-red-400 border-red-500/30',
  };
  return (
    <span
      className={`inline-flex items-center rounded-full border px-2 py-0.5 text-[10px] font-medium uppercase tracking-wide ${styles[status]}`}
    >
      {status}
    </span>
  );
}

/* ========== SWAP MODAL ========== */
function SwapModal({
  shifts,
  onClose,
  onSubmit,
}: {
  shifts: ScheduleEntry[];
  onClose: () => void;
  onSubmit: (shiftId: string) => void;
}) {
  const [selected, setSelected] = useState<string | null>(null);

  return (
    <div className="fixed inset-0 z-50 flex items-end justify-center bg-black/60 backdrop-blur-sm">
      <div className="w-full max-w-lg rounded-t-2xl border-t border-white/10 bg-slate-800 p-5 pb-8 animate-in slide-in-from-bottom">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-base font-semibold text-white">Request Shift Swap</h3>
          <button onClick={onClose} className="p-1 rounded-lg hover:bg-white/10 transition-colors">
            <X className="h-5 w-5 text-slate-400" />
          </button>
        </div>

        <p className="text-sm text-slate-400 mb-4">Select a shift you want to swap:</p>

        {shifts.length === 0 ? (
          <p className="text-sm text-slate-500 py-4 text-center">No upcoming shifts to swap.</p>
        ) : (
          <div className="space-y-2 max-h-60 overflow-y-auto">
            {shifts.map((s) => {
              const d = new Date(s.date + 'T00:00:00');
              const isSelected = selected === s.id;
              return (
                <button
                  key={s.id}
                  onClick={() => setSelected(s.id)}
                  className={`w-full flex items-center justify-between rounded-lg border p-3 text-left transition-colors ${
                    isSelected
                      ? 'border-orange-500 bg-orange-500/10'
                      : 'border-white/10 bg-white/5 hover:bg-white/10'
                  }`}
                >
                  <div>
                    <p className="text-sm font-medium text-white">
                      {DAY_NAMES[((d.getDay() + 6) % 7)]}{' '}
                      {MONTH_SHORT[d.getMonth()]} {d.getDate()}
                    </p>
                    <p className="text-xs text-slate-400 mt-0.5">
                      {formatTime(s.start_time)} - {formatTime(s.end_time)}
                      {s.station && ` \u00b7 ${s.station}`}
                    </p>
                  </div>
                  {isSelected && <Check className="h-4 w-4 text-orange-400 shrink-0" />}
                </button>
              );
            })}
          </div>
        )}

        <button
          onClick={() => selected && onSubmit(selected)}
          disabled={!selected}
          className={`w-full mt-4 rounded-lg py-3 text-sm font-semibold transition-colors ${
            selected
              ? 'bg-orange-500 hover:bg-orange-600 text-white'
              : 'bg-white/5 text-slate-600 cursor-not-allowed'
          }`}
        >
          Submit Swap Request
        </button>
      </div>
    </div>
  );
}

/* ========== SCHEDULE GRID ========== */
function ScheduleGrid({
  weekStart,
  schedules,
  loading,
}: {
  weekStart: Date;
  schedules: ScheduleEntry[];
  loading: boolean;
}) {
  const today = new Date();

  const days = useMemo(() => {
    return Array.from({ length: 7 }, (_, i) => {
      const date = addDays(weekStart, i);
      const dateStr = toDateStr(date);
      const entry = schedules.find((s) => s.date === dateStr) ?? null;
      const isToday = isSameDay(date, today);
      return { date, dateStr, entry, isToday, dayName: DAY_NAMES[i] };
    });
  }, [weekStart, schedules, today]);

  if (loading) {
    return (
      <div className="flex justify-center py-12">
        <Loader2 className="h-6 w-6 animate-spin text-slate-500" />
      </div>
    );
  }

  return (
    <div className="grid grid-cols-7 gap-1.5">
      {days.map((d) => {
        const isSwap = d.entry?.status === 'pending_swap';
        const hasShift = !!d.entry;

        let cardClasses = 'rounded-lg p-2 min-h-[100px] flex flex-col transition-colors ';
        if (d.isToday) {
          cardClasses += 'ring-2 ring-orange-500 ';
        }
        if (!hasShift) {
          cardClasses += 'border border-dashed border-slate-700 bg-slate-800/30';
        } else if (isSwap) {
          cardClasses += 'border border-amber-500/40 bg-amber-500/10';
        } else {
          cardClasses += 'border border-green-500/40 bg-green-500/10';
        }

        return (
          <div key={d.dateStr} className={cardClasses}>
            <p
              className={`text-[10px] font-semibold uppercase tracking-wide mb-1 ${
                d.isToday ? 'text-orange-400' : 'text-slate-500'
              }`}
            >
              {d.dayName}
            </p>
            <p
              className={`text-xs font-medium mb-auto ${
                d.isToday ? 'text-orange-300' : 'text-slate-400'
              }`}
            >
              {d.date.getDate()}
            </p>

            {hasShift ? (
              <div className="mt-1">
                <p
                  className={`text-[10px] font-semibold leading-tight ${
                    isSwap ? 'text-amber-300' : 'text-green-300'
                  }`}
                >
                  {formatTime(d.entry!.start_time)}
                </p>
                <p
                  className={`text-[10px] leading-tight ${
                    isSwap ? 'text-amber-400/70' : 'text-green-400/70'
                  }`}
                >
                  {formatTime(d.entry!.end_time)}
                </p>
                {d.entry!.station && (
                  <p className="text-[9px] text-slate-400 mt-0.5 truncate">
                    {d.entry!.station}
                  </p>
                )}
              </div>
            ) : (
              <p className="text-[10px] text-slate-600 mt-1">Off</p>
            )}
          </div>
        );
      })}
    </div>
  );
}

/* ========== SWAP REQUESTS SECTION ========== */
function SwapRequestsSection({
  swaps,
  loading,
  onRequestSwap,
}: {
  swaps: SwapRequest[];
  loading: boolean;
  onRequestSwap: () => void;
}) {
  return (
    <Card>
      <div className="flex items-center justify-between mb-3">
        <div className="flex items-center gap-2">
          <ArrowLeftRight className="h-5 w-5 text-orange-400" />
          <h2 className="text-sm font-semibold text-slate-300 uppercase tracking-wide">
            My Swap Requests
          </h2>
        </div>
        <button
          onClick={onRequestSwap}
          className="text-xs font-medium text-orange-400 hover:text-orange-300 transition-colors"
        >
          Request Swap
        </button>
      </div>

      {loading ? (
        <div className="flex justify-center py-4">
          <Loader2 className="h-5 w-5 animate-spin text-slate-500" />
        </div>
      ) : swaps.length === 0 ? (
        <p className="text-sm text-slate-500 py-2">No swap requests</p>
      ) : (
        <ul className="space-y-2">
          {swaps.map((s) => {
            const d = new Date(s.shift_date + 'T00:00:00');
            return (
              <li
                key={s.id}
                className="flex items-center justify-between rounded-lg border border-white/5 bg-white/5 px-3 py-2"
              >
                <div>
                  <p className="text-sm text-slate-200">
                    {DAY_NAMES[((d.getDay() + 6) % 7)]}{' '}
                    {MONTH_SHORT[d.getMonth()]} {d.getDate()}
                  </p>
                  {s.target_employee_name && (
                    <p className="text-xs text-slate-500">with {s.target_employee_name}</p>
                  )}
                  {s.reason && (
                    <p className="text-xs text-slate-500 mt-0.5">{s.reason}</p>
                  )}
                </div>
                <StatusBadge status={s.status} />
              </li>
            );
          })}
        </ul>
      )}
    </Card>
  );
}

/* ========== CLOCK IN/OUT BUTTON ========== */
function ClockButton() {
  const [clockedIn, setClockedIn] = useState(false);
  const [elapsed, setElapsed] = useState(0);
  const [clockInTime, setClockInTime] = useState<number | null>(null);

  useEffect(() => {
    if (!clockedIn || !clockInTime) return;
    const id = setInterval(() => {
      setElapsed(Math.floor((Date.now() - clockInTime) / 1000));
    }, 1000);
    return () => clearInterval(id);
  }, [clockedIn, clockInTime]);

  const toggle = useCallback(() => {
    if (clockedIn) {
      setClockedIn(false);
      setClockInTime(null);
      setElapsed(0);
    } else {
      setClockedIn(true);
      setClockInTime(Date.now());
    }
  }, [clockedIn]);

  return (
    <div className="fixed bottom-16 left-0 right-0 z-40 px-4 pb-3">
      <div className="max-w-lg mx-auto">
        {clockedIn && (
          <p className="text-center text-lg font-mono text-white mb-2 tabular-nums">
            {formatElapsed(elapsed)}
          </p>
        )}
        <button
          onClick={toggle}
          className={`w-full rounded-xl py-4 text-base font-semibold shadow-lg transition-colors ${
            clockedIn
              ? 'bg-emerald-600 hover:bg-emerald-700 text-white'
              : 'bg-orange-500 hover:bg-orange-600 text-white'
          }`}
        >
          {clockedIn ? 'Clock Out' : 'Clock In'}
        </button>
      </div>
    </div>
  );
}

/* ========== PAGE ========== */
export default function SchedulePage() {
  const user = getUser();
  const [weekOffset, setWeekOffset] = useState(0);
  const [schedules, setSchedules] = useState<ScheduleEntry[]>([]);
  const [loadingSched, setLoadingSched] = useState(true);
  const [swaps, setSwaps] = useState<SwapRequest[]>([]);
  const [loadingSwaps, setLoadingSwaps] = useState(true);
  const [showSwapModal, setShowSwapModal] = useState(false);
  const [noSchedule, setNoSchedule] = useState(false);

  const weekStart = useMemo(() => {
    const base = getMonday(new Date());
    return addDays(base, weekOffset * 7);
  }, [weekOffset]);

  // fetch schedule
  useEffect(() => {
    let cancelled = false;
    setLoadingSched(true);
    setNoSchedule(false);
    (async () => {
      try {
        const data = await api<{ schedules: ScheduleEntry[] }>(
          `/labor/schedules/employee/${user?.user_id}`,
        );
        if (cancelled) return;
        setSchedules(data.schedules ?? []);
        if (!data.schedules || data.schedules.length === 0) {
          setNoSchedule(true);
        }
      } catch {
        if (!cancelled) {
          setSchedules([]);
          setNoSchedule(true);
        }
      } finally {
        if (!cancelled) setLoadingSched(false);
      }
    })();
    return () => { cancelled = true; };
  }, [user?.user_id, weekOffset]);

  // fetch swaps
  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        const data = await api<{ swaps: SwapRequest[] }>('/labor/swaps');
        if (!cancelled) setSwaps(data.swaps ?? []);
      } catch {
        /* no swaps */
      } finally {
        if (!cancelled) setLoadingSwaps(false);
      }
    })();
    return () => { cancelled = true; };
  }, []);

  // upcoming shifts for swap modal
  const upcomingShifts = useMemo(() => {
    const todayStr = toDateStr(new Date());
    return schedules
      .filter((s) => s.date >= todayStr)
      .sort((a, b) => a.date.localeCompare(b.date));
  }, [schedules]);

  const handleSwapSubmit = useCallback((shiftId: string) => {
    // Wire to POST /api/v1/labor/swaps in the future
    console.log('Swap requested for shift:', shiftId);
    setShowSwapModal(false);
  }, []);

  return (
    <div className="min-h-screen bg-slate-900 pb-36">
      {/* header */}
      <div className="px-4 pt-4 pb-3">
        <h1 className="text-lg font-bold text-white">My Schedule</h1>
      </div>

      {/* week navigation */}
      <div className="px-4 mb-4">
        <div className="flex items-center justify-between">
          <button
            onClick={() => setWeekOffset((o) => o - 1)}
            className="p-2 rounded-lg hover:bg-white/10 transition-colors"
            aria-label="Previous week"
          >
            <ChevronLeft className="h-5 w-5 text-slate-400" />
          </button>
          <span className="text-sm font-medium text-slate-300">
            {formatDateRange(weekStart)}
          </span>
          <button
            onClick={() => setWeekOffset((o) => o + 1)}
            className="p-2 rounded-lg hover:bg-white/10 transition-colors"
            aria-label="Next week"
          >
            <ChevronRight className="h-5 w-5 text-slate-400" />
          </button>
        </div>
      </div>

      {/* schedule grid or empty state */}
      <div className="px-4 mb-4">
        {noSchedule && !loadingSched ? (
          <Card className="flex flex-col items-center py-10">
            <CalendarOff className="h-10 w-10 text-slate-600 mb-3" />
            <p className="text-sm font-medium text-slate-400">No schedule published yet</p>
            <p className="text-xs text-slate-600 mt-1">Check back later for your upcoming shifts.</p>
          </Card>
        ) : (
          <ScheduleGrid
            weekStart={weekStart}
            schedules={schedules}
            loading={loadingSched}
          />
        )}
      </div>

      {/* swap requests */}
      <div className="px-4 mb-4">
        <SwapRequestsSection
          swaps={swaps}
          loading={loadingSwaps}
          onRequestSwap={() => setShowSwapModal(true)}
        />
      </div>

      {/* clock in/out */}
      <ClockButton />

      {/* swap modal */}
      {showSwapModal && (
        <SwapModal
          shifts={upcomingShifts}
          onClose={() => setShowSwapModal(false)}
          onSubmit={handleSwapSubmit}
        />
      )}
    </div>
  );
}
