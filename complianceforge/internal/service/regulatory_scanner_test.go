package service

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// ============================================================
// RSS PARSING TESTS
// ============================================================

// TestParseRSSFeed_RSS2 verifies that a standard RSS 2.0 feed is correctly
// parsed into normalised RSSEntry values.
func TestParseRSSFeed_RSS2(t *testing.T) {
	feed := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Test Regulatory Feed</title>
    <link>https://example.com</link>
    <description>Test feed</description>
    <item>
      <title>New GDPR Enforcement Decision</title>
      <link>https://example.com/gdpr-decision-1</link>
      <description>Major enforcement action regarding data breach notification</description>
      <pubDate>Mon, 15 Jan 2026 10:00:00 +0000</pubDate>
      <guid>guid-001</guid>
      <category>enforcement</category>
    </item>
    <item>
      <title>Updated ISO 27001 Guidance</title>
      <link>https://example.com/iso-guidance</link>
      <description>New guidance on implementing Annex A controls</description>
      <pubDate>Tue, 16 Jan 2026 14:30:00 +0000</pubDate>
      <guid>guid-002</guid>
      <category>guidance</category>
    </item>
    <item>
      <title>NIS2 Implementation Deadline Reminder</title>
      <link>https://example.com/nis2-deadline</link>
      <description>Critical: All EU member states must transpose NIS2 by October 2024</description>
      <pubDate>Wed, 17 Jan 2026 09:00:00 +0000</pubDate>
      <guid>guid-003</guid>
    </item>
  </channel>
</rss>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, feed)
	}))
	defer server.Close()

	scanner := NewRegulatoryScanner(nil)
	entries, err := scanner.ParseRSSFeed(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}

	// Verify first entry
	if entries[0].Title != "New GDPR Enforcement Decision" {
		t.Errorf("expected title 'New GDPR Enforcement Decision', got '%s'", entries[0].Title)
	}
	if entries[0].Link != "https://example.com/gdpr-decision-1" {
		t.Errorf("expected link 'https://example.com/gdpr-decision-1', got '%s'", entries[0].Link)
	}
	if entries[0].GUID != "guid-001" {
		t.Errorf("expected GUID 'guid-001', got '%s'", entries[0].GUID)
	}

	// Verify published date is parsed correctly
	expected := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	if !entries[0].PublishedAt.Equal(expected) {
		t.Errorf("expected published at %v, got %v", expected, entries[0].PublishedAt)
	}
}

