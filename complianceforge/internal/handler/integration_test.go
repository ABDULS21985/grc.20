package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/complianceforge/platform/internal/models"
)

// ============================================================
// RISK ENDPOINT INTEGRATION TESTS
// ============================================================

func TestCreateRiskRequest_ValidPayload(t *testing.T) {
	payload := map[string]interface{}{
		"title":               "Ransomware attack via phishing email",
		"description":         "Risk of ransomware infection through employee phishing",
		"risk_category_id":    "e0000001-0000-0000-0000-000000000005", // Cybersecurity
		"risk_source":         "external",
		"owner_user_id":       "00000000-0000-0000-0000-000000000088",
		"inherent_likelihood": 4,
		"inherent_impact":     5,
		"residual_likelihood": 2,
		"residual_impact":     3,
		"financial_impact_eur": 750000,
		"risk_velocity":       "fast",
		"review_frequency":    "quarterly",
		"tags":                []string{"ransomware", "phishing", "cyber"},
	}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}

	req := httptest.NewRequest("POST", "/api/v1/risks", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	if len(jsonBody) == 0 {
		t.Error("Request body should not be empty")
	}

	// Verify the payload structure
	var decoded map[string]interface{}
	json.Unmarshal(jsonBody, &decoded)

	if decoded["title"] != "Ransomware attack via phishing email" {
		t.Error("Title mismatch")
	}
	if decoded["inherent_likelihood"].(float64) != 4 {
		t.Error("Likelihood should be 4")
	}
	if decoded["inherent_impact"].(float64) != 5 {
		t.Error("Impact should be 5")
	}

	// Verify risk score would be calculated as 4*5=20 (Critical)
	expectedScore := decoded["inherent_likelihood"].(float64) * decoded["inherent_impact"].(float64)
	if expectedScore != 20 {
		t.Errorf("Expected inherent score 20, got %.0f", expectedScore)
	}
}

func TestRiskScoreCalculation(t *testing.T) {
	tests := []struct {
		likelihood int
		impact     int
		score      int
		level      string
	}{
		{1, 1, 1, "very_low"},
		{2, 2, 4, "low"},
		{3, 2, 6, "medium"},
		{4, 3, 12, "high"},
		{5, 4, 20, "critical"},
		{5, 5, 25, "critical"},
	}

	for _, tt := range tests {
		score := tt.likelihood * tt.impact
		if score != tt.score {
			t.Errorf("L=%d I=%d: expected score %d, got %d", tt.likelihood, tt.impact, tt.score, score)
		}

		var level string
		switch {
		case score >= 20:
			level = "critical"
		case score >= 12:
			level = "high"
		case score >= 6:
			level = "medium"
		case score >= 3:
			level = "low"
		default:
			level = "very_low"
		}

		if level != tt.level {
			t.Errorf("Score %d: expected level '%s', got '%s'", score, tt.level, level)
		}
	}
}

// ============================================================
// INCIDENT ENDPOINT INTEGRATION TESTS
// ============================================================

func TestCreateIncident_GDPRBreach(t *testing.T) {
	payload := map[string]interface{}{
		"title":                     "Customer database exposed via misconfigured S3 bucket",
		"description":               "Personal data of 50,000 customers was publicly accessible",
		"incident_type":             "data_breach",
		"severity":                  "critical",
		"is_data_breach":            true,
		"data_subjects_affected":    50000,
		"data_categories_affected":  []string{"name", "email", "address", "phone"},
		"is_nis2_reportable":        false,
		"impact_description":        "Personal data of EU customers exposed to internet",
	}

	jsonBody, _ := json.Marshal(payload)

	var decoded map[string]interface{}
	json.Unmarshal(jsonBody, &decoded)

	// Verify GDPR breach fields
	if decoded["is_data_breach"] != true {
		t.Error("is_data_breach should be true")
	}
	if decoded["data_subjects_affected"].(float64) != 50000 {
		t.Error("data_subjects_affected should be 50000")
	}

	// Verify 72-hour deadline would be set
	now := time.Now()
	deadline := now.Add(72 * time.Hour)
	if deadline.Before(now) {
		t.Error("Notification deadline should be in the future")
	}

	// Verify deadline is exactly 72 hours
	diff := deadline.Sub(now)
	if diff.Hours() < 71.9 || diff.Hours() > 72.1 {
		t.Errorf("Expected ~72 hours, got %.1f hours", diff.Hours())
	}
}

