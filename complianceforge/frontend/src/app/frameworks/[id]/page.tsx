'use client';

import { useEffect, useState, useCallback } from 'react';
import { useParams } from 'next/navigation';
import api from '@/lib/api';
import type { ComplianceFramework, FrameworkControl, ComplianceScore } from '@/types';

const TABS = ['Controls', 'Implementation Status', 'Gap Analysis', 'Cross-Mapping'] as const;
type Tab = typeof TABS[number];

interface GapItem {
  control_code: string;
  control_title: string;
  gap_description: string;
  risk_level: string;
  recommendation: string;
  priority: string;
}

interface CrossMappingEntry {
  source_framework: string;
  source_control_code: string;
  source_control_title: string;
  target_framework: string;
  target_control_code: string;
  target_control_title: string;
  mapping_type: string;
}

export default function FrameworkDetailPage() {
  const params = useParams();
  const id = params.id as string;

  const [framework, setFramework] = useState<ComplianceFramework | null>(null);
  const [controls, setControls] = useState<FrameworkControl[]>([]);
  const [totalControls, setTotalControls] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [tab, setTab] = useState<Tab>('Controls');

  // Controls search
  const [searchQuery, setSearchQuery] = useState('');
  const [searchResults, setSearchResults] = useState<FrameworkControl[] | null>(null);
  const [searching, setSearching] = useState(false);

  // Implementation Status
  const [scores, setScores] = useState<ComplianceScore[]>([]);
  const [scoresLoading, setScoresLoading] = useState(false);

  // Gap Analysis
  const [gaps, setGaps] = useState<GapItem[]>([]);
  const [gapsLoading, setGapsLoading] = useState(false);

  // Cross-Mapping
  const [crossMappings, setCrossMappings] = useState<CrossMappingEntry[]>([]);
  const [crossMappingLoading, setCrossMappingLoading] = useState(false);

  useEffect(() => {
    setLoading(true);
    setError(null);
    Promise.all([
      api.getFramework(id),
      api.getFrameworkControls(id, 1, 20),
    ])
      .then(([fwRes, ctrlRes]) => {
        setFramework(fwRes.data);
        setControls(ctrlRes.data?.data || []);
        setTotalControls(ctrlRes.data?.pagination?.total_items || 0);
      })
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, [id]);

  const fetchPage = useCallback((p: number) => {
    setPage(p);
    api.getFrameworkControls(id, p, 20)
      .then((res) => {
        setControls(res.data?.data || []);
        setTotalControls(res.data?.pagination?.total_items || 0);
      })
      .catch(() => {});
  }, [id]);

  const handleSearch = useCallback(() => {
    if (!searchQuery.trim()) {
      setSearchResults(null);
      return;
    }
    setSearching(true);
    api.searchControls(searchQuery)
      .then((res) => {
        const results = (res.data?.data || res.data || []) as FrameworkControl[];
        setSearchResults(results.filter((c: FrameworkControl) => c.framework_id === id));
      })
      .catch(() => setSearchResults([]))
      .finally(() => setSearching(false));
  }, [searchQuery, id]);

  useEffect(() => {
    if (tab === 'Implementation Status' && scores.length === 0) {
      setScoresLoading(true);
      api.getComplianceScores()
        .then((res) => setScores(res.data?.data || res.data || []))
        .catch(() => {})
        .finally(() => setScoresLoading(false));
    }
  }, [tab, scores.length]);

  useEffect(() => {
    if (tab === 'Gap Analysis' && gaps.length === 0) {
      setGapsLoading(true);
      api.getGapAnalysis(id)
        .then((res) => setGaps(res.data?.data || res.data || []))
        .catch(() => {})
        .finally(() => setGapsLoading(false));
    }
  }, [tab, gaps.length, id]);

  useEffect(() => {
    if (tab === 'Cross-Mapping' && crossMappings.length === 0) {
      setCrossMappingLoading(true);
      api.getCrossMapping()
        .then((res) => {
          const all = res.data?.data || res.data || [];
          setCrossMappings(
            all.filter((m: CrossMappingEntry) =>
              m.source_framework === id || m.target_framework === id ||
              m.source_framework === framework?.code || m.target_framework === framework?.code
            )
          );
        })
        .catch(() => {})
        .finally(() => setCrossMappingLoading(false));
    }
  }, [tab, crossMappings.length, id, framework?.code]);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <p className="text-gray-500">Loading framework...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="rounded-lg bg-red-50 border border-red-200 p-4 text-red-700">{error}</div>
    );
  }

  if (!framework) {
    return (
      <div className="rounded-lg bg-gray-50 border border-gray-200 p-4 text-gray-600">
        Framework not found.
      </div>
    );
  }

  const displayedControls = searchResults !== null ? searchResults : controls;
  const totalPages = Math.ceil(totalControls / 20);

  const frameworkScore = scores.find(
    (s) => s.framework_id === id || s.framework_code === framework.code
  );

  const riskOrder: Record<string, number> = { critical: 0, high: 1, medium: 2, low: 3 };
  const sortedGaps = [...gaps].sort(
    (a, b) => (riskOrder[a.risk_level] ?? 4) - (riskOrder[b.risk_level] ?? 4)
  );

  return (
    <div>
      {/* Back button */}
      <a href="/frameworks" className="inline-flex items-center text-sm text-gray-500 hover:text-indigo-600 mb-4">
        &larr; Back to Frameworks
      </a>

      {/* Header */}
      <div className="flex items-start justify-between mb-6">
        <div className="flex items-start gap-4">
          <div
            className="h-12 w-12 rounded-xl flex items-center justify-center flex-shrink-0"
            style={{ backgroundColor: (framework.color_hex || '#6366F1') + '20' }}
          >
            <span className="text-sm font-bold" style={{ color: framework.color_hex || '#6366F1' }}>
              {framework.code.substring(0, 2)}
            </span>
          </div>
          <div>
            <h1 className="text-2xl font-bold text-gray-900">{framework.name}</h1>
            <p className="text-gray-500 mt-1">
              {framework.issuing_body} &middot; v{framework.version} &middot; {framework.total_controls} controls
            </p>
            {framework.description && (
              <p className="text-sm text-gray-600 mt-2 max-w-2xl">{framework.description}</p>
            )}
          </div>
        </div>
        <div className="flex gap-2">
          <span className={`badge ${framework.category === 'security' ? 'badge-info' : framework.category === 'privacy' ? 'badge-high' : framework.category === 'governance' ? 'badge-medium' : 'badge-low'}`}>
            {framework.category}
          </span>
          {framework.is_active && <span className="badge badge-low">Active</span>}
        </div>
      </div>

      {/* Tabs */}
      <div className="flex gap-1 border-b border-gray-200 mb-6">
        {TABS.map((t) => (
          <button
            key={t}
            onClick={() => setTab(t)}
            className={`px-4 py-2.5 text-sm font-medium border-b-2 transition-colors ${
              tab === t
                ? 'border-indigo-600 text-indigo-600'
                : 'border-transparent text-gray-500 hover:text-gray-700'
            }`}
          >
            {t}
          </button>
        ))}
      </div>

      {/* Controls Tab */}
      {tab === 'Controls' && (
        <div>
          {/* Search */}
          <div className="flex gap-3 mb-4">
            <input
              type="text"
              placeholder="Search controls..."
              className="input flex-1"
              value={searchQuery}
              onChange={(e) => {
                setSearchQuery(e.target.value);
                if (!e.target.value.trim()) setSearchResults(null);
              }}
              onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
            />
            <button onClick={handleSearch} className="btn-primary" disabled={searching}>
              {searching ? 'Searching...' : 'Search'}
            </button>
          </div>

          {displayedControls.length === 0 ? (
            <div className="rounded-lg bg-gray-50 border border-gray-200 p-8 text-center">
              <p className="text-gray-500">
                {searchResults !== null ? 'No controls match your search.' : 'No controls loaded for this framework.'}
              </p>
            </div>
          ) : (
            <div className="card overflow-hidden p-0">
              <table className="w-full">
                <thead>
                  <tr className="border-b border-gray-100">
                    <th className="table-header px-4 py-3">Code</th>
                    <th className="table-header px-4 py-3">Title</th>
                    <th className="table-header px-4 py-3">Type</th>
                    <th className="table-header px-4 py-3">Implementation</th>
                    <th className="table-header px-4 py-3">Priority</th>
                    <th className="table-header px-4 py-3">Mandatory</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-50">
                  {displayedControls.map((ctrl) => (
                    <tr key={ctrl.id} className="hover:bg-gray-50 transition-colors">
                      <td className="px-4 py-3 text-sm font-mono text-gray-500">
                        <a href={`/controls/${ctrl.id}`} className="hover:text-indigo-600">{ctrl.code}</a>
                      </td>
                      <td className="px-4 py-3 text-sm font-medium text-gray-900">{ctrl.title}</td>
                      <td className="px-4 py-3">
                        <span className="badge badge-info">{ctrl.control_type}</span>
                      </td>
                      <td className="px-4 py-3">
                        <span className="badge badge-medium">{ctrl.implementation_type}</span>
                      </td>
                      <td className="px-4 py-3">
                        <span className={`badge ${ctrl.priority === 'critical' || ctrl.priority === 'high' ? 'badge-critical' : ctrl.priority === 'medium' ? 'badge-medium' : 'badge-low'}`}>
                          {ctrl.priority}
                        </span>
                      </td>
                      <td className="px-4 py-3 text-sm text-gray-500">
                        {ctrl.is_mandatory ? <span className="text-red-600 font-medium">Required</span> : 'Optional'}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}

          {/* Pagination */}
          {searchResults === null && totalPages > 1 && (
            <div className="flex items-center justify-between mt-4">
              <p className="text-sm text-gray-500">
                Page {page} of {totalPages} ({totalControls} controls)
              </p>
              <div className="flex gap-2">
                <button
                  onClick={() => fetchPage(page - 1)}
                  disabled={page <= 1}
                  className="btn-secondary text-sm disabled:opacity-50"
                >
                  Previous
                </button>
                <button
                  onClick={() => fetchPage(page + 1)}
                  disabled={page >= totalPages}
                  className="btn-secondary text-sm disabled:opacity-50"
                >
                  Next
                </button>
              </div>
            </div>
          )}
        </div>
      )}

      {/* Implementation Status Tab */}
      {tab === 'Implementation Status' && (
        <div>
          {scoresLoading ? (
            <div className="flex items-center justify-center h-32">
              <p className="text-gray-500">Loading implementation status...</p>
            </div>
          ) : !frameworkScore ? (
            <div className="rounded-lg bg-gray-50 border border-gray-200 p-8 text-center">
              <p className="text-gray-500">No compliance score data available for this framework.</p>
            </div>
          ) : (
            <div>
              {/* Score Overview */}
              <div className="grid grid-cols-2 gap-4 sm:grid-cols-4 mb-6">
                <div className="card p-4 text-center">
                  <p className="text-3xl font-bold text-indigo-600">{frameworkScore.compliance_score.toFixed(1)}%</p>
                  <p className="text-xs text-gray-500 mt-1">Overall Compliance</p>
                </div>
                <div className="card p-4 text-center">
                  <p className="text-3xl font-bold text-green-600">{frameworkScore.implemented}</p>
                  <p className="text-xs text-gray-500 mt-1">Implemented</p>
                </div>
                <div className="card p-4 text-center">
                  <p className="text-3xl font-bold text-amber-600">{frameworkScore.partially_implemented}</p>
                  <p className="text-xs text-gray-500 mt-1">Partially Implemented</p>
                </div>
                <div className="card p-4 text-center">
                  <p className="text-3xl font-bold text-red-600">{frameworkScore.not_implemented}</p>
                  <p className="text-xs text-gray-500 mt-1">Not Implemented</p>
                </div>
              </div>

              {/* Progress Bar */}
              <div className="card">
                <h3 className="text-sm font-semibold text-gray-700 mb-3">Implementation Progress</h3>
                <div className="h-6 w-full rounded-full bg-gray-100 flex overflow-hidden">
                  <div
                    className="bg-green-500 h-full transition-all duration-500"
                    style={{ width: `${(frameworkScore.implemented / frameworkScore.total_controls) * 100}%` }}
                  />
                  <div
                    className="bg-amber-400 h-full transition-all duration-500"
                    style={{ width: `${(frameworkScore.partially_implemented / frameworkScore.total_controls) * 100}%` }}
                  />
                  <div
                    className="bg-red-400 h-full transition-all duration-500"
                    style={{ width: `${(frameworkScore.not_implemented / frameworkScore.total_controls) * 100}%` }}
                  />
                  <div
                    className="bg-gray-300 h-full transition-all duration-500"
                    style={{ width: `${(frameworkScore.not_applicable / frameworkScore.total_controls) * 100}%` }}
                  />
                </div>
                <div className="flex items-center gap-6 mt-3 text-xs text-gray-500">
                  <span className="flex items-center gap-1"><span className="inline-block h-2.5 w-2.5 rounded-full bg-green-500" /> Implemented ({frameworkScore.implemented})</span>
                  <span className="flex items-center gap-1"><span className="inline-block h-2.5 w-2.5 rounded-full bg-amber-400" /> Partial ({frameworkScore.partially_implemented})</span>
                  <span className="flex items-center gap-1"><span className="inline-block h-2.5 w-2.5 rounded-full bg-red-400" /> Not Implemented ({frameworkScore.not_implemented})</span>
                  <span className="flex items-center gap-1"><span className="inline-block h-2.5 w-2.5 rounded-full bg-gray-300" /> N/A ({frameworkScore.not_applicable})</span>
                </div>
              </div>

              {/* Maturity */}
              <div className="card mt-4">
                <h3 className="text-sm font-semibold text-gray-700 mb-2">Average Maturity Level</h3>
                <div className="flex items-center gap-3">
                  <div className="text-2xl font-bold text-indigo-600">{frameworkScore.maturity_avg.toFixed(1)}</div>
                  <span className="text-sm text-gray-500">/ 5.0</span>
                </div>
                <div className="h-2.5 w-full rounded-full bg-gray-100 mt-2">
                  <div
                    className="h-2.5 rounded-full bg-indigo-500 transition-all duration-500"
                    style={{ width: `${(frameworkScore.maturity_avg / 5) * 100}%` }}
                  />
                </div>
              </div>
            </div>
          )}
        </div>
      )}

      {/* Gap Analysis Tab */}
      {tab === 'Gap Analysis' && (
        <div>
          {gapsLoading ? (
            <div className="flex items-center justify-center h-32">
              <p className="text-gray-500">Loading gap analysis...</p>
            </div>
          ) : sortedGaps.length === 0 ? (
            <div className="rounded-lg bg-green-50 border border-green-200 p-8 text-center">
              <p className="text-green-700 font-medium">No compliance gaps identified for this framework.</p>
            </div>
          ) : (
            <div>
              <p className="text-sm text-gray-500 mb-4">{sortedGaps.length} gap{sortedGaps.length !== 1 ? 's' : ''} identified, sorted by risk level</p>
              <div className="space-y-3">
                {sortedGaps.map((gap, idx) => (
                  <div key={idx} className={`card border-l-4 ${gap.risk_level === 'critical' ? 'border-l-red-600' : gap.risk_level === 'high' ? 'border-l-orange-500' : gap.risk_level === 'medium' ? 'border-l-yellow-500' : 'border-l-green-500'}`}>
                    <div className="flex items-start justify-between">
                      <div>
                        <div className="flex items-center gap-2">
                          <span className="text-sm font-mono text-gray-500">{gap.control_code}</span>
                          <span className={`badge badge-${gap.risk_level}`}>{gap.risk_level}</span>
                          {gap.priority && <span className="badge badge-info">{gap.priority}</span>}
                        </div>
                        <h3 className="font-medium text-gray-900 mt-1">{gap.control_title}</h3>
                      </div>
                    </div>
                    {gap.gap_description && (
                      <p className="text-sm text-gray-600 mt-2">{gap.gap_description}</p>
                    )}
                    {gap.recommendation && (
                      <div className="mt-2 rounded-lg bg-blue-50 p-3">
                        <p className="text-xs font-medium text-blue-700">Recommendation</p>
                        <p className="text-sm text-blue-800 mt-0.5">{gap.recommendation}</p>
                      </div>
                    )}
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      )}

      {/* Cross-Mapping Tab */}
      {tab === 'Cross-Mapping' && (
        <div>
          {crossMappingLoading ? (
            <div className="flex items-center justify-center h-32">
              <p className="text-gray-500">Loading cross-mappings...</p>
            </div>
          ) : crossMappings.length === 0 ? (
            <div className="rounded-lg bg-gray-50 border border-gray-200 p-8 text-center">
              <p className="text-gray-500">No cross-mappings found for this framework.</p>
            </div>
          ) : (
            <div>
              <p className="text-sm text-gray-500 mb-4">{crossMappings.length} mapping{crossMappings.length !== 1 ? 's' : ''} to other frameworks</p>
              <div className="card overflow-hidden p-0">
                <table className="w-full">
                  <thead>
                    <tr className="border-b border-gray-100">
                      <th className="table-header px-4 py-3">Source Framework</th>
                      <th className="table-header px-4 py-3">Source Control</th>
                      <th className="table-header px-4 py-3">Mapping</th>
                      <th className="table-header px-4 py-3">Target Framework</th>
                      <th className="table-header px-4 py-3">Target Control</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-gray-50">
                    {crossMappings.map((m, idx) => (
                      <tr key={idx} className="hover:bg-gray-50">
                        <td className="px-4 py-3 text-sm font-medium text-gray-700">{m.source_framework}</td>
                        <td className="px-4 py-3">
                          <span className="text-sm font-mono text-gray-500">{m.source_control_code}</span>
                          <p className="text-xs text-gray-400 mt-0.5">{m.source_control_title}</p>
                        </td>
                        <td className="px-4 py-3 text-center">
                          <span className="badge badge-info">{m.mapping_type || 'equivalent'}</span>
                        </td>
                        <td className="px-4 py-3 text-sm font-medium text-gray-700">{m.target_framework}</td>
                        <td className="px-4 py-3">
                          <span className="text-sm font-mono text-gray-500">{m.target_control_code}</span>
                          <p className="text-xs text-gray-400 mt-0.5">{m.target_control_title}</p>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
