'use client';

import { useEffect, useState, useCallback } from 'react';
import api from '@/lib/api';

// ============================================================
// TYPES
// ============================================================

interface ActivityEntry {
  id: string;
  organization_id: string;
  actor_user_id: string | null;
  actor_email: string;
  actor_name: string;
  action: string;
  entity_type: string;
  entity_id: string;
  entity_ref: string;
  entity_title: string;
  description: string;
  changes: Record<string, unknown>;
  is_system: boolean;
  visibility: string;
  created_at: string;
}

interface UnreadCounts {
  total_unread: number;
  by_entity_type: Record<string, number>;
  unread_entities: UnreadEntityInfo[];
}

interface UnreadEntityInfo {
  entity_type: string;
  entity_id: string;
  unread_count: number;
  last_read_at: string;
}

interface PaginationData {
  page: number;
  page_size: number;
  total_items: number;
  total_pages: number;
  has_next: boolean;
  has_prev: boolean;
}

// ============================================================
// CONSTANTS
// ============================================================

const ACTION_ICONS: Record<string, string> = {
  created: '+ ',
  updated: '~ ',
  deleted: 'x ',
  status_changed: '> ',
  assigned: '@ ',
  commented: '# ',
  mentioned: '@ ',
  approved: 'v ',
  rejected: 'X ',
  submitted: '^ ',
  completed: 'v ',
  published: '! ',
  reviewed: '? ',
  escalated: '! ',
  archived: '- ',
  restored: '+ ',
};

const ACTION_COLORS: Record<string, string> = {
  created: 'bg-green-100 text-green-700 border-green-200',
  updated: 'bg-blue-100 text-blue-700 border-blue-200',
  deleted: 'bg-red-100 text-red-700 border-red-200',
  status_changed: 'bg-amber-100 text-amber-700 border-amber-200',
  assigned: 'bg-purple-100 text-purple-700 border-purple-200',
  commented: 'bg-sky-100 text-sky-700 border-sky-200',
  mentioned: 'bg-indigo-100 text-indigo-700 border-indigo-200',
  approved: 'bg-emerald-100 text-emerald-700 border-emerald-200',
  rejected: 'bg-rose-100 text-rose-700 border-rose-200',
  submitted: 'bg-cyan-100 text-cyan-700 border-cyan-200',
  completed: 'bg-teal-100 text-teal-700 border-teal-200',
  published: 'bg-lime-100 text-lime-700 border-lime-200',
  reviewed: 'bg-orange-100 text-orange-700 border-orange-200',
  escalated: 'bg-red-100 text-red-700 border-red-200',
  archived: 'bg-gray-100 text-gray-600 border-gray-200',
  restored: 'bg-green-100 text-green-700 border-green-200',
};

const ENTITY_TYPE_LABELS: Record<string, string> = {
  risk: 'Risk',
  control: 'Control',
  policy: 'Policy',
  audit: 'Audit',
  finding: 'Finding',
  incident: 'Incident',
  vendor: 'Vendor',
  asset: 'Asset',
  exception: 'Exception',
  dsr: 'DSR Request',
  bia_process: 'BIA Process',
  bc_plan: 'BC Plan',
  questionnaire: 'Questionnaire',
  workflow: 'Workflow',
};

const ACTION_FILTER_OPTIONS = [
  { value: '', label: 'All actions' },
  { value: 'created', label: 'Created' },
  { value: 'updated', label: 'Updated' },
  { value: 'deleted', label: 'Deleted' },
  { value: 'status_changed', label: 'Status changed' },
  { value: 'assigned', label: 'Assigned' },
  { value: 'commented', label: 'Commented' },
  { value: 'approved', label: 'Approved' },
  { value: 'rejected', label: 'Rejected' },
  { value: 'completed', label: 'Completed' },
  { value: 'published', label: 'Published' },
];

