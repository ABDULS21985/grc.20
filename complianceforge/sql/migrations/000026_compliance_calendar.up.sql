-- 000026_compliance_calendar.up.sql
-- Compliance Calendar & Deadline Management Engine
-- Creates tables for calendar events, user subscriptions, and sync configs.

-- ============================================================
-- ENUM TYPES
-- ============================================================

CREATE TYPE calendar_event_type AS ENUM (
    -- Policy events
    'policy_review_due',
    'policy_approval_due',
    'policy_expiry',
    'policy_attestation_due',
    -- Risk events
    'risk_review_due',
    'risk_treatment_due',
    'risk_assessment_due',
    'risk_appetite_review',
    -- Vendor events
    'vendor_assessment_due',
    'vendor_contract_expiry',
    'vendor_dpa_renewal',
    'vendor_review_due',
    -- Audit events
    'audit_planned_start',
    'audit_planned_end',
    'audit_finding_due',
    -- Evidence events
    'evidence_collection_due',
    'evidence_review_due',
    -- Exception events
    'exception_expiry',
    'exception_review_due',
    -- DSR events
    'dsr_deadline',
    'dsr_extension_deadline',
    -- Incident events
    'incident_notification_deadline',
    'incident_nis2_early_warning',
    'incident_nis2_notification',
    'incident_nis2_final_report',
    -- Regulatory events
    'regulatory_effective_date',
    'regulatory_response_deadline',
    -- Business continuity events
    'bc_plan_review_due',
    'bc_exercise_scheduled',
    'bc_plan_expiry',
    -- Board events
    'board_meeting_scheduled',
    'board_decision_action_due',
    'board_pack_due',
    -- Custom
    'custom_deadline',
    'custom_reminder',
    'custom_milestone'
);

CREATE TYPE calendar_event_category AS ENUM (
    'policy',
    'risk',
    'vendor',
    'audit',
    'evidence',
    'exception',
    'dsr',
    'incident',
    'regulatory',
    'business_continuity',
    'board',
    'custom'
);

CREATE TYPE calendar_event_priority AS ENUM (
    'critical',
    'high',
    'medium',
    'low'
);

CREATE TYPE calendar_event_status AS ENUM (
    'upcoming',
    'due_soon',
    'overdue',
    'completed',
    'cancelled',
    'snoozed'
);

CREATE TYPE calendar_recurrence_type AS ENUM (
    'none',
    'daily',
    'weekly',
    'monthly',
    'quarterly',
    'semi_annually',
    'annually',
    'custom_rrule'
);

-- ============================================================
-- TABLE: calendar_events
-- ============================================================

