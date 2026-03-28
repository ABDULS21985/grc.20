package validator_test

import (
	"testing"

	"github.com/complianceforge/platform/internal/validator"
)

func TestRequiredField(t *testing.T) {
	v := validator.New()
	v.Required("name", "")
	if v.Valid() {
		t.Error("Empty required field should fail validation")
	}
	if len(v.Errors()) != 1 {
		t.Errorf("Expected 1 error, got %d", len(v.Errors()))
	}
}

func TestRequiredFieldWithValue(t *testing.T) {
	v := validator.New()
	v.Required("name", "John")
	if !v.Valid() {
		t.Error("Non-empty required field should pass")
	}
}

func TestMinLength(t *testing.T) {
	v := validator.New()
	v.MinLength("password", "short", 12)
	if v.Valid() {
		t.Error("Short password should fail min length check")
	}
}

func TestMinLengthPass(t *testing.T) {
	v := validator.New()
	v.MinLength("password", "this-is-a-long-password", 12)
	if !v.Valid() {
		t.Error("Long password should pass min length check")
	}
}

func TestEmail(t *testing.T) {
	tests := []struct {
		email string
		valid bool
	}{
		{"user@example.com", true},
		{"test@company.co.uk", true},
		{"invalid", false},
		{"@missing-local", false},
		{"no-domain@", false},
	}

	for _, tt := range tests {
		v := validator.New()
		v.Email("email", tt.email)
		if v.Valid() != tt.valid {
			t.Errorf("Email '%s': expected valid=%v, got valid=%v", tt.email, tt.valid, v.Valid())
		}
	}
}

func TestUUID(t *testing.T) {
	v := validator.New()
	v.UUID("id", "550e8400-e29b-41d4-a716-446655440000")
	if !v.Valid() {
		t.Error("Valid UUID should pass")
	}

	v2 := validator.New()
	v2.UUID("id", "not-a-uuid")
	if v2.Valid() {
		t.Error("Invalid UUID should fail")
	}
}

func TestOneOf(t *testing.T) {
	v := validator.New()
	v.OneOf("status", "active", []string{"active", "inactive", "locked"})
	if !v.Valid() {
		t.Error("Value in allowed list should pass")
	}

	v2 := validator.New()
	v2.OneOf("status", "deleted", []string{"active", "inactive", "locked"})
	if v2.Valid() {
		t.Error("Value not in allowed list should fail")
	}
}

func TestIntRange(t *testing.T) {
	v := validator.New()
	v.IntRange("likelihood", 3, 1, 5)
	if !v.Valid() {
		t.Error("Value within range should pass")
	}

	v2 := validator.New()
	v2.IntRange("likelihood", 6, 1, 5)
	if v2.Valid() {
		t.Error("Value outside range should fail")
	}
}

func TestChainedValidation(t *testing.T) {
	v := validator.New()
	v.Required("email", "user@example.com").
		Email("email", "user@example.com").
		Required("name", "John").
		MinLength("name", "John", 2)

	if !v.Valid() {
		t.Error("All valid fields should pass chained validation")
	}
}

func TestChainedValidationWithErrors(t *testing.T) {
	v := validator.New()
	v.Required("email", "").
		Required("name", "").
		MinLength("password", "short", 12)

	if v.Valid() {
		t.Error("Multiple invalid fields should fail")
	}
	if len(v.Errors()) != 3 {
		t.Errorf("Expected 3 errors, got %d", len(v.Errors()))
	}
}

func TestErrorsToMap(t *testing.T) {
	v := validator.New()
	v.Required("email", "").Required("name", "")

	m := v.Errors().ToMap()
	if len(m) != 2 {
		t.Errorf("Expected 2 error entries in map, got %d", len(m))
	}
	if _, ok := m["email"]; !ok {
		t.Error("Expected 'email' key in error map")
	}
}
