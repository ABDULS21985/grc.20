'use client';

import { useEffect, useState } from 'react';
import api from '@/lib/api';

// ============================================================
// TYPES
// ============================================================

interface EvidenceTemplate {
  id: string;
  framework_control_code: string;
  framework_code: string;
  name: string;
  description: string;
  evidence_category: string;
  collection_method: string;
  collection_instructions: string;
  collection_frequency: string;
  typical_collection_time_minutes: number;
  validation_rules: any[];
  acceptance_criteria: string;
  common_rejection_reasons: string[];
  difficulty: string;
  auditor_priority: string;
  is_system: boolean;
  tags: string[];
}

interface EvidenceRequirement {
  id: string;
  status: string;
  is_mandatory: boolean;
  validation_status: string;
  next_collection_due: string | null;
  due_date: string | null;
  template_name: string;
  framework_control_code: string;
  framework_code: string;
  assigned_to: string | null;
  consecutive_failures: number;
  notes: string;
}

interface EvidenceGaps {
  total_requirements: number;
  collected: number;
  pending: number;
  overdue: number;
  validated: number;
  failed: number;
  coverage_percent: number;
  gaps_by_framework: FrameworkGap[];
  critical_gaps: CriticalGapItem[];
  overdue_items: OverdueItem[];
}

interface FrameworkGap {
  framework_code: string;
  total_required: number;
  collected: number;
  coverage_percent: number;
}

interface CriticalGapItem {
  requirement_id: string;
  template_name: string;
  framework_control_code: string;
  framework_code: string;
  auditor_priority: string;
  due_date: string | null;
  days_overdue: number;
}

interface OverdueItem {
  requirement_id: string;
  template_name: string;
  framework_control_code: string;
  due_date: string | null;
  days_overdue: number;
}

interface CollectionSchedule {
  upcoming_this_week: ScheduleItem[];
  upcoming_this_month: ScheduleItem[];
  upcoming_this_quarter: ScheduleItem[];
  total_scheduled: number;
}

interface ScheduleItem {
  requirement_id: string;
  template_name: string;
  framework_control_code: string;
  framework_code: string;
  collection_frequency: string;
  next_due: string | null;
  difficulty: string;
  estimated_minutes: number;
}

type TabType = 'templates' | 'requirements' | 'gaps' | 'schedule';

// ============================================================
// HELPERS
// ============================================================

const priorityColors: Record<string, string> = {
  critical: 'bg-red-100 text-red-800',
  high: 'bg-orange-100 text-orange-800',
  medium: 'bg-yellow-100 text-yellow-800',
  low: 'bg-green-100 text-green-800',
  informational: 'bg-gray-100 text-gray-800',
};

const statusColors: Record<string, string> = {
  pending: 'bg-gray-100 text-gray-800',
  in_progress: 'bg-blue-100 text-blue-800',
  collected: 'bg-green-100 text-green-800',
  validated: 'bg-emerald-100 text-emerald-800',
  rejected: 'bg-red-100 text-red-800',
  overdue: 'bg-red-100 text-red-800',
  waived: 'bg-purple-100 text-purple-800',
  not_applicable: 'bg-gray-100 text-gray-600',
};

const difficultyColors: Record<string, string> = {
  trivial: 'bg-green-100 text-green-700',
  easy: 'bg-green-100 text-green-800',
  moderate: 'bg-yellow-100 text-yellow-800',
  hard: 'bg-orange-100 text-orange-800',
  expert: 'bg-red-100 text-red-800',
};

const categoryLabels: Record<string, string> = {
  document: 'Document',
  screenshot: 'Screenshot',
  configuration_export: 'Config Export',
  log_extract: 'Log Extract',
  scan_report: 'Scan Report',
  interview_record: 'Interview',
  test_result: 'Test Result',
  certification: 'Certification',
  training_record: 'Training Record',
  policy_document: 'Policy',
  procedure_document: 'Procedure',
  meeting_minutes: 'Meeting Minutes',
  email_confirmation: 'Email',
  system_report: 'System Report',
  audit_trail: 'Audit Trail',
};

