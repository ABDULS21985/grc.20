'use client';

import { useEffect, useState, useCallback } from 'react';
import api from '@/lib/api';

// ============================================================
// TYPES
// ============================================================

interface BoardMember {
  id: string;
  name: string;
  title: string;
  email: string;
  member_type: string;
  committees: string[];
  is_active: boolean;
  portal_access_enabled: boolean;
  last_portal_access_at: string | null;
  created_at: string;
}

interface BoardMeeting {
  id: string;
  meeting_ref: string;
  title: string;
  meeting_type: string;
  date: string;
  time: string | null;
  location: string;
  status: string;
  agenda_items: any[];
  board_pack_document_path: string | null;
  board_pack_generated_at: string | null;
  attendees: string[];
  apologies: string[];
  created_at: string;
}

interface BoardDecision {
  id: string;
  decision_ref: string;
  title: string;
  description: string;
  decision_type: string;
  decision: string;
  vote_for: number;
  vote_against: number;
  vote_abstain: number;
  action_required: boolean;
  action_description: string;
  action_due_date: string | null;
  action_status: string;
  decided_at: string | null;
  decided_by: string;
  meeting_ref_display: string;
  tags: string[];
}

interface BoardReport {
  id: string;
  report_type: string;
  title: string;
  period_start: string | null;
  period_end: string | null;
  file_path: string;
  file_format: string;
  generated_at: string;
  classification: string;
  page_count: number;
}

interface BoardDashboard {
  compliance_score: number;
  risk_appetite_status: string;
  open_risks: { total: number; critical: number; high: number };
  open_incidents: { total: number; critical: number; open: number };
  pending_actions: number;
  overdue_actions: number;
  total_members: number;
  reports_generated: number;
  next_meeting: BoardMeeting | null;
  last_board_pack_at: string | null;
}

// ============================================================
// COLOUR MAPS
// ============================================================

const STATUS_COLORS: Record<string, string> = {
  planned: 'bg-blue-100 text-blue-800',
  agenda_set: 'bg-indigo-100 text-indigo-800',
  in_progress: 'bg-yellow-100 text-yellow-800',
  completed: 'bg-green-100 text-green-800',
  minutes_approved: 'bg-emerald-100 text-emerald-800',
};

const DECISION_COLORS: Record<string, string> = {
  approved: 'bg-green-100 text-green-800',
  rejected: 'bg-red-100 text-red-800',
  deferred: 'bg-yellow-100 text-yellow-800',
  conditional_approval: 'bg-orange-100 text-orange-800',
};

const ACTION_COLORS: Record<string, string> = {
  pending: 'bg-gray-100 text-gray-700',
  in_progress: 'bg-blue-100 text-blue-800',
  completed: 'bg-green-100 text-green-800',
  overdue: 'bg-red-100 text-red-800',
  cancelled: 'bg-gray-200 text-gray-500',
};

const MEMBER_TYPES: Record<string, string> = {
  executive_director: 'Executive Director',
  non_executive_director: 'Non-Executive Director',
  independent_director: 'Independent Director',
  committee_chair: 'Committee Chair',
  observer: 'Observer',
  secretary: 'Secretary',
};

const MEETING_TYPES: Record<string, string> = {
  full_board: 'Full Board',
  audit_committee: 'Audit Committee',
  risk_committee: 'Risk Committee',
  remuneration_committee: 'Remuneration Committee',
  nomination_committee: 'Nomination Committee',
  special: 'Special Meeting',
  agm: 'AGM',
  egm: 'EGM',
};

// ============================================================
// MAIN PAGE COMPONENT
// ============================================================

