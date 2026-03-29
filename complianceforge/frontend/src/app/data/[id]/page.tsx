'use client';

import { useEffect, useState, useCallback } from 'react';
import { useParams, useRouter } from 'next/navigation';
import Link from 'next/link';
import api from '@/lib/api';

// ============================================================
// TYPE DEFINITIONS
// ============================================================

interface ProcessingActivity {
  id: string;
  activity_ref: string;
  name: string;
  description: string;
  purpose: string;
  legal_basis: string;
  legal_basis_detail: string;
  status: string;
  role: string;
  joint_controller_details: string;
  data_subject_categories: string[];
  estimated_data_subjects_count: number | null;
  data_category_ids: string[];
  special_categories_processed: boolean;
  special_categories_legal_basis: string;
  recipient_categories: string[];
  recipient_vendor_ids: string[];
  involves_international_transfer: boolean;
  transfer_countries: string[];
  transfer_safeguards: string;
  transfer_safeguards_detail: string;
  tia_conducted: boolean;
  tia_date: string | null;
  tia_document_path: string;
  retention_period_months: number | null;
  retention_justification: string;
  deletion_method: string;
  deletion_responsible_user_id: string | null;
  system_ids: string[];
  storage_locations: string[];
  dpia_required: boolean;
  dpia_status: string;
  dpia_document_path: string;
  dpia_conducted_date: string | null;
  security_measures: string;
  linked_control_codes: string[];
  risk_level: string;
  data_steward_user_id: string | null;
  department: string;
  process_owner_user_id: string | null;
  last_review_date: string | null;
  next_review_date: string | null;
  review_frequency_months: number;
  created_at: string;
  updated_at: string;
}

