'use client';

import { useEffect, useState } from 'react';
import api from '@/lib/api';

// ============================================================
// TYPES
// ============================================================

interface TestSuite {
  id: string;
  name: string;
  description: string;
  test_type: string;
  schedule_cron: string;
  is_active: boolean;
  last_run_at: string | null;
  last_run_status: string | null;
  pass_threshold_percent: number;
  test_case_count: number;
  created_at: string;
}

interface TestRun {
  id: string;
  test_suite_id: string;
  status: string;
  started_at: string | null;
  completed_at: string | null;
  total_tests: number;
  passed: number;
  failed: number;
  skipped: number;
  errors: number;
  pass_rate: number;
  threshold_met: boolean;
  results: TestCaseResult[];
  triggered_by: string;
  suite_name: string;
  created_at: string;
}

interface TestCaseResult {
  test_case_id: string;
  test_case_name: string;
  status: string;
  message: string;
  is_critical: boolean;
  duration: string;
}

interface PreAuditReport {
  id: string;
  framework_code: string;
  generated_at: string;
  overall_readiness: number;
  readiness_level: string;
  total_controls: number;
  controls_with_evidence: number;
  controls_missing_evidence: number;
  evidence_completion: number;
  validation_pass_rate: number;
  critical_gaps: PreAuditGap[];
  control_readiness: ControlReady[];
  recommendations: string[];
  estimated_remediation_hours: number;
}

interface PreAuditGap {
  control_code: string;
  control_name: string;
  gap_type: string;
  severity: string;
  description: string;
  recommendation: string;
}

interface ControlReady {
  control_code: string;
  control_title: string;
  evidence_count: number;
  required_count: number;
  validation_passed: number;
  readiness_percent: number;
  status: string;
}

interface Framework {
  id: string;
  code: string;
  name: string;
}

type ViewType = 'suites' | 'runs' | 'pre-audit';

// ============================================================
// HELPERS
// ============================================================

const statusIcons: Record<string, { color: string; label: string }> = {
  pass: { color: 'text-green-600', label: 'PASS' },
  fail: { color: 'text-red-600', label: 'FAIL' },
  skip: { color: 'text-yellow-600', label: 'SKIP' },
  error: { color: 'text-red-800', label: 'ERROR' },
};

const readinessColors: Record<string, string> = {
  audit_ready: 'bg-green-100 text-green-800',
  mostly_ready: 'bg-blue-100 text-blue-800',
  significant_gaps: 'bg-yellow-100 text-yellow-800',
  not_ready: 'bg-red-100 text-red-800',
};

const readinessLabels: Record<string, string> = {
  audit_ready: 'Audit Ready',
  mostly_ready: 'Mostly Ready',
  significant_gaps: 'Significant Gaps',
  not_ready: 'Not Ready',
};

