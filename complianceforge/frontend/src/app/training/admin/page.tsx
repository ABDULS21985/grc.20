'use client';

import { useEffect, useState, useCallback } from 'react';
import api from '@/lib/api';

const STATUS_COLORS: Record<string, string> = {
  draft: 'bg-gray-100 text-gray-600',
  active: 'bg-green-100 text-green-800',
  archived: 'bg-yellow-100 text-yellow-800',
};

const CATEGORY_LABELS: Record<string, string> = {
  security_awareness: 'Security Awareness',
  privacy: 'Privacy',
  compliance: 'Compliance',
  technical: 'Technical',
  management: 'Management',
  onboarding: 'Onboarding',
  custom: 'Custom',
};

const AUDIENCE_LABELS: Record<string, string> = {
  all_employees: 'All Employees',
  management: 'Management',
  technical_staff: 'Technical Staff',
  new_joiners: 'New Joiners',
  specific_roles: 'Specific Roles',
  board_members: 'Board Members',
};

interface TrainingProgramme {
  id: string;
  programme_ref: string;
  name: string;
  description: string;
  category: string;
  target_audience: string;
  passing_score: number;
  max_attempts: number;
  duration_minutes: number;
  is_mandatory: boolean;
  recurrence_months: number | null;
  status: string;
  completion_rate: number;
  total_assigned: number;
  total_completed: number;
  created_at: string;
}

interface TrainingDashboard {
  total_programmes: number;
  active_programmes: number;
  total_assignments: number;
  completed_count: number;
  in_progress_count: number;
  overdue_count: number;
  overall_completion_rate: number;
  average_score: number;
  by_category: Record<string, number>;
  by_status: Record<string, number>;
}

interface ComplianceMatrix {
  programmes: { id: string; name: string; ref: string }[];
  users: { id: string; full_name: string; email: string }[];
  cells: { status: string; score: number | null; completed_at: string | null; due_date: string | null }[][];
}

interface PhishingSimulation {
  id: string;
  name: string;
  difficulty: string;
  status: string;
  target_count: number;
  total_sent: number;
  click_rate: number;
  report_rate: number;
  launched_at: string | null;
  completed_at: string | null;
}

interface PhishingTrendPoint {
  simulation_name: string;
  completed_at: string;
  click_rate: number;
  report_rate: number;
}

type Tab = 'programmes' | 'dashboard' | 'matrix' | 'phishing';

