-- ============================================================
-- ComplianceForge GRC Platform
-- Migration 006: Advanced Views, Functions & Dashboard Support
-- ============================================================

-- ============================================================
-- CROSS-FRAMEWORK COVERAGE VIEW
-- Shows how implementing one framework covers another
-- ============================================================

CREATE OR REPLACE VIEW v_cross_framework_coverage AS
SELECT
    sf.code AS source_framework,
    sf.name AS source_framework_name,
    tf.code AS target_framework,
    tf.name AS target_framework_name,
    COUNT(*) AS total_mappings,
    COUNT(*) FILTER (WHERE m.mapping_type = 'equivalent') AS equivalent_count,
    COUNT(*) FILTER (WHERE m.mapping_type = 'partial') AS partial_count,
    COUNT(*) FILTER (WHERE m.mapping_type = 'related') AS related_count,
    ROUND(AVG(m.mapping_strength)::DECIMAL, 2) AS avg_strength,
    ROUND(
        (COUNT(*) FILTER (WHERE m.mapping_type IN ('equivalent', 'partial'))::DECIMAL / 
        NULLIF(COUNT(*), 0)) * 100, 1
    ) AS coverage_percentage
FROM framework_control_mappings m
JOIN framework_controls sc ON m.source_control_id = sc.id
JOIN framework_controls tc ON m.target_control_id = tc.id
JOIN compliance_frameworks sf ON sc.framework_id = sf.id
JOIN compliance_frameworks tf ON tc.framework_id = tf.id
GROUP BY sf.code, sf.name, tf.code, tf.name
ORDER BY sf.code, tf.code;

-- ============================================================
-- POLICY COMPLIANCE STATUS VIEW
-- ============================================================

CREATE OR REPLACE VIEW v_policy_compliance_status AS
SELECT
    p.organization_id,
    COUNT(*) AS total_policies,
    COUNT(*) FILTER (WHERE p.status = 'published') AS published,
    COUNT(*) FILTER (WHERE p.status = 'draft') AS draft,
    COUNT(*) FILTER (WHERE p.status = 'under_review') AS under_review,
    COUNT(*) FILTER (WHERE p.status = 'archived') AS archived,
    COUNT(*) FILTER (WHERE p.review_status = 'overdue') AS reviews_overdue,
    COUNT(*) FILTER (WHERE p.next_review_date BETWEEN CURRENT_DATE AND CURRENT_DATE + INTERVAL '30 days') AS reviews_due_30_days,
    ROUND(
        (COUNT(*) FILTER (WHERE p.status = 'published')::DECIMAL / NULLIF(COUNT(*), 0)) * 100, 1
    ) AS publish_rate
FROM policies p
WHERE p.deleted_at IS NULL
GROUP BY p.organization_id;

-- ============================================================
-- INCIDENT METRICS VIEW
-- ============================================================

CREATE OR REPLACE VIEW v_incident_metrics AS
SELECT
    i.organization_id,
    COUNT(*) AS total_incidents,
    COUNT(*) FILTER (WHERE i.status NOT IN ('resolved', 'closed')) AS open_incidents,
    COUNT(*) FILTER (WHERE i.severity = 'critical' AND i.status NOT IN ('resolved', 'closed')) AS critical_open,
    COUNT(*) FILTER (WHERE i.is_data_breach = true) AS total_breaches,
    COUNT(*) FILTER (WHERE i.is_data_breach = true AND i.dpa_notified_at IS NULL AND i.notification_required = true) AS breaches_pending_notification,
    COUNT(*) FILTER (WHERE i.is_nis2_reportable = true) AS nis2_reportable,
    COUNT(*) FILTER (WHERE i.is_nis2_reportable = true AND i.nis2_early_warning_at IS NULL) AS nis2_pending_early_warning,
    COALESCE(AVG(EXTRACT(EPOCH FROM (i.resolved_at - i.reported_at)) / 3600)
        FILTER (WHERE i.resolved_at IS NOT NULL), 0)::DECIMAL(10,1) AS avg_resolution_hours,
    COALESCE(AVG(EXTRACT(EPOCH FROM (i.contained_at - i.reported_at)) / 3600)
        FILTER (WHERE i.contained_at IS NOT NULL), 0)::DECIMAL(10,1) AS avg_containment_hours
