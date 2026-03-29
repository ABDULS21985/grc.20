package service

import (
	"testing"
)

// ============================================================
// SEARCH ENGINE TESTS
// Unit tests for query parsing, snippet generation, slugifying,
// reading time estimation, and tsquery building.
// ============================================================

// ── BuildTSQuery Tests ──────────────────────────────────────

func TestBuildTSQuery_SimpleWord(t *testing.T) {
	result := BuildTSQuery("firewall")
	if result == "" {
		t.Fatal("expected non-empty tsquery for simple word")
	}
	if !contains(result, "firewall") {
		t.Errorf("expected tsquery to contain 'firewall', got %q", result)
	}
}

func TestBuildTSQuery_MultipleWords(t *testing.T) {
	result := BuildTSQuery("access control policy")
	if result == "" {
		t.Fatal("expected non-empty tsquery")
	}
	// Should join with &
	if !contains(result, "&") {
		t.Errorf("expected '&' conjunction in %q", result)
	}
}

func TestBuildTSQuery_QuotedPhrase(t *testing.T) {
	result := BuildTSQuery(`"access control"`)
	if result == "" {
		t.Fatal("expected non-empty tsquery for quoted phrase")
	}
	// Should use <-> for phrase matching
	if !contains(result, "<->") {
		t.Errorf("expected '<->' phrase operator in %q", result)
	}
}

func TestBuildTSQuery_OROperator(t *testing.T) {
	result := BuildTSQuery("encryption OR cryptography")
	if result == "" {
		t.Fatal("expected non-empty tsquery")
	}
	if !contains(result, "|") {
		t.Errorf("expected '|' OR operator in %q", result)
	}
}

func TestBuildTSQuery_Negation(t *testing.T) {
	result := BuildTSQuery("policy -obsolete")
	if result == "" {
		t.Fatal("expected non-empty tsquery")
	}
	if !contains(result, "!") {
		t.Errorf("expected '!' negation in %q", result)
	}
}

func TestBuildTSQuery_Empty(t *testing.T) {
	result := BuildTSQuery("")
	if result != "" {
		t.Errorf("expected empty tsquery for empty input, got %q", result)
	}
}

func TestBuildTSQuery_WhitespaceOnly(t *testing.T) {
	result := BuildTSQuery("   ")
	if result != "" {
		t.Errorf("expected empty tsquery for whitespace, got %q", result)
	}
}

func TestBuildTSQuery_MixedQuotedAndFree(t *testing.T) {
	result := BuildTSQuery(`"risk matrix" assessment`)
	if result == "" {
		t.Fatal("expected non-empty tsquery")
	}
	if !contains(result, "<->") {
		t.Errorf("expected phrase operator in %q", result)
	}
	if !contains(result, "&") {
		t.Errorf("expected conjunction operator in %q", result)
	}
}

func TestBuildTSQuery_PrefixMatch(t *testing.T) {
	result := BuildTSQuery("comp")
	if !contains(result, ":*") {
		t.Errorf("expected prefix match operator ':*' in %q", result)
	}
}

// ── Snippet Generation Tests ────────────────────────────────

func TestGenerateSnippet_ShortBody(t *testing.T) {
	body := "A short text."
	snippet := generateSnippet(body, "short", 200)
	if snippet != body {
		t.Errorf("expected full body for short text, got %q", snippet)
	}
}

func TestGenerateSnippet_LongBody_CenteredOnQuery(t *testing.T) {
	body := "Lorem ipsum dolor sit amet, consectetur adipiscing elit. " +
		"The firewall configuration must be reviewed quarterly. " +
		"Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua."
	snippet := generateSnippet(body, "firewall", 80)
	if len(snippet) == 0 {
		t.Fatal("expected non-empty snippet")
	}
	// Snippet should contain the query term
	if !contains(snippet, "firewall") {
		t.Errorf("expected snippet to contain 'firewall', got %q", snippet)
	}
}

func TestGenerateSnippet_EmptyBody(t *testing.T) {
	snippet := generateSnippet("", "test", 200)
	if snippet != "" {
		t.Errorf("expected empty snippet for empty body, got %q", snippet)
	}
}

