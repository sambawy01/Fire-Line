import { useState, useEffect } from 'react';
import { ChefHat, Clock, Zap, CheckCircle, AlertTriangle } from 'lucide-react';
import { useLocationStore } from '../stores/location';
import { useCapacity, useKDSTickets, useKDSMetrics, useBumpItem } from '../hooks/useKitchen';
import KPICard from '../components/ui/KPICard';
import LoadingSpinner from '../components/ui/LoadingSpinner';
import ErrorBanner from '../components/ui/ErrorBanner';
import type { KDSTicket, KitchenStation } from '../lib/api';

// ── helpers ──────────────────────────────────────────────────────────────────

function fmtSecs(secs: number): string {
  const m = Math.floor(secs / 60);
  const s = secs % 60;
  return `${m}:${String(s).padStart(2, '0')}`;
}

function loadColor(pct: number): string {
  if (pct >= 80) return 'bg-red-500';
  if (pct >= 50) return 'bg-yellow-400';
  return 'bg-green-500';
}

function urgencyClasses(elapsedSecs: number): string {
  const mins = elapsedSecs / 60;
  if (mins >= 10) return 'border-red-500 bg-red-500/10';
  if (mins >= 5) return 'border-yellow-400 bg-yellow-400/10';
  return 'border-green-400 bg-green-400/10';
}

function urgencyTextColor(elapsedSecs: number): string {
  const mins = elapsedSecs / 60;
  if (mins >= 10) return 'text-red-600';
  if (mins >= 5) return 'text-yellow-600';
  return 'text-green-600';
}

const CHANNEL_LABELS: Record<string, string> = {
  dine_in: 'Dine-in',
  takeout: 'Takeout',
  delivery: 'Delivery',
  drive_thru: 'Drive-Thru',
  catering: 'Catering',
  online: 'Online',
};

const STATION_COLORS: Record<string, string> = {
  grill: 'bg-orange-500/20 text-orange-400',
  fry: 'bg-yellow-500/20 text-yellow-400',
  salad: 'bg-green-500/20 text-green-400',
  dessert: 'bg-pink-500/20 text-pink-400',
  beverage: 'bg-blue-500/20 text-blue-400',
  expo: 'bg-purple-500/20 text-purple-400',
};

const ITEM_STATUS_COLORS: Record<string, string> = {
  pending: 'bg-white/10 text-slate-300',
  cooking: 'bg-orange-500/20 text-orange-400',
  ready: 'bg-green-500/20 text-green-400',
};

// ── sub-components ────────────────────────────────────────────────────────────

function StationCard({ station }: { station: KitchenStation }) {
  const barColor = loadColor(station.load_pct ?? 0);
  return (
    <div className="bg-white/5 rounded-xl border border-white/10 p-4 shadow-sm">
      <div className="flex items-center justify-between mb-3">
        <div>
          <p className="font-semibold text-white">{station.name}</p>
          <p className="text-xs text-slate-300 capitalize">{station.station_type}</p>
        </div>
        <span
          className={`text-xs font-bold px-2 py-0.5 rounded-full ${
            station.status === 'active'
              ? 'bg-green-500/20 text-green-400'
              : 'bg-white/10 text-slate-400'
          }`}
        >
          {station.status}
        </span>
      </div>

      {/* load bar */}
      <div className="mb-2">
        <div className="flex justify-between text-xs text-slate-400 mb-1">
          <span>{station.current_load}/{station.max_concurrent} slots</span>
          <span className={(station.load_pct ?? 0) >= 80 ? 'text-red-600 font-semibold' : ''}>
            {(station.load_pct ?? 0).toFixed(0)}%
          </span>
        </div>
        <div className="h-2.5 bg-white/10 rounded-full overflow-hidden">
          <div
            className={`h-full rounded-full transition-all duration-500 ${barColor}`}
            style={{ width: `${Math.min(station.load_pct ?? 0, 100)}%` }}
          />
        </div>
      </div>
    </div>
  );
}

function ElapsedTimer({ baseSecs }: { baseSecs: number }) {
  const [elapsed, setElapsed] = useState(baseSecs);

  useEffect(() => {
    setElapsed(baseSecs);
    const id = setInterval(() => setElapsed((e) => e + 1), 1000);
    return () => clearInterval(id);
  }, [baseSecs]);

  return (
    <span className={`font-mono font-bold text-sm ${urgencyTextColor(elapsed)}`}>
      {fmtSecs(elapsed)}
    </span>
  );
}

