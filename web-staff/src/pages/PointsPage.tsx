import { useState, useEffect, useMemo } from 'react';
import {
  Trophy,
  CheckCircle,
  Zap,
  Clock,
  AlertTriangle,
  XCircle,
  Heart,
  Settings,
  Loader2,
  Medal,
  TrendingUp,
  Flame,
  Hash,
} from 'lucide-react';
import { api } from '../lib/api';
import { getUser } from '../stores/auth';

/* ---------- types ---------- */
interface PointEvent {
  id: string;
  reason: string;
  description: string;
  points: number;
  created_at: string;
}

interface LeaderboardEntry {
  user_id: string;
  display_name: string;
  role: string;
  total_points: number;
  rank: number;
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
  if (days < 7) return `${days}d ago`;
  const weeks = Math.floor(days / 7);
  return `${weeks}w ago`;
}

function getInitials(name: string): string {
  return name
    .split(' ')
    .map((w) => w[0])
    .join('')
    .toUpperCase()
    .slice(0, 2);
}

const REASON_CONFIG: Record<
  string,
  { icon: typeof CheckCircle; color: string; bgColor: string }
> = {
  task_completion: { icon: CheckCircle, color: 'text-green-400', bgColor: 'bg-green-500/15' },
  speed_bonus: { icon: Zap, color: 'text-blue-400', bgColor: 'bg-blue-500/15' },
  attendance: { icon: Clock, color: 'text-green-400', bgColor: 'bg-green-500/15' },
  late: { icon: AlertTriangle, color: 'text-red-400', bgColor: 'bg-red-500/15' },
  no_show: { icon: XCircle, color: 'text-red-400', bgColor: 'bg-red-500/15' },
  peer_nominated: { icon: Heart, color: 'text-pink-400', bgColor: 'bg-pink-500/15' },
  manager_adjustment: { icon: Settings, color: 'text-slate-400', bgColor: 'bg-slate-500/15' },
};

function getReasonConfig(reason: string) {
  return REASON_CONFIG[reason] ?? { icon: Trophy, color: 'text-slate-400', bgColor: 'bg-slate-500/15' };
}

/* ---------- card wrapper ---------- */
function Card({ children, className = '' }: { children: React.ReactNode; className?: string }) {
  return (
    <div className={`rounded-xl border border-white/10 bg-white/5 p-4 ${className}`}>
      {children}
    </div>
  );
}

/* ========== HERO POINTS ========== */
function HeroPoints({ total }: { total: number }) {
  return (
    <div className="flex flex-col items-center py-6">
      <div className="relative mb-2">
        <div className="h-24 w-24 rounded-full bg-gradient-to-br from-orange-500/20 to-amber-500/20 border-2 border-orange-500/40 flex items-center justify-center">
          <Trophy className="h-10 w-10 text-orange-400" />
        </div>
      </div>
      <p className="text-5xl font-bold text-white tabular-nums">{total.toLocaleString()}</p>
      <p className="text-sm text-slate-400 mt-1">Staff Points</p>
    </div>
  );
}

/* ========== MINI STAT CARDS ========== */
function MiniStats({
  weekPoints,
  streak,
  rank,
}: {
  weekPoints: number;
  streak: number;
  rank: number | null;
}) {
  const stats = [
    {
      label: 'This Week',
      value: weekPoints > 0 ? `+${weekPoints}` : '0',
      icon: TrendingUp,
      color: 'text-green-400',
    },
    {
      label: 'Streak',
      value: `${streak}d`,
      icon: Flame,
      color: 'text-orange-400',
    },
    {
      label: 'Rank',
      value: rank ? `#${rank}` : '--',
      icon: Hash,
      color: 'text-blue-400',
    },
  ];

  return (
    <div className="grid grid-cols-3 gap-2">
      {stats.map((s) => (
        <Card key={s.label} className="flex flex-col items-center py-3 px-2">
          <s.icon className={`h-4 w-4 ${s.color} mb-1`} />
          <p className="text-lg font-bold text-white tabular-nums">{s.value}</p>
          <p className="text-[10px] text-slate-500 uppercase tracking-wide">{s.label}</p>
        </Card>
      ))}
    </div>
  );
}

