package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// ============================================================
// ANALYTICS ENGINE — Snapshot Collector, Trend Calculator,
//                     Risk Predictor, Benchmark Engine
// ============================================================

// AnalyticsEngine handles analytics computation including snapshots,
// compliance trends, risk predictions, and peer benchmarking.
type AnalyticsEngine struct {
	pool *pgxpool.Pool
}

// NewAnalyticsEngine creates a new AnalyticsEngine.
func NewAnalyticsEngine(pool *pgxpool.Pool) *AnalyticsEngine {
	return &AnalyticsEngine{pool: pool}
}

// Pool exposes the database pool for scheduler use.
func (e *AnalyticsEngine) Pool() *pgxpool.Pool {
	return e.pool
}

// ============================================================
// STRUCT DEFINITIONS
// ============================================================

// SnapshotMetrics captures the comprehensive state of an organization at a point in time.
type SnapshotMetrics struct {
	// Risk metrics
	TotalRisks        int64            `json:"total_risks"`
	RisksBySeverity   map[string]int64 `json:"risks_by_severity"`
	RisksByStatus     map[string]int64 `json:"risks_by_status"`
	AvgRiskScore      float64          `json:"avg_risk_score"`
	OverdueRiskReview int64            `json:"overdue_risk_reviews"`

	// Control metrics
	TotalControls         int64   `json:"total_controls"`
	ImplementedControls   int64   `json:"implemented_controls"`
	ControlCoverageRate   float64 `json:"control_coverage_rate"`
	AvgMaturityLevel      float64 `json:"avg_maturity_level"`

	// Incident metrics
	TotalIncidents   int64            `json:"total_incidents"`
	OpenIncidents    int64            `json:"open_incidents"`
	IncidentsByType  map[string]int64 `json:"incidents_by_type"`
	DataBreaches     int64            `json:"data_breaches"`

	// Policy metrics
	TotalPolicies        int64 `json:"total_policies"`
	PublishedPolicies    int64 `json:"published_policies"`
	PoliciesDueReview    int64 `json:"policies_due_review"`

	// Vendor metrics
	TotalVendors         int64            `json:"total_vendors"`
	ActiveVendors        int64            `json:"active_vendors"`
	VendorsByRiskTier    map[string]int64 `json:"vendors_by_risk_tier"`

	// Finding metrics
	TotalFindings        int64            `json:"total_findings"`
	OpenFindings         int64            `json:"open_findings"`
	FindingsBySeverity   map[string]int64 `json:"findings_by_severity"`

	// Compliance metrics
	OverallComplianceScore float64                  `json:"overall_compliance_score"`
	FrameworkScores        map[string]float64       `json:"framework_scores"`

	// Timestamps
	CapturedAt           time.Time `json:"captured_at"`
}

// ComplianceTrend represents a compliance score measurement for a specific framework and date.
type ComplianceTrend struct {
	ID                  uuid.UUID `json:"id"`
	OrganizationID      uuid.UUID `json:"organization_id"`
	FrameworkID         uuid.UUID `json:"framework_id"`
	FrameworkCode       string    `json:"framework_code"`
	MeasurementDate     time.Time `json:"measurement_date"`
	ComplianceScore     float64   `json:"compliance_score"`
	ControlsImplemented int       `json:"controls_implemented"`
	ControlsTotal       int       `json:"controls_total"`
	MaturityAvg         float64   `json:"maturity_avg"`
	ScoreChange7d       float64   `json:"score_change_7d"`
	ScoreChange30d      float64   `json:"score_change_30d"`
	ScoreChange90d      float64   `json:"score_change_90d"`
	TrendDirection      string    `json:"trend_direction"`
	CreatedAt           time.Time `json:"created_at"`
}

// RiskPrediction holds the predicted trajectory for a specific risk.
type RiskPrediction struct {
	RiskID            uuid.UUID          `json:"risk_id"`
	CurrentScore      float64            `json:"current_score"`
	Forecasts         []ForecastPoint    `json:"forecasts"`
	ModelVersion      string             `json:"model_version"`
	SmoothingAlpha    float64            `json:"smoothing_alpha"`
	InputFeatures     map[string]float64 `json:"input_features"`
	ConfidenceLevel   float64            `json:"confidence_level"`
	GeneratedAt       time.Time          `json:"generated_at"`
}

// ForecastPoint is a single prediction for a future date.
type ForecastPoint struct {
	Date                  time.Time `json:"date"`
	PredictedValue        float64   `json:"predicted_value"`
	ConfidenceIntervalLow float64   `json:"confidence_interval_low"`
	ConfidenceIntervalHigh float64  `json:"confidence_interval_high"`
	DaysAhead             int       `json:"days_ahead"`
}

// BreachPrediction contains the estimated probability of a data breach.
type BreachPrediction struct {
	OrganizationID        uuid.UUID          `json:"organization_id"`
	BreachProbability30d  float64            `json:"breach_probability_30d"`
	BreachProbability90d  float64            `json:"breach_probability_90d"`
	BreachProbability365d float64            `json:"breach_probability_365d"`
	ConfidenceLevel       float64            `json:"confidence_level"`
	RiskFactors           []RiskFactor       `json:"risk_factors"`
	ModelVersion          string             `json:"model_version"`
	InputFeatures         map[string]float64 `json:"input_features"`
	GeneratedAt           time.Time          `json:"generated_at"`
}

// RiskFactor describes a single contributing factor to breach probability.
type RiskFactor struct {
	Name        string  `json:"name"`
	Weight      float64 `json:"weight"`
	Value       float64 `json:"value"`
	Contribution float64 `json:"contribution"`
	Description string  `json:"description"`
}

// PeerComparison shows an organization's position relative to anonymized benchmarks.
type PeerComparison struct {
	OrganizationID  uuid.UUID          `json:"organization_id"`
	Metrics         []PeerMetric       `json:"metrics"`
	BenchmarkPeriod string             `json:"benchmark_period"`
	SampleSize      int                `json:"sample_size"`
	GeneratedAt     time.Time          `json:"generated_at"`
}

// PeerMetric is a single benchmark comparison for one metric.
type PeerMetric struct {
	MetricName    string  `json:"metric_name"`
	OrgValue      float64 `json:"org_value"`
	Percentile25  float64 `json:"percentile_25"`
	Percentile50  float64 `json:"percentile_50"`
	Percentile75  float64 `json:"percentile_75"`
	Percentile90  float64 `json:"percentile_90"`
	PercentilePos float64 `json:"percentile_position"`
	SampleSize    int     `json:"sample_size"`
}

