-- 000035_developer_portal.down.sql
-- Rollback: Public API, Webhooks Marketplace & Developer Portal

-- ============================================================
-- DROP INDEXES — api_keys new columns
-- ============================================================

DROP INDEX IF EXISTS idx_api_keys_tier;
DROP INDEX IF EXISTS idx_api_keys_scope;

-- ============================================================
-- DROP INDEXES — sandbox_environments
-- ============================================================

DROP INDEX IF EXISTS idx_sandbox_expires;
DROP INDEX IF EXISTS idx_sandbox_status;
DROP INDEX IF EXISTS idx_sandbox_org;

-- ============================================================
-- DROP INDEXES — api_usage_logs
-- ============================================================

DROP INDEX IF EXISTS idx_api_usage_method_path;
DROP INDEX IF EXISTS idx_api_usage_key_created;
DROP INDEX IF EXISTS idx_api_usage_created;
DROP INDEX IF EXISTS idx_api_usage_key;
DROP INDEX IF EXISTS idx_api_usage_org;

-- ============================================================
-- DROP INDEXES — webhook_deliveries
-- ============================================================

DROP INDEX IF EXISTS idx_webhook_del_idempotency;
DROP INDEX IF EXISTS idx_webhook_del_retry;
DROP INDEX IF EXISTS idx_webhook_del_created;
DROP INDEX IF EXISTS idx_webhook_del_status;
DROP INDEX IF EXISTS idx_webhook_del_org;
DROP INDEX IF EXISTS idx_webhook_del_sub;

-- ============================================================
-- DROP INDEXES — webhook_subscriptions
-- ============================================================

DROP INDEX IF EXISTS idx_webhook_subs_events;
DROP INDEX IF EXISTS idx_webhook_subs_org_status;
DROP INDEX IF EXISTS idx_webhook_subs_org;

-- ============================================================
-- DROP TRIGGERS
-- ============================================================

DROP TRIGGER IF EXISTS set_sandbox_environments_updated_at ON sandbox_environments;
DROP TRIGGER IF EXISTS set_webhook_subscriptions_updated_at ON webhook_subscriptions;

-- ============================================================
-- DROP POLICIES
-- ============================================================

DROP POLICY IF EXISTS sandbox_environments_org_insert ON sandbox_environments;
DROP POLICY IF EXISTS sandbox_environments_org_isolation ON sandbox_environments;
DROP POLICY IF EXISTS api_usage_logs_org_insert ON api_usage_logs;
DROP POLICY IF EXISTS api_usage_logs_org_isolation ON api_usage_logs;
DROP POLICY IF EXISTS webhook_deliveries_org_insert ON webhook_deliveries;
DROP POLICY IF EXISTS webhook_deliveries_org_isolation ON webhook_deliveries;
DROP POLICY IF EXISTS webhook_subscriptions_org_insert ON webhook_subscriptions;
DROP POLICY IF EXISTS webhook_subscriptions_org_isolation ON webhook_subscriptions;

-- ============================================================
-- REMOVE COLUMNS FROM api_keys
-- ============================================================

ALTER TABLE api_keys
    DROP COLUMN IF EXISTS metadata,
    DROP COLUMN IF EXISTS allowed_origins,
    DROP COLUMN IF EXISTS allowed_ip_ranges,
    DROP COLUMN IF EXISTS rate_limit_per_day,
    DROP COLUMN IF EXISTS tier,
    DROP COLUMN IF EXISTS scope;

-- ============================================================
-- DROP TABLES (order matters for FK constraints)
-- ============================================================

DROP TABLE IF EXISTS sandbox_environments;
DROP TABLE IF EXISTS api_usage_logs;
DROP TABLE IF EXISTS webhook_deliveries;
DROP TABLE IF EXISTS webhook_subscriptions;

-- ============================================================
-- DROP ENUMS
-- ============================================================

DROP TYPE IF EXISTS sandbox_status;
DROP TYPE IF EXISTS webhook_delivery_status;
DROP TYPE IF EXISTS webhook_status;
