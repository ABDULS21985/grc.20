package service

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// ============================================================
// RSS / ATOM FEED DATA TYPES
// ============================================================

// RSSFeed represents a parsed RSS 2.0 feed.
type RSSFeed struct {
	XMLName xml.Name    `xml:"rss"`
	Channel RSSChannel  `xml:"channel"`
}

// RSSChannel is the <channel> element inside an RSS feed.
type RSSChannel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	Items       []RSSItem `xml:"item"`
}

// RSSItem is a single <item> inside an RSS channel.
type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	GUID        string `xml:"guid"`
	Category    string `xml:"category"`
}

// AtomFeed represents a parsed Atom feed.
type AtomFeed struct {
	XMLName xml.Name    `xml:"feed"`
	Title   string      `xml:"title"`
	Entries []AtomEntry `xml:"entry"`
}

// AtomEntry is a single <entry> inside an Atom feed.
type AtomEntry struct {
	Title   string     `xml:"title"`
	Link    AtomLink   `xml:"link"`
	Summary string     `xml:"summary"`
	Content string     `xml:"content"`
	Updated string     `xml:"updated"`
	ID      string     `xml:"id"`
}

// AtomLink represents the <link> element inside an Atom entry.
type AtomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
}

// RSSEntry is the normalised representation of a feed entry,
// regardless of whether the source was RSS 2.0 or Atom.
type RSSEntry struct {
	Title       string
	Link        string
	Description string
	PublishedAt time.Time
	GUID        string
	Category    string
}

// ScanResult summarises the outcome of scanning a single source.
type ScanResult struct {
	SourceID      uuid.UUID `json:"source_id"`
	SourceName    string    `json:"source_name"`
	EntriesFound  int       `json:"entries_found"`
	NewChanges    int       `json:"new_changes"`
	Duplicates    int       `json:"duplicates"`
	Errors        []string  `json:"errors,omitempty"`
	ScannedAt     time.Time `json:"scanned_at"`
}

// ============================================================
// SCANNER
// ============================================================

// RegulatoryScanner fetches and parses regulatory RSS/Atom feeds,
// de-duplicates entries, classifies their severity and affected
// frameworks, and persists new regulatory changes to the database.
type RegulatoryScanner struct {
	pool       *pgxpool.Pool
	httpClient *http.Client
}

