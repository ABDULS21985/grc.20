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

// AssessmentQuestionnaire is the master questionnaire template.
type AssessmentQuestionnaire struct {
	ID                        uuid.UUID       `json:"id"`
	OrganizationID            *uuid.UUID      `json:"organization_id"`
	Name                      string          `json:"name"`
	Description               string          `json:"description"`
	QuestionnaireType         string          `json:"questionnaire_type"`
	Version                   int             `json:"version"`
	Status                    string          `json:"status"`
	TotalQuestions            int             `json:"total_questions"`
	TotalSections             int             `json:"total_sections"`
	EstimatedCompletionMinutes int            `json:"estimated_completion_minutes"`
	ScoringMethod             string          `json:"scoring_method"`
	PassThreshold             float64         `json:"pass_threshold"`
	RiskTierThresholds        json.RawMessage `json:"risk_tier_thresholds"`
	ApplicableVendorTiers     []string        `json:"applicable_vendor_tiers"`
	IsSystem                  bool            `json:"is_system"`
	IsTemplate                bool            `json:"is_template"`
	CreatedBy                 *uuid.UUID      `json:"created_by"`
	CreatedAt                 time.Time       `json:"created_at"`
	UpdatedAt                 time.Time       `json:"updated_at"`
	// Nested data (populated when fetching full detail)
	Sections []QuestionnaireSection `json:"sections,omitempty"`
}

// QuestionnaireSection groups questions within a questionnaire.
type QuestionnaireSection struct {
	ID                 uuid.UUID               `json:"id"`
	QuestionnaireID    uuid.UUID               `json:"questionnaire_id"`
	Name               string                  `json:"name"`
	Description        string                  `json:"description"`
	SortOrder          int                     `json:"sort_order"`
	Weight             float64                 `json:"weight"`
	FrameworkDomainCode string                 `json:"framework_domain_code"`
	CreatedAt          time.Time               `json:"created_at"`
	Questions          []QuestionnaireQuestion `json:"questions,omitempty"`
}

// QuestionnaireQuestion is a single question within a section.
type QuestionnaireQuestion struct {
	ID                uuid.UUID       `json:"id"`
	SectionID         uuid.UUID       `json:"section_id"`
	QuestionText      string          `json:"question_text"`
	QuestionType      string          `json:"question_type"`
	Options           json.RawMessage `json:"options"`
	IsRequired        bool            `json:"is_required"`
	Weight            float64         `json:"weight"`
	RiskImpact        string          `json:"risk_impact"`
	GuidanceText      string          `json:"guidance_text"`
	EvidenceRequired  bool            `json:"evidence_required"`
	EvidenceGuidance  string          `json:"evidence_guidance"`
	MappedControlCodes []string       `json:"mapped_control_codes"`
	ConditionalOn     json.RawMessage `json:"conditional_on"`
	SortOrder         int             `json:"sort_order"`
	Tags              []string        `json:"tags"`
	CreatedAt         time.Time       `json:"created_at"`
}

// ScoreResult contains the computed score and breakdown for an assessment.
type ScoreResult struct {
	AssessmentID     uuid.UUID               `json:"assessment_id"`
	OverallScore     float64                 `json:"overall_score"`
	RiskRating       string                  `json:"risk_rating"`
	PassFail         string                  `json:"pass_fail"`
	CriticalFindings int                     `json:"critical_findings"`
	HighFindings     int                     `json:"high_findings"`
	SectionScores    map[string]SectionScore `json:"section_scores"`
}

// SectionScore holds the computed score for a single section.
type SectionScore struct {
	SectionID   uuid.UUID `json:"section_id"`
	SectionName string    `json:"section_name"`
	Score       float64   `json:"score"`
	Weight      float64   `json:"weight"`
	Answered    int       `json:"answered"`
	Total       int       `json:"total"`
}

// VendorComparison provides a side-by-side comparison of multiple vendor assessments.
type VendorComparison struct {
	Assessments []ComparisonEntry `json:"assessments"`
	BestScore   float64           `json:"best_score"`
	WorstScore  float64           `json:"worst_score"`
	AvgScore    float64           `json:"avg_score"`
}

// ComparisonEntry is a single entry in a vendor comparison.
type ComparisonEntry struct {
	AssessmentID   uuid.UUID               `json:"assessment_id"`
	AssessmentRef  string                  `json:"assessment_ref"`
	VendorID       uuid.UUID               `json:"vendor_id"`
	VendorName     string                  `json:"vendor_name"`
	OverallScore   float64                 `json:"overall_score"`
	RiskRating     string                  `json:"risk_rating"`
	PassFail       string                  `json:"pass_fail"`
	SectionScores  map[string]SectionScore `json:"section_scores,omitempty"`
	CompletedAt    *time.Time              `json:"completed_at"`
}

