package service

import (
	"math"
	"testing"
	"time"

	"github.com/google/uuid"
)

// ============================================================
// LINEAR REGRESSION TESTS
// ============================================================

func TestLinearRegressionSlope_Increasing(t *testing.T) {
	values := []float64{10, 20, 30, 40, 50}
	slope := LinearRegressionSlope(values)

	if math.Abs(slope-10.0) > 0.001 {
		t.Errorf("Expected slope ~10.0 for linearly increasing data, got %f", slope)
	}
}

func TestLinearRegressionSlope_Decreasing(t *testing.T) {
	values := []float64{50, 40, 30, 20, 10}
	slope := LinearRegressionSlope(values)

	if math.Abs(slope-(-10.0)) > 0.001 {
		t.Errorf("Expected slope ~-10.0 for linearly decreasing data, got %f", slope)
	}
}

func TestLinearRegressionSlope_Flat(t *testing.T) {
	values := []float64{42, 42, 42, 42, 42}
	slope := LinearRegressionSlope(values)

	if math.Abs(slope) > 0.001 {
		t.Errorf("Expected slope ~0 for constant data, got %f", slope)
	}
}

func TestLinearRegressionSlope_SinglePoint(t *testing.T) {
	values := []float64{42}
	slope := LinearRegressionSlope(values)

	if slope != 0 {
		t.Errorf("Expected slope 0 for single point, got %f", slope)
	}
}

func TestLinearRegressionSlope_Empty(t *testing.T) {
	slope := LinearRegressionSlope(nil)

	if slope != 0 {
		t.Errorf("Expected slope 0 for empty data, got %f", slope)
	}
}

func TestLinearRegressionSlope_NoisyIncreasing(t *testing.T) {
	// Noisy upward trend
	values := []float64{10, 12, 9, 15, 14, 18, 16, 20, 19, 22}
	slope := LinearRegressionSlope(values)

	if slope <= 0 {
		t.Errorf("Expected positive slope for noisy increasing data, got %f", slope)
	}
}

func TestLinearRegressionSlope_TwoPoints(t *testing.T) {
	values := []float64{0, 100}
	slope := LinearRegressionSlope(values)

	if math.Abs(slope-100.0) > 0.001 {
		t.Errorf("Expected slope ~100 for two-point data, got %f", slope)
	}
}

// ============================================================
// EXPONENTIAL SMOOTHING TESTS
// ============================================================

func TestExponentialSmoothing_Basic(t *testing.T) {
	values := []float64{10, 20, 30, 40, 50}
	alpha := 0.3
	smoothed := ExponentialSmoothing(values, alpha)

	if len(smoothed) != len(values) {
		t.Fatalf("Expected %d smoothed values, got %d", len(values), len(smoothed))
	}

	// First value should equal first observed value
	if smoothed[0] != values[0] {
		t.Errorf("First smoothed value should equal first observed: %f != %f", smoothed[0], values[0])
	}

	// Verify SES formula: s_t = alpha * y_t + (1 - alpha) * s_{t-1}
	for i := 1; i < len(values); i++ {
		expected := alpha*values[i] + (1-alpha)*smoothed[i-1]
		if math.Abs(smoothed[i]-expected) > 0.0001 {
			t.Errorf("Smoothed[%d]: expected %f, got %f", i, expected, smoothed[i])
		}
	}
}

func TestExponentialSmoothing_AlphaZero(t *testing.T) {
	values := []float64{10, 20, 30, 40, 50}
	smoothed := ExponentialSmoothing(values, 0.0)

	// With alpha=0, smoothed values should all equal the first value
	for i, s := range smoothed {
		if s != values[0] {
			t.Errorf("With alpha=0, smoothed[%d] should be %f, got %f", i, values[0], s)
		}
	}
}

func TestExponentialSmoothing_AlphaOne(t *testing.T) {
	values := []float64{10, 20, 30, 40, 50}
	smoothed := ExponentialSmoothing(values, 1.0)

	// With alpha=1, smoothed values should equal observed values
	for i, s := range smoothed {
		if s != values[i] {
			t.Errorf("With alpha=1, smoothed[%d] should be %f, got %f", i, values[i], s)
		}
	}
}

