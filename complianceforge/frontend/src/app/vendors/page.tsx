'use client';

import { useEffect, useState } from 'react';
import api from '@/lib/api';
import type { Vendor } from '@/types';

export default function VendorsPage() {
  const [vendors, setVendors] = useState<Vendor[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    api.getVendors()
      .then((res) => setVendors(res.data?.data || []))
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <div className="flex items-center justify-center h-64"><p className="text-gray-500">Loading vendors...</p></div>;
  if (error) return <div className="rounded-lg bg-red-50 border border-red-200 p-4 text-red-700">{error}</div>;

  const missingDPA = vendors.filter(v => v.data_processing && !v.dpa_in_place);
  const criticalVendors = vendors.filter(v => v.risk_tier === 'critical');

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Vendor Risk Management</h1>
          <p className="text-gray-500 mt-1">GDPR Article 28 compliant third-party oversight</p>
        </div>
        <button className="btn-primary">+ Onboard Vendor</button>
      </div>

      {/* DPA Warning */}
      {missingDPA.length > 0 && (
        <div className="mb-6 rounded-lg border-2 border-red-300 bg-red-50 p-4">
          <div className="flex items-center gap-2 mb-2">
            <span className="text-xl">🔓</span>
            <h2 className="font-bold text-red-800">GDPR Compliance Alert: {missingDPA.length} vendor{missingDPA.length > 1 ? 's' : ''} missing Data Processing Agreement</h2>
          </div>
          <p className="text-sm text-red-700 mb-3">GDPR Article 28 requires a DPA before sharing personal data with processors.</p>
          <div className="flex flex-wrap gap-2">
            {missingDPA.map(v => (
              <span key={v.id} className="badge badge-critical">{v.name}</span>
            ))}
          </div>
        </div>
      )}

      {/* Stats */}
      <div className="grid grid-cols-5 gap-4 mb-6">
        <div className="card p-4 text-center">
          <p className="text-2xl font-bold text-indigo-600">{vendors.length}</p>
          <p className="text-xs text-gray-500 mt-1">Total Vendors</p>
        </div>
        <div className="card p-4 text-center">
          <p className="text-2xl font-bold text-red-600">{criticalVendors.length}</p>
          <p className="text-xs text-gray-500 mt-1">Critical Risk</p>
        </div>
        <div className="card p-4 text-center">
          <p className="text-2xl font-bold text-orange-600">{vendors.filter(v => v.risk_tier === 'high').length}</p>
          <p className="text-xs text-gray-500 mt-1">High Risk</p>
        </div>
        <div className="card p-4 text-center">
          <p className="text-2xl font-bold text-red-600">{missingDPA.length}</p>
          <p className="text-xs text-gray-500 mt-1">Missing DPA</p>
        </div>
        <div className="card p-4 text-center">
          <p className="text-2xl font-bold text-gray-600">€{(vendors.reduce((s, v) => s + v.contract_value, 0) / 1000).toFixed(0)}K</p>
          <p className="text-xs text-gray-500 mt-1">Total Value</p>
        </div>
      </div>

      {/* Vendor Table */}
      <div className="card overflow-hidden p-0">
        <table className="w-full">
          <thead>
            <tr className="border-b border-gray-100">
              <th className="table-header px-4 py-3">Vendor</th>
              <th className="table-header px-4 py-3">Country</th>
              <th className="table-header px-4 py-3">Risk Tier</th>
              <th className="table-header px-4 py-3">Score</th>
              <th className="table-header px-4 py-3">Data Processing</th>
              <th className="table-header px-4 py-3">DPA</th>
              <th className="table-header px-4 py-3">Certifications</th>
              <th className="table-header px-4 py-3">Next Assessment</th>
              <th className="table-header px-4 py-3">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-50">
            {vendors.map(v => (
              <tr key={v.id} className={`hover:bg-gray-50 ${v.data_processing && !v.dpa_in_place ? 'bg-red-50' : ''}`}>
                <td className="px-4 py-3">
                  <a href={`/vendors/${v.id}`} className="text-sm font-medium text-gray-900 hover:text-indigo-600">{v.name}</a>
                  <span className="ml-1 text-xs text-gray-400">{v.vendor_ref}</span>
                </td>
                <td className="px-4 py-3 text-sm">{v.country_code}</td>
                <td className="px-4 py-3"><span className={`badge badge-${v.risk_tier}`}>{v.risk_tier}</span></td>
                <td className="px-4 py-3">
                  <span className={`text-sm font-bold ${v.risk_score >= 70 ? 'text-green-600' : v.risk_score >= 50 ? 'text-amber-600' : 'text-red-600'}`}>{v.risk_score}</span>
                </td>
                <td className="px-4 py-3">{v.data_processing ? <span className="badge badge-info">Yes</span> : <span className="text-gray-400 text-sm">No</span>}</td>
                <td className="px-4 py-3">
                  {v.data_processing ? (
                    v.dpa_in_place ? <span className="badge badge-low">✓ In Place</span> : <span className="badge badge-critical">✗ Missing</span>
                  ) : <span className="text-gray-400 text-sm">N/A</span>}
                </td>
                <td className="px-4 py-3">
                  <div className="flex flex-wrap gap-1">
                    {v.certifications.length > 0 ? v.certifications.map(c => (
                      <span key={c} className="badge badge-info text-xs">{c}</span>
                    )) : <span className="text-gray-400 text-xs">None</span>}
                  </div>
                </td>
                <td className="px-4 py-3 text-sm text-gray-500">{v.next_assessment_date || '—'}</td>
                <td className="px-4 py-3">
                  <a href={`/vendors/${v.id}`} className="text-sm text-indigo-600 hover:underline">View</a>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
