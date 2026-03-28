-- ============================================================
-- ComplianceForge GRC Platform
-- Migration 002: Organizations, Users, RBAC & Audit Log
-- ============================================================

-- ============================================================
-- ORGANIZATIONS
-- ============================================================

CREATE TABLE organizations (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name                    VARCHAR(255) NOT NULL,
    slug                    VARCHAR(100) UNIQUE NOT NULL,
    legal_name              VARCHAR(500),
    registration_number     VARCHAR(100),
    tax_id                  VARCHAR(100),
    industry                VARCHAR(100) NOT NULL,
    sector                  VARCHAR(100),
    country_code            CHAR(2) NOT NULL,
    headquarters_address    JSONB DEFAULT '{}',
    status                  org_status DEFAULT 'trial',
    tier                    org_tier DEFAULT 'starter',
    settings                JSONB DEFAULT '{}',
    branding                JSONB DEFAULT '{}',
    timezone                VARCHAR(50) DEFAULT 'Europe/London',
    default_language        VARCHAR(10) DEFAULT 'en',
    supported_languages     TEXT[] DEFAULT '{en}',
    employee_count_range    VARCHAR(20),
    annual_revenue_range    VARCHAR(30),
    parent_organization_id  UUID REFERENCES organizations(id),
    metadata                JSONB DEFAULT '{}',
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at              TIMESTAMPTZ
);

