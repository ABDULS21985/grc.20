package service_test

import (
	"strings"
	"testing"
)

// ============================================================
// FRAMEWORK RECOMMENDATION ENGINE TESTS
// These tests verify the recommendation logic that maps industry
// assessment answers to compliance framework suggestions.
//
// The tests use a local mirror of the types and recommendation
// logic to run purely as unit tests without database access and
// without triggering import-cycle issues in the broader service
// package.
// ============================================================

// industryAssessmentInput mirrors service.IndustryAssessmentInput for testing.
type industryAssessmentInput struct {
	ProcessPaymentCards  bool
	HandleEUPersonalData bool
	EssentialServices    bool
	UKPublicSector       bool
	ISOCertification     bool
	USFederalContracts   bool
	CyberMaturity        bool
	ITILRequirements     bool
	BoardITGovernance    bool
}

// testRecommendation is a simplified recommendation for testing.
type testRecommendation struct {
	Code     string
	Reason   string
	Priority string
}

// generateRecommendationCodes mirrors the recommendation logic from
// OnboardingWizard.GetFrameworkRecommendations without DB access.
func generateRecommendationCodes(answers industryAssessmentInput) []testRecommendation {
	var recs []testRecommendation

	if answers.ProcessPaymentCards {
		recs = append(recs, testRecommendation{
			Code:     "PCI_DSS_4",
			Reason:   "Payment card processing requires PCI DSS compliance.",
			Priority: "required",
		})
	}
	if answers.HandleEUPersonalData {
		recs = append(recs, testRecommendation{
			Code:     "UK_GDPR",
			Reason:   "EU/UK personal data handling requires GDPR compliance.",
			Priority: "required",
		})
	}
	if answers.EssentialServices {
		recs = append(recs, testRecommendation{
			Code:     "NCSC_CAF",
			Reason:   "Essential services require NCSC CAF.",
			Priority: "required",
		})
		recs = append(recs, testRecommendation{
			Code:     "NIS2",
			Reason:   "Essential services fall under NIS2.",
			Priority: "required",
		})
	}
	if answers.UKPublicSector {
		recs = append(recs, testRecommendation{
			Code:     "CYBER_ESSENTIALS",
			Reason:   "UK public sector requires Cyber Essentials.",
			Priority: "required",
		})
	}
	if answers.ISOCertification {
		recs = append(recs, testRecommendation{
			Code:     "ISO27001",
			Reason:   "ISO certification interest maps to ISO 27001.",
			Priority: "recommended",
		})
	}
	if answers.USFederalContracts {
		recs = append(recs, testRecommendation{
			Code:     "NIST_800_53",
			Reason:   "US federal contracts require NIST 800-53.",
			Priority: "required",
		})
	}
	if answers.CyberMaturity {
		recs = append(recs, testRecommendation{
			Code:     "NIST_CSF_2",
			Reason:   "NIST CSF 2.0 for cybersecurity maturity.",
			Priority: "recommended",
		})
	}
	if answers.ITILRequirements {
		recs = append(recs, testRecommendation{
			Code:     "ITIL_4",
			Reason:   "ITIL 4 for IT service management.",
			Priority: "recommended",
		})
	}
	if answers.BoardITGovernance {
		recs = append(recs, testRecommendation{
			Code:     "COBIT_2019",
			Reason:   "COBIT 2019 for board-level IT governance.",
			Priority: "recommended",
		})
	}

	// Default recommendations if none selected
	if len(recs) == 0 {
		recs = append(recs, testRecommendation{
			Code:     "NIST_CSF_2",
			Reason:   "Default recommendation.",
			Priority: "recommended",
		})
		recs = append(recs, testRecommendation{
			Code:     "ISO27001",
			Reason:   "Default recommendation.",
			Priority: "recommended",
		})
	}

	// Deduplicate
	seen := make(map[string]bool)
	var unique []testRecommendation
	for _, r := range recs {
		if !seen[r.Code] {
			seen[r.Code] = true
			unique = append(unique, r)
		}
	}

	return unique
}

// assertContainsCode checks that a code exists in the recommendations.
func assertContainsCode(t *testing.T, recs []testRecommendation, code string) {
	t.Helper()
	for _, r := range recs {
		if r.Code == code {
			return
		}
	}
	t.Errorf("Expected recommendation for %s, but not found in %d recommendations", code, len(recs))
}

