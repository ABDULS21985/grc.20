'use client';

import { useEffect, useState, useCallback } from 'react';
import api from '@/lib/api';

// ============================================================
// TYPES
// ============================================================

interface ComplianceException {
  id: string;
  exception_ref: string;
  title: string;
  description: string;
  exception_type: string;
  status: string;
  priority: string;
  scope_type: string;
  residual_risk_level: string;
  has_compensating_controls: boolean;
  compensating_effectiveness: string;
  requested_by: string;
  requested_at: string;
  approved_by: string | null;
  approved_at: string | null;
  effective_date: string;
  expiry_date: string | null;
  next_review_date: string | null;
  renewal_count: number;
  framework_control_codes: string[];
  tags: string[];
  created_at: string;
}

interface ExceptionDashboard {
  active_count: number;
  draft_count: number;
  pending_approval_count: number;
  rejected_count: number;
  expired_count: number;
  by_risk_level: Record<string, number>;
  by_priority: Record<string, number>;
  by_type: Record<string, number>;
  expiring_30_days: number;
  expiring_60_days: number;
  expiring_90_days: number;
  overdue_reviews: number;
  average_age_days: number;
  top_excepted_frameworks: { framework_code: string; count: number }[];
  recent_exceptions: ComplianceException[];
}

// ============================================================
// CONSTANTS
// ============================================================

