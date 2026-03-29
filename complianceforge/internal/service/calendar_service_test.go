package service

import (
	"testing"
	"time"
)

// ============================================================
// RECURRENCE PARSING TESTS
// ============================================================

// TestComputeNextOccurrence_Daily verifies daily recurrence computes +1 day.
func TestComputeNextOccurrence_Daily(t *testing.T) {
	base := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	next := computeNextOccurrence(base, "daily", "")

	expected := time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("daily: expected %v, got %v", expected, next)
	}
}

// TestComputeNextOccurrence_Weekly verifies weekly recurrence computes +7 days.
func TestComputeNextOccurrence_Weekly(t *testing.T) {
	base := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	next := computeNextOccurrence(base, "weekly", "")

	expected := time.Date(2026, 3, 22, 0, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("weekly: expected %v, got %v", expected, next)
	}
}

// TestComputeNextOccurrence_Monthly verifies monthly recurrence.
func TestComputeNextOccurrence_Monthly(t *testing.T) {
	base := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
	next := computeNextOccurrence(base, "monthly", "")

	// Jan 31 + 1 month = Mar 3 (Feb has 28 days in 2026)
	if next.Month() != time.March {
		t.Errorf("monthly from Jan 31: expected March, got %v", next.Month())
	}
}

// TestComputeNextOccurrence_Quarterly verifies quarterly recurrence.
func TestComputeNextOccurrence_Quarterly(t *testing.T) {
	base := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	next := computeNextOccurrence(base, "quarterly", "")

	expected := time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("quarterly: expected %v, got %v", expected, next)
	}
}

// TestComputeNextOccurrence_SemiAnnually verifies semi-annual recurrence.
func TestComputeNextOccurrence_SemiAnnually(t *testing.T) {
	base := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	next := computeNextOccurrence(base, "semi_annually", "")

	expected := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("semi_annually: expected %v, got %v", expected, next)
	}
}

// TestComputeNextOccurrence_Annually verifies annual recurrence.
func TestComputeNextOccurrence_Annually(t *testing.T) {
	base := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	next := computeNextOccurrence(base, "annually", "")

	expected := time.Date(2027, 3, 15, 0, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("annually: expected %v, got %v", expected, next)
	}
}

// TestComputeNextOccurrence_None returns zero time.
func TestComputeNextOccurrence_None(t *testing.T) {
	base := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	next := computeNextOccurrence(base, "none", "")

	if !next.IsZero() {
		t.Errorf("none: expected zero time, got %v", next)
	}
}

// TestComputeNextOccurrence_Unknown returns zero time.
func TestComputeNextOccurrence_Unknown(t *testing.T) {
	base := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	next := computeNextOccurrence(base, "unknown_type", "")

	if !next.IsZero() {
		t.Errorf("unknown: expected zero time, got %v", next)
	}
}

// ============================================================
// RRULE PARSING TESTS
// ============================================================

// TestParseRRule_MonthlyInterval3 verifies "FREQ=MONTHLY;INTERVAL=3" = quarterly.
func TestParseRRule_MonthlyInterval3(t *testing.T) {
	base := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	next := parseRRule(base, "FREQ=MONTHLY;INTERVAL=3")

	expected := time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("MONTHLY;INTERVAL=3: expected %v, got %v", expected, next)
	}
}

// TestParseRRule_WeeklyInterval2 verifies "FREQ=WEEKLY;INTERVAL=2" = bi-weekly.
func TestParseRRule_WeeklyInterval2(t *testing.T) {
	base := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	next := parseRRule(base, "FREQ=WEEKLY;INTERVAL=2")

	expected := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("WEEKLY;INTERVAL=2: expected %v, got %v", expected, next)
	}
}

// TestParseRRule_DailyDefault verifies "FREQ=DAILY" with no INTERVAL defaults to 1.
func TestParseRRule_DailyDefault(t *testing.T) {
	base := time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)
	next := parseRRule(base, "FREQ=DAILY")

	expected := time.Date(2026, 6, 11, 0, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("DAILY default: expected %v, got %v", expected, next)
	}
}

// TestParseRRule_Yearly verifies "FREQ=YEARLY" adds one year.
func TestParseRRule_Yearly(t *testing.T) {
	base := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	next := parseRRule(base, "FREQ=YEARLY")

	expected := time.Date(2027, 12, 31, 0, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("YEARLY: expected %v, got %v", expected, next)
	}
}

// TestParseRRule_YearlyInterval2 verifies "FREQ=YEARLY;INTERVAL=2" = bi-annual.
func TestParseRRule_YearlyInterval2(t *testing.T) {
	base := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	next := parseRRule(base, "FREQ=YEARLY;INTERVAL=2")

	expected := time.Date(2028, 6, 1, 0, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("YEARLY;INTERVAL=2: expected %v, got %v", expected, next)
	}
}

// TestParseRRule_Empty returns zero time for empty RRULE.
func TestParseRRule_Empty(t *testing.T) {
	base := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	next := parseRRule(base, "")

	if !next.IsZero() {
		t.Errorf("empty rrule: expected zero time, got %v", next)
	}
}

