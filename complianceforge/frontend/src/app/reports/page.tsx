'use client';

import { useEffect, useState } from 'react';
import api from '@/lib/api';
import type { ReportDefinition, ReportSchedule, ReportRun, ReportType, ReportFormat } from '@/types';

const REPORT_TYPE_LABELS: Record<string, string> = {
  compliance_status: 'Compliance Status',
  risk_register: 'Risk Register',
  risk_heatmap: 'Risk Heatmap',
  audit_summary: 'Audit Summary',
  audit_findings: 'Audit Findings',
  incident_summary: 'Incident Summary',
  breach_register: 'Breach Register',
  vendor_risk: 'Vendor Risk',
  policy_status: 'Policy Status',
  attestation_report: 'Attestation Report',
  gap_analysis: 'Gap Analysis',
  cross_framework_mapping: 'Cross-Framework Mapping',
  executive_summary: 'Executive Summary',
  kri_dashboard: 'KRI Dashboard',
  treatment_progress: 'Treatment Progress',
  custom: 'Custom Report',
};

const QUICK_REPORTS: { type: ReportType; label: string; description: string }[] = [
  { type: 'compliance_status', label: 'Compliance Status', description: 'Full compliance overview across all frameworks' },
  { type: 'risk_register', label: 'Risk Register', description: 'Complete risk register with heatmap' },
  { type: 'executive_summary', label: 'Executive Summary', description: 'Board-level 3-5 page summary' },
  { type: 'audit_findings', label: 'Audit Findings', description: 'Open findings by severity' },
  { type: 'incident_summary', label: 'Incident Summary', description: 'Incidents and breach register' },
  { type: 'gap_analysis', label: 'Gap Analysis', description: 'Compliance gaps with remediation roadmap' },
];

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

