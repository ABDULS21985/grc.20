package service

import (
	"context"
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

// EvidenceTemplate represents a template defining what evidence is needed
// for a particular framework control, including collection guidance and
// validation rules.
type EvidenceTemplate struct {
	ID                          uuid.UUID       `json:"id"`
	OrganizationID              *uuid.UUID      `json:"organization_id"`
	FrameworkControlCode        string          `json:"framework_control_code"`
	FrameworkCode               string          `json:"framework_code"`
	Name                        string          `json:"name"`
	Description                 string          `json:"description"`
	EvidenceCategory            string          `json:"evidence_category"`
	CollectionMethod            string          `json:"collection_method"`
	CollectionInstructions      string          `json:"collection_instructions"`
	CollectionFrequency         string          `json:"collection_frequency"`
	TypicalCollectionTimeMinutes int            `json:"typical_collection_time_minutes"`
	ValidationRules             json.RawMessage `json:"validation_rules"`
	AcceptanceCriteria          string          `json:"acceptance_criteria"`
	CommonRejectionReasons      []string        `json:"common_rejection_reasons"`
	TemplateFields              json.RawMessage `json:"template_fields"`
	SampleEvidenceDescription   string          `json:"sample_evidence_description"`
	SampleFilePath              string          `json:"sample_file_path"`
	ApplicableTo                []string        `json:"applicable_to"`
	Difficulty                  string          `json:"difficulty"`
	AuditorPriority             string          `json:"auditor_priority"`
	IsSystem                    bool            `json:"is_system"`
	Tags                        []string        `json:"tags"`
	CreatedAt                   time.Time       `json:"created_at"`
	UpdatedAt                   time.Time       `json:"updated_at"`
}

// EvidenceRequirement represents an organization-specific evidence requirement
// linked to a control implementation and evidence template.
type EvidenceRequirement struct {
	ID                          uuid.UUID       `json:"id"`
	OrganizationID              uuid.UUID       `json:"organization_id"`
	ControlImplementationID     *uuid.UUID      `json:"control_implementation_id"`
	EvidenceTemplateID          *uuid.UUID      `json:"evidence_template_id"`
	Status                      string          `json:"status"`
	IsMandatory                 bool            `json:"is_mandatory"`
	CollectionFrequencyOverride *string         `json:"collection_frequency_override"`
	AssignedTo                  *uuid.UUID      `json:"assigned_to"`
	DueDate                     *time.Time      `json:"due_date"`
	LastCollectedAt             *time.Time      `json:"last_collected_at"`
	LastValidatedAt             *time.Time      `json:"last_validated_at"`
	LastEvidenceID              *uuid.UUID      `json:"last_evidence_id"`
	ValidationStatus            string          `json:"validation_status"`
	ValidationResults           json.RawMessage `json:"validation_results"`
	NextCollectionDue           *time.Time      `json:"next_collection_due"`
	ConsecutiveFailures         int             `json:"consecutive_failures"`
	Notes                       string          `json:"notes"`
	CreatedAt                   time.Time       `json:"created_at"`
	UpdatedAt                   time.Time       `json:"updated_at"`
	// Joined fields
	TemplateName                string          `json:"template_name,omitempty"`
	FrameworkControlCode        string          `json:"framework_control_code,omitempty"`
	FrameworkCode               string          `json:"framework_code,omitempty"`
}

// ValidationRule represents a single validation rule that can be applied
// to evidence to verify it meets requirements.
type ValidationRule struct {
	RuleType string                 `json:"rule_type"`
	Params   map[string]interface{} `json:"params,omitempty"`
}

// ValidationResult holds the outcome of validating evidence against rules.
type ValidationResult struct {
	Valid           bool                   `json:"valid"`
	OverallStatus   string                 `json:"overall_status"`
	RuleResults     []RuleResult           `json:"rule_results"`
	ValidatedAt     time.Time              `json:"validated_at"`
	EvidenceID      uuid.UUID              `json:"evidence_id"`
	RequirementID   uuid.UUID              `json:"requirement_id"`
}

// RuleResult holds the outcome of a single validation rule check.
type RuleResult struct {
	RuleType string `json:"rule_type"`
	Passed   bool   `json:"passed"`
	Message  string `json:"message"`
}

// EvidenceGapsResult holds the analysis of evidence collection gaps.
type EvidenceGapsResult struct {
	TotalRequirements int               `json:"total_requirements"`
	Collected         int               `json:"collected"`
	Pending           int               `json:"pending"`
	Overdue           int               `json:"overdue"`
	Validated         int               `json:"validated"`
	Failed            int               `json:"failed"`
	CoveragePercent   float64           `json:"coverage_percent"`
	GapsByFramework   []FrameworkGap    `json:"gaps_by_framework"`
	CriticalGaps      []CriticalGapItem `json:"critical_gaps"`
	OverdueItems      []OverdueItem     `json:"overdue_items"`
}

// FrameworkGap holds gap statistics for a single framework.
type FrameworkGap struct {
	FrameworkCode   string  `json:"framework_code"`
	TotalRequired   int     `json:"total_required"`
	Collected       int     `json:"collected"`
	CoveragePercent float64 `json:"coverage_percent"`
}

// CriticalGapItem represents a high-priority evidence gap.
type CriticalGapItem struct {
	RequirementID        uuid.UUID  `json:"requirement_id"`
	TemplateName         string     `json:"template_name"`
	FrameworkControlCode string     `json:"framework_control_code"`
	FrameworkCode        string     `json:"framework_code"`
	AuditorPriority      string     `json:"auditor_priority"`
	DueDate              *time.Time `json:"due_date"`
	DaysOverdue          int        `json:"days_overdue"`
}

// OverdueItem represents an overdue evidence requirement.
type OverdueItem struct {
	RequirementID        uuid.UUID  `json:"requirement_id"`
	TemplateName         string     `json:"template_name"`
	FrameworkControlCode string     `json:"framework_control_code"`
	DueDate              *time.Time `json:"due_date"`
	DaysOverdue          int        `json:"days_overdue"`
	AssignedTo           *uuid.UUID `json:"assigned_to"`
}

// CollectionSchedule provides an overview of upcoming evidence collection tasks.
type CollectionSchedule struct {
	UpcomingThisWeek  []ScheduleItem `json:"upcoming_this_week"`
	UpcomingThisMonth []ScheduleItem `json:"upcoming_this_month"`
	UpcomingThisQuarter []ScheduleItem `json:"upcoming_this_quarter"`
	TotalScheduled    int            `json:"total_scheduled"`
}

// ScheduleItem represents a single scheduled evidence collection task.
type ScheduleItem struct {
	RequirementID        uuid.UUID  `json:"requirement_id"`
	TemplateName         string     `json:"template_name"`
	FrameworkControlCode string     `json:"framework_control_code"`
	FrameworkCode        string     `json:"framework_code"`
	CollectionFrequency  string     `json:"collection_frequency"`
	NextDue              *time.Time `json:"next_due"`
	AssignedTo           *uuid.UUID `json:"assigned_to"`
	Difficulty           string     `json:"difficulty"`
	EstimatedMinutes     int        `json:"estimated_minutes"`
}

// GenerateResult holds the outcome of generating evidence requirements.
type GenerateResult struct {
	Created int `json:"created"`
	Skipped int `json:"skipped"`
	Total   int `json:"total"`
	Message string `json:"message"`
}

// TemplateListResult wraps a paginated template list.
type TemplateListResult struct {
	Templates []EvidenceTemplate `json:"templates"`
	Total     int64              `json:"total"`
}

// RequirementListResult wraps a paginated requirement list.
type RequirementListResult struct {
	Requirements []EvidenceRequirement `json:"requirements"`
	Total        int64                 `json:"total"`
}

// ============================================================
// FILTER / REQUEST TYPES
// ============================================================

// TemplateFilter holds filter parameters for listing evidence templates.
type TemplateFilter struct {
	FrameworkCode    string `json:"framework_code"`
	ControlCode      string `json:"control_code"`
	Category         string `json:"category"`
	Difficulty       string `json:"difficulty"`
	AuditorPriority  string `json:"auditor_priority"`
	Search           string `json:"search"`
	IsSystem         *bool  `json:"is_system"`
	Page             int    `json:"page"`
	PageSize         int    `json:"page_size"`
}

// RequirementFilter holds filter parameters for listing evidence requirements.
type RequirementFilter struct {
	Status           string     `json:"status"`
	ValidationStatus string     `json:"validation_status"`
	FrameworkCode    string     `json:"framework_code"`
	AssignedTo       *uuid.UUID `json:"assigned_to"`
	IsMandatory      *bool      `json:"is_mandatory"`
	Page             int        `json:"page"`
	PageSize         int        `json:"page_size"`
}

// CreateTemplateRequest defines the input for creating a custom evidence template.
type CreateTemplateRequest struct {
	FrameworkControlCode        string          `json:"framework_control_code"`
	FrameworkCode               string          `json:"framework_code"`
	Name                        string          `json:"name"`
	Description                 string          `json:"description"`
	EvidenceCategory            string          `json:"evidence_category"`
	CollectionMethod            string          `json:"collection_method"`
	CollectionInstructions      string          `json:"collection_instructions"`
	CollectionFrequency         string          `json:"collection_frequency"`
	TypicalCollectionTimeMinutes int            `json:"typical_collection_time_minutes"`
	ValidationRules             json.RawMessage `json:"validation_rules"`
	AcceptanceCriteria          string          `json:"acceptance_criteria"`
	CommonRejectionReasons      []string        `json:"common_rejection_reasons"`
	TemplateFields              json.RawMessage `json:"template_fields"`
	Difficulty                  string          `json:"difficulty"`
	AuditorPriority             string          `json:"auditor_priority"`
	Tags                        []string        `json:"tags"`
}

// UpdateRequirementRequest defines the input for updating an evidence requirement.
type UpdateRequirementRequest struct {
	Status                      *string    `json:"status"`
	IsMandatory                 *bool      `json:"is_mandatory"`
	CollectionFrequencyOverride *string    `json:"collection_frequency_override"`
	AssignedTo                  *uuid.UUID `json:"assigned_to"`
	DueDate                     *string    `json:"due_date"`
	Notes                       *string    `json:"notes"`
}

// ============================================================
// SERVICE
// ============================================================

// EvidenceTemplateService provides business logic for evidence template
// management, requirement tracking, evidence validation, and gap analysis.
type EvidenceTemplateService struct {
	pool *pgxpool.Pool
}

// NewEvidenceTemplateService creates a new EvidenceTemplateService.
func NewEvidenceTemplateService(pool *pgxpool.Pool) *EvidenceTemplateService {
	return &EvidenceTemplateService{pool: pool}
}

// setRLS sets the RLS context for the current transaction.
func (s *EvidenceTemplateService) setRLS(ctx context.Context, tx pgx.Tx, orgID uuid.UUID) error {
	_, err := tx.Exec(ctx, "SET LOCAL app.current_org = '"+orgID.String()+"'")
	return err
}

// ============================================================
// TEMPLATE OPERATIONS
// ============================================================

// GetTemplatesForControl returns all evidence templates applicable to a specific control.
func (s *EvidenceTemplateService) GetTemplatesForControl(ctx context.Context, controlCode, frameworkCode string) ([]EvidenceTemplate, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, organization_id, framework_control_code, framework_code,
		       name, description, evidence_category, collection_method,
		       collection_instructions, collection_frequency,
		       typical_collection_time_minutes, validation_rules,
		       acceptance_criteria, common_rejection_reasons,
		       template_fields, sample_evidence_description, sample_file_path,
		       applicable_to, difficulty, auditor_priority, is_system, tags,
		       created_at, updated_at
		FROM evidence_templates
		WHERE framework_control_code = $1 AND framework_code = $2
		  AND (organization_id IS NULL OR is_system = true)
		ORDER BY auditor_priority ASC, name ASC`,
		controlCode, frameworkCode)
	if err != nil {
		return nil, fmt.Errorf("query templates for control: %w", err)
	}
	defer rows.Close()

	return scanTemplates(rows)
}

// GetTemplatesForFramework returns all evidence templates for a given framework,
// including both system templates and org-specific custom templates.
func (s *EvidenceTemplateService) GetTemplatesForFramework(ctx context.Context, orgID uuid.UUID, frameworkCode string) ([]EvidenceTemplate, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.setRLS(ctx, tx, orgID); err != nil {
		return nil, fmt.Errorf("set rls: %w", err)
	}

	rows, err := tx.Query(ctx, `
		SELECT id, organization_id, framework_control_code, framework_code,
		       name, description, evidence_category, collection_method,
		       collection_instructions, collection_frequency,
		       typical_collection_time_minutes, validation_rules,
		       acceptance_criteria, common_rejection_reasons,
		       template_fields, sample_evidence_description, sample_file_path,
		       applicable_to, difficulty, auditor_priority, is_system, tags,
		       created_at, updated_at
		FROM evidence_templates
		WHERE framework_code = $1
		ORDER BY framework_control_code ASC, auditor_priority ASC`,
		frameworkCode)
	if err != nil {
		return nil, fmt.Errorf("query templates for framework: %w", err)
	}
	defer rows.Close()

	return scanTemplates(rows)
}

// ListTemplates returns a filtered, paginated list of evidence templates.
func (s *EvidenceTemplateService) ListTemplates(ctx context.Context, orgID uuid.UUID, filter TemplateFilter) (*TemplateListResult, error) {
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

	if err := s.setRLS(ctx, tx, orgID); err != nil {
		return nil, fmt.Errorf("set rls: %w", err)
	}

	var conditions []string
	var args []interface{}
	argIdx := 1

	if filter.FrameworkCode != "" {
		conditions = append(conditions, fmt.Sprintf("framework_code = $%d", argIdx))
		args = append(args, filter.FrameworkCode)
		argIdx++
	}
	if filter.ControlCode != "" {
		conditions = append(conditions, fmt.Sprintf("framework_control_code = $%d", argIdx))
		args = append(args, filter.ControlCode)
		argIdx++
	}
	if filter.Category != "" {
		conditions = append(conditions, fmt.Sprintf("evidence_category = $%d", argIdx))
		args = append(args, filter.Category)
		argIdx++
	}
	if filter.Difficulty != "" {
		conditions = append(conditions, fmt.Sprintf("difficulty = $%d", argIdx))
		args = append(args, filter.Difficulty)
		argIdx++
	}
	if filter.AuditorPriority != "" {
		conditions = append(conditions, fmt.Sprintf("auditor_priority = $%d", argIdx))
		args = append(args, filter.AuditorPriority)
		argIdx++
	}
	if filter.IsSystem != nil {
		conditions = append(conditions, fmt.Sprintf("is_system = $%d", argIdx))
		args = append(args, *filter.IsSystem)
		argIdx++
	}
	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf(
			"(name ILIKE '%%' || $%d || '%%' OR description ILIKE '%%' || $%d || '%%' OR framework_control_code ILIKE '%%' || $%d || '%%')",
			argIdx, argIdx, argIdx))
		args = append(args, filter.Search)
		argIdx++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count query
	var total int64
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM evidence_templates %s", where)
	if err := tx.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count templates: %w", err)
	}

	// Data query
	dataSQL := fmt.Sprintf(`
		SELECT id, organization_id, framework_control_code, framework_code,
		       name, description, evidence_category, collection_method,
		       collection_instructions, collection_frequency,
		       typical_collection_time_minutes, validation_rules,
		       acceptance_criteria, common_rejection_reasons,
		       template_fields, sample_evidence_description, sample_file_path,
		       applicable_to, difficulty, auditor_priority, is_system, tags,
		       created_at, updated_at
		FROM evidence_templates
		%s
		ORDER BY framework_code ASC, framework_control_code ASC, name ASC
		LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)

	args = append(args, filter.PageSize, offset)

	rows, err := tx.Query(ctx, dataSQL, args...)
	if err != nil {
		return nil, fmt.Errorf("query templates: %w", err)
	}
	defer rows.Close()

	templates, err := scanTemplates(rows)
	if err != nil {
		return nil, err
	}

	return &TemplateListResult{Templates: templates, Total: total}, nil
}