CREATE TABLE calendar_events (
    id                       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id          UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    event_ref                VARCHAR(30) NOT NULL,

    -- Event classification
    event_type               calendar_event_type NOT NULL,
    category                 calendar_event_category NOT NULL,
    priority                 calendar_event_priority NOT NULL DEFAULT 'medium',
    status                   calendar_event_status NOT NULL DEFAULT 'upcoming',

    -- Content
    title                    VARCHAR(500) NOT NULL,
    description              TEXT,

    -- Source entity linkage
    source_entity_type       VARCHAR(50) NOT NULL,
    source_entity_id         UUID NOT NULL,
    source_entity_ref        VARCHAR(50),

    -- Timing
    due_date                 DATE NOT NULL,
    due_time                 TIME,
    start_date               DATE,
    end_date                 DATE,
    all_day                  BOOLEAN NOT NULL DEFAULT true,
    timezone                 VARCHAR(50) NOT NULL DEFAULT 'UTC',

    -- Recurrence (RFC 5545 RRULE)
    recurrence_type          calendar_recurrence_type NOT NULL DEFAULT 'none',
    rrule                    TEXT,
    recurrence_end_date      DATE,
    parent_event_id          UUID REFERENCES calendar_events(id) ON DELETE SET NULL,
    occurrence_date          DATE,

    -- Assignment
    assigned_to_user_id      UUID REFERENCES users(id) ON DELETE SET NULL,
    assigned_to_role         VARCHAR(100),
    owner_user_id            UUID REFERENCES users(id) ON DELETE SET NULL,

    -- Notification
    reminder_days_before     INT[] NOT NULL DEFAULT '{7,3,1}',
    last_reminder_sent_at    TIMESTAMPTZ,
    reminder_count           INT NOT NULL DEFAULT 0,
    escalation_after_days    INT NOT NULL DEFAULT 7,
    escalated                BOOLEAN NOT NULL DEFAULT false,
    escalated_to_user_id     UUID REFERENCES users(id) ON DELETE SET NULL,
    escalated_at             TIMESTAMPTZ,

    -- Completion
    completed_at             TIMESTAMPTZ,
    completed_by             UUID REFERENCES users(id) ON DELETE SET NULL,
    completion_notes         TEXT,

    -- Rescheduling
    original_due_date        DATE,
    rescheduled_count        INT NOT NULL DEFAULT 0,
    reschedule_reason        TEXT,

    -- Metadata
    url                      TEXT,
    color                    VARCHAR(7),
    tags                     TEXT[] NOT NULL DEFAULT '{}',
    metadata                 JSONB NOT NULL DEFAULT '{}',

    created_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_calendar_events_ref UNIQUE (organization_id, event_ref),
    CONSTRAINT uq_calendar_events_source UNIQUE (organization_id, source_entity_type, source_entity_id, event_type, COALESCE(occurrence_date, '1970-01-01'::DATE))
);

-- ============================================================
-- TABLE: calendar_subscriptions
-- ============================================================

CREATE TABLE calendar_subscriptions (
    id                       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id          UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id                  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Subscription preferences
    subscribed_categories    calendar_event_category[] NOT NULL DEFAULT '{policy,risk,vendor,audit,evidence,exception,dsr,incident,regulatory,business_continuity,board,custom}',
    subscribed_priorities    calendar_event_priority[] NOT NULL DEFAULT '{critical,high,medium,low}',
    email_reminders          BOOLEAN NOT NULL DEFAULT true,
    in_app_reminders         BOOLEAN NOT NULL DEFAULT true,
    daily_digest             BOOLEAN NOT NULL DEFAULT false,
    weekly_digest            BOOLEAN NOT NULL DEFAULT true,

    -- iCal export
    ical_export_enabled      BOOLEAN NOT NULL DEFAULT false,
    ical_token               VARCHAR(128),
    ical_token_expires_at    TIMESTAMPTZ,

    -- Notification preferences
    reminder_days_override   INT[],
    quiet_hours_start        TIME,
    quiet_hours_end          TIME,
    timezone                 VARCHAR(50) NOT NULL DEFAULT 'UTC',

    created_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_calendar_subscriptions_user UNIQUE (organization_id, user_id)
);

-- ============================================================
-- TABLE: calendar_sync_configs
-- ============================================================

CREATE TABLE calendar_sync_configs (
    id                       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id          UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- Module identification
    module_name              VARCHAR(50) NOT NULL,
    is_enabled               BOOLEAN NOT NULL DEFAULT true,

    -- Sync settings
    sync_frequency_minutes   INT NOT NULL DEFAULT 30,
    last_sync_at             TIMESTAMPTZ,
    last_sync_status         VARCHAR(20) NOT NULL DEFAULT 'pending',
    last_sync_events_created INT NOT NULL DEFAULT 0,
    last_sync_events_updated INT NOT NULL DEFAULT 0,
    last_sync_error          TEXT,

    -- Configuration
    auto_create_events       BOOLEAN NOT NULL DEFAULT true,
    auto_complete_events     BOOLEAN NOT NULL DEFAULT true,
    default_reminder_days    INT[] NOT NULL DEFAULT '{7,3,1}',
    default_priority         calendar_event_priority NOT NULL DEFAULT 'medium',

    created_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_calendar_sync_configs UNIQUE (organization_id, module_name)
);

-- ============================================================
-- ROW-LEVEL SECURITY
-- ============================================================

