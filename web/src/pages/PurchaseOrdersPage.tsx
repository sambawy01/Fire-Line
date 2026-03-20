import { useState } from 'react';
import { CheckCircle, X, Eye, ShoppingCart } from 'lucide-react';
import { useLocationStore } from '../stores/location';
import { usePOs, usePO, useApprovePO, useCancelPO } from '../hooks/usePurchaseOrders';
import DataTable from '../components/ui/DataTable';
import type { Column } from '../components/ui/DataTable';
import StatusBadge from '../components/ui/StatusBadge';
import ErrorBanner from '../components/ui/ErrorBanner';
import LoadingSpinner from '../components/ui/LoadingSpinner';
import type { PurchaseOrder, POLine } from '../lib/api';

// ─── Helpers ─────────────────────────────────────────────────────────────────

function dollars(cents: number): string {
  return `$${(cents / 100).toFixed(2)}`;
}

function daysSince(iso: string | null): number {
  if (!iso) return 0;
  return Math.floor((Date.now() - new Date(iso).getTime()) / (1000 * 60 * 60 * 24));
}

function formatDate(iso: string | null): string {
  if (!iso) return '—';
  return new Date(iso).toLocaleDateString();
}

function poStatusVariant(status: PurchaseOrder['status']): 'info' | 'success' | 'neutral' | 'critical' {
  switch (status) {
    case 'draft':
      return 'info';
    case 'approved':
      return 'success';
    case 'received':
      return 'neutral';
    case 'cancelled':
      return 'critical';
  }
}

// ─── PO Detail Modal ──────────────────────────────────────────────────────────

const lineColumns: Column<POLine>[] = [
  { key: 'ingredient_name', header: 'Ingredient', sortable: true },
  {
    key: 'ordered_qty',
    header: 'Ordered',
    align: 'right',
    render: (r) => `${(r.ordered_qty ?? 0).toFixed(2)} ${r.ordered_unit ?? ''}`,
  },
  {
    key: 'estimated_unit_cost',
    header: 'Est. Unit Cost',
    align: 'right',
    render: (r) => dollars(r.estimated_unit_cost),
  },
  {
    key: 'est_total',
    header: 'Est. Total',
    align: 'right',
    render: (r) => dollars(r.ordered_qty * r.estimated_unit_cost),
  },
  {
    key: 'received_qty',
    header: 'Received',
    align: 'right',
    render: (r) =>
      r.received_qty != null ? `${r.received_qty.toFixed(2)} ${r.ordered_unit}` : '—',
  },
  {
    key: 'variance_flag',
    header: 'Variance',
    align: 'center',
    render: (r) => {
      if (!r.variance_flag) return <span className="text-gray-400">—</span>;
      const variant =
        r.variance_flag === 'over' ? 'critical' : r.variance_flag === 'under' ? 'warning' : 'neutral';
      return <StatusBadge variant={variant}>{r.variance_flag}</StatusBadge>;
    },
  },
  {
    key: 'note',
    header: 'Note',
    render: (r) => <span className="text-gray-500 text-xs">{r.note || '—'}</span>,
  },
];

function PODetailModal({ poId, onClose }: { poId: string; onClose: () => void }) {
  const { data, isLoading, error } = usePO(poId);

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
      <div className="bg-white rounded-xl shadow-xl w-full max-w-5xl max-h-[90vh] flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-gray-200">
          <div>
            <h2 className="text-lg font-bold text-gray-800">
              {isLoading ? 'Loading PO…' : data ? `PO — ${data.vendor_name}` : 'Purchase Order'}
            </h2>
            {data && (
              <p className="text-xs text-gray-500 mt-0.5">
                {data.line_count ?? 0} line{(data.line_count ?? 0) !== 1 ? 's' : ''} · Est. {dollars(data.total_estimated ?? 0)}
                {data.total_actual ? ` · Actual ${dollars(data.total_actual)}` : ''}
              </p>
            )}
          </div>
          <button
            onClick={onClose}
            className="p-2 rounded-lg text-gray-400 hover:text-gray-600 hover:bg-gray-100 transition-colors"
            aria-label="Close"
          >
            <X className="h-5 w-5" />
          </button>
        </div>

        {/* Body */}
        <div className="overflow-y-auto flex-1 px-6 py-4">
          {isLoading && (
            <div className="flex justify-center py-12">
              <LoadingSpinner size="lg" />
            </div>
          )}
          {error && (
            <ErrorBanner message={error instanceof Error ? error.message : 'Failed to load PO details'} />
          )}
          {data && !isLoading && (
            <div className="space-y-4">
              {/* Meta row */}
              <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 text-sm">
                <div>
                  <p className="text-xs text-gray-500 uppercase tracking-wide mb-0.5">Status</p>
                  <StatusBadge variant={poStatusVariant(data.status)}>{data.status}</StatusBadge>
                </div>
                <div>
                  <p className="text-xs text-gray-500 uppercase tracking-wide mb-0.5">Source</p>
                  <span className="text-gray-700 capitalize">{(data.source ?? 'manual').replace('_', ' ')}</span>
                </div>
                <div>
                  <p className="text-xs text-gray-500 uppercase tracking-wide mb-0.5">Suggested</p>
                  <span className="text-gray-700">{formatDate(data.suggested_at)}</span>
                </div>
                <div>
                  <p className="text-xs text-gray-500 uppercase tracking-wide mb-0.5">Approved</p>
                  <span className="text-gray-700">{formatDate(data.approved_at)}</span>
                </div>
              </div>

              {data.notes && (
                <div className="bg-gray-50 rounded-lg px-4 py-3 text-sm text-gray-700">
                  <span className="font-medium">Notes: </span>
                  {data.notes}
                </div>
              )}

              {/* Lines table */}
              <DataTable
                columns={lineColumns}
                data={data.lines}
                keyExtractor={(r) => r.po_line_id}
                emptyTitle="No line items"
                emptyDescription="This purchase order has no line items yet."
              />
            </div>
          )}
        </div>

        <div className="px-6 py-4 border-t border-gray-200 flex justify-end">
          <button
            onClick={onClose}
            className="px-4 py-2 rounded-lg bg-gray-100 text-gray-700 text-sm font-medium hover:bg-gray-200 transition-colors"
          >
            Close
          </button>
        </div>
      </div>
    </div>
  );
}

