package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math"
	"testing"
)

// ============================================================
// Token Hashing Tests
// ============================================================

func TestHashToken_Deterministic(t *testing.T) {
	token := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	hash1 := HashToken(token)
	hash2 := HashToken(token)
	if hash1 != hash2 {
		t.Errorf("HashToken not deterministic: %s != %s", hash1, hash2)
	}
}

func TestHashToken_DifferentTokensDifferentHashes(t *testing.T) {
	token1 := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	token2 := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	hash1 := HashToken(token1)
	hash2 := HashToken(token2)
	if hash1 == hash2 {
		t.Error("Different tokens produced the same hash")
	}
}

func TestHashToken_LengthIs64Hex(t *testing.T) {
	token := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	hash := HashToken(token)
	if len(hash) != 64 {
		t.Errorf("Expected hash length 64, got %d", len(hash))
	}
}

func TestHashToken_MatchesManualSHA256(t *testing.T) {
	token := "test_token_value"
	expected := sha256.Sum256([]byte(token))
	expectedHex := hex.EncodeToString(expected[:])
	got := HashToken(token)
	if got != expectedHex {
		t.Errorf("HashToken mismatch: expected %s, got %s", expectedHex, got)
	}
}

// ============================================================
// Score Calculation Tests (unit tests for scoring logic)
// ============================================================

// testSectionData simulates section data for scoring.
type testSectionData struct {
	Name   string
	Weight float64
	Score  float64
}

// computeWeightedAverage mirrors the scoring logic in CalculateScore.
func computeWeightedAverage(sections []testSectionData) float64 {
	totalWeight := 0.0
	weightedSum := 0.0
	for _, s := range sections {
		weightedSum += s.Score * s.Weight
		totalWeight += s.Weight
	}
	if totalWeight == 0 {
		return 0
	}
	return weightedSum / totalWeight
}

// computeSimpleAverage mirrors simple_average scoring.
func computeSimpleAverage(sections []testSectionData) float64 {
	if len(sections) == 0 {
		return 0
	}
	sum := 0.0
	for _, s := range sections {
		sum += s.Score
	}
	return sum / float64(len(sections))
}

// computeMinimumScore mirrors minimum_score scoring.
func computeMinimumScore(sections []testSectionData) float64 {
	if len(sections) == 0 {
		return 0
	}
	min := 100.0
	for _, s := range sections {
		if s.Score < min {
			min = s.Score
		}
	}
	return min
}

// determineRiskRating mirrors the risk rating logic.
func determineRiskRating(score float64, thresholds map[string]float64) string {
	if score >= thresholds["low"] {
		return "low"
	}
	if score >= thresholds["medium"] {
		return "medium"
	}
	if score >= thresholds["high"] {
		return "high"
	}
	return "critical"
}

// determinePassFail mirrors the pass/fail determination.
func determinePassFail(score, threshold float64, criticalFindings int) string {
	if score >= threshold {
		return "pass"
	}
	if score >= threshold*0.9 && criticalFindings == 0 {
		return "conditional_pass"
	}
	return "fail"
}

func TestWeightedAverage_EqualWeights(t *testing.T) {
	sections := []testSectionData{
		{Name: "A", Weight: 1.0, Score: 80},
		{Name: "B", Weight: 1.0, Score: 60},
		{Name: "C", Weight: 1.0, Score: 100},
	}
	result := computeWeightedAverage(sections)
	expected := 80.0
	if math.Abs(result-expected) > 0.01 {
		t.Errorf("Expected %.2f, got %.2f", expected, result)
	}
}

func TestWeightedAverage_DifferentWeights(t *testing.T) {
	sections := []testSectionData{
		{Name: "Access Control", Weight: 1.30, Score: 90},
		{Name: "Data Protection", Weight: 1.30, Score: 70},
		{Name: "Physical Security", Weight: 0.70, Score: 50},
	}
	// Expected: (90*1.30 + 70*1.30 + 50*0.70) / (1.30 + 1.30 + 0.70)
	// = (117 + 91 + 35) / 3.30 = 243/3.30 = 73.636...
	result := computeWeightedAverage(sections)
	expected := 243.0 / 3.30
	if math.Abs(result-expected) > 0.01 {
		t.Errorf("Expected %.2f, got %.2f", expected, result)
	}
}

func TestWeightedAverage_ZeroWeight(t *testing.T) {
	sections := []testSectionData{}
	result := computeWeightedAverage(sections)
	if result != 0 {
		t.Errorf("Expected 0 for empty sections, got %.2f", result)
	}
}

func TestSimpleAverage(t *testing.T) {
	sections := []testSectionData{
		{Name: "A", Weight: 1.0, Score: 80},
		{Name: "B", Weight: 2.0, Score: 60},
		{Name: "C", Weight: 0.5, Score: 40},
	}
	result := computeSimpleAverage(sections)
	expected := 60.0 // (80+60+40)/3
	if math.Abs(result-expected) > 0.01 {
		t.Errorf("Expected %.2f, got %.2f", expected, result)
	}
}

