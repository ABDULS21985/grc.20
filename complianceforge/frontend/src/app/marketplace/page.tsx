'use client';

import { useEffect, useState, useCallback } from 'react';
import api from '@/lib/api';

// ============================================================
// Marketplace Browse Page
// Search, filter, and discover compliance control packs,
// framework templates, and policy bundles.
// ============================================================

interface ContentsSummary {
  total_controls?: number;
  total_policies?: number;
  total_templates?: number;
  control_categories?: { name: string; count: number }[];
}

interface MarketplacePackage {
  id: string;
  publisher_id: string;
  package_slug: string;
  name: string;
  description: string;
  package_type: string;
  category: string;
  applicable_frameworks: string[];
  applicable_regions: string[];
  applicable_industries: string[];
  tags: string[];
  current_version: string;
  pricing_model: string;
  price_eur: number;
  download_count: number;
  install_count: number;
  rating_avg: number;
  rating_count: number;
  featured: boolean;
  contents_summary: ContentsSummary;
  status: string;
  published_at: string;
  license: string;
  publisher_name: string;
  publisher_slug: string;
  is_verified: boolean;
}

interface Pagination {
  page: number;
  page_size: number;
  total_items: number;
  total_pages: number;
  has_next: boolean;
  has_prev: boolean;
}

const PACKAGE_TYPE_LABELS: Record<string, string> = {
  control_pack: 'Control Pack',
  framework_template: 'Framework Template',
  policy_bundle: 'Policy Bundle',
  risk_library: 'Risk Library',
  assessment_template: 'Assessment Template',
  report_template: 'Report Template',
};

const PACKAGE_TYPE_COLORS: Record<string, string> = {
  control_pack: 'bg-indigo-100 text-indigo-700',
  framework_template: 'bg-emerald-100 text-emerald-700',
  policy_bundle: 'bg-amber-100 text-amber-700',
  risk_library: 'bg-red-100 text-red-700',
  assessment_template: 'bg-purple-100 text-purple-700',
  report_template: 'bg-sky-100 text-sky-700',
};

const CATEGORY_OPTIONS = [
  { value: '', label: 'All Categories' },
  { value: 'financial_services', label: 'Financial Services' },
  { value: 'healthcare', label: 'Healthcare' },
  { value: 'technology', label: 'Technology' },
  { value: 'government', label: 'Government' },
  { value: 'energy', label: 'Energy & Utilities' },
  { value: 'retail', label: 'Retail' },
  { value: 'manufacturing', label: 'Manufacturing' },
];

const SORT_OPTIONS = [
  { value: 'downloads', label: 'Most Downloads' },
  { value: 'rating', label: 'Highest Rated' },
  { value: 'newest', label: 'Newest First' },
  { value: 'relevance', label: 'Relevance' },
];

const FRAMEWORK_OPTIONS = [
  { value: 'ISO27001', label: 'ISO 27001' },
  { value: 'NIST_CSF_2', label: 'NIST CSF 2.0' },
  { value: 'NIST_800_53', label: 'NIST 800-53' },
  { value: 'UK_GDPR', label: 'UK GDPR' },
  { value: 'PCI_DSS_4', label: 'PCI DSS 4.0' },
  { value: 'NCSC_CAF', label: 'NCSC CAF' },
  { value: 'CYBER_ESSENTIALS', label: 'Cyber Essentials' },
];

const REGION_OPTIONS = [
  { value: 'UK', label: 'United Kingdom' },
  { value: 'EU', label: 'European Union' },
  { value: 'US', label: 'United States' },
  { value: 'Global', label: 'Global' },
];

