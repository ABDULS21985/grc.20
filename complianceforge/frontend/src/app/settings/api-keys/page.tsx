'use client';

import { useEffect, useState, useCallback } from 'react';
import api from '@/lib/api';

// ============================================================
// TYPES
// ============================================================

interface APIKey {
  id: string;
  name: string;
  key_prefix: string;
  permissions: string[];
  rate_limit_per_minute: number;
  expires_at: string | null;
  last_used_at: string | null;
  last_used_ip: string;
  is_active: boolean;
  created_at: string;
}

interface CreateAPIKeyResult extends APIKey {
  raw_key: string;
}

const PERMISSION_OPTIONS = [
  { value: '*:*', label: 'Full Access', description: 'Read and write access to all resources' },
  { value: '*:read', label: 'Read-Only (All)', description: 'Read access to all resources' },
  { value: 'risks:read', label: 'Risks: Read', description: 'Read access to risk data' },
  { value: 'risks:write', label: 'Risks: Write', description: 'Write access to risk data' },
  { value: 'integrations:read', label: 'Integrations: Read', description: 'Read access to integrations' },
  { value: 'integrations:write', label: 'Integrations: Write', description: 'Write access to integrations' },
  { value: 'compliance:read', label: 'Compliance: Read', description: 'Read access to compliance data' },
  { value: 'policies:read', label: 'Policies: Read', description: 'Read access to policies' },
  { value: 'audits:read', label: 'Audits: Read', description: 'Read access to audit data' },
  { value: 'incidents:read', label: 'Incidents: Read', description: 'Read access to incidents' },
  { value: 'incidents:write', label: 'Incidents: Write', description: 'Write access to incidents' },
  { value: 'reports:read', label: 'Reports: Read', description: 'Read access to reports' },
];

const EXPIRY_OPTIONS = [
  { value: 0, label: 'Never expires' },
  { value: 30, label: '30 days' },
  { value: 90, label: '90 days' },
  { value: 180, label: '180 days' },
  { value: 365, label: '1 year' },
];

// ============================================================
// COMPONENT
// ============================================================

