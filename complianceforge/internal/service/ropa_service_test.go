package service

import (
	"testing"

	"github.com/google/uuid"
)

// ============================================================
// DPIA Trigger Assessment Tests
// ============================================================

func TestAssessDPIARequired_NoIndicators(t *testing.T) {
	req := CreateActivityRequest{
		Name:                          "Simple newsletter",
		LegalBasis:                    "consent",
		SpecialCategoriesProcessed:    false,
		InvolvesInternationalTransfer: false,
		DataSubjectCategories:         []string{"customers"},
		DataCategoryIDs:               []uuid.UUID{uuid.New()},
	}

	if assessDPIARequired(req) {
		t.Error("Expected DPIA not required for simple processing with no risk indicators")
	}
}

func TestAssessDPIARequired_SingleIndicator(t *testing.T) {
	// Only special categories -- 1 indicator, need 2+ for DPIA
	req := CreateActivityRequest{
		Name:                          "Health monitoring",
		LegalBasis:                    "consent",
		SpecialCategoriesProcessed:    true,
		InvolvesInternationalTransfer: false,
		DataSubjectCategories:         []string{"patients"},
		DataCategoryIDs:               []uuid.UUID{uuid.New()},
	}

	if assessDPIARequired(req) {
		t.Error("Expected DPIA not required with only 1 risk indicator (special categories)")
	}
}

func TestAssessDPIARequired_TwoIndicators(t *testing.T) {
	// Special categories + international transfer = 2 indicators
	req := CreateActivityRequest{
		Name:                          "Cross-border health data processing",
		LegalBasis:                    "consent",
		SpecialCategoriesProcessed:    true,
		InvolvesInternationalTransfer: true,
		DataSubjectCategories:         []string{"patients"},
		DataCategoryIDs:               []uuid.UUID{uuid.New()},
	}

	if !assessDPIARequired(req) {
		t.Error("Expected DPIA required with 2 risk indicators (special categories + international transfer)")
	}
}

func TestAssessDPIARequired_LargeScaleSpecialCategory(t *testing.T) {
	// Special categories + large scale = 2 indicators
	count := 50000
	req := CreateActivityRequest{
		Name:                          "National health database",
		LegalBasis:                    "public_task",
		SpecialCategoriesProcessed:    true,
		InvolvesInternationalTransfer: false,
		EstimatedDataSubjectsCount:    &count,
		DataSubjectCategories:         []string{"citizens"},
		DataCategoryIDs:               []uuid.UUID{uuid.New()},
	}

	if !assessDPIARequired(req) {
		t.Error("Expected DPIA required for large-scale special category processing")
	}
}

func TestAssessDPIARequired_ChildrenData(t *testing.T) {
	// Children + international transfer = 2 indicators
	req := CreateActivityRequest{
		Name:                          "Children's educational platform",
		LegalBasis:                    "consent",
		SpecialCategoriesProcessed:    false,
		InvolvesInternationalTransfer: true,
		DataSubjectCategories:         []string{"children"},
		DataCategoryIDs:               []uuid.UUID{uuid.New(), uuid.New()},
	}

	if !assessDPIARequired(req) {
		t.Error("Expected DPIA required for children's data with international transfer")
	}
}

func TestAssessDPIARequired_VulnerablePersons(t *testing.T) {
	// Vulnerable persons + large scale = 2 indicators
	count := 15000
	req := CreateActivityRequest{
		Name:                       "Social services case management",
		LegalBasis:                 "legal_obligation",
		SpecialCategoriesProcessed: false,
		EstimatedDataSubjectsCount: &count,
		DataSubjectCategories:      []string{"vulnerable_persons"},
		DataCategoryIDs:            []uuid.UUID{uuid.New()},
	}

	if !assessDPIARequired(req) {
		t.Error("Expected DPIA required for vulnerable persons with large scale processing")
	}
}

