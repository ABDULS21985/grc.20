// Package service provides DSR (Data Subject Request) business logic for GDPR compliance.
// All PII fields are encrypted at rest using AES-256-GCM via the internal crypto package.
// Every state-changing action is logged to the dsr_audit_trail table.
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"github.com/complianceforge/platform/internal/pkg/crypto"
)

// ============================================================
// DSR MODELS (service-layer structs)
// ============================================================

// DSRRequest represents a GDPR Data Subject Request.
type DSRRequest struct {
	ID                         uuid.UUID  `json:"id"`
	OrganizationID             uuid.UUID  `json:"organization_id"`
	RequestRef                 string     `json:"request_ref"`
	RequestType                string     `json:"request_type"`
	Status                     string     `json:"status"`
	Priority                   string     `json:"priority"`
	DataSubjectName            string     `json:"data_subject_name,omitempty"`
	DataSubjectEmail           string     `json:"data_subject_email,omitempty"`
	DataSubjectPhone           string     `json:"data_subject_phone,omitempty"`
	DataSubjectAddress         string     `json:"data_subject_address,omitempty"`
	DataSubjectIDVerified      bool       `json:"data_subject_id_verified"`
	IdentityVerificationMethod string     `json:"identity_verification_method,omitempty"`
	IdentityVerifiedAt         *time.Time `json:"identity_verified_at,omitempty"`
	IdentityVerifiedBy         *uuid.UUID `json:"identity_verified_by,omitempty"`
	RequestDescription         string     `json:"request_description,omitempty"`
	RequestSource              string     `json:"request_source"`
	ReceivedDate               time.Time  `json:"received_date"`
	AcknowledgedAt             *time.Time `json:"acknowledged_at,omitempty"`
	ResponseDeadline           time.Time  `json:"response_deadline"`
	ExtendedDeadline           *time.Time `json:"extended_deadline,omitempty"`
	ExtensionReason            string     `json:"extension_reason,omitempty"`
	ExtensionNotifiedAt        *time.Time `json:"extension_notified_at,omitempty"`
	AssignedTo                 *uuid.UUID `json:"assigned_to,omitempty"`
	DataSystemsAffected        []string   `json:"data_systems_affected,omitempty"`
	DataCategoriesAffected     []string   `json:"data_categories_affected,omitempty"`
	ThirdPartiesNotified       []string   `json:"third_parties_notified,omitempty"`
	ProcessingNotes            string     `json:"processing_notes,omitempty"`
	CompletedAt                *time.Time `json:"completed_at,omitempty"`
	CompletedBy                *uuid.UUID `json:"completed_by,omitempty"`
	ResponseMethod             string     `json:"response_method,omitempty"`
	ResponseDocumentPath       string     `json:"response_document_path,omitempty"`
	RejectionReason            string     `json:"rejection_reason,omitempty"`
	RejectionLegalBasis        string     `json:"rejection_legal_basis,omitempty"`
	SLAStatus                  string     `json:"sla_status"`
	DaysRemaining              int        `json:"days_remaining"`
	WasExtended                bool       `json:"was_extended"`
	WasCompletedOnTime         *bool      `json:"was_completed_on_time,omitempty"`
	CreatedAt                  time.Time  `json:"created_at"`
	UpdatedAt                  time.Time  `json:"updated_at"`
	// Populated via joins
	Tasks      []DSRTask       `json:"tasks,omitempty"`
	AuditTrail []DSRAuditEntry `json:"audit_trail,omitempty"`
}