func TestExponentialSmoothing_Empty(t *testing.T) {
	smoothed := ExponentialSmoothing(nil, 0.3)
	if smoothed != nil {
		t.Errorf("Expected nil for empty input, got %v", smoothed)
	}
}

func TestExponentialSmoothing_SingleValue(t *testing.T) {
	values := []float64{42.0}
	smoothed := ExponentialSmoothing(values, 0.3)

	if len(smoothed) != 1 || smoothed[0] != 42.0 {
		t.Errorf("Expected [42.0], got %v", smoothed)
	}
}

func TestExponentialSmoothing_SmoothingEffect(t *testing.T) {
	// Verify that smoothing reduces variance
	values := []float64{10, 50, 10, 50, 10, 50, 10, 50}
	smoothed := ExponentialSmoothing(values, 0.3)

	// Calculate variance of original vs smoothed
	origVar := variance(values)
	smoothedVar := variance(smoothed)

	if smoothedVar >= origVar {
		t.Errorf("Smoothed variance (%f) should be less than original (%f)", smoothedVar, origVar)
	}
}

// ============================================================
// STANDARD ERROR TESTS
// ============================================================

func TestCalculateStdError_Perfect(t *testing.T) {
	observed := []float64{10, 20, 30, 40, 50}
	// If smoothed equals observed, std error should be 0
	stdErr := CalculateStdError(observed, observed)

	if stdErr != 0 {
		t.Errorf("Expected 0 std error for perfect fit, got %f", stdErr)
	}
}

func TestCalculateStdError_WithResiduals(t *testing.T) {
	observed := []float64{10, 20, 30, 40, 50}
	smoothed := []float64{12, 18, 32, 38, 52}
	stdErr := CalculateStdError(observed, smoothed)

	if stdErr <= 0 {
		t.Errorf("Expected positive std error with residuals, got %f", stdErr)
	}
}

func TestCalculateStdError_InsufficientData(t *testing.T) {
	observed := []float64{10}
	smoothed := []float64{10}
	stdErr := CalculateStdError(observed, smoothed)

	if stdErr != 1.0 {
		t.Errorf("Expected default 1.0 for insufficient data, got %f", stdErr)
	}
}

func TestCalculateStdError_MismatchedLengths(t *testing.T) {
	observed := []float64{10, 20, 30}
	smoothed := []float64{10, 20}
	stdErr := CalculateStdError(observed, smoothed)

	if stdErr != 1.0 {
		t.Errorf("Expected default 1.0 for mismatched lengths, got %f", stdErr)
	}
}

// ============================================================
// PERCENTILE TESTS
// ============================================================

func TestPercentile_Median(t *testing.T) {
	sorted := []float64{10, 20, 30, 40, 50}
	p50 := Percentile(sorted, 50)

	if p50 != 30 {
		t.Errorf("Expected median 30, got %f", p50)
	}
}

func TestPercentile_25th(t *testing.T) {
	sorted := []float64{10, 20, 30, 40, 50}
	p25 := Percentile(sorted, 25)

	if p25 != 20 {
		t.Errorf("Expected P25 = 20, got %f", p25)
	}
}

func TestPercentile_75th(t *testing.T) {
	sorted := []float64{10, 20, 30, 40, 50}
	p75 := Percentile(sorted, 75)

	if p75 != 40 {
		t.Errorf("Expected P75 = 40, got %f", p75)
	}
}

func TestPercentile_90th(t *testing.T) {
	sorted := []float64{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}
	p90 := Percentile(sorted, 90)

	expected := 91.0 // linear interpolation: 90 + 0.1*(100-90) = 91
	if math.Abs(p90-expected) > 0.01 {
		t.Errorf("Expected P90 ~ %f, got %f", expected, p90)
	}
}

func TestPercentile_Empty(t *testing.T) {
	p := Percentile(nil, 50)
	if p != 0 {
		t.Errorf("Expected 0 for empty slice, got %f", p)
	}
}

