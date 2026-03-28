-- ============================================================
-- Migration 013: Integration Hub
-- Provides integration adapters, SSO configuration, and API key
-- management for the ComplianceForge GRC platform.
-- ============================================================

-- ============================================================
-- ENUMS
-- ============================================================

CREATE TYPE integration_type AS ENUM (
    'sso_saml',
    'sso_oidc',
    'cloud_aws',
    'cloud_azure',
    'cloud_gcp',
    'siem_splunk',
    'siem_elastic',
    'siem_sentinel',
    'itsm_servicenow',
    'itsm_jira',
    'itsm_freshservice',
    'email_smtp',
    'email_sendgrid',
    'slack',
    'teams',
    'webhook_inbound',
    'webhook_outbound',
    'custom_api'
);

CREATE TYPE integration_status AS ENUM (
    'active',
    'inactive',
    'error',
    'pending_setup'
);

CREATE TYPE integration_health_status AS ENUM (
    'healthy',
    'degraded',
    'unhealthy',
    'unknown'
);

CREATE TYPE sync_status AS ENUM (
    'running',
    'completed',
    'failed',
    'cancelled'
);

CREATE TYPE sso_protocol AS ENUM (
    'saml2',
    'oidc'
);

-- ============================================================
-- TABLE: integrations
-- Stores all third-party integration configurations.
-- ============================================================

CREATE TABLE integrations (
    id                       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id          UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    integration_type         integration_type NOT NULL,
    name                     VARCHAR(255) NOT NULL,
    description              TEXT,
    status                   integration_status NOT NULL DEFAULT 'pending_setup',
    configuration_encrypted  TEXT,
    health_status            integration_health_status NOT NULL DEFAULT 'unknown',
    last_health_check_at     TIMESTAMPTZ,
    last_sync_at             TIMESTAMPTZ,
    sync_frequency_minutes   INTEGER DEFAULT 60,
    error_count              INTEGER NOT NULL DEFAULT 0,
    last_error_message       TEXT,
    capabilities             TEXT[],
    created_by               UUID REFERENCES users(id) ON DELETE SET NULL,
    metadata                 JSONB NOT NULL DEFAULT '{}',
    created_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at               TIMESTAMPTZ
);

-- RLS
ALTER TABLE integrations ENABLE ROW LEVEL SECURITY;

CREATE POLICY integrations_tenant_isolation ON integrations
    USING (organization_id = current_setting('app.current_org_id', true)::uuid);

CREATE POLICY integrations_tenant_insert ON integrations
    FOR INSERT WITH CHECK (organization_id = current_setting('app.current_org_id', true)::uuid);

-- Indexes
CREATE INDEX idx_integrations_org_id ON integrations(organization_id);
CREATE INDEX idx_integrations_org_type ON integrations(organization_id, integration_type);
CREATE INDEX idx_integrations_org_status ON integrations(organization_id, status);
CREATE INDEX idx_integrations_health ON integrations(organization_id, health_status);
CREATE INDEX idx_integrations_deleted ON integrations(deleted_at) WHERE deleted_at IS NULL;

-- Trigger: auto-update updated_at
CREATE TRIGGER set_integrations_updated_at
    BEFORE UPDATE ON integrations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================
-- TABLE: integration_sync_logs
-- Records every synchronisation run for auditing.
-- ============================================================

