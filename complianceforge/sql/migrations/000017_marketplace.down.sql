-- 000017_marketplace.down.sql
-- Rollback: Control Library Marketplace & Framework Template Exchange

BEGIN;

-- Drop triggers
DROP TRIGGER IF EXISTS trg_mp_reviews_updated_at ON marketplace_reviews;
DROP TRIGGER IF EXISTS trg_mp_installations_updated_at ON marketplace_installations;
DROP TRIGGER IF EXISTS trg_mp_packages_updated_at ON marketplace_packages;
DROP TRIGGER IF EXISTS trg_mp_publishers_updated_at ON marketplace_publishers;

-- Drop trigger function
DROP FUNCTION IF EXISTS update_marketplace_updated_at();

-- Drop RLS policies
DROP POLICY IF EXISTS marketplace_reviews_tenant ON marketplace_reviews;
DROP POLICY IF EXISTS marketplace_installations_tenant ON marketplace_installations;

-- Drop tables in dependency order
DROP TABLE IF EXISTS marketplace_reviews;
DROP TABLE IF EXISTS marketplace_installations;
DROP TABLE IF EXISTS marketplace_package_versions;
DROP TABLE IF EXISTS marketplace_packages;
DROP TABLE IF EXISTS marketplace_publishers;

-- Drop enum types
DROP TYPE IF EXISTS marketplace_installation_status;
DROP TYPE IF EXISTS marketplace_package_status;
DROP TYPE IF EXISTS marketplace_pricing_model;
DROP TYPE IF EXISTS marketplace_package_type;

COMMIT;
