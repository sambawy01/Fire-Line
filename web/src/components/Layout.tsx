import { useEffect } from 'react';
import { Outlet, NavLink, useNavigate } from 'react-router-dom';
import {
  LayoutDashboard,
  Package,
  DollarSign,
  Bell,
  Plug,
  LogOut,
  Flame,
  User,
  MapPin,
  UtensilsCrossed,
  Users,
  Truck,
  UserCheck,
  Activity,
} from 'lucide-react';
import { useAuthStore } from '../stores/auth';
import { useLocationStore } from '../stores/location';
import { useAlertCount } from '../hooks/useAlerts';

const navItems = [
  { to: '/', label: 'Dashboard', icon: LayoutDashboard },
  { to: '/inventory', label: 'Inventory', icon: Package },
  { to: '/financial', label: 'Financial', icon: DollarSign },
  { to: '/menu', label: 'Menu', icon: UtensilsCrossed },
  { to: '/labor', label: 'Labor', icon: Users },
  { to: '/vendors', label: 'Vendors', icon: Truck },
  { to: '/customers', label: 'Customers', icon: UserCheck },
  { to: '/operations', label: 'Operations', icon: Activity },
  { to: '/alerts', label: 'Alerts', icon: Bell, showBadge: true },
  { to: '/adapters', label: 'Adapters', icon: Plug },
];

export default function Layout() {
  const navigate = useNavigate();
  const logout = useAuthStore((s) => s.logout);
  const role = useAuthStore((s) => s.role);
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);

  const { locations, selectedLocationId, setLocation, loadLocations } = useLocationStore();
  const { data: alertCount } = useAlertCount(selectedLocationId);

  useEffect(() => {
    if (isAuthenticated) {
      loadLocations();
    }
  }, [isAuthenticated, loadLocations]);

  const handleLogout = () => {
    useLocationStore.getState().clear();
    logout();
    navigate('/login');
  };

  return (
    <div className="flex h-screen bg-gray-100">
      {/* Sidebar */}
      <aside className="hidden md:flex md:flex-col md:w-64 md:fixed md:inset-y-0 bg-[#1E293B] text-white">
        {/* Logo */}
        <div className="flex items-center gap-3 px-6 py-5 border-b border-white/10">
          <Flame className="h-8 w-8 text-[#F97316]" />
          <div>
            <h1 className="text-lg font-bold tracking-tight">FireLine</h1>
            <p className="text-xs text-gray-400">by OpsNerve</p>
          </div>
        </div>

        {/* Location Switcher */}
        {locations.length > 1 && (
          <div className="px-3 py-3 border-b border-white/10">
            <div className="flex items-center gap-2 px-3 mb-1.5">
              <MapPin className="h-3.5 w-3.5 text-gray-400" />
              <span className="text-xs text-gray-400 uppercase tracking-wider">Location</span>
            </div>
            <select
              value={selectedLocationId ?? ''}
              onChange={(e) => setLocation(e.target.value)}
              className="w-full bg-white/10 text-white text-sm rounded-md px-3 py-1.5 border border-white/10 focus:outline-none focus:ring-1 focus:ring-[#F97316]"
            >
              {locations.map((loc) => (
                <option key={loc.id} value={loc.id} className="bg-[#1E293B]">
                  {loc.name}
                </option>
              ))}
            </select>
          </div>
        )}

        {/* Navigation */}
        <nav className="flex-1 px-3 py-4 space-y-1 overflow-y-auto">
          {navItems.map(({ to, label, icon: Icon, showBadge }) => (
            <NavLink
              key={to}
              to={to}
              className={({ isActive }) =>
                `flex items-center gap-3 px-3 py-2.5 rounded-md text-sm font-medium transition-colors ${
                  isActive
                    ? 'border-l-[3px] border-[#F97316] text-[#F97316] bg-white/5'
                    : 'border-l-[3px] border-transparent text-gray-300 hover:text-white hover:bg-white/5'
                }`
              }
            >
              <Icon className="h-5 w-5 shrink-0" />
              {label}
              {showBadge && alertCount?.count != null && alertCount.count > 0 && (
                <span className="ml-auto inline-flex items-center justify-center rounded-full bg-[#F97316] px-2 py-0.5 text-xs font-bold text-white min-w-[20px]">
                  {alertCount.count}
                </span>
              )}
            </NavLink>
          ))}
        </nav>

        {/* Logout */}
        <div className="px-3 py-4 border-t border-white/10">
          <button
            onClick={handleLogout}
            className="flex items-center gap-3 w-full px-3 py-2.5 rounded-md text-sm font-medium text-gray-300 hover:text-white hover:bg-white/5 transition-colors"
          >
            <LogOut className="h-5 w-5 shrink-0" />
            Logout
          </button>
        </div>
      </aside>

      {/* Main content */}
      <div className="flex-1 md:ml-64 flex flex-col min-h-screen">
        {/* Top header */}
        <header className="sticky top-0 z-10 bg-white border-b border-gray-200 px-6 py-4 flex items-center justify-between">
          <h2 className="text-xl font-semibold text-gray-800">
            {locations.length === 1 ? locations[0].name : 'FireLine'}
          </h2>
          <div className="flex items-center gap-3">
            <div className="text-right hidden sm:block">
              <p className="text-sm font-medium text-gray-700">
                {role ?? 'Operator'}
              </p>
              <p className="text-xs text-gray-400">Restaurant Manager</p>
            </div>
            <div className="h-9 w-9 rounded-full bg-[#1E293B] flex items-center justify-center">
              <User className="h-5 w-5 text-white" />
            </div>
          </div>
        </header>

        {/* Page content */}
        <main className="flex-1 p-6 overflow-y-auto">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
