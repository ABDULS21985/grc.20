package service

import (
	"encoding/json"
	"testing"
	"time"
)

// ============================================================
// CONDITION EVALUATION TESTS
// ============================================================

func TestEvaluateCondition_Equals(t *testing.T) {
	metadata := json.RawMessage(`{"risk_level": "high", "score": 85}`)

	tests := []struct {
		name     string
		expr     ConditionExpression
		expected bool
	}{
		{
			name:     "equals match",
			expr:     ConditionExpression{Field: "risk_level", Operator: "eq", Value: "high"},
			expected: true,
		},
		{
			name:     "equals no match",
			expr:     ConditionExpression{Field: "risk_level", Operator: "eq", Value: "low"},
			expected: false,
		},
		{
			name:     "not equals match",
			expr:     ConditionExpression{Field: "risk_level", Operator: "neq", Value: "low"},
			expected: true,
		},
		{
			name:     "not equals no match",
			expr:     ConditionExpression{Field: "risk_level", Operator: "neq", Value: "high"},
			expected: false,
		},
		{
			name:     "double equals syntax",
			expr:     ConditionExpression{Field: "risk_level", Operator: "==", Value: "high"},
			expected: true,
		},
		{
			name:     "not equals != syntax",
			expr:     ConditionExpression{Field: "risk_level", Operator: "!=", Value: "low"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := evaluateCondition(tt.expr, metadata)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEvaluateCondition_In(t *testing.T) {
	metadata := json.RawMessage(`{"risk_level": "critical", "status": "open"}`)

	tests := []struct {
		name     string
		expr     ConditionExpression
		expected bool
	}{
		{
			name: "in list match",
			expr: ConditionExpression{
				Field:    "risk_level",
				Operator: "in",
				Value:    []interface{}{"critical", "high"},
			},
			expected: true,
		},
		{
			name: "in list no match",
			expr: ConditionExpression{
				Field:    "risk_level",
				Operator: "in",
				Value:    []interface{}{"low", "medium"},
			},
			expected: false,
		},
		{
			name: "not_in list match (value not in list)",
			expr: ConditionExpression{
				Field:    "risk_level",
				Operator: "not_in",
				Value:    []interface{}{"low", "medium"},
			},
			expected: true,
		},
		{
			name: "not_in list no match (value in list)",
			expr: ConditionExpression{
				Field:    "risk_level",
				Operator: "not_in",
				Value:    []interface{}{"critical", "high"},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := evaluateCondition(tt.expr, metadata)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEvaluateCondition_NumericComparisons(t *testing.T) {
	metadata := json.RawMessage(`{"score": 75, "threshold": 50}`)

	tests := []struct {
		name     string
		expr     ConditionExpression
		expected bool
	}{
		{
			name:     "greater than true",
			expr:     ConditionExpression{Field: "score", Operator: "gt", Value: float64(50)},
			expected: true,
		},
		{
			name:     "greater than false",
			expr:     ConditionExpression{Field: "score", Operator: "gt", Value: float64(80)},
			expected: false,
		},
		{
			name:     "greater than equal true",
			expr:     ConditionExpression{Field: "score", Operator: "gte", Value: float64(75)},
			expected: true,
		},
		{
			name:     "less than true",
			expr:     ConditionExpression{Field: "score", Operator: "lt", Value: float64(100)},
			expected: true,
		},
		{
			name:     "less than false",
			expr:     ConditionExpression{Field: "score", Operator: "lt", Value: float64(50)},
			expected: false,
		},
		{
			name:     "less than equal true",
			expr:     ConditionExpression{Field: "score", Operator: "lte", Value: float64(75)},
			expected: true,
		},
		{
			name:     "> syntax",
			expr:     ConditionExpression{Field: "score", Operator: ">", Value: float64(50)},
			expected: true,
		},
		{
			name:     ">= syntax",
			expr:     ConditionExpression{Field: "score", Operator: ">=", Value: float64(75)},
			expected: true,
		},
		{
			name:     "< syntax",
			expr:     ConditionExpression{Field: "score", Operator: "<", Value: float64(100)},
			expected: true,
		},
		{
			name:     "<= syntax",
			expr:     ConditionExpression{Field: "score", Operator: "<=", Value: float64(75)},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := evaluateCondition(tt.expr, metadata)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEvaluateCondition_Contains(t *testing.T) {
	metadata := json.RawMessage(`{"description": "This is a critical security finding"}`)

	tests := []struct {
		name     string
		expr     ConditionExpression
		expected bool
	}{
		{
			name:     "contains match",
			expr:     ConditionExpression{Field: "description", Operator: "contains", Value: "critical"},
			expected: true,
		},
		{
			name:     "contains no match",
			expr:     ConditionExpression{Field: "description", Operator: "contains", Value: "minor"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := evaluateCondition(tt.expr, metadata)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEvaluateCondition_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		expr     ConditionExpression
		metadata json.RawMessage
		expected bool
	}{
		{
			name:     "empty field",
			expr:     ConditionExpression{Field: "", Operator: "eq", Value: "test"},
			metadata: json.RawMessage(`{"key": "test"}`),
			expected: false,
		},
		{
			name:     "missing field",
			expr:     ConditionExpression{Field: "nonexistent", Operator: "eq", Value: "test"},
			metadata: json.RawMessage(`{"key": "test"}`),
			expected: false,
		},
		{
			name:     "invalid metadata JSON",
			expr:     ConditionExpression{Field: "key", Operator: "eq", Value: "test"},
			metadata: json.RawMessage(`invalid json`),
			expected: false,
		},
		{
			name:     "empty metadata",
			expr:     ConditionExpression{Field: "key", Operator: "eq", Value: "test"},
			metadata: json.RawMessage(`{}`),
			expected: false,
		},
		{
			name:     "unknown operator",
			expr:     ConditionExpression{Field: "key", Operator: "regex", Value: "test"},
			metadata: json.RawMessage(`{"key": "test"}`),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := evaluateCondition(tt.expr, tt.metadata)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// ============================================================
// ACTION VALIDATION TESTS
// ============================================================

func TestIsValidAction(t *testing.T) {
	tests := []struct {
		name     string
		stepType string
		action   string
		expected bool
	}{
		// Approval step
		{"approval approve", "approval", "approve", true},
		{"approval reject", "approval", "reject", true},
		{"approval complete invalid", "approval", "complete", false},

		// Review step
		{"review approve", "review", "approve", true},
		{"review reject", "review", "reject", true},
		{"review complete", "review", "complete", true},
		{"review skip invalid", "review", "skip", false},

		// Task step
		{"task complete", "task", "complete", true},
		{"task reject", "task", "reject", true},
		{"task approve invalid", "task", "approve", false},

		// Parallel gate step
		{"parallel_gate approve", "parallel_gate", "approve", true},
		{"parallel_gate reject", "parallel_gate", "reject", true},
		{"parallel_gate complete", "parallel_gate", "complete", true},

		// Default step type
		{"unknown complete", "unknown", "complete", true},
		{"unknown approve invalid", "unknown", "approve", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidAction(tt.stepType, tt.action)
			if result != tt.expected {
				t.Errorf("isValidAction(%q, %q) = %v, expected %v", tt.stepType, tt.action, result, tt.expected)
			}
		})
	}
}

// ============================================================
// EXECUTION STATUS RESOLUTION TESTS
// ============================================================

func TestResolveExecutionStatus(t *testing.T) {
	tests := []struct {
		action   string
		expected string
	}{
		{"approve", "approved"},
		{"reject", "rejected"},
		{"complete", "completed"},
		{"skip", "skipped"},
		{"delegate", "delegated"},
		{"escalate", "escalated"},
		{"unknown", "completed"},
	}

	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			result := resolveExecutionStatus(tt.action)
			if result != tt.expected {
				t.Errorf("resolveExecutionStatus(%q) = %q, expected %q", tt.action, result, tt.expected)
			}
		})
	}
}

// ============================================================
// SLA CALCULATION TESTS
// ============================================================

func TestSLAConfigParsing(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectTotal int
		expectStep  int
	}{
		{
			name:        "full config",
			input:       `{"total_sla_hours": 120, "step_default_sla_hours": 48}`,
			expectTotal: 120,
			expectStep:  48,
		},
		{
			name:        "empty config",
			input:       `{}`,
			expectTotal: 0,
			expectStep:  0,
		},
		{
			name:        "partial config",
			input:       `{"total_sla_hours": 240}`,
			expectTotal: 240,
			expectStep:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg SLAConfig
			if err := json.Unmarshal([]byte(tt.input), &cfg); err != nil {
				t.Fatalf("failed to parse SLA config: %v", err)
			}
			if cfg.TotalSLAHours != tt.expectTotal {
				t.Errorf("TotalSLAHours = %d, expected %d", cfg.TotalSLAHours, tt.expectTotal)
			}
			if cfg.StepDefaultSLAHours != tt.expectStep {
				t.Errorf("StepDefaultSLAHours = %d, expected %d", cfg.StepDefaultSLAHours, tt.expectStep)
			}
		})
	}
}

func TestSLADeadlineCalculation(t *testing.T) {
	now := time.Now()
	slaHours := 48

	deadline := now.Add(time.Duration(slaHours) * time.Hour)

	// Deadline should be 48 hours from now
	diff := deadline.Sub(now)
	expectedDiff := 48 * time.Hour

	if diff != expectedDiff {
		t.Errorf("SLA deadline diff = %v, expected %v", diff, expectedDiff)
	}

	// Check that a deadline in the past counts as breached
	pastDeadline := now.Add(-1 * time.Hour)
	if !pastDeadline.Before(now) {
		t.Error("Past deadline should be before now")
	}

	// Check at-risk calculation (within 20% of total SLA time)
	totalHours := 120.0
	atRiskThreshold := totalHours * 0.2 // 24 hours
	hoursRemaining := deadline.Sub(now).Hours()
	isAtRisk := hoursRemaining <= atRiskThreshold

	if isAtRisk {
		t.Error("48 hours remaining should not be at-risk for a 120-hour SLA")
	}

	shortDeadline := now.Add(20 * time.Hour)
	shortRemaining := shortDeadline.Sub(now).Hours()
	isShortAtRisk := shortRemaining <= atRiskThreshold

	if !isShortAtRisk {
		t.Error("20 hours remaining should be at-risk for a 120-hour SLA")
	}
}

// ============================================================
// HELPER FUNCTION TESTS
// ============================================================

func TestDefaultJSONB(t *testing.T) {
	tests := []struct {
		name     string
		input    json.RawMessage
		expected string
	}{
		{"nil input", nil, "{}"},
		{"empty input", json.RawMessage{}, "{}"},
		{"valid input", json.RawMessage(`{"key": "value"}`), `{"key": "value"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := defaultJSONB(tt.input)
			if string(result) != tt.expected {
				t.Errorf("defaultJSONB() = %q, expected %q", string(result), tt.expected)
			}
		})
	}
}

func TestNullableApprovalMode(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "any_one"},
		{"any_one", "any_one"},
		{"all_required", "all_required"},
		{"majority", "majority"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := nullableApprovalMode(tt.input)
			if result != tt.expected {
				t.Errorf("nullableApprovalMode(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestToFloat(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected float64
	}{
		{"float64", float64(42.5), 42.5},
		{"float32", float32(42.5), float64(float32(42.5))},
		{"int", 42, 42.0},
		{"int64", int64(42), 42.0},
		{"string", "not a number", 0},
		{"nil", nil, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toFloat(tt.input)
			if result != tt.expected {
				t.Errorf("toFloat(%v) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

// ============================================================
// CONDITIONAL BRANCHING INTEGRATION TESTS
// ============================================================

func TestConditionalBranching_RiskAcceptance(t *testing.T) {
	// Simulate the Risk Acceptance workflow condition:
	// If risk_level is in [critical, high] -> route to CISO
	// Else -> route to Risk Manager

	condExpr := ConditionExpression{
		Field:    "risk_level",
		Operator: "in",
		Value:    []interface{}{"critical", "high"},
	}

	tests := []struct {
		riskLevel    string
		expectCISO   bool
	}{
		{"critical", true},
		{"high", true},
		{"medium", false},
		{"low", false},
		{"very_low", false},
	}

	for _, tt := range tests {
		t.Run(tt.riskLevel, func(t *testing.T) {
			metadata := json.RawMessage(`{"risk_level": "` + tt.riskLevel + `"}`)
			result := evaluateCondition(condExpr, metadata)
			if result != tt.expectCISO {
				t.Errorf("risk_level=%q: expected CISO route=%v, got %v",
					tt.riskLevel, tt.expectCISO, result)
			}
		})
	}
}

func TestConditionalBranching_ExceptionRequest(t *testing.T) {
	// Simulate the Exception Request workflow condition:
	// If impact_level is in [high, critical] -> route to Senior Management
	// Else -> route to Compliance Lead

	condExpr := ConditionExpression{
		Field:    "impact_level",
		Operator: "in",
		Value:    []interface{}{"high", "critical"},
	}

	tests := []struct {
		impactLevel        string
		expectSeniorMgmt   bool
	}{
		{"critical", true},
		{"high", true},
		{"medium", false},
		{"low", false},
	}

	for _, tt := range tests {
		t.Run(tt.impactLevel, func(t *testing.T) {
			metadata := json.RawMessage(`{"impact_level": "` + tt.impactLevel + `"}`)
			result := evaluateCondition(condExpr, metadata)
			if result != tt.expectSeniorMgmt {
				t.Errorf("impact_level=%q: expected senior management=%v, got %v",
					tt.impactLevel, tt.expectSeniorMgmt, result)
			}
		})
	}
}

func TestConditionalBranching_ThresholdBased(t *testing.T) {
	// Test numeric threshold-based condition
	condExpr := ConditionExpression{
		Field:    "compliance_score",
		Operator: "lt",
		Value:    float64(70),
	}

	tests := []struct {
		score    float64
		expected bool
	}{
		{50, true},
		{69.9, true},
		{70, false},
		{95, false},
	}

	for _, tt := range tests {
		t.Run("score", func(t *testing.T) {
			scoreJSON, _ := json.Marshal(map[string]float64{"compliance_score": tt.score})
			result := evaluateCondition(condExpr, json.RawMessage(scoreJSON))
			if result != tt.expected {
				t.Errorf("score=%.1f: expected %v, got %v", tt.score, tt.expected, result)
			}
		})
	}
}
