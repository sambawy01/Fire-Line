import { useState } from 'react';
import {
  Plus,
  Eye,
  Pause,
  Play,
  Clock,
  Plug,
} from 'lucide-react';

type AdapterStatus = 'active' | 'paused' | 'errored';

interface Adapter {
  id: string;
  name: string;
  type: string;
  location: string;
  status: AdapterStatus;
  brandColor: string;
  capabilities: string[];
  lastSync: string;
}

const MOCK_ADAPTERS: Adapter[] = [
  {
    id: 'adapter-001',
    name: 'Loyverse POS — Nimbu Egypt',
    type: 'Loyverse',
    location: 'All Branches',
    status: 'active',
    brandColor: '#7C3AED',
    capabilities: [
      'READ_ORDERS',
      'READ_MENU',
      'READ_INVENTORY',
      'READ_EMPLOYEES',
      'READ_CUSTOMERS',
      'WRITE_86_STATUS',
    ],
    lastSync: new Date().toISOString(),
  },
  {
    id: 'adapter-002',
    name: 'Toast POS — Nimbu El Gouna',
    type: 'Toast',
    location: 'El Gouna, Red Sea',
    status: 'active',
    brandColor: '#FF6B35',
    capabilities: [
      'READ_ORDERS',
      'READ_MENU',
      'WRITE_MENU',
      'READ_PAYMENTS',
      'READ_LABOR',
    ],
    lastSync: new Date(Date.now() - 3600000).toISOString(),
  },
  {
    id: 'adapter-003',
    name: 'Toast POS — Nimbu New Cairo',
    type: 'Toast',
    location: 'New Cairo, Cairo',
    status: 'active',
    brandColor: '#FF6B35',
    capabilities: [
      'READ_ORDERS',
      'READ_MENU',
      'WRITE_MENU',
      'READ_PAYMENTS',
      'READ_LABOR',
    ],
    lastSync: new Date(Date.now() - 7200000).toISOString(),
  },
  {
    id: 'adapter-004',
    name: 'Toast POS — Nimbu Zayed',
    type: 'Toast',
    location: 'Sheikh Zayed, Giza',
    status: 'active',
    brandColor: '#FF6B35',
    capabilities: [
      'READ_ORDERS',
      'READ_MENU',
      'READ_PAYMENTS',
    ],
    lastSync: new Date(Date.now() - 5400000).toISOString(),
  },
  {
    id: 'adapter-005',
    name: 'Toast POS — Nimbu North Coast',
    type: 'Toast',
    location: 'North Coast',
    status: 'paused',
    brandColor: '#FF6B35',
    capabilities: [
      'READ_ORDERS',
      'READ_MENU',
      'READ_PAYMENTS',
    ],
    lastSync: new Date(Date.now() - 86400000).toISOString(),
  },
];

const STATUS_CONFIG: Record<
  AdapterStatus,
  { label: string; dot: string; bg: string; text: string }
> = {
  active: {
    label: 'Active',
    dot: 'bg-green-500',
    bg: 'bg-green-50',
    text: 'text-green-700',
  },
  paused: {
    label: 'Paused',
    dot: 'bg-amber-500',
    bg: 'bg-amber-50',
    text: 'text-amber-700',
  },
  errored: {
    label: 'Errored',
    dot: 'bg-red-500',
    bg: 'bg-red-50',
    text: 'text-red-700',
  },
};

function formatTimestamp(iso: string): string {
  const date = new Date(iso);
  return date.toLocaleString('en-US', {
    month: 'short',
    day: 'numeric',
    hour: 'numeric',
    minute: '2-digit',
    hour12: true,
  });
}

