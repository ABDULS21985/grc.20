-- ============================================================
-- ComplianceForge GRC Platform
-- Migration 010 DOWN: Drop NIS2 Compliance Automation tables
-- ============================================================

DROP TABLE IF EXISTS nis2_control_mappings CASCADE;
DROP TABLE IF EXISTS nis2_management_accountability CASCADE;
DROP TABLE IF EXISTS nis2_security_measures CASCADE;
DROP TABLE IF EXISTS nis2_incident_reports CASCADE;
DROP TABLE IF EXISTS nis2_entity_assessment CASCADE;

DROP FUNCTION IF EXISTS generate_nis2_report_ref(UUID);

DROP TYPE IF EXISTS nis2_measure_status;
DROP TYPE IF EXISTS nis2_reporting_status;
DROP TYPE IF EXISTS nis2_entity_type;
