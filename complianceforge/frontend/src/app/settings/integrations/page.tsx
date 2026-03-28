'use client';

import { useEffect, useState, useCallback } from 'react';
import api from '@/lib/api';

// ============================================================
// TYPES
// ============================================================

interface Integration {
  id: string;
  integration_type: string;
  name: string;
  description: string;
  status: 'active' | 'inactive' | 'error' | 'pending_setup';
  health_status: 'healthy' | 'degraded' | 'unhealthy' | 'unknown';
  last_sync_at: string | null;
  sync_frequency_minutes: number;
  error_count: number;
  last_error_message: string;
  capabilities: string[];
  created_at: string;
}

interface CatalogEntry {
  type: string;
  category: string;
  name: string;
  description: string;
  icon: string;
  capabilities: string[];
  setup_fields: string[];
}

interface HealthSummary {
  total_integrations: number;
  active_count: number;
  error_count: number;
  healthy_count: number;
  unhealthy_count: number;
  degraded_count: number;
}

// ============================================================
// CATALOG (static — mirrors backend GetIntegrationCatalog)
// ============================================================

const CATALOG: CatalogEntry[] = [
  { type: 'sso_saml', category: 'identity', name: 'SAML 2.0 SSO', description: 'Single Sign-On via SAML 2.0 (Azure AD, Okta, ADFS, Google Workspace)', icon: 'shield-check', capabilities: ['sso', 'user_provisioning'], setup_fields: ['entity_id', 'sso_url', 'slo_url', 'certificate'] },
  { type: 'sso_oidc', category: 'identity', name: 'OpenID Connect SSO', description: 'Single Sign-On via OIDC (Azure AD, Okta, Auth0, Google)', icon: 'key', capabilities: ['sso', 'user_provisioning'], setup_fields: ['issuer_url', 'client_id', 'client_secret'] },
  { type: 'cloud_aws', category: 'cloud', name: 'Amazon Web Services', description: 'Import findings from AWS Security Hub, Config, and CloudTrail', icon: 'cloud', capabilities: ['compliance_sync', 'finding_import', 'asset_discovery'], setup_fields: ['access_key_id', 'secret_access_key', 'region', 'role_arn'] },
  { type: 'cloud_azure', category: 'cloud', name: 'Microsoft Azure', description: 'Import compliance data from Azure Security Center and Policy', icon: 'cloud', capabilities: ['compliance_sync', 'finding_import', 'asset_discovery'], setup_fields: ['tenant_id', 'client_id', 'client_secret', 'subscription_id'] },
  { type: 'cloud_gcp', category: 'cloud', name: 'Google Cloud Platform', description: 'Import findings from Security Command Center', icon: 'cloud', capabilities: ['compliance_sync', 'finding_import'], setup_fields: ['project_id', 'service_account_key'] },
  { type: 'siem_splunk', category: 'siem', name: 'Splunk', description: 'Query security events and create incidents from notable events', icon: 'search', capabilities: ['event_ingestion', 'incident_creation'], setup_fields: ['base_url', 'token', 'index'] },
  { type: 'siem_elastic', category: 'siem', name: 'Elastic Security', description: 'Ingest security alerts from Elastic SIEM', icon: 'search', capabilities: ['event_ingestion', 'alert_sync'], setup_fields: ['base_url', 'api_key', 'cloud_id'] },
  { type: 'siem_sentinel', category: 'siem', name: 'Microsoft Sentinel', description: 'Import incidents and alerts from Azure Sentinel', icon: 'search', capabilities: ['event_ingestion', 'incident_creation'], setup_fields: ['tenant_id', 'client_id', 'client_secret', 'workspace_id'] },
  { type: 'itsm_servicenow', category: 'itsm', name: 'ServiceNow', description: 'Bidirectional incident sync with ServiceNow ITSM', icon: 'ticket', capabilities: ['incident_sync', 'ticket_creation', 'status_sync'], setup_fields: ['instance_url', 'username', 'password'] },
  { type: 'itsm_jira', category: 'itsm', name: 'Jira', description: 'Create Jira issues from findings and sync statuses', icon: 'ticket', capabilities: ['issue_creation', 'status_sync'], setup_fields: ['base_url', 'email', 'api_token', 'project'] },
  { type: 'itsm_freshservice', category: 'itsm', name: 'Freshservice', description: 'Create and track tickets in Freshservice ITSM', icon: 'ticket', capabilities: ['ticket_creation', 'status_sync'], setup_fields: ['domain', 'api_key'] },
  { type: 'email_smtp', category: 'communication', name: 'Custom SMTP', description: 'Send notifications via your own SMTP server', icon: 'mail', capabilities: ['email_notifications'], setup_fields: ['host', 'port', 'username', 'password', 'from_address'] },
  { type: 'email_sendgrid', category: 'communication', name: 'SendGrid', description: 'Transactional emails via SendGrid', icon: 'mail', capabilities: ['email_notifications'], setup_fields: ['api_key', 'from_address', 'from_name'] },
  { type: 'slack', category: 'communication', name: 'Slack', description: 'Send alerts to Slack channels', icon: 'message-square', capabilities: ['notifications', 'alerts'], setup_fields: ['webhook_url', 'channel', 'bot_token'] },
  { type: 'teams', category: 'communication', name: 'Microsoft Teams', description: 'Send alerts to Teams channels', icon: 'message-square', capabilities: ['notifications', 'alerts'], setup_fields: ['webhook_url', 'channel'] },
  { type: 'webhook_inbound', category: 'automation', name: 'Inbound Webhook', description: 'Receive data from external systems', icon: 'arrow-down-circle', capabilities: ['data_ingestion'], setup_fields: ['secret'] },
  { type: 'webhook_outbound', category: 'automation', name: 'Outbound Webhook', description: 'Send events to external systems', icon: 'arrow-up-circle', capabilities: ['event_dispatch'], setup_fields: ['url', 'secret', 'events'] },
  { type: 'custom_api', category: 'automation', name: 'Custom API', description: 'Connect to any REST API', icon: 'code', capabilities: ['custom'], setup_fields: ['base_url', 'auth_type', 'api_key_or_token'] },
];

