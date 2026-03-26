import { useState, useEffect, useCallback } from 'react';
import { Link } from 'react-router-dom';
import {
  Clock,
  CheckCircle2,
  Circle,
  ChevronRight,
  Megaphone,
  Trophy,
  Loader2,
  CalendarOff,
  BellOff,
  ClipboardList,
} from 'lucide-react';
import { api } from '../lib/api';
import { getUser } from '../stores/auth';

/* ---------- types ---------- */
interface ScheduleEntry {
  date: string;
  start_time: string;
  end_time: string;
}

interface Task {
  id: string;
  title: string;
  status: 'pending' | 'in_progress' | 'completed';
}

interface Announcement {
  id: string;
  title: string;
  body: string;
  priority: 'urgent' | 'normal' | 'low';
  created_at: string;
}

/* ---------- helpers ---------- */
function relativeTime(iso: string): string {
  const diff = Date.now() - new Date(iso).getTime();
  const mins = Math.floor(diff / 60000);
  if (mins < 1) return 'just now';
  if (mins < 60) return `${mins}m ago`;
  const hrs = Math.floor(mins / 60);
  if (hrs < 24) return `${hrs}h ago`;
  const days = Math.floor(hrs / 24);
  return `${days}d ago`;
}

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

/* ---------- card wrapper ---------- */
function Card({ children, className = '' }: { children: React.ReactNode; className?: string }) {
  return (
    <div className={`rounded-xl border border-white/10 bg-white/5 p-4 ${className}`}>
      {children}
    </div>
  );
}

/* ========== SHIFT STATUS ========== */
function ShiftStatusCard() {
  const user = getUser();
  const [schedule, setSchedule] = useState<ScheduleEntry | null>(null);
  const [loading, setLoading] = useState(true);
  const [clockedIn, setClockedIn] = useState(false);
  const [elapsed, setElapsed] = useState(0);
  const [clockInTime, setClockInTime] = useState<number | null>(null);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        const data = await api<{ schedules: ScheduleEntry[] }>(
          `/labor/schedules/employee/${user?.user_id}`,
        );
        if (cancelled) return;
        const today = new Date().toISOString().slice(0, 10);
        const entry = data.schedules?.find((s) => s.date === today) ?? null;
        setSchedule(entry);
      } catch {
        /* schedule not available */
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();
    return () => { cancelled = true; };
  }, [user?.user_id]);

  // elapsed timer
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
    <Card>
      <div className="flex items-center gap-2 mb-3">
        <Clock className="h-5 w-5 text-orange-400" />
        <h2 className="text-sm font-semibold text-slate-300 uppercase tracking-wide">
          Today's Shift
        </h2>
      </div>

      {loading ? (
        <div className="flex justify-center py-4">
          <Loader2 className="h-5 w-5 animate-spin text-slate-500" />
        </div>
      ) : schedule ? (
        <p className="text-slate-400 text-sm mb-4">
          {formatTime(schedule.start_time)} &ndash; {formatTime(schedule.end_time)}
        </p>
      ) : (
        <p className="text-slate-500 text-sm mb-4">No shift scheduled</p>
      )}

      {clockedIn && (
        <p className="text-center text-2xl font-mono text-white mb-3 tabular-nums">
          {formatElapsed(elapsed)}
        </p>
      )}

      <button
        onClick={toggle}
        className={`w-full rounded-lg py-3 text-sm font-semibold transition-colors ${
          clockedIn
            ? 'bg-emerald-600 hover:bg-emerald-700 text-white'
            : 'bg-orange-500 hover:bg-orange-600 text-white'
        }`}
      >
        {clockedIn ? 'Clock Out' : 'Clock In'}
      </button>
    </Card>
  );
}

/* ========== ACTIVE TASKS ========== */
function ActiveTasksCard() {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        const data = await api<{ tasks: Task[] }>('/tasks/my');
        if (!cancelled) setTasks(data.tasks ?? []);
      } catch {
        /* no tasks */
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();
    return () => { cancelled = true; };
  }, []);

  const completed = tasks.filter((t) => t.status === 'completed').length;
  const total = tasks.length;
  const pct = total > 0 ? Math.round((completed / total) * 100) : 0;

  return (
    <Card>
      <div className="flex items-center justify-between mb-3">
        <div className="flex items-center gap-2">
          <ClipboardList className="h-5 w-5 text-orange-400" />
          <h2 className="text-sm font-semibold text-slate-300 uppercase tracking-wide">
            Active Tasks
          </h2>
        </div>
        <Link to="/tasks" className="text-xs text-orange-400 flex items-center gap-0.5">
          View All <ChevronRight className="h-3 w-3" />
        </Link>
      </div>

      {loading ? (
        <div className="flex justify-center py-4">
          <Loader2 className="h-5 w-5 animate-spin text-slate-500" />
        </div>
      ) : total === 0 ? (
        <p className="text-slate-500 text-sm py-2">No tasks assigned</p>
      ) : (
        <>
          <p className="text-slate-400 text-sm mb-2">
            {completed}/{total} Tasks Complete
          </p>

          {/* progress bar */}
          <div className="h-2 rounded-full bg-white/10 mb-4 overflow-hidden">
            <div
              className="h-full rounded-full bg-orange-500 transition-all"
              style={{ width: `${pct}%` }}
            />
          </div>

          {/* first 3 tasks */}
          <ul className="space-y-2">
            {tasks.slice(0, 3).map((t) => (
              <li key={t.id} className="flex items-center gap-2">
                {t.status === 'completed' ? (
                  <CheckCircle2 className="h-4 w-4 text-emerald-400 shrink-0" />
                ) : (
                  <Circle className="h-4 w-4 text-slate-600 shrink-0" />
                )}
                <span
                  className={`text-sm truncate ${
                    t.status === 'completed' ? 'text-slate-500 line-through' : 'text-slate-300'
                  }`}
                >
                  {t.title}
                </span>
              </li>
            ))}
          </ul>
        </>
      )}
    </Card>
  );
}

