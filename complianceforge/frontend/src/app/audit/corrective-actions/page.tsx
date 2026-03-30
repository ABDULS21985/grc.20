'use client';

import { useEffect, useState, useCallback } from 'react';
import api from '@/lib/api';

// ============================================================
// TYPES
// ============================================================

interface CorrectiveAction {
  id: string;
  finding_id: string | null;
  engagement_id: string | null;
  action_ref: string;
  title: string;
  description: string;
  action_type: string;
  priority: string;
  status: string;
  root_cause: string;
  planned_action: string;
  actual_action: string;
  responsible_user_id: string | null;
  implementer_user_id: string | null;
  due_date: string | null;
  completed_date: string | null;
  verified_by: string | null;
  verified_date: string | null;
  verification_notes: string;
  verification_status: string;
  cost_estimate: number | null;
  actual_cost: number | null;
  effectiveness_rating: string;
  follow_up_required: boolean;
  follow_up_date: string | null;
  created_at: string;
}

interface DashboardMetrics {
  total: number;
  open: number;
  inProgress: number;
  completed: number;
  verified: number;
  closed: number;
  overdue: number;
  byPriority: Record<string, number>;
  completionRate: number;
}

// ============================================================
// COLOUR HELPERS
// ============================================================

const STATUS_COLORS: Record<string, string> = {
  open: 'bg-blue-100 text-blue-800',
  in_progress: 'bg-yellow-100 text-yellow-800',
  completed: 'bg-emerald-100 text-emerald-800',
  verified: 'bg-green-100 text-green-800',
  closed: 'bg-gray-100 text-gray-700',
  overdue: 'bg-red-100 text-red-800',
  rejected: 'bg-red-100 text-red-700',
};

const PRIORITY_COLORS: Record<string, string> = {
  critical: 'bg-red-100 text-red-800',
  high: 'bg-orange-100 text-orange-800',
  medium: 'bg-yellow-100 text-yellow-800',
  low: 'bg-blue-100 text-blue-800',
};

const PRIORITY_DOT_COLORS: Record<string, string> = {
  critical: 'bg-red-500',
  high: 'bg-orange-500',
  medium: 'bg-yellow-500',
  low: 'bg-blue-400',
};

// ============================================================
// MAIN PAGE COMPONENT
// ============================================================

