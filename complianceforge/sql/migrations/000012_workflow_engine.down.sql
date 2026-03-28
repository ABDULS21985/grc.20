-- ============================================================
-- ComplianceForge GRC Platform
-- Migration 012 DOWN: Drop Compliance Workflow Engine
-- ============================================================

DROP TRIGGER IF EXISTS set_workflow_delegation_rules_updated_at ON workflow_delegation_rules;
DROP TRIGGER IF EXISTS set_workflow_instances_updated_at ON workflow_instances;
DROP TRIGGER IF EXISTS set_workflow_steps_updated_at ON workflow_steps;
DROP TRIGGER IF EXISTS set_workflow_definitions_updated_at ON workflow_definitions;

DROP TABLE IF EXISTS workflow_delegation_rules CASCADE;
DROP TABLE IF EXISTS workflow_step_executions CASCADE;
DROP TABLE IF EXISTS workflow_instances CASCADE;
DROP TABLE IF EXISTS workflow_steps CASCADE;
DROP TABLE IF EXISTS workflow_definitions CASCADE;

DROP TYPE IF EXISTS workflow_task_assignee_type;
DROP TYPE IF EXISTS workflow_step_action;
DROP TYPE IF EXISTS workflow_step_exec_status;
DROP TYPE IF EXISTS workflow_sla_status;
DROP TYPE IF EXISTS workflow_completion_outcome;
DROP TYPE IF EXISTS workflow_instance_status;
DROP TYPE IF EXISTS workflow_approval_mode;
DROP TYPE IF EXISTS workflow_approver_type;
DROP TYPE IF EXISTS workflow_step_type;
DROP TYPE IF EXISTS workflow_status;
DROP TYPE IF EXISTS workflow_type;