interface DataFlowMap {
  id: string;
  processing_activity_id: string;
  name: string;
  flow_type: string;
  source_type: string;
  source_name: string;
  destination_type: string;
  destination_name: string;
  destination_country: string;
  data_category_ids: string[];
  transfer_method: string;
  encryption_in_transit: boolean;
  encryption_at_rest: boolean;
  volume_description: string;
  frequency: string;
  legal_basis: string;
  notes: string;
  sort_order: number;
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

const ROLE_LABELS: Record<string, string> = {
  controller: 'Controller',
  joint_controller: 'Joint Controller',
  processor: 'Processor',
};

const DPIA_STATUS_COLORS: Record<string, string> = {
  not_required: 'bg-gray-100 text-gray-600',
  required: 'bg-red-100 text-red-800',
  in_progress: 'bg-yellow-100 text-yellow-800',
  completed: 'bg-green-100 text-green-800',
  review_needed: 'bg-orange-100 text-orange-800',
};

const RISK_COLORS: Record<string, string> = {
  critical: 'bg-red-100 text-red-800',
  high: 'bg-orange-100 text-orange-800',
  medium: 'bg-yellow-100 text-yellow-800',
  low: 'bg-green-100 text-green-800',
};

const SAFEGUARD_LABELS: Record<string, string> = {
  adequacy_decision: 'EU Adequacy Decision',
  standard_contractual_clauses: 'Standard Contractual Clauses (SCCs)',
  binding_corporate_rules: 'Binding Corporate Rules (BCRs)',
  approved_code_of_conduct: 'Approved Code of Conduct',
  approved_certification: 'Approved Certification Mechanism',
  derogation_art49: 'Derogation under Art. 49',
  none: 'None',
};

const FLOW_TYPE_COLORS: Record<string, string> = {
  collection: 'bg-blue-100 text-blue-700',
  storage: 'bg-gray-100 text-gray-700',
  processing: 'bg-green-100 text-green-700',
  transfer: 'bg-orange-100 text-orange-700',
  sharing: 'bg-purple-100 text-purple-700',
  deletion: 'bg-red-100 text-red-700',
};

type SectionType = 'overview' | 'dataflows' | 'dpia' | 'transfers' | 'retention';

// ============================================================
// COMPONENT
// ============================================================

export default function ProcessingActivityDetailPage() {
  const params = useParams();
  const router = useRouter();
  const activityId = params.id as string;

  const [activity, setActivity] = useState<ProcessingActivity | null>(null);
  const [flows, setFlows] = useState<DataFlowMap[]>([]);
  const [activeSection, setActiveSection] = useState<SectionType>('overview');
  const [loading, setLoading] = useState(true);
  const [editing, setEditing] = useState(false);

  // Edit form state
  const [editName, setEditName] = useState('');
  const [editDescription, setEditDescription] = useState('');
  const [editPurpose, setEditPurpose] = useState('');
  const [editStatus, setEditStatus] = useState('');
  const [editDepartment, setEditDepartment] = useState('');

  const loadActivity = useCallback(async () => {
    try {
      const res = await api.getProcessingActivity(activityId);
      if (res.data) {
        setActivity(res.data);
        setEditName(res.data.name);
        setEditDescription(res.data.description);
        setEditPurpose(res.data.purpose);
        setEditStatus(res.data.status);
        setEditDepartment(res.data.department);
      }
    } catch (err) {
      console.error('Failed to load activity:', err);
    }
  }, [activityId]);

  const loadFlows = useCallback(async () => {
    try {
      const res = await api.getActivityDataFlows(activityId);
      if (res.data) setFlows(res.data);
    } catch (err) {
      console.error('Failed to load data flows:', err);
    }
  }, [activityId]);

  useEffect(() => {
    const load = async () => {
      setLoading(true);
      await Promise.all([loadActivity(), loadFlows()]);
      setLoading(false);
    };
    load();
  }, [loadActivity, loadFlows]);

  const handleSave = async () => {
    if (!activity) return;
    try {
      await api.updateProcessingActivity(activityId, {
        name: editName,
        description: editDescription,
        purpose: editPurpose,
        status: editStatus,
        department: editDepartment,
      });
      setEditing(false);
      loadActivity();
    } catch (err) {
      console.error('Failed to update activity:', err);
    }
  };

  // Art. 30 completeness calculation
  const calculateCompleteness = (a: ProcessingActivity): { score: number; missing: string[] } => {
    const checks: [string, boolean][] = [
      ['Name', !!a.name],
      ['Description', !!a.description],
      ['Purpose', !!a.purpose],
      ['Legal Basis', !!a.legal_basis],
      ['Data Subject Categories', a.data_subject_categories?.length > 0],
      ['Data Categories', a.data_category_ids?.length > 0],
      ['Recipient Categories', a.recipient_categories?.length > 0],
      ['Retention Period', a.retention_period_months !== null],
      ['Security Measures', !!a.security_measures],
      ['Controller/Processor Role', !!a.role],
      ['Process Owner', !!a.process_owner_user_id],
      ['Department', !!a.department],
    ];

    const present = checks.filter(([, ok]) => ok).length;
    const missing = checks.filter(([, ok]) => !ok).map(([name]) => name);
    return { score: Math.round((present / checks.length) * 100), missing };
  };

  if (loading || !activity) {
    return (
      <div className="p-8">
        <div className="animate-pulse space-y-4">
          <div className="h-6 bg-gray-200 rounded w-1/4"></div>
          <div className="h-4 bg-gray-200 rounded w-1/2"></div>
          <div className="h-64 bg-gray-200 rounded"></div>
        </div>
      </div>
    );
  }

  const { score: completeness, missing } = calculateCompleteness(activity);

  const sections: { key: SectionType; label: string }[] = [
    { key: 'overview', label: 'Overview' },
    { key: 'dataflows', label: `Data Flows (${flows.length})` },
    { key: 'dpia', label: 'DPIA' },
    { key: 'transfers', label: 'Transfers' },
    { key: 'retention', label: 'Retention & Deletion' },
  ];

  return (
    <div className="p-6 max-w-7xl mx-auto">
      {/* Breadcrumb + Header */}
      <div className="mb-6">
        <nav className="text-sm text-gray-500 mb-2">
          <Link href="/data" className="hover:text-blue-600">Data Governance</Link>
          <span className="mx-2">/</span>
          <span className="text-gray-900">{activity.activity_ref}</span>
        </nav>

        <div className="flex items-start justify-between">
          <div>
            <div className="flex items-center gap-3">
              <h1 className="text-2xl font-bold text-gray-900">{activity.name}</h1>
              <span className={`px-2 py-1 rounded text-xs font-medium ${STATUS_COLORS[activity.status] || 'bg-gray-100'}`}>
                {activity.status}
              </span>
              <span className={`px-2 py-1 rounded text-xs font-medium ${RISK_COLORS[activity.risk_level] || 'bg-gray-100'}`}>
                {activity.risk_level} risk
              </span>
            </div>
            <p className="text-sm text-gray-500 mt-1">
              {activity.activity_ref} | {ROLE_LABELS[activity.role] || activity.role}
              {activity.department && ` | ${activity.department}`}
            </p>
          </div>
          <div className="flex gap-2">
            {!editing ? (
              <button onClick={() => setEditing(true)} className="px-4 py-2 bg-blue-600 text-white text-sm rounded hover:bg-blue-700">
                Edit
              </button>
            ) : (
              <>
                <button onClick={handleSave} className="px-4 py-2 bg-green-600 text-white text-sm rounded hover:bg-green-700">
                  Save
                </button>
                <button onClick={() => setEditing(false)} className="px-4 py-2 bg-gray-200 text-gray-700 text-sm rounded hover:bg-gray-300">
                  Cancel
                </button>
              </>
            )}
          </div>
        </div>
      </div>

      {/* Completeness Bar */}
      <div className="bg-white border rounded-lg p-4 mb-6">
        <div className="flex justify-between items-center mb-2">
          <span className="text-sm font-medium text-gray-700">Art. 30 Completeness</span>
          <span className={`text-sm font-bold ${completeness >= 80 ? 'text-green-600' : completeness >= 50 ? 'text-yellow-600' : 'text-red-600'}`}>
            {completeness}%
          </span>
        </div>
        <div className="w-full bg-gray-100 rounded-full h-3">
          <div
            className={`rounded-full h-3 transition-all ${completeness >= 80 ? 'bg-green-500' : completeness >= 50 ? 'bg-yellow-500' : 'bg-red-500'}`}
            style={{ width: `${completeness}%` }}
          />
        </div>
        {missing.length > 0 && (
          <p className="text-xs text-gray-500 mt-2">
            Missing: {missing.join(', ')}
          </p>
        )}
      </div>

      {/* Section Tabs */}
      <div className="border-b border-gray-200 mb-6">
        <nav className="flex space-x-6">
          {sections.map(s => (
            <button
              key={s.key}
              onClick={() => setActiveSection(s.key)}
              className={`py-3 px-1 border-b-2 text-sm font-medium transition-colors ${
                activeSection === s.key
                  ? 'border-blue-600 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700'
              }`}
            >
              {s.label}
            </button>
          ))}
        </nav>
      </div>

      {/* Section Content */}
      {activeSection === 'overview' && (
        <OverviewSection
          activity={activity}
          editing={editing}
          editName={editName} setEditName={setEditName}
          editDescription={editDescription} setEditDescription={setEditDescription}
          editPurpose={editPurpose} setEditPurpose={setEditPurpose}
          editStatus={editStatus} setEditStatus={setEditStatus}
          editDepartment={editDepartment} setEditDepartment={setEditDepartment}
        />
      )}
      {activeSection === 'dataflows' && (
        <DataFlowsSection flows={flows} activityId={activityId} onReload={loadFlows} />
      )}
      {activeSection === 'dpia' && (
        <DPIASection activity={activity} />
      )}
      {activeSection === 'transfers' && (
        <TransfersSection activity={activity} />
      )}
      {activeSection === 'retention' && (
        <RetentionSection activity={activity} />
      )}
    </div>
  );
}

// ============================================================
// OVERVIEW SECTION
// ============================================================

function OverviewSection({ activity, editing, editName, setEditName, editDescription, setEditDescription, editPurpose, setEditPurpose, editStatus, setEditStatus, editDepartment, setEditDepartment }: {
  activity: ProcessingActivity;
  editing: boolean;
  editName: string; setEditName: (v: string) => void;
  editDescription: string; setEditDescription: (v: string) => void;
  editPurpose: string; setEditPurpose: (v: string) => void;
  editStatus: string; setEditStatus: (v: string) => void;
  editDepartment: string; setEditDepartment: (v: string) => void;
}) {
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
      {/* Basic Information */}
      <div className="bg-white border rounded-lg p-5">
        <h3 className="text-sm font-semibold text-gray-700 mb-4">Basic Information</h3>
        <div className="space-y-3">
          <FieldRow label="Name" editing={editing}>
            {editing ? (
              <input value={editName} onChange={e => setEditName(e.target.value)} className="w-full px-2 py-1 border rounded text-sm" />
            ) : (
              <span className="text-sm">{activity.name}</span>
            )}
          </FieldRow>
          <FieldRow label="Description" editing={editing}>
            {editing ? (
              <textarea value={editDescription} onChange={e => setEditDescription(e.target.value)} className="w-full px-2 py-1 border rounded text-sm" rows={3} />
            ) : (
              <span className="text-sm text-gray-600">{activity.description || '-'}</span>
            )}
          </FieldRow>
          <FieldRow label="Purpose" editing={editing}>
            {editing ? (
              <textarea value={editPurpose} onChange={e => setEditPurpose(e.target.value)} className="w-full px-2 py-1 border rounded text-sm" rows={2} />
            ) : (
              <span className="text-sm text-gray-600">{activity.purpose || '-'}</span>
            )}
          </FieldRow>
          <FieldRow label="Status" editing={editing}>
            {editing ? (
              <select value={editStatus} onChange={e => setEditStatus(e.target.value)} className="px-2 py-1 border rounded text-sm">
                <option value="draft">Draft</option>
                <option value="active">Active</option>
                <option value="under_review">Under Review</option>
                <option value="suspended">Suspended</option>
                <option value="retired">Retired</option>
              </select>
            ) : (
              <span className={`px-2 py-1 rounded text-xs font-medium ${STATUS_COLORS[activity.status]}`}>{activity.status}</span>
            )}
          </FieldRow>
          <FieldRow label="Department" editing={editing}>
            {editing ? (
              <input value={editDepartment} onChange={e => setEditDepartment(e.target.value)} className="w-full px-2 py-1 border rounded text-sm" />
            ) : (
              <span className="text-sm">{activity.department || '-'}</span>
            )}
          </FieldRow>
        </div>
      </div>

      {/* Legal Basis & Role */}
      <div className="bg-white border rounded-lg p-5">
        <h3 className="text-sm font-semibold text-gray-700 mb-4">Legal Basis & Role</h3>
        <div className="space-y-3">
          <FieldRow label="Legal Basis">
            <span className="text-sm">{LEGAL_BASIS_LABELS[activity.legal_basis] || activity.legal_basis || '-'}</span>
          </FieldRow>
          {activity.legal_basis_detail && (
            <FieldRow label="Detail">
              <span className="text-sm text-gray-600">{activity.legal_basis_detail}</span>
            </FieldRow>
          )}
          <FieldRow label="Role">
            <span className="text-sm">{ROLE_LABELS[activity.role] || activity.role}</span>
          </FieldRow>
          {activity.joint_controller_details && (
            <FieldRow label="Joint Controller Details">
              <span className="text-sm text-gray-600">{activity.joint_controller_details}</span>
            </FieldRow>
          )}
        </div>

        <h3 className="text-sm font-semibold text-gray-700 mt-6 mb-4">Data Subjects</h3>
        <div className="space-y-3">
          <FieldRow label="Categories">
            <div className="flex flex-wrap gap-1">
              {activity.data_subject_categories?.map(c => (
                <span key={c} className="px-2 py-0.5 bg-blue-50 text-blue-700 rounded text-xs">{c}</span>
              ))}
              {(!activity.data_subject_categories || activity.data_subject_categories.length === 0) && (
                <span className="text-sm text-gray-400">Not specified</span>
              )}
            </div>
          </FieldRow>
          <FieldRow label="Estimated Count">
            <span className="text-sm">{activity.estimated_data_subjects_count?.toLocaleString() || '-'}</span>
          </FieldRow>
          {activity.special_categories_processed && (
            <>
              <FieldRow label="Special Categories">
                <span className="px-2 py-0.5 bg-purple-100 text-purple-700 rounded text-xs">Art. 9 Data Processed</span>
              </FieldRow>
              {activity.special_categories_legal_basis && (
                <FieldRow label="Art. 9 Legal Basis">
                  <span className="text-sm text-gray-600">{activity.special_categories_legal_basis}</span>
                </FieldRow>
              )}
            </>
          )}
        </div>
      </div>

      {/* Recipients & Security */}
      <div className="bg-white border rounded-lg p-5">
        <h3 className="text-sm font-semibold text-gray-700 mb-4">Recipients</h3>
        <div className="space-y-3">
          <FieldRow label="Recipient Categories">
            <div className="flex flex-wrap gap-1">
              {activity.recipient_categories?.map(r => (
                <span key={r} className="px-2 py-0.5 bg-gray-100 text-gray-700 rounded text-xs">{r}</span>
              ))}
              {(!activity.recipient_categories || activity.recipient_categories.length === 0) && (
                <span className="text-sm text-gray-400">Not specified</span>
              )}
            </div>
          </FieldRow>
        </div>
      </div>

      {/* Security & Controls */}
      <div className="bg-white border rounded-lg p-5">
        <h3 className="text-sm font-semibold text-gray-700 mb-4">Security Measures</h3>
        <p className="text-sm text-gray-600">{activity.security_measures || 'Not documented'}</p>
        {activity.linked_control_codes && activity.linked_control_codes.length > 0 && (
          <div className="mt-3">
            <span className="text-xs text-gray-500">Linked Controls: </span>
            <div className="flex flex-wrap gap-1 mt-1">
              {activity.linked_control_codes.map(c => (
                <span key={c} className="px-2 py-0.5 bg-blue-50 text-blue-700 rounded text-xs font-mono">{c}</span>
              ))}
            </div>
          </div>
        )}

        <h3 className="text-sm font-semibold text-gray-700 mt-6 mb-3">Review Schedule</h3>
        <div className="space-y-2 text-sm text-gray-600">
          <div className="flex justify-between">
            <span>Frequency</span>
            <span>Every {activity.review_frequency_months} months</span>
          </div>
          <div className="flex justify-between">
            <span>Last Review</span>
            <span>{activity.last_review_date ? new Date(activity.last_review_date).toLocaleDateString() : 'Never'}</span>
          </div>
          <div className="flex justify-between">
            <span>Next Review</span>
            <span className={activity.next_review_date && new Date(activity.next_review_date) < new Date() ? 'text-red-600 font-medium' : ''}>
              {activity.next_review_date ? new Date(activity.next_review_date).toLocaleDateString() : '-'}
            </span>
          </div>
        </div>
      </div>
    </div>
  );
}

