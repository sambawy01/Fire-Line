import { create } from 'zustand';
import { api } from '../lib/api';
import { savePending, loadPending, clearPending } from '../lib/offline';
import { useAuthStore } from './auth';

const OFFLINE_KEY = 'count_lines';
const AUTO_SYNC_INTERVAL_MS = 10_000;

export interface CountLine {
  count_line_id: string;
  ingredient_id: string;
  name: string;
  category: string;
  expected_qty: number;
  counted_qty: number | null;
  unit: string;
  note: string;
}

export interface PendingLine {
  ingredient_id: string;
  counted_qty: number;
  unit: string;
  note: string;
}

export interface ActiveCount {
  count_id: string;
  status: string;
  started_at: string;
}

interface CountProgress {
  counted: number;
  total: number;
}

interface StartCountResponse {
  count_id: string;
  status: string;
  started_at: string;
}

interface CountDetailResponse {
  count_id: string;
  status: string;
  started_at: string;
  lines: CountLine[];
  progress: CountProgress;
}

interface SyncLinesResponse {
  updated: number;
}

interface SubmitCountResponse {
  count_id: string;
  status: string;
}

interface CountState {
  activeCount: ActiveCount | null;
  lines: CountLine[];
  progress: CountProgress;
  pendingSync: PendingLine[];
  syncing: boolean;
  error: string | null;
  autoSyncTimer: ReturnType<typeof setInterval> | null;

  startCount: (countType: 'full' | 'spot', category?: string) => Promise<void>;
  loadCount: (countId: string) => Promise<void>;
  updateLine: (ingredientId: string, qty: number, note?: string) => void;
  syncLines: () => Promise<void>;
  submitCount: () => Promise<void>;
  resetCount: () => void;
  startAutoSync: () => void;
  stopAutoSync: () => void;
}

export const useCountStore = create<CountState>((set, get) => ({
  activeCount: null,
  lines: [],
  progress: { counted: 0, total: 0 },
  pendingSync: [],
  syncing: false,
  error: null,
  autoSyncTimer: null,

  startCount: async (countType, category) => {
    const { locationId, activeStaff } = useAuthStore.getState();
    set({ error: null });
    try {
      const data = await api.post<StartCountResponse>('/inventory/counts', {
        location_id: locationId,
        count_type: countType,
        counted_by: activeStaff?.employee_id,
        ...(category ? { category } : {}),
      });
      set({
        activeCount: {
          count_id: data.count_id,
          status: data.status,
          started_at: data.started_at,
        },
        pendingSync: [],
      });
      await get().loadCount(data.count_id);
      get().startAutoSync();
    } catch (e: any) {
      set({ error: e.message ?? 'Failed to start count' });
      throw e;
    }
  },

  loadCount: async (countId) => {
    set({ error: null });
    try {
      const data = await api.get<CountDetailResponse>(`/inventory/counts/${countId}`);
      // Restore any pending lines from offline storage
      const offlinePending = await loadPending(OFFLINE_KEY);
      set({
        activeCount: {
          count_id: data.count_id,
          status: data.status,
          started_at: data.started_at,
        },
        lines: data.lines ?? [],
        progress: data.progress ?? { counted: 0, total: 0 },
        pendingSync: offlinePending,
      });
    } catch (e: any) {
      set({ error: e.message ?? 'Failed to load count' });
      throw e;
    }
  },

  updateLine: (ingredientId, qty, note = '') => {
    const { lines, pendingSync } = get();

    // Update local line
    const updatedLines = lines.map((l) =>
      l.ingredient_id === ingredientId
        ? { ...l, counted_qty: qty, note }
        : l,
    );

    // Upsert into pendingSync
    const existingIndex = pendingSync.findIndex((p) => p.ingredient_id === ingredientId);
    const line = updatedLines.find((l) => l.ingredient_id === ingredientId);
    const newPending: PendingLine = {
      ingredient_id: ingredientId,
      counted_qty: qty,
      unit: line?.unit ?? '',
      note,
    };

    let updatedPending: PendingLine[];
    if (existingIndex >= 0) {
      updatedPending = pendingSync.map((p, i) => (i === existingIndex ? newPending : p));
    } else {
      updatedPending = [...pendingSync, newPending];
    }

    // Recalculate progress
    const counted = updatedLines.filter((l) => l.counted_qty !== null).length;
    const total = updatedLines.length;

    set({
      lines: updatedLines,
      pendingSync: updatedPending,
      progress: { counted, total },
    });

    // Persist to offline storage async (fire and forget)
    savePending(OFFLINE_KEY, updatedPending).catch(() => {});
  },

  syncLines: async () => {
    const { activeCount, pendingSync, syncing } = get();
    if (!activeCount || pendingSync.length === 0 || syncing) return;

    set({ syncing: true });
    const toSync = [...pendingSync];
    try {
      await api.post<SyncLinesResponse>(`/inventory/counts/${activeCount.count_id}/lines`, {
        lines: toSync,
      });
      // Remove only the lines we just synced (in case more were added during the request)
      const { pendingSync: current } = get();
      const remaining = current.filter(
        (p) => !toSync.some((s) => s.ingredient_id === p.ingredient_id),
      );
      set({ pendingSync: remaining, syncing: false });
      await savePending(OFFLINE_KEY, remaining);
    } catch (e: any) {
      set({ syncing: false, error: e.message ?? 'Sync failed' });
    }
  },

  submitCount: async () => {
    const { activeCount } = get();
    if (!activeCount) return;

    // Sync any remaining pending lines first
    await get().syncLines();

    try {
      await api.put<SubmitCountResponse>(`/inventory/counts/${activeCount.count_id}`, {
        status: 'submitted',
      });
      set({
        activeCount: { ...activeCount, status: 'submitted' },
      });
      await clearPending(OFFLINE_KEY);
      get().stopAutoSync();
    } catch (e: any) {
      set({ error: e.message ?? 'Failed to submit count' });
      throw e;
    }
  },

  resetCount: () => {
    get().stopAutoSync();
    clearPending(OFFLINE_KEY).catch(() => {});
    set({
      activeCount: null,
      lines: [],
      progress: { counted: 0, total: 0 },
      pendingSync: [],
      error: null,
    });
  },

  startAutoSync: () => {
    const { autoSyncTimer } = get();
    if (autoSyncTimer) return; // already running

    const timer = setInterval(() => {
      const { pendingSync } = get();
      if (pendingSync.length > 0) {
        get().syncLines();
      }
    }, AUTO_SYNC_INTERVAL_MS);

    set({ autoSyncTimer: timer });
  },

  stopAutoSync: () => {
    const { autoSyncTimer } = get();
    if (autoSyncTimer) {
      clearInterval(autoSyncTimer);
      set({ autoSyncTimer: null });
    }
  },
}));
