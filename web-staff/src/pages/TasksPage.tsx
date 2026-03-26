import { useState, useEffect, useCallback } from 'react';
import {
  CheckCircle2,
  Circle,
  Loader2,
  X,
  Camera,
  AlertCircle,
} from 'lucide-react';
import { api } from '../lib/api';

/* ---------- types ---------- */
interface DataEntryConfig {
  label: string;
  unit: string;
  expected_min?: number;
  expected_max?: number;
}

interface Task {
  id: string;
  title: string;
  status: 'pending' | 'in_progress' | 'completed';
  priority: 'urgent' | 'normal' | 'low';
  type: 'checklist' | 'ad-hoc' | 'data_entry';
  due_time?: string | null;
  completed_at?: string | null;
  data_entry_config?: DataEntryConfig | null;
}

type FilterTab = 'all' | 'pending' | 'in_progress' | 'completed';

/* ---------- helpers ---------- */
function formatTime(iso: string): string {
  const d = new Date(iso);
  return d.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit' });
}

function formatTimestamp(iso: string): string {
  const d = new Date(iso);
  return d.toLocaleString('en-US', {
    month: 'short',
    day: 'numeric',
    hour: 'numeric',
    minute: '2-digit',
  });
}

const priorityDot: Record<string, string> = {
  urgent: 'bg-red-500',
  normal: 'bg-orange-500',
  low: 'bg-slate-500',
};

const typeBadge: Record<string, { bg: string; text: string; label: string }> = {
  checklist: { bg: 'bg-blue-500/20', text: 'text-blue-400', label: 'Checklist' },
  'ad-hoc': { bg: 'bg-purple-500/20', text: 'text-purple-400', label: 'Ad-hoc' },
  data_entry: { bg: 'bg-amber-500/20', text: 'text-amber-400', label: 'Data Entry' },
};

const filterTabs: { key: FilterTab; label: string }[] = [
  { key: 'all', label: 'All' },
  { key: 'pending', label: 'Pending' },
  { key: 'in_progress', label: 'In Progress' },
  { key: 'completed', label: 'Completed' },
];