func TestPercentile_SingleValue(t *testing.T) {
	p := Percentile([]float64{42.0}, 50)
	if p != 42.0 {
		t.Errorf("Expected 42 for single value, got %f", p)
	}
}

func TestPercentile_Extremes(t *testing.T) {
	sorted := []float64{10, 20, 30, 40, 50}

	p0 := Percentile(sorted, 0)
	if p0 != 10 {
		t.Errorf("Expected P0 = 10 (minimum), got %f", p0)
	}

	p100 := Percentile(sorted, 100)
	if p100 != 50 {
		t.Errorf("Expected P100 = 50 (maximum), got %f", p100)
	}
}

// ============================================================
// SORT TESTS
// ============================================================

func TestSortFloat64s(t *testing.T) {
	tests := []struct {
		name     string
		input    []float64
		expected []float64
	}{
		{"already sorted", []float64{1, 2, 3, 4, 5}, []float64{1, 2, 3, 4, 5}},
		{"reverse sorted", []float64{5, 4, 3, 2, 1}, []float64{1, 2, 3, 4, 5}},
		{"random order", []float64{3, 1, 4, 1, 5, 9, 2, 6}, []float64{1, 1, 2, 3, 4, 5, 6, 9}},
		{"duplicates", []float64{5, 5, 5, 5}, []float64{5, 5, 5, 5}},
		{"single element", []float64{42}, []float64{42}},
		{"empty", []float64{}, []float64{}},
		{"negative values", []float64{-3, -1, -4, -1, -5}, []float64{-5, -4, -3, -1, -1}},
		{"mixed sign", []float64{-2, 0, 3, -1, 2}, []float64{-2, -1, 0, 2, 3}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sortFloat64s(tc.input)
			if len(tc.input) != len(tc.expected) {
				t.Fatalf("Length mismatch: got %d, want %d", len(tc.input), len(tc.expected))
			}
			for i := range tc.expected {
				if tc.input[i] != tc.expected[i] {
					t.Errorf("Index %d: got %f, want %f", i, tc.input[i], tc.expected[i])
				}
			}
		})
	}
}

// ============================================================
// STRUCT CONSTRUCTION TESTS
// ============================================================

func TestComplianceTrend_Fields(t *testing.T) {
	orgID := uuid.New()
	fwID := uuid.New()
	trend := ComplianceTrend{
		ID:                  uuid.New(),
		OrganizationID:      orgID,
		FrameworkID:         fwID,
		FrameworkCode:       "ISO27001",
		MeasurementDate:     time.Now(),
		ComplianceScore:     85.50,
		ControlsImplemented: 120,
		ControlsTotal:       150,
		MaturityAvg:         3.25,
		ScoreChange7d:       1.5,
		ScoreChange30d:      5.0,
		ScoreChange90d:      12.0,
		TrendDirection:      "improving",
	}

	if trend.FrameworkCode != "ISO27001" {
		t.Errorf("Expected framework_code 'ISO27001', got '%s'", trend.FrameworkCode)
	}
	if trend.ComplianceScore != 85.50 {
		t.Errorf("Expected compliance_score 85.50, got %f", trend.ComplianceScore)
	}
	if trend.TrendDirection != "improving" {
		t.Errorf("Expected trend 'improving', got '%s'", trend.TrendDirection)
	}
	if trend.ControlsImplemented != 120 {
		t.Errorf("Expected 120 implemented controls, got %d", trend.ControlsImplemented)
	}
}

