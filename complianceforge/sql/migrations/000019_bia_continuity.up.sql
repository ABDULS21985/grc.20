-- 000019_bia_continuity.up.sql
-- Business Impact Analysis & Business Continuity
-- Creates tables for business processes, BIA scenarios, continuity plans,
-- BC exercises, and process dependency mapping.

BEGIN;

-- ============================================================
-- ENUM TYPES
-- ============================================================

CREATE TYPE bp_category AS ENUM (
    'core_revenue',
    'customer_facing',
    'regulatory_required',
    'operational_support',
    'strategic',
    'administrative'
);

CREATE TYPE bp_criticality AS ENUM (
    'mission_critical',
    'business_critical',
    'important',
    'minor',
    'non_essential'
);

CREATE TYPE bp_status AS ENUM (
    'active',
    'inactive',
    'under_review'
);

CREATE TYPE impact_level AS ENUM (
    'catastrophic',
    'severe',
    'moderate',
    'minor',
    'negligible'
);

CREATE TYPE bia_scenario_type AS ENUM (
    'cyber_attack',
    'natural_disaster',
    'infrastructure_failure',
    'supply_chain',
    'pandemic',
    'regulatory_action',
    'data_loss',
    'key_person_loss',
    'utility_failure',
    'civil_unrest'
);

CREATE TYPE bia_likelihood AS ENUM (
    'almost_certain',
    'likely',
    'possible',
    'unlikely',
    'rare'
);

CREATE TYPE bia_scenario_status AS ENUM (
    'identified',
    'analysed',
    'mitigated',
    'accepted'
);

CREATE TYPE continuity_plan_type AS ENUM (
    'business_continuity_plan',
    'disaster_recovery_plan',
    'crisis_management_plan',
    'incident_response_plan',
    'pandemic_plan',
    'it_service_continuity'
);

CREATE TYPE continuity_plan_status AS ENUM (
    'draft',
    'approved',
    'active',
    'under_review',
    'archived'
);

CREATE TYPE bc_exercise_type AS ENUM (
    'tabletop',
    'walkthrough',
    'simulation',
    'full_test',
    'parallel_test',
    'component_test'
);

CREATE TYPE bc_exercise_status AS ENUM (
    'planned',
    'in_progress',
    'completed',
    'cancelled'
);

CREATE TYPE bc_exercise_rating AS ENUM (
    'pass',
    'pass_with_concerns',
    'fail'
);

CREATE TYPE dependency_type AS ENUM (
    'system',
    'application',
    'data',
    'vendor',
    'person',
    'process',
    'facility',
    'network'
);

-- ============================================================
-- TABLE: business_processes
-- Core table tracking all business processes subject to BIA.
-- ============================================================

CREATE TABLE business_processes (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id             UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    process_ref                 VARCHAR(20) NOT NULL,
    name                        VARCHAR(300) NOT NULL,
    description                 TEXT,
    process_owner_user_id       UUID REFERENCES users(id),
    department                  VARCHAR(200),
    category                    bp_category NOT NULL DEFAULT 'operational_support',
    criticality                 bp_criticality NOT NULL DEFAULT 'important',
    status                      bp_status NOT NULL DEFAULT 'active',

    -- Financial impact
    financial_impact_per_hour_eur DECIMAL(12,2),
    financial_impact_per_day_eur  DECIMAL(12,2),

    -- Impact assessments
    regulatory_impact           TEXT,
    reputational_impact         impact_level,
    legal_impact                impact_level,
    operational_impact          impact_level,
    safety_impact               impact_level,

    -- Recovery objectives
    rto_hours                   DECIMAL(8,1),
    rpo_hours                   DECIMAL(8,1),
    mtpd_hours                  DECIMAL(8,1),
    minimum_service_level       TEXT,

    -- Dependencies (denormalized for quick access)
    dependent_asset_ids         UUID[] DEFAULT '{}',
    dependent_vendor_ids        UUID[] DEFAULT '{}',
    dependent_process_ids       UUID[] DEFAULT '{}',
    key_personnel_user_ids      UUID[] DEFAULT '{}',

    -- Classification and scheduling
    data_classification         VARCHAR(50),
    peak_periods                TEXT[] DEFAULT '{}',
    last_bia_date               DATE,
    next_bia_due                DATE,
    bia_frequency_months        INT NOT NULL DEFAULT 12,

    -- Extensibility
    notes                       TEXT,
    metadata                    JSONB NOT NULL DEFAULT '{}'::jsonb,

    -- Timestamps
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at                  TIMESTAMPTZ,

    CONSTRAINT uq_business_process_ref UNIQUE (organization_id, process_ref)
);