// ModelValidation compares historical predictions with actual outcomes.
type ModelValidation struct {
	OrganizationID       uuid.UUID             `json:"organization_id"`
	TotalPredictions     int                   `json:"total_predictions"`
	ValidatedPredictions int                   `json:"validated_predictions"`
	MeanAbsoluteError    float64               `json:"mean_absolute_error"`
	MeanSquaredError     float64               `json:"mean_squared_error"`
	AccuracyByType       map[string]TypeAccuracy `json:"accuracy_by_type"`
	ModelVersion         string                `json:"model_version"`
	ValidatedAt          time.Time             `json:"validated_at"`
}

// TypeAccuracy holds accuracy metrics for a specific prediction type.
type TypeAccuracy struct {
	PredictionType       string  `json:"prediction_type"`
	Count                int     `json:"count"`
	MeanAbsoluteError    float64 `json:"mean_absolute_error"`
	WithinConfidenceRate float64 `json:"within_confidence_rate"`
}

// ============================================================
// SNAPSHOT COLLECTOR
// ============================================================

// TakeSnapshot captures a comprehensive metrics snapshot from the current state
// of the organization's GRC data. Snapshots are immutable.
func (e *AnalyticsEngine) TakeSnapshot(ctx context.Context, orgID uuid.UUID, snapshotType string) error {
	log.Info().Str("org_id", orgID.String()).Str("type", snapshotType).Msg("Taking analytics snapshot")

	metrics := &SnapshotMetrics{
		RisksBySeverity:    make(map[string]int64),
		RisksByStatus:      make(map[string]int64),
		IncidentsByType:    make(map[string]int64),
		VendorsByRiskTier:  make(map[string]int64),
		FindingsBySeverity: make(map[string]int64),
		FrameworkScores:    make(map[string]float64),
		CapturedAt:         time.Now().UTC(),
	}

	// --- Risk metrics ---
	var rCrit, rHigh, rMed, rLow, rVLow int64
	var sIdent, sAssess, sTreat, sAccept, sClose, sMon int64
	err := e.pool.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE residual_risk_level = 'critical'),
			COUNT(*) FILTER (WHERE residual_risk_level = 'high'),
			COUNT(*) FILTER (WHERE residual_risk_level = 'medium'),
			COUNT(*) FILTER (WHERE residual_risk_level = 'low'),
			COUNT(*) FILTER (WHERE residual_risk_level = 'very_low'),
			COUNT(*) FILTER (WHERE status = 'identified'),
			COUNT(*) FILTER (WHERE status = 'assessed'),
			COUNT(*) FILTER (WHERE status = 'treated'),
			COUNT(*) FILTER (WHERE status = 'accepted'),
			COUNT(*) FILTER (WHERE status = 'closed'),
			COUNT(*) FILTER (WHERE status = 'monitoring'),
			COALESCE(AVG(residual_risk_score), 0),
			COUNT(*) FILTER (WHERE next_review_date < CURRENT_DATE)
		FROM risks
		WHERE organization_id = $1 AND deleted_at IS NULL`, orgID,
	).Scan(
		&metrics.TotalRisks,
		&rCrit, &rHigh, &rMed, &rLow, &rVLow,
		&sIdent, &sAssess, &sTreat, &sAccept, &sClose, &sMon,
		&metrics.AvgRiskScore,
		&metrics.OverdueRiskReview,
	)
	if err != nil && err != pgx.ErrNoRows {
		return fmt.Errorf("snapshot risks: %w", err)
	}
	metrics.RisksBySeverity["critical"] = rCrit
	metrics.RisksBySeverity["high"] = rHigh
	metrics.RisksBySeverity["medium"] = rMed
	metrics.RisksBySeverity["low"] = rLow
	metrics.RisksBySeverity["very_low"] = rVLow
	metrics.RisksByStatus["identified"] = sIdent
	metrics.RisksByStatus["assessed"] = sAssess
	metrics.RisksByStatus["treated"] = sTreat
	metrics.RisksByStatus["accepted"] = sAccept
	metrics.RisksByStatus["closed"] = sClose
	metrics.RisksByStatus["monitoring"] = sMon

	// --- Control metrics ---
	err = e.pool.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE status IN ('implemented', 'effective')),
			COALESCE(AVG(maturity_level), 0)
		FROM control_implementations
		WHERE organization_id = $1 AND deleted_at IS NULL`, orgID,
	).Scan(
		&metrics.TotalControls,
		&metrics.ImplementedControls,
		&metrics.AvgMaturityLevel,
	)
	if err != nil && err != pgx.ErrNoRows {
		return fmt.Errorf("snapshot controls: %w", err)
	}
	if metrics.TotalControls > 0 {
		metrics.ControlCoverageRate = float64(metrics.ImplementedControls) / float64(metrics.TotalControls) * 100
	}

	// --- Incident metrics ---
	var iBreach, iSec, iOps, iComp int64
	err = e.pool.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE status NOT IN ('resolved', 'closed')),
			COUNT(*) FILTER (WHERE incident_type = 'data_breach'),
			COUNT(*) FILTER (WHERE incident_type = 'security'),
			COUNT(*) FILTER (WHERE incident_type = 'operational'),
			COUNT(*) FILTER (WHERE incident_type = 'compliance'),
			COUNT(*) FILTER (WHERE is_data_breach = true)
		FROM incidents
		WHERE organization_id = $1 AND deleted_at IS NULL`, orgID,
	).Scan(
		&metrics.TotalIncidents,
		&metrics.OpenIncidents,
		&iBreach, &iSec, &iOps, &iComp,
		&metrics.DataBreaches,
	)
	if err != nil && err != pgx.ErrNoRows {
		return fmt.Errorf("snapshot incidents: %w", err)
	}
	metrics.IncidentsByType["data_breach"] = iBreach
	metrics.IncidentsByType["security"] = iSec
	metrics.IncidentsByType["operational"] = iOps
	metrics.IncidentsByType["compliance"] = iComp

	// --- Policy metrics ---
	err = e.pool.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE status = 'published'),
			COUNT(*) FILTER (WHERE next_review_date <= CURRENT_DATE)
		FROM policies
		WHERE organization_id = $1 AND deleted_at IS NULL`, orgID,
	).Scan(
		&metrics.TotalPolicies,
		&metrics.PublishedPolicies,
		&metrics.PoliciesDueReview,
	)
	if err != nil && err != pgx.ErrNoRows {
		return fmt.Errorf("snapshot policies: %w", err)
	}

	// --- Vendor metrics ---
	var vCrit, vHigh, vMed, vLow int64
	err = e.pool.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE status = 'active'),
			COUNT(*) FILTER (WHERE risk_tier = 'critical'),
			COUNT(*) FILTER (WHERE risk_tier = 'high'),
			COUNT(*) FILTER (WHERE risk_tier = 'medium'),
			COUNT(*) FILTER (WHERE risk_tier = 'low')
		FROM vendors
		WHERE organization_id = $1 AND deleted_at IS NULL`, orgID,
	).Scan(
		&metrics.TotalVendors,
		&metrics.ActiveVendors,
		&vCrit, &vHigh, &vMed, &vLow,
	)
	if err != nil && err != pgx.ErrNoRows {
		return fmt.Errorf("snapshot vendors: %w", err)
	}
	metrics.VendorsByRiskTier["critical"] = vCrit
	metrics.VendorsByRiskTier["high"] = vHigh
	metrics.VendorsByRiskTier["medium"] = vMed
	metrics.VendorsByRiskTier["low"] = vLow

	// --- Finding metrics ---
	var fCrit, fHigh, fMed, fLow, fInfo int64
	err = e.pool.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE status IN ('open', 'in_progress')),
			COUNT(*) FILTER (WHERE severity = 'critical'),
			COUNT(*) FILTER (WHERE severity = 'high'),
			COUNT(*) FILTER (WHERE severity = 'medium'),
			COUNT(*) FILTER (WHERE severity = 'low'),
			COUNT(*) FILTER (WHERE severity = 'informational')
		FROM audit_findings
		WHERE organization_id = $1 AND deleted_at IS NULL`, orgID,
	).Scan(
		&metrics.TotalFindings,
		&metrics.OpenFindings,
		&fCrit, &fHigh, &fMed, &fLow, &fInfo,
	)
	if err != nil && err != pgx.ErrNoRows {
		return fmt.Errorf("snapshot findings: %w", err)
	}
	metrics.FindingsBySeverity["critical"] = fCrit
	metrics.FindingsBySeverity["high"] = fHigh
	metrics.FindingsBySeverity["medium"] = fMed
	metrics.FindingsBySeverity["low"] = fLow
	metrics.FindingsBySeverity["informational"] = fInfo

	// --- Framework compliance scores ---
	rows, err := e.pool.Query(ctx, `
		SELECT
			cf.code,
			COALESCE(of2.compliance_score, 0)
		FROM organization_frameworks of2
		JOIN compliance_frameworks cf ON of2.framework_id = cf.id
		WHERE of2.organization_id = $1`, orgID)
	if err != nil {
		return fmt.Errorf("snapshot framework scores: %w", err)
	}
	defer rows.Close()

	var totalScore float64
	var frameCount int
	for rows.Next() {
		var code string
		var score float64
		if err := rows.Scan(&code, &score); err != nil {
			return fmt.Errorf("scan framework score: %w", err)
		}
		metrics.FrameworkScores[code] = score
		totalScore += score
		frameCount++
	}
	if frameCount > 0 {
		metrics.OverallComplianceScore = math.Round(totalScore/float64(frameCount)*100) / 100
	}

	// Marshal metrics to JSONB
	metricsJSON, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("marshal snapshot metrics: %w", err)
	}

	// Insert immutable snapshot
	_, err = e.pool.Exec(ctx, `
		INSERT INTO analytics_snapshots (organization_id, snapshot_type, snapshot_date, metrics)
		VALUES ($1, $2, CURRENT_DATE, $3)`,
		orgID, snapshotType, metricsJSON,
	)
	if err != nil {
		return fmt.Errorf("insert snapshot: %w", err)
	}

	log.Info().Str("org_id", orgID.String()).Str("type", snapshotType).
		Float64("compliance_score", metrics.OverallComplianceScore).
		Int64("total_risks", metrics.TotalRisks).
		Msg("Analytics snapshot captured")

	return nil
}

