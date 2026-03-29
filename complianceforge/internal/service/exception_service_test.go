package service

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

// ============================================================
// COMPLIANCE EXCEPTION MODEL TESTS
// ============================================================

func TestComplianceExceptionSerialization(t *testing.T) {
	now := time.Now()
	approverID := uuid.New()
	approvedAt := now
	expiryDate := now.AddDate(1, 0, 0)
	nextReview := now.AddDate(0, 6, 0)
	policyID := uuid.New()
	riskAssessmentID := uuid.New()

	exc := ComplianceException{
		ID:                            uuid.New(),
		OrganizationID:                uuid.New(),
		ExceptionRef:                  "EXC-2026-0001",
		Title:                         "Legacy System Authentication Exception",
		Description:                   "Legacy ERP system cannot support MFA",
		ExceptionType:                 "temporary",
		Status:                        "approved",
		Priority:                      "high",
		ScopeType:                     "control_implementation",
		ControlImplementationIDs:      []uuid.UUID{uuid.New(), uuid.New()},
		FrameworkControlCodes:         []string{"ISO27001-A.9.4.2", "NIST-IA-2"},
		PolicyID:                      &policyID,
		ScopeDescription:              "Legacy ERP authentication module",
		RiskJustification:             "System replacement planned for Q3 2026",
		ResidualRiskDescription:       "Increased risk of unauthorized access",
		ResidualRiskLevel:             "high",
		RiskAssessmentID:              &riskAssessmentID,
		RiskAcceptedBy:                &approverID,
		RiskAcceptedAt:                &approvedAt,
		HasCompensatingControls:       true,
		CompensatingControlsDescription: "Enhanced logging, IP restrictions, VPN-only access",
		CompensatingControlIDs:        []uuid.UUID{uuid.New()},
		CompensatingEffectiveness:     "mostly_effective",
		RequestedBy:                   uuid.New(),
		RequestedAt:                   now,
		ApprovedBy:                    &approverID,
		ApprovedAt:                    &approvedAt,
		ApprovalComments:              "Approved with compensating controls in place",
		EffectiveDate:                 now,
		ExpiryDate:                    &expiryDate,
		ReviewFrequencyMonths:         6,
		NextReviewDate:                &nextReview,
		RenewalCount:                  0,
		Conditions:                    "Must maintain VPN access logs",
		BusinessImpactIfImplemented:   "ERP downtime of 4+ weeks during migration",
		RegulatoryNotificationRequired: false,
		Tags:                          []string{"legacy", "authentication", "erp"},
		Metadata:                      json.RawMessage(`{"system":"legacy_erp","version":"4.2"}`),
		CreatedAt:                     now,
		UpdatedAt:                     now,
	}

	data, err := json.Marshal(exc)
	if err != nil {
		t.Fatalf("Failed to marshal ComplianceException: %v", err)
	}

	var decoded ComplianceException
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal ComplianceException: %v", err)
	}

	if decoded.ExceptionRef != "EXC-2026-0001" {
		t.Errorf("Expected exception_ref 'EXC-2026-0001', got '%s'", decoded.ExceptionRef)
	}
	if decoded.Title != "Legacy System Authentication Exception" {
		t.Errorf("Expected title mismatch, got '%s'", decoded.Title)
	}
	if decoded.Status != "approved" {
		t.Errorf("Expected status 'approved', got '%s'", decoded.Status)
	}
	if decoded.ExceptionType != "temporary" {
		t.Errorf("Expected type 'temporary', got '%s'", decoded.ExceptionType)
	}
	if decoded.Priority != "high" {
		t.Errorf("Expected priority 'high', got '%s'", decoded.Priority)
	}
	if decoded.ResidualRiskLevel != "high" {
		t.Errorf("Expected risk level 'high', got '%s'", decoded.ResidualRiskLevel)
	}
	if !decoded.HasCompensatingControls {
		t.Error("Expected has_compensating_controls to be true")
	}
	if decoded.CompensatingEffectiveness != "mostly_effective" {
		t.Errorf("Expected effectiveness 'mostly_effective', got '%s'", decoded.CompensatingEffectiveness)
	}
	if len(decoded.ControlImplementationIDs) != 2 {
		t.Errorf("Expected 2 control_implementation_ids, got %d", len(decoded.ControlImplementationIDs))
	}
	if len(decoded.FrameworkControlCodes) != 2 {
		t.Errorf("Expected 2 framework_control_codes, got %d", len(decoded.FrameworkControlCodes))
	}
	if decoded.ReviewFrequencyMonths != 6 {
		t.Errorf("Expected review_frequency_months 6, got %d", decoded.ReviewFrequencyMonths)
	}
	if len(decoded.Tags) != 3 {
		t.Errorf("Expected 3 tags, got %d", len(decoded.Tags))
	}
	if decoded.RenewalCount != 0 {
		t.Errorf("Expected renewal_count 0, got %d", decoded.RenewalCount)
	}
}