// ============================================================
// REQUEST / FILTER TYPES
// ============================================================

// CreateQuestionnaireRequest is the input for creating a questionnaire.
type CreateQuestionnaireRequest struct {
	Name                      string   `json:"name"`
	Description               string   `json:"description"`
	QuestionnaireType         string   `json:"questionnaire_type"`
	ScoringMethod             string   `json:"scoring_method"`
	PassThreshold             float64  `json:"pass_threshold"`
	ApplicableVendorTiers     []string `json:"applicable_vendor_tiers"`
	EstimatedCompletionMinutes int     `json:"estimated_completion_minutes"`
	Sections                  []CreateSectionRequest `json:"sections"`
}

// CreateSectionRequest is the input for creating a section within a questionnaire.
type CreateSectionRequest struct {
	Name                string                   `json:"name"`
	Description         string                   `json:"description"`
	Weight              float64                  `json:"weight"`
	FrameworkDomainCode string                   `json:"framework_domain_code"`
	Questions           []CreateQuestionRequest  `json:"questions"`
}

// CreateQuestionRequest is the input for creating a question within a section.
type CreateQuestionRequest struct {
	QuestionText      string          `json:"question_text"`
	QuestionType      string          `json:"question_type"`
	Options           json.RawMessage `json:"options"`
	IsRequired        bool            `json:"is_required"`
	Weight            float64         `json:"weight"`
	RiskImpact        string          `json:"risk_impact"`
	GuidanceText      string          `json:"guidance_text"`
	EvidenceRequired  bool            `json:"evidence_required"`
	EvidenceGuidance  string          `json:"evidence_guidance"`
	MappedControlCodes []string       `json:"mapped_control_codes"`
	Tags              []string        `json:"tags"`
}

// UpdateQuestionnaireRequest is the input for updating questionnaire metadata.
type UpdateQuestionnaireRequest struct {
	Name                      *string  `json:"name"`
	Description               *string  `json:"description"`
	Status                    *string  `json:"status"`
	ScoringMethod             *string  `json:"scoring_method"`
	PassThreshold             *float64 `json:"pass_threshold"`
	ApplicableVendorTiers     []string `json:"applicable_vendor_tiers"`
	EstimatedCompletionMinutes *int    `json:"estimated_completion_minutes"`
}

// QuestionnaireFilter holds filters for listing questionnaires.
type QuestionnaireFilter struct {
	QuestionnaireType string `json:"questionnaire_type"`
	Status            string `json:"status"`
	IsTemplate        bool   `json:"is_template"`
	Search            string `json:"search"`
	Page              int    `json:"page"`
	PageSize          int    `json:"page_size"`
}

// ============================================================
// SERVICE
// ============================================================

// QuestionnaireService provides business logic for TPRM questionnaire management.
type QuestionnaireService struct {
	pool *pgxpool.Pool
}

// NewQuestionnaireService creates a new QuestionnaireService.
func NewQuestionnaireService(pool *pgxpool.Pool) *QuestionnaireService {
	return &QuestionnaireService{pool: pool}
}

// ============================================================
// CREATE QUESTIONNAIRE
// ============================================================

