'use client';

import { useEffect, useState } from 'react';
import api from '@/lib/api';
import type { NIS2Dashboard, NIS2SecurityMeasure, NIS2IncidentReport, NIS2ManagementRecord, NIS2EntityAssessment } from '@/types';

const MEASURE_STATUS_COLORS: Record<string, string> = {
  not_started: 'badge-critical',
  in_progress: 'badge-medium',
  implemented: 'badge-info',
  verified: 'badge-low',
};

export default function NIS2Page() {
  const [tab, setTab] = useState<'dashboard' | 'measures' | 'incidents' | 'management' | 'assessment'>('dashboard');
  const [dashboard, setDashboard] = useState<NIS2Dashboard | null>(null);
  const [measures, setMeasures] = useState<NIS2SecurityMeasure[]>([]);
  const [incidents, setIncidents] = useState<NIS2IncidentReport[]>([]);
  const [management, setManagement] = useState<NIS2ManagementRecord[]>([]);
  const [assessment, setAssessment] = useState<NIS2EntityAssessment | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadData();
  }, [tab]);

  async function loadData() {
    setLoading(true);
    try {
      if (tab === 'dashboard') {
        const res = await api.getNIS2Dashboard();
        setDashboard(res.data);
      } else if (tab === 'measures') {
        const res = await api.getNIS2Measures();
        setMeasures(res.data?.data || res.data || []);
      } else if (tab === 'incidents') {
        const res = await api.getNIS2Incidents();
        setIncidents(res.data?.data || res.data || []);
      } else if (tab === 'management') {
        const res = await api.getNIS2Management();
        setManagement(res.data?.data || res.data || []);
      } else {
        const res = await api.getNIS2Assessment();
        setAssessment(res.data);
      }
    } catch { /* ignore */ }
    setLoading(false);
  }

  const tabs = [
    { key: 'dashboard' as const, label: 'Dashboard' },
    { key: 'measures' as const, label: 'Security Measures' },
    { key: 'incidents' as const, label: 'Incident Reports' },
    { key: 'management' as const, label: 'Management' },
    { key: 'assessment' as const, label: 'Entity Assessment' },
  ];

  return (
    <div>
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">NIS2 Compliance</h1>
          <p className="text-gray-500 mt-1">EU Network and Information Security Directive compliance management</p>
        </div>
      </div>

      <div className="flex gap-1 border-b border-gray-200 mb-6 overflow-x-auto">
        {tabs.map(t => (
          <button key={t.key} onClick={() => setTab(t.key)}
            className={`px-4 py-2.5 text-sm font-medium border-b-2 whitespace-nowrap transition-colors ${tab === t.key ? 'border-indigo-600 text-indigo-600' : 'border-transparent text-gray-500 hover:text-gray-700'}`}
          >
            {t.label}
          </button>
        ))}
      </div>

      {loading ? <div className="card animate-pulse h-96" /> : tab === 'dashboard' && dashboard ? (
        <div>
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-4 mb-6">
            <div className="card">
              <p className="text-sm text-gray-500">Entity Type</p>
              <p className="text-2xl font-bold text-gray-900 mt-1 capitalize">{dashboard.entity_type}</p>
              <p className="text-xs text-gray-500 mt-1">{dashboard.is_in_scope ? 'In Scope' : 'Not in Scope'}</p>
            </div>
            <div className="card">
              <p className="text-sm text-gray-500">Measures Implemented</p>
              <p className="text-2xl font-bold text-indigo-600 mt-1">{dashboard.measures_implemented}/{dashboard.measures_total}</p>
              <div className="mt-2 h-2 bg-gray-100 rounded-full overflow-hidden">
                <div className="h-full bg-indigo-600 rounded-full" style={{ width: `${(dashboard.measures_implemented / dashboard.measures_total) * 100}%` }} />
              </div>
            </div>
            <div className="card">
              <p className="text-sm text-gray-500">Open Incidents</p>
              <p className="text-2xl font-bold text-gray-900 mt-1">{dashboard.open_incidents}</p>
              {dashboard.overdue_reports > 0 && <p className="text-xs text-red-600 mt-1">{dashboard.overdue_reports} overdue reports</p>}
            </div>
            <div className="card">
              <p className="text-sm text-gray-500">Management Training</p>
              <p className="text-2xl font-bold text-gray-900 mt-1">{dashboard.management_trained}/{dashboard.management_total}</p>
              <p className="text-xs text-gray-500 mt-1">Board members trained</p>
            </div>
          </div>
        </div>
      ) : tab === 'measures' ? (
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
          {measures.map(m => (
            <div key={m.id} className="card">
              <div className="flex items-start justify-between">
                <div className="flex-1">
                  <div className="flex items-center gap-2">
                    <span className="text-xs font-mono text-indigo-600 bg-indigo-50 px-2 py-0.5 rounded">{m.measure_code}</span>
                    <span className={`badge ${MEASURE_STATUS_COLORS[m.implementation_status] || 'badge-info'}`}>
                      {m.implementation_status.replace(/_/g, ' ')}
                    </span>
                  </div>
                  <h3 className="font-semibold text-gray-900 mt-2">{m.measure_title}</h3>
                  <p className="text-sm text-gray-600 mt-1">{m.measure_description}</p>
                  <p className="text-xs text-gray-400 mt-2">{m.article_reference}</p>
                </div>
              </div>
              {m.linked_control_ids?.length > 0 && (
                <p className="text-xs text-gray-500 mt-3 pt-3 border-t border-gray-100">
                  {m.linked_control_ids.length} linked controls
                </p>
              )}
            </div>
          ))}
          {measures.length === 0 && (
            <div className="col-span-full text-center py-12 text-gray-500">
              No NIS2 measures found. Run the seed data to populate Article 21 measures.
            </div>
          )}
        </div>
      ) : tab === 'incidents' ? (
        <div className="card overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="table-header">
                <th className="px-4 py-3">Reference</th>
                <th className="px-4 py-3">Phase 1: Early Warning (24h)</th>
                <th className="px-4 py-3">Phase 2: Notification (72h)</th>
                <th className="px-4 py-3">Phase 3: Final Report (1mo)</th>
                <th className="px-4 py-3">Actions</th>
              </tr>
            </thead>
            <tbody>
              {incidents.map(inc => (
                <tr key={inc.id} className="border-t border-gray-100">
                  <td className="px-4 py-3 font-medium text-gray-900">{inc.report_ref}</td>
                  <td className="px-4 py-3">
                    <PhaseCell status={inc.early_warning_status} deadline={inc.early_warning_deadline} submittedAt={inc.early_warning_submitted_at} />
                  </td>
                  <td className="px-4 py-3">
                    <PhaseCell status={inc.notification_status} deadline={inc.notification_deadline} submittedAt={inc.notification_submitted_at} />
                  </td>
                  <td className="px-4 py-3">
                    <PhaseCell status={inc.final_report_status} deadline={inc.final_report_deadline} submittedAt={inc.final_report_submitted_at} />
                  </td>
                  <td className="px-4 py-3">
                    <a href={`/nis2/incidents/${inc.id}`} className="text-indigo-600 hover:text-indigo-700 text-xs font-medium">View</a>
                  </td>
                </tr>
              ))}
              {incidents.length === 0 && (
                <tr><td colSpan={5} className="px-4 py-12 text-center text-gray-500">No NIS2 incident reports</td></tr>
              )}
            </tbody>
          </table>
        </div>
      ) : tab === 'management' ? (
        <div>
          <div className="flex justify-end mb-4">
            <button className="btn-primary">Add Board Member</button>
          </div>
          <div className="card overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="table-header">
                  <th className="px-4 py-3">Name</th>
                  <th className="px-4 py-3">Role</th>
                  <th className="px-4 py-3">Training</th>
                  <th className="px-4 py-3">Training Date</th>
                  <th className="px-4 py-3">Risk Measures Approved</th>
                  <th className="px-4 py-3">Next Training Due</th>
                </tr>
              </thead>
              <tbody>
                {management.map(m => (
                  <tr key={m.id} className="border-t border-gray-100">
                    <td className="px-4 py-3 font-medium text-gray-900">{m.board_member_name}</td>
                    <td className="px-4 py-3 text-gray-600">{m.board_member_role}</td>
                    <td className="px-4 py-3">
                      <span className={`badge ${m.training_completed ? 'badge-low' : 'badge-critical'}`}>
                        {m.training_completed ? 'Completed' : 'Required'}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-gray-600">{m.training_date ? new Date(m.training_date).toLocaleDateString('en-GB') : '—'}</td>
                    <td className="px-4 py-3">
                      <span className={`badge ${m.risk_measures_approved ? 'badge-low' : 'badge-medium'}`}>
                        {m.risk_measures_approved ? 'Approved' : 'Pending'}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-gray-600">{m.next_training_due ? new Date(m.next_training_due).toLocaleDateString('en-GB') : '—'}</td>
                  </tr>
                ))}
                {management.length === 0 && (
                  <tr><td colSpan={6} className="px-4 py-12 text-center text-gray-500">No management records</td></tr>
                )}
              </tbody>
            </table>
          </div>
        </div>
      ) : tab === 'assessment' ? (
        <div className="card max-w-2xl">
          {assessment ? (
            <div>
              <h2 className="font-semibold text-gray-900 mb-4">Entity Assessment</h2>
              <dl className="space-y-3 text-sm">
                <div className="flex justify-between"><dt className="text-gray-500">Entity Type</dt><dd className="font-bold capitalize">{assessment.entity_type}</dd></div>
                <div className="flex justify-between"><dt className="text-gray-500">Sector</dt><dd>{assessment.sector}</dd></div>
                <div className="flex justify-between"><dt className="text-gray-500">Sub-Sector</dt><dd>{assessment.sub_sector || '—'}</dd></div>
                <div className="flex justify-between"><dt className="text-gray-500">Employee Count</dt><dd>{assessment.employee_count?.toLocaleString()}</dd></div>
                <div className="flex justify-between"><dt className="text-gray-500">Annual Turnover</dt><dd>€{assessment.annual_turnover_eur?.toLocaleString()}</dd></div>
                <div className="flex justify-between"><dt className="text-gray-500">In Scope</dt><dd>{assessment.is_in_scope ? 'Yes' : 'No'}</dd></div>
                <div className="flex justify-between"><dt className="text-gray-500">Member State</dt><dd>{assessment.member_state}</dd></div>
                <div className="flex justify-between"><dt className="text-gray-500">Competent Authority</dt><dd>{assessment.competent_authority}</dd></div>
                <div className="flex justify-between"><dt className="text-gray-500">CSIRT</dt><dd>{assessment.csirt_name}</dd></div>
                <div className="flex justify-between"><dt className="text-gray-500">Assessment Date</dt><dd>{new Date(assessment.assessment_date).toLocaleDateString('en-GB')}</dd></div>
              </dl>
              <button className="btn-secondary mt-6">Reassess</button>
            </div>
          ) : (
            <div className="text-center py-8">
              <p className="text-gray-500 mb-4">No entity assessment completed yet.</p>
              <p className="text-sm text-gray-400 mb-6">Complete the NIS2 entity categorisation to determine if your organisation is classified as essential or important.</p>
              <button className="btn-primary">Start Assessment</button>
            </div>
          )}
        </div>
      ) : null}
    </div>
  );
}

function PhaseCell({ status, deadline, submittedAt }: { status: string; deadline: string; submittedAt: string | null }) {
  const colors: Record<string, string> = {
    not_required: 'text-gray-400',
    pending: 'text-amber-600',
    submitted: 'text-green-600',
    overdue: 'text-red-600',
  };
  return (
    <div>
      <span className={`text-sm font-medium ${colors[status] || ''}`}>{status.replace(/_/g, ' ')}</span>
      {deadline && status !== 'not_required' && (
        <p className="text-xs text-gray-400 mt-0.5">
          {submittedAt ? `Submitted ${new Date(submittedAt).toLocaleDateString('en-GB')}` : `Due ${new Date(deadline).toLocaleDateString('en-GB')}`}
        </p>
      )}
    </div>
  );
}
