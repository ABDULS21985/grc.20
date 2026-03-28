'use client';

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import api from '@/lib/api';
import type { Risk } from '@/types';

export default function RiskDetailPage() {
  const params = useParams();
  const id = params.id as string;

  const [risk, setRisk] = useState<Risk | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    api.getRisk(id)
      .then((res) => setRisk(res.data))
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, [id]);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <p className="text-gray-500">Loading risk...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="rounded-lg bg-red-50 border border-red-200 p-4 text-red-700">{error}</div>
    );
  }

  if (!risk) {
    return (
      <div className="rounded-lg bg-gray-50 border border-gray-200 p-4 text-gray-600">
        Risk not found.
      </div>
    );
  }

  const inherentScore = risk.inherent_likelihood * risk.inherent_impact;

  return (
    <div>
      {/* Back button */}
      <a href="/risks" className="inline-flex items-center text-sm text-gray-500 hover:text-indigo-600 mb-4">
        &larr; Back to Risk Register
      </a>

      {/* Header */}
      <div className="flex items-start justify-between mb-6">
        <div>
          <div className="flex items-center gap-3 mb-1">
            <span className="text-sm font-mono text-gray-400">{risk.risk_ref}</span>
            <span className={`badge badge-${risk.status === 'closed' ? 'low' : 'info'}`}>{risk.status}</span>
            <span className={`badge badge-${risk.residual_risk_level}`}>{risk.residual_risk_level}</span>
          </div>
          <h1 className="text-2xl font-bold text-gray-900">{risk.title}</h1>
        </div>
        <div className="text-right">
          <p className="text-sm text-gray-500">Residual Risk Score</p>
          <p className={`text-3xl font-bold ${risk.residual_risk_score >= 20 ? 'text-red-600' : risk.residual_risk_score >= 12 ? 'text-orange-600' : risk.residual_risk_score >= 6 ? 'text-yellow-600' : 'text-green-600'}`}>
            {risk.residual_risk_score}
          </p>
        </div>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-3 gap-4 mb-6">
        <div className="card p-4">
          <p className="text-xs font-medium text-gray-500 uppercase tracking-wide">Inherent Score</p>
          <p className={`text-2xl font-bold mt-1 ${inherentScore >= 20 ? 'text-red-600' : inherentScore >= 12 ? 'text-orange-600' : inherentScore >= 6 ? 'text-yellow-600' : 'text-green-600'}`}>
            {inherentScore}
          </p>
          <p className="text-xs text-gray-400 mt-0.5">
            {risk.inherent_likelihood} (likelihood) x {risk.inherent_impact} (impact)
          </p>
        </div>
        <div className="card p-4">
          <p className="text-xs font-medium text-gray-500 uppercase tracking-wide">Residual Score</p>
          <p className={`text-2xl font-bold mt-1 ${risk.residual_risk_score >= 20 ? 'text-red-600' : risk.residual_risk_score >= 12 ? 'text-orange-600' : risk.residual_risk_score >= 6 ? 'text-yellow-600' : 'text-green-600'}`}>
            {risk.residual_risk_score}
          </p>
          <p className="text-xs text-gray-400 mt-0.5">
            {risk.residual_likelihood} (likelihood) x {risk.residual_impact} (impact)
          </p>
        </div>
        <div className="card p-4">
          <p className="text-xs font-medium text-gray-500 uppercase tracking-wide">Financial Impact</p>
          <p className="text-2xl font-bold text-gray-900 mt-1">
            {risk.financial_impact_eur > 0 ? `€${risk.financial_impact_eur.toLocaleString()}` : '—'}
          </p>
          <p className="text-xs text-gray-400 mt-0.5">Estimated EUR exposure</p>
        </div>
      </div>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        {/* Overview */}
        <div className="card">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Overview</h2>
          <div className="space-y-3">
            {risk.description && (
              <div>
                <p className="text-xs font-medium text-gray-500 uppercase">Description</p>
                <p className="text-sm text-gray-700 mt-1">{risk.description}</p>
              </div>
            )}
            <DetailRow label="Source" value={risk.risk_source || '—'} />
            <DetailRow label="Velocity" value={risk.risk_velocity || '—'} />
            <DetailRow label="Category ID" value={risk.risk_category_id || '—'} />
            <DetailRow label="Next Review" value={risk.next_review_date ? new Date(risk.next_review_date).toLocaleDateString('en-GB') : '—'} />
            <DetailRow label="Created" value={new Date(risk.created_at).toLocaleDateString('en-GB')} />
            {risk.tags && risk.tags.length > 0 && (
              <div>
                <p className="text-xs font-medium text-gray-500 uppercase mb-1">Tags</p>
                <div className="flex flex-wrap gap-1">
                  {risk.tags.map((tag) => (
                    <span key={tag} className="badge badge-info text-xs">{tag}</span>
                  ))}
                </div>
              </div>
            )}
          </div>
        </div>

        {/* Risk Matrix */}
        <div className="card">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Risk Matrix</h2>
          <p className="text-xs text-gray-500 mb-3">Showing inherent and residual risk positions on the 5x5 matrix</p>
          <div className="grid grid-cols-5 gap-1">
            {[5, 4, 3, 2, 1].map((impact) =>
              [1, 2, 3, 4, 5].map((likelihood) => {
                const score = likelihood * impact;
                const isInherent =
                  risk.inherent_likelihood === likelihood && risk.inherent_impact === impact;
                const isResidual =
                  risk.residual_likelihood === likelihood && risk.residual_impact === impact;

                return (
                  <div
                    key={`${likelihood}-${impact}`}
                    className={`h-10 rounded text-xs flex items-center justify-center font-medium text-white relative
                      ${score >= 20 ? 'bg-red-600' : score >= 12 ? 'bg-orange-500' : score >= 6 ? 'bg-yellow-500' : score >= 3 ? 'bg-green-400' : 'bg-green-300'}
                      ${isInherent || isResidual ? 'ring-2 ring-offset-1' : ''}
                      ${isInherent ? 'ring-gray-900' : ''}
                      ${isResidual && !isInherent ? 'ring-blue-600' : ''}`}
                  >
                    {score}
                    {isInherent && (
                      <span className="absolute -top-1 -right-1 h-3 w-3 rounded-full bg-gray-900 border border-white" title="Inherent" />
                    )}
                    {isResidual && (
                      <span className="absolute -bottom-1 -right-1 h-3 w-3 rounded-full bg-blue-600 border border-white" title="Residual" />
                    )}
                  </div>
                );
              })
            )}
          </div>
          <div className="flex justify-between mt-2">
            <span className="text-xs text-gray-400">Likelihood &rarr;</span>
            <span className="text-xs text-gray-400">&uarr; Impact</span>
          </div>
          <div className="flex items-center gap-4 mt-3 text-xs text-gray-500">
            <span className="flex items-center gap-1">
              <span className="inline-block h-2.5 w-2.5 rounded-full bg-gray-900" /> Inherent ({inherentScore})
            </span>
            <span className="flex items-center gap-1">
              <span className="inline-block h-2.5 w-2.5 rounded-full bg-blue-600" /> Residual ({risk.residual_risk_score})
            </span>
          </div>
        </div>
      </div>
    </div>
  );
}

function DetailRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between py-2 border-b border-gray-50">
      <span className="text-sm text-gray-500">{label}</span>
      <span className="text-sm font-medium text-gray-900">{value}</span>
    </div>
  );
}