func TestExceptionReviewSerialization(t *testing.T) {
	now := time.Now()
	nextReview := now.AddDate(0, 6, 0)
	compensatingEffective := true

	review := ExceptionReview{
		ID:                    uuid.New(),
		OrganizationID:        uuid.New(),
		ExceptionID:           uuid.New(),
		ReviewType:            "periodic",
		ReviewerID:            uuid.New(),
		ReviewDate:            now,
		Outcome:               "continue",
		RiskLevelAtReview:     "medium",
		CompensatingEffective: &compensatingEffective,
		Findings:              "Compensating controls operating effectively",
		Recommendations:       "Continue monitoring; plan migration for Q3",
		NextReviewDate:        &nextReview,
		Attachments:           []string{"review_evidence_001.pdf"},
		Metadata:              json.RawMessage(`{"reviewed_by_role":"risk_manager"}`),
		CreatedAt:             now,
	}

	data, err := json.Marshal(review)
	if err != nil {
		t.Fatalf("Failed to marshal ExceptionReview: %v", err)
	}

	var decoded ExceptionReview
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal ExceptionReview: %v", err)
	}

	if decoded.ReviewType != "periodic" {
		t.Errorf("Expected review_type 'periodic', got '%s'", decoded.ReviewType)
	}
	if decoded.Outcome != "continue" {
		t.Errorf("Expected outcome 'continue', got '%s'", decoded.Outcome)
	}
	if decoded.RiskLevelAtReview != "medium" {
		t.Errorf("Expected risk_level_at_review 'medium', got '%s'", decoded.RiskLevelAtReview)
	}
	if decoded.CompensatingEffective == nil || !*decoded.CompensatingEffective {
		t.Error("Expected compensating_effective to be true")
	}
}

func TestExceptionAuditEntrySerialization(t *testing.T) {
	now := time.Now()

	entry := ExceptionAuditEntry{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		ExceptionID:    uuid.New(),
		Action:         "approved",
		ActorID:        uuid.New(),
		ActorEmail:     "admin@example.com",
		OldStatus:      "pending_approval",
		NewStatus:      "approved",
		Details:        "Approved with compensating controls in place",
		IPAddress:      "192.168.1.100",
		UserAgent:      "Mozilla/5.0",
		Metadata:       json.RawMessage(`{}`),
		CreatedAt:      now,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("Failed to marshal ExceptionAuditEntry: %v", err)
	}

	var decoded ExceptionAuditEntry
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal ExceptionAuditEntry: %v", err)
	}

	if decoded.Action != "approved" {
		t.Errorf("Expected action 'approved', got '%s'", decoded.Action)
	}
	if decoded.OldStatus != "pending_approval" {
		t.Errorf("Expected old_status 'pending_approval', got '%s'", decoded.OldStatus)
	}
	if decoded.NewStatus != "approved" {
		t.Errorf("Expected new_status 'approved', got '%s'", decoded.NewStatus)
	}
	if decoded.ActorEmail != "admin@example.com" {
		t.Errorf("Expected actor_email 'admin@example.com', got '%s'", decoded.ActorEmail)
	}
}

