package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// ============================================================
// ContinuityService
// ============================================================

// ContinuityService implements business logic for Business Continuity management,
// covering BIA scenarios, continuity plans, BC exercises, and the BC dashboard.
type ContinuityService struct {
	pool *pgxpool.Pool
}

// NewContinuityService creates a new ContinuityService with the given database pool.
func NewContinuityService(pool *pgxpool.Pool) *ContinuityService {
	return &ContinuityService{pool: pool}
}

// ============================================================
// DATA TYPES
// ============================================================

// BIAScenario represents a disruption scenario for impact analysis.
type BIAScenario struct {
	ID                      uuid.UUID       `json:"id"`
	OrganizationID          uuid.UUID       `json:"organization_id"`
	ScenarioRef             string          `json:"scenario_ref"`
	Name                    string          `json:"name"`
	Description             string          `json:"description"`
	ScenarioType            string          `json:"scenario_type"`
	Likelihood              string          `json:"likelihood"`
	AffectedProcessIDs      []uuid.UUID     `json:"affected_process_ids"`
	AffectedAssetIDs        []uuid.UUID     `json:"affected_asset_ids"`
	ImpactTimeline          json.RawMessage `json:"impact_timeline"`
	EstimatedFinancialLoss  *float64        `json:"estimated_financial_loss_eur"`
	MitigationStrategy      string          `json:"mitigation_strategy"`
	Status                  string          `json:"status"`
	CreatedAt               time.Time       `json:"created_at"`
	UpdatedAt               time.Time       `json:"updated_at"`
}

// ContinuityPlan represents a business continuity or disaster recovery plan.
type ContinuityPlan struct {
	ID                    uuid.UUID       `json:"id"`
	OrganizationID        uuid.UUID       `json:"organization_id"`
	PlanRef               string          `json:"plan_ref"`
	Name                  string          `json:"name"`
	PlanType              string          `json:"plan_type"`
	Status                string          `json:"status"`
	Version               int             `json:"version"`
	ScopeDescription      string          `json:"scope_description"`
	CoveredScenarioIDs    []uuid.UUID     `json:"covered_scenario_ids"`
	CoveredProcessIDs     []uuid.UUID     `json:"covered_process_ids"`
	ActivationCriteria    string          `json:"activation_criteria"`
	ActivationAuthority   string          `json:"activation_authority"`
	CommandStructure      json.RawMessage `json:"command_structure"`
	CommunicationPlan     json.RawMessage `json:"communication_plan"`
	RecoveryProcedures    json.RawMessage `json:"recovery_procedures"`
	ResourceRequirements  json.RawMessage `json:"resource_requirements"`
	AlternateSiteDetails  json.RawMessage `json:"alternate_site_details"`
	OwnerUserID           *uuid.UUID      `json:"owner_user_id"`
	ApprovedBy            *uuid.UUID      `json:"approved_by"`
	ApprovedAt            *time.Time      `json:"approved_at"`
	NextReviewDate        *time.Time      `json:"next_review_date"`
	ReviewFrequencyMonths int             `json:"review_frequency_months"`
	DocumentPath          string          `json:"document_path"`
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
}

