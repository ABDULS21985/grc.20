-- 000034_training_certification.up.sql
-- Training & Certification Tracking — programmes, assignments, content,
-- professional certifications, phishing simulations, and results.

-- ============================================================
-- ENUMS
-- ============================================================

CREATE TYPE training_programme_status AS ENUM (
    'draft',
    'active',
    'archived'
);

CREATE TYPE training_category AS ENUM (
    'security_awareness',
    'privacy',
    'compliance',
    'technical',
    'management',
    'onboarding',
    'custom'
);

CREATE TYPE training_target_audience AS ENUM (
    'all_employees',
    'management',
    'technical_staff',
    'new_joiners',
    'specific_roles',
    'board_members'
);

CREATE TYPE assignment_status AS ENUM (
    'assigned',
    'in_progress',
    'completed',
    'failed',
    'overdue',
    'exempted'
);

CREATE TYPE content_type AS ENUM (
    'video',
    'document',
    'quiz',
    'interactive',
    'external_link',
    'scorm'
);

CREATE TYPE phishing_simulation_status AS ENUM (
    'draft',
    'scheduled',
    'active',
    'completed',
    'cancelled'
);

CREATE TYPE phishing_difficulty AS ENUM (
    'easy',
    'medium',
    'hard',
    'expert'
);

CREATE TYPE phishing_action AS ENUM (
    'delivered',
    'opened',
    'clicked',
    'submitted_data',
    'reported'
);

CREATE TYPE certification_status AS ENUM (
    'active',
    'expired',
    'expiring_soon',
    'revoked',
    'pending_renewal'
);

-- ============================================================
-- TABLE: training_programmes
-- Defines a training programme / course template.
-- ============================================================

CREATE TABLE training_programmes (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id         UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- Identity
    programme_ref           VARCHAR(30) NOT NULL,
    name                    VARCHAR(300) NOT NULL,
    description             TEXT,
    category                training_category NOT NULL DEFAULT 'security_awareness',
    target_audience         training_target_audience NOT NULL DEFAULT 'all_employees',
    target_roles            TEXT[] DEFAULT '{}',

    -- Content & Assessment
    passing_score           INTEGER NOT NULL DEFAULT 80,
    max_attempts            INTEGER NOT NULL DEFAULT 3,
    duration_minutes        INTEGER NOT NULL DEFAULT 30,
    content_version         VARCHAR(30) DEFAULT '1.0',

    -- Scheduling
    is_mandatory            BOOLEAN NOT NULL DEFAULT true,
    recurrence_months       INTEGER DEFAULT 12,
    due_within_days         INTEGER DEFAULT 30,

    -- Compliance mapping
    applicable_frameworks   TEXT[] DEFAULT '{}',
    applicable_controls     TEXT[] DEFAULT '{}',

    -- Status
    status                  training_programme_status NOT NULL DEFAULT 'draft',
    is_template             BOOLEAN NOT NULL DEFAULT false,

    -- Metadata
    metadata                JSONB DEFAULT '{}'::jsonb,
    created_by              UUID REFERENCES users(id),
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at              TIMESTAMPTZ,

    CONSTRAINT uq_training_prog_ref UNIQUE (organization_id, programme_ref)
);

-- RLS
ALTER TABLE training_programmes ENABLE ROW LEVEL SECURITY;

CREATE POLICY training_programmes_org_isolation ON training_programmes
    USING (organization_id = current_setting('app.current_org', true)::uuid);