/* ========== ANNOUNCEMENTS ========== */
function AnnouncementsCard() {
  const [items, setItems] = useState<Announcement[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        const data = await api<{ announcements: Announcement[] }>('/announcements');
        if (!cancelled) setItems(data.announcements ?? []);
      } catch {
        /* none */
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();
    return () => { cancelled = true; };
  }, []);

  const borderColor: Record<string, string> = {
    urgent: 'border-l-red-500',
    normal: 'border-l-orange-500',
    low: 'border-l-slate-500',
  };

  return (
    <Card>
      <div className="flex items-center gap-2 mb-3">
        <Megaphone className="h-5 w-5 text-orange-400" />
        <h2 className="text-sm font-semibold text-slate-300 uppercase tracking-wide">
          Announcements
        </h2>
      </div>

      {loading ? (
        <div className="flex justify-center py-4">
          <Loader2 className="h-5 w-5 animate-spin text-slate-500" />
        </div>
      ) : items.length === 0 ? (
        <div className="flex flex-col items-center py-4 text-slate-500">
          <BellOff className="h-8 w-8 mb-1" />
          <p className="text-sm">No announcements</p>
        </div>
      ) : (
        <ul className="space-y-2">
          {items.map((a) => (
            <li
              key={a.id}
              className={`border-l-2 pl-3 py-1 ${borderColor[a.priority] ?? 'border-l-slate-500'}`}
            >
              <div className="flex items-center justify-between">
                <p className="text-sm font-medium text-slate-200 truncate pr-2">{a.title}</p>
                <span className="text-[10px] text-slate-500 whitespace-nowrap">
                  {relativeTime(a.created_at)}
                </span>
              </div>
              <p className="text-xs text-slate-500 line-clamp-2 mt-0.5">{a.body}</p>
            </li>
          ))}
        </ul>
      )}
    </Card>
  );
}

/* ========== MY POINTS ========== */
function PointsCard() {
  const user = getUser();
  const points = user?.staff_points ?? 0;

  // fake sparkline data (7 bars)
  const bars = [3, 5, 2, 7, 4, 6, 8];
  const max = Math.max(...bars);

  return (
    <Card>
      <div className="flex items-center justify-between mb-3">
        <div className="flex items-center gap-2">
          <Trophy className="h-5 w-5 text-orange-400" />
          <h2 className="text-sm font-semibold text-slate-300 uppercase tracking-wide">
            My Points
          </h2>
        </div>
        <Link to="/points" className="text-xs text-orange-400 flex items-center gap-0.5">
          Leaderboard <ChevronRight className="h-3 w-3" />
        </Link>
      </div>

      <div className="flex items-end justify-between">
        <p className="text-4xl font-bold text-white tabular-nums">{points.toLocaleString()}</p>

        {/* sparkline */}
        <div className="flex items-end gap-1 h-10">
          {bars.map((v, i) => (
            <div
              key={i}
              className="w-2 rounded-sm bg-orange-500/60"
              style={{ height: `${(v / max) * 100}%` }}
            />
          ))}
        </div>
      </div>
    </Card>
  );
}

/* ========== PAGE ========== */
export default function HomePage() {
  const user = getUser();

  return (
    <div className="p-4 pb-24 space-y-4 max-w-lg mx-auto">
      {/* greeting */}
      <div>
        <h1 className="text-lg font-bold text-white">
          Hey, {user?.display_name?.split(' ')[0] ?? 'there'}
        </h1>
        <p className="text-xs text-slate-500">
          {new Date().toLocaleDateString('en-US', {
            weekday: 'long',
            month: 'long',
            day: 'numeric',
          })}
        </p>
      </div>

      <ShiftStatusCard />
      <ActiveTasksCard />
      <AnnouncementsCard />
      <PointsCard />
    </div>
  );
}