// ============================================================
// TREND CALCULATOR
// ============================================================

// CalculateComplianceTrends computes compliance trends for each framework
// adopted by the organization, including score deltas and trend direction
// determined by linear regression over the last 30 snapshots.
func (e *AnalyticsEngine) CalculateComplianceTrends(ctx context.Context, orgID uuid.UUID) error {
	log.Info().Str("org_id", orgID.String()).Msg("Calculating compliance trends")

	// Get all adopted frameworks for this org
	rows, err := e.pool.Query(ctx, `
		SELECT of2.id, of2.framework_id, cf.code,
			COALESCE(of2.compliance_score, 0)
		FROM organization_frameworks of2
		JOIN compliance_frameworks cf ON of2.framework_id = cf.id
		WHERE of2.organization_id = $1`, orgID)
	if err != nil {
		return fmt.Errorf("query frameworks: %w", err)
	}
	defer rows.Close()

	type fwData struct {
		OrgFwID     uuid.UUID
		FrameworkID uuid.UUID
		Code        string
		Score       float64
	}
	var frameworks []fwData
	for rows.Next() {
		var fw fwData
		if err := rows.Scan(&fw.OrgFwID, &fw.FrameworkID, &fw.Code, &fw.Score); err != nil {
			return fmt.Errorf("scan framework: %w", err)
		}
		frameworks = append(frameworks, fw)
	}

	today := time.Now().UTC().Truncate(24 * time.Hour)

	for _, fw := range frameworks {
		// Get control implementation counts
		var implemented, total int
		err := e.pool.QueryRow(ctx, `
			SELECT
				COUNT(*) FILTER (WHERE status IN ('implemented', 'effective')),
				COUNT(*)
			FROM control_implementations
			WHERE organization_id = $1 AND org_framework_id = $2 AND deleted_at IS NULL`,
			orgID, fw.OrgFwID,
		).Scan(&implemented, &total)
		if err != nil && err != pgx.ErrNoRows {
			log.Error().Err(err).Str("framework", fw.Code).Msg("Failed to query control counts")
			continue
		}

		// Get maturity average
		var maturityAvg float64
		err = e.pool.QueryRow(ctx, `
			SELECT COALESCE(AVG(maturity_level), 0)
			FROM control_implementations
			WHERE organization_id = $1 AND org_framework_id = $2
				AND status != 'not_applicable' AND deleted_at IS NULL`,
			orgID, fw.OrgFwID,
		).Scan(&maturityAvg)
		if err != nil && err != pgx.ErrNoRows {
			log.Error().Err(err).Str("framework", fw.Code).Msg("Failed to query maturity")
			continue
		}

		// Get historical scores for delta calculation
		scoreChange7d := e.getScoreDelta(ctx, orgID, fw.FrameworkID, today, 7)
		scoreChange30d := e.getScoreDelta(ctx, orgID, fw.FrameworkID, today, 30)
		scoreChange90d := e.getScoreDelta(ctx, orgID, fw.FrameworkID, today, 90)

		// Determine trend direction via linear regression on last 30 data points
		direction := e.calculateTrendDirection(ctx, orgID, fw.FrameworkID)

		// Upsert trend record
		_, err = e.pool.Exec(ctx, `
			INSERT INTO analytics_compliance_trends
				(organization_id, framework_id, framework_code, measurement_date,
				 compliance_score, controls_implemented, controls_total,
				 maturity_avg, score_change_7d, score_change_30d, score_change_90d,
				 trend_direction)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
			ON CONFLICT (organization_id, framework_id, measurement_date)
			DO UPDATE SET
				compliance_score = EXCLUDED.compliance_score,
				controls_implemented = EXCLUDED.controls_implemented,
				controls_total = EXCLUDED.controls_total,
				maturity_avg = EXCLUDED.maturity_avg,
				score_change_7d = EXCLUDED.score_change_7d,
				score_change_30d = EXCLUDED.score_change_30d,
				score_change_90d = EXCLUDED.score_change_90d,
				trend_direction = EXCLUDED.trend_direction`,
			orgID, fw.FrameworkID, fw.Code, today,
			fw.Score, implemented, total,
			math.Round(maturityAvg*100)/100,
			scoreChange7d, scoreChange30d, scoreChange90d,
			direction,
		)
		if err != nil {
			log.Error().Err(err).Str("framework", fw.Code).Msg("Failed to upsert compliance trend")
			continue
		}
	}

	return nil
}