// ============================================================
// COMPLIANCE IMPACT CALCULATION TESTS
// ============================================================

func TestComplianceImpactHighRisk(t *testing.T) {
	impact := calculateImpactForTest(
		"critical",           // residualRiskLevel
		5,                    // controlCount
		true,                 // hasCompensating
		"partially_effective", // effectiveness
		"temporary",          // exceptionType
		1,                    // renewalCount
		nil,                  // expiryDate
	)

	// Critical risk with 5 controls = 5 * 2.5 = 12.5
	if impact.ComplianceScoreImpact != 12.5 {
		t.Errorf("Expected compliance_score_impact 12.5, got %v", impact.ComplianceScoreImpact)
	}

	if impact.RiskExposureIncrease != "significant" {
		t.Errorf("Expected risk_exposure 'significant', got '%s'", impact.RiskExposureIncrease)
	}

	if impact.CompensatingCoverage != 50.0 {
		t.Errorf("Expected compensating_coverage 50.0, got %v", impact.CompensatingCoverage)
	}

	// 50% coverage is moderate
	if impact.NetRiskDelta != "moderate" {
		t.Errorf("Expected net_risk_delta 'moderate', got '%s'", impact.NetRiskDelta)
	}
}

func TestComplianceImpactLowRiskFullCompensation(t *testing.T) {
	impact := calculateImpactForTest(
		"low",             // residualRiskLevel
		2,                 // controlCount
		true,              // hasCompensating
		"fully_effective", // effectiveness
		"permanent",       // exceptionType
		0,                 // renewalCount
		nil,               // expiryDate
	)

	// Low risk with 2 controls = 2 * 0.3 = 0.6
	if impact.ComplianceScoreImpact != 0.6 {
		t.Errorf("Expected compliance_score_impact 0.6, got %v", impact.ComplianceScoreImpact)
	}

	if impact.RiskExposureIncrease != "minimal" {
		t.Errorf("Expected risk_exposure 'minimal', got '%s'", impact.RiskExposureIncrease)
	}

	if impact.CompensatingCoverage != 95.0 {
		t.Errorf("Expected compensating_coverage 95.0, got %v", impact.CompensatingCoverage)
	}

	// 95% coverage = low risk delta
	if impact.NetRiskDelta != "low" {
		t.Errorf("Expected net_risk_delta 'low', got '%s'", impact.NetRiskDelta)
	}
}

func TestComplianceImpactNoCompensating(t *testing.T) {
	impact := calculateImpactForTest(
		"high",         // residualRiskLevel
		3,              // controlCount
		false,          // hasCompensating
		"not_assessed", // effectiveness
		"temporary",    // exceptionType
		0,              // renewalCount
		nil,            // expiryDate
	)

	// High risk with 3 controls = 3 * 1.5 = 4.5
	if impact.ComplianceScoreImpact != 4.5 {
		t.Errorf("Expected compliance_score_impact 4.5, got %v", impact.ComplianceScoreImpact)
	}

	if impact.CompensatingCoverage != 0.0 {
		t.Errorf("Expected compensating_coverage 0.0, got %v", impact.CompensatingCoverage)
	}

	if impact.NetRiskDelta != "high" {
		t.Errorf("Expected net_risk_delta 'high', got '%s'", impact.NetRiskDelta)
	}

	// Should recommend compensating controls
	found := false
	for _, rec := range impact.Recommendations {
		if rec == "Consider implementing compensating controls to reduce risk exposure" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected recommendation to implement compensating controls")
	}
}

func TestComplianceImpactMaxRenewals(t *testing.T) {
	impact := calculateImpactForTest(
		"medium",          // residualRiskLevel
		1,                 // controlCount
		true,              // hasCompensating
		"mostly_effective", // effectiveness
		"temporary",       // exceptionType
		2,                 // renewalCount (max)
		nil,               // expiryDate
	)

	// Should recommend permanent remediation
	found := false
	for _, rec := range impact.Recommendations {
		if rec == "Maximum renewals reached; plan for permanent remediation" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected recommendation for permanent remediation when max renewals reached")
	}
}