export default function ReportsPage() {
  const [tab, setTab] = useState<'generate' | 'definitions' | 'schedules' | 'history'>('generate');
  const [definitions, setDefinitions] = useState<ReportDefinition[]>([]);
  const [schedules, setSchedules] = useState<ReportSchedule[]>([]);
  const [history, setHistory] = useState<ReportRun[]>([]);
  const [loading, setLoading] = useState(false);
  const [generating, setGenerating] = useState<string | null>(null);

  useEffect(() => {
    loadData();
  }, [tab]);

  async function loadData() {
    setLoading(true);
    try {
      if (tab === 'definitions') {
        const res = await api.getReportDefinitions();
        setDefinitions(res.data?.data || res.data || []);
      } else if (tab === 'schedules') {
        const res = await api.getReportSchedules();
        setSchedules(res.data?.data || res.data || []);
      } else if (tab === 'history') {
        const res = await api.getReportHistory();
        setHistory(res.data?.data || res.data || []);
      }
    } catch { /* ignore */ }
    setLoading(false);
  }

  async function quickGenerate(type: ReportType, format: ReportFormat = 'pdf') {
    setGenerating(type);
    try {
      const res = await api.generateReport({ report_type: type, format });
      const runId = res.data?.id;
      if (runId) pollReportStatus(runId);
    } catch (e: any) {
      alert('Generation failed: ' + e.message);
      setGenerating(null);
    }
  }

  async function pollReportStatus(runId: string) {
    const check = async () => {
      try {
        const res = await api.getReportStatus(runId);
        const status = res.data?.status;
        if (status === 'completed') {
          setGenerating(null);
          window.open(`${API_BASE}/reports/download/${runId}`, '_blank');
        } else if (status === 'failed') {
          setGenerating(null);
          alert('Report generation failed: ' + (res.data?.error_message || 'Unknown error'));
        } else {
          setTimeout(check, 2000);
        }
      } catch {
        setGenerating(null);
      }
    };
    check();
  }

  const tabs = [
    { key: 'generate' as const, label: 'Quick Generate' },
    { key: 'definitions' as const, label: 'Saved Reports' },
    { key: 'schedules' as const, label: 'Schedules' },
    { key: 'history' as const, label: 'History' },
  ];

  return (
    <div>
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Reports</h1>
          <p className="text-gray-500 mt-1">Generate professional PDF and XLSX compliance reports</p>
        </div>
        <button className="btn-primary">Create Report Definition</button>
      </div>

      <div className="flex gap-1 border-b border-gray-200 mb-6">
        {tabs.map(t => (
          <button key={t.key} onClick={() => setTab(t.key)}
            className={`px-4 py-2.5 text-sm font-medium border-b-2 transition-colors ${tab === t.key ? 'border-indigo-600 text-indigo-600' : 'border-transparent text-gray-500 hover:text-gray-700'}`}
          >
            {t.label}
          </button>
        ))}
      </div>

      {tab === 'generate' ? (
        <div>
          <h2 className="font-semibold text-gray-900 mb-4">Quick Generate</h2>
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
            {QUICK_REPORTS.map(r => (
              <div key={r.type} className="card hover:border-indigo-300 hover:shadow-md transition-all">
                <h3 className="font-semibold text-gray-900">{r.label}</h3>
                <p className="text-sm text-gray-600 mt-1">{r.description}</p>
                <div className="flex gap-2 mt-4">
                  <button onClick={() => quickGenerate(r.type, 'pdf')} disabled={generating === r.type} className="btn-primary text-xs py-1.5">
                    {generating === r.type ? 'Generating...' : 'PDF'}
                  </button>
                  <button onClick={() => quickGenerate(r.type, 'xlsx')} disabled={generating === r.type} className="btn-secondary text-xs py-1.5">
                    XLSX
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>
      ) : loading ? <div className="card animate-pulse h-64" /> : tab === 'definitions' ? (
        <div className="card overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="table-header">
                <th className="px-4 py-3">Name</th>
                <th className="px-4 py-3">Type</th>
                <th className="px-4 py-3">Format</th>
                <th className="px-4 py-3">Classification</th>
                <th className="px-4 py-3">Actions</th>
              </tr>
            </thead>
            <tbody>
              {definitions.map(def => (
                <tr key={def.id} className="border-t border-gray-100">
                  <td className="px-4 py-3 font-medium text-gray-900">{def.name}</td>
                  <td className="px-4 py-3"><span className="badge badge-info">{REPORT_TYPE_LABELS[def.report_type] || def.report_type}</span></td>
                  <td className="px-4 py-3 uppercase text-gray-600">{def.format}</td>
                  <td className="px-4 py-3 text-gray-600">{def.classification}</td>
                  <td className="px-4 py-3 flex gap-2">
                    <button onClick={() => api.generateFromDefinition(def.id)} className="text-indigo-600 hover:text-indigo-700 text-xs font-medium">Generate</button>
                    <button className="text-gray-500 hover:text-gray-700 text-xs font-medium">Edit</button>
                  </td>
                </tr>
              ))}
              {definitions.length === 0 && (
                <tr><td colSpan={5} className="px-4 py-12 text-center text-gray-500">No saved report definitions</td></tr>
              )}
            </tbody>
          </table>
        </div>
      ) : tab === 'schedules' ? (
        <div>
          <div className="flex justify-end mb-4">
            <button className="btn-primary">Create Schedule</button>
          </div>
          <div className="card overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="table-header">
                  <th className="px-4 py-3">Name</th>
                  <th className="px-4 py-3">Frequency</th>
                  <th className="px-4 py-3">Delivery</th>
                  <th className="px-4 py-3">Last Run</th>
                  <th className="px-4 py-3">Next Run</th>
                  <th className="px-4 py-3">Active</th>
                  <th className="px-4 py-3">Actions</th>
                </tr>
              </thead>
              <tbody>
                {schedules.map(sch => (
                  <tr key={sch.id} className="border-t border-gray-100">
                    <td className="px-4 py-3 font-medium text-gray-900">{sch.name}</td>
                    <td className="px-4 py-3 capitalize text-gray-600">{sch.frequency}</td>
                    <td className="px-4 py-3 capitalize text-gray-600">{sch.delivery_channel}</td>
                    <td className="px-4 py-3 text-gray-600">{sch.last_run_at ? new Date(sch.last_run_at).toLocaleDateString('en-GB') : 'Never'}</td>
                    <td className="px-4 py-3 text-gray-600">{sch.next_run_at ? new Date(sch.next_run_at).toLocaleDateString('en-GB') : '—'}</td>
                    <td className="px-4 py-3"><span className={`badge ${sch.is_active ? 'badge-low' : 'badge-medium'}`}>{sch.is_active ? 'Active' : 'Paused'}</span></td>
                    <td className="px-4 py-3"><button className="text-indigo-600 hover:text-indigo-700 text-xs font-medium">Edit</button></td>
                  </tr>
                ))}
                {schedules.length === 0 && (
                  <tr><td colSpan={7} className="px-4 py-12 text-center text-gray-500">No report schedules</td></tr>
                )}
              </tbody>
            </table>
          </div>
        </div>
      ) : tab === 'history' ? (
        <div className="card overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="table-header">
                <th className="px-4 py-3">Report</th>
                <th className="px-4 py-3">Format</th>
                <th className="px-4 py-3">Status</th>
                <th className="px-4 py-3">Generated</th>
                <th className="px-4 py-3">Size</th>
                <th className="px-4 py-3">Time</th>
                <th className="px-4 py-3">Download</th>
              </tr>
            </thead>
            <tbody>
              {history.map(run => (
                <tr key={run.id} className="border-t border-gray-100">
                  <td className="px-4 py-3 font-medium text-gray-900">{run.report_definition_id.slice(0, 8)}...</td>
                  <td className="px-4 py-3 uppercase text-gray-600">{run.format}</td>
                  <td className="px-4 py-3">
                    <span className={`badge ${run.status === 'completed' ? 'badge-low' : run.status === 'failed' ? 'badge-critical' : run.status === 'generating' ? 'badge-info' : 'badge-medium'}`}>{run.status}</span>
                  </td>
                  <td className="px-4 py-3 text-gray-600">{new Date(run.created_at).toLocaleDateString('en-GB', { day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit' })}</td>
                  <td className="px-4 py-3 text-gray-600">{run.file_size_bytes ? `${(run.file_size_bytes / 1024).toFixed(0)} KB` : '—'}</td>
                  <td className="px-4 py-3 text-gray-600">{run.generation_time_ms ? `${(run.generation_time_ms / 1000).toFixed(1)}s` : '—'}</td>
                  <td className="px-4 py-3">
                    {run.status === 'completed' && (
                      <a href={`${API_BASE}/reports/download/${run.id}`} target="_blank" className="text-indigo-600 hover:text-indigo-700 text-xs font-medium">Download</a>
                    )}
                  </td>
                </tr>
              ))}
              {history.length === 0 && (
                <tr><td colSpan={7} className="px-4 py-12 text-center text-gray-500">No report history</td></tr>
              )}
            </tbody>
          </table>
        </div>
      ) : null}
    </div>
  );
}
