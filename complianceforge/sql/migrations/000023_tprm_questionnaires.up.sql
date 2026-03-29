-- ============================================================
-- ComplianceForge GRC Platform
-- Migration 023: Third-Party Risk Management (TPRM)
--                Assessment Questionnaires
-- ============================================================
-- Provides assessment questionnaire templates, vendor assessment
-- lifecycle management, response collection, scoring, and a
-- vendor self-service portal for completing assessments.
-- ============================================================

BEGIN;

-- ============================================================
-- ENUM TYPES
-- ============================================================

CREATE TYPE questionnaire_type AS ENUM (
    'security',
    'privacy',
    'compliance',
    'operational',
    'financial',
    'esg',
    'custom'
);

CREATE TYPE questionnaire_status AS ENUM (
    'draft',
    'active',
    'deprecated'
);

CREATE TYPE scoring_method_type AS ENUM (
    'weighted_average',
    'simple_average',
    'minimum_score',
    'custom_formula'
);

CREATE TYPE question_type AS ENUM (
    'yes_no',
    'multiple_choice',
    'single_choice',
    'likert_scale',
    'numeric',
    'text',
    'date',
    'file_upload',
    'matrix'
);

CREATE TYPE question_risk_impact AS ENUM (
    'critical',
    'high',
    'medium',
    'low',
    'informational'
);

CREATE TYPE vendor_assessment_status AS ENUM (
    'draft',
    'sent',
    'in_progress',
    'submitted',
    'under_review',
    'completed',
    'expired',
    'cancelled'
);

CREATE TYPE vendor_assessment_pass_fail AS ENUM (
    'pass',
    'fail',
    'conditional_pass',
    'pending'
);

CREATE TYPE vendor_response_flag AS ENUM (
    'none',
    'acceptable',
    'needs_clarification',
    'concern',
    'critical_finding'
);

-- ============================================================
-- ASSESSMENT QUESTIONNAIRES
-- Master template that defines an assessment questionnaire.
-- organization_id NULL = system-level (built-in) template.
-- ============================================================

