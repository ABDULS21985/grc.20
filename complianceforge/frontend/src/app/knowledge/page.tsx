'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';
import api from '@/lib/api';

// ============================================================
// KNOWLEDGE BASE — /knowledge
// Browse compliance guidance articles by category, framework,
// and difficulty. Search within the knowledge base. View
// recommended articles and bookmarks.
// ============================================================

interface KnowledgeArticle {
  id: string;
  article_type: string;
  title: string;
  slug: string;
  summary: string;
  applicable_frameworks: string[];
  applicable_control_codes: string[];
  tags: string[];
  difficulty: string;
  reading_time_minutes: number;
  is_system: boolean;
  is_bookmarked: boolean;
  view_count: number;
  helpful_count: number;
  created_at: string;
}

const articleTypeLabels: Record<string, string> = {
  implementation_guide: 'Implementation Guide',
  regulatory_guide: 'Regulatory Guide',
  best_practice: 'Best Practice',
  framework_overview: 'Framework Overview',
  control_guidance: 'Control Guidance',
  audit_preparation: 'Audit Preparation',
  incident_response: 'Incident Response',
  risk_management: 'Risk Management',
  vendor_management: 'Vendor Management',
};

const articleTypeColors: Record<string, string> = {
  implementation_guide: 'bg-blue-100 text-blue-800',
  regulatory_guide: 'bg-purple-100 text-purple-800',
  best_practice: 'bg-green-100 text-green-800',
  framework_overview: 'bg-cyan-100 text-cyan-800',
  control_guidance: 'bg-orange-100 text-orange-800',
  audit_preparation: 'bg-yellow-100 text-yellow-800',
  incident_response: 'bg-red-100 text-red-800',
  risk_management: 'bg-pink-100 text-pink-800',
  vendor_management: 'bg-indigo-100 text-indigo-800',
};

const difficultyLabels: Record<string, string> = {
  beginner: 'Beginner',
  intermediate: 'Intermediate',
  advanced: 'Advanced',
  expert: 'Expert',
};

const difficultyColors: Record<string, string> = {
  beginner: 'bg-green-100 text-green-700',
  intermediate: 'bg-blue-100 text-blue-700',
  advanced: 'bg-orange-100 text-orange-700',
  expert: 'bg-red-100 text-red-700',
};