// TestParseRSSFeed_Atom verifies that an Atom feed is correctly parsed.
func TestParseRSSFeed_Atom(t *testing.T) {
	feed := `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <title>Regulatory Authority Updates</title>
  <entry>
    <title>New Cybersecurity Framework Published</title>
    <link href="https://example.com/cyber-framework" rel="alternate"/>
    <summary>A comprehensive new cybersecurity framework for critical infrastructure</summary>
    <updated>2026-02-10T08:00:00Z</updated>
    <id>urn:uuid:atom-entry-001</id>
  </entry>
  <entry>
    <title>Data Protection Impact Assessment Guidelines</title>
    <link href="https://example.com/dpia-guide" rel="alternate"/>
    <content>Detailed guidelines for conducting DPIAs under GDPR Article 35</content>
    <updated>2026-02-12T14:00:00Z</updated>
    <id>urn:uuid:atom-entry-002</id>
  </entry>
</feed>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/atom+xml")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, feed)
	}))
	defer server.Close()

	scanner := NewRegulatoryScanner(nil)
	entries, err := scanner.ParseRSSFeed(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	if entries[0].Title != "New Cybersecurity Framework Published" {
		t.Errorf("expected first entry title 'New Cybersecurity Framework Published', got '%s'", entries[0].Title)
	}
	if entries[0].Link != "https://example.com/cyber-framework" {
		t.Errorf("expected link 'https://example.com/cyber-framework', got '%s'", entries[0].Link)
	}

	// Second entry should use <content> as description since <summary> is absent
	if entries[1].Description != "Detailed guidelines for conducting DPIAs under GDPR Article 35" {
		t.Errorf("expected description from <content>, got '%s'", entries[1].Description)
	}
}

// TestParseRSSFeed_HTTPError verifies correct error handling for non-200 responses.
func TestParseRSSFeed_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	scanner := NewRegulatoryScanner(nil)
	_, err := scanner.ParseRSSFeed(context.Background(), server.URL)
	if err == nil {
		t.Fatal("expected error for non-200 response, got nil")
	}
}

// TestParseRSSFeed_InvalidXML verifies correct error handling for invalid XML.
func TestParseRSSFeed_InvalidXML(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "this is not valid XML at all")
	}))
	defer server.Close()

	scanner := NewRegulatoryScanner(nil)
	_, err := scanner.ParseRSSFeed(context.Background(), server.URL)
	if err == nil {
		t.Fatal("expected error for invalid XML, got nil")
	}
}

// TestParseRSSFeed_EmptyFeed verifies handling of a valid but empty feed.
func TestParseRSSFeed_EmptyFeed(t *testing.T) {
	feed := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Empty Feed</title>
    <link>https://example.com</link>
  </channel>
</rss>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, feed)
	}))
	defer server.Close()

	scanner := NewRegulatoryScanner(nil)
	_, err := scanner.ParseRSSFeed(context.Background(), server.URL)
	// An empty feed should fail to parse as both RSS and Atom
	if err == nil {
		t.Fatal("expected error for empty feed, got nil")
	}
}

// ============================================================
// RSS/ATOM XML NORMALISATION TESTS
// ============================================================

// TestNormaliseRSSItems verifies that RSS items are correctly normalised.
func TestNormaliseRSSItems(t *testing.T) {
	items := []RSSItem{
		{
			Title:       "  Test Item  ",
			Link:        "  https://example.com/test  ",
			Description: "A test item description",
			PubDate:     "Mon, 01 Jan 2026 00:00:00 +0000",
			GUID:        "guid-test",
			Category:    "testing",
		},
	}

	entries := normaliseRSSItems(items)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	if entries[0].Title != "Test Item" {
		t.Errorf("expected trimmed title 'Test Item', got '%s'", entries[0].Title)
	}
	if entries[0].Link != "https://example.com/test" {
		t.Errorf("expected trimmed link, got '%s'", entries[0].Link)
	}
	if entries[0].GUID != "guid-test" {
		t.Errorf("expected GUID 'guid-test', got '%s'", entries[0].GUID)
	}
	if entries[0].Category != "testing" {
		t.Errorf("expected category 'testing', got '%s'", entries[0].Category)
	}
}

// TestNormaliseAtomEntries verifies that Atom entries are correctly normalised.
func TestNormaliseAtomEntries(t *testing.T) {
	entries := normaliseAtomEntries([]AtomEntry{
		{
			Title:   "Atom Entry",
			Link:    AtomLink{Href: "https://example.com/atom", Rel: "alternate"},
			Summary: "Summary text",
			Updated: "2026-03-15T12:00:00Z",
			ID:      "urn:test:1",
		},
		{
			Title:   "Atom Entry No Summary",
			Link:    AtomLink{Href: "https://example.com/atom2"},
			Content: "Content fallback",
			Updated: "2026-03-16T12:00:00Z",
			ID:      "urn:test:2",
		},
	})

	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	if entries[0].Description != "Summary text" {
		t.Errorf("expected description 'Summary text', got '%s'", entries[0].Description)
	}
	if entries[1].Description != "Content fallback" {
		t.Errorf("expected content fallback as description, got '%s'", entries[1].Description)
	}
}

// ============================================================
// DATE PARSING TESTS
// ============================================================

// TestParseFlexibleDate verifies parsing of various date formats used in RSS/Atom feeds.
func TestParseFlexibleDate(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		year   int
		month  time.Month
		day    int
	}{
		{"RFC1123Z", "Mon, 02 Jan 2006 15:04:05 -0700", 2006, time.January, 2},
		{"RFC3339", "2026-01-15T10:00:00Z", 2026, time.January, 15},
		{"ISO8601 offset", "2026-06-20T14:30:00+02:00", 2026, time.June, 20},
		{"Date only", "2026-03-28", 2026, time.March, 28},
		{"RFC1123", "Tue, 16 Jan 2026 14:30:00 UTC", 2026, time.January, 16},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseFlexibleDate(tt.input)
			if result.Year() != tt.year {
				t.Errorf("expected year %d, got %d", tt.year, result.Year())
			}
			if result.Month() != tt.month {
				t.Errorf("expected month %v, got %v", tt.month, result.Month())
			}
			if result.Day() != tt.day {
				t.Errorf("expected day %d, got %d", tt.day, result.Day())
			}
		})
	}
}

// TestParseFlexibleDate_Empty verifies that an empty string returns current time.
func TestParseFlexibleDate_Empty(t *testing.T) {
	before := time.Now().UTC().Add(-1 * time.Second)
	result := parseFlexibleDate("")
	after := time.Now().UTC().Add(1 * time.Second)

	if result.Before(before) || result.After(after) {
		t.Errorf("expected current time for empty input, got %v", result)
	}
}

// ============================================================
// CLASSIFICATION TESTS
// ============================================================

// TestClassifyChangeType verifies keyword-based change type classification.
func TestClassifyChangeType(t *testing.T) {
	tests := []struct {
		text     string
		expected string
	}{
		{"Major enforcement action and fine imposed by ICO", "enforcement_decision"},
		{"New regulation on digital markets adopted by EU", "new_regulation"},
		{"Amendment to the Data Protection Act 2018", "amendment"},
		{"Public consultation on AI governance framework", "consultation"},
		{"Updated guidance on cloud security best practices", "guidance"},
		{"Court ruling on data subject access rights", "court_ruling"},
		{"ISO standard revision for 27001:2022", "standard_revision"},
		{"PCI DSS standard update v4.0.1 released", "standard_update"},
		{"Industry bulletin on emerging cyber threats", "industry_bulletin"},
		{"Regular news about technology trends", "guidance"},
	}

	for _, tt := range tests {
		t.Run(tt.text[:30], func(t *testing.T) {
			result := classifyChangeType(tt.text)
			if result != tt.expected {
				t.Errorf("classifyChangeType(%q) = %q, want %q", tt.text, result, tt.expected)
			}
		})
	}
}

// TestClassifySeverity verifies keyword-based severity classification.
func TestClassifySeverity(t *testing.T) {
	tests := []struct {
		text     string
		expected string
	}{
		{"Critical zero-day vulnerability requires immediate action", "critical"},
		{"Emergency cybersecurity advisory: mandatory patching", "critical"},
		{"Significant enforcement fine for non-compliance", "high"},
		{"Mandatory requirement with upcoming deadline", "high"},
		{"Standard amendment and revision to existing controls", "medium"},
		{"Minor informational newsletter summary", "low"},
		{"Regular organisational meeting notes", "informational"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := classifySeverity(tt.text)
			if result != tt.expected {
				t.Errorf("classifySeverity(%q) = %q, want %q", tt.text, result, tt.expected)
			}
		})
	}
}

// TestDetectRegions verifies keyword-based region detection.
func TestDetectRegions(t *testing.T) {
	tests := []struct {
		text     string
		expected []string
	}{
		{"UK ICO publishes new guidance for united kingdom organisations", []string{"GB"}},
		{"European Union directive affects EU member states", []string{"EU"}},
		{"Germany and france adopt joint cybersecurity strategy", []string{"DE", "FR"}},
		{"Global international standard update worldwide", []string{"GLOBAL"}},
		{"No region mentioned", nil},
	}

	for _, tt := range tests {
		name := tt.text
		if len(name) > 20 {
			name = name[:20]
		}
		t.Run(name, func(t *testing.T) {
			result := detectRegions(tt.text)
			if tt.expected == nil && len(result) == 0 {
				return // Both empty, OK
			}
			for _, exp := range tt.expected {
				found := false
				for _, r := range result {
					if r == exp {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("detectRegions(%q): expected region %q in result %v", tt.text, exp, result)
				}
			}
		})
	}
}

// TestDetectTags verifies keyword-based tag detection.
func TestDetectTags(t *testing.T) {
	tests := []struct {
		text     string
		expected string
	}{
		{"New GDPR enforcement decision", "GDPR"},
		{"ISO 27001 certification update", "ISO27001"},
		{"NIST cybersecurity framework revision", "NIST"},
		{"PCI DSS compliance requirements", "PCI-DSS"},
		{"NIS2 directive implementation", "NIS2"},
		{"Ransomware attack advisory", "ransomware"},
		{"Supply chain security guidance", "supply-chain"},
		{"Cloud security best practices", "cloud"},
		{"Artificial intelligence governance", "artificial-intelligence"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			tags := detectTags(tt.text)
			found := false
			for _, tag := range tags {
				if tag == tt.expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("detectTags(%q): expected tag %q in %v", tt.text, tt.expected, tags)
			}
		})
	}
}

// ============================================================
// CLASSIFICATION INTEGRATION TEST
// ============================================================

// TestClassifyChange_Integration verifies that ClassifyChange correctly
// populates all classification fields on a RegulatoryChange struct.
func TestClassifyChange_Integration(t *testing.T) {
	scanner := NewRegulatoryScanner(nil)

	change := &RegulatoryChange{
		Title:   "Critical: New EU GDPR Enforcement Fine for Data Breach",
		Summary: "The European Commission has imposed a significant enforcement fine on a major technology company for failing to comply with GDPR data breach notification requirements. This affects all organisations operating in the European Union.",
	}

	err := scanner.ClassifyChange(context.Background(), change)
	if err != nil {
		t.Fatalf("ClassifyChange returned error: %v", err)
	}

	// Should be classified as enforcement_decision
	if change.ChangeType != "enforcement_decision" {
		t.Errorf("expected change_type 'enforcement_decision', got '%s'", change.ChangeType)
	}

	// Should be critical severity
	if change.Severity != "critical" {
		t.Errorf("expected severity 'critical', got '%s'", change.Severity)
	}

	// Should detect EU region
	hasEU := false
	for _, r := range change.AffectedRegions {
		if r == "EU" {
			hasEU = true
		}
	}
	if !hasEU {
		t.Errorf("expected EU in affected_regions, got %v", change.AffectedRegions)
	}

	// Should detect GDPR tag
	hasGDPR := false
	for _, tag := range change.Tags {
		if tag == "GDPR" {
			hasGDPR = true
		}
	}
	if !hasGDPR {
		t.Errorf("expected GDPR in tags, got %v", change.Tags)
	}
}

// ============================================================
// XML STRUCT MARSHALLING TESTS
// ============================================================

// TestRSSFeedXMLRoundtrip verifies that the RSS XML structs can correctly
// unmarshal and marshal RSS 2.0 XML content.
func TestRSSFeedXMLRoundtrip(t *testing.T) {
	input := `<rss version="2.0"><channel><title>Test</title><link>https://test.com</link><description>Desc</description><item><title>Item 1</title><link>https://test.com/1</link><description>Item desc</description><pubDate>Mon, 01 Jan 2026 00:00:00 +0000</pubDate></item></channel></rss>`

	var feed RSSFeed
	if err := xml.Unmarshal([]byte(input), &feed); err != nil {
		t.Fatalf("failed to unmarshal RSS feed: %v", err)
	}

	if feed.Channel.Title != "Test" {
		t.Errorf("expected channel title 'Test', got '%s'", feed.Channel.Title)
	}
	if len(feed.Channel.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(feed.Channel.Items))
	}
	if feed.Channel.Items[0].Title != "Item 1" {
		t.Errorf("expected item title 'Item 1', got '%s'", feed.Channel.Items[0].Title)
	}
}