func TestRiskPrediction_Fields(t *testing.T) {
	riskID := uuid.New()
	prediction := RiskPrediction{
		RiskID:          riskID,
		CurrentScore:    12.5,
		ModelVersion:    "ses-v1.0",
		SmoothingAlpha:  0.3,
		ConfidenceLevel: 0.95,
		Forecasts: []ForecastPoint{
			{
				Date:                   time.Now().AddDate(0, 0, 30),
				PredictedValue:         11.8,
				ConfidenceIntervalLow:  8.2,
				ConfidenceIntervalHigh: 15.4,
				DaysAhead:              30,
			},
			{
				Date:                   time.Now().AddDate(0, 0, 60),
				PredictedValue:         11.8,
				ConfidenceIntervalLow:  6.5,
				ConfidenceIntervalHigh: 17.1,
				DaysAhead:              60,
			},
		},
	}

	if prediction.RiskID != riskID {
		t.Error("Risk ID mismatch")
	}
	if prediction.ConfidenceLevel != 0.95 {
		t.Errorf("Expected confidence 0.95, got %f", prediction.ConfidenceLevel)
	}
	if len(prediction.Forecasts) != 2 {
		t.Fatalf("Expected 2 forecasts, got %d", len(prediction.Forecasts))
	}
	if prediction.Forecasts[0].DaysAhead != 30 {
		t.Errorf("Expected 30 days ahead, got %d", prediction.Forecasts[0].DaysAhead)
	}
	if prediction.Forecasts[0].ConfidenceIntervalLow >= prediction.Forecasts[0].PredictedValue {
		t.Error("Confidence interval low should be less than predicted value")
	}
	if prediction.Forecasts[0].ConfidenceIntervalHigh <= prediction.Forecasts[0].PredictedValue {
		t.Error("Confidence interval high should be greater than predicted value")
	}
}

func TestBreachPrediction_Fields(t *testing.T) {
	orgID := uuid.New()
	prediction := BreachPrediction{
		OrganizationID:        orgID,
		BreachProbability30d:  0.12,
		BreachProbability90d:  0.32,
		BreachProbability365d: 0.78,
		ConfidenceLevel:       0.95,
		ModelVersion:          "logistic-v1.0",
		RiskFactors: []RiskFactor{
			{Name: "Critical Risks", Weight: 0.20, Value: 3, Contribution: 0.06},
			{Name: "Control Coverage Gap", Weight: 0.15, Value: 0.25, Contribution: 0.0375},
		},
	}

	if prediction.BreachProbability30d < 0 || prediction.BreachProbability30d > 1 {
		t.Errorf("Breach probability should be [0,1], got %f", prediction.BreachProbability30d)
	}
	if prediction.BreachProbability90d <= prediction.BreachProbability30d {
		t.Error("90-day probability should be greater than 30-day")
	}
	if prediction.BreachProbability365d <= prediction.BreachProbability90d {
		t.Error("365-day probability should be greater than 90-day")
	}
	if len(prediction.RiskFactors) != 2 {
		t.Fatalf("Expected 2 risk factors, got %d", len(prediction.RiskFactors))
	}
}

func TestModelValidation_Fields(t *testing.T) {
	orgID := uuid.New()
	validation := ModelValidation{
		OrganizationID:       orgID,
		TotalPredictions:     100,
		ValidatedPredictions: 85,
		MeanAbsoluteError:    1.5,
		MeanSquaredError:     3.2,
		ModelVersion:         "ses-v1.0",
		AccuracyByType: map[string]TypeAccuracy{
			"score_forecast": {
				PredictionType:       "score_forecast",
				Count:                60,
				MeanAbsoluteError:    1.2,
				WithinConfidenceRate: 0.92,
			},
			"breach_probability": {
				PredictionType:       "breach_probability",
				Count:                25,
				MeanAbsoluteError:    0.08,
				WithinConfidenceRate: 0.88,
			},
		},
	}

	if validation.TotalPredictions != 100 {
		t.Errorf("Expected 100 total predictions, got %d", validation.TotalPredictions)
	}
	if validation.MeanAbsoluteError <= 0 {
		t.Error("MAE should be positive")
	}
	if sf, ok := validation.AccuracyByType["score_forecast"]; ok {
		if sf.WithinConfidenceRate < 0 || sf.WithinConfidenceRate > 1 {
			t.Errorf("WithinConfidenceRate should be [0,1], got %f", sf.WithinConfidenceRate)
		}
	} else {
		t.Error("Expected 'score_forecast' in accuracy map")
	}
}