// ─── Section: Suggested POs ──────────────────────────────────────────────────

function SuggestedPOs({
  pos,
  onReview,
}: {
  pos: PurchaseOrder[];
  onReview: (id: string) => void;
}) {
  const approveMutation = useApprovePO();

  if (pos.length === 0) return null;

  return (
    <section>
      <div className="mb-3">
        <h2 className="text-lg font-semibold text-gray-800">Suggested Purchase Orders</h2>
        <p className="text-xs text-gray-500 mt-0.5">
          AI-recommended orders based on PAR levels and usage trends
        </p>
      </div>
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
        {pos.map((po) => (
          <div
            key={po.purchase_order_id}
            className="bg-white rounded-xl border border-gray-200 shadow-sm p-5 flex flex-col gap-4"
          >
            <div className="flex items-start justify-between">
              <div>
                <p className="font-semibold text-gray-800">{po.vendor_name}</p>
                <p className="text-xs text-gray-500 mt-0.5">
                  {po.line_count ?? 0} item{(po.line_count ?? 0) !== 1 ? 's' : ''} · Est. {dollars(po.total_estimated ?? 0)}
                </p>
              </div>
              <StatusBadge variant="info">Suggested</StatusBadge>
            </div>

            {po.notes && (
              <p className="text-xs text-gray-500 line-clamp-2">{po.notes}</p>
            )}

            <div className="flex gap-2 mt-auto">
              <button
                onClick={() => approveMutation.mutate(po.purchase_order_id)}
                disabled={approveMutation.isPending}
                className="flex-1 flex items-center justify-center gap-1.5 px-3 py-2 rounded-lg bg-emerald-600 text-white text-sm font-medium hover:bg-emerald-700 disabled:opacity-50 transition-colors"
              >
                <CheckCircle className="h-4 w-4" />
                Approve
              </button>
              <button
                onClick={() => onReview(po.purchase_order_id)}
                className="flex-1 flex items-center justify-center gap-1.5 px-3 py-2 rounded-lg border border-gray-300 text-gray-700 text-sm font-medium hover:bg-gray-50 transition-colors"
              >
                <Eye className="h-4 w-4" />
                Review
              </button>
            </div>
          </div>
        ))}
      </div>
    </section>
  );
}

// ─── Section: Active POs ──────────────────────────────────────────────────────

const activeColumns = (onReview: (id: string) => void): Column<PurchaseOrder>[] => [
  { key: 'vendor_name', header: 'Vendor', sortable: true },
  {
    key: 'line_count',
    header: 'Items',
    align: 'right',
    render: (r) => String(r.line_count ?? 0),
  },
  {
    key: 'total_estimated',
    header: 'Est. Total',
    align: 'right',
    sortable: true,
    render: (r) => dollars(r.total_estimated ?? 0),
  },
  {
    key: 'approved_at',
    header: 'Approved',
    render: (r) => formatDate(r.approved_at),
  },
  {
    key: 'days_waiting',
    header: 'Days Waiting',
    align: 'right',
    render: (r) => {
      const days = daysSince(r.approved_at);
      const color = days >= 5 ? 'text-red-600 font-semibold' : days >= 3 ? 'text-amber-600' : 'text-gray-700';
      return <span className={color}>{days}d</span>;
    },
  },
  {
    key: 'status',
    header: 'Status',
    align: 'center',
    render: (r) => <StatusBadge variant={poStatusVariant(r.status)}>{r.status}</StatusBadge>,
  },
  {
    key: 'actions',
    header: '',
    render: (r) => (
      <button
        onClick={() => onReview(r.purchase_order_id)}
        className="inline-flex items-center gap-1 px-2.5 py-1 rounded text-xs font-medium text-gray-600 hover:bg-gray-100 transition-colors"
      >
        <Eye className="h-3.5 w-3.5" />
        View
      </button>
    ),
  },
];

