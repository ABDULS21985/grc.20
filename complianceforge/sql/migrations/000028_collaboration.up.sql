-- 000028_collaboration.up.sql
-- Collaboration, Comments, Mentions & Activity Feed System

-- ============================================================
-- ENUM TYPES
-- ============================================================

CREATE TYPE comment_visibility AS ENUM ('public', 'internal');

CREATE TYPE activity_action AS ENUM (
    'created',
    'updated',
    'deleted',
    'status_changed',
    'assigned',
    'unassigned',
    'commented',
    'mentioned',
    'approved',
    'rejected',
    'submitted',
    'completed',
    'archived',
    'restored',
    'linked',
    'unlinked',
    'attached',
    'detached',
    'escalated',
    'transferred',
    'reviewed',
    'published',
    'revoked',
    'scored',
    'imported',
    'exported'
);

CREATE TYPE activity_visibility AS ENUM ('all', 'internal', 'admin_only');

CREATE TYPE follow_type AS ENUM ('watching', 'participating', 'mentioned');

CREATE TYPE reaction_type AS ENUM (
    'thumbs_up',
    'thumbs_down',
    'check',
    'eyes',
    'rocket',
    'warning'
);

-- ============================================================
-- TABLE: comments
-- Threaded comments on any entity (controls, risks, audits, etc).
-- Supports markdown content, @mentions, attachments, reactions,
-- internal-only notes, and resolution notes.
-- ============================================================

