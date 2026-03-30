-- ============================================================
-- ComplianceForge GRC Platform
-- Migration 033 DOWN: Drop Advanced Audit Management tables
-- ============================================================

DROP TABLE IF EXISTS audit_corrective_actions CASCADE;
DROP TABLE IF EXISTS audit_test_procedures CASCADE;
DROP TABLE IF EXISTS audit_samples CASCADE;
DROP TABLE IF EXISTS audit_workpapers CASCADE;
DROP TABLE IF EXISTS audit_engagements CASCADE;
DROP TABLE IF EXISTS audit_universe CASCADE;
DROP TABLE IF EXISTS audit_programmes CASCADE;