// GetTemplate returns a single evidence template by ID.
func (s *EvidenceTemplateService) GetTemplate(ctx context.Context, orgID, templateID uuid.UUID) (*EvidenceTemplate, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.setRLS(ctx, tx, orgID); err != nil {
		return nil, fmt.Errorf("set rls: %w", err)
	}

	var t EvidenceTemplate
	err = tx.QueryRow(ctx, `
		SELECT id, organization_id, framework_control_code, framework_code,
		       name, description, evidence_category, collection_method,
		       collection_instructions, collection_frequency,
		       typical_collection_time_minutes, validation_rules,
		       acceptance_criteria, common_rejection_reasons,
		       template_fields, sample_evidence_description, sample_file_path,
		       applicable_to, difficulty, auditor_priority, is_system, tags,
		       created_at, updated_at
		FROM evidence_templates
		WHERE id = $1`, templateID).Scan(
		&t.ID, &t.OrganizationID, &t.FrameworkControlCode, &t.FrameworkCode,
		&t.Name, &t.Description, &t.EvidenceCategory, &t.CollectionMethod,
		&t.CollectionInstructions, &t.CollectionFrequency,
		&t.TypicalCollectionTimeMinutes, &t.ValidationRules,
		&t.AcceptanceCriteria, &t.CommonRejectionReasons,
		&t.TemplateFields, &t.SampleEvidenceDescription, &t.SampleFilePath,
		&t.ApplicableTo, &t.Difficulty, &t.AuditorPriority, &t.IsSystem, &t.Tags,
		&t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("evidence template not found")
		}
		return nil, fmt.Errorf("get template: %w", err)
	}

	return &t, nil
}

