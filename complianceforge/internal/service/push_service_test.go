package service

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

// ============================================================
// MOCK PUSH PROVIDER
// ============================================================

// mockPushProvider records push sends for testing.
type mockPushProvider struct {
	mu       sync.Mutex
	sends    []mockPushSend
	failNext bool
	failWith string
}

type mockPushSend struct {
	Platform     string
	Token        string
	Notification *PushNotification
}

func (m *mockPushProvider) Send(_ context.Context, platform, token string, notification *PushNotification) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.sends = append(m.sends, mockPushSend{
		Platform:     platform,
		Token:        token,
		Notification: notification,
	})

	if m.failNext {
		m.failNext = false
		if m.failWith != "" {
			return fmt.Errorf("%s", m.failWith)
		}
		return fmt.Errorf("mock send failure")
	}

	return nil
}

func (m *mockPushProvider) getSends() []mockPushSend {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]mockPushSend, len(m.sends))
	copy(result, m.sends)
	return result
}

func (m *mockPushProvider) reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sends = nil
	m.failNext = false
	m.failWith = ""
}

// ============================================================
// TOKEN HASH TESTS
// ============================================================

func TestHashToken_Deterministic(t *testing.T) {
	token := "test-push-token-abc123"
	hash1 := hashToken(token)
	hash2 := hashToken(token)

	if hash1 != hash2 {
		t.Errorf("hashToken should be deterministic: %s != %s", hash1, hash2)
	}

	if len(hash1) != 64 {
		t.Errorf("SHA-256 hex hash should be 64 chars, got %d", len(hash1))
	}
}

func TestHashToken_DifferentTokens(t *testing.T) {
	hash1 := hashToken("token-A")
	hash2 := hashToken("token-B")

	if hash1 == hash2 {
		t.Error("Different tokens should produce different hashes")
	}
}

func TestHashToken_EmptyToken(t *testing.T) {
	hash := hashToken("")
	if hash == "" {
		t.Error("hashToken of empty string should not be empty")
	}
	if len(hash) != 64 {
		t.Errorf("Expected 64-char hash, got %d", len(hash))
	}
}

func TestHashToken_Exported(t *testing.T) {
	token := "exported-token-test"
	if HashToken(token) != hashToken(token) {
		t.Error("HashToken (exported) should produce same result as hashToken (internal)")
	}
}

// ============================================================
// PREFERENCE ENFORCEMENT TESTS
// ============================================================

func TestIsTypeEnabled_BreachAlwaysOn(t *testing.T) {
	svc := &PushService{}

	// Even with everything disabled, breach alerts should be enabled
	prefs := &MobilePreferences{
		PushEnabled:           false,
		PushBreachAlerts:      false, // attempt to disable
		PushApprovalRequests:  false,
		PushIncidentAlerts:    false,
		PushDeadlineReminders: false,
		PushMentions:          false,
		PushComments:          false,
	}

	if !svc.isTypeEnabled(prefs, PushTypeBreachAlert) {
		t.Error("Breach alerts must ALWAYS be enabled regardless of preferences")
	}
}

