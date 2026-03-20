import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  CheckCircle,
  Circle,
  ChevronRight,
  ChevronLeft,
  Zap,
  UtensilsCrossed,
  BarChart3,
  Users,
  ShoppingCart,
  Calendar,
  PackageOpen,
  TrendingUp,
  Star,
  Leaf,
  Clock,
  DollarSign,
  Building2,
  Loader2,
} from 'lucide-react';
import { useStartOnboarding, useUpdateStep, useCompleteChecklistItem, useOnboardingChecklist } from '../hooks/useOnboarding';
import { onboardingApi, type FirstInsights, type ChecklistItem } from '../lib/api';

// ─── Constants ───────────────────────────────────────────────────────────────

const STEPS = [
  'profile',
  'pos_connect',
  'importing',
  'first_insights',
  'concept_type',
  'priorities',
  'modules',
  'checklist',
] as const;

type WizardStep = (typeof STEPS)[number];

const STEP_LABELS: Record<WizardStep, string> = {
  profile: 'Profile',
  pos_connect: 'POS',
  importing: 'Import',
  first_insights: 'Insights',
  concept_type: 'Concept',
  priorities: 'Priorities',
  modules: 'Modules',
  checklist: 'Checklist',
};

const POS_SYSTEMS = [
  { id: 'toast', name: 'Toast', color: 'bg-red-500', letter: 'T' },
  { id: 'square', name: 'Square', color: 'bg-blue-600', letter: 'S' },
  { id: 'clover', name: 'Clover', color: 'bg-green-500', letter: 'C' },
  { id: 'csv', name: 'CSV Import', color: 'bg-gray-600', letter: '↑' },
];

const CONCEPT_TYPES: Record<string, { label: string; icon: typeof UtensilsCrossed; desc: string }> = {
  quick_service: {
    label: 'Quick Service',
    icon: Zap,
    desc: 'Fast, counter-service focused with high volume and speed priorities.',
  },
  fast_casual: {
    label: 'Fast Casual',
    icon: UtensilsCrossed,
    desc: 'Quality food with quick-service convenience and counter ordering.',
  },
  casual_dining: {
    label: 'Casual Dining',
    icon: UtensilsCrossed,
    desc: 'Table service with a comfortable, relaxed dining experience.',
  },
  upscale_casual: {
    label: 'Upscale Casual',
    icon: Star,
    desc: 'Elevated cuisine with attentive service in a polished environment.',
  },
  fine_dining: {
    label: 'Fine Dining',
    icon: Star,
    desc: 'Premium dining experience with exceptional service and cuisine.',
  },
};

const PRIORITY_OPTIONS = [
  {
    id: 'reduce_waste',
    label: 'Reduce Waste',
    desc: 'Cut food waste and control shrinkage',
    icon: Leaf,
  },
  {
    id: 'boost_revenue',
    label: 'Boost Revenue',
    desc: 'Drive sales and grow top-line performance',
    icon: TrendingUp,
  },
  {
    id: 'labor_efficiency',
    label: 'Labor Efficiency',
    desc: 'Optimize scheduling and reduce labor costs',
    icon: Clock,
  },
  {
    id: 'food_cost_control',
    label: 'Food Cost Control',
    desc: 'Lower COGS through better purchasing and recipes',
    icon: DollarSign,
  },
  {
    id: 'guest_experience',
    label: 'Guest Experience',
    desc: 'Improve satisfaction, loyalty, and retention',
    icon: Users,
  },
  {
    id: 'growth_insights',
    label: 'Growth Insights',
    desc: 'Multi-location visibility and portfolio analytics',
    icon: BarChart3,
  },
];