export default function BoardManagementPage() {
  const [activeTab, setActiveTab] = useState<'dashboard' | 'members' | 'meetings' | 'decisions' | 'reports'>('dashboard');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Data state
  const [dashboard, setDashboard] = useState<BoardDashboard | null>(null);
  const [members, setMembers] = useState<BoardMember[]>([]);
  const [meetings, setMeetings] = useState<BoardMeeting[]>([]);
  const [decisions, setDecisions] = useState<BoardDecision[]>([]);
  const [reports, setReports] = useState<BoardReport[]>([]);

  // Modals
  const [showCreateMember, setShowCreateMember] = useState(false);
  const [showCreateMeeting, setShowCreateMeeting] = useState(false);
  const [showRecordDecision, setShowRecordDecision] = useState(false);
  const [showGenerateReport, setShowGenerateReport] = useState(false);

  // Form state for new member
  const [newMember, setNewMember] = useState({
    name: '', title: '', email: '', member_type: 'non_executive_director',
    committees: '' as string, portal_access_enabled: false,
  });

  // Form state for new meeting
  const [newMeeting, setNewMeeting] = useState({
    title: '', meeting_type: 'full_board', date: '', time: '', location: '',
  });

  // Form state for new decision
  const [newDecision, setNewDecision] = useState({
    meeting_id: '', title: '', description: '', decision_type: 'general',
    decision: 'approved', vote_for: 0, vote_against: 0, vote_abstain: 0,
    rationale: '', action_required: false, action_description: '', action_due_date: '',
    decided_by: '',
  });

  // Form state for report generation
  const [reportForm, setReportForm] = useState({
    report_type: 'board_pack', title: '', period_start: '', period_end: '', file_format: 'pdf',
  });

  const loadDashboard = useCallback(async () => {
    try {
      const res = await api.request<any>('/board/dashboard');
      if (res.data) setDashboard(res.data);
    } catch { /* ignore */ }
  }, []);

  const loadMembers = useCallback(async () => {
    try {
      const res = await api.request<any>('/board/members');
      if (res.data) setMembers(res.data);
    } catch { /* ignore */ }
  }, []);

  const loadMeetings = useCallback(async () => {
    try {
      const res = await api.request<any>('/board/meetings');
      if (res.data) setMeetings(res.data);
    } catch { /* ignore */ }
  }, []);

  const loadDecisions = useCallback(async () => {
    try {
      const res = await api.request<any>('/board/decisions');
      if (res.data) setDecisions(res.data);
    } catch { /* ignore */ }
  }, []);

  const loadReports = useCallback(async () => {
    try {
      const res = await api.request<any>('/board/reports');
      if (res.data) setReports(res.data);
    } catch { /* ignore */ }
  }, []);

  useEffect(() => {
    const loadAll = async () => {
      setLoading(true);
      setError(null);
      try {
        await Promise.all([loadDashboard(), loadMembers(), loadMeetings(), loadDecisions(), loadReports()]);
      } catch (err: any) {
        setError(err.message || 'Failed to load board data');
      } finally {
        setLoading(false);
      }
    };
    loadAll();
  }, [loadDashboard, loadMembers, loadMeetings, loadDecisions, loadReports]);

  const handleCreateMember = async () => {
    try {
      await api.request<any>('/board/members', {
        method: 'POST',
        body: {
          ...newMember,
          committees: newMember.committees ? newMember.committees.split(',').map(s => s.trim()) : [],
        },
      });
      setShowCreateMember(false);
      setNewMember({ name: '', title: '', email: '', member_type: 'non_executive_director', committees: '', portal_access_enabled: false });
      await loadMembers();
    } catch (err: any) {
      setError(err.message);
    }
  };

  const handleCreateMeeting = async () => {
    try {
      await api.request<any>('/board/meetings', {
        method: 'POST',
        body: newMeeting,
      });
      setShowCreateMeeting(false);
      setNewMeeting({ title: '', meeting_type: 'full_board', date: '', time: '', location: '' });
      await loadMeetings();
    } catch (err: any) {
      setError(err.message);
    }
  };

  const handleRecordDecision = async () => {
    try {
      await api.request<any>('/board/decisions', {
        method: 'POST',
        body: {
          ...newDecision,
          vote_for: Number(newDecision.vote_for),
          vote_against: Number(newDecision.vote_against),
          vote_abstain: Number(newDecision.vote_abstain),
          tags: [],
        },
      });
      setShowRecordDecision(false);
      setNewDecision({
        meeting_id: '', title: '', description: '', decision_type: 'general',
        decision: 'approved', vote_for: 0, vote_against: 0, vote_abstain: 0,
        rationale: '', action_required: false, action_description: '', action_due_date: '',
        decided_by: '',
      });
      await loadDecisions();
    } catch (err: any) {
      setError(err.message);
    }
  };

  const handleGenerateBoardPack = async (meetingId: string) => {
    try {
      await api.request<any>(`/board/meetings/${meetingId}/generate-pack`, { method: 'POST' });
      await Promise.all([loadMeetings(), loadReports()]);
    } catch (err: any) {
      setError(err.message);
    }
  };

  const handleGenerateReport = async () => {
    try {
      await api.request<any>('/board/reports/generate', {
        method: 'POST',
        body: reportForm,
      });
      setShowGenerateReport(false);
      setReportForm({ report_type: 'board_pack', title: '', period_start: '', period_end: '', file_format: 'pdf' });
      await loadReports();
    } catch (err: any) {
      setError(err.message);
    }
  };

  const handleUpdateAction = async (decisionId: string, status: string) => {
    try {
      await api.request<any>(`/board/decisions/${decisionId}/action`, {
        method: 'PUT',
        body: { action_status: status },
      });
      await loadDecisions();
    } catch (err: any) {
      setError(err.message);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-600" />
        <span className="ml-3 text-gray-600">Loading board data...</span>
      </div>
    );
  }

  return (
    <div className="max-w-7xl mx-auto px-4 py-8">
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">Board Reporting & Governance</h1>
          <p className="text-gray-500 mt-1">Executive oversight, meeting management, and governance dashboards</p>
        </div>
      </div>

      {error && (
        <div className="mb-4 bg-red-50 border border-red-200 rounded-lg p-4">
          <p className="text-red-800 text-sm">{error}</p>
          <button onClick={() => setError(null)} className="text-red-600 text-xs underline mt-1">Dismiss</button>
        </div>
      )}

      {/* Tab Navigation */}
      <div className="border-b border-gray-200 mb-6">
        <nav className="flex space-x-8" aria-label="Tabs">
          {(['dashboard', 'members', 'meetings', 'decisions', 'reports'] as const).map(tab => (
            <button
              key={tab}
              onClick={() => setActiveTab(tab)}
              className={`py-4 px-1 border-b-2 font-medium text-sm capitalize ${
                activeTab === tab
                  ? 'border-indigo-500 text-indigo-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              {tab}
            </button>
          ))}
        </nav>
      </div>

      {/* ── DASHBOARD TAB ── */}
      {activeTab === 'dashboard' && dashboard && (
        <div className="space-y-6">
          {/* Key Metrics */}
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <MetricCard label="Compliance Score" value={`${dashboard.compliance_score.toFixed(1)}%`}
              color={dashboard.compliance_score >= 80 ? 'green' : dashboard.compliance_score >= 60 ? 'yellow' : 'red'} />
            <MetricCard label="Risk Appetite" value={dashboard.risk_appetite_status.replace(/_/g, ' ')}
              color={dashboard.risk_appetite_status === 'within_appetite' ? 'green' : dashboard.risk_appetite_status === 'at_limit' ? 'yellow' : 'red'} />
            <MetricCard label="Pending Actions" value={String(dashboard.pending_actions)}
              color={dashboard.pending_actions === 0 ? 'green' : 'yellow'} />
            <MetricCard label="Overdue Actions" value={String(dashboard.overdue_actions)}
              color={dashboard.overdue_actions === 0 ? 'green' : 'red'} />
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <MetricCard label="Open Risks" value={`${dashboard.open_risks.total} (${dashboard.open_risks.critical} critical)`}
              color={dashboard.open_risks.critical > 0 ? 'red' : 'blue'} />
            <MetricCard label="Open Incidents" value={`${dashboard.open_incidents.total} total, ${dashboard.open_incidents.open} open`}
              color={dashboard.open_incidents.critical > 0 ? 'red' : 'blue'} />
            <MetricCard label="Board Members" value={String(dashboard.total_members)} color="blue" />
          </div>

          {/* Next Meeting */}
          {dashboard.next_meeting && (
            <div className="bg-white rounded-lg shadow p-6">
              <h3 className="text-lg font-semibold text-gray-900 mb-2">Next Board Meeting</h3>
              <div className="flex items-center space-x-4">
                <span className="text-2xl font-bold text-indigo-600">{dashboard.next_meeting.date}</span>
                <div>
                  <p className="font-medium">{dashboard.next_meeting.title}</p>
                  <p className="text-sm text-gray-500">{dashboard.next_meeting.meeting_ref} &middot; {dashboard.next_meeting.location || 'TBC'}</p>
                </div>
                <span className={`px-2 py-1 rounded text-xs font-medium ${STATUS_COLORS[dashboard.next_meeting.status] || 'bg-gray-100'}`}>
                  {dashboard.next_meeting.status.replace(/_/g, ' ')}
                </span>
              </div>
            </div>
          )}
        </div>
      )}

      {/* ── MEMBERS TAB ── */}
      {activeTab === 'members' && (
        <div>
          <div className="flex justify-between items-center mb-4">
            <h2 className="text-xl font-semibold">Board Members ({members.length})</h2>
            <button onClick={() => setShowCreateMember(true)}
              className="bg-indigo-600 text-white px-4 py-2 rounded-lg text-sm hover:bg-indigo-700">
              Add Member
            </button>
          </div>
          <div className="bg-white shadow rounded-lg overflow-hidden">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Name</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Type</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Committees</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Portal</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {members.map(m => (
                  <tr key={m.id} className="hover:bg-gray-50">
                    <td className="px-6 py-4">
                      <div className="font-medium text-gray-900">{m.name}</div>
                      <div className="text-sm text-gray-500">{m.title}</div>
                      <div className="text-xs text-gray-400">{m.email}</div>
                    </td>
                    <td className="px-6 py-4 text-sm">{MEMBER_TYPES[m.member_type] || m.member_type}</td>
                    <td className="px-6 py-4 text-sm">
                      {m.committees.length > 0 ? m.committees.join(', ') : <span className="text-gray-400">None</span>}
                    </td>
                    <td className="px-6 py-4">
                      <span className={`px-2 py-1 rounded text-xs font-medium ${m.is_active ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-600'}`}>
                        {m.is_active ? 'Active' : 'Inactive'}
                      </span>
                    </td>
                    <td className="px-6 py-4 text-sm">
                      {m.portal_access_enabled ? (
                        <span className="text-green-600">Enabled</span>
                      ) : (
                        <span className="text-gray-400">Disabled</span>
                      )}
                    </td>
                  </tr>
                ))}
                {members.length === 0 && (
                  <tr><td colSpan={5} className="px-6 py-8 text-center text-gray-500">No board members yet. Add your first member above.</td></tr>
                )}
              </tbody>
            </table>
          </div>

          {/* Create Member Modal */}
          {showCreateMember && (
            <Modal title="Add Board Member" onClose={() => setShowCreateMember(false)}>
              <div className="space-y-4">
                <Input label="Full Name" value={newMember.name} onChange={v => setNewMember({ ...newMember, name: v })} required />
                <Input label="Title" value={newMember.title} onChange={v => setNewMember({ ...newMember, title: v })} />
                <Input label="Email" value={newMember.email} onChange={v => setNewMember({ ...newMember, email: v })} />
                <Select label="Member Type" value={newMember.member_type}
                  onChange={v => setNewMember({ ...newMember, member_type: v })}
                  options={Object.entries(MEMBER_TYPES).map(([k, v]) => ({ value: k, label: v }))} />
                <Input label="Committees (comma-separated)" value={newMember.committees}
                  onChange={v => setNewMember({ ...newMember, committees: v })} />
                <label className="flex items-center space-x-2">
                  <input type="checkbox" checked={newMember.portal_access_enabled}
                    onChange={e => setNewMember({ ...newMember, portal_access_enabled: e.target.checked })} />
                  <span className="text-sm">Enable portal access</span>
                </label>
                <div className="flex justify-end space-x-3 pt-4">
                  <button onClick={() => setShowCreateMember(false)} className="px-4 py-2 border rounded-lg text-sm">Cancel</button>
                  <button onClick={handleCreateMember} className="px-4 py-2 bg-indigo-600 text-white rounded-lg text-sm hover:bg-indigo-700">Create</button>
                </div>
              </div>
            </Modal>
          )}
        </div>
      )}

      {/* ── MEETINGS TAB ── */}
      {activeTab === 'meetings' && (
        <div>
          <div className="flex justify-between items-center mb-4">
            <h2 className="text-xl font-semibold">Board Meetings ({meetings.length})</h2>
            <button onClick={() => setShowCreateMeeting(true)}
              className="bg-indigo-600 text-white px-4 py-2 rounded-lg text-sm hover:bg-indigo-700">
              Schedule Meeting
            </button>
          </div>

          <div className="space-y-4">
            {meetings.map(m => (
              <div key={m.id} className="bg-white shadow rounded-lg p-6">
                <div className="flex items-center justify-between">
                  <div>
                    <div className="flex items-center space-x-3">
                      <span className="text-sm font-mono text-gray-500">{m.meeting_ref}</span>
                      <span className={`px-2 py-1 rounded text-xs font-medium ${STATUS_COLORS[m.status] || 'bg-gray-100'}`}>
                        {m.status.replace(/_/g, ' ')}
                      </span>
                    </div>
                    <h3 className="text-lg font-semibold text-gray-900 mt-1">{m.title}</h3>
                    <p className="text-sm text-gray-500 mt-1">
                      {m.date}{m.time ? ` at ${m.time}` : ''} &middot; {MEETING_TYPES[m.meeting_type] || m.meeting_type}
                      {m.location && ` &middot; ${m.location}`}
                    </p>
                    <p className="text-xs text-gray-400 mt-1">
                      {m.attendees.length} attendees &middot; {m.apologies.length} apologies
                    </p>
                  </div>
                  <div className="flex space-x-2">
                    {!m.board_pack_document_path && (
                      <button onClick={() => handleGenerateBoardPack(m.id)}
                        className="px-3 py-1.5 bg-emerald-600 text-white rounded text-xs hover:bg-emerald-700">
                        Generate Pack
                      </button>
                    )}
                    {m.board_pack_document_path && (
                      <span className="px-3 py-1.5 bg-green-100 text-green-700 rounded text-xs">
                        Pack Ready
                      </span>
                    )}
                  </div>
                </div>
              </div>
            ))}
            {meetings.length === 0 && (
              <div className="bg-white shadow rounded-lg p-8 text-center text-gray-500">
                No meetings scheduled. Click &ldquo;Schedule Meeting&rdquo; to get started.
              </div>
            )}
          </div>

          {/* Create Meeting Modal */}
          {showCreateMeeting && (
            <Modal title="Schedule Board Meeting" onClose={() => setShowCreateMeeting(false)}>
              <div className="space-y-4">
                <Input label="Title" value={newMeeting.title} onChange={v => setNewMeeting({ ...newMeeting, title: v })} required />
                <Select label="Meeting Type" value={newMeeting.meeting_type}
                  onChange={v => setNewMeeting({ ...newMeeting, meeting_type: v })}
                  options={Object.entries(MEETING_TYPES).map(([k, v]) => ({ value: k, label: v }))} />
                <Input label="Date" type="date" value={newMeeting.date} onChange={v => setNewMeeting({ ...newMeeting, date: v })} required />
                <Input label="Time" type="time" value={newMeeting.time} onChange={v => setNewMeeting({ ...newMeeting, time: v })} />
                <Input label="Location" value={newMeeting.location} onChange={v => setNewMeeting({ ...newMeeting, location: v })} />
                <div className="flex justify-end space-x-3 pt-4">
                  <button onClick={() => setShowCreateMeeting(false)} className="px-4 py-2 border rounded-lg text-sm">Cancel</button>
                  <button onClick={handleCreateMeeting} className="px-4 py-2 bg-indigo-600 text-white rounded-lg text-sm hover:bg-indigo-700">Schedule</button>
                </div>
              </div>
            </Modal>
          )}
        </div>
      )}

      {/* ── DECISIONS TAB ── */}
      {activeTab === 'decisions' && (
        <div>
          <div className="flex justify-between items-center mb-4">
            <h2 className="text-xl font-semibold">Board Decisions ({decisions.length})</h2>
            <button onClick={() => setShowRecordDecision(true)}
              className="bg-indigo-600 text-white px-4 py-2 rounded-lg text-sm hover:bg-indigo-700">
              Record Decision
            </button>
          </div>

          <div className="space-y-4">
            {decisions.map(d => (
              <div key={d.id} className="bg-white shadow rounded-lg p-6">
                <div className="flex items-center justify-between mb-2">
                  <div className="flex items-center space-x-3">
                    <span className="text-sm font-mono text-gray-500">{d.decision_ref}</span>
                    <span className={`px-2 py-1 rounded text-xs font-medium ${DECISION_COLORS[d.decision] || 'bg-gray-100'}`}>
                      {d.decision.replace(/_/g, ' ')}
                    </span>
                    {d.meeting_ref_display && (
                      <span className="text-xs text-gray-400">@ {d.meeting_ref_display}</span>
                    )}
                  </div>
                  <div className="text-xs text-gray-400">
                    {d.decided_at ? new Date(d.decided_at).toLocaleDateString() : ''}
                    {d.decided_by ? ` by ${d.decided_by}` : ''}
                  </div>
                </div>
                <h3 className="text-base font-semibold text-gray-900">{d.title}</h3>
                {d.description && <p className="text-sm text-gray-600 mt-1">{d.description}</p>}

                <div className="flex items-center space-x-4 mt-3 text-xs text-gray-500">
                  <span>For: {d.vote_for}</span>
                  <span>Against: {d.vote_against}</span>
                  <span>Abstain: {d.vote_abstain}</span>
                  {d.tags.length > 0 && (
                    <span>Tags: {d.tags.join(', ')}</span>
                  )}
                </div>

                {d.action_required && (
                  <div className="mt-3 pt-3 border-t border-gray-100">
                    <div className="flex items-center justify-between">
                      <div>
                        <span className="text-xs font-medium text-gray-500">ACTION REQUIRED: </span>
                        <span className="text-sm text-gray-700">{d.action_description || 'No description'}</span>
                        {d.action_due_date && (
                          <span className="text-xs text-gray-400 ml-2">Due: {d.action_due_date}</span>
                        )}
                      </div>
                      <div className="flex items-center space-x-2">
                        <span className={`px-2 py-1 rounded text-xs font-medium ${ACTION_COLORS[d.action_status] || 'bg-gray-100'}`}>
                          {d.action_status.replace(/_/g, ' ')}
                        </span>
                        {d.action_status !== 'completed' && d.action_status !== 'cancelled' && (
                          <button
                            onClick={() => handleUpdateAction(d.id, 'completed')}
                            className="px-2 py-1 bg-green-600 text-white rounded text-xs hover:bg-green-700"
                          >
                            Complete
                          </button>
                        )}
                      </div>
                    </div>
                  </div>
                )}
              </div>
            ))}
            {decisions.length === 0 && (
              <div className="bg-white shadow rounded-lg p-8 text-center text-gray-500">
                No decisions recorded yet. Record your first board decision.
              </div>
            )}
          </div>

          {/* Record Decision Modal */}
          {showRecordDecision && (
            <Modal title="Record Board Decision" onClose={() => setShowRecordDecision(false)}>
              <div className="space-y-4 max-h-[70vh] overflow-y-auto">
                <Select label="Meeting" value={newDecision.meeting_id}
                  onChange={v => setNewDecision({ ...newDecision, meeting_id: v })}
                  options={[{ value: '', label: 'Select meeting...' }, ...meetings.map(m => ({ value: m.id, label: `${m.meeting_ref} - ${m.title}` }))]} />
                <Input label="Title" value={newDecision.title} onChange={v => setNewDecision({ ...newDecision, title: v })} required />
                <Textarea label="Description" value={newDecision.description} onChange={v => setNewDecision({ ...newDecision, description: v })} />
                <Select label="Decision Type" value={newDecision.decision_type}
                  onChange={v => setNewDecision({ ...newDecision, decision_type: v })}
                  options={[
                    { value: 'general', label: 'General' }, { value: 'policy_approval', label: 'Policy Approval' },
                    { value: 'risk_acceptance', label: 'Risk Acceptance' }, { value: 'budget_approval', label: 'Budget Approval' },
                    { value: 'strategy_approval', label: 'Strategy Approval' }, { value: 'compliance_action', label: 'Compliance Action' },
                    { value: 'incident_response', label: 'Incident Response' }, { value: 'vendor_approval', label: 'Vendor Approval' },
                    { value: 'regulatory_response', label: 'Regulatory Response' },
                  ]} />
                <Select label="Outcome" value={newDecision.decision}
                  onChange={v => setNewDecision({ ...newDecision, decision: v })}
                  options={[
                    { value: 'approved', label: 'Approved' }, { value: 'rejected', label: 'Rejected' },
                    { value: 'deferred', label: 'Deferred' }, { value: 'conditional_approval', label: 'Conditional Approval' },
                  ]} />
                <div className="grid grid-cols-3 gap-3">
                  <Input label="Votes For" type="number" value={String(newDecision.vote_for)}
                    onChange={v => setNewDecision({ ...newDecision, vote_for: parseInt(v) || 0 })} />
                  <Input label="Votes Against" type="number" value={String(newDecision.vote_against)}
                    onChange={v => setNewDecision({ ...newDecision, vote_against: parseInt(v) || 0 })} />
                  <Input label="Abstentions" type="number" value={String(newDecision.vote_abstain)}
                    onChange={v => setNewDecision({ ...newDecision, vote_abstain: parseInt(v) || 0 })} />
                </div>
                <Textarea label="Rationale" value={newDecision.rationale} onChange={v => setNewDecision({ ...newDecision, rationale: v })} />
                <Input label="Decided By" value={newDecision.decided_by} onChange={v => setNewDecision({ ...newDecision, decided_by: v })} />
                <label className="flex items-center space-x-2">
                  <input type="checkbox" checked={newDecision.action_required}
                    onChange={e => setNewDecision({ ...newDecision, action_required: e.target.checked })} />
                  <span className="text-sm">Requires follow-up action</span>
                </label>
                {newDecision.action_required && (
                  <>
                    <Textarea label="Action Description" value={newDecision.action_description}
                      onChange={v => setNewDecision({ ...newDecision, action_description: v })} />
                    <Input label="Action Due Date" type="date" value={newDecision.action_due_date}
                      onChange={v => setNewDecision({ ...newDecision, action_due_date: v })} />
                  </>
                )}
                <div className="flex justify-end space-x-3 pt-4">
                  <button onClick={() => setShowRecordDecision(false)} className="px-4 py-2 border rounded-lg text-sm">Cancel</button>
                  <button onClick={handleRecordDecision} className="px-4 py-2 bg-indigo-600 text-white rounded-lg text-sm hover:bg-indigo-700">Record Decision</button>
                </div>
              </div>
            </Modal>
          )}
        </div>
      )}

      {/* ── REPORTS TAB ── */}
      {activeTab === 'reports' && (
        <div>
          <div className="flex justify-between items-center mb-4">
            <h2 className="text-xl font-semibold">Board Reports ({reports.length})</h2>
            <div className="flex space-x-2">
              <button onClick={() => setShowGenerateReport(true)}
                className="bg-indigo-600 text-white px-4 py-2 rounded-lg text-sm hover:bg-indigo-700">
                Generate Report
              </button>
              <button onClick={async () => {
                try {
                  await api.request<any>('/board/nis2-governance');
                  await loadReports();
                } catch (err: any) { setError(err.message); }
              }}
                className="bg-emerald-600 text-white px-4 py-2 rounded-lg text-sm hover:bg-emerald-700">
                NIS2 Governance Report
              </button>
            </div>
          </div>

          <div className="bg-white shadow rounded-lg overflow-hidden">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Report</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Type</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Period</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Format</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Generated</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Classification</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {reports.map(r => (
                  <tr key={r.id} className="hover:bg-gray-50">
                    <td className="px-6 py-4">
                      <div className="font-medium text-gray-900 text-sm">{r.title}</div>
                      <div className="text-xs text-gray-400">{r.page_count} pages</div>
                    </td>
                    <td className="px-6 py-4 text-sm capitalize">{r.report_type.replace(/_/g, ' ')}</td>
                    <td className="px-6 py-4 text-sm text-gray-500">
                      {r.period_start && r.period_end ? `${r.period_start} - ${r.period_end}` : '-'}
                    </td>
                    <td className="px-6 py-4 text-sm uppercase">{r.file_format}</td>
                    <td className="px-6 py-4 text-sm text-gray-500">{new Date(r.generated_at).toLocaleDateString()}</td>
                    <td className="px-6 py-4">
                      <span className="px-2 py-1 rounded text-xs font-medium bg-purple-100 text-purple-800">
                        {r.classification.replace(/_/g, ' ')}
                      </span>
                    </td>
                  </tr>
                ))}
                {reports.length === 0 && (
                  <tr><td colSpan={6} className="px-6 py-8 text-center text-gray-500">No reports generated yet.</td></tr>
                )}
              </tbody>
            </table>
          </div>

          {/* Generate Report Modal */}
          {showGenerateReport && (
            <Modal title="Generate Board Report" onClose={() => setShowGenerateReport(false)}>
              <div className="space-y-4">
                <Select label="Report Type" value={reportForm.report_type}
                  onChange={v => setReportForm({ ...reportForm, report_type: v })}
                  options={[
                    { value: 'board_pack', label: 'Board Pack' }, { value: 'compliance_summary', label: 'Compliance Summary' },
                    { value: 'risk_dashboard', label: 'Risk Dashboard' }, { value: 'incident_report', label: 'Incident Report' },
                    { value: 'regulatory_update', label: 'Regulatory Update' }, { value: 'nis2_governance', label: 'NIS2 Governance' },
                    { value: 'quarterly_review', label: 'Quarterly Review' }, { value: 'annual_review', label: 'Annual Review' },
                    { value: 'custom', label: 'Custom Report' },
                  ]} />
                <Input label="Title" value={reportForm.title} onChange={v => setReportForm({ ...reportForm, title: v })} />
                <div className="grid grid-cols-2 gap-3">
                  <Input label="Period Start" type="date" value={reportForm.period_start}
                    onChange={v => setReportForm({ ...reportForm, period_start: v })} />
                  <Input label="Period End" type="date" value={reportForm.period_end}
                    onChange={v => setReportForm({ ...reportForm, period_end: v })} />
                </div>
                <Select label="Format" value={reportForm.file_format}
                  onChange={v => setReportForm({ ...reportForm, file_format: v })}
                  options={[
                    { value: 'pdf', label: 'PDF' }, { value: 'html', label: 'HTML' },
                    { value: 'docx', label: 'DOCX' }, { value: 'xlsx', label: 'Excel' },
                  ]} />
                <div className="flex justify-end space-x-3 pt-4">
                  <button onClick={() => setShowGenerateReport(false)} className="px-4 py-2 border rounded-lg text-sm">Cancel</button>
                  <button onClick={handleGenerateReport} className="px-4 py-2 bg-indigo-600 text-white rounded-lg text-sm hover:bg-indigo-700">Generate</button>
                </div>
              </div>
            </Modal>
          )}
        </div>
      )}
    </div>
  );
}