func TestGenerateSnippet_QueryNotFound(t *testing.T) {
	body := "This is a long text that does not contain the search term anywhere in its content padding padding padding."
	snippet := generateSnippet(body, "zzzzz", 40)
	if len(snippet) == 0 {
		t.Fatal("expected non-empty snippet even when query not found")
	}
}

// ── Slugify Tests ───────────────────────────────────────────

func TestSlugify_Basic(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello World", "hello-world"},
		{"GDPR Article 30 ROPA", "gdpr-article-30-ropa"},
		{"ISO 27001 — Annex A.5", "iso-27001-annex-a5"},
		{"  Leading and Trailing  ", "leading-and-trailing"},
		{"Special!@#$Characters", "specialcharacters"},
		{"Multiple---Hyphens", "multiple-hyphens"},
		{"already-slugified", "already-slugified"},
		{"", ""},
	}

	for _, tt := range tests {
		got := slugify(tt.input)
		if got != tt.expected {
			t.Errorf("slugify(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

// ── Reading Time Estimation Tests ───────────────────────────

func TestEstimateReadingTime_Short(t *testing.T) {
	content := "Just a few words here."
	got := estimateReadingTime(content)
	if got != 1 {
		t.Errorf("expected 1 minute for short content, got %d", got)
	}
}

func TestEstimateReadingTime_Medium(t *testing.T) {
	// ~600 words = 3 minutes at 200 WPM
	words := make([]byte, 0, 4000)
	for i := 0; i < 600; i++ {
		words = append(words, []byte("word ")...)
	}
	got := estimateReadingTime(string(words))
	if got != 3 {
		t.Errorf("expected 3 minutes for 600 words, got %d", got)
	}
}

func TestEstimateReadingTime_Long(t *testing.T) {
	// ~2000 words = 10 minutes
	words := make([]byte, 0, 12000)
	for i := 0; i < 2000; i++ {
		words = append(words, []byte("word ")...)
	}
	got := estimateReadingTime(string(words))
	if got != 10 {
		t.Errorf("expected 10 minutes for 2000 words, got %d", got)
	}
}

// ── SearchRequest Defaults ──────────────────────────────────

func TestSearchRequest_Defaults(t *testing.T) {
	req := SearchRequest{}
	if req.Page != 0 {
		t.Errorf("default page should be 0 (handled by Search), got %d", req.Page)
	}
	if req.SortBy != "" {
		t.Errorf("default sort should be empty (handled by Search), got %q", req.SortBy)
	}
}

func TestSearchRequest_EntityTypes(t *testing.T) {
	req := SearchRequest{
		EntityTypes: []string{"risk", "policy", "control"},
	}
	if len(req.EntityTypes) != 3 {
		t.Errorf("expected 3 entity types, got %d", len(req.EntityTypes))
	}
}

// ── FacetBucket ─────────────────────────────────────────────

func TestFacetBucket_Struct(t *testing.T) {
	b := FacetBucket{Value: "risk", Count: 42}
	if b.Value != "risk" || b.Count != 42 {
		t.Errorf("unexpected facet bucket: %+v", b)
	}
}

// ── IndexRecord ─────────────────────────────────────────────

func TestIndexRecord_NilDefaults(t *testing.T) {
	rec := IndexRecord{
		EntityType: "risk",
		Title:      "Test Risk",
	}
	if rec.Tags != nil {
		t.Errorf("expected nil tags by default")
	}
	if rec.FrameworkCodes != nil {
		t.Errorf("expected nil framework codes by default")
	}
}

// ── AutocompleteResult ──────────────────────────────────────

func TestAutocompleteResult_Fields(t *testing.T) {
	r := AutocompleteResult{
		EntityType: "policy",
		Title:      "Data Protection Policy",
		EntityRef:  "POL-0001",
		Status:     "active",
	}
	if r.EntityType != "policy" || r.Title != "Data Protection Policy" {
		t.Errorf("unexpected autocomplete result: %+v", r)
	}
}

// ── Helper ──────────────────────────────────────────────────

func contains(s, substr string) bool {
	return len(s) >= len(substr) && containsSubstr(s, substr)
}

func containsSubstr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