// ─── Section: PO History ──────────────────────────────────────────────────────

const historyColumns = (onReview: (id: string) => void): Column<PurchaseOrder>[] => [
  { key: 'vendor_name', header: 'Vendor', sortable: true },
  {
    key: 'total_estimated',
    header: 'Ordered $',
    align: 'right',
    sortable: true,
    render: (r) => dollars(r.total_estimated),
  },
  {
    key: 'total_actual',
    header: 'Actual $',
    align: 'right',
    render: (r) => (r.total_actual ? dollars(r.total_actual) : '—'),
  },
  {
    key: 'variance',
    header: 'Variance $',
    align: 'right',
    render: (r) => {
      if (!r.total_actual) return <span className="text-gray-400">—</span>;
      const diff = r.total_actual - r.total_estimated;
      const formatted = `${diff > 0 ? '+' : ''}${dollars(Math.abs(diff))}`;
      const color = diff < 0 ? 'text-emerald-600 font-medium' : diff > 0 ? 'text-red-600 font-medium' : 'text-gray-600';
      return <span className={color}>{diff > 0 ? `+${dollars(diff)}` : diff < 0 ? `-${dollars(Math.abs(diff))}` : '$0.00'}</span>;
    },
  },
  {
    key: 'status',
    header: 'Status',
    align: 'center',
    render: (r) => <StatusBadge variant={poStatusVariant(r.status)}>{r.status}</StatusBadge>,
  },
  {
    key: 'received_at',
    header: 'Date',
    render: (r) => formatDate(r.received_at ?? r.approved_at),
  },
  {
    key: 'actions',
    header: '',
    render: (r) => (
      <button
        onClick={() => onReview(r.purchase_order_id)}
        className="inline-flex items-center gap-1 px-2.5 py-1 rounded text-xs font-medium text-gray-600 hover:bg-gray-100 transition-colors"
      >
        <Eye className="h-3.5 w-3.5" />
        View
      </button>
    ),
  },
];

// ─── Page ─────────────────────────────────────────────────────────────────────

export default function PurchaseOrdersPage() {
  const locationId = useLocationStore((s) => s.selectedLocationId);
  const [selectedPoId, setSelectedPoId] = useState<string | null>(null);

  // Fetch all POs (no status filter — we'll partition client-side)
  const { data, isLoading, error, refetch } = usePOs(locationId, undefined);

  if (!locationId) return <LoadingSpinner fullPage />;

  const allPOs = data?.purchase_orders ?? [];

  const suggestedPOs = allPOs.filter(
    (po) => po.status === 'draft' && po.source === 'system_recommended'
  );
  const activePOs = allPOs.filter((po) => po.status === 'approved');
  const historyPOs = allPOs.filter(
    (po) => po.status === 'received' || po.status === 'cancelled'
  );

  return (
    <div className="space-y-8">
      {/* Page header */}
      <div className="flex items-center gap-3">
        <ShoppingCart className="h-7 w-7 text-orange-500" />
        <div>
          <h1 className="text-2xl font-bold text-gray-800">Purchase Orders</h1>
          <p className="text-sm text-gray-500 mt-0.5">
            AI-suggested orders, active deliveries, and receiving history
          </p>
        </div>
      </div>

      {error && (
        <ErrorBanner
          message={error instanceof Error ? error.message : 'Failed to load purchase orders'}
          retry={() => refetch()}
        />
      )}

      {isLoading && (
        <div className="flex justify-center py-16">
          <LoadingSpinner size="lg" />
        </div>
      )}

      {!isLoading && !error && (
        <>
          {/* Section 1 — Suggested */}
          <SuggestedPOs pos={suggestedPOs} onReview={setSelectedPoId} />

          {/* Section 2 — Active */}
          <section>
            <div className="mb-3">
              <h2 className="text-lg font-semibold text-gray-800">Active Orders</h2>
              <p className="text-xs text-gray-500 mt-0.5">Approved orders awaiting delivery</p>
            </div>
            <DataTable
              columns={activeColumns(setSelectedPoId)}
              data={activePOs}
              keyExtractor={(r) => r.purchase_order_id}
              emptyTitle="No active orders"
              emptyDescription="Approve a suggested order or create one manually to see it here."
            />
          </section>

          {/* Section 3 — History */}
          <section>
            <div className="mb-3">
              <h2 className="text-lg font-semibold text-gray-800">Order History</h2>
              <p className="text-xs text-gray-500 mt-0.5">
                Received and cancelled orders — green variance = savings, red = overage
              </p>
            </div>
            <DataTable
              columns={historyColumns(setSelectedPoId)}
              data={historyPOs}
              keyExtractor={(r) => r.purchase_order_id}
              emptyTitle="No order history"
              emptyDescription="Received and cancelled orders will appear here."
            />
          </section>
        </>
      )}

      {/* Detail modal */}
      {selectedPoId && (
        <PODetailModal poId={selectedPoId} onClose={() => setSelectedPoId(null)} />
      )}
    </div>
  );
}