function StarRating({ rating, count }: { rating: number; count: number }) {
  const fullStars = Math.floor(rating);
  const halfStar = rating - fullStars >= 0.5;
  return (
    <span className="flex items-center gap-1">
      {[...Array(5)].map((_, i) => (
        <svg
          key={i}
          className={`h-4 w-4 ${
            i < fullStars
              ? 'text-amber-400'
              : i === fullStars && halfStar
                ? 'text-amber-300'
                : 'text-gray-200'
          }`}
          fill="currentColor"
          viewBox="0 0 20 20"
        >
          <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
        </svg>
      ))}
      <span className="text-sm text-gray-500 ml-1">
        {rating > 0 ? rating.toFixed(1) : '--'} ({count})
      </span>
    </span>
  );
}

export default function MarketplacePage() {
  const [packages, setPackages] = useState<MarketplacePackage[]>([]);
  const [featured, setFeatured] = useState<MarketplacePackage[]>([]);
  const [pagination, setPagination] = useState<Pagination | null>(null);
  const [loading, setLoading] = useState(true);

  // Filters
  const [query, setQuery] = useState('');
  const [category, setCategory] = useState('');
  const [sortBy, setSortBy] = useState('downloads');
  const [selectedFrameworks, setSelectedFrameworks] = useState<string[]>([]);
  const [selectedRegions, setSelectedRegions] = useState<string[]>([]);
  const [pricingFilter, setPricingFilter] = useState('');
  const [page, setPage] = useState(1);

  const loadFeatured = useCallback(() => {
    api.getMarketplaceFeatured()
      .then(res => setFeatured(res.data || []))
      .catch(() => {});
  }, []);

  const loadPackages = useCallback(() => {
    setLoading(true);
    const params = new URLSearchParams();
    if (query) params.set('q', query);
    if (category) params.set('category', category);
    if (sortBy) params.set('sort', sortBy);
    if (pricingFilter) params.set('pricing', pricingFilter);
    if (selectedFrameworks.length > 0) params.set('framework', selectedFrameworks.join(','));
    if (selectedRegions.length > 0) params.set('region', selectedRegions.join(','));
    params.set('page', String(page));
    params.set('page_size', '12');

    api.getMarketplacePackages(params.toString())
      .then(res => {
        setPackages(res.data || []);
        setPagination(res.pagination || null);
      })
      .catch(() => setPackages([]))
      .finally(() => setLoading(false));
  }, [query, category, sortBy, pricingFilter, selectedFrameworks, selectedRegions, page]);

  useEffect(() => { loadFeatured(); }, [loadFeatured]);
  useEffect(() => { loadPackages(); }, [loadPackages]);

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    setPage(1);
    loadPackages();
  };

  const toggleFramework = (fw: string) => {
    setSelectedFrameworks(prev =>
      prev.includes(fw) ? prev.filter(f => f !== fw) : [...prev, fw]
    );
    setPage(1);
  };

  const toggleRegion = (reg: string) => {
    setSelectedRegions(prev =>
      prev.includes(reg) ? prev.filter(r => r !== reg) : [...prev, reg]
    );
    setPage(1);
  };

  const clearFilters = () => {
    setQuery('');
    setCategory('');
    setSortBy('downloads');
    setSelectedFrameworks([]);
    setSelectedRegions([]);
    setPricingFilter('');
    setPage(1);
  };

  const hasActiveFilters = query || category || selectedFrameworks.length > 0 || selectedRegions.length > 0 || pricingFilter;

  return (
    <div>
      {/* Header */}
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-gray-900">Marketplace</h1>
        <p className="text-gray-500 mt-1">
          Discover and install compliance control packs, framework templates, and policy bundles
        </p>
      </div>

      {/* Featured Packages */}
      {featured.length > 0 && !hasActiveFilters && (
        <div className="mb-10">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Featured Packages</h2>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            {featured.slice(0, 3).map(pkg => (
              <a
                key={pkg.id}
                href={`/marketplace/${pkg.publisher_slug}/${pkg.package_slug}`}
                className="card border-2 border-indigo-100 hover:border-indigo-300 transition-colors"
              >
                <div className="flex items-start justify-between mb-3">
                  <span className={`text-xs font-medium px-2 py-0.5 rounded-full ${PACKAGE_TYPE_COLORS[pkg.package_type] || 'bg-gray-100 text-gray-600'}`}>
                    {PACKAGE_TYPE_LABELS[pkg.package_type] || pkg.package_type}
                  </span>
                  <span className="text-xs text-indigo-600 font-semibold">Featured</span>
                </div>
                <h3 className="font-semibold text-gray-900 mb-1">{pkg.name}</h3>
                <p className="text-sm text-gray-500 mb-3 line-clamp-2">{pkg.description}</p>
                <div className="flex items-center justify-between text-xs text-gray-400">
                  <span className="flex items-center gap-1">
                    {pkg.is_verified && (
                      <svg className="h-3.5 w-3.5 text-blue-500" fill="currentColor" viewBox="0 0 20 20">
                        <path fillRule="evenodd" d="M6.267 3.455a3.066 3.066 0 001.745-.723 3.066 3.066 0 013.976 0 3.066 3.066 0 001.745.723 3.066 3.066 0 012.812 2.812c.051.643.304 1.254.723 1.745a3.066 3.066 0 010 3.976 3.066 3.066 0 00-.723 1.745 3.066 3.066 0 01-2.812 2.812 3.066 3.066 0 00-1.745.723 3.066 3.066 0 01-3.976 0 3.066 3.066 0 00-1.745-.723 3.066 3.066 0 01-2.812-2.812 3.066 3.066 0 00-.723-1.745 3.066 3.066 0 010-3.976 3.066 3.066 0 00.723-1.745 3.066 3.066 0 012.812-2.812zm7.44 5.252a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                      </svg>
                    )}
                    {pkg.publisher_name}
                  </span>
                  <span>v{pkg.current_version}</span>
                </div>
                <div className="flex items-center justify-between mt-3 pt-3 border-t border-gray-100">
                  <StarRating rating={pkg.rating_avg} count={pkg.rating_count} />
                  <span className="text-xs text-gray-400">{pkg.download_count.toLocaleString()} downloads</span>
                </div>
              </a>
            ))}
          </div>
        </div>
      )}

      {/* Search & Filters */}
      <div className="flex flex-col lg:flex-row gap-6">
        {/* Sidebar Filters */}
        <aside className="w-full lg:w-64 flex-shrink-0">
          <div className="card space-y-5">
            <div className="flex items-center justify-between">
              <h3 className="font-semibold text-gray-900 text-sm">Filters</h3>
              {hasActiveFilters && (
                <button onClick={clearFilters} className="text-xs text-indigo-600 hover:text-indigo-800">
                  Clear all
                </button>
              )}
            </div>

            {/* Category */}
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1.5">Category</label>
              <select
                value={category}
                onChange={e => { setCategory(e.target.value); setPage(1); }}
                className="w-full border border-gray-300 rounded-md px-2.5 py-1.5 text-sm focus:ring-indigo-500 focus:border-indigo-500"
              >
                {CATEGORY_OPTIONS.map(opt => (
                  <option key={opt.value} value={opt.value}>{opt.label}</option>
                ))}
              </select>
            </div>

            {/* Pricing */}
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1.5">Pricing</label>
              <select
                value={pricingFilter}
                onChange={e => { setPricingFilter(e.target.value); setPage(1); }}
                className="w-full border border-gray-300 rounded-md px-2.5 py-1.5 text-sm focus:ring-indigo-500 focus:border-indigo-500"
              >
                <option value="">All</option>
                <option value="free">Free</option>
                <option value="one_time">One-Time Purchase</option>
                <option value="subscription">Subscription</option>
              </select>
            </div>

            {/* Frameworks */}
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1.5">Frameworks</label>
              <div className="space-y-1.5">
                {FRAMEWORK_OPTIONS.map(fw => (
                  <label key={fw.value} className="flex items-center gap-2 text-sm text-gray-600 cursor-pointer">
                    <input
                      type="checkbox"
                      checked={selectedFrameworks.includes(fw.value)}
                      onChange={() => toggleFramework(fw.value)}
                      className="rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
                    />
                    {fw.label}
                  </label>
                ))}
              </div>
            </div>

            {/* Regions */}
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1.5">Regions</label>
              <div className="space-y-1.5">
                {REGION_OPTIONS.map(reg => (
                  <label key={reg.value} className="flex items-center gap-2 text-sm text-gray-600 cursor-pointer">
                    <input
                      type="checkbox"
                      checked={selectedRegions.includes(reg.value)}
                      onChange={() => toggleRegion(reg.value)}
                      className="rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
                    />
                    {reg.label}
                  </label>
                ))}
              </div>
            </div>
          </div>
        </aside>

        {/* Main Content */}
        <div className="flex-1">
          {/* Search Bar + Sort */}
          <div className="flex flex-col sm:flex-row gap-3 mb-6">
            <form onSubmit={handleSearch} className="flex-1">
              <div className="relative">
                <input
                  type="text"
                  value={query}
                  onChange={e => setQuery(e.target.value)}
                  placeholder="Search packages..."
                  className="w-full border border-gray-300 rounded-md pl-10 pr-4 py-2 text-sm focus:ring-indigo-500 focus:border-indigo-500"
                />
                <svg
                  className="absolute left-3 top-2.5 h-4 w-4 text-gray-400"
                  fill="none" stroke="currentColor" viewBox="0 0 24 24"
                >
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                </svg>
              </div>
            </form>
            <select
              value={sortBy}
              onChange={e => { setSortBy(e.target.value); setPage(1); }}
              className="border border-gray-300 rounded-md px-3 py-2 text-sm focus:ring-indigo-500 focus:border-indigo-500"
            >
              {SORT_OPTIONS.map(opt => (
                <option key={opt.value} value={opt.value}>{opt.label}</option>
              ))}
            </select>
          </div>

          {/* Results count */}
          {pagination && (
            <p className="text-sm text-gray-500 mb-4">
              {pagination.total_items} package{pagination.total_items !== 1 ? 's' : ''} found
            </p>
          )}

          {/* Loading */}
          {loading && (
            <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
              {Array.from({ length: 6 }).map((_, i) => (
                <div key={i} className="card animate-pulse">
                  <div className="h-4 bg-gray-200 rounded w-24 mb-3" />
                  <div className="h-5 bg-gray-200 rounded w-3/4 mb-2" />
                  <div className="h-3 bg-gray-200 rounded w-full mb-1" />
                  <div className="h-3 bg-gray-200 rounded w-2/3 mb-4" />
                  <div className="h-3 bg-gray-200 rounded w-1/3" />
                </div>
              ))}
            </div>
          )}

          {/* Package Cards */}
          {!loading && (
            <>
              <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
                {packages.map(pkg => (
                  <a
                    key={pkg.id}
                    href={`/marketplace/${pkg.publisher_slug}/${pkg.package_slug}`}
                    className="card hover:shadow-md transition-shadow"
                  >
                    <div className="flex items-start justify-between mb-2">
                      <span className={`text-xs font-medium px-2 py-0.5 rounded-full ${PACKAGE_TYPE_COLORS[pkg.package_type] || 'bg-gray-100 text-gray-600'}`}>
                        {PACKAGE_TYPE_LABELS[pkg.package_type] || pkg.package_type}
                      </span>
                      {pkg.pricing_model === 'free' ? (
                        <span className="text-xs font-medium text-green-600">Free</span>
                      ) : (
                        <span className="text-xs font-medium text-gray-600">
                          EUR {pkg.price_eur.toFixed(2)}
                        </span>
                      )}
                    </div>
                    <h3 className="font-semibold text-gray-900 text-sm mb-1 line-clamp-1">{pkg.name}</h3>
                    <p className="text-xs text-gray-500 mb-3 line-clamp-2">{pkg.description}</p>

                    {/* Contents summary */}
                    {pkg.contents_summary && (
                      <div className="flex gap-3 mb-3 text-xs text-gray-400">
                        {pkg.contents_summary.total_controls != null && (
                          <span>{pkg.contents_summary.total_controls} controls</span>
                        )}
                        {pkg.contents_summary.total_policies != null && (
                          <span>{pkg.contents_summary.total_policies} policies</span>
                        )}
                      </div>
                    )}

                    {/* Tags */}
                    {pkg.tags && pkg.tags.length > 0 && (
                      <div className="flex flex-wrap gap-1 mb-3">
                        {pkg.tags.slice(0, 3).map(tag => (
                          <span key={tag} className="text-xs bg-gray-100 text-gray-500 px-1.5 py-0.5 rounded">
                            {tag}
                          </span>
                        ))}
                        {pkg.tags.length > 3 && (
                          <span className="text-xs text-gray-400">+{pkg.tags.length - 3}</span>
                        )}
                      </div>
                    )}

                    {/* Publisher & Version */}
                    <div className="flex items-center justify-between text-xs text-gray-400 mb-2">
                      <span className="flex items-center gap-1">
                        {pkg.is_verified && (
                          <svg className="h-3 w-3 text-blue-500" fill="currentColor" viewBox="0 0 20 20">
                            <path fillRule="evenodd" d="M6.267 3.455a3.066 3.066 0 001.745-.723 3.066 3.066 0 013.976 0 3.066 3.066 0 001.745.723 3.066 3.066 0 012.812 2.812c.051.643.304 1.254.723 1.745a3.066 3.066 0 010 3.976 3.066 3.066 0 00-.723 1.745 3.066 3.066 0 01-2.812 2.812 3.066 3.066 0 00-1.745.723 3.066 3.066 0 01-3.976 0 3.066 3.066 0 00-1.745-.723 3.066 3.066 0 01-2.812-2.812 3.066 3.066 0 00-.723-1.745 3.066 3.066 0 010-3.976 3.066 3.066 0 00.723-1.745 3.066 3.066 0 012.812-2.812zm7.44 5.252a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                          </svg>
                        )}
                        {pkg.publisher_name}
                      </span>
                      <span>v{pkg.current_version}</span>
                    </div>

                    {/* Rating & Downloads */}
                    <div className="flex items-center justify-between pt-2 border-t border-gray-100">
                      <StarRating rating={pkg.rating_avg} count={pkg.rating_count} />
                      <span className="text-xs text-gray-400">
                        {pkg.download_count.toLocaleString()} dl
                      </span>
                    </div>
                  </a>
                ))}
              </div>

              {/* Empty State */}
              {packages.length === 0 && (
                <div className="text-center py-16">
                  <svg className="mx-auto h-12 w-12 text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
                  </svg>
                  <h3 className="mt-4 text-gray-900 font-medium">No packages found</h3>
                  <p className="text-sm text-gray-500 mt-1">Try adjusting your search or filters.</p>
                  <button onClick={clearFilters} className="mt-4 btn-secondary text-sm">
                    Clear Filters
                  </button>
                </div>
              )}

              {/* Pagination */}
              {pagination && pagination.total_pages > 1 && (
                <div className="flex items-center justify-between mt-8">
                  <button
                    onClick={() => setPage(p => Math.max(1, p - 1))}
                    disabled={!pagination.has_prev}
                    className="btn-secondary text-sm disabled:opacity-50"
                  >
                    Previous
                  </button>
                  <span className="text-sm text-gray-500">
                    Page {pagination.page} of {pagination.total_pages}
                  </span>
                  <button
                    onClick={() => setPage(p => p + 1)}
                    disabled={!pagination.has_next}
                    className="btn-secondary text-sm disabled:opacity-50"
                  >
                    Next
                  </button>
                </div>
              )}
            </>
          )}
        </div>
      </div>
    </div>
  );
}
