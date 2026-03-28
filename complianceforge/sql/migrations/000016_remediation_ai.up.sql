-- 000016_remediation_ai.up.sql
-- AI-Assisted Compliance Remediation Planner
-- Creates tables for remediation plans, actions, and AI interaction logging.

BEGIN;

-- ============================================================
-- ENUM TYPES
-- ============================================================

CREATE TYPE remediation_plan_type AS ENUM (
    'gap_remediation',
    'audit_finding',
    'risk_treatment',
    'continuous_improvement',
    'incident_response',
    'regulatory_change'
);

CREATE TYPE remediation_plan_status AS ENUM (
    'draft',
    'pending_approval',
    'approved',
    'in_progress',
    'on_hold',
    'completed',
    'cancelled'
);

CREATE TYPE remediation_priority AS ENUM (
    'critical',
    'high',
    'medium',
    'low'
);

CREATE TYPE remediation_action_type AS ENUM (
    'implement_control',
    'update_policy',
    'deploy_technical',
    'conduct_training',
    'perform_assessment',
    'gather_evidence',
    'review_process',
    'third_party_engagement',
    'documentation',
    'monitoring_setup'
);

CREATE TYPE remediation_action_status AS ENUM (
    'pending',
    'assigned',
    'in_progress',
    'blocked',
    'in_review',
    'completed',
    'deferred',
    'cancelled'
);

-- ============================================================
-- TABLE: remediation_plans
-- Master table for AI-generated and manual compliance
-- remediation plans with full audit trail.
-- ============================================================

CREATE TABLE remediation_plans (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id             UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    plan_ref                    VARCHAR(20) NOT NULL,
    name                        VARCHAR(500) NOT NULL,
    description                 TEXT,
    plan_type                   remediation_plan_type NOT NULL DEFAULT 'gap_remediation',
    status                      remediation_plan_status NOT NULL DEFAULT 'draft',
    scope_framework_ids         UUID[] DEFAULT '{}',
    scope_description           TEXT,
    priority                    remediation_priority NOT NULL DEFAULT 'medium',

    -- AI generation metadata
    ai_generated                BOOLEAN NOT NULL DEFAULT false,
    ai_model                    VARCHAR(100),
    ai_prompt_summary           TEXT,
    ai_generation_date          TIMESTAMPTZ,
    ai_confidence_score         DECIMAL(3,2) CHECK (ai_confidence_score >= 0 AND ai_confidence_score <= 1),
    human_reviewed              BOOLEAN NOT NULL DEFAULT false,
    human_reviewed_by           UUID REFERENCES users(id),
    human_reviewed_at           TIMESTAMPTZ,

    -- Planning and tracking
    target_completion_date      DATE,
    estimated_total_hours       DECIMAL(10,1) DEFAULT 0,
    estimated_total_cost_eur    DECIMAL(12,2) DEFAULT 0,
    actual_completion_date      DATE,
    completion_percentage       DECIMAL(5,2) NOT NULL DEFAULT 0 CHECK (completion_percentage >= 0 AND completion_percentage <= 100),

    -- Ownership
    owner_user_id               UUID REFERENCES users(id),
    created_by                  UUID NOT NULL REFERENCES users(id),
    approved_by                 UUID REFERENCES users(id),
    approved_at                 TIMESTAMPTZ,

    -- Extensibility
    metadata                    JSONB NOT NULL DEFAULT '{}'::jsonb,

    -- Timestamps
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at                  TIMESTAMPTZ,

    CONSTRAINT uq_remediation_plan_ref UNIQUE (organization_id, plan_ref)
);

-- ============================================================
-- TABLE: remediation_actions
-- Individual actions within a remediation plan, with AI-generated
-- guidance, evidence suggestions, and cross-framework benefits.
-- ============================================================

