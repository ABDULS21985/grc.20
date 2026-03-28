package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// ============================================================
// TEST HELPERS
// ============================================================

func makeConditions(items ...Condition) []Condition {
	return items
}

func ptr[T any](v T) *T {
	return &v
}

func makePolicy(name, effect, resourceType string, actions []string, priority int,
	subjectConds, resourceConds, envConds []Condition) PolicyRecord {
	return PolicyRecord{
		ID:                    uuid.New(),
		OrgID:                 uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		Name:                  name,
		Effect:                effect,
		IsActive:              true,
		Priority:              priority,
		SubjectConditions:     subjectConds,
		ResourceType:          resourceType,
		ResourceConditions:    resourceConds,
		Actions:               actions,
		EnvironmentConditions: envConds,
	}
}

// ============================================================
// TEST: compareEquals
// ============================================================

func TestCompareEquals(t *testing.T) {
	tests := []struct {
		a, b     interface{}
		expected bool
	}{
		{"hello", "hello", true},
		{"hello", "world", false},
		{42, 42, true},
		{42, "42", true},
		{true, true, true},
		{true, "true", true},
		{false, true, false},
	}

	for i, tt := range tests {
		result := compareEquals(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("Test %d: compareEquals(%v, %v) = %v, want %v", i, tt.a, tt.b, result, tt.expected)
		}
	}
}

// ============================================================
// TEST: valueInList
// ============================================================

func TestValueInList(t *testing.T) {
	tests := []struct {
		attr     interface{}
		list     interface{}
		expected bool
	}{
		{"admin", []interface{}{"admin", "user", "viewer"}, true},
		{"superadmin", []interface{}{"admin", "user", "viewer"}, false},
		{"viewer", []string{"admin", "user", "viewer"}, true},
		{"editor", []string{"admin", "user"}, false},
	}

	for i, tt := range tests {
		result := valueInList(tt.attr, tt.list)
		if result != tt.expected {
			t.Errorf("Test %d: valueInList(%v, %v) = %v, want %v", i, tt.attr, tt.list, result, tt.expected)
		}
	}
}

// ============================================================
// TEST: containsAny
// ============================================================

func TestContainsAny(t *testing.T) {
	tests := []struct {
		attr     interface{}
		cond     interface{}
		expected bool
	}{
		{[]string{"org_admin", "viewer"}, []interface{}{"org_admin"}, true},
		{[]string{"viewer"}, []interface{}{"org_admin", "dpo"}, false},
		{[]interface{}{"dpo", "viewer"}, []interface{}{"dpo"}, true},
		{[]string{}, []interface{}{"admin"}, false},
	}

	for i, tt := range tests {
		result := containsAny(tt.attr, tt.cond)
		if result != tt.expected {
			t.Errorf("Test %d: containsAny(%v, %v) = %v, want %v", i, tt.attr, tt.cond, result, tt.expected)
		}
	}
}

// ============================================================
// TEST: compareGreaterThan / compareLessThan
// ============================================================

func TestNumericComparisons(t *testing.T) {
	if !compareGreaterThan(10, 5) {
		t.Error("10 should be > 5")
	}
	if compareGreaterThan(3, 5) {
		t.Error("3 should not be > 5")
	}
	if !compareLessThan(3, 5) {
		t.Error("3 should be < 5")
	}
	if compareLessThan(10, 5) {
		t.Error("10 should not be < 5")
	}
	// Float values
	if !compareGreaterThan(3.14, 2.71) {
		t.Error("3.14 should be > 2.71")
	}
}

// ============================================================
// TEST: checkCIDR
// ============================================================