// DSRTask represents a task within a DSR workflow.
type DSRTask struct {
	ID             uuid.UUID  `json:"id"`
	OrganizationID uuid.UUID  `json:"organization_id"`
	DSRRequestID   uuid.UUID  `json:"dsr_request_id"`
	TaskType       string     `json:"task_type"`
	Description    string     `json:"description"`
	SystemName     string     `json:"system_name,omitempty"`
	AssignedTo     *uuid.UUID `json:"assigned_to,omitempty"`
	Status         string     `json:"status"`
	DueDate        *time.Time `json:"due_date,omitempty"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	CompletedBy    *uuid.UUID `json:"completed_by,omitempty"`
	Notes          string     `json:"notes,omitempty"`
	EvidencePath   string     `json:"evidence_path,omitempty"`
	SortOrder      int        `json:"sort_order"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// DSRAuditEntry represents an entry in the DSR audit trail.
type DSRAuditEntry struct {
	ID             uuid.UUID  `json:"id"`
	OrganizationID uuid.UUID  `json:"organization_id"`
	DSRRequestID   uuid.UUID  `json:"dsr_request_id"`
	Action         string     `json:"action"`
	PerformedBy    *uuid.UUID `json:"performed_by,omitempty"`
	Description    string     `json:"description"`
	CreatedAt      time.Time  `json:"created_at"`
}

// DSRResponseTemplate represents a response template.
type DSRResponseTemplate struct {
	ID             uuid.UUID  `json:"id"`
	OrganizationID *uuid.UUID `json:"organization_id,omitempty"`
	RequestType    string     `json:"request_type"`
	Name           string     `json:"name"`
	Subject        string     `json:"subject"`
	BodyHTML       string     `json:"body_html"`
	BodyText       string     `json:"body_text"`
	IsSystem       bool       `json:"is_system"`
	Language       string     `json:"language"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// CreateDSRRequest is the payload for creating a new DSR.
type CreateDSRRequest struct {
	RequestType        string     `json:"request_type" validate:"required,oneof=access erasure rectification portability restriction objection automated_decision"`
	Priority           string     `json:"priority" validate:"omitempty,oneof=standard urgent complex"`
	DataSubjectName    string     `json:"data_subject_name" validate:"required"`
	DataSubjectEmail   string     `json:"data_subject_email" validate:"required,email"`
	DataSubjectPhone   string     `json:"data_subject_phone,omitempty"`
	DataSubjectAddress string     `json:"data_subject_address,omitempty"`
	RequestDescription string     `json:"request_description,omitempty"`
	RequestSource      string     `json:"request_source" validate:"required,oneof=email form phone letter in_person portal"`
	ReceivedDate       *time.Time `json:"received_date,omitempty"`
	AssignedTo         *uuid.UUID `json:"assigned_to,omitempty"`
}

// UpdateDSRRequest is the payload for updating a DSR.
type UpdateDSRRequest struct {
	Priority               string     `json:"priority,omitempty"`
	RequestDescription     string     `json:"request_description,omitempty"`
	DataSystemsAffected    []string   `json:"data_systems_affected,omitempty"`
	DataCategoriesAffected []string   `json:"data_categories_affected,omitempty"`
	ThirdPartiesNotified   []string   `json:"third_parties_notified,omitempty"`
	ProcessingNotes        string     `json:"processing_notes,omitempty"`
	AssignedTo             *uuid.UUID `json:"assigned_to,omitempty"`
}

// DSRDashboard holds summary metrics for the DSR management dashboard.
type DSRDashboard struct {
	TotalRequests         int64   `json:"total_requests"`
	ReceivedCount         int64   `json:"received_count"`
	VerificationCount     int64   `json:"verification_count"`
	InProgressCount       int64   `json:"in_progress_count"`
	ExtendedCount         int64   `json:"extended_count"`
	CompletedCount        int64   `json:"completed_count"`
	RejectedCount         int64   `json:"rejected_count"`
	WithdrawnCount        int64   `json:"withdrawn_count"`
	OnTrackCount          int64   `json:"on_track_count"`
	AtRiskCount           int64   `json:"at_risk_count"`
	OverdueCount          int64   `json:"overdue_count"`
	AccessCount           int64   `json:"access_count"`
	ErasureCount          int64   `json:"erasure_count"`
	RectificationCount    int64   `json:"rectification_count"`
	PortabilityCount      int64   `json:"portability_count"`
	RestrictionCount      int64   `json:"restriction_count"`
	ObjectionCount        int64   `json:"objection_count"`
	AutomatedDecisionCount int64  `json:"automated_decision_count"`
	AvgCompletionDays     float64 `json:"avg_completion_days"`
	CompletedOnTimeCount  int64   `json:"completed_on_time_count"`
	CompletedLateCount    int64   `json:"completed_late_count"`
}

// ============================================================
// TASK CHECKLIST DEFINITIONS
// ============================================================

type taskTemplate struct {
	TaskType    string
	Description string
	SortOrder   int
}

var accessTaskChecklist = []taskTemplate{
	{"verify_identity", "Verify the identity of the data subject", 1},
	{"locate_data", "Locate all personal data held across systems", 2},
	{"extract_data", "Extract personal data from identified systems", 3},
	{"review_data", "Review extracted data for third-party information and exemptions", 4},
	{"compile_response", "Compile response package with all personal data", 5},
	{"send_response", "Send response to the data subject", 6},
}

var erasureTaskChecklist = []taskTemplate{
	{"verify_identity", "Verify the identity of the data subject", 1},
	{"locate_data", "Locate all personal data held across systems", 2},
	{"review_exemptions", "Review legal exemptions under GDPR Article 17(3)", 3},
	{"execute_erasure", "Execute erasure of personal data from all systems", 4},
	{"confirm_erasure", "Confirm erasure completion and obtain evidence", 5},
	{"notify_third_parties", "Notify third-party processors of erasure requirement", 6},
	{"send_confirmation", "Send erasure confirmation to the data subject", 7},
}

var rectificationTaskChecklist = []taskTemplate{
	{"verify_identity", "Verify the identity of the data subject", 1},
	{"locate_data", "Locate all instances of the data to be corrected", 2},
	{"verify_correction", "Verify the accuracy of the requested corrections", 3},
	{"execute_correction", "Apply corrections across all relevant systems", 4},
	{"notify_third_parties", "Notify third parties of the rectification per Article 19", 5},
	{"send_confirmation", "Send rectification confirmation to the data subject", 6},
}

var portabilityTaskChecklist = []taskTemplate{
	{"verify_identity", "Verify the identity of the data subject", 1},
	{"locate_data", "Locate all personal data provided by the data subject", 2},
	{"extract_in_machine_readable", "Extract data in a structured, machine-readable format (JSON/CSV)", 3},
	{"review_data", "Review extracted data for completeness and accuracy", 4},
	{"send_response", "Send portable data package to the data subject", 5},
}

var restrictionTaskChecklist = []taskTemplate{
	{"verify_identity", "Verify the identity of the data subject", 1},
	{"locate_data", "Locate all personal data subject to restriction", 2},
	{"review_exemptions", "Review grounds for restriction under Article 18", 3},
	{"execute_correction", "Apply processing restriction markers to relevant data", 4},
	{"notify_third_parties", "Notify third parties of the restriction per Article 19", 5},
	{"send_confirmation", "Send restriction confirmation to the data subject", 6},
}

var objectionTaskChecklist = []taskTemplate{
	{"verify_identity", "Verify the identity of the data subject", 1},
	{"review_data", "Assess objection grounds and legitimate interests", 2},
	{"compile_response", "Compile response with decision and reasoning", 3},
	{"send_response", "Send objection decision to the data subject", 4},
}

var automatedDecisionTaskChecklist = []taskTemplate{
	{"verify_identity", "Verify the identity of the data subject", 1},
	{"locate_data", "Identify automated decision-making processes affecting the subject", 2},
	{"review_data", "Review the logic and impact of automated decisions", 3},
	{"compile_response", "Compile meaningful information about the logic involved", 4},
	{"send_response", "Provide response with information about automated processing", 5},
}

func getTaskChecklist(requestType string) []taskTemplate {
	switch requestType {
	case "access":
		return accessTaskChecklist
	case "erasure":
		return erasureTaskChecklist
	case "rectification":
		return rectificationTaskChecklist
	case "portability":
		return portabilityTaskChecklist
	case "restriction":
		return restrictionTaskChecklist
	case "objection":
		return objectionTaskChecklist
	case "automated_decision":
		return automatedDecisionTaskChecklist
	default:
		return accessTaskChecklist
	}
}

// ============================================================
// DSR SERVICE
// ============================================================

// DSRService provides business logic for GDPR Data Subject Requests.
type DSRService struct {
	pool      *pgxpool.Pool
	encryptor *crypto.Encryptor
}

// NewDSRService creates a new DSRService.
func NewDSRService(pool *pgxpool.Pool, encryptor *crypto.Encryptor) *DSRService {
	return &DSRService{
		pool:      pool,
		encryptor: encryptor,
	}
}

// ============================================================
// CREATE REQUEST
// ============================================================

// CreateRequest creates a new DSR with auto-generated reference, 30-day deadline,
// and a default task checklist based on the request type.
func (s *DSRService) CreateRequest(ctx context.Context, orgID uuid.UUID, userID uuid.UUID, req CreateDSRRequest) (*DSRRequest, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Set tenant context for RLS
	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_tenant', $1, true)", orgID.String()); err != nil {
		return nil, fmt.Errorf("failed to set tenant: %w", err)
	}

	// Generate reference number
	var requestRef string
	if err := tx.QueryRow(ctx, "SELECT generate_dsr_ref($1)", orgID).Scan(&requestRef); err != nil {
		return nil, fmt.Errorf("failed to generate DSR reference: %w", err)
	}

	// Encrypt PII fields
	nameEnc, err := s.encryptor.EncryptString(req.DataSubjectName)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt name: %w", err)
	}
	emailEnc, err := s.encryptor.EncryptString(req.DataSubjectEmail)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt email: %w", err)
	}
	phoneEnc := ""
	if req.DataSubjectPhone != "" {
		phoneEnc, err = s.encryptor.EncryptString(req.DataSubjectPhone)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt phone: %w", err)
		}
	}
	addressEnc := ""
	if req.DataSubjectAddress != "" {
		addressEnc, err = s.encryptor.EncryptString(req.DataSubjectAddress)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt address: %w", err)
		}
	}

	// Calculate received date (default to today)
	receivedDate := time.Now()
	if req.ReceivedDate != nil {
		receivedDate = *req.ReceivedDate
	}

	// Calculate 30-day response deadline (GDPR Article 12(3))
	responseDeadline := receivedDate.AddDate(0, 0, 30)

	priority := req.Priority
	if priority == "" {
		priority = "standard"
	}

	now := time.Now()
	requestID := uuid.New()

	// Insert DSR request
	_, err = tx.Exec(ctx, `
		INSERT INTO dsr_requests (
			id, organization_id, request_ref, request_type, status, priority,
			data_subject_name_encrypted, data_subject_email_encrypted,
			data_subject_phone_encrypted, data_subject_address_encrypted,
			request_description, request_source, received_date, acknowledged_at,
			response_deadline, assigned_to, metadata, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)`,
		requestID, orgID, requestRef, req.RequestType, "received", priority,
		nameEnc, emailEnc, phoneEnc, addressEnc,
		req.RequestDescription, req.RequestSource, receivedDate.Format("2006-01-02"), now,
		responseDeadline.Format("2006-01-02"), req.AssignedTo, "{}", now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert DSR request: %w", err)
	}

	// Create default task checklist based on request type
	tasks := getTaskChecklist(req.RequestType)
	for _, t := range tasks {
		taskID := uuid.New()
		_, err = tx.Exec(ctx, `
			INSERT INTO dsr_tasks (
				id, organization_id, dsr_request_id, task_type, description,
				status, due_date, sort_order, created_at, updated_at
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
			taskID, orgID, requestID, t.TaskType, t.Description,
			"pending", responseDeadline.Format("2006-01-02"), t.SortOrder, now, now,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to insert task %s: %w", t.TaskType, err)
		}
	}

	// Log to audit trail
	if err := s.logAudit(ctx, tx, orgID, requestID, &userID, "request_created",
		fmt.Sprintf("DSR %s created: %s request via %s", requestRef, req.RequestType, req.RequestSource)); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}

	// Calculate initial days remaining
	daysRemaining := int(time.Until(responseDeadline).Hours() / 24)

	result := &DSRRequest{
		ID:                 requestID,
		OrganizationID:     orgID,
		RequestRef:         requestRef,
		RequestType:        req.RequestType,
		Status:             "received",
		Priority:           priority,
		DataSubjectName:    req.DataSubjectName,
		DataSubjectEmail:   req.DataSubjectEmail,
		DataSubjectPhone:   req.DataSubjectPhone,
		DataSubjectAddress: req.DataSubjectAddress,
		RequestDescription: req.RequestDescription,
		RequestSource:      req.RequestSource,
		ReceivedDate:       receivedDate,
		AcknowledgedAt:     &now,
		ResponseDeadline:   responseDeadline,
		AssignedTo:         req.AssignedTo,
		SLAStatus:          "on_track",
		DaysRemaining:      daysRemaining,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	log.Info().
		Str("request_ref", requestRef).
		Str("org_id", orgID.String()).
		Str("type", req.RequestType).
		Str("deadline", responseDeadline.Format("2006-01-02")).
		Msg("DSR request created")

	return result, nil
}

