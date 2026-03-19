import { Navigate, useSearchParams } from 'react-router-dom';
import { useAuthStore } from '../stores/auth';

export function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const [params] = useSearchParams();

  // Demo mode: skip auth for UI testing (?demo=true on any page)
  const demoMode = params.get('demo') === 'true' || sessionStorage.getItem('fireline_demo') === 'true';
  if (demoMode) {
    sessionStorage.setItem('fireline_demo', 'true');
    return <>{children}</>;
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  return <>{children}</>;
}
