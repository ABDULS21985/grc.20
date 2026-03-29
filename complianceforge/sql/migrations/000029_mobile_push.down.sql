-- 000029_mobile_push.down.sql
-- Rollback: Mobile-Optimised API, Push Notifications & Responsive Design

-- Drop triggers first
DROP TRIGGER IF EXISTS trg_mobile_prefs_updated_at ON user_mobile_preferences;
DROP TRIGGER IF EXISTS trg_push_tokens_updated_at ON push_notification_tokens;
DROP FUNCTION IF EXISTS update_push_tokens_updated_at();

-- Drop RLS policies
DROP POLICY IF EXISTS mobile_prefs_org_isolation ON user_mobile_preferences;
DROP POLICY IF EXISTS push_log_org_isolation ON push_notification_log;
DROP POLICY IF EXISTS push_tokens_org_isolation ON push_notification_tokens;

-- Drop tables (order matters for foreign keys)
DROP TABLE IF EXISTS user_mobile_preferences;
DROP TABLE IF EXISTS push_notification_log;
DROP TABLE IF EXISTS push_notification_tokens;

-- Drop enums
DROP TYPE IF EXISTS push_delivery_status;
DROP TYPE IF EXISTS push_platform;
