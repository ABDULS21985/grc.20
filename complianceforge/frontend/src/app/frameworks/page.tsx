'use client';

import { useEffect, useState } from 'react';
import api from '@/lib/api';
import type { ComplianceFramework } from '@/types';

const CATEGORY_COLORS: Record<string, string> = {
  security: 'badge-info',
  privacy: 'badge-high',
  governance: 'badge-medium',
  risk: 'badge-critical',
  operational: 'badge-low',
};

export default function FrameworksPage() {
  const [frameworks, setFrameworks] = useState<ComplianceFramework[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    api.getFrameworks()
      .then((res) => setFrameworks(res.data?.data || res.data || []))
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, []);

  if (loading) {
    return (
      <div>
        <div className="flex items-center justify-between mb-8">
          <div>
            <h1 className="text-2xl font-bold text-gray-900">Compliance Frameworks</h1>
            <p className="text-gray-500 mt-1">Loading frameworks...</p>
          </div>
        </div>
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
          {Array.from({ length: 9 }).map((_, i) => (
            <div key={i} className="card animate-pulse">
              <div className="flex items-start gap-3">
                <div className="h-10 w-10 rounded-lg bg-gray-200" />
                <div className="flex-1 space-y-2">
                  <div className="h-4 bg-gray-200 rounded w-3/4" />
                  <div className="h-3 bg-gray-200 rounded w-1/2" />
                </div>
              </div>
              <div className="h-3 bg-gray-200 rounded w-full mt-3" />
              <div className="h-3 bg-gray-200 rounded w-2/3 mt-2" />
            </div>
          ))}
        </div>
      </div>
    );
  }

  if (error) return <div className="rounded-lg bg-red-50 border border-red-200 p-4 text-red-700">{error}</div>;

  return (
    <div>
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Compliance Frameworks</h1>
          <p className="text-gray-500 mt-1">{frameworks.length} standards covering security, privacy, governance, and operations</p>
        </div>
        <button className="btn-primary">Adopt Framework</button>
      </div>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
        {frameworks.map(fw => (
          <a
            key={fw.id}
            href={`/frameworks/${fw.id}`}
            className="card hover:border-indigo-300 hover:shadow-md transition-all group"
          >
            <div className="flex items-start gap-3">
              <div
                className="h-10 w-10 rounded-lg flex items-center justify-center flex-shrink-0"
                style={{ backgroundColor: (fw.color_hex || '#4f46e5') + '20' }}
              >
                <span className="text-sm font-bold" style={{ color: fw.color_hex || '#4f46e5' }}>
                  {fw.code.substring(0, 2)}
                </span>
              </div>
              <div className="flex-1 min-w-0">
                <h3 className="font-semibold text-gray-900 group-hover:text-indigo-600 transition-colors">{fw.name}</h3>
                <p className="text-xs text-gray-500 mt-0.5">{fw.issuing_body} • v{fw.version}</p>
              </div>
              <span className={`badge ${CATEGORY_COLORS[fw.category] || 'badge-info'}`}>
                {fw.category}
              </span>
            </div>

            <p className="text-sm text-gray-600 mt-3">{fw.description}</p>

            <div className="flex items-center justify-between mt-4 pt-3 border-t border-gray-100">
              <span className="text-sm text-gray-500">{fw.total_controls} controls</span>
              <span className="text-xs font-medium text-indigo-600 group-hover:underline">View Controls →</span>
            </div>
          </a>
        ))}
      </div>

      {frameworks.length === 0 && !loading && (
        <div className="text-center py-12">
          <p className="text-gray-500">No frameworks adopted yet.</p>
          <button className="btn-primary mt-4">Adopt Your First Framework</button>
        </div>
      )}
    </div>
  );
}
