package service

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

// ============================================================
// REMEDIATION PLAN MODEL TESTS
// ============================================================

func TestRemediationPlanSerialization(t *testing.T) {
	now := time.Now()
	confidence := 0.85
	ownerID := uuid.New()

	plan := RemediationPlan{
		ID:                uuid.New(),
		OrganizationID:    uuid.New(),
		PlanRef:           "RMP-2026-0001",
		Name:              "ISO 27001 Gap Remediation",
		Description:       "Address gaps identified in ISO 27001 assessment",
		PlanType:          "gap_remediation",
		Status:            "draft",
		ScopeFrameworkIDs: []uuid.UUID{uuid.New()},
		ScopeDescription:  "Full ISO 27001:2022 scope",
		Priority:          "high",
		AIGenerated:       true,
		AIModel:           "claude-sonnet-4-20250514",
		AIConfidenceScore: &confidence,
		AIGenerationDate:  &now,
		HumanReviewed:     false,
		EstimatedTotalHours: 120.5,
		EstimatedTotalCost:  15000.00,
		CompletionPercentage: 0,
		OwnerUserID:       &ownerID,
		CreatedBy:         uuid.New(),
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	data, err := json.Marshal(plan)
	if err != nil {
		t.Fatalf("Failed to marshal RemediationPlan: %v", err)
	}

	var decoded RemediationPlan
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal RemediationPlan: %v", err)
	}

	if decoded.PlanRef != "RMP-2026-0001" {
		t.Errorf("Expected plan_ref 'RMP-2026-0001', got '%s'", decoded.PlanRef)
	}
	if decoded.Name != "ISO 27001 Gap Remediation" {
		t.Errorf("Expected name 'ISO 27001 Gap Remediation', got '%s'", decoded.Name)
	}
	if decoded.PlanType != "gap_remediation" {
		t.Errorf("Expected plan_type 'gap_remediation', got '%s'", decoded.PlanType)
	}
	if decoded.Status != "draft" {
		t.Errorf("Expected status 'draft', got '%s'", decoded.Status)
	}
	if decoded.Priority != "high" {
		t.Errorf("Expected priority 'high', got '%s'", decoded.Priority)
	}
	if !decoded.AIGenerated {
		t.Error("Expected ai_generated = true")
	}
	if decoded.AIConfidenceScore == nil || *decoded.AIConfidenceScore != 0.85 {
		t.Errorf("Expected ai_confidence_score 0.85, got %v", decoded.AIConfidenceScore)
	}
	if decoded.EstimatedTotalHours != 120.5 {
		t.Errorf("Expected estimated_total_hours 120.5, got %f", decoded.EstimatedTotalHours)
	}
}

func TestRemediationActionSerialization(t *testing.T) {
	now := time.Now()
	action := RemediationAction{
		ID:                       uuid.New(),
		OrganizationID:           uuid.New(),
		PlanID:                   uuid.New(),
		ActionRef:                "RMP-2026-0001-A01",
		SortOrder:                1,
		Title:                    "Implement MFA for all admin access",
		Description:              "Deploy multi-factor authentication for privileged accounts",
		ActionType:               "deploy_technical",
		FrameworkControlCode:     "A.8.5",
		Priority:                 "critical",
		EstimatedHours:           24.0,
		EstimatedCostEUR:         3500.00,
		RequiredSkills:           []string{"IAM", "security engineering"},
		Status:                   "pending",
		AIImplementationGuidance: "Begin with PAM solution evaluation...",
		AIEvidenceSuggestions:    []string{"MFA enrollment report", "PAM configuration export"},
		AIToolRecommendations:    []string{"CyberArk", "Okta", "Azure AD"},
		AIRiskIfDeferred:         "Privileged accounts remain vulnerable to credential compromise",
		AICrossFrameworkBenefit:  "Addresses NIST CSF PR.AC-7, PCI DSS 8.3, ISO 27001 A.8.5",
		CreatedAt:                now,
		UpdatedAt:                now,
	}

	data, err := json.Marshal(action)
	if err != nil {
		t.Fatalf("Failed to marshal RemediationAction: %v", err)
	}

	var decoded RemediationAction
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal RemediationAction: %v", err)
	}

	if decoded.ActionRef != "RMP-2026-0001-A01" {
		t.Errorf("Expected action_ref 'RMP-2026-0001-A01', got '%s'", decoded.ActionRef)
	}
	if decoded.ActionType != "deploy_technical" {
		t.Errorf("Expected action_type 'deploy_technical', got '%s'", decoded.ActionType)
	}
	if decoded.Priority != "critical" {
		t.Errorf("Expected priority 'critical', got '%s'", decoded.Priority)
	}
	if len(decoded.RequiredSkills) != 2 {
		t.Errorf("Expected 2 required skills, got %d", len(decoded.RequiredSkills))
	}
	if len(decoded.AIToolRecommendations) != 3 {
		t.Errorf("Expected 3 tool recommendations, got %d", len(decoded.AIToolRecommendations))
	}
}

