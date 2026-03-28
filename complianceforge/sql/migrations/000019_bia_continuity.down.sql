-- 000019_bia_continuity.down.sql
-- Rollback: Remove BIA & Business Continuity tables, types, and policies.

BEGIN;

-- Drop triggers
DROP TRIGGER IF EXISTS trg_process_deps_updated_at ON process_dependencies_map;
DROP FUNCTION IF EXISTS update_process_deps_updated_at();

DROP TRIGGER IF EXISTS trg_bc_exercises_updated_at ON bc_exercises;
DROP FUNCTION IF EXISTS update_bc_exercises_updated_at();

DROP TRIGGER IF EXISTS trg_continuity_plans_updated_at ON continuity_plans;
DROP FUNCTION IF EXISTS update_continuity_plans_updated_at();

DROP TRIGGER IF EXISTS trg_bia_scenarios_updated_at ON bia_scenarios;
DROP FUNCTION IF EXISTS update_bia_scenarios_updated_at();

DROP TRIGGER IF EXISTS trg_business_processes_updated_at ON business_processes;
DROP FUNCTION IF EXISTS update_business_processes_updated_at();

-- Drop ref generation functions
DROP FUNCTION IF EXISTS generate_bcx_ref(UUID);
DROP FUNCTION IF EXISTS generate_bcp_ref(UUID);
DROP FUNCTION IF EXISTS generate_scn_ref(UUID);
DROP FUNCTION IF EXISTS generate_bp_ref(UUID);

-- Drop tables (order matters due to foreign keys)
DROP TABLE IF EXISTS process_dependencies_map CASCADE;
DROP TABLE IF EXISTS bc_exercises CASCADE;
DROP TABLE IF EXISTS continuity_plans CASCADE;
DROP TABLE IF EXISTS bia_scenarios CASCADE;
DROP TABLE IF EXISTS business_processes CASCADE;

-- Drop enum types
DROP TYPE IF EXISTS dependency_type;
DROP TYPE IF EXISTS bc_exercise_rating;
DROP TYPE IF EXISTS bc_exercise_status;
DROP TYPE IF EXISTS bc_exercise_type;
DROP TYPE IF EXISTS continuity_plan_status;
DROP TYPE IF EXISTS continuity_plan_type;
DROP TYPE IF EXISTS bia_scenario_status;
DROP TYPE IF EXISTS bia_likelihood;
DROP TYPE IF EXISTS bia_scenario_type;
DROP TYPE IF EXISTS impact_level;
DROP TYPE IF EXISTS bp_status;
DROP TYPE IF EXISTS bp_criticality;
DROP TYPE IF EXISTS bp_category;

COMMIT;
