package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ============================================================
// ANALYTICS QUERY — Time Series, Comparisons, Top Movers,
//                    Distributions, and Data Export
// ============================================================

// AnalyticsQuery provides read-only query operations over analytics data.
type AnalyticsQuery struct {
	pool *pgxpool.Pool
}

// NewAnalyticsQuery creates a new AnalyticsQuery.
func NewAnalyticsQuery(pool *pgxpool.Pool) *AnalyticsQuery {
	return &AnalyticsQuery{pool: pool}
}

// ============================================================
// STRUCT DEFINITIONS
// ============================================================

// TimeSeriesPoint represents a single data point in a time series.
type TimeSeriesPoint struct {
	Date  time.Time `json:"date"`
	Value float64   `json:"value"`
	Label string    `json:"label,omitempty"`
}

// MetricComparison compares a metric across two time periods.
type MetricComparison struct {
	Metric         string  `json:"metric"`
	CurrentPeriod  string  `json:"current_period"`
	PreviousPeriod string  `json:"previous_period"`
	CurrentValue   float64 `json:"current_value"`
	PreviousValue  float64 `json:"previous_value"`
	AbsoluteChange float64 `json:"absolute_change"`
	PercentChange  float64 `json:"percent_change"`
	Direction      string  `json:"direction"` // "up", "down", "flat"
}

// TopMover represents an entity that has changed the most for a given metric.
type TopMover struct {
	EntityID   string  `json:"entity_id"`
	EntityName string  `json:"entity_name"`
	EntityType string  `json:"entity_type"`
	OldValue   float64 `json:"old_value"`
	NewValue   float64 `json:"new_value"`
	Change     float64 `json:"change"`
	ChangePercent float64 `json:"change_percent"`
}

// DistributionEntry represents a single category in a distribution.
type DistributionEntry struct {
	Category string  `json:"category"`
	Count    int64   `json:"count"`
	Percent  float64 `json:"percent"`
}

// AnalyticsExportConfig configures the data export operation.
type AnalyticsExportConfig struct {
	Format      string   `json:"format"`       // "csv" or "json"
	Metrics     []string `json:"metrics"`       // metrics to include
	StartDate   string   `json:"start_date"`    // YYYY-MM-DD
	EndDate     string   `json:"end_date"`      // YYYY-MM-DD
	Granularity string   `json:"granularity"`   // "daily", "weekly", "monthly"
}

// ============================================================
// TIME SERIES QUERIES
// ============================================================

// validMetrics defines the set of extractable metrics from snapshot JSONB.
var validMetrics = map[string]string{
	"compliance_score":      "metrics->>'overall_compliance_score'",
	"total_risks":           "metrics->>'total_risks'",
	"avg_risk_score":        "metrics->>'avg_risk_score'",
	"open_incidents":        "metrics->>'open_incidents'",
	"open_findings":         "metrics->>'open_findings'",
	"control_coverage":      "metrics->>'control_coverage_rate'",
	"avg_maturity":          "metrics->>'avg_maturity_level'",
	"total_controls":        "metrics->>'total_controls'",
	"implemented_controls":  "metrics->>'implemented_controls'",
	"total_policies":        "metrics->>'total_policies'",
	"policies_due_review":   "metrics->>'policies_due_review'",
	"total_vendors":         "metrics->>'total_vendors'",
	"active_vendors":        "metrics->>'active_vendors'",
	"data_breaches":         "metrics->>'data_breaches'",
}

// GetMetricTimeSeries extracts a time series of a named metric from snapshots.
// Supported granularities: "daily", "weekly", "monthly".
// Period format: "30d", "90d", "6m", "12m", "1y".
func (q *AnalyticsQuery) GetMetricTimeSeries(ctx context.Context, orgID uuid.UUID, metric, period, granularity string) ([]TimeSeriesPoint, error) {
	jsonPath, ok := validMetrics[metric]
	if !ok {
		return nil, fmt.Errorf("unknown metric: %s", metric)
	}

	startDate, err := parsePeriodToDate(period)
	if err != nil {
		return nil, err
	}

	snapshotFilter := "daily"
	switch granularity {
	case "weekly":
		snapshotFilter = "weekly"
	case "monthly":
		snapshotFilter = "monthly"
	}

	query := fmt.Sprintf(`
		SELECT snapshot_date, COALESCE((%s)::FLOAT, 0)
		FROM analytics_snapshots
		WHERE organization_id = $1
			AND snapshot_type = $2
			AND snapshot_date >= $3
		ORDER BY snapshot_date ASC`, jsonPath)

	rows, err := q.pool.Query(ctx, query, orgID, snapshotFilter, startDate)
	if err != nil {
		return nil, fmt.Errorf("query time series: %w", err)
	}
	defer rows.Close()

	var points []TimeSeriesPoint
	for rows.Next() {
		var p TimeSeriesPoint
		if err := rows.Scan(&p.Date, &p.Value); err != nil {
			return nil, fmt.Errorf("scan time series point: %w", err)
		}
		p.Label = metric
		points = append(points, p)
	}

	return points, nil
}

