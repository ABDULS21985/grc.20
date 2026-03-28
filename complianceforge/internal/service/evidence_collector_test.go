package service

import (
	"encoding/json"
	"testing"
)

// ============================================================
// EVIDENCE VALIDATOR TESTS
// ============================================================

func TestValidateEvidence_NoCriteria(t *testing.T) {
	ec := &EvidenceCollector{}

	data := json.RawMessage(`{"status": "ok"}`)
	results, allPassed := ec.ValidateEvidence(json.RawMessage(`[]`), data)

	if !allPassed {
		t.Error("Expected all criteria to pass when no criteria defined")
	}
	if len(results) != 0 {
		t.Errorf("Expected 0 results, got %d", len(results))
	}
}

func TestValidateEvidence_NilCriteria(t *testing.T) {
	ec := &EvidenceCollector{}

	data := json.RawMessage(`{"status": "ok"}`)
	_, allPassed := ec.ValidateEvidence(nil, data)

	if !allPassed {
		t.Error("Expected all criteria to pass when criteria is nil")
	}
}

func TestValidateEvidence_InvalidCriteriaJSON(t *testing.T) {
	ec := &EvidenceCollector{}

	data := json.RawMessage(`{"status": "ok"}`)
	_, allPassed := ec.ValidateEvidence(json.RawMessage(`{invalid`), data)

	if !allPassed {
		t.Error("Expected all criteria to pass when criteria JSON is invalid (treated as no criteria)")
	}
}

func TestValidateEvidence_InvalidDataJSON(t *testing.T) {
	ec := &EvidenceCollector{}

	criteria := json.RawMessage(`[{"field":"status","operator":"equals","value":"ok"}]`)
	data := json.RawMessage(`not json`)
	results, allPassed := ec.ValidateEvidence(criteria, data)

	if allPassed {
		t.Error("Expected validation to fail for invalid JSON data")
	}
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	if results[0].Passed {
		t.Error("Expected the result to indicate failure")
	}
}

func TestValidateEvidence_Equals_Pass(t *testing.T) {
	ec := &EvidenceCollector{}

	criteria := json.RawMessage(`[{"field":"status","operator":"equals","value":"healthy"}]`)
	data := json.RawMessage(`{"status": "healthy"}`)
	results, allPassed := ec.ValidateEvidence(criteria, data)

	if !allPassed {
		t.Error("Expected equals check to pass")
	}
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	if !results[0].Passed {
		t.Errorf("Expected result to pass, got actual='%s'", results[0].Actual)
	}
}

func TestValidateEvidence_Equals_Fail(t *testing.T) {
	ec := &EvidenceCollector{}

	criteria := json.RawMessage(`[{"field":"status","operator":"equals","value":"healthy"}]`)
	data := json.RawMessage(`{"status": "degraded"}`)
	results, allPassed := ec.ValidateEvidence(criteria, data)

	if allPassed {
		t.Error("Expected equals check to fail")
	}
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	if results[0].Passed {
		t.Error("Expected result to fail")
	}
	if results[0].Actual != "degraded" {
		t.Errorf("Expected actual 'degraded', got '%s'", results[0].Actual)
	}
}

func TestValidateEvidence_NotEquals(t *testing.T) {
	ec := &EvidenceCollector{}

	criteria := json.RawMessage(`[{"field":"status","operator":"not_equals","value":"error"}]`)
	data := json.RawMessage(`{"status": "ok"}`)
	results, allPassed := ec.ValidateEvidence(criteria, data)

	if !allPassed {
		t.Error("Expected not_equals check to pass")
	}
	if len(results) != 1 || !results[0].Passed {
		t.Error("Expected result to pass")
	}
}

func TestValidateEvidence_NotEquals_Fail(t *testing.T) {
	ec := &EvidenceCollector{}

	criteria := json.RawMessage(`[{"field":"status","operator":"not_equals","value":"error"}]`)
	data := json.RawMessage(`{"status": "error"}`)
	_, allPassed := ec.ValidateEvidence(criteria, data)

	if allPassed {
		t.Error("Expected not_equals check to fail when values match")
	}
}

func TestValidateEvidence_GreaterThan_Pass(t *testing.T) {
	ec := &EvidenceCollector{}

	criteria := json.RawMessage(`[{"field":"score","operator":"greater_than","value":"80"}]`)
	data := json.RawMessage(`{"score": 95}`)
	results, allPassed := ec.ValidateEvidence(criteria, data)

	if !allPassed {
		t.Error("Expected greater_than check to pass (95 > 80)")
	}
	if len(results) != 1 || !results[0].Passed {
		t.Error("Expected result to pass")
	}
}