const ENTITY_FILTER_OPTIONS = [
  { value: '', label: 'All entities' },
  { value: 'risk', label: 'Risks' },
  { value: 'control', label: 'Controls' },
  { value: 'policy', label: 'Policies' },
  { value: 'audit', label: 'Audits' },
  { value: 'incident', label: 'Incidents' },
  { value: 'vendor', label: 'Vendors' },
  { value: 'exception', label: 'Exceptions' },
];

// ============================================================
// HELPERS
// ============================================================

function formatTimeAgo(dateStr: string): string {
  const date = new Date(dateStr);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMin = Math.floor(diffMs / 60000);
  const diffHour = Math.floor(diffMin / 60);
  const diffDay = Math.floor(diffHour / 24);

  if (diffMin < 1) return 'just now';
  if (diffMin < 60) return `${diffMin}m ago`;
  if (diffHour < 24) return `${diffHour}h ago`;
  if (diffDay < 7) return `${diffDay}d ago`;
  return date.toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric' });
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('en-GB', {
    weekday: 'short',
    day: 'numeric',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

function getActionLabel(action: string): string {
  return action.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase());
}

function groupByDate(entries: ActivityEntry[]): Map<string, ActivityEntry[]> {
  const groups = new Map<string, ActivityEntry[]>();
  for (const entry of entries) {
    const date = new Date(entry.created_at).toLocaleDateString('en-GB', {
      day: 'numeric',
      month: 'long',
      year: 'numeric',
    });
    const existing = groups.get(date) || [];
    existing.push(entry);
    groups.set(date, existing);
  }
  return groups;
}

// ============================================================
// COMPONENT
// ============================================================

export default function ActivityPage() {
  const [entries, setEntries] = useState<ActivityEntry[]>([]);
  const [unread, setUnread] = useState<UnreadCounts | null>(null);
  const [pagination, setPagination] = useState<PaginationData | null>(null);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [viewMode, setViewMode] = useState<'personal' | 'org'>('personal');

  // Filters
  const [actionFilter, setActionFilter] = useState('');
  const [entityTypeFilter, setEntityTypeFilter] = useState('');

  const fetchFeed = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const params = new URLSearchParams({
        page: String(page),
        page_size: '20',
      });
      if (actionFilter) params.set('action', actionFilter);
      if (entityTypeFilter) params.set('entity_type', entityTypeFilter);

      let data;
      if (viewMode === 'personal') {
        data = await api.getActivityFeed(page, 20, {
          action: actionFilter,
          entity_type: entityTypeFilter,
        });
      } else {
        data = await api.getOrgActivityFeed(page, 20, {
          action: actionFilter,
          entity_type: entityTypeFilter,
        });
      }
      setEntries(data.data || []);
      setPagination(data.pagination || null);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to load activity feed');
    } finally {
      setLoading(false);
    }
  }, [page, viewMode, actionFilter, entityTypeFilter]);

  const fetchUnread = useCallback(async () => {
    try {
      const data = await api.getUnreadCounts();
      setUnread(data.data || null);
    } catch {
      // Silently fail for unread counts
    }
  }, []);

  useEffect(() => {
    fetchFeed();
    fetchUnread();
  }, [fetchFeed, fetchUnread]);

  const handleFilterChange = () => {
    setPage(1);
  };

  const dateGroups = groupByDate(entries);

  return (
    <div className="max-w-5xl mx-auto px-4 py-8">
      {/* Header */}
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Activity Feed</h1>
          <p className="text-gray-500 mt-1">
            Track changes, comments, and updates across your organization.
          </p>
        </div>
        {unread && unread.total_unread > 0 && (
          <div className="bg-blue-50 border border-blue-200 rounded-lg px-4 py-2">
            <span className="text-sm font-medium text-blue-700">
              {unread.total_unread} unread
            </span>
          </div>
        )}
      </div>

      {/* View Toggle & Filters */}
      <div className="bg-white border rounded-lg p-4 mb-6">
        <div className="flex flex-wrap items-center gap-4">
          {/* View mode toggle */}
          <div className="flex bg-gray-100 rounded-lg p-0.5">
            <button
              onClick={() => { setViewMode('personal'); setPage(1); }}
              className={`px-4 py-1.5 rounded-md text-sm font-medium transition-colors ${
                viewMode === 'personal'
                  ? 'bg-white text-gray-900 shadow-sm'
                  : 'text-gray-500 hover:text-gray-700'
              }`}
            >
              My Feed
            </button>
            <button
              onClick={() => { setViewMode('org'); setPage(1); }}
              className={`px-4 py-1.5 rounded-md text-sm font-medium transition-colors ${
                viewMode === 'org'
                  ? 'bg-white text-gray-900 shadow-sm'
                  : 'text-gray-500 hover:text-gray-700'
              }`}
            >
              Organization
            </button>
          </div>

          {/* Action filter */}
          <select
            value={actionFilter}
            onChange={(e) => { setActionFilter(e.target.value); handleFilterChange(); }}
            className="border rounded-lg px-3 py-1.5 text-sm bg-white"
          >
            {ACTION_FILTER_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>{opt.label}</option>
            ))}
          </select>

          {/* Entity type filter */}
          <select
            value={entityTypeFilter}
            onChange={(e) => { setEntityTypeFilter(e.target.value); handleFilterChange(); }}
            className="border rounded-lg px-3 py-1.5 text-sm bg-white"
          >
            {ENTITY_FILTER_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>{opt.label}</option>
            ))}
          </select>
        </div>

        {/* Unread by entity type */}
        {unread && Object.keys(unread.by_entity_type).length > 0 && (
          <div className="flex flex-wrap gap-2 mt-3 pt-3 border-t">
            {Object.entries(unread.by_entity_type).map(([type_, count]) => (
              <span
                key={type_}
                className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-50 text-blue-700"
              >
                {ENTITY_TYPE_LABELS[type_] || type_}: {count}
              </span>
            ))}
          </div>
        )}
      </div>

      {/* Error state */}
      {error && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-6">
          <p className="text-red-700 text-sm">{error}</p>
          <button
            onClick={fetchFeed}
            className="mt-2 text-red-600 text-sm underline hover:text-red-800"
          >
            Retry
          </button>
        </div>
      )}

      {/* Loading state */}
      {loading && (
        <div className="flex items-center justify-center py-12">
          <div className="animate-pulse text-gray-400">Loading activity...</div>
        </div>
      )}

      {/* Empty state */}
      {!loading && entries.length === 0 && (
        <div className="text-center py-16 bg-white border rounded-lg">
          <div className="text-4xl mb-4 text-gray-300">&mdash;</div>
          <h3 className="text-lg font-medium text-gray-700">No activity yet</h3>
          <p className="text-gray-500 mt-1 max-w-md mx-auto">
            {viewMode === 'personal'
              ? 'Follow entities to see their activity here. Comments, status changes, and updates will appear in your feed.'
              : 'No organization-wide activity has been recorded yet.'}
          </p>
        </div>
      )}

      {/* Timeline */}
      {!loading && entries.length > 0 && (
        <div className="space-y-8">
          {Array.from(dateGroups.entries()).map(([date, dateEntries]) => (
            <div key={date}>
              {/* Date header */}
              <div className="sticky top-0 z-10 bg-gray-50 -mx-4 px-4 py-2 mb-4">
                <h3 className="text-sm font-semibold text-gray-600">{date}</h3>
              </div>

              {/* Timeline entries */}
              <div className="relative">
                {/* Vertical timeline line */}
                <div className="absolute left-4 top-0 bottom-0 w-px bg-gray-200" />

                <div className="space-y-4">
                  {dateEntries.map((entry) => (
                    <div key={entry.id} className="relative pl-10">
                      {/* Timeline dot */}
                      <div
                        className={`absolute left-2.5 top-2 w-3 h-3 rounded-full border-2 ${
                          ACTION_COLORS[entry.action]
                            ? ACTION_COLORS[entry.action].split(' ')[0]
                            : 'bg-gray-200'
                        } border-white ring-2 ring-gray-100`}
                      />

                      {/* Entry card */}
                      <div className="bg-white border rounded-lg p-4 hover:shadow-sm transition-shadow">
                        <div className="flex items-start justify-between">
                          <div className="flex-1 min-w-0">
                            {/* Action badge + entity info */}
                            <div className="flex items-center gap-2 mb-1">
                              <span
                                className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium border ${
                                  ACTION_COLORS[entry.action] || 'bg-gray-100 text-gray-600 border-gray-200'
                                }`}
                              >
                                {ACTION_ICONS[entry.action] || '  '}
                                {getActionLabel(entry.action)}
                              </span>
                              <span className="text-xs text-gray-400 px-1.5 py-0.5 bg-gray-50 rounded">
                                {ENTITY_TYPE_LABELS[entry.entity_type] || entry.entity_type}
                              </span>
                              {entry.entity_ref && (
                                <span className="text-xs font-mono text-gray-500">
                                  {entry.entity_ref}
                                </span>
                              )}
                              {entry.is_system && (
                                <span className="text-xs text-gray-400 italic">system</span>
                              )}
                            </div>

                            {/* Title */}
                            {entry.entity_title && (
                              <h4 className="text-sm font-medium text-gray-900 truncate">
                                {entry.entity_title}
                              </h4>
                            )}

                            {/* Description */}
                            {entry.description && (
                              <p className="text-sm text-gray-600 mt-0.5 line-clamp-2">
                                {entry.description}
                              </p>
                            )}

                            {/* Changes diff */}
                            {entry.changes && Object.keys(entry.changes).length > 0 && (
                              <div className="mt-2 text-xs bg-gray-50 rounded p-2 border">
                                {Object.entries(entry.changes).map(([field, change]) => (
                                  <div key={field} className="flex items-center gap-1 text-gray-500">
                                    <span className="font-medium">{field}:</span>
                                    <span className="text-red-500 line-through">
                                      {String((change as Record<string, unknown>)?.old || '')}
                                    </span>
                                    <span className="text-gray-400">{'->'}</span>
                                    <span className="text-green-600">
                                      {String((change as Record<string, unknown>)?.new || '')}
                                    </span>
                                  </div>
                                ))}
                              </div>
                            )}
                          </div>

                          {/* Timestamp + actor */}
                          <div className="text-right ml-4 flex-shrink-0">
                            <div className="text-xs text-gray-400" title={formatDate(entry.created_at)}>
                              {formatTimeAgo(entry.created_at)}
                            </div>
                            {entry.actor_name && (
                              <div className="text-xs text-gray-500 mt-0.5">
                                {entry.actor_name}
                              </div>
                            )}
                            {!entry.actor_name && entry.actor_email && (
                              <div className="text-xs text-gray-500 mt-0.5">
                                {entry.actor_email}
                              </div>
                            )}
                          </div>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Pagination */}
      {pagination && pagination.total_pages > 1 && (
        <div className="flex items-center justify-between mt-8 pt-6 border-t">
          <p className="text-sm text-gray-500">
            Page {pagination.page} of {pagination.total_pages} ({pagination.total_items} total)
          </p>
          <div className="flex gap-2">
            <button
              onClick={() => setPage(p => Math.max(1, p - 1))}
              disabled={!pagination.has_prev}
              className="px-3 py-1.5 text-sm border rounded-lg disabled:opacity-40 hover:bg-gray-50 transition-colors"
            >
              Previous
            </button>
            <button
              onClick={() => setPage(p => p + 1)}
              disabled={!pagination.has_next}
              className="px-3 py-1.5 text-sm border rounded-lg disabled:opacity-40 hover:bg-gray-50 transition-colors"
            >
              Next
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
