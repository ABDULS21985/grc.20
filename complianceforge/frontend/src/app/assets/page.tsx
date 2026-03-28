'use client';

import { useEffect, useState } from 'react';
import api from '@/lib/api';

interface Asset {
  id: string;
  asset_ref: string;
  name: string;
  asset_type: string;
  category: string;
  criticality: string;
  classification: string;
  owner_name: string;
  location: string;
  processes_personal_data: boolean;
  status: string;
  description: string;
}

interface AssetStats {
  total: number;
  critical: number;
  high: number;
  medium: number;
  low: number;
  personal_data: number;
  by_type: Record<string, number>;
}

export default function AssetsPage() {
  const [assets, setAssets] = useState<Asset[]>([]);
  const [stats, setStats] = useState<AssetStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    Promise.all([
      api.getAssets(),
      api.getAssetStats(),
    ])
      .then(([assetsRes, statsRes]) => {
        setAssets(assetsRes.data?.data || []);
        setStats(statsRes.data);
      })
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, []);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <p className="text-gray-500">Loading assets...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="rounded-lg bg-red-50 border border-red-200 p-4 text-red-700">{error}</div>
    );
  }

  const totalAssets = stats?.total ?? assets.length;
  const criticalAssets = stats?.critical ?? assets.filter((a) => a.criticality === 'critical').length;
  const personalDataAssets = stats?.personal_data ?? assets.filter((a) => a.processes_personal_data).length;

  return (
    <div>
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Asset Register</h1>
          <p className="text-gray-500 mt-1">{totalAssets} assets tracked across the organisation</p>
        </div>
        <button className="btn-primary">+ Register Asset</button>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-2 gap-4 sm:grid-cols-4 mb-6">
        <div className="card p-4 text-center">
          <p className="text-2xl font-bold text-indigo-600">{totalAssets}</p>
          <p className="text-xs text-gray-500 mt-1">Total Assets</p>
        </div>
        <div className="card p-4 text-center">
          <p className="text-2xl font-bold text-red-600">{criticalAssets}</p>
          <p className="text-xs text-gray-500 mt-1">Critical</p>
        </div>
        <div className="card p-4 text-center">
          <p className="text-2xl font-bold text-purple-600">{personalDataAssets}</p>
          <p className="text-xs text-gray-500 mt-1">Personal Data</p>
        </div>
        <div className="card p-4 text-center">
          <p className="text-2xl font-bold text-gray-600">
            {stats?.by_type ? Object.keys(stats.by_type).length : '—'}
          </p>
          <p className="text-xs text-gray-500 mt-1">Asset Types</p>
        </div>
      </div>

      {/* Asset Type Breakdown */}
      {stats?.by_type && Object.keys(stats.by_type).length > 0 && (
        <div className="card mb-6">
          <h2 className="text-sm font-semibold text-gray-700 mb-3">Assets by Type</h2>
          <div className="flex flex-wrap gap-3">
            {Object.entries(stats.by_type).map(([type, count]) => (
              <div key={type} className="rounded-lg bg-gray-50 px-3 py-2 text-center">
                <p className="text-sm font-bold text-gray-700">{count}</p>
                <p className="text-xs text-gray-500 capitalize">{type.replace(/_/g, ' ')}</p>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Assets Table */}
      {assets.length === 0 ? (
        <div className="rounded-lg bg-gray-50 border border-gray-200 p-8 text-center">
          <p className="text-gray-500">No assets registered yet.</p>
          <p className="text-sm text-gray-400 mt-1">Register your first asset to begin tracking.</p>
        </div>
      ) : (
        <div className="card overflow-hidden p-0">
          <table className="w-full">
            <thead>
              <tr className="border-b border-gray-100">
                <th className="table-header px-4 py-3">Ref</th>
                <th className="table-header px-4 py-3">Name</th>
                <th className="table-header px-4 py-3">Type</th>
                <th className="table-header px-4 py-3">Criticality</th>
                <th className="table-header px-4 py-3">Classification</th>
                <th className="table-header px-4 py-3">Personal Data</th>
                <th className="table-header px-4 py-3">Owner</th>
                <th className="table-header px-4 py-3">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-50">
              {assets.map((asset) => (
                <tr key={asset.id} className="hover:bg-gray-50 transition-colors">
                  <td className="px-4 py-3 text-sm font-mono text-gray-500">{asset.asset_ref}</td>
                  <td className="px-4 py-3">
                    <a href={`/assets/${asset.id}`} className="text-sm font-medium text-gray-900 hover:text-indigo-600">
                      {asset.name}
                    </a>
                  </td>
                  <td className="px-4 py-3">
                    <span className="badge badge-info">{asset.asset_type}</span>
                  </td>
                  <td className="px-4 py-3">
                    <span className={`badge badge-${asset.criticality === 'critical' ? 'critical' : asset.criticality === 'high' ? 'high' : asset.criticality === 'medium' ? 'medium' : 'low'}`}>
                      {asset.criticality}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-sm text-gray-600">{asset.classification}</td>
                  <td className="px-4 py-3">
                    {asset.processes_personal_data ? (
                      <span className="badge badge-high">Yes</span>
                    ) : (
                      <span className="text-sm text-gray-400">No</span>
                    )}
                  </td>
                  <td className="px-4 py-3 text-sm text-gray-600">{asset.owner_name || '—'}</td>
                  <td className="px-4 py-3">
                    <a href={`/assets/${asset.id}`} className="text-sm text-indigo-600 hover:underline">View</a>
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
