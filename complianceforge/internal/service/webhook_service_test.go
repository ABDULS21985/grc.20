package service

import (
	"strings"
	"testing"
)

// ============================================================
// HMAC-SHA256 Signature Tests
// ============================================================

func TestGenerateSignature_Deterministic(t *testing.T) {
	svc := &WebhookService{}
	payload := []byte(`{"event":"risk.created","data":{"id":"abc123"}}`)
	secret := "whsec_test_secret_12345"

	sig1 := svc.GenerateSignature(payload, secret)
	sig2 := svc.GenerateSignature(payload, secret)

	if sig1 != sig2 {
		t.Errorf("GenerateSignature is not deterministic: %q != %q", sig1, sig2)
	}

	// Must start with sha256= prefix
	if !strings.HasPrefix(sig1, "sha256=") {
		t.Errorf("Signature should start with 'sha256=', got: %q", sig1)
	}

	// Must be sha256= prefix + 64 hex chars
	hexPart := strings.TrimPrefix(sig1, "sha256=")
	if len(hexPart) != 64 {
		t.Errorf("Expected 64 hex characters after 'sha256=', got %d: %q", len(hexPart), hexPart)
	}
}

func TestVerifySignature_Valid(t *testing.T) {
	svc := &WebhookService{}
	payload := []byte(`{"event":"control.updated","data":{"id":"xyz789"}}`)
	secret := "whsec_valid_secret_abcdef"

	sig := svc.GenerateSignature(payload, secret)
	if !svc.VerifySignature(payload, secret, sig) {
		t.Error("VerifySignature returned false for a valid signature")
	}
}

func TestVerifySignature_Invalid(t *testing.T) {
	svc := &WebhookService{}
	payload := []byte(`{"event":"incident.created"}`)
	secret := "whsec_my_secret"

	validSig := svc.GenerateSignature(payload, secret)
	// Tamper with the signature
	tampered := validSig[:len(validSig)-4] + "dead"

	if svc.VerifySignature(payload, secret, tampered) {
		t.Error("VerifySignature should return false for a tampered signature")
	}
}

func TestVerifySignature_Empty(t *testing.T) {
	svc := &WebhookService{}
	payload := []byte(`{"event":"ping"}`)

	// Empty signature
	if svc.VerifySignature(payload, "secret", "") {
		t.Error("VerifySignature should return false for an empty signature")
	}

	// Empty secret
	if svc.VerifySignature(payload, "", "sha256=abc") {
		t.Error("VerifySignature should return false for an empty secret")
	}

	// Both empty
	if svc.VerifySignature(payload, "", "") {
		t.Error("VerifySignature should return false when both are empty")
	}
}

func TestVerifySignature_WrongSecret(t *testing.T) {
	svc := &WebhookService{}
	payload := []byte(`{"event":"policy.published"}`)
	secret1 := "whsec_secret_one"
	secret2 := "whsec_secret_two"

	sig := svc.GenerateSignature(payload, secret1)
	if svc.VerifySignature(payload, secret2, sig) {
		t.Error("VerifySignature should return false when verified with a different secret")
	}
}

// ============================================================
// API Key Generation Tests
// ============================================================

func TestGenerateAPIKey_Format(t *testing.T) {
	fullKey, prefix, hash, err := GenerateAPIKeyString()
	if err != nil {
		t.Fatalf("GenerateAPIKeyString returned error: %v", err)
	}

	// Full key must start with cf_live_
	if !strings.HasPrefix(fullKey, "cf_live_") {
		t.Errorf("API key should start with 'cf_live_', got: %q", fullKey[:min(20, len(fullKey))])
	}

	// Key body after prefix should be 64 hex chars (32 bytes hex-encoded)
	body := strings.TrimPrefix(fullKey, "cf_live_")
	if len(body) != 64 {
		t.Errorf("Key body should be 64 hex characters, got %d: %q", len(body), body)
	}

	// Prefix should be first 16 characters of the full key
	if prefix != fullKey[:16] {
		t.Errorf("Prefix should be first 16 chars of full key, got prefix=%q vs expected=%q", prefix, fullKey[:16])
	}

	// Hash should be 64 hex chars (SHA-256)
	if len(hash) != 64 {
		t.Errorf("Hash should be 64 hex characters (SHA-256), got %d", len(hash))
	}
}

func TestGenerateAPIKey_UniquePrefix(t *testing.T) {
	keys := make(map[string]bool)
	for i := 0; i < 100; i++ {
		fullKey, _, _, err := GenerateAPIKeyString()
		if err != nil {
			t.Fatalf("GenerateAPIKeyString returned error on iteration %d: %v", i, err)
		}
		if keys[fullKey] {
			t.Fatalf("Duplicate key generated on iteration %d: %q", i, fullKey)
		}
		keys[fullKey] = true
	}
}

func TestAPIKeyHash_Consistent(t *testing.T) {
	key := "cf_live_abcdef0123456789abcdef0123456789abcdef0123456789abcdef01234567"

	hash1 := HashAPIKey(key)
	hash2 := HashAPIKey(key)

	if hash1 != hash2 {
		t.Errorf("HashAPIKey is not consistent: %q != %q", hash1, hash2)
	}

	// SHA-256 hash should be 64 hex chars
	if len(hash1) != 64 {
		t.Errorf("Hash should be 64 hex characters, got %d", len(hash1))
	}
}

func TestAPIKeyHash_DifferentKeys(t *testing.T) {
	key1 := "cf_live_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	key2 := "cf_live_bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"

	hash1 := HashAPIKey(key1)
	hash2 := HashAPIKey(key2)

	if hash1 == hash2 {
		t.Error("Different keys should produce different hashes")
	}
}

