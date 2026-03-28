import { cn } from '@/lib/utils';
import { TrendingUp, TrendingDown, Minus } from 'lucide-react';

// ── Types ────────────────────────────────────────────────
type StatColor = 'green' | 'amber' | 'red' | 'blue' | 'indigo';
type TrendDirection = 'up' | 'down' | 'flat';

interface StatCardProps {
  label: string;
  value: string;
  subtitle?: string;
  color?: StatColor;
  trend?: TrendDirection;
  icon?: React.ReactNode;
}

// ── Color Map ────────────────────────────────────────────
const COLOR_STYLES: Record<StatColor, { border: string; bg: string; iconBg: string; iconText: string }> = {
  green: {
    border: 'border-l-green-500',
    bg: 'bg-green-50',
    iconBg: 'bg-green-100',
    iconText: 'text-green-600',
  },
  amber: {
    border: 'border-l-amber-500',
    bg: 'bg-amber-50',
    iconBg: 'bg-amber-100',
    iconText: 'text-amber-600',
  },
  red: {
    border: 'border-l-red-500',
    bg: 'bg-red-50',
    iconBg: 'bg-red-100',
    iconText: 'text-red-600',
  },
  blue: {
    border: 'border-l-blue-500',
    bg: 'bg-blue-50',
    iconBg: 'bg-blue-100',
    iconText: 'text-blue-600',
  },
  indigo: {
    border: 'border-l-indigo-500',
    bg: 'bg-indigo-50',
    iconBg: 'bg-indigo-100',
    iconText: 'text-indigo-600',
  },
};

const TREND_CONFIG: Record<TrendDirection, { icon: typeof TrendingUp; color: string }> = {
  up: { icon: TrendingUp, color: 'text-green-600' },
  down: { icon: TrendingDown, color: 'text-red-600' },
  flat: { icon: Minus, color: 'text-gray-400' },
};

// ── Component ────────────────────────────────────────────
export function StatCard({
  label,
  value,
  subtitle,
  color = 'indigo',
  trend,
  icon,
}: StatCardProps) {
  const styles = COLOR_STYLES[color];
  const TrendIcon = trend ? TREND_CONFIG[trend].icon : null;
  const trendColor = trend ? TREND_CONFIG[trend].color : '';

  return (
    <div
      className={cn(
        'card flex items-start gap-4 border-l-4',
        styles.border,
      )}
    >
      {/* Icon */}
      {icon && (
        <div
          className={cn(
            'flex h-11 w-11 shrink-0 items-center justify-center rounded-lg',
            styles.iconBg,
            styles.iconText,
          )}
        >
          {icon}
        </div>
      )}

      {/* Content */}
      <div className="min-w-0 flex-1">
        <p className="text-xs font-medium uppercase tracking-wider text-gray-500">
          {label}
        </p>
        <div className="mt-1 flex items-baseline gap-2">
          <p className="text-2xl font-bold text-gray-900">{value}</p>
          {TrendIcon && (
            <TrendIcon className={cn('h-4 w-4', trendColor)} />
          )}
        </div>
        {subtitle && (
          <p className="mt-0.5 text-xs text-gray-500">{subtitle}</p>
        )}
      </div>
    </div>
  );
}