-- Indexes
CREATE INDEX idx_training_programmes_org      ON training_programmes (organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_training_programmes_status   ON training_programmes (organization_id, status) WHERE deleted_at IS NULL;
CREATE INDEX idx_training_programmes_category ON training_programmes (organization_id, category) WHERE deleted_at IS NULL;

-- Trigger for updated_at
CREATE TRIGGER set_training_programmes_updated_at
    BEFORE UPDATE ON training_programmes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================
-- TABLE: training_content
-- Individual content items / quiz questions within a programme.
-- ============================================================

CREATE TABLE training_content (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id         UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    programme_id            UUID NOT NULL REFERENCES training_programmes(id) ON DELETE CASCADE,

    -- Content
    title                   VARCHAR(500) NOT NULL,
    content_type            content_type NOT NULL DEFAULT 'document',
    content_body            TEXT,
    content_url             VARCHAR(1000),
    sequence_order          INTEGER NOT NULL DEFAULT 0,
    duration_minutes        INTEGER DEFAULT 5,

    -- Quiz fields
    is_quiz_question        BOOLEAN NOT NULL DEFAULT false,
    question_text           TEXT,
    answer_options          JSONB DEFAULT '[]'::jsonb,
    correct_answer_index    INTEGER,
    explanation             TEXT,

    -- Metadata
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- RLS
ALTER TABLE training_content ENABLE ROW LEVEL SECURITY;

CREATE POLICY training_content_org_isolation ON training_content
    USING (organization_id = current_setting('app.current_org', true)::uuid);

-- Indexes
CREATE INDEX idx_training_content_programme ON training_content (programme_id, sequence_order);
CREATE INDEX idx_training_content_org       ON training_content (organization_id);

CREATE TRIGGER set_training_content_updated_at
    BEFORE UPDATE ON training_content
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================
-- TABLE: training_assignments
-- Tracks individual user assignments to training programmes.
-- ============================================================

CREATE TABLE training_assignments (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id         UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    programme_id            UUID NOT NULL REFERENCES training_programmes(id) ON DELETE CASCADE,
    user_id                 UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Status tracking
    status                  assignment_status NOT NULL DEFAULT 'assigned',
    assigned_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    due_date                TIMESTAMPTZ,
    started_at              TIMESTAMPTZ,
    completed_at            TIMESTAMPTZ,

    -- Assessment results
    score                   INTEGER,
    passed                  BOOLEAN,
    attempts                INTEGER NOT NULL DEFAULT 0,
    time_spent_minutes      INTEGER DEFAULT 0,

    -- Exemption
    exempted_by             UUID REFERENCES users(id),
    exemption_reason        TEXT,

    -- Certificate
    certificate_url         VARCHAR(1000),
    certificate_issued_at   TIMESTAMPTZ,

    -- Metadata
    metadata                JSONB DEFAULT '{}'::jsonb,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_training_assignment UNIQUE (organization_id, programme_id, user_id)
);

-- RLS
ALTER TABLE training_assignments ENABLE ROW LEVEL SECURITY;

CREATE POLICY training_assignments_org_isolation ON training_assignments
    USING (organization_id = current_setting('app.current_org', true)::uuid);

-- Indexes
CREATE INDEX idx_training_assignments_org         ON training_assignments (organization_id);
CREATE INDEX idx_training_assignments_user        ON training_assignments (user_id, status);
CREATE INDEX idx_training_assignments_programme   ON training_assignments (programme_id, status);
CREATE INDEX idx_training_assignments_status      ON training_assignments (organization_id, status);
CREATE INDEX idx_training_assignments_due         ON training_assignments (organization_id, due_date) WHERE status IN ('assigned', 'in_progress');
CREATE INDEX idx_training_assignments_overdue     ON training_assignments (organization_id) WHERE status = 'overdue';

CREATE TRIGGER set_training_assignments_updated_at
    BEFORE UPDATE ON training_assignments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================
-- TABLE: professional_certifications
-- Tracks professional certifications held by team members.
-- ============================================================

CREATE TABLE professional_certifications (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id         UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id                 UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Certification details
    certification_name      VARCHAR(300) NOT NULL,
    issuing_body            VARCHAR(300) NOT NULL,
    credential_id           VARCHAR(200),
    certification_url       VARCHAR(1000),

    -- Dates
    date_obtained           DATE NOT NULL,
    expiry_date             DATE,
    renewal_date            DATE,

    -- Status
    status                  certification_status NOT NULL DEFAULT 'active',

    -- Metadata
    notes                   TEXT,
    evidence_document_url   VARCHAR(1000),
    metadata                JSONB DEFAULT '{}'::jsonb,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at              TIMESTAMPTZ
);

-- RLS
ALTER TABLE professional_certifications ENABLE ROW LEVEL SECURITY;

CREATE POLICY professional_certifications_org_isolation ON professional_certifications
    USING (organization_id = current_setting('app.current_org', true)::uuid);

-- Indexes
CREATE INDEX idx_prof_certs_org       ON professional_certifications (organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_prof_certs_user      ON professional_certifications (user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_prof_certs_expiry    ON professional_certifications (organization_id, expiry_date) WHERE deleted_at IS NULL AND expiry_date IS NOT NULL;
CREATE INDEX idx_prof_certs_status    ON professional_certifications (organization_id, status) WHERE deleted_at IS NULL;

CREATE TRIGGER set_professional_certifications_updated_at
    BEFORE UPDATE ON professional_certifications
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================
-- TABLE: phishing_simulations
-- Defines phishing simulation campaigns.
-- ============================================================

CREATE TABLE phishing_simulations (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id         UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- Campaign details
    name                    VARCHAR(300) NOT NULL,
    description             TEXT,
    difficulty              phishing_difficulty NOT NULL DEFAULT 'medium',
    email_template_subject  VARCHAR(500),
    email_template_body     TEXT,
    landing_page_url        VARCHAR(1000),

    -- Targeting
    target_user_ids         UUID[] DEFAULT '{}',
    target_department       VARCHAR(200),
    target_count            INTEGER DEFAULT 0,

    -- Scheduling
    status                  phishing_simulation_status NOT NULL DEFAULT 'draft',
    scheduled_at            TIMESTAMPTZ,
    launched_at             TIMESTAMPTZ,
    completed_at            TIMESTAMPTZ,

    -- Aggregate results (computed on completion)
    total_sent              INTEGER DEFAULT 0,
    total_opened            INTEGER DEFAULT 0,
    total_clicked           INTEGER DEFAULT 0,
    total_submitted         INTEGER DEFAULT 0,
    total_reported          INTEGER DEFAULT 0,
    open_rate               NUMERIC(5,2) DEFAULT 0,
    click_rate              NUMERIC(5,2) DEFAULT 0,
    submit_rate             NUMERIC(5,2) DEFAULT 0,
    report_rate             NUMERIC(5,2) DEFAULT 0,

    -- Metadata
    created_by              UUID REFERENCES users(id),
    metadata                JSONB DEFAULT '{}'::jsonb,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- RLS
ALTER TABLE phishing_simulations ENABLE ROW LEVEL SECURITY;

CREATE POLICY phishing_simulations_org_isolation ON phishing_simulations
    USING (organization_id = current_setting('app.current_org', true)::uuid);

-- Indexes
CREATE INDEX idx_phishing_simulations_org    ON phishing_simulations (organization_id);
CREATE INDEX idx_phishing_simulations_status ON phishing_simulations (organization_id, status);

CREATE TRIGGER set_phishing_simulations_updated_at
    BEFORE UPDATE ON phishing_simulations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================
-- TABLE: phishing_simulation_results
-- Individual user results from a phishing simulation.
-- ============================================================

CREATE TABLE phishing_simulation_results (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id         UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    simulation_id           UUID NOT NULL REFERENCES phishing_simulations(id) ON DELETE CASCADE,
    user_id                 UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Interaction tracking
    email_sent_at           TIMESTAMPTZ,
    email_opened_at         TIMESTAMPTZ,
    link_clicked_at         TIMESTAMPTZ,
    data_submitted_at       TIMESTAMPTZ,
    reported_at             TIMESTAMPTZ,

    -- Final action (worst action taken)
    final_action            phishing_action NOT NULL DEFAULT 'delivered',

    -- Metadata
    user_agent              TEXT,
    ip_address              VARCHAR(45),
    metadata                JSONB DEFAULT '{}'::jsonb,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_phishing_result UNIQUE (simulation_id, user_id)
);

-- RLS
ALTER TABLE phishing_simulation_results ENABLE ROW LEVEL SECURITY;

CREATE POLICY phishing_simulation_results_org_isolation ON phishing_simulation_results
    USING (organization_id = current_setting('app.current_org', true)::uuid);

-- Indexes
CREATE INDEX idx_phishing_results_simulation ON phishing_simulation_results (simulation_id);
CREATE INDEX idx_phishing_results_user       ON phishing_simulation_results (user_id);
CREATE INDEX idx_phishing_results_org        ON phishing_simulation_results (organization_id);
CREATE INDEX idx_phishing_results_action     ON phishing_simulation_results (simulation_id, final_action);

CREATE TRIGGER set_phishing_simulation_results_updated_at
    BEFORE UPDATE ON phishing_simulation_results
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================
-- HELPER: generate_training_ref(org_id)
-- Auto-generates TP-NNN references per organisation.
-- ============================================================

CREATE OR REPLACE FUNCTION generate_training_ref(p_org_id UUID)
RETURNS VARCHAR AS $$
DECLARE
    next_num INTEGER;
BEGIN
    SELECT COALESCE(MAX(CAST(SUBSTRING(programme_ref FROM 4) AS INTEGER)), 0) + 1
    INTO next_num
    FROM training_programmes
    WHERE organization_id = p_org_id;

    RETURN 'TP-' || LPAD(next_num::TEXT, 3, '0');
END;
$$ LANGUAGE plpgsql;
