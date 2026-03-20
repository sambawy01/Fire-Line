import { useState, type FormEvent } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import {
  Flame,
  Mail,
  Lock,
  Building2,
  Hash,
  User,
  Loader2,
  AlertCircle,
} from 'lucide-react';
import { useAuthStore } from '../stores/auth';

export default function SignupPage() {
  const navigate = useNavigate();
  const { signup, isLoading, error, clearError } = useAuthStore();

  const [orgName, setOrgName] = useState('');
  const [orgSlug, setOrgSlug] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [displayName, setDisplayName] = useState('');

  /** Auto-generate a slug from the org name when the user hasn't manually edited the slug. */
  const [slugTouched, setSlugTouched] = useState(false);

  const deriveSlug = (name: string) =>
    name
      .toLowerCase()
      .replace(/[^a-z0-9]+/g, '-')
      .replace(/^-|-$/g, '');

  const handleOrgNameChange = (value: string) => {
    setOrgName(value);
    if (!slugTouched) {
      setOrgSlug(deriveSlug(value));
    }
  };

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    await signup({
      org_name: orgName,
      org_slug: orgSlug,
      email,
      password,
      display_name: displayName,
    });
    if (useAuthStore.getState().isAuthenticated) {
      navigate('/', { replace: true });
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-slate-900 px-4 py-12">
      <div className="w-full max-w-md">
        {/* Branding */}
        <div className="mb-8 text-center">
          <div className="mx-auto mb-3 flex h-14 w-14 items-center justify-center rounded-xl bg-slate-900">
            <Flame className="h-8 w-8 text-orange-500" />
          </div>
          <h1 className="text-2xl font-bold tracking-tight text-white">
            Create your FireLine account
          </h1>
          <p className="mt-1 text-sm text-slate-400">
            Set up your organization and start managing operations
          </p>
        </div>

        {/* Card */}
        <div className="rounded-xl border border-white/10 bg-white/5 p-8 shadow-sm">
          <form onSubmit={handleSubmit} className="space-y-5">
            {/* Error banner */}
            {error && (
              <div className="flex items-start gap-2 rounded-lg border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-400">
                <AlertCircle className="mt-0.5 h-4 w-4 shrink-0" />
                <span>{error}</span>
                <button
                  type="button"
                  onClick={clearError}
                  className="ml-auto text-red-400 hover:text-red-600"
                  aria-label="Dismiss error"
                >
                  &times;
                </button>
              </div>
            )}

            {/* Organization name */}
            <div>
              <label
                htmlFor="orgName"
                className="mb-1.5 block text-sm font-medium text-slate-200"
              >
                Organization name
              </label>
              <div className="relative">
                <Building2 className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-300" />
                <input
                  id="orgName"
                  type="text"
                  required
                  placeholder="Bistro Cloud Kitchen"
                  value={orgName}
                  onChange={(e) => handleOrgNameChange(e.target.value)}
                  className="w-full rounded-lg border border-white/15 bg-white/10 py-2.5 pl-10 pr-3 text-sm text-white placeholder:text-slate-300 focus:border-orange-500 focus:outline-none focus:ring-2 focus:ring-orange-500/20"
                />
              </div>
            </div>

            {/* Organization slug */}
            <div>
              <label
                htmlFor="orgSlug"
                className="mb-1.5 block text-sm font-medium text-slate-200"
              >
                Organization slug
              </label>
              <div className="relative">
                <Hash className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-300" />
                <input
                  id="orgSlug"
                  type="text"
                  required
                  pattern="[a-z0-9]+(?:-[a-z0-9]+)*"
                  title="Lowercase letters, numbers, and hyphens only"
                  placeholder="bistro-cloud-kitchen"
                  value={orgSlug}
                  onChange={(e) => {
                    setSlugTouched(true);
                    setOrgSlug(e.target.value);
                  }}
                  className="w-full rounded-lg border border-white/15 bg-white/10 py-2.5 pl-10 pr-3 text-sm text-white placeholder:text-slate-300 focus:border-orange-500 focus:outline-none focus:ring-2 focus:ring-orange-500/20"
                />
              </div>
              <p className="mt-1 text-xs text-slate-300">
                Used in URLs. Lowercase letters, numbers, and hyphens only.
              </p>
            </div>

            {/* Display name */}
            <div>
              <label
                htmlFor="displayName"
                className="mb-1.5 block text-sm font-medium text-slate-200"
              >
                Your name
              </label>
              <div className="relative">
                <User className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-300" />
                <input
                  id="displayName"
                  type="text"
                  required
                  autoComplete="name"
                  placeholder="Jane Smith"
                  value={displayName}
                  onChange={(e) => setDisplayName(e.target.value)}
                  className="w-full rounded-lg border border-white/15 bg-white/10 py-2.5 pl-10 pr-3 text-sm text-white placeholder:text-slate-300 focus:border-orange-500 focus:outline-none focus:ring-2 focus:ring-orange-500/20"
                />
              </div>
            </div>

            {/* Email */}
            <div>
              <label
                htmlFor="email"
                className="mb-1.5 block text-sm font-medium text-slate-200"
              >
                Email
              </label>
              <div className="relative">
                <Mail className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-300" />
                <input
                  id="email"
                  type="email"
                  required
                  autoComplete="email"
                  placeholder="you@restaurant.com"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  className="w-full rounded-lg border border-white/15 bg-white/10 py-2.5 pl-10 pr-3 text-sm text-white placeholder:text-slate-300 focus:border-orange-500 focus:outline-none focus:ring-2 focus:ring-orange-500/20"
                />
              </div>
            </div>

            {/* Password */}
            <div>
              <label
                htmlFor="password"
                className="mb-1.5 block text-sm font-medium text-slate-200"
              >
                Password
              </label>
              <div className="relative">
                <Lock className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-300" />
                <input
                  id="password"
                  type="password"
                  required
                  minLength={8}
                  autoComplete="new-password"
                  placeholder="At least 8 characters"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  className="w-full rounded-lg border border-white/15 bg-white/10 py-2.5 pl-10 pr-3 text-sm text-white placeholder:text-slate-300 focus:border-orange-500 focus:outline-none focus:ring-2 focus:ring-orange-500/20"
                />
              </div>
            </div>

            {/* Submit */}
            <button
              type="submit"
              disabled={isLoading}
              className="flex w-full items-center justify-center gap-2 rounded-lg bg-orange-500 px-4 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-orange-600 focus:outline-none focus:ring-2 focus:ring-orange-500/50 focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-60"
            >
              {isLoading ? (
                <>
                  <Loader2 className="h-4 w-4 animate-spin" />
                  Creating account...
                </>
              ) : (
                'Create account'
              )}
            </button>
          </form>

          {/* Footer link */}
          <p className="mt-6 text-center text-sm text-slate-400">
            Already have an account?{' '}
            <Link
              to="/login"
              className="font-medium text-orange-500 hover:text-orange-600"
            >
              Sign in
            </Link>
          </p>
        </div>
      </div>
    </div>
  );
}
