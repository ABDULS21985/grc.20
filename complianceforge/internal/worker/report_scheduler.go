// Package worker — report_scheduler.go runs scheduled report generation.
// Checks report_schedules every minute for due reports (next_run_at <= NOW()
// AND is_active), generates them, and advances the schedule.
package worker

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/complianceforge/platform/internal/service"
)

// ReportScheduler periodically checks for due report schedules and
// triggers report generation. It runs as a background goroutine
// alongside the main job processor.
type ReportScheduler struct {
	engine   *service.ReportEngine
	interval time.Duration
}

// NewReportScheduler creates a new scheduler that checks for due reports
// at the specified interval (typically every 60 seconds).
func NewReportScheduler(engine *service.ReportEngine) *ReportScheduler {
	return &ReportScheduler{
		engine:   engine,
		interval: 1 * time.Minute,
	}
}

// Start begins the scheduler loop. It blocks until the context is cancelled.
// Should be launched as a goroutine: go scheduler.Start(ctx)
func (s *ReportScheduler) Start(ctx context.Context) {
	log.Info().Dur("interval", s.interval).Msg("Report scheduler started")

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	// Run immediately on start, then on each tick
	s.tick(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Report scheduler shutting down")
			return
		case <-ticker.C:
			s.tick(ctx)
		}
	}
}

// tick performs a single scheduler pass: finds due schedules and generates reports.
func (s *ReportScheduler) tick(ctx context.Context) {
	schedules, err := s.engine.GetDueSchedules(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch due report schedules")
		return
	}

	if len(schedules) == 0 {
		return
	}

	log.Info().Int("count", len(schedules)).Msg("Processing due report schedules")

	for i := range schedules {
		schedule := &schedules[i]

		logger := log.With().
			Str("schedule_id", schedule.ID.String()).
			Str("schedule_name", schedule.Name).
			Str("frequency", schedule.Frequency).
			Str("org_id", schedule.OrganizationID.String()).
			Logger()

		logger.Info().Msg("Executing scheduled report generation")

		// Use a system user ID for scheduled reports. If recipient_user_ids
		// is populated, use the first one as the "generated_by" user.
		generatedBy := systemUserID(schedule)

		run, genErr := s.engine.GenerateFromSchedule(ctx, schedule, generatedBy)
		if genErr != nil {
			logger.Error().Err(genErr).Msg("Scheduled report generation failed")
			// Still advance the schedule so it does not get stuck
		} else {
			logger.Info().
				Str("run_id", run.ID.String()).
				Str("status", run.Status).
				Msg("Scheduled report generated")
		}

		// Advance the schedule: update last_run_at and calculate next_run_at
		if advErr := s.engine.AdvanceSchedule(ctx, schedule); advErr != nil {
			logger.Error().Err(advErr).Msg("Failed to advance schedule")
		} else {
			logger.Info().Msg("Schedule advanced to next run")
		}
	}
}

// systemUserID returns the user ID to attribute the scheduled generation to.
// Prefers the first recipient user; falls back to a nil UUID representing the system.
func systemUserID(schedule *service.ReportSchedule) uuid.UUID {
	if len(schedule.RecipientUserIDs) > 0 {
		return schedule.RecipientUserIDs[0]
	}
	// Return the report_definition creator as a fallback would require
	// an extra query. For now, use a nil UUID that the engine interprets
	// as "System".
	return uuid.Nil
}
