interface StaffUser {
  user_id: string;
  org_id: string;
  role: string;
  display_name: string;
  staff_points: number;
  location_id: string;
}

let listeners: Array<() => void> = [];

function notify() { listeners.forEach(fn => fn()); }

export function getUser(): StaffUser | null {
  const raw = localStorage.getItem('staff_user');
  return raw ? JSON.parse(raw) : null;
}

export function getToken(): string | null {
  return localStorage.getItem('staff_token');
}

export function isAuthenticated(): boolean {
  return !!getToken();
}

export function login(user: StaffUser, token: string) {
  localStorage.setItem('staff_user', JSON.stringify(user));
  localStorage.setItem('staff_token', token);
  notify();
}

export function logout() {
  localStorage.removeItem('staff_user');
  localStorage.removeItem('staff_token');
  notify();
}

export function subscribe(fn: () => void) {
  listeners.push(fn);
  return () => { listeners = listeners.filter(l => l !== fn); };
}