/* ========== POINTS HISTORY ========== */
function PointsHistory({
  events,
  loading,
}: {
  events: PointEvent[];
  loading: boolean;
}) {
  if (loading) {
    return (
      <div className="flex justify-center py-8">
        <Loader2 className="h-6 w-6 animate-spin text-slate-500" />
      </div>
    );
  }

  if (events.length === 0) {
    return (
      <p className="text-sm text-slate-500 py-4 text-center">No points activity yet.</p>
    );
  }

  return (
    <ul className="space-y-1">
      {events.map((e) => {
        const config = getReasonConfig(e.reason);
        const Icon = config.icon;
        const isPositive = e.points >= 0;

        return (
          <li
            key={e.id}
            className="flex items-center gap-3 rounded-lg px-3 py-2.5 hover:bg-white/5 transition-colors"
          >
            <div
              className={`h-8 w-8 rounded-full ${config.bgColor} flex items-center justify-center shrink-0`}
            >
              <Icon className={`h-4 w-4 ${config.color}`} />
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-sm text-slate-200 truncate">{e.description}</p>
              <p className="text-[10px] text-slate-500">{relativeTime(e.created_at)}</p>
            </div>
            <span
              className={`text-sm font-semibold tabular-nums shrink-0 ${
                isPositive ? 'text-green-400' : 'text-red-400'
              }`}
            >
              {isPositive ? '+' : ''}{e.points}
            </span>
          </li>
        );
      })}
    </ul>
  );
}

/* ========== LEADERBOARD ========== */
const MEDAL_COLORS = [
  'bg-yellow-500', // gold
  'bg-slate-300',  // silver
  'bg-amber-700',  // bronze
];

function Leaderboard({
  entries,
  loading,
  currentUserId,
  userRole,
}: {
  entries: LeaderboardEntry[];
  loading: boolean;
  currentUserId: string;
  userRole: string;
}) {
  const showFullNames = userRole !== 'staff';

  const displayName = (entry: LeaderboardEntry) => {
    if (entry.user_id === currentUserId) return 'You';
    return showFullNames ? entry.display_name : getInitials(entry.display_name);
  };

  if (loading) {
    return (
      <div className="flex justify-center py-8">
        <Loader2 className="h-6 w-6 animate-spin text-slate-500" />
      </div>
    );
  }

  if (entries.length === 0) {
    return (
      <p className="text-sm text-slate-500 py-4 text-center">No leaderboard data yet.</p>
    );
  }

  const top3 = entries.slice(0, 3);
  const rest = entries.slice(3);

  return (
    <div>
      {/* top 3 podium */}
      {top3.length > 0 && (
        <div className="flex items-end justify-center gap-3 mb-6 pt-4">
          {/* reorder: 2nd, 1st, 3rd */}
          {[top3[1], top3[0], top3[2]]
            .filter(Boolean)
            .map((entry, idx) => {
              if (!entry) return null;
              const actualRank = entry.rank;
              const isMe = entry.user_id === currentUserId;
              const heights = ['h-16', 'h-20', 'h-12'];
              const sizes = ['h-12 w-12', 'h-14 w-14', 'h-10 w-10'];
              const order = [1, 0, 2]; // maps display position to height/size index

              return (
                <div key={entry.user_id} className="flex flex-col items-center">
                  <div className="relative mb-1">
                    <div
                      className={`${sizes[order[idx]]} rounded-full ${
                        isMe ? 'bg-orange-500/30 border-2 border-orange-500' : 'bg-white/10 border-2 border-white/20'
                      } flex items-center justify-center`}
                    >
                      <span
                        className={`text-xs font-bold ${
                          isMe ? 'text-orange-300' : 'text-slate-300'
                        }`}
                      >
                        {getInitials(entry.display_name)}
                      </span>
                    </div>
                    <div
                      className={`absolute -top-1 -right-1 h-5 w-5 rounded-full ${
                        MEDAL_COLORS[actualRank - 1] ?? 'bg-slate-600'
                      } flex items-center justify-center`}
                    >
                      <Medal className="h-3 w-3 text-white" />
                    </div>
                  </div>
                  <p
                    className={`text-xs font-medium truncate max-w-[70px] text-center ${
                      isMe ? 'text-orange-300' : 'text-slate-300'
                    }`}
                  >
                    {displayName(entry)}
                  </p>
                  <p className="text-[10px] text-slate-500">{entry.total_points.toLocaleString()} pts</p>
                  <div
                    className={`${heights[order[idx]]} w-14 rounded-t-lg mt-1 ${
                      isMe ? 'bg-orange-500/20' : 'bg-white/5'
                    }`}
                  />
                </div>
              );
            })}
        </div>
      )}

      {/* rest of leaderboard */}
      {rest.length > 0 && (
        <ul className="space-y-1">
          {rest.map((entry) => {
            const isMe = entry.user_id === currentUserId;
            return (
              <li
                key={entry.user_id}
                className={`flex items-center gap-3 rounded-lg px-3 py-2.5 ${
                  isMe ? 'bg-orange-500/10 border border-orange-500/30' : 'hover:bg-white/5'
                } transition-colors`}
              >
                <span className="text-sm font-bold text-slate-500 w-6 text-right tabular-nums">
                  {entry.rank}
                </span>
                <div
                  className={`h-8 w-8 rounded-full ${
                    isMe ? 'bg-orange-500/20 border border-orange-500/40' : 'bg-white/10'
                  } flex items-center justify-center shrink-0`}
                >
                  <span
                    className={`text-[10px] font-bold ${
                      isMe ? 'text-orange-300' : 'text-slate-400'
                    }`}
                  >
                    {getInitials(entry.display_name)}
                  </span>
                </div>
                <div className="flex-1 min-w-0">
                  <p
                    className={`text-sm font-medium truncate ${
                      isMe ? 'text-orange-200' : 'text-slate-300'
                    }`}
                  >
                    {displayName(entry)}
                  </p>
                  {showFullNames && (
                    <p className="text-[10px] text-slate-500 capitalize">{entry.role}</p>
                  )}
                </div>
                <span
                  className={`text-sm font-semibold tabular-nums ${
                    isMe ? 'text-orange-300' : 'text-slate-400'
                  }`}
                >
                  {entry.total_points.toLocaleString()}
                </span>
              </li>
            );
          })}
        </ul>
      )}
    </div>
  );
}

