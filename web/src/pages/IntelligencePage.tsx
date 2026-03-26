import { useState, useMemo, useCallback } from 'react';
import {
  Shield,
  RefreshCw,
  AlertTriangle,
  DollarSign,
  Clock,
  Package,
  BarChart,
  ChevronDown,
  ChevronUp,
  X,
  Users,
  HeartPulse,
  Eye,
  Zap,
} from 'lucide-react';
import { useAuthStore } from '../stores/auth';
import { useCEOBriefing, useIntelligenceAnomalies, useResolveAnomaly } from '../hooks/useIntelligence';
import KPICard from '../components/ui/KPICard';
import LoadingSpinner from '../components/ui/LoadingSpinner';
import ErrorBanner from '../components/ui/ErrorBanner';
import Modal from '../components/ui/Modal';
import type { IntelligenceAnomaly } from '../lib/api';

// ── Helpers ─────────────────────────────────────────────────────────────────

function scoreColor(score: number): string {
  if (score >= 70) return 'text-emerald-400';
  if (score >= 40) return 'text-amber-400';
  return 'text-red-400';
}

function scoreBg(score: number): string {
  if (score >= 70) return 'bg-emerald-500/20';
  if (score >= 40) return 'bg-amber-500/20';
  return 'bg-red-500/20';
}

function riskBadge(level: string) {
  const colors: Record<string, string> = {
    low: 'bg-emerald-500/20 text-emerald-400 border-emerald-500/30',
    medium: 'bg-amber-500/20 text-amber-400 border-amber-500/30',
    high: 'bg-red-500/20 text-red-400 border-red-500/30',
  };
  return (
    <span className={`inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-semibold ${colors[level] || colors.low}`}>
      {level.toUpperCase()}
    </span>
  );
}

