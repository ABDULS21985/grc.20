-- 000016_remediation_ai.down.sql
-- Rollback: Remove AI-Assisted Compliance Remediation Planner tables, types, and policies.

BEGIN;

-- Drop triggers
DROP TRIGGER IF EXISTS trg_remediation_actions_updated_at ON remediation_actions;
DROP FUNCTION IF EXISTS update_remediation_actions_updated_at();

DROP TRIGGER IF EXISTS trg_remediation_plans_updated_at ON remediation_plans;
DROP FUNCTION IF EXISTS update_remediation_plans_updated_at();

-- Drop ref generation functions
DROP FUNCTION IF EXISTS generate_remediation_action_ref(UUID);
DROP FUNCTION IF EXISTS generate_remediation_plan_ref(UUID);

-- Drop tables (order matters due to foreign keys)
DROP TABLE IF EXISTS ai_interaction_logs CASCADE;
DROP TABLE IF EXISTS remediation_actions CASCADE;
DROP TABLE IF EXISTS remediation_plans CASCADE;

-- Drop enum types
DROP TYPE IF EXISTS remediation_action_status;
DROP TYPE IF EXISTS remediation_action_type;
DROP TYPE IF EXISTS remediation_priority;
DROP TYPE IF EXISTS remediation_plan_status;
DROP TYPE IF EXISTS remediation_plan_type;

COMMIT;
