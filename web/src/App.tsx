import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ProtectedRoute } from './components/ProtectedRoute';
import Layout from './components/Layout';
import LoginPage from './pages/LoginPage';
import SignupPage from './pages/SignupPage';
import DashboardPage from './pages/DashboardPage';
import InventoryPage from './pages/InventoryPage';
import FinancialPage from './pages/FinancialPage';
import AlertsPage from './pages/AlertsPage';
import AdaptersPage from './pages/AdaptersPage';
import MenuPage from './pages/MenuPage';
import LaborPage from './pages/LaborPage';
import VendorPage from './pages/VendorPage';
import CustomerPage from './pages/CustomerPage';
import OperationsPage from './pages/OperationsPage';
import ReportsPage from './pages/ReportsPage';
import PurchaseOrdersPage from './pages/PurchaseOrdersPage';
import SchedulingPage from './pages/SchedulingPage';
import KitchenPage from './pages/KitchenPage';
import MarketingPage from './pages/MarketingPage';
import PortfolioPage from './pages/PortfolioPage';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { retry: 1, staleTime: 30_000 },
  },
});

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/signup" element={<SignupPage />} />
          <Route
            path="/"
            element={
              <ProtectedRoute>
                <Layout />
              </ProtectedRoute>
            }
          >
            <Route index element={<DashboardPage />} />
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
            <Route path="reports" element={<ReportsPage />} />
            <Route path="marketing" element={<MarketingPage />} />
            <Route path="portfolio" element={<PortfolioPage />} />
            <Route path="alerts" element={<AlertsPage />} />
            <Route path="adapters" element={<AdaptersPage />} />
          </Route>
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </BrowserRouter>
    </QueryClientProvider>
  );
}
