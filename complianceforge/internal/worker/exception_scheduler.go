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
// EXCEPTION SCHEDULER — Periodic expiry and review checks
// ============================================================

// ExceptionScheduler runs periodic checks for expiring exceptions
// and overdue reviews across all active organizations.
type ExceptionScheduler struct {
	svc  *service.ExceptionService
	pool *pgxpool.Pool
}

// NewExceptionScheduler creates a new ExceptionScheduler.
func NewExceptionScheduler(svc *service.ExceptionService, pool *pgxpool.Pool) *ExceptionScheduler {
	return &ExceptionScheduler{
		svc:  svc,
		pool: pool,
	}
}

// Start begins the exception scheduler. It runs:
//   - Daily at 01:00 UTC: Check for expiring and expired exceptions
//   - Daily at 01:30 UTC: Check for overdue reviews
//
// Blocks until ctx is cancelled.
func (s *ExceptionScheduler) Start(ctx context.Context) {
	log.Info().Msg("Exception scheduler started")

	// Daily expiry check at 01:00 UTC
	go s.runScheduled(ctx, "exception_daily_expiry", 1*time.Minute, func(now time.Time) bool {
		return now.Hour() == 1 && now.Minute() == 0
	}, func(ctx context.Context) {
		if err := s.RunDailyCheck(ctx); err != nil {
			log.Error().Err(err).Msg("Exception daily expiry check failed")
		}
	})

	// Daily overdue review check at 01:30 UTC
	go s.runScheduled(ctx, "exception_overdue_reviews", 1*time.Minute, func(now time.Time) bool {
		return now.Hour() == 1 && now.Minute() == 30
	}, func(ctx context.Context) {
		if err := s.CheckOverdueReviews(ctx); err != nil {
			log.Error().Err(err).Msg("Exception overdue review check failed")
		}
	})

	<-ctx.Done()
	log.Info().Msg("Exception scheduler shutting down")
}

// runScheduled checks the schedule condition at the given interval and runs
// the job function when the condition is met.
func (s *ExceptionScheduler) runScheduled(ctx context.Context, name string, checkInterval time.Duration, shouldRun func(time.Time) bool, job func(context.Context)) {
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
				if nowUTC.Sub(lastRun) < 2*time.Minute {
					continue
				}
				lastRun = nowUTC
				log.Info().Str("job", name).Msg("Running scheduled exception job")
				job(ctx)
			}
		}
	}
}

// ============================================================
// JOB IMPLEMENTATIONS
// ============================================================

// RunDailyCheck finds expiring exceptions and auto-expires past-due ones.
// It checks for exceptions expiring at 30d, 14d, 7d, and 1d thresholds
// and logs notifications. It also transitions expired exceptions.
func (s *ExceptionScheduler) RunDailyCheck(ctx context.Context) error {
	log.Info().Msg("Running exception daily expiry check")

	orgIDs, err := s.getActiveOrgIDs(ctx)
	if err != nil {
		return err
	}

	totalExpired := int64(0)
	totalExpiring := 0

	for _, orgID := range orgIDs {
		// Auto-expire past-due exceptions
		expired, err := s.svc.AutoExpireExceptions(ctx, orgID)
		if err != nil {
			log.Error().Err(err).Str("org_id", orgID.String()).Msg("Failed to auto-expire exceptions")
			continue
		}
		totalExpired += expired

		// Check for expiring exceptions at notification thresholds
		thresholds := []int{30, 14, 7, 1}
		for _, days := range thresholds {
			expiring, err := s.svc.GetExpiringExceptions(ctx, orgID, days)
			if err != nil {
				log.Error().Err(err).
					Str("org_id", orgID.String()).
					Int("days", days).
					Msg("Failed to check expiring exceptions")
				continue
			}

			for _, exc := range expiring {
				if exc.ExpiryDate == nil {
					continue
				}
				daysUntilExpiry := int(time.Until(*exc.ExpiryDate).Hours() / 24)

				// Only log for the matching threshold window
				if daysUntilExpiry > days {
					continue
				}

				// Find the tightest threshold that matches
				matchedThreshold := 0
				for _, t := range thresholds {
					if daysUntilExpiry <= t {
						matchedThreshold = t
					}
				}
				if matchedThreshold != days {
					continue
				}

				log.Warn().
					Str("org_id", orgID.String()).
					Str("exception_ref", exc.ExceptionRef).
					Str("title", exc.Title).
					Int("days_until_expiry", daysUntilExpiry).
					Int("threshold_days", days).
					Msg("Exception expiring soon")
				totalExpiring++
			}
		}
	}

	log.Info().
		Int64("auto_expired", totalExpired).
		Int("expiring_notifications", totalExpiring).
		Int("orgs_processed", len(orgIDs)).
		Msg("Exception daily expiry check complete")

	return nil
}

// CheckOverdueReviews finds exceptions that are overdue for their periodic review.
func (s *ExceptionScheduler) CheckOverdueReviews(ctx context.Context) error {
	log.Info().Msg("Running exception overdue review check")

	orgIDs, err := s.getActiveOrgIDs(ctx)
	if err != nil {
		return err
	}

	totalOverdue := 0

	for _, orgID := range orgIDs {
		overdue, err := s.svc.GetOverdueReviews(ctx, orgID)
		if err != nil {
			log.Error().Err(err).Str("org_id", orgID.String()).Msg("Failed to check overdue reviews")
			continue
		}

		for _, exc := range overdue {
			daysPast := 0
			if exc.NextReviewDate != nil {
				daysPast = int(time.Since(*exc.NextReviewDate).Hours() / 24)
			}

			log.Warn().
				Str("org_id", orgID.String()).
				Str("exception_ref", exc.ExceptionRef).
				Str("title", exc.Title).
				Int("days_overdue", daysPast).
				Msg("Exception overdue for review")
			totalOverdue++
		}
	}

	log.Info().
		Int("total_overdue", totalOverdue).
		Int("orgs_processed", len(orgIDs)).
		Msg("Exception overdue review check complete")

	return nil
}

// ============================================================
// HELPERS
// ============================================================

// getActiveOrgIDs returns all active organization IDs.
func (s *ExceptionScheduler) getActiveOrgIDs(ctx context.Context) ([]uuid.UUID, error) {
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
