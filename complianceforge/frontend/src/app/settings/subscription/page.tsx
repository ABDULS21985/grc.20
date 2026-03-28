'use client';

import { useEffect, useState, useCallback } from 'react';
import api from '@/lib/api';

// ============================================================
// Subscription Management Page
// ============================================================

interface Plan {
  id: string;
  name: string;
  slug: string;
  description: string;
  tier: string;
  pricing_eur_monthly: number;
  pricing_eur_annual: number;
  max_users: number;
  max_frameworks: number;
  max_risks: number;
  max_vendors: number;
  max_storage_gb: number;
  features: Record<string, boolean>;
  monthly_savings: number;
}

interface Subscription {
  id: string;
  plan_name: string;
  status: string;
  billing_cycle: string;
  current_period_start: string;
  current_period_end: string;
  trial_ends_at: string | null;
  cancelled_at: string | null;
  cancel_reason: string;
  max_users: number;
  max_frameworks: number;
  plan: Plan | null;
}

interface ResourceUsage {
  current: number;
  max: number;
  remaining: number;
  at_limit: boolean;
}

interface UsageSummary {
  users: ResourceUsage;
  frameworks: ResourceUsage;
  risks: ResourceUsage;
  vendors: ResourceUsage;
  recent_events: UsageEvent[];
}

interface UsageEvent {
  id: string;
  event_type: string;
  quantity: number;
  created_at: string;
}