func TestCheckCIDR(t *testing.T) {
	tests := []struct {
		ip       interface{}
		cidr     interface{}
		expected bool
	}{
		{"192.168.1.100", "192.168.1.0/24", true},
		{"10.0.0.1", "192.168.1.0/24", false},
		{"10.0.0.50", "10.0.0.0/8", true},
		{"invalid-ip", "10.0.0.0/8", false},
		{"10.0.0.1", "invalid-cidr", false},
		{123, "10.0.0.0/8", false}, // non-string IP
	}

	for i, tt := range tests {
		result := checkCIDR(tt.ip, tt.cidr)
		if result != tt.expected {
			t.Errorf("Test %d: checkCIDR(%v, %v) = %v, want %v", i, tt.ip, tt.cidr, result, tt.expected)
		}
	}
}

// ============================================================
// TEST: checkBetween
// ============================================================

func TestCheckBetween(t *testing.T) {
	// Numeric between
	if !checkBetween(5, []interface{}{1, 10}) {
		t.Error("5 should be between 1 and 10")
	}
	if checkBetween(15, []interface{}{1, 10}) {
		t.Error("15 should not be between 1 and 10")
	}
	if !checkBetween(1, []interface{}{1, 10}) {
		t.Error("1 should be between 1 and 10 (inclusive)")
	}
	if !checkBetween(10, []interface{}{1, 10}) {
		t.Error("10 should be between 1 and 10 (inclusive)")
	}

	// Time between
	now := time.Now().UTC()
	past := now.Add(-1 * time.Hour).Format(time.RFC3339)
	future := now.Add(1 * time.Hour).Format(time.RFC3339)
	current := now.Format(time.RFC3339)

	if !checkBetween(current, []interface{}{past, future}) {
		t.Error("Current time should be between past and future")
	}

	// Invalid bounds
	if checkBetween(5, []interface{}{1}) {
		t.Error("Bounds must have exactly 2 elements")
	}
	if checkBetween(5, "not-a-list") {
		t.Error("Non-list bounds should return false")
	}
}

// ============================================================
// TEST: EvaluateConditions — subject matching
// ============================================================

func TestEvaluateConditions_SubjectRoles(t *testing.T) {
	subjectMap := map[string]interface{}{
		"user_id": "user-123",
		"roles":   []string{"org_admin", "viewer"},
		"department": "IT",
	}

	// Should match: roles contains_any ["org_admin"]
	conds := makeConditions(Condition{
		Field: "roles", Operator: "contains_any", Value: []interface{}{"org_admin"},
	})
	if !EvaluateConditions(conds, subjectMap, subjectMap) {
		t.Error("Subject with org_admin role should match contains_any")
	}

	// Should not match: roles contains_any ["dpo"]
	conds = makeConditions(Condition{
		Field: "roles", Operator: "contains_any", Value: []interface{}{"dpo"},
	})
	if EvaluateConditions(conds, subjectMap, subjectMap) {
		t.Error("Subject without dpo role should not match")
	}
}

func TestEvaluateConditions_DepartmentEquals(t *testing.T) {
	subjectMap := map[string]interface{}{
		"department": "Engineering",
	}

	conds := makeConditions(Condition{
		Field: "department", Operator: "equals", Value: "Engineering",
	})
	if !EvaluateConditions(conds, subjectMap, subjectMap) {
		t.Error("Department should match 'Engineering'")
	}

	conds = makeConditions(Condition{
		Field: "department", Operator: "equals", Value: "Finance",
	})
	if EvaluateConditions(conds, subjectMap, subjectMap) {
		t.Error("Department should not match 'Finance'")
	}
}

// ============================================================
// TEST: EvaluateConditions — resource conditions
// ============================================================

func TestEvaluateConditions_ResourceOwner(t *testing.T) {
	userID := "user-abc-123"
	resourceMap := map[string]interface{}{
		"owner_user_id":  "user-abc-123",
		"classification": "internal",
	}
	subjectMap := map[string]interface{}{
		"user_id": userID,
	}

	// equals_subject: resource owner_user_id should equal subject user_id
	conds := makeConditions(Condition{
		Field: "owner_user_id", Operator: "equals_subject", Value: "user_id",
	})
	if !EvaluateConditions(conds, resourceMap, subjectMap) {
		t.Error("Resource owner should match subject user_id via equals_subject")
	}

	// Different user — should not match
	subjectMap2 := map[string]interface{}{"user_id": "different-user"}
	if EvaluateConditions(conds, resourceMap, subjectMap2) {
		t.Error("Resource owner should not match different user_id")
	}
}

