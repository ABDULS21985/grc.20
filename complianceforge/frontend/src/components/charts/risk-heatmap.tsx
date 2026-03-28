'use client';

import { cn } from '@/lib/utils';

// ── Types ────────────────────────────────────────────────
interface HeatmapCell {
  likelihood: number;
  impact: number;
  count: number;
  risks?: { id: string; title: string }[];
}

interface RiskHeatmapProps {
  data: HeatmapCell[];
  onClick?: (cell: HeatmapCell) => void;
}

// ── Helpers ──────────────────────────────────────────────
const AXIS_LABELS = ['Very Low', 'Low', 'Medium', 'High', 'Critical'];

/**
 * Risk score = likelihood * impact.
 * Color bands:
 *   1-3  = green   (low)
 *   4-8  = yellow  (medium)
 *   9-14 = orange  (high)
 *   15-25 = red    (critical)
 */
function getCellColor(score: number): string {
  if (score >= 15) return 'bg-red-500 hover:bg-red-600 text-white';
  if (score >= 9) return 'bg-orange-400 hover:bg-orange-500 text-white';
  if (score >= 4) return 'bg-yellow-300 hover:bg-yellow-400 text-yellow-900';
  return 'bg-green-300 hover:bg-green-400 text-green-900';
}

function getCellBg(score: number): string {
  if (score >= 15) return 'bg-red-100';
  if (score >= 9) return 'bg-orange-100';
  if (score >= 4) return 'bg-yellow-100';
  return 'bg-green-100';
}

// ── Component ────────────────────────────────────────────
export function RiskHeatmap({ data, onClick }: RiskHeatmapProps) {
  // Build a lookup: key "likelihood-impact" -> cell
  const cellMap = new Map<string, HeatmapCell>();
  for (const cell of data) {
    cellMap.set(`${cell.likelihood}-${cell.impact}`, cell);
  }

  return (
    <div className="w-full">
      <div className="flex">
        {/* Y-axis label */}
        <div className="flex w-8 shrink-0 items-center justify-center">
          <span className="-rotate-90 whitespace-nowrap text-xs font-medium uppercase tracking-wider text-gray-500">
            Impact
          </span>
        </div>

        <div className="flex-1">
          {/* Grid rows: impact 5 (top) to 1 (bottom) */}
          <div className="grid grid-cols-[auto_repeat(5,1fr)] gap-1">
            {/* Header spacer */}
            <div />
            {/* Empty top-right cells */}
            {Array.from({ length: 5 }).map((_, i) => (
              <div key={`spacer-${i}`} />
            ))}

            {[5, 4, 3, 2, 1].map((impact) => (
              <>
                {/* Y-axis tick */}
                <div
                  key={`label-${impact}`}
                  className="flex items-center justify-end pr-2 text-xs text-gray-500"
                >
                  {AXIS_LABELS[impact - 1]}
                </div>

                {/* Cells for this impact row */}
                {[1, 2, 3, 4, 5].map((likelihood) => {
                  const score = likelihood * impact;
                  const cell = cellMap.get(`${likelihood}-${impact}`);
                  const count = cell?.count ?? 0;

                  return (
                    <button
                      key={`${likelihood}-${impact}`}
                      onClick={
                        onClick && cell && count > 0
                          ? () => onClick(cell)
                          : undefined
                      }
                      className={cn(
                        'flex aspect-square items-center justify-center rounded-lg text-sm font-semibold transition-colors',
                        count > 0
                          ? getCellColor(score)
                          : getCellBg(score),
                        count > 0 && onClick && 'cursor-pointer',
                        count === 0 && 'cursor-default opacity-60',
                      )}
                      title={`Likelihood: ${likelihood}, Impact: ${impact}, Score: ${score}, Risks: ${count}`}
                    >
                      {count > 0 ? count : ''}
                    </button>
                  );
                })}
              </>
            ))}

            {/* X-axis labels */}
            <div />
            {AXIS_LABELS.map((label) => (
              <div
                key={label}
                className="text-center text-xs text-gray-500 pt-1"
              >
                {label}
              </div>
            ))}
          </div>

          {/* X-axis title */}
          <p className="mt-2 text-center text-xs font-medium uppercase tracking-wider text-gray-500">
            Likelihood
          </p>
        </div>
      </div>

      {/* Legend */}
      <div className="mt-4 flex flex-wrap items-center justify-center gap-4 text-xs text-gray-600">
        <div className="flex items-center gap-1.5">
          <span className="inline-block h-3 w-3 rounded bg-green-300" />
          Low (1-3)
        </div>
        <div className="flex items-center gap-1.5">
          <span className="inline-block h-3 w-3 rounded bg-yellow-300" />
          Medium (4-8)
        </div>
        <div className="flex items-center gap-1.5">
          <span className="inline-block h-3 w-3 rounded bg-orange-400" />
          High (9-14)
        </div>
        <div className="flex items-center gap-1.5">
          <span className="inline-block h-3 w-3 rounded bg-red-500" />
          Critical (15-25)
        </div>
      </div>
    </div>
  );
}
