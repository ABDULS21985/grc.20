package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
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

// ComplianceException represents a compliance exception with optional
// compensating controls.
type ComplianceException struct {
	ID                            uuid.UUID        `json:"id"`
	OrganizationID                uuid.UUID        `json:"organization_id"`
	ExceptionRef                  string           `json:"exception_ref"`
	Title                         string           `json:"title"`
	Description                   string           `json:"description"`
	ExceptionType                 string           `json:"exception_type"`
	Status                        string           `json:"status"`
	Priority                      string           `json:"priority"`
	ScopeType                     string           `json:"scope_type"`
	ControlImplementationIDs      []uuid.UUID      `json:"control_implementation_ids"`
	FrameworkControlCodes         []string         `json:"framework_control_codes"`
	PolicyID                      *uuid.UUID       `json:"policy_id"`
	ScopeDescription              string           `json:"scope_description"`
	RiskJustification             string           `json:"risk_justification"`
	ResidualRiskDescription       string           `json:"residual_risk_description"`
	ResidualRiskLevel             string           `json:"residual_risk_level"`
	RiskAssessmentID              *uuid.UUID       `json:"risk_assessment_id"`
	RiskAcceptedBy                *uuid.UUID       `json:"risk_accepted_by"`
	RiskAcceptedAt                *time.Time       `json:"risk_accepted_at"`
	HasCompensatingControls       bool             `json:"has_compensating_controls"`
	CompensatingControlsDescription string         `json:"compensating_controls_description"`
	CompensatingControlIDs        []uuid.UUID      `json:"compensating_control_ids"`
	CompensatingEffectiveness     string           `json:"compensating_effectiveness"`
	RequestedBy                   uuid.UUID        `json:"requested_by"`
	RequestedAt                   time.Time        `json:"requested_at"`
	ApprovedBy                    *uuid.UUID       `json:"approved_by"`
	ApprovedAt                    *time.Time       `json:"approved_at"`
	ApprovalComments              string           `json:"approval_comments"`
	RejectionReason               string           `json:"rejection_reason"`
	WorkflowInstanceID            *uuid.UUID       `json:"workflow_instance_id"`
	EffectiveDate                 time.Time        `json:"effective_date"`
	ExpiryDate                    *time.Time       `json:"expiry_date"`
	ReviewFrequencyMonths         int              `json:"review_frequency_months"`
	NextReviewDate                *time.Time       `json:"next_review_date"`
	LastReviewDate                *time.Time       `json:"last_review_date"`
	LastReviewedBy                *uuid.UUID       `json:"last_reviewed_by"`
	RenewalCount                  int              `json:"renewal_count"`
	Conditions                    string           `json:"conditions"`
	BusinessImpactIfImplemented   string           `json:"business_impact_if_implemented"`
	RegulatoryNotificationRequired bool            `json:"regulatory_notification_required"`
	RegulatorNotifiedAt           *time.Time       `json:"regulator_notified_at"`
	AuditEvidencePath             string           `json:"audit_evidence_path"`
	Tags                          []string         `json:"tags"`
	Metadata                      json.RawMessage  `json:"metadata"`
	CreatedAt                     time.Time        `json:"created_at"`
	UpdatedAt                     time.Time        `json:"updated_at"`
	DeletedAt                     *time.Time       `json:"deleted_at,omitempty"`
}

// ExceptionReview represents a periodic or ad-hoc review of an exception.
type ExceptionReview struct {
	ID                   uuid.UUID       `json:"id"`
	OrganizationID       uuid.UUID       `json:"organization_id"`
	ExceptionID          uuid.UUID       `json:"exception_id"`
	ReviewType           string          `json:"review_type"`
	ReviewerID           uuid.UUID       `json:"reviewer_id"`
	ReviewDate           time.Time       `json:"review_date"`
	Outcome              string          `json:"outcome"`
	RiskLevelAtReview    string          `json:"risk_level_at_review"`
	CompensatingEffective *bool          `json:"compensating_effective"`
	Findings             string          `json:"findings"`
	Recommendations      string          `json:"recommendations"`
	NextReviewDate       *time.Time      `json:"next_review_date"`
	Attachments          []string        `json:"attachments"`
	Metadata             json.RawMessage `json:"metadata"`
	CreatedAt            time.Time       `json:"created_at"`
}

// ExceptionAuditEntry represents an entry in the exception audit trail.
type ExceptionAuditEntry struct {
	ID              uuid.UUID       `json:"id"`
	OrganizationID  uuid.UUID       `json:"organization_id"`
	ExceptionID     uuid.UUID       `json:"exception_id"`
	Action          string          `json:"action"`
	ActorID         uuid.UUID       `json:"actor_id"`
	ActorEmail      string          `json:"actor_email"`
	OldStatus       string          `json:"old_status"`
	NewStatus       string          `json:"new_status"`
	Details         string          `json:"details"`
	IPAddress       string          `json:"ip_address"`
	UserAgent       string          `json:"user_agent"`
	Metadata        json.RawMessage `json:"metadata"`
	CreatedAt       time.Time       `json:"created_at"`
}

// ============================================================
// REQUEST / RESPONSE TYPES
// ============================================================

// CreateExceptionRequest is the request body for creating a new exception.
type CreateExceptionRequest struct {
	Title                         string      `json:"title"`
	Description                   string      `json:"description"`
	ExceptionType                 string      `json:"exception_type"`
	Priority                      string      `json:"priority"`
	ScopeType                     string      `json:"scope_type"`
	ControlImplementationIDs      []uuid.UUID `json:"control_implementation_ids"`
	FrameworkControlCodes         []string    `json:"framework_control_codes"`
	PolicyID                      *uuid.UUID  `json:"policy_id"`
	ScopeDescription              string      `json:"scope_description"`
	RiskJustification             string      `json:"risk_justification"`
	ResidualRiskDescription       string      `json:"residual_risk_description"`
	ResidualRiskLevel             string      `json:"residual_risk_level"`
	HasCompensatingControls       bool        `json:"has_compensating_controls"`
	CompensatingControlsDescription string   `json:"compensating_controls_description"`
	CompensatingControlIDs        []uuid.UUID `json:"compensating_control_ids"`
	CompensatingEffectiveness     string      `json:"compensating_effectiveness"`
	EffectiveDate                 string      `json:"effective_date"`
	ExpiryDate                    string      `json:"expiry_date"`
	ReviewFrequencyMonths         int         `json:"review_frequency_months"`
	Conditions                    string      `json:"conditions"`
	BusinessImpactIfImplemented   string      `json:"business_impact_if_implemented"`
	RegulatoryNotificationRequired bool       `json:"regulatory_notification_required"`
	Tags                          []string    `json:"tags"`
	Metadata                      json.RawMessage `json:"metadata"`
}

// UpdateExceptionRequest is the request body for updating an existing exception.
type UpdateExceptionRequest struct {
	Title                         *string     `json:"title"`
	Description                   *string     `json:"description"`
	ExceptionType                 *string     `json:"exception_type"`
	Priority                      *string     `json:"priority"`
	ScopeType                     *string     `json:"scope_type"`
	ControlImplementationIDs      []uuid.UUID `json:"control_implementation_ids"`
	FrameworkControlCodes         []string    `json:"framework_control_codes"`
	PolicyID                      *uuid.UUID  `json:"policy_id"`
	ScopeDescription              *string     `json:"scope_description"`
	RiskJustification             *string     `json:"risk_justification"`
	ResidualRiskDescription       *string     `json:"residual_risk_description"`
	ResidualRiskLevel             *string     `json:"residual_risk_level"`
	HasCompensatingControls       *bool       `json:"has_compensating_controls"`
	CompensatingControlsDescription *string   `json:"compensating_controls_description"`
	CompensatingControlIDs        []uuid.UUID `json:"compensating_control_ids"`
	CompensatingEffectiveness     *string     `json:"compensating_effectiveness"`
	EffectiveDate                 *string     `json:"effective_date"`
	ExpiryDate                    *string     `json:"expiry_date"`
	ReviewFrequencyMonths         *int        `json:"review_frequency_months"`
	Conditions                    *string     `json:"conditions"`
	BusinessImpactIfImplemented   *string     `json:"business_impact_if_implemented"`
	RegulatoryNotificationRequired *bool      `json:"regulatory_notification_required"`
	Tags                          []string    `json:"tags"`
	Metadata                      json.RawMessage `json:"metadata"`
}

// ExceptionFilter holds filtering parameters for listing exceptions.
type ExceptionFilter struct {
	Status        string     `json:"status"`
	ExceptionType string     `json:"exception_type"`
	Priority      string     `json:"priority"`
	ScopeType     string     `json:"scope_type"`
	RiskLevel     string     `json:"risk_level"`
	Search        string     `json:"search"`
	Tags          []string   `json:"tags"`
	Page          int        `json:"page"`
	PageSize      int        `json:"page_size"`
}

// ExceptionListResult wraps the paginated list of exceptions.
type ExceptionListResult struct {
	Exceptions []ComplianceException `json:"exceptions"`
	Total      int64                 `json:"total"`
}

// ExceptionReviewRequest is the request body for submitting an exception review.
type ExceptionReviewRequest struct {
	ReviewType            string  `json:"review_type"`
	Outcome               string  `json:"outcome"`
	RiskLevelAtReview     string  `json:"risk_level_at_review"`
	CompensatingEffective *bool   `json:"compensating_effective"`
	Findings              string  `json:"findings"`
	Recommendations       string  `json:"recommendations"`
	NextReviewDate        string  `json:"next_review_date"`
}

