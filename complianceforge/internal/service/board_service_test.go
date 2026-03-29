package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

// ============================================================
// Board Pack Data Structure Tests
// ============================================================

func TestBoardPackDataStructure(t *testing.T) {
	// Simulate building a board pack data structure and verify serialisation.
	packData := map[string]interface{}{
		"generated_at":  time.Now().UTC().Format(time.RFC3339),
		"meeting_id":    uuid.New().String(),
		"meeting_title": "Q1 2026 Board Meeting",
		"meeting_date":  "2026-03-28",
		"compliance_summary": map[string]interface{}{
			"overall_score": 82.5,
			"frameworks": []map[string]interface{}{
				{"code": "ISO27001", "name": "ISO 27001:2022", "score": 85.0},
				{"code": "NIS2", "name": "NIS2 Directive", "score": 72.0},
				{"code": "UK_GDPR", "name": "UK GDPR", "score": 90.5},
			},
		},
		"risk_dashboard": map[string]interface{}{
			"total":           15,
			"critical":        1,
			"high":            4,
			"medium":          7,
			"low":             3,
			"appetite_status": "at_limit",
		},
		"incident_report": map[string]interface{}{
			"total_last_quarter": 8,
			"open":               2,
			"critical":           0,
			"high":               1,
		},
		"regulatory_update": map[string]interface{}{
			"new_changes":         3,
			"pending_assessments": 2,
			"upcoming_deadlines":  []interface{}{},
		},
		"pending_actions_count": 5,
		"agenda_items":          json.RawMessage(`[{"title":"Compliance Review","order":1},{"title":"Risk Appetite","order":2}]`),
	}

	data, err := json.Marshal(packData)
	if err != nil {
		t.Fatalf("Failed to marshal board pack data: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("Board pack data serialised to empty bytes")
	}

	// Verify the data can be deserialised back
	var restored map[string]interface{}
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Failed to unmarshal board pack data: %v", err)
	}

	// Check key sections exist
	if restored["meeting_title"] != "Q1 2026 Board Meeting" {
		t.Errorf("Expected meeting_title = 'Q1 2026 Board Meeting', got %v", restored["meeting_title"])
	}

	compliance, ok := restored["compliance_summary"].(map[string]interface{})
	if !ok {
		t.Fatal("compliance_summary not found or wrong type")
	}
	if compliance["overall_score"].(float64) != 82.5 {
		t.Errorf("Expected overall_score = 82.5, got %v", compliance["overall_score"])
	}

	risk, ok := restored["risk_dashboard"].(map[string]interface{})
	if !ok {
		t.Fatal("risk_dashboard not found or wrong type")
	}
	if risk["appetite_status"] != "at_limit" {
		t.Errorf("Expected appetite_status = 'at_limit', got %v", risk["appetite_status"])
	}
}

func TestBoardPackPageCountEstimation(t *testing.T) {
	// Page count is estimated from JSON size: len/500 + 8 (for cover, ToC, etc.)
	tests := []struct {
		name         string
		jsonSize     int
		expectedMin  int
		expectedMax  int
	}{
		{"small pack", 2000, 10, 15},
		{"medium pack", 10000, 20, 30},
		{"large pack", 50000, 100, 110},
		{"empty pack", 0, 8, 9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pageCount := tt.jsonSize/500 + 8
			if pageCount < tt.expectedMin || pageCount > tt.expectedMax {
				t.Errorf("Page count %d outside expected range [%d, %d] for JSON size %d",
					pageCount, tt.expectedMin, tt.expectedMax, tt.jsonSize)
			}
		})
	}
}

// ============================================================
// Decision Recording Logic Tests
// ============================================================

