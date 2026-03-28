-- ============================================================
-- ComplianceForge GRC Platform
-- Migration 009 DOWN: Drop GDPR DSR Management Tables
-- ============================================================

DROP VIEW IF EXISTS v_dsr_dashboard;

DROP TRIGGER IF EXISTS trg_dsr_templates_updated_at ON dsr_response_templates;
DROP TRIGGER IF EXISTS trg_dsr_tasks_updated_at ON dsr_tasks;
DROP TRIGGER IF EXISTS trg_dsr_requests_updated_at ON dsr_requests;

DROP POLICY IF EXISTS dsr_templates_tenant ON dsr_response_templates;
DROP POLICY IF EXISTS dsr_audit_trail_tenant ON dsr_audit_trail;
DROP POLICY IF EXISTS dsr_tasks_tenant ON dsr_tasks;
DROP POLICY IF EXISTS dsr_requests_tenant ON dsr_requests;

DROP TABLE IF EXISTS dsr_response_templates;
DROP TABLE IF EXISTS dsr_audit_trail;
DROP TABLE IF EXISTS dsr_tasks;
DROP TABLE IF EXISTS dsr_requests;

DROP FUNCTION IF EXISTS generate_dsr_ref(UUID);

DROP TYPE IF EXISTS dsr_task_status;
DROP TYPE IF EXISTS dsr_task_type;
DROP TYPE IF EXISTS dsr_sla_status;
DROP TYPE IF EXISTS dsr_response_method;
DROP TYPE IF EXISTS dsr_request_source;
DROP TYPE IF EXISTS dsr_priority;
DROP TYPE IF EXISTS dsr_status;
DROP TYPE IF EXISTS dsr_request_type;
