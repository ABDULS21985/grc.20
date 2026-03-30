'use client';

import { useEffect, useState, useCallback } from 'react';
import api from '@/lib/api';

// ============================================================
// TYPES
// ============================================================

interface ResidencyConfig {
  organization_id: string;
  data_region: string;
  data_residency_enforced: boolean;
  display_name: string;
  description: string;
  allowed_countries: string[];
  blocked_countries: string[];
  allowed_cloud_regions: Record<string, unknown>;
  primary_cloud_region: string;
  failover_cloud_region: string;
  compliance_frameworks: string[];
  legal_basis: string;
  data_protection_authority: string;
  dpa_contact_url: string;
  gdpr_adequate_countries: string[];
  enforcement_mode: string;
  allow_cross_region_search: boolean;
  allow_cross_region_backup: boolean;
}

interface StorageLocation {
  type: string;
  label: string;
  region: string;
  provider: string;
  status: string;
  within_jurisdiction: boolean;
}

interface AuditEntry {
  id: string;
  organization_id: string;
  action: string;
  user_email: string;
  ip_address: string;
  source_region: string;
  destination_region: string;
  source_country: string;
  destination_country: string;
  resource_type: string;
  allowed: boolean;
  blocked_reason: string;
  vendor_name: string;
  transfer_mechanism: string;
  created_at: string;
}

interface ResidencyDashboard {
  current_region: string;
  region_display_name: string;
  enforced: boolean;
  enforcement_mode: string;
  compliance_status: string;
  compliance_frameworks: string[];
  storage_locations: StorageLocation[];
  recent_blocked: AuditEntry[];
  total_blocked_count: number;
  last_audit_entry: string | null;
  allowed_cloud_regions: Record<string, unknown>;
}

interface DataRegion {
  id: string;
  region: string;
  display_name: string;
  description: string;
  compliance_frameworks: string[];
  enforcement_mode: string;
  is_active: boolean;
}

interface ValidationResult {
  allowed: boolean;
  source_region: string;
  destination_region: string;
  reason: string;
  required_safeguards: string[];
  legal_basis: string;
}

interface TransferValidation {
  allowed: boolean;
  destination_country: string;
  is_gdpr_adequate: boolean;
  requires_additional_safeguards: boolean;
  transfer_mechanisms: string[];
  required_safeguards: string[];
  reason: string;
  data_protection_authority: string;
}

// ============================================================
// CONSTANTS
// ============================================================

const TABS = ['Overview', 'Storage', 'Audit Log', 'Export Validation', 'Transfer Validation'];

const ACTION_LABELS: Record<string, string> = {
  data_access: 'Data Access',
  data_export: 'Data Export',
  data_transfer: 'Data Transfer',
  cross_region_blocked: 'Cross-Region Blocked',
  config_change: 'Config Change',
  region_migration: 'Region Migration',
};

const STATUS_COLOURS: Record<string, string> = {
  compliant: 'bg-green-100 text-green-800',
  monitoring: 'bg-yellow-100 text-yellow-800',
  not_enforced: 'bg-gray-100 text-gray-700',
};

const COUNTRY_NAMES: Record<string, string> = {
  AD: 'Andorra', AR: 'Argentina', AT: 'Austria', AU: 'Australia',
  BE: 'Belgium', BG: 'Bulgaria', BR: 'Brazil', CA: 'Canada',
  CH: 'Switzerland', CN: 'China', CY: 'Cyprus', CZ: 'Czechia',
  DE: 'Germany', DK: 'Denmark', EE: 'Estonia', ES: 'Spain',
  FI: 'Finland', FO: 'Faroe Islands', FR: 'France', GB: 'United Kingdom',
  GG: 'Guernsey', GR: 'Greece', HR: 'Croatia', HU: 'Hungary',
  IE: 'Ireland', IL: 'Israel', IM: 'Isle of Man', IN: 'India',
  IS: 'Iceland', IT: 'Italy', JE: 'Jersey', JP: 'Japan',
  KR: 'South Korea', LI: 'Liechtenstein', LT: 'Lithuania', LU: 'Luxembourg',
  LV: 'Latvia', MT: 'Malta', MX: 'Mexico', NL: 'Netherlands',
  NO: 'Norway', NZ: 'New Zealand', PL: 'Poland', PT: 'Portugal',
  RO: 'Romania', RU: 'Russia', SE: 'Sweden', SI: 'Slovenia',
  SK: 'Slovakia', US: 'United States', UY: 'Uruguay', ZA: 'South Africa',
};

