-- ============================================================
-- ComplianceForge GRC Platform
-- Migration 018: Regulatory Change Management & Horizon Scanning
-- Tables for tracking regulatory sources, changes, org subscriptions,
-- and per-org impact assessments with automated RSS-based scanning.
-- ============================================================

BEGIN;

-- ============================================================
-- ENUM TYPES
-- ============================================================

CREATE TYPE regulatory_source_type AS ENUM (
    'supervisory_authority',
    'standards_body',
    'government',
    'industry_body',
    'legal_publisher',
    'custom'
);

CREATE TYPE scan_frequency AS ENUM ('hourly', 'daily', 'weekly');

CREATE TYPE regulatory_change_type AS ENUM (
    'new_regulation',
    'amendment',
    'guidance',
    'enforcement_decision',
    'standard_revision',
    'standard_update',
    'industry_bulletin',
    'court_ruling',
    'consultation'
);

CREATE TYPE regulatory_severity AS ENUM (
    'critical',
    'high',
    'medium',
    'low',
    'informational'
);

CREATE TYPE regulatory_change_status AS ENUM (
    'new',
    'under_assessment',
    'assessed',
    'action_required',
    'implemented',
    'not_applicable',
    'monitoring'
);

CREATE TYPE regulatory_impact_level AS ENUM (
    'none',
    'low',
    'moderate',
    'significant',
    'critical'
);

CREATE TYPE impact_assessment_status AS ENUM (
    'pending',
    'in_progress',
    'completed'
);

-- ============================================================
-- TABLE: regulatory_sources
-- Tracks external regulatory bodies, standards organisations,
-- and legal publishers whose output is scanned for changes.
-- ============================================================

