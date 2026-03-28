-- ============================================================
-- ComplianceForge GRC Platform
-- Seed 019: Default Workflow Definitions
-- 5 system workflow definitions with steps
-- ============================================================

-- ============================================================
-- 1. POLICY APPROVAL WORKFLOW
-- 3 steps: Compliance Officer Review -> Approver Approval -> Auto-action Update Status
-- ============================================================

INSERT INTO workflow_definitions (id, organization_id, name, description, workflow_type, entity_type, version, status, trigger_conditions, sla_config, is_system)
VALUES (
    'a0000000-0000-0000-0000-000000000001',
    NULL,
    'Policy Approval Workflow',
    'Standard 3-step workflow for policy approval. A compliance officer reviews the policy, then an approver signs off, and the policy status is automatically updated to published.',
    'policy_approval',
    'policy',
    1,
    'active',
    '{"on_status_change": "pending_approval"}',
    '{"total_sla_hours": 120, "step_default_sla_hours": 48}',
    true
);

-- Step 1: Compliance Officer Review
INSERT INTO workflow_steps (id, workflow_definition_id, organization_id, step_order, name, description, step_type, approver_type, approval_mode, minimum_approvals, sla_hours, can_delegate)
VALUES (
    'b0000000-0000-0000-0000-000000000001',
    'a0000000-0000-0000-0000-000000000001',
    NULL, 1,
    'Compliance Officer Review',
    'Compliance officer reviews the policy for accuracy, regulatory alignment, and completeness.',
    'review',
    'role',
    'any_one', 1, 48, true
);

-- Step 2: Approver Approval
INSERT INTO workflow_steps (id, workflow_definition_id, organization_id, step_order, name, description, step_type, approver_type, approval_mode, minimum_approvals, sla_hours, can_delegate)
VALUES (
    'b0000000-0000-0000-0000-000000000002',
    'a0000000-0000-0000-0000-000000000001',
    NULL, 2,
    'Management Approval',
    'Designated approver or org admin reviews and formally approves the policy.',
    'approval',
    'org_admin',
    'any_one', 1, 48, true
);

-- Step 3: Auto-action - Update policy status to published
INSERT INTO workflow_steps (id, workflow_definition_id, organization_id, step_order, name, description, step_type, auto_action, sla_hours, can_delegate)
VALUES (
    'b0000000-0000-0000-0000-000000000003',
    'a0000000-0000-0000-0000-000000000001',
    NULL, 3,
    'Publish Policy',
    'Automatically update the policy status to published upon approval.',
    'auto_action',
    '{"action": "update_entity_status", "target_status": "published", "entity_type": "policy"}',
    1, false
);

-- ============================================================
-- 2. RISK ACCEPTANCE WORKFLOW
-- Conditional: high/critical -> CISO approval, else -> Risk Manager
-- ============================================================

INSERT INTO workflow_definitions (id, organization_id, name, description, workflow_type, entity_type, version, status, trigger_conditions, sla_config, is_system)
VALUES (
    'a0000000-0000-0000-0000-000000000002',
    NULL,
    'Risk Acceptance Workflow',
    'Conditional workflow for risk acceptance. High/critical risks require CISO approval; medium/low risks are routed to the Risk Manager.',
    'risk_acceptance',
    'risk',
    1,
    'active',
    '{"on_treatment": "accept"}',
    '{"total_sla_hours": 240, "step_default_sla_hours": 72}',
    true
);

-- Step 1: Condition - Evaluate risk level
INSERT INTO workflow_steps (id, workflow_definition_id, organization_id, step_order, name, description, step_type, condition_expression, condition_true_step_id, condition_false_step_id)
VALUES (
    'b0000000-0000-0000-0000-000000000010',
    'a0000000-0000-0000-0000-000000000002',
    NULL, 1,
    'Evaluate Risk Level',
    'Route high/critical risks to CISO, others to Risk Manager.',
    'condition',
    '{"field": "risk_level", "operator": "in", "value": ["critical", "high"]}',
    'b0000000-0000-0000-0000-000000000011',
    'b0000000-0000-0000-0000-000000000012'
);

-- Step 2a: CISO Approval (for high/critical)
INSERT INTO workflow_steps (id, workflow_definition_id, organization_id, step_order, name, description, step_type, approver_type, approval_mode, minimum_approvals, sla_hours, can_delegate)
VALUES (
    'b0000000-0000-0000-0000-000000000011',
    'a0000000-0000-0000-0000-000000000002',
    NULL, 2,
    'CISO Approval',
    'Chief Information Security Officer must approve acceptance of high/critical risks.',
    'approval',
    'role',
    'any_one', 1, 72, true
);

