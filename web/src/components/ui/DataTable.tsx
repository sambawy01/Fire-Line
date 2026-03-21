import { useState, type ReactNode } from 'react';
import LoadingSpinner from './LoadingSpinner';
import EmptyState from './EmptyState';

export interface Column<T> {
  key: string;
  header: string;
  render?: (row: T) => ReactNode;
  sortable?: boolean;
  align?: 'left' | 'center' | 'right';
}

interface DataTableProps<T> {
  columns: Column<T>[];
  data: T[];
  keyExtractor: (row: T) => string;
  isLoading?: boolean;
  emptyTitle?: string;
  emptyDescription?: string;
  onRowClick?: (row: T) => void;
  rowClassName?: (row: T) => string;
  /** If provided, the row with this key will render expandedRowContent below it. */
  expandedRowId?: string;
  /** Render a node below the row when expandedRowId matches. */
  renderExpanded?: (row: T) => ReactNode;
}

export default function DataTable<T>({
  columns,
  data,
  keyExtractor,
  isLoading = false,
  emptyTitle = 'No data',
  emptyDescription,
  onRowClick,
  rowClassName,
  expandedRowId,
  renderExpanded,
}: DataTableProps<T>) {
  const [sortKey, setSortKey] = useState<string | null>(null);
  const [sortAsc, setSortAsc] = useState(true);

  if (isLoading) {
    return (
      <div className="bg-white/5 rounded-xl border border-white/10 p-12 flex justify-center">
        <LoadingSpinner size="lg" />
      </div>
    );
  }

  if (data.length === 0) {
    return <EmptyState title={emptyTitle} description={emptyDescription} />;
  }

  const sorted = sortKey
    ? [...data].sort((a, b) => {
        const aVal = (a as Record<string, unknown>)[sortKey];
        const bVal = (b as Record<string, unknown>)[sortKey];
        if (typeof aVal === 'number' && typeof bVal === 'number') {
          return sortAsc ? aVal - bVal : bVal - aVal;
        }
        return sortAsc
          ? String(aVal).localeCompare(String(bVal))
          : String(bVal).localeCompare(String(aVal));
      })
    : data;

  function handleSort(key: string) {
    if (sortKey === key) {
      setSortAsc(!sortAsc);
    } else {
      setSortKey(key);
      setSortAsc(true);
    }
  }

  const alignClass = (align?: string) =>
    align === 'right' ? 'text-right' : align === 'center' ? 'text-center' : 'text-left';

  return (
    <div className="bg-white/5 rounded-xl border border-white/10 overflow-hidden">
      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="bg-white/5 text-slate-400 uppercase tracking-wider text-xs">
              {columns.map((col) => (
                <th
                  key={col.key}
                  className={`px-6 py-3 font-medium ${alignClass(col.align)} ${col.sortable ? 'cursor-pointer select-none hover:text-slate-200' : ''}`}
                  onClick={col.sortable ? () => handleSort(col.key) : undefined}
                >
                  {col.header}
                  {col.sortable && sortKey === col.key && (
                    <span className="ml-1">{sortAsc ? '↑' : '↓'}</span>
                  )}
                </th>
              ))}
            </tr>
          </thead>
          <tbody className="divide-y divide-white/5">
            {sorted.map((row) => {
              const key = keyExtractor(row);
              const isExpanded = expandedRowId === key;
              return (
                <>
                  <tr
                    key={key}
                    className={`hover:bg-white/5 transition-colors ${onRowClick ? 'cursor-pointer' : ''} ${isExpanded ? 'bg-white/[0.07]' : ''} ${rowClassName ? rowClassName(row) : ''}`}
                    onClick={onRowClick ? () => onRowClick(row) : undefined}
                  >
                    {columns.map((col) => (
                      <td key={col.key} className={`px-6 py-3 ${alignClass(col.align)} text-slate-300`}>
                        {col.render
                          ? col.render(row)
                          : String((row as Record<string, unknown>)[col.key] ?? '')}
                      </td>
                    ))}
                  </tr>
                  {isExpanded && renderExpanded && (
                    <tr key={`${key}-expanded`}>
                      <td colSpan={columns.length} className="p-0">
                        {renderExpanded(row)}
                      </td>
                    </tr>
                  )}
                </>
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );
}