CREATE INDEX idx_organizations_slug ON organizations(slug);
CREATE INDEX idx_organizations_status ON organizations(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_organizations_country ON organizations(country_code);
CREATE INDEX idx_organizations_parent ON organizations(parent_organization_id) WHERE parent_organization_id IS NOT NULL;

CREATE TRIGGER trg_organizations_updated_at
    BEFORE UPDATE ON organizations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================
-- ORGANIZATION SUBSCRIPTIONS
-- ============================================================

CREATE TABLE organization_subscriptions (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id         UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    plan_name               VARCHAR(100) NOT NULL,
    status                  VARCHAR(20) NOT NULL DEFAULT 'trialing',
    max_users               INT DEFAULT 5,
    max_frameworks          INT DEFAULT 3,
    features_enabled        JSONB DEFAULT '{}',
    billing_cycle           VARCHAR(20) DEFAULT 'monthly',
    current_period_start    TIMESTAMPTZ,
    current_period_end      TIMESTAMPTZ,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_org_subscriptions_org ON organization_subscriptions(organization_id);

-- ============================================================
-- USERS
-- ============================================================

CREATE TABLE users (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id             UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    email                       VARCHAR(320) NOT NULL,
    password_hash               VARCHAR(255),
    first_name                  VARCHAR(100) NOT NULL,
    last_name                   VARCHAR(100) NOT NULL,
    job_title                   VARCHAR(200),
    department                  VARCHAR(200),
    phone                       VARCHAR(50),
    avatar_url                  TEXT,
    status                      user_status DEFAULT 'pending_verification',
    is_super_admin              BOOLEAN DEFAULT false,
    timezone                    VARCHAR(50) DEFAULT 'Europe/London',
    language                    VARCHAR(10) DEFAULT 'en',
    last_login_at               TIMESTAMPTZ,
    last_login_ip               INET,
    password_changed_at         TIMESTAMPTZ,
    failed_login_attempts       INT DEFAULT 0,
    locked_until                TIMESTAMPTZ,
    notification_preferences    JSONB DEFAULT '{}',
    metadata                    JSONB DEFAULT '{}',
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at                  TIMESTAMPTZ,
    UNIQUE(organization_id, email)
);

CREATE INDEX idx_users_org ON users(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_status ON users(organization_id, status) WHERE deleted_at IS NULL;

CREATE TRIGGER trg_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================
-- USER MFA
-- ============================================================

CREATE TABLE user_mfa (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    method              mfa_method NOT NULL,
    secret_encrypted    BYTEA,
    is_primary          BOOLEAN DEFAULT false,
    is_verified         BOOLEAN DEFAULT false,
    recovery_codes_hash TEXT[],
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_mfa_user ON user_mfa(user_id);

-- ============================================================
-- USER SESSIONS
-- ============================================================

CREATE TABLE user_sessions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    organization_id     UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    token_hash          VARCHAR(255) NOT NULL,
    refresh_token_hash  VARCHAR(255),
    ip_address          INET,
    user_agent          TEXT,
    expires_at          TIMESTAMPTZ NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sessions_user ON user_sessions(user_id);
CREATE INDEX idx_sessions_token ON user_sessions(token_hash);
CREATE INDEX idx_sessions_expires ON user_sessions(expires_at);

-- ============================================================
-- PASSWORD RESET TOKENS
-- ============================================================

CREATE TABLE password_reset_tokens (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash      VARCHAR(255) NOT NULL,
    expires_at      TIMESTAMPTZ NOT NULL,
    used_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_password_reset_user ON password_reset_tokens(user_id);
CREATE INDEX idx_password_reset_token ON password_reset_tokens(token_hash);

-- ============================================================
-- ROLES
-- ============================================================

CREATE TABLE roles (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id     UUID REFERENCES organizations(id) ON DELETE CASCADE,
    name                VARCHAR(100) NOT NULL,
    slug                VARCHAR(100) NOT NULL,
    description         TEXT,
    is_system_role      BOOLEAN DEFAULT false,
    is_custom           BOOLEAN DEFAULT false,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_roles_unique_slug ON roles(COALESCE(organization_id, '00000000-0000-0000-0000-000000000000'::UUID), slug)
    WHERE deleted_at IS NULL;

-- ============================================================
-- PERMISSIONS
-- ============================================================

CREATE TABLE permissions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    resource    VARCHAR(100) NOT NULL,
    action      VARCHAR(30) NOT NULL,
    description TEXT,
    UNIQUE(resource, action)
);

-- ============================================================
-- ROLE ↔ PERMISSION MAPPING
-- ============================================================

CREATE TABLE role_permissions (
    role_id         UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id   UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

-- ============================================================
-- USER ↔ ROLE MAPPING
-- ============================================================

CREATE TABLE user_roles (
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id         UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    assigned_by     UUID REFERENCES users(id),
    assigned_at     TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (user_id, role_id, organization_id)
);

CREATE INDEX idx_user_roles_user ON user_roles(user_id);
CREATE INDEX idx_user_roles_org ON user_roles(organization_id);

-- ============================================================
-- ENTITY-LEVEL PERMISSIONS (Granular)
-- ============================================================

CREATE TABLE user_entity_permissions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    organization_id     UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    entity_type         VARCHAR(50) NOT NULL,
    entity_id           UUID NOT NULL,
    permission_level    VARCHAR(20) NOT NULL,
    granted_by          UUID REFERENCES users(id),
    expires_at          TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_entity_perms_user ON user_entity_permissions(user_id, entity_type, entity_id);
CREATE INDEX idx_entity_perms_org ON user_entity_permissions(organization_id);

-- ============================================================
-- AUDIT LOG (Partitioned by month)
-- ============================================================

CREATE TABLE audit_logs (
    id                  UUID NOT NULL DEFAULT gen_random_uuid(),
    organization_id     UUID NOT NULL,
    user_id             UUID,
    action              VARCHAR(50) NOT NULL,
    entity_type         VARCHAR(100),
    entity_id           UUID,
    changes             JSONB DEFAULT '{}',
    ip_address          INET,
    user_agent          TEXT,
    metadata            JSONB DEFAULT '{}',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
) PARTITION BY RANGE (created_at);

-- Create partitions for the next 12 months
CREATE TABLE audit_logs_2026_01 PARTITION OF audit_logs FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');
CREATE TABLE audit_logs_2026_02 PARTITION OF audit_logs FOR VALUES FROM ('2026-02-01') TO ('2026-03-01');
CREATE TABLE audit_logs_2026_03 PARTITION OF audit_logs FOR VALUES FROM ('2026-03-01') TO ('2026-04-01');
CREATE TABLE audit_logs_2026_04 PARTITION OF audit_logs FOR VALUES FROM ('2026-04-01') TO ('2026-05-01');
CREATE TABLE audit_logs_2026_05 PARTITION OF audit_logs FOR VALUES FROM ('2026-05-01') TO ('2026-06-01');
CREATE TABLE audit_logs_2026_06 PARTITION OF audit_logs FOR VALUES FROM ('2026-06-01') TO ('2026-07-01');
CREATE TABLE audit_logs_2026_07 PARTITION OF audit_logs FOR VALUES FROM ('2026-07-01') TO ('2026-08-01');
CREATE TABLE audit_logs_2026_08 PARTITION OF audit_logs FOR VALUES FROM ('2026-08-01') TO ('2026-09-01');
CREATE TABLE audit_logs_2026_09 PARTITION OF audit_logs FOR VALUES FROM ('2026-09-01') TO ('2026-10-01');
CREATE TABLE audit_logs_2026_10 PARTITION OF audit_logs FOR VALUES FROM ('2026-10-01') TO ('2026-11-01');
CREATE TABLE audit_logs_2026_11 PARTITION OF audit_logs FOR VALUES FROM ('2026-11-01') TO ('2026-12-01');
CREATE TABLE audit_logs_2026_12 PARTITION OF audit_logs FOR VALUES FROM ('2026-12-01') TO ('2027-01-01');

CREATE INDEX idx_audit_logs_org ON audit_logs(organization_id, created_at DESC);
CREATE INDEX idx_audit_logs_user ON audit_logs(user_id, created_at DESC);
CREATE INDEX idx_audit_logs_entity ON audit_logs(entity_type, entity_id);

-- ============================================================
-- ROW-LEVEL SECURITY POLICIES
-- ============================================================

-- Users table RLS
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
CREATE POLICY users_tenant_isolation ON users
    USING (organization_id = get_current_tenant() OR get_current_tenant() IS NULL);

-- User sessions RLS
ALTER TABLE user_sessions ENABLE ROW LEVEL SECURITY;
CREATE POLICY sessions_tenant_isolation ON user_sessions
    USING (organization_id = get_current_tenant() OR get_current_tenant() IS NULL);

-- User roles RLS
ALTER TABLE user_roles ENABLE ROW LEVEL SECURITY;
CREATE POLICY user_roles_tenant_isolation ON user_roles
    USING (organization_id = get_current_tenant() OR get_current_tenant() IS NULL);

-- Entity permissions RLS
ALTER TABLE user_entity_permissions ENABLE ROW LEVEL SECURITY;
CREATE POLICY entity_perms_tenant_isolation ON user_entity_permissions
    USING (organization_id = get_current_tenant() OR get_current_tenant() IS NULL);

-- Audit logs RLS
ALTER TABLE audit_logs ENABLE ROW LEVEL SECURITY;
CREATE POLICY audit_logs_tenant_isolation ON audit_logs
    USING (organization_id = get_current_tenant() OR get_current_tenant() IS NULL);
