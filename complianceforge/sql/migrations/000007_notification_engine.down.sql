-- ============================================================
-- ComplianceForge GRC Platform
-- Migration 007 (Down): Drop Enterprise Notification Engine
-- ============================================================

-- Drop RLS policies
DROP POLICY IF EXISTS rls_notification_templates ON notification_templates;
DROP POLICY IF EXISTS rls_notification_preferences ON notification_preferences;
DROP POLICY IF EXISTS rls_notifications ON notifications;
DROP POLICY IF EXISTS rls_notification_rules ON notification_rules;
DROP POLICY IF EXISTS rls_notification_channels ON notification_channels;

-- Drop tables in dependency order
DROP TABLE IF EXISTS notification_preferences;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS notification_rules;
DROP TABLE IF EXISTS notification_channels;
DROP TABLE IF EXISTS notification_templates;

-- Drop enum types
DROP TYPE IF EXISTS digest_frequency;
DROP TYPE IF EXISTS notification_recipient_type;
DROP TYPE IF EXISTS notification_status;
DROP TYPE IF EXISTS notification_channel_type;
