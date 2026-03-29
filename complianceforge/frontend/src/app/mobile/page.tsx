'use client';

import { useEffect, useState, useCallback } from 'react';
import api from '@/lib/api';

// ============================================================
// TYPES
// ============================================================

interface MobileDashboard {
  compliance_score: number;
  risk_counts: {
    critical: number;
    high: number;
    medium: number;
    low: number;
  };
  open_incidents: number;
  pending_approvals: number;
  overdue_deadlines: number;
  unread_notifications: number;
  last_updated: string;
}

interface ApprovalItem {
  id: string;
  instance_id: string;
  step_name: string;
  workflow_name: string;
  entity_type: string;
  entity_name: string;
  requested_by: string;
  requested_at: string;
  priority: string;
}

interface IncidentItem {
  id: string;
  reference: string;
  title: string;
  severity: string;
  status: string;
  category: string;
  assigned_to: string;
  reported_at: string;
  is_breachable: boolean;
}

interface DeadlineItem {
  id: string;
  type: string;
  title: string;
  due_date: string;
  days_left: number;
  priority: string;
  overdue: boolean;
}

interface ActivityItem {
  id: string;
  type: string;
  title: string;
  actor: string;
  timestamp: string;
}

type ActiveTab = 'overview' | 'approvals' | 'incidents' | 'deadlines' | 'activity';

// ============================================================
// MOBILE DASHBOARD PAGE
// ============================================================