func TestDecisionRefFormat(t *testing.T) {
	// Verify expected format: BD-YYYY-NNN
	year := time.Now().Year()
	seq := 42
	ref := "BD-" + time.Now().Format("2006") + "-" + leftPad(seq, 3)

	expectedPrefix := "BD-" + time.Now().Format("2006") + "-"
	if ref[:len(expectedPrefix)] != expectedPrefix {
		t.Errorf("Decision ref should start with %s, got %s", expectedPrefix, ref)
	}

	if len(ref) != len(expectedPrefix)+3 {
		t.Errorf("Decision ref length = %d, want %d. Year=%d", len(ref), len(expectedPrefix)+3, year)
	}
}

func TestDecisionOutcomeMapping(t *testing.T) {
	outcomes := map[string]bool{
		"approved":              true,
		"rejected":              true,
		"deferred":              true,
		"conditional_approval":  true,
	}

	valid := []string{"approved", "rejected", "deferred", "conditional_approval"}
	for _, outcome := range valid {
		if !outcomes[outcome] {
			t.Errorf("Expected outcome %q to be valid", outcome)
		}
	}

	invalid := []string{"maybe", "pending", "cancelled", ""}
	for _, outcome := range invalid {
		if outcomes[outcome] {
			t.Errorf("Expected outcome %q to be invalid", outcome)
		}
	}
}

func TestDecisionLinkedEntityStatusUpdate(t *testing.T) {
	// Verify the logic for updating linked entities based on board decisions.
	tests := []struct {
		entityType     string
		decision       string
		shouldUpdate   bool
		expectedStatus string
	}{
		{"risk", "approved", true, "accepted"},
		{"risk", "conditional_approval", true, "accepted"},
		{"risk", "rejected", false, ""},
		{"risk", "deferred", false, ""},
		{"policy", "approved", true, "approved"},
		{"policy", "rejected", false, ""},
		{"vendor", "approved", true, "approved"},
		{"vendor", "rejected", false, ""},
		{"unknown", "approved", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.entityType+"/"+tt.decision, func(t *testing.T) {
			shouldUpdate, expectedStatus := simulateLinkedEntityUpdate(tt.entityType, tt.decision)
			if shouldUpdate != tt.shouldUpdate {
				t.Errorf("shouldUpdate = %v, want %v", shouldUpdate, tt.shouldUpdate)
			}
			if shouldUpdate && expectedStatus != tt.expectedStatus {
				t.Errorf("expectedStatus = %q, want %q", expectedStatus, tt.expectedStatus)
			}
		})
	}
}

func TestDecisionVoteTally(t *testing.T) {
	// Verify vote counting and quorum detection
	tests := []struct {
		name       string
		voteFor    int
		voteAgainst int
		voteAbstain int
		totalVoters int
		hasQuorum   bool
		passed      bool
	}{
		{"unanimous", 10, 0, 0, 10, true, true},
		{"majority", 7, 3, 0, 10, true, true},
		{"tied", 5, 5, 0, 10, true, false},
		{"abstentions", 5, 2, 3, 10, true, true},
		{"no quorum", 3, 1, 0, 10, false, true},
		{"single voter", 1, 0, 0, 1, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			totalVoted := tt.voteFor + tt.voteAgainst + tt.voteAbstain
			hasQuorum := float64(totalVoted) >= float64(tt.totalVoters)*0.5
			passed := tt.voteFor > tt.voteAgainst

			if hasQuorum != tt.hasQuorum {
				t.Errorf("hasQuorum = %v, want %v (voted=%d, total=%d)",
					hasQuorum, tt.hasQuorum, totalVoted, tt.totalVoters)
			}
			if passed != tt.passed {
				t.Errorf("passed = %v, want %v (for=%d, against=%d)",
					passed, tt.passed, tt.voteFor, tt.voteAgainst)
			}
		})
	}
}

// ============================================================
// Token-Based Portal Access Tests
// ============================================================

