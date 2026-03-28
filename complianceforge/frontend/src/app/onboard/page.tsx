'use client';

import { useEffect, useState, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import api from '@/lib/api';

// ============================================================
// Onboarding Wizard - 7-Step Self-Service Setup
// ============================================================

const STEPS = [
  { number: 1, title: 'Organisation Profile', description: 'Tell us about your organisation' },
  { number: 2, title: 'Industry Assessment', description: 'Help us understand your compliance needs' },
  { number: 3, title: 'Framework Selection', description: 'Choose your compliance frameworks' },
  { number: 4, title: 'Team Setup', description: 'Invite your team members' },
  { number: 5, title: 'Risk Appetite', description: 'Define your risk tolerance' },
  { number: 6, title: 'Quick Assessment', description: 'Rate your current maturity' },
  { number: 7, title: 'Complete Setup', description: 'Review and finalise' },
];

interface OnboardingProgress {
  current_step: number;
  total_steps: number;
  completed_steps: number[];
  is_completed: boolean;
  skipped_steps: number[];
  org_profile_data: Record<string, string>;
  industry_assessment_data: Record<string, boolean>;
  selected_framework_ids: string[];
  team_invitations: TeamInvitation[];
  risk_appetite_data: Record<string, string>;
  quick_assessment_data: Record<string, number>;
}

interface TeamInvitation {
  email: string;
  first_name: string;
  last_name: string;
  role: string;
  job_title: string;
}

interface FrameworkRecommendation {
  framework_id: string;
  framework_code: string;
  framework_name: string;
  reason: string;
  priority: string;
  description: string;
}

export default function OnboardingWizardPage() {
  const router = useRouter();
  const [step, setStep] = useState(1);
  const [progress, setProgress] = useState<OnboardingProgress | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Step data
  const [orgProfile, setOrgProfile] = useState({
    display_name: '', legal_name: '', industry: '', country_code: '',
    timezone: 'Europe/London', employee_count_range: '', website: '',
  });
  const [assessment, setAssessment] = useState({
    process_payment_cards: false, handle_eu_personal_data: false,
    essential_services: false, uk_public_sector: false,
    iso_certification: false, us_federal_contracts: false,
    cyber_maturity: false, itil_requirements: false, board_it_governance: false,
  });
  const [recommendations, setRecommendations] = useState<FrameworkRecommendation[]>([]);
  const [selectedFrameworks, setSelectedFrameworks] = useState<string[]>([]);
  const [invitations, setInvitations] = useState<TeamInvitation[]>([
    { email: '', first_name: '', last_name: '', role: 'compliance_manager', job_title: '' },
  ]);
  const [riskAppetite, setRiskAppetite] = useState({
    overall_appetite: 'moderate', acceptable_risk_level: 'medium',
    review_frequency: 'quarterly', notes: '',
  });
  const [quickAssessment, setQuickAssessment] = useState({
    security_maturity: 3, policy_documentation: 3, incident_response_ready: 3,
    data_protection_level: 3, access_control_maturity: 3, third_party_risk_mgmt: 3,
    business_continuity: 3, security_awareness: 3,
  });

  const loadProgress = useCallback(async () => {
    try {
      const res = await api.get<{ data: OnboardingProgress }>('/onboard/progress');
      const p = res.data;
      setProgress(p);
      if (p.current_step) setStep(p.current_step);
      if (p.org_profile_data && Object.keys(p.org_profile_data).length > 0) {
        setOrgProfile(prev => ({ ...prev, ...p.org_profile_data }));
      }
      if (p.industry_assessment_data && Object.keys(p.industry_assessment_data).length > 0) {
        setAssessment(prev => ({ ...prev, ...p.industry_assessment_data }));
      }
      if (p.selected_framework_ids?.length) {
        setSelectedFrameworks(p.selected_framework_ids);
      }
      if (p.team_invitations?.length) {
        setInvitations(p.team_invitations);
      }
      if (p.risk_appetite_data && Object.keys(p.risk_appetite_data).length > 0) {
        setRiskAppetite(prev => ({ ...prev, ...p.risk_appetite_data }));
      }
      if (p.quick_assessment_data && Object.keys(p.quick_assessment_data).length > 0) {
        setQuickAssessment(prev => ({ ...prev, ...p.quick_assessment_data }));
      }
      if (p.is_completed) {
        router.push('/dashboard');
      }
    } catch {
      // First time, no progress yet
    } finally {
      setLoading(false);
    }
  }, [router]);

  useEffect(() => { loadProgress(); }, [loadProgress]);

  const saveStep = async () => {
    setSaving(true);
    setError(null);
    try {
      let body: unknown;
      switch (step) {
        case 1: body = orgProfile; break;
        case 2: body = assessment; break;
        case 3: body = { framework_ids: selectedFrameworks }; break;
        case 4: body = { invitations: invitations.filter(i => i.email) }; break;
        case 5: body = riskAppetite; break;
        case 6: body = quickAssessment; break;
        default: return;
      }
      await api.put(`/onboard/step/${step}`, { body });
      if (step < 7) setStep(step + 1);
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Failed to save step';
      setError(message);
    } finally {
      setSaving(false);
    }
  };

  const skipStep = async () => {
    try {
      await api.post(`/onboard/step/${step}/skip`);
      if (step < 7) setStep(step + 1);
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Failed to skip step';
      setError(message);
    }
  };

  const completeOnboarding = async () => {
    setSaving(true);
    setError(null);
    try {
      await api.post('/onboard/complete');
      router.push('/dashboard');
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Failed to complete onboarding';
      setError(message);
    } finally {
      setSaving(false);
    }
  };

  const fetchRecommendations = async () => {
    try {
      const res = await api.post<{ data: FrameworkRecommendation[] }>('/onboard/recommendations', { body: assessment });
      setRecommendations(res.data || []);
    } catch {
      // Silently fail
    }
  };

  useEffect(() => {
    if (step === 3) fetchRecommendations();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [step]);

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <p className="text-gray-500">Loading onboarding wizard...</p>
      </div>
    );
  }

  const completedSteps = progress?.completed_steps || [];
  const progressPct = Math.round((completedSteps.length / 7) * 100);

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <div className="bg-white border-b border-gray-200 px-6 py-4">
        <div className="max-w-4xl mx-auto flex items-center justify-between">
          <div>
            <h1 className="text-xl font-bold text-gray-900">ComplianceForge Setup</h1>
            <p className="text-sm text-gray-500 mt-0.5">
              Step {step} of 7 &mdash; {STEPS[step - 1]?.title}
            </p>
          </div>
          <div className="text-right">
            <div className="text-sm font-medium text-gray-700">{progressPct}% complete</div>
          </div>
        </div>
      </div>

      {/* Progress Bar */}
      <div className="bg-white border-b border-gray-100">
        <div className="max-w-4xl mx-auto px-6 py-3">
          <div className="flex items-center gap-1">
            {STEPS.map((s) => (
              <div key={s.number} className="flex-1 flex flex-col items-center">
                <button
                  onClick={() => setStep(s.number)}
                  className={`w-8 h-8 rounded-full flex items-center justify-center text-xs font-medium transition-colors ${
                    s.number === step
                      ? 'bg-indigo-600 text-white'
                      : completedSteps.includes(s.number)
                      ? 'bg-green-100 text-green-700 border border-green-300'
                      : 'bg-gray-100 text-gray-400 border border-gray-200'
                  }`}
                >
                  {completedSteps.includes(s.number) ? '\u2713' : s.number}
                </button>
                <span className={`text-xs mt-1 hidden sm:block ${s.number === step ? 'text-indigo-600 font-medium' : 'text-gray-400'}`}>
                  {s.title}
                </span>
              </div>
            ))}
          </div>
          <div className="mt-2 h-1.5 bg-gray-100 rounded-full overflow-hidden">
            <div className="h-full bg-indigo-600 rounded-full transition-all duration-300" style={{ width: `${progressPct}%` }} />
          </div>
        </div>
      </div>

      {/* Error Banner */}
      {error && (
        <div className="max-w-4xl mx-auto px-6 mt-4">
          <div className="rounded-lg bg-red-50 border border-red-200 p-3 text-sm text-red-700">
            {error}
          </div>
        </div>
      )}

      {/* Step Content */}
      <div className="max-w-4xl mx-auto px-6 py-8">
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
          {step === 1 && <Step1OrgProfile data={orgProfile} onChange={setOrgProfile} />}
          {step === 2 && <Step2Assessment data={assessment} onChange={setAssessment} />}
          {step === 3 && (
            <Step3Frameworks
              recommendations={recommendations}
              selected={selectedFrameworks}
              onToggle={(id) => setSelectedFrameworks(prev =>
                prev.includes(id) ? prev.filter(f => f !== id) : [...prev, id]
              )}
            />
          )}
          {step === 4 && <Step4Team invitations={invitations} onChange={setInvitations} />}
          {step === 5 && <Step5RiskAppetite data={riskAppetite} onChange={setRiskAppetite} />}
          {step === 6 && <Step6QuickAssessment data={quickAssessment} onChange={setQuickAssessment} />}
          {step === 7 && (
            <Step7Complete
              progress={progress}
              selectedFrameworks={selectedFrameworks}
              invitations={invitations}
              recommendations={recommendations}
            />
          )}
        </div>

        {/* Navigation */}
        <div className="flex items-center justify-between mt-6">
          <button
            onClick={() => setStep(Math.max(1, step - 1))}
            disabled={step === 1}
            className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            Previous
          </button>

          <div className="flex items-center gap-3">
            {step < 7 && step > 1 && (
              <button onClick={skipStep} className="px-4 py-2 text-sm text-gray-500 hover:text-gray-700">
                Skip this step
              </button>
            )}
            {step < 7 ? (
              <button
                onClick={saveStep}
                disabled={saving}
                className="px-6 py-2 text-sm font-medium text-white bg-indigo-600 rounded-lg hover:bg-indigo-700 disabled:opacity-50"
              >
                {saving ? 'Saving...' : 'Save & Continue'}
              </button>
            ) : (
              <button
                onClick={completeOnboarding}
                disabled={saving}
                className="px-6 py-2 text-sm font-medium text-white bg-green-600 rounded-lg hover:bg-green-700 disabled:opacity-50"
              >
                {saving ? 'Completing...' : 'Complete Setup'}
              </button>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

// ============================================================
// STEP COMPONENTS
// ============================================================

function Step1OrgProfile({ data, onChange }: {
  data: Record<string, string>;
  onChange: (d: Record<string, string>) => void;
}) {
  const update = (key: string, value: string) => onChange({ ...data, [key]: value });

  return (
    <div>
      <h2 className="text-lg font-semibold text-gray-900 mb-4">Organisation Profile</h2>
      <p className="text-sm text-gray-500 mb-6">Provide basic information about your organisation. This helps us customise the platform for your needs.</p>
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <FormField label="Organisation Name *" value={data.display_name} onChange={v => update('display_name', v)} placeholder="Acme Corporation" />
        <FormField label="Legal Name" value={data.legal_name} onChange={v => update('legal_name', v)} placeholder="Acme Corporation Ltd" />
        <FormSelect label="Industry *" value={data.industry} onChange={v => update('industry', v)} options={[
          { value: '', label: 'Select industry...' },
          { value: 'technology', label: 'Technology' },
          { value: 'financial_services', label: 'Financial Services' },
          { value: 'healthcare', label: 'Healthcare' },
          { value: 'government', label: 'Government / Public Sector' },
          { value: 'energy', label: 'Energy & Utilities' },
          { value: 'manufacturing', label: 'Manufacturing' },
          { value: 'retail', label: 'Retail & E-commerce' },
          { value: 'education', label: 'Education' },
          { value: 'professional_services', label: 'Professional Services' },
          { value: 'other', label: 'Other' },
        ]} />
        <FormSelect label="Country *" value={data.country_code} onChange={v => update('country_code', v)} options={[
          { value: '', label: 'Select country...' },
          { value: 'GB', label: 'United Kingdom' },
          { value: 'DE', label: 'Germany' },
          { value: 'FR', label: 'France' },
          { value: 'NL', label: 'Netherlands' },
          { value: 'IE', label: 'Ireland' },
          { value: 'US', label: 'United States' },
          { value: 'CA', label: 'Canada' },
          { value: 'AU', label: 'Australia' },
        ]} />
        <FormSelect label="Employee Count" value={data.employee_count_range} onChange={v => update('employee_count_range', v)} options={[
          { value: '', label: 'Select range...' },
          { value: '1-10', label: '1-10' },
          { value: '11-49', label: '11-49' },
          { value: '50-249', label: '50-249' },
          { value: '250-999', label: '250-999' },
          { value: '1000+', label: '1000+' },
        ]} />
        <FormField label="Website" value={data.website} onChange={v => update('website', v)} placeholder="https://acme.com" />
      </div>
    </div>
  );
}

function Step2Assessment({ data, onChange }: {
  data: Record<string, boolean>;
  onChange: (d: Record<string, boolean>) => void;
}) {
  const toggle = (key: string) => onChange({ ...data, [key]: !data[key] });

  const questions = [
    { key: 'process_payment_cards', label: 'Do you process payment card data?', hint: 'Credit/debit card transactions' },
    { key: 'handle_eu_personal_data', label: 'Do you handle EU/UK personal data?', hint: 'Customer or employee data of EU/UK residents' },
    { key: 'essential_services', label: 'Are you an essential services provider?', hint: 'Energy, transport, health, water, digital infrastructure' },
    { key: 'uk_public_sector', label: 'Are you a UK public sector organisation?', hint: 'Government, NHS, local authority, or public body' },
    { key: 'iso_certification', label: 'Are you pursuing ISO certification?', hint: 'ISO 27001 or similar management system certification' },
    { key: 'us_federal_contracts', label: 'Do you have US federal government contracts?', hint: 'DoD, civilian agency, or subcontractor' },
    { key: 'cyber_maturity', label: 'Do you need a cybersecurity maturity model?', hint: 'Structured approach to improving security posture' },
    { key: 'itil_requirements', label: 'Do you have ITIL/ITSM requirements?', hint: 'IT service management alignment needed' },
    { key: 'board_it_governance', label: 'Does your board require IT governance reporting?', hint: 'Board-level oversight of IT and technology' },
  ];

  return (
    <div>
      <h2 className="text-lg font-semibold text-gray-900 mb-4">Industry Assessment</h2>
      <p className="text-sm text-gray-500 mb-6">Answer these questions to help us recommend the right compliance frameworks for your organisation.</p>
      <div className="space-y-3">
        {questions.map(q => (
          <label key={q.key} className={`flex items-start gap-3 p-4 rounded-lg border cursor-pointer transition-colors ${data[q.key] ? 'border-indigo-300 bg-indigo-50' : 'border-gray-200 hover:bg-gray-50'}`}>
            <input type="checkbox" checked={data[q.key] || false} onChange={() => toggle(q.key)}
              className="mt-0.5 h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-500" />
            <div>
              <div className="text-sm font-medium text-gray-900">{q.label}</div>
              <div className="text-xs text-gray-500 mt-0.5">{q.hint}</div>
            </div>
          </label>
        ))}
      </div>
    </div>
  );
}

function Step3Frameworks({ recommendations, selected, onToggle }: {
  recommendations: FrameworkRecommendation[];
  selected: string[];
  onToggle: (id: string) => void;
}) {
  return (
    <div>
      <h2 className="text-lg font-semibold text-gray-900 mb-4">Framework Selection</h2>
      <p className="text-sm text-gray-500 mb-6">
        Based on your assessment, we recommend the following frameworks. Select the ones you want to adopt.
      </p>
      {recommendations.length === 0 ? (
        <p className="text-sm text-gray-400 italic">Complete the Industry Assessment to see recommendations.</p>
      ) : (
        <div className="space-y-3">
          {recommendations.map(fw => (
            <label key={fw.framework_id || fw.framework_code}
              className={`flex items-start gap-3 p-4 rounded-lg border cursor-pointer transition-colors ${
                selected.includes(fw.framework_id)
                  ? 'border-indigo-300 bg-indigo-50'
                  : 'border-gray-200 hover:bg-gray-50'
              }`}
            >
              <input type="checkbox" checked={selected.includes(fw.framework_id)}
                onChange={() => onToggle(fw.framework_id)}
                className="mt-0.5 h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
                disabled={!fw.framework_id}
              />
              <div className="flex-1">
                <div className="flex items-center gap-2">
                  <span className="text-sm font-medium text-gray-900">{fw.framework_name}</span>
                  <span className={`text-xs px-2 py-0.5 rounded-full ${
                    fw.priority === 'required'
                      ? 'bg-red-100 text-red-700'
                      : fw.priority === 'recommended'
                      ? 'bg-amber-100 text-amber-700'
                      : 'bg-gray-100 text-gray-600'
                  }`}>
                    {fw.priority}
                  </span>
                </div>
                <p className="text-xs text-gray-500 mt-1">{fw.reason}</p>
              </div>
            </label>
          ))}
        </div>
      )}
      <div className="mt-4 text-xs text-gray-400">
        Selected: {selected.length} framework{selected.length !== 1 ? 's' : ''}
      </div>
    </div>
  );
}

function Step4Team({ invitations, onChange }: {
  invitations: TeamInvitation[];
  onChange: (inv: TeamInvitation[]) => void;
}) {
  const updateInv = (idx: number, key: keyof TeamInvitation, value: string) => {
    const updated = [...invitations];
    updated[idx] = { ...updated[idx], [key]: value };
    onChange(updated);
  };

  const addInvitation = () => {
    onChange([...invitations, { email: '', first_name: '', last_name: '', role: 'compliance_manager', job_title: '' }]);
  };

  const removeInvitation = (idx: number) => {
    onChange(invitations.filter((_, i) => i !== idx));
  };

  return (
    <div>
      <h2 className="text-lg font-semibold text-gray-900 mb-4">Team Setup</h2>
      <p className="text-sm text-gray-500 mb-6">Invite team members to help manage your compliance programme.</p>
      <div className="space-y-4">
        {invitations.map((inv, idx) => (
          <div key={idx} className="p-4 rounded-lg border border-gray-200 bg-gray-50">
            <div className="flex items-center justify-between mb-3">
              <span className="text-sm font-medium text-gray-700">Team Member {idx + 1}</span>
              {invitations.length > 1 && (
                <button onClick={() => removeInvitation(idx)} className="text-xs text-red-500 hover:text-red-700">Remove</button>
              )}
            </div>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <FormField label="Email" value={inv.email} onChange={v => updateInv(idx, 'email', v)} placeholder="user@company.com" />
              <FormField label="First Name" value={inv.first_name} onChange={v => updateInv(idx, 'first_name', v)} placeholder="Jane" />
              <FormField label="Last Name" value={inv.last_name} onChange={v => updateInv(idx, 'last_name', v)} placeholder="Smith" />
              <FormSelect label="Role" value={inv.role} onChange={v => updateInv(idx, 'role', v)} options={[
                { value: 'compliance_manager', label: 'Compliance Manager' },
                { value: 'risk_manager', label: 'Risk Manager' },
                { value: 'auditor', label: 'Auditor' },
                { value: 'viewer', label: 'Viewer' },
              ]} />
            </div>
          </div>
        ))}
      </div>
      <button onClick={addInvitation} className="mt-4 text-sm text-indigo-600 hover:text-indigo-700 font-medium">
        + Add another team member
      </button>
    </div>
  );
}

function Step5RiskAppetite({ data, onChange }: {
  data: Record<string, string>;
  onChange: (d: Record<string, string>) => void;
}) {
  const update = (key: string, value: string) => onChange({ ...data, [key]: value });

  return (
    <div>
      <h2 className="text-lg font-semibold text-gray-900 mb-4">Risk Appetite</h2>
      <p className="text-sm text-gray-500 mb-6">Define how your organisation approaches risk. This will guide risk assessments and treatment decisions.</p>
      <div className="space-y-4">
        <FormSelect label="Overall Risk Appetite" value={data.overall_appetite} onChange={v => update('overall_appetite', v)} options={[
          { value: 'conservative', label: 'Conservative - Minimise risk at all costs' },
          { value: 'moderate', label: 'Moderate - Balanced approach to risk' },
          { value: 'aggressive', label: 'Aggressive - Accept higher risk for greater reward' },
        ]} />
        <FormSelect label="Acceptable Risk Level" value={data.acceptable_risk_level} onChange={v => update('acceptable_risk_level', v)} options={[
          { value: 'low', label: 'Low - Only accept very low risks' },
          { value: 'medium', label: 'Medium - Accept moderate risks with mitigation' },
          { value: 'high', label: 'High - Accept higher risks if justified' },
        ]} />
        <FormSelect label="Risk Review Frequency" value={data.review_frequency} onChange={v => update('review_frequency', v)} options={[
          { value: 'monthly', label: 'Monthly' },
          { value: 'quarterly', label: 'Quarterly' },
          { value: 'annually', label: 'Annually' },
        ]} />
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Additional Notes</label>
          <textarea
            value={data.notes || ''}
            onChange={e => update('notes', e.target.value)}
            rows={3}
            className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500"
            placeholder="Any specific risk considerations for your organisation..."
          />
        </div>
      </div>
    </div>
  );
}

function Step6QuickAssessment({ data, onChange }: {
  data: Record<string, number>;
  onChange: (d: Record<string, number>) => void;
}) {
  const update = (key: string, value: number) => onChange({ ...data, [key]: value });

  const areas = [
    { key: 'security_maturity', label: 'Overall Security Maturity' },
    { key: 'policy_documentation', label: 'Policy Documentation' },
    { key: 'incident_response_ready', label: 'Incident Response Readiness' },
    { key: 'data_protection_level', label: 'Data Protection Level' },
    { key: 'access_control_maturity', label: 'Access Control Maturity' },
    { key: 'third_party_risk_mgmt', label: 'Third-Party Risk Management' },
    { key: 'business_continuity', label: 'Business Continuity' },
    { key: 'security_awareness', label: 'Security Awareness' },
  ];

  const levels = ['Ad-hoc', 'Repeatable', 'Defined', 'Managed', 'Optimised'];

  return (
    <div>
      <h2 className="text-lg font-semibold text-gray-900 mb-4">Quick Self-Assessment</h2>
      <p className="text-sm text-gray-500 mb-6">Rate your current maturity level in each area (1=Ad-hoc to 5=Optimised). This gives us a baseline for improvement tracking.</p>
      <div className="space-y-5">
        {areas.map(area => (
          <div key={area.key}>
            <div className="flex items-center justify-between mb-2">
              <label className="text-sm font-medium text-gray-700">{area.label}</label>
              <span className="text-xs text-indigo-600 font-medium">
                {data[area.key]}/5 &mdash; {levels[(data[area.key] || 1) - 1]}
              </span>
            </div>
            <input
              type="range" min={1} max={5} step={1}
              value={data[area.key] || 3}
              onChange={e => update(area.key, parseInt(e.target.value))}
              className="w-full h-2 bg-gray-200 rounded-lg appearance-none cursor-pointer accent-indigo-600"
            />
            <div className="flex justify-between text-xs text-gray-400 mt-1">
              {levels.map((l, i) => <span key={i}>{i + 1}</span>)}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

function Step7Complete({ progress, selectedFrameworks, invitations, recommendations }: {
  progress: OnboardingProgress | null;
  selectedFrameworks: string[];
  invitations: TeamInvitation[];
  recommendations: FrameworkRecommendation[];
}) {
  const fwNames = selectedFrameworks.map(id => {
    const rec = recommendations.find(r => r.framework_id === id);
    return rec?.framework_name || id;
  });

  const validInvitations = invitations.filter(i => i.email);

  return (
    <div>
      <h2 className="text-lg font-semibold text-gray-900 mb-4">Review & Complete</h2>
      <p className="text-sm text-gray-500 mb-6">Review your setup and click Complete to finalise your onboarding.</p>

      <div className="space-y-4">
        <SummaryCard title="Frameworks to Adopt" value={`${fwNames.length} selected`}>
          {fwNames.length > 0 ? (
            <ul className="text-sm text-gray-600 list-disc list-inside">
              {fwNames.map((name, i) => <li key={i}>{name}</li>)}
            </ul>
          ) : (
            <p className="text-sm text-gray-400">No frameworks selected</p>
          )}
        </SummaryCard>

        <SummaryCard title="Team Invitations" value={`${validInvitations.length} member${validInvitations.length !== 1 ? 's' : ''}`}>
          {validInvitations.length > 0 ? (
            <ul className="text-sm text-gray-600 list-disc list-inside">
              {validInvitations.map((inv, i) => <li key={i}>{inv.email} ({inv.role})</li>)}
            </ul>
          ) : (
            <p className="text-sm text-gray-400">No team members invited</p>
          )}
        </SummaryCard>

        <SummaryCard title="Steps Completed" value={`${progress?.completed_steps?.length || 0} of 7`}>
          <div className="flex gap-2 flex-wrap">
            {STEPS.map(s => (
              <span key={s.number} className={`text-xs px-2 py-1 rounded ${
                progress?.completed_steps?.includes(s.number) ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-500'
              }`}>
                {s.title}
              </span>
            ))}
          </div>
        </SummaryCard>
      </div>

      <div className="mt-6 p-4 bg-green-50 border border-green-200 rounded-lg">
        <p className="text-sm text-green-700">
          Clicking <strong>Complete Setup</strong> will adopt your selected frameworks, create control implementations,
          send team invitations, and set up your default risk matrix. This may take a few moments.
        </p>
      </div>
    </div>
  );
}

// ============================================================
// SHARED FORM COMPONENTS
// ============================================================

function FormField({ label, value, onChange, placeholder, type = 'text' }: {
  label: string; value: string; onChange: (v: string) => void; placeholder?: string; type?: string;
}) {
  return (
    <div>
      <label className="block text-sm font-medium text-gray-700 mb-1">{label}</label>
      <input type={type} value={value} onChange={e => onChange(e.target.value)} placeholder={placeholder}
        className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500" />
    </div>
  );
}

function FormSelect({ label, value, onChange, options }: {
  label: string; value: string; onChange: (v: string) => void;
  options: { value: string; label: string }[];
}) {
  return (
    <div>
      <label className="block text-sm font-medium text-gray-700 mb-1">{label}</label>
      <select value={value} onChange={e => onChange(e.target.value)}
        className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 bg-white">
        {options.map(o => <option key={o.value} value={o.value}>{o.label}</option>)}
      </select>
    </div>
  );
}

function SummaryCard({ title, value, children }: { title: string; value: string; children: React.ReactNode }) {
  return (
    <div className="p-4 rounded-lg border border-gray-200">
      <div className="flex items-center justify-between mb-2">
        <h3 className="text-sm font-medium text-gray-900">{title}</h3>
        <span className="text-xs text-gray-500">{value}</span>
      </div>
      {children}
    </div>
  );
}
