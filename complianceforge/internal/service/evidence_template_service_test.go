package service

import (
	"encoding/json"
	"testing"
	"time"
)

// ============================================================
// VALIDATION RULE ENGINE TESTS
// ============================================================

func TestRunValidationRules_FileNotEmpty_Pass(t *testing.T) {
	rules := []ValidationRule{
		{RuleType: "file_not_empty"},
	}

	results := RunValidationRules(rules, 1024, "report.pdf", "application/pdf", time.Now())

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	if !results[0].Passed {
		t.Error("Expected file_not_empty to pass for file size 1024")
	}
	if results[0].RuleType != "file_not_empty" {
		t.Errorf("Expected rule_type 'file_not_empty', got '%s'", results[0].RuleType)
	}
}

func TestRunValidationRules_FileNotEmpty_Fail(t *testing.T) {
	rules := []ValidationRule{
		{RuleType: "file_not_empty"},
	}

	results := RunValidationRules(rules, 0, "empty.pdf", "application/pdf", time.Now())

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	if results[0].Passed {
		t.Error("Expected file_not_empty to fail for file size 0")
	}
}

func TestRunValidationRules_DateWithin_Pass(t *testing.T) {
	rules := []ValidationRule{
		{
			RuleType: "date_within",
			Params:   map[string]interface{}{"days": float64(90)},
		},
	}

	// Evidence collected 30 days ago — within 90-day window
	collectedAt := time.Now().AddDate(0, 0, -30)
	results := RunValidationRules(rules, 100, "report.pdf", "application/pdf", collectedAt)

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	if !results[0].Passed {
		t.Error("Expected date_within to pass for evidence collected 30 days ago with 90-day window")
	}
}

func TestRunValidationRules_DateWithin_Fail(t *testing.T) {
	rules := []ValidationRule{
		{
			RuleType: "date_within",
			Params:   map[string]interface{}{"days": float64(90)},
		},
	}

	// Evidence collected 120 days ago — outside 90-day window
	collectedAt := time.Now().AddDate(0, 0, -120)
	results := RunValidationRules(rules, 100, "report.pdf", "application/pdf", collectedAt)

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	if results[0].Passed {
		t.Error("Expected date_within to fail for evidence collected 120 days ago with 90-day window")
	}
}

func TestRunValidationRules_DateWithin_DefaultDays(t *testing.T) {
	rules := []ValidationRule{
		{RuleType: "date_within"},
	}

	// Evidence collected 300 days ago — within default 365-day window
	collectedAt := time.Now().AddDate(0, 0, -300)
	results := RunValidationRules(rules, 100, "report.pdf", "application/pdf", collectedAt)

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	if !results[0].Passed {
		t.Error("Expected date_within to pass with default 365-day window for 300-day-old evidence")
	}
}

func TestRunValidationRules_DateWithin_DefaultDays_Fail(t *testing.T) {
	rules := []ValidationRule{
		{RuleType: "date_within"},
	}

	// Evidence collected 400 days ago — outside default 365-day window
	collectedAt := time.Now().AddDate(0, 0, -400)
	results := RunValidationRules(rules, 100, "report.pdf", "application/pdf", collectedAt)

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	if results[0].Passed {
		t.Error("Expected date_within to fail with default 365-day window for 400-day-old evidence")
	}
}

func TestRunValidationRules_ContainsText_Pass(t *testing.T) {
	rules := []ValidationRule{
		{
			RuleType: "contains_text",
			Params:   map[string]interface{}{"text": "approved"},
		},
	}

	results := RunValidationRules(rules, 100, "policy_approved_2026.pdf", "application/pdf", time.Now())

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	if !results[0].Passed {
		t.Error("Expected contains_text to pass for filename containing 'approved'")
	}
}

func TestRunValidationRules_ContainsText_CaseInsensitive(t *testing.T) {
	rules := []ValidationRule{
		{
			RuleType: "contains_text",
			Params:   map[string]interface{}{"text": "APPROVED"},
		},
	}

	results := RunValidationRules(rules, 100, "policy_approved_2026.pdf", "application/pdf", time.Now())

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	if !results[0].Passed {
		t.Error("Expected contains_text to pass with case-insensitive matching")
	}
}