FROM incidents i
WHERE i.deleted_at IS NULL
GROUP BY i.organization_id;

-- ============================================================
-- VENDOR RISK SUMMARY VIEW
-- ============================================================

CREATE OR REPLACE VIEW v_vendor_risk_summary AS
SELECT
    v.organization_id,
    COUNT(*) AS total_vendors,
    COUNT(*) FILTER (WHERE v.status = 'active') AS active_vendors,
    COUNT(*) FILTER (WHERE v.risk_tier = 'critical') AS critical_risk,
    COUNT(*) FILTER (WHERE v.risk_tier = 'high') AS high_risk,
    COUNT(*) FILTER (WHERE v.risk_tier = 'medium') AS medium_risk,
    COUNT(*) FILTER (WHERE v.risk_tier = 'low') AS low_risk,
    COUNT(*) FILTER (WHERE v.data_processing = true) AS data_processors,
    COUNT(*) FILTER (WHERE v.data_processing = true AND v.dpa_in_place = false) AS missing_dpa,
    COUNT(*) FILTER (WHERE v.next_assessment_date < CURRENT_DATE AND v.status = 'active') AS assessments_overdue,
    COALESCE(SUM(v.contract_value), 0) AS total_contract_value
FROM vendors v
WHERE v.deleted_at IS NULL
GROUP BY v.organization_id;

-- ============================================================
-- AUDIT FINDINGS SUMMARY VIEW
-- ============================================================

CREATE OR REPLACE VIEW v_audit_findings_summary AS
SELECT
    af.organization_id,
    COUNT(*) AS total_findings,
    COUNT(*) FILTER (WHERE af.status IN ('open', 'in_progress')) AS open_findings,
    COUNT(*) FILTER (WHERE af.status = 'open' AND af.due_date < CURRENT_DATE) AS overdue_findings,
    COUNT(*) FILTER (WHERE af.severity = 'critical' AND af.status NOT IN ('remediated', 'closed', 'accepted')) AS critical_open,
    COUNT(*) FILTER (WHERE af.severity = 'high' AND af.status NOT IN ('remediated', 'closed', 'accepted')) AS high_open,
    COUNT(*) FILTER (WHERE af.status = 'remediated') AS remediated,
    COUNT(*) FILTER (WHERE af.status = 'closed') AS closed,
    COUNT(*) FILTER (WHERE af.finding_type = 'non_conformity') AS non_conformities,
    COUNT(*) FILTER (WHERE af.finding_type = 'observation') AS observations
FROM audit_findings af
WHERE af.deleted_at IS NULL
GROUP BY af.organization_id;

-- ============================================================
-- RISK TREATMENT PROGRESS VIEW
-- ============================================================

CREATE OR REPLACE VIEW v_risk_treatment_progress AS
SELECT
    rt.organization_id,
    COUNT(*) AS total_treatments,
    COUNT(*) FILTER (WHERE rt.status = 'completed') AS completed,
    COUNT(*) FILTER (WHERE rt.status = 'in_progress') AS in_progress,
    COUNT(*) FILTER (WHERE rt.status = 'planned') AS planned,
    COUNT(*) FILTER (WHERE rt.status NOT IN ('completed', 'cancelled') AND rt.target_date < CURRENT_DATE) AS overdue,
    COALESCE(SUM(rt.estimated_cost_eur), 0) AS total_estimated_cost,
    COALESCE(SUM(rt.actual_cost_eur), 0) AS total_actual_cost,
    ROUND(
        (COUNT(*) FILTER (WHERE rt.status = 'completed')::DECIMAL / NULLIF(COUNT(*), 0)) * 100, 1
    ) AS completion_rate