func TestValidateEvidence_GreaterThan_Fail(t *testing.T) {
	ec := &EvidenceCollector{}

	criteria := json.RawMessage(`[{"field":"score","operator":"greater_than","value":"80"}]`)
	data := json.RawMessage(`{"score": 50}`)
	_, allPassed := ec.ValidateEvidence(criteria, data)

	if allPassed {
		t.Error("Expected greater_than check to fail (50 < 80)")
	}
}

func TestValidateEvidence_LessThan_Pass(t *testing.T) {
	ec := &EvidenceCollector{}

	criteria := json.RawMessage(`[{"field":"error_rate","operator":"less_than","value":"5"}]`)
	data := json.RawMessage(`{"error_rate": 2.5}`)
	_, allPassed := ec.ValidateEvidence(criteria, data)

	if !allPassed {
		t.Error("Expected less_than check to pass (2.5 < 5)")
	}
}

func TestValidateEvidence_LessThan_Fail(t *testing.T) {
	ec := &EvidenceCollector{}

	criteria := json.RawMessage(`[{"field":"error_rate","operator":"less_than","value":"5"}]`)
	data := json.RawMessage(`{"error_rate": 10}`)
	_, allPassed := ec.ValidateEvidence(criteria, data)

	if allPassed {
		t.Error("Expected less_than check to fail (10 > 5)")
	}
}

func TestValidateEvidence_Contains_Pass(t *testing.T) {
	ec := &EvidenceCollector{}

	criteria := json.RawMessage(`[{"field":"message","operator":"contains","value":"compliant"}]`)
	data := json.RawMessage(`{"message": "System is fully compliant with requirements"}`)
	_, allPassed := ec.ValidateEvidence(criteria, data)

	if !allPassed {
		t.Error("Expected contains check to pass")
	}
}

func TestValidateEvidence_Contains_Fail(t *testing.T) {
	ec := &EvidenceCollector{}

	criteria := json.RawMessage(`[{"field":"message","operator":"contains","value":"compliant"}]`)
	data := json.RawMessage(`{"message": "System is not ready"}`)
	_, allPassed := ec.ValidateEvidence(criteria, data)

	if allPassed {
		t.Error("Expected contains check to fail")
	}
}

func TestValidateEvidence_MatchesRegex_Pass(t *testing.T) {
	ec := &EvidenceCollector{}

	criteria := json.RawMessage(`[{"field":"version","operator":"matches_regex","value":"^v[0-9]+\\.[0-9]+\\.[0-9]+$"}]`)
	data := json.RawMessage(`{"version": "v1.2.3"}`)
	_, allPassed := ec.ValidateEvidence(criteria, data)

	if !allPassed {
		t.Error("Expected matches_regex check to pass")
	}
}

func TestValidateEvidence_MatchesRegex_Fail(t *testing.T) {
	ec := &EvidenceCollector{}

	criteria := json.RawMessage(`[{"field":"version","operator":"matches_regex","value":"^v[0-9]+\\.[0-9]+\\.[0-9]+$"}]`)
	data := json.RawMessage(`{"version": "latest"}`)
	_, allPassed := ec.ValidateEvidence(criteria, data)

	if allPassed {
		t.Error("Expected matches_regex check to fail for 'latest'")
	}
}

func TestValidateEvidence_MatchesRegex_InvalidPattern(t *testing.T) {
	ec := &EvidenceCollector{}

	criteria := json.RawMessage(`[{"field":"value","operator":"matches_regex","value":"[invalid"}]`)
	data := json.RawMessage(`{"value": "test"}`)
	results, allPassed := ec.ValidateEvidence(criteria, data)

	if allPassed {
		t.Error("Expected invalid regex to cause failure")
	}
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	if results[0].Passed {
		t.Error("Expected result to fail for invalid regex")
	}
}

func TestValidateEvidence_MissingField(t *testing.T) {
	ec := &EvidenceCollector{}

	criteria := json.RawMessage(`[{"field":"nonexistent","operator":"equals","value":"x"}]`)
	data := json.RawMessage(`{"status": "ok"}`)
	results, allPassed := ec.ValidateEvidence(criteria, data)

	if allPassed {
		t.Error("Expected check to fail for missing field")
	}
	if len(results) != 1 || results[0].Passed {
		t.Error("Expected result to fail")
	}
}