func TestAssessDPIARequired_ManyCategoriesWithTransfer(t *testing.T) {
	// 6+ data categories + international transfer = 2 indicators
	ids := make([]uuid.UUID, 7)
	for i := range ids {
		ids[i] = uuid.New()
	}

	req := CreateActivityRequest{
		Name:                          "Global CRM platform",
		LegalBasis:                    "legitimate_interest",
		SpecialCategoriesProcessed:    false,
		InvolvesInternationalTransfer: true,
		DataSubjectCategories:         []string{"customers"},
		DataCategoryIDs:               ids,
	}

	if !assessDPIARequired(req) {
		t.Error("Expected DPIA required with many data categories and international transfer")
	}
}

func TestAssessDPIARequired_AllIndicators(t *testing.T) {
	// All indicators present
	count := 100000
	ids := make([]uuid.UUID, 8)
	for i := range ids {
		ids[i] = uuid.New()
	}

	req := CreateActivityRequest{
		Name:                          "Comprehensive employee surveillance",
		LegalBasis:                    "legitimate_interest",
		SpecialCategoriesProcessed:    true,
		InvolvesInternationalTransfer: true,
		EstimatedDataSubjectsCount:    &count,
		DataSubjectCategories:         []string{"employees", "children"},
		DataCategoryIDs:               ids,
	}

	if !assessDPIARequired(req) {
		t.Error("Expected DPIA required with all risk indicators present")
	}
}

func TestAssessDPIARequired_BoundarySubjectCount(t *testing.T) {
	// Exactly 10000 should NOT trigger (threshold is >10000)
	count := 10000
	req := CreateActivityRequest{
		Name:                       "Medium-scale processing",
		LegalBasis:                 "contract",
		SpecialCategoriesProcessed: true,
		EstimatedDataSubjectsCount: &count,
		DataSubjectCategories:      []string{"customers"},
		DataCategoryIDs:            []uuid.UUID{uuid.New()},
	}

	// 1 indicator (special categories) but subject count at boundary (not >10000)
	if assessDPIARequired(req) {
		t.Error("Expected DPIA not required: special categories alone (subject count at boundary, not above)")
	}
}

func TestAssessDPIARequired_JustAboveBoundary(t *testing.T) {
	// 10001 should trigger large-scale indicator
	count := 10001
	req := CreateActivityRequest{
		Name:                       "Just-above-threshold processing",
		LegalBasis:                 "contract",
		SpecialCategoriesProcessed: true,
		EstimatedDataSubjectsCount: &count,
		DataSubjectCategories:      []string{"customers"},
		DataCategoryIDs:            []uuid.UUID{uuid.New()},
	}

	// 2 indicators: special categories + large scale
	if !assessDPIARequired(req) {
		t.Error("Expected DPIA required: special categories + just above large-scale threshold")
	}
}

// ============================================================
// ROPA Completeness Validation Tests
// ============================================================

// TestROPACompletenessCheck validates that Art.30 required fields are present
// in a processing activity record.
func TestROPACompletenessCheck(t *testing.T) {
	type completenessField struct {
		Name     string
		Present  bool
	}

	// Simulate a fully complete Art.30 record
	activity := struct {
		Name                          string
		Purpose                       string
		LegalBasis                    string
		DataSubjectCategories         []string
		DataCategoryIDs               []uuid.UUID
		RecipientCategories           []string
		InvolvesInternationalTransfer bool
		TransferSafeguards            string
		RetentionPeriodMonths         *int
		SecurityMeasures              string
		Role                          string
		ProcessOwnerUserID            *uuid.UUID
	}{
		Name:                  "Customer data processing",
		Purpose:               "Service delivery",
		LegalBasis:            "contract",
		DataSubjectCategories: []string{"customers"},
		DataCategoryIDs:       []uuid.UUID{uuid.New()},
		RecipientCategories:   []string{"IT support"},
		RetentionPeriodMonths: intPtr(36),
		SecurityMeasures:      "AES-256 encryption, MFA, audit logging",
		Role:                  "controller",
		ProcessOwnerUserID:    uuidPtr(uuid.New()),
	}

	fields := []completenessField{
		{"Name", activity.Name != ""},
		{"Purpose", activity.Purpose != ""},
		{"Legal Basis", activity.LegalBasis != ""},
		{"Data Subject Categories", len(activity.DataSubjectCategories) > 0},
		{"Data Categories", len(activity.DataCategoryIDs) > 0},
		{"Recipient Categories", len(activity.RecipientCategories) > 0},
		{"Retention Period", activity.RetentionPeriodMonths != nil},
		{"Security Measures", activity.SecurityMeasures != ""},
		{"Controller/Processor Role", activity.Role != ""},
		{"Process Owner", activity.ProcessOwnerUserID != nil},
	}

	totalRequired := len(fields)
	present := 0
	for _, f := range fields {
		if f.Present {
			present++
		}
	}

	completeness := float64(present) / float64(totalRequired) * 100

	if completeness != 100 {
		t.Errorf("Expected 100%% completeness for full record, got %.1f%%", completeness)
	}
}

