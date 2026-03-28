'use client';

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import api from '@/lib/api';
import type { DSRRequest, DSRTask, DSRAuditEntry } from '@/types';

const TYPE_LABELS: Record<string, string> = {
  access: 'Right of Access (Art.15)',
  erasure: 'Right to Erasure (Art.17)',
  rectification: 'Right to Rectification (Art.16)',
  portability: 'Right to Portability (Art.20)',
  restriction: 'Right to Restriction (Art.18)',
  objection: 'Right to Object (Art.21)',
  automated_decision: 'Automated Decision Making (Art.22)',
};

export default function DSRDetailPage() {
  const { id } = useParams<{ id: string }>();
  const [request, setRequest] = useState<DSRRequest | null>(null);
  const [tasks, setTasks] = useState<DSRTask[]>([]);
  const [auditTrail, setAuditTrail] = useState<DSRAuditEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [tab, setTab] = useState<'tasks' | 'audit'>('tasks');

  useEffect(() => {
    api.getDSRRequest(id)
      .then(res => {
        const data = res.data;
        setRequest(data?.request || data);
        setTasks(data?.tasks || []);
        setAuditTrail(data?.audit_trail || []);
      })
      .finally(() => setLoading(false));
  }, [id]);

  async function completeTask(taskId: string) {
    try {
      await api.updateDSRTask(id, taskId, { status: 'completed' });
      setTasks(prev => prev.map(t => t.id === taskId ? { ...t, status: 'completed', completed_at: new Date().toISOString() } : t));
    } catch { /* ignore */ }
  }

  if (loading) return <div className="card animate-pulse h-96" />;
  if (!request) return <div className="card p-8 text-center text-gray-500">DSR request not found</div>;

  const isOverdue = request.days_remaining <= 0;
  const isAtRisk = request.days_remaining > 0 && request.days_remaining <= 7;

  return (
    <div>
      <div className="flex items-center gap-3 mb-2">
        <a href="/dsr" className="text-gray-400 hover:text-gray-600">← Back to DSR</a>
      </div>

      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">{request.request_ref}</h1>
          <p className="text-gray-500 mt-1">{TYPE_LABELS[request.request_type] || request.request_type}</p>
        </div>
        <div className="flex gap-2">
          {request.status === 'received' && (
            <button onClick={() => api.verifyDSRIdentity(id, { method: 'document' })} className="btn-secondary">Verify Identity</button>
          )}
          {request.status === 'in_progress' && (
            <>
              <button onClick={() => api.extendDSRDeadline(id, { reason: 'Complex request' })} className="btn-secondary">Extend Deadline</button>
              <button onClick={() => api.completeDSR(id, { response_method: 'email' })} className="btn-primary">Complete Request</button>
            </>
          )}
          {!['completed', 'rejected', 'withdrawn'].includes(request.status) && (
            <button onClick={() => api.rejectDSR(id, { reason: '', legal_basis: 'Article 17(3)' })} className="btn-danger">Reject</button>
          )}
        </div>
      </div>

      {/* Status Banner */}
      {(isOverdue || isAtRisk) && (
        <div className={`rounded-lg p-4 mb-6 ${isOverdue ? 'bg-red-50 border border-red-200' : 'bg-amber-50 border border-amber-200'}`}>
          <p className={`font-semibold ${isOverdue ? 'text-red-700' : 'text-amber-700'}`}>
            {isOverdue ? 'OVERDUE — Response deadline has passed!' : `${request.days_remaining} days remaining until deadline`}
          </p>
          <p className={`text-sm mt-1 ${isOverdue ? 'text-red-600' : 'text-amber-600'}`}>
            Deadline: {new Date(request.extended_deadline || request.response_deadline).toLocaleDateString('en-GB', { day: 'numeric', month: 'long', year: 'numeric' })}
            {request.was_extended && ' (Extended)'}
          </p>
        </div>
      )}

      {/* Info Grid */}
      <div className="grid grid-cols-1 gap-6 md:grid-cols-2 mb-6">
        <div className="card">
          <h2 className="font-semibold text-gray-900 mb-4">Request Details</h2>
          <dl className="space-y-3 text-sm">
            <div className="flex justify-between"><dt className="text-gray-500">Status</dt><dd className="font-medium">{request.status.replace(/_/g, ' ')}</dd></div>
            <div className="flex justify-between"><dt className="text-gray-500">Priority</dt><dd className="font-medium">{request.priority}</dd></div>
            <div className="flex justify-between"><dt className="text-gray-500">Source</dt><dd>{request.request_source}</dd></div>
            <div className="flex justify-between"><dt className="text-gray-500">Received</dt><dd>{new Date(request.received_date).toLocaleDateString('en-GB')}</dd></div>
            <div className="flex justify-between"><dt className="text-gray-500">Deadline</dt><dd className="font-medium">{new Date(request.extended_deadline || request.response_deadline).toLocaleDateString('en-GB')}</dd></div>
            <div className="flex justify-between"><dt className="text-gray-500">ID Verified</dt><dd>{request.data_subject_id_verified ? 'Yes' : 'No'}</dd></div>
          </dl>
        </div>
        <div className="card">
          <h2 className="font-semibold text-gray-900 mb-4">Data Subject</h2>
          <dl className="space-y-3 text-sm">
            <div className="flex justify-between"><dt className="text-gray-500">Name</dt><dd className="font-medium">{request.data_subject_name || '—'}</dd></div>
            <div className="flex justify-between"><dt className="text-gray-500">Email</dt><dd>{request.data_subject_email || '—'}</dd></div>
          </dl>
          <div className="mt-4">
            <p className="text-sm text-gray-500 mb-1">Description</p>
            <p className="text-sm text-gray-700">{request.request_description}</p>
          </div>
          {request.data_systems_affected?.length > 0 && (
            <div className="mt-4">
              <p className="text-sm text-gray-500 mb-1">Systems Affected</p>
              <div className="flex flex-wrap gap-1">
                {request.data_systems_affected.map(s => <span key={s} className="badge badge-info">{s}</span>)}
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Tabs */}
      <div className="flex gap-1 border-b border-gray-200 mb-6">
        {(['tasks', 'audit'] as const).map(t => (
          <button key={t} onClick={() => setTab(t)}
            className={`px-4 py-2.5 text-sm font-medium border-b-2 transition-colors ${tab === t ? 'border-indigo-600 text-indigo-600' : 'border-transparent text-gray-500 hover:text-gray-700'}`}
          >
            {t === 'tasks' ? `Task Checklist (${tasks.filter(t => t.status === 'completed').length}/${tasks.length})` : 'Audit Trail'}
          </button>
        ))}
      </div>

      {tab === 'tasks' ? (
        <div className="card">
          <div className="space-y-3">
            {tasks.sort((a, b) => a.sort_order - b.sort_order).map(task => (
              <div key={task.id} className="flex items-center gap-3 p-3 rounded-lg border border-gray-100 hover:bg-gray-50">
                <button
                  onClick={() => task.status !== 'completed' && completeTask(task.id)}
                  className={`h-5 w-5 rounded border flex-shrink-0 flex items-center justify-center ${task.status === 'completed' ? 'bg-green-500 border-green-500 text-white' : 'border-gray-300 hover:border-indigo-500'}`}
                >
                  {task.status === 'completed' && <svg className="h-3 w-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M5 13l4 4L19 7" /></svg>}
                </button>
                <div className="flex-1">
                  <p className={`text-sm font-medium ${task.status === 'completed' ? 'text-gray-400 line-through' : 'text-gray-900'}`}>{task.description}</p>
                  {task.system_name && <p className="text-xs text-gray-500">System: {task.system_name}</p>}
                </div>
                <span className={`badge ${task.status === 'completed' ? 'badge-low' : task.status === 'blocked' ? 'badge-critical' : task.status === 'in_progress' ? 'badge-info' : 'badge-medium'}`}>
                  {task.status}
                </span>
              </div>
            ))}
            {tasks.length === 0 && <p className="text-center text-gray-500 py-8">No tasks</p>}
          </div>
        </div>
      ) : (
        <div className="card">
          <div className="space-y-4">
            {auditTrail.map(entry => (
              <div key={entry.id} className="flex gap-4">
                <div className="flex flex-col items-center">
                  <div className="h-2.5 w-2.5 rounded-full bg-indigo-500 mt-1.5" />
                  <div className="flex-1 w-px bg-gray-200" />
                </div>
                <div className="pb-4">
                  <p className="text-sm font-medium text-gray-900">{entry.action.replace(/_/g, ' ')}</p>
                  <p className="text-sm text-gray-600 mt-0.5">{entry.description}</p>
                  <p className="text-xs text-gray-400 mt-1">
                    {new Date(entry.created_at).toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit' })}
                  </p>
                </div>
              </div>
            ))}
            {auditTrail.length === 0 && <p className="text-center text-gray-500 py-8">No audit trail entries</p>}
          </div>
        </div>
      )}
    </div>
  );
}
