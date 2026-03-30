'use client';

import { useEffect, useState, useCallback } from 'react';
import api from '@/lib/api';

// ============================================================
// Developer Portal — API Key Management
// Create, list, revoke API keys with usage dashboards.
// Provides "copy to clipboard" for newly created keys (shown once).
// ============================================================

interface APIKey {
  id: string;
  organization_id: string;
  name: string;
  key_prefix: string;
  key?: string; // only present on creation
  scope: string[];
  tier: string;
  rate_limit_per_minute: number;
  rate_limit_per_day: number;
  allowed_ip_ranges: string[];
  allowed_origins: string[];
  expires_at: string | null;
  last_used_at: string | null;
  last_used_ip: string | null;
  is_active: boolean;
  created_at: string;
  updated_at?: string;
}

interface UsageStats {
  key_id: string;
  period: string;
  total_requests: number;
  success_count: number;
  error_count: number;
  avg_duration_ms: number;
  top_endpoints: { method: string; path: string; count: number }[];
  requests_by_day: { date: string; count: number; errors: number }[];
  errors_by_code: Record<string, number>;
}

const TIER_LABELS: Record<string, string> = {
  standard: 'Standard',
  professional: 'Professional',
  enterprise: 'Enterprise',
};

const TIER_COLORS: Record<string, string> = {
  standard: 'bg-gray-100 text-gray-700',
  professional: 'bg-blue-100 text-blue-700',
  enterprise: 'bg-purple-100 text-purple-700',
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

export default function DeveloperPortalPage() {
  const [keys, setKeys] = useState<APIKey[]>([]);
  const [loading, setLoading] = useState(true);
  const [showCreate, setShowCreate] = useState(false);
  const [newKey, setNewKey] = useState<APIKey | null>(null);
  const [copied, setCopied] = useState(false);
  const [selectedKeyUsage, setSelectedKeyUsage] = useState<UsageStats | null>(null);
  const [usageKeyId, setUsageKeyId] = useState<string | null>(null);
  const [usagePeriod, setUsagePeriod] = useState('7d');
  const [usageLoading, setUsageLoading] = useState(false);

  // Create form state
  const [formName, setFormName] = useState('');
  const [formTier, setFormTier] = useState('standard');
  const [formScopes, setFormScopes] = useState<string[]>([]);
  const [formRateMin, setFormRateMin] = useState(60);
  const [formRateDay, setFormRateDay] = useState(10000);
  const [formExpiryDays, setFormExpiryDays] = useState(0);
  const [createError, setCreateError] = useState('');

  const loadKeys = useCallback(() => {
    setLoading(true);
    api.get('/developer/api-keys')
      .then(res => setKeys(res.data || []))
      .catch(() => setKeys([]))
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => { loadKeys(); }, [loadKeys]);

  const loadUsage = useCallback((keyId: string, period: string) => {
    setUsageLoading(true);
    api.get(`/developer/api-keys/${keyId}/usage?period=${period}`)
      .then(res => setSelectedKeyUsage(res.data || null))
      .catch(() => setSelectedKeyUsage(null))
      .finally(() => setUsageLoading(false));
  }, []);

  const handleViewUsage = (keyId: string) => {
    if (usageKeyId === keyId) {
      setUsageKeyId(null);
      setSelectedKeyUsage(null);
      return;
    }
    setUsageKeyId(keyId);
    loadUsage(keyId, usagePeriod);
  };

  const handlePeriodChange = (period: string) => {
    setUsagePeriod(period);
    if (usageKeyId) {
      loadUsage(usageKeyId, period);
    }
  };

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    setCreateError('');

    if (!formName.trim()) {
      setCreateError('Name is required');
      return;
    }

    try {
      const res = await api.post('/developer/api-keys', {
        name: formName,
        tier: formTier,
        scope: formScopes,
        rate_limit_per_minute: formRateMin,
        rate_limit_per_day: formRateDay,
        expires_in_days: formExpiryDays > 0 ? formExpiryDays : undefined,
      });
      setNewKey(res.data);
      setShowCreate(false);
      setFormName('');
      setFormScopes([]);
      loadKeys();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to create API key';
      setCreateError(msg);
    }
  };

  const handleRevoke = async (keyId: string) => {
    if (!confirm('Are you sure you want to revoke this API key? This action cannot be undone.')) {
      return;
    }
    try {
      await api.delete(`/developer/api-keys/${keyId}`);
      loadKeys();
    } catch {
      // error handled silently
    }
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text).then(() => {
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    });
  };

  const activeKeys = keys.filter(k => k.is_active);
  const revokedKeys = keys.filter(k => !k.is_active);

  return (
    <div>
      {/* Header */}
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Developer Portal</h1>
          <p className="text-gray-500 mt-1">
            Manage API keys, view usage analytics, and configure webhook integrations
          </p>
        </div>
        <div className="flex gap-3">
          <a href="/developer/webhooks" className="btn-secondary text-sm">
            Webhooks
          </a>
          <button onClick={() => setShowCreate(true)} className="btn-primary text-sm">
            Generate API Key
          </button>
        </div>
      </div>

      {/* Newly Created Key Banner — shown once */}
      {newKey?.key && (
        <div className="mb-6 bg-emerald-50 border border-emerald-200 rounded-lg p-4">
          <div className="flex items-start justify-between">
            <div>
              <h3 className="font-semibold text-emerald-900">API Key Created Successfully</h3>
              <p className="text-sm text-emerald-700 mt-1">
                Copy this key now. It will not be shown again.
              </p>
            </div>
            <button onClick={() => setNewKey(null)} className="text-emerald-500 hover:text-emerald-700">
              <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
          <div className="mt-3 flex items-center gap-2">
            <code className="flex-1 bg-white border border-emerald-300 rounded px-3 py-2 text-sm font-mono text-gray-800 break-all">
              {newKey.key}
            </code>
            <button
              onClick={() => copyToClipboard(newKey.key!)}
              className="btn-primary text-sm whitespace-nowrap"
            >
              {copied ? 'Copied!' : 'Copy to Clipboard'}
            </button>
          </div>
          <p className="text-xs text-emerald-600 mt-2">
            Name: {newKey.name} | Prefix: {newKey.key_prefix} | Tier: {newKey.tier}
          </p>
        </div>
      )}

      {/* Create API Key Modal */}
      {showCreate && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg shadow-xl w-full max-w-lg mx-4 p-6">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-lg font-semibold text-gray-900">Generate New API Key</h2>
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
                <label className="block text-sm font-medium text-gray-700 mb-1">Key Name</label>
                <input
                  type="text"
                  value={formName}
                  onChange={e => setFormName(e.target.value)}
                  placeholder="e.g. Production Integration"
                  className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:ring-indigo-500 focus:border-indigo-500"
                  required
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Tier</label>
                <select
                  value={formTier}
                  onChange={e => setFormTier(e.target.value)}
                  className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:ring-indigo-500 focus:border-indigo-500"
                >
                  <option value="standard">Standard (60 req/min)</option>
                  <option value="professional">Professional (300 req/min)</option>
                  <option value="enterprise">Enterprise (1000 req/min)</option>
                </select>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Rate Limit (per min)</label>
                  <input
                    type="number"
                    value={formRateMin}
                    onChange={e => setFormRateMin(Number(e.target.value))}
                    min={1}
                    max={10000}
                    className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:ring-indigo-500 focus:border-indigo-500"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Rate Limit (per day)</label>
                  <input
                    type="number"
                    value={formRateDay}
                    onChange={e => setFormRateDay(Number(e.target.value))}
                    min={1}
                    max={1000000}
                    className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:ring-indigo-500 focus:border-indigo-500"
                  />
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Expiry (days, 0 = never)</label>
                <input
                  type="number"
                  value={formExpiryDays}
                  onChange={e => setFormExpiryDays(Number(e.target.value))}
                  min={0}
                  max={365}
                  className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:ring-indigo-500 focus:border-indigo-500"
                />
              </div>

              <div className="flex justify-end gap-3 pt-2">
                <button type="button" onClick={() => setShowCreate(false)} className="btn-secondary text-sm">
                  Cancel
                </button>
                <button type="submit" className="btn-primary text-sm">
                  Generate Key
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Active API Keys */}
      <div className="mb-8">
        <h2 className="text-lg font-semibold text-gray-900 mb-4">
          Active API Keys ({activeKeys.length})
        </h2>

        {loading && (
          <div className="space-y-3">
            {[1, 2, 3].map(i => (
              <div key={i} className="card animate-pulse">
                <div className="h-4 bg-gray-200 rounded w-1/4 mb-2" />
                <div className="h-3 bg-gray-200 rounded w-1/2" />
              </div>
            ))}
          </div>
        )}

        {!loading && activeKeys.length === 0 && (
          <div className="card text-center py-10">
            <svg className="mx-auto h-10 w-10 text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
            </svg>
            <h3 className="mt-3 text-gray-900 font-medium">No API Keys</h3>
            <p className="text-sm text-gray-500 mt-1">Generate your first API key to get started.</p>
            <button onClick={() => setShowCreate(true)} className="mt-4 btn-primary text-sm">
              Generate API Key
            </button>
          </div>
        )}

        {!loading && activeKeys.map(key => (
          <div key={key.id} className="card mb-3">
            <div className="flex items-start justify-between">
              <div className="flex-1">
                <div className="flex items-center gap-2 mb-1">
                  <h3 className="font-semibold text-gray-900">{key.name}</h3>
                  <span className={`text-xs font-medium px-2 py-0.5 rounded-full ${TIER_COLORS[key.tier] || TIER_COLORS.standard}`}>
                    {TIER_LABELS[key.tier] || key.tier}
                  </span>
                  <span className="inline-flex items-center gap-1 text-xs text-emerald-600">
                    <span className="w-1.5 h-1.5 bg-emerald-500 rounded-full" />
                    Active
                  </span>
                </div>
                <div className="flex flex-wrap items-center gap-4 text-sm text-gray-500">
                  <span className="font-mono text-xs bg-gray-100 px-2 py-0.5 rounded">
                    {key.key_prefix}...
                  </span>
                  <span>Created {formatDate(key.created_at)}</span>
                  <span>Last used {formatDate(key.last_used_at)}</span>
                  {key.expires_at && (
                    <span className="text-amber-600">Expires {formatDate(key.expires_at)}</span>
                  )}
                </div>
                {key.scope && key.scope.length > 0 && (
                  <div className="flex flex-wrap gap-1 mt-2">
                    {key.scope.slice(0, 5).map(s => (
                      <span key={s} className="text-xs bg-indigo-50 text-indigo-600 px-1.5 py-0.5 rounded">
                        {s}
                      </span>
                    ))}
                    {key.scope.length > 5 && (
                      <span className="text-xs text-gray-400">+{key.scope.length - 5} more</span>
                    )}
                  </div>
                )}
              </div>
              <div className="flex gap-2 ml-4">
                <button
                  onClick={() => handleViewUsage(key.id)}
                  className="btn-secondary text-xs"
                >
                  {usageKeyId === key.id ? 'Hide Usage' : 'Usage'}
                </button>
                <button
                  onClick={() => handleRevoke(key.id)}
                  className="text-xs text-red-600 hover:text-red-800 border border-red-200 rounded px-3 py-1.5 hover:bg-red-50"
                >
                  Revoke
                </button>
              </div>
            </div>

            {/* Usage Dashboard (inline) */}
            {usageKeyId === key.id && (
              <div className="mt-4 pt-4 border-t border-gray-100">
                <div className="flex items-center justify-between mb-3">
                  <h4 className="text-sm font-medium text-gray-700">Usage Dashboard</h4>
                  <div className="flex gap-1">
                    {['24h', '7d', '30d', '90d'].map(p => (
                      <button
                        key={p}
                        onClick={() => handlePeriodChange(p)}
                        className={`text-xs px-2 py-1 rounded ${usagePeriod === p ? 'bg-indigo-100 text-indigo-700 font-medium' : 'text-gray-500 hover:bg-gray-100'}`}
                      >
                        {p}
                      </button>
                    ))}
                  </div>
                </div>

                {usageLoading && (
                  <div className="animate-pulse space-y-2">
                    <div className="h-8 bg-gray-200 rounded w-full" />
                    <div className="h-8 bg-gray-200 rounded w-3/4" />
                  </div>
                )}

                {!usageLoading && selectedKeyUsage && (
                  <div className="space-y-4">
                    {/* Summary Cards */}
                    <div className="grid grid-cols-4 gap-3">
                      <div className="bg-gray-50 rounded-lg px-3 py-2">
                        <p className="text-xs text-gray-500">Total Requests</p>
                        <p className="text-lg font-bold text-gray-900">
                          {selectedKeyUsage.total_requests.toLocaleString()}
                        </p>
                      </div>
                      <div className="bg-emerald-50 rounded-lg px-3 py-2">
                        <p className="text-xs text-emerald-600">Successful</p>
                        <p className="text-lg font-bold text-emerald-700">
                          {selectedKeyUsage.success_count.toLocaleString()}
                        </p>
                      </div>
                      <div className="bg-red-50 rounded-lg px-3 py-2">
                        <p className="text-xs text-red-600">Errors</p>
                        <p className="text-lg font-bold text-red-700">
                          {selectedKeyUsage.error_count.toLocaleString()}
                        </p>
                      </div>
                      <div className="bg-blue-50 rounded-lg px-3 py-2">
                        <p className="text-xs text-blue-600">Avg Latency</p>
                        <p className="text-lg font-bold text-blue-700">
                          {selectedKeyUsage.avg_duration_ms.toFixed(0)}ms
                        </p>
                      </div>
                    </div>

                    {/* Top Endpoints */}
                    {selectedKeyUsage.top_endpoints.length > 0 && (
                      <div>
                        <h5 className="text-xs font-medium text-gray-700 mb-2">Top Endpoints</h5>
                        <div className="space-y-1">
                          {selectedKeyUsage.top_endpoints.slice(0, 5).map((ep, i) => (
                            <div key={i} className="flex items-center gap-2 text-xs">
                              <span className="font-mono font-medium text-indigo-600 w-12">{ep.method}</span>
                              <span className="font-mono text-gray-600 flex-1 truncate">{ep.path}</span>
                              <span className="text-gray-500">{ep.count.toLocaleString()} req</span>
                            </div>
                          ))}
                        </div>
                      </div>
                    )}

                    {/* Daily Chart (simple bar representation) */}
                    {selectedKeyUsage.requests_by_day.length > 0 && (
                      <div>
                        <h5 className="text-xs font-medium text-gray-700 mb-2">Requests by Day</h5>
                        <div className="flex items-end gap-1 h-16">
                          {selectedKeyUsage.requests_by_day.map((day, i) => {
                            const maxCount = Math.max(...selectedKeyUsage.requests_by_day.map(d => d.count), 1);
                            const height = Math.max((day.count / maxCount) * 100, 2);
                            return (
                              <div key={i} className="flex-1 flex flex-col items-center" title={`${day.date}: ${day.count} requests`}>
                                <div
                                  className="w-full bg-indigo-400 rounded-t"
                                  style={{ height: `${height}%` }}
                                />
                              </div>
                            );
                          })}
                        </div>
                        <div className="flex justify-between text-xs text-gray-400 mt-1">
                          <span>{selectedKeyUsage.requests_by_day[0]?.date}</span>
                          <span>{selectedKeyUsage.requests_by_day[selectedKeyUsage.requests_by_day.length - 1]?.date}</span>
                        </div>
                      </div>
                    )}
                  </div>
                )}

                {!usageLoading && !selectedKeyUsage && (
                  <p className="text-sm text-gray-500">No usage data available for this period.</p>
                )}
              </div>
            )}
          </div>
        ))}
      </div>

      {/* Revoked API Keys */}
      {revokedKeys.length > 0 && (
        <div>
          <h2 className="text-lg font-semibold text-gray-900 mb-4">
            Revoked Keys ({revokedKeys.length})
          </h2>
          <div className="space-y-2">
            {revokedKeys.map(key => (
              <div key={key.id} className="card bg-gray-50 opacity-75">
                <div className="flex items-center justify-between">
                  <div>
                    <div className="flex items-center gap-2">
                      <h3 className="font-medium text-gray-600">{key.name}</h3>
                      <span className="text-xs text-red-500 bg-red-50 px-2 py-0.5 rounded-full">Revoked</span>
                    </div>
                    <div className="flex items-center gap-3 text-xs text-gray-400 mt-1">
                      <span className="font-mono">{key.key_prefix}...</span>
                      <span>Created {formatDate(key.created_at)}</span>
                    </div>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
