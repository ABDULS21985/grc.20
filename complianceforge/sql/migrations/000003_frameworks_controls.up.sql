-- ============================================================
-- ComplianceForge GRC Platform
-- Migration 003: Compliance Frameworks, Controls & Cross-Mappings
-- ============================================================

-- ============================================================
-- COMPLIANCE FRAMEWORKS
-- ============================================================

CREATE TABLE compliance_frameworks (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id         UUID REFERENCES organizations(id) ON DELETE CASCADE,
    code                    VARCHAR(50) NOT NULL,
    name                    VARCHAR(255) NOT NULL,
    full_name               TEXT,
    version                 VARCHAR(20) NOT NULL,
    description             TEXT,
    issuing_body            VARCHAR(255),
    category                VARCHAR(50) NOT NULL, -- security, privacy, governance, risk, operational
    applicable_regions      TEXT[] DEFAULT '{global}',
    applicable_industries   TEXT[] DEFAULT '{all}',
    is_system_framework     BOOLEAN DEFAULT false,
    is_active               BOOLEAN DEFAULT true,
    effective_date          DATE,
    sunset_date             DATE,
    total_controls          INT DEFAULT 0,
    icon_url                TEXT,
    color_hex               VARCHAR(7),
    metadata                JSONB DEFAULT '{}',
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at              TIMESTAMPTZ,
    UNIQUE(COALESCE(organization_id, '00000000-0000-0000-0000-000000000000'::UUID), code, version)
);

