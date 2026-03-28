'use client';

import { useEffect, useState, useCallback } from 'react';
import api from '@/lib/api';

// ── Types ────────────────────────────────────────────────────

interface Condition {
  field: string;
  operator: string;
  value: unknown;
}

interface AccessPolicy {
  id: string;
  org_id: string;
  name: string;
  description: string;
  priority: number;
  effect: 'allow' | 'deny';
  is_active: boolean;
  subject_conditions: Condition[];
  resource_type: string;
  resource_conditions: Condition[];
  actions: string[];
  environment_conditions: Condition[];
  valid_from?: string;
  valid_until?: string;
  created_at: string;
  updated_at: string;
}

interface PolicyAssignment {
  id: string;
  org_id: string;
  access_policy_id: string;
  assignee_type: string;
  assignee_id?: string;
  valid_from?: string;
  valid_until?: string;
  created_at: string;
}

interface AuditLogEntry {
  id: string;
  user_id: string;
  action: string;
  resource_type: string;
  resource_id?: string;
  decision: 'allow' | 'deny';
  matched_policy_id?: string;
  evaluation_time_us: number;
  subject_attributes: Record<string, unknown>;
  resource_attributes: Record<string, unknown>;
  environment_attributes: Record<string, unknown>;
  created_at: string;
}

interface Permission {
  policy_id: string;
  policy_name: string;
  resource_type: string;
  actions: string[];
  priority: number;
  conditions: boolean;
}

type Tab = 'policies' | 'audit' | 'permissions' | 'evaluate';

const EFFECT_COLORS: Record<string, string> = {
  allow: 'badge-low',
  deny: 'badge-critical',
};

const OPERATORS = [
  'equals', 'not_equals', 'in', 'not_in', 'contains_any',
  'greater_than', 'less_than', 'in_cidr', 'between', 'equals_subject',
];

const RESOURCE_TYPES = [
  '*', 'control', 'risk', 'incident', 'policy', 'dsr_request',
  'vendor', 'audit', 'asset', 'report',
];

const ACTIONS = [
  'read', 'create', 'update', 'delete', 'export', 'approve', 'assign',
];

// ── Main Page ────────────────────────────────────────────────

