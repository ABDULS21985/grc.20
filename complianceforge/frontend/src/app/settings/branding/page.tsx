'use client';

import { useEffect, useState, useCallback, useRef } from 'react';
import api from '@/lib/api';

// ============================================================
// TYPES
// ============================================================

interface BrandingConfig {
  id: string;
  organization_id: string;
  product_name: string;
  tagline: string;
  company_name: string;
  support_email: string;
  support_url: string;
  privacy_policy_url: string;
  terms_of_service_url: string;
  logo_full_url: string;
  logo_icon_url: string;
  logo_dark_url: string;
  logo_light_url: string;
  favicon_url: string;
  email_logo_url: string;
  color_primary: string;
  color_primary_hover: string;
  color_secondary: string;
  color_secondary_hover: string;
  color_accent: string;
  color_background: string;
  color_surface: string;
  color_text_primary: string;
  color_text_secondary: string;
  color_border: string;
  color_success: string;
  color_warning: string;
  color_error: string;
  color_info: string;
  color_sidebar_bg: string;
  color_sidebar_text: string;
  font_family_heading: string;
  font_family_body: string;
  font_size_base: string;
  sidebar_style: string;
  corner_radius: string;
  density: string;
  custom_domain: string;
  domain_verification_status: string;
  domain_verified_at: string | null;
  ssl_status: string;
  custom_css: string;
  show_powered_by: boolean;
  show_help_widget: boolean;
  show_marketplace: boolean;
  show_knowledge_base: boolean;
}

interface DomainStatus {
  custom_domain: string;
  domain_verification_status: string;
  domain_verified_at: string | null;
  ssl_status: string;
  ssl_provisioned_at: string | null;
  ssl_expires_at: string | null;
  verification_token: string;
}

const TABS = ['Branding', 'Colours', 'Typography & Layout', 'Custom Domain', 'Custom CSS', 'Preview'];

const FONT_OPTIONS = [
  'Inter', 'Roboto', 'Open Sans', 'Lato', 'Montserrat', 'Poppins',
  'Source Sans Pro', 'Nunito', 'Raleway', 'Work Sans', 'DM Sans', 'IBM Plex Sans',
];

const LOGO_TYPES = [
  { key: 'full', label: 'Full Logo', desc: 'Used in the header and login page' },
  { key: 'icon', label: 'Icon Logo', desc: 'Used in collapsed sidebar and tabs' },
  { key: 'dark', label: 'Dark Background Logo', desc: 'Used on dark backgrounds' },
  { key: 'light', label: 'Light Background Logo', desc: 'Used on light backgrounds' },
  { key: 'favicon', label: 'Favicon', desc: 'Browser tab icon (recommended: 32x32px)' },
  { key: 'email', label: 'Email Logo', desc: 'Used in email notifications' },
];

const COLOUR_FIELDS: { key: string; label: string; group: string }[] = [
  { key: 'color_primary', label: 'Primary', group: 'Brand' },
  { key: 'color_primary_hover', label: 'Primary Hover', group: 'Brand' },
  { key: 'color_secondary', label: 'Secondary', group: 'Brand' },
  { key: 'color_secondary_hover', label: 'Secondary Hover', group: 'Brand' },
  { key: 'color_accent', label: 'Accent', group: 'Brand' },
  { key: 'color_background', label: 'Background', group: 'Surfaces' },
  { key: 'color_surface', label: 'Surface', group: 'Surfaces' },
  { key: 'color_text_primary', label: 'Text Primary', group: 'Text' },
  { key: 'color_text_secondary', label: 'Text Secondary', group: 'Text' },
  { key: 'color_border', label: 'Border', group: 'Text' },
  { key: 'color_success', label: 'Success', group: 'Status' },
  { key: 'color_warning', label: 'Warning', group: 'Status' },
  { key: 'color_error', label: 'Error', group: 'Status' },
  { key: 'color_info', label: 'Info', group: 'Status' },
  { key: 'color_sidebar_bg', label: 'Sidebar Background', group: 'Sidebar' },
  { key: 'color_sidebar_text', label: 'Sidebar Text', group: 'Sidebar' },
];

// ============================================================
// MAIN PAGE COMPONENT
// ============================================================

