-- ============================================================
-- Rollback Migration 000022: Evidence Template Library & Automated Evidence Testing
-- ============================================================

BEGIN;

-- Drop triggers
DROP TRIGGER IF EXISTS trg_evidence_test_cases_updated_at ON evidence_test_cases;
DROP TRIGGER IF EXISTS trg_evidence_test_suites_updated_at ON evidence_test_suites;
DROP TRIGGER IF EXISTS trg_evidence_requirements_updated_at ON evidence_requirements;
DROP TRIGGER IF EXISTS trg_evidence_templates_updated_at ON evidence_templates;
DROP FUNCTION IF EXISTS update_evidence_templates_updated_at();

-- Drop tables (reverse dependency order)
DROP TABLE IF EXISTS evidence_test_runs CASCADE;
DROP TABLE IF EXISTS evidence_test_cases CASCADE;
DROP TABLE IF EXISTS evidence_test_suites CASCADE;
DROP TABLE IF EXISTS evidence_requirements CASCADE;
DROP TABLE IF EXISTS evidence_templates CASCADE;

-- Drop enum types
DROP TYPE IF EXISTS evidence_run_trigger_type;
DROP TYPE IF EXISTS evidence_test_case_type;
DROP TYPE IF EXISTS evidence_test_run_status;
DROP TYPE IF EXISTS evidence_test_suite_type;
DROP TYPE IF EXISTS evidence_test_type;
DROP TYPE IF EXISTS evidence_validation_status;
DROP TYPE IF EXISTS evidence_requirement_status;
DROP TYPE IF EXISTS evidence_auditor_priority;
DROP TYPE IF EXISTS evidence_difficulty;
DROP TYPE IF EXISTS evidence_collection_frequency;
DROP TYPE IF EXISTS evidence_collection_method;
DROP TYPE IF EXISTS evidence_category_type;

COMMIT;
