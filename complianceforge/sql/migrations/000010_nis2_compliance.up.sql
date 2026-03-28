-- ============================================================
-- ComplianceForge GRC Platform
-- Migration 010: NIS2 Compliance Automation Module
-- Tables for entity assessment, incident reporting (3-phase),
-- security measures (Article 21), and management accountability.
-- ============================================================

-- ============================================================
-- ENUM TYPES FOR NIS2
-- ============================================================

CREATE TYPE nis2_entity_type AS ENUM ('essential', 'important', 'not_applicable');

CREATE TYPE nis2_reporting_status AS ENUM ('not_required', 'pending', 'submitted', 'overdue');

CREATE TYPE nis2_measure_status AS ENUM ('not_started', 'in_progress', 'implemented', 'verified');

-- ============================================================
-- NIS2 ENTITY ASSESSMENT
-- Determines whether an organisation is 'essential' or 'important'
-- under the NIS2 Directive based on sector, size, and turnover.
-- ============================================================

CREATE TABLE nis2_entity_assessment (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id         UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    entity_type             nis2_entity_type NOT NULL DEFAULT 'not_applicable',
    sector                  VARCHAR(200) NOT NULL,
    sub_sector              VARCHAR(200),
    assessment_criteria     JSONB DEFAULT '{}',
    employee_count          INT DEFAULT 0,
    annual_turnover_eur     DECIMAL(15,2) DEFAULT 0,
    assessment_date         DATE NOT NULL DEFAULT CURRENT_DATE,
    assessed_by             UUID REFERENCES users(id),
    is_in_scope             BOOLEAN DEFAULT false,
    member_state            VARCHAR(5),
    competent_authority     VARCHAR(200),
    csirt_name              VARCHAR(200),
    csirt_contact_email     VARCHAR(200),
    csirt_reporting_url     TEXT,
    notes                   TEXT,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_nis2_entity_org ON nis2_entity_assessment(organization_id);
CREATE INDEX idx_nis2_entity_type ON nis2_entity_assessment(organization_id, entity_type);
CREATE INDEX idx_nis2_entity_sector ON nis2_entity_assessment(sector);

ALTER TABLE nis2_entity_assessment ENABLE ROW LEVEL SECURITY;
CREATE POLICY nis2_entity_assessment_tenant ON nis2_entity_assessment
    USING (organization_id = get_current_tenant() OR get_current_tenant() IS NULL);

CREATE TRIGGER trg_nis2_entity_assessment_updated_at
    BEFORE UPDATE ON nis2_entity_assessment
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================
-- NIS2 INCIDENT REPORTS (3-Phase Reporting)
-- Phase 1: Early Warning (24 hours)
-- Phase 2: Incident Notification (72 hours)
-- Phase 3: Final Report (1 month from notification)
-- Per NIS2 Article 23
-- ============================================================

CREATE OR REPLACE FUNCTION generate_nis2_report_ref(org_id UUID)
RETURNS VARCHAR AS $$
DECLARE next_num INT; BEGIN
    SELECT COALESCE(MAX(CAST(SUBSTRING(report_ref FROM 11) AS INT)), 0) + 1
    INTO next_num FROM nis2_incident_reports WHERE organization_id = org_id;
    RETURN 'NIS2-' || EXTRACT(YEAR FROM NOW())::TEXT || '-' || LPAD(next_num::TEXT, 4, '0');
END;
$$ LANGUAGE plpgsql;

CREATE TABLE nis2_incident_reports (
    id                              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id                 UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    incident_id                     UUID NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    report_ref                      VARCHAR(30) NOT NULL,

    -- Phase 1: Early Warning (24 hours from detection)
    early_warning_status            nis2_reporting_status DEFAULT 'not_required',
    early_warning_deadline          TIMESTAMPTZ,
    early_warning_submitted_at      TIMESTAMPTZ,
    early_warning_submitted_by      UUID REFERENCES users(id),
    early_warning_content           JSONB DEFAULT '{}',
    early_warning_csirt_reference   VARCHAR(100),

    -- Phase 2: Incident Notification (72 hours from detection)
    notification_status             nis2_reporting_status DEFAULT 'not_required',
    notification_deadline           TIMESTAMPTZ,
    notification_submitted_at       TIMESTAMPTZ,
    notification_submitted_by       UUID REFERENCES users(id),
    notification_content            JSONB DEFAULT '{}',
    notification_csirt_reference    VARCHAR(100),

    -- Phase 3: Final Report (1 month from notification submission)
    final_report_status             nis2_reporting_status DEFAULT 'not_required',
    final_report_deadline           TIMESTAMPTZ,
    final_report_submitted_at       TIMESTAMPTZ,
    final_report_submitted_by       UUID REFERENCES users(id),
    final_report_content            JSONB DEFAULT '{}',
    final_report_document_path      TEXT,

    created_at                      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(organization_id, report_ref)
);

CREATE INDEX idx_nis2_incidents_org ON nis2_incident_reports(organization_id);
CREATE INDEX idx_nis2_incidents_incident ON nis2_incident_reports(incident_id);
CREATE INDEX idx_nis2_incidents_ew_status ON nis2_incident_reports(organization_id, early_warning_status)
    WHERE early_warning_status IN ('pending', 'overdue');
CREATE INDEX idx_nis2_incidents_notif_status ON nis2_incident_reports(organization_id, notification_status)
    WHERE notification_status IN ('pending', 'overdue');
CREATE INDEX idx_nis2_incidents_final_status ON nis2_incident_reports(organization_id, final_report_status)
    WHERE final_report_status IN ('pending', 'overdue');

ALTER TABLE nis2_incident_reports ENABLE ROW LEVEL SECURITY;
CREATE POLICY nis2_incident_reports_tenant ON nis2_incident_reports
    USING (organization_id = get_current_tenant() OR get_current_tenant() IS NULL);

CREATE TRIGGER trg_nis2_incident_reports_updated_at
    BEFORE UPDATE ON nis2_incident_reports
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================
-- NIS2 SECURITY MEASURES (Article 21)
-- Tracks implementation status of the 10 mandatory security
-- measures required under NIS2 Article 21(2).
-- ============================================================

CREATE TABLE nis2_security_measures (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id         UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    measure_code            VARCHAR(20) NOT NULL,
    measure_title           VARCHAR(500) NOT NULL,
    measure_description     TEXT,
    article_reference       VARCHAR(50),
    implementation_status   nis2_measure_status DEFAULT 'not_started',
    owner_user_id           UUID REFERENCES users(id),
    evidence_description    TEXT,
    last_assessed_at        TIMESTAMPTZ,
    next_assessment_date    DATE,
    linked_control_ids      UUID[] DEFAULT '{}',
    notes                   TEXT,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(organization_id, measure_code)
);

CREATE INDEX idx_nis2_measures_org ON nis2_security_measures(organization_id);
CREATE INDEX idx_nis2_measures_status ON nis2_security_measures(organization_id, implementation_status);
CREATE INDEX idx_nis2_measures_code ON nis2_security_measures(measure_code);

ALTER TABLE nis2_security_measures ENABLE ROW LEVEL SECURITY;
CREATE POLICY nis2_security_measures_tenant ON nis2_security_measures
    USING (organization_id = get_current_tenant() OR get_current_tenant() IS NULL);

CREATE TRIGGER trg_nis2_security_measures_updated_at
    BEFORE UPDATE ON nis2_security_measures
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================
-- NIS2 MANAGEMENT ACCOUNTABILITY
-- Tracks board-level training and risk measure approvals
-- per NIS2 Article 20 (Governance).
-- ============================================================

CREATE TABLE nis2_management_accountability (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id             UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    board_member_name           VARCHAR(200) NOT NULL,
    board_member_role           VARCHAR(200) NOT NULL,
    training_completed          BOOLEAN DEFAULT false,
    training_date               DATE,
    training_provider           VARCHAR(200),
    training_certificate_path   TEXT,
    risk_measures_approved      BOOLEAN DEFAULT false,
    approval_date               DATE,
    approval_document_path      TEXT,
    next_training_due           DATE,
    notes                       TEXT,
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_nis2_mgmt_org ON nis2_management_accountability(organization_id);
CREATE INDEX idx_nis2_mgmt_training ON nis2_management_accountability(organization_id, training_completed);
CREATE INDEX idx_nis2_mgmt_next_due ON nis2_management_accountability(next_training_due)
    WHERE next_training_due IS NOT NULL;

ALTER TABLE nis2_management_accountability ENABLE ROW LEVEL SECURITY;
CREATE POLICY nis2_management_accountability_tenant ON nis2_management_accountability
    USING (organization_id = get_current_tenant() OR get_current_tenant() IS NULL);

CREATE TRIGGER trg_nis2_mgmt_accountability_updated_at
    BEFORE UPDATE ON nis2_management_accountability
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================
-- NIS2 CONTROL MAPPINGS (cross-reference to ISO 27001)
-- Maps each NIS2 Article 21 measure to ISO 27001 controls.
-- ============================================================

CREATE TABLE nis2_control_mappings (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nis2_measure_code       VARCHAR(20) NOT NULL,
    iso_control_code        VARCHAR(20) NOT NULL,
    mapping_strength        DECIMAL(3,2) DEFAULT 0.80,
    notes                   TEXT,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_nis2_ctrl_map_measure ON nis2_control_mappings(nis2_measure_code);
CREATE INDEX idx_nis2_ctrl_map_iso ON nis2_control_mappings(iso_control_code);
CREATE UNIQUE INDEX idx_nis2_ctrl_map_unique ON nis2_control_mappings(nis2_measure_code, iso_control_code);
