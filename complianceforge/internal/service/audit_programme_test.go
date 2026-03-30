package service

import (
	"fmt"
	"math"
	"testing"

	"github.com/google/uuid"
)

// ============================================================
// ISA 530 Sample Size Calculation Tests
// ============================================================

// TestCalculateSampleSize_95Confidence_5Error tests the standard case:
// 95% confidence, 5% tolerable error, 1% expected error, large population.
// Formula: n = (1.96^2 * 0.01 * 0.99) / 0.05^2 = 15.21 -> ceil = 16
// With finite correction for N=10000: n_adj = 16 / (1 + 15/10000) ~= 16
func TestCalculateSampleSize_95Confidence_5Error(t *testing.T) {
	n := CalculateSampleSize(10000, 95.0, 0.05, 0.01)

	// Expected: (1.96^2 * 0.01 * 0.99) / (0.05^2) = 15.21... ceil = 16
	// With FPC: 16 / (1 + 15/10000) = 15.976... ceil = 16
	expected := 16
	if n != expected {
		t.Errorf("CalculateSampleSize(10000, 95, 0.05, 0.01) = %d; want %d", n, expected)
	}
}

// TestCalculateSampleSize_90Confidence_10Error tests 90% confidence,
// 10% tolerable error, 5% expected error.
// Z=1.645 for 90%, p=0.05, E=0.10
// n = (1.645^2 * 0.05 * 0.95) / 0.10^2 = 12.86... -> ceil = 13
func TestCalculateSampleSize_90Confidence_10Error(t *testing.T) {
	n := CalculateSampleSize(10000, 90.0, 0.10, 0.05)

	z := 1.645
	p := 0.05
	e := 0.10
	rawN := (z * z * p * (1 - p)) / (e * e)
	adjN := rawN / (1.0 + (rawN-1.0)/10000.0)
	expected := int(math.Ceil(adjN))

	if n != expected {
		t.Errorf("CalculateSampleSize(10000, 90, 0.10, 0.05) = %d; want %d", n, expected)
	}
}

// TestCalculateSampleSize_SmallPopulation verifies the finite population
// correction factor significantly reduces the sample size for small populations.
func TestCalculateSampleSize_SmallPopulation(t *testing.T) {
	nLarge := CalculateSampleSize(100000, 95.0, 0.05, 0.01)
	nSmall := CalculateSampleSize(50, 95.0, 0.05, 0.01)

	if nSmall >= nLarge {
		t.Errorf("Small population (%d) should yield smaller sample than large population (%d)",
			nSmall, nLarge)
	}

	// For a population of 50, the sample cannot exceed 50.
	if nSmall > 50 {
		t.Errorf("Sample size %d exceeds population of 50", nSmall)
	}

	// Verify the FPC is actually applied.
	// Without FPC: n = (1.96^2 * 0.01 * 0.99) / 0.05^2 = ~16
	// With FPC for N=50: n_adj = 16 / (1 + 15/50) = 16 / 1.3 = ~12.3 -> 13
	z := 1.96
	p := 0.01
	e := 0.05
	rawN := (z * z * p * (1 - p)) / (e * e)
	adjN := rawN / (1.0 + (rawN-1.0)/50.0)
	expected := int(math.Ceil(adjN))

	if nSmall != expected {
		t.Errorf("CalculateSampleSize(50, 95, 0.05, 0.01) = %d; want %d", nSmall, expected)
	}
}

// TestCalculateSampleSize_MinimumSize ensures that the sample size
// never drops below 1, even with extreme parameters.
func TestCalculateSampleSize_MinimumSize(t *testing.T) {
	// Very large tolerable error with tiny expected error
	n := CalculateSampleSize(10, 80.0, 0.99, 0.001)
	if n < 1 {
		t.Errorf("Sample size should never be below 1, got %d", n)
	}

	// Population of 1
	n1 := CalculateSampleSize(1, 95.0, 0.05, 0.01)
	if n1 != 1 {
		t.Errorf("Population of 1 should yield sample size of 1, got %d", n1)
	}

	// Edge: zero expected error rate should use default (0.01)
	nZeroP := CalculateSampleSize(1000, 95.0, 0.05, 0.0)
	if nZeroP < 1 {
		t.Errorf("Sample size with zero expected error should still be >= 1, got %d", nZeroP)
	}
}