// CreateQuestionnaire creates a new questionnaire with sections and questions.
func (s *QuestionnaireService) CreateQuestionnaire(ctx context.Context, orgID uuid.UUID, req CreateQuestionnaireRequest) (*AssessmentQuestionnaire, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("questionnaire name is required")
	}
	if req.QuestionnaireType == "" {
		req.QuestionnaireType = "security"
	}
	if req.ScoringMethod == "" {
		req.ScoringMethod = "weighted_average"
	}
	if req.PassThreshold <= 0 {
		req.PassThreshold = 70.00
	}
	if req.ApplicableVendorTiers == nil {
		req.ApplicableVendorTiers = []string{}
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Set RLS context
	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return nil, fmt.Errorf("set org context: %w", err)
	}

	totalQuestions := 0
	for _, sec := range req.Sections {
		totalQuestions += len(sec.Questions)
	}

	riskTierThresholds := json.RawMessage(`{"critical":40,"high":55,"medium":70,"low":85}`)

	var q AssessmentQuestionnaire
	err = tx.QueryRow(ctx, `
		INSERT INTO assessment_questionnaires
			(organization_id, name, description, questionnaire_type, status,
			 total_questions, total_sections, estimated_completion_minutes,
			 scoring_method, pass_threshold, risk_tier_thresholds,
			 applicable_vendor_tiers, is_system, is_template, created_by)
		VALUES ($1, $2, $3, $4, 'draft', $5, $6, $7, $8, $9, $10, $11, false, true, $12)
		RETURNING id, organization_id, name, description, questionnaire_type,
		          version, status, total_questions, total_sections,
		          estimated_completion_minutes, scoring_method, pass_threshold,
		          risk_tier_thresholds, applicable_vendor_tiers, is_system, is_template,
		          created_by, created_at, updated_at`,
		orgID, req.Name, req.Description, req.QuestionnaireType,
		totalQuestions, len(req.Sections), req.EstimatedCompletionMinutes,
		req.ScoringMethod, req.PassThreshold, riskTierThresholds,
		req.ApplicableVendorTiers, contextUserID(ctx),
	).Scan(
		&q.ID, &q.OrganizationID, &q.Name, &q.Description, &q.QuestionnaireType,
		&q.Version, &q.Status, &q.TotalQuestions, &q.TotalSections,
		&q.EstimatedCompletionMinutes, &q.ScoringMethod, &q.PassThreshold,
		&q.RiskTierThresholds, &q.ApplicableVendorTiers, &q.IsSystem, &q.IsTemplate,
		&q.CreatedBy, &q.CreatedAt, &q.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert questionnaire: %w", err)
	}

	// Insert sections and questions
	for i, secReq := range req.Sections {
		secWeight := secReq.Weight
		if secWeight <= 0 {
			secWeight = 1.00
		}

		var secID uuid.UUID
		err = tx.QueryRow(ctx, `
			INSERT INTO questionnaire_sections
				(questionnaire_id, name, description, sort_order, weight, framework_domain_code)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id`,
			q.ID, secReq.Name, secReq.Description, i+1, secWeight, secReq.FrameworkDomainCode,
		).Scan(&secID)
		if err != nil {
			return nil, fmt.Errorf("insert section %d: %w", i+1, err)
		}

		for j, qReq := range secReq.Questions {
			qWeight := qReq.Weight
			if qWeight <= 0 {
				qWeight = 1.00
			}
			if qReq.RiskImpact == "" {
				qReq.RiskImpact = "medium"
			}
			if qReq.Options == nil {
				qReq.Options = json.RawMessage(`[]`)
			}
			if qReq.MappedControlCodes == nil {
				qReq.MappedControlCodes = []string{}
			}
			if qReq.Tags == nil {
				qReq.Tags = []string{}
			}

			_, err = tx.Exec(ctx, `
				INSERT INTO questionnaire_questions
					(section_id, question_text, question_type, options, is_required,
					 weight, risk_impact, guidance_text, evidence_required, evidence_guidance,
					 mapped_control_codes, sort_order, tags)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
				secID, qReq.QuestionText, qReq.QuestionType, qReq.Options,
				qReq.IsRequired, qWeight, qReq.RiskImpact, qReq.GuidanceText,
				qReq.EvidenceRequired, qReq.EvidenceGuidance,
				qReq.MappedControlCodes, j+1, qReq.Tags,
			)
			if err != nil {
				return nil, fmt.Errorf("insert question %d/%d: %w", i+1, j+1, err)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	log.Info().
		Str("questionnaire_id", q.ID.String()).
		Str("name", q.Name).
		Int("questions", totalQuestions).
		Msg("questionnaire created")

	return &q, nil
}

// ============================================================
// GET QUESTIONNAIRE (full detail with sections & questions)
// ============================================================

// GetQuestionnaire returns a questionnaire with its sections and questions.
func (s *QuestionnaireService) GetQuestionnaire(ctx context.Context, orgID uuid.UUID, qID uuid.UUID) (*AssessmentQuestionnaire, error) {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("acquire conn: %w", err)
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return nil, fmt.Errorf("set org context: %w", err)
	}

	var q AssessmentQuestionnaire
	err = conn.QueryRow(ctx, `
		SELECT id, organization_id, name, description, questionnaire_type,
		       version, status, total_questions, total_sections,
		       estimated_completion_minutes, scoring_method, pass_threshold,
		       risk_tier_thresholds, applicable_vendor_tiers, is_system, is_template,
		       created_by, created_at, updated_at
		FROM assessment_questionnaires
		WHERE id = $1`, qID).Scan(
		&q.ID, &q.OrganizationID, &q.Name, &q.Description, &q.QuestionnaireType,
		&q.Version, &q.Status, &q.TotalQuestions, &q.TotalSections,
		&q.EstimatedCompletionMinutes, &q.ScoringMethod, &q.PassThreshold,
		&q.RiskTierThresholds, &q.ApplicableVendorTiers, &q.IsSystem, &q.IsTemplate,
		&q.CreatedBy, &q.CreatedAt, &q.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("questionnaire not found")
		}
		return nil, fmt.Errorf("get questionnaire: %w", err)
	}

	// Fetch sections
	secRows, err := conn.Query(ctx, `
		SELECT id, questionnaire_id, name, description, sort_order, weight,
		       COALESCE(framework_domain_code, ''), created_at
		FROM questionnaire_sections
		WHERE questionnaire_id = $1
		ORDER BY sort_order`, qID)
	if err != nil {
		return nil, fmt.Errorf("query sections: %w", err)
	}
	defer secRows.Close()

	sectionMap := make(map[uuid.UUID]*QuestionnaireSection)
	for secRows.Next() {
		var sec QuestionnaireSection
		if err := secRows.Scan(
			&sec.ID, &sec.QuestionnaireID, &sec.Name, &sec.Description,
			&sec.SortOrder, &sec.Weight, &sec.FrameworkDomainCode, &sec.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan section: %w", err)
		}
		sec.Questions = []QuestionnaireQuestion{}
		q.Sections = append(q.Sections, sec)
		sectionMap[sec.ID] = &q.Sections[len(q.Sections)-1]
	}
	if err := secRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate sections: %w", err)
	}

	if len(q.Sections) == 0 {
		q.Sections = []QuestionnaireSection{}
		return &q, nil
	}

	// Collect section IDs for question query
	secIDs := make([]uuid.UUID, 0, len(q.Sections))
	for _, sec := range q.Sections {
		secIDs = append(secIDs, sec.ID)
	}

	// Fetch all questions for all sections in one query
	qRows, err := conn.Query(ctx, `
		SELECT id, section_id, question_text, question_type, options,
		       is_required, weight, risk_impact, guidance_text,
		       evidence_required, evidence_guidance, mapped_control_codes,
		       conditional_on, sort_order, tags, created_at
		FROM questionnaire_questions
		WHERE section_id = ANY($1)
		ORDER BY section_id, sort_order`, secIDs)
	if err != nil {
		return nil, fmt.Errorf("query questions: %w", err)
	}
	defer qRows.Close()

	for qRows.Next() {
		var qq QuestionnaireQuestion
		if err := qRows.Scan(
			&qq.ID, &qq.SectionID, &qq.QuestionText, &qq.QuestionType, &qq.Options,
			&qq.IsRequired, &qq.Weight, &qq.RiskImpact, &qq.GuidanceText,
			&qq.EvidenceRequired, &qq.EvidenceGuidance, &qq.MappedControlCodes,
			&qq.ConditionalOn, &qq.SortOrder, &qq.Tags, &qq.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan question: %w", err)
		}
		if sec, ok := sectionMap[qq.SectionID]; ok {
			sec.Questions = append(sec.Questions, qq)
		}
	}
	if err := qRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate questions: %w", err)
	}

	return &q, nil
}

// ============================================================
// LIST QUESTIONNAIRES
// ============================================================

// ListQuestionnaires returns filtered questionnaires visible to the organisation.
func (s *QuestionnaireService) ListQuestionnaires(ctx context.Context, orgID uuid.UUID, filter QuestionnaireFilter) ([]AssessmentQuestionnaire, int64, error) {
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

	// System templates are always visible; org templates filtered by RLS
	if filter.QuestionnaireType != "" {
		conditions = append(conditions, fmt.Sprintf("questionnaire_type = $%d", argIdx))
		args = append(args, filter.QuestionnaireType)
		argIdx++
	}
	if filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, filter.Status)
		argIdx++
	}
	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf(
			"(name ILIKE '%%' || $%d || '%%' OR description ILIKE '%%' || $%d || '%%')", argIdx, argIdx))
		args = append(args, filter.Search)
		argIdx++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count
	countSQL := fmt.Sprintf(`SELECT COUNT(*) FROM assessment_questionnaires %s`, where)
	var total int64
	if err := conn.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count questionnaires: %w", err)
	}

	// Data
	dataSQL := fmt.Sprintf(`
		SELECT id, organization_id, name, description, questionnaire_type,
		       version, status, total_questions, total_sections,
		       estimated_completion_minutes, scoring_method, pass_threshold,
		       risk_tier_thresholds, applicable_vendor_tiers, is_system, is_template,
		       created_by, created_at, updated_at
		FROM assessment_questionnaires
		%s
		ORDER BY is_system DESC, name ASC
		LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)
	args = append(args, filter.PageSize, offset)

	rows, err := conn.Query(ctx, dataSQL, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query questionnaires: %w", err)
	}
	defer rows.Close()

	var list []AssessmentQuestionnaire
	for rows.Next() {
		var q AssessmentQuestionnaire
		if err := rows.Scan(
			&q.ID, &q.OrganizationID, &q.Name, &q.Description, &q.QuestionnaireType,
			&q.Version, &q.Status, &q.TotalQuestions, &q.TotalSections,
			&q.EstimatedCompletionMinutes, &q.ScoringMethod, &q.PassThreshold,
			&q.RiskTierThresholds, &q.ApplicableVendorTiers, &q.IsSystem, &q.IsTemplate,
			&q.CreatedBy, &q.CreatedAt, &q.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan questionnaire: %w", err)
		}
		list = append(list, q)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate questionnaires: %w", err)
	}

	if list == nil {
		list = []AssessmentQuestionnaire{}
	}
	return list, total, nil
}

// ============================================================
// UPDATE QUESTIONNAIRE
// ============================================================

// UpdateQuestionnaire updates questionnaire metadata (not sections/questions).
func (s *QuestionnaireService) UpdateQuestionnaire(ctx context.Context, orgID, qID uuid.UUID, req UpdateQuestionnaireRequest) error {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire conn: %w", err)
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return fmt.Errorf("set org context: %w", err)
	}

	// Verify questionnaire exists and is not a system template
	var isSystem bool
	err = conn.QueryRow(ctx, `
		SELECT is_system FROM assessment_questionnaires WHERE id = $1`, qID).Scan(&isSystem)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("questionnaire not found")
		}
		return fmt.Errorf("check questionnaire: %w", err)
	}
	if isSystem {
		return fmt.Errorf("cannot modify system questionnaire templates")
	}

	// Build dynamic update
	var setClauses []string
	var args []interface{}
	argIdx := 1

	if req.Name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, *req.Name)
		argIdx++
	}
	if req.Description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", argIdx))
		args = append(args, *req.Description)
		argIdx++
	}
	if req.Status != nil {
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *req.Status)
		argIdx++
	}
	if req.ScoringMethod != nil {
		setClauses = append(setClauses, fmt.Sprintf("scoring_method = $%d", argIdx))
		args = append(args, *req.ScoringMethod)
		argIdx++
	}
	if req.PassThreshold != nil {
		setClauses = append(setClauses, fmt.Sprintf("pass_threshold = $%d", argIdx))
		args = append(args, *req.PassThreshold)
		argIdx++
	}
	if req.ApplicableVendorTiers != nil {
		setClauses = append(setClauses, fmt.Sprintf("applicable_vendor_tiers = $%d", argIdx))
		args = append(args, req.ApplicableVendorTiers)
		argIdx++
	}
	if req.EstimatedCompletionMinutes != nil {
		setClauses = append(setClauses, fmt.Sprintf("estimated_completion_minutes = $%d", argIdx))
		args = append(args, *req.EstimatedCompletionMinutes)
		argIdx++
	}

	if len(setClauses) == 0 {
		return nil // nothing to update
	}

	sql := fmt.Sprintf(`UPDATE assessment_questionnaires SET %s WHERE id = $%d`,
		strings.Join(setClauses, ", "), argIdx)
	args = append(args, qID)

	tag, err := conn.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("update questionnaire: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("questionnaire not found")
	}

	return nil
}

