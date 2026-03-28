-- ============================================================
-- ComplianceForge GRC Platform
-- Migration 018 DOWN: Regulatory Change Management & Horizon Scanning
-- Rollback: Remove all regulatory change tables, functions, and types.
-- ============================================================

BEGIN;

-- Drop triggers
DROP TRIGGER IF EXISTS trg_regulatory_impact_assessments_updated_at ON regulatory_impact_assessments;
DROP TRIGGER IF EXISTS trg_regulatory_changes_updated_at ON regulatory_changes;
DROP TRIGGER IF EXISTS trg_regulatory_sources_updated_at ON regulatory_sources;

-- Drop tables in FK-safe order
DROP TABLE IF EXISTS regulatory_impact_assessments CASCADE;
DROP TABLE IF EXISTS regulatory_subscriptions CASCADE;
DROP TABLE IF EXISTS regulatory_changes CASCADE;
DROP TABLE IF EXISTS regulatory_sources CASCADE;

-- Drop functions
DROP FUNCTION IF EXISTS generate_change_ref();

-- Drop enum types
DROP TYPE IF EXISTS impact_assessment_status;
DROP TYPE IF EXISTS regulatory_impact_level;
DROP TYPE IF EXISTS regulatory_change_status;
DROP TYPE IF EXISTS regulatory_severity;
DROP TYPE IF EXISTS regulatory_change_type;
DROP TYPE IF EXISTS scan_frequency;
DROP TYPE IF EXISTS regulatory_source_type;

COMMIT;
