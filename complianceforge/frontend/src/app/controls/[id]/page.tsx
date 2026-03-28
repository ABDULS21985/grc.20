'use client';

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import api from '@/lib/api';
import type { FrameworkControl, ComplianceScore } from '@/types';

interface ControlImplementation {
  control_id: string;
  control_code: string;
  control_title: string;
  framework_id: string;
  framework_code: string;
  framework_name: string;
  status: string;
  maturity_level: number;
  implementation_description: string;
  gap_description: string;
  evidence_list: EvidenceItem[];
  test_history: TestRecord[];
}

interface EvidenceItem {
  id: string;
  name: string;
  file_type: string;
  uploaded_at: string;
  uploaded_by: string;
}

interface TestRecord {
  id: string;
  test_date: string;
  result: string;
  tester_name: string;
  notes: string;
}

const MATURITY_LABELS: Record<number, string> = {
  0: 'Non-existent',
  1: 'Initial / Ad-hoc',
  2: 'Managed',
  3: 'Defined',
  4: 'Quantitatively Managed',
  5: 'Optimizing',
};

const STATUS_OPTIONS = [
  'not_implemented',
  'partially_implemented',
  'implemented',
  'not_applicable',
];

export default function ControlDetailPage() {
  const params = useParams();
  const id = params.id as string;

  const [control, setControl] = useState<ControlImplementation | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    // Attempt to load control data from compliance scores or search
    // Since there is no direct getControl(id) endpoint, we search for it
    setLoading(true);
    setError(null);

    api.searchControls('')
      .then(async (searchRes) => {
        const controls: FrameworkControl[] = searchRes.data?.data || searchRes.data || [];
        const found = controls.find((c: FrameworkControl) => c.id === id);

        // Build a control implementation object from available data
        let scoresData: ComplianceScore[] = [];
        try {
          const scoresRes = await api.getComplianceScores();
          scoresData = scoresRes.data?.data || scoresRes.data || [];
        } catch {
          // Scores not critical
        }

        const frameworkScore = found
          ? scoresData.find((s: ComplianceScore) => s.framework_id === found.framework_id)
          : null;

        setControl({
          control_id: id,
          control_code: found?.code || id,
          control_title: found?.title || 'Control',
          framework_id: found?.framework_id || '',
          framework_code: frameworkScore?.framework_code || '',
          framework_name: frameworkScore?.framework_name || '',
          status: 'not_implemented',
          maturity_level: 0,
          implementation_description: found?.description || '',
          gap_description: found?.guidance || '',
          evidence_list: [],
          test_history: [],
        });
      })
      .catch((err) => {
        // Fallback: show a basic placeholder with the ID
        setControl({
          control_id: id,
          control_code: id,
          control_title: 'Control Detail',
          framework_id: '',
          framework_code: '',
          framework_name: '',
          status: 'not_implemented',
          maturity_level: 0,
          implementation_description: '',
          gap_description: '',
          evidence_list: [],
          test_history: [],
        });
        setError(null); // Show the page even if search failed
      })
      .finally(() => setLoading(false));
  }, [id]);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <p className="text-gray-500">Loading control...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="rounded-lg bg-red-50 border border-red-200 p-4 text-red-700">{error}</div>
    );
  }

  if (!control) {
    return (
      <div className="rounded-lg bg-gray-50 border border-gray-200 p-4 text-gray-600">
        Control not found.
      </div>
    );
  }

  const statusColor: Record<string, string> = {
    not_implemented: 'badge-critical',
    partially_implemented: 'badge-medium',
    implemented: 'badge-low',
    not_applicable: 'badge-info',
  };

  return (
    <div>
      {/* Back button */}
      {control.framework_id ? (
        <a href={`/frameworks/${control.framework_id}`} className="inline-flex items-center text-sm text-gray-500 hover:text-indigo-600 mb-4">
          &larr; Back to Framework
        </a>
      ) : (
        <a href="/frameworks" className="inline-flex items-center text-sm text-gray-500 hover:text-indigo-600 mb-4">
          &larr; Back to Frameworks
        </a>
      )}

      {/* Header */}
      <div className="flex items-start justify-between mb-6">
        <div>
          <div className="flex items-center gap-3 mb-1">
            <span className="text-sm font-mono text-gray-400">{control.control_code}</span>
            {control.framework_name && (
              <span className="badge badge-info">{control.framework_name}</span>
            )}
            <span className={`badge ${statusColor[control.status] || 'badge-info'}`}>
              {control.status.replace(/_/g, ' ')}
            </span>
          </div>
          <h1 className="text-2xl font-bold text-gray-900">{control.control_title}</h1>
        </div>
        <div className="text-right">
          <p className="text-sm text-gray-500">Maturity Level</p>
          <p className="text-3xl font-bold text-indigo-600">{control.maturity_level}/5</p>
          <p className="text-xs text-gray-400">{MATURITY_LABELS[control.maturity_level] || ''}</p>
        </div>
      </div>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        {/* Left Column */}
        <div className="space-y-6">
          {/* Status */}
          <div className="card">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">Implementation Status</h2>
            <div className="space-y-4">
              <div>
                <p className="text-xs font-medium text-gray-500 uppercase mb-2">Current Status</p>
                <div className="flex flex-wrap gap-2">
                  {STATUS_OPTIONS.map((s) => (
                    <span
                      key={s}
                      className={`px-3 py-1.5 rounded-lg text-xs font-medium border transition-colors ${
                        control.status === s
                          ? 'bg-indigo-50 border-indigo-300 text-indigo-700'
                          : 'bg-gray-50 border-gray-200 text-gray-500'
                      }`}
                    >
                      {s.replace(/_/g, ' ')}
                    </span>
                  ))}
                </div>
              </div>

              <div>
                <p className="text-xs font-medium text-gray-500 uppercase mb-2">Maturity Level</p>
                <div className="flex items-center gap-1">
                  {[1, 2, 3, 4, 5].map((level) => (
                    <div
                      key={level}
                      className={`h-8 flex-1 rounded flex items-center justify-center text-xs font-medium ${
                        level <= control.maturity_level
                          ? 'bg-indigo-500 text-white'
                          : 'bg-gray-100 text-gray-400'
                      }`}
                    >
                      {level}
                    </div>
                  ))}
                </div>
                <p className="text-xs text-gray-400 mt-1">{MATURITY_LABELS[control.maturity_level] || 'Not assessed'}</p>
              </div>
            </div>
          </div>

          {/* Implementation Description */}
          <div className="card">
            <h2 className="text-lg font-semibold text-gray-900 mb-3">Implementation Description</h2>
            {control.implementation_description ? (
              <p className="text-sm text-gray-700 whitespace-pre-wrap">{control.implementation_description}</p>
            ) : (
              <p className="text-sm text-gray-400">No implementation description provided.</p>
            )}
          </div>

          {/* Gap Description */}
          <div className="card">
            <h2 className="text-lg font-semibold text-gray-900 mb-3">Gap Description / Guidance</h2>
            {control.gap_description ? (
              <p className="text-sm text-gray-700 whitespace-pre-wrap">{control.gap_description}</p>
            ) : (
              <p className="text-sm text-gray-400">No gap description available.</p>
            )}
          </div>
        </div>

        {/* Right Column */}
        <div className="space-y-6">
          {/* Evidence */}
          <div className="card">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-lg font-semibold text-gray-900">Evidence ({control.evidence_list.length})</h2>
              <button className="btn-primary text-sm">Upload Evidence</button>
            </div>
            {control.evidence_list.length > 0 ? (
              <div className="space-y-2">
                {control.evidence_list.map((ev) => (
                  <div key={ev.id} className="flex items-center justify-between rounded-lg border border-gray-100 p-3">
                    <div className="flex items-center gap-3">
                      <div className="h-8 w-8 rounded-lg bg-indigo-50 flex items-center justify-center">
                        <span className="text-xs font-bold text-indigo-600">{ev.file_type?.toUpperCase().slice(0, 3) || 'DOC'}</span>
                      </div>
                      <div>
                        <p className="text-sm font-medium text-gray-900">{ev.name}</p>
                        <p className="text-xs text-gray-400">
                          Uploaded by {ev.uploaded_by} on {new Date(ev.uploaded_at).toLocaleDateString('en-GB')}
                        </p>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="rounded-lg bg-gray-50 border border-gray-200 p-6 text-center">
                <p className="text-sm text-gray-500">No evidence uploaded yet.</p>
                <p className="text-xs text-gray-400 mt-1">Upload documents, screenshots, or configuration exports.</p>
              </div>
            )}
          </div>

          {/* Test History */}
          <div className="card">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-lg font-semibold text-gray-900">Test History ({control.test_history.length})</h2>
              <button className="btn-secondary text-sm">Record Test</button>
            </div>
            {control.test_history.length > 0 ? (
              <div className="space-y-3">
                {control.test_history.map((test) => (
                  <div key={test.id} className="rounded-lg border border-gray-100 p-3">
                    <div className="flex items-center justify-between mb-1">
                      <span className="text-sm font-medium text-gray-900">
                        {new Date(test.test_date).toLocaleDateString('en-GB')}
                      </span>
                      <span className={`badge ${test.result === 'pass' ? 'badge-low' : test.result === 'fail' ? 'badge-critical' : 'badge-medium'}`}>
                        {test.result}
                      </span>
                    </div>
                    <p className="text-xs text-gray-500">Tested by {test.tester_name}</p>
                    {test.notes && <p className="text-xs text-gray-400 mt-1">{test.notes}</p>}
                  </div>
                ))}
              </div>
            ) : (
              <div className="rounded-lg bg-gray-50 border border-gray-200 p-6 text-center">
                <p className="text-sm text-gray-500">No test records found.</p>
                <p className="text-xs text-gray-400 mt-1">Record control testing results to track effectiveness over time.</p>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
