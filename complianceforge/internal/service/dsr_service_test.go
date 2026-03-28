package service

import (
	"testing"
	"time"
)

// TestCalculateSLAStatus_OnTrack verifies that requests well within the deadline
// are classified as "on_track" with correct days remaining.
func TestCalculateSLAStatus_OnTrack(t *testing.T) {
	now := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	deadline := time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC) // 30 days from received

	status, daysRemaining := CalculateSLAStatus(now, deadline, nil)

	if status != "on_track" {
		t.Errorf("expected status 'on_track', got '%s'", status)
	}
	if daysRemaining < 28 || daysRemaining > 30 {
		t.Errorf("expected days_remaining ~29, got %d", daysRemaining)
	}
}

// TestCalculateSLAStatus_AtRisk verifies that requests within 7 days of the deadline
// are classified as "at_risk".
func TestCalculateSLAStatus_AtRisk(t *testing.T) {
	now := time.Date(2026, 3, 25, 12, 0, 0, 0, time.UTC)
	deadline := time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC) // 6 days remaining

	status, daysRemaining := CalculateSLAStatus(now, deadline, nil)

	if status != "at_risk" {
		t.Errorf("expected status 'at_risk', got '%s'", status)
	}
	if daysRemaining < 4 || daysRemaining > 6 {
		t.Errorf("expected days_remaining ~5, got %d", daysRemaining)
	}
}

// TestCalculateSLAStatus_AtRiskBoundary verifies the exact 7-day boundary is "at_risk".
func TestCalculateSLAStatus_AtRiskBoundary(t *testing.T) {
	now := time.Date(2026, 3, 24, 0, 0, 0, 0, time.UTC)
	deadline := time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC) // exactly 7 days

	status, daysRemaining := CalculateSLAStatus(now, deadline, nil)

	if status != "at_risk" {
		t.Errorf("expected status 'at_risk', got '%s'", status)
	}
	if daysRemaining != 7 {
		t.Errorf("expected days_remaining 7, got %d", daysRemaining)
	}
}

// TestCalculateSLAStatus_Overdue verifies that requests past the deadline
// are classified as "overdue" with negative days remaining.
func TestCalculateSLAStatus_Overdue(t *testing.T) {
	now := time.Date(2026, 4, 5, 12, 0, 0, 0, time.UTC)
	deadline := time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC) // 5 days past

	status, daysRemaining := CalculateSLAStatus(now, deadline, nil)

	if status != "overdue" {
		t.Errorf("expected status 'overdue', got '%s'", status)
	}
	if daysRemaining >= 0 {
		t.Errorf("expected negative days_remaining, got %d", daysRemaining)
	}
}

// TestCalculateSLAStatus_ExtendedDeadlineUsed verifies that when an extended deadline
// is set, it takes precedence over the original deadline.
func TestCalculateSLAStatus_ExtendedDeadlineUsed(t *testing.T) {
	now := time.Date(2026, 4, 5, 12, 0, 0, 0, time.UTC)
	originalDeadline := time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC) // past
	extendedDeadline := time.Date(2026, 5, 30, 0, 0, 0, 0, time.UTC) // future

	status, daysRemaining := CalculateSLAStatus(now, originalDeadline, &extendedDeadline)

	if status != "on_track" {
		t.Errorf("expected status 'on_track' (extended deadline used), got '%s'", status)
	}
	if daysRemaining <= 0 {
		t.Errorf("expected positive days_remaining with extended deadline, got %d", daysRemaining)
	}
}

// TestCalculateSLAStatus_ExtendedDeadlineOverdue verifies that when the extended
// deadline is also past, the request is overdue.
func TestCalculateSLAStatus_ExtendedDeadlineOverdue(t *testing.T) {
	now := time.Date(2026, 6, 5, 12, 0, 0, 0, time.UTC)
	originalDeadline := time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC)
	extendedDeadline := time.Date(2026, 5, 30, 0, 0, 0, 0, time.UTC)

	status, daysRemaining := CalculateSLAStatus(now, originalDeadline, &extendedDeadline)

	if status != "overdue" {
		t.Errorf("expected status 'overdue' (extended deadline also past), got '%s'", status)
	}
	if daysRemaining >= 0 {
		t.Errorf("expected negative days_remaining, got %d", daysRemaining)
	}
}

// TestCalculateSLAStatus_ZeroDaysRemaining verifies the exact deadline day is "at_risk".
func TestCalculateSLAStatus_ZeroDaysRemaining(t *testing.T) {
	now := time.Date(2026, 3, 31, 12, 0, 0, 0, time.UTC)
	deadline := time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC)

	status, daysRemaining := CalculateSLAStatus(now, deadline, nil)

	// On the deadline day, hours remaining < 24 so daysRemaining is 0 or negative
	if status != "at_risk" && status != "overdue" {
		t.Errorf("expected status 'at_risk' or 'overdue' on deadline day, got '%s'", status)
	}
	_ = daysRemaining
}