func TestPeerComparison_Fields(t *testing.T) {
	orgID := uuid.New()
	comparison := PeerComparison{
		OrganizationID:  orgID,
		BenchmarkPeriod: "2024-01",
		SampleSize:      150,
		Metrics: []PeerMetric{
			{
				MetricName:    "compliance_score",
				OrgValue:      82.5,
				Percentile25:  55.0,
				Percentile50:  68.0,
				Percentile75:  80.0,
				Percentile90:  92.0,
				PercentilePos: 76.5,
				SampleSize:    150,
			},
		},
	}

	if comparison.SampleSize != 150 {
		t.Errorf("Expected sample size 150, got %d", comparison.SampleSize)
	}
	if len(comparison.Metrics) != 1 {
		t.Fatalf("Expected 1 metric, got %d", len(comparison.Metrics))
	}
	m := comparison.Metrics[0]
	if m.PercentilePos < 0 || m.PercentilePos > 100 {
		t.Errorf("Percentile position should be [0,100], got %f", m.PercentilePos)
	}
}

// ============================================================
// PERCENTILE POSITION ESTIMATION TESTS
// ============================================================

func TestEstimatePercentilePosition_BelowP25(t *testing.T) {
	pos := estimatePercentilePosition(10, 20, 50, 70, 90)
	if pos > 25 {
		t.Errorf("Value below P25 should have position <= 25, got %f", pos)
	}
}

func TestEstimatePercentilePosition_AtP50(t *testing.T) {
	pos := estimatePercentilePosition(50, 20, 50, 70, 90)
	if math.Abs(pos-50) > 0.01 {
		t.Errorf("Value at P50 should have position ~50, got %f", pos)
	}
}

func TestEstimatePercentilePosition_AboveP90(t *testing.T) {
	pos := estimatePercentilePosition(95, 20, 50, 70, 90)
	if pos <= 90 {
		t.Errorf("Value above P90 should have position > 90, got %f", pos)
	}
}

func TestEstimatePercentilePosition_BetweenP50P75(t *testing.T) {
	pos := estimatePercentilePosition(60, 20, 50, 70, 90)
	if pos < 50 || pos > 75 {
		t.Errorf("Value between P50 and P75 should have position in [50,75], got %f", pos)
	}
}

// ============================================================
// SNAPSHOT METRICS STRUCT TESTS
// ============================================================

func TestSnapshotMetrics_DefaultMaps(t *testing.T) {
	metrics := SnapshotMetrics{
		RisksBySeverity:    make(map[string]int64),
		RisksByStatus:      make(map[string]int64),
		IncidentsByType:    make(map[string]int64),
		VendorsByRiskTier:  make(map[string]int64),
		FindingsBySeverity: make(map[string]int64),
		FrameworkScores:    make(map[string]float64),
	}

	metrics.RisksBySeverity["critical"] = 5
	metrics.RisksBySeverity["high"] = 10
	metrics.FrameworkScores["ISO27001"] = 82.5

	if metrics.RisksBySeverity["critical"] != 5 {
		t.Errorf("Expected 5 critical risks, got %d", metrics.RisksBySeverity["critical"])
	}
	if metrics.FrameworkScores["ISO27001"] != 82.5 {
		t.Errorf("Expected ISO27001 score 82.5, got %f", metrics.FrameworkScores["ISO27001"])
	}
}

func TestSnapshotMetrics_CoverageCalculation(t *testing.T) {
	metrics := SnapshotMetrics{
		TotalControls:       100,
		ImplementedControls: 75,
	}

	coverage := float64(metrics.ImplementedControls) / float64(metrics.TotalControls) * 100
	if coverage != 75.0 {
		t.Errorf("Expected 75%% coverage, got %f", coverage)
	}
}

func TestSnapshotMetrics_ZeroControls(t *testing.T) {
	metrics := SnapshotMetrics{
		TotalControls:       0,
		ImplementedControls: 0,
	}

	// Should not divide by zero
	coverage := 0.0
	if metrics.TotalControls > 0 {
		coverage = float64(metrics.ImplementedControls) / float64(metrics.TotalControls) * 100
	}

	if coverage != 0 {
		t.Errorf("Expected 0 coverage for zero controls, got %f", coverage)
	}
}

