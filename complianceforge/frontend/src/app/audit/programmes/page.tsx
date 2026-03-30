'use client';

import { useEffect, useState, useCallback } from 'react';
import api from '@/lib/api';

// ============================================================
// TYPES
// ============================================================

interface AuditProgramme {
  id: string;
  programme_ref: string;
  name: string;
  description: string;
  status: string;
  programme_type: string;
  period_start: string;
  period_end: string;
  total_budget_days: number;
  used_budget_days: number;
  objectives: string;
  risk_appetite: string;
  methodology: string;
  owner_user_id: string | null;
  created_at: string;
}

interface AuditEngagement {
  id: string;
  programme_id: string;
  engagement_ref: string;
  name: string;
  engagement_type: string;
  status: string;
  priority: string;
  risk_rating: string;
  scope: string;
  objectives: string;
  lead_auditor_id: string | null;
  planned_start_date: string | null;
  planned_end_date: string | null;
  actual_start_date: string | null;
  actual_end_date: string | null;
  budget_days: number;
  actual_days: number;
  fieldwork_complete: boolean;
  report_issued: boolean;
  overall_opinion: string;
  created_at: string;
}

interface AuditWorkpaper {
  id: string;
  engagement_id: string;
  workpaper_ref: string;
  title: string;
  description: string;
  workpaper_type: string;
  status: string;
  content: string;
  prepared_by: string;
  prepared_date: string;
  reviewed_by: string | null;
  reviewed_date: string | null;
  review_comments: string;
}

interface AuditSample {
  id: string;
  engagement_id: string;
  sample_ref: string;
  name: string;
  sampling_method: string;
  population_size: number;
  sample_size: number;
  confidence_level: number;
  tolerable_error_rate: number;
  items_tested: number;
  items_passed: number;
  items_failed: number;
  status: string;
  actual_error_rate: number | null;
}

interface AuditTestProcedure {
  id: string;
  engagement_id: string;
  procedure_ref: string;
  title: string;
  test_type: string;
  control_ref: string;
  expected_result: string;
  actual_result: string;
  result: string | null;
  tested_by: string | null;
  tested_date: string | null;
}

interface AuditCorrectiveAction {
  id: string;
  action_ref: string;
  title: string;
  action_type: string;
  priority: string;
  status: string;
  due_date: string | null;
  completed_date: string | null;
  verified_by: string | null;
}

// ============================================================
// COLOUR HELPERS
// ============================================================

const STATUS_COLORS: Record<string, string> = {
  draft: 'bg-gray-100 text-gray-700',
  active: 'bg-green-100 text-green-800',
  approved: 'bg-blue-100 text-blue-800',
  completed: 'bg-emerald-100 text-emerald-800',
  cancelled: 'bg-red-100 text-red-700',
  planning: 'bg-gray-100 text-gray-700',
  fieldwork: 'bg-yellow-100 text-yellow-800',
  review: 'bg-orange-100 text-orange-800',
  reporting: 'bg-purple-100 text-purple-800',
};

const PRIORITY_COLORS: Record<string, string> = {
  critical: 'bg-red-100 text-red-800',
  high: 'bg-orange-100 text-orange-800',
  medium: 'bg-yellow-100 text-yellow-800',
  low: 'bg-blue-100 text-blue-800',
};

const RISK_COLORS: Record<string, string> = {
  critical: 'text-red-700 font-semibold',
  high: 'text-orange-600 font-semibold',
  medium: 'text-yellow-600',
  low: 'text-blue-600',
  very_low: 'text-gray-500',
};

type TabName = 'overview' | 'workpapers' | 'tests' | 'samples' | 'findings' | 'corrective-actions';

// ============================================================
// MAIN PAGE COMPONENT
// ============================================================

