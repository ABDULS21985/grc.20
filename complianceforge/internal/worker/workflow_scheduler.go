// Package worker provides the WorkflowScheduler for periodic background tasks
// related to the compliance workflow engine: SLA breach detection, step escalation,
// and timer step processing.
package worker

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/complianceforge/platform/internal/service"
)

// WorkflowScheduler runs periodic checks for workflow SLA breaches, escalations,
// and timer-based step advancement.
type WorkflowScheduler struct {
	engine   *service.WorkflowEngine
	interval time.Duration
}

// NewWorkflowScheduler creates a new WorkflowScheduler that checks every 5 minutes.
func NewWorkflowScheduler(engine *service.WorkflowEngine) *WorkflowScheduler {
	return &WorkflowScheduler{
		engine:   engine,
		interval: 5 * time.Minute,
	}
}

// Start begins the periodic workflow check loop. Blocks until ctx is cancelled.
func (s *WorkflowScheduler) Start(ctx context.Context) {
	log.Info().Dur("interval", s.interval).Msg("Workflow scheduler started")

	// Run immediately on startup
	s.runChecks(ctx)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Workflow scheduler shutting down")
			return
		case <-ticker.C:
			s.runChecks(ctx)
		}
	}
}

// runChecks performs all periodic workflow checks in sequence.
func (s *WorkflowScheduler) runChecks(ctx context.Context) {
	startTime := time.Now()
	log.Debug().Msg("Workflow scheduler: starting periodic check")

	// 1. Check for SLA breaches and update statuses
	breached := s.checkSLABreaches(ctx)

	// 2. Escalate overdue steps that have been breached
	escalated := s.escalateOverdueSteps(ctx)

	// 3. Process timer steps whose delay has expired
	timersProcessed := s.processTimerSteps(ctx)

	duration := time.Since(startTime)
	if breached > 0 || escalated > 0 || timersProcessed > 0 {
		log.Info().
			Int("sla_breaches", breached).
			Int("escalated", escalated).
			Int("timers_processed", timersProcessed).
			Dur("duration", duration).
			Msg("Workflow scheduler: periodic check completed")
	} else {
		log.Debug().Dur("duration", duration).Msg("Workflow scheduler: periodic check completed (no actions)")
	}
}

// checkSLABreaches checks all active step executions for SLA deadline violations.
func (s *WorkflowScheduler) checkSLABreaches(ctx context.Context) int {
	breachedIDs, err := s.engine.CheckSLABreaches(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Workflow scheduler: SLA breach check failed")
		return 0
	}

	if len(breachedIDs) > 0 {
		log.Warn().
			Int("count", len(breachedIDs)).
			Msg("Workflow scheduler: SLA breaches detected")
	}

	return len(breachedIDs)
}

// escalateOverdueSteps finds step executions that have breached their SLA
// and triggers escalation if escalation users are configured.
func (s *WorkflowScheduler) escalateOverdueSteps(ctx context.Context) int {
	// Query breached steps that have not yet been escalated
	rows, err := s.engine.Pool().Query(ctx, `
		SELECT e.id, e.organization_id
		FROM workflow_step_executions e
		JOIN workflow_steps st ON st.id = e.workflow_step_id
		JOIN workflow_instances i ON i.id = e.workflow_instance_id
		WHERE e.status IN ('pending', 'in_progress')
			AND e.sla_status = 'breached'
			AND e.escalated_at IS NULL
			AND st.escalation_user_ids IS NOT NULL
			AND array_length(st.escalation_user_ids, 1) > 0
			AND i.status = 'active'
		LIMIT 50`)
	if err != nil {
		log.Error().Err(err).Msg("Workflow scheduler: failed to query overdue steps for escalation")
		return 0
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		var execID, orgID uuid.UUID
		if err := rows.Scan(&execID, &orgID); err != nil {
			log.Error().Err(err).Msg("Workflow scheduler: failed to scan overdue step")
			continue
		}

		if err := s.engine.EscalateStep(ctx, orgID, execID); err != nil {
			log.Error().Err(err).
				Str("execution_id", execID.String()).
				Msg("Workflow scheduler: failed to escalate step")
			continue
		}

		count++
	}

	return count
}

// processTimerSteps finds timer steps whose delay has expired and advances the workflow.
func (s *WorkflowScheduler) processTimerSteps(ctx context.Context) int {
	overdueTimers, err := s.engine.GetOverdueTimerSteps(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Workflow scheduler: failed to query overdue timer steps")
		return 0
	}

	var count int
	for i := range overdueTimers {
		exec := overdueTimers[i]
		if err := s.engine.ProcessTimerStep(ctx, exec); err != nil {
			log.Error().Err(err).
				Str("execution_id", exec.ID.String()).
				Msg("Workflow scheduler: failed to process timer step")
			continue
		}
		count++
	}

	return count
}
