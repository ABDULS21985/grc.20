'use client';

import { useEffect, useState } from 'react';
import api from '@/lib/api';
import type { Incident } from '@/types';

export default function IncidentsPage() {
  const [incidents, setIncidents] = useState<Incident[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    api.getIncidents()
      .then((res) => setIncidents(res.data?.data || []))
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <div className="flex items-center justify-center h-64"><p className="text-gray-500">Loading incidents...</p></div>;
  if (error) return <div className="rounded-lg bg-red-50 border border-red-200 p-4 text-red-700">{error}</div>;

  const breaches = incidents.filter(i => i.is_data_breach && !i.dpa_notified_at);

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Incident Management</h1>
          <p className="text-gray-500 mt-1">GDPR Article 33 &amp; NIS2 Article 23 compliant</p>
        </div>
        <button className="btn-danger">🚨 Report Incident</button>
      </div>

      {/* GDPR Breach Alert */}
      {breaches.length > 0 && (
        <div className="mb-6 rounded-lg border-2 border-red-300 bg-red-50 p-5">
          <div className="flex items-center gap-2 mb-3">
            <span className="text-2xl">⚠️</span>
            <h2 className="text-lg font-bold text-red-800">GDPR Breach Notifications Pending</h2>
          </div>
          {breaches.map(b => (
            <div key={b.id} className="flex items-center justify-between bg-white rounded-lg p-3 mb-2 border border-red-200">
              <div>
                <span className="font-mono text-sm text-red-600">{b.incident_ref}</span>
                <span className="ml-2 text-sm text-gray-900">{b.title}</span>
                <span className="ml-2 badge badge-critical">{b.data_subjects_affected.toLocaleString()} subjects</span>
              </div>
              <div className="flex items-center gap-3">
                <span className="text-sm text-red-600 font-medium">
                  Deadline: {new Date(b.notification_deadline).toLocaleString('en-GB', { dateStyle: 'short', timeStyle: 'short' })}
                </span>
                <button className="btn-danger text-xs py-1 px-3">Notify DPA</button>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Stats */}
      <div className="grid grid-cols-4 gap-4 mb-6">
        <StatCard label="Open Incidents" value={incidents.filter(i => !['resolved', 'closed'].includes(i.status)).length} color="red" />
        <StatCard label="Data Breaches" value={incidents.filter(i => i.is_data_breach).length} color="red" />
        <StatCard label="Resolved" value={incidents.filter(i => i.status === 'resolved').length} color="green" />
        <StatCard label="Total (30 days)" value={incidents.length} color="blue" />
      </div>

      {/* Incident Table */}
      <div className="card overflow-hidden p-0">
        <table className="w-full">
          <thead>
            <tr className="border-b border-gray-100">
              <th className="table-header px-4 py-3">Ref</th>
              <th className="table-header px-4 py-3">Incident</th>
              <th className="table-header px-4 py-3">Type</th>
              <th className="table-header px-4 py-3">Severity</th>
              <th className="table-header px-4 py-3">Status</th>
              <th className="table-header px-4 py-3">Breach</th>
              <th className="table-header px-4 py-3">Reported</th>
              <th className="table-header px-4 py-3">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-50">
            {incidents.map(inc => (
              <tr key={inc.id} className={`hover:bg-gray-50 ${inc.is_data_breach && !inc.dpa_notified_at ? 'bg-red-50' : ''}`}>
                <td className="px-4 py-3 text-sm font-mono text-gray-500">{inc.incident_ref}</td>
                <td className="px-4 py-3 text-sm font-medium text-gray-900">{inc.title}</td>
                <td className="px-4 py-3"><span className="badge badge-info">{inc.incident_type.replace('_', ' ')}</span></td>
                <td className="px-4 py-3"><span className={`badge badge-${inc.severity}`}>{inc.severity}</span></td>
                <td className="px-4 py-3"><span className="badge badge-info">{inc.status}</span></td>
                <td className="px-4 py-3">
                  {inc.is_data_breach ? (
                    <span className="badge badge-critical">🔓 Breach ({inc.data_subjects_affected.toLocaleString()})</span>
                  ) : (
                    <span className="text-gray-400 text-sm">—</span>
                  )}
                </td>
                <td className="px-4 py-3 text-sm text-gray-500">{new Date(inc.reported_at).toLocaleDateString('en-GB')}</td>
                <td className="px-4 py-3">
                  <a href={`/incidents/${inc.id}`} className="text-sm text-indigo-600 hover:underline">View</a>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function StatCard({ label, value, color }: { label: string; value: number; color: string }) {
  const colors = { red: 'text-red-600 bg-red-50', green: 'text-green-600 bg-green-50', blue: 'text-blue-600 bg-blue-50' };
  return (
    <div className={`rounded-lg p-4 ${(colors as any)[color] || colors.blue}`}>
      <p className="text-xs font-medium text-gray-500 uppercase">{label}</p>
      <p className={`text-2xl font-bold mt-1 ${(colors as any)[color]?.split(' ')[0]}`}>{value}</p>
    </div>
  );
}
