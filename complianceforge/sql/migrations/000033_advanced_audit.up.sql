-- ============================================================
-- ComplianceForge GRC Platform
-- Migration 033: Advanced Audit Management
-- Audit programmes, universe, engagements, workpapers,
-- sampling, test procedures, and corrective actions.
-- ============================================================

-- ============================================================
-- AUDIT PROGRAMMES
-- An audit programme defines the overall audit plan for a period
-- (typically annual), grouping multiple audit engagements.
-- ============================================================

CREATE TABLE audit_programmes (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id     UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    programme_ref       VARCHAR(30) NOT NULL,
    name                VARCHAR(500) NOT NULL,
    description         TEXT,
    status              VARCHAR(30) NOT NULL DEFAULT 'draft',
    programme_type      VARCHAR(30) NOT NULL DEFAULT 'annual',
    period_start        DATE NOT NULL,
    period_end          DATE NOT NULL,
    total_budget_days   INT DEFAULT 0,
    used_budget_days    INT DEFAULT 0,
    objectives          TEXT,
    risk_appetite       VARCHAR(20) DEFAULT 'medium',
    methodology         TEXT,
    approved_by         UUID REFERENCES users(id),
    approved_at         TIMESTAMPTZ,
    owner_user_id       UUID REFERENCES users(id),
    metadata            JSONB DEFAULT '{}',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(organization_id, programme_ref)
);

CREATE INDEX idx_audit_programmes_org ON audit_programmes(organization_id);
CREATE INDEX idx_audit_programmes_status ON audit_programmes(organization_id, status);
CREATE INDEX idx_audit_programmes_period ON audit_programmes(organization_id, period_start, period_end);

ALTER TABLE audit_programmes ENABLE ROW LEVEL SECURITY;
CREATE POLICY audit_programmes_tenant ON audit_programmes
    USING (organization_id = get_current_tenant() OR get_current_tenant() IS NULL);

CREATE TRIGGER trg_audit_programmes_updated_at BEFORE UPDATE ON audit_programmes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================
-- AUDIT UNIVERSE
-- The complete inventory of auditable entities within the
-- organisation, each with a risk rating and audit cycle.
-- ============================================================

CREATE TABLE audit_universe (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id         UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    entity_ref              VARCHAR(30) NOT NULL,
    name                    VARCHAR(500) NOT NULL,
    entity_type             VARCHAR(50) NOT NULL,
    description             TEXT,
    risk_rating             VARCHAR(20) NOT NULL DEFAULT 'medium',
    risk_score              DECIMAL(5,2) DEFAULT 0,
    business_owner_id       UUID REFERENCES users(id),
    department              VARCHAR(200),
    location                VARCHAR(200),
    regulatory_relevance    TEXT[] DEFAULT '{}',
    last_audit_date         DATE,
    next_audit_due          DATE,
    audit_frequency_months  INT DEFAULT 12,
    status                  VARCHAR(20) NOT NULL DEFAULT 'active',
    linked_framework_ids    UUID[] DEFAULT '{}',
    linked_risk_ids         UUID[] DEFAULT '{}',
    metadata                JSONB DEFAULT '{}',
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(organization_id, entity_ref)
);

CREATE INDEX idx_audit_universe_org ON audit_universe(organization_id);
CREATE INDEX idx_audit_universe_risk ON audit_universe(organization_id, risk_rating);
CREATE INDEX idx_audit_universe_next ON audit_universe(next_audit_due) WHERE status = 'active';
CREATE INDEX idx_audit_universe_type ON audit_universe(organization_id, entity_type);

ALTER TABLE audit_universe ENABLE ROW LEVEL SECURITY;
CREATE POLICY audit_universe_tenant ON audit_universe
    USING (organization_id = get_current_tenant() OR get_current_tenant() IS NULL);

CREATE TRIGGER trg_audit_universe_updated_at BEFORE UPDATE ON audit_universe
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================
-- AUDIT ENGAGEMENTS
-- A specific audit engagement within a programme, linked to an
-- auditable entity and to the existing audits table.
-- ============================================================

CREATE TABLE audit_engagements (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id         UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    programme_id            UUID NOT NULL REFERENCES audit_programmes(id) ON DELETE CASCADE,
    audit_id                UUID REFERENCES audits(id),
    auditable_entity_id     UUID REFERENCES audit_universe(id),
    engagement_ref          VARCHAR(30) NOT NULL,
    name                    VARCHAR(500) NOT NULL,
    engagement_type         VARCHAR(30) NOT NULL DEFAULT 'assurance',
    status                  VARCHAR(30) NOT NULL DEFAULT 'planning',
    priority                VARCHAR(20) DEFAULT 'medium',
    risk_rating             VARCHAR(20) DEFAULT 'medium',
    scope                   TEXT,
    objectives              TEXT,
    methodology             TEXT,
    lead_auditor_id         UUID REFERENCES users(id),
    audit_team_ids          UUID[] DEFAULT '{}',
    planned_start_date      DATE,
    planned_end_date        DATE,
    actual_start_date       DATE,
    actual_end_date         DATE,
    budget_days             INT DEFAULT 0,
    actual_days             INT DEFAULT 0,
    fieldwork_complete      BOOLEAN DEFAULT false,
    report_issued           BOOLEAN DEFAULT false,
    report_issued_date      DATE,
    overall_opinion         VARCHAR(30),
    metadata                JSONB DEFAULT '{}',
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(organization_id, engagement_ref)
);