function Badge({ text, className = '' }: { text: string; className?: string }) {
  return (
    <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${className}`}>
      {text}
    </span>
  );
}

// ============================================================
// MAIN PAGE COMPONENT
// ============================================================

export default function EvidenceTestingPage() {
  const [view, setView] = useState<ViewType>('suites');
  const [suites, setSuites] = useState<TestSuite[]>([]);
  const [selectedSuiteRuns, setSelectedSuiteRuns] = useState<TestRun[]>([]);
  const [selectedSuiteId, setSelectedSuiteId] = useState<string | null>(null);
  const [selectedRun, setSelectedRun] = useState<TestRun | null>(null);
  const [preAuditReport, setPreAuditReport] = useState<PreAuditReport | null>(null);
  const [frameworks, setFrameworks] = useState<Framework[]>([]);
  const [selectedFramework, setSelectedFramework] = useState('');
  const [loading, setLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadSuites();
    loadFrameworks();
  }, []);

  async function loadSuites() {
    setLoading(true);
    try {
      const res = await api.getEvidenceTestSuites();
      setSuites(res.data || []);
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }

  async function loadFrameworks() {
    try {
      const res = await api.getFrameworks();
      setFrameworks(res.data || []);
    } catch {
      // Non-critical
    }
  }

  async function handleRunSuite(suiteId: string) {
    setActionLoading(true);
    try {
      const res = await api.runEvidenceTestSuite(suiteId);
      const run = res.data as TestRun;
      setSelectedRun(run);
      setView('runs');
      loadSuites(); // Refresh suite statuses
    } catch (err: any) {
      setError(err.message);
    } finally {
      setActionLoading(false);
    }
  }

  async function handleViewRuns(suiteId: string) {
    setLoading(true);
    setSelectedSuiteId(suiteId);
    try {
      const res = await api.getEvidenceTestRunResults(suiteId);
      setSelectedSuiteRuns(res.data || []);
      setView('runs');
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }

  async function handlePreAuditCheck() {
    if (!selectedFramework) {
      setError('Please select a framework');
      return;
    }
    setActionLoading(true);
    try {
      const res = await api.runPreAuditCheck({ framework_id: selectedFramework });
      setPreAuditReport(res.data);
      setView('pre-audit');
    } catch (err: any) {
      setError(err.message);
    } finally {
      setActionLoading(false);
    }
  }

  return (
    <div className="p-6 max-w-7xl mx-auto">
      <div className="flex justify-between items-start mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Evidence Testing</h1>
          <p className="text-gray-600 mt-1">
            Manage test suites, run evidence checks, and assess audit readiness.
          </p>
        </div>
        <div className="flex gap-2">
          <button
            onClick={() => { setView('suites'); setSelectedRun(null); }}
            className={`px-4 py-2 text-sm rounded ${view === 'suites' ? 'bg-blue-600 text-white' : 'bg-gray-100 text-gray-700'}`}
          >
            Test Suites
          </button>
          <button
            onClick={() => setView('pre-audit')}
            className={`px-4 py-2 text-sm rounded ${view === 'pre-audit' ? 'bg-blue-600 text-white' : 'bg-gray-100 text-gray-700'}`}
          >
            Pre-Audit Check
          </button>
        </div>
      </div>

      {error && (
        <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded mb-4 flex justify-between">
          <span>{error}</span>
          <button onClick={() => setError(null)} className="text-red-500 hover:text-red-700 font-bold">X</button>
        </div>
      )}

      {loading ? (
        <div className="text-center py-12 text-gray-500">Loading...</div>
      ) : (
        <>
          {view === 'suites' && (
            <TestSuitesView
              suites={suites}
              onRunSuite={handleRunSuite}
              onViewRuns={handleViewRuns}
              actionLoading={actionLoading}
            />
          )}

          {view === 'runs' && (
            <TestRunsView
              runs={selectedSuiteRuns}
              selectedRun={selectedRun}
              onBack={() => { setView('suites'); setSelectedRun(null); }}
              onSelectRun={setSelectedRun}
            />
          )}

          {view === 'pre-audit' && (
            <PreAuditView
              report={preAuditReport}
              frameworks={frameworks}
              selectedFramework={selectedFramework}
              onSelectFramework={setSelectedFramework}
              onRunCheck={handlePreAuditCheck}
              actionLoading={actionLoading}
            />
          )}
        </>
      )}
    </div>
  );
}

// ============================================================
// TEST SUITES VIEW
// ============================================================

function TestSuitesView({
  suites, onRunSuite, onViewRuns, actionLoading,
}: {
  suites: TestSuite[];
  onRunSuite: (id: string) => void;
  onViewRuns: (id: string) => void;
  actionLoading: boolean;
}) {
  return (
    <div className="space-y-4">
      {suites.length === 0 ? (
        <div className="bg-white shadow rounded-lg p-8 text-center">
          <p className="text-gray-500 mb-4">No test suites created yet.</p>
          <p className="text-sm text-gray-400">Create a test suite via the API to get started with evidence testing.</p>
        </div>
      ) : (
        suites.map((suite) => (
          <div key={suite.id} className="bg-white shadow rounded-lg p-6">
            <div className="flex justify-between items-start">
              <div>
                <div className="flex items-center gap-2">
                  <h3 className="text-lg font-semibold text-gray-900">{suite.name}</h3>
                  <Badge
                    text={suite.is_active ? 'Active' : 'Inactive'}
                    className={suite.is_active ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-600'}
                  />
                  <Badge text={suite.test_type.replace(/_/g, ' ')} className="bg-blue-100 text-blue-700" />
                </div>
                {suite.description && <p className="text-sm text-gray-600 mt-1">{suite.description}</p>}
              </div>
              <div className="flex gap-2">
                <button
                  onClick={() => onViewRuns(suite.id)}
                  className="px-3 py-1 text-sm border border-gray-300 rounded hover:bg-gray-50"
                >
                  View Runs
                </button>
                <button
                  onClick={() => onRunSuite(suite.id)}
                  disabled={actionLoading}
                  className="px-3 py-1 text-sm bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50"
                >
                  {actionLoading ? 'Running...' : 'Run Now'}
                </button>
              </div>
            </div>

            <div className="grid grid-cols-4 gap-4 mt-4 pt-4 border-t">
              <div>
                <div className="text-xs text-gray-500">Test Cases</div>
                <div className="text-sm font-medium">{suite.test_case_count}</div>
              </div>
              <div>
                <div className="text-xs text-gray-500">Pass Threshold</div>
                <div className="text-sm font-medium">{suite.pass_threshold_percent}%</div>
              </div>
              <div>
                <div className="text-xs text-gray-500">Last Run</div>
                <div className="text-sm font-medium">
                  {suite.last_run_at ? new Date(suite.last_run_at).toLocaleString() : 'Never'}
                </div>
              </div>
              <div>
                <div className="text-xs text-gray-500">Last Status</div>
                <div className="text-sm font-medium">
                  {suite.last_run_status ? (
                    <Badge
                      text={suite.last_run_status}
                      className={suite.last_run_status === 'completed' ? 'bg-green-100 text-green-800' : 'bg-yellow-100 text-yellow-800'}
                    />
                  ) : '-'}
                </div>
              </div>
            </div>
          </div>
        ))
      )}
    </div>
  );
}

// ============================================================
// TEST RUNS VIEW
// ============================================================

function TestRunsView({
  runs, selectedRun, onBack, onSelectRun,
}: {
  runs: TestRun[];
  selectedRun: TestRun | null;
  onBack: () => void;
  onSelectRun: (run: TestRun | null) => void;
}) {
  const currentRun = selectedRun || (runs.length > 0 ? runs[0] : null);

  return (
    <div>
      <button onClick={onBack} className="text-blue-600 hover:text-blue-800 mb-4 text-sm">&larr; Back to suites</button>

      {currentRun && (
        <div className="bg-white shadow rounded-lg p-6 mb-6">
          <div className="flex justify-between items-start mb-4">
            <div>
              <h3 className="text-lg font-semibold text-gray-900">
                Test Run: {currentRun.suite_name || currentRun.test_suite_id.slice(0, 8)}
              </h3>
              <p className="text-sm text-gray-500">
                {currentRun.started_at && new Date(currentRun.started_at).toLocaleString()}
                {' - Triggered: '}{currentRun.triggered_by}
              </p>
            </div>
            <Badge
              text={currentRun.threshold_met ? 'PASSED' : 'FAILED'}
              className={currentRun.threshold_met ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'}
            />
          </div>

          {/* Run Summary */}
          <div className="grid grid-cols-6 gap-4 mb-6">
            <div className="bg-gray-50 rounded p-3 text-center">
              <div className="text-xs text-gray-500">Total</div>
              <div className="text-xl font-bold text-gray-900">{currentRun.total_tests}</div>
            </div>
            <div className="bg-green-50 rounded p-3 text-center">
              <div className="text-xs text-gray-500">Passed</div>
              <div className="text-xl font-bold text-green-600">{currentRun.passed}</div>
            </div>
            <div className="bg-red-50 rounded p-3 text-center">
              <div className="text-xs text-gray-500">Failed</div>
              <div className="text-xl font-bold text-red-600">{currentRun.failed}</div>
            </div>
            <div className="bg-yellow-50 rounded p-3 text-center">
              <div className="text-xs text-gray-500">Skipped</div>
              <div className="text-xl font-bold text-yellow-600">{currentRun.skipped}</div>
            </div>
            <div className="bg-gray-50 rounded p-3 text-center">
              <div className="text-xs text-gray-500">Errors</div>
              <div className="text-xl font-bold text-gray-600">{currentRun.errors}</div>
            </div>
            <div className="bg-blue-50 rounded p-3 text-center">
              <div className="text-xs text-gray-500">Pass Rate</div>
              <div className={`text-xl font-bold ${currentRun.pass_rate >= 80 ? 'text-green-600' : 'text-red-600'}`}>
                {currentRun.pass_rate.toFixed(1)}%
              </div>
            </div>
          </div>

          {/* Individual Test Case Results */}
          {currentRun.results && currentRun.results.length > 0 && (
            <div>
              <h4 className="text-sm font-semibold text-gray-900 mb-3">Test Case Results</h4>
              <div className="space-y-2">
                {currentRun.results.map((result, idx) => {
                  const icon = statusIcons[result.status] || { color: 'text-gray-600', label: result.status.toUpperCase() };
                  return (
                    <div key={idx} className="flex items-center justify-between bg-gray-50 rounded p-3">
                      <div className="flex items-center gap-3">
                        <span className={`font-mono text-sm font-bold ${icon.color}`}>{icon.label}</span>
                        <span className="text-sm text-gray-900">{result.test_case_name}</span>
                        {result.is_critical && <Badge text="Critical" className="bg-red-100 text-red-800" />}
                      </div>
                      <div className="flex items-center gap-4">
                        <span className="text-xs text-gray-500 max-w-xs truncate">{result.message}</span>
                        <span className="text-xs text-gray-400">{result.duration}</span>
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>
          )}
        </div>
      )}

      {/* Run History */}
      {runs.length > 1 && (
        <div className="bg-white shadow rounded-lg p-6">
          <h3 className="text-lg font-semibold text-gray-900 mb-4">Run History</h3>
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500">Date</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500">Status</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500">Pass Rate</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500">Tests</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500">Trigger</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500">Threshold</th>
                <th className="px-4 py-3"></th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {runs.map((run) => (
                <tr key={run.id} className="hover:bg-gray-50">
                  <td className="px-4 py-3 text-sm text-gray-600">
                    {run.started_at ? new Date(run.started_at).toLocaleString() : '-'}
                  </td>
                  <td className="px-4 py-3 text-sm">
                    <Badge text={run.status} className={run.status === 'completed' ? 'bg-green-100 text-green-800' : 'bg-yellow-100 text-yellow-800'} />
                  </td>
                  <td className="px-4 py-3 text-sm">
                    <span className={run.pass_rate >= 80 ? 'text-green-600 font-medium' : 'text-red-600 font-medium'}>
                      {run.pass_rate.toFixed(1)}%
                    </span>
                  </td>
                  <td className="px-4 py-3 text-sm text-gray-600">
                    {run.passed}/{run.total_tests} passed
                  </td>
                  <td className="px-4 py-3 text-sm text-gray-500 capitalize">{run.triggered_by}</td>
                  <td className="px-4 py-3 text-sm">
                    <Badge
                      text={run.threshold_met ? 'Met' : 'Not Met'}
                      className={run.threshold_met ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'}
                    />
                  </td>
                  <td className="px-4 py-3 text-sm">
                    <button onClick={() => onSelectRun(run)} className="text-blue-600 hover:text-blue-800 text-sm">
                      Details
                    </button>
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

// ============================================================
// PRE-AUDIT VIEW
// ============================================================

function PreAuditView({
  report, frameworks, selectedFramework, onSelectFramework, onRunCheck, actionLoading,
}: {
  report: PreAuditReport | null;
  frameworks: Framework[];
  selectedFramework: string;
  onSelectFramework: (id: string) => void;
  onRunCheck: () => void;
  actionLoading: boolean;
}) {
  return (
    <div className="space-y-6">
      {/* Run Pre-Audit Check */}
      <div className="bg-white shadow rounded-lg p-6">
        <h3 className="text-lg font-semibold text-gray-900 mb-4">Run Pre-Audit Check</h3>
        <p className="text-sm text-gray-600 mb-4">
          Select a framework to assess your evidence readiness for an upcoming audit.
          This will analyze all evidence requirements, validate collected evidence,
          and generate a readiness report with recommendations.
        </p>
        <div className="flex gap-3">
          <select
            value={selectedFramework}
            onChange={(e) => onSelectFramework(e.target.value)}
            className="border rounded px-3 py-2 text-sm flex-1 max-w-md"
          >
            <option value="">Select Framework...</option>
            {frameworks.map((fw) => (
              <option key={fw.id} value={fw.id}>{fw.name} ({fw.code})</option>
            ))}
          </select>
          <button
            onClick={onRunCheck}
            disabled={actionLoading || !selectedFramework}
            className="px-4 py-2 bg-blue-600 text-white text-sm rounded hover:bg-blue-700 disabled:opacity-50"
          >
            {actionLoading ? 'Running...' : 'Run Pre-Audit Check'}
          </button>
        </div>
      </div>

      {/* Pre-Audit Report */}
      {report && (
        <>
          {/* Readiness Summary */}
          <div className="bg-white shadow rounded-lg p-6">
            <div className="flex justify-between items-start mb-6">
              <div>
                <h3 className="text-lg font-semibold text-gray-900">
                  Pre-Audit Readiness Report: {report.framework_code}
                </h3>
                <p className="text-sm text-gray-500">
                  Generated: {new Date(report.generated_at).toLocaleString()}
                </p>
              </div>
              <Badge
                text={readinessLabels[report.readiness_level] || report.readiness_level}
                className={readinessColors[report.readiness_level] || 'bg-gray-100'}
              />
            </div>

            <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
              <div className="bg-blue-50 rounded-lg p-4 text-center">
                <div className="text-xs text-gray-500">Overall Readiness</div>
                <div className={`text-3xl font-bold ${report.overall_readiness >= 80 ? 'text-green-600' : report.overall_readiness >= 60 ? 'text-yellow-600' : 'text-red-600'}`}>
                  {report.overall_readiness.toFixed(1)}%
                </div>
              </div>
              <div className="bg-gray-50 rounded-lg p-4 text-center">
                <div className="text-xs text-gray-500">Evidence Completion</div>
                <div className="text-3xl font-bold text-gray-900">{report.evidence_completion.toFixed(1)}%</div>
              </div>
              <div className="bg-gray-50 rounded-lg p-4 text-center">
                <div className="text-xs text-gray-500">Validation Pass Rate</div>
                <div className="text-3xl font-bold text-gray-900">{report.validation_pass_rate.toFixed(1)}%</div>
              </div>
              <div className="bg-gray-50 rounded-lg p-4 text-center">
                <div className="text-xs text-gray-500">Est. Remediation</div>
                <div className="text-3xl font-bold text-gray-900">{report.estimated_remediation_hours.toFixed(0)}h</div>
              </div>
            </div>

            <div className="grid grid-cols-3 gap-4 mb-6">
              <div className="text-center p-3 bg-green-50 rounded">
                <div className="text-2xl font-bold text-green-600">{report.controls_with_evidence}</div>
                <div className="text-xs text-gray-500">Controls with Evidence</div>
              </div>
              <div className="text-center p-3 bg-red-50 rounded">
                <div className="text-2xl font-bold text-red-600">{report.controls_missing_evidence}</div>
                <div className="text-xs text-gray-500">Controls Missing Evidence</div>
              </div>
              <div className="text-center p-3 bg-gray-50 rounded">
                <div className="text-2xl font-bold text-gray-600">{report.total_controls}</div>
                <div className="text-xs text-gray-500">Total Controls</div>
              </div>
            </div>
          </div>

          {/* Recommendations */}
          {report.recommendations.length > 0 && (
            <div className="bg-white shadow rounded-lg p-6">
              <h3 className="text-lg font-semibold text-gray-900 mb-4">Recommendations</h3>
              <ul className="space-y-3">
                {report.recommendations.map((rec, idx) => (
                  <li key={idx} className="flex items-start gap-3 text-sm text-gray-700">
                    <span className="bg-blue-100 text-blue-700 rounded-full w-6 h-6 flex items-center justify-center flex-shrink-0 text-xs font-bold">
                      {idx + 1}
                    </span>
                    {rec}
                  </li>
                ))}
              </ul>
            </div>
          )}

          {/* Critical Gaps */}
          {report.critical_gaps.length > 0 && (
            <div className="bg-white shadow rounded-lg p-6">
              <h3 className="text-lg font-semibold text-gray-900 mb-4">Critical Gaps ({report.critical_gaps.length})</h3>
              <div className="space-y-2">
                {report.critical_gaps.map((gap, idx) => (
                  <div key={idx} className="bg-red-50 border border-red-100 rounded p-3">
                    <div className="flex justify-between items-start">
                      <div>
                        <span className="font-mono text-sm font-semibold text-blue-600">{gap.control_code}</span>
                        <span className="text-sm text-gray-700 ml-2">{gap.control_name}</span>
                      </div>
                      <Badge text={gap.severity} className="bg-red-100 text-red-800" />
                    </div>
                    <p className="text-sm text-gray-600 mt-1">{gap.description}</p>
                    <p className="text-sm text-blue-700 mt-1">{gap.recommendation}</p>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Control Readiness Detail */}
          {report.control_readiness.length > 0 && (
            <div className="bg-white shadow rounded-lg p-6">
              <h3 className="text-lg font-semibold text-gray-900 mb-4">Control Readiness Detail</h3>
              <div className="overflow-x-auto">
                <table className="min-w-full divide-y divide-gray-200">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500">Control</th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500">Title</th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500">Evidence</th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500">Validated</th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500">Readiness</th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500">Status</th>
                    </tr>
                  </thead>
                  <tbody className="bg-white divide-y divide-gray-200">
                    {report.control_readiness.map((cr) => (
                      <tr key={cr.control_code} className="hover:bg-gray-50">
                        <td className="px-4 py-2 text-sm font-mono text-blue-600">{cr.control_code}</td>
                        <td className="px-4 py-2 text-sm text-gray-900 max-w-xs truncate">{cr.control_title}</td>
                        <td className="px-4 py-2 text-sm text-gray-600">{cr.evidence_count}/{cr.required_count}</td>
                        <td className="px-4 py-2 text-sm text-gray-600">{cr.validation_passed}</td>
                        <td className="px-4 py-2 text-sm">
                          <div className="flex items-center gap-2">
                            <div className="w-16 bg-gray-200 rounded-full h-2">
                              <div
                                className={`h-2 rounded-full ${cr.readiness_percent >= 80 ? 'bg-green-500' : cr.readiness_percent >= 50 ? 'bg-yellow-500' : 'bg-red-500'}`}
                                style={{ width: `${Math.min(100, cr.readiness_percent)}%` }}
                              />
                            </div>
                            <span className="text-xs">{cr.readiness_percent.toFixed(0)}%</span>
                          </div>
                        </td>
                        <td className="px-4 py-2 text-sm">
                          <Badge
                            text={cr.status}
                            className={cr.status === 'complete' ? 'bg-green-100 text-green-800' : cr.status === 'partial' ? 'bg-yellow-100 text-yellow-800' : 'bg-red-100 text-red-800'}
                          />
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  );
}
