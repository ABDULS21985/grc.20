// Package queue provides a Redis-backed job queue for asynchronous task processing.
// Used for email notifications, report generation, compliance score recalculation,
// and other background tasks that should not block API responses.
package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// Job types supported by the queue.
const (
	JobTypeSendEmail              = "send_email"
	JobTypeGenerateReport         = "generate_report"
	JobTypeRecalculateScore       = "recalculate_compliance_score"
	JobTypePolicyReviewReminder   = "policy_review_reminder"
	JobTypeBreachDeadlineAlert    = "breach_deadline_alert"
	JobTypeFindingEscalation      = "finding_escalation"
	JobTypeVendorAssessmentDue    = "vendor_assessment_due"
	JobTypeAttestationReminder    = "attestation_reminder"
	JobTypeAuditLogExport         = "audit_log_export"
)

// Queue names for priority-based processing.
const (
	QueueCritical = "complianceforge:queue:critical" // breach alerts, security incidents
	QueueHigh     = "complianceforge:queue:high"     // reports, score recalculation
	QueueDefault  = "complianceforge:queue:default"  // emails, reminders
	QueueLow      = "complianceforge:queue:low"      // exports, cleanup
)

// Job represents a unit of work to be processed asynchronously.
type Job struct {
	ID             string          `json:"id"`
	Type           string          `json:"type"`
	Queue          string          `json:"queue"`
	OrganizationID string          `json:"organization_id"`
	Payload        json.RawMessage `json:"payload"`
	Attempts       int             `json:"attempts"`
	MaxAttempts    int             `json:"max_attempts"`
	CreatedAt      time.Time       `json:"created_at"`
	ScheduledAt    *time.Time      `json:"scheduled_at,omitempty"`
}

// EmailPayload is the payload for email jobs.
type EmailPayload struct {
	To       []string `json:"to"`
	Subject  string   `json:"subject"`
	Body     string   `json:"body"`
	Template string   `json:"template,omitempty"`
	Data     map[string]interface{} `json:"data,omitempty"`
}

// ReportPayload is the payload for report generation jobs.
type ReportPayload struct {
	ReportType string `json:"report_type"` // compliance, risk, audit, executive
	Format     string `json:"format"`      // pdf, xlsx, csv
	UserID     string `json:"user_id"`
	Filters    map[string]interface{} `json:"filters,omitempty"`
}

// ScorePayload is the payload for compliance score recalculation.
type ScorePayload struct {
	OrgFrameworkID string `json:"org_framework_id"`
}

// Queue manages the Redis-backed job queue.
type Queue struct {
	rdb *redis.Client
}

// New creates a new Queue client.
func New(rdb *redis.Client) *Queue {
	return &Queue{rdb: rdb}
}

// NewFromAddr creates a Queue client from Redis address and password.
func NewFromAddr(addr, password string, db int) (*Queue, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis for queue: %w", err)
	}

	return &Queue{rdb: rdb}, nil
}

// Enqueue adds a job to the specified queue.
func (q *Queue) Enqueue(ctx context.Context, jobType, queueName, orgID string, payload interface{}) (*Job, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	job := &Job{
		ID:             uuid.New().String(),
		Type:           jobType,
		Queue:          queueName,
		OrganizationID: orgID,
		Payload:        payloadBytes,
		Attempts:       0,
		MaxAttempts:    3,
		CreatedAt:      time.Now(),
	}

	jobBytes, err := json.Marshal(job)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal job: %w", err)
	}

	if err := q.rdb.LPush(ctx, queueName, jobBytes).Err(); err != nil {
		return nil, fmt.Errorf("failed to enqueue job: %w", err)
	}

	log.Debug().
		Str("job_id", job.ID).
		Str("type", jobType).
		Str("queue", queueName).
		Msg("Job enqueued")

	return job, nil
}