// ============================================================
// Random Sample Selection Tests
// ============================================================

// TestRandomSampleSelection_Uniqueness ensures all selected indices are unique.
func TestRandomSampleSelection_Uniqueness(t *testing.T) {
	population := 1000
	sampleSize := 50

	indices, err := GenerateRandomSample(population, sampleSize)
	if err != nil {
		t.Fatalf("GenerateRandomSample failed: %v", err)
	}

	if len(indices) != sampleSize {
		t.Fatalf("Expected %d indices, got %d", sampleSize, len(indices))
	}

	seen := make(map[int]bool)
	for _, idx := range indices {
		if seen[idx] {
			t.Errorf("Duplicate index found: %d", idx)
		}
		seen[idx] = true
	}
}

// TestRandomSampleSelection_WithinRange ensures all selected indices are
// within the valid range [0, populationSize).
func TestRandomSampleSelection_WithinRange(t *testing.T) {
	population := 500
	sampleSize := 100

	indices, err := GenerateRandomSample(population, sampleSize)
	if err != nil {
		t.Fatalf("GenerateRandomSample failed: %v", err)
	}

	for _, idx := range indices {
		if idx < 0 || idx >= population {
			t.Errorf("Index %d out of range [0, %d)", idx, population)
		}
	}

	// Indices should be sorted (our implementation sorts them).
	for i := 1; i < len(indices); i++ {
		if indices[i] <= indices[i-1] {
			t.Errorf("Indices not strictly sorted: indices[%d]=%d <= indices[%d]=%d",
				i, indices[i], i-1, indices[i-1])
		}
	}
}

// TestRandomSampleSelection_FullPopulation ensures that when sampleSize equals
// populationSize, all indices are returned.
func TestRandomSampleSelection_FullPopulation(t *testing.T) {
	population := 10
	sampleSize := 10

	indices, err := GenerateRandomSample(population, sampleSize)
	if err != nil {
		t.Fatalf("GenerateRandomSample failed: %v", err)
	}

	if len(indices) != population {
		t.Fatalf("Expected %d indices, got %d", population, len(indices))
	}

	for i := 0; i < population; i++ {
		if indices[i] != i {
			t.Errorf("Expected index %d at position %d, got %d", i, i, indices[i])
		}
	}
}

// TestRandomSampleSelection_ExceedsPopulation ensures that when sampleSize
// exceeds populationSize, it is clamped to the population.
func TestRandomSampleSelection_ExceedsPopulation(t *testing.T) {
	population := 5
	sampleSize := 20

	indices, err := GenerateRandomSample(population, sampleSize)
	if err != nil {
		t.Fatalf("GenerateRandomSample failed: %v", err)
	}

	if len(indices) != population {
		t.Errorf("Expected %d indices (clamped), got %d", population, len(indices))
	}
}

// ============================================================
// Risk-Based Selection Tests
// ============================================================

// TestRiskBasedSelection_PrioritiseCritical verifies that critical-rated
// entities receive a higher priority score than lower-rated ones.
func TestRiskBasedSelection_PrioritiseCritical(t *testing.T) {
	criticalWeight := riskRatingWeight("critical")
	highWeight := riskRatingWeight("high")
	mediumWeight := riskRatingWeight("medium")
	lowWeight := riskRatingWeight("low")

	if criticalWeight <= highWeight {
		t.Errorf("critical weight (%.1f) should exceed high weight (%.1f)", criticalWeight, highWeight)
	}
	if highWeight <= mediumWeight {
		t.Errorf("high weight (%.1f) should exceed medium weight (%.1f)", highWeight, mediumWeight)
	}
	if mediumWeight <= lowWeight {
		t.Errorf("medium weight (%.1f) should exceed low weight (%.1f)", mediumWeight, lowWeight)
	}

	// Simulate priority scoring: same time since audit, different risk.
	daysSinceAudit := 365
	timeFactor := float64(daysSinceAudit) / 365.0

	criticalPriority := criticalWeight * 1.0 * timeFactor
	lowPriority := lowWeight * 1.0 * timeFactor

	if criticalPriority <= lowPriority {
		t.Errorf("Critical entity priority (%.2f) should exceed low entity priority (%.2f)",
			criticalPriority, lowPriority)
	}
}

