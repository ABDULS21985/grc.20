import { cn } from '@/lib/utils';

// ── Skeleton Primitive ───────────────────────────────────
function Bone({ className }: { className?: string }) {
  return (
    <div
      className={cn('animate-pulse rounded bg-gray-200', className)}
    />
  );
}

// ── Table Skeleton ───────────────────────────────────────
interface TableSkeletonProps {
  rows?: number;
  columns?: number;
}

export function TableSkeleton({ rows = 5, columns = 6 }: TableSkeletonProps) {
  return (
    <div className="overflow-hidden rounded-xl border border-gray-200 bg-white">
      {/* Header */}
      <div className="grid gap-4 border-b border-gray-200 bg-gray-50 px-6 py-3" style={{ gridTemplateColumns: `repeat(${columns}, 1fr)` }}>
        {Array.from({ length: columns }).map((_, i) => (
          <Bone key={`h-${i}`} className="h-3 w-20" />
        ))}
      </div>

      {/* Rows */}
      {Array.from({ length: rows }).map((_, rowIdx) => (
        <div
          key={`r-${rowIdx}`}
          className="grid gap-4 border-b border-gray-100 px-6 py-4 last:border-b-0"
          style={{ gridTemplateColumns: `repeat(${columns}, 1fr)` }}
        >
          {Array.from({ length: columns }).map((_, colIdx) => (
            <Bone
              key={`c-${rowIdx}-${colIdx}`}
              className={cn('h-4', colIdx === 0 ? 'w-32' : 'w-24')}
            />
          ))}
        </div>
      ))}
    </div>
  );
}

// ── Card Skeleton ────────────────────────────────────────
export function CardSkeleton() {
  return (
    <div className="card space-y-4">
      <Bone className="h-4 w-1/3" />
      <Bone className="h-3 w-full" />
      <Bone className="h-3 w-2/3" />
      <Bone className="h-8 w-24" />
    </div>
  );
}

// ── Stat Card Skeleton ───────────────────────────────────
export function StatCardSkeleton() {
  return (
    <div className="card flex items-center gap-4">
      <Bone className="h-12 w-12 rounded-lg" />
      <div className="flex-1 space-y-2">
        <Bone className="h-3 w-20" />
        <Bone className="h-6 w-16" />
        <Bone className="h-3 w-24" />
      </div>
    </div>
  );
}

// ── Full Page Skeleton ───────────────────────────────────
export function PageSkeleton() {
  return (
    <div className="space-y-6">
      {/* Stat cards row */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {Array.from({ length: 4 }).map((_, i) => (
          <StatCardSkeleton key={`stat-${i}`} />
        ))}
      </div>

      {/* Table */}
      <TableSkeleton rows={8} columns={6} />
    </div>
  );
}
