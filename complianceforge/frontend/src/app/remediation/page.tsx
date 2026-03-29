'use client';

import { useEffect, useState } from 'react';
import api from '@/lib/api';

interface RemediationPlan {
  id: string;
  plan_ref: string;
  name: string;
  description: string;
  plan_type: string;
  status: string;
  priority: string;
  ai_generated: boolean;
  ai_confidence_score: number | null;
  human_reviewed: boolean;
  completion_percentage: number;
  estimated_total_hours: number;
  estimated_total_cost_eur: number;
  target_completion_date: string | null;
  owner_user_id: string | null;
  created_at: string;
  updated_at: string;
}

interface GenerateFormState {
  name: string;
  plan_type: string;
  scope_description: string;
  priority: string;
  risk_appetite: string;
  budget_eur: string;
  timeline_months: string;
  industry_context: string;
  gaps: Array<{
    control_code: string;
    control_title: string;
    framework_code: string;
    current_status: string;
    target_status: string;
    gap_severity: string;
  }>;
}

const statusColors: Record<string, string> = {
  draft: 'bg-gray-100 text-gray-700',
  pending_approval: 'bg-yellow-100 text-yellow-800',
  approved: 'bg-blue-100 text-blue-800',
  in_progress: 'bg-indigo-100 text-indigo-800',
  on_hold: 'bg-orange-100 text-orange-800',
  completed: 'bg-green-100 text-green-800',
  cancelled: 'bg-red-100 text-red-700',
};

const priorityColors: Record<string, string> = {
  critical: 'bg-red-600',
  high: 'bg-orange-500',
  medium: 'bg-yellow-500',
  low: 'bg-green-500',
};

const planTypeLabels: Record<string, string> = {
  gap_remediation: 'Gap Remediation',
  audit_finding: 'Audit Finding',
  risk_treatment: 'Risk Treatment',
  continuous_improvement: 'Continuous Improvement',
  incident_response: 'Incident Response',
  regulatory_change: 'Regulatory Change',
};

