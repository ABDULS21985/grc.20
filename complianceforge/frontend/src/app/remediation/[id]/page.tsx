'use client';

import { useEffect, useState, useCallback } from 'react';
import { useParams } from 'next/navigation';
import api from '@/lib/api';

interface RemediationAction {
  id: string;
  action_ref: string;
  sort_order: number;
  title: string;
  description: string;
  action_type: string;
  framework_control_code: string;
  priority: string;
  estimated_hours: number;
  estimated_cost_eur: number;
  required_skills: string[];
  assigned_to: string | null;
  target_start_date: string | null;
  target_end_date: string | null;
  status: string;
  actual_hours: number | null;
  actual_cost_eur: number | null;
  completion_notes: string;
  evidence_paths: string[];
  ai_implementation_guidance: string;
  ai_evidence_suggestions: string[];
  ai_tool_recommendations: string[];
  ai_risk_if_deferred: string;
  ai_cross_framework_benefit: string;
}

interface RemediationPlan {
  id: string;
  plan_ref: string;
  name: string;
  description: string;
  plan_type: string;
  status: string;
  priority: string;
  ai_generated: boolean;
  ai_model: string;
  ai_confidence_score: number | null;
  human_reviewed: boolean;
  target_completion_date: string | null;
  estimated_total_hours: number;
  estimated_total_cost_eur: number;
  completion_percentage: number;
  created_at: string;
  updated_at: string;
  actions: RemediationAction[];
}

interface PlanProgress {
  completion_percentage: number;
  total_actions: number;
  completed_actions: number;
  in_progress_actions: number;
  blocked_actions: number;
  pending_actions: number;
  estimated_hours_total: number;
  actual_hours_total: number;
  estimated_cost_total: number;
  actual_cost_total: number;
  on_track: boolean;
  days_remaining: number;
  days_overdue: number;
  critical_path_actions: Array<{
    action_id: string;
    action_ref: string;
    title: string;
    status: string;
    priority: string;
    days_left: number;
  }>;
}

const kanbanColumns = [
  { key: 'pending', label: 'Pending', color: 'border-gray-300 bg-gray-50' },
  { key: 'assigned', label: 'Assigned', color: 'border-blue-300 bg-blue-50' },
  { key: 'in_progress', label: 'In Progress', color: 'border-indigo-300 bg-indigo-50' },
  { key: 'in_review', label: 'In Review', color: 'border-yellow-300 bg-yellow-50' },
  { key: 'completed', label: 'Completed', color: 'border-green-300 bg-green-50' },
];

const statusColors: Record<string, string> = {
  draft: 'bg-gray-100 text-gray-700',
  pending_approval: 'bg-yellow-100 text-yellow-800',
  approved: 'bg-blue-100 text-blue-800',
  in_progress: 'bg-indigo-100 text-indigo-800',
  on_hold: 'bg-orange-100 text-orange-800',
  completed: 'bg-green-100 text-green-800',
  cancelled: 'bg-red-100 text-red-700',
};

const priorityConfig: Record<string, { color: string; bg: string }> = {
  critical: { color: 'text-red-700', bg: 'bg-red-100' },
  high: { color: 'text-orange-700', bg: 'bg-orange-100' },
  medium: { color: 'text-yellow-700', bg: 'bg-yellow-100' },
  low: { color: 'text-green-700', bg: 'bg-green-100' },
};

const actionTypeLabels: Record<string, string> = {
  implement_control: 'Implement Control',
  update_policy: 'Update Policy',
  deploy_technical: 'Deploy Technical',
  conduct_training: 'Training',
  perform_assessment: 'Assessment',
  gather_evidence: 'Gather Evidence',
  review_process: 'Review Process',
  third_party_engagement: 'Third Party',
  documentation: 'Documentation',
  monitoring_setup: 'Monitoring',
};

