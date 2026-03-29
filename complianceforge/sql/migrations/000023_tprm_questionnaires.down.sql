-- ============================================================
-- ComplianceForge GRC Platform
-- Migration 023 DOWN: Drop TPRM Questionnaire tables
-- ============================================================

BEGIN;

-- Drop triggers
DROP TRIGGER IF EXISTS trg_vendor_assessment_responses_updated ON vendor_assessment_responses;
DROP TRIGGER IF EXISTS trg_vendor_assessments_updated ON vendor_assessments;
DROP TRIGGER IF EXISTS trg_assessment_questionnaires_updated ON assessment_questionnaires;
DROP TRIGGER IF EXISTS trg_vendor_assessment_ref ON vendor_assessments;

-- Drop function
DROP FUNCTION IF EXISTS generate_assessment_ref();

-- Drop tables in dependency order
DROP TABLE IF EXISTS vendor_portal_sessions CASCADE;
DROP TABLE IF EXISTS vendor_assessment_responses CASCADE;
DROP TABLE IF EXISTS vendor_assessments CASCADE;
DROP TABLE IF EXISTS questionnaire_questions CASCADE;
DROP TABLE IF EXISTS questionnaire_sections CASCADE;
DROP TABLE IF EXISTS assessment_questionnaires CASCADE;

-- Drop enum types
DROP TYPE IF EXISTS vendor_response_flag;
DROP TYPE IF EXISTS vendor_assessment_pass_fail;
DROP TYPE IF EXISTS vendor_assessment_status;
DROP TYPE IF EXISTS question_risk_impact;
DROP TYPE IF EXISTS question_type;
DROP TYPE IF EXISTS scoring_method_type;
DROP TYPE IF EXISTS questionnaire_status;
DROP TYPE IF EXISTS questionnaire_type;

COMMIT;