ALTER TABLE calendar_events ENABLE ROW LEVEL SECURITY;
ALTER TABLE calendar_subscriptions ENABLE ROW LEVEL SECURITY;
ALTER TABLE calendar_sync_configs ENABLE ROW LEVEL SECURITY;

CREATE POLICY calendar_events_org_isolation ON calendar_events
    USING (organization_id = current_setting('app.current_org', true)::uuid);

CREATE POLICY calendar_subscriptions_org_isolation ON calendar_subscriptions
    USING (organization_id = current_setting('app.current_org', true)::uuid);

CREATE POLICY calendar_sync_configs_org_isolation ON calendar_sync_configs
    USING (organization_id = current_setting('app.current_org', true)::uuid);

-- ============================================================
-- AUTO-REF FUNCTION: CE-YYYY-NNNNN
-- ============================================================

CREATE OR REPLACE FUNCTION generate_calendar_event_ref()
RETURNS TRIGGER AS $$
DECLARE
    yr  TEXT;
    seq INT;
BEGIN
    yr := EXTRACT(YEAR FROM NOW())::TEXT;

    SELECT COALESCE(MAX(
        CASE WHEN event_ref ~ ('^CE-' || yr || '-\d+$')
             THEN CAST(SUBSTRING(event_ref FROM '\d+$') AS INT)
             ELSE 0 END
    ), 0) + 1
    INTO seq
    FROM calendar_events
    WHERE organization_id = NEW.organization_id;

    NEW.event_ref := 'CE-' || yr || '-' || LPAD(seq::TEXT, 5, '0');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_calendar_event_ref
    BEFORE INSERT ON calendar_events
    FOR EACH ROW
    WHEN (NEW.event_ref IS NULL OR NEW.event_ref = '')
    EXECUTE FUNCTION generate_calendar_event_ref();

-- ============================================================
-- AUTO-UPDATE updated_at TRIGGERS
-- ============================================================

CREATE TRIGGER set_calendar_events_updated_at
    BEFORE UPDATE ON calendar_events
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER set_calendar_subscriptions_updated_at
    BEFORE UPDATE ON calendar_subscriptions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER set_calendar_sync_configs_updated_at
    BEFORE UPDATE ON calendar_sync_configs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================
-- INDEXES
-- ============================================================

-- calendar_events
CREATE INDEX idx_calendar_events_org ON calendar_events(organization_id);
CREATE INDEX idx_calendar_events_due_date ON calendar_events(organization_id, due_date);
CREATE INDEX idx_calendar_events_status ON calendar_events(organization_id, status);
CREATE INDEX idx_calendar_events_category ON calendar_events(organization_id, category);
CREATE INDEX idx_calendar_events_priority ON calendar_events(organization_id, priority);
CREATE INDEX idx_calendar_events_assigned ON calendar_events(assigned_to_user_id) WHERE assigned_to_user_id IS NOT NULL;
CREATE INDEX idx_calendar_events_source ON calendar_events(organization_id, source_entity_type, source_entity_id);
CREATE INDEX idx_calendar_events_overdue ON calendar_events(organization_id, due_date) WHERE status IN ('upcoming', 'due_soon');
CREATE INDEX idx_calendar_events_type ON calendar_events(organization_id, event_type);
CREATE INDEX idx_calendar_events_parent ON calendar_events(parent_event_id) WHERE parent_event_id IS NOT NULL;

-- calendar_subscriptions
CREATE INDEX idx_calendar_subscriptions_org ON calendar_subscriptions(organization_id);
CREATE INDEX idx_calendar_subscriptions_user ON calendar_subscriptions(user_id);
CREATE INDEX idx_calendar_subscriptions_ical ON calendar_subscriptions(ical_token) WHERE ical_token IS NOT NULL;

-- calendar_sync_configs
CREATE INDEX idx_calendar_sync_configs_org ON calendar_sync_configs(organization_id);
CREATE INDEX idx_calendar_sync_configs_module ON calendar_sync_configs(organization_id, module_name);
