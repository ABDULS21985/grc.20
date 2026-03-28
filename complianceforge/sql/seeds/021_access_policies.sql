-- 021_access_policies.sql
-- Default ABAC policies for the ComplianceForge GRC platform.
-- These are seeded per-organisation during onboarding.
--
-- Each policy uses JSONB conditions with the structure:
--   [{"field": "<attr>", "operator": "<op>", "value": <json_value>}]
--
-- Supported operators: equals, not_equals, in, not_in, contains_any,
--   greater_than, less_than, in_cidr, between, equals_subject

-- ============================================================
-- Policy (a): Org Admin — Full Access
-- Grants allow on all resource types and all actions to org_admin.
-- ============================================================
INSERT INTO access_policies (id, org_id, name, description, priority, effect, is_active,
    subject_conditions, resource_type, resource_conditions, actions, environment_conditions,
    created_at, updated_at)
SELECT
    '00000000-0000-0000-0000-000000000a01'::uuid,
    o.id,
    'Org Admin Full Access',
    'Grants organisation administrators unrestricted access to all resources and actions.',
    1,
    'allow',
    true,
    '[{"field":"roles","operator":"contains_any","value":["org_admin"]}]'::jsonb,
    '*',
    '[]'::jsonb,
    '{read,create,update,delete,export,approve,assign}',
    '[]'::jsonb,
    now(), now()
FROM organizations o
LIMIT 1;

-- Assignment: role = org_admin
INSERT INTO access_policy_assignments (id, org_id, access_policy_id, assignee_type, created_at)
SELECT
    '00000000-0000-0000-0000-000000000b01'::uuid,
    ap.org_id,
    ap.id,
    'role',
    now()
FROM access_policies ap WHERE ap.id = '00000000-0000-0000-0000-000000000a01'::uuid;

-- ============================================================
-- Policy (b): Control Owner — Own Controls Only
-- Allows control owners to manage controls where they are the owner.
-- ============================================================
INSERT INTO access_policies (id, org_id, name, description, priority, effect, is_active,
    subject_conditions, resource_type, resource_conditions, actions, environment_conditions,
    created_at, updated_at)
SELECT
    '00000000-0000-0000-0000-000000000a02'::uuid,
    o.id,
    'Control Owner — Own Controls Only',
    'Allows control owners to read and update controls they own. Uses equals_subject to match owner_user_id against the requesting user.',
    50,
    'allow',
    true,
    '[{"field":"roles","operator":"contains_any","value":["control_owner","risk_manager"]}]'::jsonb,
    'control',
    '[{"field":"owner_user_id","operator":"equals_subject","value":"user_id"}]'::jsonb,
    '{read,update,assign}',
    '[]'::jsonb,
    now(), now()
FROM organizations o
LIMIT 1;

INSERT INTO access_policy_assignments (id, org_id, access_policy_id, assignee_type, created_at)
SELECT
    '00000000-0000-0000-0000-000000000b02'::uuid,
    ap.org_id,
    ap.id,
    'role',
    now()
FROM access_policies ap WHERE ap.id = '00000000-0000-0000-0000-000000000a02'::uuid;

-- ============================================================
-- Policy (c): DPO — Privacy Incidents
-- Grants the Data Protection Officer access to incidents flagged as data breaches.
-- ============================================================
INSERT INTO access_policies (id, org_id, name, description, priority, effect, is_active,
    subject_conditions, resource_type, resource_conditions, actions, environment_conditions,
    created_at, updated_at)
SELECT
    '00000000-0000-0000-0000-000000000a03'::uuid,
    o.id,
    'DPO — Privacy Incidents Only',
    'Grants the DPO read/update access to incidents where is_data_breach is true.',
    30,
    'allow',
    true,
    '[{"field":"roles","operator":"contains_any","value":["dpo","privacy_officer"]}]'::jsonb,
    'incident',
    '[{"field":"is_data_breach","operator":"equals","value":true}]'::jsonb,
    '{read,update,approve,export}',
    '[]'::jsonb,
    now(), now()
FROM organizations o
LIMIT 1;

