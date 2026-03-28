-- ============================================================
-- ComplianceForge GRC Platform
-- Migration 005: Policies, Audits, Incidents, Vendors, Assets
-- ============================================================

-- ============================================================
-- POLICY CATEGORIES
-- ============================================================

CREATE TABLE policy_categories (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id     UUID REFERENCES organizations(id) ON DELETE CASCADE,
    name                VARCHAR(255) NOT NULL,
    code                VARCHAR(50) NOT NULL,
    description         TEXT,
    parent_category_id  UUID REFERENCES policy_categories(id),
    sort_order          INT DEFAULT 0,
    is_system_default   BOOLEAN DEFAULT false,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_policy_cats_org ON policy_categories(organization_id);

-- ============================================================
-- POLICIES
-- ============================================================

CREATE OR REPLACE FUNCTION generate_policy_ref(org_id UUID)
RETURNS VARCHAR AS $$
DECLARE next_num INT; BEGIN
    SELECT COALESCE(MAX(CAST(SUBSTRING(policy_ref FROM 5) AS INT)), 0) + 1
    INTO next_num FROM policies WHERE organization_id = org_id;
    RETURN 'POL-' || LPAD(next_num::TEXT, 4, '0');
END;
$$ LANGUAGE plpgsql;

CREATE TABLE policies (
    id                              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id                 UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    policy_ref                      VARCHAR(30) NOT NULL,
    title                           VARCHAR(500) NOT NULL,
    category_id                     UUID REFERENCES policy_categories(id),
    status                          policy_status DEFAULT 'draft',
    classification                  classification_level DEFAULT 'internal',
    owner_user_id                   UUID REFERENCES users(id),
    author_user_id                  UUID REFERENCES users(id),
    approver_user_id                UUID REFERENCES users(id),
    current_version                 INT DEFAULT 1,
    current_version_id              UUID,
    review_frequency_months         INT DEFAULT 12,
    last_review_date                DATE,
    next_review_date                DATE,
    review_status                   VARCHAR(20) DEFAULT 'current',
    applies_to_all                  BOOLEAN DEFAULT true,
    applicable_departments          UUID[] DEFAULT '{}',
    applicable_roles                TEXT[] DEFAULT '{}',
    applicable_locations            TEXT[] DEFAULT '{}',
    linked_framework_ids            UUID[] DEFAULT '{}',
    linked_control_ids              UUID[] DEFAULT '{}',
    linked_risk_ids                 UUID[] DEFAULT '{}',
    parent_policy_id                UUID REFERENCES policies(id),
    supersedes_policy_id            UUID REFERENCES policies(id),
    effective_date                  DATE,
    expiry_date                     DATE,
    tags                            TEXT[] DEFAULT '{}',
    is_mandatory                    BOOLEAN DEFAULT true,
    requires_attestation            BOOLEAN DEFAULT true,
    attestation_frequency_months    INT DEFAULT 12,
    metadata                        JSONB DEFAULT '{}',
    search_vector                   tsvector,
    created_at                      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at                      TIMESTAMPTZ,
    UNIQUE(organization_id, policy_ref)
);

CREATE INDEX idx_policies_org ON policies(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_policies_status ON policies(organization_id, status) WHERE deleted_at IS NULL;
CREATE INDEX idx_policies_review ON policies(next_review_date) WHERE deleted_at IS NULL;
CREATE INDEX idx_policies_search ON policies USING GIN(search_vector);

ALTER TABLE policies ENABLE ROW LEVEL SECURITY;
CREATE POLICY policies_tenant ON policies
    USING (organization_id = get_current_tenant() OR get_current_tenant() IS NULL);

CREATE TRIGGER trg_policies_updated_at BEFORE UPDATE ON policies FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- Auto-update search vector
CREATE OR REPLACE FUNCTION update_policy_search_vector()
RETURNS TRIGGER AS $$
BEGIN
    NEW.search_vector := to_tsvector('english',
        COALESCE(NEW.policy_ref, '') || ' ' || COALESCE(NEW.title, '')
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_policies_search BEFORE INSERT OR UPDATE ON policies
    FOR EACH ROW EXECUTE FUNCTION update_policy_search_vector();

-- ============================================================
-- POLICY VERSIONS
-- ============================================================

CREATE TABLE policy_versions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    policy_id           UUID NOT NULL REFERENCES policies(id) ON DELETE CASCADE,
    organization_id     UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    version_number      INT NOT NULL,
    version_label       VARCHAR(20),
    title               VARCHAR(500) NOT NULL,
    content_html        TEXT,
    content_text        TEXT,
    summary             TEXT,
    change_description  TEXT,
    change_type         VARCHAR(20) DEFAULT 'major',
    language            VARCHAR(10) DEFAULT 'en',
    word_count          INT DEFAULT 0,
    status              VARCHAR(20) DEFAULT 'draft',
    created_by          UUID NOT NULL REFERENCES users(id),
    published_at        TIMESTAMPTZ,
    published_by        UUID REFERENCES users(id),
    file_path           TEXT,
    file_hash           VARCHAR(128),
    metadata            JSONB DEFAULT '{}',
    search_vector       tsvector,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(policy_id, version_number)
);

CREATE INDEX idx_policy_versions_policy ON policy_versions(policy_id);
CREATE INDEX idx_policy_versions_org ON policy_versions(organization_id);

ALTER TABLE policy_versions ENABLE ROW LEVEL SECURITY;
CREATE POLICY policy_versions_tenant ON policy_versions
    USING (organization_id = get_current_tenant() OR get_current_tenant() IS NULL);

-- Add FK from policies to policy_versions now that both exist
ALTER TABLE policies ADD CONSTRAINT fk_policies_current_version
    FOREIGN KEY (current_version_id) REFERENCES policy_versions(id);

-- ============================================================
-- POLICY ATTESTATION CAMPAIGNS
-- ============================================================

CREATE TABLE policy_attestation_campaigns (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id         UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name                    VARCHAR(255) NOT NULL,
    description             TEXT,
    policy_ids              UUID[] NOT NULL,
    target_audience         JSONB DEFAULT '{}',
    status                  VARCHAR(20) DEFAULT 'draft',
    start_date              DATE,
    due_date                DATE,
    total_recipients        INT DEFAULT 0,
    attested_count          INT DEFAULT 0,
    completion_rate         DECIMAL(5,2) DEFAULT 0.00,
    auto_remind             BOOLEAN DEFAULT true,
    reminder_frequency_days INT DEFAULT 7,
    escalation_after_days   INT DEFAULT 30,
    escalation_to           UUID REFERENCES users(id),
    created_by              UUID NOT NULL REFERENCES users(id),
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_attest_campaigns_org ON policy_attestation_campaigns(organization_id);

ALTER TABLE policy_attestation_campaigns ENABLE ROW LEVEL SECURITY;
CREATE POLICY attest_campaigns_tenant ON policy_attestation_campaigns
    USING (organization_id = get_current_tenant() OR get_current_tenant() IS NULL);

-- ============================================================
-- POLICY ATTESTATIONS
-- ============================================================

CREATE TABLE policy_attestations (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    policy_id           UUID NOT NULL REFERENCES policies(id) ON DELETE CASCADE,
    policy_version_id   UUID NOT NULL REFERENCES policy_versions(id),
    organization_id     UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id             UUID NOT NULL REFERENCES users(id),
    campaign_id         UUID REFERENCES policy_attestation_campaigns(id),
    status              VARCHAR(20) DEFAULT 'pending',
    attested_at         TIMESTAMPTZ,
    attested_from_ip    INET,
    attestation_method  VARCHAR(20) DEFAULT 'digital_click',
    attestation_text    TEXT,
    declined_reason     TEXT,
    due_date            DATE,
    reminder_count      INT DEFAULT 0,
    last_reminder_at    TIMESTAMPTZ,
    expires_at          TIMESTAMPTZ,
    metadata            JSONB DEFAULT '{}',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_attestations_org ON policy_attestations(organization_id);
CREATE INDEX idx_attestations_user ON policy_attestations(user_id, policy_id);
CREATE INDEX idx_attestations_campaign ON policy_attestations(campaign_id) WHERE campaign_id IS NOT NULL;
CREATE INDEX idx_attestations_status ON policy_attestations(organization_id, status);

ALTER TABLE policy_attestations ENABLE ROW LEVEL SECURITY;
CREATE POLICY attestations_tenant ON policy_attestations
    USING (organization_id = get_current_tenant() OR get_current_tenant() IS NULL);

-- ============================================================
-- POLICY EXCEPTIONS
-- ============================================================

CREATE TABLE policy_exceptions (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id         UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    policy_id               UUID NOT NULL REFERENCES policies(id) ON DELETE CASCADE,
    exception_ref           VARCHAR(20) NOT NULL,
    title                   VARCHAR(500) NOT NULL,
    description             TEXT,
    justification           TEXT,
    risk_assessment         TEXT,
    compensating_controls   TEXT,
    status                  VARCHAR(20) DEFAULT 'requested',
    requested_by            UUID NOT NULL REFERENCES users(id),
    approved_by             UUID REFERENCES users(id),
    approved_at             TIMESTAMPTZ,
    effective_date          DATE,
    expiry_date             DATE,
    review_date             DATE,
    risk_level              VARCHAR(10),
    linked_risk_id          UUID REFERENCES risks(id),
    conditions              TEXT,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_policy_exceptions_org ON policy_exceptions(organization_id);

ALTER TABLE policy_exceptions ENABLE ROW LEVEL SECURITY;
CREATE POLICY policy_exceptions_tenant ON policy_exceptions
    USING (organization_id = get_current_tenant() OR get_current_tenant() IS NULL);

-- ============================================================
-- AUDITS
-- ============================================================

CREATE OR REPLACE FUNCTION generate_audit_ref(org_id UUID)
RETURNS VARCHAR AS $$
DECLARE next_num INT; BEGIN
    SELECT COALESCE(MAX(CAST(SUBSTRING(audit_ref FROM 5) AS INT)), 0) + 1
    INTO next_num FROM audits WHERE organization_id = org_id;
    RETURN 'AUD-' || LPAD(next_num::TEXT, 4, '0');
END;
$$ LANGUAGE plpgsql;

CREATE TABLE audits (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id         UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    audit_ref               VARCHAR(20) NOT NULL,
    title                   VARCHAR(500) NOT NULL,
    audit_type              VARCHAR(30) NOT NULL, -- internal, external, regulatory, compliance
    status                  audit_status DEFAULT 'planned',
    description             TEXT,
    scope                   TEXT,
    methodology             TEXT,
    lead_auditor_id         UUID REFERENCES users(id),
    audit_team_ids          UUID[] DEFAULT '{}',
    linked_framework_ids    UUID[] DEFAULT '{}',
    planned_start_date      DATE,
    planned_end_date        DATE,
    actual_start_date       DATE,
    actual_end_date         DATE,
    total_findings          INT DEFAULT 0,
    critical_findings       INT DEFAULT 0,
    high_findings           INT DEFAULT 0,
    medium_findings         INT DEFAULT 0,
    low_findings            INT DEFAULT 0,
    report_file_path        TEXT,
    conclusion              TEXT,
    tags                    TEXT[] DEFAULT '{}',
    metadata                JSONB DEFAULT '{}',
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at              TIMESTAMPTZ,
    UNIQUE(organization_id, audit_ref)
);

CREATE INDEX idx_audits_org ON audits(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_audits_status ON audits(organization_id, status) WHERE deleted_at IS NULL;

ALTER TABLE audits ENABLE ROW LEVEL SECURITY;
CREATE POLICY audits_tenant ON audits
    USING (organization_id = get_current_tenant() OR get_current_tenant() IS NULL);

CREATE TRIGGER trg_audits_updated_at BEFORE UPDATE ON audits FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================
-- AUDIT FINDINGS
-- ============================================================

CREATE TABLE audit_findings (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id         UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    audit_id                UUID NOT NULL REFERENCES audits(id) ON DELETE CASCADE,
    finding_ref             VARCHAR(30) NOT NULL,
    title                   VARCHAR(500) NOT NULL,
    description             TEXT NOT NULL,
    severity                finding_severity NOT NULL,
    status                  VARCHAR(30) DEFAULT 'open',
    finding_type            VARCHAR(30) DEFAULT 'non_conformity',
    control_id              UUID REFERENCES control_implementations(id),
    root_cause              TEXT,
    recommendation          TEXT,
    management_response     TEXT,
    responsible_user_id     UUID REFERENCES users(id),
    due_date                DATE,
    closed_date             DATE,
    evidence_ids            UUID[] DEFAULT '{}',
    linked_risk_id          UUID REFERENCES risks(id),
    metadata                JSONB DEFAULT '{}',
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at              TIMESTAMPTZ
);

CREATE INDEX idx_findings_org ON audit_findings(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_findings_audit ON audit_findings(audit_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_findings_status ON audit_findings(organization_id, status) WHERE deleted_at IS NULL;
CREATE INDEX idx_findings_severity ON audit_findings(organization_id, severity) WHERE deleted_at IS NULL;

ALTER TABLE audit_findings ENABLE ROW LEVEL SECURITY;
CREATE POLICY findings_tenant ON audit_findings
    USING (organization_id = get_current_tenant() OR get_current_tenant() IS NULL);

CREATE TRIGGER trg_findings_updated_at BEFORE UPDATE ON audit_findings FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================
-- INCIDENTS
-- ============================================================

CREATE OR REPLACE FUNCTION generate_incident_ref(org_id UUID)
RETURNS VARCHAR AS $$
DECLARE next_num INT; BEGIN
    SELECT COALESCE(MAX(CAST(SUBSTRING(incident_ref FROM 5) AS INT)), 0) + 1
    INTO next_num FROM incidents WHERE organization_id = org_id;
    RETURN 'INC-' || LPAD(next_num::TEXT, 4, '0');
END;
$$ LANGUAGE plpgsql;

CREATE TABLE incidents (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id             UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    incident_ref                VARCHAR(20) NOT NULL,
    title                       VARCHAR(500) NOT NULL,
    description                 TEXT NOT NULL,
    incident_type               incident_type NOT NULL,
    severity                    incident_severity NOT NULL,
    status                      incident_status DEFAULT 'open',
    reported_by                 UUID NOT NULL REFERENCES users(id),
    reported_at                 TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    assigned_to                 UUID REFERENCES users(id),
    -- GDPR Data Breach (72-hour notification)
    is_data_breach              BOOLEAN DEFAULT false,
    data_subjects_affected      INT DEFAULT 0,
    data_categories_affected    TEXT[] DEFAULT '{}',
    notification_required       BOOLEAN DEFAULT false,
    dpa_notified_at             TIMESTAMPTZ,
    data_subjects_notified_at   TIMESTAMPTZ,
    notification_deadline       TIMESTAMPTZ,
    -- NIS2 Incident Reporting
    is_nis2_reportable          BOOLEAN DEFAULT false,
    nis2_early_warning_at       TIMESTAMPTZ,
    nis2_notification_at        TIMESTAMPTZ,
    nis2_final_report_at        TIMESTAMPTZ,
    -- Impact & Resolution
    impact_description          TEXT,
    financial_impact_eur        DECIMAL(15,2) DEFAULT 0,
    root_cause                  TEXT,
    containment_actions         TEXT,
    remediation_actions         TEXT,
    lessons_learned             TEXT,
    contained_at                TIMESTAMPTZ,
    resolved_at                 TIMESTAMPTZ,
    closed_at                   TIMESTAMPTZ,
    linked_control_ids          UUID[] DEFAULT '{}',
    linked_risk_ids             UUID[] DEFAULT '{}',
    tags                        TEXT[] DEFAULT '{}',
    metadata                    JSONB DEFAULT '{}',
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at                  TIMESTAMPTZ,
    UNIQUE(organization_id, incident_ref)
);

CREATE INDEX idx_incidents_org ON incidents(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_incidents_status ON incidents(organization_id, status) WHERE deleted_at IS NULL;
CREATE INDEX idx_incidents_severity ON incidents(organization_id, severity) WHERE deleted_at IS NULL;
CREATE INDEX idx_incidents_breach ON incidents(organization_id) WHERE is_data_breach = true AND deleted_at IS NULL;
CREATE INDEX idx_incidents_nis2 ON incidents(organization_id) WHERE is_nis2_reportable = true AND deleted_at IS NULL;

ALTER TABLE incidents ENABLE ROW LEVEL SECURITY;
CREATE POLICY incidents_tenant ON incidents
    USING (organization_id = get_current_tenant() OR get_current_tenant() IS NULL);

CREATE TRIGGER trg_incidents_updated_at BEFORE UPDATE ON incidents FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- Auto-set GDPR 72-hour notification deadline
CREATE OR REPLACE FUNCTION set_breach_deadline()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.is_data_breach = true AND NEW.notification_deadline IS NULL THEN
        NEW.notification_deadline := NEW.reported_at + INTERVAL '72 hours';
        NEW.notification_required := true;
    END IF;
    -- NIS2 early warning deadline is 24 hours
    IF NEW.is_nis2_reportable = true AND NEW.nis2_early_warning_at IS NULL THEN
        -- Just flag it; actual notification is manual
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_incidents_breach_deadline
    BEFORE INSERT OR UPDATE ON incidents
    FOR EACH ROW EXECUTE FUNCTION set_breach_deadline();

-- ============================================================
-- VENDORS / THIRD PARTIES
-- ============================================================

CREATE OR REPLACE FUNCTION generate_vendor_ref(org_id UUID)
RETURNS VARCHAR AS $$
DECLARE next_num INT; BEGIN
    SELECT COALESCE(MAX(CAST(SUBSTRING(vendor_ref FROM 5) AS INT)), 0) + 1
    INTO next_num FROM vendors WHERE organization_id = org_id;
    RETURN 'VND-' || LPAD(next_num::TEXT, 4, '0');
END;
$$ LANGUAGE plpgsql;

CREATE TABLE vendors (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id         UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    vendor_ref              VARCHAR(20) NOT NULL,
    name                    VARCHAR(255) NOT NULL,
    legal_name              VARCHAR(500),
    website                 TEXT,
    industry                VARCHAR(100),
    country_code            CHAR(2),
    contact_name            VARCHAR(200),
    contact_email           VARCHAR(320),
    contact_phone           VARCHAR(50),
    status                  vendor_status DEFAULT 'pending',
    risk_tier               vendor_risk_tier DEFAULT 'medium',
    risk_score              DECIMAL(5,2) DEFAULT 0,
    service_description     TEXT,
    data_processing         BOOLEAN DEFAULT false,
    data_categories         TEXT[] DEFAULT '{}',
    contract_start_date     DATE,
    contract_end_date       DATE,
    contract_value          DECIMAL(15,2) DEFAULT 0,
    last_assessment_date    DATE,
    next_assessment_date    DATE,
    assessment_frequency    VARCHAR(20) DEFAULT 'annually',
    certifications          TEXT[] DEFAULT '{}',
    dpa_in_place            BOOLEAN DEFAULT false,
    dpa_signed_date         DATE,
    owner_user_id           UUID REFERENCES users(id),
    tags                    TEXT[] DEFAULT '{}',
    metadata                JSONB DEFAULT '{}',
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at              TIMESTAMPTZ,
    UNIQUE(organization_id, vendor_ref)
);

CREATE INDEX idx_vendors_org ON vendors(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_vendors_status ON vendors(organization_id, status) WHERE deleted_at IS NULL;
CREATE INDEX idx_vendors_risk ON vendors(organization_id, risk_tier) WHERE deleted_at IS NULL;

ALTER TABLE vendors ENABLE ROW LEVEL SECURITY;
CREATE POLICY vendors_tenant ON vendors
    USING (organization_id = get_current_tenant() OR get_current_tenant() IS NULL);

CREATE TRIGGER trg_vendors_updated_at BEFORE UPDATE ON vendors FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================
-- ASSETS
-- ============================================================

CREATE TABLE assets (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id             UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    asset_ref                   VARCHAR(20) NOT NULL,
    name                        VARCHAR(255) NOT NULL,
    asset_type                  VARCHAR(30) NOT NULL,
    category                    VARCHAR(100),
    description                 TEXT,
    status                      VARCHAR(20) DEFAULT 'active',
    criticality                 VARCHAR(10) DEFAULT 'medium',
    owner_user_id               UUID REFERENCES users(id),
    custodian_user_id           UUID REFERENCES users(id),
    location                    VARCHAR(255),
    ip_address                  INET,
    classification              classification_level DEFAULT 'internal',
    processes_personal_data     BOOLEAN DEFAULT false,
    linked_vendor_id            UUID REFERENCES vendors(id),
    tags                        TEXT[] DEFAULT '{}',
    metadata                    JSONB DEFAULT '{}',
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at                  TIMESTAMPTZ,
    UNIQUE(organization_id, asset_ref)
);

CREATE INDEX idx_assets_org ON assets(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_assets_type ON assets(organization_id, asset_type) WHERE deleted_at IS NULL;

ALTER TABLE assets ENABLE ROW LEVEL SECURITY;
CREATE POLICY assets_tenant ON assets
    USING (organization_id = get_current_tenant() OR get_current_tenant() IS NULL);

CREATE TRIGGER trg_assets_updated_at BEFORE UPDATE ON assets FOR EACH ROW EXECUTE FUNCTION update_updated_at();