// ============================================================
// METRIC COMPARISON
// ============================================================

// GetMetricComparison compares a metric's average value across two periods.
// Period format: "2024-01", "2024-Q1", or "2024-W05".
func (q *AnalyticsQuery) GetMetricComparison(ctx context.Context, orgID uuid.UUID, metric, currentPeriod, previousPeriod string) (*MetricComparison, error) {
	jsonPath, ok := validMetrics[metric]
	if !ok {
		return nil, fmt.Errorf("unknown metric: %s", metric)
	}

	currentStart, currentEnd, err := parsePeriodRange(currentPeriod)
	if err != nil {
		return nil, fmt.Errorf("parse current period: %w", err)
	}
	previousStart, previousEnd, err := parsePeriodRange(previousPeriod)
	if err != nil {
		return nil, fmt.Errorf("parse previous period: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT COALESCE(AVG((%s)::FLOAT), 0)
		FROM analytics_snapshots
		WHERE organization_id = $1
			AND snapshot_date >= $2 AND snapshot_date <= $3`, jsonPath)

	var currentValue, previousValue float64

	err = q.pool.QueryRow(ctx, query, orgID, currentStart, currentEnd).Scan(&currentValue)
	if err != nil {
		return nil, fmt.Errorf("query current period: %w", err)
	}

	err = q.pool.QueryRow(ctx, query, orgID, previousStart, previousEnd).Scan(&previousValue)
	if err != nil {
		return nil, fmt.Errorf("query previous period: %w", err)
	}

	absoluteChange := currentValue - previousValue
	percentChange := 0.0
	if previousValue != 0 {
		percentChange = (absoluteChange / previousValue) * 100
	}

	direction := "flat"
	if absoluteChange > 0.01 {
		direction = "up"
	} else if absoluteChange < -0.01 {
		direction = "down"
	}

	return &MetricComparison{
		Metric:         metric,
		CurrentPeriod:  currentPeriod,
		PreviousPeriod: previousPeriod,
		CurrentValue:   currentValue,
		PreviousValue:  previousValue,
		AbsoluteChange: absoluteChange,
		PercentChange:  percentChange,
		Direction:      direction,
	}, nil
}

// ============================================================
// TOP MOVERS
// ============================================================

// GetTopMovers returns the entities that have changed the most for a metric.
// direction: "improving" or "declining".
func (q *AnalyticsQuery) GetTopMovers(ctx context.Context, orgID uuid.UUID, metric, period, direction string, limit int) ([]TopMover, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	startDate, err := parsePeriodToDate(period)
	if err != nil {
		return nil, err
	}

	orderDir := "DESC"
	if direction == "declining" {
		orderDir = "ASC"
	}

	// For framework compliance scores, compare trend records
	if metric == "compliance_score" {
		query := fmt.Sprintf(`
			WITH latest AS (
				SELECT DISTINCT ON (framework_code) framework_code, framework_id, compliance_score
				FROM analytics_compliance_trends
				WHERE organization_id = $1
				ORDER BY framework_code, measurement_date DESC
			),
			earliest AS (
				SELECT DISTINCT ON (framework_code) framework_code, framework_id, compliance_score
				FROM analytics_compliance_trends
				WHERE organization_id = $1 AND measurement_date >= $2
				ORDER BY framework_code, measurement_date ASC
			)
			SELECT l.framework_id::TEXT, l.framework_code, 'framework',
				e.compliance_score, l.compliance_score,
				l.compliance_score - e.compliance_score
			FROM latest l
			JOIN earliest e ON l.framework_code = e.framework_code
			ORDER BY (l.compliance_score - e.compliance_score) %s
			LIMIT $3`, orderDir)

		rows, err := q.pool.Query(ctx, query, orgID, startDate, limit)
		if err != nil {
			return nil, fmt.Errorf("query top movers: %w", err)
		}
		defer rows.Close()

		var movers []TopMover
		for rows.Next() {
			var m TopMover
			if err := rows.Scan(&m.EntityID, &m.EntityName, &m.EntityType,
				&m.OldValue, &m.NewValue, &m.Change); err != nil {
				return nil, fmt.Errorf("scan top mover: %w", err)
			}
			if m.OldValue != 0 {
				m.ChangePercent = (m.Change / m.OldValue) * 100
			}
			movers = append(movers, m)
		}
		return movers, nil
	}

	// Generic top movers from snapshots: compare earliest and latest values
	jsonPath, ok := validMetrics[metric]
	if !ok {
		return nil, fmt.Errorf("unknown metric: %s", metric)
	}

	query := fmt.Sprintf(`
		WITH latest AS (
			SELECT organization_id, (%s)::FLOAT AS val, snapshot_date
			FROM analytics_snapshots
			WHERE organization_id = $1 AND snapshot_type = 'daily'
			ORDER BY snapshot_date DESC LIMIT 1
		),
		earliest AS (
			SELECT organization_id, (%s)::FLOAT AS val, snapshot_date
			FROM analytics_snapshots
			WHERE organization_id = $1 AND snapshot_type = 'daily' AND snapshot_date >= $2
			ORDER BY snapshot_date ASC LIMIT 1
		)
		SELECT 'org'::TEXT, 'organization'::TEXT, 'metric',
			COALESCE(e.val, 0), COALESCE(l.val, 0),
			COALESCE(l.val, 0) - COALESCE(e.val, 0)
		FROM latest l
		LEFT JOIN earliest e ON l.organization_id = e.organization_id
		LIMIT $3`, jsonPath, jsonPath)

	rows, err := q.pool.Query(ctx, query, orgID, startDate, limit)
	if err != nil {
		return nil, fmt.Errorf("query generic top movers: %w", err)
	}
	defer rows.Close()

	var movers []TopMover
	for rows.Next() {
		var m TopMover
		if err := rows.Scan(&m.EntityID, &m.EntityName, &m.EntityType,
			&m.OldValue, &m.NewValue, &m.Change); err != nil {
			return nil, fmt.Errorf("scan top mover: %w", err)
		}
		if m.OldValue != 0 {
			m.ChangePercent = (m.Change / m.OldValue) * 100
		}
		movers = append(movers, m)
	}
	return movers, nil
}

// ============================================================
// DISTRIBUTION
// ============================================================

// GetDistribution returns the count-based distribution of an entity grouped by a category.
// Supported entities: "risks", "incidents", "findings", "vendors", "controls", "policies".
func (q *AnalyticsQuery) GetDistribution(ctx context.Context, orgID uuid.UUID, entity, groupBy string) ([]DistributionEntry, error) {
	var query string

	switch entity {
	case "risks":
		col := "residual_risk_level::TEXT"
		if groupBy == "status" {
			col = "status::TEXT"
		} else if groupBy == "category" {
			col = "COALESCE(risk_source, 'unknown')"
		}
		query = fmt.Sprintf(`
			SELECT %s AS category, COUNT(*) AS cnt
			FROM risks
			WHERE organization_id = $1 AND deleted_at IS NULL
			GROUP BY category ORDER BY cnt DESC`, col)

	case "incidents":
		col := "severity::TEXT"
		if groupBy == "type" {
			col = "incident_type::TEXT"
		} else if groupBy == "status" {
			col = "status::TEXT"
		}
		query = fmt.Sprintf(`
			SELECT %s AS category, COUNT(*) AS cnt
			FROM incidents
			WHERE organization_id = $1 AND deleted_at IS NULL
			GROUP BY category ORDER BY cnt DESC`, col)

	case "findings":
		col := "severity::TEXT"
		if groupBy == "status" {
			col = "status::TEXT"
		}
		query = fmt.Sprintf(`
			SELECT %s AS category, COUNT(*) AS cnt
			FROM audit_findings
			WHERE organization_id = $1 AND deleted_at IS NULL
			GROUP BY category ORDER BY cnt DESC`, col)

	case "vendors":
		col := "risk_tier::TEXT"
		if groupBy == "status" {
			col = "status::TEXT"
		}
		query = fmt.Sprintf(`
			SELECT %s AS category, COUNT(*) AS cnt
			FROM vendors
			WHERE organization_id = $1 AND deleted_at IS NULL
			GROUP BY category ORDER BY cnt DESC`, col)

	case "controls":
		col := "status::TEXT"
		if groupBy == "type" {
			col = "COALESCE(control_type::TEXT, 'unknown')"
		}
		query = fmt.Sprintf(`
			SELECT %s AS category, COUNT(*) AS cnt
			FROM control_implementations
			WHERE organization_id = $1 AND deleted_at IS NULL
			GROUP BY category ORDER BY cnt DESC`, col)

	case "policies":
		col := "status::TEXT"
		if groupBy == "classification" {
			col = "classification::TEXT"
		}
		query = fmt.Sprintf(`
			SELECT %s AS category, COUNT(*) AS cnt
			FROM policies
			WHERE organization_id = $1 AND deleted_at IS NULL
			GROUP BY category ORDER BY cnt DESC`, col)

	default:
		return nil, fmt.Errorf("unsupported entity: %s", entity)
	}

	rows, err := q.pool.Query(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("query distribution: %w", err)
	}
	defer rows.Close()

	var entries []DistributionEntry
	var total int64
	for rows.Next() {
		var e DistributionEntry
		if err := rows.Scan(&e.Category, &e.Count); err != nil {
			return nil, fmt.Errorf("scan distribution entry: %w", err)
		}
		total += e.Count
		entries = append(entries, e)
	}

	// Calculate percentages
	for i := range entries {
		if total > 0 {
			entries[i].Percent = float64(entries[i].Count) / float64(total) * 100
		}
	}

	return entries, nil
}

// ============================================================
// DATA EXPORT
// ============================================================

// ExportAnalyticsData exports analytics data in the requested format.
// Returns the data bytes and the content-type header value.
func (q *AnalyticsQuery) ExportAnalyticsData(ctx context.Context, orgID uuid.UUID, config AnalyticsExportConfig) ([]byte, string, error) {
	startDate, err := time.Parse("2006-01-02", config.StartDate)
	if err != nil {
		startDate = time.Now().UTC().AddDate(0, -12, 0)
	}
	endDate, err := time.Parse("2006-01-02", config.EndDate)
	if err != nil {
		endDate = time.Now().UTC()
	}

	granularity := "daily"
	if config.Granularity != "" {
		granularity = config.Granularity
	}

	// Query snapshots within the range
	rows, err := q.pool.Query(ctx, `
		SELECT snapshot_date, snapshot_type::TEXT, metrics
		FROM analytics_snapshots
		WHERE organization_id = $1
			AND snapshot_type = $2
			AND snapshot_date >= $3 AND snapshot_date <= $4
		ORDER BY snapshot_date ASC`,
		orgID, granularity, startDate, endDate,
	)
	if err != nil {
		return nil, "", fmt.Errorf("query export data: %w", err)
	}
	defer rows.Close()

	var exportRows []exportRow
	for rows.Next() {
		var date time.Time
		var snapType string
		var metricsJSON []byte
		if err := rows.Scan(&date, &snapType, &metricsJSON); err != nil {
			return nil, "", fmt.Errorf("scan export row: %w", err)
		}

		var metrics map[string]interface{}
		if err := json.Unmarshal(metricsJSON, &metrics); err != nil {
			continue
		}

		// Filter to requested metrics if specified
		if len(config.Metrics) > 0 {
			filtered := make(map[string]interface{})
			for _, m := range config.Metrics {
				if v, ok := metrics[m]; ok {
					filtered[m] = v
				}
			}
			metrics = filtered
		}

		exportRows = append(exportRows, exportRow{
			Date:         date.Format("2006-01-02"),
			SnapshotType: snapType,
			Metrics:      metrics,
		})
	}

	switch config.Format {
	case "csv":
		return q.exportCSV(exportRows, config.Metrics)
	default:
		return q.exportJSON(exportRows)
	}
}

// exportJSON marshals export data to JSON.
func (q *AnalyticsQuery) exportJSON(rows []exportRow) ([]byte, string, error) {
	data, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		return nil, "", fmt.Errorf("marshal export JSON: %w", err)
	}
	return data, "application/json", nil
}

// exportCSV converts export data to CSV format.
func (q *AnalyticsQuery) exportCSV(rows []exportRow, metrics []string) ([]byte, string, error) {
	if len(rows) == 0 {
		return []byte("date,snapshot_type\n"), "text/csv", nil
	}

	// Determine column headers from metrics or from first row's keys
	columns := metrics
	if len(columns) == 0 && len(rows) > 0 {
		for k := range rows[0].Metrics {
			columns = append(columns, k)
		}
		// Sort columns for consistency
		sortStrings(columns)
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write header
	header := []string{"date", "snapshot_type"}
	header = append(header, columns...)
	if err := writer.Write(header); err != nil {
		return nil, "", fmt.Errorf("write CSV header: %w", err)
	}

	// Write rows
	for _, row := range rows {
		record := []string{row.Date, row.SnapshotType}
		for _, col := range columns {
			val := ""
			if v, ok := row.Metrics[col]; ok {
				val = fmt.Sprintf("%v", v)
			}
			record = append(record, val)
		}
		if err := writer.Write(record); err != nil {
			return nil, "", fmt.Errorf("write CSV row: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, "", fmt.Errorf("flush CSV: %w", err)
	}

	return buf.Bytes(), "text/csv", nil
}

// ============================================================
// HELPER FUNCTIONS
// ============================================================

type exportRow struct {
	Date         string                 `json:"date"`
	SnapshotType string                 `json:"snapshot_type"`
	Metrics      map[string]interface{} `json:"metrics"`
}

// parsePeriodToDate converts a period string to a start date.
// Supported formats: "30d", "90d", "6m", "12m", "1y", "2y".
func parsePeriodToDate(period string) (time.Time, error) {
	now := time.Now().UTC()
	if len(period) < 2 {
		return now.AddDate(0, -12, 0), nil
	}

	unit := period[len(period)-1:]
	numStr := period[:len(period)-1]

	var num int
	for _, c := range numStr {
		if c < '0' || c > '9' {
			return time.Time{}, fmt.Errorf("invalid period format: %s", period)
		}
		num = num*10 + int(c-'0')
	}

	switch unit {
	case "d":
		return now.AddDate(0, 0, -num), nil
	case "m":
		return now.AddDate(0, -num, 0), nil
	case "y":
		return now.AddDate(-num, 0, 0), nil
	default:
		return time.Time{}, fmt.Errorf("unknown period unit: %s", unit)
	}
}

// parsePeriodRange converts a period string to a start and end date.
// Supported formats: "2024-01" (month), "2024-Q1" (quarter).
func parsePeriodRange(period string) (time.Time, time.Time, error) {
	// Month format: "2024-01"
	if len(period) == 7 && period[4] == '-' {
		t, err := time.Parse("2006-01", period)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("parse month period: %w", err)
		}
		start := t
		end := t.AddDate(0, 1, -1)
		return start, end, nil
	}

	// Quarter format: "2024-Q1"
	if len(period) == 7 && period[5] == 'Q' {
		yearStr := period[:4]
		qStr := period[6:]
		var year, q int
		for _, c := range yearStr {
			year = year*10 + int(c-'0')
		}
		for _, c := range qStr {
			q = q*10 + int(c-'0')
		}
		if q < 1 || q > 4 {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid quarter: %d", q)
		}
		startMonth := time.Month((q-1)*3 + 1)
		start := time.Date(year, startMonth, 1, 0, 0, 0, 0, time.UTC)
		end := start.AddDate(0, 3, -1)
		return start, end, nil
	}

	return time.Time{}, time.Time{}, fmt.Errorf("unsupported period format: %s", period)
}

// sortStrings sorts a string slice using insertion sort.
func sortStrings(a []string) {
	for i := 1; i < len(a); i++ {
		key := a[i]
		j := i - 1
		for j >= 0 && a[j] > key {
			a[j+1] = a[j]
			j--
		}
		a[j+1] = key
	}
}
