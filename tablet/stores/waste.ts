import { create } from 'zustand';
import { api } from '../lib/api';
import { useAuthStore } from './auth';

export interface WasteLog {
  waste_id: string;
  ingredient_id: string;
  ingredient_name: string;
  quantity: number;
  unit: string;
  reason: string;
  logged_by_name: string;
  logged_at: string;
  note: string;
}

export interface WasteIngredient {
  ingredient_id: string;
  name: string;
  unit: string;
  category: string;
}

export type WasteReason =
  | 'expired'
  | 'dropped'
  | 'overcooked'
  | 'contaminated'
  | 'overproduction'
  | 'other';

export interface LogWasteInput {
  ingredient_id: string;
  quantity: number;
  unit: string;
  reason: WasteReason;
  note?: string;
}

interface WasteLogResponse {
  waste_id: string;
  logged_at: string;
}

interface WasteListResponse {
  logs: WasteLog[];
}

interface UsageIngredient {
  ingredient_id: string;
  name: string;
  unit: string;
  category: string;
}

interface UsageResponse {
  ingredients: UsageIngredient[];
}

interface WasteState {
  todaysLogs: WasteLog[];
  ingredients: WasteIngredient[];
  loading: boolean;
  error: string | null;

  loadLogs: () => Promise<void>;
  loadIngredients: () => Promise<void>;
  logWaste: (input: LogWasteInput) => Promise<void>;
}

function todayRange(): { from: string; to: string } {
  const now = new Date();
  const from = new Date(now.getFullYear(), now.getMonth(), now.getDate());
  const to = new Date(from.getTime() + 24 * 60 * 60 * 1000);
  return {
    from: from.toISOString().split('T')[0],
    to: to.toISOString().split('T')[0],
  };
}

export const useWasteStore = create<WasteState>((set, get) => ({
  todaysLogs: [],
  ingredients: [],
  loading: false,
  error: null,

  loadLogs: async () => {
    const { locationId } = useAuthStore.getState();
    set({ loading: true, error: null });
    try {
      const { from, to } = todayRange();
      const data = await api.get<WasteListResponse>(
        `/inventory/waste?location_id=${locationId}&from=${from}&to=${to}`,
      );
      set({ todaysLogs: data.logs ?? [], loading: false });
    } catch (e: any) {
      set({ error: e.message ?? 'Failed to load waste logs', loading: false });
    }
  },

  loadIngredients: async () => {
    const { locationId } = useAuthStore.getState();
    set({ error: null });
    try {
      const data = await api.get<UsageResponse>(`/inventory/usage?location_id=${locationId}`);
      set({
        ingredients: (data.ingredients ?? []).map((i) => ({
          ingredient_id: i.ingredient_id,
          name: i.name,
          unit: i.unit,
          category: i.category,
        })),
      });
    } catch (e: any) {
      set({ error: e.message ?? 'Failed to load ingredients' });
    }
  },

  logWaste: async (input) => {
    const { locationId, activeStaff } = useAuthStore.getState();
    set({ error: null });
    try {
      const data = await api.post<WasteLogResponse>('/inventory/waste', {
        location_id: locationId,
        ingredient_id: input.ingredient_id,
        quantity: input.quantity,
        unit: input.unit,
        reason: input.reason,
        logged_by: activeStaff?.employee_id,
        note: input.note ?? '',
      });

      const { ingredients } = get();
      const ingredient = ingredients.find((i) => i.ingredient_id === input.ingredient_id);

      const newLog: WasteLog = {
        waste_id: data.waste_id,
        ingredient_id: input.ingredient_id,
        ingredient_name: ingredient?.name ?? input.ingredient_id,
        quantity: input.quantity,
        unit: input.unit,
        reason: input.reason,
        logged_by_name: activeStaff?.display_name ?? '',
        logged_at: data.logged_at,
        note: input.note ?? '',
      };

      set((state) => ({ todaysLogs: [newLog, ...state.todaysLogs] }));
    } catch (e: any) {
      set({ error: e.message ?? 'Failed to log waste' });
      throw e;
    }
  },
}));
