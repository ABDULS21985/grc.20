'use client';

import {
  ResponsiveContainer,
  RadarChart,
  PolarGrid,
  PolarAngleAxis,
  PolarRadiusAxis,
  Radar,
  Tooltip,
  Legend,
} from 'recharts';

// ── Types ────────────────────────────────────────────────
interface RadarDataPoint {
  name: string;
  score: number;
  target?: number;
}

interface ComplianceRadarProps {
  data: RadarDataPoint[];
}

// ── Component ────────────────────────────────────────────
export function ComplianceRadar({ data }: ComplianceRadarProps) {
  const hasTarget = data.some((d) => d.target !== undefined);

  if (data.length === 0) {
    return (
      <div className="flex h-64 items-center justify-center text-sm text-gray-500">
        No framework scores available.
      </div>
    );
  }

  return (
    <ResponsiveContainer width="100%" height={360}>
      <RadarChart cx="50%" cy="50%" outerRadius="75%" data={data}>
        <PolarGrid stroke="#e5e7eb" />
        <PolarAngleAxis
          dataKey="name"
          tick={{ fill: '#6b7280', fontSize: 12 }}
        />
        <PolarRadiusAxis
          angle={90}
          domain={[0, 100]}
          tick={{ fill: '#9ca3af', fontSize: 10 }}
          tickCount={6}
        />

        {/* Target area (shown behind actual scores) */}
        {hasTarget && (
          <Radar
            name="Target"
            dataKey="target"
            stroke="#c7d2fe"
            fill="#c7d2fe"
            fillOpacity={0.2}
            strokeDasharray="4 4"
          />
        )}

        {/* Actual scores */}
        <Radar
          name="Score"
          dataKey="score"
          stroke="#4f46e5"
          fill="#4f46e5"
          fillOpacity={0.25}
          strokeWidth={2}
        />

        <Tooltip
          contentStyle={{
            backgroundColor: '#fff',
            border: '1px solid #e5e7eb',
            borderRadius: '0.75rem',
            fontSize: '0.875rem',
          }}
          formatter={(value: number) => [`${value}%`, '']}
        />
        <Legend
          wrapperStyle={{ fontSize: '0.75rem', paddingTop: '0.5rem' }}
        />
      </RadarChart>
    </ResponsiveContainer>
  );
}
