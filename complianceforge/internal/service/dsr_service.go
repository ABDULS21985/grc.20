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

type DSRRequest struct {
	ID                         uuid.UUID       `json:"id"`
	OrganizationID             uuid.UUID       `json:"organization_id"`
	RequestRef                 string          `json:"request_ref"`
	RequestType                string          `json:"request_type"`
	Status                     string          `json:"status"`
	Priority                   string          `json:"priority"`
	DataSubjectName            string          `json:"data_subject_name,omitempty"`
	DataSubjectEmail           string          `json:"data_subject_email,omitempty"`
	DataSubjectPhone           string          `json:"data_subject_phone,omitempty"`
	DataSubjectAddress         string          `json:"data_subject_address,omitempty"`
	DataSubjectIDVerified      bool            `json:"data_subject_id_verified"`
	IdentityVerificationMethod string          `json:"identity_verification_method,omitempty"`
	IdentityVerifiedAt         *time.Time      `json:"identity_verified_at,omitempty"`
	IdentityVerifiedBy         *uuid.UUID      `json:"identity_verified_by,omitempty"`
	RequestDescription         string          `json:"request_description,omitempty"`
	RequestSource              string          `json:"request_source"`
	ReceivedDate               time.Time       `json:"received_date"`
	AcknowledgedAt             *time.Time      `json:"acknowledged_at,omitempty"`
	ResponseDeadline           time.Time       `json:"response_deadline"`
	ExtendedDeadline           *time.Time      `json:"extended_deadline,omitempty"`
	ExtensionReason            string          `json:"extension_reason,omitempty"`
	ExtensionNotifiedAt        *time.Time      `json:"extension_notified_at,omitempty"`
	AssignedTo                 *uuid.UUID      `json:"assigned_to,omitempty"`
	DataSystemsAffected        []string        `json:"data_systems_affected,omitempty"`
	DataCategoriesAffected     []string        `json:"data_categories_affected,omitempty"`
	ThirdPartiesNotified       []string        `json:"third_parties_notified,omitempty"`
	ProcessingNotes            string          `json:"processing_notes,omitempty"`
	CompletedAt                *time.Time      `json:"completed_at,omitempty"`
	CompletedBy                *uuid.UUID      `json:"completed_by,omitempty"`
	ResponseMethod             string          `json:"response_method,omitempty"`
	ResponseDocumentPath       string          `json:"response_document_path,omitempty"`
	RejectionReason            string          `json:"rejection_reason,omitempty"`
	RejectionLegalBasis        string          `json:"rejection_legal_basis,omitempty"`
	SLAStatus                  string          `json:"sla_status"`
	DaysRemaining              int             `json:"days_remaining"`
	WasExtended                bool            `json:"was_extended"`
	WasCompletedOnTime         *bool           `json:"was_completed_on_time,omitempty"`
	CreatedAt                  time.Time       `json:"created_at"`
	UpdatedAt                  time.Time       `json:"updated_at"`
	Tasks                      []DSRTask       `json:"tasks,omitempty"`
	AuditTrail                 []DSRAuditEntry `json:"audit_trail,omitempty"`
}

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

type DSRAuditEntry struct {
	ID             uuid.UUID  `json:"id"`
	OrganizationID uuid.UUID  `json:"organization_id"`
	DSRRequestID   uuid.UUID  `json:"dsr_request_id"`
	Action         string     `json:"action"`
	PerformedBy    *uuid.UUID `json:"performed_by,omitempty"`
	Description    string     `json:"description"`
	CreatedAt      time.Time  `json:"created_at"`
}

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

type CreateDSRRequest struct {
	RequestType        string     `json:"request_type"`
	Priority           string     `json:"priority"`
	DataSubjectName    string     `json:"data_subject_name"`
	DataSubjectEmail   string     `json:"data_subject_email"`
	DataSubjectPhone   string     `json:"data_subject_phone,omitempty"`
	DataSubjectAddress string     `json:"data_subject_address,omitempty"`
	RequestDescription string     `json:"request_description,omitempty"`
	RequestSource      string     `json:"request_source"`
	ReceivedDate       *time.Time `json:"received_date,omitempty"`
	AssignedTo         *uuid.UUID `json:"assigned_to,omitempty"`
}