CREATE TABLE remediation_actions (
    id                                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id                     UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    plan_id                             UUID NOT NULL REFERENCES remediation_plans(id) ON DELETE CASCADE,
    action_ref                          VARCHAR(30) NOT NULL,
    sort_order                          INT NOT NULL DEFAULT 0,
    title                               VARCHAR(500) NOT NULL,
    description                         TEXT,
    action_type                         remediation_action_type NOT NULL DEFAULT 'implement_control',

    -- Linkages to other GRC entities
    linked_control_implementation_id    UUID,
    linked_finding_id                   UUID,
    linked_risk_treatment_id            UUID,
    framework_control_code              VARCHAR(100),

    -- Effort estimation
    priority                            remediation_priority NOT NULL DEFAULT 'medium',
    estimated_hours                     DECIMAL(8,1) DEFAULT 0,
    estimated_cost_eur                  DECIMAL(10,2) DEFAULT 0,
    required_skills                     TEXT[] DEFAULT '{}',
    dependencies                        UUID[] DEFAULT '{}',

    -- Assignment and scheduling
    assigned_to                         UUID REFERENCES users(id),
    target_start_date                   DATE,
    target_end_date                     DATE,
    status                              remediation_action_status NOT NULL DEFAULT 'pending',

    -- Actuals
    actual_start_date                   DATE,
    actual_end_date                     DATE,
    actual_hours                        DECIMAL(8,1),
    actual_cost_eur                     DECIMAL(10,2),
    completion_notes                    TEXT,
    evidence_paths                      TEXT[] DEFAULT '{}',

    -- AI-generated content
    ai_implementation_guidance          TEXT,
    ai_evidence_suggestions             TEXT[] DEFAULT '{}',
    ai_tool_recommendations             TEXT[] DEFAULT '{}',
    ai_risk_if_deferred                 TEXT,
    ai_cross_framework_benefit          TEXT,

    -- Timestamps
    created_at                          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at                          TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at                          TIMESTAMPTZ,

    CONSTRAINT uq_remediation_action_ref UNIQUE (organization_id, action_ref)
);

-- ============================================================
-- TABLE: ai_interaction_logs
-- Audit log for all AI API interactions with cost tracking,
-- latency metrics, and user feedback for model governance.
-- ============================================================

