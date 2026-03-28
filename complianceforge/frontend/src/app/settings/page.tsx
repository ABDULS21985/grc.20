'use client';

import { useEffect, useState } from 'react';
import api from '@/lib/api';
import type { User, Role } from '@/types';

const TABS = ['Organisation', 'Users', 'Roles', 'Audit Log'];

interface OrgData {
  name: string;
  legal_name: string;
  industry: string;
  country: string;
  timezone: string;
  subscription: string;
  data_residency: string;
  frameworks_adopted: number;
  total_frameworks: number;
  user_count: number;
  user_limit: number;
}

interface AuditLogEntry {
  id: string;
  user_name: string;
  action: string;
  entity_type: string;
  entity_ref: string;
  detail: string;
  created_at: string;
}

export default function SettingsPage() {
  const [tab, setTab] = useState('Organisation');

  return (
    <div>
      <h1 className="text-2xl font-bold text-gray-900 mb-6">Organisation Settings</h1>

      {/* Tabs */}
      <div className="flex gap-1 border-b border-gray-200 mb-6">
        {TABS.map(t => (
          <button key={t} onClick={() => setTab(t)}
            className={`px-4 py-2.5 text-sm font-medium border-b-2 transition-colors ${tab === t ? 'border-indigo-600 text-indigo-600' : 'border-transparent text-gray-500 hover:text-gray-700'}`}>
            {t}
          </button>
        ))}
      </div>

      {tab === 'Organisation' && <OrgTab />}
      {tab === 'Users' && <UsersTab />}
      {tab === 'Roles' && <RolesTab />}
      {tab === 'Audit Log' && <AuditLogTab />}
    </div>
  );
}

function OrgTab() {
  const [org, setOrg] = useState<OrgData | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api.getOrganization()
      .then((res) => setOrg(res.data))
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <p className="text-gray-500">Loading...</p>;
  if (!org) return <p className="text-gray-500">Failed to load organisation details.</p>;

  return (
    <div className="card max-w-2xl">
      <h2 className="text-lg font-semibold text-gray-900 mb-4">Organisation Details</h2>
      <div className="space-y-4">
        <Field label="Organisation Name" value={org.name} />
        <Field label="Legal Name" value={org.legal_name} />
        <Field label="Industry" value={org.industry} />
        <Field label="Country" value={org.country} />
        <Field label="Timezone" value={org.timezone} />
        <Field label="Subscription" value={org.subscription} badge />
        <Field label="Data Residency" value={org.data_residency} />
        <Field label="Frameworks Adopted" value={`${org.frameworks_adopted} / ${org.total_frameworks}`} />
        <Field label="Users" value={`${org.user_count} / ${org.user_limit}`} />
      </div>
      <div className="mt-6 pt-4 border-t border-gray-100">
        <button className="btn-secondary">Edit Settings</button>
      </div>
    </div>
  );
}

