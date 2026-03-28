// Package worker contains the Scheduler which runs periodic background tasks
// for evidence collection, compliance monitoring, drift detection,
// DSR deadline checks, NIS2 deadline checks, and report schedule processing.
package worker

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/complianceforge/platform/internal/service"
)

// Scheduler runs periodic background checks for Batch 3 modules.
type Scheduler struct {
	evidenceCollector  *service.EvidenceCollector
	complianceMonitor  *service.ComplianceMonitor
	driftDetector      *service.DriftDetector
	reportEngine       *service.ReportEngine
}

// NewScheduler creates a new periodic task scheduler.
func NewScheduler(
	ec *service.EvidenceCollector,
	cm *service.ComplianceMonitor,
	dd *service.DriftDetector,
	re *service.ReportEngine,
) *Scheduler {
	return &Scheduler{
		evidenceCollector:  ec,
		complianceMonitor:  cm,
		driftDetector:      dd,
		reportEngine:       re,
	}
}

// Start begins all periodic background tasks. Blocks until ctx is cancelled.
func (s *Scheduler) Start(ctx context.Context) {
	log.Info().Msg("Background scheduler started")

	// Evidence collection: check every 1 minute
	go s.runPeriodic(ctx, "evidence_collection", 1*time.Minute, func(ctx context.Context) {
		if s.evidenceCollector != nil {
			if err := s.evidenceCollector.RunScheduledCollections(ctx); err != nil {
				log.Error().Err(err).Msg("Evidence collection scheduler error")
			}
		}
	})

	// Compliance monitors: check every 5 minutes
	go s.runPeriodic(ctx, "compliance_monitors", 5*time.Minute, func(ctx context.Context) {
		if s.complianceMonitor != nil {
			if err := s.complianceMonitor.RunScheduledChecks(ctx); err != nil {
				log.Error().Err(err).Msg("Compliance monitor scheduler error")
			}
		}
	})

	// Drift detection: every 15 minutes
	go s.runPeriodic(ctx, "drift_detection", 15*time.Minute, func(ctx context.Context) {
		if s.driftDetector != nil {
			// Drift analysis runs per-org; for now we query active orgs
			s.runDriftAnalysisAllOrgs(ctx)
		}
	})

	// Report schedule check: every 5 minutes
	go s.runPeriodic(ctx, "report_schedules", 5*time.Minute, func(ctx context.Context) {
		if s.reportEngine != nil {
			s.processReportSchedules(ctx)
		}
	})

	// Block until context cancelled
	<-ctx.Done()
	log.Info().Msg("Background scheduler shutting down")
}

// runPeriodic runs a function at a fixed interval until the context is cancelled.
func (s *Scheduler) runPeriodic(ctx context.Context, name string, interval time.Duration, fn func(context.Context)) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Run immediately on startup
	fn(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fn(ctx)
		}
	}
}

// runDriftAnalysisAllOrgs queries active organisations and runs drift analysis on each.
func (s *Scheduler) runDriftAnalysisAllOrgs(ctx context.Context) {
	pool := s.complianceMonitor.Pool()
	rows, err := pool.Query(ctx, `SELECT id FROM organizations WHERE status = 'active' LIMIT 100`)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query orgs for drift analysis")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var orgID uuid.UUID
		if err := rows.Scan(&orgID); err != nil {
			continue
		}
		if err := s.driftDetector.Analyze(ctx, orgID); err != nil {
			log.Error().Err(err).Msg("Drift analysis error")
		}
	}
}

// processReportSchedules finds due report schedules and generates reports.
func (s *Scheduler) processReportSchedules(ctx context.Context) {
	schedules, err := s.reportEngine.GetDueSchedules(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get due report schedules")
		return
	}

	for i := range schedules {
		schedule := &schedules[i]
		log.Info().Str("schedule_id", schedule.ID.String()).Str("name", schedule.Name).Msg("Processing due report schedule")

		_, err := s.reportEngine.GenerateFromSchedule(ctx, schedule, schedule.RecipientUserIDs[0])
		if err != nil {
			log.Error().Err(err).Str("schedule_id", schedule.ID.String()).Msg("Failed to generate scheduled report")
			continue
		}

		if err := s.reportEngine.AdvanceSchedule(ctx, schedule); err != nil {
			log.Error().Err(err).Str("schedule_id", schedule.ID.String()).Msg("Failed to advance schedule")
		}
	}
}
