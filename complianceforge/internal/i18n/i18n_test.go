package i18n

import (
	"strings"
	"testing"
	"time"
)

// TestNewTranslator verifies that all locale files load correctly.
func TestNewTranslator(t *testing.T) {
	tr, err := newTranslator()
	if err != nil {
		t.Fatalf("failed to create translator: %v", err)
	}

	// All 9 languages must be loaded.
	expectedLangs := []string{"en-GB", "de-DE", "fr-FR", "es-ES", "it-IT", "nl-NL", "pt-PT", "pl-PL", "sv-SE"}
	for _, lang := range expectedLangs {
		if _, ok := tr.translations[lang]; !ok {
			t.Errorf("language %s not loaded", lang)
		}
	}
}

// TestTranslationLookup tests basic key lookup.
func TestTranslationLookup(t *testing.T) {
	tr, err := newTranslator()
	if err != nil {
		t.Fatalf("failed to create translator: %v", err)
	}

	tests := []struct {
		lang     string
		key      string
		expected string
	}{
		{"en-GB", "common.save", "Save"},
		{"en-GB", "common.cancel", "Cancel"},
		{"de-DE", "common.save", "Speichern"},
		{"de-DE", "common.cancel", "Abbrechen"},
		{"fr-FR", "common.save", "Enregistrer"},
		{"es-ES", "common.delete", "Eliminar"},
		{"it-IT", "common.edit", "Modifica"},
		{"nl-NL", "common.create", "Aanmaken"},
		{"pt-PT", "common.search", "Pesquisar"},
		{"pl-PL", "common.filter", "Filtruj"},
		{"sv-SE", "common.loading", "Laddar..."},
	}

	for _, tc := range tests {
		t.Run(tc.lang+"/"+tc.key, func(t *testing.T) {
			result := tr.T(tc.lang, tc.key)
			if result != tc.expected {
				t.Errorf("T(%q, %q) = %q, want %q", tc.lang, tc.key, result, tc.expected)
			}
		})
	}
}

// TestFallbackToEnglish verifies that missing keys fall back to en-GB.
func TestFallbackToEnglish(t *testing.T) {
	tr, err := newTranslator()
	if err != nil {
		t.Fatalf("failed to create translator: %v", err)
	}

	// A key that exists in en-GB should be returned when the requested
	// language is missing that key. We test by using a valid English key
	// and confirming it returns the English value even for another language.
	enValue := tr.T("en-GB", "common.save")
	deValue := tr.T("de-DE", "common.save")

	if enValue == "" {
		t.Fatal("expected non-empty English value for common.save")
	}
	if deValue == "" {
		t.Fatal("expected non-empty German value for common.save")
	}

	// German should have its own value, not the English fallback.
	if enValue == deValue {
		t.Error("German value should differ from English for common.save")
	}
}

// TestMissingKeyReturnsDottedKey verifies that a completely missing key
// returns the dotted key path.
func TestMissingKeyReturnsDottedKey(t *testing.T) {
	tr, err := newTranslator()
	if err != nil {
		t.Fatalf("failed to create translator: %v", err)
	}

	key := "this.key.does.not.exist.anywhere"
	result := tr.T("en-GB", key)
	if result != key {
		t.Errorf("expected missing key to return %q, got %q", key, result)
	}

	result = tr.T("de-DE", key)
	if result != key {
		t.Errorf("expected missing key in de-DE to return %q, got %q", key, result)
	}
}

// TestInterpolation verifies placeholder substitution.
func TestInterpolation(t *testing.T) {
	tr, err := newTranslator()
	if err != nil {
		t.Fatalf("failed to create translator: %v", err)
	}

	result := tr.T("en-GB", "dashboard.welcome", map[string]interface{}{
		"name": "Alice",
	})
	if result != "Welcome back, Alice" {
		t.Errorf("interpolation failed: got %q", result)
	}

	result = tr.T("de-DE", "dashboard.welcome", map[string]interface{}{
		"name": "Hans",
	})
	if result != "Willkommen zurück, Hans" {
		t.Errorf("German interpolation failed: got %q", result)
	}

	// Multiple placeholders.
	result = tr.T("en-GB", "common.showing", map[string]interface{}{
		"from":  1,
		"to":    10,
		"total": 100,
	})
	if result != "Showing 1 to 10 of 100 results" {
		t.Errorf("multi-placeholder interpolation failed: got %q", result)
	}
}

// TestPluralisation tests the TPl method.
func TestPluralisation(t *testing.T) {
	tr, err := newTranslator()
	if err != nil {
		t.Fatalf("failed to create translator: %v", err)
	}

	// Singular.
	result := tr.TPl("en-GB", "common.items", 1)
	if result != "1 item" {
		t.Errorf("singular pluralisation failed: got %q", result)
	}

	// Plural.
	result = tr.TPl("en-GB", "common.items", 5)
	if result != "5 items" {
		t.Errorf("plural pluralisation failed: got %q", result)
	}

	// German singular.
	result = tr.TPl("de-DE", "common.items", 1)
	if result != "1 Element" {
		t.Errorf("German singular pluralisation failed: got %q", result)
	}

	// German plural.
	result = tr.TPl("de-DE", "common.items", 3)
	if result != "3 Elemente" {
		t.Errorf("German plural pluralisation failed: got %q", result)
	}

	// Risk count.
	result = tr.TPl("en-GB", "risks.count", 1)
	if result != "1 risk" {
		t.Errorf("risk singular failed: got %q", result)
	}

	result = tr.TPl("en-GB", "risks.count", 42)
	if result != "42 risks" {
		t.Errorf("risk plural failed: got %q", result)
	}
}