// ============================================================
// VERIFY IDENTITY
// ============================================================

// VerifyIdentity records that the data subject's identity has been verified.
func (s *DSRService) VerifyIdentity(ctx context.Context, orgID, requestID uuid.UUID, method string, verifiedBy uuid.UUID) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_tenant', $1, true)", orgID.String()); err != nil {
		return fmt.Errorf("failed to set tenant: %w", err)
	}

	now := time.Now()
	tag, err := tx.Exec(ctx, `
		UPDATE dsr_requests SET
			data_subject_id_verified = true,
			identity_verification_method = $1,
			identity_verified_at = $2,
			identity_verified_by = $3,
			status = CASE WHEN status = 'received' THEN 'in_progress'::dsr_status ELSE status END
		WHERE id = $4 AND organization_id = $5 AND deleted_at IS NULL`,
		method, now, verifiedBy, requestID, orgID,
	)
	if err != nil {
		return fmt.Errorf("failed to update identity verification: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("DSR request not found")
	}

	if err := s.logAudit(ctx, tx, orgID, requestID, &verifiedBy, "identity_verified",
		fmt.Sprintf("Identity verified using method: %s", method)); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// ============================================================
// ASSIGN REQUEST
// ============================================================

// AssignRequest assigns a DSR to a user.
func (s *DSRService) AssignRequest(ctx context.Context, orgID, requestID, assigneeID, performedBy uuid.UUID) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_tenant', $1, true)", orgID.String()); err != nil {
		return fmt.Errorf("failed to set tenant: %w", err)
	}

	tag, err := tx.Exec(ctx, `
		UPDATE dsr_requests SET assigned_to = $1
		WHERE id = $2 AND organization_id = $3 AND deleted_at IS NULL`,
		assigneeID, requestID, orgID,
	)
	if err != nil {
		return fmt.Errorf("failed to assign request: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("DSR request not found")
	}

	if err := s.logAudit(ctx, tx, orgID, requestID, &performedBy, "request_assigned",
		fmt.Sprintf("Request assigned to user %s", assigneeID.String())); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// ============================================================
// EXTEND DEADLINE
// ============================================================

// ExtendDeadline extends the response deadline by 60 days (GDPR Article 12(3)).
func (s *DSRService) ExtendDeadline(ctx context.Context, orgID, requestID uuid.UUID, reason string, performedBy uuid.UUID) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_tenant', $1, true)", orgID.String()); err != nil {
		return fmt.Errorf("failed to set tenant: %w", err)
	}

	now := time.Now()

	// Get the current deadline
	var currentDeadline time.Time
	err = tx.QueryRow(ctx, `
		SELECT response_deadline FROM dsr_requests
		WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL`,
		requestID, orgID,
	).Scan(&currentDeadline)
	if err != nil {
		return fmt.Errorf("DSR request not found: %w", err)
	}

	// Extend by 60 days (total allowed extension under GDPR is 2 months)
	extendedDeadline := currentDeadline.AddDate(0, 0, 60)

	tag, err := tx.Exec(ctx, `
		UPDATE dsr_requests SET
			status = 'extended'::dsr_status,
			extended_deadline = $1,
			extension_reason = $2,
			extension_notified_at = $3,
			was_extended = true
		WHERE id = $4 AND organization_id = $5 AND deleted_at IS NULL`,
		extendedDeadline.Format("2006-01-02"), reason, now, requestID, orgID,
	)
	if err != nil {
		return fmt.Errorf("failed to extend deadline: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("DSR request not found")
	}

	if err := s.logAudit(ctx, tx, orgID, requestID, &performedBy, "deadline_extended",
		fmt.Sprintf("Deadline extended to %s. Reason: %s", extendedDeadline.Format("2006-01-02"), reason)); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// ============================================================
// COMPLETE TASK
// ============================================================

// CompleteTask marks a DSR task as completed.
func (s *DSRService) CompleteTask(ctx context.Context, orgID, taskID, completedBy uuid.UUID, notes, evidencePath string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_tenant', $1, true)", orgID.String()); err != nil {
		return fmt.Errorf("failed to set tenant: %w", err)
	}

	now := time.Now()

	// Update the task
	var dsrRequestID uuid.UUID
	var taskType string
	err = tx.QueryRow(ctx, `
		UPDATE dsr_tasks SET
			status = 'completed'::dsr_task_status,
			completed_at = $1,
			completed_by = $2,
			notes = CASE WHEN $3 = '' THEN notes ELSE $3 END,
			evidence_path = CASE WHEN $4 = '' THEN evidence_path ELSE $4 END
		WHERE id = $5 AND organization_id = $6
		RETURNING dsr_request_id, task_type`,
		now, completedBy, notes, evidencePath, taskID, orgID,
	).Scan(&dsrRequestID, &taskType)
	if err != nil {
		return fmt.Errorf("failed to complete task: %w", err)
	}

	if err := s.logAudit(ctx, tx, orgID, dsrRequestID, &completedBy, "task_completed",
		fmt.Sprintf("Task '%s' completed", taskType)); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// ============================================================
// UPDATE TASK
// ============================================================

// UpdateTask updates a DSR task's status and metadata.
func (s *DSRService) UpdateTask(ctx context.Context, orgID, taskID uuid.UUID, status, notes, evidencePath string, assignedTo *uuid.UUID, performedBy uuid.UUID) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_tenant', $1, true)", orgID.String()); err != nil {
		return fmt.Errorf("failed to set tenant: %w", err)
	}

	now := time.Now()
	var dsrRequestID uuid.UUID
	var taskType string

	var completedAt *time.Time
	var completedByPtr *uuid.UUID
	if status == "completed" {
		completedAt = &now
		completedByPtr = &performedBy
	}

	err = tx.QueryRow(ctx, `
		UPDATE dsr_tasks SET
			status = $1::dsr_task_status,
			notes = CASE WHEN $2 = '' THEN notes ELSE $2 END,
			evidence_path = CASE WHEN $3 = '' THEN evidence_path ELSE $3 END,
			assigned_to = COALESCE($4, assigned_to),
			completed_at = COALESCE($5, completed_at),
			completed_by = COALESCE($6, completed_by)
		WHERE id = $7 AND organization_id = $8
		RETURNING dsr_request_id, task_type`,
		status, notes, evidencePath, assignedTo, completedAt, completedByPtr, taskID, orgID,
	).Scan(&dsrRequestID, &taskType)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	if err := s.logAudit(ctx, tx, orgID, dsrRequestID, &performedBy, "task_updated",
		fmt.Sprintf("Task '%s' updated to status: %s", taskType, status)); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// ============================================================
