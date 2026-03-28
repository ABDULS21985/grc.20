-- ============================================================
-- ComplianceForge GRC Platform
-- Migration 009: GDPR Data Subject Request (DSR) Management
-- ============================================================

-- ============================================================
-- ENUM TYPES
-- ============================================================

CREATE TYPE dsr_request_type AS ENUM (
    'access',
    'erasure',
    'rectification',
    'portability',
    'restriction',
    'objection',
    'automated_decision'
);

CREATE TYPE dsr_status AS ENUM (
    'received',
    'identity_verification',
    'in_progress',
    'extended',
    'completed',
    'rejected',
    'withdrawn'
);

CREATE TYPE dsr_priority AS ENUM (
    'standard',
    'urgent',
    'complex'
);

CREATE TYPE dsr_request_source AS ENUM (
    'email',
    'form',
    'phone',
    'letter',
    'in_person',
    'portal'
);

CREATE TYPE dsr_response_method AS ENUM (
    'email',
    'post',
    'portal',
    'in_person'
);

CREATE TYPE dsr_sla_status AS ENUM (
    'on_track',
    'at_risk',
    'overdue'
);

CREATE TYPE dsr_task_type AS ENUM (
    'verify_identity',
    'locate_data',
    'extract_data',
    'review_data',
    'compile_response',
    'notify_processors',
    'execute_erasure',
    'confirm_erasure',
    'send_response',
    'notify_third_parties',
    'review_exemptions',
    'verify_correction',
    'execute_correction',
    'extract_in_machine_readable',
    'send_confirmation'
);

CREATE TYPE dsr_task_status AS ENUM (
    'pending',
    'in_progress',
    'completed',
    'blocked',
    'not_applicable'
);

-- ============================================================
-- SEQUENCE FOR DSR REFERENCE NUMBERS
-- ============================================================

CREATE OR REPLACE FUNCTION generate_dsr_ref(org_id UUID)
RETURNS VARCHAR AS $$
DECLARE
    next_num INT;
    ref_val VARCHAR;
    current_year TEXT;
BEGIN
    current_year := EXTRACT(YEAR FROM CURRENT_DATE)::TEXT;
    SELECT COALESCE(MAX(
        CAST(SUBSTRING(request_ref FROM '-[0-9]{4}-([0-9]+)$') AS INT)
    ), 0) + 1 INTO next_num
    FROM dsr_requests
    WHERE organization_id = org_id
      AND request_ref LIKE 'DSR-' || current_year || '-%';
    ref_val := 'DSR-' || current_year || '-' || LPAD(next_num::TEXT, 4, '0');
    RETURN ref_val;
END;
$$ LANGUAGE plpgsql;

-- ============================================================
-- DSR REQUESTS
-- ============================================================

CREATE TABLE dsr_requests (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id             UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    request_ref                 VARCHAR(20) NOT NULL,
    request_type                dsr_request_type NOT NULL,
    status                      dsr_status NOT NULL DEFAULT 'received',
    priority                    dsr_priority NOT NULL DEFAULT 'standard',

    -- Data Subject PII (encrypted at application layer via AES-256-GCM)
    data_subject_name_encrypted     TEXT,
    data_subject_email_encrypted    TEXT,
    data_subject_phone_encrypted    TEXT,
    data_subject_address_encrypted  TEXT,

    -- Identity Verification
    data_subject_id_verified        BOOLEAN DEFAULT false,
    identity_verification_method    VARCHAR(100),
    identity_verified_at            TIMESTAMPTZ,
    identity_verified_by            UUID REFERENCES users(id),

    -- Request Details
    request_description             TEXT,
    request_source                  dsr_request_source NOT NULL DEFAULT 'email',
    received_date                   DATE NOT NULL DEFAULT CURRENT_DATE,
    acknowledged_at                 TIMESTAMPTZ,

    -- Deadlines (GDPR Article 12(3): 30 days, extendable by 60 days)
    response_deadline               DATE NOT NULL,
    extended_deadline               DATE,
    extension_reason                TEXT,
    extension_notified_at           TIMESTAMPTZ,

    -- Assignment
    assigned_to                     UUID REFERENCES users(id),

    -- Data Scope
    data_systems_affected           TEXT[] DEFAULT '{}',
    data_categories_affected        TEXT[] DEFAULT '{}',
    third_parties_notified          TEXT[] DEFAULT '{}',
    processing_notes                TEXT,

    -- Completion
    completed_at                    TIMESTAMPTZ,
    completed_by                    UUID REFERENCES users(id),
    response_method                 dsr_response_method,
    response_document_path          TEXT,

    -- Rejection
    rejection_reason                TEXT,
    rejection_legal_basis           TEXT,

    -- SLA Tracking (computed fields updated by scheduler)
    sla_status                      dsr_sla_status DEFAULT 'on_track',
    days_remaining                  INT,
    was_extended                    BOOLEAN DEFAULT false,
    was_completed_on_time           BOOLEAN,

    -- Metadata
    metadata                        JSONB DEFAULT '{}',
    created_at                      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at                      TIMESTAMPTZ,

    UNIQUE(organization_id, request_ref)
);