// NewRegulatoryScanner creates a new scanner backed by the given pool.
func NewRegulatoryScanner(pool *pgxpool.Pool) *RegulatoryScanner {
	return &RegulatoryScanner{
		pool: pool,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ScanRSSFeeds scans all active sources that have an RSS feed URL configured.
// It returns nil on complete success, or a combined error summarising failures.
func (s *RegulatoryScanner) ScanRSSFeeds(ctx context.Context) error {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, rss_feed_url, relevance_frameworks
		FROM regulatory_sources
		WHERE is_active = true AND rss_feed_url IS NOT NULL AND rss_feed_url != ''
		ORDER BY name`)
	if err != nil {
		return fmt.Errorf("query active sources: %w", err)
	}
	defer rows.Close()

	type sourceInfo struct {
		ID         uuid.UUID
		Name       string
		FeedURL    string
		Frameworks []string
	}

	var sources []sourceInfo
	for rows.Next() {
		var si sourceInfo
		if err := rows.Scan(&si.ID, &si.Name, &si.FeedURL, &si.Frameworks); err != nil {
			log.Error().Err(err).Msg("scan source row")
			continue
		}
		sources = append(sources, si)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate sources: %w", err)
	}

	var scanErrors []string
	for _, src := range sources {
		result, err := s.ScanSource(ctx, src.ID)
		if err != nil {
			scanErrors = append(scanErrors, fmt.Sprintf("%s: %v", src.Name, err))
			log.Error().Err(err).Str("source", src.Name).Msg("failed to scan source")
			continue
		}
		log.Info().
			Str("source", src.Name).
			Int("found", result.EntriesFound).
			Int("new", result.NewChanges).
			Int("dupes", result.Duplicates).
			Msg("source scan complete")
	}

	if len(scanErrors) > 0 {
		return fmt.Errorf("scan errors (%d/%d sources): %s",
			len(scanErrors), len(sources), strings.Join(scanErrors, "; "))
	}
	return nil
}

// ScanSource scans a single regulatory source by its ID,
// fetching its RSS feed, de-duplicating entries, classifying
// severity, and inserting new regulatory_changes rows.
func (s *RegulatoryScanner) ScanSource(ctx context.Context, sourceID uuid.UUID) (*ScanResult, error) {
	var feedURL, name string
	var frameworks []string
	err := s.pool.QueryRow(ctx, `
		SELECT name, rss_feed_url, relevance_frameworks
		FROM regulatory_sources
		WHERE id = $1 AND is_active = true`, sourceID).Scan(&name, &feedURL, &frameworks)
	if err != nil {
		return nil, fmt.Errorf("get source %s: %w", sourceID, err)
	}

	if feedURL == "" {
		return nil, fmt.Errorf("source %s has no RSS feed URL", name)
	}

	entries, err := s.ParseRSSFeed(ctx, feedURL)
	if err != nil {
		return nil, fmt.Errorf("parse feed %s: %w", feedURL, err)
	}

	result := &ScanResult{
		SourceID:     sourceID,
		SourceName:   name,
		EntriesFound: len(entries),
		ScannedAt:    time.Now().UTC(),
	}

	for _, entry := range entries {
		isDupe, err := s.DeduplicateChange(ctx, entry.Title, entry.Link)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("dedup check: %v", err))
			continue
		}
		if isDupe {
			result.Duplicates++
			continue
		}

		change := &RegulatoryChange{
			SourceID:      &sourceID,
			Title:         entry.Title,
			Summary:       truncate(entry.Description, 2000),
			FullTextURL:   entry.Link,
			PublishedDate: &entry.PublishedAt,
			Status:        "new",
			Severity:      "medium",
			AffectedFrameworks: frameworks,
		}

		if err := s.ClassifyChange(ctx, change); err != nil {
			log.Warn().Err(err).Str("title", entry.Title).Msg("classification failed, using defaults")
		}

		_, err = s.pool.Exec(ctx, `
			INSERT INTO regulatory_changes
				(source_id, title, summary, full_text_url, published_date,
				 change_type, severity, status, affected_frameworks, affected_regions, tags, metadata)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
			change.SourceID,
			change.Title,
			change.Summary,
			change.FullTextURL,
			change.PublishedDate,
			change.ChangeType,
			change.Severity,
			change.Status,
			change.AffectedFrameworks,
			change.AffectedRegions,
			change.Tags,
			`{}`,
		)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("insert '%s': %v", truncate(entry.Title, 60), err))
			continue
		}
		result.NewChanges++
	}

	// Update last_scanned_at on the source
	_, _ = s.pool.Exec(ctx,
		`UPDATE regulatory_sources SET last_scanned_at = NOW() WHERE id = $1`, sourceID)

	return result, nil
}

// ParseRSSFeed fetches the given URL and parses it as either an RSS 2.0 or
// Atom XML feed, returning a normalised slice of RSSEntry values.
func (s *RegulatoryScanner) ParseRSSFeed(ctx context.Context, feedURL string) ([]RSSEntry, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "ComplianceForge-RegulatoryScanner/1.0")
	req.Header.Set("Accept", "application/rss+xml, application/atom+xml, application/xml, text/xml")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch feed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("feed returned HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024)) // 10 MB limit
	if err != nil {
		return nil, fmt.Errorf("read feed body: %w", err)
	}

	// Try RSS 2.0 first
	var rssFeed RSSFeed
	if err := xml.Unmarshal(body, &rssFeed); err == nil && len(rssFeed.Channel.Items) > 0 {
		return normaliseRSSItems(rssFeed.Channel.Items), nil
	}

	// Try Atom
	var atomFeed AtomFeed
	if err := xml.Unmarshal(body, &atomFeed); err == nil && len(atomFeed.Entries) > 0 {
		return normaliseAtomEntries(atomFeed.Entries), nil
	}

	return nil, fmt.Errorf("unable to parse feed as RSS 2.0 or Atom")
}

