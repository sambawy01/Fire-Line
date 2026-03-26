import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Flame, Delete } from 'lucide-react';
import { api } from '../lib/api';
import { login } from '../stores/auth';

const LOCATIONS = [
  { id: 'loc_nimbu_main', name: 'Nimbu - Main St' },
  { id: 'loc_nimbu_downtown', name: 'Nimbu - Downtown' },
  { id: 'loc_nimbu_westside', name: 'Nimbu - Westside' },
  { id: 'loc_nimbu_northpark', name: 'Nimbu - North Park' },
];

export default function LoginPage() {
  const navigate = useNavigate();
  const [locationId, setLocationId] = useState(LOCATIONS[0].id);
  const [pin, setPin] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  function handleDigit(digit: string) {
    if (pin.length < 6) {
      setPin(prev => prev + digit);
      setError('');
    }
  }

  function handleBackspace() {
    setPin(prev => prev.slice(0, -1));
    setError('');
  }

  async function handleSubmit() {
    if (pin.length < 4) {
      setError('PIN must be at least 4 digits');
      return;
    }
    setLoading(true);
    setError('');
    try {
      const res = await api<{
        token: string;
        user: {
          user_id: string;
          org_id: string;
          role: string;
          display_name: string;
          staff_points: number;
          location_id: string;
        };
      }>('/auth/pin-login', {
        method: 'POST',
        body: JSON.stringify({ location_id: locationId, pin }),
      });
      login(res.user, res.token);
      navigate('/', { replace: true });
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Login failed');
      setPin('');
    } finally {
      setLoading(false);
    }
  }

  const digits = ['1', '2', '3', '4', '5', '6', '7', '8', '9'];

  return (
    <div className="flex flex-col items-center justify-center min-h-screen bg-slate-900 text-slate-100 px-4">
      {/* Logo */}
      <div className="flex flex-col items-center mb-8">
        <div className="w-16 h-16 rounded-2xl bg-orange-500/20 flex items-center justify-center mb-3">
          <Flame size={36} className="text-orange-400" />
        </div>
        <h1 className="text-2xl font-bold text-white">Staff Login</h1>
        <p className="text-sm text-slate-400 mt-1">FireLine by OpsNerve</p>
      </div>

      {/* Location selector */}
      <div className="w-full max-w-xs mb-6">
        <label htmlFor="location" className="block text-xs font-medium text-slate-400 mb-1.5">
          Location
        </label>
        <select
          id="location"
          value={locationId}
          onChange={e => setLocationId(e.target.value)}
          className="w-full px-3 py-2.5 rounded-lg bg-slate-800 border border-slate-700 text-sm text-white focus:outline-none focus:ring-2 focus:ring-orange-500"
        >
          {LOCATIONS.map(loc => (
            <option key={loc.id} value={loc.id}>{loc.name}</option>
          ))}
        </select>
      </div>

      {/* PIN display */}
      <div className="flex gap-2 mb-2" aria-label="PIN entry">
        {Array.from({ length: 6 }).map((_, i) => (
          <div
            key={i}
            className={`w-10 h-12 rounded-lg border-2 flex items-center justify-center text-xl font-mono ${
              i < pin.length
                ? 'border-orange-500 bg-orange-500/10 text-white'
                : 'border-slate-700 bg-slate-800 text-slate-600'
            }`}
          >
            {i < pin.length ? '\u2022' : ''}
          </div>
        ))}
      </div>

      {/* Error message */}
      {error && (
        <p className="text-red-400 text-sm mb-2" role="alert">{error}</p>
      )}

      {/* Keypad */}
      <div className="grid grid-cols-3 gap-2 w-full max-w-xs mt-4">
        {digits.map(d => (
          <button
            key={d}
            onClick={() => handleDigit(d)}
            disabled={loading}
            className="h-14 rounded-xl bg-slate-800 hover:bg-slate-700 active:bg-slate-600 text-xl font-semibold text-white transition-colors disabled:opacity-50"
          >
            {d}
          </button>
        ))}
        <button
          onClick={handleBackspace}
          disabled={loading || pin.length === 0}
          className="h-14 rounded-xl bg-slate-800 hover:bg-slate-700 active:bg-slate-600 flex items-center justify-center text-slate-400 transition-colors disabled:opacity-50"
          aria-label="Backspace"
        >
          <Delete size={24} />
        </button>
        <button
          onClick={() => handleDigit('0')}
          disabled={loading}
          className="h-14 rounded-xl bg-slate-800 hover:bg-slate-700 active:bg-slate-600 text-xl font-semibold text-white transition-colors disabled:opacity-50"
        >
          0
        </button>
        <button
          onClick={handleSubmit}
          disabled={loading || pin.length < 4}
          className="h-14 rounded-xl bg-orange-500 hover:bg-orange-600 active:bg-orange-700 text-white text-sm font-bold transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {loading ? '...' : 'Enter'}
        </button>
      </div>
    </div>
  );
}
