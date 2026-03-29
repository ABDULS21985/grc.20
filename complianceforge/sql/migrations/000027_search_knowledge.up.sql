-- ============================================================
-- ComplianceForge GRC Platform
-- Migration 027: Advanced Search, Knowledge Base &
--                Compliance Guidance Engine
-- ============================================================

-- ============================================================
-- ENUM TYPES
-- ============================================================

CREATE TYPE article_type AS ENUM (
    'implementation_guide',
    'regulatory_guide',
    'best_practice',
    'framework_overview',
    'control_guidance',
    'audit_preparation',
    'incident_response',
    'risk_management',
    'vendor_management'
);

CREATE TYPE article_difficulty AS ENUM (
    'beginner',
    'intermediate',
    'advanced',
    'expert'
);

-- ============================================================
-- SEARCH INDEX
-- Unified full-text search index across all entity types.
-- Entity-agnostic: any type (risk, control, policy, etc.) can
-- be indexed here for fast cross-platform search.
-- ============================================================

CREATE TABLE search_index (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id     UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    entity_type         VARCHAR(50) NOT NULL,
    entity_id           UUID NOT NULL,
    entity_ref          VARCHAR(100),
    title               TEXT NOT NULL,
    body                TEXT,
    tags                TEXT[] DEFAULT '{}',
    framework_codes     TEXT[] DEFAULT '{}',
    status              VARCHAR(50),
    severity            VARCHAR(50),
    category            VARCHAR(100),
    search_vector       TSVECTOR,
    metadata            JSONB DEFAULT '{}',
    indexed_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE(organization_id, entity_type, entity_id)
);

ALTER TABLE search_index ENABLE ROW LEVEL SECURITY;

CREATE POLICY search_index_tenant_isolation ON search_index
    USING (organization_id = get_current_tenant());

CREATE INDEX idx_search_index_org_id ON search_index(organization_id);
CREATE INDEX idx_search_index_entity_type ON search_index(organization_id, entity_type);
CREATE INDEX idx_search_index_search_vector ON search_index USING GIN(search_vector);
CREATE INDEX idx_search_index_tags ON search_index USING GIN(tags);
CREATE INDEX idx_search_index_framework_codes ON search_index USING GIN(framework_codes);
CREATE INDEX idx_search_index_status ON search_index(organization_id, status);
CREATE INDEX idx_search_index_severity ON search_index(organization_id, severity);
CREATE INDEX idx_search_index_category ON search_index(organization_id, category);

-- Trigger to automatically update search_vector on INSERT/UPDATE
CREATE OR REPLACE FUNCTION search_index_update_vector() RETURNS TRIGGER AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('english', COALESCE(NEW.title, '')), 'A') ||
        setweight(to_tsvector('english', COALESCE(NEW.entity_ref, '')), 'A') ||
        setweight(to_tsvector('english', COALESCE(array_to_string(NEW.tags, ' '), '')), 'B') ||
        setweight(to_tsvector('english', COALESCE(array_to_string(NEW.framework_codes, ' '), '')), 'B') ||
        setweight(to_tsvector('english', COALESCE(NEW.body, '')), 'C') ||
        setweight(to_tsvector('english', COALESCE(NEW.category, '')), 'D') ||
        setweight(to_tsvector('english', COALESCE(NEW.status, '')), 'D');
    NEW.updated_at := NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_search_index_vector
    BEFORE INSERT OR UPDATE ON search_index
    FOR EACH ROW EXECUTE FUNCTION search_index_update_vector();

-- ============================================================
-- KNOWLEDGE ARTICLES
-- System and organization-specific compliance guidance articles.
-- System articles (is_system = true) have NULL org_id and are
-- shared across all tenants.
-- ============================================================

