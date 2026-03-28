'use client';

import { useEffect, useState } from 'react';
import api from '@/lib/api';

interface WorkflowDefinition {
  id: string;
  organization_id: string | null;
  name: string;
  description: string;
  workflow_type: string;
  entity_type: string;
  version: number;
  status: string;
  trigger_conditions: Record<string, unknown>;
  sla_config: Record<string, unknown>;
  is_system: boolean;
  created_at: string;
  updated_at: string;
}

interface WorkflowStep {
  id: string;
  workflow_definition_id: string;
  step_order: number;
  name: string;
  description: string;
  step_type: string;
  approver_type: string | null;
  approval_mode: string;
  minimum_approvals: number;
  sla_hours: number | null;
  is_optional: boolean;
  can_delegate: boolean;
}

const STATUS_COLORS: Record<string, string> = {
  draft: 'badge-medium',
  active: 'badge-low',
  deprecated: 'badge-critical',
};

const TYPE_LABELS: Record<string, string> = {
  policy_approval: 'Policy Approval',
  risk_acceptance: 'Risk Acceptance',
  exception_request: 'Exception Request',
  audit_finding_remediation: 'Audit Finding Remediation',
  vendor_onboarding: 'Vendor Onboarding',
  control_change: 'Control Change',
  incident_response: 'Incident Response',
  access_review: 'Access Review',
  custom: 'Custom',
};

const ENTITY_LABELS: Record<string, string> = {
  policy: 'Policy',
  risk: 'Risk',
  exception: 'Exception',
  audit_finding: 'Audit Finding',
  vendor: 'Vendor',
  control: 'Control',
  incident: 'Incident',
};

const STEP_TYPE_COLORS: Record<string, string> = {
  approval: 'bg-blue-100 text-blue-800',
  review: 'bg-purple-100 text-purple-800',
  task: 'bg-green-100 text-green-800',
  notification: 'bg-yellow-100 text-yellow-800',
  condition: 'bg-orange-100 text-orange-800',
  parallel_gate: 'bg-pink-100 text-pink-800',
  timer: 'bg-gray-100 text-gray-800',
  auto_action: 'bg-cyan-100 text-cyan-800',
};

