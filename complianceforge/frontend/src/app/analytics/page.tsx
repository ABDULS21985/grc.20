'use client';

import { useEffect, useState, useCallback } from 'react';
import api from '@/lib/api';

// ============================================================
// TYPES
// ============================================================

interface ComplianceTrend {
  id: string;
  framework_code: string;
  measurement_date: string;
  compliance_score: number;
  controls_implemented: number;
  controls_total: number;
  maturity_avg: number;
  score_change_7d: number;
  score_change_30d: number;
  score_change_90d: number;
  trend_direction: 'improving' | 'stable' | 'declining';
}

interface BreachPrediction {
  breach_probability_30d: number;
  breach_probability_90d: number;
  breach_probability_365d: number;
  confidence_level: number;
  risk_factors: RiskFactor[];
  model_version: string;
}

interface RiskFactor {
  name: string;
  weight: number;
  value: number;
  contribution: number;
  description: string;
}

interface PeerComparison {
  metrics: PeerMetric[];
  benchmark_period: string;
  sample_size: number;
}

interface PeerMetric {
  metric_name: string;
  org_value: number;
  percentile_25: number;
  percentile_50: number;
  percentile_75: number;
  percentile_90: number;
  percentile_position: number;
  sample_size: number;
}

interface TimeSeriesPoint {
  date: string;
  value: number;
  label: string;
}

interface DistributionEntry {
  category: string;
  count: number;
  percent: number;
}

interface Dashboard {
  id: string;
  name: string;
  description: string;
  layout: WidgetLayout[];
  is_default: boolean;
  is_shared: boolean;
  created_at: string;
  updated_at: string;
}

interface WidgetLayout {
  widget_type: string;
  metric: string;
  title: string;
  x: number;
  y: number;
  w: number;
  h: number;
  config: Record<string, unknown>;
}

interface WidgetTypeInfo {
  id: string;
  widget_type: string;
  name: string;
  description: string;
  available_metrics: string[];
  default_config: Record<string, unknown>;
  min_width: number;
  min_height: number;
}

// ============================================================
// CONSTANTS
// ============================================================

const TREND_COLORS: Record<string, { bg: string; text: string; arrow: string }> = {
  improving: { bg: 'bg-green-50', text: 'text-green-700', arrow: 'M5 10l7-7m0 0l7 7m-7-7v18' },
  stable:    { bg: 'bg-gray-50',  text: 'text-gray-600',  arrow: 'M5 12h14' },
  declining: { bg: 'bg-red-50',   text: 'text-red-700',   arrow: 'M19 14l-7 7m0 0l-7-7m7 7V3' },
};

const METRIC_LABELS: Record<string, string> = {
  compliance_score: 'Compliance Score',
  control_coverage: 'Control Coverage',
  avg_maturity: 'Avg Maturity',
  total_risks: 'Total Risks',
  open_incidents: 'Open Incidents',
  open_findings: 'Open Findings',
  avg_risk_score: 'Avg Risk Score',
  policies_due_review: 'Policies Due Review',
};

const PROB_THRESHOLDS = {
  low: 0.15,
  medium: 0.35,
  high: 0.60,
};

// ============================================================
// MAIN PAGE COMPONENT
// ============================================================