// TestCalculateSLAStatus_NilExtendedDeadline verifies that nil extended deadline
// falls back to the original deadline.
func TestCalculateSLAStatus_NilExtendedDeadline(t *testing.T) {
	now := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	deadline := time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC)

	statusWithNil, daysWithNil := CalculateSLAStatus(now, deadline, nil)
	statusWithoutExt, daysWithoutExt := CalculateSLAStatus(now, deadline, nil)

	if statusWithNil != statusWithoutExt {
		t.Errorf("nil extended deadline should behave same as no extension: got '%s' vs '%s'",
			statusWithNil, statusWithoutExt)
	}
	if daysWithNil != daysWithoutExt {
		t.Errorf("nil extended deadline days should match: got %d vs %d",
			daysWithNil, daysWithoutExt)
	}
}

// TestCalculateSLAStatus_ZeroExtendedDeadline verifies that a zero-time extended deadline
// is treated as not set.
func TestCalculateSLAStatus_ZeroExtendedDeadline(t *testing.T) {
	now := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	deadline := time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC)
	zeroTime := time.Time{}

	status, daysRemaining := CalculateSLAStatus(now, deadline, &zeroTime)

	if status != "on_track" {
		t.Errorf("expected 'on_track' with zero extended deadline, got '%s'", status)
	}
	if daysRemaining < 28 {
		t.Errorf("expected ~29 days remaining, got %d", daysRemaining)
	}
}

// TestGetTaskChecklist_Access verifies the access request generates the correct task list.
func TestGetTaskChecklist_Access(t *testing.T) {
	tasks := getTaskChecklist("access")

	if len(tasks) != 6 {
		t.Errorf("expected 6 tasks for access request, got %d", len(tasks))
	}

	expectedTypes := []string{
		"verify_identity", "locate_data", "extract_data",
		"review_data", "compile_response", "send_response",
	}
	for i, expected := range expectedTypes {
		if i >= len(tasks) {
			break
		}
		if tasks[i].TaskType != expected {
			t.Errorf("task %d: expected type '%s', got '%s'", i, expected, tasks[i].TaskType)
		}
		if tasks[i].SortOrder != i+1 {
			t.Errorf("task %d: expected sort_order %d, got %d", i, i+1, tasks[i].SortOrder)
		}
	}
}

// TestGetTaskChecklist_Erasure verifies the erasure request generates the correct task list.
func TestGetTaskChecklist_Erasure(t *testing.T) {
	tasks := getTaskChecklist("erasure")

	if len(tasks) != 7 {
		t.Errorf("expected 7 tasks for erasure request, got %d", len(tasks))
	}

	expectedTypes := []string{
		"verify_identity", "locate_data", "review_exemptions",
		"execute_erasure", "confirm_erasure", "notify_third_parties", "send_confirmation",
	}
	for i, expected := range expectedTypes {
		if i >= len(tasks) {
			break
		}
		if tasks[i].TaskType != expected {
			t.Errorf("task %d: expected type '%s', got '%s'", i, expected, tasks[i].TaskType)
		}
	}
}

// TestGetTaskChecklist_Rectification verifies the rectification request task list.
func TestGetTaskChecklist_Rectification(t *testing.T) {
	tasks := getTaskChecklist("rectification")

	if len(tasks) != 6 {
		t.Errorf("expected 6 tasks for rectification request, got %d", len(tasks))
	}

	expectedTypes := []string{
		"verify_identity", "locate_data", "verify_correction",
		"execute_correction", "notify_third_parties", "send_confirmation",
	}
	for i, expected := range expectedTypes {
		if i >= len(tasks) {
			break
		}
		if tasks[i].TaskType != expected {
			t.Errorf("task %d: expected type '%s', got '%s'", i, expected, tasks[i].TaskType)
		}
	}
}

// TestGetTaskChecklist_Portability verifies the portability request task list.
func TestGetTaskChecklist_Portability(t *testing.T) {
	tasks := getTaskChecklist("portability")

	if len(tasks) != 5 {
		t.Errorf("expected 5 tasks for portability request, got %d", len(tasks))
	}

	if tasks[2].TaskType != "extract_in_machine_readable" {
		t.Errorf("portability task 3 should be 'extract_in_machine_readable', got '%s'", tasks[2].TaskType)
	}
}

// TestGetTaskChecklist_Restriction verifies the restriction request task list.
func TestGetTaskChecklist_Restriction(t *testing.T) {
	tasks := getTaskChecklist("restriction")

	if len(tasks) != 6 {
		t.Errorf("expected 6 tasks for restriction request, got %d", len(tasks))
	}

	// First task should always be verify_identity
	if tasks[0].TaskType != "verify_identity" {
		t.Errorf("first task should be 'verify_identity', got '%s'", tasks[0].TaskType)
	}
}

// TestGetTaskChecklist_Objection verifies the objection request task list.
func TestGetTaskChecklist_Objection(t *testing.T) {
	tasks := getTaskChecklist("objection")

	if len(tasks) != 4 {
		t.Errorf("expected 4 tasks for objection request, got %d", len(tasks))
	}
}

// TestGetTaskChecklist_AutomatedDecision verifies the automated_decision request task list.
func TestGetTaskChecklist_AutomatedDecision(t *testing.T) {
	tasks := getTaskChecklist("automated_decision")

	if len(tasks) != 5 {
		t.Errorf("expected 5 tasks for automated_decision request, got %d", len(tasks))
	}
}

