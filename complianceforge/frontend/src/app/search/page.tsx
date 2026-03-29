'use client';

import { useState, useEffect, useCallback } from 'react';
import api from '@/lib/api';

// ============================================================
// SEARCH PAGE — /search
// Full search results page with faceted filtering, search bar,
// result list with highlighted snippets, and pagination.
// ============================================================

interface SearchResult {
  entity_type: string;
  entity_id: string;
  entity_ref: string;
  title: string;
  snippet: string;
  score: number;
  status: string;
  severity: string;
  category: string;
  framework_codes: string[];
  tags: string[];
  updated_at: string;
}

interface FacetBucket {
  value: string;
  count: number;
}

interface SearchFacets {
  entity_types: FacetBucket[];
  frameworks: FacetBucket[];
  statuses: FacetBucket[];
  severities: FacetBucket[];
  categories: FacetBucket[];
}

interface SearchResponse {
  results: SearchResult[];
  facets: SearchFacets;
  total_results: number;
  query_time_ms: number;
  page: number;
  page_size: number;
  total_pages: number;
}

const entityTypeLabels: Record<string, string> = {
  risk: 'Risk',
  policy: 'Policy',
  control: 'Control',
  incident: 'Incident',
  vendor: 'Vendor',
  asset: 'Asset',
  finding: 'Finding',
  exception: 'Exception',
  dsr_request: 'DSR Request',
  regulatory_change: 'Regulatory Change',
  processing_activity: 'Processing Activity',
  remediation_action: 'Remediation Action',
};

const entityTypeColors: Record<string, string> = {
  risk: 'bg-red-100 text-red-800',
  policy: 'bg-blue-100 text-blue-800',
  control: 'bg-green-100 text-green-800',
  incident: 'bg-orange-100 text-orange-800',
  vendor: 'bg-purple-100 text-purple-800',
  asset: 'bg-cyan-100 text-cyan-800',
  finding: 'bg-yellow-100 text-yellow-800',
  exception: 'bg-pink-100 text-pink-800',
  dsr_request: 'bg-indigo-100 text-indigo-800',
  regulatory_change: 'bg-teal-100 text-teal-800',
};

const entityRoutes: Record<string, string> = {
  risk: '/risks',
  policy: '/policies',
  control: '/controls',
  incident: '/incidents',
  vendor: '/vendors',
  asset: '/assets',
  finding: '/audits',
  exception: '/exceptions',
  dsr_request: '/dsr',
  regulatory_change: '/regulatory',
  processing_activity: '/data/ropa',
  remediation_action: '/remediation',
};

