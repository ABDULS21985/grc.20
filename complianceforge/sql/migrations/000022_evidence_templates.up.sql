-- ============================================================
-- Migration 000022: Evidence Template Library & Automated Evidence Testing
-- ============================================================

BEGIN;

-- ============================================================
-- ENUM TYPES
-- ============================================================

CREATE TYPE evidence_category_type AS ENUM (
    'document',
    'screenshot',
    'configuration_export',
    'log_extract',
    'scan_report',
    'interview_record',
    'test_result',
    'certification',
    'training_record',
    'policy_document',
    'procedure_document',
    'meeting_minutes',
    'email_confirmation',
    'system_report',
    'audit_trail'
);

CREATE TYPE evidence_collection_method AS ENUM (
    'manual_upload',
    'automated_pull',
    'api_integration',
    'screenshot_capture',
    'system_export',
    'interview',
    'observation',
    'document_review',
    'automated_scan'
);

CREATE TYPE evidence_collection_frequency AS ENUM (
    'once',
    'daily',
    'weekly',
    'monthly',
    'quarterly',
    'semi_annually',
    'annually',
    'on_change',
    'continuous'
);

CREATE TYPE evidence_difficulty AS ENUM (
    'trivial',
    'easy',
    'moderate',
    'hard',
    'expert'
);

CREATE TYPE evidence_auditor_priority AS ENUM (
    'critical',
    'high',
    'medium',
    'low',
    'informational'
);

CREATE TYPE evidence_requirement_status AS ENUM (
    'pending',
    'in_progress',
    'collected',
    'validated',
    'rejected',
    'overdue',
    'waived',
    'not_applicable'
);

CREATE TYPE evidence_validation_status AS ENUM (
    'not_validated',
    'pass',
    'fail',
    'partial',
    'expired'
);

CREATE TYPE evidence_test_type AS ENUM (
    'completeness',
    'freshness',
    'accuracy',
    'consistency',
    'format_compliance',
    'coverage',
    'timeliness'
);

CREATE TYPE evidence_test_suite_type AS ENUM (
    'pre_audit',
    'continuous',
    'on_demand',
    'regression',
    'framework_specific'
);

CREATE TYPE evidence_test_run_status AS ENUM (
    'pending',
    'running',
    'completed',
    'failed',
    'cancelled'
);

CREATE TYPE evidence_test_case_type AS ENUM (
    'exists',
    'not_empty',
    'date_within',
    'contains_text',
    'file_type_check',
    'file_size_check',
    'freshness_check',
    'approval_check',
    'coverage_check',
    'custom_query'
);

CREATE TYPE evidence_run_trigger_type AS ENUM (
    'manual',
    'scheduled',
    'pre_audit',
    'on_evidence_upload',
    'api'
);

-- ============================================================
-- TABLE: evidence_templates
-- ============================================================