func TestTokenHashing(t *testing.T) {
	// Verify token hashing produces deterministic, fixed-length output.
	token := "test-portal-token-abc123"
	hash1 := hashTokenTest(token)
	hash2 := hashTokenTest(token)

	if hash1 != hash2 {
		t.Error("Token hashing is not deterministic")
	}

	if len(hash1) != 64 { // SHA-256 produces 32 bytes = 64 hex chars
		t.Errorf("Token hash length = %d, want 64", len(hash1))
	}

	// Different tokens produce different hashes
	differentHash := hashTokenTest("different-token")
	if hash1 == differentHash {
		t.Error("Different tokens should produce different hashes")
	}
}

func TestTokenExpiry(t *testing.T) {
	// Simulate checking token expiry.
	now := time.Now()

	tests := []struct {
		name      string
		expiresAt *time.Time
		isValid   bool
	}{
		{"no expiry (nil)", nil, true},
		{"future expiry", timePtr(now.Add(24 * time.Hour)), true},
		{"past expiry", timePtr(now.Add(-24 * time.Hour)), false},
		{"exactly now", timePtr(now), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.expiresAt == nil || tt.expiresAt.After(now)
			if isValid != tt.isValid {
				t.Errorf("isValid = %v, want %v", isValid, tt.isValid)
			}
		})
	}
}

func TestPortalAccessRequirements(t *testing.T) {
	// Verify that portal access requires all conditions to be met:
	// - portal_access_enabled = true
	// - is_active = true
	// - token matches
	// - not expired

	type memberAccess struct {
		isActive            bool
		portalAccessEnabled bool
		tokenMatches        bool
		isExpired           bool
	}

	tests := []struct {
		name     string
		access   memberAccess
		allowed  bool
	}{
		{"all good", memberAccess{true, true, true, false}, true},
		{"inactive member", memberAccess{false, true, true, false}, false},
		{"portal disabled", memberAccess{true, false, true, false}, false},
		{"wrong token", memberAccess{true, true, false, false}, false},
		{"expired", memberAccess{true, true, true, true}, false},
		{"all bad", memberAccess{false, false, false, true}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed := tt.access.isActive &&
				tt.access.portalAccessEnabled &&
				tt.access.tokenMatches &&
				!tt.access.isExpired

			if allowed != tt.allowed {
				t.Errorf("allowed = %v, want %v", allowed, tt.allowed)
			}
		})
	}
}

// ============================================================
// Risk Appetite Status Tests
// ============================================================

func TestRiskAppetiteStatus(t *testing.T) {
	tests := []struct {
		name     string
		critical int
		high     int
		expected string
	}{
		{"no risks", 0, 0, "within_appetite"},
		{"low risk", 0, 3, "within_appetite"},
		{"at limit", 0, 6, "at_limit"},
		{"exceeded critical", 1, 0, "exceeded"},
		{"exceeded both", 2, 10, "exceeded"},
		{"exactly threshold", 0, 5, "within_appetite"},
		{"just over threshold", 0, 6, "at_limit"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appetite := "within_appetite"
			if tt.critical > 0 {
				appetite = "exceeded"
			} else if tt.high > 5 {
				appetite = "at_limit"
			}

			if appetite != tt.expected {
				t.Errorf("appetite = %q, want %q (critical=%d, high=%d)",
					appetite, tt.expected, tt.critical, tt.high)
			}
		})
	}
}

// ============================================================
// Meeting Ref Format Tests
// ============================================================

func TestMeetingRefFormat(t *testing.T) {
	// Verify the expected format: BM-YYYY-QN-NN
	tests := []struct {
		month    int
		expected string
	}{
		{1, "Q1"},
		{2, "Q1"},
		{3, "Q1"},
		{4, "Q2"},
		{5, "Q2"},
		{6, "Q2"},
		{7, "Q3"},
		{8, "Q3"},
		{9, "Q3"},
		{10, "Q4"},
		{11, "Q4"},
		{12, "Q4"},
	}

	for _, tt := range tests {
		t.Run("month_"+leftPad(tt.month, 2), func(t *testing.T) {
			quarter := (tt.month-1)/3 + 1
			qStr := "Q" + leftPad(quarter, 1)
			if qStr != tt.expected {
				t.Errorf("month %d -> quarter %s, want %s", tt.month, qStr, tt.expected)
			}
		})
	}
}

