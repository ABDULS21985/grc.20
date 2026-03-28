package service

import (
	"testing"
	"time"
)

// ============================================================
// DetermineEntityType Tests
// ============================================================

func TestDetermineEntityType_EssentialSector_LargeEntity(t *testing.T) {
	tests := []struct {
		name              string
		sector            string
		employeeCount     int
		annualTurnoverEUR float64
		expected          string
	}{
		{
			name:              "energy sector with 250+ employees is essential",
			sector:            "energy",
			employeeCount:     500,
			annualTurnoverEUR: 100_000_000,
			expected:          "essential",
		},
		{
			name:              "transport sector with 50M+ turnover is essential",
			sector:            "transport",
			employeeCount:     100,
			annualTurnoverEUR: 60_000_000,
			expected:          "essential",
		},
		{
			name:              "banking sector exactly at 250 employees is essential",
			sector:            "banking",
			employeeCount:     250,
			annualTurnoverEUR: 5_000_000,
			expected:          "essential",
		},
		{
			name:              "health sector exactly at 50M turnover is essential",
			sector:            "health",
			employeeCount:     50,
			annualTurnoverEUR: 50_000_000,
			expected:          "essential",
		},
		{
			name:              "digital_infrastructure large entity is essential",
			sector:            "digital_infrastructure",
			employeeCount:     300,
			annualTurnoverEUR: 200_000_000,
			expected:          "essential",
		},
		{
			name:              "space sector large entity is essential",
			sector:            "space",
			employeeCount:     1000,
			annualTurnoverEUR: 500_000_000,
			expected:          "essential",
		},
		{
			name:              "drinking_water large entity is essential",
			sector:            "drinking_water",
			employeeCount:     250,
			annualTurnoverEUR: 20_000_000,
			expected:          "essential",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetermineEntityType(tt.sector, tt.employeeCount, tt.annualTurnoverEUR)
			if result != tt.expected {
				t.Errorf("DetermineEntityType(%q, %d, %.0f) = %q, want %q",
					tt.sector, tt.employeeCount, tt.annualTurnoverEUR, result, tt.expected)
			}
		})
	}
}

func TestDetermineEntityType_EssentialSector_MediumEntity(t *testing.T) {
	tests := []struct {
		name              string
		sector            string
		employeeCount     int
		annualTurnoverEUR float64
		expected          string
	}{
		{
			name:              "energy medium entity (50-249 employees) is important",
			sector:            "energy",
			employeeCount:     100,
			annualTurnoverEUR: 20_000_000,
			expected:          "important",
		},
		{
			name:              "transport medium entity at lower bound is important",
			sector:            "transport",
			employeeCount:     50,
			annualTurnoverEUR: 5_000_000,
			expected:          "important",
		},
		{
			name:              "banking medium by turnover only is important",
			sector:            "banking",
			employeeCount:     30,
			annualTurnoverEUR: 15_000_000,
			expected:          "important",
		},
		{
			name:              "health at exactly 10M turnover is important",
			sector:            "health",
			employeeCount:     20,
			annualTurnoverEUR: 10_000_000,
			expected:          "important",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetermineEntityType(tt.sector, tt.employeeCount, tt.annualTurnoverEUR)
			if result != tt.expected {
				t.Errorf("DetermineEntityType(%q, %d, %.0f) = %q, want %q",
					tt.sector, tt.employeeCount, tt.annualTurnoverEUR, result, tt.expected)
			}
		})
	}
}

func TestDetermineEntityType_ImportantSector(t *testing.T) {
	tests := []struct {
		name              string
		sector            string
		employeeCount     int
		annualTurnoverEUR float64
		expected          string
	}{
		{
			name:              "manufacturing large entity is important (Annex II)",
			sector:            "manufacturing",
			employeeCount:     500,
			annualTurnoverEUR: 100_000_000,
			expected:          "important",
		},
		{
			name:              "food medium entity is important",
			sector:            "food",
			employeeCount:     75,
			annualTurnoverEUR: 25_000_000,
			expected:          "important",
		},
		{
			name:              "chemicals medium entity is important",
			sector:            "chemicals",
			employeeCount:     50,
			annualTurnoverEUR: 12_000_000,
			expected:          "important",
		},
		{
			name:              "digital_providers large entity is important",
			sector:            "digital_providers",
			employeeCount:     300,
			annualTurnoverEUR: 80_000_000,
			expected:          "important",
		},
		{
			name:              "postal_courier medium entity is important",
			sector:            "postal_courier",
			employeeCount:     60,
			annualTurnoverEUR: 11_000_000,
			expected:          "important",
		},
		{
			name:              "waste_management medium by turnover is important",
			sector:            "waste_management",
			employeeCount:     30,
			annualTurnoverEUR: 10_000_000,
			expected:          "important",
		},
		{
			name:              "research medium entity is important",
			sector:            "research",
			employeeCount:     80,
			annualTurnoverEUR: 15_000_000,
			expected:          "important",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetermineEntityType(tt.sector, tt.employeeCount, tt.annualTurnoverEUR)
			if result != tt.expected {
				t.Errorf("DetermineEntityType(%q, %d, %.0f) = %q, want %q",
					tt.sector, tt.employeeCount, tt.annualTurnoverEUR, result, tt.expected)
			}
		})
	}
}

