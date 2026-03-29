-- ============================================================
-- ComplianceForge GRC Platform
-- Migration 027 DOWN: Drop Search & Knowledge Base tables
-- ============================================================

DROP TABLE IF EXISTS knowledge_article_feedback CASCADE;
DROP TABLE IF EXISTS recent_searches CASCADE;
DROP TABLE IF EXISTS knowledge_bookmarks CASCADE;
DROP TABLE IF EXISTS knowledge_articles CASCADE;
DROP TABLE IF EXISTS search_index CASCADE;

DROP FUNCTION IF EXISTS knowledge_articles_update_vector() CASCADE;
DROP FUNCTION IF EXISTS search_index_update_vector() CASCADE;

DROP TYPE IF EXISTS article_difficulty CASCADE;
DROP TYPE IF EXISTS article_type CASCADE;
