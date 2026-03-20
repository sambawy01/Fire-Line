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

export interface Budget {
  budget_id: string;
  period_type: 'daily' | 'weekly' | 'monthly';
  period_start: string;
  period_end: string;
  revenue_target: number;
  food_cost_pct_target: number;
  labor_cost_pct_target: number;
  cogs_target: number;
}

export interface BudgetVariance {
  budget: Budget;
  actual_revenue: number;
  actual_cogs: number;
  actual_food_cost_pct: number;
  revenue_variance: number;
  revenue_variance_pct: number;
  cogs_variance: number;
  cogs_variance_pct: number;
  food_cost_pct_delta: number;
  status: 'on_track' | 'over' | 'under';
}

export interface IngredientCostEntry {
  ingredient_id: string;
  ingredient_name: string;
  total_cost: number;
  unit_cost: number;
  quantity_used: number;
  unit: string;
  cost_pct: number;
}

export interface CostCenter {
  category: string;
  cogs: number;
  cogs_pct: number;
  revenue_pct: number;
  ingredient_count: number;
  top_ingredients: IngredientCostEntry[];
}

export interface TransactionAnomaly {
  type: string;
  description: string;
  current_value: number;
  baseline: number;
  z_score: number;
  severity: string;
  detected_at: string;
}

export type ProfitAndLoss = PnL;

export interface PeriodComparison {
  current: ProfitAndLoss;
  last_week: ProfitAndLoss | null;
  last_month: ProfitAndLoss | null;
  revenue_vs_last_week_pct: number;
  revenue_vs_last_month_pct: number;
  cogs_vs_last_week_pct: number;
  cogs_vs_last_month_pct: number;
}