-- ============================================================
-- TABLE: bia_scenarios
-- Disruption scenarios that may affect business processes.
-- ============================================================

CREATE TABLE bia_scenarios (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id             UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    scenario_ref                VARCHAR(20) NOT NULL,
    name                        VARCHAR(300) NOT NULL,
    description                 TEXT,
    scenario_type               bia_scenario_type NOT NULL,
    likelihood                  bia_likelihood NOT NULL DEFAULT 'possible',
    affected_process_ids        UUID[] DEFAULT '{}',
    affected_asset_ids          UUID[] DEFAULT '{}',
    impact_timeline             JSONB NOT NULL DEFAULT '{}'::jsonb,
    estimated_financial_loss_eur DECIMAL(12,2),
    mitigation_strategy         TEXT,
    status                      bia_scenario_status NOT NULL DEFAULT 'identified',

    -- Timestamps
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at                  TIMESTAMPTZ,

    CONSTRAINT uq_bia_scenario_ref UNIQUE (organization_id, scenario_ref)
);

-- ============================================================
-- TABLE: continuity_plans
-- Business continuity and disaster recovery plans.
-- ============================================================

CREATE TABLE continuity_plans (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id             UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    plan_ref                    VARCHAR(20) NOT NULL,
    name                        VARCHAR(300) NOT NULL,
    plan_type                   continuity_plan_type NOT NULL,
    status                      continuity_plan_status NOT NULL DEFAULT 'draft',
    version                     INT NOT NULL DEFAULT 1,
    scope_description           TEXT,

    -- Coverage
    covered_scenario_ids        UUID[] DEFAULT '{}',
    covered_process_ids         UUID[] DEFAULT '{}',

    -- Activation
    activation_criteria         TEXT,
    activation_authority        TEXT,

    -- Plan content (structured)
    command_structure            JSONB NOT NULL DEFAULT '{}'::jsonb,
    communication_plan          JSONB NOT NULL DEFAULT '{}'::jsonb,
    recovery_procedures         JSONB NOT NULL DEFAULT '{}'::jsonb,
    resource_requirements       JSONB NOT NULL DEFAULT '{}'::jsonb,
    alternate_site_details      JSONB NOT NULL DEFAULT '{}'::jsonb,

    -- Ownership and approval
    owner_user_id               UUID REFERENCES users(id),
    approved_by                 UUID REFERENCES users(id),
    approved_at                 TIMESTAMPTZ,

    -- Review schedule
    next_review_date            DATE,
    review_frequency_months     INT NOT NULL DEFAULT 6,
    document_path               TEXT,

    -- Timestamps
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at                  TIMESTAMPTZ,

    CONSTRAINT uq_continuity_plan_ref UNIQUE (organization_id, plan_ref)
);

-- ============================================================
-- TABLE: bc_exercises
-- Business continuity exercises / tests linked to plans and scenarios.
-- ============================================================

CREATE TABLE bc_exercises (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id             UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    exercise_ref                VARCHAR(20) NOT NULL,
    name                        VARCHAR(300) NOT NULL,
    exercise_type               bc_exercise_type NOT NULL,
    plan_id                     UUID REFERENCES continuity_plans(id) ON DELETE SET NULL,
    scenario_id                 UUID REFERENCES bia_scenarios(id) ON DELETE SET NULL,
    status                      bc_exercise_status NOT NULL DEFAULT 'planned',

    -- Scheduling
    scheduled_date              DATE,
    actual_date                 DATE,
    participants                JSONB NOT NULL DEFAULT '[]'::jsonb,

    -- Results
    rto_achieved_hours          DECIMAL(8,1),
    rpo_achieved_hours          DECIMAL(8,1),
    objectives_met              BOOLEAN,
    lessons_learned             TEXT,
    gaps_identified             TEXT,
    improvement_actions         JSONB NOT NULL DEFAULT '[]'::jsonb,
    overall_rating              bc_exercise_rating,
    report_document_path        TEXT,

    -- Ownership
    conducted_by                UUID REFERENCES users(id),

    -- Timestamps
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at                  TIMESTAMPTZ,

    CONSTRAINT uq_bc_exercise_ref UNIQUE (organization_id, exercise_ref)
);