CREATE TABLE comments (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id         UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    entity_type             VARCHAR(50) NOT NULL,
    entity_id               UUID NOT NULL,
    parent_comment_id       UUID REFERENCES comments(id) ON DELETE CASCADE,
    thread_depth            INT NOT NULL DEFAULT 0 CHECK (thread_depth <= 3),

    -- Author
    author_user_id          UUID NOT NULL,

    -- Content
    content                 TEXT NOT NULL,
    content_html            TEXT NOT NULL DEFAULT '',

    -- Flags
    is_internal             BOOLEAN NOT NULL DEFAULT false,
    is_resolution_note      BOOLEAN NOT NULL DEFAULT false,
    is_pinned               BOOLEAN NOT NULL DEFAULT false,

    -- Mentions
    mentioned_user_ids      UUID[] DEFAULT '{}',
    mentioned_role_slugs    TEXT[] DEFAULT '{}',

    -- Attachments (stored as parallel arrays)
    attachment_paths        TEXT[] DEFAULT '{}',
    attachment_names        TEXT[] DEFAULT '{}',
    attachment_sizes        BIGINT[] DEFAULT '{}',

    -- Reactions (JSONB: {"thumbs_up": ["user_id1"], "rocket": ["user_id2"]})
    reactions               JSONB NOT NULL DEFAULT '{}',

    -- Edit tracking
    is_edited               BOOLEAN NOT NULL DEFAULT false,
    edited_at               TIMESTAMPTZ,

    -- Soft delete
    is_deleted              BOOLEAN NOT NULL DEFAULT false,
    deleted_at              TIMESTAMPTZ,
    deleted_by_user_id      UUID,

    -- Timestamps
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE comments ENABLE ROW LEVEL SECURITY;

CREATE POLICY comments_tenant_isolation ON comments
    USING (organization_id = current_setting('app.current_org')::uuid);

-- Indexes for comments
CREATE INDEX idx_comments_org_id ON comments(organization_id);
CREATE INDEX idx_comments_entity ON comments(organization_id, entity_type, entity_id);
CREATE INDEX idx_comments_parent ON comments(parent_comment_id) WHERE parent_comment_id IS NOT NULL;
CREATE INDEX idx_comments_author ON comments(author_user_id);
CREATE INDEX idx_comments_created ON comments(created_at DESC);
CREATE INDEX idx_comments_pinned ON comments(organization_id, entity_type, entity_id, is_pinned)
    WHERE is_pinned = true AND is_deleted = false;
CREATE INDEX idx_comments_mentioned_users ON comments USING GIN(mentioned_user_ids)
    WHERE is_deleted = false;
CREATE INDEX idx_comments_not_deleted ON comments(organization_id, entity_type, entity_id, created_at)
    WHERE is_deleted = false;

-- ============================================================
-- TABLE: activity_feed
-- Immutable activity log for all entity changes. Supports
-- structured diffs, system events, and visibility control.
-- ============================================================

CREATE TABLE activity_feed (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id         UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    actor_user_id           UUID,
    action                  activity_action NOT NULL,
    entity_type             VARCHAR(50) NOT NULL,
    entity_id               UUID NOT NULL,
    entity_ref              VARCHAR(50),
    entity_title            VARCHAR(500),
    description             TEXT NOT NULL DEFAULT '',
    changes                 JSONB DEFAULT '{}',
    is_system               BOOLEAN NOT NULL DEFAULT false,
    visibility              activity_visibility NOT NULL DEFAULT 'all',
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE activity_feed ENABLE ROW LEVEL SECURITY;

CREATE POLICY activity_feed_tenant_isolation ON activity_feed
    USING (organization_id = current_setting('app.current_org')::uuid);

-- Indexes for activity_feed
CREATE INDEX idx_activity_feed_org_id ON activity_feed(organization_id);
CREATE INDEX idx_activity_feed_entity ON activity_feed(organization_id, entity_type, entity_id);
CREATE INDEX idx_activity_feed_actor ON activity_feed(actor_user_id);
CREATE INDEX idx_activity_feed_created ON activity_feed(organization_id, created_at DESC);
CREATE INDEX idx_activity_feed_action ON activity_feed(organization_id, action);
CREATE INDEX idx_activity_feed_changes ON activity_feed USING GIN(changes)
    WHERE changes != '{}';
CREATE INDEX idx_activity_feed_personal ON activity_feed(organization_id, actor_user_id, created_at DESC);

-- ============================================================
-- TABLE: user_follows
-- Tracks which entities a user is following (watching, participating,
-- or mentioned). Controls notification delivery.
-- ============================================================

CREATE TABLE user_follows (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id         UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id                 UUID NOT NULL,
    entity_type             VARCHAR(50) NOT NULL,
    entity_id               UUID NOT NULL,
    follow_type             follow_type NOT NULL DEFAULT 'watching',
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_user_follows_entity UNIQUE (organization_id, user_id, entity_type, entity_id)
);

ALTER TABLE user_follows ENABLE ROW LEVEL SECURITY;

CREATE POLICY user_follows_tenant_isolation ON user_follows
    USING (organization_id = current_setting('app.current_org')::uuid);

-- Indexes for user_follows
CREATE INDEX idx_user_follows_org_id ON user_follows(organization_id);
CREATE INDEX idx_user_follows_user ON user_follows(organization_id, user_id);
CREATE INDEX idx_user_follows_entity ON user_follows(organization_id, entity_type, entity_id);
CREATE INDEX idx_user_follows_user_type ON user_follows(organization_id, user_id, follow_type);

-- ============================================================
-- TABLE: user_read_markers
-- Tracks the last-read timestamp and unread count per user per entity.
-- Used for unread badges and notification counts.
-- ============================================================

CREATE TABLE user_read_markers (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id         UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id                 UUID NOT NULL,
    entity_type             VARCHAR(50) NOT NULL,
    entity_id               UUID NOT NULL,
    last_read_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    unread_count            INT NOT NULL DEFAULT 0,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_user_read_markers_entity UNIQUE (organization_id, user_id, entity_type, entity_id)
);

ALTER TABLE user_read_markers ENABLE ROW LEVEL SECURITY;

CREATE POLICY user_read_markers_tenant_isolation ON user_read_markers
    USING (organization_id = current_setting('app.current_org')::uuid);

-- Indexes for user_read_markers
CREATE INDEX idx_user_read_markers_org_id ON user_read_markers(organization_id);
CREATE INDEX idx_user_read_markers_user ON user_read_markers(organization_id, user_id);
CREATE INDEX idx_user_read_markers_entity ON user_read_markers(organization_id, entity_type, entity_id);
CREATE INDEX idx_user_read_markers_unread ON user_read_markers(organization_id, user_id, unread_count)
    WHERE unread_count > 0;

-- ============================================================
-- TRIGGER: auto-update updated_at on comments
-- ============================================================

CREATE OR REPLACE FUNCTION update_comment_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_comment_updated_at
    BEFORE UPDATE ON comments
    FOR EACH ROW
    EXECUTE FUNCTION update_comment_updated_at();

-- ============================================================
-- TRIGGER: auto-update updated_at on user_read_markers
-- ============================================================

CREATE OR REPLACE FUNCTION update_read_marker_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_read_marker_updated_at
    BEFORE UPDATE ON user_read_markers
    FOR EACH ROW
    EXECUTE FUNCTION update_read_marker_updated_at();