// COMPLETE REQUEST
// ============================================================

// CompleteRequest marks a DSR as completed.
func (s *DSRService) CompleteRequest(ctx context.Context, orgID, requestID uuid.UUID, responseMethod, documentPath string, completedBy uuid.UUID) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_tenant', $1, true)", orgID.String()); err != nil {
		return fmt.Errorf("failed to set tenant: %w", err)
	}

	now := time.Now()

	// Get the deadlines to determine if completed on time
	var responseDeadline time.Time
	var extendedDeadline *time.Time
	err = tx.QueryRow(ctx, `
		SELECT response_deadline, extended_deadline FROM dsr_requests
		WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL`,
		requestID, orgID,
	).Scan(&responseDeadline, &extendedDeadline)
	if err != nil {
		return fmt.Errorf("DSR request not found: %w", err)
	}

	effectiveDeadline := responseDeadline
	if extendedDeadline != nil {
		effectiveDeadline = *extendedDeadline
	}
	wasOnTime := now.Before(effectiveDeadline.AddDate(0, 0, 1)) // include the deadline day itself

	tag, err := tx.Exec(ctx, `
		UPDATE dsr_requests SET
			status = 'completed'::dsr_status,
			completed_at = $1,
			completed_by = $2,
			response_method = $3::dsr_response_method,
			response_document_path = $4,
			was_completed_on_time = $5,
			sla_status = 'on_track'::dsr_sla_status,
			days_remaining = 0
		WHERE id = $6 AND organization_id = $7 AND deleted_at IS NULL`,
		now, completedBy, responseMethod, documentPath, wasOnTime, requestID, orgID,
	)
	if err != nil {
		return fmt.Errorf("failed to complete request: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("DSR request not found")
	}

	if err := s.logAudit(ctx, tx, orgID, requestID, &completedBy, "request_completed",
		fmt.Sprintf("Request completed via %s. On time: %v", responseMethod, wasOnTime)); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// ============================================================
// REJECT REQUEST
// ============================================================

// RejectRequest rejects a DSR with a reason and legal basis.
func (s *DSRService) RejectRequest(ctx context.Context, orgID, requestID uuid.UUID, reason, legalBasis string, rejectedBy uuid.UUID) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_tenant', $1, true)", orgID.String()); err != nil {
		return fmt.Errorf("failed to set tenant: %w", err)
	}

	now := time.Now()
	tag, err := tx.Exec(ctx, `
		UPDATE dsr_requests SET
			status = 'rejected'::dsr_status,
			rejection_reason = $1,
			rejection_legal_basis = $2,
			completed_at = $3,
			completed_by = $4
		WHERE id = $5 AND organization_id = $6 AND deleted_at IS NULL`,
		reason, legalBasis, now, rejectedBy, requestID, orgID,
	)
	if err != nil {
		return fmt.Errorf("failed to reject request: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("DSR request not found")
	}

	if err := s.logAudit(ctx, tx, orgID, requestID, &rejectedBy, "request_rejected",
		fmt.Sprintf("Request rejected. Reason: %s. Legal basis: %s", reason, legalBasis)); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// ============================================================
// UPDATE REQUEST
// ============================================================

// UpdateRequest updates mutable fields of a DSR.
func (s *DSRService) UpdateRequest(ctx context.Context, orgID, requestID uuid.UUID, req UpdateDSRRequest, performedBy uuid.UUID) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_tenant', $1, true)", orgID.String()); err != nil {
		return fmt.Errorf("failed to set tenant: %w", err)
	}

	tag, err := tx.Exec(ctx, `
		UPDATE dsr_requests SET
			priority = CASE WHEN $1 = '' THEN priority ELSE $1::dsr_priority END,
			request_description = CASE WHEN $2 = '' THEN request_description ELSE $2 END,
			data_systems_affected = CASE WHEN $3::text[] IS NULL THEN data_systems_affected ELSE $3 END,
			data_categories_affected = CASE WHEN $4::text[] IS NULL THEN data_categories_affected ELSE $4 END,
			third_parties_notified = CASE WHEN $5::text[] IS NULL THEN third_parties_notified ELSE $5 END,
			processing_notes = CASE WHEN $6 = '' THEN processing_notes ELSE $6 END,
			assigned_to = COALESCE($7, assigned_to)
		WHERE id = $8 AND organization_id = $9 AND deleted_at IS NULL`,
		req.Priority, req.RequestDescription,
		req.DataSystemsAffected, req.DataCategoriesAffected,
		req.ThirdPartiesNotified, req.ProcessingNotes,
		req.AssignedTo, requestID, orgID,
	)
	if err != nil {
		return fmt.Errorf("failed to update request: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("DSR request not found")
	}

	if err := s.logAudit(ctx, tx, orgID, requestID, &performedBy, "request_updated",
		"Request details updated"); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// ============================================================
// GET REQUEST
// ============================================================

// GetRequest returns a DSR by ID with tasks and audit trail.
func (s *DSRService) GetRequest(ctx context.Context, orgID, requestID uuid.UUID) (*DSRRequest, error) {
	dsr := &DSRRequest{}
	var nameEnc, emailEnc, phoneEnc, addressEnc string
	var responseMethod, rejectionReason, rejectionLegalBasis, extensionReason, processingNotes, requestDesc *string
	var wasCompletedOnTime *bool

	err := s.pool.QueryRow(ctx, `
		SELECT
			id, organization_id, request_ref, request_type::text, status::text, priority::text,
			COALESCE(data_subject_name_encrypted, ''), COALESCE(data_subject_email_encrypted, ''),
			COALESCE(data_subject_phone_encrypted, ''), COALESCE(data_subject_address_encrypted, ''),
			data_subject_id_verified, COALESCE(identity_verification_method, ''),
			identity_verified_at, identity_verified_by,
			request_description, request_source::text, received_date, acknowledged_at,
			response_deadline, extended_deadline, extension_reason, extension_notified_at,
			assigned_to, data_systems_affected, data_categories_affected, third_parties_notified,
			processing_notes, completed_at, completed_by,
			response_method::text, response_document_path,
			rejection_reason, rejection_legal_basis,
			COALESCE(sla_status::text, 'on_track'), COALESCE(days_remaining, 0),
			was_extended, was_completed_on_time,
			created_at, updated_at
		FROM dsr_requests
		WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL`,
		requestID, orgID,
	).Scan(
		&dsr.ID, &dsr.OrganizationID, &dsr.RequestRef, &dsr.RequestType, &dsr.Status, &dsr.Priority,
		&nameEnc, &emailEnc, &phoneEnc, &addressEnc,
		&dsr.DataSubjectIDVerified, &dsr.IdentityVerificationMethod,
		&dsr.IdentityVerifiedAt, &dsr.IdentityVerifiedBy,
		&requestDesc, &dsr.RequestSource, &dsr.ReceivedDate, &dsr.AcknowledgedAt,
		&dsr.ResponseDeadline, &dsr.ExtendedDeadline, &extensionReason, &dsr.ExtensionNotifiedAt,
		&dsr.AssignedTo, &dsr.DataSystemsAffected, &dsr.DataCategoriesAffected, &dsr.ThirdPartiesNotified,
		&processingNotes, &dsr.CompletedAt, &dsr.CompletedBy,
		&responseMethod, &dsr.ResponseDocumentPath,
		&rejectionReason, &rejectionLegalBasis,
		&dsr.SLAStatus, &dsr.DaysRemaining,
		&dsr.WasExtended, &wasCompletedOnTime,
		&dsr.CreatedAt, &dsr.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("DSR request not found: %w", err)
	}

	// Assign optional string fields
	if requestDesc != nil {
		dsr.RequestDescription = *requestDesc
	}
	if extensionReason != nil {
		dsr.ExtensionReason = *extensionReason
	}
	if processingNotes != nil {
		dsr.ProcessingNotes = *processingNotes
	}
	if responseMethod != nil {
		dsr.ResponseMethod = *responseMethod
	}
	if rejectionReason != nil {
		dsr.RejectionReason = *rejectionReason
	}
	if rejectionLegalBasis != nil {
		dsr.RejectionLegalBasis = *rejectionLegalBasis
	}
	if wasCompletedOnTime != nil {
		dsr.WasCompletedOnTime = wasCompletedOnTime
	}

	// Decrypt PII fields
	if nameEnc != "" {
		dsr.DataSubjectName, _ = s.encryptor.DecryptString(nameEnc)
	}
	if emailEnc != "" {
		dsr.DataSubjectEmail, _ = s.encryptor.DecryptString(emailEnc)
	}
	if phoneEnc != "" {
		dsr.DataSubjectPhone, _ = s.encryptor.DecryptString(phoneEnc)
	}
	if addressEnc != "" {
		dsr.DataSubjectAddress, _ = s.encryptor.DecryptString(addressEnc)
	}

	// Fetch tasks
	tasks, err := s.getTasksByRequestID(ctx, orgID, requestID)
	if err == nil {
		dsr.Tasks = tasks
	}

	// Fetch audit trail
	trail, err := s.getAuditTrail(ctx, orgID, requestID)
	if err == nil {
		dsr.AuditTrail = trail
	}

	return dsr, nil
}

// ============================================================
// LIST REQUESTS
// ============================================================

// ListRequests returns paginated DSR requests for an organization.
func (s *DSRService) ListRequests(ctx context.Context, orgID uuid.UUID, page, pageSize int, status, requestType, slaStatus string) ([]DSRRequest, int64, error) {
	// Count total
	countQuery := `SELECT COUNT(*) FROM dsr_requests WHERE organization_id = $1 AND deleted_at IS NULL`
	args := []interface{}{orgID}
	argIdx := 2

	if status != "" {
		countQuery += fmt.Sprintf(" AND status = $%d::dsr_status", argIdx)
		args = append(args, status)
		argIdx++
	}
	if requestType != "" {
		countQuery += fmt.Sprintf(" AND request_type = $%d::dsr_request_type", argIdx)
		args = append(args, requestType)
		argIdx++
	}
	if slaStatus != "" {
		countQuery += fmt.Sprintf(" AND sla_status = $%d::dsr_sla_status", argIdx)
		args = append(args, slaStatus)
		argIdx++
	}

	var total int64
	if err := s.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count DSR requests: %w", err)
	}

	// Fetch rows
	query := `
		SELECT
			id, organization_id, request_ref, request_type::text, status::text, priority::text,
			data_subject_id_verified,
			request_source::text, received_date, response_deadline,
			extended_deadline, assigned_to,
			COALESCE(sla_status::text, 'on_track'), COALESCE(days_remaining, 0),
			was_extended, was_completed_on_time, completed_at,
			created_at, updated_at
		FROM dsr_requests
		WHERE organization_id = $1 AND deleted_at IS NULL`

	listArgs := []interface{}{orgID}
	listArgIdx := 2

	if status != "" {
		query += fmt.Sprintf(" AND status = $%d::dsr_status", listArgIdx)
		listArgs = append(listArgs, status)
		listArgIdx++
	}
	if requestType != "" {
		query += fmt.Sprintf(" AND request_type = $%d::dsr_request_type", listArgIdx)
		listArgs = append(listArgs, requestType)
		listArgIdx++
	}
	if slaStatus != "" {
		query += fmt.Sprintf(" AND sla_status = $%d::dsr_sla_status", listArgIdx)
		listArgs = append(listArgs, slaStatus)
		listArgIdx++
	}

	query += " ORDER BY created_at DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", listArgIdx, listArgIdx+1)
	listArgs = append(listArgs, pageSize, (page-1)*pageSize)

	rows, err := s.pool.Query(ctx, query, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list DSR requests: %w", err)
	}
	defer rows.Close()

	var dsrs []DSRRequest
	for rows.Next() {
		var d DSRRequest
		var wasCompletedOnTime *bool
		if err := rows.Scan(
			&d.ID, &d.OrganizationID, &d.RequestRef, &d.RequestType, &d.Status, &d.Priority,
			&d.DataSubjectIDVerified,
			&d.RequestSource, &d.ReceivedDate, &d.ResponseDeadline,
			&d.ExtendedDeadline, &d.AssignedTo,
			&d.SLAStatus, &d.DaysRemaining,
			&d.WasExtended, &wasCompletedOnTime, &d.CompletedAt,
			&d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan DSR request: %w", err)
		}
		if wasCompletedOnTime != nil {
			d.WasCompletedOnTime = wasCompletedOnTime
		}
		dsrs = append(dsrs, d)
	}

	return dsrs, total, nil
}

