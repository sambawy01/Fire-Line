import { create } from 'zustand';
import { locationApi, type Location } from '../lib/api';

interface LocationState {
  selectedLocationId: string | null;
  locations: Location[];
  isLoading: boolean;
  setLocation: (id: string) => void;
  loadLocations: () => Promise<void>;
  clear: () => void;
}

export const useLocationStore = create<LocationState>((set, get) => ({
  selectedLocationId: localStorage.getItem('selected_location_id'),
  locations: [],
  isLoading: false,

  setLocation: (id: string) => {
    localStorage.setItem('selected_location_id', id);
    set({ selectedLocationId: id });
  },

  loadLocations: async () => {
    set({ isLoading: true });
    try {
      const { locations } = await locationApi.getLocations();
      const current = get().selectedLocationId;
      const validSelection = locations.some((l) => l.id === current);
      set({
        locations,
        isLoading: false,
        selectedLocationId: validSelection ? current : locations[0]?.id ?? null,
      });
      // Persist the auto-selected location
      const finalId = validSelection ? current : locations[0]?.id ?? null;
      if (finalId) localStorage.setItem('selected_location_id', finalId);
    } catch {
      set({ isLoading: false });
    }
  },

  clear: () => {
    localStorage.removeItem('selected_location_id');
    set({ selectedLocationId: null, locations: [] });
  },
}));
