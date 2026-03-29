'use client';

import { useEffect, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import api from '@/lib/api';

// ============================================================
// Package Detail Page
// Displays full package information with tabs for Overview,
// Contents, Versions, and Reviews. Includes install button.
// ============================================================

interface ContentsSummary {
  total_controls?: number;
  total_policies?: number;
  total_templates?: number;
  control_categories?: { name: string; count: number }[];
  framework_mappings?: number;
  evidence_templates?: number;
  test_procedures?: number;
}

interface MarketplacePackage {
  id: string;
  publisher_id: string;
  package_slug: string;
  name: string;
  description: string;
  long_description: string;
  package_type: string;
  category: string;
  applicable_frameworks: string[];
  applicable_regions: string[];
  applicable_industries: string[];
  tags: string[];
  current_version: string;
  min_platform_version: string;
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
  deprecated_at: string | null;
  deprecation_message: string;
  license: string;
  publisher_name: string;
  publisher_slug: string;
  is_verified: boolean;
}

interface PackageVersion {
  id: string;
  package_id: string;
  version: string;
  release_notes: string;
  package_hash: string;
  file_size_bytes: number;
  is_breaking_change: boolean;
  migration_notes: string;
  published_at: string;
}

interface PackageStats {
  total_downloads: number;
  total_installs: number;
  average_rating: number;
  total_reviews: number;
  version_count: number;
  latest_version: string;
  first_published: string;
}

interface Review {
  id: string;
  package_id: string;
  organization_id: string;
  user_id: string;
  rating: number;
  title: string;
  review_text: string;
  helpful_count: number;
  is_verified_install: boolean;
  created_at: string;
}

interface PackageDetail {
  package: MarketplacePackage;
  versions: PackageVersion[];
  stats: PackageStats;
}

const PACKAGE_TYPE_LABELS: Record<string, string> = {
  control_pack: 'Control Pack',
  framework_template: 'Framework Template',
  policy_bundle: 'Policy Bundle',
  risk_library: 'Risk Library',
  assessment_template: 'Assessment Template',
  report_template: 'Report Template',
};

const FRAMEWORK_LABELS: Record<string, string> = {
  ISO27001: 'ISO 27001',
  NIST_CSF_2: 'NIST CSF 2.0',
  NIST_800_53: 'NIST 800-53',
  UK_GDPR: 'UK GDPR',
  PCI_DSS_4: 'PCI DSS 4.0',
  NCSC_CAF: 'NCSC CAF',
  CYBER_ESSENTIALS: 'Cyber Essentials',
  COBIT_2019: 'COBIT 2019',
  ITIL_4: 'ITIL 4',
};

type TabID = 'overview' | 'contents' | 'versions' | 'reviews';

function StarRating({ rating, count, large }: { rating: number; count: number; large?: boolean }) {
  const size = large ? 'h-5 w-5' : 'h-4 w-4';
  const fullStars = Math.floor(rating);
  const halfStar = rating - fullStars >= 0.5;
  return (
    <span className="flex items-center gap-1">
      {[...Array(5)].map((_, i) => (
        <svg
          key={i}
          className={`${size} ${
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
      <span className={`text-gray-500 ml-1 ${large ? 'text-base' : 'text-sm'}`}>
        {rating > 0 ? rating.toFixed(1) : '--'} ({count} review{count !== 1 ? 's' : ''})
      </span>
    </span>
  );
}

function formatBytes(bytes: number): string {
  if (bytes < 1024) return bytes + ' B';
  if (bytes < 1048576) return (bytes / 1024).toFixed(1) + ' KB';
  return (bytes / 1048576).toFixed(1) + ' MB';
}

function formatDate(dateStr: string): string {
  if (!dateStr) return '--';
  return new Date(dateStr).toLocaleDateString('en-GB', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
  });
}

export default function PackageDetailPage() {
  const params = useParams();
  const router = useRouter();
  const publisherSlug = params.publisher as string;
  const packageSlug = params.slug as string;

  const [detail, setDetail] = useState<PackageDetail | null>(null);
  const [reviews, setReviews] = useState<Review[]>([]);
  const [reviewPage, setReviewPage] = useState(1);
  const [reviewTotal, setReviewTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [installing, setInstalling] = useState(false);
  const [installSuccess, setInstallSuccess] = useState(false);
  const [activeTab, setActiveTab] = useState<TabID>('overview');

  // Review form
  const [showReviewForm, setShowReviewForm] = useState(false);
  const [reviewRating, setReviewRating] = useState(5);
  const [reviewTitle, setReviewTitle] = useState('');
  const [reviewText, setReviewText] = useState('');
  const [submittingReview, setSubmittingReview] = useState(false);

  useEffect(() => {
    setLoading(true);
    api.getMarketplacePackageDetail(publisherSlug, packageSlug)
      .then(res => setDetail(res.data))
      .catch(() => setDetail(null))
      .finally(() => setLoading(false));
  }, [publisherSlug, packageSlug]);

  useEffect(() => {
    if (activeTab === 'reviews') {
      api.getMarketplacePackageReviews(publisherSlug, packageSlug, reviewPage, 10)
        .then(res => {
          setReviews(res.data || []);
          setReviewTotal(res.pagination?.total_items || 0);
        })
        .catch(() => setReviews([]));
    }
  }, [publisherSlug, packageSlug, activeTab, reviewPage]);

  const handleInstall = async () => {
    if (!detail) return;
    const latestVersion = detail.versions[0];
    if (!latestVersion) return;

    setInstalling(true);
    try {
      await api.installMarketplacePackage({
        package_id: detail.package.id,
        version_id: latestVersion.id,
        configuration: {},
      });
      setInstallSuccess(true);
      setTimeout(() => setInstallSuccess(false), 5000);
    } catch (err: any) {
      alert(err.message || 'Installation failed');
    } finally {
      setInstalling(false);
    }
  };

  const handleSubmitReview = async () => {
    if (!detail) return;
    setSubmittingReview(true);
    try {
      await api.submitMarketplaceReview({
        package_id: detail.package.id,
        rating: reviewRating,
        title: reviewTitle,
        review_text: reviewText,
      });
      setShowReviewForm(false);
      setReviewTitle('');
      setReviewText('');
      setReviewRating(5);
      // Reload reviews
      setReviewPage(1);
      setActiveTab('reviews');
    } catch (err: any) {
      alert(err.message || 'Failed to submit review');
    } finally {
      setSubmittingReview(false);
    }
  };

  if (loading) {
    return (
      <div>
        <div className="animate-pulse">
          <div className="h-6 bg-gray-200 rounded w-64 mb-2" />
          <div className="h-4 bg-gray-200 rounded w-96 mb-6" />
          <div className="card h-96" />
        </div>
      </div>
    );
  }

  if (!detail) {
    return (
      <div className="text-center py-16">
        <h2 className="text-xl font-bold text-gray-900">Package not found</h2>
        <p className="text-gray-500 mt-2">The package you are looking for does not exist or has been removed.</p>
        <a href="/marketplace" className="btn-primary mt-4 inline-block">
          Back to Marketplace
        </a>
      </div>
    );
  }

  const pkg = detail.package;
  const versions = detail.versions || [];
  const stats = detail.stats;

  const tabs: { id: TabID; label: string }[] = [
    { id: 'overview', label: 'Overview' },
    { id: 'contents', label: 'Contents' },
    { id: 'versions', label: `Versions (${versions.length})` },
    { id: 'reviews', label: `Reviews (${stats.total_reviews})` },
  ];

  return (
    <div>
      {/* Breadcrumb */}
      <nav className="flex items-center gap-2 text-sm text-gray-500 mb-6">
        <a href="/marketplace" className="hover:text-indigo-600">Marketplace</a>
        <span>/</span>
        <a href={`/marketplace?q=${pkg.publisher_slug}`} className="hover:text-indigo-600">{pkg.publisher_name}</a>
        <span>/</span>
        <span className="text-gray-900 font-medium">{pkg.name}</span>
      </nav>

      {/* Deprecation Banner */}
      {pkg.deprecated_at && (
        <div className="bg-amber-50 border border-amber-200 rounded-lg p-4 mb-6">
          <div className="flex items-start gap-3">
            <svg className="h-5 w-5 text-amber-500 mt-0.5 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
            </svg>
            <div>
              <p className="font-medium text-amber-800">This package has been deprecated</p>
              <p className="text-sm text-amber-700 mt-0.5">{pkg.deprecation_message || 'This package is no longer maintained.'}</p>
            </div>
          </div>
        </div>
      )}

      {/* Header */}
      <div className="flex flex-col lg:flex-row gap-6 mb-8">
        <div className="flex-1">
          <div className="flex items-center gap-3 mb-2">
            <h1 className="text-2xl font-bold text-gray-900">{pkg.name}</h1>
            {pkg.featured && (
              <span className="text-xs bg-indigo-100 text-indigo-700 px-2 py-0.5 rounded-full font-medium">
                Featured
              </span>
            )}
          </div>
          <p className="text-gray-500 mb-4">{pkg.description}</p>

          {/* Publisher */}
          <div className="flex items-center gap-2 text-sm text-gray-600 mb-3">
            <span>By</span>
            <span className="font-medium flex items-center gap-1">
              {pkg.is_verified && (
                <svg className="h-4 w-4 text-blue-500" fill="currentColor" viewBox="0 0 20 20">
                  <path fillRule="evenodd" d="M6.267 3.455a3.066 3.066 0 001.745-.723 3.066 3.066 0 013.976 0 3.066 3.066 0 001.745.723 3.066 3.066 0 012.812 2.812c.051.643.304 1.254.723 1.745a3.066 3.066 0 010 3.976 3.066 3.066 0 00-.723 1.745 3.066 3.066 0 01-2.812 2.812 3.066 3.066 0 00-1.745.723 3.066 3.066 0 01-3.976 0 3.066 3.066 0 00-1.745-.723 3.066 3.066 0 01-2.812-2.812 3.066 3.066 0 00-.723-1.745 3.066 3.066 0 010-3.976 3.066 3.066 0 00.723-1.745 3.066 3.066 0 012.812-2.812zm7.44 5.252a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                </svg>
              )}
              {pkg.publisher_name}
            </span>
          </div>

          <StarRating rating={pkg.rating_avg} count={pkg.rating_count} large />

          {/* Frameworks */}
          {pkg.applicable_frameworks && pkg.applicable_frameworks.length > 0 && (
            <div className="flex flex-wrap gap-2 mt-4">
              {pkg.applicable_frameworks.map(fw => (
                <span key={fw} className="text-xs bg-gray-100 text-gray-600 px-2 py-1 rounded-md">
                  {FRAMEWORK_LABELS[fw] || fw}
                </span>
              ))}
            </div>
          )}
        </div>

        {/* Install Card */}
        <div className="lg:w-72 flex-shrink-0">
          <div className="card border border-gray-200">
            <div className="text-center mb-4">
              {pkg.pricing_model === 'free' ? (
                <span className="text-2xl font-bold text-green-600">Free</span>
              ) : (
                <span className="text-2xl font-bold text-gray-900">EUR {pkg.price_eur.toFixed(2)}</span>
              )}
            </div>

            <button
              onClick={handleInstall}
              disabled={installing || installSuccess}
              className={`w-full py-2.5 px-4 rounded-lg font-medium text-sm transition-colors ${
                installSuccess
                  ? 'bg-green-600 text-white'
                  : 'bg-indigo-600 text-white hover:bg-indigo-700 disabled:opacity-50'
              }`}
            >
              {installing ? 'Installing...' : installSuccess ? 'Installed' : 'Install Package'}
            </button>

            <div className="mt-4 space-y-3 text-sm">
              <div className="flex justify-between text-gray-600">
                <span>Version</span>
                <span className="font-medium text-gray-900">{pkg.current_version}</span>
              </div>
              <div className="flex justify-between text-gray-600">
                <span>Type</span>
                <span className="font-medium text-gray-900">{PACKAGE_TYPE_LABELS[pkg.package_type] || pkg.package_type}</span>
              </div>
              <div className="flex justify-between text-gray-600">
                <span>License</span>
                <span className="font-medium text-gray-900">{pkg.license}</span>
              </div>
              <div className="flex justify-between text-gray-600">
                <span>Downloads</span>
                <span className="font-medium text-gray-900">{pkg.download_count.toLocaleString()}</span>
              </div>
              <div className="flex justify-between text-gray-600">
                <span>Active Installs</span>
                <span className="font-medium text-gray-900">{pkg.install_count.toLocaleString()}</span>
              </div>
              {pkg.published_at && (
                <div className="flex justify-between text-gray-600">
                  <span>Published</span>
                  <span className="font-medium text-gray-900">{formatDate(pkg.published_at)}</span>
                </div>
              )}
              {pkg.min_platform_version && (
                <div className="flex justify-between text-gray-600">
                  <span>Min Platform</span>
                  <span className="font-medium text-gray-900">v{pkg.min_platform_version}</span>
                </div>
              )}
            </div>

            {/* Regions */}
            {pkg.applicable_regions && pkg.applicable_regions.length > 0 && (
              <div className="mt-4 pt-4 border-t border-gray-100">
                <p className="text-xs text-gray-500 mb-2">Applicable Regions</p>
                <div className="flex flex-wrap gap-1">
                  {pkg.applicable_regions.map(reg => (
                    <span key={reg} className="text-xs bg-blue-50 text-blue-600 px-2 py-0.5 rounded">
                      {reg}
                    </span>
                  ))}
                </div>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Tabs */}
      <div className="border-b border-gray-200 mb-6">
        <nav className="flex gap-6">
          {tabs.map(tab => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={`pb-3 text-sm font-medium border-b-2 transition-colors ${
                activeTab === tab.id
                  ? 'border-indigo-600 text-indigo-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700'
              }`}
            >
              {tab.label}
            </button>
          ))}
        </nav>
      </div>

      {/* Tab Content */}
      <div className="min-h-[400px]">
        {/* Overview Tab */}
        {activeTab === 'overview' && (
          <div className="prose prose-gray max-w-none">
            {pkg.long_description ? (
              <div
                className="text-sm text-gray-700 leading-relaxed"
                style={{ whiteSpace: 'pre-wrap' }}
              >
                {pkg.long_description}
              </div>
            ) : (
              <p className="text-gray-500">{pkg.description}</p>
            )}

            {/* Tags */}
            {pkg.tags && pkg.tags.length > 0 && (
              <div className="mt-8">
                <h3 className="text-sm font-semibold text-gray-900 mb-3">Tags</h3>
                <div className="flex flex-wrap gap-2">
                  {pkg.tags.map(tag => (
                    <a
                      key={tag}
                      href={`/marketplace?tag=${tag}`}
                      className="text-sm bg-gray-100 text-gray-600 px-3 py-1 rounded-full hover:bg-gray-200 transition-colors"
                    >
                      {tag}
                    </a>
                  ))}
                </div>
              </div>
            )}

            {/* Industries */}
            {pkg.applicable_industries && pkg.applicable_industries.length > 0 && (
              <div className="mt-6">
                <h3 className="text-sm font-semibold text-gray-900 mb-3">Applicable Industries</h3>
                <div className="flex flex-wrap gap-2">
                  {pkg.applicable_industries.map(ind => (
                    <span key={ind} className="text-sm bg-gray-100 text-gray-600 px-3 py-1 rounded-full capitalize">
                      {ind.replace(/_/g, ' ')}
                    </span>
                  ))}
                </div>
              </div>
            )}
          </div>
        )}

        {/* Contents Tab */}
        {activeTab === 'contents' && (
          <div>
            {pkg.contents_summary ? (
              <div className="space-y-6">
                {/* Summary Stats */}
                <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                  {pkg.contents_summary.total_controls != null && (
                    <div className="card text-center">
                      <p className="text-3xl font-bold text-indigo-600">{pkg.contents_summary.total_controls}</p>
                      <p className="text-sm text-gray-500 mt-1">Controls</p>
                    </div>
                  )}
                  {pkg.contents_summary.framework_mappings != null && (
                    <div className="card text-center">
                      <p className="text-3xl font-bold text-emerald-600">{pkg.contents_summary.framework_mappings}</p>
                      <p className="text-sm text-gray-500 mt-1">Framework Mappings</p>
                    </div>
                  )}
                  {pkg.contents_summary.evidence_templates != null && (
                    <div className="card text-center">
                      <p className="text-3xl font-bold text-amber-600">{pkg.contents_summary.evidence_templates}</p>
                      <p className="text-sm text-gray-500 mt-1">Evidence Templates</p>
                    </div>
                  )}
                  {pkg.contents_summary.test_procedures != null && (
                    <div className="card text-center">
                      <p className="text-3xl font-bold text-purple-600">{pkg.contents_summary.test_procedures}</p>
                      <p className="text-sm text-gray-500 mt-1">Test Procedures</p>
                    </div>
                  )}
                </div>

                {/* Control Categories */}
                {pkg.contents_summary.control_categories && pkg.contents_summary.control_categories.length > 0 && (
                  <div className="card">
                    <h3 className="font-semibold text-gray-900 mb-4">Control Categories</h3>
                    <div className="space-y-3">
                      {pkg.contents_summary.control_categories.map(cat => {
                        const total = pkg.contents_summary.total_controls || 1;
                        const pct = Math.round((cat.count / total) * 100);
                        return (
                          <div key={cat.name}>
                            <div className="flex items-center justify-between text-sm mb-1">
                              <span className="text-gray-700">{cat.name}</span>
                              <span className="text-gray-500">{cat.count} controls</span>
                            </div>
                            <div className="w-full bg-gray-100 rounded-full h-2">
                              <div
                                className="bg-indigo-500 h-2 rounded-full transition-all"
                                style={{ width: `${pct}%` }}
                              />
                            </div>
                          </div>
                        );
                      })}
                    </div>
                  </div>
                )}
              </div>
            ) : (
              <p className="text-gray-500">No detailed contents information available.</p>
            )}
          </div>
        )}

        {/* Versions Tab */}
        {activeTab === 'versions' && (
          <div className="space-y-4">
            {versions.length === 0 && (
              <p className="text-gray-500">No versions published yet.</p>
            )}
            {versions.map((ver, idx) => (
              <div key={ver.id} className={`card ${idx === 0 ? 'border-2 border-indigo-100' : ''}`}>
                <div className="flex items-start justify-between mb-2">
                  <div className="flex items-center gap-2">
                    <h3 className="font-semibold text-gray-900">v{ver.version}</h3>
                    {idx === 0 && (
                      <span className="text-xs bg-indigo-100 text-indigo-700 px-2 py-0.5 rounded-full">Latest</span>
                    )}
                    {ver.is_breaking_change && (
                      <span className="text-xs bg-red-100 text-red-700 px-2 py-0.5 rounded-full">Breaking Change</span>
                    )}
                  </div>
                  <span className="text-xs text-gray-400">{formatDate(ver.published_at)}</span>
                </div>
                {ver.release_notes && (
                  <p className="text-sm text-gray-600 mb-2">{ver.release_notes}</p>
                )}
                {ver.migration_notes && (
                  <div className="text-sm text-amber-700 bg-amber-50 rounded p-2 mb-2">
                    <strong>Migration Notes:</strong> {ver.migration_notes}
                  </div>
                )}
                <div className="flex gap-4 text-xs text-gray-400">
                  <span>Size: {formatBytes(ver.file_size_bytes)}</span>
                  <span className="font-mono truncate max-w-xs" title={ver.package_hash}>
                    SHA-256: {ver.package_hash.substring(0, 16)}...
                  </span>
                </div>
              </div>
            ))}
          </div>
        )}

        {/* Reviews Tab */}
        {activeTab === 'reviews' && (
          <div>
            <div className="flex items-center justify-between mb-6">
              <h3 className="font-semibold text-gray-900">Reviews ({reviewTotal})</h3>
              <button
                onClick={() => setShowReviewForm(!showReviewForm)}
                className="btn-primary text-sm"
              >
                Write a Review
              </button>
            </div>

            {/* Review Form */}
            {showReviewForm && (
              <div className="card border border-indigo-100 mb-6">
                <h4 className="font-medium text-gray-900 mb-4">Your Review</h4>
                <div className="mb-4">
                  <label className="block text-sm font-medium text-gray-700 mb-1">Rating</label>
                  <div className="flex gap-1">
                    {[1, 2, 3, 4, 5].map(star => (
                      <button
                        key={star}
                        onClick={() => setReviewRating(star)}
                        className="focus:outline-none"
                      >
                        <svg
                          className={`h-8 w-8 ${star <= reviewRating ? 'text-amber-400' : 'text-gray-200'} hover:text-amber-300 transition-colors`}
                          fill="currentColor"
                          viewBox="0 0 20 20"
                        >
                          <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
                        </svg>
                      </button>
                    ))}
                  </div>
                </div>
                <div className="mb-4">
                  <label className="block text-sm font-medium text-gray-700 mb-1">Title</label>
                  <input
                    type="text"
                    value={reviewTitle}
                    onChange={e => setReviewTitle(e.target.value)}
                    placeholder="Summarise your experience..."
                    className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:ring-indigo-500 focus:border-indigo-500"
                  />
                </div>
                <div className="mb-4">
                  <label className="block text-sm font-medium text-gray-700 mb-1">Review</label>
                  <textarea
                    value={reviewText}
                    onChange={e => setReviewText(e.target.value)}
                    rows={4}
                    placeholder="Share details of your experience with this package..."
                    className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:ring-indigo-500 focus:border-indigo-500"
                  />
                </div>
                <div className="flex justify-end gap-3">
                  <button onClick={() => setShowReviewForm(false)} className="btn-secondary text-sm">Cancel</button>
                  <button
                    onClick={handleSubmitReview}
                    disabled={submittingReview || !reviewTitle.trim()}
                    className="btn-primary text-sm disabled:opacity-50"
                  >
                    {submittingReview ? 'Submitting...' : 'Submit Review'}
                  </button>
                </div>
              </div>
            )}

            {/* Review List */}
            {reviews.length === 0 && !showReviewForm && (
              <div className="text-center py-12">
                <p className="text-gray-500">No reviews yet. Be the first to review this package.</p>
              </div>
            )}
            <div className="space-y-4">
              {reviews.map(review => (
                <div key={review.id} className="card">
                  <div className="flex items-start justify-between mb-2">
                    <div>
                      <div className="flex items-center gap-2 mb-1">
                        <StarRating rating={review.rating} count={0} />
                        {review.is_verified_install && (
                          <span className="text-xs bg-green-100 text-green-700 px-2 py-0.5 rounded-full">
                            Verified Install
                          </span>
                        )}
                      </div>
                      <h4 className="font-medium text-gray-900">{review.title}</h4>
                    </div>
                    <span className="text-xs text-gray-400">{formatDate(review.created_at)}</span>
                  </div>
                  {review.review_text && (
                    <p className="text-sm text-gray-600 mt-2">{review.review_text}</p>
                  )}
                  {review.helpful_count > 0 && (
                    <p className="text-xs text-gray-400 mt-2">
                      {review.helpful_count} {review.helpful_count === 1 ? 'person' : 'people'} found this helpful
                    </p>
                  )}
                </div>
              ))}
            </div>

            {/* Review Pagination */}
            {reviewTotal > 10 && (
              <div className="flex items-center justify-center gap-4 mt-6">
                <button
                  onClick={() => setReviewPage(p => Math.max(1, p - 1))}
                  disabled={reviewPage <= 1}
                  className="btn-secondary text-sm disabled:opacity-50"
                >
                  Previous
                </button>
                <span className="text-sm text-gray-500">Page {reviewPage}</span>
                <button
                  onClick={() => setReviewPage(p => p + 1)}
                  disabled={reviews.length < 10}
                  className="btn-secondary text-sm disabled:opacity-50"
                >
                  Next
                </button>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