// CreateTemplate creates a new custom evidence template for an organization.
func (s *EvidenceTemplateService) CreateTemplate(ctx context.Context, orgID uuid.UUID, req CreateTemplateRequest) (*EvidenceTemplate, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("template name is required")
	}
	if req.FrameworkControlCode == "" || req.FrameworkCode == "" {
		return nil, fmt.Errorf("framework_control_code and framework_code are required")
	}
	if req.EvidenceCategory == "" {
		req.EvidenceCategory = "document"
	}
	if req.CollectionMethod == "" {
		req.CollectionMethod = "manual_upload"
	}
	if req.CollectionFrequency == "" {
		req.CollectionFrequency = "annually"
	}
	if req.Difficulty == "" {
		req.Difficulty = "moderate"
	}
	if req.AuditorPriority == "" {
		req.AuditorPriority = "medium"
	}
	if req.ValidationRules == nil {
		req.ValidationRules = json.RawMessage(`[]`)
	}
	if req.TemplateFields == nil {
		req.TemplateFields = json.RawMessage(`[]`)
	}
	if req.CommonRejectionReasons == nil {
		req.CommonRejectionReasons = []string{}
	}
	if req.Tags == nil {
		req.Tags = []string{}
	}
	if req.TypicalCollectionTimeMinutes <= 0 {
		req.TypicalCollectionTimeMinutes = 30
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.setRLS(ctx, tx, orgID); err != nil {
		return nil, fmt.Errorf("set rls: %w", err)
	}

	var t EvidenceTemplate
	err = tx.QueryRow(ctx, `
		INSERT INTO evidence_templates (
			organization_id, framework_control_code, framework_code,
			name, description, evidence_category, collection_method,
			collection_instructions, collection_frequency,
			typical_collection_time_minutes, validation_rules,
			acceptance_criteria, common_rejection_reasons,
			template_fields, difficulty, auditor_priority, is_system, tags
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,false,$17)
		RETURNING id, organization_id, framework_control_code, framework_code,
		          name, description, evidence_category, collection_method,
		          collection_instructions, collection_frequency,
		          typical_collection_time_minutes, validation_rules,
		          acceptance_criteria, common_rejection_reasons,
		          template_fields, sample_evidence_description, sample_file_path,
		          applicable_to, difficulty, auditor_priority, is_system, tags,
		          created_at, updated_at`,
		orgID, req.FrameworkControlCode, req.FrameworkCode,
		req.Name, req.Description, req.EvidenceCategory, req.CollectionMethod,
		req.CollectionInstructions, req.CollectionFrequency,
		req.TypicalCollectionTimeMinutes, req.ValidationRules,
		req.AcceptanceCriteria, req.CommonRejectionReasons,
		req.TemplateFields, req.Difficulty, req.AuditorPriority, req.Tags,
	).Scan(
		&t.ID, &t.OrganizationID, &t.FrameworkControlCode, &t.FrameworkCode,
		&t.Name, &t.Description, &t.EvidenceCategory, &t.CollectionMethod,
		&t.CollectionInstructions, &t.CollectionFrequency,
		&t.TypicalCollectionTimeMinutes, &t.ValidationRules,
		&t.AcceptanceCriteria, &t.CommonRejectionReasons,
		&t.TemplateFields, &t.SampleEvidenceDescription, &t.SampleFilePath,
		&t.ApplicableTo, &t.Difficulty, &t.AuditorPriority, &t.IsSystem, &t.Tags,
		&t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert template: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	log.Info().Str("template_id", t.ID.String()).Str("name", t.Name).Msg("evidence template created")
	return &t, nil
}

// ============================================================
// REQUIREMENT OPERATIONS
// ============================================================

// GenerateEvidenceRequirements automatically creates evidence requirements for
// all controls of a framework that have matching system templates.
func (s *EvidenceTemplateService) GenerateEvidenceRequirements(ctx context.Context, orgID, frameworkID uuid.UUID) (*GenerateResult, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.setRLS(ctx, tx, orgID); err != nil {
		return nil, fmt.Errorf("set rls: %w", err)
	}

	// Get framework code for the given framework ID
	var frameworkCode string
	err = tx.QueryRow(ctx, `SELECT code FROM compliance_frameworks WHERE id = $1`, frameworkID).Scan(&frameworkCode)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("framework not found")
		}
		return nil, fmt.Errorf("get framework code: %w", err)
	}

	// Get all system templates for this framework
	rows, err := tx.Query(ctx, `
		SELECT et.id, et.framework_control_code, et.collection_frequency
		FROM evidence_templates et
		WHERE et.framework_code = $1 AND et.is_system = true`, frameworkCode)
	if err != nil {
		return nil, fmt.Errorf("query templates: %w", err)
	}
	defer rows.Close()

	type templateInfo struct {
		ID          uuid.UUID
		ControlCode string
		Frequency   string
	}

	var templates []templateInfo
	for rows.Next() {
		var ti templateInfo
		if err := rows.Scan(&ti.ID, &ti.ControlCode, &ti.Frequency); err != nil {
			return nil, fmt.Errorf("scan template: %w", err)
		}
		templates = append(templates, ti)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate templates: %w", err)
	}

	result := &GenerateResult{Total: len(templates)}

	for _, tmpl := range templates {
		// Check if requirement already exists
		var exists bool
		err := tx.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM evidence_requirements
				WHERE organization_id = $1 AND evidence_template_id = $2
			)`, orgID, tmpl.ID).Scan(&exists)
		if err != nil {
			return nil, fmt.Errorf("check existing: %w", err)
		}

		if exists {
			result.Skipped++
			continue
		}

		// Calculate next collection due date based on frequency
		nextDue := calculateNextDue(tmpl.Frequency)

		_, err = tx.Exec(ctx, `
			INSERT INTO evidence_requirements (
				organization_id, evidence_template_id, status,
				is_mandatory, next_collection_due
			) VALUES ($1, $2, 'pending', true, $3)`,
			orgID, tmpl.ID, nextDue)
		if err != nil {
			log.Warn().Err(err).Str("template_id", tmpl.ID.String()).Msg("failed to create evidence requirement")
			result.Skipped++
			continue
		}
		result.Created++
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	result.Message = fmt.Sprintf("Generated %d evidence requirements (%d skipped, already exist)", result.Created, result.Skipped)
	log.Info().
		Str("org_id", orgID.String()).
		Str("framework", frameworkCode).
		Int("created", result.Created).
		Int("skipped", result.Skipped).
		Msg("evidence requirements generated")

	return result, nil
}

// ListRequirements returns a filtered, paginated list of evidence requirements.
func (s *EvidenceTemplateService) ListRequirements(ctx context.Context, orgID uuid.UUID, filter RequirementFilter) (*RequirementListResult, error) {
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

	if err := s.setRLS(ctx, tx, orgID); err != nil {
		return nil, fmt.Errorf("set rls: %w", err)
	}

	var conditions []string
	var args []interface{}
	argIdx := 1

	conditions = append(conditions, fmt.Sprintf("er.organization_id = $%d", argIdx))
	args = append(args, orgID)
	argIdx++

	if filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("er.status = $%d", argIdx))
		args = append(args, filter.Status)
		argIdx++
	}
	if filter.ValidationStatus != "" {
		conditions = append(conditions, fmt.Sprintf("er.validation_status = $%d", argIdx))
		args = append(args, filter.ValidationStatus)
		argIdx++
	}
	if filter.FrameworkCode != "" {
		conditions = append(conditions, fmt.Sprintf("et.framework_code = $%d", argIdx))
		args = append(args, filter.FrameworkCode)
		argIdx++
	}
	if filter.AssignedTo != nil {
		conditions = append(conditions, fmt.Sprintf("er.assigned_to = $%d", argIdx))
		args = append(args, *filter.AssignedTo)
		argIdx++
	}
	if filter.IsMandatory != nil {
		conditions = append(conditions, fmt.Sprintf("er.is_mandatory = $%d", argIdx))
		args = append(args, *filter.IsMandatory)
		argIdx++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count query
	var total int64
	countSQL := fmt.Sprintf(`
		SELECT COUNT(*) FROM evidence_requirements er
		LEFT JOIN evidence_templates et ON er.evidence_template_id = et.id
		%s`, where)
	if err := tx.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count requirements: %w", err)
	}

	// Data query
	dataSQL := fmt.Sprintf(`
		SELECT er.id, er.organization_id, er.control_implementation_id,
		       er.evidence_template_id, er.status, er.is_mandatory,
		       er.collection_frequency_override, er.assigned_to, er.due_date,
		       er.last_collected_at, er.last_validated_at, er.last_evidence_id,
		       er.validation_status, er.validation_results, er.next_collection_due,
		       er.consecutive_failures, er.notes, er.created_at, er.updated_at,
		       COALESCE(et.name, '') AS template_name,
		       COALESCE(et.framework_control_code, '') AS framework_control_code,
		       COALESCE(et.framework_code, '') AS framework_code
		FROM evidence_requirements er
		LEFT JOIN evidence_templates et ON er.evidence_template_id = et.id
		%s
		ORDER BY er.status ASC, er.due_date ASC NULLS LAST, er.created_at DESC
		LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)

	args = append(args, filter.PageSize, offset)

	rows, err := tx.Query(ctx, dataSQL, args...)
	if err != nil {
		return nil, fmt.Errorf("query requirements: %w", err)
	}
	defer rows.Close()

	var requirements []EvidenceRequirement
	for rows.Next() {
		var r EvidenceRequirement
		if err := rows.Scan(
			&r.ID, &r.OrganizationID, &r.ControlImplementationID,
			&r.EvidenceTemplateID, &r.Status, &r.IsMandatory,
			&r.CollectionFrequencyOverride, &r.AssignedTo, &r.DueDate,
			&r.LastCollectedAt, &r.LastValidatedAt, &r.LastEvidenceID,
			&r.ValidationStatus, &r.ValidationResults, &r.NextCollectionDue,
			&r.ConsecutiveFailures, &r.Notes, &r.CreatedAt, &r.UpdatedAt,
			&r.TemplateName, &r.FrameworkControlCode, &r.FrameworkCode,
		); err != nil {
			return nil, fmt.Errorf("scan requirement: %w", err)
		}
		requirements = append(requirements, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate requirements: %w", err)
	}
	if requirements == nil {
		requirements = []EvidenceRequirement{}
	}

	return &RequirementListResult{Requirements: requirements, Total: total}, nil
}

