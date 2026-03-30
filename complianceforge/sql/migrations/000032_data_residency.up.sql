-- 000032_data_residency.up.sql
-- Multi-Region Deployment & Data Residency

-- ============================================================
-- ENUMS
-- ============================================================

CREATE TYPE data_residency_region AS ENUM (
    'eu',
    'uk',
    'dach',
    'nordics',
    'france',
    'global'
);

CREATE TYPE residency_enforcement_mode AS ENUM (
    'enforce',
    'audit',
    'disabled'
);

CREATE TYPE audit_action_type AS ENUM (
    'data_access',
    'data_export',
    'data_transfer',
    'cross_region_blocked',
    'config_change',
    'region_migration'
);

-- ============================================================
-- ALTER organizations TABLE
-- Add data residency columns to the organisations table.
-- ============================================================

ALTER TABLE organizations
    ADD COLUMN data_region         data_residency_region NOT NULL DEFAULT 'global',
    ADD COLUMN data_residency_enforced BOOLEAN NOT NULL DEFAULT false;

-- ============================================================
-- TABLE: data_residency_configs
-- Stores per-region data residency configuration including
-- allowed cloud regions, compliance frameworks, and storage.
-- ============================================================

CREATE TABLE data_residency_configs (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    region                      data_residency_region NOT NULL UNIQUE,
    display_name                VARCHAR(100) NOT NULL,
    description                 TEXT,

    -- Geographic boundaries
    allowed_countries           TEXT[] NOT NULL DEFAULT '{}',
    blocked_countries           TEXT[] NOT NULL DEFAULT '{}',

    -- Cloud provider regions
    allowed_cloud_regions       JSONB NOT NULL DEFAULT '{}',
    primary_cloud_region        VARCHAR(100),
    failover_cloud_region       VARCHAR(100),

    -- Storage configuration
    storage_config              JSONB NOT NULL DEFAULT '{}'::jsonb,

    -- Compliance & legal
    compliance_frameworks       TEXT[] NOT NULL DEFAULT '{}',
    legal_basis                 TEXT,
    data_protection_authority   VARCHAR(200),
    dpa_contact_url             VARCHAR(500),

    -- GDPR adequacy decisions for transfer validation
    gdpr_adequate_countries     TEXT[] NOT NULL DEFAULT '{}',

    -- Enforcement
    enforcement_mode            residency_enforcement_mode NOT NULL DEFAULT 'enforce',

    -- Feature flags
    allow_cross_region_search   BOOLEAN NOT NULL DEFAULT false,
    allow_cross_region_backup   BOOLEAN NOT NULL DEFAULT false,

    is_active                   BOOLEAN NOT NULL DEFAULT true,
    metadata                    JSONB NOT NULL DEFAULT '{}',
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- TABLE: data_residency_audit_log
-- Immutable audit trail for all data residency events
-- including cross-region access attempts and data transfers.
-- ============================================================

CREATE TABLE data_residency_audit_log (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id             UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    action                      audit_action_type NOT NULL,

    -- Actor
    user_id                     UUID,
    user_email                  VARCHAR(300),
    ip_address                  INET,

    -- Geographic context
    source_region               data_residency_region,
    destination_region          VARCHAR(50),
    source_country              CHAR(2),
    destination_country         CHAR(2),

    -- Resource
    resource_type               VARCHAR(100),
    resource_id                 UUID,

    -- Outcome
    allowed                     BOOLEAN NOT NULL DEFAULT true,
    blocked_reason              TEXT,

    -- Transfer-specific
    vendor_id                   UUID,
    vendor_name                 VARCHAR(200),
    transfer_mechanism          VARCHAR(100),

    -- Details
    details                     JSONB NOT NULL DEFAULT '{}',

    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- ROW-LEVEL SECURITY
-- ============================================================

ALTER TABLE data_residency_audit_log ENABLE ROW LEVEL SECURITY;

-- data_residency_audit_log: only the owning organisation can see its audit entries
CREATE POLICY data_residency_audit_log_org_isolation ON data_residency_audit_log
    USING (organization_id = current_setting('app.current_org', true)::uuid);

-- data_residency_configs is a global reference table, no RLS needed
-- (access controlled at handler level).

-- ============================================================
-- AUTO-UPDATE updated_at TRIGGERS
-- ============================================================

CREATE TRIGGER set_data_residency_configs_updated_at
    BEFORE UPDATE ON data_residency_configs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================
-- INDEXES
-- ============================================================

-- Organizations
CREATE INDEX idx_organizations_data_region ON organizations(data_region);

-- data_residency_configs
CREATE INDEX idx_data_residency_configs_region ON data_residency_configs(region);
CREATE INDEX idx_data_residency_configs_active ON data_residency_configs(is_active)
    WHERE is_active = true;

-- data_residency_audit_log
CREATE INDEX idx_residency_audit_org ON data_residency_audit_log(organization_id);
CREATE INDEX idx_residency_audit_org_created ON data_residency_audit_log(organization_id, created_at DESC);
CREATE INDEX idx_residency_audit_action ON data_residency_audit_log(action);
CREATE INDEX idx_residency_audit_allowed ON data_residency_audit_log(organization_id, allowed)
    WHERE allowed = false;
CREATE INDEX idx_residency_audit_source_region ON data_residency_audit_log(source_region);
CREATE INDEX idx_residency_audit_ip ON data_residency_audit_log(ip_address)
    WHERE ip_address IS NOT NULL;