CREATE TABLE ai_interaction_logs (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id     UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    interaction_type    VARCHAR(100) NOT NULL,
    prompt_text         TEXT NOT NULL,
    response_text       TEXT,
    model               VARCHAR(100) NOT NULL,
    input_tokens        INT DEFAULT 0,
    output_tokens       INT DEFAULT 0,
    latency_ms          INT DEFAULT 0,
    cost_eur            DECIMAL(8,4) DEFAULT 0,
    user_id             UUID REFERENCES users(id),
    rating              INT CHECK (rating >= 1 AND rating <= 5),
    feedback            TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ============================================================
-- INDEXES
-- ============================================================

-- remediation_plans indexes
CREATE INDEX idx_remediation_plans_org_id ON remediation_plans(organization_id);
CREATE INDEX idx_remediation_plans_status ON remediation_plans(organization_id, status);
CREATE INDEX idx_remediation_plans_priority ON remediation_plans(organization_id, priority);
CREATE INDEX idx_remediation_plans_owner ON remediation_plans(owner_user_id) WHERE owner_user_id IS NOT NULL;
CREATE INDEX idx_remediation_plans_created_at ON remediation_plans(organization_id, created_at DESC);
CREATE INDEX idx_remediation_plans_plan_type ON remediation_plans(organization_id, plan_type);
CREATE INDEX idx_remediation_plans_ai_generated ON remediation_plans(organization_id, ai_generated) WHERE ai_generated = true;

-- remediation_actions indexes
CREATE INDEX idx_remediation_actions_org_id ON remediation_actions(organization_id);
CREATE INDEX idx_remediation_actions_plan_id ON remediation_actions(plan_id);
CREATE INDEX idx_remediation_actions_status ON remediation_actions(organization_id, status);
CREATE INDEX idx_remediation_actions_assigned_to ON remediation_actions(assigned_to) WHERE assigned_to IS NOT NULL;
CREATE INDEX idx_remediation_actions_sort ON remediation_actions(plan_id, sort_order);
CREATE INDEX idx_remediation_actions_priority ON remediation_actions(organization_id, priority);
CREATE INDEX idx_remediation_actions_control ON remediation_actions(linked_control_implementation_id) WHERE linked_control_implementation_id IS NOT NULL;
CREATE INDEX idx_remediation_actions_finding ON remediation_actions(linked_finding_id) WHERE linked_finding_id IS NOT NULL;

-- ai_interaction_logs indexes
CREATE INDEX idx_ai_interaction_logs_org_id ON ai_interaction_logs(organization_id);
CREATE INDEX idx_ai_interaction_logs_type ON ai_interaction_logs(organization_id, interaction_type);
CREATE INDEX idx_ai_interaction_logs_user ON ai_interaction_logs(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX idx_ai_interaction_logs_created ON ai_interaction_logs(organization_id, created_at DESC);
CREATE INDEX idx_ai_interaction_logs_model ON ai_interaction_logs(model);

-- ============================================================
-- ROW LEVEL SECURITY
-- ============================================================

ALTER TABLE remediation_plans ENABLE ROW LEVEL SECURITY;
ALTER TABLE remediation_actions ENABLE ROW LEVEL SECURITY;
ALTER TABLE ai_interaction_logs ENABLE ROW LEVEL SECURITY;

CREATE POLICY rls_remediation_plans ON remediation_plans
    USING (organization_id = current_setting('app.current_org')::uuid);

CREATE POLICY rls_remediation_actions ON remediation_actions
    USING (organization_id = current_setting('app.current_org')::uuid);

CREATE POLICY rls_ai_interaction_logs ON ai_interaction_logs
    USING (organization_id = current_setting('app.current_org')::uuid);

-- ============================================================
-- TRIGGER: auto-update updated_at timestamps
-- ============================================================

CREATE OR REPLACE FUNCTION update_remediation_plans_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_remediation_plans_updated_at
    BEFORE UPDATE ON remediation_plans
    FOR EACH ROW EXECUTE FUNCTION update_remediation_plans_updated_at();

CREATE OR REPLACE FUNCTION update_remediation_actions_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_remediation_actions_updated_at
    BEFORE UPDATE ON remediation_actions
    FOR EACH ROW EXECUTE FUNCTION update_remediation_actions_updated_at();

-- ============================================================
-- FUNCTION: auto-generate plan_ref as RMP-YYYY-NNNN
-- ============================================================

CREATE OR REPLACE FUNCTION generate_remediation_plan_ref(p_org_id UUID)
RETURNS VARCHAR(20) AS $$
DECLARE
    v_year TEXT;
    v_seq  INT;
    v_ref  VARCHAR(20);
BEGIN
    v_year := to_char(now(), 'YYYY');

    SELECT COALESCE(MAX(
        CAST(SUBSTRING(plan_ref FROM 'RMP-' || v_year || '-(\d+)') AS INT)
    ), 0) + 1
    INTO v_seq
    FROM remediation_plans
    WHERE organization_id = p_org_id
      AND plan_ref LIKE 'RMP-' || v_year || '-%';

    v_ref := 'RMP-' || v_year || '-' || LPAD(v_seq::TEXT, 4, '0');
    RETURN v_ref;
END;
$$ LANGUAGE plpgsql;

-- ============================================================
-- FUNCTION: auto-generate action_ref as RMP-NNNN-ANN
-- ============================================================

CREATE OR REPLACE FUNCTION generate_remediation_action_ref(p_plan_id UUID)
RETURNS VARCHAR(30) AS $$
DECLARE
    v_plan_ref VARCHAR(20);
    v_seq      INT;
    v_ref      VARCHAR(30);
BEGIN
    SELECT plan_ref INTO v_plan_ref
    FROM remediation_plans
    WHERE id = p_plan_id;

    SELECT COALESCE(MAX(
        CAST(SUBSTRING(action_ref FROM v_plan_ref || '-A(\d+)') AS INT)
    ), 0) + 1
    INTO v_seq
    FROM remediation_actions
    WHERE plan_id = p_plan_id;

    v_ref := v_plan_ref || '-A' || LPAD(v_seq::TEXT, 2, '0');
    RETURN v_ref;
END;
$$ LANGUAGE plpgsql;

COMMIT;
