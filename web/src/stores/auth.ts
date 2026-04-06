import { create } from 'zustand';
import { authApi } from '../lib/api';

interface AuthState {
  accessToken: string | null;
  orgId: string | null;
  userId: string | null;
  role: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;
  login: (email: string, password: string) => Promise<void>;
  signup: (data: { org_name: string; org_slug: string; email: string; password: string; display_name: string }) => Promise<void>;
  logout: () => Promise<void>;
  clearError: () => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  accessToken: localStorage.getItem('access_token'),
  orgId: localStorage.getItem('org_id'),
  userId: localStorage.getItem('user_id'),
  role: localStorage.getItem('role'),
  isAuthenticated: !!localStorage.getItem('access_token'),
  isLoading: false,
  error: null,

  login: async (email, password) => {
    set({ isLoading: true, error: null });
    try {
      const result = await authApi.login({ email, password });
      if (result.mfa_required) {
        set({ isLoading: false, error: 'MFA required (not yet supported in UI)' });
        return;
      }
      localStorage.setItem('access_token', result.access_token);
      if (result.org_id) localStorage.setItem('org_id', result.org_id);
      if (result.user_id) localStorage.setItem('user_id', result.user_id);
      if (result.role) localStorage.setItem('role', result.role);
      set({
        accessToken: result.access_token,
        orgId: result.org_id || null,
        userId: result.user_id || null,
        role: result.role || null,
        isAuthenticated: true,
        isLoading: false,
      });
    } catch (err: any) {
      set({ isLoading: false, error: err.message || 'Login failed' });
    }
  },

  signup: async (data) => {
    set({ isLoading: true, error: null });
    try {
      const result = await authApi.signup(data);
      localStorage.setItem('access_token', result.access_token);
      if (result.org_id) localStorage.setItem('org_id', result.org_id);
      if (result.user_id) localStorage.setItem('user_id', result.user_id);
      set({
        accessToken: result.access_token,
        orgId: result.org_id || null,
        userId: result.user_id || null,
        role: 'owner',
        isAuthenticated: true,
        isLoading: false,
      });
    } catch (err: any) {
      set({ isLoading: false, error: err.message || 'Signup failed' });
    }
  },

  logout: async () => {
    await authApi.logout();
    localStorage.removeItem('access_token');
    localStorage.removeItem('org_id');
    localStorage.removeItem('user_id');
    localStorage.removeItem('role');
    set({
      accessToken: null,
      orgId: null,
      userId: null,
      role: null,
      isAuthenticated: false,
    });
  },

  clearError: () => set({ error: null }),
}));