// ============================================================
// DASHBOARD
// ============================================================

// GetDSRDashboard returns summary metrics for the DSR management dashboard.
func (s *DSRService) GetDSRDashboard(ctx context.Context, orgID uuid.UUID) (*DSRDashboard, error) {
	dash := &DSRDashboard{}

	err := s.pool.QueryRow(ctx, `
		SELECT
			COALESCE(total_requests, 0), COALESCE(received_count, 0),
			COALESCE(verification_count, 0), COALESCE(in_progress_count, 0),
			COALESCE(extended_count, 0), COALESCE(completed_count, 0),
			COALESCE(rejected_count, 0), COALESCE(withdrawn_count, 0),
			COALESCE(on_track_count, 0), COALESCE(at_risk_count, 0), COALESCE(overdue_count, 0),
			COALESCE(access_count, 0), COALESCE(erasure_count, 0),
			COALESCE(rectification_count, 0), COALESCE(portability_count, 0),
			COALESCE(restriction_count, 0), COALESCE(objection_count, 0),
			COALESCE(automated_decision_count, 0),
			COALESCE(avg_completion_days, 0),
			COALESCE(completed_on_time_count, 0), COALESCE(completed_late_count, 0)
		FROM v_dsr_dashboard WHERE organization_id = $1`, orgID,
	).Scan(
		&dash.TotalRequests, &dash.ReceivedCount,
		&dash.VerificationCount, &dash.InProgressCount,
		&dash.ExtendedCount, &dash.CompletedCount,
		&dash.RejectedCount, &dash.WithdrawnCount,
		&dash.OnTrackCount, &dash.AtRiskCount, &dash.OverdueCount,
		&dash.AccessCount, &dash.ErasureCount,
		&dash.RectificationCount, &dash.PortabilityCount,
		&dash.RestrictionCount, &dash.ObjectionCount,
		&dash.AutomatedDecisionCount,
		&dash.AvgCompletionDays,
		&dash.CompletedOnTimeCount, &dash.CompletedLateCount,
	)
	if err != nil {
		// No data yet; return zero-initialized dashboard
		if err == pgx.ErrNoRows {
			return dash, nil
		}
		return nil, fmt.Errorf("failed to get DSR dashboard: %w", err)
	}

	return dash, nil
}

