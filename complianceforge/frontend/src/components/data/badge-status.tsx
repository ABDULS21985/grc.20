import { cn } from '@/lib/utils';

// ── Types ────────────────────────────────────────────────
type BadgeType = 'severity' | 'risk' | 'status' | 'policy';

interface BadgeStatusProps {
  status: string;
  type?: BadgeType;
}

// ── Class Mappings ───────────────────────────────────────
const SEVERITY_MAP: Record<string, string> = {
  critical: 'badge-critical',
  high: 'badge-high',
  medium: 'badge-medium',
  low: 'badge-low',
  very_low: 'badge-low',
  informational: 'badge-info',
};

const RISK_MAP: Record<string, string> = {
  critical: 'badge-critical',
  high: 'badge-high',
  medium: 'badge-medium',
  low: 'badge-low',
  very_low: 'badge-low',
  identified: 'badge-info',
  assessed: 'badge-medium',
  treated: 'badge-low',
  accepted: 'bg-indigo-100 text-indigo-700',
  closed: 'bg-gray-100 text-gray-700',
  monitoring: 'badge-info',
};

const STATUS_MAP: Record<string, string> = {
  open: 'badge-high',
  investigating: 'badge-medium',
  contained: 'badge-info',
  resolved: 'badge-low',
  closed: 'bg-gray-100 text-gray-700',
  planned: 'badge-info',
  in_progress: 'badge-medium',
  completed: 'badge-low',
  cancelled: 'bg-gray-100 text-gray-700',
  active: 'badge-low',
  inactive: 'bg-gray-100 text-gray-700',
  locked: 'badge-critical',
};

const POLICY_MAP: Record<string, string> = {
  draft: 'bg-gray-100 text-gray-700',
  under_review: 'badge-medium',
  pending_approval: 'badge-info',
  approved: 'badge-low',
  published: 'bg-indigo-100 text-indigo-700',
  archived: 'bg-gray-100 text-gray-700',
  retired: 'bg-gray-100 text-gray-500',
};

function getTypeMap(type: BadgeType): Record<string, string> {
  switch (type) {
    case 'severity':
      return SEVERITY_MAP;
    case 'risk':
      return RISK_MAP;
    case 'status':
      return STATUS_MAP;
    case 'policy':
      return POLICY_MAP;
  }
}

function formatLabel(status: string): string {
  return status
    .replace(/_/g, ' ')
    .replace(/\b\w/g, (c) => c.toUpperCase());
}

// ── Component ────────────────────────────────────────────
export function BadgeStatus({ status, type = 'severity' }: BadgeStatusProps) {
  const map = getTypeMap(type);
  const colorClass = map[status.toLowerCase()] ?? 'badge-info';

  return (
    <span className={cn('badge', colorClass)}>
      {formatLabel(status)}
    </span>
  );
}
