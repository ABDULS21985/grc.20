-- 000030_white_labelling.down.sql
-- Rollback: Multi-Tenant White-Labelling, Custom Branding & Theming Engine

-- ============================================================
-- DROP INDEXES
-- ============================================================

DROP INDEX IF EXISTS idx_partner_tenant_mappings_onboarded;
DROP INDEX IF EXISTS idx_partner_tenant_mappings_org;
DROP INDEX IF EXISTS idx_partner_tenant_mappings_partner;
DROP INDEX IF EXISTS idx_white_label_partners_active;
DROP INDEX IF EXISTS idx_white_label_partners_slug;
DROP INDEX IF EXISTS idx_tenant_branding_domain_verified;
DROP INDEX IF EXISTS idx_tenant_branding_custom_domain;
DROP INDEX IF EXISTS idx_tenant_branding_org;

-- ============================================================
-- DROP TRIGGERS
-- ============================================================

DROP TRIGGER IF EXISTS set_white_label_partners_updated_at ON white_label_partners;
DROP TRIGGER IF EXISTS set_tenant_branding_updated_at ON tenant_branding;

-- ============================================================
-- DROP POLICIES
-- ============================================================

DROP POLICY IF EXISTS partner_tenant_mappings_org_isolation ON partner_tenant_mappings;
DROP POLICY IF EXISTS tenant_branding_org_isolation ON tenant_branding;

-- ============================================================
-- DROP TABLES (order matters for FK constraints)
-- ============================================================

DROP TABLE IF EXISTS partner_tenant_mappings;
DROP TABLE IF EXISTS white_label_partners;
DROP TABLE IF EXISTS tenant_branding;

-- ============================================================
-- DROP ENUMS
-- ============================================================

DROP TYPE IF EXISTS ssl_status;
DROP TYPE IF EXISTS domain_verification_status;
DROP TYPE IF EXISTS ui_density;
DROP TYPE IF EXISTS corner_radius_style;
DROP TYPE IF EXISTS sidebar_style;