export default function AccessPoliciesPage() {
  const [tab, setTab] = useState<Tab>('policies');

  const tabs: { key: Tab; label: string }[] = [
    { key: 'policies', label: 'Access Policies' },
    { key: 'audit', label: 'Decision Log' },
    { key: 'permissions', label: 'My Permissions' },
    { key: 'evaluate', label: 'Policy Tester' },
  ];

  return (
    <div>
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Access Control (ABAC)</h1>
          <p className="text-gray-500 mt-1">
            Attribute-based access policies for fine-grained authorization
          </p>
        </div>
      </div>

      <div className="flex gap-1 border-b border-gray-200 mb-6">
        {tabs.map(t => (
          <button
            key={t.key}
            onClick={() => setTab(t.key)}
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

      {tab === 'policies' && <PoliciesTab />}
      {tab === 'audit' && <AuditLogTab />}
      {tab === 'permissions' && <PermissionsTab />}
      {tab === 'evaluate' && <EvaluateTab />}
    </div>
  );
}

// ── Policies Tab ─────────────────────────────────────────────

function PoliciesTab() {
  const [policies, setPolicies] = useState<AccessPolicy[]>([]);
  const [loading, setLoading] = useState(true);
  const [showCreate, setShowCreate] = useState(false);
  const [expandedId, setExpandedId] = useState<string | null>(null);
  const [assignments, setAssignments] = useState<Record<string, PolicyAssignment[]>>({});

  const loadPolicies = useCallback(async () => {
    setLoading(true);
    try {
      const res = await api.getAccessPolicies();
      setPolicies(res.data?.data || res.data || []);
    } catch { /* ignore */ }
    setLoading(false);
  }, []);

  useEffect(() => { loadPolicies(); }, [loadPolicies]);

  async function loadAssignments(policyId: string) {
    try {
      const res = await api.getAccessPolicyAssignments(policyId);
      setAssignments(prev => ({ ...prev, [policyId]: res.data?.data || res.data || [] }));
    } catch { /* ignore */ }
  }

  function toggleExpand(id: string) {
    if (expandedId === id) {
      setExpandedId(null);
    } else {
      setExpandedId(id);
      if (!assignments[id]) loadAssignments(id);
    }
  }

  async function toggleActive(policy: AccessPolicy) {
    try {
      await api.updateAccessPolicy(policy.id, { is_active: !policy.is_active });
      loadPolicies();
    } catch { /* ignore */ }
  }

  async function deletePolicy(id: string) {
    if (!confirm('Delete this access policy? This cannot be undone.')) return;
    try {
      await api.deleteAccessPolicy(id);
      loadPolicies();
    } catch { /* ignore */ }
  }

  if (loading) return <div className="card animate-pulse h-64" />;

  return (
    <div>
      <div className="flex justify-between items-center mb-4">
        <p className="text-sm text-gray-500">
          {policies.length} polic{policies.length === 1 ? 'y' : 'ies'} configured
        </p>
        <button className="btn-primary" onClick={() => setShowCreate(true)}>
          + Create Policy
        </button>
      </div>

      {showCreate && (
        <CreatePolicyForm
          onClose={() => setShowCreate(false)}
          onCreated={() => { setShowCreate(false); loadPolicies(); }}
        />
      )}

      <div className="space-y-3">
        {policies.map(policy => (
          <div key={policy.id} className="card">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className={`h-3 w-3 rounded-full ${policy.is_active ? 'bg-green-400' : 'bg-gray-300'}`} />
                <div>
                  <div className="flex items-center gap-2">
                    <h3 className="font-medium text-gray-900">{policy.name}</h3>
                    <span className={`badge ${EFFECT_COLORS[policy.effect] || 'badge-info'}`}>
                      {policy.effect.toUpperCase()}
                    </span>
                    <span className="text-xs text-gray-400">Priority: {policy.priority}</span>
                  </div>
                  <p className="text-sm text-gray-500 mt-0.5">{policy.description}</p>
                </div>
              </div>
              <div className="flex items-center gap-2">
                <span className="badge badge-info text-xs">{policy.resource_type}</span>
                <span className="text-xs text-gray-400">
                  {policy.actions.join(', ')}
                </span>
                <button
                  onClick={() => toggleActive(policy)}
                  className="text-xs text-indigo-600 hover:underline"
                >
                  {policy.is_active ? 'Disable' : 'Enable'}
                </button>
                <button
                  onClick={() => toggleExpand(policy.id)}
                  className="text-xs text-indigo-600 hover:underline"
                >
                  {expandedId === policy.id ? 'Collapse' : 'Details'}
                </button>
                <button
                  onClick={() => deletePolicy(policy.id)}
                  className="text-xs text-red-600 hover:underline"
                >
                  Delete
                </button>
              </div>
            </div>

            {expandedId === policy.id && (
              <PolicyDetails
                policy={policy}
                assignments={assignments[policy.id] || []}
                onAssignmentCreated={() => loadAssignments(policy.id)}
              />
            )}
          </div>
        ))}

        {policies.length === 0 && (
          <div className="text-center py-12">
            <p className="text-gray-500 mb-4">No access policies configured yet.</p>
            <button className="btn-primary" onClick={() => setShowCreate(true)}>
              Create Your First Policy
            </button>
          </div>
        )}
      </div>
    </div>
  );
}

// ── Policy Details (expanded) ────────────────────────────────

function PolicyDetails({
  policy,
  assignments,
  onAssignmentCreated,
}: {
  policy: AccessPolicy;
  assignments: PolicyAssignment[];
  onAssignmentCreated: () => void;
}) {
  const [showAssign, setShowAssign] = useState(false);

  async function removeAssignment(assignmentId: string) {
    try {
      await api.deleteAccessPolicyAssignment(policy.id, assignmentId);
      onAssignmentCreated();
    } catch { /* ignore */ }
  }

  return (
    <div className="mt-4 pt-4 border-t border-gray-100 space-y-4">
      {/* Conditions */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <ConditionBlock title="Subject Conditions" conditions={policy.subject_conditions} />
        <ConditionBlock title="Resource Conditions" conditions={policy.resource_conditions} />
        <ConditionBlock title="Environment Conditions" conditions={policy.environment_conditions} />
      </div>

      {/* Temporal */}
      {(policy.valid_from || policy.valid_until) && (
        <div className="text-sm text-gray-500">
          <span className="font-medium">Temporal: </span>
          {policy.valid_from && <span>From {new Date(policy.valid_from).toLocaleDateString('en-GB')}</span>}
          {policy.valid_from && policy.valid_until && <span> to </span>}
          {policy.valid_until && <span>{new Date(policy.valid_until).toLocaleDateString('en-GB')}</span>}
        </div>
      )}

      {/* Assignments */}
      <div>
        <div className="flex items-center justify-between mb-2">
          <h4 className="text-sm font-semibold text-gray-700">Assignments</h4>
          <button className="text-xs text-indigo-600 hover:underline" onClick={() => setShowAssign(!showAssign)}>
            + Add Assignment
          </button>
        </div>

        {showAssign && (
          <AssignmentForm
            policyId={policy.id}
            onClose={() => setShowAssign(false)}
            onCreated={() => { setShowAssign(false); onAssignmentCreated(); }}
          />
        )}

        {assignments.length > 0 ? (
          <div className="space-y-1">
            {assignments.map(a => (
              <div key={a.id} className="flex items-center justify-between text-sm bg-gray-50 rounded px-3 py-2">
                <div>
                  <span className="badge badge-info text-xs mr-2">{a.assignee_type}</span>
                  {a.assignee_id && <span className="text-gray-500 font-mono text-xs">{a.assignee_id}</span>}
                </div>
                <button
                  onClick={() => removeAssignment(a.id)}
                  className="text-xs text-red-600 hover:underline"
                >
                  Remove
                </button>
              </div>
            ))}
          </div>
        ) : (
          <p className="text-xs text-gray-400">No assignments yet</p>
        )}
      </div>
    </div>
  );
}

// ── Condition Display Block ──────────────────────────────────

function ConditionBlock({ title, conditions }: { title: string; conditions: Condition[] }) {
  return (
    <div>
      <h5 className="text-xs font-semibold text-gray-500 uppercase mb-1">{title}</h5>
      {conditions && conditions.length > 0 ? (
        <div className="space-y-1">
          {conditions.map((c, i) => (
            <div key={i} className="text-xs bg-gray-50 rounded px-2 py-1 font-mono">
              <span className="text-indigo-600">{c.field}</span>
              {' '}
              <span className="text-orange-600">{c.operator}</span>
              {' '}
              <span className="text-gray-700">{JSON.stringify(c.value)}</span>
            </div>
          ))}
        </div>
      ) : (
        <p className="text-xs text-gray-400">None (no restrictions)</p>
      )}
    </div>
  );
}

// ── Create Policy Form ───────────────────────────────────────

function CreatePolicyForm({
  onClose,
  onCreated,
}: {
  onClose: () => void;
  onCreated: () => void;
}) {
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [effect, setEffect] = useState<'allow' | 'deny'>('allow');
  const [priority, setPriority] = useState(100);
  const [resourceType, setResourceType] = useState('*');
  const [actions, setActions] = useState<string[]>(['read']);
  const [isActive, setIsActive] = useState(true);
  const [saving, setSaving] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setSaving(true);
    try {
      await api.createAccessPolicy({
        name,
        description,
        effect,
        priority,
        resource_type: resourceType,
        actions,
        is_active: isActive,
        subject_conditions: [],
        resource_conditions: [],
        environment_conditions: [],
      });
      onCreated();
    } catch { /* ignore */ }
    setSaving(false);
  }

  return (
    <div className="card mb-4 border-2 border-indigo-100">
      <div className="flex items-center justify-between mb-4">
        <h3 className="font-semibold text-gray-900">Create Access Policy</h3>
        <button onClick={onClose} className="text-gray-400 hover:text-gray-600 text-lg">&times;</button>
      </div>
      <form onSubmit={handleSubmit} className="space-y-4">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Name</label>
            <input
              className="input w-full"
              value={name}
              onChange={e => setName(e.target.value)}
              placeholder="Policy name"
              required
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Effect</label>
            <select className="input w-full" value={effect} onChange={e => setEffect(e.target.value as 'allow' | 'deny')}>
              <option value="allow">Allow</option>
              <option value="deny">Deny</option>
            </select>
          </div>
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
          <textarea
            className="input w-full"
            value={description}
            onChange={e => setDescription(e.target.value)}
            rows={2}
            placeholder="Describe when this policy applies"
          />
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Priority</label>
            <input
              type="number"
              className="input w-full"
              value={priority}
              onChange={e => setPriority(Number(e.target.value))}
              min={1}
              max={999}
            />
            <p className="text-xs text-gray-400 mt-1">Lower = higher priority</p>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Resource Type</label>
            <select className="input w-full" value={resourceType} onChange={e => setResourceType(e.target.value)}>
              {RESOURCE_TYPES.map(rt => (
                <option key={rt} value={rt}>{rt === '*' ? 'All Resources' : rt}</option>
              ))}
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Status</label>
            <label className="flex items-center gap-2 mt-2">
              <input
                type="checkbox"
                checked={isActive}
                onChange={e => setIsActive(e.target.checked)}
                className="rounded border-gray-300 text-indigo-600"
              />
              <span className="text-sm text-gray-600">Active</span>
            </label>
          </div>
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Actions</label>
          <div className="flex flex-wrap gap-2">
            {ACTIONS.map(a => (
              <label key={a} className="flex items-center gap-1">
                <input
                  type="checkbox"
                  checked={actions.includes(a)}
                  onChange={e => {
                    if (e.target.checked) setActions(prev => [...prev, a]);
                    else setActions(prev => prev.filter(x => x !== a));
                  }}
                  className="rounded border-gray-300 text-indigo-600"
                />
                <span className="text-sm text-gray-600">{a}</span>
              </label>
            ))}
          </div>
        </div>

        <div className="flex justify-end gap-2">
          <button type="button" onClick={onClose} className="btn-secondary">Cancel</button>
          <button type="submit" disabled={saving || !name} className="btn-primary">
            {saving ? 'Creating...' : 'Create Policy'}
          </button>
        </div>
      </form>
    </div>
  );
}

// ── Assignment Form ──────────────────────────────────────────

function AssignmentForm({
  policyId,
  onClose,
  onCreated,
}: {
  policyId: string;
  onClose: () => void;
  onCreated: () => void;
}) {
  const [assigneeType, setAssigneeType] = useState('role');
  const [assigneeId, setAssigneeId] = useState('');
  const [saving, setSaving] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setSaving(true);
    try {
      await api.createAccessPolicyAssignment(policyId, {
        assignee_type: assigneeType,
        assignee_id: assigneeId || undefined,
      });
      onCreated();
    } catch { /* ignore */ }
    setSaving(false);
  }

  return (
    <form onSubmit={handleSubmit} className="flex items-end gap-2 mb-3 bg-gray-50 rounded p-3">
      <div>
        <label className="text-xs font-medium text-gray-600">Type</label>
        <select className="input text-sm" value={assigneeType} onChange={e => setAssigneeType(e.target.value)}>
          <option value="user">User</option>
          <option value="role">Role</option>
          <option value="group">Group</option>
          <option value="all_users">All Users</option>
        </select>
      </div>
      {assigneeType !== 'all_users' && (
        <div className="flex-1">
          <label className="text-xs font-medium text-gray-600">Assignee ID</label>
          <input
            className="input text-sm w-full"
            value={assigneeId}
            onChange={e => setAssigneeId(e.target.value)}
            placeholder="UUID of user, role, or group"
          />
        </div>
      )}
      <button type="submit" disabled={saving} className="btn-primary text-sm py-1.5">
        {saving ? '...' : 'Assign'}
      </button>
      <button type="button" onClick={onClose} className="btn-secondary text-sm py-1.5">Cancel</button>
    </form>
  );
}

// ── Audit Log Tab ────────────────────────────────────────────

function AuditLogTab() {
  const [entries, setEntries] = useState<AuditLogEntry[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api.getAccessAuditLog()
      .then(res => setEntries(res.data?.data || res.data || []))
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <div className="card animate-pulse h-64" />;

  return (
    <div>
      <p className="text-sm text-gray-500 mb-4">
        Immutable log of all ABAC policy evaluation decisions
      </p>
      <div className="card overflow-hidden p-0">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-gray-100">
              <th className="table-header px-4 py-3">Time</th>
              <th className="table-header px-4 py-3">User</th>
              <th className="table-header px-4 py-3">Action</th>
              <th className="table-header px-4 py-3">Resource</th>
              <th className="table-header px-4 py-3">Decision</th>
              <th className="table-header px-4 py-3">Time (us)</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-50">
            {entries.map(entry => (
              <tr key={entry.id} className="hover:bg-gray-50">
                <td className="px-4 py-3 text-xs font-mono text-gray-500 whitespace-nowrap">
                  {new Date(entry.created_at).toLocaleString('en-GB', {
                    dateStyle: 'short',
                    timeStyle: 'medium',
                  })}
                </td>
                <td className="px-4 py-3 text-xs font-mono text-gray-600">
                  {entry.user_id.slice(0, 8)}...
                </td>
                <td className="px-4 py-3">
                  <span className="badge badge-info text-xs">{entry.action}</span>
                </td>
                <td className="px-4 py-3 text-sm text-gray-700">
                  {entry.resource_type}
                  {entry.resource_id && (
                    <span className="text-gray-400 ml-1 text-xs">({entry.resource_id.slice(0, 8)})</span>
                  )}
                </td>
                <td className="px-4 py-3">
                  <span className={`badge ${entry.decision === 'allow' ? 'badge-low' : 'badge-critical'}`}>
                    {entry.decision}
                  </span>
                </td>
                <td className="px-4 py-3 text-xs text-gray-500 font-mono">
                  {entry.evaluation_time_us}
                </td>
              </tr>
            ))}
            {entries.length === 0 && (
              <tr>
                <td colSpan={6} className="px-4 py-8 text-center text-gray-500">
                  No access decisions logged yet
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}

// ── Permissions Tab ──────────────────────────────────────────

function PermissionsTab() {
  const [permissions, setPermissions] = useState<Permission[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api.getMyAccessPermissions()
      .then(res => setPermissions(res.data?.permissions || []))
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <div className="card animate-pulse h-64" />;

  return (
    <div>
      <p className="text-sm text-gray-500 mb-4">
        These are the access permissions granted to your account based on active ABAC policies.
      </p>

      {permissions.length > 0 ? (
        <div className="space-y-3">
          {permissions.map(perm => (
            <div key={perm.policy_id} className="card flex items-center justify-between">
              <div>
                <div className="flex items-center gap-2">
                  <h3 className="font-medium text-gray-900">{perm.policy_name}</h3>
                  <span className="badge badge-info text-xs">{perm.resource_type}</span>
                  {perm.conditions && (
                    <span className="text-xs text-orange-500">Conditional</span>
                  )}
                </div>
                <div className="flex gap-1 mt-1">
                  {perm.actions.map(a => (
                    <span key={a} className="text-xs bg-gray-100 text-gray-600 rounded px-1.5 py-0.5">
                      {a}
                    </span>
                  ))}
                </div>
              </div>
              <span className="text-xs text-gray-400">Priority: {perm.priority}</span>
            </div>
          ))}
        </div>
      ) : (
        <div className="card text-center py-12">
          <p className="text-gray-500">No permissions found for your account.</p>
          <p className="text-sm text-gray-400 mt-1">Contact your administrator to assign access policies.</p>
        </div>
      )}
    </div>
  );
}

// ── Evaluate Tab (Policy Tester) ─────────────────────────────

function EvaluateTab() {
  const [subjectId, setSubjectId] = useState('');
  const [action, setAction] = useState('read');
  const [resourceType, setResourceType] = useState('control');
  const [resourceId, setResourceId] = useState('');
  const [roles, setRoles] = useState('');
  const [mfaVerified, setMfaVerified] = useState(false);
  const [ip, setIp] = useState('');
  const [result, setResult] = useState<Record<string, unknown> | null>(null);
  const [testing, setTesting] = useState(false);

  async function handleEvaluate(e: React.FormEvent) {
    e.preventDefault();
    setTesting(true);
    setResult(null);
    try {
      const res = await api.evaluateAccessPolicy({
        subject_id: subjectId,
        action,
        resource_type: resourceType,
        resource_id: resourceId || undefined,
        roles: roles.split(',').map(r => r.trim()).filter(Boolean),
        mfa_verified: mfaVerified,
        ip: ip || undefined,
      });
      setResult(res.data);
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Evaluation failed';
      setResult({ error: message });
    }
    setTesting(false);
  }

  return (
    <div>
      <p className="text-sm text-gray-500 mb-4">
        Test ABAC policy evaluation without making real access requests. This is an admin diagnostic tool.
      </p>

      <div className="card">
        <h3 className="font-semibold text-gray-900 mb-4">Policy Evaluation Tester</h3>
        <form onSubmit={handleEvaluate} className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Subject (User ID)</label>
              <input
                className="input w-full"
                value={subjectId}
                onChange={e => setSubjectId(e.target.value)}
                placeholder="UUID of the user to test"
                required
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Roles (comma separated)</label>
              <input
                className="input w-full"
                value={roles}
                onChange={e => setRoles(e.target.value)}
                placeholder="e.g. org_admin, viewer"
              />
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Action</label>
              <select className="input w-full" value={action} onChange={e => setAction(e.target.value)}>
                {ACTIONS.map(a => <option key={a} value={a}>{a}</option>)}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Resource Type</label>
              <select className="input w-full" value={resourceType} onChange={e => setResourceType(e.target.value)}>
                {RESOURCE_TYPES.filter(r => r !== '*').map(rt => (
                  <option key={rt} value={rt}>{rt}</option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Resource ID (optional)</label>
              <input
                className="input w-full"
                value={resourceId}
                onChange={e => setResourceId(e.target.value)}
                placeholder="UUID"
              />
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Client IP (optional)</label>
              <input
                className="input w-full"
                value={ip}
                onChange={e => setIp(e.target.value)}
                placeholder="e.g. 192.168.1.100"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">MFA</label>
              <label className="flex items-center gap-2 mt-2">
                <input
                  type="checkbox"
                  checked={mfaVerified}
                  onChange={e => setMfaVerified(e.target.checked)}
                  className="rounded border-gray-300 text-indigo-600"
                />
                <span className="text-sm text-gray-600">MFA Verified</span>
              </label>
            </div>
          </div>

          <div className="flex justify-end">
            <button type="submit" disabled={testing || !subjectId} className="btn-primary">
              {testing ? 'Evaluating...' : 'Evaluate Access'}
            </button>
          </div>
        </form>

        {result && (
          <div className="mt-6 pt-4 border-t border-gray-200">
            <h4 className="font-semibold text-gray-900 mb-2">Evaluation Result</h4>
            <div className={`p-4 rounded-lg ${
              result.decision === 'allow'
                ? 'bg-green-50 border border-green-200'
                : 'bg-red-50 border border-red-200'
            }`}>
              <div className="flex items-center gap-3 mb-2">
                <span className={`badge text-sm ${
                  result.decision === 'allow' ? 'badge-low' : 'badge-critical'
                }`}>
                  {String(result.decision || 'ERROR').toUpperCase()}
                </span>
                {result.evaluation_time_us && (
                  <span className="text-xs text-gray-500">
                    Evaluated in {String(result.evaluation_time_us)} us
                  </span>
                )}
              </div>
              {result.reason && (
                <p className="text-sm text-gray-700">{String(result.reason)}</p>
              )}
              {result.policy_name && (
                <p className="text-xs text-gray-500 mt-1">
                  Matched policy: {String(result.policy_name)}
                </p>
              )}
              {result.error && (
                <p className="text-sm text-red-600">{String(result.error)}</p>
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
