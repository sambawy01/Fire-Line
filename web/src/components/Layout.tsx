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
} from 'lucide-react';
import { useAuthStore } from '../stores/auth';

const navItems = [
  { to: '/', label: 'Dashboard', icon: LayoutDashboard },
  { to: '/inventory', label: 'Inventory', icon: Package },
  { to: '/financial', label: 'Financial', icon: DollarSign },
  { to: '/alerts', label: 'Alerts', icon: Bell },
  { to: '/adapters', label: 'Adapters', icon: Plug },
];

export default function Layout() {
  const navigate = useNavigate();
  const logout = useAuthStore((s) => s.logout);
  const role = useAuthStore((s) => s.role);

  const handleLogout = () => {
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

        {/* Navigation */}
        <nav className="flex-1 px-3 py-4 space-y-1 overflow-y-auto">
          {navItems.map(({ to, label, icon: Icon }) => (
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
          <h2 className="text-xl font-semibold text-gray-800">FireLine</h2>
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