FROM risk_treatments rt
GROUP BY rt.organization_id;

-- ============================================================
-- KRI DASHBOARD VIEW
-- ============================================================

CREATE OR REPLACE VIEW v_kri_dashboard AS
SELECT
    ri.organization_id,
    ri.id AS indicator_id,
    ri.name,
    ri.metric_type,
    ri.current_value,
    ri.threshold_green,
    ri.threshold_amber,
    ri.threshold_red,
    ri.trend,
    ri.last_updated_at,
    ri.collection_frequency,
    CASE
        WHEN ri.current_value <= ri.threshold_green THEN 'green'
        WHEN ri.current_value <= ri.threshold_amber THEN 'amber'
        ELSE 'red'
    END AS status,
    r.risk_ref,
    r.title AS risk_title,
    COALESCE(u.first_name || ' ' || u.last_name, 'Unassigned') AS owner_name
FROM risk_indicators ri
LEFT JOIN risks r ON ri.risk_id = r.id
LEFT JOIN users u ON ri.owner_user_id = u.id
ORDER BY
    CASE
        WHEN ri.current_value > ri.threshold_red THEN 1
        WHEN ri.current_value > ri.threshold_amber THEN 2
        ELSE 3
    END,
    ri.name;

-- ============================================================
-- COMPLIANCE MATURITY OVER TIME (Aggregate Function)
-- ============================================================

