package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// ============================================================
// TrainingService
// ============================================================

// TrainingService implements business logic for training programme
// management, assignment tracking, and compliance dashboards.
type TrainingService struct {
	pool *pgxpool.Pool
}

// NewTrainingService creates a new TrainingService with the given database pool.
func NewTrainingService(pool *pgxpool.Pool) *TrainingService {
	return &TrainingService{pool: pool}
}

// ============================================================
// DATA TYPES
// ============================================================

// TrainingProgramme represents a training programme / course.
type TrainingProgramme struct {
	ID                   uuid.UUID       `json:"id"`
	OrganizationID       uuid.UUID       `json:"organization_id"`
	ProgrammeRef         string          `json:"programme_ref"`
	Name                 string          `json:"name"`
	Description          string          `json:"description"`
	Category             string          `json:"category"`
	TargetAudience       string          `json:"target_audience"`
	TargetRoles          []string        `json:"target_roles"`
	PassingScore         int             `json:"passing_score"`
	MaxAttempts          int             `json:"max_attempts"`
	DurationMinutes      int             `json:"duration_minutes"`
	ContentVersion       string          `json:"content_version"`
	IsMandatory          bool            `json:"is_mandatory"`
	RecurrenceMonths     *int            `json:"recurrence_months"`
	DueWithinDays        *int            `json:"due_within_days"`
	ApplicableFrameworks []string        `json:"applicable_frameworks"`
	ApplicableControls   []string        `json:"applicable_controls"`
	Status               string          `json:"status"`
	IsTemplate           bool            `json:"is_template"`
	Metadata             json.RawMessage `json:"metadata"`
	CreatedBy            *uuid.UUID      `json:"created_by"`
	CreatedAt            time.Time       `json:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at"`
	CompletionRate       *float64        `json:"completion_rate,omitempty"`
	TotalAssigned        *int            `json:"total_assigned,omitempty"`
	TotalCompleted       *int            `json:"total_completed,omitempty"`
}

// TrainingAssignment tracks a user's assignment to a training programme.
type TrainingAssignment struct {
	ID                 uuid.UUID       `json:"id"`
	OrganizationID     uuid.UUID       `json:"organization_id"`
	ProgrammeID        uuid.UUID       `json:"programme_id"`
	UserID             uuid.UUID       `json:"user_id"`
	Status             string          `json:"status"`
	AssignedAt         time.Time       `json:"assigned_at"`
	DueDate            *time.Time      `json:"due_date"`
	StartedAt          *time.Time      `json:"started_at"`
	CompletedAt        *time.Time      `json:"completed_at"`
	Score              *int            `json:"score"`
	Passed             *bool           `json:"passed"`
	Attempts           int             `json:"attempts"`
	TimeSpentMinutes   int             `json:"time_spent_minutes"`
	ExemptedBy         *uuid.UUID      `json:"exempted_by"`
	ExemptionReason    string          `json:"exemption_reason"`
	CertificateURL     string          `json:"certificate_url"`
	CertificateIssuedAt *time.Time     `json:"certificate_issued_at"`
	Metadata           json.RawMessage `json:"metadata"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
	ProgrammeName      string          `json:"programme_name,omitempty"`
	ProgrammeCategory  string          `json:"programme_category,omitempty"`
	UserEmail          string          `json:"user_email,omitempty"`
	UserFullName       string          `json:"user_full_name,omitempty"`
}

// TrainingContent represents a content item or quiz question within a programme.
type TrainingContent struct {
	ID                 uuid.UUID       `json:"id"`
	OrganizationID     uuid.UUID       `json:"organization_id"`
	ProgrammeID        uuid.UUID       `json:"programme_id"`
	Title              string          `json:"title"`
	ContentType        string          `json:"content_type"`
	ContentBody        string          `json:"content_body"`
	ContentURL         string          `json:"content_url"`
	SequenceOrder      int             `json:"sequence_order"`
	DurationMinutes    *int            `json:"duration_minutes"`
	IsQuizQuestion     bool            `json:"is_quiz_question"`
	QuestionText       string          `json:"question_text"`
	AnswerOptions      json.RawMessage `json:"answer_options"`
	CorrectAnswerIndex *int            `json:"correct_answer_index"`
	Explanation        string          `json:"explanation"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}

// AssignmentResult holds the result of generating assignments.
type AssignmentResult struct {
	ProgrammeID    uuid.UUID `json:"programme_id"`
	TotalAssigned  int       `json:"total_assigned"`
	NewlyAssigned  int       `json:"newly_assigned"`
	AlreadyExists  int       `json:"already_exists"`
	DueDate        time.Time `json:"due_date"`
}

