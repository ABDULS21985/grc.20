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
// ANALYTICS SCHEDULER — Periodic snapshot, trend, and benchmark jobs
// ============================================================

// AnalyticsScheduler runs periodic analytics jobs including daily snapshots,
// weekly/monthly/quarterly aggregations, trend calculations, and benchmarks.
type AnalyticsScheduler struct {
	engine *service.AnalyticsEngine
	pool   *pgxpool.Pool
}

// NewAnalyticsScheduler creates a new AnalyticsScheduler.
func NewAnalyticsScheduler(engine *service.AnalyticsEngine, pool *pgxpool.Pool) *AnalyticsScheduler {
	return &AnalyticsScheduler{
		engine: engine,
		pool:   pool,
	}
}

// Start begins the analytics scheduler. It runs periodic jobs on the following cadence:
//   - Daily at 00:00 UTC: Take daily snapshot, calculate compliance trends
//   - Weekly on Monday at 00:30 UTC: Take weekly snapshot
//   - Monthly on the 1st at 01:00 UTC: Take monthly snapshot, recalculate benchmarks
//   - Quarterly (Jan 1, Apr 1, Jul 1, Oct 1) at 02:00 UTC: Take quarterly snapshot
//
// Blocks until ctx is cancelled.
func (s *AnalyticsScheduler) Start(ctx context.Context) {
	log.Info().Msg("Analytics scheduler started")

	// Daily jobs: check every minute, fire at 00:00 UTC
	go s.runScheduled(ctx, "analytics_daily", 1*time.Minute, func(now time.Time) bool {
		return now.Hour() == 0 && now.Minute() == 0
	}, func(ctx context.Context) {
		if err := s.RunDailyJobs(ctx); err != nil {
			log.Error().Err(err).Msg("Analytics daily jobs failed")
		}
	})

	// Weekly jobs: check every minute, fire Monday at 00:30 UTC
	go s.runScheduled(ctx, "analytics_weekly", 1*time.Minute, func(now time.Time) bool {
		return now.Weekday() == time.Monday && now.Hour() == 0 && now.Minute() == 30
	}, func(ctx context.Context) {
		if err := s.RunWeeklyJobs(ctx); err != nil {
			log.Error().Err(err).Msg("Analytics weekly jobs failed")
		}
	})

	// Monthly jobs: check every minute, fire 1st of month at 01:00 UTC
	go s.runScheduled(ctx, "analytics_monthly", 1*time.Minute, func(now time.Time) bool {
		return now.Day() == 1 && now.Hour() == 1 && now.Minute() == 0
	}, func(ctx context.Context) {
		if err := s.runMonthlyJobs(ctx); err != nil {
			log.Error().Err(err).Msg("Analytics monthly jobs failed")
		}
	})

	// Quarterly jobs: check every minute, fire on quarter start at 02:00 UTC
	go s.runScheduled(ctx, "analytics_quarterly", 1*time.Minute, func(now time.Time) bool {
		return isQuarterStart(now) && now.Hour() == 2 && now.Minute() == 0
	}, func(ctx context.Context) {
		if err := s.runQuarterlyJobs(ctx); err != nil {
			log.Error().Err(err).Msg("Analytics quarterly jobs failed")
		}
	})

	<-ctx.Done()
	log.Info().Msg("Analytics scheduler shutting down")
}

// runScheduled checks the schedule condition at the given interval and runs
// the job function when the condition is met. It prevents duplicate runs
// within the same time window.
func (s *AnalyticsScheduler) runScheduled(ctx context.Context, name string, checkInterval time.Duration, shouldRun func(time.Time) bool, job func(context.Context)) {
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	var lastRun time.Time

	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			nowUTC := now.UTC()
			if shouldRun(nowUTC) {
				// Prevent running more than once per minute window
				if nowUTC.Sub(lastRun) < 2*time.Minute {
					continue
				}
				lastRun = nowUTC
				log.Info().Str("job", name).Msg("Running scheduled analytics job")
				job(ctx)
			}
		}
	}
}

// ============================================================
// JOB IMPLEMENTATIONS
// ============================================================