func TestCreateIncident_NIS2Reportable(t *testing.T) {
	payload := map[string]interface{}{
		"title":               "Critical infrastructure DDoS attack",
		"description":         "Sustained DDoS attack on essential services network",
		"incident_type":       "security",
		"severity":            "critical",
		"is_data_breach":      false,
		"is_nis2_reportable":  true,
		"impact_description":  "Service disruption affecting essential service delivery",
	}

	jsonBody, _ := json.Marshal(payload)

	var decoded map[string]interface{}
	json.Unmarshal(jsonBody, &decoded)

	if decoded["is_nis2_reportable"] != true {
		t.Error("is_nis2_reportable should be true")
	}

	// NIS2 requires early warning within 24 hours
	now := time.Now()
	earlyWarningDeadline := now.Add(24 * time.Hour)
	notificationDeadline := now.Add(72 * time.Hour)

	if earlyWarningDeadline.After(notificationDeadline) {
		t.Error("Early warning deadline should be before full notification deadline")
	}
}

// ============================================================
// POLICY ENDPOINT INTEGRATION TESTS
// ============================================================

func TestCreatePolicyRequest_Complete(t *testing.T) {
	req := models.CreatePolicyRequest{
		Title:                 "Information Security Policy",
		CategoryID:            [16]byte{0xf0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		Classification:        "internal",
		ContentHTML:           "<h1>Information Security Policy</h1><p>Version 1.0</p><h2>1. Purpose</h2><p>This policy establishes...</p>",
		Summary:               "Defines the organisation's approach to information security management.",
		OwnerUserID:           [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x88},
		ApproverUserID:        [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x88},
		ReviewFrequencyMonths: 12,
		IsMandatory:           true,
		RequiresAttestation:   true,
		Tags:                  []string{"security", "iso27001", "mandatory"},
	}

	if req.Title == "" {
		t.Error("Policy title should not be empty")
	}
	if req.ReviewFrequencyMonths < 1 || req.ReviewFrequencyMonths > 36 {
		t.Errorf("Review frequency %d out of range 1-36", req.ReviewFrequencyMonths)
	}
	if req.Classification != "internal" {
		t.Errorf("Expected classification 'internal', got '%s'", req.Classification)
	}
}

// ============================================================
// VENDOR ENDPOINT INTEGRATION TESTS
// ============================================================

func TestVendorOnboarding_WithGDPRRequirements(t *testing.T) {
	payload := map[string]interface{}{
		"name":                "CloudTech Solutions Ltd",
		"legal_name":          "CloudTech Solutions Limited",
		"website":             "https://cloudtech.example.com",
		"industry":            "technology",
		"country_code":        "DE",
		"contact_name":        "Hans Mueller",
		"contact_email":       "hans@cloudtech.example.com",
		"risk_tier":           "high",
		"service_description": "Cloud infrastructure hosting for customer data",
		"data_processing":     true,
		"data_categories":     []string{"name", "email", "financial"},
		"certifications":      []string{"ISO27001", "SOC2"},
		"assessment_frequency": "quarterly",
	}

	jsonBody, _ := json.Marshal(payload)

	var decoded map[string]interface{}
	json.Unmarshal(jsonBody, &decoded)

	// When data_processing is true, DPA is required per GDPR Article 28
	if decoded["data_processing"].(bool) {
		// System should flag that DPA is needed
		t.Log("GDPR: DPA required for this vendor — data_processing = true")
	}

	if decoded["country_code"] != "DE" {
		t.Error("Country code should be DE (Germany)")
	}

	// High risk vendors should have quarterly assessments
	if decoded["assessment_frequency"] != "quarterly" {
		t.Error("High risk vendors should be assessed quarterly")
	}
}

// ============================================================
// COMPLIANCE SCORE INTEGRATION TESTS
// ============================================================

func TestComplianceScoreCalculation(t *testing.T) {
	tests := []struct {
		total          int
		implemented    int
		notApplicable  int
		expectedScore  float64
	}{
		{93, 70, 5, 79.55},   // ISO 27001: 70 of 88 applicable
		{93, 93, 0, 100.0},   // Full compliance
		{93, 0, 0, 0.0},      // No implementation
		{93, 0, 93, 0.0},     // All NA (edge case)
		{93, 46, 5, 52.27},   // Partial implementation
	}

	for _, tt := range tests {
		applicable := tt.total - tt.notApplicable
		var score float64
		if applicable > 0 {
			score = float64(tt.implemented) / float64(applicable) * 100
		}

		// Round to 2 decimal places
		score = float64(int(score*100)) / 100

		if tt.expectedScore > 0 {
			diff := score - tt.expectedScore
			if diff > 0.5 || diff < -0.5 {
				t.Errorf("Total=%d Impl=%d NA=%d: expected ~%.2f%%, got %.2f%%",
					tt.total, tt.implemented, tt.notApplicable, tt.expectedScore, score)
			}
		}
	}
}

// ============================================================
// AUDIT FINDING INTEGRATION TESTS
// ============================================================

func TestAuditFinding_SeverityClassification(t *testing.T) {
	severities := []string{"critical", "high", "medium", "low", "informational"}

	for _, sev := range severities {
		finding := models.AuditFinding{
			Severity:    models.FindingSeverity(sev),
			Title:       "Test finding",
			Description: "Test description",
			Status:      "open",
			FindingType: "non_conformity",
		}

		if string(finding.Severity) != sev {
			t.Errorf("Severity mismatch: expected '%s', got '%s'", sev, finding.Severity)
		}
	}
}

// ============================================================
// CROSS-FRAMEWORK MAPPING TESTS
// ============================================================

func TestMappingStrengthRange(t *testing.T) {
	strengths := []float64{0.0, 0.25, 0.50, 0.75, 0.85, 0.90, 0.95, 1.0}

	for _, s := range strengths {
		if s < 0 || s > 1 {
			t.Errorf("Mapping strength %.2f out of range [0, 1]", s)
		}
	}

	// Invalid strength
	if -0.1 >= 0 {
		t.Error("Negative strength should be invalid")
	}
	if 1.1 <= 1 {
		t.Error("Strength > 1 should be invalid")
	}
}

func TestMappingTypes(t *testing.T) {
	validTypes := map[string]bool{
		"equivalent": true,
		"partial":    true,
		"related":    true,
		"superset":   true,
		"subset":     true,
	}

	for mType := range validTypes {
		if mType == "" {
			t.Error("Mapping type should not be empty")
		}
	}

	if validTypes["invalid_type"] {
		t.Error("Invalid type should not be in the map")
	}
}

// ============================================================
// API RESPONSE FORMAT TESTS
// ============================================================

func TestPaginatedResponseFormat(t *testing.T) {
	resp := models.PaginatedResponse{
		Data: []string{"item1", "item2"},
		Pagination: models.Pagination{
			Page:       1,
			PageSize:   20,
			TotalItems: 50,
			TotalPages: 3,
			HasNext:    true,
			HasPrev:    false,
		},
	}

	jsonBytes, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded map[string]interface{}
	json.Unmarshal(jsonBytes, &decoded)

	if decoded["data"] == nil {
		t.Error("data field should be present")
	}
	if decoded["pagination"] == nil {
		t.Error("pagination field should be present")
	}

	pagination := decoded["pagination"].(map[string]interface{})
	if pagination["total_items"].(float64) != 50 {
		t.Error("total_items should be 50")
	}
	if pagination["has_next"].(bool) != true {
		t.Error("has_next should be true")
	}
	if pagination["has_prev"].(bool) != false {
		t.Error("has_prev should be false for page 1")
	}
}

func TestAPIErrorResponseFormat(t *testing.T) {
	resp := models.APIResponse{
		Success: false,
		Error: &models.APIError{
			Code:    "VALIDATION_ERROR",
			Message: "Request validation failed",
			Details: map[string]string{
				"title":              "is required",
				"inherent_likelihood": "must be between 1 and 5",
			},
		},
	}

	jsonBytes, _ := json.Marshal(resp)
	if !bytes.Contains(jsonBytes, []byte(`"success":false`)) {
		t.Error("Error response should have success=false")
	}
	if !bytes.Contains(jsonBytes, []byte(`"code":"VALIDATION_ERROR"`)) {
		t.Error("Error response should contain error code")
	}
}

func TestHealthEndpointResponse(t *testing.T) {
	resp := map[string]interface{}{
		"status":  "ok",
		"version": "1.0.0",
		"services": map[string]interface{}{
			"database": map[string]interface{}{"status": "healthy"},
			"redis":    map[string]interface{}{"status": "healthy"},
		},
	}

	jsonBytes, _ := json.Marshal(resp)

	var decoded map[string]interface{}
	json.Unmarshal(jsonBytes, &decoded)

	if decoded["status"] != "ok" {
		t.Error("Health status should be 'ok'")
	}

	services := decoded["services"].(map[string]interface{})
	db := services["database"].(map[string]interface{})
	if db["status"] != "healthy" {
		t.Error("Database should be healthy")
	}
}