// ============================================================
// MAIN PAGE COMPONENT
// ============================================================

export default function DataResidencyPage() {
  const [tab, setTab] = useState('Overview');
  const [dashboard, setDashboard] = useState<ResidencyDashboard | null>(null);
  const [config, setConfig] = useState<ResidencyConfig | null>(null);
  const [regions, setRegions] = useState<DataRegion[]>([]);
  const [auditLog, setAuditLog] = useState<AuditEntry[]>([]);
  const [auditTotal, setAuditTotal] = useState(0);
  const [auditPage, setAuditPage] = useState(1);
  const [auditFilter, setAuditFilter] = useState<{ action: string; allowed: string }>({ action: '', allowed: '' });
  const [loading, setLoading] = useState(true);
  const [message, setMessage] = useState<{ type: 'success' | 'error' | 'info'; text: string } | null>(null);

  // Export validation
  const [exportRegion, setExportRegion] = useState('');
  const [exportResult, setExportResult] = useState<ValidationResult | null>(null);
  const [exportValidating, setExportValidating] = useState(false);

  // Transfer validation
  const [transferCountry, setTransferCountry] = useState('');
  const [transferResult, setTransferResult] = useState<TransferValidation | null>(null);
  const [transferValidating, setTransferValidating] = useState(false);

  // ============================================================
  // DATA LOADING
  // ============================================================

  const loadDashboard = useCallback(async () => {
    try {
      const res = await api.get('/residency/status');
      setDashboard(res.data);
    } catch {
      setMessage({ type: 'error', text: 'Failed to load residency dashboard.' });
    }
  }, []);

  const loadConfig = useCallback(async () => {
    try {
      const res = await api.get('/residency/config');
      setConfig(res.data);
    } catch {
      setMessage({ type: 'error', text: 'Failed to load residency configuration.' });
    }
  }, []);

  const loadRegions = useCallback(async () => {
    try {
      const res = await api.get('/residency/regions');
      setRegions(res.data || []);
    } catch {
      // Non-critical
    }
  }, []);

  const loadAuditLog = useCallback(async (page: number = 1) => {
    try {
      const params = new URLSearchParams({ page: String(page), page_size: '15' });
      if (auditFilter.action) params.set('action', auditFilter.action);
      if (auditFilter.allowed) params.set('allowed', auditFilter.allowed);
      const res = await api.get(`/residency/audit-log?${params.toString()}`);
      setAuditLog(res.data || []);
      setAuditTotal(res.pagination?.total_items || 0);
      setAuditPage(page);
    } catch {
      setMessage({ type: 'error', text: 'Failed to load audit log.' });
    }
  }, [auditFilter]);

  useEffect(() => {
    const init = async () => {
      setLoading(true);
      await Promise.all([loadDashboard(), loadConfig(), loadRegions()]);
      setLoading(false);
    };
    init();
  }, [loadDashboard, loadConfig, loadRegions]);

  useEffect(() => {
    if (tab === 'Audit Log') {
      loadAuditLog(1);
    }
  }, [tab, loadAuditLog]);

  // ============================================================
  // HANDLERS
  // ============================================================

  const handleValidateExport = async () => {
    if (!exportRegion.trim()) {
      setMessage({ type: 'error', text: 'Please enter a destination region.' });
      return;
    }
    setExportValidating(true);
    setExportResult(null);
    setMessage(null);
    try {
      const res = await api.post('/residency/validate-export', {
        destination_region: exportRegion.trim().toLowerCase(),
      });
      setExportResult(res.data);
    } catch {
      setMessage({ type: 'error', text: 'Failed to validate export.' });
    } finally {
      setExportValidating(false);
    }
  };

  const handleValidateTransfer = async () => {
    if (!transferCountry.trim()) {
      setMessage({ type: 'error', text: 'Please enter a destination country code.' });
      return;
    }
    setTransferValidating(true);
    setTransferResult(null);
    setMessage(null);
    try {
      const res = await api.post('/residency/validate-transfer', {
        destination_country: transferCountry.trim().toUpperCase(),
      });
      setTransferResult(res.data);
    } catch {
      setMessage({ type: 'error', text: 'Failed to validate transfer.' });
    } finally {
      setTransferValidating(false);
    }
  };

  // ============================================================
  // RENDER
  // ============================================================

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin h-8 w-8 border-4 border-indigo-500 border-t-transparent rounded-full mx-auto mb-4" />
          <p className="text-gray-500">Loading data residency configuration...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-2xl font-bold text-gray-900">Data Residency & Multi-Region Deployment</h1>
          <p className="mt-1 text-sm text-gray-500">
            Manage where your data is stored, control cross-region access, and ensure compliance with local regulations.
          </p>
        </div>

        {/* Message Banner */}
        {message && (
          <div className={`mb-6 p-4 rounded-lg text-sm ${
            message.type === 'error' ? 'bg-red-50 text-red-700 border border-red-200' :
            message.type === 'success' ? 'bg-green-50 text-green-700 border border-green-200' :
            'bg-blue-50 text-blue-700 border border-blue-200'
          }`}>
            {message.text}
            <button className="ml-4 font-medium underline" onClick={() => setMessage(null)}>Dismiss</button>
          </div>
        )}

        {/* Region Indicator Bar */}
        {dashboard && (
          <div className="mb-6 bg-white rounded-lg shadow-sm border border-gray-200 p-4">
            <div className="flex items-center justify-between flex-wrap gap-4">
              <div className="flex items-center gap-4">
                <div className="flex items-center gap-2">
                  <div className="h-3 w-3 rounded-full bg-indigo-500" />
                  <span className="text-sm font-medium text-gray-700">Current Region:</span>
                  <span className="text-sm font-bold text-indigo-700">{dashboard.region_display_name}</span>
                </div>
                <div className={`px-2.5 py-0.5 rounded-full text-xs font-medium ${STATUS_COLOURS[dashboard.compliance_status] || 'bg-gray-100 text-gray-700'}`}>
                  {dashboard.compliance_status === 'compliant' ? 'Compliant' :
                   dashboard.compliance_status === 'monitoring' ? 'Monitoring (Audit Mode)' :
                   'Not Enforced'}
                </div>
              </div>
              <div className="flex items-center gap-4 text-sm text-gray-500">
                <span>Enforcement: <strong className="text-gray-700">{dashboard.enforcement_mode}</strong></span>
                {dashboard.total_blocked_count > 0 && (
                  <span className="text-red-600">
                    {dashboard.total_blocked_count} blocked attempt{dashboard.total_blocked_count !== 1 ? 's' : ''}
                  </span>
                )}
              </div>
            </div>
          </div>
        )}

        {/* Tabs */}
        <div className="border-b border-gray-200 mb-6">
          <nav className="-mb-px flex space-x-8">
            {TABS.map(t => (
              <button
                key={t}
                onClick={() => setTab(t)}
                className={`py-3 px-1 border-b-2 text-sm font-medium transition-colors ${
                  tab === t
                    ? 'border-indigo-500 text-indigo-600'
                    : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                }`}
              >
                {t}
              </button>
            ))}
          </nav>
        </div>

        {/* Tab Content */}
        {tab === 'Overview' && <OverviewTab dashboard={dashboard} config={config} regions={regions} />}
        {tab === 'Storage' && <StorageTab dashboard={dashboard} />}
        {tab === 'Audit Log' && (
          <AuditLogTab
            entries={auditLog}
            total={auditTotal}
            page={auditPage}
            filter={auditFilter}
            onFilterChange={setAuditFilter}
            onPageChange={loadAuditLog}
          />
        )}
        {tab === 'Export Validation' && (
          <ExportValidationTab
            regions={regions}
            exportRegion={exportRegion}
            setExportRegion={setExportRegion}
            result={exportResult}
            validating={exportValidating}
            onValidate={handleValidateExport}
          />
        )}
        {tab === 'Transfer Validation' && (
          <TransferValidationTab
            transferCountry={transferCountry}
            setTransferCountry={setTransferCountry}
            result={transferResult}
            validating={transferValidating}
            onValidate={handleValidateTransfer}
          />
        )}
      </div>
    </div>
  );
}