// RunDailyJobs takes a daily snapshot and calculates compliance trends
// for all active organizations.
func (s *AnalyticsScheduler) RunDailyJobs(ctx context.Context) error {
	log.Info().Msg("Running analytics daily jobs")

	orgIDs, err := s.getActiveOrgIDs(ctx)
	if err != nil {
		return err
	}

	for _, orgID := range orgIDs {
		// Take daily snapshot
		if err := s.engine.TakeSnapshot(ctx, orgID, "daily"); err != nil {
			log.Error().Err(err).Str("org_id", orgID.String()).Msg("Failed to take daily snapshot")
			continue
		}

		// Calculate compliance trends
		if err := s.engine.CalculateComplianceTrends(ctx, orgID); err != nil {
			log.Error().Err(err).Str("org_id", orgID.String()).Msg("Failed to calculate compliance trends")
			continue
		}
	}

	log.Info().Int("orgs_processed", len(orgIDs)).Msg("Analytics daily jobs complete")
	return nil
}

// RunWeeklyJobs takes a weekly snapshot for all active organizations.
func (s *AnalyticsScheduler) RunWeeklyJobs(ctx context.Context) error {
	log.Info().Msg("Running analytics weekly jobs")

	orgIDs, err := s.getActiveOrgIDs(ctx)
	if err != nil {
		return err
	}

	for _, orgID := range orgIDs {
		if err := s.engine.TakeSnapshot(ctx, orgID, "weekly"); err != nil {
			log.Error().Err(err).Str("org_id", orgID.String()).Msg("Failed to take weekly snapshot")
			continue
		}
	}

	log.Info().Int("orgs_processed", len(orgIDs)).Msg("Analytics weekly jobs complete")
	return nil
}

// runMonthlyJobs takes a monthly snapshot for all active organizations
// and recalculates anonymized benchmarks.
func (s *AnalyticsScheduler) runMonthlyJobs(ctx context.Context) error {
	log.Info().Msg("Running analytics monthly jobs")

	orgIDs, err := s.getActiveOrgIDs(ctx)
	if err != nil {
		return err
	}

	for _, orgID := range orgIDs {
		if err := s.engine.TakeSnapshot(ctx, orgID, "monthly"); err != nil {
			log.Error().Err(err).Str("org_id", orgID.String()).Msg("Failed to take monthly snapshot")
			continue
		}
	}

	// Recalculate anonymized benchmarks across all organizations
	if err := s.engine.CalculateBenchmarks(ctx); err != nil {
		log.Error().Err(err).Msg("Failed to calculate benchmarks")
	}

	log.Info().Int("orgs_processed", len(orgIDs)).Msg("Analytics monthly jobs complete")
	return nil
}

// runQuarterlyJobs takes a quarterly snapshot for all active organizations.
func (s *AnalyticsScheduler) runQuarterlyJobs(ctx context.Context) error {
	log.Info().Msg("Running analytics quarterly jobs")

	orgIDs, err := s.getActiveOrgIDs(ctx)
	if err != nil {
		return err
	}

	for _, orgID := range orgIDs {
		if err := s.engine.TakeSnapshot(ctx, orgID, "quarterly"); err != nil {
			log.Error().Err(err).Str("org_id", orgID.String()).Msg("Failed to take quarterly snapshot")
			continue
		}
	}

	log.Info().Int("orgs_processed", len(orgIDs)).Msg("Analytics quarterly jobs complete")
	return nil
}

// ============================================================
// HELPERS
// ============================================================

// getActiveOrgIDs returns all active organization IDs.
func (s *AnalyticsScheduler) getActiveOrgIDs(ctx context.Context) ([]uuid.UUID, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id FROM organizations
		WHERE status = 'active'
		ORDER BY id
		LIMIT 500`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			continue
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// isQuarterStart returns true if the date is the first day of a calendar quarter.
func isQuarterStart(t time.Time) bool {
	return t.Day() == 1 && (t.Month() == time.January || t.Month() == time.April ||
		t.Month() == time.July || t.Month() == time.October)
}