// ExceptionDashboard provides at-a-glance metrics for exception management.
type ExceptionDashboard struct {
	ActiveCount           int                `json:"active_count"`
	DraftCount            int                `json:"draft_count"`
	PendingApprovalCount  int                `json:"pending_approval_count"`
	RejectedCount         int                `json:"rejected_count"`
	ExpiredCount          int                `json:"expired_count"`
	ByRiskLevel           map[string]int     `json:"by_risk_level"`
	ByPriority            map[string]int     `json:"by_priority"`
	ByType                map[string]int     `json:"by_type"`
	Expiring30Days        int                `json:"expiring_30_days"`
	Expiring60Days        int                `json:"expiring_60_days"`
	Expiring90Days        int                `json:"expiring_90_days"`
	OverdueReviews        int                `json:"overdue_reviews"`
	AverageAgeDays        float64            `json:"average_age_days"`
	TopExceptedFrameworks []FrameworkExCount `json:"top_excepted_frameworks"`
	RecentExceptions      []ComplianceException `json:"recent_exceptions"`
}

// FrameworkExCount holds a framework code and exception count.
type FrameworkExCount struct {
	FrameworkCode string `json:"framework_code"`
	Count         int    `json:"count"`
}

// ComplianceImpact shows the effect an exception has on compliance posture.
type ComplianceImpact struct {
	ExceptionID              uuid.UUID         `json:"exception_id"`
	ExceptionRef             string            `json:"exception_ref"`
	AffectedControlCount     int               `json:"affected_control_count"`
	AffectedFrameworks       []string          `json:"affected_frameworks"`
	ComplianceScoreImpact    float64           `json:"compliance_score_impact"`
	RiskExposureIncrease     string            `json:"risk_exposure_increase"`
	CompensatingCoverage     float64           `json:"compensating_coverage"`
	NetRiskDelta             string            `json:"net_risk_delta"`
	AffectedControls         []AffectedControl `json:"affected_controls"`
	Recommendations          []string          `json:"recommendations"`
}

// AffectedControl is a control impacted by the exception.
type AffectedControl struct {
	ControlCode string `json:"control_code"`
	ControlName string `json:"control_name"`
	Framework   string `json:"framework"`
	Impact      string `json:"impact"`
}

// RenewExceptionRequest is the request body for renewing an exception.
type RenewExceptionRequest struct {
	NewExpiry     string `json:"new_expiry"`
	Justification string `json:"justification"`
}

// ApproveExceptionRequest is the request body for approving an exception.
type ApproveExceptionRequest struct {
	Comments string `json:"comments"`
}

// RejectExceptionRequest is the request body for rejecting an exception.
type RejectExceptionRequest struct {
	Reason string `json:"reason"`
}

// RevokeExceptionRequest is the request body for revoking an exception.
type RevokeExceptionRequest struct {
	Reason string `json:"reason"`
}

// ============================================================
// SERVICE
// ============================================================

// ExceptionService provides business logic for exception management,
// including creation, approval workflows, reviews, and compliance impact analysis.
type ExceptionService struct {
	pool *pgxpool.Pool
}

// NewExceptionService creates a new ExceptionService.
func NewExceptionService(pool *pgxpool.Pool) *ExceptionService {
	return &ExceptionService{pool: pool}
}

// setOrgRLS sets the RLS context for the current transaction.
func (s *ExceptionService) setOrgRLS(ctx context.Context, tx pgx.Tx, orgID uuid.UUID) error {
	_, err := tx.Exec(ctx, "SET LOCAL app.current_org = '"+orgID.String()+"'")
	return err
}

// ============================================================
// CREATE
// ============================================================