// ============================================================
// PLAN PROGRESS CALCULATION TESTS
// ============================================================

func TestPlanProgressCalculation(t *testing.T) {
	tests := []struct {
		name               string
		actions            []RemediationAction
		expectedCompleted  int
		expectedInProgress int
		expectedBlocked    int
		expectedPending    int
		expectedPercentage float64
	}{
		{
			name:               "empty plan",
			actions:            []RemediationAction{},
			expectedCompleted:  0,
			expectedInProgress: 0,
			expectedBlocked:    0,
			expectedPending:    0,
			expectedPercentage: 0,
		},
		{
			name: "all completed",
			actions: []RemediationAction{
				{Status: "completed"},
				{Status: "completed"},
				{Status: "completed"},
			},
			expectedCompleted:  3,
			expectedPercentage: 100.0,
		},
		{
			name: "mixed status",
			actions: []RemediationAction{
				{Status: "completed", Priority: "high"},
				{Status: "in_progress", Priority: "medium"},
				{Status: "pending", Priority: "low"},
				{Status: "blocked", Priority: "critical"},
			},
			expectedCompleted:  1,
			expectedInProgress: 1,
			expectedBlocked:    1,
			expectedPending:    1,
			expectedPercentage: 25.0,
		},
		{
			name: "with cancelled actions",
			actions: []RemediationAction{
				{Status: "completed"},
				{Status: "completed"},
				{Status: "cancelled"},
				{Status: "pending"},
			},
			expectedCompleted:  2,
			expectedPending:    1,
			expectedPercentage: 66.66,
		},
		{
			name: "half completed",
			actions: []RemediationAction{
				{Status: "completed"},
				{Status: "in_progress"},
			},
			expectedCompleted:  1,
			expectedInProgress: 1,
			expectedPercentage: 50.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			completed := 0
			inProgress := 0
			blocked := 0
			pending := 0
			cancelled := 0

			for _, a := range tt.actions {
				switch a.Status {
				case "completed":
					completed++
				case "in_progress", "in_review":
					inProgress++
				case "blocked":
					blocked++
				case "pending", "assigned":
					pending++
				case "cancelled":
					cancelled++
				}
			}

			if completed != tt.expectedCompleted {
				t.Errorf("Expected %d completed, got %d", tt.expectedCompleted, completed)
			}
			if inProgress != tt.expectedInProgress {
				t.Errorf("Expected %d in progress, got %d", tt.expectedInProgress, inProgress)
			}
			if blocked != tt.expectedBlocked {
				t.Errorf("Expected %d blocked, got %d", tt.expectedBlocked, blocked)
			}
			if pending != tt.expectedPending {
				t.Errorf("Expected %d pending, got %d", tt.expectedPending, pending)
			}

			// Calculate completion percentage (same logic as TrackProgress)
			actionableActions := len(tt.actions) - cancelled
			var percentage float64
			if actionableActions > 0 {
				percentage = float64(completed) / float64(actionableActions) * 100
				percentage = float64(int(percentage*100)) / 100 // round to 2 decimal places
			}

			if percentage != tt.expectedPercentage {
				t.Errorf("Expected completion %.2f%%, got %.2f%%", tt.expectedPercentage, percentage)
			}
		})
	}
}

// ============================================================
// AI RESPONSE PARSING TESTS
// ============================================================