// ============================================================
// SLA COMPLIANCE CHECK
// ============================================================

// CheckSLACompliance returns DSR requests that are at risk or overdue.
func (s *DSRService) CheckSLACompliance(ctx context.Context, orgID uuid.UUID) ([]DSRRequest, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT
			id, organization_id, request_ref, request_type::text, status::text, priority::text,
			request_source::text, received_date, response_deadline,
			extended_deadline, assigned_to,
			COALESCE(sla_status::text, 'on_track'), COALESCE(days_remaining, 0),
			was_extended, created_at, updated_at
		FROM dsr_requests
		WHERE organization_id = $1
			AND deleted_at IS NULL
			AND status NOT IN ('completed', 'rejected', 'withdrawn')
			AND (sla_status = 'at_risk' OR sla_status = 'overdue')
		ORDER BY days_remaining ASC`, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check SLA compliance: %w", err)
	}
	defer rows.Close()

	var dsrs []DSRRequest
	for rows.Next() {
		var d DSRRequest
		if err := rows.Scan(
			&d.ID, &d.OrganizationID, &d.RequestRef, &d.RequestType, &d.Status, &d.Priority,
			&d.RequestSource, &d.ReceivedDate, &d.ResponseDeadline,
			&d.ExtendedDeadline, &d.AssignedTo,
			&d.SLAStatus, &d.DaysRemaining,
			&d.WasExtended, &d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan DSR request: %w", err)
		}
		dsrs = append(dsrs, d)
	}

	return dsrs, nil
}

// ============================================================
// OVERDUE REQUESTS
// ============================================================

// GetOverdueRequests returns DSR requests past their deadline.
func (s *DSRService) GetOverdueRequests(ctx context.Context, orgID uuid.UUID) ([]DSRRequest, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT
			id, organization_id, request_ref, request_type::text, status::text, priority::text,
			request_source::text, received_date, response_deadline,
			extended_deadline, assigned_to,
			COALESCE(sla_status::text, 'overdue'), COALESCE(days_remaining, 0),
			was_extended, created_at, updated_at
		FROM dsr_requests
		WHERE organization_id = $1
			AND deleted_at IS NULL
			AND status NOT IN ('completed', 'rejected', 'withdrawn')
			AND sla_status = 'overdue'
		ORDER BY response_deadline ASC`, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get overdue requests: %w", err)
	}
	defer rows.Close()

	var dsrs []DSRRequest
	for rows.Next() {
		var d DSRRequest
		if err := rows.Scan(
			&d.ID, &d.OrganizationID, &d.RequestRef, &d.RequestType, &d.Status, &d.Priority,
			&d.RequestSource, &d.ReceivedDate, &d.ResponseDeadline,
			&d.ExtendedDeadline, &d.AssignedTo,
			&d.SLAStatus, &d.DaysRemaining,
			&d.WasExtended, &d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan overdue request: %w", err)
		}
		dsrs = append(dsrs, d)
	}

	return dsrs, nil
}

