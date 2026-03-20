import type { LucideIcon } from 'lucide-react';
import { Inbox } from 'lucide-react';

interface EmptyStateProps {
  icon?: LucideIcon;
  title: string;
  description?: string;
  action?: { label: string; onClick: () => void };
}

export default function EmptyState({ icon: Icon = Inbox, title, description, action }: EmptyStateProps) {
  return (
    <div className="rounded-xl border border-dashed border-gray-300 bg-white py-16 text-center">
      <Icon className="mx-auto mb-3 h-10 w-10 text-gray-300" />
      <p className="text-lg font-medium text-gray-700">{title}</p>
      {description && <p className="mt-1 text-sm text-slate-300">{description}</p>}
      {action && (
        <button
          onClick={action.onClick}
          className="mt-4 rounded-lg bg-[#F97316] px-4 py-2 text-sm font-medium text-white hover:bg-[#EA580C] transition-colors"
        >
          {action.label}
        </button>
      )}
    </div>
  );
}
