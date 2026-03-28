'use client';

import { useEffect, useState } from 'react';
import api from '@/lib/api';
import type { Policy } from '@/types';

export default function PoliciesPage() {
  const [policies, setPolicies] = useState<Policy[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    api.getPolicies()
      .then((res) => setPolicies(res.data?.data || []))
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <div className="flex items-center justify-center h-64"><p className="text-gray-500">Loading policies...</p></div>;
  if (error) return <div className="rounded-lg bg-red-50 border border-red-200 p-4 text-red-700">{error}</div>;

  const overdue = policies.filter(p => p.review_status === 'overdue').length;

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Policy Management</h1>
          <p className="text-gray-500 mt-1">{policies.length} policies</p>
        </div>
        <button className="btn-primary">+ Draft Policy</button>
      </div>

      {overdue > 0 && (
        <div className="mb-6 rounded-lg bg-amber-50 border border-amber-200 p-4 flex items-center gap-3">
          <span className="text-amber-600 text-xl">📋</span>
          <p className="text-amber-800 font-medium">{overdue} polic{overdue > 1 ? 'ies' : 'y'} overdue for review</p>
        </div>
      )}

      {/* Stats Row */}
      <div className="grid grid-cols-4 gap-4 mb-6">
        <div className="card p-4 text-center"><p className="text-2xl font-bold text-green-600">{policies.filter(p => p.status === 'published').length}</p><p className="text-xs text-gray-500 mt-1">Published</p></div>
        <div className="card p-4 text-center"><p className="text-2xl font-bold text-blue-600">{policies.filter(p => p.status === 'draft').length}</p><p className="text-xs text-gray-500 mt-1">Draft</p></div>
        <div className="card p-4 text-center"><p className="text-2xl font-bold text-amber-600">{policies.filter(p => p.status === 'under_review').length}</p><p className="text-xs text-gray-500 mt-1">Under Review</p></div>
        <div className="card p-4 text-center"><p className="text-2xl font-bold text-red-600">{overdue}</p><p className="text-xs text-gray-500 mt-1">Review Overdue</p></div>
      </div>

      {/* Policy Table */}
      <div className="card overflow-hidden p-0">
        <table className="w-full">
          <thead>
            <tr className="border-b border-gray-100">
              <th className="table-header px-4 py-3">Ref</th>
              <th className="table-header px-4 py-3">Policy Title</th>
              <th className="table-header px-4 py-3">Status</th>
              <th className="table-header px-4 py-3">Version</th>
              <th className="table-header px-4 py-3">Review Status</th>
              <th className="table-header px-4 py-3">Next Review</th>
              <th className="table-header px-4 py-3">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-50">
            {policies.map(pol => (
              <tr key={pol.id} className={`hover:bg-gray-50 ${pol.review_status === 'overdue' ? 'bg-amber-50' : ''}`}>
                <td className="px-4 py-3 text-sm font-mono text-gray-500">{pol.policy_ref}</td>
                <td className="px-4 py-3">
                  <a href={`/policies/${pol.id}`} className="text-sm font-medium text-gray-900 hover:text-indigo-600">{pol.title}</a>
                  <span className="ml-2 badge badge-info text-xs">{pol.classification}</span>
                </td>
                <td className="px-4 py-3">
                  <span className={`badge ${pol.status === 'published' ? 'badge-low' : pol.status === 'draft' ? 'badge-info' : 'badge-medium'}`}>
                    {pol.status.replace('_', ' ')}
                  </span>
                </td>
                <td className="px-4 py-3 text-sm text-gray-600">v{pol.current_version}.0</td>
                <td className="px-4 py-3">
                  <span className={`badge ${pol.review_status === 'overdue' ? 'badge-critical' : pol.review_status === 'review_due' ? 'badge-medium' : 'badge-low'}`}>
                    {pol.review_status.replace('_', ' ')}
                  </span>
                </td>
                <td className="px-4 py-3 text-sm text-gray-500">{pol.next_review_date || '—'}</td>
                <td className="px-4 py-3">
                  <a href={`/policies/${pol.id}`} className="text-sm text-indigo-600 hover:underline">View</a>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
