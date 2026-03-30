'use client';

import { useEffect, useState, useCallback } from 'react';
import api from '@/lib/api';

// ============================================================
// Webhook Management Page
// List subscriptions with health status, create new subscriptions,
// view delivery logs, and replay failed deliveries.
// ============================================================

interface WebhookSubscription {
  id: string;
  organization_id: string;
  url: string;
  description: string;
  secret?: string;
  events: string[];
  status: string;
  version: string;
  failure_count: number;
  max_retries: number;
  last_triggered_at: string | null;
  last_success_at: string | null;
  last_failure_at: string | null;
  last_failure_reason: string | null;
  created_at: string;
  updated_at: string;
}

interface WebhookDelivery {
  id: string;
  subscription_id: string;
  event_type: string;
  payload: unknown;
  response_status: number | null;
  status: string;
  attempt_count: number;
  max_attempts: number;
  duration_ms: number | null;
  error_message: string | null;
  created_at: string;
  completed_at: string | null;
}

interface EventTypeInfo {
  event_type: string;
  category: string;
  description: string;
  version: string;
}

interface Pagination {
  page: number;
  page_size: number;
  total_items: number;
  total_pages: number;
  has_next: boolean;
  has_prev: boolean;
}

const STATUS_COLORS: Record<string, string> = {
  active: 'bg-emerald-100 text-emerald-700',
  paused: 'bg-amber-100 text-amber-700',
  disabled: 'bg-gray-100 text-gray-600',
  error: 'bg-red-100 text-red-700',
};

const DELIVERY_STATUS_COLORS: Record<string, string> = {
  pending: 'bg-blue-100 text-blue-700',
  success: 'bg-emerald-100 text-emerald-700',
  failed: 'bg-red-100 text-red-700',
  retrying: 'bg-amber-100 text-amber-700',
};