// UpdateRequirement updates an evidence requirement.
func (s *EvidenceTemplateService) UpdateRequirement(ctx context.Context, orgID, reqID uuid.UUID, req UpdateRequirementRequest) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.setRLS(ctx, tx, orgID); err != nil {
		return fmt.Errorf("set rls: %w", err)
	}

	var setClauses []string
	var args []interface{}
	argIdx := 1

	if req.Status != nil {
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *req.Status)
		argIdx++
	}
	if req.IsMandatory != nil {
		setClauses = append(setClauses, fmt.Sprintf("is_mandatory = $%d", argIdx))
		args = append(args, *req.IsMandatory)
		argIdx++
	}
	if req.CollectionFrequencyOverride != nil {
		setClauses = append(setClauses, fmt.Sprintf("collection_frequency_override = $%d", argIdx))
		args = append(args, *req.CollectionFrequencyOverride)
		argIdx++
	}
	if req.AssignedTo != nil {
		setClauses = append(setClauses, fmt.Sprintf("assigned_to = $%d", argIdx))
		args = append(args, *req.AssignedTo)
		argIdx++
	}
	if req.DueDate != nil {
		setClauses = append(setClauses, fmt.Sprintf("due_date = $%d", argIdx))
		args = append(args, *req.DueDate)
		argIdx++
	}
	if req.Notes != nil {
		setClauses = append(setClauses, fmt.Sprintf("notes = $%d", argIdx))
		args = append(args, *req.Notes)
		argIdx++
	}

	if len(setClauses) == 0 {
		return fmt.Errorf("no fields to update")
	}

	setClauses = append(setClauses, "updated_at = NOW()")

	sql := fmt.Sprintf(`
		UPDATE evidence_requirements SET %s
		WHERE id = $%d AND organization_id = $%d`,
		strings.Join(setClauses, ", "), argIdx, argIdx+1)
	args = append(args, reqID, orgID)

	tag, err := tx.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("update requirement: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("evidence requirement not found")
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	return nil
}