function TicketCard({
  ticket,
  onBump,
  bumping,
}: {
  ticket: KDSTicket;
  onBump: (itemId: string, status: string) => void;
  bumping: boolean;
}) {
  return (
    <div
      className={`rounded-xl border-2 p-4 shadow-sm transition-colors ${urgencyClasses(ticket.elapsed_secs)}`}
    >
      <div className="flex items-center justify-between mb-3">
        <div className="flex items-center gap-2">
          <span className="font-bold text-white text-lg">#{ticket.order_number}</span>
          <span className="text-xs px-2 py-0.5 rounded-full bg-white border border-white/10 text-slate-300">
            {CHANNEL_LABELS[ticket.channel] ?? ticket.channel}
          </span>
        </div>
        <div className="flex items-center gap-1">
          <Clock className="h-3.5 w-3.5 text-slate-300" />
          <ElapsedTimer baseSecs={ticket.elapsed_secs} />
        </div>
      </div>

      <div className="space-y-2">
        {(ticket.items ?? []).map((item) => {
          const stationClass =
            STATION_COLORS[item.station_type] ?? 'bg-gray-100 text-slate-300';
          const statusClass =
            ITEM_STATUS_COLORS[item.status] ?? 'bg-gray-100 text-slate-300';

          return (
            <div
              key={item.ticket_item_id}
              className="flex items-center justify-between bg-white/10 rounded-lg px-3 py-2"
            >
              <div className="flex items-center gap-2 flex-1 min-w-0">
                <span className="font-semibold text-slate-200 text-sm">
                  {item.quantity}×
                </span>
                <span className="text-sm text-slate-200 truncate">{item.item_name}</span>
                <span className={`text-xs px-1.5 py-0.5 rounded-full capitalize ${stationClass}`}>
                  {item.station_type}
                </span>
              </div>
              <div className="flex items-center gap-1.5 shrink-0 ml-2">
                <span className={`text-xs px-1.5 py-0.5 rounded-full capitalize ${statusClass}`}>
                  {item.status}
                </span>
                {item.status === 'pending' && (
                  <button
                    disabled={bumping}
                    onClick={() => onBump(item.ticket_item_id, 'cooking')}
                    className="text-xs px-2 py-0.5 rounded-md bg-orange-500 text-white hover:bg-orange-600 disabled:opacity-50 transition-colors"
                  >
                    Start
                  </button>
                )}
                {item.status === 'cooking' && (
                  <button
                    disabled={bumping}
                    onClick={() => onBump(item.ticket_item_id, 'ready')}
                    className="text-xs px-2 py-0.5 rounded-md bg-green-500 text-white hover:bg-green-600 disabled:opacity-50 transition-colors"
                  >
                    Done
                  </button>
                )}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}

// ── main page ─────────────────────────────────────────────────────────────────

export default function KitchenPage() {
  const { selectedLocationId } = useLocationStore();

  const { data: capacityData, isLoading: capLoading, error: capErr } = useCapacity(selectedLocationId);
  const { data: ticketsData, isLoading: ticketsLoading, error: ticketsErr } = useKDSTickets(selectedLocationId);
  const { data: metricsData, isLoading: metricsLoading } = useKDSMetrics(selectedLocationId);
  const bumpMutation = useBumpItem();

  const isLoading = capLoading || ticketsLoading;
  const error = capErr || ticketsErr;

  if (isLoading) return <LoadingSpinner />;
  if (error) return <ErrorBanner message={(error as Error).message} />;

  const stations = capacityData?.stations ?? [];
  const tickets = ticketsData?.tickets ?? [];

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex items-center gap-3">
        <ChefHat className="h-7 w-7 text-[#F97316]" />
        <div>
          <h1 className="text-2xl font-bold text-white">Kitchen Operations</h1>
          <p className="text-sm text-slate-400">Live station load and active ticket management</p>
        </div>
      </div>

      {/* ── Section 1: KDS Metrics ─────────────────────────────────────────── */}
      {metricsData && !metricsLoading && (
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
          <KPICard
            label="Avg Ticket Time"
            value={fmtSecs(metricsData.avg_ticket_time_secs ?? 0)}
            icon={Clock}
            iconColor="text-[#F97316]"
            bgTint="bg-orange-500/10"
          />
          <KPICard
            label="Items / Hour"
            value={(metricsData.items_per_hour ?? 0).toFixed(0)}
            icon={Zap}
            iconColor="text-blue-500"
            bgTint="bg-blue-500/10"
          />
          <KPICard
            label="Tickets Completed Today"
            value={(metricsData.tickets_completed ?? 0).toLocaleString()}
            icon={CheckCircle}
            iconColor="text-green-500"
            bgTint="bg-green-500/10"
          />
        </div>
      )}

      {/* ── Section 2: Station Load ────────────────────────────────────────── */}
      <div>
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold text-white">Station Load</h2>
          {capacityData && (
            <div className="flex items-center gap-4 text-sm text-slate-400">
              <span>
                Overall:{' '}
                <span
                  className={`font-bold ${
                    (capacityData.total_capacity_pct ?? 0) >= 80
                      ? 'text-red-600'
                      : (capacityData.total_capacity_pct ?? 0) >= 50
                      ? 'text-yellow-600'
                      : 'text-green-600'
                  }`}
                >
                  {(capacityData.total_capacity_pct ?? 0).toFixed(0)}%
                </span>
              </span>
              <span>{capacityData.active_tickets ?? 0} active tickets</span>
            </div>
          )}
        </div>

        {stations.length === 0 ? (
          <div className="text-center py-10 text-slate-300">No stations configured.</div>
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
            {stations.map((s) => (
              <StationCard key={s.station_id} station={s} />
            ))}
          </div>
        )}
      </div>

      {/* ── Section 3: Active Tickets (expo view) ──────────────────────────── */}
      <div>
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold text-white">Active Tickets</h2>
          <div className="flex items-center gap-3 text-xs text-slate-400">
            <span className="flex items-center gap-1">
              <span className="inline-block w-2.5 h-2.5 rounded-full bg-green-400" /> &lt;5 min
            </span>
            <span className="flex items-center gap-1">
              <span className="inline-block w-2.5 h-2.5 rounded-full bg-yellow-400" /> 5–10 min
            </span>
            <span className="flex items-center gap-1">
              <AlertTriangle className="h-3 w-3 text-red-500" /> &gt;10 min
            </span>
          </div>
        </div>

        {tickets.length === 0 ? (
          <div className="text-center py-16 text-slate-300">
            <ChefHat className="h-12 w-12 mx-auto mb-3 opacity-30" />
            <p className="text-lg font-medium">No active tickets</p>
            <p className="text-sm">Kitchen is clear</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
            {tickets.map((ticket) => (
              <TicketCard
                key={ticket.ticket_id}
                ticket={ticket}
                onBump={(itemId, status) => bumpMutation.mutate({ itemId, status })}
                bumping={bumpMutation.isPending}
              />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
