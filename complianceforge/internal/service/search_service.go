package service

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// ============================================================
// SEARCH ENGINE
// Unified full-text search across all ComplianceForge entities.
// Uses PostgreSQL TSVECTOR/ts_query for relevance-ranked search
// with faceted filtering, autocomplete, and related-entity
// discovery.
// ============================================================

// SearchEngine provides unified search across the platform.
type SearchEngine struct {
	pool *pgxpool.Pool
}

// NewSearchEngine creates a new SearchEngine.
func NewSearchEngine(pool *pgxpool.Pool) *SearchEngine {
	return &SearchEngine{pool: pool}
}

// ============================================================
// DATA TYPES
// ============================================================

// SearchRequest holds parameters for a unified search query.
type SearchRequest struct {
	Query       string   `json:"query"`
	EntityTypes []string `json:"entity_types,omitempty"`
	Frameworks  []string `json:"frameworks,omitempty"`
	Statuses    []string `json:"statuses,omitempty"`
	Severities  []string `json:"severities,omitempty"`
	Categories  []string `json:"categories,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	DateFrom    string   `json:"date_from,omitempty"`
	DateTo      string   `json:"date_to,omitempty"`
	SortBy      string   `json:"sort_by,omitempty"` // relevance, date, title, severity
	Page        int      `json:"page,omitempty"`
	PageSize    int      `json:"page_size,omitempty"`
}

// SearchResult represents a single result from the unified search.
type SearchResult struct {
	EntityType     string    `json:"entity_type"`
	EntityID       uuid.UUID `json:"entity_id"`
	EntityRef      string    `json:"entity_ref"`
	Title          string    `json:"title"`
	Snippet        string    `json:"snippet"`
	Score          float64   `json:"score"`
	Status         string    `json:"status,omitempty"`
	Severity       string    `json:"severity,omitempty"`
	Category       string    `json:"category,omitempty"`
	FrameworkCodes []string  `json:"framework_codes,omitempty"`
	Tags           []string  `json:"tags,omitempty"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// FacetBucket holds a value and its count for faceted search.
type FacetBucket struct {
	Value string `json:"value"`
	Count int    `json:"count"`
}

// SearchFacets holds aggregated filter counts.
type SearchFacets struct {
	EntityTypes []FacetBucket `json:"entity_types"`
	Frameworks  []FacetBucket `json:"frameworks"`
	Statuses    []FacetBucket `json:"statuses"`
	Severities  []FacetBucket `json:"severities"`
	Categories  []FacetBucket `json:"categories"`
}

// SearchResponse wraps results, facets, and metadata.
type SearchResponse struct {
	Results      []SearchResult `json:"results"`
	Facets       SearchFacets   `json:"facets"`
	TotalResults int            `json:"total_results"`
	QueryTimeMs  int64          `json:"query_time_ms"`
	Page         int            `json:"page"`
	PageSize     int            `json:"page_size"`
	TotalPages   int            `json:"total_pages"`
}

// AutocompleteResult represents a single autocomplete suggestion.
type AutocompleteResult struct {
	EntityType string    `json:"entity_type"`
	EntityID   uuid.UUID `json:"entity_id"`
	EntityRef  string    `json:"entity_ref"`
	Title      string    `json:"title"`
	Status     string    `json:"status,omitempty"`
}

// RelatedEntity groups related items by category.
type RelatedEntity struct {
	Relationship string         `json:"relationship"`
	Items        []SearchResult `json:"items"`
}

// RecentSearch represents a user's recent search query.
type RecentSearch struct {
	ID              uuid.UUID  `json:"id"`
	Query           string     `json:"query"`
	ResultCount     int        `json:"result_count"`
	ClickedType     *string    `json:"clicked_entity_type,omitempty"`
	ClickedEntityID *uuid.UUID `json:"clicked_entity_id,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

// IndexStats holds counts for search index health monitoring.
type IndexStats struct {
	EntityType   string    `json:"entity_type"`
	IndexedCount int       `json:"indexed_count"`
	LastIndexed  time.Time `json:"last_indexed"`
}

// ============================================================
// SEARCH
// ============================================================

// BuildTSQuery converts a user query string into a PostgreSQL
// tsquery. Supports quoted phrases, OR, and negation (-term).
func BuildTSQuery(raw string) string {
	if raw == "" {
		return ""
	}

	raw = strings.TrimSpace(raw)
	var parts []string
	inQuote := false
	var current strings.Builder

	for i := 0; i < len(raw); i++ {
		ch := raw[i]
		if ch == '"' {
			if inQuote {
				// End of quoted phrase — use <-> (phrase operator)
				phrase := strings.TrimSpace(current.String())
				if phrase != "" {
					words := strings.Fields(phrase)
					parts = append(parts, strings.Join(words, " <-> "))
				}
				current.Reset()
				inQuote = false
			} else {
				// Flush any accumulated words before the quote
				flushWords(&current, &parts)
				inQuote = true
			}
			continue
		}
		current.WriteByte(ch)
	}
	// Flush remaining
	flushWords(&current, &parts)

	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, " & ")
}

func flushWords(buf *strings.Builder, parts *[]string) {
	text := strings.TrimSpace(buf.String())
	buf.Reset()
	if text == "" {
		return
	}
	for _, w := range strings.Fields(text) {
		upper := strings.ToUpper(w)
		if upper == "OR" {
			// Replace last & with |
			if len(*parts) > 0 {
				(*parts)[len(*parts)-1] = (*parts)[len(*parts)-1] + " |"
			}
			continue
		}
		if strings.HasPrefix(w, "-") && len(w) > 1 {
			*parts = append(*parts, "!"+w[1:]+":*")
			continue
		}
		// Prefix match for partial words
		*parts = append(*parts, w+":*")
	}
}

// Search performs a unified full-text search across all indexed entities.
func (s *SearchEngine) Search(ctx context.Context, orgID uuid.UUID, req SearchRequest) (*SearchResponse, error) {
	start := time.Now()

	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}
	if req.SortBy == "" {
		req.SortBy = "relevance"
	}

	tsq := BuildTSQuery(req.Query)

	// Build dynamic WHERE clauses
	args := []interface{}{orgID}
	argN := 2
	var conditions []string
	conditions = append(conditions, "si.organization_id = $1")

	if tsq != "" {
		conditions = append(conditions, fmt.Sprintf("si.search_vector @@ to_tsquery('english', $%d)", argN))
		args = append(args, tsq)
		argN++
	}

	if len(req.EntityTypes) > 0 {
		conditions = append(conditions, fmt.Sprintf("si.entity_type = ANY($%d)", argN))
		args = append(args, req.EntityTypes)
		argN++
	}
	if len(req.Frameworks) > 0 {
		conditions = append(conditions, fmt.Sprintf("si.framework_codes && $%d", argN))
		args = append(args, req.Frameworks)
		argN++
	}
	if len(req.Statuses) > 0 {
		conditions = append(conditions, fmt.Sprintf("si.status = ANY($%d)", argN))
		args = append(args, req.Statuses)
		argN++
	}
	if len(req.Severities) > 0 {
		conditions = append(conditions, fmt.Sprintf("si.severity = ANY($%d)", argN))
		args = append(args, req.Severities)
		argN++
	}
	if len(req.Categories) > 0 {
		conditions = append(conditions, fmt.Sprintf("si.category = ANY($%d)", argN))
		args = append(args, req.Categories)
		argN++
	}
	if len(req.Tags) > 0 {
		conditions = append(conditions, fmt.Sprintf("si.tags && $%d", argN))
		args = append(args, req.Tags)
		argN++
	}
	if req.DateFrom != "" {
		conditions = append(conditions, fmt.Sprintf("si.updated_at >= $%d::timestamptz", argN))
		args = append(args, req.DateFrom)
		argN++
	}
	if req.DateTo != "" {
		conditions = append(conditions, fmt.Sprintf("si.updated_at <= $%d::timestamptz", argN))
		args = append(args, req.DateTo)
		argN++
	}

	where := strings.Join(conditions, " AND ")

	// Score expression
	scoreExpr := "0"
	if tsq != "" {
		scoreExpr = fmt.Sprintf("ts_rank_cd(si.search_vector, to_tsquery('english', '%s'))", strings.ReplaceAll(tsq, "'", "''"))
	}

	// Order by
	orderBy := "score DESC, si.updated_at DESC"
	switch req.SortBy {
	case "date":
		orderBy = "si.updated_at DESC"
	case "title":
		orderBy = "si.title ASC"
	case "severity":
		orderBy = "si.severity ASC, score DESC"
	}

	offset := (req.Page - 1) * req.PageSize

	// Main query
	query := fmt.Sprintf(`
		SELECT
			si.entity_type,
			si.entity_id,
			COALESCE(si.entity_ref, '') AS entity_ref,
			si.title,
			COALESCE(si.body, '') AS body,
			%s AS score,
			COALESCE(si.status, '') AS status,
			COALESCE(si.severity, '') AS severity,
			COALESCE(si.category, '') AS category,
			COALESCE(si.framework_codes, '{}') AS framework_codes,
			COALESCE(si.tags, '{}') AS tags,
			si.updated_at
		FROM search_index si
		WHERE %s
		ORDER BY %s
		LIMIT %d OFFSET %d
	`, scoreExpr, where, orderBy, req.PageSize, offset)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("search begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return nil, fmt.Errorf("search set org: %w", err)
	}

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("search query: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var r SearchResult
		var body string
		if err := rows.Scan(
			&r.EntityType, &r.EntityID, &r.EntityRef, &r.Title,
			&body, &r.Score, &r.Status, &r.Severity, &r.Category,
			&r.FrameworkCodes, &r.Tags, &r.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("search scan: %w", err)
		}

		// Generate snippet
		r.Snippet = generateSnippet(body, req.Query, 200)
		results = append(results, r)
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM search_index si WHERE %s", where)
	var total int
	if err := tx.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("search count: %w", err)
	}

	// Facets
	facets := SearchFacets{}
	facets.EntityTypes = s.queryFacet(ctx, tx, "entity_type", where, args)
	facets.Statuses = s.queryFacet(ctx, tx, "status", where, args)
	facets.Severities = s.queryFacet(ctx, tx, "severity", where, args)
	facets.Categories = s.queryFacet(ctx, tx, "category", where, args)
	facets.Frameworks = s.queryFrameworkFacet(ctx, tx, where, args)

	tx.Commit(ctx)

	elapsed := time.Since(start).Milliseconds()

	return &SearchResponse{
		Results:      results,
		Facets:       facets,
		TotalResults: total,
		QueryTimeMs:  elapsed,
		Page:         req.Page,
		PageSize:     req.PageSize,
		TotalPages:   int(math.Ceil(float64(total) / float64(req.PageSize))),
	}, nil
}

func (s *SearchEngine) queryFacet(ctx context.Context, tx pgx.Tx, column, where string, args []interface{}) []FacetBucket {
	q := fmt.Sprintf(`
		SELECT COALESCE(%s, '') AS val, COUNT(*) AS cnt
		FROM search_index si
		WHERE %s AND %s IS NOT NULL AND %s != ''
		GROUP BY val ORDER BY cnt DESC LIMIT 20
	`, column, where, column, column)

	rows, err := tx.Query(ctx, q, args...)
	if err != nil {
		log.Warn().Err(err).Str("column", column).Msg("facet query failed")
		return nil
	}
	defer rows.Close()

	var buckets []FacetBucket
	for rows.Next() {
		var b FacetBucket
		if err := rows.Scan(&b.Value, &b.Count); err == nil && b.Value != "" {
			buckets = append(buckets, b)
		}
	}
	return buckets
}

func (s *SearchEngine) queryFrameworkFacet(ctx context.Context, tx pgx.Tx, where string, args []interface{}) []FacetBucket {
	q := fmt.Sprintf(`
		SELECT fw, COUNT(*) AS cnt
		FROM search_index si, LATERAL unnest(si.framework_codes) AS fw
		WHERE %s
		GROUP BY fw ORDER BY cnt DESC LIMIT 20
	`, where)

	rows, err := tx.Query(ctx, q, args...)
	if err != nil {
		log.Warn().Err(err).Msg("framework facet query failed")
		return nil
	}
	defer rows.Close()

	var buckets []FacetBucket
	for rows.Next() {
		var b FacetBucket
		if err := rows.Scan(&b.Value, &b.Count); err == nil {
			buckets = append(buckets, b)
		}
	}
	return buckets
}

// generateSnippet creates a truncated text snippet around query terms.
func generateSnippet(body, query string, maxLen int) string {
	if body == "" {
		return ""
	}
	if len(body) <= maxLen {
		return body
	}

	lowerBody := strings.ToLower(body)
	queryWords := strings.Fields(strings.ToLower(query))

	bestPos := 0
	for _, w := range queryWords {
		idx := strings.Index(lowerBody, w)
		if idx > 0 {
			bestPos = idx
			break
		}
	}

	start := bestPos - maxLen/2
	if start < 0 {
		start = 0
	}
	end := start + maxLen
	if end > len(body) {
		end = len(body)
		start = end - maxLen
		if start < 0 {
			start = 0
		}
	}

	snippet := body[start:end]
	if start > 0 {
		snippet = "..." + snippet
	}
	if end < len(body) {
		snippet = snippet + "..."
	}
	return snippet
}

// ============================================================
// AUTOCOMPLETE
// ============================================================

// Autocomplete returns quick prefix-match suggestions for the search bar.
func (s *SearchEngine) Autocomplete(ctx context.Context, orgID uuid.UUID, prefix string, limit int) ([]AutocompleteResult, error) {
	if len(prefix) < 2 {
		return nil, nil
	}
	if limit < 1 || limit > 20 {
		limit = 10
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("autocomplete begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return nil, fmt.Errorf("autocomplete set org: %w", err)
	}

	tsq := prefix + ":*"
	query := `
		SELECT entity_type, entity_id, COALESCE(entity_ref, ''), title, COALESCE(status, '')
		FROM search_index
		WHERE organization_id = $1
		  AND search_vector @@ to_tsquery('english', $2)
		ORDER BY ts_rank_cd(search_vector, to_tsquery('english', $2)) DESC
		LIMIT $3
	`

	rows, err := tx.Query(ctx, query, orgID, tsq, limit)
	if err != nil {
		return nil, fmt.Errorf("autocomplete query: %w", err)
	}
	defer rows.Close()

	var results []AutocompleteResult
	for rows.Next() {
		var r AutocompleteResult
		if err := rows.Scan(&r.EntityType, &r.EntityID, &r.EntityRef, &r.Title, &r.Status); err != nil {
			return nil, fmt.Errorf("autocomplete scan: %w", err)
		}
		results = append(results, r)
	}

	tx.Commit(ctx)
	return results, nil
}

// ============================================================
// RELATED ENTITIES
// ============================================================

// GetRelatedEntities finds items related to a given entity
// using shared tags, framework codes, and text similarity.
func (s *SearchEngine) GetRelatedEntities(ctx context.Context, orgID uuid.UUID, entityType string, entityID uuid.UUID) ([]RelatedEntity, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("related begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return nil, fmt.Errorf("related set org: %w", err)
	}

	// Get the source entity from the index
	var sourceTags []string
	var sourceFrameworks []string
	var sourceVector string
	err = tx.QueryRow(ctx, `
		SELECT COALESCE(tags, '{}'), COALESCE(framework_codes, '{}'), search_vector::text
		FROM search_index
		WHERE organization_id = $1 AND entity_type = $2 AND entity_id = $3
	`, orgID, entityType, entityID).Scan(&sourceTags, &sourceFrameworks, &sourceVector)
	if err != nil {
		tx.Commit(ctx)
		return nil, nil // Entity not in index
	}

	var groups []RelatedEntity

	// 1. Same-framework items (different type)
	if len(sourceFrameworks) > 0 {
		rows, err := tx.Query(ctx, `
			SELECT entity_type, entity_id, COALESCE(entity_ref, ''), title,
				   COALESCE(status, ''), COALESCE(severity, ''),
				   COALESCE(framework_codes, '{}'), updated_at
			FROM search_index
			WHERE organization_id = $1
			  AND (entity_type != $2 OR entity_id != $3)
			  AND framework_codes && $4
			ORDER BY updated_at DESC
			LIMIT 10
		`, orgID, entityType, entityID, sourceFrameworks)
		if err == nil {
			items := scanRelatedResults(rows)
			if len(items) > 0 {
				groups = append(groups, RelatedEntity{Relationship: "same_framework", Items: items})
			}
		}
	}

	// 2. Same-tag items
	if len(sourceTags) > 0 {
		rows, err := tx.Query(ctx, `
			SELECT entity_type, entity_id, COALESCE(entity_ref, ''), title,
				   COALESCE(status, ''), COALESCE(severity, ''),
				   COALESCE(framework_codes, '{}'), updated_at
			FROM search_index
			WHERE organization_id = $1
			  AND (entity_type != $2 OR entity_id != $3)
			  AND tags && $4
			ORDER BY updated_at DESC
			LIMIT 10
		`, orgID, entityType, entityID, sourceTags)
		if err == nil {
			items := scanRelatedResults(rows)
			if len(items) > 0 {
				groups = append(groups, RelatedEntity{Relationship: "shared_tags", Items: items})
			}
		}
	}

	// 3. Same type (sibling entities)
	rows, err := tx.Query(ctx, `
		SELECT entity_type, entity_id, COALESCE(entity_ref, ''), title,
			   COALESCE(status, ''), COALESCE(severity, ''),
			   COALESCE(framework_codes, '{}'), updated_at
		FROM search_index
		WHERE organization_id = $1
		  AND entity_type = $2
		  AND entity_id != $3
		ORDER BY updated_at DESC
		LIMIT 5
	`, orgID, entityType, entityID)
	if err == nil {
		items := scanRelatedResults(rows)
		if len(items) > 0 {
			groups = append(groups, RelatedEntity{Relationship: "same_type", Items: items})
		}
	}

	tx.Commit(ctx)
	return groups, nil
}

func scanRelatedResults(rows pgx.Rows) []SearchResult {
	defer rows.Close()
	var items []SearchResult
	for rows.Next() {
		var r SearchResult
		if err := rows.Scan(
			&r.EntityType, &r.EntityID, &r.EntityRef, &r.Title,
			&r.Status, &r.Severity, &r.FrameworkCodes, &r.UpdatedAt,
		); err == nil {
			items = append(items, r)
		}
	}
	return items
}

// ============================================================
// INDEXING
// ============================================================

// IndexRecord is the data needed to upsert a search index entry.
type IndexRecord struct {
	EntityType     string
	EntityID       uuid.UUID
	EntityRef      string
	Title          string
	Body           string
	Tags           []string
	FrameworkCodes []string
	Status         string
	Severity       string
	Category       string
	Metadata       map[string]interface{}
}

// IndexEntity upserts a single entity into the search index.
func (s *SearchEngine) IndexEntity(ctx context.Context, orgID uuid.UUID, rec IndexRecord) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("index begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return fmt.Errorf("index set org: %w", err)
	}

	query := `
		INSERT INTO search_index (
			organization_id, entity_type, entity_id, entity_ref,
			title, body, tags, framework_codes,
			status, severity, category, metadata, indexed_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW())
		ON CONFLICT (organization_id, entity_type, entity_id)
		DO UPDATE SET
			entity_ref = EXCLUDED.entity_ref,
			title = EXCLUDED.title,
			body = EXCLUDED.body,
			tags = EXCLUDED.tags,
			framework_codes = EXCLUDED.framework_codes,
			status = EXCLUDED.status,
			severity = EXCLUDED.severity,
			category = EXCLUDED.category,
			metadata = EXCLUDED.metadata,
			indexed_at = NOW()
	`

	metadataJSON := "{}"
	if rec.Metadata != nil {
		// Simple JSON encoding
		pairs := make([]string, 0, len(rec.Metadata))
		for k, v := range rec.Metadata {
			pairs = append(pairs, fmt.Sprintf(`"%s":"%v"`, k, v))
		}
		metadataJSON = "{" + strings.Join(pairs, ",") + "}"
	}

	if rec.Tags == nil {
		rec.Tags = []string{}
	}
	if rec.FrameworkCodes == nil {
		rec.FrameworkCodes = []string{}
	}

	_, err = tx.Exec(ctx, query,
		orgID, rec.EntityType, rec.EntityID, rec.EntityRef,
		rec.Title, rec.Body, rec.Tags, rec.FrameworkCodes,
		rec.Status, rec.Severity, rec.Category, metadataJSON,
	)
	if err != nil {
		return fmt.Errorf("index upsert: %w", err)
	}

	return tx.Commit(ctx)
}

// RemoveEntity removes an entity from the search index.
func (s *SearchEngine) RemoveEntity(ctx context.Context, orgID uuid.UUID, entityType string, entityID uuid.UUID) error {
	_, err := s.pool.Exec(ctx, `
		DELETE FROM search_index
		WHERE organization_id = $1 AND entity_type = $2 AND entity_id = $3
	`, orgID, entityType, entityID)
	return err
}

// IndexAllEntities triggers a full re-index for an organization.
// Each entity type is indexed in batch.
func (s *SearchEngine) IndexAllEntities(ctx context.Context, orgID uuid.UUID) (*IndexReport, error) {
	report := &IndexReport{StartedAt: time.Now()}

	entityTypes := []string{
		"control", "risk", "policy", "incident", "vendor",
		"asset", "finding", "exception", "dsr_request",
		"regulatory_change", "processing_activity", "remediation_action",
	}

	for _, et := range entityTypes {
		count, errs := s.indexEntityType(ctx, orgID, et)
		report.TypeCounts = append(report.TypeCounts, IndexTypeCount{
			EntityType: et,
			Indexed:    count,
			Errors:     errs,
		})
		report.TotalIndexed += count
		report.TotalErrors += errs
	}

	report.CompletedAt = time.Now()
	report.DurationMs = report.CompletedAt.Sub(report.StartedAt).Milliseconds()
	return report, nil
}

// IndexReport holds the results of a full re-index operation.
type IndexReport struct {
	TotalIndexed int              `json:"total_indexed"`
	TotalErrors  int              `json:"total_errors"`
	TypeCounts   []IndexTypeCount `json:"type_counts"`
	StartedAt    time.Time        `json:"started_at"`
	CompletedAt  time.Time        `json:"completed_at"`
	DurationMs   int64            `json:"duration_ms"`
}

// IndexTypeCount holds counts for a single entity type.
type IndexTypeCount struct {
	EntityType string `json:"entity_type"`
	Indexed    int    `json:"indexed"`
	Errors     int    `json:"errors"`
}

// indexEntityType indexes all entities of a given type for an org.
func (s *SearchEngine) indexEntityType(ctx context.Context, orgID uuid.UUID, entityType string) (int, int) {
	tableMap := map[string]string{
		"risk":                "risks",
		"policy":             "policies",
		"incident":           "incidents",
		"vendor":             "vendors",
		"asset":              "assets",
		"finding":            "audit_findings",
		"exception":          "exceptions",
		"dsr_request":        "dsr_requests",
		"regulatory_change":  "regulatory_changes",
		"processing_activity": "processing_activities",
		"remediation_action": "remediation_actions",
		"control":            "control_implementations",
	}

	table, ok := tableMap[entityType]
	if !ok {
		return 0, 0
	}

	// Generic indexing: fetch id, title-like column, and body-like column
	// Each table has different column names, so use a query builder
	query := buildIndexQuery(entityType, table, orgID)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		log.Error().Err(err).Str("entity_type", entityType).Msg("index begin tx failed")
		return 0, 1
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return 0, 1
	}

	rows, err := tx.Query(ctx, query, orgID)
	if err != nil {
		log.Warn().Err(err).Str("entity_type", entityType).Str("table", table).Msg("index query failed — table may not exist")
		return 0, 0
	}
	defer rows.Close()

	count := 0
	errors := 0
	for rows.Next() {
		var id uuid.UUID
		var ref, title, body, status, severity, category string
		if err := rows.Scan(&id, &ref, &title, &body, &status, &severity, &category); err != nil {
			errors++
			continue
		}

		rec := IndexRecord{
			EntityType: entityType,
			EntityID:   id,
			EntityRef:  ref,
			Title:      title,
			Body:       body,
			Status:     status,
			Severity:   severity,
			Category:   category,
		}

		if err := s.indexSingleInTx(ctx, tx, orgID, rec); err != nil {
			errors++
			continue
		}
		count++
	}

	tx.Commit(ctx)
	return count, errors
}

func (s *SearchEngine) indexSingleInTx(ctx context.Context, tx pgx.Tx, orgID uuid.UUID, rec IndexRecord) error {
	if rec.Tags == nil {
		rec.Tags = []string{}
	}
	if rec.FrameworkCodes == nil {
		rec.FrameworkCodes = []string{}
	}

	_, err := tx.Exec(ctx, `
		INSERT INTO search_index (
			organization_id, entity_type, entity_id, entity_ref,
			title, body, tags, framework_codes,
			status, severity, category, indexed_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW())
		ON CONFLICT (organization_id, entity_type, entity_id)
		DO UPDATE SET
			entity_ref = EXCLUDED.entity_ref,
			title = EXCLUDED.title,
			body = EXCLUDED.body,
			tags = EXCLUDED.tags,
			framework_codes = EXCLUDED.framework_codes,
			status = EXCLUDED.status,
			severity = EXCLUDED.severity,
			category = EXCLUDED.category,
			indexed_at = NOW()
	`, orgID, rec.EntityType, rec.EntityID, rec.EntityRef,
		rec.Title, rec.Body, rec.Tags, rec.FrameworkCodes,
		rec.Status, rec.Severity, rec.Category,
	)
	return err
}

// buildIndexQuery returns a SELECT query that fetches indexable
// data from the source table for a given entity type.
func buildIndexQuery(entityType, table string, orgID uuid.UUID) string {
	switch entityType {
	case "risk":
		return `SELECT id, COALESCE(reference_id, ''), title,
			COALESCE(description, ''), COALESCE(status, ''),
			COALESCE(inherent_likelihood || 'x' || inherent_impact, ''), COALESCE(risk_category, '')
			FROM risks WHERE organization_id = $1`
	case "policy":
		return `SELECT id, COALESCE(reference_id, ''), title,
			COALESCE(content, ''), COALESCE(status, ''), '', COALESCE(category, '')
			FROM policies WHERE organization_id = $1`
	case "incident":
		return `SELECT id, COALESCE(reference_id, ''), title,
			COALESCE(description, ''), COALESCE(status, ''),
			COALESCE(severity, ''), COALESCE(incident_type, '')
			FROM incidents WHERE organization_id = $1`
	case "vendor":
		return `SELECT id, COALESCE(reference_id, ''), company_name,
			COALESCE(description, ''), COALESCE(risk_tier, ''),
			COALESCE(overall_risk_level, ''), ''
			FROM vendors WHERE organization_id = $1`
	case "asset":
		return `SELECT id, COALESCE(reference_id, ''), name,
			COALESCE(description, ''), COALESCE(status, ''),
			COALESCE(criticality, ''), COALESCE(asset_type, '')
			FROM assets WHERE organization_id = $1`
	case "finding":
		return `SELECT id, COALESCE(reference_id, ''), title,
			COALESCE(description, ''), COALESCE(status, ''),
			COALESCE(severity, ''), ''
			FROM audit_findings WHERE organization_id = $1`
	case "exception":
		return `SELECT id, COALESCE(reference_id, ''), title,
			COALESCE(justification, ''), COALESCE(status, ''),
			COALESCE(risk_level, ''), ''
			FROM exceptions WHERE organization_id = $1`
	case "dsr_request":
		return `SELECT id, COALESCE(reference_id, ''), COALESCE(request_type, ''),
			COALESCE(description, ''), COALESCE(status, ''), '', ''
			FROM dsr_requests WHERE organization_id = $1`
	case "regulatory_change":
		return `SELECT id, COALESCE(reference_id, ''), title,
			COALESCE(summary, ''), COALESCE(status, ''),
			COALESCE(severity, ''), COALESCE(change_type, '')
			FROM regulatory_changes WHERE organization_id = $1`
	case "processing_activity":
		return `SELECT id, COALESCE(reference_id, ''), name,
			COALESCE(description, ''), COALESCE(status, ''),
			COALESCE(risk_level, ''), COALESCE(lawful_basis, '')
			FROM processing_activities WHERE organization_id = $1`
	case "remediation_action":
		return `SELECT id, COALESCE(reference_id, ''), title,
			COALESCE(description, ''), COALESCE(status, ''),
			COALESCE(priority, ''), ''
			FROM remediation_actions WHERE organization_id = $1`
	case "control":
		return `SELECT id, COALESCE(reference_id, ''), COALESCE(control_code, ''),
			COALESCE(notes, ''), COALESCE(status, ''), '', ''
			FROM control_implementations WHERE organization_id = $1`
	default:
		return fmt.Sprintf("SELECT id, '', '', '', '', '', '' FROM %s WHERE organization_id = $1 LIMIT 0", table)
	}
}

// ============================================================
// RECENT SEARCHES
// ============================================================

// RecordSearch saves a user's search query for analytics and suggestions.
func (s *SearchEngine) RecordSearch(ctx context.Context, orgID, userID uuid.UUID, query string, resultCount int) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO recent_searches (organization_id, user_id, query, result_count)
		VALUES ($1, $2, $3, $4)
	`, orgID, userID, query, resultCount)
	return err
}