export default function BrandingSettingsPage() {
  const [tab, setTab] = useState('Branding');
  const [branding, setBranding] = useState<BrandingConfig | null>(null);
  const [draft, setDraft] = useState<Partial<BrandingConfig>>({});
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [previewCSS, setPreviewCSS] = useState('');
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

  const loadBranding = useCallback(async () => {
    try {
      const res = await api.getBranding();
      setBranding(res.data);
      setDraft({});
    } catch {
      setMessage({ type: 'error', text: 'Failed to load branding configuration.' });
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadBranding();
  }, [loadBranding]);

  const updateDraft = (field: string, value: string | boolean) => {
    setDraft(prev => ({ ...prev, [field]: value }));
  };

  const currentValue = (field: string): string => {
    if (field in draft) return draft[field as keyof BrandingConfig] as string;
    if (branding) return (branding as Record<string, unknown>)[field] as string || '';
    return '';
  };

  const currentBool = (field: string): boolean => {
    if (field in draft) return draft[field as keyof BrandingConfig] as boolean;
    if (branding) return (branding as Record<string, unknown>)[field] as boolean;
    return true;
  };

  const handleSave = async () => {
    if (Object.keys(draft).length === 0) {
      setMessage({ type: 'error', text: 'No changes to save.' });
      return;
    }
    setSaving(true);
    setMessage(null);
    try {
      const res = await api.updateBranding(draft);
      setBranding(res.data);
      setDraft({});
      setMessage({ type: 'success', text: 'Branding updated successfully.' });
    } catch (err: unknown) {
      const errorMsg = err instanceof Error ? err.message : 'Failed to save branding.';
      setMessage({ type: 'error', text: errorMsg });
    } finally {
      setSaving(false);
    }
  };

  const handleReset = () => {
    setDraft({});
    setMessage({ type: 'success', text: 'Changes discarded. Showing saved branding.' });
  };

  const handlePreview = async () => {
    try {
      const merged = { ...branding, ...draft };
      const res = await api.previewBranding(merged);
      setPreviewCSS(res.data?.css || '');
      setTab('Preview');
    } catch {
      setMessage({ type: 'error', text: 'Failed to generate preview.' });
    }
  };

  const hasDraftChanges = Object.keys(draft).length > 0;

  if (loading) return <p className="text-gray-500 p-6">Loading branding settings...</p>;

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Branding & Theming</h1>
          <p className="text-sm text-gray-500 mt-1">Customise the look and feel of your ComplianceForge instance</p>
        </div>
        <div className="flex gap-2">
          {hasDraftChanges && (
            <button onClick={handleReset} className="btn-secondary">Discard Changes</button>
          )}
          <button onClick={handlePreview} className="btn-secondary">Preview</button>
          <button onClick={handleSave} disabled={saving || !hasDraftChanges} className="btn-primary">
            {saving ? 'Saving...' : 'Save Changes'}
          </button>
        </div>
      </div>

      {/* Status message */}
      {message && (
        <div className={`mb-4 p-3 rounded-lg text-sm ${message.type === 'success' ? 'bg-green-50 text-green-700 border border-green-200' : 'bg-red-50 text-red-700 border border-red-200'}`}>
          {message.text}
        </div>
      )}

      {/* Draft indicator */}
      {hasDraftChanges && (
        <div className="mb-4 p-3 rounded-lg bg-yellow-50 border border-yellow-200 text-sm text-yellow-700">
          You have unsaved changes ({Object.keys(draft).length} field{Object.keys(draft).length > 1 ? 's' : ''} modified).
        </div>
      )}

      {/* Tabs */}
      <div className="flex gap-1 border-b border-gray-200 mb-6">
        {TABS.map(t => (
          <button key={t} onClick={() => setTab(t)}
            className={`px-4 py-2.5 text-sm font-medium border-b-2 transition-colors ${tab === t ? 'border-indigo-600 text-indigo-600' : 'border-transparent text-gray-500 hover:text-gray-700'}`}>
            {t}
          </button>
        ))}
      </div>

      {tab === 'Branding' && (
        <BrandingTab
          currentValue={currentValue}
          currentBool={currentBool}
          updateDraft={updateDraft}
          orgId={branding?.organization_id || ''}
          onLogoChange={loadBranding}
        />
      )}
      {tab === 'Colours' && (
        <ColoursTab currentValue={currentValue} updateDraft={updateDraft} />
      )}
      {tab === 'Typography & Layout' && (
        <TypographyLayoutTab currentValue={currentValue} updateDraft={updateDraft} />
      )}
      {tab === 'Custom Domain' && (
        <CustomDomainTab />
      )}
      {tab === 'Custom CSS' && (
        <CustomCSSTab currentValue={currentValue} updateDraft={updateDraft} />
      )}
      {tab === 'Preview' && (
        <PreviewTab branding={branding} draft={draft} previewCSS={previewCSS} />
      )}
    </div>
  );
}

// ============================================================
// BRANDING TAB (Identity + Logos + Feature Flags)
// ============================================================