export default function CorrectiveActionsPage() {
  const [actions, setActions] = useState<CorrectiveAction[]>([]);
  const [loading, setLoading] = useState(true);
  const [filterStatus, setFilterStatus] = useState('');
  const [filterPriority, setFilterPriority] = useState('');
  const [showOverdueOnly, setShowOverdueOnly] = useState(false);
  const [page, setPage] = useState(1);
  const [totalItems, setTotalItems] = useState(0);
  const pageSize = 20;

  // --------------------------------------------------------
  // DATA FETCHING
  // --------------------------------------------------------

  const fetchActions = useCallback(async () => {
    try {
      setLoading(true);
      let url = `/audit/corrective-actions?page=${page}&page_size=${pageSize}`;
      if (filterStatus) url += `&status=${filterStatus}`;
      if (filterPriority) url += `&priority=${filterPriority}`;

      const res = await api.get<{
        data: CorrectiveAction[];
        pagination: { total_items: number };
      }>(url);

      let data = res.data || [];

      // Client-side overdue filter
      if (showOverdueOnly) {
        const now = new Date();
        data = data.filter((a) => {
          if (!a.due_date) return false;
          return (
            new Date(a.due_date) < now &&
            !['closed', 'verified', 'completed'].includes(a.status)
          );
        });
      }

      setActions(data);
      setTotalItems(res.pagination?.total_items || data.length);
    } catch {
      console.error('Failed to fetch corrective actions');
    } finally {
      setLoading(false);
    }
  }, [page, filterStatus, filterPriority, showOverdueOnly]);

  useEffect(() => {
    fetchActions();
  }, [fetchActions]);

  // --------------------------------------------------------
  // DASHBOARD METRICS
  // --------------------------------------------------------

  const computeMetrics = (items: CorrectiveAction[]): DashboardMetrics => {
    const now = new Date();
    let open = 0,
      inProgress = 0,
      completed = 0,
      verified = 0,
      closed = 0,
      overdue = 0;
    const byPriority: Record<string, number> = {};

    for (const a of items) {
      switch (a.status) {
        case 'open':
          open++;
          break;
        case 'in_progress':
          inProgress++;
          break;
        case 'completed':
          completed++;
          break;
        case 'verified':
          verified++;
          break;
        case 'closed':
          closed++;
          break;
      }

      // Overdue check
      if (
        a.due_date &&
        new Date(a.due_date) < now &&
        !['closed', 'verified', 'completed'].includes(a.status)
      ) {
        overdue++;
      }

      byPriority[a.priority] = (byPriority[a.priority] || 0) + 1;
    }

    const total = items.length;
    const resolvedCount = completed + verified + closed;
    const completionRate = total > 0 ? Math.round((resolvedCount / total) * 100) : 0;

    return {
      total,
      open,
      inProgress,
      completed,
      verified,
      closed,
      overdue,
      byPriority,
      completionRate,
    };
  };

  const metrics = computeMetrics(actions);

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

  const isOverdue = (a: CorrectiveAction) => {
    if (!a.due_date) return false;
    return (
      new Date(a.due_date) < new Date() &&
      !['closed', 'verified', 'completed'].includes(a.status)
    );
  };

  const totalPages = Math.ceil(totalItems / pageSize);

  // --------------------------------------------------------
  // RENDER
  // --------------------------------------------------------

  return (
    <div className="p-6 max-w-7xl mx-auto">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900">Corrective Actions Tracker</h1>
        <p className="text-gray-500 text-sm mt-1">
          Organisation-wide view of corrective actions from all audit engagements.
        </p>
      </div>

      {/* Dashboard Cards */}
      <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-7 gap-4 mb-6">
        <div className="bg-white border rounded-lg p-4">
          <div className="text-2xl font-bold text-gray-900">{metrics.total}</div>
          <div className="text-xs text-gray-500 mt-1">Total Actions</div>
        </div>
        <div className="bg-white border rounded-lg p-4">
          <div className="text-2xl font-bold text-blue-600">{metrics.open}</div>
          <div className="text-xs text-gray-500 mt-1">Open</div>
        </div>
        <div className="bg-white border rounded-lg p-4">
          <div className="text-2xl font-bold text-yellow-600">{metrics.inProgress}</div>
          <div className="text-xs text-gray-500 mt-1">In Progress</div>
        </div>
        <div className="bg-white border rounded-lg p-4">
          <div className="text-2xl font-bold text-emerald-600">{metrics.completed}</div>
          <div className="text-xs text-gray-500 mt-1">Completed</div>
        </div>
        <div className="bg-white border rounded-lg p-4">
          <div className="text-2xl font-bold text-red-600">{metrics.overdue}</div>
          <div className="text-xs text-gray-500 mt-1">Overdue</div>
        </div>
        <div className="bg-white border rounded-lg p-4">
          <div className="text-2xl font-bold text-gray-600">{metrics.closed}</div>
          <div className="text-xs text-gray-500 mt-1">Closed</div>
        </div>
        <div className="bg-white border rounded-lg p-4">
          <div className="text-2xl font-bold text-green-700">{metrics.completionRate}%</div>
          <div className="text-xs text-gray-500 mt-1">Completion Rate</div>
        </div>
      </div>

      {/* Priority Breakdown */}
      <div className="bg-white border rounded-lg p-4 mb-6">
        <h3 className="text-sm font-semibold text-gray-700 mb-3">By Priority</h3>
        <div className="flex gap-6">
          {['critical', 'high', 'medium', 'low'].map((p) => (
            <div key={p} className="flex items-center gap-2">
              <div className={`w-3 h-3 rounded-full ${PRIORITY_DOT_COLORS[p] || 'bg-gray-400'}`} />
              <span className="text-sm text-gray-700 capitalize">{p}</span>
              <span className="text-sm font-semibold text-gray-900">{metrics.byPriority[p] || 0}</span>
            </div>
          ))}
        </div>
      </div>

      {/* Filters */}
      <div className="flex flex-wrap items-center gap-4 mb-4">
        <div>
          <label className="block text-xs text-gray-500 mb-1">Status</label>
          <select
            value={filterStatus}
            onChange={(e) => {
              setFilterStatus(e.target.value);
              setPage(1);
            }}
            className="px-3 py-1.5 border rounded text-sm"
          >
            <option value="">All</option>
            <option value="open">Open</option>
            <option value="in_progress">In Progress</option>
            <option value="completed">Completed</option>
            <option value="verified">Verified</option>
            <option value="closed">Closed</option>
          </select>
        </div>
        <div>
          <label className="block text-xs text-gray-500 mb-1">Priority</label>
          <select
            value={filterPriority}
            onChange={(e) => {
              setFilterPriority(e.target.value);
              setPage(1);
            }}
            className="px-3 py-1.5 border rounded text-sm"
          >
            <option value="">All</option>
            <option value="critical">Critical</option>
            <option value="high">High</option>
            <option value="medium">Medium</option>
            <option value="low">Low</option>
          </select>
        </div>
        <div className="flex items-end">
          <label className="flex items-center gap-2 text-sm text-gray-700 cursor-pointer">
            <input
              type="checkbox"
              checked={showOverdueOnly}
              onChange={(e) => {
                setShowOverdueOnly(e.target.checked);
                setPage(1);
              }}
              className="rounded border-gray-300"
            />
            Overdue only
          </label>
        </div>
      </div>

      {/* Actions Table */}
      {loading ? (
        <div className="text-center py-12 text-gray-500">Loading corrective actions...</div>
      ) : actions.length === 0 ? (
        <div className="bg-white border rounded-lg p-12 text-center">
          <h3 className="text-lg font-medium text-gray-900 mb-2">No Corrective Actions</h3>
          <p className="text-gray-500 text-sm">
            {filterStatus || filterPriority || showOverdueOnly
              ? 'No actions match the current filters.'
              : 'No corrective actions have been created yet.'}
          </p>
        </div>
      ) : (
        <div className="bg-white border rounded-lg overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-gray-50 text-left">
              <tr>
                <th className="px-4 py-3 font-medium text-gray-600">Ref</th>
                <th className="px-4 py-3 font-medium text-gray-600">Title</th>
                <th className="px-4 py-3 font-medium text-gray-600">Type</th>
                <th className="px-4 py-3 font-medium text-gray-600">Priority</th>
                <th className="px-4 py-3 font-medium text-gray-600">Status</th>
                <th className="px-4 py-3 font-medium text-gray-600">Due Date</th>
                <th className="px-4 py-3 font-medium text-gray-600">Completed</th>
                <th className="px-4 py-3 font-medium text-gray-600">Verified</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {actions.map((a) => (
                <tr
                  key={a.id}
                  className={`hover:bg-gray-50 ${isOverdue(a) ? 'bg-red-50' : ''}`}
                >
                  <td className="px-4 py-3 font-mono text-gray-500">{a.action_ref}</td>
                  <td className="px-4 py-3">
                    <div className="text-gray-900 font-medium">{a.title}</div>
                    {a.description && (
                      <div className="text-gray-500 text-xs mt-0.5 line-clamp-1">
                        {a.description}
                      </div>
                    )}
                  </td>
                  <td className="px-4 py-3 text-gray-600 capitalize">{a.action_type}</td>
                  <td className="px-4 py-3">
                    <span
                      className={`px-2 py-0.5 rounded text-xs font-medium ${PRIORITY_COLORS[a.priority] || 'bg-gray-100 text-gray-700'}`}
                    >
                      {a.priority}
                    </span>
                  </td>
                  <td className="px-4 py-3">
                    <span
                      className={`px-2 py-0.5 rounded text-xs font-medium ${
                        isOverdue(a)
                          ? STATUS_COLORS.overdue
                          : STATUS_COLORS[a.status] || 'bg-gray-100 text-gray-700'
                      }`}
                    >
                      {isOverdue(a) ? 'overdue' : a.status}
                    </span>
                  </td>
                  <td className="px-4 py-3">
                    <span className={isOverdue(a) ? 'text-red-700 font-medium' : 'text-gray-600'}>
                      {formatDate(a.due_date)}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-gray-600">{formatDate(a.completed_date)}</td>
                  <td className="px-4 py-3">
                    {a.verified_date ? (
                      <span className="text-green-700 text-xs font-medium">
                        {formatDate(a.verified_date)}
                      </span>
                    ) : (
                      <span className="text-gray-400">-</span>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>

          {/* Pagination */}
          {totalPages > 1 && (
            <div className="flex items-center justify-between px-4 py-3 border-t bg-gray-50">
              <span className="text-sm text-gray-500">
                Showing {(page - 1) * pageSize + 1}--
                {Math.min(page * pageSize, totalItems)} of {totalItems}
              </span>
              <div className="flex gap-2">
                <button
                  onClick={() => setPage((p) => Math.max(1, p - 1))}
                  disabled={page <= 1}
                  className="px-3 py-1 text-sm border rounded hover:bg-gray-100 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  Previous
                </button>
                <button
                  onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                  disabled={page >= totalPages}
                  className="px-3 py-1 text-sm border rounded hover:bg-gray-100 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  Next
                </button>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