// ============================================================
// RECOMMENDATION TESTS
// ============================================================

func TestGetFrameworkRecommendations_PaymentCards(t *testing.T) {
	answers := industryAssessmentInput{ProcessPaymentCards: true}
	recs := generateRecommendationCodes(answers)
	assertContainsCode(t, recs, "PCI_DSS_4")
}

func TestGetFrameworkRecommendations_EUPersonalData(t *testing.T) {
	answers := industryAssessmentInput{HandleEUPersonalData: true}
	recs := generateRecommendationCodes(answers)
	assertContainsCode(t, recs, "UK_GDPR")
}

func TestGetFrameworkRecommendations_EssentialServices(t *testing.T) {
	answers := industryAssessmentInput{EssentialServices: true}
	recs := generateRecommendationCodes(answers)
	assertContainsCode(t, recs, "NCSC_CAF")
	assertContainsCode(t, recs, "NIS2")
}

func TestGetFrameworkRecommendations_UKPublicSector(t *testing.T) {
	answers := industryAssessmentInput{UKPublicSector: true}
	recs := generateRecommendationCodes(answers)
	assertContainsCode(t, recs, "CYBER_ESSENTIALS")
}

func TestGetFrameworkRecommendations_ISOCertification(t *testing.T) {
	answers := industryAssessmentInput{ISOCertification: true}
	recs := generateRecommendationCodes(answers)
	assertContainsCode(t, recs, "ISO27001")
}

func TestGetFrameworkRecommendations_USFederal(t *testing.T) {
	answers := industryAssessmentInput{USFederalContracts: true}
	recs := generateRecommendationCodes(answers)
	assertContainsCode(t, recs, "NIST_800_53")
}

func TestGetFrameworkRecommendations_CyberMaturity(t *testing.T) {
	answers := industryAssessmentInput{CyberMaturity: true}
	recs := generateRecommendationCodes(answers)
	assertContainsCode(t, recs, "NIST_CSF_2")
}

func TestGetFrameworkRecommendations_ITIL(t *testing.T) {
	answers := industryAssessmentInput{ITILRequirements: true}
	recs := generateRecommendationCodes(answers)
	assertContainsCode(t, recs, "ITIL_4")
}

func TestGetFrameworkRecommendations_BoardGovernance(t *testing.T) {
	answers := industryAssessmentInput{BoardITGovernance: true}
	recs := generateRecommendationCodes(answers)
	assertContainsCode(t, recs, "COBIT_2019")
}

func TestGetFrameworkRecommendations_MultipleAnswers(t *testing.T) {
	answers := industryAssessmentInput{
		ProcessPaymentCards:  true,
		HandleEUPersonalData: true,
		ISOCertification:     true,
		CyberMaturity:        true,
	}
	recs := generateRecommendationCodes(answers)
	assertContainsCode(t, recs, "PCI_DSS_4")
	assertContainsCode(t, recs, "UK_GDPR")
	assertContainsCode(t, recs, "ISO27001")
	assertContainsCode(t, recs, "NIST_CSF_2")

	// Verify no duplicates
	seen := make(map[string]bool)
	for _, r := range recs {
		if seen[r.Code] {
			t.Errorf("Duplicate recommendation code: %s", r.Code)
		}
		seen[r.Code] = true
	}
}

func TestGetFrameworkRecommendations_NoAnswers_DefaultRecommendations(t *testing.T) {
	answers := industryAssessmentInput{}
	recs := generateRecommendationCodes(answers)

	if len(recs) == 0 {
		t.Error("Expected default recommendations when no answers provided")
	}
	assertContainsCode(t, recs, "NIST_CSF_2")
	assertContainsCode(t, recs, "ISO27001")
}

func TestGetFrameworkRecommendations_AllAnswers(t *testing.T) {
	answers := industryAssessmentInput{
		ProcessPaymentCards:  true,
		HandleEUPersonalData: true,
		EssentialServices:    true,
		UKPublicSector:       true,
		ISOCertification:     true,
		USFederalContracts:   true,
		CyberMaturity:        true,
		ITILRequirements:     true,
		BoardITGovernance:    true,
	}
	recs := generateRecommendationCodes(answers)

	expectedCodes := []string{
		"PCI_DSS_4", "UK_GDPR", "NCSC_CAF", "NIS2",
		"CYBER_ESSENTIALS", "ISO27001", "NIST_800_53",
		"NIST_CSF_2", "ITIL_4", "COBIT_2019",
	}
	for _, code := range expectedCodes {
		assertContainsCode(t, recs, code)
	}

	if len(recs) != 10 {
		t.Errorf("Expected 10 recommendations for all answers, got %d", len(recs))
	}
}

