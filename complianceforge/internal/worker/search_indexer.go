package worker

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"github.com/complianceforge/platform/internal/service"
)

// ============================================================
// SEARCH INDEXER
// Background worker that listens for entity change events and
// maintains the search index. Runs a nightly full re-index at
// 03:00 UTC and provides an index health check.
// ============================================================

// SearchIndexer manages background search index maintenance.
type SearchIndexer struct {
	pool   *pgxpool.Pool
	engine *service.SearchEngine
}

// NewSearchIndexer creates a new SearchIndexer.
func NewSearchIndexer(pool *pgxpool.Pool, engine *service.SearchEngine) *SearchIndexer {
	return &SearchIndexer{pool: pool, engine: engine}
}

// HandleEntityChange processes entity change events and re-indexes
// the affected entity in the search index.
func (w *SearchIndexer) HandleEntityChange(ctx context.Context, orgID uuid.UUID, entityType string, entityID uuid.UUID) {
	rec := service.IndexRecord{
		EntityType: entityType,
		EntityID:   entityID,
	}
	if err := w.engine.IndexEntity(ctx, orgID, rec); err != nil {
		log.Warn().Err(err).
			Str("entity_type", entityType).
			Str("entity_id", entityID.String()).
			Msg("search indexer: failed to re-index entity")
	}
}

// HandleEntityDelete removes a deleted entity from the search index.
func (w *SearchIndexer) HandleEntityDelete(ctx context.Context, orgID uuid.UUID, entityType string, entityID uuid.UUID) {
	if err := w.engine.RemoveEntity(ctx, orgID, entityType, entityID); err != nil {
		log.Warn().Err(err).
			Str("entity_type", entityType).
			Str("entity_id", entityID.String()).
			Msg("search indexer: failed to remove entity from index")
	}
}

// StartNightlyReindex runs a nightly full re-index at 03:00 UTC.
// It indexes all organisations' entities to catch any missed updates.
func (w *SearchIndexer) StartNightlyReindex(ctx context.Context) {
	log.Info().Msg("search indexer: nightly re-index scheduler started")
	for {
		now := time.Now().UTC()
		next := time.Date(now.Year(), now.Month(), now.Day()+1, 3, 0, 0, 0, time.UTC)
		if now.Hour() < 3 {
			next = time.Date(now.Year(), now.Month(), now.Day(), 3, 0, 0, 0, time.UTC)
		}

		timer := time.NewTimer(time.Until(next))
		select {
		case <-ctx.Done():
			timer.Stop()
			log.Info().Msg("search indexer: nightly re-index scheduler stopped")
			return
		case <-timer.C:
			w.runFullReindex(ctx)
		}
	}
}

func (w *SearchIndexer) runFullReindex(ctx context.Context) {
	log.Info().Msg("search indexer: starting nightly full re-index")

	// Get all active organisation IDs
	rows, err := w.pool.Query(ctx, "SELECT id FROM organizations WHERE is_active = true")
	if err != nil {
		log.Error().Err(err).Msg("search indexer: failed to list organisations")
		return
	}
	defer rows.Close()

	var orgIDs []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err == nil {
			orgIDs = append(orgIDs, id)
		}
	}

	for _, orgID := range orgIDs {
		report, err := w.engine.IndexAllEntities(ctx, orgID)
		if err != nil {
			log.Error().Err(err).Str("org_id", orgID.String()).
				Msg("search indexer: full re-index failed for org")
			continue
		}
		log.Info().
			Str("org_id", orgID.String()).
			Int("indexed", report.TotalIndexed).
			Int("errors", report.TotalErrors).
			Int64("duration_ms", report.DurationMs).
			Msg("search indexer: org re-index complete")
	}

	log.Info().Int("org_count", len(orgIDs)).Msg("search indexer: nightly full re-index complete")
}

// RunHealthCheck compares entity counts in source tables against
// the search index to detect index drift.
func (w *SearchIndexer) RunHealthCheck(ctx context.Context, orgID uuid.UUID) ([]IndexHealthEntry, error) {
	entityTypes := map[string]string{
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

	var entries []IndexHealthEntry
	for entityType, table := range entityTypes {
		var sourceCount, indexCount int

		err := w.pool.QueryRow(ctx,
			"SELECT COUNT(*) FROM "+table+" WHERE organization_id = $1", orgID,
		).Scan(&sourceCount)
		if err != nil {
			// Table may not exist yet
			sourceCount = -1
		}

		err = w.pool.QueryRow(ctx,
			"SELECT COUNT(*) FROM search_index WHERE organization_id = $1 AND entity_type = $2",
			orgID, entityType,
		).Scan(&indexCount)
		if err != nil {
			indexCount = -1
		}

		entries = append(entries, IndexHealthEntry{
			EntityType:  entityType,
			SourceCount: sourceCount,
			IndexCount:  indexCount,
			InSync:      sourceCount == indexCount,
		})
	}

	return entries, nil
}

// IndexHealthEntry holds health info for one entity type.
type IndexHealthEntry struct {
	EntityType  string `json:"entity_type"`
	SourceCount int    `json:"source_count"`
	IndexCount  int    `json:"index_count"`
	InSync      bool   `json:"in_sync"`
}