const CATEGORIES: { key: string; label: string }[] = [
  { key: 'all', label: 'All' },
  { key: 'identity', label: 'Identity & SSO' },
  { key: 'cloud', label: 'Cloud Providers' },
  { key: 'siem', label: 'SIEM' },
  { key: 'itsm', label: 'ITSM & Ticketing' },
  { key: 'communication', label: 'Communication' },
  { key: 'automation', label: 'Automation' },
];

// ============================================================
// COMPONENT
// ============================================================

export default function IntegrationsPage() {
  const [integrations, setIntegrations] = useState<Integration[]>([]);
  const [healthSummary, setHealthSummary] = useState<HealthSummary | null>(null);
  const [loading, setLoading] = useState(true);
  const [category, setCategory] = useState('all');
  const [view, setView] = useState<'active' | 'marketplace'>('active');
  const [showSetup, setShowSetup] = useState<CatalogEntry | null>(null);
  const [configFields, setConfigFields] = useState<Record<string, string>>({});
  const [setupName, setSetupName] = useState('');
  const [creating, setCreating] = useState(false);
  const [testingId, setTestingId] = useState<string | null>(null);
  const [syncingId, setSyncingId] = useState<string | null>(null);

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const [intRes, healthRes] = await Promise.all([
        api.getIntegrations(),
        api.getIntegrationHealthSummary(),
      ]);
      setIntegrations(intRes.data || []);
      setHealthSummary(healthRes.data || null);
    } catch {
      /* silent */
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { fetchData(); }, [fetchData]);

  const handleCreate = async () => {
    if (!showSetup || !setupName) return;
    setCreating(true);
    try {
      const config: Record<string, string> = {};
      showSetup.setup_fields.forEach(f => { if (configFields[f]) config[f] = configFields[f]; });
      await api.createIntegration({
        integration_type: showSetup.type,
        name: setupName,
        description: showSetup.description,
        configuration: config,
        capabilities: showSetup.capabilities,
      });
      setShowSetup(null);
      setConfigFields({});
      setSetupName('');
      fetchData();
    } catch {
      /* silent */
    } finally {
      setCreating(false);
    }
  };

  const handleTest = async (id: string) => {
    setTestingId(id);
    try {
      await api.testIntegration(id);
      fetchData();
    } catch { /* silent */ } finally { setTestingId(null); }
  };

  const handleSync = async (id: string) => {
    setSyncingId(id);
    try {
      await api.syncIntegration(id);
      fetchData();
    } catch { /* silent */ } finally { setSyncingId(null); }
  };

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this integration?')) return;
    try {
      await api.deleteIntegration(id);
      fetchData();
    } catch { /* silent */ }
  };

  const catalogFiltered = category === 'all'
    ? CATALOG
    : CATALOG.filter(c => c.category === category);

  const connectedTypes = new Set(integrations.map(i => i.integration_type));

  // ── Status / Health badges ──
  const statusColor: Record<string, string> = {
    active: 'bg-green-100 text-green-700',
    inactive: 'bg-gray-100 text-gray-600',
    error: 'bg-red-100 text-red-700',
    pending_setup: 'bg-yellow-100 text-yellow-700',
  };

  const healthColor: Record<string, string> = {
    healthy: 'bg-green-100 text-green-700',
    degraded: 'bg-yellow-100 text-yellow-700',
    unhealthy: 'bg-red-100 text-red-700',
    unknown: 'bg-gray-100 text-gray-600',
  };

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Integration Hub</h1>
          <p className="text-sm text-gray-500 mt-1">Connect ComplianceForge with your security stack</p>
        </div>
        <div className="flex gap-2">
          <button
            onClick={() => setView('active')}
            className={`px-4 py-2 text-sm font-medium rounded-lg transition-colors ${view === 'active' ? 'bg-indigo-600 text-white' : 'bg-white text-gray-700 border border-gray-200 hover:bg-gray-50'}`}
          >Active Integrations</button>
          <button
            onClick={() => setView('marketplace')}
            className={`px-4 py-2 text-sm font-medium rounded-lg transition-colors ${view === 'marketplace' ? 'bg-indigo-600 text-white' : 'bg-white text-gray-700 border border-gray-200 hover:bg-gray-50'}`}
          >Marketplace</button>
        </div>
      </div>

      {/* Health Summary Cards */}
      {healthSummary && (
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
          <SummaryCard label="Total" value={healthSummary.total_integrations} color="text-gray-900" />
          <SummaryCard label="Active" value={healthSummary.active_count} color="text-green-600" />
          <SummaryCard label="Healthy" value={healthSummary.healthy_count} color="text-green-600" />
          <SummaryCard label="Errors" value={healthSummary.error_count} color="text-red-600" />
        </div>
      )}

      {/* Active Integrations View */}
      {view === 'active' && (
        <div>
          {loading ? (
            <p className="text-gray-500">Loading integrations...</p>
          ) : integrations.length === 0 ? (
            <div className="card text-center py-12">
              <p className="text-gray-500 text-lg mb-2">No integrations configured yet</p>
              <p className="text-gray-400 text-sm mb-4">Browse the Marketplace to connect your first integration</p>
              <button onClick={() => setView('marketplace')} className="btn-primary">Browse Marketplace</button>
            </div>
          ) : (
            <div className="space-y-3">
              {integrations.map(intg => (
                <div key={intg.id} className="card flex items-center justify-between">
                  <div className="flex items-center gap-4 flex-1">
                    <div className="h-10 w-10 rounded-lg bg-indigo-100 flex items-center justify-center">
                      <span className="text-indigo-600 text-xs font-bold uppercase">{intg.integration_type.slice(0, 3)}</span>
                    </div>
                    <div className="flex-1">
                      <div className="flex items-center gap-2">
                        <h3 className="font-medium text-gray-900">{intg.name}</h3>
                        <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${statusColor[intg.status] || 'bg-gray-100 text-gray-600'}`}>
                          {intg.status.replace('_', ' ')}
                        </span>
                        <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${healthColor[intg.health_status] || 'bg-gray-100 text-gray-600'}`}>
                          {intg.health_status}
                        </span>
                      </div>
                      <div className="flex items-center gap-4 mt-1">
                        <span className="text-xs text-gray-500">Type: {intg.integration_type}</span>
                        {intg.last_sync_at && (
                          <span className="text-xs text-gray-400">
                            Last sync: {new Date(intg.last_sync_at).toLocaleString('en-GB', { dateStyle: 'short', timeStyle: 'short' })}
                          </span>
                        )}
                        {intg.error_count > 0 && (
                          <span className="text-xs text-red-500">{intg.error_count} error(s)</span>
                        )}
                      </div>
                      {intg.last_error_message && (
                        <p className="text-xs text-red-500 mt-1 truncate max-w-md">{intg.last_error_message}</p>
                      )}
                    </div>
                  </div>
                  <div className="flex items-center gap-2 ml-4">
                    <button
                      onClick={() => handleTest(intg.id)}
                      disabled={testingId === intg.id}
                      className="text-xs text-indigo-600 hover:underline disabled:opacity-50"
                    >{testingId === intg.id ? 'Testing...' : 'Test'}</button>
                    <button
                      onClick={() => handleSync(intg.id)}
                      disabled={syncingId === intg.id}
                      className="text-xs text-indigo-600 hover:underline disabled:opacity-50"
                    >{syncingId === intg.id ? 'Syncing...' : 'Sync'}</button>
                    <button
                      onClick={() => handleDelete(intg.id)}
                      className="text-xs text-red-600 hover:underline"
                    >Remove</button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* Marketplace View */}
      {view === 'marketplace' && (
        <div>
          {/* Category Tabs */}
          <div className="flex gap-1 border-b border-gray-200 mb-6 overflow-x-auto">
            {CATEGORIES.map(cat => (
              <button
                key={cat.key}
                onClick={() => setCategory(cat.key)}
                className={`px-4 py-2.5 text-sm font-medium border-b-2 transition-colors whitespace-nowrap ${
                  category === cat.key
                    ? 'border-indigo-600 text-indigo-600'
                    : 'border-transparent text-gray-500 hover:text-gray-700'
                }`}
              >{cat.label}</button>
            ))}
          </div>

          {/* Integration Cards Grid */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {catalogFiltered.map(entry => {
              const isConnected = connectedTypes.has(entry.type);
              return (
                <div key={entry.type} className="card hover:shadow-md transition-shadow">
                  <div className="flex items-start justify-between mb-3">
                    <div className="flex items-center gap-3">
                      <div className="h-10 w-10 rounded-lg bg-indigo-50 flex items-center justify-center">
                        <span className="text-indigo-600 text-xs font-bold uppercase">{entry.type.slice(0, 3)}</span>
                      </div>
                      <div>
                        <h3 className="font-medium text-gray-900">{entry.name}</h3>
                        <span className="text-xs text-gray-400 capitalize">{entry.category}</span>
                      </div>
                    </div>
                    {isConnected && (
                      <span className="text-xs bg-green-100 text-green-700 px-2 py-0.5 rounded-full font-medium">Connected</span>
                    )}
                  </div>
                  <p className="text-sm text-gray-600 mb-3">{entry.description}</p>
                  <div className="flex flex-wrap gap-1 mb-4">
                    {entry.capabilities.map(cap => (
                      <span key={cap} className="text-xs bg-gray-100 text-gray-600 px-2 py-0.5 rounded">{cap.replace('_', ' ')}</span>
                    ))}
                  </div>
                  <button
                    onClick={() => { setShowSetup(entry); setSetupName(entry.name); setConfigFields({}); }}
                    className={`w-full py-2 text-sm font-medium rounded-lg transition-colors ${
                      isConnected
                        ? 'bg-gray-100 text-gray-600 hover:bg-gray-200'
                        : 'bg-indigo-600 text-white hover:bg-indigo-700'
                    }`}
                  >{isConnected ? 'Add Another' : 'Connect'}</button>
                </div>
              );
            })}
          </div>
        </div>
      )}

      {/* Setup Modal */}
      {showSetup && (
        <div className="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4">
          <div className="bg-white rounded-xl shadow-2xl w-full max-w-lg max-h-[90vh] overflow-y-auto">
            <div className="p-6">
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-lg font-semibold text-gray-900">Connect {showSetup.name}</h2>
                <button onClick={() => setShowSetup(null)} className="text-gray-400 hover:text-gray-600 text-xl">&times;</button>
              </div>
              <p className="text-sm text-gray-500 mb-6">{showSetup.description}</p>

              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Display Name</label>
                  <input
                    type="text"
                    value={setupName}
                    onChange={e => setSetupName(e.target.value)}
                    className="input w-full"
                    placeholder={showSetup.name}
                  />
                </div>
                {showSetup.setup_fields.map(field => (
                  <div key={field}>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      {field.replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase())}
                    </label>
                    {field.includes('secret') || field.includes('password') || field.includes('token') || field.includes('key') && !field.includes('key_id') ? (
                      <input
                        type="password"
                        value={configFields[field] || ''}
                        onChange={e => setConfigFields(prev => ({ ...prev, [field]: e.target.value }))}
                        className="input w-full"
                        placeholder={`Enter ${field.replace(/_/g, ' ')}`}
                      />
                    ) : field === 'certificate' || field === 'service_account_key' ? (
                      <textarea
                        value={configFields[field] || ''}
                        onChange={e => setConfigFields(prev => ({ ...prev, [field]: e.target.value }))}
                        className="input w-full h-24 font-mono text-xs"
                        placeholder={`Paste ${field.replace(/_/g, ' ')} here`}
                      />
                    ) : (
                      <input
                        type="text"
                        value={configFields[field] || ''}
                        onChange={e => setConfigFields(prev => ({ ...prev, [field]: e.target.value }))}
                        className="input w-full"
                        placeholder={`Enter ${field.replace(/_/g, ' ')}`}
                      />
                    )}
                  </div>
                ))}
              </div>

              <div className="flex justify-end gap-3 mt-6 pt-4 border-t border-gray-100">
                <button
                  onClick={() => setShowSetup(null)}
                  className="btn-secondary"
                >Cancel</button>
                <button
                  onClick={handleCreate}
                  disabled={creating || !setupName}
                  className="btn-primary disabled:opacity-50"
                >{creating ? 'Connecting...' : 'Connect Integration'}</button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

// ============================================================
// SUB-COMPONENTS
// ============================================================

function SummaryCard({ label, value, color }: { label: string; value: number; color: string }) {
  return (
    <div className="card">
      <p className="text-sm text-gray-500">{label}</p>
      <p className={`text-2xl font-bold ${color}`}>{value}</p>
    </div>
  );
}
