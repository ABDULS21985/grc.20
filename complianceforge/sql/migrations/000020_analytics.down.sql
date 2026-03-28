-- ============================================================
-- ComplianceForge GRC Platform
-- Migration 020 (DOWN): Drop Analytics, Predictive Risk
--                        Scoring & BI Dashboard
-- ============================================================

-- Drop triggers
DROP TRIGGER IF EXISTS set_analytics_dashboards_updated_at ON analytics_custom_dashboards;

-- Drop tables (reverse order of creation)
DROP TABLE IF EXISTS analytics_widget_types CASCADE;
DROP TABLE IF EXISTS analytics_custom_dashboards CASCADE;
DROP TABLE IF EXISTS analytics_benchmarks CASCADE;
DROP TABLE IF EXISTS analytics_risk_predictions CASCADE;
DROP TABLE IF EXISTS analytics_compliance_trends CASCADE;
DROP TABLE IF EXISTS analytics_snapshots CASCADE;

-- Drop enum types
DROP TYPE IF EXISTS benchmark_type;
DROP TYPE IF EXISTS prediction_type;
DROP TYPE IF EXISTS trend_direction;
DROP TYPE IF EXISTS snapshot_type;