func TestIsTypeEnabled_AllTypes(t *testing.T) {
	svc := &PushService{}

	tests := []struct {
		name      string
		notifType string
		prefs     *MobilePreferences
		expected  bool
	}{
		{
			name:      "approval_request enabled",
			notifType: PushTypeApprovalRequest,
			prefs:     &MobilePreferences{PushApprovalRequests: true},
			expected:  true,
		},
		{
			name:      "approval_request disabled",
			notifType: PushTypeApprovalRequest,
			prefs:     &MobilePreferences{PushApprovalRequests: false},
			expected:  false,
		},
		{
			name:      "incident_created enabled",
			notifType: PushTypeIncidentCreated,
			prefs:     &MobilePreferences{PushIncidentAlerts: true},
			expected:  true,
		},
		{
			name:      "incident_created disabled",
			notifType: PushTypeIncidentCreated,
			prefs:     &MobilePreferences{PushIncidentAlerts: false},
			expected:  false,
		},
		{
			name:      "deadline_reminder enabled",
			notifType: PushTypeDeadlineReminder,
			prefs:     &MobilePreferences{PushDeadlineReminders: true},
			expected:  true,
		},
		{
			name:      "deadline_reminder disabled",
			notifType: PushTypeDeadlineReminder,
			prefs:     &MobilePreferences{PushDeadlineReminders: false},
			expected:  false,
		},
		{
			name:      "mention enabled",
			notifType: PushTypeMention,
			prefs:     &MobilePreferences{PushMentions: true},
			expected:  true,
		},
		{
			name:      "mention disabled",
			notifType: PushTypeMention,
			prefs:     &MobilePreferences{PushMentions: false},
			expected:  false,
		},
		{
			name:      "comment enabled",
			notifType: PushTypeComment,
			prefs:     &MobilePreferences{PushComments: true},
			expected:  true,
		},
		{
			name:      "comment disabled",
			notifType: PushTypeComment,
			prefs:     &MobilePreferences{PushComments: false},
			expected:  false,
		},
		{
			name:      "breach_alert always enabled",
			notifType: PushTypeBreachAlert,
			prefs:     &MobilePreferences{PushBreachAlerts: false},
			expected:  true,
		},
		{
			name:      "unknown type uses global push_enabled true",
			notifType: "some_new_type",
			prefs:     &MobilePreferences{PushEnabled: true},
			expected:  true,
		},
		{
			name:      "unknown type uses global push_enabled false",
			notifType: "some_new_type",
			prefs:     &MobilePreferences{PushEnabled: false},
			expected:  false,
		},
		{
			name:      "nil prefs defaults to enabled",
			notifType: PushTypeIncidentCreated,
			prefs:     nil,
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.isTypeEnabled(tt.prefs, tt.notifType)
			if result != tt.expected {
				t.Errorf("isTypeEnabled(%q) = %v, want %v", tt.notifType, result, tt.expected)
			}
		})
	}
}

// ============================================================
// QUIET HOURS TESTS
// ============================================================

func TestIsTimeInRange_SameDayRange(t *testing.T) {
	tests := []struct {
		current  string
		start    string
		end      string
		expected bool
	}{
		{"10:00", "09:00", "17:00", true},  // within range
		{"08:00", "09:00", "17:00", false}, // before range
		{"17:00", "09:00", "17:00", false}, // at end (exclusive)
		{"09:00", "09:00", "17:00", true},  // at start (inclusive)
		{"16:59", "09:00", "17:00", true},  // just before end
	}

	for _, tt := range tests {
		result := isTimeInRange(tt.current, tt.start, tt.end)
		if result != tt.expected {
			t.Errorf("isTimeInRange(%q, %q, %q) = %v, want %v",
				tt.current, tt.start, tt.end, result, tt.expected)
		}
	}
}

func TestIsTimeInRange_OvernightRange(t *testing.T) {
	tests := []struct {
		current  string
		start    string
		end      string
		expected bool
	}{
		{"23:00", "22:00", "08:00", true},  // evening, within quiet hours
		{"02:00", "22:00", "08:00", true},  // early morning, within quiet hours
		{"07:59", "22:00", "08:00", true},  // just before end
		{"08:00", "22:00", "08:00", false}, // at end (exclusive)
		{"22:00", "22:00", "08:00", true},  // at start (inclusive)
		{"12:00", "22:00", "08:00", false}, // midday, outside quiet hours
		{"21:59", "22:00", "08:00", false}, // just before start
	}

	for _, tt := range tests {
		result := isTimeInRange(tt.current, tt.start, tt.end)
		if result != tt.expected {
			t.Errorf("isTimeInRange(%q, %q, %q) = %v, want %v",
				tt.current, tt.start, tt.end, result, tt.expected)
		}
	}
}

func TestIsInQuietHours_Disabled(t *testing.T) {
	svc := &PushService{}
	prefs := &MobilePreferences{
		QuietHoursEnabled:  false,
		QuietHoursStart:    "22:00",
		QuietHoursEnd:      "08:00",
		QuietHoursTimezone: "UTC",
	}

	// Should never be in quiet hours when disabled
	if svc.IsInQuietHours(prefs) {
		t.Error("Should not be in quiet hours when quiet_hours_enabled is false")
	}
}

