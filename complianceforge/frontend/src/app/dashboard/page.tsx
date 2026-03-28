'use client';

import { useEffect, useState } from 'react';
import api from '@/lib/api';
import type { DashboardSummary } from '@/types';

export default function DashboardPage() {
  const [data, setData] = useState<DashboardSummary | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    api.getDashboard()
      .then((res) => setData(res.data))
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <div className="flex items-center justify-center h-64"><p className="text-gray-500">Loading dashboard...</p></div>;
  if (error) return <div className="rounded-lg bg-red-50 border border-red-200 p-4 text-red-700">{error}</div>;
  if (!data) return null;

  return (
    <div>
      {/* Header */}
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-gray-900">Executive Dashboard</h1>
        <p className="text-gray-500 mt-1">Compliance posture overview across all frameworks</p>
      </div>

      {/* Alert Banner */}
      {data.breaches_near_deadline > 0 && (
        <div className="mb-6 rounded-lg bg-red-50 border border-red-200 p-4 flex items-center gap-3">
          <span className="text-red-600 text-xl">⚠️</span>
          <div>
            <p className="font-semibold text-red-800">
              {data.breaches_near_deadline} data breach{data.breaches_near_deadline > 1 ? 'es' : ''} approaching 72-hour GDPR notification deadline
            </p>
            <p className="text-red-600 text-sm mt-0.5">Immediate action required per GDPR Article 33</p>
          </div>
          <a href="/incidents?filter=breaches_urgent" className="ml-auto rounded-lg bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-700">
            View Breaches
          </a>
        </div>
      )}

      {/* KPI Cards */}
      <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-6 mb-8">
        <KPICard
          label="Compliance Score"
          value={`${data.overall_compliance_score}%`}
          color={data.overall_compliance_score >= 80 ? 'green' : data.overall_compliance_score >= 60 ? 'amber' : 'red'}
        />
        <KPICard label="Open Risks" value={String(Object.values(data.risk_summary).reduce((a, b) => a + b, 0))} subtitle={`${data.risk_summary.critical} critical`} color="red" />
        <KPICard label="Open Incidents" value={String(data.open_incidents)} color={data.open_incidents > 5 ? 'red' : 'amber'} />
        <KPICard label="Open Findings" value={String(data.open_findings)} color={data.open_findings > 10 ? 'amber' : 'green'} />
        <KPICard label="Policies Due" value={String(data.policies_due_for_review)} color={data.policies_due_for_review > 3 ? 'amber' : 'green'} />
        <KPICard label="High-Risk Vendors" value={String(data.vendors_high_risk)} color={data.vendors_high_risk > 5 ? 'red' : 'amber'} />
      </div>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        {/* Framework Compliance Scores */}
        <div className="rounded-xl border border-gray-200 bg-white p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Framework Compliance</h2>
          <div className="space-y-4">
            {data.framework_scores.map((fw) => (
              <div key={fw.code}>
                <div className="flex items-center justify-between mb-1">
                  <span className="text-sm font-medium text-gray-700">{fw.name}</span>
                  <span className={`text-sm font-bold ${getScoreColor(fw.score)}`}>{fw.score.toFixed(1)}%</span>
                </div>
                <div className="h-2.5 w-full rounded-full bg-gray-100">
                  <div
                    className={`h-2.5 rounded-full transition-all duration-500 ${getBarColor(fw.score)}`}
                    style={{ width: `${Math.min(fw.score, 100)}%` }}
                  />
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Risk Distribution */}
        <div className="rounded-xl border border-gray-200 bg-white p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Risk Distribution</h2>
          <div className="space-y-3">
            {Object.entries(data.risk_summary).map(([level, count]) => (
              <div key={level} className="flex items-center gap-4">
                <span className={`inline-flex h-3 w-3 rounded-full ${getRiskDot(level)}`} />
                <span className="w-20 text-sm font-medium text-gray-700 capitalize">{level}</span>
                <div className="flex-1 h-6 rounded bg-gray-100 relative">
                  <div
                    className={`h-6 rounded ${getRiskBar(level)} transition-all duration-500 flex items-center px-2`}
                    style={{ width: `${Math.max((count / 48) * 100, 8)}%` }}
                  >
                    <span className="text-xs font-bold text-white">{count}</span>
                  </div>
                </div>
              </div>
            ))}
          </div>

          {/* Risk Heatmap Mini */}
          <div className="mt-6 pt-4 border-t border-gray-100">
            <h3 className="text-sm font-medium text-gray-600 mb-3">5×5 Risk Heatmap</h3>
            <div className="grid grid-cols-5 gap-1">
              {[5,4,3,2,1].map(impact => (
                [1,2,3,4,5].map(likelihood => {
                  const score = likelihood * impact;
                  return (
                    <div
                      key={`${likelihood}-${impact}`}
                      className={`h-8 rounded text-xs flex items-center justify-center font-medium text-white
                        ${score >= 20 ? 'bg-red-600' : score >= 12 ? 'bg-orange-500' : score >= 6 ? 'bg-yellow-500' : score >= 3 ? 'bg-green-400' : 'bg-green-300'}`}
                    >
                      {score}
                    </div>
                  );
                })
              ))}
            </div>
            <div className="flex justify-between mt-1">
              <span className="text-xs text-gray-400">Likelihood →</span>
              <span className="text-xs text-gray-400">↑ Impact</span>
            </div>
          </div>
        </div>
      </div>

      {/* Quick Actions */}
      <div className="mt-8 rounded-xl border border-gray-200 bg-white p-6">
        <h2 className="text-lg font-semibold text-gray-900 mb-4">Quick Actions</h2>
        <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
          <ActionButton href="/risks?action=new" label="Register Risk" icon="+" />
          <ActionButton href="/incidents?action=report" label="Report Incident" icon="!" />
          <ActionButton href="/policies?action=new" label="Draft Policy" icon="📝" />
          <ActionButton href="/audits?action=new" label="Plan Audit" icon="📋" />
        </div>
      </div>
    </div>
  );
}

// ── Helper Components ──────────────────────────────────
function KPICard({ label, value, subtitle, color }: { label: string; value: string; subtitle?: string; color: 'green' | 'amber' | 'red' }) {
  const borderColor = { green: 'border-green-200', amber: 'border-amber-200', red: 'border-red-200' }[color];
  const bgColor = { green: 'bg-green-50', amber: 'bg-amber-50', red: 'bg-red-50' }[color];
  const textColor = { green: 'text-green-700', amber: 'text-amber-700', red: 'text-red-700' }[color];

  return (
    <div className={`rounded-xl border ${borderColor} ${bgColor} p-4`}>
      <p className="text-xs font-medium text-gray-500 uppercase tracking-wide">{label}</p>
      <p className={`text-2xl font-bold ${textColor} mt-1`}>{value}</p>
      {subtitle && <p className="text-xs text-gray-500 mt-0.5">{subtitle}</p>}
    </div>
  );
}

function ActionButton({ href, label, icon }: { href: string; label: string; icon: string }) {
  return (
    <a href={href} className="flex items-center gap-2 rounded-lg border border-gray-200 bg-white p-3 text-sm font-medium text-gray-700 hover:bg-gray-50 hover:border-indigo-300 transition-colors">
      <span className="text-lg">{icon}</span>
      {label}
    </a>
  );
}

function getScoreColor(score: number) {
  if (score >= 80) return 'text-green-600';
  if (score >= 60) return 'text-amber-600';
  return 'text-red-600';
}

function getBarColor(score: number) {
  if (score >= 80) return 'bg-green-500';
  if (score >= 60) return 'bg-amber-500';
  return 'bg-red-500';
}

function getRiskDot(level: string) {
  return { critical: 'bg-red-600', high: 'bg-orange-500', medium: 'bg-yellow-500', low: 'bg-green-500' }[level] || 'bg-gray-400';
}

function getRiskBar(level: string) {
  return { critical: 'bg-red-600', high: 'bg-orange-500', medium: 'bg-yellow-500', low: 'bg-green-500' }[level] || 'bg-gray-400';
}
