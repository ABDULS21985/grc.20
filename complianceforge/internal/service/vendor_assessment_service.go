package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// ============================================================
// MODEL TYPES
// ============================================================

// VendorAssessment tracks the lifecycle of a single vendor assessment.
type VendorAssessment struct {
	ID                    uuid.UUID       `json:"id"`
	OrganizationID        uuid.UUID       `json:"organization_id"`
	VendorID              uuid.UUID       `json:"vendor_id"`
	QuestionnaireID       uuid.UUID       `json:"questionnaire_id"`
	AssessmentRef         string          `json:"assessment_ref"`
	Status                string          `json:"status"`
	SentAt                *time.Time      `json:"sent_at"`
	SentToEmail           string          `json:"sent_to_email"`
	SentToName            string          `json:"sent_to_name"`
	ReminderCount         int             `json:"reminder_count"`
	LastReminderAt        *time.Time      `json:"last_reminder_at"`
	DueDate               *time.Time      `json:"due_date"`
	SubmittedAt           *time.Time      `json:"submitted_at"`
	OverallScore          *float64        `json:"overall_score"`
	RiskRating            *string         `json:"risk_rating"`
	SectionScores         json.RawMessage `json:"section_scores"`
	CriticalFindings      int             `json:"critical_findings"`
	HighFindings          int             `json:"high_findings"`
	PassFail              string          `json:"pass_fail"`
	ReviewedBy            *uuid.UUID      `json:"reviewed_by"`
	ReviewedAt            *time.Time      `json:"reviewed_at"`
	ReviewNotes           string          `json:"review_notes"`
	ReviewerOverrideScore *float64        `json:"reviewer_override_score"`
	ReviewerOverrideReason string         `json:"reviewer_override_reason"`
	FollowUpRequired      bool            `json:"follow_up_required"`
	FollowUpItems         json.RawMessage `json:"follow_up_items"`
	NextAssessmentDate    *time.Time      `json:"next_assessment_date"`
	Metadata              json.RawMessage `json:"metadata"`
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
	// Joined fields
	VendorName            string          `json:"vendor_name,omitempty"`
	QuestionnaireName     string          `json:"questionnaire_name,omitempty"`
	Responses             []VendorAssessmentResponse `json:"responses,omitempty"`
}

