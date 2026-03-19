const API_BASE = '/api/v1';

class ApiError extends Error {
  status: number;
  code: string;
  constructor(status: number, code: string, message: string) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.code = code;
  }
}

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const token = localStorage.getItem('access_token');
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
  };

  const res = await fetch(`${API_BASE}${path}`, { ...options, headers });

  if (!res.ok) {
    const body = await res.json().catch(() => ({}));

    // 401: clear auth state and redirect to login (unless in demo mode)
    if (res.status === 401) {
      const isDemo = sessionStorage.getItem('fireline_demo') === 'true';
      if (!isDemo) {
        localStorage.removeItem('access_token');
        localStorage.removeItem('refresh_token');
        localStorage.removeItem('org_id');
        localStorage.removeItem('user_id');
        localStorage.removeItem('role');
        window.location.href = '/login';
      }
    }

    throw new ApiError(res.status, body.error?.code || 'UNKNOWN', body.error?.message || res.statusText);
  }

  return res.json();
}

// Location
export interface Location {
  id: string;
  name: string;
  org_id: string;
}

export const locationApi = {
  getLocations() {
    return request<{ locations: Location[] }>('/locations');
  },
};

// Auth
export interface AuthTokens {
  access_token: string;
  refresh_token: string;
  org_id?: string;
  user_id?: string;
  role?: string;
}

export const authApi = {
  signup(data: { org_name: string; org_slug: string; email: string; password: string; display_name: string }) {
    return request<AuthTokens>('/auth/signup', { method: 'POST', body: JSON.stringify(data) });
  },
  login(data: { email: string; password: string }) {
    return request<AuthTokens & { mfa_required?: boolean }>('/auth/login', { method: 'POST', body: JSON.stringify(data) });
  },
  refresh(refresh_token: string) {
    return request<AuthTokens>('/auth/refresh', { method: 'POST', body: JSON.stringify({ refresh_token }) });
  },
  logout(refresh_token: string) {
    return request<{ status: string }>('/auth/logout', { method: 'POST', body: JSON.stringify({ refresh_token }) });
  },
};

// Financial
export interface PnL {
  location_id: string;
  period_start: string;
  period_end: string;
  gross_revenue: number;
  discounts: number;
  net_revenue: number;
  cogs: number;
  gross_profit: number;
  gross_margin: number;
  tax_collected: number;
  tips: number;
  check_count: number;
  avg_check_size: number;
  by_channel: ChannelBreakdown[];
}

export interface ChannelBreakdown {
  channel: string;
  revenue: number;
  cogs: number;
  gross_profit: number;
  gross_margin: number;
  check_count: number;
  avg_check_size: number;
}

export interface Anomaly {
  metric_name: string;
  current_value: number;
  mean: number;
  std_dev: number;
  z_score: number;
  severity: 'warning' | 'critical';
  detected_at: string;
}

export const financialApi = {
  getPnL(locationId: string, from?: string, to?: string) {
    const params = new URLSearchParams({ location_id: locationId });
    if (from) params.set('from', from);
    if (to) params.set('to', to);
    return request<PnL>(`/financial/pnl?${params}`);
  },
  getAnomalies(locationId: string) {
    return request<{ anomalies: Anomaly[] }>(`/financial/anomalies?location_id=${locationId}`);
  },
};

// Inventory
export interface TheoreticalUsage {
  ingredient_id: string;
  ingredient_name: string;
  total_used: number;
  unit: string;
  cost_per_unit: number;
  total_cost: number;
}

export interface PARStatus {
  ingredient_id: string;
  ingredient_name: string;
  current_level: number;
  par_level: number;
  reorder_point: number;
  unit: string;
  needs_reorder: boolean;
  suggested_qty: number;
}

export const inventoryApi = {
  getUsage(locationId: string, from?: string, to?: string) {
    const params = new URLSearchParams({ location_id: locationId });
    if (from) params.set('from', from);
    if (to) params.set('to', to);
    return request<{ usage: TheoreticalUsage[]; period_start: string; period_end: string }>(`/inventory/usage?${params}`);
  },
  getPARStatus(locationId: string) {
    return request<{ par_status: PARStatus[] }>(`/inventory/par?location_id=${locationId}`);
  },
};

// Alerting
export interface Alert {
  alert_id: string;
  org_id: string;
  location_id: string;
  rule_id: string;
  severity: 'info' | 'warning' | 'critical';
  title: string;
  description: string;
  module: string;
  status: string;
  created_at: string;
  acked_at: string | null;
  resolved_at: string | null;
}

export const alertsApi = {
  getQueue(locationId?: string) {
    const params = locationId ? `?location_id=${locationId}` : '';
    return request<{ alerts: Alert[] }>(`/alerts${params}`);
  },
  getCount() {
    return request<{ count: number }>('/alerts/count');
  },
  acknowledge(alertId: string) {
    return request<{ status: string }>(`/alerts/${alertId}/acknowledge`, { method: 'POST' });
  },
  resolve(alertId: string) {
    return request<{ status: string }>(`/alerts/${alertId}/resolve`, { method: 'POST' });
  },
};

// Menu Intelligence
export interface MenuItemAnalysis {
  menu_item_id: string;
  name: string;
  category: string;
  price: number;
  food_cost: number;
  units_sold: number;
  contrib_margin: number;
  contrib_margin_pct: number;
  popularity_pct: number;
  health_score: number;
  classification: 'powerhouse' | 'hidden_gem' | 'crowd_pleaser' | 'underperformer';
  by_channel: ChannelMarginData[];
}

export interface ChannelMarginData {
  channel: string;
  revenue: number;
  commission: number;
  food_cost: number;
  margin: number;
  margin_pct: number;
  units_sold: number;
}

export interface MenuSummary {
  total_items: number;
  avg_margin_pct: number;
  powerhouse_count: number;
  underperform_count: number;
  categories: CategorySummaryData[];
}

export interface CategorySummaryData {
  category: string;
  item_count: number;
  avg_margin_pct: number;
  top_item: string;
}

export const menuApi = {
  getItems(locationId: string, from?: string, to?: string) {
    const params = new URLSearchParams({ location_id: locationId });
    if (from) params.set('from', from);
    if (to) params.set('to', to);
    return request<{ items: MenuItemAnalysis[] }>(`/menu/items?${params}`);
  },
  getSummary(locationId: string, from?: string, to?: string) {
    const params = new URLSearchParams({ location_id: locationId });
    if (from) params.set('from', from);
    if (to) params.set('to', to);
    return request<MenuSummary>(`/menu/summary?${params}`);
  },
};

export { ApiError };
