-- ============================================================
-- ComplianceForge GRC Platform
-- Migration 008: Advanced Reporting Engine (DOWN)
-- ============================================================

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS report_runs CASCADE;
DROP TABLE IF EXISTS report_schedules CASCADE;
DROP TABLE IF EXISTS report_definitions CASCADE;

-- Drop enum types
DROP TYPE IF EXISTS report_delivery_channel;
DROP TYPE IF EXISTS report_run_status;
DROP TYPE IF EXISTS report_frequency;
DROP TYPE IF EXISTS report_format;
DROP TYPE IF EXISTS report_type;