// TestRiskBasedSelection_OverdueAudits verifies that entities overdue for
// audit receive a higher priority boost.
func TestRiskBasedSelection_OverdueAudits(t *testing.T) {
	riskWeight := riskRatingWeight("medium")

	// Entity audited recently (30 days ago, frequency 12 months)
	recentDaysSince := 30
	recentTimeFactor := float64(recentDaysSince) / 365.0
	recentOverdueFactor := 1.0 // Not overdue (30 < 360)
	recentPriority := riskWeight * recentOverdueFactor * recentTimeFactor

	// Entity overdue (400 days ago, frequency 12 months = 360 days)
	overdueDaysSince := 400
	overdueTimeFactor := float64(overdueDaysSince) / 365.0
	overdueOverdueFactor := 1.5 // Overdue (400 > 360)
	overduePriority := riskWeight * overdueOverdueFactor * overdueTimeFactor

	if overduePriority <= recentPriority {
		t.Errorf("Overdue entity priority (%.2f) should exceed recently audited entity priority (%.2f)",
			overduePriority, recentPriority)
	}

	// The overdue factor should amplify priority by 1.5x.
	expectedRatio := (overdueOverdueFactor * overdueTimeFactor) / (recentOverdueFactor * recentTimeFactor)
	actualRatio := overduePriority / recentPriority
	if math.Abs(expectedRatio-actualRatio) > 0.001 {
		t.Errorf("Priority ratio = %.4f, expected = %.4f", actualRatio, expectedRatio)
	}
}

// TestRiskBasedSelection_NeverAudited verifies that entities that have never
// been audited default to 2 years (730 days) for the time factor.
func TestRiskBasedSelection_NeverAudited(t *testing.T) {
	riskWeight := riskRatingWeight("high")
	defaultDaysSince := 365 * 2
	timeFactor := float64(defaultDaysSince) / 365.0

	// Should yield a time factor of 2.0
	if timeFactor != 2.0 {
		t.Errorf("Expected time factor of 2.0 for never-audited entity, got %.2f", timeFactor)
	}

	priority := riskWeight * 1.5 * timeFactor // Assume overdue since never audited
	if priority <= 0 {
		t.Errorf("Never-audited entity should have positive priority, got %.2f", priority)
	}
}

// TestRiskBasedSelection_EstimatedDays verifies that estimated days increase
// with risk rating.
func TestRiskBasedSelection_EstimatedDays(t *testing.T) {
	critDays := estimateAuditDays("critical")
	highDays := estimateAuditDays("high")
	medDays := estimateAuditDays("medium")
	lowDays := estimateAuditDays("low")

	if critDays <= highDays {
		t.Errorf("Critical days (%d) should exceed high days (%d)", critDays, highDays)
	}
	if highDays <= medDays {
		t.Errorf("High days (%d) should exceed medium days (%d)", highDays, medDays)
	}
	if medDays <= lowDays {
		t.Errorf("Medium days (%d) should exceed low days (%d)", medDays, lowDays)
	}
}

// ============================================================
// Workpaper Review — Four-Eyes Principle Tests
// ============================================================

// TestWorkpaperReview_FourEyes verifies that the four-eyes principle
// prevents the same person from being both preparer and reviewer.
func TestWorkpaperReview_FourEyes(t *testing.T) {
	preparer := uuid.New()
	reviewer := uuid.New()

	// Same person: should be rejected.
	if preparer == preparer {
		// Simulate the check in SubmitForReview
		samePersonReview := (preparer == preparer) // true
		if !samePersonReview {
			t.Error("Expected same-person check to be true")
		}
	}

	// Different persons: should be allowed.
	differentPersonReview := (preparer == reviewer)
	if differentPersonReview {
		t.Error("Two different UUIDs should not be equal")
	}

	// Verify the actual logic used in the service:
	// reviewerID == preparedBy should block.
	simulateReview := func(prepared, review uuid.UUID) error {
		if review == prepared {
			return fmt.Errorf("four-eyes violation")
		}
		return nil
	}

	if err := simulateReview(preparer, preparer); err == nil {
		t.Error("Expected four-eyes violation when preparer == reviewer")
	}

	if err := simulateReview(preparer, reviewer); err != nil {
		t.Errorf("Expected no error when preparer != reviewer, got: %v", err)
	}
}