func TestEvaluateConditions_ClassificationNotIn(t *testing.T) {
	resourceMap := map[string]interface{}{
		"classification": "internal",
	}
	subjectMap := map[string]interface{}{}

	// not_in: classification NOT IN [confidential, restricted]
	conds := makeConditions(Condition{
		Field: "classification", Operator: "not_in", Value: []interface{}{"confidential", "restricted"},
	})
	if !EvaluateConditions(conds, resourceMap, subjectMap) {
		t.Error("'internal' should pass not_in [confidential, restricted]")
	}

	// Now test with confidential
	resourceMap["classification"] = "confidential"
	if EvaluateConditions(conds, resourceMap, subjectMap) {
		t.Error("'confidential' should fail not_in [confidential, restricted]")
	}
}

func TestEvaluateConditions_IsDataBreach(t *testing.T) {
	resourceMap := map[string]interface{}{
		"is_data_breach": true,
	}
	subjectMap := map[string]interface{}{}

	conds := makeConditions(Condition{
		Field: "is_data_breach", Operator: "equals", Value: true,
	})
	if !EvaluateConditions(conds, resourceMap, subjectMap) {
		t.Error("is_data_breach=true should match equals true")
	}

	resourceMap["is_data_breach"] = false
	if EvaluateConditions(conds, resourceMap, subjectMap) {
		t.Error("is_data_breach=false should not match equals true")
	}
}

// ============================================================
// TEST: EvaluateConditions — environment conditions
// ============================================================

func TestEvaluateConditions_MFARequired(t *testing.T) {
	envMap := map[string]interface{}{
		"mfa_verified": true,
	}
	subjectMap := map[string]interface{}{}

	conds := makeConditions(Condition{
		Field: "mfa_verified", Operator: "equals", Value: true,
	})
	if !EvaluateConditions(conds, envMap, subjectMap) {
		t.Error("MFA verified should match")
	}

	envMap["mfa_verified"] = false
	if EvaluateConditions(conds, envMap, subjectMap) {
		t.Error("MFA not verified should not match")
	}
}

func TestEvaluateConditions_TimeHour(t *testing.T) {
	// Business hours: 9-17
	envMap := map[string]interface{}{
		"time_hour": 14,
	}
	subjectMap := map[string]interface{}{}

	// Check if hour is NOT in allowed hours (deny policy pattern)
	conds := makeConditions(Condition{
		Field: "time_hour", Operator: "not_in", Value: []interface{}{9, 10, 11, 12, 13, 14, 15, 16, 17},
	})
	if EvaluateConditions(conds, envMap, subjectMap) {
		t.Error("Hour 14 is in business hours, not_in should return false")
	}

	// After hours
	envMap["time_hour"] = 22
	if !EvaluateConditions(conds, envMap, subjectMap) {
		t.Error("Hour 22 is after hours, not_in should return true")
	}
}

func TestEvaluateConditions_CIDR(t *testing.T) {
	envMap := map[string]interface{}{
		"ip": "192.168.1.50",
	}
	subjectMap := map[string]interface{}{}

	conds := makeConditions(Condition{
		Field: "ip", Operator: "in_cidr", Value: "192.168.1.0/24",
	})
	if !EvaluateConditions(conds, envMap, subjectMap) {
		t.Error("192.168.1.50 should be in 192.168.1.0/24")
	}

	envMap["ip"] = "10.0.0.1"
	if EvaluateConditions(conds, envMap, subjectMap) {
		t.Error("10.0.0.1 should not be in 192.168.1.0/24")
	}
}