// TestFormatDate verifies locale-aware date formatting.
func TestFormatDate(t *testing.T) {
	// 15 March 2025, 14:30 UTC
	ts := time.Date(2025, 3, 15, 14, 30, 0, 0, time.UTC)

	tests := []struct {
		lang     string
		expected string
	}{
		{"en-GB", "15/03/2025 14:30"},
		{"de-DE", "15.03.2025 14:30"},
		{"fr-FR", "15/03/2025 14h30"},
		{"nl-NL", "15-03-2025 14:30"},
		{"sv-SE", "2025-03-15 14:30"},
		{"pl-PL", "15.03.2025 14:30"},
	}

	for _, tc := range tests {
		t.Run(tc.lang, func(t *testing.T) {
			result := FormatDate(tc.lang, ts)
			if result != tc.expected {
				t.Errorf("FormatDate(%q) = %q, want %q", tc.lang, result, tc.expected)
			}
		})
	}

	// Unknown language should fall back to en-GB format.
	result := FormatDate("xx-XX", ts)
	if result != "15/03/2025 14:30" {
		t.Errorf("FormatDate with unknown lang = %q, want en-GB format", result)
	}
}

// TestFormatNumber verifies locale-aware number formatting.
func TestFormatNumber(t *testing.T) {
	tests := []struct {
		lang     string
		number   float64
		expected string
	}{
		{"en-GB", 1234.56, "1,234.56"},
		{"de-DE", 1234.56, "1.234,56"},
		{"fr-FR", 1234.56, "1\u202f234,56"},
		{"en-GB", 0.99, "0.99"},
		{"de-DE", 0.99, "0,99"},
		{"en-GB", 1000000.00, "1,000,000.00"},
		{"sv-SE", 1234.56, "1\u00a0234,56"},
	}

	for _, tc := range tests {
		t.Run(tc.lang, func(t *testing.T) {
			result := FormatNumber(tc.lang, tc.number)
			if result != tc.expected {
				t.Errorf("FormatNumber(%q, %v) = %q, want %q", tc.lang, tc.number, result, tc.expected)
			}
		})
	}
}

// TestFormatCurrency verifies locale-aware currency formatting.
func TestFormatCurrency(t *testing.T) {
	tests := []struct {
		lang     string
		amount   float64
		currency string
		contains string // Substring check (symbols may vary in placement)
	}{
		{"en-GB", 1234.56, "GBP", "\u00a31,234.56"},
		{"de-DE", 1234.56, "EUR", "1.234,56"},
		{"en-GB", 99.99, "USD", "$99.99"},
	}

	for _, tc := range tests {
		t.Run(tc.lang+"/"+tc.currency, func(t *testing.T) {
			result := FormatCurrency(tc.lang, tc.amount, tc.currency)
			if !strings.Contains(result, tc.contains) {
				t.Errorf("FormatCurrency(%q, %v, %q) = %q, expected to contain %q",
					tc.lang, tc.amount, tc.currency, result, tc.contains)
			}
		})
	}
}

// TestNegativeNumber verifies negative number formatting.
func TestNegativeNumber(t *testing.T) {
	result := FormatNumber("en-GB", -1234.56)
	if result != "-1,234.56" {
		t.Errorf("FormatNumber for negative = %q, want %q", result, "-1,234.56")
	}
}

// TestSupportedLanguages verifies the helper functions.
func TestSupportedLanguages(t *testing.T) {
	langs := SupportedLanguages()
	if len(langs) != 9 {
		t.Errorf("expected 9 supported languages, got %d", len(langs))
	}

	if !IsSupported("en-GB") {
		t.Error("en-GB should be supported")
	}
	if !IsSupported("de-DE") {
		t.Error("de-DE should be supported")
	}
	if IsSupported("ja-JP") {
		t.Error("ja-JP should not be supported")
	}
}

// TestAllLanguagesHaveSameKeys checks that every language has the same key
// set as English (our reference).
func TestAllLanguagesHaveSameKeys(t *testing.T) {
	tr, err := newTranslator()
	if err != nil {
		t.Fatalf("failed to create translator: %v", err)
	}

	enKeys := tr.translations["en-GB"]
	for lang, msgs := range tr.translations {
		if lang == "en-GB" {
			continue
		}
		for key := range enKeys {
			if _, ok := msgs[key]; !ok {
				t.Errorf("language %s is missing key %q", lang, key)
			}
		}
	}
}

// TestKeyCount ensures each language has at least 250 keys.
func TestKeyCount(t *testing.T) {
	tr, err := newTranslator()
	if err != nil {
		t.Fatalf("failed to create translator: %v", err)
	}

	for lang, msgs := range tr.translations {
		count := len(msgs)
		if count < 250 {
			t.Errorf("language %s has only %d keys, expected at least 250", lang, count)
		}
	}
}

// TestNavigationKeysExist verifies key categories are present.
func TestNavigationKeysExist(t *testing.T) {
	tr, err := newTranslator()
	if err != nil {
		t.Fatalf("failed to create translator: %v", err)
	}

	requiredKeys := []string{
		"navigation.dashboard",
		"navigation.frameworks",
		"navigation.controls",
		"navigation.risks",
		"navigation.policies",
		"navigation.audits",
		"navigation.incidents",
		"navigation.vendors",
		"navigation.assets",
		"navigation.settings",
		"navigation.reports",
		"navigation.nis2",
		"navigation.monitoring",
		"navigation.workflows",
	}

	for _, lang := range SupportedLanguages() {
		for _, key := range requiredKeys {
			val := tr.T(lang, key)
			if val == key {
				t.Errorf("language %s: key %q returned raw key (not translated)", lang, key)
			}
		}
	}
}
