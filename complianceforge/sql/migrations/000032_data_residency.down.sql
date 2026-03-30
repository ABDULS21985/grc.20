-- 000032_data_residency.down.sql
-- Rollback: Multi-Region Deployment & Data Residency

-- Drop indexes
DROP INDEX IF EXISTS idx_residency_audit_ip;
DROP INDEX IF EXISTS idx_residency_audit_source_region;
DROP INDEX IF EXISTS idx_residency_audit_allowed;
DROP INDEX IF EXISTS idx_residency_audit_action;
DROP INDEX IF EXISTS idx_residency_audit_org_created;
DROP INDEX IF EXISTS idx_residency_audit_org;
DROP INDEX IF EXISTS idx_data_residency_configs_active;
DROP INDEX IF EXISTS idx_data_residency_configs_region;
DROP INDEX IF EXISTS idx_organizations_data_region;

-- Drop triggers
DROP TRIGGER IF EXISTS set_data_residency_configs_updated_at ON data_residency_configs;

-- Drop RLS policies
DROP POLICY IF EXISTS data_residency_audit_log_org_isolation ON data_residency_audit_log;

-- Drop tables
DROP TABLE IF EXISTS data_residency_audit_log;
DROP TABLE IF EXISTS data_residency_configs;

-- Remove columns from organizations
ALTER TABLE organizations
    DROP COLUMN IF EXISTS data_region,
    DROP COLUMN IF EXISTS data_residency_enforced;

-- Drop enums
DROP TYPE IF EXISTS audit_action_type;
DROP TYPE IF EXISTS residency_enforcement_mode;
DROP TYPE IF EXISTS data_residency_region;