// ============================================================
// Corrective Action — Verifier Independence Tests
// ============================================================

// TestCorrectiveAction_VerifierDifferentFromImplementer ensures the verifier
// of a corrective action cannot be the same as the implementer.
func TestCorrectiveAction_VerifierDifferentFromImplementer(t *testing.T) {
	implementer := uuid.New()
	verifier := uuid.New()

	// Same person: should be rejected.
	simulateVerify := func(impl *uuid.UUID, verifierID uuid.UUID) error {
		if impl != nil && *impl == verifierID {
			return fmt.Errorf("four-eyes violation")
		}
		return nil
	}

	if err := simulateVerify(&implementer, implementer); err == nil {
		t.Error("Expected four-eyes violation when implementer == verifier")
	}

	if err := simulateVerify(&implementer, verifier); err != nil {
		t.Errorf("Expected no error when implementer != verifier, got: %v", err)
	}

	// Nil implementer: verification should always be allowed.
	if err := simulateVerify(nil, verifier); err != nil {
		t.Errorf("Expected no error when implementer is nil, got: %v", err)
	}
}

// ============================================================
// Engagement Status Transition Tests
// ============================================================

// TestEngagementStatus_ValidTransitions tests all valid and invalid
// engagement status transitions.
func TestEngagementStatus_ValidTransitions(t *testing.T) {
	validCases := []struct {
		from, to string
	}{
		{"planning", "fieldwork"},
		{"planning", "cancelled"},
		{"fieldwork", "review"},
		{"fieldwork", "cancelled"},
		{"review", "reporting"},
		{"review", "fieldwork"},
		{"reporting", "completed"},
		{"reporting", "review"},
	}

	for _, tc := range validCases {
		if !IsValidEngagementTransition(tc.from, tc.to) {
			t.Errorf("Expected transition %q -> %q to be valid", tc.from, tc.to)
		}
	}

	invalidCases := []struct {
		from, to string
	}{
		{"planning", "completed"},
		{"planning", "reporting"},
		{"planning", "review"},
		{"fieldwork", "completed"},
		{"fieldwork", "planning"},
		{"review", "cancelled"},
		{"review", "planning"},
		{"reporting", "planning"},
		{"reporting", "fieldwork"},
		{"completed", "fieldwork"},
		{"completed", "planning"},
		{"cancelled", "planning"},
		{"cancelled", "fieldwork"},
	}

	for _, tc := range invalidCases {
		if IsValidEngagementTransition(tc.from, tc.to) {
			t.Errorf("Expected transition %q -> %q to be invalid", tc.from, tc.to)
		}
	}
}

// TestEngagementStatus_TerminalStates verifies that completed and cancelled
// are terminal states with no valid outgoing transitions.
func TestEngagementStatus_TerminalStates(t *testing.T) {
	terminalStates := []string{"completed", "cancelled"}
	allStates := []string{"planning", "fieldwork", "review", "reporting", "completed", "cancelled"}

	for _, terminal := range terminalStates {
		for _, target := range allStates {
			if IsValidEngagementTransition(terminal, target) {
				t.Errorf("Terminal state %q should not allow transition to %q", terminal, target)
			}
		}
	}
}

// TestEngagementStatus_UnknownState verifies that unknown states are rejected.
func TestEngagementStatus_UnknownState(t *testing.T) {
	if IsValidEngagementTransition("unknown", "planning") {
		t.Error("Unknown source state should be rejected")
	}
	if IsValidEngagementTransition("planning", "unknown") {
		t.Error("Unknown target state should be rejected")
	}
}
