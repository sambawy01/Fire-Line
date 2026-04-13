import { useState } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { useLocationStore } from '../stores/location';
import {
  useLaborSummary,
  useLaborEmployees,
  useProfiles,
  useLeaderboard,
  usePointHistory,
} from '../hooks/useLabor';
import { laborApi } from '../lib/api';
import KPICard from '../components/ui/KPICard';
import DataTable from '../components/ui/DataTable';
import type { Column } from '../components/ui/DataTable';
import StatusBadge from '../components/ui/StatusBadge';
import LoadingSpinner from '../components/ui/LoadingSpinner';
import ErrorBanner from '../components/ui/ErrorBanner';
import type { EmployeeDetail, EmployeeProfile, LeaderboardEntry } from '../lib/api';
import { DollarSign, Percent, Users, Clock, ChevronDown, ChevronUp } from 'lucide-react';

// ─── Helpers ────────────────────────────────────────────────────────────────

function cents(v: number): string {
  return `EGP ${(v / 100).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
}

function capitalize(s: string): string {
  return s.charAt(0).toUpperCase() + s.slice(1).toLowerCase();
}

function statusVariant(status: string): 'success' | 'neutral' | 'critical' {
  if (status === 'active') return 'success';
  if (status === 'terminated') return 'critical';
  return 'neutral';
}

function eluAvg(ratings: Record<string, number>): number {
  const vals = Object.values(ratings);
  if (!vals.length) return 0;
  return vals.reduce((a, b) => a + b, 0) / vals.length;
}

/** Color for ELU bar fill — 0-1.5 red, 1.5-3.0 amber, 3.0-5.0 green (0-5 scale) */
function eluBarColor(score: number): string {
  if (score <= 1.5) return 'bg-red-500';
  if (score <= 3.0) return 'bg-amber-400';
  return 'bg-green-500';
}

function eluTextColor(score: number): string {
  if (score <= 1.5) return 'text-red-400';
  if (score <= 3.0) return 'text-amber-400';
  return 'text-green-400';
}

type Trend = 'up' | 'down' | 'stable';

function TrendArrow({ trend }: { trend: Trend }) {
  if (trend === 'up') return <span className="text-green-500 font-bold">↑</span>;
  if (trend === 'down') return <span className="text-red-500 font-bold">↓</span>;
  return <span className="text-slate-300 font-bold">→</span>;
}

const ELU_STATIONS = ['grill', 'saute', 'prep', 'expo', 'bar', 'fryer', 'dish'];

/** Format a timestamp into a human-readable relative time string */
function formatRelativeTime(dateStr: string): string {
  const diff = Date.now() - new Date(dateStr).getTime();
  const mins = Math.floor(diff / 60_000);
  if (mins < 1) return 'Just now';
  if (mins < 60) return `${mins} min ago`;
  const hours = Math.floor(mins / 60);
  if (hours < 24) return `${hours} hour${hours !== 1 ? 's' : ''} ago`;
  const days = Math.floor(hours / 24);
  if (days === 1) return 'Yesterday';
  if (days < 7) return `${days} days ago`;
  return `${Math.floor(days / 7)} week${Math.floor(days / 7) !== 1 ? 's' : ''} ago`;
}

// ─── Expanded Profile Panel ──────────────────────────────────────────────────

interface ExpandedProfilePanelProps {
  profile: EmployeeProfile;
  overviewEmployee?: EmployeeDetail;
  onClose: () => void;
}

function ExpandedProfilePanel({ profile, overviewEmployee, onClose }: ExpandedProfilePanelProps) {
  const { data: pointHistoryData, isLoading: pointsLoading } = usePointHistory(profile.employee_id);
  const pointEvents = pointHistoryData?.events ?? [];
  const avg = eluAvg(profile.elu_ratings);

  // Build station map — show all 7 canonical stations, default 0 if missing
  const stationScores: Record<string, number> = {};
  ELU_STATIONS.forEach((s) => {
    stationScores[s] = profile.elu_ratings[s] ?? 0;
  });

  return (
    <div className="bg-white/5 border border-white/10 rounded-xl p-5 mx-2 mb-2 space-y-6">
      {/* ── Header ── */}
      <div className="flex items-start justify-between gap-4">
        <div className="flex flex-col gap-1">
          <h3 className="text-xl font-bold text-white leading-tight">{profile.display_name}</h3>
          <div className="flex items-center gap-2 flex-wrap">
            <span className="px-2 py-0.5 rounded-full bg-white/10 text-slate-200 text-xs font-medium capitalize">
              {profile.role}
            </span>
            <StatusBadge variant={statusVariant(profile.status)}>
              {capitalize(profile.status)}
            </StatusBadge>
          </div>
        </div>
        <div className="flex flex-col items-end gap-1 shrink-0">
          <div className="flex items-center gap-1">
            <span className="text-2xl font-bold text-white">{profile.staff_points.toLocaleString()}</span>
            <span className="text-xs text-slate-400 self-end pb-0.5">pts</span>
          </div>
          <div className="flex items-center gap-1 text-sm text-slate-400">
            <TrendArrow trend={profile.points_trend} />
            <span className="text-xs capitalize">{profile.points_trend}</span>
          </div>
        </div>
        <button
          onClick={onClose}
          className="text-slate-400 hover:text-white text-xl leading-none shrink-0 self-start"
          aria-label="Collapse profile"
        >
          ×
        </button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {/* ── Section 1: ELU Ratings ── */}
        <div>
          <h4 className="text-sm font-semibold text-slate-200 mb-3 uppercase tracking-wide">
            ELU Station Ratings
            <span className={`ml-2 text-xs font-bold ${eluTextColor(avg)}`}>
              avg {avg.toFixed(2)}
            </span>
          </h4>
          <div className="space-y-2.5">
            {ELU_STATIONS.map((station) => {
              const score = stationScores[station];
              return (
                <div key={station} className="flex items-center gap-3">
                  <span className="text-xs text-slate-400 w-16 capitalize">{station}</span>
                  <div className="flex-1 h-3 bg-white/10 rounded-full overflow-hidden">
                    <div
                      className={`h-full rounded-full ${eluBarColor(score)}`}
                      style={{ width: `${(score / 5) * 100}%` }}
                    />
                  </div>
                  <span className="text-xs font-bold text-white w-8 text-right">{score.toFixed(1)}</span>
                </div>
              );
            })}
          </div>
        </div>

        {/* ── Section 2: Performance Metrics ── */}
        <div>
          <h4 className="text-sm font-semibold text-slate-200 mb-3 uppercase tracking-wide">
            Performance Metrics
          </h4>
          <div className="grid grid-cols-2 gap-3">
            <div className="bg-white/5 rounded-lg p-3">
              <p className="text-xs text-slate-400 mb-1">Staff Points</p>
              <p className="text-lg font-bold text-white">{profile.staff_points.toLocaleString()}</p>
              <div className="flex items-center gap-1 mt-0.5">
                <TrendArrow trend={profile.points_trend} />
                <span className="text-xs text-slate-400 capitalize">{profile.points_trend}</span>
              </div>
            </div>
            {overviewEmployee ? (
              <>
                <div className="bg-white/5 rounded-lg p-3">
                  <p className="text-xs text-slate-400 mb-1">Shifts</p>
                  <p className="text-lg font-bold text-white">{overviewEmployee.shift_count}</p>
                </div>
                <div className="bg-white/5 rounded-lg p-3">
                  <p className="text-xs text-slate-400 mb-1">Hours Worked</p>
                  <p className="text-lg font-bold text-white">{overviewEmployee.hours_worked.toFixed(1)}</p>
                </div>
                <div className="bg-white/5 rounded-lg p-3">
                  <p className="text-xs text-slate-400 mb-1">Avg Hrs / Shift</p>
                  <p className="text-lg font-bold text-white">{overviewEmployee.avg_hours_per_shift.toFixed(1)}</p>
                </div>
              </>
            ) : (
              <div className="bg-white/5 rounded-lg p-3 col-span-1">
                <p className="text-xs text-slate-400 mb-1">ELU Average</p>
                <p className={`text-lg font-bold ${eluTextColor(avg)}`}>{avg.toFixed(2)}</p>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* ── Section 3: Point History ── */}
      <div>
        <h4 className="text-sm font-semibold text-slate-200 mb-3 uppercase tracking-wide">
          Recent Point History
        </h4>
        {pointsLoading ? (
          <div className="flex justify-center py-4">
            <LoadingSpinner />
          </div>
        ) : pointEvents.length === 0 ? (
          <p className="text-sm text-slate-500 italic">No point history recorded</p>
        ) : (
          <ul className="space-y-2">
            {pointEvents.map((ev) => (
              <li key={ev.event_id} className="flex items-center gap-3 text-sm">
                <span
                  className={`font-bold w-10 text-right shrink-0 ${
                    ev.points >= 0 ? 'text-green-400' : 'text-red-400'
                  }`}
                >
                  {ev.points >= 0 ? '+' : ''}{ev.points}
                </span>
                <div className="flex-1 min-w-0">
                  <span className="font-medium text-slate-200">{ev.reason}</span>
                  {ev.description && (
                    <>
                      <span className="text-slate-400 mx-1">—</span>
                      <span className="text-slate-400">{ev.description}</span>
                    </>
                  )}
                </div>
                <span className="text-xs text-slate-500 shrink-0">{formatRelativeTime(ev.created_at)}</span>
              </li>
            ))}
          </ul>
        )}
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {/* ── Section 4: Certifications ── */}
        <div>
          <h4 className="text-sm font-semibold text-slate-200 mb-3 uppercase tracking-wide">
            Certifications
          </h4>
          {profile.certifications.length === 0 ? (
            <p className="text-sm text-slate-500 italic">No certifications recorded</p>
          ) : (
            <div className="flex flex-wrap gap-2">
              {profile.certifications.map((cert) => (
                <span
                  key={cert}
                  className="px-2.5 py-1 rounded-full bg-blue-500/20 border border-blue-400/30 text-blue-300 text-xs font-medium"
                >
                  {cert}
                </span>
              ))}
            </div>
          )}
        </div>

        {/* ── Section 5: Availability ── */}
        <div>
          <h4 className="text-sm font-semibold text-slate-200 mb-3 uppercase tracking-wide">
            Availability
          </h4>
          {!profile.availability || Object.keys(profile.availability).length === 0 ? (
            <p className="text-sm text-slate-500 italic">Availability not set</p>
          ) : (
            <div className="space-y-1.5">
              {Object.entries(profile.availability).map(([day, times]) => (
                <div key={day} className="flex items-center gap-3">
                  <span className="text-xs text-slate-400 w-20 capitalize">{day}</span>
                  <span className="text-xs text-slate-200">
                    {Array.isArray(times) ? times.join(', ') : String(times)}
                  </span>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

// ─── Overview Tab ────────────────────────────────────────────────────────────

interface OverviewTabProps {
  employees: EmployeeDetail[];
  isLoading: boolean;
  profiles: EmployeeProfile[];
}

function OverviewEmployeeTable({ employees, isLoading, profiles }: OverviewTabProps) {
  const [expandedId, setExpandedId] = useState<string | null>(null);

  const employeeColumns: Column<EmployeeDetail>[] = [
    {
      key: 'display_name',
      header: 'Employee',
      sortable: true,
      render: (r) => (
        <div className="flex items-center gap-2">
          <span className="font-semibold text-white">{r.display_name}</span>
          {expandedId === r.employee_id ? (
            <ChevronUp className="w-3 h-3 text-slate-400" />
          ) : (
            <ChevronDown className="w-3 h-3 text-slate-400" />
          )}
        </div>
      ),
    },
    {
      key: 'role',
      header: 'Role',
      sortable: true,
      render: (r) => capitalize(r.role),
    },
    {
      key: 'status',
      header: 'Status',
      sortable: true,
      render: (r) => (
        <StatusBadge variant={statusVariant(r.status)}>
          {capitalize(r.status)}
        </StatusBadge>
      ),
    },
    { key: 'shift_count', header: 'Shifts', align: 'right', sortable: true },
    {
      key: 'hours_worked',
      header: 'Hours',
      align: 'right',
      sortable: true,
      render: (r) => r.hours_worked.toFixed(1),
    },
    {
      key: 'labor_cost',
      header: 'Cost ($)',
      align: 'right',
      sortable: true,
      render: (r) => cents(r.labor_cost),
    },
    {
      key: 'avg_hours_per_shift',
      header: 'Avg Hrs/Shift',
      align: 'right',
      sortable: true,
      render: (r) => r.avg_hours_per_shift.toFixed(1),
    },
    {
      key: 'hourly_rate',
      header: 'Rate ($/hr)',
      align: 'right',
      sortable: true,
      render: (r) => cents(r.hourly_rate),
    },
  ];

  return (
    <div className="space-y-0">
      <DataTable
        columns={employeeColumns}
        data={employees}
        keyExtractor={(r) => r.employee_id}
        isLoading={isLoading}
        emptyTitle="No employees found"
        emptyDescription="No employee data is available for this location and period."
        onRowClick={(r) =>
          setExpandedId(expandedId === r.employee_id ? null : r.employee_id)
        }
        expandedRowId={expandedId ?? undefined}
        renderExpanded={(r) => {
          const profile = profiles.find((p) => p.employee_id === r.employee_id);
          if (!profile) return null;
          return (
            <ExpandedProfilePanel
              profile={profile}
              overviewEmployee={r}
              onClose={() => setExpandedId(null)}
            />
          );
        }}
      />
    </div>
  );
}

// ─── Staff Profiles Tab ──────────────────────────────────────────────────────

function StaffProfilesTab({ locationId }: { locationId: string }) {
  const { data, isLoading, error, refetch } = useProfiles(locationId);
  const [expandedId, setExpandedId] = useState<string | null>(null);
  const profiles = data?.profiles ?? [];

  const profileColumns: Column<EmployeeProfile>[] = [
    {
      key: 'display_name',
      header: 'Employee',
      sortable: true,
      render: (r) => (
        <div className="flex items-center gap-2">
          <span className="font-semibold text-white">{r.display_name}</span>
          {expandedId === r.employee_id ? (
            <ChevronUp className="w-3 h-3 text-slate-400" />
          ) : (
            <ChevronDown className="w-3 h-3 text-slate-400" />
          )}
        </div>
      ),
    },
    {
      key: 'role',
      header: 'Role',
      sortable: true,
      render: (r) => capitalize(r.role),
    },
    {
      key: 'status',
      header: 'Status',
      sortable: true,
      render: (r) => (
        <StatusBadge variant={statusVariant(r.status)}>
          {capitalize(r.status)}
        </StatusBadge>
      ),
    },
    {
      key: 'elu_ratings',
      header: 'ELU Avg',
      align: 'right',
      sortable: false,
      render: (r) => {
        const avg = eluAvg(r.elu_ratings);
        return (
          <span className={`font-semibold ${eluTextColor(avg)}`}>
            {avg.toFixed(2)}
          </span>
        );
      },
    },
    {
      key: 'staff_points',
      header: 'Points',
      align: 'right',
      sortable: true,
      render: (r) => r.staff_points.toLocaleString(),
    },
    {
      key: 'points_trend',
      header: 'Trend',
      align: 'center',
      sortable: false,
      render: (r) => <TrendArrow trend={r.points_trend} />,
    },
    {
      key: 'certifications',
      header: 'Certs',
      align: 'right',
      sortable: false,
      render: (r) => (
        <span className="text-slate-300">{r.certifications.length}</span>
      ),
    },
  ];

  if (error) {
    const msg = error instanceof Error ? error.message : 'Failed to load profiles';
    return <ErrorBanner message={msg} retry={() => void refetch()} />;
  }

  return (
    <div className="space-y-0">
      <DataTable
        columns={profileColumns}
        data={profiles}
        keyExtractor={(r) => r.employee_id}
        isLoading={isLoading}
        emptyTitle="No profiles found"
        emptyDescription="No staff profile data is available for this location."
        onRowClick={(r) =>
          setExpandedId(expandedId === r.employee_id ? null : r.employee_id)
        }
        expandedRowId={expandedId ?? undefined}
        renderExpanded={(r) => (
          <ExpandedProfilePanel
            profile={r}
            onClose={() => setExpandedId(null)}
          />
        )}
      />
    </div>
  );
}

// ─── Leaderboard Tab ─────────────────────────────────────────────────────────

const medalColors: Record<number, string> = {
  1: 'bg-yellow-400 text-yellow-900',
  2: 'bg-gray-300 text-white',
  3: 'bg-amber-600 text-amber-100',
};

function LeaderboardTab({ locationId }: { locationId: string }) {
  const { data, isLoading, error, refetch } = useLeaderboard(locationId);
  const entries = data?.leaderboard ?? [];

  if (isLoading) {
    return (
      <div className="flex justify-center py-12">
        <LoadingSpinner />
      </div>
    );
  }

  if (error) {
    const msg = error instanceof Error ? error.message : 'Failed to load leaderboard';
    return <ErrorBanner message={msg} retry={() => void refetch()} />;
  }

  if (!entries.length) {
    return (
      <p className="text-center text-slate-300 py-12">No leaderboard data available.</p>
    );
  }

  return (
    <div className="space-y-3 max-w-2xl">
      {entries.map((entry: LeaderboardEntry, idx: number) => {
        const rank = idx + 1;
        const medal = medalColors[rank];
        return (
          <div
            key={entry.employee_id}
            className={`flex items-center gap-4 p-4 rounded-xl border ${
              rank <= 3
                ? 'border-yellow-200 bg-yellow-50'
                : 'border-white/10 bg-white/5'
            } shadow-sm`}
          >
            {/* Rank badge */}
            <div
              className={`w-9 h-9 rounded-full flex items-center justify-center font-bold text-sm shrink-0 ${
                medal ?? 'bg-white/10 text-slate-300'
              }`}
            >
              #{rank}
            </div>

            {/* Name + role */}
            <div className="flex-1 min-w-0">
              <p className="font-semibold text-white truncate">{entry.display_name}</p>
              <p className="text-xs text-slate-400 capitalize">{entry.role}</p>
            </div>

            {/* Points + trend */}
            <div className="flex items-center gap-2 shrink-0">
              <span className="text-lg font-bold text-white">
                {entry.staff_points.toLocaleString()}
              </span>
              <span className="text-xs text-slate-400">pts</span>
              <TrendArrow trend={entry.points_trend} />
            </div>
          </div>
        );
      })}
    </div>
  );
}

// ─── ELU Management Tab ──────────────────────────────────────────────────────

function ELUManagementTab({ locationId }: { locationId: string }) {
  const { data, isLoading, error, refetch } = useProfiles(locationId);
  const qc = useQueryClient();

  const profiles = data?.profiles ?? [];
  const [selectedId, setSelectedId] = useState<string>('');
  const [ratings, setRatings] = useState<Record<string, number>>({});
  const [saving, setSaving] = useState(false);
  const [saveError, setSaveError] = useState<string | null>(null);
  const [saveSuccess, setSaveSuccess] = useState(false);

  const selectedProfile = profiles.find((p) => p.employee_id === selectedId);

  const handleSelectEmployee = (id: string) => {
    setSelectedId(id);
    setSaveError(null);
    setSaveSuccess(false);
    const profile = profiles.find((p) => p.employee_id === id);
    if (profile) {
      const defaults: Record<string, number> =
        Object.keys(profile.elu_ratings).length > 0
          ? { ...profile.elu_ratings }
          : { grill: 0, fryer: 0, prep: 0, expo: 0, cashier: 0 };
      setRatings(defaults);
    }
  };

  const handleSliderChange = (station: string, value: number) => {
    setRatings((prev) => ({ ...prev, [station]: value }));
  };

  const handleSave = async () => {
    if (!selectedId) return;
    setSaving(true);
    setSaveError(null);
    setSaveSuccess(false);
    try {
      await laborApi.updateELU(selectedId, ratings);
      setSaveSuccess(true);
      void qc.invalidateQueries({ queryKey: ['labor', 'profiles', locationId] });
    } catch (err) {
      setSaveError(err instanceof Error ? err.message : 'Failed to save ELU ratings');
    } finally {
      setSaving(false);
    }
  };

  if (isLoading) {
    return (
      <div className="flex justify-center py-12">
        <LoadingSpinner />
      </div>
    );
  }

  if (error) {
    const msg = error instanceof Error ? error.message : 'Failed to load profiles';
    return <ErrorBanner message={msg} retry={() => void refetch()} />;
  }

  return (
    <div className="max-w-xl space-y-6">
      {/* Employee selector */}
      <div>
        <label className="block text-sm font-medium text-slate-200 mb-1">
          Select Employee
        </label>
        <select
          className="w-full rounded-lg border border-white/20 px-3 py-2 text-sm bg-white/10 text-white focus:outline-none focus:ring-2 focus:ring-red-400"
          value={selectedId}
          onChange={(e) => handleSelectEmployee(e.target.value)}
        >
          <option value="">— choose an employee —</option>
          {profiles.map((p) => (
            <option key={p.employee_id} value={p.employee_id}>
              {p.display_name} ({capitalize(p.role)})
            </option>
          ))}
        </select>
      </div>

      {/* Sliders */}
      {selectedProfile && (
        <div className="bg-white/5 border border-white/10 rounded-xl p-5 shadow-sm space-y-5">
          <h3 className="font-semibold text-white">
            ELU Ratings — {selectedProfile.display_name}
          </h3>

          {Object.entries(ratings).map(([station, score]) => (
            <div key={station}>
              <div className="flex justify-between mb-1">
                <label className="text-sm text-slate-200 capitalize">{station}</label>
                <span className={`text-sm font-bold ${eluTextColor(score)}`}>
                  {score.toFixed(1)}
                </span>
              </div>
              {/* Color indicator strip */}
              <div className="relative mb-1">
                <div className="w-full h-2 rounded-full bg-gradient-to-r from-red-400 via-yellow-300 to-green-500 opacity-30" />
                <div
                  className={`absolute top-0 h-2 rounded-full ${eluBarColor(score)} opacity-80`}
                  style={{ width: `${Math.min((score / 5.0) * 100, 100)}%` }}
                />
              </div>
              <input
                type="range"
                min={0}
                max={5}
                step={0.1}
                value={score}
                onChange={(e) => handleSliderChange(station, parseFloat(e.target.value))}
                className="w-full accent-red-500"
              />
              <div className="flex justify-between text-xs text-slate-300">
                <span>0.0</span>
                <span>2.5</span>
                <span>5.0</span>
              </div>
            </div>
          ))}

          {saveError && (
            <p className="text-sm text-red-400 bg-red-900/20 border border-red-500/30 rounded p-2">
              {saveError}
            </p>
          )}
          {saveSuccess && (
            <p className="text-sm text-green-400 bg-green-900/20 border border-green-500/30 rounded p-2">
              ELU ratings saved successfully.
            </p>
          )}

          <button
            onClick={() => void handleSave()}
            disabled={saving}
            className="w-full py-2 px-4 rounded-lg bg-red-600 hover:bg-red-700 disabled:opacity-50 text-white font-semibold text-sm transition-colors"
          >
            {saving ? 'Saving…' : 'Save ELU Ratings'}
          </button>
        </div>
      )}
    </div>
  );
}

// ─── Main Page ────────────────────────────────────────────────────────────────

type Tab = 'overview' | 'profiles' | 'leaderboard' | 'elu';

const TABS: { id: Tab; label: string }[] = [
  { id: 'overview', label: 'Overview' },
  { id: 'profiles', label: 'Staff Profiles' },
  { id: 'leaderboard', label: 'Leaderboard' },
  { id: 'elu', label: 'ELU Management' },
];

export default function LaborPage() {
  const locationId = useLocationStore((s) => s.selectedLocationId);
  const [activeTab, setActiveTab] = useState<Tab>('overview');

  const {
    data: summary,
    isLoading: summaryLoading,
    error: summaryError,
    refetch: refetchSummary,
  } = useLaborSummary(locationId);

  const {
    data: employeesData,
    isLoading: employeesLoading,
    error: employeesError,
    refetch: refetchEmployees,
  } = useLaborEmployees(locationId);

  // Fetch profiles here so Overview tab can look up ELU + availability data
  const { data: profilesData } = useProfiles(locationId);
  const profiles = profilesData?.profiles ?? [];

  if (!locationId) return <LoadingSpinner fullPage />;

  const employees = employeesData?.employees ?? [];
  const overviewError = summaryError ?? employeesError;
  const overviewErrorMsg =
    overviewError instanceof Error ? overviewError.message : 'Failed to load labor data';

  return (
    <div className="space-y-6">
      {/* Page header */}
      <div>
        <h1 className="text-2xl font-bold text-white">Labor Intelligence</h1>
        <p className="text-sm text-slate-400 mt-1">
          Workforce costs, hours, employee performance, and station readiness
        </p>
      </div>

      {/* Tab bar */}
      <div className="flex gap-1 border-b border-white/10">
        {TABS.map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            className={`px-4 py-2 text-sm font-medium rounded-t-lg border-b-2 transition-colors ${
              activeTab === tab.id
                ? 'border-red-500 text-red-400 bg-red-500/10'
                : 'border-transparent text-slate-400 hover:text-slate-200 hover:bg-white/5'
            }`}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {/* Tab content */}
      {activeTab === 'overview' && (
        <div className="space-y-8">
          {overviewError && (
            <ErrorBanner
              message={overviewErrorMsg}
              retry={() => {
                void refetchSummary();
                void refetchEmployees();
              }}
            />
          )}

          {/* KPI Cards */}
          {summaryLoading ? (
            <div className="flex justify-center py-8">
              <LoadingSpinner />
            </div>
          ) : (
            <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
              <KPICard
                label="Labor Cost"
                value={summary ? cents(summary.total_labor_cost) : '$—'}
                icon={DollarSign}
                iconColor="text-red-600"
                bgTint="bg-red-50"
              />
              <KPICard
                label="Labor Cost %"
                value={summary ? `${summary.labor_cost_pct.toFixed(1)}%` : '—'}
                icon={Percent}
                iconColor="text-blue-600"
                bgTint="bg-blue-50"
              />
              <KPICard
                label="Active Employees"
                value={summary ? String(summary.employee_count) : '—'}
                icon={Users}
                iconColor="text-slate-300"
                bgTint="bg-white/10"
              />
              <KPICard
                label="Total Hours"
                value={summary ? summary.total_hours.toFixed(1) : '—'}
                icon={Clock}
                iconColor="text-purple-600"
                bgTint="bg-purple-50"
              />
            </div>
          )}

          {/* Employee table — expandable rows */}
          <div>
            <h2 className="text-lg font-semibold text-white mb-3">
              Employee Detail
              <span className="ml-2 text-xs font-normal text-slate-400">
                — click any row to expand profile
              </span>
            </h2>
            <OverviewEmployeeTable
              employees={employees}
              isLoading={employeesLoading}
              profiles={profiles}
            />
          </div>
        </div>
      )}

      {activeTab === 'profiles' && <StaffProfilesTab locationId={locationId} />}
      {activeTab === 'leaderboard' && <LeaderboardTab locationId={locationId} />}
      {activeTab === 'elu' && <ELUManagementTab locationId={locationId} />}
    </div>
  );
}
