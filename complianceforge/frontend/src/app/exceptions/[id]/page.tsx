'use client';

import { useEffect, useState, useCallback } from 'react';
import { useParams } from 'next/navigation';
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
  scope_description: string;
  control_implementation_ids: string[];
  framework_control_codes: string[];
  policy_id: string | null;
  risk_justification: string;
  residual_risk_description: string;
  residual_risk_level: string;
  risk_assessment_id: string | null;
  risk_accepted_by: string | null;
  risk_accepted_at: string | null;
  has_compensating_controls: boolean;
  compensating_controls_description: string;
  compensating_control_ids: string[];
  compensating_effectiveness: string;
  requested_by: string;
  requested_at: string;
  approved_by: string | null;
  approved_at: string | null;
  approval_comments: string;
  rejection_reason: string;
  workflow_instance_id: string | null;
  effective_date: string;
  expiry_date: string | null;
  review_frequency_months: number;
  next_review_date: string | null;
  last_review_date: string | null;
  last_reviewed_by: string | null;
  renewal_count: number;
  conditions: string;
  business_impact_if_implemented: string;
  regulatory_notification_required: boolean;
  regulator_notified_at: string | null;
  audit_evidence_path: string;
  tags: string[];
  metadata: Record<string, any>;
  created_at: string;
  updated_at: string;
}

interface ExceptionReview {
  id: string;
  review_type: string;
  reviewer_id: string;
  review_date: string;
  outcome: string;
  risk_level_at_review: string;
  compensating_effective: boolean | null;
  findings: string;
  recommendations: string;
  next_review_date: string | null;
  created_at: string;
}

interface AuditEntry {
  id: string;
  action: string;
  actor_id: string;
  actor_email: string;
  old_status: string;
  new_status: string;
  details: string;
  created_at: string;
}

interface ComplianceImpact {
  exception_id: string;
  exception_ref: string;
  affected_control_count: number;
  affected_frameworks: string[];
  compliance_score_impact: number;
  risk_exposure_increase: string;
  compensating_coverage: number;
  net_risk_delta: string;
  affected_controls: { control_code: string; control_name: string; framework: string; impact: string }[];
  recommendations: string[];
}

// ============================================================
// CONSTANTS
// ============================================================

const STATUS_COLORS: Record<string, string> = {
  draft: 'bg-gray-100 text-gray-700 border-gray-300',
  pending_risk_assessment: 'bg-yellow-100 text-yellow-800 border-yellow-300',
  pending_approval: 'bg-blue-100 text-blue-800 border-blue-300',
  approved: 'bg-green-100 text-green-800 border-green-300',
  rejected: 'bg-red-100 text-red-800 border-red-300',
  expired: 'bg-orange-100 text-orange-800 border-orange-300',
  revoked: 'bg-red-200 text-red-900 border-red-400',
  renewal_pending: 'bg-amber-100 text-amber-800 border-amber-300',
};

const RISK_COLORS: Record<string, string> = {
  critical: 'bg-red-100 text-red-800 border-red-300',
  high: 'bg-orange-100 text-orange-800 border-orange-300',
  medium: 'bg-yellow-100 text-yellow-800 border-yellow-300',
  low: 'bg-blue-100 text-blue-800 border-blue-300',
  very_low: 'bg-gray-100 text-gray-600 border-gray-300',
};

const EFFECTIVENESS_LABELS: Record<string, string> = {
  fully_effective: 'Fully Effective',
  mostly_effective: 'Mostly Effective',
  partially_effective: 'Partially Effective',
  minimally_effective: 'Minimally Effective',
  not_effective: 'Not Effective',
  not_assessed: 'Not Assessed',
};

const EFFECTIVENESS_COLORS: Record<string, string> = {
  fully_effective: 'text-green-700',
  mostly_effective: 'text-green-600',
  partially_effective: 'text-yellow-600',
  minimally_effective: 'text-orange-600',
  not_effective: 'text-red-600',
  not_assessed: 'text-gray-500',
};

const OUTCOME_COLORS: Record<string, string> = {
  continue: 'bg-green-100 text-green-800',
  modify: 'bg-yellow-100 text-yellow-800',
  escalate: 'bg-orange-100 text-orange-800',
  revoke: 'bg-red-100 text-red-800',
  renew: 'bg-blue-100 text-blue-800',
};

const AUDIT_ACTION_ICONS: Record<string, string> = {
  created: '[ + ]',
  updated: '[ ~ ]',
  submitted_for_approval: '[ >> ]',
  approved: '[ OK ]',
  rejected: '[ X ]',
  revoked: '[ !! ]',
  renewed: '[ <> ]',
  reviewed: '[ ? ]',
};

// ============================================================
// COMPONENT
// ============================================================