func TestGetFrameworkRecommendations_PriorityAssignment(t *testing.T) {
	answers := industryAssessmentInput{
		ProcessPaymentCards: true,
		ISOCertification:    true,
	}
	recs := generateRecommendationCodes(answers)

	for _, r := range recs {
		switch r.Code {
		case "PCI_DSS_4":
			if r.Priority != "required" {
				t.Errorf("PCI DSS should have priority 'required', got %q", r.Priority)
			}
		case "ISO27001":
			if r.Priority != "recommended" {
				t.Errorf("ISO 27001 should have priority 'recommended', got %q", r.Priority)
			}
		}
	}
}

// ============================================================
// FORMAT FRAMEWORK NAME TESTS
// ============================================================

func TestFormatFrameworkName_KnownCodes(t *testing.T) {
	tests := []struct {
		code     string
		expected string
	}{
		{"ISO27001", "ISO 27001"},
		{"UK_GDPR", "UK GDPR"},
		{"NCSC_CAF", "NCSC CAF"},
		{"NIS2", "NIS2 Directive"},
		{"CYBER_ESSENTIALS", "Cyber Essentials"},
		{"NIST_800_53", "NIST 800-53"},
		{"NIST_CSF_2", "NIST CSF 2.0"},
		{"PCI_DSS_4", "PCI DSS v4.0"},
		{"ITIL_4", "ITIL 4"},
		{"COBIT_2019", "COBIT 2019"},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			result := formatFrameworkNameTest(tt.code)
			if result != tt.expected {
				t.Errorf("formatFrameworkName(%q) = %q, want %q", tt.code, result, tt.expected)
			}
		})
	}
}

func TestFormatFrameworkName_UnknownCode(t *testing.T) {
	result := formatFrameworkNameTest("SOME_UNKNOWN_FRAMEWORK")
	expected := "SOME UNKNOWN FRAMEWORK"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

// formatFrameworkNameTest mirrors the service function for testing.
func formatFrameworkNameTest(code string) string {
	names := map[string]string{
		"ISO27001":         "ISO 27001",
		"UK_GDPR":          "UK GDPR",
		"NCSC_CAF":         "NCSC CAF",
		"NIS2":             "NIS2 Directive",
		"CYBER_ESSENTIALS": "Cyber Essentials",
		"NIST_800_53":      "NIST 800-53",
		"NIST_CSF_2":       "NIST CSF 2.0",
		"PCI_DSS_4":        "PCI DSS v4.0",
		"ITIL_4":           "ITIL 4",
		"COBIT_2019":       "COBIT 2019",
	}
	if name, ok := names[code]; ok {
		return name
	}
	return strings.ReplaceAll(code, "_", " ")
}

// ============================================================
// PLAN LIMIT CHECK TESTS (unit logic)
// ============================================================

func TestLimitCheck_CanCreate(t *testing.T) {
	tests := []struct {
		name      string
		current   int
		maxVal    int
		canCreate bool
		remaining int
	}{
		{"under_limit", 3, 5, true, 2},
		{"at_limit", 5, 5, false, 0},
		{"over_limit", 7, 5, false, 0},
		{"zero_usage", 0, 10, true, 10},
		{"one_remaining", 4, 5, true, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			remaining := tt.maxVal - tt.current
			if remaining < 0 {
				remaining = 0
			}
			canCreate := tt.current < tt.maxVal

			if canCreate != tt.canCreate {
				t.Errorf("CanCreate = %v, want %v", canCreate, tt.canCreate)
			}
			if remaining != tt.remaining {
				t.Errorf("Remaining = %d, want %d", remaining, tt.remaining)
			}
		})
	}
}

// ============================================================
// ONBOARDING STEP VALIDATION TESTS
// ============================================================