// ClassifyChange applies rule-based classification to determine
// severity, change_type, affected regions, and tags from the title
// and summary content using keyword matching.
func (s *RegulatoryScanner) ClassifyChange(ctx context.Context, change *RegulatoryChange) error {
	combined := strings.ToLower(change.Title + " " + change.Summary)

	// Classify change type
	change.ChangeType = classifyChangeType(combined)

	// Classify severity
	change.Severity = classifySeverity(combined)

	// Detect affected regions from keywords
	change.AffectedRegions = detectRegions(combined)

	// Generate tags from keywords
	change.Tags = detectTags(combined)

	return nil
}

// DeduplicateChange checks whether a regulatory change with the same URL
// or a highly similar title already exists. Returns true if a duplicate
// is found (meaning the entry should be skipped).
func (s *RegulatoryScanner) DeduplicateChange(ctx context.Context, title, url string) (bool, error) {
	// Exact URL match
	if url != "" {
		var count int
		err := s.pool.QueryRow(ctx,
			`SELECT COUNT(*) FROM regulatory_changes WHERE full_text_url = $1`, url).Scan(&count)
		if err != nil {
			return false, fmt.Errorf("check URL duplicate: %w", err)
		}
		if count > 0 {
			return true, nil
		}
	}

	// Title similarity: case-insensitive containment check
	if title != "" {
		normalised := strings.TrimSpace(strings.ToLower(title))
		var count int
		err := s.pool.QueryRow(ctx,
			`SELECT COUNT(*) FROM regulatory_changes WHERE LOWER(title) = $1`, normalised).Scan(&count)
		if err != nil {
			return false, fmt.Errorf("check title duplicate: %w", err)
		}
		if count > 0 {
			return true, nil
		}
	}

	return false, nil
}

// ============================================================
// INTERNAL HELPERS
// ============================================================

// normaliseRSSItems converts RSS 2.0 items into the normalised RSSEntry slice.
func normaliseRSSItems(items []RSSItem) []RSSEntry {
	entries := make([]RSSEntry, 0, len(items))
	for _, item := range items {
		pub := parseFlexibleDate(item.PubDate)
		entries = append(entries, RSSEntry{
			Title:       strings.TrimSpace(item.Title),
			Link:        strings.TrimSpace(item.Link),
			Description: strings.TrimSpace(item.Description),
			PublishedAt: pub,
			GUID:        item.GUID,
			Category:    item.Category,
		})
	}
	return entries
}

// normaliseAtomEntries converts Atom entries into the normalised RSSEntry slice.
func normaliseAtomEntries(entries []AtomEntry) []RSSEntry {
	result := make([]RSSEntry, 0, len(entries))
	for _, entry := range entries {
		pub := parseFlexibleDate(entry.Updated)
		desc := entry.Summary
		if desc == "" {
			desc = entry.Content
		}
		link := entry.Link.Href
		result = append(result, RSSEntry{
			Title:       strings.TrimSpace(entry.Title),
			Link:        strings.TrimSpace(link),
			Description: strings.TrimSpace(desc),
			PublishedAt: pub,
			GUID:        entry.ID,
		})
	}
	return result
}

// parseFlexibleDate attempts to parse a date string using common RSS/Atom formats.
func parseFlexibleDate(s string) time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Now().UTC()
	}
	formats := []string{
		time.RFC1123Z,                    // Mon, 02 Jan 2006 15:04:05 -0700
		time.RFC1123,                     // Mon, 02 Jan 2006 15:04:05 MST
		time.RFC3339,                     // 2006-01-02T15:04:05Z07:00
		"2006-01-02T15:04:05Z",          // ISO 8601 UTC
		"2006-01-02T15:04:05-07:00",     // ISO 8601 with offset
		"2006-01-02",                     // date only
		"Mon, 2 Jan 2006 15:04:05 -0700", // single-digit day
		"02 Jan 2006 15:04:05 -0700",    // no weekday
	}
	for _, layout := range formats {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC()
		}
	}
	return time.Now().UTC()
}