-- ============================================================
-- TABLE: process_dependencies_map
-- Detailed dependency mapping for each business process.
-- ============================================================

CREATE TABLE process_dependencies_map (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id             UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    process_id                  UUID NOT NULL REFERENCES business_processes(id) ON DELETE CASCADE,
    dependency_type             dependency_type NOT NULL,
    dependency_entity_type      VARCHAR(50),
    dependency_entity_id        UUID,
    dependency_name             VARCHAR(300) NOT NULL,
    is_critical                 BOOLEAN NOT NULL DEFAULT false,
    alternative_available       BOOLEAN NOT NULL DEFAULT false,
    alternative_description     TEXT,
    recovery_sequence           INT,
    notes                       TEXT,

    -- Timestamps
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ============================================================
-- INDEXES
-- ============================================================

-- business_processes indexes
CREATE INDEX idx_business_processes_org_id ON business_processes(organization_id);
CREATE INDEX idx_business_processes_status ON business_processes(organization_id, status);
CREATE INDEX idx_business_processes_criticality ON business_processes(organization_id, criticality);
CREATE INDEX idx_business_processes_category ON business_processes(organization_id, category);
CREATE INDEX idx_business_processes_owner ON business_processes(process_owner_user_id) WHERE process_owner_user_id IS NOT NULL;
CREATE INDEX idx_business_processes_next_bia ON business_processes(organization_id, next_bia_due) WHERE next_bia_due IS NOT NULL;

-- bia_scenarios indexes
CREATE INDEX idx_bia_scenarios_org_id ON bia_scenarios(organization_id);
CREATE INDEX idx_bia_scenarios_status ON bia_scenarios(organization_id, status);
CREATE INDEX idx_bia_scenarios_type ON bia_scenarios(organization_id, scenario_type);

-- continuity_plans indexes
CREATE INDEX idx_continuity_plans_org_id ON continuity_plans(organization_id);
CREATE INDEX idx_continuity_plans_status ON continuity_plans(organization_id, status);
CREATE INDEX idx_continuity_plans_type ON continuity_plans(organization_id, plan_type);
CREATE INDEX idx_continuity_plans_review ON continuity_plans(organization_id, next_review_date) WHERE next_review_date IS NOT NULL;

-- bc_exercises indexes
CREATE INDEX idx_bc_exercises_org_id ON bc_exercises(organization_id);
CREATE INDEX idx_bc_exercises_status ON bc_exercises(organization_id, status);
CREATE INDEX idx_bc_exercises_plan ON bc_exercises(plan_id) WHERE plan_id IS NOT NULL;
CREATE INDEX idx_bc_exercises_scenario ON bc_exercises(scenario_id) WHERE scenario_id IS NOT NULL;
CREATE INDEX idx_bc_exercises_scheduled ON bc_exercises(organization_id, scheduled_date);

-- process_dependencies_map indexes
CREATE INDEX idx_process_deps_org_id ON process_dependencies_map(organization_id);
CREATE INDEX idx_process_deps_process_id ON process_dependencies_map(process_id);
CREATE INDEX idx_process_deps_entity ON process_dependencies_map(dependency_entity_id) WHERE dependency_entity_id IS NOT NULL;
CREATE INDEX idx_process_deps_critical ON process_dependencies_map(process_id, is_critical) WHERE is_critical = true;

-- ============================================================
-- ROW LEVEL SECURITY
-- ============================================================

ALTER TABLE business_processes ENABLE ROW LEVEL SECURITY;
ALTER TABLE bia_scenarios ENABLE ROW LEVEL SECURITY;
ALTER TABLE continuity_plans ENABLE ROW LEVEL SECURITY;
ALTER TABLE bc_exercises ENABLE ROW LEVEL SECURITY;
ALTER TABLE process_dependencies_map ENABLE ROW LEVEL SECURITY;

CREATE POLICY rls_business_processes ON business_processes
    USING (organization_id = current_setting('app.current_org')::uuid);

CREATE POLICY rls_bia_scenarios ON bia_scenarios
    USING (organization_id = current_setting('app.current_org')::uuid);

CREATE POLICY rls_continuity_plans ON continuity_plans
    USING (organization_id = current_setting('app.current_org')::uuid);

CREATE POLICY rls_bc_exercises ON bc_exercises
    USING (organization_id = current_setting('app.current_org')::uuid);

CREATE POLICY rls_process_dependencies_map ON process_dependencies_map
    USING (organization_id = current_setting('app.current_org')::uuid);

-- ============================================================
-- TRIGGERS: auto-update updated_at timestamps
-- ============================================================

CREATE OR REPLACE FUNCTION update_business_processes_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_business_processes_updated_at
    BEFORE UPDATE ON business_processes
    FOR EACH ROW EXECUTE FUNCTION update_business_processes_updated_at();

CREATE OR REPLACE FUNCTION update_bia_scenarios_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_bia_scenarios_updated_at
    BEFORE UPDATE ON bia_scenarios
    FOR EACH ROW EXECUTE FUNCTION update_bia_scenarios_updated_at();

CREATE OR REPLACE FUNCTION update_continuity_plans_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_continuity_plans_updated_at
    BEFORE UPDATE ON continuity_plans
    FOR EACH ROW EXECUTE FUNCTION update_continuity_plans_updated_at();

CREATE OR REPLACE FUNCTION update_bc_exercises_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_bc_exercises_updated_at
    BEFORE UPDATE ON bc_exercises
    FOR EACH ROW EXECUTE FUNCTION update_bc_exercises_updated_at();

CREATE OR REPLACE FUNCTION update_process_deps_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_process_deps_updated_at
    BEFORE UPDATE ON process_dependencies_map
    FOR EACH ROW EXECUTE FUNCTION update_process_deps_updated_at();

-- ============================================================
-- FUNCTION: auto-generate process_ref as BP-NNN
-- ============================================================

CREATE OR REPLACE FUNCTION generate_bp_ref(p_org_id UUID)
RETURNS VARCHAR(20) AS $$
DECLARE
    v_seq INT;
    v_ref VARCHAR(20);
BEGIN
    SELECT COALESCE(MAX(
        CAST(SUBSTRING(process_ref FROM 'BP-(\d+)') AS INT)
    ), 0) + 1
    INTO v_seq
    FROM business_processes
    WHERE organization_id = p_org_id;

    v_ref := 'BP-' || LPAD(v_seq::TEXT, 3, '0');
    RETURN v_ref;
END;
$$ LANGUAGE plpgsql;

-- ============================================================
-- FUNCTION: auto-generate scenario_ref as SCN-NNN
-- ============================================================

CREATE OR REPLACE FUNCTION generate_scn_ref(p_org_id UUID)
RETURNS VARCHAR(20) AS $$
DECLARE
    v_seq INT;
    v_ref VARCHAR(20);
BEGIN
    SELECT COALESCE(MAX(
        CAST(SUBSTRING(scenario_ref FROM 'SCN-(\d+)') AS INT)
    ), 0) + 1
    INTO v_seq
    FROM bia_scenarios
    WHERE organization_id = p_org_id;

    v_ref := 'SCN-' || LPAD(v_seq::TEXT, 3, '0');
    RETURN v_ref;
END;
$$ LANGUAGE plpgsql;

-- ============================================================
-- FUNCTION: auto-generate plan_ref as BCP-NNN
-- ============================================================

CREATE OR REPLACE FUNCTION generate_bcp_ref(p_org_id UUID)
RETURNS VARCHAR(20) AS $$
DECLARE
    v_seq INT;
    v_ref VARCHAR(20);
BEGIN
    SELECT COALESCE(MAX(
        CAST(SUBSTRING(plan_ref FROM 'BCP-(\d+)') AS INT)
    ), 0) + 1
    INTO v_seq
    FROM continuity_plans
    WHERE organization_id = p_org_id;

    v_ref := 'BCP-' || LPAD(v_seq::TEXT, 3, '0');
    RETURN v_ref;
END;
$$ LANGUAGE plpgsql;

-- ============================================================
-- FUNCTION: auto-generate exercise_ref as BCX-NNN
-- ============================================================

CREATE OR REPLACE FUNCTION generate_bcx_ref(p_org_id UUID)
RETURNS VARCHAR(20) AS $$
DECLARE
    v_seq INT;
    v_ref VARCHAR(20);
BEGIN
    SELECT COALESCE(MAX(
        CAST(SUBSTRING(exercise_ref FROM 'BCX-(\d+)') AS INT)
    ), 0) + 1
    INTO v_seq
    FROM bc_exercises
    WHERE organization_id = p_org_id;

    v_ref := 'BCX-' || LPAD(v_seq::TEXT, 3, '0');
    RETURN v_ref;
END;
$$ LANGUAGE plpgsql;

COMMIT;