// ============================================================
// Webhook Event Types Tests
// ============================================================

func TestWebhookEventTypes_Complete(t *testing.T) {
	svc := &DeveloperPortalService{}
	events := svc.ListWebhookEventTypes()

	if len(events) < 40 {
		t.Errorf("Expected at least 40 webhook event types, got %d", len(events))
	}

	// Verify key categories exist
	categoryMap := make(map[string]int)
	for _, e := range events {
		categoryMap[e.Category]++
	}

	requiredCategories := []string{
		"Risk Management",
		"Control Management",
		"Policy Management",
		"Audit Management",
		"Incident Management",
		"Compliance",
		"Vendor Management",
		"Evidence",
		"Regulatory",
		"User Management",
		"Workflow",
		"System",
	}

	for _, cat := range requiredCategories {
		if categoryMap[cat] == 0 {
			t.Errorf("Missing required event category: %q", cat)
		}
	}

	// Verify specific critical events exist
	eventMap := make(map[string]bool)
	for _, e := range events {
		eventMap[e.EventType] = true
	}

	criticalEvents := []string{
		"risk.created",
		"control.status_changed",
		"policy.published",
		"incident.created",
		"incident.breach_detected",
		"compliance.score_changed",
		"vendor.risk_changed",
		"evidence.uploaded",
		"regulatory.change_detected",
		"ping",
	}

	for _, ev := range criticalEvents {
		if !eventMap[ev] {
			t.Errorf("Missing critical event type: %q", ev)
		}
	}

	// All events must have non-empty fields
	for _, e := range events {
		if e.EventType == "" {
			t.Error("Found event with empty EventType")
		}
		if e.Category == "" {
			t.Errorf("Event %q has empty Category", e.EventType)
		}
		if e.Description == "" {
			t.Errorf("Event %q has empty Description", e.EventType)
		}
		if e.Version == "" {
			t.Errorf("Event %q has empty Version", e.EventType)
		}
	}
}

// ============================================================
// API Scopes Tests
// ============================================================

func TestAPIScopes_ReadWriteDelete(t *testing.T) {
	svc := &DeveloperPortalService{}
	scopes := svc.ListAPIScopes()

	if len(scopes) == 0 {
		t.Fatal("Expected at least some API scopes, got 0")
	}

	// Verify access types are present
	accessTypes := make(map[string]int)
	for _, s := range scopes {
		accessTypes[s.Access]++
	}

	if accessTypes["read"] == 0 {
		t.Error("No read scopes found")
	}
	if accessTypes["write"] == 0 {
		t.Error("No write scopes found")
	}
	if accessTypes["delete"] == 0 {
		t.Error("No delete scopes found")
	}

	// Verify scope format: resource:access
	for _, s := range scopes {
		parts := strings.SplitN(s.Scope, ":", 2)
		if len(parts) != 2 {
			t.Errorf("Scope %q does not follow resource:access format", s.Scope)
		}
		if parts[0] == "" || parts[1] == "" {
			t.Errorf("Scope %q has empty resource or access part", s.Scope)
		}
	}

	// Verify key resource categories have read scopes
	scopeMap := make(map[string]bool)
	for _, s := range scopes {
		scopeMap[s.Scope] = true
	}

	requiredReadScopes := []string{
		"risks:read",
		"controls:read",
		"policies:read",
		"audits:read",
		"incidents:read",
		"vendors:read",
		"evidence:read",
		"compliance:read",
		"analytics:read",
	}

	for _, rs := range requiredReadScopes {
		if !scopeMap[rs] {
			t.Errorf("Missing required read scope: %q", rs)
		}
	}
}

// ============================================================
// Subscription Validation Tests
// ============================================================

func TestSubscriptionValidation_RequiresHTTPS(t *testing.T) {
	svc := &WebhookService{}

	// HTTPS URLs should pass validation (Subscribe will fail without DB, but
	// we test the URL validation logic via the error message)
	_, err := svc.Subscribe(nil, [16]byte{}, CreateSubscriptionReq{
		URL:    "http://example.com/webhook",
		Events: []string{"risk.created"},
	})

	if err == nil {
		t.Fatal("Expected error for HTTP URL, got nil")
	}

	if !strings.Contains(err.Error(), "HTTPS") {
		t.Errorf("Expected error message to mention HTTPS, got: %q", err.Error())
	}
}

func TestSubscriptionValidation_RejectsHTTP(t *testing.T) {
	svc := &WebhookService{}

	httpURLs := []string{
		"http://example.com/webhook",
		"http://my-server.local/hooks",
		"http://192.168.1.1:8080/callback",
	}

	for _, url := range httpURLs {
		_, err := svc.Subscribe(nil, [16]byte{}, CreateSubscriptionReq{
			URL:    url,
			Events: []string{"risk.created"},
		})
		if err == nil {
			t.Errorf("Expected error for HTTP URL %q, got nil", url)
			continue
		}
		if !strings.Contains(err.Error(), "HTTPS") {
			t.Errorf("Expected HTTPS error for URL %q, got: %q", url, err.Error())
		}
	}
}

func TestSubscriptionValidation_RequiresEvents(t *testing.T) {
	svc := &WebhookService{}

	_, err := svc.Subscribe(nil, [16]byte{}, CreateSubscriptionReq{
		URL:    "https://example.com/webhook",
		Events: []string{},
	})

	if err == nil {
		t.Fatal("Expected error for empty events, got nil")
	}

	if !strings.Contains(err.Error(), "at least one") {
		t.Errorf("Expected error about events requirement, got: %q", err.Error())
	}
}

// ============================================================
// Helpers
// ============================================================

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