type UpdateDSRRequest struct {
	Priority               string     `json:"priority,omitempty"`
	RequestDescription     string     `json:"request_description,omitempty"`
	DataSystemsAffected    []string   `json:"data_systems_affected,omitempty"`
	DataCategoriesAffected []string   `json:"data_categories_affected,omitempty"`
	ThirdPartiesNotified   []string   `json:"third_parties_notified,omitempty"`
	ProcessingNotes        string     `json:"processing_notes,omitempty"`
	AssignedTo             *uuid.UUID `json:"assigned_to,omitempty"`
}

type DSRDashboard struct {
	TotalRequests          int64   `json:"total_requests"`
	ReceivedCount          int64   `json:"received_count"`
	VerificationCount      int64   `json:"verification_count"`
	InProgressCount        int64   `json:"in_progress_count"`
	ExtendedCount          int64   `json:"extended_count"`
	CompletedCount         int64   `json:"completed_count"`
	RejectedCount          int64   `json:"rejected_count"`
	WithdrawnCount         int64   `json:"withdrawn_count"`
	OnTrackCount           int64   `json:"on_track_count"`
	AtRiskCount            int64   `json:"at_risk_count"`
	OverdueCount           int64   `json:"overdue_count"`
	AccessCount            int64   `json:"access_count"`
	ErasureCount           int64   `json:"erasure_count"`
	RectificationCount     int64   `json:"rectification_count"`
	PortabilityCount       int64   `json:"portability_count"`
	RestrictionCount       int64   `json:"restriction_count"`
	ObjectionCount         int64   `json:"objection_count"`
	AutomatedDecisionCount int64   `json:"automated_decision_count"`
	AvgCompletionDays      float64 `json:"avg_completion_days"`
	CompletedOnTimeCount   int64   `json:"completed_on_time_count"`
	CompletedLateCount     int64   `json:"completed_late_count"`
}

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
	{"extract_in_machine_readable", "Extract data in structured machine-readable format", 3},
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
	case "access":        return accessTaskChecklist
	case "erasure":       return erasureTaskChecklist
	case "rectification": return rectificationTaskChecklist
	case "portability":   return portabilityTaskChecklist
	case "restriction":   return restrictionTaskChecklist
	case "objection":     return objectionTaskChecklist
	case "automated_decision": return automatedDecisionTaskChecklist
	default:              return accessTaskChecklist
	}
}

// DSRService provides business logic for GDPR Data Subject Requests.
type DSRService struct {
	pool      *pgxpool.Pool
	encryptor *crypto.Encryptor
}

// NewDSRService creates a new DSRService.
func NewDSRService(pool *pgxpool.Pool, enc *crypto.Encryptor) *DSRService {
	return &DSRService{pool: pool, encryptor: enc}
}

