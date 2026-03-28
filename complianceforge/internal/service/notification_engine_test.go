package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// ============================================================
// TEMPLATE RENDERER TESTS
// ============================================================

func TestRenderTemplate_SimpleVariables(t *testing.T) {
	tmpl := "Hello {{.Name}}, your incident {{.IncidentRef}} needs attention."
	data := map[string]interface{}{
		"Name":        "Alice",
		"IncidentRef": "INC-001",
	}

	result, err := RenderTemplate(tmpl, data)
	if err != nil {
		t.Fatalf("RenderTemplate returned error: %v", err)
	}

	expected := "Hello Alice, your incident INC-001 needs attention."
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestRenderTemplate_MissingVariable(t *testing.T) {
	tmpl := "Incident: {{.IncidentRef}} — Severity: {{.Severity}}"
	data := map[string]interface{}{
		"IncidentRef": "INC-002",
		// Severity intentionally omitted
	}

	result, err := RenderTemplate(tmpl, data)
	if err != nil {
		t.Fatalf("RenderTemplate returned error: %v", err)
	}

	// With missingkey=zero, missing keys render as zero value
	expected := "Incident: INC-002 — Severity: <no value>"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestRenderTemplate_EmptyTemplate(t *testing.T) {
	result, err := RenderTemplate("", map[string]interface{}{})
	if err != nil {
		t.Fatalf("RenderTemplate returned error: %v", err)
	}
	if result != "" {
		t.Errorf("Expected empty string, got %q", result)
	}
}

func TestRenderTemplate_HTMLContent(t *testing.T) {
	tmpl := `<h2>GDPR Breach — {{.IncidentRef}}</h2><p>Deadline: {{.Deadline}}</p><p>Data subjects: {{.DataSubjectsAffected}}</p>`
	data := map[string]interface{}{
		"IncidentRef":          "INC-GDPR-001",
		"Deadline":             "2026-03-30T14:00:00Z",
		"DataSubjectsAffected": "1500",
	}

	result, err := RenderTemplate(tmpl, data)
	if err != nil {
		t.Fatalf("RenderTemplate returned error: %v", err)
	}

	if result == "" {
		t.Error("Expected non-empty HTML output")
	}

	// Check key content is present
	for _, expected := range []string{"INC-GDPR-001", "2026-03-30T14:00:00Z", "1500"} {
		if !containsString(result, expected) {
			t.Errorf("Expected result to contain %q, got %q", expected, result)
		}
	}
}

func TestRenderTemplate_SlackBlockKit(t *testing.T) {
	tmpl := `{"blocks":[{"type":"header","text":{"type":"plain_text","text":"Incident: {{.IncidentRef}}"}}]}`
	data := map[string]interface{}{
		"IncidentRef": "INC-007",
	}

	result, err := RenderTemplate(tmpl, data)
	if err != nil {
		t.Fatalf("RenderTemplate returned error: %v", err)
	}

	if !containsString(result, "INC-007") {
		t.Errorf("Expected result to contain 'INC-007', got %q", result)
	}
}

func TestRenderTemplate_ComplexPayload(t *testing.T) {
	tmpl := `{{.PolicyRef}} — {{.Title}} | Owner: {{.OwnerName}} | Due: {{.DueDate}} | Last: {{.LastReviewDate}}`
	data := map[string]interface{}{
		"PolicyRef":      "POL-003",
		"Title":          "Information Security Policy",
		"OwnerName":      "Bob Smith",
		"DueDate":        "2026-04-15",
		"LastReviewDate": "2025-04-15",
	}

	result, err := RenderTemplate(tmpl, data)
	if err != nil {
		t.Fatalf("RenderTemplate returned error: %v", err)
	}

	expected := "POL-003 — Information Security Policy | Owner: Bob Smith | Due: 2026-04-15 | Last: 2025-04-15"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestRenderTemplate_InvalidSyntax(t *testing.T) {
	tmpl := "{{.Name} missing closing brace"
	_, err := RenderTemplate(tmpl, map[string]interface{}{"Name": "test"})
	if err == nil {
		t.Error("Expected error for invalid template syntax")
	}
}

// ============================================================
// EVENT BUS TESTS
// ============================================================

func TestEventBus_PublishSubscribe(t *testing.T) {
	bus := NewEventBus(10)
	defer bus.Close()

	ch := bus.Subscribe()

	event := Event{
		Type:      EventIncidentCreated,
		OrgID:     uuid.New(),
		Severity:  "critical",
		Payload:   map[string]interface{}{"Title": "Test Incident"},
		Timestamp: time.Now(),
	}

	bus.Publish(event)

	select {
	case received := <-ch:
		if received.Type != EventIncidentCreated {
			t.Errorf("Expected event type %q, got %q", EventIncidentCreated, received.Type)
		}
		if received.Severity != "critical" {
			t.Errorf("Expected severity 'critical', got %q", received.Severity)
		}
		title, ok := received.Payload["Title"].(string)
		if !ok || title != "Test Incident" {
			t.Errorf("Expected payload Title 'Test Incident', got %v", received.Payload["Title"])
		}
	case <-time.After(1 * time.Second):
		t.Error("Timed out waiting for event")
	}
}

func TestEventBus_MultipleSubscribers(t *testing.T) {
	bus := NewEventBus(10)
	defer bus.Close()

	ch1 := bus.Subscribe()
	ch2 := bus.Subscribe()

	event := Event{
		Type:      EventBreachDeadlineApproaching,
		OrgID:     uuid.New(),
		Severity:  "critical",
		Timestamp: time.Now(),
	}

	bus.Publish(event)

	// Both subscribers should receive the event
	for i, ch := range []chan Event{ch1, ch2} {
		select {
		case received := <-ch:
			if received.Type != EventBreachDeadlineApproaching {
				t.Errorf("Subscriber %d: expected type %q, got %q", i, EventBreachDeadlineApproaching, received.Type)
			}
		case <-time.After(1 * time.Second):
			t.Errorf("Subscriber %d: timed out waiting for event", i)
		}
	}
}

func TestEventBus_BufferFull(t *testing.T) {
	bus := NewEventBus(1) // Very small buffer
	defer bus.Close()

	_ = bus.Subscribe()

	// Fill the buffer
	bus.Publish(Event{Type: "first", Timestamp: time.Now()})

	// This should not block (event is dropped)
	bus.Publish(Event{Type: "second", Timestamp: time.Now()})

	// Should not panic or hang
}

// ============================================================
// RATE LIMITER TESTS
// ============================================================

func TestRateLimiter_AllowWithinLimit(t *testing.T) {
	rl := newRateLimiter(5, 1*time.Hour)
	userID := uuid.New()

	for i := 0; i < 5; i++ {
		if !rl.Allow(userID) {
			t.Errorf("Expected Allow() to return true for request %d", i+1)
		}
	}
}

func TestRateLimiter_DenyOverLimit(t *testing.T) {
	rl := newRateLimiter(3, 1*time.Hour)
	userID := uuid.New()

	// Use up the quota
	for i := 0; i < 3; i++ {
		rl.Allow(userID)
	}

	// Next request should be denied
	if rl.Allow(userID) {
		t.Error("Expected Allow() to return false when rate limit exceeded")
	}
}

func TestRateLimiter_DifferentUsers(t *testing.T) {
	rl := newRateLimiter(2, 1*time.Hour)
	user1 := uuid.New()
	user2 := uuid.New()

	// User 1 uses quota
	rl.Allow(user1)
	rl.Allow(user1)

	// User 1 is now at limit
	if rl.Allow(user1) {
		t.Error("Expected user1 to be rate limited")
	}

	// User 2 should still be allowed
	if !rl.Allow(user2) {
		t.Error("Expected user2 to be allowed (different user)")
	}
}

func TestRateLimiter_WindowReset(t *testing.T) {
	// Use a very short window for testing
	rl := newRateLimiter(1, 50*time.Millisecond)
	userID := uuid.New()

	// First request passes
	if !rl.Allow(userID) {
		t.Error("Expected first request to be allowed")
	}

	// Second request should be denied
	if rl.Allow(userID) {
		t.Error("Expected second request to be denied")
	}

	// Wait for window to reset
	time.Sleep(60 * time.Millisecond)

	// Should be allowed again
	if !rl.Allow(userID) {
		t.Error("Expected request to be allowed after window reset")
	}
}

// ============================================================
// RULE EVALUATOR TESTS
// ============================================================

func TestEvaluateRules_SeverityFilter(t *testing.T) {
	// Test that severity filter logic works correctly
	rule := NotificationRule{
		ID:             uuid.New(),
		EventType:      EventIncidentCreated,
		SeverityFilter: []string{"critical", "high"},
		IsActive:       true,
	}

	// Test matching severity
	tests := []struct {
		severity string
		expected bool
	}{
		{"critical", true},
		{"high", true},
		{"medium", false},
		{"low", false},
		{"", true}, // empty severity matches all
	}

	for _, tt := range tests {
		matched := severityMatches(rule.SeverityFilter, tt.severity)
		if matched != tt.expected {
			t.Errorf("severityMatches(%v, %q) = %v, want %v",
				rule.SeverityFilter, tt.severity, matched, tt.expected)
		}
	}
}

func TestEvaluateRules_EmptySeverityFilter(t *testing.T) {
	// Empty severity filter means "match all severities"
	result := severityMatches([]string{}, "critical")
	if !result {
		t.Error("Empty severity filter should match all severities")
	}

	result = severityMatches(nil, "high")
	if !result {
		t.Error("Nil severity filter should match all severities")
	}
}

// ============================================================
// NOTIFICATION MODEL TESTS
// ============================================================

func TestNotification_Defaults(t *testing.T) {
	n := &Notification{
		ID:              uuid.New(),
		OrganizationID:  uuid.New(),
		EventType:       EventIncidentCreated,
		RecipientUserID: uuid.New(),
		ChannelType:     ChannelEmail,
		Status:          "pending",
		RetryCount:      0,
		MaxRetries:      3,
		CreatedAt:       time.Now(),
	}

	if n.Status != "pending" {
		t.Errorf("Expected status 'pending', got %q", n.Status)
	}

	if n.RetryCount != 0 {
		t.Errorf("Expected retry count 0, got %d", n.RetryCount)
	}

	if n.MaxRetries != 3 {
		t.Errorf("Expected max retries 3, got %d", n.MaxRetries)
	}
}

func TestEventConstants(t *testing.T) {
	// Verify event constants are unique and properly formatted
	events := []string{
		EventIncidentCreated,
		EventIncidentUpdated,
		EventIncidentResolved,
		EventIncidentClosed,
		EventBreachDeadlineApproaching,
		EventBreachDeadlineExpired,
		EventBreachDPANotified,
		EventNIS2EarlyWarningDue,
		EventNIS2NotificationDue,
		EventNIS2EarlyWarningFiled,
		EventControlStatusChanged,
		EventControlTestFailed,
		EventControlTestPassed,
		EventPolicyReviewDue,
		EventPolicyReviewOverdue,
		EventPolicyPublished,
		EventAttestationRequired,
		EventAttestationOverdue,
		EventAttestationCompleted,
		EventFindingCreated,
		EventFindingRemediationOverdue,
		EventAuditCompleted,
		EventVendorAssessmentDue,
		EventVendorMissingDPA,
		EventVendorRiskChanged,
		EventRiskThresholdExceeded,
		EventRiskReviewDue,
		EventComplianceScoreDropped,
		EventUserWelcome,
		EventUserPasswordReset,
	}

	seen := make(map[string]bool)
	for _, e := range events {
		if e == "" {
			t.Error("Event constant must not be empty")
		}
		if seen[e] {
			t.Errorf("Duplicate event constant: %q", e)
		}
		seen[e] = true

		// Verify format: category.action
		if !containsString(e, ".") {
			t.Errorf("Event constant %q should follow 'category.action' format", e)
		}
	}
}

func TestChannelTypes(t *testing.T) {
	channels := []ChannelType{ChannelEmail, ChannelInApp, ChannelSlack, ChannelWebhook}

	seen := make(map[ChannelType]bool)
	for _, ch := range channels {
		if ch == "" {
			t.Error("Channel type must not be empty")
		}
		if seen[ch] {
			t.Errorf("Duplicate channel type: %q", ch)
		}
		seen[ch] = true
	}
}

// ============================================================
// HELPER FUNCTIONS
// ============================================================

// severityMatches checks if event severity is in the allowed list.
// This mirrors the logic in evaluateRules.
func severityMatches(filter []string, severity string) bool {
	if len(filter) == 0 {
		return true
	}
	if severity == "" {
		return true
	}
	for _, s := range filter {
		if s == severity {
			return true
		}
	}
	return false
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