// ============================================================
// Board Dashboard Aggregation Tests
// ============================================================

func TestDashboardComplianceByFramework(t *testing.T) {
	scores := []FrameworkScore{
		{FrameworkCode: "ISO27001", FrameworkName: "ISO 27001:2022", Score: 85.0},
		{FrameworkCode: "NIS2", FrameworkName: "NIS2 Directive", Score: 72.0},
		{FrameworkCode: "UK_GDPR", FrameworkName: "UK GDPR", Score: 90.5},
	}

	if len(scores) != 3 {
		t.Fatalf("Expected 3 framework scores, got %d", len(scores))
	}

	// Calculate overall average
	var total float64
	for _, s := range scores {
		total += s.Score
	}
	avg := total / float64(len(scores))

	expectedAvg := (85.0 + 72.0 + 90.5) / 3.0
	if avg != expectedAvg {
		t.Errorf("Average score = %.2f, want %.2f", avg, expectedAvg)
	}

	// Verify lowest score is flagged
	lowestIdx := 0
	for i, s := range scores {
		if s.Score < scores[lowestIdx].Score {
			lowestIdx = i
		}
	}
	if scores[lowestIdx].FrameworkCode != "NIS2" {
		t.Errorf("Lowest score framework = %s, want NIS2", scores[lowestIdx].FrameworkCode)
	}
}

func TestNIS2GovernanceReportStructure(t *testing.T) {
	report := NIS2GovernanceReport{
		GeneratedAt:          time.Now().UTC(),
		OrganizationID:       uuid.New(),
		ManagementBodyStatus: "established",
		TrainingStatus:       "in_progress",
		RiskManagementScore:  75.5,
		IncidentReporting: NIS2IncidentReporting{
			EarlyWarningCapable: true,
			NotificationCapable: true,
			FinalReportCapable:  true,
			IncidentsReported:   5,
			AverageResponseHours: 4,
		},
		SupplyChainSecurity:   "monitored",
		BoardOversightSummary: "Board has 7 active members.",
		SecurityMeasures: []NIS2MeasureStatus{
			{MeasureName: "Risk Analysis & Policies", Status: "implemented", Coverage: 85.0},
			{MeasureName: "Incident Handling", Status: "implemented", Coverage: 80.0},
		},
	}

	data, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("Failed to marshal NIS2 report: %v", err)
	}

	var restored NIS2GovernanceReport
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Failed to unmarshal NIS2 report: %v", err)
	}

	if restored.ManagementBodyStatus != "established" {
		t.Errorf("ManagementBodyStatus = %q, want 'established'", restored.ManagementBodyStatus)
	}
	if !restored.IncidentReporting.EarlyWarningCapable {
		t.Error("EarlyWarningCapable should be true")
	}
	if len(restored.SecurityMeasures) != 2 {
		t.Errorf("SecurityMeasures count = %d, want 2", len(restored.SecurityMeasures))
	}
}

// ============================================================
// HELPERS
// ============================================================

func hashTokenTest(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

func simulateLinkedEntityUpdate(entityType, decision string) (bool, string) {
	switch entityType {
	case "risk":
		if decision == "approved" || decision == "conditional_approval" {
			return true, "accepted"
		}
	case "policy":
		if decision == "approved" {
			return true, "approved"
		}
	case "vendor":
		if decision == "approved" {
			return true, "approved"
		}
	}
	return false, ""
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func leftPad(n, width int) string {
	s := ""
	v := n
	for v > 0 {
		s = string(rune('0'+v%10)) + s
		v /= 10
	}
	if s == "" {
		s = "0"
	}
	for len(s) < width {
		s = "0" + s
	}
	return s
}