// getScoreDelta returns the score change over the given number of days.
func (e *AnalyticsEngine) getScoreDelta(ctx context.Context, orgID, frameworkID uuid.UUID, today time.Time, days int) float64 {
	pastDate := today.AddDate(0, 0, -days)
	var pastScore float64
	err := e.pool.QueryRow(ctx, `
		SELECT compliance_score FROM analytics_compliance_trends
		WHERE organization_id = $1 AND framework_id = $2
			AND measurement_date <= $3
		ORDER BY measurement_date DESC LIMIT 1`,
		orgID, frameworkID, pastDate,
	).Scan(&pastScore)
	if err != nil {
		return 0 // no historical data available
	}

	var currentScore float64
	err = e.pool.QueryRow(ctx, `
		SELECT compliance_score FROM analytics_compliance_trends
		WHERE organization_id = $1 AND framework_id = $2
		ORDER BY measurement_date DESC LIMIT 1`,
		orgID, frameworkID,
	).Scan(&currentScore)
	if err != nil {
		return 0
	}

	return math.Round((currentScore-pastScore)*100) / 100
}

// calculateTrendDirection uses ordinary least squares linear regression
// on the last 30 compliance trend measurements to determine direction.
func (e *AnalyticsEngine) calculateTrendDirection(ctx context.Context, orgID, frameworkID uuid.UUID) string {
	rows, err := e.pool.Query(ctx, `
		SELECT compliance_score, measurement_date
		FROM analytics_compliance_trends
		WHERE organization_id = $1 AND framework_id = $2
		ORDER BY measurement_date DESC LIMIT 30`, orgID, frameworkID)
	if err != nil {
		return "stable"
	}
	defer rows.Close()

	var scores []float64
	for rows.Next() {
		var score float64
		var date time.Time
		if err := rows.Scan(&score, &date); err != nil {
			continue
		}
		scores = append(scores, score)
	}

	if len(scores) < 3 {
		return "stable"
	}

	// Reverse so index 0 = oldest
	for i, j := 0, len(scores)-1; i < j; i, j = i+1, j-1 {
		scores[i], scores[j] = scores[j], scores[i]
	}

	slope := LinearRegressionSlope(scores)

	if slope > 0.5 {
		return "improving"
	} else if slope < -0.5 {
		return "declining"
	}
	return "stable"
}

// LinearRegressionSlope computes the slope of the best-fit line using
// ordinary least squares. X-values are sequential indices 0..n-1.
func LinearRegressionSlope(values []float64) float64 {
	n := float64(len(values))
	if n < 2 {
		return 0
	}

	var sumX, sumY, sumXY, sumX2 float64
	for i, y := range values {
		x := float64(i)
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	denominator := n*sumX2 - sumX*sumX
	if denominator == 0 {
		return 0
	}
	return (n*sumXY - sumX*sumY) / denominator
}

// GetComplianceTrends retrieves historical compliance trend data for a framework.
func (e *AnalyticsEngine) GetComplianceTrends(ctx context.Context, orgID uuid.UUID, frameworkCode string, months int) ([]ComplianceTrend, error) {
	if months <= 0 {
		months = 12
	}
	startDate := time.Now().UTC().AddDate(0, -months, 0)

	query := `
		SELECT id, organization_id, framework_id, framework_code,
			measurement_date, compliance_score, controls_implemented,
			controls_total, maturity_avg, score_change_7d, score_change_30d,
			score_change_90d, trend_direction::TEXT, created_at
		FROM analytics_compliance_trends
		WHERE organization_id = $1 AND measurement_date >= $2`
	args := []interface{}{orgID, startDate}

	if frameworkCode != "" {
		query += " AND framework_code = $3"
		args = append(args, frameworkCode)
	}
	query += " ORDER BY measurement_date ASC"

	rows, err := e.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query compliance trends: %w", err)
	}
	defer rows.Close()

	var trends []ComplianceTrend
	for rows.Next() {
		var t ComplianceTrend
		if err := rows.Scan(
			&t.ID, &t.OrganizationID, &t.FrameworkID, &t.FrameworkCode,
			&t.MeasurementDate, &t.ComplianceScore, &t.ControlsImplemented,
			&t.ControlsTotal, &t.MaturityAvg, &t.ScoreChange7d, &t.ScoreChange30d,
			&t.ScoreChange90d, &t.TrendDirection, &t.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan compliance trend: %w", err)
		}
		trends = append(trends, t)
	}
	return trends, nil
}