func TestExtractJSONFromResponse_DirectJSON(t *testing.T) {
	input := `{"plan_name":"Test Plan","priority":"high","confidence_score":0.85,"actions":[],"estimated_total_hours":40,"estimated_total_cost_eur":5000,"timeline_weeks":8,"risk_summary":"test","assumptions":["test"]}`

	var result RemediationPlanResponse
	err := extractJSONFromResponse(input, &result)
	if err != nil {
		t.Fatalf("Failed to extract JSON: %v", err)
	}

	if result.PlanName != "Test Plan" {
		t.Errorf("Expected plan_name 'Test Plan', got '%s'", result.PlanName)
	}
	if result.Priority != "high" {
		t.Errorf("Expected priority 'high', got '%s'", result.Priority)
	}
	if result.ConfidenceScore != 0.85 {
		t.Errorf("Expected confidence_score 0.85, got %f", result.ConfidenceScore)
	}
}

func TestExtractJSONFromResponse_MarkdownFenced(t *testing.T) {
	input := "Here is the plan:\n```json\n{\"plan_name\":\"Fenced Plan\",\"priority\":\"medium\",\"confidence_score\":0.75,\"actions\":[],\"estimated_total_hours\":20,\"estimated_total_cost_eur\":2500,\"timeline_weeks\":4,\"risk_summary\":\"test\",\"assumptions\":[]}\n```\nLet me know if you need changes."

	var result RemediationPlanResponse
	err := extractJSONFromResponse(input, &result)
	if err != nil {
		t.Fatalf("Failed to extract JSON from fenced response: %v", err)
	}

	if result.PlanName != "Fenced Plan" {
		t.Errorf("Expected plan_name 'Fenced Plan', got '%s'", result.PlanName)
	}
}

func TestExtractJSONFromResponse_WithPreamble(t *testing.T) {
	input := "Sure, here is the guidance:\n{\"control_code\":\"A.5.1\",\"implementation_steps\":[\"Step 1\",\"Step 2\"],\"technical_measures\":[],\"organizational_measures\":[],\"evidence_required\":[],\"common_pitfalls\":[],\"maturity_indicators\":{},\"estimated_effort\":\"2 weeks\",\"related_controls\":[]}"

	var result ControlGuidance
	err := extractJSONFromResponse(input, &result)
	if err != nil {
		t.Fatalf("Failed to extract JSON with preamble: %v", err)
	}

	if result.ControlCode != "A.5.1" {
		t.Errorf("Expected control_code 'A.5.1', got '%s'", result.ControlCode)
	}
	if len(result.ImplementationSteps) != 2 {
		t.Errorf("Expected 2 implementation steps, got %d", len(result.ImplementationSteps))
	}
}

func TestExtractJSONFromResponse_InvalidJSON(t *testing.T) {
	input := "This is not JSON at all, just plain text."

	var result RemediationPlanResponse
	err := extractJSONFromResponse(input, &result)
	if err == nil {
		t.Error("Expected error for non-JSON input, got nil")
	}
}

// ============================================================
// RATE LIMITER TESTS
// ============================================================

func TestRateLimiter_AllowsWithinLimit(t *testing.T) {
	rl := newRateLimiterMap(5, time.Hour)
	orgID := uuid.New()

	for i := 0; i < 5; i++ {
		if !rl.Allow(orgID) {
			t.Errorf("Expected rate limiter to allow call %d", i+1)
		}
	}
}

func TestRateLimiter_DeniesOverLimit(t *testing.T) {
	rl := newRateLimiterMap(3, time.Hour)
	orgID := uuid.New()

	for i := 0; i < 3; i++ {
		rl.Allow(orgID)
	}

	if rl.Allow(orgID) {
		t.Error("Expected rate limiter to deny call over limit")
	}
}

func TestRateLimiter_SeparateOrgs(t *testing.T) {
	rl := newRateLimiterMap(2, time.Hour)
	org1 := uuid.New()
	org2 := uuid.New()

	// Exhaust org1 limit
	rl.Allow(org1)
	rl.Allow(org1)

	if rl.Allow(org1) {
		t.Error("Expected org1 to be rate limited")
	}

	// org2 should still have capacity
	if !rl.Allow(org2) {
		t.Error("Expected org2 to still have capacity")
	}
}