export default function SearchPage() {
  const [query, setQuery] = useState('');
  const [response, setResponse] = useState<SearchResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [page, setPage] = useState(1);

  // Filters
  const [selectedTypes, setSelectedTypes] = useState<string[]>([]);
  const [selectedFrameworks, setSelectedFrameworks] = useState<string[]>([]);
  const [selectedStatuses, setSelectedStatuses] = useState<string[]>([]);
  const [selectedSeverities, setSelectedSeverities] = useState<string[]>([]);
  const [sortBy, setSortBy] = useState('relevance');

  const doSearch = useCallback(async () => {
    if (!query.trim()) return;
    setLoading(true);
    try {
      const params = new URLSearchParams({
        q: query,
        page: String(page),
        page_size: '20',
        sort: sortBy,
      });
      if (selectedTypes.length) params.set('types', selectedTypes.join(','));
      if (selectedFrameworks.length) params.set('frameworks', selectedFrameworks.join(','));
      if (selectedStatuses.length) params.set('statuses', selectedStatuses.join(','));
      if (selectedSeverities.length) params.set('severities', selectedSeverities.join(','));

      const res = await api.searchEntities(params.toString());
      setResponse(res.data);
    } catch (err) {
      console.error('Search failed:', err);
    } finally {
      setLoading(false);
    }
  }, [query, page, sortBy, selectedTypes, selectedFrameworks, selectedStatuses, selectedSeverities]);

  useEffect(() => {
    const timer = setTimeout(() => {
      if (query.trim()) doSearch();
    }, 300);
    return () => clearTimeout(timer);
  }, [query, doSearch]);

  useEffect(() => {
    if (query.trim()) doSearch();
  }, [page, sortBy, selectedTypes, selectedFrameworks, selectedStatuses, selectedSeverities, doSearch]);

  const toggleFilter = (list: string[], setList: (v: string[]) => void, value: string) => {
    setList(list.includes(value) ? list.filter(v => v !== value) : [...list, value]);
    setPage(1);
  };

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Search Header */}
      <div className="bg-white border-b sticky top-0 z-10">
        <div className="max-w-7xl mx-auto px-4 py-4">
          <div className="flex items-center gap-4">
            <div className="flex-1 relative">
              <input
                type="text"
                value={query}
                onChange={(e) => { setQuery(e.target.value); setPage(1); }}
                placeholder="Search across all entities... (supports &quot;exact phrases&quot;, OR, and -negation)"
                className="w-full px-4 py-3 pl-10 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-lg"
                autoFocus
              />
              <svg className="absolute left-3 top-3.5 w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
              </svg>
            </div>
            <select
              value={sortBy}
              onChange={(e) => setSortBy(e.target.value)}
              className="px-3 py-3 border rounded-lg bg-white"
            >
              <option value="relevance">Relevance</option>
              <option value="date">Newest</option>
              <option value="title">Title A-Z</option>
            </select>
          </div>
          {response && (
            <div className="mt-2 text-sm text-gray-500">
              {response.total_results} results in {response.query_time_ms}ms
            </div>
          )}
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-4 py-6 flex gap-6">
        {/* Faceted Filters Sidebar */}
        <div className="w-64 flex-shrink-0">
          <div className="bg-white rounded-lg shadow p-4 space-y-6">
            {/* Entity Type Filter */}
            {response?.facets?.entity_types && response.facets.entity_types.length > 0 && (
              <div>
                <h3 className="font-semibold text-sm text-gray-700 mb-2">Entity Type</h3>
                {response.facets.entity_types.map(f => (
                  <label key={f.value} className="flex items-center gap-2 py-1 cursor-pointer">
                    <input
                      type="checkbox"
                      checked={selectedTypes.includes(f.value)}
                      onChange={() => toggleFilter(selectedTypes, setSelectedTypes, f.value)}
                      className="rounded"
                    />
                    <span className="text-sm">{entityTypeLabels[f.value] || f.value}</span>
                    <span className="text-xs text-gray-400 ml-auto">{f.count}</span>
                  </label>
                ))}
              </div>
            )}

            {/* Framework Filter */}
            {response?.facets?.frameworks && response.facets.frameworks.length > 0 && (
              <div>
                <h3 className="font-semibold text-sm text-gray-700 mb-2">Framework</h3>
                {response.facets.frameworks.map(f => (
                  <label key={f.value} className="flex items-center gap-2 py-1 cursor-pointer">
                    <input
                      type="checkbox"
                      checked={selectedFrameworks.includes(f.value)}
                      onChange={() => toggleFilter(selectedFrameworks, setSelectedFrameworks, f.value)}
                      className="rounded"
                    />
                    <span className="text-sm">{f.value}</span>
                    <span className="text-xs text-gray-400 ml-auto">{f.count}</span>
                  </label>
                ))}
              </div>
            )}

            {/* Status Filter */}
            {response?.facets?.statuses && response.facets.statuses.length > 0 && (
              <div>
                <h3 className="font-semibold text-sm text-gray-700 mb-2">Status</h3>
                {response.facets.statuses.map(f => (
                  <label key={f.value} className="flex items-center gap-2 py-1 cursor-pointer">
                    <input
                      type="checkbox"
                      checked={selectedStatuses.includes(f.value)}
                      onChange={() => toggleFilter(selectedStatuses, setSelectedStatuses, f.value)}
                      className="rounded"
                    />
                    <span className="text-sm capitalize">{f.value}</span>
                    <span className="text-xs text-gray-400 ml-auto">{f.count}</span>
                  </label>
                ))}
              </div>
            )}

            {/* Severity Filter */}
            {response?.facets?.severities && response.facets.severities.length > 0 && (
              <div>
                <h3 className="font-semibold text-sm text-gray-700 mb-2">Severity</h3>
                {response.facets.severities.map(f => (
                  <label key={f.value} className="flex items-center gap-2 py-1 cursor-pointer">
                    <input
                      type="checkbox"
                      checked={selectedSeverities.includes(f.value)}
                      onChange={() => toggleFilter(selectedSeverities, setSelectedSeverities, f.value)}
                      className="rounded"
                    />
                    <span className="text-sm capitalize">{f.value}</span>
                    <span className="text-xs text-gray-400 ml-auto">{f.count}</span>
                  </label>
                ))}
              </div>
            )}
          </div>
        </div>

        {/* Results */}
        <div className="flex-1 space-y-4">
          {loading && (
            <div className="text-center py-12 text-gray-500">Searching...</div>
          )}

          {!loading && !response && (
            <div className="text-center py-12">
              <div className="text-6xl mb-4">&#128269;</div>
              <h2 className="text-xl font-semibold text-gray-700">Search ComplianceForge</h2>
              <p className="text-gray-500 mt-2">
                Search across risks, policies, controls, incidents, vendors, assets, findings, and more.
              </p>
              <p className="text-gray-400 mt-1 text-sm">
                Use &quot;quotes&quot; for exact phrases, OR for alternatives, and -minus to exclude terms.
              </p>
            </div>
          )}

          {!loading && response?.results?.map((result, i) => (
            <a
              key={`${result.entity_type}-${result.entity_id}-${i}`}
              href={`${entityRoutes[result.entity_type] || '#'}/${result.entity_id}`}
              className="block bg-white rounded-lg shadow p-4 hover:shadow-md transition"
            >
              <div className="flex items-start gap-3">
                <span className={`px-2 py-0.5 rounded text-xs font-medium ${entityTypeColors[result.entity_type] || 'bg-gray-100 text-gray-800'}`}>
                  {entityTypeLabels[result.entity_type] || result.entity_type}
                </span>
                {result.entity_ref && (
                  <span className="text-xs text-gray-400 font-mono">{result.entity_ref}</span>
                )}
                {result.status && (
                  <span className="text-xs px-2 py-0.5 rounded bg-gray-100 text-gray-600 capitalize ml-auto">
                    {result.status}
                  </span>
                )}
              </div>
              <h3 className="text-lg font-medium text-gray-900 mt-2">{result.title}</h3>
              {result.snippet && (
                <p className="text-sm text-gray-600 mt-1 line-clamp-2">{result.snippet}</p>
              )}
              <div className="flex items-center gap-3 mt-2 text-xs text-gray-400">
                {result.severity && <span className="capitalize">{result.severity}</span>}
                {result.framework_codes?.length > 0 && (
                  <span>{result.framework_codes.join(', ')}</span>
                )}
                <span className="ml-auto">{new Date(result.updated_at).toLocaleDateString()}</span>
              </div>
            </a>
          ))}

          {!loading && response && response.results?.length === 0 && (
            <div className="text-center py-12">
              <div className="text-4xl mb-4">&#128533;</div>
              <h3 className="text-lg font-semibold text-gray-700">No results found</h3>
              <p className="text-gray-500 mt-1">Try different keywords or remove some filters.</p>
            </div>
          )}

          {/* Pagination */}
          {response && response.total_pages > 1 && (
            <div className="flex justify-center gap-2 py-4">
              <button
                onClick={() => setPage(p => Math.max(1, p - 1))}
                disabled={page <= 1}
                className="px-3 py-1 rounded border disabled:opacity-50"
              >
                Previous
              </button>
              <span className="px-3 py-1 text-sm text-gray-600">
                Page {response.page} of {response.total_pages}
              </span>
              <button
                onClick={() => setPage(p => Math.min(response.total_pages, p + 1))}
                disabled={page >= response.total_pages}
                className="px-3 py-1 rounded border disabled:opacity-50"
              >
                Next
              </button>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