/* ---------- completion modal ---------- */
function CompletionModal({
  task,
  onClose,
  onComplete,
}: {
  task: Task;
  onClose: () => void;
  onComplete: (task: Task) => void;
}) {
  const [value, setValue] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState('');

  const isDataEntry = task.type === 'data_entry' && task.data_entry_config;
  const config = task.data_entry_config;

  const handleSubmit = useCallback(async () => {
    setError('');
    setSubmitting(true);

    try {
      if (isDataEntry && config) {
        const num = parseFloat(value);
        if (isNaN(num)) {
          setError('Please enter a valid number');
          setSubmitting(false);
          return;
        }
        if (config.expected_min !== undefined && num < config.expected_min) {
          setError(`Value below expected range (min ${config.expected_min})`);
          setSubmitting(false);
          return;
        }
        if (config.expected_max !== undefined && num > config.expected_max) {
          setError(`Value above expected range (max ${config.expected_max})`);
          setSubmitting(false);
          return;
        }
        await api(`/tasks/${task.id}/complete`, {
          method: 'PUT',
          body: JSON.stringify({
            status: 'completed',
            data_entry_value: { value: num, unit: config.unit },
          }),
        });
      } else {
        await api(`/tasks/${task.id}/complete`, {
          method: 'PUT',
          body: JSON.stringify({ status: 'completed' }),
        });
      }
      onComplete({ ...task, status: 'completed', completed_at: new Date().toISOString() });
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to complete task');
    } finally {
      setSubmitting(false);
    }
  }, [task, isDataEntry, config, value, onComplete]);

  return (
    <div className="fixed inset-0 z-50 flex items-end sm:items-center justify-center">
      {/* backdrop */}
      <div className="absolute inset-0 bg-black/60" onClick={onClose} />

      {/* panel */}
      <div className="relative w-full max-w-md bg-slate-800 rounded-t-2xl sm:rounded-2xl border border-white/10 p-5 mx-4 mb-0 sm:mb-0">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-base font-semibold text-white">Complete Task</h3>
          <button onClick={onClose} className="text-slate-400 hover:text-white">
            <X className="h-5 w-5" />
          </button>
        </div>

        <p className="text-sm text-slate-300 mb-4">{task.title}</p>

        {isDataEntry && config && (
          <div className="mb-4">
            <label className="block text-xs text-slate-400 mb-1">{config.label}</label>
            <div className="flex items-center gap-2">
              <input
                type="number"
                value={value}
                onChange={(e) => setValue(e.target.value)}
                placeholder={
                  config.expected_min !== undefined && config.expected_max !== undefined
                    ? `${config.expected_min} - ${config.expected_max}`
                    : 'Enter value'
                }
                className="flex-1 rounded-lg bg-white/5 border border-white/10 px-3 py-2.5 text-sm text-white placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-orange-500/50"
              />
              <span className="text-sm text-slate-400 shrink-0">{config.unit}</span>
            </div>
            {config.expected_min !== undefined && config.expected_max !== undefined && (
              <p className="text-[11px] text-slate-500 mt-1">
                Expected range: {config.expected_min} &ndash; {config.expected_max} {config.unit}
              </p>
            )}
          </div>
        )}

        {error && (
          <div className="flex items-center gap-2 text-red-400 text-xs mb-3">
            <AlertCircle className="h-3.5 w-3.5 shrink-0" />
            {error}
          </div>
        )}

        {/* photo placeholder */}
        <button
          type="button"
          className="flex items-center gap-2 text-xs text-slate-500 mb-4 hover:text-slate-400 transition-colors"
          disabled
        >
          <Camera className="h-4 w-4" />
          Attach photo (coming soon)
        </button>

        <div className="flex gap-3">
          <button
            onClick={onClose}
            className="flex-1 rounded-lg border border-white/10 py-2.5 text-sm font-medium text-slate-400 hover:text-white transition-colors"
          >
            Cancel
          </button>
          <button
            onClick={handleSubmit}
            disabled={submitting || (isDataEntry && !value)}
            className="flex-1 rounded-lg bg-emerald-600 hover:bg-emerald-700 disabled:opacity-50 disabled:cursor-not-allowed py-2.5 text-sm font-semibold text-white transition-colors flex items-center justify-center gap-2"
          >
            {submitting && <Loader2 className="h-4 w-4 animate-spin" />}
            Complete
          </button>
        </div>
      </div>
    </div>
  );
}