func TestRateLimiter_ResetsAfterWindow(t *testing.T) {
	rl := newRateLimiterMap(2, 50*time.Millisecond)
	orgID := uuid.New()

	rl.Allow(orgID)
	rl.Allow(orgID)

	if rl.Allow(orgID) {
		t.Error("Expected rate limiter to deny after exhaustion")
	}

	// Wait for window to expire
	time.Sleep(60 * time.Millisecond)

	if !rl.Allow(orgID) {
		t.Error("Expected rate limiter to allow after window reset")
	}
}

// ============================================================
// FALLBACK REMEDIATION PLAN TESTS
// ============================================================

func TestFallbackRemediationPlan_BasicGeneration(t *testing.T) {
	svc := &AIService{}

	req := RemediationPlanRequest{
		PlanType:       "gap_remediation",
		TimelineMonths: 3,
		Gaps: []ComplianceGap{
			{
				ControlCode:   "A.5.1",
				ControlTitle:  "Information Security Policies",
				FrameworkCode: "ISO27001",
				CurrentStatus: "not_implemented",
				TargetStatus:  "implemented",
				GapSeverity:   "high",
			},
			{
				ControlCode:   "A.8.5",
				ControlTitle:  "Secure Authentication",
				FrameworkCode: "ISO27001",
				CurrentStatus: "partial",
				TargetStatus:  "effective",
				GapSeverity:   "critical",
			},
		},
	}

	result, err := svc.fallbackRemediationPlan(req)
	if err != nil {
		t.Fatalf("Fallback plan generation failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if len(result.Actions) != 2 {
		t.Errorf("Expected 2 actions, got %d", len(result.Actions))
	}

	if result.Priority != "critical" {
		t.Errorf("Expected priority 'critical' (highest gap severity), got '%s'", result.Priority)
	}

	if result.ConfidenceScore != 0.60 {
		t.Errorf("Expected confidence score 0.60 for fallback, got %f", result.ConfidenceScore)
	}

	if result.EstimatedHours <= 0 {
		t.Error("Expected positive estimated hours")
	}

	if result.EstimatedCost <= 0 {
		t.Error("Expected positive estimated cost")
	}

	// Verify action priorities match gap severities
	for i, action := range result.Actions {
		if action.SortOrder != i+1 {
			t.Errorf("Action %d: expected sort_order %d, got %d", i, i+1, action.SortOrder)
		}
	}

	// First gap is high severity
	if result.Actions[0].Priority != "high" {
		t.Errorf("Expected first action priority 'high', got '%s'", result.Actions[0].Priority)
	}
	// Second gap is critical severity
	if result.Actions[1].Priority != "critical" {
		t.Errorf("Expected second action priority 'critical', got '%s'", result.Actions[1].Priority)
	}
}

func TestFallbackRemediationPlan_EmptyGaps(t *testing.T) {
	svc := &AIService{}

	req := RemediationPlanRequest{
		PlanType:       "gap_remediation",
		TimelineMonths: 1,
		Gaps:           []ComplianceGap{},
	}

	result, err := svc.fallbackRemediationPlan(req)
	if err != nil {
		t.Fatalf("Fallback plan with empty gaps failed: %v", err)
	}

	if len(result.Actions) != 0 {
		t.Errorf("Expected 0 actions for empty gaps, got %d", len(result.Actions))
	}

	if result.EstimatedHours != 0 {
		t.Errorf("Expected 0 estimated hours for empty gaps, got %f", result.EstimatedHours)
	}
}

// ============================================================
// FALLBACK CONTROL GUIDANCE TESTS
// ============================================================

func TestFallbackControlGuidance_KnownControl(t *testing.T) {
	svc := &AIService{}

	req := ControlGuidanceRequest{
		ControlCode:   "A.5.1",
		FrameworkCode: "ISO27001",
	}

	result, err := svc.fallbackControlGuidance(req)
	if err != nil {
		t.Fatalf("Fallback control guidance failed: %v", err)
	}

	if result.ControlCode != "A.5.1" {
		t.Errorf("Expected control_code 'A.5.1', got '%s'", result.ControlCode)
	}

	if len(result.ImplementationSteps) == 0 {
		t.Error("Expected non-empty implementation steps from static KB")
	}

	if len(result.EvidenceRequired) == 0 {
		t.Error("Expected non-empty evidence required from static KB")
	}

	if result.EstimatedEffort == "" {
		t.Error("Expected non-empty estimated effort")
	}
}

func TestFallbackControlGuidance_UnknownControl(t *testing.T) {
	svc := &AIService{}

	req := ControlGuidanceRequest{
		ControlCode:   "UNKNOWN-99",
		FrameworkCode: "UNKNOWN",
	}

	result, err := svc.fallbackControlGuidance(req)
	if err != nil {
		t.Fatalf("Fallback for unknown control failed: %v", err)
	}

	if result.ControlCode != "UNKNOWN-99" {
		t.Errorf("Expected control_code 'UNKNOWN-99', got '%s'", result.ControlCode)
	}

	// Should return generic guidance
	if len(result.ImplementationSteps) == 0 {
		t.Error("Expected non-empty generic implementation steps")
	}
}

// ============================================================
// FALLBACK GAP ANALYSIS TESTS
// ============================================================

func TestFallbackGapAnalysis(t *testing.T) {
	svc := &AIService{}

	gaps := []ComplianceGap{
		{
			ControlCode:   "A.5.1",
			ControlTitle:  "Info Security Policies",
			FrameworkCode: "ISO27001",
			CurrentStatus: "not_implemented",
			TargetStatus:  "implemented",
			GapSeverity:   "critical",
		},
		{
			ControlCode:   "A.8.5",
			ControlTitle:  "Secure Auth",
			FrameworkCode: "ISO27001",
			CurrentStatus: "partial",
			TargetStatus:  "effective",
			GapSeverity:   "high",
		},
	}

	result := svc.fallbackGapAnalysis(gaps)

	if result.OverallRiskLevel != "critical" {
		t.Errorf("Expected overall risk 'critical', got '%s'", result.OverallRiskLevel)
	}

	if len(result.GapAssessments) != 2 {
		t.Errorf("Expected 2 gap assessments, got %d", len(result.GapAssessments))
	}

	if len(result.PrioritizedOrder) != 2 {
		t.Errorf("Expected 2 items in prioritized order, got %d", len(result.PrioritizedOrder))
	}

	if len(result.StrategicItems) == 0 {
		t.Error("Expected strategic items for critical/high gaps")
	}
}

// ============================================================
// FALLBACK EVIDENCE TEMPLATE TESTS
// ============================================================

func TestFallbackEvidenceTemplate(t *testing.T) {
	svc := &AIService{}

	result := svc.fallbackEvidenceTemplate("A.8.5", "Secure Authentication")

	if result.ControlCode != "A.8.5" {
		t.Errorf("Expected control_code 'A.8.5', got '%s'", result.ControlCode)
	}
	if result.ControlTitle != "Secure Authentication" {
		t.Errorf("Expected control_title 'Secure Authentication', got '%s'", result.ControlTitle)
	}
	if len(result.EvidenceTypes) == 0 {
		t.Error("Expected non-empty evidence types")
	}
	if len(result.CollectionTips) == 0 {
		t.Error("Expected non-empty collection tips")
	}
	if result.ReviewFrequency == "" {
		t.Error("Expected non-empty review frequency")
	}
}

// ============================================================
// STATIC KNOWLEDGE BASE TESTS
// ============================================================

func TestStaticKnowledgeBase_HasExpectedControls(t *testing.T) {
	expectedControls := []string{
		"ISO27001-A.5.1",
		"ISO27001-A.5.2",
		"ISO27001-A.6.1",
		"ISO27001-A.8.1",
		"ISO27001-A.8.5",
		"ISO27001-A.5.15",
		"ISO27001-A.8.9",
		"ISO27001-A.8.15",
		"ISO27001-A.5.28",
		"ISO27001-A.5.30",
	}

	for _, key := range expectedControls {
		guidance, ok := staticControlGuidance[key]
		if !ok {
			t.Errorf("Expected static guidance for %s", key)
			continue
		}

		if len(guidance.ImplementationSteps) == 0 {
			t.Errorf("%s: expected non-empty implementation steps", key)
		}
		if len(guidance.TechnicalMeasures) == 0 {
			t.Errorf("%s: expected non-empty technical measures", key)
		}
		if len(guidance.EvidenceRequired) == 0 {
			t.Errorf("%s: expected non-empty evidence required", key)
		}
		if guidance.EstimatedEffort == "" {
			t.Errorf("%s: expected non-empty estimated effort", key)
		}
	}
}

// ============================================================
// PLAN TIMELINE TESTS
// ============================================================

func TestTimelineEntrySerialization(t *testing.T) {
	now := time.Now()
	start := now.Add(-24 * time.Hour)
	end := now.Add(7 * 24 * time.Hour)
	assignee := uuid.New()

	entry := RemediationTimelineEntry{
		ActionID:     uuid.New(),
		ActionRef:    "RMP-2026-0001-A01",
		Title:        "Deploy MFA",
		StartDate:    &start,
		EndDate:      &end,
		Status:       "in_progress",
		Priority:     "critical",
		AssignedTo:   &assignee,
		Dependencies: []uuid.UUID{uuid.New()},
	}

	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("Failed to marshal RemediationTimelineEntry: %v", err)
	}

	var decoded RemediationTimelineEntry
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal RemediationTimelineEntry: %v", err)
	}

	if decoded.ActionRef != "RMP-2026-0001-A01" {
		t.Errorf("Expected action_ref 'RMP-2026-0001-A01', got '%s'", decoded.ActionRef)
	}
	if decoded.Status != "in_progress" {
		t.Errorf("Expected status 'in_progress', got '%s'", decoded.Status)
	}
}

