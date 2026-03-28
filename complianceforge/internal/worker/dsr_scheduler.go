// Package worker provides the DSR SLA scheduler for background compliance monitoring.
// The scheduler runs daily to update SLA statuses and emit notifications for DSR requests
// approaching or past their GDPR response deadlines.
package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"github.com/complianceforge/platform/internal/pkg/queue"
	"github.com/complianceforge/platform/internal/service"
)

// DSRScheduler runs daily SLA compliance checks for all active DSR requests.
type DSRScheduler struct {
	pool     *pgxpool.Pool
	dsrSvc   *service.DSRService
	jobQueue *queue.Queue
	interval time.Duration
}

// NewDSRScheduler creates a new DSR SLA scheduler.
func NewDSRScheduler(pool *pgxpool.Pool, dsrSvc *service.DSRService, jobQueue *queue.Queue) *DSRScheduler {
	return &DSRScheduler{
		pool:     pool,
		dsrSvc:   dsrSvc,
		jobQueue: jobQueue,
		interval: 24 * time.Hour,
	}
}

// Start begins the daily SLA check loop. It blocks until the context is cancelled.
func (s *DSRScheduler) Start(ctx context.Context) {
	log.Info().Dur("interval", s.interval).Msg("DSR SLA scheduler started")

	// Run immediately on start, then on the interval
	s.runCheck(ctx)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("DSR SLA scheduler shutting down")
			return
		case <-ticker.C:
			s.runCheck(ctx)
		}
	}
}

// runCheck performs a single SLA compliance check across all organizations.
func (s *DSRScheduler) runCheck(ctx context.Context) {
	startTime := time.Now()
	log.Info().Msg("DSR SLA check: starting daily run")

	// Get all active organizations that have DSR requests
	orgIDs, err := s.getActiveOrganizations(ctx)
	if err != nil {
		log.Error().Err(err).Msg("DSR SLA check: failed to get organizations")
		return
	}

	var totalUpdated int64
	var totalNotifications int

	for _, orgID := range orgIDs {
		// Update SLA statuses for all active DSRs in this organization
		updated, err := s.dsrSvc.UpdateSLAStatuses(ctx, orgID)
		if err != nil {
			log.Error().Err(err).Str("org_id", orgID.String()).Msg("DSR SLA check: failed to update statuses")
			continue
		}
		totalUpdated += updated

		// Emit notifications for at-risk and overdue requests
		notifications, err := s.emitNotifications(ctx, orgID)
		if err != nil {
			log.Error().Err(err).Str("org_id", orgID.String()).Msg("DSR SLA check: failed to emit notifications")
			continue
		}
		totalNotifications += notifications
	}

	duration := time.Since(startTime)
	log.Info().
		Int("organizations", len(orgIDs)).
		Int64("requests_updated", totalUpdated).
		Int("notifications_sent", totalNotifications).
		Dur("duration", duration).
		Msg("DSR SLA check: daily run completed")
}

