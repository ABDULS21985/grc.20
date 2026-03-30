-- 032_webhook_events.sql
-- Seed data: Webhook event type documentation and API scope reference.
-- These are informational inserts used by the developer portal to display
-- available events and scopes. They are also defined in-code in the
-- DeveloperPortalService for runtime access without database round-trips.

BEGIN;

-- ============================================================
-- WEBHOOK EVENT TYPES — Reference Documentation
-- ============================================================

-- We create a lightweight reference table for event type documentation.
-- This supplements the in-code list and allows UI-driven filtering.

CREATE TABLE IF NOT EXISTS webhook_event_type_docs (
    event_type   VARCHAR(100) PRIMARY KEY,
    category     VARCHAR(50) NOT NULL,
    description  TEXT NOT NULL,
    version      VARCHAR(10) NOT NULL DEFAULT '2024-01-01',
    payload_schema JSONB NOT NULL DEFAULT '{}',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Risk Management Events
INSERT INTO webhook_event_type_docs (event_type, category, description, version, payload_schema) VALUES
('risk.created',            'Risk Management',     'A new risk has been created',                       '2024-01-01', '{"type":"object","properties":{"id":{"type":"string"},"title":{"type":"string"},"severity":{"type":"string"}}}'),
('risk.updated',            'Risk Management',     'A risk has been updated',                           '2024-01-01', '{"type":"object","properties":{"id":{"type":"string"},"changes":{"type":"object"}}}'),
('risk.deleted',            'Risk Management',     'A risk has been deleted',                           '2024-01-01', '{"type":"object","properties":{"id":{"type":"string"}}}'),
('risk.status_changed',     'Risk Management',     'A risk status has changed',                         '2024-01-01', '{"type":"object","properties":{"id":{"type":"string"},"old_status":{"type":"string"},"new_status":{"type":"string"}}}'),
('risk.review_due',         'Risk Management',     'A risk review is due',                              '2024-01-01', '{"type":"object","properties":{"id":{"type":"string"},"due_date":{"type":"string"}}}'),
('risk.score_changed',      'Risk Management',     'A risk score has changed',                          '2024-01-01', '{"type":"object","properties":{"id":{"type":"string"},"old_score":{"type":"number"},"new_score":{"type":"number"}}}'),

-- Control Management Events
('control.created',            'Control Management',  'A new control has been created',                 '2024-01-01', '{"type":"object","properties":{"id":{"type":"string"},"title":{"type":"string"}}}'),
('control.updated',            'Control Management',  'A control has been updated',                     '2024-01-01', '{"type":"object","properties":{"id":{"type":"string"},"changes":{"type":"object"}}}'),
('control.status_changed',     'Control Management',  'A control implementation status has changed',    '2024-01-01', '{"type":"object","properties":{"id":{"type":"string"},"old_status":{"type":"string"},"new_status":{"type":"string"}}}'),
('control.evidence_attached',  'Control Management',  'Evidence has been attached to a control',        '2024-01-01', '{"type":"object","properties":{"control_id":{"type":"string"},"evidence_id":{"type":"string"}}}'),
('control.test_completed',     'Control Management',  'A control test has been completed',              '2024-01-01', '{"type":"object","properties":{"control_id":{"type":"string"},"result":{"type":"string"}}}'),

-- Policy Management Events
('policy.created',       'Policy Management',  'A new policy has been created',             '2024-01-01', '{"type":"object","properties":{"id":{"type":"string"},"title":{"type":"string"}}}'),
('policy.updated',       'Policy Management',  'A policy has been updated',                 '2024-01-01', '{"type":"object","properties":{"id":{"type":"string"},"version":{"type":"string"}}}'),
('policy.published',     'Policy Management',  'A policy has been published',               '2024-01-01', '{"type":"object","properties":{"id":{"type":"string"},"version":{"type":"string"}}}'),
('policy.approved',      'Policy Management',  'A policy has been approved',                '2024-01-01', '{"type":"object","properties":{"id":{"type":"string"},"approved_by":{"type":"string"}}}'),
('policy.review_due',    'Policy Management',  'A policy review is coming due',             '2024-01-01', '{"type":"object","properties":{"id":{"type":"string"},"due_date":{"type":"string"}}}'),
('policy.acknowledged',  'Policy Management',  'A policy has been acknowledged by a user',  '2024-01-01', '{"type":"object","properties":{"policy_id":{"type":"string"},"user_id":{"type":"string"}}}'),

-- Audit Management Events
('audit.created',          'Audit Management',   'A new audit has been created',          '2024-01-01', '{"type":"object","properties":{"id":{"type":"string"},"title":{"type":"string"}}}'),
('audit.completed',        'Audit Management',   'An audit has been completed',           '2024-01-01', '{"type":"object","properties":{"id":{"type":"string"},"findings_count":{"type":"integer"}}}'),
('audit.finding_created',  'Audit Management',   'A new audit finding has been created',  '2024-01-01', '{"type":"object","properties":{"id":{"type":"string"},"audit_id":{"type":"string"},"severity":{"type":"string"}}}'),
('audit.finding_resolved', 'Audit Management',   'An audit finding has been resolved',    '2024-01-01', '{"type":"object","properties":{"id":{"type":"string"},"audit_id":{"type":"string"}}}'),

-- Incident Management Events
('incident.created',         'Incident Management', 'A new incident has been reported',     '2024-01-01', '{"type":"object","properties":{"id":{"type":"string"},"severity":{"type":"string"}}}'),
('incident.updated',         'Incident Management', 'An incident has been updated',         '2024-01-01', '{"type":"object","properties":{"id":{"type":"string"},"changes":{"type":"object"}}}'),
('incident.resolved',        'Incident Management', 'An incident has been resolved',        '2024-01-01', '{"type":"object","properties":{"id":{"type":"string"},"resolution":{"type":"string"}}}'),
('incident.escalated',       'Incident Management', 'An incident has been escalated',       '2024-01-01', '{"type":"object","properties":{"id":{"type":"string"},"escalated_to":{"type":"string"}}}'),
('incident.breach_detected', 'Incident Management', 'A data breach has been detected',      '2024-01-01', '{"type":"object","properties":{"id":{"type":"string"},"breach_type":{"type":"string"},"affected_records":{"type":"integer"}}}'),

-- Compliance Events
('compliance.score_changed',        'Compliance', 'Overall compliance score has changed',        '2024-01-01', '{"type":"object","properties":{"old_score":{"type":"number"},"new_score":{"type":"number"}}}'),
('compliance.framework_mapped',     'Compliance', 'A framework mapping has been updated',        '2024-01-01', '{"type":"object","properties":{"framework_code":{"type":"string"}}}'),
('compliance.gap_identified',       'Compliance', 'A compliance gap has been identified',        '2024-01-01', '{"type":"object","properties":{"gap_id":{"type":"string"},"control_id":{"type":"string"}}}'),
('compliance.deadline_approaching', 'Compliance', 'A compliance deadline is approaching',        '2024-01-01', '{"type":"object","properties":{"deadline":{"type":"string"},"framework":{"type":"string"}}}'),

-- Vendor Management Events
('vendor.created',                'Vendor Management', 'A new vendor has been added',                 '2024-01-01', '{"type":"object","properties":{"id":{"type":"string"},"name":{"type":"string"}}}'),
('vendor.risk_changed',           'Vendor Management', 'Vendor risk level has changed',               '2024-01-01', '{"type":"object","properties":{"id":{"type":"string"},"old_risk":{"type":"string"},"new_risk":{"type":"string"}}}'),
('vendor.assessment_completed',   'Vendor Management', 'A vendor assessment has been completed',      '2024-01-01', '{"type":"object","properties":{"vendor_id":{"type":"string"},"score":{"type":"number"}}}'),
('vendor.contract_expiring',      'Vendor Management', 'A vendor contract is expiring',               '2024-01-01', '{"type":"object","properties":{"vendor_id":{"type":"string"},"expiry_date":{"type":"string"}}}'),

-- Evidence Events
('evidence.uploaded',        'Evidence', 'New evidence has been uploaded',       '2024-01-01', '{"type":"object","properties":{"id":{"type":"string"},"control_id":{"type":"string"}}}'),
('evidence.expired',         'Evidence', 'Evidence has expired',                 '2024-01-01', '{"type":"object","properties":{"id":{"type":"string"},"expired_at":{"type":"string"}}}'),
('evidence.review_required', 'Evidence', 'Evidence requires review',             '2024-01-01', '{"type":"object","properties":{"id":{"type":"string"},"due_date":{"type":"string"}}}'),

-- Regulatory Events
('regulatory.change_detected',      'Regulatory', 'A regulatory change has been detected',               '2024-01-01', '{"type":"object","properties":{"change_id":{"type":"string"},"jurisdiction":{"type":"string"}}}'),
('regulatory.impact_assessed',      'Regulatory', 'Impact of a regulatory change has been assessed',     '2024-01-01', '{"type":"object","properties":{"change_id":{"type":"string"},"impact_level":{"type":"string"}}}'),
('regulatory.deadline_approaching', 'Regulatory', 'A regulatory deadline is approaching',                '2024-01-01', '{"type":"object","properties":{"change_id":{"type":"string"},"deadline":{"type":"string"}}}'),

-- User Management Events
('user.created',      'User Management', 'A new user has been created',      '2024-01-01', '{"type":"object","properties":{"id":{"type":"string"},"email":{"type":"string"}}}'),
('user.deactivated',  'User Management', 'A user has been deactivated',      '2024-01-01', '{"type":"object","properties":{"id":{"type":"string"}}}'),
('user.role_changed', 'User Management', 'A user role has been changed',     '2024-01-01', '{"type":"object","properties":{"id":{"type":"string"},"old_role":{"type":"string"},"new_role":{"type":"string"}}}'),

-- Workflow Events
('workflow.started',            'Workflow', 'A workflow has been started',              '2024-01-01', '{"type":"object","properties":{"id":{"type":"string"},"type":{"type":"string"}}}'),
('workflow.completed',          'Workflow', 'A workflow has been completed',            '2024-01-01', '{"type":"object","properties":{"id":{"type":"string"},"outcome":{"type":"string"}}}'),
('workflow.step_completed',     'Workflow', 'A workflow step has been completed',       '2024-01-01', '{"type":"object","properties":{"workflow_id":{"type":"string"},"step_id":{"type":"string"}}}'),
('workflow.approval_required',  'Workflow', 'A workflow step requires approval',        '2024-01-01', '{"type":"object","properties":{"workflow_id":{"type":"string"},"approver_id":{"type":"string"}}}'),

-- System Events
('ping',             'System', 'Connectivity test event',          '2024-01-01', '{"type":"object","properties":{"message":{"type":"string"}}}'),
('api_key.created',  'System', 'A new API key has been created',   '2024-01-01', '{"type":"object","properties":{"key_id":{"type":"string"},"prefix":{"type":"string"}}}'),
('api_key.revoked',  'System', 'An API key has been revoked',      '2024-01-01', '{"type":"object","properties":{"key_id":{"type":"string"}}}')

ON CONFLICT (event_type) DO NOTHING;

-- ============================================================
-- API SCOPES — Reference Documentation
-- ============================================================

CREATE TABLE IF NOT EXISTS api_scope_docs (
    scope       VARCHAR(100) PRIMARY KEY,
    category    VARCHAR(50) NOT NULL,
    description TEXT NOT NULL,
    access      VARCHAR(10) NOT NULL CHECK (access IN ('read', 'write', 'delete')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO api_scope_docs (scope, category, description, access) VALUES
-- Risk
('risks:read',        'Risk Management',       'Read risk records',                     'read'),
('risks:write',       'Risk Management',       'Create and update risk records',        'write'),
('risks:delete',      'Risk Management',       'Delete risk records',                   'delete'),
-- Controls
('controls:read',     'Control Management',    'Read control records',                  'read'),
('controls:write',    'Control Management',    'Create and update control records',     'write'),
('controls:delete',   'Control Management',    'Delete control records',                'delete'),
-- Policies
('policies:read',     'Policy Management',     'Read policy documents',                 'read'),
('policies:write',    'Policy Management',     'Create and update policy documents',    'write'),
('policies:delete',   'Policy Management',     'Delete policy documents',               'delete'),
-- Audits
('audits:read',       'Audit Management',      'Read audit records and findings',       'read'),
('audits:write',      'Audit Management',      'Create and update audits',              'write'),
('audits:delete',     'Audit Management',      'Delete audit records',                  'delete'),
-- Incidents
('incidents:read',    'Incident Management',   'Read incident reports',                 'read'),
('incidents:write',   'Incident Management',   'Create and update incidents',           'write'),
('incidents:delete',  'Incident Management',   'Delete incident reports',               'delete'),
-- Vendors
('vendors:read',      'Vendor Management',     'Read vendor information',               'read'),
('vendors:write',     'Vendor Management',     'Create and update vendor records',      'write'),
('vendors:delete',    'Vendor Management',     'Delete vendor records',                 'delete'),
-- Evidence
('evidence:read',     'Evidence',              'Read evidence records',                 'read'),
('evidence:write',    'Evidence',              'Upload and update evidence',            'write'),
('evidence:delete',   'Evidence',              'Delete evidence records',               'delete'),
-- Compliance
('compliance:read',   'Compliance',            'Read compliance data and scores',       'read'),
('compliance:write',  'Compliance',            'Update compliance mappings',            'write'),
-- Frameworks
('frameworks:read',   'Frameworks',            'Read framework definitions',            'read'),
('frameworks:write',  'Frameworks',            'Create custom frameworks',              'write'),
-- Users
('users:read',        'User Management',       'Read user information',                 'read'),
('users:write',       'User Management',       'Create and update users',               'write'),
('users:delete',      'User Management',       'Deactivate users',                      'delete'),
-- Reports
('reports:read',      'Reports',               'Generate and read reports',             'read'),
('reports:write',     'Reports',               'Create report templates',               'write'),
-- Webhooks
('webhooks:read',     'Webhooks',              'Read webhook subscriptions',            'read'),
('webhooks:write',    'Webhooks',              'Create and update webhooks',            'write'),
('webhooks:delete',   'Webhooks',              'Delete webhook subscriptions',          'delete'),
-- Analytics
('analytics:read',    'Analytics',             'Read analytics and dashboards',         'read'),
-- Settings
('settings:read',     'Settings',              'Read organisation settings',            'read'),
('settings:write',    'Settings',              'Update organisation settings',          'write')

ON CONFLICT (scope) DO NOTHING;

COMMIT;
