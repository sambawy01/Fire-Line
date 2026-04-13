import {
  Clock,
  Plug,
  RefreshCw,
  Download,
  AlertCircle,
  Loader2,
  Wifi,
  WifiOff,
} from 'lucide-react';
import {
  useLoyverseStatus,
  useLoyverseSync,
  useLoyverseImport,
} from '../hooks/useAdapters';
import type { AdapterStatusValue } from '../lib/api';

// Loyverse capabilities are fixed and known from the adapter implementation.
const LOYVERSE_CAPABILITIES = [
  'READ_ORDERS',
  'READ_MENU',
  'READ_INVENTORY',
  'READ_EMPLOYEES',
  'WRITE_86_STATUS',
];

const STATUS_CONFIG: Record<
  AdapterStatusValue,
  { label: string; dot: string; bg: string; text: string }
> = {
  active: {
    label: 'Active',
    dot: 'bg-green-500',
    bg: 'bg-green-50',
    text: 'text-green-700',
  },
  initializing: {
    label: 'Initializing',
    dot: 'bg-blue-500',
    bg: 'bg-blue-50',
    text: 'text-blue-700',
  },
  errored: {
    label: 'Errored',
    dot: 'bg-red-500',
    bg: 'bg-red-50',
    text: 'text-red-700',
  },
  disconnected: {
    label: 'Disconnected',
    dot: 'bg-slate-400',
    bg: 'bg-slate-100',
    text: 'text-slate-600',
  },
};

function formatTimestamp(iso: string): string {
  const date = new Date(iso);
  return date.toLocaleString('en-US', {
    month: 'short',
    day: 'numeric',
    hour: 'numeric',
    minute: '2-digit',
    hour12: true,
  });
}

/** Find the most recent sync time across all freshness entries. */
function latestSyncTime(freshness: Record<string, { LastSyncAt: string }>): string | null {
  let latest: string | null = null;
  for (const entry of Object.values(freshness)) {
    if (!latest || entry.LastSyncAt > latest) {
      latest = entry.LastSyncAt;
    }
  }
  return latest;
}