// BCExercise represents a business continuity exercise or test.
type BCExercise struct {
	ID                 uuid.UUID       `json:"id"`
	OrganizationID     uuid.UUID       `json:"organization_id"`
	ExerciseRef        string          `json:"exercise_ref"`
	Name               string          `json:"name"`
	ExerciseType       string          `json:"exercise_type"`
	PlanID             *uuid.UUID      `json:"plan_id"`
	ScenarioID         *uuid.UUID      `json:"scenario_id"`
	Status             string          `json:"status"`
	ScheduledDate      *time.Time      `json:"scheduled_date"`
	ActualDate         *time.Time      `json:"actual_date"`
	Participants       json.RawMessage `json:"participants"`
	RTOAchievedHours   *float64        `json:"rto_achieved_hours"`
	RPOAchievedHours   *float64        `json:"rpo_achieved_hours"`
	ObjectivesMet      *bool           `json:"objectives_met"`
	LessonsLearned     string          `json:"lessons_learned"`
	GapsIdentified     string          `json:"gaps_identified"`
	ImprovementActions json.RawMessage `json:"improvement_actions"`
	OverallRating      *string         `json:"overall_rating"`
	ReportDocumentPath string          `json:"report_document_path"`
	ConductedBy        *uuid.UUID      `json:"conducted_by"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}

// BCDashboard aggregates business continuity metrics for the organisation.
type BCDashboard struct {
	ProcessesWithoutBIA     int        `json:"processes_without_bia"`
	PlansRequiringReview    int        `json:"plans_requiring_review"`
	LastExerciseDate        *time.Time `json:"last_exercise_date"`
	LastExerciseResult      string     `json:"last_exercise_result"`
	RTOCoveragePercent      float64    `json:"rto_coverage_percent"`
	RPOCoveragePercent      float64    `json:"rpo_coverage_percent"`
	SPOFCount               int        `json:"spof_count"`
	TotalProcesses          int        `json:"total_processes"`
	TotalScenarios          int        `json:"total_scenarios"`
	TotalPlans              int        `json:"total_plans"`
	TotalExercises          int        `json:"total_exercises"`
	ActivePlans             int        `json:"active_plans"`
	DraftPlans              int        `json:"draft_plans"`
	MissionCriticalCount    int        `json:"mission_critical_count"`
	BusinessCriticalCount   int        `json:"business_critical_count"`
	PlannedExercises        int        `json:"planned_exercises"`
	CompletedExercises      int        `json:"completed_exercises"`
}

// ============================================================
// REQUEST TYPES
// ============================================================

// CreateScenarioReq is the request body for creating a scenario.
type CreateScenarioReq struct {
	Name                    string      `json:"name"`
	Description             string      `json:"description"`
	ScenarioType            string      `json:"scenario_type"`
	Likelihood              string      `json:"likelihood"`
	AffectedProcessIDs      []uuid.UUID `json:"affected_process_ids"`
	AffectedAssetIDs        []uuid.UUID `json:"affected_asset_ids"`
	ImpactTimeline          json.RawMessage `json:"impact_timeline"`
	EstimatedFinancialLoss  *float64    `json:"estimated_financial_loss_eur"`
	MitigationStrategy      string      `json:"mitigation_strategy"`
}

// UpdateScenarioReq is the request body for updating a scenario.
type UpdateScenarioReq struct {
	Name                    *string         `json:"name"`
	Description             *string         `json:"description"`
	ScenarioType            *string         `json:"scenario_type"`
	Likelihood              *string         `json:"likelihood"`
	AffectedProcessIDs      []uuid.UUID     `json:"affected_process_ids"`
	AffectedAssetIDs        []uuid.UUID     `json:"affected_asset_ids"`
	ImpactTimeline          json.RawMessage `json:"impact_timeline"`
	EstimatedFinancialLoss  *float64        `json:"estimated_financial_loss_eur"`
	MitigationStrategy      *string         `json:"mitigation_strategy"`
	Status                  *string         `json:"status"`
}

// CreatePlanReq is the request body for creating a continuity plan.
type CreatePlanReq struct {
	Name                  string          `json:"name"`
	PlanType              string          `json:"plan_type"`
	ScopeDescription      string          `json:"scope_description"`
	CoveredScenarioIDs    []uuid.UUID     `json:"covered_scenario_ids"`
	CoveredProcessIDs     []uuid.UUID     `json:"covered_process_ids"`
	ActivationCriteria    string          `json:"activation_criteria"`
	ActivationAuthority   string          `json:"activation_authority"`
	CommandStructure      json.RawMessage `json:"command_structure"`
	CommunicationPlan     json.RawMessage `json:"communication_plan"`
	RecoveryProcedures    json.RawMessage `json:"recovery_procedures"`
	ResourceRequirements  json.RawMessage `json:"resource_requirements"`
	AlternateSiteDetails  json.RawMessage `json:"alternate_site_details"`
	OwnerUserID           *uuid.UUID      `json:"owner_user_id"`
	ReviewFrequencyMonths int             `json:"review_frequency_months"`
	DocumentPath          string          `json:"document_path"`
}

// UpdatePlanReq is the request body for updating a continuity plan.
type UpdatePlanReq struct {
	Name                  *string         `json:"name"`
	PlanType              *string         `json:"plan_type"`
	Status                *string         `json:"status"`
	ScopeDescription      *string         `json:"scope_description"`
	CoveredScenarioIDs    []uuid.UUID     `json:"covered_scenario_ids"`
	CoveredProcessIDs     []uuid.UUID     `json:"covered_process_ids"`
	ActivationCriteria    *string         `json:"activation_criteria"`
	ActivationAuthority   *string         `json:"activation_authority"`
	CommandStructure      json.RawMessage `json:"command_structure"`
	CommunicationPlan     json.RawMessage `json:"communication_plan"`
	RecoveryProcedures    json.RawMessage `json:"recovery_procedures"`
	ResourceRequirements  json.RawMessage `json:"resource_requirements"`
	AlternateSiteDetails  json.RawMessage `json:"alternate_site_details"`
	OwnerUserID           *uuid.UUID      `json:"owner_user_id"`
	ReviewFrequencyMonths *int            `json:"review_frequency_months"`
	DocumentPath          *string         `json:"document_path"`
}

// CreateExerciseReq is the request body for scheduling a BC exercise.
type CreateExerciseReq struct {
	Name          string     `json:"name"`
	ExerciseType  string     `json:"exercise_type"`
	PlanID        *uuid.UUID `json:"plan_id"`
	ScenarioID    *uuid.UUID `json:"scenario_id"`
	ScheduledDate string     `json:"scheduled_date"`
	Participants  json.RawMessage `json:"participants"`
	ConductedBy   *uuid.UUID `json:"conducted_by"`
}

// CompleteExerciseReq is the request body for recording exercise results.
type CompleteExerciseReq struct {
	ActualDate         string          `json:"actual_date"`
	RTOAchievedHours   *float64        `json:"rto_achieved_hours"`
	RPOAchievedHours   *float64        `json:"rpo_achieved_hours"`
	ObjectivesMet      *bool           `json:"objectives_met"`
	LessonsLearned     string          `json:"lessons_learned"`
	GapsIdentified     string          `json:"gaps_identified"`
	ImprovementActions json.RawMessage `json:"improvement_actions"`
	OverallRating      *string         `json:"overall_rating"`
	ReportDocumentPath string          `json:"report_document_path"`
}

// ============================================================
// SCENARIOS
// ============================================================

// CreateScenario creates a new BIA scenario with auto-generated SCN-NNN ref.
func (s *ContinuityService) CreateScenario(ctx context.Context, orgID uuid.UUID, req CreateScenarioReq) (*BIAScenario, error) {
	if req.Likelihood == "" {
		req.Likelihood = "possible"
	}

	impactTimeline := req.ImpactTimeline
	if len(impactTimeline) == 0 {
		impactTimeline = json.RawMessage(`{}`)
	}

	sc := &BIAScenario{}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO bia_scenarios (
			organization_id, scenario_ref, name, description,
			scenario_type, likelihood, affected_process_ids,
			affected_asset_ids, impact_timeline,
			estimated_financial_loss_eur, mitigation_strategy
		) VALUES (
			$1, generate_scn_ref($1), $2, $3,
			$4::bia_scenario_type, $5::bia_likelihood, $6,
			$7, $8::jsonb,
			$9, $10
		)
		RETURNING id, organization_id, scenario_ref, name,
			COALESCE(description, ''), scenario_type::TEXT,
			likelihood::TEXT, COALESCE(affected_process_ids, '{}'),
			COALESCE(affected_asset_ids, '{}'),
			COALESCE(impact_timeline, '{}'::jsonb),
			estimated_financial_loss_eur,
			COALESCE(mitigation_strategy, ''), status::TEXT,
			created_at, updated_at`,
		orgID, req.Name, req.Description,
		req.ScenarioType, req.Likelihood, req.AffectedProcessIDs,
		req.AffectedAssetIDs, impactTimeline,
		req.EstimatedFinancialLoss, req.MitigationStrategy,
	).Scan(
		&sc.ID, &sc.OrganizationID, &sc.ScenarioRef, &sc.Name,
		&sc.Description, &sc.ScenarioType,
		&sc.Likelihood, &sc.AffectedProcessIDs,
		&sc.AffectedAssetIDs, &sc.ImpactTimeline,
		&sc.EstimatedFinancialLoss,
		&sc.MitigationStrategy, &sc.Status,
		&sc.CreatedAt, &sc.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create BIA scenario: %w", err)
	}

	log.Info().
		Str("org_id", orgID.String()).
		Str("scenario_ref", sc.ScenarioRef).
		Str("type", sc.ScenarioType).
		Msg("BIA scenario created")

	return sc, nil
}