// TrainingDashboard aggregates training metrics for an org.
type TrainingDashboard struct {
	TotalProgrammes     int                    `json:"total_programmes"`
	ActiveProgrammes    int                    `json:"active_programmes"`
	TotalAssignments    int                    `json:"total_assignments"`
	CompletedCount      int                    `json:"completed_count"`
	InProgressCount     int                    `json:"in_progress_count"`
	OverdueCount        int                    `json:"overdue_count"`
	OverallCompletionRate float64              `json:"overall_completion_rate"`
	AverageScore        float64                `json:"average_score"`
	ByCategory          map[string]int         `json:"by_category"`
	ByStatus            map[string]int         `json:"by_status"`
	RecentCompletions   []TrainingAssignment   `json:"recent_completions"`
}

// ComplianceMatrix is a grid of users x programmes with completion status.
type ComplianceMatrix struct {
	Programmes []MatrixProgramme `json:"programmes"`
	Users      []MatrixUser      `json:"users"`
	Cells      [][]MatrixCell    `json:"cells"`
}

// MatrixProgramme is a column header in the compliance matrix.
type MatrixProgramme struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Ref  string    `json:"ref"`
}

// MatrixUser is a row header in the compliance matrix.
type MatrixUser struct {
	ID       uuid.UUID `json:"id"`
	FullName string    `json:"full_name"`
	Email    string    `json:"email"`
}

// MatrixCell holds the status of one user x programme intersection.
type MatrixCell struct {
	Status      string     `json:"status"`
	Score       *int       `json:"score,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	DueDate     *time.Time `json:"due_date,omitempty"`
}

// ComplianceMatrixFilter controls which rows and columns are in the matrix.
type ComplianceMatrixFilter struct {
	ProgrammeIDs []uuid.UUID `json:"programme_ids"`
	UserIDs      []uuid.UUID `json:"user_ids"`
	Department   string      `json:"department"`
}

// ============================================================
// REQUEST TYPES
// ============================================================

// CreateProgrammeReq is the request body for creating a training programme.
type CreateProgrammeReq struct {
	Name                 string   `json:"name"`
	Description          string   `json:"description"`
	Category             string   `json:"category"`
	TargetAudience       string   `json:"target_audience"`
	TargetRoles          []string `json:"target_roles"`
	PassingScore         int      `json:"passing_score"`
	MaxAttempts          int      `json:"max_attempts"`
	DurationMinutes      int      `json:"duration_minutes"`
	IsMandatory          bool     `json:"is_mandatory"`
	RecurrenceMonths     *int     `json:"recurrence_months"`
	DueWithinDays        *int     `json:"due_within_days"`
	ApplicableFrameworks []string `json:"applicable_frameworks"`
	ApplicableControls   []string `json:"applicable_controls"`
}

// UpdateProgrammeReq is the request body for updating a training programme.
type UpdateProgrammeReq struct {
	Name                 *string  `json:"name"`
	Description          *string  `json:"description"`
	Category             *string  `json:"category"`
	TargetAudience       *string  `json:"target_audience"`
	TargetRoles          []string `json:"target_roles"`
	PassingScore         *int     `json:"passing_score"`
	MaxAttempts          *int     `json:"max_attempts"`
	DurationMinutes      *int     `json:"duration_minutes"`
	IsMandatory          *bool    `json:"is_mandatory"`
	RecurrenceMonths     *int     `json:"recurrence_months"`
	DueWithinDays        *int     `json:"due_within_days"`
	ApplicableFrameworks []string `json:"applicable_frameworks"`
	ApplicableControls   []string `json:"applicable_controls"`
	Status               *string  `json:"status"`
}

// CompleteAssignmentReq is the request body for completing a training assignment.
type CompleteAssignmentReq struct {
	Score            int `json:"score"`
	TimeSpentMinutes int `json:"time_spent_minutes"`
}

// ExemptAssignmentReq is the request body for exempting a user from training.
type ExemptAssignmentReq struct {
	Reason string `json:"reason"`
}

// ============================================================
// PROGRAMME CRUD
// ============================================================

// ListProgrammes returns a paginated list of training programmes.
func (s *TrainingService) ListProgrammes(ctx context.Context, orgID uuid.UUID, page, pageSize int) ([]TrainingProgramme, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var total int64
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM training_programmes
		WHERE organization_id = $1 AND deleted_at IS NULL`, orgID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count training programmes: %w", err)
	}

	rows, err := s.pool.Query(ctx, `
		SELECT
			tp.id, tp.organization_id, tp.programme_ref, tp.name,
			COALESCE(tp.description, ''), tp.category::TEXT, tp.target_audience::TEXT,
			COALESCE(tp.target_roles, '{}'), tp.passing_score, tp.max_attempts,
			tp.duration_minutes, COALESCE(tp.content_version, '1.0'),
			tp.is_mandatory, tp.recurrence_months, tp.due_within_days,
			COALESCE(tp.applicable_frameworks, '{}'), COALESCE(tp.applicable_controls, '{}'),
			tp.status::TEXT, tp.is_template,
			COALESCE(tp.metadata, '{}'::jsonb), tp.created_by,
			tp.created_at, tp.updated_at,
			COALESCE(agg.total_assigned, 0),
			COALESCE(agg.total_completed, 0),
			CASE WHEN COALESCE(agg.total_assigned, 0) > 0
				THEN ROUND(COALESCE(agg.total_completed, 0)::numeric / agg.total_assigned * 100, 1)
				ELSE 0
			END AS completion_rate
		FROM training_programmes tp
		LEFT JOIN LATERAL (
			SELECT
				COUNT(*) AS total_assigned,
				COUNT(*) FILTER (WHERE ta.status = 'completed') AS total_completed
			FROM training_assignments ta
			WHERE ta.programme_id = tp.id AND ta.organization_id = tp.organization_id
		) agg ON true
		WHERE tp.organization_id = $1 AND tp.deleted_at IS NULL
		ORDER BY tp.created_at DESC
		LIMIT $2 OFFSET $3`,
		orgID, pageSize, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list training programmes: %w", err)
	}
	defer rows.Close()

	var programmes []TrainingProgramme
	for rows.Next() {
		var p TrainingProgramme
		var totalAssigned, totalCompleted int
		var completionRate float64
		if err := rows.Scan(
			&p.ID, &p.OrganizationID, &p.ProgrammeRef, &p.Name,
			&p.Description, &p.Category, &p.TargetAudience,
			&p.TargetRoles, &p.PassingScore, &p.MaxAttempts,
			&p.DurationMinutes, &p.ContentVersion,
			&p.IsMandatory, &p.RecurrenceMonths, &p.DueWithinDays,
			&p.ApplicableFrameworks, &p.ApplicableControls,
			&p.Status, &p.IsTemplate,
			&p.Metadata, &p.CreatedBy,
			&p.CreatedAt, &p.UpdatedAt,
			&totalAssigned, &totalCompleted, &completionRate,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan training programme: %w", err)
		}
		p.TotalAssigned = &totalAssigned
		p.TotalCompleted = &totalCompleted
		p.CompletionRate = &completionRate
		programmes = append(programmes, p)
	}
	return programmes, total, nil
}

