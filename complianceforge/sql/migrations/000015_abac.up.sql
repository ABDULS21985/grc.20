-- 000015_abac.up.sql
-- Advanced Attribute-Based Access Control (ABAC) Engine
-- Provides fine-grained, attribute-driven authorization for the ComplianceForge GRC platform.

BEGIN;

-- ============================================================
-- ENUM TYPES
-- ============================================================

CREATE TYPE access_policy_effect AS ENUM ('allow', 'deny');

CREATE TYPE assignee_type AS ENUM ('user', 'role', 'group', 'all_users');

CREATE TYPE access_decision_type AS ENUM ('allow', 'deny');

CREATE TYPE field_permission_type AS ENUM ('visible', 'masked', 'hidden');

-- ============================================================
-- TABLE: access_policies
-- Core ABAC policy definitions storing subject, resource, and
-- environment conditions as JSONB for maximum flexibility.
-- ============================================================

CREATE TABLE access_policies (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id                  UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name                    VARCHAR(255) NOT NULL,
    description             TEXT,
    priority                INT NOT NULL DEFAULT 100,
    effect                  access_policy_effect NOT NULL,
    is_active               BOOLEAN NOT NULL DEFAULT true,
    subject_conditions      JSONB NOT NULL DEFAULT '[]'::jsonb,
    resource_type           VARCHAR(100) NOT NULL,
    resource_conditions     JSONB NOT NULL DEFAULT '[]'::jsonb,
    actions                 TEXT[] NOT NULL DEFAULT '{}',
    environment_conditions  JSONB NOT NULL DEFAULT '[]'::jsonb,
    valid_from              TIMESTAMPTZ,
    valid_until             TIMESTAMPTZ,
    created_by              UUID REFERENCES users(id),
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ============================================================
-- TABLE: access_policy_assignments
-- Links policies to subjects (users, roles, groups, or all).
-- ============================================================

CREATE TABLE access_policy_assignments (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id              UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    access_policy_id    UUID NOT NULL REFERENCES access_policies(id) ON DELETE CASCADE,
    assignee_type       assignee_type NOT NULL,
    assignee_id         UUID,
    valid_from          TIMESTAMPTZ,
    valid_until         TIMESTAMPTZ,
    created_by          UUID REFERENCES users(id),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ============================================================
-- TABLE: access_audit_log
-- Immutable record of every ABAC evaluation decision.
-- ============================================================

CREATE TABLE access_audit_log (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id                  UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id                 UUID NOT NULL,
    action                  VARCHAR(100) NOT NULL,
    resource_type           VARCHAR(100) NOT NULL,
    resource_id             UUID,
    decision                access_decision_type NOT NULL,
    matched_policy_id       UUID REFERENCES access_policies(id) ON DELETE SET NULL,
    evaluation_time_us      INT NOT NULL DEFAULT 0,
    subject_attributes      JSONB NOT NULL DEFAULT '{}'::jsonb,
    resource_attributes     JSONB NOT NULL DEFAULT '{}'::jsonb,
    environment_attributes  JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ============================================================
-- TABLE: field_level_permissions
-- Per-field visibility/masking rules tied to access policies.
-- ============================================================

CREATE TABLE field_level_permissions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id              UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    access_policy_id    UUID NOT NULL REFERENCES access_policies(id) ON DELETE CASCADE,
    resource_type       VARCHAR(100) NOT NULL,
    field_name          VARCHAR(255) NOT NULL,
    permission          field_permission_type NOT NULL DEFAULT 'visible',
    mask_pattern        VARCHAR(255),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ============================================================
-- ROW-LEVEL SECURITY
-- ============================================================

ALTER TABLE access_policies ENABLE ROW LEVEL SECURITY;
ALTER TABLE access_policy_assignments ENABLE ROW LEVEL SECURITY;
ALTER TABLE access_audit_log ENABLE ROW LEVEL SECURITY;
ALTER TABLE field_level_permissions ENABLE ROW LEVEL SECURITY;

-- access_policies RLS
CREATE POLICY access_policies_tenant ON access_policies
    USING (org_id::text = current_setting('app.current_tenant', true));

-- access_policy_assignments RLS
CREATE POLICY access_policy_assignments_tenant ON access_policy_assignments
    USING (org_id::text = current_setting('app.current_tenant', true));

-- access_audit_log RLS
CREATE POLICY access_audit_log_tenant ON access_audit_log
    USING (org_id::text = current_setting('app.current_tenant', true));

-- field_level_permissions RLS
CREATE POLICY field_level_permissions_tenant ON field_level_permissions
    USING (org_id::text = current_setting('app.current_tenant', true));

-- ============================================================
-- INDEXES
-- ============================================================

-- access_policies: org + active filter (most common query path)
CREATE INDEX idx_access_policies_org_id ON access_policies(org_id);
CREATE INDEX idx_access_policies_resource_type ON access_policies(org_id, resource_type);

-- Partial index: only active policies matter for evaluation
CREATE INDEX idx_access_policies_active ON access_policies(org_id, resource_type, priority)
    WHERE is_active = true;

-- access_policy_assignments: fast lookup by policy, by assignee
CREATE INDEX idx_apa_policy_id ON access_policy_assignments(access_policy_id);
CREATE INDEX idx_apa_assignee ON access_policy_assignments(org_id, assignee_type, assignee_id);

-- access_audit_log: time-series queries and per-user lookups
CREATE INDEX idx_aal_org_created ON access_audit_log(org_id, created_at DESC);
CREATE INDEX idx_aal_user ON access_audit_log(org_id, user_id, created_at DESC);
CREATE INDEX idx_aal_resource ON access_audit_log(org_id, resource_type, resource_id);

-- field_level_permissions
CREATE INDEX idx_flp_policy ON field_level_permissions(access_policy_id);
CREATE INDEX idx_flp_resource ON field_level_permissions(org_id, resource_type);

-- ============================================================
-- TRIGGER: updated_at auto-maintenance
-- ============================================================

CREATE OR REPLACE FUNCTION update_access_policies_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_access_policies_updated_at
    BEFORE UPDATE ON access_policies
    FOR EACH ROW EXECUTE FUNCTION update_access_policies_updated_at();

COMMIT;