CREATE INDEX idx_audit_engagements_org ON audit_engagements(organization_id);
CREATE INDEX idx_audit_engagements_prog ON audit_engagements(programme_id);
CREATE INDEX idx_audit_engagements_status ON audit_engagements(organization_id, status);
CREATE INDEX idx_audit_engagements_entity ON audit_engagements(auditable_entity_id);
CREATE INDEX idx_audit_engagements_audit ON audit_engagements(audit_id) WHERE audit_id IS NOT NULL;

ALTER TABLE audit_engagements ENABLE ROW LEVEL SECURITY;
CREATE POLICY audit_engagements_tenant ON audit_engagements
    USING (organization_id = get_current_tenant() OR get_current_tenant() IS NULL);

CREATE TRIGGER trg_audit_engagements_updated_at BEFORE UPDATE ON audit_engagements
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================
-- AUDIT WORKPAPERS
-- Documentation created during fieldwork supporting audit
-- conclusions, with four-eyes review workflow.
-- ============================================================

CREATE TABLE audit_workpapers (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id     UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    engagement_id       UUID NOT NULL REFERENCES audit_engagements(id) ON DELETE CASCADE,
    workpaper_ref       VARCHAR(30) NOT NULL,
    title               VARCHAR(500) NOT NULL,
    description         TEXT,
    workpaper_type      VARCHAR(30) NOT NULL DEFAULT 'general',
    status              VARCHAR(30) NOT NULL DEFAULT 'draft',
    content             TEXT,
    prepared_by         UUID NOT NULL REFERENCES users(id),
    prepared_date       DATE NOT NULL DEFAULT CURRENT_DATE,
    reviewed_by         UUID REFERENCES users(id),
    reviewed_date       DATE,
    review_comments     TEXT,
    linked_control_ids  UUID[] DEFAULT '{}',
    linked_risk_ids     UUID[] DEFAULT '{}',
    attachments         JSONB DEFAULT '[]',
    metadata            JSONB DEFAULT '{}',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(organization_id, workpaper_ref)
);

CREATE INDEX idx_audit_workpapers_org ON audit_workpapers(organization_id);
CREATE INDEX idx_audit_workpapers_engagement ON audit_workpapers(engagement_id);
CREATE INDEX idx_audit_workpapers_status ON audit_workpapers(organization_id, status);

ALTER TABLE audit_workpapers ENABLE ROW LEVEL SECURITY;
CREATE POLICY audit_workpapers_tenant ON audit_workpapers
    USING (organization_id = get_current_tenant() OR get_current_tenant() IS NULL);

CREATE TRIGGER trg_audit_workpapers_updated_at BEFORE UPDATE ON audit_workpapers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================
-- AUDIT SAMPLES
-- ISA 530 statistical sampling for audit testing, including
-- sample size calculation parameters and item results.
-- ============================================================

CREATE TABLE audit_samples (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id         UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    engagement_id           UUID NOT NULL REFERENCES audit_engagements(id) ON DELETE CASCADE,
    sample_ref              VARCHAR(30) NOT NULL,
    name                    VARCHAR(500) NOT NULL,
    description             TEXT,
    sampling_method         VARCHAR(30) NOT NULL DEFAULT 'statistical',
    population_size         INT NOT NULL,
    sample_size             INT NOT NULL,
    confidence_level        DECIMAL(5,2) NOT NULL DEFAULT 95.00,
    tolerable_error_rate    DECIMAL(5,4) NOT NULL DEFAULT 0.0500,
    expected_error_rate     DECIMAL(5,4) NOT NULL DEFAULT 0.0100,
    z_score                 DECIMAL(6,4) NOT NULL DEFAULT 1.9600,
    selection_method        VARCHAR(30) NOT NULL DEFAULT 'random',
    selected_items          JSONB NOT NULL DEFAULT '[]',
    items_tested            INT DEFAULT 0,
    items_passed            INT DEFAULT 0,
    items_failed            INT DEFAULT 0,
    items_inconclusive      INT DEFAULT 0,
    actual_error_rate       DECIMAL(5,4),
    conclusion              TEXT,
    status                  VARCHAR(30) NOT NULL DEFAULT 'pending',
    created_by              UUID NOT NULL REFERENCES users(id),
    metadata                JSONB DEFAULT '{}',
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(organization_id, sample_ref)
);

CREATE INDEX idx_audit_samples_org ON audit_samples(organization_id);
CREATE INDEX idx_audit_samples_engagement ON audit_samples(engagement_id);
CREATE INDEX idx_audit_samples_status ON audit_samples(organization_id, status);