export default function APIKeysPage() {
  const [keys, setKeys] = useState<APIKey[]>([]);
  const [loading, setLoading] = useState(true);
  const [showCreate, setShowCreate] = useState(false);
  const [showCreatedKey, setShowCreatedKey] = useState<CreateAPIKeyResult | null>(null);
  const [copied, setCopied] = useState(false);

  // Create form state
  const [name, setName] = useState('');
  const [permissions, setPermissions] = useState<string[]>(['*:read']);
  const [rateLimit, setRateLimit] = useState(60);
  const [expiresInDays, setExpiresInDays] = useState(0);
  const [creating, setCreating] = useState(false);

  const fetchKeys = useCallback(async () => {
    setLoading(true);
    try {
      const res = await api.getAPIKeys();
      setKeys(res.data || []);
    } catch {
      /* silent */
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { fetchKeys(); }, [fetchKeys]);

  const handleCreate = async () => {
    if (!name.trim()) return;
    setCreating(true);
    try {
      const res = await api.createAPIKey({
        name: name.trim(),
        permissions,
        rate_limit_per_minute: rateLimit,
        expires_in_days: expiresInDays,
      });
      setShowCreatedKey(res.data);
      setShowCreate(false);
      setName('');
      setPermissions(['*:read']);
      setRateLimit(60);
      setExpiresInDays(0);
      fetchKeys();
    } catch {
      /* silent */
    } finally {
      setCreating(false);
    }
  };

  const handleRevoke = async (id: string) => {
    if (!confirm('Are you sure you want to revoke this API key? This action cannot be undone.')) return;
    try {
      await api.revokeAPIKey(id);
      fetchKeys();
    } catch { /* silent */ }
  };

  const handleCopy = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch { /* silent */ }
  };

  const togglePermission = (perm: string) => {
    setPermissions(prev =>
      prev.includes(perm) ? prev.filter(p => p !== perm) : [...prev, perm]
    );
  };

  const activeKeys = keys.filter(k => k.is_active);
  const revokedKeys = keys.filter(k => !k.is_active);

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">API Keys</h1>
          <p className="text-sm text-gray-500 mt-1">Manage programmatic access to the ComplianceForge API</p>
        </div>
        <button onClick={() => setShowCreate(true)} className="btn-primary">+ Create API Key</button>
      </div>

      {/* Warning Banner */}
      <div className="bg-amber-50 border border-amber-200 rounded-lg p-4 mb-6">
        <div className="flex items-start gap-3">
          <span className="text-amber-600 text-lg">!</span>
          <div>
            <p className="text-sm font-medium text-amber-800">API keys grant programmatic access to your organisation's data</p>
            <p className="text-xs text-amber-600 mt-1">
              Treat API keys like passwords. Do not share them in public repositories, client-side code, or insecure channels.
              Keys are shown only once upon creation and cannot be retrieved later.
            </p>
          </div>
        </div>
      </div>

      {loading ? (
        <p className="text-gray-500">Loading API keys...</p>
      ) : (
        <>
          {/* Active Keys */}
          <h2 className="text-lg font-semibold text-gray-900 mb-3">Active Keys ({activeKeys.length})</h2>
          {activeKeys.length === 0 ? (
            <div className="card text-center py-8 mb-8">
              <p className="text-gray-500">No active API keys</p>
              <p className="text-sm text-gray-400 mt-1">Create your first key to enable programmatic access</p>
            </div>
          ) : (
            <div className="space-y-3 mb-8">
              {activeKeys.map(key => (
                <div key={key.id} className="card">
                  <div className="flex items-center justify-between">
                    <div className="flex-1">
                      <div className="flex items-center gap-3">
                        <h3 className="font-medium text-gray-900">{key.name}</h3>
                        <code className="text-xs bg-gray-100 px-2 py-0.5 rounded font-mono text-gray-600">
                          {key.key_prefix}...
                        </code>
                      </div>
                      <div className="flex items-center gap-4 mt-2 text-xs text-gray-500">
                        <span>Created: {new Date(key.created_at).toLocaleDateString('en-GB')}</span>
                        {key.expires_at && (
                          <span className={new Date(key.expires_at) < new Date() ? 'text-red-500 font-medium' : ''}>
                            Expires: {new Date(key.expires_at).toLocaleDateString('en-GB')}
                          </span>
                        )}
                        {key.last_used_at ? (
                          <span>Last used: {new Date(key.last_used_at).toLocaleString('en-GB', { dateStyle: 'short', timeStyle: 'short' })}</span>
                        ) : (
                          <span className="text-gray-400">Never used</span>
                        )}
                        {key.last_used_ip && <span>IP: {key.last_used_ip}</span>}
                        <span>Rate limit: {key.rate_limit_per_minute}/min</span>
                      </div>
                      <div className="flex flex-wrap gap-1 mt-2">
                        {key.permissions.map(perm => (
                          <span key={perm} className="text-xs bg-indigo-50 text-indigo-600 px-2 py-0.5 rounded font-mono">{perm}</span>
                        ))}
                      </div>
                    </div>
                    <button
                      onClick={() => handleRevoke(key.id)}
                      className="text-sm text-red-600 hover:underline ml-4"
                    >Revoke</button>
                  </div>
                </div>
              ))}
            </div>
          )}

          {/* Revoked Keys */}
          {revokedKeys.length > 0 && (
            <>
              <h2 className="text-lg font-semibold text-gray-500 mb-3">Revoked Keys ({revokedKeys.length})</h2>
              <div className="space-y-3">
                {revokedKeys.map(key => (
                  <div key={key.id} className="card opacity-50">
                    <div className="flex items-center gap-3">
                      <h3 className="font-medium text-gray-600 line-through">{key.name}</h3>
                      <code className="text-xs bg-gray-100 px-2 py-0.5 rounded font-mono text-gray-400">{key.key_prefix}...</code>
                      <span className="text-xs bg-red-100 text-red-600 px-2 py-0.5 rounded font-medium">Revoked</span>
                    </div>
                    <p className="text-xs text-gray-400 mt-1">Created: {new Date(key.created_at).toLocaleDateString('en-GB')}</p>
                  </div>
                ))}
              </div>
            </>
          )}
        </>
      )}

      {/* Create Modal */}
      {showCreate && (
        <div className="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4">
          <div className="bg-white rounded-xl shadow-2xl w-full max-w-lg max-h-[90vh] overflow-y-auto">
            <div className="p-6">
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-lg font-semibold text-gray-900">Create API Key</h2>
                <button onClick={() => setShowCreate(false)} className="text-gray-400 hover:text-gray-600 text-xl">&times;</button>
              </div>

              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Key Name</label>
                  <input
                    type="text"
                    value={name}
                    onChange={e => setName(e.target.value)}
                    className="input w-full"
                    placeholder="e.g. CI/CD Pipeline, SIEM Integration"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Permissions</label>
                  <div className="space-y-2 max-h-48 overflow-y-auto border border-gray-200 rounded-lg p-3">
                    {PERMISSION_OPTIONS.map(opt => (
                      <label key={opt.value} className="flex items-start gap-2 cursor-pointer">
                        <input
                          type="checkbox"
                          checked={permissions.includes(opt.value)}
                          onChange={() => togglePermission(opt.value)}
                          className="mt-0.5 rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
                        />
                        <div>
                          <span className="text-sm font-medium text-gray-700">{opt.label}</span>
                          <p className="text-xs text-gray-500">{opt.description}</p>
                        </div>
                      </label>
                    ))}
                  </div>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Rate Limit (requests/min)</label>
                  <input
                    type="number"
                    value={rateLimit}
                    onChange={e => setRateLimit(Math.max(1, parseInt(e.target.value) || 60))}
                    className="input w-32"
                    min={1}
                    max={1000}
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Expiry</label>
                  <select
                    value={expiresInDays}
                    onChange={e => setExpiresInDays(parseInt(e.target.value))}
                    className="input w-48"
                  >
                    {EXPIRY_OPTIONS.map(opt => (
                      <option key={opt.value} value={opt.value}>{opt.label}</option>
                    ))}
                  </select>
                </div>
              </div>

              <div className="flex justify-end gap-3 mt-6 pt-4 border-t border-gray-100">
                <button onClick={() => setShowCreate(false)} className="btn-secondary">Cancel</button>
                <button
                  onClick={handleCreate}
                  disabled={creating || !name.trim() || permissions.length === 0}
                  className="btn-primary disabled:opacity-50"
                >{creating ? 'Creating...' : 'Create API Key'}</button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Created Key Modal — show raw key once */}
      {showCreatedKey && (
        <div className="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4">
          <div className="bg-white rounded-xl shadow-2xl w-full max-w-lg">
            <div className="p-6">
              <div className="flex items-center gap-2 mb-4">
                <span className="text-green-600 text-xl">&#10003;</span>
                <h2 className="text-lg font-semibold text-gray-900">API Key Created</h2>
              </div>

              <div className="bg-amber-50 border border-amber-200 rounded-lg p-3 mb-4">
                <p className="text-sm font-medium text-amber-800">
                  Copy your API key now. It will not be shown again.
                </p>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">API Key</label>
                <div className="flex items-center gap-2">
                  <code className="flex-1 bg-gray-50 border border-gray-200 rounded-lg px-3 py-2 text-sm font-mono text-gray-900 break-all select-all">
                    {showCreatedKey.raw_key}
                  </code>
                  <button
                    onClick={() => handleCopy(showCreatedKey.raw_key)}
                    className="btn-secondary whitespace-nowrap"
                  >{copied ? 'Copied!' : 'Copy'}</button>
                </div>
              </div>

              <div className="mt-4 text-sm text-gray-600">
                <p><strong>Name:</strong> {showCreatedKey.name}</p>
                <p><strong>Prefix:</strong> {showCreatedKey.key_prefix}...</p>
                <p><strong>Permissions:</strong> {showCreatedKey.permissions.join(', ')}</p>
                {showCreatedKey.expires_at && (
                  <p><strong>Expires:</strong> {new Date(showCreatedKey.expires_at).toLocaleDateString('en-GB')}</p>
                )}
              </div>

              <div className="bg-gray-50 rounded-lg p-3 mt-4 text-xs text-gray-600">
                <p className="font-medium mb-1">Usage example:</p>
                <code className="block text-xs">
                  curl -H &quot;X-API-Key: {showCreatedKey.key_prefix}...&quot; \<br />
                  &nbsp;&nbsp;{typeof window !== 'undefined' ? window.location.origin : ''}/api/v1/risks
                </code>
              </div>

              <div className="flex justify-end mt-6 pt-4 border-t border-gray-100">
                <button
                  onClick={() => setShowCreatedKey(null)}
                  className="btn-primary"
                >I have copied the key</button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