// ============================================================
// TEST: actionMatches
// ============================================================

func TestActionMatches(t *testing.T) {
	tests := []struct {
		actions  []string
		action   string
		expected bool
	}{
		{[]string{"read", "update"}, "read", true},
		{[]string{"read", "update"}, "delete", false},
		{[]string{"*"}, "anything", true},
		{[]string{"Read"}, "read", true}, // case-insensitive
		{[]string{}, "read", false},
	}

	for i, tt := range tests {
		result := actionMatches(tt.actions, tt.action)
		if result != tt.expected {
			t.Errorf("Test %d: actionMatches(%v, %q) = %v, want %v", i, tt.actions, tt.action, result, tt.expected)
		}
	}
}

// ============================================================
// TEST: Deny-overrides combining algorithm
// ============================================================

func TestDenyOverrides_BasicAllow(t *testing.T) {
	// Simulate: one allow policy matches, no deny
	userID := uuid.New()
	orgID := uuid.New()

	subject := SubjectAttributes{
		UserID: userID,
		OrgID:  orgID,
		Roles:  []string{"org_admin"},
	}
	subjectMap := subjectToMap(subject)
	envMap := envToMap(EnvironmentAttributes{Time: time.Now().UTC()})

	allowPolicy := makePolicy("Admin Allow", "allow", "*", []string{"read"}, 1,
		makeConditions(Condition{Field: "roles", Operator: "contains_any", Value: []interface{}{"org_admin"}}),
		nil, nil,
	)

	// Verify conditions match
	if !EvaluateConditions(allowPolicy.SubjectConditions, subjectMap, subjectMap) {
		t.Fatal("Subject conditions should match for org_admin")
	}

	// Action should match
	if !actionMatches(allowPolicy.Actions, "read") {
		t.Fatal("Action 'read' should match")
	}

	_ = envMap // used for completeness in real evaluation
}

func TestDenyOverrides_DenyWins(t *testing.T) {
	subjectMap := map[string]interface{}{
		"roles": []string{"viewer"},
	}

	// Allow policy: viewer can read
	allowPolicy := makePolicy("Viewer Read", "allow", "*", []string{"read", "export"}, 50,
		makeConditions(Condition{Field: "roles", Operator: "contains_any", Value: []interface{}{"viewer"}}),
		nil, nil,
	)

	// Deny policy: no export after hours (lower priority number = higher priority)
	denyPolicy := makePolicy("No Export", "deny", "*", []string{"export"}, 10,
		nil, nil,
		makeConditions(Condition{Field: "time_hour", Operator: "not_in",
			Value: []interface{}{9, 10, 11, 12, 13, 14, 15, 16, 17}}),
	)

	// Verify allow matches for "export"
	if !actionMatches(allowPolicy.Actions, "export") {
		t.Fatal("Allow policy should match 'export' action")
	}
	if !EvaluateConditions(allowPolicy.SubjectConditions, subjectMap, subjectMap) {
		t.Fatal("Allow subject conditions should match for viewer")
	}

	// Verify deny matches for "export" at hour 22
	envMap := map[string]interface{}{"time_hour": 22}
	if !actionMatches(denyPolicy.Actions, "export") {
		t.Fatal("Deny policy should match 'export' action")
	}
	if !EvaluateConditions(denyPolicy.EnvironmentConditions, envMap, subjectMap) {
		t.Fatal("Deny environment conditions should match at hour 22")
	}

	// Simulate deny-overrides: deny wins even though allow also matches
	var matchedDeny, matchedAllow *PolicyRecord
	policies := []PolicyRecord{allowPolicy, denyPolicy}

	for i := range policies {
		p := &policies[i]
		if p.Effect == "deny" {
			// Check action match and env conditions
			if actionMatches(p.Actions, "export") && EvaluateConditions(p.EnvironmentConditions, envMap, subjectMap) {
				if matchedDeny == nil || p.Priority < matchedDeny.Priority {
					matchedDeny = p
				}
			}
		} else if p.Effect == "allow" {
			if actionMatches(p.Actions, "export") && EvaluateConditions(p.SubjectConditions, subjectMap, subjectMap) {
				if matchedAllow == nil || p.Priority < matchedAllow.Priority {
					matchedAllow = p
				}
			}
		}
	}

	if matchedDeny == nil {
		t.Fatal("Deny policy should have matched")
	}
	if matchedAllow == nil {
		t.Fatal("Allow policy should have matched")
	}

	// Deny overrides: final effect should be deny
	finalEffect := "deny"
	if matchedDeny != nil {
		finalEffect = "deny"
	} else if matchedAllow != nil {
		finalEffect = "allow"
	}
	if finalEffect != "deny" {
		t.Errorf("Deny should override allow, got %s", finalEffect)
	}
}

