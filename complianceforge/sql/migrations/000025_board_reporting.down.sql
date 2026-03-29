-- 000025_board_reporting.down.sql
-- Rollback: Executive Board Reporting Portal & Governance Dashboards

-- Drop triggers
DROP TRIGGER IF EXISTS set_board_decisions_updated_at ON board_decisions;
DROP TRIGGER IF EXISTS set_board_meetings_updated_at ON board_meetings;
DROP TRIGGER IF EXISTS set_board_members_updated_at ON board_members;
DROP TRIGGER IF EXISTS trg_board_decision_ref ON board_decisions;
DROP TRIGGER IF EXISTS trg_board_meeting_ref ON board_meetings;

-- Drop functions
DROP FUNCTION IF EXISTS generate_board_decision_ref();
DROP FUNCTION IF EXISTS generate_board_meeting_ref();

-- Drop tables (order matters for FK constraints)
DROP TABLE IF EXISTS board_reports;
DROP TABLE IF EXISTS board_decisions;
DROP TABLE IF EXISTS board_meetings;
DROP TABLE IF EXISTS board_members;

-- Drop enums
DROP TYPE IF EXISTS board_report_format;
DROP TYPE IF EXISTS board_report_type;
DROP TYPE IF EXISTS board_action_status;
DROP TYPE IF EXISTS board_decision_outcome;
DROP TYPE IF EXISTS board_decision_type;
DROP TYPE IF EXISTS board_meeting_status;
DROP TYPE IF EXISTS board_meeting_type;
DROP TYPE IF EXISTS board_member_type;
