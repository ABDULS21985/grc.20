'use client';

import { useEffect, useState } from 'react';
import api from '@/lib/api';

interface StepExecution {
  id: string;
  workflow_instance_id: string;
  workflow_step_id: string;
  step_order: number;
  status: string;
  assigned_to: string | null;
  delegated_to: string | null;
  action: string | null;
  comments: string;
  decision_reason: string;
  sla_deadline: string | null;
  sla_status: string;
  started_at: string | null;
  completed_at: string | null;
  created_at: string;
  step_name: string;
  step_type: string;
  entity_ref: string;
  entity_type: string;
  instance_status: string;
}

const STATUS_COLORS: Record<string, string> = {
  pending: 'badge-medium',
  in_progress: 'badge-info',
  approved: 'badge-low',
  rejected: 'badge-critical',
  completed: 'badge-low',
  skipped: 'badge-medium',
  escalated: 'badge-high',
  delegated: 'badge-info',
  timed_out: 'badge-critical',
};

const SLA_COLORS: Record<string, string> = {
  on_track: 'text-green-600',
  at_risk: 'text-amber-600',
  breached: 'text-red-600',
};

const STEP_TYPE_LABELS: Record<string, string> = {
  approval: 'Approval',
  review: 'Review',
  task: 'Task',
  notification: 'Notification',
  condition: 'Condition',
  parallel_gate: 'Parallel Gate',
  timer: 'Timer',
  auto_action: 'Auto Action',
};

const ENTITY_TYPE_LABELS: Record<string, string> = {
  policy: 'Policy',
  risk: 'Risk',
  exception: 'Exception',
  audit_finding: 'Audit Finding',
  vendor: 'Vendor',
  control: 'Control',
  incident: 'Incident',
};