// ============================================================
// RISK PREDICTOR — Statistical Models (not AI)
// ============================================================

// PredictRiskScoreTrajectory uses simple exponential smoothing (alpha=0.3)
// on historical risk scores to forecast +30, +60, and +90 day values
// with 95% confidence intervals.
func (e *AnalyticsEngine) PredictRiskScoreTrajectory(ctx context.Context, orgID, riskID uuid.UUID) (*RiskPrediction, error) {
	// Get current risk score
	var currentScore float64
	err := e.pool.QueryRow(ctx, `
		SELECT COALESCE(residual_risk_score, 0)
		FROM risks WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL`,
		riskID, orgID,
	).Scan(&currentScore)
	if err != nil {
		return nil, fmt.Errorf("query current risk score: %w", err)
	}

	// Get historical scores from snapshots
	historicalScores := e.getRiskHistoricalScores(ctx, orgID, riskID)
	if len(historicalScores) == 0 {
		historicalScores = []float64{currentScore}
	}

	// Simple Exponential Smoothing (SES) with alpha=0.3
	const alpha = 0.3
	smoothed := ExponentialSmoothing(historicalScores, alpha)

	// Calculate standard error from residuals for confidence intervals
	stdErr := CalculateStdError(historicalScores, smoothed)

	// z-value for 95% confidence interval
	const z95 = 1.96

	prediction := &RiskPrediction{
		RiskID:         riskID,
		CurrentScore:   currentScore,
		ModelVersion:   "ses-v1.0",
		SmoothingAlpha: alpha,
		ConfidenceLevel: 0.95,
		InputFeatures: map[string]float64{
			"historical_data_points": float64(len(historicalScores)),
			"current_score":          currentScore,
			"smoothed_latest":        smoothed[len(smoothed)-1],
			"std_error":              stdErr,
		},
		GeneratedAt: time.Now().UTC(),
	}

	lastSmoothed := smoothed[len(smoothed)-1]

	// Forecast at +30, +60, +90 days
	for _, days := range []int{30, 60, 90} {
		// For SES, forecast is the last smoothed value (flat forecast)
		// Confidence interval widens with horizon
		horizonFactor := math.Sqrt(float64(days) / 30.0)
		ciWidth := z95 * stdErr * horizonFactor

		low := math.Max(0, lastSmoothed-ciWidth)
		high := math.Min(25, lastSmoothed+ciWidth) // risk scores cap at 25

		prediction.Forecasts = append(prediction.Forecasts, ForecastPoint{
			Date:                   time.Now().UTC().AddDate(0, 0, days),
			PredictedValue:         math.Round(lastSmoothed*10000) / 10000,
			ConfidenceIntervalLow:  math.Round(low*10000) / 10000,
			ConfidenceIntervalHigh: math.Round(high*10000) / 10000,
			DaysAhead:              days,
		})
	}

	// Store predictions in the database
	for _, fc := range prediction.Forecasts {
		featJSON, _ := json.Marshal(prediction.InputFeatures)
		_, err := e.pool.Exec(ctx, `
			INSERT INTO analytics_risk_predictions
				(organization_id, risk_id, prediction_date, prediction_type,
				 predicted_value, confidence_interval_low, confidence_interval_high,
				 confidence_level, model_version, input_features)
			VALUES ($1, $2, $3, 'score_forecast', $4, $5, $6, $7, $8, $9)`,
			orgID, riskID, fc.Date, fc.PredictedValue,
			fc.ConfidenceIntervalLow, fc.ConfidenceIntervalHigh,
			prediction.ConfidenceLevel, prediction.ModelVersion, featJSON,
		)
		if err != nil {
			log.Error().Err(err).Msg("Failed to store risk prediction")
		}
	}

	return prediction, nil
}

// getRiskHistoricalScores extracts historical average risk scores from daily snapshots.
func (e *AnalyticsEngine) getRiskHistoricalScores(ctx context.Context, orgID, riskID uuid.UUID) []float64 {
	rows, err := e.pool.Query(ctx, `
		SELECT (metrics->>'avg_risk_score')::FLOAT
		FROM analytics_snapshots
		WHERE organization_id = $1 AND snapshot_type = 'daily'
		ORDER BY snapshot_date ASC
		LIMIT 90`, orgID)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var scores []float64
	for rows.Next() {
		var s float64
		if err := rows.Scan(&s); err != nil {
			continue
		}
		scores = append(scores, s)
	}
	return scores
}

// ExponentialSmoothing applies simple exponential smoothing to a time series.
// s_t = alpha * y_t + (1 - alpha) * s_{t-1}
func ExponentialSmoothing(values []float64, alpha float64) []float64 {
	if len(values) == 0 {
		return nil
	}
	smoothed := make([]float64, len(values))
	smoothed[0] = values[0]
	for i := 1; i < len(values); i++ {
		smoothed[i] = alpha*values[i] + (1-alpha)*smoothed[i-1]
	}
	return smoothed
}

// CalculateStdError computes the standard error of the residuals between
// observed and smoothed values.
func CalculateStdError(observed, smoothed []float64) float64 {
	if len(observed) < 2 || len(observed) != len(smoothed) {
		return 1.0 // default when insufficient data
	}

	var sumSqResiduals float64
	for i := range observed {
		residual := observed[i] - smoothed[i]
		sumSqResiduals += residual * residual
	}
	mse := sumSqResiduals / float64(len(observed)-1)
	return math.Sqrt(mse)
}