export default function AdaptersPage() {
  const [adapters, setAdapters] = useState<Adapter[]>(MOCK_ADAPTERS);

  function handleTogglePause(id: string) {
    setAdapters((prev) =>
      prev.map((a) => {
        if (a.id !== id) return a;
        const nextStatus: AdapterStatus =
          a.status === 'paused' ? 'active' : 'paused';
        console.log(`Adapter ${id} toggled to ${nextStatus}`);
        return { ...a, status: nextStatus };
      }),
    );
  }

  function handleViewDetails(id: string) {
    console.log(`View details for adapter: ${id}`);
  }

  return (
    <div className="min-h-screen">
      <div className="mx-auto max-w-5xl px-4 py-8 sm:px-6 lg:px-8">
        {/* Header */}
        <div className="mb-6 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <Plug className="h-7 w-7 text-white" />
            <h1 className="text-2xl font-bold text-white">
              POS Connections
            </h1>
          </div>
          <button className="inline-flex items-center gap-2 rounded-lg bg-[#F97316] px-4 py-2 text-sm font-semibold text-white transition-colors hover:bg-[#EA580C]">
            <Plus className="h-4 w-4" />
            Connect New POS
          </button>
        </div>

        {/* Adapter Grid */}
        <div className="grid gap-6 sm:grid-cols-2">
          {adapters.map((adapter) => {
            const status = STATUS_CONFIG[adapter.status];
            const initial = adapter.type.charAt(0).toUpperCase();

            return (
              <div
                key={adapter.id}
                className="rounded-xl border border-white/10 bg-white/5 shadow-sm"
              >
                <div className="p-5">
                  {/* Top row: logo + name + status */}
                  <div className="mb-4 flex items-start gap-3">
                    {/* Logo placeholder */}
                    <div
                      className="flex h-11 w-11 shrink-0 items-center justify-center rounded-full text-lg font-bold text-white"
                      style={{ backgroundColor: adapter.brandColor }}
                    >
                      {initial}
                    </div>

                    <div className="min-w-0 flex-1">
                      <h3 className="truncate text-base font-semibold text-white">
                        {adapter.name}
                      </h3>
                      <p className="text-sm text-slate-400">
                        {adapter.type} &middot; {adapter.location}
                      </p>
                    </div>

                    {/* Status badge */}
                    <span
                      className={`inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-xs font-semibold ${status.bg} ${status.text}`}
                    >
                      <span
                        className={`inline-block h-2 w-2 rounded-full ${status.dot}`}
                      />
                      {status.label}
                    </span>
                  </div>

                  {/* Capabilities */}
                  <div className="mb-4 flex flex-wrap gap-1.5">
                    {adapter.capabilities.map((cap) => (
                      <span
                        key={cap}
                        className="rounded-md bg-gray-100 px-2 py-0.5 text-xs font-medium text-slate-300"
                      >
                        {cap}
                      </span>
                    ))}
                  </div>

                  {/* Last sync */}
                  <p className="mb-4 flex items-center gap-1 text-xs text-slate-300">
                    <Clock className="h-3.5 w-3.5" />
                    Last sync: {formatTimestamp(adapter.lastSync)}
                  </p>

                  {/* Actions */}
                  <div className="flex items-center gap-3 border-t border-white/5 pt-4">
                    <button
                      onClick={() => handleViewDetails(adapter.id)}
                      className="inline-flex items-center gap-1.5 rounded-lg bg-white px-4 py-1.5 text-sm font-medium text-slate-200 ring-1 ring-gray-200 transition-colors hover:bg-white/5"
                    >
                      <Eye className="h-4 w-4" />
                      View Details
                    </button>
                    <button
                      onClick={() => handleTogglePause(adapter.id)}
                      className={`inline-flex items-center gap-1.5 rounded-lg px-4 py-1.5 text-sm font-medium transition-colors ${
                        adapter.status === 'paused'
                          ? 'bg-[#F97316] text-white hover:bg-[#EA580C]'
                          : 'bg-white text-slate-200 ring-1 ring-gray-200 hover:bg-white/5'
                      }`}
                    >
                      {adapter.status === 'paused' ? (
                        <>
                          <Play className="h-4 w-4" />
                          Resume
                        </>
                      ) : (
                        <>
                          <Pause className="h-4 w-4" />
                          Pause
                        </>
                      )}
                    </button>
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}