function Badge({ text, className = '' }: { text: string; className?: string }) {
  return (
    <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${className}`}>
      {text}
    </span>
  );
}

// ============================================================
// MAIN PAGE COMPONENT
// ============================================================

export default function EvidencePage() {
  const [activeTab, setActiveTab] = useState<TabType>('templates');
  const [templates, setTemplates] = useState<EvidenceTemplate[]>([]);
  const [requirements, setRequirements] = useState<EvidenceRequirement[]>([]);
  const [gaps, setGaps] = useState<EvidenceGaps | null>(null);
  const [schedule, setSchedule] = useState<CollectionSchedule | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [templatePage, setTemplatePage] = useState(1);
  const [templateTotal, setTemplateTotal] = useState(0);
  const [requirementPage, setRequirementPage] = useState(1);
  const [requirementTotal, setRequirementTotal] = useState(0);
  const [frameworkFilter, setFrameworkFilter] = useState('');
  const [categoryFilter, setCategoryFilter] = useState('');
  const [priorityFilter, setPriorityFilter] = useState('');
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedTemplate, setSelectedTemplate] = useState<EvidenceTemplate | null>(null);
  const pageSize = 20;

  useEffect(() => {
    loadData();
  }, [activeTab, templatePage, requirementPage, frameworkFilter, categoryFilter, priorityFilter, searchQuery]);

  async function loadData() {
    setLoading(true);
    setError(null);
    try {
      switch (activeTab) {
        case 'templates': {
          let qs = `page=${templatePage}&page_size=${pageSize}`;
          if (frameworkFilter) qs += `&framework_code=${frameworkFilter}`;
          if (categoryFilter) qs += `&category=${categoryFilter}`;
          if (priorityFilter) qs += `&auditor_priority=${priorityFilter}`;
          if (searchQuery) qs += `&search=${encodeURIComponent(searchQuery)}`;
          const res = await api.getEvidenceTemplates(qs);
          setTemplates(res.data || []);
          setTemplateTotal(res.pagination?.total_items || 0);
          break;
        }
        case 'requirements': {
          let qs = `page=${requirementPage}&page_size=${pageSize}`;
          if (frameworkFilter) qs += `&framework_code=${frameworkFilter}`;
          const res = await api.getEvidenceRequirements(qs);
          setRequirements(res.data || []);
          setRequirementTotal(res.pagination?.total_items || 0);
          break;
        }
        case 'gaps': {
          const res = await api.getEvidenceGaps();
          setGaps(res.data);
          break;
        }
        case 'schedule': {
          const res = await api.getCollectionSchedule();
          setSchedule(res.data);
          break;
        }
      }
    } catch (err: any) {
      setError(err.message || 'Failed to load data');
    } finally {
      setLoading(false);
    }
  }

  const tabs: { key: TabType; label: string }[] = [
    { key: 'templates', label: 'Template Library' },
    { key: 'requirements', label: 'Requirements' },
    { key: 'gaps', label: 'Evidence Gaps' },
    { key: 'schedule', label: 'Collection Schedule' },
  ];

  return (
    <div className="p-6 max-w-7xl mx-auto">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900">Evidence Management</h1>
        <p className="text-gray-600 mt-1">
          Browse the evidence template library, track collection requirements, and identify gaps.
        </p>
      </div>

      {/* Tab Navigation */}
      <div className="border-b border-gray-200 mb-6">
        <nav className="-mb-px flex space-x-8">
          {tabs.map((tab) => (
            <button
              key={tab.key}
              onClick={() => { setActiveTab(tab.key); }}
              className={`whitespace-nowrap py-3 px-1 border-b-2 font-medium text-sm ${
                activeTab === tab.key
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              {tab.label}
            </button>
          ))}
        </nav>
      </div>

      {error && (
        <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded mb-4">
          {error}
        </div>
      )}

      {loading ? (
        <div className="text-center py-12 text-gray-500">Loading...</div>
      ) : (
        <>
          {activeTab === 'templates' && (
            <TemplatesTab
              templates={templates}
              total={templateTotal}
              page={templatePage}
              pageSize={pageSize}
              frameworkFilter={frameworkFilter}
              categoryFilter={categoryFilter}
              priorityFilter={priorityFilter}
              searchQuery={searchQuery}
              selectedTemplate={selectedTemplate}
              onPageChange={setTemplatePage}
              onFrameworkChange={setFrameworkFilter}
              onCategoryChange={setCategoryFilter}
              onPriorityChange={setPriorityFilter}
              onSearchChange={setSearchQuery}
              onSelectTemplate={setSelectedTemplate}
            />
          )}
          {activeTab === 'requirements' && (
            <RequirementsTab
              requirements={requirements}
              total={requirementTotal}
              page={requirementPage}
              pageSize={pageSize}
              onPageChange={setRequirementPage}
            />
          )}
          {activeTab === 'gaps' && gaps && <GapsTab gaps={gaps} />}
          {activeTab === 'schedule' && schedule && <ScheduleTab schedule={schedule} />}
        </>
      )}
    </div>
  );
}

