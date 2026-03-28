'use client';

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import api from '@/lib/api';
import type { Policy } from '@/types';

const TABS = ['Content', 'Versions', 'Attestations'] as const;
type Tab = typeof TABS[number];

interface PolicyDetail extends Policy {
  content?: string;
  approved_by?: string;
  approved_at?: string;
  owner_name?: string;
  department?: string;
  versions?: PolicyVersion[];
  attestations?: PolicyAttestation[];
}

interface PolicyVersion {
  version: number;
  change_summary: string;
  created_by: string;
  created_at: string;
  status: string;
}

interface PolicyAttestation {
  id: string;
  user_name: string;
  user_email: string;
  attested_at: string;
  status: string;
}

export default function PolicyDetailPage() {
  const params = useParams();
  const id = params.id as string;

  const [policy, setPolicy] = useState<PolicyDetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [tab, setTab] = useState<Tab>('Content');
  const [actionLoading, setActionLoading] = useState(false);
  const [actionMessage, setActionMessage] = useState<string | null>(null);

  useEffect(() => {
    api.getPolicy(id)
      .then((res) => setPolicy(res.data))
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, [id]);

  async function handlePublish() {
    setActionLoading(true);
    setActionMessage(null);
    try {
      await api.publishPolicy(id);
      setActionMessage('Policy published successfully.');
      // Refresh
      const res = await api.getPolicy(id);
      setPolicy(res.data);
    } catch (err: unknown) {
      setActionMessage(`Failed to publish: ${err instanceof Error ? err.message : 'Unknown error'}`);
    } finally {
      setActionLoading(false);
    }
  }

  async function handleAttest() {
    setActionLoading(true);
    setActionMessage(null);
    try {
      await api.attestPolicy(id);
      setActionMessage('Attestation recorded successfully.');
      const res = await api.getPolicy(id);
      setPolicy(res.data);
    } catch (err: unknown) {
      setActionMessage(`Failed to attest: ${err instanceof Error ? err.message : 'Unknown error'}`);
    } finally {
      setActionLoading(false);
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <p className="text-gray-500">Loading policy...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="rounded-lg bg-red-50 border border-red-200 p-4 text-red-700">{error}</div>
    );
  }

  if (!policy) {
    return (
      <div className="rounded-lg bg-gray-50 border border-gray-200 p-4 text-gray-600">
        Policy not found.
      </div>
    );
  }

  const statusColor: Record<string, string> = {
    draft: 'badge-info',
    under_review: 'badge-medium',
    pending_approval: 'badge-medium',
    approved: 'badge-low',
    published: 'badge-low',
    archived: 'badge-info',
    retired: 'badge-info',
  };

  return (
    <div>
      {/* Back button */}
      <a href="/policies" className="inline-flex items-center text-sm text-gray-500 hover:text-indigo-600 mb-4">
        &larr; Back to Policies
      </a>

      {/* Header */}
      <div className="flex items-start justify-between mb-6">
        <div>
          <div className="flex items-center gap-3 mb-1">
            <span className="text-sm font-mono text-gray-400">{policy.policy_ref}</span>
            <span className={`badge ${statusColor[policy.status] || 'badge-info'}`}>{policy.status.replace(/_/g, ' ')}</span>
            <span className="badge badge-info">{policy.classification}</span>
            {policy.is_mandatory && <span className="badge badge-critical">Mandatory</span>}
          </div>
          <h1 className="text-2xl font-bold text-gray-900">{policy.title}</h1>
          <p className="text-sm text-gray-500 mt-1">
            Version {policy.current_version}.0 &middot; Review every {policy.review_frequency_months} months
            {policy.next_review_date && ` &middot; Next review: ${new Date(policy.next_review_date).toLocaleDateString('en-GB')}`}
          </p>
        </div>

        {/* Action Buttons */}
        <div className="flex gap-2">
          {policy.status === 'approved' && (
            <button onClick={handlePublish} className="btn-primary" disabled={actionLoading}>
              {actionLoading ? 'Publishing...' : 'Publish Policy'}
            </button>
          )}
          {policy.status === 'published' && (
            <button onClick={handleAttest} className="btn-primary" disabled={actionLoading}>
              {actionLoading ? 'Attesting...' : 'Attest Policy'}
            </button>
          )}
        </div>
      </div>

      {/* Action Message */}
      {actionMessage && (
        <div className={`mb-4 rounded-lg p-3 text-sm ${actionMessage.startsWith('Failed') ? 'bg-red-50 border border-red-200 text-red-700' : 'bg-green-50 border border-green-200 text-green-700'}`}>
          {actionMessage}
        </div>
      )}

      {/* Policy Info Cards */}
      <div className="grid grid-cols-4 gap-4 mb-6">
        <div className="card p-4 text-center">
          <p className="text-xs font-medium text-gray-500 uppercase">Status</p>
          <p className="text-sm font-bold text-gray-900 mt-1 capitalize">{policy.status.replace(/_/g, ' ')}</p>
        </div>
        <div className="card p-4 text-center">
          <p className="text-xs font-medium text-gray-500 uppercase">Version</p>
          <p className="text-sm font-bold text-gray-900 mt-1">v{policy.current_version}.0</p>
        </div>
        <div className="card p-4 text-center">
          <p className="text-xs font-medium text-gray-500 uppercase">Review Status</p>
          <p className={`text-sm font-bold mt-1 ${policy.review_status === 'overdue' ? 'text-red-600' : 'text-gray-900'}`}>
            {(policy.review_status || '—').replace(/_/g, ' ')}
          </p>
        </div>
        <div className="card p-4 text-center">
          <p className="text-xs font-medium text-gray-500 uppercase">Attestation</p>
          <p className="text-sm font-bold text-gray-900 mt-1">
            {policy.requires_attestation ? 'Required' : 'Not Required'}
          </p>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex gap-1 border-b border-gray-200 mb-6">
        {TABS.map((t) => (
          <button
            key={t}
            onClick={() => setTab(t)}
            className={`px-4 py-2.5 text-sm font-medium border-b-2 transition-colors ${
              tab === t
                ? 'border-indigo-600 text-indigo-600'
                : 'border-transparent text-gray-500 hover:text-gray-700'
            }`}
          >
            {t}
          </button>
        ))}
      </div>

      {/* Content Tab */}
      {tab === 'Content' && (
        <div className="card">
          {policy.content ? (
            <div className="prose prose-sm max-w-none">
              <div className="whitespace-pre-wrap text-sm text-gray-700 leading-relaxed">
                {policy.content}
              </div>
            </div>
          ) : (
            <div className="rounded-lg bg-gray-50 border border-gray-200 p-8 text-center">
              <p className="text-gray-500">Policy content is not available in the detail response.</p>
              <p className="text-xs text-gray-400 mt-1">Content may be stored separately or requires a dedicated endpoint.</p>
            </div>
          )}
        </div>
      )}

      {/* Versions Tab */}
      {tab === 'Versions' && (
        <div>
          {policy.versions && policy.versions.length > 0 ? (
            <div className="space-y-3">
              {policy.versions.map((v) => (
                <div key={v.version} className="card flex items-start justify-between">
                  <div>
                    <div className="flex items-center gap-2">
                      <span className="text-sm font-bold text-gray-900">Version {v.version}.0</span>
                      <span className={`badge ${v.status === 'published' ? 'badge-low' : 'badge-info'}`}>{v.status}</span>
                    </div>
                    <p className="text-sm text-gray-600 mt-1">{v.change_summary}</p>
                    <p className="text-xs text-gray-400 mt-1">
                      By {v.created_by} on {new Date(v.created_at).toLocaleDateString('en-GB')}
                    </p>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="card">
              <div className="space-y-3">
                <div className="flex items-start gap-3">
                  <div className="h-8 w-8 rounded-full bg-indigo-100 flex items-center justify-center flex-shrink-0">
                    <span className="text-indigo-600 text-xs font-bold">v{policy.current_version}</span>
                  </div>
                  <div>
                    <p className="text-sm font-medium text-gray-900">Version {policy.current_version}.0 (Current)</p>
                    <p className="text-xs text-gray-500 mt-0.5">
                      Created {new Date(policy.created_at).toLocaleDateString('en-GB')}
                    </p>
                    <p className="text-xs text-gray-400 mt-0.5">
                      Status: {policy.status.replace(/_/g, ' ')}
                    </p>
                  </div>
                </div>
              </div>
              {policy.current_version <= 1 && (
                <p className="text-xs text-gray-400 mt-4">This is the initial version of the policy.</p>
              )}
            </div>
          )}
        </div>
      )}

      {/* Attestations Tab */}
      {tab === 'Attestations' && (
        <div>
          {!policy.requires_attestation ? (
            <div className="rounded-lg bg-gray-50 border border-gray-200 p-8 text-center">
              <p className="text-gray-500">This policy does not require attestation.</p>
            </div>
          ) : policy.attestations && policy.attestations.length > 0 ? (
            <div className="card overflow-hidden p-0">
              <table className="w-full">
                <thead>
                  <tr className="border-b border-gray-100">
                    <th className="table-header px-4 py-3">User</th>
                    <th className="table-header px-4 py-3">Email</th>
                    <th className="table-header px-4 py-3">Status</th>
                    <th className="table-header px-4 py-3">Attested At</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-50">
                  {policy.attestations.map((a) => (
                    <tr key={a.id} className="hover:bg-gray-50">
                      <td className="px-4 py-3 text-sm font-medium text-gray-900">{a.user_name}</td>
                      <td className="px-4 py-3 text-sm text-gray-600">{a.user_email}</td>
                      <td className="px-4 py-3">
                        <span className={`badge ${a.status === 'attested' ? 'badge-low' : 'badge-medium'}`}>{a.status}</span>
                      </td>
                      <td className="px-4 py-3 text-sm text-gray-500">
                        {new Date(a.attested_at).toLocaleString('en-GB', { dateStyle: 'short', timeStyle: 'short' })}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <div className="rounded-lg bg-amber-50 border border-amber-200 p-8 text-center">
              <p className="text-amber-700 font-medium">No attestations recorded yet.</p>
              <p className="text-sm text-amber-600 mt-1">
                {policy.status === 'published'
                  ? 'Users can now attest to this policy.'
                  : 'Attestation will be available after the policy is published.'}
              </p>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
