// Package validator provides input validation for API requests.
package validator

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// ValidationError holds field-level validation errors.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors is a collection of validation errors.
type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	msgs := make([]string, len(ve))
	for i, e := range ve {
		msgs[i] = fmt.Sprintf("%s: %s", e.Field, e.Message)
	}
	return strings.Join(msgs, "; ")
}

// HasErrors returns true if there are any validation errors.
func (ve ValidationErrors) HasErrors() bool {
	return len(ve) > 0
}

// ToMap converts errors to a map for JSON responses.
func (ve ValidationErrors) ToMap() map[string]string {
	m := make(map[string]string)
	for _, e := range ve {
		m[e.Field] = e.Message
	}
	return m
}

// Validate performs basic validation on common fields.
type Validator struct {
	errors ValidationErrors
}

func New() *Validator {
	return &Validator{}
}

func (v *Validator) Required(field, value string) *Validator {
	if strings.TrimSpace(value) == "" {
		v.errors = append(v.errors, ValidationError{Field: field, Message: "is required"})
	}
	return v
}

func (v *Validator) MinLength(field, value string, min int) *Validator {
	if len(value) < min {
		v.errors = append(v.errors, ValidationError{
			Field: field, Message: fmt.Sprintf("must be at least %d characters", min),
		})
	}
	return v
}

func (v *Validator) MaxLength(field, value string, max int) *Validator {
	if len(value) > max {
		v.errors = append(v.errors, ValidationError{
			Field: field, Message: fmt.Sprintf("must be at most %d characters", max),
		})
	}
	return v
}

func (v *Validator) Email(field, value string) *Validator {
	if !strings.Contains(value, "@") || !strings.Contains(value, ".") {
		v.errors = append(v.errors, ValidationError{Field: field, Message: "must be a valid email address"})
	}
	return v
}

func (v *Validator) UUID(field, value string) *Validator {
	if _, err := uuid.Parse(value); err != nil {
		v.errors = append(v.errors, ValidationError{Field: field, Message: "must be a valid UUID"})
	}
	return v
}

func (v *Validator) OneOf(field, value string, allowed []string) *Validator {
	for _, a := range allowed {
		if value == a {
			return v
		}
	}
	v.errors = append(v.errors, ValidationError{
		Field: field, Message: fmt.Sprintf("must be one of: %s", strings.Join(allowed, ", ")),
	})
	return v
}

func (v *Validator) IntRange(field string, value, min, max int) *Validator {
	if value < min || value > max {
		v.errors = append(v.errors, ValidationError{
			Field: field, Message: fmt.Sprintf("must be between %d and %d", min, max),
		})
	}
	return v
}

func (v *Validator) Errors() ValidationErrors {
	return v.errors
}

func (v *Validator) Valid() bool {
	return len(v.errors) == 0
}