// ListScenarios returns a paginated list of BIA scenarios.
func (s *ContinuityService) ListScenarios(ctx context.Context, orgID uuid.UUID, page, pageSize int) ([]BIAScenario, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var total int64
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM bia_scenarios
		WHERE organization_id = $1 AND deleted_at IS NULL`, orgID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count scenarios: %w", err)
	}

	rows, err := s.pool.Query(ctx, `
		SELECT id, organization_id, scenario_ref, name,
			COALESCE(description, ''), scenario_type::TEXT,
			likelihood::TEXT, COALESCE(affected_process_ids, '{}'),
			COALESCE(affected_asset_ids, '{}'),
			COALESCE(impact_timeline, '{}'::jsonb),
			estimated_financial_loss_eur,
			COALESCE(mitigation_strategy, ''), status::TEXT,
			created_at, updated_at
		FROM bia_scenarios
		WHERE organization_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`,
		orgID, pageSize, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list scenarios: %w", err)
	}
	defer rows.Close()

	var scenarios []BIAScenario
	for rows.Next() {
		var sc BIAScenario
		if err := rows.Scan(
			&sc.ID, &sc.OrganizationID, &sc.ScenarioRef, &sc.Name,
			&sc.Description, &sc.ScenarioType,
			&sc.Likelihood, &sc.AffectedProcessIDs,
			&sc.AffectedAssetIDs, &sc.ImpactTimeline,
			&sc.EstimatedFinancialLoss,
			&sc.MitigationStrategy, &sc.Status,
			&sc.CreatedAt, &sc.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan scenario: %w", err)
		}
		scenarios = append(scenarios, sc)
	}
	return scenarios, total, nil
}

// GetScenario retrieves a single BIA scenario by ID.
func (s *ContinuityService) GetScenario(ctx context.Context, orgID, scenarioID uuid.UUID) (*BIAScenario, error) {
	sc := &BIAScenario{}
	err := s.pool.QueryRow(ctx, `
		SELECT id, organization_id, scenario_ref, name,
			COALESCE(description, ''), scenario_type::TEXT,
			likelihood::TEXT, COALESCE(affected_process_ids, '{}'),
			COALESCE(affected_asset_ids, '{}'),
			COALESCE(impact_timeline, '{}'::jsonb),
			estimated_financial_loss_eur,
			COALESCE(mitigation_strategy, ''), status::TEXT,
			created_at, updated_at
		FROM bia_scenarios
		WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL`,
		scenarioID, orgID,
	).Scan(
		&sc.ID, &sc.OrganizationID, &sc.ScenarioRef, &sc.Name,
		&sc.Description, &sc.ScenarioType,
		&sc.Likelihood, &sc.AffectedProcessIDs,
		&sc.AffectedAssetIDs, &sc.ImpactTimeline,
		&sc.EstimatedFinancialLoss,
		&sc.MitigationStrategy, &sc.Status,
		&sc.CreatedAt, &sc.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("BIA scenario not found: %w", err)
	}
	return sc, nil
}

// UpdateScenario updates mutable fields on a BIA scenario.
func (s *ContinuityService) UpdateScenario(ctx context.Context, orgID, scenarioID uuid.UUID, req UpdateScenarioReq) error {
	ct, err := s.pool.Exec(ctx, `
		UPDATE bia_scenarios SET
			name = COALESCE($3, name),
			description = COALESCE($4, description),
			scenario_type = COALESCE($5::bia_scenario_type, scenario_type),
			likelihood = COALESCE($6::bia_likelihood, likelihood),
			affected_process_ids = COALESCE($7, affected_process_ids),
			affected_asset_ids = COALESCE($8, affected_asset_ids),
			impact_timeline = COALESCE($9::jsonb, impact_timeline),
			estimated_financial_loss_eur = COALESCE($10, estimated_financial_loss_eur),
			mitigation_strategy = COALESCE($11, mitigation_strategy),
			status = COALESCE($12::bia_scenario_status, status)
		WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL`,
		scenarioID, orgID,
		req.Name, req.Description, req.ScenarioType, req.Likelihood,
		req.AffectedProcessIDs, req.AffectedAssetIDs,
		req.ImpactTimeline, req.EstimatedFinancialLoss,
		req.MitigationStrategy, req.Status,
	)
	if err != nil {
		return fmt.Errorf("failed to update scenario: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("scenario not found")
	}
	return nil
}

// ============================================================
// PLANS
// ============================================================

// CreatePlan creates a new continuity plan with auto-generated BCP-NNN ref.
func (s *ContinuityService) CreatePlan(ctx context.Context, orgID uuid.UUID, req CreatePlanReq) (*ContinuityPlan, error) {
	if req.ReviewFrequencyMonths <= 0 {
		req.ReviewFrequencyMonths = 6
	}

	// Default JSONB fields
	cmdStruct := req.CommandStructure
	if len(cmdStruct) == 0 {
		cmdStruct = json.RawMessage(`{}`)
	}
	commPlan := req.CommunicationPlan
	if len(commPlan) == 0 {
		commPlan = json.RawMessage(`{}`)
	}
	recovProcs := req.RecoveryProcedures
	if len(recovProcs) == 0 {
		recovProcs = json.RawMessage(`{}`)
	}
	resReqs := req.ResourceRequirements
	if len(resReqs) == 0 {
		resReqs = json.RawMessage(`{}`)
	}
	altSite := req.AlternateSiteDetails
	if len(altSite) == 0 {
		altSite = json.RawMessage(`{}`)
	}

	nextReview := time.Now().AddDate(0, req.ReviewFrequencyMonths, 0)

	plan := &ContinuityPlan{}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO continuity_plans (
			organization_id, plan_ref, name, plan_type, scope_description,
			covered_scenario_ids, covered_process_ids,
			activation_criteria, activation_authority,
			command_structure, communication_plan,
			recovery_procedures, resource_requirements,
			alternate_site_details, owner_user_id,
			review_frequency_months, next_review_date, document_path
		) VALUES (
			$1, generate_bcp_ref($1), $2, $3::continuity_plan_type, $4,
			$5, $6, $7, $8,
			$9::jsonb, $10::jsonb, $11::jsonb, $12::jsonb,
			$13::jsonb, $14, $15, $16, $17
		)
		RETURNING id, organization_id, plan_ref, name,
			plan_type::TEXT, status::TEXT, version,
			COALESCE(scope_description, ''),
			COALESCE(covered_scenario_ids, '{}'), COALESCE(covered_process_ids, '{}'),
			COALESCE(activation_criteria, ''), COALESCE(activation_authority, ''),
			COALESCE(command_structure, '{}'::jsonb),
			COALESCE(communication_plan, '{}'::jsonb),
			COALESCE(recovery_procedures, '{}'::jsonb),
			COALESCE(resource_requirements, '{}'::jsonb),
			COALESCE(alternate_site_details, '{}'::jsonb),
			owner_user_id, approved_by, approved_at,
			next_review_date, review_frequency_months,
			COALESCE(document_path, ''),
			created_at, updated_at`,
		orgID, req.Name, req.PlanType, req.ScopeDescription,
		req.CoveredScenarioIDs, req.CoveredProcessIDs,
		req.ActivationCriteria, req.ActivationAuthority,
		cmdStruct, commPlan, recovProcs, resReqs,
		altSite, req.OwnerUserID,
		req.ReviewFrequencyMonths, nextReview, req.DocumentPath,
	).Scan(
		&plan.ID, &plan.OrganizationID, &plan.PlanRef, &plan.Name,
		&plan.PlanType, &plan.Status, &plan.Version,
		&plan.ScopeDescription,
		&plan.CoveredScenarioIDs, &plan.CoveredProcessIDs,
		&plan.ActivationCriteria, &plan.ActivationAuthority,
		&plan.CommandStructure, &plan.CommunicationPlan,
		&plan.RecoveryProcedures, &plan.ResourceRequirements,
		&plan.AlternateSiteDetails,
		&plan.OwnerUserID, &plan.ApprovedBy, &plan.ApprovedAt,
		&plan.NextReviewDate, &plan.ReviewFrequencyMonths,
		&plan.DocumentPath,
		&plan.CreatedAt, &plan.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create continuity plan: %w", err)
	}

	log.Info().
		Str("org_id", orgID.String()).
		Str("plan_ref", plan.PlanRef).
		Str("type", plan.PlanType).
		Msg("Continuity plan created")

	return plan, nil
}

