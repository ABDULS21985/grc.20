-- ============================================================
-- ComplianceForge GRC Platform
-- Migration 012: Compliance Workflow Engine
-- ============================================================

-- ============================================================
-- ENUM TYPES
-- ============================================================

CREATE TYPE workflow_type AS ENUM (
    'policy_approval', 'risk_acceptance', 'exception_request',
    'audit_finding_remediation', 'vendor_onboarding', 'control_change',
    'incident_response', 'access_review', 'custom'
);

CREATE TYPE workflow_status AS ENUM (
    'draft', 'active', 'deprecated'
);

CREATE TYPE workflow_step_type AS ENUM (
    'approval', 'review', 'task', 'notification',
    'condition', 'parallel_gate', 'timer', 'auto_action'
);

CREATE TYPE workflow_approver_type AS ENUM (
    'specific_users', 'role', 'manager', 'org_admin', 'dynamic'
);

CREATE TYPE workflow_approval_mode AS ENUM (
    'any_one', 'all_required', 'majority'
);

CREATE TYPE workflow_instance_status AS ENUM (
    'active', 'completed', 'cancelled', 'suspended', 'failed'
);

CREATE TYPE workflow_completion_outcome AS ENUM (
    'approved', 'rejected', 'completed', 'cancelled', 'expired'
);

CREATE TYPE workflow_sla_status AS ENUM (
    'on_track', 'at_risk', 'breached'
);

CREATE TYPE workflow_step_exec_status AS ENUM (
    'pending', 'in_progress', 'approved', 'rejected',
    'completed', 'skipped', 'escalated', 'delegated', 'timed_out'
);

CREATE TYPE workflow_step_action AS ENUM (
    'approve', 'reject', 'complete', 'skip',
    'delegate', 'escalate', 'cancel', 'auto_complete'
);

CREATE TYPE workflow_task_assignee_type AS ENUM (
    'specific_users', 'role', 'entity_owner', 'dynamic'
);

-- ============================================================
-- WORKFLOW DEFINITIONS
-- ============================================================

CREATE TABLE workflow_definitions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id     UUID REFERENCES organizations(id) ON DELETE CASCADE,
    name                VARCHAR(200) NOT NULL,
    description         TEXT,
    workflow_type       workflow_type NOT NULL,
    entity_type         VARCHAR(100) NOT NULL,
    version             INT NOT NULL DEFAULT 1,
    status              workflow_status NOT NULL DEFAULT 'draft',
    trigger_conditions  JSONB DEFAULT '{}',
    sla_config          JSONB DEFAULT '{}',
    metadata            JSONB DEFAULT '{}',
    created_by          UUID REFERENCES users(id) ON DELETE SET NULL,
    is_system           BOOLEAN NOT NULL DEFAULT false,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- WORKFLOW STEPS
-- ============================================================

CREATE TABLE workflow_steps (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_definition_id      UUID NOT NULL REFERENCES workflow_definitions(id) ON DELETE CASCADE,
    organization_id             UUID REFERENCES organizations(id) ON DELETE CASCADE,
    step_order                  INT NOT NULL,
    name                        VARCHAR(200) NOT NULL,
    description                 TEXT,
    step_type                   workflow_step_type NOT NULL,

    -- Approval fields
    approver_type               workflow_approver_type,
    approver_ids                UUID[],
    approval_mode               workflow_approval_mode DEFAULT 'any_one',
    minimum_approvals           INT DEFAULT 1,

    -- Task fields
    task_description            TEXT,
    task_assignee_type          workflow_task_assignee_type,
    task_assignee_ids           UUID[],

    -- Condition fields
    condition_expression        JSONB DEFAULT '{}',
    condition_true_step_id      UUID REFERENCES workflow_steps(id) ON DELETE SET NULL,
    condition_false_step_id     UUID REFERENCES workflow_steps(id) ON DELETE SET NULL,

    -- Auto-action fields
    auto_action                 JSONB DEFAULT '{}',

    -- Timer fields
    timer_hours                 INT,
    timer_business_hours_only   BOOLEAN DEFAULT false,

    -- SLA & escalation
    sla_hours                   INT,
    escalation_user_ids         UUID[],

    -- Flags
    is_optional                 BOOLEAN NOT NULL DEFAULT false,
    can_delegate                BOOLEAN NOT NULL DEFAULT true,

    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (workflow_definition_id, step_order)
);

-- ============================================================
-- WORKFLOW INSTANCES
-- ============================================================

