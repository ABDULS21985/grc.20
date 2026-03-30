-- 000035_developer_portal.up.sql
-- Public API, Webhooks Marketplace & Developer Portal
-- Adds webhook subscriptions, delivery tracking, API usage logging,
-- sandbox environments, and enhanced API key management.

-- ============================================================
-- ENUMS
-- ============================================================

CREATE TYPE webhook_status AS ENUM (
    'active',
    'paused',
    'disabled',
    'error'
);

CREATE TYPE webhook_delivery_status AS ENUM (
    'pending',
    'success',
    'failed',
    'retrying'
);

CREATE TYPE sandbox_status AS ENUM (
    'provisioning',
    'active',
    'expired',
    'destroyed'
);

-- ============================================================
-- TABLE: webhook_subscriptions
-- Stores webhook endpoint registrations per organisation.
-- Each subscription targets a URL and listens for specific events.
-- ============================================================

CREATE TABLE webhook_subscriptions (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id      UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    url                  TEXT NOT NULL,
    description          TEXT NOT NULL DEFAULT '',
    secret               VARCHAR(128) NOT NULL,
    events               TEXT[] NOT NULL DEFAULT '{}',
    status               webhook_status NOT NULL DEFAULT 'active',
    version              VARCHAR(10) NOT NULL DEFAULT '2024-01-01',
    headers              JSONB NOT NULL DEFAULT '{}',
    failure_count        INT NOT NULL DEFAULT 0,
    max_retries          INT NOT NULL DEFAULT 5,
    last_triggered_at    TIMESTAMPTZ,
    last_success_at      TIMESTAMPTZ,
    last_failure_at      TIMESTAMPTZ,
    last_failure_reason  TEXT,
    created_by           UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- TABLE: webhook_deliveries
-- Tracks every webhook delivery attempt with request/response detail.
-- ============================================================

CREATE TABLE webhook_deliveries (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subscription_id      UUID NOT NULL REFERENCES webhook_subscriptions(id) ON DELETE CASCADE,
    organization_id      UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    event_type           VARCHAR(100) NOT NULL,
    payload              JSONB NOT NULL DEFAULT '{}',
    request_headers      JSONB NOT NULL DEFAULT '{}',
    response_status      INT,
    response_body        TEXT,
    response_headers     JSONB NOT NULL DEFAULT '{}',
    status               webhook_delivery_status NOT NULL DEFAULT 'pending',
    attempt_count        INT NOT NULL DEFAULT 0,
    max_attempts         INT NOT NULL DEFAULT 5,
    next_retry_at        TIMESTAMPTZ,
    duration_ms          INT,
    error_message        TEXT,
    idempotency_key      VARCHAR(64) NOT NULL DEFAULT gen_random_uuid()::text,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at         TIMESTAMPTZ
);

-- ============================================================
-- TABLE: api_usage_logs
-- Records every API request for usage analytics and billing.
-- ============================================================

CREATE TABLE api_usage_logs (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id      UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    api_key_id           UUID REFERENCES api_keys(id) ON DELETE SET NULL,
    method               VARCHAR(10) NOT NULL,
    path                 VARCHAR(500) NOT NULL,
    status_code          INT NOT NULL,
    request_size_bytes   INT NOT NULL DEFAULT 0,
    response_size_bytes  INT NOT NULL DEFAULT 0,
    duration_ms          INT NOT NULL DEFAULT 0,
    ip_address           VARCHAR(45),
    user_agent           TEXT,
    error_code           VARCHAR(50),
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- TABLE: sandbox_environments
-- Isolated sandbox environments for API testing per organisation.
-- ============================================================

CREATE TABLE sandbox_environments (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id      UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name                 VARCHAR(255) NOT NULL DEFAULT 'Default Sandbox',
    status               sandbox_status NOT NULL DEFAULT 'provisioning',
    api_base_url         TEXT NOT NULL DEFAULT '',
    sandbox_api_key      VARCHAR(128),
    seed_data_loaded     BOOLEAN NOT NULL DEFAULT false,
    expires_at           TIMESTAMPTZ NOT NULL DEFAULT (NOW() + INTERVAL '30 days'),
    last_accessed_at     TIMESTAMPTZ,
    metadata             JSONB NOT NULL DEFAULT '{}',
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_sandbox_org UNIQUE (organization_id)
);

-- ============================================================
-- ALTER api_keys — add developer portal columns
-- ============================================================

ALTER TABLE api_keys
    ADD COLUMN IF NOT EXISTS scope                TEXT[] NOT NULL DEFAULT '{}',
    ADD COLUMN IF NOT EXISTS tier                  VARCHAR(20) NOT NULL DEFAULT 'standard',
    ADD COLUMN IF NOT EXISTS rate_limit_per_day    INT NOT NULL DEFAULT 10000,
    ADD COLUMN IF NOT EXISTS allowed_ip_ranges     TEXT[] NOT NULL DEFAULT '{}',
    ADD COLUMN IF NOT EXISTS allowed_origins       TEXT[] NOT NULL DEFAULT '{}',
    ADD COLUMN IF NOT EXISTS metadata              JSONB NOT NULL DEFAULT '{}';

-- ============================================================
-- ROW-LEVEL SECURITY
-- ============================================================

ALTER TABLE webhook_subscriptions ENABLE ROW LEVEL SECURITY;
ALTER TABLE webhook_deliveries ENABLE ROW LEVEL SECURITY;
ALTER TABLE api_usage_logs ENABLE ROW LEVEL SECURITY;
ALTER TABLE sandbox_environments ENABLE ROW LEVEL SECURITY;

-- webhook_subscriptions: tenant isolation
CREATE POLICY webhook_subscriptions_org_isolation ON webhook_subscriptions
    USING (organization_id = current_setting('app.current_org', true)::uuid);

CREATE POLICY webhook_subscriptions_org_insert ON webhook_subscriptions
    FOR INSERT WITH CHECK (organization_id = current_setting('app.current_org', true)::uuid);

-- webhook_deliveries: tenant isolation
CREATE POLICY webhook_deliveries_org_isolation ON webhook_deliveries
    USING (organization_id = current_setting('app.current_org', true)::uuid);

CREATE POLICY webhook_deliveries_org_insert ON webhook_deliveries
    FOR INSERT WITH CHECK (organization_id = current_setting('app.current_org', true)::uuid);

-- api_usage_logs: tenant isolation
CREATE POLICY api_usage_logs_org_isolation ON api_usage_logs
    USING (organization_id = current_setting('app.current_org', true)::uuid);

CREATE POLICY api_usage_logs_org_insert ON api_usage_logs
    FOR INSERT WITH CHECK (organization_id = current_setting('app.current_org', true)::uuid);

-- sandbox_environments: tenant isolation
CREATE POLICY sandbox_environments_org_isolation ON sandbox_environments
    USING (organization_id = current_setting('app.current_org', true)::uuid);

CREATE POLICY sandbox_environments_org_insert ON sandbox_environments
    FOR INSERT WITH CHECK (organization_id = current_setting('app.current_org', true)::uuid);

-- ============================================================
-- AUTO-UPDATE updated_at TRIGGERS
-- ============================================================

CREATE TRIGGER set_webhook_subscriptions_updated_at
    BEFORE UPDATE ON webhook_subscriptions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER set_sandbox_environments_updated_at
    BEFORE UPDATE ON sandbox_environments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================
-- INDEXES
-- ============================================================

-- webhook_subscriptions
CREATE INDEX idx_webhook_subs_org ON webhook_subscriptions(organization_id);
CREATE INDEX idx_webhook_subs_org_status ON webhook_subscriptions(organization_id, status)
    WHERE status = 'active';
CREATE INDEX idx_webhook_subs_events ON webhook_subscriptions USING gin(events);

-- webhook_deliveries
CREATE INDEX idx_webhook_del_sub ON webhook_deliveries(subscription_id);
CREATE INDEX idx_webhook_del_org ON webhook_deliveries(organization_id);
CREATE INDEX idx_webhook_del_status ON webhook_deliveries(status)
    WHERE status IN ('pending', 'retrying');
CREATE INDEX idx_webhook_del_created ON webhook_deliveries(subscription_id, created_at DESC);
CREATE INDEX idx_webhook_del_retry ON webhook_deliveries(next_retry_at)
    WHERE status = 'retrying' AND next_retry_at IS NOT NULL;
CREATE INDEX idx_webhook_del_idempotency ON webhook_deliveries(idempotency_key);

-- api_usage_logs
CREATE INDEX idx_api_usage_org ON api_usage_logs(organization_id);
CREATE INDEX idx_api_usage_key ON api_usage_logs(api_key_id);
CREATE INDEX idx_api_usage_created ON api_usage_logs(organization_id, created_at DESC);
CREATE INDEX idx_api_usage_key_created ON api_usage_logs(api_key_id, created_at DESC);
CREATE INDEX idx_api_usage_method_path ON api_usage_logs(organization_id, method, path);

-- sandbox_environments
CREATE INDEX idx_sandbox_org ON sandbox_environments(organization_id);
CREATE INDEX idx_sandbox_status ON sandbox_environments(status)
    WHERE status = 'active';
CREATE INDEX idx_sandbox_expires ON sandbox_environments(expires_at)
    WHERE status = 'active';

-- api_keys new columns
CREATE INDEX idx_api_keys_scope ON api_keys USING gin(scope);
CREATE INDEX idx_api_keys_tier ON api_keys(tier);