export default function MobileDashboardPage() {
  const [dashboard, setDashboard] = useState<MobileDashboard | null>(null);
  const [approvals, setApprovals] = useState<ApprovalItem[]>([]);
  const [incidents, setIncidents] = useState<IncidentItem[]>([]);
  const [deadlines, setDeadlines] = useState<DeadlineItem[]>([]);
  const [activities, setActivities] = useState<ActivityItem[]>([]);
  const [activeTab, setActiveTab] = useState<ActiveTab>('overview');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [actionLoading, setActionLoading] = useState<string | null>(null);

  // Load dashboard data
  useEffect(() => {
    api.getMobileDashboard()
      .then((res) => setDashboard(res.data))
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, []);

  // Load tab-specific data
  useEffect(() => {
    switch (activeTab) {
      case 'approvals':
        api.getMobileApprovals().then((res) => setApprovals(res.data || [])).catch(() => {});
        break;
      case 'incidents':
        api.getMobileIncidents().then((res) => setIncidents(res.data || [])).catch(() => {});
        break;
      case 'deadlines':
        api.getMobileDeadlines(14).then((res) => setDeadlines(res.data?.deadlines || [])).catch(() => {});
        break;
      case 'activity':
        api.getMobileActivity(20).then((res) => setActivities(res.data?.activities || [])).catch(() => {});
        break;
    }
  }, [activeTab]);

  const handleApprove = useCallback(async (id: string) => {
    setActionLoading(id);
    try {
      await api.mobileApprove(id);
      setApprovals((prev) => prev.filter((a) => a.id !== id));
      if (dashboard) {
        setDashboard({ ...dashboard, pending_approvals: Math.max(0, dashboard.pending_approvals - 1) });
      }
    } catch (err: any) {
      alert(err.message || 'Failed to approve');
    } finally {
      setActionLoading(null);
    }
  }, [dashboard]);

  const handleReject = useCallback(async (id: string) => {
    const reason = prompt('Reason for rejection:');
    if (!reason) return;
    setActionLoading(id);
    try {
      await api.mobileReject(id, reason);
      setApprovals((prev) => prev.filter((a) => a.id !== id));
      if (dashboard) {
        setDashboard({ ...dashboard, pending_approvals: Math.max(0, dashboard.pending_approvals - 1) });
      }
    } catch (err: any) {
      alert(err.message || 'Failed to reject');
    } finally {
      setActionLoading(null);
    }
  }, [dashboard]);

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-gray-50">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-600 mx-auto" />
          <p className="text-gray-500 mt-3 text-sm">Loading dashboard...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-4 min-h-screen bg-gray-50">
        <div className="rounded-lg bg-red-50 border border-red-200 p-4 text-red-700 text-sm">
          {error}
        </div>
      </div>
    );
  }

  if (!dashboard) return null;

  return (
    <div className="min-h-screen bg-gray-50 pb-20">
      {/* Header */}
      <header className="bg-white border-b border-gray-200 px-4 py-3 sticky top-0 z-10">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-lg font-bold text-gray-900">ComplianceForge</h1>
            <p className="text-xs text-gray-500">Mobile Dashboard</p>
          </div>
          <div className="flex items-center gap-3">
            {dashboard.unread_notifications > 0 && (
              <a href="/notifications" className="relative">
                <svg className="w-6 h-6 text-gray-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
                    d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
                </svg>
                <span className="absolute -top-1 -right-1 bg-red-500 text-white text-xs rounded-full h-4 w-4 flex items-center justify-center">
                  {dashboard.unread_notifications > 9 ? '9+' : dashboard.unread_notifications}
                </span>
              </a>
            )}
          </div>
        </div>
      </header>

      {/* Tab Navigation */}
      <nav className="bg-white border-b border-gray-200 px-2 overflow-x-auto">
        <div className="flex gap-1 min-w-max">
          {([
            { key: 'overview', label: 'Overview' },
            { key: 'approvals', label: 'Approvals', badge: dashboard.pending_approvals },
            { key: 'incidents', label: 'Incidents', badge: dashboard.open_incidents },
            { key: 'deadlines', label: 'Deadlines', badge: dashboard.overdue_deadlines },
            { key: 'activity', label: 'Activity' },
          ] as { key: ActiveTab; label: string; badge?: number }[]).map((tab) => (
            <button
              key={tab.key}
              onClick={() => setActiveTab(tab.key)}
              className={`px-3 py-2.5 text-sm font-medium whitespace-nowrap border-b-2 transition-colors ${
                activeTab === tab.key
                  ? 'border-indigo-600 text-indigo-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700'
              }`}
            >
              {tab.label}
              {tab.badge !== undefined && tab.badge > 0 && (
                <span className={`ml-1.5 inline-flex items-center justify-center px-1.5 py-0.5 rounded-full text-xs font-bold ${
                  activeTab === tab.key ? 'bg-indigo-100 text-indigo-700' : 'bg-gray-100 text-gray-600'
                }`}>
                  {tab.badge}
                </span>
              )}
            </button>
          ))}
        </div>
      </nav>

      {/* Content */}
      <main className="p-4">
        {activeTab === 'overview' && <OverviewTab dashboard={dashboard} onTabChange={setActiveTab} />}
        {activeTab === 'approvals' && (
          <ApprovalsTab
            approvals={approvals}
            onApprove={handleApprove}
            onReject={handleReject}
            actionLoading={actionLoading}
          />
        )}
        {activeTab === 'incidents' && <IncidentsTab incidents={incidents} />}
        {activeTab === 'deadlines' && <DeadlinesTab deadlines={deadlines} />}
        {activeTab === 'activity' && <ActivityTab activities={activities} />}
      </main>
    </div>
  );
}

// ============================================================
// TAB COMPONENTS
// ============================================================