-- Indexes
CREATE INDEX idx_dsr_requests_org ON dsr_requests(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_dsr_requests_status ON dsr_requests(organization_id, status) WHERE deleted_at IS NULL;
CREATE INDEX idx_dsr_requests_type ON dsr_requests(organization_id, request_type) WHERE deleted_at IS NULL;
CREATE INDEX idx_dsr_requests_deadline ON dsr_requests(response_deadline) WHERE deleted_at IS NULL AND status NOT IN ('completed', 'rejected', 'withdrawn');
CREATE INDEX idx_dsr_requests_sla ON dsr_requests(organization_id, sla_status) WHERE deleted_at IS NULL AND status NOT IN ('completed', 'rejected', 'withdrawn');
CREATE INDEX idx_dsr_requests_assigned ON dsr_requests(assigned_to) WHERE deleted_at IS NULL;
CREATE INDEX idx_dsr_requests_received ON dsr_requests(organization_id, received_date DESC) WHERE deleted_at IS NULL;

-- Row-Level Security
ALTER TABLE dsr_requests ENABLE ROW LEVEL SECURITY;
CREATE POLICY dsr_requests_tenant ON dsr_requests
    USING (organization_id = get_current_tenant() OR get_current_tenant() IS NULL);

-- Updated-at trigger
CREATE TRIGGER trg_dsr_requests_updated_at
    BEFORE UPDATE ON dsr_requests
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================
-- DSR TASKS
-- ============================================================

CREATE TABLE dsr_tasks (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id     UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    dsr_request_id      UUID NOT NULL REFERENCES dsr_requests(id) ON DELETE CASCADE,
    task_type           dsr_task_type NOT NULL,
    description         TEXT,
    system_name         VARCHAR(200),
    assigned_to         UUID REFERENCES users(id),
    status              dsr_task_status NOT NULL DEFAULT 'pending',
    due_date            DATE,
    completed_at        TIMESTAMPTZ,
    completed_by        UUID REFERENCES users(id),
    notes               TEXT,
    evidence_path       TEXT,
    sort_order          INT DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_dsr_tasks_request ON dsr_tasks(dsr_request_id);
CREATE INDEX idx_dsr_tasks_org ON dsr_tasks(organization_id);
CREATE INDEX idx_dsr_tasks_status ON dsr_tasks(organization_id, status);
CREATE INDEX idx_dsr_tasks_assigned ON dsr_tasks(assigned_to) WHERE status NOT IN ('completed', 'not_applicable');

-- Row-Level Security
ALTER TABLE dsr_tasks ENABLE ROW LEVEL SECURITY;
CREATE POLICY dsr_tasks_tenant ON dsr_tasks
    USING (organization_id = get_current_tenant() OR get_current_tenant() IS NULL);

-- Updated-at trigger
CREATE TRIGGER trg_dsr_tasks_updated_at
    BEFORE UPDATE ON dsr_tasks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================
-- DSR AUDIT TRAIL
-- ============================================================

CREATE TABLE dsr_audit_trail (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id     UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    dsr_request_id      UUID NOT NULL REFERENCES dsr_requests(id) ON DELETE CASCADE,
    action              VARCHAR(100) NOT NULL,
    performed_by        UUID REFERENCES users(id),
    description         TEXT,
    metadata            JSONB DEFAULT '{}',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_dsr_audit_request ON dsr_audit_trail(dsr_request_id);
CREATE INDEX idx_dsr_audit_org ON dsr_audit_trail(organization_id);
CREATE INDEX idx_dsr_audit_created ON dsr_audit_trail(dsr_request_id, created_at DESC);

-- Row-Level Security
ALTER TABLE dsr_audit_trail ENABLE ROW LEVEL SECURITY;
CREATE POLICY dsr_audit_trail_tenant ON dsr_audit_trail
    USING (organization_id = get_current_tenant() OR get_current_tenant() IS NULL);

-- ============================================================
-- DSR RESPONSE TEMPLATES
-- ============================================================

CREATE TABLE dsr_response_templates (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id     UUID REFERENCES organizations(id) ON DELETE CASCADE,
    request_type        dsr_request_type NOT NULL,
    name                VARCHAR(255) NOT NULL,
    subject             TEXT NOT NULL,
    body_html           TEXT NOT NULL,
    body_text           TEXT NOT NULL,
    is_system           BOOLEAN DEFAULT false,
    language            VARCHAR(10) DEFAULT 'en',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_dsr_templates_org ON dsr_response_templates(organization_id);
CREATE INDEX idx_dsr_templates_type ON dsr_response_templates(request_type);
CREATE INDEX idx_dsr_templates_system ON dsr_response_templates(is_system) WHERE is_system = true;

-- Row-Level Security
ALTER TABLE dsr_response_templates ENABLE ROW LEVEL SECURITY;
CREATE POLICY dsr_templates_tenant ON dsr_response_templates
    USING (organization_id = get_current_tenant() OR organization_id IS NULL OR get_current_tenant() IS NULL);

-- Updated-at trigger
CREATE TRIGGER trg_dsr_templates_updated_at
    BEFORE UPDATE ON dsr_response_templates
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================
-- VIEWS
-- ============================================================

CREATE OR REPLACE VIEW v_dsr_dashboard AS
SELECT
    dr.organization_id,
    COUNT(*) AS total_requests,
    COUNT(*) FILTER (WHERE dr.status = 'received') AS received_count,
    COUNT(*) FILTER (WHERE dr.status = 'identity_verification') AS verification_count,
    COUNT(*) FILTER (WHERE dr.status = 'in_progress') AS in_progress_count,
    COUNT(*) FILTER (WHERE dr.status = 'extended') AS extended_count,
    COUNT(*) FILTER (WHERE dr.status = 'completed') AS completed_count,
    COUNT(*) FILTER (WHERE dr.status = 'rejected') AS rejected_count,
    COUNT(*) FILTER (WHERE dr.status = 'withdrawn') AS withdrawn_count,
    COUNT(*) FILTER (WHERE dr.sla_status = 'on_track') AS on_track_count,
    COUNT(*) FILTER (WHERE dr.sla_status = 'at_risk') AS at_risk_count,
    COUNT(*) FILTER (WHERE dr.sla_status = 'overdue') AS overdue_count,
    COUNT(*) FILTER (WHERE dr.request_type = 'access') AS access_count,
    COUNT(*) FILTER (WHERE dr.request_type = 'erasure') AS erasure_count,
    COUNT(*) FILTER (WHERE dr.request_type = 'rectification') AS rectification_count,
    COUNT(*) FILTER (WHERE dr.request_type = 'portability') AS portability_count,
    COUNT(*) FILTER (WHERE dr.request_type = 'restriction') AS restriction_count,
    COUNT(*) FILTER (WHERE dr.request_type = 'objection') AS objection_count,
    COUNT(*) FILTER (WHERE dr.request_type = 'automated_decision') AS automated_decision_count,
    COALESCE(AVG(
        CASE WHEN dr.completed_at IS NOT NULL
        THEN EXTRACT(EPOCH FROM (dr.completed_at - dr.created_at)) / 86400.0
        END
    ), 0) AS avg_completion_days,
    COUNT(*) FILTER (WHERE dr.was_completed_on_time = true) AS completed_on_time_count,
    COUNT(*) FILTER (WHERE dr.was_completed_on_time = false) AS completed_late_count
FROM dsr_requests dr
WHERE dr.deleted_at IS NULL
GROUP BY dr.organization_id;
