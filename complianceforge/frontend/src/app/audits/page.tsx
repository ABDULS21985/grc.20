'use client';

import { useEffect, useState } from 'react';
import api from '@/lib/api';
import type { Audit } from '@/types';

interface FindingsStats {
  total_findings: number;
  critical_findings: number;
  high_findings: number;
  open_findings: number;
  overdue_findings: number;
}

export default function AuditsPage() {
  const [audits, setAudits] = useState<Audit[]>([]);
  const [stats, setStats] = useState<FindingsStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    Promise.all([
      api.getAudits(),
      api.getFindingsStats(),
    ])
      .then(([auditsRes, statsRes]) => {
        setAudits(auditsRes.data?.data || []);
        setStats(statsRes.data);
      })
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, []);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <p className="text-gray-500">Loading audits...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="rounded-lg bg-red-50 border border-red-200 p-4 text-red-700">{error}</div>
    );
  }

  const planned = audits.filter((a) => a.status === 'planned').length;
  const inProgress = audits.filter((a) => a.status === 'in_progress').length;
  const completed = audits.filter((a) => a.status === 'completed').length;

  return (
    <div>
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Audit Management</h1>
          <p className="text-gray-500 mt-1">{audits.length} audits across all programmes</p>
        </div>
        <button className="btn-primary">+ Plan Audit</button>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-5 gap-4 mb-6">
        <div className="card p-4 text-center">
          <p className="text-2xl font-bold text-blue-600">{planned}</p>
          <p className="text-xs text-gray-500 mt-1">Planned</p>
        </div>
        <div className="card p-4 text-center">
          <p className="text-2xl font-bold text-amber-600">{inProgress}</p>
          <p className="text-xs text-gray-500 mt-1">In Progress</p>
        </div>
        <div className="card p-4 text-center">
          <p className="text-2xl font-bold text-green-600">{completed}</p>
          <p className="text-xs text-gray-500 mt-1">Completed</p>
        </div>
        <div className="card p-4 text-center">
          <p className="text-2xl font-bold text-indigo-600">{stats?.total_findings ?? 0}</p>
          <p className="text-xs text-gray-500 mt-1">Total Findings</p>
        </div>
        <div className="card p-4 text-center">
          <p className="text-2xl font-bold text-red-600">{stats?.critical_findings ?? 0}</p>
          <p className="text-xs text-gray-500 mt-1">Critical Findings</p>
        </div>
      </div>

      {/* Audits Table */}
      {audits.length === 0 ? (
        <div className="rounded-lg bg-gray-50 border border-gray-200 p-8 text-center">
          <p className="text-gray-500">No audits found.</p>
          <p className="text-sm text-gray-400 mt-1">Plan your first audit to get started.</p>
        </div>
      ) : (
        <div className="card overflow-hidden p-0">
          <table className="w-full">
            <thead>
              <tr className="border-b border-gray-100">
                <th className="table-header px-4 py-3">Ref</th>
                <th className="table-header px-4 py-3">Title</th>
                <th className="table-header px-4 py-3">Type</th>
                <th className="table-header px-4 py-3">Status</th>
                <th className="table-header px-4 py-3">Start Date</th>
                <th className="table-header px-4 py-3">End Date</th>
                <th className="table-header px-4 py-3">Findings</th>
                <th className="table-header px-4 py-3">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-50">
              {audits.map((audit) => (
                <tr key={audit.id} className="hover:bg-gray-50 transition-colors">
                  <td className="px-4 py-3 text-sm font-mono text-gray-500">{audit.audit_ref}</td>
                  <td className="px-4 py-3">
                    <a href={`/audits/${audit.id}`} className="text-sm font-medium text-gray-900 hover:text-indigo-600">
                      {audit.title}
                    </a>
                  </td>
                  <td className="px-4 py-3">
                    <span className="badge badge-info">{audit.audit_type}</span>
                  </td>
                  <td className="px-4 py-3">
                    <span className={`badge ${
                      audit.status === 'completed' ? 'badge-low' :
                      audit.status === 'in_progress' ? 'badge-medium' :
                      audit.status === 'cancelled' ? 'badge-critical' : 'badge-info'
                    }`}>
                      {audit.status.replace(/_/g, ' ')}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-sm text-gray-500">
                    {audit.planned_start_date ? new Date(audit.planned_start_date).toLocaleDateString('en-GB') : '—'}
                  </td>
                  <td className="px-4 py-3 text-sm text-gray-500">
                    {audit.planned_end_date ? new Date(audit.planned_end_date).toLocaleDateString('en-GB') : '—'}
                  </td>
                  <td className="px-4 py-3">
                    <span className="text-sm text-gray-700">{audit.total_findings}</span>
                    {audit.critical_findings > 0 && (
                      <span className="ml-1 badge badge-critical text-xs">{audit.critical_findings} critical</span>
                    )}
                  </td>
                  <td className="px-4 py-3">
                    <a href={`/audits/${audit.id}`} className="text-sm text-indigo-600 hover:underline">View</a>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
