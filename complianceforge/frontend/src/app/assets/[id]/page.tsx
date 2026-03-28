'use client';

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
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
  owner_department: string;
  location: string;
  description: string;
  processes_personal_data: boolean;
  data_types?: string[];
  retention_period?: string;
  legal_basis?: string;
  status: string;
  created_at: string;
  updated_at: string;
}

export default function AssetDetailPage() {
  const params = useParams();
  const id = params.id as string;

  const [asset, setAsset] = useState<Asset | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    // The API exposes getAssets() for lists. We attempt to fetch a single asset
    // by fetching the list and filtering, or using a direct endpoint if available.
    api.getAssets(1, 100)
      .then((res) => {
        const all: Asset[] = res.data?.data || [];
        const found = all.find((a) => a.id === id);
        if (found) {
          setAsset(found);
        } else {
          setError('Asset not found in the current page of results.');
        }
      })
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, [id]);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <p className="text-gray-500">Loading asset...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div>
        <a href="/assets" className="inline-flex items-center text-sm text-gray-500 hover:text-indigo-600 mb-4">
          &larr; Back to Assets
        </a>
        <div className="rounded-lg bg-red-50 border border-red-200 p-4 text-red-700">{error}</div>
      </div>
    );
  }

  if (!asset) {
    return (
      <div>
        <a href="/assets" className="inline-flex items-center text-sm text-gray-500 hover:text-indigo-600 mb-4">
          &larr; Back to Assets
        </a>
        <div className="rounded-lg bg-gray-50 border border-gray-200 p-4 text-gray-600">
          Asset not found.
        </div>
      </div>
    );
  }

  const critColor: Record<string, string> = {
    critical: 'badge-critical',
    high: 'badge-high',
    medium: 'badge-medium',
    low: 'badge-low',
  };

  return (
    <div>
      {/* Back button */}
      <a href="/assets" className="inline-flex items-center text-sm text-gray-500 hover:text-indigo-600 mb-4">
        &larr; Back to Assets
      </a>

      {/* Header */}
      <div className="flex items-start justify-between mb-6">
        <div>
          <div className="flex items-center gap-3 mb-1">
            <span className="text-sm font-mono text-gray-400">{asset.asset_ref}</span>
            <span className="badge badge-info">{asset.asset_type}</span>
            <span className={`badge ${critColor[asset.criticality] || 'badge-info'}`}>{asset.criticality}</span>
            {asset.processes_personal_data && <span className="badge badge-high">Personal Data</span>}
          </div>
          <h1 className="text-2xl font-bold text-gray-900">{asset.name}</h1>
          {asset.description && (
            <p className="text-sm text-gray-500 mt-1">{asset.description}</p>
          )}
        </div>
      </div>

      {/* GDPR ROPA Notice */}
      {asset.processes_personal_data && (
        <div className="mb-6 rounded-xl border-2 border-purple-300 bg-purple-50 p-5">
          <div className="flex items-center gap-2 mb-2">
            <span className="text-xl">&#128203;</span>
            <h2 className="font-bold text-purple-800">GDPR Record of Processing Activities (ROPA)</h2>
          </div>
          <p className="text-sm text-purple-700">
            This asset processes personal data and must be included in your Record of Processing Activities
            per GDPR Article 30. Ensure appropriate legal basis, data types, and retention periods are documented.
          </p>
        </div>
      )}

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        {/* Asset Details */}
        <div className="card">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Asset Details</h2>
          <div className="space-y-3">
            <DetailRow label="Name" value={asset.name} />
            <DetailRow label="Type" value={asset.asset_type} />
            <DetailRow label="Category" value={asset.category || '—'} />
            <DetailRow label="Criticality" value={asset.criticality} />
            <DetailRow label="Classification" value={asset.classification || '—'} />
            <DetailRow label="Status" value={asset.status || '—'} />
            <DetailRow label="Location" value={asset.location || '—'} />
          </div>
        </div>

        {/* Ownership */}
        <div className="card">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Ownership</h2>
          <div className="space-y-3">
            <DetailRow label="Owner" value={asset.owner_name || '—'} />
            <DetailRow label="Department" value={asset.owner_department || '—'} />
            <DetailRow
              label="Created"
              value={asset.created_at ? new Date(asset.created_at).toLocaleDateString('en-GB') : '—'}
            />
            <DetailRow
              label="Last Updated"
              value={asset.updated_at ? new Date(asset.updated_at).toLocaleDateString('en-GB') : '—'}
            />
          </div>
        </div>

        {/* Personal Data Processing */}
        {asset.processes_personal_data && (
          <div className="card lg:col-span-2">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">Personal Data Processing</h2>
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
              <div>
                <p className="text-xs font-medium text-gray-500 uppercase mb-2">Data Types</p>
                {asset.data_types && asset.data_types.length > 0 ? (
                  <div className="flex flex-wrap gap-1">
                    {asset.data_types.map((dt) => (
                      <span key={dt} className="badge badge-info text-xs">{dt}</span>
                    ))}
                  </div>
                ) : (
                  <p className="text-sm text-gray-400">Not specified</p>
                )}
              </div>
              <div>
                <p className="text-xs font-medium text-gray-500 uppercase mb-2">Legal Basis</p>
                <p className="text-sm text-gray-700">{asset.legal_basis || 'Not specified'}</p>
              </div>
              <div>
                <p className="text-xs font-medium text-gray-500 uppercase mb-2">Retention Period</p>
                <p className="text-sm text-gray-700">{asset.retention_period || 'Not specified'}</p>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

function DetailRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between py-2 border-b border-gray-50">
      <span className="text-sm text-gray-500">{label}</span>
      <span className="text-sm font-medium text-gray-900 capitalize">{value}</span>
    </div>
  );
}