// getActiveOrganizations returns organization IDs that have active (non-terminal) DSR requests.
func (s *DSRScheduler) getActiveOrganizations(ctx context.Context) ([]uuid.UUID, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT DISTINCT organization_id FROM dsr_requests
		WHERE deleted_at IS NULL
			AND status NOT IN ('completed', 'rejected', 'withdrawn')`)
	if err != nil {
		return nil, fmt.Errorf("failed to query active organizations: %w", err)
	}
	defer rows.Close()

	var orgIDs []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan org ID: %w", err)
		}
		orgIDs = append(orgIDs, id)
	}

	return orgIDs, nil
}

// emitNotifications sends alerts for DSR requests at specific SLA thresholds.
// Notifications are emitted at: 7 days remaining, 3 days remaining, and overdue.
func (s *DSRScheduler) emitNotifications(ctx context.Context, orgID uuid.UUID) (int, error) {
	now := time.Now()
	count := 0

	// Query active DSRs with their effective deadlines
	rows, err := s.pool.Query(ctx, `
		SELECT
			id, request_ref, request_type::text, status::text,
			response_deadline, extended_deadline,
			COALESCE(days_remaining, 0),
			assigned_to
		FROM dsr_requests
		WHERE organization_id = $1
			AND deleted_at IS NULL
			AND status NOT IN ('completed', 'rejected', 'withdrawn')
		ORDER BY COALESCE(extended_deadline, response_deadline) ASC`,
		orgID,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to query DSR requests: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			requestID        uuid.UUID
			requestRef       string
			requestType      string
			status           string
			responseDeadline time.Time
			extendedDeadline *time.Time
			daysRemaining    int
			assignedTo       *uuid.UUID
		)

		if err := rows.Scan(
			&requestID, &requestRef, &requestType, &status,
			&responseDeadline, &extendedDeadline,
			&daysRemaining, &assignedTo,
		); err != nil {
			log.Error().Err(err).Msg("DSR SLA check: failed to scan request")
			continue
		}

		// Suppress unused variable warnings
		_ = status
		_ = assignedTo

		// Determine effective deadline
		effectiveDeadline := responseDeadline
		if extendedDeadline != nil {
			effectiveDeadline = *extendedDeadline
		}

		// Recalculate days remaining (in case the DB value is stale)
		actualDaysRemaining := int(effectiveDeadline.Sub(now).Hours() / 24)

		// Emit notifications at threshold points
		switch {
		case actualDaysRemaining < 0:
			// OVERDUE
			if err := s.sendDSRAlert(ctx, orgID.String(), requestRef, requestType, "overdue",
				fmt.Sprintf("DSR %s is OVERDUE by %d day(s). Immediate action required.", requestRef, -actualDaysRemaining),
				effectiveDeadline.Format("2006-01-02")); err != nil {
				log.Error().Err(err).Str("ref", requestRef).Msg("Failed to send overdue alert")
			} else {
				count++
			}

		case actualDaysRemaining <= 3:
			// 3 DAYS REMAINING - CRITICAL WARNING
			if err := s.sendDSRAlert(ctx, orgID.String(), requestRef, requestType, "critical_warning",
				fmt.Sprintf("DSR %s has only %d day(s) remaining. Deadline: %s", requestRef, actualDaysRemaining, effectiveDeadline.Format("2006-01-02")),
				effectiveDeadline.Format("2006-01-02")); err != nil {
				log.Error().Err(err).Str("ref", requestRef).Msg("Failed to send critical warning")
			} else {
				count++
			}

		case actualDaysRemaining <= 7:
			// 7 DAYS REMAINING - WARNING
			if err := s.sendDSRAlert(ctx, orgID.String(), requestRef, requestType, "warning",
				fmt.Sprintf("DSR %s has %d day(s) remaining until deadline: %s", requestRef, actualDaysRemaining, effectiveDeadline.Format("2006-01-02")),
				effectiveDeadline.Format("2006-01-02")); err != nil {
				log.Error().Err(err).Str("ref", requestRef).Msg("Failed to send warning")
			} else {
				count++
			}
		}
	}

	return count, nil
}

// sendDSRAlert enqueues a notification job for a DSR SLA alert.
func (s *DSRScheduler) sendDSRAlert(ctx context.Context, orgID, requestRef, requestType, alertLevel, message, deadline string) error {
	subject := fmt.Sprintf("[DSR %s] %s - %s", alertLevel, requestRef, requestType)

	payload := queue.EmailPayload{
		Subject:  subject,
		Template: "dsr_sla_alert",
		Data: map[string]interface{}{
			"request_ref":  requestRef,
			"request_type": requestType,
			"alert_level":  alertLevel,
			"message":      message,
			"deadline":     deadline,
		},
	}

	queueName := queue.QueueDefault
	if alertLevel == "overdue" || alertLevel == "critical_warning" {
		queueName = queue.QueueHigh
	}

	_, err := s.jobQueue.Enqueue(ctx, "dsr_sla_alert", queueName, orgID, payload)
	return err
}
