-- ============================================================
-- Seed 024: Analytics Widget Types
-- Defines the available widget types for the BI dashboard builder.
-- ============================================================

INSERT INTO analytics_widget_types (widget_type, name, description, available_metrics, default_config, min_width, min_height)
VALUES
(
    'line_chart',
    'Line Chart',
    'Displays metric trends over time as a continuous line graph. Ideal for compliance scores, risk trajectories, and incident volume tracking.',
    ARRAY['compliance_score', 'risk_score', 'incident_count', 'finding_count', 'control_coverage', 'maturity_score', 'vendor_risk_avg', 'policy_compliance_rate', 'dsr_response_time', 'evidence_freshness'],
    '{"show_legend": true, "show_grid": true, "line_style": "smooth", "fill_area": false, "time_range": "6m", "granularity": "weekly"}',
    4,
    3
),
(
    'bar_chart',
    'Bar Chart',
    'Compares values across categories using horizontal or vertical bars. Useful for framework comparisons, risk distribution by category, and departmental benchmarking.',
    ARRAY['compliance_score', 'risk_count', 'incident_count', 'finding_count', 'control_count', 'vendor_count', 'policy_count', 'maturity_score'],
    '{"orientation": "vertical", "show_values": true, "show_legend": true, "stacked": false, "group_by": "framework"}',
    3,
    3
),
(
    'donut_chart',
    'Donut Chart',
    'Shows proportional distribution of a metric across categories. Best for risk severity distribution, control status breakdown, and incident type composition.',
    ARRAY['risk_distribution', 'control_status_distribution', 'incident_type_distribution', 'finding_severity_distribution', 'vendor_tier_distribution', 'policy_status_distribution'],
    '{"show_legend": true, "show_percentages": true, "inner_radius": 60, "animate": true}',
    3,
    3
),
(
    'kpi_card',
    'KPI Card',
    'Single-value metric display with trend indicator and optional sparkline. Perfect for executive dashboards showing headline numbers like overall compliance score, total open risks, and SLA adherence.',
    ARRAY['compliance_score', 'risk_count', 'incident_count', 'finding_count', 'control_coverage', 'maturity_score', 'vendor_risk_avg', 'policy_compliance_rate', 'dsr_response_time', 'breach_probability', 'overdue_actions'],
    '{"show_trend": true, "show_sparkline": true, "trend_period": "30d", "format": "number", "threshold_warning": 60, "threshold_danger": 40}',
    2,
    2
),
(
    'heatmap',
    'Heatmap',
    'Grid visualization using color intensity to represent values across two dimensions. Ideal for risk matrices (likelihood vs impact), framework coverage maps, and temporal activity analysis.',
    ARRAY['risk_heatmap', 'control_coverage_map', 'incident_frequency_map', 'compliance_gap_map', 'vendor_assessment_map'],
    '{"color_scheme": "red_green", "show_values": true, "x_axis": "likelihood", "y_axis": "impact", "cell_size": "auto"}',
    4,
    4
),
(
    'radar',
    'Radar Chart',
    'Multi-axis chart comparing performance across multiple dimensions simultaneously. Best for peer benchmarking, multi-framework maturity comparison, and holistic security posture assessment.',
    ARRAY['compliance_score', 'maturity_score', 'risk_score', 'control_coverage', 'incident_response_time', 'vendor_risk_avg', 'evidence_freshness', 'policy_compliance_rate'],
    '{"show_legend": true, "fill_opacity": 0.2, "max_value": 100, "show_scale": true, "comparison_mode": "peer"}',
    4,
    4
),
(
    'table',
    'Data Table',
    'Sortable, filterable tabular view of detailed data records. Suitable for risk registers, control implementation status, audit findings, and any detailed drill-down view.',
    ARRAY['risk_register', 'control_implementations', 'audit_findings', 'incident_log', 'vendor_assessments', 'policy_reviews', 'compliance_gaps', 'action_items'],
    '{"page_size": 10, "sortable": true, "filterable": true, "show_search": true, "compact": false, "striped": true}',
    6,
    4
),
(
    'gauge',
    'Gauge',
    'Circular progress indicator showing a single value against a maximum. Ideal for overall compliance score, SLA adherence percentage, and risk appetite utilization.',
    ARRAY['compliance_score', 'control_coverage', 'maturity_score', 'sla_adherence', 'risk_appetite_utilization', 'policy_compliance_rate', 'training_completion'],
    '{"min": 0, "max": 100, "thresholds": [{"value": 40, "color": "red"}, {"value": 70, "color": "amber"}, {"value": 100, "color": "green"}], "show_value": true, "show_label": true}',
    2,
    2
),
(
    'sparkline',
    'Sparkline',
    'Compact inline trend visualization showing directional movement without axis labels. Designed for embedding within summary tables and cards to show quick historical context.',
    ARRAY['compliance_score', 'risk_score', 'incident_count', 'finding_count', 'control_coverage', 'maturity_score'],
    '{"time_range": "30d", "line_color": "blue", "fill": false, "height": 32, "show_min_max": false}',
    2,
    1
),
(
    'trend_arrow',
    'Trend Arrow',
    'Directional indicator showing whether a metric is improving, stable, or declining. Minimalist widget for at-a-glance status in dense dashboard layouts.',
    ARRAY['compliance_score', 'risk_score', 'incident_count', 'finding_count', 'control_coverage', 'maturity_score', 'vendor_risk_avg'],
    '{"comparison_period": "7d", "improving_color": "green", "stable_color": "gray", "declining_color": "red", "show_percentage": true, "show_label": true}',
    1,
    1
)
ON CONFLICT (widget_type) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    available_metrics = EXCLUDED.available_metrics,
    default_config = EXCLUDED.default_config,
    min_width = EXCLUDED.min_width,
    min_height = EXCLUDED.min_height;