// ============================================================
// CLONE TEMPLATE
// ============================================================

// CloneTemplate creates an editable copy of a system or org template.
func (s *QuestionnaireService) CloneTemplate(ctx context.Context, orgID, templateID uuid.UUID) (*AssessmentQuestionnaire, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return nil, fmt.Errorf("set org context: %w", err)
	}

	// Verify template exists
	var origName string
	err = tx.QueryRow(ctx, `
		SELECT name FROM assessment_questionnaires WHERE id = $1`, templateID).Scan(&origName)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("template not found")
		}
		return nil, fmt.Errorf("check template: %w", err)
	}

	cloneName := origName + " (Copy)"

	// Clone the questionnaire
	var newID uuid.UUID
	err = tx.QueryRow(ctx, `
		INSERT INTO assessment_questionnaires
			(organization_id, name, description, questionnaire_type, version, status,
			 total_questions, total_sections, estimated_completion_minutes,
			 scoring_method, pass_threshold, risk_tier_thresholds,
			 applicable_vendor_tiers, is_system, is_template, created_by)
		SELECT $1, $2, description, questionnaire_type, 1, 'draft',
		       total_questions, total_sections, estimated_completion_minutes,
		       scoring_method, pass_threshold, risk_tier_thresholds,
		       applicable_vendor_tiers, false, true, $3
		FROM assessment_questionnaires WHERE id = $4
		RETURNING id`,
		orgID, cloneName, contextUserID(ctx), templateID,
	).Scan(&newID)
	if err != nil {
		return nil, fmt.Errorf("clone questionnaire: %w", err)
	}

	// Clone sections
	secRows, err := tx.Query(ctx, `
		SELECT id, name, description, sort_order, weight, framework_domain_code
		FROM questionnaire_sections
		WHERE questionnaire_id = $1
		ORDER BY sort_order`, templateID)
	if err != nil {
		return nil, fmt.Errorf("query template sections: %w", err)
	}

	type sectionPair struct {
		oldID uuid.UUID
		newID uuid.UUID
	}
	var sectionPairs []sectionPair

	for secRows.Next() {
		var (
			oldSecID            uuid.UUID
			name, desc          string
			sortOrder           int
			weight              float64
			frameworkDomainCode *string
		)
		if err := secRows.Scan(&oldSecID, &name, &desc, &sortOrder, &weight, &frameworkDomainCode); err != nil {
			secRows.Close()
			return nil, fmt.Errorf("scan template section: %w", err)
		}

		var newSecID uuid.UUID
		err = tx.QueryRow(ctx, `
			INSERT INTO questionnaire_sections
				(questionnaire_id, name, description, sort_order, weight, framework_domain_code)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id`,
			newID, name, desc, sortOrder, weight, frameworkDomainCode,
		).Scan(&newSecID)
		if err != nil {
			secRows.Close()
			return nil, fmt.Errorf("clone section: %w", err)
		}

		sectionPairs = append(sectionPairs, sectionPair{oldID: oldSecID, newID: newSecID})
	}
	secRows.Close()
	if err := secRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate template sections: %w", err)
	}

	// Clone questions for each section
	for _, pair := range sectionPairs {
		_, err = tx.Exec(ctx, `
			INSERT INTO questionnaire_questions
				(section_id, question_text, question_type, options, is_required,
				 weight, risk_impact, guidance_text, evidence_required, evidence_guidance,
				 mapped_control_codes, conditional_on, sort_order, tags)
			SELECT $1, question_text, question_type, options, is_required,
			       weight, risk_impact, guidance_text, evidence_required, evidence_guidance,
			       mapped_control_codes, conditional_on, sort_order, tags
			FROM questionnaire_questions
			WHERE section_id = $2
			ORDER BY sort_order`,
			pair.newID, pair.oldID,
		)
		if err != nil {
			return nil, fmt.Errorf("clone questions: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	log.Info().
		Str("new_id", newID.String()).
		Str("template_id", templateID.String()).
		Msg("questionnaire template cloned")

	// Return the cloned questionnaire
	return s.GetQuestionnaire(ctx, orgID, newID)
}

// ============================================================
// CALCULATE SCORE
// ============================================================

// CalculateScore computes the overall and per-section scores for an assessment.
func (s *QuestionnaireService) CalculateScore(ctx context.Context, assessmentID uuid.UUID) (*ScoreResult, error) {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("acquire conn: %w", err)
	}
	defer conn.Release()

	// Get assessment details to determine scoring method and thresholds
	var (
		scoringMethod  string
		passThreshold  float64
		thresholdsJSON json.RawMessage
		qID            uuid.UUID
		assessOrgID    uuid.UUID
	)
	err = conn.QueryRow(ctx, `
		SELECT va.id, va.organization_id, va.questionnaire_id,
		       aq.scoring_method, aq.pass_threshold, aq.risk_tier_thresholds
		FROM vendor_assessments va
		JOIN assessment_questionnaires aq ON va.questionnaire_id = aq.id
		WHERE va.id = $1`, assessmentID).Scan(
		&assessmentID, &assessOrgID, &qID,
		&scoringMethod, &passThreshold, &thresholdsJSON,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("assessment not found")
		}
		return nil, fmt.Errorf("get assessment: %w", err)
	}

	if _, err := conn.Exec(ctx, "SET LOCAL app.current_org = $1", assessOrgID.String()); err != nil {
		return nil, fmt.Errorf("set org context: %w", err)
	}

	// Parse risk tier thresholds
	var thresholds struct {
		Critical float64 `json:"critical"`
		High     float64 `json:"high"`
		Medium   float64 `json:"medium"`
		Low      float64 `json:"low"`
	}
	if err := json.Unmarshal(thresholdsJSON, &thresholds); err != nil {
		thresholds.Critical = 40
		thresholds.High = 55
		thresholds.Medium = 70
		thresholds.Low = 85
	}

	// Fetch sections with weights
	secRows, err := conn.Query(ctx, `
		SELECT qs.id, qs.name, qs.weight
		FROM questionnaire_sections qs
		WHERE qs.questionnaire_id = $1
		ORDER BY qs.sort_order`, qID)
	if err != nil {
		return nil, fmt.Errorf("query sections: %w", err)
	}
	defer secRows.Close()

	type sectionInfo struct {
		ID     uuid.UUID
		Name   string
		Weight float64
	}
	var sections []sectionInfo
	for secRows.Next() {
		var si sectionInfo
		if err := secRows.Scan(&si.ID, &si.Name, &si.Weight); err != nil {
			return nil, fmt.Errorf("scan section: %w", err)
		}
		sections = append(sections, si)
	}
	if err := secRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate sections: %w", err)
	}

	// Fetch all questions with responses for this assessment
	qRows, err := conn.Query(ctx, `
		SELECT qq.id, qq.section_id, qq.weight, qq.risk_impact,
		       var.answer_score, qq.options, qq.question_type
		FROM questionnaire_questions qq
		JOIN questionnaire_sections qs ON qq.section_id = qs.id
		LEFT JOIN vendor_assessment_responses var
			ON var.question_id = qq.id AND var.assessment_id = $1
		WHERE qs.questionnaire_id = $2
		ORDER BY qs.sort_order, qq.sort_order`, assessmentID, qID)
	if err != nil {
		return nil, fmt.Errorf("query questions: %w", err)
	}
	defer qRows.Close()

	// Accumulate scores per section
	type questionData struct {
		SectionID   uuid.UUID
		Weight      float64
		RiskImpact  string
		AnswerScore *float64
	}
	var questions []questionData
	for qRows.Next() {
		var qd questionData
		var optionsRaw json.RawMessage
		var qType string
		if err := qRows.Scan(
			new(uuid.UUID), &qd.SectionID, &qd.Weight, &qd.RiskImpact,
			&qd.AnswerScore, &optionsRaw, &qType,
		); err != nil {
			return nil, fmt.Errorf("scan question data: %w", err)
		}
		questions = append(questions, qd)
	}
	if err := qRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate question data: %w", err)
	}

	// Compute per-section scores
	sectionScores := make(map[string]SectionScore)
	sectionWeightedSum := make(map[uuid.UUID]float64)
	sectionWeightTotal := make(map[uuid.UUID]float64)
	sectionAnswered := make(map[uuid.UUID]int)
	sectionTotal := make(map[uuid.UUID]int)
	criticalFindings := 0
	highFindings := 0

	for _, q := range questions {
		sectionTotal[q.SectionID]++
		if q.AnswerScore != nil {
			sectionWeightedSum[q.SectionID] += (*q.AnswerScore) * q.Weight
			sectionWeightTotal[q.SectionID] += q.Weight
			sectionAnswered[q.SectionID]++

			// Count findings (score < 50 for critical/high impact questions)
			if *q.AnswerScore < 50.0 {
				switch q.RiskImpact {
				case "critical":
					criticalFindings++
				case "high":
					highFindings++
				}
			}
		}
	}

	// Map section IDs to info
	sectionInfoMap := make(map[uuid.UUID]sectionInfo)
	for _, si := range sections {
		sectionInfoMap[si.ID] = si
	}

	for _, si := range sections {
		score := 0.0
		if sectionWeightTotal[si.ID] > 0 {
			score = sectionWeightedSum[si.ID] / sectionWeightTotal[si.ID]
		}
		sectionScores[si.Name] = SectionScore{
			SectionID:   si.ID,
			SectionName: si.Name,
			Score:       score,
			Weight:      si.Weight,
			Answered:    sectionAnswered[si.ID],
			Total:       sectionTotal[si.ID],
		}
	}

	// Compute overall score based on scoring method
	var overallScore float64
	switch scoringMethod {
	case "weighted_average":
		totalWeight := 0.0
		weightedSum := 0.0
		for _, ss := range sectionScores {
			if ss.Answered > 0 {
				weightedSum += ss.Score * ss.Weight
				totalWeight += ss.Weight
			}
		}
		if totalWeight > 0 {
			overallScore = weightedSum / totalWeight
		}
	case "simple_average":
		sum := 0.0
		count := 0
		for _, ss := range sectionScores {
			if ss.Answered > 0 {
				sum += ss.Score
				count++
			}
		}
		if count > 0 {
			overallScore = sum / float64(count)
		}
	case "minimum_score":
		overallScore = 100.0
		for _, ss := range sectionScores {
			if ss.Answered > 0 && ss.Score < overallScore {
				overallScore = ss.Score
			}
		}
		if len(sectionScores) == 0 {
			overallScore = 0
		}
	default:
		// Default to weighted average
		totalWeight := 0.0
		weightedSum := 0.0
		for _, ss := range sectionScores {
			if ss.Answered > 0 {
				weightedSum += ss.Score * ss.Weight
				totalWeight += ss.Weight
			}
		}
		if totalWeight > 0 {
			overallScore = weightedSum / totalWeight
		}
	}

	// Determine risk rating from thresholds
	riskRating := "critical"
	if overallScore >= thresholds.Low {
		riskRating = "low"
	} else if overallScore >= thresholds.Medium {
		riskRating = "medium"
	} else if overallScore >= thresholds.High {
		riskRating = "high"
	}

	// Determine pass/fail
	passFail := "fail"
	if overallScore >= passThreshold {
		passFail = "pass"
	} else if overallScore >= passThreshold*0.9 && criticalFindings == 0 {
		passFail = "conditional_pass"
	}

	result := &ScoreResult{
		AssessmentID:     assessmentID,
		OverallScore:     overallScore,
		RiskRating:       riskRating,
		PassFail:         passFail,
		CriticalFindings: criticalFindings,
		HighFindings:     highFindings,
		SectionScores:    sectionScores,
	}

	// Store computed scores on the assessment
	sectionScoresJSON, _ := json.Marshal(sectionScores)
	_, err = conn.Exec(ctx, `
		UPDATE vendor_assessments SET
			overall_score = $1, risk_rating = $2, pass_fail = $3,
			critical_findings = $4, high_findings = $5, section_scores = $6
		WHERE id = $7`,
		overallScore, riskRating, passFail,
		criticalFindings, highFindings, sectionScoresJSON,
		assessmentID)
	if err != nil {
		log.Warn().Err(err).Msg("failed to persist score on assessment")
	}

	return result, nil
}

