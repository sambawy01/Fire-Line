import { useState } from 'react';
import {
  Wrench, Server, ClipboardList, Calendar, BarChart3,
  AlertTriangle, CheckCircle, Clock, DollarSign, XCircle,
  Settings, Thermometer, Wind, Droplets, Zap, Shield, HelpCircle,
  ChevronRight, Plus, User,
} from 'lucide-react';
import { useLocationStore } from '../stores/location';
import {
  useEquipment,
  useMaintenanceTickets,
  useOverdueEquipment,
  useMaintenanceStats,
  useTicketDetail,
  useCompleteTicket,
  useAddMaintenanceLog,
  useCreateTicket,
} from '../hooks/useMaintenance';
import KPICard from '../components/ui/KPICard';
import LoadingSpinner from '../components/ui/LoadingSpinner';
import Modal from '../components/ui/Modal';
import type { Equipment, MaintenanceTicket } from '../lib/api';

// ── helpers ──────────────────────────────────────────────────────────────────

const TABS = ['Equipment', 'Work Orders', 'Schedule', 'Analytics'] as const;
type Tab = typeof TABS[number];

const CATEGORY_STYLES: Record<string, string> = {
  cooking: 'bg-red-500/20 text-red-400',
  refrigeration: 'bg-blue-500/20 text-blue-400',
  hvac: 'bg-purple-500/20 text-purple-400',
  plumbing: 'bg-teal-500/20 text-teal-400',
  electrical: 'bg-amber-500/20 text-amber-400',
  safety: 'bg-green-500/20 text-green-400',
  other: 'bg-slate-500/20 text-slate-400',
};

const CATEGORY_ICONS: Record<string, React.ElementType> = {
  cooking: Thermometer,
  refrigeration: Thermometer,
  hvac: Wind,
  plumbing: Droplets,
  electrical: Zap,
  safety: Shield,
  other: HelpCircle,
};

const STATUS_STYLES: Record<string, string> = {
  operational: 'bg-green-500/20 text-green-400',
  needs_maintenance: 'bg-amber-500/20 text-amber-400',
  under_repair: 'bg-orange-500/20 text-orange-400',
  out_of_service: 'bg-red-500/20 text-red-400',
  retired: 'bg-slate-500/20 text-slate-400',
};

const PRIORITY_STYLES: Record<string, string> = {
  critical: 'bg-red-500/20 text-red-400 border-red-500/30',
  high: 'bg-orange-500/20 text-orange-400 border-orange-500/30',
  medium: 'bg-blue-500/20 text-blue-400 border-blue-500/30',
  low: 'bg-slate-500/20 text-slate-400 border-slate-500/30',
};

const TICKET_STATUS_STYLES: Record<string, string> = {
  open: 'bg-red-500/20 text-red-400',
  in_progress: 'bg-amber-500/20 text-amber-400',
  on_hold: 'bg-slate-500/20 text-slate-400',
  completed: 'bg-green-500/20 text-green-400',
  cancelled: 'bg-slate-500/20 text-slate-400',
};

function healthColor(score: number): string {
  if (score >= 80) return 'bg-green-500';
  if (score >= 50) return 'bg-amber-500';
  return 'bg-red-500';
}

function healthTextColor(score: number): string {
  if (score >= 80) return 'text-green-400';
  if (score >= 50) return 'text-amber-400';
  return 'text-red-400';
}

function fmtCost(cents: number): string {
  return `EGP ${(cents / 100).toLocaleString('en-US', { minimumFractionDigits: 0, maximumFractionDigits: 0 })}`;
}

function fmtDate(d: string | null): string {
  if (!d) return '--';
  try {
    return new Date(d).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
  } catch {
    return d;
  }
}

function ticketAge(createdAt: string): string {
  const diff = Date.now() - new Date(createdAt).getTime();
  const days = Math.floor(diff / (1000 * 60 * 60 * 24));
  if (days === 0) return 'Today';
  if (days === 1) return '1 day ago';
  return `${days} days ago`;
}

function isOverdue(nextMaintenance: string | null): boolean {
  if (!nextMaintenance) return false;
  return new Date(nextMaintenance) < new Date();
}

function warrantyStatus(expiryDate: string | null): { label: string; style: string } {
  if (!expiryDate) return { label: 'N/A', style: 'text-slate-500' };
  const expiry = new Date(expiryDate);
  const now = new Date();
  const daysLeft = Math.floor((expiry.getTime() - now.getTime()) / (1000 * 60 * 60 * 24));
  if (daysLeft < 0) return { label: 'Expired', style: 'text-red-400' };
  if (daysLeft < 90) return { label: `${daysLeft}d left`, style: 'text-amber-400' };
  return { label: 'Active', style: 'text-green-400' };
}

// ── Main Page ────────────────────────────────────────────────────────────────

