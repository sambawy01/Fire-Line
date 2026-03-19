import { create } from 'zustand';
import { api, setToken } from '../lib/api';

interface ActiveStaff {
  employee_id: string;
  display_name: string;
  role: string;
}

interface Location {
  id: string;
  name: string;
}

interface AuthState {
  managerToken: string | null;
  activeStaff: ActiveStaff | null;
  locationId: string | null;
  locationName: string | null;
  locations: Location[];
  lastActivity: number;

  login: (email: string, password: string) => Promise<void>;
  selectLocation: (id: string, name: string) => void;
  pinVerify: (pin: string) => Promise<void>;
  checkTimeout: () => void;
  touchActivity: () => void;
  logout: () => void;
}

interface LoginResponse {
  token: string;
}

interface LocationsResponse {
  locations: Location[];
}

interface PinVerifyResponse {
  employee_id: string;
  display_name: string;
  role: string;
}

const INACTIVITY_TIMEOUT_MS = 120_000; // 2 minutes

export const useAuthStore = create<AuthState>((set, get) => ({
  managerToken: null,
  activeStaff: null,
  locationId: null,
  locationName: null,
  locations: [],
  lastActivity: Date.now(),

  login: async (email: string, password: string) => {
    const data = await api.post<LoginResponse>('/auth/login', {
      email,
      password,
    });
    setToken(data.token);

    const locData = await api.get<LocationsResponse>('/locations');

    set({
      managerToken: data.token,
      locations: locData.locations ?? [],
      activeStaff: null,
      locationId: null,
      locationName: null,
    });
  },

  selectLocation: (id: string, name: string) => {
    set({ locationId: id, locationName: name });
  },

  pinVerify: async (pin: string) => {
    const { locationId } = get();
    const data = await api.post<PinVerifyResponse>('/auth/pin-verify', {
      pin,
      location_id: locationId,
    });
    set({
      activeStaff: {
        employee_id: data.employee_id,
        display_name: data.display_name,
        role: data.role,
      },
      lastActivity: Date.now(),
    });
  },

  checkTimeout: () => {
    const { lastActivity } = get();
    if (Date.now() - lastActivity > INACTIVITY_TIMEOUT_MS) {
      set({ activeStaff: null });
    }
  },

  touchActivity: () => {
    set({ lastActivity: Date.now() });
  },

  logout: () => {
    setToken(null);
    set({
      managerToken: null,
      activeStaff: null,
      locationId: null,
      locationName: null,
      locations: [],
      lastActivity: Date.now(),
    });
  },
}));