const MODULE_INFO: Record<string, { label: string; icon: typeof BarChart3; desc: string }> = {
  inventory: { label: 'Inventory', icon: PackageOpen, desc: 'PAR tracking, usage analysis, and waste logging.' },
  financial: { label: 'Financial', icon: DollarSign, desc: 'P&L, budget vs. actual, anomaly detection.' },
  labor: { label: 'Labor', icon: Users, desc: 'Staff performance, labor cost percentage tracking.' },
  scheduling: { label: 'Scheduling', icon: Calendar, desc: 'AI-assisted shift scheduling and coverage analysis.' },
  marketing: { label: 'Marketing', icon: TrendingUp, desc: 'Guest campaigns, loyalty programs, and promotions.' },
  vendor: { label: 'Vendors', icon: ShoppingCart, desc: 'Purchase orders, vendor scoring, and cost benchmarks.' },
  customers: { label: 'Customers', icon: Users, desc: 'Guest profiles, visit history, and preferences.' },
  menu_scoring: { label: 'Menu Scoring', icon: Star, desc: 'Menu engineering with star/dog/plow analysis.' },
  reporting: { label: 'Reporting', icon: BarChart3, desc: 'Automated reports and scheduled summaries.' },
  portfolio: { label: 'Portfolio', icon: Building2, desc: 'Cross-location KPIs and multi-site analytics.' },
  operations: { label: 'Operations', icon: Zap, desc: 'Kitchen display routing and service metrics.' },
};

// ─── Helpers ─────────────────────────────────────────────────────────────────