func TestComplianceImpactExpiringSoon(t *testing.T) {
	expiry := time.Now().AddDate(0, 0, 15) // 15 days from now
	impact := calculateImpactForTest(
		"medium",          // residualRiskLevel
		2,                 // controlCount
		true,              // hasCompensating
		"mostly_effective", // effectiveness
		"temporary",       // exceptionType
		0,                 // renewalCount
		&expiry,           // expiryDate
	)

	// Should recommend renewal planning
	found := false
	for _, rec := range impact.Recommendations {
		if rec == "Exception expiring soon; initiate renewal or remediation planning" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected recommendation for expiring exception")
	}
}

func TestComplianceImpactMediumRisk(t *testing.T) {
	impact := calculateImpactForTest(
		"medium",           // residualRiskLevel
		4,                  // controlCount
		true,               // hasCompensating
		"minimally_effective", // effectiveness
		"conditional",      // exceptionType
		0,                  // renewalCount
		nil,                // expiryDate
	)

	// Medium risk with 4 controls = 4 * 0.8 = 3.2
	if impact.ComplianceScoreImpact != 3.2 {
		t.Errorf("Expected compliance_score_impact 3.2, got %v", impact.ComplianceScoreImpact)
	}

	if impact.RiskExposureIncrease != "limited" {
		t.Errorf("Expected risk_exposure 'limited', got '%s'", impact.RiskExposureIncrease)
	}

	// Minimally effective = 25% coverage
	if impact.CompensatingCoverage != 25.0 {
		t.Errorf("Expected compensating_coverage 25.0, got %v", impact.CompensatingCoverage)
	}
}

func TestComplianceImpactNotEffective(t *testing.T) {
	impact := calculateImpactForTest(
		"high",          // residualRiskLevel
		1,               // controlCount
		true,            // hasCompensating
		"not_effective", // effectiveness
		"temporary",     // exceptionType
		0,               // renewalCount
		nil,             // expiryDate
	)

	if impact.CompensatingCoverage != 5.0 {
		t.Errorf("Expected compensating_coverage 5.0, got %v", impact.CompensatingCoverage)
	}

	// Should recommend improving compensating controls
	found := false
	for _, rec := range impact.Recommendations {
		if rec == "Assess or improve compensating control effectiveness" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected recommendation to improve compensating controls")
	}
}

// ============================================================
// EXPIRY HANDLING TESTS
// ============================================================