// PredictBreachProbability estimates the probability of a data breach using
// a logistic-style model based on risk counts, severity distribution,
// incident history, and control coverage.
func (e *AnalyticsEngine) PredictBreachProbability(ctx context.Context, orgID uuid.UUID) (*BreachPrediction, error) {
	features := make(map[string]float64)

	// Gather input features

	// 1. Risk count by severity
	var critRisks, highRisks, medRisks, lowRisks int64
	e.pool.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE residual_risk_level = 'critical'),
			COUNT(*) FILTER (WHERE residual_risk_level = 'high'),
			COUNT(*) FILTER (WHERE residual_risk_level = 'medium'),
			COUNT(*) FILTER (WHERE residual_risk_level = 'low')
		FROM risks WHERE organization_id = $1 AND status != 'closed' AND deleted_at IS NULL`,
		orgID,
	).Scan(&critRisks, &highRisks, &medRisks, &lowRisks)
	features["critical_risks"] = float64(critRisks)
	features["high_risks"] = float64(highRisks)
	features["medium_risks"] = float64(medRisks)
	features["low_risks"] = float64(lowRisks)

	// 2. Past incident count (last 12 months)
	var incidentCount12m, breachCount12m int64
	e.pool.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE is_data_breach = true)
		FROM incidents
		WHERE organization_id = $1 AND created_at >= NOW() - INTERVAL '12 months'
			AND deleted_at IS NULL`,
		orgID,
	).Scan(&incidentCount12m, &breachCount12m)
	features["incidents_12m"] = float64(incidentCount12m)
	features["breaches_12m"] = float64(breachCount12m)

	// 3. Control coverage
	var totalControls, implControls int64
	e.pool.QueryRow(ctx, `
		SELECT COUNT(*), COUNT(*) FILTER (WHERE status IN ('implemented', 'effective'))
		FROM control_implementations
		WHERE organization_id = $1 AND deleted_at IS NULL`, orgID,
	).Scan(&totalControls, &implControls)
	controlCoverage := 0.0
	if totalControls > 0 {
		controlCoverage = float64(implControls) / float64(totalControls)
	}
	features["control_coverage"] = controlCoverage

	// 4. Open critical/high findings
	var critFindings, highFindings int64
	e.pool.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE severity = 'critical'),
			COUNT(*) FILTER (WHERE severity = 'high')
		FROM audit_findings
		WHERE organization_id = $1 AND status IN ('open', 'in_progress')
			AND deleted_at IS NULL`, orgID,
	).Scan(&critFindings, &highFindings)
	features["critical_findings"] = float64(critFindings)
	features["high_findings"] = float64(highFindings)

	// 5. Average compliance score
	var avgCompliance float64
	e.pool.QueryRow(ctx, `
		SELECT COALESCE(AVG(compliance_score), 0)
		FROM organization_frameworks
		WHERE organization_id = $1`, orgID,
	).Scan(&avgCompliance)
	features["avg_compliance_score"] = avgCompliance

	// Logistic-style risk score calculation
	// Higher score = higher risk of breach
	riskFactors := []RiskFactor{
		{
			Name:   "Critical Risks",
			Weight: 0.20,
			Value:  features["critical_risks"],
			Description: "Number of open critical-level risks",
		},
		{
			Name:   "High Risks",
			Weight: 0.12,
			Value:  features["high_risks"],
			Description: "Number of open high-level risks",
		},
		{
			Name:   "Past Breaches",
			Weight: 0.18,
			Value:  features["breaches_12m"],
			Description: "Data breaches in the last 12 months",
		},
		{
			Name:   "Incident Volume",
			Weight: 0.10,
			Value:  features["incidents_12m"],
			Description: "Total security incidents in the last 12 months",
		},
		{
			Name:   "Control Coverage Gap",
			Weight: 0.15,
			Value:  1 - controlCoverage,
			Description: "Proportion of controls not fully implemented",
		},
		{
			Name:   "Critical Findings",
			Weight: 0.12,
			Value:  features["critical_findings"],
			Description: "Open critical audit findings",
		},
		{
			Name:   "Compliance Gap",
			Weight: 0.13,
			Value:  math.Max(0, 100-avgCompliance) / 100.0,
			Description: "Gap from full compliance (normalized)",
		},
	}

	// Compute weighted linear combination, then apply sigmoid
	var linearScore float64
	for i := range riskFactors {
		// Normalize values: cap at reasonable maximums to keep in [0,1] range
		normalizedValue := riskFactors[i].Value
		switch riskFactors[i].Name {
		case "Critical Risks":
			normalizedValue = math.Min(normalizedValue/10.0, 1.0)
		case "High Risks":
			normalizedValue = math.Min(normalizedValue/20.0, 1.0)
		case "Past Breaches":
			normalizedValue = math.Min(normalizedValue/5.0, 1.0)
		case "Incident Volume":
			normalizedValue = math.Min(normalizedValue/50.0, 1.0)
		case "Critical Findings":
			normalizedValue = math.Min(normalizedValue/10.0, 1.0)
		}

		contribution := riskFactors[i].Weight * normalizedValue
		riskFactors[i].Contribution = math.Round(contribution*10000) / 10000
		linearScore += contribution
	}

	// Sigmoid: P = 1 / (1 + e^(-k*(x-0.5)))
	// k controls steepness; shift by 0.5 to center
	sigmoid := func(x, k float64) float64 {
		return 1.0 / (1.0 + math.Exp(-k*(x-0.5)))
	}

	prob30d := sigmoid(linearScore, 6.0)
	prob90d := 1.0 - math.Pow(1.0-prob30d, 3.0)    // compound over 3 periods
	prob365d := 1.0 - math.Pow(1.0-prob30d, 12.0)   // compound over 12 periods

	prediction := &BreachPrediction{
		OrganizationID:        orgID,
		BreachProbability30d:  math.Round(prob30d*10000) / 10000,
		BreachProbability90d:  math.Round(prob90d*10000) / 10000,
		BreachProbability365d: math.Round(prob365d*10000) / 10000,
		ConfidenceLevel:       0.95,
		RiskFactors:           riskFactors,
		ModelVersion:          "logistic-v1.0",
		InputFeatures:         features,
		GeneratedAt:           time.Now().UTC(),
	}

	// Store prediction
	featJSON, _ := json.Marshal(features)
	_, err := e.pool.Exec(ctx, `
		INSERT INTO analytics_risk_predictions
			(organization_id, prediction_date, prediction_type,
			 predicted_value, confidence_interval_low, confidence_interval_high,
			 confidence_level, model_version, input_features)
		VALUES ($1, CURRENT_DATE, 'breach_probability', $2, $3, $4, $5, $6, $7)`,
		orgID, prediction.BreachProbability30d,
		math.Max(0, prediction.BreachProbability30d-0.05),
		math.Min(1, prediction.BreachProbability30d+0.05),
		prediction.ConfidenceLevel,
		prediction.ModelVersion, featJSON,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to store breach prediction")
	}

	return prediction, nil
}

// ValidateModel compares past predictions with actual observed values to
// measure model accuracy.
func (e *AnalyticsEngine) ValidateModel(ctx context.Context, orgID uuid.UUID) (*ModelValidation, error) {
	rows, err := e.pool.Query(ctx, `
		SELECT prediction_type::TEXT, predicted_value, actual_value,
			confidence_interval_low, confidence_interval_high
		FROM analytics_risk_predictions
		WHERE organization_id = $1 AND actual_value IS NOT NULL
		ORDER BY created_at DESC
		LIMIT 1000`, orgID)
	if err != nil {
		return nil, fmt.Errorf("query predictions for validation: %w", err)
	}
	defer rows.Close()

	validation := &ModelValidation{
		OrganizationID: orgID,
		AccuracyByType: make(map[string]TypeAccuracy),
		ModelVersion:   "ses-v1.0",
		ValidatedAt:    time.Now().UTC(),
	}

	type predRecord struct {
		predType string
		predicted, actual, ciLow, ciHigh float64
	}

	var records []predRecord
	for rows.Next() {
		var r predRecord
		if err := rows.Scan(&r.predType, &r.predicted, &r.actual, &r.ciLow, &r.ciHigh); err != nil {
			continue
		}
		records = append(records, r)
	}

	validation.TotalPredictions = len(records)
	validation.ValidatedPredictions = len(records)

	if len(records) == 0 {
		return validation, nil
	}

	// Aggregate errors by type
	typeErrors := make(map[string][]float64)
	typeCIHits := make(map[string]int)
	typeCounts := make(map[string]int)

	var totalMAE, totalMSE float64
	for _, r := range records {
		absErr := math.Abs(r.predicted - r.actual)
		sqErr := (r.predicted - r.actual) * (r.predicted - r.actual)
		totalMAE += absErr
		totalMSE += sqErr

		typeErrors[r.predType] = append(typeErrors[r.predType], absErr)
		typeCounts[r.predType]++
		if r.actual >= r.ciLow && r.actual <= r.ciHigh {
			typeCIHits[r.predType]++
		}
	}

	n := float64(len(records))
	validation.MeanAbsoluteError = math.Round(totalMAE/n*10000) / 10000
	validation.MeanSquaredError = math.Round(totalMSE/n*10000) / 10000

	for pType, errors := range typeErrors {
		var sumAbs float64
		for _, e := range errors {
			sumAbs += e
		}
		count := typeCounts[pType]
		ciRate := 0.0
		if count > 0 {
			ciRate = float64(typeCIHits[pType]) / float64(count)
		}
		validation.AccuracyByType[pType] = TypeAccuracy{
			PredictionType:       pType,
			Count:                count,
			MeanAbsoluteError:    math.Round(sumAbs/float64(count)*10000) / 10000,
			WithinConfidenceRate: math.Round(ciRate*10000) / 10000,
		}
	}

	return validation, nil
}

// ============================================================
// BENCHMARK ENGINE
// ============================================================

// orgMetric holds anonymized per-organization metrics for benchmark aggregation.
type orgMetric struct {
	Industry string
	Size     string
	Region   string
	Metrics  SnapshotMetrics
}

// CalculateBenchmarks aggregates anonymized metrics across all organizations
// and computes percentiles by category. No individual org data is exposed.
func (e *AnalyticsEngine) CalculateBenchmarks(ctx context.Context) error {
	log.Info().Msg("Calculating analytics benchmarks")

	// Collect per-org aggregate metrics from the most recent daily snapshots
	rows, err := e.pool.Query(ctx, `
		SELECT DISTINCT ON (s.organization_id)
			s.organization_id,
			o.industry,
			o.company_size,
			o.country,
			s.metrics
		FROM analytics_snapshots s
		JOIN organizations o ON s.organization_id = o.id
		WHERE s.snapshot_type = 'daily'
		ORDER BY s.organization_id, s.snapshot_date DESC`)
	if err != nil {
		return fmt.Errorf("query org snapshots for benchmarks: %w", err)
	}
	defer rows.Close()

	var allOrgs []orgMetric
	for rows.Next() {
		var orgID uuid.UUID
		var industry, size, country string
		var metricsJSON []byte
		if err := rows.Scan(&orgID, &industry, &size, &country, &metricsJSON); err != nil {
			log.Error().Err(err).Msg("Scan benchmark org data")
			continue
		}
		var sm SnapshotMetrics
		if err := json.Unmarshal(metricsJSON, &sm); err != nil {
			continue
		}
		allOrgs = append(allOrgs, orgMetric{
			Industry: industry,
			Size:     size,
			Region:   country,
			Metrics:  sm,
		})
	}

	if len(allOrgs) < 5 {
		log.Info().Int("org_count", len(allOrgs)).Msg("Insufficient organizations for benchmarking (minimum 5)")
		return nil
	}

	// Define metrics to benchmark
	type metricExtractor func(SnapshotMetrics) float64
	metricDefs := map[string]metricExtractor{
		"compliance_score":     func(m SnapshotMetrics) float64 { return m.OverallComplianceScore },
		"control_coverage":     func(m SnapshotMetrics) float64 { return m.ControlCoverageRate },
		"avg_maturity":         func(m SnapshotMetrics) float64 { return m.AvgMaturityLevel },
		"total_risks":          func(m SnapshotMetrics) float64 { return float64(m.TotalRisks) },
		"open_incidents":       func(m SnapshotMetrics) float64 { return float64(m.OpenIncidents) },
		"open_findings":        func(m SnapshotMetrics) float64 { return float64(m.OpenFindings) },
		"avg_risk_score":       func(m SnapshotMetrics) float64 { return m.AvgRiskScore },
		"policies_due_review":  func(m SnapshotMetrics) float64 { return float64(m.PoliciesDueReview) },
	}

	period := time.Now().UTC().Format("2006-01")

	// Calculate overall benchmarks
	for metricName, extractor := range metricDefs {
		values := make([]float64, 0, len(allOrgs))
		for _, org := range allOrgs {
			values = append(values, extractor(org.Metrics))
		}
		e.storeBenchmark(ctx, "overall", "all", metricName, period, values)
	}

	// Calculate benchmarks by industry
	industryGroups := groupBy(allOrgs, func(o orgMetric) string { return o.Industry })
	for industry, orgs := range industryGroups {
		if len(orgs) < 3 {
			continue // not enough data for anonymity
		}
		for metricName, extractor := range metricDefs {
			values := make([]float64, 0, len(orgs))
			for _, org := range orgs {
				values = append(values, extractor(org.Metrics))
			}
			e.storeBenchmark(ctx, "industry", industry, metricName, period, values)
		}
	}

	// Calculate benchmarks by size
	sizeGroups := groupBy(allOrgs, func(o orgMetric) string { return o.Size })
	for sz, orgs := range sizeGroups {
		if len(orgs) < 3 {
			continue
		}
		for metricName, extractor := range metricDefs {
			values := make([]float64, 0, len(orgs))
			for _, org := range orgs {
				values = append(values, extractor(org.Metrics))
			}
			e.storeBenchmark(ctx, "size", sz, metricName, period, values)
		}
	}

	log.Info().Int("total_orgs", len(allOrgs)).Msg("Benchmark calculation complete")
	return nil
}

// storeBenchmark computes percentiles from a sorted list and upserts the benchmark row.
func (e *AnalyticsEngine) storeBenchmark(ctx context.Context, bmType, category, metricName, period string, values []float64) {
	if len(values) == 0 {
		return
	}
	sortFloat64s(values)

	p25 := Percentile(values, 25)
	p50 := Percentile(values, 50)
	p75 := Percentile(values, 75)
	p90 := Percentile(values, 90)

	_, err := e.pool.Exec(ctx, `
		INSERT INTO analytics_benchmarks
			(benchmark_type, category, metric_name, period,
			 percentile_25, percentile_50, percentile_75, percentile_90,
			 sample_size, calculated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())`,
		bmType, category, metricName, period,
		p25, p50, p75, p90, len(values),
	)
	if err != nil {
		log.Error().Err(err).Str("metric", metricName).Msg("Failed to store benchmark")
	}
}

// CompareToPeers returns the organization's percentile position relative
// to anonymized peer benchmarks for each key metric.
func (e *AnalyticsEngine) CompareToPeers(ctx context.Context, orgID uuid.UUID) (*PeerComparison, error) {
	// Get org's latest snapshot metrics
	var metricsJSON []byte
	err := e.pool.QueryRow(ctx, `
		SELECT metrics FROM analytics_snapshots
		WHERE organization_id = $1 AND snapshot_type = 'daily'
		ORDER BY snapshot_date DESC LIMIT 1`, orgID,
	).Scan(&metricsJSON)
	if err != nil {
		return nil, fmt.Errorf("query org snapshot: %w", err)
	}

	var orgMetrics SnapshotMetrics
	if err := json.Unmarshal(metricsJSON, &orgMetrics); err != nil {
		return nil, fmt.Errorf("unmarshal org metrics: %w", err)
	}

	// Get the latest overall benchmarks
	rows, err := e.pool.Query(ctx, `
		SELECT DISTINCT ON (metric_name)
			metric_name, percentile_25, percentile_50, percentile_75,
			percentile_90, sample_size
		FROM analytics_benchmarks
		WHERE benchmark_type = 'overall'
		ORDER BY metric_name, calculated_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("query benchmarks: %w", err)
	}
	defer rows.Close()

	comparison := &PeerComparison{
		OrganizationID:  orgID,
		BenchmarkPeriod: time.Now().UTC().Format("2006-01"),
		GeneratedAt:     time.Now().UTC(),
	}

	metricValues := map[string]float64{
		"compliance_score":     orgMetrics.OverallComplianceScore,
		"control_coverage":     orgMetrics.ControlCoverageRate,
		"avg_maturity":         orgMetrics.AvgMaturityLevel,
		"total_risks":          float64(orgMetrics.TotalRisks),
		"open_incidents":       float64(orgMetrics.OpenIncidents),
		"open_findings":        float64(orgMetrics.OpenFindings),
		"avg_risk_score":       orgMetrics.AvgRiskScore,
		"policies_due_review":  float64(orgMetrics.PoliciesDueReview),
	}

	for rows.Next() {
		var pm PeerMetric
		if err := rows.Scan(
			&pm.MetricName, &pm.Percentile25, &pm.Percentile50,
			&pm.Percentile75, &pm.Percentile90, &pm.SampleSize,
		); err != nil {
			continue
		}

		if val, ok := metricValues[pm.MetricName]; ok {
			pm.OrgValue = val
			pm.PercentilePos = estimatePercentilePosition(val, pm.Percentile25, pm.Percentile50, pm.Percentile75, pm.Percentile90)
		}

		comparison.Metrics = append(comparison.Metrics, pm)
		comparison.SampleSize = pm.SampleSize
	}

	return comparison, nil
}