function formatDate(dateStr: string | null): string {
  if (!dateStr) return 'Never';
  return new Date(dateStr).toLocaleDateString('en-GB', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

function getHealthStatus(sub: WebhookSubscription): { label: string; color: string } {
  if (sub.status !== 'active') {
    return { label: sub.status.charAt(0).toUpperCase() + sub.status.slice(1), color: STATUS_COLORS[sub.status] || STATUS_COLORS.disabled };
  }
  if (sub.failure_count >= 5) {
    return { label: 'Unhealthy', color: 'bg-red-100 text-red-700' };
  }
  if (sub.failure_count > 0) {
    return { label: 'Degraded', color: 'bg-amber-100 text-amber-700' };
  }
  return { label: 'Healthy', color: 'bg-emerald-100 text-emerald-700' };
}

export default function WebhooksPage() {
  const [subscriptions, setSubscriptions] = useState<WebhookSubscription[]>([]);
  const [loading, setLoading] = useState(true);
  const [showCreate, setShowCreate] = useState(false);
  const [eventTypes, setEventTypes] = useState<EventTypeInfo[]>([]);
  const [newWebhookSecret, setNewWebhookSecret] = useState<string | null>(null);
  const [secretCopied, setSecretCopied] = useState(false);

  // Create form
  const [formUrl, setFormUrl] = useState('');
  const [formDescription, setFormDescription] = useState('');
  const [formEvents, setFormEvents] = useState<string[]>([]);
  const [formMaxRetries, setFormMaxRetries] = useState(5);
  const [createError, setCreateError] = useState('');
  const [eventSearch, setEventSearch] = useState('');

  // Delivery log
  const [selectedSubId, setSelectedSubId] = useState<string | null>(null);
  const [deliveries, setDeliveries] = useState<WebhookDelivery[]>([]);
  const [deliveryPagination, setDeliveryPagination] = useState<Pagination | null>(null);
  const [deliveryPage, setDeliveryPage] = useState(1);
  const [deliveryLoading, setDeliveryLoading] = useState(false);

  const loadSubscriptions = useCallback(() => {
    setLoading(true);
    api.get('/developer/webhooks')
      .then(res => setSubscriptions(res.data || []))
      .catch(() => setSubscriptions([]))
      .finally(() => setLoading(false));
  }, []);

  const loadEventTypes = useCallback(() => {
    api.get('/developer/events')
      .then(res => setEventTypes(res.data || []))
      .catch(() => setEventTypes([]));
  }, []);

  useEffect(() => { loadSubscriptions(); loadEventTypes(); }, [loadSubscriptions, loadEventTypes]);

  const loadDeliveries = useCallback((subId: string, page: number) => {
    setDeliveryLoading(true);
    api.get(`/developer/webhooks/${subId}/deliveries?page=${page}&page_size=10`)
      .then(res => {
        setDeliveries(res.data || []);
        setDeliveryPagination(res.pagination || null);
      })
      .catch(() => { setDeliveries([]); setDeliveryPagination(null); })
      .finally(() => setDeliveryLoading(false));
  }, []);

  const handleViewDeliveries = (subId: string) => {
    if (selectedSubId === subId) {
      setSelectedSubId(null);
      setDeliveries([]);
      return;
    }
    setSelectedSubId(subId);
    setDeliveryPage(1);
    loadDeliveries(subId, 1);
  };

  const handleDeliveryPageChange = (newPage: number) => {
    setDeliveryPage(newPage);
    if (selectedSubId) {
      loadDeliveries(selectedSubId, newPage);
    }
  };

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    setCreateError('');

    if (!formUrl.startsWith('https://')) {
      setCreateError('Webhook URL must use HTTPS');
      return;
    }
    if (formEvents.length === 0) {
      setCreateError('Select at least one event type');
      return;
    }

    try {
      const res = await api.post('/developer/webhooks', {
        url: formUrl,
        description: formDescription,
        events: formEvents,
        max_retries: formMaxRetries,
      });
      // Show the secret once
      if (res.data?.secret) {
        setNewWebhookSecret(res.data.secret);
      }
      setShowCreate(false);
      setFormUrl('');
      setFormDescription('');
      setFormEvents([]);
      loadSubscriptions();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to create webhook';
      setCreateError(msg);
    }
  };

  const handleTest = async (subId: string) => {
    try {
      await api.post(`/developer/webhooks/${subId}/test`, {});
      alert('Ping event queued for delivery');
    } catch {
      alert('Failed to send ping');
    }
  };

  const handleDelete = async (subId: string) => {
    if (!confirm('Are you sure you want to delete this webhook subscription?')) return;
    try {
      await api.delete(`/developer/webhooks/${subId}`);
      loadSubscriptions();
      if (selectedSubId === subId) {
        setSelectedSubId(null);
        setDeliveries([]);
      }
    } catch {
      // handled silently
    }
  };

  const handleReplay = async (deliveryId: string) => {
    try {
      await api.post(`/developer/webhooks/deliveries/${deliveryId}/replay`, {});
      if (selectedSubId) {
        loadDeliveries(selectedSubId, deliveryPage);
      }
    } catch {
      alert('Failed to replay delivery');
    }
  };

  const toggleEvent = (eventType: string) => {
    setFormEvents(prev =>
      prev.includes(eventType) ? prev.filter(e => e !== eventType) : [...prev, eventType]
    );
  };

  const toggleAllCategory = (category: string) => {
    const categoryEvents = eventTypes.filter(e => e.category === category).map(e => e.event_type);
    const allSelected = categoryEvents.every(e => formEvents.includes(e));
    if (allSelected) {
      setFormEvents(prev => prev.filter(e => !categoryEvents.includes(e)));
    } else {
      setFormEvents(prev => [...new Set([...prev, ...categoryEvents])]);
    }
  };

  const copySecret = (text: string) => {
    navigator.clipboard.writeText(text).then(() => {
      setSecretCopied(true);
      setTimeout(() => setSecretCopied(false), 2000);
    });
  };

  // Group event types by category
  const eventsByCategory = eventTypes.reduce<Record<string, EventTypeInfo[]>>((acc, e) => {
    if (!acc[e.category]) acc[e.category] = [];
    acc[e.category].push(e);
    return acc;
  }, {});

  const filteredCategories = Object.entries(eventsByCategory).filter(([category, events]) => {
    if (!eventSearch) return true;
    const q = eventSearch.toLowerCase();
    return category.toLowerCase().includes(q) ||
      events.some(e => e.event_type.toLowerCase().includes(q) || e.description.toLowerCase().includes(q));
  });

  return (
    <div>
      {/* Header */}
      <div className="flex items-center justify-between mb-8">
        <div>
          <div className="flex items-center gap-2 mb-1">
            <a href="/developer" className="text-sm text-indigo-600 hover:text-indigo-800">Developer Portal</a>
            <span className="text-gray-400">/</span>
            <span className="text-sm text-gray-600">Webhooks</span>
          </div>
          <h1 className="text-2xl font-bold text-gray-900">Webhook Subscriptions</h1>
          <p className="text-gray-500 mt-1">
            Subscribe to events and receive real-time notifications via HTTPS webhooks
          </p>
        </div>
        <button onClick={() => setShowCreate(true)} className="btn-primary text-sm">
          Create Subscription
        </button>
      </div>

      {/* New Webhook Secret Banner */}
      {newWebhookSecret && (
        <div className="mb-6 bg-emerald-50 border border-emerald-200 rounded-lg p-4">
          <div className="flex items-start justify-between">
            <div>
              <h3 className="font-semibold text-emerald-900">Webhook Created Successfully</h3>
              <p className="text-sm text-emerald-700 mt-1">
                Copy the signing secret now. It will not be shown again.
              </p>
            </div>
            <button onClick={() => setNewWebhookSecret(null)} className="text-emerald-500 hover:text-emerald-700">
              <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
          <div className="mt-3 flex items-center gap-2">
            <code className="flex-1 bg-white border border-emerald-300 rounded px-3 py-2 text-sm font-mono text-gray-800 break-all">
              {newWebhookSecret}
            </code>
            <button
              onClick={() => copySecret(newWebhookSecret)}
              className="btn-primary text-sm whitespace-nowrap"
            >
              {secretCopied ? 'Copied!' : 'Copy Secret'}
            </button>
          </div>
        </div>
      )}

      {/* Create Webhook Modal */}
      {showCreate && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg shadow-xl w-full max-w-2xl mx-4 p-6 max-h-[90vh] overflow-y-auto">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-lg font-semibold text-gray-900">Create Webhook Subscription</h2>
              <button onClick={() => setShowCreate(false)} className="text-gray-400 hover:text-gray-600">
                <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>

            {createError && (
              <div className="mb-4 bg-red-50 text-red-700 text-sm rounded px-3 py-2">{createError}</div>
            )}

            <form onSubmit={handleCreate} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Endpoint URL (HTTPS required)</label>
                <input
                  type="url"
                  value={formUrl}
                  onChange={e => setFormUrl(e.target.value)}
                  placeholder="https://your-server.com/webhook"
                  className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:ring-indigo-500 focus:border-indigo-500"
                  required
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
                <input
                  type="text"
                  value={formDescription}
                  onChange={e => setFormDescription(e.target.value)}
                  placeholder="e.g. SIEM integration for incidents"
                  className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:ring-indigo-500 focus:border-indigo-500"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Max Retries</label>
                <input
                  type="number"
                  value={formMaxRetries}
                  onChange={e => setFormMaxRetries(Number(e.target.value))}
                  min={1}
                  max={10}
                  className="w-32 border border-gray-300 rounded-md px-3 py-2 text-sm focus:ring-indigo-500 focus:border-indigo-500"
                />
              </div>

              {/* Event Selection */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Events ({formEvents.length} selected)
                </label>
                <input
                  type="text"
                  value={eventSearch}
                  onChange={e => setEventSearch(e.target.value)}
                  placeholder="Search events..."
                  className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm mb-3 focus:ring-indigo-500 focus:border-indigo-500"
                />
                <div className="border border-gray-200 rounded-lg max-h-60 overflow-y-auto">
                  {filteredCategories.map(([category, events]) => (
                    <div key={category} className="border-b border-gray-100 last:border-b-0">
                      <div
                        className="flex items-center justify-between px-3 py-2 bg-gray-50 cursor-pointer hover:bg-gray-100"
                        onClick={() => toggleAllCategory(category)}
                      >
                        <span className="text-xs font-semibold text-gray-600 uppercase tracking-wider">{category}</span>
                        <span className="text-xs text-gray-400">
                          {events.filter(e => formEvents.includes(e.event_type)).length}/{events.length}
                        </span>
                      </div>
                      <div className="px-3 py-1">
                        {events
                          .filter(e => !eventSearch || e.event_type.toLowerCase().includes(eventSearch.toLowerCase()) || e.description.toLowerCase().includes(eventSearch.toLowerCase()))
                          .map(event => (
                            <label key={event.event_type} className="flex items-center gap-2 py-1 cursor-pointer">
                              <input
                                type="checkbox"
                                checked={formEvents.includes(event.event_type)}
                                onChange={() => toggleEvent(event.event_type)}
                                className="rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
                              />
                              <span className="text-sm font-mono text-gray-700">{event.event_type}</span>
                              <span className="text-xs text-gray-400 ml-auto">{event.description}</span>
                            </label>
                          ))}
                      </div>
                    </div>
                  ))}
                </div>
              </div>

              <div className="flex justify-end gap-3 pt-2">
                <button type="button" onClick={() => setShowCreate(false)} className="btn-secondary text-sm">
                  Cancel
                </button>
                <button type="submit" className="btn-primary text-sm">
                  Create Subscription
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Subscription List */}
      {loading && (
        <div className="space-y-3">
          {[1, 2, 3].map(i => (
            <div key={i} className="card animate-pulse">
              <div className="h-4 bg-gray-200 rounded w-1/3 mb-2" />
              <div className="h-3 bg-gray-200 rounded w-2/3" />
            </div>
          ))}
        </div>
      )}

      {!loading && subscriptions.length === 0 && (
        <div className="card text-center py-12">
          <svg className="mx-auto h-10 w-10 text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
          </svg>
          <h3 className="mt-3 text-gray-900 font-medium">No Webhook Subscriptions</h3>
          <p className="text-sm text-gray-500 mt-1">Create your first webhook to receive real-time event notifications.</p>
          <button onClick={() => setShowCreate(true)} className="mt-4 btn-primary text-sm">
            Create Subscription
          </button>
        </div>
      )}

      {!loading && subscriptions.map(sub => {
        const health = getHealthStatus(sub);

        return (
          <div key={sub.id} className="card mb-3">
            {/* Subscription Header */}
            <div className="flex items-start justify-between">
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2 mb-1">
                  <span className={`text-xs font-medium px-2 py-0.5 rounded-full ${health.color}`}>
                    {health.label}
                  </span>
                  <span className={`text-xs px-2 py-0.5 rounded-full ${STATUS_COLORS[sub.status] || STATUS_COLORS.disabled}`}>
                    {sub.status}
                  </span>
                </div>
                <h3 className="font-mono text-sm text-gray-900 truncate mb-1">{sub.url}</h3>
                {sub.description && (
                  <p className="text-sm text-gray-500 mb-2">{sub.description}</p>
                )}
                <div className="flex flex-wrap items-center gap-3 text-xs text-gray-400">
                  <span>Created {formatDate(sub.created_at)}</span>
                  <span>Last triggered {formatDate(sub.last_triggered_at)}</span>
                  <span>Last success {formatDate(sub.last_success_at)}</span>
                  {sub.failure_count > 0 && (
                    <span className="text-red-500">{sub.failure_count} consecutive failures</span>
                  )}
                </div>
                {sub.last_failure_reason && (
                  <p className="text-xs text-red-500 mt-1 truncate">
                    Last error: {sub.last_failure_reason}
                  </p>
                )}
                {/* Events */}
                <div className="flex flex-wrap gap-1 mt-2">
                  {sub.events.slice(0, 6).map(ev => (
                    <span key={ev} className="text-xs bg-indigo-50 text-indigo-600 px-1.5 py-0.5 rounded font-mono">
                      {ev}
                    </span>
                  ))}
                  {sub.events.length > 6 && (
                    <span className="text-xs text-gray-400">+{sub.events.length - 6} more</span>
                  )}
                </div>
              </div>

              <div className="flex gap-2 ml-4 flex-shrink-0">
                <button
                  onClick={() => handleTest(sub.id)}
                  className="btn-secondary text-xs"
                  title="Send a test ping"
                >
                  Ping
                </button>
                <button
                  onClick={() => handleViewDeliveries(sub.id)}
                  className="btn-secondary text-xs"
                >
                  {selectedSubId === sub.id ? 'Hide Log' : 'Deliveries'}
                </button>
                <button
                  onClick={() => handleDelete(sub.id)}
                  className="text-xs text-red-600 hover:text-red-800 border border-red-200 rounded px-3 py-1.5 hover:bg-red-50"
                >
                  Delete
                </button>
              </div>
            </div>

            {/* Delivery Log Timeline */}
            {selectedSubId === sub.id && (
              <div className="mt-4 pt-4 border-t border-gray-100">
                <h4 className="text-sm font-medium text-gray-700 mb-3">Delivery Log</h4>

                {deliveryLoading && (
                  <div className="animate-pulse space-y-2">
                    <div className="h-6 bg-gray-200 rounded w-full" />
                    <div className="h-6 bg-gray-200 rounded w-3/4" />
                    <div className="h-6 bg-gray-200 rounded w-5/6" />
                  </div>
                )}

                {!deliveryLoading && deliveries.length === 0 && (
                  <p className="text-sm text-gray-500">No deliveries yet.</p>
                )}

                {!deliveryLoading && deliveries.length > 0 && (
                  <div className="space-y-0">
                    {deliveries.map((d, idx) => (
                      <div key={d.id} className="flex items-start gap-3 relative">
                        {/* Timeline line */}
                        {idx < deliveries.length - 1 && (
                          <div className="absolute left-[7px] top-5 bottom-0 w-px bg-gray-200" />
                        )}
                        {/* Timeline dot */}
                        <div className={`w-4 h-4 rounded-full flex-shrink-0 mt-0.5 border-2 ${
                          d.status === 'success' ? 'bg-emerald-500 border-emerald-500' :
                          d.status === 'failed' ? 'bg-red-500 border-red-500' :
                          d.status === 'retrying' ? 'bg-amber-500 border-amber-500' :
                          'bg-blue-500 border-blue-500'
                        }`} />

                        <div className="flex-1 pb-4">
                          <div className="flex items-center gap-2 mb-0.5">
                            <span className={`text-xs font-medium px-1.5 py-0.5 rounded ${DELIVERY_STATUS_COLORS[d.status] || 'bg-gray-100 text-gray-600'}`}>
                              {d.status}
                            </span>
                            <span className="text-xs font-mono text-indigo-600">{d.event_type}</span>
                            {d.response_status && (
                              <span className={`text-xs font-mono ${d.response_status >= 200 && d.response_status < 300 ? 'text-emerald-600' : 'text-red-600'}`}>
                                HTTP {d.response_status}
                              </span>
                            )}
                            {d.duration_ms != null && (
                              <span className="text-xs text-gray-400">{d.duration_ms}ms</span>
                            )}
                          </div>
                          <div className="flex items-center gap-3 text-xs text-gray-400">
                            <span>{formatDate(d.created_at)}</span>
                            <span>Attempt {d.attempt_count}/{d.max_attempts}</span>
                            {d.error_message && (
                              <span className="text-red-400 truncate max-w-xs" title={d.error_message}>
                                {d.error_message}
                              </span>
                            )}
                          </div>
                          {/* Replay button for failed deliveries */}
                          {d.status === 'failed' && (
                            <button
                              onClick={() => handleReplay(d.id)}
                              className="mt-1 text-xs text-indigo-600 hover:text-indigo-800 font-medium"
                            >
                              Replay Delivery
                            </button>
                          )}
                        </div>
                      </div>
                    ))}
                  </div>
                )}

                {/* Delivery Pagination */}
                {deliveryPagination && deliveryPagination.total_pages > 1 && (
                  <div className="flex items-center justify-between mt-3 pt-3 border-t border-gray-100">
                    <button
                      onClick={() => handleDeliveryPageChange(deliveryPage - 1)}
                      disabled={!deliveryPagination.has_prev}
                      className="btn-secondary text-xs disabled:opacity-50"
                    >
                      Previous
                    </button>
                    <span className="text-xs text-gray-500">
                      Page {deliveryPagination.page} of {deliveryPagination.total_pages}
                    </span>
                    <button
                      onClick={() => handleDeliveryPageChange(deliveryPage + 1)}
                      disabled={!deliveryPagination.has_next}
                      className="btn-secondary text-xs disabled:opacity-50"
                    >
                      Next
                    </button>
                  </div>
                )}
              </div>
            )}
          </div>
        );
      })}
    </div>
  );
}