function cents(v: number) {
  return `$${(v / 100).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
}

function formatHour(h: number) {
  if (h === 0) return '12 AM';
  if (h < 12) return `${h} AM`;
  if (h === 12) return '12 PM';
  return `${h - 12} PM`;
}

// ─── Sub-components ───────────────────────────────────────────────────────────

function ProgressBar({ currentIndex }: { currentIndex: number }) {
  return (
    <div className="flex items-center justify-center gap-2 mb-10">
      {STEPS.map((step, i) => (
        <div key={step} className="flex items-center gap-2">
          <div
            className={`w-8 h-8 rounded-full flex items-center justify-center text-xs font-semibold transition-all ${
              i < currentIndex
                ? 'bg-orange-500 text-white'
                : i === currentIndex
                ? 'bg-orange-500 text-white ring-4 ring-orange-500/30'
                : 'bg-white/10 text-slate-500'
            }`}
            title={STEP_LABELS[step]}
          >
            {i < currentIndex ? <CheckCircle className="w-4 h-4" /> : i + 1}
          </div>
          {i < STEPS.length - 1 && (
            <div className={`h-0.5 w-8 ${i < currentIndex ? 'bg-orange-500' : 'bg-white/15'}`} />
          )}
        </div>
      ))}
    </div>
  );
}

function StepTitle({ title, subtitle }: { title: string; subtitle?: string }) {
  return (
    <div className="text-center mb-8">
      <h1 className="text-2xl font-bold text-white">{title}</h1>
      {subtitle && <p className="mt-2 text-slate-400">{subtitle}</p>}
    </div>
  );
}

function NavButtons({
  onBack,
  onNext,
  nextLabel = 'Continue',
  nextDisabled = false,
  loading = false,
}: {
  onBack?: () => void;
  onNext: () => void;
  nextLabel?: string;
  nextDisabled?: boolean;
  loading?: boolean;
}) {
  return (
    <div className="flex justify-between mt-8">
      {onBack ? (
        <button
          onClick={onBack}
          className="flex items-center gap-2 px-4 py-2 text-slate-300 hover:text-white transition-colors"
        >
          <ChevronLeft className="w-4 h-4" /> Back
        </button>
      ) : (
        <div />
      )}
      <button
        onClick={onNext}
        disabled={nextDisabled || loading}
        className="flex items-center gap-2 px-6 py-2.5 bg-orange-500 hover:bg-orange-600 disabled:opacity-50 disabled:cursor-not-allowed text-white rounded-lg font-medium transition-colors"
      >
        {loading && <Loader2 className="w-4 h-4 animate-spin" />}
        {nextLabel}
        {!loading && <ChevronRight className="w-4 h-4" />}
      </button>
    </div>
  );
}

// ─── Steps ────────────────────────────────────────────────────────────────────

function ProfileStep({
  onNext,
}: {
  onNext: (data: Record<string, string>) => void;
}) {
  const [form, setForm] = useState({
    restaurant_name: '',
    address: '',
    timezone: 'America/New_York',
    cuisine_type: '',
    seating_capacity: '',
  });

  const valid = form.restaurant_name.trim() !== '' && form.cuisine_type !== '';

  const set = (k: keyof typeof form) => (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) =>
    setForm((f) => ({ ...f, [k]: e.target.value }));

  return (
    <div>
      <StepTitle title="Tell us about your restaurant" subtitle="We'll personalize your experience based on your operation." />
      <div className="grid grid-cols-1 gap-4 max-w-lg mx-auto">
        <div>
          <label className="block text-sm font-medium text-slate-200 mb-1">Restaurant Name *</label>
          <input
            value={form.restaurant_name}
            onChange={set('restaurant_name')}
            placeholder="e.g. The Rustic Fork"
            className="w-full px-3 py-2 bg-white/10 text-white border border-white/15 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-400 placeholder:text-slate-500"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-slate-200 mb-1">Address</label>
          <input
            value={form.address}
            onChange={set('address')}
            placeholder="123 Main St, Chicago, IL 60601"
            className="w-full px-3 py-2 bg-white/10 text-white border border-white/15 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-400 placeholder:text-slate-500"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-slate-200 mb-1">Timezone</label>
          <select
            value={form.timezone}
            onChange={set('timezone')}
            className="w-full px-3 py-2 bg-white/10 text-white border border-white/15 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-400 placeholder:text-slate-500"
          >
            <option value="America/New_York">Eastern (ET)</option>
            <option value="America/Chicago">Central (CT)</option>
            <option value="America/Denver">Mountain (MT)</option>
            <option value="America/Los_Angeles">Pacific (PT)</option>
            <option value="Pacific/Honolulu">Hawaii (HT)</option>
            <option value="America/Anchorage">Alaska (AKT)</option>
          </select>
        </div>
        <div>
          <label className="block text-sm font-medium text-slate-200 mb-1">Cuisine Type *</label>
          <select
            value={form.cuisine_type}
            onChange={set('cuisine_type')}
            className="w-full px-3 py-2 bg-white/10 text-white border border-white/15 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-400 placeholder:text-slate-500"
          >
            <option value="">Select a cuisine…</option>
            <option value="american">American</option>
            <option value="italian">Italian</option>
            <option value="mexican">Mexican</option>
            <option value="asian">Asian / Pan-Asian</option>
            <option value="seafood">Seafood</option>
            <option value="steakhouse">Steakhouse</option>
            <option value="mediterranean">Mediterranean</option>
            <option value="pizza">Pizza</option>
            <option value="burgers">Burgers</option>
            <option value="breakfast_brunch">Breakfast / Brunch</option>
            <option value="bar_grill">Bar & Grill</option>
            <option value="other">Other</option>
          </select>
        </div>
        <div>
          <label className="block text-sm font-medium text-slate-200 mb-1">Seating Capacity</label>
          <input
            type="number"
            min={1}
            value={form.seating_capacity}
            onChange={set('seating_capacity')}
            placeholder="e.g. 80"
            className="w-full px-3 py-2 bg-white/10 text-white border border-white/15 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-400 placeholder:text-slate-500"
          />
        </div>
      </div>
      <NavButtons onNext={() => onNext(form)} nextDisabled={!valid} />
    </div>
  );
}

function POSConnectStep({ onBack, onNext }: { onBack: () => void; onNext: () => void }) {
  const [selected, setSelected] = useState<string | null>(null);
  const [connecting, setConnecting] = useState(false);
  const [connected, setConnected] = useState(false);

  function handleConnect() {
    if (!selected) return;
    setConnecting(true);
    setTimeout(() => {
      setConnecting(false);
      setConnected(true);
    }, 2000);
  }

  return (
    <div>
      <StepTitle
        title="Connect your POS"
        subtitle="Select your point-of-sale system to import historical data."
      />
      <div className="grid grid-cols-2 gap-4 max-w-md mx-auto mb-6">
        {POS_SYSTEMS.map((pos) => (
          <button
            key={pos.id}
            onClick={() => { setSelected(pos.id); setConnected(false); }}
            className={`p-5 rounded-xl border-2 flex flex-col items-center gap-3 transition-all ${
              selected === pos.id
                ? 'border-orange-500 bg-orange-500/10'
                : 'border-white/10 hover:border-white/15 bg-white/5'
            }`}
          >
            <div className={`w-12 h-12 rounded-xl ${pos.color} flex items-center justify-center text-white text-xl font-bold`}>
              {pos.letter}
            </div>
            <span className="text-sm font-medium text-white">{pos.name}</span>
          </button>
        ))}
      </div>

      {connected ? (
        <div className="text-center py-3 px-6 bg-green-500/10 border border-green-500/30 rounded-lg text-green-400 font-medium max-w-md mx-auto mb-4 flex items-center justify-center gap-2">
          <CheckCircle className="w-5 h-5" /> Connected successfully!
        </div>
      ) : (
        <div className="flex justify-center mb-4">
          <button
            onClick={handleConnect}
            disabled={!selected || connecting}
            className="flex items-center gap-2 px-6 py-2.5 bg-gray-800 hover:bg-slate-900 disabled:opacity-50 disabled:cursor-not-allowed text-white rounded-lg font-medium transition-colors"
          >
            {connecting && <Loader2 className="w-4 h-4 animate-spin" />}
            {connecting ? 'Connecting…' : 'Connect'}
          </button>
        </div>
      )}

      <NavButtons onBack={onBack} onNext={onNext} nextDisabled={!connected} />
    </div>
  );
}

const IMPORT_ITEMS = [
  '64 menu items imported',
  '125 transactions synced',
  '18 staff members loaded',
  'Historical sales data: 90 days',
  'Inventory levels baseline set',
  'Vendor catalog linked',
];

function ImportingStep({ onNext }: { onNext: () => void }) {
  const [visible, setVisible] = useState(0);
  const [done, setDone] = useState(false);

  useEffect(() => {
    if (visible < IMPORT_ITEMS.length) {
      const t = setTimeout(() => setVisible((v) => v + 1), 700);
      return () => clearTimeout(t);
    } else {
      const t = setTimeout(() => setDone(true), 800);
      return () => clearTimeout(t);
    }
  }, [visible]);

  return (
    <div>
      <StepTitle title="Importing your data" subtitle="We're pulling in your historical records. This usually takes less than a minute." />
      <div className="max-w-sm mx-auto space-y-3 mb-8">
        {IMPORT_ITEMS.map((item, i) => (
          <div
            key={item}
            className={`flex items-center gap-3 transition-all duration-500 ${
              i < visible ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-2'
            }`}
          >
            <CheckCircle className="w-5 h-5 text-green-500 flex-shrink-0" />
            <span className="text-slate-200">{item}</span>
          </div>
        ))}
        {!done && (
          <div className="flex items-center gap-3 text-slate-500">
            <Loader2 className="w-5 h-5 animate-spin" />
            <span>Processing…</span>
          </div>
        )}
      </div>
      {done && (
        <div className="text-center">
          <div className="inline-flex items-center gap-2 px-5 py-2.5 bg-green-500/10 border border-green-500/30 text-green-400 rounded-lg font-medium mb-6">
            <CheckCircle className="w-5 h-5" /> All data imported!
          </div>
          <div>
            <button
              onClick={onNext}
              className="flex items-center gap-2 px-6 py-2.5 bg-orange-500 hover:bg-orange-600 text-white rounded-lg font-medium transition-colors mx-auto"
            >
              See your first insights <ChevronRight className="w-4 h-4" />
            </button>
          </div>
        </div>
      )}
    </div>
  );
}

function FirstInsightsStep({
  insights,
  onBack,
  onNext,
}: {
  insights: FirstInsights | null;
  onBack: () => void;
  onNext: () => void;
}) {
  const demo: FirstInsights = insights ?? {
    daily_revenue_avg: 487500,
    top_sellers: ['Grilled Salmon', 'House Burger', 'Caesar Salad', 'Truffle Fries', 'Chicken Piccata'],
    peak_hour: 19,
    avg_check: 3850,
    void_rate: 2.3,
    staff_count: 18,
    check_count: 125,
  };

  const kpis = [
    { label: 'Daily Revenue Avg', value: cents(demo.daily_revenue_avg), icon: DollarSign, color: 'text-green-600' },
    { label: 'Avg Check Size', value: cents(demo.avg_check), icon: TrendingUp, color: 'text-blue-600' },
    { label: 'Peak Hour', value: formatHour(demo.peak_hour), icon: Clock, color: 'text-orange-600' },
    { label: 'Void Rate', value: `${demo.void_rate.toFixed(1)}%`, icon: BarChart3, color: 'text-red-500' },
    { label: 'Staff Count', value: String(demo.staff_count), icon: Users, color: 'text-purple-600' },
    { label: 'Checks Synced', value: String(demo.check_count), icon: CheckCircle, color: 'text-slate-300' },
  ];

  return (
    <div>
      <StepTitle title="Your first insights" subtitle="Based on your imported data, here's a snapshot of your operation." />
      <div className="grid grid-cols-2 md:grid-cols-3 gap-4 max-w-2xl mx-auto mb-6">
        {kpis.map((k) => (
          <div key={k.label} className="bg-white/5 border border-white/10 rounded-xl p-4">
            <div className="flex items-center gap-2 mb-1">
              <k.icon className={`w-4 h-4 ${k.color}`} />
              <span className="text-xs text-slate-400 uppercase tracking-wide">{k.label}</span>
            </div>
            <div className="text-xl font-bold text-white">{k.value}</div>
          </div>
        ))}
      </div>
      {demo.top_sellers.length > 0 && (
        <div className="max-w-2xl mx-auto bg-white/5 border border-white/10 rounded-xl p-4 mb-4">
          <p className="text-xs text-slate-400 uppercase tracking-wide mb-2">Top Sellers</p>
          <div className="flex flex-wrap gap-2">
            {demo.top_sellers.map((s) => (
              <span key={s} className="px-3 py-1 bg-orange-500/20 text-orange-400 rounded-full text-sm font-medium">
                {s}
              </span>
            ))}
          </div>
        </div>
      )}
      <NavButtons onBack={onBack} onNext={onNext} />
    </div>
  );
}

function ConceptTypeStep({
  conceptType,
  onBack,
  onNext,
}: {
  conceptType: string;
  onBack: () => void;
  onNext: (type: string) => void;
}) {
  const [selected, setSelected] = useState(conceptType || 'casual_dining');
  const info = CONCEPT_TYPES[selected] ?? CONCEPT_TYPES['casual_dining'];
  const Icon = info.icon;

  return (
    <div>
      <StepTitle title="Your restaurant concept" subtitle="We inferred your concept from your average check size. Is this right?" />
      <div className="max-w-sm mx-auto mb-6">
        <div className="p-6 bg-orange-500/10 border-2 border-orange-500/40 rounded-xl text-center mb-4">
          <Icon className="w-10 h-10 text-orange-500 mx-auto mb-2" />
          <div className="text-lg font-bold text-white">{info.label}</div>
          <p className="text-sm text-slate-300 mt-1">{info.desc}</p>
        </div>
        <p className="text-sm text-slate-400 text-center mb-3">Not quite right? Choose below:</p>
        <div className="grid grid-cols-1 gap-2">
          {Object.entries(CONCEPT_TYPES).map(([key, ct]) => {
            const Ic = ct.icon;
            return (
              <button
                key={key}
                onClick={() => setSelected(key)}
                className={`flex items-center gap-3 px-4 py-2.5 rounded-lg border transition-all text-left ${
                  selected === key
                    ? 'border-orange-500 bg-orange-500/10 text-orange-400'
                    : 'border-white/10 hover:border-white/15 text-slate-200'
                }`}
              >
                <Ic className="w-4 h-4 flex-shrink-0" />
                <span className="text-sm font-medium">{ct.label}</span>
              </button>
            );
          })}
        </div>
      </div>
      <NavButtons onBack={onBack} onNext={() => onNext(selected)} nextLabel="Confirm" />
    </div>
  );
}

function PrioritiesStep({
  onBack,
  onNext,
}: {
  onBack: () => void;
  onNext: (priorities: string[]) => void;
}) {
  const [selected, setSelected] = useState<string[]>([]);

  function toggle(id: string) {
    setSelected((prev) => {
      if (prev.includes(id)) return prev.filter((p) => p !== id);
      if (prev.length >= 3) return prev;
      return [...prev, id];
    });
  }

  return (
    <div>
      <StepTitle
        title="What are your top priorities?"
        subtitle="Choose up to 3. We'll recommend modules and build your personalized checklist."
      />
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 max-w-xl mx-auto mb-2">
        {PRIORITY_OPTIONS.map((p) => {
          const Icon = p.icon;
          const active = selected.includes(p.id);
          const maxed = selected.length >= 3 && !active;
          return (
            <button
              key={p.id}
              onClick={() => toggle(p.id)}
              disabled={maxed}
              className={`flex items-start gap-3 p-4 rounded-xl border-2 text-left transition-all ${
                active
                  ? 'border-orange-500 bg-orange-500/10'
                  : maxed
                  ? 'border-white/5 bg-white/5 opacity-50 cursor-not-allowed'
                  : 'border-white/10 hover:border-white/15 bg-white/5'
              }`}
            >
              <div className={`w-9 h-9 rounded-lg flex items-center justify-center flex-shrink-0 ${active ? 'bg-orange-500/20' : 'bg-white/10'}`}>
                <Icon className={`w-5 h-5 ${active ? 'text-orange-400' : 'text-slate-400'}`} />
              </div>
              <div>
                <div className="font-semibold text-sm text-white">{p.label}</div>
                <div className="text-xs text-slate-400 mt-0.5">{p.desc}</div>
              </div>
              {active && <CheckCircle className="w-5 h-5 text-orange-500 ml-auto flex-shrink-0" />}
            </button>
          );
        })}
      </div>
      <p className="text-center text-sm text-slate-500 mb-2">{selected.length}/3 selected</p>
      <NavButtons onBack={onBack} onNext={() => onNext(selected)} nextDisabled={selected.length === 0} />
    </div>
  );
}

function ModulesStep({
  recommendedModules,
  onBack,
  onNext,
}: {
  recommendedModules: string[];
  onBack: () => void;
  onNext: (modules: string[]) => void;
}) {
  const [active, setActive] = useState<string[]>(recommendedModules);

  function toggle(id: string) {
    setActive((prev) => (prev.includes(id) ? prev.filter((m) => m !== id) : [...prev, id]));
  }

  const allModules = Object.keys(MODULE_INFO);

  return (
    <div>
      <StepTitle
        title="Choose your modules"
        subtitle="We've pre-selected based on your priorities. Toggle anything to customize."
      />
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 max-w-xl mx-auto mb-4">
        {allModules.map((id) => {
          const m = MODULE_INFO[id];
          if (!m) return null;
          const Icon = m.icon;
          const isActive = active.includes(id);
          const isRecommended = recommendedModules.includes(id);
          return (
            <button
              key={id}
              onClick={() => toggle(id)}
              className={`flex items-start gap-3 p-4 rounded-xl border-2 text-left transition-all ${
                isActive
                  ? 'border-orange-500 bg-orange-500/10'
                  : 'border-white/10 hover:border-white/15 bg-white/5'
              }`}
            >
              <div className={`w-9 h-9 rounded-lg flex items-center justify-center flex-shrink-0 ${isActive ? 'bg-orange-500/20' : 'bg-white/10'}`}>
                <Icon className={`w-5 h-5 ${isActive ? 'text-orange-400' : 'text-slate-400'}`} />
              </div>
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2">
                  <span className="font-semibold text-sm text-white">{m.label}</span>
                  {isRecommended && (
                    <span className="text-xs px-1.5 py-0.5 bg-orange-500/20 text-orange-400 rounded font-medium">
                      Recommended
                    </span>
                  )}
                </div>
                <div className="text-xs text-slate-400 mt-0.5">{m.desc}</div>
              </div>
              <div className={`w-5 h-5 rounded border-2 flex-shrink-0 flex items-center justify-center ${isActive ? 'bg-orange-500 border-orange-500' : 'border-white/15'}`}>
                {isActive && <CheckCircle className="w-3 h-3 text-white" />}
              </div>
            </button>
          );
        })}
      </div>
      <NavButtons onBack={onBack} onNext={() => onNext(active)} nextDisabled={active.length === 0} />
    </div>
  );
}

function ChecklistStep({
  items,
  onCompleteItem,
  onFinish,
  onBack,
}: {
  items: ChecklistItem[];
  onCompleteItem: (id: string) => void;
  onFinish: () => void;
  onBack: () => void;
}) {
  const completed = items.filter((i) => i.completed).length;
  const total = items.length;
  const pct = total > 0 ? Math.round((completed / total) * 100) : 0;

  const byCategory = items.reduce<Record<string, ChecklistItem[]>>((acc, item) => {
    if (!acc[item.category]) acc[item.category] = [];
    acc[item.category].push(item);
    return acc;
  }, {});

  return (
    <div>
      <StepTitle
        title="Your personalized checklist"
        subtitle="Complete these steps to get the most out of FireLine."
      />
      <div className="max-w-xl mx-auto mb-6">
        <div className="flex items-center justify-between text-sm text-slate-300 mb-2">
          <span>{completed} of {total} complete</span>
          <span className="font-semibold text-orange-600">{pct}%</span>
        </div>
        <div className="w-full bg-white/10 rounded-full h-2.5 mb-6">
          <div
            className="bg-orange-500 h-2.5 rounded-full transition-all"
            style={{ width: `${pct}%` }}
          />
        </div>

        {Object.entries(byCategory).map(([category, catItems]) => (
          <div key={category} className="mb-5">
            <h3 className="text-xs font-semibold text-slate-500 uppercase tracking-widest mb-2 capitalize">
              {category.replace(/_/g, ' ')}
            </h3>
            <div className="space-y-2">
              {catItems.map((item) => (
                <div
                  key={item.item_id}
                  className={`flex items-start gap-3 p-3 rounded-lg border transition-all ${
                    item.completed ? 'bg-white/5 border-white/10 opacity-70' : 'bg-white/5 border-white/10'
                  }`}
                >
                  <button
                    onClick={() => !item.completed && onCompleteItem(item.item_id)}
                    className="flex-shrink-0 mt-0.5"
                    disabled={item.completed}
                  >
                    {item.completed ? (
                      <CheckCircle className="w-5 h-5 text-green-500" />
                    ) : (
                      <Circle className="w-5 h-5 text-slate-500 hover:text-orange-400 transition-colors" />
                    )}
                  </button>
                  <div>
                    <div className={`text-sm font-medium ${item.completed ? 'line-through text-slate-500' : 'text-white'}`}>
                      {item.title}
                    </div>
                    {item.description && (
                      <div className="text-xs text-slate-400 mt-0.5">{item.description}</div>
                    )}
                  </div>
                </div>
              ))}
            </div>
          </div>
        ))}
      </div>

      <div className="flex justify-between max-w-xl mx-auto">
        <button
          onClick={onBack}
          className="flex items-center gap-2 px-4 py-2 text-slate-300 hover:text-white transition-colors"
        >
          <ChevronLeft className="w-4 h-4" /> Back
        </button>
        <button
          onClick={onFinish}
          className="flex items-center gap-2 px-6 py-2.5 bg-orange-500 hover:bg-orange-600 text-white rounded-lg font-medium transition-colors"
        >
          Go to Dashboard <ChevronRight className="w-4 h-4" />
        </button>
      </div>
    </div>
  );
}

// ─── Main Page ────────────────────────────────────────────────────────────────

export default function OnboardingPage() {
  const navigate = useNavigate();
  const [step, setStep] = useState<WizardStep>('profile');
  const [sessionId, setSessionId] = useState<string>('');
  const [insights, setInsights] = useState<FirstInsights | null>(null);
  const [conceptType, setConceptType] = useState('casual_dining');
  const [recommendedModules, setRecommendedModules] = useState<string[]>([]);
  const [checklistItems, setChecklistItems] = useState<ChecklistItem[]>([]);

  const startOnboarding = useStartOnboarding();
  const updateStep = useUpdateStep();
  const completeItem = useCompleteChecklistItem();
  const { data: checklistData } = useOnboardingChecklist();

  // Keep checklist in sync with server data
  useEffect(() => {
    if (checklistData?.items) {
      setChecklistItems(checklistData.items);
    }
  }, [checklistData]);

  // Start session on mount using stored user_id
  useEffect(() => {
    const userId = localStorage.getItem('user_id') || 'demo-user';
    startOnboarding.mutate(userId, {
      onSuccess: (data) => {
        setSessionId(data.session.session_id);
      },
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const stepIndex = STEPS.indexOf(step);

  function goTo(s: WizardStep) {
    setStep(s);
  }

  async function advanceStep(targetStep: WizardStep, data: Record<string, unknown> = {}) {
    if (sessionId) {
      try {
        await updateStep.mutateAsync({ sessionId, step: targetStep, data });
      } catch {
        // Non-fatal: proceed anyway in wizard flow
      }
    }
    goTo(targetStep);
  }

  // ── Step handlers ──

  async function handleProfile(data: Record<string, string>) {
    await advanceStep('pos_connect', data);
  }

  function handlePOSConnect() {
    goTo('importing');
  }

  function handleImporting() {
    // Fetch simulated insights
    onboardingApi.getInsights('demo-location').then((r) => setInsights(r.insights)).catch(() => {});
    goTo('first_insights');
  }

  async function handleFirstInsights() {
    await advanceStep('concept_type', {});
    // Fetch concept inference
    onboardingApi.inferConcept('demo-location').then((r) => setConceptType(r.concept_type)).catch(() => {});
  }

  async function handleConceptType(type: string) {
    setConceptType(type);
    await advanceStep('priorities', { concept_type: type });
  }

  async function handlePriorities(priorities: string[]) {
    // Fetch recommended modules
    const mods = await onboardingApi.recommendModules(priorities).catch(() => ({ modules: [] }));
    setRecommendedModules(mods.modules);
    await advanceStep('modules', { priorities });
  }

  async function handleModules(modules: string[]) {
    setRecommendedModules(modules);
    // Generate checklist on server
    try {
      const res = await onboardingApi.getChecklist();
      setChecklistItems(res.items);
    } catch {
      // Will generate from defaults if not available
    }
    await advanceStep('checklist', { modules });
  }

  function handleCompleteItem(itemId: string) {
    // Optimistic update
    setChecklistItems((prev) =>
      prev.map((it) => (it.item_id === itemId ? { ...it, completed: true } : it))
    );
    completeItem.mutate(itemId);
  }

  async function handleFinish() {
    await advanceStep('complete', {});
    navigate('/');
  }

  return (
    <div className="min-h-screen bg-slate-900 flex flex-col">
      {/* Header */}
      <header className="bg-white/5 border-b border-white/10 px-6 py-4">
        <div className="max-w-3xl mx-auto flex items-center gap-2">
          <div className="w-7 h-7 bg-orange-500 rounded-md" />
          <span className="font-bold text-white text-lg">FireLine</span>
          <span className="text-slate-500 mx-2">|</span>
          <span className="text-slate-400 text-sm">Setup Wizard</span>
        </div>
      </header>

      {/* Content */}
      <main className="flex-1 px-4 py-10">
        <div className="max-w-3xl mx-auto">
          <ProgressBar currentIndex={stepIndex} />

          {step === 'profile' && <ProfileStep onNext={handleProfile} />}

          {step === 'pos_connect' && (
            <POSConnectStep
              onBack={() => goTo('profile')}
              onNext={handlePOSConnect}
            />
          )}

          {step === 'importing' && <ImportingStep onNext={handleImporting} />}

          {step === 'first_insights' && (
            <FirstInsightsStep
              insights={insights}
              onBack={() => goTo('pos_connect')}
              onNext={handleFirstInsights}
            />
          )}

          {step === 'concept_type' && (
            <ConceptTypeStep
              conceptType={conceptType}
              onBack={() => goTo('first_insights')}
              onNext={handleConceptType}
            />
          )}

          {step === 'priorities' && (
            <PrioritiesStep
              onBack={() => goTo('concept_type')}
              onNext={handlePriorities}
            />
          )}

          {step === 'modules' && (
            <ModulesStep
              recommendedModules={recommendedModules}
              onBack={() => goTo('priorities')}
              onNext={handleModules}
            />
          )}

          {step === 'checklist' && (
            <ChecklistStep
              items={checklistItems}
              onCompleteItem={handleCompleteItem}
              onFinish={handleFinish}
              onBack={() => goTo('modules')}
            />
          )}
        </div>
      </main>
    </div>
  );
}