func TestIsInQuietHours_NilPrefs(t *testing.T) {
	svc := &PushService{}
	if svc.IsInQuietHours(nil) {
		t.Error("Should not be in quiet hours with nil preferences")
	}
}

func TestIsInQuietHours_InvalidTimezone(t *testing.T) {
	svc := &PushService{}
	prefs := &MobilePreferences{
		QuietHoursEnabled:  true,
		QuietHoursStart:    "00:00",
		QuietHoursEnd:      "23:59",
		QuietHoursTimezone: "Invalid/Timezone",
	}

	// Should not panic, falls back to UTC
	_ = svc.IsInQuietHours(prefs)
}

// ============================================================
// INVALID TOKEN ERROR DETECTION
// ============================================================

func TestIsInvalidTokenError(t *testing.T) {
	tests := []struct {
		err      error
		expected bool
	}{
		{nil, false},
		{fmt.Errorf("network timeout"), false},
		{fmt.Errorf("server error 500"), false},
		{fmt.Errorf("Invalid registration token"), true},
		{fmt.Errorf("token is unregistered"), true},
		{fmt.Errorf("device Not Registered"), true},
		{fmt.Errorf("expired token for device"), true},
		{fmt.Errorf("INVALID_ARGUMENT: invalid token"), true},
	}

	for _, tt := range tests {
		result := isInvalidTokenError(tt.err)
		if result != tt.expected {
			errMsg := "<nil>"
			if tt.err != nil {
				errMsg = tt.err.Error()
			}
			t.Errorf("isInvalidTokenError(%q) = %v, want %v", errMsg, result, tt.expected)
		}
	}
}

// ============================================================
// NOTIFICATION TYPE CONSTANTS
// ============================================================

func TestPushTypeConstants(t *testing.T) {
	types := []string{
		PushTypeBreachAlert,
		PushTypeApprovalRequest,
		PushTypeIncidentCreated,
		PushTypeDeadlineReminder,
		PushTypeMention,
		PushTypeComment,
	}

	seen := make(map[string]bool)
	for _, pt := range types {
		if pt == "" {
			t.Error("Push type constant must not be empty")
		}
		if seen[pt] {
			t.Errorf("Duplicate push type constant: %q", pt)
		}
		seen[pt] = true
	}
}

// ============================================================
// MOBILE PREFERENCES DEFAULT VALUES
// ============================================================

func TestMobilePreferences_Defaults(t *testing.T) {
	prefs := MobilePreferences{
		UserID:                uuid.New(),
		OrganizationID:        uuid.New(),
		PushEnabled:           true,
		PushBreachAlerts:      true,
		PushApprovalRequests:  true,
		PushIncidentAlerts:    true,
		PushDeadlineReminders: true,
		PushMentions:          true,
		PushComments:          false,
		QuietHoursEnabled:     false,
		QuietHoursStart:       "22:00",
		QuietHoursEnd:         "08:00",
		QuietHoursTimezone:    "UTC",
	}

	if !prefs.PushEnabled {
		t.Error("Default push_enabled should be true")
	}
	if !prefs.PushBreachAlerts {
		t.Error("Default push_breach_alerts should be true")
	}
	if prefs.PushComments {
		t.Error("Default push_comments should be false")
	}
	if prefs.QuietHoursEnabled {
		t.Error("Default quiet_hours_enabled should be false")
	}
	if prefs.QuietHoursTimezone != "UTC" {
		t.Errorf("Default timezone should be 'UTC', got %q", prefs.QuietHoursTimezone)
	}
}

// ============================================================
// PUSH NOTIFICATION MODEL TESTS
// ============================================================

func TestPushNotification_Fields(t *testing.T) {
	n := &PushNotification{
		Type:     PushTypeBreachAlert,
		Title:    "Test Alert",
		Body:     "Test body",
		Priority: "high",
		Sound:    "critical_alert",
		Data: map[string]interface{}{
			"incident_id": uuid.New().String(),
		},
	}

	if n.Type != PushTypeBreachAlert {
		t.Errorf("Expected type %q, got %q", PushTypeBreachAlert, n.Type)
	}
	if n.Priority != "high" {
		t.Errorf("Expected priority 'high', got %q", n.Priority)
	}
	if n.Sound != "critical_alert" {
		t.Errorf("Expected sound 'critical_alert', got %q", n.Sound)
	}
	if n.Data["incident_id"] == nil {
		t.Error("Expected incident_id in data")
	}
}