export default function WorkflowDefinitionsPage() {
  const [definitions, setDefinitions] = useState<WorkflowDefinition[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedDef, setSelectedDef] = useState<WorkflowDefinition | null>(null);
  const [steps, setSteps] = useState<WorkflowStep[]>([]);
  const [stepsLoading, setStepsLoading] = useState(false);
  const [showCreate, setShowCreate] = useState(false);
  const [createForm, setCreateForm] = useState({
    name: '',
    description: '',
    workflow_type: 'policy_approval',
    entity_type: 'policy',
  });

  const loadDefinitions = () => {
    setLoading(true);
    api.getWorkflowDefinitions()
      .then(res => setDefinitions(res.data || []))
      .finally(() => setLoading(false));
  };

  useEffect(() => { loadDefinitions(); }, []);

  const loadSteps = async (def: WorkflowDefinition) => {
    setSelectedDef(def);
    setStepsLoading(true);
    try {
      const res = await api.getWorkflowSteps(def.id);
      setSteps(res.data || []);
    } catch {
      setSteps([]);
    } finally {
      setStepsLoading(false);
    }
  };

  const handleCreate = async () => {
    if (!createForm.name.trim()) {
      alert('Name is required');
      return;
    }
    try {
      await api.createWorkflowDefinition(createForm);
      setShowCreate(false);
      setCreateForm({ name: '', description: '', workflow_type: 'policy_approval', entity_type: 'policy' });
      loadDefinitions();
    } catch (err: any) {
      alert(err.message || 'Failed to create definition');
    }
  };

  const handleActivate = async (defId: string) => {
    try {
      await api.activateWorkflowDefinition(defId);
      loadDefinitions();
      if (selectedDef?.id === defId) {
        setSelectedDef(prev => prev ? { ...prev, status: 'active' } : null);
      }
    } catch (err: any) {
      alert(err.message || 'Failed to activate definition');
    }
  };

  if (loading) {
    return (
      <div>
        <h1 className="text-2xl font-bold text-gray-900 mb-8">Workflow Definitions</h1>
        <div className="card animate-pulse h-96" />
      </div>
    );
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Workflow Definitions</h1>
          <p className="text-gray-500 mt-1">Manage compliance workflow templates, steps, and routing logic</p>
        </div>
        <button onClick={() => setShowCreate(true)} className="btn-primary">New Definition</button>
      </div>

      {/* Summary */}
      <div className="grid grid-cols-1 gap-4 md:grid-cols-3 mb-8">
        <div className="card">
          <p className="text-sm text-gray-500">Total Definitions</p>
          <p className="text-3xl font-bold text-gray-900 mt-1">{definitions.length}</p>
        </div>
        <div className="card">
          <p className="text-sm text-gray-500">Active</p>
          <p className="text-3xl font-bold text-green-600 mt-1">
            {definitions.filter(d => d.status === 'active').length}
          </p>
        </div>
        <div className="card">
          <p className="text-sm text-gray-500">System</p>
          <p className="text-3xl font-bold text-indigo-600 mt-1">
            {definitions.filter(d => d.is_system).length}
          </p>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Definitions List */}
        <div className="card overflow-hidden">
          <div className="px-4 py-3 border-b border-gray-200 bg-gray-50">
            <h2 className="font-semibold text-gray-900">Definitions</h2>
          </div>
          <div className="divide-y divide-gray-100 max-h-[600px] overflow-y-auto">
            {definitions.map(def => (
              <div
                key={def.id}
                onClick={() => loadSteps(def)}
                className={`px-4 py-3 cursor-pointer hover:bg-gray-50 transition-colors ${
                  selectedDef?.id === def.id ? 'bg-indigo-50 border-l-4 border-indigo-500' : ''
                }`}
              >
                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="font-medium text-gray-900">{def.name}</h3>
                    <p className="text-xs text-gray-500 mt-0.5">
                      {TYPE_LABELS[def.workflow_type] || def.workflow_type} | {ENTITY_LABELS[def.entity_type] || def.entity_type} | v{def.version}
                    </p>
                  </div>
                  <div className="flex items-center gap-2">
                    {def.is_system && (
                      <span className="text-xs px-2 py-0.5 bg-gray-100 text-gray-600 rounded-full">System</span>
                    )}
                    <span className={`badge ${STATUS_COLORS[def.status] || 'badge-info'}`}>
                      {def.status}
                    </span>
                  </div>
                </div>
                {def.description && (
                  <p className="text-xs text-gray-500 mt-1 line-clamp-2">{def.description}</p>
                )}
              </div>
            ))}
            {definitions.length === 0 && (
              <p className="px-4 py-8 text-center text-gray-500">No workflow definitions found</p>
            )}
          </div>
        </div>

        {/* Steps Panel */}
        <div className="card overflow-hidden">
          <div className="px-4 py-3 border-b border-gray-200 bg-gray-50 flex items-center justify-between">
            <h2 className="font-semibold text-gray-900">
              {selectedDef ? `Steps: ${selectedDef.name}` : 'Select a definition'}
            </h2>
            {selectedDef && selectedDef.status === 'draft' && !selectedDef.is_system && (
              <button
                onClick={() => handleActivate(selectedDef.id)}
                className="btn-sm bg-green-600 text-white hover:bg-green-700"
              >
                Activate
              </button>
            )}
          </div>

          {!selectedDef && (
            <p className="px-4 py-12 text-center text-gray-500">
              Click a workflow definition to view its steps
            </p>
          )}

          {selectedDef && stepsLoading && (
            <div className="px-4 py-8">
              <div className="animate-pulse space-y-3">
                {Array.from({ length: 3 }).map((_, i) => <div key={i} className="h-16 bg-gray-100 rounded" />)}
              </div>
            </div>
          )}

          {selectedDef && !stepsLoading && (
            <div className="divide-y divide-gray-100 max-h-[600px] overflow-y-auto">
              {steps.map((step, idx) => (
                <div key={step.id} className="px-4 py-3">
                  <div className="flex items-start gap-3">
                    {/* Step number circle */}
                    <div className="flex-shrink-0 w-8 h-8 rounded-full bg-indigo-100 text-indigo-700 flex items-center justify-center text-sm font-bold">
                      {step.step_order}
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2">
                        <h4 className="font-medium text-gray-900">{step.name}</h4>
                        <span className={`text-xs px-2 py-0.5 rounded-full ${STEP_TYPE_COLORS[step.step_type] || 'bg-gray-100 text-gray-800'}`}>
                          {step.step_type.replace(/_/g, ' ')}
                        </span>
                        {step.is_optional && (
                          <span className="text-xs px-2 py-0.5 bg-yellow-100 text-yellow-800 rounded-full">Optional</span>
                        )}
                      </div>
                      {step.description && (
                        <p className="text-xs text-gray-500 mt-1">{step.description}</p>
                      )}
                      <div className="flex items-center gap-4 mt-1 text-xs text-gray-500">
                        {step.sla_hours && <span>SLA: {step.sla_hours}h</span>}
                        {step.approver_type && <span>Approver: {step.approver_type.replace(/_/g, ' ')}</span>}
                        {step.approval_mode && step.step_type !== 'condition' && (
                          <span>Mode: {step.approval_mode.replace(/_/g, ' ')}</span>
                        )}
                        {step.can_delegate && <span>Delegatable</span>}
                      </div>
                    </div>
                  </div>
                  {/* Connector line */}
                  {idx < steps.length - 1 && (
                    <div className="ml-4 mt-2 mb-0 border-l-2 border-indigo-200 h-3" />
                  )}
                </div>
              ))}
              {steps.length === 0 && (
                <p className="px-4 py-8 text-center text-gray-500">No steps defined yet</p>
              )}
            </div>
          )}
        </div>
      </div>

      {/* Create Modal */}
      {showCreate && (
        <div className="fixed inset-0 bg-black/30 z-50 flex items-center justify-center p-4">
          <div className="bg-white rounded-lg shadow-xl max-w-lg w-full p-6">
            <h2 className="text-lg font-bold text-gray-900 mb-4">New Workflow Definition</h2>

            <label className="block text-sm font-medium text-gray-700 mb-1">Name</label>
            <input
              type="text"
              value={createForm.name}
              onChange={e => setCreateForm(f => ({ ...f, name: e.target.value }))}
              className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm mb-3 focus:ring-indigo-500 focus:border-indigo-500"
              placeholder="e.g., Custom Policy Review"
            />

            <label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
            <textarea
              value={createForm.description}
              onChange={e => setCreateForm(f => ({ ...f, description: e.target.value }))}
              rows={3}
              className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm mb-3 focus:ring-indigo-500 focus:border-indigo-500"
              placeholder="Describe the workflow purpose..."
            />

            <div className="grid grid-cols-2 gap-3 mb-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Workflow Type</label>
                <select
                  value={createForm.workflow_type}
                  onChange={e => setCreateForm(f => ({ ...f, workflow_type: e.target.value }))}
                  className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:ring-indigo-500 focus:border-indigo-500"
                >
                  {Object.entries(TYPE_LABELS).map(([val, label]) => (
                    <option key={val} value={val}>{label}</option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Entity Type</label>
                <select
                  value={createForm.entity_type}
                  onChange={e => setCreateForm(f => ({ ...f, entity_type: e.target.value }))}
                  className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:ring-indigo-500 focus:border-indigo-500"
                >
                  {Object.entries(ENTITY_LABELS).map(([val, label]) => (
                    <option key={val} value={val}>{label}</option>
                  ))}
                </select>
              </div>
            </div>

            <div className="flex justify-end gap-3">
              <button onClick={() => setShowCreate(false)} className="btn-secondary">Cancel</button>
              <button onClick={handleCreate} className="btn-primary">Create</button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
