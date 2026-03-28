-- ============================================================
-- ComplianceForge GRC Platform
-- Migration 020: Advanced Analytics, Predictive Risk Scoring
--                & BI Dashboard
-- ============================================================

-- ============================================================
-- ENUM TYPES
-- ============================================================

CREATE TYPE snapshot_type AS ENUM (
    'daily', 'weekly', 'monthly', 'quarterly'
);

CREATE TYPE trend_direction AS ENUM (
    'improving', 'stable', 'declining'
);

CREATE TYPE prediction_type AS ENUM (
    'score_forecast', 'breach_probability',
    'treatment_effectiveness', 'escalation_likelihood'
);

CREATE TYPE benchmark_type AS ENUM (
    'industry', 'size', 'region', 'framework', 'overall'
);

-- ============================================================
-- ANALYTICS SNAPSHOTS
-- Immutable point-in-time captures of organizational metrics.
-- ============================================================

CREATE TABLE analytics_snapshots (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id     UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    snapshot_type       snapshot_type NOT NULL,
    snapshot_date       DATE NOT NULL,
    metrics             JSONB NOT NULL DEFAULT '{}',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE analytics_snapshots ENABLE ROW LEVEL SECURITY;

CREATE POLICY analytics_snapshots_tenant_isolation ON analytics_snapshots
    USING (organization_id = get_current_tenant());

CREATE INDEX idx_analytics_snapshots_org_id ON analytics_snapshots(organization_id);
CREATE INDEX idx_analytics_snapshots_date ON analytics_snapshots(snapshot_date);
CREATE INDEX idx_analytics_snapshots_org_type_date ON analytics_snapshots(organization_id, snapshot_type, snapshot_date);
CREATE INDEX idx_analytics_snapshots_metrics ON analytics_snapshots USING GIN(metrics);

-- ============================================================
-- ANALYTICS COMPLIANCE TRENDS
-- Tracks compliance score movement over time per framework.
-- ============================================================

CREATE TABLE analytics_compliance_trends (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id         UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    framework_id            UUID,
    framework_code          VARCHAR(20),
    measurement_date        DATE NOT NULL,
    compliance_score        DECIMAL(5,2),
    controls_implemented    INT DEFAULT 0,
    controls_total          INT DEFAULT 0,
    maturity_avg            DECIMAL(3,2),
    score_change_7d         DECIMAL(5,2),
    score_change_30d        DECIMAL(5,2),
    score_change_90d        DECIMAL(5,2),
    trend_direction         trend_direction DEFAULT 'stable',
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(organization_id, framework_id, measurement_date)
);

ALTER TABLE analytics_compliance_trends ENABLE ROW LEVEL SECURITY;

CREATE POLICY analytics_compliance_trends_tenant_isolation ON analytics_compliance_trends
    USING (organization_id = get_current_tenant());

CREATE INDEX idx_analytics_compliance_trends_org_id ON analytics_compliance_trends(organization_id);
CREATE INDEX idx_analytics_compliance_trends_date ON analytics_compliance_trends(measurement_date);
CREATE INDEX idx_analytics_compliance_trends_org_fw_date ON analytics_compliance_trends(organization_id, framework_id, measurement_date);

-- ============================================================
-- ANALYTICS RISK PREDICTIONS
-- Statistical model outputs for risk trajectory forecasting.
-- ============================================================

CREATE TABLE analytics_risk_predictions (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id             UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    risk_id                     UUID,
    prediction_date             DATE,
    prediction_type             prediction_type NOT NULL,
    predicted_value             DECIMAL(10,4),
    confidence_interval_low     DECIMAL(10,4),
    confidence_interval_high    DECIMAL(10,4),
    confidence_level            DECIMAL(3,2) DEFAULT 0.95,
    model_version               VARCHAR(50) DEFAULT 'v1.0',
    input_features              JSONB,
    actual_value                DECIMAL(10,4),
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE analytics_risk_predictions ENABLE ROW LEVEL SECURITY;

CREATE POLICY analytics_risk_predictions_tenant_isolation ON analytics_risk_predictions
    USING (organization_id = get_current_tenant());

CREATE INDEX idx_analytics_risk_predictions_org_id ON analytics_risk_predictions(organization_id);
CREATE INDEX idx_analytics_risk_predictions_date ON analytics_risk_predictions(prediction_date);
CREATE INDEX idx_analytics_risk_predictions_risk ON analytics_risk_predictions(organization_id, risk_id, prediction_date);

-- ============================================================
-- ANALYTICS BENCHMARKS
-- Anonymized aggregate metrics across organizations for
-- peer comparison. No individual org data is exposed.
-- ============================================================

CREATE TABLE analytics_benchmarks (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    benchmark_type      benchmark_type NOT NULL,
    category            VARCHAR(100),
    metric_name         VARCHAR(200) NOT NULL,
    period              VARCHAR(20),
    percentile_25       DECIMAL(10,2),
    percentile_50       DECIMAL(10,2),
    percentile_75       DECIMAL(10,2),
    percentile_90       DECIMAL(10,2),
    sample_size         INT DEFAULT 0,
    calculated_at       TIMESTAMPTZ DEFAULT NOW(),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_analytics_benchmarks_type_metric ON analytics_benchmarks(benchmark_type, metric_name);
CREATE INDEX idx_analytics_benchmarks_period ON analytics_benchmarks(period);

-- ============================================================
-- ANALYTICS CUSTOM DASHBOARDS
-- User-created BI dashboards with configurable widget layouts.
-- ============================================================

CREATE TABLE analytics_custom_dashboards (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id     UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name                VARCHAR(200) NOT NULL,
    description         TEXT,
    layout              JSONB NOT NULL DEFAULT '[]',
    is_default          BOOLEAN DEFAULT false,
    is_shared           BOOLEAN DEFAULT false,
    owner_user_id       UUID,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE analytics_custom_dashboards ENABLE ROW LEVEL SECURITY;

CREATE POLICY analytics_custom_dashboards_tenant_isolation ON analytics_custom_dashboards
    USING (organization_id = get_current_tenant());

CREATE INDEX idx_analytics_custom_dashboards_org_id ON analytics_custom_dashboards(organization_id);
CREATE INDEX idx_analytics_custom_dashboards_owner ON analytics_custom_dashboards(owner_user_id);

CREATE TRIGGER set_analytics_dashboards_updated_at
    BEFORE UPDATE ON analytics_custom_dashboards
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================
-- ANALYTICS WIDGET TYPES
-- Registry of available widget types for dashboard builder.
-- ============================================================

CREATE TABLE analytics_widget_types (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    widget_type         VARCHAR(50) NOT NULL UNIQUE,
    name                VARCHAR(200) NOT NULL,
    description         TEXT,
    available_metrics   TEXT[],
    default_config      JSONB DEFAULT '{}',
    min_width           INT DEFAULT 1,
    min_height          INT DEFAULT 1,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_analytics_widget_types_type ON analytics_widget_types(widget_type);
