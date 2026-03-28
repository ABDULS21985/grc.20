-- ============================================================
-- ComplianceForge GRC Platform
-- Migration 011 (DOWN): Drop Continuous Monitoring tables
-- ============================================================

DROP TABLE IF EXISTS compliance_drift_events CASCADE;
DROP TABLE IF EXISTS compliance_monitor_results CASCADE;
DROP TABLE IF EXISTS compliance_monitors CASCADE;
DROP TABLE IF EXISTS evidence_collection_runs CASCADE;
DROP TABLE IF EXISTS evidence_collection_configs CASCADE;

DROP TYPE IF EXISTS drift_severity;
DROP TYPE IF EXISTS drift_type;
DROP TYPE IF EXISTS monitor_result_status;
DROP TYPE IF EXISTS monitor_check_status;
DROP TYPE IF EXISTS compliance_monitor_type;
DROP TYPE IF EXISTS collection_run_status;
DROP TYPE IF EXISTS evidence_collection_method;
