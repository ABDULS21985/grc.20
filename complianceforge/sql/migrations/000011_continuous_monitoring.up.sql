-- ============================================================
-- ComplianceForge GRC Platform
-- Migration 011: Continuous Monitoring & Automated Evidence Collection
-- ============================================================

-- ============================================================
-- ENUM TYPES
-- ============================================================

CREATE TYPE evidence_collection_method AS ENUM (
    'manual', 'api_fetch', 'file_watch', 'script_execution', 'email_parse', 'webhook_receive'
);

CREATE TYPE collection_run_status AS ENUM (
    'scheduled', 'running', 'success', 'failed', 'timeout', 'validation_failed'
);

CREATE TYPE compliance_monitor_type AS ENUM (
    'control_effectiveness', 'evidence_freshness', 'kri_threshold',
    'policy_attestation', 'vendor_assessment', 'training_completion'
);

CREATE TYPE monitor_check_status AS ENUM (
    'passing', 'failing', 'unknown'
);

CREATE TYPE monitor_result_status AS ENUM (
    'passing', 'failing'
);

CREATE TYPE drift_type AS ENUM (
    'control_degraded', 'evidence_expired', 'kri_breached', 'policy_unattested',
    'vendor_overdue', 'training_expired', 'score_dropped'
);

CREATE TYPE drift_severity AS ENUM (
    'critical', 'high', 'medium', 'low'
);

-- ============================================================
-- EVIDENCE COLLECTION CONFIGS
-- ============================================================

