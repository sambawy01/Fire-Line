import { create } from 'zustand';
import { api } from '../lib/api';
import { savePending, loadPending, clearPending } from '../lib/offline';
import { useAuthStore } from './auth';

const OFFLINE_KEY = 'receive_lines';

export interface PurchaseOrder {
  purchase_order_id: string;
  vendor_name: string;
  line_count: number;
  total_estimated: number;
  approved_at: string;
}

export interface POLine {
  po_line_id: string;
  ingredient_name: string;
  ordered_qty: number;
  ordered_unit: string;
  estimated_unit_cost: number;
}

export interface POWithLines {
  purchase_order_id: string;
  vendor_name: string;
  lines: POLine[];
}

export interface ReceivedLineEntry {
  received_qty: number;
  received_unit_cost: number;
  note: string;
  verified: boolean;
}

export type DiscrepancyFlag = 'exact' | 'short' | 'over' | 'not_received';

export interface Discrepancy {
  po_line_id: string;
  ingredient_name: string;
  ordered_qty: number;
  ordered_unit: string;
  received_qty: number;
  flag: DiscrepancyFlag;
}

interface PendingPOsResponse {
  pending: PurchaseOrder[];
}

interface SubmitReceivingResponse {
  status: string;
  total_actual: number;
  discrepancies: unknown[];
}

interface ReceiveState {
  pendingPOs: PurchaseOrder[];
  activePO: POWithLines | null;
  receivedLines: Record<string, ReceivedLineEntry>;
  phase: 'list' | 'receiving' | 'review';
  loading: boolean;
  submitting: boolean;
  error: string | null;

  loadPending: () => Promise<void>;
  startReceiving: (poId: string) => Promise<void>;
  updateLine: (poLineId: string, qty: number, cost: number, note: string) => void;
  markNotReceived: (poLineId: string) => void;
  getProgress: () => { verified: number; total: number };
  getDiscrepancies: () => Discrepancy[];
  submitReceiving: () => Promise<void>;
  reset: () => void;
}

function computeFlag(orderedQty: number, receivedQty: number): DiscrepancyFlag {
  if (receivedQty === 0) return 'not_received';
  const ratio = receivedQty / orderedQty;
  if (ratio >= 0.98 && ratio <= 1.02) return 'exact';
  if (receivedQty < orderedQty) return 'short';
  return 'over';
}

export const useReceiveStore = create<ReceiveState>((set, get) => ({
  pendingPOs: [],
  activePO: null,
  receivedLines: {},
  phase: 'list',
  loading: false,
  submitting: false,
  error: null,

  loadPending: async () => {
    const { locationId } = useAuthStore.getState();
    set({ loading: true, error: null });
    try {
      const data = await api.get<PendingPOsResponse>(
        `/inventory/po/pending?location_id=${locationId ?? ''}`,
      );
      set({ pendingPOs: data.pending ?? [], loading: false });
    } catch (e: any) {
      set({ error: e.message ?? 'Failed to load pending POs', loading: false });
    }
  },

  startReceiving: async (poId: string) => {
    set({ loading: true, error: null });
    try {
      const data = await api.get<POWithLines>(`/inventory/po/${poId}`);

      // Pre-fill receivedLines with ordered values
      const prefilled: Record<string, ReceivedLineEntry> = {};
      for (const line of data.lines ?? []) {
        prefilled[line.po_line_id] = {
          received_qty: line.ordered_qty,
          received_unit_cost: line.estimated_unit_cost,
          note: '',
          verified: false,
        };
      }

      // Restore any offline-saved progress
      const offlineSaved = await loadPending(OFFLINE_KEY);
      if (Array.isArray(offlineSaved) && offlineSaved.length > 0) {
        for (const entry of offlineSaved) {
          if (entry.po_line_id && prefilled[entry.po_line_id]) {
            prefilled[entry.po_line_id] = {
              received_qty: entry.received_qty,
              received_unit_cost: entry.received_unit_cost,
              note: entry.note ?? '',
              verified: entry.verified ?? false,
            };
          }
        }
      }

      set({
        activePO: data,
        receivedLines: prefilled,
        phase: 'receiving',
        loading: false,
      });
    } catch (e: any) {
      set({ error: e.message ?? 'Failed to load PO', loading: false });
    }
  },

  updateLine: (poLineId: string, qty: number, cost: number, note: string) => {
    const { receivedLines } = get();
    const updated = {
      ...receivedLines,
      [poLineId]: {
        received_qty: qty,
        received_unit_cost: cost,
        note,
        verified: true,
      },
    };
    set({ receivedLines: updated });

    // Persist to offline storage
    const toSave = Object.entries(updated).map(([id, entry]) => ({
      po_line_id: id,
      ...entry,
    }));
    savePending(OFFLINE_KEY, toSave).catch(() => {});
  },

  markNotReceived: (poLineId: string) => {
    const { receivedLines } = get();
    const existing = receivedLines[poLineId];
    const updated = {
      ...receivedLines,
      [poLineId]: {
        received_qty: 0,
        received_unit_cost: existing?.received_unit_cost ?? 0,
        note: existing?.note ?? '',
        verified: true,
      },
    };
    set({ receivedLines: updated });

    const toSave = Object.entries(updated).map(([id, entry]) => ({
      po_line_id: id,
      ...entry,
    }));
    savePending(OFFLINE_KEY, toSave).catch(() => {});
  },

  getProgress: () => {
    const { activePO, receivedLines } = get();
    if (!activePO) return { verified: 0, total: 0 };
    const total = activePO.lines.length;
    const verified = activePO.lines.filter(
      (l) => receivedLines[l.po_line_id]?.verified,
    ).length;
    return { verified, total };
  },

  getDiscrepancies: () => {
    const { activePO, receivedLines } = get();
    if (!activePO) return [];
    return activePO.lines
      .map((line) => {
        const entry = receivedLines[line.po_line_id];
        const flag = computeFlag(line.ordered_qty, entry?.received_qty ?? 0);
        return { po_line_id: line.po_line_id, ingredient_name: line.ingredient_name, ordered_qty: line.ordered_qty, ordered_unit: line.ordered_unit, received_qty: entry?.received_qty ?? 0, flag };
      })
      .filter((d) => d.flag !== 'exact');
  },

  submitReceiving: async () => {
    const { activePO, receivedLines } = get();
    if (!activePO) return;

    set({ submitting: true, error: null });
    try {
      const lines = activePO.lines.map((line) => {
        const entry = receivedLines[line.po_line_id];
        return {
          po_line_id: line.po_line_id,
          received_qty: entry?.received_qty ?? 0,
          received_unit_cost: entry?.received_unit_cost ?? 0,
          note: entry?.note ?? '',
        };
      });

      await api.post<SubmitReceivingResponse>(
        `/inventory/po/${activePO.purchase_order_id}/receive`,
        { lines },
      );

      await clearPending(OFFLINE_KEY);
      set({
        activePO: null,
        receivedLines: {},
        phase: 'list',
        submitting: false,
      });
    } catch (e: any) {
      set({ error: e.message ?? 'Failed to submit receiving', submitting: false });
      throw e;
    }
  },

  reset: () => {
    clearPending(OFFLINE_KEY).catch(() => {});
    set({
      activePO: null,
      receivedLines: {},
      phase: 'list',
      error: null,
    });
  },
}));
