'use client';

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import api from '@/lib/api';
import type { Audit, AuditFinding } from '@/types';

interface AuditDetail extends Audit {
  description?: string;
  scope?: string;
  lead_auditor_name?: string;
  methodology?: string;
}

export default function AuditDetailPage() {
  const params = useParams();
  const id = params.id as string;

  const [audit, setAudit] = useState<AuditDetail | null>(null);
  const [findings, setFindings] = useState<AuditFinding[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setLoading(true);
    setError(null);
    Promise.all([
      api.getAudit(id),
      api.getAuditFindings(id),
    ])
      .then(([auditRes, findingsRes]) => {
        setAudit(auditRes.data);
        setFindings(findingsRes.data?.data || findingsRes.data || []);
      })
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, [id]);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <p className="text-gray-500">Loading audit...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="rounded-lg bg-red-50 border border-red-200 p-4 text-red-700">{error}</div>
    );
  }

  if (!audit) {
    return (
      <div className="rounded-lg bg-gray-50 border border-gray-200 p-4 text-gray-600">
        Audit not found.
      </div>
    );
  }

  const statusColor: Record<string, string> = {
    planned: 'badge-info',
    in_progress: 'badge-medium',
    completed: 'badge-low',
    cancelled: 'badge-critical',
  };

  const severityColor: Record<string, string> = {
    critical: 'badge-critical',
    high: 'badge-high',
    medium: 'badge-medium',
    low: 'badge-low',
    informational: 'badge-info',
  };

  return (
    <div>
      {/* Back button */}
      <a href="/audits" className="inline-flex items-center text-sm text-gray-500 hover:text-indigo-600 mb-4">
        &larr; Back to Audits
      </a>

      {/* Header */}
      <div className="flex items-start justify-between mb-6">
        <div>
          <div className="flex items-center gap-3 mb-1">
            <span className="text-sm font-mono text-gray-400">{audit.audit_ref}</span>
            <span className="badge badge-info">{audit.audit_type}</span>
            <span className={`badge ${statusColor[audit.status] || 'badge-info'}`}>
              {audit.status.replace(/_/g, ' ')}
            </span>
          </div>
          <h1 className="text-2xl font-bold text-gray-900">{audit.title}</h1>
        </div>
        <button className="btn-primary">+ Add Finding</button>
      </div>

      {/* Audit Info */}
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2 mb-6">
        <div className="card">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Audit Details</h2>
          <div className="space-y-3">
            <DetailRow label="Audit Type" value={audit.audit_type} />
            <DetailRow label="Status" value={audit.status.replace(/_/g, ' ')} />
            <DetailRow
              label="Planned Start"
              value={audit.planned_start_date ? new Date(audit.planned_start_date).toLocaleDateString('en-GB') : '—'}
            />
            <DetailRow
              label="Planned End"
              value={audit.planned_end_date ? new Date(audit.planned_end_date).toLocaleDateString('en-GB') : '—'}
            />
            <DetailRow label="Lead Auditor" value={audit.lead_auditor_name || audit.lead_auditor_id || '—'} />
            {audit.scope && <DetailRow label="Scope" value={audit.scope} />}
            {audit.methodology && <DetailRow label="Methodology" value={audit.methodology} />}
          </div>
        </div>

        <div className="card">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Findings Summary</h2>
          <div className="grid grid-cols-2 gap-3">
            <div className="rounded-lg bg-gray-50 p-3 text-center">
              <p className="text-2xl font-bold text-indigo-600">{audit.total_findings}</p>
              <p className="text-xs text-gray-500 mt-1">Total Findings</p>
            </div>
            <div className="rounded-lg bg-red-50 p-3 text-center">
              <p className="text-2xl font-bold text-red-600">{audit.critical_findings}</p>
              <p className="text-xs text-gray-500 mt-1">Critical</p>
            </div>
            <div className="rounded-lg bg-orange-50 p-3 text-center">
              <p className="text-2xl font-bold text-orange-600">{audit.high_findings}</p>
              <p className="text-xs text-gray-500 mt-1">High</p>
            </div>
            <div className="rounded-lg bg-green-50 p-3 text-center">
              <p className="text-2xl font-bold text-green-600">
                {audit.total_findings - audit.critical_findings - audit.high_findings}
              </p>
              <p className="text-xs text-gray-500 mt-1">Medium / Low</p>
            </div>
          </div>
          {audit.description && (
            <div className="mt-4 pt-3 border-t border-gray-100">
              <p className="text-xs font-medium text-gray-500 uppercase mb-1">Description</p>
              <p className="text-sm text-gray-700">{audit.description}</p>
            </div>
          )}
        </div>
      </div>

      {/* Findings Table */}
      <div className="mb-4">
        <h2 className="text-lg font-semibold text-gray-900">Findings ({findings.length})</h2>
      </div>

      {findings.length === 0 ? (
        <div className="rounded-lg bg-gray-50 border border-gray-200 p-8 text-center">
          <p className="text-gray-500">No findings recorded for this audit.</p>
          <p className="text-sm text-gray-400 mt-1">Add findings as you conduct the audit.</p>
        </div>
      ) : (
        <div className="card overflow-hidden p-0">
          <table className="w-full">
            <thead>
              <tr className="border-b border-gray-100">
                <th className="table-header px-4 py-3">Ref</th>
                <th className="table-header px-4 py-3">Title</th>
                <th className="table-header px-4 py-3">Severity</th>
                <th className="table-header px-4 py-3">Status</th>
                <th className="table-header px-4 py-3">Type</th>
                <th className="table-header px-4 py-3">Due Date</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-50">
              {findings.map((finding) => (
                <tr key={finding.id} className="hover:bg-gray-50 transition-colors">
                  <td className="px-4 py-3 text-sm font-mono text-gray-500">{finding.finding_ref}</td>
                  <td className="px-4 py-3">
                    <p className="text-sm font-medium text-gray-900">{finding.title}</p>
                    {finding.description && (
                      <p className="text-xs text-gray-400 mt-0.5 truncate max-w-xs">{finding.description}</p>
                    )}
                  </td>
                  <td className="px-4 py-3">
                    <span className={`badge ${severityColor[finding.severity] || 'badge-info'}`}>
                      {finding.severity}
                    </span>
                  </td>
                  <td className="px-4 py-3">
                    <span className="badge badge-info">{finding.status}</span>
                  </td>
                  <td className="px-4 py-3">
                    <span className="text-sm text-gray-600">{finding.finding_type}</span>
                  </td>
                  <td className="px-4 py-3 text-sm text-gray-500">
                    {finding.due_date ? new Date(finding.due_date).toLocaleDateString('en-GB') : '—'}
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

function DetailRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between py-2 border-b border-gray-50">
      <span className="text-sm text-gray-500">{label}</span>
      <span className="text-sm font-medium text-gray-900 capitalize">{value}</span>
    </div>
  );
}
