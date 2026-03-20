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
import { DollarSign, Percent, Users, Clock } from 'lucide-react';

// ─── Helpers ────────────────────────────────────────────────────────────────

function cents(v: number): string {
  return `$${(v / 100).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
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

function eluColor(score: number): string {
  if (score <= 0.5) return 'bg-red-500';
  if (score <= 1.0) return 'bg-yellow-400';
  return 'bg-green-500';
}

function eluTextColor(score: number): string {
  if (score <= 0.5) return 'text-red-600';
  if (score <= 1.0) return 'text-yellow-600';
  return 'text-green-600';
}

type Trend = 'up' | 'down' | 'stable';

function TrendArrow({ trend }: { trend: Trend }) {
  if (trend === 'up') return <span className="text-green-500 font-bold">↑</span>;
  if (trend === 'down') return <span className="text-red-500 font-bold">↓</span>;
  return <span className="text-gray-400 font-bold">→</span>;
}

// ─── Overview Tab ────────────────────────────────────────────────────────────

const employeeColumns: Column<EmployeeDetail>[] = [
  {
    key: 'display_name',
    header: 'Employee',
    sortable: true,
    render: (r) => <span className="font-semibold text-gray-800">{r.display_name}</span>,
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

// ─── Staff Profiles Tab ──────────────────────────────────────────────────────

function ExpandedProfile({ profile }: { profile: EmployeeProfile }) {
  const { data: histData, isLoading } = usePointHistory(profile.employee_id);
  const events = histData?.events?.slice(0, 5) ?? [];

  return (
    <div className="p-4 bg-gray-50 border-t border-gray-200 grid grid-cols-1 md:grid-cols-2 gap-6">
      {/* ELU Bars */}
      <div>
        <h4 className="text-sm font-semibold text-gray-700 mb-3">ELU Station Ratings</h4>
        {Object.keys(profile.elu_ratings).length === 0 ? (
          <p className="text-sm text-gray-400 italic">No ELU ratings recorded</p>
        ) : (
          <div className="space-y-2">
            {Object.entries(profile.elu_ratings).map(([station, score]) => (
              <div key={station}>
                <div className="flex justify-between mb-0.5">
                  <span className="text-xs text-gray-600 capitalize">{station}</span>
                  <span className={`text-xs font-semibold ${eluTextColor(score)}`}>
                    {score.toFixed(1)}
                  </span>
                </div>
                <div className="w-full bg-gray-200 rounded-full h-2">
                  <div
                    className={`h-2 rounded-full ${eluColor(score)}`}
                    style={{ width: `${Math.min((score / 2.0) * 100, 100)}%` }}
                  />
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Recent Point Events */}
      <div>
        <h4 className="text-sm font-semibold text-gray-700 mb-3">Recent Point Events</h4>
        {isLoading ? (
          <LoadingSpinner />
        ) : events.length === 0 ? (
          <p className="text-sm text-gray-400 italic">No point history</p>
        ) : (
          <ul className="space-y-2">
            {events.map((ev) => (
              <li key={ev.event_id} className="flex items-start justify-between text-sm">
                <div>
                  <span className="font-medium text-gray-700 capitalize">{ev.reason}</span>
                  {ev.description && (
                    <p className="text-gray-400 text-xs">{ev.description}</p>
                  )}
                </div>
                <span
                  className={`ml-3 font-bold whitespace-nowrap ${
                    ev.points >= 0 ? 'text-green-600' : 'text-red-500'
                  }`}
                >
                  {ev.points >= 0 ? '+' : ''}{ev.points} pts
                </span>
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  );
}

function StaffProfilesTab({ locationId }: { locationId: string }) {
  const { data, isLoading, error, refetch } = useProfiles(locationId);
  const [expandedId, setExpandedId] = useState<string | null>(null);
  const profiles = data?.profiles ?? [];

  const profileColumns: Column<EmployeeProfile>[] = [
    {
      key: 'display_name',
      header: 'Employee',
      sortable: true,
      render: (r) => <span className="font-semibold text-gray-800">{r.display_name}</span>,
    },
    {
      key: 'role',
      header: 'Role',
      sortable: true,
      render: (r) => capitalize(r.role),
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
        <span className="text-gray-600">{r.certifications.length}</span>
      ),
    },
  ];

  if (error) {
    const msg = error instanceof Error ? error.message : 'Failed to load profiles';
    return <ErrorBanner message={msg} retry={() => void refetch()} />;
  }

  return (
    <div className="space-y-2">
      <DataTable
        columns={profileColumns}
        data={profiles}
        keyExtractor={(r) => r.employee_id}
        isLoading={isLoading}
        emptyTitle="No profiles found"
        emptyDescription="No staff profile data is available for this location."
        onRowClick={(r) => setExpandedId(expandedId === r.employee_id ? null : r.employee_id)}
      />
      {expandedId && (() => {
        const profile = profiles.find((p) => p.employee_id === expandedId);
        return profile ? (
          <div className="border border-gray-200 rounded-lg overflow-hidden shadow-sm">
            <div className="flex items-center justify-between px-4 py-2 bg-white border-b border-gray-200">
              <span className="font-semibold text-gray-800">{profile.display_name}</span>
              <button
                onClick={() => setExpandedId(null)}
                className="text-gray-400 hover:text-gray-600 text-lg leading-none"
              >
                ×
              </button>
            </div>
            <ExpandedProfile profile={profile} />
          </div>
        ) : null;
      })()}
    </div>
  );
}

// ─── Leaderboard Tab ─────────────────────────────────────────────────────────

const medalColors: Record<number, string> = {
  1: 'bg-yellow-400 text-yellow-900',
  2: 'bg-gray-300 text-gray-800',
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
      <p className="text-center text-gray-400 py-12">No leaderboard data available.</p>
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
                : 'border-gray-200 bg-white'
            } shadow-sm`}
          >
            {/* Rank badge */}
            <div
              className={`w-9 h-9 rounded-full flex items-center justify-center font-bold text-sm shrink-0 ${
                medal ?? 'bg-gray-100 text-gray-600'
              }`}
            >
              #{rank}
            </div>

            {/* Name + role */}
            <div className="flex-1 min-w-0">
              <p className="font-semibold text-gray-800 truncate">{entry.display_name}</p>
              <p className="text-xs text-gray-500 capitalize">{entry.role}</p>
            </div>

            {/* Points + trend */}
            <div className="flex items-center gap-2 shrink-0">
              <span className="text-lg font-bold text-gray-800">
                {entry.staff_points.toLocaleString()}
              </span>
              <span className="text-xs text-gray-500">pts</span>
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
      // Start with existing ratings or a default set
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
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Select Employee
        </label>
        <select
          className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm bg-white focus:outline-none focus:ring-2 focus:ring-red-400"
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
        <div className="bg-white border border-gray-200 rounded-xl p-5 shadow-sm space-y-5">
          <h3 className="font-semibold text-gray-800">
            ELU Ratings — {selectedProfile.display_name}
          </h3>

          {Object.entries(ratings).map(([station, score]) => (
            <div key={station}>
              <div className="flex justify-between mb-1">
                <label className="text-sm text-gray-700 capitalize">{station}</label>
                <span className={`text-sm font-bold ${eluTextColor(score)}`}>
                  {score.toFixed(1)}
                </span>
              </div>
              {/* Color indicator strip */}
              <div className="relative mb-1">
                <div className="w-full h-2 rounded-full bg-gradient-to-r from-red-400 via-yellow-300 to-green-500 opacity-30" />
                <div
                  className={`absolute top-0 h-2 rounded-full ${eluColor(score)} opacity-80`}
                  style={{ width: `${Math.min((score / 2.0) * 100, 100)}%` }}
                />
              </div>
              <input
                type="range"
                min={0}
                max={2}
                step={0.1}
                value={score}
                onChange={(e) => handleSliderChange(station, parseFloat(e.target.value))}
                className="w-full accent-red-500"
              />
              <div className="flex justify-between text-xs text-gray-400">
                <span>0.0</span>
                <span>1.0</span>
                <span>2.0</span>
              </div>
            </div>
          ))}

          {saveError && (
            <p className="text-sm text-red-600 bg-red-50 border border-red-200 rounded p-2">
              {saveError}
            </p>
          )}
          {saveSuccess && (
            <p className="text-sm text-green-700 bg-green-50 border border-green-200 rounded p-2">
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

  if (!locationId) return <LoadingSpinner fullPage />;

  const employees = employeesData?.employees ?? [];
  const overviewError = summaryError ?? employeesError;
  const overviewErrorMsg =
    overviewError instanceof Error ? overviewError.message : 'Failed to load labor data';

  return (
    <div className="space-y-6">
      {/* Page header */}
      <div>
        <h1 className="text-2xl font-bold text-gray-800">Labor Intelligence</h1>
        <p className="text-sm text-gray-500 mt-1">
          Workforce costs, hours, employee performance, and station readiness
        </p>
      </div>

      {/* Tab bar */}
      <div className="flex gap-1 border-b border-gray-200">
        {TABS.map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            className={`px-4 py-2 text-sm font-medium rounded-t-lg border-b-2 transition-colors ${
              activeTab === tab.id
                ? 'border-red-500 text-red-600 bg-red-50'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:bg-gray-50'
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
                iconColor="text-gray-600"
                bgTint="bg-gray-100"
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

          {/* Employee table */}
          <div>
            <h2 className="text-lg font-semibold text-gray-800 mb-3">Employee Detail</h2>
            <DataTable
              columns={employeeColumns}
              data={employees}
              keyExtractor={(r) => r.employee_id}
              isLoading={employeesLoading}
              emptyTitle="No employees found"
              emptyDescription="No employee data is available for this location and period."
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
