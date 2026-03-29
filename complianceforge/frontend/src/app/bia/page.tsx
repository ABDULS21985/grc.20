'use client';

import { useEffect, useState, useCallback } from 'react';
import api from '@/lib/api';

const CRITICALITY_COLORS: Record<string, string> = {
  mission_critical: 'bg-red-100 text-red-800',
  business_critical: 'bg-orange-100 text-orange-800',
  important: 'bg-yellow-100 text-yellow-800',
  minor: 'bg-blue-100 text-blue-800',
  non_essential: 'bg-gray-100 text-gray-600',
};

const IMPACT_COLORS: Record<string, string> = {
  catastrophic: 'text-red-700',
  severe: 'text-orange-600',
  moderate: 'text-yellow-600',
  minor: 'text-blue-600',
  negligible: 'text-gray-500',
};

interface BusinessProcess {
  id: string;
  process_ref: string;
  name: string;
  description: string;
  department: string;
  category: string;
  criticality: string;
  status: string;
  financial_impact_per_hour_eur: number | null;
  financial_impact_per_day_eur: number | null;
  reputational_impact: string | null;
  operational_impact: string | null;
  rto_hours: number | null;
  rpo_hours: number | null;
  mtpd_hours: number | null;
  last_bia_date: string | null;
  next_bia_due: string | null;
  created_at: string;
}

interface SinglePointOfFailure {
  entity_name: string;
  entity_type: string;
  dependency_type: string;
  affected_process_count: number;
  is_critical: boolean;
  alternative_available: boolean;
  affected_processes: {
    process_id: string;
    process_ref: string;
    process_name: string;
    criticality: string;
    is_direct: boolean;
  }[];
}

interface BIAReport {
  generated_at: string;
  total_processes: number;
  processes_by_criticality: Record<string, number>;
  processes_by_category: Record<string, number>;
  critical_processes: {
    id: string;
    process_ref: string;
    name: string;
    criticality: string;
    category: string;
    rto_hours: number | null;
    rpo_hours: number | null;
    financial_impact_per_day_eur: number | null;
  }[];
  rto_rpo_summary: {
    processes_with_rto: number;
    processes_with_rpo: number;
    min_rto_hours: number | null;
    max_rto_hours: number | null;
    avg_rto_hours: number | null;
    min_rpo_hours: number | null;
    max_rpo_hours: number | null;
    avg_rpo_hours: number | null;
  };
  financial_impact: {
    total_hourly_impact_eur: number;
    total_daily_impact_eur: number;
    total_weekly_impact_eur: number;
  };
  single_points_of_failure: SinglePointOfFailure[];
  processes_without_bia: number;
  overdue_bias: number;
}

type Tab = 'processes' | 'report' | 'spof' | 'graph';

