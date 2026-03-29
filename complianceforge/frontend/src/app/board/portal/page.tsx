'use client';

import { useEffect, useState, useCallback } from 'react';

// ============================================================
// TYPES
// ============================================================

interface BoardDashboard {
  compliance_score: number;
  compliance_by_framework: { framework_code: string; framework_name: string; score: number }[];
  risk_appetite_status: string;
  open_risks: { total: number; critical: number; high: number; medium: number; low: number };
  open_incidents: { total: number; critical: number; high: number; open: number };
  upcoming_decisions: BoardDecision[];
  recent_decisions: BoardDecision[];
  regulatory_horizon: { title: string; severity: string; deadline: string; days_left: number }[];
  next_meeting: BoardMeeting | null;
  pending_actions: number;
  overdue_actions: number;
  total_members: number;
  reports_generated: number;
  last_board_pack_at: string | null;
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
  board_pack_document_path: string | null;
  board_pack_generated_at: string | null;
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
}

// ============================================================
// PORTAL API — uses separate base URL for public endpoints
// ============================================================

const PORTAL_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

async function portalFetch<T>(path: string): Promise<T> {
  const response = await fetch(`${PORTAL_BASE}${path}`);
  const data = await response.json();
  if (!response.ok) {
    throw new Error(data.error?.message || 'Request failed');
  }
  return data;
}

// ============================================================
// COLOUR HELPERS
// ============================================================

function scoreColor(score: number): string {
  if (score >= 80) return 'text-green-600';
  if (score >= 60) return 'text-yellow-600';
  return 'text-red-600';
}

function scoreBg(score: number): string {
  if (score >= 80) return 'bg-green-500';
  if (score >= 60) return 'bg-yellow-500';
  return 'bg-red-500';
}

function appetiteLabel(status: string): { text: string; color: string } {
  switch (status) {
    case 'within_appetite': return { text: 'Within Appetite', color: 'text-green-700 bg-green-100' };
    case 'at_limit': return { text: 'At Limit', color: 'text-yellow-700 bg-yellow-100' };
    case 'exceeded': return { text: 'Exceeded', color: 'text-red-700 bg-red-100' };
    default: return { text: status, color: 'text-gray-700 bg-gray-100' };
  }
}

const SEVERITY_COLORS: Record<string, string> = {
  critical: 'bg-red-100 text-red-800',
  high: 'bg-orange-100 text-orange-800',
  medium: 'bg-yellow-100 text-yellow-800',
  low: 'bg-blue-100 text-blue-800',
};

// ============================================================
// MAIN PORTAL COMPONENT
// ============================================================