func TestSimpleAverage_Empty(t *testing.T) {
	result := computeSimpleAverage(nil)
	if result != 0 {
		t.Errorf("Expected 0 for empty sections, got %.2f", result)
	}
}

func TestMinimumScore(t *testing.T) {
	sections := []testSectionData{
		{Name: "A", Weight: 1.0, Score: 90},
		{Name: "B", Weight: 1.0, Score: 45},
		{Name: "C", Weight: 1.0, Score: 70},
	}
	result := computeMinimumScore(sections)
	expected := 45.0
	if math.Abs(result-expected) > 0.01 {
		t.Errorf("Expected %.2f, got %.2f", expected, result)
	}
}

func TestMinimumScore_Empty(t *testing.T) {
	result := computeMinimumScore(nil)
	if result != 0 {
		t.Errorf("Expected 0 for empty sections, got %.2f", result)
	}
}

func TestMinimumScore_AllSame(t *testing.T) {
	sections := []testSectionData{
		{Name: "A", Weight: 1.0, Score: 75},
		{Name: "B", Weight: 1.0, Score: 75},
	}
	result := computeMinimumScore(sections)
	if math.Abs(result-75.0) > 0.01 {
		t.Errorf("Expected 75.00, got %.2f", result)
	}
}

// ============================================================
// Risk Rating Tests
// ============================================================

func TestRiskRating_Critical(t *testing.T) {
	thresholds := map[string]float64{"critical": 40, "high": 55, "medium": 70, "low": 85}
	rating := determineRiskRating(30, thresholds)
	if rating != "critical" {
		t.Errorf("Expected 'critical' for score 30, got '%s'", rating)
	}
}

func TestRiskRating_High(t *testing.T) {
	thresholds := map[string]float64{"critical": 40, "high": 55, "medium": 70, "low": 85}
	rating := determineRiskRating(60, thresholds)
	if rating != "high" {
		t.Errorf("Expected 'high' for score 60, got '%s'", rating)
	}
}

func TestRiskRating_Medium(t *testing.T) {
	thresholds := map[string]float64{"critical": 40, "high": 55, "medium": 70, "low": 85}
	rating := determineRiskRating(75, thresholds)
	if rating != "medium" {
		t.Errorf("Expected 'medium' for score 75, got '%s'", rating)
	}
}

func TestRiskRating_Low(t *testing.T) {
	thresholds := map[string]float64{"critical": 40, "high": 55, "medium": 70, "low": 85}
	rating := determineRiskRating(90, thresholds)
	if rating != "low" {
		t.Errorf("Expected 'low' for score 90, got '%s'", rating)
	}
}

func TestRiskRating_ExactBoundary(t *testing.T) {
	thresholds := map[string]float64{"critical": 40, "high": 55, "medium": 70, "low": 85}
	// At exact threshold, should be that category
	rating := determineRiskRating(85, thresholds)
	if rating != "low" {
		t.Errorf("Expected 'low' for score exactly 85, got '%s'", rating)
	}
	rating = determineRiskRating(70, thresholds)
	if rating != "medium" {
		t.Errorf("Expected 'medium' for score exactly 70, got '%s'", rating)
	}
}

func TestRiskRating_Zero(t *testing.T) {
	thresholds := map[string]float64{"critical": 40, "high": 55, "medium": 70, "low": 85}
	rating := determineRiskRating(0, thresholds)
	if rating != "critical" {
		t.Errorf("Expected 'critical' for score 0, got '%s'", rating)
	}
}

func TestRiskRating_Perfect(t *testing.T) {
	thresholds := map[string]float64{"critical": 40, "high": 55, "medium": 70, "low": 85}
	rating := determineRiskRating(100, thresholds)
	if rating != "low" {
		t.Errorf("Expected 'low' for score 100, got '%s'", rating)
	}
}

// ============================================================
// Pass/Fail Tests
// ============================================================

func TestPassFail_Pass(t *testing.T) {
	result := determinePassFail(80, 70, 0)
	if result != "pass" {
		t.Errorf("Expected 'pass' for score 80 / threshold 70, got '%s'", result)
	}
}

func TestPassFail_Fail(t *testing.T) {
	result := determinePassFail(50, 70, 0)
	if result != "fail" {
		t.Errorf("Expected 'fail' for score 50 / threshold 70, got '%s'", result)
	}
}

func TestPassFail_ConditionalPass(t *testing.T) {
	// Score is >= 90% of threshold and no critical findings
	result := determinePassFail(65, 70, 0)
	if result != "conditional_pass" {
		t.Errorf("Expected 'conditional_pass' for score 65 / threshold 70 / 0 criticals, got '%s'", result)
	}
}

func TestPassFail_ConditionalPassBlockedByCriticals(t *testing.T) {
	// Score is in conditional range but has critical findings
	result := determinePassFail(65, 70, 1)
	if result != "fail" {
		t.Errorf("Expected 'fail' for score 65 / threshold 70 / 1 critical, got '%s'", result)
	}
}

func TestPassFail_ExactThreshold(t *testing.T) {
	result := determinePassFail(70, 70, 0)
	if result != "pass" {
		t.Errorf("Expected 'pass' for score exactly at threshold, got '%s'", result)
	}
}