// ============================================================
// OVERVIEW TAB
// ============================================================

function OverviewTab({ dashboard, config, regions }: {
  dashboard: ResidencyDashboard | null;
  config: ResidencyConfig | null;
  regions: DataRegion[];
}) {
  if (!dashboard || !config) return null;

  return (
    <div className="space-y-6">
      {/* Region Details */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
        <h2 className="text-lg font-semibold text-gray-900 mb-4">Region Configuration</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          <div>
            <label className="block text-xs font-medium text-gray-500 uppercase tracking-wider">Data Region</label>
            <p className="mt-1 text-sm font-medium text-gray-900">{config.display_name}</p>
            <p className="text-xs text-gray-500">{config.description}</p>
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-500 uppercase tracking-wider">Primary Cloud Region</label>
            <p className="mt-1 text-sm font-medium text-gray-900">{config.primary_cloud_region || 'Not configured'}</p>
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-500 uppercase tracking-wider">Failover Region</label>
            <p className="mt-1 text-sm font-medium text-gray-900">{config.failover_cloud_region || 'Not configured'}</p>
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-500 uppercase tracking-wider">Enforcement Mode</label>
            <p className="mt-1 text-sm font-medium text-gray-900 capitalize">{config.enforcement_mode}</p>
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-500 uppercase tracking-wider">Legal Basis</label>
            <p className="mt-1 text-sm font-medium text-gray-900">{config.legal_basis || 'Not specified'}</p>
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-500 uppercase tracking-wider">Data Protection Authority</label>
            <p className="mt-1 text-sm font-medium text-gray-900">{config.data_protection_authority || 'Not specified'}</p>
          </div>
        </div>
      </div>

      {/* Compliance Frameworks */}
      {config.compliance_frameworks && config.compliance_frameworks.length > 0 && (
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Compliance Frameworks</h2>
          <div className="flex flex-wrap gap-2">
            {config.compliance_frameworks.map(fw => (
              <span key={fw} className="px-3 py-1 rounded-full bg-indigo-50 text-indigo-700 text-xs font-medium">
                {fw}
              </span>
            ))}
          </div>
        </div>
      )}

      {/* Recent Blocked Attempts */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
        <h2 className="text-lg font-semibold text-gray-900 mb-4">
          Recent Blocked Cross-Region Attempts
          {dashboard.total_blocked_count > 0 && (
            <span className="ml-2 text-sm font-normal text-red-600">({dashboard.total_blocked_count} total)</span>
          )}
        </h2>
        {dashboard.recent_blocked.length === 0 ? (
          <div className="text-center py-8 text-gray-400">
            <p className="text-sm">No blocked attempts recorded. All data access is within jurisdiction.</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="min-w-full text-sm">
              <thead>
                <tr className="border-b border-gray-200">
                  <th className="text-left py-2 px-3 text-xs font-medium text-gray-500 uppercase">Time</th>
                  <th className="text-left py-2 px-3 text-xs font-medium text-gray-500 uppercase">Action</th>
                  <th className="text-left py-2 px-3 text-xs font-medium text-gray-500 uppercase">User</th>
                  <th className="text-left py-2 px-3 text-xs font-medium text-gray-500 uppercase">From</th>
                  <th className="text-left py-2 px-3 text-xs font-medium text-gray-500 uppercase">To</th>
                  <th className="text-left py-2 px-3 text-xs font-medium text-gray-500 uppercase">Reason</th>
                </tr>
              </thead>
              <tbody>
                {dashboard.recent_blocked.map(entry => (
                  <tr key={entry.id} className="border-b border-gray-100 hover:bg-gray-50">
                    <td className="py-2 px-3 text-gray-600 whitespace-nowrap">
                      {new Date(entry.created_at).toLocaleString()}
                    </td>
                    <td className="py-2 px-3">
                      <span className="px-2 py-0.5 rounded bg-red-50 text-red-700 text-xs font-medium">
                        {ACTION_LABELS[entry.action] || entry.action}
                      </span>
                    </td>
                    <td className="py-2 px-3 text-gray-700">{entry.user_email || 'System'}</td>
                    <td className="py-2 px-3 text-gray-700">{entry.source_region || entry.source_country || '-'}</td>
                    <td className="py-2 px-3 text-gray-700">{entry.destination_region || entry.destination_country || '-'}</td>
                    <td className="py-2 px-3 text-gray-500 max-w-xs truncate">{entry.blocked_reason}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Available Regions */}
      {regions.length > 0 && (
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Available Deployment Regions</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {regions.map(r => (
              <div
                key={r.id}
                className={`p-4 rounded-lg border ${
                  r.region === config?.data_region
                    ? 'border-indigo-300 bg-indigo-50'
                    : 'border-gray-200 bg-gray-50'
                }`}
              >
                <div className="flex items-center justify-between mb-2">
                  <h3 className="font-medium text-gray-900">{r.display_name}</h3>
                  {r.region === config?.data_region && (
                    <span className="px-2 py-0.5 rounded-full bg-indigo-100 text-indigo-700 text-xs font-medium">Current</span>
                  )}
                </div>
                <p className="text-xs text-gray-500 mb-2">{r.description}</p>
                {r.compliance_frameworks && r.compliance_frameworks.length > 0 && (
                  <div className="flex flex-wrap gap-1">
                    {r.compliance_frameworks.slice(0, 4).map(fw => (
                      <span key={fw} className="px-1.5 py-0.5 rounded bg-gray-200 text-gray-600 text-xs">
                        {fw}
                      </span>
                    ))}
                    {r.compliance_frameworks.length > 4 && (
                      <span className="px-1.5 py-0.5 rounded bg-gray-200 text-gray-600 text-xs">
                        +{r.compliance_frameworks.length - 4} more
                      </span>
                    )}
                  </div>
                )}
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

// ============================================================
// STORAGE TAB
// ============================================================

function StorageTab({ dashboard }: { dashboard: ResidencyDashboard | null }) {
  if (!dashboard) return null;

  return (
    <div className="space-y-6">
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
        <h2 className="text-lg font-semibold text-gray-900 mb-4">Storage Locations</h2>
        <p className="text-sm text-gray-500 mb-6">
          All data storage components and their geographic locations. All services must remain within the configured jurisdiction.
        </p>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {dashboard.storage_locations.map(loc => (
            <div key={loc.type} className="p-4 rounded-lg border border-gray-200 bg-gray-50">
              <div className="flex items-start justify-between">
                <div>
                  <h3 className="font-medium text-gray-900">{loc.label}</h3>
                  <div className="mt-2 space-y-1">
                    <p className="text-xs text-gray-500">
                      Region: <span className="font-medium text-gray-700">{loc.region}</span>
                    </p>
                    <p className="text-xs text-gray-500">
                      Provider: <span className="font-medium text-gray-700">{loc.provider}</span>
                    </p>
                    <p className="text-xs text-gray-500">
                      Status: <span className="font-medium text-gray-700 capitalize">{loc.status}</span>
                    </p>
                  </div>
                </div>
                <div className="flex-shrink-0">
                  {loc.within_jurisdiction ? (
                    <span className="inline-flex items-center px-2.5 py-0.5 rounded-full bg-green-100 text-green-800 text-xs font-medium">
                      Within Jurisdiction
                    </span>
                  ) : (
                    <span className="inline-flex items-center px-2.5 py-0.5 rounded-full bg-red-100 text-red-800 text-xs font-medium">
                      Outside Jurisdiction
                    </span>
                  )}
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Cloud Regions */}
      {dashboard.allowed_cloud_regions && Object.keys(dashboard.allowed_cloud_regions).length > 0 && (
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Allowed Cloud Regions</h2>
          <div className="space-y-3">
            {Object.entries(dashboard.allowed_cloud_regions).map(([provider, providerRegions]) => (
              <div key={provider}>
                <h3 className="text-sm font-medium text-gray-700 mb-1 capitalize">{provider}</h3>
                <div className="flex flex-wrap gap-2">
                  {(Array.isArray(providerRegions) ? providerRegions : [String(providerRegions)]).map((region: string) => (
                    <span key={region} className="px-2 py-1 rounded bg-blue-50 text-blue-700 text-xs font-mono">
                      {region}
                    </span>
                  ))}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

// ============================================================
// AUDIT LOG TAB
// ============================================================

function AuditLogTab({ entries, total, page, filter, onFilterChange, onPageChange }: {
  entries: AuditEntry[];
  total: number;
  page: number;
  filter: { action: string; allowed: string };
  onFilterChange: (f: { action: string; allowed: string }) => void;
  onPageChange: (p: number) => void;
}) {
  const totalPages = Math.ceil(total / 15);

  return (
    <div className="space-y-6">
      {/* Filters */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4">
        <div className="flex flex-wrap items-center gap-4">
          <div>
            <label className="block text-xs font-medium text-gray-500 mb-1">Action Type</label>
            <select
              className="block w-48 rounded-md border-gray-300 shadow-sm text-sm focus:ring-indigo-500 focus:border-indigo-500"
              value={filter.action}
              onChange={e => onFilterChange({ ...filter, action: e.target.value })}
            >
              <option value="">All Actions</option>
              <option value="data_access">Data Access</option>
              <option value="data_export">Data Export</option>
              <option value="data_transfer">Data Transfer</option>
              <option value="cross_region_blocked">Cross-Region Blocked</option>
              <option value="config_change">Config Change</option>
              <option value="region_migration">Region Migration</option>
            </select>
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-500 mb-1">Status</label>
            <select
              className="block w-40 rounded-md border-gray-300 shadow-sm text-sm focus:ring-indigo-500 focus:border-indigo-500"
              value={filter.allowed}
              onChange={e => onFilterChange({ ...filter, allowed: e.target.value })}
            >
              <option value="">All</option>
              <option value="true">Allowed</option>
              <option value="false">Blocked</option>
            </select>
          </div>
          <div className="ml-auto text-sm text-gray-500">
            {total} entr{total === 1 ? 'y' : 'ies'} total
          </div>
        </div>
      </div>

      {/* Audit Log Table */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
        {entries.length === 0 ? (
          <div className="text-center py-12 text-gray-400">
            <p className="text-sm">No audit log entries found.</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="min-w-full text-sm">
              <thead className="bg-gray-50">
                <tr>
                  <th className="text-left py-3 px-4 text-xs font-medium text-gray-500 uppercase">Timestamp</th>
                  <th className="text-left py-3 px-4 text-xs font-medium text-gray-500 uppercase">Action</th>
                  <th className="text-left py-3 px-4 text-xs font-medium text-gray-500 uppercase">User</th>
                  <th className="text-left py-3 px-4 text-xs font-medium text-gray-500 uppercase">IP Address</th>
                  <th className="text-left py-3 px-4 text-xs font-medium text-gray-500 uppercase">Source</th>
                  <th className="text-left py-3 px-4 text-xs font-medium text-gray-500 uppercase">Destination</th>
                  <th className="text-left py-3 px-4 text-xs font-medium text-gray-500 uppercase">Status</th>
                  <th className="text-left py-3 px-4 text-xs font-medium text-gray-500 uppercase">Details</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {entries.map(entry => (
                  <tr key={entry.id} className="hover:bg-gray-50">
                    <td className="py-3 px-4 text-gray-600 whitespace-nowrap">
                      {new Date(entry.created_at).toLocaleString()}
                    </td>
                    <td className="py-3 px-4">
                      <span className="px-2 py-0.5 rounded bg-gray-100 text-gray-700 text-xs font-medium">
                        {ACTION_LABELS[entry.action] || entry.action}
                      </span>
                    </td>
                    <td className="py-3 px-4 text-gray-700">{entry.user_email || 'System'}</td>
                    <td className="py-3 px-4 text-gray-500 font-mono text-xs">{entry.ip_address || '-'}</td>
                    <td className="py-3 px-4 text-gray-700">
                      {entry.source_region || entry.source_country || '-'}
                    </td>
                    <td className="py-3 px-4 text-gray-700">
                      {entry.destination_region || entry.destination_country || '-'}
                    </td>
                    <td className="py-3 px-4">
                      {entry.allowed ? (
                        <span className="px-2 py-0.5 rounded-full bg-green-100 text-green-700 text-xs font-medium">Allowed</span>
                      ) : (
                        <span className="px-2 py-0.5 rounded-full bg-red-100 text-red-700 text-xs font-medium">Blocked</span>
                      )}
                    </td>
                    <td className="py-3 px-4 text-gray-500 max-w-xs truncate">
                      {entry.blocked_reason || entry.vendor_name || entry.resource_type || '-'}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        {/* Pagination */}
        {totalPages > 1 && (
          <div className="bg-gray-50 px-4 py-3 flex items-center justify-between border-t border-gray-200">
            <div className="text-sm text-gray-500">
              Page {page} of {totalPages}
            </div>
            <div className="flex gap-2">
              <button
                onClick={() => onPageChange(page - 1)}
                disabled={page <= 1}
                className="px-3 py-1.5 rounded border border-gray-300 text-sm text-gray-700 hover:bg-gray-100 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                Previous
              </button>
              <button
                onClick={() => onPageChange(page + 1)}
                disabled={page >= totalPages}
                className="px-3 py-1.5 rounded border border-gray-300 text-sm text-gray-700 hover:bg-gray-100 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                Next
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

// ============================================================
// EXPORT VALIDATION TAB
// ============================================================

function ExportValidationTab({ regions, exportRegion, setExportRegion, result, validating, onValidate }: {
  regions: DataRegion[];
  exportRegion: string;
  setExportRegion: (v: string) => void;
  result: ValidationResult | null;
  validating: boolean;
  onValidate: () => void;
}) {
  return (
    <div className="space-y-6">
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
        <h2 className="text-lg font-semibold text-gray-900 mb-2">Validate Data Export</h2>
        <p className="text-sm text-gray-500 mb-6">
          Check whether exporting data to another deployment region is permitted under your residency policy.
        </p>
        <div className="flex items-end gap-4">
          <div className="flex-1 max-w-sm">
            <label className="block text-sm font-medium text-gray-700 mb-1">Destination Region</label>
            <select
              className="block w-full rounded-md border-gray-300 shadow-sm text-sm focus:ring-indigo-500 focus:border-indigo-500"
              value={exportRegion}
              onChange={e => setExportRegion(e.target.value)}
            >
              <option value="">Select a region...</option>
              {regions.map(r => (
                <option key={r.region} value={r.region}>{r.display_name} ({r.region})</option>
              ))}
              <option value="us-east">US East (custom)</option>
              <option value="ap-southeast">Asia Pacific Southeast (custom)</option>
            </select>
          </div>
          <button
            onClick={onValidate}
            disabled={validating || !exportRegion}
            className="px-6 py-2 bg-indigo-600 text-white text-sm font-medium rounded-md hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            {validating ? 'Validating...' : 'Validate Export'}
          </button>
        </div>
      </div>

      {/* Export Result */}
      {result && (
        <div className={`bg-white rounded-lg shadow-sm border p-6 ${
          result.allowed ? 'border-green-200' : 'border-red-200'
        }`}>
          <div className="flex items-center gap-3 mb-4">
            <div className={`h-8 w-8 rounded-full flex items-center justify-center ${
              result.allowed ? 'bg-green-100 text-green-600' : 'bg-red-100 text-red-600'
            }`}>
              {result.allowed ? (
                <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" /></svg>
              ) : (
                <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" /></svg>
              )}
            </div>
            <h3 className="text-lg font-semibold text-gray-900">
              {result.allowed ? 'Export Permitted' : 'Export Blocked'}
            </h3>
          </div>
          <div className="space-y-3">
            <div className="flex gap-8 text-sm">
              <div><span className="text-gray-500">From:</span> <span className="font-medium">{result.source_region}</span></div>
              <div><span className="text-gray-500">To:</span> <span className="font-medium">{result.destination_region}</span></div>
            </div>
            <p className="text-sm text-gray-700">{result.reason}</p>
            {result.legal_basis && (
              <p className="text-sm text-gray-500">Legal Basis: <span className="font-medium">{result.legal_basis}</span></p>
            )}
            {result.required_safeguards && result.required_safeguards.length > 0 && (
              <div>
                <p className="text-sm font-medium text-gray-700 mb-1">Required Safeguards:</p>
                <ul className="list-disc list-inside text-sm text-gray-600 space-y-0.5">
                  {result.required_safeguards.map(s => <li key={s}>{s}</li>)}
                </ul>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}

// ============================================================
// TRANSFER VALIDATION TAB
// ============================================================

function TransferValidationTab({ transferCountry, setTransferCountry, result, validating, onValidate }: {
  transferCountry: string;
  setTransferCountry: (v: string) => void;
  result: TransferValidation | null;
  validating: boolean;
  onValidate: () => void;
}) {
  return (
    <div className="space-y-6">
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
        <h2 className="text-lg font-semibold text-gray-900 mb-2">Validate Vendor Data Transfer</h2>
        <p className="text-sm text-gray-500 mb-6">
          Check whether transferring data to a vendor in a specific country is compliant with GDPR adequacy decisions
          and your residency policy.
        </p>
        <div className="flex items-end gap-4">
          <div className="flex-1 max-w-sm">
            <label className="block text-sm font-medium text-gray-700 mb-1">Destination Country (ISO code)</label>
            <input
              type="text"
              placeholder="e.g. US, JP, CN, IN"
              maxLength={2}
              className="block w-full rounded-md border-gray-300 shadow-sm text-sm focus:ring-indigo-500 focus:border-indigo-500 uppercase"
              value={transferCountry}
              onChange={e => setTransferCountry(e.target.value.toUpperCase())}
            />
            {transferCountry && COUNTRY_NAMES[transferCountry] && (
              <p className="mt-1 text-xs text-gray-500">{COUNTRY_NAMES[transferCountry]}</p>
            )}
          </div>
          <button
            onClick={onValidate}
            disabled={validating || !transferCountry.trim()}
            className="px-6 py-2 bg-indigo-600 text-white text-sm font-medium rounded-md hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            {validating ? 'Validating...' : 'Validate Transfer'}
          </button>
        </div>
      </div>

      {/* Transfer Result */}
      {result && (
        <div className={`bg-white rounded-lg shadow-sm border p-6 ${
          result.allowed ? 'border-green-200' : 'border-red-200'
        }`}>
          <div className="flex items-center gap-3 mb-4">
            <div className={`h-8 w-8 rounded-full flex items-center justify-center ${
              result.allowed ? 'bg-green-100 text-green-600' : 'bg-red-100 text-red-600'
            }`}>
              {result.allowed ? (
                <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" /></svg>
              ) : (
                <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" /></svg>
              )}
            </div>
            <h3 className="text-lg font-semibold text-gray-900">
              {result.allowed ? 'Transfer Permitted' : 'Transfer Requires Safeguards'}
            </h3>
          </div>
          <div className="space-y-3">
            <div className="flex gap-8 text-sm">
              <div>
                <span className="text-gray-500">Destination:</span>{' '}
                <span className="font-medium">{COUNTRY_NAMES[result.destination_country] || result.destination_country}</span>
                <span className="text-gray-400 ml-1">({result.destination_country})</span>
              </div>
            </div>
            <div className="flex gap-4 text-sm">
              <span className={`px-2 py-0.5 rounded-full text-xs font-medium ${
                result.is_gdpr_adequate ? 'bg-green-100 text-green-700' : 'bg-yellow-100 text-yellow-700'
              }`}>
                {result.is_gdpr_adequate ? 'GDPR Adequate' : 'No Adequacy Decision'}
              </span>
              {result.requires_additional_safeguards && (
                <span className="px-2 py-0.5 rounded-full bg-orange-100 text-orange-700 text-xs font-medium">
                  Additional Safeguards Required
                </span>
              )}
            </div>
            <p className="text-sm text-gray-700">{result.reason}</p>
            {result.data_protection_authority && (
              <p className="text-sm text-gray-500">
                DPA: <span className="font-medium">{result.data_protection_authority}</span>
              </p>
            )}
            {result.transfer_mechanisms && result.transfer_mechanisms.length > 0 && (
              <div>
                <p className="text-sm font-medium text-gray-700 mb-1">Transfer Mechanisms:</p>
                <ul className="list-disc list-inside text-sm text-gray-600 space-y-0.5">
                  {result.transfer_mechanisms.map(m => <li key={m}>{m}</li>)}
                </ul>
              </div>
            )}
            {result.required_safeguards && result.required_safeguards.length > 0 && (
              <div>
                <p className="text-sm font-medium text-gray-700 mb-1">Required Safeguards:</p>
                <ul className="list-disc list-inside text-sm text-gray-600 space-y-0.5">
                  {result.required_safeguards.map(s => <li key={s}>{s}</li>)}
                </ul>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