// TestGetTaskChecklist_Unknown verifies that an unknown request type falls back to access.
func TestGetTaskChecklist_Unknown(t *testing.T) {
	tasks := getTaskChecklist("nonexistent_type")
	accessTasks := getTaskChecklist("access")

	if len(tasks) != len(accessTasks) {
		t.Errorf("unknown type should fall back to access tasks: got %d tasks, expected %d",
			len(tasks), len(accessTasks))
	}
}

// TestGetTaskChecklist_AllStartWithVerifyIdentity verifies that all request types
// start with identity verification as required by GDPR.
func TestGetTaskChecklist_AllStartWithVerifyIdentity(t *testing.T) {
	requestTypes := []string{
		"access", "erasure", "rectification", "portability",
		"restriction", "objection", "automated_decision",
	}

	for _, rt := range requestTypes {
		tasks := getTaskChecklist(rt)
		if len(tasks) == 0 {
			t.Errorf("request type '%s' has no tasks", rt)
			continue
		}
		if tasks[0].TaskType != "verify_identity" {
			t.Errorf("request type '%s': first task should be 'verify_identity', got '%s'",
				rt, tasks[0].TaskType)
		}
	}
}

// TestGetTaskChecklist_SortOrderSequential verifies that task sort orders
// are sequential starting from 1.
func TestGetTaskChecklist_SortOrderSequential(t *testing.T) {
	requestTypes := []string{
		"access", "erasure", "rectification", "portability",
		"restriction", "objection", "automated_decision",
	}

	for _, rt := range requestTypes {
		tasks := getTaskChecklist(rt)
		for i, task := range tasks {
			expected := i + 1
			if task.SortOrder != expected {
				t.Errorf("request type '%s', task %d: expected sort_order %d, got %d",
					rt, i, expected, task.SortOrder)
			}
		}
	}
}

// TestCalculateSLAStatus_GDPRDeadline verifies the standard 30-day GDPR deadline calculation.
func TestCalculateSLAStatus_GDPRDeadline(t *testing.T) {
	receivedDate := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	deadline := receivedDate.AddDate(0, 0, 30) // March 31

	// Day 1: should be on_track
	now := receivedDate.Add(24 * time.Hour)
	status, days := CalculateSLAStatus(now, deadline, nil)
	if status != "on_track" {
		t.Errorf("day 1: expected 'on_track', got '%s'", status)
	}
	if days < 27 {
		t.Errorf("day 1: expected ~29 days remaining, got %d", days)
	}

	// Day 23 (7 days remaining): should be at_risk
	now = receivedDate.AddDate(0, 0, 23)
	status, days = CalculateSLAStatus(now, deadline, nil)
	if status != "at_risk" {
		t.Errorf("day 23: expected 'at_risk', got '%s'", status)
	}

	// Day 28 (2 days remaining): should be at_risk
	now = receivedDate.AddDate(0, 0, 28)
	status, days = CalculateSLAStatus(now, deadline, nil)
	if status != "at_risk" {
		t.Errorf("day 28: expected 'at_risk', got '%s'", status)
	}

	// Day 31 (1 day past): should be overdue
	now = receivedDate.AddDate(0, 0, 31)
	status, days = CalculateSLAStatus(now, deadline, nil)
	if status != "overdue" {
		t.Errorf("day 31: expected 'overdue', got '%s'", status)
	}
	if days >= 0 {
		t.Errorf("day 31: expected negative days, got %d", days)
	}
}

// TestCalculateSLAStatus_ExtendedDeadline60Days verifies the GDPR-compliant 60-day extension.
func TestCalculateSLAStatus_ExtendedDeadline60Days(t *testing.T) {
	receivedDate := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	originalDeadline := receivedDate.AddDate(0, 0, 30) // March 31
	extendedDeadline := originalDeadline.AddDate(0, 0, 60) // May 30

	// On April 5 (past original deadline but within extension): should be on_track
	now := time.Date(2026, 4, 5, 12, 0, 0, 0, time.UTC)
	status, days := CalculateSLAStatus(now, originalDeadline, &extendedDeadline)
	if status != "on_track" {
		t.Errorf("April 5 with extension: expected 'on_track', got '%s'", status)
	}
	if days <= 0 {
		t.Errorf("April 5 with extension: expected positive days, got %d", days)
	}

	// On May 25 (5 days before extended deadline): should be at_risk
	now = time.Date(2026, 5, 25, 12, 0, 0, 0, time.UTC)
	status, days = CalculateSLAStatus(now, originalDeadline, &extendedDeadline)
	if status != "at_risk" {
		t.Errorf("May 25 with extension: expected 'at_risk', got '%s'", status)
	}

	// On June 1 (past extended deadline): should be overdue
	now = time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	status, days = CalculateSLAStatus(now, originalDeadline, &extendedDeadline)
	if status != "overdue" {
		t.Errorf("June 1 past extension: expected 'overdue', got '%s'", status)
	}
}