// CreateProgramme creates a new training programme.
func (s *TrainingService) CreateProgramme(ctx context.Context, orgID uuid.UUID, req CreateProgrammeReq) (*TrainingProgramme, error) {
	if req.Category == "" {
		req.Category = "security_awareness"
	}
	if req.TargetAudience == "" {
		req.TargetAudience = "all_employees"
	}
	if req.PassingScore <= 0 {
		req.PassingScore = 80
	}
	if req.MaxAttempts <= 0 {
		req.MaxAttempts = 3
	}
	if req.DurationMinutes <= 0 {
		req.DurationMinutes = 30
	}

	p := &TrainingProgramme{}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO training_programmes (
			organization_id, programme_ref, name, description,
			category, target_audience, target_roles,
			passing_score, max_attempts, duration_minutes,
			is_mandatory, recurrence_months, due_within_days,
			applicable_frameworks, applicable_controls
		) VALUES (
			$1, generate_training_ref($1), $2, $3,
			$4::training_category, $5::training_target_audience, $6,
			$7, $8, $9,
			$10, $11, $12,
			$13, $14
		)
		RETURNING id, organization_id, programme_ref, name,
			COALESCE(description, ''), category::TEXT, target_audience::TEXT,
			COALESCE(target_roles, '{}'), passing_score, max_attempts,
			duration_minutes, COALESCE(content_version, '1.0'),
			is_mandatory, recurrence_months, due_within_days,
			COALESCE(applicable_frameworks, '{}'), COALESCE(applicable_controls, '{}'),
			status::TEXT, is_template,
			COALESCE(metadata, '{}'::jsonb), created_by,
			created_at, updated_at`,
		orgID, req.Name, req.Description,
		req.Category, req.TargetAudience, req.TargetRoles,
		req.PassingScore, req.MaxAttempts, req.DurationMinutes,
		req.IsMandatory, req.RecurrenceMonths, req.DueWithinDays,
		req.ApplicableFrameworks, req.ApplicableControls,
	).Scan(
		&p.ID, &p.OrganizationID, &p.ProgrammeRef, &p.Name,
		&p.Description, &p.Category, &p.TargetAudience,
		&p.TargetRoles, &p.PassingScore, &p.MaxAttempts,
		&p.DurationMinutes, &p.ContentVersion,
		&p.IsMandatory, &p.RecurrenceMonths, &p.DueWithinDays,
		&p.ApplicableFrameworks, &p.ApplicableControls,
		&p.Status, &p.IsTemplate,
		&p.Metadata, &p.CreatedBy,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create training programme: %w", err)
	}

	log.Info().
		Str("org_id", orgID.String()).
		Str("programme_ref", p.ProgrammeRef).
		Str("name", p.Name).
		Msg("Training programme created")

	return p, nil
}

// GetProgramme retrieves a single training programme by ID.
func (s *TrainingService) GetProgramme(ctx context.Context, orgID, programmeID uuid.UUID) (*TrainingProgramme, error) {
	p := &TrainingProgramme{}
	err := s.pool.QueryRow(ctx, `
		SELECT id, organization_id, programme_ref, name,
			COALESCE(description, ''), category::TEXT, target_audience::TEXT,
			COALESCE(target_roles, '{}'), passing_score, max_attempts,
			duration_minutes, COALESCE(content_version, '1.0'),
			is_mandatory, recurrence_months, due_within_days,
			COALESCE(applicable_frameworks, '{}'), COALESCE(applicable_controls, '{}'),
			status::TEXT, is_template,
			COALESCE(metadata, '{}'::jsonb), created_by,
			created_at, updated_at
		FROM training_programmes
		WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL`,
		programmeID, orgID,
	).Scan(
		&p.ID, &p.OrganizationID, &p.ProgrammeRef, &p.Name,
		&p.Description, &p.Category, &p.TargetAudience,
		&p.TargetRoles, &p.PassingScore, &p.MaxAttempts,
		&p.DurationMinutes, &p.ContentVersion,
		&p.IsMandatory, &p.RecurrenceMonths, &p.DueWithinDays,
		&p.ApplicableFrameworks, &p.ApplicableControls,
		&p.Status, &p.IsTemplate,
		&p.Metadata, &p.CreatedBy,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("training programme not found: %w", err)
	}
	return p, nil
}

// UpdateProgramme updates mutable fields on a training programme.
func (s *TrainingService) UpdateProgramme(ctx context.Context, orgID, programmeID uuid.UUID, req UpdateProgrammeReq) error {
	ct, err := s.pool.Exec(ctx, `
		UPDATE training_programmes SET
			name = COALESCE($3, name),
			description = COALESCE($4, description),
			category = COALESCE($5::training_category, category),
			target_audience = COALESCE($6::training_target_audience, target_audience),
			target_roles = COALESCE($7, target_roles),
			passing_score = COALESCE($8, passing_score),
			max_attempts = COALESCE($9, max_attempts),
			duration_minutes = COALESCE($10, duration_minutes),
			is_mandatory = COALESCE($11, is_mandatory),
			recurrence_months = COALESCE($12, recurrence_months),
			due_within_days = COALESCE($13, due_within_days),
			applicable_frameworks = COALESCE($14, applicable_frameworks),
			applicable_controls = COALESCE($15, applicable_controls),
			status = COALESCE($16::training_programme_status, status)
		WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL`,
		programmeID, orgID,
		req.Name, req.Description,
		req.Category, req.TargetAudience, req.TargetRoles,
		req.PassingScore, req.MaxAttempts, req.DurationMinutes,
		req.IsMandatory, req.RecurrenceMonths, req.DueWithinDays,
		req.ApplicableFrameworks, req.ApplicableControls,
		req.Status,
	)
	if err != nil {
		return fmt.Errorf("failed to update training programme: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("training programme not found")
	}
	return nil
}

// ============================================================
// ASSIGNMENT GENERATION
// ============================================================

// GenerateAssignments creates training assignments for all target users
// based on the programme's target audience configuration.
func (s *TrainingService) GenerateAssignments(ctx context.Context, orgID, programmeID uuid.UUID) (*AssignmentResult, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Set RLS context
	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return nil, fmt.Errorf("failed to set org context: %w", err)
	}

	// Fetch the programme
	var targetAudience string
	var targetRoles []string
	var dueWithinDays *int
	err = tx.QueryRow(ctx, `
		SELECT target_audience::TEXT, COALESCE(target_roles, '{}'), due_within_days
		FROM training_programmes
		WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL AND status = 'active'`,
		programmeID, orgID,
	).Scan(&targetAudience, &targetRoles, &dueWithinDays)
	if err != nil {
		return nil, fmt.Errorf("training programme not found or not active: %w", err)
	}

	// Calculate due date
	dueDays := 30
	if dueWithinDays != nil && *dueWithinDays > 0 {
		dueDays = *dueWithinDays
	}
	dueDate := time.Now().AddDate(0, 0, dueDays)

	// Build user query based on target audience
	var userQuery string
	var userArgs []interface{}
	switch targetAudience {
	case "all_employees":
		userQuery = `SELECT id FROM users WHERE organization_id = $1 AND status = 'active'`
		userArgs = []interface{}{orgID}
	case "management":
		userQuery = `SELECT id FROM users WHERE organization_id = $1 AND status = 'active' AND role IN ('admin', 'manager', 'org_admin')`
		userArgs = []interface{}{orgID}
	case "technical_staff":
		userQuery = `SELECT id FROM users WHERE organization_id = $1 AND status = 'active' AND role IN ('analyst', 'engineer', 'developer')`
		userArgs = []interface{}{orgID}
	case "new_joiners":
		userQuery = `SELECT id FROM users WHERE organization_id = $1 AND status = 'active' AND created_at > now() - interval '90 days'`
		userArgs = []interface{}{orgID}
	case "specific_roles":
		userQuery = `SELECT id FROM users WHERE organization_id = $1 AND status = 'active' AND role = ANY($2)`
		userArgs = []interface{}{orgID, targetRoles}
	case "board_members":
		userQuery = `SELECT id FROM users WHERE organization_id = $1 AND status = 'active' AND role IN ('board_member', 'org_admin')`
		userArgs = []interface{}{orgID}
	default:
		userQuery = `SELECT id FROM users WHERE organization_id = $1 AND status = 'active'`
		userArgs = []interface{}{orgID}
	}

	rows, err := tx.Query(ctx, userQuery, userArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to query target users: %w", err)
	}
	defer rows.Close()

	var userIDs []uuid.UUID
	for rows.Next() {
		var uid uuid.UUID
		if err := rows.Scan(&uid); err != nil {
			return nil, fmt.Errorf("failed to scan user ID: %w", err)
		}
		userIDs = append(userIDs, uid)
	}

	result := &AssignmentResult{
		ProgrammeID: programmeID,
		DueDate:     dueDate,
	}

	for _, uid := range userIDs {
		result.TotalAssigned++

		ct, err := tx.Exec(ctx, `
			INSERT INTO training_assignments (
				organization_id, programme_id, user_id, status, due_date
			) VALUES ($1, $2, $3, 'assigned', $4)
			ON CONFLICT (organization_id, programme_id, user_id) DO NOTHING`,
			orgID, programmeID, uid, dueDate,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create assignment for user %s: %w", uid, err)
		}
		if ct.RowsAffected() > 0 {
			result.NewlyAssigned++
		} else {
			result.AlreadyExists++
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit assignments: %w", err)
	}

	log.Info().
		Str("org_id", orgID.String()).
		Str("programme_id", programmeID.String()).
		Int("total", result.TotalAssigned).
		Int("new", result.NewlyAssigned).
		Msg("Training assignments generated")

	return result, nil
}

// ============================================================
// ASSIGNMENT CRUD
// ============================================================

// GetMyAssignments returns all training assignments for the current user.
func (s *TrainingService) GetMyAssignments(ctx context.Context, orgID, userID uuid.UUID) ([]TrainingAssignment, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT
			ta.id, ta.organization_id, ta.programme_id, ta.user_id,
			ta.status::TEXT, ta.assigned_at, ta.due_date, ta.started_at, ta.completed_at,
			ta.score, ta.passed, ta.attempts, COALESCE(ta.time_spent_minutes, 0),
			ta.exempted_by, COALESCE(ta.exemption_reason, ''),
			COALESCE(ta.certificate_url, ''), ta.certificate_issued_at,
			COALESCE(ta.metadata, '{}'::jsonb), ta.created_at, ta.updated_at,
			tp.name, tp.category::TEXT
		FROM training_assignments ta
		JOIN training_programmes tp ON tp.id = ta.programme_id
		WHERE ta.organization_id = $1 AND ta.user_id = $2
		ORDER BY
			CASE ta.status
				WHEN 'overdue' THEN 1
				WHEN 'in_progress' THEN 2
				WHEN 'assigned' THEN 3
				WHEN 'completed' THEN 4
				WHEN 'failed' THEN 5
				WHEN 'exempted' THEN 6
			END,
			ta.due_date ASC NULLS LAST`,
		orgID, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get my assignments: %w", err)
	}
	defer rows.Close()

	var assignments []TrainingAssignment
	for rows.Next() {
		var a TrainingAssignment
		if err := rows.Scan(
			&a.ID, &a.OrganizationID, &a.ProgrammeID, &a.UserID,
			&a.Status, &a.AssignedAt, &a.DueDate, &a.StartedAt, &a.CompletedAt,
			&a.Score, &a.Passed, &a.Attempts, &a.TimeSpentMinutes,
			&a.ExemptedBy, &a.ExemptionReason,
			&a.CertificateURL, &a.CertificateIssuedAt,
			&a.Metadata, &a.CreatedAt, &a.UpdatedAt,
			&a.ProgrammeName, &a.ProgrammeCategory,
		); err != nil {
			return nil, fmt.Errorf("failed to scan assignment: %w", err)
		}
		assignments = append(assignments, a)
	}
	return assignments, nil
}

