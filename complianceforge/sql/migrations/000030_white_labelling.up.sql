-- 000030_white_labelling.up.sql
-- Multi-Tenant White-Labelling, Custom Branding & Theming Engine

-- ============================================================
-- ENUMS
-- ============================================================

CREATE TYPE sidebar_style AS ENUM (
    'light',
    'dark',
    'branded'
);

CREATE TYPE corner_radius_style AS ENUM (
    'none',
    'small',
    'medium',
    'large',
    'full'
);

CREATE TYPE ui_density AS ENUM (
    'compact',
    'comfortable',
    'spacious'
);

CREATE TYPE domain_verification_status AS ENUM (
    'pending',
    'dns_configured',
    'verified',
    'failed'
);

CREATE TYPE ssl_status AS ENUM (
    'pending',
    'provisioning',
    'active',
    'expired',
    'error'
);

-- ============================================================
-- TABLE: tenant_branding
-- Stores per-organisation branding & theming configuration.
-- One row per organisation (UNIQUE on organization_id).
-- ============================================================

CREATE TABLE tenant_branding (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id             UUID NOT NULL UNIQUE REFERENCES organizations(id) ON DELETE CASCADE,

    -- Identity
    product_name                VARCHAR(100) NOT NULL DEFAULT 'ComplianceForge',
    tagline                     VARCHAR(300),
    company_name                VARCHAR(200),
    support_email               VARCHAR(300),
    support_url                 VARCHAR(500),
    privacy_policy_url          VARCHAR(500),
    terms_of_service_url        VARCHAR(500),

    -- Logos (6 variants)
    logo_full_url               TEXT,
    logo_icon_url               TEXT,
    logo_dark_url               TEXT,
    logo_light_url              TEXT,
    favicon_url                 TEXT,
    email_logo_url              TEXT,

    -- Colours (16 fields, all VARCHAR(7) hex e.g. #1A2B3C)
    color_primary               VARCHAR(7) NOT NULL DEFAULT '#4F46E5',
    color_primary_hover         VARCHAR(7) NOT NULL DEFAULT '#4338CA',
    color_secondary             VARCHAR(7) NOT NULL DEFAULT '#7C3AED',
    color_secondary_hover       VARCHAR(7) NOT NULL DEFAULT '#6D28D9',
    color_accent                VARCHAR(7) NOT NULL DEFAULT '#06B6D4',
    color_background            VARCHAR(7) NOT NULL DEFAULT '#F9FAFB',
    color_surface               VARCHAR(7) NOT NULL DEFAULT '#FFFFFF',
    color_text_primary          VARCHAR(7) NOT NULL DEFAULT '#111827',
    color_text_secondary        VARCHAR(7) NOT NULL DEFAULT '#6B7280',
    color_border                VARCHAR(7) NOT NULL DEFAULT '#E5E7EB',
    color_success               VARCHAR(7) NOT NULL DEFAULT '#10B981',
    color_warning               VARCHAR(7) NOT NULL DEFAULT '#F59E0B',
    color_error                 VARCHAR(7) NOT NULL DEFAULT '#EF4444',
    color_info                  VARCHAR(7) NOT NULL DEFAULT '#3B82F6',
    color_sidebar_bg            VARCHAR(7) NOT NULL DEFAULT '#1F2937',
    color_sidebar_text          VARCHAR(7) NOT NULL DEFAULT '#F9FAFB',

    -- Typography
    font_family_heading         VARCHAR(100) NOT NULL DEFAULT 'Inter',
    font_family_body            VARCHAR(100) NOT NULL DEFAULT 'Inter',
    font_size_base              VARCHAR(10) NOT NULL DEFAULT '14px',

    -- Layout & UI Style
    sidebar_style               sidebar_style NOT NULL DEFAULT 'dark',
    corner_radius               corner_radius_style NOT NULL DEFAULT 'medium',
    density                     ui_density NOT NULL DEFAULT 'comfortable',

    -- Custom Domain
    custom_domain               VARCHAR(253),
    domain_verification_token   VARCHAR(128),
    domain_verification_status  domain_verification_status NOT NULL DEFAULT 'pending',
    domain_verified_at          TIMESTAMPTZ,
    ssl_status                  ssl_status NOT NULL DEFAULT 'pending',
    ssl_provisioned_at          TIMESTAMPTZ,
    ssl_expires_at              TIMESTAMPTZ,

    -- Custom CSS (sanitized on write)
    custom_css                  TEXT,

    -- Feature Flags
    show_powered_by             BOOLEAN NOT NULL DEFAULT true,
    show_help_widget            BOOLEAN NOT NULL DEFAULT true,
    show_marketplace            BOOLEAN NOT NULL DEFAULT true,
    show_knowledge_base         BOOLEAN NOT NULL DEFAULT true,

    -- Timestamps
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- TABLE: white_label_partners
-- Partners who resell ComplianceForge under their own brand.
-- ============================================================

CREATE TABLE white_label_partners (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    partner_name                VARCHAR(200) NOT NULL,
    partner_slug                VARCHAR(100) NOT NULL UNIQUE,
    contact_email               VARCHAR(300) NOT NULL,
    default_branding_id         UUID REFERENCES tenant_branding(id) ON DELETE SET NULL,
    revenue_share_percent       NUMERIC(5,2) NOT NULL DEFAULT 0.00
        CONSTRAINT chk_revenue_share CHECK (revenue_share_percent >= 0 AND revenue_share_percent <= 100),
    max_tenants                 INT NOT NULL DEFAULT 100,
    is_active                   BOOLEAN NOT NULL DEFAULT true,
    metadata                    JSONB NOT NULL DEFAULT '{}',
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- TABLE: partner_tenant_mappings
-- Links a white-label partner to the organisations they manage.
-- ============================================================

CREATE TABLE partner_tenant_mappings (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    partner_id                  UUID NOT NULL REFERENCES white_label_partners(id) ON DELETE CASCADE,
    organization_id             UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    onboarded_at                TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_partner_org UNIQUE (partner_id, organization_id)
);

-- ============================================================
-- ROW-LEVEL SECURITY
-- ============================================================

ALTER TABLE tenant_branding ENABLE ROW LEVEL SECURITY;
ALTER TABLE partner_tenant_mappings ENABLE ROW LEVEL SECURITY;

-- tenant_branding: only the owning organisation can see its branding
CREATE POLICY tenant_branding_org_isolation ON tenant_branding
    USING (organization_id = current_setting('app.current_org', true)::uuid);

-- partner_tenant_mappings: visible if you are the mapped org
CREATE POLICY partner_tenant_mappings_org_isolation ON partner_tenant_mappings
    USING (organization_id = current_setting('app.current_org', true)::uuid);

-- white_label_partners is a global table managed by super admins,
-- no RLS needed (access controlled at handler level).

-- ============================================================
-- AUTO-UPDATE updated_at TRIGGERS
-- ============================================================

CREATE TRIGGER set_tenant_branding_updated_at
    BEFORE UPDATE ON tenant_branding
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER set_white_label_partners_updated_at
    BEFORE UPDATE ON white_label_partners
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================
-- INDEXES
-- ============================================================

CREATE INDEX idx_tenant_branding_org ON tenant_branding(organization_id);
CREATE INDEX idx_tenant_branding_custom_domain ON tenant_branding(custom_domain)
    WHERE custom_domain IS NOT NULL;
CREATE INDEX idx_tenant_branding_domain_verified ON tenant_branding(custom_domain)
    WHERE domain_verification_status = 'verified';

CREATE INDEX idx_white_label_partners_slug ON white_label_partners(partner_slug);
CREATE INDEX idx_white_label_partners_active ON white_label_partners(is_active)
    WHERE is_active = true;

CREATE INDEX idx_partner_tenant_mappings_partner ON partner_tenant_mappings(partner_id);
CREATE INDEX idx_partner_tenant_mappings_org ON partner_tenant_mappings(organization_id);
CREATE INDEX idx_partner_tenant_mappings_onboarded ON partner_tenant_mappings(partner_id, onboarded_at DESC);
