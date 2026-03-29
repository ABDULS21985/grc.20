-- 000026_compliance_calendar.down.sql
-- Rollback: Compliance Calendar & Deadline Management Engine

-- Drop triggers
DROP TRIGGER IF EXISTS set_calendar_sync_configs_updated_at ON calendar_sync_configs;
DROP TRIGGER IF EXISTS set_calendar_subscriptions_updated_at ON calendar_subscriptions;
DROP TRIGGER IF EXISTS set_calendar_events_updated_at ON calendar_events;
DROP TRIGGER IF EXISTS trg_calendar_event_ref ON calendar_events;

-- Drop functions
DROP FUNCTION IF EXISTS generate_calendar_event_ref();

-- Drop tables (order matters for FK constraints)
DROP TABLE IF EXISTS calendar_sync_configs;
DROP TABLE IF EXISTS calendar_subscriptions;
DROP TABLE IF EXISTS calendar_events;

-- Drop enums
DROP TYPE IF EXISTS calendar_recurrence_type;
DROP TYPE IF EXISTS calendar_event_status;
DROP TYPE IF EXISTS calendar_event_priority;
DROP TYPE IF EXISTS calendar_event_category;
DROP TYPE IF EXISTS calendar_event_type;