function severityBadge(severity: string) {
  const map: Record<string, string> = {
    info: 'bg-blue-500/20 text-blue-400 border-blue-500/30',
    warning: 'bg-amber-500/20 text-amber-400 border-amber-500/30',
    critical: 'bg-red-500/20 text-red-400 border-red-500/30',
  };
  return (
    <span className={`inline-flex items-center gap-1 rounded-full border px-2.5 py-0.5 text-xs font-semibold ${map[severity] || map.info}`}>
      {severity === 'critical' && <span className="relative flex h-2 w-2"><span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-red-400 opacity-75" /><span className="relative inline-flex rounded-full h-2 w-2 bg-red-500" /></span>}
      {severity.toUpperCase()}
    </span>
  );
}

function statusBadge(status: string) {
  const map: Record<string, string> = {
    open: 'bg-red-500/20 text-red-400 border-red-500/30',
    investigating: 'bg-amber-500/20 text-amber-400 border-amber-500/30',
    resolved: 'bg-emerald-500/20 text-emerald-400 border-emerald-500/30',
    false_positive: 'bg-slate-500/20 text-slate-400 border-slate-500/30',
  };
  return (
    <span className={`inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-semibold ${map[status] || map.open}`}>
      {status.replace('_', ' ').toUpperCase()}
    </span>
  );
}

function anomalyIcon(type: string) {
  const map: Record<string, typeof AlertTriangle> = {
    void_pattern: AlertTriangle,
    cash_variance: DollarSign,
    clock_irregularity: Clock,
    shrinkage: Package,
    transaction_pattern: BarChart,
  };
  const Icon = map[type] || AlertTriangle;
  return <Icon className="h-5 w-5 text-slate-300" />;
}

function timeAgo(iso: string): string {
  const diff = Date.now() - new Date(iso).getTime();
  const mins = Math.floor(diff / 60_000);
  if (mins < 1) return 'just now';
  if (mins < 60) return `${mins}m ago`;
  const hrs = Math.floor(mins / 60);
  if (hrs < 24) return `${hrs}h ago`;
  return `${Math.floor(hrs / 24)}d ago`;
}

function fmtPct(v: number): string {
  return `${v.toFixed(1)}%`;
}

// ── Anomaly filter tabs ─────────────────────────────────────────────────────

type AnomalyFilter = 'all' | 'open' | 'investigating' | 'resolved';

const FILTER_TABS: { id: AnomalyFilter; label: string }[] = [
  { id: 'all', label: 'All' },
  { id: 'open', label: 'Open' },
  { id: 'investigating', label: 'Investigating' },
  { id: 'resolved', label: 'Resolved' },
];

// ── Component ───────────────────────────────────────────────────────────────

export default function IntelligencePage() {
  const role = useAuthStore((s) => s.role);
  const canResolve = role === 'ops_director' || role === 'owner' || role === 'ceo';

  const {
    data: briefing,
    isLoading: briefingLoading,
    error: briefingError,
    refetch: refetchBriefing,
  } = useCEOBriefing();

  const {
    data: anomalyData,
    isLoading: anomalyLoading,
    error: anomalyError,
    refetch: refetchAnomalies,
  } = useIntelligenceAnomalies();

  const resolveMutation = useResolveAnomaly();

  const [anomalyFilter, setAnomalyFilter] = useState<AnomalyFilter>('all');
  const [expandedAnomaly, setExpandedAnomaly] = useState<string | null>(null);
  const [resolveModal, setResolveModal] = useState<IntelligenceAnomaly | null>(null);
  const [resolveStatus, setResolveStatus] = useState('resolved');
  const [resolveNotes, setResolveNotes] = useState('');

  const anomalies = anomalyData?.anomalies ?? [];

  const filteredAnomalies = useMemo(() => {
    if (anomalyFilter === 'all') return anomalies;
    return anomalies.filter((a) => a.status === anomalyFilter);
  }, [anomalies, anomalyFilter]);

  const handleRefresh = useCallback(() => {
    refetchBriefing();
    refetchAnomalies();
  }, [refetchBriefing, refetchAnomalies]);

  const handleResolve = async () => {
    if (!resolveModal) return;
    await resolveMutation.mutateAsync({
      id: resolveModal.anomaly_id,
      status: resolveStatus,
      notes: resolveNotes,
    });
    setResolveModal(null);
    setResolveStatus('resolved');
    setResolveNotes('');
  };

  const isLoading = briefingLoading || anomalyLoading;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div className="flex items-center gap-3">
          <div className="bg-violet-500/20 p-2.5 rounded-lg">
            <Shield className="h-6 w-6 text-violet-400" />
          </div>
          <div>
            <h1 className="text-2xl font-bold text-white">Intelligence Center</h1>
            <p className="text-sm text-slate-400">Surveillance &amp; anomaly detection</p>
          </div>
        </div>

        <button
          onClick={handleRefresh}
          disabled={isLoading}
          className="flex items-center gap-2 bg-white/5 hover:bg-white/10 border border-white/10 text-white text-sm font-medium px-4 py-2 rounded-lg transition-colors disabled:opacity-50"
        >
          <RefreshCw className={`h-4 w-4 ${isLoading ? 'animate-spin' : ''}`} />
          Refresh
        </button>
      </div>

      {/* Errors */}
      {briefingError && (
        <ErrorBanner
          message={(briefingError as Error).message || 'Failed to load CEO briefing'}
          retry={() => refetchBriefing()}
        />
      )}
      {anomalyError && (
        <ErrorBanner
          message={(anomalyError as Error).message || 'Failed to load anomalies'}
          retry={() => refetchAnomalies()}
        />
      )}

      {isLoading && <LoadingSpinner fullPage />}

      {/* ── CEO Briefing Section ──────────────────────────────────────────────── */}
      {briefing && (
        <>
          {/* Score Cards */}
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
            <div className={`bg-white/5 rounded-xl border border-white/10 p-5 flex items-start gap-4`}>
              <div className={`${scoreBg(briefing.fraud_risk_score)} p-3 rounded-lg`}>
                <Eye className={`h-6 w-6 ${scoreColor(briefing.fraud_risk_score)}`} />
              </div>
              <div>
                <p className="text-sm text-slate-400">Fraud Risk Score</p>
                <p className={`text-2xl font-bold mt-0.5 ${scoreColor(briefing.fraud_risk_score)}`}>
                  {briefing.fraud_risk_score}
                </p>
              </div>
            </div>

            <div className="bg-white/5 rounded-xl border border-white/10 p-5 flex items-start gap-4">
              <div className={`${scoreBg(briefing.workforce_health_score)} p-3 rounded-lg`}>
                <HeartPulse className={`h-6 w-6 ${scoreColor(briefing.workforce_health_score)}`} />
              </div>
              <div>
                <p className="text-sm text-slate-400">Workforce Health</p>
                <p className={`text-2xl font-bold mt-0.5 ${scoreColor(briefing.workforce_health_score)}`}>
                  {briefing.workforce_health_score}
                </p>
              </div>
            </div>

            <div className="bg-white/5 rounded-xl border border-white/10 p-5 flex items-start gap-4">
              <div className={`${briefing.open_anomalies > 0 ? 'bg-red-500/20' : 'bg-slate-500/20'} p-3 rounded-lg`}>
                <AlertTriangle className={`h-6 w-6 ${briefing.open_anomalies > 0 ? 'text-red-400' : 'text-slate-400'}`} />
              </div>
              <div>
                <p className="text-sm text-slate-400">Open Anomalies</p>
                <p className={`text-2xl font-bold mt-0.5 ${briefing.open_anomalies > 0 ? 'text-red-400' : 'text-white'}`}>
                  {briefing.open_anomalies}
                </p>
              </div>
            </div>

            <div className="bg-white/5 rounded-xl border border-white/10 p-5 flex items-start gap-4">
              <div className={`${briefing.critical_anomalies > 0 ? 'bg-red-500/20' : 'bg-slate-500/20'} p-3 rounded-lg`}>
                <Zap className={`h-6 w-6 ${briefing.critical_anomalies > 0 ? 'text-red-400' : 'text-slate-400'}`} />
              </div>
              <div>
                <p className="text-sm text-slate-400">Critical Anomalies</p>
                <div className="flex items-center gap-2">
                  <p className={`text-2xl font-bold mt-0.5 ${briefing.critical_anomalies > 0 ? 'text-red-400' : 'text-white'}`}>
                    {briefing.critical_anomalies}
                  </p>
                  {briefing.critical_anomalies > 0 && (
                    <span className="relative flex h-3 w-3">
                      <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-red-400 opacity-75" />
                      <span className="relative inline-flex rounded-full h-3 w-3 bg-red-500" />
                    </span>
                  )}
                </div>
              </div>
            </div>
          </div>

          {/* Location Risk Map */}
          <div className="bg-white/5 rounded-xl border border-white/10 p-5">
            <h2 className="text-lg font-semibold text-white mb-4">Location Risk Map</h2>
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
              {briefing.location_risks.map((loc) => (
                <div key={loc.location_id} className="bg-white/5 rounded-lg border border-white/10 p-4 space-y-3">
                  <div className="flex items-center justify-between">
                    <h3 className="text-sm font-semibold text-white truncate">{loc.location_name}</h3>
                    {riskBadge(loc.risk_level)}
                  </div>
                  <div className="space-y-1.5 text-xs">
                    <div className="flex justify-between">
                      <span className="text-slate-400">Anomalies</span>
                      <span className={loc.anomaly_count > 0 ? 'text-red-400 font-medium' : 'text-slate-300'}>
                        {loc.anomaly_count}
                      </span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-slate-400">Task Completion</span>
                      <span className="text-slate-300">{fmtPct(loc.task_completion_pct)}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-slate-400">Attendance</span>
                      <span className="text-slate-300">{fmtPct(loc.attendance_pct)}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-slate-400">Labor Cost</span>
                      <span className="text-slate-300">{fmtPct(loc.labor_cost_pct)}</span>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* Turnover Risk + Staffing Alerts Row */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {/* Turnover Risk Panel */}
            <div className="bg-white/5 rounded-xl border border-white/10 p-5">
              <h2 className="text-lg font-semibold text-white mb-4">Turnover Risk</h2>
              {briefing.turnover_risks.length === 0 ? (
                <p className="text-sm text-slate-400">No at-risk employees detected.</p>
              ) : (
                <div className="space-y-3 max-h-80 overflow-y-auto">
                  {[...briefing.turnover_risks]
                    .sort((a, b) => b.risk_score - a.risk_score)
                    .map((emp) => (
                      <div key={emp.employee_id} className="bg-white/5 rounded-lg p-3 space-y-2">
                        <div className="flex items-center justify-between">
                          <div>
                            <p className="text-sm font-medium text-white">{emp.display_name}</p>
                            <p className="text-xs text-slate-400">
                              {emp.role} &middot; {emp.location_name}
                            </p>
                          </div>
                          <span className={`text-sm font-bold ${scoreColor(100 - emp.risk_score)}`}>
                            {emp.risk_score}
                          </span>
                        </div>
                        {/* Risk bar */}
                        <div className="w-full bg-white/10 rounded-full h-2">
                          <div
                            className={`h-2 rounded-full transition-all ${
                              emp.risk_score >= 70
                                ? 'bg-red-500'
                                : emp.risk_score >= 40
                                ? 'bg-amber-500'
                                : 'bg-emerald-500'
                            }`}
                            style={{ width: `${Math.min(emp.risk_score, 100)}%` }}
                          />
                        </div>
                        {/* Signal badges */}
                        <div className="flex flex-wrap gap-1">
                          {emp.signals.map((signal) => (
                            <span
                              key={signal}
                              className="inline-flex items-center rounded-full bg-white/10 px-2 py-0.5 text-[10px] text-slate-300"
                            >
                              {signal}
                            </span>
                          ))}
                        </div>
                      </div>
                    ))}
                </div>
              )}
            </div>

            {/* Staffing Alerts Panel */}
            <div className="bg-white/5 rounded-xl border border-white/10 p-5">
              <h2 className="text-lg font-semibold text-white mb-4">Staffing Alerts</h2>
              {briefing.staffing_gaps.length === 0 ? (
                <p className="text-sm text-slate-400">No upcoming staffing gaps.</p>
              ) : (
                <div className="space-y-3 max-h-80 overflow-y-auto">
                  {briefing.staffing_gaps.map((gap, i) => (
                    <div key={i} className="bg-white/5 rounded-lg p-4 space-y-2">
                      <div className="flex items-center justify-between">
                        <h3 className="text-sm font-semibold text-white">{gap.location_name}</h3>
                        <span className="text-xs text-slate-400">{gap.date}</span>
                      </div>
                      <div className="flex items-center gap-4 text-sm">
                        <div>
                          <span className="text-slate-400">Scheduled: </span>
                          <span className="text-white font-medium">{gap.scheduled}</span>
                        </div>
                        <div>
                          <span className="text-slate-400">Required: </span>
                          <span className="text-white font-medium">{gap.required}</span>
                        </div>
                        <div>
                          <span className="text-slate-400">Gap: </span>
                          <span className="text-red-400 font-bold">{gap.gap}</span>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>

          {/* Top Performers */}
          <div className="bg-white/5 rounded-xl border border-white/10 p-5">
            <h2 className="text-lg font-semibold text-white mb-4">Top Performers</h2>
            {briefing.top_performers.length === 0 ? (
              <p className="text-sm text-slate-400">No performer data available.</p>
            ) : (
              <div className="flex gap-4 overflow-x-auto pb-2">
                {briefing.top_performers.map((p) => (
                  <div
                    key={p.employee_id}
                    className="flex-shrink-0 w-48 bg-white/5 rounded-lg border border-white/10 p-4 text-center"
                  >
                    <div className="w-12 h-12 mx-auto rounded-full bg-[#F97316]/20 flex items-center justify-center mb-2">
                      <span className="text-sm font-bold text-[#F97316]">
                        {p.display_name
                          .split(' ')
                          .map((w) => w[0])
                          .join('')
                          .slice(0, 2)
                          .toUpperCase()}
                      </span>
                    </div>
                    <p className="text-sm font-medium text-white truncate">{p.display_name}</p>
                    <p className="text-lg font-bold text-[#F97316] mt-1">{p.points} pts</p>
                    <p className="text-xs text-slate-400 mt-1 truncate">{p.location_name}</p>
                    <p className="text-xs text-slate-500 capitalize">{p.role}</p>
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* Training ROI */}
          <div className="bg-white/5 rounded-xl border border-white/10 overflow-hidden">
            <div className="px-5 py-4 border-b border-white/10">
              <h2 className="text-lg font-semibold text-white">Training ROI</h2>
            </div>
            {briefing.training_roi.length === 0 ? (
              <div className="p-5">
                <p className="text-sm text-slate-400">No training data available.</p>
              </div>
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full text-sm" role="table">
                  <thead>
                    <tr className="border-b border-white/10 text-left">
                      <th className="px-5 py-3 text-xs font-medium text-slate-400 uppercase tracking-wider">Certification</th>
                      <th className="px-5 py-3 text-xs font-medium text-slate-400 uppercase tracking-wider text-right">Certified</th>
                      <th className="px-5 py-3 text-xs font-medium text-slate-400 uppercase tracking-wider text-right">Avg ELU (Certified)</th>
                      <th className="px-5 py-3 text-xs font-medium text-slate-400 uppercase tracking-wider text-right">Avg ELU (Uncertified)</th>
                      <th className="px-5 py-3 text-xs font-medium text-slate-400 uppercase tracking-wider text-right">Lift %</th>
                    </tr>
                  </thead>
                  <tbody>
                    {briefing.training_roi.map((row) => (
                      <tr key={row.certification} className="border-b border-white/5 hover:bg-white/5 transition-colors">
                        <td className="px-5 py-3 text-white font-medium">{row.certification}</td>
                        <td className="px-5 py-3 text-slate-300 text-right">{row.certified_count}</td>
                        <td className="px-5 py-3 text-slate-300 text-right">{row.avg_elu_certified.toFixed(1)}</td>
                        <td className="px-5 py-3 text-slate-300 text-right">{row.avg_elu_uncertified.toFixed(1)}</td>
                        <td className={`px-5 py-3 text-right font-semibold ${row.lift_pct >= 0 ? 'text-emerald-400' : 'text-red-400'}`}>
                          {row.lift_pct >= 0 ? '+' : ''}{row.lift_pct.toFixed(1)}%
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        </>
      )}

      {/* ── Anomaly Investigation Section ─────────────────────────────────────── */}
      {!anomalyLoading && (
        <div className="bg-white/5 rounded-xl border border-white/10">
          <div className="px-5 py-4 border-b border-white/10">
            <h2 className="text-lg font-semibold text-white">Anomaly Investigation</h2>
          </div>

          {/* Filter Tabs */}
          <div className="px-5 pt-4 flex gap-2 flex-wrap">
            {FILTER_TABS.map((tab) => (
              <button
                key={tab.id}
                onClick={() => setAnomalyFilter(tab.id)}
                className={`px-3 py-1.5 rounded-lg text-sm font-medium transition-colors ${
                  anomalyFilter === tab.id
                    ? 'bg-[#F97316] text-white'
                    : 'bg-white/5 text-slate-300 hover:bg-white/10'
                }`}
              >
                {tab.label}
                {tab.id !== 'all' && (
                  <span className="ml-1.5 text-xs opacity-75">
                    ({anomalies.filter((a) => a.status === tab.id).length})
                  </span>
                )}
              </button>
            ))}
          </div>

          {/* Anomaly Feed */}
          <div className="p-5 space-y-3">
            {filteredAnomalies.length === 0 ? (
              <p className="text-sm text-slate-400 py-4 text-center">No anomalies found for this filter.</p>
            ) : (
              filteredAnomalies.map((anomaly) => {
                const isExpanded = expandedAnomaly === anomaly.anomaly_id;
                return (
                  <div
                    key={anomaly.anomaly_id}
                    className="bg-white/5 rounded-lg border border-white/10 overflow-hidden"
                  >
                    {/* Anomaly Header */}
                    <button
                      onClick={() => setExpandedAnomaly(isExpanded ? null : anomaly.anomaly_id)}
                      className="w-full px-4 py-3 flex items-start gap-3 text-left hover:bg-white/5 transition-colors"
                    >
                      <div className="mt-0.5 shrink-0">{anomalyIcon(anomaly.type)}</div>
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2 flex-wrap">
                          {severityBadge(anomaly.severity)}
                          {statusBadge(anomaly.status)}
                        </div>
                        <p className="text-sm font-medium text-white mt-1">{anomaly.title}</p>
                        <p className="text-xs text-slate-400 mt-0.5 truncate">{anomaly.description}</p>
                        <div className="flex items-center gap-3 mt-1.5 text-xs text-slate-500">
                          <span>{anomaly.location_name}</span>
                          <span>&middot;</span>
                          <span>{timeAgo(anomaly.detected_at)}</span>
                        </div>
                      </div>
                      <div className="shrink-0 mt-1">
                        {isExpanded ? (
                          <ChevronUp className="h-4 w-4 text-slate-400" />
                        ) : (
                          <ChevronDown className="h-4 w-4 text-slate-400" />
                        )}
                      </div>
                    </button>

                    {/* Expanded Details */}
                    {isExpanded && (
                      <div className="px-4 pb-4 border-t border-white/5 pt-3 space-y-3">
                        <div>
                          <p className="text-xs font-medium text-slate-400 uppercase tracking-wider mb-1">Full Description</p>
                          <p className="text-sm text-slate-300">{anomaly.description}</p>
                        </div>

                        {/* Evidence JSON */}
                        {anomaly.evidence && Object.keys(anomaly.evidence).length > 0 && (
                          <div>
                            <p className="text-xs font-medium text-slate-400 uppercase tracking-wider mb-1">Evidence</p>
                            <pre className="bg-slate-900 rounded-lg p-3 text-xs text-slate-300 overflow-x-auto max-h-48">
                              {JSON.stringify(anomaly.evidence, null, 2)}
                            </pre>
                          </div>
                        )}

                        {/* Resolution notes */}
                        {anomaly.resolution_notes && (
                          <div>
                            <p className="text-xs font-medium text-slate-400 uppercase tracking-wider mb-1">Resolution Notes</p>
                            <p className="text-sm text-slate-300">{anomaly.resolution_notes}</p>
                          </div>
                        )}

                        {/* Resolve button */}
                        {canResolve && (anomaly.status === 'open' || anomaly.status === 'investigating') && (
                          <button
                            onClick={(e) => {
                              e.stopPropagation();
                              setResolveModal(anomaly);
                              setResolveStatus('resolved');
                              setResolveNotes('');
                            }}
                            className="flex items-center gap-2 bg-[#F97316] hover:bg-[#EA580C] text-white text-sm font-medium px-4 py-2 rounded-lg transition-colors"
                          >
                            Resolve Anomaly
                          </button>
                        )}
                      </div>
                    )}
                  </div>
                );
              })
            )}
          </div>
        </div>
      )}

      {/* ── Resolve Modal ──────────────────────────────────────────────────────── */}
      <Modal
        open={!!resolveModal}
        onClose={() => setResolveModal(null)}
        title="Resolve Anomaly"
        footer={
          <>
            <button
              onClick={() => setResolveModal(null)}
              className="px-4 py-2 text-sm font-medium text-slate-300 hover:text-white transition-colors"
            >
              Cancel
            </button>
            <button
              onClick={handleResolve}
              disabled={resolveMutation.isPending}
              className="flex items-center gap-2 bg-[#F97316] hover:bg-[#EA580C] text-white text-sm font-medium px-4 py-2 rounded-lg transition-colors disabled:opacity-50"
            >
              {resolveMutation.isPending ? 'Saving...' : 'Confirm'}
            </button>
          </>
        }
      >
        <div className="space-y-4">
          {resolveModal && (
            <p className="text-sm text-slate-300">{resolveModal.title}</p>
          )}

          <div>
            <label className="block text-sm font-medium text-slate-300 mb-1" htmlFor="resolve-status">
              Resolution Status
            </label>
            <select
              id="resolve-status"
              value={resolveStatus}
              onChange={(e) => setResolveStatus(e.target.value)}
              className="w-full bg-white/5 border border-white/10 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:ring-1 focus:ring-[#F97316]"
            >
              <option value="resolved" className="bg-slate-800">Confirmed &amp; Resolved</option>
              <option value="false_positive" className="bg-slate-800">False Positive</option>
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-300 mb-1" htmlFor="resolve-notes">
              Notes
            </label>
            <textarea
              id="resolve-notes"
              value={resolveNotes}
              onChange={(e) => setResolveNotes(e.target.value)}
              rows={3}
              placeholder="Add investigation notes..."
              className="w-full bg-white/5 border border-white/10 rounded-lg px-3 py-2 text-sm text-white placeholder-slate-500 focus:outline-none focus:ring-1 focus:ring-[#F97316] resize-none"
            />
          </div>
        </div>
      </Modal>
    </div>
  );
}