// ============================================================
// SUB-COMPONENTS
// ============================================================

function MetricCard({ label, value, color }: { label: string; value: string; color: string }) {
  const colorMap: Record<string, string> = {
    green: 'border-green-200 bg-green-50',
    yellow: 'border-yellow-200 bg-yellow-50',
    red: 'border-red-200 bg-red-50',
    blue: 'border-blue-200 bg-blue-50',
  };
  return (
    <div className={`rounded-lg border p-4 ${colorMap[color] || colorMap.blue}`}>
      <p className="text-xs font-medium text-gray-500 uppercase">{label}</p>
      <p className="text-xl font-bold text-gray-900 mt-1 capitalize">{value}</p>
    </div>
  );
}

function Modal({ title, onClose, children }: { title: string; onClose: () => void; children: React.ReactNode }) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <div className="bg-white rounded-lg shadow-xl w-full max-w-lg mx-4 p-6">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold text-gray-900">{title}</h3>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600 text-xl leading-none">&times;</button>
        </div>
        {children}
      </div>
    </div>
  );
}

function Input({ label, value, onChange, type = 'text', required = false }: {
  label: string; value: string; onChange: (v: string) => void; type?: string; required?: boolean;
}) {
  return (
    <div>
      <label className="block text-sm font-medium text-gray-700 mb-1">{label}{required && <span className="text-red-500"> *</span>}</label>
      <input type={type} value={value} onChange={e => onChange(e.target.value)}
        className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500" />
    </div>
  );
}

function Textarea({ label, value, onChange }: { label: string; value: string; onChange: (v: string) => void }) {
  return (
    <div>
      <label className="block text-sm font-medium text-gray-700 mb-1">{label}</label>
      <textarea value={value} onChange={e => onChange(e.target.value)} rows={3}
        className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500" />
    </div>
  );
}

function Select({ label, value, onChange, options }: {
  label: string; value: string; onChange: (v: string) => void;
  options: { value: string; label: string }[];
}) {
  return (
    <div>
      <label className="block text-sm font-medium text-gray-700 mb-1">{label}</label>
      <select value={value} onChange={e => onChange(e.target.value)}
        className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500">
        {options.map(o => <option key={o.value} value={o.value}>{o.label}</option>)}
      </select>
    </div>
  );
}