/* ---------- task card ---------- */
function TaskCard({
  task,
  onUpdate,
}: {
  task: Task;
  onUpdate: (updated: Task) => void;
}) {
  const [acting, setActing] = useState(false);
  const [showModal, setShowModal] = useState(false);

  const handleStart = useCallback(async () => {
    setActing(true);
    try {
      await api(`/tasks/${task.id}/status`, {
        method: 'PUT',
        body: JSON.stringify({ status: 'in_progress' }),
      });
      onUpdate({ ...task, status: 'in_progress' });
    } catch {
      /* swallow for now */
    } finally {
      setActing(false);
    }
  }, [task, onUpdate]);

  const badge = typeBadge[task.type] ?? typeBadge['ad-hoc'];

  return (
    <>
      <div className="rounded-xl border border-white/10 bg-white/5 p-4">
        <div className="flex items-start gap-3">
          {/* priority dot */}
          <div className={`mt-1.5 h-2.5 w-2.5 rounded-full shrink-0 ${priorityDot[task.priority]}`} />

          <div className="flex-1 min-w-0">
            {/* title row */}
            <div className="flex items-center gap-2 flex-wrap">
              <h3 className="text-sm font-semibold text-white truncate">{task.title}</h3>
              <span className={`text-[10px] px-1.5 py-0.5 rounded-full font-medium ${badge.bg} ${badge.text}`}>
                {badge.label}
              </span>
            </div>

            {/* due time */}
            {task.due_time && task.status !== 'completed' && (
              <p className="text-xs text-slate-500 mt-1">Due by {formatTime(task.due_time)}</p>
            )}

            {/* completed timestamp */}
            {task.status === 'completed' && task.completed_at && (
              <p className="text-xs text-emerald-500/70 mt-1">
                Completed {formatTimestamp(task.completed_at)}
              </p>
            )}

            {/* actions */}
            <div className="mt-3">
              {task.status === 'pending' && (
                <button
                  onClick={handleStart}
                  disabled={acting}
                  className="rounded-lg bg-orange-500/20 hover:bg-orange-500/30 text-orange-400 text-xs font-semibold px-4 py-1.5 transition-colors disabled:opacity-50 flex items-center gap-1.5"
                >
                  {acting && <Loader2 className="h-3 w-3 animate-spin" />}
                  Start
                </button>
              )}
              {task.status === 'in_progress' && (
                <button
                  onClick={() => setShowModal(true)}
                  className="rounded-lg bg-emerald-500/20 hover:bg-emerald-500/30 text-emerald-400 text-xs font-semibold px-4 py-1.5 transition-colors"
                >
                  Complete
                </button>
              )}
              {task.status === 'completed' && (
                <div className="flex items-center gap-1.5 text-emerald-400">
                  <CheckCircle2 className="h-4 w-4" />
                  <span className="text-xs font-medium">Done</span>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>

      {showModal && (
        <CompletionModal
          task={task}
          onClose={() => setShowModal(false)}
          onComplete={(updated) => {
            onUpdate(updated);
            setShowModal(false);
          }}
        />
      )}
    </>
  );
}

/* ========== PAGE ========== */
export default function TasksPage() {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState<FilterTab>('pending');

  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        const data = await api<{ tasks: Task[] }>('/tasks/my');
        if (!cancelled) setTasks(data.tasks ?? []);
      } catch {
        /* empty */
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();
    return () => { cancelled = true; };
  }, []);

  const handleUpdate = useCallback((updated: Task) => {
    setTasks((prev) => prev.map((t) => (t.id === updated.id ? updated : t)));
  }, []);

  const filtered = filter === 'all' ? tasks : tasks.filter((t) => t.status === filter);

  return (
    <div className="p-4 pb-24 max-w-lg mx-auto">
      <h1 className="text-lg font-bold text-white mb-4">Tasks</h1>

      {/* filter tabs */}
      <div className="flex gap-2 overflow-x-auto pb-3 -mx-4 px-4 no-scrollbar">
        {filterTabs.map((tab) => (
          <button
            key={tab.key}
            onClick={() => setFilter(tab.key)}
            className={`shrink-0 rounded-full px-4 py-1.5 text-xs font-medium transition-colors ${
              filter === tab.key
                ? 'bg-orange-500 text-white'
                : 'bg-white/5 text-slate-400 hover:bg-white/10'
            }`}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {/* content */}
      {loading ? (
        <div className="flex justify-center py-16">
          <Loader2 className="h-6 w-6 animate-spin text-slate-500" />
        </div>
      ) : filtered.length === 0 ? (
        <div className="flex flex-col items-center py-16 text-slate-500">
          {filter === 'completed' ? (
            <>
              <Circle className="h-12 w-12 mb-3 text-slate-600" />
              <p className="text-sm">No completed tasks yet.</p>
            </>
          ) : (
            <>
              <CheckCircle2 className="h-12 w-12 mb-3 text-emerald-600/50" />
              <p className="text-sm font-medium text-slate-400">All caught up!</p>
              <p className="text-xs mt-1">No {filter === 'all' ? '' : filter.replace('_', ' ')} tasks.</p>
            </>
          )}
        </div>
      ) : (
        <div className="space-y-3">
          {filtered.map((task) => (
            <TaskCard key={task.id} task={task} onUpdate={handleUpdate} />
          ))}
        </div>
      )}
    </div>
  );
}