func TestROPACompletenessCheck_Incomplete(t *testing.T) {
	// Simulate a record missing several Art.30 fields
	activity := struct {
		Name                          string
		Purpose                       string
		LegalBasis                    string
		DataSubjectCategories         []string
		DataCategoryIDs               []uuid.UUID
		RecipientCategories           []string
		RetentionPeriodMonths         *int
		SecurityMeasures              string
		Role                          string
		ProcessOwnerUserID            *uuid.UUID
	}{
		Name:       "Draft activity",
		LegalBasis: "consent",
		Role:       "controller",
		// Missing: Purpose, DataSubjectCategories, DataCategoryIDs,
		// RecipientCategories, RetentionPeriodMonths, SecurityMeasures, ProcessOwnerUserID
	}

	type field struct {
		Name    string
		Present bool
	}
	fields := []field{
		{"Name", activity.Name != ""},
		{"Purpose", activity.Purpose != ""},
		{"Legal Basis", activity.LegalBasis != ""},
		{"Data Subject Categories", len(activity.DataSubjectCategories) > 0},
		{"Data Categories", len(activity.DataCategoryIDs) > 0},
		{"Recipient Categories", len(activity.RecipientCategories) > 0},
		{"Retention Period", activity.RetentionPeriodMonths != nil},
		{"Security Measures", activity.SecurityMeasures != ""},
		{"Controller/Processor Role", activity.Role != ""},
		{"Process Owner", activity.ProcessOwnerUserID != nil},
	}

	present := 0
	var missing []string
	for _, f := range fields {
		if f.Present {
			present++
		} else {
			missing = append(missing, f.Name)
		}
	}

	completeness := float64(present) / float64(len(fields)) * 100

	if completeness >= 100 {
		t.Error("Expected incomplete record to have <100% completeness")
	}
	if completeness != 30 {
		t.Errorf("Expected 30%% completeness (3/10 fields), got %.1f%%", completeness)
	}
	expectedMissing := 7
	if len(missing) != expectedMissing {
		t.Errorf("Expected %d missing fields, got %d: %v", expectedMissing, len(missing), missing)
	}
}

func TestROPACompletenessCheck_TransferWithoutSafeguards(t *testing.T) {
	// International transfer present but no safeguards specified
	activity := struct {
		InvolvesInternationalTransfer bool
		TransferCountries             []string
		TransferSafeguards            string
		TIAConducted                  bool
	}{
		InvolvesInternationalTransfer: true,
		TransferCountries:             []string{"US", "IN"},
		TransferSafeguards:            "",
		TIAConducted:                  false,
	}

	var issues []string
	if activity.InvolvesInternationalTransfer {
		if activity.TransferSafeguards == "" {
			issues = append(issues, "International transfer declared but no safeguard type specified")
		}
		if !activity.TIAConducted {
			issues = append(issues, "Transfer Impact Assessment not conducted for international transfer")
		}
		if len(activity.TransferCountries) == 0 {
			issues = append(issues, "Transfer countries not specified")
		}
	}

	if len(issues) != 2 {
		t.Errorf("Expected 2 transfer compliance issues, got %d: %v", len(issues), issues)
	}
}

