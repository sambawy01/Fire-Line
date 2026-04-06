import { lazy, Suspense } from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ErrorBoundary } from './components/ErrorBoundary';
import { ProtectedRoute } from './components/ProtectedRoute';
import Layout from './components/Layout';

const LoginPage = lazy(() => import('./pages/LoginPage'));
const SignupPage = lazy(() => import('./pages/SignupPage'));
const OnboardingPage = lazy(() => import('./pages/OnboardingPage'));
const DashboardPage = lazy(() => import('./pages/DashboardPage'));
const InventoryPage = lazy(() => import('./pages/InventoryPage'));
const FinancialPage = lazy(() => import('./pages/FinancialPage'));
const AlertsPage = lazy(() => import('./pages/AlertsPage'));
const AdaptersPage = lazy(() => import('./pages/AdaptersPage'));
const MenuPage = lazy(() => import('./pages/MenuPage'));
const LaborPage = lazy(() => import('./pages/LaborPage'));
const VendorPage = lazy(() => import('./pages/VendorPage'));
const CustomerPage = lazy(() => import('./pages/CustomerPage'));
const OperationsPage = lazy(() => import('./pages/OperationsPage'));
const ReportsPage = lazy(() => import('./pages/ReportsPage'));
const PurchaseOrdersPage = lazy(() => import('./pages/PurchaseOrdersPage'));
const SchedulingPage = lazy(() => import('./pages/SchedulingPage'));
const KitchenPage = lazy(() => import('./pages/KitchenPage'));
const MaintenancePage = lazy(() => import('./pages/MaintenancePage'));
const MarketingPage = lazy(() => import('./pages/MarketingPage'));
const PayrollPage = lazy(() => import('./pages/PayrollPage'));
const IntelligencePage = lazy(() => import('./pages/IntelligencePage'));
const PortfolioPage = lazy(() => import('./pages/PortfolioPage'));

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { retry: 1, staleTime: 30_000 },
  },
});

export default function App() {
  return (
    <ErrorBoundary>
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <Suspense fallback={<div className="flex items-center justify-center h-screen"><div className="animate-spin rounded-full h-8 w-8 border-b-2 border-orange-500"></div></div>}>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/signup" element={<SignupPage />} />
          <Route path="/onboarding" element={<OnboardingPage />} />
          <Route
            path="/"
            element={
              <ProtectedRoute>
                <Layout />
              </ProtectedRoute>
            }
          >
            <Route index element={<PortfolioPage />} />
            <Route path="dashboard" element={<DashboardPage />} />
            <Route path="inventory" element={<InventoryPage />} />
            <Route path="purchase-orders" element={<PurchaseOrdersPage />} />
            <Route path="financial" element={<FinancialPage />} />
            <Route path="menu" element={<MenuPage />} />
            <Route path="labor" element={<LaborPage />} />
            <Route path="scheduling" element={<SchedulingPage />} />
            <Route path="vendors" element={<VendorPage />} />
            <Route path="customers" element={<CustomerPage />} />
            <Route path="operations" element={<OperationsPage />} />
            <Route path="kitchen" element={<KitchenPage />} />
            <Route path="maintenance" element={<MaintenancePage />} />
            <Route path="reports" element={<ReportsPage />} />
            <Route path="marketing" element={<MarketingPage />} />
            <Route path="payroll" element={<PayrollPage />} />
            <Route path="intelligence" element={<IntelligencePage />} />
            <Route path="alerts" element={<AlertsPage />} />
            <Route path="adapters" element={<AdaptersPage />} />
          </Route>
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
        </Suspense>
      </BrowserRouter>
    </QueryClientProvider>
    </ErrorBoundary>
  );
}