// ListPlans returns a paginated list of continuity plans.
func (s *ContinuityService) ListPlans(ctx context.Context, orgID uuid.UUID, page, pageSize int) ([]ContinuityPlan, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var total int64
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM continuity_plans
		WHERE organization_id = $1 AND deleted_at IS NULL`, orgID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count plans: %w", err)
	}

	rows, err := s.pool.Query(ctx, `
		SELECT id, organization_id, plan_ref, name,
			plan_type::TEXT, status::TEXT, version,
			COALESCE(scope_description, ''),
			COALESCE(covered_scenario_ids, '{}'), COALESCE(covered_process_ids, '{}'),
			COALESCE(activation_criteria, ''), COALESCE(activation_authority, ''),
			COALESCE(command_structure, '{}'::jsonb),
			COALESCE(communication_plan, '{}'::jsonb),
			COALESCE(recovery_procedures, '{}'::jsonb),
			COALESCE(resource_requirements, '{}'::jsonb),
			COALESCE(alternate_site_details, '{}'::jsonb),
			owner_user_id, approved_by, approved_at,
			next_review_date, review_frequency_months,
			COALESCE(document_path, ''),
			created_at, updated_at
		FROM continuity_plans
		WHERE organization_id = $1 AND deleted_at IS NULL
		ORDER BY
			CASE status
				WHEN 'active' THEN 1
				WHEN 'approved' THEN 2
				WHEN 'under_review' THEN 3
				WHEN 'draft' THEN 4
				WHEN 'archived' THEN 5
			END,
			name ASC
		LIMIT $2 OFFSET $3`,
		orgID, pageSize, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list plans: %w", err)
	}
	defer rows.Close()

	var plans []ContinuityPlan
	for rows.Next() {
		var p ContinuityPlan
		if err := rows.Scan(
			&p.ID, &p.OrganizationID, &p.PlanRef, &p.Name,
			&p.PlanType, &p.Status, &p.Version,
			&p.ScopeDescription,
			&p.CoveredScenarioIDs, &p.CoveredProcessIDs,
			&p.ActivationCriteria, &p.ActivationAuthority,
			&p.CommandStructure, &p.CommunicationPlan,
			&p.RecoveryProcedures, &p.ResourceRequirements,
			&p.AlternateSiteDetails,
			&p.OwnerUserID, &p.ApprovedBy, &p.ApprovedAt,
			&p.NextReviewDate, &p.ReviewFrequencyMonths,
			&p.DocumentPath,
			&p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan plan: %w", err)
		}
		plans = append(plans, p)
	}
	return plans, total, nil
}

// GetPlan retrieves a single continuity plan by ID.
func (s *ContinuityService) GetPlan(ctx context.Context, orgID, planID uuid.UUID) (*ContinuityPlan, error) {
	p := &ContinuityPlan{}
	err := s.pool.QueryRow(ctx, `
		SELECT id, organization_id, plan_ref, name,
			plan_type::TEXT, status::TEXT, version,
			COALESCE(scope_description, ''),
			COALESCE(covered_scenario_ids, '{}'), COALESCE(covered_process_ids, '{}'),
			COALESCE(activation_criteria, ''), COALESCE(activation_authority, ''),
			COALESCE(command_structure, '{}'::jsonb),
			COALESCE(communication_plan, '{}'::jsonb),
			COALESCE(recovery_procedures, '{}'::jsonb),
			COALESCE(resource_requirements, '{}'::jsonb),
			COALESCE(alternate_site_details, '{}'::jsonb),
			owner_user_id, approved_by, approved_at,
			next_review_date, review_frequency_months,
			COALESCE(document_path, ''),
			created_at, updated_at
		FROM continuity_plans
		WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL`,
		planID, orgID,
	).Scan(
		&p.ID, &p.OrganizationID, &p.PlanRef, &p.Name,
		&p.PlanType, &p.Status, &p.Version,
		&p.ScopeDescription,
		&p.CoveredScenarioIDs, &p.CoveredProcessIDs,
		&p.ActivationCriteria, &p.ActivationAuthority,
		&p.CommandStructure, &p.CommunicationPlan,
		&p.RecoveryProcedures, &p.ResourceRequirements,
		&p.AlternateSiteDetails,
		&p.OwnerUserID, &p.ApprovedBy, &p.ApprovedAt,
		&p.NextReviewDate, &p.ReviewFrequencyMonths,
		&p.DocumentPath,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("continuity plan not found: %w", err)
	}
	return p, nil
}

// UpdatePlan updates mutable fields on a continuity plan and increments version.
func (s *ContinuityService) UpdatePlan(ctx context.Context, orgID, planID uuid.UUID, req UpdatePlanReq) error {
	ct, err := s.pool.Exec(ctx, `
		UPDATE continuity_plans SET
			name = COALESCE($3, name),
			plan_type = COALESCE($4::continuity_plan_type, plan_type),
			status = COALESCE($5::continuity_plan_status, status),
			scope_description = COALESCE($6, scope_description),
			covered_scenario_ids = COALESCE($7, covered_scenario_ids),
			covered_process_ids = COALESCE($8, covered_process_ids),
			activation_criteria = COALESCE($9, activation_criteria),
			activation_authority = COALESCE($10, activation_authority),
			command_structure = COALESCE($11::jsonb, command_structure),
			communication_plan = COALESCE($12::jsonb, communication_plan),
			recovery_procedures = COALESCE($13::jsonb, recovery_procedures),
			resource_requirements = COALESCE($14::jsonb, resource_requirements),
			alternate_site_details = COALESCE($15::jsonb, alternate_site_details),
			owner_user_id = COALESCE($16, owner_user_id),
			review_frequency_months = COALESCE($17, review_frequency_months),
			document_path = COALESCE($18, document_path),
			version = version + 1
		WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL`,
		planID, orgID,
		req.Name, req.PlanType, req.Status, req.ScopeDescription,
		req.CoveredScenarioIDs, req.CoveredProcessIDs,
		req.ActivationCriteria, req.ActivationAuthority,
		req.CommandStructure, req.CommunicationPlan,
		req.RecoveryProcedures, req.ResourceRequirements,
		req.AlternateSiteDetails, req.OwnerUserID,
		req.ReviewFrequencyMonths, req.DocumentPath,
	)
	if err != nil {
		return fmt.Errorf("failed to update continuity plan: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("continuity plan not found")
	}
	return nil
}

// ApprovePlan marks a continuity plan as approved, records the approver,
// sets the next review date, and transitions status to 'approved'.
func (s *ContinuityService) ApprovePlan(ctx context.Context, orgID, planID, approverID uuid.UUID) error {
	// Fetch review frequency
	var freqMonths int
	err := s.pool.QueryRow(ctx, `
		SELECT review_frequency_months FROM continuity_plans
		WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL`,
		planID, orgID,
	).Scan(&freqMonths)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("continuity plan not found")
		}
		return fmt.Errorf("failed to fetch plan: %w", err)
	}

	now := time.Now()
	nextReview := now.AddDate(0, freqMonths, 0)

	ct, err := s.pool.Exec(ctx, `
		UPDATE continuity_plans SET
			status = 'approved'::continuity_plan_status,
			approved_by = $3,
			approved_at = $4,
			next_review_date = $5
		WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL`,
		planID, orgID, approverID, now, nextReview,
	)
	if err != nil {
		return fmt.Errorf("failed to approve plan: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("continuity plan not found")
	}

	log.Info().
		Str("org_id", orgID.String()).
		Str("plan_id", planID.String()).
		Str("approved_by", approverID.String()).
		Msg("Continuity plan approved")

	return nil
}

// ============================================================
// EXERCISES
// ============================================================

// ScheduleExercise creates a new BC exercise with auto-generated BCX-NNN ref.
func (s *ContinuityService) ScheduleExercise(ctx context.Context, orgID uuid.UUID, req CreateExerciseReq) (*BCExercise, error) {
	var scheduledDate *time.Time
	if req.ScheduledDate != "" {
		t, err := time.Parse("2006-01-02", req.ScheduledDate)
		if err != nil {
			return nil, fmt.Errorf("invalid scheduled_date format, expected YYYY-MM-DD: %w", err)
		}
		scheduledDate = &t
	}

	participants := req.Participants
	if len(participants) == 0 {
		participants = json.RawMessage(`[]`)
	}

	ex := &BCExercise{}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO bc_exercises (
			organization_id, exercise_ref, name, exercise_type,
			plan_id, scenario_id, scheduled_date,
			participants, conducted_by
		) VALUES (
			$1, generate_bcx_ref($1), $2, $3::bc_exercise_type,
			$4, $5, $6, $7::jsonb, $8
		)
		RETURNING id, organization_id, exercise_ref, name,
			exercise_type::TEXT, plan_id, scenario_id, status::TEXT,
			scheduled_date, actual_date,
			COALESCE(participants, '[]'::jsonb),
			rto_achieved_hours, rpo_achieved_hours,
			objectives_met,
			COALESCE(lessons_learned, ''),
			COALESCE(gaps_identified, ''),
			COALESCE(improvement_actions, '[]'::jsonb),
			overall_rating::TEXT,
			COALESCE(report_document_path, ''),
			conducted_by, created_at, updated_at`,
		orgID, req.Name, req.ExerciseType,
		req.PlanID, req.ScenarioID, scheduledDate,
		participants, req.ConductedBy,
	).Scan(
		&ex.ID, &ex.OrganizationID, &ex.ExerciseRef, &ex.Name,
		&ex.ExerciseType, &ex.PlanID, &ex.ScenarioID, &ex.Status,
		&ex.ScheduledDate, &ex.ActualDate,
		&ex.Participants,
		&ex.RTOAchievedHours, &ex.RPOAchievedHours,
		&ex.ObjectivesMet,
		&ex.LessonsLearned, &ex.GapsIdentified,
		&ex.ImprovementActions,
		&ex.OverallRating,
		&ex.ReportDocumentPath,
		&ex.ConductedBy, &ex.CreatedAt, &ex.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to schedule BC exercise: %w", err)
	}

	log.Info().
		Str("org_id", orgID.String()).
		Str("exercise_ref", ex.ExerciseRef).
		Str("type", ex.ExerciseType).
		Msg("BC exercise scheduled")

	return ex, nil
}

