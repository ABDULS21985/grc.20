'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import api from '@/lib/api';

// ============================================================
// TYPES
// ============================================================

interface VendorAssessment {
  id: string;
  assessment_ref: string;
  vendor_id: string;
  vendor_name: string;
  questionnaire_name: string;
  status: string;
  overall_score: number | null;
  risk_rating: string | null;
  pass_fail: string;
  due_date: string | null;
  submitted_at: string | null;
  sent_to_email: string;
  reminder_count: number;
  critical_findings: number;
  high_findings: number;
  created_at: string;
}

interface AssessmentDashboard {
  total_assessments: number;
  in_progress: number;
  awaiting_review: number;
  completed: number;
  overdue: number;
  pass_rate: number;
  avg_score: number;
  status_breakdown: Record<string, number>;
  risk_breakdown: Record<string, number>;
  recent_assessments: VendorAssessment[];
  upcoming_due_dates: DueDateEntry[];
  score_distribution: Record<string, number>;
}

interface DueDateEntry {
  assessment_id: string;
  assessment_ref: string;
  vendor_name: string;
  due_date: string;
  days_remaining: number;
  status: string;
}

interface Questionnaire {
  id: string;
  name: string;
  questionnaire_type: string;
  status: string;
  total_questions: number;
  total_sections: number;
  estimated_completion_minutes: number;
  is_system: boolean;
  is_template: boolean;
}

// ============================================================
// HELPERS
// ============================================================

const statusColors: Record<string, string> = {
  draft: 'bg-gray-100 text-gray-800',
  sent: 'bg-blue-100 text-blue-800',
  in_progress: 'bg-yellow-100 text-yellow-800',
  submitted: 'bg-purple-100 text-purple-800',
  under_review: 'bg-indigo-100 text-indigo-800',
  completed: 'bg-green-100 text-green-800',
  expired: 'bg-red-100 text-red-800',
  cancelled: 'bg-gray-100 text-gray-500',
};

const riskColors: Record<string, string> = {
  critical: 'bg-red-100 text-red-800',
  high: 'bg-orange-100 text-orange-800',
  medium: 'bg-yellow-100 text-yellow-800',
  low: 'bg-green-100 text-green-800',
};

const passFailColors: Record<string, string> = {
  pass: 'bg-green-100 text-green-800',
  fail: 'bg-red-100 text-red-800',
  conditional_pass: 'bg-yellow-100 text-yellow-800',
  pending: 'bg-gray-100 text-gray-600',
};

function formatDate(dateStr: string | null): string {
  if (!dateStr) return '-';
  return new Date(dateStr).toLocaleDateString('en-GB', {
    day: '2-digit', month: 'short', year: 'numeric',
  });
}

function formatScore(score: number | null): string {
  if (score === null || score === undefined) return '-';
  return score.toFixed(1) + '%';
}

// ============================================================
// COMPONENT
// ============================================================