CREATE TABLE assessment_questionnaires (
    id                           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id              UUID REFERENCES organizations(id) ON DELETE CASCADE,
    name                         VARCHAR(300) NOT NULL,
    description                  TEXT NOT NULL DEFAULT '',
    questionnaire_type           questionnaire_type NOT NULL DEFAULT 'security',
    version                      INT NOT NULL DEFAULT 1,
    status                       questionnaire_status NOT NULL DEFAULT 'draft',
    total_questions              INT NOT NULL DEFAULT 0,
    total_sections               INT NOT NULL DEFAULT 0,
    estimated_completion_minutes INT NOT NULL DEFAULT 0,
    scoring_method               scoring_method_type NOT NULL DEFAULT 'weighted_average',
    pass_threshold               DECIMAL(5,2) NOT NULL DEFAULT 70.00,
    risk_tier_thresholds         JSONB NOT NULL DEFAULT '{"critical": 40, "high": 55, "medium": 70, "low": 85}',
    applicable_vendor_tiers      TEXT[] NOT NULL DEFAULT '{}',
    is_system                    BOOLEAN NOT NULL DEFAULT false,
    is_template                  BOOLEAN NOT NULL DEFAULT false,
    created_by                   UUID REFERENCES users(id),
    created_at                   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE assessment_questionnaires IS 'Master questionnaire templates for vendor risk assessments';
COMMENT ON COLUMN assessment_questionnaires.organization_id IS 'NULL for system-level built-in templates';
COMMENT ON COLUMN assessment_questionnaires.is_system IS 'True for immutable system templates seeded at startup';

-- ============================================================
-- QUESTIONNAIRE SECTIONS
-- Logical groupings of questions within a questionnaire.
-- ============================================================

CREATE TABLE questionnaire_sections (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    questionnaire_id    UUID NOT NULL REFERENCES assessment_questionnaires(id) ON DELETE CASCADE,
    name                VARCHAR(200) NOT NULL,
    description         TEXT NOT NULL DEFAULT '',
    sort_order          INT NOT NULL DEFAULT 0,
    weight              DECIMAL(5,2) NOT NULL DEFAULT 1.00,
    framework_domain_code VARCHAR(50),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE questionnaire_sections IS 'Sections within a questionnaire template';

-- ============================================================
-- QUESTIONNAIRE QUESTIONS
-- Individual questions within a section.
-- ============================================================

CREATE TABLE questionnaire_questions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    section_id          UUID NOT NULL REFERENCES questionnaire_sections(id) ON DELETE CASCADE,
    question_text       TEXT NOT NULL,
    question_type       question_type NOT NULL DEFAULT 'single_choice',
    options             JSONB NOT NULL DEFAULT '[]',
    is_required         BOOLEAN NOT NULL DEFAULT true,
    weight              DECIMAL(5,2) NOT NULL DEFAULT 1.00,
    risk_impact         question_risk_impact NOT NULL DEFAULT 'medium',
    guidance_text       TEXT NOT NULL DEFAULT '',
    evidence_required   BOOLEAN NOT NULL DEFAULT false,
    evidence_guidance   TEXT NOT NULL DEFAULT '',
    mapped_control_codes TEXT[] NOT NULL DEFAULT '{}',
    conditional_on      JSONB,
    sort_order          INT NOT NULL DEFAULT 0,
    tags                TEXT[] NOT NULL DEFAULT '{}',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE questionnaire_questions IS 'Questions within a questionnaire section';
COMMENT ON COLUMN questionnaire_questions.options IS 'JSON array of {value, label, score} for choice-type questions';
COMMENT ON COLUMN questionnaire_questions.conditional_on IS 'JSON {question_id, expected_value} — show only when condition is met';

-- ============================================================
-- VENDOR ASSESSMENTS
-- Tracks a single assessment sent to / completed by a vendor.
-- ============================================================

CREATE TABLE vendor_assessments (
    id                       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id          UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    vendor_id                UUID NOT NULL REFERENCES vendors(id) ON DELETE CASCADE,
    questionnaire_id         UUID NOT NULL REFERENCES assessment_questionnaires(id) ON DELETE RESTRICT,
    assessment_ref           VARCHAR(20) NOT NULL,
    status                   vendor_assessment_status NOT NULL DEFAULT 'draft',
    sent_at                  TIMESTAMPTZ,
    sent_to_email            VARCHAR(300),
    sent_to_name             VARCHAR(200),
    access_token_hash        VARCHAR(128),
    reminder_count           INT NOT NULL DEFAULT 0,
    last_reminder_at         TIMESTAMPTZ,
    due_date                 DATE,
    submitted_at             TIMESTAMPTZ,
    overall_score            DECIMAL(5,2),
    risk_rating              VARCHAR(20),
    section_scores           JSONB NOT NULL DEFAULT '{}',
    critical_findings        INT NOT NULL DEFAULT 0,
    high_findings            INT NOT NULL DEFAULT 0,
    pass_fail                vendor_assessment_pass_fail NOT NULL DEFAULT 'pending',
    reviewed_by              UUID REFERENCES users(id),
    reviewed_at              TIMESTAMPTZ,
    review_notes             TEXT NOT NULL DEFAULT '',
    reviewer_override_score  DECIMAL(5,2),
    reviewer_override_reason TEXT NOT NULL DEFAULT '',
    follow_up_required       BOOLEAN NOT NULL DEFAULT false,
    follow_up_items          JSONB NOT NULL DEFAULT '[]',
    next_assessment_date     DATE,
    metadata                 JSONB NOT NULL DEFAULT '{}',
    created_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE vendor_assessments IS 'Tracks assessment lifecycle for a specific vendor';
COMMENT ON COLUMN vendor_assessments.access_token_hash IS 'SHA-256 of the 32-byte random token sent to the vendor';

-- ============================================================
-- VENDOR ASSESSMENT RESPONSES
-- Individual answers provided by the vendor.
-- ============================================================

CREATE TABLE vendor_assessment_responses (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id     UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    assessment_id       UUID NOT NULL REFERENCES vendor_assessments(id) ON DELETE CASCADE,
    question_id         UUID NOT NULL REFERENCES questionnaire_questions(id) ON DELETE CASCADE,
    answer_value        TEXT NOT NULL DEFAULT '',
    answer_score        DECIMAL(5,2),
    evidence_paths      TEXT[] NOT NULL DEFAULT '{}',
    evidence_notes      TEXT NOT NULL DEFAULT '',
    reviewer_comment    TEXT NOT NULL DEFAULT '',
    reviewer_flag       vendor_response_flag NOT NULL DEFAULT 'none',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE vendor_assessment_responses IS 'Individual vendor responses to assessment questions';

-- ============================================================
-- VENDOR PORTAL SESSIONS
-- Tracks vendor portal access sessions for audit purposes.
-- ============================================================

CREATE TABLE vendor_portal_sessions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    assessment_id       UUID NOT NULL REFERENCES vendor_assessments(id) ON DELETE CASCADE,
    access_token_hash   VARCHAR(128) NOT NULL,
    vendor_email        VARCHAR(300) NOT NULL DEFAULT '',
    ip_address          VARCHAR(45) NOT NULL DEFAULT '',
    user_agent          TEXT NOT NULL DEFAULT '',
    started_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_activity_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at        TIMESTAMPTZ,
    progress_percentage DECIMAL(5,2) NOT NULL DEFAULT 0.00,
    is_active           BOOLEAN NOT NULL DEFAULT true
);

COMMENT ON TABLE vendor_portal_sessions IS 'Audit trail of vendor portal access';

-- ============================================================
-- ROW LEVEL SECURITY
-- ============================================================

ALTER TABLE vendor_assessments ENABLE ROW LEVEL SECURITY;
ALTER TABLE vendor_assessment_responses ENABLE ROW LEVEL SECURITY;

-- Org-scoped RLS for vendor_assessments
CREATE POLICY vendor_assessments_org_isolation ON vendor_assessments
    USING (organization_id = current_setting('app.current_org', true)::uuid);

-- Org-scoped RLS for vendor_assessment_responses
CREATE POLICY vendor_assessment_responses_org_isolation ON vendor_assessment_responses
    USING (organization_id = current_setting('app.current_org', true)::uuid);

-- Note: assessment_questionnaires supports both system (NULL org_id) and org-scoped
-- entries. RLS uses a combined policy.
ALTER TABLE assessment_questionnaires ENABLE ROW LEVEL SECURITY;

CREATE POLICY assessment_questionnaires_org_isolation ON assessment_questionnaires
    USING (
        organization_id IS NULL
        OR organization_id = current_setting('app.current_org', true)::uuid
    );

-- ============================================================
-- AUTO-GENERATE ASSESSMENT REFERENCE
-- Generates references like TPRM-2026-000001
-- ============================================================

CREATE OR REPLACE FUNCTION generate_assessment_ref()
RETURNS TRIGGER AS $$
DECLARE
    year_str TEXT;
    seq_num  INT;
BEGIN
    year_str := EXTRACT(YEAR FROM NOW())::TEXT;
    SELECT COALESCE(MAX(
        CAST(SUBSTRING(assessment_ref FROM 10) AS INT)
    ), 0) + 1
    INTO seq_num
    FROM vendor_assessments
    WHERE assessment_ref LIKE 'TPRM-' || year_str || '-%'
      AND organization_id = NEW.organization_id;

    NEW.assessment_ref := 'TPRM-' || year_str || '-' || LPAD(seq_num::TEXT, 6, '0');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_vendor_assessment_ref
    BEFORE INSERT ON vendor_assessments
    FOR EACH ROW
    WHEN (NEW.assessment_ref IS NULL OR NEW.assessment_ref = '')
    EXECUTE FUNCTION generate_assessment_ref();

-- ============================================================
-- AUTO-UPDATE updated_at
-- ============================================================

CREATE TRIGGER trg_assessment_questionnaires_updated
    BEFORE UPDATE ON assessment_questionnaires
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_vendor_assessments_updated
    BEFORE UPDATE ON vendor_assessments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_vendor_assessment_responses_updated
    BEFORE UPDATE ON vendor_assessment_responses
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================
-- INDEXES
-- ============================================================

-- assessment_questionnaires
CREATE INDEX idx_aq_org_id ON assessment_questionnaires(organization_id);
CREATE INDEX idx_aq_status ON assessment_questionnaires(status);
CREATE INDEX idx_aq_type ON assessment_questionnaires(questionnaire_type);
CREATE INDEX idx_aq_system ON assessment_questionnaires(is_system) WHERE is_system = true;

-- questionnaire_sections
CREATE INDEX idx_qs_questionnaire ON questionnaire_sections(questionnaire_id);
CREATE INDEX idx_qs_sort ON questionnaire_sections(questionnaire_id, sort_order);

-- questionnaire_questions
CREATE INDEX idx_qq_section ON questionnaire_questions(section_id);
CREATE INDEX idx_qq_sort ON questionnaire_questions(section_id, sort_order);
CREATE INDEX idx_qq_risk_impact ON questionnaire_questions(risk_impact);

-- vendor_assessments
CREATE INDEX idx_va_org ON vendor_assessments(organization_id);
CREATE INDEX idx_va_vendor ON vendor_assessments(vendor_id);
CREATE INDEX idx_va_questionnaire ON vendor_assessments(questionnaire_id);
CREATE INDEX idx_va_status ON vendor_assessments(organization_id, status);
CREATE INDEX idx_va_ref ON vendor_assessments(assessment_ref);
CREATE INDEX idx_va_token ON vendor_assessments(access_token_hash) WHERE access_token_hash IS NOT NULL;
CREATE INDEX idx_va_due ON vendor_assessments(due_date) WHERE status IN ('sent', 'in_progress');
CREATE INDEX idx_va_pass_fail ON vendor_assessments(organization_id, pass_fail);

-- vendor_assessment_responses
CREATE INDEX idx_var_assessment ON vendor_assessment_responses(assessment_id);
CREATE INDEX idx_var_question ON vendor_assessment_responses(question_id);
CREATE UNIQUE INDEX idx_var_assessment_question ON vendor_assessment_responses(assessment_id, question_id);

-- vendor_portal_sessions
CREATE INDEX idx_vps_assessment ON vendor_portal_sessions(assessment_id);
CREATE INDEX idx_vps_token ON vendor_portal_sessions(access_token_hash);

COMMIT;