func TestDetermineEntityType_NotApplicable(t *testing.T) {
	tests := []struct {
		name              string
		sector            string
		employeeCount     int
		annualTurnoverEUR float64
		expected          string
	}{
		{
			name:              "essential sector below thresholds is not applicable",
			sector:            "energy",
			employeeCount:     10,
			annualTurnoverEUR: 5_000_000,
			expected:          "not_applicable",
		},
		{
			name:              "important sector below thresholds is not applicable",
			sector:            "manufacturing",
			employeeCount:     20,
			annualTurnoverEUR: 5_000_000,
			expected:          "not_applicable",
		},
		{
			name:              "non-listed sector regardless of size is not applicable",
			sector:            "retail",
			employeeCount:     5000,
			annualTurnoverEUR: 1_000_000_000,
			expected:          "not_applicable",
		},
		{
			name:              "empty sector is not applicable",
			sector:            "",
			employeeCount:     500,
			annualTurnoverEUR: 100_000_000,
			expected:          "not_applicable",
		},
		{
			name:              "unknown sector is not applicable",
			sector:            "entertainment",
			employeeCount:     300,
			annualTurnoverEUR: 50_000_000,
			expected:          "not_applicable",
		},
		{
			name:              "zero employees and turnover is not applicable",
			sector:            "energy",
			employeeCount:     0,
			annualTurnoverEUR: 0,
			expected:          "not_applicable",
		},
		{
			name:              "important sector just below medium threshold",
			sector:            "food",
			employeeCount:     49,
			annualTurnoverEUR: 9_999_999,
			expected:          "not_applicable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetermineEntityType(tt.sector, tt.employeeCount, tt.annualTurnoverEUR)
			if result != tt.expected {
				t.Errorf("DetermineEntityType(%q, %d, %.0f) = %q, want %q",
					tt.sector, tt.employeeCount, tt.annualTurnoverEUR, result, tt.expected)
			}
		})
	}
}

func TestDetermineEntityType_CaseInsensitive(t *testing.T) {
	tests := []struct {
		name   string
		sector string
	}{
		{"uppercase", "ENERGY"},
		{"mixed case", "Energy"},
		{"with spaces", "  energy  "},
		{"mixed with spaces", "  Transport  "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetermineEntityType(tt.sector, 500, 100_000_000)
			if result != "essential" {
				t.Errorf("DetermineEntityType(%q, 500, 100000000) = %q, want 'essential' (case insensitive)",
					tt.sector, result)
			}
		})
	}
}

// ============================================================
// CalculateDeadlines Tests
// ============================================================

func TestCalculateDeadlines(t *testing.T) {
	detectionTime := time.Date(2026, 3, 15, 10, 0, 0, 0, time.UTC)

	ew, notif, final := CalculateDeadlines(detectionTime)

	// Phase 1: 24 hours
	expectedEW := detectionTime.Add(24 * time.Hour)
	if !ew.Equal(expectedEW) {
		t.Errorf("Early warning deadline = %v, want %v", ew, expectedEW)
	}

	// Phase 2: 72 hours
	expectedNotif := detectionTime.Add(72 * time.Hour)
	if !notif.Equal(expectedNotif) {
		t.Errorf("Notification deadline = %v, want %v", notif, expectedNotif)
	}

	// Phase 3: 1 month from notification deadline
	expectedFinal := expectedNotif.AddDate(0, 1, 0)
	if !final.Equal(expectedFinal) {
		t.Errorf("Final report deadline = %v, want %v", final, expectedFinal)
	}
}