CREATE TABLE evidence_collection_configs (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id             UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    control_implementation_id   UUID REFERENCES control_implementations(id) ON DELETE SET NULL,
    name                        VARCHAR(200) NOT NULL,
    collection_method           evidence_collection_method NOT NULL,
    schedule_cron               VARCHAR(100),
    schedule_description        VARCHAR(200),
    api_config                  JSONB DEFAULT '{}',
    file_config                 JSONB DEFAULT '{}',
    script_config               JSONB DEFAULT '{}',
    webhook_config              JSONB DEFAULT '{}',
    acceptance_criteria         JSONB DEFAULT '[]',
    failure_threshold           INT DEFAULT 1,
    auto_update_control_status  BOOLEAN DEFAULT false,
    is_active                   BOOLEAN DEFAULT true,
    last_collection_at          TIMESTAMPTZ,
    last_collection_status      VARCHAR(50),
    next_collection_at          TIMESTAMPTZ,
    consecutive_failures        INT DEFAULT 0,
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_evidence_configs_org ON evidence_collection_configs(organization_id);
CREATE INDEX idx_evidence_configs_active ON evidence_collection_configs(organization_id, is_active)
    WHERE is_active = true;
CREATE INDEX idx_evidence_configs_next ON evidence_collection_configs(next_collection_at)
    WHERE is_active = true AND next_collection_at IS NOT NULL;
CREATE INDEX idx_evidence_configs_ctrl ON evidence_collection_configs(control_implementation_id)
    WHERE control_implementation_id IS NOT NULL;

ALTER TABLE evidence_collection_configs ENABLE ROW LEVEL SECURITY;
CREATE POLICY evidence_configs_tenant ON evidence_collection_configs
    USING (organization_id = get_current_tenant() OR get_current_tenant() IS NULL);

CREATE TRIGGER trg_evidence_configs_updated_at
    BEFORE UPDATE ON evidence_collection_configs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================
-- EVIDENCE COLLECTION RUNS
-- ============================================================

CREATE TABLE evidence_collection_runs (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id             UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    config_id                   UUID NOT NULL REFERENCES evidence_collection_configs(id) ON DELETE CASCADE,
    control_implementation_id   UUID REFERENCES control_implementations(id) ON DELETE SET NULL,
    status                      collection_run_status NOT NULL DEFAULT 'scheduled',
    started_at                  TIMESTAMPTZ,
    completed_at                TIMESTAMPTZ,
    duration_ms                 INT,
    collected_data              JSONB DEFAULT '{}',
    validation_results          JSONB DEFAULT '[]',
    all_criteria_passed         BOOLEAN,
    evidence_id                 UUID REFERENCES control_evidence(id) ON DELETE SET NULL,
    error_message               TEXT,
    metadata                    JSONB DEFAULT '{}',
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_collection_runs_org ON evidence_collection_runs(organization_id);
CREATE INDEX idx_collection_runs_config ON evidence_collection_runs(config_id, created_at DESC);
CREATE INDEX idx_collection_runs_status ON evidence_collection_runs(organization_id, status);
CREATE INDEX idx_collection_runs_time ON evidence_collection_runs(created_at DESC);

ALTER TABLE evidence_collection_runs ENABLE ROW LEVEL SECURITY;
CREATE POLICY collection_runs_tenant ON evidence_collection_runs
    USING (organization_id = get_current_tenant() OR get_current_tenant() IS NULL);

-- ============================================================
-- COMPLIANCE MONITORS
-- ============================================================

CREATE TABLE compliance_monitors (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id         UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name                    VARCHAR(200) NOT NULL,
    monitor_type            compliance_monitor_type NOT NULL,
    target_entity_type      VARCHAR(50),
    target_entity_id        UUID,
    check_frequency_cron    VARCHAR(100),
    conditions              JSONB DEFAULT '{}',
    alert_on_failure        BOOLEAN DEFAULT true,
    alert_severity          VARCHAR(20) DEFAULT 'high',
    is_active               BOOLEAN DEFAULT true,
    last_check_at           TIMESTAMPTZ,
    last_check_status       monitor_check_status DEFAULT 'unknown',
    consecutive_failures    INT DEFAULT 0,
    failure_since           TIMESTAMPTZ,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_monitors_org ON compliance_monitors(organization_id);
CREATE INDEX idx_monitors_active ON compliance_monitors(organization_id, is_active)
    WHERE is_active = true;
CREATE INDEX idx_monitors_type ON compliance_monitors(organization_id, monitor_type);
CREATE INDEX idx_monitors_status ON compliance_monitors(organization_id, last_check_status);
CREATE INDEX idx_monitors_entity ON compliance_monitors(target_entity_type, target_entity_id);

ALTER TABLE compliance_monitors ENABLE ROW LEVEL SECURITY;
CREATE POLICY monitors_tenant ON compliance_monitors
    USING (organization_id = get_current_tenant() OR get_current_tenant() IS NULL);

CREATE TRIGGER trg_monitors_updated_at
    BEFORE UPDATE ON compliance_monitors
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================
-- COMPLIANCE MONITOR RESULTS
-- ============================================================

CREATE TABLE compliance_monitor_results (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id     UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    monitor_id          UUID NOT NULL REFERENCES compliance_monitors(id) ON DELETE CASCADE,
    status              monitor_result_status NOT NULL,
    check_time          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    result_data         JSONB DEFAULT '{}',
    message             TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_monitor_results_org ON compliance_monitor_results(organization_id);
CREATE INDEX idx_monitor_results_monitor ON compliance_monitor_results(monitor_id, created_at DESC);
CREATE INDEX idx_monitor_results_time ON compliance_monitor_results(created_at DESC);
CREATE INDEX idx_monitor_results_status ON compliance_monitor_results(organization_id, status);

ALTER TABLE compliance_monitor_results ENABLE ROW LEVEL SECURITY;
CREATE POLICY monitor_results_tenant ON compliance_monitor_results
    USING (organization_id = get_current_tenant() OR get_current_tenant() IS NULL);

-- ============================================================
-- COMPLIANCE DRIFT EVENTS
-- ============================================================

CREATE TABLE compliance_drift_events (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id     UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    drift_type          drift_type NOT NULL,
    severity            drift_severity NOT NULL,
    entity_type         VARCHAR(50),
    entity_id           UUID,
    entity_ref          VARCHAR(50),
    description         TEXT,
    previous_state      VARCHAR(100),
    current_state       VARCHAR(100),
    detected_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    acknowledged_at     TIMESTAMPTZ,
    acknowledged_by     UUID REFERENCES users(id),
    resolved_at         TIMESTAMPTZ,
    resolved_by         UUID REFERENCES users(id),
    resolution_notes    TEXT,
    notification_sent   BOOLEAN DEFAULT false,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_drift_events_org ON compliance_drift_events(organization_id);
CREATE INDEX idx_drift_events_active ON compliance_drift_events(organization_id)
    WHERE resolved_at IS NULL;
CREATE INDEX idx_drift_events_severity ON compliance_drift_events(organization_id, severity)
    WHERE resolved_at IS NULL;
CREATE INDEX idx_drift_events_type ON compliance_drift_events(organization_id, drift_type);
CREATE INDEX idx_drift_events_entity ON compliance_drift_events(entity_type, entity_id);
CREATE INDEX idx_drift_events_detected ON compliance_drift_events(detected_at DESC);

ALTER TABLE compliance_drift_events ENABLE ROW LEVEL SECURITY;
CREATE POLICY drift_events_tenant ON compliance_drift_events
    USING (organization_id = get_current_tenant() OR get_current_tenant() IS NULL);