CREATE OR REPLACE FUNCTION calculate_org_maturity(p_org_id UUID)
RETURNS TABLE (
    framework_code VARCHAR,
    framework_name VARCHAR,
    avg_maturity DECIMAL,
    maturity_label VARCHAR,
    total_controls BIGINT,
    assessed_controls BIGINT
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        cf.code::VARCHAR,
        cf.name::VARCHAR,
        ROUND(AVG(ci.maturity_level)::DECIMAL, 2) AS avg_maturity,
        CASE
            WHEN AVG(ci.maturity_level) >= 4.5 THEN 'Optimizing'
            WHEN AVG(ci.maturity_level) >= 3.5 THEN 'Quantitatively Managed'
            WHEN AVG(ci.maturity_level) >= 2.5 THEN 'Defined'
            WHEN AVG(ci.maturity_level) >= 1.5 THEN 'Managed'
            WHEN AVG(ci.maturity_level) >= 0.5 THEN 'Initial'
            ELSE 'Non-existent'
        END::VARCHAR AS maturity_label,
        COUNT(*)::BIGINT AS total_controls,
        COUNT(*) FILTER (WHERE ci.maturity_level > 0)::BIGINT AS assessed_controls
    FROM control_implementations ci
    JOIN organization_frameworks of2 ON ci.org_framework_id = of2.id
    JOIN compliance_frameworks cf ON of2.framework_id = cf.id
    WHERE ci.organization_id = p_org_id
        AND ci.status != 'not_applicable'
        AND ci.deleted_at IS NULL
    GROUP BY cf.code, cf.name
    ORDER BY cf.code;
END;
$$ LANGUAGE plpgsql STABLE;

-- ============================================================
-- EXECUTIVE DASHBOARD AGGREGATE FUNCTION
-- ============================================================

CREATE OR REPLACE FUNCTION get_executive_dashboard(p_org_id UUID)
RETURNS JSONB AS $$
DECLARE
    result JSONB;
BEGIN
    SELECT jsonb_build_object(
        'compliance', (
            SELECT jsonb_build_object(
                'overall_score', COALESCE(ROUND(AVG(compliance_score)::DECIMAL, 1), 0),
                'frameworks_adopted', COUNT(*),
                'highest_score', COALESCE(MAX(compliance_score), 0),
                'lowest_score', COALESCE(MIN(compliance_score), 0)
            )
            FROM v_compliance_score_by_framework
            WHERE organization_id = p_org_id
        ),
        'risks', (
            SELECT jsonb_build_object(
                'total', COUNT(*),
                'critical', COUNT(*) FILTER (WHERE residual_risk_level = 'critical'),
                'high', COUNT(*) FILTER (WHERE residual_risk_level = 'high'),
                'medium', COUNT(*) FILTER (WHERE residual_risk_level = 'medium'),
                'low', COUNT(*) FILTER (WHERE residual_risk_level = 'low')
            )
            FROM risks
            WHERE organization_id = p_org_id AND deleted_at IS NULL
        ),
        'incidents', (
            SELECT row_to_json(v.*)::JSONB
            FROM v_incident_metrics v
            WHERE v.organization_id = p_org_id
        ),
        'vendors', (
            SELECT row_to_json(v.*)::JSONB
            FROM v_vendor_risk_summary v
            WHERE v.organization_id = p_org_id
        ),
        'findings', (
            SELECT row_to_json(v.*)::JSONB
            FROM v_audit_findings_summary v
            WHERE v.organization_id = p_org_id
        ),
        'policies', (
            SELECT row_to_json(v.*)::JSONB
            FROM v_policy_compliance_status v
            WHERE v.organization_id = p_org_id
        ),
        'treatments', (
            SELECT row_to_json(v.*)::JSONB
            FROM v_risk_treatment_progress v
            WHERE v.organization_id = p_org_id
        )
    ) INTO result;

    RETURN COALESCE(result, '{}'::JSONB);
END;
$$ LANGUAGE plpgsql STABLE;

-- ============================================================
-- AUTO-CREATE AUDIT LOG PARTITIONS (12 months ahead)
-- ============================================================

CREATE OR REPLACE FUNCTION create_audit_log_partitions()
RETURNS VOID AS $$
DECLARE
    start_date DATE;
    end_date DATE;
    partition_name TEXT;
BEGIN
    FOR i IN 0..11 LOOP
        start_date := DATE_TRUNC('month', NOW() + (i || ' months')::INTERVAL);
        end_date := start_date + INTERVAL '1 month';
        partition_name := 'audit_logs_' || TO_CHAR(start_date, 'YYYY_MM');

        BEGIN
            EXECUTE format(
                'CREATE TABLE IF NOT EXISTS %I PARTITION OF audit_logs FOR VALUES FROM (%L) TO (%L)',
                partition_name, start_date, end_date
            );
        EXCEPTION WHEN duplicate_table THEN
            -- Partition already exists, skip
            NULL;
        END;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Create partitions for 2027
SELECT create_audit_log_partitions();

-- ============================================================
-- COMPLIANCE SCORE RECALCULATION TRIGGER
-- When a control_implementation status changes, recalculate the framework score
-- ============================================================

CREATE OR REPLACE FUNCTION recalculate_framework_score()
RETURNS TRIGGER AS $$
DECLARE
    v_score DECIMAL(5,2);
    v_total INT;
    v_implemented INT;
    v_not_applicable INT;
BEGIN
    SELECT
        COUNT(*),
        COUNT(*) FILTER (WHERE status IN ('implemented', 'effective')),
        COUNT(*) FILTER (WHERE status = 'not_applicable')
    INTO v_total, v_implemented, v_not_applicable
    FROM control_implementations
    WHERE org_framework_id = NEW.org_framework_id AND deleted_at IS NULL;

    IF (v_total - v_not_applicable) > 0 THEN
        v_score := ROUND((v_implemented::DECIMAL / (v_total - v_not_applicable)) * 100, 2);
    ELSE
        v_score := 0;
    END IF;

    UPDATE organization_frameworks
    SET compliance_score = v_score, last_assessment_date = NOW()
    WHERE id = NEW.org_framework_id;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_recalculate_score
    AFTER INSERT OR UPDATE OF status ON control_implementations
    FOR EACH ROW
    EXECUTE FUNCTION recalculate_framework_score();
