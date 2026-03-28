// Package worker processes background jobs from the Redis queue.
package worker

import (
	"context"
	"encoding/json"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/complianceforge/platform/internal/pkg/email"
	"github.com/complianceforge/platform/internal/pkg/queue"
	"github.com/complianceforge/platform/internal/service"
)

// Processor handles background job execution.
type Processor struct {
	queue            *queue.Queue
	emailSender      *email.Sender
	complianceEngine *service.ComplianceEngine
	reportingSvc     *service.ReportingService
}

// NewProcessor creates a new job processor.
func NewProcessor(
	q *queue.Queue,
	emailSender *email.Sender,
	complianceEngine *service.ComplianceEngine,
	reportingSvc *service.ReportingService,
) *Processor {
	return &Processor{
		queue:            q,
		emailSender:      emailSender,
		complianceEngine: complianceEngine,
		reportingSvc:     reportingSvc,
	}
}

// Start begins processing jobs from all queues in priority order.
// It blocks until the context is cancelled.
func (p *Processor) Start(ctx context.Context) {
	log.Info().Msg("Job processor started")

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Job processor shutting down")
			return
		default:
			// Poll queues in priority order: critical → high → default → low
			job, err := p.queue.Dequeue(ctx, 5*time.Second,
				queue.QueueCritical,
				queue.QueueHigh,
				queue.QueueDefault,
				queue.QueueLow,
			)
			if err != nil {
				log.Error().Err(err).Msg("Failed to dequeue job")
				time.Sleep(1 * time.Second)
				continue
			}
			if job == nil {
				continue // Timeout, no jobs available
			}

			p.processJob(ctx, job)
		}
	}
}

func (p *Processor) processJob(ctx context.Context, job *queue.Job) {
	startTime := time.Now()
	logger := log.With().
		Str("job_id", job.ID).
		Str("type", job.Type).
		Str("org_id", job.OrganizationID).
		Int("attempt", job.Attempts).
		Logger()

	logger.Info().Msg("Processing job")

	var err error
	switch job.Type {
	case queue.JobTypeSendEmail:
		err = p.handleSendEmail(ctx, job)
	case queue.JobTypeBreachDeadlineAlert:
		err = p.handleBreachAlert(ctx, job)
	case queue.JobTypeGenerateReport:
		err = p.handleGenerateReport(ctx, job)
	case queue.JobTypeRecalculateScore:
		err = p.handleRecalculateScore(ctx, job)
	case queue.JobTypePolicyReviewReminder:
		err = p.handlePolicyReviewReminder(ctx, job)
	case queue.JobTypeFindingEscalation:
		err = p.handleFindingEscalation(ctx, job)
	case queue.JobTypeVendorAssessmentDue:
		err = p.handleVendorAssessment(ctx, job)
	case queue.JobTypeAttestationReminder:
		err = p.handleAttestationReminder(ctx, job)
	default:
		logger.Warn().Msg("Unknown job type — skipping")
		return
	}

	duration := time.Since(startTime)

	if err != nil {
		logger.Error().Err(err).Dur("duration", duration).Msg("Job failed")
		if retryErr := p.queue.Retry(ctx, job); retryErr != nil {
			logger.Error().Err(retryErr).Msg("Failed to retry job")
		}
		return
	}

	logger.Info().Dur("duration", duration).Msg("Job completed successfully")
}

// ============================================================
// JOB HANDLERS
// ============================================================

func (p *Processor) handleSendEmail(ctx context.Context, job *queue.Job) error {
	var payload queue.EmailPayload
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return err
	}

	if payload.Template != "" {
		return p.emailSender.SendTemplated(payload.To, payload.Subject, payload.Template, payload.Data)
	}

	return p.emailSender.Send(payload.To, payload.Subject, payload.Body)
}

func (p *Processor) handleBreachAlert(ctx context.Context, job *queue.Job) error {
	var payload queue.EmailPayload
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return err
	}

	log.Warn().
		Str("org_id", job.OrganizationID).
		Strs("to", payload.To).
		Msg("CRITICAL: Sending GDPR breach deadline alert")

	return p.emailSender.SendTemplated(payload.To, payload.Subject, "breach_alert", payload.Data)
}

func (p *Processor) handleGenerateReport(ctx context.Context, job *queue.Job) error {
	var payload queue.ReportPayload
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return err
	}

	log.Info().
		Str("report_type", payload.ReportType).
		Str("format", payload.Format).
		Str("user_id", payload.UserID).
		Msg("Generating report")

	// Report generation happens here
	// In production, generate PDF/XLSX and store in file storage
	// Then notify the user that the report is ready

	return nil
}

func (p *Processor) handleRecalculateScore(ctx context.Context, job *queue.Job) error {
	var payload queue.ScorePayload
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return err
	}

	log.Info().
		Str("org_framework_id", payload.OrgFrameworkID).
		Msg("Recalculating compliance score")

	// Score recalculation handled by the compliance engine
	// p.complianceEngine.CalculateFrameworkScore(ctx, orgID, orgFrameworkID)

	return nil
}

func (p *Processor) handlePolicyReviewReminder(ctx context.Context, job *queue.Job) error {
	var payload queue.EmailPayload
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return err
	}

	return p.emailSender.SendTemplated(payload.To, payload.Subject, "policy_review", payload.Data)
}

func (p *Processor) handleFindingEscalation(ctx context.Context, job *queue.Job) error {
	var payload queue.EmailPayload
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return err
	}

	return p.emailSender.SendTemplated(payload.To, payload.Subject, "finding_escalation", payload.Data)
}

func (p *Processor) handleVendorAssessment(ctx context.Context, job *queue.Job) error {
	var payload queue.EmailPayload
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return err
	}

	return p.emailSender.SendTemplated(payload.To, payload.Subject, "vendor_assessment_due", payload.Data)
}

func (p *Processor) handleAttestationReminder(ctx context.Context, job *queue.Job) error {
	var payload queue.EmailPayload
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return err
	}

	return p.emailSender.SendTemplated(payload.To, payload.Subject, "attestation_reminder", payload.Data)
}