// TestAtomFeedXMLRoundtrip verifies that the Atom XML structs can correctly
// unmarshal Atom XML content.
func TestAtomFeedXMLRoundtrip(t *testing.T) {
	input := `<feed xmlns="http://www.w3.org/2005/Atom"><title>Atom Test</title><entry><title>Entry 1</title><link href="https://test.com/1" rel="alternate"/><summary>Entry summary</summary><updated>2026-01-01T00:00:00Z</updated><id>urn:test:1</id></entry></feed>`

	var feed AtomFeed
	if err := xml.Unmarshal([]byte(input), &feed); err != nil {
		t.Fatalf("failed to unmarshal Atom feed: %v", err)
	}

	if feed.Title != "Atom Test" {
		t.Errorf("expected feed title 'Atom Test', got '%s'", feed.Title)
	}
	if len(feed.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(feed.Entries))
	}
	if feed.Entries[0].Title != "Entry 1" {
		t.Errorf("expected entry title 'Entry 1', got '%s'", feed.Entries[0].Title)
	}
	if feed.Entries[0].Link.Href != "https://test.com/1" {
		t.Errorf("expected link href 'https://test.com/1', got '%s'", feed.Entries[0].Link.Href)
	}
}

// ============================================================
// HELPER FUNCTION TESTS
// ============================================================

// TestTruncate verifies the truncate helper function.
func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"exactly10!", 10, "exactly10!"},
		{"this is a longer string that should be truncated", 20, "this is a longer ..."},
		{"", 5, ""},
	}

	for _, tt := range tests {
		result := truncate(tt.input, tt.maxLen)
		if result != tt.expected {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
		}
	}
}

// TestRegContainsAny verifies the regContainsAny helper function.
func TestRegContainsAny(t *testing.T) {
	if !regContainsAny("hello world", "world") {
		t.Error("expected regContainsAny to find 'world' in 'hello world'")
	}
	if !regContainsAny("test enforcement action", "enforcement", "fine") {
		t.Error("expected regContainsAny to find 'enforcement'")
	}
	if regContainsAny("no match here", "missing", "absent") {
		t.Error("expected regContainsAny to return false when no terms match")
	}
}