function BrandingTab({
  currentValue, currentBool, updateDraft, orgId, onLogoChange
}: {
  currentValue: (f: string) => string;
  currentBool: (f: string) => boolean;
  updateDraft: (f: string, v: string | boolean) => void;
  orgId: string;
  onLogoChange: () => void;
}) {
  return (
    <div className="space-y-6">
      {/* Identity */}
      <div className="card">
        <h2 className="text-lg font-semibold text-gray-900 mb-4">Identity</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <TextInput label="Product Name" value={currentValue('product_name')} onChange={v => updateDraft('product_name', v)} />
          <TextInput label="Tagline" value={currentValue('tagline')} onChange={v => updateDraft('tagline', v)} />
          <TextInput label="Company Name" value={currentValue('company_name')} onChange={v => updateDraft('company_name', v)} />
          <TextInput label="Support Email" type="email" value={currentValue('support_email')} onChange={v => updateDraft('support_email', v)} />
          <TextInput label="Support URL" type="url" value={currentValue('support_url')} onChange={v => updateDraft('support_url', v)} />
          <TextInput label="Privacy Policy URL" type="url" value={currentValue('privacy_policy_url')} onChange={v => updateDraft('privacy_policy_url', v)} />
          <TextInput label="Terms of Service URL" type="url" value={currentValue('terms_of_service_url')} onChange={v => updateDraft('terms_of_service_url', v)} />
        </div>
      </div>

      {/* Logos */}
      <div className="card">
        <h2 className="text-lg font-semibold text-gray-900 mb-4">Logos</h2>
        <p className="text-sm text-gray-500 mb-4">Upload SVG, PNG, or JPEG files. Maximum 5MB per file.</p>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {LOGO_TYPES.map(logo => (
            <LogoUploadCard
              key={logo.key}
              logoType={logo.key}
              label={logo.label}
              description={logo.desc}
              currentUrl={currentValue(`logo_${logo.key === 'favicon' ? 'favicon' : logo.key === 'email' ? 'email_logo' : logo.key}_url`)}
              onUpload={onLogoChange}
            />
          ))}
        </div>
      </div>

      {/* Feature Flags */}
      <div className="card">
        <h2 className="text-lg font-semibold text-gray-900 mb-4">Feature Visibility</h2>
        <div className="space-y-3">
          <ToggleSwitch
            label="Show 'Powered by ComplianceForge'"
            description="Mandatory for non-Enterprise plans"
            checked={currentBool('show_powered_by')}
            onChange={v => updateDraft('show_powered_by', v)}
          />
          <ToggleSwitch
            label="Show Help Widget"
            description="In-app help and documentation widget"
            checked={currentBool('show_help_widget')}
            onChange={v => updateDraft('show_help_widget', v)}
          />
          <ToggleSwitch
            label="Show Marketplace"
            description="Access to the compliance package marketplace"
            checked={currentBool('show_marketplace')}
            onChange={v => updateDraft('show_marketplace', v)}
          />
          <ToggleSwitch
            label="Show Knowledge Base"
            description="Access to the knowledge base and guides"
            checked={currentBool('show_knowledge_base')}
            onChange={v => updateDraft('show_knowledge_base', v)}
          />
        </div>
      </div>
    </div>
  );
}

// ============================================================
// COLOURS TAB
// ============================================================

function ColoursTab({
  currentValue, updateDraft
}: {
  currentValue: (f: string) => string;
  updateDraft: (f: string, v: string) => void;
}) {
  const groups = ['Brand', 'Surfaces', 'Text', 'Status', 'Sidebar'];

  return (
    <div className="space-y-6">
      {groups.map(group => (
        <div key={group} className="card">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">{group} Colours</h2>
          <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
            {COLOUR_FIELDS.filter(f => f.group === group).map(field => (
              <ColourPicker
                key={field.key}
                label={field.label}
                value={currentValue(field.key)}
                onChange={v => updateDraft(field.key, v)}
              />
            ))}
          </div>
        </div>
      ))}
    </div>
  );
}

// ============================================================
// TYPOGRAPHY & LAYOUT TAB
// ============================================================

