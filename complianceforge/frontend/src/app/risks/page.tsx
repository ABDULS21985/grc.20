'use client';

import { useEffect, useState } from 'react';
import api from '@/lib/api';
import type { Risk } from '@/types';

export default function RisksPage() {
  const [risks, setRisks] = useState<Risk[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [sortBy, setSortBy] = useState('residual_risk_score');
  const [filter, setFilter] = useState('all');

  useEffect(() => {
    api.getRisks()
      .then((res) => setRisks(res.data?.data || []))
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <div className="flex items-center justify-center h-64"><p className="text-gray-500">Loading risks...</p></div>;
  if (error) return <div className="rounded-lg bg-red-50 border border-red-200 p-4 text-red-700">{error}</div>;

  const filtered = risks
    .filter(r => filter === 'all' || r.residual_risk_level === filter)
    .sort((a, b) => b.residual_risk_score - a.residual_risk_score);

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Risk Register</h1>
          <p className="text-gray-500 mt-1">{risks.length} risks across all categories</p>
        </div>
        <button className="btn-primary">+ Register Risk</button>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-5 gap-3 mb-6">
        {[
          { label: 'Critical', count: risks.filter(r => r.residual_risk_level === 'critical').length, color: 'bg-red-600' },
          { label: 'High', count: risks.filter(r => r.residual_risk_level === 'high').length, color: 'bg-orange-500' },
          { label: 'Medium', count: risks.filter(r => r.residual_risk_level === 'medium').length, color: 'bg-yellow-500' },
          { label: 'Low', count: risks.filter(r => r.residual_risk_level === 'low').length, color: 'bg-green-500' },
          { label: 'Total', count: risks.length, color: 'bg-indigo-600' },
        ].map(s => (
          <button key={s.label} onClick={() => setFilter(s.label === 'Total' ? 'all' : s.label.toLowerCase())}
            className={`rounded-lg p-3 text-center transition-all ${filter === s.label.toLowerCase() || (s.label === 'Total' && filter === 'all') ? 'ring-2 ring-indigo-500' : ''}`}>
            <div className={`inline-flex h-8 w-8 items-center justify-center rounded-full ${s.color} text-white text-sm font-bold`}>{s.count}</div>
            <p className="text-xs text-gray-600 mt-1">{s.label}</p>
          </button>
        ))}
      </div>

      {/* Risk Table */}
      <div className="card overflow-hidden p-0">
        <table className="w-full">
          <thead>
            <tr className="border-b border-gray-100">
              <th className="table-header px-4 py-3">Ref</th>
              <th className="table-header px-4 py-3">Risk Title</th>
              <th className="table-header px-4 py-3">Category</th>
              <th className="table-header px-4 py-3 cursor-pointer hover:text-indigo-600" onClick={() => setSortBy('residual_risk_score')}>Score ↓</th>
              <th className="table-header px-4 py-3">Level</th>
              <th className="table-header px-4 py-3">Status</th>
              <th className="table-header px-4 py-3">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-50">
            {filtered.map(risk => (
              <tr key={risk.id} className="hover:bg-gray-50 transition-colors">
                <td className="px-4 py-3 text-sm font-mono text-gray-500">{risk.risk_ref}</td>
                <td className="px-4 py-3">
                  <a href={`/risks/${risk.id}`} className="text-sm font-medium text-gray-900 hover:text-indigo-600">{risk.title}</a>
                </td>
                <td className="px-4 py-3 text-sm text-gray-600">{risk.risk_source}</td>
                <td className="px-4 py-3">
                  <span className={`text-sm font-bold ${risk.residual_risk_score >= 20 ? 'text-red-600' : risk.residual_risk_score >= 12 ? 'text-orange-600' : risk.residual_risk_score >= 6 ? 'text-yellow-600' : 'text-green-600'}`}>
                    {risk.residual_risk_score}
                  </span>
                </td>
                <td className="px-4 py-3">
                  <span className={`badge badge-${risk.residual_risk_level}`}>{risk.residual_risk_level}</span>
                </td>
                <td className="px-4 py-3">
                  <span className="badge badge-info">{risk.status}</span>
                </td>
                <td className="px-4 py-3">
                  <a href={`/risks/${risk.id}`} className="text-sm text-indigo-600 hover:underline">View</a>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