CREATE TABLE workflow_instances (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id         UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    workflow_definition_id  UUID NOT NULL REFERENCES workflow_definitions(id) ON DELETE RESTRICT,
    entity_type             VARCHAR(100) NOT NULL,
    entity_id               UUID NOT NULL,
    entity_ref              VARCHAR(200),
    status                  workflow_instance_status NOT NULL DEFAULT 'active',
    current_step_id         UUID REFERENCES workflow_steps(id) ON DELETE SET NULL,
    current_step_order      INT,
    started_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_by              UUID REFERENCES users(id) ON DELETE SET NULL,
    completed_at            TIMESTAMPTZ,
    completion_outcome      workflow_completion_outcome,
    total_duration_hours    NUMERIC(10,2),
    sla_status              workflow_sla_status DEFAULT 'on_track',
    sla_deadline            TIMESTAMPTZ,
    metadata                JSONB DEFAULT '{}',
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- WORKFLOW STEP EXECUTIONS
-- ============================================================

CREATE TABLE workflow_step_executions (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id         UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    workflow_instance_id    UUID NOT NULL REFERENCES workflow_instances(id) ON DELETE CASCADE,
    workflow_step_id        UUID NOT NULL REFERENCES workflow_steps(id) ON DELETE CASCADE,
    step_order              INT NOT NULL,
    status                  workflow_step_exec_status NOT NULL DEFAULT 'pending',
    assigned_to             UUID REFERENCES users(id) ON DELETE SET NULL,
    delegated_to            UUID REFERENCES users(id) ON DELETE SET NULL,
    delegated_by            UUID REFERENCES users(id) ON DELETE SET NULL,
    delegated_at            TIMESTAMPTZ,
    action_taken_by         UUID REFERENCES users(id) ON DELETE SET NULL,
    action_taken_at         TIMESTAMPTZ,
    action                  workflow_step_action,
    comments                TEXT,
    decision_reason         TEXT,
    attachments_paths       TEXT[],
    sla_deadline            TIMESTAMPTZ,
    sla_status              workflow_sla_status DEFAULT 'on_track',
    escalated_at            TIMESTAMPTZ,
    escalated_to            UUID REFERENCES users(id) ON DELETE SET NULL,
    started_at              TIMESTAMPTZ,
    completed_at            TIMESTAMPTZ,
    duration_hours          NUMERIC(10,2),
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- WORKFLOW DELEGATION RULES
-- ============================================================

CREATE TABLE workflow_delegation_rules (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id     UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    delegator_user_id   UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    delegate_user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    workflow_types      TEXT[] DEFAULT '{}',
    valid_from          TIMESTAMPTZ NOT NULL,
    valid_until         TIMESTAMPTZ,
    reason              TEXT,
    is_active           BOOLEAN NOT NULL DEFAULT true,
    created_by          UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- ROW LEVEL SECURITY
-- ============================================================

ALTER TABLE workflow_definitions ENABLE ROW LEVEL SECURITY;
ALTER TABLE workflow_steps ENABLE ROW LEVEL SECURITY;
ALTER TABLE workflow_instances ENABLE ROW LEVEL SECURITY;
ALTER TABLE workflow_step_executions ENABLE ROW LEVEL SECURITY;
ALTER TABLE workflow_delegation_rules ENABLE ROW LEVEL SECURITY;

-- RLS policies: system rows (organization_id IS NULL) are visible to all; org rows are tenant-scoped
CREATE POLICY workflow_definitions_org_isolation ON workflow_definitions
    USING (organization_id IS NULL OR organization_id = current_setting('app.current_org_id', true)::uuid);

CREATE POLICY workflow_steps_org_isolation ON workflow_steps
    USING (organization_id IS NULL OR organization_id = current_setting('app.current_org_id', true)::uuid);

CREATE POLICY workflow_instances_org_isolation ON workflow_instances
    USING (organization_id = current_setting('app.current_org_id', true)::uuid);

CREATE POLICY workflow_step_executions_org_isolation ON workflow_step_executions
    USING (organization_id = current_setting('app.current_org_id', true)::uuid);

CREATE POLICY workflow_delegation_rules_org_isolation ON workflow_delegation_rules
    USING (organization_id = current_setting('app.current_org_id', true)::uuid);

-- ============================================================
-- INDEXES
-- ============================================================

-- Workflow definitions
CREATE INDEX idx_workflow_definitions_org ON workflow_definitions(organization_id);
CREATE INDEX idx_workflow_definitions_type ON workflow_definitions(workflow_type);
CREATE INDEX idx_workflow_definitions_status ON workflow_definitions(status);
CREATE INDEX idx_workflow_definitions_entity_type ON workflow_definitions(entity_type);

-- Workflow steps
CREATE INDEX idx_workflow_steps_definition ON workflow_steps(workflow_definition_id);
CREATE INDEX idx_workflow_steps_order ON workflow_steps(workflow_definition_id, step_order);
CREATE INDEX idx_workflow_steps_type ON workflow_steps(step_type);

-- Workflow instances
CREATE INDEX idx_workflow_instances_org ON workflow_instances(organization_id);
CREATE INDEX idx_workflow_instances_definition ON workflow_instances(workflow_definition_id);
CREATE INDEX idx_workflow_instances_entity ON workflow_instances(entity_type, entity_id);
CREATE INDEX idx_workflow_instances_status ON workflow_instances(status);
CREATE INDEX idx_workflow_instances_sla ON workflow_instances(sla_status) WHERE status = 'active';
CREATE INDEX idx_workflow_instances_current_step ON workflow_instances(current_step_id);

-- Workflow step executions
CREATE INDEX idx_workflow_step_exec_instance ON workflow_step_executions(workflow_instance_id);
CREATE INDEX idx_workflow_step_exec_step ON workflow_step_executions(workflow_step_id);
CREATE INDEX idx_workflow_step_exec_assigned ON workflow_step_executions(assigned_to) WHERE status IN ('pending', 'in_progress');
CREATE INDEX idx_workflow_step_exec_status ON workflow_step_executions(status);
CREATE INDEX idx_workflow_step_exec_sla ON workflow_step_executions(sla_deadline) WHERE status IN ('pending', 'in_progress');

-- Delegation rules
CREATE INDEX idx_workflow_delegation_delegator ON workflow_delegation_rules(delegator_user_id);
CREATE INDEX idx_workflow_delegation_delegate ON workflow_delegation_rules(delegate_user_id);
CREATE INDEX idx_workflow_delegation_active ON workflow_delegation_rules(organization_id, is_active) WHERE is_active = true;

-- ============================================================
-- TRIGGERS FOR updated_at
-- ============================================================

CREATE TRIGGER set_workflow_definitions_updated_at
    BEFORE UPDATE ON workflow_definitions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER set_workflow_steps_updated_at
    BEFORE UPDATE ON workflow_steps
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER set_workflow_instances_updated_at
    BEFORE UPDATE ON workflow_instances
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER set_workflow_delegation_rules_updated_at
    BEFORE UPDATE ON workflow_delegation_rules
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