func TestValidateEvidence_NestedField(t *testing.T) {
	ec := &EvidenceCollector{}

	criteria := json.RawMessage(`[{"field":"data.status","operator":"equals","value":"active"}]`)
	data := json.RawMessage(`{"data": {"status": "active", "count": 10}}`)
	results, allPassed := ec.ValidateEvidence(criteria, data)

	if !allPassed {
		t.Error("Expected nested field access to work")
	}
	if len(results) != 1 || !results[0].Passed {
		t.Error("Expected result to pass")
	}
}

func TestValidateEvidence_DeeplyNestedField(t *testing.T) {
	ec := &EvidenceCollector{}

	criteria := json.RawMessage(`[{"field":"response.body.results.status","operator":"equals","value":"passed"}]`)
	data := json.RawMessage(`{
		"response": {
			"body": {
				"results": {
					"status": "passed",
					"score": 100
				}
			}
		}
	}`)
	_, allPassed := ec.ValidateEvidence(criteria, data)

	if !allPassed {
		t.Error("Expected deeply nested field access to work")
	}
}

func TestValidateEvidence_MultipleCriteria_AllPass(t *testing.T) {
	ec := &EvidenceCollector{}

	criteria := json.RawMessage(`[
		{"field":"status","operator":"equals","value":"ok"},
		{"field":"score","operator":"greater_than","value":"90"},
		{"field":"errors","operator":"less_than","value":"5"}
	]`)
	data := json.RawMessage(`{"status": "ok", "score": 95, "errors": 0}`)
	results, allPassed := ec.ValidateEvidence(criteria, data)

	if !allPassed {
		t.Error("Expected all criteria to pass")
	}
	if len(results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(results))
	}
	for i, r := range results {
		if !r.Passed {
			t.Errorf("Criterion %d should have passed", i)
		}
	}
}

func TestValidateEvidence_MultipleCriteria_OneFails(t *testing.T) {
	ec := &EvidenceCollector{}

	criteria := json.RawMessage(`[
		{"field":"status","operator":"equals","value":"ok"},
		{"field":"score","operator":"greater_than","value":"90"},
		{"field":"errors","operator":"less_than","value":"5"}
	]`)
	data := json.RawMessage(`{"status": "ok", "score": 85, "errors": 0}`)
	results, allPassed := ec.ValidateEvidence(criteria, data)

	if allPassed {
		t.Error("Expected allPassed to be false when one criterion fails")
	}
	if len(results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(results))
	}
	// First should pass (status equals ok)
	if !results[0].Passed {
		t.Error("First criterion should have passed")
	}
	// Second should fail (85 < 90)
	if results[1].Passed {
		t.Error("Second criterion should have failed (score 85 not > 90)")
	}
	// Third should pass (0 < 5)
	if !results[2].Passed {
		t.Error("Third criterion should have passed")
	}
}

func TestValidateEvidence_CustomMessage(t *testing.T) {
	ec := &EvidenceCollector{}

	criteria := json.RawMessage(`[{"field":"uptime","operator":"greater_than","value":"99.9","message":"Uptime SLA not met"}]`)
	data := json.RawMessage(`{"uptime": 98.5}`)
	results, allPassed := ec.ValidateEvidence(criteria, data)

	if allPassed {
		t.Error("Expected check to fail")
	}
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	if results[0].Message != "Uptime SLA not met" {
		t.Errorf("Expected custom message, got '%s'", results[0].Message)
	}
}

func TestValidateEvidence_UnsupportedOperator(t *testing.T) {
	ec := &EvidenceCollector{}

	criteria := json.RawMessage(`[{"field":"x","operator":"between","value":"10"}]`)
	data := json.RawMessage(`{"x": 5}`)
	results, allPassed := ec.ValidateEvidence(criteria, data)

	if allPassed {
		t.Error("Expected unsupported operator to fail")
	}
	if len(results) != 1 || results[0].Passed {
		t.Error("Expected result to fail for unsupported operator")
	}
}

// ============================================================
// RESOLVE FIELD TESTS
// ============================================================

func TestResolveField_TopLevel(t *testing.T) {
	data := map[string]interface{}{
		"key": "value",
	}
	val := resolveField("key", data)
	if val != "value" {
		t.Errorf("Expected 'value', got '%v'", val)
	}
}