-- Step 2b: Risk Manager Approval (for medium/low)
INSERT INTO workflow_steps (id, workflow_definition_id, organization_id, step_order, name, description, step_type, approver_type, approval_mode, minimum_approvals, sla_hours, can_delegate)
VALUES (
    'b0000000-0000-0000-0000-000000000012',
    'a0000000-0000-0000-0000-000000000002',
    NULL, 3,
    'Risk Manager Approval',
    'Risk Manager reviews and approves acceptance of medium/low risks.',
    'approval',
    'role',
    'any_one', 1, 48, true
);

-- Step 3: Auto-action - Update risk status
INSERT INTO workflow_steps (id, workflow_definition_id, organization_id, step_order, name, description, step_type, auto_action, sla_hours, can_delegate)
VALUES (
    'b0000000-0000-0000-0000-000000000013',
    'a0000000-0000-0000-0000-000000000002',
    NULL, 4,
    'Accept Risk',
    'Automatically update risk status to accepted.',
    'auto_action',
    '{"action": "update_entity_status", "target_status": "accepted", "entity_type": "risk"}',
    1, false
);

-- ============================================================
-- 3. EXCEPTION REQUEST WORKFLOW
-- 4 steps: submit justification -> condition (impact level) -> approval -> notification
-- ============================================================

INSERT INTO workflow_definitions (id, organization_id, name, description, workflow_type, entity_type, version, status, trigger_conditions, sla_config, is_system)
VALUES (
    'a0000000-0000-0000-0000-000000000003',
    NULL,
    'Exception Request Workflow',
    'Four-step workflow for compliance exception requests. Includes justification review, conditional routing based on impact, formal approval, and stakeholder notification.',
    'exception_request',
    'exception',
    1,
    'active',
    '{"on_create": true}',
    '{"total_sla_hours": 168, "step_default_sla_hours": 48}',
    true
);

-- Step 1: Justification Review
INSERT INTO workflow_steps (id, workflow_definition_id, organization_id, step_order, name, description, step_type, approver_type, approval_mode, minimum_approvals, sla_hours, can_delegate)
VALUES (
    'b0000000-0000-0000-0000-000000000020',
    'a0000000-0000-0000-0000-000000000003',
    NULL, 1,
    'Justification Review',
    'Compliance team reviews the exception justification, compensating controls, and risk assessment.',
    'review',
    'role',
    'any_one', 1, 48, true
);

-- Step 2: Condition - Impact Level
INSERT INTO workflow_steps (id, workflow_definition_id, organization_id, step_order, name, description, step_type, condition_expression, condition_true_step_id, condition_false_step_id)
VALUES (
    'b0000000-0000-0000-0000-000000000021',
    'a0000000-0000-0000-0000-000000000003',
    NULL, 2,
    'Evaluate Impact Level',
    'Route high-impact exceptions to senior management; low-impact to compliance lead.',
    'condition',
    '{"field": "impact_level", "operator": "in", "value": ["high", "critical"]}',
    'b0000000-0000-0000-0000-000000000022',
    'b0000000-0000-0000-0000-000000000023'
);

-- Step 3a: Senior Management Approval (high/critical impact)
INSERT INTO workflow_steps (id, workflow_definition_id, organization_id, step_order, name, description, step_type, approver_type, approval_mode, minimum_approvals, sla_hours, can_delegate)
VALUES (
    'b0000000-0000-0000-0000-000000000022',
    'a0000000-0000-0000-0000-000000000003',
    NULL, 3,
    'Senior Management Approval',
    'Senior management approval required for high-impact exceptions.',
    'approval',
    'org_admin',
    'all_required', 1, 72, true
);

-- Step 3b: Compliance Lead Approval (low/medium impact)
INSERT INTO workflow_steps (id, workflow_definition_id, organization_id, step_order, name, description, step_type, approver_type, approval_mode, minimum_approvals, sla_hours, can_delegate)
VALUES (
    'b0000000-0000-0000-0000-000000000023',
    'a0000000-0000-0000-0000-000000000003',
    NULL, 4,
    'Compliance Lead Approval',
    'Compliance lead approval for low/medium impact exceptions.',
    'approval',
    'role',
    'any_one', 1, 48, true
);

-- Step 4: Stakeholder Notification
INSERT INTO workflow_steps (id, workflow_definition_id, organization_id, step_order, name, description, step_type, auto_action)
VALUES (
    'b0000000-0000-0000-0000-000000000024',
    'a0000000-0000-0000-0000-000000000003',
    NULL, 5,
    'Notify Stakeholders',
    'Send notification to all relevant stakeholders about the exception decision.',
    'notification',
    '{"action": "send_notification", "template": "exception_decision", "recipients": "stakeholders"}'
);

-- ============================================================
-- 4. AUDIT FINDING REMEDIATION WORKFLOW
-- 3 steps: assign remediation task -> verify remediation -> close finding
-- ============================================================

