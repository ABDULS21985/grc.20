-- 000017_marketplace.up.sql
-- Control Library Marketplace & Framework Template Exchange
-- Provides a marketplace for sharing, discovering, and installing compliance
-- control packs, framework templates, and policy bundles across organisations.

BEGIN;

-- ============================================================
-- ENUM TYPES
-- ============================================================

CREATE TYPE marketplace_package_type AS ENUM (
    'control_pack',
    'framework_template',
    'policy_bundle',
    'risk_library',
    'assessment_template',
    'report_template'
);

CREATE TYPE marketplace_pricing_model AS ENUM (
    'free',
    'one_time',
    'subscription'
);

CREATE TYPE marketplace_package_status AS ENUM (
    'draft',
    'in_review',
    'published',
    'suspended',
    'deprecated',
    'archived'
);

CREATE TYPE marketplace_installation_status AS ENUM (
    'installing',
    'installed',
    'update_available',
    'updating',
    'uninstalling',
    'uninstalled',
    'failed'
);

-- ============================================================
-- TABLE: marketplace_publishers
-- Organisations or individuals that publish packages to the
-- marketplace. Verified publishers get a trust badge.
-- ============================================================

CREATE TABLE marketplace_publishers (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id   UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    publisher_name    VARCHAR(255) NOT NULL,
    publisher_slug    VARCHAR(255) NOT NULL UNIQUE,
    description       TEXT,
    website           VARCHAR(512),
    logo_url          VARCHAR(512),
    is_verified       BOOLEAN NOT NULL DEFAULT false,
    verification_date DATE,
    is_official       BOOLEAN NOT NULL DEFAULT false,
    total_packages    INT NOT NULL DEFAULT 0,
    total_downloads   INT NOT NULL DEFAULT 0,
    rating_avg        DECIMAL(3,2) DEFAULT 0.00,
    rating_count      INT NOT NULL DEFAULT 0,
    contact_email     VARCHAR(255),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ============================================================
-- TABLE: marketplace_packages
-- Individual packages published to the marketplace. Each package
-- has versioned releases, categorisation, and usage statistics.
-- ============================================================

CREATE TABLE marketplace_packages (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    publisher_id          UUID NOT NULL REFERENCES marketplace_publishers(id) ON DELETE CASCADE,
    package_slug          VARCHAR(255) NOT NULL,
    name                  VARCHAR(255) NOT NULL,
    description           TEXT,
    long_description      TEXT,
    package_type          marketplace_package_type NOT NULL,
    category              VARCHAR(100),
    applicable_frameworks TEXT[] DEFAULT '{}',
    applicable_regions    TEXT[] DEFAULT '{}',
    applicable_industries TEXT[] DEFAULT '{}',
    tags                  TEXT[] DEFAULT '{}',
    current_version       VARCHAR(20),
    min_platform_version  VARCHAR(20),
    pricing_model         marketplace_pricing_model NOT NULL DEFAULT 'free',
    price_eur             DECIMAL(10,2) DEFAULT 0.00,
    download_count        INT NOT NULL DEFAULT 0,
    install_count         INT NOT NULL DEFAULT 0,
    rating_avg            DECIMAL(3,2) DEFAULT 0.00,
    rating_count          INT NOT NULL DEFAULT 0,
    featured              BOOLEAN NOT NULL DEFAULT false,
    contents_summary      JSONB DEFAULT '{}'::jsonb,
    status                marketplace_package_status NOT NULL DEFAULT 'draft',
    published_at          TIMESTAMPTZ,
    deprecated_at         TIMESTAMPTZ,
    deprecation_message   TEXT,
    license               VARCHAR(100) NOT NULL DEFAULT 'CC-BY-4.0',
    created_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_publisher_package_slug UNIQUE (publisher_id, package_slug)
);

-- ============================================================
-- TABLE: marketplace_package_versions
-- Immutable release snapshots. package_data holds the full
-- serialised content (controls, mappings, policies, etc.).
-- ============================================================

CREATE TABLE marketplace_package_versions (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    package_id        UUID NOT NULL REFERENCES marketplace_packages(id) ON DELETE CASCADE,
    version           VARCHAR(20) NOT NULL,
    release_notes     TEXT,
    package_data      JSONB NOT NULL,
    package_hash      VARCHAR(128) NOT NULL,
    file_size_bytes   BIGINT NOT NULL DEFAULT 0,
    is_breaking_change BOOLEAN NOT NULL DEFAULT false,
    migration_notes   TEXT,
    published_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_package_version UNIQUE (package_id, version)
);

-- ============================================================
-- TABLE: marketplace_installations
-- Tracks which organisations have installed which packages,
-- their current version, configuration, and import results.
-- ============================================================

CREATE TABLE marketplace_installations (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id   UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    package_id        UUID NOT NULL REFERENCES marketplace_packages(id) ON DELETE CASCADE,
    version_id        UUID NOT NULL REFERENCES marketplace_package_versions(id),
    installed_version VARCHAR(20) NOT NULL,
    status            marketplace_installation_status NOT NULL DEFAULT 'installing',
    installed_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    installed_by      UUID REFERENCES users(id),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    uninstalled_at    TIMESTAMPTZ,
    configuration     JSONB DEFAULT '{}'::jsonb,
    import_summary    JSONB DEFAULT '{}'::jsonb,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_org_package UNIQUE (organization_id, package_id)
);

-- ============================================================
-- TABLE: marketplace_reviews
-- User reviews and ratings for installed packages. Only one
-- review per organisation per package is permitted.
-- ============================================================

CREATE TABLE marketplace_reviews (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    package_id        UUID NOT NULL REFERENCES marketplace_packages(id) ON DELETE CASCADE,
    organization_id   UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id           UUID NOT NULL REFERENCES users(id),
    rating            INT NOT NULL CHECK (rating >= 1 AND rating <= 5),
    title             VARCHAR(255),
    review_text       TEXT,
    helpful_count     INT NOT NULL DEFAULT 0,
    is_verified_install BOOLEAN NOT NULL DEFAULT false,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_package_org_review UNIQUE (package_id, organization_id)
);

-- ============================================================
-- ROW-LEVEL SECURITY
-- Installations and reviews are org-scoped.
-- ============================================================

ALTER TABLE marketplace_installations ENABLE ROW LEVEL SECURITY;
ALTER TABLE marketplace_reviews ENABLE ROW LEVEL SECURITY;

CREATE POLICY marketplace_installations_tenant ON marketplace_installations
    USING (organization_id::text = current_setting('app.current_tenant', true));

CREATE POLICY marketplace_reviews_tenant ON marketplace_reviews
    USING (organization_id::text = current_setting('app.current_tenant', true));

-- ============================================================
-- INDEXES
-- ============================================================

-- marketplace_publishers
CREATE INDEX idx_mp_publishers_org ON marketplace_publishers(organization_id);
CREATE INDEX idx_mp_publishers_slug ON marketplace_publishers(publisher_slug);

-- marketplace_packages: main discovery queries
CREATE INDEX idx_mp_packages_publisher ON marketplace_packages(publisher_id);
CREATE INDEX idx_mp_packages_slug ON marketplace_packages(package_slug);
CREATE INDEX idx_mp_packages_status ON marketplace_packages(status);
CREATE INDEX idx_mp_packages_category ON marketplace_packages(category);
CREATE INDEX idx_mp_packages_type ON marketplace_packages(package_type);
CREATE INDEX idx_mp_packages_featured ON marketplace_packages(featured) WHERE featured = true;

-- Full-text search index on name + description
CREATE INDEX idx_mp_packages_search ON marketplace_packages
    USING gin(to_tsvector('english', coalesce(name, '') || ' ' || coalesce(description, '')));

-- GIN indexes for array containment queries
CREATE INDEX idx_mp_packages_frameworks ON marketplace_packages USING gin(applicable_frameworks);
CREATE INDEX idx_mp_packages_regions ON marketplace_packages USING gin(applicable_regions);
CREATE INDEX idx_mp_packages_industries ON marketplace_packages USING gin(applicable_industries);
CREATE INDEX idx_mp_packages_tags ON marketplace_packages USING gin(tags);

-- marketplace_package_versions
CREATE INDEX idx_mp_versions_package ON marketplace_package_versions(package_id);
CREATE INDEX idx_mp_versions_published ON marketplace_package_versions(package_id, published_at DESC);

-- marketplace_installations
CREATE INDEX idx_mp_installations_org ON marketplace_installations(organization_id);
CREATE INDEX idx_mp_installations_package ON marketplace_installations(package_id);
CREATE INDEX idx_mp_installations_status ON marketplace_installations(organization_id, status);

-- marketplace_reviews
CREATE INDEX idx_mp_reviews_package ON marketplace_reviews(package_id);
CREATE INDEX idx_mp_reviews_org ON marketplace_reviews(organization_id);
CREATE INDEX idx_mp_reviews_rating ON marketplace_reviews(package_id, rating);

-- ============================================================
-- TRIGGER: updated_at auto-maintenance
-- ============================================================

CREATE OR REPLACE FUNCTION update_marketplace_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_mp_publishers_updated_at
    BEFORE UPDATE ON marketplace_publishers
    FOR EACH ROW EXECUTE FUNCTION update_marketplace_updated_at();

CREATE TRIGGER trg_mp_packages_updated_at
    BEFORE UPDATE ON marketplace_packages
    FOR EACH ROW EXECUTE FUNCTION update_marketplace_updated_at();

CREATE TRIGGER trg_mp_installations_updated_at
    BEFORE UPDATE ON marketplace_installations
    FOR EACH ROW EXECUTE FUNCTION update_marketplace_updated_at();

CREATE TRIGGER trg_mp_reviews_updated_at
    BEFORE UPDATE ON marketplace_reviews
    FOR EACH ROW EXECUTE FUNCTION update_marketplace_updated_at();

COMMIT;