func (s *DSRService) CreateRequest(ctx context.Context, orgID, userID uuid.UUID, req CreateDSRRequest) (*DSRRequest, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil { return nil, fmt.Errorf("begin tx: %w", err) }
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_tenant', $1, true)", orgID.String()); err != nil {
		return nil, fmt.Errorf("set tenant: %w", err)
	}
	var requestRef string
	if err := tx.QueryRow(ctx, "SELECT generate_dsr_ref($1)", orgID).Scan(&requestRef); err != nil {
		return nil, fmt.Errorf("generate ref: %w", err)
	}
	nameEnc, err := s.encryptor.EncryptString(req.DataSubjectName)
	if err != nil { return nil, fmt.Errorf("encrypt name: %w", err) }
	emailEnc, err := s.encryptor.EncryptString(req.DataSubjectEmail)
	if err != nil { return nil, fmt.Errorf("encrypt email: %w", err) }
	var phoneEnc, addressEnc string
	if req.DataSubjectPhone != "" {
		if phoneEnc, err = s.encryptor.EncryptString(req.DataSubjectPhone); err != nil { return nil, err }
	}
	if req.DataSubjectAddress != "" {
		if addressEnc, err = s.encryptor.EncryptString(req.DataSubjectAddress); err != nil { return nil, err }
	}
	receivedDate := time.Now()
	if req.ReceivedDate != nil { receivedDate = *req.ReceivedDate }
	responseDeadline := receivedDate.AddDate(0, 0, 30)
	priority := req.Priority
	if priority == "" { priority = "standard" }
	now := time.Now()
	requestID := uuid.New()
	_, err = tx.Exec(ctx, `INSERT INTO dsr_requests (id,organization_id,request_ref,request_type,status,priority,data_subject_name_encrypted,data_subject_email_encrypted,data_subject_phone_encrypted,data_subject_address_encrypted,request_description,request_source,received_date,acknowledged_at,response_deadline,assigned_to,metadata,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)`,
		requestID, orgID, requestRef, req.RequestType, "received", priority, nameEnc, emailEnc, phoneEnc, addressEnc, req.RequestDescription, req.RequestSource, receivedDate.Format("2006-01-02"), now, responseDeadline.Format("2006-01-02"), req.AssignedTo, "{}", now, now)
	if err != nil { return nil, fmt.Errorf("insert dsr: %w", err) }
	for _, t := range getTaskChecklist(req.RequestType) {
		_, err = tx.Exec(ctx, `INSERT INTO dsr_tasks (id,organization_id,dsr_request_id,task_type,description,status,due_date,sort_order,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
			uuid.New(), orgID, requestID, t.TaskType, t.Description, "pending", responseDeadline.Format("2006-01-02"), t.SortOrder, now, now)
		if err != nil { return nil, fmt.Errorf("insert task %s: %w", t.TaskType, err) }
	}
	if err := s.logAudit(ctx, tx, orgID, requestID, &userID, "request_created", fmt.Sprintf("DSR %s created: %s via %s", requestRef, req.RequestType, req.RequestSource)); err != nil { return nil, err }
	if err := tx.Commit(ctx); err != nil { return nil, fmt.Errorf("commit: %w", err) }
	log.Info().Str("ref", requestRef).Str("org", orgID.String()).Str("type", req.RequestType).Msg("DSR created")
	return &DSRRequest{ID: requestID, OrganizationID: orgID, RequestRef: requestRef, RequestType: req.RequestType, Status: "received", Priority: priority, DataSubjectName: req.DataSubjectName, DataSubjectEmail: req.DataSubjectEmail, DataSubjectPhone: req.DataSubjectPhone, DataSubjectAddress: req.DataSubjectAddress, RequestDescription: req.RequestDescription, RequestSource: req.RequestSource, ReceivedDate: receivedDate, AcknowledgedAt: &now, ResponseDeadline: responseDeadline, AssignedTo: req.AssignedTo, SLAStatus: "on_track", DaysRemaining: int(time.Until(responseDeadline).Hours() / 24), CreatedAt: now, UpdatedAt: now}, nil
}

func (s *DSRService) VerifyIdentity(ctx context.Context, orgID, requestID uuid.UUID, method string, verifiedBy uuid.UUID) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil { return err }
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_tenant', $1, true)", orgID.String()); err != nil { return err }
	now := time.Now()
	tag, err := tx.Exec(ctx, `UPDATE dsr_requests SET data_subject_id_verified=true, identity_verification_method=$1, identity_verified_at=$2, identity_verified_by=$3, status=CASE WHEN status='received' THEN 'in_progress'::dsr_status ELSE status END WHERE id=$4 AND organization_id=$5 AND deleted_at IS NULL`, method, now, verifiedBy, requestID, orgID)
	if err != nil { return err }
	if tag.RowsAffected() == 0 { return fmt.Errorf("not found") }
	if err := s.logAudit(ctx, tx, orgID, requestID, &verifiedBy, "identity_verified", fmt.Sprintf("Verified via %s", method)); err != nil { return err }
	return tx.Commit(ctx)
}

func (s *DSRService) AssignRequest(ctx context.Context, orgID, requestID, assigneeID, performedBy uuid.UUID) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil { return err }
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_tenant', $1, true)", orgID.String()); err != nil { return err }
	tag, err := tx.Exec(ctx, `UPDATE dsr_requests SET assigned_to=$1 WHERE id=$2 AND organization_id=$3 AND deleted_at IS NULL`, assigneeID, requestID, orgID)
	if err != nil { return err }
	if tag.RowsAffected() == 0 { return fmt.Errorf("not found") }
	if err := s.logAudit(ctx, tx, orgID, requestID, &performedBy, "assigned", fmt.Sprintf("Assigned to %s", assigneeID)); err != nil { return err }
	return tx.Commit(ctx)
}

func (s *DSRService) ExtendDeadline(ctx context.Context, orgID, requestID uuid.UUID, reason string, performedBy uuid.UUID) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil { return err }
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_tenant', $1, true)", orgID.String()); err != nil { return err }
	var dl time.Time
	if err := tx.QueryRow(ctx, `SELECT response_deadline FROM dsr_requests WHERE id=$1 AND organization_id=$2 AND deleted_at IS NULL`, requestID, orgID).Scan(&dl); err != nil { return err }
	ext := dl.AddDate(0, 0, 60)
	now := time.Now()
	tag, err := tx.Exec(ctx, `UPDATE dsr_requests SET status='extended'::dsr_status, extended_deadline=$1, extension_reason=$2, extension_notified_at=$3, was_extended=true WHERE id=$4 AND organization_id=$5 AND deleted_at IS NULL`, ext.Format("2006-01-02"), reason, now, requestID, orgID)
	if err != nil { return err }
	if tag.RowsAffected() == 0 { return fmt.Errorf("not found") }
	if err := s.logAudit(ctx, tx, orgID, requestID, &performedBy, "deadline_extended", fmt.Sprintf("Extended to %s: %s", ext.Format("2006-01-02"), reason)); err != nil { return err }
	return tx.Commit(ctx)
}

func (s *DSRService) UpdateTask(ctx context.Context, orgID, taskID uuid.UUID, status, notes, evidencePath string, assignedTo *uuid.UUID, performedBy uuid.UUID) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil { return err }
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_tenant', $1, true)", orgID.String()); err != nil { return err }
	now := time.Now()
	var cat *time.Time
	var cby *uuid.UUID
	if status == "completed" { cat = &now; cby = &performedBy }
	var rid uuid.UUID
	var tt string
	if err := tx.QueryRow(ctx, `UPDATE dsr_tasks SET status=$1::dsr_task_status, notes=CASE WHEN $2='' THEN notes ELSE $2 END, evidence_path=CASE WHEN $3='' THEN evidence_path ELSE $3 END, assigned_to=COALESCE($4,assigned_to), completed_at=COALESCE($5,completed_at), completed_by=COALESCE($6,completed_by) WHERE id=$7 AND organization_id=$8 RETURNING dsr_request_id, task_type`, status, notes, evidencePath, assignedTo, cat, cby, taskID, orgID).Scan(&rid, &tt); err != nil { return err }
	if err := s.logAudit(ctx, tx, orgID, rid, &performedBy, "task_updated", fmt.Sprintf("Task '%s' -> %s", tt, status)); err != nil { return err }
	return tx.Commit(ctx)
}

func (s *DSRService) CompleteRequest(ctx context.Context, orgID, requestID uuid.UUID, responseMethod, documentPath string, completedBy uuid.UUID) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil { return err }
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_tenant', $1, true)", orgID.String()); err != nil { return err }
	now := time.Now()
	var dl time.Time
	var edl *time.Time
	if err := tx.QueryRow(ctx, `SELECT response_deadline, extended_deadline FROM dsr_requests WHERE id=$1 AND organization_id=$2 AND deleted_at IS NULL`, requestID, orgID).Scan(&dl, &edl); err != nil { return err }
	eff := dl
	if edl != nil { eff = *edl }
	onTime := now.Before(eff.AddDate(0, 0, 1))
	tag, err := tx.Exec(ctx, `UPDATE dsr_requests SET status='completed'::dsr_status, completed_at=$1, completed_by=$2, response_method=$3::dsr_response_method, response_document_path=$4, was_completed_on_time=$5, sla_status='on_track'::dsr_sla_status, days_remaining=0 WHERE id=$6 AND organization_id=$7 AND deleted_at IS NULL`, now, completedBy, responseMethod, documentPath, onTime, requestID, orgID)
	if err != nil { return err }
	if tag.RowsAffected() == 0 { return fmt.Errorf("not found") }
	if err := s.logAudit(ctx, tx, orgID, requestID, &completedBy, "completed", fmt.Sprintf("Completed via %s, on_time=%v", responseMethod, onTime)); err != nil { return err }
	return tx.Commit(ctx)
}

func (s *DSRService) RejectRequest(ctx context.Context, orgID, requestID uuid.UUID, reason, legalBasis string, rejectedBy uuid.UUID) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil { return err }
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_tenant', $1, true)", orgID.String()); err != nil { return err }
	now := time.Now()
	tag, err := tx.Exec(ctx, `UPDATE dsr_requests SET status='rejected'::dsr_status, rejection_reason=$1, rejection_legal_basis=$2, completed_at=$3, completed_by=$4 WHERE id=$5 AND organization_id=$6 AND deleted_at IS NULL`, reason, legalBasis, now, rejectedBy, requestID, orgID)
	if err != nil { return err }
	if tag.RowsAffected() == 0 { return fmt.Errorf("not found") }
	if err := s.logAudit(ctx, tx, orgID, requestID, &rejectedBy, "rejected", fmt.Sprintf("Reason: %s. Basis: %s", reason, legalBasis)); err != nil { return err }
	return tx.Commit(ctx)
}

func (s *DSRService) UpdateRequest(ctx context.Context, orgID, requestID uuid.UUID, req UpdateDSRRequest, performedBy uuid.UUID) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil { return err }
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_tenant', $1, true)", orgID.String()); err != nil { return err }
	tag, err := tx.Exec(ctx, `UPDATE dsr_requests SET priority=CASE WHEN $1='' THEN priority ELSE $1::dsr_priority END, request_description=CASE WHEN $2='' THEN request_description ELSE $2 END, data_systems_affected=CASE WHEN $3::text[] IS NULL THEN data_systems_affected ELSE $3 END, data_categories_affected=CASE WHEN $4::text[] IS NULL THEN data_categories_affected ELSE $4 END, third_parties_notified=CASE WHEN $5::text[] IS NULL THEN third_parties_notified ELSE $5 END, processing_notes=CASE WHEN $6='' THEN processing_notes ELSE $6 END, assigned_to=COALESCE($7,assigned_to) WHERE id=$8 AND organization_id=$9 AND deleted_at IS NULL`, req.Priority, req.RequestDescription, req.DataSystemsAffected, req.DataCategoriesAffected, req.ThirdPartiesNotified, req.ProcessingNotes, req.AssignedTo, requestID, orgID)
	if err != nil { return err }
	if tag.RowsAffected() == 0 { return fmt.Errorf("not found") }
	if err := s.logAudit(ctx, tx, orgID, requestID, &performedBy, "updated", "Request updated"); err != nil { return err }
	return tx.Commit(ctx)
}

func (s *DSRService) GetRequest(ctx context.Context, orgID, requestID uuid.UUID) (*DSRRequest, error) {
	d := &DSRRequest{}
	var ne, ee, pe, ae string
	var rm, rr, rlb, er, pn, rd *string
	var wot *bool
	err := s.pool.QueryRow(ctx, `SELECT id,organization_id,request_ref,request_type::text,status::text,priority::text,COALESCE(data_subject_name_encrypted,''),COALESCE(data_subject_email_encrypted,''),COALESCE(data_subject_phone_encrypted,''),COALESCE(data_subject_address_encrypted,''),data_subject_id_verified,COALESCE(identity_verification_method,''),identity_verified_at,identity_verified_by,request_description,request_source::text,received_date,acknowledged_at,response_deadline,extended_deadline,extension_reason,extension_notified_at,assigned_to,data_systems_affected,data_categories_affected,third_parties_notified,processing_notes,completed_at,completed_by,response_method::text,response_document_path,rejection_reason,rejection_legal_basis,COALESCE(sla_status::text,'on_track'),COALESCE(days_remaining,0),was_extended,was_completed_on_time,created_at,updated_at FROM dsr_requests WHERE id=$1 AND organization_id=$2 AND deleted_at IS NULL`, requestID, orgID).Scan(&d.ID, &d.OrganizationID, &d.RequestRef, &d.RequestType, &d.Status, &d.Priority, &ne, &ee, &pe, &ae, &d.DataSubjectIDVerified, &d.IdentityVerificationMethod, &d.IdentityVerifiedAt, &d.IdentityVerifiedBy, &rd, &d.RequestSource, &d.ReceivedDate, &d.AcknowledgedAt, &d.ResponseDeadline, &d.ExtendedDeadline, &er, &d.ExtensionNotifiedAt, &d.AssignedTo, &d.DataSystemsAffected, &d.DataCategoriesAffected, &d.ThirdPartiesNotified, &pn, &d.CompletedAt, &d.CompletedBy, &rm, &d.ResponseDocumentPath, &rr, &rlb, &d.SLAStatus, &d.DaysRemaining, &d.WasExtended, &wot, &d.CreatedAt, &d.UpdatedAt)
	if err != nil { return nil, fmt.Errorf("not found: %w", err) }
	if rd != nil { d.RequestDescription = *rd }
	if er != nil { d.ExtensionReason = *er }
	if pn != nil { d.ProcessingNotes = *pn }
	if rm != nil { d.ResponseMethod = *rm }
	if rr != nil { d.RejectionReason = *rr }
	if rlb != nil { d.RejectionLegalBasis = *rlb }
	if wot != nil { d.WasCompletedOnTime = wot }
	if ne != "" { d.DataSubjectName, _ = s.encryptor.DecryptString(ne) }
	if ee != "" { d.DataSubjectEmail, _ = s.encryptor.DecryptString(ee) }
	if pe != "" { d.DataSubjectPhone, _ = s.encryptor.DecryptString(pe) }
	if ae != "" { d.DataSubjectAddress, _ = s.encryptor.DecryptString(ae) }
	if tasks, err := s.getTasksByRequestID(ctx, orgID, requestID); err == nil { d.Tasks = tasks }
	if trail, err := s.getAuditTrail(ctx, orgID, requestID); err == nil { d.AuditTrail = trail }
	return d, nil
}

func (s *DSRService) ListRequests(ctx context.Context, orgID uuid.UUID, page, pageSize int, status, requestType, slaStatus string) ([]DSRRequest, int64, error) {
	cq := `SELECT COUNT(*) FROM dsr_requests WHERE organization_id=$1 AND deleted_at IS NULL`
	a := []interface{}{orgID}
	ai := 2
	if status != "" { cq += fmt.Sprintf(" AND status=$%d::dsr_status", ai); a = append(a, status); ai++ }
	if requestType != "" { cq += fmt.Sprintf(" AND request_type=$%d::dsr_request_type", ai); a = append(a, requestType); ai++ }
	if slaStatus != "" { cq += fmt.Sprintf(" AND sla_status=$%d::dsr_sla_status", ai); a = append(a, slaStatus); ai++ }
	var total int64
	if err := s.pool.QueryRow(ctx, cq, a...).Scan(&total); err != nil { return nil, 0, err }
	q := `SELECT id,organization_id,request_ref,request_type::text,status::text,priority::text,data_subject_id_verified,request_source::text,received_date,response_deadline,extended_deadline,assigned_to,COALESCE(sla_status::text,'on_track'),COALESCE(days_remaining,0),was_extended,was_completed_on_time,completed_at,created_at,updated_at FROM dsr_requests WHERE organization_id=$1 AND deleted_at IS NULL`
	la := []interface{}{orgID}
	li := 2
	if status != "" { q += fmt.Sprintf(" AND status=$%d::dsr_status", li); la = append(la, status); li++ }
	if requestType != "" { q += fmt.Sprintf(" AND request_type=$%d::dsr_request_type", li); la = append(la, requestType); li++ }
	if slaStatus != "" { q += fmt.Sprintf(" AND sla_status=$%d::dsr_sla_status", li); la = append(la, slaStatus); li++ }
	q += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", li, li+1)
	la = append(la, pageSize, (page-1)*pageSize)
	rows, err := s.pool.Query(ctx, q, la...)
	if err != nil { return nil, 0, err }
	defer rows.Close()
	var dsrs []DSRRequest
	for rows.Next() {
		var d DSRRequest
		var wot *bool
		if err := rows.Scan(&d.ID, &d.OrganizationID, &d.RequestRef, &d.RequestType, &d.Status, &d.Priority, &d.DataSubjectIDVerified, &d.RequestSource, &d.ReceivedDate, &d.ResponseDeadline, &d.ExtendedDeadline, &d.AssignedTo, &d.SLAStatus, &d.DaysRemaining, &d.WasExtended, &wot, &d.CompletedAt, &d.CreatedAt, &d.UpdatedAt); err != nil { return nil, 0, err }
		if wot != nil { d.WasCompletedOnTime = wot }
		dsrs = append(dsrs, d)
	}
	return dsrs, total, nil
}

func (s *DSRService) GetDSRDashboard(ctx context.Context, orgID uuid.UUID) (*DSRDashboard, error) {
	d := &DSRDashboard{}
	err := s.pool.QueryRow(ctx, `SELECT COALESCE(total_requests,0),COALESCE(received_count,0),COALESCE(verification_count,0),COALESCE(in_progress_count,0),COALESCE(extended_count,0),COALESCE(completed_count,0),COALESCE(rejected_count,0),COALESCE(withdrawn_count,0),COALESCE(on_track_count,0),COALESCE(at_risk_count,0),COALESCE(overdue_count,0),COALESCE(access_count,0),COALESCE(erasure_count,0),COALESCE(rectification_count,0),COALESCE(portability_count,0),COALESCE(restriction_count,0),COALESCE(objection_count,0),COALESCE(automated_decision_count,0),COALESCE(avg_completion_days,0),COALESCE(completed_on_time_count,0),COALESCE(completed_late_count,0) FROM v_dsr_dashboard WHERE organization_id=$1`, orgID).Scan(&d.TotalRequests, &d.ReceivedCount, &d.VerificationCount, &d.InProgressCount, &d.ExtendedCount, &d.CompletedCount, &d.RejectedCount, &d.WithdrawnCount, &d.OnTrackCount, &d.AtRiskCount, &d.OverdueCount, &d.AccessCount, &d.ErasureCount, &d.RectificationCount, &d.PortabilityCount, &d.RestrictionCount, &d.ObjectionCount, &d.AutomatedDecisionCount, &d.AvgCompletionDays, &d.CompletedOnTimeCount, &d.CompletedLateCount)
	if err != nil {
		if err == pgx.ErrNoRows { return d, nil }
		return nil, err
	}
	return d, nil
}

func (s *DSRService) CheckSLACompliance(ctx context.Context, orgID uuid.UUID) ([]DSRRequest, error) {
	rows, err := s.pool.Query(ctx, `SELECT id,organization_id,request_ref,request_type::text,status::text,priority::text,request_source::text,received_date,response_deadline,extended_deadline,assigned_to,COALESCE(sla_status::text,'on_track'),COALESCE(days_remaining,0),was_extended,created_at,updated_at FROM dsr_requests WHERE organization_id=$1 AND deleted_at IS NULL AND status NOT IN ('completed','rejected','withdrawn') AND (sla_status='at_risk' OR sla_status='overdue') ORDER BY days_remaining ASC`, orgID)
	if err != nil { return nil, err }
	defer rows.Close()
	var dsrs []DSRRequest
	for rows.Next() {
		var d DSRRequest
		if err := rows.Scan(&d.ID, &d.OrganizationID, &d.RequestRef, &d.RequestType, &d.Status, &d.Priority, &d.RequestSource, &d.ReceivedDate, &d.ResponseDeadline, &d.ExtendedDeadline, &d.AssignedTo, &d.SLAStatus, &d.DaysRemaining, &d.WasExtended, &d.CreatedAt, &d.UpdatedAt); err != nil { return nil, err }
		dsrs = append(dsrs, d)
	}
	return dsrs, nil
}

func (s *DSRService) GetOverdueRequests(ctx context.Context, orgID uuid.UUID) ([]DSRRequest, error) {
	rows, err := s.pool.Query(ctx, `SELECT id,organization_id,request_ref,request_type::text,status::text,priority::text,request_source::text,received_date,response_deadline,extended_deadline,assigned_to,COALESCE(sla_status::text,'overdue'),COALESCE(days_remaining,0),was_extended,created_at,updated_at FROM dsr_requests WHERE organization_id=$1 AND deleted_at IS NULL AND status NOT IN ('completed','rejected','withdrawn') AND sla_status='overdue' ORDER BY response_deadline ASC`, orgID)
	if err != nil { return nil, err }
	defer rows.Close()
	var dsrs []DSRRequest
	for rows.Next() {
		var d DSRRequest
		if err := rows.Scan(&d.ID, &d.OrganizationID, &d.RequestRef, &d.RequestType, &d.Status, &d.Priority, &d.RequestSource, &d.ReceivedDate, &d.ResponseDeadline, &d.ExtendedDeadline, &d.AssignedTo, &d.SLAStatus, &d.DaysRemaining, &d.WasExtended, &d.CreatedAt, &d.UpdatedAt); err != nil { return nil, err }
		dsrs = append(dsrs, d)
	}
	return dsrs, nil
}

func (s *DSRService) GetResponseTemplates(ctx context.Context, orgID uuid.UUID) ([]DSRResponseTemplate, error) {
	rows, err := s.pool.Query(ctx, `SELECT id,organization_id,request_type::text,name,subject,body_html,body_text,is_system,language,created_at,updated_at FROM dsr_response_templates WHERE (organization_id=$1 OR organization_id IS NULL) ORDER BY is_system DESC, request_type, name`, orgID)
	if err != nil { return nil, err }
	defer rows.Close()
	var ts []DSRResponseTemplate
	for rows.Next() {
		var t DSRResponseTemplate
		if err := rows.Scan(&t.ID, &t.OrganizationID, &t.RequestType, &t.Name, &t.Subject, &t.BodyHTML, &t.BodyText, &t.IsSystem, &t.Language, &t.CreatedAt, &t.UpdatedAt); err != nil { return nil, err }
		ts = append(ts, t)
	}
	return ts, nil
}

// CalculateSLAStatus computes the SLA status for a given deadline.
func CalculateSLAStatus(now time.Time, responseDeadline time.Time, extendedDeadline *time.Time) (string, int) {
	eff := responseDeadline
	if extendedDeadline != nil && !extendedDeadline.IsZero() { eff = *extendedDeadline }
	dr := int(eff.Sub(now).Hours() / 24)
	if dr < 0 { return "overdue", dr }
	if dr <= 7 { return "at_risk", dr }
	return "on_track", dr
}

// UpdateSLAStatuses batch-updates SLA fields for all active DSR requests in an organization.
func (s *DSRService) UpdateSLAStatuses(ctx context.Context, orgID uuid.UUID) (int64, error) {
	tag, err := s.pool.Exec(ctx, `UPDATE dsr_requests SET days_remaining=CASE WHEN extended_deadline IS NOT NULL THEN EXTRACT(DAY FROM (extended_deadline-$1::date))::int ELSE EXTRACT(DAY FROM (response_deadline-$1::date))::int END, sla_status=CASE WHEN COALESCE(extended_deadline,response_deadline)<$1::date THEN 'overdue'::dsr_sla_status WHEN COALESCE(extended_deadline,response_deadline)-$1::date<=7 THEN 'at_risk'::dsr_sla_status ELSE 'on_track'::dsr_sla_status END WHERE organization_id=$2 AND deleted_at IS NULL AND status NOT IN ('completed','rejected','withdrawn')`, time.Now().Format("2006-01-02"), orgID)
	if err != nil { return 0, err }
	return tag.RowsAffected(), nil
}

// GetTasks returns all tasks for a given DSR request.
func (s *DSRService) GetTasks(ctx context.Context, orgID, requestID uuid.UUID) ([]DSRTask, error) {
	return s.getTasksByRequestID(ctx, orgID, requestID)
}

// GetAuditTrail returns the audit trail for a given DSR request.
func (s *DSRService) GetAuditTrail(ctx context.Context, orgID, requestID uuid.UUID) ([]DSRAuditEntry, error) {
	return s.getAuditTrail(ctx, orgID, requestID)
}

func (s *DSRService) logAudit(ctx context.Context, tx pgx.Tx, orgID, requestID uuid.UUID, performedBy *uuid.UUID, action, description string) error {
	_, err := tx.Exec(ctx, `INSERT INTO dsr_audit_trail (id,organization_id,dsr_request_id,action,performed_by,description,metadata,created_at) VALUES ($1,$2,$3,$4,$5,$6,'{}',$7)`, uuid.New(), orgID, requestID, action, performedBy, description, time.Now())
	return err
}

func (s *DSRService) getTasksByRequestID(ctx context.Context, orgID, requestID uuid.UUID) ([]DSRTask, error) {
	rows, err := s.pool.Query(ctx, `SELECT id,organization_id,dsr_request_id,task_type::text,COALESCE(description,''),COALESCE(system_name,''),assigned_to,status::text,due_date,completed_at,completed_by,COALESCE(notes,''),COALESCE(evidence_path,''),sort_order,created_at,updated_at FROM dsr_tasks WHERE dsr_request_id=$1 AND organization_id=$2 ORDER BY sort_order ASC`, requestID, orgID)
	if err != nil { return nil, err }
	defer rows.Close()
	var tasks []DSRTask
	for rows.Next() {
		var t DSRTask
		if err := rows.Scan(&t.ID, &t.OrganizationID, &t.DSRRequestID, &t.TaskType, &t.Description, &t.SystemName, &t.AssignedTo, &t.Status, &t.DueDate, &t.CompletedAt, &t.CompletedBy, &t.Notes, &t.EvidencePath, &t.SortOrder, &t.CreatedAt, &t.UpdatedAt); err != nil { return nil, err }
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (s *DSRService) getAuditTrail(ctx context.Context, orgID, requestID uuid.UUID) ([]DSRAuditEntry, error) {
	rows, err := s.pool.Query(ctx, `SELECT id,organization_id,dsr_request_id,action,performed_by,COALESCE(description,''),created_at FROM dsr_audit_trail WHERE dsr_request_id=$1 AND organization_id=$2 ORDER BY created_at DESC`, requestID, orgID)
	if err != nil { return nil, err }
	defer rows.Close()
	var es []DSRAuditEntry
	for rows.Next() {
		var e DSRAuditEntry
		if err := rows.Scan(&e.ID, &e.OrganizationID, &e.DSRRequestID, &e.Action, &e.PerformedBy, &e.Description, &e.CreatedAt); err != nil { return nil, err }
		es = append(es, e)
	}
	return es, nil
}