func TestPassFail_JustBelowConditional(t *testing.T) {
	// 90% of 70 = 63. Score 62 should fail.
	result := determinePassFail(62, 70, 0)
	if result != "fail" {
		t.Errorf("Expected 'fail' for score 62 / threshold 70, got '%s'", result)
	}
}

// ============================================================
// Score-to-Answer Mapping Tests
// ============================================================

func TestOptionScoreLookup(t *testing.T) {
	optionsJSON := `[
		{"value":"yes_current","label":"Yes, reviewed within 12 months","score":100},
		{"value":"yes_outdated","label":"Yes, but not recently reviewed","score":60},
		{"value":"in_progress","label":"In development","score":30},
		{"value":"no","label":"No","score":0}
	]`

	var opts []struct {
		Value string  `json:"value"`
		Score float64 `json:"score"`
	}
	if err := json.Unmarshal([]byte(optionsJSON), &opts); err != nil {
		t.Fatalf("Failed to unmarshal options: %v", err)
	}

	scores := make(map[string]float64)
	for _, opt := range opts {
		scores[opt.Value] = opt.Score
	}

	tests := []struct {
		answer   string
		expected float64
		exists   bool
	}{
		{"yes_current", 100, true},
		{"yes_outdated", 60, true},
		{"in_progress", 30, true},
		{"no", 0, true},
		{"invalid", 0, false},
	}

	for _, tt := range tests {
		score, ok := scores[tt.answer]
		if ok != tt.exists {
			t.Errorf("Answer '%s': expected exists=%v, got %v", tt.answer, tt.exists, ok)
			continue
		}
		if ok && math.Abs(score-tt.expected) > 0.01 {
			t.Errorf("Answer '%s': expected score %.2f, got %.2f", tt.answer, tt.expected, score)
		}
	}
}

// ============================================================
// End-to-End Score Calculation Simulation
// ============================================================

func TestFullScoreCalculation(t *testing.T) {
	// Simulate a full assessment with 3 sections
	sections := []testSectionData{
		{Name: "Security Governance", Weight: 1.20, Score: 85},
		{Name: "Access Control", Weight: 1.30, Score: 72},
		{Name: "Data Protection", Weight: 1.30, Score: 60},
	}

	// Weighted average
	overall := computeWeightedAverage(sections)
	t.Logf("Overall score: %.2f", overall)

	// Risk rating
	thresholds := map[string]float64{"critical": 40, "high": 55, "medium": 70, "low": 85}
	rating := determineRiskRating(overall, thresholds)
	t.Logf("Risk rating: %s", rating)

	// Pass/fail with threshold 70
	passFail := determinePassFail(overall, 70, 0)
	t.Logf("Pass/fail: %s", passFail)

	// Verify reasonable values
	if overall < 60 || overall > 90 {
		t.Errorf("Overall score %.2f is out of expected range [60, 90]", overall)
	}
	if rating != "medium" && rating != "high" {
		t.Errorf("Expected 'medium' or 'high' risk rating, got '%s'", rating)
	}
	if passFail != "pass" && passFail != "conditional_pass" {
		t.Errorf("Expected 'pass' or 'conditional_pass', got '%s'", passFail)
	}
}

func TestFullScoreCalculation_FailingVendor(t *testing.T) {
	sections := []testSectionData{
		{Name: "Security Governance", Weight: 1.20, Score: 30},
		{Name: "Access Control", Weight: 1.30, Score: 20},
		{Name: "Data Protection", Weight: 1.30, Score: 40},
	}

	overall := computeWeightedAverage(sections)
	t.Logf("Overall score: %.2f", overall)

	thresholds := map[string]float64{"critical": 40, "high": 55, "medium": 70, "low": 85}
	rating := determineRiskRating(overall, thresholds)
	t.Logf("Risk rating: %s", rating)

	passFail := determinePassFail(overall, 70, 2)
	t.Logf("Pass/fail: %s", passFail)

	if overall > 40 {
		t.Errorf("Overall score %.2f is too high for failing vendor", overall)
	}
	if rating != "critical" {
		t.Errorf("Expected 'critical' risk rating, got '%s'", rating)
	}
	if passFail != "fail" {
		t.Errorf("Expected 'fail', got '%s'", passFail)
	}
}

func TestFullScoreCalculation_MinimumScoreMethod(t *testing.T) {
	sections := []testSectionData{
		{Name: "Security Governance", Weight: 1.0, Score: 90},
		{Name: "Access Control", Weight: 1.0, Score: 35},
		{Name: "Data Protection", Weight: 1.0, Score: 80},
	}

	// Minimum score method uses the worst section
	overall := computeMinimumScore(sections)
	t.Logf("Overall score (minimum): %.2f", overall)

	if math.Abs(overall-35) > 0.01 {
		t.Errorf("Expected 35.00, got %.2f", overall)
	}

	// With weighted average, the score would be much higher
	weighted := computeWeightedAverage(sections)
	t.Logf("Weighted average (for comparison): %.2f", weighted)

	if overall >= weighted {
		t.Error("Minimum score should be less than or equal to weighted average")
	}
}