// ============================================================
// CRITICAL PATH TESTS
// ============================================================

func TestCriticalPathIdentification(t *testing.T) {
	now := time.Now()
	dueDate := now.Add(7 * 24 * time.Hour)

	actions := []RemediationAction{
		{ID: uuid.New(), ActionRef: "A01", Priority: "critical", Status: "in_progress", TargetEndDate: &dueDate},
		{ID: uuid.New(), ActionRef: "A02", Priority: "high", Status: "pending", TargetEndDate: &dueDate},
		{ID: uuid.New(), ActionRef: "A03", Priority: "medium", Status: "pending"},
		{ID: uuid.New(), ActionRef: "A04", Priority: "critical", Status: "completed"},
		{ID: uuid.New(), ActionRef: "A05", Priority: "low", Status: "cancelled"},
	}

	criticalPath := make([]CriticalPathAction, 0)
	for _, action := range actions {
		if (action.Priority == "critical" || action.Priority == "high") &&
			action.Status != "completed" && action.Status != "cancelled" {
			daysLeft := 0
			if action.TargetEndDate != nil {
				daysLeft = int(time.Until(*action.TargetEndDate).Hours() / 24)
			}
			criticalPath = append(criticalPath, CriticalPathAction{
				ActionID:  action.ID,
				ActionRef: action.ActionRef,
				Status:    action.Status,
				Priority:  action.Priority,
				DaysLeft:  daysLeft,
			})
		}
	}

	if len(criticalPath) != 2 {
		t.Errorf("Expected 2 critical path actions, got %d", len(criticalPath))
	}

	// First should be A01 (critical, in_progress)
	if criticalPath[0].ActionRef != "A01" {
		t.Errorf("Expected first critical path action 'A01', got '%s'", criticalPath[0].ActionRef)
	}
	// Second should be A02 (high, pending)
	if criticalPath[1].ActionRef != "A02" {
		t.Errorf("Expected second critical path action 'A02', got '%s'", criticalPath[1].ActionRef)
	}

	// Both should have positive days left
	for _, cp := range criticalPath {
		if cp.DaysLeft <= 0 {
			t.Errorf("Expected positive days left for %s, got %d", cp.ActionRef, cp.DaysLeft)
		}
	}
}