// ============================================================
// DATA FLOWS SECTION
// ============================================================

function DataFlowsSection({ flows, activityId, onReload }: { flows: DataFlowMap[]; activityId: string; onReload: () => void }) {
  return (
    <div className="space-y-4">
      <div className="flex justify-between items-center">
        <h3 className="text-lg font-semibold">Data Flow Diagram</h3>
        <span className="text-sm text-gray-500">{flows.length} flows mapped</span>
      </div>

      {/* Visual Flow Diagram */}
      {flows.length > 0 && (
        <div className="bg-white border rounded-lg p-6">
          <div className="space-y-4">
            {flows.map((f, idx) => (
              <div key={f.id} className="flex items-center gap-3">
                <div className="flex-shrink-0 px-3 py-2 bg-blue-50 border border-blue-200 rounded text-sm text-center min-w-[120px]">
                  <div className="text-xs text-blue-500">{f.source_type}</div>
                  <div className="font-medium text-blue-800">{f.source_name}</div>
                </div>
                <div className="flex-1 flex flex-col items-center">
                  <span className={`px-2 py-0.5 rounded text-xs font-medium mb-1 ${FLOW_TYPE_COLORS[f.flow_type] || 'bg-gray-100'}`}>
                    {f.flow_type}
                  </span>
                  <div className="w-full border-t-2 border-dashed border-gray-300 relative">
                    <div className="absolute right-0 top-[-5px] text-gray-400">&#9654;</div>
                  </div>
                  <div className="flex gap-1 mt-1">
                    {f.encryption_in_transit && (
                      <span className="px-1 py-0.5 bg-green-50 text-green-600 rounded text-xs">Encrypted</span>
                    )}
                    {f.frequency && (
                      <span className="px-1 py-0.5 bg-gray-50 text-gray-500 rounded text-xs">{f.frequency}</span>
                    )}
                  </div>
                </div>
                <div className="flex-shrink-0 px-3 py-2 bg-orange-50 border border-orange-200 rounded text-sm text-center min-w-[120px]">
                  <div className="text-xs text-orange-500">{f.destination_type}</div>
                  <div className="font-medium text-orange-800">{f.destination_name}</div>
                  {f.destination_country && (
                    <div className="text-xs text-orange-400 mt-0.5">{f.destination_country}</div>
                  )}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Flow Table */}
      <div className="bg-white border rounded-lg overflow-hidden">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Flow</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Type</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Source</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Destination</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Method</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Encryption</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200">
            {flows.map(f => (
              <tr key={f.id} className="hover:bg-gray-50">
                <td className="px-4 py-3 text-sm font-medium">{f.name}</td>
                <td className="px-4 py-3">
                  <span className={`px-2 py-1 rounded text-xs ${FLOW_TYPE_COLORS[f.flow_type] || 'bg-gray-100'}`}>{f.flow_type}</span>
                </td>
                <td className="px-4 py-3 text-sm text-gray-600">{f.source_name}</td>
                <td className="px-4 py-3 text-sm text-gray-600">
                  {f.destination_name}
                  {f.destination_country && <span className="ml-1 text-xs text-blue-500">({f.destination_country})</span>}
                </td>
                <td className="px-4 py-3 text-sm text-gray-500">{f.transfer_method || '-'}</td>
                <td className="px-4 py-3">
                  <div className="flex gap-1">
                    {f.encryption_in_transit && <span className="px-1.5 py-0.5 bg-green-100 text-green-700 rounded text-xs">Transit</span>}
                    {f.encryption_at_rest && <span className="px-1.5 py-0.5 bg-green-100 text-green-700 rounded text-xs">At Rest</span>}
                    {!f.encryption_in_transit && !f.encryption_at_rest && <span className="text-xs text-gray-400">None</span>}
                  </div>
                </td>
              </tr>
            ))}
            {flows.length === 0 && (
              <tr>
                <td colSpan={6} className="px-4 py-8 text-center text-gray-400 text-sm">
                  No data flows mapped yet. Add flows to visualize data movement.
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
// DPIA SECTION
// ============================================================

function DPIASection({ activity }: { activity: ProcessingActivity }) {
  return (
    <div className="space-y-6">
      <div className="bg-white border rounded-lg p-5">
        <h3 className="text-sm font-semibold text-gray-700 mb-4">Data Protection Impact Assessment</h3>
        <div className="grid grid-cols-2 gap-4">
          <div>
            <span className="text-xs text-gray-500">DPIA Required</span>
            <p className="text-sm font-medium mt-1">
              {activity.dpia_required ? (
                <span className="px-2 py-1 bg-red-100 text-red-700 rounded">Yes</span>
              ) : (
                <span className="px-2 py-1 bg-green-100 text-green-700 rounded">No</span>
              )}
            </p>
          </div>
          <div>
            <span className="text-xs text-gray-500">DPIA Status</span>
            <p className="mt-1">
              <span className={`px-2 py-1 rounded text-xs font-medium ${DPIA_STATUS_COLORS[activity.dpia_status] || 'bg-gray-100'}`}>
                {activity.dpia_status.replace(/_/g, ' ')}
              </span>
            </p>
          </div>
          {activity.dpia_conducted_date && (
            <div>
              <span className="text-xs text-gray-500">Conducted Date</span>
              <p className="text-sm mt-1">{new Date(activity.dpia_conducted_date).toLocaleDateString()}</p>
            </div>
          )}
          {activity.dpia_document_path && (
            <div>
              <span className="text-xs text-gray-500">Document</span>
              <p className="text-sm text-blue-600 mt-1">{activity.dpia_document_path}</p>
            </div>
          )}
        </div>
      </div>

      {/* Risk Assessment */}
      <div className="bg-white border rounded-lg p-5">
        <h3 className="text-sm font-semibold text-gray-700 mb-4">Risk Indicators</h3>
        <div className="space-y-2">
          <RiskIndicator label="Special Category Data (Art. 9)" active={activity.special_categories_processed} />
          <RiskIndicator label="International Transfers" active={activity.involves_international_transfer} />
          <RiskIndicator label="Large-Scale Processing (&gt;10,000)" active={activity.estimated_data_subjects_count !== null && activity.estimated_data_subjects_count > 10000} />
          <RiskIndicator label="Children or Vulnerable Persons" active={activity.data_subject_categories?.some(c => c === 'children' || c === 'vulnerable_persons') || false} />
          <RiskIndicator label="Multiple Data Categories (&gt;5)" active={activity.data_category_ids?.length > 5} />
        </div>
      </div>
    </div>
  );
}

// ============================================================
// TRANSFERS SECTION
// ============================================================

function TransfersSection({ activity }: { activity: ProcessingActivity }) {
  return (
    <div className="space-y-6">
      <div className="bg-white border rounded-lg p-5">
        <h3 className="text-sm font-semibold text-gray-700 mb-4">International Transfer Details</h3>
        {activity.involves_international_transfer ? (
          <div className="space-y-4">
            <div>
              <span className="text-xs text-gray-500">Transfer Countries</span>
              <div className="flex flex-wrap gap-1 mt-1">
                {activity.transfer_countries?.map(c => (
                  <span key={c} className="px-2 py-1 bg-blue-50 text-blue-700 rounded text-sm">{c}</span>
                ))}
              </div>
            </div>
            <div>
              <span className="text-xs text-gray-500">Safeguards</span>
              <p className="text-sm mt-1">{SAFEGUARD_LABELS[activity.transfer_safeguards] || activity.transfer_safeguards || 'Not specified'}</p>
              {activity.transfer_safeguards_detail && (
                <p className="text-xs text-gray-500 mt-1">{activity.transfer_safeguards_detail}</p>
              )}
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <span className="text-xs text-gray-500">TIA Conducted</span>
                <p className="mt-1">
                  <span className={`px-2 py-1 rounded text-xs ${activity.tia_conducted ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'}`}>
                    {activity.tia_conducted ? 'Yes' : 'No'}
                  </span>
                </p>
              </div>
              {activity.tia_date && (
                <div>
                  <span className="text-xs text-gray-500">TIA Date</span>
                  <p className="text-sm mt-1">{new Date(activity.tia_date).toLocaleDateString()}</p>
                </div>
              )}
            </div>
          </div>
        ) : (
          <p className="text-sm text-gray-400">This activity does not involve international data transfers.</p>
        )}
      </div>
    </div>
  );
}

// ============================================================
// RETENTION SECTION
// ============================================================

function RetentionSection({ activity }: { activity: ProcessingActivity }) {
  return (
    <div className="space-y-6">
      <div className="bg-white border rounded-lg p-5">
        <h3 className="text-sm font-semibold text-gray-700 mb-4">Retention & Deletion</h3>
        <div className="space-y-3">
          <FieldRow label="Retention Period">
            <span className="text-sm">
              {activity.retention_period_months ? `${activity.retention_period_months} months` : 'Not specified'}
            </span>
          </FieldRow>
          <FieldRow label="Justification">
            <span className="text-sm text-gray-600">{activity.retention_justification || 'Not documented'}</span>
          </FieldRow>
          <FieldRow label="Deletion Method">
            <span className="text-sm text-gray-600">{activity.deletion_method || 'Not specified'}</span>
          </FieldRow>
          <FieldRow label="Storage Locations">
            <div className="flex flex-wrap gap-1">
              {activity.storage_locations?.map(l => (
                <span key={l} className="px-2 py-0.5 bg-gray-100 text-gray-700 rounded text-xs">{l}</span>
              ))}
              {(!activity.storage_locations || activity.storage_locations.length === 0) && (
                <span className="text-sm text-gray-400">Not specified</span>
              )}
            </div>
          </FieldRow>
        </div>
      </div>
    </div>
  );
}

// ============================================================
// SHARED COMPONENTS
// ============================================================

function FieldRow({ label, editing, children }: { label: string; editing?: boolean; children: React.ReactNode }) {
  return (
    <div className="flex flex-col sm:flex-row sm:items-start gap-1">
      <span className="text-xs text-gray-500 sm:w-40 sm:flex-shrink-0 pt-1">{label}</span>
      <div className="flex-1">{children}</div>
    </div>
  );
}

function RiskIndicator({ label, active }: { label: string; active: boolean }) {
  return (
    <div className="flex items-center justify-between py-2 px-3 rounded bg-gray-50">
      <span className="text-sm text-gray-700">{label}</span>
      <span className={`px-2 py-0.5 rounded text-xs font-medium ${active ? 'bg-red-100 text-red-700' : 'bg-gray-100 text-gray-500'}`}>
        {active ? 'Yes' : 'No'}
      </span>
    </div>
  );
}