export default function VendorAssessmentsPage() {
  const [tab, setTab] = useState<'dashboard' | 'assessments' | 'templates'>('dashboard');
  const [dashboard, setDashboard] = useState<AssessmentDashboard | null>(null);
  const [assessments, setAssessments] = useState<VendorAssessment[]>([]);
  const [templates, setTemplates] = useState<Questionnaire[]>([]);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [statusFilter, setStatusFilter] = useState('');
  const [searchQuery, setSearchQuery] = useState('');

  // ── Load dashboard ──────────────────────────────────────
  useEffect(() => {
    if (tab === 'dashboard') {
      setLoading(true);
      api.request<any>('/vendor-assessments/dashboard')
        .then((res) => setDashboard(res.data))
        .catch(console.error)
        .finally(() => setLoading(false));
    }
  }, [tab]);

  // ── Load assessments ────────────────────────────────────
  useEffect(() => {
    if (tab === 'assessments') {
      setLoading(true);
      const params = new URLSearchParams({ page: String(page), page_size: '20' });
      if (statusFilter) params.set('status', statusFilter);
      if (searchQuery) params.set('search', searchQuery);
      api.request<any>(`/vendor-assessments?${params.toString()}`)
        .then((res) => {
          setAssessments(res.data || []);
          setTotalPages(res.pagination?.total_pages || 1);
        })
        .catch(console.error)
        .finally(() => setLoading(false));
    }
  }, [tab, page, statusFilter, searchQuery]);

  // ── Load templates ──────────────────────────────────────
  useEffect(() => {
    if (tab === 'templates') {
      setLoading(true);
      api.request<any>('/questionnaires?page_size=50&status=active')
        .then((res) => setTemplates(res.data || []))
        .catch(console.error)
        .finally(() => setLoading(false));
    }
  }, [tab]);

  // ============================================================
  // RENDER: DASHBOARD TAB
  // ============================================================

  const renderDashboard = () => {
    if (!dashboard) return <p className="text-gray-500">No data available.</p>;

    return (
      <div className="space-y-6">
        {/* Summary cards */}
        <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-4">
          <SummaryCard label="Total" value={dashboard.total_assessments} color="bg-blue-50" />
          <SummaryCard label="In Progress" value={dashboard.in_progress} color="bg-yellow-50" />
          <SummaryCard label="Awaiting Review" value={dashboard.awaiting_review} color="bg-purple-50" />
          <SummaryCard label="Completed" value={dashboard.completed} color="bg-green-50" />
          <SummaryCard label="Overdue" value={dashboard.overdue} color="bg-red-50" />
          <SummaryCard label="Avg Score" value={`${dashboard.avg_score.toFixed(1)}%`} color="bg-indigo-50" />
        </div>

        {/* Pass rate + score distribution */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <div className="bg-white border border-gray-200 rounded-lg p-6">
            <h3 className="text-sm font-semibold text-gray-700 mb-4">Pass Rate</h3>
            <div className="flex items-center gap-4">
              <div className="text-4xl font-bold text-green-600">{dashboard.pass_rate.toFixed(0)}%</div>
              <div className="flex-1">
                <div className="w-full bg-gray-200 rounded-full h-4">
                  <div
                    className="bg-green-500 h-4 rounded-full"
                    style={{ width: `${Math.min(dashboard.pass_rate, 100)}%` }}
                  />
                </div>
              </div>
            </div>
          </div>

          <div className="bg-white border border-gray-200 rounded-lg p-6">
            <h3 className="text-sm font-semibold text-gray-700 mb-4">Score Distribution</h3>
            <div className="flex items-end gap-2 h-24">
              {['0-20', '20-40', '40-60', '60-80', '80-100'].map((bucket) => {
                const count = dashboard.score_distribution[bucket] || 0;
                const maxCount = Math.max(...Object.values(dashboard.score_distribution), 1);
                const height = (count / maxCount) * 100;
                return (
                  <div key={bucket} className="flex-1 flex flex-col items-center">
                    <span className="text-xs text-gray-500 mb-1">{count}</span>
                    <div
                      className="w-full bg-blue-400 rounded-t"
                      style={{ height: `${Math.max(height, 4)}%` }}
                    />
                    <span className="text-xs text-gray-500 mt-1">{bucket}</span>
                  </div>
                );
              })}
            </div>
          </div>
        </div>

        {/* Risk breakdown + Status breakdown */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <div className="bg-white border border-gray-200 rounded-lg p-6">
            <h3 className="text-sm font-semibold text-gray-700 mb-4">Risk Breakdown</h3>
            <div className="space-y-2">
              {Object.entries(dashboard.risk_breakdown).map(([rating, count]) => (
                <div key={rating} className="flex items-center justify-between">
                  <span className={`px-2 py-1 rounded text-xs font-medium capitalize ${riskColors[rating] || 'bg-gray-100 text-gray-600'}`}>
                    {rating}
                  </span>
                  <span className="text-sm font-medium text-gray-700">{count}</span>
                </div>
              ))}
            </div>
          </div>

          <div className="bg-white border border-gray-200 rounded-lg p-6">
            <h3 className="text-sm font-semibold text-gray-700 mb-4">Status Breakdown</h3>
            <div className="space-y-2">
              {Object.entries(dashboard.status_breakdown).map(([status, count]) => (
                <div key={status} className="flex items-center justify-between">
                  <span className={`px-2 py-1 rounded text-xs font-medium ${statusColors[status] || 'bg-gray-100 text-gray-600'}`}>
                    {status.replace(/_/g, ' ')}
                  </span>
                  <span className="text-sm font-medium text-gray-700">{count}</span>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* Upcoming due dates */}
        {dashboard.upcoming_due_dates.length > 0 && (
          <div className="bg-white border border-gray-200 rounded-lg p-6">
            <h3 className="text-sm font-semibold text-gray-700 mb-4">Upcoming Due Dates</h3>
            <div className="divide-y divide-gray-100">
              {dashboard.upcoming_due_dates.map((d) => (
                <div key={d.assessment_id} className="py-2 flex items-center justify-between">
                  <div>
                    <Link href={`/vendor-assessments/${d.assessment_id}`} className="text-sm font-medium text-blue-600 hover:underline">
                      {d.assessment_ref}
                    </Link>
                    <span className="text-sm text-gray-500 ml-2">{d.vendor_name}</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <span className={`text-xs font-medium ${d.days_remaining <= 7 ? 'text-red-600' : 'text-gray-600'}`}>
                      {d.days_remaining} days
                    </span>
                    <span className="text-xs text-gray-500">{formatDate(d.due_date)}</span>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Recent assessments */}
        {dashboard.recent_assessments.length > 0 && (
          <div className="bg-white border border-gray-200 rounded-lg p-6">
            <h3 className="text-sm font-semibold text-gray-700 mb-4">Recent Assessments</h3>
            <table className="min-w-full text-sm">
              <thead>
                <tr className="text-left text-gray-500 border-b">
                  <th className="pb-2 font-medium">Ref</th>
                  <th className="pb-2 font-medium">Vendor</th>
                  <th className="pb-2 font-medium">Status</th>
                  <th className="pb-2 font-medium">Score</th>
                  <th className="pb-2 font-medium">Risk</th>
                  <th className="pb-2 font-medium">Result</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {dashboard.recent_assessments.map((a) => (
                  <tr key={a.id}>
                    <td className="py-2">
                      <Link href={`/vendor-assessments/${a.id}`} className="text-blue-600 hover:underline">
                        {a.assessment_ref}
                      </Link>
                    </td>
                    <td className="py-2 text-gray-700">{a.vendor_name}</td>
                    <td className="py-2">
                      <span className={`px-2 py-0.5 rounded text-xs font-medium ${statusColors[a.status] || ''}`}>
                        {a.status.replace(/_/g, ' ')}
                      </span>
                    </td>
                    <td className="py-2 font-medium">{formatScore(a.overall_score)}</td>
                    <td className="py-2">
                      {a.risk_rating && (
                        <span className={`px-2 py-0.5 rounded text-xs font-medium capitalize ${riskColors[a.risk_rating] || ''}`}>
                          {a.risk_rating}
                        </span>
                      )}
                    </td>
                    <td className="py-2">
                      <span className={`px-2 py-0.5 rounded text-xs font-medium ${passFailColors[a.pass_fail] || ''}`}>
                        {a.pass_fail.replace(/_/g, ' ')}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    );
  };

  // ============================================================
  // RENDER: ASSESSMENTS TAB
  // ============================================================

  const renderAssessments = () => (
    <div className="space-y-4">
      {/* Filters */}
      <div className="flex flex-wrap gap-3 items-center">
        <input
          type="text"
          placeholder="Search by ref or vendor..."
          value={searchQuery}
          onChange={(e) => { setSearchQuery(e.target.value); setPage(1); }}
          className="px-3 py-2 border border-gray-300 rounded-md text-sm w-64"
        />
        <select
          value={statusFilter}
          onChange={(e) => { setStatusFilter(e.target.value); setPage(1); }}
          className="px-3 py-2 border border-gray-300 rounded-md text-sm"
        >
          <option value="">All Statuses</option>
          <option value="draft">Draft</option>
          <option value="sent">Sent</option>
          <option value="in_progress">In Progress</option>
          <option value="submitted">Submitted</option>
          <option value="under_review">Under Review</option>
          <option value="completed">Completed</option>
          <option value="expired">Expired</option>
        </select>
      </div>

      {/* Table */}
      <div className="bg-white border border-gray-200 rounded-lg overflow-hidden">
        <table className="min-w-full text-sm">
          <thead className="bg-gray-50">
            <tr className="text-left text-gray-600">
              <th className="px-4 py-3 font-medium">Reference</th>
              <th className="px-4 py-3 font-medium">Vendor</th>
              <th className="px-4 py-3 font-medium">Status</th>
              <th className="px-4 py-3 font-medium">Score</th>
              <th className="px-4 py-3 font-medium">Risk</th>
              <th className="px-4 py-3 font-medium">Result</th>
              <th className="px-4 py-3 font-medium">Due Date</th>
              <th className="px-4 py-3 font-medium">Findings</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-100">
            {assessments.length === 0 ? (
              <tr><td colSpan={8} className="px-4 py-8 text-center text-gray-500">No assessments found.</td></tr>
            ) : (
              assessments.map((a) => (
                <tr key={a.id} className="hover:bg-gray-50">
                  <td className="px-4 py-3">
                    <Link href={`/vendor-assessments/${a.id}`} className="text-blue-600 hover:underline font-medium">
                      {a.assessment_ref}
                    </Link>
                  </td>
                  <td className="px-4 py-3 text-gray-700">{a.vendor_name}</td>
                  <td className="px-4 py-3">
                    <span className={`px-2 py-0.5 rounded text-xs font-medium ${statusColors[a.status] || ''}`}>
                      {a.status.replace(/_/g, ' ')}
                    </span>
                  </td>
                  <td className="px-4 py-3 font-medium">{formatScore(a.overall_score)}</td>
                  <td className="px-4 py-3">
                    {a.risk_rating ? (
                      <span className={`px-2 py-0.5 rounded text-xs font-medium capitalize ${riskColors[a.risk_rating] || ''}`}>
                        {a.risk_rating}
                      </span>
                    ) : '-'}
                  </td>
                  <td className="px-4 py-3">
                    <span className={`px-2 py-0.5 rounded text-xs font-medium ${passFailColors[a.pass_fail] || ''}`}>
                      {a.pass_fail.replace(/_/g, ' ')}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-gray-600">{formatDate(a.due_date)}</td>
                  <td className="px-4 py-3">
                    {(a.critical_findings > 0 || a.high_findings > 0) ? (
                      <span className="text-xs">
                        {a.critical_findings > 0 && <span className="text-red-600 font-medium mr-1">{a.critical_findings}C</span>}
                        {a.high_findings > 0 && <span className="text-orange-600 font-medium">{a.high_findings}H</span>}
                      </span>
                    ) : '-'}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex justify-center gap-2">
          <button
            onClick={() => setPage(Math.max(1, page - 1))}
            disabled={page <= 1}
            className="px-3 py-1 text-sm border rounded disabled:opacity-50"
          >
            Previous
          </button>
          <span className="px-3 py-1 text-sm text-gray-600">
            Page {page} of {totalPages}
          </span>
          <button
            onClick={() => setPage(Math.min(totalPages, page + 1))}
            disabled={page >= totalPages}
            className="px-3 py-1 text-sm border rounded disabled:opacity-50"
          >
            Next
          </button>
        </div>
      )}
    </div>
  );

  // ============================================================
  // RENDER: TEMPLATES TAB
  // ============================================================

  const renderTemplates = () => (
    <div className="space-y-4">
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {templates.map((tpl) => (
          <div key={tpl.id} className="bg-white border border-gray-200 rounded-lg p-5">
            <div className="flex items-start justify-between mb-2">
              <h3 className="font-semibold text-gray-900">{tpl.name}</h3>
              {tpl.is_system && (
                <span className="px-2 py-0.5 bg-blue-50 text-blue-700 text-xs rounded font-medium">System</span>
              )}
            </div>
            <div className="flex flex-wrap gap-3 text-xs text-gray-500 mb-3">
              <span>{tpl.total_questions} questions</span>
              <span>{tpl.total_sections} sections</span>
              <span>~{tpl.estimated_completion_minutes} min</span>
              <span className="capitalize">{tpl.questionnaire_type}</span>
            </div>
            <div className="flex gap-2">
              <Link
                href={`/questionnaires/${tpl.id}`}
                className="px-3 py-1 text-xs font-medium text-blue-600 border border-blue-200 rounded hover:bg-blue-50"
              >
                View
              </Link>
              <button
                onClick={async () => {
                  try {
                    await api.request<any>(`/questionnaires/${tpl.id}/clone`, { method: 'POST' });
                    alert('Template cloned successfully.');
                  } catch (e) {
                    alert('Failed to clone template.');
                  }
                }}
                className="px-3 py-1 text-xs font-medium text-gray-600 border border-gray-200 rounded hover:bg-gray-50"
              >
                Clone
              </button>
            </div>
          </div>
        ))}
      </div>
      {templates.length === 0 && (
        <p className="text-gray-500 text-center py-8">No questionnaire templates available.</p>
      )}
    </div>
  );

  // ============================================================
  // MAIN RENDER
  // ============================================================

  return (
    <div className="p-6 max-w-7xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Third-Party Risk Management</h1>
          <p className="text-sm text-gray-500 mt-1">Manage vendor security assessments and questionnaires</p>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex gap-1 mb-6 border-b border-gray-200">
        {(['dashboard', 'assessments', 'templates'] as const).map((t) => (
          <button
            key={t}
            onClick={() => { setTab(t); setPage(1); }}
            className={`px-4 py-2 text-sm font-medium border-b-2 capitalize ${
              tab === t
                ? 'border-blue-500 text-blue-600'
                : 'border-transparent text-gray-500 hover:text-gray-700'
            }`}
          >
            {t}
          </button>
        ))}
      </div>

      {/* Content */}
      {loading ? (
        <div className="flex items-center justify-center py-12">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500" />
          <span className="ml-3 text-gray-500">Loading...</span>
        </div>
      ) : (
        <>
          {tab === 'dashboard' && renderDashboard()}
          {tab === 'assessments' && renderAssessments()}
          {tab === 'templates' && renderTemplates()}
        </>
      )}
    </div>
  );
}

// ============================================================
// SUB-COMPONENTS
// ============================================================

function SummaryCard({ label, value, color }: { label: string; value: string | number; color: string }) {
  return (
    <div className={`${color} rounded-lg p-4 border border-gray-100`}>
      <p className="text-xs text-gray-500 mb-1">{label}</p>
      <p className="text-2xl font-bold text-gray-900">{value}</p>
    </div>
  );
}