function OverviewTab({ dashboard, onTabChange }: { dashboard: MobileDashboard; onTabChange: (tab: ActiveTab) => void }) {
  const totalRisks = dashboard.risk_counts.critical + dashboard.risk_counts.high +
    dashboard.risk_counts.medium + dashboard.risk_counts.low;

  return (
    <div className="space-y-4">
      {/* Compliance Score */}
      <div className="bg-white rounded-xl border border-gray-200 p-4">
        <p className="text-xs font-medium text-gray-500 uppercase tracking-wide">Compliance Score</p>
        <div className="flex items-end gap-2 mt-1">
          <span className={`text-3xl font-bold ${getScoreColor(dashboard.compliance_score)}`}>
            {dashboard.compliance_score.toFixed(1)}%
          </span>
        </div>
        <div className="mt-2 h-2 w-full rounded-full bg-gray-100">
          <div
            className={`h-2 rounded-full transition-all ${getBarColor(dashboard.compliance_score)}`}
            style={{ width: `${Math.min(dashboard.compliance_score, 100)}%` }}
          />
        </div>
      </div>

      {/* Quick Stats Grid */}
      <div className="grid grid-cols-2 gap-3">
        <StatCard
          label="Open Risks"
          value={totalRisks}
          subtitle={`${dashboard.risk_counts.critical} critical`}
          color={dashboard.risk_counts.critical > 0 ? 'red' : 'amber'}
          onClick={() => {}}
        />
        <StatCard
          label="Open Incidents"
          value={dashboard.open_incidents}
          color={dashboard.open_incidents > 5 ? 'red' : 'amber'}
          onClick={() => onTabChange('incidents')}
        />
        <StatCard
          label="Pending Approvals"
          value={dashboard.pending_approvals}
          color={dashboard.pending_approvals > 0 ? 'amber' : 'green'}
          onClick={() => onTabChange('approvals')}
        />
        <StatCard
          label="Overdue"
          value={dashboard.overdue_deadlines}
          color={dashboard.overdue_deadlines > 0 ? 'red' : 'green'}
          onClick={() => onTabChange('deadlines')}
        />
      </div>

      {/* Risk Breakdown */}
      <div className="bg-white rounded-xl border border-gray-200 p-4">
        <h3 className="text-sm font-semibold text-gray-900 mb-3">Risk Distribution</h3>
        <div className="space-y-2">
          {([
            { level: 'Critical', count: dashboard.risk_counts.critical, color: 'bg-red-500' },
            { level: 'High', count: dashboard.risk_counts.high, color: 'bg-orange-500' },
            { level: 'Medium', count: dashboard.risk_counts.medium, color: 'bg-yellow-500' },
            { level: 'Low', count: dashboard.risk_counts.low, color: 'bg-green-500' },
          ]).map((risk) => (
            <div key={risk.level} className="flex items-center gap-3">
              <span className={`h-2.5 w-2.5 rounded-full ${risk.color}`} />
              <span className="text-sm text-gray-600 w-16">{risk.level}</span>
              <div className="flex-1 h-4 bg-gray-100 rounded">
                <div
                  className={`h-4 rounded ${risk.color} transition-all flex items-center px-1`}
                  style={{ width: `${totalRisks > 0 ? Math.max((risk.count / totalRisks) * 100, 4) : 0}%` }}
                >
                  {risk.count > 0 && <span className="text-xs font-bold text-white">{risk.count}</span>}
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Last Updated */}
      <p className="text-center text-xs text-gray-400">
        Last updated: {new Date(dashboard.last_updated).toLocaleTimeString()}
      </p>
    </div>
  );
}

function ApprovalsTab({
  approvals,
  onApprove,
  onReject,
  actionLoading,
}: {
  approvals: ApprovalItem[];
  onApprove: (id: string) => void;
  onReject: (id: string) => void;
  actionLoading: string | null;
}) {
  if (approvals.length === 0) {
    return <EmptyState message="No pending approvals" />;
  }

  return (
    <div className="space-y-3">
      {approvals.map((a) => (
        <div key={a.id} className="bg-white rounded-xl border border-gray-200 p-4">
          <div className="flex items-start justify-between">
            <div className="flex-1 min-w-0">
              <p className="font-medium text-gray-900 text-sm truncate">{a.step_name}</p>
              <p className="text-xs text-gray-500 mt-0.5">{a.workflow_name}</p>
              {a.entity_name && (
                <p className="text-xs text-gray-600 mt-1 truncate">{a.entity_type}: {a.entity_name}</p>
              )}
              <p className="text-xs text-gray-400 mt-1">
                Requested by {a.requested_by} &middot; {formatTimeAgo(a.requested_at)}
              </p>
            </div>
            <PriorityBadge priority={a.priority} />
          </div>
          <div className="flex gap-2 mt-3">
            <button
              onClick={() => onApprove(a.id)}
              disabled={actionLoading === a.id}
              className="flex-1 bg-green-600 text-white text-sm font-medium py-2 px-3 rounded-lg hover:bg-green-700 disabled:opacity-50 transition-colors"
            >
              {actionLoading === a.id ? 'Processing...' : 'Approve'}
            </button>
            <button
              onClick={() => onReject(a.id)}
              disabled={actionLoading === a.id}
              className="flex-1 bg-red-50 text-red-700 text-sm font-medium py-2 px-3 rounded-lg hover:bg-red-100 disabled:opacity-50 border border-red-200 transition-colors"
            >
              Reject
            </button>
          </div>
        </div>
      ))}
    </div>
  );
}

function IncidentsTab({ incidents }: { incidents: IncidentItem[] }) {
  if (incidents.length === 0) {
    return <EmptyState message="No active incidents" />;
  }

  return (
    <div className="space-y-3">
      {incidents.map((inc) => (
        <a key={inc.id} href={`/incidents?id=${inc.id}`} className="block bg-white rounded-xl border border-gray-200 p-4 hover:border-indigo-300 transition-colors">
          <div className="flex items-start justify-between">
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-2">
                <span className="text-xs font-mono text-gray-500">{inc.reference}</span>
                {inc.is_breachable && (
                  <span className="text-xs bg-red-100 text-red-700 px-1.5 py-0.5 rounded font-medium">BREACH</span>
                )}
              </div>
              <p className="font-medium text-gray-900 text-sm mt-1 truncate">{inc.title}</p>
              <p className="text-xs text-gray-500 mt-1">
                {inc.assigned_to} &middot; {formatTimeAgo(inc.reported_at)}
              </p>
            </div>
            <SeverityBadge severity={inc.severity} />
          </div>
          <div className="flex items-center gap-2 mt-2">
            <StatusPill status={inc.status} />
            {inc.category && <span className="text-xs text-gray-400">{inc.category}</span>}
          </div>
        </a>
      ))}
    </div>
  );
}

function DeadlinesTab({ deadlines }: { deadlines: DeadlineItem[] }) {
  if (deadlines.length === 0) {
    return <EmptyState message="No upcoming deadlines" />;
  }

  return (
    <div className="space-y-3">
      {deadlines.map((d) => (
        <div key={d.id} className={`bg-white rounded-xl border p-4 ${d.overdue ? 'border-red-200' : 'border-gray-200'}`}>
          <div className="flex items-start justify-between">
            <div className="flex-1 min-w-0">
              <p className="font-medium text-gray-900 text-sm truncate">{d.title}</p>
              <p className="text-xs text-gray-500 mt-0.5 capitalize">{d.type.replace(/_/g, ' ')}</p>
            </div>
            <PriorityBadge priority={d.priority} />
          </div>
          <div className="flex items-center justify-between mt-2">
            <span className="text-xs text-gray-500">
              {new Date(d.due_date).toLocaleDateString()}
            </span>
            <span className={`text-xs font-medium ${d.overdue ? 'text-red-600' : d.days_left <= 2 ? 'text-orange-600' : 'text-gray-600'}`}>
              {d.overdue ? `${Math.abs(d.days_left)} days overdue` : d.days_left === 0 ? 'Due today' : `${d.days_left} days left`}
            </span>
          </div>
        </div>
      ))}
    </div>
  );
}

function ActivityTab({ activities }: { activities: ActivityItem[] }) {
  if (activities.length === 0) {
    return <EmptyState message="No recent activity" />;
  }

  return (
    <div className="space-y-2">
      {activities.map((a) => (
        <div key={a.id} className="bg-white rounded-lg border border-gray-200 px-4 py-3">
          <div className="flex items-center gap-3">
            <ActivityIcon type={a.type} />
            <div className="flex-1 min-w-0">
              <p className="text-sm text-gray-900 truncate">{a.title}</p>
              <p className="text-xs text-gray-500">{a.actor} &middot; {formatTimeAgo(a.timestamp)}</p>
            </div>
          </div>
        </div>
      ))}
    </div>
  );
}

// ============================================================
// SHARED COMPONENTS
// ============================================================

function StatCard({ label, value, subtitle, color, onClick }: {
  label: string;
  value: number;
  subtitle?: string;
  color: 'green' | 'amber' | 'red';
  onClick?: () => void;
}) {
  const styles = {
    green: { border: 'border-green-200', bg: 'bg-green-50', text: 'text-green-700' },
    amber: { border: 'border-amber-200', bg: 'bg-amber-50', text: 'text-amber-700' },
    red: { border: 'border-red-200', bg: 'bg-red-50', text: 'text-red-700' },
  }[color];

  return (
    <button onClick={onClick} className={`rounded-xl border ${styles.border} ${styles.bg} p-4 text-left w-full`}>
      <p className="text-xs font-medium text-gray-500 uppercase tracking-wide">{label}</p>
      <p className={`text-2xl font-bold ${styles.text} mt-1`}>{value}</p>
      {subtitle && <p className="text-xs text-gray-500 mt-0.5">{subtitle}</p>}
    </button>
  );
}

function SeverityBadge({ severity }: { severity: string }) {
  const styles: Record<string, string> = {
    critical: 'bg-red-100 text-red-700',
    high: 'bg-orange-100 text-orange-700',
    medium: 'bg-yellow-100 text-yellow-700',
    low: 'bg-green-100 text-green-700',
  };
  return (
    <span className={`text-xs font-medium px-2 py-0.5 rounded ${styles[severity] || 'bg-gray-100 text-gray-700'}`}>
      {severity.toUpperCase()}
    </span>
  );
}

function PriorityBadge({ priority }: { priority: string }) {
  const styles: Record<string, string> = {
    critical: 'bg-red-100 text-red-700',
    high: 'bg-orange-100 text-orange-700',
    medium: 'bg-yellow-100 text-yellow-700',
    low: 'bg-green-100 text-green-700',
    normal: 'bg-gray-100 text-gray-700',
  };
  return (
    <span className={`text-xs font-medium px-2 py-0.5 rounded ${styles[priority] || 'bg-gray-100 text-gray-700'}`}>
      {priority}
    </span>
  );
}

function StatusPill({ status }: { status: string }) {
  const styles: Record<string, string> = {
    open: 'bg-blue-100 text-blue-700',
    investigating: 'bg-purple-100 text-purple-700',
    contained: 'bg-yellow-100 text-yellow-700',
    resolved: 'bg-green-100 text-green-700',
    closed: 'bg-gray-100 text-gray-600',
  };
  return (
    <span className={`text-xs font-medium px-2 py-0.5 rounded capitalize ${styles[status] || 'bg-gray-100 text-gray-700'}`}>
      {status.replace(/_/g, ' ')}
    </span>
  );
}

function ActivityIcon({ type }: { type: string }) {
  const icons: Record<string, { bg: string; symbol: string }> = {
    incident: { bg: 'bg-red-100', symbol: '!' },
    policy: { bg: 'bg-blue-100', symbol: 'P' },
    risk: { bg: 'bg-orange-100', symbol: 'R' },
    audit: { bg: 'bg-purple-100', symbol: 'A' },
    vendor: { bg: 'bg-teal-100', symbol: 'V' },
  };
  const icon = icons[type] || { bg: 'bg-gray-100', symbol: '?' };
  return (
    <span className={`flex items-center justify-center h-8 w-8 rounded-full ${icon.bg} text-xs font-bold text-gray-700`}>
      {icon.symbol}
    </span>
  );
}

function EmptyState({ message }: { message: string }) {
  return (
    <div className="text-center py-12">
      <div className="text-gray-400 text-4xl mb-3">---</div>
      <p className="text-gray-500 text-sm">{message}</p>
    </div>
  );
}

// ============================================================
// HELPERS
// ============================================================

function formatTimeAgo(dateStr: string): string {
  const now = Date.now();
  const then = new Date(dateStr).getTime();
  const diffMs = now - then;
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMs / 3600000);
  const diffDays = Math.floor(diffMs / 86400000);

  if (diffMins < 1) return 'just now';
  if (diffMins < 60) return `${diffMins}m ago`;
  if (diffHours < 24) return `${diffHours}h ago`;
  if (diffDays < 7) return `${diffDays}d ago`;
  return new Date(dateStr).toLocaleDateString();
}

function getScoreColor(score: number): string {
  if (score >= 80) return 'text-green-600';
  if (score >= 60) return 'text-amber-600';
  return 'text-red-600';
}

function getBarColor(score: number): string {
  if (score >= 80) return 'bg-green-500';
  if (score >= 60) return 'bg-amber-500';
  return 'bg-red-500';
}