// ============================================================
// DEVICE INFO & TOKEN MODEL TESTS
// ============================================================

func TestDeviceInfo_Fields(t *testing.T) {
	info := DeviceInfo{
		DeviceName:  "iPhone 15 Pro",
		DeviceModel: "iPhone16,1",
		OSVersion:   "iOS 17.4",
		AppVersion:  "2.5.0",
	}

	if info.DeviceName != "iPhone 15 Pro" {
		t.Errorf("Expected device name 'iPhone 15 Pro', got %q", info.DeviceName)
	}
	if info.AppVersion != "2.5.0" {
		t.Errorf("Expected app version '2.5.0', got %q", info.AppVersion)
	}
}

func TestPushToken_Fields(t *testing.T) {
	now := time.Now()
	token := PushToken{
		ID:             uuid.New(),
		UserID:         uuid.New(),
		OrganizationID: uuid.New(),
		Platform:       "ios",
		Token:          "abc123",
		TokenHash:      hashToken("abc123"),
		DeviceName:     "Test Device",
		IsActive:       true,
		LastUsedAt:     now,
		CreatedAt:      now,
	}

	if !token.IsActive {
		t.Error("Expected token to be active")
	}
	if token.Platform != "ios" {
		t.Errorf("Expected platform 'ios', got %q", token.Platform)
	}
}

// ============================================================
// STUB PUSH PROVIDER TESTS
// ============================================================

func TestStubPushProvider_NoError(t *testing.T) {
	provider := &StubPushProvider{}
	ctx := context.Background()

	notification := &PushNotification{
		Type:  PushTypeIncidentCreated,
		Title: "Test",
		Body:  "Test body",
	}

	err := provider.Send(ctx, "ios", "test-token-123", notification)
	if err != nil {
		t.Errorf("StubPushProvider should not return error, got: %v", err)
	}
}

// ============================================================
// TRUNCATE TOKEN HELPER
// ============================================================