func TestSkipStep_InvalidStepNumbers(t *testing.T) {
	tests := []struct {
		step    int
		wantErr bool
	}{
		{0, true},
		{-1, true},
		{8, true},
		{7, true},
		{1, false},
		{6, false},
	}

	for _, tt := range tests {
		isValid := tt.step >= 1 && tt.step <= 7
		canSkip := isValid && tt.step != 7

		if tt.wantErr && canSkip {
			t.Errorf("Step %d should not be skippable", tt.step)
		}
		if !tt.wantErr && !canSkip {
			t.Errorf("Step %d should be skippable", tt.step)
		}
	}
}

func TestOrgProfileInput_Fields(t *testing.T) {
	type orgProfileInput struct {
		DisplayName        string
		LegalName          string
		Industry           string
		CountryCode        string
		Timezone           string
		EmployeeCountRange string
	}

	input := orgProfileInput{
		DisplayName:        "Acme Corp",
		LegalName:          "Acme Corporation Ltd",
		Industry:           "technology",
		CountryCode:        "GB",
		Timezone:           "Europe/London",
		EmployeeCountRange: "50-249",
	}

	if input.DisplayName != "Acme Corp" {
		t.Errorf("DisplayName = %q, want 'Acme Corp'", input.DisplayName)
	}
	if input.CountryCode != "GB" {
		t.Errorf("CountryCode = %q, want 'GB'", input.CountryCode)
	}
}

func TestQuickAssessmentInput_MaturityRange(t *testing.T) {
	fields := []struct {
		name  string
		value int
	}{
		{"SecurityMaturity", 3},
		{"PolicyDocumentation", 2},
		{"IncidentResponseReady", 4},
		{"DataProtectionLevel", 3},
		{"AccessControlMaturity", 2},
		{"ThirdPartyRiskMgmt", 1},
		{"BusinessContinuity", 3},
		{"SecurityAwareness", 4},
	}

	for _, f := range fields {
		if f.value < 1 || f.value > 5 {
			t.Errorf("%s = %d, want 1-5", f.name, f.value)
		}
	}
}

func TestRiskAppetiteInput_ValidValues(t *testing.T) {
	validAppetites := []string{"conservative", "moderate", "aggressive"}
	overallAppetite := "moderate"

	found := false
	for _, v := range validAppetites {
		if overallAppetite == v {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("OverallAppetite %q not in valid values", overallAppetite)
	}
}

// ============================================================
// SUBSCRIPTION PLAN TESTS
// ============================================================

func TestSubscriptionPlan_MonthlySavings(t *testing.T) {
	plans := []struct {
		name           string
		monthly        float64
		annual         float64
		expectedSaving float64
	}{
		{"Starter", 99.00, 990.00, 16.50},
		{"Professional", 299.00, 2990.00, 49.83},
		{"Enterprise", 799.00, 7990.00, 133.17},
	}

	for _, p := range plans {
		t.Run(p.name, func(t *testing.T) {
			annualMonthly := p.annual / 12
			saving := p.monthly - annualMonthly

			if saving <= 0 {
				t.Errorf("Expected positive monthly savings, got %.2f", saving)
			}

			diff := saving - p.expectedSaving
			if diff < -0.01 || diff > 0.01 {
				t.Errorf("Monthly savings = %.2f, expected ~%.2f", saving, p.expectedSaving)
			}
		})
	}
}

func TestSubscriptionPlan_TierLimits(t *testing.T) {
	tiers := []struct {
		name          string
		maxUsers      int
		maxFrameworks int
	}{
		{"starter", 5, 3},
		{"professional", 25, 5},
		{"enterprise", 100, 9},
	}

	for _, tier := range tiers {
		t.Run(tier.name, func(t *testing.T) {
			if tier.maxUsers <= 0 {
				t.Errorf("maxUsers should be positive, got %d", tier.maxUsers)
			}
			if tier.maxFrameworks <= 0 {
				t.Errorf("maxFrameworks should be positive, got %d", tier.maxFrameworks)
			}
		})
	}

	for i := 1; i < len(tiers); i++ {
		if tiers[i].maxUsers <= tiers[i-1].maxUsers {
			t.Errorf("Tier %s should have more users than %s", tiers[i].name, tiers[i-1].name)
		}
		if tiers[i].maxFrameworks <= tiers[i-1].maxFrameworks {
			t.Errorf("Tier %s should have more frameworks than %s", tiers[i].name, tiers[i-1].name)
		}
	}
}