// EnqueueEmail enqueues an email job.
func (q *Queue) EnqueueEmail(ctx context.Context, orgID string, to []string, subject, body string) (*Job, error) {
	payload := EmailPayload{
		To:      to,
		Subject: subject,
		Body:    body,
	}
	return q.Enqueue(ctx, JobTypeSendEmail, QueueDefault, orgID, payload)
}

// EnqueueBreachAlert enqueues a critical breach deadline alert.
func (q *Queue) EnqueueBreachAlert(ctx context.Context, orgID string, to []string, incidentRef, deadline string) (*Job, error) {
	payload := EmailPayload{
		To:       to,
		Subject:  fmt.Sprintf("[URGENT] GDPR Breach Notification Deadline — %s", incidentRef),
		Template: "breach_alert",
		Data: map[string]interface{}{
			"incident_ref": incidentRef,
			"deadline":     deadline,
		},
	}
	return q.Enqueue(ctx, JobTypeBreachDeadlineAlert, QueueCritical, orgID, payload)
}

// EnqueueReportGeneration enqueues a report generation job.
func (q *Queue) EnqueueReportGeneration(ctx context.Context, orgID, reportType, format, userID string) (*Job, error) {
	payload := ReportPayload{
		ReportType: reportType,
		Format:     format,
		UserID:     userID,
	}
	return q.Enqueue(ctx, JobTypeGenerateReport, QueueHigh, orgID, payload)
}

// EnqueueScoreRecalculation enqueues a compliance score recalculation.
func (q *Queue) EnqueueScoreRecalculation(ctx context.Context, orgID, orgFrameworkID string) (*Job, error) {
	payload := ScorePayload{
		OrgFrameworkID: orgFrameworkID,
	}
	return q.Enqueue(ctx, JobTypeRecalculateScore, QueueHigh, orgID, payload)
}

// Dequeue retrieves and removes the next job from the queue.
// Uses BRPOP for blocking wait (up to timeout).
func (q *Queue) Dequeue(ctx context.Context, timeout time.Duration, queues ...string) (*Job, error) {
	result, err := q.rdb.BRPop(ctx, timeout, queues...).Result()
	if err == redis.Nil {
		return nil, nil // Timeout, no jobs available
	}
	if err != nil {
		return nil, fmt.Errorf("dequeue error: %w", err)
	}

	// result[0] = queue name, result[1] = job data
	var job Job
	if err := json.Unmarshal([]byte(result[1]), &job); err != nil {
		return nil, fmt.Errorf("failed to unmarshal job: %w", err)
	}

	job.Attempts++
	return &job, nil
}

// Retry re-enqueues a failed job with incremented attempt count.
func (q *Queue) Retry(ctx context.Context, job *Job) error {
	if job.Attempts >= job.MaxAttempts {
		// Move to dead letter queue
		return q.moveToDLQ(ctx, job)
	}

	jobBytes, err := json.Marshal(job)
	if err != nil {
		return err
	}

	// Re-enqueue with delay (exponential backoff)
	delay := time.Duration(job.Attempts*job.Attempts) * time.Second * 30
	time.AfterFunc(delay, func() {
		q.rdb.LPush(context.Background(), job.Queue, jobBytes)
	})

	return nil
}

// moveToDLQ moves a failed job to the dead letter queue.
func (q *Queue) moveToDLQ(ctx context.Context, job *Job) error {
	dlqKey := job.Queue + ":dlq"
	jobBytes, _ := json.Marshal(job)
	return q.rdb.LPush(ctx, dlqKey, jobBytes).Err()
}

// QueueStats returns the number of pending jobs per queue.
func (q *Queue) QueueStats(ctx context.Context) (map[string]int64, error) {
	queues := []string{QueueCritical, QueueHigh, QueueDefault, QueueLow}
	stats := make(map[string]int64)

	for _, queueName := range queues {
		length, err := q.rdb.LLen(ctx, queueName).Result()
		if err != nil {
			return nil, err
		}
		stats[queueName] = length
	}

	return stats, nil
}