// ============================================================
// TEMPLATES TAB
// ============================================================

function TemplatesTab({
  templates, total, page, pageSize,
  frameworkFilter, categoryFilter, priorityFilter, searchQuery,
  selectedTemplate,
  onPageChange, onFrameworkChange, onCategoryChange, onPriorityChange,
  onSearchChange, onSelectTemplate,
}: {
  templates: EvidenceTemplate[];
  total: number;
  page: number;
  pageSize: number;
  frameworkFilter: string;
  categoryFilter: string;
  priorityFilter: string;
  searchQuery: string;
  selectedTemplate: EvidenceTemplate | null;
  onPageChange: (p: number) => void;
  onFrameworkChange: (f: string) => void;
  onCategoryChange: (c: string) => void;
  onPriorityChange: (p: string) => void;
  onSearchChange: (s: string) => void;
  onSelectTemplate: (t: EvidenceTemplate | null) => void;
}) {
  const totalPages = Math.ceil(total / pageSize);

  if (selectedTemplate) {
    return (
      <div>
        <button onClick={() => onSelectTemplate(null)} className="text-blue-600 hover:text-blue-800 mb-4 text-sm">
          &larr; Back to templates
        </button>
        <TemplateDetail template={selectedTemplate} />
      </div>
    );
  }

  return (
    <div>
      {/* Filters */}
      <div className="grid grid-cols-1 md:grid-cols-5 gap-3 mb-4">
        <input
          type="text"
          placeholder="Search templates..."
          value={searchQuery}
          onChange={(e) => onSearchChange(e.target.value)}
          className="border rounded px-3 py-2 text-sm"
        />
        <select value={frameworkFilter} onChange={(e) => onFrameworkChange(e.target.value)} className="border rounded px-3 py-2 text-sm">
          <option value="">All Frameworks</option>
          <option value="ISO27001">ISO 27001</option>
          <option value="PCI_DSS_4">PCI DSS 4.0</option>
          <option value="NIST_800_53">NIST 800-53</option>
        </select>
        <select value={categoryFilter} onChange={(e) => onCategoryChange(e.target.value)} className="border rounded px-3 py-2 text-sm">
          <option value="">All Categories</option>
          {Object.entries(categoryLabels).map(([k, v]) => (
            <option key={k} value={k}>{v}</option>
          ))}
        </select>
        <select value={priorityFilter} onChange={(e) => onPriorityChange(e.target.value)} className="border rounded px-3 py-2 text-sm">
          <option value="">All Priorities</option>
          <option value="critical">Critical</option>
          <option value="high">High</option>
          <option value="medium">Medium</option>
          <option value="low">Low</option>
        </select>
        <div className="text-sm text-gray-500 flex items-center">{total} templates found</div>
      </div>

      {/* Template List */}
      <div className="bg-white shadow rounded-lg overflow-hidden">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Control</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Template Name</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Category</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Frequency</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Priority</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Difficulty</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Time</th>
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {templates.map((t) => (
              <tr key={t.id} className="hover:bg-gray-50 cursor-pointer" onClick={() => onSelectTemplate(t)}>
                <td className="px-4 py-3 text-sm">
                  <span className="font-mono text-blue-600">{t.framework_control_code}</span>
                  <br />
                  <span className="text-xs text-gray-400">{t.framework_code}</span>
                </td>
                <td className="px-4 py-3 text-sm font-medium text-gray-900 max-w-xs truncate">{t.name}</td>
                <td className="px-4 py-3 text-sm text-gray-500">{categoryLabels[t.evidence_category] || t.evidence_category}</td>
                <td className="px-4 py-3 text-sm text-gray-500 capitalize">{t.collection_frequency.replace('_', ' ')}</td>
                <td className="px-4 py-3 text-sm">
                  <Badge text={t.auditor_priority} className={priorityColors[t.auditor_priority] || 'bg-gray-100'} />
                </td>
                <td className="px-4 py-3 text-sm">
                  <Badge text={t.difficulty} className={difficultyColors[t.difficulty] || 'bg-gray-100'} />
                </td>
                <td className="px-4 py-3 text-sm text-gray-500">{t.typical_collection_time_minutes}m</td>
              </tr>
            ))}
            {templates.length === 0 && (
              <tr>
                <td colSpan={7} className="text-center py-8 text-gray-500">No templates found matching your filters.</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex justify-between items-center mt-4">
          <button
            onClick={() => onPageChange(Math.max(1, page - 1))}
            disabled={page <= 1}
            className="px-3 py-1 border rounded text-sm disabled:opacity-50"
          >
            Previous
          </button>
          <span className="text-sm text-gray-600">Page {page} of {totalPages}</span>
          <button
            onClick={() => onPageChange(Math.min(totalPages, page + 1))}
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
// TEMPLATE DETAIL
// ============================================================

function TemplateDetail({ template: t }: { template: EvidenceTemplate }) {
  return (
    <div className="bg-white shadow rounded-lg p-6">
      <div className="flex justify-between items-start mb-4">
        <div>
          <h2 className="text-xl font-bold text-gray-900">{t.name}</h2>
          <p className="text-sm text-gray-500 mt-1">
            <span className="font-mono">{t.framework_control_code}</span> ({t.framework_code})
          </p>
        </div>
        <div className="flex gap-2">
          <Badge text={t.auditor_priority} className={priorityColors[t.auditor_priority] || 'bg-gray-100'} />
          <Badge text={t.difficulty} className={difficultyColors[t.difficulty] || 'bg-gray-100'} />
          {t.is_system && <Badge text="System" className="bg-blue-100 text-blue-800" />}
        </div>
      </div>

      <p className="text-gray-700 mb-6">{t.description}</p>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
        <div className="bg-gray-50 rounded p-3">
          <div className="text-xs text-gray-500 uppercase mb-1">Category</div>
          <div className="text-sm font-medium">{categoryLabels[t.evidence_category] || t.evidence_category}</div>
        </div>
        <div className="bg-gray-50 rounded p-3">
          <div className="text-xs text-gray-500 uppercase mb-1">Collection Method</div>
          <div className="text-sm font-medium capitalize">{t.collection_method.replace(/_/g, ' ')}</div>
        </div>
        <div className="bg-gray-50 rounded p-3">
          <div className="text-xs text-gray-500 uppercase mb-1">Collection Frequency</div>
          <div className="text-sm font-medium capitalize">{t.collection_frequency.replace(/_/g, ' ')}</div>
        </div>
      </div>

      {t.collection_instructions && (
        <div className="mb-6">
          <h3 className="text-sm font-semibold text-gray-900 mb-2">Collection Instructions</h3>
          <p className="text-sm text-gray-700 bg-blue-50 border border-blue-100 rounded p-4 whitespace-pre-wrap">
            {t.collection_instructions}
          </p>
        </div>
      )}

      {t.acceptance_criteria && (
        <div className="mb-6">
          <h3 className="text-sm font-semibold text-gray-900 mb-2">Acceptance Criteria</h3>
          <p className="text-sm text-gray-700 bg-green-50 border border-green-100 rounded p-3">{t.acceptance_criteria}</p>
        </div>
      )}

      {t.common_rejection_reasons && t.common_rejection_reasons.length > 0 && (
        <div className="mb-6">
          <h3 className="text-sm font-semibold text-gray-900 mb-2">Common Rejection Reasons</h3>
          <ul className="text-sm text-gray-700 space-y-1">
            {t.common_rejection_reasons.map((reason, i) => (
              <li key={i} className="flex items-start">
                <span className="text-red-500 mr-2">!</span> {reason}
              </li>
            ))}
          </ul>
        </div>
      )}

      {t.validation_rules && t.validation_rules.length > 0 && (
        <div className="mb-6">
          <h3 className="text-sm font-semibold text-gray-900 mb-2">Validation Rules</h3>
          <div className="space-y-1">
            {t.validation_rules.map((rule: any, i: number) => (
              <div key={i} className="text-sm bg-gray-50 rounded px-3 py-2">
                <span className="font-mono text-blue-600">{rule.rule_type}</span>
                {rule.params && <span className="text-gray-500 ml-2">{JSON.stringify(rule.params)}</span>}
              </div>
            ))}
          </div>
        </div>
      )}

      {t.tags && t.tags.length > 0 && (
        <div>
          <h3 className="text-sm font-semibold text-gray-900 mb-2">Tags</h3>
          <div className="flex flex-wrap gap-1">
            {t.tags.map((tag, i) => (
              <Badge key={i} text={tag} className="bg-gray-100 text-gray-700" />
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

// ============================================================
// REQUIREMENTS TAB
// ============================================================

function RequirementsTab({
  requirements, total, page, pageSize, onPageChange,
}: {
  requirements: EvidenceRequirement[];
  total: number;
  page: number;
  pageSize: number;
  onPageChange: (p: number) => void;
}) {
  const totalPages = Math.ceil(total / pageSize);

  return (
    <div>
      <div className="flex justify-between items-center mb-4">
        <div className="text-sm text-gray-500">{total} requirements</div>
      </div>

      <div className="bg-white shadow rounded-lg overflow-hidden">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Control</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Evidence Template</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Validation</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Due Date</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Mandatory</th>
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {requirements.map((req) => (
              <tr key={req.id} className="hover:bg-gray-50">
                <td className="px-4 py-3 text-sm">
                  <span className="font-mono text-blue-600">{req.framework_control_code}</span>
                  <br />
                  <span className="text-xs text-gray-400">{req.framework_code}</span>
                </td>
                <td className="px-4 py-3 text-sm text-gray-900 max-w-sm truncate">{req.template_name}</td>
                <td className="px-4 py-3 text-sm">
                  <Badge text={req.status} className={statusColors[req.status] || 'bg-gray-100'} />
                </td>
                <td className="px-4 py-3 text-sm">
                  <Badge
                    text={req.validation_status}
                    className={req.validation_status === 'pass' ? 'bg-green-100 text-green-800' : req.validation_status === 'fail' ? 'bg-red-100 text-red-800' : 'bg-gray-100 text-gray-600'}
                  />
                </td>
                <td className="px-4 py-3 text-sm text-gray-500">
                  {req.due_date ? new Date(req.due_date).toLocaleDateString() : req.next_collection_due ? new Date(req.next_collection_due).toLocaleDateString() : '-'}
                </td>
                <td className="px-4 py-3 text-sm">{req.is_mandatory ? 'Yes' : 'No'}</td>
              </tr>
            ))}
            {requirements.length === 0 && (
              <tr>
                <td colSpan={6} className="text-center py-8 text-gray-500">
                  No requirements found. Generate requirements from a framework to get started.
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      {totalPages > 1 && (
        <div className="flex justify-between items-center mt-4">
          <button onClick={() => onPageChange(Math.max(1, page - 1))} disabled={page <= 1} className="px-3 py-1 border rounded text-sm disabled:opacity-50">Previous</button>
          <span className="text-sm text-gray-600">Page {page} of {totalPages}</span>
          <button onClick={() => onPageChange(Math.min(totalPages, page + 1))} disabled={page >= totalPages} className="px-3 py-1 border rounded text-sm disabled:opacity-50">Next</button>
        </div>
      )}
    </div>
  );
}

// ============================================================
// GAPS TAB
// ============================================================

function GapsTab({ gaps }: { gaps: EvidenceGaps }) {
  return (
    <div className="space-y-6">
      {/* Summary Cards */}
      <div className="grid grid-cols-2 md:grid-cols-6 gap-4">
        <SummaryCard title="Total" value={gaps.total_requirements} color="text-gray-900" />
        <SummaryCard title="Collected" value={gaps.collected} color="text-green-600" />
        <SummaryCard title="Pending" value={gaps.pending} color="text-yellow-600" />
        <SummaryCard title="Overdue" value={gaps.overdue} color="text-red-600" />
        <SummaryCard title="Validated" value={gaps.validated} color="text-emerald-600" />
        <SummaryCard title="Coverage" value={`${gaps.coverage_percent.toFixed(1)}%`} color="text-blue-600" />
      </div>

      {/* Coverage by Framework */}
      {gaps.gaps_by_framework.length > 0 && (
        <div className="bg-white shadow rounded-lg p-6">
          <h3 className="text-lg font-semibold text-gray-900 mb-4">Coverage by Framework</h3>
          <div className="space-y-3">
            {gaps.gaps_by_framework.map((fg) => (
              <div key={fg.framework_code} className="flex items-center gap-4">
                <span className="text-sm font-medium w-28">{fg.framework_code}</span>
                <div className="flex-1 bg-gray-200 rounded-full h-4">
                  <div
                    className={`h-4 rounded-full ${fg.coverage_percent >= 80 ? 'bg-green-500' : fg.coverage_percent >= 50 ? 'bg-yellow-500' : 'bg-red-500'}`}
                    style={{ width: `${Math.min(100, fg.coverage_percent)}%` }}
                  />
                </div>
                <span className="text-sm text-gray-600 w-24 text-right">
                  {fg.collected}/{fg.total_required} ({fg.coverage_percent.toFixed(1)}%)
                </span>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Critical Gaps */}
      {gaps.critical_gaps.length > 0 && (
        <div className="bg-white shadow rounded-lg p-6">
          <h3 className="text-lg font-semibold text-gray-900 mb-4">Critical Gaps</h3>
          <div className="space-y-2">
            {gaps.critical_gaps.map((cg) => (
              <div key={cg.requirement_id} className="flex items-center justify-between bg-red-50 rounded p-3">
                <div>
                  <span className="font-mono text-sm text-blue-600">{cg.framework_control_code}</span>
                  <span className="text-sm text-gray-700 ml-2">{cg.template_name}</span>
                </div>
                <div className="flex items-center gap-2">
                  <Badge text={cg.auditor_priority} className={priorityColors[cg.auditor_priority]} />
                  {cg.days_overdue > 0 && <span className="text-xs text-red-600">{cg.days_overdue}d overdue</span>}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Overdue Items */}
      {gaps.overdue_items.length > 0 && (
        <div className="bg-white shadow rounded-lg p-6">
          <h3 className="text-lg font-semibold text-gray-900 mb-4">Overdue Evidence ({gaps.overdue_items.length})</h3>
          <div className="space-y-2">
            {gaps.overdue_items.map((oi) => (
              <div key={oi.requirement_id} className="flex items-center justify-between border-b pb-2">
                <div>
                  <span className="font-mono text-sm text-blue-600">{oi.framework_control_code}</span>
                  <span className="text-sm text-gray-700 ml-2">{oi.template_name}</span>
                </div>
                <span className="text-sm text-red-600 font-medium">{oi.days_overdue} days overdue</span>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

// ============================================================
// SCHEDULE TAB
// ============================================================

function ScheduleTab({ schedule }: { schedule: CollectionSchedule }) {
  return (
    <div className="space-y-6">
      <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
        <p className="text-sm text-blue-800">
          <strong>{schedule.total_scheduled}</strong> evidence collection tasks scheduled in the next 90 days.
        </p>
      </div>

      <ScheduleSection title="Due This Week" items={schedule.upcoming_this_week} urgency="high" />
      <ScheduleSection title="Due This Month" items={schedule.upcoming_this_month} urgency="medium" />
      <ScheduleSection title="Due This Quarter" items={schedule.upcoming_this_quarter} urgency="low" />
    </div>
  );
}

function ScheduleSection({ title, items, urgency }: { title: string; items: ScheduleItem[]; urgency: string }) {
  if (items.length === 0) return null;

  const borderColor = urgency === 'high' ? 'border-red-200' : urgency === 'medium' ? 'border-yellow-200' : 'border-gray-200';

  return (
    <div className={`bg-white shadow rounded-lg p-6 border-l-4 ${borderColor}`}>
      <h3 className="text-lg font-semibold text-gray-900 mb-4">{title} ({items.length})</h3>
      <div className="space-y-3">
        {items.map((item) => (
          <div key={item.requirement_id} className="flex items-center justify-between py-2 border-b border-gray-100 last:border-0">
            <div>
              <span className="font-mono text-sm text-blue-600">{item.framework_control_code}</span>
              <span className="text-sm text-gray-500 ml-1">({item.framework_code})</span>
              <p className="text-sm text-gray-800">{item.template_name}</p>
            </div>
            <div className="flex items-center gap-3 text-sm">
              <Badge text={item.difficulty} className={difficultyColors[item.difficulty] || 'bg-gray-100'} />
              <span className="text-gray-500">{item.estimated_minutes}m</span>
              {item.next_due && <span className="text-gray-600">{new Date(item.next_due).toLocaleDateString()}</span>}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

// ============================================================
// SHARED COMPONENTS
// ============================================================

function SummaryCard({ title, value, color }: { title: string; value: number | string; color: string }) {
  return (
    <div className="bg-white shadow rounded-lg p-4">
      <div className="text-xs text-gray-500 uppercase">{title}</div>
      <div className={`text-2xl font-bold ${color}`}>{value}</div>
    </div>
  );
}
