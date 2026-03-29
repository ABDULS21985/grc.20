'use client';

import { useEffect, useState, useCallback } from 'react';
import api from '@/lib/api';

const PLAN_STATUS_COLORS: Record<string, string> = {
  active: 'bg-green-100 text-green-800',
  approved: 'bg-blue-100 text-blue-800',
  draft: 'bg-gray-100 text-gray-600',
  under_review: 'bg-amber-100 text-amber-800',
  archived: 'bg-gray-100 text-gray-400',
};

const SCENARIO_STATUS_COLORS: Record<string, string> = {
  identified: 'bg-gray-100 text-gray-700',
  analysed: 'bg-blue-100 text-blue-700',
  mitigated: 'bg-green-100 text-green-700',
  accepted: 'bg-amber-100 text-amber-700',
};

const EXERCISE_STATUS_COLORS: Record<string, string> = {
  planned: 'bg-blue-100 text-blue-700',
  in_progress: 'bg-amber-100 text-amber-700',
  completed: 'bg-green-100 text-green-700',
  cancelled: 'bg-gray-100 text-gray-400',
};

const RATING_COLORS: Record<string, string> = {
  pass: 'text-green-600',
  pass_with_concerns: 'text-amber-600',
  fail: 'text-red-600',
};

const LIKELIHOOD_ORDER = ['almost_certain', 'likely', 'possible', 'unlikely', 'rare'];

interface BCDashboard {
  processes_without_bia: number;
  plans_requiring_review: number;
  last_exercise_date: string | null;
  last_exercise_result: string;
  rto_coverage_percent: number;
  rpo_coverage_percent: number;
  spof_count: number;
  total_processes: number;
  total_scenarios: number;
  total_plans: number;
  total_exercises: number;
  active_plans: number;
  draft_plans: number;
  mission_critical_count: number;
  business_critical_count: number;
  planned_exercises: number;
  completed_exercises: number;
}

interface BIAScenario {
  id: string;
  scenario_ref: string;
  name: string;
  description: string;
  scenario_type: string;
  likelihood: string;
  affected_process_ids: string[];
  estimated_financial_loss_eur: number | null;
  mitigation_strategy: string;
  status: string;
  created_at: string;
}

interface ContinuityPlan {
  id: string;
  plan_ref: string;
  name: string;
  plan_type: string;
  status: string;
  version: number;
  scope_description: string;
  covered_scenario_ids: string[];
  covered_process_ids: string[];
  owner_user_id: string | null;
  approved_by: string | null;
  approved_at: string | null;
  next_review_date: string | null;
  review_frequency_months: number;
  created_at: string;
}

interface BCExercise {
  id: string;
  exercise_ref: string;
  name: string;
  exercise_type: string;
  plan_id: string | null;
  scenario_id: string | null;
  status: string;
  scheduled_date: string | null;
  actual_date: string | null;
  rto_achieved_hours: number | null;
  rpo_achieved_hours: number | null;
  objectives_met: boolean | null;
  overall_rating: string | null;
  lessons_learned: string;
  gaps_identified: string;
  created_at: string;
}

type Tab = 'dashboard' | 'scenarios' | 'plans' | 'exercises';

