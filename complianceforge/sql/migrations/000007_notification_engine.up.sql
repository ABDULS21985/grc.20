-- ============================================================
-- ComplianceForge GRC Platform
-- Migration 007: Enterprise Notification Engine
-- ============================================================

-- ============================================================
-- ENUM TYPES
-- ============================================================

CREATE TYPE notification_channel_type AS ENUM ('email', 'in_app', 'slack', 'webhook', 'sms');
CREATE TYPE notification_status AS ENUM ('pending', 'sent', 'delivered', 'failed', 'bounced');
CREATE TYPE notification_recipient_type AS ENUM ('user', 'role', 'team', 'all_admins', 'all_users', 'custom');
CREATE TYPE digest_frequency AS ENUM ('realtime', 'hourly', 'daily', 'weekly', 'none');

-- ============================================================
-- NOTIFICATION TEMPLATES
-- ============================================================

CREATE TABLE notification_templates (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    name            VARCHAR(255) NOT NULL,
    event_type      VARCHAR(100) NOT NULL,
    subject_template        TEXT NOT NULL DEFAULT '',
    body_html_template      TEXT NOT NULL DEFAULT '',
    body_text_template      TEXT NOT NULL DEFAULT '',
    in_app_title_template   TEXT NOT NULL DEFAULT '',
    in_app_body_template    TEXT NOT NULL DEFAULT '',
    slack_template          JSONB NOT NULL DEFAULT '{}',
    webhook_payload_template JSONB NOT NULL DEFAULT '{}',
    variables       TEXT[] NOT NULL DEFAULT '{}',
    is_system       BOOLEAN NOT NULL DEFAULT false,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notification_templates_org ON notification_templates(organization_id) WHERE organization_id IS NOT NULL;
CREATE INDEX idx_notification_templates_event ON notification_templates(event_type);
CREATE INDEX idx_notification_templates_system ON notification_templates(is_system) WHERE is_system = true;

CREATE TRIGGER trg_notification_templates_updated
    BEFORE UPDATE ON notification_templates
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================
-- NOTIFICATION CHANNELS
-- ============================================================

CREATE TABLE notification_channels (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    channel_type    notification_channel_type NOT NULL,
    name            VARCHAR(255) NOT NULL,
    configuration   JSONB NOT NULL DEFAULT '{}',
    is_active       BOOLEAN NOT NULL DEFAULT true,
    is_default      BOOLEAN NOT NULL DEFAULT false,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX idx_notification_channels_org ON notification_channels(organization_id);
CREATE INDEX idx_notification_channels_type ON notification_channels(organization_id, channel_type);
CREATE INDEX idx_notification_channels_active ON notification_channels(organization_id) WHERE is_active = true AND deleted_at IS NULL;

CREATE TRIGGER trg_notification_channels_updated
    BEFORE UPDATE ON notification_channels
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================
-- NOTIFICATION RULES
-- ============================================================

CREATE TABLE notification_rules (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id         UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name                    VARCHAR(255) NOT NULL,
    event_type              VARCHAR(100) NOT NULL,
    severity_filter         TEXT[] NOT NULL DEFAULT '{}',
    conditions              JSONB NOT NULL DEFAULT '{}',
    channel_ids             UUID[] NOT NULL DEFAULT '{}',
    recipient_type          notification_recipient_type NOT NULL DEFAULT 'all_admins',
    recipient_ids           UUID[] NOT NULL DEFAULT '{}',
    template_id             UUID REFERENCES notification_templates(id) ON DELETE SET NULL,
    is_active               BOOLEAN NOT NULL DEFAULT true,
    cooldown_minutes        INT NOT NULL DEFAULT 0,
    escalation_after_minutes INT NOT NULL DEFAULT 0,
    escalation_channel_ids  UUID[] NOT NULL DEFAULT '{}',
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notification_rules_org ON notification_rules(organization_id);
CREATE INDEX idx_notification_rules_event ON notification_rules(organization_id, event_type);
CREATE INDEX idx_notification_rules_active ON notification_rules(organization_id) WHERE is_active = true;

CREATE TRIGGER trg_notification_rules_updated
    BEFORE UPDATE ON notification_rules
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================
-- NOTIFICATIONS (delivery log)
-- ============================================================

CREATE TABLE notifications (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id     UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    rule_id             UUID REFERENCES notification_rules(id) ON DELETE SET NULL,
    event_type          VARCHAR(100) NOT NULL,
    event_payload       JSONB NOT NULL DEFAULT '{}',
    recipient_user_id   UUID REFERENCES users(id) ON DELETE CASCADE,
    channel_type        notification_channel_type NOT NULL,
    channel_id          UUID REFERENCES notification_channels(id) ON DELETE SET NULL,
    subject             TEXT NOT NULL DEFAULT '',
    body                TEXT NOT NULL DEFAULT '',
    status              notification_status NOT NULL DEFAULT 'pending',
    sent_at             TIMESTAMPTZ,
    delivered_at        TIMESTAMPTZ,
    read_at             TIMESTAMPTZ,
    acknowledged_at     TIMESTAMPTZ,
    error_message       TEXT,
    retry_count         INT NOT NULL DEFAULT 0,
    max_retries         INT NOT NULL DEFAULT 3,
    next_retry_at       TIMESTAMPTZ,
    metadata            JSONB NOT NULL DEFAULT '{}',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_org ON notifications(organization_id);
CREATE INDEX idx_notifications_user ON notifications(recipient_user_id, created_at DESC);
CREATE INDEX idx_notifications_user_unread ON notifications(recipient_user_id)
    WHERE channel_type = 'in_app' AND read_at IS NULL;
CREATE INDEX idx_notifications_status ON notifications(status) WHERE status IN ('pending', 'failed');
CREATE INDEX idx_notifications_retry ON notifications(next_retry_at) WHERE status = 'failed' AND retry_count < max_retries;
CREATE INDEX idx_notifications_event ON notifications(organization_id, event_type, created_at DESC);

-- ============================================================
-- NOTIFICATION PREFERENCES (per user per event)
-- ============================================================

CREATE TABLE notification_preferences (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    organization_id     UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    event_type          VARCHAR(100) NOT NULL,
    email_enabled       BOOLEAN NOT NULL DEFAULT true,
    in_app_enabled      BOOLEAN NOT NULL DEFAULT true,
    slack_enabled       BOOLEAN NOT NULL DEFAULT false,
    digest_frequency    digest_frequency NOT NULL DEFAULT 'realtime',
    quiet_hours_start   TIME,
    quiet_hours_end     TIME,
    quiet_hours_timezone VARCHAR(50) NOT NULL DEFAULT 'UTC',
    UNIQUE(user_id, event_type)
);

CREATE INDEX idx_notification_prefs_user ON notification_preferences(user_id);
CREATE INDEX idx_notification_prefs_org ON notification_preferences(organization_id);

-- ============================================================
-- ROW-LEVEL SECURITY POLICIES
-- ============================================================

-- notification_channels
ALTER TABLE notification_channels ENABLE ROW LEVEL SECURITY;
CREATE POLICY rls_notification_channels ON notification_channels
    USING (organization_id = get_current_tenant());

-- notification_rules
ALTER TABLE notification_rules ENABLE ROW LEVEL SECURITY;
CREATE POLICY rls_notification_rules ON notification_rules
    USING (organization_id = get_current_tenant());

-- notifications
ALTER TABLE notifications ENABLE ROW LEVEL SECURITY;
CREATE POLICY rls_notifications ON notifications
    USING (organization_id = get_current_tenant());

-- notification_preferences
ALTER TABLE notification_preferences ENABLE ROW LEVEL SECURITY;
CREATE POLICY rls_notification_preferences ON notification_preferences
    USING (organization_id = get_current_tenant());

-- notification_templates: system templates (organization_id IS NULL) are visible to all;
-- org-specific templates are restricted to that org.
ALTER TABLE notification_templates ENABLE ROW LEVEL SECURITY;
CREATE POLICY rls_notification_templates ON notification_templates
    USING (organization_id IS NULL OR organization_id = get_current_tenant());