INSERT INTO access_policy_assignments (id, org_id, access_policy_id, assignee_type, created_at)
SELECT
    '00000000-0000-0000-0000-000000000b03'::uuid,
    ap.org_id,
    ap.id,
    'role',
    now()
FROM access_policies ap WHERE ap.id = '00000000-0000-0000-0000-000000000a03'::uuid;

-- ============================================================
-- Policy (d): DPO — All DSR Requests
-- Grants full access to all Data Subject Request records.
-- ============================================================
INSERT INTO access_policies (id, org_id, name, description, priority, effect, is_active,
    subject_conditions, resource_type, resource_conditions, actions, environment_conditions,
    created_at, updated_at)
SELECT
    '00000000-0000-0000-0000-000000000a04'::uuid,
    o.id,
    'DPO — All DSR Requests',
    'Grants the DPO full lifecycle access to all Data Subject Requests for GDPR compliance.',
    30,
    'allow',
    true,
    '[{"field":"roles","operator":"contains_any","value":["dpo","privacy_officer"]}]'::jsonb,
    'dsr_request',
    '[]'::jsonb,
    '{read,create,update,delete,approve,export,assign}',
    '[]'::jsonb,
    now(), now()
FROM organizations o
LIMIT 1;

INSERT INTO access_policy_assignments (id, org_id, access_policy_id, assignee_type, created_at)
SELECT
    '00000000-0000-0000-0000-000000000b04'::uuid,
    ap.org_id,
    ap.id,
    'role',
    now()
FROM access_policies ap WHERE ap.id = '00000000-0000-0000-0000-000000000a04'::uuid;

-- ============================================================
-- Policy (e): External Auditor — Read-Only During Engagement
-- Temporal + MFA-required: allows read access only within the
-- engagement window and only when MFA has been verified.
-- ============================================================
INSERT INTO access_policies (id, org_id, name, description, priority, effect, is_active,
    subject_conditions, resource_type, resource_conditions, actions, environment_conditions,
    valid_from, valid_until,
    created_at, updated_at)
SELECT
    '00000000-0000-0000-0000-000000000a05'::uuid,
    o.id,
    'External Auditor — Read-Only During Engagement',
    'Grants external auditors read-only access to all resources. Enforces temporal bounds (engagement period) and requires MFA verification.',
    40,
    'allow',
    true,
    '[{"field":"roles","operator":"contains_any","value":["external_auditor","auditor"]}]'::jsonb,
    '*',
    '[]'::jsonb,
    '{read}',
    '[{"field":"mfa_verified","operator":"equals","value":true}]'::jsonb,
    now(),
    now() + interval '90 days',
    now(), now()
FROM organizations o
LIMIT 1;

INSERT INTO access_policy_assignments (id, org_id, access_policy_id, assignee_type, created_at)
SELECT
    '00000000-0000-0000-0000-000000000b05'::uuid,
    ap.org_id,
    ap.id,
    'role',
    now()
FROM access_policies ap WHERE ap.id = '00000000-0000-0000-0000-000000000a05'::uuid;

-- ============================================================
-- Policy (f): Viewer — Read-Only No Confidential
-- Viewers can read everything except confidential / restricted resources.
-- ============================================================
INSERT INTO access_policies (id, org_id, name, description, priority, effect, is_active,
    subject_conditions, resource_type, resource_conditions, actions, environment_conditions,
    created_at, updated_at)
SELECT
    '00000000-0000-0000-0000-000000000a06'::uuid,
    o.id,
    'Viewer — Read-Only No Confidential',
    'Grants viewers read-only access to all resources except those classified as confidential or restricted.',
    60,
    'allow',
    true,
    '[{"field":"roles","operator":"contains_any","value":["viewer","readonly"]}]'::jsonb,
    '*',
    '[{"field":"classification","operator":"not_in","value":["confidential","restricted"]}]'::jsonb,
    '{read}',
    '[]'::jsonb,
    now(), now()
FROM organizations o
LIMIT 1;

INSERT INTO access_policy_assignments (id, org_id, access_policy_id, assignee_type, created_at)
SELECT
    '00000000-0000-0000-0000-000000000b06'::uuid,
    ap.org_id,
    ap.id,
    'role',
    now()
