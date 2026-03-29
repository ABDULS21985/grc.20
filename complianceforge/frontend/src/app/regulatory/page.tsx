'use client';

import { useEffect, useState } from 'react';
import api from '@/lib/api';

// ============================================================
// TYPES
// ============================================================

interface RegulatoryChange {
  id: string;
  change_ref: string;
  title: string;
  summary: string;
  full_text_url: string;
  published_date: string | null;
  effective_date: string | null;
  change_type: string;
  severity: string;
  status: string;
  affected_frameworks: string[];
  affected_regions: string[];
  tags: string[];
  source_name: string;
  deadline: string | null;
}

interface RegulatorySource {
  id: string;
  name: string;
  source_type: string;
  country_code: string;
  region: string;
  url: string;
  rss_feed_url: string;
  scan_frequency: string;
  last_scanned_at: string | null;
  is_active: boolean;
  relevance_frameworks: string[];
}

interface RegSubscription {
  id: string;
  source_id: string;
  source_name: string;
  source_type: string;
  is_active: boolean;
  notification_on_new: boolean;
  auto_assess: boolean;
}

interface RegDashboard {
  new_changes_count: number;
  pending_assessments: number;
  action_required: number;
  subscribed_sources_count: number;
  severity_breakdown: Record<string, number>;
  status_breakdown: Record<string, number>;
  upcoming_deadlines: DeadlineEntry[];
  recent_changes: RegulatoryChange[];
}

interface DeadlineEntry {
  change_id: string;
  change_ref: string;
  title: string;
  deadline: string;
  severity: string;
  days_left: number;
}

interface TimelineEntry {
  date: string;
  change_ref: string;
  title: string;
  change_type: string;
  severity: string;
  status: string;
  event_type: string;
}

// ============================================================
// CONSTANTS
// ============================================================

const SEVERITY_COLORS: Record<string, string> = {
  critical: 'bg-red-100 text-red-700',
  high: 'bg-orange-100 text-orange-700',
  medium: 'bg-amber-100 text-amber-700',
  low: 'bg-blue-100 text-blue-700',
  informational: 'bg-gray-100 text-gray-600',
};

const STATUS_COLORS: Record<string, string> = {
  new: 'bg-indigo-100 text-indigo-700',
  under_assessment: 'bg-yellow-100 text-yellow-700',
  assessed: 'bg-blue-100 text-blue-700',
  action_required: 'bg-red-100 text-red-700',
  implemented: 'bg-green-100 text-green-700',
  not_applicable: 'bg-gray-100 text-gray-500',
  monitoring: 'bg-purple-100 text-purple-700',
};

const CHANGE_TYPE_LABELS: Record<string, string> = {
  new_regulation: 'New Regulation',
  amendment: 'Amendment',
  guidance: 'Guidance',
  enforcement_decision: 'Enforcement',
  standard_revision: 'Std Revision',
  standard_update: 'Std Update',
  industry_bulletin: 'Bulletin',
  court_ruling: 'Court Ruling',
  consultation: 'Consultation',
};

const EVENT_TYPE_COLORS: Record<string, string> = {
  published: 'bg-blue-500',
  effective: 'bg-green-500',
  deadline: 'bg-red-500',
};

const SOURCE_TYPE_LABELS: Record<string, string> = {
  supervisory_authority: 'Supervisory Authority',
  standards_body: 'Standards Body',
  government: 'Government',
  industry_body: 'Industry Body',
  legal_publisher: 'Legal Publisher',
  custom: 'Custom',
};

type Tab = 'dashboard' | 'changes' | 'sources' | 'subscriptions' | 'timeline';

// ============================================================
// COMPONENT
// ============================================================