ALTER TABLE audit_samples ENABLE ROW LEVEL SECURITY;
CREATE POLICY audit_samples_tenant ON audit_samples
    USING (organization_id = get_current_tenant() OR get_current_tenant() IS NULL);

CREATE TRIGGER trg_audit_samples_updated_at BEFORE UPDATE ON audit_samples
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================
-- AUDIT TEST PROCEDURES
-- Specific test procedures within an engagement, linked to
-- controls and producing pass/fail results.
-- ============================================================

CREATE TABLE audit_test_procedures (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id         UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    engagement_id           UUID NOT NULL REFERENCES audit_engagements(id) ON DELETE CASCADE,
    procedure_ref           VARCHAR(30) NOT NULL,
    title                   VARCHAR(500) NOT NULL,
    description             TEXT,
    test_type               VARCHAR(30) NOT NULL DEFAULT 'substantive',
    control_id              UUID,
    control_ref             VARCHAR(50),
    expected_result         TEXT,
    actual_result           TEXT,
    result                  VARCHAR(20),
    tested_by               UUID REFERENCES users(id),
    tested_date             DATE,
    workpaper_id            UUID REFERENCES audit_workpapers(id),
    sample_id               UUID REFERENCES audit_samples(id),
    finding_id              UUID REFERENCES audit_findings(id),
    evidence_refs           TEXT[] DEFAULT '{}',
    notes                   TEXT,
    metadata                JSONB DEFAULT '{}',
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(organization_id, procedure_ref)
);

CREATE INDEX idx_audit_test_procs_org ON audit_test_procedures(organization_id);
CREATE INDEX idx_audit_test_procs_engagement ON audit_test_procedures(engagement_id);
CREATE INDEX idx_audit_test_procs_result ON audit_test_procedures(organization_id, result) WHERE result IS NOT NULL;

ALTER TABLE audit_test_procedures ENABLE ROW LEVEL SECURITY;
CREATE POLICY audit_test_procs_tenant ON audit_test_procedures
    USING (organization_id = get_current_tenant() OR get_current_tenant() IS NULL);

CREATE TRIGGER trg_audit_test_procs_updated_at BEFORE UPDATE ON audit_test_procedures
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================
-- AUDIT CORRECTIVE ACTIONS
-- Actions arising from audit findings, tracked through to
-- closure with independent verification (four-eyes).
-- ============================================================

CREATE TABLE audit_corrective_actions (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id         UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    finding_id              UUID REFERENCES audit_findings(id),
    engagement_id           UUID REFERENCES audit_engagements(id),
    action_ref              VARCHAR(30) NOT NULL,
    title                   VARCHAR(500) NOT NULL,
    description             TEXT,
    action_type             VARCHAR(30) NOT NULL DEFAULT 'corrective',
    priority                VARCHAR(20) NOT NULL DEFAULT 'medium',
    status                  VARCHAR(30) NOT NULL DEFAULT 'open',
    root_cause              TEXT,
    planned_action          TEXT NOT NULL,
    actual_action           TEXT,
    responsible_user_id     UUID REFERENCES users(id),
    implementer_user_id     UUID REFERENCES users(id),
    due_date                DATE,
    completed_date          DATE,
    verified_by             UUID REFERENCES users(id),
    verified_date           DATE,
    verification_notes      TEXT,
    verification_status     VARCHAR(30),
    evidence_refs           TEXT[] DEFAULT '{}',
    cost_estimate           DECIMAL(15,2),
    actual_cost             DECIMAL(15,2),
    effectiveness_rating    VARCHAR(20),
    follow_up_required      BOOLEAN DEFAULT false,
    follow_up_date          DATE,
    metadata                JSONB DEFAULT '{}',
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(organization_id, action_ref)
);

CREATE INDEX idx_audit_corr_actions_org ON audit_corrective_actions(organization_id);
CREATE INDEX idx_audit_corr_actions_finding ON audit_corrective_actions(finding_id) WHERE finding_id IS NOT NULL;
CREATE INDEX idx_audit_corr_actions_engagement ON audit_corrective_actions(engagement_id) WHERE engagement_id IS NOT NULL;
CREATE INDEX idx_audit_corr_actions_status ON audit_corrective_actions(organization_id, status);
CREATE INDEX idx_audit_corr_actions_priority ON audit_corrective_actions(organization_id, priority);
CREATE INDEX idx_audit_corr_actions_due ON audit_corrective_actions(due_date) WHERE status NOT IN ('closed', 'verified');
CREATE INDEX idx_audit_corr_actions_responsible ON audit_corrective_actions(responsible_user_id) WHERE status NOT IN ('closed', 'verified');

ALTER TABLE audit_corrective_actions ENABLE ROW LEVEL SECURITY;
CREATE POLICY audit_corr_actions_tenant ON audit_corrective_actions
    USING (organization_id = get_current_tenant() OR get_current_tenant() IS NULL);

CREATE TRIGGER trg_audit_corr_actions_updated_at BEFORE UPDATE ON audit_corrective_actions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