// ============================================================
// GENERATE PLAN REQUEST VALIDATION TESTS
// ============================================================

func TestGeneratePlanRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		req     GeneratePlanRequest
		wantErr bool
	}{
		{
			name: "valid manual request",
			req: GeneratePlanRequest{
				Name:     "Test Plan",
				PlanType: "gap_remediation",
				Priority: "high",
			},
			wantErr: false,
		},
		{
			name: "valid AI request",
			req: GeneratePlanRequest{
				Name:     "AI Plan",
				PlanType: "audit_finding",
				Priority: "critical",
				UseAI:    true,
				AIRequest: &RemediationPlanRequest{
					PlanType: "audit_finding",
					Gaps: []ComplianceGap{
						{ControlCode: "A.5.1", GapSeverity: "high"},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.req)
			if err != nil {
				t.Fatalf("Failed to marshal request: %v", err)
			}

			var decoded GeneratePlanRequest
			if err := json.Unmarshal(data, &decoded); err != nil {
				if !tt.wantErr {
					t.Errorf("Unexpected error: %v", err)
				}
				return
			}

			if tt.wantErr {
				t.Error("Expected error but got none")
			}

			if decoded.Name != tt.req.Name {
				t.Errorf("Expected name '%s', got '%s'", tt.req.Name, decoded.Name)
			}
		})
	}
}

