-- ============================================================
-- ComplianceForge GRC Platform
-- Migration 004: Risk Management
-- ============================================================

-- ============================================================
-- RISK CATEGORIES
-- ============================================================

CREATE TABLE risk_categories (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id     UUID REFERENCES organizations(id) ON DELETE CASCADE,
    name                VARCHAR(255) NOT NULL,
    code                VARCHAR(50) NOT NULL,
    description         TEXT,
    parent_category_id  UUID REFERENCES risk_categories(id),
    color_hex           VARCHAR(7),
    icon                VARCHAR(50),
    sort_order          INT DEFAULT 0,
    is_system_default   BOOLEAN DEFAULT false,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_risk_categories_org ON risk_categories(organization_id);

-- ============================================================
-- RISK MATRICES
-- ============================================================

CREATE TABLE risk_matrices (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id     UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name                VARCHAR(255) NOT NULL,
    description         TEXT,
    likelihood_scale    JSONB NOT NULL,
    impact_scale        JSONB NOT NULL,
    risk_levels         JSONB NOT NULL,
    matrix_size         INT DEFAULT 5,
    is_default          BOOLEAN DEFAULT false,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_risk_matrices_org ON risk_matrices(organization_id);

ALTER TABLE risk_matrices ENABLE ROW LEVEL SECURITY;
CREATE POLICY risk_matrices_tenant ON risk_matrices
    USING (organization_id = get_current_tenant() OR get_current_tenant() IS NULL);

-- ============================================================
-- RISK APPETITE STATEMENTS
-- ============================================================

CREATE TABLE risk_appetite_statements (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id             UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    risk_category_id            UUID NOT NULL REFERENCES risk_categories(id),
    appetite_level              VARCHAR(20) NOT NULL,
    appetite_description        TEXT,
    quantitative_threshold_low  DECIMAL(12,2),
    quantitative_threshold_high DECIMAL(12,2),
    threshold_metric            VARCHAR(100),
    tolerance_level             VARCHAR(20),
    approved_by                 UUID REFERENCES users(id),
    approved_at                 TIMESTAMPTZ,
    review_date                 DATE,
    status                      VARCHAR(20) DEFAULT 'draft',
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_risk_appetite_org ON risk_appetite_statements(organization_id);

ALTER TABLE risk_appetite_statements ENABLE ROW LEVEL SECURITY;
CREATE POLICY risk_appetite_tenant ON risk_appetite_statements
    USING (organization_id = get_current_tenant() OR get_current_tenant() IS NULL);

-- ============================================================
-- RISKS (Core Risk Register)
-- ============================================================

-- Sequence for risk references per org
CREATE OR REPLACE FUNCTION generate_risk_ref(org_id UUID)
RETURNS VARCHAR AS $$
DECLARE
    next_num INT;
    ref_val VARCHAR;
BEGIN
    SELECT COALESCE(MAX(
        CAST(SUBSTRING(risk_ref FROM 5) AS INT)
    ), 0) + 1 INTO next_num
    FROM risks WHERE organization_id = org_id;
    ref_val := 'RSK-' || LPAD(next_num::TEXT, 4, '0');
    RETURN ref_val;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE risks (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id         UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    risk_ref                VARCHAR(20) NOT NULL,
    title                   VARCHAR(500) NOT NULL,
    description             TEXT,
    risk_category_id        UUID NOT NULL REFERENCES risk_categories(id),
    risk_source             VARCHAR(100) NOT NULL,
    risk_type               VARCHAR(50),
    status                  risk_status DEFAULT 'identified',
    owner_user_id           UUID REFERENCES users(id),
    delegate_user_id        UUID REFERENCES users(id),
    business_unit_id        UUID,
    risk_matrix_id          UUID REFERENCES risk_matrices(id),
    -- Inherent Risk
    inherent_likelihood     INT CHECK(inherent_likelihood BETWEEN 1 AND 5),
    inherent_impact         INT CHECK(inherent_impact BETWEEN 1 AND 5),
    inherent_risk_score     DECIMAL(5,2) DEFAULT 0,
    inherent_risk_level     risk_level,
    -- Residual Risk
    residual_likelihood     INT CHECK(residual_likelihood BETWEEN 1 AND 5),
    residual_impact         INT CHECK(residual_impact BETWEEN 1 AND 5),
    residual_risk_score     DECIMAL(5,2) DEFAULT 0,
    residual_risk_level     risk_level,
    -- Target Risk
    target_likelihood       INT CHECK(target_likelihood BETWEEN 1 AND 5),
    target_impact           INT CHECK(target_impact BETWEEN 1 AND 5),
    target_risk_score       DECIMAL(5,2) DEFAULT 0,
    target_risk_level       risk_level,
    -- Financial & Impact
    financial_impact_eur    DECIMAL(15,2) DEFAULT 0,
    impact_description      TEXT,
    impact_categories       JSONB DEFAULT '{}',
    risk_velocity           VARCHAR(20),
    risk_proximity          VARCHAR(20),
    -- Dates
    identified_date         DATE DEFAULT CURRENT_DATE,
    last_assessed_date      DATE,
    next_review_date        DATE,
    review_frequency        VARCHAR(20) DEFAULT 'quarterly',
    -- Linkages
    linked_regulations      TEXT[] DEFAULT '{}',
    linked_control_ids      UUID[] DEFAULT '{}',
    tags                    TEXT[] DEFAULT '{}',
    is_emerging             BOOLEAN DEFAULT false,
    metadata                JSONB DEFAULT '{}',
    search_vector           tsvector,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at              TIMESTAMPTZ,
    UNIQUE(organization_id, risk_ref)
);

CREATE INDEX idx_risks_org ON risks(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_risks_status ON risks(organization_id, status) WHERE deleted_at IS NULL;
CREATE INDEX idx_risks_level ON risks(organization_id, residual_risk_level) WHERE deleted_at IS NULL;
CREATE INDEX idx_risks_owner ON risks(owner_user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_risks_category ON risks(risk_category_id);
CREATE INDEX idx_risks_search ON risks USING GIN(search_vector);
CREATE INDEX idx_risks_review ON risks(next_review_date) WHERE deleted_at IS NULL AND next_review_date IS NOT NULL;

ALTER TABLE risks ENABLE ROW LEVEL SECURITY;
CREATE POLICY risks_tenant ON risks
    USING (organization_id = get_current_tenant() OR get_current_tenant() IS NULL);

CREATE TRIGGER trg_risks_updated_at
    BEFORE UPDATE ON risks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- Auto-calculate risk scores
CREATE OR REPLACE FUNCTION calculate_risk_scores()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.inherent_likelihood IS NOT NULL AND NEW.inherent_impact IS NOT NULL THEN
        NEW.inherent_risk_score := NEW.inherent_likelihood * NEW.inherent_impact;
        NEW.inherent_risk_level := CASE
            WHEN NEW.inherent_risk_score >= 20 THEN 'critical'
            WHEN NEW.inherent_risk_score >= 12 THEN 'high'
            WHEN NEW.inherent_risk_score >= 6 THEN 'medium'
            WHEN NEW.inherent_risk_score >= 3 THEN 'low'
            ELSE 'very_low'
        END::risk_level;
    END IF;
    IF NEW.residual_likelihood IS NOT NULL AND NEW.residual_impact IS NOT NULL THEN
        NEW.residual_risk_score := NEW.residual_likelihood * NEW.residual_impact;
        NEW.residual_risk_level := CASE
            WHEN NEW.residual_risk_score >= 20 THEN 'critical'
            WHEN NEW.residual_risk_score >= 12 THEN 'high'
            WHEN NEW.residual_risk_score >= 6 THEN 'medium'
            WHEN NEW.residual_risk_score >= 3 THEN 'low'
            ELSE 'very_low'
        END::risk_level;
    END IF;
    IF NEW.target_likelihood IS NOT NULL AND NEW.target_impact IS NOT NULL THEN
        NEW.target_risk_score := NEW.target_likelihood * NEW.target_impact;
        NEW.target_risk_level := CASE
            WHEN NEW.target_risk_score >= 20 THEN 'critical'
            WHEN NEW.target_risk_score >= 12 THEN 'high'
            WHEN NEW.target_risk_score >= 6 THEN 'medium'
            WHEN NEW.target_risk_score >= 3 THEN 'low'
            ELSE 'very_low'
        END::risk_level;
    END IF;
    -- Update search vector
    NEW.search_vector := to_tsvector('english',
        COALESCE(NEW.risk_ref, '') || ' ' ||
        COALESCE(NEW.title, '') || ' ' ||
        COALESCE(NEW.description, '')
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_risks_calc_scores
    BEFORE INSERT OR UPDATE ON risks
    FOR EACH ROW EXECUTE FUNCTION calculate_risk_scores();

-- ============================================================
-- RISK ASSESSMENTS (Historical)
-- ============================================================

CREATE TABLE risk_assessments (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id     UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    risk_id             UUID NOT NULL REFERENCES risks(id) ON DELETE CASCADE,
    assessment_type     VARCHAR(30) NOT NULL,
    assessor_user_id    UUID REFERENCES users(id),
    assessment_date     DATE NOT NULL DEFAULT CURRENT_DATE,
    likelihood_before   INT,
    impact_before       INT,
    score_before        DECIMAL(5,2),
    likelihood_after    INT,
    impact_after        INT,
    score_after         DECIMAL(5,2),
    assessment_notes    TEXT,
    methodology         VARCHAR(50),
    confidence_level    VARCHAR(20),
    data_sources        TEXT[] DEFAULT '{}',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_risk_assessments_org ON risk_assessments(organization_id);
CREATE INDEX idx_risk_assessments_risk ON risk_assessments(risk_id);

ALTER TABLE risk_assessments ENABLE ROW LEVEL SECURITY;
CREATE POLICY risk_assessments_tenant ON risk_assessments
    USING (organization_id = get_current_tenant() OR get_current_tenant() IS NULL);

-- ============================================================
-- RISK TREATMENTS
-- ============================================================

CREATE TABLE risk_treatments (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id             UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    risk_id                     UUID NOT NULL REFERENCES risks(id) ON DELETE CASCADE,
    treatment_type              treatment_type NOT NULL,
    title                       VARCHAR(500) NOT NULL,
    description                 TEXT,
    status                      VARCHAR(20) DEFAULT 'planned',
    priority                    VARCHAR(10) DEFAULT 'medium',
    owner_user_id               UUID REFERENCES users(id),
    start_date                  DATE,
    target_date                 DATE,
    completed_date              DATE,
    estimated_cost_eur          DECIMAL(12,2) DEFAULT 0,
    actual_cost_eur             DECIMAL(12,2) DEFAULT 0,
    expected_risk_reduction     DECIMAL(5,2) DEFAULT 0,
    progress_percentage         INT DEFAULT 0 CHECK(progress_percentage BETWEEN 0 AND 100),
    linked_control_ids          UUID[] DEFAULT '{}',
    notes                       TEXT,
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_treatments_org ON risk_treatments(organization_id);
CREATE INDEX idx_treatments_risk ON risk_treatments(risk_id);
CREATE INDEX idx_treatments_status ON risk_treatments(organization_id, status);

ALTER TABLE risk_treatments ENABLE ROW LEVEL SECURITY;
CREATE POLICY treatments_tenant ON risk_treatments
    USING (organization_id = get_current_tenant() OR get_current_tenant() IS NULL);

CREATE TRIGGER trg_treatments_updated_at
    BEFORE UPDATE ON risk_treatments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================
-- KEY RISK INDICATORS (KRI)
-- ============================================================

CREATE TABLE risk_indicators (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id         UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    risk_id                 UUID REFERENCES risks(id) ON DELETE SET NULL,
    name                    VARCHAR(255) NOT NULL,
    description             TEXT,
    metric_type             VARCHAR(30) NOT NULL,
    measurement_unit        VARCHAR(50),
    collection_frequency    VARCHAR(20) DEFAULT 'monthly',
    data_source             VARCHAR(255),
    threshold_green         DECIMAL(12,2),
    threshold_amber         DECIMAL(12,2),
    threshold_red           DECIMAL(12,2),
    current_value           DECIMAL(12,2) DEFAULT 0,
    trend                   VARCHAR(10) DEFAULT 'stable',
    owner_user_id           UUID REFERENCES users(id),
    last_updated_at         TIMESTAMPTZ,
    is_automated            BOOLEAN DEFAULT false,
    automation_config       JSONB DEFAULT '{}',
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_indicators_org ON risk_indicators(organization_id);
CREATE INDEX idx_indicators_risk ON risk_indicators(risk_id) WHERE risk_id IS NOT NULL;

ALTER TABLE risk_indicators ENABLE ROW LEVEL SECURITY;
CREATE POLICY indicators_tenant ON risk_indicators
    USING (organization_id = get_current_tenant() OR get_current_tenant() IS NULL);

-- ============================================================
-- KRI VALUES (Historical measurements)
-- ============================================================

CREATE TABLE risk_indicator_values (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    indicator_id    UUID NOT NULL REFERENCES risk_indicators(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    value           DECIMAL(12,2) NOT NULL,
    status          VARCHAR(10) NOT NULL,
    measured_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    measured_by     UUID REFERENCES users(id),
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_indicator_values_indicator ON risk_indicator_values(indicator_id, measured_at DESC);
CREATE INDEX idx_indicator_values_org ON risk_indicator_values(organization_id);

ALTER TABLE risk_indicator_values ENABLE ROW LEVEL SECURITY;
CREATE POLICY indicator_values_tenant ON risk_indicator_values
    USING (organization_id = get_current_tenant() OR get_current_tenant() IS NULL);

-- ============================================================
-- RISK ↔ CONTROL MAPPINGS
-- ============================================================

CREATE TABLE risk_control_mappings (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id             UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    risk_id                     UUID NOT NULL REFERENCES risks(id) ON DELETE CASCADE,
    control_implementation_id   UUID NOT NULL REFERENCES control_implementations(id) ON DELETE CASCADE,
    effectiveness               VARCHAR(20) DEFAULT 'not_tested',
    contribution_percentage     DECIMAL(5,2) DEFAULT 0,
    notes                       TEXT,
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(risk_id, control_implementation_id)
);

CREATE INDEX idx_risk_ctrl_map_org ON risk_control_mappings(organization_id);
CREATE INDEX idx_risk_ctrl_map_risk ON risk_control_mappings(risk_id);

ALTER TABLE risk_control_mappings ENABLE ROW LEVEL SECURITY;
CREATE POLICY risk_ctrl_map_tenant ON risk_control_mappings
    USING (organization_id = get_current_tenant() OR get_current_tenant() IS NULL);

-- ============================================================
-- VIEWS
-- ============================================================

CREATE OR REPLACE VIEW v_risk_heatmap AS
SELECT
    r.organization_id,
    r.id AS risk_id,
    r.risk_ref,
    r.title,
    rc.name AS category_name,
    r.inherent_likelihood,
    r.inherent_impact,
    r.inherent_risk_score,
    r.residual_likelihood,
    r.residual_impact,
    r.residual_risk_score,
    r.residual_risk_level,
    COALESCE(u.first_name || ' ' || u.last_name, 'Unassigned') AS owner_name,
    r.status::TEXT
FROM risks r
LEFT JOIN risk_categories rc ON r.risk_category_id = rc.id
LEFT JOIN users u ON r.owner_user_id = u.id
WHERE r.deleted_at IS NULL;

CREATE OR REPLACE VIEW v_top_risks AS
SELECT
    r.organization_id,
    r.id AS risk_id,
    r.risk_ref,
    r.title,
    rc.name AS category_name,
    r.residual_risk_score,
    r.residual_risk_level,
    r.financial_impact_eur,
    r.status::TEXT,
    r.next_review_date,
    COALESCE(u.first_name || ' ' || u.last_name, 'Unassigned') AS owner_name,
    (SELECT COUNT(*) FROM risk_treatments rt WHERE rt.risk_id = r.id AND rt.status NOT IN ('completed', 'cancelled')) AS open_treatments
FROM risks r
LEFT JOIN risk_categories rc ON r.risk_category_id = rc.id
LEFT JOIN users u ON r.owner_user_id = u.id
WHERE r.deleted_at IS NULL
ORDER BY r.residual_risk_score DESC;
