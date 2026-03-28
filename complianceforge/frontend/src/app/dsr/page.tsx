'use client';

import { useEffect, useState } from 'react';
import api from '@/lib/api';
import type { DSRRequest, DSRDashboard } from '@/types';

const TYPE_LABELS: Record<string, string> = {
  access: 'Access (Art.15)',
  erasure: 'Erasure (Art.17)',
  rectification: 'Rectification (Art.16)',
  portability: 'Portability (Art.20)',
  restriction: 'Restriction (Art.18)',
  objection: 'Objection (Art.21)',
  automated_decision: 'Automated Decision (Art.22)',
};

const STATUS_COLORS: Record<string, string> = {
  received: 'badge-info',
  identity_verification: 'badge-medium',
  in_progress: 'badge-info',
  extended: 'badge-high',
  completed: 'badge-low',
  rejected: 'badge-critical',
  withdrawn: 'badge-medium',
};

const SLA_COLORS: Record<string, string> = {
  on_track: 'text-green-600',
  at_risk: 'text-amber-600',
  overdue: 'text-red-600',
};

export default function DSRPage() {
  const [requests, setRequests] = useState<DSRRequest[]>([]);
  const [dashboard, setDashboard] = useState<DSRDashboard | null>(null);
  const [loading, setLoading] = useState(true);
  const [showCreate, setShowCreate] = useState(false);

  useEffect(() => {
    Promise.all([
      api.getDSRRequests().then(res => setRequests(res.data?.data || [])),
      api.getDSRDashboard().then(res => setDashboard(res.data)),
    ]).finally(() => setLoading(false));
  }, []);

  if (loading) {
    return (
      <div>
        <h1 className="text-2xl font-bold text-gray-900 mb-8">GDPR Data Subject Requests</h1>
        <div className="grid grid-cols-1 gap-4 md:grid-cols-4 mb-8">
          {Array.from({ length: 4 }).map((_, i) => <div key={i} className="card animate-pulse h-24" />)}
        </div>
        <div className="card animate-pulse h-96" />
      </div>
    );
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">GDPR Data Subject Requests</h1>
          <p className="text-gray-500 mt-1">Manage DSAR, erasure, rectification, and portability requests per GDPR Articles 12–23</p>
        </div>
        <button onClick={() => setShowCreate(true)} className="btn-primary">New DSR Request</button>
      </div>

      {dashboard && (
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-4 mb-8">
          <div className="card">
            <p className="text-sm text-gray-500">Total Requests</p>
            <p className="text-3xl font-bold text-gray-900 mt-1">{dashboard.total_requests}</p>
          </div>
          <div className="card">
            <p className="text-sm text-gray-500">SLA Compliance</p>
            <p className="text-3xl font-bold text-green-600 mt-1">{dashboard.sla_compliance_rate}%</p>
          </div>
          <div className="card">
            <p className="text-sm text-gray-500">Overdue</p>
            <p className="text-3xl font-bold text-red-600 mt-1">{dashboard.overdue_count}</p>
          </div>
          <div className="card">
            <p className="text-sm text-gray-500">Avg Completion</p>
            <p className="text-3xl font-bold text-gray-900 mt-1">{dashboard.avg_completion_days}d</p>
          </div>
        </div>
      )}

      <div className="card overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="table-header">
              <th className="px-4 py-3">Reference</th>
              <th className="px-4 py-3">Type</th>
              <th className="px-4 py-3">Data Subject</th>
              <th className="px-4 py-3">Status</th>
              <th className="px-4 py-3">Received</th>
              <th className="px-4 py-3">Deadline</th>
              <th className="px-4 py-3">Days Left</th>
              <th className="px-4 py-3">SLA</th>
            </tr>
          </thead>
          <tbody>
            {requests.map(req => (
              <tr key={req.id} className="border-t border-gray-100 hover:bg-gray-50">
                <td className="px-4 py-3">
                  <a href={`/dsr/${req.id}`} className="font-medium text-indigo-600 hover:text-indigo-700">{req.request_ref}</a>
                </td>
                <td className="px-4 py-3">
                  <span className="badge badge-info">{TYPE_LABELS[req.request_type] || req.request_type}</span>
                </td>
                <td className="px-4 py-3 text-gray-700">
                  {req.data_subject_name ? req.data_subject_name.replace(/(?<=.).(?=.*@)/g, '*') : '—'}
                </td>
                <td className="px-4 py-3">
                  <span className={`badge ${STATUS_COLORS[req.status] || 'badge-info'}`}>{req.status.replace(/_/g, ' ')}</span>
                </td>
                <td className="px-4 py-3 text-gray-600">{new Date(req.received_date).toLocaleDateString('en-GB')}</td>
                <td className="px-4 py-3 text-gray-600">
                  {new Date(req.extended_deadline || req.response_deadline).toLocaleDateString('en-GB')}
                </td>
                <td className="px-4 py-3">
                  <span className={`font-semibold ${req.days_remaining <= 0 ? 'text-red-600' : req.days_remaining <= 7 ? 'text-amber-600' : 'text-green-600'}`}>
                    {req.days_remaining <= 0 ? 'Overdue' : `${req.days_remaining}d`}
                  </span>
                </td>
                <td className="px-4 py-3">
                  <span className={`font-medium ${SLA_COLORS[req.sla_status] || ''}`}>
                    {req.sla_status?.replace(/_/g, ' ')}
                  </span>
                </td>
              </tr>
            ))}
            {requests.length === 0 && (
              <tr><td colSpan={8} className="px-4 py-12 text-center text-gray-500">No DSR requests yet</td></tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