// ============================================================
// COMPARE VENDORS
// ============================================================

// CompareVendors provides a side-by-side comparison of multiple assessments.
func (s *QuestionnaireService) CompareVendors(ctx context.Context, orgID uuid.UUID, assessmentIDs []uuid.UUID) (*VendorComparison, error) {
	if len(assessmentIDs) == 0 {
		return nil, fmt.Errorf("at least one assessment ID is required")
	}
	if len(assessmentIDs) > 10 {
		return nil, fmt.Errorf("maximum 10 assessments can be compared")
	}

	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("acquire conn: %w", err)
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return nil, fmt.Errorf("set org context: %w", err)
	}

	rows, err := conn.Query(ctx, `
		SELECT va.id, va.assessment_ref, va.vendor_id,
		       COALESCE(v.name, '') AS vendor_name,
		       COALESCE(va.overall_score, 0), COALESCE(va.risk_rating, 'pending'),
		       va.pass_fail, va.section_scores, va.reviewed_at
		FROM vendor_assessments va
		LEFT JOIN vendors v ON va.vendor_id = v.id
		WHERE va.id = ANY($1) AND va.organization_id = $2
		ORDER BY va.overall_score DESC NULLS LAST`, assessmentIDs, orgID)
	if err != nil {
		return nil, fmt.Errorf("query assessments: %w", err)
	}
	defer rows.Close()

	comparison := &VendorComparison{
		BestScore:  0,
		WorstScore: 100,
	}
	totalScore := 0.0

	for rows.Next() {
		var entry ComparisonEntry
		var sectionScoresRaw json.RawMessage
		if err := rows.Scan(
			&entry.AssessmentID, &entry.AssessmentRef, &entry.VendorID,
			&entry.VendorName, &entry.OverallScore, &entry.RiskRating,
			&entry.PassFail, &sectionScoresRaw, &entry.CompletedAt,
		); err != nil {
			return nil, fmt.Errorf("scan comparison entry: %w", err)
		}

		// Parse section scores
		if len(sectionScoresRaw) > 2 {
			entry.SectionScores = make(map[string]SectionScore)
			_ = json.Unmarshal(sectionScoresRaw, &entry.SectionScores)
		}

		if entry.OverallScore > comparison.BestScore {
			comparison.BestScore = entry.OverallScore
		}
		if entry.OverallScore < comparison.WorstScore {
			comparison.WorstScore = entry.OverallScore
		}
		totalScore += entry.OverallScore

		comparison.Assessments = append(comparison.Assessments, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate comparison: %w", err)
	}

	if len(comparison.Assessments) > 0 {
		comparison.AvgScore = totalScore / float64(len(comparison.Assessments))
	}
	if comparison.Assessments == nil {
		comparison.Assessments = []ComparisonEntry{}
	}

	return comparison, nil
}