// classifyChangeType determines the regulatory_change_type based on keywords.
func classifyChangeType(text string) string {
	switch {
	case regContainsAny(text, "enforcement", "fine", "penalty", "sanction", "infringement"):
		return "enforcement_decision"
	case regContainsAny(text, "new regulation", "new law", "new directive", "enacted", "adopted regulation"):
		return "new_regulation"
	case regContainsAny(text, "amendment", "amend", "revised", "revision to"):
		return "amendment"
	case regContainsAny(text, "consultation", "call for evidence", "request for comment", "rfc", "public comment"):
		return "consultation"
	case regContainsAny(text, "guidance", "guideline", "advisory", "recommendation", "best practice"):
		return "guidance"
	case regContainsAny(text, "court ruling", "judgment", "court decision", "tribunal"):
		return "court_ruling"
	case regContainsAny(text, "standard revision", "iso update", "nist update", "framework update"):
		return "standard_revision"
	case regContainsAny(text, "standard update", "pci update", "itil update"):
		return "standard_update"
	case regContainsAny(text, "bulletin", "newsletter", "announcement", "press release"):
		return "industry_bulletin"
	default:
		return "guidance"
	}
}

// classifySeverity determines severity based on keyword presence.
func classifySeverity(text string) string {
	switch {
	case regContainsAny(text, "critical", "emergency", "immediate action", "mandatory",
		"breach", "zero-day", "severe", "urgent"):
		return "critical"
	case regContainsAny(text, "high risk", "significant", "enforcement", "fine", "penalty",
		"non-compliance", "deadline", "mandatory requirement"):
		return "high"
	case regContainsAny(text, "update", "amendment", "revision", "change", "moderate"):
		return "medium"
	case regContainsAny(text, "minor", "low risk", "informational", "newsletter",
		"summary", "digest", "overview"):
		return "low"
	default:
		return "informational"
	}
}

// detectRegions extracts affected region codes from text content.
func detectRegions(text string) []string {
	regionMap := map[string][]string{
		"uk":              {"GB"},
		"united kingdom":  {"GB"},
		"britain":         {"GB"},
		"eu":              {"EU"},
		"european union":  {"EU"},
		"europe":          {"EU"},
		"germany":         {"DE"},
		"deutschland":     {"DE"},
		"france":          {"FR"},
		"united states":   {"US"},
		"usa":             {"US"},
		"global":          {"GLOBAL"},
		"international":   {"GLOBAL"},
		"worldwide":       {"GLOBAL"},
	}

	seen := make(map[string]bool)
	var regions []string
	for keyword, codes := range regionMap {
		if strings.Contains(text, keyword) {
			for _, code := range codes {
				if !seen[code] {
					seen[code] = true
					regions = append(regions, code)
				}
			}
		}
	}
	return regions
}

// detectTags generates relevant tags from keywords in the content.
func detectTags(text string) []string {
	tagKeywords := map[string]string{
		"gdpr":            "GDPR",
		"data protection": "data-protection",
		"privacy":         "privacy",
		"cybersecurity":   "cybersecurity",
		"cyber security":  "cybersecurity",
		"nis2":            "NIS2",
		"iso 27001":       "ISO27001",
		"iso27001":        "ISO27001",
		"nist":            "NIST",
		"pci dss":         "PCI-DSS",
		"pci-dss":         "PCI-DSS",
		"ransomware":      "ransomware",
		"supply chain":    "supply-chain",
		"cloud":           "cloud",
		"ai":              "artificial-intelligence",
		"artificial intelligence": "artificial-intelligence",
		"dora":            "DORA",
		"financial":       "financial-services",
		"healthcare":      "healthcare",
		"incident":        "incident-response",
		"encryption":      "encryption",
		"access control":  "access-control",
	}

	seen := make(map[string]bool)
	var tags []string
	for keyword, tag := range tagKeywords {
		if strings.Contains(text, keyword) && !seen[tag] {
			seen[tag] = true
			tags = append(tags, tag)
		}
	}
	return tags
}

// regContainsAny returns true if text contains any of the given substrings.
func regContainsAny(text string, terms ...string) bool {
	for _, term := range terms {
		if strings.Contains(text, term) {
			return true
		}
	}
	return false
}

// truncate shortens a string to at most maxLen characters.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