export default function AuditProgrammesPage() {
  const [programmes, setProgrammes] = useState<AuditProgramme[]>([]);
  const [engagements, setEngagements] = useState<AuditEngagement[]>([]);
  const [loading, setLoading] = useState(true);
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [selectedProgramme, setSelectedProgramme] = useState<AuditProgramme | null>(null);
  const [selectedEngagement, setSelectedEngagement] = useState<AuditEngagement | null>(null);
  const [activeTab, setActiveTab] = useState<TabName>('overview');

  // Engagement workspace data
  const [workpapers, setWorkpapers] = useState<AuditWorkpaper[]>([]);
  const [samples, setSamples] = useState<AuditSample[]>([]);
  const [testProcedures, setTestProcedures] = useState<AuditTestProcedure[]>([]);
  const [correctiveActions, setCorrectiveActions] = useState<AuditCorrectiveAction[]>([]);

  // Create form state
  const [newProgramme, setNewProgramme] = useState({
    name: '',
    description: '',
    programme_type: 'annual',
    period_start: '',
    period_end: '',
    total_budget_days: 200,
    objectives: '',
    risk_appetite: 'medium',
    methodology: '',
  });

  // --------------------------------------------------------
  // DATA FETCHING
  // --------------------------------------------------------

  const fetchProgrammes = useCallback(async () => {
    try {
      setLoading(true);
      const res = await api.get<{ data: AuditProgramme[]; pagination: unknown }>(
        '/audit/programmes'
      );
      setProgrammes(res.data || []);
    } catch {
      console.error('Failed to fetch audit programmes');
    } finally {
      setLoading(false);
    }
  }, []);

  const fetchEngagements = useCallback(async (programmeId: string) => {
    try {
      const res = await api.get<{ data: AuditEngagement[]; pagination: unknown }>(
        `/audit/engagements?programme_id=${programmeId}`
      );
      setEngagements(res.data || []);
    } catch {
      console.error('Failed to fetch engagements');
    }
  }, []);

  const fetchEngagementWorkspace = useCallback(async (engagementId: string) => {
    try {
      const [wpRes] = await Promise.all([
        api.get<{ success: boolean; data: AuditWorkpaper[] }>(
          `/audit/engagements/${engagementId}/workpapers`
        ),
      ]);
      setWorkpapers(wpRes.data || []);
    } catch {
      console.error('Failed to fetch engagement workspace data');
    }
  }, []);

  useEffect(() => {
    fetchProgrammes();
  }, [fetchProgrammes]);

  useEffect(() => {
    if (selectedProgramme) {
      fetchEngagements(selectedProgramme.id);
    }
  }, [selectedProgramme, fetchEngagements]);

  useEffect(() => {
    if (selectedEngagement) {
      fetchEngagementWorkspace(selectedEngagement.id);
    }
  }, [selectedEngagement, fetchEngagementWorkspace]);

  // --------------------------------------------------------
  // CREATE PROGRAMME
  // --------------------------------------------------------

  const handleCreateProgramme = async () => {
    try {
      const res = await api.post<{ success: boolean; data: AuditProgramme }>(
        '/audit/programmes',
        newProgramme
      );
      if (res.success) {
        setShowCreateForm(false);
        setNewProgramme({
          name: '',
          description: '',
          programme_type: 'annual',
          period_start: '',
          period_end: '',
          total_budget_days: 200,
          objectives: '',
          risk_appetite: 'medium',
          methodology: '',
        });
        fetchProgrammes();
      }
    } catch {
      console.error('Failed to create programme');
    }
  };

  // --------------------------------------------------------
  // HELPERS
  // --------------------------------------------------------

  const formatDate = (d: string | null) => {
    if (!d) return '-';
    return new Date(d).toLocaleDateString('en-GB', {
      day: '2-digit',
      month: 'short',
      year: 'numeric',
    });
  };

  const utilisationPct = (p: AuditProgramme) => {
    if (p.total_budget_days === 0) return 0;
    return Math.round((p.used_budget_days / p.total_budget_days) * 100);
  };

  // --------------------------------------------------------
  // ENGAGEMENT WORKSPACE
  // --------------------------------------------------------

  if (selectedEngagement) {
    const tabs: { key: TabName; label: string }[] = [
      { key: 'overview', label: 'Overview' },
      { key: 'workpapers', label: 'Workpapers' },
      { key: 'tests', label: 'Tests' },
      { key: 'samples', label: 'Samples' },
      { key: 'findings', label: 'Findings' },
      { key: 'corrective-actions', label: 'Corrective Actions' },
    ];

    return (
      <div className="p-6 max-w-7xl mx-auto">
        {/* Back navigation */}
        <button
          onClick={() => {
            setSelectedEngagement(null);
            setActiveTab('overview');
          }}
          className="mb-4 text-sm text-blue-600 hover:text-blue-800 flex items-center gap-1"
        >
          &larr; Back to Programme
        </button>

        {/* Engagement header */}
        <div className="mb-6">
          <div className="flex items-center gap-3 mb-2">
            <span className="text-sm font-mono text-gray-500">
              {selectedEngagement.engagement_ref}
            </span>
            <span
              className={`px-2 py-0.5 rounded text-xs font-medium ${STATUS_COLORS[selectedEngagement.status] || 'bg-gray-100 text-gray-700'}`}
            >
              {selectedEngagement.status}
            </span>
            <span
              className={`px-2 py-0.5 rounded text-xs font-medium ${RISK_COLORS[selectedEngagement.risk_rating] || ''}`}
            >
              Risk: {selectedEngagement.risk_rating}
            </span>
          </div>
          <h1 className="text-2xl font-bold text-gray-900">{selectedEngagement.name}</h1>
          <p className="text-gray-600 mt-1">{selectedEngagement.scope}</p>
          <div className="mt-3 flex gap-6 text-sm text-gray-500">
            <span>Start: {formatDate(selectedEngagement.planned_start_date)}</span>
            <span>End: {formatDate(selectedEngagement.planned_end_date)}</span>
            <span>Budget: {selectedEngagement.budget_days} days</span>
            <span>Actual: {selectedEngagement.actual_days} days</span>
          </div>
        </div>

        {/* Tab navigation */}
        <div className="border-b border-gray-200 mb-6">
          <nav className="flex gap-6 -mb-px">
            {tabs.map((tab) => (
              <button
                key={tab.key}
                onClick={() => setActiveTab(tab.key)}
                className={`pb-3 text-sm font-medium border-b-2 transition-colors ${
                  activeTab === tab.key
                    ? 'border-blue-600 text-blue-600'
                    : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                }`}
              >
                {tab.label}
              </button>
            ))}
          </nav>
        </div>

        {/* Tab content */}
        {activeTab === 'overview' && (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="bg-white border rounded-lg p-5">
              <h3 className="font-semibold text-gray-900 mb-3">Details</h3>
              <dl className="space-y-2 text-sm">
                <div className="flex justify-between">
                  <dt className="text-gray-500">Type</dt>
                  <dd className="text-gray-900">{selectedEngagement.engagement_type}</dd>
                </div>
                <div className="flex justify-between">
                  <dt className="text-gray-500">Priority</dt>
                  <dd>
                    <span className={`px-2 py-0.5 rounded text-xs ${PRIORITY_COLORS[selectedEngagement.priority] || 'bg-gray-100 text-gray-700'}`}>
                      {selectedEngagement.priority}
                    </span>
                  </dd>
                </div>
                <div className="flex justify-between">
                  <dt className="text-gray-500">Fieldwork Complete</dt>
                  <dd className="text-gray-900">{selectedEngagement.fieldwork_complete ? 'Yes' : 'No'}</dd>
                </div>
                <div className="flex justify-between">
                  <dt className="text-gray-500">Report Issued</dt>
                  <dd className="text-gray-900">{selectedEngagement.report_issued ? 'Yes' : 'No'}</dd>
                </div>
                {selectedEngagement.overall_opinion && (
                  <div className="flex justify-between">
                    <dt className="text-gray-500">Opinion</dt>
                    <dd className="text-gray-900 font-medium">{selectedEngagement.overall_opinion}</dd>
                  </div>
                )}
              </dl>
            </div>
            <div className="bg-white border rounded-lg p-5">
              <h3 className="font-semibold text-gray-900 mb-3">Objectives</h3>
              <p className="text-sm text-gray-700 whitespace-pre-line">
                {selectedEngagement.objectives || 'No objectives defined.'}
              </p>
            </div>
          </div>
        )}

        {activeTab === 'workpapers' && (
          <div className="bg-white border rounded-lg overflow-hidden">
            <div className="p-4 border-b flex justify-between items-center">
              <h3 className="font-semibold text-gray-900">Workpapers</h3>
              <span className="text-sm text-gray-500">{workpapers.length} workpapers</span>
            </div>
            {workpapers.length === 0 ? (
              <div className="p-8 text-center text-gray-500 text-sm">
                No workpapers yet. Create one to start documenting fieldwork.
              </div>
            ) : (
              <table className="w-full text-sm">
                <thead className="bg-gray-50 text-left">
                  <tr>
                    <th className="px-4 py-3 font-medium text-gray-600">Ref</th>
                    <th className="px-4 py-3 font-medium text-gray-600">Title</th>
                    <th className="px-4 py-3 font-medium text-gray-600">Type</th>
                    <th className="px-4 py-3 font-medium text-gray-600">Status</th>
                    <th className="px-4 py-3 font-medium text-gray-600">Prepared</th>
                    <th className="px-4 py-3 font-medium text-gray-600">Reviewed</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-100">
                  {workpapers.map((wp) => (
                    <tr key={wp.id} className="hover:bg-gray-50">
                      <td className="px-4 py-3 font-mono text-gray-500">{wp.workpaper_ref}</td>
                      <td className="px-4 py-3 text-gray-900">{wp.title}</td>
                      <td className="px-4 py-3 text-gray-600">{wp.workpaper_type}</td>
                      <td className="px-4 py-3">
                        <span className={`px-2 py-0.5 rounded text-xs ${STATUS_COLORS[wp.status] || 'bg-gray-100 text-gray-700'}`}>
                          {wp.status}
                        </span>
                      </td>
                      <td className="px-4 py-3 text-gray-600">{formatDate(wp.prepared_date)}</td>
                      <td className="px-4 py-3 text-gray-600">{formatDate(wp.reviewed_date)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        )}

        {activeTab === 'tests' && (
          <div className="bg-white border rounded-lg overflow-hidden">
            <div className="p-4 border-b">
              <h3 className="font-semibold text-gray-900">Test Procedures</h3>
            </div>
            {testProcedures.length === 0 ? (
              <div className="p-8 text-center text-gray-500 text-sm">
                No test procedures defined yet.
              </div>
            ) : (
              <table className="w-full text-sm">
                <thead className="bg-gray-50 text-left">
                  <tr>
                    <th className="px-4 py-3 font-medium text-gray-600">Ref</th>
                    <th className="px-4 py-3 font-medium text-gray-600">Title</th>
                    <th className="px-4 py-3 font-medium text-gray-600">Type</th>
                    <th className="px-4 py-3 font-medium text-gray-600">Control</th>
                    <th className="px-4 py-3 font-medium text-gray-600">Result</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-100">
                  {testProcedures.map((tp) => (
                    <tr key={tp.id} className="hover:bg-gray-50">
                      <td className="px-4 py-3 font-mono text-gray-500">{tp.procedure_ref}</td>
                      <td className="px-4 py-3 text-gray-900">{tp.title}</td>
                      <td className="px-4 py-3 text-gray-600">{tp.test_type}</td>
                      <td className="px-4 py-3 text-gray-600">{tp.control_ref || '-'}</td>
                      <td className="px-4 py-3">
                        {tp.result ? (
                          <span
                            className={`px-2 py-0.5 rounded text-xs font-medium ${
                              tp.result === 'pass'
                                ? 'bg-green-100 text-green-800'
                                : tp.result === 'fail'
                                  ? 'bg-red-100 text-red-800'
                                  : 'bg-yellow-100 text-yellow-800'
                            }`}
                          >
                            {tp.result}
                          </span>
                        ) : (
                          <span className="text-gray-400">Pending</span>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        )}

        {activeTab === 'samples' && (
          <div className="bg-white border rounded-lg overflow-hidden">
            <div className="p-4 border-b">
              <h3 className="font-semibold text-gray-900">Audit Samples (ISA 530)</h3>
            </div>
            {samples.length === 0 ? (
              <div className="p-8 text-center text-gray-500 text-sm">
                No samples generated yet. Use the API to generate a statistical sample.
              </div>
            ) : (
              <div className="divide-y divide-gray-100">
                {samples.map((s) => (
                  <div key={s.id} className="p-4">
                    <div className="flex items-center justify-between mb-2">
                      <div>
                        <span className="font-mono text-sm text-gray-500 mr-2">{s.sample_ref}</span>
                        <span className="font-medium text-gray-900">{s.name}</span>
                      </div>
                      <span className={`px-2 py-0.5 rounded text-xs ${STATUS_COLORS[s.status] || 'bg-gray-100 text-gray-700'}`}>
                        {s.status}
                      </span>
                    </div>
                    <div className="grid grid-cols-2 md:grid-cols-5 gap-4 text-sm">
                      <div>
                        <span className="text-gray-500">Population:</span>{' '}
                        <span className="text-gray-900">{s.population_size.toLocaleString()}</span>
                      </div>
                      <div>
                        <span className="text-gray-500">Sample Size:</span>{' '}
                        <span className="text-gray-900">{s.sample_size}</span>
                      </div>
                      <div>
                        <span className="text-gray-500">Confidence:</span>{' '}
                        <span className="text-gray-900">{s.confidence_level}%</span>
                      </div>
                      <div>
                        <span className="text-gray-500">Tested:</span>{' '}
                        <span className="text-gray-900">
                          {s.items_tested}/{s.sample_size}
                        </span>
                      </div>
                      <div>
                        <span className="text-gray-500">Error Rate:</span>{' '}
                        <span className="text-gray-900">
                          {s.actual_error_rate !== null
                            ? `${(s.actual_error_rate * 100).toFixed(1)}%`
                            : '-'}
                        </span>
                      </div>
                    </div>
                    {s.items_tested > 0 && (
                      <div className="mt-2 flex gap-4 text-xs">
                        <span className="text-green-700">Passed: {s.items_passed}</span>
                        <span className="text-red-700">Failed: {s.items_failed}</span>
                      </div>
                    )}
                  </div>
                ))}
              </div>
            )}
          </div>
        )}

        {activeTab === 'findings' && (
          <div className="bg-white border rounded-lg p-8 text-center text-gray-500 text-sm">
            Findings are managed from the main Audits module. Linked findings
            for this engagement will appear here.
          </div>
        )}

        {activeTab === 'corrective-actions' && (
          <div className="bg-white border rounded-lg overflow-hidden">
            <div className="p-4 border-b">
              <h3 className="font-semibold text-gray-900">Corrective Actions</h3>
            </div>
            {correctiveActions.length === 0 ? (
              <div className="p-8 text-center text-gray-500 text-sm">
                No corrective actions for this engagement.
              </div>
            ) : (
              <table className="w-full text-sm">
                <thead className="bg-gray-50 text-left">
                  <tr>
                    <th className="px-4 py-3 font-medium text-gray-600">Ref</th>
                    <th className="px-4 py-3 font-medium text-gray-600">Title</th>
                    <th className="px-4 py-3 font-medium text-gray-600">Priority</th>
                    <th className="px-4 py-3 font-medium text-gray-600">Status</th>
                    <th className="px-4 py-3 font-medium text-gray-600">Due</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-100">
                  {correctiveActions.map((ca) => (
                    <tr key={ca.id} className="hover:bg-gray-50">
                      <td className="px-4 py-3 font-mono text-gray-500">{ca.action_ref}</td>
                      <td className="px-4 py-3 text-gray-900">{ca.title}</td>
                      <td className="px-4 py-3">
                        <span className={`px-2 py-0.5 rounded text-xs ${PRIORITY_COLORS[ca.priority] || 'bg-gray-100 text-gray-700'}`}>
                          {ca.priority}
                        </span>
                      </td>
                      <td className="px-4 py-3">
                        <span className={`px-2 py-0.5 rounded text-xs ${STATUS_COLORS[ca.status] || 'bg-gray-100 text-gray-700'}`}>
                          {ca.status}
                        </span>
                      </td>
                      <td className="px-4 py-3 text-gray-600">{formatDate(ca.due_date)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        )}
      </div>
    );
  }

  // --------------------------------------------------------
  // PROGRAMME LIST + ENGAGEMENT LIST
  // --------------------------------------------------------

  if (selectedProgramme) {
    return (
      <div className="p-6 max-w-7xl mx-auto">
        <button
          onClick={() => {
            setSelectedProgramme(null);
            setEngagements([]);
          }}
          className="mb-4 text-sm text-blue-600 hover:text-blue-800 flex items-center gap-1"
        >
          &larr; Back to Programmes
        </button>

        {/* Programme header */}
        <div className="bg-white border rounded-lg p-6 mb-6">
          <div className="flex items-center gap-3 mb-2">
            <span className="font-mono text-sm text-gray-500">{selectedProgramme.programme_ref}</span>
            <span
              className={`px-2 py-0.5 rounded text-xs font-medium ${STATUS_COLORS[selectedProgramme.status] || 'bg-gray-100 text-gray-700'}`}
            >
              {selectedProgramme.status}
            </span>
          </div>
          <h1 className="text-2xl font-bold text-gray-900 mb-2">{selectedProgramme.name}</h1>
          <p className="text-gray-600 mb-4">{selectedProgramme.description}</p>

          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
            <div>
              <span className="text-gray-500 block">Period</span>
              <span className="text-gray-900 font-medium">
                {formatDate(selectedProgramme.period_start)} &mdash;{' '}
                {formatDate(selectedProgramme.period_end)}
              </span>
            </div>
            <div>
              <span className="text-gray-500 block">Type</span>
              <span className="text-gray-900 font-medium">{selectedProgramme.programme_type}</span>
            </div>
            <div>
              <span className="text-gray-500 block">Risk Appetite</span>
              <span className={`font-medium ${RISK_COLORS[selectedProgramme.risk_appetite] || 'text-gray-900'}`}>
                {selectedProgramme.risk_appetite}
              </span>
            </div>
            <div>
              <span className="text-gray-500 block">Budget Utilisation</span>
              <div className="flex items-center gap-2">
                <div className="flex-1 bg-gray-200 rounded-full h-2">
                  <div
                    className="bg-blue-600 h-2 rounded-full"
                    style={{ width: `${Math.min(utilisationPct(selectedProgramme), 100)}%` }}
                  />
                </div>
                <span className="text-gray-900 font-medium text-xs">
                  {selectedProgramme.used_budget_days}/{selectedProgramme.total_budget_days} days (
                  {utilisationPct(selectedProgramme)}%)
                </span>
              </div>
            </div>
          </div>
        </div>

        {/* Engagements */}
        <div className="bg-white border rounded-lg overflow-hidden">
          <div className="p-4 border-b flex justify-between items-center">
            <h2 className="font-semibold text-gray-900">Audit Engagements</h2>
            <span className="text-sm text-gray-500">{engagements.length} engagements</span>
          </div>
          {engagements.length === 0 ? (
            <div className="p-8 text-center text-gray-500 text-sm">
              No engagements in this programme yet.
            </div>
          ) : (
            <table className="w-full text-sm">
              <thead className="bg-gray-50 text-left">
                <tr>
                  <th className="px-4 py-3 font-medium text-gray-600">Ref</th>
                  <th className="px-4 py-3 font-medium text-gray-600">Name</th>
                  <th className="px-4 py-3 font-medium text-gray-600">Type</th>
                  <th className="px-4 py-3 font-medium text-gray-600">Status</th>
                  <th className="px-4 py-3 font-medium text-gray-600">Risk</th>
                  <th className="px-4 py-3 font-medium text-gray-600">Period</th>
                  <th className="px-4 py-3 font-medium text-gray-600">Budget</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {engagements.map((eng) => (
                  <tr
                    key={eng.id}
                    className="hover:bg-gray-50 cursor-pointer"
                    onClick={() => setSelectedEngagement(eng)}
                  >
                    <td className="px-4 py-3 font-mono text-gray-500">{eng.engagement_ref}</td>
                    <td className="px-4 py-3 text-gray-900 font-medium">{eng.name}</td>
                    <td className="px-4 py-3 text-gray-600">{eng.engagement_type}</td>
                    <td className="px-4 py-3">
                      <span
                        className={`px-2 py-0.5 rounded text-xs font-medium ${STATUS_COLORS[eng.status] || 'bg-gray-100 text-gray-700'}`}
                      >
                        {eng.status}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <span className={RISK_COLORS[eng.risk_rating] || 'text-gray-600'}>
                        {eng.risk_rating}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-gray-600 text-xs">
                      {formatDate(eng.planned_start_date)} &mdash; {formatDate(eng.planned_end_date)}
                    </td>
                    <td className="px-4 py-3 text-gray-600">{eng.budget_days}d</td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </div>
    );
  }

  // --------------------------------------------------------
  // PROGRAMME LIST (TOP LEVEL)
  // --------------------------------------------------------

  return (
    <div className="p-6 max-w-7xl mx-auto">
      <div className="flex justify-between items-center mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Audit Programmes</h1>
          <p className="text-gray-500 text-sm mt-1">
            Manage audit programmes, engagements, and the audit lifecycle.
          </p>
        </div>
        <button
          onClick={() => setShowCreateForm(true)}
          className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 text-sm font-medium"
        >
          New Programme
        </button>
      </div>

      {/* Create Programme Form */}
      {showCreateForm && (
        <div className="bg-white border rounded-lg p-6 mb-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Create Audit Programme</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Name</label>
              <input
                type="text"
                value={newProgramme.name}
                onChange={(e) => setNewProgramme({ ...newProgramme, name: e.target.value })}
                className="w-full px-3 py-2 border rounded-lg text-sm"
                placeholder="Annual Internal Audit Programme 2026"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Type</label>
              <select
                value={newProgramme.programme_type}
                onChange={(e) => setNewProgramme({ ...newProgramme, programme_type: e.target.value })}
                className="w-full px-3 py-2 border rounded-lg text-sm"
              >
                <option value="annual">Annual</option>
                <option value="multi_year">Multi-Year</option>
                <option value="special">Special/Ad Hoc</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Period Start</label>
              <input
                type="date"
                value={newProgramme.period_start}
                onChange={(e) => setNewProgramme({ ...newProgramme, period_start: e.target.value })}
                className="w-full px-3 py-2 border rounded-lg text-sm"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Period End</label>
              <input
                type="date"
                value={newProgramme.period_end}
                onChange={(e) => setNewProgramme({ ...newProgramme, period_end: e.target.value })}
                className="w-full px-3 py-2 border rounded-lg text-sm"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Budget (Days)</label>
              <input
                type="number"
                value={newProgramme.total_budget_days}
                onChange={(e) =>
                  setNewProgramme({ ...newProgramme, total_budget_days: parseInt(e.target.value) || 0 })
                }
                className="w-full px-3 py-2 border rounded-lg text-sm"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Risk Appetite</label>
              <select
                value={newProgramme.risk_appetite}
                onChange={(e) => setNewProgramme({ ...newProgramme, risk_appetite: e.target.value })}
                className="w-full px-3 py-2 border rounded-lg text-sm"
              >
                <option value="low">Low</option>
                <option value="medium">Medium</option>
                <option value="high">High</option>
              </select>
            </div>
            <div className="md:col-span-2">
              <label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
              <textarea
                value={newProgramme.description}
                onChange={(e) => setNewProgramme({ ...newProgramme, description: e.target.value })}
                className="w-full px-3 py-2 border rounded-lg text-sm"
                rows={2}
                placeholder="Programme scope and description..."
              />
            </div>
            <div className="md:col-span-2">
              <label className="block text-sm font-medium text-gray-700 mb-1">Objectives</label>
              <textarea
                value={newProgramme.objectives}
                onChange={(e) => setNewProgramme({ ...newProgramme, objectives: e.target.value })}
                className="w-full px-3 py-2 border rounded-lg text-sm"
                rows={2}
                placeholder="Key audit objectives..."
              />
            </div>
          </div>
          <div className="flex gap-3 mt-4">
            <button
              onClick={handleCreateProgramme}
              className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 text-sm font-medium"
            >
              Create Programme
            </button>
            <button
              onClick={() => setShowCreateForm(false)}
              className="px-4 py-2 border text-gray-700 rounded-lg hover:bg-gray-50 text-sm"
            >
              Cancel
            </button>
          </div>
        </div>
      )}

      {/* Programme List */}
      {loading ? (
        <div className="text-center py-12 text-gray-500">Loading audit programmes...</div>
      ) : programmes.length === 0 ? (
        <div className="bg-white border rounded-lg p-12 text-center">
          <h3 className="text-lg font-medium text-gray-900 mb-2">No Audit Programmes</h3>
          <p className="text-gray-500 text-sm mb-4">
            Create your first audit programme to begin planning audit engagements.
          </p>
          <button
            onClick={() => setShowCreateForm(true)}
            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 text-sm font-medium"
          >
            Create Programme
          </button>
        </div>
      ) : (
        <div className="space-y-4">
          {programmes.map((p) => (
            <div
              key={p.id}
              className="bg-white border rounded-lg p-5 hover:shadow-md transition-shadow cursor-pointer"
              onClick={() => setSelectedProgramme(p)}
            >
              <div className="flex justify-between items-start mb-3">
                <div>
                  <div className="flex items-center gap-2 mb-1">
                    <span className="font-mono text-sm text-gray-500">{p.programme_ref}</span>
                    <span
                      className={`px-2 py-0.5 rounded text-xs font-medium ${STATUS_COLORS[p.status] || 'bg-gray-100 text-gray-700'}`}
                    >
                      {p.status}
                    </span>
                    <span className="px-2 py-0.5 rounded text-xs bg-gray-100 text-gray-600">
                      {p.programme_type}
                    </span>
                  </div>
                  <h3 className="text-lg font-semibold text-gray-900">{p.name}</h3>
                  {p.description && (
                    <p className="text-sm text-gray-600 mt-1 line-clamp-2">{p.description}</p>
                  )}
                </div>
                <div className="text-right text-sm">
                  <div className="text-gray-500">
                    {formatDate(p.period_start)} &mdash; {formatDate(p.period_end)}
                  </div>
                  <div className={`mt-1 ${RISK_COLORS[p.risk_appetite] || 'text-gray-600'}`}>
                    Risk appetite: {p.risk_appetite}
                  </div>
                </div>
              </div>

              {/* Budget utilisation bar */}
              <div className="flex items-center gap-3">
                <div className="flex-1 bg-gray-200 rounded-full h-2">
                  <div
                    className={`h-2 rounded-full ${
                      utilisationPct(p) > 90
                        ? 'bg-red-500'
                        : utilisationPct(p) > 70
                          ? 'bg-yellow-500'
                          : 'bg-blue-600'
                    }`}
                    style={{ width: `${Math.min(utilisationPct(p), 100)}%` }}
                  />
                </div>
                <span className="text-xs text-gray-500 whitespace-nowrap">
                  {p.used_budget_days}/{p.total_budget_days} days ({utilisationPct(p)}%)
                </span>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