export default function MaintenancePage() {
  const [tab, setTab] = useState<Tab>('Equipment');
  const selectedLocationId = useLocationStore((s) => s.selectedLocationId);

  return (
    <div className="space-y-6">
      {/* Page header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="p-2.5 rounded-lg bg-orange-500/20">
            <Wrench className="h-6 w-6 text-orange-400" />
          </div>
          <div>
            <h1 className="text-2xl font-bold text-white">Maintenance</h1>
            <p className="text-sm text-slate-400">Equipment & work order management</p>
          </div>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex gap-1 bg-white/5 rounded-lg p-1 w-fit">
        {TABS.map((t) => {
          const icons = { Equipment: Server, 'Work Orders': ClipboardList, Schedule: Calendar, Analytics: BarChart3 };
          const Icon = icons[t];
          return (
            <button
              key={t}
              onClick={() => setTab(t)}
              className={`flex items-center gap-2 px-4 py-2 rounded-md text-sm font-medium transition-colors ${
                tab === t ? 'bg-orange-500/20 text-orange-400' : 'text-slate-400 hover:text-white hover:bg-white/5'
              }`}
            >
              <Icon className="h-4 w-4" />
              {t}
            </button>
          );
        })}
      </div>

      {/* Tab content */}
      {tab === 'Equipment' && <EquipmentTab locationId={selectedLocationId} />}
      {tab === 'Work Orders' && <WorkOrdersTab locationId={selectedLocationId} />}
      {tab === 'Schedule' && <ScheduleTab locationId={selectedLocationId} />}
      {tab === 'Analytics' && <AnalyticsTab locationId={selectedLocationId} />}
    </div>
  );
}

// ── Tab 1: Equipment Registry ────────────────────────────────────────────────

