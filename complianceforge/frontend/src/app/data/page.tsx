'use client';

import { useEffect, useState, useCallback } from 'react';
import Link from 'next/link';
import api from '@/lib/api';

// ============================================================
// TYPE DEFINITIONS
// ============================================================

interface ROPADashboard {
  total_activities: number;
  active_activities: number;
  draft_activities: number;
  by_legal_basis: Record<string, number>;
  by_department: Record<string, number>;
  special_categories_count: number;
  international_transfers: number;
  dpia_required: number;
  dpia_completed: number;
  dpia_in_progress: number;
  dpia_pending: number;
  overdue_reviews: number;
  upcoming_reviews_30_days: number;
  total_data_flows: number;
  high_risk_activities: number;
  by_role: Record<string, number>;
  last_export_date: string | null;
}

interface ProcessingActivity {
  id: string;
  activity_ref: string;
  name: string;
  description: string;
  purpose: string;
  legal_basis: string;
  status: string;
  role: string;
  department: string;
  special_categories_processed: boolean;
  involves_international_transfer: boolean;
  dpia_required: boolean;
  dpia_status: string;
  risk_level: string;
  next_review_date: string | null;
  created_at: string;
}

interface DataClassification {
  id: string;
  name: string;
  level: number;
  description: string;
  handling_requirements: string;
  encryption_required: boolean;
  access_restriction_required: boolean;
  data_masking_required: boolean;
  color_hex: string;
  is_system: boolean;
}

interface DataCategory {
  id: string;
  name: string;
  category_type: string;
  gdpr_special_category: boolean;
  description: string;
  examples: string[];
  retention_period_months: number | null;
  is_system: boolean;
}

interface ROPAExport {
  id: string;
  export_ref: string;
  export_date: string;
  format: string;
  activities_included: number;
  export_reason: string;
  notes: string;
}

interface TransferEntry {
  activity_id: string;
  activity_ref: string;
  activity_name: string;
  transfer_countries: string[];
  safeguards: string;
  safeguards_detail: string;
  tia_conducted: boolean;
  tia_date: string | null;
  legal_basis: string;
  department: string;
  status: string;
}

// ============================================================
// CONSTANTS
// ============================================================

const STATUS_COLORS: Record<string, string> = {
  draft: 'bg-gray-100 text-gray-700',
  active: 'bg-green-100 text-green-800',
  under_review: 'bg-yellow-100 text-yellow-800',
  suspended: 'bg-red-100 text-red-800',
  retired: 'bg-gray-200 text-gray-500',
};

const LEGAL_BASIS_LABELS: Record<string, string> = {
  consent: 'Consent (Art. 6(1)(a))',
  contract: 'Contract (Art. 6(1)(b))',
  legal_obligation: 'Legal Obligation (Art. 6(1)(c))',
  vital_interest: 'Vital Interest (Art. 6(1)(d))',
  public_task: 'Public Task (Art. 6(1)(e))',
  legitimate_interest: 'Legitimate Interest (Art. 6(1)(f))',
};

const RISK_COLORS: Record<string, string> = {
  critical: 'bg-red-100 text-red-800',
  high: 'bg-orange-100 text-orange-800',
  medium: 'bg-yellow-100 text-yellow-800',
  low: 'bg-green-100 text-green-800',
};

const DPIA_STATUS_COLORS: Record<string, string> = {
  not_required: 'bg-gray-100 text-gray-600',
  required: 'bg-red-100 text-red-800',
  in_progress: 'bg-yellow-100 text-yellow-800',
  completed: 'bg-green-100 text-green-800',
  review_needed: 'bg-orange-100 text-orange-800',
};

const CATEGORY_TYPE_LABELS: Record<string, string> = {
  personal_data: 'Personal Data',
  special_category: 'Special Category (Art. 9)',
  financial: 'Financial',
  technical: 'Technical',
  business: 'Business',
  public: 'Public',
  proprietary: 'Proprietary',
};

type TabType = 'dashboard' | 'activities' | 'classifications' | 'categories' | 'exports' | 'transfers';

// ============================================================
// COMPONENT
// ============================================================