export default function RegulatoryPage() {
  const [tab, setTab] = useState<Tab>('dashboard');
  const [dashboard, setDashboard] = useState<RegDashboard | null>(null);
  const [changes, setChanges] = useState<RegulatoryChange[]>([]);
  const [sources, setSources] = useState<RegulatorySource[]>([]);
  const [subscriptions, setSubscriptions] = useState<RegSubscription[]>([]);
  const [timeline, setTimeline] = useState<TimelineEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [totalChanges, setTotalChanges] = useState(0);
  const [page, setPage] = useState(1);

  // Filters
  const [severityFilter, setSeverityFilter] = useState('');
  const [statusFilter, setStatusFilter] = useState('');
  const [regionFilter, setRegionFilter] = useState('');
  const [searchFilter, setSearchFilter] = useState('');

  useEffect(() => {
    loadData();
  }, [tab, page, severityFilter, statusFilter, regionFilter]);

  async function loadData() {
    setLoading(true);
    try {
      if (tab === 'dashboard') {
        const res = await api.getRegulatoryDashboard();
        setDashboard(res.data);
      } else if (tab === 'changes') {
        const params = new URLSearchParams({ page: String(page), page_size: '20' });
        if (severityFilter) params.set('severity', severityFilter);
        if (statusFilter) params.set('status', statusFilter);
        if (regionFilter) params.set('region', regionFilter);
        if (searchFilter) params.set('search', searchFilter);
        const res = await api.getRegulatoryChanges(params.toString());
        setChanges(res.data || []);
        setTotalChanges(res.pagination?.total_items || 0);
      } else if (tab === 'sources') {
        const res = await api.getRegulatorySources();
        setSources(res.data || []);
      } else if (tab === 'subscriptions') {
        const res = await api.getRegulatorySubscriptions();
        setSubscriptions(res.data || []);
      } else if (tab === 'timeline') {
        const res = await api.getRegulatoryTimeline(12);
        setTimeline(res.data || []);
      }
    } catch { /* swallow */ }
    setLoading(false);
  }

  async function handleSubscribe(sourceId: string) {
    try {
      await api.subscribeRegulatory({ source_id: sourceId, notification_on_new: true });
      loadData();
    } catch { /* swallow */ }
  }

  async function handleUnsubscribe(subId: string) {
    try {
      await api.unsubscribeRegulatory(subId);
      loadData();
    } catch { /* swallow */ }
  }

  function handleSearch() {
    setPage(1);
    loadData();
  }

  const tabs: { key: Tab; label: string }[] = [
    { key: 'dashboard', label: 'Dashboard' },
    { key: 'changes', label: 'Change Feed' },
    { key: 'sources', label: 'Sources' },
    { key: 'subscriptions', label: 'Subscriptions' },
    { key: 'timeline', label: 'Timeline' },
  ];

  return (
    <div>
      {/* Header */}
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Regulatory Change Management</h1>
          <p className="text-gray-500 mt-1">
            Horizon scanning, impact assessment, and regulatory change tracking
          </p>
        </div>
      </div>

      {/* Tab Bar */}
      <div className="flex gap-1 border-b border-gray-200 mb-6">
        {tabs.map(t => (
          <button
            key={t.key}
            onClick={() => { setTab(t.key); setPage(1); }}
            className={`px-4 py-2.5 text-sm font-medium border-b-2 transition-colors ${
              tab === t.key
                ? 'border-indigo-600 text-indigo-600'
                : 'border-transparent text-gray-500 hover:text-gray-700'
            }`}
          >
            {t.label}
            {t.key === 'changes' && dashboard?.new_changes_count ? (
              <span className="ml-1.5 inline-flex h-5 w-5 items-center justify-center rounded-full bg-red-500 text-[10px] text-white font-bold">
                {dashboard.new_changes_count}
              </span>
            ) : null}
          </button>
        ))}
      </div>

      {/* Content */}
      {loading ? (
        <div className="card animate-pulse h-96" />
      ) : tab === 'dashboard' && dashboard ? (
        /* ============================================================
           DASHBOARD TAB
           ============================================================ */
        <div>
          {/* Summary Cards */}
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-4 mb-6">
            <div className="card">
              <p className="text-sm text-gray-500">New Changes</p>
              <p className="text-3xl font-bold text-indigo-600 mt-1">{dashboard.new_changes_count}</p>
              <p className="text-xs text-gray-400 mt-1">Awaiting review</p>
            </div>
            <div className="card">
              <p className="text-sm text-gray-500">Pending Assessments</p>
              <p className="text-3xl font-bold text-amber-600 mt-1">{dashboard.pending_assessments}</p>
              <p className="text-xs text-gray-400 mt-1">Impact assessment needed</p>
            </div>
            <div className="card">
              <p className="text-sm text-gray-500">Action Required</p>
              <p className="text-3xl font-bold text-red-600 mt-1">{dashboard.action_required}</p>
              <p className="text-xs text-gray-400 mt-1">Remediation needed</p>
            </div>
            <div className="card">
              <p className="text-sm text-gray-500">Subscribed Sources</p>
              <p className="text-3xl font-bold text-gray-900 mt-1">{dashboard.subscribed_sources_count}</p>
              <p className="text-xs text-gray-400 mt-1">Active monitoring</p>
            </div>
          </div>

          {/* Severity Breakdown */}
          <div className="grid grid-cols-1 gap-4 lg:grid-cols-2 mb-6">
            <div className="card">
              <h3 className="font-semibold text-gray-900 mb-4">Severity Breakdown</h3>
              <div className="space-y-2">
                {Object.entries(dashboard.severity_breakdown || {}).map(([sev, count]) => (
                  <div key={sev} className="flex items-center justify-between">
                    <span className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${SEVERITY_COLORS[sev] || 'bg-gray-100 text-gray-600'}`}>
                      {sev}
                    </span>
                    <span className="text-sm font-medium text-gray-900">{count}</span>
                  </div>
                ))}
                {Object.keys(dashboard.severity_breakdown || {}).length === 0 && (
                  <p className="text-sm text-gray-400">No changes tracked yet</p>
                )}
              </div>
            </div>

            <div className="card">
              <h3 className="font-semibold text-gray-900 mb-4">Status Breakdown</h3>
              <div className="space-y-2">
                {Object.entries(dashboard.status_breakdown || {}).map(([status, count]) => (
                  <div key={status} className="flex items-center justify-between">
                    <span className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${STATUS_COLORS[status] || 'bg-gray-100 text-gray-600'}`}>
                      {status.replace(/_/g, ' ')}
                    </span>
                    <span className="text-sm font-medium text-gray-900">{count}</span>
                  </div>
                ))}
                {Object.keys(dashboard.status_breakdown || {}).length === 0 && (
                  <p className="text-sm text-gray-400">No changes tracked yet</p>
                )}
              </div>
            </div>
          </div>

          {/* Upcoming Deadlines */}
          {dashboard.upcoming_deadlines.length > 0 && (
            <div className="card mb-6">
              <h3 className="font-semibold text-gray-900 mb-4">Upcoming Deadlines</h3>
              <div className="space-y-3">
                {dashboard.upcoming_deadlines.map(d => (
                  <div key={d.change_id} className="flex items-center justify-between py-2 border-b border-gray-50 last:border-0">
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium text-gray-900 truncate">{d.title}</p>
                      <p className="text-xs text-gray-500">{d.change_ref}</p>
                    </div>
                    <div className="flex items-center gap-3 ml-4">
                      <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${SEVERITY_COLORS[d.severity]}`}>
                        {d.severity}
                      </span>
                      <span className={`text-sm font-medium ${d.days_left <= 7 ? 'text-red-600' : d.days_left <= 30 ? 'text-amber-600' : 'text-gray-600'}`}>
                        {d.days_left}d left
                      </span>
                      <span className="text-xs text-gray-400">
                        {new Date(d.deadline).toLocaleDateString('en-GB')}
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Recent Changes */}
          <div className="card">
            <h3 className="font-semibold text-gray-900 mb-4">Recent Changes</h3>
            <div className="space-y-3">
              {dashboard.recent_changes.map(c => (
                <div key={c.id} className="flex items-start gap-3 py-2 border-b border-gray-50 last:border-0">
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium text-gray-900 truncate">{c.title}</p>
                    <div className="flex items-center gap-2 mt-1">
                      <span className="text-xs text-gray-400">{c.change_ref}</span>
                      {c.source_name && (
                        <span className="text-xs text-gray-400">from {c.source_name}</span>
                      )}
                      {c.published_date && (
                        <span className="text-xs text-gray-400">
                          {new Date(c.published_date).toLocaleDateString('en-GB')}
                        </span>
                      )}
                    </div>
                  </div>
                  <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${SEVERITY_COLORS[c.severity]}`}>
                    {c.severity}
                  </span>
                  <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${STATUS_COLORS[c.status]}`}>
                    {c.status.replace(/_/g, ' ')}
                  </span>
                </div>
              ))}
              {dashboard.recent_changes.length === 0 && (
                <p className="text-sm text-gray-400 py-4 text-center">No regulatory changes detected yet</p>
              )}
            </div>
          </div>
        </div>

      ) : tab === 'changes' ? (
        /* ============================================================
           CHANGES TAB
           ============================================================ */
        <div>
          {/* Filters */}
          <div className="flex flex-wrap gap-3 mb-4">
            <input
              type="text"
              placeholder="Search changes..."
              value={searchFilter}
              onChange={e => setSearchFilter(e.target.value)}
              onKeyDown={e => e.key === 'Enter' && handleSearch()}
              className="rounded-lg border border-gray-300 px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
            />
            <select
              value={severityFilter}
              onChange={e => { setSeverityFilter(e.target.value); setPage(1); }}
              className="rounded-lg border border-gray-300 px-3 py-1.5 text-sm"
            >
              <option value="">All Severities</option>
              <option value="critical">Critical</option>
              <option value="high">High</option>
              <option value="medium">Medium</option>
              <option value="low">Low</option>
              <option value="informational">Informational</option>
            </select>
            <select
              value={statusFilter}
              onChange={e => { setStatusFilter(e.target.value); setPage(1); }}
              className="rounded-lg border border-gray-300 px-3 py-1.5 text-sm"
            >
              <option value="">All Statuses</option>
              <option value="new">New</option>
              <option value="under_assessment">Under Assessment</option>
              <option value="assessed">Assessed</option>
              <option value="action_required">Action Required</option>
              <option value="implemented">Implemented</option>
              <option value="not_applicable">Not Applicable</option>
              <option value="monitoring">Monitoring</option>
            </select>
            <select
              value={regionFilter}
              onChange={e => { setRegionFilter(e.target.value); setPage(1); }}
              className="rounded-lg border border-gray-300 px-3 py-1.5 text-sm"
            >
              <option value="">All Regions</option>
              <option value="GB">United Kingdom</option>
              <option value="EU">European Union</option>
              <option value="DE">Germany</option>
              <option value="FR">France</option>
              <option value="US">United States</option>
              <option value="GLOBAL">Global</option>
            </select>
            <button
              onClick={handleSearch}
              className="rounded-lg bg-indigo-600 px-4 py-1.5 text-sm text-white hover:bg-indigo-700"
            >
              Search
            </button>
          </div>

          {/* Changes Table */}
          <div className="card overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="table-header">
                  <th className="px-4 py-3 text-left">Ref</th>
                  <th className="px-4 py-3 text-left">Title</th>
                  <th className="px-4 py-3 text-left">Source</th>
                  <th className="px-4 py-3 text-left">Type</th>
                  <th className="px-4 py-3 text-left">Severity</th>
                  <th className="px-4 py-3 text-left">Status</th>
                  <th className="px-4 py-3 text-left">Published</th>
                  <th className="px-4 py-3 text-left">Frameworks</th>
                </tr>
              </thead>
              <tbody>
                {changes.map(c => (
                  <tr key={c.id} className="border-t border-gray-100 hover:bg-gray-50">
                    <td className="px-4 py-3 font-mono text-xs text-gray-500 whitespace-nowrap">{c.change_ref}</td>
                    <td className="px-4 py-3">
                      <div className="max-w-sm">
                        <p className="font-medium text-gray-900 truncate">{c.title}</p>
                        {c.full_text_url && (
                          <a
                            href={c.full_text_url}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="text-xs text-indigo-600 hover:text-indigo-700"
                          >
                            View source
                          </a>
                        )}
                      </div>
                    </td>
                    <td className="px-4 py-3 text-gray-600 text-xs">{c.source_name || '-'}</td>
                    <td className="px-4 py-3">
                      <span className="inline-flex items-center rounded-full bg-gray-100 px-2 py-0.5 text-xs font-medium text-gray-600">
                        {CHANGE_TYPE_LABELS[c.change_type] || c.change_type || '-'}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${SEVERITY_COLORS[c.severity]}`}>
                        {c.severity}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${STATUS_COLORS[c.status]}`}>
                        {c.status.replace(/_/g, ' ')}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-gray-600 text-xs whitespace-nowrap">
                      {c.published_date ? new Date(c.published_date).toLocaleDateString('en-GB') : '-'}
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex flex-wrap gap-1">
                        {(c.affected_frameworks || []).slice(0, 3).map(f => (
                          <span key={f} className="inline-flex items-center rounded bg-indigo-50 px-1.5 py-0.5 text-[10px] font-medium text-indigo-600">
                            {f}
                          </span>
                        ))}
                        {(c.affected_frameworks || []).length > 3 && (
                          <span className="text-[10px] text-gray-400">+{c.affected_frameworks.length - 3}</span>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
                {changes.length === 0 && (
                  <tr>
                    <td colSpan={8} className="px-4 py-12 text-center text-gray-500">
                      No regulatory changes found matching your filters
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>

          {/* Pagination */}
          {totalChanges > 20 && (
            <div className="flex items-center justify-between mt-4">
              <p className="text-sm text-gray-500">
                Showing {(page - 1) * 20 + 1} to {Math.min(page * 20, totalChanges)} of {totalChanges}
              </p>
              <div className="flex gap-2">
                <button
                  onClick={() => setPage(p => Math.max(1, p - 1))}
                  disabled={page === 1}
                  className="rounded-lg border border-gray-300 px-3 py-1 text-sm disabled:opacity-50"
                >
                  Previous
                </button>
                <button
                  onClick={() => setPage(p => p + 1)}
                  disabled={page * 20 >= totalChanges}
                  className="rounded-lg border border-gray-300 px-3 py-1 text-sm disabled:opacity-50"
                >
                  Next
                </button>
              </div>
            </div>
          )}
        </div>

      ) : tab === 'sources' ? (
        /* ============================================================
           SOURCES TAB
           ============================================================ */
        <div>
          <div className="card overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="table-header">
                  <th className="px-4 py-3 text-left">Name</th>
                  <th className="px-4 py-3 text-left">Type</th>
                  <th className="px-4 py-3 text-left">Country</th>
                  <th className="px-4 py-3 text-left">Region</th>
                  <th className="px-4 py-3 text-left">Frequency</th>
                  <th className="px-4 py-3 text-left">Last Scanned</th>
                  <th className="px-4 py-3 text-left">Frameworks</th>
                  <th className="px-4 py-3 text-left">Status</th>
                  <th className="px-4 py-3 text-left">Actions</th>
                </tr>
              </thead>
              <tbody>
                {sources.map(src => (
                  <tr key={src.id} className="border-t border-gray-100 hover:bg-gray-50">
                    <td className="px-4 py-3">
                      <div>
                        <p className="font-medium text-gray-900">{src.name}</p>
                        {src.url && (
                          <a
                            href={src.url}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="text-xs text-indigo-600 hover:text-indigo-700"
                          >
                            Visit website
                          </a>
                        )}
                      </div>
                    </td>
                    <td className="px-4 py-3 text-xs text-gray-600">
                      {SOURCE_TYPE_LABELS[src.source_type] || src.source_type}
                    </td>
                    <td className="px-4 py-3 text-gray-600">{src.country_code || '-'}</td>
                    <td className="px-4 py-3 text-gray-600">{src.region || '-'}</td>
                    <td className="px-4 py-3">
                      <span className="inline-flex items-center rounded-full bg-gray-100 px-2 py-0.5 text-xs font-medium text-gray-600">
                        {src.scan_frequency}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-xs text-gray-500">
                      {src.last_scanned_at
                        ? new Date(src.last_scanned_at).toLocaleDateString('en-GB', {
                            day: '2-digit', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit',
                          })
                        : 'Never'}
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex flex-wrap gap-1">
                        {(src.relevance_frameworks || []).slice(0, 3).map(f => (
                          <span key={f} className="inline-flex items-center rounded bg-indigo-50 px-1.5 py-0.5 text-[10px] font-medium text-indigo-600">
                            {f}
                          </span>
                        ))}
                        {(src.relevance_frameworks || []).length > 3 && (
                          <span className="text-[10px] text-gray-400">+{src.relevance_frameworks.length - 3}</span>
                        )}
                      </div>
                    </td>
                    <td className="px-4 py-3">
                      <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${
                        src.is_active ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-500'
                      }`}>
                        {src.is_active ? 'Active' : 'Inactive'}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <button
                        onClick={() => handleSubscribe(src.id)}
                        className="text-indigo-600 hover:text-indigo-700 text-xs font-medium"
                      >
                        Subscribe
                      </button>
                    </td>
                  </tr>
                ))}
                {sources.length === 0 && (
                  <tr>
                    <td colSpan={9} className="px-4 py-12 text-center text-gray-500">
                      No regulatory sources configured
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </div>

      ) : tab === 'subscriptions' ? (
        /* ============================================================
           SUBSCRIPTIONS TAB
           ============================================================ */
        <div>
          <div className="card overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="table-header">
                  <th className="px-4 py-3 text-left">Source</th>
                  <th className="px-4 py-3 text-left">Type</th>
                  <th className="px-4 py-3 text-left">Notifications</th>
                  <th className="px-4 py-3 text-left">Auto-Assess</th>
                  <th className="px-4 py-3 text-left">Status</th>
                  <th className="px-4 py-3 text-left">Actions</th>
                </tr>
              </thead>
              <tbody>
                {subscriptions.map(sub => (
                  <tr key={sub.id} className="border-t border-gray-100 hover:bg-gray-50">
                    <td className="px-4 py-3 font-medium text-gray-900">{sub.source_name}</td>
                    <td className="px-4 py-3 text-xs text-gray-600">
                      {SOURCE_TYPE_LABELS[sub.source_type] || sub.source_type}
                    </td>
                    <td className="px-4 py-3">
                      <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${
                        sub.notification_on_new ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-500'
                      }`}>
                        {sub.notification_on_new ? 'On' : 'Off'}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${
                        sub.auto_assess ? 'bg-blue-100 text-blue-700' : 'bg-gray-100 text-gray-500'
                      }`}>
                        {sub.auto_assess ? 'Enabled' : 'Disabled'}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${
                        sub.is_active ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-500'
                      }`}>
                        {sub.is_active ? 'Active' : 'Inactive'}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <button
                        onClick={() => handleUnsubscribe(sub.id)}
                        className="text-red-600 hover:text-red-700 text-xs font-medium"
                      >
                        Unsubscribe
                      </button>
                    </td>
                  </tr>
                ))}
                {subscriptions.length === 0 && (
                  <tr>
                    <td colSpan={6} className="px-4 py-12 text-center text-gray-500">
                      No active subscriptions. Go to Sources to subscribe.
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </div>

      ) : tab === 'timeline' ? (
        /* ============================================================
           TIMELINE TAB
           ============================================================ */
        <div className="card">
          <h3 className="font-semibold text-gray-900 mb-6">Regulatory Change Timeline (12 months)</h3>
          <div className="relative">
            {/* Vertical line */}
            <div className="absolute left-4 top-0 bottom-0 w-0.5 bg-gray-200" />
            <div className="space-y-4 pl-12">
              {timeline.map((entry, idx) => (
                <div key={`${entry.change_ref}-${entry.event_type}-${idx}`} className="relative">
                  {/* Dot */}
                  <div className={`absolute -left-8 top-1 h-3 w-3 rounded-full border-2 border-white ${
                    EVENT_TYPE_COLORS[entry.event_type] || 'bg-gray-400'
                  }`} />
                  <div className="flex items-start gap-3">
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2">
                        <span className="text-xs font-medium text-gray-400">
                          {new Date(entry.date).toLocaleDateString('en-GB', {
                            day: '2-digit', month: 'short', year: 'numeric',
                          })}
                        </span>
                        <span className={`inline-flex items-center rounded-full px-1.5 py-0.5 text-[10px] font-medium ${
                          entry.event_type === 'deadline' ? 'bg-red-100 text-red-700'
                            : entry.event_type === 'effective' ? 'bg-green-100 text-green-700'
                            : 'bg-blue-100 text-blue-700'
                        }`}>
                          {entry.event_type}
                        </span>
                        <span className={`inline-flex items-center rounded-full px-1.5 py-0.5 text-[10px] font-medium ${SEVERITY_COLORS[entry.severity]}`}>
                          {entry.severity}
                        </span>
                      </div>
                      <p className="text-sm font-medium text-gray-900 mt-0.5 truncate">{entry.title}</p>
                      <p className="text-xs text-gray-400">{entry.change_ref} - {CHANGE_TYPE_LABELS[entry.change_type] || entry.change_type}</p>
                    </div>
                    <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${STATUS_COLORS[entry.status]}`}>
                      {entry.status.replace(/_/g, ' ')}
                    </span>
                  </div>
                </div>
              ))}
              {timeline.length === 0 && (
                <p className="text-sm text-gray-400 py-4 text-center">No timeline events found</p>
              )}
            </div>
          </div>
        </div>

      ) : null}
    </div>
  );
}
