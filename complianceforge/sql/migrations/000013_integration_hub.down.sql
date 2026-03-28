-- ============================================================
-- Migration 013 DOWN: Drop Integration Hub
-- ============================================================

DROP VIEW IF EXISTS vw_integration_summary;

DROP TRIGGER IF EXISTS set_api_keys_updated_at ON api_keys;
DROP TRIGGER IF EXISTS set_sso_configurations_updated_at ON sso_configurations;
DROP TRIGGER IF EXISTS set_integrations_updated_at ON integrations;

DROP TABLE IF EXISTS api_keys CASCADE;
DROP TABLE IF EXISTS sso_configurations CASCADE;
DROP TABLE IF EXISTS integration_sync_logs CASCADE;
DROP TABLE IF EXISTS integrations CASCADE;

DROP TYPE IF EXISTS sso_protocol;
DROP TYPE IF EXISTS sync_status;
DROP TYPE IF EXISTS integration_health_status;
DROP TYPE IF EXISTS integration_status;
DROP TYPE IF EXISTS integration_type;
