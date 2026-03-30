-- 000031_compliance_as_code.down.sql
-- Rollback: Compliance-as-Code Engine

-- Drop triggers
DROP TRIGGER IF EXISTS trg_cac_drift_events_updated ON cac_drift_events;
DROP TRIGGER IF EXISTS trg_cac_resource_mappings_updated ON cac_resource_mappings;
DROP TRIGGER IF EXISTS trg_cac_sync_runs_updated ON cac_sync_runs;
DROP TRIGGER IF EXISTS trg_cac_repositories_updated ON cac_repositories;

-- Drop function
DROP FUNCTION IF EXISTS update_cac_updated_at();

-- Drop tables (order matters due to foreign keys)
DROP TABLE IF EXISTS cac_drift_events;
DROP TABLE IF EXISTS cac_resource_mappings;
DROP TABLE IF EXISTS cac_sync_runs;
DROP TABLE IF EXISTS cac_repositories;

-- Drop enums
DROP TYPE IF EXISTS cac_diff_action_type;
DROP TYPE IF EXISTS cac_drift_status;
DROP TYPE IF EXISTS cac_drift_direction;
DROP TYPE IF EXISTS cac_resource_status;
DROP TYPE IF EXISTS cac_sync_direction;
DROP TYPE IF EXISTS cac_sync_status;
DROP TYPE IF EXISTS cac_provider;
