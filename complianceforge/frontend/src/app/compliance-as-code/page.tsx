'use client';

import { useEffect, useState, useCallback } from 'react';
import api from '@/lib/api';

// ============================================================
// TYPES
// ============================================================

interface CaCRepository {
  id: string;
  name: string;
  description: string;
  provider: string;
  repo_url: string;
  branch: string;
  base_path: string;
  sync_direction: string;
  auto_sync: boolean;
  require_approval: boolean;
  last_sync_at: string | null;
  last_sync_status: string | null;
  resource_count: number;
  is_active: boolean;
  created_at: string;
}

interface CaCSyncRun {
  id: string;
  repository_id: string;
  repository_name: string;
  status: string;
  direction: string;
  trigger_type: string;
  commit_sha: string;
  branch: string;
  files_changed: number;
  resources_added: number;
  resources_updated: number;
  resources_deleted: number;
  resources_unchanged: number;
  errors: string[];
  started_at: string | null;
  completed_at: string | null;
  created_at: string;
}

interface CaCDriftEvent {
  id: string;
  repository_id: string;
  repository_name: string;
  drift_direction: string;
  status: string;
  kind: string;
  resource_name: string;
  resource_uid: string;
  field_path: string;
  repo_value: string;
  platform_value: string;
  description: string;
  detected_at: string;
  resolved_at: string | null;
  resolution_action: string;
}

interface CaCResourceMapping {
  id: string;
  repository_id: string;
  repository_name: string;
  file_path: string;
  api_version: string;
  kind: string;
  resource_name: string;
  resource_uid: string;
  platform_entity_type: string;
  status: string;
  content_hash: string;
  last_synced_at: string | null;
  created_at: string;
}

interface PaginationInfo {
  page: number;
  page_size: number;
  total_items: number;
  total_pages: number;
  has_next: boolean;
  has_prev: boolean;
}

// ============================================================
// CONSTANTS
// ============================================================

const SYNC_STATUS_COLORS: Record<string, string> = {
  pending: 'bg-gray-100 text-gray-600',
  running: 'bg-blue-100 text-blue-700',
  awaiting_approval: 'bg-amber-100 text-amber-700',
  approved: 'bg-emerald-100 text-emerald-700',
  rejected: 'bg-red-100 text-red-700',
  applying: 'bg-indigo-100 text-indigo-700',
  completed: 'bg-green-100 text-green-700',
  failed: 'bg-red-100 text-red-700',
};

const DRIFT_DIRECTION_COLORS: Record<string, string> = {
  repo_ahead: 'bg-blue-100 text-blue-700',
  platform_ahead: 'bg-amber-100 text-amber-700',
  conflict: 'bg-red-100 text-red-700',
};

const DRIFT_STATUS_COLORS: Record<string, string> = {
  open: 'bg-red-100 text-red-700',
  acknowledged: 'bg-amber-100 text-amber-700',
  resolved: 'bg-green-100 text-green-700',
  suppressed: 'bg-gray-100 text-gray-500',
};

const RESOURCE_STATUS_COLORS: Record<string, string> = {
  synced: 'bg-green-100 text-green-700',
  pending: 'bg-amber-100 text-amber-700',
  drift_detected: 'bg-red-100 text-red-700',
  conflict: 'bg-red-100 text-red-700',
  orphaned: 'bg-gray-100 text-gray-500',
  error: 'bg-red-100 text-red-700',
};

const PROVIDER_LABELS: Record<string, string> = {
  github: 'GitHub',
  gitlab: 'GitLab',
  bitbucket: 'Bitbucket',
  azure_devops: 'Azure DevOps',
};

const DIRECTION_LABELS: Record<string, string> = {
  repo_ahead: 'Repo Ahead',
  platform_ahead: 'Platform Ahead',
  conflict: 'Conflict',
};

const KIND_LABELS: Record<string, string> = {
  ControlImplementation: 'Control',
  Policy: 'Policy',
  RiskAcceptance: 'Risk Acceptance',
  EvidenceConfig: 'Evidence',
  Framework: 'Framework',
  RiskTreatment: 'Risk Treatment',
  AssetClassification: 'Asset Class.',
  AuditSchedule: 'Audit Schedule',
  IncidentPlaybook: 'Playbook',
  VendorAssessment: 'Vendor Assess.',
};