CREATE INDEX idx_frameworks_org ON compliance_frameworks(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_frameworks_code ON compliance_frameworks(code) WHERE deleted_at IS NULL;
CREATE INDEX idx_frameworks_active ON compliance_frameworks(is_active) WHERE deleted_at IS NULL;

CREATE TRIGGER trg_frameworks_updated_at
    BEFORE UPDATE ON compliance_frameworks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================
-- FRAMEWORK DOMAINS (Top-level groupings)
-- ============================================================

CREATE TABLE framework_domains (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    framework_id        UUID NOT NULL REFERENCES compliance_frameworks(id) ON DELETE CASCADE,
    code                VARCHAR(50) NOT NULL,
    name                VARCHAR(255) NOT NULL,
    description         TEXT,
    sort_order          INT DEFAULT 0,
    parent_domain_id    UUID REFERENCES framework_domains(id),
    depth_level         INT DEFAULT 0,
    total_controls      INT DEFAULT 0,
    metadata            JSONB DEFAULT '{}',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_domains_framework ON framework_domains(framework_id);
CREATE INDEX idx_domains_parent ON framework_domains(parent_domain_id) WHERE parent_domain_id IS NOT NULL;
CREATE UNIQUE INDEX idx_domains_unique_code ON framework_domains(framework_id, code);

-- ============================================================
-- FRAMEWORK CONTROLS
-- ============================================================

CREATE TABLE framework_controls (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    framework_id            UUID NOT NULL REFERENCES compliance_frameworks(id) ON DELETE CASCADE,
    domain_id               UUID NOT NULL REFERENCES framework_domains(id) ON DELETE CASCADE,
    code                    VARCHAR(100) NOT NULL,
    title                   VARCHAR(500) NOT NULL,
    description             TEXT,
    guidance                TEXT,
    objective               TEXT,
    control_type            control_type DEFAULT 'preventive',
    implementation_type     implementation_type DEFAULT 'administrative',
    is_mandatory            BOOLEAN DEFAULT true,
    priority                VARCHAR(10) DEFAULT 'medium',
    sort_order              INT DEFAULT 0,
    parent_control_id       UUID REFERENCES framework_controls(id),
    depth_level             INT DEFAULT 0,
    evidence_requirements   JSONB DEFAULT '[]',
    test_procedures         JSONB DEFAULT '[]',
    references              JSONB DEFAULT '[]',
    keywords                TEXT[] DEFAULT '{}',
    metadata                JSONB DEFAULT '{}',
    search_vector           tsvector GENERATED ALWAYS AS (
        to_tsvector('english',
            COALESCE(code, '') || ' ' ||
            COALESCE(title, '') || ' ' ||
            COALESCE(description, '')
        )
    ) STORED,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(framework_id, code)
);

CREATE INDEX idx_controls_framework ON framework_controls(framework_id);
CREATE INDEX idx_controls_domain ON framework_controls(domain_id);
CREATE INDEX idx_controls_parent ON framework_controls(parent_control_id) WHERE parent_control_id IS NOT NULL;
CREATE INDEX idx_controls_search ON framework_controls USING GIN(search_vector);
CREATE INDEX idx_controls_keywords ON framework_controls USING GIN(keywords);

-- ============================================================
-- CROSS-FRAMEWORK CONTROL MAPPINGS
-- ============================================================

CREATE TABLE framework_control_mappings (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_control_id   UUID NOT NULL REFERENCES framework_controls(id) ON DELETE CASCADE,
    target_control_id   UUID NOT NULL REFERENCES framework_controls(id) ON DELETE CASCADE,
    mapping_type        mapping_type NOT NULL DEFAULT 'related',
    mapping_strength    DECIMAL(3,2) DEFAULT 0.50,
    notes               TEXT,
    is_verified         BOOLEAN DEFAULT false,
    verified_by         UUID REFERENCES users(id),
    verified_at         TIMESTAMPTZ,
    metadata            JSONB DEFAULT '{}',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(source_control_id, target_control_id),
    CHECK(source_control_id <> target_control_id),
    CHECK(mapping_strength >= 0.00 AND mapping_strength <= 1.00)
);

CREATE INDEX idx_mappings_source ON framework_control_mappings(source_control_id);
CREATE INDEX idx_mappings_target ON framework_control_mappings(target_control_id);

-- ============================================================
-- ORGANIZATION FRAMEWORK ADOPTION
-- ============================================================

CREATE TABLE organization_frameworks (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id         UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    framework_id            UUID NOT NULL REFERENCES compliance_frameworks(id) ON DELETE CASCADE,
    status                  VARCHAR(30) NOT NULL DEFAULT 'not_started',
    adoption_date           DATE,
    target_completion_date  DATE,
    certification_date      DATE,
    certification_expiry    DATE,
    certifying_body         VARCHAR(255),
    certificate_number      VARCHAR(100),
    scope_description       TEXT,
    scope_business_units    UUID[] DEFAULT '{}',
    compliance_score        DECIMAL(5,2) DEFAULT 0.00,
    last_assessment_date    TIMESTAMPTZ,
    assessment_frequency    VARCHAR(20) DEFAULT 'quarterly',
    responsible_user_id     UUID REFERENCES users(id),
    metadata                JSONB DEFAULT '{}',
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(organization_id, framework_id)
);

CREATE INDEX idx_org_frameworks_org ON organization_frameworks(organization_id);
CREATE INDEX idx_org_frameworks_status ON organization_frameworks(organization_id, status);

ALTER TABLE organization_frameworks ENABLE ROW LEVEL SECURITY;
CREATE POLICY org_frameworks_tenant ON organization_frameworks
    USING (organization_id = get_current_tenant() OR get_current_tenant() IS NULL);

CREATE TRIGGER trg_org_frameworks_updated_at
    BEFORE UPDATE ON organization_frameworks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================
-- CONTROL IMPLEMENTATIONS
-- ============================================================

CREATE TABLE control_implementations (
    id                                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id                     UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    framework_control_id                UUID NOT NULL REFERENCES framework_controls(id) ON DELETE CASCADE,
    org_framework_id                    UUID NOT NULL REFERENCES organization_frameworks(id) ON DELETE CASCADE,
    status                              control_status DEFAULT 'not_implemented',
    implementation_status               VARCHAR(30) DEFAULT 'not_started',
    maturity_level                      INT DEFAULT 0 CHECK(maturity_level >= 0 AND maturity_level <= 5),
    owner_user_id                       UUID REFERENCES users(id),
    reviewer_user_id                    UUID REFERENCES users(id),
    implementation_description          TEXT,
    implementation_notes                TEXT,
    compensating_control_description    TEXT,
    gap_description                     TEXT,
    remediation_plan                    TEXT,
    remediation_due_date                DATE,
    test_frequency                      VARCHAR(20) DEFAULT 'quarterly',
    last_tested_at                      TIMESTAMPTZ,
    last_tested_by                      UUID REFERENCES users(id),
    last_test_result                    VARCHAR(20),
    effectiveness_score                 DECIMAL(5,2) DEFAULT 0.00,
    risk_if_not_implemented             VARCHAR(10) DEFAULT 'medium',
    automation_level                    VARCHAR(20) DEFAULT 'manual',
    tags                                TEXT[] DEFAULT '{}',
    metadata                            JSONB DEFAULT '{}',
    created_at                          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at                          TIMESTAMPTZ,
    UNIQUE(organization_id, framework_control_id)
);

CREATE INDEX idx_ctrl_impl_org ON control_implementations(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_ctrl_impl_status ON control_implementations(organization_id, status) WHERE deleted_at IS NULL;
CREATE INDEX idx_ctrl_impl_framework ON control_implementations(org_framework_id);
CREATE INDEX idx_ctrl_impl_owner ON control_implementations(owner_user_id) WHERE owner_user_id IS NOT NULL;

ALTER TABLE control_implementations ENABLE ROW LEVEL SECURITY;
CREATE POLICY ctrl_impl_tenant ON control_implementations
    USING (organization_id = get_current_tenant() OR get_current_tenant() IS NULL);

CREATE TRIGGER trg_ctrl_impl_updated_at
    BEFORE UPDATE ON control_implementations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================
-- CONTROL EVIDENCE
-- ============================================================

CREATE TABLE control_evidence (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id             UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    control_implementation_id   UUID NOT NULL REFERENCES control_implementations(id) ON DELETE CASCADE,
    title                       VARCHAR(500) NOT NULL,
    description                 TEXT,
    evidence_type               VARCHAR(30) NOT NULL,
    file_path                   TEXT,
    file_name                   VARCHAR(255),
    file_size_bytes             BIGINT DEFAULT 0,
    mime_type                   VARCHAR(100),
    file_hash                   VARCHAR(128),
    collection_method           VARCHAR(30) DEFAULT 'manual_upload',
    collected_at                TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    collected_by                UUID NOT NULL REFERENCES users(id),
    valid_from                  DATE,
    valid_until                 DATE,
    is_current                  BOOLEAN DEFAULT true,
    review_status               VARCHAR(20) DEFAULT 'pending',
    reviewed_by                 UUID REFERENCES users(id),
    reviewed_at                 TIMESTAMPTZ,
    review_notes                TEXT,
    metadata                    JSONB DEFAULT '{}',
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at                  TIMESTAMPTZ
);

CREATE INDEX idx_evidence_org ON control_evidence(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_evidence_ctrl ON control_evidence(control_implementation_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_evidence_current ON control_evidence(control_implementation_id, is_current) WHERE is_current = true;

ALTER TABLE control_evidence ENABLE ROW LEVEL SECURITY;
CREATE POLICY evidence_tenant ON control_evidence
    USING (organization_id = get_current_tenant() OR get_current_tenant() IS NULL);

-- ============================================================
-- CONTROL TEST RESULTS
-- ============================================================

CREATE TABLE control_test_results (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id             UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    control_implementation_id   UUID NOT NULL REFERENCES control_implementations(id) ON DELETE CASCADE,
    test_type                   VARCHAR(30) NOT NULL,
    test_procedure              TEXT,
    result                      VARCHAR(20) NOT NULL,
    findings                    TEXT,
    recommendations             TEXT,
    tested_by                   UUID NOT NULL REFERENCES users(id),
    tested_at                   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    evidence_ids                UUID[] DEFAULT '{}',
    next_test_date              DATE,
    metadata                    JSONB DEFAULT '{}',
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_test_results_org ON control_test_results(organization_id);
CREATE INDEX idx_test_results_ctrl ON control_test_results(control_implementation_id);

ALTER TABLE control_test_results ENABLE ROW LEVEL SECURITY;
CREATE POLICY test_results_tenant ON control_test_results
    USING (organization_id = get_current_tenant() OR get_current_tenant() IS NULL);

-- ============================================================
-- VIEWS
-- ============================================================

-- Compliance score aggregation per framework per org
CREATE OR REPLACE VIEW v_compliance_score_by_framework AS
SELECT
    ci.organization_id,
    of2.framework_id,
    cf.code AS framework_code,
    cf.name AS framework_name,
    COUNT(*) AS total_controls,
    COUNT(*) FILTER (WHERE ci.status = 'implemented' OR ci.status = 'effective') AS implemented,
    COUNT(*) FILTER (WHERE ci.status = 'partial') AS partially_implemented,
    COUNT(*) FILTER (WHERE ci.status = 'not_implemented' OR ci.status = 'planned') AS not_implemented,
    COUNT(*) FILTER (WHERE ci.status = 'not_applicable') AS not_applicable,
    ROUND(
        (COUNT(*) FILTER (WHERE ci.status IN ('implemented', 'effective'))::DECIMAL /
        NULLIF(COUNT(*) FILTER (WHERE ci.status <> 'not_applicable'), 0)) * 100, 2
    ) AS compliance_score,
    ROUND(AVG(ci.maturity_level)::DECIMAL, 2) AS maturity_avg
FROM control_implementations ci
JOIN organization_frameworks of2 ON ci.org_framework_id = of2.id
JOIN compliance_frameworks cf ON of2.framework_id = cf.id
WHERE ci.deleted_at IS NULL
GROUP BY ci.organization_id, of2.framework_id, cf.code, cf.name;

-- Control gap analysis
CREATE OR REPLACE VIEW v_control_gap_analysis AS
SELECT
    ci.organization_id,
    cf.code AS framework_code,
    cf.name AS framework_name,
    fc.code AS control_code,
    fc.title AS control_title,
    ci.status,
    ci.maturity_level,
    ci.gap_description,
    ci.remediation_plan,
    ci.remediation_due_date,
    ci.risk_if_not_implemented,
    u.first_name || ' ' || u.last_name AS owner_name
FROM control_implementations ci
JOIN framework_controls fc ON ci.framework_control_id = fc.id
JOIN compliance_frameworks cf ON fc.framework_id = cf.id
LEFT JOIN users u ON ci.owner_user_id = u.id
WHERE ci.status NOT IN ('implemented', 'effective', 'not_applicable')
AND ci.deleted_at IS NULL;
