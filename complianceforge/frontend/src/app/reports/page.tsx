'use client';

import { useState } from 'react';
import api from '@/lib/api';

interface ComplianceReportData {
  generated_at: string;
  overall_score: number;
  frameworks: FrameworkReportEntry[];
  total_controls: number;
  implemented_controls: number;
  gaps_count: number;
  recommendations: string[];
}

interface FrameworkReportEntry {
  code: string;
  name: string;
  score: number;
  total_controls: number;
  implemented: number;
  gaps: number;
}

interface RiskReportData {
  generated_at: string;
  total_risks: number;
  risk_by_level: Record<string, number>;
  top_risks: RiskReportEntry[];
  average_risk_score: number;
  total_financial_exposure: number;
  mitigation_rate: number;
}

interface RiskReportEntry {
  risk_ref: string;
  title: string;
  risk_level: string;
  risk_score: number;
  financial_impact: number;
  status: string;
}

export default function ReportsPage() {
  const [complianceReport, setComplianceReport] = useState<ComplianceReportData | null>(null);
  const [riskReport, setRiskReport] = useState<RiskReportData | null>(null);
  const [complianceLoading, setComplianceLoading] = useState(false);
  const [riskLoading, setRiskLoading] = useState(false);
  const [complianceError, setComplianceError] = useState<string | null>(null);
  const [riskError, setRiskError] = useState<string | null>(null);

  async function generateComplianceReport() {
    setComplianceLoading(true);
    setComplianceError(null);
    try {
      const res = await api.getComplianceReport();
      setComplianceReport(res.data);
    } catch (err: unknown) {
      setComplianceError(err instanceof Error ? err.message : 'Failed to generate report');
    } finally {
      setComplianceLoading(false);
    }
  }

  async function generateRiskReport() {
    setRiskLoading(true);
    setRiskError(null);
    try {
      const res = await api.getRiskReport();
      setRiskReport(res.data);
    } catch (err: unknown) {
      setRiskError(err instanceof Error ? err.message : 'Failed to generate report');
    } finally {
      setRiskLoading(false);
    }
  }

  return (
    <div>
      {/* Header */}
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-gray-900">Reports</h1>
        <p className="text-gray-500 mt-1">Generate compliance and risk reports for stakeholder review</p>
      </div>

      {/* Report Cards */}
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2 mb-8">
        {/* Compliance Report Card */}
        <div className="card">
          <div className="flex items-center gap-3 mb-3">
            <div className="h-10 w-10 rounded-lg bg-indigo-50 flex items-center justify-center">
              <span className="text-indigo-600 font-bold text-sm">CR</span>
            </div>
            <div>
              <h2 className="text-lg font-semibold text-gray-900">Compliance Report</h2>
              <p className="text-xs text-gray-500">Framework compliance posture across all standards</p>
            </div>
          </div>
          <p className="text-sm text-gray-600 mb-4">
            Generates a comprehensive overview of compliance scores, implementation status,
            gaps, and recommendations across all adopted frameworks.
          </p>
          <button
            onClick={generateComplianceReport}
            disabled={complianceLoading}
            className="btn-primary"
          >
            {complianceLoading ? 'Generating...' : 'Generate Compliance Report'}
          </button>
        </div>

        {/* Risk Report Card */}
        <div className="card">
          <div className="flex items-center gap-3 mb-3">
            <div className="h-10 w-10 rounded-lg bg-red-50 flex items-center justify-center">
              <span className="text-red-600 font-bold text-sm">RR</span>
            </div>
            <div>
              <h2 className="text-lg font-semibold text-gray-900">Risk Report</h2>
              <p className="text-xs text-gray-500">Enterprise risk posture and financial exposure</p>
            </div>
          </div>
          <p className="text-sm text-gray-600 mb-4">
            Generates a risk landscape summary including risk distribution, top risks,
            financial exposure, and mitigation effectiveness.
          </p>
          <button
            onClick={generateRiskReport}
            disabled={riskLoading}
            className="btn-primary"
          >
            {riskLoading ? 'Generating...' : 'Generate Risk Report'}
          </button>
        </div>
      </div>

      {/* Compliance Report Results */}
      {complianceError && (
        <div className="mb-6 rounded-lg bg-red-50 border border-red-200 p-4 text-red-700">{complianceError}</div>
      )}

      {complianceReport && (
        <div className="mb-8">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-bold text-gray-900">Compliance Report</h2>
            <span className="text-xs text-gray-400">
              Generated {complianceReport.generated_at ? new Date(complianceReport.generated_at).toLocaleString('en-GB') : 'just now'}
            </span>
          </div>

          {/* KPIs */}
          <div className="grid grid-cols-4 gap-4 mb-6">
            <div className="card p-4 text-center">
              <p className="text-3xl font-bold text-indigo-600">{complianceReport.overall_score?.toFixed(1) ?? '—'}%</p>
              <p className="text-xs text-gray-500 mt-1">Overall Score</p>
            </div>
            <div className="card p-4 text-center">
              <p className="text-3xl font-bold text-green-600">{complianceReport.implemented_controls ?? '—'}</p>
              <p className="text-xs text-gray-500 mt-1">Implemented Controls</p>
            </div>
            <div className="card p-4 text-center">
              <p className="text-3xl font-bold text-gray-600">{complianceReport.total_controls ?? '—'}</p>
              <p className="text-xs text-gray-500 mt-1">Total Controls</p>
            </div>
            <div className="card p-4 text-center">
              <p className="text-3xl font-bold text-red-600">{complianceReport.gaps_count ?? '—'}</p>
              <p className="text-xs text-gray-500 mt-1">Open Gaps</p>
            </div>
          </div>

          {/* Framework Breakdown */}
          {complianceReport.frameworks && complianceReport.frameworks.length > 0 && (
            <div className="card mb-6">
              <h3 className="text-sm font-semibold text-gray-700 mb-4">Framework Breakdown</h3>
              <div className="space-y-3">
                {complianceReport.frameworks.map((fw) => (
                  <div key={fw.code}>
                    <div className="flex items-center justify-between mb-1">
                      <span className="text-sm font-medium text-gray-700">{fw.name}</span>
                      <div className="flex items-center gap-3">
                        <span className="text-xs text-gray-400">{fw.implemented}/{fw.total_controls} controls</span>
                        <span className={`text-sm font-bold ${fw.score >= 80 ? 'text-green-600' : fw.score >= 60 ? 'text-amber-600' : 'text-red-600'}`}>
                          {fw.score.toFixed(1)}%
                        </span>
                      </div>
                    </div>
                    <div className="h-2 w-full rounded-full bg-gray-100">
                      <div
                        className={`h-2 rounded-full transition-all duration-500 ${fw.score >= 80 ? 'bg-green-500' : fw.score >= 60 ? 'bg-amber-500' : 'bg-red-500'}`}
                        style={{ width: `${Math.min(fw.score, 100)}%` }}
                      />
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Recommendations */}
          {complianceReport.recommendations && complianceReport.recommendations.length > 0 && (
            <div className="card">
              <h3 className="text-sm font-semibold text-gray-700 mb-3">Recommendations</h3>
              <ul className="space-y-2">
                {complianceReport.recommendations.map((rec, idx) => (
                  <li key={idx} className="flex items-start gap-2">
                    <span className="inline-flex h-5 w-5 items-center justify-center rounded-full bg-indigo-100 text-xs font-medium text-indigo-600 flex-shrink-0 mt-0.5">
                      {idx + 1}
                    </span>
                    <span className="text-sm text-gray-700">{rec}</span>
                  </li>
                ))}
              </ul>
            </div>
          )}
        </div>
      )}

      {/* Risk Report Results */}
      {riskError && (
        <div className="mb-6 rounded-lg bg-red-50 border border-red-200 p-4 text-red-700">{riskError}</div>
      )}

      {riskReport && (
        <div className="mb-8">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-bold text-gray-900">Risk Report</h2>
            <span className="text-xs text-gray-400">
              Generated {riskReport.generated_at ? new Date(riskReport.generated_at).toLocaleString('en-GB') : 'just now'}
            </span>
          </div>

          {/* KPIs */}
          <div className="grid grid-cols-4 gap-4 mb-6">
            <div className="card p-4 text-center">
              <p className="text-3xl font-bold text-indigo-600">{riskReport.total_risks ?? '—'}</p>
              <p className="text-xs text-gray-500 mt-1">Total Risks</p>
            </div>
            <div className="card p-4 text-center">
              <p className="text-3xl font-bold text-amber-600">{riskReport.average_risk_score?.toFixed(1) ?? '—'}</p>
              <p className="text-xs text-gray-500 mt-1">Avg Risk Score</p>
            </div>
            <div className="card p-4 text-center">
              <p className="text-3xl font-bold text-red-600">
                {riskReport.total_financial_exposure != null ? `€${(riskReport.total_financial_exposure / 1000).toFixed(0)}K` : '—'}
              </p>
              <p className="text-xs text-gray-500 mt-1">Financial Exposure</p>
            </div>
            <div className="card p-4 text-center">
              <p className="text-3xl font-bold text-green-600">{riskReport.mitigation_rate?.toFixed(0) ?? '—'}%</p>
              <p className="text-xs text-gray-500 mt-1">Mitigation Rate</p>
            </div>
          </div>

          {/* Risk Distribution */}
          {riskReport.risk_by_level && (
            <div className="card mb-6">
              <h3 className="text-sm font-semibold text-gray-700 mb-3">Risk Distribution</h3>
              <div className="space-y-2">
                {Object.entries(riskReport.risk_by_level).map(([level, count]) => {
                  const colors: Record<string, string> = {
                    critical: 'bg-red-600',
                    high: 'bg-orange-500',
                    medium: 'bg-yellow-500',
                    low: 'bg-green-500',
                    very_low: 'bg-green-300',
                  };
                  const maxCount = Math.max(...Object.values(riskReport.risk_by_level), 1);
                  return (
                    <div key={level} className="flex items-center gap-3">
                      <span className="w-16 text-sm font-medium text-gray-600 capitalize">{level.replace(/_/g, ' ')}</span>
                      <div className="flex-1 h-6 rounded bg-gray-100">
                        <div
                          className={`h-6 rounded ${colors[level] || 'bg-gray-400'} flex items-center px-2 transition-all duration-500`}
                          style={{ width: `${Math.max((count / maxCount) * 100, 8)}%` }}
                        >
                          <span className="text-xs font-bold text-white">{count}</span>
                        </div>
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>
          )}

          {/* Top Risks */}
          {riskReport.top_risks && riskReport.top_risks.length > 0 && (
            <div className="card overflow-hidden p-0">
              <div className="px-4 py-3 border-b border-gray-100">
                <h3 className="text-sm font-semibold text-gray-700">Top Risks</h3>
              </div>
              <table className="w-full">
                <thead>
                  <tr className="border-b border-gray-100">
                    <th className="table-header px-4 py-3">Ref</th>
                    <th className="table-header px-4 py-3">Risk</th>
                    <th className="table-header px-4 py-3">Level</th>
                    <th className="table-header px-4 py-3">Score</th>
                    <th className="table-header px-4 py-3">Financial Impact</th>
                    <th className="table-header px-4 py-3">Status</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-50">
                  {riskReport.top_risks.map((risk) => (
                    <tr key={risk.risk_ref} className="hover:bg-gray-50">
                      <td className="px-4 py-3 text-sm font-mono text-gray-500">{risk.risk_ref}</td>
                      <td className="px-4 py-3 text-sm font-medium text-gray-900">{risk.title}</td>
                      <td className="px-4 py-3">
                        <span className={`badge badge-${risk.risk_level}`}>{risk.risk_level}</span>
                      </td>
                      <td className="px-4 py-3">
                        <span className={`text-sm font-bold ${risk.risk_score >= 20 ? 'text-red-600' : risk.risk_score >= 12 ? 'text-orange-600' : 'text-yellow-600'}`}>
                          {risk.risk_score}
                        </span>
                      </td>
                      <td className="px-4 py-3 text-sm text-gray-600">
                        {risk.financial_impact > 0 ? `€${risk.financial_impact.toLocaleString()}` : '—'}
                      </td>
                      <td className="px-4 py-3">
                        <span className="badge badge-info">{risk.status}</span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