export default function RemediationPlansPage() {
  const [plans, setPlans] = useState<RemediationPlan[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showGenerateModal, setShowGenerateModal] = useState(false);
  const [generating, setGenerating] = useState(false);
  const [statusFilter, setStatusFilter] = useState('all');

  const [form, setForm] = useState<GenerateFormState>({
    name: '',
    plan_type: 'gap_remediation',
    scope_description: '',
    priority: 'medium',
    risk_appetite: 'moderate',
    budget_eur: '',
    timeline_months: '3',
    industry_context: '',
    gaps: [{ control_code: '', control_title: '', framework_code: 'ISO27001', current_status: 'not_implemented', target_status: 'implemented', gap_severity: 'medium' }],
  });

  useEffect(() => {
    loadPlans();
  }, []);

  const loadPlans = () => {
    setLoading(true);
    api.get('/remediation/plans')
      .then((res: { data: { data: RemediationPlan[] } }) => setPlans(res.data?.data || []))
      .catch((err: Error) => setError(err.message))
      .finally(() => setLoading(false));
  };

  const handleGenerate = async () => {
    setGenerating(true);
    setError(null);

    try {
      const payload = {
        name: form.name || 'AI-Generated Remediation Plan',
        plan_type: form.plan_type,
        scope_description: form.scope_description,
        priority: form.priority,
        use_ai: true,
        ai_request: {
          plan_type: form.plan_type,
          scope_description: form.scope_description,
          risk_appetite: form.risk_appetite,
          budget_eur: parseFloat(form.budget_eur) || 0,
          timeline_months: parseInt(form.timeline_months) || 3,
          industry_context: form.industry_context,
          gaps: form.gaps.filter(g => g.control_code !== ''),
        },
      };

      await api.post('/remediation/plans/generate', payload);
      setShowGenerateModal(false);
      loadPlans();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to generate plan');
    } finally {
      setGenerating(false);
    }
  };

  const addGap = () => {
    setForm(prev => ({
      ...prev,
      gaps: [...prev.gaps, { control_code: '', control_title: '', framework_code: 'ISO27001', current_status: 'not_implemented', target_status: 'implemented', gap_severity: 'medium' }],
    }));
  };

  const removeGap = (index: number) => {
    setForm(prev => ({
      ...prev,
      gaps: prev.gaps.filter((_, i) => i !== index),
    }));
  };

  const updateGap = (index: number, field: string, value: string) => {
    setForm(prev => ({
      ...prev,
      gaps: prev.gaps.map((g, i) => i === index ? { ...g, [field]: value } : g),
    }));
  };

  const filteredPlans = plans.filter(p =>
    statusFilter === 'all' || p.status === statusFilter
  );

  const stats = {
    total: plans.length,
    aiGenerated: plans.filter(p => p.ai_generated).length,
    inProgress: plans.filter(p => p.status === 'in_progress').length,
    completed: plans.filter(p => p.status === 'completed').length,
    avgCompletion: plans.length > 0
      ? Math.round(plans.reduce((sum, p) => sum + p.completion_percentage, 0) / plans.length)
      : 0,
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <p className="text-gray-500">Loading remediation plans...</p>
      </div>
    );
  }

  return (
    <div>
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Remediation Plans</h1>
          <p className="text-gray-500 mt-1">AI-assisted compliance remediation planning and tracking</p>
        </div>
        <div className="flex gap-2">
          <button
            onClick={() => setShowGenerateModal(true)}
            className="inline-flex items-center px-4 py-2 bg-indigo-600 text-white text-sm font-medium rounded-lg hover:bg-indigo-700 transition-colors"
          >
            <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
            </svg>
            Generate AI Plan
          </button>
        </div>
      </div>

      {/* Error display */}
      {error && (
        <div className="rounded-lg bg-red-50 border border-red-200 p-4 text-red-700 mb-6">
          {error}
          <button onClick={() => setError(null)} className="ml-4 text-sm underline">Dismiss</button>
        </div>
      )}

      {/* Summary cards */}
      <div className="grid grid-cols-5 gap-3 mb-6">
        {[
          { label: 'Total Plans', value: stats.total, color: 'bg-indigo-600' },
          { label: 'AI Generated', value: stats.aiGenerated, color: 'bg-purple-600' },
          { label: 'In Progress', value: stats.inProgress, color: 'bg-blue-500' },
          { label: 'Completed', value: stats.completed, color: 'bg-green-500' },
          { label: 'Avg Completion', value: `${stats.avgCompletion}%`, color: 'bg-yellow-500' },
        ].map(s => (
          <div key={s.label} className="rounded-lg bg-white border border-gray-200 p-4 text-center">
            <div className={`inline-flex h-10 w-10 items-center justify-center rounded-full ${s.color} text-white text-sm font-bold`}>
              {s.value}
            </div>
            <p className="text-xs text-gray-600 mt-2">{s.label}</p>
          </div>
        ))}
      </div>

      {/* Status filter */}
      <div className="flex gap-2 mb-4">
        {['all', 'draft', 'pending_approval', 'approved', 'in_progress', 'completed'].map(status => (
          <button
            key={status}
            onClick={() => setStatusFilter(status)}
            className={`px-3 py-1 text-xs rounded-full transition-colors ${
              statusFilter === status
                ? 'bg-indigo-600 text-white'
                : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
            }`}
          >
            {status === 'all' ? 'All' : status.replace(/_/g, ' ')}
          </button>
        ))}
      </div>

      {/* Plans list */}
      {filteredPlans.length === 0 ? (
        <div className="rounded-lg bg-gray-50 border border-gray-200 p-12 text-center">
          <p className="text-gray-500 mb-4">No remediation plans found.</p>
          <button
            onClick={() => setShowGenerateModal(true)}
            className="inline-flex items-center px-4 py-2 bg-indigo-600 text-white text-sm font-medium rounded-lg hover:bg-indigo-700"
          >
            Generate Your First AI Plan
          </button>
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-4">
          {filteredPlans.map(plan => (
            <a
              key={plan.id}
              href={`/remediation/${plan.id}`}
              className="block rounded-lg bg-white border border-gray-200 p-5 hover:border-indigo-300 hover:shadow-md transition-all"
            >
              <div className="flex items-start justify-between">
                <div className="flex-1">
                  <div className="flex items-center gap-2 mb-1">
                    <span className="text-xs font-mono text-gray-400">{plan.plan_ref}</span>
                    <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${statusColors[plan.status] || 'bg-gray-100 text-gray-700'}`}>
                      {plan.status.replace(/_/g, ' ')}
                    </span>
                    <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs font-medium bg-gray-100 text-gray-600">
                      <span className={`w-2 h-2 rounded-full ${priorityColors[plan.priority] || 'bg-gray-400'}`} />
                      {plan.priority}
                    </span>
                    {plan.ai_generated && (
                      <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-purple-100 text-purple-700">
                        AI Generated
                        {plan.ai_confidence_score !== null && ` (${Math.round(plan.ai_confidence_score * 100)}%)`}
                      </span>
                    )}
                    {plan.ai_generated && !plan.human_reviewed && (
                      <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-amber-100 text-amber-700">
                        Needs Review
                      </span>
                    )}
                  </div>
                  <h3 className="text-base font-semibold text-gray-900">{plan.name}</h3>
                  {plan.description && (
                    <p className="text-sm text-gray-500 mt-1 line-clamp-2">{plan.description}</p>
                  )}
                  <div className="flex items-center gap-4 mt-3 text-xs text-gray-400">
                    <span>{planTypeLabels[plan.plan_type] || plan.plan_type}</span>
                    {plan.target_completion_date && (
                      <span>Due: {new Date(plan.target_completion_date).toLocaleDateString()}</span>
                    )}
                    {plan.estimated_total_hours > 0 && (
                      <span>{plan.estimated_total_hours}h estimated</span>
                    )}
                    {plan.estimated_total_cost_eur > 0 && (
                      <span>EUR {plan.estimated_total_cost_eur.toLocaleString()}</span>
                    )}
                  </div>
                </div>
                <div className="ml-6 flex-shrink-0 w-24">
                  <div className="text-right">
                    <span className="text-lg font-bold text-gray-900">{Math.round(plan.completion_percentage)}%</span>
                    <div className="w-full bg-gray-200 rounded-full h-2 mt-1">
                      <div
                        className={`h-2 rounded-full transition-all ${
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
                </div>
              </div>
            </a>
          ))}
        </div>
      )}

      {/* Generate AI Plan Modal */}
      {showGenerateModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="bg-white rounded-xl shadow-2xl w-full max-w-2xl max-h-[90vh] overflow-y-auto p-6">
            <div className="flex items-center justify-between mb-6">
              <div>
                <h2 className="text-xl font-bold text-gray-900">Generate AI Remediation Plan</h2>
                <p className="text-sm text-gray-500 mt-1">Provide context for the AI to generate an optimized remediation plan</p>
              </div>
              <button onClick={() => setShowGenerateModal(false)} className="text-gray-400 hover:text-gray-600">
                <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>

            <div className="space-y-4">
              {/* Plan name */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Plan Name</label>
                <input
                  type="text"
                  value={form.name}
                  onChange={e => setForm(prev => ({ ...prev, name: e.target.value }))}
                  placeholder="e.g., ISO 27001 Gap Remediation Q2 2026"
                  className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500"
                />
              </div>

              {/* Type and Priority */}
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Plan Type</label>
                  <select
                    value={form.plan_type}
                    onChange={e => setForm(prev => ({ ...prev, plan_type: e.target.value }))}
                    className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500"
                  >
                    {Object.entries(planTypeLabels).map(([k, v]) => (
                      <option key={k} value={k}>{v}</option>
                    ))}
                  </select>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Priority</label>
                  <select
                    value={form.priority}
                    onChange={e => setForm(prev => ({ ...prev, priority: e.target.value }))}
                    className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500"
                  >
                    <option value="critical">Critical</option>
                    <option value="high">High</option>
                    <option value="medium">Medium</option>
                    <option value="low">Low</option>
                  </select>
                </div>
              </div>

              {/* Scope */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Scope Description</label>
                <textarea
                  value={form.scope_description}
                  onChange={e => setForm(prev => ({ ...prev, scope_description: e.target.value }))}
                  placeholder="Describe the scope of this remediation plan..."
                  rows={2}
                  className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500"
                />
              </div>

              {/* Budget and Timeline */}
              <div className="grid grid-cols-3 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Budget (EUR)</label>
                  <input
                    type="number"
                    value={form.budget_eur}
                    onChange={e => setForm(prev => ({ ...prev, budget_eur: e.target.value }))}
                    placeholder="50000"
                    className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Timeline (months)</label>
                  <input
                    type="number"
                    value={form.timeline_months}
                    onChange={e => setForm(prev => ({ ...prev, timeline_months: e.target.value }))}
                    className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Risk Appetite</label>
                  <select
                    value={form.risk_appetite}
                    onChange={e => setForm(prev => ({ ...prev, risk_appetite: e.target.value }))}
                    className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500"
                  >
                    <option value="very_low">Very Low</option>
                    <option value="low">Low</option>
                    <option value="moderate">Moderate</option>
                    <option value="high">High</option>
                  </select>
                </div>
              </div>

              {/* Industry */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Industry Context</label>
                <input
                  type="text"
                  value={form.industry_context}
                  onChange={e => setForm(prev => ({ ...prev, industry_context: e.target.value }))}
                  placeholder="e.g., Financial Services, Healthcare, Government"
                  className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500"
                />
              </div>

              {/* Compliance Gaps */}
              <div>
                <div className="flex items-center justify-between mb-2">
                  <label className="block text-sm font-medium text-gray-700">Compliance Gaps</label>
                  <button
                    onClick={addGap}
                    className="text-xs text-indigo-600 hover:text-indigo-800 font-medium"
                  >
                    + Add Gap
                  </button>
                </div>
                <div className="space-y-3">
                  {form.gaps.map((gap, index) => (
                    <div key={index} className="p-3 bg-gray-50 rounded-lg border border-gray-200">
                      <div className="flex items-center justify-between mb-2">
                        <span className="text-xs font-medium text-gray-500">Gap {index + 1}</span>
                        {form.gaps.length > 1 && (
                          <button onClick={() => removeGap(index)} className="text-xs text-red-500 hover:text-red-700">Remove</button>
                        )}
                      </div>
                      <div className="grid grid-cols-3 gap-2">
                        <input
                          type="text"
                          value={gap.control_code}
                          onChange={e => updateGap(index, 'control_code', e.target.value)}
                          placeholder="Control code (e.g., A.5.1)"
                          className="rounded border border-gray-300 px-2 py-1 text-xs focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500"
                        />
                        <input
                          type="text"
                          value={gap.control_title}
                          onChange={e => updateGap(index, 'control_title', e.target.value)}
                          placeholder="Control title"
                          className="rounded border border-gray-300 px-2 py-1 text-xs focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500"
                        />
                        <select
                          value={gap.framework_code}
                          onChange={e => updateGap(index, 'framework_code', e.target.value)}
                          className="rounded border border-gray-300 px-2 py-1 text-xs focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500"
                        >
                          <option value="ISO27001">ISO 27001</option>
                          <option value="NIST_CSF_2">NIST CSF 2.0</option>
                          <option value="PCI_DSS_4">PCI DSS 4.0</option>
                          <option value="UK_GDPR">UK GDPR</option>
                          <option value="NIST_800_53">NIST 800-53</option>
                        </select>
                      </div>
                      <div className="grid grid-cols-3 gap-2 mt-2">
                        <select
                          value={gap.current_status}
                          onChange={e => updateGap(index, 'current_status', e.target.value)}
                          className="rounded border border-gray-300 px-2 py-1 text-xs focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500"
                        >
                          <option value="not_implemented">Not Implemented</option>
                          <option value="planned">Planned</option>
                          <option value="partial">Partial</option>
                        </select>
                        <select
                          value={gap.target_status}
                          onChange={e => updateGap(index, 'target_status', e.target.value)}
                          className="rounded border border-gray-300 px-2 py-1 text-xs focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500"
                        >
                          <option value="implemented">Implemented</option>
                          <option value="effective">Effective</option>
                        </select>
                        <select
                          value={gap.gap_severity}
                          onChange={e => updateGap(index, 'gap_severity', e.target.value)}
                          className="rounded border border-gray-300 px-2 py-1 text-xs focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500"
                        >
                          <option value="critical">Critical</option>
                          <option value="high">High</option>
                          <option value="medium">Medium</option>
                          <option value="low">Low</option>
                        </select>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </div>

            {/* Modal footer */}
            <div className="flex justify-end gap-3 mt-6 pt-4 border-t border-gray-200">
              <button
                onClick={() => setShowGenerateModal(false)}
                className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50"
              >
                Cancel
              </button>
              <button
                onClick={handleGenerate}
                disabled={generating}
                className="inline-flex items-center px-4 py-2 text-sm font-medium text-white bg-indigo-600 rounded-lg hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {generating ? (
                  <>
                    <svg className="animate-spin -ml-1 mr-2 h-4 w-4 text-white" fill="none" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                    </svg>
                    Generating...
                  </>
                ) : (
                  <>
                    <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
                    </svg>
                    Generate Plan
                  </>
                )}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
