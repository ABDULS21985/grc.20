'use client';

import { useEffect, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import Link from 'next/link';
import api from '@/lib/api';

// ============================================================
// TYPES
// ============================================================

interface VendorAssessment {
  id: string;
  assessment_ref: string;
  vendor_id: string;
  vendor_name: string;
  questionnaire_name: string;
  questionnaire_id: string;
  status: string;
  overall_score: number | null;
  risk_rating: string | null;
  pass_fail: string;
  due_date: string | null;
  sent_at: string | null;
  sent_to_email: string;
  sent_to_name: string;
  submitted_at: string | null;
  reviewed_at: string | null;
  reviewed_by: string | null;
  review_notes: string;
  reviewer_override_score: number | null;
  reviewer_override_reason: string;
  reminder_count: number;
  critical_findings: number;
  high_findings: number;
  follow_up_required: boolean;
  follow_up_items: unknown[];
  next_assessment_date: string | null;
  section_scores: Record<string, SectionScore>;
  responses: AssessmentResponse[];
  created_at: string;
}

interface SectionScore {
  section_id: string;
  section_name: string;
  score: number;
  weight: number;
  answered: number;
  total: number;
}

interface AssessmentResponse {
  id: string;
  question_id: string;
  question_text: string;
  question_type: string;
  section_name: string;
  risk_impact: string;
  answer_value: string;
  answer_score: number | null;
  evidence_paths: string[];
  evidence_notes: string;
  reviewer_comment: string;
  reviewer_flag: string;
}

// ============================================================
// HELPERS
// ============================================================

const statusColors: Record<string, string> = {
  draft: 'bg-gray-100 text-gray-800',
  sent: 'bg-blue-100 text-blue-800',
  in_progress: 'bg-yellow-100 text-yellow-800',
  submitted: 'bg-purple-100 text-purple-800',
  under_review: 'bg-indigo-100 text-indigo-800',
  completed: 'bg-green-100 text-green-800',
  expired: 'bg-red-100 text-red-800',
  cancelled: 'bg-gray-100 text-gray-500',
};

const riskColors: Record<string, string> = {
  critical: 'bg-red-100 text-red-800 border-red-200',
  high: 'bg-orange-100 text-orange-800 border-orange-200',
  medium: 'bg-yellow-100 text-yellow-800 border-yellow-200',
  low: 'bg-green-100 text-green-800 border-green-200',
};

const passFailColors: Record<string, string> = {
  pass: 'bg-green-100 text-green-800',
  fail: 'bg-red-100 text-red-800',
  conditional_pass: 'bg-yellow-100 text-yellow-800',
  pending: 'bg-gray-100 text-gray-600',
};

const flagColors: Record<string, string> = {
  none: 'bg-gray-50 text-gray-500',
  acceptable: 'bg-green-50 text-green-700',
  needs_clarification: 'bg-yellow-50 text-yellow-700',
  concern: 'bg-orange-50 text-orange-700',
  critical_finding: 'bg-red-50 text-red-700',
};

const impactColors: Record<string, string> = {
  critical: 'text-red-600',
  high: 'text-orange-600',
  medium: 'text-yellow-600',
  low: 'text-green-600',
  informational: 'text-gray-500',
};

function formatDate(dateStr: string | null): string {
  if (!dateStr) return '-';
  return new Date(dateStr).toLocaleDateString('en-GB', {
    day: '2-digit', month: 'short', year: 'numeric',
  });
}

function formatDateTime(dateStr: string | null): string {
  if (!dateStr) return '-';
  return new Date(dateStr).toLocaleString('en-GB', {
    day: '2-digit', month: 'short', year: 'numeric',
    hour: '2-digit', minute: '2-digit',
  });
}

function formatScore(score: number | null): string {
  if (score === null || score === undefined) return '-';
  return score.toFixed(1) + '%';
}

function getScoreColor(score: number | null): string {
  if (score === null) return 'text-gray-500';
  if (score >= 80) return 'text-green-600';
  if (score >= 60) return 'text-yellow-600';
  if (score >= 40) return 'text-orange-600';
  return 'text-red-600';
}

// ============================================================
// COMPONENT
// ============================================================

export default function VendorAssessmentDetailPage() {
  const params = useParams();
  const router = useRouter();
  const assessmentId = params.id as string;

  const [assessment, setAssessment] = useState<VendorAssessment | null>(null);
  const [loading, setLoading] = useState(true);
  const [activeSection, setActiveSection] = useState<string>('');
  const [showReviewForm, setShowReviewForm] = useState(false);
  const [reviewData, setReviewData] = useState({
    review_notes: '',
    pass_fail: '',
    override_score: '',
    override_reason: '',
    follow_up_required: false,
  });
  const [submitting, setSubmitting] = useState(false);

  // ── Load assessment ─────────────────────────────────────
  useEffect(() => {
    setLoading(true);
    api.request<any>(`/vendor-assessments/${assessmentId}`)
      .then((res) => {
        setAssessment(res.data);
        // Set the first section as active
        if (res.data?.responses?.length > 0) {
          setActiveSection(res.data.responses[0].section_name);
        }
      })
      .catch(console.error)
      .finally(() => setLoading(false));
  }, [assessmentId]);

  // ── Action handlers ─────────────────────────────────────
  const handleSendReminder = async () => {
    try {
      await api.request<any>(`/vendor-assessments/${assessmentId}/reminder`, { method: 'POST' });
      alert('Reminder sent successfully.');
      // Refresh
      const res = await api.request<any>(`/vendor-assessments/${assessmentId}`);
      setAssessment(res.data);
    } catch {
      alert('Failed to send reminder.');
    }
  };

  const handleSubmitReview = async () => {
    if (!reviewData.pass_fail) {
      alert('Please select a Pass/Fail result.');
      return;
    }
    setSubmitting(true);
    try {
      const body: Record<string, unknown> = {
        review_notes: reviewData.review_notes,
        pass_fail: reviewData.pass_fail,
        follow_up_required: reviewData.follow_up_required,
      };
      if (reviewData.override_score) {
        body.override_score = parseFloat(reviewData.override_score);
        body.override_reason = reviewData.override_reason;
      }
      await api.request<any>(`/vendor-assessments/${assessmentId}/review`, {
        method: 'POST',
        body,
      });
      alert('Review submitted successfully.');
      const res = await api.request<any>(`/vendor-assessments/${assessmentId}`);
      setAssessment(res.data);
      setShowReviewForm(false);
    } catch {
      alert('Failed to submit review.');
    } finally {
      setSubmitting(false);
    }
  };

  const handleFlagResponse = async (responseId: string, flag: string, comment: string) => {
    try {
      await api.request<any>(`/vendor-assessments/${assessmentId}/review`, {
        method: 'POST',
        body: {
          review_notes: assessment?.review_notes || '',
          pass_fail: assessment?.pass_fail || 'pending',
          response_flags: [{ response_id: responseId, flag, comment }],
        },
      });
      const res = await api.request<any>(`/vendor-assessments/${assessmentId}`);
      setAssessment(res.data);
    } catch {
      alert('Failed to update flag.');
    }
  };

  // ── Loading state ───────────────────────────────────────
  if (loading) {
    return (
      <div className="p-6 max-w-7xl mx-auto flex items-center justify-center py-16">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500" />
        <span className="ml-3 text-gray-500">Loading assessment...</span>
      </div>
    );
  }

  if (!assessment) {
    return (
      <div className="p-6 max-w-7xl mx-auto">
        <p className="text-red-500">Assessment not found.</p>
        <Link href="/vendor-assessments" className="text-blue-600 hover:underline text-sm mt-2 inline-block">
          Back to Assessments
        </Link>
      </div>
    );
  }

  // ── Group responses by section ──────────────────────────
  const sections = new Map<string, AssessmentResponse[]>();
  (assessment.responses || []).forEach((resp) => {
    const existing = sections.get(resp.section_name) || [];
    existing.push(resp);
    sections.set(resp.section_name, existing);
  });
  const sectionNames = Array.from(sections.keys());

  const canReview = assessment.status === 'submitted' || assessment.status === 'under_review';
  const canRemind = assessment.status === 'sent' || assessment.status === 'in_progress';

  // ============================================================
  // RENDER
  // ============================================================

  return (
    <div className="p-6 max-w-7xl mx-auto">
      {/* Header */}
      <div className="flex items-center gap-2 text-sm text-gray-500 mb-4">
        <Link href="/vendor-assessments" className="hover:text-blue-600">Vendor Assessments</Link>
        <span>/</span>
        <span className="text-gray-900 font-medium">{assessment.assessment_ref}</span>
      </div>

      <div className="flex items-start justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">{assessment.assessment_ref}</h1>
          <p className="text-sm text-gray-500 mt-1">
            {assessment.vendor_name} &mdash; {assessment.questionnaire_name}
          </p>
        </div>
        <div className="flex gap-2">
          {canRemind && (
            <button
              onClick={handleSendReminder}
              className="px-4 py-2 text-sm font-medium text-gray-700 border border-gray-300 rounded-md hover:bg-gray-50"
            >
              Send Reminder
            </button>
          )}
          {canReview && (
            <button
              onClick={() => setShowReviewForm(true)}
              className="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700"
            >
              Review Assessment
            </button>
          )}
        </div>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-4 mb-6">
        <div className="bg-white border border-gray-200 rounded-lg p-4">
          <p className="text-xs text-gray-500">Status</p>
          <span className={`mt-1 inline-block px-2 py-0.5 rounded text-xs font-medium ${statusColors[assessment.status] || ''}`}>
            {assessment.status.replace(/_/g, ' ')}
          </span>
        </div>
        <div className="bg-white border border-gray-200 rounded-lg p-4">
          <p className="text-xs text-gray-500">Overall Score</p>
          <p className={`text-xl font-bold ${getScoreColor(assessment.overall_score)}`}>
            {formatScore(assessment.overall_score)}
          </p>
        </div>
        <div className="bg-white border border-gray-200 rounded-lg p-4">
          <p className="text-xs text-gray-500">Risk Rating</p>
          {assessment.risk_rating ? (
            <span className={`mt-1 inline-block px-2 py-0.5 rounded text-xs font-medium capitalize ${riskColors[assessment.risk_rating] || ''}`}>
              {assessment.risk_rating}
            </span>
          ) : <p className="text-sm text-gray-400 mt-1">-</p>}
        </div>
        <div className="bg-white border border-gray-200 rounded-lg p-4">
          <p className="text-xs text-gray-500">Result</p>
          <span className={`mt-1 inline-block px-2 py-0.5 rounded text-xs font-medium ${passFailColors[assessment.pass_fail] || ''}`}>
            {assessment.pass_fail.replace(/_/g, ' ')}
          </span>
        </div>
        <div className="bg-white border border-gray-200 rounded-lg p-4">
          <p className="text-xs text-gray-500">Critical / High</p>
          <p className="text-sm font-medium mt-1">
            <span className="text-red-600">{assessment.critical_findings}C</span>
            {' / '}
            <span className="text-orange-600">{assessment.high_findings}H</span>
          </p>
        </div>
        <div className="bg-white border border-gray-200 rounded-lg p-4">
          <p className="text-xs text-gray-500">Due Date</p>
          <p className="text-sm font-medium mt-1">{formatDate(assessment.due_date)}</p>
        </div>
      </div>

      {/* Assessment Details */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-6">
        <div className="bg-white border border-gray-200 rounded-lg p-5">
          <h3 className="text-sm font-semibold text-gray-700 mb-3">Assessment Details</h3>
          <dl className="space-y-2 text-sm">
            <div className="flex justify-between">
              <dt className="text-gray-500">Sent to</dt>
              <dd className="text-gray-900">{assessment.sent_to_email}</dd>
            </div>
            <div className="flex justify-between">
              <dt className="text-gray-500">Sent at</dt>
              <dd className="text-gray-900">{formatDateTime(assessment.sent_at)}</dd>
            </div>
            <div className="flex justify-between">
              <dt className="text-gray-500">Submitted at</dt>
              <dd className="text-gray-900">{formatDateTime(assessment.submitted_at)}</dd>
            </div>
            <div className="flex justify-between">
              <dt className="text-gray-500">Reminders sent</dt>
              <dd className="text-gray-900">{assessment.reminder_count}</dd>
            </div>
            <div className="flex justify-between">
              <dt className="text-gray-500">Follow-up required</dt>
              <dd className="text-gray-900">{assessment.follow_up_required ? 'Yes' : 'No'}</dd>
            </div>
          </dl>
        </div>

        {/* Section Score Breakdown */}
        <div className="lg:col-span-2 bg-white border border-gray-200 rounded-lg p-5">
          <h3 className="text-sm font-semibold text-gray-700 mb-3">Section Scores</h3>
          {assessment.section_scores && Object.keys(assessment.section_scores).length > 0 ? (
            <div className="space-y-3">
              {Object.entries(assessment.section_scores).map(([name, sec]) => (
                <div key={name}>
                  <div className="flex items-center justify-between text-sm mb-1">
                    <span className="text-gray-700">{name}</span>
                    <span className={`font-medium ${getScoreColor(sec.score)}`}>
                      {sec.score.toFixed(1)}% ({sec.answered}/{sec.total})
                    </span>
                  </div>
                  <div className="w-full bg-gray-200 rounded-full h-2">
                    <div
                      className={`h-2 rounded-full ${
                        sec.score >= 80 ? 'bg-green-500' :
                        sec.score >= 60 ? 'bg-yellow-500' :
                        sec.score >= 40 ? 'bg-orange-500' : 'bg-red-500'
                      }`}
                      style={{ width: `${Math.min(sec.score, 100)}%` }}
                    />
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-sm text-gray-500">Scores not yet calculated.</p>
          )}
        </div>
      </div>

      {/* Review Notes (if reviewed) */}
      {assessment.review_notes && (
        <div className="bg-white border border-gray-200 rounded-lg p-5 mb-6">
          <h3 className="text-sm font-semibold text-gray-700 mb-2">Review Notes</h3>
          <p className="text-sm text-gray-700 whitespace-pre-wrap">{assessment.review_notes}</p>
          {assessment.reviewer_override_score !== null && (
            <div className="mt-3 p-3 bg-yellow-50 border border-yellow-200 rounded">
              <p className="text-sm font-medium text-yellow-800">
                Reviewer Override Score: {assessment.reviewer_override_score}%
              </p>
              {assessment.reviewer_override_reason && (
                <p className="text-sm text-yellow-700 mt-1">{assessment.reviewer_override_reason}</p>
              )}
            </div>
          )}
        </div>
      )}

      {/* Responses by Section */}
      {sectionNames.length > 0 && (
        <div className="bg-white border border-gray-200 rounded-lg overflow-hidden mb-6">
          <div className="border-b border-gray-200 px-4 py-3 flex items-center justify-between">
            <h3 className="text-sm font-semibold text-gray-700">Responses</h3>
            <span className="text-xs text-gray-500">{assessment.responses?.length || 0} responses</span>
          </div>

          {/* Section tabs */}
          <div className="border-b border-gray-200 px-4 flex gap-1 overflow-x-auto">
            {sectionNames.map((name) => (
              <button
                key={name}
                onClick={() => setActiveSection(name)}
                className={`px-3 py-2 text-xs font-medium whitespace-nowrap border-b-2 ${
                  activeSection === name
                    ? 'border-blue-500 text-blue-600'
                    : 'border-transparent text-gray-500 hover:text-gray-700'
                }`}
              >
                {name}
              </button>
            ))}
          </div>

          {/* Active section responses */}
          <div className="divide-y divide-gray-100">
            {(sections.get(activeSection) || []).map((resp, idx) => (
              <div key={resp.id} className="p-4">
                <div className="flex items-start justify-between mb-2">
                  <div className="flex-1">
                    <div className="flex items-center gap-2 mb-1">
                      <span className="text-xs text-gray-400">Q{idx + 1}</span>
                      <span className={`text-xs font-medium capitalize ${impactColors[resp.risk_impact] || ''}`}>
                        {resp.risk_impact}
                      </span>
                    </div>
                    <p className="text-sm text-gray-900">{resp.question_text}</p>
                  </div>
                  <div className="ml-4 text-right">
                    {resp.answer_score !== null ? (
                      <span className={`text-lg font-bold ${getScoreColor(resp.answer_score)}`}>
                        {resp.answer_score.toFixed(0)}
                      </span>
                    ) : (
                      <span className="text-sm text-gray-400">-</span>
                    )}
                  </div>
                </div>

                {/* Answer */}
                <div className="bg-gray-50 rounded p-3 mb-2">
                  <p className="text-sm text-gray-700">
                    <span className="font-medium">Answer:</span> {resp.answer_value || 'No response'}
                  </p>
                  {resp.evidence_notes && (
                    <p className="text-xs text-gray-500 mt-1">
                      <span className="font-medium">Evidence notes:</span> {resp.evidence_notes}
                    </p>
                  )}
                  {resp.evidence_paths.length > 0 && (
                    <div className="mt-1">
                      <span className="text-xs font-medium text-gray-500">Evidence files: </span>
                      {resp.evidence_paths.map((path, i) => (
                        <span key={i} className="text-xs text-blue-600 mr-2">{path}</span>
                      ))}
                    </div>
                  )}
                </div>

                {/* Reviewer flag / comment */}
                <div className="flex items-center gap-2">
                  <span className={`px-2 py-0.5 rounded text-xs font-medium ${flagColors[resp.reviewer_flag] || flagColors.none}`}>
                    {resp.reviewer_flag.replace(/_/g, ' ')}
                  </span>
                  {resp.reviewer_comment && (
                    <span className="text-xs text-gray-500 italic">{resp.reviewer_comment}</span>
                  )}
                  {canReview && (
                    <div className="ml-auto flex gap-1">
                      {['acceptable', 'needs_clarification', 'concern', 'critical_finding'].map((flag) => (
                        <button
                          key={flag}
                          onClick={() => {
                            const comment = flag === 'acceptable' ? '' : prompt('Comment for this flag:') || '';
                            handleFlagResponse(resp.id, flag, comment);
                          }}
                          className={`px-2 py-0.5 rounded text-xs ${
                            resp.reviewer_flag === flag
                              ? flagColors[flag]
                              : 'bg-gray-100 text-gray-500 hover:bg-gray-200'
                          }`}
                          title={flag.replace(/_/g, ' ')}
                        >
                          {flag === 'acceptable' ? 'OK' :
                           flag === 'needs_clarification' ? '?' :
                           flag === 'concern' ? '!' : '!!'}
                        </button>
                      ))}
                    </div>
                  )}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Review Form Modal */}
      {showReviewForm && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 w-full max-w-lg mx-4">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">Review Assessment</h2>

            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Result</label>
                <select
                  value={reviewData.pass_fail}
                  onChange={(e) => setReviewData({ ...reviewData, pass_fail: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm"
                >
                  <option value="">Select result...</option>
                  <option value="pass">Pass</option>
                  <option value="conditional_pass">Conditional Pass</option>
                  <option value="fail">Fail</option>
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Review Notes</label>
                <textarea
                  value={reviewData.review_notes}
                  onChange={(e) => setReviewData({ ...reviewData, review_notes: e.target.value })}
                  rows={4}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm"
                  placeholder="Provide your assessment of the vendor's responses..."
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Override Score (optional)
                </label>
                <input
                  type="number"
                  min="0"
                  max="100"
                  step="0.1"
                  value={reviewData.override_score}
                  onChange={(e) => setReviewData({ ...reviewData, override_score: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm"
                  placeholder="Leave empty to use calculated score"
                />
              </div>

              {reviewData.override_score && (
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Override Reason
                  </label>
                  <input
                    type="text"
                    value={reviewData.override_reason}
                    onChange={(e) => setReviewData({ ...reviewData, override_reason: e.target.value })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm"
                    placeholder="Reason for score override"
                  />
                </div>
              )}

              <div className="flex items-center gap-2">
                <input
                  type="checkbox"
                  id="follow_up"
                  checked={reviewData.follow_up_required}
                  onChange={(e) => setReviewData({ ...reviewData, follow_up_required: e.target.checked })}
                  className="rounded border-gray-300"
                />
                <label htmlFor="follow_up" className="text-sm text-gray-700">
                  Follow-up required
                </label>
              </div>
            </div>

            <div className="flex justify-end gap-3 mt-6">
              <button
                onClick={() => setShowReviewForm(false)}
                className="px-4 py-2 text-sm font-medium text-gray-700 border border-gray-300 rounded-md hover:bg-gray-50"
              >
                Cancel
              </button>
              <button
                onClick={handleSubmitReview}
                disabled={submitting || !reviewData.pass_fail}
                className="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700 disabled:opacity-50"
              >
                {submitting ? 'Submitting...' : 'Submit Review'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