export default function BoardPortalPage() {
  const [token, setToken] = useState<string>('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [activeView, setActiveView] = useState<'dashboard' | 'meetings' | 'decisions'>('dashboard');

  const [dashboard, setDashboard] = useState<BoardDashboard | null>(null);
  const [meetings, setMeetings] = useState<BoardMeeting[]>([]);
  const [decisions, setDecisions] = useState<BoardDecision[]>([]);

  // Extract token from URL hash or query param on load
  useEffect(() => {
    if (typeof window !== 'undefined') {
      const params = new URLSearchParams(window.location.search);
      const t = params.get('token') || window.location.hash.replace('#', '');
      if (t) {
        setToken(t);
      }
    }
  }, []);

  const loadPortalData = useCallback(async () => {
    if (!token) return;
    setLoading(true);
    setError(null);
    try {
      const [dashRes, meetRes, decRes] = await Promise.all([
        portalFetch<any>(`/board-portal/${token}`),
        portalFetch<any>(`/board-portal/${token}/meetings`),
        portalFetch<any>(`/board-portal/${token}/decisions`),
      ]);
      if (dashRes.data) setDashboard(dashRes.data);
      if (meetRes.data) setMeetings(meetRes.data);
      if (decRes.data) setDecisions(decRes.data);
    } catch (err: any) {
      setError(err.message || 'Failed to load portal data. Please check your access link.');
    } finally {
      setLoading(false);
    }
  }, [token]);

  useEffect(() => {
    if (token) loadPortalData();
  }, [token, loadPortalData]);

  // ── Token Entry Screen ──
  if (!token || (!dashboard && !loading && !error)) {
    return (
      <div className="min-h-screen bg-slate-50 flex items-center justify-center">
        <div className="bg-white rounded-xl shadow-lg p-10 max-w-md w-full text-center">
          <div className="mb-6">
            <div className="w-16 h-16 bg-indigo-100 rounded-full flex items-center justify-center mx-auto mb-4">
              <svg className="w-8 h-8 text-indigo-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
              </svg>
            </div>
            <h1 className="text-2xl font-bold text-gray-900">Board Portal</h1>
            <p className="text-gray-500 mt-2">ComplianceForge Executive Board Portal</p>
          </div>
          <div className="mb-4">
            <input
              type="text"
              placeholder="Enter your portal access token"
              value={token}
              onChange={e => setToken(e.target.value)}
              className="w-full border border-gray-300 rounded-lg px-4 py-3 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
            />
          </div>
          <button
            onClick={loadPortalData}
            disabled={!token}
            className="w-full bg-indigo-600 text-white px-4 py-3 rounded-lg text-sm font-medium hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            Access Portal
          </button>
          {error && (
            <p className="mt-4 text-sm text-red-600">{error}</p>
          )}
          <p className="mt-6 text-xs text-gray-400">
            Your access token was provided by the board secretary. If you do not have a token, contact your organisation administrator.
          </p>
        </div>
      </div>
    );
  }

  if (loading) {
    return (
      <div className="min-h-screen bg-slate-50 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-10 w-10 border-b-2 border-indigo-600 mx-auto" />
          <p className="mt-4 text-gray-500">Loading executive dashboard...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen bg-slate-50 flex items-center justify-center">
        <div className="bg-white rounded-xl shadow-lg p-10 max-w-md text-center">
          <p className="text-red-600 font-medium">{error}</p>
          <button onClick={() => { setToken(''); setError(null); }}
            className="mt-4 text-indigo-600 underline text-sm">
            Try a different token
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-slate-50">
      {/* Header */}
      <header className="bg-white border-b border-gray-200 shadow-sm">
        <div className="max-w-7xl mx-auto px-6 py-4 flex items-center justify-between">
          <div>
            <h1 className="text-xl font-bold text-gray-900">ComplianceForge</h1>
            <p className="text-sm text-gray-500">Executive Board Portal</p>
          </div>
          <nav className="flex space-x-1">
            {(['dashboard', 'meetings', 'decisions'] as const).map(view => (
              <button
                key={view}
                onClick={() => setActiveView(view)}
                className={`px-4 py-2 rounded-lg text-sm font-medium capitalize transition-colors ${
                  activeView === view
                    ? 'bg-indigo-100 text-indigo-700'
                    : 'text-gray-500 hover:text-gray-700 hover:bg-gray-100'
                }`}
              >
                {view}
              </button>
            ))}
          </nav>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-6 py-8">
        {/* ── DASHBOARD VIEW ── */}
        {activeView === 'dashboard' && dashboard && (
          <div className="space-y-8">
            {/* Compliance Gauge */}
            <div className="bg-white rounded-xl shadow p-8">
              <h2 className="text-lg font-semibold text-gray-800 mb-6">Compliance Posture</h2>
              <div className="flex items-center space-x-8">
                <div className="text-center">
                  <div className="relative w-32 h-32">
                    <svg className="w-32 h-32 transform -rotate-90" viewBox="0 0 120 120">
                      <circle cx="60" cy="60" r="50" fill="none" stroke="#e5e7eb" strokeWidth="10" />
                      <circle cx="60" cy="60" r="50" fill="none" stroke="currentColor"
                        strokeWidth="10" strokeLinecap="round" strokeDasharray={`${dashboard.compliance_score * 3.14} 314`}
                        className={scoreColor(dashboard.compliance_score)} />
                    </svg>
                    <div className="absolute inset-0 flex items-center justify-center">
                      <span className={`text-2xl font-bold ${scoreColor(dashboard.compliance_score)}`}>
                        {dashboard.compliance_score.toFixed(0)}%
                      </span>
                    </div>
                  </div>
                  <p className="text-sm text-gray-500 mt-2">Overall Score</p>
                </div>
                <div className="flex-1 space-y-3">
                  {dashboard.compliance_by_framework.map(fw => (
                    <div key={fw.framework_code}>
                      <div className="flex justify-between text-sm mb-1">
                        <span className="text-gray-700">{fw.framework_name}</span>
                        <span className={`font-medium ${scoreColor(fw.score)}`}>{fw.score.toFixed(1)}%</span>
                      </div>
                      <div className="w-full bg-gray-200 rounded-full h-2">
                        <div className={`h-2 rounded-full ${scoreBg(fw.score)}`} style={{ width: `${Math.min(fw.score, 100)}%` }} />
                      </div>
                    </div>
                  ))}
                  {dashboard.compliance_by_framework.length === 0 && (
                    <p className="text-sm text-gray-400">No framework data available</p>
                  )}
                </div>
              </div>
            </div>

            {/* Risk & Incidents */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div className="bg-white rounded-xl shadow p-6">
                <h3 className="text-lg font-semibold text-gray-800 mb-4">Risk Status</h3>
                <div className="flex items-center justify-between mb-4">
                  <span className="text-sm text-gray-500">Risk Appetite</span>
                  {(() => {
                    const a = appetiteLabel(dashboard.risk_appetite_status);
                    return <span className={`px-3 py-1 rounded-full text-xs font-medium ${a.color}`}>{a.text}</span>;
                  })()}
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div className="text-center p-3 bg-gray-50 rounded-lg">
                    <p className="text-2xl font-bold text-gray-900">{dashboard.open_risks.total}</p>
                    <p className="text-xs text-gray-500">Total Open</p>
                  </div>
                  <div className="text-center p-3 bg-red-50 rounded-lg">
                    <p className="text-2xl font-bold text-red-600">{dashboard.open_risks.critical}</p>
                    <p className="text-xs text-gray-500">Critical</p>
                  </div>
                  <div className="text-center p-3 bg-orange-50 rounded-lg">
                    <p className="text-2xl font-bold text-orange-600">{dashboard.open_risks.high}</p>
                    <p className="text-xs text-gray-500">High</p>
                  </div>
                  <div className="text-center p-3 bg-yellow-50 rounded-lg">
                    <p className="text-2xl font-bold text-yellow-600">{dashboard.open_risks.medium}</p>
                    <p className="text-xs text-gray-500">Medium</p>
                  </div>
                </div>
              </div>

              <div className="bg-white rounded-xl shadow p-6">
                <h3 className="text-lg font-semibold text-gray-800 mb-4">Incident Summary</h3>
                <div className="grid grid-cols-2 gap-4">
                  <div className="text-center p-3 bg-gray-50 rounded-lg">
                    <p className="text-2xl font-bold text-gray-900">{dashboard.open_incidents.total}</p>
                    <p className="text-xs text-gray-500">Last Quarter</p>
                  </div>
                  <div className="text-center p-3 bg-blue-50 rounded-lg">
                    <p className="text-2xl font-bold text-blue-600">{dashboard.open_incidents.open}</p>
                    <p className="text-xs text-gray-500">Currently Open</p>
                  </div>
                  <div className="text-center p-3 bg-red-50 rounded-lg">
                    <p className="text-2xl font-bold text-red-600">{dashboard.open_incidents.critical}</p>
                    <p className="text-xs text-gray-500">Critical</p>
                  </div>
                  <div className="text-center p-3 bg-orange-50 rounded-lg">
                    <p className="text-2xl font-bold text-orange-600">{dashboard.open_incidents.high}</p>
                    <p className="text-xs text-gray-500">High</p>
                  </div>
                </div>
              </div>
            </div>

            {/* Actions & Regulatory */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div className="bg-white rounded-xl shadow p-6">
                <h3 className="text-lg font-semibold text-gray-800 mb-4">Pending Board Actions</h3>
                <div className="flex items-center space-x-6">
                  <div className="text-center">
                    <p className="text-3xl font-bold text-yellow-600">{dashboard.pending_actions}</p>
                    <p className="text-xs text-gray-500">Pending</p>
                  </div>
                  <div className="text-center">
                    <p className="text-3xl font-bold text-red-600">{dashboard.overdue_actions}</p>
                    <p className="text-xs text-gray-500">Overdue</p>
                  </div>
                </div>
                {dashboard.upcoming_decisions.length > 0 && (
                  <div className="mt-4 space-y-2">
                    {dashboard.upcoming_decisions.slice(0, 3).map(d => (
                      <div key={d.id} className="flex items-center justify-between py-2 border-t border-gray-100">
                        <div>
                          <p className="text-sm font-medium text-gray-800">{d.title}</p>
                          <p className="text-xs text-gray-400">{d.decision_ref} &middot; Due: {d.action_due_date || 'TBD'}</p>
                        </div>
                        <span className={`px-2 py-1 rounded text-xs font-medium ${
                          d.action_status === 'overdue' ? 'bg-red-100 text-red-700' : 'bg-yellow-100 text-yellow-700'
                        }`}>
                          {d.action_status}
                        </span>
                      </div>
                    ))}
                  </div>
                )}
              </div>

              <div className="bg-white rounded-xl shadow p-6">
                <h3 className="text-lg font-semibold text-gray-800 mb-4">Regulatory Horizon</h3>
                {dashboard.regulatory_horizon.length > 0 ? (
                  <div className="space-y-3">
                    {dashboard.regulatory_horizon.map((item, idx) => (
                      <div key={idx} className="flex items-center justify-between py-2 border-b border-gray-100 last:border-0">
                        <div className="flex-1 min-w-0">
                          <p className="text-sm font-medium text-gray-800 truncate">{item.title}</p>
                          <p className="text-xs text-gray-400">Deadline: {item.deadline}</p>
                        </div>
                        <div className="flex items-center space-x-2 ml-3">
                          <span className={`px-2 py-1 rounded text-xs font-medium ${SEVERITY_COLORS[item.severity] || 'bg-gray-100'}`}>
                            {item.severity}
                          </span>
                          <span className="text-xs text-gray-500 whitespace-nowrap">{item.days_left}d</span>
                        </div>
                      </div>
                    ))}
                  </div>
                ) : (
                  <p className="text-sm text-gray-400">No upcoming regulatory deadlines</p>
                )}
              </div>
            </div>

            {/* Next Meeting */}
            {dashboard.next_meeting && (
              <div className="bg-white rounded-xl shadow p-6">
                <h3 className="text-lg font-semibold text-gray-800 mb-3">Next Board Meeting</h3>
                <div className="flex items-center space-x-6">
                  <div className="bg-indigo-50 rounded-lg p-4 text-center min-w-[80px]">
                    <p className="text-xs text-indigo-600 uppercase font-medium">
                      {new Date(dashboard.next_meeting.date).toLocaleDateString('en-GB', { month: 'short' })}
                    </p>
                    <p className="text-2xl font-bold text-indigo-700">
                      {new Date(dashboard.next_meeting.date).getDate()}
                    </p>
                  </div>
                  <div>
                    <p className="font-semibold text-gray-900">{dashboard.next_meeting.title}</p>
                    <p className="text-sm text-gray-500">
                      {dashboard.next_meeting.meeting_ref} &middot;
                      {dashboard.next_meeting.time || 'Time TBC'} &middot;
                      {dashboard.next_meeting.location || 'Location TBC'}
                    </p>
                    {dashboard.next_meeting.board_pack_document_path ? (
                      <span className="inline-block mt-2 px-3 py-1 bg-green-100 text-green-700 rounded text-xs font-medium">
                        Board Pack Available
                      </span>
                    ) : (
                      <span className="inline-block mt-2 px-3 py-1 bg-gray-100 text-gray-500 rounded text-xs">
                        Board Pack Pending
                      </span>
                    )}
                  </div>
                </div>
              </div>
            )}
          </div>
        )}

        {/* ── MEETINGS VIEW ── */}
        {activeView === 'meetings' && (
          <div className="space-y-4">
            <h2 className="text-xl font-semibold text-gray-800 mb-4">Board Meetings</h2>
            {meetings.map(m => (
              <div key={m.id} className="bg-white rounded-xl shadow p-6">
                <div className="flex items-center justify-between">
                  <div>
                    <div className="flex items-center space-x-3 mb-1">
                      <span className="text-xs font-mono text-gray-400">{m.meeting_ref}</span>
                      <span className={`px-2 py-0.5 rounded text-xs font-medium ${
                        m.status === 'completed' ? 'bg-green-100 text-green-700' :
                        m.status === 'planned' ? 'bg-blue-100 text-blue-700' :
                        'bg-gray-100 text-gray-600'
                      }`}>
                        {m.status.replace(/_/g, ' ')}
                      </span>
                    </div>
                    <p className="text-base font-semibold text-gray-900">{m.title}</p>
                    <p className="text-sm text-gray-500 mt-1">
                      {m.date}{m.time ? ` at ${m.time}` : ''} &middot; {m.meeting_type.replace(/_/g, ' ')}
                      {m.location ? ` &middot; ${m.location}` : ''}
                    </p>
                  </div>
                  {m.board_pack_document_path && (
                    <button className="px-4 py-2 bg-indigo-600 text-white rounded-lg text-sm hover:bg-indigo-700">
                      Download Pack
                    </button>
                  )}
                </div>
              </div>
            ))}
            {meetings.length === 0 && (
              <div className="bg-white rounded-xl shadow p-8 text-center text-gray-400">
                No meetings available
              </div>
            )}
          </div>
        )}

        {/* ── DECISIONS VIEW ── */}
        {activeView === 'decisions' && (
          <div className="space-y-4">
            <h2 className="text-xl font-semibold text-gray-800 mb-4">Board Decisions</h2>
            {decisions.map(d => (
              <div key={d.id} className="bg-white rounded-xl shadow p-6">
                <div className="flex items-center space-x-3 mb-2">
                  <span className="text-xs font-mono text-gray-400">{d.decision_ref}</span>
                  <span className={`px-2 py-0.5 rounded text-xs font-medium ${
                    d.decision === 'approved' ? 'bg-green-100 text-green-700' :
                    d.decision === 'rejected' ? 'bg-red-100 text-red-700' :
                    d.decision === 'deferred' ? 'bg-yellow-100 text-yellow-700' :
                    'bg-orange-100 text-orange-700'
                  }`}>
                    {d.decision.replace(/_/g, ' ')}
                  </span>
                  <span className="text-xs text-gray-400 capitalize">{d.decision_type.replace(/_/g, ' ')}</span>
                </div>
                <p className="text-base font-semibold text-gray-900">{d.title}</p>
                {d.description && <p className="text-sm text-gray-600 mt-1">{d.description}</p>}

                <div className="flex items-center space-x-4 mt-3 text-xs text-gray-500">
                  <span>For: {d.vote_for}</span>
                  <span>Against: {d.vote_against}</span>
                  <span>Abstain: {d.vote_abstain}</span>
                  {d.decided_at && <span>&middot; {new Date(d.decided_at).toLocaleDateString()}</span>}
                  {d.decided_by && <span>by {d.decided_by}</span>}
                </div>

                {d.action_required && (
                  <div className="mt-3 pt-3 border-t border-gray-100 flex items-center justify-between">
                    <div>
                      <span className="text-xs font-semibold text-gray-500 uppercase">Action: </span>
                      <span className="text-sm text-gray-700">{d.action_description}</span>
                      {d.action_due_date && <span className="text-xs text-gray-400 ml-2">Due: {d.action_due_date}</span>}
                    </div>
                    <span className={`px-2 py-1 rounded text-xs font-medium ${
                      d.action_status === 'completed' ? 'bg-green-100 text-green-700' :
                      d.action_status === 'overdue' ? 'bg-red-100 text-red-700' :
                      'bg-yellow-100 text-yellow-700'
                    }`}>
                      {d.action_status.replace(/_/g, ' ')}
                    </span>
                  </div>
                )}
              </div>
            ))}
            {decisions.length === 0 && (
              <div className="bg-white rounded-xl shadow p-8 text-center text-gray-400">
                No decisions available
              </div>
            )}
          </div>
        )}
      </main>

      {/* Footer */}
      <footer className="border-t border-gray-200 mt-12">
        <div className="max-w-7xl mx-auto px-6 py-4 text-center text-xs text-gray-400">
          ComplianceForge Board Portal &middot; Confidential &middot; Classification: Board Confidential
        </div>
      </footer>
    </div>
  );
}
