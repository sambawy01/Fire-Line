import { Outlet, NavLink, useNavigate } from 'react-router-dom';
import { Home, ClipboardList, Calendar, Trophy, LogOut } from 'lucide-react';
import { getUser, logout } from '../stores/auth';

const tabs = [
  { to: '/', label: 'Home', icon: Home },
  { to: '/tasks', label: 'Tasks', icon: ClipboardList },
  { to: '/schedule', label: 'Schedule', icon: Calendar },
  { to: '/points', label: 'Points', icon: Trophy },
] as const;

export default function Layout() {
  const user = getUser();
  const navigate = useNavigate();

  function handleLogout() {
    logout();
    navigate('/login', { replace: true });
  }

  return (
    <div className="flex flex-col min-h-screen bg-slate-900 text-slate-100">
      {/* Header */}
      <header className="flex items-center justify-between px-4 py-3 bg-slate-800 border-b border-slate-700">
        <div className="flex items-center gap-3">
          <div>
            <p className="text-sm font-semibold text-white leading-tight">
              {user?.display_name ?? 'Staff'}
            </p>
            <span className="inline-block mt-0.5 px-2 py-0.5 text-[10px] font-medium uppercase tracking-wider rounded-full bg-orange-500/20 text-orange-400">
              {user?.role ?? 'crew'}
            </span>
          </div>
        </div>
        <button
          onClick={handleLogout}
          className="flex items-center gap-1.5 text-xs text-slate-400 hover:text-white transition-colors"
          aria-label="Log out"
        >
          <LogOut size={16} />
          <span className="hidden sm:inline">Logout</span>
        </button>
      </header>

      {/* Page content */}
      <main className="flex-1 overflow-y-auto pb-20">
        <Outlet />
      </main>

      {/* Bottom tab bar */}
      <nav
        className="fixed bottom-0 inset-x-0 bg-slate-800 border-t border-slate-700 safe-area-pb"
        aria-label="Main navigation"
      >
        <ul className="flex justify-around items-center h-16">
          {tabs.map(({ to, label, icon: Icon }) => (
            <li key={to}>
              <NavLink
                to={to}
                end={to === '/'}
                className={({ isActive }) =>
                  `flex flex-col items-center gap-0.5 px-3 py-1 text-[11px] font-medium transition-colors ${
                    isActive
                      ? 'text-orange-400'
                      : 'text-slate-400 hover:text-slate-200'
                  }`
                }
              >
                <Icon size={22} />
                {label}
              </NavLink>
            </li>
          ))}
        </ul>
      </nav>
    </div>
  );
}
