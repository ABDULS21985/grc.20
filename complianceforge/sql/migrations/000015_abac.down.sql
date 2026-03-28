-- 000015_abac.down.sql
-- Rollback: Remove Advanced ABAC Engine tables, types, and policies.

BEGIN;

-- Drop triggers
DROP TRIGGER IF EXISTS trg_access_policies_updated_at ON access_policies;
DROP FUNCTION IF EXISTS update_access_policies_updated_at();

-- Drop tables (order matters due to foreign keys)
DROP TABLE IF EXISTS field_level_permissions CASCADE;
DROP TABLE IF EXISTS access_audit_log CASCADE;
DROP TABLE IF EXISTS access_policy_assignments CASCADE;
DROP TABLE IF EXISTS access_policies CASCADE;

-- Drop enum types
DROP TYPE IF EXISTS field_permission_type;
DROP TYPE IF EXISTS access_decision_type;
DROP TYPE IF EXISTS assignee_type;
DROP TYPE IF EXISTS access_policy_effect;

COMMIT;
