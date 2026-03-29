-- 000028_collaboration.down.sql
-- Rollback Collaboration, Comments, Mentions & Activity Feed

DROP TRIGGER IF EXISTS trg_read_marker_updated_at ON user_read_markers;
DROP FUNCTION IF EXISTS update_read_marker_updated_at();

DROP TRIGGER IF EXISTS trg_comment_updated_at ON comments;
DROP FUNCTION IF EXISTS update_comment_updated_at();

DROP TABLE IF EXISTS user_read_markers CASCADE;
DROP TABLE IF EXISTS user_follows CASCADE;
DROP TABLE IF EXISTS activity_feed CASCADE;
DROP TABLE IF EXISTS comments CASCADE;

DROP TYPE IF EXISTS reaction_type;
DROP TYPE IF EXISTS follow_type;
DROP TYPE IF EXISTS activity_visibility;
DROP TYPE IF EXISTS activity_action;
DROP TYPE IF EXISTS comment_visibility;
