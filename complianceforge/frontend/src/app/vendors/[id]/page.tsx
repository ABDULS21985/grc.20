'use client';

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import api from '@/lib/api';
import type { Vendor } from '@/types';

interface VendorDetail extends Vendor {
  description?: string;
  website?: string;
  primary_contact_name?: string;
  primary_contact_email?: string;
  service_category?: string;
  contract_start_date?: string;
  contract_end_date?: string;
  data_types_processed?: string[];
  data_transfer_mechanism?: string;
  sub_processors?: string[];
  last_assessment_date?: string;
  assessment_score?: number;
}

export default function VendorDetailPage() {
  const params = useParams();
  const id = params.id as string;

  const [vendor, setVendor] = useState<VendorDetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    api.getVendor(id)
      .then((res) => setVendor(res.data))
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, [id]);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <p className="text-gray-500">Loading vendor...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="rounded-lg bg-red-50 border border-red-200 p-4 text-red-700">{error}</div>
    );
  }

  if (!vendor) {
    return (
      <div className="rounded-lg bg-gray-50 border border-gray-200 p-4 text-gray-600">
        Vendor not found.
      </div>
    );
  }

  const tierColor: Record<string, string> = {
    critical: 'badge-critical',
    high: 'badge-high',
    medium: 'badge-medium',
    low: 'badge-low',
  };

  return (
    <div>
      {/* Back button */}
      <a href="/vendors" className="inline-flex items-center text-sm text-gray-500 hover:text-indigo-600 mb-4">
        &larr; Back to Vendors
      </a>

      {/* Header */}
      <div className="flex items-start justify-between mb-6">
        <div>
          <div className="flex items-center gap-3 mb-1">
            <span className="text-sm font-mono text-gray-400">{vendor.vendor_ref}</span>
            <span className={`badge ${tierColor[vendor.risk_tier] || 'badge-info'}`}>{vendor.risk_tier} risk</span>
            <span className="badge badge-info">{vendor.status}</span>
            <span className="text-sm text-gray-500">{vendor.country_code}</span>
          </div>
          <h1 className="text-2xl font-bold text-gray-900">{vendor.name}</h1>
          {vendor.description && (
            <p className="text-sm text-gray-500 mt-1">{vendor.description}</p>
          )}
        </div>
        <div className="text-right">
          <p className="text-sm text-gray-500">Risk Score</p>
          <p className={`text-3xl font-bold ${vendor.risk_score >= 70 ? 'text-green-600' : vendor.risk_score >= 50 ? 'text-amber-600' : 'text-red-600'}`}>
            {vendor.risk_score}
          </p>
        </div>
      </div>

      {/* DPA Warning */}
      {vendor.data_processing && !vendor.dpa_in_place && (
        <div className="mb-6 rounded-xl border-2 border-red-300 bg-red-50 p-5">
          <div className="flex items-center gap-2 mb-2">
            <span className="text-xl">&#9888;</span>
            <h2 className="font-bold text-red-800">GDPR Compliance Alert: Data Processing Agreement Missing</h2>
          </div>
          <p className="text-sm text-red-700">
            This vendor processes personal data but does not have a Data Processing Agreement (DPA) in place.
            Under GDPR Article 28, a DPA is legally required before sharing personal data with any processor.
          </p>
        </div>
      )}

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        {/* Overview */}
        <div className="card">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Overview</h2>
          <div className="space-y-3">
            <DetailRow label="Vendor Reference" value={vendor.vendor_ref} />
            <DetailRow label="Country" value={vendor.country_code} />
            <DetailRow label="Status" value={vendor.status} />
            {vendor.website && <DetailRow label="Website" value={vendor.website} />}
            {vendor.service_category && <DetailRow label="Service Category" value={vendor.service_category} />}
            {vendor.primary_contact_name && <DetailRow label="Primary Contact" value={vendor.primary_contact_name} />}
            {vendor.primary_contact_email && <DetailRow label="Contact Email" value={vendor.primary_contact_email} />}
          </div>
        </div>

        {/* Risk Assessment */}
        <div className="card">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Risk Assessment</h2>
          <div className="space-y-3">
            <DetailRow label="Risk Tier" value={vendor.risk_tier} />
            <div className="flex items-center justify-between py-2 border-b border-gray-50">
              <span className="text-sm text-gray-500">Risk Score</span>
              <div className="flex items-center gap-2">
                <div className="w-24 h-2 rounded-full bg-gray-100">
                  <div
                    className={`h-2 rounded-full transition-all ${vendor.risk_score >= 70 ? 'bg-green-500' : vendor.risk_score >= 50 ? 'bg-amber-500' : 'bg-red-500'}`}
                    style={{ width: `${vendor.risk_score}%` }}
                  />
                </div>
                <span className="text-sm font-bold text-gray-900">{vendor.risk_score}/100</span>
              </div>
            </div>
            <DetailRow
              label="Next Assessment"
              value={vendor.next_assessment_date ? new Date(vendor.next_assessment_date).toLocaleDateString('en-GB') : '—'}
            />
            {vendor.last_assessment_date && (
              <DetailRow
                label="Last Assessment"
                value={new Date(vendor.last_assessment_date).toLocaleDateString('en-GB')}
              />
            )}
          </div>
        </div>

        {/* Certifications */}
        <div className="card">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Certifications</h2>
          {vendor.certifications && vendor.certifications.length > 0 ? (
            <div className="flex flex-wrap gap-2">
              {vendor.certifications.map((cert) => (
                <span key={cert} className="badge badge-info">{cert}</span>
              ))}
            </div>
          ) : (
            <p className="text-sm text-gray-500">No certifications recorded.</p>
          )}
        </div>

        {/* Data Processing */}
        <div className="card">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Data Processing</h2>
          <div className="space-y-3">
            <div className="flex items-center justify-between py-2 border-b border-gray-50">
              <span className="text-sm text-gray-500">Processes Personal Data</span>
              {vendor.data_processing ? (
                <span className="badge badge-high">Yes</span>
              ) : (
                <span className="text-sm font-medium text-gray-900">No</span>
              )}
            </div>
            <div className="flex items-center justify-between py-2 border-b border-gray-50">
              <span className="text-sm text-gray-500">DPA In Place</span>
              {vendor.data_processing ? (
                vendor.dpa_in_place ? (
                  <span className="badge badge-low">Yes</span>
                ) : (
                  <span className="badge badge-critical">Missing</span>
                )
              ) : (
                <span className="text-sm font-medium text-gray-400">N/A</span>
              )}
            </div>
            {vendor.data_types_processed && vendor.data_types_processed.length > 0 && (
              <div>
                <p className="text-xs font-medium text-gray-500 uppercase mb-1">Data Types Processed</p>
                <div className="flex flex-wrap gap-1">
                  {vendor.data_types_processed.map((dt) => (
                    <span key={dt} className="badge badge-info text-xs">{dt}</span>
                  ))}
                </div>
              </div>
            )}
            {vendor.data_transfer_mechanism && (
              <DetailRow label="Transfer Mechanism" value={vendor.data_transfer_mechanism} />
            )}
            {vendor.sub_processors && vendor.sub_processors.length > 0 && (
              <div>
                <p className="text-xs font-medium text-gray-500 uppercase mb-1">Sub-Processors</p>
                <div className="flex flex-wrap gap-1">
                  {vendor.sub_processors.map((sp) => (
                    <span key={sp} className="badge badge-info text-xs">{sp}</span>
                  ))}
                </div>
              </div>
            )}
          </div>
        </div>

        {/* Contract Info */}
        <div className="card lg:col-span-2">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Contract Information</h2>
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
            <div className="rounded-lg bg-gray-50 p-4 text-center">
              <p className="text-xs font-medium text-gray-500 uppercase">Contract Value</p>
              <p className="text-xl font-bold text-gray-900 mt-1">
                {vendor.contract_value > 0 ? `€${vendor.contract_value.toLocaleString()}` : '—'}
              </p>
            </div>
            <div className="rounded-lg bg-gray-50 p-4 text-center">
              <p className="text-xs font-medium text-gray-500 uppercase">Contract Start</p>
              <p className="text-xl font-bold text-gray-900 mt-1">
                {vendor.contract_start_date ? new Date(vendor.contract_start_date).toLocaleDateString('en-GB') : '—'}
              </p>
            </div>
            <div className="rounded-lg bg-gray-50 p-4 text-center">
              <p className="text-xs font-medium text-gray-500 uppercase">Contract End</p>
              <p className="text-xl font-bold text-gray-900 mt-1">
                {vendor.contract_end_date ? new Date(vendor.contract_end_date).toLocaleDateString('en-GB') : '—'}
              </p>
            </div>
          </div>
        </div>
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