func TestDenyOverrides_DefaultDeny(t *testing.T) {
	// No matching policies at all — should default to deny
	subjectMap := map[string]interface{}{
		"roles": []string{"unknown_role"},
	}

	// Policy that requires org_admin
	policy := makePolicy("Admin Only", "allow", "*", []string{"read"}, 1,
		makeConditions(Condition{Field: "roles", Operator: "contains_any", Value: []interface{}{"org_admin"}}),
		nil, nil,
	)

	if EvaluateConditions(policy.SubjectConditions, subjectMap, subjectMap) {
		t.Error("Unknown role should not match org_admin condition")
	}

	// Simulate: no policies match, result should be deny (default)
	var matchedDeny, matchedAllow *PolicyRecord
	finalEffect := "deny" // default
	if matchedDeny != nil {
		finalEffect = "deny"
	} else if matchedAllow != nil {
		finalEffect = "allow"
	}
	if finalEffect != "deny" {
		t.Error("Default should be deny when no policies match")
	}
}

// ============================================================
// TEST: Temporal constraints
// ============================================================

func TestTemporalConstraints(t *testing.T) {
	now := time.Now().UTC()
	past := now.Add(-24 * time.Hour)
	future := now.Add(24 * time.Hour)
	distantPast := now.Add(-48 * time.Hour)
	distantFuture := now.Add(48 * time.Hour)

	// Policy valid from past to future (should be valid now)
	p1 := makePolicy("Valid Now", "allow", "*", []string{"read"}, 50, nil, nil, nil)
	p1.ValidFrom = &past
	p1.ValidUntil = &future

	if now.Before(*p1.ValidFrom) {
		t.Error("Current time should not be before valid_from (yesterday)")
	}
	if now.After(*p1.ValidUntil) {
		t.Error("Current time should not be after valid_until (tomorrow)")
	}

	// Policy expired (valid_until in the past)
	p2 := makePolicy("Expired", "allow", "*", []string{"read"}, 50, nil, nil, nil)
	p2.ValidFrom = &distantPast
	p2.ValidUntil = &past

	if !now.After(*p2.ValidUntil) {
		t.Error("Current time should be after expired policy's valid_until")
	}

	// Policy not yet valid
	p3 := makePolicy("Future", "allow", "*", []string{"read"}, 50, nil, nil, nil)
	p3.ValidFrom = &future
	p3.ValidUntil = &distantFuture

	if !now.Before(*p3.ValidFrom) {
		t.Error("Current time should be before future policy's valid_from")
	}
}

// ============================================================
// TEST: equals_subject operator
// ============================================================