export default function WorkflowApprovalsPage() {
  const [approvals, setApprovals] = useState<StepExecution[]>([]);
  const [loading, setLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState<string | null>(null);
  const [rejectModal, setRejectModal] = useState<StepExecution | null>(null);
  const [delegateModal, setDelegateModal] = useState<StepExecution | null>(null);
  const [rejectReason, setRejectReason] = useState('');
  const [delegateUserId, setDelegateUserId] = useState('');
  const [comments, setComments] = useState('');

  const loadApprovals = () => {
    setLoading(true);
    api.getMyApprovals()
      .then(res => setApprovals(res.data || []))
      .finally(() => setLoading(false));
  };

  useEffect(() => { loadApprovals(); }, []);

  const handleApprove = async (execId: string) => {
    setActionLoading(execId);
    try {
      await api.approveExecution(execId, { comments });
      setComments('');
      loadApprovals();
    } catch (err: any) {
      alert(err.message || 'Failed to approve');
    } finally {
      setActionLoading(null);
    }
  };

  const handleReject = async () => {
    if (!rejectModal) return;
    if (!rejectReason.trim()) {
      alert('Reason is required when rejecting');
      return;
    }
    setActionLoading(rejectModal.id);
    try {
      await api.rejectExecution(rejectModal.id, { comments, reason: rejectReason });
      setRejectModal(null);
      setRejectReason('');
      setComments('');
      loadApprovals();
    } catch (err: any) {
      alert(err.message || 'Failed to reject');
    } finally {
      setActionLoading(null);
    }
  };

  const handleDelegate = async () => {
    if (!delegateModal || !delegateUserId.trim()) return;
    setActionLoading(delegateModal.id);
    try {
      await api.delegateExecution(delegateModal.id, { delegate_user_id: delegateUserId });
      setDelegateModal(null);
      setDelegateUserId('');
      loadApprovals();
    } catch (err: any) {
      alert(err.message || 'Failed to delegate');
    } finally {
      setActionLoading(null);
    }
  };

  if (loading) {
    return (
      <div>
        <h1 className="text-2xl font-bold text-gray-900 mb-8">My Approvals</h1>
        <div className="grid grid-cols-1 gap-4 md:grid-cols-3 mb-8">
          {Array.from({ length: 3 }).map((_, i) => <div key={i} className="card animate-pulse h-24" />)}
        </div>
        <div className="card animate-pulse h-96" />
      </div>
    );
  }

  const pendingCount = approvals.filter(a => a.status === 'pending').length;
  const atRiskCount = approvals.filter(a => a.sla_status === 'at_risk').length;
  const breachedCount = approvals.filter(a => a.sla_status === 'breached').length;

  return (
    <div>
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">My Approvals</h1>
          <p className="text-gray-500 mt-1">Review, approve, reject, or delegate pending workflow items assigned to you</p>
        </div>
        <button onClick={loadApprovals} className="btn-secondary">Refresh</button>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-1 gap-4 md:grid-cols-3 mb-8">
        <div className="card">
          <p className="text-sm text-gray-500">Pending</p>
          <p className="text-3xl font-bold text-indigo-600 mt-1">{pendingCount}</p>
        </div>
        <div className="card">
          <p className="text-sm text-gray-500">At Risk (SLA)</p>
          <p className="text-3xl font-bold text-amber-600 mt-1">{atRiskCount}</p>
        </div>
        <div className="card">
          <p className="text-sm text-gray-500">SLA Breached</p>
          <p className="text-3xl font-bold text-red-600 mt-1">{breachedCount}</p>
        </div>
      </div>

      {/* Approvals Table */}
      <div className="card overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="table-header">
              <th className="px-4 py-3">Entity</th>
              <th className="px-4 py-3">Type</th>
              <th className="px-4 py-3">Step</th>
              <th className="px-4 py-3">Step Type</th>
              <th className="px-4 py-3">Status</th>
              <th className="px-4 py-3">SLA</th>
              <th className="px-4 py-3">Deadline</th>
              <th className="px-4 py-3">Actions</th>
            </tr>
          </thead>
          <tbody>
            {approvals.map(exec => (
              <tr key={exec.id} className="border-t border-gray-100 hover:bg-gray-50">
                <td className="px-4 py-3">
                  <span className="font-medium text-gray-900">{exec.entity_ref || '---'}</span>
                  <span className="block text-xs text-gray-500">{ENTITY_TYPE_LABELS[exec.entity_type] || exec.entity_type}</span>
                </td>
                <td className="px-4 py-3">
                  <span className="badge badge-info">{ENTITY_TYPE_LABELS[exec.entity_type] || exec.entity_type}</span>
                </td>
                <td className="px-4 py-3 text-gray-700">{exec.step_name}</td>
                <td className="px-4 py-3">
                  <span className="text-gray-600">{STEP_TYPE_LABELS[exec.step_type] || exec.step_type}</span>
                </td>
                <td className="px-4 py-3">
                  <span className={`badge ${STATUS_COLORS[exec.status] || 'badge-info'}`}>
                    {exec.status.replace(/_/g, ' ')}
                  </span>
                </td>
                <td className="px-4 py-3">
                  <span className={`font-medium ${SLA_COLORS[exec.sla_status] || ''}`}>
                    {exec.sla_status?.replace(/_/g, ' ') || 'N/A'}
                  </span>
                </td>
                <td className="px-4 py-3 text-gray-600">
                  {exec.sla_deadline ? new Date(exec.sla_deadline).toLocaleDateString('en-GB', {
                    day: '2-digit', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit',
                  }) : '---'}
                </td>
                <td className="px-4 py-3">
                  <div className="flex gap-2">
                    {(exec.step_type === 'approval' || exec.step_type === 'review' || exec.step_type === 'parallel_gate') && (
                      <button
                        onClick={() => handleApprove(exec.id)}
                        disabled={actionLoading === exec.id}
                        className="btn-sm bg-green-600 text-white hover:bg-green-700 disabled:opacity-50"
                      >
                        {actionLoading === exec.id ? '...' : 'Approve'}
                      </button>
                    )}
                    {(exec.step_type === 'task') && (
                      <button
                        onClick={() => handleApprove(exec.id)}
                        disabled={actionLoading === exec.id}
                        className="btn-sm bg-green-600 text-white hover:bg-green-700 disabled:opacity-50"
                      >
                        {actionLoading === exec.id ? '...' : 'Complete'}
                      </button>
                    )}
                    <button
                      onClick={() => setRejectModal(exec)}
                      className="btn-sm bg-red-600 text-white hover:bg-red-700"
                    >
                      Reject
                    </button>
                    <button
                      onClick={() => setDelegateModal(exec)}
                      className="btn-sm bg-gray-200 text-gray-700 hover:bg-gray-300"
                    >
                      Delegate
                    </button>
                  </div>
                </td>
              </tr>
            ))}
            {approvals.length === 0 && (
              <tr>
                <td colSpan={8} className="px-4 py-12 text-center text-gray-500">
                  No pending approvals. You are all caught up.
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      {/* Reject Modal */}
      {rejectModal && (
        <div className="fixed inset-0 bg-black/30 z-50 flex items-center justify-center p-4">
          <div className="bg-white rounded-lg shadow-xl max-w-md w-full p-6">
            <h2 className="text-lg font-bold text-gray-900 mb-4">Reject Step</h2>
            <p className="text-sm text-gray-600 mb-4">
              You are rejecting: <strong>{rejectModal.step_name}</strong> for{' '}
              <strong>{rejectModal.entity_ref}</strong>
            </p>
            <label className="block text-sm font-medium text-gray-700 mb-1">Reason (required)</label>
            <textarea
              value={rejectReason}
              onChange={e => setRejectReason(e.target.value)}
              rows={3}
              className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm mb-4 focus:ring-indigo-500 focus:border-indigo-500"
              placeholder="Explain why this is being rejected..."
            />
            <label className="block text-sm font-medium text-gray-700 mb-1">Comments (optional)</label>
            <textarea
              value={comments}
              onChange={e => setComments(e.target.value)}
              rows={2}
              className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm mb-4 focus:ring-indigo-500 focus:border-indigo-500"
              placeholder="Additional comments..."
            />
            <div className="flex justify-end gap-3">
              <button onClick={() => { setRejectModal(null); setRejectReason(''); }} className="btn-secondary">
                Cancel
              </button>
              <button
                onClick={handleReject}
                disabled={actionLoading === rejectModal.id}
                className="btn-sm bg-red-600 text-white hover:bg-red-700 disabled:opacity-50 px-4 py-2"
              >
                {actionLoading === rejectModal.id ? 'Rejecting...' : 'Reject'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Delegate Modal */}
      {delegateModal && (
        <div className="fixed inset-0 bg-black/30 z-50 flex items-center justify-center p-4">
          <div className="bg-white rounded-lg shadow-xl max-w-md w-full p-6">
            <h2 className="text-lg font-bold text-gray-900 mb-4">Delegate Step</h2>
            <p className="text-sm text-gray-600 mb-4">
              Delegate <strong>{delegateModal.step_name}</strong> for{' '}
              <strong>{delegateModal.entity_ref}</strong> to another user.
            </p>
            <label className="block text-sm font-medium text-gray-700 mb-1">Delegate User ID</label>
            <input
              type="text"
              value={delegateUserId}
              onChange={e => setDelegateUserId(e.target.value)}
              className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm mb-4 focus:ring-indigo-500 focus:border-indigo-500"
              placeholder="Enter user ID to delegate to..."
            />
            <div className="flex justify-end gap-3">
              <button onClick={() => { setDelegateModal(null); setDelegateUserId(''); }} className="btn-secondary">
                Cancel
              </button>
              <button
                onClick={handleDelegate}
                disabled={actionLoading === delegateModal.id || !delegateUserId.trim()}
                className="btn-primary disabled:opacity-50"
              >
                {actionLoading === delegateModal.id ? 'Delegating...' : 'Delegate'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