CREATE TABLE evidence_templates (
    id                              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id                 UUID REFERENCES organizations(id) ON DELETE CASCADE,
    framework_control_code          VARCHAR(50) NOT NULL,
    framework_code                  VARCHAR(20) NOT NULL,
    name                            VARCHAR(300) NOT NULL,
    description                     TEXT NOT NULL DEFAULT '',
    evidence_category               evidence_category_type NOT NULL DEFAULT 'document',
    collection_method               evidence_collection_method NOT NULL DEFAULT 'manual_upload',
    collection_instructions         TEXT NOT NULL DEFAULT '',
    collection_frequency            evidence_collection_frequency NOT NULL DEFAULT 'annually',
    typical_collection_time_minutes INT NOT NULL DEFAULT 30,
    validation_rules                JSONB NOT NULL DEFAULT '[]'::jsonb,
    acceptance_criteria             TEXT NOT NULL DEFAULT '',
    common_rejection_reasons        TEXT[] NOT NULL DEFAULT '{}',
    template_fields                 JSONB NOT NULL DEFAULT '[]'::jsonb,
    sample_evidence_description     TEXT NOT NULL DEFAULT '',
    sample_file_path                TEXT NOT NULL DEFAULT '',
    applicable_to                   TEXT[] NOT NULL DEFAULT '{}',
    difficulty                      evidence_difficulty NOT NULL DEFAULT 'moderate',
    auditor_priority                evidence_auditor_priority NOT NULL DEFAULT 'medium',
    is_system                       BOOLEAN NOT NULL DEFAULT false,
    tags                            TEXT[] NOT NULL DEFAULT '{}',
    created_at                      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE evidence_templates IS 'Library of evidence templates mapping to framework controls with collection guidance and validation rules.';
COMMENT ON COLUMN evidence_templates.organization_id IS 'NULL for system-wide templates, set for org-specific custom templates.';
COMMENT ON COLUMN evidence_templates.validation_rules IS 'Array of rule objects: [{rule_type, params}] e.g. [{rule_type:"file_not_empty"},{rule_type:"date_within",params:{days:90}}]';

-- ============================================================
-- TABLE: evidence_requirements
-- ============================================================

CREATE TABLE evidence_requirements (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id             UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    control_implementation_id   UUID,
    evidence_template_id        UUID REFERENCES evidence_templates(id) ON DELETE SET NULL,
    status                      evidence_requirement_status NOT NULL DEFAULT 'pending',
    is_mandatory                BOOLEAN NOT NULL DEFAULT true,
    collection_frequency_override evidence_collection_frequency,
    assigned_to                 UUID,
    due_date                    DATE,
    last_collected_at           TIMESTAMPTZ,
    last_validated_at           TIMESTAMPTZ,
    last_evidence_id            UUID,
    validation_status           evidence_validation_status NOT NULL DEFAULT 'not_validated',
    validation_results          JSONB NOT NULL DEFAULT '{}'::jsonb,
    next_collection_due         DATE,
    consecutive_failures        INT NOT NULL DEFAULT 0,
    notes                       TEXT NOT NULL DEFAULT '',
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE evidence_requirements IS 'Organization-specific evidence requirements linked to control implementations and evidence templates.';

-- ============================================================
-- TABLE: evidence_test_suites
-- ============================================================

CREATE TABLE evidence_test_suites (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id         UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name                    VARCHAR(200) NOT NULL,
    description             TEXT NOT NULL DEFAULT '',
    test_type               evidence_test_suite_type NOT NULL DEFAULT 'on_demand',
    schedule_cron           VARCHAR(100) NOT NULL DEFAULT '',
    is_active               BOOLEAN NOT NULL DEFAULT true,
    last_run_at             TIMESTAMPTZ,
    last_run_status         evidence_test_run_status,
    pass_threshold_percent  DECIMAL(5,2) NOT NULL DEFAULT 80.00,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE evidence_test_suites IS 'Collections of evidence test cases grouped into suites for batch execution.';

-- ============================================================
-- TABLE: evidence_test_cases
-- ============================================================

CREATE TABLE evidence_test_cases (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id             UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    test_suite_id               UUID NOT NULL REFERENCES evidence_test_suites(id) ON DELETE CASCADE,
    name                        VARCHAR(300) NOT NULL,
    description                 TEXT NOT NULL DEFAULT '',
    test_type                   evidence_test_case_type NOT NULL DEFAULT 'exists',
    target_control_code         VARCHAR(50) NOT NULL DEFAULT '',
    target_evidence_template_id UUID REFERENCES evidence_templates(id) ON DELETE SET NULL,
    test_config                 JSONB NOT NULL DEFAULT '{}'::jsonb,
    expected_result             VARCHAR(100) NOT NULL DEFAULT 'pass',
    sort_order                  INT NOT NULL DEFAULT 0,
    is_critical                 BOOLEAN NOT NULL DEFAULT false,
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE evidence_test_cases IS 'Individual test case definitions within a test suite, each verifying a specific evidence condition.';

-- ============================================================
-- TABLE: evidence_test_runs
-- ============================================================

CREATE TABLE evidence_test_runs (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id     UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    test_suite_id       UUID NOT NULL REFERENCES evidence_test_suites(id) ON DELETE CASCADE,
    status              evidence_test_run_status NOT NULL DEFAULT 'pending',
    started_at          TIMESTAMPTZ,
    completed_at        TIMESTAMPTZ,
    total_tests         INT NOT NULL DEFAULT 0,
    passed              INT NOT NULL DEFAULT 0,
    failed              INT NOT NULL DEFAULT 0,
    skipped             INT NOT NULL DEFAULT 0,
    errors              INT NOT NULL DEFAULT 0,
    pass_rate           DECIMAL(5,2) NOT NULL DEFAULT 0.00,
    threshold_met       BOOLEAN NOT NULL DEFAULT false,
    results             JSONB NOT NULL DEFAULT '[]'::jsonb,
    triggered_by        evidence_run_trigger_type NOT NULL DEFAULT 'manual',
    triggered_by_user   UUID,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE evidence_test_runs IS 'Records of test suite executions with aggregated results and individual test case outcomes.';

-- ============================================================
-- ROW LEVEL SECURITY
-- ============================================================

ALTER TABLE evidence_templates ENABLE ROW LEVEL SECURITY;
ALTER TABLE evidence_requirements ENABLE ROW LEVEL SECURITY;
ALTER TABLE evidence_test_suites ENABLE ROW LEVEL SECURITY;
ALTER TABLE evidence_test_cases ENABLE ROW LEVEL SECURITY;
ALTER TABLE evidence_test_runs ENABLE ROW LEVEL SECURITY;

-- evidence_templates: system templates (org_id IS NULL) visible to all, org templates filtered by org
CREATE POLICY evidence_templates_isolation ON evidence_templates
    USING (
        organization_id IS NULL
        OR organization_id::text = current_setting('app.current_org', true)
    );

CREATE POLICY evidence_requirements_isolation ON evidence_requirements
    USING (organization_id::text = current_setting('app.current_org', true));

CREATE POLICY evidence_test_suites_isolation ON evidence_test_suites
    USING (organization_id::text = current_setting('app.current_org', true));

CREATE POLICY evidence_test_cases_isolation ON evidence_test_cases
    USING (organization_id::text = current_setting('app.current_org', true));

CREATE POLICY evidence_test_runs_isolation ON evidence_test_runs
    USING (organization_id::text = current_setting('app.current_org', true));

-- ============================================================
-- INDEXES
-- ============================================================

-- evidence_templates
CREATE INDEX idx_evidence_templates_org ON evidence_templates(organization_id) WHERE organization_id IS NOT NULL;
CREATE INDEX idx_evidence_templates_framework ON evidence_templates(framework_code);
CREATE INDEX idx_evidence_templates_control ON evidence_templates(framework_control_code);
CREATE INDEX idx_evidence_templates_framework_control ON evidence_templates(framework_code, framework_control_code);
CREATE INDEX idx_evidence_templates_system ON evidence_templates(is_system) WHERE is_system = true;
CREATE INDEX idx_evidence_templates_category ON evidence_templates(evidence_category);
CREATE INDEX idx_evidence_templates_priority ON evidence_templates(auditor_priority);
CREATE INDEX idx_evidence_templates_tags ON evidence_templates USING gin(tags);

-- evidence_requirements
CREATE INDEX idx_evidence_requirements_org ON evidence_requirements(organization_id);
CREATE INDEX idx_evidence_requirements_control ON evidence_requirements(control_implementation_id);
CREATE INDEX idx_evidence_requirements_template ON evidence_requirements(evidence_template_id);
CREATE INDEX idx_evidence_requirements_status ON evidence_requirements(organization_id, status);
CREATE INDEX idx_evidence_requirements_assigned ON evidence_requirements(assigned_to) WHERE assigned_to IS NOT NULL;
CREATE INDEX idx_evidence_requirements_due ON evidence_requirements(organization_id, due_date) WHERE due_date IS NOT NULL;
CREATE INDEX idx_evidence_requirements_validation ON evidence_requirements(organization_id, validation_status);
CREATE INDEX idx_evidence_requirements_next_due ON evidence_requirements(organization_id, next_collection_due) WHERE next_collection_due IS NOT NULL;

-- evidence_test_suites
CREATE INDEX idx_evidence_test_suites_org ON evidence_test_suites(organization_id);
CREATE INDEX idx_evidence_test_suites_active ON evidence_test_suites(organization_id, is_active) WHERE is_active = true;

-- evidence_test_cases
CREATE INDEX idx_evidence_test_cases_suite ON evidence_test_cases(test_suite_id);
CREATE INDEX idx_evidence_test_cases_org ON evidence_test_cases(organization_id);
CREATE INDEX idx_evidence_test_cases_control ON evidence_test_cases(target_control_code) WHERE target_control_code != '';

-- evidence_test_runs
CREATE INDEX idx_evidence_test_runs_suite ON evidence_test_runs(test_suite_id);
CREATE INDEX idx_evidence_test_runs_org ON evidence_test_runs(organization_id);
CREATE INDEX idx_evidence_test_runs_status ON evidence_test_runs(organization_id, status);
CREATE INDEX idx_evidence_test_runs_created ON evidence_test_runs(organization_id, created_at DESC);

-- ============================================================
-- TRIGGER: updated_at
-- ============================================================

CREATE OR REPLACE FUNCTION update_evidence_templates_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_evidence_templates_updated_at
    BEFORE UPDATE ON evidence_templates
    FOR EACH ROW EXECUTE FUNCTION update_evidence_templates_updated_at();

CREATE TRIGGER trg_evidence_requirements_updated_at
    BEFORE UPDATE ON evidence_requirements
    FOR EACH ROW EXECUTE FUNCTION update_evidence_templates_updated_at();

CREATE TRIGGER trg_evidence_test_suites_updated_at
    BEFORE UPDATE ON evidence_test_suites
    FOR EACH ROW EXECUTE FUNCTION update_evidence_templates_updated_at();

CREATE TRIGGER trg_evidence_test_cases_updated_at
    BEFORE UPDATE ON evidence_test_cases
    FOR EACH ROW EXECUTE FUNCTION update_evidence_templates_updated_at();

COMMIT;
