-- 000029_mobile_push.up.sql
-- Mobile-Optimised API, Push Notifications & Responsive Design

-- ============================================================
-- ENUMS
-- ============================================================

CREATE TYPE push_platform AS ENUM ('ios', 'android', 'web');

CREATE TYPE push_delivery_status AS ENUM ('sent', 'delivered', 'failed', 'invalid_token');

-- ============================================================
-- TABLE: push_notification_tokens
-- Stores device push tokens for each user, across platforms.
-- ============================================================

CREATE TABLE push_notification_tokens (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    platform        push_platform NOT NULL,
    token           TEXT NOT NULL,
    token_hash      VARCHAR(128) NOT NULL,
    device_name     VARCHAR(255) DEFAULT '',
    device_model    VARCHAR(255) DEFAULT '',
    os_version      VARCHAR(64) DEFAULT '',
    app_version     VARCHAR(64) DEFAULT '',
    is_active       BOOLEAN NOT NULL DEFAULT true,
    last_used_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_push_token_hash UNIQUE (token_hash)
);

CREATE INDEX idx_push_tokens_user_active ON push_notification_tokens(user_id, is_active) WHERE is_active = true;
CREATE INDEX idx_push_tokens_org ON push_notification_tokens(organization_id);
CREATE INDEX idx_push_tokens_last_used ON push_notification_tokens(last_used_at);

-- ============================================================
-- TABLE: push_notification_log
-- Tracks every push notification sent, for audit and debugging.
-- ============================================================

CREATE TABLE push_notification_log (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    organization_id   UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    token_id          UUID REFERENCES push_notification_tokens(id) ON DELETE SET NULL,
    notification_type VARCHAR(100) NOT NULL,
    title             VARCHAR(512) NOT NULL,
    body              TEXT NOT NULL DEFAULT '',
    data              JSONB NOT NULL DEFAULT '{}',
    status            push_delivery_status NOT NULL DEFAULT 'sent',
    platform          push_platform NOT NULL,
    sent_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    delivered_at      TIMESTAMPTZ,
    error_message     TEXT,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_push_log_user ON push_notification_log(user_id, sent_at DESC);
CREATE INDEX idx_push_log_org ON push_notification_log(organization_id, sent_at DESC);
CREATE INDEX idx_push_log_status ON push_notification_log(status) WHERE status = 'failed';
CREATE INDEX idx_push_log_token ON push_notification_log(token_id);
CREATE INDEX idx_push_log_type ON push_notification_log(notification_type, sent_at DESC);

-- ============================================================
-- TABLE: user_mobile_preferences
-- Per-user push notification preferences with quiet hours.
-- push_breach_alerts is always true and enforced at app level.
-- ============================================================

CREATE TABLE user_mobile_preferences (
    id                       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id                  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    organization_id          UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    push_enabled             BOOLEAN NOT NULL DEFAULT true,
    push_breach_alerts       BOOLEAN NOT NULL DEFAULT true,
    push_approval_requests   BOOLEAN NOT NULL DEFAULT true,
    push_incident_alerts     BOOLEAN NOT NULL DEFAULT true,
    push_deadline_reminders  BOOLEAN NOT NULL DEFAULT true,
    push_mentions            BOOLEAN NOT NULL DEFAULT true,
    push_comments            BOOLEAN NOT NULL DEFAULT false,
    quiet_hours_enabled      BOOLEAN NOT NULL DEFAULT false,
    quiet_hours_start        TIME DEFAULT '22:00',
    quiet_hours_end          TIME DEFAULT '08:00',
    quiet_hours_timezone     VARCHAR(64) NOT NULL DEFAULT 'UTC',
    created_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- push_breach_alerts cannot be set to false; enforced by CHECK
    CONSTRAINT chk_breach_alerts_always_on CHECK (push_breach_alerts = true),
    CONSTRAINT uq_user_mobile_prefs UNIQUE (user_id)
);

CREATE INDEX idx_mobile_prefs_org ON user_mobile_preferences(organization_id);

-- ============================================================
-- ROW LEVEL SECURITY
-- ============================================================

ALTER TABLE push_notification_tokens ENABLE ROW LEVEL SECURITY;
ALTER TABLE push_notification_log ENABLE ROW LEVEL SECURITY;
ALTER TABLE user_mobile_preferences ENABLE ROW LEVEL SECURITY;

-- push_notification_tokens: users see only their own org's tokens
CREATE POLICY push_tokens_org_isolation ON push_notification_tokens
    USING (organization_id = current_setting('app.current_org', true)::uuid);

-- push_notification_log: org isolation
CREATE POLICY push_log_org_isolation ON push_notification_log
    USING (organization_id = current_setting('app.current_org', true)::uuid);

-- user_mobile_preferences: org isolation
CREATE POLICY mobile_prefs_org_isolation ON user_mobile_preferences
    USING (organization_id = current_setting('app.current_org', true)::uuid);

-- ============================================================
-- TRIGGER: auto-update updated_at
-- ============================================================

CREATE OR REPLACE FUNCTION update_push_tokens_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_push_tokens_updated_at
    BEFORE UPDATE ON push_notification_tokens
    FOR EACH ROW EXECUTE FUNCTION update_push_tokens_updated_at();

CREATE TRIGGER trg_mobile_prefs_updated_at
    BEFORE UPDATE ON user_mobile_preferences
    FOR EACH ROW EXECUTE FUNCTION update_push_tokens_updated_at();