/* ========== PAGE ========== */
export default function PointsPage() {
  const user = getUser();
  const totalPoints = user?.staff_points ?? 0;

  const [pointEvents, setPointEvents] = useState<PointEvent[]>([]);
  const [loadingPoints, setLoadingPoints] = useState(true);
  const [leaderboard, setLeaderboard] = useState<LeaderboardEntry[]>([]);
  const [loadingBoard, setLoadingBoard] = useState(true);
  const [activeTab, setActiveTab] = useState<'history' | 'leaderboard'>('history');

  // fetch points history
  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        const data = await api<{ points: PointEvent[] }>(
          `/labor/points/${user?.user_id}`,
        );
        if (!cancelled) setPointEvents(data.points ?? []);
      } catch {
        /* no points data */
      } finally {
        if (!cancelled) setLoadingPoints(false);
      }
    })();
    return () => { cancelled = true; };
  }, [user?.user_id]);

  // fetch leaderboard
  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        const data = await api<{ leaderboard: LeaderboardEntry[] }>('/labor/leaderboard');
        if (!cancelled) setLeaderboard(data.leaderboard ?? []);
      } catch {
        /* no leaderboard data */
      } finally {
        if (!cancelled) setLoadingBoard(false);
      }
    })();
    return () => { cancelled = true; };
  }, []);

  // computed: week points
  const weekPoints = useMemo(() => {
    const now = Date.now();
    const sevenDaysAgo = now - 7 * 24 * 60 * 60 * 1000;
    return pointEvents
      .filter((e) => new Date(e.created_at).getTime() >= sevenDaysAgo)
      .reduce((sum, e) => sum + e.points, 0);
  }, [pointEvents]);

  // computed: user rank
  const userRank = useMemo(() => {
    const entry = leaderboard.find((e) => e.user_id === user?.user_id);
    return entry?.rank ?? null;
  }, [leaderboard, user?.user_id]);

  return (
    <div className="min-h-screen bg-slate-900 pb-24">
      {/* hero */}
      <HeroPoints total={totalPoints} />

      {/* mini stats */}
      <div className="px-4 mb-5">
        <MiniStats weekPoints={weekPoints} streak={5} rank={userRank} />
      </div>

      {/* tab switcher */}
      <div className="px-4 mb-3">
        <div className="flex rounded-lg border border-white/10 bg-white/5 p-1">
          <button
            onClick={() => setActiveTab('history')}
            className={`flex-1 rounded-md py-2 text-xs font-semibold transition-colors ${
              activeTab === 'history'
                ? 'bg-orange-500 text-white'
                : 'text-slate-400 hover:text-slate-300'
            }`}
          >
            Points History
          </button>
          <button
            onClick={() => setActiveTab('leaderboard')}
            className={`flex-1 rounded-md py-2 text-xs font-semibold transition-colors ${
              activeTab === 'leaderboard'
                ? 'bg-orange-500 text-white'
                : 'text-slate-400 hover:text-slate-300'
            }`}
          >
            Location Leaderboard
          </button>
        </div>
      </div>

      {/* tab content */}
      <div className="px-4">
        {activeTab === 'history' ? (
          <Card>
            <PointsHistory events={pointEvents} loading={loadingPoints} />
          </Card>
        ) : (
          <Card>
            <div className="flex items-center gap-2 mb-3">
              <Trophy className="h-5 w-5 text-orange-400" />
              <h2 className="text-sm font-semibold text-slate-300 uppercase tracking-wide">
                Location Leaderboard
              </h2>
            </div>
            <Leaderboard
              entries={leaderboard}
              loading={loadingBoard}
              currentUserId={user?.user_id ?? ''}
              userRole={user?.role ?? 'staff'}
            />
          </Card>
        )}
      </div>
    </div>
  );
}