func TestExceptionExpiryValidation(t *testing.T) {
	tests := []struct {
		name          string
		effectiveDate string
		expiryDate    string
		expectError   bool
	}{
		{
			name:          "valid dates",
			effectiveDate: "2026-01-01",
			expiryDate:    "2027-01-01",
			expectError:   false,
		},
		{
			name:          "expiry before effective",
			effectiveDate: "2026-06-01",
			expiryDate:    "2026-01-01",
			expectError:   true,
		},
		{
			name:          "same date",
			effectiveDate: "2026-06-01",
			expiryDate:    "2026-06-01",
			expectError:   true,
		},
		{
			name:          "no expiry date",
			effectiveDate: "2026-01-01",
			expiryDate:    "",
			expectError:   false,
		},
		{
			name:          "invalid effective format",
			effectiveDate: "01-01-2026",
			expiryDate:    "2027-01-01",
			expectError:   true,
		},
		{
			name:          "invalid expiry format",
			effectiveDate: "2026-01-01",
			expiryDate:    "invalid",
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateExceptionDates(tt.effectiveDate, tt.expiryDate)
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestRenewalCountValidation(t *testing.T) {
	tests := []struct {
		name         string
		excType      string
		renewalCount int
		expectError  bool
	}{
		{
			name:         "temporary first renewal",
			excType:      "temporary",
			renewalCount: 0,
			expectError:  false,
		},
		{
			name:         "temporary second renewal",
			excType:      "temporary",
			renewalCount: 1,
			expectError:  false,
		},
		{
			name:         "temporary max renewals reached",
			excType:      "temporary",
			renewalCount: 2,
			expectError:  true,
		},
		{
			name:         "permanent many renewals",
			excType:      "permanent",
			renewalCount: 5,
			expectError:  false,
		},
		{
			name:         "conditional renewal",
			excType:      "conditional",
			renewalCount: 3,
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRenewal(tt.excType, tt.renewalCount)
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestStatusTransitions(t *testing.T) {
	tests := []struct {
		name          string
		currentStatus string
		targetAction  string
		expectValid   bool
	}{
		{"draft to submit", "draft", "submit", true},
		{"pending to approve", "pending_risk_assessment", "approve", true},
		{"pending_approval to approve", "pending_approval", "approve", true},
		{"pending to reject", "pending_risk_assessment", "reject", true},
		{"approved to revoke", "approved", "revoke", true},
		{"approved to renew", "approved", "renew", true},
		{"expired to renew", "expired", "renew", true},
		{"draft to approve", "draft", "approve", false},
		{"approved to approve", "approved", "approve", false},
		{"rejected to approve", "rejected", "approve", false},
		{"draft to revoke", "draft", "revoke", false},
		{"rejected to revoke", "rejected", "revoke", false},
		{"revoked to renew", "revoked", "renew", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := isValidTransition(tt.currentStatus, tt.targetAction)
			if valid != tt.expectValid {
				t.Errorf("Expected transition %s -> %s to be %v, got %v",
					tt.currentStatus, tt.targetAction, tt.expectValid, valid)
			}
		})
	}
}

func TestCreateExceptionRequestValidation(t *testing.T) {
	tests := []struct {
		name        string
		req         CreateExceptionRequest
		expectError bool
	}{
		{
			name: "valid minimal request",
			req: CreateExceptionRequest{
				Title:             "Test Exception",
				Description:       "Test description",
				RiskJustification: "Test justification",
				EffectiveDate:     "2026-01-01",
			},
			expectError: false,
		},
		{
			name: "missing title",
			req: CreateExceptionRequest{
				Description:       "Test description",
				RiskJustification: "Test justification",
				EffectiveDate:     "2026-01-01",
			},
			expectError: true,
		},
		{
			name: "missing description",
			req: CreateExceptionRequest{
				Title:             "Test Exception",
				RiskJustification: "Test justification",
				EffectiveDate:     "2026-01-01",
			},
			expectError: true,
		},
		{
			name: "missing risk justification",
			req: CreateExceptionRequest{
				Title:         "Test Exception",
				Description:   "Test description",
				EffectiveDate: "2026-01-01",
			},
			expectError: true,
		},
		{
			name: "missing effective date",
			req: CreateExceptionRequest{
				Title:             "Test Exception",
				Description:       "Test description",
				RiskJustification: "Test justification",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCreateRequest(tt.req)
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestExceptionDashboardStructure(t *testing.T) {
	dash := ExceptionDashboard{
		ActiveCount:          5,
		DraftCount:           3,
		PendingApprovalCount: 2,
		RejectedCount:        1,
		ExpiredCount:         4,
		ByRiskLevel: map[string]int{
			"critical": 1,
			"high":     2,
			"medium":   3,
			"low":      1,
		},
		ByPriority: map[string]int{
			"critical": 0,
			"high":     3,
			"medium":   4,
			"low":      1,
		},
		ByType: map[string]int{
			"temporary":   4,
			"permanent":   2,
			"conditional": 1,
		},
		Expiring30Days:  2,
		Expiring60Days:  3,
		Expiring90Days:  5,
		OverdueReviews:  1,
		AverageAgeDays:  45.3,
		TopExceptedFrameworks: []FrameworkExCount{
			{FrameworkCode: "ISO27001", Count: 5},
			{FrameworkCode: "NIST_800_53", Count: 3},
		},
		RecentExceptions: []ComplianceException{},
	}

	data, err := json.Marshal(dash)
	if err != nil {
		t.Fatalf("Failed to marshal ExceptionDashboard: %v", err)
	}

	var decoded ExceptionDashboard
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal ExceptionDashboard: %v", err)
	}

	if decoded.ActiveCount != 5 {
		t.Errorf("Expected active_count 5, got %d", decoded.ActiveCount)
	}
	if decoded.Expiring30Days != 2 {
		t.Errorf("Expected expiring_30_days 2, got %d", decoded.Expiring30Days)
	}
	if decoded.OverdueReviews != 1 {
		t.Errorf("Expected overdue_reviews 1, got %d", decoded.OverdueReviews)
	}
	if decoded.AverageAgeDays != 45.3 {
		t.Errorf("Expected average_age_days 45.3, got %v", decoded.AverageAgeDays)
	}
	if len(decoded.TopExceptedFrameworks) != 2 {
		t.Errorf("Expected 2 top frameworks, got %d", len(decoded.TopExceptedFrameworks))
	}
	if decoded.ByRiskLevel["critical"] != 1 {
		t.Errorf("Expected 1 critical risk, got %d", decoded.ByRiskLevel["critical"])
	}
}

func TestComplianceImpactStructure(t *testing.T) {
	impact := ComplianceImpact{
		ExceptionID:           uuid.New(),
		ExceptionRef:          "EXC-2026-0042",
		AffectedControlCount:  3,
		AffectedFrameworks:    []string{"ISO27001", "NIST_800_53"},
		ComplianceScoreImpact: 4.5,
		RiskExposureIncrease:  "moderate",
		CompensatingCoverage:  75.0,
		NetRiskDelta:          "low",
		AffectedControls: []AffectedControl{
			{
				ControlCode: "A.9.4.2",
				ControlName: "Secure Log-on Procedures",
				Framework:   "ISO27001",
				Impact:      "excepted",
			},
		},
		Recommendations: []string{
			"Monitor compensating control effectiveness quarterly",
		},
	}

	data, err := json.Marshal(impact)
	if err != nil {
		t.Fatalf("Failed to marshal ComplianceImpact: %v", err)
	}

	var decoded ComplianceImpact
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal ComplianceImpact: %v", err)
	}

	if decoded.AffectedControlCount != 3 {
		t.Errorf("Expected 3 affected controls, got %d", decoded.AffectedControlCount)
	}
	if decoded.ComplianceScoreImpact != 4.5 {
		t.Errorf("Expected score impact 4.5, got %v", decoded.ComplianceScoreImpact)
	}
	if decoded.CompensatingCoverage != 75.0 {
		t.Errorf("Expected coverage 75.0, got %v", decoded.CompensatingCoverage)
	}
	if len(decoded.AffectedControls) != 1 {
		t.Errorf("Expected 1 affected control, got %d", len(decoded.AffectedControls))
	}
}

// ============================================================
// HELPER FUNCTIONS FOR TESTS
// ============================================================

// calculateImpactForTest simulates the compliance impact calculation
// used in CalculateComplianceImpact without requiring a DB connection.
func calculateImpactForTest(
	residualRiskLevel string,
	controlCount int,
	hasCompensating bool,
	effectiveness string,
	exceptionType string,
	renewalCount int,
	expiryDate *time.Time,
) ComplianceImpact {
	impact := ComplianceImpact{}

	// Score impact
	var scoreImpact float64
	switch residualRiskLevel {
	case "critical":
		scoreImpact = float64(controlCount) * 2.5
	case "high":
		scoreImpact = float64(controlCount) * 1.5
	case "medium":
		scoreImpact = float64(controlCount) * 0.8
	case "low":
		scoreImpact = float64(controlCount) * 0.3
	default:
		scoreImpact = float64(controlCount) * 0.5
	}
	impact.ComplianceScoreImpact = float64(int(scoreImpact*10+0.5)) / 10

	// Risk exposure
	switch residualRiskLevel {
	case "critical":
		impact.RiskExposureIncrease = "significant"
	case "high":
		impact.RiskExposureIncrease = "moderate"
	case "medium":
		impact.RiskExposureIncrease = "limited"
	default:
		impact.RiskExposureIncrease = "minimal"
	}

	// Compensating coverage
	if hasCompensating {
		switch effectiveness {
		case "fully_effective":
			impact.CompensatingCoverage = 95.0
		case "mostly_effective":
			impact.CompensatingCoverage = 75.0
		case "partially_effective":
			impact.CompensatingCoverage = 50.0
		case "minimally_effective":
			impact.CompensatingCoverage = 25.0
		case "not_effective":
			impact.CompensatingCoverage = 5.0
		default:
			impact.CompensatingCoverage = 0.0
		}
	}

	// Net risk delta
	if impact.CompensatingCoverage >= 75 {
		impact.NetRiskDelta = "low"
	} else if impact.CompensatingCoverage >= 50 {
		impact.NetRiskDelta = "moderate"
	} else if residualRiskLevel == "critical" || residualRiskLevel == "high" {
		impact.NetRiskDelta = "high"
	} else {
		impact.NetRiskDelta = "moderate"
	}

	// Recommendations
	var recs []string
	if !hasCompensating {
		recs = append(recs, "Consider implementing compensating controls to reduce risk exposure")
	}
	if effectiveness == "not_assessed" || effectiveness == "not_effective" {
		recs = append(recs, "Assess or improve compensating control effectiveness")
	}
	if exceptionType == "temporary" && renewalCount >= 2 {
		recs = append(recs, "Maximum renewals reached; plan for permanent remediation")
	}
	if residualRiskLevel == "critical" || residualRiskLevel == "high" {
		recs = append(recs, "High residual risk; schedule more frequent reviews")
	}
	if expiryDate != nil && expiryDate.Before(time.Now().AddDate(0, 0, 30)) {
		recs = append(recs, "Exception expiring soon; initiate renewal or remediation planning")
	}
	if len(recs) == 0 {
		recs = append(recs, "Exception is well-managed with adequate compensating controls")
	}
	impact.Recommendations = recs

	return impact
}

// validateExceptionDates validates effective and expiry date strings.
func validateExceptionDates(effectiveDateStr, expiryDateStr string) error {
	effectiveDate, err := time.Parse("2006-01-02", effectiveDateStr)
	if err != nil {
		return fmt.Errorf("invalid effective_date format: use YYYY-MM-DD")
	}

	if expiryDateStr != "" {
		expiryDate, err := time.Parse("2006-01-02", expiryDateStr)
		if err != nil {
			return fmt.Errorf("invalid expiry_date format: use YYYY-MM-DD")
		}
		if !expiryDate.After(effectiveDate) {
			return fmt.Errorf("expiry_date must be after effective_date")
		}
	}

	return nil
}

// validateRenewal checks if an exception can be renewed based on type and count.
func validateRenewal(exceptionType string, currentRenewalCount int) error {
	if exceptionType == "temporary" && currentRenewalCount >= 2 {
		return fmt.Errorf("maximum renewal count (2) reached for temporary exceptions")
	}
	return nil
}

// isValidTransition checks if a status transition is valid for a given action.
func isValidTransition(currentStatus, action string) bool {
	switch action {
	case "submit":
		return currentStatus == "draft"
	case "approve":
		return currentStatus == "pending_risk_assessment" || currentStatus == "pending_approval"
	case "reject":
		return currentStatus == "pending_risk_assessment" || currentStatus == "pending_approval"
	case "revoke":
		return currentStatus == "approved"
	case "renew":
		return currentStatus == "approved" || currentStatus == "renewal_pending" || currentStatus == "expired"
	default:
		return false
	}
}

// validateCreateRequest validates a CreateExceptionRequest.
func validateCreateRequest(req CreateExceptionRequest) error {
	if req.Title == "" {
		return fmt.Errorf("title is required")
	}
	if req.Description == "" {
		return fmt.Errorf("description is required")
	}
	if req.RiskJustification == "" {
		return fmt.Errorf("risk_justification is required")
	}
	if req.EffectiveDate == "" {
		return fmt.Errorf("effective_date is required")
	}
	return nil
}