export default function TrainingAdminPage() {
  const [tab, setTab] = useState<Tab>('programmes');
  const [programmes, setProgrammes] = useState<TrainingProgramme[]>([]);
  const [dashboard, setDashboard] = useState<TrainingDashboard | null>(null);
  const [matrix, setMatrix] = useState<ComplianceMatrix | null>(null);
  const [simulations, setSimulations] = useState<PhishingSimulation[]>([]);
  const [phishingTrend, setPhishingTrend] = useState<PhishingTrendPoint[]>([]);
  const [loading, setLoading] = useState(true);
  const [totalItems, setTotalItems] = useState(0);
  const [page, setPage] = useState(1);

  // Create programme wizard state
  const [showCreate, setShowCreate] = useState(false);
  const [createStep, setCreateStep] = useState(1);
  const [newProgramme, setNewProgramme] = useState({
    name: '',
    description: '',
    category: 'security_awareness',
    target_audience: 'all_employees',
    passing_score: 80,
    max_attempts: 3,
    duration_minutes: 30,
    is_mandatory: true,
    recurrence_months: 12,
    due_within_days: 30,
  });

  const loadData = useCallback(async () => {
    setLoading(true);
    try {
      if (tab === 'programmes') {
        const res = await api.get(`/training/programmes?page=${page}&page_size=20`);
        setProgrammes(res.data?.data || []);
        setTotalItems(res.data?.pagination?.total_items || 0);
      } else if (tab === 'dashboard') {
        const res = await api.get('/training/dashboard');
        setDashboard(res.data?.data || null);
      } else if (tab === 'matrix') {
        const res = await api.get('/training/compliance-matrix');
        setMatrix(res.data?.data || null);
      } else if (tab === 'phishing') {
        const [simRes, trendRes] = await Promise.all([
          api.get('/training/phishing/simulations?page=1&page_size=20'),
          api.get('/training/phishing/trend'),
        ]);
        setSimulations(simRes.data?.data || []);
        setPhishingTrend(trendRes.data?.data || []);
      }
    } catch {
      /* ignore */
    }
    setLoading(false);
  }, [tab, page]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const handleCreateProgramme = async () => {
    try {
      await api.post('/training/programmes', newProgramme);
      setShowCreate(false);
      setCreateStep(1);
      setNewProgramme({
        name: '',
        description: '',
        category: 'security_awareness',
        target_audience: 'all_employees',
        passing_score: 80,
        max_attempts: 3,
        duration_minutes: 30,
        is_mandatory: true,
        recurrence_months: 12,
        due_within_days: 30,
      });
      loadData();
    } catch {
      /* ignore */
    }
  };

  const handleGenerateAssignments = async (programmeId: string) => {
    try {
      await api.post(`/training/programmes/${programmeId}/generate-assignments`);
      loadData();
    } catch {
      /* ignore */
    }
  };

  const handleExportMatrix = () => {
    window.open('/api/v1/training/compliance-matrix/export', '_blank');
  };

  const MATRIX_CELL_COLORS: Record<string, string> = {
    completed: 'bg-green-500 text-white',
    in_progress: 'bg-yellow-400 text-white',
    assigned: 'bg-blue-400 text-white',
    overdue: 'bg-red-500 text-white',
    failed: 'bg-red-300 text-white',
    exempted: 'bg-gray-300 text-gray-700',
    not_assigned: 'bg-gray-100 text-gray-400',
  };

  const tabs = [
    { key: 'programmes' as const, label: 'Programmes' },
    { key: 'dashboard' as const, label: 'Dashboard' },
    { key: 'matrix' as const, label: 'Compliance Matrix' },
    { key: 'phishing' as const, label: 'Phishing Simulations' },
  ];

  return (
    <div>
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Training Administration</h1>
          <p className="text-gray-500 mt-1">
            Manage training programmes, track compliance, and run phishing simulations
          </p>
        </div>
        <button
          onClick={() => setShowCreate(true)}
          className="px-4 py-2 bg-indigo-600 text-white text-sm rounded-lg hover:bg-indigo-700 transition-colors"
        >
          Create Programme
        </button>
      </div>

      <div className="flex gap-1 border-b border-gray-200 mb-6 overflow-x-auto">
        {tabs.map((t) => (
          <button
            key={t.key}
            onClick={() => { setTab(t.key); setPage(1); }}
            className={`px-4 py-2.5 text-sm font-medium border-b-2 whitespace-nowrap transition-colors ${
              tab === t.key
                ? 'border-indigo-600 text-indigo-600'
                : 'border-transparent text-gray-500 hover:text-gray-700'
            }`}
          >
            {t.label}
          </button>
        ))}
      </div>

      {/* Create Programme Wizard Modal */}
      {showCreate && (
        <div className="fixed inset-0 bg-black/30 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl shadow-xl w-full max-w-lg p-6">
            <h2 className="text-lg font-semibold mb-4">
              Create Training Programme - Step {createStep} of 3
            </h2>

            {/* Step indicators */}
            <div className="flex gap-2 mb-6">
              {[1, 2, 3].map((s) => (
                <div
                  key={s}
                  className={`h-1.5 flex-1 rounded-full ${s <= createStep ? 'bg-indigo-600' : 'bg-gray-200'}`}
                />
              ))}
            </div>

            {createStep === 1 && (
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Programme Name</label>
                  <input
                    type="text"
                    value={newProgramme.name}
                    onChange={(e) => setNewProgramme({ ...newProgramme, name: e.target.value })}
                    className="w-full border rounded-lg px-3 py-2 text-sm"
                    placeholder="e.g., Security Awareness Training 2025"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
                  <textarea
                    value={newProgramme.description}
                    onChange={(e) => setNewProgramme({ ...newProgramme, description: e.target.value })}
                    className="w-full border rounded-lg px-3 py-2 text-sm"
                    rows={3}
                    placeholder="Describe the programme content and objectives..."
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Category</label>
                  <select
                    value={newProgramme.category}
                    onChange={(e) => setNewProgramme({ ...newProgramme, category: e.target.value })}
                    className="w-full border rounded-lg px-3 py-2 text-sm"
                  >
                    {Object.entries(CATEGORY_LABELS).map(([k, v]) => (
                      <option key={k} value={k}>{v}</option>
                    ))}
                  </select>
                </div>
              </div>
            )}

            {createStep === 2 && (
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Target Audience</label>
                  <select
                    value={newProgramme.target_audience}
                    onChange={(e) => setNewProgramme({ ...newProgramme, target_audience: e.target.value })}
                    className="w-full border rounded-lg px-3 py-2 text-sm"
                  >
                    {Object.entries(AUDIENCE_LABELS).map(([k, v]) => (
                      <option key={k} value={k}>{v}</option>
                    ))}
                  </select>
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Passing Score (%)</label>
                    <input
                      type="number"
                      value={newProgramme.passing_score}
                      onChange={(e) => setNewProgramme({ ...newProgramme, passing_score: parseInt(e.target.value) || 80 })}
                      className="w-full border rounded-lg px-3 py-2 text-sm"
                      min={1}
                      max={100}
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Max Attempts</label>
                    <input
                      type="number"
                      value={newProgramme.max_attempts}
                      onChange={(e) => setNewProgramme({ ...newProgramme, max_attempts: parseInt(e.target.value) || 3 })}
                      className="w-full border rounded-lg px-3 py-2 text-sm"
                      min={1}
                    />
                  </div>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Duration (minutes)</label>
                  <input
                    type="number"
                    value={newProgramme.duration_minutes}
                    onChange={(e) => setNewProgramme({ ...newProgramme, duration_minutes: parseInt(e.target.value) || 30 })}
                    className="w-full border rounded-lg px-3 py-2 text-sm"
                    min={5}
                  />
                </div>
              </div>
            )}

            {createStep === 3 && (
              <div className="space-y-4">
                <div className="flex items-center gap-3">
                  <input
                    type="checkbox"
                    checked={newProgramme.is_mandatory}
                    onChange={(e) => setNewProgramme({ ...newProgramme, is_mandatory: e.target.checked })}
                    className="w-4 h-4 text-indigo-600"
                  />
                  <label className="text-sm font-medium text-gray-700">Mandatory Training</label>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Recurrence (months, 0 for one-time)
                  </label>
                  <input
                    type="number"
                    value={newProgramme.recurrence_months}
                    onChange={(e) => setNewProgramme({ ...newProgramme, recurrence_months: parseInt(e.target.value) || 0 })}
                    className="w-full border rounded-lg px-3 py-2 text-sm"
                    min={0}
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Due Within (days)</label>
                  <input
                    type="number"
                    value={newProgramme.due_within_days}
                    onChange={(e) => setNewProgramme({ ...newProgramme, due_within_days: parseInt(e.target.value) || 30 })}
                    className="w-full border rounded-lg px-3 py-2 text-sm"
                    min={1}
                  />
                </div>

                {/* Summary */}
                <div className="p-4 bg-gray-50 rounded-lg">
                  <h4 className="text-sm font-medium text-gray-700 mb-2">Summary</h4>
                  <dl className="space-y-1 text-sm">
                    <div className="flex justify-between">
                      <dt className="text-gray-500">Name:</dt>
                      <dd className="text-gray-900">{newProgramme.name || '-'}</dd>
                    </div>
                    <div className="flex justify-between">
                      <dt className="text-gray-500">Category:</dt>
                      <dd className="text-gray-900">{CATEGORY_LABELS[newProgramme.category]}</dd>
                    </div>
                    <div className="flex justify-between">
                      <dt className="text-gray-500">Audience:</dt>
                      <dd className="text-gray-900">{AUDIENCE_LABELS[newProgramme.target_audience]}</dd>
                    </div>
                    <div className="flex justify-between">
                      <dt className="text-gray-500">Pass Threshold:</dt>
                      <dd className="text-gray-900">{newProgramme.passing_score}%</dd>
                    </div>
                  </dl>
                </div>
              </div>
            )}

            <div className="flex justify-between mt-6">
              <button
                onClick={() => {
                  if (createStep > 1) setCreateStep(createStep - 1);
                  else setShowCreate(false);
                }}
                className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800"
              >
                {createStep > 1 ? 'Back' : 'Cancel'}
              </button>
              {createStep < 3 ? (
                <button
                  onClick={() => setCreateStep(createStep + 1)}
                  disabled={createStep === 1 && !newProgramme.name}
                  className="px-4 py-2 bg-indigo-600 text-white text-sm rounded-lg hover:bg-indigo-700 disabled:opacity-50"
                >
                  Next
                </button>
              ) : (
                <button
                  onClick={handleCreateProgramme}
                  disabled={!newProgramme.name}
                  className="px-4 py-2 bg-green-600 text-white text-sm rounded-lg hover:bg-green-700 disabled:opacity-50"
                >
                  Create Programme
                </button>
              )}
            </div>
          </div>
        </div>
      )}

      {loading ? (
        <div className="card animate-pulse h-96" />
      ) : tab === 'programmes' ? (
        <div>
          <div className="flex justify-between items-center mb-4">
            <p className="text-sm text-gray-500">{totalItems} programmes</p>
          </div>
          <div className="card overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="table-header">
                  <th className="px-4 py-3 text-left">Ref</th>
                  <th className="px-4 py-3 text-left">Name</th>
                  <th className="px-4 py-3 text-left">Category</th>
                  <th className="px-4 py-3 text-left">Audience</th>
                  <th className="px-4 py-3 text-left">Status</th>
                  <th className="px-4 py-3 text-right">Assigned</th>
                  <th className="px-4 py-3 text-right">Completion</th>
                  <th className="px-4 py-3 text-left">Actions</th>
                </tr>
              </thead>
              <tbody>
                {programmes.map((p) => (
                  <tr key={p.id} className="border-t border-gray-100 hover:bg-gray-50">
                    <td className="px-4 py-3 font-mono text-xs text-indigo-600">{p.programme_ref}</td>
                    <td className="px-4 py-3">
                      <p className="font-medium text-gray-900">{p.name}</p>
                      {p.description && (
                        <p className="text-xs text-gray-400 mt-0.5 truncate max-w-xs">{p.description}</p>
                      )}
                    </td>
                    <td className="px-4 py-3 text-gray-500">
                      {CATEGORY_LABELS[p.category] || p.category}
                    </td>
                    <td className="px-4 py-3 text-gray-500">
                      {AUDIENCE_LABELS[p.target_audience] || p.target_audience}
                    </td>
                    <td className="px-4 py-3">
                      <span
                        className={`text-xs px-2 py-0.5 rounded-full ${STATUS_COLORS[p.status] || 'bg-gray-100 text-gray-600'}`}
                      >
                        {p.status}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-right">{p.total_assigned}</td>
                    <td className="px-4 py-3 text-right">
                      <div className="flex items-center justify-end gap-2">
                        <div className="w-16 bg-gray-200 rounded-full h-2">
                          <div
                            className="bg-green-500 h-2 rounded-full"
                            style={{ width: `${Math.min(p.completion_rate, 100)}%` }}
                          />
                        </div>
                        <span className="text-xs text-gray-500">{p.completion_rate.toFixed(0)}%</span>
                      </div>
                    </td>
                    <td className="px-4 py-3">
                      {p.status === 'active' && (
                        <button
                          onClick={() => handleGenerateAssignments(p.id)}
                          className="text-xs text-indigo-600 hover:text-indigo-800 font-medium"
                        >
                          Assign
                        </button>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      ) : tab === 'dashboard' && dashboard ? (
        <div>
          {/* KPI Cards */}
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
            <div className="card p-4">
              <p className="text-xs text-gray-500 uppercase tracking-wide">Active Programmes</p>
              <p className="text-2xl font-bold text-gray-900 mt-1">{dashboard.active_programmes}</p>
            </div>
            <div className="card p-4">
              <p className="text-xs text-gray-500 uppercase tracking-wide">Completion Rate</p>
              <p className="text-2xl font-bold text-green-600 mt-1">
                {dashboard.overall_completion_rate.toFixed(1)}%
              </p>
            </div>
            <div className="card p-4">
              <p className="text-xs text-gray-500 uppercase tracking-wide">Average Score</p>
              <p className="text-2xl font-bold text-indigo-600 mt-1">
                {dashboard.average_score.toFixed(1)}%
              </p>
            </div>
            <div className="card p-4">
              <p className="text-xs text-gray-500 uppercase tracking-wide">Overdue</p>
              <p className="text-2xl font-bold text-red-600 mt-1">{dashboard.overdue_count}</p>
            </div>
          </div>

          {/* Status Breakdown */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="card p-6">
              <h3 className="font-semibold text-gray-800 mb-4">By Status</h3>
              <div className="space-y-3">
                {Object.entries(dashboard.by_status).map(([status, count]) => (
                  <div key={status} className="flex items-center justify-between">
                    <span className="text-sm text-gray-600 capitalize">{status.replace('_', ' ')}</span>
                    <div className="flex items-center gap-2">
                      <div className="w-32 bg-gray-200 rounded-full h-2">
                        <div
                          className="bg-indigo-500 h-2 rounded-full"
                          style={{
                            width: `${dashboard.total_assignments > 0 ? (count / dashboard.total_assignments) * 100 : 0}%`,
                          }}
                        />
                      </div>
                      <span className="text-sm font-medium text-gray-700 w-8 text-right">{count}</span>
                    </div>
                  </div>
                ))}
              </div>
            </div>
            <div className="card p-6">
              <h3 className="font-semibold text-gray-800 mb-4">By Category</h3>
              <div className="space-y-3">
                {Object.entries(dashboard.by_category).map(([cat, count]) => (
                  <div key={cat} className="flex items-center justify-between">
                    <span className="text-sm text-gray-600">{CATEGORY_LABELS[cat] || cat}</span>
                    <span className="text-sm font-medium text-gray-700">{count}</span>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      ) : tab === 'matrix' && matrix ? (
        <div>
          <div className="flex justify-between items-center mb-4">
            <p className="text-sm text-gray-500">
              {matrix.users.length} users x {matrix.programmes.length} programmes
            </p>
            <button
              onClick={handleExportMatrix}
              className="px-3 py-1.5 bg-gray-100 text-gray-700 text-sm rounded-lg hover:bg-gray-200"
            >
              Export CSV
            </button>
          </div>
          <div className="card overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="table-header">
                  <th className="px-4 py-3 text-left sticky left-0 bg-gray-50 z-10">User</th>
                  {matrix.programmes.map((p) => (
                    <th key={p.id} className="px-3 py-3 text-center min-w-[120px]">
                      <div className="text-xs">{p.ref}</div>
                      <div className="text-xs text-gray-500 truncate max-w-[100px]">{p.name}</div>
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {matrix.users.map((u, i) => (
                  <tr key={u.id} className="border-t border-gray-100">
                    <td className="px-4 py-2 sticky left-0 bg-white z-10">
                      <p className="font-medium text-gray-900 text-xs">{u.full_name}</p>
                      <p className="text-xs text-gray-400">{u.email}</p>
                    </td>
                    {matrix.cells[i]?.map((cell, j) => (
                      <td key={j} className="px-1 py-1 text-center">
                        <span
                          className={`inline-block w-full text-xs px-2 py-1 rounded ${MATRIX_CELL_COLORS[cell.status] || 'bg-gray-100 text-gray-400'}`}
                          title={`${cell.status}${cell.score !== null ? ` (${cell.score}%)` : ''}`}
                        >
                          {cell.status === 'completed'
                            ? cell.score !== null
                              ? `${cell.score}%`
                              : 'Done'
                            : cell.status === 'not_assigned'
                              ? '-'
                              : cell.status.replace('_', ' ')}
                        </span>
                      </td>
                    ))}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      ) : tab === 'phishing' ? (
        <div>
          {/* Phishing Trend */}
          {phishingTrend.length > 0 && (
            <div className="card p-6 mb-6">
              <h3 className="font-semibold text-gray-800 mb-4">Click Rate Trend</h3>
              <div className="flex items-end gap-2 h-40">
                {phishingTrend.map((point, i) => (
                  <div key={i} className="flex-1 flex flex-col items-center">
                    <div className="w-full flex flex-col items-center gap-1">
                      <div
                        className="w-full bg-red-400 rounded-t"
                        style={{ height: `${Math.max(point.click_rate, 2)}%` }}
                        title={`Click: ${point.click_rate.toFixed(1)}%`}
                      />
                      <div
                        className="w-full bg-green-400 rounded-t"
                        style={{ height: `${Math.max(point.report_rate, 2)}%` }}
                        title={`Report: ${point.report_rate.toFixed(1)}%`}
                      />
                    </div>
                    <p className="text-xs text-gray-400 mt-1 truncate max-w-[80px]">
                      {point.simulation_name}
                    </p>
                  </div>
                ))}
              </div>
              <div className="flex gap-4 mt-2 text-xs text-gray-500">
                <span className="flex items-center gap-1">
                  <span className="w-3 h-3 bg-red-400 rounded" /> Click Rate
                </span>
                <span className="flex items-center gap-1">
                  <span className="w-3 h-3 bg-green-400 rounded" /> Report Rate
                </span>
              </div>
            </div>
          )}

          {/* Simulation List */}
          <div className="card overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="table-header">
                  <th className="px-4 py-3 text-left">Name</th>
                  <th className="px-4 py-3 text-left">Difficulty</th>
                  <th className="px-4 py-3 text-left">Status</th>
                  <th className="px-4 py-3 text-right">Targets</th>
                  <th className="px-4 py-3 text-right">Click Rate</th>
                  <th className="px-4 py-3 text-right">Report Rate</th>
                  <th className="px-4 py-3 text-left">Launched</th>
                </tr>
              </thead>
              <tbody>
                {simulations.map((s) => (
                  <tr key={s.id} className="border-t border-gray-100 hover:bg-gray-50">
                    <td className="px-4 py-3 font-medium text-gray-900">{s.name}</td>
                    <td className="px-4 py-3">
                      <span className={`text-xs px-2 py-0.5 rounded-full ${
                        s.difficulty === 'easy' ? 'bg-green-100 text-green-800' :
                        s.difficulty === 'medium' ? 'bg-yellow-100 text-yellow-800' :
                        s.difficulty === 'hard' ? 'bg-orange-100 text-orange-800' :
                        'bg-red-100 text-red-800'
                      }`}>
                        {s.difficulty}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <span className={`text-xs px-2 py-0.5 rounded-full ${STATUS_COLORS[s.status] || 'bg-gray-100 text-gray-600'}`}>
                        {s.status}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-right">{s.target_count || s.total_sent}</td>
                    <td className="px-4 py-3 text-right">
                      {s.status === 'completed' ? `${s.click_rate.toFixed(1)}%` : '-'}
                    </td>
                    <td className="px-4 py-3 text-right">
                      {s.status === 'completed' ? `${s.report_rate.toFixed(1)}%` : '-'}
                    </td>
                    <td className="px-4 py-3 text-gray-500">
                      {s.launched_at ? new Date(s.launched_at).toLocaleDateString() : '-'}
                    </td>
                  </tr>
                ))}
                {simulations.length === 0 && (
                  <tr>
                    <td colSpan={7} className="px-4 py-8 text-center text-gray-400">
                      No phishing simulations yet
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </div>
      ) : null}
    </div>
  );
}