// CreateException creates a new compliance exception in draft status.
func (s *ExceptionService) CreateException(ctx context.Context, orgID uuid.UUID, req CreateExceptionRequest) (*ComplianceException, error) {
	if req.Title == "" {
		return nil, fmt.Errorf("title is required")
	}
	if req.Description == "" {
		return nil, fmt.Errorf("description is required")
	}
	if req.RiskJustification == "" {
		return nil, fmt.Errorf("risk_justification is required")
	}
	if req.EffectiveDate == "" {
		return nil, fmt.Errorf("effective_date is required")
	}

	effectiveDate, err := time.Parse("2006-01-02", req.EffectiveDate)
	if err != nil {
		return nil, fmt.Errorf("invalid effective_date format: use YYYY-MM-DD")
	}

	var expiryDate *time.Time
	if req.ExpiryDate != "" {
		t, err := time.Parse("2006-01-02", req.ExpiryDate)
		if err != nil {
			return nil, fmt.Errorf("invalid expiry_date format: use YYYY-MM-DD")
		}
		if t.Before(effectiveDate) {
			return nil, fmt.Errorf("expiry_date must be after effective_date")
		}
		expiryDate = &t
	}

	// Defaults
	if req.ExceptionType == "" {
		req.ExceptionType = "temporary"
	}
	if req.Priority == "" {
		req.Priority = "medium"
	}
	if req.ScopeType == "" {
		req.ScopeType = "control_implementation"
	}
	if req.ResidualRiskLevel == "" {
		req.ResidualRiskLevel = "medium"
	}
	if req.CompensatingEffectiveness == "" {
		req.CompensatingEffectiveness = "not_assessed"
	}
	if req.ReviewFrequencyMonths < 1 {
		req.ReviewFrequencyMonths = 12
	}
	if req.ControlImplementationIDs == nil {
		req.ControlImplementationIDs = []uuid.UUID{}
	}
	if req.FrameworkControlCodes == nil {
		req.FrameworkControlCodes = []string{}
	}
	if req.CompensatingControlIDs == nil {
		req.CompensatingControlIDs = []uuid.UUID{}
	}
	if req.Tags == nil {
		req.Tags = []string{}
	}
	if req.Metadata == nil {
		req.Metadata = json.RawMessage(`{}`)
	}

	// Extract requester from context
	requesterID := contextUserID(ctx)
	if requesterID == nil {
		return nil, fmt.Errorf("user ID not found in context")
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.setOrgRLS(ctx, tx, orgID); err != nil {
		return nil, fmt.Errorf("set RLS: %w", err)
	}

	var exc ComplianceException
	err = tx.QueryRow(ctx, `
		INSERT INTO compliance_exceptions (
			organization_id, title, description, exception_type, status, priority,
			scope_type, control_implementation_ids, framework_control_codes,
			policy_id, scope_description,
			risk_justification, residual_risk_description, residual_risk_level,
			has_compensating_controls, compensating_controls_description,
			compensating_control_ids, compensating_effectiveness,
			requested_by, effective_date, expiry_date,
			review_frequency_months, conditions, business_impact_if_implemented,
			regulatory_notification_required, tags, metadata
		) VALUES (
			$1, $2, $3, $4, 'draft', $5,
			$6, $7, $8,
			$9, $10,
			$11, $12, $13,
			$14, $15,
			$16, $17,
			$18, $19, $20,
			$21, $22, $23,
			$24, $25, $26
		)
		RETURNING id, organization_id, exception_ref, title, description,
			exception_type::text, status::text, priority,
			scope_type::text, control_implementation_ids, framework_control_codes,
			policy_id, COALESCE(scope_description, ''),
			risk_justification, COALESCE(residual_risk_description, ''), residual_risk_level,
			risk_assessment_id, risk_accepted_by, risk_accepted_at,
			has_compensating_controls, COALESCE(compensating_controls_description, ''),
			compensating_control_ids, compensating_effectiveness::text,
			requested_by, requested_at, approved_by, approved_at,
			COALESCE(approval_comments, ''), COALESCE(rejection_reason, ''),
			workflow_instance_id,
			effective_date, expiry_date, review_frequency_months,
			next_review_date, last_review_date, last_reviewed_by, renewal_count,
			COALESCE(conditions, ''), COALESCE(business_impact_if_implemented, ''),
			regulatory_notification_required, regulator_notified_at,
			COALESCE(audit_evidence_path, ''),
			tags, metadata, created_at, updated_at, deleted_at`,
		orgID, req.Title, req.Description, req.ExceptionType, req.Priority,
		req.ScopeType, req.ControlImplementationIDs, req.FrameworkControlCodes,
		req.PolicyID, req.ScopeDescription,
		req.RiskJustification, req.ResidualRiskDescription, req.ResidualRiskLevel,
		req.HasCompensatingControls, req.CompensatingControlsDescription,
		req.CompensatingControlIDs, req.CompensatingEffectiveness,
		*requesterID, effectiveDate, expiryDate,
		req.ReviewFrequencyMonths, req.Conditions, req.BusinessImpactIfImplemented,
		req.RegulatoryNotificationRequired, req.Tags, req.Metadata,
	).Scan(
		&exc.ID, &exc.OrganizationID, &exc.ExceptionRef, &exc.Title, &exc.Description,
		&exc.ExceptionType, &exc.Status, &exc.Priority,
		&exc.ScopeType, &exc.ControlImplementationIDs, &exc.FrameworkControlCodes,
		&exc.PolicyID, &exc.ScopeDescription,
		&exc.RiskJustification, &exc.ResidualRiskDescription, &exc.ResidualRiskLevel,
		&exc.RiskAssessmentID, &exc.RiskAcceptedBy, &exc.RiskAcceptedAt,
		&exc.HasCompensatingControls, &exc.CompensatingControlsDescription,
		&exc.CompensatingControlIDs, &exc.CompensatingEffectiveness,
		&exc.RequestedBy, &exc.RequestedAt, &exc.ApprovedBy, &exc.ApprovedAt,
		&exc.ApprovalComments, &exc.RejectionReason,
		&exc.WorkflowInstanceID,
		&exc.EffectiveDate, &exc.ExpiryDate, &exc.ReviewFrequencyMonths,
		&exc.NextReviewDate, &exc.LastReviewDate, &exc.LastReviewedBy, &exc.RenewalCount,
		&exc.Conditions, &exc.BusinessImpactIfImplemented,
		&exc.RegulatoryNotificationRequired, &exc.RegulatorNotifiedAt,
		&exc.AuditEvidencePath,
		&exc.Tags, &exc.Metadata, &exc.CreatedAt, &exc.UpdatedAt, &exc.DeletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert exception: %w", err)
	}

	// Audit trail
	_, err = tx.Exec(ctx, `
		INSERT INTO exception_audit_trail
			(organization_id, exception_id, action, actor_id, new_status, details)
		VALUES ($1, $2, 'created', $3, 'draft', $4)`,
		orgID, exc.ID, *requesterID, fmt.Sprintf("Exception %s created: %s", exc.ExceptionRef, exc.Title))
	if err != nil {
		log.Warn().Err(err).Msg("Failed to write exception audit trail")
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	log.Info().
		Str("exception_id", exc.ID.String()).
		Str("ref", exc.ExceptionRef).
		Msg("Compliance exception created")

	return &exc, nil
}

// ============================================================
// GET
// ============================================================

// GetException retrieves a single exception by ID.
func (s *ExceptionService) GetException(ctx context.Context, orgID, exceptionID uuid.UUID) (*ComplianceException, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.setOrgRLS(ctx, tx, orgID); err != nil {
		return nil, fmt.Errorf("set RLS: %w", err)
	}

	exc, err := s.scanException(ctx, tx, exceptionID)
	if err != nil {
		return nil, err
	}

	tx.Commit(ctx)
	return exc, nil
}

// scanException scans a single exception by ID within an existing transaction.
func (s *ExceptionService) scanException(ctx context.Context, tx pgx.Tx, exceptionID uuid.UUID) (*ComplianceException, error) {
	var exc ComplianceException
	err := tx.QueryRow(ctx, `
		SELECT id, organization_id, exception_ref, title, description,
			exception_type::text, status::text, priority,
			scope_type::text, control_implementation_ids, framework_control_codes,
			policy_id, COALESCE(scope_description, ''),
			risk_justification, COALESCE(residual_risk_description, ''), residual_risk_level,
			risk_assessment_id, risk_accepted_by, risk_accepted_at,
			has_compensating_controls, COALESCE(compensating_controls_description, ''),
			compensating_control_ids, compensating_effectiveness::text,
			requested_by, requested_at, approved_by, approved_at,
			COALESCE(approval_comments, ''), COALESCE(rejection_reason, ''),
			workflow_instance_id,
			effective_date, expiry_date, review_frequency_months,
			next_review_date, last_review_date, last_reviewed_by, renewal_count,
			COALESCE(conditions, ''), COALESCE(business_impact_if_implemented, ''),
			regulatory_notification_required, regulator_notified_at,
			COALESCE(audit_evidence_path, ''),
			tags, metadata, created_at, updated_at, deleted_at
		FROM compliance_exceptions
		WHERE id = $1 AND deleted_at IS NULL`, exceptionID).Scan(
		&exc.ID, &exc.OrganizationID, &exc.ExceptionRef, &exc.Title, &exc.Description,
		&exc.ExceptionType, &exc.Status, &exc.Priority,
		&exc.ScopeType, &exc.ControlImplementationIDs, &exc.FrameworkControlCodes,
		&exc.PolicyID, &exc.ScopeDescription,
		&exc.RiskJustification, &exc.ResidualRiskDescription, &exc.ResidualRiskLevel,
		&exc.RiskAssessmentID, &exc.RiskAcceptedBy, &exc.RiskAcceptedAt,
		&exc.HasCompensatingControls, &exc.CompensatingControlsDescription,
		&exc.CompensatingControlIDs, &exc.CompensatingEffectiveness,
		&exc.RequestedBy, &exc.RequestedAt, &exc.ApprovedBy, &exc.ApprovedAt,
		&exc.ApprovalComments, &exc.RejectionReason,
		&exc.WorkflowInstanceID,
		&exc.EffectiveDate, &exc.ExpiryDate, &exc.ReviewFrequencyMonths,
		&exc.NextReviewDate, &exc.LastReviewDate, &exc.LastReviewedBy, &exc.RenewalCount,
		&exc.Conditions, &exc.BusinessImpactIfImplemented,
		&exc.RegulatoryNotificationRequired, &exc.RegulatorNotifiedAt,
		&exc.AuditEvidencePath,
		&exc.Tags, &exc.Metadata, &exc.CreatedAt, &exc.UpdatedAt, &exc.DeletedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("exception not found")
		}
		return nil, fmt.Errorf("get exception: %w", err)
	}
	return &exc, nil
}

// ============================================================
// LIST
// ============================================================

// ListExceptions returns a filtered, paginated list of compliance exceptions.
func (s *ExceptionService) ListExceptions(ctx context.Context, orgID uuid.UUID, filter ExceptionFilter) (*ExceptionListResult, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 20
	}
	offset := (filter.Page - 1) * filter.PageSize

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.setOrgRLS(ctx, tx, orgID); err != nil {
		return nil, fmt.Errorf("set RLS: %w", err)
	}

	var conditions []string
	var args []interface{}
	argIdx := 1

	conditions = append(conditions, "deleted_at IS NULL")

	if filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d::exception_status", argIdx))
		args = append(args, filter.Status)
		argIdx++
	}
	if filter.ExceptionType != "" {
		conditions = append(conditions, fmt.Sprintf("exception_type = $%d::exception_type", argIdx))
		args = append(args, filter.ExceptionType)
		argIdx++
	}
	if filter.Priority != "" {
		conditions = append(conditions, fmt.Sprintf("priority = $%d", argIdx))
		args = append(args, filter.Priority)
		argIdx++
	}
	if filter.ScopeType != "" {
		conditions = append(conditions, fmt.Sprintf("scope_type = $%d::exception_scope_type", argIdx))
		args = append(args, filter.ScopeType)
		argIdx++
	}
	if filter.RiskLevel != "" {
		conditions = append(conditions, fmt.Sprintf("residual_risk_level = $%d", argIdx))
		args = append(args, filter.RiskLevel)
		argIdx++
	}
	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf(
			"(title ILIKE '%%' || $%d || '%%' OR description ILIKE '%%' || $%d || '%%' OR exception_ref ILIKE '%%' || $%d || '%%')",
			argIdx, argIdx, argIdx))
		args = append(args, filter.Search)
		argIdx++
	}

	where := "WHERE " + strings.Join(conditions, " AND ")

	// Count
	countSQL := fmt.Sprintf(`SELECT COUNT(*) FROM compliance_exceptions %s`, where)
	var total int64
	if err := tx.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count exceptions: %w", err)
	}

	// Data
	dataSQL := fmt.Sprintf(`
		SELECT id, organization_id, exception_ref, title, description,
			exception_type::text, status::text, priority,
			scope_type::text, control_implementation_ids, framework_control_codes,
			policy_id, COALESCE(scope_description, ''),
			risk_justification, COALESCE(residual_risk_description, ''), residual_risk_level,
			risk_assessment_id, risk_accepted_by, risk_accepted_at,
			has_compensating_controls, COALESCE(compensating_controls_description, ''),
			compensating_control_ids, compensating_effectiveness::text,
			requested_by, requested_at, approved_by, approved_at,
			COALESCE(approval_comments, ''), COALESCE(rejection_reason, ''),
			workflow_instance_id,
			effective_date, expiry_date, review_frequency_months,
			next_review_date, last_review_date, last_reviewed_by, renewal_count,
			COALESCE(conditions, ''), COALESCE(business_impact_if_implemented, ''),
			regulatory_notification_required, regulator_notified_at,
			COALESCE(audit_evidence_path, ''),
			tags, metadata, created_at, updated_at, deleted_at
		FROM compliance_exceptions
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)

	args = append(args, filter.PageSize, offset)

	rows, err := tx.Query(ctx, dataSQL, args...)
	if err != nil {
		return nil, fmt.Errorf("query exceptions: %w", err)
	}
	defer rows.Close()

	var exceptions []ComplianceException
	for rows.Next() {
		var exc ComplianceException
		if err := rows.Scan(
			&exc.ID, &exc.OrganizationID, &exc.ExceptionRef, &exc.Title, &exc.Description,
			&exc.ExceptionType, &exc.Status, &exc.Priority,
			&exc.ScopeType, &exc.ControlImplementationIDs, &exc.FrameworkControlCodes,
			&exc.PolicyID, &exc.ScopeDescription,
			&exc.RiskJustification, &exc.ResidualRiskDescription, &exc.ResidualRiskLevel,
			&exc.RiskAssessmentID, &exc.RiskAcceptedBy, &exc.RiskAcceptedAt,
			&exc.HasCompensatingControls, &exc.CompensatingControlsDescription,
			&exc.CompensatingControlIDs, &exc.CompensatingEffectiveness,
			&exc.RequestedBy, &exc.RequestedAt, &exc.ApprovedBy, &exc.ApprovedAt,
			&exc.ApprovalComments, &exc.RejectionReason,
			&exc.WorkflowInstanceID,
			&exc.EffectiveDate, &exc.ExpiryDate, &exc.ReviewFrequencyMonths,
			&exc.NextReviewDate, &exc.LastReviewDate, &exc.LastReviewedBy, &exc.RenewalCount,
			&exc.Conditions, &exc.BusinessImpactIfImplemented,
			&exc.RegulatoryNotificationRequired, &exc.RegulatorNotifiedAt,
			&exc.AuditEvidencePath,
			&exc.Tags, &exc.Metadata, &exc.CreatedAt, &exc.UpdatedAt, &exc.DeletedAt,
		); err != nil {
			return nil, fmt.Errorf("scan exception: %w", err)
		}
		exceptions = append(exceptions, exc)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate exceptions: %w", err)
	}

	if exceptions == nil {
		exceptions = []ComplianceException{}
	}

	tx.Commit(ctx)
	return &ExceptionListResult{Exceptions: exceptions, Total: total}, nil
}

// ============================================================
// UPDATE
// ============================================================

// UpdateException updates an exception. Only allowed when status is 'draft'.
func (s *ExceptionService) UpdateException(ctx context.Context, orgID, exceptionID uuid.UUID, req UpdateExceptionRequest) (*ComplianceException, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.setOrgRLS(ctx, tx, orgID); err != nil {
		return nil, fmt.Errorf("set RLS: %w", err)
	}

	// Verify status is draft
	var currentStatus string
	err = tx.QueryRow(ctx,
		`SELECT status::text FROM compliance_exceptions WHERE id = $1 AND deleted_at IS NULL`,
		exceptionID).Scan(&currentStatus)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("exception not found")
		}
		return nil, fmt.Errorf("check status: %w", err)
	}
	if currentStatus != "draft" {
		return nil, fmt.Errorf("only draft exceptions can be updated; current status: %s", currentStatus)
	}

	// Build dynamic update
	var setClauses []string
	var args []interface{}
	argIdx := 1

	if req.Title != nil {
		setClauses = append(setClauses, fmt.Sprintf("title = $%d", argIdx))
		args = append(args, *req.Title)
		argIdx++
	}
	if req.Description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", argIdx))
		args = append(args, *req.Description)
		argIdx++
	}
	if req.ExceptionType != nil {
		setClauses = append(setClauses, fmt.Sprintf("exception_type = $%d::exception_type", argIdx))
		args = append(args, *req.ExceptionType)
		argIdx++
	}
	if req.Priority != nil {
		setClauses = append(setClauses, fmt.Sprintf("priority = $%d", argIdx))
		args = append(args, *req.Priority)
		argIdx++
	}
	if req.ScopeType != nil {
		setClauses = append(setClauses, fmt.Sprintf("scope_type = $%d::exception_scope_type", argIdx))
		args = append(args, *req.ScopeType)
		argIdx++
	}
	if req.ControlImplementationIDs != nil {
		setClauses = append(setClauses, fmt.Sprintf("control_implementation_ids = $%d", argIdx))
		args = append(args, req.ControlImplementationIDs)
		argIdx++
	}
	if req.FrameworkControlCodes != nil {
		setClauses = append(setClauses, fmt.Sprintf("framework_control_codes = $%d", argIdx))
		args = append(args, req.FrameworkControlCodes)
		argIdx++
	}
	if req.PolicyID != nil {
		setClauses = append(setClauses, fmt.Sprintf("policy_id = $%d", argIdx))
		args = append(args, *req.PolicyID)
		argIdx++
	}
	if req.ScopeDescription != nil {
		setClauses = append(setClauses, fmt.Sprintf("scope_description = $%d", argIdx))
		args = append(args, *req.ScopeDescription)
		argIdx++
	}
	if req.RiskJustification != nil {
		setClauses = append(setClauses, fmt.Sprintf("risk_justification = $%d", argIdx))
		args = append(args, *req.RiskJustification)
		argIdx++
	}
	if req.ResidualRiskDescription != nil {
		setClauses = append(setClauses, fmt.Sprintf("residual_risk_description = $%d", argIdx))
		args = append(args, *req.ResidualRiskDescription)
		argIdx++
	}
	if req.ResidualRiskLevel != nil {
		setClauses = append(setClauses, fmt.Sprintf("residual_risk_level = $%d", argIdx))
		args = append(args, *req.ResidualRiskLevel)
		argIdx++
	}
	if req.HasCompensatingControls != nil {
		setClauses = append(setClauses, fmt.Sprintf("has_compensating_controls = $%d", argIdx))
		args = append(args, *req.HasCompensatingControls)
		argIdx++
	}
	if req.CompensatingControlsDescription != nil {
		setClauses = append(setClauses, fmt.Sprintf("compensating_controls_description = $%d", argIdx))
		args = append(args, *req.CompensatingControlsDescription)
		argIdx++
	}
	if req.CompensatingControlIDs != nil {
		setClauses = append(setClauses, fmt.Sprintf("compensating_control_ids = $%d", argIdx))
		args = append(args, req.CompensatingControlIDs)
		argIdx++
	}
	if req.CompensatingEffectiveness != nil {
		setClauses = append(setClauses, fmt.Sprintf("compensating_effectiveness = $%d::compensating_effectiveness", argIdx))
		args = append(args, *req.CompensatingEffectiveness)
		argIdx++
	}
	if req.EffectiveDate != nil {
		setClauses = append(setClauses, fmt.Sprintf("effective_date = $%d", argIdx))
		t, err := time.Parse("2006-01-02", *req.EffectiveDate)
		if err != nil {
			return nil, fmt.Errorf("invalid effective_date format")
		}
		args = append(args, t)
		argIdx++
	}
	if req.ExpiryDate != nil {
		setClauses = append(setClauses, fmt.Sprintf("expiry_date = $%d", argIdx))
		t, err := time.Parse("2006-01-02", *req.ExpiryDate)
		if err != nil {
			return nil, fmt.Errorf("invalid expiry_date format")
		}
		args = append(args, t)
		argIdx++
	}
	if req.ReviewFrequencyMonths != nil {
		setClauses = append(setClauses, fmt.Sprintf("review_frequency_months = $%d", argIdx))
		args = append(args, *req.ReviewFrequencyMonths)
		argIdx++
	}
	if req.Conditions != nil {
		setClauses = append(setClauses, fmt.Sprintf("conditions = $%d", argIdx))
		args = append(args, *req.Conditions)
		argIdx++
	}
	if req.BusinessImpactIfImplemented != nil {
		setClauses = append(setClauses, fmt.Sprintf("business_impact_if_implemented = $%d", argIdx))
		args = append(args, *req.BusinessImpactIfImplemented)
		argIdx++
	}
	if req.RegulatoryNotificationRequired != nil {
		setClauses = append(setClauses, fmt.Sprintf("regulatory_notification_required = $%d", argIdx))
		args = append(args, *req.RegulatoryNotificationRequired)
		argIdx++
	}
	if req.Tags != nil {
		setClauses = append(setClauses, fmt.Sprintf("tags = $%d", argIdx))
		args = append(args, req.Tags)
		argIdx++
	}
	if req.Metadata != nil {
		setClauses = append(setClauses, fmt.Sprintf("metadata = $%d", argIdx))
		args = append(args, req.Metadata)
		argIdx++
	}

	if len(setClauses) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}

	updateSQL := fmt.Sprintf(
		`UPDATE compliance_exceptions SET %s WHERE id = $%d AND deleted_at IS NULL`,
		strings.Join(setClauses, ", "), argIdx)
	args = append(args, exceptionID)

	tag, err := tx.Exec(ctx, updateSQL, args...)
	if err != nil {
		return nil, fmt.Errorf("update exception: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil, fmt.Errorf("exception not found or update failed")
	}

	exc, err := s.scanException(ctx, tx, exceptionID)
	if err != nil {
		return nil, err
	}

	// Audit trail
	actorID := contextUserID(ctx)
	actorUUID := uuid.Nil
	if actorID != nil {
		actorUUID = *actorID
	}
	_, _ = tx.Exec(ctx, `
		INSERT INTO exception_audit_trail
			(organization_id, exception_id, action, actor_id, details)
		VALUES ($1, $2, 'updated', $3, 'Exception updated in draft status')`,
		orgID, exceptionID, actorUUID)

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return exc, nil
}

// ============================================================
// SUBMIT FOR APPROVAL
// ============================================================

// SubmitForApproval transitions an exception from draft to pending_risk_assessment.
func (s *ExceptionService) SubmitForApproval(ctx context.Context, orgID, exceptionID uuid.UUID) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.setOrgRLS(ctx, tx, orgID); err != nil {
		return fmt.Errorf("set RLS: %w", err)
	}

	// Get current exception
	exc, err := s.scanException(ctx, tx, exceptionID)
	if err != nil {
		return err
	}

	if exc.Status != "draft" {
		return fmt.Errorf("exception must be in draft status to submit; current: %s", exc.Status)
	}

	// Validate required fields
	if exc.Title == "" || exc.Description == "" || exc.RiskJustification == "" {
		return fmt.Errorf("title, description, and risk_justification are required before submission")
	}
	if len(exc.ControlImplementationIDs) == 0 && len(exc.FrameworkControlCodes) == 0 && exc.PolicyID == nil {
		return fmt.Errorf("at least one control, framework code, or policy must be specified in scope")
	}

	tag, err := tx.Exec(ctx, `
		UPDATE compliance_exceptions
		SET status = 'pending_risk_assessment'
		WHERE id = $1 AND deleted_at IS NULL`,
		exceptionID)
	if err != nil {
		return fmt.Errorf("update status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("exception not found")
	}

	// Audit trail
	actorID := contextUserID(ctx)
	actorUUID := uuid.Nil
	if actorID != nil {
		actorUUID = *actorID
	}
	_, _ = tx.Exec(ctx, `
		INSERT INTO exception_audit_trail
			(organization_id, exception_id, action, actor_id, old_status, new_status, details)
		VALUES ($1, $2, 'submitted_for_approval', $3, 'draft', 'pending_risk_assessment', 'Exception submitted for risk assessment and approval')`,
		orgID, exceptionID, actorUUID)

	return tx.Commit(ctx)
}

// ============================================================
// APPROVE
// ============================================================

// ApproveException approves an exception and sets validity dates.
func (s *ExceptionService) ApproveException(ctx context.Context, orgID, exceptionID, approverID uuid.UUID, comments string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.setOrgRLS(ctx, tx, orgID); err != nil {
		return fmt.Errorf("set RLS: %w", err)
	}

	exc, err := s.scanException(ctx, tx, exceptionID)
	if err != nil {
		return err
	}

	allowedStatuses := map[string]bool{
		"pending_risk_assessment": true,
		"pending_approval":       true,
	}
	if !allowedStatuses[exc.Status] {
		return fmt.Errorf("exception must be pending approval; current: %s", exc.Status)
	}

	now := time.Now()
	nextReview := now.AddDate(0, exc.ReviewFrequencyMonths, 0)

	// Calculate expiry for temporary exceptions without one
	var expiryDate *time.Time
	if exc.ExpiryDate != nil {
		expiryDate = exc.ExpiryDate
	} else if exc.ExceptionType == "temporary" {
		expiry := now.AddDate(1, 0, 0) // Default 1 year for temporary
		expiryDate = &expiry
	}

	tag, err := tx.Exec(ctx, `
		UPDATE compliance_exceptions
		SET status = 'approved',
			approved_by = $1,
			approved_at = $2,
			approval_comments = $3,
			risk_accepted_by = $1,
			risk_accepted_at = $2,
			next_review_date = $4,
			expiry_date = COALESCE($5, expiry_date)
		WHERE id = $6 AND deleted_at IS NULL`,
		approverID, now, comments, nextReview, expiryDate, exceptionID)
	if err != nil {
		return fmt.Errorf("approve exception: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("exception not found")
	}

	// Audit trail
	_, _ = tx.Exec(ctx, `
		INSERT INTO exception_audit_trail
			(organization_id, exception_id, action, actor_id, old_status, new_status, details)
		VALUES ($1, $2, 'approved', $3, $4, 'approved', $5)`,
		orgID, exceptionID, approverID, exc.Status,
		fmt.Sprintf("Approved with comments: %s", comments))

	log.Info().
		Str("exception_id", exceptionID.String()).
		Str("approver_id", approverID.String()).
		Msg("Exception approved")

	return tx.Commit(ctx)
}

// ============================================================
// REJECT
// ============================================================

// RejectException rejects an exception with a reason.
func (s *ExceptionService) RejectException(ctx context.Context, orgID, exceptionID, rejectorID uuid.UUID, reason string) error {
	if reason == "" {
		return fmt.Errorf("rejection reason is required")
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.setOrgRLS(ctx, tx, orgID); err != nil {
		return fmt.Errorf("set RLS: %w", err)
	}

	var currentStatus string
	err = tx.QueryRow(ctx,
		`SELECT status::text FROM compliance_exceptions WHERE id = $1 AND deleted_at IS NULL`,
		exceptionID).Scan(&currentStatus)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("exception not found")
		}
		return fmt.Errorf("check status: %w", err)
	}

	allowedStatuses := map[string]bool{
		"pending_risk_assessment": true,
		"pending_approval":       true,
	}
	if !allowedStatuses[currentStatus] {
		return fmt.Errorf("exception must be pending to reject; current: %s", currentStatus)
	}

	tag, err := tx.Exec(ctx, `
		UPDATE compliance_exceptions
		SET status = 'rejected', rejection_reason = $1
		WHERE id = $2 AND deleted_at IS NULL`,
		reason, exceptionID)
	if err != nil {
		return fmt.Errorf("reject exception: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("exception not found")
	}

	_, _ = tx.Exec(ctx, `
		INSERT INTO exception_audit_trail
			(organization_id, exception_id, action, actor_id, old_status, new_status, details)
		VALUES ($1, $2, 'rejected', $3, $4, 'rejected', $5)`,
		orgID, exceptionID, rejectorID, currentStatus,
		fmt.Sprintf("Rejected: %s", reason))

	return tx.Commit(ctx)
}

// ============================================================
// REVOKE
// ============================================================

// RevokeException revokes an approved exception.
func (s *ExceptionService) RevokeException(ctx context.Context, orgID, exceptionID, revokerID uuid.UUID, reason string) error {
	if reason == "" {
		return fmt.Errorf("revocation reason is required")
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.setOrgRLS(ctx, tx, orgID); err != nil {
		return fmt.Errorf("set RLS: %w", err)
	}

	var currentStatus string
	err = tx.QueryRow(ctx,
		`SELECT status::text FROM compliance_exceptions WHERE id = $1 AND deleted_at IS NULL`,
		exceptionID).Scan(&currentStatus)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("exception not found")
		}
		return fmt.Errorf("check status: %w", err)
	}

	if currentStatus != "approved" {
		return fmt.Errorf("only approved exceptions can be revoked; current: %s", currentStatus)
	}

	tag, err := tx.Exec(ctx, `
		UPDATE compliance_exceptions
		SET status = 'revoked'
		WHERE id = $1 AND deleted_at IS NULL`,
		exceptionID)
	if err != nil {
		return fmt.Errorf("revoke exception: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("exception not found")
	}

	_, _ = tx.Exec(ctx, `
		INSERT INTO exception_audit_trail
			(organization_id, exception_id, action, actor_id, old_status, new_status, details)
		VALUES ($1, $2, 'revoked', $3, 'approved', 'revoked', $4)`,
		orgID, exceptionID, revokerID, fmt.Sprintf("Revoked: %s", reason))

	return tx.Commit(ctx)
}

// ============================================================
// RENEW
// ============================================================

// RenewException renews a temporary exception. Maximum 2 renewals allowed.
func (s *ExceptionService) RenewException(ctx context.Context, orgID, exceptionID uuid.UUID, newExpiry time.Time, justification string) error {
	if justification == "" {
		return fmt.Errorf("renewal justification is required")
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.setOrgRLS(ctx, tx, orgID); err != nil {
		return fmt.Errorf("set RLS: %w", err)
	}

	exc, err := s.scanException(ctx, tx, exceptionID)
	if err != nil {
		return err
	}

	allowedStatuses := map[string]bool{
		"approved":        true,
		"renewal_pending": true,
		"expired":         true,
	}
	if !allowedStatuses[exc.Status] {
		return fmt.Errorf("exception cannot be renewed in status: %s", exc.Status)
	}

	if exc.ExceptionType == "temporary" && exc.RenewalCount >= 2 {
		return fmt.Errorf("maximum renewal count (2) reached for temporary exceptions")
	}

	if newExpiry.Before(time.Now()) {
		return fmt.Errorf("new expiry date must be in the future")
	}

	nextReview := time.Now().AddDate(0, exc.ReviewFrequencyMonths, 0)

	tag, err := tx.Exec(ctx, `
		UPDATE compliance_exceptions
		SET status = 'approved',
			expiry_date = $1,
			renewal_count = renewal_count + 1,
			next_review_date = $2
		WHERE id = $3 AND deleted_at IS NULL`,
		newExpiry, nextReview, exceptionID)
	if err != nil {
		return fmt.Errorf("renew exception: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("exception not found")
	}

	actorID := contextUserID(ctx)
	actorUUID := uuid.Nil
	if actorID != nil {
		actorUUID = *actorID
	}

	_, _ = tx.Exec(ctx, `
		INSERT INTO exception_audit_trail
			(organization_id, exception_id, action, actor_id, old_status, new_status, details)
		VALUES ($1, $2, 'renewed', $3, $4, 'approved', $5)`,
		orgID, exceptionID, actorUUID, exc.Status,
		fmt.Sprintf("Renewed until %s. Justification: %s. Renewal #%d",
			newExpiry.Format("2006-01-02"), justification, exc.RenewalCount+1))

	log.Info().
		Str("exception_id", exceptionID.String()).
		Int("renewal_count", exc.RenewalCount+1).
		Msg("Exception renewed")

	return tx.Commit(ctx)
}

// ============================================================
// REVIEW
// ============================================================

// ReviewException records a review of an exception.
func (s *ExceptionService) ReviewException(ctx context.Context, orgID, exceptionID uuid.UUID, req ExceptionReviewRequest) error {
	if req.Outcome == "" {
		return fmt.Errorf("review outcome is required")
	}
	if req.ReviewType == "" {
		req.ReviewType = "periodic"
	}

	reviewerID := contextUserID(ctx)
	if reviewerID == nil {
		return fmt.Errorf("user ID not found in context")
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.setOrgRLS(ctx, tx, orgID); err != nil {
		return fmt.Errorf("set RLS: %w", err)
	}

	// Verify exception exists and is active
	var currentStatus string
	err = tx.QueryRow(ctx,
		`SELECT status::text FROM compliance_exceptions WHERE id = $1 AND deleted_at IS NULL`,
		exceptionID).Scan(&currentStatus)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("exception not found")
		}
		return fmt.Errorf("check status: %w", err)
	}

	// Parse next review date if provided
	var nextReview *time.Time
	if req.NextReviewDate != "" {
		t, err := time.Parse("2006-01-02", req.NextReviewDate)
		if err != nil {
			return fmt.Errorf("invalid next_review_date format")
		}
		nextReview = &t
	}

	// Insert review record
	_, err = tx.Exec(ctx, `
		INSERT INTO exception_reviews
			(organization_id, exception_id, review_type, reviewer_id,
			 outcome, risk_level_at_review, compensating_effective,
			 findings, recommendations, next_review_date)
		VALUES ($1, $2, $3::exception_review_type, $4,
			$5::exception_review_outcome, $6, $7,
			$8, $9, $10)`,
		orgID, exceptionID, req.ReviewType, *reviewerID,
		req.Outcome, req.RiskLevelAtReview, req.CompensatingEffective,
		req.Findings, req.Recommendations, nextReview)
	if err != nil {
		return fmt.Errorf("insert review: %w", err)
	}

	// Update exception with review info
	now := time.Now()
	updateSQL := `
		UPDATE compliance_exceptions
		SET last_review_date = $1, last_reviewed_by = $2`
	updateArgs := []interface{}{now, *reviewerID}
	argIdx := 3

	if nextReview != nil {
		updateSQL += fmt.Sprintf(`, next_review_date = $%d`, argIdx)
		updateArgs = append(updateArgs, *nextReview)
		argIdx++
	}

	if req.RiskLevelAtReview != "" {
		updateSQL += fmt.Sprintf(`, residual_risk_level = $%d`, argIdx)
		updateArgs = append(updateArgs, req.RiskLevelAtReview)
		argIdx++
	}

	// Handle outcome-based status changes
	switch req.Outcome {
	case "revoke":
		updateSQL += `, status = 'revoked'`
	case "escalate":
		updateSQL += `, status = 'pending_approval'`
	}

	updateSQL += fmt.Sprintf(` WHERE id = $%d AND deleted_at IS NULL`, argIdx)
	updateArgs = append(updateArgs, exceptionID)

	_, err = tx.Exec(ctx, updateSQL, updateArgs...)
	if err != nil {
		return fmt.Errorf("update exception after review: %w", err)
	}

	// Audit trail
	_, _ = tx.Exec(ctx, `
		INSERT INTO exception_audit_trail
			(organization_id, exception_id, action, actor_id, details)
		VALUES ($1, $2, 'reviewed', $3, $4)`,
		orgID, exceptionID, *reviewerID,
		fmt.Sprintf("Review type: %s, outcome: %s, findings: %s",
			req.ReviewType, req.Outcome, req.Findings))

	return tx.Commit(ctx)
}

// ============================================================
// EXPIRING EXCEPTIONS
// ============================================================

// GetExpiringExceptions returns exceptions expiring within the specified number of days.
func (s *ExceptionService) GetExpiringExceptions(ctx context.Context, orgID uuid.UUID, withinDays int) ([]ComplianceException, error) {
	if withinDays < 1 {
		withinDays = 30
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.setOrgRLS(ctx, tx, orgID); err != nil {
		return nil, fmt.Errorf("set RLS: %w", err)
	}

	rows, err := tx.Query(ctx, `
		SELECT id, organization_id, exception_ref, title, description,
			exception_type::text, status::text, priority,
			scope_type::text, control_implementation_ids, framework_control_codes,
			policy_id, COALESCE(scope_description, ''),
			risk_justification, COALESCE(residual_risk_description, ''), residual_risk_level,
			risk_assessment_id, risk_accepted_by, risk_accepted_at,
			has_compensating_controls, COALESCE(compensating_controls_description, ''),
			compensating_control_ids, compensating_effectiveness::text,
			requested_by, requested_at, approved_by, approved_at,
			COALESCE(approval_comments, ''), COALESCE(rejection_reason, ''),
			workflow_instance_id,
			effective_date, expiry_date, review_frequency_months,
			next_review_date, last_review_date, last_reviewed_by, renewal_count,
			COALESCE(conditions, ''), COALESCE(business_impact_if_implemented, ''),
			regulatory_notification_required, regulator_notified_at,
			COALESCE(audit_evidence_path, ''),
			tags, metadata, created_at, updated_at, deleted_at
		FROM compliance_exceptions
		WHERE status = 'approved'
			AND expiry_date IS NOT NULL
			AND expiry_date <= CURRENT_DATE + ($1 || ' days')::interval
			AND expiry_date >= CURRENT_DATE
			AND deleted_at IS NULL
		ORDER BY expiry_date ASC`, withinDays)
	if err != nil {
		return nil, fmt.Errorf("query expiring exceptions: %w", err)
	}
	defer rows.Close()

	var exceptions []ComplianceException
	for rows.Next() {
		var exc ComplianceException
		if err := rows.Scan(
			&exc.ID, &exc.OrganizationID, &exc.ExceptionRef, &exc.Title, &exc.Description,
			&exc.ExceptionType, &exc.Status, &exc.Priority,
			&exc.ScopeType, &exc.ControlImplementationIDs, &exc.FrameworkControlCodes,
			&exc.PolicyID, &exc.ScopeDescription,
			&exc.RiskJustification, &exc.ResidualRiskDescription, &exc.ResidualRiskLevel,
			&exc.RiskAssessmentID, &exc.RiskAcceptedBy, &exc.RiskAcceptedAt,
			&exc.HasCompensatingControls, &exc.CompensatingControlsDescription,
			&exc.CompensatingControlIDs, &exc.CompensatingEffectiveness,
			&exc.RequestedBy, &exc.RequestedAt, &exc.ApprovedBy, &exc.ApprovedAt,
			&exc.ApprovalComments, &exc.RejectionReason,
			&exc.WorkflowInstanceID,
			&exc.EffectiveDate, &exc.ExpiryDate, &exc.ReviewFrequencyMonths,
			&exc.NextReviewDate, &exc.LastReviewDate, &exc.LastReviewedBy, &exc.RenewalCount,
			&exc.Conditions, &exc.BusinessImpactIfImplemented,
			&exc.RegulatoryNotificationRequired, &exc.RegulatorNotifiedAt,
			&exc.AuditEvidencePath,
			&exc.Tags, &exc.Metadata, &exc.CreatedAt, &exc.UpdatedAt, &exc.DeletedAt,
		); err != nil {
			return nil, fmt.Errorf("scan expiring exception: %w", err)
		}
		exceptions = append(exceptions, exc)
	}

	if exceptions == nil {
		exceptions = []ComplianceException{}
	}

	tx.Commit(ctx)
	return exceptions, nil
}

// ============================================================
// DASHBOARD
// ============================================================

// GetExceptionDashboard returns aggregated metrics for exception management.
func (s *ExceptionService) GetExceptionDashboard(ctx context.Context, orgID uuid.UUID) (*ExceptionDashboard, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.setOrgRLS(ctx, tx, orgID); err != nil {
		return nil, fmt.Errorf("set RLS: %w", err)
	}

	dash := &ExceptionDashboard{
		ByRiskLevel: make(map[string]int),
		ByPriority:  make(map[string]int),
		ByType:      make(map[string]int),
	}

	// Status counts
	_ = tx.QueryRow(ctx,
		`SELECT COUNT(*) FROM compliance_exceptions WHERE status = 'approved' AND deleted_at IS NULL`).Scan(&dash.ActiveCount)
	_ = tx.QueryRow(ctx,
		`SELECT COUNT(*) FROM compliance_exceptions WHERE status = 'draft' AND deleted_at IS NULL`).Scan(&dash.DraftCount)
	_ = tx.QueryRow(ctx,
		`SELECT COUNT(*) FROM compliance_exceptions WHERE status IN ('pending_risk_assessment', 'pending_approval') AND deleted_at IS NULL`).Scan(&dash.PendingApprovalCount)
	_ = tx.QueryRow(ctx,
		`SELECT COUNT(*) FROM compliance_exceptions WHERE status = 'rejected' AND deleted_at IS NULL`).Scan(&dash.RejectedCount)
	_ = tx.QueryRow(ctx,
		`SELECT COUNT(*) FROM compliance_exceptions WHERE status = 'expired' AND deleted_at IS NULL`).Scan(&dash.ExpiredCount)

	// By risk level
	riskRows, err := tx.Query(ctx, `
		SELECT residual_risk_level, COUNT(*)
		FROM compliance_exceptions
		WHERE status = 'approved' AND deleted_at IS NULL
		GROUP BY residual_risk_level`)
	if err == nil {
		defer riskRows.Close()
		for riskRows.Next() {
			var level string
			var cnt int
			if riskRows.Scan(&level, &cnt) == nil {
				dash.ByRiskLevel[level] = cnt
			}
		}
	}

	// By priority
	priRows, err := tx.Query(ctx, `
		SELECT priority, COUNT(*)
		FROM compliance_exceptions
		WHERE status = 'approved' AND deleted_at IS NULL
		GROUP BY priority`)
	if err == nil {
		defer priRows.Close()
		for priRows.Next() {
			var pri string
			var cnt int
			if priRows.Scan(&pri, &cnt) == nil {
				dash.ByPriority[pri] = cnt
			}
		}
	}

	// By type
	typeRows, err := tx.Query(ctx, `
		SELECT exception_type::text, COUNT(*)
		FROM compliance_exceptions
		WHERE status = 'approved' AND deleted_at IS NULL
		GROUP BY exception_type`)
	if err == nil {
		defer typeRows.Close()
		for typeRows.Next() {
			var et string
			var cnt int
			if typeRows.Scan(&et, &cnt) == nil {
				dash.ByType[et] = cnt
			}
		}
	}

	// Expiring counts
	_ = tx.QueryRow(ctx, `
		SELECT COUNT(*) FROM compliance_exceptions
		WHERE status = 'approved' AND expiry_date IS NOT NULL
			AND expiry_date <= CURRENT_DATE + INTERVAL '30 days'
			AND expiry_date >= CURRENT_DATE
			AND deleted_at IS NULL`).Scan(&dash.Expiring30Days)

	_ = tx.QueryRow(ctx, `
		SELECT COUNT(*) FROM compliance_exceptions
		WHERE status = 'approved' AND expiry_date IS NOT NULL
			AND expiry_date <= CURRENT_DATE + INTERVAL '60 days'
			AND expiry_date >= CURRENT_DATE
			AND deleted_at IS NULL`).Scan(&dash.Expiring60Days)

	_ = tx.QueryRow(ctx, `
		SELECT COUNT(*) FROM compliance_exceptions
		WHERE status = 'approved' AND expiry_date IS NOT NULL
			AND expiry_date <= CURRENT_DATE + INTERVAL '90 days'
			AND expiry_date >= CURRENT_DATE
			AND deleted_at IS NULL`).Scan(&dash.Expiring90Days)

	// Overdue reviews
	_ = tx.QueryRow(ctx, `
		SELECT COUNT(*) FROM compliance_exceptions
		WHERE status = 'approved'
			AND next_review_date IS NOT NULL
			AND next_review_date < CURRENT_DATE
			AND deleted_at IS NULL`).Scan(&dash.OverdueReviews)

	// Average age
	_ = tx.QueryRow(ctx, `
		SELECT COALESCE(AVG(EXTRACT(EPOCH FROM (NOW() - created_at)) / 86400), 0)
		FROM compliance_exceptions
		WHERE status = 'approved' AND deleted_at IS NULL`).Scan(&dash.AverageAgeDays)
	dash.AverageAgeDays = math.Round(dash.AverageAgeDays*10) / 10

	// Top excepted frameworks
	fwRows, err := tx.Query(ctx, `
		SELECT code, COUNT(*) as cnt
		FROM compliance_exceptions, UNNEST(framework_control_codes) AS code
		WHERE status = 'approved' AND deleted_at IS NULL
		GROUP BY code
		ORDER BY cnt DESC
		LIMIT 10`)
	if err == nil {
		defer fwRows.Close()
		for fwRows.Next() {
			var fc FrameworkExCount
			if fwRows.Scan(&fc.FrameworkCode, &fc.Count) == nil {
				dash.TopExceptedFrameworks = append(dash.TopExceptedFrameworks, fc)
			}
		}
	}
	if dash.TopExceptedFrameworks == nil {
		dash.TopExceptedFrameworks = []FrameworkExCount{}
	}

	// Recent exceptions
	recRows, err := tx.Query(ctx, `
		SELECT id, organization_id, exception_ref, title, description,
			exception_type::text, status::text, priority,
			scope_type::text, control_implementation_ids, framework_control_codes,
			policy_id, COALESCE(scope_description, ''),
			risk_justification, COALESCE(residual_risk_description, ''), residual_risk_level,
			risk_assessment_id, risk_accepted_by, risk_accepted_at,
			has_compensating_controls, COALESCE(compensating_controls_description, ''),
			compensating_control_ids, compensating_effectiveness::text,
			requested_by, requested_at, approved_by, approved_at,
			COALESCE(approval_comments, ''), COALESCE(rejection_reason, ''),
			workflow_instance_id,
			effective_date, expiry_date, review_frequency_months,
			next_review_date, last_review_date, last_reviewed_by, renewal_count,
			COALESCE(conditions, ''), COALESCE(business_impact_if_implemented, ''),
			regulatory_notification_required, regulator_notified_at,
			COALESCE(audit_evidence_path, ''),
			tags, metadata, created_at, updated_at, deleted_at
		FROM compliance_exceptions
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT 5`)
	if err == nil {
		defer recRows.Close()
		for recRows.Next() {
			var exc ComplianceException
			if recRows.Scan(
				&exc.ID, &exc.OrganizationID, &exc.ExceptionRef, &exc.Title, &exc.Description,
				&exc.ExceptionType, &exc.Status, &exc.Priority,
				&exc.ScopeType, &exc.ControlImplementationIDs, &exc.FrameworkControlCodes,
				&exc.PolicyID, &exc.ScopeDescription,
				&exc.RiskJustification, &exc.ResidualRiskDescription, &exc.ResidualRiskLevel,
				&exc.RiskAssessmentID, &exc.RiskAcceptedBy, &exc.RiskAcceptedAt,
				&exc.HasCompensatingControls, &exc.CompensatingControlsDescription,
				&exc.CompensatingControlIDs, &exc.CompensatingEffectiveness,
				&exc.RequestedBy, &exc.RequestedAt, &exc.ApprovedBy, &exc.ApprovedAt,
				&exc.ApprovalComments, &exc.RejectionReason,
				&exc.WorkflowInstanceID,
				&exc.EffectiveDate, &exc.ExpiryDate, &exc.ReviewFrequencyMonths,
				&exc.NextReviewDate, &exc.LastReviewDate, &exc.LastReviewedBy, &exc.RenewalCount,
				&exc.Conditions, &exc.BusinessImpactIfImplemented,
				&exc.RegulatoryNotificationRequired, &exc.RegulatorNotifiedAt,
				&exc.AuditEvidencePath,
				&exc.Tags, &exc.Metadata, &exc.CreatedAt, &exc.UpdatedAt, &exc.DeletedAt,
			) == nil {
				dash.RecentExceptions = append(dash.RecentExceptions, exc)
			}
		}
	}
	if dash.RecentExceptions == nil {
		dash.RecentExceptions = []ComplianceException{}
	}

	tx.Commit(ctx)
	return dash, nil
}

// ============================================================
// COMPLIANCE IMPACT
// ============================================================

// CalculateComplianceImpact calculates the compliance posture impact of an exception.
func (s *ExceptionService) CalculateComplianceImpact(ctx context.Context, orgID, exceptionID uuid.UUID) (*ComplianceImpact, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.setOrgRLS(ctx, tx, orgID); err != nil {
		return nil, fmt.Errorf("set RLS: %w", err)
	}

	exc, err := s.scanException(ctx, tx, exceptionID)
	if err != nil {
		return nil, err
	}

	impact := &ComplianceImpact{
		ExceptionID:  exceptionID,
		ExceptionRef: exc.ExceptionRef,
	}

	// Count affected controls from control_implementation_ids
	controlCount := len(exc.ControlImplementationIDs)

	// Count affected framework controls
	frameworkSet := make(map[string]bool)
	var affectedControls []AffectedControl

	for _, code := range exc.FrameworkControlCodes {
		parts := strings.SplitN(code, ".", 2)
		if len(parts) > 0 {
			frameworkSet[parts[0]] = true
		}
		affectedControls = append(affectedControls, AffectedControl{
			ControlCode: code,
			ControlName: code,
			Framework:   parts[0],
			Impact:      "excepted",
		})
	}

	// Also look up control implementations to get their framework mappings
	if len(exc.ControlImplementationIDs) > 0 {
		ciRows, err := tx.Query(ctx, `
			SELECT ci.id, fc.control_code, fc.title, f.code
			FROM control_implementations ci
			JOIN framework_controls fc ON ci.framework_control_id = fc.id
			JOIN frameworks f ON fc.framework_id = f.id
			WHERE ci.id = ANY($1)`, exc.ControlImplementationIDs)
		if err == nil {
			defer ciRows.Close()
			for ciRows.Next() {
				var ciID uuid.UUID
				var ac AffectedControl
				if ciRows.Scan(&ciID, &ac.ControlCode, &ac.ControlName, &ac.Framework) == nil {
					ac.Impact = "excepted"
					affectedControls = append(affectedControls, ac)
					frameworkSet[ac.Framework] = true
				}
			}
		}
	}

	controlCount += len(exc.FrameworkControlCodes)
	impact.AffectedControlCount = controlCount

	var frameworks []string
	for fw := range frameworkSet {
		frameworks = append(frameworks, fw)
	}
	impact.AffectedFrameworks = frameworks
	if impact.AffectedFrameworks == nil {
		impact.AffectedFrameworks = []string{}
	}

	impact.AffectedControls = affectedControls
	if impact.AffectedControls == nil {
		impact.AffectedControls = []AffectedControl{}
	}

	// Estimate compliance score impact based on risk level and control count
	var scoreImpact float64
	switch exc.ResidualRiskLevel {
	case "critical":
		scoreImpact = float64(controlCount) * 2.5
	case "high":
		scoreImpact = float64(controlCount) * 1.5
	case "medium":
		scoreImpact = float64(controlCount) * 0.8
	case "low":
		scoreImpact = float64(controlCount) * 0.3
	default:
		scoreImpact = float64(controlCount) * 0.5
	}
	impact.ComplianceScoreImpact = math.Round(scoreImpact*10) / 10

	// Risk exposure increase
	switch exc.ResidualRiskLevel {
	case "critical":
		impact.RiskExposureIncrease = "significant"
	case "high":
		impact.RiskExposureIncrease = "moderate"
	case "medium":
		impact.RiskExposureIncrease = "limited"
	default:
		impact.RiskExposureIncrease = "minimal"
	}

	// Compensating coverage
	if exc.HasCompensatingControls {
		switch exc.CompensatingEffectiveness {
		case "fully_effective":
			impact.CompensatingCoverage = 95.0
		case "mostly_effective":
			impact.CompensatingCoverage = 75.0
		case "partially_effective":
			impact.CompensatingCoverage = 50.0
		case "minimally_effective":
			impact.CompensatingCoverage = 25.0
		case "not_effective":
			impact.CompensatingCoverage = 5.0
		default:
			impact.CompensatingCoverage = 0.0
		}
	}

	// Net risk delta
	if impact.CompensatingCoverage >= 75 {
		impact.NetRiskDelta = "low"
	} else if impact.CompensatingCoverage >= 50 {
		impact.NetRiskDelta = "moderate"
	} else if exc.ResidualRiskLevel == "critical" || exc.ResidualRiskLevel == "high" {
		impact.NetRiskDelta = "high"
	} else {
		impact.NetRiskDelta = "moderate"
	}

	// Recommendations
	var recs []string
	if !exc.HasCompensatingControls {
		recs = append(recs, "Consider implementing compensating controls to reduce risk exposure")
	}
	if exc.CompensatingEffectiveness == "not_assessed" || exc.CompensatingEffectiveness == "not_effective" {
		recs = append(recs, "Assess or improve compensating control effectiveness")
	}
	if exc.ExceptionType == "temporary" && exc.RenewalCount >= 2 {
		recs = append(recs, "Maximum renewals reached; plan for permanent remediation")
	}
	if exc.ResidualRiskLevel == "critical" || exc.ResidualRiskLevel == "high" {
		recs = append(recs, "High residual risk; schedule more frequent reviews")
	}
	if exc.ExpiryDate != nil && exc.ExpiryDate.Before(time.Now().AddDate(0, 0, 30)) {
		recs = append(recs, "Exception expiring soon; initiate renewal or remediation planning")
	}
	if len(recs) == 0 {
		recs = append(recs, "Exception is well-managed with adequate compensating controls")
	}
	impact.Recommendations = recs

	tx.Commit(ctx)
	return impact, nil
}

// ============================================================
// AUDIT TRAIL QUERY
// ============================================================

// GetExceptionAuditTrail returns the audit trail for an exception.
func (s *ExceptionService) GetExceptionAuditTrail(ctx context.Context, orgID, exceptionID uuid.UUID) ([]ExceptionAuditEntry, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.setOrgRLS(ctx, tx, orgID); err != nil {
		return nil, fmt.Errorf("set RLS: %w", err)
	}

	rows, err := tx.Query(ctx, `
		SELECT id, organization_id, exception_id, action, actor_id,
			COALESCE(actor_email, ''), COALESCE(old_status, ''), COALESCE(new_status, ''),
			COALESCE(details, ''), COALESCE(ip_address, ''), COALESCE(user_agent, ''),
			metadata, created_at
		FROM exception_audit_trail
		WHERE exception_id = $1
		ORDER BY created_at DESC
		LIMIT 200`, exceptionID)
	if err != nil {
		return nil, fmt.Errorf("query audit trail: %w", err)
	}
	defer rows.Close()

	var entries []ExceptionAuditEntry
	for rows.Next() {
		var e ExceptionAuditEntry
		if err := rows.Scan(
			&e.ID, &e.OrganizationID, &e.ExceptionID, &e.Action, &e.ActorID,
			&e.ActorEmail, &e.OldStatus, &e.NewStatus,
			&e.Details, &e.IPAddress, &e.UserAgent,
			&e.Metadata, &e.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan audit entry: %w", err)
		}
		entries = append(entries, e)
	}

	if entries == nil {
		entries = []ExceptionAuditEntry{}
	}

	tx.Commit(ctx)
	return entries, nil
}

// ============================================================
// REVIEWS QUERY
// ============================================================

// GetExceptionReviews returns all reviews for an exception.
func (s *ExceptionService) GetExceptionReviews(ctx context.Context, orgID, exceptionID uuid.UUID) ([]ExceptionReview, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.setOrgRLS(ctx, tx, orgID); err != nil {
		return nil, fmt.Errorf("set RLS: %w", err)
	}

	rows, err := tx.Query(ctx, `
		SELECT id, organization_id, exception_id, review_type::text, reviewer_id,
			review_date, outcome::text, COALESCE(risk_level_at_review, ''),
			compensating_effective, COALESCE(findings, ''), COALESCE(recommendations, ''),
			next_review_date, attachments, metadata, created_at
		FROM exception_reviews
		WHERE exception_id = $1
		ORDER BY review_date DESC`, exceptionID)
	if err != nil {
		return nil, fmt.Errorf("query reviews: %w", err)
	}
	defer rows.Close()

	var reviews []ExceptionReview
	for rows.Next() {
		var r ExceptionReview
		if err := rows.Scan(
			&r.ID, &r.OrganizationID, &r.ExceptionID, &r.ReviewType, &r.ReviewerID,
			&r.ReviewDate, &r.Outcome, &r.RiskLevelAtReview,
			&r.CompensatingEffective, &r.Findings, &r.Recommendations,
			&r.NextReviewDate, &r.Attachments, &r.Metadata, &r.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan review: %w", err)
		}
		reviews = append(reviews, r)
	}

	if reviews == nil {
		reviews = []ExceptionReview{}
	}

	tx.Commit(ctx)
	return reviews, nil
}

// ============================================================
// EXPIRY MANAGEMENT
// ============================================================

// AutoExpireExceptions transitions past-due approved exceptions to expired status.
func (s *ExceptionService) AutoExpireExceptions(ctx context.Context, orgID uuid.UUID) (int64, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.setOrgRLS(ctx, tx, orgID); err != nil {
		return 0, fmt.Errorf("set RLS: %w", err)
	}

	tag, err := tx.Exec(ctx, `
		UPDATE compliance_exceptions
		SET status = 'expired'
		WHERE status = 'approved'
			AND expiry_date IS NOT NULL
			AND expiry_date < CURRENT_DATE
			AND deleted_at IS NULL`)
	if err != nil {
		return 0, fmt.Errorf("auto-expire: %w", err)
	}

	count := tag.RowsAffected()
	if count > 0 {
		log.Info().Int64("expired_count", count).Msg("Auto-expired past-due exceptions")
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit: %w", err)
	}

	return count, nil
}

// GetOverdueReviews returns approved exceptions that are overdue for review.
func (s *ExceptionService) GetOverdueReviews(ctx context.Context, orgID uuid.UUID) ([]ComplianceException, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.setOrgRLS(ctx, tx, orgID); err != nil {
		return nil, fmt.Errorf("set RLS: %w", err)
	}

	rows, err := tx.Query(ctx, `
		SELECT id, organization_id, exception_ref, title, description,
			exception_type::text, status::text, priority,
			scope_type::text, control_implementation_ids, framework_control_codes,
			policy_id, COALESCE(scope_description, ''),
			risk_justification, COALESCE(residual_risk_description, ''), residual_risk_level,
			risk_assessment_id, risk_accepted_by, risk_accepted_at,
			has_compensating_controls, COALESCE(compensating_controls_description, ''),
			compensating_control_ids, compensating_effectiveness::text,
			requested_by, requested_at, approved_by, approved_at,
			COALESCE(approval_comments, ''), COALESCE(rejection_reason, ''),
			workflow_instance_id,
			effective_date, expiry_date, review_frequency_months,
			next_review_date, last_review_date, last_reviewed_by, renewal_count,
			COALESCE(conditions, ''), COALESCE(business_impact_if_implemented, ''),
			regulatory_notification_required, regulator_notified_at,
			COALESCE(audit_evidence_path, ''),
			tags, metadata, created_at, updated_at, deleted_at
		FROM compliance_exceptions
		WHERE status = 'approved'
			AND next_review_date IS NOT NULL
			AND next_review_date < CURRENT_DATE
			AND deleted_at IS NULL
		ORDER BY next_review_date ASC`)
	if err != nil {
		return nil, fmt.Errorf("query overdue reviews: %w", err)
	}
	defer rows.Close()

	var exceptions []ComplianceException
	for rows.Next() {
		var exc ComplianceException
		if err := rows.Scan(
			&exc.ID, &exc.OrganizationID, &exc.ExceptionRef, &exc.Title, &exc.Description,
			&exc.ExceptionType, &exc.Status, &exc.Priority,
			&exc.ScopeType, &exc.ControlImplementationIDs, &exc.FrameworkControlCodes,
			&exc.PolicyID, &exc.ScopeDescription,
			&exc.RiskJustification, &exc.ResidualRiskDescription, &exc.ResidualRiskLevel,
			&exc.RiskAssessmentID, &exc.RiskAcceptedBy, &exc.RiskAcceptedAt,
			&exc.HasCompensatingControls, &exc.CompensatingControlsDescription,
			&exc.CompensatingControlIDs, &exc.CompensatingEffectiveness,
			&exc.RequestedBy, &exc.RequestedAt, &exc.ApprovedBy, &exc.ApprovedAt,
			&exc.ApprovalComments, &exc.RejectionReason,
			&exc.WorkflowInstanceID,
			&exc.EffectiveDate, &exc.ExpiryDate, &exc.ReviewFrequencyMonths,
			&exc.NextReviewDate, &exc.LastReviewDate, &exc.LastReviewedBy, &exc.RenewalCount,
			&exc.Conditions, &exc.BusinessImpactIfImplemented,
			&exc.RegulatoryNotificationRequired, &exc.RegulatorNotifiedAt,
			&exc.AuditEvidencePath,
			&exc.Tags, &exc.Metadata, &exc.CreatedAt, &exc.UpdatedAt, &exc.DeletedAt,
		); err != nil {
			return nil, fmt.Errorf("scan overdue exception: %w", err)
		}
		exceptions = append(exceptions, exc)
	}

	if exceptions == nil {
		exceptions = []ComplianceException{}
	}

	tx.Commit(ctx)
	return exceptions, nil
}