export interface ItemCost {
  menu_item_id: string;
  name: string;
  category: string;
  revenue: number;
  cogs: number;
  gross_profit: number;
  gross_margin: number;
  units_sold: number;
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
  budgetVariance: (locationId: string) =>
    request<BudgetVariance>(`/financial/budget-variance?location_id=${locationId}`),
  costCenters: (locationId: string) =>
    request<{ cost_centers: CostCenter[] }>(`/financial/cost-centers?location_id=${locationId}`),
  txAnomalies: (locationId: string) =>
    request<{ anomalies: TransactionAnomaly[] }>(`/financial/transaction-anomalies?location_id=${locationId}`),
  periodComparison: (locationId: string) =>
    request<PeriodComparison>(`/financial/period-comparison?location_id=${locationId}`),
  drilldownItems: (locationId: string, category: string) =>
    request<{ items: ItemCost[] }>(`/financial/drilldown/items?location_id=${locationId}&category=${category}`),
  drilldownIngredients: (locationId: string, menuItemId: string) =>
    request<{ ingredients: IngredientCostEntry[] }>(`/financial/drilldown/ingredients?location_id=${locationId}&menu_item_id=${menuItemId}`),
  createBudget: (data: any) =>
    request<Budget>('/financial/budgets', { method: 'POST', body: JSON.stringify(data) }),
  listBudgets: (locationId: string, periodType?: string) =>
    request<{ budgets: Budget[] }>(`/financial/budgets?location_id=${locationId}${periodType ? `&period_type=${periodType}` : ''}`),
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

export interface CountVariance {
  variance_id: string;
  ingredient_id: string;
  ingredient_name: string;
  category: string;
  theoretical_usage: number;
  actual_usage: number;
  variance_qty: number;
  variance_pct: number;
  variance_cents: number;
  cause_probabilities: Record<string, number>;
  severity: 'info' | 'warning' | 'critical';
  created_at: string;
}

export interface WasteLogEntry {
  waste_id: string;
  ingredient_id: string;
  ingredient_name: string;
  quantity: number;
  unit: string;
  reason: string;
  logged_by: string;
  logged_by_name: string;
  logged_at: string;
  note: string;
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
  getVariances: (locationId: string) =>
    request<{ variances: CountVariance[] }>(`/inventory/variances?location_id=${locationId}`),
  getWasteLogs: (locationId: string) =>
    request<{ waste_logs: WasteLogEntry[] }>(
      `/inventory/waste?location_id=${locationId}&from=${new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString()}&to=${new Date().toISOString()}`
    ),
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

// Menu Scoring (SP16)
export interface MenuItemScore {
  menu_item_id: string;
  name: string;
  category: string;
  price: number;
  margin_score: number;
  velocity_score: number;
  complexity_score: number;
  satisfaction_score: number;
  strategic_score: number;
  classification:
    | 'powerhouse'
    | 'hidden_gem'
    | 'crowd_pleaser'
    | 'workhorse'
    | 'complex_star'
    | 'declining_star'
    | 'underperformer'
    | 'strategic_anchor';
  contribution_margin: number;
  units_sold: number;
}

export interface AffectedItem {
  menu_item_id: string;
  name: string;
  shared: boolean;
  margin_delta?: number;
}

export interface SimulationResult {
  simulation_type: string;
  current_revenue: number;
  projected_revenue: number;
  revenue_delta: number;
  current_profit: number;
  projected_profit: number;
  profit_delta: number;
  affected_items?: AffectedItem[];
}

export interface IngredientDependency {
  ingredient_id: string;
  ingredient_name: string;
  menu_item_count: number;
  menu_items: string[];
}

export interface CrossSellPair {
  item_a_name: string;
  item_b_name: string;
  co_occurrences: number;
  affinity: number;
}

export const menuScoringApi = {
  triggerScore(locationId: string) {
    return request<{ status: string }>('/menu/score', {
      method: 'POST',
      body: JSON.stringify({ location_id: locationId }),
    });
  },
  getScores(locationId: string) {
    return request<{ items: MenuItemScore[] }>(`/menu/scores?location_id=${locationId}`);
  },
  simulatePrice(locationId: string, menuItemId: string, newPrice: number) {
    return request<SimulationResult>('/menu/simulate/price', {
      method: 'POST',
      body: JSON.stringify({ location_id: locationId, menu_item_id: menuItemId, new_price: newPrice }),
    });
  },
  simulateRemoval(locationId: string, menuItemId: string) {
    return request<SimulationResult>('/menu/simulate/removal', {
      method: 'POST',
      body: JSON.stringify({ location_id: locationId, menu_item_id: menuItemId }),
    });
  },
  simulateIngredientCost(locationId: string, ingredientId: string, newCostPerUnit: number) {
    return request<SimulationResult>('/menu/simulate/ingredient-cost', {
      method: 'POST',
      body: JSON.stringify({ location_id: locationId, ingredient_id: ingredientId, new_cost_per_unit: newCostPerUnit }),
    });
  },
  getDependencies(locationId: string) {
    return request<{ dependencies: IngredientDependency[] }>(`/menu/dependencies?location_id=${locationId}`);
  },
  getCrossSell(locationId: string, limit = 10) {
    return request<{ pairs: CrossSellPair[] }>(`/menu/cross-sell?location_id=${locationId}&limit=${limit}`);
  },
};

// Labor Profiles / ELU / Points
export interface EmployeeProfile {
  employee_id: string;
  display_name: string;
  role: string;
  status: string;
  elu_ratings: Record<string, number>;
  staff_points: number;
  points_trend: 'up' | 'down' | 'stable';
  certifications: string[];
  availability: Record<string, any>;
}

export interface PointEvent {
  event_id: string;
  employee_id: string;
  points: number;
  reason: string;
  description: string;
  created_at: string;
}

export interface LeaderboardEntry {
  employee_id: string;
  display_name: string;
  role: string;
  staff_points: number;
  points_trend: 'up' | 'down' | 'stable';
}

// Labor Intelligence
export interface LaborSummary {
  total_labor_cost: number;
  labor_cost_pct: number;
  net_revenue: number;
  employee_count: number;
  total_hours: number;
  total_shifts: number;
}

export interface EmployeeDetail {
  employee_id: string;
  display_name: string;
  role: string;
  status: string;
  shift_count: number;
  hours_worked: number;
  labor_cost: number;
  avg_hours_per_shift: number;
  hourly_rate: number;
}

export const laborApi = {
  getSummary(locationId: string, from?: string, to?: string) {
    const params = new URLSearchParams({ location_id: locationId });
    if (from) params.set('from', from);
    if (to) params.set('to', to);
    return request<LaborSummary>(`/labor/summary?${params}`);
  },
  getEmployees(locationId: string, from?: string, to?: string) {
    const params = new URLSearchParams({ location_id: locationId });
    if (from) params.set('from', from);
    if (to) params.set('to', to);
    return request<{ employees: EmployeeDetail[] }>(`/labor/employees?${params}`);
  },
  getProfiles: (locationId: string) =>
    request<{ profiles: EmployeeProfile[] }>(`/labor/profiles?location_id=${locationId}`),
  getProfile: (id: string) =>
    request<EmployeeProfile>(`/labor/profiles/${id}`),
  updateELU: (id: string, ratings: Record<string, number>) =>
    request<any>(`/labor/profiles/${id}/elu`, { method: 'PUT', body: JSON.stringify({ ratings }) }),
  awardPoints: (data: any) =>
    request<any>('/labor/points', { method: 'POST', body: JSON.stringify(data) }),
  getPointHistory: (employeeId: string) =>
    request<{ events: PointEvent[] }>(`/labor/points/${employeeId}?limit=20`),
  getLeaderboard: (locationId: string) =>
    request<{ leaderboard: LeaderboardEntry[] }>(`/labor/leaderboard?location_id=${locationId}&limit=10`),
};

// Vendor Intelligence
export interface VendorAnalysis {
  vendor_name: string;
  items_supplied: number;
  total_spend: number;
  spend_pct: number;
  avg_cost_per_item: number;
  score: number;
}

export interface VendorSummary {
  total_vendors: number;
  total_spend: number;
  top_vendor_name: string;
  top_vendor_pct: number;
  avg_items_per_vendor: number;
}

export const vendorApi = {
  getVendors(locationId: string, from?: string, to?: string) {
    const params = new URLSearchParams({ location_id: locationId });
    if (from) params.set('from', from);
    if (to) params.set('to', to);
    return request<{ vendors: VendorAnalysis[] }>(`/vendors?${params}`);
  },
  getSummary(locationId: string, from?: string, to?: string) {
    const params = new URLSearchParams({ location_id: locationId });
    if (from) params.set('from', from);
    if (to) params.set('to', to);
    return request<VendorSummary>(`/vendors/summary?${params}`);
  },
};

// Vendor Scoring (SP17)
export interface VendorScore {
  vendor_name: string;
  overall_score: number;
  price_score: number;
  delivery_score: number;
  quality_score: number;
  accuracy_score: number;
  total_orders: number;
  otif_rate: number;
  on_time_rate: number;
  in_full_rate: number;
  avg_lead_days: number;
}

export interface PriceAnomaly {
  ingredient_name: string;
  vendor_name: string;
  current_price: number;
  avg_price: number;
  z_score: number;
  severity: 'warning' | 'critical';
}

export interface VendorRecommendation {
  vendor_name: string;
  score: number;
  unit_cost: number;
  reasoning: string;
}

export interface PricePoint {
  unit_cost: number;
  recorded_at: string;
}

export interface VendorCompareEntry {
  vendor_name: string;
  overall_score: number;
  unit_cost: number;
  otif_rate: number;
  avg_lead_days: number;
}

export const vendorScoringApi = {
  calculateScores(locationId: string) {
    return request<{ status: string }>('/vendors/scores/calculate', {
      method: 'POST',
      body: JSON.stringify({ location_id: locationId }),
    });
  },
  getScores(locationId: string) {
    return request<{ vendor_scores: VendorScore[] }>(`/vendors/scores?location_id=${locationId}`);
  },
  getScorecard(locationId: string, vendorName: string) {
    return request<VendorScore>(`/vendors/scorecard?location_id=${locationId}&vendor_name=${encodeURIComponent(vendorName)}`);
  },
  compare(locationId: string, ingredientId: string) {
    return request<{ ingredient_name: string; vendors: VendorCompareEntry[]; recommended: string }>(
      `/vendors/compare?location_id=${locationId}&ingredient_id=${ingredientId}`
    );
  },
  priceTrend(ingredientId: string, vendorName: string, months = 6) {
    return request<{ prices: PricePoint[] }>(
      `/vendors/price-trend?ingredient_id=${ingredientId}&vendor_name=${encodeURIComponent(vendorName)}&months=${months}`
    );
  },
  priceAnomalies(locationId: string) {
    return request<{ anomalies: PriceAnomaly[] }>(`/vendors/price-anomalies?location_id=${locationId}`);
  },
  recommend(locationId: string, ingredientId: string) {
    return request<VendorRecommendation>(`/vendors/recommend?location_id=${locationId}&ingredient_id=${ingredientId}`);
  },
};

// Guest Profiles (SP15)
export interface GuestProfile {
  guest_id: string;
  privacy_tier: string;
  first_name: string | null;
  total_visits: number;
  total_spend: number;
  avg_check: number;
  clv_score: number;
  segment: string;
  churn_risk: string;
  churn_probability: number;
  last_visit_at: string | null;
}

export interface SegmentDistribution {
  segment: string;
  count: number;
}

export interface ChurnDistribution {
  risk: string;
  count: number;
}

export interface CLVBucket {
  range: string;
  count: number;
}

export const guestApi = {
  list: (locationId: string, sortBy?: string) =>
    request<{ guests: GuestProfile[] }>(`/customers/guests?location_id=${locationId}&sort_by=${sortBy || 'clv_score'}&limit=50`),
  get: (id: string) =>
    request<GuestProfile>(`/customers/guests/${id}`),
  segments: () =>
    request<{ segments: SegmentDistribution[] }>('/customers/analytics/segments'),
  churn: () =>
    request<{ distribution: ChurnDistribution[] }>('/customers/analytics/churn'),
  clv: () =>
    request<{ buckets: CLVBucket[] }>('/customers/analytics/clv'),
  refresh: () =>
    request<any>('/customers/analytics/refresh', { method: 'POST' }),
};

// Customer Intelligence (legacy)
export interface CustomerDetail {
  customer_id: string;
  name: string;
  email: string;
  phone: string;
  first_visit: string | null;
  last_visit: string | null;
  total_visits: number;
  total_spend: number;
  avg_check: number;
  segment: 'new' | 'regular' | 'vip' | 'lapsed' | 'at_risk';
  ai_summary: string;
  ai_summary_updated_at: string | null;
}

export interface CustomerSummary {
  total_customers: number;
  avg_lifetime_value: number;
  vip_count: number;
  at_risk_count: number;
  segment_counts: Record<string, number>;
}

export interface AnalyzeResult {
  analyzed: number;
  errors: number;
  message: string;
}

export const customerApi = {
  getCustomers(locationId: string) {
    return request<{ customers: CustomerDetail[] }>(`/customers?location_id=${locationId}`);
  },
  getSummary(locationId: string) {
    return request<CustomerSummary>(`/customers/summary?location_id=${locationId}`);
  },
  analyze(locationId: string) {
    return request<AnalyzeResult>(`/customers/analyze?location_id=${locationId}`, { method: 'POST' });
  },
};

// Operations Intelligence
export interface OperationsSummary {
  orders_today: number;
  avg_ticket_time: number;
  orders_per_hour: number;
  active_tickets: number;
  longest_open_min: number;
  revenue_per_hour: number;
  void_rate: number;
  channel_performance: ChannelPerf[];
}

export interface ChannelPerf {
  channel: string;
  orders: number;
  pct_of_total: number;
  avg_ticket_time: number;
  revenue: number;
}

export interface HourlyData {
  hour: number;
  orders: number;
  revenue: number;
}

export const operationsApi = {
  getSummary(locationId: string, from?: string, to?: string) {
    const params = new URLSearchParams({ location_id: locationId });
    if (from) params.set('from', from);
    if (to) params.set('to', to);
    return request<OperationsSummary>(`/operations/summary?${params}`);
  },
  getHourly(locationId: string, from?: string, to?: string) {
    const params = new URLSearchParams({ location_id: locationId });
    if (from) params.set('from', from);
    if (to) params.set('to', to);
    return request<{ hourly: HourlyData[] }>(`/operations/hourly?${params}`);
  },
};

// Reporting
export interface DailyReport {
  location_id: string;
  location_name: string;
  report_date: string;
  health_score: number;
  net_revenue: number;
  gross_margin_pct: number;
  labor_cost_pct: number;
  orders_today: number;
  avg_ticket_time: number;
  active_alerts: number;
  critical_count: number;
  critical_issues: CriticalIssue[];
  channels: ReportChannel[];
  top_items: ReportMenuItem[];
  worst_item: ReportMenuItem | null;
  zero_sales_items: string[];
  category_revenue: CategoryRevData[];
  staff_summary: StaffEntry[];
  total_labor_cost: number;
  total_hours_worked: number;
  overtime_flags: string[];
  reorder_needed: ReorderItem[];
}

export interface CriticalIssue { title: string; module: string; created_at: string; }
export interface ReportChannel { channel: string; orders: number; revenue: number; pct_of_total: number; avg_ticket_time: number; }
export interface ReportMenuItem { name: string; category: string; units_sold: number; revenue: number; margin_pct: number; }
export interface CategoryRevData { category: string; revenue: number; pct_of_total: number; item_count: number; }
export interface StaffEntry { name: string; role: string; hours_worked: number; labor_cost: number; is_overtime: boolean; }
export interface ReorderItem { name: string; current_level: number; par_level: number; unit: string; }

export const reportsApi = {
  getDaily(locationId: string, from?: string, to?: string) {
    const params = new URLSearchParams({ location_id: locationId });
    if (from) params.set('from', from);
    if (to) params.set('to', to);
    return request<DailyReport>(`/reports/daily?${params}`);
  },
  async downloadPdf(locationId: string) {
    const token = localStorage.getItem('access_token');
    const res = await fetch(`/api/v1/reports/daily/pdf?location_id=${locationId}`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    if (!res.ok) throw new Error('PDF download failed');
    const blob = await res.blob();
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `fireline-daily-report.pdf`;
    a.click();
    URL.revokeObjectURL(url);
  },
};

// Purchase Orders
export interface PurchaseOrder {
  purchase_order_id: string;
  vendor_name: string;
  status: 'draft' | 'approved' | 'received' | 'cancelled';
  source: 'manual' | 'system_recommended';
  line_count: number;
  total_estimated: number; // cents
  total_actual: number;
  suggested_at: string | null;
  approved_at: string | null;
  received_at: string | null;
  notes: string;
}

export interface POLine {
  po_line_id: string;
  ingredient_id: string;
  ingredient_name: string;
  ordered_qty: number;
  ordered_unit: string;
  estimated_unit_cost: number;
  received_qty: number | null;
  received_unit_cost: number | null;
  variance_qty: number | null;
  variance_flag: string | null;
  note: string;
}

export interface POWithLines extends PurchaseOrder {
  lines: POLine[];
}

export interface PARBreach {
  ingredient_id: string;
  ingredient_name: string;
  current_level: number;
  reorder_point: number;
  par_level: number;
  avg_daily_usage: number;
  projected_stockout_days: number;
  vendor_name: string;
  has_pending_po: boolean;
}

export const poApi = {
  list: (locationId: string, status?: string) =>
    request<{ purchase_orders: PurchaseOrder[] }>(
      `/inventory/po?location_id=${locationId}${status ? `&status=${status}` : ''}`
    ),
  get: (id: string) => request<POWithLines>(`/inventory/po/${id}`),
  approve: (id: string) =>
    request<any>(`/inventory/po/${id}`, { method: 'PUT', body: JSON.stringify({ status: 'approved' }) }),
  cancel: (id: string) =>
    request<any>(`/inventory/po/${id}`, { method: 'PUT', body: JSON.stringify({ status: 'cancelled' }) }),
  parBreaches: (locationId: string) =>
    request<{ breaches: PARBreach[] }>(`/inventory/par-breaches?location_id=${locationId}`),
};

// Scheduling
export interface ScheduledShift {
  scheduled_shift_id: string;
  employee_id: string;
  employee_name: string;
  shift_date: string;
  start_time: string;
  end_time: string;
  station: string;
  status: string;
}

export interface ScheduleWithShifts {
  schedule_id: string;
  location_id: string;
  week_start: string;
  status: string;
  shifts: ScheduledShift[];
}

export interface LaborCostProjection {
  total_hours: number;
  total_cost: number;
  labor_cost_pct: number;
  budget_target_pct: number;
  over_under: string;
}

export interface ForecastBlock {
  time_block: string;
  forecasted_covers: number;
  required_headcount: number;
}

export interface OvertimeRisk {
  employee_id: string;
  employee_name: string;
  scheduled_hours: number;
  severity: string;
}

export interface SwapRequest {
  swap_id: string;
  requester_name: string;
  target_name: string;
  status: string;
  reason: string;
  created_at: string;
}

export const schedulingApi = {
  getSchedule: (locationId: string, weekStart: string) =>
    request<ScheduleWithShifts>(`/labor/schedules?location_id=${locationId}&week_start=${weekStart}`),
  generate: (locationId: string, weekStart: string) =>
    request<ScheduleWithShifts>('/labor/schedules/generate', { method: 'POST', body: JSON.stringify({ location_id: locationId, week_start: weekStart }) }),
  publish: (id: string) => request<any>(`/labor/schedules/${id}/publish`, { method: 'POST' }),
  cost: (id: string) => request<LaborCostProjection>(`/labor/schedules/${id}/cost`),
  forecast: (locationId: string, date: string) => request<{ forecast: ForecastBlock[] }>(`/labor/forecast?location_id=${locationId}&date=${date}`),
  swaps: (locationId: string) => request<{ swap_requests: SwapRequest[] }>(`/labor/swaps?location_id=${locationId}&status=pending`),
  reviewSwap: (id: string, approved: boolean) => request<any>(`/labor/swaps/${id}`, { method: 'PUT', body: JSON.stringify({ approved }) }),
  overtimeRisk: (locationId: string, weekStart: string) => request<{ risks: OvertimeRisk[] }>(`/labor/overtime-risk?location_id=${locationId}&week_start=${weekStart}`),
};

// Kitchen Operations
export interface KitchenStation {
  station_id: string;
  name: string;
  station_type: string;
  max_concurrent: number;
  current_load: number;
  load_pct: number;
  status: string;
}

export interface KitchenCapacity {
  stations: KitchenStation[];
  total_capacity_pct: number;
  active_tickets: number;
  avg_ticket_time_secs: number;
}

export interface KDSTicketItem {
  ticket_item_id: string;
  item_name: string;
  quantity: number;
  station_type: string;
  status: string;
}

export interface KDSTicket {
  ticket_id: string;
  order_number: string;
  channel: string;
  status: string;
  items: KDSTicketItem[];
  elapsed_secs: number;
  created_at: string;
}

export interface KDSMetricsByStation {
  station_type: string;
  avg_time_secs: number;
  items_completed: number;
}

export interface KDSMetrics {
  avg_ticket_time_secs: number;
  items_per_hour: number;
  tickets_completed: number;
  by_station: KDSMetricsByStation[];
}

export const kitchenApi = {
  capacity: (locationId: string) =>
    request<KitchenCapacity>(`/operations/capacity?location_id=${locationId}`),
  tickets: (locationId: string) =>
    request<{ tickets: KDSTicket[] }>(`/operations/kds/tickets?location_id=${locationId}`),
  stationTickets: (stationType: string, locationId: string) =>
    request<{ tickets: KDSTicket[] }>(`/operations/kds/station/${stationType}?location_id=${locationId}`),
  metrics: (locationId: string) =>
    request<KDSMetrics>(`/operations/kds/metrics?location_id=${locationId}`),
  bumpItem: (itemId: string, status: string) =>
    request<any>(`/operations/kds/items/${itemId}/bump`, {
      method: 'PUT',
      body: JSON.stringify({ status }),
    }),
};

// Operations Command Center (SP18)
export interface OperationalHealth {
  overall_score: number;
  kitchen_score: number;
  ticket_score: number;
  staff_score: number;
  financial_score: number;
  inventory_score: number;
  status: string;
}

export interface SuggestedAction {
  action_type: string;
  description: string;
  impact: string;
}

export interface OverloadStatus {
  is_overloaded: boolean;
  capacity_pct: number;
  severity: string;
  suggested_actions: SuggestedAction[];
}

export interface TicketPriority {
  ticket_id: string;
  order_number: string;
  channel: string;
  priority_score: number;
  sla_minutes: number;
  elapsed_minutes: number;
  urgency: string;
}

export interface RealtimeHorizon {
  health: OperationalHealth;
  overload: OverloadStatus;
  active_tickets: number;
  avg_ticket_time: number;
  station_loads: Record<string, number>;
}

export interface ShiftHorizon {
  forecasted_covers: number;
  scheduled_staff: number;
  required_staff: number;
  staff_gap: number;
  expected_revenue: number;
}

export interface DailyHorizon {
  prep_items: number;
  expected_deliveries: number;
  scheduled_shifts: number;
  forecasted_revenue: number;
}

export interface WeeklyHorizon {
  total_hours: number;
  pending_pos: number;
  projected_labor_cost: number;
  projected_revenue: number;
}

export interface StrategicHorizon {
  revenue_30d: number;
  revenue_delta_pct: number;
  cogs_30d: number;
  labor_cost_pct: number;
  labor_trend: string;
}

export const opsCommandApi = {
  getHealth(locationId: string) {
    return request<OperationalHealth>(`/operations/health?location_id=${locationId}`);
  },
  getOverload(locationId: string) {
    return request<OverloadStatus>(`/operations/overload?location_id=${locationId}`);
  },
  getPriorities(locationId: string) {
    return request<{ priorities: TicketPriority[] }>(`/operations/priority?location_id=${locationId}`);
  },
  getRealtimeHorizon(locationId: string) {
    return request<RealtimeHorizon>(`/operations/horizon/realtime?location_id=${locationId}`);
  },
  getShiftHorizon(locationId: string) {
    return request<ShiftHorizon>(`/operations/horizon/shift?location_id=${locationId}`);
  },
  getDailyHorizon(locationId: string) {
    return request<DailyHorizon>(`/operations/horizon/daily?location_id=${locationId}`);
  },
  getWeeklyHorizon(locationId: string) {
    return request<WeeklyHorizon>(`/operations/horizon/weekly?location_id=${locationId}`);
  },
  getStrategicHorizon(locationId: string) {
    return request<StrategicHorizon>(`/operations/horizon/strategic?location_id=${locationId}`);
  },
};

export { ApiError };