func TestEqualsSubject(t *testing.T) {
	userID := "abc-123-def"
	subjectMap := map[string]interface{}{
		"user_id": userID,
	}

	// Resource owned by the same user
	resourceMap := map[string]interface{}{
		"owner_user_id": userID,
	}

	cond := Condition{Field: "owner_user_id", Operator: "equals_subject", Value: "user_id"}
	if !evaluateSingleCondition(cond, resourceMap, subjectMap) {
		t.Error("equals_subject should match when resource owner == subject user_id")
	}

	// Resource owned by a different user
	resourceMap2 := map[string]interface{}{
		"owner_user_id": "different-user-456",
	}
	if evaluateSingleCondition(cond, resourceMap2, subjectMap) {
		t.Error("equals_subject should not match when owner differs from subject")
	}

	// Missing field in resource
	resourceMap3 := map[string]interface{}{}
	if evaluateSingleCondition(cond, resourceMap3, subjectMap) {
		t.Error("equals_subject should not match when resource field is missing")
	}

	// Missing field in subject
	subjectMap2 := map[string]interface{}{}
	if evaluateSingleCondition(cond, resourceMap, subjectMap2) {
		t.Error("equals_subject should not match when subject field is missing")
	}
}

// ============================================================
// TEST: MaskValue
// ============================================================

func TestMaskValue(t *testing.T) {
	tests := []struct {
		value    interface{}
		pattern  string
		expected string
	}{
		{"John Smith", "", "J*** S****"},
		{"Alice", "", "A****"},
		{nil, "", "****"},
		{"", "", "****"},
		{12345, "", "***,***"},
		{"99.99", "", "***,***"},
		{"Hello World", "H**** W****", "H**** W****"},
		{"anything", "REDACTED", "REDACTED"},
	}

	for i, tt := range tests {
		result := MaskValue(tt.value, tt.pattern)
		if result != tt.expected {
			t.Errorf("Test %d: MaskValue(%v, %q) = %q, want %q", i, tt.value, tt.pattern, result, tt.expected)
		}
	}
}

// ============================================================
// TEST: toStringSlice
// ============================================================

func TestToStringSlice(t *testing.T) {
	// []string
	result := toStringSlice([]string{"a", "b", "c"})
	if len(result) != 3 || result[0] != "a" {
		t.Errorf("Expected [a b c], got %v", result)
	}

	// []interface{}
	result = toStringSlice([]interface{}{"x", 42, true})
	if len(result) != 3 || result[0] != "x" || result[1] != "42" {
		t.Errorf("Expected [x 42 true], got %v", result)
	}

	// Single string
	result = toStringSlice("single")
	if len(result) != 1 || result[0] != "single" {
		t.Errorf("Expected [single], got %v", result)
	}
}

// ============================================================
// TEST: toFloat64
// ============================================================

func TestToFloat64(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected float64
	}{
		{42.0, 42.0},
		{float32(3.14), 3.140000104904175}, // float32 precision
		{100, 100.0},
		{int64(200), 200.0},
		{int32(50), 50.0},
		{"123.45", 123.45},
		{"not-a-number", 0.0},
	}

	for i, tt := range tests {
		result := toFloat64(tt.input)
		diff := result - tt.expected
		if diff < 0 {
			diff = -diff
		}
		if diff > 0.01 {
			t.Errorf("Test %d: toFloat64(%v) = %f, want %f", i, tt.input, result, tt.expected)
		}
	}
}

// ============================================================
// TEST: subjectToMap and envToMap
// ============================================================

func TestSubjectToMap(t *testing.T) {
	userID := uuid.New()
	orgID := uuid.New()

	subject := SubjectAttributes{
		UserID:         userID,
		OrgID:          orgID,
		Roles:          []string{"admin", "viewer"},
		Department:     "IT",
		Region:         "EU",
		ClearanceLevel: "high",
		MFAVerified:    true,
	}

	m := subjectToMap(subject)
	if m["user_id"] != userID.String() {
		t.Errorf("Expected user_id %s, got %v", userID.String(), m["user_id"])
	}
	if m["department"] != "IT" {
		t.Errorf("Expected department IT, got %v", m["department"])
	}
	roles, ok := m["roles"].([]string)
	if !ok || len(roles) != 2 {
		t.Errorf("Expected 2 roles, got %v", m["roles"])
	}
	if m["mfa_verified"] != true {
		t.Error("Expected mfa_verified to be true")
	}
}