export default function SubscriptionPage() {
  const [sub, setSub] = useState<Subscription | null>(null);
  const [usage, setUsage] = useState<UsageSummary | null>(null);
  const [plans, setPlans] = useState<Plan[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [actionLoading, setActionLoading] = useState(false);
  const [showCancelModal, setShowCancelModal] = useState(false);
  const [cancelReason, setCancelReason] = useState('');
  const [billingToggle, setBillingToggle] = useState<'monthly' | 'annual'>('annual');

  const loadData = useCallback(async () => {
    try {
      const [subRes, plansRes] = await Promise.all([
        api.get<{ data: { subscription: Subscription; usage: UsageSummary } }>('/subscription'),
        api.get<{ data: Plan[] }>('/subscription/plans'),
      ]);
      setSub(subRes.data.subscription);
      setUsage(subRes.data.usage);
      setPlans(plansRes.data || []);
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Failed to load subscription data';
      setError(message);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { loadData(); }, [loadData]);

  const changePlan = async (planSlug: string) => {
    setActionLoading(true);
    setError(null);
    try {
      await api.put('/subscription/plan', { body: { plan_slug: planSlug } });
      await loadData();
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Failed to change plan';
      setError(message);
    } finally {
      setActionLoading(false);
    }
  };

  const cancelSubscription = async () => {
    setActionLoading(true);
    setError(null);
    try {
      await api.post('/subscription/cancel', { body: { reason: cancelReason } });
      setShowCancelModal(false);
      setCancelReason('');
      await loadData();
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Failed to cancel subscription';
      setError(message);
    } finally {
      setActionLoading(false);
    }
  };

  const pauseSubscription = async () => {
    setActionLoading(true);
    try {
      await api.post('/subscription/pause');
      await loadData();
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Failed to pause subscription';
      setError(message);
    } finally {
      setActionLoading(false);
    }
  };

  const resumeSubscription = async () => {
    setActionLoading(true);
    try {
      await api.post('/subscription/resume');
      await loadData();
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Failed to resume subscription';
      setError(message);
    } finally {
      setActionLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <p className="text-gray-500">Loading subscription...</p>
      </div>
    );
  }

  const statusColors: Record<string, string> = {
    active: 'bg-green-100 text-green-700',
    trialing: 'bg-blue-100 text-blue-700',
    past_due: 'bg-amber-100 text-amber-700',
    paused: 'bg-gray-100 text-gray-600',
    cancelled: 'bg-red-100 text-red-700',
  };

  return (
    <div>
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-gray-900">Subscription Management</h1>
        <p className="text-gray-500 mt-1">Manage your plan, billing, and usage limits</p>
      </div>

      {error && (
        <div className="mb-6 rounded-lg bg-red-50 border border-red-200 p-3 text-sm text-red-700">
          {error}
          <button onClick={() => setError(null)} className="ml-2 text-red-500 hover:text-red-700 font-medium">Dismiss</button>
        </div>
      )}

      {/* Current Subscription */}
      {sub && (
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 mb-6">
          <div className="flex items-start justify-between">
            <div>
              <div className="flex items-center gap-3">
                <h2 className="text-lg font-semibold text-gray-900">{sub.plan?.name || sub.plan_name} Plan</h2>
                <span className={`text-xs px-2.5 py-1 rounded-full font-medium ${statusColors[sub.status] || 'bg-gray-100 text-gray-600'}`}>
                  {sub.status.replace('_', ' ')}
                </span>
              </div>
              <p className="text-sm text-gray-500 mt-1">{sub.plan?.description}</p>
              <div className="flex items-center gap-4 mt-3 text-sm text-gray-600">
                <span>Billing: <strong className="text-gray-900">{sub.billing_cycle}</strong></span>
                {sub.current_period_end && (
                  <span>Renews: <strong className="text-gray-900">{new Date(sub.current_period_end).toLocaleDateString()}</strong></span>
                )}
                {sub.trial_ends_at && sub.status === 'trialing' && (
                  <span className="text-blue-600">Trial ends: <strong>{new Date(sub.trial_ends_at).toLocaleDateString()}</strong></span>
                )}
              </div>
            </div>
            <div className="text-right">
              {sub.plan && (
                <div className="text-2xl font-bold text-gray-900">
                  &euro;{sub.billing_cycle === 'annual' ? sub.plan.pricing_eur_annual.toLocaleString() : sub.plan.pricing_eur_monthly.toLocaleString()}
                  <span className="text-sm font-normal text-gray-500">/{sub.billing_cycle === 'annual' ? 'yr' : 'mo'}</span>
                </div>
              )}
            </div>
          </div>

          {/* Action Buttons */}
          <div className="flex items-center gap-3 mt-4 pt-4 border-t border-gray-100">
            {sub.status === 'active' && (
              <>
                <button onClick={pauseSubscription} disabled={actionLoading}
                  className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 disabled:opacity-50">
                  Pause Subscription
                </button>
                <button onClick={() => setShowCancelModal(true)} disabled={actionLoading}
                  className="px-4 py-2 text-sm font-medium text-red-600 bg-white border border-red-200 rounded-lg hover:bg-red-50 disabled:opacity-50">
                  Cancel Subscription
                </button>
              </>
            )}
            {sub.status === 'paused' && (
              <button onClick={resumeSubscription} disabled={actionLoading}
                className="px-4 py-2 text-sm font-medium text-white bg-green-600 rounded-lg hover:bg-green-700 disabled:opacity-50">
                Resume Subscription
              </button>
            )}
            {sub.cancelled_at && (
              <div className="text-sm text-gray-500">
                Cancelled on {new Date(sub.cancelled_at).toLocaleDateString()}
                {sub.cancel_reason && <span> &mdash; Reason: {sub.cancel_reason}</span>}
              </div>
            )}
          </div>
        </div>
      )}

      {/* Usage Summary */}
      {usage && (
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 mb-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Current Usage</h2>
          <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
            <UsageCard label="Users" usage={usage.users} />
            <UsageCard label="Frameworks" usage={usage.frameworks} />
            <UsageCard label="Active Risks" usage={usage.risks} />
            <UsageCard label="Vendors" usage={usage.vendors} />
          </div>
        </div>
      )}

      {/* Available Plans */}
      <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 mb-6">
        <div className="flex items-center justify-between mb-6">
          <h2 className="text-lg font-semibold text-gray-900">Available Plans</h2>
          <div className="flex items-center bg-gray-100 rounded-lg p-1">
            <button
              onClick={() => setBillingToggle('monthly')}
              className={`px-3 py-1.5 text-sm rounded-md font-medium transition-colors ${
                billingToggle === 'monthly' ? 'bg-white text-gray-900 shadow-sm' : 'text-gray-500'
              }`}
            >
              Monthly
            </button>
            <button
              onClick={() => setBillingToggle('annual')}
              className={`px-3 py-1.5 text-sm rounded-md font-medium transition-colors ${
                billingToggle === 'annual' ? 'bg-white text-gray-900 shadow-sm' : 'text-gray-500'
              }`}
            >
              Annual <span className="text-green-600 text-xs ml-1">Save ~17%</span>
            </button>
          </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          {plans.map(plan => {
            const isCurrent = sub?.plan_name === plan.slug;
            const price = billingToggle === 'annual' ? plan.pricing_eur_annual : plan.pricing_eur_monthly;
            const period = billingToggle === 'annual' ? '/year' : '/month';

            return (
              <div key={plan.id} className={`rounded-xl border-2 p-5 transition-colors ${
                isCurrent ? 'border-indigo-400 bg-indigo-50' : 'border-gray-200 hover:border-gray-300'
              }`}>
                <div className="flex items-center justify-between mb-3">
                  <h3 className="text-lg font-bold text-gray-900">{plan.name}</h3>
                  {isCurrent && (
                    <span className="text-xs px-2 py-0.5 bg-indigo-100 text-indigo-700 rounded-full font-medium">Current</span>
                  )}
                </div>
                <div className="mb-4">
                  <span className="text-3xl font-bold text-gray-900">&euro;{price.toLocaleString()}</span>
                  <span className="text-sm text-gray-500">{period}</span>
                  {billingToggle === 'annual' && (
                    <div className="text-xs text-green-600 mt-1">
                      Save &euro;{((plan.pricing_eur_monthly * 12) - plan.pricing_eur_annual).toFixed(0)}/year
                    </div>
                  )}
                </div>
                <p className="text-sm text-gray-500 mb-4">{plan.description}</p>

                <ul className="space-y-2 mb-4 text-sm">
                  <PlanFeature label={`Up to ${plan.max_users} users`} />
                  <PlanFeature label={`${plan.max_frameworks} frameworks`} />
                  <PlanFeature label={`${plan.max_risks.toLocaleString()} active risks`} />
                  <PlanFeature label={`${plan.max_vendors} vendors`} />
                  <PlanFeature label={`${plan.max_storage_gb} GB storage`} />
                  {plan.features.sso && <PlanFeature label="SSO / SAML" />}
                  {plan.features.api_access && <PlanFeature label="API Access" />}
                  {plan.features.advanced_reporting && <PlanFeature label="Advanced Reporting" />}
                  {plan.features.custom_branding && <PlanFeature label="Custom Branding" />}
                  {plan.features.dedicated_csm && <PlanFeature label="Dedicated CSM" />}
                </ul>

                {isCurrent ? (
                  <div className="text-center text-sm text-indigo-600 font-medium py-2">Your current plan</div>
                ) : (
                  <button
                    onClick={() => changePlan(plan.slug)}
                    disabled={actionLoading}
                    className={`w-full py-2 rounded-lg text-sm font-medium transition-colors disabled:opacity-50 ${
                      plan.tier === 'enterprise'
                        ? 'bg-indigo-600 text-white hover:bg-indigo-700'
                        : 'bg-white text-gray-700 border border-gray-300 hover:bg-gray-50'
                    }`}
                  >
                    {actionLoading ? 'Updating...' : isUpgrade(sub?.plan_name, plan.slug) ? 'Upgrade' : 'Switch'}
                  </button>
                )}
              </div>
            );
          })}
        </div>
      </div>

      {/* Recent Activity */}
      {usage && usage.recent_events && usage.recent_events.length > 0 && (
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Recent Activity</h2>
          <div className="divide-y divide-gray-100">
            {usage.recent_events.slice(0, 10).map(evt => (
              <div key={evt.id} className="flex items-center justify-between py-3">
                <div>
                  <span className="text-sm text-gray-900">{formatEventType(evt.event_type)}</span>
                  {evt.quantity > 1 && <span className="text-xs text-gray-500 ml-2">(x{evt.quantity})</span>}
                </div>
                <span className="text-xs text-gray-500">{new Date(evt.created_at).toLocaleString()}</span>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Cancel Modal */}
      {showCancelModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl shadow-lg p-6 w-full max-w-md mx-4">
            <h3 className="text-lg font-semibold text-gray-900 mb-2">Cancel Subscription</h3>
            <p className="text-sm text-gray-500 mb-4">
              Are you sure? Your data will be retained but access will be limited after the current billing period ends.
            </p>
            <textarea
              value={cancelReason}
              onChange={e => setCancelReason(e.target.value)}
              placeholder="Help us improve - why are you cancelling? (optional)"
              rows={3}
              className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm mb-4 focus:border-red-500 focus:ring-1 focus:ring-red-500"
            />
            <div className="flex justify-end gap-3">
              <button onClick={() => setShowCancelModal(false)}
                className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50">
                Keep Subscription
              </button>
              <button onClick={cancelSubscription} disabled={actionLoading}
                className="px-4 py-2 text-sm font-medium text-white bg-red-600 rounded-lg hover:bg-red-700 disabled:opacity-50">
                {actionLoading ? 'Cancelling...' : 'Confirm Cancellation'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

// ============================================================
// HELPER COMPONENTS
// ============================================================

function UsageCard({ label, usage }: { label: string; usage: ResourceUsage }) {
  const pct = usage.max > 0 ? Math.round((usage.current / usage.max) * 100) : 0;
  const barColor = pct >= 90 ? 'bg-red-500' : pct >= 70 ? 'bg-amber-500' : 'bg-green-500';

  return (
    <div className="p-4 rounded-lg border border-gray-200">
      <div className="flex items-center justify-between mb-2">
        <span className="text-sm font-medium text-gray-700">{label}</span>
        <span className={`text-xs font-medium ${usage.at_limit ? 'text-red-600' : 'text-gray-500'}`}>
          {usage.current}/{usage.max}
        </span>
      </div>
      <div className="h-2 bg-gray-100 rounded-full overflow-hidden">
        <div className={`h-full rounded-full transition-all ${barColor}`} style={{ width: `${Math.min(pct, 100)}%` }} />
      </div>
      {usage.at_limit && (
        <p className="text-xs text-red-500 mt-1 font-medium">Limit reached - upgrade to add more</p>
      )}
    </div>
  );
}

function PlanFeature({ label }: { label: string }) {
  return (
    <li className="flex items-center gap-2 text-gray-600">
      <svg className="w-4 h-4 text-green-500 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
        <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
      </svg>
      {label}
    </li>
  );
}

// ============================================================
// HELPER FUNCTIONS
// ============================================================

function isUpgrade(currentSlug: string | undefined, targetSlug: string): boolean {
  const order: Record<string, number> = { starter: 1, professional: 2, enterprise: 3 };
  return (order[targetSlug] || 0) > (order[currentSlug || ''] || 0);
}

function formatEventType(eventType: string): string {
  const labels: Record<string, string> = {
    plan_change: 'Plan changed',
    subscription_cancelled: 'Subscription cancelled',
    subscription_paused: 'Subscription paused',
    subscription_resumed: 'Subscription resumed',
    user_created: 'User created',
    framework_adopted: 'Framework adopted',
    risk_created: 'Risk created',
    vendor_added: 'Vendor added',
  };
  return labels[eventType] || eventType.replace(/_/g, ' ');
}