// ============================================================
// FORECAST POINT CONFIDENCE INTERVAL TESTS
// ============================================================

func TestForecastPoint_ConfidenceIntervalWidens(t *testing.T) {
	alpha := 0.3
	historical := []float64{10, 11, 12, 11, 13, 12, 14, 13, 15, 14}
	smoothed := ExponentialSmoothing(historical, alpha)
	stdErr := CalculateStdError(historical, smoothed)

	const z95 = 1.96
	lastSmoothed := smoothed[len(smoothed)-1]

	var forecasts []ForecastPoint
	for _, days := range []int{30, 60, 90} {
		horizonFactor := math.Sqrt(float64(days) / 30.0)
		ciWidth := z95 * stdErr * horizonFactor

		forecasts = append(forecasts, ForecastPoint{
			PredictedValue:         lastSmoothed,
			ConfidenceIntervalLow:  lastSmoothed - ciWidth,
			ConfidenceIntervalHigh: lastSmoothed + ciWidth,
			DaysAhead:              days,
		})
	}

	// Verify confidence intervals widen with horizon
	for i := 1; i < len(forecasts); i++ {
		prevWidth := forecasts[i-1].ConfidenceIntervalHigh - forecasts[i-1].ConfidenceIntervalLow
		currWidth := forecasts[i].ConfidenceIntervalHigh - forecasts[i].ConfidenceIntervalLow
		if currWidth <= prevWidth {
			t.Errorf("CI should widen: %d-day width (%f) <= %d-day width (%f)",
				forecasts[i].DaysAhead, currWidth,
				forecasts[i-1].DaysAhead, prevWidth)
		}
	}
}

// ============================================================
// BENCHMARK AGGREGATION TESTS
// ============================================================

func TestPercentile_LargerDataset(t *testing.T) {
	// Simulate 20 org compliance scores
	scores := []float64{
		32, 41, 45, 48, 52, 55, 58, 62, 65, 68,
		71, 73, 76, 78, 81, 84, 87, 90, 93, 97,
	}

	p25 := Percentile(scores, 25)
	p50 := Percentile(scores, 50)
	p75 := Percentile(scores, 75)
	p90 := Percentile(scores, 90)

	// Verify ordering
	if p25 >= p50 {
		t.Errorf("P25 (%f) should be < P50 (%f)", p25, p50)
	}
	if p50 >= p75 {
		t.Errorf("P50 (%f) should be < P75 (%f)", p50, p75)
	}
	if p75 >= p90 {
		t.Errorf("P75 (%f) should be < P90 (%f)", p75, p90)
	}

	// P50 should be around the middle of the dataset
	if p50 < 60 || p50 > 75 {
		t.Errorf("P50 should be roughly in the 60-75 range, got %f", p50)
	}
}

func TestGroupBy(t *testing.T) {
	orgs := []orgMetric{
		{Industry: "finance", Size: "large"},
		{Industry: "finance", Size: "small"},
		{Industry: "healthcare", Size: "large"},
		{Industry: "", Size: "small"}, // empty industry should be skipped
	}

	groups := groupBy(orgs, func(o orgMetric) string { return o.Industry })

	if len(groups) != 2 {
		t.Errorf("Expected 2 groups (finance, healthcare), got %d", len(groups))
	}
	if len(groups["finance"]) != 2 {
		t.Errorf("Expected 2 finance orgs, got %d", len(groups["finance"]))
	}
	if len(groups["healthcare"]) != 1 {
		t.Errorf("Expected 1 healthcare org, got %d", len(groups["healthcare"]))
	}
}

// ============================================================
// HELPERS
// ============================================================

func variance(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}
	mean := 0.0
	for _, v := range values {
		mean += v
	}
	mean /= float64(len(values))

	var sumSq float64
	for _, v := range values {
		diff := v - mean
		sumSq += diff * diff
	}
	return sumSq / float64(len(values)-1)
}