func TestRunValidationRules_ContainsText_Fail(t *testing.T) {
	rules := []ValidationRule{
		{
			RuleType: "contains_text",
			Params:   map[string]interface{}{"text": "signed"},
		},
	}

	results := RunValidationRules(rules, 100, "policy_draft.pdf", "application/pdf", time.Now())

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	if results[0].Passed {
		t.Error("Expected contains_text to fail for filename not containing 'signed'")
	}
}

func TestRunValidationRules_ContainsText_EmptyText(t *testing.T) {
	rules := []ValidationRule{
		{
			RuleType: "contains_text",
			Params:   map[string]interface{}{"text": ""},
		},
	}

	results := RunValidationRules(rules, 100, "policy.pdf", "application/pdf", time.Now())

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	if !results[0].Passed {
		t.Error("Expected contains_text to pass for empty text parameter")
	}
}

func TestRunValidationRules_FileType_Pass(t *testing.T) {
	rules := []ValidationRule{
		{
			RuleType: "file_type",
			Params:   map[string]interface{}{"allowed": "pdf,docx"},
		},
	}

	results := RunValidationRules(rules, 100, "report.pdf", "application/pdf", time.Now())

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	if !results[0].Passed {
		t.Error("Expected file_type to pass for PDF file with 'pdf,docx' allowed")
	}
}

func TestRunValidationRules_FileType_ByMimeType(t *testing.T) {
	rules := []ValidationRule{
		{
			RuleType: "file_type",
			Params:   map[string]interface{}{"allowed": "pdf"},
		},
	}

	results := RunValidationRules(rules, 100, "report.document", "application/pdf", time.Now())

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	if !results[0].Passed {
		t.Error("Expected file_type to pass when MIME type contains 'pdf'")
	}
}

func TestRunValidationRules_FileType_Fail(t *testing.T) {
	rules := []ValidationRule{
		{
			RuleType: "file_type",
			Params:   map[string]interface{}{"allowed": "pdf,docx"},
		},
	}

	results := RunValidationRules(rules, 100, "report.exe", "application/x-executable", time.Now())

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	if results[0].Passed {
		t.Error("Expected file_type to fail for .exe file with 'pdf,docx' allowed")
	}
}

func TestRunValidationRules_FileSize_WithinLimits(t *testing.T) {
	rules := []ValidationRule{
		{
			RuleType: "file_size",
			Params:   map[string]interface{}{"min_bytes": float64(100), "max_bytes": float64(10000000)},
		},
	}

	results := RunValidationRules(rules, 5000, "report.pdf", "application/pdf", time.Now())

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	if !results[0].Passed {
		t.Error("Expected file_size to pass for 5000 bytes within 100-10000000 range")
	}
}

func TestRunValidationRules_FileSize_TooSmall(t *testing.T) {
	rules := []ValidationRule{
		{
			RuleType: "file_size",
			Params:   map[string]interface{}{"min_bytes": float64(1000)},
		},
	}

	results := RunValidationRules(rules, 500, "report.pdf", "application/pdf", time.Now())

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	if results[0].Passed {
		t.Error("Expected file_size to fail for 500 bytes with min_bytes 1000")
	}
}

func TestRunValidationRules_FileSize_TooLarge(t *testing.T) {
	rules := []ValidationRule{
		{
			RuleType: "file_size",
			Params:   map[string]interface{}{"max_bytes": float64(1000)},
		},
	}

	results := RunValidationRules(rules, 5000, "report.pdf", "application/pdf", time.Now())

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	if results[0].Passed {
		t.Error("Expected file_size to fail for 5000 bytes with max_bytes 1000")
	}
}

func TestRunValidationRules_UnknownRuleType(t *testing.T) {
	rules := []ValidationRule{
		{RuleType: "custom_unknown_check"},
	}

	results := RunValidationRules(rules, 100, "report.pdf", "application/pdf", time.Now())

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	if !results[0].Passed {
		t.Error("Expected unknown rule type to pass (skip with true)")
	}
	if results[0].RuleType != "custom_unknown_check" {
		t.Errorf("Expected rule_type 'custom_unknown_check', got '%s'", results[0].RuleType)
	}
}

