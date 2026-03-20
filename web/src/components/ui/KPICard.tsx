import type { LucideIcon } from 'lucide-react';

interface KPICardProps {
  label: string;
  value: string;
  icon: LucideIcon;
  iconColor?: string;
  bgTint?: string;
}

export default function KPICard({
  label,
  value,
  icon: Icon,
  iconColor = 'text-slate-300',
  bgTint = 'bg-gray-50',
}: KPICardProps) {
  return (
    <div className="bg-white/5 rounded-xl border border-white/10 p-5 flex items-start gap-4">
      <div className={`${bgTint} p-3 rounded-lg`}>
        <Icon className={`h-6 w-6 ${iconColor}`} />
      </div>
      <div>
        <p className="text-sm text-slate-400">{label}</p>
        <p className="text-2xl font-bold text-white mt-0.5">{value}</p>
      </div>
    </div>
  );
}