CREATE TABLE regulatory_sources (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name                VARCHAR(300) NOT NULL,
    source_type         regulatory_source_type NOT NULL,
    country_code        VARCHAR(5),
    region              VARCHAR(50),
    url                 TEXT,
    rss_feed_url        TEXT,
    api_url             TEXT,
    relevance_frameworks TEXT[] DEFAULT '{}',
    scan_frequency      scan_frequency NOT NULL DEFAULT 'daily',
    last_scanned_at     TIMESTAMPTZ,
    is_active           BOOLEAN NOT NULL DEFAULT true,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_regulatory_sources_active ON regulatory_sources(is_active);
CREATE INDEX idx_regulatory_sources_frequency ON regulatory_sources(scan_frequency) WHERE is_active = true;
CREATE INDEX idx_regulatory_sources_type ON regulatory_sources(source_type);

CREATE TRIGGER trg_regulatory_sources_updated_at
    BEFORE UPDATE ON regulatory_sources
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================
-- FUNCTION: generate_change_ref
-- Auto-generates RC-YYYY-NNNN references for regulatory changes.
-- ============================================================

CREATE OR REPLACE FUNCTION generate_change_ref()
RETURNS VARCHAR AS $$
DECLARE
    next_num INT;
    current_year TEXT;
BEGIN
    current_year := EXTRACT(YEAR FROM NOW())::TEXT;
    SELECT COALESCE(MAX(
        CAST(SUBSTRING(change_ref FROM 9) AS INT)
    ), 0) + 1
    INTO next_num
    FROM regulatory_changes
    WHERE change_ref LIKE 'RC-' || current_year || '-%';
    RETURN 'RC-' || current_year || '-' || LPAD(next_num::TEXT, 4, '0');
END;
$$ LANGUAGE plpgsql;

-- ============================================================
-- TABLE: regulatory_changes
-- Central register of detected regulatory changes, amendments,
-- guidance updates, enforcement decisions, and standards revisions.
-- ============================================================

CREATE TABLE regulatory_changes (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_id               UUID REFERENCES regulatory_sources(id) ON DELETE SET NULL,
    change_ref              VARCHAR(30) NOT NULL UNIQUE DEFAULT generate_change_ref(),
    title                   VARCHAR(500) NOT NULL,
    summary                 TEXT,
    full_text_url           TEXT,
    published_date          DATE,
    effective_date          DATE,
    change_type             regulatory_change_type,
    severity                regulatory_severity NOT NULL DEFAULT 'medium',
    status                  regulatory_change_status NOT NULL DEFAULT 'new',
    affected_frameworks     TEXT[] DEFAULT '{}',
    affected_regions        TEXT[] DEFAULT '{}',
    affected_industries     TEXT[] DEFAULT '{}',
    affected_control_codes  TEXT[] DEFAULT '{}',
    impact_assessment       TEXT,
    impact_level            regulatory_impact_level,
    compliance_gap_created  BOOLEAN NOT NULL DEFAULT false,
    required_actions        TEXT,
    deadline                DATE,
    assessed_by             UUID REFERENCES users(id) ON DELETE SET NULL,
    assessed_at             TIMESTAMPTZ,
    response_plan_id        UUID,
    assigned_to             UUID REFERENCES users(id) ON DELETE SET NULL,
    notes                   TEXT,
    tags                    TEXT[] DEFAULT '{}',
    metadata                JSONB DEFAULT '{}',
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_reg_changes_source ON regulatory_changes(source_id);
CREATE INDEX idx_reg_changes_status ON regulatory_changes(status);
CREATE INDEX idx_reg_changes_severity ON regulatory_changes(severity);
CREATE INDEX idx_reg_changes_published ON regulatory_changes(published_date DESC);
CREATE INDEX idx_reg_changes_ref ON regulatory_changes(change_ref);
CREATE INDEX idx_reg_changes_type ON regulatory_changes(change_type);
CREATE INDEX idx_reg_changes_effective ON regulatory_changes(effective_date) WHERE effective_date IS NOT NULL;
CREATE INDEX idx_reg_changes_deadline ON regulatory_changes(deadline) WHERE deadline IS NOT NULL;
CREATE INDEX idx_reg_changes_frameworks ON regulatory_changes USING gin(affected_frameworks);
CREATE INDEX idx_reg_changes_regions ON regulatory_changes USING gin(affected_regions);

CREATE TRIGGER trg_regulatory_changes_updated_at
    BEFORE UPDATE ON regulatory_changes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================
-- TABLE: regulatory_subscriptions
-- Per-org subscriptions to regulatory sources, controlling
-- which sources an organisation monitors and notification prefs.
-- ============================================================

CREATE TABLE regulatory_subscriptions (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id             UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    source_id                   UUID NOT NULL REFERENCES regulatory_sources(id) ON DELETE CASCADE,
    is_active                   BOOLEAN NOT NULL DEFAULT true,
    notification_on_new         BOOLEAN NOT NULL DEFAULT true,
    notification_severity_filter TEXT[] DEFAULT '{}',
    auto_assess                 BOOLEAN NOT NULL DEFAULT false,
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_reg_subscription_org_source UNIQUE (organization_id, source_id)
);

CREATE INDEX idx_reg_subs_org ON regulatory_subscriptions(organization_id);
CREATE INDEX idx_reg_subs_source ON regulatory_subscriptions(source_id);
CREATE INDEX idx_reg_subs_active ON regulatory_subscriptions(organization_id, is_active) WHERE is_active = true;

ALTER TABLE regulatory_subscriptions ENABLE ROW LEVEL SECURITY;
CREATE POLICY regulatory_subscriptions_tenant ON regulatory_subscriptions
    USING (organization_id = get_current_tenant() OR get_current_tenant() IS NULL);

-- ============================================================
-- TABLE: regulatory_impact_assessments
-- Per-organisation impact assessment for a specific regulatory
-- change, including gap analysis, cost estimates, and plans.
-- ============================================================

CREATE TABLE regulatory_impact_assessments (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id         UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    change_id               UUID NOT NULL REFERENCES regulatory_changes(id) ON DELETE CASCADE,
    status                  impact_assessment_status NOT NULL DEFAULT 'pending',
    impact_on_frameworks    JSONB DEFAULT '{}',
    gap_analysis            JSONB DEFAULT '{}',
    existing_coverage       DECIMAL(5,2) DEFAULT 0.00,
    estimated_effort_hours  DECIMAL(8,1) DEFAULT 0.0,
    estimated_cost_eur      DECIMAL(10,2) DEFAULT 0.00,
    ai_assessment           TEXT,
    human_assessment        TEXT,
    assessed_by             UUID REFERENCES users(id) ON DELETE SET NULL,
    assessed_at             TIMESTAMPTZ,
    remediation_plan_id     UUID,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_reg_impact_org_change UNIQUE (organization_id, change_id)
);

CREATE INDEX idx_reg_impact_org ON regulatory_impact_assessments(organization_id);
CREATE INDEX idx_reg_impact_change ON regulatory_impact_assessments(change_id);
CREATE INDEX idx_reg_impact_status ON regulatory_impact_assessments(organization_id, status);

ALTER TABLE regulatory_impact_assessments ENABLE ROW LEVEL SECURITY;
CREATE POLICY regulatory_impact_assessments_tenant ON regulatory_impact_assessments
    USING (organization_id = get_current_tenant() OR get_current_tenant() IS NULL);

CREATE TRIGGER trg_regulatory_impact_assessments_updated_at
    BEFORE UPDATE ON regulatory_impact_assessments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

COMMIT;