func TestTruncateToken(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"short", "short"},
		{"exactly12ch", "exactly12ch"},
		{"a-long-push-token-that-should-be-truncated", "a-long-push-..."},
		{"", ""},
	}

	for _, tt := range tests {
		result := truncateToken(tt.input)
		if result != tt.expected {
			t.Errorf("truncateToken(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

// ============================================================
// MAX TOKENS CONSTANT
// ============================================================

func TestMaxTokensPerUser(t *testing.T) {
	if MaxTokensPerUser != 5 {
		t.Errorf("MaxTokensPerUser should be 5, got %d", MaxTokensPerUser)
	}
}

// ============================================================
// PUSH INTEGRATION EVENT MAPPING TESTS
// ============================================================

func TestEventPushMappings_BreachEventsAreCritical(t *testing.T) {
	breachEvents := []string{
		EventBreachDeadlineApproaching,
		EventBreachDeadlineExpired,
		EventBreachDPANotified,
		EventNIS2EarlyWarningDue,
		EventNIS2NotificationDue,
	}

	for _, event := range breachEvents {
		mapping, exists := eventPushMappings[event]
		if !exists {
			t.Errorf("Expected push mapping for breach event %q", event)
			continue
		}
		if !mapping.IsCritical {
			t.Errorf("Breach event %q should be marked as critical", event)
		}
		if mapping.PushType != PushTypeBreachAlert {
			t.Errorf("Breach event %q should have push type %q, got %q",
				event, PushTypeBreachAlert, mapping.PushType)
		}
		if mapping.Priority != "high" {
			t.Errorf("Breach event %q should have high priority, got %q",
				event, mapping.Priority)
		}
	}
}

func TestEventPushMappings_NonBreachEventsNotCritical(t *testing.T) {
	nonCriticalEvents := []string{
		EventIncidentCreated,
		EventIncidentUpdated,
		EventPolicyReviewDue,
		EventPolicyReviewOverdue,
		EventFindingCreated,
		EventRiskReviewDue,
		EventVendorAssessmentDue,
	}

	for _, event := range nonCriticalEvents {
		mapping, exists := eventPushMappings[event]
		if !exists {
			t.Errorf("Expected push mapping for event %q", event)
			continue
		}
		if mapping.IsCritical {
			t.Errorf("Non-breach event %q should NOT be marked as critical", event)
		}
	}
}

func TestEventPushMappings_AllHaveRequiredFields(t *testing.T) {
	for event, mapping := range eventPushMappings {
		if mapping.PushType == "" {
			t.Errorf("Event %q mapping has empty PushType", event)
		}
		if mapping.TitleFmt == "" {
			t.Errorf("Event %q mapping has empty TitleFmt", event)
		}
		if mapping.BodyFmt == "" {
			t.Errorf("Event %q mapping has empty BodyFmt", event)
		}
		if mapping.Priority == "" {
			t.Errorf("Event %q mapping has empty Priority", event)
		}
		if mapping.Priority != "high" && mapping.Priority != "normal" {
			t.Errorf("Event %q mapping has invalid priority %q (must be 'high' or 'normal')",
				event, mapping.Priority)
		}
	}
}

// ============================================================
// ENRICH TITLE/BODY TESTS
// ============================================================

func TestEnrichTitle_WithReference(t *testing.T) {
	event := Event{
		Type:     EventIncidentCreated,
		OrgID:    uuid.New(),
		Severity: "critical",
		Payload: map[string]interface{}{
			"reference": "INC-001",
		},
	}

	title := enrichTitle("New Incident", event)
	if title != "[CRITICAL] New Incident [INC-001]" {
		t.Errorf("Expected '[CRITICAL] New Incident [INC-001]', got %q", title)
	}
}

func TestEnrichTitle_NoSeverityNoReference(t *testing.T) {
	event := Event{
		Type:    EventPolicyReviewDue,
		OrgID:   uuid.New(),
		Payload: map[string]interface{}{},
	}

	title := enrichTitle("Policy Review Due", event)
	if title != "Policy Review Due" {
		t.Errorf("Expected 'Policy Review Due', got %q", title)
	}
}

func TestEnrichBody_WithTitle(t *testing.T) {
	event := Event{
		Type:  EventIncidentCreated,
		OrgID: uuid.New(),
		Payload: map[string]interface{}{
			"title": "Server breach detected",
		},
	}

	body := enrichBody("A new incident has been reported.", event)
	expected := "A new incident has been reported. — Server breach detected"
	if body != expected {
		t.Errorf("Expected %q, got %q", expected, body)
	}
}

func TestEnrichBody_NoPayloadTitle(t *testing.T) {
	event := Event{
		Type:    EventIncidentCreated,
		OrgID:   uuid.New(),
		Payload: map[string]interface{}{},
	}

	body := enrichBody("A new incident has been reported.", event)
	if body != "A new incident has been reported." {
		t.Errorf("Expected original body, got %q", body)
	}
}

// ============================================================
// APPEND UNIQUE UUID
// ============================================================

func TestAppendUniqueUUID_NewID(t *testing.T) {
	id1 := uuid.New()
	id2 := uuid.New()
	slice := []uuid.UUID{id1}

	result := appendUniqueUUID(slice, id2)
	if len(result) != 2 {
		t.Errorf("Expected 2 elements, got %d", len(result))
	}
}

func TestAppendUniqueUUID_Duplicate(t *testing.T) {
	id1 := uuid.New()
	slice := []uuid.UUID{id1}

	result := appendUniqueUUID(slice, id1)
	if len(result) != 1 {
		t.Errorf("Expected 1 element (no duplicate), got %d", len(result))
	}
}

func TestAppendUniqueUUID_EmptySlice(t *testing.T) {
	id := uuid.New()
	result := appendUniqueUUID(nil, id)
	if len(result) != 1 {
		t.Errorf("Expected 1 element, got %d", len(result))
	}
}