function TypographyLayoutTab({
  currentValue, updateDraft
}: {
  currentValue: (f: string) => string;
  updateDraft: (f: string, v: string) => void;
}) {
  return (
    <div className="space-y-6">
      {/* Typography */}
      <div className="card">
        <h2 className="text-lg font-semibold text-gray-900 mb-4">Typography</h2>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <SelectInput
            label="Heading Font"
            value={currentValue('font_family_heading')}
            options={FONT_OPTIONS.map(f => ({ value: f, label: f }))}
            onChange={v => updateDraft('font_family_heading', v)}
          />
          <SelectInput
            label="Body Font"
            value={currentValue('font_family_body')}
            options={FONT_OPTIONS.map(f => ({ value: f, label: f }))}
            onChange={v => updateDraft('font_family_body', v)}
          />
          <SelectInput
            label="Base Font Size"
            value={currentValue('font_size_base')}
            options={[
              { value: '12px', label: '12px (Small)' },
              { value: '13px', label: '13px' },
              { value: '14px', label: '14px (Default)' },
              { value: '15px', label: '15px' },
              { value: '16px', label: '16px (Large)' },
            ]}
            onChange={v => updateDraft('font_size_base', v)}
          />
        </div>
      </div>

      {/* Layout */}
      <div className="card">
        <h2 className="text-lg font-semibold text-gray-900 mb-4">Layout Options</h2>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <SelectInput
            label="Sidebar Style"
            value={currentValue('sidebar_style')}
            options={[
              { value: 'light', label: 'Light' },
              { value: 'dark', label: 'Dark' },
              { value: 'branded', label: 'Branded (uses primary colour)' },
            ]}
            onChange={v => updateDraft('sidebar_style', v)}
          />
          <SelectInput
            label="Corner Radius"
            value={currentValue('corner_radius')}
            options={[
              { value: 'none', label: 'None (sharp corners)' },
              { value: 'small', label: 'Small (4px)' },
              { value: 'medium', label: 'Medium (8px)' },
              { value: 'large', label: 'Large (12px)' },
              { value: 'full', label: 'Full (pill shape)' },
            ]}
            onChange={v => updateDraft('corner_radius', v)}
          />
          <SelectInput
            label="UI Density"
            value={currentValue('density')}
            options={[
              { value: 'compact', label: 'Compact' },
              { value: 'comfortable', label: 'Comfortable (Default)' },
              { value: 'spacious', label: 'Spacious' },
            ]}
            onChange={v => updateDraft('density', v)}
          />
        </div>
      </div>

      {/* Preview Samples */}
      <div className="card">
        <h2 className="text-lg font-semibold text-gray-900 mb-4">Preview Samples</h2>
        <div className="flex gap-3 flex-wrap">
          <button
            className="px-4 py-2 text-white text-sm font-medium transition-colors"
            style={{
              backgroundColor: currentValue('color_primary'),
              borderRadius: currentValue('corner_radius') === 'none' ? '0' :
                currentValue('corner_radius') === 'small' ? '4px' :
                currentValue('corner_radius') === 'large' ? '12px' :
                currentValue('corner_radius') === 'full' ? '9999px' : '8px',
              fontFamily: currentValue('font_family_body'),
            }}
          >
            Primary Button
          </button>
          <button
            className="px-4 py-2 text-white text-sm font-medium transition-colors"
            style={{
              backgroundColor: currentValue('color_secondary'),
              borderRadius: currentValue('corner_radius') === 'none' ? '0' :
                currentValue('corner_radius') === 'small' ? '4px' :
                currentValue('corner_radius') === 'large' ? '12px' :
                currentValue('corner_radius') === 'full' ? '9999px' : '8px',
              fontFamily: currentValue('font_family_body'),
            }}
          >
            Secondary Button
          </button>
          <span
            className="px-3 py-1 text-xs font-medium"
            style={{
              backgroundColor: currentValue('color_success') + '20',
              color: currentValue('color_success'),
              borderRadius: currentValue('corner_radius') === 'none' ? '0' :
                currentValue('corner_radius') === 'small' ? '4px' :
                currentValue('corner_radius') === 'full' ? '9999px' : '6px',
            }}
          >
            Success Badge
          </span>
          <span
            className="px-3 py-1 text-xs font-medium"
            style={{
              backgroundColor: currentValue('color_warning') + '20',
              color: currentValue('color_warning'),
              borderRadius: currentValue('corner_radius') === 'none' ? '0' :
                currentValue('corner_radius') === 'small' ? '4px' :
                currentValue('corner_radius') === 'full' ? '9999px' : '6px',
            }}
          >
            Warning Badge
          </span>
          <span
            className="px-3 py-1 text-xs font-medium"
            style={{
              backgroundColor: currentValue('color_error') + '20',
              color: currentValue('color_error'),
              borderRadius: currentValue('corner_radius') === 'none' ? '0' :
                currentValue('corner_radius') === 'small' ? '4px' :
                currentValue('corner_radius') === 'full' ? '9999px' : '6px',
            }}
          >
            Error Badge
          </span>
        </div>
      </div>
    </div>
  );
}

// ============================================================
// CUSTOM DOMAIN TAB (4-step wizard)
// ============================================================

