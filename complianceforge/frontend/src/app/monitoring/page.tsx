'use client';

import { useEffect, useState } from 'react';
import api from '@/lib/api';
import type { MonitoringDashboard, EvidenceCollectionConfig, ComplianceMonitor, DriftEvent } from '@/types';

const DRIFT_SEVERITY_COLORS: Record<string, string> = {
  critical: 'badge-critical',
  high: 'badge-high',
  medium: 'badge-medium',
  low: 'badge-low',
};

const MONITOR_STATUS_COLORS: Record<string, string> = {
  passing: 'bg-green-500',
  failing: 'bg-red-500',
  unknown: 'bg-gray-400',
};

const HEALTH_COLORS: Record<string, { bg: string; text: string; label: string }> = {
  healthy: { bg: 'bg-green-100', text: 'text-green-700', label: 'Healthy' },
  degraded: { bg: 'bg-amber-100', text: 'text-amber-700', label: 'Degraded' },
  critical: { bg: 'bg-red-100', text: 'text-red-700', label: 'Critical' },
};

export default function MonitoringPage() {
  const [tab, setTab] = useState<'dashboard' | 'configs' | 'monitors' | 'drift'>('dashboard');
  const [dashboard, setDashboard] = useState<MonitoringDashboard | null>(null);
  const [configs, setConfigs] = useState<EvidenceCollectionConfig[]>([]);
  const [monitors, setMonitors] = useState<ComplianceMonitor[]>([]);
  const [driftEvents, setDriftEvents] = useState<DriftEvent[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadData();
  }, [tab]);

  async function loadData() {
    setLoading(true);
    try {
      if (tab === 'dashboard') {
        const res = await api.getMonitoringDashboard();
        setDashboard(res.data);
      } else if (tab === 'configs') {
        const res = await api.getMonitoringConfigs();
        setConfigs(res.data?.data || res.data || []);
      } else if (tab === 'monitors') {
        const res = await api.getComplianceMonitors();
        setMonitors(res.data?.data || res.data || []);
      } else {
        const res = await api.getDriftEvents();
        setDriftEvents(res.data?.data || res.data || []);
      }
    } catch { /* ignore */ }
    setLoading(false);
  }

  async function runNow(configId: string) {
    try {
      await api.runMonitoringConfigNow(configId);
      alert('Collection triggered!');
    } catch (e: any) {
      alert('Failed: ' + e.message);
    }
  }

  async function acknowledgeDriftEvent(id: string) {
    try {
      await api.acknowledgeDrift(id);
      setDriftEvents(prev => prev.map(d => d.id === id ? { ...d, acknowledged_at: new Date().toISOString() } : d));
    } catch { /* ignore */ }
  }

  const tabs = [
    { key: 'dashboard' as const, label: 'Dashboard' },
    { key: 'configs' as const, label: 'Evidence Collection' },
    { key: 'monitors' as const, label: 'Monitors' },
    { key: 'drift' as const, label: 'Drift Events' },
  ];

  return (
    <div>
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Continuous Monitoring</h1>
          <p className="text-gray-500 mt-1">Automated evidence collection, compliance monitoring, and drift detection</p>
        </div>
      </div>

      <div className="flex gap-1 border-b border-gray-200 mb-6">
        {tabs.map(t => (
          <button key={t.key} onClick={() => setTab(t.key)}
            className={`px-4 py-2.5 text-sm font-medium border-b-2 transition-colors ${tab === t.key ? 'border-indigo-600 text-indigo-600' : 'border-transparent text-gray-500 hover:text-gray-700'}`}
          >
            {t.label}
            {t.key === 'drift' && driftEvents.filter(d => !d.acknowledged_at).length > 0 && (
              <span className="ml-1.5 inline-flex h-5 w-5 items-center justify-center rounded-full bg-red-500 text-[10px] text-white font-bold">
                {driftEvents.filter(d => !d.acknowledged_at).length}
              </span>
            )}
          </button>
        ))}
      </div>

      {loading ? <div className="card animate-pulse h-96" /> : tab === 'dashboard' && dashboard ? (
        <div>
          {/* Health Status */}
          <div className={`rounded-xl p-6 mb-6 ${HEALTH_COLORS[dashboard.health_status]?.bg || 'bg-gray-100'}`}>
            <div className="flex items-center gap-3">
              <div className={`h-4 w-4 rounded-full ${dashboard.health_status === 'healthy' ? 'bg-green-500' : dashboard.health_status === 'degraded' ? 'bg-amber-500' : 'bg-red-500'}`} />
              <h2 className={`text-xl font-bold ${HEALTH_COLORS[dashboard.health_status]?.text || 'text-gray-700'}`}>
                {HEALTH_COLORS[dashboard.health_status]?.label || 'Unknown'}
              </h2>
            </div>
            <p className={`text-sm mt-1 ${HEALTH_COLORS[dashboard.health_status]?.text || 'text-gray-600'}`}>
              {dashboard.monitors_passing} of {dashboard.monitors_total} monitors passing
            </p>
          </div>

          <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-4 mb-6">
            <div className="card">
              <p className="text-sm text-gray-500">Active Drift Events</p>
              <p className="text-3xl font-bold text-gray-900 mt-1">{dashboard.active_drift_events}</p>
              <div className="flex gap-2 mt-2">
                {Object.entries(dashboard.drift_by_severity || {}).map(([sev, count]) => (
                  <span key={sev} className={`badge ${DRIFT_SEVERITY_COLORS[sev] || 'badge-info'}`}>{sev}: {count}</span>
                ))}
              </div>
            </div>
            <div className="card">
              <p className="text-sm text-gray-500">Evidence Success (24h)</p>
              <p className="text-3xl font-bold text-green-600 mt-1">{dashboard.evidence_success_rate_24h}%</p>
            </div>
            <div className="card">
              <p className="text-sm text-gray-500">Evidence Success (7d)</p>
              <p className="text-3xl font-bold text-gray-900 mt-1">{dashboard.evidence_success_rate_7d}%</p>
            </div>
            <div className="card">
              <p className="text-sm text-gray-500">Monitors Failing</p>
              <p className="text-3xl font-bold text-red-600 mt-1">{dashboard.monitors_failing}</p>
            </div>
          </div>
        </div>
      ) : tab === 'configs' ? (
        <div>
          <div className="flex justify-end mb-4">
            <button className="btn-primary">Add Collection Config</button>
          </div>
          <div className="card overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="table-header">
                  <th className="px-4 py-3">Name</th>
                  <th className="px-4 py-3">Method</th>
                  <th className="px-4 py-3">Schedule</th>
                  <th className="px-4 py-3">Last Collection</th>
                  <th className="px-4 py-3">Status</th>
                  <th className="px-4 py-3">Failures</th>
                  <th className="px-4 py-3">Actions</th>
                </tr>
              </thead>
              <tbody>
                {configs.map(cfg => (
                  <tr key={cfg.id} className="border-t border-gray-100">
                    <td className="px-4 py-3 font-medium text-gray-900">{cfg.name}</td>
                    <td className="px-4 py-3"><span className="badge badge-info">{cfg.collection_method.replace(/_/g, ' ')}</span></td>
                    <td className="px-4 py-3 text-gray-600">{cfg.schedule_description || cfg.schedule_cron}</td>
                    <td className="px-4 py-3 text-gray-600">{cfg.last_collection_at ? new Date(cfg.last_collection_at).toLocaleDateString('en-GB') : 'Never'}</td>
                    <td className="px-4 py-3">
                      <span className={`badge ${cfg.last_collection_status === 'success' ? 'badge-low' : cfg.last_collection_status === 'failed' ? 'badge-critical' : 'badge-info'}`}>
                        {cfg.last_collection_status || 'pending'}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <span className={cfg.consecutive_failures > 0 ? 'text-red-600 font-medium' : 'text-gray-400'}>{cfg.consecutive_failures}</span>
                    </td>
                    <td className="px-4 py-3">
                      <button onClick={() => runNow(cfg.id)} className="text-indigo-600 hover:text-indigo-700 text-xs font-medium">Run Now</button>
                    </td>
                  </tr>
                ))}
                {configs.length === 0 && (
                  <tr><td colSpan={7} className="px-4 py-12 text-center text-gray-500">No evidence collection configs</td></tr>
                )}
              </tbody>
            </table>
          </div>
        </div>
      ) : tab === 'monitors' ? (
        <div>
          <div className="flex justify-end mb-4">
            <button className="btn-primary">Create Monitor</button>
          </div>
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
            {monitors.map(m => (
              <div key={m.id} className="card">
                <div className="flex items-center justify-between">
                  <h3 className="font-semibold text-gray-900">{m.name}</h3>
                  <div className={`h-3 w-3 rounded-full ${MONITOR_STATUS_COLORS[m.last_check_status] || 'bg-gray-400'}`} />
                </div>
                <p className="text-xs text-gray-500 mt-1">{m.monitor_type.replace(/_/g, ' ')}</p>
                <div className="mt-3 space-y-1.5 text-sm">
                  <div className="flex justify-between">
                    <span className="text-gray-500">Status</span>
                    <span className={`font-medium ${m.last_check_status === 'passing' ? 'text-green-600' : m.last_check_status === 'failing' ? 'text-red-600' : 'text-gray-400'}`}>
                      {m.last_check_status}
                    </span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-gray-500">Consecutive Failures</span>
                    <span className={m.consecutive_failures > 0 ? 'text-red-600 font-medium' : 'text-gray-900'}>{m.consecutive_failures}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-gray-500">Last Check</span>
                    <span className="text-gray-600">{m.last_check_at ? new Date(m.last_check_at).toLocaleDateString('en-GB') : 'Never'}</span>
                  </div>
                </div>
              </div>
            ))}
            {monitors.length === 0 && (
              <div className="col-span-full text-center py-12 text-gray-500">No compliance monitors configured</div>
            )}
          </div>
        </div>
      ) : tab === 'drift' ? (
        <div className="card overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="table-header">
                <th className="px-4 py-3">Type</th>
                <th className="px-4 py-3">Severity</th>
                <th className="px-4 py-3">Entity</th>
                <th className="px-4 py-3">Description</th>
                <th className="px-4 py-3">Detected</th>
                <th className="px-4 py-3">Status</th>
                <th className="px-4 py-3">Actions</th>
              </tr>
            </thead>
            <tbody>
              {driftEvents.map(d => (
                <tr key={d.id} className={`border-t border-gray-100 ${!d.acknowledged_at ? 'bg-red-50/30' : ''}`}>
                  <td className="px-4 py-3 font-medium text-gray-900">{d.drift_type.replace(/_/g, ' ')}</td>
                  <td className="px-4 py-3"><span className={`badge ${DRIFT_SEVERITY_COLORS[d.severity]}`}>{d.severity}</span></td>
                  <td className="px-4 py-3 text-gray-600">{d.entity_ref || d.entity_type}</td>
                  <td className="px-4 py-3 text-gray-600 max-w-xs truncate">{d.description}</td>
                  <td className="px-4 py-3 text-gray-600">{new Date(d.detected_at).toLocaleDateString('en-GB')}</td>
                  <td className="px-4 py-3">
                    {d.resolved_at ? <span className="badge badge-low">Resolved</span>
                      : d.acknowledged_at ? <span className="badge badge-medium">Acknowledged</span>
                      : <span className="badge badge-critical">Active</span>}
                  </td>
                  <td className="px-4 py-3">
                    {!d.acknowledged_at && (
                      <button onClick={() => acknowledgeDriftEvent(d.id)} className="text-indigo-600 hover:text-indigo-700 text-xs font-medium">Acknowledge</button>
                    )}
                  </td>
                </tr>
              ))}
              {driftEvents.length === 0 && (
                <tr><td colSpan={7} className="px-4 py-12 text-center text-gray-500">No drift events detected</td></tr>
              )}
            </tbody>
          </table>
        </div>
      ) : null}
    </div>
  );
}