export default function BIAPage() {
  const [tab, setTab] = useState<Tab>('processes');
  const [processes, setProcesses] = useState<BusinessProcess[]>([]);
  const [report, setReport] = useState<BIAReport | null>(null);
  const [spofs, setSPOFs] = useState<SinglePointOfFailure[]>([]);
  const [loading, setLoading] = useState(true);
  const [totalItems, setTotalItems] = useState(0);
  const [page, setPage] = useState(1);

  const loadData = useCallback(async () => {
    setLoading(true);
    try {
      if (tab === 'processes') {
        const res = await api.get(`/bia/processes?page=${page}&page_size=20`);
        const data = res.data;
        setProcesses(data?.data || []);
        setTotalItems(data?.pagination?.total_items || 0);
      } else if (tab === 'report') {
        const res = await api.get('/bia/report');
        setReport(res.data?.data || null);
      } else if (tab === 'spof') {
        const res = await api.get('/bia/single-points-of-failure');
        setSPOFs(res.data?.data || []);
      }
    } catch {
      /* ignore */
    }
    setLoading(false);
  }, [tab, page]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const tabs = [
    { key: 'processes' as const, label: 'Business Processes' },
    { key: 'report' as const, label: 'BIA Report' },
    { key: 'spof' as const, label: 'Single Points of Failure' },
    { key: 'graph' as const, label: 'Dependency Graph' },
  ];

  return (
    <div>
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Business Impact Analysis</h1>
          <p className="text-gray-500 mt-1">
            Assess business processes, identify dependencies, and evaluate disruption impacts
          </p>
        </div>
      </div>

      <div className="flex gap-1 border-b border-gray-200 mb-6 overflow-x-auto">
        {tabs.map((t) => (
          <button
            key={t.key}
            onClick={() => { setTab(t.key); setPage(1); }}
            className={`px-4 py-2.5 text-sm font-medium border-b-2 whitespace-nowrap transition-colors ${
              tab === t.key
                ? 'border-indigo-600 text-indigo-600'
                : 'border-transparent text-gray-500 hover:text-gray-700'
            }`}
          >
            {t.label}
          </button>
        ))}
      </div>

      {loading ? (
        <div className="card animate-pulse h-96" />
      ) : tab === 'processes' ? (
        <div>
          <div className="flex justify-between items-center mb-4">
            <p className="text-sm text-gray-500">{totalItems} processes</p>
          </div>
          <div className="card overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="table-header">
                  <th className="px-4 py-3 text-left">Ref</th>
                  <th className="px-4 py-3 text-left">Name</th>
                  <th className="px-4 py-3 text-left">Department</th>
                  <th className="px-4 py-3 text-left">Criticality</th>
                  <th className="px-4 py-3 text-right">RTO (hrs)</th>
                  <th className="px-4 py-3 text-right">RPO (hrs)</th>
                  <th className="px-4 py-3 text-right">Impact / Day</th>
                  <th className="px-4 py-3 text-left">Last BIA</th>
                  <th className="px-4 py-3 text-left">Status</th>
                </tr>
              </thead>
              <tbody>
                {processes.map((p) => (
                  <tr key={p.id} className="border-t border-gray-100 hover:bg-gray-50">
                    <td className="px-4 py-3 font-mono text-xs text-indigo-600">{p.process_ref}</td>
                    <td className="px-4 py-3">
                      <a href={`/bia/processes/${p.id}`} className="font-medium text-gray-900 hover:text-indigo-600">
                        {p.name}
                      </a>
                      {p.description && (
                        <p className="text-xs text-gray-400 mt-0.5 truncate max-w-xs">{p.description}</p>
                      )}
                    </td>
                    <td className="px-4 py-3 text-gray-600">{p.department || '--'}</td>
                    <td className="px-4 py-3">
                      <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${CRITICALITY_COLORS[p.criticality] || 'bg-gray-100 text-gray-600'}`}>
                        {p.criticality.replace(/_/g, ' ')}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-right font-mono text-sm">
                      {p.rto_hours != null ? p.rto_hours.toFixed(1) : '--'}
                    </td>
                    <td className="px-4 py-3 text-right font-mono text-sm">
                      {p.rpo_hours != null ? p.rpo_hours.toFixed(1) : '--'}
                    </td>
                    <td className="px-4 py-3 text-right font-mono text-sm">
                      {p.financial_impact_per_day_eur != null
                        ? `EUR ${p.financial_impact_per_day_eur.toLocaleString()}`
                        : '--'}
                    </td>
                    <td className="px-4 py-3 text-gray-500 text-xs">
                      {p.last_bia_date
                        ? new Date(p.last_bia_date).toLocaleDateString('en-GB')
                        : <span className="text-amber-600">Not assessed</span>}
                    </td>
                    <td className="px-4 py-3">
                      <span className={`text-xs font-medium ${p.status === 'active' ? 'text-green-600' : p.status === 'inactive' ? 'text-gray-400' : 'text-amber-600'}`}>
                        {p.status}
                      </span>
                    </td>
                  </tr>
                ))}
                {processes.length === 0 && (
                  <tr>
                    <td colSpan={9} className="px-4 py-12 text-center text-gray-500">
                      No business processes found. Create one to begin your BIA.
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
          {totalItems > 20 && (
            <div className="flex justify-center gap-2 mt-4">
              <button
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page <= 1}
                className="px-3 py-1.5 text-sm border rounded disabled:opacity-50"
              >
                Previous
              </button>
              <span className="px-3 py-1.5 text-sm text-gray-600">Page {page}</span>
              <button
                onClick={() => setPage((p) => p + 1)}
                disabled={page * 20 >= totalItems}
                className="px-3 py-1.5 text-sm border rounded disabled:opacity-50"
              >
                Next
              </button>
            </div>
          )}
        </div>
      ) : tab === 'report' && report ? (
        <div>
          {/* Summary Cards */}
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-4 mb-6">
            <div className="card">
              <p className="text-sm text-gray-500">Total Processes</p>
              <p className="text-2xl font-bold text-gray-900 mt-1">{report.total_processes}</p>
              {report.processes_without_bia > 0 && (
                <p className="text-xs text-amber-600 mt-1">
                  {report.processes_without_bia} without BIA assessment
                </p>
              )}
            </div>
            <div className="card">
              <p className="text-sm text-gray-500">Total Daily Impact</p>
              <p className="text-2xl font-bold text-red-600 mt-1">
                EUR {report.financial_impact.total_daily_impact_eur.toLocaleString()}
              </p>
              <p className="text-xs text-gray-500 mt-1">
                EUR {report.financial_impact.total_hourly_impact_eur.toLocaleString()} / hour
              </p>
            </div>
            <div className="card">
              <p className="text-sm text-gray-500">RTO Coverage</p>
              <p className="text-2xl font-bold text-indigo-600 mt-1">
                {report.rto_rpo_summary.processes_with_rto} / {report.total_processes}
              </p>
              {report.rto_rpo_summary.avg_rto_hours != null && (
                <p className="text-xs text-gray-500 mt-1">
                  Avg: {report.rto_rpo_summary.avg_rto_hours.toFixed(1)}h | Min: {report.rto_rpo_summary.min_rto_hours?.toFixed(1)}h | Max: {report.rto_rpo_summary.max_rto_hours?.toFixed(1)}h
                </p>
              )}
            </div>
            <div className="card">
              <p className="text-sm text-gray-500">Single Points of Failure</p>
              <p className={`text-2xl font-bold mt-1 ${report.single_points_of_failure.length > 0 ? 'text-red-600' : 'text-green-600'}`}>
                {report.single_points_of_failure.length}
              </p>
              {report.overdue_bias > 0 && (
                <p className="text-xs text-amber-600 mt-1">{report.overdue_bias} overdue BIAs</p>
              )}
            </div>
          </div>

          {/* Criticality Breakdown */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
            <div className="card">
              <h3 className="font-semibold text-gray-900 mb-4">Processes by Criticality</h3>
              <div className="space-y-3">
                {Object.entries(report.processes_by_criticality).map(([level, count]) => (
                  <div key={level} className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${CRITICALITY_COLORS[level] || 'bg-gray-100'}`}>
                        {level.replace(/_/g, ' ')}
                      </span>
                    </div>
                    <div className="flex items-center gap-3">
                      <div className="w-32 h-2 bg-gray-100 rounded-full overflow-hidden">
                        <div
                          className="h-full bg-indigo-500 rounded-full"
                          style={{ width: `${(count / report.total_processes) * 100}%` }}
                        />
                      </div>
                      <span className="text-sm font-medium text-gray-700 w-8 text-right">{count}</span>
                    </div>
                  </div>
                ))}
              </div>
            </div>

            <div className="card">
              <h3 className="font-semibold text-gray-900 mb-4">Processes by Category</h3>
              <div className="space-y-3">
                {Object.entries(report.processes_by_category).map(([cat, count]) => (
                  <div key={cat} className="flex items-center justify-between">
                    <span className="text-sm text-gray-600 capitalize">{cat.replace(/_/g, ' ')}</span>
                    <span className="text-sm font-medium text-gray-700">{count}</span>
                  </div>
                ))}
              </div>
            </div>
          </div>

          {/* Critical Processes Table */}
          <div className="card overflow-x-auto">
            <h3 className="font-semibold text-gray-900 mb-4">All Processes Ranked by Criticality</h3>
            <table className="w-full text-sm">
              <thead>
                <tr className="table-header">
                  <th className="px-4 py-3 text-left">Ref</th>
                  <th className="px-4 py-3 text-left">Name</th>
                  <th className="px-4 py-3 text-left">Criticality</th>
                  <th className="px-4 py-3 text-left">Category</th>
                  <th className="px-4 py-3 text-right">RTO (hrs)</th>
                  <th className="px-4 py-3 text-right">RPO (hrs)</th>
                  <th className="px-4 py-3 text-right">Impact / Day</th>
                </tr>
              </thead>
              <tbody>
                {report.critical_processes.map((p) => (
                  <tr key={p.id} className="border-t border-gray-100">
                    <td className="px-4 py-3 font-mono text-xs text-indigo-600">{p.process_ref}</td>
                    <td className="px-4 py-3 font-medium text-gray-900">{p.name}</td>
                    <td className="px-4 py-3">
                      <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${CRITICALITY_COLORS[p.criticality] || 'bg-gray-100'}`}>
                        {p.criticality.replace(/_/g, ' ')}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-gray-600 capitalize text-xs">{p.category.replace(/_/g, ' ')}</td>
                    <td className="px-4 py-3 text-right font-mono">
                      {p.rto_hours != null ? p.rto_hours.toFixed(1) : '--'}
                    </td>
                    <td className="px-4 py-3 text-right font-mono">
                      {p.rpo_hours != null ? p.rpo_hours.toFixed(1) : '--'}
                    </td>
                    <td className="px-4 py-3 text-right font-mono">
                      {p.financial_impact_per_day_eur != null
                        ? `EUR ${p.financial_impact_per_day_eur.toLocaleString()}`
                        : '--'}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      ) : tab === 'spof' ? (
        <div>
          {spofs.length === 0 ? (
            <div className="card text-center py-12">
              <p className="text-gray-500 mb-2">No single points of failure detected</p>
              <p className="text-sm text-gray-400">
                SPoF analysis requires at least two critical processes sharing a common dependency.
              </p>
            </div>
          ) : (
            <div className="space-y-4">
              <div className="card bg-red-50 border-red-200">
                <p className="text-sm text-red-800 font-medium">
                  {spofs.length} single point{spofs.length !== 1 ? 's' : ''} of failure detected across critical business processes
                </p>
              </div>
              {spofs.map((spof, idx) => (
                <div key={idx} className="card">
                  <div className="flex items-start justify-between mb-3">
                    <div>
                      <h3 className="font-semibold text-gray-900">{spof.entity_name}</h3>
                      <div className="flex items-center gap-2 mt-1">
                        <span className="text-xs bg-gray-100 text-gray-600 px-2 py-0.5 rounded">
                          {spof.dependency_type}
                        </span>
                        {spof.entity_type && (
                          <span className="text-xs bg-gray-100 text-gray-600 px-2 py-0.5 rounded">
                            {spof.entity_type}
                          </span>
                        )}
                        {spof.is_critical && (
                          <span className="text-xs bg-red-100 text-red-700 px-2 py-0.5 rounded font-medium">
                            Critical dependency
                          </span>
                        )}
                        {!spof.alternative_available && (
                          <span className="text-xs bg-amber-100 text-amber-700 px-2 py-0.5 rounded font-medium">
                            No alternative
                          </span>
                        )}
                      </div>
                    </div>
                    <span className="text-lg font-bold text-red-600">{spof.affected_process_count} processes</span>
                  </div>
                  <div className="border-t border-gray-100 pt-3">
                    <p className="text-xs text-gray-500 mb-2">Affected critical processes:</p>
                    <div className="flex flex-wrap gap-2">
                      {spof.affected_processes.map((ap) => (
                        <span
                          key={ap.process_id}
                          className={`inline-flex items-center gap-1 text-xs px-2 py-1 rounded ${CRITICALITY_COLORS[ap.criticality] || 'bg-gray-100'}`}
                        >
                          <span className="font-mono">{ap.process_ref}</span>
                          <span>{ap.process_name}</span>
                          {!ap.is_direct && <span className="text-gray-400">(transitive)</span>}
                        </span>
                      ))}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      ) : tab === 'graph' ? (
        <div className="card text-center py-16">
          <svg className="mx-auto h-16 w-16 text-gray-300 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
          </svg>
          <h3 className="text-lg font-medium text-gray-900 mb-2">Dependency Graph</h3>
          <p className="text-sm text-gray-500 max-w-md mx-auto">
            The interactive dependency graph visualization shows process-to-process
            and process-to-entity relationships. Data is available via the API endpoint
            <code className="bg-gray-100 px-1.5 py-0.5 rounded text-xs ml-1">/bia/dependency-graph</code>.
          </p>
        </div>
      ) : null}
    </div>
  );
}