export default function BCPage() {
  const [tab, setTab] = useState<Tab>('dashboard');
  const [dashboard, setDashboard] = useState<BCDashboard | null>(null);
  const [scenarios, setScenarios] = useState<BIAScenario[]>([]);
  const [plans, setPlans] = useState<ContinuityPlan[]>([]);
  const [exercises, setExercises] = useState<BCExercise[]>([]);
  const [loading, setLoading] = useState(true);

  const loadData = useCallback(async () => {
    setLoading(true);
    try {
      if (tab === 'dashboard') {
        const res = await api.get('/bc/dashboard');
        setDashboard(res.data?.data || null);
      } else if (tab === 'scenarios') {
        const res = await api.get('/bc/scenarios?page=1&page_size=50');
        setScenarios(res.data?.data || []);
      } else if (tab === 'plans') {
        const res = await api.get('/bc/plans?page=1&page_size=50');
        setPlans(res.data?.data || []);
      } else if (tab === 'exercises') {
        const res = await api.get('/bc/exercises?page=1&page_size=50');
        setExercises(res.data?.data || []);
      }
    } catch {
      /* ignore */
    }
    setLoading(false);
  }, [tab]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const tabs = [
    { key: 'dashboard' as const, label: 'Dashboard' },
    { key: 'scenarios' as const, label: 'Scenarios' },
    { key: 'plans' as const, label: 'Continuity Plans' },
    { key: 'exercises' as const, label: 'Exercises' },
  ];

  return (
    <div>
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Business Continuity</h1>
          <p className="text-gray-500 mt-1">
            Manage disruption scenarios, continuity plans, and exercises
          </p>
        </div>
      </div>

      <div className="flex gap-1 border-b border-gray-200 mb-6 overflow-x-auto">
        {tabs.map((t) => (
          <button
            key={t.key}
            onClick={() => setTab(t.key)}
            className={`px-4 py-2.5 text-sm font-medium border-b-2 whitespace-nowrap transition-colors ${
              tab === t.key
                ? 'border-indigo-600 text-indigo-600'
                : 'border-transparent text-gray-500 hover:text-gray-700'
            }`}
          >
            {t.label}
          </button>
        ))}
      </div>

      {loading ? (
        <div className="card animate-pulse h-96" />
      ) : tab === 'dashboard' && dashboard ? (
        <div>
          {/* Summary Cards */}
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-4 mb-6">
            <div className="card">
              <p className="text-sm text-gray-500">Business Processes</p>
              <p className="text-2xl font-bold text-gray-900 mt-1">{dashboard.total_processes}</p>
              <div className="flex gap-3 mt-2 text-xs">
                <span className="text-red-600">{dashboard.mission_critical_count} mission critical</span>
                <span className="text-orange-600">{dashboard.business_critical_count} business critical</span>
              </div>
            </div>
            <div className="card">
              <p className="text-sm text-gray-500">Continuity Plans</p>
              <p className="text-2xl font-bold text-indigo-600 mt-1">{dashboard.total_plans}</p>
              <div className="flex gap-3 mt-2 text-xs">
                <span className="text-green-600">{dashboard.active_plans} active</span>
                <span className="text-gray-500">{dashboard.draft_plans} draft</span>
              </div>
              {dashboard.plans_requiring_review > 0 && (
                <p className="text-xs text-amber-600 mt-1">{dashboard.plans_requiring_review} need review</p>
              )}
            </div>
            <div className="card">
              <p className="text-sm text-gray-500">RTO / RPO Coverage</p>
              <div className="flex gap-4 mt-2">
                <div>
                  <p className="text-lg font-bold text-gray-900">{dashboard.rto_coverage_percent.toFixed(0)}%</p>
                  <p className="text-xs text-gray-500">RTO defined</p>
                </div>
                <div>
                  <p className="text-lg font-bold text-gray-900">{dashboard.rpo_coverage_percent.toFixed(0)}%</p>
                  <p className="text-xs text-gray-500">RPO defined</p>
                </div>
              </div>
              {dashboard.processes_without_bia > 0 && (
                <p className="text-xs text-amber-600 mt-2">{dashboard.processes_without_bia} processes without BIA</p>
              )}
            </div>
            <div className="card">
              <p className="text-sm text-gray-500">Exercises</p>
              <p className="text-2xl font-bold text-gray-900 mt-1">{dashboard.total_exercises}</p>
              <div className="flex gap-3 mt-2 text-xs">
                <span className="text-blue-600">{dashboard.planned_exercises} planned</span>
                <span className="text-green-600">{dashboard.completed_exercises} completed</span>
              </div>
              {dashboard.last_exercise_date && (
                <p className="text-xs text-gray-500 mt-1">
                  Last: {new Date(dashboard.last_exercise_date).toLocaleDateString('en-GB')}
                  {dashboard.last_exercise_result && (
                    <span className={`ml-1 font-medium ${RATING_COLORS[dashboard.last_exercise_result] || ''}`}>
                      ({dashboard.last_exercise_result.replace(/_/g, ' ')})
                    </span>
                  )}
                </p>
              )}
            </div>
          </div>

          {/* Alerts */}
          <div className="space-y-3 mb-6">
            {dashboard.spof_count > 0 && (
              <div className="card bg-red-50 border-red-200 flex items-center gap-3">
                <div className="flex-shrink-0 w-8 h-8 rounded-full bg-red-100 flex items-center justify-center">
                  <span className="text-red-600 text-sm font-bold">!</span>
                </div>
                <div>
                  <p className="text-sm font-medium text-red-800">
                    {dashboard.spof_count} Single Point{dashboard.spof_count !== 1 ? 's' : ''} of Failure Detected
                  </p>
                  <p className="text-xs text-red-600">
                    Review in BIA &gt; Single Points of Failure to assess mitigation options.
                  </p>
                </div>
              </div>
            )}
            {dashboard.plans_requiring_review > 0 && (
              <div className="card bg-amber-50 border-amber-200 flex items-center gap-3">
                <div className="flex-shrink-0 w-8 h-8 rounded-full bg-amber-100 flex items-center justify-center">
                  <span className="text-amber-700 text-sm font-bold">R</span>
                </div>
                <div>
                  <p className="text-sm font-medium text-amber-800">
                    {dashboard.plans_requiring_review} plan{dashboard.plans_requiring_review !== 1 ? 's' : ''} overdue for review
                  </p>
                  <p className="text-xs text-amber-600">
                    Navigate to Continuity Plans to schedule reviews.
                  </p>
                </div>
              </div>
            )}
          </div>

          {/* Scenario Count */}
          <div className="card">
            <h3 className="font-semibold text-gray-900 mb-2">Scenarios Tracked</h3>
            <p className="text-3xl font-bold text-gray-900">{dashboard.total_scenarios}</p>
            <p className="text-sm text-gray-500 mt-1">
              Disruption scenarios assessed across all business processes.
            </p>
          </div>
        </div>
      ) : tab === 'scenarios' ? (
        <div>
          <div className="card overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="table-header">
                  <th className="px-4 py-3 text-left">Ref</th>
                  <th className="px-4 py-3 text-left">Scenario</th>
                  <th className="px-4 py-3 text-left">Type</th>
                  <th className="px-4 py-3 text-left">Likelihood</th>
                  <th className="px-4 py-3 text-right">Est. Loss (EUR)</th>
                  <th className="px-4 py-3 text-left">Status</th>
                  <th className="px-4 py-3 text-right">Processes</th>
                </tr>
              </thead>
              <tbody>
                {scenarios.map((sc) => (
                  <tr key={sc.id} className="border-t border-gray-100 hover:bg-gray-50">
                    <td className="px-4 py-3 font-mono text-xs text-indigo-600">{sc.scenario_ref}</td>
                    <td className="px-4 py-3">
                      <p className="font-medium text-gray-900">{sc.name}</p>
                      {sc.description && (
                        <p className="text-xs text-gray-400 mt-0.5 truncate max-w-sm">{sc.description}</p>
                      )}
                    </td>
                    <td className="px-4 py-3 text-gray-600 text-xs capitalize">{sc.scenario_type.replace(/_/g, ' ')}</td>
                    <td className="px-4 py-3">
                      <LikelihoodBadge likelihood={sc.likelihood} />
                    </td>
                    <td className="px-4 py-3 text-right font-mono text-sm">
                      {sc.estimated_financial_loss_eur != null
                        ? sc.estimated_financial_loss_eur.toLocaleString()
                        : '--'}
                    </td>
                    <td className="px-4 py-3">
                      <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${SCENARIO_STATUS_COLORS[sc.status] || 'bg-gray-100'}`}>
                        {sc.status}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-right text-gray-600">{sc.affected_process_ids?.length || 0}</td>
                  </tr>
                ))}
                {scenarios.length === 0 && (
                  <tr>
                    <td colSpan={7} className="px-4 py-12 text-center text-gray-500">
                      No disruption scenarios defined. Create one to begin continuity planning.
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </div>
      ) : tab === 'plans' ? (
        <div>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {plans.map((plan) => (
              <div key={plan.id} className="card hover:shadow-md transition-shadow">
                <div className="flex items-start justify-between mb-3">
                  <div>
                    <span className="text-xs font-mono text-indigo-600 bg-indigo-50 px-2 py-0.5 rounded">
                      {plan.plan_ref}
                    </span>
                    <span className={`ml-2 inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${PLAN_STATUS_COLORS[plan.status] || 'bg-gray-100'}`}>
                      {plan.status.replace(/_/g, ' ')}
                    </span>
                  </div>
                  <span className="text-xs text-gray-400">v{plan.version}</span>
                </div>
                <h3 className="font-semibold text-gray-900 mb-1">{plan.name}</h3>
                <p className="text-xs text-gray-500 mb-3 capitalize">{plan.plan_type.replace(/_/g, ' ')}</p>
                {plan.scope_description && (
                  <p className="text-xs text-gray-600 mb-3 line-clamp-2">{plan.scope_description}</p>
                )}
                <div className="flex items-center justify-between text-xs text-gray-400 pt-3 border-t border-gray-100">
                  <span>{plan.covered_process_ids?.length || 0} processes covered</span>
                  {plan.next_review_date && (
                    <span className={new Date(plan.next_review_date) < new Date() ? 'text-amber-600 font-medium' : ''}>
                      Review: {new Date(plan.next_review_date).toLocaleDateString('en-GB')}
                    </span>
                  )}
                </div>
                {plan.approved_at && (
                  <p className="text-xs text-green-600 mt-2">
                    Approved {new Date(plan.approved_at).toLocaleDateString('en-GB')}
                  </p>
                )}
              </div>
            ))}
            {plans.length === 0 && (
              <div className="col-span-full text-center py-12 text-gray-500">
                No continuity plans found. Create one to document your recovery procedures.
              </div>
            )}
          </div>
        </div>
      ) : tab === 'exercises' ? (
        <div>
          <div className="card overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="table-header">
                  <th className="px-4 py-3 text-left">Ref</th>
                  <th className="px-4 py-3 text-left">Exercise</th>
                  <th className="px-4 py-3 text-left">Type</th>
                  <th className="px-4 py-3 text-left">Status</th>
                  <th className="px-4 py-3 text-left">Scheduled</th>
                  <th className="px-4 py-3 text-right">RTO Achieved</th>
                  <th className="px-4 py-3 text-right">RPO Achieved</th>
                  <th className="px-4 py-3 text-left">Rating</th>
                  <th className="px-4 py-3 text-left">Objectives Met</th>
                </tr>
              </thead>
              <tbody>
                {exercises.map((ex) => (
                  <tr key={ex.id} className="border-t border-gray-100 hover:bg-gray-50">
                    <td className="px-4 py-3 font-mono text-xs text-indigo-600">{ex.exercise_ref}</td>
                    <td className="px-4 py-3">
                      <p className="font-medium text-gray-900">{ex.name}</p>
                      {ex.lessons_learned && (
                        <p className="text-xs text-gray-400 mt-0.5 truncate max-w-xs">{ex.lessons_learned}</p>
                      )}
                    </td>
                    <td className="px-4 py-3 text-gray-600 text-xs capitalize">{ex.exercise_type.replace(/_/g, ' ')}</td>
                    <td className="px-4 py-3">
                      <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${EXERCISE_STATUS_COLORS[ex.status] || 'bg-gray-100'}`}>
                        {ex.status.replace(/_/g, ' ')}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-gray-500 text-xs">
                      {ex.scheduled_date
                        ? new Date(ex.scheduled_date).toLocaleDateString('en-GB')
                        : '--'}
                    </td>
                    <td className="px-4 py-3 text-right font-mono">
                      {ex.rto_achieved_hours != null ? `${ex.rto_achieved_hours.toFixed(1)}h` : '--'}
                    </td>
                    <td className="px-4 py-3 text-right font-mono">
                      {ex.rpo_achieved_hours != null ? `${ex.rpo_achieved_hours.toFixed(1)}h` : '--'}
                    </td>
                    <td className="px-4 py-3">
                      {ex.overall_rating ? (
                        <span className={`text-sm font-medium ${RATING_COLORS[ex.overall_rating] || ''}`}>
                          {ex.overall_rating.replace(/_/g, ' ')}
                        </span>
                      ) : (
                        <span className="text-gray-400">--</span>
                      )}
                    </td>
                    <td className="px-4 py-3">
                      {ex.objectives_met === true ? (
                        <span className="text-green-600 text-sm font-medium">Yes</span>
                      ) : ex.objectives_met === false ? (
                        <span className="text-red-600 text-sm font-medium">No</span>
                      ) : (
                        <span className="text-gray-400">--</span>
                      )}
                    </td>
                  </tr>
                ))}
                {exercises.length === 0 && (
                  <tr>
                    <td colSpan={9} className="px-4 py-12 text-center text-gray-500">
                      No exercises scheduled. Schedule a tabletop or simulation exercise to test your plans.
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </div>
      ) : null}
    </div>
  );
}

function LikelihoodBadge({ likelihood }: { likelihood: string }) {
  const colors: Record<string, string> = {
    almost_certain: 'bg-red-100 text-red-800',
    likely: 'bg-orange-100 text-orange-800',
    possible: 'bg-yellow-100 text-yellow-800',
    unlikely: 'bg-blue-100 text-blue-800',
    rare: 'bg-gray-100 text-gray-600',
  };
  return (
    <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${colors[likelihood] || 'bg-gray-100'}`}>
      {likelihood.replace(/_/g, ' ')}
    </span>
  );
}