// ============================================================
// VALIDATION ENGINE
// ============================================================

// ValidateEvidence validates collected evidence against the requirement's template rules.
func (s *EvidenceTemplateService) ValidateEvidence(ctx context.Context, orgID, requirementID, evidenceID uuid.UUID) (*ValidationResult, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.setRLS(ctx, tx, orgID); err != nil {
		return nil, fmt.Errorf("set rls: %w", err)
	}

	// Get the requirement and template validation rules
	var rulesJSON json.RawMessage
	var templateID *uuid.UUID
	err = tx.QueryRow(ctx, `
		SELECT er.evidence_template_id, COALESCE(et.validation_rules, '[]'::jsonb)
		FROM evidence_requirements er
		LEFT JOIN evidence_templates et ON er.evidence_template_id = et.id
		WHERE er.id = $1 AND er.organization_id = $2`,
		requirementID, orgID).Scan(&templateID, &rulesJSON)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("evidence requirement not found")
		}
		return nil, fmt.Errorf("get requirement: %w", err)
	}

	// Get evidence details
	var fileSize int64
	var fileName string
	var mimeType string
	var collectedAt time.Time
	err = tx.QueryRow(ctx, `
		SELECT COALESCE(file_size_bytes, 0), COALESCE(file_name, ''),
		       COALESCE(mime_type, ''), COALESCE(collected_at, NOW())
		FROM control_evidence
		WHERE id = $1 AND organization_id = $2`,
		evidenceID, orgID).Scan(&fileSize, &fileName, &mimeType, &collectedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("evidence not found")
		}
		return nil, fmt.Errorf("get evidence: %w", err)
	}

	// Parse validation rules
	var rules []ValidationRule
	if err := json.Unmarshal(rulesJSON, &rules); err != nil {
		return nil, fmt.Errorf("parse validation rules: %w", err)
	}

	// Run validation rules
	ruleResults := RunValidationRules(rules, fileSize, fileName, mimeType, collectedAt)

	// Determine overall status
	allPassed := true
	for _, rr := range ruleResults {
		if !rr.Passed {
			allPassed = false
			break
		}
	}

	overallStatus := "pass"
	if !allPassed {
		overallStatus = "fail"
	}

	result := &ValidationResult{
		Valid:         allPassed,
		OverallStatus: overallStatus,
		RuleResults:   ruleResults,
		ValidatedAt:   time.Now(),
		EvidenceID:    evidenceID,
		RequirementID: requirementID,
	}

	// Update the requirement with validation results
	resultsJSON, _ := json.Marshal(result)
	validationStatus := "pass"
	if !allPassed {
		validationStatus = "fail"
	}

	consecutiveFailuresUpdate := "0"
	if !allPassed {
		consecutiveFailuresUpdate = "consecutive_failures + 1"
	}

	_, err = tx.Exec(ctx, fmt.Sprintf(`
		UPDATE evidence_requirements SET
			last_validated_at = NOW(),
			last_evidence_id = $1,
			validation_status = $2,
			validation_results = $3,
			consecutive_failures = %s,
			status = CASE WHEN $4 = 'pass' THEN 'validated' ELSE 'rejected' END,
			updated_at = NOW()
		WHERE id = $5 AND organization_id = $6`, consecutiveFailuresUpdate),
		evidenceID, validationStatus, resultsJSON, validationStatus,
		requirementID, orgID)
	if err != nil {
		return nil, fmt.Errorf("update requirement validation: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return result, nil
}

// RunValidationRules executes a set of validation rules against evidence metadata.
// This is exported for testing.
func RunValidationRules(rules []ValidationRule, fileSize int64, fileName, mimeType string, collectedAt time.Time) []RuleResult {
	var results []RuleResult

	for _, rule := range rules {
		switch rule.RuleType {
		case "file_not_empty":
			results = append(results, RuleResult{
				RuleType: "file_not_empty",
				Passed:   fileSize > 0,
				Message:  fmt.Sprintf("File size: %d bytes", fileSize),
			})

		case "date_within":
			days := 365
			if d, ok := rule.Params["days"]; ok {
				switch v := d.(type) {
				case float64:
					days = int(v)
				case int:
					days = v
				}
			}
			cutoff := time.Now().AddDate(0, 0, -days)
			passed := collectedAt.After(cutoff)
			results = append(results, RuleResult{
				RuleType: "date_within",
				Passed:   passed,
				Message:  fmt.Sprintf("Evidence collected at %s (must be within %d days)", collectedAt.Format("2006-01-02"), days),
			})

		case "contains_text":
			text := ""
			if t, ok := rule.Params["text"]; ok {
				if s, ok := t.(string); ok {
					text = s
				}
			}
			// For file-based evidence, we check the filename as a basic proxy.
			// A production system would inspect file contents.
			passed := text == "" || strings.Contains(strings.ToLower(fileName), strings.ToLower(text))
			results = append(results, RuleResult{
				RuleType: "contains_text",
				Passed:   passed,
				Message:  fmt.Sprintf("Checking for text '%s' in filename", text),
			})

		case "file_type":
			allowedTypes := ""
			if ft, ok := rule.Params["allowed"]; ok {
				if s, ok := ft.(string); ok {
					allowedTypes = s
				}
			}
			passed := allowedTypes == "" || containsFileType(mimeType, fileName, allowedTypes)
			results = append(results, RuleResult{
				RuleType: "file_type",
				Passed:   passed,
				Message:  fmt.Sprintf("File type: %s (allowed: %s)", mimeType, allowedTypes),
			})

		case "file_size":
			maxBytes := int64(0)
			minBytes := int64(0)
			if max, ok := rule.Params["max_bytes"]; ok {
				switch v := max.(type) {
				case float64:
					maxBytes = int64(v)
				case int:
					maxBytes = int64(v)
				}
			}
			if min, ok := rule.Params["min_bytes"]; ok {
				switch v := min.(type) {
				case float64:
					minBytes = int64(v)
				case int:
					minBytes = int64(v)
				}
			}
			passed := true
			msg := fmt.Sprintf("File size: %d bytes", fileSize)
			if maxBytes > 0 && fileSize > maxBytes {
				passed = false
				msg += fmt.Sprintf(" (exceeds max %d)", maxBytes)
			}
			if minBytes > 0 && fileSize < minBytes {
				passed = false
				msg += fmt.Sprintf(" (below min %d)", minBytes)
			}
			results = append(results, RuleResult{
				RuleType: "file_size",
				Passed:   passed,
				Message:  msg,
			})

		default:
			results = append(results, RuleResult{
				RuleType: rule.RuleType,
				Passed:   true,
				Message:  fmt.Sprintf("Unknown rule type '%s' — skipped", rule.RuleType),
			})
		}
	}

	return results
}

// ============================================================
// GAP ANALYSIS & COLLECTION SCHEDULE
// ============================================================

// GetEvidenceGaps returns an analysis of evidence collection gaps for the organization.
func (s *EvidenceTemplateService) GetEvidenceGaps(ctx context.Context, orgID uuid.UUID) (*EvidenceGapsResult, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.setRLS(ctx, tx, orgID); err != nil {
		return nil, fmt.Errorf("set rls: %w", err)
	}

	gaps := &EvidenceGapsResult{}

	// Overall counts
	_ = tx.QueryRow(ctx, `SELECT COUNT(*) FROM evidence_requirements WHERE organization_id = $1`, orgID).Scan(&gaps.TotalRequirements)
	_ = tx.QueryRow(ctx, `SELECT COUNT(*) FROM evidence_requirements WHERE organization_id = $1 AND status IN ('collected','validated')`, orgID).Scan(&gaps.Collected)
	_ = tx.QueryRow(ctx, `SELECT COUNT(*) FROM evidence_requirements WHERE organization_id = $1 AND status = 'pending'`, orgID).Scan(&gaps.Pending)
	_ = tx.QueryRow(ctx, `SELECT COUNT(*) FROM evidence_requirements WHERE organization_id = $1 AND status = 'overdue'`, orgID).Scan(&gaps.Overdue)
	_ = tx.QueryRow(ctx, `SELECT COUNT(*) FROM evidence_requirements WHERE organization_id = $1 AND validation_status = 'pass'`, orgID).Scan(&gaps.Validated)
	_ = tx.QueryRow(ctx, `SELECT COUNT(*) FROM evidence_requirements WHERE organization_id = $1 AND validation_status = 'fail'`, orgID).Scan(&gaps.Failed)

	// Also count items where due_date has passed as overdue
	_ = tx.QueryRow(ctx, `
		SELECT COUNT(*) FROM evidence_requirements
		WHERE organization_id = $1 AND due_date < CURRENT_DATE AND status NOT IN ('collected','validated','waived','not_applicable')`,
		orgID).Scan(&gaps.Overdue)

	if gaps.TotalRequirements > 0 {
		gaps.CoveragePercent = float64(gaps.Collected) / float64(gaps.TotalRequirements) * 100
	}

	// Gaps by framework
	fwRows, err := tx.Query(ctx, `
		SELECT et.framework_code,
		       COUNT(*) AS total_required,
		       COUNT(*) FILTER (WHERE er.status IN ('collected','validated')) AS collected
		FROM evidence_requirements er
		JOIN evidence_templates et ON er.evidence_template_id = et.id
		WHERE er.organization_id = $1
		GROUP BY et.framework_code
		ORDER BY et.framework_code`, orgID)
	if err == nil {
		defer fwRows.Close()
		for fwRows.Next() {
			var fg FrameworkGap
			if fwRows.Scan(&fg.FrameworkCode, &fg.TotalRequired, &fg.Collected) == nil {
				if fg.TotalRequired > 0 {
					fg.CoveragePercent = float64(fg.Collected) / float64(fg.TotalRequired) * 100
				}
				gaps.GapsByFramework = append(gaps.GapsByFramework, fg)
			}
		}
	}
	if gaps.GapsByFramework == nil {
		gaps.GapsByFramework = []FrameworkGap{}
	}

	// Critical gaps (high/critical priority that are pending or overdue)
	critRows, err := tx.Query(ctx, `
		SELECT er.id, COALESCE(et.name, ''), COALESCE(et.framework_control_code, ''),
		       COALESCE(et.framework_code, ''), COALESCE(et.auditor_priority::text, 'medium'),
		       er.due_date
		FROM evidence_requirements er
		JOIN evidence_templates et ON er.evidence_template_id = et.id
		WHERE er.organization_id = $1
		  AND et.auditor_priority IN ('critical','high')
		  AND er.status IN ('pending','overdue','in_progress')
		ORDER BY et.auditor_priority ASC, er.due_date ASC NULLS LAST
		LIMIT 20`, orgID)
	if err == nil {
		defer critRows.Close()
		for critRows.Next() {
			var cg CriticalGapItem
			if critRows.Scan(&cg.RequirementID, &cg.TemplateName, &cg.FrameworkControlCode,
				&cg.FrameworkCode, &cg.AuditorPriority, &cg.DueDate) == nil {
				if cg.DueDate != nil && cg.DueDate.Before(time.Now()) {
					cg.DaysOverdue = int(time.Since(*cg.DueDate).Hours() / 24)
				}
				gaps.CriticalGaps = append(gaps.CriticalGaps, cg)
			}
		}
	}
	if gaps.CriticalGaps == nil {
		gaps.CriticalGaps = []CriticalGapItem{}
	}

	// Overdue items
	overdueRows, err := tx.Query(ctx, `
		SELECT er.id, COALESCE(et.name, ''), COALESCE(et.framework_control_code, ''),
		       er.due_date, er.assigned_to
		FROM evidence_requirements er
		LEFT JOIN evidence_templates et ON er.evidence_template_id = et.id
		WHERE er.organization_id = $1
		  AND er.due_date < CURRENT_DATE
		  AND er.status NOT IN ('collected','validated','waived','not_applicable')
		ORDER BY er.due_date ASC
		LIMIT 50`, orgID)
	if err == nil {
		defer overdueRows.Close()
		for overdueRows.Next() {
			var oi OverdueItem
			if overdueRows.Scan(&oi.RequirementID, &oi.TemplateName, &oi.FrameworkControlCode,
				&oi.DueDate, &oi.AssignedTo) == nil {
				if oi.DueDate != nil {
					oi.DaysOverdue = int(time.Since(*oi.DueDate).Hours() / 24)
				}
				gaps.OverdueItems = append(gaps.OverdueItems, oi)
			}
		}
	}
	if gaps.OverdueItems == nil {
		gaps.OverdueItems = []OverdueItem{}
	}

	return gaps, nil
}

// GetCollectionSchedule returns upcoming evidence collection tasks organized by time period.
func (s *EvidenceTemplateService) GetCollectionSchedule(ctx context.Context, orgID uuid.UUID) (*CollectionSchedule, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.setRLS(ctx, tx, orgID); err != nil {
		return nil, fmt.Errorf("set rls: %w", err)
	}

	schedule := &CollectionSchedule{}

	rows, err := tx.Query(ctx, `
		SELECT er.id, COALESCE(et.name, ''), COALESCE(et.framework_control_code, ''),
		       COALESCE(et.framework_code, ''),
		       COALESCE(er.collection_frequency_override::text, et.collection_frequency::text, 'annually') AS freq,
		       er.next_collection_due, er.assigned_to,
		       COALESCE(et.difficulty::text, 'moderate'),
		       COALESCE(et.typical_collection_time_minutes, 30)
		FROM evidence_requirements er
		LEFT JOIN evidence_templates et ON er.evidence_template_id = et.id
		WHERE er.organization_id = $1
		  AND er.status NOT IN ('waived','not_applicable')
		  AND er.next_collection_due IS NOT NULL
		  AND er.next_collection_due <= CURRENT_DATE + INTERVAL '90 days'
		ORDER BY er.next_collection_due ASC`, orgID)
	if err != nil {
		return nil, fmt.Errorf("query schedule: %w", err)
	}
	defer rows.Close()

	now := time.Now()
	weekEnd := now.AddDate(0, 0, 7)
	monthEnd := now.AddDate(0, 1, 0)

	for rows.Next() {
		var item ScheduleItem
		if err := rows.Scan(
			&item.RequirementID, &item.TemplateName, &item.FrameworkControlCode,
			&item.FrameworkCode, &item.CollectionFrequency, &item.NextDue,
			&item.AssignedTo, &item.Difficulty, &item.EstimatedMinutes,
		); err != nil {
			continue
		}

		schedule.TotalScheduled++

		if item.NextDue != nil {
			if item.NextDue.Before(weekEnd) {
				schedule.UpcomingThisWeek = append(schedule.UpcomingThisWeek, item)
			} else if item.NextDue.Before(monthEnd) {
				schedule.UpcomingThisMonth = append(schedule.UpcomingThisMonth, item)
			} else {
				schedule.UpcomingThisQuarter = append(schedule.UpcomingThisQuarter, item)
			}
		}
	}

	if schedule.UpcomingThisWeek == nil {
		schedule.UpcomingThisWeek = []ScheduleItem{}
	}
	if schedule.UpcomingThisMonth == nil {
		schedule.UpcomingThisMonth = []ScheduleItem{}
	}
	if schedule.UpcomingThisQuarter == nil {
		schedule.UpcomingThisQuarter = []ScheduleItem{}
	}

	return schedule, nil
}

// ============================================================
// HELPERS
// ============================================================

// scanTemplates scans pgx rows into a slice of EvidenceTemplate.
func scanTemplates(rows pgx.Rows) ([]EvidenceTemplate, error) {
	var templates []EvidenceTemplate
	for rows.Next() {
		var t EvidenceTemplate
		if err := rows.Scan(
			&t.ID, &t.OrganizationID, &t.FrameworkControlCode, &t.FrameworkCode,
			&t.Name, &t.Description, &t.EvidenceCategory, &t.CollectionMethod,
			&t.CollectionInstructions, &t.CollectionFrequency,
			&t.TypicalCollectionTimeMinutes, &t.ValidationRules,
			&t.AcceptanceCriteria, &t.CommonRejectionReasons,
			&t.TemplateFields, &t.SampleEvidenceDescription, &t.SampleFilePath,
			&t.ApplicableTo, &t.Difficulty, &t.AuditorPriority, &t.IsSystem, &t.Tags,
			&t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan template: %w", err)
		}
		templates = append(templates, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate templates: %w", err)
	}
	if templates == nil {
		templates = []EvidenceTemplate{}
	}
	return templates, nil
}

// calculateNextDue computes the next collection due date from a frequency string.
func calculateNextDue(frequency string) time.Time {
	now := time.Now()
	switch frequency {
	case "daily":
		return now.AddDate(0, 0, 1)
	case "weekly":
		return now.AddDate(0, 0, 7)
	case "monthly":
		return now.AddDate(0, 1, 0)
	case "quarterly":
		return now.AddDate(0, 3, 0)
	case "semi_annually":
		return now.AddDate(0, 6, 0)
	case "annually":
		return now.AddDate(1, 0, 0)
	case "once":
		return now.AddDate(0, 1, 0)
	default:
		return now.AddDate(0, 3, 0)
	}
}

// containsFileType checks if a file matches allowed types.
func containsFileType(mimeType, fileName, allowedTypes string) bool {
	allowed := strings.Split(strings.ToLower(allowedTypes), ",")
	mLower := strings.ToLower(mimeType)
	fLower := strings.ToLower(fileName)

	for _, a := range allowed {
		a = strings.TrimSpace(a)
		if a == "" {
			continue
		}
		// Check MIME type match
		if strings.Contains(mLower, a) {
			return true
		}
		// Check extension match
		if strings.HasSuffix(fLower, "."+a) {
			return true
		}
	}
	return false
}
