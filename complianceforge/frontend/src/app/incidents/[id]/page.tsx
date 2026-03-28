'use client';

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import api from '@/lib/api';
import type { Incident } from '@/types';

interface IncidentDetail extends Incident {
  root_cause?: string;
  containment_actions?: string;
  remediation_actions?: string;
  lessons_learned?: string;
  reporter_name?: string;
  assigned_to_name?: string;
  nis2_early_warning_submitted?: boolean;
  nis2_early_warning_at?: string;
}

export default function IncidentDetailPage() {
  const params = useParams();
  const id = params.id as string;

  const [incident, setIncident] = useState<IncidentDetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [actionLoading, setActionLoading] = useState(false);
  const [actionMessage, setActionMessage] = useState<string | null>(null);

  useEffect(() => {
    api.getIncident(id)
      .then((res) => setIncident(res.data))
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, [id]);

  async function handleNotifyDPA() {
    setActionLoading(true);
    setActionMessage(null);
    try {
      await api.notifyDPA(id);
      setActionMessage('DPA notification submitted successfully.');
      const res = await api.getIncident(id);
      setIncident(res.data);
    } catch (err: unknown) {
      setActionMessage(`Failed to notify DPA: ${err instanceof Error ? err.message : 'Unknown error'}`);
    } finally {
      setActionLoading(false);
    }
  }

  async function handleNIS2EarlyWarning() {
    setActionLoading(true);
    setActionMessage(null);
    try {
      await api.submitNIS2EarlyWarning(id);
      setActionMessage('NIS2 early warning submitted successfully.');
      const res = await api.getIncident(id);
      setIncident(res.data);
    } catch (err: unknown) {
      setActionMessage(`Failed to submit early warning: ${err instanceof Error ? err.message : 'Unknown error'}`);
    } finally {
      setActionLoading(false);
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <p className="text-gray-500">Loading incident...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="rounded-lg bg-red-50 border border-red-200 p-4 text-red-700">{error}</div>
    );
  }

  if (!incident) {
    return (
      <div className="rounded-lg bg-gray-50 border border-gray-200 p-4 text-gray-600">
        Incident not found.
      </div>
    );
  }

  // Calculate hours remaining to notification deadline
  let hoursRemaining: number | null = null;
  if (incident.notification_deadline) {
    const deadlineMs = new Date(incident.notification_deadline).getTime();
    const nowMs = Date.now();
    hoursRemaining = Math.max(0, (deadlineMs - nowMs) / (1000 * 60 * 60));
  }

  const severityColor: Record<string, string> = {
    critical: 'badge-critical',
    high: 'badge-high',
    medium: 'badge-medium',
    low: 'badge-low',
  };

  const statusColor: Record<string, string> = {
    open: 'badge-critical',
    investigating: 'badge-high',
    contained: 'badge-medium',
    resolved: 'badge-low',
    closed: 'badge-info',
  };

  return (
    <div>
      {/* Back button */}
      <a href="/incidents" className="inline-flex items-center text-sm text-gray-500 hover:text-indigo-600 mb-4">
        &larr; Back to Incidents
      </a>

      {/* Header */}
      <div className="flex items-start justify-between mb-6">
        <div>
          <div className="flex items-center gap-3 mb-1">
            <span className="text-sm font-mono text-gray-400">{incident.incident_ref}</span>
            <span className={`badge ${severityColor[incident.severity] || 'badge-info'}`}>{incident.severity}</span>
            <span className={`badge ${statusColor[incident.status] || 'badge-info'}`}>{incident.status}</span>
            {incident.is_data_breach && <span className="badge badge-critical">Data Breach</span>}
            {incident.is_nis2_reportable && <span className="badge badge-high">NIS2 Reportable</span>}
          </div>
          <h1 className="text-2xl font-bold text-gray-900">{incident.title}</h1>
          <p className="text-sm text-gray-500 mt-1">
            Reported {new Date(incident.reported_at).toLocaleString('en-GB', { dateStyle: 'medium', timeStyle: 'short' })}
            {incident.reporter_name && ` by ${incident.reporter_name}`}
          </p>
        </div>
      </div>

      {/* Action Message */}
      {actionMessage && (
        <div className={`mb-4 rounded-lg p-3 text-sm ${actionMessage.startsWith('Failed') ? 'bg-red-50 border border-red-200 text-red-700' : 'bg-green-50 border border-green-200 text-green-700'}`}>
          {actionMessage}
        </div>
      )}

      {/* GDPR Data Breach Panel */}
      {incident.is_data_breach && (
        <div className="mb-6 rounded-xl border-2 border-red-300 bg-red-50 p-6">
          <div className="flex items-center gap-2 mb-4">
            <span className="text-2xl">&#9888;</span>
            <h2 className="text-lg font-bold text-red-800">GDPR Data Breach - Article 33 Notification Required</h2>
          </div>

          <div className="grid grid-cols-2 gap-4 sm:grid-cols-4 mb-4">
            <div className="rounded-lg bg-white p-3 border border-red-200">
              <p className="text-xs font-medium text-red-600 uppercase">Notification Deadline</p>
              <p className="text-sm font-bold text-red-800 mt-1">
                {incident.notification_deadline
                  ? new Date(incident.notification_deadline).toLocaleString('en-GB', { dateStyle: 'short', timeStyle: 'short' })
                  : '—'}
              </p>
            </div>
            <div className="rounded-lg bg-white p-3 border border-red-200">
              <p className="text-xs font-medium text-red-600 uppercase">Hours Remaining</p>
              <p className={`text-sm font-bold mt-1 ${hoursRemaining !== null && hoursRemaining < 12 ? 'text-red-800' : 'text-orange-700'}`}>
                {hoursRemaining !== null ? `${hoursRemaining.toFixed(1)} hours` : '—'}
              </p>
            </div>
            <div className="rounded-lg bg-white p-3 border border-red-200">
              <p className="text-xs font-medium text-red-600 uppercase">Data Subjects Affected</p>
              <p className="text-sm font-bold text-red-800 mt-1">
                {incident.data_subjects_affected > 0 ? incident.data_subjects_affected.toLocaleString() : '—'}
              </p>
            </div>
            <div className="rounded-lg bg-white p-3 border border-red-200">
              <p className="text-xs font-medium text-red-600 uppercase">DPA Notification</p>
              <p className="text-sm font-bold mt-1">
                {incident.dpa_notified_at ? (
                  <span className="text-green-700">
                    Notified {new Date(incident.dpa_notified_at).toLocaleDateString('en-GB')}
                  </span>
                ) : (
                  <span className="text-red-800">Pending</span>
                )}
              </p>
            </div>
          </div>

          {!incident.dpa_notified_at && (
            <button
              onClick={handleNotifyDPA}
              disabled={actionLoading}
              className="btn-danger"
            >
              {actionLoading ? 'Submitting...' : 'Notify DPA Now'}
            </button>
          )}
        </div>
      )}

      {/* NIS2 Section */}
      {incident.is_nis2_reportable && (
        <div className="mb-6 rounded-xl border-2 border-orange-300 bg-orange-50 p-6">
          <div className="flex items-center gap-2 mb-3">
            <span className="text-xl">&#9888;</span>
            <h2 className="text-lg font-bold text-orange-800">NIS2 Directive - Article 23 Reporting</h2>
          </div>
          <p className="text-sm text-orange-700 mb-4">
            This incident is classified as NIS2 reportable. An early warning must be submitted to the competent authority
            within 24 hours of becoming aware of the significant incident.
          </p>

          {incident.nis2_early_warning_submitted || incident.nis2_early_warning_at ? (
            <div className="rounded-lg bg-white p-3 border border-orange-200 inline-block">
              <p className="text-sm text-green-700 font-medium">
                Early warning submitted
                {incident.nis2_early_warning_at && ` on ${new Date(incident.nis2_early_warning_at).toLocaleString('en-GB', { dateStyle: 'short', timeStyle: 'short' })}`}
              </p>
            </div>
          ) : (
            <button
              onClick={handleNIS2EarlyWarning}
              disabled={actionLoading}
              className="rounded-lg bg-orange-600 px-4 py-2 text-sm font-medium text-white hover:bg-orange-700 transition-colors"
            >
              {actionLoading ? 'Submitting...' : 'Submit Early Warning'}
            </button>
          )}
        </div>
      )}

      {/* Incident Details */}
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        <div className="card">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Incident Details</h2>
          <div className="space-y-3">
            <DetailRow label="Type" value={incident.incident_type.replace(/_/g, ' ')} />
            <DetailRow label="Severity" value={incident.severity} />
            <DetailRow label="Status" value={incident.status} />
            <DetailRow
              label="Reported At"
              value={new Date(incident.reported_at).toLocaleString('en-GB', { dateStyle: 'medium', timeStyle: 'short' })}
            />
            {incident.assigned_to_name && <DetailRow label="Assigned To" value={incident.assigned_to_name} />}
            <DetailRow
              label="Financial Impact"
              value={incident.financial_impact_eur > 0 ? `€${incident.financial_impact_eur.toLocaleString()}` : '—'}
            />
          </div>
        </div>

        <div className="card">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Description &amp; Timeline</h2>
          {incident.description ? (
            <div className="space-y-4">
              <div>
                <p className="text-xs font-medium text-gray-500 uppercase mb-1">Description</p>
                <p className="text-sm text-gray-700 whitespace-pre-wrap">{incident.description}</p>
              </div>
              {incident.root_cause && (
                <div>
                  <p className="text-xs font-medium text-gray-500 uppercase mb-1">Root Cause</p>
                  <p className="text-sm text-gray-700">{incident.root_cause}</p>
                </div>
              )}
              {incident.containment_actions && (
                <div>
                  <p className="text-xs font-medium text-gray-500 uppercase mb-1">Containment Actions</p>
                  <p className="text-sm text-gray-700">{incident.containment_actions}</p>
                </div>
              )}
              {incident.remediation_actions && (
                <div>
                  <p className="text-xs font-medium text-gray-500 uppercase mb-1">Remediation Actions</p>
                  <p className="text-sm text-gray-700">{incident.remediation_actions}</p>
                </div>
              )}
              {incident.lessons_learned && (
                <div>
                  <p className="text-xs font-medium text-gray-500 uppercase mb-1">Lessons Learned</p>
                  <p className="text-sm text-gray-700">{incident.lessons_learned}</p>
                </div>
              )}
            </div>
          ) : (
            <p className="text-sm text-gray-500">No description available.</p>
          )}
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
