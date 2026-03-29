'use client';

import { useState, useEffect } from 'react';
import { useParams } from 'next/navigation';
import Link from 'next/link';
import api from '@/lib/api';

// ============================================================
// KNOWLEDGE ARTICLE DETAIL — /knowledge/[slug]
// Full article view with rendered markdown, table of contents,
// related controls sidebar, helpful/not-helpful buttons, and
// bookmark toggle.
// ============================================================

interface KnowledgeArticle {
  id: string;
  article_type: string;
  title: string;
  slug: string;
  content_markdown: string;
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
  not_helpful_count: number;
  created_at: string;
  updated_at: string;
}

const difficultyColors: Record<string, string> = {
  beginner: 'bg-green-100 text-green-700',
  intermediate: 'bg-blue-100 text-blue-700',
  advanced: 'bg-orange-100 text-orange-700',
  expert: 'bg-red-100 text-red-700',
};

export default function KnowledgeArticlePage() {
  const params = useParams();
  const slug = params.slug as string;

  const [article, setArticle] = useState<KnowledgeArticle | null>(null);
  const [loading, setLoading] = useState(true);
  const [feedbackGiven, setFeedbackGiven] = useState<string | null>(null);

  useEffect(() => {
    const fetchArticle = async () => {
      try {
        const res = await api.getKnowledgeArticle(slug);
        setArticle(res.data);
      } catch (err) {
        console.error('Failed to fetch article:', err);
      } finally {
        setLoading(false);
      }
    };
    if (slug) fetchArticle();
  }, [slug]);

  const giveFeedback = async (action: string) => {
    if (!article || feedbackGiven) return;
    try {
      await api.submitArticleFeedback(article.id, { action });
      setFeedbackGiven(action);
      if (action === 'helpful') {
        setArticle(prev => prev ? { ...prev, helpful_count: prev.helpful_count + 1 } : null);
      } else if (action === 'not_helpful') {
        setArticle(prev => prev ? { ...prev, not_helpful_count: prev.not_helpful_count + 1 } : null);
      }
    } catch (err) {
      console.error('Feedback failed:', err);
    }
  };

  const toggleBookmark = async () => {
    if (!article) return;
    try {
      await api.toggleKnowledgeBookmark(article.id);
      setArticle(prev => prev ? { ...prev, is_bookmarked: !prev.is_bookmarked } : null);
    } catch (err) {
      console.error('Bookmark toggle failed:', err);
    }
  };

  // Extract headings for table of contents
  const extractHeadings = (markdown: string) => {
    const headings: { level: number; text: string; id: string }[] = [];
    const lines = markdown.split('\n');
    for (const line of lines) {
      const match = line.match(/^(#{1,3})\s+(.+)/);
      if (match) {
        const level = match[1].length;
        const text = match[2].trim();
        const id = text.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '');
        headings.push({ level, text, id });
      }
    }
    return headings;
  };

  // Simple markdown to HTML conversion
  const renderMarkdown = (md: string) => {
    let html = md
      // Headers
      .replace(/^### (.+)$/gm, '<h3 id="$1" class="text-lg font-semibold mt-6 mb-2 text-gray-900">$1</h3>')
      .replace(/^## (.+)$/gm, '<h2 id="$1" class="text-xl font-bold mt-8 mb-3 text-gray-900">$1</h2>')
      .replace(/^# (.+)$/gm, '<h1 id="$1" class="text-2xl font-bold mt-8 mb-4 text-gray-900">$1</h1>')
      // Bold
      .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
      // Italic
      .replace(/\*(.+?)\*/g, '<em>$1</em>')
      // Code blocks
      .replace(/```[\s\S]*?```/g, (match) => {
        const code = match.replace(/```\w*\n?/g, '').replace(/```/g, '');
        return `<pre class="bg-gray-900 text-green-300 p-4 rounded-lg overflow-x-auto my-4 text-sm"><code>${code}</code></pre>`;
      })
      // Inline code
      .replace(/`(.+?)`/g, '<code class="bg-gray-100 px-1.5 py-0.5 rounded text-sm text-pink-600">$1</code>')
      // Unordered lists
      .replace(/^- (.+)$/gm, '<li class="ml-4 list-disc">$1</li>')
      // Ordered lists
      .replace(/^\d+\. (.+)$/gm, '<li class="ml-4 list-decimal">$1</li>')
      // Checkboxes
      .replace(/^- \[x\] (.+)$/gm, '<li class="ml-4 flex items-center gap-2"><input type="checkbox" checked disabled class="rounded" /><span class="line-through text-gray-500">$1</span></li>')
      .replace(/^- \[ \] (.+)$/gm, '<li class="ml-4 flex items-center gap-2"><input type="checkbox" disabled class="rounded" /><span>$1</span></li>')
      // Tables (basic)
      .replace(/\|(.+)\|/g, (match) => {
        const cells = match.split('|').filter(c => c.trim());
        if (cells.every(c => c.trim().match(/^-+$/))) return '';
        const tds = cells.map(c => `<td class="px-3 py-2 border">${c.trim()}</td>`).join('');
        return `<tr>${tds}</tr>`;
      })
      // Paragraphs
      .replace(/\n\n/g, '</p><p class="my-3 text-gray-700 leading-relaxed">')
      // Line breaks
      .replace(/\n/g, '<br/>');

    // Fix heading IDs
    html = html.replace(/id="([^"]+)"/g, (_, text) => {
      const id = text.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '');
      return `id="${id}"`;
    });

    return `<div class="prose max-w-none"><p class="my-3 text-gray-700 leading-relaxed">${html}</p></div>`;
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-gray-500">Loading article...</div>
      </div>
    );
  }

  if (!article) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="text-4xl mb-3">&#128218;</div>
          <h2 className="text-xl font-semibold text-gray-700">Article Not Found</h2>
          <Link href="/knowledge" className="text-blue-600 mt-2 inline-block">
            Back to Knowledge Base
          </Link>
        </div>
      </div>
    );
  }

  const headings = extractHeadings(article.content_markdown);

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <div className="bg-white border-b">
        <div className="max-w-7xl mx-auto px-4 py-6">
          <Link href="/knowledge" className="text-sm text-blue-600 hover:underline mb-2 inline-block">
            &larr; Knowledge Base
          </Link>
          <div className="flex items-center gap-3 mb-2">
            <span className="px-2 py-0.5 rounded text-xs font-medium bg-blue-100 text-blue-800 capitalize">
              {article.article_type.replace(/_/g, ' ')}
            </span>
            <span className={`px-2 py-0.5 rounded text-xs ${difficultyColors[article.difficulty] || ''}`}>
              {article.difficulty}
            </span>
            <span className="text-xs text-gray-400">{article.reading_time_minutes} min read</span>
            <span className="text-xs text-gray-400">{article.view_count} views</span>
          </div>
          <h1 className="text-2xl font-bold text-gray-900">{article.title}</h1>
          {article.summary && (
            <p className="text-gray-600 mt-2">{article.summary}</p>
          )}
          <div className="flex items-center gap-2 mt-3">
            <button onClick={toggleBookmark} className="text-sm text-gray-500 hover:text-blue-600">
              {article.is_bookmarked ? '\u2605 Bookmarked' : '\u2606 Bookmark'}
            </button>
          </div>
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-4 py-6 flex gap-8">
        {/* Table of Contents */}
        {headings.length > 0 && (
          <div className="w-64 flex-shrink-0 hidden lg:block">
            <div className="sticky top-20 bg-white rounded-lg shadow p-4">
              <h3 className="font-semibold text-sm text-gray-700 mb-3">Contents</h3>
              <nav className="space-y-1">
                {headings.map((h, i) => (
                  <a
                    key={i}
                    href={`#${h.id}`}
                    className={`block text-sm text-gray-600 hover:text-blue-600 ${
                      h.level === 1 ? '' : h.level === 2 ? 'pl-3' : 'pl-6'
                    }`}
                  >
                    {h.text}
                  </a>
                ))}
              </nav>

              {/* Related Controls */}
              {article.applicable_control_codes?.length > 0 && (
                <div className="mt-6 pt-4 border-t">
                  <h3 className="font-semibold text-sm text-gray-700 mb-2">Related Controls</h3>
                  <div className="flex flex-wrap gap-1">
                    {article.applicable_control_codes.map(code => (
                      <span key={code} className="px-2 py-0.5 rounded bg-gray-100 text-xs text-gray-700 font-mono">
                        {code}
                      </span>
                    ))}
                  </div>
                </div>
              )}

              {/* Frameworks */}
              {article.applicable_frameworks?.length > 0 && (
                <div className="mt-4 pt-4 border-t">
                  <h3 className="font-semibold text-sm text-gray-700 mb-2">Frameworks</h3>
                  <div className="flex flex-wrap gap-1">
                    {article.applicable_frameworks.map(fw => (
                      <span key={fw} className="px-2 py-0.5 rounded bg-blue-50 text-xs text-blue-700">
                        {fw}
                      </span>
                    ))}
                  </div>
                </div>
              )}

              {/* Tags */}
              {article.tags?.length > 0 && (
                <div className="mt-4 pt-4 border-t">
                  <h3 className="font-semibold text-sm text-gray-700 mb-2">Tags</h3>
                  <div className="flex flex-wrap gap-1">
                    {article.tags.map(tag => (
                      <span key={tag} className="px-2 py-0.5 rounded bg-gray-100 text-xs text-gray-600">
                        #{tag}
                      </span>
                    ))}
                  </div>
                </div>
              )}
            </div>
          </div>
        )}

        {/* Article Content */}
        <div className="flex-1 bg-white rounded-lg shadow p-8">
          <div
            dangerouslySetInnerHTML={{ __html: renderMarkdown(article.content_markdown) }}
          />

          {/* Feedback Section */}
          <div className="mt-12 pt-6 border-t text-center">
            <p className="text-gray-600 mb-3">Was this article helpful?</p>
            <div className="flex justify-center gap-4">
              <button
                onClick={() => giveFeedback('helpful')}
                disabled={!!feedbackGiven}
                className={`px-4 py-2 rounded-lg border ${
                  feedbackGiven === 'helpful'
                    ? 'bg-green-50 border-green-300 text-green-700'
                    : 'hover:bg-gray-50'
                } disabled:cursor-not-allowed`}
              >
                &#128077; Helpful ({article.helpful_count})
              </button>
              <button
                onClick={() => giveFeedback('not_helpful')}
                disabled={!!feedbackGiven}
                className={`px-4 py-2 rounded-lg border ${
                  feedbackGiven === 'not_helpful'
                    ? 'bg-red-50 border-red-300 text-red-700'
                    : 'hover:bg-gray-50'
                } disabled:cursor-not-allowed`}
              >
                &#128078; Not Helpful ({article.not_helpful_count})
              </button>
            </div>
            {feedbackGiven && (
              <p className="text-sm text-gray-500 mt-2">Thanks for your feedback!</p>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