function UsersTab() {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api.getUsers()
      .then((res) => setUsers(res.data?.data || []))
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <p className="text-gray-500">Loading users...</p>;

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <input type="text" placeholder="Search users..." className="input w-64" />
        <button className="btn-primary">+ Add User</button>
      </div>
      <div className="card overflow-hidden p-0">
        <table className="w-full">
          <thead>
            <tr className="border-b border-gray-100">
              <th className="table-header px-4 py-3">User</th>
              <th className="table-header px-4 py-3">Role</th>
              <th className="table-header px-4 py-3">Department</th>
              <th className="table-header px-4 py-3">Status</th>
              <th className="table-header px-4 py-3">Last Login</th>
              <th className="table-header px-4 py-3">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-50">
            {users.map(u => (
              <tr key={u.id} className="hover:bg-gray-50">
                <td className="px-4 py-3">
                  <div className="flex items-center gap-3">
                    <div className="h-8 w-8 rounded-full bg-indigo-100 flex items-center justify-center">
                      <span className="text-indigo-600 text-xs font-medium">{u.first_name[0]}{u.last_name[0]}</span>
                    </div>
                    <div>
                      <p className="text-sm font-medium text-gray-900">{u.first_name} {u.last_name}</p>
                      <p className="text-xs text-gray-500">{u.email}</p>
                    </div>
                  </div>
                </td>
                <td className="px-4 py-3"><span className="badge badge-info">{u.roles?.[0]?.name || '—'}</span></td>
                <td className="px-4 py-3 text-sm text-gray-600">{u.department}</td>
                <td className="px-4 py-3">
                  <span className={`badge ${u.status === 'active' ? 'badge-low' : 'badge-medium'}`}>{u.status.replace('_', ' ')}</span>
                </td>
                <td className="px-4 py-3 text-sm text-gray-500">
                  {u.last_login_at ? new Date(u.last_login_at).toLocaleString('en-GB', { dateStyle: 'short', timeStyle: 'short' }) : '—'}
                </td>
                <td className="px-4 py-3">
                  <button className="text-sm text-indigo-600 hover:underline mr-3">Edit</button>
                  <button className="text-sm text-red-600 hover:underline">Deactivate</button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function RolesTab() {
  const [roles, setRoles] = useState<Role[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api.getRoles()
      .then((res) => setRoles(res.data?.data || res.data || []))
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <p className="text-gray-500">Loading roles...</p>;

  return (
    <div className="space-y-3">
      {roles.map(r => (
        <div key={r.id} className="card flex items-center justify-between">
          <div>
            <div className="flex items-center gap-2">
              <h3 className="font-medium text-gray-900">{r.name}</h3>
              {r.is_system_role && <span className="badge badge-info text-xs">System</span>}
            </div>
            <p className="text-sm text-gray-500 mt-0.5">{r.description}</p>
          </div>
          <div className="text-right">
            <button className="text-xs text-indigo-600 hover:underline mt-1">View Permissions</button>
          </div>
        </div>
      ))}
    </div>
  );
}

function AuditLogTab() {
  const [logs, setLogs] = useState<AuditLogEntry[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api.getAuditLog()
      .then((res) => setLogs(res.data?.data || []))
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <p className="text-gray-500">Loading audit log...</p>;

  return (
    <div>
      <p className="text-sm text-gray-500 mb-4">Immutable audit trail — ISO 27001 A.8.15 compliant</p>
      <div className="card overflow-hidden p-0">
        <table className="w-full">
          <thead>
            <tr className="border-b border-gray-100">
              <th className="table-header px-4 py-3">Time</th>
              <th className="table-header px-4 py-3">User</th>
              <th className="table-header px-4 py-3">Action</th>
              <th className="table-header px-4 py-3">Entity</th>
              <th className="table-header px-4 py-3">Details</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-50">
            {logs.map(log => (
              <tr key={log.id} className="hover:bg-gray-50">
                <td className="px-4 py-3 text-xs text-gray-500 font-mono whitespace-nowrap">
                  {new Date(log.created_at).toLocaleString('en-GB', { dateStyle: 'short', timeStyle: 'medium' })}
                </td>
                <td className="px-4 py-3 text-sm text-gray-700">{log.user_name}</td>
                <td className="px-4 py-3">
                  <span className={`badge ${log.action.includes('create') || log.action.includes('upload') ? 'badge-low' : log.action.includes('login') ? 'badge-info' : 'badge-medium'}`}>
                    {log.action}
                  </span>
                </td>
                <td className="px-4 py-3 text-sm font-mono text-gray-600">{log.entity_ref}</td>
                <td className="px-4 py-3 text-sm text-gray-600">{log.detail}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function Field({ label, value, badge }: { label: string; value: string; badge?: boolean }) {
  return (
    <div className="flex items-center justify-between py-2 border-b border-gray-50">
      <span className="text-sm text-gray-500">{label}</span>
      {badge ? <span className="badge badge-info">{value}</span> : <span className="text-sm font-medium text-gray-900">{value}</span>}
    </div>
  );
}