// ============================================================
// RESPONSE TEMPLATES
// ============================================================

// GetResponseTemplates returns response templates, combining system and org-specific ones.
func (s *DSRService) GetResponseTemplates(ctx context.Context, orgID uuid.UUID) ([]DSRResponseTemplate, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, organization_id, request_type::text, name, subject, body_html, body_text,
			is_system, language, created_at, updated_at
		FROM dsr_response_templates
		WHERE (organization_id = $1 OR organization_id IS NULL)
		ORDER BY is_system DESC, request_type, name`, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get templates: %w", err)
	}
	defer rows.Close()

	var templates []DSRResponseTemplate
	for rows.Next() {
		var t DSRResponseTemplate
		if err := rows.Scan(
			&t.ID, &t.OrganizationID, &t.RequestType, &t.Name, &t.Subject,
			&t.BodyHTML, &t.BodyText, &t.IsSystem, &t.Language,
			&t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan template: %w", err)
		}
		templates = append(templates, t)
	}

	return templates, nil
}

// ============================================================
// SLA STATUS COMPUTATION (used by scheduler)
// ============================================================

// CalculateSLAStatus computes the SLA status for a given deadline.
// Returns the SLA status string and the number of days remaining.
func CalculateSLAStatus(now time.Time, responseDeadline time.Time, extendedDeadline *time.Time) (string, int) {
	effectiveDeadline := responseDeadline
	if extendedDeadline != nil && !extendedDeadline.IsZero() {
		effectiveDeadline = *extendedDeadline
	}

	daysRemaining := int(effectiveDeadline.Sub(now).Hours() / 24)

	if daysRemaining < 0 {
		return "overdue", daysRemaining
	}
	if daysRemaining <= 7 {
		return "at_risk", daysRemaining
	}
	return "on_track", daysRemaining
}

// UpdateSLAStatuses batch-updates SLA fields for all active DSR requests in an organization.
func (s *DSRService) UpdateSLAStatuses(ctx context.Context, orgID uuid.UUID) (int64, error) {
	now := time.Now()

	// Use a single UPDATE statement with computed CASE expressions
	tag, err := s.pool.Exec(ctx, `
		UPDATE dsr_requests SET
			days_remaining = CASE
				WHEN extended_deadline IS NOT NULL
				THEN EXTRACT(DAY FROM (extended_deadline - $1::date))::int
				ELSE EXTRACT(DAY FROM (response_deadline - $1::date))::int
			END,
			sla_status = CASE
				WHEN COALESCE(extended_deadline, response_deadline) < $1::date THEN 'overdue'::dsr_sla_status
				WHEN COALESCE(extended_deadline, response_deadline) - $1::date <= 7 THEN 'at_risk'::dsr_sla_status
				ELSE 'on_track'::dsr_sla_status
			END
		WHERE organization_id = $2
			AND deleted_at IS NULL
			AND status NOT IN ('completed', 'rejected', 'withdrawn')`,
		now.Format("2006-01-02"), orgID,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to update SLA statuses: %w", err)
	}

	return tag.RowsAffected(), nil
}

// ============================================================
// INTERNAL HELPERS
// ============================================================

func (s *DSRService) logAudit(ctx context.Context, tx pgx.Tx, orgID, requestID uuid.UUID, performedBy *uuid.UUID, action, description string) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO dsr_audit_trail (id, organization_id, dsr_request_id, action, performed_by, description, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, '{}', $7)`,
		uuid.New(), orgID, requestID, action, performedBy, description, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to log audit: %w", err)
	}
	return nil
}