FROM access_policies ap WHERE ap.id = '00000000-0000-0000-0000-000000000a06'::uuid;

-- ============================================================
-- Policy (g): Export Restriction After Hours (DENY)
-- Denies export actions outside business hours (09:00-18:00 UTC).
-- Uses the between operator on the environment time attribute.
-- ============================================================
INSERT INTO access_policies (id, org_id, name, description, priority, effect, is_active,
    subject_conditions, resource_type, resource_conditions, actions, environment_conditions,
    created_at, updated_at)
SELECT
    '00000000-0000-0000-0000-000000000a07'::uuid,
    o.id,
    'Export Restriction — After Hours',
    'Denies all export actions outside standard business hours (09:00–18:00 UTC). Applies to all users.',
    10,
    'deny',
    true,
    '[]'::jsonb,
    '*',
    '[]'::jsonb,
    '{export}',
    '[{"field":"time_hour","operator":"not_in","value":[9,10,11,12,13,14,15,16,17]}]'::jsonb,
    now(), now()
FROM organizations o
LIMIT 1;

INSERT INTO access_policy_assignments (id, org_id, access_policy_id, assignee_type, created_at)
SELECT
    '00000000-0000-0000-0000-000000000b07'::uuid,
    ap.org_id,
    ap.id,
    'all_users',
    now()
FROM access_policies ap WHERE ap.id = '00000000-0000-0000-0000-000000000a07'::uuid;

-- ============================================================
-- Policy (h): Field Masking — Financial Data for Non-Managers
-- Allows read but masks financial fields for non-manager roles.
-- The field_level_permissions table controls which fields are masked.
-- ============================================================
INSERT INTO access_policies (id, org_id, name, description, priority, effect, is_active,
    subject_conditions, resource_type, resource_conditions, actions, environment_conditions,
    created_at, updated_at)
SELECT
    '00000000-0000-0000-0000-000000000a08'::uuid,
    o.id,
    'Field Masking — Financial Data for Non-Managers',
    'Allows read access to risk and vendor records but masks financial fields (estimated cost, financial impact) for users without manager-level roles.',
    70,
    'allow',
    true,
    '[{"field":"roles","operator":"not_in","value":["org_admin","risk_manager","finance_manager"]}]'::jsonb,
    'risk',
    '[]'::jsonb,
    '{read}',
    '[]'::jsonb,
    now(), now()
FROM organizations o
LIMIT 1;

INSERT INTO access_policy_assignments (id, org_id, access_policy_id, assignee_type, created_at)
SELECT
    '00000000-0000-0000-0000-000000000b08'::uuid,
    ap.org_id,
    ap.id,
    'all_users',
    now()
FROM access_policies ap WHERE ap.id = '00000000-0000-0000-0000-000000000a08'::uuid;

-- Field-level masking rules for the financial data policy
INSERT INTO field_level_permissions (id, org_id, access_policy_id, resource_type, field_name, permission, mask_pattern, created_at)
SELECT
    '00000000-0000-0000-0000-000000000c01'::uuid,
    ap.org_id,
    ap.id,
    'risk',
    'financial_impact_eur',
    'masked',
    '***,***',
    now()
FROM access_policies ap WHERE ap.id = '00000000-0000-0000-0000-000000000a08'::uuid;

INSERT INTO field_level_permissions (id, org_id, access_policy_id, resource_type, field_name, permission, mask_pattern, created_at)
SELECT
    '00000000-0000-0000-0000-000000000c02'::uuid,
    ap.org_id,
    ap.id,
    'risk',
    'estimated_cost',
    'masked',
    '***,***',
    now()
FROM access_policies ap WHERE ap.id = '00000000-0000-0000-0000-000000000a08'::uuid;

INSERT INTO field_level_permissions (id, org_id, access_policy_id, resource_type, field_name, permission, mask_pattern, created_at)
SELECT
    '00000000-0000-0000-0000-000000000c03'::uuid,
    ap.org_id,
    ap.id,
    'vendor',
    'contract_value',
    'hidden',
    NULL,
    now()
FROM access_policies ap WHERE ap.id = '00000000-0000-0000-0000-000000000a08'::uuid;
