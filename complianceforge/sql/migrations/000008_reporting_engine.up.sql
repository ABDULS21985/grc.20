-- ============================================================
-- ComplianceForge GRC Platform
-- Migration 008: Advanced Reporting Engine
-- ============================================================

-- ============================================================
-- ENUM TYPES
-- ============================================================

CREATE TYPE report_type AS ENUM (
    'compliance_status',
    'risk_register',
    'risk_heatmap',
    'audit_summary',
    'audit_findings',
    'incident_summary',
    'breach_register',
    'vendor_risk',
    'policy_status',
    'attestation_report',
    'gap_analysis',
    'cross_framework_mapping',
    'executive_summary',
    'kri_dashboard',
    'treatment_progress',
    'custom'
);

CREATE TYPE report_format AS ENUM ('pdf', 'xlsx', 'csv', 'json');

CREATE TYPE report_frequency AS ENUM ('daily', 'weekly', 'monthly', 'quarterly', 'annually');

CREATE TYPE report_run_status AS ENUM ('pending', 'generating', 'completed', 'failed');

CREATE TYPE report_delivery_channel AS ENUM ('email', 'storage', 'both');

-- ============================================================
-- REPORT DEFINITIONS
-- ============================================================

CREATE TABLE report_definitions (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id             UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name                        VARCHAR(500) NOT NULL,
    description                 TEXT,
    report_type                 report_type NOT NULL,
    format                      report_format NOT NULL DEFAULT 'pdf',
    filters                     JSONB DEFAULT '{}',
    sections                    JSONB DEFAULT '[]',
    classification              VARCHAR(50) DEFAULT 'internal',
    include_executive_summary   BOOLEAN DEFAULT true,
    include_appendices          BOOLEAN DEFAULT false,
    branding                    JSONB DEFAULT '{}',
    created_by                  UUID NOT NULL REFERENCES users(id),
    is_template                 BOOLEAN DEFAULT false,
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_report_definitions_org ON report_definitions(organization_id);
CREATE INDEX idx_report_definitions_type ON report_definitions(organization_id, report_type);
CREATE INDEX idx_report_definitions_template ON report_definitions(is_template) WHERE is_template = true;
CREATE INDEX idx_report_definitions_created_by ON report_definitions(created_by);

ALTER TABLE report_definitions ENABLE ROW LEVEL SECURITY;
CREATE POLICY report_definitions_tenant ON report_definitions
    USING (organization_id = get_current_tenant() OR get_current_tenant() IS NULL);

CREATE TRIGGER trg_report_definitions_updated_at
    BEFORE UPDATE ON report_definitions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================
-- REPORT SCHEDULES
-- ============================================================

CREATE TABLE report_schedules (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id             UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    report_definition_id        UUID NOT NULL REFERENCES report_definitions(id) ON DELETE CASCADE,
    name                        VARCHAR(500) NOT NULL,
    frequency                   report_frequency NOT NULL,
    day_of_week                 INT CHECK (day_of_week >= 0 AND day_of_week <= 6),
    day_of_month                INT CHECK (day_of_month >= 1 AND day_of_month <= 31),
    time_of_day                 TIME NOT NULL DEFAULT '08:00',
    timezone                    VARCHAR(50) NOT NULL DEFAULT 'Europe/London',
    recipient_user_ids          UUID[] DEFAULT '{}',
    recipient_emails            TEXT[] DEFAULT '{}',
    delivery_channel            report_delivery_channel NOT NULL DEFAULT 'both',
    is_active                   BOOLEAN DEFAULT true,
    last_run_at                 TIMESTAMPTZ,
    next_run_at                 TIMESTAMPTZ,
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_report_schedules_org ON report_schedules(organization_id);
CREATE INDEX idx_report_schedules_definition ON report_schedules(report_definition_id);
CREATE INDEX idx_report_schedules_next_run ON report_schedules(next_run_at)
    WHERE is_active = true;
CREATE INDEX idx_report_schedules_active ON report_schedules(is_active, next_run_at)
    WHERE is_active = true;

ALTER TABLE report_schedules ENABLE ROW LEVEL SECURITY;
CREATE POLICY report_schedules_tenant ON report_schedules
    USING (organization_id = get_current_tenant() OR get_current_tenant() IS NULL);

CREATE TRIGGER trg_report_schedules_updated_at
    BEFORE UPDATE ON report_schedules
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================
-- REPORT RUNS
-- ============================================================

CREATE TABLE report_runs (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id             UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    report_definition_id        UUID NOT NULL REFERENCES report_definitions(id) ON DELETE CASCADE,
    schedule_id                 UUID REFERENCES report_schedules(id) ON DELETE SET NULL,
    status                      report_run_status NOT NULL DEFAULT 'pending',
    format                      report_format NOT NULL DEFAULT 'pdf',
    file_path                   TEXT,
    file_size_bytes             BIGINT,
    page_count                  INT,
    generation_time_ms          INT,
    parameters                  JSONB DEFAULT '{}',
    generated_by                UUID NOT NULL REFERENCES users(id),
    error_message               TEXT,
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at                TIMESTAMPTZ
);

CREATE INDEX idx_report_runs_org ON report_runs(organization_id);
CREATE INDEX idx_report_runs_definition ON report_runs(report_definition_id);
CREATE INDEX idx_report_runs_schedule ON report_runs(schedule_id) WHERE schedule_id IS NOT NULL;
CREATE INDEX idx_report_runs_status ON report_runs(organization_id, status);
CREATE INDEX idx_report_runs_created ON report_runs(organization_id, created_at DESC);
CREATE INDEX idx_report_runs_generated_by ON report_runs(generated_by);

ALTER TABLE report_runs ENABLE ROW LEVEL SECURITY;
CREATE POLICY report_runs_tenant ON report_runs
    USING (organization_id = get_current_tenant() OR get_current_tenant() IS NULL);