function EquipmentTab({ locationId }: { locationId: string | null }) {
  const { data, isLoading } = useEquipment(locationId);
  const [selectedEquipment, setSelectedEquipment] = useState<Equipment | null>(null);

  if (isLoading) return <LoadingSpinner size="lg" />;

  const equipment = data?.equipment ?? [];
  const operational = equipment.filter((e) => e.status === 'operational').length;
  const needsMaint = equipment.filter((e) => e.status === 'needs_maintenance').length;
  const outOfService = equipment.filter((e) => e.status === 'out_of_service' || e.status === 'under_repair').length;

  return (
    <div className="space-y-6">
      {/* KPI cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <KPICard label="Total Equipment" value={String(equipment.length)} icon={Server} iconColor="text-blue-400" bgTint="bg-blue-500/20" />
        <KPICard label="Operational" value={String(operational)} icon={CheckCircle} iconColor="text-green-400" bgTint="bg-green-500/20" />
        <KPICard label="Needs Maintenance" value={String(needsMaint)} icon={AlertTriangle} iconColor="text-amber-400" bgTint="bg-amber-500/20" />
        <KPICard label="Out of Service" value={String(outOfService)} icon={XCircle} iconColor="text-red-400" bgTint="bg-red-500/20" />
      </div>

      {/* Equipment table */}
      <div className="bg-white/5 rounded-xl border border-white/10 overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="bg-white/5 text-slate-400 uppercase tracking-wider text-xs">
                <th className="px-6 py-3 text-left font-medium">Equipment</th>
                <th className="px-6 py-3 text-left font-medium">Category</th>
                <th className="px-6 py-3 text-left font-medium">Status</th>
                <th className="px-6 py-3 text-left font-medium">Health</th>
                <th className="px-6 py-3 text-left font-medium">Last Maint.</th>
                <th className="px-6 py-3 text-left font-medium">Next Maint.</th>
                <th className="px-6 py-3 text-left font-medium">Warranty</th>
                <th className="px-6 py-3 text-center font-medium"></th>
              </tr>
            </thead>
            <tbody className="divide-y divide-white/5">
              {equipment.map((eq) => {
                const CatIcon = CATEGORY_ICONS[eq.category] ?? Settings;
                const warranty = warrantyStatus(eq.warranty_expiry);
                const overdue = isOverdue(eq.next_maintenance);
                return (
                  <tr
                    key={eq.equipment_id}
                    className="hover:bg-white/5 cursor-pointer transition-colors"
                    onClick={() => setSelectedEquipment(eq)}
                  >
                    <td className="px-6 py-3">
                      <div>
                        <p className="text-white font-medium">{eq.name}</p>
                        <p className="text-xs text-slate-500">{eq.make ?? ''} {eq.model ?? ''}</p>
                      </div>
                    </td>
                    <td className="px-6 py-3">
                      <span className={`inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium ${CATEGORY_STYLES[eq.category]}`}>
                        <CatIcon className="h-3 w-3" />
                        {eq.category}
                      </span>
                    </td>
                    <td className="px-6 py-3">
                      <span className={`inline-flex px-2.5 py-0.5 rounded-full text-xs font-medium ${STATUS_STYLES[eq.status]}`}>
                        {eq.status.replace(/_/g, ' ')}
                      </span>
                    </td>
                    <td className="px-6 py-3">
                      <div className="flex items-center gap-2">
                        <div className="w-16 h-2 bg-white/10 rounded-full overflow-hidden">
                          <div className={`h-full rounded-full ${healthColor(eq.health_score)}`} style={{ width: `${eq.health_score}%` }} />
                        </div>
                        <span className={`text-xs font-bold ${healthTextColor(eq.health_score)}`}>{eq.health_score}</span>
                      </div>
                    </td>
                    <td className="px-6 py-3 text-slate-300 text-xs">{fmtDate(eq.last_maintenance)}</td>
                    <td className="px-6 py-3">
                      <span className={`text-xs ${overdue ? 'text-red-400 font-bold' : 'text-slate-300'}`}>
                        {fmtDate(eq.next_maintenance)}
                        {overdue && <span className="ml-1.5 px-1.5 py-0.5 rounded bg-red-500/20 text-red-400 text-[10px] font-bold uppercase">Overdue</span>}
                      </span>
                    </td>
                    <td className="px-6 py-3">
                      <span className={`text-xs font-medium ${warranty.style}`}>{warranty.label}</span>
                    </td>
                    <td className="px-6 py-3 text-center">
                      <ChevronRight className="h-4 w-4 text-slate-500" />
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      </div>

      {/* Equipment detail modal */}
      {selectedEquipment && (
        <EquipmentDetailModal equipment={selectedEquipment} onClose={() => setSelectedEquipment(null)} />
      )}
    </div>
  );
}

function EquipmentDetailModal({ equipment: eq, onClose }: { equipment: Equipment; onClose: () => void }) {
  const warranty = warrantyStatus(eq.warranty_expiry);
  const CatIcon = CATEGORY_ICONS[eq.category] ?? Settings;

  return (
    <Modal open onClose={onClose} title={eq.name}>
      <div className="space-y-4">
        <div className="grid grid-cols-2 gap-4">
          <div>
            <p className="text-xs text-slate-400 mb-1">Category</p>
            <span className={`inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium ${CATEGORY_STYLES[eq.category]}`}>
              <CatIcon className="h-3 w-3" />
              {eq.category}
            </span>
          </div>
          <div>
            <p className="text-xs text-slate-400 mb-1">Status</p>
            <span className={`inline-flex px-2.5 py-0.5 rounded-full text-xs font-medium ${STATUS_STYLES[eq.status]}`}>
              {eq.status.replace(/_/g, ' ')}
            </span>
          </div>
          <div>
            <p className="text-xs text-slate-400 mb-1">Health Score</p>
            <div className="flex items-center gap-2">
              <div className="w-20 h-2.5 bg-white/10 rounded-full overflow-hidden">
                <div className={`h-full rounded-full ${healthColor(eq.health_score)}`} style={{ width: `${eq.health_score}%` }} />
              </div>
              <span className={`text-sm font-bold ${healthTextColor(eq.health_score)}`}>{eq.health_score}%</span>
            </div>
          </div>
          <div>
            <p className="text-xs text-slate-400 mb-1">Warranty</p>
            <span className={`text-sm font-medium ${warranty.style}`}>{warranty.label}</span>
          </div>
        </div>

        <div className="border-t border-white/10 pt-4 grid grid-cols-2 gap-3 text-sm">
          <div><span className="text-slate-400">Make:</span> <span className="text-white ml-1">{eq.make ?? 'N/A'}</span></div>
          <div><span className="text-slate-400">Model:</span> <span className="text-white ml-1">{eq.model ?? 'N/A'}</span></div>
          <div><span className="text-slate-400">Serial:</span> <span className="text-white ml-1">{eq.serial_number ?? 'N/A'}</span></div>
          <div><span className="text-slate-400">Install Date:</span> <span className="text-white ml-1">{fmtDate(eq.install_date)}</span></div>
          <div><span className="text-slate-400">Last Maintenance:</span> <span className="text-white ml-1">{fmtDate(eq.last_maintenance)}</span></div>
          <div><span className="text-slate-400">Next Maintenance:</span> <span className="text-white ml-1">{fmtDate(eq.next_maintenance)}</span></div>
          <div><span className="text-slate-400">Interval:</span> <span className="text-white ml-1">{eq.maintenance_interval_days} days</span></div>
          <div><span className="text-slate-400">Warranty Expiry:</span> <span className="text-white ml-1">{fmtDate(eq.warranty_expiry)}</span></div>
        </div>

        {eq.notes && (
          <div className="border-t border-white/10 pt-4">
            <p className="text-xs text-slate-400 mb-1">Notes</p>
            <p className="text-sm text-slate-300">{eq.notes}</p>
          </div>
        )}
      </div>
    </Modal>
  );
}

// ── Tab 2: Work Orders ───────────────────────────────────────────────────────

function WorkOrdersTab({ locationId }: { locationId: string | null }) {
  const [statusFilter, setStatusFilter] = useState('');
  const [priorityFilter, setPriorityFilter] = useState('');
  const [selectedTicketId, setSelectedTicketId] = useState<string | null>(null);

  const { data, isLoading } = useMaintenanceTickets(locationId, statusFilter || undefined, priorityFilter || undefined);

  if (isLoading) return <LoadingSpinner size="lg" />;

  const tickets = data?.tickets ?? [];

  return (
    <div className="space-y-6">
      {/* Filter bar */}
      <div className="flex items-center gap-4">
        <select
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value)}
          className="bg-white/5 border border-white/10 text-white text-sm rounded-lg px-3 py-2 focus:outline-none focus:ring-1 focus:ring-orange-500"
        >
          <option value="">All Statuses</option>
          <option value="open">Open</option>
          <option value="in_progress">In Progress</option>
          <option value="on_hold">On Hold</option>
          <option value="completed">Completed</option>
          <option value="cancelled">Cancelled</option>
        </select>
        <select
          value={priorityFilter}
          onChange={(e) => setPriorityFilter(e.target.value)}
          className="bg-white/5 border border-white/10 text-white text-sm rounded-lg px-3 py-2 focus:outline-none focus:ring-1 focus:ring-orange-500"
        >
          <option value="">All Priorities</option>
          <option value="critical">Critical</option>
          <option value="high">High</option>
          <option value="medium">Medium</option>
          <option value="low">Low</option>
        </select>
        <span className="text-sm text-slate-400">{tickets.length} ticket{tickets.length !== 1 ? 's' : ''}</span>
      </div>

      {/* Ticket cards */}
      {tickets.length === 0 ? (
        <div className="text-center py-12 text-slate-400">No tickets match the selected filters.</div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
          {tickets.map((t) => (
            <div
              key={t.ticket_id}
              onClick={() => setSelectedTicketId(t.ticket_id)}
              className={`bg-white/5 rounded-xl border p-5 cursor-pointer hover:bg-white/[0.08] transition-colors ${
                PRIORITY_STYLES[t.priority]?.split(' ').pop() ?? 'border-white/10'
              }`}
            >
              <div className="flex items-start justify-between mb-3">
                <div>
                  <p className="text-xs text-slate-500 font-mono">{t.ticket_number}</p>
                  <p className="text-white font-semibold mt-0.5">{t.title}</p>
                </div>
                <span className={`inline-flex px-2 py-0.5 rounded-full text-[10px] font-bold uppercase ${PRIORITY_STYLES[t.priority]}`}>
                  {t.priority}
                </span>
              </div>

              <div className="flex items-center gap-2 mb-3">
                <span className={`inline-flex px-2 py-0.5 rounded-full text-xs font-medium ${TICKET_STATUS_STYLES[t.status]}`}>
                  {t.status.replace(/_/g, ' ')}
                </span>
                <span className="text-xs text-slate-500 px-2 py-0.5 rounded-full bg-white/5 capitalize">{t.type}</span>
              </div>

              <div className="space-y-1.5 text-xs text-slate-400">
                <div className="flex items-center gap-1.5">
                  <Server className="h-3 w-3" />
                  {t.equipment_name}
                </div>
                {t.assigned_to && (
                  <div className="flex items-center gap-1.5">
                    <User className="h-3 w-3" />
                    {t.assigned_to}
                  </div>
                )}
                <div className="flex items-center gap-1.5">
                  <Clock className="h-3 w-3" />
                  {ticketAge(t.created_at)}
                </div>
                {t.estimated_cost > 0 && (
                  <div className="flex items-center gap-1.5">
                    <DollarSign className="h-3 w-3" />
                    Est. {fmtCost(t.estimated_cost)}
                    {t.actual_cost > 0 && <span className="ml-1">/ Actual {fmtCost(t.actual_cost)}</span>}
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Ticket detail modal */}
      {selectedTicketId && (
        <TicketDetailModal ticketId={selectedTicketId} onClose={() => setSelectedTicketId(null)} />
      )}
    </div>
  );
}

function TicketDetailModal({ ticketId, onClose }: { ticketId: string; onClose: () => void }) {
  const { data: ticket, isLoading } = useTicketDetail(ticketId);
  const completeMut = useCompleteTicket();
  const addLogMut = useAddMaintenanceLog();
  const [resolution, setResolution] = useState('');
  const [actualCost, setActualCost] = useState('');
  const [logAction, setLogAction] = useState('');
  const [logNotes, setLogNotes] = useState('');

  if (isLoading || !ticket) {
    return <Modal open onClose={onClose} title="Loading..."><LoadingSpinner /></Modal>;
  }

  const t = ticket;
  const logs = t.logs ?? [];

  const handleComplete = () => {
    completeMut.mutate({ id: ticketId, resolution, actualCost: parseInt(actualCost || '0') * 100 }, {
      onSuccess: onClose,
    });
  };

  const handleAddLog = () => {
    if (!logAction.trim()) return;
    addLogMut.mutate({ ticketId, data: { action: logAction, notes: logNotes || undefined } }, {
      onSuccess: () => { setLogAction(''); setLogNotes(''); },
    });
  };

  return (
    <Modal
      open
      onClose={onClose}
      title={`${t.ticket_number} - ${t.title}`}
      footer={
        t.status !== 'completed' && t.status !== 'cancelled' ? (
          <>
            <button
              onClick={handleComplete}
              disabled={!resolution.trim()}
              className="px-4 py-2 rounded-lg bg-green-600 text-white text-sm font-medium hover:bg-green-500 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              Complete Ticket
            </button>
          </>
        ) : undefined
      }
    >
      <div className="space-y-4 max-h-[60vh] overflow-y-auto">
        {/* Ticket info */}
        <div className="grid grid-cols-2 gap-3 text-sm">
          <div>
            <span className="text-slate-400">Type:</span>
            <span className="text-white ml-1 capitalize">{t.type}</span>
          </div>
          <div>
            <span className="text-slate-400">Priority:</span>
            <span className={`ml-1 px-2 py-0.5 rounded-full text-xs font-medium ${PRIORITY_STYLES[t.priority]}`}>{t.priority}</span>
          </div>
          <div>
            <span className="text-slate-400">Status:</span>
            <span className={`ml-1 px-2 py-0.5 rounded-full text-xs font-medium ${TICKET_STATUS_STYLES[t.status]}`}>{t.status.replace(/_/g, ' ')}</span>
          </div>
          <div>
            <span className="text-slate-400">Equipment:</span>
            <span className="text-white ml-1">{t.equipment_name}</span>
          </div>
          {t.assigned_to && <div><span className="text-slate-400">Assigned:</span> <span className="text-white ml-1">{t.assigned_to}</span></div>}
          {t.scheduled_date && <div><span className="text-slate-400">Scheduled:</span> <span className="text-white ml-1">{fmtDate(t.scheduled_date)}</span></div>}
          <div><span className="text-slate-400">Est. Cost:</span> <span className="text-white ml-1">{fmtCost(t.estimated_cost)}</span></div>
          {t.actual_cost > 0 && <div><span className="text-slate-400">Actual Cost:</span> <span className="text-white ml-1">{fmtCost(t.actual_cost)}</span></div>}
        </div>

        {t.description && (
          <div className="border-t border-white/10 pt-3">
            <p className="text-xs text-slate-400 mb-1">Description</p>
            <p className="text-sm text-slate-300">{t.description}</p>
          </div>
        )}

        {t.resolution && (
          <div className="border-t border-white/10 pt-3">
            <p className="text-xs text-slate-400 mb-1">Resolution</p>
            <p className="text-sm text-green-400">{t.resolution}</p>
          </div>
        )}

        {/* Logs */}
        <div className="border-t border-white/10 pt-3">
          <p className="text-xs text-slate-400 mb-2 uppercase tracking-wider">Activity Log</p>
          {logs.length === 0 ? (
            <p className="text-sm text-slate-500">No activity recorded yet.</p>
          ) : (
            <div className="space-y-2">
              {logs.map((log) => (
                <div key={log.log_id} className="flex gap-3 text-sm">
                  <div className="w-1.5 h-1.5 rounded-full bg-orange-500 mt-1.5 shrink-0" />
                  <div className="flex-1">
                    <p className="text-white font-medium">{log.action}</p>
                    {log.notes && <p className="text-slate-400 text-xs">{log.notes}</p>}
                    <div className="flex items-center gap-3 mt-0.5 text-xs text-slate-500">
                      {log.performed_by && <span>{log.performed_by}</span>}
                      <span>{fmtDate(log.performed_at)}</span>
                      {log.cost > 0 && <span>{fmtCost(log.cost)}</span>}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Add log form (only if not completed/cancelled) */}
        {t.status !== 'completed' && t.status !== 'cancelled' && (
          <div className="border-t border-white/10 pt-3 space-y-2">
            <p className="text-xs text-slate-400 uppercase tracking-wider">Add Log Entry</p>
            <input
              type="text"
              placeholder="Action..."
              value={logAction}
              onChange={(e) => setLogAction(e.target.value)}
              className="w-full bg-white/5 border border-white/10 text-white text-sm rounded-lg px-3 py-2 focus:outline-none focus:ring-1 focus:ring-orange-500"
            />
            <input
              type="text"
              placeholder="Notes (optional)"
              value={logNotes}
              onChange={(e) => setLogNotes(e.target.value)}
              className="w-full bg-white/5 border border-white/10 text-white text-sm rounded-lg px-3 py-2 focus:outline-none focus:ring-1 focus:ring-orange-500"
            />
            <button
              onClick={handleAddLog}
              disabled={!logAction.trim()}
              className="px-3 py-1.5 rounded-lg bg-white/10 text-white text-sm hover:bg-white/20 disabled:opacity-50 transition-colors"
            >
              Add Entry
            </button>
          </div>
        )}

        {/* Complete form */}
        {t.status !== 'completed' && t.status !== 'cancelled' && (
          <div className="border-t border-white/10 pt-3 space-y-2">
            <p className="text-xs text-slate-400 uppercase tracking-wider">Complete Ticket</p>
            <textarea
              placeholder="Resolution description..."
              value={resolution}
              onChange={(e) => setResolution(e.target.value)}
              rows={2}
              className="w-full bg-white/5 border border-white/10 text-white text-sm rounded-lg px-3 py-2 focus:outline-none focus:ring-1 focus:ring-orange-500"
            />
            <input
              type="number"
              placeholder="Actual cost ($)"
              value={actualCost}
              onChange={(e) => setActualCost(e.target.value)}
              className="w-full bg-white/5 border border-white/10 text-white text-sm rounded-lg px-3 py-2 focus:outline-none focus:ring-1 focus:ring-orange-500"
            />
          </div>
        )}
      </div>
    </Modal>
  );
}

// ── Tab 3: Maintenance Schedule ──────────────────────────────────────────────

function ScheduleTab({ locationId }: { locationId: string | null }) {
  const { data: overdueData, isLoading: overdueLoading } = useOverdueEquipment(locationId);
  const { data: equipmentData, isLoading: eqLoading } = useEquipment(locationId);
  const createTicketMut = useCreateTicket();

  if (overdueLoading || eqLoading) return <LoadingSpinner size="lg" />;

  const overdue = overdueData?.equipment ?? [];
  const allEquipment = equipmentData?.equipment ?? [];

  // Build schedule: equipment sorted by next_maintenance date
  const scheduled = [...allEquipment]
    .filter((e) => e.next_maintenance)
    .sort((a, b) => {
      const da = new Date(a.next_maintenance!).getTime();
      const db = new Date(b.next_maintenance!).getTime();
      return da - db;
    });

  const overdueIds = new Set(overdue.map((e) => e.equipment_id));

  const handleCreateTicket = (eq: Equipment) => {
    createTicketMut.mutate({
      location_id: eq.location_id,
      equipment_id: eq.equipment_id,
      type: 'preventive',
      priority: 'high',
      title: `Scheduled maintenance: ${eq.name}`,
      description: `Overdue preventive maintenance for ${eq.name}. Last maintained: ${eq.last_maintenance ?? 'Never'}.`,
      estimated_cost: 0,
    });
  };

  return (
    <div className="space-y-6">
      {/* Overdue alert */}
      {overdue.length > 0 && (
        <div className="bg-red-500/10 border border-red-500/20 rounded-xl p-4">
          <div className="flex items-center gap-2 mb-2">
            <AlertTriangle className="h-5 w-5 text-red-400" />
            <h3 className="text-red-400 font-semibold">{overdue.length} Overdue Equipment</h3>
          </div>
          <p className="text-sm text-slate-400 mb-3">The following equipment has passed its scheduled maintenance date.</p>
        </div>
      )}

      {/* Timeline */}
      <div className="bg-white/5 rounded-xl border border-white/10 overflow-hidden">
        <div className="px-6 py-4 border-b border-white/10">
          <h3 className="text-white font-semibold">Maintenance Schedule</h3>
          <p className="text-xs text-slate-400 mt-0.5">Equipment sorted by next maintenance date</p>
        </div>
        <div className="divide-y divide-white/5">
          {scheduled.map((eq) => {
            const isOd = overdueIds.has(eq.equipment_id);
            const CatIcon = CATEGORY_ICONS[eq.category] ?? Settings;
            return (
              <div key={eq.equipment_id} className={`px-6 py-4 flex items-center justify-between ${isOd ? 'bg-red-500/5' : ''}`}>
                <div className="flex items-center gap-4">
                  <div className={`p-2 rounded-lg ${isOd ? 'bg-red-500/20' : 'bg-white/5'}`}>
                    <CatIcon className={`h-5 w-5 ${isOd ? 'text-red-400' : 'text-slate-400'}`} />
                  </div>
                  <div>
                    <div className="flex items-center gap-2">
                      <p className="text-white font-medium">{eq.name}</p>
                      {isOd && (
                        <span className="px-1.5 py-0.5 rounded bg-red-500/20 text-red-400 text-[10px] font-bold uppercase">Overdue</span>
                      )}
                    </div>
                    <p className="text-xs text-slate-500">{eq.make ?? ''} {eq.model ?? ''}</p>
                  </div>
                </div>
                <div className="flex items-center gap-4">
                  <div className="text-right">
                    <p className={`text-sm font-medium ${isOd ? 'text-red-400' : 'text-slate-300'}`}>
                      {fmtDate(eq.next_maintenance)}
                    </p>
                    <p className="text-xs text-slate-500">Last: {fmtDate(eq.last_maintenance)}</p>
                  </div>
                  <div className="flex items-center gap-2">
                    <div className="w-12 h-2 bg-white/10 rounded-full overflow-hidden">
                      <div className={`h-full rounded-full ${healthColor(eq.health_score)}`} style={{ width: `${eq.health_score}%` }} />
                    </div>
                    <span className={`text-xs font-bold ${healthTextColor(eq.health_score)}`}>{eq.health_score}</span>
                  </div>
                  {isOd && (
                    <button
                      onClick={() => handleCreateTicket(eq)}
                      disabled={createTicketMut.isPending}
                      className="flex items-center gap-1 px-3 py-1.5 rounded-lg bg-orange-500/20 text-orange-400 text-xs font-medium hover:bg-orange-500/30 transition-colors"
                    >
                      <Plus className="h-3 w-3" />
                      Create Ticket
                    </button>
                  )}
                </div>
              </div>
            );
          })}
          {scheduled.length === 0 && (
            <div className="px-6 py-12 text-center text-slate-400">No equipment with scheduled maintenance.</div>
          )}
        </div>
      </div>
    </div>
  );
}

// ── Tab 4: Analytics ─────────────────────────────────────────────────────────

/** Derive monthly maintenance costs from completed tickets over the last 6 months */
function computeMonthlyCosts(tickets: MaintenanceTicket[]): { month: string; cost: number }[] {
  const now = new Date();
  const monthNames = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
  const months: { month: string; year: number; monthIdx: number; cost: number }[] = [];

  for (let i = 5; i >= 0; i--) {
    const d = new Date(now.getFullYear(), now.getMonth() - i, 1);
    months.push({ month: monthNames[d.getMonth()], year: d.getFullYear(), monthIdx: d.getMonth(), cost: 0 });
  }

  // Sum actual_cost from completed tickets into the appropriate month bucket
  tickets.forEach((t) => {
    if (t.status !== 'completed' || !t.completed_at) return;
    const completedDate = new Date(t.completed_at);
    const bucket = months.find(
      (m) => m.monthIdx === completedDate.getMonth() && m.year === completedDate.getFullYear()
    );
    if (bucket) {
      bucket.cost += t.actual_cost;
    }
  });

  return months.map(({ month, cost }) => ({ month, cost }));
}

function AnalyticsTab({ locationId }: { locationId: string | null }) {
  const { data: stats, isLoading } = useMaintenanceStats(locationId);
  const { data: ticketsData, isLoading: ticketsLoading } = useMaintenanceTickets(locationId);

  if (isLoading || ticketsLoading) return <LoadingSpinner size="lg" />;
  if (!stats) return <div className="text-slate-400 text-center py-12">No analytics data available.</div>;

  const uptimePct = stats.total_equipment > 0
    ? Math.round((stats.operational_count / stats.total_equipment) * 100)
    : 0;

  // Compute monthly costs from real completed tickets
  const allTickets = ticketsData?.tickets ?? [];
  const monthlyCosts = computeMonthlyCosts(allTickets);
  const maxCost = Math.max(...monthlyCosts.map((m) => m.cost), 1);

  // Ticket type colors
  const typeColors: Record<string, string> = {
    preventive: 'bg-blue-500',
    corrective: 'bg-amber-500',
    emergency: 'bg-red-500',
    inspection: 'bg-green-500',
  };

  const totalTickets = (stats.tickets_by_type ?? []).reduce((sum, t) => sum + t.count, 0) || 1;

  // Health distribution colors
  const healthColors: Record<string, string> = {
    'Good (80-100)': 'bg-green-500',
    'Fair (50-79)': 'bg-amber-500',
    'Poor (0-49)': 'bg-red-500',
  };

  return (
    <div className="space-y-6">
      {/* KPI cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <KPICard
          label="Open Tickets"
          value={String(stats.open_tickets + stats.in_progress_tickets)}
          icon={ClipboardList}
          iconColor="text-orange-400"
          bgTint="bg-orange-500/20"
        />
        <KPICard
          label="Avg Resolution"
          value={`${stats.avg_resolution_hours.toFixed(1)}h`}
          icon={Clock}
          iconColor="text-blue-400"
          bgTint="bg-blue-500/20"
        />
        <KPICard
          label="Cost This Month"
          value={fmtCost(stats.total_cost_this_month)}
          icon={DollarSign}
          iconColor="text-green-400"
          bgTint="bg-green-500/20"
        />
        <KPICard
          label="Equipment Uptime"
          value={`${uptimePct}%`}
          icon={CheckCircle}
          iconColor="text-emerald-400"
          bgTint="bg-emerald-500/20"
        />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Tickets by type pie chart (bar-based approximation) */}
        <div className="bg-white/5 rounded-xl border border-white/10 p-5">
          <h3 className="text-white font-semibold mb-4">Tickets by Type</h3>
          <div className="space-y-3">
            {(stats.tickets_by_type ?? []).map((t) => {
              const pct = Math.round((t.count / totalTickets) * 100);
              return (
                <div key={t.type}>
                  <div className="flex justify-between text-sm mb-1">
                    <span className="text-slate-300 capitalize">{t.type}</span>
                    <span className="text-slate-400">{t.count} ({pct}%)</span>
                  </div>
                  <div className="h-3 bg-white/10 rounded-full overflow-hidden">
                    <div className={`h-full rounded-full ${typeColors[t.type] ?? 'bg-slate-500'}`} style={{ width: `${pct}%` }} />
                  </div>
                </div>
              );
            })}
          </div>
        </div>

        {/* Maintenance cost by month */}
        <div className="bg-white/5 rounded-xl border border-white/10 p-5">
          <h3 className="text-white font-semibold mb-4">Monthly Maintenance Cost</h3>
          <div className="flex items-end gap-2 h-48">
            {monthlyCosts.map((m) => {
              const h = Math.max((m.cost / maxCost) * 100, 4);
              return (
                <div key={m.month} className="flex-1 flex flex-col items-center justify-end">
                  <span className="text-xs text-slate-400 mb-1">{fmtCost(m.cost)}</span>
                  <div
                    className="w-full bg-orange-500/80 rounded-t-md transition-all duration-500"
                    style={{ height: `${h}%` }}
                  />
                  <span className="text-xs text-slate-400 mt-2">{m.month}</span>
                </div>
              );
            })}
          </div>
        </div>

        {/* Equipment health distribution */}
        <div className="bg-white/5 rounded-xl border border-white/10 p-5">
          <h3 className="text-white font-semibold mb-4">Equipment Health Distribution</h3>
          <div className="space-y-3">
            {(stats.health_distribution ?? []).map((h) => {
              const pct = stats.total_equipment > 0 ? Math.round((h.count / stats.total_equipment) * 100) : 0;
              return (
                <div key={h.range}>
                  <div className="flex justify-between text-sm mb-1">
                    <span className="text-slate-300">{h.range}</span>
                    <span className="text-slate-400">{h.count} ({pct}%)</span>
                  </div>
                  <div className="h-3 bg-white/10 rounded-full overflow-hidden">
                    <div className={`h-full rounded-full ${healthColors[h.range] ?? 'bg-slate-500'}`} style={{ width: `${pct}%` }} />
                  </div>
                </div>
              );
            })}
          </div>
        </div>

        {/* Summary stats */}
        <div className="bg-white/5 rounded-xl border border-white/10 p-5">
          <h3 className="text-white font-semibold mb-4">Overview</h3>
          <div className="grid grid-cols-2 gap-4">
            <div className="bg-white/5 rounded-lg p-3 text-center">
              <p className="text-2xl font-bold text-white">{stats.total_equipment}</p>
              <p className="text-xs text-slate-400">Total Equipment</p>
            </div>
            <div className="bg-white/5 rounded-lg p-3 text-center">
              <p className="text-2xl font-bold text-green-400">{stats.operational_count}</p>
              <p className="text-xs text-slate-400">Operational</p>
            </div>
            <div className="bg-white/5 rounded-lg p-3 text-center">
              <p className="text-2xl font-bold text-amber-400">{stats.needs_maintenance_count}</p>
              <p className="text-xs text-slate-400">Needs Maintenance</p>
            </div>
            <div className="bg-white/5 rounded-lg p-3 text-center">
              <p className="text-2xl font-bold text-red-400">{stats.overdue_count}</p>
              <p className="text-xs text-slate-400">Overdue</p>
            </div>
            <div className="bg-white/5 rounded-lg p-3 text-center col-span-2">
              <p className="text-2xl font-bold text-blue-400">{stats.avg_health_score.toFixed(0)}%</p>
              <p className="text-xs text-slate-400">Average Health Score</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