export default function PlanDetailPage() {
  const params = useParams();
  const id = params.id as string;

  const [plan, setPlan] = useState<RemediationPlan | null>(null);
  const [progress, setProgress] = useState<PlanProgress | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedAction, setSelectedAction] = useState<RemediationAction | null>(null);
  const [activeTab, setActiveTab] = useState<'kanban' | 'timeline' | 'progress'>('kanban');

  const loadPlan = useCallback(() => {
    setLoading(true);
    Promise.all([
      api.get(`/remediation/plans/${id}`),
      api.get(`/remediation/plans/${id}/progress`),
    ])
      .then(([planRes, progressRes]: [{ data: { data: RemediationPlan } }, { data: { data: PlanProgress } }]) => {
        setPlan(planRes.data?.data || null);
        setProgress(progressRes.data?.data || null);
      })
      .catch((err: Error) => setError(err.message))
      .finally(() => setLoading(false));
  }, [id]);

  useEffect(() => {
    loadPlan();
  }, [loadPlan]);

  const handleStatusChange = async (actionId: string, newStatus: string) => {
    try {
      await api.put(`/remediation/actions/${actionId}`, { status: newStatus });
      loadPlan();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update action');
    }
  };

  const handleApprove = async () => {
    try {
      await api.post(`/remediation/plans/${id}/approve`, {});
      loadPlan();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to approve plan');
    }
  };

  const handleCompleteAction = async (actionId: string) => {
    const notes = prompt('Enter completion notes:');
    if (notes === null) return;

    const hoursStr = prompt('Actual hours spent:', '0');
    const hours = parseFloat(hoursStr || '0');

    try {
      await api.post(`/remediation/actions/${actionId}/complete`, {
        completion_notes: notes,
        actual_hours: hours,
        actual_cost_eur: 0,
        evidence_paths: [],
      });
      setSelectedAction(null);
      loadPlan();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to complete action');
    }
  };

  const getColumnActions = (status: string): RemediationAction[] => {
    if (!plan?.actions) return [];
    if (status === 'pending') {
      return plan.actions.filter(a => a.status === 'pending' || a.status === 'blocked' || a.status === 'deferred');
    }
    return plan.actions.filter(a => a.status === status);
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <p className="text-gray-500">Loading plan...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="rounded-lg bg-red-50 border border-red-200 p-4 text-red-700">{error}</div>
    );
  }

  if (!plan) {
    return (
      <div className="rounded-lg bg-gray-50 border border-gray-200 p-4 text-gray-600">
        Plan not found.
      </div>
    );
  }

  return (
    <div>
      {/* Back link */}
      <a href="/remediation" className="inline-flex items-center text-sm text-gray-500 hover:text-indigo-600 mb-4">
        &larr; Back to Remediation Plans
      </a>

      {/* Header */}
      <div className="flex items-start justify-between mb-6">
        <div>
          <div className="flex items-center gap-2 mb-1">
            <span className="text-sm font-mono text-gray-400">{plan.plan_ref}</span>
            <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${statusColors[plan.status] || 'bg-gray-100'}`}>
              {plan.status.replace(/_/g, ' ')}
            </span>
            {plan.ai_generated && (
              <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-purple-100 text-purple-700">
                AI Generated
                {plan.ai_confidence_score !== null && ` - ${Math.round(plan.ai_confidence_score * 100)}% confidence`}
              </span>
            )}
          </div>
          <h1 className="text-2xl font-bold text-gray-900">{plan.name}</h1>
          {plan.description && (
            <p className="text-gray-500 mt-1 max-w-2xl">{plan.description}</p>
          )}
        </div>
        <div className="flex gap-2">
          {(plan.status === 'draft' || plan.status === 'pending_approval') && (
            <button
              onClick={handleApprove}
              className="px-4 py-2 text-sm font-medium text-white bg-green-600 rounded-lg hover:bg-green-700"
            >
              Approve Plan
            </button>
          )}
        </div>
      </div>

      {/* Progress summary */}
      {progress && (
        <div className="grid grid-cols-6 gap-3 mb-6">
          <div className="rounded-lg bg-white border border-gray-200 p-3 text-center">
            <p className="text-2xl font-bold text-indigo-600">{Math.round(progress.completion_percentage)}%</p>
            <p className="text-xs text-gray-500">Complete</p>
          </div>
          <div className="rounded-lg bg-white border border-gray-200 p-3 text-center">
            <p className="text-2xl font-bold text-gray-900">{progress.total_actions}</p>
            <p className="text-xs text-gray-500">Total Actions</p>
          </div>
          <div className="rounded-lg bg-white border border-gray-200 p-3 text-center">
            <p className="text-2xl font-bold text-green-600">{progress.completed_actions}</p>
            <p className="text-xs text-gray-500">Completed</p>
          </div>
          <div className="rounded-lg bg-white border border-gray-200 p-3 text-center">
            <p className="text-2xl font-bold text-blue-600">{progress.in_progress_actions}</p>
            <p className="text-xs text-gray-500">In Progress</p>
          </div>
          <div className="rounded-lg bg-white border border-gray-200 p-3 text-center">
            <p className={`text-2xl font-bold ${progress.on_track ? 'text-green-600' : 'text-red-600'}`}>
              {progress.on_track ? 'On Track' : 'At Risk'}
            </p>
            <p className="text-xs text-gray-500">
              {progress.days_overdue > 0 ? `${progress.days_overdue}d overdue` : `${progress.days_remaining}d left`}
            </p>
          </div>
          <div className="rounded-lg bg-white border border-gray-200 p-3 text-center">
            <p className="text-2xl font-bold text-gray-900">{progress.actual_hours_total.toFixed(1)}h</p>
            <p className="text-xs text-gray-500">of {progress.estimated_hours_total.toFixed(1)}h est.</p>
          </div>
        </div>
      )}

      {/* Overall progress bar */}
      <div className="mb-6">
        <div className="w-full bg-gray-200 rounded-full h-3">
          <div
            className={`h-3 rounded-full transition-all ${
              plan.completion_percentage === 100
                ? 'bg-green-500'
                : plan.completion_percentage >= 50
                ? 'bg-indigo-500'
                : plan.completion_percentage > 0
                ? 'bg-yellow-500'
                : 'bg-gray-300'
            }`}
            style={{ width: `${Math.min(plan.completion_percentage, 100)}%` }}
          />
        </div>
      </div>

      {/* Tabs */}
      <div className="flex gap-1 mb-6 border-b border-gray-200">
        {([['kanban', 'Kanban Board'], ['timeline', 'Timeline'], ['progress', 'Analytics']] as const).map(([key, label]) => (
          <button
            key={key}
            onClick={() => setActiveTab(key)}
            className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
              activeTab === key
                ? 'border-indigo-600 text-indigo-600'
                : 'border-transparent text-gray-500 hover:text-gray-700'
            }`}
          >
            {label}
          </button>
        ))}
      </div>

      {/* Kanban view */}
      {activeTab === 'kanban' && (
        <div className="flex gap-4 overflow-x-auto pb-4">
          {kanbanColumns.map(column => {
            const columnActions = getColumnActions(column.key);
            return (
              <div
                key={column.key}
                className={`flex-shrink-0 w-72 rounded-lg border-2 ${column.color} p-3`}
              >
                <div className="flex items-center justify-between mb-3">
                  <h3 className="text-sm font-semibold text-gray-700">{column.label}</h3>
                  <span className="inline-flex items-center justify-center w-6 h-6 rounded-full bg-white text-xs font-bold text-gray-600 border border-gray-300">
                    {columnActions.length}
                  </span>
                </div>
                <div className="space-y-2 min-h-[200px]">
                  {columnActions.map(action => (
                    <div
                      key={action.id}
                      onClick={() => setSelectedAction(action)}
                      className="bg-white rounded-lg border border-gray-200 p-3 cursor-pointer hover:shadow-md transition-shadow"
                    >
                      <div className="flex items-center gap-1 mb-1">
                        <span className="text-[10px] font-mono text-gray-400">{action.action_ref}</span>
                        <span className={`inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-medium ${priorityConfig[action.priority]?.bg || 'bg-gray-100'} ${priorityConfig[action.priority]?.color || 'text-gray-700'}`}>
                          {action.priority}
                        </span>
                      </div>
                      <p className="text-sm font-medium text-gray-900 line-clamp-2">{action.title}</p>
                      <div className="flex items-center gap-2 mt-2 text-[10px] text-gray-400">
                        <span>{actionTypeLabels[action.action_type] || action.action_type}</span>
                        {action.estimated_hours > 0 && <span>{action.estimated_hours}h</span>}
                        {action.framework_control_code && <span className="font-mono">{action.framework_control_code}</span>}
                      </div>
                      {action.status === 'blocked' && (
                        <span className="inline-flex items-center mt-2 px-1.5 py-0.5 rounded text-[10px] font-medium bg-red-100 text-red-700">
                          Blocked
                        </span>
                      )}
                      {action.status === 'deferred' && (
                        <span className="inline-flex items-center mt-2 px-1.5 py-0.5 rounded text-[10px] font-medium bg-orange-100 text-orange-700">
                          Deferred
                        </span>
                      )}

                      {/* Quick status change buttons */}
                      {column.key !== 'completed' && (
                        <div className="flex gap-1 mt-2 pt-2 border-t border-gray-100">
                          {column.key === 'pending' && (
                            <button
                              onClick={(e) => { e.stopPropagation(); handleStatusChange(action.id, 'in_progress'); }}
                              className="text-[10px] px-2 py-0.5 rounded bg-indigo-50 text-indigo-600 hover:bg-indigo-100"
                            >
                              Start
                            </button>
                          )}
                          {column.key === 'assigned' && (
                            <button
                              onClick={(e) => { e.stopPropagation(); handleStatusChange(action.id, 'in_progress'); }}
                              className="text-[10px] px-2 py-0.5 rounded bg-indigo-50 text-indigo-600 hover:bg-indigo-100"
                            >
                              Start Work
                            </button>
                          )}
                          {column.key === 'in_progress' && (
                            <>
                              <button
                                onClick={(e) => { e.stopPropagation(); handleStatusChange(action.id, 'in_review'); }}
                                className="text-[10px] px-2 py-0.5 rounded bg-yellow-50 text-yellow-700 hover:bg-yellow-100"
                              >
                                Review
                              </button>
                              <button
                                onClick={(e) => { e.stopPropagation(); handleCompleteAction(action.id); }}
                                className="text-[10px] px-2 py-0.5 rounded bg-green-50 text-green-700 hover:bg-green-100"
                              >
                                Complete
                              </button>
                            </>
                          )}
                          {column.key === 'in_review' && (
                            <button
                              onClick={(e) => { e.stopPropagation(); handleCompleteAction(action.id); }}
                              className="text-[10px] px-2 py-0.5 rounded bg-green-50 text-green-700 hover:bg-green-100"
                            >
                              Approve & Complete
                            </button>
                          )}
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              </div>
            );
          })}
        </div>
      )}

      {/* Timeline view */}
      {activeTab === 'timeline' && plan.actions && (
        <div className="space-y-3">
          {plan.actions.sort((a, b) => a.sort_order - b.sort_order).map((action, index) => (
            <div
              key={action.id}
              className="flex items-start gap-4 p-4 bg-white rounded-lg border border-gray-200"
            >
              <div className="flex-shrink-0 flex flex-col items-center">
                <div className={`w-8 h-8 rounded-full flex items-center justify-center text-sm font-bold text-white ${
                  action.status === 'completed' ? 'bg-green-500' :
                  action.status === 'in_progress' ? 'bg-indigo-500' :
                  action.status === 'blocked' ? 'bg-red-500' :
                  'bg-gray-300'
                }`}>
                  {action.status === 'completed' ? (
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M5 13l4 4L19 7" />
                    </svg>
                  ) : (
                    index + 1
                  )}
                </div>
                {index < plan.actions.length - 1 && (
                  <div className={`w-0.5 h-8 mt-1 ${action.status === 'completed' ? 'bg-green-300' : 'bg-gray-200'}`} />
                )}
              </div>
              <div className="flex-1">
                <div className="flex items-center gap-2 mb-1">
                  <span className="text-xs font-mono text-gray-400">{action.action_ref}</span>
                  <span className={`inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-medium ${priorityConfig[action.priority]?.bg || 'bg-gray-100'} ${priorityConfig[action.priority]?.color || 'text-gray-700'}`}>
                    {action.priority}
                  </span>
                  <span className={`inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-medium ${statusColors[action.status] || 'bg-gray-100 text-gray-700'}`}>
                    {action.status.replace(/_/g, ' ')}
                  </span>
                </div>
                <h4 className="text-sm font-semibold text-gray-900">{action.title}</h4>
                {action.description && (
                  <p className="text-xs text-gray-500 mt-1">{action.description}</p>
                )}
                <div className="flex items-center gap-4 mt-2 text-xs text-gray-400">
                  {action.target_start_date && <span>Start: {new Date(action.target_start_date).toLocaleDateString()}</span>}
                  {action.target_end_date && <span>End: {new Date(action.target_end_date).toLocaleDateString()}</span>}
                  <span>{action.estimated_hours}h estimated</span>
                  {action.framework_control_code && <span className="font-mono">{action.framework_control_code}</span>}
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Progress analytics view */}
      {activeTab === 'progress' && progress && (
        <div className="grid grid-cols-2 gap-6">
          {/* Critical Path */}
          <div className="bg-white rounded-lg border border-gray-200 p-5">
            <h3 className="text-sm font-semibold text-gray-900 mb-4">Critical Path Actions</h3>
            {progress.critical_path_actions && progress.critical_path_actions.length > 0 ? (
              <div className="space-y-3">
                {progress.critical_path_actions.map(action => (
                  <div key={action.action_id} className="flex items-center justify-between p-3 bg-red-50 rounded-lg border border-red-200">
                    <div>
                      <span className="text-xs font-mono text-gray-400">{action.action_ref}</span>
                      <p className="text-sm font-medium text-gray-900">{action.title}</p>
                    </div>
                    <div className="text-right">
                      <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${priorityConfig[action.priority]?.bg || 'bg-gray-100'} ${priorityConfig[action.priority]?.color || 'text-gray-700'}`}>
                        {action.priority}
                      </span>
                      <p className="text-xs text-gray-500 mt-1">
                        {action.days_left > 0 ? `${action.days_left}d left` : 'Overdue'}
                      </p>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-sm text-gray-500">No critical path actions identified.</p>
            )}
          </div>

          {/* Cost tracking */}
          <div className="bg-white rounded-lg border border-gray-200 p-5">
            <h3 className="text-sm font-semibold text-gray-900 mb-4">Cost & Effort Tracking</h3>
            <div className="space-y-4">
              <div>
                <div className="flex justify-between text-sm mb-1">
                  <span className="text-gray-500">Hours</span>
                  <span className="font-medium">{progress.actual_hours_total.toFixed(1)} / {progress.estimated_hours_total.toFixed(1)}h</span>
                </div>
                <div className="w-full bg-gray-200 rounded-full h-2">
                  <div
                    className="bg-indigo-500 h-2 rounded-full"
                    style={{ width: `${Math.min(progress.estimated_hours_total > 0 ? (progress.actual_hours_total / progress.estimated_hours_total) * 100 : 0, 100)}%` }}
                  />
                </div>
              </div>
              <div>
                <div className="flex justify-between text-sm mb-1">
                  <span className="text-gray-500">Cost (EUR)</span>
                  <span className="font-medium">
                    {progress.actual_cost_total.toLocaleString()} / {progress.estimated_cost_total.toLocaleString()}
                  </span>
                </div>
                <div className="w-full bg-gray-200 rounded-full h-2">
                  <div
                    className="bg-green-500 h-2 rounded-full"
                    style={{ width: `${Math.min(progress.estimated_cost_total > 0 ? (progress.actual_cost_total / progress.estimated_cost_total) * 100 : 0, 100)}%` }}
                  />
                </div>
              </div>
            </div>
          </div>

          {/* Status breakdown */}
          <div className="bg-white rounded-lg border border-gray-200 p-5">
            <h3 className="text-sm font-semibold text-gray-900 mb-4">Status Breakdown</h3>
            <div className="space-y-2">
              {[
                { label: 'Completed', count: progress.completed_actions, color: 'bg-green-500' },
                { label: 'In Progress', count: progress.in_progress_actions, color: 'bg-indigo-500' },
                { label: 'Blocked', count: progress.blocked_actions, color: 'bg-red-500' },
                { label: 'Pending', count: progress.pending_actions, color: 'bg-gray-400' },
              ].map(item => (
                <div key={item.label} className="flex items-center gap-3">
                  <span className={`w-3 h-3 rounded-full ${item.color}`} />
                  <span className="text-sm text-gray-600 flex-1">{item.label}</span>
                  <span className="text-sm font-semibold text-gray-900">{item.count}</span>
                  <span className="text-xs text-gray-400">
                    ({progress.total_actions > 0 ? Math.round((item.count / progress.total_actions) * 100) : 0}%)
                  </span>
                </div>
              ))}
            </div>
          </div>

          {/* AI info */}
          {plan.ai_generated && (
            <div className="bg-white rounded-lg border border-gray-200 p-5">
              <h3 className="text-sm font-semibold text-gray-900 mb-4">AI Generation Info</h3>
              <div className="space-y-2 text-sm">
                <div className="flex justify-between">
                  <span className="text-gray-500">Model</span>
                  <span className="font-mono text-gray-900">{plan.ai_model}</span>
                </div>
                {plan.ai_confidence_score !== null && (
                  <div className="flex justify-between">
                    <span className="text-gray-500">Confidence</span>
                    <span className="font-medium text-gray-900">{Math.round(plan.ai_confidence_score * 100)}%</span>
                  </div>
                )}
                <div className="flex justify-between">
                  <span className="text-gray-500">Human Reviewed</span>
                  <span className={`font-medium ${plan.human_reviewed ? 'text-green-600' : 'text-amber-600'}`}>
                    {plan.human_reviewed ? 'Yes' : 'Pending'}
                  </span>
                </div>
              </div>
            </div>
          )}
        </div>
      )}

      {/* Action detail modal */}
      {selectedAction && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="bg-white rounded-xl shadow-2xl w-full max-w-2xl max-h-[90vh] overflow-y-auto p-6">
            <div className="flex items-start justify-between mb-4">
              <div>
                <div className="flex items-center gap-2 mb-1">
                  <span className="text-xs font-mono text-gray-400">{selectedAction.action_ref}</span>
                  <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${priorityConfig[selectedAction.priority]?.bg || 'bg-gray-100'} ${priorityConfig[selectedAction.priority]?.color || 'text-gray-700'}`}>
                    {selectedAction.priority}
                  </span>
                  <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${statusColors[selectedAction.status] || 'bg-gray-100 text-gray-700'}`}>
                    {selectedAction.status.replace(/_/g, ' ')}
                  </span>
                </div>
                <h2 className="text-lg font-bold text-gray-900">{selectedAction.title}</h2>
              </div>
              <button onClick={() => setSelectedAction(null)} className="text-gray-400 hover:text-gray-600">
                <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>

            <div className="space-y-4">
              {selectedAction.description && (
                <div>
                  <h4 className="text-xs font-semibold text-gray-500 uppercase mb-1">Description</h4>
                  <p className="text-sm text-gray-700">{selectedAction.description}</p>
                </div>
              )}

              <div className="grid grid-cols-3 gap-3">
                <div>
                  <h4 className="text-xs font-semibold text-gray-500 uppercase mb-1">Type</h4>
                  <p className="text-sm text-gray-700">{actionTypeLabels[selectedAction.action_type] || selectedAction.action_type}</p>
                </div>
                <div>
                  <h4 className="text-xs font-semibold text-gray-500 uppercase mb-1">Control Code</h4>
                  <p className="text-sm font-mono text-gray-700">{selectedAction.framework_control_code || 'N/A'}</p>
                </div>
                <div>
                  <h4 className="text-xs font-semibold text-gray-500 uppercase mb-1">Estimated</h4>
                  <p className="text-sm text-gray-700">{selectedAction.estimated_hours}h / EUR {selectedAction.estimated_cost_eur.toLocaleString()}</p>
                </div>
              </div>

              {selectedAction.required_skills && selectedAction.required_skills.length > 0 && (
                <div>
                  <h4 className="text-xs font-semibold text-gray-500 uppercase mb-1">Required Skills</h4>
                  <div className="flex flex-wrap gap-1">
                    {selectedAction.required_skills.map((skill, i) => (
                      <span key={i} className="inline-flex items-center px-2 py-0.5 rounded text-xs bg-gray-100 text-gray-700">{skill}</span>
                    ))}
                  </div>
                </div>
              )}

              {selectedAction.ai_implementation_guidance && (
                <div className="bg-purple-50 rounded-lg p-4 border border-purple-200">
                  <h4 className="text-xs font-semibold text-purple-700 uppercase mb-2">AI Implementation Guidance</h4>
                  <p className="text-sm text-gray-700">{selectedAction.ai_implementation_guidance}</p>
                </div>
              )}

              {selectedAction.ai_evidence_suggestions && selectedAction.ai_evidence_suggestions.length > 0 && (
                <div>
                  <h4 className="text-xs font-semibold text-gray-500 uppercase mb-1">AI Evidence Suggestions</h4>
                  <ul className="list-disc list-inside text-sm text-gray-700 space-y-0.5">
                    {selectedAction.ai_evidence_suggestions.map((s, i) => (
                      <li key={i}>{s}</li>
                    ))}
                  </ul>
                </div>
              )}

              {selectedAction.ai_tool_recommendations && selectedAction.ai_tool_recommendations.length > 0 && (
                <div>
                  <h4 className="text-xs font-semibold text-gray-500 uppercase mb-1">AI Tool Recommendations</h4>
                  <div className="flex flex-wrap gap-1">
                    {selectedAction.ai_tool_recommendations.map((tool, i) => (
                      <span key={i} className="inline-flex items-center px-2 py-0.5 rounded text-xs bg-blue-50 text-blue-700 border border-blue-200">{tool}</span>
                    ))}
                  </div>
                </div>
              )}

              {selectedAction.ai_risk_if_deferred && (
                <div className="bg-red-50 rounded-lg p-3 border border-red-200">
                  <h4 className="text-xs font-semibold text-red-700 uppercase mb-1">Risk if Deferred</h4>
                  <p className="text-sm text-gray-700">{selectedAction.ai_risk_if_deferred}</p>
                </div>
              )}

              {selectedAction.ai_cross_framework_benefit && (
                <div className="bg-green-50 rounded-lg p-3 border border-green-200">
                  <h4 className="text-xs font-semibold text-green-700 uppercase mb-1">Cross-Framework Benefit</h4>
                  <p className="text-sm text-gray-700">{selectedAction.ai_cross_framework_benefit}</p>
                </div>
              )}
            </div>

            {/* Modal footer with actions */}
            <div className="flex justify-end gap-2 mt-6 pt-4 border-t border-gray-200">
              {selectedAction.status !== 'completed' && selectedAction.status !== 'cancelled' && (
                <>
                  {selectedAction.status === 'pending' && (
                    <button
                      onClick={() => { handleStatusChange(selectedAction.id, 'in_progress'); setSelectedAction(null); }}
                      className="px-3 py-1.5 text-sm font-medium text-white bg-indigo-600 rounded-lg hover:bg-indigo-700"
                    >
                      Start Work
                    </button>
                  )}
                  {selectedAction.status === 'in_progress' && (
                    <button
                      onClick={() => handleCompleteAction(selectedAction.id)}
                      className="px-3 py-1.5 text-sm font-medium text-white bg-green-600 rounded-lg hover:bg-green-700"
                    >
                      Mark Complete
                    </button>
                  )}
                </>
              )}
              <button
                onClick={() => setSelectedAction(null)}
                className="px-3 py-1.5 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50"
              >
                Close
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