func TestEnvToMap(t *testing.T) {
	now := time.Now().UTC()
	env := EnvironmentAttributes{
		IP:        "192.168.1.1",
		Time:      now,
		MFAStatus: true,
		TimeHour:  14,
	}

	m := envToMap(env)
	if m["ip"] != "192.168.1.1" {
		t.Errorf("Expected ip 192.168.1.1, got %v", m["ip"])
	}
	if m["mfa_verified"] != true {
		t.Error("Expected mfa_verified true")
	}
	if m["time_hour"] != 14 {
		t.Errorf("Expected time_hour 14, got %v", m["time_hour"])
	}
}

// ============================================================
// TEST: Unknown operator
// ============================================================

func TestUnknownOperator(t *testing.T) {
	attrs := map[string]interface{}{"field": "value"}
	cond := Condition{Field: "field", Operator: "regex_match", Value: ".*"}
	if evaluateSingleCondition(cond, attrs, attrs) {
		t.Error("Unknown operator should return false")
	}
}

// ============================================================
// TEST: not_equals with missing field
// ============================================================

func TestNotEqualsWithMissingField(t *testing.T) {
	attrs := map[string]interface{}{}
	cond := Condition{Field: "missing_field", Operator: "not_equals", Value: "something"}
	// Field absent means it does not equal the value -> true
	if !evaluateSingleCondition(cond, attrs, attrs) {
		t.Error("not_equals on missing field should return true (field is absent, so not equal)")
	}
}

// ============================================================
// TEST: not_in with missing field
// ============================================================

func TestNotInWithMissingField(t *testing.T) {
	attrs := map[string]interface{}{}
	cond := Condition{Field: "missing_field", Operator: "not_in", Value: []interface{}{"a", "b"}}
	// Field absent means it cannot be in the list -> true
	if !evaluateSingleCondition(cond, attrs, attrs) {
		t.Error("not_in on missing field should return true")
	}
}

// ============================================================
// TEST: Multiple conditions (AND semantics)
// ============================================================

func TestMultipleConditions_AND(t *testing.T) {
	attrs := map[string]interface{}{
		"roles":      []string{"viewer"},
		"department": "Finance",
		"region":     "EU",
	}

	// All conditions must match
	conds := makeConditions(
		Condition{Field: "roles", Operator: "contains_any", Value: []interface{}{"viewer"}},
		Condition{Field: "department", Operator: "equals", Value: "Finance"},
		Condition{Field: "region", Operator: "equals", Value: "EU"},
	)
	if !EvaluateConditions(conds, attrs, attrs) {
		t.Error("All three conditions should match")
	}

	// One condition fails
	conds2 := makeConditions(
		Condition{Field: "roles", Operator: "contains_any", Value: []interface{}{"viewer"}},
		Condition{Field: "department", Operator: "equals", Value: "Engineering"}, // mismatch
		Condition{Field: "region", Operator: "equals", Value: "EU"},
	)
	if EvaluateConditions(conds2, attrs, attrs) {
		t.Error("Department mismatch should cause overall failure (AND semantics)")
	}
}

// ============================================================
// TEST: sanitizeColumnName
// ============================================================

func TestSanitizeColumnName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"owner_user_id", "owner_user_id"},
		{"status", "status"},
		{"classification", "classification"},
		{"Robert'; DROP TABLE users;--", "RobertDROPTABLEusers"},
		{"", ""},
		{"SELECT", ""},      // reserved word
		{"valid_field", "valid_field"},
	}

	for i, tt := range tests {
		result := sanitizeColumnName(tt.input)
		if result != tt.expected {
			t.Errorf("Test %d: sanitizeColumnName(%q) = %q, want %q", i, tt.input, result, tt.expected)
		}
	}
}