func (s *DSRService) getTasksByRequestID(ctx context.Context, orgID, requestID uuid.UUID) ([]DSRTask, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, organization_id, dsr_request_id, task_type::text, COALESCE(description, ''),
			COALESCE(system_name, ''), assigned_to, status::text,
			due_date, completed_at, completed_by, COALESCE(notes, ''), COALESCE(evidence_path, ''),
			sort_order, created_at, updated_at
		FROM dsr_tasks
		WHERE dsr_request_id = $1 AND organization_id = $2
		ORDER BY sort_order ASC`, requestID, orgID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []DSRTask
	for rows.Next() {
		var t DSRTask
		if err := rows.Scan(
			&t.ID, &t.OrganizationID, &t.DSRRequestID, &t.TaskType, &t.Description,
			&t.SystemName, &t.AssignedTo, &t.Status,
			&t.DueDate, &t.CompletedAt, &t.CompletedBy, &t.Notes, &t.EvidencePath,
			&t.SortOrder, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}

	return tasks, nil
}

func (s *DSRService) getAuditTrail(ctx context.Context, orgID, requestID uuid.UUID) ([]DSRAuditEntry, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, organization_id, dsr_request_id, action, performed_by, COALESCE(description, ''), created_at
		FROM dsr_audit_trail
		WHERE dsr_request_id = $1 AND organization_id = $2
		ORDER BY created_at DESC`, requestID, orgID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []DSRAuditEntry
	for rows.Next() {
		var e DSRAuditEntry
		if err := rows.Scan(
			&e.ID, &e.OrganizationID, &e.DSRRequestID, &e.Action, &e.PerformedBy, &e.Description, &e.CreatedAt,
		); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}

	return entries, nil
}