CREATE TABLE knowledge_articles (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id             UUID REFERENCES organizations(id) ON DELETE CASCADE,
    article_type                article_type NOT NULL,
    title                       VARCHAR(500) NOT NULL,
    slug                        VARCHAR(500) NOT NULL,
    content_markdown            TEXT NOT NULL,
    summary                     TEXT,
    applicable_frameworks       TEXT[] DEFAULT '{}',
    applicable_control_codes    TEXT[] DEFAULT '{}',
    tags                        TEXT[] DEFAULT '{}',
    difficulty                  article_difficulty NOT NULL DEFAULT 'intermediate',
    reading_time_minutes        INT DEFAULT 5,
    is_system                   BOOLEAN NOT NULL DEFAULT false,
    is_published                BOOLEAN NOT NULL DEFAULT false,
    view_count                  INT DEFAULT 0,
    helpful_count               INT DEFAULT 0,
    not_helpful_count           INT DEFAULT 0,
    author_user_id              UUID,
    search_vector               TSVECTOR,
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE knowledge_articles ENABLE ROW LEVEL SECURITY;

-- System articles (is_system = true, org_id IS NULL) are visible to all.
-- Org-specific articles are restricted to their organization.
CREATE POLICY knowledge_articles_access ON knowledge_articles
    USING (
        is_system = true
        OR organization_id IS NULL
        OR organization_id = get_current_tenant()
    );

CREATE INDEX idx_knowledge_articles_org_id ON knowledge_articles(organization_id);
CREATE INDEX idx_knowledge_articles_slug ON knowledge_articles(slug);
CREATE INDEX idx_knowledge_articles_type ON knowledge_articles(article_type);
CREATE INDEX idx_knowledge_articles_search_vector ON knowledge_articles USING GIN(search_vector);
CREATE INDEX idx_knowledge_articles_tags ON knowledge_articles USING GIN(tags);
CREATE INDEX idx_knowledge_articles_frameworks ON knowledge_articles USING GIN(applicable_frameworks);
CREATE INDEX idx_knowledge_articles_control_codes ON knowledge_articles USING GIN(applicable_control_codes);
CREATE INDEX idx_knowledge_articles_published ON knowledge_articles(is_published, is_system);
CREATE INDEX idx_knowledge_articles_difficulty ON knowledge_articles(difficulty);

-- Unique slug per organization (or globally for system articles)
CREATE UNIQUE INDEX idx_knowledge_articles_unique_slug
    ON knowledge_articles (COALESCE(organization_id, '00000000-0000-0000-0000-000000000000'), slug);

-- Trigger to update search_vector
CREATE OR REPLACE FUNCTION knowledge_articles_update_vector() RETURNS TRIGGER AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('english', COALESCE(NEW.title, '')), 'A') ||
        setweight(to_tsvector('english', COALESCE(NEW.summary, '')), 'B') ||
        setweight(to_tsvector('english', COALESCE(array_to_string(NEW.tags, ' '), '')), 'B') ||
        setweight(to_tsvector('english', COALESCE(array_to_string(NEW.applicable_frameworks, ' '), '')), 'B') ||
        setweight(to_tsvector('english', COALESCE(NEW.content_markdown, '')), 'C');
    NEW.updated_at := NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_knowledge_articles_vector
    BEFORE INSERT OR UPDATE ON knowledge_articles
    FOR EACH ROW EXECUTE FUNCTION knowledge_articles_update_vector();

-- ============================================================
-- KNOWLEDGE BOOKMARKS
-- Users can bookmark articles for quick access.
-- ============================================================

CREATE TABLE knowledge_bookmarks (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID NOT NULL,
    article_id          UUID NOT NULL REFERENCES knowledge_articles(id) ON DELETE CASCADE,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE(user_id, article_id)
);

CREATE INDEX idx_knowledge_bookmarks_user ON knowledge_bookmarks(user_id);
CREATE INDEX idx_knowledge_bookmarks_article ON knowledge_bookmarks(article_id);

-- ============================================================
-- RECENT SEARCHES
-- Tracks user search queries for autocomplete suggestions
-- and search analytics.
-- ============================================================

CREATE TABLE recent_searches (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id     UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id             UUID NOT NULL,
    query               TEXT NOT NULL,
    result_count        INT DEFAULT 0,
    clicked_entity_type VARCHAR(50),
    clicked_entity_id   UUID,
    filters_used        JSONB DEFAULT '{}',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE recent_searches ENABLE ROW LEVEL SECURITY;

CREATE POLICY recent_searches_tenant_isolation ON recent_searches
    USING (organization_id = get_current_tenant());

CREATE INDEX idx_recent_searches_org_user ON recent_searches(organization_id, user_id);
CREATE INDEX idx_recent_searches_query ON recent_searches(organization_id, query);
CREATE INDEX idx_recent_searches_created ON recent_searches(created_at);

-- ============================================================
-- ARTICLE FEEDBACK
-- Per-user feedback on knowledge articles for quality tracking.
-- ============================================================

CREATE TABLE knowledge_article_feedback (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    article_id          UUID NOT NULL REFERENCES knowledge_articles(id) ON DELETE CASCADE,
    user_id             UUID NOT NULL,
    action              VARCHAR(50) NOT NULL,
    comment             TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE(article_id, user_id, action)
);

CREATE INDEX idx_knowledge_article_feedback_article ON knowledge_article_feedback(article_id);
CREATE INDEX idx_knowledge_article_feedback_user ON knowledge_article_feedback(user_id);