const STATUS_COLORS: Record<string, string> = {
  draft: 'bg-gray-100 text-gray-700',
  pending_risk_assessment: 'bg-yellow-100 text-yellow-800',
  pending_approval: 'bg-blue-100 text-blue-800',
  approved: 'bg-green-100 text-green-800',
  rejected: 'bg-red-100 text-red-800',
  expired: 'bg-orange-100 text-orange-800',
  revoked: 'bg-red-200 text-red-900',
  renewal_pending: 'bg-amber-100 text-amber-800',
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

const STATUS_OPTIONS = [
  { value: '', label: 'All Statuses' },
  { value: 'draft', label: 'Draft' },
  { value: 'pending_risk_assessment', label: 'Pending Risk Assessment' },
  { value: 'pending_approval', label: 'Pending Approval' },
  { value: 'approved', label: 'Approved' },
  { value: 'rejected', label: 'Rejected' },
  { value: 'expired', label: 'Expired' },
  { value: 'revoked', label: 'Revoked' },
];

const PRIORITY_OPTIONS = [
  { value: '', label: 'All Priorities' },
  { value: 'critical', label: 'Critical' },
  { value: 'high', label: 'High' },
  { value: 'medium', label: 'Medium' },
  { value: 'low', label: 'Low' },
];

const TYPE_OPTIONS = [
  { value: '', label: 'All Types' },
  { value: 'temporary', label: 'Temporary' },
  { value: 'permanent', label: 'Permanent' },
  { value: 'conditional', label: 'Conditional' },
];

// ============================================================
// COMPONENT
// ============================================================

export default function ExceptionsPage() {
  const [dashboard, setDashboard] = useState<ExceptionDashboard | null>(null);
  const [exceptions, setExceptions] = useState<ComplianceException[]>([]);
  const [loading, setLoading] = useState(true);
  const [listLoading, setListLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [totalItems, setTotalItems] = useState(0);

  // Filters
  const [statusFilter, setStatusFilter] = useState('');
  const [priorityFilter, setPriorityFilter] = useState('');
  const [typeFilter, setTypeFilter] = useState('');
  const [searchQuery, setSearchQuery] = useState('');

  // Create modal
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [createStep, setCreateStep] = useState(1);
  const [createForm, setCreateForm] = useState({
    title: '',
    description: '',
    exception_type: 'temporary',
    priority: 'medium',
    scope_type: 'control_implementation',
    risk_justification: '',
    residual_risk_level: 'medium',
    effective_date: '',
    expiry_date: '',
    review_frequency_months: 12,
    has_compensating_controls: false,
    compensating_controls_description: '',
    compensating_effectiveness: 'not_assessed',
    conditions: '',
    business_impact_if_implemented: '',
    regulatory_notification_required: false,
    tags: '',
  });
  const [createError, setCreateError] = useState('');
  const [creating, setCreating] = useState(false);

  // ── Data Fetching ─────────────────────────────────────────

  const loadDashboard = useCallback(async () => {
    try {
      const res = await api.request<any>('/exceptions/dashboard');
      setDashboard(res.data);
    } catch (err: any) {
      console.error('Failed to load dashboard:', err);
    }
  }, []);

  const loadExceptions = useCallback(async () => {
    setListLoading(true);
    try {
      const params = new URLSearchParams({ page: String(page), page_size: '20' });
      if (statusFilter) params.set('status', statusFilter);
      if (priorityFilter) params.set('priority', priorityFilter);
      if (typeFilter) params.set('exception_type', typeFilter);
      if (searchQuery) params.set('search', searchQuery);

      const res = await api.request<any>(`/exceptions?${params.toString()}`);
      setExceptions(res.data || []);
      if (res.pagination) {
        setTotalPages(res.pagination.total_pages || 1);
        setTotalItems(res.pagination.total_items || 0);
      }
    } catch (err: any) {
      setError(err.message);
    } finally {
      setListLoading(false);
    }
  }, [page, statusFilter, priorityFilter, typeFilter, searchQuery]);

  useEffect(() => {
    Promise.all([loadDashboard(), loadExceptions()])
      .finally(() => setLoading(false));
  }, [loadDashboard, loadExceptions]);

  // ── Create Exception Handler ──────────────────────────────

  const handleCreate = async () => {
    setCreateError('');
    if (!createForm.title) { setCreateError('Title is required'); return; }
    if (!createForm.description) { setCreateError('Description is required'); return; }
    if (!createForm.risk_justification) { setCreateError('Risk justification is required'); return; }
    if (!createForm.effective_date) { setCreateError('Effective date is required'); return; }

    setCreating(true);
    try {
      const body: any = {
        ...createForm,
        review_frequency_months: Number(createForm.review_frequency_months),
        tags: createForm.tags ? createForm.tags.split(',').map((t: string) => t.trim()).filter(Boolean) : [],
      };
      await api.request<any>('/exceptions', { method: 'POST', body });
      setShowCreateModal(false);
      setCreateStep(1);
      setCreateForm({
        title: '', description: '', exception_type: 'temporary', priority: 'medium',
        scope_type: 'control_implementation', risk_justification: '', residual_risk_level: 'medium',
        effective_date: '', expiry_date: '', review_frequency_months: 12,
        has_compensating_controls: false, compensating_controls_description: '',
        compensating_effectiveness: 'not_assessed', conditions: '',
        business_impact_if_implemented: '', regulatory_notification_required: false, tags: '',
      });
      loadDashboard();
      loadExceptions();
    } catch (err: any) {
      setCreateError(err.message || 'Failed to create exception');
    } finally {
      setCreating(false);
    }
  };

  // ── Helpers ───────────────────────────────────────────────

  const formatDate = (d: string | null) => {
    if (!d) return '-';
    return new Date(d).toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric' });
  };

  const formatStatus = (s: string) => s.replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase());

  const daysUntil = (d: string | null) => {
    if (!d) return null;
    const diff = Math.ceil((new Date(d).getTime() - Date.now()) / (1000 * 60 * 60 * 24));
    return diff;
  };

  // ── Risk distribution chart (simple bar) ──────────────────

  const renderRiskDistribution = () => {
    if (!dashboard) return null;
    const levels = ['critical', 'high', 'medium', 'low', 'very_low'];
    const colors: Record<string, string> = {
      critical: 'bg-red-500', high: 'bg-orange-500', medium: 'bg-yellow-500',
      low: 'bg-blue-500', very_low: 'bg-gray-400',
    };
    const total = Object.values(dashboard.by_risk_level).reduce((a, b) => a + b, 0) || 1;

    return (
      <div className="space-y-2">
        {levels.map(level => {
          const count = dashboard.by_risk_level[level] || 0;
          const pct = Math.round((count / total) * 100);
          return (
            <div key={level} className="flex items-center gap-2">
              <span className="w-20 text-xs text-gray-600 capitalize">{level.replace('_', ' ')}</span>
              <div className="flex-1 bg-gray-100 rounded-full h-4 overflow-hidden">
                <div className={`h-full rounded-full ${colors[level]}`} style={{ width: `${pct}%` }} />
              </div>
              <span className="text-xs text-gray-500 w-12 text-right">{count} ({pct}%)</span>
            </div>
          );
        })}
      </div>
    );
  };

  // ── Loading State ─────────────────────────────────────────

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <p className="text-gray-500">Loading exception management...</p>
      </div>
    );
  }

  if (error && !exceptions.length) {
    return (
      <div className="rounded-lg bg-red-50 border border-red-200 p-4 text-red-700">{error}</div>
    );
  }

  // ── Render ────────────────────────────────────────────────

  return (
    <div>
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Exception Management</h1>
          <p className="text-gray-500 text-sm mt-1">
            Manage compliance exceptions and compensating controls
          </p>
        </div>
        <button
          onClick={() => setShowCreateModal(true)}
          className="bg-indigo-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-indigo-700 transition"
        >
          + New Exception
        </button>
      </div>

      {/* KPI Cards */}
      {dashboard && (
        <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-4 mb-6">
          <KPICard label="Active" value={dashboard.active_count} color="text-green-700" bg="bg-green-50" />
          <KPICard label="Draft" value={dashboard.draft_count} color="text-gray-700" bg="bg-gray-50" />
          <KPICard label="Pending Approval" value={dashboard.pending_approval_count} color="text-blue-700" bg="bg-blue-50" />
          <KPICard label="Expiring (30d)" value={dashboard.expiring_30_days} color="text-orange-700" bg="bg-orange-50" />
          <KPICard label="Overdue Reviews" value={dashboard.overdue_reviews} color="text-red-700" bg="bg-red-50" />
          <KPICard label="Avg Age (days)" value={dashboard.average_age_days} color="text-indigo-700" bg="bg-indigo-50" />
        </div>
      )}

      {/* Risk Distribution & Type Breakdown */}
      {dashboard && (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-6">
          <div className="bg-white border border-gray-200 rounded-lg p-4">
            <h3 className="text-sm font-semibold text-gray-700 mb-3">Risk Distribution (Active)</h3>
            {renderRiskDistribution()}
          </div>
          <div className="bg-white border border-gray-200 rounded-lg p-4">
            <h3 className="text-sm font-semibold text-gray-700 mb-3">Expiry Horizon</h3>
            <div className="space-y-3">
              <div className="flex justify-between items-center">
                <span className="text-sm text-gray-600">Within 30 days</span>
                <span className="text-sm font-semibold text-red-600">{dashboard.expiring_30_days}</span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-sm text-gray-600">Within 60 days</span>
                <span className="text-sm font-semibold text-orange-600">{dashboard.expiring_60_days}</span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-sm text-gray-600">Within 90 days</span>
                <span className="text-sm font-semibold text-yellow-600">{dashboard.expiring_90_days}</span>
              </div>
              <hr className="my-2" />
              <div className="flex justify-between items-center">
                <span className="text-sm text-gray-600">Expired</span>
                <span className="text-sm font-semibold text-gray-500">{dashboard.expired_count}</span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-sm text-gray-600">Rejected</span>
                <span className="text-sm font-semibold text-gray-500">{dashboard.rejected_count}</span>
              </div>
            </div>
            {dashboard.top_excepted_frameworks.length > 0 && (
              <>
                <h3 className="text-sm font-semibold text-gray-700 mt-4 mb-2">Top Excepted Frameworks</h3>
                <div className="space-y-1">
                  {dashboard.top_excepted_frameworks.slice(0, 5).map(fw => (
                    <div key={fw.framework_code} className="flex justify-between text-sm">
                      <span className="text-gray-600 font-mono">{fw.framework_code}</span>
                      <span className="text-gray-800 font-semibold">{fw.count}</span>
                    </div>
                  ))}
                </div>
              </>
            )}
          </div>
        </div>
      )}

      {/* Filters */}
      <div className="bg-white border border-gray-200 rounded-lg p-4 mb-4">
        <div className="grid grid-cols-1 md:grid-cols-5 gap-3">
          <input
            type="text"
            placeholder="Search exceptions..."
            value={searchQuery}
            onChange={e => { setSearchQuery(e.target.value); setPage(1); }}
            className="border border-gray-300 rounded-md px-3 py-2 text-sm"
          />
          <select
            value={statusFilter}
            onChange={e => { setStatusFilter(e.target.value); setPage(1); }}
            className="border border-gray-300 rounded-md px-3 py-2 text-sm"
          >
            {STATUS_OPTIONS.map(o => <option key={o.value} value={o.value}>{o.label}</option>)}
          </select>
          <select
            value={priorityFilter}
            onChange={e => { setPriorityFilter(e.target.value); setPage(1); }}
            className="border border-gray-300 rounded-md px-3 py-2 text-sm"
          >
            {PRIORITY_OPTIONS.map(o => <option key={o.value} value={o.value}>{o.label}</option>)}
          </select>
          <select
            value={typeFilter}
            onChange={e => { setTypeFilter(e.target.value); setPage(1); }}
            className="border border-gray-300 rounded-md px-3 py-2 text-sm"
          >
            {TYPE_OPTIONS.map(o => <option key={o.value} value={o.value}>{o.label}</option>)}
          </select>
          <div className="text-sm text-gray-500 flex items-center">
            {totalItems} exception{totalItems !== 1 ? 's' : ''} found
          </div>
        </div>
      </div>

      {/* Exception List Table */}
      <div className="bg-white border border-gray-200 rounded-lg overflow-hidden">
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Ref</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Title</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Priority</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Risk</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Type</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Expiry</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Comp. Controls</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {listLoading ? (
                <tr>
                  <td colSpan={8} className="px-4 py-8 text-center text-gray-500">Loading...</td>
                </tr>
              ) : exceptions.length === 0 ? (
                <tr>
                  <td colSpan={8} className="px-4 py-8 text-center text-gray-500">
                    No exceptions found. Click &quot;+ New Exception&quot; to create one.
                  </td>
                </tr>
              ) : (
                exceptions.map(exc => {
                  const expiryDays = daysUntil(exc.expiry_date);
                  return (
                    <tr key={exc.id} className="hover:bg-gray-50 cursor-pointer" onClick={() => window.location.href = `/exceptions/${exc.id}`}>
                      <td className="px-4 py-3 text-sm font-mono text-gray-500">{exc.exception_ref}</td>
                      <td className="px-4 py-3 text-sm font-medium text-gray-900 max-w-xs truncate">{exc.title}</td>
                      <td className="px-4 py-3">
                        <span className={`inline-block px-2 py-0.5 rounded text-xs font-medium ${STATUS_COLORS[exc.status] || 'bg-gray-100 text-gray-700'}`}>
                          {formatStatus(exc.status)}
                        </span>
                      </td>
                      <td className="px-4 py-3">
                        <span className={`inline-block px-2 py-0.5 rounded text-xs font-medium ${PRIORITY_COLORS[exc.priority] || ''}`}>
                          {exc.priority}
                        </span>
                      </td>
                      <td className="px-4 py-3">
                        <span className={`text-sm capitalize ${RISK_COLORS[exc.residual_risk_level] || ''}`}>
                          {exc.residual_risk_level?.replace('_', ' ')}
                        </span>
                      </td>
                      <td className="px-4 py-3 text-sm text-gray-600 capitalize">{exc.exception_type}</td>
                      <td className="px-4 py-3 text-sm">
                        {exc.expiry_date ? (
                          <span className={expiryDays !== null && expiryDays <= 30 ? 'text-red-600 font-semibold' : 'text-gray-600'}>
                            {formatDate(exc.expiry_date)}
                            {expiryDays !== null && expiryDays > 0 && (
                              <span className="text-xs ml-1">({expiryDays}d)</span>
                            )}
                          </span>
                        ) : (
                          <span className="text-gray-400">No expiry</span>
                        )}
                      </td>
                      <td className="px-4 py-3 text-sm">
                        {exc.has_compensating_controls ? (
                          <span className="text-green-600">Yes</span>
                        ) : (
                          <span className="text-gray-400">No</span>
                        )}
                      </td>
                    </tr>
                  );
                })
              )}
            </tbody>
          </table>
        </div>

        {/* Pagination */}
        {totalPages > 1 && (
          <div className="flex items-center justify-between px-4 py-3 border-t border-gray-200 bg-gray-50">
            <button
              onClick={() => setPage(p => Math.max(1, p - 1))}
              disabled={page <= 1}
              className="px-3 py-1 text-sm rounded border border-gray-300 disabled:opacity-50 hover:bg-gray-100"
            >
              Previous
            </button>
            <span className="text-sm text-gray-600">
              Page {page} of {totalPages}
            </span>
            <button
              onClick={() => setPage(p => Math.min(totalPages, p + 1))}
              disabled={page >= totalPages}
              className="px-3 py-1 text-sm rounded border border-gray-300 disabled:opacity-50 hover:bg-gray-100"
            >
              Next
            </button>
          </div>
        )}
      </div>

      {/* Create Exception Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 z-50 flex items-center justify-center p-4">
          <div className="bg-white rounded-xl shadow-xl w-full max-w-2xl max-h-[90vh] overflow-y-auto">
            <div className="p-6">
              <div className="flex items-center justify-between mb-6">
                <h2 className="text-lg font-bold text-gray-900">New Compliance Exception</h2>
                <button onClick={() => { setShowCreateModal(false); setCreateStep(1); setCreateError(''); }}
                  className="text-gray-400 hover:text-gray-600 text-xl">&times;</button>
              </div>

              {/* Step indicator */}
              <div className="flex items-center gap-2 mb-6">
                {[1, 2, 3].map(step => (
                  <div key={step} className="flex items-center gap-2">
                    <div className={`w-8 h-8 rounded-full flex items-center justify-center text-sm font-medium ${
                      createStep >= step ? 'bg-indigo-600 text-white' : 'bg-gray-200 text-gray-500'
                    }`}>{step}</div>
                    <span className={`text-xs ${createStep >= step ? 'text-indigo-600' : 'text-gray-400'}`}>
                      {step === 1 ? 'Details' : step === 2 ? 'Risk & Scope' : 'Controls'}
                    </span>
                    {step < 3 && <div className="w-8 h-px bg-gray-300" />}
                  </div>
                ))}
              </div>

              {createError && (
                <div className="bg-red-50 border border-red-200 rounded-md p-3 text-sm text-red-700 mb-4">{createError}</div>
              )}

              {/* Step 1: Basic Details */}
              {createStep === 1 && (
                <div className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Title *</label>
                    <input type="text" value={createForm.title}
                      onChange={e => setCreateForm({ ...createForm, title: e.target.value })}
                      className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm"
                      placeholder="e.g., Legacy System Authentication Exception" />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Description *</label>
                    <textarea value={createForm.description}
                      onChange={e => setCreateForm({ ...createForm, description: e.target.value })}
                      className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm" rows={3}
                      placeholder="Describe why this exception is needed" />
                  </div>
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">Type</label>
                      <select value={createForm.exception_type}
                        onChange={e => setCreateForm({ ...createForm, exception_type: e.target.value })}
                        className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm">
                        <option value="temporary">Temporary</option>
                        <option value="permanent">Permanent</option>
                        <option value="conditional">Conditional</option>
                      </select>
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">Priority</label>
                      <select value={createForm.priority}
                        onChange={e => setCreateForm({ ...createForm, priority: e.target.value })}
                        className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm">
                        <option value="critical">Critical</option>
                        <option value="high">High</option>
                        <option value="medium">Medium</option>
                        <option value="low">Low</option>
                      </select>
                    </div>
                  </div>
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">Effective Date *</label>
                      <input type="date" value={createForm.effective_date}
                        onChange={e => setCreateForm({ ...createForm, effective_date: e.target.value })}
                        className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm" />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">Expiry Date</label>
                      <input type="date" value={createForm.expiry_date}
                        onChange={e => setCreateForm({ ...createForm, expiry_date: e.target.value })}
                        className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm" />
                    </div>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Tags (comma-separated)</label>
                    <input type="text" value={createForm.tags}
                      onChange={e => setCreateForm({ ...createForm, tags: e.target.value })}
                      className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm"
                      placeholder="legacy, authentication, erp" />
                  </div>
                </div>
              )}

              {/* Step 2: Risk & Scope */}
              {createStep === 2 && (
                <div className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Risk Justification *</label>
                    <textarea value={createForm.risk_justification}
                      onChange={e => setCreateForm({ ...createForm, risk_justification: e.target.value })}
                      className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm" rows={3}
                      placeholder="Explain why the risk is acceptable" />
                  </div>
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">Residual Risk Level</label>
                      <select value={createForm.residual_risk_level}
                        onChange={e => setCreateForm({ ...createForm, residual_risk_level: e.target.value })}
                        className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm">
                        <option value="critical">Critical</option>
                        <option value="high">High</option>
                        <option value="medium">Medium</option>
                        <option value="low">Low</option>
                        <option value="very_low">Very Low</option>
                      </select>
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">Scope Type</label>
                      <select value={createForm.scope_type}
                        onChange={e => setCreateForm({ ...createForm, scope_type: e.target.value })}
                        className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm">
                        <option value="control_implementation">Control Implementation</option>
                        <option value="framework_control">Framework Control</option>
                        <option value="policy">Policy</option>
                        <option value="process">Process</option>
                        <option value="system">System</option>
                        <option value="organization_wide">Organization Wide</option>
                      </select>
                    </div>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Business Impact if Implemented</label>
                    <textarea value={createForm.business_impact_if_implemented}
                      onChange={e => setCreateForm({ ...createForm, business_impact_if_implemented: e.target.value })}
                      className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm" rows={2}
                      placeholder="What would happen if the control were fully implemented now?" />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Conditions</label>
                    <textarea value={createForm.conditions}
                      onChange={e => setCreateForm({ ...createForm, conditions: e.target.value })}
                      className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm" rows={2}
                      placeholder="Any conditions that must be met for this exception" />
                  </div>
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">Review Frequency (months)</label>
                      <input type="number" min={1} max={36} value={createForm.review_frequency_months}
                        onChange={e => setCreateForm({ ...createForm, review_frequency_months: parseInt(e.target.value) || 12 })}
                        className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm" />
                    </div>
                    <div className="flex items-center pt-6">
                      <label className="flex items-center gap-2 text-sm text-gray-700">
                        <input type="checkbox" checked={createForm.regulatory_notification_required}
                          onChange={e => setCreateForm({ ...createForm, regulatory_notification_required: e.target.checked })}
                          className="rounded border-gray-300" />
                        Regulatory notification required
                      </label>
                    </div>
                  </div>
                </div>
              )}

              {/* Step 3: Compensating Controls */}
              {createStep === 3 && (
                <div className="space-y-4">
                  <div className="flex items-center gap-2">
                    <label className="flex items-center gap-2 text-sm font-medium text-gray-700">
                      <input type="checkbox" checked={createForm.has_compensating_controls}
                        onChange={e => setCreateForm({ ...createForm, has_compensating_controls: e.target.checked })}
                        className="rounded border-gray-300" />
                      Has Compensating Controls
                    </label>
                  </div>
                  {createForm.has_compensating_controls && (
                    <>
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Description of Compensating Controls</label>
                        <textarea value={createForm.compensating_controls_description}
                          onChange={e => setCreateForm({ ...createForm, compensating_controls_description: e.target.value })}
                          className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm" rows={3}
                          placeholder="Describe the compensating controls in place" />
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Effectiveness Assessment</label>
                        <select value={createForm.compensating_effectiveness}
                          onChange={e => setCreateForm({ ...createForm, compensating_effectiveness: e.target.value })}
                          className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm">
                          <option value="not_assessed">Not Assessed</option>
                          <option value="fully_effective">Fully Effective</option>
                          <option value="mostly_effective">Mostly Effective</option>
                          <option value="partially_effective">Partially Effective</option>
                          <option value="minimally_effective">Minimally Effective</option>
                          <option value="not_effective">Not Effective</option>
                        </select>
                      </div>
                    </>
                  )}

                  <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 text-sm text-blue-800">
                    <p className="font-medium mb-1">Summary</p>
                    <ul className="list-disc list-inside space-y-1 text-blue-700">
                      <li>Type: {createForm.exception_type} | Priority: {createForm.priority}</li>
                      <li>Risk Level: {createForm.residual_risk_level}</li>
                      <li>Effective: {createForm.effective_date || 'Not set'} | Expiry: {createForm.expiry_date || 'None'}</li>
                      <li>Compensating Controls: {createForm.has_compensating_controls ? 'Yes' : 'No'}</li>
                      <li>Review every {createForm.review_frequency_months} month(s)</li>
                    </ul>
                  </div>
                </div>
              )}

              {/* Navigation */}
              <div className="flex items-center justify-between mt-6 pt-4 border-t border-gray-200">
                {createStep > 1 ? (
                  <button onClick={() => setCreateStep(s => s - 1)}
                    className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800">Back</button>
                ) : (
                  <div />
                )}
                {createStep < 3 ? (
                  <button onClick={() => setCreateStep(s => s + 1)}
                    className="bg-indigo-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-indigo-700">
                    Next
                  </button>
                ) : (
                  <button onClick={handleCreate} disabled={creating}
                    className="bg-indigo-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-indigo-700 disabled:opacity-50">
                    {creating ? 'Creating...' : 'Create Exception'}
                  </button>
                )}
              </div>
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

function KPICard({ label, value, color, bg }: { label: string; value: number; color: string; bg: string }) {
  return (
    <div className={`${bg} rounded-lg p-4 border border-gray-100`}>
      <p className="text-xs text-gray-500 mb-1">{label}</p>
      <p className={`text-2xl font-bold ${color}`}>{typeof value === 'number' ? value.toLocaleString() : value}</p>
    </div>
  );
}
