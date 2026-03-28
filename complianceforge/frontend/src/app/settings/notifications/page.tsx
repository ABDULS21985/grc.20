'use client';

import { useEffect, useState } from 'react';
import api from '@/lib/api';
import type { NotificationRule, NotificationChannel, NotificationPreference } from '@/types';

const EVENT_TYPES = [
  'incident.created', 'incident.severity_changed', 'breach.detected',
  'breach.deadline_approaching', 'breach.deadline_expired',
  'control.status_changed', 'policy.review_due', 'policy.review_overdue',
  'policy.published', 'policy.attestation_required',
  'finding.created', 'finding.overdue', 'finding.escalated',
  'risk.created', 'risk.threshold_exceeded', 'risk.review_due',
  'vendor.assessment_due', 'vendor.dpa_missing',
  'compliance.score_dropped', 'dsr.received', 'dsr.deadline_approaching',
];

export default function NotificationSettingsPage() {
  const [tab, setTab] = useState<'preferences' | 'rules' | 'channels'>('preferences');
  const [preferences, setPreferences] = useState<NotificationPreference[]>([]);
  const [rules, setRules] = useState<NotificationRule[]>([]);
  const [channels, setChannels] = useState<NotificationChannel[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadData();
  }, [tab]);

  async function loadData() {
    setLoading(true);
    try {
      if (tab === 'preferences') {
        const res = await api.getNotificationPreferences();
        setPreferences(res.data?.data || res.data || []);
      } else if (tab === 'rules') {
        const res = await api.getNotificationRules();
        setRules(res.data?.data || res.data || []);
      } else {
        const res = await api.getNotificationChannels();
        setChannels(res.data?.data || res.data || []);
      }
    } catch { /* ignore */ }
    setLoading(false);
  }

  async function testChannel(id: string) {
    try {
      await api.testNotificationChannel(id);
      alert('Test notification sent!');
    } catch (e: any) {
      alert('Test failed: ' + e.message);
    }
  }

  const tabs = [
    { key: 'preferences' as const, label: 'My Preferences' },
    { key: 'rules' as const, label: 'Notification Rules' },
    { key: 'channels' as const, label: 'Channels' },
  ];

  return (
    <div>
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Notification Settings</h1>
          <p className="text-gray-500 mt-1">Configure how you receive notifications</p>
        </div>
      </div>

      <div className="flex gap-1 border-b border-gray-200 mb-6">
        {tabs.map(t => (
          <button
            key={t.key}
            onClick={() => setTab(t.key)}
            className={`px-4 py-2.5 text-sm font-medium border-b-2 transition-colors ${tab === t.key ? 'border-indigo-600 text-indigo-600' : 'border-transparent text-gray-500 hover:text-gray-700'}`}
          >
            {t.label}
          </button>
        ))}
      </div>

      {loading ? (
        <div className="card animate-pulse h-64" />
      ) : tab === 'preferences' ? (
        <div className="card">
          <h2 className="font-semibold text-gray-900 mb-4">Notification Preferences by Event Type</h2>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="table-header">
                  <th className="px-4 py-3">Event Type</th>
                  <th className="px-4 py-3 text-center">Email</th>
                  <th className="px-4 py-3 text-center">In-App</th>
                  <th className="px-4 py-3 text-center">Slack</th>
                  <th className="px-4 py-3">Frequency</th>
                </tr>
              </thead>
              <tbody>
                {(preferences.length > 0 ? preferences : EVENT_TYPES.map(et => ({
                  id: et, event_type: et, email_enabled: true, in_app_enabled: true,
                  slack_enabled: false, digest_frequency: 'immediate' as const,
                  quiet_hours_start: null, quiet_hours_end: null, quiet_hours_timezone: null,
                }))).map(pref => (
                  <tr key={pref.id} className="border-t border-gray-100">
                    <td className="px-4 py-3 font-medium text-gray-900">
                      {pref.event_type.replace(/\./g, ' ').replace(/\b\w/g, l => l.toUpperCase())}
                    </td>
                    <td className="px-4 py-3 text-center">
                      <input type="checkbox" defaultChecked={pref.email_enabled} className="rounded border-gray-300 text-indigo-600" />
                    </td>
                    <td className="px-4 py-3 text-center">
                      <input type="checkbox" defaultChecked={pref.in_app_enabled} className="rounded border-gray-300 text-indigo-600" />
                    </td>
                    <td className="px-4 py-3 text-center">
                      <input type="checkbox" defaultChecked={pref.slack_enabled} className="rounded border-gray-300 text-indigo-600" />
                    </td>
                    <td className="px-4 py-3">
                      <select defaultValue={pref.digest_frequency} className="input text-xs py-1">
                        <option value="immediate">Immediate</option>
                        <option value="hourly">Hourly</option>
                        <option value="daily">Daily</option>
                        <option value="weekly">Weekly</option>
                      </select>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          <div className="mt-4 flex justify-end">
            <button className="btn-primary">Save Preferences</button>
          </div>
        </div>
      ) : tab === 'rules' ? (
        <div>
          <div className="flex justify-end mb-4">
            <button className="btn-primary">Create Rule</button>
          </div>
          <div className="card overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="table-header">
                  <th className="px-4 py-3">Name</th>
                  <th className="px-4 py-3">Event Type</th>
                  <th className="px-4 py-3">Severity Filter</th>
                  <th className="px-4 py-3">Recipient</th>
                  <th className="px-4 py-3">Active</th>
                  <th className="px-4 py-3">Actions</th>
                </tr>
              </thead>
              <tbody>
                {rules.map(rule => (
                  <tr key={rule.id} className="border-t border-gray-100">
                    <td className="px-4 py-3 font-medium text-gray-900">{rule.name}</td>
                    <td className="px-4 py-3 text-gray-600">{rule.event_type}</td>
                    <td className="px-4 py-3">
                      {rule.severity_filter ? rule.severity_filter.map(s => (
                        <span key={s} className={`badge badge-${s === 'critical' ? 'critical' : s === 'high' ? 'high' : 'medium'} mr-1`}>{s}</span>
                      )) : <span className="text-gray-400">All</span>}
                    </td>
                    <td className="px-4 py-3 text-gray-600">{rule.recipient_type}</td>
                    <td className="px-4 py-3">
                      <span className={`badge ${rule.is_active ? 'badge-low' : 'badge-medium'}`}>
                        {rule.is_active ? 'Active' : 'Inactive'}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <button className="text-indigo-600 hover:text-indigo-700 text-xs font-medium">Edit</button>
                    </td>
                  </tr>
                ))}
                {rules.length === 0 && (
                  <tr><td colSpan={6} className="px-4 py-8 text-center text-gray-500">No notification rules configured</td></tr>
                )}
              </tbody>
            </table>
          </div>
        </div>
      ) : (
        <div>
          <div className="flex justify-end mb-4">
            <button className="btn-primary">Add Channel</button>
          </div>
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
            {channels.map(ch => (
              <div key={ch.id} className="card">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <div className={`h-10 w-10 rounded-lg flex items-center justify-center ${ch.channel_type === 'email' ? 'bg-blue-100' : ch.channel_type === 'slack' ? 'bg-purple-100' : ch.channel_type === 'webhook' ? 'bg-orange-100' : 'bg-green-100'}`}>
                      <span className="text-xs font-bold uppercase">{ch.channel_type.slice(0, 2)}</span>
                    </div>
                    <div>
                      <h3 className="font-semibold text-gray-900">{ch.name}</h3>
                      <p className="text-xs text-gray-500">{ch.channel_type}</p>
                    </div>
                  </div>
                  <span className={`badge ${ch.is_active ? 'badge-low' : 'badge-medium'}`}>
                    {ch.is_active ? 'Active' : 'Inactive'}
                  </span>
                </div>
                <div className="mt-4 flex gap-2">
                  <button onClick={() => testChannel(ch.id)} className="btn-secondary text-xs py-1.5">Test</button>
                  <button className="btn-secondary text-xs py-1.5">Edit</button>
                </div>
              </div>
            ))}
            {channels.length === 0 && (
              <div className="col-span-full text-center py-12">
                <p className="text-gray-500">No notification channels configured</p>
                <button className="btn-primary mt-4">Set Up Your First Channel</button>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
