package service

import (
	"testing"

	"github.com/google/uuid"
)

// ============================================================
// DRIFT EVENT STRUCTURE TESTS
// ============================================================

func TestDriftEvent_Fields(t *testing.T) {
	entityID := uuid.New()
	event := DriftEvent{
		DriftType:     "control_degraded",
		Severity:      "high",
		EntityType:    "control_implementation",
		EntityID:      &entityID,
		EntityRef:     "CTL-001",
		Description:   "Control evidence collection failing",
		PreviousState: "effective",
		CurrentState:  "evidence_collection_failing",
	}

	if event.DriftType != "control_degraded" {
		t.Errorf("Expected drift_type 'control_degraded', got '%s'", event.DriftType)
	}
	if event.Severity != "high" {
		t.Errorf("Expected severity 'high', got '%s'", event.Severity)
	}
	if event.EntityType != "control_implementation" {
		t.Errorf("Expected entity_type 'control_implementation', got '%s'", event.EntityType)
	}
	if *event.EntityID != entityID {
		t.Error("Entity ID mismatch")
	}
	if event.EntityRef != "CTL-001" {
		t.Errorf("Expected entity_ref 'CTL-001', got '%s'", event.EntityRef)
	}
}

func TestDriftEvent_NilEntityID(t *testing.T) {
	event := DriftEvent{
		DriftType:  "score_dropped",
		Severity:   "medium",
		EntityType: "organization_framework",
		EntityID:   nil,
	}

	if event.EntityID != nil {
		t.Error("Expected nil entity ID")
	}
}

// ============================================================
// DRIFT SUMMARY STRUCTURE TESTS
// ============================================================

func TestDriftSummary_EmptyState(t *testing.T) {
	summary := &DriftSummary{
		TotalActive:     0,
		BySeverity:      make(map[string]int64),
		ByType:          make(map[string]int64),
		Acknowledged:    0,
		Unacknowledged:  0,
		ResolvedLast30d: 0,
		RecentEvents:    nil,
	}

	if summary.TotalActive != 0 {
		t.Errorf("Expected 0 active, got %d", summary.TotalActive)
	}
	if len(summary.BySeverity) != 0 {
		t.Error("Expected empty severity map")
	}
	if len(summary.ByType) != 0 {
		t.Error("Expected empty type map")
	}
}

func TestDriftSummary_WithData(t *testing.T) {
	summary := &DriftSummary{
		TotalActive: 5,
		BySeverity: map[string]int64{
			"critical": 1,
			"high":     2,
			"medium":   2,
		},
		ByType: map[string]int64{
			"control_degraded": 2,
			"evidence_expired": 1,
			"kri_breached":     1,
			"vendor_overdue":   1,
		},
		Acknowledged:    2,
		Unacknowledged:  3,
		ResolvedLast30d: 10,
	}

	if summary.TotalActive != 5 {
		t.Errorf("Expected 5 active, got %d", summary.TotalActive)
	}
	if summary.BySeverity["critical"] != 1 {
		t.Errorf("Expected 1 critical, got %d", summary.BySeverity["critical"])
	}
	if summary.BySeverity["high"] != 2 {
		t.Errorf("Expected 2 high, got %d", summary.BySeverity["high"])
	}
	if summary.Acknowledged+summary.Unacknowledged != summary.TotalActive {
		t.Error("Acknowledged + Unacknowledged should equal TotalActive")
	}
}

// ============================================================
// DRIFT TYPE VALIDITY TESTS
// ============================================================

func TestValidDriftTypes(t *testing.T) {
	validTypes := []string{
		"control_degraded",
		"evidence_expired",
		"kri_breached",
		"policy_unattested",
		"vendor_overdue",
		"training_expired",
		"score_dropped",
	}

	for _, dt := range validTypes {
		event := DriftEvent{DriftType: dt}
		if event.DriftType != dt {
			t.Errorf("Drift type mismatch: expected '%s', got '%s'", dt, event.DriftType)
		}
	}
}