func TestRunValidationRules_MultipleRules_AllPass(t *testing.T) {
	rules := []ValidationRule{
		{RuleType: "file_not_empty"},
		{
			RuleType: "date_within",
			Params:   map[string]interface{}{"days": float64(365)},
		},
		{
			RuleType: "file_type",
			Params:   map[string]interface{}{"allowed": "pdf"},
		},
	}

	results := RunValidationRules(rules, 1024, "report.pdf", "application/pdf", time.Now())

	if len(results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(results))
	}

	allPassed := true
	for _, r := range results {
		if !r.Passed {
			allPassed = false
			t.Errorf("Rule '%s' failed: %s", r.RuleType, r.Message)
		}
	}
	if !allPassed {
		t.Error("Expected all rules to pass")
	}
}

func TestRunValidationRules_MultipleRules_SomeFail(t *testing.T) {
	rules := []ValidationRule{
		{RuleType: "file_not_empty"},
		{
			RuleType: "date_within",
			Params:   map[string]interface{}{"days": float64(30)},
		},
		{
			RuleType: "file_type",
			Params:   map[string]interface{}{"allowed": "pdf"},
		},
	}

	// File is non-empty (pass), but collected 60 days ago (fail for 30-day window)
	collectedAt := time.Now().AddDate(0, 0, -60)
	results := RunValidationRules(rules, 1024, "report.pdf", "application/pdf", collectedAt)

	if len(results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(results))
	}

	if !results[0].Passed {
		t.Error("Expected file_not_empty to pass")
	}
	if results[1].Passed {
		t.Error("Expected date_within to fail for 60-day-old evidence with 30-day window")
	}
	if !results[2].Passed {
		t.Error("Expected file_type to pass for PDF")
	}
}

func TestRunValidationRules_EmptyRules(t *testing.T) {
	rules := []ValidationRule{}

	results := RunValidationRules(rules, 1024, "report.pdf", "application/pdf", time.Now())

	if len(results) != 0 {
		t.Errorf("Expected 0 results for empty rules, got %d", len(results))
	}
}

func TestRunValidationRules_NilParams(t *testing.T) {
	rules := []ValidationRule{
		{RuleType: "date_within", Params: nil},
		{RuleType: "file_type", Params: nil},
		{RuleType: "file_size", Params: nil},
		{RuleType: "contains_text", Params: nil},
	}

	// Should not panic with nil params
	results := RunValidationRules(rules, 1024, "report.pdf", "application/pdf", time.Now())

	if len(results) != 4 {
		t.Fatalf("Expected 4 results, got %d", len(results))
	}

	// date_within with nil params defaults to 365 days
	if !results[0].Passed {
		t.Error("Expected date_within with nil params to pass (uses default 365 days)")
	}
	// file_type with nil params should pass (empty allowed = all allowed)
	if !results[1].Passed {
		t.Error("Expected file_type with nil params to pass")
	}
	// file_size with nil params should pass (no limits set)
	if !results[2].Passed {
		t.Error("Expected file_size with nil params to pass")
	}
	// contains_text with nil params should pass (empty text)
	if !results[3].Passed {
		t.Error("Expected contains_text with nil params to pass")
	}
}

// ============================================================
// HELPER FUNCTION TESTS
// ============================================================

func TestCalculateNextDue(t *testing.T) {
	now := time.Now()

	tests := []struct {
		frequency string
		minDays   int
		maxDays   int
	}{
		{"daily", 0, 2},
		{"weekly", 6, 8},
		{"monthly", 27, 32},
		{"quarterly", 88, 93},
		{"semi_annually", 178, 185},
		{"annually", 364, 367},
		{"once", 27, 32},
		{"unknown", 88, 93}, // defaults to quarterly
	}

	for _, tc := range tests {
		t.Run(tc.frequency, func(t *testing.T) {
			result := calculateNextDue(tc.frequency)
			daysDiff := int(result.Sub(now).Hours() / 24)
			if daysDiff < tc.minDays || daysDiff > tc.maxDays {
				t.Errorf("calculateNextDue(%q) = %d days, want between %d and %d",
					tc.frequency, daysDiff, tc.minDays, tc.maxDays)
			}
		})
	}
}