// RecordSearchClick records that a user clicked a search result.
func (s *SearchEngine) RecordSearchClick(ctx context.Context, orgID, userID uuid.UUID, query, entityType string, entityID uuid.UUID) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE recent_searches
		SET clicked_entity_type = $4, clicked_entity_id = $5
		WHERE id = (
			SELECT id FROM recent_searches
			WHERE organization_id = $1 AND user_id = $2 AND query = $3
			ORDER BY created_at DESC LIMIT 1
		)
	`, orgID, userID, query, entityType, entityID)
	return err
}

// GetRecentSearches returns a user's recent search queries.
func (s *SearchEngine) GetRecentSearches(ctx context.Context, orgID, userID uuid.UUID, limit int) ([]RecentSearch, error) {
	if limit < 1 || limit > 50 {
		limit = 10
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return nil, err
	}

	rows, err := tx.Query(ctx, `
		SELECT id, query, result_count, clicked_entity_type, clicked_entity_id, created_at
		FROM recent_searches
		WHERE organization_id = $1 AND user_id = $2
		ORDER BY created_at DESC
		LIMIT $3
	`, orgID, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []RecentSearch
	for rows.Next() {
		var r RecentSearch
		if err := rows.Scan(&r.ID, &r.Query, &r.ResultCount, &r.ClickedType, &r.ClickedEntityID, &r.CreatedAt); err != nil {
			continue
		}
		results = append(results, r)
	}

	tx.Commit(ctx)
	return results, nil
}

// GetSearchAnalytics returns popular and zero-result queries.
func (s *SearchEngine) GetSearchAnalytics(ctx context.Context, orgID uuid.UUID, days int) (map[string]interface{}, error) {
	if days < 1 {
		days = 30
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return nil, err
	}

	since := time.Now().AddDate(0, 0, -days)

	// Popular queries
	popularRows, err := tx.Query(ctx, `
		SELECT query, COUNT(*) AS cnt, AVG(result_count) AS avg_results
		FROM recent_searches
		WHERE organization_id = $1 AND created_at >= $2
		GROUP BY query ORDER BY cnt DESC LIMIT 20
	`, orgID, since)
	if err != nil {
		return nil, err
	}

	type popularQuery struct {
		Query      string  `json:"query"`
		Count      int     `json:"count"`
		AvgResults float64 `json:"avg_results"`
	}
	var popular []popularQuery
	for popularRows.Next() {
		var p popularQuery
		if err := popularRows.Scan(&p.Query, &p.Count, &p.AvgResults); err == nil {
			popular = append(popular, p)
		}
	}
	popularRows.Close()

	// Zero-result queries
	zeroRows, err := tx.Query(ctx, `
		SELECT query, COUNT(*) AS cnt
		FROM recent_searches
		WHERE organization_id = $1 AND created_at >= $2 AND result_count = 0
		GROUP BY query ORDER BY cnt DESC LIMIT 20
	`, orgID, since)
	if err != nil {
		return nil, err
	}

	type zeroQuery struct {
		Query string `json:"query"`
		Count int    `json:"count"`
	}
	var zero []zeroQuery
	for zeroRows.Next() {
		var z zeroQuery
		if err := zeroRows.Scan(&z.Query, &z.Count); err == nil {
			zero = append(zero, z)
		}
	}
	zeroRows.Close()

	tx.Commit(ctx)

	return map[string]interface{}{
		"popular_queries":     popular,
		"zero_result_queries": zero,
		"period_days":         days,
	}, nil
}

// GetIndexStats returns counts per entity type in the search index.
func (s *SearchEngine) GetIndexStats(ctx context.Context, orgID uuid.UUID) ([]IndexStats, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return nil, err
	}

	rows, err := tx.Query(ctx, `
		SELECT entity_type, COUNT(*), MAX(indexed_at)
		FROM search_index
		WHERE organization_id = $1
		GROUP BY entity_type
		ORDER BY entity_type
	`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []IndexStats
	for rows.Next() {
		var s IndexStats
		if err := rows.Scan(&s.EntityType, &s.IndexedCount, &s.LastIndexed); err == nil {
			stats = append(stats, s)
		}
	}

	tx.Commit(ctx)
	return stats, nil
}