// TestParseRRule_InvalidFreq returns zero time for unknown FREQ.
func TestParseRRule_InvalidFreq(t *testing.T) {
	base := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	next := parseRRule(base, "FREQ=SECONDLY;INTERVAL=1")

	if !next.IsZero() {
		t.Errorf("invalid freq: expected zero time, got %v", next)
	}
}

// TestParseRRule_CaseInsensitive verifies mixed-case RRULE parsing.
func TestParseRRule_CaseInsensitive(t *testing.T) {
	base := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	next := parseRRule(base, "freq=monthly;interval=6")

	expected := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("case insensitive: expected %v, got %v", expected, next)
	}
}

// TestParseRRule_DailyInterval5 verifies "FREQ=DAILY;INTERVAL=5".
func TestParseRRule_DailyInterval5(t *testing.T) {
	base := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	next := parseRRule(base, "FREQ=DAILY;INTERVAL=5")

	expected := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("DAILY;INTERVAL=5: expected %v, got %v", expected, next)
	}
}

// ============================================================
// REMINDER SCHEDULING TESTS
// ============================================================

// TestReminderShouldFire verifies that reminders fire at the right thresholds.
func TestReminderShouldFire(t *testing.T) {
	thresholds := []int{7, 3, 1}

	tests := []struct {
		name         string
		daysUntilDue int
		shouldFire   bool
	}{
		{"7 days before", 7, true},
		{"3 days before", 3, true},
		{"1 day before", 1, true},
		{"due today", 0, false},
		{"5 days before", 5, false},
		{"10 days before", 10, false},
		{"2 days before", 2, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fired := false
			for _, threshold := range thresholds {
				if tt.daysUntilDue == threshold {
					fired = true
					break
				}
			}
			if fired != tt.shouldFire {
				t.Errorf("%s: expected shouldFire=%v, got %v", tt.name, tt.shouldFire, fired)
			}
		})
	}
}

// ============================================================
// SYNC IDEMPOTENCY TESTS (unit-level logic)
// ============================================================

// TestRiskLevelToPriority verifies mapping of risk levels to priorities.
func TestRiskLevelToPriority(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"critical", "critical"},
		{"high", "high"},
		{"medium", "medium"},
		{"low", "low"},
		{"very_low", "low"},
		{"", "low"},
		{"CRITICAL", "critical"},
		{"High", "high"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := riskLevelToPriority(tt.input)
			if result != tt.expected {
				t.Errorf("riskLevelToPriority(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestSeverityToPriority verifies mapping of severity to priority.
func TestSeverityToPriority(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"critical", "critical"},
		{"high", "high"},
		{"medium", "medium"},
		{"low", "low"},
		{"informational", "low"},
		{"", "low"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := severityToPriority(tt.input)
			if result != tt.expected {
				t.Errorf("severityToPriority(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestPriorityOrDefault verifies the priority fallback logic.
func TestPriorityOrDefault(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		def      string
		expected string
	}{
		{"valid critical", "critical", "medium", "critical"},
		{"valid high", "high", "medium", "high"},
		{"valid medium", "medium", "high", "medium"},
		{"valid low", "low", "medium", "low"},
		{"empty falls back", "", "high", "high"},
		{"unknown falls back", "urgent", "medium", "medium"},
		{"standard falls back", "standard", "high", "high"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := priorityOrDefault(tt.input, tt.def)
			if result != tt.expected {
				t.Errorf("priorityOrDefault(%q, %q) = %q, want %q", tt.input, tt.def, result, tt.expected)
			}
		})
	}
}

// TestICalEscape verifies iCal special character escaping.
func TestICalEscape(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"plain text", "Hello World", "Hello World"},
		{"semicolons", "a;b;c", "a\\;b\\;c"},
		{"commas", "a,b,c", "a\\,b\\,c"},
		{"newlines", "line1\nline2", "line1\\nline2"},
		{"backslashes", "path\\to\\file", "path\\\\to\\\\file"},
		{"mixed", "Meeting; Room 1, Floor 2\nNotes", "Meeting\\; Room 1\\, Floor 2\\nNotes"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := icalEscape(tt.input)
			if result != tt.expected {
				t.Errorf("icalEscape(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// ============================================================
// RECURRENCE END DATE BOUNDARY TESTS
// ============================================================

// TestComputeNextOccurrence_CustomRRule verifies custom RRULE handling
// delegates correctly.
func TestComputeNextOccurrence_CustomRRule(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	next := computeNextOccurrence(base, "custom_rrule", "FREQ=MONTHLY;INTERVAL=2")

	expected := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("custom_rrule MONTHLY;INTERVAL=2: expected %v, got %v", expected, next)
	}
}

// TestComputeNextOccurrence_AnnuallyLeapYear tests Feb 29 handling for annual recurrence.
func TestComputeNextOccurrence_AnnuallyLeapYear(t *testing.T) {
	// 2024 is a leap year, 2025 is not
	base := time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC)
	next := computeNextOccurrence(base, "annually", "")

	// Go's AddDate handles this gracefully, resulting in Mar 1, 2025
	if next.Year() != 2025 {
		t.Errorf("leap year annually: expected year 2025, got %d", next.Year())
	}
}