type Tab = 'repositories' | 'sync-history' | 'drift' | 'mappings';

// ============================================================
// COMPONENT
// ============================================================

export default function ComplianceAsCodePage() {
  const [tab, setTab] = useState<Tab>('repositories');
  const [loading, setLoading] = useState(true);

  // Repositories
  const [repos, setRepos] = useState<CaCRepository[]>([]);
  const [repoTotal, setRepoTotal] = useState(0);
  const [repoPage, setRepoPage] = useState(1);
  const [showConnectForm, setShowConnectForm] = useState(false);

  // Sync history
  const [syncRuns, setSyncRuns] = useState<CaCSyncRun[]>([]);
  const [syncTotal, setSyncTotal] = useState(0);
  const [syncPage, setSyncPage] = useState(1);

  // Drift
  const [driftEvents, setDriftEvents] = useState<CaCDriftEvent[]>([]);
  const [driftTotal, setDriftTotal] = useState(0);
  const [driftPage, setDriftPage] = useState(1);
  const [driftStatusFilter, setDriftStatusFilter] = useState('');
  const [driftDirectionFilter, setDriftDirectionFilter] = useState('');

  // Resource mappings
  const [mappings, setMappings] = useState<CaCResourceMapping[]>([]);
  const [mappingTotal, setMappingTotal] = useState(0);
  const [mappingPage, setMappingPage] = useState(1);
  const [mappingStatusFilter, setMappingStatusFilter] = useState('');
  const [mappingKindFilter, setMappingKindFilter] = useState('');

  // Connect form state
  const [formName, setFormName] = useState('');
  const [formRepoURL, setFormRepoURL] = useState('');
  const [formProvider, setFormProvider] = useState('github');
  const [formBranch, setFormBranch] = useState('main');
  const [formBasePath, setFormBasePath] = useState('/');
  const [formDirection, setFormDirection] = useState('pull');
  const [formAutoSync, setFormAutoSync] = useState(false);
  const [formRequireApproval, setFormRequireApproval] = useState(true);

  const loadData = useCallback(async () => {
    setLoading(true);
    try {
      if (tab === 'repositories') {
        const params = new URLSearchParams({
          page: String(repoPage),
          page_size: '20',
          is_active: 'true',
        });
        const res = await api.request<any>(`/cac/repositories?${params}`);
        setRepos(res.data || []);
        setRepoTotal(res.pagination?.total_items || 0);
      } else if (tab === 'sync-history') {
        const params = new URLSearchParams({
          page: String(syncPage),
          page_size: '20',
        });
        const res = await api.request<any>(`/cac/sync-runs?${params}`);
        setSyncRuns(res.data || []);
        setSyncTotal(res.pagination?.total_items || 0);
      } else if (tab === 'drift') {
        const params = new URLSearchParams({
          page: String(driftPage),
          page_size: '20',
        });
        if (driftStatusFilter) params.set('status', driftStatusFilter);
        if (driftDirectionFilter) params.set('direction', driftDirectionFilter);
        const res = await api.request<any>(`/cac/drift?${params}`);
        setDriftEvents(res.data || []);
        setDriftTotal(res.pagination?.total_items || 0);
      } else if (tab === 'mappings') {
        const params = new URLSearchParams({
          page: String(mappingPage),
          page_size: '20',
        });
        if (mappingStatusFilter) params.set('status', mappingStatusFilter);
        if (mappingKindFilter) params.set('kind', mappingKindFilter);
        const res = await api.request<any>(`/cac/resource-mappings?${params}`);
        setMappings(res.data || []);
        setMappingTotal(res.pagination?.total_items || 0);
      }
    } catch {
      /* swallow */
    }
    setLoading(false);
  }, [tab, repoPage, syncPage, driftPage, driftStatusFilter, driftDirectionFilter, mappingPage, mappingStatusFilter, mappingKindFilter]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  async function handleConnectRepo() {
    try {
      await api.request<any>('/cac/repositories', {
        method: 'POST',
        body: {
          name: formName,
          repo_url: formRepoURL,
          provider: formProvider,
          branch: formBranch,
          base_path: formBasePath,
          sync_direction: formDirection,
          auto_sync: formAutoSync,
          require_approval: formRequireApproval,
        },
      });
      setShowConnectForm(false);
      resetForm();
      loadData();
    } catch {
      /* swallow */
    }
  }

  async function handleTriggerSync(repoId: string) {
    try {
      await api.request<any>(`/cac/repositories/${repoId}/sync`, {
        method: 'POST',
        body: {},
      });
      loadData();
    } catch {
      /* swallow */
    }
  }

  async function handleDeleteRepo(repoId: string) {
    try {
      await api.request<any>(`/cac/repositories/${repoId}`, {
        method: 'DELETE',
      });
      loadData();
    } catch {
      /* swallow */
    }
  }

  async function handleApproveSyncRun(syncRunId: string) {
    try {
      await api.request<any>(`/cac/sync-runs/${syncRunId}/approve`, {
        method: 'POST',
        body: {},
      });
      loadData();
    } catch {
      /* swallow */
    }
  }

  async function handleRejectSyncRun(syncRunId: string) {
    try {
      await api.request<any>(`/cac/sync-runs/${syncRunId}/reject`, {
        method: 'POST',
        body: { reason: 'Rejected from UI' },
      });
      loadData();
    } catch {
      /* swallow */
    }
  }

  async function handleResolveDrift(driftId: string) {
    try {
      await api.request<any>(`/cac/drift/${driftId}/resolve`, {
        method: 'POST',
        body: {
          resolution_action: 'accept_platform',
          resolution_notes: 'Resolved from dashboard',
        },
      });
      loadData();
    } catch {
      /* swallow */
    }
  }

  function resetForm() {
    setFormName('');
    setFormRepoURL('');
    setFormProvider('github');
    setFormBranch('main');
    setFormBasePath('/');
    setFormDirection('pull');
    setFormAutoSync(false);
    setFormRequireApproval(true);
  }

  // Compute drift summary counts
  const driftByDirection = driftEvents.reduce(
    (acc, ev) => {
      if (ev.status === 'open' || ev.status === 'acknowledged') {
        acc[ev.drift_direction] = (acc[ev.drift_direction] || 0) + 1;
      }
      return acc;
    },
    {} as Record<string, number>
  );

  const tabs: { key: Tab; label: string }[] = [
    { key: 'repositories', label: 'Repositories' },
    { key: 'sync-history', label: 'Sync History' },
    { key: 'drift', label: 'Drift Detection' },
    { key: 'mappings', label: 'Resource Mappings' },
  ];

  return (
    <div>
      {/* Header */}
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">
            Compliance as Code
          </h1>
          <p className="text-gray-500 mt-1">
            Manage compliance policies, controls, and evidence as versioned code
            in Git repositories
          </p>
        </div>
        {tab === 'repositories' && (
          <button
            onClick={() => setShowConnectForm(true)}
            className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-700"
          >
            Connect Repository
          </button>
        )}
      </div>

      {/* Tab Bar */}
      <div className="flex gap-1 border-b border-gray-200 mb-6">
        {tabs.map((t) => (
          <button
            key={t.key}
            onClick={() => {
              setTab(t.key);
            }}
            className={`px-4 py-2.5 text-sm font-medium border-b-2 transition-colors ${
              tab === t.key
                ? 'border-indigo-600 text-indigo-600'
                : 'border-transparent text-gray-500 hover:text-gray-700'
            }`}
          >
            {t.label}
          </button>
        ))}
      </div>

      {/* Connect Repository Modal */}
      {showConnectForm && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
          <div className="bg-white rounded-xl shadow-xl w-full max-w-lg mx-4 p-6">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">
              Connect Repository
            </h2>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Name
                </label>
                <input
                  type="text"
                  value={formName}
                  onChange={(e) => setFormName(e.target.value)}
                  placeholder="e.g. compliance-policies"
                  className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Repository URL
                </label>
                <input
                  type="text"
                  value={formRepoURL}
                  onChange={(e) => setFormRepoURL(e.target.value)}
                  placeholder="https://github.com/org/repo"
                  className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                />
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Provider
                  </label>
                  <select
                    value={formProvider}
                    onChange={(e) => setFormProvider(e.target.value)}
                    className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm"
                  >
                    <option value="github">GitHub</option>
                    <option value="gitlab">GitLab</option>
                    <option value="bitbucket">Bitbucket</option>
                    <option value="azure_devops">Azure DevOps</option>
                  </select>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Branch
                  </label>
                  <input
                    type="text"
                    value={formBranch}
                    onChange={(e) => setFormBranch(e.target.value)}
                    className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  />
                </div>
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Base Path
                  </label>
                  <input
                    type="text"
                    value={formBasePath}
                    onChange={(e) => setFormBasePath(e.target.value)}
                    className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Sync Direction
                  </label>
                  <select
                    value={formDirection}
                    onChange={(e) => setFormDirection(e.target.value)}
                    className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm"
                  >
                    <option value="pull">Pull (Repo to Platform)</option>
                    <option value="push">Push (Platform to Repo)</option>
                    <option value="bidirectional">Bidirectional</option>
                  </select>
                </div>
              </div>
              <div className="flex gap-6">
                <label className="flex items-center gap-2 text-sm text-gray-700">
                  <input
                    type="checkbox"
                    checked={formAutoSync}
                    onChange={(e) => setFormAutoSync(e.target.checked)}
                    className="rounded border-gray-300"
                  />
                  Auto-sync on push
                </label>
                <label className="flex items-center gap-2 text-sm text-gray-700">
                  <input
                    type="checkbox"
                    checked={formRequireApproval}
                    onChange={(e) =>
                      setFormRequireApproval(e.target.checked)
                    }
                    className="rounded border-gray-300"
                  />
                  Require approval
                </label>
              </div>
            </div>
            <div className="flex justify-end gap-3 mt-6">
              <button
                onClick={() => {
                  setShowConnectForm(false);
                  resetForm();
                }}
                className="rounded-lg border border-gray-300 px-4 py-2 text-sm text-gray-700 hover:bg-gray-50"
              >
                Cancel
              </button>
              <button
                onClick={handleConnectRepo}
                disabled={!formName || !formRepoURL}
                className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-700 disabled:opacity-50"
              >
                Connect
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Content */}
      {loading ? (
        <div className="card animate-pulse h-96" />
      ) : tab === 'repositories' ? (
        /* ============================================================
           REPOSITORIES TAB
           ============================================================ */
        <div>
          {repos.length === 0 ? (
            <div className="card text-center py-16">
              <p className="text-gray-400 text-lg mb-2">
                No repositories connected
              </p>
              <p className="text-gray-400 text-sm mb-6">
                Connect a Git repository to start managing compliance as code
              </p>
              <button
                onClick={() => setShowConnectForm(true)}
                className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-700"
              >
                Connect Your First Repository
              </button>
            </div>
          ) : (
            <div className="grid gap-4">
              {repos.map((repo) => (
                <div key={repo.id} className="card">
                  <div className="flex items-start justify-between">
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-3">
                        <h3 className="font-semibold text-gray-900">
                          {repo.name}
                        </h3>
                        <span
                          className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${
                            repo.is_active
                              ? 'bg-green-100 text-green-700'
                              : 'bg-gray-100 text-gray-500'
                          }`}
                        >
                          {repo.is_active ? 'Active' : 'Inactive'}
                        </span>
                        {repo.last_sync_status && (
                          <span
                            className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${
                              SYNC_STATUS_COLORS[repo.last_sync_status] ||
                              'bg-gray-100 text-gray-600'
                            }`}
                          >
                            {repo.last_sync_status.replace(/_/g, ' ')}
                          </span>
                        )}
                      </div>
                      {repo.description && (
                        <p className="text-sm text-gray-500 mt-1">
                          {repo.description}
                        </p>
                      )}
                      <div className="flex items-center gap-4 mt-2 text-xs text-gray-400">
                        <span>
                          {PROVIDER_LABELS[repo.provider] || repo.provider}
                        </span>
                        <span className="font-mono">{repo.branch}</span>
                        <span>{repo.resource_count} resources</span>
                        <span>
                          {repo.sync_direction === 'pull'
                            ? 'Pull'
                            : repo.sync_direction === 'push'
                            ? 'Push'
                            : 'Bidirectional'}
                        </span>
                        {repo.auto_sync && (
                          <span className="text-indigo-500">Auto-sync</span>
                        )}
                        {repo.require_approval && (
                          <span className="text-amber-500">
                            Approval required
                          </span>
                        )}
                        {repo.last_sync_at && (
                          <span>
                            Last sync:{' '}
                            {new Date(repo.last_sync_at).toLocaleDateString(
                              'en-GB',
                              {
                                day: '2-digit',
                                month: 'short',
                                year: 'numeric',
                                hour: '2-digit',
                                minute: '2-digit',
                              }
                            )}
                          </span>
                        )}
                      </div>
                      <p className="text-xs text-gray-400 mt-1 font-mono truncate">
                        {repo.repo_url}
                      </p>
                    </div>
                    <div className="flex items-center gap-2 ml-4">
                      <button
                        onClick={() => handleTriggerSync(repo.id)}
                        className="rounded-lg bg-indigo-50 px-3 py-1.5 text-xs font-medium text-indigo-600 hover:bg-indigo-100"
                      >
                        Sync Now
                      </button>
                      <button
                        onClick={() => handleDeleteRepo(repo.id)}
                        className="rounded-lg bg-red-50 px-3 py-1.5 text-xs font-medium text-red-600 hover:bg-red-100"
                      >
                        Disconnect
                      </button>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}

          {/* Pagination */}
          {repoTotal > 20 && (
            <div className="flex items-center justify-between mt-4">
              <p className="text-sm text-gray-500">
                Showing {(repoPage - 1) * 20 + 1} to{' '}
                {Math.min(repoPage * 20, repoTotal)} of {repoTotal}
              </p>
              <div className="flex gap-2">
                <button
                  onClick={() => setRepoPage((p) => Math.max(1, p - 1))}
                  disabled={repoPage === 1}
                  className="rounded-lg border border-gray-300 px-3 py-1 text-sm disabled:opacity-50"
                >
                  Previous
                </button>
                <button
                  onClick={() => setRepoPage((p) => p + 1)}
                  disabled={repoPage * 20 >= repoTotal}
                  className="rounded-lg border border-gray-300 px-3 py-1 text-sm disabled:opacity-50"
                >
                  Next
                </button>
              </div>
            </div>
          )}
        </div>
      ) : tab === 'sync-history' ? (
        /* ============================================================
           SYNC HISTORY TAB
           ============================================================ */
        <div>
          <div className="card">
            <h3 className="font-semibold text-gray-900 mb-4">
              Sync Run Timeline
            </h3>
            {syncRuns.length === 0 ? (
              <p className="text-sm text-gray-400 py-8 text-center">
                No sync runs recorded yet
              </p>
            ) : (
              <div className="relative">
                <div className="absolute left-4 top-0 bottom-0 w-0.5 bg-gray-200" />
                <div className="space-y-4 pl-12">
                  {syncRuns.map((run) => (
                    <div key={run.id} className="relative">
                      <div
                        className={`absolute -left-8 top-1 h-3 w-3 rounded-full border-2 border-white ${
                          run.status === 'completed'
                            ? 'bg-green-500'
                            : run.status === 'failed'
                            ? 'bg-red-500'
                            : run.status === 'running' ||
                              run.status === 'applying'
                            ? 'bg-blue-500'
                            : run.status === 'awaiting_approval'
                            ? 'bg-amber-500'
                            : 'bg-gray-400'
                        }`}
                      />
                      <div className="flex items-start justify-between">
                        <div className="flex-1 min-w-0">
                          <div className="flex items-center gap-2">
                            <span className="text-xs font-medium text-gray-400">
                              {new Date(run.created_at).toLocaleDateString(
                                'en-GB',
                                {
                                  day: '2-digit',
                                  month: 'short',
                                  year: 'numeric',
                                  hour: '2-digit',
                                  minute: '2-digit',
                                }
                              )}
                            </span>
                            <span
                              className={`inline-flex items-center rounded-full px-2 py-0.5 text-[10px] font-medium ${
                                SYNC_STATUS_COLORS[run.status] ||
                                'bg-gray-100 text-gray-600'
                              }`}
                            >
                              {run.status.replace(/_/g, ' ')}
                            </span>
                            <span className="text-xs text-gray-400">
                              {run.trigger_type}
                            </span>
                          </div>
                          <p className="text-sm font-medium text-gray-900 mt-0.5">
                            {run.repository_name || 'Repository'}
                          </p>
                          <div className="flex items-center gap-3 mt-1 text-xs text-gray-400">
                            <span>{run.direction}</span>
                            {run.branch && (
                              <span className="font-mono">{run.branch}</span>
                            )}
                            {run.commit_sha && (
                              <span className="font-mono">
                                {run.commit_sha.substring(0, 8)}
                              </span>
                            )}
                            {run.status === 'completed' && (
                              <span>
                                +{run.resources_added} ~
                                {run.resources_updated} -{run.resources_deleted}{' '}
                                ={run.resources_unchanged}
                              </span>
                            )}
                          </div>
                        </div>
                        {run.status === 'awaiting_approval' && (
                          <div className="flex gap-2 ml-4">
                            <button
                              onClick={() => handleApproveSyncRun(run.id)}
                              className="rounded-lg bg-green-50 px-3 py-1 text-xs font-medium text-green-700 hover:bg-green-100"
                            >
                              Approve
                            </button>
                            <button
                              onClick={() => handleRejectSyncRun(run.id)}
                              className="rounded-lg bg-red-50 px-3 py-1 text-xs font-medium text-red-700 hover:bg-red-100"
                            >
                              Reject
                            </button>
                          </div>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>

          {syncTotal > 20 && (
            <div className="flex items-center justify-between mt-4">
              <p className="text-sm text-gray-500">
                Showing {(syncPage - 1) * 20 + 1} to{' '}
                {Math.min(syncPage * 20, syncTotal)} of {syncTotal}
              </p>
              <div className="flex gap-2">
                <button
                  onClick={() => setSyncPage((p) => Math.max(1, p - 1))}
                  disabled={syncPage === 1}
                  className="rounded-lg border border-gray-300 px-3 py-1 text-sm disabled:opacity-50"
                >
                  Previous
                </button>
                <button
                  onClick={() => setSyncPage((p) => p + 1)}
                  disabled={syncPage * 20 >= syncTotal}
                  className="rounded-lg border border-gray-300 px-3 py-1 text-sm disabled:opacity-50"
                >
                  Next
                </button>
              </div>
            </div>
          )}
        </div>
      ) : tab === 'drift' ? (
        /* ============================================================
           DRIFT TAB
           ============================================================ */
        <div>
          {/* Drift Dashboard Summary */}
          <div className="grid grid-cols-1 gap-4 md:grid-cols-3 mb-6">
            <div className="card">
              <p className="text-sm text-gray-500">Repo Ahead</p>
              <p className="text-3xl font-bold text-blue-600 mt-1">
                {driftByDirection['repo_ahead'] || 0}
              </p>
              <p className="text-xs text-gray-400 mt-1">
                Changes in repo not yet in platform
              </p>
            </div>
            <div className="card">
              <p className="text-sm text-gray-500">Platform Ahead</p>
              <p className="text-3xl font-bold text-amber-600 mt-1">
                {driftByDirection['platform_ahead'] || 0}
              </p>
              <p className="text-xs text-gray-400 mt-1">
                Platform changes not in repo
              </p>
            </div>
            <div className="card">
              <p className="text-sm text-gray-500">Conflicts</p>
              <p className="text-3xl font-bold text-red-600 mt-1">
                {driftByDirection['conflict'] || 0}
              </p>
              <p className="text-xs text-gray-400 mt-1">
                Both sides changed independently
              </p>
            </div>
          </div>

          {/* Filters */}
          <div className="flex flex-wrap gap-3 mb-4">
            <select
              value={driftStatusFilter}
              onChange={(e) => {
                setDriftStatusFilter(e.target.value);
                setDriftPage(1);
              }}
              className="rounded-lg border border-gray-300 px-3 py-1.5 text-sm"
            >
              <option value="">All Statuses</option>
              <option value="open">Open</option>
              <option value="acknowledged">Acknowledged</option>
              <option value="resolved">Resolved</option>
              <option value="suppressed">Suppressed</option>
            </select>
            <select
              value={driftDirectionFilter}
              onChange={(e) => {
                setDriftDirectionFilter(e.target.value);
                setDriftPage(1);
              }}
              className="rounded-lg border border-gray-300 px-3 py-1.5 text-sm"
            >
              <option value="">All Directions</option>
              <option value="repo_ahead">Repo Ahead</option>
              <option value="platform_ahead">Platform Ahead</option>
              <option value="conflict">Conflict</option>
            </select>
          </div>

          {/* Drift Events Table */}
          <div className="card overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="table-header">
                  <th className="px-4 py-3 text-left">Resource</th>
                  <th className="px-4 py-3 text-left">Kind</th>
                  <th className="px-4 py-3 text-left">Repository</th>
                  <th className="px-4 py-3 text-left">Direction</th>
                  <th className="px-4 py-3 text-left">Status</th>
                  <th className="px-4 py-3 text-left">Detected</th>
                  <th className="px-4 py-3 text-left">Actions</th>
                </tr>
              </thead>
              <tbody>
                {driftEvents.map((ev) => (
                  <tr
                    key={ev.id}
                    className="border-t border-gray-100 hover:bg-gray-50"
                  >
                    <td className="px-4 py-3">
                      <div>
                        <p className="font-medium text-gray-900">
                          {ev.resource_name}
                        </p>
                        <p className="text-xs text-gray-400 font-mono">
                          {ev.resource_uid}
                        </p>
                      </div>
                    </td>
                    <td className="px-4 py-3">
                      <span className="inline-flex items-center rounded bg-gray-100 px-2 py-0.5 text-xs font-medium text-gray-600">
                        {KIND_LABELS[ev.kind] || ev.kind}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-gray-600 text-xs">
                      {ev.repository_name || '-'}
                    </td>
                    <td className="px-4 py-3">
                      <span
                        className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${
                          DRIFT_DIRECTION_COLORS[ev.drift_direction] ||
                          'bg-gray-100 text-gray-600'
                        }`}
                      >
                        {DIRECTION_LABELS[ev.drift_direction] ||
                          ev.drift_direction}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <span
                        className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${
                          DRIFT_STATUS_COLORS[ev.status] ||
                          'bg-gray-100 text-gray-600'
                        }`}
                      >
                        {ev.status}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-xs text-gray-500 whitespace-nowrap">
                      {new Date(ev.detected_at).toLocaleDateString('en-GB', {
                        day: '2-digit',
                        month: 'short',
                        year: 'numeric',
                      })}
                    </td>
                    <td className="px-4 py-3">
                      {(ev.status === 'open' ||
                        ev.status === 'acknowledged') && (
                        <button
                          onClick={() => handleResolveDrift(ev.id)}
                          className="text-indigo-600 hover:text-indigo-700 text-xs font-medium"
                        >
                          Resolve
                        </button>
                      )}
                    </td>
                  </tr>
                ))}
                {driftEvents.length === 0 && (
                  <tr>
                    <td
                      colSpan={7}
                      className="px-4 py-12 text-center text-gray-500"
                    >
                      No drift events detected
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>

          {driftTotal > 20 && (
            <div className="flex items-center justify-between mt-4">
              <p className="text-sm text-gray-500">
                Showing {(driftPage - 1) * 20 + 1} to{' '}
                {Math.min(driftPage * 20, driftTotal)} of {driftTotal}
              </p>
              <div className="flex gap-2">
                <button
                  onClick={() => setDriftPage((p) => Math.max(1, p - 1))}
                  disabled={driftPage === 1}
                  className="rounded-lg border border-gray-300 px-3 py-1 text-sm disabled:opacity-50"
                >
                  Previous
                </button>
                <button
                  onClick={() => setDriftPage((p) => p + 1)}
                  disabled={driftPage * 20 >= driftTotal}
                  className="rounded-lg border border-gray-300 px-3 py-1 text-sm disabled:opacity-50"
                >
                  Next
                </button>
              </div>
            </div>
          )}
        </div>
      ) : tab === 'mappings' ? (
        /* ============================================================
           RESOURCE MAPPINGS TAB
           ============================================================ */
        <div>
          {/* Filters */}
          <div className="flex flex-wrap gap-3 mb-4">
            <select
              value={mappingStatusFilter}
              onChange={(e) => {
                setMappingStatusFilter(e.target.value);
                setMappingPage(1);
              }}
              className="rounded-lg border border-gray-300 px-3 py-1.5 text-sm"
            >
              <option value="">All Statuses</option>
              <option value="synced">Synced</option>
              <option value="pending">Pending</option>
              <option value="drift_detected">Drift Detected</option>
              <option value="conflict">Conflict</option>
              <option value="orphaned">Orphaned</option>
              <option value="error">Error</option>
            </select>
            <select
              value={mappingKindFilter}
              onChange={(e) => {
                setMappingKindFilter(e.target.value);
                setMappingPage(1);
              }}
              className="rounded-lg border border-gray-300 px-3 py-1.5 text-sm"
            >
              <option value="">All Kinds</option>
              <option value="ControlImplementation">Control Implementation</option>
              <option value="Policy">Policy</option>
              <option value="RiskAcceptance">Risk Acceptance</option>
              <option value="EvidenceConfig">Evidence Config</option>
              <option value="Framework">Framework</option>
              <option value="RiskTreatment">Risk Treatment</option>
              <option value="AssetClassification">Asset Classification</option>
              <option value="AuditSchedule">Audit Schedule</option>
              <option value="IncidentPlaybook">Incident Playbook</option>
              <option value="VendorAssessment">Vendor Assessment</option>
            </select>
          </div>

          {/* Mappings Table */}
          <div className="card overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="table-header">
                  <th className="px-4 py-3 text-left">Resource</th>
                  <th className="px-4 py-3 text-left">Kind</th>
                  <th className="px-4 py-3 text-left">File Path</th>
                  <th className="px-4 py-3 text-left">Repository</th>
                  <th className="px-4 py-3 text-left">Platform Entity</th>
                  <th className="px-4 py-3 text-left">Status</th>
                  <th className="px-4 py-3 text-left">Last Synced</th>
                </tr>
              </thead>
              <tbody>
                {mappings.map((m) => (
                  <tr
                    key={m.id}
                    className="border-t border-gray-100 hover:bg-gray-50"
                  >
                    <td className="px-4 py-3">
                      <div>
                        <p className="font-medium text-gray-900">
                          {m.resource_name}
                        </p>
                        <p className="text-xs text-gray-400 font-mono">
                          {m.resource_uid}
                        </p>
                      </div>
                    </td>
                    <td className="px-4 py-3">
                      <span className="inline-flex items-center rounded bg-indigo-50 px-2 py-0.5 text-xs font-medium text-indigo-600">
                        {KIND_LABELS[m.kind] || m.kind}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-xs text-gray-500 font-mono max-w-xs truncate">
                      {m.file_path}
                    </td>
                    <td className="px-4 py-3 text-gray-600 text-xs">
                      {m.repository_name || '-'}
                    </td>
                    <td className="px-4 py-3 text-xs text-gray-500">
                      {m.platform_entity_type
                        ? m.platform_entity_type.replace(/_/g, ' ')
                        : '-'}
                    </td>
                    <td className="px-4 py-3">
                      <span
                        className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${
                          RESOURCE_STATUS_COLORS[m.status] ||
                          'bg-gray-100 text-gray-600'
                        }`}
                      >
                        {m.status.replace(/_/g, ' ')}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-xs text-gray-500 whitespace-nowrap">
                      {m.last_synced_at
                        ? new Date(m.last_synced_at).toLocaleDateString(
                            'en-GB',
                            {
                              day: '2-digit',
                              month: 'short',
                              year: 'numeric',
                            }
                          )
                        : 'Never'}
                    </td>
                  </tr>
                ))}
                {mappings.length === 0 && (
                  <tr>
                    <td
                      colSpan={7}
                      className="px-4 py-12 text-center text-gray-500"
                    >
                      No resource mappings found
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>

          {mappingTotal > 20 && (
            <div className="flex items-center justify-between mt-4">
              <p className="text-sm text-gray-500">
                Showing {(mappingPage - 1) * 20 + 1} to{' '}
                {Math.min(mappingPage * 20, mappingTotal)} of {mappingTotal}
              </p>
              <div className="flex gap-2">
                <button
                  onClick={() => setMappingPage((p) => Math.max(1, p - 1))}
                  disabled={mappingPage === 1}
                  className="rounded-lg border border-gray-300 px-3 py-1 text-sm disabled:opacity-50"
                >
                  Previous
                </button>
                <button
                  onClick={() => setMappingPage((p) => p + 1)}
                  disabled={mappingPage * 20 >= mappingTotal}
                  className="rounded-lg border border-gray-300 px-3 py-1 text-sm disabled:opacity-50"
                >
                  Next
                </button>
              </div>
            </div>
          )}
        </div>
      ) : null}
    </div>
  );
}