func TestResolveField_Nested(t *testing.T) {
	data := map[string]interface{}{
		"a": map[string]interface{}{
			"b": map[string]interface{}{
				"c": "deep",
			},
		},
	}
	val := resolveField("a.b.c", data)
	if val != "deep" {
		t.Errorf("Expected 'deep', got '%v'", val)
	}
}

func TestResolveField_Missing(t *testing.T) {
	data := map[string]interface{}{
		"key": "value",
	}
	val := resolveField("missing", data)
	if val != nil {
		t.Errorf("Expected nil for missing field, got '%v'", val)
	}
}

func TestResolveField_PartialPath(t *testing.T) {
	data := map[string]interface{}{
		"a": "not_a_map",
	}
	val := resolveField("a.b", data)
	if val != nil {
		t.Errorf("Expected nil when path goes through non-map, got '%v'", val)
	}
}

// ============================================================
// COMPARE NUMERIC TESTS
// ============================================================

func TestCompareNumeric_FloatGreaterThan(t *testing.T) {
	if !compareNumeric(float64(10), "5", ">") {
		t.Error("Expected 10 > 5 to be true")
	}
}

func TestCompareNumeric_FloatLessThan(t *testing.T) {
	if !compareNumeric(float64(3), "5", "<") {
		t.Error("Expected 3 < 5 to be true")
	}
}

func TestCompareNumeric_IntConversion(t *testing.T) {
	if !compareNumeric(int(100), "50", ">") {
		t.Error("Expected int 100 > 50 to be true")
	}
}

func TestCompareNumeric_Int64Conversion(t *testing.T) {
	if !compareNumeric(int64(100), "50", ">") {
		t.Error("Expected int64 100 > 50 to be true")
	}
}

func TestCompareNumeric_StringConversion(t *testing.T) {
	if !compareNumeric("99.5", "50", ">") {
		t.Error("Expected string '99.5' > 50 to be true")
	}
}

func TestCompareNumeric_InvalidExpected(t *testing.T) {
	if compareNumeric(float64(10), "not_a_number", ">") {
		t.Error("Expected false for non-numeric expected value")
	}
}

func TestCompareNumeric_InvalidActual(t *testing.T) {
	if compareNumeric("not_a_number", "10", ">") {
		t.Error("Expected false for non-numeric actual value")
	}
}

func TestCompareNumeric_UnsupportedOp(t *testing.T) {
	if compareNumeric(float64(10), "5", "==") {
		t.Error("Expected false for unsupported operator")
	}
}

// ============================================================
// WEBHOOK SIGNATURE VALIDATION TESTS
// ============================================================

func TestValidateWebhookSignature_Valid(t *testing.T) {
	ec := &EvidenceCollector{}

	payload := []byte(`{"event":"test"}`)
	secret := "mysecret"

	// Compute expected signature
	expectedSig := computeHMACSHA256(payload, secret)

	if !ec.ValidateWebhookSignature(payload, expectedSig, secret) {
		t.Error("Expected valid signature to pass")
	}
}

func TestValidateWebhookSignature_ValidWithPrefix(t *testing.T) {
	ec := &EvidenceCollector{}

	payload := []byte(`{"event":"test"}`)
	secret := "mysecret"

	expectedSig := "sha256=" + computeHMACSHA256(payload, secret)

	if !ec.ValidateWebhookSignature(payload, expectedSig, secret) {
		t.Error("Expected valid signature with sha256= prefix to pass")
	}
}

func TestValidateWebhookSignature_Invalid(t *testing.T) {
	ec := &EvidenceCollector{}

	payload := []byte(`{"event":"test"}`)
	if ec.ValidateWebhookSignature(payload, "invalid_signature", "mysecret") {
		t.Error("Expected invalid signature to fail")
	}
}

func TestValidateWebhookSignature_WrongSecret(t *testing.T) {
	ec := &EvidenceCollector{}

	payload := []byte(`{"event":"test"}`)
	sig := computeHMACSHA256(payload, "correct_secret")

	if ec.ValidateWebhookSignature(payload, sig, "wrong_secret") {
		t.Error("Expected signature with wrong secret to fail")
	}
}

// helper to compute HMAC-SHA256 hex for tests
func computeHMACSHA256(payload []byte, secret string) string {
	import_hmac := hmacNew(sha256New, []byte(secret))
	import_hmac.Write(payload)
	return hexEncodeToString(import_hmac.Sum(nil))
}