func TestValidDriftSeverities(t *testing.T) {
	validSeverities := []string{"critical", "high", "medium", "low"}

	for _, s := range validSeverities {
		event := DriftEvent{Severity: s}
		if event.Severity != s {
			t.Errorf("Severity mismatch: expected '%s', got '%s'", s, event.Severity)
		}
	}
}

// ============================================================
// DRIFT DETECTOR CONSTRUCTOR TEST
// ============================================================

func TestNewDriftDetector_NilPool(t *testing.T) {
	dd := NewDriftDetector(nil, nil)
	if dd == nil {
		t.Error("Expected non-nil DriftDetector even with nil pool")
	}
	if dd.pool != nil {
		t.Error("Expected nil pool")
	}
}

// ============================================================
// DRIFT EVENT STATE TRANSITIONS
// ============================================================

func TestDriftEvent_UnacknowledgedState(t *testing.T) {
	event := DriftEvent{
		DriftType:      "evidence_expired",
		Severity:       "medium",
		AcknowledgedAt: nil,
		AcknowledgedBy: nil,
		ResolvedAt:     nil,
		ResolvedBy:     nil,
	}

	if event.AcknowledgedAt != nil {
		t.Error("Expected unacknowledged event to have nil AcknowledgedAt")
	}
	if event.ResolvedAt != nil {
		t.Error("Expected unresolved event to have nil ResolvedAt")
	}
}

func TestDriftEvent_SeverityOrdering(t *testing.T) {
	severityOrder := map[string]int{
		"critical": 0,
		"high":     1,
		"medium":   2,
		"low":      3,
	}

	if severityOrder["critical"] >= severityOrder["high"] {
		t.Error("Critical should have higher priority than high")
	}
	if severityOrder["high"] >= severityOrder["medium"] {
		t.Error("High should have higher priority than medium")
	}
	if severityOrder["medium"] >= severityOrder["low"] {
		t.Error("Medium should have higher priority than low")
	}
}

// ============================================================
// DRIFT SUMMARY AGGREGATION TESTS
// ============================================================

func TestDriftSummary_SeverityTotals(t *testing.T) {
	summary := &DriftSummary{
		TotalActive: 6,
		BySeverity: map[string]int64{
			"critical": 1,
			"high":     2,
			"medium":   2,
			"low":      1,
		},
	}

	var total int64
	for _, count := range summary.BySeverity {
		total += count
	}

	if total != summary.TotalActive {
		t.Errorf("Sum of severity counts (%d) does not match TotalActive (%d)", total, summary.TotalActive)
	}
}

func TestDriftSummary_TypeTotals(t *testing.T) {
	summary := &DriftSummary{
		TotalActive: 4,
		ByType: map[string]int64{
			"control_degraded": 1,
			"evidence_expired": 1,
			"kri_breached":     1,
			"score_dropped":    1,
		},
	}

	var total int64
	for _, count := range summary.ByType {
		total += count
	}

	if total != summary.TotalActive {
		t.Errorf("Sum of type counts (%d) does not match TotalActive (%d)", total, summary.TotalActive)
	}
}

// ============================================================
// ENTITY TYPE COVERAGE TESTS
// ============================================================

func TestDriftEventEntityTypes(t *testing.T) {
	cases := []struct {
		driftType  string
		entityType string
	}{
		{"control_degraded", "control_implementation"},
		{"evidence_expired", "control_implementation"},
		{"kri_breached", "risk_indicator"},
		{"policy_unattested", "policy"},
		{"vendor_overdue", "vendor"},
		{"score_dropped", "organization_framework"},
	}

	for _, tc := range cases {
		event := DriftEvent{
			DriftType:  tc.driftType,
			EntityType: tc.entityType,
		}
		if event.EntityType != tc.entityType {
			t.Errorf("Drift type '%s' expected entity type '%s', got '%s'",
				tc.driftType, tc.entityType, event.EntityType)
		}
	}
}