// VendorAssessmentResponse is a single response from the vendor.
type VendorAssessmentResponse struct {
	ID              uuid.UUID  `json:"id"`
	OrganizationID  uuid.UUID  `json:"organization_id"`
	AssessmentID    uuid.UUID  `json:"assessment_id"`
	QuestionID      uuid.UUID  `json:"question_id"`
	AnswerValue     string     `json:"answer_value"`
	AnswerScore     *float64   `json:"answer_score"`
	EvidencePaths   []string   `json:"evidence_paths"`
	EvidenceNotes   string     `json:"evidence_notes"`
	ReviewerComment string     `json:"reviewer_comment"`
	ReviewerFlag    string     `json:"reviewer_flag"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	// Joined fields
	QuestionText    string     `json:"question_text,omitempty"`
	QuestionType    string     `json:"question_type,omitempty"`
	SectionName     string     `json:"section_name,omitempty"`
	RiskImpact      string     `json:"risk_impact,omitempty"`
}

// AssessmentDashboard provides at-a-glance TPRM metrics.
type AssessmentDashboard struct {
	TotalAssessments    int                `json:"total_assessments"`
	InProgress          int                `json:"in_progress"`
	AwaitingReview      int                `json:"awaiting_review"`
	Completed           int                `json:"completed"`
	Overdue             int                `json:"overdue"`
	PassRate            float64            `json:"pass_rate"`
	AvgScore            float64            `json:"avg_score"`
	StatusBreakdown     map[string]int     `json:"status_breakdown"`
	RiskBreakdown       map[string]int     `json:"risk_breakdown"`
	RecentAssessments   []VendorAssessment `json:"recent_assessments"`
	UpcomingDueDates    []DueDateEntry     `json:"upcoming_due_dates"`
	ScoreDistribution   map[string]int     `json:"score_distribution"`
}

// DueDateEntry represents an upcoming assessment due date.
type DueDateEntry struct {
	AssessmentID   uuid.UUID `json:"assessment_id"`
	AssessmentRef  string    `json:"assessment_ref"`
	VendorName     string    `json:"vendor_name"`
	DueDate        time.Time `json:"due_date"`
	DaysRemaining  int       `json:"days_remaining"`
	Status         string    `json:"status"`
}

// ProgressResult reports progress for the vendor portal.
type ProgressResult struct {
	TotalQuestions int     `json:"total_questions"`
	Answered       int     `json:"answered"`
	Required       int     `json:"required"`
	RequiredDone   int     `json:"required_done"`
	Percentage     float64 `json:"percentage"`
	IsComplete     bool    `json:"is_complete"`
}

// ============================================================
// REQUEST TYPES
// ============================================================

// ResponseInput holds a single response from the vendor.
type ResponseInput struct {
	QuestionID    uuid.UUID `json:"question_id"`
	AnswerValue   string    `json:"answer_value"`
	EvidencePaths []string  `json:"evidence_paths"`
	EvidenceNotes string    `json:"evidence_notes"`
}

// ReviewInput holds reviewer feedback for an assessment.
type ReviewInput struct {
	ReviewNotes           string                `json:"review_notes"`
	OverrideScore         *float64              `json:"override_score"`
	OverrideReason        string                `json:"override_reason"`
	PassFail              string                `json:"pass_fail"`
	FollowUpRequired      bool                  `json:"follow_up_required"`
	FollowUpItems         json.RawMessage       `json:"follow_up_items"`
	NextAssessmentDate    *time.Time            `json:"next_assessment_date"`
	ResponseFlags         []ResponseFlagInput   `json:"response_flags"`
}

// ResponseFlagInput is reviewer flag/comment on an individual response.
type ResponseFlagInput struct {
	ResponseID uuid.UUID `json:"response_id"`
	Flag       string    `json:"flag"`
	Comment    string    `json:"comment"`
}

// AssessmentFilter holds filters for listing assessments.
type AssessmentFilter struct {
	VendorID         *uuid.UUID `json:"vendor_id"`
	QuestionnaireID  *uuid.UUID `json:"questionnaire_id"`
	Status           string     `json:"status"`
	PassFail         string     `json:"pass_fail"`
	RiskRating       string     `json:"risk_rating"`
	Search           string     `json:"search"`
	Page             int        `json:"page"`
	PageSize         int        `json:"page_size"`
}

// ============================================================
// SERVICE
// ============================================================

// VendorAssessmentService manages the vendor assessment lifecycle.
type VendorAssessmentService struct {
	pool *pgxpool.Pool
	qSvc *QuestionnaireService
}

// NewVendorAssessmentService creates a new VendorAssessmentService.
func NewVendorAssessmentService(pool *pgxpool.Pool) *VendorAssessmentService {
	return &VendorAssessmentService{
		pool: pool,
		qSvc: NewQuestionnaireService(pool),
	}
}

// ============================================================
// SEND ASSESSMENT
// ============================================================

// SendAssessment creates and sends an assessment to a vendor contact.
func (s *VendorAssessmentService) SendAssessment(
	ctx context.Context, orgID, vendorID, questionnaireID uuid.UUID,
	dueDate time.Time, contactEmail string,
) (*VendorAssessment, error) {
	if contactEmail == "" {
		return nil, fmt.Errorf("contact email is required")
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return nil, fmt.Errorf("set org context: %w", err)
	}

	// Verify vendor exists
	var vendorName string
	err = tx.QueryRow(ctx, `SELECT name FROM vendors WHERE id = $1 AND organization_id = $2`,
		vendorID, orgID).Scan(&vendorName)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("vendor not found")
		}
		return nil, fmt.Errorf("check vendor: %w", err)
	}

	// Verify questionnaire exists and is active
	var qName string
	err = tx.QueryRow(ctx, `
		SELECT name FROM assessment_questionnaires
		WHERE id = $1 AND status = 'active'`, questionnaireID).Scan(&qName)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("questionnaire not found or not active")
		}
		return nil, fmt.Errorf("check questionnaire: %w", err)
	}

	// Generate 32-byte random access token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}
	rawToken := hex.EncodeToString(tokenBytes)

	// Store SHA-256 hash of the token
	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])

	var va VendorAssessment
	err = tx.QueryRow(ctx, `
		INSERT INTO vendor_assessments
			(organization_id, vendor_id, questionnaire_id, status,
			 sent_at, sent_to_email, sent_to_name, access_token_hash, due_date)
		VALUES ($1, $2, $3, 'sent', NOW(), $4, $5, $6, $7)
		RETURNING id, organization_id, vendor_id, questionnaire_id, assessment_ref,
		          status, sent_at, sent_to_email, sent_to_name, reminder_count,
		          last_reminder_at, due_date, submitted_at, overall_score, risk_rating,
		          section_scores, critical_findings, high_findings, pass_fail,
		          reviewed_by, reviewed_at, review_notes,
		          reviewer_override_score, reviewer_override_reason,
		          follow_up_required, follow_up_items, next_assessment_date,
		          metadata, created_at, updated_at`,
		orgID, vendorID, questionnaireID, contactEmail, vendorName, tokenHash, dueDate,
	).Scan(
		&va.ID, &va.OrganizationID, &va.VendorID, &va.QuestionnaireID, &va.AssessmentRef,
		&va.Status, &va.SentAt, &va.SentToEmail, &va.SentToName, &va.ReminderCount,
		&va.LastReminderAt, &va.DueDate, &va.SubmittedAt, &va.OverallScore, &va.RiskRating,
		&va.SectionScores, &va.CriticalFindings, &va.HighFindings, &va.PassFail,
		&va.ReviewedBy, &va.ReviewedAt, &va.ReviewNotes,
		&va.ReviewerOverrideScore, &va.ReviewerOverrideReason,
		&va.FollowUpRequired, &va.FollowUpItems, &va.NextAssessmentDate,
		&va.Metadata, &va.CreatedAt, &va.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert assessment: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	va.VendorName = vendorName
	va.QuestionnaireName = qName

	// Attach the raw token so the caller can send it in an email.
	// It is NOT stored in the database — only the hash is.
	va.Metadata, _ = json.Marshal(map[string]string{
		"access_token": rawToken,
		"portal_url":   fmt.Sprintf("/vendor-portal/%s", rawToken),
	})

	log.Info().
		Str("assessment_id", va.ID.String()).
		Str("assessment_ref", va.AssessmentRef).
		Str("vendor", vendorName).
		Str("email", contactEmail).
		Msg("vendor assessment sent")

	return &va, nil
}

// ============================================================
// GET ASSESSMENT
// ============================================================

// GetAssessment returns a full assessment with responses.
func (s *VendorAssessmentService) GetAssessment(ctx context.Context, orgID, assessmentID uuid.UUID) (*VendorAssessment, error) {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("acquire conn: %w", err)
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return nil, fmt.Errorf("set org context: %w", err)
	}

	va, err := s.scanAssessment(ctx, conn, `
		SELECT va.id, va.organization_id, va.vendor_id, va.questionnaire_id, va.assessment_ref,
		       va.status, va.sent_at, va.sent_to_email, va.sent_to_name, va.reminder_count,
		       va.last_reminder_at, va.due_date, va.submitted_at, va.overall_score, va.risk_rating,
		       va.section_scores, va.critical_findings, va.high_findings, va.pass_fail,
		       va.reviewed_by, va.reviewed_at, va.review_notes,
		       va.reviewer_override_score, va.reviewer_override_reason,
		       va.follow_up_required, va.follow_up_items, va.next_assessment_date,
		       va.metadata, va.created_at, va.updated_at,
		       COALESCE(v.name, '') AS vendor_name,
		       COALESCE(aq.name, '') AS questionnaire_name
		FROM vendor_assessments va
		LEFT JOIN vendors v ON va.vendor_id = v.id
		LEFT JOIN assessment_questionnaires aq ON va.questionnaire_id = aq.id
		WHERE va.id = $1 AND va.organization_id = $2`, assessmentID, orgID)
	if err != nil {
		return nil, err
	}

	// Fetch responses with question details
	respRows, err := conn.Query(ctx, `
		SELECT var.id, var.organization_id, var.assessment_id, var.question_id,
		       var.answer_value, var.answer_score, var.evidence_paths, var.evidence_notes,
		       var.reviewer_comment, var.reviewer_flag, var.created_at, var.updated_at,
		       qq.question_text, qq.question_type,
		       qs.name AS section_name, qq.risk_impact
		FROM vendor_assessment_responses var
		JOIN questionnaire_questions qq ON var.question_id = qq.id
		JOIN questionnaire_sections qs ON qq.section_id = qs.id
		WHERE var.assessment_id = $1
		ORDER BY qs.sort_order, qq.sort_order`, assessmentID)
	if err != nil {
		return nil, fmt.Errorf("query responses: %w", err)
	}
	defer respRows.Close()

	for respRows.Next() {
		var r VendorAssessmentResponse
		if err := respRows.Scan(
			&r.ID, &r.OrganizationID, &r.AssessmentID, &r.QuestionID,
			&r.AnswerValue, &r.AnswerScore, &r.EvidencePaths, &r.EvidenceNotes,
			&r.ReviewerComment, &r.ReviewerFlag, &r.CreatedAt, &r.UpdatedAt,
			&r.QuestionText, &r.QuestionType, &r.SectionName, &r.RiskImpact,
		); err != nil {
			return nil, fmt.Errorf("scan response: %w", err)
		}
		va.Responses = append(va.Responses, r)
	}
	if err := respRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate responses: %w", err)
	}
	if va.Responses == nil {
		va.Responses = []VendorAssessmentResponse{}
	}

	return va, nil
}

// ============================================================
// LIST ASSESSMENTS
// ============================================================

// ListAssessments returns filtered assessments for an organisation.
func (s *VendorAssessmentService) ListAssessments(ctx context.Context, orgID uuid.UUID, filter AssessmentFilter) ([]VendorAssessment, int64, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 20
	}
	offset := (filter.Page - 1) * filter.PageSize

	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("acquire conn: %w", err)
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return nil, 0, fmt.Errorf("set org context: %w", err)
	}

	var conditions []string
	var args []interface{}
	argIdx := 1

	conditions = append(conditions, fmt.Sprintf("va.organization_id = $%d", argIdx))
	args = append(args, orgID)
	argIdx++

	if filter.VendorID != nil {
		conditions = append(conditions, fmt.Sprintf("va.vendor_id = $%d", argIdx))
		args = append(args, *filter.VendorID)
		argIdx++
	}
	if filter.QuestionnaireID != nil {
		conditions = append(conditions, fmt.Sprintf("va.questionnaire_id = $%d", argIdx))
		args = append(args, *filter.QuestionnaireID)
		argIdx++
	}
	if filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("va.status = $%d", argIdx))
		args = append(args, filter.Status)
		argIdx++
	}
	if filter.PassFail != "" {
		conditions = append(conditions, fmt.Sprintf("va.pass_fail = $%d", argIdx))
		args = append(args, filter.PassFail)
		argIdx++
	}
	if filter.RiskRating != "" {
		conditions = append(conditions, fmt.Sprintf("va.risk_rating = $%d", argIdx))
		args = append(args, filter.RiskRating)
		argIdx++
	}
	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf(
			"(va.assessment_ref ILIKE '%%' || $%d || '%%' OR v.name ILIKE '%%' || $%d || '%%')",
			argIdx, argIdx))
		args = append(args, filter.Search)
		argIdx++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count
	countSQL := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM vendor_assessments va
		LEFT JOIN vendors v ON va.vendor_id = v.id
		%s`, where)
	var total int64
	if err := conn.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count assessments: %w", err)
	}

	// Data
	dataSQL := fmt.Sprintf(`
		SELECT va.id, va.organization_id, va.vendor_id, va.questionnaire_id, va.assessment_ref,
		       va.status, va.sent_at, va.sent_to_email, va.sent_to_name, va.reminder_count,
		       va.last_reminder_at, va.due_date, va.submitted_at, va.overall_score, va.risk_rating,
		       va.section_scores, va.critical_findings, va.high_findings, va.pass_fail,
		       va.reviewed_by, va.reviewed_at, va.review_notes,
		       va.reviewer_override_score, va.reviewer_override_reason,
		       va.follow_up_required, va.follow_up_items, va.next_assessment_date,
		       va.metadata, va.created_at, va.updated_at,
		       COALESCE(v.name, '') AS vendor_name,
		       COALESCE(aq.name, '') AS questionnaire_name
		FROM vendor_assessments va
		LEFT JOIN vendors v ON va.vendor_id = v.id
		LEFT JOIN assessment_questionnaires aq ON va.questionnaire_id = aq.id
		%s
		ORDER BY va.created_at DESC
		LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)
	args = append(args, filter.PageSize, offset)

	rows, err := conn.Query(ctx, dataSQL, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query assessments: %w", err)
	}
	defer rows.Close()

	var list []VendorAssessment
	for rows.Next() {
		var va VendorAssessment
		if err := rows.Scan(
			&va.ID, &va.OrganizationID, &va.VendorID, &va.QuestionnaireID, &va.AssessmentRef,
			&va.Status, &va.SentAt, &va.SentToEmail, &va.SentToName, &va.ReminderCount,
			&va.LastReminderAt, &va.DueDate, &va.SubmittedAt, &va.OverallScore, &va.RiskRating,
			&va.SectionScores, &va.CriticalFindings, &va.HighFindings, &va.PassFail,
			&va.ReviewedBy, &va.ReviewedAt, &va.ReviewNotes,
			&va.ReviewerOverrideScore, &va.ReviewerOverrideReason,
			&va.FollowUpRequired, &va.FollowUpItems, &va.NextAssessmentDate,
			&va.Metadata, &va.CreatedAt, &va.UpdatedAt,
			&va.VendorName, &va.QuestionnaireName,
		); err != nil {
			return nil, 0, fmt.Errorf("scan assessment: %w", err)
		}
		list = append(list, va)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate assessments: %w", err)
	}

	if list == nil {
		list = []VendorAssessment{}
	}
	return list, total, nil
}

// ============================================================
// GET ASSESSMENT BY TOKEN (vendor portal)
// ============================================================

// GetAssessmentByToken retrieves an assessment using the portal token hash.
// This is used by the public vendor portal — no org context required.
func (s *VendorAssessmentService) GetAssessmentByToken(ctx context.Context, tokenHash string) (*VendorAssessment, error) {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("acquire conn: %w", err)
	}
	defer conn.Release()

	va, err := s.scanAssessmentNoRLS(ctx, conn, `
		SELECT va.id, va.organization_id, va.vendor_id, va.questionnaire_id, va.assessment_ref,
		       va.status, va.sent_at, va.sent_to_email, va.sent_to_name, va.reminder_count,
		       va.last_reminder_at, va.due_date, va.submitted_at, va.overall_score, va.risk_rating,
		       va.section_scores, va.critical_findings, va.high_findings, va.pass_fail,
		       va.reviewed_by, va.reviewed_at, va.review_notes,
		       va.reviewer_override_score, va.reviewer_override_reason,
		       va.follow_up_required, va.follow_up_items, va.next_assessment_date,
		       va.metadata, va.created_at, va.updated_at,
		       COALESCE(v.name, '') AS vendor_name,
		       COALESCE(aq.name, '') AS questionnaire_name
		FROM vendor_assessments va
		LEFT JOIN vendors v ON va.vendor_id = v.id
		LEFT JOIN assessment_questionnaires aq ON va.questionnaire_id = aq.id
		WHERE va.access_token_hash = $1
		  AND va.status IN ('sent', 'in_progress')`, tokenHash)
	if err != nil {
		return nil, err
	}

	return va, nil
}

// ============================================================
// SAVE RESPONSES (vendor portal)
// ============================================================

// SaveResponses saves or updates vendor responses for an assessment.
func (s *VendorAssessmentService) SaveResponses(ctx context.Context, tokenHash string, responses []ResponseInput) error {
	if len(responses) == 0 {
		return nil
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Validate token and get assessment
	var assessmentID, orgID uuid.UUID
	var qID uuid.UUID
	err = tx.QueryRow(ctx, `
		SELECT id, organization_id, questionnaire_id
		FROM vendor_assessments
		WHERE access_token_hash = $1
		  AND status IN ('sent', 'in_progress')`, tokenHash).Scan(&assessmentID, &orgID, &qID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("assessment not found or not accessible")
		}
		return fmt.Errorf("validate token: %w", err)
	}

	// Build set of valid question IDs for this questionnaire
	validQIDs := make(map[uuid.UUID]bool)
	qRows, err := tx.Query(ctx, `
		SELECT qq.id FROM questionnaire_questions qq
		JOIN questionnaire_sections qs ON qq.section_id = qs.id
		WHERE qs.questionnaire_id = $1`, qID)
	if err != nil {
		return fmt.Errorf("query questions: %w", err)
	}
	for qRows.Next() {
		var id uuid.UUID
		if qRows.Scan(&id) == nil {
			validQIDs[id] = true
		}
	}
	qRows.Close()

	// Build a map of question options for score lookup
	optionScores := make(map[uuid.UUID]map[string]float64)
	optRows, err := tx.Query(ctx, `
		SELECT qq.id, qq.options FROM questionnaire_questions qq
		JOIN questionnaire_sections qs ON qq.section_id = qs.id
		WHERE qs.questionnaire_id = $1`, qID)
	if err != nil {
		return fmt.Errorf("query options: %w", err)
	}
	for optRows.Next() {
		var id uuid.UUID
		var optRaw json.RawMessage
		if optRows.Scan(&id, &optRaw) == nil {
			var opts []struct {
				Value string  `json:"value"`
				Score float64 `json:"score"`
			}
			if json.Unmarshal(optRaw, &opts) == nil {
				scores := make(map[string]float64)
				for _, opt := range opts {
					scores[opt.Value] = opt.Score
				}
				optionScores[id] = scores
			}
		}
	}
	optRows.Close()

	// Upsert each response
	for _, resp := range responses {
		if !validQIDs[resp.QuestionID] {
			continue // skip invalid question IDs
		}

		// Determine score from options
		var answerScore *float64
		if scores, ok := optionScores[resp.QuestionID]; ok {
			if sc, found := scores[resp.AnswerValue]; found {
				answerScore = &sc
			}
		}

		if resp.EvidencePaths == nil {
			resp.EvidencePaths = []string{}
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO vendor_assessment_responses
				(organization_id, assessment_id, question_id, answer_value, answer_score,
				 evidence_paths, evidence_notes)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (assessment_id, question_id) DO UPDATE SET
				answer_value = EXCLUDED.answer_value,
				answer_score = EXCLUDED.answer_score,
				evidence_paths = EXCLUDED.evidence_paths,
				evidence_notes = EXCLUDED.evidence_notes,
				updated_at = NOW()`,
			orgID, assessmentID, resp.QuestionID, resp.AnswerValue, answerScore,
			resp.EvidencePaths, resp.EvidenceNotes,
		)
		if err != nil {
			return fmt.Errorf("upsert response: %w", err)
		}
	}

	// Update status to in_progress if still sent
	_, _ = tx.Exec(ctx, `
		UPDATE vendor_assessments SET status = 'in_progress'
		WHERE id = $1 AND status = 'sent'`, assessmentID)

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	return nil
}