func TestCalculateDeadlines_SpecificValues(t *testing.T) {
	// Test with a specific date to verify exact calculations
	detection := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	ew, notif, final := CalculateDeadlines(detection)

	// 24 hours later
	if ew.Day() != 2 || ew.Hour() != 0 {
		t.Errorf("Early warning: expected Jan 2 00:00, got %v", ew)
	}

	// 72 hours later
	if notif.Day() != 4 || notif.Hour() != 0 {
		t.Errorf("Notification: expected Jan 4 00:00, got %v", notif)
	}

	// 1 month from Jan 4 = Feb 4
	if final.Month() != time.February || final.Day() != 4 {
		t.Errorf("Final report: expected Feb 4, got %v", final)
	}
}

func TestCalculateDeadlines_MonthBoundary(t *testing.T) {
	// Detection at end of January -- tests month boundary
	detection := time.Date(2026, 1, 29, 14, 30, 0, 0, time.UTC)

	ew, notif, final := CalculateDeadlines(detection)

	// EW: 24h = Jan 30 14:30
	if ew.Month() != time.January || ew.Day() != 30 {
		t.Errorf("Early warning: expected Jan 30, got %v", ew)
	}

	// Notification: 72h = Feb 1 14:30
	if notif.Month() != time.February || notif.Day() != 1 {
		t.Errorf("Notification: expected Feb 1, got %v", notif)
	}

	// Final: 1 month from Feb 1 = Mar 1
	if final.Month() != time.March || final.Day() != 1 {
		t.Errorf("Final report: expected Mar 1, got %v", final)
	}
}

func TestCalculateDeadlines_OrderPreserved(t *testing.T) {
	detection := time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC)

	ew, notif, final := CalculateDeadlines(detection)

	if !ew.Before(notif) {
		t.Error("Early warning deadline should be before notification deadline")
	}
	if !notif.Before(final) {
		t.Error("Notification deadline should be before final report deadline")
	}
}

// ============================================================
// Sector Classification Tests
// ============================================================

func TestAllEssentialSectors(t *testing.T) {
	sectors := []string{
		"energy", "transport", "banking", "financial_market",
		"health", "drinking_water", "waste_water",
		"digital_infrastructure", "ict_service_management",
		"public_administration", "space",
	}

	for _, sector := range sectors {
		t.Run(sector, func(t *testing.T) {
			result := DetermineEntityType(sector, 500, 100_000_000)
			if result != "essential" {
				t.Errorf("Sector %q with 500 employees and 100M turnover should be essential, got %q",
					sector, result)
			}
		})
	}
}

func TestAllImportantSectors(t *testing.T) {
	sectors := []string{
		"postal_courier", "waste_management", "chemicals",
		"food", "manufacturing", "digital_providers", "research",
	}

	for _, sector := range sectors {
		t.Run(sector, func(t *testing.T) {
			// Large Annex II entities are important (not essential)
			result := DetermineEntityType(sector, 500, 100_000_000)
			if result != "important" {
				t.Errorf("Sector %q (Annex II) with 500 employees should be important, got %q",
					sector, result)
			}
		})
	}
}

// ============================================================
// Boundary Tests
// ============================================================

func TestEntityType_BoundaryValues(t *testing.T) {
	tests := []struct {
		name     string
		sector   string
		emp      int
		turnover float64
		expected string
	}{
		{"essential sector, 249 employees, 49.99M", "energy", 249, 49_999_999, "important"},
		{"essential sector, 250 employees, 0 turnover", "energy", 250, 0, "essential"},
		{"essential sector, 0 employees, 50M turnover", "energy", 0, 50_000_000, "essential"},
		{"essential sector, 49 employees, 9.99M", "energy", 49, 9_999_999, "not_applicable"},
		{"essential sector, 50 employees, 0 turnover", "energy", 50, 0, "important"},
		{"essential sector, 0 employees, 10M turnover", "energy", 0, 10_000_000, "important"},
		{"important sector, 249 employees, 49.99M", "manufacturing", 249, 49_999_999, "important"},
		{"important sector, 49 employees, 9.99M", "manufacturing", 49, 9_999_999, "not_applicable"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetermineEntityType(tt.sector, tt.emp, tt.turnover)
			if result != tt.expected {
				t.Errorf("DetermineEntityType(%q, %d, %.0f) = %q, want %q",
					tt.sector, tt.emp, tt.turnover, result, tt.expected)
			}
		})
	}
}