INSERT INTO workflow_definitions (id, organization_id, name, description, workflow_type, entity_type, version, status, trigger_conditions, sla_config, is_system)
VALUES (
    'a0000000-0000-0000-0000-000000000004',
    NULL,
    'Audit Finding Remediation Workflow',
    'Three-step workflow for tracking audit finding remediation. Assigns a task to the responsible owner, then verification by the auditor, and finally auto-closes the finding.',
    'audit_finding_remediation',
    'audit_finding',
    1,
    'active',
    '{"on_finding_created": true}',
    '{"total_sla_hours": 720, "step_default_sla_hours": 240}',
    true
);

-- Step 1: Remediation Task
INSERT INTO workflow_steps (id, workflow_definition_id, organization_id, step_order, name, description, step_type, task_assignee_type, task_description, sla_hours, can_delegate)
VALUES (
    'b0000000-0000-0000-0000-000000000030',
    'a0000000-0000-0000-0000-000000000004',
    NULL, 1,
    'Remediation Task',
    'Finding owner implements the remediation plan and provides evidence of completion.',
    'task',
    'entity_owner',
    'Implement the remediation plan for this audit finding. Upload evidence of remediation and mark as complete.',
    480, true
);

-- Step 2: Auditor Verification
INSERT INTO workflow_steps (id, workflow_definition_id, organization_id, step_order, name, description, step_type, approver_type, approval_mode, minimum_approvals, sla_hours, can_delegate)
VALUES (
    'b0000000-0000-0000-0000-000000000031',
    'a0000000-0000-0000-0000-000000000004',
    NULL, 2,
    'Auditor Verification',
    'Auditor verifies the remediation evidence and confirms the finding has been adequately addressed.',
    'review',
    'role',
    'any_one', 1, 168, true
);

-- Step 3: Auto-close Finding
INSERT INTO workflow_steps (id, workflow_definition_id, organization_id, step_order, name, description, step_type, auto_action, sla_hours, can_delegate)
VALUES (
    'b0000000-0000-0000-0000-000000000032',
    'a0000000-0000-0000-0000-000000000004',
    NULL, 3,
    'Close Finding',
    'Automatically mark the audit finding as remediated and closed.',
    'auto_action',
    '{"action": "update_entity_status", "target_status": "closed", "entity_type": "audit_finding"}',
    1, false
);

-- ============================================================
-- 5. VENDOR ONBOARDING WORKFLOW
-- Parallel gate: Security Review + Legal Review -> Final Approval -> Auto-action
-- ============================================================

INSERT INTO workflow_definitions (id, organization_id, name, description, workflow_type, entity_type, version, status, trigger_conditions, sla_config, is_system)
VALUES (
    'a0000000-0000-0000-0000-000000000005',
    NULL,
    'Vendor Onboarding Workflow',
    'Parallel-gate workflow for vendor onboarding. Security and Legal reviews happen simultaneously, followed by a final management approval and automatic status update.',
    'vendor_onboarding',
    'vendor',
    1,
    'active',
    '{"on_create": true}',
    '{"total_sla_hours": 336, "step_default_sla_hours": 96}',
    true
);

-- Step 1: Parallel Gate - Security + Legal Review
INSERT INTO workflow_steps (id, workflow_definition_id, organization_id, step_order, name, description, step_type, approver_type, approver_ids, approval_mode, minimum_approvals, sla_hours, can_delegate)
VALUES (
    'b0000000-0000-0000-0000-000000000040',
    'a0000000-0000-0000-0000-000000000005',
    NULL, 1,
    'Security & Legal Review',
    'Parallel review gate: both Security and Legal teams must complete their assessments. Security evaluates risk posture; Legal reviews contractual and data processing terms.',
    'parallel_gate',
    'role',
    '{}',
    'all_required', 2, 168, true
);

-- Step 2: Final Management Approval
INSERT INTO workflow_steps (id, workflow_definition_id, organization_id, step_order, name, description, step_type, approver_type, approval_mode, minimum_approvals, sla_hours, can_delegate)
VALUES (
    'b0000000-0000-0000-0000-000000000041',
    'a0000000-0000-0000-0000-000000000005',
    NULL, 2,
    'Final Approval',
    'Management reviews the combined Security and Legal assessments and grants final vendor approval.',
    'approval',
    'org_admin',
    'any_one', 1, 96, true
);

-- Step 3: Auto-action - Activate vendor
INSERT INTO workflow_steps (id, workflow_definition_id, organization_id, step_order, name, description, step_type, auto_action, sla_hours, can_delegate)
VALUES (
    'b0000000-0000-0000-0000-000000000042',
    'a0000000-0000-0000-0000-000000000005',
    NULL, 3,
    'Activate Vendor',
    'Automatically update vendor status to approved/active.',
    'auto_action',
    '{"action": "update_entity_status", "target_status": "approved", "entity_type": "vendor"}',
    1, false
);