export default function ExceptionDetailPage() {
  const params = useParams();
  const id = params.id as string;

  const [exception, setException] = useState<ComplianceException | null>(null);
  const [reviews, setReviews] = useState<ExceptionReview[]>([]);
  const [auditTrail, setAuditTrail] = useState<AuditEntry[]>([]);
  const [impact, setImpact] = useState<ComplianceImpact | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [actionLoading, setActionLoading] = useState(false);
  const [actionMessage, setActionMessage] = useState('');
  const [activeTab, setActiveTab] = useState<'details' | 'impact' | 'reviews' | 'audit'>('details');

  // Action modals
  const [showApproveModal, setShowApproveModal] = useState(false);
  const [showRejectModal, setShowRejectModal] = useState(false);
  const [showRevokeModal, setShowRevokeModal] = useState(false);
  const [showRenewModal, setShowRenewModal] = useState(false);
  const [showReviewModal, setShowReviewModal] = useState(false);
  const [approveComments, setApproveComments] = useState('');
  const [rejectReason, setRejectReason] = useState('');
  const [revokeReason, setRevokeReason] = useState('');
  const [renewExpiry, setRenewExpiry] = useState('');
  const [renewJustification, setRenewJustification] = useState('');
  const [reviewForm, setReviewForm] = useState({
    review_type: 'periodic',
    outcome: 'continue',
    risk_level_at_review: '',
    findings: '',
    recommendations: '',
    next_review_date: '',
  });

  // ── Data Loading ──────────────────────────────────────────

  const loadException = useCallback(async () => {
    try {
      const res = await api.request<any>(`/exceptions/${id}`);
      setException(res.data);
    } catch (err: any) {
      setError(err.message);
    }
  }, [id]);

  const loadReviews = useCallback(async () => {
    try {
      const res = await api.request<any>(`/exceptions/${id}/reviews`);
      setReviews(res.data || []);
    } catch { /* ignore */ }
  }, [id]);

  const loadAuditTrail = useCallback(async () => {
    try {
      const res = await api.request<any>(`/exceptions/${id}/audit-trail`);
      setAuditTrail(res.data || []);
    } catch { /* ignore */ }
  }, [id]);

  const loadImpact = useCallback(async () => {
    try {
      const res = await api.request<any>(`/exceptions/impact/${id}`);
      setImpact(res.data);
    } catch { /* ignore */ }
  }, [id]);

  useEffect(() => {
    Promise.all([loadException(), loadReviews(), loadAuditTrail(), loadImpact()])
      .finally(() => setLoading(false));
  }, [loadException, loadReviews, loadAuditTrail, loadImpact]);

  // ── Action Handlers ───────────────────────────────────────

  const handleSubmit = async () => {
    setActionLoading(true);
    try {
      await api.request<any>(`/exceptions/${id}/submit`, { method: 'POST' });
      setActionMessage('Exception submitted for approval');
      loadException();
      loadAuditTrail();
    } catch (err: any) {
      setActionMessage(`Error: ${err.message}`);
    } finally {
      setActionLoading(false);
    }
  };

  const handleApprove = async () => {
    setActionLoading(true);
    try {
      await api.request<any>(`/exceptions/${id}/approve`, {
        method: 'POST', body: { comments: approveComments },
      });
      setActionMessage('Exception approved');
      setShowApproveModal(false);
      setApproveComments('');
      loadException();
      loadAuditTrail();
    } catch (err: any) {
      setActionMessage(`Error: ${err.message}`);
    } finally {
      setActionLoading(false);
    }
  };

  const handleReject = async () => {
    if (!rejectReason) { setActionMessage('Rejection reason is required'); return; }
    setActionLoading(true);
    try {
      await api.request<any>(`/exceptions/${id}/reject`, {
        method: 'POST', body: { reason: rejectReason },
      });
      setActionMessage('Exception rejected');
      setShowRejectModal(false);
      setRejectReason('');
      loadException();
      loadAuditTrail();
    } catch (err: any) {
      setActionMessage(`Error: ${err.message}`);
    } finally {
      setActionLoading(false);
    }
  };

  const handleRevoke = async () => {
    if (!revokeReason) { setActionMessage('Revocation reason is required'); return; }
    setActionLoading(true);
    try {
      await api.request<any>(`/exceptions/${id}/revoke`, {
        method: 'POST', body: { reason: revokeReason },
      });
      setActionMessage('Exception revoked');
      setShowRevokeModal(false);
      setRevokeReason('');
      loadException();
      loadAuditTrail();
    } catch (err: any) {
      setActionMessage(`Error: ${err.message}`);
    } finally {
      setActionLoading(false);
    }
  };

  const handleRenew = async () => {
    if (!renewExpiry || !renewJustification) {
      setActionMessage('Expiry date and justification are required');
      return;
    }
    setActionLoading(true);
    try {
      await api.request<any>(`/exceptions/${id}/renew`, {
        method: 'POST', body: { new_expiry: renewExpiry, justification: renewJustification },
      });
      setActionMessage('Exception renewed');
      setShowRenewModal(false);
      setRenewExpiry('');
      setRenewJustification('');
      loadException();
      loadAuditTrail();
    } catch (err: any) {
      setActionMessage(`Error: ${err.message}`);
    } finally {
      setActionLoading(false);
    }
  };

  const handleReview = async () => {
    if (!reviewForm.outcome) { setActionMessage('Review outcome is required'); return; }
    setActionLoading(true);
    try {
      await api.request<any>(`/exceptions/${id}/review`, {
        method: 'POST', body: reviewForm,
      });
      setActionMessage('Review recorded');
      setShowReviewModal(false);
      setReviewForm({ review_type: 'periodic', outcome: 'continue', risk_level_at_review: '', findings: '', recommendations: '', next_review_date: '' });
      loadException();
      loadReviews();
      loadAuditTrail();
    } catch (err: any) {
      setActionMessage(`Error: ${err.message}`);
    } finally {
      setActionLoading(false);
    }
  };

  // ── Helpers ───────────────────────────────────────────────

  const formatDate = (d: string | null) => {
    if (!d) return '-';
    return new Date(d).toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric' });
  };

  const formatDateTime = (d: string | null) => {
    if (!d) return '-';
    return new Date(d).toLocaleString('en-GB', {
      day: '2-digit', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit',
    });
  };

  const formatStatus = (s: string) => s.replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase());

  const daysUntil = (d: string | null) => {
    if (!d) return null;
    return Math.ceil((new Date(d).getTime() - Date.now()) / (1000 * 60 * 60 * 24));
  };

  // ── Loading / Error States ────────────────────────────────

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <p className="text-gray-500">Loading exception details...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="rounded-lg bg-red-50 border border-red-200 p-4 text-red-700">{error}</div>
    );
  }

  if (!exception) {
    return (
      <div className="rounded-lg bg-gray-50 border border-gray-200 p-4 text-gray-600">Exception not found.</div>
    );
  }

  const exc = exception;
  const expiryDays = daysUntil(exc.expiry_date);
  const reviewDays = daysUntil(exc.next_review_date);

  // ── Render ────────────────────────────────────────────────

  return (
    <div>
      {/* Back link */}
      <a href="/exceptions" className="inline-flex items-center text-sm text-gray-500 hover:text-indigo-600 mb-4">
        &larr; Back to Exceptions
      </a>

      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-start md:justify-between gap-4 mb-6">
        <div className="flex-1">
          <div className="flex items-center gap-3 mb-2 flex-wrap">
            <span className="text-sm font-mono text-gray-400">{exc.exception_ref}</span>
            <span className={`inline-block px-2.5 py-0.5 rounded-full text-xs font-medium border ${STATUS_COLORS[exc.status] || 'bg-gray-100'}`}>
              {formatStatus(exc.status)}
            </span>
            <span className={`inline-block px-2.5 py-0.5 rounded-full text-xs font-medium border ${RISK_COLORS[exc.residual_risk_level] || ''}`}>
              {exc.residual_risk_level?.replace('_', ' ')} risk
            </span>
            <span className="inline-block px-2 py-0.5 rounded text-xs font-medium bg-gray-100 text-gray-700 capitalize">
              {exc.exception_type}
            </span>
            <span className="inline-block px-2 py-0.5 rounded text-xs font-medium bg-gray-100 text-gray-600 capitalize">
              {exc.priority} priority
            </span>
          </div>
          <h1 className="text-xl font-bold text-gray-900 mb-1">{exc.title}</h1>
          <p className="text-sm text-gray-600">{exc.description}</p>
          {exc.tags && exc.tags.length > 0 && (
            <div className="flex gap-1 mt-2 flex-wrap">
              {exc.tags.map(tag => (
                <span key={tag} className="inline-block px-2 py-0.5 bg-indigo-50 text-indigo-700 rounded text-xs">{tag}</span>
              ))}
            </div>
          )}
        </div>

        {/* Action Buttons */}
        <div className="flex flex-wrap gap-2">
          {exc.status === 'draft' && (
            <button onClick={handleSubmit} disabled={actionLoading}
              className="bg-blue-600 text-white px-3 py-1.5 rounded text-sm font-medium hover:bg-blue-700 disabled:opacity-50">
              Submit for Approval
            </button>
          )}
          {(exc.status === 'pending_risk_assessment' || exc.status === 'pending_approval') && (
            <>
              <button onClick={() => setShowApproveModal(true)}
                className="bg-green-600 text-white px-3 py-1.5 rounded text-sm font-medium hover:bg-green-700">
                Approve
              </button>
              <button onClick={() => setShowRejectModal(true)}
                className="bg-red-600 text-white px-3 py-1.5 rounded text-sm font-medium hover:bg-red-700">
                Reject
              </button>
            </>
          )}
          {exc.status === 'approved' && (
            <>
              <button onClick={() => setShowReviewModal(true)}
                className="bg-indigo-600 text-white px-3 py-1.5 rounded text-sm font-medium hover:bg-indigo-700">
                Record Review
              </button>
              <button onClick={() => setShowRenewModal(true)}
                className="bg-amber-600 text-white px-3 py-1.5 rounded text-sm font-medium hover:bg-amber-700">
                Renew
              </button>
              <button onClick={() => setShowRevokeModal(true)}
                className="bg-red-600 text-white px-3 py-1.5 rounded text-sm font-medium hover:bg-red-700">
                Revoke
              </button>
            </>
          )}
          {exc.status === 'expired' && (
            <button onClick={() => setShowRenewModal(true)}
              className="bg-amber-600 text-white px-3 py-1.5 rounded text-sm font-medium hover:bg-amber-700">
              Renew
            </button>
          )}
        </div>
      </div>

      {/* Action message */}
      {actionMessage && (
        <div className={`rounded-lg p-3 text-sm mb-4 ${actionMessage.startsWith('Error') ? 'bg-red-50 text-red-700 border border-red-200' : 'bg-green-50 text-green-700 border border-green-200'}`}>
          {actionMessage}
          <button onClick={() => setActionMessage('')} className="ml-2 text-xs underline">dismiss</button>
        </div>
      )}

      {/* Tabs */}
      <div className="flex border-b border-gray-200 mb-6">
        {(['details', 'impact', 'reviews', 'audit'] as const).map(tab => (
          <button key={tab} onClick={() => setActiveTab(tab)}
            className={`px-4 py-2 text-sm font-medium border-b-2 -mb-px ${
              activeTab === tab ? 'border-indigo-600 text-indigo-600' : 'border-transparent text-gray-500 hover:text-gray-700'
            }`}>
            {tab === 'details' ? 'Details' : tab === 'impact' ? 'Compliance Impact' : tab === 'reviews' ? `Reviews (${reviews.length})` : `Audit Trail (${auditTrail.length})`}
          </button>
        ))}
      </div>

      {/* Details Tab */}
      {activeTab === 'details' && (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {/* Scope & Affected Controls */}
          <div className="bg-white border border-gray-200 rounded-lg p-5">
            <h3 className="text-sm font-semibold text-gray-700 mb-3">Scope</h3>
            <dl className="space-y-2 text-sm">
              <div className="flex justify-between">
                <dt className="text-gray-500">Scope Type</dt>
                <dd className="text-gray-900 capitalize">{exc.scope_type?.replace(/_/g, ' ')}</dd>
              </div>
              {exc.scope_description && (
                <div><dt className="text-gray-500 mb-1">Scope Description</dt><dd className="text-gray-700">{exc.scope_description}</dd></div>
              )}
              {exc.framework_control_codes && exc.framework_control_codes.length > 0 && (
                <div>
                  <dt className="text-gray-500 mb-1">Framework Controls</dt>
                  <dd className="flex flex-wrap gap-1">
                    {exc.framework_control_codes.map(code => (
                      <span key={code} className="px-2 py-0.5 bg-gray-100 rounded text-xs font-mono">{code}</span>
                    ))}
                  </dd>
                </div>
              )}
              {exc.control_implementation_ids && exc.control_implementation_ids.length > 0 && (
                <div className="flex justify-between">
                  <dt className="text-gray-500">Control Implementations</dt>
                  <dd className="text-gray-900">{exc.control_implementation_ids.length} control(s)</dd>
                </div>
              )}
            </dl>
          </div>

          {/* Risk & Justification */}
          <div className="bg-white border border-gray-200 rounded-lg p-5">
            <h3 className="text-sm font-semibold text-gray-700 mb-3">Risk Assessment</h3>
            <dl className="space-y-2 text-sm">
              <div>
                <dt className="text-gray-500 mb-1">Risk Justification</dt>
                <dd className="text-gray-700">{exc.risk_justification}</dd>
              </div>
              {exc.residual_risk_description && (
                <div>
                  <dt className="text-gray-500 mb-1">Residual Risk Description</dt>
                  <dd className="text-gray-700">{exc.residual_risk_description}</dd>
                </div>
              )}
              <div className="flex justify-between">
                <dt className="text-gray-500">Risk Accepted By</dt>
                <dd className="text-gray-900 font-mono text-xs">{exc.risk_accepted_by || '-'}</dd>
              </div>
              <div className="flex justify-between">
                <dt className="text-gray-500">Risk Accepted At</dt>
                <dd className="text-gray-900">{formatDateTime(exc.risk_accepted_at)}</dd>
              </div>
            </dl>
          </div>

          {/* Compensating Controls */}
          <div className="bg-white border border-gray-200 rounded-lg p-5">
            <h3 className="text-sm font-semibold text-gray-700 mb-3">Compensating Controls</h3>
            {exc.has_compensating_controls ? (
              <dl className="space-y-2 text-sm">
                <div>
                  <dt className="text-gray-500 mb-1">Description</dt>
                  <dd className="text-gray-700">{exc.compensating_controls_description || '-'}</dd>
                </div>
                <div className="flex justify-between">
                  <dt className="text-gray-500">Effectiveness</dt>
                  <dd className={`font-medium ${EFFECTIVENESS_COLORS[exc.compensating_effectiveness] || ''}`}>
                    {EFFECTIVENESS_LABELS[exc.compensating_effectiveness] || exc.compensating_effectiveness}
                  </dd>
                </div>
                {exc.compensating_control_ids && exc.compensating_control_ids.length > 0 && (
                  <div className="flex justify-between">
                    <dt className="text-gray-500">Linked Controls</dt>
                    <dd className="text-gray-900">{exc.compensating_control_ids.length} control(s)</dd>
                  </div>
                )}
              </dl>
            ) : (
              <p className="text-sm text-gray-400">No compensating controls defined</p>
            )}
          </div>

          {/* Lifecycle & Validity */}
          <div className="bg-white border border-gray-200 rounded-lg p-5">
            <h3 className="text-sm font-semibold text-gray-700 mb-3">Lifecycle</h3>
            <dl className="space-y-2 text-sm">
              <div className="flex justify-between">
                <dt className="text-gray-500">Effective Date</dt>
                <dd className="text-gray-900">{formatDate(exc.effective_date)}</dd>
              </div>
              <div className="flex justify-between">
                <dt className="text-gray-500">Expiry Date</dt>
                <dd className={expiryDays !== null && expiryDays <= 30 ? 'text-red-600 font-semibold' : 'text-gray-900'}>
                  {exc.expiry_date ? `${formatDate(exc.expiry_date)}${expiryDays !== null && expiryDays > 0 ? ` (${expiryDays}d)` : ''}` : 'No expiry'}
                </dd>
              </div>
              <div className="flex justify-between">
                <dt className="text-gray-500">Review Frequency</dt>
                <dd className="text-gray-900">Every {exc.review_frequency_months} month(s)</dd>
              </div>
              <div className="flex justify-between">
                <dt className="text-gray-500">Next Review</dt>
                <dd className={reviewDays !== null && reviewDays < 0 ? 'text-red-600 font-semibold' : 'text-gray-900'}>
                  {exc.next_review_date ? `${formatDate(exc.next_review_date)}${reviewDays !== null && reviewDays < 0 ? ' (OVERDUE)' : ''}` : '-'}
                </dd>
              </div>
              <div className="flex justify-between">
                <dt className="text-gray-500">Last Reviewed</dt>
                <dd className="text-gray-900">{formatDate(exc.last_review_date)}</dd>
              </div>
              <div className="flex justify-between">
                <dt className="text-gray-500">Renewal Count</dt>
                <dd className="text-gray-900">{exc.renewal_count}{exc.exception_type === 'temporary' ? ' / 2 max' : ''}</dd>
              </div>
              <div className="flex justify-between">
                <dt className="text-gray-500">Approved By</dt>
                <dd className="text-gray-900 font-mono text-xs">{exc.approved_by || '-'}</dd>
              </div>
              <div className="flex justify-between">
                <dt className="text-gray-500">Approved At</dt>
                <dd className="text-gray-900">{formatDateTime(exc.approved_at)}</dd>
              </div>
              {exc.approval_comments && (
                <div>
                  <dt className="text-gray-500 mb-1">Approval Comments</dt>
                  <dd className="text-gray-700">{exc.approval_comments}</dd>
                </div>
              )}
              {exc.rejection_reason && (
                <div>
                  <dt className="text-gray-500 mb-1">Rejection Reason</dt>
                  <dd className="text-red-700">{exc.rejection_reason}</dd>
                </div>
              )}
            </dl>
          </div>

          {/* Conditions & Business Impact */}
          {(exc.conditions || exc.business_impact_if_implemented) && (
            <div className="bg-white border border-gray-200 rounded-lg p-5 md:col-span-2">
              <h3 className="text-sm font-semibold text-gray-700 mb-3">Conditions &amp; Business Impact</h3>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
                {exc.conditions && (
                  <div>
                    <dt className="text-gray-500 mb-1">Conditions</dt>
                    <dd className="text-gray-700">{exc.conditions}</dd>
                  </div>
                )}
                {exc.business_impact_if_implemented && (
                  <div>
                    <dt className="text-gray-500 mb-1">Business Impact if Implemented</dt>
                    <dd className="text-gray-700">{exc.business_impact_if_implemented}</dd>
                  </div>
                )}
              </div>
            </div>
          )}
        </div>
      )}

      {/* Compliance Impact Tab */}
      {activeTab === 'impact' && (
        <div>
          {impact ? (
            <div className="space-y-6">
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                <div className="bg-white border rounded-lg p-4">
                  <p className="text-xs text-gray-500">Affected Controls</p>
                  <p className="text-2xl font-bold text-gray-900">{impact.affected_control_count}</p>
                </div>
                <div className="bg-white border rounded-lg p-4">
                  <p className="text-xs text-gray-500">Score Impact</p>
                  <p className="text-2xl font-bold text-red-600">-{impact.compliance_score_impact}%</p>
                </div>
                <div className="bg-white border rounded-lg p-4">
                  <p className="text-xs text-gray-500">Compensating Coverage</p>
                  <p className="text-2xl font-bold text-blue-600">{impact.compensating_coverage}%</p>
                </div>
                <div className="bg-white border rounded-lg p-4">
                  <p className="text-xs text-gray-500">Net Risk Delta</p>
                  <p className={`text-2xl font-bold capitalize ${
                    impact.net_risk_delta === 'high' ? 'text-red-600' : impact.net_risk_delta === 'moderate' ? 'text-orange-600' : 'text-green-600'
                  }`}>{impact.net_risk_delta}</p>
                </div>
              </div>

              {impact.affected_frameworks.length > 0 && (
                <div className="bg-white border rounded-lg p-5">
                  <h3 className="text-sm font-semibold text-gray-700 mb-2">Affected Frameworks</h3>
                  <div className="flex gap-2 flex-wrap">
                    {impact.affected_frameworks.map(fw => (
                      <span key={fw} className="px-3 py-1 bg-indigo-50 text-indigo-700 rounded-full text-sm font-medium">{fw}</span>
                    ))}
                  </div>
                </div>
              )}

              {impact.affected_controls.length > 0 && (
                <div className="bg-white border rounded-lg p-5">
                  <h3 className="text-sm font-semibold text-gray-700 mb-3">Affected Controls</h3>
                  <table className="min-w-full text-sm">
                    <thead>
                      <tr className="border-b">
                        <th className="text-left py-2 text-xs text-gray-500">Code</th>
                        <th className="text-left py-2 text-xs text-gray-500">Name</th>
                        <th className="text-left py-2 text-xs text-gray-500">Framework</th>
                        <th className="text-left py-2 text-xs text-gray-500">Impact</th>
                      </tr>
                    </thead>
                    <tbody>
                      {impact.affected_controls.map((ctrl, i) => (
                        <tr key={i} className="border-b last:border-0">
                          <td className="py-2 font-mono text-xs">{ctrl.control_code}</td>
                          <td className="py-2">{ctrl.control_name}</td>
                          <td className="py-2">{ctrl.framework}</td>
                          <td className="py-2 capitalize text-orange-600">{ctrl.impact}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}

              {impact.recommendations.length > 0 && (
                <div className="bg-amber-50 border border-amber-200 rounded-lg p-5">
                  <h3 className="text-sm font-semibold text-amber-800 mb-2">Recommendations</h3>
                  <ul className="list-disc list-inside space-y-1 text-sm text-amber-700">
                    {impact.recommendations.map((rec, i) => <li key={i}>{rec}</li>)}
                  </ul>
                </div>
              )}
            </div>
          ) : (
            <div className="text-sm text-gray-500 bg-gray-50 rounded-lg p-8 text-center">
              Compliance impact analysis not available.
            </div>
          )}
        </div>
      )}

      {/* Reviews Tab */}
      {activeTab === 'reviews' && (
        <div>
          {reviews.length === 0 ? (
            <div className="text-sm text-gray-500 bg-gray-50 rounded-lg p-8 text-center">
              No reviews recorded yet.
            </div>
          ) : (
            <div className="space-y-4">
              {reviews.map(review => (
                <div key={review.id} className="bg-white border border-gray-200 rounded-lg p-4">
                  <div className="flex items-center justify-between mb-2">
                    <div className="flex items-center gap-2">
                      <span className={`px-2 py-0.5 rounded text-xs font-medium ${OUTCOME_COLORS[review.outcome] || 'bg-gray-100'}`}>
                        {review.outcome}
                      </span>
                      <span className="text-xs text-gray-500 capitalize">{review.review_type?.replace(/_/g, ' ')}</span>
                    </div>
                    <span className="text-xs text-gray-400">{formatDateTime(review.review_date)}</span>
                  </div>
                  {review.risk_level_at_review && (
                    <p className="text-xs text-gray-500 mb-1">Risk at review: <span className="capitalize font-medium">{review.risk_level_at_review}</span></p>
                  )}
                  {review.findings && (
                    <div className="mb-2">
                      <p className="text-xs text-gray-500">Findings:</p>
                      <p className="text-sm text-gray-700">{review.findings}</p>
                    </div>
                  )}
                  {review.recommendations && (
                    <div>
                      <p className="text-xs text-gray-500">Recommendations:</p>
                      <p className="text-sm text-gray-700">{review.recommendations}</p>
                    </div>
                  )}
                  {review.compensating_effective !== null && (
                    <p className="text-xs text-gray-500 mt-1">
                      Compensating controls effective: <span className={review.compensating_effective ? 'text-green-600' : 'text-red-600'}>{review.compensating_effective ? 'Yes' : 'No'}</span>
                    </p>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* Audit Trail Tab */}
      {activeTab === 'audit' && (
        <div>
          {auditTrail.length === 0 ? (
            <div className="text-sm text-gray-500 bg-gray-50 rounded-lg p-8 text-center">
              No audit trail entries.
            </div>
          ) : (
            <div className="relative">
              <div className="absolute left-4 top-0 bottom-0 w-px bg-gray-200" />
              <div className="space-y-4">
                {auditTrail.map(entry => (
                  <div key={entry.id} className="relative pl-10">
                    <div className="absolute left-2 top-1 w-5 h-5 rounded-full bg-white border-2 border-gray-300 flex items-center justify-center">
                      <span className="text-[8px] text-gray-400">
                        {entry.action === 'approved' ? 'OK' : entry.action === 'rejected' ? 'X' : entry.action === 'created' ? '+' : '~'}
                      </span>
                    </div>
                    <div className="bg-white border border-gray-200 rounded-lg p-3">
                      <div className="flex items-center justify-between mb-1">
                        <span className="text-sm font-medium text-gray-900 capitalize">{entry.action.replace(/_/g, ' ')}</span>
                        <span className="text-xs text-gray-400">{formatDateTime(entry.created_at)}</span>
                      </div>
                      {(entry.old_status || entry.new_status) && (
                        <p className="text-xs text-gray-500 mb-1">
                          {entry.old_status && <span className="capitalize">{formatStatus(entry.old_status)}</span>}
                          {entry.old_status && entry.new_status && ' -> '}
                          {entry.new_status && <span className="capitalize font-medium">{formatStatus(entry.new_status)}</span>}
                        </p>
                      )}
                      {entry.details && <p className="text-sm text-gray-600">{entry.details}</p>}
                      {entry.actor_email && <p className="text-xs text-gray-400 mt-1">By: {entry.actor_email}</p>}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      )}

      {/* ── Action Modals ──────────────────────────────────── */}

      {/* Approve Modal */}
      {showApproveModal && (
        <ActionModal title="Approve Exception" onClose={() => setShowApproveModal(false)} onConfirm={handleApprove}
          confirmLabel="Approve" confirmColor="bg-green-600 hover:bg-green-700" loading={actionLoading}>
          <label className="block text-sm font-medium text-gray-700 mb-1">Comments (optional)</label>
          <textarea value={approveComments} onChange={e => setApproveComments(e.target.value)}
            className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm" rows={3}
            placeholder="Add approval comments..." />
        </ActionModal>
      )}

      {/* Reject Modal */}
      {showRejectModal && (
        <ActionModal title="Reject Exception" onClose={() => setShowRejectModal(false)} onConfirm={handleReject}
          confirmLabel="Reject" confirmColor="bg-red-600 hover:bg-red-700" loading={actionLoading}>
          <label className="block text-sm font-medium text-gray-700 mb-1">Reason *</label>
          <textarea value={rejectReason} onChange={e => setRejectReason(e.target.value)}
            className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm" rows={3}
            placeholder="Provide a reason for rejection..." />
        </ActionModal>
      )}

      {/* Revoke Modal */}
      {showRevokeModal && (
        <ActionModal title="Revoke Exception" onClose={() => setShowRevokeModal(false)} onConfirm={handleRevoke}
          confirmLabel="Revoke" confirmColor="bg-red-600 hover:bg-red-700" loading={actionLoading}>
          <label className="block text-sm font-medium text-gray-700 mb-1">Reason *</label>
          <textarea value={revokeReason} onChange={e => setRevokeReason(e.target.value)}
            className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm" rows={3}
            placeholder="Provide a reason for revocation..." />
        </ActionModal>
      )}

      {/* Renew Modal */}
      {showRenewModal && (
        <ActionModal title="Renew Exception" onClose={() => setShowRenewModal(false)} onConfirm={handleRenew}
          confirmLabel="Renew" confirmColor="bg-amber-600 hover:bg-amber-700" loading={actionLoading}>
          <div className="space-y-3">
            {exc.exception_type === 'temporary' && exc.renewal_count >= 2 && (
              <div className="bg-red-50 border border-red-200 rounded p-2 text-sm text-red-700">
                Maximum renewal count (2) reached. This exception cannot be renewed.
              </div>
            )}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">New Expiry Date *</label>
              <input type="date" value={renewExpiry} onChange={e => setRenewExpiry(e.target.value)}
                className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Justification *</label>
              <textarea value={renewJustification} onChange={e => setRenewJustification(e.target.value)}
                className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm" rows={3}
                placeholder="Explain why renewal is needed..." />
            </div>
            <p className="text-xs text-gray-500">Current renewal count: {exc.renewal_count}</p>
          </div>
        </ActionModal>
      )}

      {/* Review Modal */}
      {showReviewModal && (
        <ActionModal title="Record Review" onClose={() => setShowReviewModal(false)} onConfirm={handleReview}
          confirmLabel="Submit Review" confirmColor="bg-indigo-600 hover:bg-indigo-700" loading={actionLoading}>
          <div className="space-y-3">
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Review Type</label>
                <select value={reviewForm.review_type}
                  onChange={e => setReviewForm({ ...reviewForm, review_type: e.target.value })}
                  className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm">
                  <option value="periodic">Periodic</option>
                  <option value="risk_reassessment">Risk Reassessment</option>
                  <option value="incident_triggered">Incident Triggered</option>
                  <option value="audit_triggered">Audit Triggered</option>
                  <option value="renewal">Renewal</option>
                  <option value="ad_hoc">Ad Hoc</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Outcome *</label>
                <select value={reviewForm.outcome}
                  onChange={e => setReviewForm({ ...reviewForm, outcome: e.target.value })}
                  className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm">
                  <option value="continue">Continue</option>
                  <option value="modify">Modify</option>
                  <option value="escalate">Escalate</option>
                  <option value="revoke">Revoke</option>
                  <option value="renew">Renew</option>
                </select>
              </div>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Risk Level at Review</label>
              <select value={reviewForm.risk_level_at_review}
                onChange={e => setReviewForm({ ...reviewForm, risk_level_at_review: e.target.value })}
                className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm">
                <option value="">-- Unchanged --</option>
                <option value="critical">Critical</option>
                <option value="high">High</option>
                <option value="medium">Medium</option>
                <option value="low">Low</option>
                <option value="very_low">Very Low</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Findings</label>
              <textarea value={reviewForm.findings}
                onChange={e => setReviewForm({ ...reviewForm, findings: e.target.value })}
                className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm" rows={2} />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Recommendations</label>
              <textarea value={reviewForm.recommendations}
                onChange={e => setReviewForm({ ...reviewForm, recommendations: e.target.value })}
                className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm" rows={2} />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Next Review Date</label>
              <input type="date" value={reviewForm.next_review_date}
                onChange={e => setReviewForm({ ...reviewForm, next_review_date: e.target.value })}
                className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm" />
            </div>
          </div>
        </ActionModal>
      )}
    </div>
  );
}

// ============================================================
// ACTION MODAL COMPONENT
// ============================================================

function ActionModal({
  title, onClose, onConfirm, confirmLabel, confirmColor, loading, children,
}: {
  title: string;
  onClose: () => void;
  onConfirm: () => void;
  confirmLabel: string;
  confirmColor: string;
  loading: boolean;
  children: React.ReactNode;
}) {
  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 z-50 flex items-center justify-center p-4">
      <div className="bg-white rounded-xl shadow-xl w-full max-w-md">
        <div className="p-6">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-bold text-gray-900">{title}</h3>
            <button onClick={onClose} className="text-gray-400 hover:text-gray-600 text-xl">&times;</button>
          </div>
          {children}
          <div className="flex justify-end gap-2 mt-4 pt-4 border-t border-gray-200">
            <button onClick={onClose} className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800">
              Cancel
            </button>
            <button onClick={onConfirm} disabled={loading}
              className={`${confirmColor} text-white px-4 py-2 rounded-lg text-sm font-medium disabled:opacity-50`}>
              {loading ? 'Processing...' : confirmLabel}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