export default function AnalyticsPage() {
  const [tab, setTab] = useState<'overview' | 'trends' | 'predictions' | 'benchmarks' | 'dashboards'>('overview');
  const [loading, setLoading] = useState(true);

  // Overview data
  const [complianceTrends, setComplianceTrends] = useState<ComplianceTrend[]>([]);
  const [riskTimeSeries, setRiskTimeSeries] = useState<TimeSeriesPoint[]>([]);
  const [riskDistribution, setRiskDistribution] = useState<DistributionEntry[]>([]);

  // Prediction data
  const [breachPrediction, setBreachPrediction] = useState<BreachPrediction | null>(null);

  // Benchmark data
  const [peerComparison, setPeerComparison] = useState<PeerComparison | null>(null);

  // Dashboard data
  const [dashboards, setDashboards] = useState<Dashboard[]>([]);
  const [widgetTypes, setWidgetTypes] = useState<WidgetTypeInfo[]>([]);
  const [showCreateDashboard, setShowCreateDashboard] = useState(false);
  const [newDashboardName, setNewDashboardName] = useState('');
  const [newDashboardDesc, setNewDashboardDesc] = useState('');

  const loadOverview = useCallback(async () => {
    setLoading(true);
    try {
      const [trendsRes, riskRes, distRes] = await Promise.all([
        api.request<{ data: ComplianceTrend[] }>('/analytics/trends/compliance?months=12'),
        api.request<{ data: TimeSeriesPoint[] }>('/analytics/metrics/avg_risk_score?period=6m&granularity=daily'),
        api.request<{ data: DistributionEntry[] }>('/analytics/distribution/risks?group_by=severity'),
      ]);
      setComplianceTrends(trendsRes.data || []);
      setRiskTimeSeries(riskRes.data || []);
      setRiskDistribution(distRes.data || []);
    } catch { /* ignore */ }
    setLoading(false);
  }, []);

  const loadPredictions = useCallback(async () => {
    setLoading(true);
    try {
      const res = await api.request<{ data: BreachPrediction }>('/analytics/predictions/breach-probability');
      setBreachPrediction(res.data);
    } catch { /* ignore */ }
    setLoading(false);
  }, []);

  const loadBenchmarks = useCallback(async () => {
    setLoading(true);
    try {
      const res = await api.request<{ data: PeerComparison }>('/analytics/benchmarks');
      setPeerComparison(res.data);
    } catch { /* ignore */ }
    setLoading(false);
  }, []);

  const loadDashboards = useCallback(async () => {
    setLoading(true);
    try {
      const [dashRes, widgetRes] = await Promise.all([
        api.request<{ data: Dashboard[] }>('/analytics/dashboards'),
        api.request<{ data: WidgetTypeInfo[] }>('/analytics/widget-types'),
      ]);
      setDashboards(dashRes.data || []);
      setWidgetTypes(widgetRes.data || []);
    } catch { /* ignore */ }
    setLoading(false);
  }, []);

  useEffect(() => {
    if (tab === 'overview') loadOverview();
    else if (tab === 'trends') loadOverview();
    else if (tab === 'predictions') loadPredictions();
    else if (tab === 'benchmarks') loadBenchmarks();
    else if (tab === 'dashboards') loadDashboards();
  }, [tab, loadOverview, loadPredictions, loadBenchmarks, loadDashboards]);

  async function createDashboard() {
    if (!newDashboardName.trim()) return;
    try {
      await api.request('/analytics/dashboards', {
        method: 'POST',
        body: {
          name: newDashboardName,
          description: newDashboardDesc,
          layout: [],
          is_default: false,
          is_shared: false,
        },
      });
      setShowCreateDashboard(false);
      setNewDashboardName('');
      setNewDashboardDesc('');
      loadDashboards();
    } catch { /* ignore */ }
  }

  async function deleteDashboard(id: string) {
    if (!confirm('Delete this dashboard?')) return;
    try {
      await api.request(`/analytics/dashboards/${id}`, { method: 'DELETE' });
      loadDashboards();
    } catch { /* ignore */ }
  }

  async function handleExport(format: 'csv' | 'json') {
    try {
      const res = await api.request<Blob>('/analytics/export', {
        method: 'POST',
        body: {
          format,
          start_date: new Date(Date.now() - 365 * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
          end_date: new Date().toISOString().split('T')[0],
          granularity: 'daily',
        },
      });
      // Trigger download if we have data
      if (res) {
        const blob = new Blob([JSON.stringify(res)], { type: format === 'csv' ? 'text/csv' : 'application/json' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `analytics_export.${format}`;
        a.click();
        URL.revokeObjectURL(url);
      }
    } catch { /* ignore */ }
  }

  const tabs = [
    { key: 'overview' as const, label: 'Overview' },
    { key: 'trends' as const, label: 'Compliance Trends' },
    { key: 'predictions' as const, label: 'Risk Predictions' },
    { key: 'benchmarks' as const, label: 'Peer Benchmarks' },
    { key: 'dashboards' as const, label: 'Custom Dashboards' },
  ];

  return (
    <div>
      {/* Header */}
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Advanced Analytics</h1>
          <p className="text-gray-500 mt-1">Predictive risk scoring, compliance trends, and BI dashboards</p>
        </div>
        <div className="flex gap-2">
          <button
            onClick={() => handleExport('csv')}
            className="rounded-lg border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
          >
            Export CSV
          </button>
          <button
            onClick={() => handleExport('json')}
            className="rounded-lg border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
          >
            Export JSON
          </button>
        </div>
      </div>

      {/* Tabs */}
      <div className="mb-6 border-b border-gray-200">
        <nav className="flex gap-6">
          {tabs.map(t => (
            <button
              key={t.key}
              onClick={() => setTab(t.key)}
              className={`pb-3 text-sm font-medium border-b-2 transition-colors ${
                tab === t.key
                  ? 'border-blue-600 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              {t.label}
            </button>
          ))}
        </nav>
      </div>

      {loading && (
        <div className="flex items-center justify-center h-64">
          <p className="text-gray-500">Loading analytics data...</p>
        </div>
      )}

      {/* OVERVIEW TAB */}
      {!loading && tab === 'overview' && (
        <div className="space-y-6">
          {/* KPI Cards */}
          <KPICardGrid trends={complianceTrends} riskDistribution={riskDistribution} />

          <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
            {/* Compliance Score by Framework */}
            <div className="rounded-xl border border-gray-200 bg-white p-6">
              <h2 className="text-lg font-semibold text-gray-900 mb-4">Framework Compliance Scores</h2>
              <ComplianceScoreList trends={getLatestTrendPerFramework(complianceTrends)} />
            </div>

            {/* Risk Distribution */}
            <div className="rounded-xl border border-gray-200 bg-white p-6">
              <h2 className="text-lg font-semibold text-gray-900 mb-4">Risk Severity Distribution</h2>
              <DistributionBars entries={riskDistribution} />
            </div>
          </div>

          {/* Risk Score Timeline */}
          <div className="rounded-xl border border-gray-200 bg-white p-6">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">Average Risk Score (6 Months)</h2>
            <TimeSeriesChart points={riskTimeSeries} />
          </div>
        </div>
      )}

      {/* TRENDS TAB */}
      {!loading && tab === 'trends' && (
        <div className="space-y-6">
          <div className="rounded-xl border border-gray-200 bg-white p-6">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">Compliance Score Trends (12 Months)</h2>
            {complianceTrends.length === 0 ? (
              <p className="text-gray-500 text-sm">No trend data available yet. Trends are calculated from daily snapshots.</p>
            ) : (
              <ComplianceTrendTable trends={complianceTrends} />
            )}
          </div>
        </div>
      )}

      {/* PREDICTIONS TAB */}
      {!loading && tab === 'predictions' && (
        <div className="space-y-6">
          {breachPrediction ? (
            <>
              <BreachProbabilityCard prediction={breachPrediction} />
              <RiskFactorBreakdown factors={breachPrediction.risk_factors} />
            </>
          ) : (
            <div className="rounded-xl border border-gray-200 bg-white p-6">
              <p className="text-gray-500 text-sm">No prediction data available. Ensure daily snapshots are running.</p>
            </div>
          )}
        </div>
      )}

      {/* BENCHMARKS TAB */}
      {!loading && tab === 'benchmarks' && (
        <div className="space-y-6">
          {peerComparison && peerComparison.metrics && peerComparison.metrics.length > 0 ? (
            <>
              <div className="rounded-xl border border-gray-200 bg-white p-6">
                <div className="flex items-center justify-between mb-4">
                  <h2 className="text-lg font-semibold text-gray-900">Peer Comparison</h2>
                  <span className="text-sm text-gray-500">
                    Based on {peerComparison.sample_size} anonymized organizations | Period: {peerComparison.benchmark_period}
                  </span>
                </div>
                <PeerComparisonRadar metrics={peerComparison.metrics} />
              </div>
              <div className="rounded-xl border border-gray-200 bg-white p-6">
                <h2 className="text-lg font-semibold text-gray-900 mb-4">Benchmark Detail</h2>
                <BenchmarkTable metrics={peerComparison.metrics} />
              </div>
            </>
          ) : (
            <div className="rounded-xl border border-gray-200 bg-white p-6">
              <p className="text-gray-500 text-sm">
                Peer benchmarks require at least 5 organizations. Benchmark data is anonymized and calculated monthly.
              </p>
            </div>
          )}
        </div>
      )}

      {/* DASHBOARDS TAB */}
      {!loading && tab === 'dashboards' && (
        <div className="space-y-6">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold text-gray-900">Custom Dashboards</h2>
            <button
              onClick={() => setShowCreateDashboard(true)}
              className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
            >
              Create Dashboard
            </button>
          </div>

          {/* Create Dashboard Form */}
          {showCreateDashboard && (
            <div className="rounded-xl border border-blue-200 bg-blue-50 p-6">
              <h3 className="text-md font-semibold text-gray-900 mb-4">New Dashboard</h3>
              <div className="space-y-3">
                <input
                  type="text"
                  placeholder="Dashboard name"
                  value={newDashboardName}
                  onChange={e => setNewDashboardName(e.target.value)}
                  className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm"
                />
                <textarea
                  placeholder="Description (optional)"
                  value={newDashboardDesc}
                  onChange={e => setNewDashboardDesc(e.target.value)}
                  className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm"
                  rows={2}
                />
                <div className="flex gap-2">
                  <button
                    onClick={createDashboard}
                    className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
                  >
                    Create
                  </button>
                  <button
                    onClick={() => setShowCreateDashboard(false)}
                    className="rounded-lg border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
                  >
                    Cancel
                  </button>
                </div>
              </div>
            </div>
          )}

          {/* Dashboard List */}
          {dashboards.length === 0 ? (
            <div className="rounded-xl border border-gray-200 bg-white p-6">
              <p className="text-gray-500 text-sm">No custom dashboards yet. Create one to get started.</p>
            </div>
          ) : (
            <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
              {dashboards.map(d => (
                <div key={d.id} className="rounded-xl border border-gray-200 bg-white p-5 hover:shadow-md transition-shadow">
                  <div className="flex items-start justify-between">
                    <div>
                      <h3 className="font-semibold text-gray-900">{d.name}</h3>
                      {d.description && <p className="text-sm text-gray-500 mt-1">{d.description}</p>}
                    </div>
                    <div className="flex gap-1">
                      {d.is_default && (
                        <span className="inline-flex items-center rounded-full bg-blue-100 px-2 py-0.5 text-xs font-medium text-blue-700">
                          Default
                        </span>
                      )}
                      {d.is_shared && (
                        <span className="inline-flex items-center rounded-full bg-green-100 px-2 py-0.5 text-xs font-medium text-green-700">
                          Shared
                        </span>
                      )}
                    </div>
                  </div>
                  <div className="mt-3 flex items-center justify-between text-xs text-gray-400">
                    <span>{(d.layout || []).length} widgets</span>
                    <span>Updated {new Date(d.updated_at).toLocaleDateString()}</span>
                  </div>
                  <div className="mt-3 flex gap-2">
                    <button className="text-sm text-blue-600 hover:text-blue-800 font-medium">Edit</button>
                    <button
                      onClick={() => deleteDashboard(d.id)}
                      className="text-sm text-red-600 hover:text-red-800 font-medium"
                    >
                      Delete
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}

          {/* Widget Types Catalog */}
          <div className="rounded-xl border border-gray-200 bg-white p-6">
            <h3 className="text-md font-semibold text-gray-900 mb-4">Available Widget Types</h3>
            <div className="grid grid-cols-2 gap-3 md:grid-cols-3 lg:grid-cols-5">
              {widgetTypes.map(wt => (
                <div key={wt.id} className="rounded-lg border border-gray-100 bg-gray-50 p-3 text-center">
                  <WidgetIcon type={wt.widget_type} />
                  <p className="text-sm font-medium text-gray-800 mt-2">{wt.name}</p>
                  <p className="text-xs text-gray-500 mt-1">Min: {wt.min_width}x{wt.min_height}</p>
                </div>
              ))}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

// ============================================================
// SUB-COMPONENTS
// ============================================================

function KPICardGrid({ trends, riskDistribution }: { trends: ComplianceTrend[]; riskDistribution: DistributionEntry[] }) {
  const latest = getLatestTrendPerFramework(trends);
  const avgScore = latest.length > 0
    ? latest.reduce((sum, t) => sum + t.compliance_score, 0) / latest.length
    : 0;
  const totalRisks = riskDistribution.reduce((sum, e) => sum + e.count, 0);
  const criticalRisks = riskDistribution.find(e => e.category === 'critical')?.count || 0;
  const avgMaturity = latest.length > 0
    ? latest.reduce((sum, t) => sum + t.maturity_avg, 0) / latest.length
    : 0;

  return (
    <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-5">
      <KPICard
        label="Avg Compliance"
        value={`${avgScore.toFixed(1)}%`}
        color={avgScore >= 80 ? 'green' : avgScore >= 60 ? 'amber' : 'red'}
      />
      <KPICard
        label="Frameworks"
        value={String(latest.length)}
        color="blue"
      />
      <KPICard
        label="Total Risks"
        value={String(totalRisks)}
        subtitle={`${criticalRisks} critical`}
        color={criticalRisks > 0 ? 'red' : 'green'}
      />
      <KPICard
        label="Avg Maturity"
        value={avgMaturity.toFixed(2)}
        color={avgMaturity >= 3 ? 'green' : avgMaturity >= 2 ? 'amber' : 'red'}
      />
      <KPICard
        label="Trend"
        value={latest.length > 0 ? (latest.filter(t => t.trend_direction === 'improving').length > latest.length / 2 ? 'Improving' : 'Stable') : 'N/A'}
        color={latest.filter(t => t.trend_direction === 'improving').length > latest.length / 2 ? 'green' : 'gray'}
      />
    </div>
  );
}

function KPICard({ label, value, subtitle, color = 'gray' }: {
  label: string; value: string; subtitle?: string; color?: string;
}) {
  const colors: Record<string, string> = {
    green: 'border-green-200 bg-green-50',
    amber: 'border-amber-200 bg-amber-50',
    red: 'border-red-200 bg-red-50',
    blue: 'border-blue-200 bg-blue-50',
    gray: 'border-gray-200 bg-gray-50',
  };

  return (
    <div className={`rounded-xl border p-4 ${colors[color] || colors.gray}`}>
      <p className="text-xs font-medium text-gray-500 uppercase tracking-wide">{label}</p>
      <p className="text-2xl font-bold text-gray-900 mt-1">{value}</p>
      {subtitle && <p className="text-xs text-gray-500 mt-0.5">{subtitle}</p>}
    </div>
  );
}

function ComplianceScoreList({ trends }: { trends: ComplianceTrend[] }) {
  if (trends.length === 0) {
    return <p className="text-gray-500 text-sm">No compliance trend data available.</p>;
  }

  return (
    <div className="space-y-4">
      {trends.map(t => (
        <div key={t.framework_code}>
          <div className="flex items-center justify-between mb-1">
            <div className="flex items-center gap-2">
              <span className="text-sm font-medium text-gray-700">{t.framework_code}</span>
              <TrendBadge direction={t.trend_direction} change={t.score_change_30d} />
            </div>
            <span className={`text-sm font-bold ${getScoreTextColor(t.compliance_score)}`}>
              {t.compliance_score.toFixed(1)}%
            </span>
          </div>
          <div className="h-2.5 w-full rounded-full bg-gray-100">
            <div
              className={`h-2.5 rounded-full transition-all duration-500 ${getScoreBarColor(t.compliance_score)}`}
              style={{ width: `${Math.min(t.compliance_score, 100)}%` }}
            />
          </div>
          <div className="flex items-center justify-between mt-1 text-xs text-gray-400">
            <span>{t.controls_implemented}/{t.controls_total} controls</span>
            <span>Maturity: {t.maturity_avg.toFixed(2)}</span>
          </div>
        </div>
      ))}
    </div>
  );
}

function TrendBadge({ direction, change }: { direction: string; change: number }) {
  const style = TREND_COLORS[direction] || TREND_COLORS.stable;
  return (
    <span className={`inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium ${style.bg} ${style.text}`}>
      <svg className="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
        <path strokeLinecap="round" strokeLinejoin="round" d={style.arrow} />
      </svg>
      {change > 0 ? '+' : ''}{change.toFixed(1)}
    </span>
  );
}

function DistributionBars({ entries }: { entries: DistributionEntry[] }) {
  if (entries.length === 0) {
    return <p className="text-gray-500 text-sm">No distribution data available.</p>;
  }

  const maxCount = Math.max(...entries.map(e => e.count), 1);
  const severityColors: Record<string, string> = {
    critical: 'bg-red-500',
    high: 'bg-orange-500',
    medium: 'bg-amber-400',
    low: 'bg-green-500',
    very_low: 'bg-green-300',
  };

  return (
    <div className="space-y-3">
      {entries.map(e => (
        <div key={e.category} className="flex items-center gap-4">
          <span className="w-20 text-sm font-medium text-gray-700 capitalize">{e.category.replace('_', ' ')}</span>
          <div className="flex-1 h-6 rounded bg-gray-100 relative">
            <div
              className={`h-6 rounded transition-all duration-500 flex items-center px-2 ${severityColors[e.category] || 'bg-blue-400'}`}
              style={{ width: `${Math.max((e.count / maxCount) * 100, 8)}%` }}
            >
              <span className="text-xs font-bold text-white">{e.count}</span>
            </div>
          </div>
          <span className="text-xs text-gray-400 w-12 text-right">{e.percent.toFixed(0)}%</span>
        </div>
      ))}
    </div>
  );
}

function TimeSeriesChart({ points }: { points: TimeSeriesPoint[] }) {
  if (points.length === 0) {
    return <p className="text-gray-500 text-sm">No time series data available.</p>;
  }

  const maxVal = Math.max(...points.map(p => p.value), 1);
  const minVal = Math.min(...points.map(p => p.value), 0);
  const range = maxVal - minVal || 1;
  const chartHeight = 200;

  // Sample every Nth point if too many
  const step = Math.max(1, Math.floor(points.length / 60));
  const sampled = points.filter((_, i) => i % step === 0);

  const pathPoints = sampled.map((p, i) => {
    const x = (i / Math.max(sampled.length - 1, 1)) * 100;
    const y = chartHeight - ((p.value - minVal) / range) * chartHeight;
    return `${x},${y}`;
  });

  return (
    <div>
      <svg viewBox={`0 0 100 ${chartHeight}`} className="w-full h-48" preserveAspectRatio="none">
        {/* Grid lines */}
        {[0, 0.25, 0.5, 0.75, 1].map(f => (
          <line key={f} x1="0" y1={chartHeight * (1 - f)} x2="100" y2={chartHeight * (1 - f)}
            stroke="#e5e7eb" strokeWidth="0.3" />
        ))}
        {/* Area fill */}
        <polygon
          points={`0,${chartHeight} ${pathPoints.join(' ')} 100,${chartHeight}`}
          fill="url(#areaGradient)" opacity="0.3"
        />
        {/* Line */}
        <polyline points={pathPoints.join(' ')} fill="none" stroke="#3b82f6" strokeWidth="0.5" />
        <defs>
          <linearGradient id="areaGradient" x1="0%" y1="0%" x2="0%" y2="100%">
            <stop offset="0%" stopColor="#3b82f6" stopOpacity="0.4" />
            <stop offset="100%" stopColor="#3b82f6" stopOpacity="0" />
          </linearGradient>
        </defs>
      </svg>
      <div className="flex justify-between text-xs text-gray-400 mt-1">
        <span>{sampled[0]?.date ? new Date(sampled[0].date).toLocaleDateString() : ''}</span>
        <span>Value range: {minVal.toFixed(1)} - {maxVal.toFixed(1)}</span>
        <span>{sampled[sampled.length - 1]?.date ? new Date(sampled[sampled.length - 1].date).toLocaleDateString() : ''}</span>
      </div>
    </div>
  );
}

function ComplianceTrendTable({ trends }: { trends: ComplianceTrend[] }) {
  const frameworks = [...new Set(trends.map(t => t.framework_code))];

  return (
    <div className="overflow-x-auto">
      <table className="min-w-full divide-y divide-gray-200">
        <thead className="bg-gray-50">
          <tr>
            <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Framework</th>
            <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Date</th>
            <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">Score</th>
            <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">7d Change</th>
            <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">30d Change</th>
            <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">90d Change</th>
            <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase">Trend</th>
            <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">Controls</th>
            <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">Maturity</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-gray-200">
          {trends.slice(-20).map(t => (
            <tr key={t.id} className="hover:bg-gray-50">
              <td className="px-4 py-2 text-sm font-medium text-gray-900">{t.framework_code}</td>
              <td className="px-4 py-2 text-sm text-gray-600">{new Date(t.measurement_date).toLocaleDateString()}</td>
              <td className={`px-4 py-2 text-sm text-right font-semibold ${getScoreTextColor(t.compliance_score)}`}>
                {t.compliance_score.toFixed(1)}%
              </td>
              <td className={`px-4 py-2 text-sm text-right ${t.score_change_7d >= 0 ? 'text-green-600' : 'text-red-600'}`}>
                {t.score_change_7d > 0 ? '+' : ''}{t.score_change_7d.toFixed(1)}
              </td>
              <td className={`px-4 py-2 text-sm text-right ${t.score_change_30d >= 0 ? 'text-green-600' : 'text-red-600'}`}>
                {t.score_change_30d > 0 ? '+' : ''}{t.score_change_30d.toFixed(1)}
              </td>
              <td className={`px-4 py-2 text-sm text-right ${t.score_change_90d >= 0 ? 'text-green-600' : 'text-red-600'}`}>
                {t.score_change_90d > 0 ? '+' : ''}{t.score_change_90d.toFixed(1)}
              </td>
              <td className="px-4 py-2 text-center">
                <TrendBadge direction={t.trend_direction} change={t.score_change_30d} />
              </td>
              <td className="px-4 py-2 text-sm text-right text-gray-600">
                {t.controls_implemented}/{t.controls_total}
              </td>
              <td className="px-4 py-2 text-sm text-right text-gray-600">
                {t.maturity_avg.toFixed(2)}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function BreachProbabilityCard({ prediction }: { prediction: BreachPrediction }) {
  const probColor = (p: number) => {
    if (p >= PROB_THRESHOLDS.high) return 'text-red-600';
    if (p >= PROB_THRESHOLDS.medium) return 'text-amber-600';
    if (p >= PROB_THRESHOLDS.low) return 'text-amber-500';
    return 'text-green-600';
  };

  const probBgColor = (p: number) => {
    if (p >= PROB_THRESHOLDS.high) return 'bg-red-50 border-red-200';
    if (p >= PROB_THRESHOLDS.medium) return 'bg-amber-50 border-amber-200';
    if (p >= PROB_THRESHOLDS.low) return 'bg-amber-50 border-amber-200';
    return 'bg-green-50 border-green-200';
  };

  return (
    <div className="rounded-xl border border-gray-200 bg-white p-6">
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-lg font-semibold text-gray-900">Breach Probability Forecast</h2>
        <span className="text-xs text-gray-400">Model: {prediction.model_version} | Confidence: {(prediction.confidence_level * 100).toFixed(0)}%</span>
      </div>
      <div className="grid grid-cols-3 gap-4">
        {[
          { label: '30-Day', value: prediction.breach_probability_30d },
          { label: '90-Day', value: prediction.breach_probability_90d },
          { label: '365-Day', value: prediction.breach_probability_365d },
        ].map(item => (
          <div key={item.label} className={`rounded-xl border p-5 text-center ${probBgColor(item.value)}`}>
            <p className="text-xs font-medium text-gray-500 uppercase tracking-wide">{item.label}</p>
            <p className={`text-3xl font-bold mt-2 ${probColor(item.value)}`}>
              {(item.value * 100).toFixed(1)}%
            </p>
            <p className="text-xs text-gray-400 mt-1">probability</p>
          </div>
        ))}
      </div>
    </div>
  );
}

function RiskFactorBreakdown({ factors }: { factors: RiskFactor[] }) {
  if (!factors || factors.length === 0) return null;

  const maxContrib = Math.max(...factors.map(f => f.contribution), 0.01);

  return (
    <div className="rounded-xl border border-gray-200 bg-white p-6">
      <h2 className="text-lg font-semibold text-gray-900 mb-4">Risk Factor Contribution</h2>
      <div className="space-y-3">
        {factors.map(f => (
          <div key={f.name}>
            <div className="flex items-center justify-between mb-1">
              <span className="text-sm font-medium text-gray-700">{f.name}</span>
              <span className="text-xs text-gray-400">
                Weight: {(f.weight * 100).toFixed(0)}% | Contribution: {(f.contribution * 100).toFixed(1)}%
              </span>
            </div>
            <div className="h-3 w-full rounded-full bg-gray-100">
              <div
                className="h-3 rounded-full bg-red-400 transition-all duration-500"
                style={{ width: `${(f.contribution / maxContrib) * 100}%` }}
              />
            </div>
            <p className="text-xs text-gray-400 mt-0.5">{f.description}</p>
          </div>
        ))}
      </div>
    </div>
  );
}

function PeerComparisonRadar({ metrics }: { metrics: PeerMetric[] }) {
  // SVG-based radar chart
  const size = 300;
  const center = size / 2;
  const radius = size * 0.38;
  const n = metrics.length;

  if (n === 0) return null;

  const angleStep = (2 * Math.PI) / n;

  const getPoint = (index: number, value: number): { x: number; y: number } => {
    const angle = (index * angleStep) - Math.PI / 2;
    const r = (value / 100) * radius;
    return {
      x: center + r * Math.cos(angle),
      y: center + r * Math.sin(angle),
    };
  };

  const orgPoints = metrics.map((m, i) => getPoint(i, m.percentile_position));
  const p50Points = metrics.map((m, i) => getPoint(i, 50)); // median line
  const orgPath = orgPoints.map(p => `${p.x},${p.y}`).join(' ');
  const p50Path = p50Points.map(p => `${p.x},${p.y}`).join(' ');

  return (
    <div className="flex justify-center">
      <svg width={size} height={size} viewBox={`0 0 ${size} ${size}`}>
        {/* Grid circles */}
        {[25, 50, 75, 100].map(pct => (
          <circle key={pct} cx={center} cy={center} r={(pct / 100) * radius}
            fill="none" stroke="#e5e7eb" strokeWidth="0.5" />
        ))}
        {/* Axis lines */}
        {metrics.map((_, i) => {
          const endPoint = getPoint(i, 100);
          return (
            <line key={i} x1={center} y1={center} x2={endPoint.x} y2={endPoint.y}
              stroke="#e5e7eb" strokeWidth="0.5" />
          );
        })}
        {/* Median polygon */}
        <polygon points={p50Path} fill="#e5e7eb" fillOpacity="0.3"
          stroke="#9ca3af" strokeWidth="0.5" strokeDasharray="3,3" />
        {/* Org polygon */}
        <polygon points={orgPath} fill="#3b82f6" fillOpacity="0.2"
          stroke="#3b82f6" strokeWidth="1.5" />
        {/* Org data points */}
        {orgPoints.map((p, i) => (
          <circle key={i} cx={p.x} cy={p.y} r="3" fill="#3b82f6" />
        ))}
        {/* Labels */}
        {metrics.map((m, i) => {
          const labelPoint = getPoint(i, 120);
          return (
            <text key={i} x={labelPoint.x} y={labelPoint.y}
              textAnchor="middle" dominantBaseline="middle"
              className="text-[8px] fill-gray-600">
              {METRIC_LABELS[m.metric_name] || m.metric_name}
            </text>
          );
        })}
      </svg>
    </div>
  );
}

function BenchmarkTable({ metrics }: { metrics: PeerMetric[] }) {
  return (
    <div className="overflow-x-auto">
      <table className="min-w-full divide-y divide-gray-200">
        <thead className="bg-gray-50">
          <tr>
            <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Metric</th>
            <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">Your Value</th>
            <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">P25</th>
            <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">P50 (Median)</th>
            <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">P75</th>
            <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">P90</th>
            <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">Your Percentile</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-gray-200">
          {metrics.map(m => (
            <tr key={m.metric_name} className="hover:bg-gray-50">
              <td className="px-4 py-2 text-sm font-medium text-gray-900">
                {METRIC_LABELS[m.metric_name] || m.metric_name}
              </td>
              <td className="px-4 py-2 text-sm text-right font-semibold text-blue-600">
                {m.org_value.toFixed(1)}
              </td>
              <td className="px-4 py-2 text-sm text-right text-gray-600">{m.percentile_25.toFixed(1)}</td>
              <td className="px-4 py-2 text-sm text-right text-gray-600">{m.percentile_50.toFixed(1)}</td>
              <td className="px-4 py-2 text-sm text-right text-gray-600">{m.percentile_75.toFixed(1)}</td>
              <td className="px-4 py-2 text-sm text-right text-gray-600">{m.percentile_90.toFixed(1)}</td>
              <td className="px-4 py-2 text-sm text-right">
                <PercentileBar position={m.percentile_position} />
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function PercentileBar({ position }: { position: number }) {
  const color = position >= 75 ? 'bg-green-500' : position >= 50 ? 'bg-blue-500' : position >= 25 ? 'bg-amber-500' : 'bg-red-500';
  return (
    <div className="flex items-center gap-2 justify-end">
      <div className="w-20 h-2 rounded-full bg-gray-100 relative">
        <div
          className={`h-2 rounded-full ${color} transition-all duration-500`}
          style={{ width: `${Math.min(position, 100)}%` }}
        />
      </div>
      <span className="text-xs font-medium text-gray-700 w-8 text-right">P{position.toFixed(0)}</span>
    </div>
  );
}

function WidgetIcon({ type }: { type: string }) {
  const iconPaths: Record<string, string> = {
    line_chart: 'M3 17l6-6 4 4 8-8',
    bar_chart: 'M9 17V9m4 8V5m4 12V11M5 21h14',
    donut_chart: 'M12 2a10 10 0 100 20 10 10 0 000-20zm0 6a4 4 0 110 8 4 4 0 010-8z',
    kpi_card: 'M4 5a1 1 0 011-1h14a1 1 0 011 1v14a1 1 0 01-1 1H5a1 1 0 01-1-1V5z',
    heatmap: 'M4 4h4v4H4V4zm6 0h4v4h-4V4zm6 0h4v4h-4V4zM4 10h4v4H4v-4zm6 0h4v4h-4v-4z',
    radar: 'M12 2L2 12l10 10 10-10L12 2z',
    table: 'M3 10h18M3 14h18M3 6h18M3 18h18',
    gauge: 'M12 2a10 10 0 00-7.07 17.07',
    sparkline: 'M3 12h2l2-4 3 8 2-4 3 3 2-1h3',
    trend_arrow: 'M5 10l7-7m0 0l7 7m-7-7v18',
  };

  return (
    <svg className="h-8 w-8 mx-auto text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
      <path strokeLinecap="round" strokeLinejoin="round" d={iconPaths[type] || iconPaths.kpi_card} />
    </svg>
  );
}

// ============================================================
// HELPERS
// ============================================================

function getLatestTrendPerFramework(trends: ComplianceTrend[]): ComplianceTrend[] {
  const latest: Record<string, ComplianceTrend> = {};
  for (const t of trends) {
    if (!latest[t.framework_code] || t.measurement_date > latest[t.framework_code].measurement_date) {
      latest[t.framework_code] = t;
    }
  }
  return Object.values(latest);
}

function getScoreTextColor(score: number): string {
  if (score >= 80) return 'text-green-600';
  if (score >= 60) return 'text-amber-600';
  return 'text-red-600';
}

function getScoreBarColor(score: number): string {
  if (score >= 80) return 'bg-green-500';
  if (score >= 60) return 'bg-amber-400';
  return 'bg-red-500';
}