export default function AdaptersPage() {
  const { data: statusData, isLoading, error } = useLoyverseStatus();
  const syncMutation = useLoyverseSync();
  const importMutation = useLoyverseImport();

  const adapterStatus = statusData?.status ?? 'disconnected';
  const statusCfg = STATUS_CONFIG[adapterStatus];
  const isActive = adapterStatus === 'active';
  const lastSync = statusData?.freshness
    ? latestSyncTime(statusData.freshness)
    : null;

  return (
    <div className="min-h-screen">
      <div className="mx-auto max-w-5xl px-4 py-8 sm:px-6 lg:px-8">
        {/* Header */}
        <div className="mb-6 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <Plug className="h-7 w-7 text-white" />
            <h1 className="text-2xl font-bold text-white">
              POS Connections
            </h1>
          </div>
        </div>

        {/* Loading State */}
        {isLoading && (
          <div className="flex items-center justify-center py-20">
            <Loader2 className="h-8 w-8 animate-spin text-slate-400" />
            <span className="ml-3 text-slate-400">Loading adapter status...</span>
          </div>
        )}

        {/* Error State */}
        {error && !statusData && (
          <div className="rounded-xl border border-red-500/20 bg-red-500/10 p-6">
            <div className="flex items-center gap-3">
              <AlertCircle className="h-6 w-6 text-red-400" />
              <div>
                <p className="font-semibold text-red-300">
                  Failed to load adapter status
                </p>
                <p className="mt-1 text-sm text-red-400/80">
                  {error instanceof Error ? error.message : 'Unknown error'}
                </p>
              </div>
            </div>
          </div>
        )}

        {/* Adapter Card */}
        {!isLoading && statusData && (
          <div className="grid gap-6 sm:grid-cols-2">
            <div className="rounded-xl border border-white/10 bg-white/5 shadow-sm">
              <div className="p-5">
                {/* Top row: logo + name + status */}
                <div className="mb-4 flex items-start gap-3">
                  {/* Logo */}
                  <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-full bg-[#7C3AED] text-lg font-bold text-white">
                    L
                  </div>

                  <div className="min-w-0 flex-1">
                    <h3 className="truncate text-base font-semibold text-white">
                      Loyverse POS
                    </h3>
                    <p className="text-sm text-slate-400">
                      Loyverse &middot; All Branches
                    </p>
                  </div>

                  {/* Status badge */}
                  <span
                    className={`inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-xs font-semibold ${statusCfg.bg} ${statusCfg.text}`}
                  >
                    <span
                      className={`inline-block h-2 w-2 rounded-full ${statusCfg.dot}`}
                    />
                    {statusCfg.label}
                  </span>
                </div>

                {/* Capabilities */}
                <div className="mb-4 flex flex-wrap gap-1.5">
                  {LOYVERSE_CAPABILITIES.map((cap) => (
                    <span
                      key={cap}
                      className="rounded-md bg-gray-100 px-2 py-0.5 text-xs font-medium text-slate-300"
                    >
                      {cap}
                    </span>
                  ))}
                </div>

                {/* Last sync */}
                <p className="mb-4 flex items-center gap-1 text-xs text-slate-300">
                  <Clock className="h-3.5 w-3.5" />
                  {lastSync
                    ? `Last sync: ${formatTimestamp(lastSync)}`
                    : 'No sync data yet'}
                </p>

                {/* Freshness details */}
                {statusData.freshness && Object.keys(statusData.freshness).length > 0 && (
                  <div className="mb-4 space-y-1">
                    {Object.entries(statusData.freshness).map(([key, entry]) => (
                      <p key={key} className="text-xs text-slate-400">
                        <span className="capitalize">{key}</span>: {entry.RecordCount} records, synced {formatTimestamp(entry.LastSyncAt)}
                      </p>
                    ))}
                  </div>
                )}

                {/* Actions */}
                <div className="flex flex-wrap items-center gap-3 border-t border-white/5 pt-4">
                  {isActive ? (
                    <>
                      <button
                        onClick={() => syncMutation.mutate()}
                        disabled={syncMutation.isPending}
                        className="inline-flex items-center gap-1.5 rounded-lg bg-[#F97316] px-4 py-1.5 text-sm font-medium text-white transition-colors hover:bg-[#EA580C] disabled:opacity-50"
                      >
                        {syncMutation.isPending ? (
                          <Loader2 className="h-4 w-4 animate-spin" />
                        ) : (
                          <RefreshCw className="h-4 w-4" />
                        )}
                        {syncMutation.isPending ? 'Syncing...' : 'Trigger Sync'}
                      </button>
                      <button
                        onClick={() => importMutation.mutate(30)}
                        disabled={importMutation.isPending}
                        className="inline-flex items-center gap-1.5 rounded-lg bg-white px-4 py-1.5 text-sm font-medium text-slate-200 ring-1 ring-gray-200 transition-colors hover:bg-white/5 disabled:opacity-50"
                      >
                        {importMutation.isPending ? (
                          <Loader2 className="h-4 w-4 animate-spin" />
                        ) : (
                          <Download className="h-4 w-4" />
                        )}
                        {importMutation.isPending ? 'Importing...' : 'Import Historical'}
                      </button>
                    </>
                  ) : (
                    <div className="flex items-center gap-3">
                      <WifiOff className="h-4 w-4 text-slate-400" />
                      <span className="text-sm text-slate-400">
                        Adapter is {adapterStatus}
                      </span>
                    </div>
                  )}
                </div>

                {/* Mutation feedback */}
                {syncMutation.isSuccess && (
                  <p className="mt-3 flex items-center gap-1.5 text-xs text-green-400">
                    <Wifi className="h-3.5 w-3.5" />
                    Sync triggered successfully
                  </p>
                )}
                {syncMutation.isError && (
                  <p className="mt-3 flex items-center gap-1.5 text-xs text-red-400">
                    <AlertCircle className="h-3.5 w-3.5" />
                    Sync failed: {syncMutation.error instanceof Error ? syncMutation.error.message : 'Unknown error'}
                  </p>
                )}
                {importMutation.isSuccess && (
                  <p className="mt-3 flex items-center gap-1.5 text-xs text-green-400">
                    <Wifi className="h-3.5 w-3.5" />
                    Import complete &mdash; {importMutation.data.orders_synced} orders, {importMutation.data.items_synced} items, {importMutation.data.employees_synced} employees
                  </p>
                )}
                {importMutation.isError && (
                  <p className="mt-3 flex items-center gap-1.5 text-xs text-red-400">
                    <AlertCircle className="h-3.5 w-3.5" />
                    Import failed: {importMutation.error instanceof Error ? importMutation.error.message : 'Unknown error'}
                  </p>
                )}
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