// ============================================================
// SUBMIT ASSESSMENT (vendor portal)
// ============================================================

// SubmitAssessment marks an assessment as submitted by the vendor.
// Validates that all required questions are answered, then calculates scores.
func (s *VendorAssessmentService) SubmitAssessment(ctx context.Context, tokenHash string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Get assessment
	var assessmentID, orgID, qID uuid.UUID
	err = tx.QueryRow(ctx, `
		SELECT id, organization_id, questionnaire_id
		FROM vendor_assessments
		WHERE access_token_hash = $1
		  AND status IN ('sent', 'in_progress')`, tokenHash).Scan(&assessmentID, &orgID, &qID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("assessment not found or already submitted")
		}
		return fmt.Errorf("validate token: %w", err)
	}

	// Check all required questions are answered
	var unanswered int
	err = tx.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM questionnaire_questions qq
		JOIN questionnaire_sections qs ON qq.section_id = qs.id
		LEFT JOIN vendor_assessment_responses var
			ON var.question_id = qq.id AND var.assessment_id = $1
		WHERE qs.questionnaire_id = $2
		  AND qq.is_required = true
		  AND (var.answer_value IS NULL OR var.answer_value = '')`,
		assessmentID, qID).Scan(&unanswered)
	if err != nil {
		return fmt.Errorf("check required questions: %w", err)
	}
	if unanswered > 0 {
		return fmt.Errorf("%d required questions remain unanswered", unanswered)
	}

	// Update status
	_, err = tx.Exec(ctx, `
		UPDATE vendor_assessments SET
			status = 'submitted', submitted_at = NOW()
		WHERE id = $1`, assessmentID)
	if err != nil {
		return fmt.Errorf("update status: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	// Calculate scores asynchronously (best-effort)
	go func() {
		scoreCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if _, err := s.qSvc.CalculateScore(scoreCtx, assessmentID); err != nil {
			log.Warn().Err(err).
				Str("assessment_id", assessmentID.String()).
				Msg("failed to auto-calculate score on submission")
		}
	}()

	log.Info().
		Str("assessment_id", assessmentID.String()).
		Msg("vendor assessment submitted")

	return nil
}

// ============================================================
// REVIEW ASSESSMENT
// ============================================================

// ReviewAssessment records reviewer feedback and final verdict on an assessment.
func (s *VendorAssessmentService) ReviewAssessment(ctx context.Context, orgID, assessmentID uuid.UUID, review ReviewInput) error {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire conn: %w", err)
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return fmt.Errorf("set org context: %w", err)
	}

	reviewerID := contextUserID(ctx)
	if review.FollowUpItems == nil {
		review.FollowUpItems = json.RawMessage(`[]`)
	}

	// Determine final pass_fail
	passFail := review.PassFail
	if passFail == "" {
		passFail = "pending"
	}

	tag, err := conn.Exec(ctx, `
		UPDATE vendor_assessments SET
			status = 'completed',
			reviewed_by = $1, reviewed_at = NOW(), review_notes = $2,
			reviewer_override_score = $3, reviewer_override_reason = $4,
			pass_fail = $5, follow_up_required = $6, follow_up_items = $7,
			next_assessment_date = $8
		WHERE id = $9 AND organization_id = $10
		  AND status IN ('submitted', 'under_review')`,
		reviewerID, review.ReviewNotes,
		review.OverrideScore, review.OverrideReason,
		passFail, review.FollowUpRequired, review.FollowUpItems,
		review.NextAssessmentDate,
		assessmentID, orgID,
	)
	if err != nil {
		return fmt.Errorf("update assessment: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("assessment not found or not in reviewable state")
	}

	// Update individual response flags
	for _, rf := range review.ResponseFlags {
		_, _ = conn.Exec(ctx, `
			UPDATE vendor_assessment_responses SET
				reviewer_flag = $1, reviewer_comment = $2
			WHERE id = $3 AND assessment_id = $4`,
			rf.Flag, rf.Comment, rf.ResponseID, assessmentID,
		)
	}

	log.Info().
		Str("assessment_id", assessmentID.String()).
		Str("pass_fail", passFail).
		Msg("vendor assessment reviewed")

	return nil
}

// ============================================================
// SEND REMINDER
// ============================================================

// SendReminder increments the reminder count and updates last_reminder_at.
func (s *VendorAssessmentService) SendReminder(ctx context.Context, orgID, assessmentID uuid.UUID) error {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire conn: %w", err)
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return fmt.Errorf("set org context: %w", err)
	}

	tag, err := conn.Exec(ctx, `
		UPDATE vendor_assessments SET
			reminder_count = reminder_count + 1,
			last_reminder_at = NOW()
		WHERE id = $1 AND organization_id = $2
		  AND status IN ('sent', 'in_progress')`, assessmentID, orgID)
	if err != nil {
		return fmt.Errorf("send reminder: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("assessment not found or not in remindable state")
	}

	log.Info().
		Str("assessment_id", assessmentID.String()).
		Msg("vendor assessment reminder sent")

	return nil
}

// ============================================================
// DASHBOARD
// ============================================================

// GetAssessmentDashboard returns aggregated TPRM metrics for an organisation.
func (s *VendorAssessmentService) GetAssessmentDashboard(ctx context.Context, orgID uuid.UUID) (*AssessmentDashboard, error) {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("acquire conn: %w", err)
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return nil, fmt.Errorf("set org context: %w", err)
	}

	dash := &AssessmentDashboard{
		StatusBreakdown:   make(map[string]int),
		RiskBreakdown:     make(map[string]int),
		ScoreDistribution: make(map[string]int),
	}

	// Total assessments
	_ = conn.QueryRow(ctx, `
		SELECT COUNT(*) FROM vendor_assessments WHERE organization_id = $1`,
		orgID).Scan(&dash.TotalAssessments)

	// In progress
	_ = conn.QueryRow(ctx, `
		SELECT COUNT(*) FROM vendor_assessments
		WHERE organization_id = $1 AND status IN ('sent', 'in_progress')`,
		orgID).Scan(&dash.InProgress)

	// Awaiting review
	_ = conn.QueryRow(ctx, `
		SELECT COUNT(*) FROM vendor_assessments
		WHERE organization_id = $1 AND status = 'submitted'`,
		orgID).Scan(&dash.AwaitingReview)

	// Completed
	_ = conn.QueryRow(ctx, `
		SELECT COUNT(*) FROM vendor_assessments
		WHERE organization_id = $1 AND status = 'completed'`,
		orgID).Scan(&dash.Completed)

	// Overdue
	_ = conn.QueryRow(ctx, `
		SELECT COUNT(*) FROM vendor_assessments
		WHERE organization_id = $1 AND status IN ('sent', 'in_progress')
		  AND due_date < CURRENT_DATE`,
		orgID).Scan(&dash.Overdue)

	// Pass rate
	var passCount, completedTotal int
	_ = conn.QueryRow(ctx, `
		SELECT COUNT(*) FILTER (WHERE pass_fail IN ('pass', 'conditional_pass')),
		       COUNT(*)
		FROM vendor_assessments
		WHERE organization_id = $1 AND status = 'completed'`,
		orgID).Scan(&passCount, &completedTotal)
	if completedTotal > 0 {
		dash.PassRate = float64(passCount) / float64(completedTotal) * 100
	}

	// Average score
	_ = conn.QueryRow(ctx, `
		SELECT COALESCE(AVG(overall_score), 0)
		FROM vendor_assessments
		WHERE organization_id = $1 AND overall_score IS NOT NULL`,
		orgID).Scan(&dash.AvgScore)

	// Status breakdown
	statRows, err := conn.Query(ctx, `
		SELECT status::text, COUNT(*)
		FROM vendor_assessments
		WHERE organization_id = $1
		GROUP BY status`, orgID)
	if err == nil {
		defer statRows.Close()
		for statRows.Next() {
			var st string
			var cnt int
			if statRows.Scan(&st, &cnt) == nil {
				dash.StatusBreakdown[st] = cnt
			}
		}
	}

	// Risk rating breakdown
	riskRows, err := conn.Query(ctx, `
		SELECT COALESCE(risk_rating, 'unrated'), COUNT(*)
		FROM vendor_assessments
		WHERE organization_id = $1
		GROUP BY risk_rating`, orgID)
	if err == nil {
		defer riskRows.Close()
		for riskRows.Next() {
			var rating string
			var cnt int
			if riskRows.Scan(&rating, &cnt) == nil {
				dash.RiskBreakdown[rating] = cnt
			}
		}
	}

	// Score distribution (0-20, 20-40, 40-60, 60-80, 80-100)
	scoreRows, err := conn.Query(ctx, `
		SELECT
			CASE
				WHEN overall_score < 20 THEN '0-20'
				WHEN overall_score < 40 THEN '20-40'
				WHEN overall_score < 60 THEN '40-60'
				WHEN overall_score < 80 THEN '60-80'
				ELSE '80-100'
			END AS bucket,
			COUNT(*)
		FROM vendor_assessments
		WHERE organization_id = $1 AND overall_score IS NOT NULL
		GROUP BY bucket
		ORDER BY bucket`, orgID)
	if err == nil {
		defer scoreRows.Close()
		for scoreRows.Next() {
			var bucket string
			var cnt int
			if scoreRows.Scan(&bucket, &cnt) == nil {
				dash.ScoreDistribution[bucket] = cnt
			}
		}
	}

	// Recent assessments (last 10)
	recentRows, err := conn.Query(ctx, `
		SELECT va.id, va.assessment_ref, va.vendor_id,
		       COALESCE(v.name, '') AS vendor_name,
		       va.status, va.overall_score, va.risk_rating, va.pass_fail,
		       va.due_date, va.submitted_at, va.created_at
		FROM vendor_assessments va
		LEFT JOIN vendors v ON va.vendor_id = v.id
		WHERE va.organization_id = $1
		ORDER BY va.created_at DESC
		LIMIT 10`, orgID)
	if err == nil {
		defer recentRows.Close()
		for recentRows.Next() {
			var a VendorAssessment
			if recentRows.Scan(
				&a.ID, &a.AssessmentRef, &a.VendorID, &a.VendorName,
				&a.Status, &a.OverallScore, &a.RiskRating, &a.PassFail,
				&a.DueDate, &a.SubmittedAt, &a.CreatedAt,
			) == nil {
				dash.RecentAssessments = append(dash.RecentAssessments, a)
			}
		}
	}
	if dash.RecentAssessments == nil {
		dash.RecentAssessments = []VendorAssessment{}
	}

	// Upcoming due dates (next 30 days)
	dueRows, err := conn.Query(ctx, `
		SELECT va.id, va.assessment_ref, COALESCE(v.name, ''), va.due_date, va.status
		FROM vendor_assessments va
		LEFT JOIN vendors v ON va.vendor_id = v.id
		WHERE va.organization_id = $1
		  AND va.status IN ('sent', 'in_progress')
		  AND va.due_date IS NOT NULL
		  AND va.due_date >= CURRENT_DATE
		  AND va.due_date <= CURRENT_DATE + INTERVAL '30 days'
		ORDER BY va.due_date ASC
		LIMIT 10`, orgID)
	if err == nil {
		defer dueRows.Close()
		for dueRows.Next() {
			var d DueDateEntry
			if dueRows.Scan(&d.AssessmentID, &d.AssessmentRef, &d.VendorName, &d.DueDate, &d.Status) == nil {
				d.DaysRemaining = int(time.Until(d.DueDate).Hours() / 24)
				dash.UpcomingDueDates = append(dash.UpcomingDueDates, d)
			}
		}
	}
	if dash.UpcomingDueDates == nil {
		dash.UpcomingDueDates = []DueDateEntry{}
	}

	return dash, nil
}

// ============================================================
// GET PROGRESS (vendor portal)
// ============================================================

// GetProgress returns completion progress for an assessment.
func (s *VendorAssessmentService) GetProgress(ctx context.Context, tokenHash string) (*ProgressResult, error) {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("acquire conn: %w", err)
	}
	defer conn.Release()

	var assessmentID, qID uuid.UUID
	err = conn.QueryRow(ctx, `
		SELECT id, questionnaire_id FROM vendor_assessments
		WHERE access_token_hash = $1
		  AND status IN ('sent', 'in_progress', 'submitted')`,
		tokenHash).Scan(&assessmentID, &qID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("assessment not found")
		}
		return nil, fmt.Errorf("get assessment: %w", err)
	}

	result := &ProgressResult{}

	// Total questions and required count
	err = conn.QueryRow(ctx, `
		SELECT COUNT(*),
		       COUNT(*) FILTER (WHERE qq.is_required = true)
		FROM questionnaire_questions qq
		JOIN questionnaire_sections qs ON qq.section_id = qs.id
		WHERE qs.questionnaire_id = $1`, qID).Scan(&result.TotalQuestions, &result.Required)
	if err != nil {
		return nil, fmt.Errorf("count questions: %w", err)
	}

	// Answered and required-answered counts
	err = conn.QueryRow(ctx, `
		SELECT COUNT(*),
		       COUNT(*) FILTER (WHERE qq.is_required = true)
		FROM vendor_assessment_responses var
		JOIN questionnaire_questions qq ON var.question_id = qq.id
		WHERE var.assessment_id = $1 AND var.answer_value != ''`, assessmentID).Scan(&result.Answered, &result.RequiredDone)
	if err != nil {
		return nil, fmt.Errorf("count answered: %w", err)
	}

	if result.TotalQuestions > 0 {
		result.Percentage = float64(result.Answered) / float64(result.TotalQuestions) * 100
	}
	result.IsComplete = result.RequiredDone >= result.Required

	return result, nil
}

// ============================================================
// HELPERS
// ============================================================

// HashToken computes SHA-256 of a raw hex token string.
func HashToken(rawToken string) string {
	hash := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(hash[:])
}

// scanAssessment scans a single VendorAssessment from a query row (with RLS join).
func (s *VendorAssessmentService) scanAssessment(ctx context.Context, conn *pgxpool.Conn, query string, args ...interface{}) (*VendorAssessment, error) {
	var va VendorAssessment
	err := conn.QueryRow(ctx, query, args...).Scan(
		&va.ID, &va.OrganizationID, &va.VendorID, &va.QuestionnaireID, &va.AssessmentRef,
		&va.Status, &va.SentAt, &va.SentToEmail, &va.SentToName, &va.ReminderCount,
		&va.LastReminderAt, &va.DueDate, &va.SubmittedAt, &va.OverallScore, &va.RiskRating,
		&va.SectionScores, &va.CriticalFindings, &va.HighFindings, &va.PassFail,
		&va.ReviewedBy, &va.ReviewedAt, &va.ReviewNotes,
		&va.ReviewerOverrideScore, &va.ReviewerOverrideReason,
		&va.FollowUpRequired, &va.FollowUpItems, &va.NextAssessmentDate,
		&va.Metadata, &va.CreatedAt, &va.UpdatedAt,
		&va.VendorName, &va.QuestionnaireName,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("assessment not found")
		}
		return nil, fmt.Errorf("scan assessment: %w", err)
	}
	return &va, nil
}

// scanAssessmentNoRLS is the same as scanAssessment but does not require
// RLS context — used for public vendor portal endpoints.
func (s *VendorAssessmentService) scanAssessmentNoRLS(ctx context.Context, conn *pgxpool.Conn, query string, args ...interface{}) (*VendorAssessment, error) {
	return s.scanAssessment(ctx, conn, query, args...)
}