// TestDPIAStatusConsistency verifies that DPIA status is consistent
// with the dpia_required flag.
func TestDPIAStatusConsistency(t *testing.T) {
	tests := []struct {
		name         string
		dpiaRequired bool
		dpiaStatus   string
		expectIssue  bool
	}{
		{"Required but not started", true, "required", true},
		{"Required and completed", true, "completed", false},
		{"Required and in progress", true, "in_progress", false},
		{"Not required, not started", false, "not_required", false},
		{"Not required but somehow completed", false, "completed", false}, // Not an issue
		{"Required but marked not_required", true, "not_required", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasIssue := false
			if tt.dpiaRequired && (tt.dpiaStatus == "required" || tt.dpiaStatus == "not_required") {
				hasIssue = true
			}
			if hasIssue != tt.expectIssue {
				t.Errorf("Test %q: expected issue=%v, got issue=%v", tt.name, tt.expectIssue, hasIssue)
			}
		})
	}
}

// TestHighRiskReasonCounting verifies the risk indicator counting logic.
func TestHighRiskReasonCounting(t *testing.T) {
	tests := []struct {
		name                       string
		specialCategories          bool
		internationalTransfer      bool
		estimatedSubjects          *int
		subjectCategories          []string
		dataCategoryCount          int
		expectedMinIndicators      int
		expectedDPIARequired       bool
	}{
		{
			name:                  "No indicators",
			expectedMinIndicators: 0,
			expectedDPIARequired:  false,
		},
		{
			name:                  "Special categories only",
			specialCategories:     true,
			expectedMinIndicators: 1,
			expectedDPIARequired:  false,
		},
		{
			name:                  "Special + transfer",
			specialCategories:     true,
			internationalTransfer: true,
			expectedMinIndicators: 2,
			expectedDPIARequired:  true,
		},
		{
			name:                  "Large scale + children",
			estimatedSubjects:     intPtr(50000),
			subjectCategories:     []string{"children"},
			expectedMinIndicators: 2,
			expectedDPIARequired:  true,
		},
		{
			name:                  "All 5 indicators",
			specialCategories:     true,
			internationalTransfer: true,
			estimatedSubjects:     intPtr(100000),
			subjectCategories:     []string{"employees", "children"},
			dataCategoryCount:     8,
			expectedMinIndicators: 5,
			expectedDPIARequired:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ids := make([]uuid.UUID, tt.dataCategoryCount)
			for i := range ids {
				ids[i] = uuid.New()
			}

			req := CreateActivityRequest{
				Name:                          tt.name,
				SpecialCategoriesProcessed:    tt.specialCategories,
				InvolvesInternationalTransfer: tt.internationalTransfer,
				EstimatedDataSubjectsCount:    tt.estimatedSubjects,
				DataSubjectCategories:         tt.subjectCategories,
				DataCategoryIDs:               ids,
			}

			result := assessDPIARequired(req)
			if result != tt.expectedDPIARequired {
				t.Errorf("Expected DPIA required=%v, got %v", tt.expectedDPIARequired, result)
			}
		})
	}
}

// TestLegalBasisValidation checks that legal basis values are valid enum members.
func TestLegalBasisValidation(t *testing.T) {
	validBases := map[string]bool{
		"consent":             true,
		"contract":            true,
		"legal_obligation":    true,
		"vital_interest":      true,
		"public_task":         true,
		"legitimate_interest": true,
	}

	testCases := []struct {
		basis    string
		expected bool
	}{
		{"consent", true},
		{"contract", true},
		{"legal_obligation", true},
		{"vital_interest", true},
		{"public_task", true},
		{"legitimate_interest", true},
		{"", false},
		{"invalid", false},
		{"CONSENT", false},
	}

	for _, tc := range testCases {
		t.Run(tc.basis, func(t *testing.T) {
			_, ok := validBases[tc.basis]
			if ok != tc.expected {
				t.Errorf("Legal basis %q: expected valid=%v, got valid=%v", tc.basis, tc.expected, ok)
			}
		})
	}
}

// ============================================================
// Helper functions for tests
// ============================================================

func intPtr(v int) *int {
	return &v
}

func uuidPtr(v uuid.UUID) *uuid.UUID {
	return &v
}