func TestContainsFileType(t *testing.T) {
	tests := []struct {
		name         string
		mimeType     string
		fileName     string
		allowedTypes string
		expected     bool
	}{
		{"PDF by extension", "application/octet-stream", "report.pdf", "pdf", true},
		{"PDF by MIME", "application/pdf", "report.bin", "pdf", true},
		{"DOCX by extension", "application/octet-stream", "doc.docx", "pdf,docx", true},
		{"EXE rejected", "application/x-executable", "malware.exe", "pdf,docx", false},
		{"Empty allowed", "application/pdf", "report.pdf", "", false},
		{"Case insensitive", "application/PDF", "report.PDF", "pdf", true},
		{"CSV by extension", "text/csv", "data.csv", "csv,xlsx", true},
		{"PNG by MIME", "image/png", "screenshot.png", "png,jpg", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := containsFileType(tc.mimeType, tc.fileName, tc.allowedTypes)
			if result != tc.expected {
				t.Errorf("containsFileType(%q, %q, %q) = %v, want %v",
					tc.mimeType, tc.fileName, tc.allowedTypes, result, tc.expected)
			}
		})
	}
}

// ============================================================
// MODEL SERIALIZATION TESTS
// ============================================================

func TestEvidenceTemplateSerialization(t *testing.T) {
	tmpl := EvidenceTemplate{
		Name:                  "Information Security Policy Document",
		FrameworkControlCode:  "A.5.1",
		FrameworkCode:         "ISO27001",
		EvidenceCategory:      "policy_document",
		CollectionMethod:      "document_review",
		CollectionFrequency:   "annually",
		Difficulty:            "easy",
		AuditorPriority:       "critical",
		IsSystem:              true,
		ValidationRules:       json.RawMessage(`[{"rule_type":"file_not_empty"},{"rule_type":"date_within","params":{"days":365}}]`),
		Tags:                  []string{"policy", "governance"},
		CommonRejectionReasons: []string{"Missing signature", "Outdated"},
	}

	data, err := json.Marshal(tmpl)
	if err != nil {
		t.Fatalf("Failed to marshal EvidenceTemplate: %v", err)
	}

	var decoded EvidenceTemplate
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal EvidenceTemplate: %v", err)
	}

	if decoded.Name != "Information Security Policy Document" {
		t.Errorf("Expected name 'Information Security Policy Document', got '%s'", decoded.Name)
	}
	if decoded.FrameworkControlCode != "A.5.1" {
		t.Errorf("Expected framework_control_code 'A.5.1', got '%s'", decoded.FrameworkControlCode)
	}
	if decoded.FrameworkCode != "ISO27001" {
		t.Errorf("Expected framework_code 'ISO27001', got '%s'", decoded.FrameworkCode)
	}
	if decoded.EvidenceCategory != "policy_document" {
		t.Errorf("Expected evidence_category 'policy_document', got '%s'", decoded.EvidenceCategory)
	}
	if !decoded.IsSystem {
		t.Error("Expected is_system = true")
	}
	if decoded.Difficulty != "easy" {
		t.Errorf("Expected difficulty 'easy', got '%s'", decoded.Difficulty)
	}
	if decoded.AuditorPriority != "critical" {
		t.Errorf("Expected auditor_priority 'critical', got '%s'", decoded.AuditorPriority)
	}
	if len(decoded.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(decoded.Tags))
	}
	if len(decoded.CommonRejectionReasons) != 2 {
		t.Errorf("Expected 2 rejection reasons, got %d", len(decoded.CommonRejectionReasons))
	}
}

