-- 000031_compliance_as_code.up.sql
-- Compliance-as-Code Engine: Git-based policy-as-code with drift detection

-- ============================================================
-- ENUMS
-- ============================================================

CREATE TYPE cac_provider AS ENUM (
    'github',
    'gitlab',
    'bitbucket',
    'azure_devops'
);

CREATE TYPE cac_sync_status AS ENUM (
    'pending',
    'running',
    'awaiting_approval',
    'approved',
    'rejected',
    'applying',
    'completed',
    'failed'
);

CREATE TYPE cac_sync_direction AS ENUM (
    'pull',
    'push',
    'bidirectional'
);

CREATE TYPE cac_resource_status AS ENUM (
    'synced',
    'pending',
    'drift_detected',
    'conflict',
    'orphaned',
    'error'
);

CREATE TYPE cac_drift_direction AS ENUM (
    'repo_ahead',
    'platform_ahead',
    'conflict'
);

CREATE TYPE cac_drift_status AS ENUM (
    'open',
    'acknowledged',
    'resolved',
    'suppressed'
);

CREATE TYPE cac_diff_action_type AS ENUM (
    'create',
    'update',
    'delete',
    'no_change'
);

-- ============================================================
-- TABLE: cac_repositories
-- ============================================================

CREATE TABLE cac_repositories (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    provider        cac_provider NOT NULL DEFAULT 'github',
    repo_url        TEXT NOT NULL,
    branch          TEXT NOT NULL DEFAULT 'main',
    base_path       TEXT NOT NULL DEFAULT '/',
    webhook_secret  TEXT,
    access_token    TEXT,
    sync_direction  cac_sync_direction NOT NULL DEFAULT 'pull',
    auto_sync       BOOLEAN NOT NULL DEFAULT false,
    require_approval BOOLEAN NOT NULL DEFAULT true,
    last_sync_at    TIMESTAMPTZ,
    last_sync_status cac_sync_status,
    resource_count  INTEGER NOT NULL DEFAULT 0,
    is_active       BOOLEAN NOT NULL DEFAULT true,
    metadata        JSONB NOT NULL DEFAULT '{}',
    created_by      UUID REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- TABLE: cac_sync_runs
-- ============================================================

CREATE TABLE cac_sync_runs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    repository_id   UUID NOT NULL REFERENCES cac_repositories(id) ON DELETE CASCADE,
    status          cac_sync_status NOT NULL DEFAULT 'pending',
    direction       cac_sync_direction NOT NULL DEFAULT 'pull',
    trigger_type    TEXT NOT NULL DEFAULT 'manual',
    commit_sha      TEXT,
    branch          TEXT,
    files_changed   INTEGER NOT NULL DEFAULT 0,
    resources_added INTEGER NOT NULL DEFAULT 0,
    resources_updated INTEGER NOT NULL DEFAULT 0,
    resources_deleted INTEGER NOT NULL DEFAULT 0,
    resources_unchanged INTEGER NOT NULL DEFAULT 0,
    errors          JSONB NOT NULL DEFAULT '[]',
    diff_plan       JSONB NOT NULL DEFAULT '{}',
    started_at      TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,
    approved_by     UUID REFERENCES users(id),
    approved_at     TIMESTAMPTZ,
    rejected_by     UUID REFERENCES users(id),
    rejected_at     TIMESTAMPTZ,
    rejection_reason TEXT,
    triggered_by    UUID REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- TABLE: cac_resource_mappings
-- ============================================================

CREATE TABLE cac_resource_mappings (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    repository_id   UUID NOT NULL REFERENCES cac_repositories(id) ON DELETE CASCADE,
    file_path       TEXT NOT NULL,
    api_version     TEXT NOT NULL DEFAULT 'complianceforge.io/v1',
    kind            TEXT NOT NULL,
    resource_name   TEXT NOT NULL,
    resource_uid    TEXT NOT NULL,
    platform_entity_type TEXT,
    platform_entity_id   UUID,
    status          cac_resource_status NOT NULL DEFAULT 'pending',
    content_hash    TEXT NOT NULL DEFAULT '',
    last_synced_at  TIMESTAMPTZ,
    last_sync_run_id UUID REFERENCES cac_sync_runs(id),
    metadata        JSONB NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (organization_id, repository_id, resource_uid)
);

-- ============================================================
-- TABLE: cac_drift_events
-- ============================================================

CREATE TABLE cac_drift_events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    repository_id   UUID NOT NULL REFERENCES cac_repositories(id) ON DELETE CASCADE,
    resource_mapping_id UUID REFERENCES cac_resource_mappings(id) ON DELETE SET NULL,
    drift_direction cac_drift_direction NOT NULL,
    status          cac_drift_status NOT NULL DEFAULT 'open',
    kind            TEXT NOT NULL,
    resource_name   TEXT NOT NULL,
    resource_uid    TEXT NOT NULL,
    field_path      TEXT,
    repo_value      TEXT,
    platform_value  TEXT,
    description     TEXT NOT NULL DEFAULT '',
    detected_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at     TIMESTAMPTZ,
    resolved_by     UUID REFERENCES users(id),
    resolution_action TEXT,
    resolution_notes TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- INDEXES
-- ============================================================

-- cac_repositories
CREATE INDEX idx_cac_repositories_org ON cac_repositories(organization_id);
CREATE INDEX idx_cac_repositories_org_active ON cac_repositories(organization_id) WHERE is_active = true;
CREATE INDEX idx_cac_repositories_provider ON cac_repositories(provider);

-- cac_sync_runs
CREATE INDEX idx_cac_sync_runs_org ON cac_sync_runs(organization_id);
CREATE INDEX idx_cac_sync_runs_repo ON cac_sync_runs(repository_id);
CREATE INDEX idx_cac_sync_runs_status ON cac_sync_runs(status);
CREATE INDEX idx_cac_sync_runs_repo_created ON cac_sync_runs(repository_id, created_at DESC);

-- cac_resource_mappings
CREATE INDEX idx_cac_resource_mappings_org ON cac_resource_mappings(organization_id);
CREATE INDEX idx_cac_resource_mappings_repo ON cac_resource_mappings(repository_id);
CREATE INDEX idx_cac_resource_mappings_kind ON cac_resource_mappings(kind);
CREATE INDEX idx_cac_resource_mappings_status ON cac_resource_mappings(status);
CREATE INDEX idx_cac_resource_mappings_uid ON cac_resource_mappings(organization_id, resource_uid);

-- cac_drift_events
CREATE INDEX idx_cac_drift_events_org ON cac_drift_events(organization_id);
CREATE INDEX idx_cac_drift_events_repo ON cac_drift_events(repository_id);
CREATE INDEX idx_cac_drift_events_status ON cac_drift_events(status);
CREATE INDEX idx_cac_drift_events_direction ON cac_drift_events(drift_direction);
CREATE INDEX idx_cac_drift_events_open ON cac_drift_events(organization_id, status) WHERE status = 'open';
CREATE INDEX idx_cac_drift_events_detected ON cac_drift_events(detected_at DESC);

-- ============================================================
-- ROW-LEVEL SECURITY
-- ============================================================

ALTER TABLE cac_repositories ENABLE ROW LEVEL SECURITY;
ALTER TABLE cac_sync_runs ENABLE ROW LEVEL SECURITY;
ALTER TABLE cac_resource_mappings ENABLE ROW LEVEL SECURITY;
ALTER TABLE cac_drift_events ENABLE ROW LEVEL SECURITY;

-- cac_repositories
CREATE POLICY cac_repositories_tenant_isolation ON cac_repositories
    USING (organization_id::text = current_setting('app.current_org', true));

-- cac_sync_runs
CREATE POLICY cac_sync_runs_tenant_isolation ON cac_sync_runs
    USING (organization_id::text = current_setting('app.current_org', true));

-- cac_resource_mappings
CREATE POLICY cac_resource_mappings_tenant_isolation ON cac_resource_mappings
    USING (organization_id::text = current_setting('app.current_org', true));

-- cac_drift_events
CREATE POLICY cac_drift_events_tenant_isolation ON cac_drift_events
    USING (organization_id::text = current_setting('app.current_org', true));

-- ============================================================
-- TRIGGERS — auto-update updated_at
-- ============================================================

CREATE OR REPLACE FUNCTION update_cac_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_cac_repositories_updated
    BEFORE UPDATE ON cac_repositories
    FOR EACH ROW EXECUTE FUNCTION update_cac_updated_at();

CREATE TRIGGER trg_cac_sync_runs_updated
    BEFORE UPDATE ON cac_sync_runs
    FOR EACH ROW EXECUTE FUNCTION update_cac_updated_at();

CREATE TRIGGER trg_cac_resource_mappings_updated
    BEFORE UPDATE ON cac_resource_mappings
    FOR EACH ROW EXECUTE FUNCTION update_cac_updated_at();

CREATE TRIGGER trg_cac_drift_events_updated
    BEFORE UPDATE ON cac_drift_events
    FOR EACH ROW EXECUTE FUNCTION update_cac_updated_at();