// ListExercises returns a paginated list of BC exercises.
func (s *ContinuityService) ListExercises(ctx context.Context, orgID uuid.UUID, page, pageSize int) ([]BCExercise, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var total int64
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM bc_exercises
		WHERE organization_id = $1 AND deleted_at IS NULL`, orgID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count exercises: %w", err)
	}

	rows, err := s.pool.Query(ctx, `
		SELECT id, organization_id, exercise_ref, name,
			exercise_type::TEXT, plan_id, scenario_id, status::TEXT,
			scheduled_date, actual_date,
			COALESCE(participants, '[]'::jsonb),
			rto_achieved_hours, rpo_achieved_hours,
			objectives_met,
			COALESCE(lessons_learned, ''),
			COALESCE(gaps_identified, ''),
			COALESCE(improvement_actions, '[]'::jsonb),
			overall_rating::TEXT,
			COALESCE(report_document_path, ''),
			conducted_by, created_at, updated_at
		FROM bc_exercises
		WHERE organization_id = $1 AND deleted_at IS NULL
		ORDER BY COALESCE(scheduled_date, '9999-12-31') DESC, created_at DESC
		LIMIT $2 OFFSET $3`,
		orgID, pageSize, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list exercises: %w", err)
	}
	defer rows.Close()

	var exercises []BCExercise
	for rows.Next() {
		var ex BCExercise
		if err := rows.Scan(
			&ex.ID, &ex.OrganizationID, &ex.ExerciseRef, &ex.Name,
			&ex.ExerciseType, &ex.PlanID, &ex.ScenarioID, &ex.Status,
			&ex.ScheduledDate, &ex.ActualDate,
			&ex.Participants,
			&ex.RTOAchievedHours, &ex.RPOAchievedHours,
			&ex.ObjectivesMet,
			&ex.LessonsLearned, &ex.GapsIdentified,
			&ex.ImprovementActions,
			&ex.OverallRating,
			&ex.ReportDocumentPath,
			&ex.ConductedBy, &ex.CreatedAt, &ex.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan exercise: %w", err)
		}
		exercises = append(exercises, ex)
	}
	return exercises, total, nil
}

// CompleteExercise records the results of a BC exercise and marks it as completed.
func (s *ContinuityService) CompleteExercise(ctx context.Context, orgID, exerciseID uuid.UUID, req CompleteExerciseReq) error {
	var actualDate *time.Time
	if req.ActualDate != "" {
		t, err := time.Parse("2006-01-02", req.ActualDate)
		if err != nil {
			return fmt.Errorf("invalid actual_date format, expected YYYY-MM-DD: %w", err)
		}
		actualDate = &t
	} else {
		now := time.Now()
		actualDate = &now
	}

	improvementActions := req.ImprovementActions
	if len(improvementActions) == 0 {
		improvementActions = json.RawMessage(`[]`)
	}

	ct, err := s.pool.Exec(ctx, `
		UPDATE bc_exercises SET
			status = 'completed'::bc_exercise_status,
			actual_date = $3,
			rto_achieved_hours = $4,
			rpo_achieved_hours = $5,
			objectives_met = $6,
			lessons_learned = $7,
			gaps_identified = $8,
			improvement_actions = $9::jsonb,
			overall_rating = $10::bc_exercise_rating,
			report_document_path = $11
		WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL`,
		exerciseID, orgID,
		actualDate,
		req.RTOAchievedHours, req.RPOAchievedHours,
		req.ObjectivesMet,
		req.LessonsLearned, req.GapsIdentified,
		improvementActions, req.OverallRating,
		req.ReportDocumentPath,
	)
	if err != nil {
		return fmt.Errorf("failed to complete exercise: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("exercise not found")
	}

	log.Info().
		Str("org_id", orgID.String()).
		Str("exercise_id", exerciseID.String()).
		Msg("BC exercise completed")

	return nil
}

// ============================================================
// DASHBOARD
// ============================================================

// GetBCDashboard aggregates business continuity metrics for the organisation.
func (s *ContinuityService) GetBCDashboard(ctx context.Context, orgID uuid.UUID) (*BCDashboard, error) {
	dash := &BCDashboard{}

	// Process stats
	err := s.pool.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE last_bia_date IS NULL),
			COUNT(*) FILTER (WHERE criticality = 'mission_critical'),
			COUNT(*) FILTER (WHERE criticality = 'business_critical'),
			COUNT(*) FILTER (WHERE rto_hours IS NOT NULL),
			COUNT(*) FILTER (WHERE rpo_hours IS NOT NULL)
		FROM business_processes
		WHERE organization_id = $1 AND deleted_at IS NULL AND status = 'active'`, orgID,
	).Scan(
		&dash.TotalProcesses, &dash.ProcessesWithoutBIA,
		&dash.MissionCriticalCount, &dash.BusinessCriticalCount,
		&dash.RTOCoveragePercent, &dash.RPOCoveragePercent, // reused as temp counts
	)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("failed to get process stats: %w", err)
	}

	// Calculate coverage percentages
	if dash.TotalProcesses > 0 {
		rtoCount := dash.RTOCoveragePercent
		rpoCount := dash.RPOCoveragePercent
		dash.RTOCoveragePercent = (rtoCount / float64(dash.TotalProcesses)) * 100
		dash.RPOCoveragePercent = (rpoCount / float64(dash.TotalProcesses)) * 100
	}

	// Plan stats
	err = s.pool.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE status = 'active'),
			COUNT(*) FILTER (WHERE status = 'draft'),
			COUNT(*) FILTER (WHERE next_review_date IS NOT NULL AND next_review_date <= CURRENT_DATE)
		FROM continuity_plans
		WHERE organization_id = $1 AND deleted_at IS NULL`, orgID,
	).Scan(
		&dash.TotalPlans, &dash.ActivePlans,
		&dash.DraftPlans, &dash.PlansRequiringReview,
	)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("failed to get plan stats: %w", err)
	}

	// Scenario count
	err = s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM bia_scenarios
		WHERE organization_id = $1 AND deleted_at IS NULL`, orgID,
	).Scan(&dash.TotalScenarios)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("failed to get scenario count: %w", err)
	}

	// Exercise stats
	err = s.pool.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE status = 'planned'),
			COUNT(*) FILTER (WHERE status = 'completed')
		FROM bc_exercises
		WHERE organization_id = $1 AND deleted_at IS NULL`, orgID,
	).Scan(
		&dash.TotalExercises, &dash.PlannedExercises,
		&dash.CompletedExercises,
	)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("failed to get exercise stats: %w", err)
	}

	// Last exercise date and result
	var lastRating *string
	err = s.pool.QueryRow(ctx, `
		SELECT actual_date, overall_rating::TEXT
		FROM bc_exercises
		WHERE organization_id = $1 AND status = 'completed' AND deleted_at IS NULL
		ORDER BY actual_date DESC NULLS LAST
		LIMIT 1`, orgID,
	).Scan(&dash.LastExerciseDate, &lastRating)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("failed to get last exercise: %w", err)
	}
	if lastRating != nil {
		dash.LastExerciseResult = *lastRating
	}

	// SPoF count (count distinct entities with 2+ critical process dependencies)
	spofRows, err := s.pool.Query(ctx, `
		SELECT dependency_name, COUNT(DISTINCT pdm.process_id)
		FROM process_dependencies_map pdm
		JOIN business_processes bp ON bp.id = pdm.process_id
			AND bp.organization_id = pdm.organization_id
			AND bp.deleted_at IS NULL
			AND bp.criticality IN ('mission_critical', 'business_critical')
		WHERE pdm.organization_id = $1
		GROUP BY dependency_name, dependency_entity_id
		HAVING COUNT(DISTINCT pdm.process_id) >= 2`, orgID,
	)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("failed to count SPoFs: %w", err)
	}
	if spofRows != nil {
		defer spofRows.Close()
		for spofRows.Next() {
			var name string
			var cnt int
			_ = spofRows.Scan(&name, &cnt)
			dash.SPOFCount++
		}
	}

	return dash, nil
}