func TestValidationResultSerialization(t *testing.T) {
	result := ValidationResult{
		Valid:         false,
		OverallStatus: "fail",
		RuleResults: []RuleResult{
			{RuleType: "file_not_empty", Passed: true, Message: "File size: 1024 bytes"},
			{RuleType: "date_within", Passed: false, Message: "Evidence older than 90 days"},
		},
		ValidatedAt: time.Now(),
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to marshal ValidationResult: %v", err)
	}

	var decoded ValidationResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal ValidationResult: %v", err)
	}

	if decoded.Valid {
		t.Error("Expected valid = false")
	}
	if decoded.OverallStatus != "fail" {
		t.Errorf("Expected overall_status 'fail', got '%s'", decoded.OverallStatus)
	}
	if len(decoded.RuleResults) != 2 {
		t.Fatalf("Expected 2 rule results, got %d", len(decoded.RuleResults))
	}
	if !decoded.RuleResults[0].Passed {
		t.Error("Expected first rule to pass")
	}
	if decoded.RuleResults[1].Passed {
		t.Error("Expected second rule to fail")
	}
}

func TestEvidenceGapsResultSerialization(t *testing.T) {
	gaps := EvidenceGapsResult{
		TotalRequirements: 100,
		Collected:         75,
		Pending:           20,
		Overdue:           5,
		Validated:         70,
		Failed:            5,
		CoveragePercent:   75.0,
		GapsByFramework: []FrameworkGap{
			{FrameworkCode: "ISO27001", TotalRequired: 60, Collected: 50, CoveragePercent: 83.3},
			{FrameworkCode: "PCI_DSS_4", TotalRequired: 40, Collected: 25, CoveragePercent: 62.5},
		},
		CriticalGaps: []CriticalGapItem{},
		OverdueItems: []OverdueItem{},
	}

	data, err := json.Marshal(gaps)
	if err != nil {
		t.Fatalf("Failed to marshal EvidenceGapsResult: %v", err)
	}

	var decoded EvidenceGapsResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal EvidenceGapsResult: %v", err)
	}

	if decoded.TotalRequirements != 100 {
		t.Errorf("Expected total_requirements 100, got %d", decoded.TotalRequirements)
	}
	if decoded.CoveragePercent != 75.0 {
		t.Errorf("Expected coverage_percent 75.0, got %f", decoded.CoveragePercent)
	}
	if len(decoded.GapsByFramework) != 2 {
		t.Errorf("Expected 2 framework gaps, got %d", len(decoded.GapsByFramework))
	}
}

func TestPreAuditReportSerialization(t *testing.T) {
	report := PreAuditReport{
		FrameworkCode:    "ISO27001",
		OverallReadiness: 72.5,
		ReadinessLevel:   "mostly_ready",
		TotalControls:    93,
		ControlsWithEvidence: 70,
		ControlsMissingEvidence: 23,
		EvidenceCompletion: 75.3,
		ValidationPassRate: 92.1,
		CriticalGaps: []PreAuditGap{
			{
				ControlCode: "A.8.24",
				ControlName: "Encryption Configuration Evidence",
				GapType:     "missing_evidence",
				Severity:    "high",
			},
		},
		Recommendations: []string{
			"Focus on collecting evidence for critical controls first.",
			"Address 23 controls with no evidence collected.",
		},
		EstimatedRemediationHours: 92.0,
	}

	data, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("Failed to marshal PreAuditReport: %v", err)
	}

	var decoded PreAuditReport
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal PreAuditReport: %v", err)
	}

	if decoded.OverallReadiness != 72.5 {
		t.Errorf("Expected overall_readiness 72.5, got %f", decoded.OverallReadiness)
	}
	if decoded.ReadinessLevel != "mostly_ready" {
		t.Errorf("Expected readiness_level 'mostly_ready', got '%s'", decoded.ReadinessLevel)
	}
	if decoded.TotalControls != 93 {
		t.Errorf("Expected total_controls 93, got %d", decoded.TotalControls)
	}
	if len(decoded.CriticalGaps) != 1 {
		t.Errorf("Expected 1 critical gap, got %d", len(decoded.CriticalGaps))
	}
	if len(decoded.Recommendations) != 2 {
		t.Errorf("Expected 2 recommendations, got %d", len(decoded.Recommendations))
	}
	if decoded.EstimatedRemediationHours != 92.0 {
		t.Errorf("Expected estimated_remediation_hours 92.0, got %f", decoded.EstimatedRemediationHours)
	}
}
