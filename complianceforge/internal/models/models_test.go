package models_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/complianceforge/platform/internal/models"
)

// ============================================================
// USER MODEL TESTS
// ============================================================

func TestUserFullName(t *testing.T) {
	user := &models.User{FirstName: "John", LastName: "Smith"}
	expected := "John Smith"
	if user.FullName() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, user.FullName())
	}
}

func TestUserIsLocked_NotLocked(t *testing.T) {
	user := &models.User{LockedUntil: nil}
	if user.IsLocked() {
		t.Error("User with nil LockedUntil should not be locked")
	}
}

func TestUserIsLocked_ExpiredLock(t *testing.T) {
	past := time.Now().Add(-1 * time.Hour)
	user := &models.User{LockedUntil: &past}
	if user.IsLocked() {
		t.Error("User with expired lock should not be locked")
	}
}

func TestUserIsLocked_ActiveLock(t *testing.T) {
	future := time.Now().Add(1 * time.Hour)
	user := &models.User{LockedUntil: &future}
	if !user.IsLocked() {
		t.Error("User with future lock should be locked")
	}
}

func TestUserPasswordNotExposed(t *testing.T) {
	user := &models.User{
		BaseModel:    models.BaseModel{ID: uuid.New()},
		Email:        "test@example.com",
		PasswordHash: "secret-hash-value",
		FirstName:    "Test",
		LastName:     "User",
	}

	jsonBytes, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	jsonStr := string(jsonBytes)
	if contains(jsonStr, "secret-hash-value") {
		t.Error("Password hash should not appear in JSON output")
	}
	if contains(jsonStr, "password_hash") {
		t.Error("password_hash field should not appear in JSON output")
	}
}

// ============================================================
// JSONB TYPE TESTS
// ============================================================

func TestJSONBMarshal(t *testing.T) {
	j := models.JSONB(`{"key": "value"}`)
	bytes, err := json.Marshal(j)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	if string(bytes) != `{"key": "value"}` {
		t.Errorf("Unexpected output: %s", string(bytes))
	}
}

func TestJSONBMarshalEmpty(t *testing.T) {
	j := models.JSONB(nil)
	bytes, err := json.Marshal(j)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	if string(bytes) != `{}` {
		t.Errorf("Expected '{}', got '%s'", string(bytes))
	}
}

// ============================================================
// ENUM TESTS
// ============================================================

func TestRiskLevelValues(t *testing.T) {
	levels := []models.RiskLevel{
		models.RiskLevelCritical,
		models.RiskLevelHigh,
		models.RiskLevelMedium,
		models.RiskLevelLow,
		models.RiskLevelVeryLow,
	}

	for _, level := range levels {
		if level == "" {
			t.Error("Risk level should not be empty")
		}
	}

	if len(levels) != 5 {
		t.Errorf("Expected 5 risk levels, got %d", len(levels))
	}
}

func TestControlStatusValues(t *testing.T) {
	statuses := []models.ControlStatus{
		models.ControlStatusNotApplicable,
		models.ControlStatusNotImplemented,
		models.ControlStatusPlanned,
		models.ControlStatusPartial,
		models.ControlStatusImplemented,
		models.ControlStatusEffective,
	}

	if len(statuses) != 6 {
		t.Errorf("Expected 6 control statuses, got %d", len(statuses))
	}
}

func TestPolicyStatusValues(t *testing.T) {
	statuses := []models.PolicyStatus{
		models.PolicyStatusDraft,
		models.PolicyStatusUnderReview,
		models.PolicyStatusPendingApproval,
		models.PolicyStatusApproved,
		models.PolicyStatusPublished,
		models.PolicyStatusArchived,
		models.PolicyStatusRetired,
	}

	if len(statuses) != 7 {
		t.Errorf("Expected 7 policy statuses, got %d", len(statuses))
	}
}

func TestFrameworkCodeValues(t *testing.T) {
	codes := []models.FrameworkCode{
		models.FrameworkISO27001,
		models.FrameworkUKGDPR,
		models.FrameworkNCSCCAF,
		models.FrameworkCyberEssentials,
		models.FrameworkNIST80053,
		models.FrameworkNISTCSF2,
		models.FrameworkPCIDSS,
		models.FrameworkITIL4,
		models.FrameworkCOBIT2019,
	}

	if len(codes) != 9 {
		t.Errorf("Expected 9 framework codes, got %d", len(codes))
	}
}

// ============================================================
// REQUEST VALIDATION TESTS
// ============================================================

func TestCreateRiskRequestMarshal(t *testing.T) {
	req := models.CreateRiskRequest{
		Title:              "Data breach via unpatched vulnerability",
		RiskCategoryID:     uuid.New(),
		RiskSource:         "external",
		OwnerUserID:        uuid.New(),
		InherentLikelihood: 4,
		InherentImpact:     5,
		ResidualLikelihood: 2,
		ResidualImpact:     3,
		FinancialImpactEUR: 500000,
		RiskVelocity:       "fast",
		ReviewFrequency:    "quarterly",
		Tags:               []string{"cyber", "vulnerability"},
	}

	bytes, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded models.CreateRiskRequest
	if err := json.Unmarshal(bytes, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.Title != req.Title {
		t.Errorf("Title mismatch: %s != %s", decoded.Title, req.Title)
	}
	if decoded.InherentLikelihood != 4 {
		t.Errorf("Expected likelihood 4, got %d", decoded.InherentLikelihood)
	}
}

func TestCreatePolicyRequestMarshal(t *testing.T) {
	req := models.CreatePolicyRequest{
		Title:                 "Information Security Policy",
		CategoryID:            uuid.New(),
		Classification:        "internal",
		ContentHTML:           "<h1>Policy</h1><p>Content here</p>",
		Summary:               "Defines the organisation's approach to information security.",
		OwnerUserID:           uuid.New(),
		ApproverUserID:        uuid.New(),
		ReviewFrequencyMonths: 12,
		IsMandatory:           true,
		RequiresAttestation:   true,
		Tags:                  []string{"security", "iso27001"},
	}

	bytes, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	if len(bytes) == 0 {
		t.Error("Marshalled bytes should not be empty")
	}
}

// ============================================================
// COMPLIANCE SCORE TESTS
// ============================================================

func TestComplianceScoreSummary(t *testing.T) {
	summary := models.ComplianceScoreSummary{
		FrameworkCode:   "ISO27001",
		FrameworkName:   "ISO 27001",
		TotalControls:   93,
		Implemented:     70,
		PartiallyImpl:   10,
		NotImplemented:  8,
		NotApplicable:   5,
		ComplianceScore: 79.55,
		MaturityAvg:     2.8,
	}

	// Verify applicable controls calculation
	applicable := summary.TotalControls - summary.NotApplicable
	expectedScore := float64(summary.Implemented) / float64(applicable) * 100

	if applicable != 88 {
		t.Errorf("Expected 88 applicable controls, got %d", applicable)
	}

	// Allow small floating point difference
	if diff := expectedScore - summary.ComplianceScore; diff > 0.1 || diff < -0.1 {
		t.Errorf("Score calculation mismatch: expected ~%.2f, got %.2f", expectedScore, summary.ComplianceScore)
	}
}

// ============================================================
// HELPERS
// ============================================================

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