export default function KnowledgeBasePage() {
  const [articles, setArticles] = useState<KnowledgeArticle[]>([]);
  const [recommended, setRecommended] = useState<KnowledgeArticle[]>([]);
  const [loading, setLoading] = useState(true);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);

  // Filters
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedType, setSelectedType] = useState('');
  const [selectedDifficulty, setSelectedDifficulty] = useState('');
  const [selectedFramework, setSelectedFramework] = useState('');
  const [sortBy, setSortBy] = useState('date');
  const [showBookmarks, setShowBookmarks] = useState(false);

  const fetchArticles = async () => {
    setLoading(true);
    try {
      if (showBookmarks) {
        const res = await api.getKnowledgeBookmarks(page, 20);
        setArticles(res.data || []);
        setTotal(res.total_count || 0);
        setTotalPages(Math.ceil((res.total_count || 0) / 20));
      } else {
        const params = new URLSearchParams({
          page: String(page),
          page_size: '20',
          sort: sortBy,
        });
        if (searchQuery) params.set('query', searchQuery);
        if (selectedType) params.set('types', selectedType);
        if (selectedDifficulty) params.set('difficulty', selectedDifficulty);
        if (selectedFramework) params.set('frameworks', selectedFramework);

        const res = await api.browseKnowledge(params.toString());
        const data = res.data;
        setArticles(data.articles || []);
        setTotal(data.total || 0);
        setTotalPages(data.total_pages || 1);
      }
    } catch (err) {
      console.error('Failed to fetch articles:', err);
    } finally {
      setLoading(false);
    }
  };

  const fetchRecommended = async () => {
    try {
      const res = await api.getRecommendedArticles(5);
      setRecommended(res.data || []);
    } catch (err) {
      console.error('Failed to fetch recommendations:', err);
    }
  };

  useEffect(() => {
    fetchArticles();
  }, [page, sortBy, selectedType, selectedDifficulty, selectedFramework, showBookmarks]);

  useEffect(() => {
    fetchRecommended();
  }, []);

  useEffect(() => {
    const timer = setTimeout(() => {
      if (page === 1) fetchArticles();
      else setPage(1);
    }, 300);
    return () => clearTimeout(timer);
  }, [searchQuery]);

  const toggleBookmark = async (articleId: string) => {
    try {
      await api.toggleKnowledgeBookmark(articleId);
      setArticles(prev => prev.map(a =>
        a.id === articleId ? { ...a, is_bookmarked: !a.is_bookmarked } : a
      ));
    } catch (err) {
      console.error('Bookmark toggle failed:', err);
    }
  };

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <div className="bg-white border-b">
        <div className="max-w-7xl mx-auto px-4 py-6">
          <h1 className="text-2xl font-bold text-gray-900">Knowledge Base</h1>
          <p className="text-gray-600 mt-1">
            Compliance guidance, implementation guides, and best practices.
          </p>
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-4 py-6">
        {/* Recommended Section */}
        {recommended.length > 0 && !showBookmarks && !searchQuery && (
          <div className="mb-8">
            <h2 className="text-lg font-semibold text-gray-800 mb-3">Recommended for You</h2>
            <div className="grid grid-cols-1 md:grid-cols-3 lg:grid-cols-5 gap-4">
              {recommended.map(article => (
                <Link
                  key={article.id}
                  href={`/knowledge/${article.slug}`}
                  className="bg-white rounded-lg shadow p-4 hover:shadow-md transition"
                >
                  <span className={`px-2 py-0.5 rounded text-xs font-medium ${articleTypeColors[article.article_type] || 'bg-gray-100'}`}>
                    {articleTypeLabels[article.article_type] || article.article_type}
                  </span>
                  <h3 className="text-sm font-medium text-gray-900 mt-2 line-clamp-2">{article.title}</h3>
                  <div className="flex items-center gap-2 mt-2 text-xs text-gray-400">
                    <span>{article.reading_time_minutes} min</span>
                    <span className={`px-1.5 py-0.5 rounded ${difficultyColors[article.difficulty] || ''}`}>
                      {difficultyLabels[article.difficulty] || article.difficulty}
                    </span>
                  </div>
                </Link>
              ))}
            </div>
          </div>
        )}

        {/* Filters Bar */}
        <div className="bg-white rounded-lg shadow p-4 mb-6">
          <div className="flex flex-wrap items-center gap-4">
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Search knowledge base..."
              className="flex-1 min-w-[200px] px-3 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500"
            />
            <select
              value={selectedType}
              onChange={(e) => { setSelectedType(e.target.value); setPage(1); }}
              className="px-3 py-2 border rounded-lg bg-white"
            >
              <option value="">All Types</option>
              {Object.entries(articleTypeLabels).map(([k, v]) => (
                <option key={k} value={k}>{v}</option>
              ))}
            </select>
            <select
              value={selectedDifficulty}
              onChange={(e) => { setSelectedDifficulty(e.target.value); setPage(1); }}
              className="px-3 py-2 border rounded-lg bg-white"
            >
              <option value="">All Levels</option>
              {Object.entries(difficultyLabels).map(([k, v]) => (
                <option key={k} value={k}>{v}</option>
              ))}
            </select>
            <select
              value={sortBy}
              onChange={(e) => setSortBy(e.target.value)}
              className="px-3 py-2 border rounded-lg bg-white"
            >
              <option value="date">Newest</option>
              <option value="popular">Most Viewed</option>
              <option value="helpful">Most Helpful</option>
              {searchQuery && <option value="relevance">Relevance</option>}
            </select>
            <button
              onClick={() => { setShowBookmarks(!showBookmarks); setPage(1); }}
              className={`px-3 py-2 rounded-lg border ${showBookmarks ? 'bg-blue-50 border-blue-300 text-blue-700' : ''}`}
            >
              Bookmarks
            </button>
          </div>
        </div>

        {/* Results */}
        <div className="text-sm text-gray-500 mb-4">
          {total} {showBookmarks ? 'bookmarked' : ''} article{total !== 1 ? 's' : ''}
        </div>

        {loading ? (
          <div className="text-center py-12 text-gray-500">Loading articles...</div>
        ) : articles.length === 0 ? (
          <div className="text-center py-12 bg-white rounded-lg shadow">
            <div className="text-4xl mb-3">&#128218;</div>
            <h3 className="text-lg font-semibold text-gray-700">
              {showBookmarks ? 'No bookmarked articles' : 'No articles found'}
            </h3>
            <p className="text-gray-500 mt-1">
              {showBookmarks ? 'Bookmark articles to find them quickly later.' : 'Try adjusting your filters.'}
            </p>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {articles.map(article => (
              <div key={article.id} className="bg-white rounded-lg shadow hover:shadow-md transition">
                <Link href={`/knowledge/${article.slug}`} className="block p-4">
                  <div className="flex items-center gap-2 mb-2">
                    <span className={`px-2 py-0.5 rounded text-xs font-medium ${articleTypeColors[article.article_type] || 'bg-gray-100'}`}>
                      {articleTypeLabels[article.article_type] || article.article_type}
                    </span>
                    <span className={`px-1.5 py-0.5 rounded text-xs ${difficultyColors[article.difficulty] || ''}`}>
                      {difficultyLabels[article.difficulty] || article.difficulty}
                    </span>
                  </div>
                  <h3 className="text-base font-semibold text-gray-900 line-clamp-2">{article.title}</h3>
                  {article.summary && (
                    <p className="text-sm text-gray-600 mt-1 line-clamp-2">{article.summary}</p>
                  )}
                  {article.applicable_frameworks?.length > 0 && (
                    <div className="flex flex-wrap gap-1 mt-2">
                      {article.applicable_frameworks.map(fw => (
                        <span key={fw} className="px-1.5 py-0.5 rounded bg-gray-100 text-xs text-gray-600">{fw}</span>
                      ))}
                    </div>
                  )}
                  <div className="flex items-center gap-3 mt-3 text-xs text-gray-400">
                    <span>{article.reading_time_minutes} min read</span>
                    <span>{article.view_count} views</span>
                    <span>{article.helpful_count} found helpful</span>
                  </div>
                </Link>
                <div className="px-4 pb-3 flex justify-end">
                  <button
                    onClick={(e) => { e.preventDefault(); toggleBookmark(article.id); }}
                    className="text-sm text-gray-400 hover:text-blue-600"
                    title={article.is_bookmarked ? 'Remove bookmark' : 'Bookmark'}
                  >
                    {article.is_bookmarked ? '\u2605' : '\u2606'}
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}

        {/* Pagination */}
        {totalPages > 1 && (
          <div className="flex justify-center gap-2 py-6">
            <button
              onClick={() => setPage(p => Math.max(1, p - 1))}
              disabled={page <= 1}
              className="px-3 py-1 rounded border disabled:opacity-50"
            >
              Previous
            </button>
            <span className="px-3 py-1 text-sm text-gray-600">
              Page {page} of {totalPages}
            </span>
            <button
              onClick={() => setPage(p => Math.min(totalPages, p + 1))}
              disabled={page >= totalPages}
              className="px-3 py-1 rounded border disabled:opacity-50"
            >
              Next
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