function CustomDomainTab() {
  const [step, setStep] = useState(1);
  const [domain, setDomain] = useState('');
  const [domainStatus, setDomainStatus] = useState<DomainStatus | null>(null);
  const [verifying, setVerifying] = useState(false);
  const [loading, setLoading] = useState(true);
  const [msg, setMsg] = useState('');

  useEffect(() => {
    api.getDomainStatus()
      .then(res => {
        const ds = res.data;
        setDomainStatus(ds);
        if (ds?.custom_domain) {
          setDomain(ds.custom_domain);
          if (ds.domain_verification_status === 'verified' && ds.ssl_status === 'active') {
            setStep(4);
          } else if (ds.domain_verification_status === 'verified') {
            setStep(3);
          } else if (ds.custom_domain) {
            setStep(2);
          }
        }
      })
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  const handleVerify = async () => {
    if (!domain.trim()) return;
    setVerifying(true);
    setMsg('');
    try {
      const res = await api.verifyDomain({ domain: domain.trim() });
      setMsg(res.data?.message || 'Verification initiated.');
      const status = res.data?.status;
      if (status === 'verified') {
        setStep(3);
      } else {
        setStep(2);
      }
      // Refresh domain status
      const statusRes = await api.getDomainStatus();
      setDomainStatus(statusRes.data);
    } catch (err: unknown) {
      const errorMsg = err instanceof Error ? err.message : 'Verification failed.';
      setMsg(errorMsg);
    } finally {
      setVerifying(false);
    }
  };

  if (loading) return <p className="text-gray-500">Loading domain status...</p>;

  return (
    <div className="space-y-6">
      {/* Progress Steps */}
      <div className="card">
        <div className="flex items-center gap-4 mb-6">
          {[
            { n: 1, label: 'Enter Domain' },
            { n: 2, label: 'Configure DNS' },
            { n: 3, label: 'SSL Certificate' },
            { n: 4, label: 'Active' },
          ].map(s => (
            <div key={s.n} className="flex items-center gap-2">
              <div className={`w-8 h-8 rounded-full flex items-center justify-center text-sm font-medium ${step >= s.n ? 'bg-indigo-600 text-white' : 'bg-gray-200 text-gray-500'}`}>
                {step > s.n ? '\u2713' : s.n}
              </div>
              <span className={`text-sm ${step >= s.n ? 'text-gray-900 font-medium' : 'text-gray-400'}`}>{s.label}</span>
              {s.n < 4 && <div className={`w-8 h-0.5 ${step > s.n ? 'bg-indigo-600' : 'bg-gray-200'}`} />}
            </div>
          ))}
        </div>

        {/* Step 1: Enter Domain */}
        {step === 1 && (
          <div>
            <h3 className="text-md font-semibold text-gray-900 mb-2">Enter your custom domain</h3>
            <p className="text-sm text-gray-500 mb-4">
              Enter the domain you want to use to access your ComplianceForge instance (e.g., grc.yourdomain.com).
            </p>
            <div className="flex gap-3 items-end">
              <div className="flex-1">
                <label className="block text-sm font-medium text-gray-700 mb-1">Domain</label>
                <input
                  type="text"
                  value={domain}
                  onChange={e => setDomain(e.target.value)}
                  placeholder="grc.yourdomain.com"
                  className="input w-full"
                />
              </div>
              <button onClick={handleVerify} disabled={verifying || !domain.trim()} className="btn-primary">
                {verifying ? 'Verifying...' : 'Verify Domain'}
              </button>
            </div>
          </div>
        )}

        {/* Step 2: DNS Configuration */}
        {step === 2 && domainStatus && (
          <div>
            <h3 className="text-md font-semibold text-gray-900 mb-2">Configure DNS Records</h3>
            <p className="text-sm text-gray-500 mb-4">
              Add one of the following DNS records to verify domain ownership:
            </p>
            <div className="space-y-4">
              <div className="bg-gray-50 p-4 rounded-lg">
                <p className="text-sm font-medium text-gray-700 mb-1">Option 1: CNAME Record</p>
                <div className="font-mono text-sm bg-white p-3 rounded border border-gray-200">
                  <span className="text-gray-500">Host:</span> {domain}<br />
                  <span className="text-gray-500">Value:</span> app.complianceforge.io
                </div>
              </div>
              <div className="bg-gray-50 p-4 rounded-lg">
                <p className="text-sm font-medium text-gray-700 mb-1">Option 2: TXT Record</p>
                <div className="font-mono text-sm bg-white p-3 rounded border border-gray-200">
                  <span className="text-gray-500">Host:</span> {domain}<br />
                  <span className="text-gray-500">Value:</span> cf-verify={domainStatus.verification_token}
                </div>
              </div>
            </div>
            <div className="flex gap-3 mt-4">
              <button onClick={handleVerify} disabled={verifying} className="btn-primary">
                {verifying ? 'Checking...' : 'Check DNS Records'}
              </button>
              <button onClick={() => setStep(1)} className="btn-secondary">Change Domain</button>
            </div>
          </div>
        )}

        {/* Step 3: SSL Provisioning */}
        {step === 3 && (
          <div>
            <h3 className="text-md font-semibold text-gray-900 mb-2">SSL Certificate</h3>
            <p className="text-sm text-gray-500 mb-4">
              Domain verified successfully. An SSL certificate is being provisioned for <strong>{domain}</strong>.
            </p>
            <div className="bg-blue-50 p-4 rounded-lg border border-blue-200">
              <p className="text-sm text-blue-700">
                SSL Status: <strong>{domainStatus?.ssl_status || 'provisioning'}</strong>
              </p>
              <p className="text-sm text-blue-600 mt-1">
                This typically takes 5-15 minutes. The page will update automatically when ready.
              </p>
            </div>
            <button onClick={handleVerify} className="btn-secondary mt-4">Refresh Status</button>
          </div>
        )}

        {/* Step 4: Active */}
        {step === 4 && (
          <div>
            <h3 className="text-md font-semibold text-gray-900 mb-2">Domain Active</h3>
            <div className="bg-green-50 p-4 rounded-lg border border-green-200">
              <p className="text-sm text-green-700">
                Your custom domain <strong>{domain}</strong> is active and secured with SSL.
              </p>
              {domainStatus?.ssl_expires_at && (
                <p className="text-sm text-green-600 mt-1">
                  SSL certificate expires: {new Date(domainStatus.ssl_expires_at).toLocaleDateString('en-GB')}
                </p>
              )}
            </div>
            <button onClick={() => setStep(1)} className="btn-secondary mt-4">Change Domain</button>
          </div>
        )}

        {msg && <p className="text-sm text-gray-600 mt-3">{msg}</p>}
      </div>
    </div>
  );
}

// ============================================================
// CUSTOM CSS TAB
// ============================================================

function CustomCSSTab({
  currentValue, updateDraft
}: {
  currentValue: (f: string) => string;
  updateDraft: (f: string, v: string) => void;
}) {
  return (
    <div className="space-y-6">
      <div className="card">
        <h2 className="text-lg font-semibold text-gray-900 mb-2">Custom CSS</h2>
        <p className="text-sm text-gray-500 mb-4">
          Add custom CSS to further customise the appearance. Only safe properties are allowed
          (color, background, font, border, margin, padding, display, etc.).
          JavaScript, data: URIs, and @import are blocked.
        </p>
        <textarea
          value={currentValue('custom_css')}
          onChange={e => updateDraft('custom_css', e.target.value)}
          className="w-full h-80 font-mono text-sm border border-gray-300 rounded-lg p-4 focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
          placeholder={`/* Custom CSS */\n.card {\n  border-color: #your-color;\n  border-radius: 12px;\n}`}
          spellCheck={false}
        />
        <div className="mt-3 flex items-center gap-4">
          <div className="text-xs text-gray-400">
            {currentValue('custom_css').length} characters
          </div>
          <button
            onClick={() => updateDraft('custom_css', '')}
            className="text-xs text-red-600 hover:underline"
          >
            Clear CSS
          </button>
        </div>
      </div>

      <div className="card bg-amber-50 border-amber-200">
        <h3 className="text-sm font-semibold text-amber-800 mb-2">Allowed CSS Properties</h3>
        <p className="text-xs text-amber-700">
          color, background, background-color, font, font-family, font-size, font-weight,
          text-align, text-decoration, border, border-radius, border-color,
          margin, padding, width, height, display, visibility, opacity,
          flex, flex-direction, gap, align-items, justify-content, box-shadow, transition, transform.
          CSS custom properties (--*) are also allowed.
        </p>
      </div>
    </div>
  );
}

// ============================================================
// PREVIEW TAB
// ============================================================

function PreviewTab({
  branding, draft, previewCSS
}: {
  branding: BrandingConfig | null;
  draft: Partial<BrandingConfig>;
  previewCSS: string;
}) {
  const merged = { ...branding, ...draft } as BrandingConfig;

  return (
    <div className="space-y-6">
      <div className="card">
        <h2 className="text-lg font-semibold text-gray-900 mb-4">Live Preview</h2>
        <p className="text-sm text-gray-500 mb-4">This is a preview of your branding changes. Save to apply them.</p>

        {/* Mini app preview */}
        <div className="border border-gray-200 rounded-lg overflow-hidden" style={{ fontFamily: merged.font_family_body }}>
          {/* Header */}
          <div className="flex items-center justify-between px-4 py-3" style={{ backgroundColor: merged.color_surface, borderBottom: `1px solid ${merged.color_border}` }}>
            <div className="flex items-center gap-3">
              {merged.logo_full_url ? (
                <img src={merged.logo_full_url} alt="Logo" className="h-8" />
              ) : (
                <div className="h-8 w-8 rounded flex items-center justify-center text-white text-xs font-bold" style={{ backgroundColor: merged.color_primary }}>
                  {(merged.product_name || 'CF')[0]}
                </div>
              )}
              <span className="font-semibold" style={{ color: merged.color_text_primary, fontFamily: merged.font_family_heading }}>
                {merged.product_name || 'ComplianceForge'}
              </span>
            </div>
            <span className="text-xs" style={{ color: merged.color_text_secondary }}>{merged.tagline}</span>
          </div>

          <div className="flex" style={{ minHeight: 300 }}>
            {/* Sidebar */}
            <div className="w-48 p-3 space-y-1" style={{ backgroundColor: merged.color_sidebar_bg }}>
              {['Dashboard', 'Frameworks', 'Risks', 'Policies', 'Settings'].map((item, i) => (
                <div key={item} className={`px-3 py-2 text-sm rounded ${i === 0 ? 'font-medium' : ''}`}
                  style={{
                    color: merged.color_sidebar_text,
                    backgroundColor: i === 0 ? merged.color_primary + '30' : 'transparent',
                    borderRadius: merged.corner_radius === 'none' ? '0' :
                      merged.corner_radius === 'small' ? '4px' :
                      merged.corner_radius === 'large' ? '12px' :
                      merged.corner_radius === 'full' ? '9999px' : '6px',
                  }}>
                  {item}
                </div>
              ))}
            </div>

            {/* Content */}
            <div className="flex-1 p-4" style={{ backgroundColor: merged.color_background }}>
              <h3 className="text-lg font-semibold mb-3" style={{ color: merged.color_text_primary, fontFamily: merged.font_family_heading }}>
                Dashboard
              </h3>
              <div className="grid grid-cols-3 gap-3 mb-4">
                {[
                  { label: 'Compliance', value: '87%', color: merged.color_success },
                  { label: 'Open Risks', value: '12', color: merged.color_warning },
                  { label: 'Incidents', value: '3', color: merged.color_error },
                ].map(stat => (
                  <div key={stat.label} className="p-3"
                    style={{
                      backgroundColor: merged.color_surface,
                      border: `1px solid ${merged.color_border}`,
                      borderRadius: merged.corner_radius === 'none' ? '0' :
                        merged.corner_radius === 'small' ? '4px' :
                        merged.corner_radius === 'large' ? '12px' :
                        merged.corner_radius === 'full' ? '16px' : '8px',
                    }}>
                    <p className="text-xs" style={{ color: merged.color_text_secondary }}>{stat.label}</p>
                    <p className="text-xl font-bold mt-1" style={{ color: stat.color }}>{stat.value}</p>
                  </div>
                ))}
              </div>
              <div className="flex gap-2">
                <button className="px-4 py-2 text-white text-sm font-medium"
                  style={{
                    backgroundColor: merged.color_primary,
                    borderRadius: merged.corner_radius === 'none' ? '0' :
                      merged.corner_radius === 'small' ? '4px' :
                      merged.corner_radius === 'large' ? '12px' :
                      merged.corner_radius === 'full' ? '9999px' : '8px',
                  }}>
                  Primary Action
                </button>
                <button className="px-4 py-2 text-sm font-medium"
                  style={{
                    border: `1px solid ${merged.color_border}`,
                    color: merged.color_text_primary,
                    backgroundColor: merged.color_surface,
                    borderRadius: merged.corner_radius === 'none' ? '0' :
                      merged.corner_radius === 'small' ? '4px' :
                      merged.corner_radius === 'large' ? '12px' :
                      merged.corner_radius === 'full' ? '9999px' : '8px',
                  }}>
                  Secondary
                </button>
              </div>
            </div>
          </div>

          {/* Footer */}
          {merged.show_powered_by && (
            <div className="text-center py-2 text-xs" style={{ color: merged.color_text_secondary, borderTop: `1px solid ${merged.color_border}` }}>
              Powered by ComplianceForge
            </div>
          )}
        </div>
      </div>

      {/* Generated CSS */}
      {previewCSS && (
        <div className="card">
          <h3 className="text-md font-semibold text-gray-900 mb-2">Generated CSS Variables</h3>
          <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg text-xs overflow-x-auto max-h-64">
            {previewCSS}
          </pre>
        </div>
      )}
    </div>
  );
}

// ============================================================
// SHARED INPUT COMPONENTS
// ============================================================

function TextInput({
  label, value, onChange, type = 'text', placeholder
}: {
  label: string; value: string; onChange: (v: string) => void;
  type?: string; placeholder?: string;
}) {
  return (
    <div>
      <label className="block text-sm font-medium text-gray-700 mb-1">{label}</label>
      <input
        type={type}
        value={value}
        onChange={e => onChange(e.target.value)}
        placeholder={placeholder}
        className="input w-full"
      />
    </div>
  );
}

function SelectInput({
  label, value, options, onChange
}: {
  label: string; value: string;
  options: { value: string; label: string }[];
  onChange: (v: string) => void;
}) {
  return (
    <div>
      <label className="block text-sm font-medium text-gray-700 mb-1">{label}</label>
      <select
        value={value}
        onChange={e => onChange(e.target.value)}
        className="input w-full"
      >
        {options.map(o => (
          <option key={o.value} value={o.value}>{o.label}</option>
        ))}
      </select>
    </div>
  );
}

function ColourPicker({
  label, value, onChange
}: {
  label: string; value: string; onChange: (v: string) => void;
}) {
  return (
    <div>
      <label className="block text-sm font-medium text-gray-700 mb-1">{label}</label>
      <div className="flex items-center gap-2">
        <input
          type="color"
          value={value || '#000000'}
          onChange={e => onChange(e.target.value.toUpperCase())}
          className="w-10 h-10 rounded border border-gray-300 cursor-pointer p-0.5"
        />
        <input
          type="text"
          value={value}
          onChange={e => onChange(e.target.value)}
          placeholder="#FFFFFF"
          className="input flex-1 font-mono text-sm"
          maxLength={7}
        />
      </div>
    </div>
  );
}

function ToggleSwitch({
  label, description, checked, onChange
}: {
  label: string; description?: string; checked: boolean;
  onChange: (v: boolean) => void;
}) {
  return (
    <div className="flex items-center justify-between py-2">
      <div>
        <p className="text-sm font-medium text-gray-700">{label}</p>
        {description && <p className="text-xs text-gray-400">{description}</p>}
      </div>
      <button
        role="switch"
        aria-checked={checked}
        onClick={() => onChange(!checked)}
        className={`relative w-11 h-6 rounded-full transition-colors ${checked ? 'bg-indigo-600' : 'bg-gray-300'}`}
      >
        <span className={`absolute top-0.5 left-0.5 w-5 h-5 rounded-full bg-white shadow transition-transform ${checked ? 'translate-x-5' : ''}`} />
      </button>
    </div>
  );
}

function LogoUploadCard({
  logoType, label, description, currentUrl, onUpload
}: {
  logoType: string; label: string; description: string;
  currentUrl: string; onUpload: () => void;
}) {
  const fileRef = useRef<HTMLInputElement>(null);
  const [uploading, setUploading] = useState(false);

  const handleUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    setUploading(true);
    try {
      await api.uploadLogo(logoType, file);
      onUpload();
    } catch {
      alert('Failed to upload logo.');
    } finally {
      setUploading(false);
      if (fileRef.current) fileRef.current.value = '';
    }
  };

  const handleDelete = async () => {
    if (!confirm('Remove this logo?')) return;
    try {
      await api.deleteLogo(logoType);
      onUpload();
    } catch {
      alert('Failed to remove logo.');
    }
  };

  return (
    <div className="border border-gray-200 rounded-lg p-4">
      <p className="text-sm font-medium text-gray-700">{label}</p>
      <p className="text-xs text-gray-400 mb-3">{description}</p>
      <div className="h-16 flex items-center justify-center bg-gray-50 rounded mb-3 border border-dashed border-gray-300">
        {currentUrl ? (
          <img src={currentUrl} alt={label} className="max-h-14 max-w-full object-contain" />
        ) : (
          <span className="text-xs text-gray-400">No logo uploaded</span>
        )}
      </div>
      <div className="flex gap-2">
        <input type="file" ref={fileRef} onChange={handleUpload} accept=".svg,.png,.jpg,.jpeg" className="hidden" />
        <button onClick={() => fileRef.current?.click()} disabled={uploading} className="text-xs text-indigo-600 hover:underline">
          {uploading ? 'Uploading...' : 'Upload'}
        </button>
        {currentUrl && (
          <button onClick={handleDelete} className="text-xs text-red-600 hover:underline">Remove</button>
        )}
      </div>
    </div>
  );
}