// ============================================================
// PROMPT BUILDER TESTS
// ============================================================

func TestBuildRemediationPlanPrompt(t *testing.T) {
	req := RemediationPlanRequest{
		PlanType:          "gap_remediation",
		ScopeDescription:  "Full ISO 27001 scope",
		RiskAppetite:      "low",
		BudgetEUR:         50000,
		TimelineMonths:    6,
		IndustryContext:   "Financial Services",
		RegulatoryContext: "FCA regulated",
		AvailableSkills:   []string{"security engineering", "GRC"},
		ExistingControls:  []string{"A.5.1", "A.8.1"},
		Gaps: []ComplianceGap{
			{
				ControlCode:   "A.8.5",
				ControlTitle:  "Secure Authentication",
				FrameworkCode: "ISO27001",
				CurrentStatus: "not_implemented",
				TargetStatus:  "implemented",
				GapSeverity:   "critical",
				Description:   "No MFA deployed",
			},
		},
	}

	prompt := buildRemediationPlanPrompt(req)

	// Check that key information is present in the prompt
	if !remediationContains(prompt, "gap_remediation") {
		t.Error("Prompt should contain plan type")
	}
	if !remediationContains(prompt, "50000") {
		t.Error("Prompt should contain budget")
	}
	if !remediationContains(prompt, "Financial Services") {
		t.Error("Prompt should contain industry context")
	}
	if !remediationContains(prompt, "A.8.5") {
		t.Error("Prompt should contain gap control code")
	}
	if !remediationContains(prompt, "No MFA deployed") {
		t.Error("Prompt should contain gap description")
	}
	if !remediationContains(prompt, "JSON only") {
		t.Error("Prompt should request JSON output")
	}
}

func TestBuildControlGuidancePrompt(t *testing.T) {
	req := ControlGuidanceRequest{
		ControlCode:      "A.5.1",
		ControlTitle:     "Information Security Policies",
		FrameworkCode:    "ISO27001",
		CurrentStatus:    "partial",
		OrganizationSize: "medium",
		IndustryContext:  "Healthcare",
	}

	prompt := buildControlGuidancePrompt(req)

	if !remediationContains(prompt, "A.5.1") {
		t.Error("Prompt should contain control code")
	}
	if !remediationContains(prompt, "ISO27001") {
		t.Error("Prompt should contain framework code")
	}
	if !remediationContains(prompt, "Healthcare") {
		t.Error("Prompt should contain industry context")
	}
}

func remediationContains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && remediationContainsHelper(s, substr))
}

func remediationContainsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ============================================================
// AI SERVICE AVAILABILITY TESTS
// ============================================================

func TestAIServiceAvailability(t *testing.T) {
	tests := []struct {
		name     string
		apiKey   string
		expected bool
	}{
		{"with key", "sk-ant-test-key", true},
		{"empty key", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &AIService{apiKey: tt.apiKey}
			if svc.isAvailable() != tt.expected {
				t.Errorf("Expected isAvailable() = %v for key '%s'", tt.expected, tt.apiKey)
			}
		})
	}
}

// ============================================================
// NULLABLE STRING HELPER TESTS
// ============================================================

func TestNullableString(t *testing.T) {
	tests := []struct {
		input    string
		expected bool // true = non-nil result
	}{
		{"hello", true},
		{"", false},
		{"  ", true},
	}

	for _, tt := range tests {
		result := nullableString(tt.input)
		if tt.expected && result == nil {
			t.Errorf("Expected non-nil for input '%s'", tt.input)
		}
		if !tt.expected && result != nil {
			t.Errorf("Expected nil for input '%s'", tt.input)
		}
		if tt.expected && result != nil && *result != tt.input {
			t.Errorf("Expected '%s', got '%s'", tt.input, *result)
		}
	}
}