// estimatePercentilePosition estimates where a value falls within the benchmark distribution.
func estimatePercentilePosition(value, p25, p50, p75, p90 float64) float64 {
	switch {
	case value <= p25:
		if p25 == 0 {
			return 12.5
		}
		return math.Min(25, value/p25*25)
	case value <= p50:
		return 25 + (value-p25)/(p50-p25)*25
	case value <= p75:
		return 50 + (value-p50)/(p75-p50)*25
	case value <= p90:
		return 75 + (value-p75)/(p90-p75)*15
	default:
		return math.Min(99, 90+(value-p90)/(p90*0.5)*10)
	}
}

// ============================================================
// MATH HELPERS
// ============================================================

// Percentile calculates the p-th percentile of a sorted slice using linear interpolation.
func Percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	if len(sorted) == 1 {
		return sorted[0]
	}

	rank := p / 100.0 * float64(len(sorted)-1)
	lower := int(math.Floor(rank))
	upper := int(math.Ceil(rank))

	if lower == upper || upper >= len(sorted) {
		return sorted[lower]
	}

	frac := rank - float64(lower)
	return sorted[lower] + frac*(sorted[upper]-sorted[lower])
}

// sortFloat64s sorts a float64 slice in ascending order using insertion sort
// (suitable for the small dataset sizes used in benchmarking).
func sortFloat64s(a []float64) {
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

// groupBy groups a slice of orgMetric by a key extractor function.
func groupBy(orgs []orgMetric, keyFn func(orgMetric) string) map[string][]orgMetric {
	groups := make(map[string][]orgMetric)
	for _, org := range orgs {
		key := keyFn(org)
		if key != "" {
			groups[key] = append(groups[key], org)
		}
	}
	return groups
}