// ListAssignments returns a paginated list of all training assignments for an org.
func (s *TrainingService) ListAssignments(ctx context.Context, orgID uuid.UUID, page, pageSize int) ([]TrainingAssignment, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var total int64
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM training_assignments
		WHERE organization_id = $1`, orgID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count assignments: %w", err)
	}

	rows, err := s.pool.Query(ctx, `
		SELECT
			ta.id, ta.organization_id, ta.programme_id, ta.user_id,
			ta.status::TEXT, ta.assigned_at, ta.due_date, ta.started_at, ta.completed_at,
			ta.score, ta.passed, ta.attempts, COALESCE(ta.time_spent_minutes, 0),
			ta.exempted_by, COALESCE(ta.exemption_reason, ''),
			COALESCE(ta.certificate_url, ''), ta.certificate_issued_at,
			COALESCE(ta.metadata, '{}'::jsonb), ta.created_at, ta.updated_at,
			tp.name, tp.category::TEXT,
			u.email, COALESCE(u.full_name, u.email)
		FROM training_assignments ta
		JOIN training_programmes tp ON tp.id = ta.programme_id
		JOIN users u ON u.id = ta.user_id
		WHERE ta.organization_id = $1
		ORDER BY ta.assigned_at DESC
		LIMIT $2 OFFSET $3`,
		orgID, pageSize, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list assignments: %w", err)
	}
	defer rows.Close()

	var assignments []TrainingAssignment
	for rows.Next() {
		var a TrainingAssignment
		if err := rows.Scan(
			&a.ID, &a.OrganizationID, &a.ProgrammeID, &a.UserID,
			&a.Status, &a.AssignedAt, &a.DueDate, &a.StartedAt, &a.CompletedAt,
			&a.Score, &a.Passed, &a.Attempts, &a.TimeSpentMinutes,
			&a.ExemptedBy, &a.ExemptionReason,
			&a.CertificateURL, &a.CertificateIssuedAt,
			&a.Metadata, &a.CreatedAt, &a.UpdatedAt,
			&a.ProgrammeName, &a.ProgrammeCategory,
			&a.UserEmail, &a.UserFullName,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan assignment: %w", err)
		}
		assignments = append(assignments, a)
	}
	return assignments, total, nil
}

// StartAssignment marks a training assignment as in-progress.
func (s *TrainingService) StartAssignment(ctx context.Context, orgID, assignmentID uuid.UUID) error {
	ct, err := s.pool.Exec(ctx, `
		UPDATE training_assignments SET
			status = 'in_progress',
			started_at = COALESCE(started_at, now())
		WHERE id = $1 AND organization_id = $2 AND status IN ('assigned', 'overdue')`,
		assignmentID, orgID,
	)
	if err != nil {
		return fmt.Errorf("failed to start assignment: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("assignment not found or cannot be started")
	}
	return nil
}

// CompleteAssignment records a training completion attempt and checks pass/fail.
func (s *TrainingService) CompleteAssignment(ctx context.Context, orgID, assignmentID uuid.UUID, score, timeSpent int) error {
	return s.RecordCompletion(ctx, orgID, assignmentID, score, timeSpent)
}

// RecordCompletion processes a training completion attempt, evaluating the score
// against the programme's passing threshold and max attempts.
func (s *TrainingService) RecordCompletion(ctx context.Context, orgID, assignmentID uuid.UUID, score, timeSpent int) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Fetch assignment and programme details
	var currentAttempts int
	var programmeID uuid.UUID
	var passingScore, maxAttempts int
	var status string
	err = tx.QueryRow(ctx, `
		SELECT ta.attempts, ta.programme_id, ta.status::TEXT,
			tp.passing_score, tp.max_attempts
		FROM training_assignments ta
		JOIN training_programmes tp ON tp.id = ta.programme_id
		WHERE ta.id = $1 AND ta.organization_id = $2`,
		assignmentID, orgID,
	).Scan(&currentAttempts, &programmeID, &status, &passingScore, &maxAttempts)
	if err != nil {
		return fmt.Errorf("assignment not found: %w", err)
	}

	// Cannot complete an already completed or exempted assignment
	if status == "completed" || status == "exempted" {
		return fmt.Errorf("assignment already %s", status)
	}

	newAttempts := currentAttempts + 1
	passed := score >= passingScore

	if passed {
		// Passed: mark as completed
		now := time.Now()
		certURL := fmt.Sprintf("/api/v1/training/assignments/%s/certificate", assignmentID)
		_, err = tx.Exec(ctx, `
			UPDATE training_assignments SET
				status = 'completed',
				score = $3,
				passed = true,
				attempts = $4,
				time_spent_minutes = COALESCE(time_spent_minutes, 0) + $5,
				completed_at = $6,
				certificate_url = $7,
				certificate_issued_at = $6
			WHERE id = $1 AND organization_id = $2`,
			assignmentID, orgID, score, newAttempts, timeSpent, now, certURL,
		)
	} else if newAttempts >= maxAttempts {
		// Failed and no more attempts
		_, err = tx.Exec(ctx, `
			UPDATE training_assignments SET
				status = 'failed',
				score = $3,
				passed = false,
				attempts = $4,
				time_spent_minutes = COALESCE(time_spent_minutes, 0) + $5
			WHERE id = $1 AND organization_id = $2`,
			assignmentID, orgID, score, newAttempts, timeSpent,
		)
	} else {
		// Failed but has remaining attempts: stay in_progress
		_, err = tx.Exec(ctx, `
			UPDATE training_assignments SET
				status = 'in_progress',
				score = $3,
				passed = false,
				attempts = $4,
				time_spent_minutes = COALESCE(time_spent_minutes, 0) + $5
			WHERE id = $1 AND organization_id = $2`,
			assignmentID, orgID, score, newAttempts, timeSpent,
		)
	}
	if err != nil {
		return fmt.Errorf("failed to update assignment: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit completion: %w", err)
	}

	log.Info().
		Str("assignment_id", assignmentID.String()).
		Int("score", score).
		Bool("passed", passed).
		Int("attempt", newAttempts).
		Int("max_attempts", maxAttempts).
		Msg("Training completion recorded")

	return nil
}

// ExemptAssignment marks a training assignment as exempted.
func (s *TrainingService) ExemptAssignment(ctx context.Context, orgID, assignmentID, exemptedByUserID uuid.UUID, reason string) error {
	ct, err := s.pool.Exec(ctx, `
		UPDATE training_assignments SET
			status = 'exempted',
			exempted_by = $3,
			exemption_reason = $4
		WHERE id = $1 AND organization_id = $2 AND status NOT IN ('completed', 'exempted')`,
		assignmentID, orgID, exemptedByUserID, reason,
	)
	if err != nil {
		return fmt.Errorf("failed to exempt assignment: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("assignment not found or already completed/exempted")
	}
	return nil
}

// ============================================================
// DASHBOARD & REPORTING
// ============================================================

// GetTrainingDashboard returns aggregated training metrics for an organisation.
func (s *TrainingService) GetTrainingDashboard(ctx context.Context, orgID uuid.UUID) (*TrainingDashboard, error) {
	d := &TrainingDashboard{
		ByCategory: make(map[string]int),
		ByStatus:   make(map[string]int),
	}

	// Programme counts
	err := s.pool.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE status = 'active')
		FROM training_programmes
		WHERE organization_id = $1 AND deleted_at IS NULL`, orgID,
	).Scan(&d.TotalProgrammes, &d.ActiveProgrammes)
	if err != nil {
		return nil, fmt.Errorf("failed to count programmes: %w", err)
	}

	// Assignment aggregate counts
	err = s.pool.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE status = 'completed'),
			COUNT(*) FILTER (WHERE status = 'in_progress'),
			COUNT(*) FILTER (WHERE status = 'overdue'),
			COALESCE(AVG(score) FILTER (WHERE score IS NOT NULL), 0)
		FROM training_assignments
		WHERE organization_id = $1`, orgID,
	).Scan(&d.TotalAssignments, &d.CompletedCount, &d.InProgressCount, &d.OverdueCount, &d.AverageScore)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate assignments: %w", err)
	}

	if d.TotalAssignments > 0 {
		d.OverallCompletionRate = float64(d.CompletedCount) / float64(d.TotalAssignments) * 100
	}

	// Assignments by status
	statusRows, err := s.pool.Query(ctx, `
		SELECT status::TEXT, COUNT(*)
		FROM training_assignments
		WHERE organization_id = $1
		GROUP BY status`, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get status breakdown: %w", err)
	}
	defer statusRows.Close()

	for statusRows.Next() {
		var st string
		var cnt int
		if err := statusRows.Scan(&st, &cnt); err != nil {
			return nil, fmt.Errorf("failed to scan status: %w", err)
		}
		d.ByStatus[st] = cnt
	}

	// Assignments by category
	catRows, err := s.pool.Query(ctx, `
		SELECT tp.category::TEXT, COUNT(*)
		FROM training_assignments ta
		JOIN training_programmes tp ON tp.id = ta.programme_id
		WHERE ta.organization_id = $1
		GROUP BY tp.category`, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get category breakdown: %w", err)
	}
	defer catRows.Close()

	for catRows.Next() {
		var cat string
		var cnt int
		if err := catRows.Scan(&cat, &cnt); err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		d.ByCategory[cat] = cnt
	}

	// Recent completions (last 10)
	recentRows, err := s.pool.Query(ctx, `
		SELECT
			ta.id, ta.organization_id, ta.programme_id, ta.user_id,
			ta.status::TEXT, ta.assigned_at, ta.due_date, ta.started_at, ta.completed_at,
			ta.score, ta.passed, ta.attempts, COALESCE(ta.time_spent_minutes, 0),
			ta.exempted_by, COALESCE(ta.exemption_reason, ''),
			COALESCE(ta.certificate_url, ''), ta.certificate_issued_at,
			COALESCE(ta.metadata, '{}'::jsonb), ta.created_at, ta.updated_at,
			tp.name, tp.category::TEXT,
			u.email, COALESCE(u.full_name, u.email)
		FROM training_assignments ta
		JOIN training_programmes tp ON tp.id = ta.programme_id
		JOIN users u ON u.id = ta.user_id
		WHERE ta.organization_id = $1 AND ta.status = 'completed'
		ORDER BY ta.completed_at DESC
		LIMIT 10`, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent completions: %w", err)
	}
	defer recentRows.Close()

	for recentRows.Next() {
		var a TrainingAssignment
		if err := recentRows.Scan(
			&a.ID, &a.OrganizationID, &a.ProgrammeID, &a.UserID,
			&a.Status, &a.AssignedAt, &a.DueDate, &a.StartedAt, &a.CompletedAt,
			&a.Score, &a.Passed, &a.Attempts, &a.TimeSpentMinutes,
			&a.ExemptedBy, &a.ExemptionReason,
			&a.CertificateURL, &a.CertificateIssuedAt,
			&a.Metadata, &a.CreatedAt, &a.UpdatedAt,
			&a.ProgrammeName, &a.ProgrammeCategory,
			&a.UserEmail, &a.UserFullName,
		); err != nil {
			return nil, fmt.Errorf("failed to scan recent completion: %w", err)
		}
		d.RecentCompletions = append(d.RecentCompletions, a)
	}

	return d, nil
}

// GetComplianceMatrix builds a user x programme grid showing completion status.
func (s *TrainingService) GetComplianceMatrix(ctx context.Context, orgID uuid.UUID, filter ComplianceMatrixFilter) (*ComplianceMatrix, error) {
	matrix := &ComplianceMatrix{}

	// Fetch active mandatory programmes
	progQuery := `
		SELECT id, name, programme_ref
		FROM training_programmes
		WHERE organization_id = $1 AND deleted_at IS NULL
			AND status = 'active' AND is_mandatory = true
		ORDER BY name`
	progRows, err := s.pool.Query(ctx, progQuery, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to query programmes for matrix: %w", err)
	}
	defer progRows.Close()

	for progRows.Next() {
		var mp MatrixProgramme
		if err := progRows.Scan(&mp.ID, &mp.Name, &mp.Ref); err != nil {
			return nil, fmt.Errorf("failed to scan matrix programme: %w", err)
		}
		matrix.Programmes = append(matrix.Programmes, mp)
	}

	if len(matrix.Programmes) == 0 {
		return matrix, nil
	}

	// Fetch active users
	userQuery := `
		SELECT id, COALESCE(full_name, email), email
		FROM users
		WHERE organization_id = $1 AND status = 'active'
		ORDER BY full_name, email`
	userRows, err := s.pool.Query(ctx, userQuery, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to query users for matrix: %w", err)
	}
	defer userRows.Close()

	for userRows.Next() {
		var mu MatrixUser
		if err := userRows.Scan(&mu.ID, &mu.FullName, &mu.Email); err != nil {
			return nil, fmt.Errorf("failed to scan matrix user: %w", err)
		}
		matrix.Users = append(matrix.Users, mu)
	}

	if len(matrix.Users) == 0 {
		return matrix, nil
	}

	// Load all assignments into a lookup map
	assignmentMap := make(map[string]MatrixCell) // key = "userID:programmeID"
	aRows, err := s.pool.Query(ctx, `
		SELECT user_id, programme_id, status::TEXT, score, completed_at, due_date
		FROM training_assignments
		WHERE organization_id = $1`,
		orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query assignments for matrix: %w", err)
	}
	defer aRows.Close()

	for aRows.Next() {
		var userID, programmeID uuid.UUID
		var cell MatrixCell
		if err := aRows.Scan(&userID, &programmeID, &cell.Status, &cell.Score, &cell.CompletedAt, &cell.DueDate); err != nil {
			return nil, fmt.Errorf("failed to scan matrix assignment: %w", err)
		}
		key := userID.String() + ":" + programmeID.String()
		assignmentMap[key] = cell
	}

	// Build the cells grid
	matrix.Cells = make([][]MatrixCell, len(matrix.Users))
	for i, u := range matrix.Users {
		matrix.Cells[i] = make([]MatrixCell, len(matrix.Programmes))
		for j, p := range matrix.Programmes {
			key := u.ID.String() + ":" + p.ID.String()
			if cell, ok := assignmentMap[key]; ok {
				matrix.Cells[i][j] = cell
			} else {
				matrix.Cells[i][j] = MatrixCell{Status: "not_assigned"}
			}
		}
	}

	return matrix, nil
}