CREATE TABLE integration_sync_logs (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id   UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    integration_id    UUID NOT NULL REFERENCES integrations(id) ON DELETE CASCADE,
    sync_type         VARCHAR(100) NOT NULL DEFAULT 'full',
    status            sync_status NOT NULL DEFAULT 'running',
    records_processed INTEGER NOT NULL DEFAULT 0,
    records_created   INTEGER NOT NULL DEFAULT 0,
    records_updated   INTEGER NOT NULL DEFAULT 0,
    records_failed    INTEGER NOT NULL DEFAULT 0,
    started_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at      TIMESTAMPTZ,
    duration_ms       INTEGER,
    error_message     TEXT,
    details           JSONB NOT NULL DEFAULT '{}',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- RLS
ALTER TABLE integration_sync_logs ENABLE ROW LEVEL SECURITY;

CREATE POLICY sync_logs_tenant_isolation ON integration_sync_logs
    USING (organization_id = current_setting('app.current_org_id', true)::uuid);

CREATE POLICY sync_logs_tenant_insert ON integration_sync_logs
    FOR INSERT WITH CHECK (organization_id = current_setting('app.current_org_id', true)::uuid);

-- Indexes
CREATE INDEX idx_sync_logs_org_id ON integration_sync_logs(organization_id);
CREATE INDEX idx_sync_logs_integration ON integration_sync_logs(integration_id);
CREATE INDEX idx_sync_logs_integration_status ON integration_sync_logs(integration_id, status);
CREATE INDEX idx_sync_logs_started ON integration_sync_logs(started_at DESC);

-- ============================================================
-- TABLE: sso_configurations
-- Stores SAML 2.0 / OIDC configuration per organisation.
-- ============================================================

CREATE TABLE sso_configurations (
    id                           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id              UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    protocol                     sso_protocol NOT NULL DEFAULT 'oidc',
    is_enabled                   BOOLEAN NOT NULL DEFAULT false,
    is_enforced                  BOOLEAN NOT NULL DEFAULT false,

    -- SAML 2.0 fields
    saml_entity_id               VARCHAR(512),
    saml_sso_url                 VARCHAR(1024),
    saml_slo_url                 VARCHAR(1024),
    saml_certificate             TEXT,
    saml_name_id_format          VARCHAR(255) DEFAULT 'urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress',
    saml_attribute_mapping       JSONB NOT NULL DEFAULT '{"email":"email","first_name":"firstName","last_name":"lastName","groups":"groups"}',

    -- OIDC fields
    oidc_issuer_url              VARCHAR(1024),
    oidc_client_id               VARCHAR(512),
    oidc_client_secret_encrypted TEXT,
    oidc_scopes                  TEXT[] DEFAULT ARRAY['openid','profile','email'],
    oidc_claim_mapping           JSONB NOT NULL DEFAULT '{"email":"email","name":"name","groups":"groups"}',

    -- Provisioning
    auto_provision_users         BOOLEAN NOT NULL DEFAULT false,
    default_role_id              UUID REFERENCES roles(id) ON DELETE SET NULL,
    allowed_domains              TEXT[],
    group_to_role_mapping        JSONB NOT NULL DEFAULT '{}',
    jit_provisioning             BOOLEAN NOT NULL DEFAULT false,

    created_at                   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                   TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_sso_org UNIQUE (organization_id)
);

-- RLS
ALTER TABLE sso_configurations ENABLE ROW LEVEL SECURITY;

CREATE POLICY sso_config_tenant_isolation ON sso_configurations
    USING (organization_id = current_setting('app.current_org_id', true)::uuid);

CREATE POLICY sso_config_tenant_insert ON sso_configurations
    FOR INSERT WITH CHECK (organization_id = current_setting('app.current_org_id', true)::uuid);

-- Indexes
CREATE INDEX idx_sso_config_org ON sso_configurations(organization_id);

-- Trigger: auto-update updated_at
CREATE TRIGGER set_sso_configurations_updated_at
    BEFORE UPDATE ON sso_configurations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================
-- TABLE: api_keys
-- Stores hashed API keys for programmatic access.
-- ============================================================

CREATE TABLE api_keys (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id      UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name                 VARCHAR(255) NOT NULL,
    key_prefix           VARCHAR(10) NOT NULL,
    key_hash             VARCHAR(128) NOT NULL,
    permissions          TEXT[] NOT NULL DEFAULT '{}',
    rate_limit_per_minute INTEGER NOT NULL DEFAULT 60,
    expires_at           TIMESTAMPTZ,
    last_used_at         TIMESTAMPTZ,
    last_used_ip         VARCHAR(45),
    is_active            BOOLEAN NOT NULL DEFAULT true,
    created_by           UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- RLS
ALTER TABLE api_keys ENABLE ROW LEVEL SECURITY;

CREATE POLICY api_keys_tenant_isolation ON api_keys
    USING (organization_id = current_setting('app.current_org_id', true)::uuid);

CREATE POLICY api_keys_tenant_insert ON api_keys
    FOR INSERT WITH CHECK (organization_id = current_setting('app.current_org_id', true)::uuid);

-- Indexes
CREATE INDEX idx_api_keys_org ON api_keys(organization_id);
CREATE INDEX idx_api_keys_prefix ON api_keys(key_prefix);
CREATE INDEX idx_api_keys_active ON api_keys(organization_id, is_active) WHERE is_active = true;
CREATE UNIQUE INDEX idx_api_keys_hash ON api_keys(key_hash);

-- Trigger: auto-update updated_at
CREATE TRIGGER set_api_keys_updated_at
    BEFORE UPDATE ON api_keys
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================
-- VIEWS
-- ============================================================

-- Integration summary per org
CREATE OR REPLACE VIEW vw_integration_summary AS
SELECT
    i.organization_id,
    COUNT(*)                                                AS total_integrations,
    COUNT(*) FILTER (WHERE i.status = 'active')             AS active_count,
    COUNT(*) FILTER (WHERE i.status = 'error')              AS error_count,
    COUNT(*) FILTER (WHERE i.health_status = 'healthy')     AS healthy_count,
    COUNT(*) FILTER (WHERE i.health_status = 'unhealthy')   AS unhealthy_count,
    COUNT(*) FILTER (WHERE i.health_status = 'degraded')    AS degraded_count,
    MAX(i.last_sync_at)                                     AS last_sync_at
FROM integrations i
WHERE i.deleted_at IS NULL
GROUP BY i.organization_id;
