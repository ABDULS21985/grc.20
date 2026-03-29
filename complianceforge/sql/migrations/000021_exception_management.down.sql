-- 000021_exception_management.down.sql
-- Rollback Exception Management & Compensating Controls

DROP TRIGGER IF EXISTS trg_exception_updated_at ON compliance_exceptions;
DROP FUNCTION IF EXISTS update_exception_updated_at();

DROP TRIGGER IF EXISTS trg_exception_ref ON compliance_exceptions;
DROP FUNCTION IF EXISTS generate_exception_ref();

DROP TABLE IF EXISTS exception_audit_trail CASCADE;
DROP TABLE IF EXISTS exception_reviews CASCADE;
DROP TABLE IF EXISTS compliance_exceptions CASCADE;

DROP TYPE IF EXISTS compensating_effectiveness;
DROP TYPE IF EXISTS exception_review_outcome;
DROP TYPE IF EXISTS exception_review_type;
DROP TYPE IF EXISTS exception_scope_type;
DROP TYPE IF EXISTS exception_status;
DROP TYPE IF EXISTS exception_type;
