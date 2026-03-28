'use client';

import { ArrowUp, ArrowDown, ArrowUpDown } from 'lucide-react';
import { cn } from '@/lib/utils';
import { EmptyState } from './empty-state';

// ── Types ────────────────────────────────────────────────
export interface Column<T> {
  key: string;
  label: string;
  sortable?: boolean;
  render?: (row: T) => React.ReactNode;
}

type SortDir = 'asc' | 'desc';

interface DataTableProps<T> {
  columns: Column<T>[];
  data: T[];
  loading?: boolean;
  emptyMessage?: string;
  onSort?: (key: string, dir: SortDir) => void;
  sortBy?: string;
  sortDir?: SortDir;
  onRowClick?: (row: T) => void;
}

// ── Skeleton Row ─────────────────────────────────────────
function SkeletonRow({ columns }: { columns: number }) {
  return (
    <tr>
      {Array.from({ length: columns }).map((_, i) => (
        <td key={i} className="px-6 py-4">
          <div className="h-4 w-3/4 animate-pulse rounded bg-gray-200" />
        </td>
      ))}
    </tr>
  );
}

// ── Sort Icon ────────────────────────────────────────────
function SortIcon({ active, dir }: { active: boolean; dir?: SortDir }) {
  if (!active) return <ArrowUpDown className="ml-1 inline h-3.5 w-3.5 text-gray-400" />;
  if (dir === 'asc') return <ArrowUp className="ml-1 inline h-3.5 w-3.5 text-indigo-600" />;
  return <ArrowDown className="ml-1 inline h-3.5 w-3.5 text-indigo-600" />;
}

// ── Component ────────────────────────────────────────────
export function DataTable<T extends Record<string, unknown>>({
  columns,
  data,
  loading = false,
  emptyMessage = 'No records found.',
  onSort,
  sortBy,
  sortDir,
  onRowClick,
}: DataTableProps<T>) {
  function handleSort(key: string) {
    if (!onSort) return;
    const newDir: SortDir =
      sortBy === key && sortDir === 'asc' ? 'desc' : 'asc';
    onSort(key, newDir);
  }

  // Loading state
  if (loading) {
    return (
      <div className="overflow-hidden rounded-xl border border-gray-200 bg-white">
        <table className="min-w-full divide-y divide-gray-200">
          <thead>
            <tr>
              {columns.map((col) => (
                <th key={col.key} className="table-header px-6 py-3">
                  {col.label}
                </th>
              ))}
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-100">
            {Array.from({ length: 5 }).map((_, i) => (
              <SkeletonRow key={i} columns={columns.length} />
            ))}
          </tbody>
        </table>
      </div>
    );
  }

  // Empty state
  if (data.length === 0) {
    return (
      <div className="overflow-hidden rounded-xl border border-gray-200 bg-white">
        <table className="min-w-full">
          <thead>
            <tr>
              {columns.map((col) => (
                <th key={col.key} className="table-header px-6 py-3">
                  {col.label}
                </th>
              ))}
            </tr>
          </thead>
        </table>
        <EmptyState title="Nothing here yet" description={emptyMessage} />
      </div>
    );
  }

  return (
    <div className="overflow-hidden rounded-xl border border-gray-200 bg-white">
      <div className="overflow-x-auto">
        <table className="min-w-full divide-y divide-gray-200">
          <thead>
            <tr>
              {columns.map((col) => (
                <th
                  key={col.key}
                  className={cn(
                    'table-header px-6 py-3',
                    col.sortable && 'cursor-pointer select-none hover:text-gray-700',
                  )}
                  onClick={col.sortable ? () => handleSort(col.key) : undefined}
                >
                  <span className="inline-flex items-center">
                    {col.label}
                    {col.sortable && (
                      <SortIcon
                        active={sortBy === col.key}
                        dir={sortBy === col.key ? sortDir : undefined}
                      />
                    )}
                  </span>
                </th>
              ))}
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-100 bg-white">
            {data.map((row, rowIdx) => (
              <tr
                key={rowIdx}
                className={cn(
                  'transition-colors',
                  onRowClick && 'cursor-pointer hover:bg-gray-50',
                )}
                onClick={onRowClick ? () => onRowClick(row) : undefined}
              >
                {columns.map((col) => (
                  <td key={col.key} className="whitespace-nowrap px-6 py-4 text-sm text-gray-700">
                    {col.render
                      ? col.render(row)
                      : (row[col.key] as React.ReactNode) ?? '-'}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