export default function DataGovernancePage() {
  const [activeTab, setActiveTab] = useState<TabType>('dashboard');
  const [dashboard, setDashboard] = useState<ROPADashboard | null>(null);
  const [activities, setActivities] = useState<ProcessingActivity[]>([]);
  const [activitiesTotal, setActivitiesTotal] = useState(0);
  const [activitiesPage, setActivitiesPage] = useState(1);
  const [classifications, setClassifications] = useState<DataClassification[]>([]);
  const [categories, setCategories] = useState<DataCategory[]>([]);
  const [exports, setExports] = useState<ROPAExport[]>([]);
  const [transfers, setTransfers] = useState<TransferEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [filterStatus, setFilterStatus] = useState('');
  const [filterLegalBasis, setFilterLegalBasis] = useState('');
  const [searchQuery, setSearchQuery] = useState('');

  const loadDashboard = useCallback(async () => {
    try {
      const res = await api.getROPADashboard();
      if (res.data) setDashboard(res.data);
    } catch (err) {
      console.error('Failed to load ROPA dashboard:', err);
    }
  }, []);

  const loadActivities = useCallback(async (page = 1) => {
    try {
      let qs = `page=${page}&page_size=20`;
      if (filterStatus) qs += `&status=${filterStatus}`;
      if (filterLegalBasis) qs += `&legal_basis=${filterLegalBasis}`;
      if (searchQuery) qs += `&search=${encodeURIComponent(searchQuery)}`;
      const res = await api.getProcessingActivities(qs);
      if (res.data) {
        setActivities(res.data);
        setActivitiesTotal(res.pagination?.total_items || 0);
      }
    } catch (err) {
      console.error('Failed to load processing activities:', err);
    }
  }, [filterStatus, filterLegalBasis, searchQuery]);

  const loadClassifications = useCallback(async () => {
    try {
      const res = await api.getDataClassifications();
      if (res.data) setClassifications(res.data);
    } catch (err) {
      console.error('Failed to load classifications:', err);
    }
  }, []);

  const loadCategories = useCallback(async () => {
    try {
      const res = await api.getDataCategories();
      if (res.data) setCategories(res.data);
    } catch (err) {
      console.error('Failed to load categories:', err);
    }
  }, []);

  const loadExports = useCallback(async () => {
    try {
      const res = await api.getROPAExports();
      if (res.data) setExports(res.data);
    } catch (err) {
      console.error('Failed to load ROPA exports:', err);
    }
  }, []);

  const loadTransfers = useCallback(async () => {
    try {
      const res = await api.getTransferRegister();
      if (res.data) setTransfers(res.data);
    } catch (err) {
      console.error('Failed to load transfer register:', err);
    }
  }, []);

  useEffect(() => {
    const load = async () => {
      setLoading(true);
      await Promise.all([loadDashboard(), loadActivities(1), loadClassifications()]);
      setLoading(false);
    };
    load();
  }, [loadDashboard, loadActivities, loadClassifications]);

  useEffect(() => {
    if (activeTab === 'categories') loadCategories();
    if (activeTab === 'exports') loadExports();
    if (activeTab === 'transfers') loadTransfers();
  }, [activeTab, loadCategories, loadExports, loadTransfers]);

  useEffect(() => {
    if (activeTab === 'activities') {
      loadActivities(activitiesPage);
    }
  }, [activeTab, activitiesPage, loadActivities]);

  const handleExportROPA = async (format: string) => {
    try {
      await api.exportROPA({ format, reason: 'ad_hoc' });
      loadExports();
      loadDashboard();
    } catch (err) {
      console.error('Failed to export ROPA:', err);
    }
  };

  if (loading) {
    return (
      <div className="p-8">
        <div className="animate-pulse space-y-4">
          <div className="h-8 bg-gray-200 rounded w-1/3"></div>
          <div className="grid grid-cols-4 gap-4">
            {[1, 2, 3, 4].map(i => (
              <div key={i} className="h-24 bg-gray-200 rounded"></div>
            ))}
          </div>
        </div>
      </div>
    );
  }

  const tabs: { key: TabType; label: string }[] = [
    { key: 'dashboard', label: 'Dashboard' },
    { key: 'activities', label: 'Processing Activities' },
    { key: 'classifications', label: 'Classifications' },
    { key: 'categories', label: 'Data Categories' },
    { key: 'exports', label: 'ROPA Exports' },
    { key: 'transfers', label: 'Transfer Register' },
  ];

  return (
    <div className="p-6 max-w-7xl mx-auto">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900">Data Governance Hub</h1>
        <p className="text-gray-600 mt-1">
          ROPA management, data classification, data categories, and GDPR Art. 30 compliance
        </p>
      </div>

      {/* Tabs */}
      <div className="border-b border-gray-200 mb-6">
        <nav className="flex space-x-6">
          {tabs.map(tab => (
            <button
              key={tab.key}
              onClick={() => setActiveTab(tab.key)}
              className={`py-3 px-1 border-b-2 text-sm font-medium transition-colors ${
                activeTab === tab.key
                  ? 'border-blue-600 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              {tab.label}
            </button>
          ))}
        </nav>
      </div>

      {/* Tab Content */}
      {activeTab === 'dashboard' && dashboard && (
        <DashboardTab dashboard={dashboard} onExport={handleExportROPA} />
      )}
      {activeTab === 'activities' && (
        <ActivitiesTab
          activities={activities}
          total={activitiesTotal}
          page={activitiesPage}
          onPageChange={setActivitiesPage}
          filterStatus={filterStatus}
          setFilterStatus={setFilterStatus}
          filterLegalBasis={filterLegalBasis}
          setFilterLegalBasis={setFilterLegalBasis}
          searchQuery={searchQuery}
          setSearchQuery={setSearchQuery}
        />
      )}
      {activeTab === 'classifications' && (
        <ClassificationsTab classifications={classifications} />
      )}
      {activeTab === 'categories' && (
        <CategoriesTab categories={categories} />
      )}
      {activeTab === 'exports' && (
        <ExportsTab exports={exports} onExport={handleExportROPA} />
      )}
      {activeTab === 'transfers' && (
        <TransfersTab transfers={transfers} />
      )}
    </div>
  );
}

// ============================================================
// DASHBOARD TAB
// ============================================================

function DashboardTab({ dashboard, onExport }: { dashboard: ROPADashboard; onExport: (format: string) => void }) {
  return (
    <div className="space-y-6">
      {/* KPI Cards */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <KPICard title="Total Activities" value={dashboard.total_activities} subtitle={`${dashboard.active_activities} active`} />
        <KPICard title="Special Categories" value={dashboard.special_categories_count} subtitle="Art. 9 data" color="purple" />
        <KPICard title="International Transfers" value={dashboard.international_transfers} subtitle="Cross-border" color="blue" />
        <KPICard title="High Risk" value={dashboard.high_risk_activities} subtitle="DPIA needed" color="red" />
      </div>

      {/* DPIA Status + Overdue Reviews */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div className="bg-white border rounded-lg p-5">
          <h3 className="text-sm font-semibold text-gray-700 mb-4">DPIA Status</h3>
          <div className="space-y-3">
            <ProgressBar label="Required" value={dashboard.dpia_required} total={dashboard.total_activities} color="bg-red-500" />
            <ProgressBar label="Completed" value={dashboard.dpia_completed} total={dashboard.dpia_required || 1} color="bg-green-500" />
            <ProgressBar label="In Progress" value={dashboard.dpia_in_progress} total={dashboard.dpia_required || 1} color="bg-yellow-500" />
            <ProgressBar label="Pending" value={dashboard.dpia_pending} total={dashboard.dpia_required || 1} color="bg-orange-500" />
          </div>
        </div>

        <div className="bg-white border rounded-lg p-5">
          <h3 className="text-sm font-semibold text-gray-700 mb-4">Review Status</h3>
          <div className="space-y-4">
            <div className="flex justify-between items-center">
              <span className="text-sm text-gray-600">Overdue Reviews</span>
              <span className={`px-2 py-1 rounded text-xs font-medium ${dashboard.overdue_reviews > 0 ? 'bg-red-100 text-red-800' : 'bg-green-100 text-green-800'}`}>
                {dashboard.overdue_reviews}
              </span>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-sm text-gray-600">Due in 30 days</span>
              <span className="px-2 py-1 rounded text-xs font-medium bg-yellow-100 text-yellow-800">
                {dashboard.upcoming_reviews_30_days}
              </span>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-sm text-gray-600">Total Data Flows</span>
              <span className="px-2 py-1 rounded text-xs font-medium bg-blue-100 text-blue-800">
                {dashboard.total_data_flows}
              </span>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-sm text-gray-600">Last ROPA Export</span>
              <span className="text-xs text-gray-500">
                {dashboard.last_export_date ? new Date(dashboard.last_export_date).toLocaleDateString() : 'Never'}
              </span>
            </div>
          </div>
        </div>
      </div>

      {/* Legal Basis Breakdown */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div className="bg-white border rounded-lg p-5">
          <h3 className="text-sm font-semibold text-gray-700 mb-4">By Legal Basis</h3>
          <div className="space-y-2">
            {Object.entries(dashboard.by_legal_basis).map(([basis, count]) => (
              <div key={basis} className="flex justify-between items-center">
                <span className="text-sm text-gray-600">{LEGAL_BASIS_LABELS[basis] || basis}</span>
                <span className="text-sm font-medium">{count}</span>
              </div>
            ))}
            {Object.keys(dashboard.by_legal_basis).length === 0 && (
              <p className="text-sm text-gray-400">No activities recorded yet</p>
            )}
          </div>
        </div>

        <div className="bg-white border rounded-lg p-5">
          <h3 className="text-sm font-semibold text-gray-700 mb-4">By Department</h3>
          <div className="space-y-2">
            {Object.entries(dashboard.by_department).map(([dept, count]) => (
              <div key={dept} className="flex justify-between items-center">
                <span className="text-sm text-gray-600">{dept}</span>
                <span className="text-sm font-medium">{count}</span>
              </div>
            ))}
            {Object.keys(dashboard.by_department).length === 0 && (
              <p className="text-sm text-gray-400">No departments assigned yet</p>
            )}
          </div>
        </div>
      </div>

      {/* Quick Export */}
      <div className="bg-white border rounded-lg p-5">
        <h3 className="text-sm font-semibold text-gray-700 mb-3">Quick ROPA Export</h3>
        <div className="flex gap-3">
          <button onClick={() => onExport('pdf')} className="px-4 py-2 bg-blue-600 text-white text-sm rounded hover:bg-blue-700">
            Export PDF
          </button>
          <button onClick={() => onExport('xlsx')} className="px-4 py-2 bg-green-600 text-white text-sm rounded hover:bg-green-700">
            Export XLSX
          </button>
          <button onClick={() => onExport('csv')} className="px-4 py-2 bg-gray-600 text-white text-sm rounded hover:bg-gray-700">
            Export CSV
          </button>
        </div>
      </div>
    </div>
  );
}

// ============================================================
// ACTIVITIES TAB
// ============================================================

function ActivitiesTab({
  activities, total, page, onPageChange,
  filterStatus, setFilterStatus,
  filterLegalBasis, setFilterLegalBasis,
  searchQuery, setSearchQuery,
}: {
  activities: ProcessingActivity[];
  total: number;
  page: number;
  onPageChange: (p: number) => void;
  filterStatus: string;
  setFilterStatus: (v: string) => void;
  filterLegalBasis: string;
  setFilterLegalBasis: (v: string) => void;
  searchQuery: string;
  setSearchQuery: (v: string) => void;
}) {
  const totalPages = Math.ceil(total / 20);

  return (
    <div className="space-y-4">
      {/* Filters */}
      <div className="flex flex-wrap gap-3 items-center">
        <input
          type="text"
          placeholder="Search activities..."
          value={searchQuery}
          onChange={e => { setSearchQuery(e.target.value); onPageChange(1); }}
          className="px-3 py-2 border rounded text-sm w-64"
        />
        <select value={filterStatus} onChange={e => { setFilterStatus(e.target.value); onPageChange(1); }} className="px-3 py-2 border rounded text-sm">
          <option value="">All Statuses</option>
          <option value="draft">Draft</option>
          <option value="active">Active</option>
          <option value="under_review">Under Review</option>
          <option value="suspended">Suspended</option>
          <option value="retired">Retired</option>
        </select>
        <select value={filterLegalBasis} onChange={e => { setFilterLegalBasis(e.target.value); onPageChange(1); }} className="px-3 py-2 border rounded text-sm">
          <option value="">All Legal Bases</option>
          {Object.entries(LEGAL_BASIS_LABELS).map(([k, v]) => (
            <option key={k} value={k}>{v}</option>
          ))}
        </select>
        <span className="text-sm text-gray-500 ml-auto">{total} activities</span>
      </div>

      {/* Activity Table */}
      <div className="bg-white border rounded-lg overflow-hidden">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Ref</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Name</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Legal Basis</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Risk</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">DPIA</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Dept</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Flags</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200">
            {activities.map(a => (
              <tr key={a.id} className="hover:bg-gray-50">
                <td className="px-4 py-3 text-sm font-mono text-blue-600">
                  <Link href={`/data/${a.id}`}>{a.activity_ref}</Link>
                </td>
                <td className="px-4 py-3 text-sm">
                  <Link href={`/data/${a.id}`} className="text-gray-900 hover:text-blue-600">
                    {a.name}
                  </Link>
                </td>
                <td className="px-4 py-3">
                  <span className={`px-2 py-1 rounded text-xs font-medium ${STATUS_COLORS[a.status] || 'bg-gray-100'}`}>
                    {a.status}
                  </span>
                </td>
                <td className="px-4 py-3 text-xs text-gray-600">
                  {LEGAL_BASIS_LABELS[a.legal_basis] || a.legal_basis || '-'}
                </td>
                <td className="px-4 py-3">
                  <span className={`px-2 py-1 rounded text-xs font-medium ${RISK_COLORS[a.risk_level] || 'bg-gray-100'}`}>
                    {a.risk_level}
                  </span>
                </td>
                <td className="px-4 py-3">
                  <span className={`px-2 py-1 rounded text-xs font-medium ${DPIA_STATUS_COLORS[a.dpia_status] || 'bg-gray-100'}`}>
                    {a.dpia_status}
                  </span>
                </td>
                <td className="px-4 py-3 text-xs text-gray-600">{a.department || '-'}</td>
                <td className="px-4 py-3">
                  <div className="flex gap-1">
                    {a.special_categories_processed && (
                      <span className="px-1.5 py-0.5 bg-purple-100 text-purple-700 rounded text-xs" title="Special Categories">Art.9</span>
                    )}
                    {a.involves_international_transfer && (
                      <span className="px-1.5 py-0.5 bg-blue-100 text-blue-700 rounded text-xs" title="International Transfer">Intl</span>
                    )}
                  </div>
                </td>
              </tr>
            ))}
            {activities.length === 0 && (
              <tr>
                <td colSpan={8} className="px-4 py-8 text-center text-gray-400 text-sm">
                  No processing activities found. Create your first activity to begin Art. 30 compliance.
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex justify-between items-center">
          <button
            onClick={() => onPageChange(page - 1)}
            disabled={page <= 1}
            className="px-3 py-1 border rounded text-sm disabled:opacity-50"
          >
            Previous
          </button>
          <span className="text-sm text-gray-600">Page {page} of {totalPages}</span>
          <button
            onClick={() => onPageChange(page + 1)}
            disabled={page >= totalPages}
            className="px-3 py-1 border rounded text-sm disabled:opacity-50"
          >
            Next
          </button>
        </div>
      )}
    </div>
  );
}

// ============================================================
// CLASSIFICATIONS TAB
// ============================================================

function ClassificationsTab({ classifications }: { classifications: DataClassification[] }) {
  return (
    <div className="space-y-4">
      <div className="flex justify-between items-center">
        <h3 className="text-lg font-semibold">Data Classification Levels</h3>
      </div>
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {classifications.map(c => (
          <div key={c.id} className="bg-white border rounded-lg p-4">
            <div className="flex items-center gap-3 mb-3">
              <div className="w-4 h-4 rounded-full" style={{ backgroundColor: c.color_hex || '#94a3b8' }} />
              <h4 className="font-semibold text-gray-900">
                Level {c.level}: {c.name}
              </h4>
            </div>
            <p className="text-sm text-gray-600 mb-3">{c.description}</p>
            <div className="space-y-1.5 text-xs">
              <div className="flex gap-2">
                <span className={`px-2 py-0.5 rounded ${c.encryption_required ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-500'}`}>
                  Encryption {c.encryption_required ? 'Required' : 'Optional'}
                </span>
                <span className={`px-2 py-0.5 rounded ${c.access_restriction_required ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-500'}`}>
                  Access Control {c.access_restriction_required ? 'Required' : 'Optional'}
                </span>
              </div>
              {c.data_masking_required && (
                <span className="px-2 py-0.5 rounded bg-purple-100 text-purple-700">Data Masking Required</span>
              )}
            </div>
            {c.handling_requirements && (
              <p className="text-xs text-gray-500 mt-3 border-t pt-2">{c.handling_requirements}</p>
            )}
          </div>
        ))}
        {classifications.length === 0 && (
          <p className="text-gray-400 text-sm col-span-full">No classification levels configured yet.</p>
        )}
      </div>
    </div>
  );
}

// ============================================================
// CATEGORIES TAB
// ============================================================

function CategoriesTab({ categories }: { categories: DataCategory[] }) {
  const grouped = categories.reduce((acc, cat) => {
    const type = cat.category_type;
    if (!acc[type]) acc[type] = [];
    acc[type].push(cat);
    return acc;
  }, {} as Record<string, DataCategory[]>);

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h3 className="text-lg font-semibold">Data Categories</h3>
        <span className="text-sm text-gray-500">{categories.length} categories</span>
      </div>
      {Object.entries(grouped).map(([type, cats]) => (
        <div key={type}>
          <h4 className="text-sm font-semibold text-gray-700 mb-3 flex items-center gap-2">
            {CATEGORY_TYPE_LABELS[type] || type}
            <span className="bg-gray-100 text-gray-500 px-2 py-0.5 rounded text-xs">{cats.length}</span>
          </h4>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
            {cats.map(cat => (
              <div key={cat.id} className="bg-white border rounded p-3">
                <div className="flex items-center gap-2 mb-1">
                  <span className="text-sm font-medium text-gray-900">{cat.name}</span>
                  {cat.gdpr_special_category && (
                    <span className="px-1.5 py-0.5 bg-purple-100 text-purple-700 rounded text-xs">Art.9</span>
                  )}
                </div>
                {cat.description && (
                  <p className="text-xs text-gray-500 mb-2">{cat.description}</p>
                )}
                {cat.examples && cat.examples.length > 0 && (
                  <p className="text-xs text-gray-400">e.g. {cat.examples.slice(0, 2).join(', ')}</p>
                )}
                {cat.retention_period_months && (
                  <p className="text-xs text-gray-400 mt-1">Retention: {cat.retention_period_months} months</p>
                )}
              </div>
            ))}
          </div>
        </div>
      ))}
      {categories.length === 0 && (
        <p className="text-gray-400 text-sm">No data categories configured yet. Run the seed script to populate defaults.</p>
      )}
    </div>
  );
}

// ============================================================
// EXPORTS TAB
// ============================================================

function ExportsTab({ exports, onExport }: { exports: ROPAExport[]; onExport: (format: string) => void }) {
  return (
    <div className="space-y-4">
      <div className="flex justify-between items-center">
        <h3 className="text-lg font-semibold">ROPA Export History</h3>
        <div className="flex gap-2">
          <button onClick={() => onExport('pdf')} className="px-3 py-1.5 bg-blue-600 text-white text-sm rounded hover:bg-blue-700">
            New PDF Export
          </button>
          <button onClick={() => onExport('xlsx')} className="px-3 py-1.5 bg-green-600 text-white text-sm rounded hover:bg-green-700">
            New XLSX Export
          </button>
        </div>
      </div>
      <div className="bg-white border rounded-lg overflow-hidden">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Ref</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Date</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Format</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Activities</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Reason</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200">
            {exports.map(e => (
              <tr key={e.id} className="hover:bg-gray-50">
                <td className="px-4 py-3 text-sm font-mono">{e.export_ref}</td>
                <td className="px-4 py-3 text-sm text-gray-600">{new Date(e.export_date).toLocaleString()}</td>
                <td className="px-4 py-3">
                  <span className="px-2 py-1 bg-blue-100 text-blue-700 rounded text-xs uppercase">{e.format}</span>
                </td>
                <td className="px-4 py-3 text-sm">{e.activities_included}</td>
                <td className="px-4 py-3 text-sm text-gray-600">{e.export_reason.replace(/_/g, ' ')}</td>
              </tr>
            ))}
            {exports.length === 0 && (
              <tr>
                <td colSpan={5} className="px-4 py-8 text-center text-gray-400 text-sm">
                  No ROPA exports yet. Generate your first export above.
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}

// ============================================================
// TRANSFERS TAB
// ============================================================

function TransfersTab({ transfers }: { transfers: TransferEntry[] }) {
  return (
    <div className="space-y-4">
      <div className="flex justify-between items-center">
        <h3 className="text-lg font-semibold">International Transfer Register</h3>
        <span className="text-sm text-gray-500">{transfers.length} transfers</span>
      </div>
      <div className="bg-white border rounded-lg overflow-hidden">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Ref</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Activity</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Countries</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Safeguards</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">TIA</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200">
            {transfers.map(t => (
              <tr key={t.activity_id} className="hover:bg-gray-50">
                <td className="px-4 py-3 text-sm font-mono text-blue-600">
                  <Link href={`/data/${t.activity_id}`}>{t.activity_ref}</Link>
                </td>
                <td className="px-4 py-3 text-sm">{t.activity_name}</td>
                <td className="px-4 py-3">
                  <div className="flex flex-wrap gap-1">
                    {t.transfer_countries.map(c => (
                      <span key={c} className="px-1.5 py-0.5 bg-blue-50 text-blue-700 rounded text-xs">{c}</span>
                    ))}
                  </div>
                </td>
                <td className="px-4 py-3 text-xs text-gray-600">{t.safeguards.replace(/_/g, ' ') || 'None specified'}</td>
                <td className="px-4 py-3">
                  <span className={`px-2 py-1 rounded text-xs font-medium ${t.tia_conducted ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'}`}>
                    {t.tia_conducted ? 'Completed' : 'Not done'}
                  </span>
                </td>
                <td className="px-4 py-3">
                  <span className={`px-2 py-1 rounded text-xs font-medium ${STATUS_COLORS[t.status] || 'bg-gray-100'}`}>
                    {t.status}
                  </span>
                </td>
              </tr>
            ))}
            {transfers.length === 0 && (
              <tr>
                <td colSpan={6} className="px-4 py-8 text-center text-gray-400 text-sm">
                  No international transfers registered.
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}

// ============================================================
// SHARED COMPONENTS
// ============================================================

function KPICard({ title, value, subtitle, color = 'default' }: {
  title: string; value: number; subtitle: string; color?: string;
}) {
  const borderColors: Record<string, string> = {
    default: 'border-l-gray-400',
    red: 'border-l-red-500',
    blue: 'border-l-blue-500',
    purple: 'border-l-purple-500',
    green: 'border-l-green-500',
  };

  return (
    <div className={`bg-white border rounded-lg p-4 border-l-4 ${borderColors[color] || borderColors.default}`}>
      <p className="text-xs text-gray-500 uppercase tracking-wide">{title}</p>
      <p className="text-2xl font-bold text-gray-900 mt-1">{value}</p>
      <p className="text-xs text-gray-400 mt-1">{subtitle}</p>
    </div>
  );
}

function ProgressBar({ label, value, total, color }: { label: string; value: number; total: number; color: string }) {
  const pct = total > 0 ? Math.min(100, (value / total) * 100) : 0;
  return (
    <div>
      <div className="flex justify-between text-xs text-gray-600 mb-1">
        <span>{label}</span>
        <span>{value} / {total}</span>
      </div>
      <div className="w-full bg-gray-100 rounded-full h-2">
        <div className={`${color} rounded-full h-2 transition-all`} style={{ width: `${pct}%` }} />
      </div>
    </div>
  );
}
