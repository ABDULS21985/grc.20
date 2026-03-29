// Package service contains the Remediation Planner for ComplianceForge.
// It orchestrates compliance remediation plan lifecycle management including
// AI-assisted plan generation, progress tracking, action management, and approval workflows.
package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// ============================================================
// DOMAIN TYPES
// ============================================================

// RemediationPlan represents a compliance remediation plan.
type RemediationPlan struct {
	ID                   uuid.UUID        `json:"id"`
	OrganizationID       uuid.UUID        `json:"organization_id"`
	PlanRef              string           `json:"plan_ref"`
	Name                 string           `json:"name"`
	Description          string           `json:"description"`
	PlanType             string           `json:"plan_type"`
	Status               string           `json:"status"`
	ScopeFrameworkIDs    []uuid.UUID      `json:"scope_framework_ids"`
	ScopeDescription     string           `json:"scope_description"`
	Priority             string           `json:"priority"`
	AIGenerated          bool             `json:"ai_generated"`
	AIModel              string           `json:"ai_model,omitempty"`
	AIPromptSummary      string           `json:"ai_prompt_summary,omitempty"`
	AIGenerationDate     *time.Time       `json:"ai_generation_date,omitempty"`
	AIConfidenceScore    *float64         `json:"ai_confidence_score,omitempty"`
	HumanReviewed        bool             `json:"human_reviewed"`
	HumanReviewedBy      *uuid.UUID       `json:"human_reviewed_by,omitempty"`
	HumanReviewedAt      *time.Time       `json:"human_reviewed_at,omitempty"`
	TargetCompletionDate *time.Time       `json:"target_completion_date,omitempty"`
	EstimatedTotalHours  float64          `json:"estimated_total_hours"`
	EstimatedTotalCost   float64          `json:"estimated_total_cost_eur"`
	ActualCompletionDate *time.Time       `json:"actual_completion_date,omitempty"`
	CompletionPercentage float64          `json:"completion_percentage"`
	OwnerUserID          *uuid.UUID       `json:"owner_user_id,omitempty"`
	CreatedBy            uuid.UUID        `json:"created_by"`
	ApprovedBy           *uuid.UUID       `json:"approved_by,omitempty"`
	ApprovedAt           *time.Time       `json:"approved_at,omitempty"`
	Metadata             []byte           `json:"metadata,omitempty"`
	CreatedAt            time.Time        `json:"created_at"`
	UpdatedAt            time.Time        `json:"updated_at"`
	Actions              []RemediationAction `json:"actions,omitempty"`
}

// RemediationAction represents a single action within a remediation plan.
type RemediationAction struct {
	ID                            uuid.UUID  `json:"id"`
	OrganizationID                uuid.UUID  `json:"organization_id"`
	PlanID                        uuid.UUID  `json:"plan_id"`
	ActionRef                     string     `json:"action_ref"`
	SortOrder                     int        `json:"sort_order"`
	Title                         string     `json:"title"`
	Description                   string     `json:"description"`
	ActionType                    string     `json:"action_type"`
	LinkedControlImplementationID *uuid.UUID `json:"linked_control_implementation_id,omitempty"`
	LinkedFindingID               *uuid.UUID `json:"linked_finding_id,omitempty"`
	LinkedRiskTreatmentID         *uuid.UUID `json:"linked_risk_treatment_id,omitempty"`
	FrameworkControlCode          string     `json:"framework_control_code"`
	Priority                      string     `json:"priority"`
	EstimatedHours                float64    `json:"estimated_hours"`
	EstimatedCostEUR              float64    `json:"estimated_cost_eur"`
	RequiredSkills                []string   `json:"required_skills"`
	Dependencies                  []uuid.UUID `json:"dependencies"`
	AssignedTo                    *uuid.UUID `json:"assigned_to,omitempty"`
	TargetStartDate               *time.Time `json:"target_start_date,omitempty"`
	TargetEndDate                 *time.Time `json:"target_end_date,omitempty"`
	Status                        string     `json:"status"`
	ActualStartDate               *time.Time `json:"actual_start_date,omitempty"`
	ActualEndDate                 *time.Time `json:"actual_end_date,omitempty"`
	ActualHours                   *float64   `json:"actual_hours,omitempty"`
	ActualCostEUR                 *float64   `json:"actual_cost_eur,omitempty"`
	CompletionNotes               string     `json:"completion_notes,omitempty"`
	EvidencePaths                 []string   `json:"evidence_paths"`
	AIImplementationGuidance      string     `json:"ai_implementation_guidance,omitempty"`
	AIEvidenceSuggestions         []string   `json:"ai_evidence_suggestions"`
	AIToolRecommendations         []string   `json:"ai_tool_recommendations"`
	AIRiskIfDeferred              string     `json:"ai_risk_if_deferred,omitempty"`
	AICrossFrameworkBenefit       string     `json:"ai_cross_framework_benefit,omitempty"`
	CreatedAt                     time.Time  `json:"created_at"`
	UpdatedAt                     time.Time  `json:"updated_at"`
}

// GeneratePlanRequest is the request to generate a new remediation plan.
type GeneratePlanRequest struct {
	Name             string            `json:"name"`
	Description      string            `json:"description"`
	PlanType         string            `json:"plan_type"`
	ScopeFrameworkIDs []uuid.UUID      `json:"scope_framework_ids"`
	ScopeDescription string            `json:"scope_description"`
	Priority         string            `json:"priority"`
	OwnerUserID      *uuid.UUID        `json:"owner_user_id"`
	TargetDate       *time.Time        `json:"target_completion_date"`
	UseAI            bool              `json:"use_ai"`
	AIRequest        *RemediationPlanRequest `json:"ai_request,omitempty"`
}

// UpdatePlanRequest is the request to update an existing plan.
type UpdatePlanRequest struct {
	Name                 *string    `json:"name,omitempty"`
	Description          *string    `json:"description,omitempty"`
	Priority             *string    `json:"priority,omitempty"`
	Status               *string    `json:"status,omitempty"`
	OwnerUserID          *uuid.UUID `json:"owner_user_id,omitempty"`
	TargetCompletionDate *time.Time `json:"target_completion_date,omitempty"`
	ScopeDescription     *string    `json:"scope_description,omitempty"`
}

// UpdateActionRequest is the request to update an existing action.
type UpdateActionRequest struct {
	Title           *string    `json:"title,omitempty"`
	Description     *string    `json:"description,omitempty"`
	Priority        *string    `json:"priority,omitempty"`
	Status          *string    `json:"status,omitempty"`
	AssignedTo      *uuid.UUID `json:"assigned_to,omitempty"`
	TargetStartDate *time.Time `json:"target_start_date,omitempty"`
	TargetEndDate   *time.Time `json:"target_end_date,omitempty"`
	SortOrder       *int       `json:"sort_order,omitempty"`
}

// CompleteActionRequest is the request to mark an action as complete.
type CompleteActionRequest struct {
	CompletionNotes string   `json:"completion_notes"`
	ActualHours     float64  `json:"actual_hours"`
	ActualCostEUR   float64  `json:"actual_cost_eur"`
	EvidencePaths   []string `json:"evidence_paths"`
}

// PlanTimeline represents the timeline view of a remediation plan.
type PlanTimeline struct {
	PlanID          uuid.UUID            `json:"plan_id"`
	PlanRef         string               `json:"plan_ref"`
	PlanName        string               `json:"plan_name"`
	StartDate       *time.Time           `json:"start_date"`
	EndDate         *time.Time           `json:"end_date"`
	TotalWeeks      int                  `json:"total_weeks"`
	TimelineEntries []RemediationTimelineEntry `json:"entries"`
	CriticalPath    []uuid.UUID          `json:"critical_path"`
	Milestones      []TimelineMilestone  `json:"milestones"`
}

// RemediationTimelineEntry represents a single action on the remediation timeline.
type RemediationTimelineEntry struct {
	ActionID    uuid.UUID  `json:"action_id"`
	ActionRef   string     `json:"action_ref"`
	Title       string     `json:"title"`
	StartDate   *time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
	Status      string     `json:"status"`
	Priority    string     `json:"priority"`
	AssignedTo  *uuid.UUID `json:"assigned_to,omitempty"`
	Dependencies []uuid.UUID `json:"dependencies"`
}

// TimelineMilestone represents a milestone in the timeline.
type TimelineMilestone struct {
	Date        time.Time `json:"date"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
}

// PlanProgress represents progress tracking data for a plan.
type PlanProgress struct {
	PlanID               uuid.UUID            `json:"plan_id"`
	PlanRef              string               `json:"plan_ref"`
	CompletionPercentage float64              `json:"completion_percentage"`
	TotalActions         int                  `json:"total_actions"`
	CompletedActions     int                  `json:"completed_actions"`
	InProgressActions    int                  `json:"in_progress_actions"`
	BlockedActions       int                  `json:"blocked_actions"`
	PendingActions       int                  `json:"pending_actions"`
	DeferredActions      int                  `json:"deferred_actions"`
	CancelledActions     int                  `json:"cancelled_actions"`
	EstimatedHoursTotal  float64              `json:"estimated_hours_total"`
	ActualHoursTotal     float64              `json:"actual_hours_total"`
	EstimatedCostTotal   float64              `json:"estimated_cost_total"`
	ActualCostTotal      float64              `json:"actual_cost_total"`
	OnTrack              bool                 `json:"on_track"`
	DaysRemaining        int                  `json:"days_remaining"`
	DaysOverdue          int                  `json:"days_overdue"`
	StatusBreakdown      map[string]int       `json:"status_breakdown"`
	PriorityBreakdown    map[string]int       `json:"priority_breakdown"`
	CriticalPathActions  []CriticalPathAction `json:"critical_path_actions"`
}

// CriticalPathAction represents an action on the critical path.
type CriticalPathAction struct {
	ActionID  uuid.UUID `json:"action_id"`
	ActionRef string    `json:"action_ref"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	Priority  string    `json:"priority"`
	DaysLeft  int       `json:"days_left"`
}

// ============================================================
// REMEDIATION PLANNER SERVICE
// ============================================================

// RemediationPlanner orchestrates compliance remediation plan lifecycle.
type RemediationPlanner struct {
	pool      *pgxpool.Pool
	aiService *AIService
}

// NewRemediationPlanner creates a new RemediationPlanner.
func NewRemediationPlanner(pool *pgxpool.Pool, aiService *AIService) *RemediationPlanner {
	return &RemediationPlanner{
		pool:      pool,
		aiService: aiService,
	}
}

// GeneratePlan creates a remediation plan, optionally using AI for generation.
func (rp *RemediationPlanner) GeneratePlan(ctx context.Context, orgID uuid.UUID, req GeneratePlanRequest) (*RemediationPlan, error) {
	tx, err := rp.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Set RLS context
	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_org', $1, true)", orgID.String()); err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	// Generate plan reference
	var planRef string
	err = tx.QueryRow(ctx, "SELECT generate_remediation_plan_ref($1)", orgID).Scan(&planRef)
	if err != nil {
		return nil, fmt.Errorf("failed to generate plan reference: %w", err)
	}

	// Get the user ID from context (stored in the plan)
	userID := uuid.Nil
	if uid, ok := ctx.Value("user_id").(uuid.UUID); ok {
		userID = uid
	}

	plan := &RemediationPlan{
		OrganizationID: orgID,
		PlanRef:        planRef,
		Name:           req.Name,
		Description:    req.Description,
		PlanType:       req.PlanType,
		Status:         "draft",
		ScopeFrameworkIDs: req.ScopeFrameworkIDs,
		ScopeDescription: req.ScopeDescription,
		Priority:        req.Priority,
		OwnerUserID:     req.OwnerUserID,
		CreatedBy:       userID,
	}

	if req.TargetDate != nil {
		plan.TargetCompletionDate = req.TargetDate
	}

	var aiActions []RemediationActionSuggestion

	// Attempt AI generation if requested
	if req.UseAI && req.AIRequest != nil && rp.aiService != nil && rp.aiService.isAvailable() {
		aiResp, err := rp.aiService.GenerateRemediationPlan(ctx, orgID, *req.AIRequest)
		if err != nil {
			log.Warn().Err(err).Msg("AI plan generation failed, continuing with manual plan")
		} else {
			plan.AIGenerated = true
			plan.AIModel = rp.aiService.model
			plan.AIPromptSummary = fmt.Sprintf("Generated plan for %d gaps across %d frameworks",
				len(req.AIRequest.Gaps), len(req.AIRequest.FrameworkIDs))
			now := time.Now()
			plan.AIGenerationDate = &now
			plan.AIConfidenceScore = &aiResp.ConfidenceScore

			if aiResp.PlanName != "" {
				plan.Name = aiResp.PlanName
			}
			if aiResp.PlanDescription != "" {
				plan.Description = aiResp.PlanDescription
			}
			if aiResp.Priority != "" {
				plan.Priority = aiResp.Priority
			}
			plan.EstimatedTotalHours = aiResp.EstimatedHours
			plan.EstimatedTotalCost = aiResp.EstimatedCost
			aiActions = aiResp.Actions
		}
	} else if req.UseAI && req.AIRequest != nil {
		// AI requested but not available, use fallback
		aiResp, err := rp.aiService.fallbackRemediationPlan(*req.AIRequest)
		if err == nil {
			plan.AIGenerated = false
			plan.EstimatedTotalHours = aiResp.EstimatedHours
			plan.EstimatedTotalCost = aiResp.EstimatedCost
			aiActions = aiResp.Actions
		}
	}

	// Validate plan_type
	validPlanTypes := map[string]bool{
		"gap_remediation": true, "audit_finding": true, "risk_treatment": true,
		"continuous_improvement": true, "incident_response": true, "regulatory_change": true,
	}
	if !validPlanTypes[plan.PlanType] {
		plan.PlanType = "gap_remediation"
	}

	// Validate priority
	validPriorities := map[string]bool{"critical": true, "high": true, "medium": true, "low": true}
	if !validPriorities[plan.Priority] {
		plan.Priority = "medium"
	}

	// Insert the plan
	err = tx.QueryRow(ctx, `
		INSERT INTO remediation_plans (
			organization_id, plan_ref, name, description, plan_type, status,
			scope_framework_ids, scope_description, priority,
			ai_generated, ai_model, ai_prompt_summary, ai_generation_date, ai_confidence_score,
			target_completion_date, estimated_total_hours, estimated_total_cost_eur,
			owner_user_id, created_by, metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9,
			$10, $11, $12, $13, $14,
			$15, $16, $17,
			$18, $19, '{}'::jsonb
		) RETURNING id, created_at, updated_at
	`,
		plan.OrganizationID, plan.PlanRef, plan.Name, plan.Description, plan.PlanType, plan.Status,
		plan.ScopeFrameworkIDs, plan.ScopeDescription, plan.Priority,
		plan.AIGenerated, nullableString(plan.AIModel), nullableString(plan.AIPromptSummary), plan.AIGenerationDate, plan.AIConfidenceScore,
		plan.TargetCompletionDate, plan.EstimatedTotalHours, plan.EstimatedTotalCost,
		plan.OwnerUserID, plan.CreatedBy,
	).Scan(&plan.ID, &plan.CreatedAt, &plan.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to insert remediation plan: %w", err)
	}

	// Insert actions
	plan.Actions = make([]RemediationAction, 0, len(aiActions))
	for _, suggestion := range aiActions {
		action, err := rp.insertActionFromSuggestion(ctx, tx, orgID, plan.ID, suggestion)
		if err != nil {
			log.Warn().Err(err).Str("action_title", suggestion.Title).Msg("Failed to insert action")
			continue
		}
		plan.Actions = append(plan.Actions, *action)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit plan creation: %w", err)
	}

	return plan, nil
}

// insertActionFromSuggestion creates a remediation action from an AI suggestion.
func (rp *RemediationPlanner) insertActionFromSuggestion(ctx context.Context, tx pgx.Tx, orgID, planID uuid.UUID, s RemediationActionSuggestion) (*RemediationAction, error) {
	var actionRef string
	err := tx.QueryRow(ctx, "SELECT generate_remediation_action_ref($1)", planID).Scan(&actionRef)
	if err != nil {
		return nil, fmt.Errorf("failed to generate action reference: %w", err)
	}

	// Validate action_type
	validTypes := map[string]bool{
		"implement_control": true, "update_policy": true, "deploy_technical": true,
		"conduct_training": true, "perform_assessment": true, "gather_evidence": true,
		"review_process": true, "third_party_engagement": true, "documentation": true,
		"monitoring_setup": true,
	}
	actionType := s.ActionType
	if !validTypes[actionType] {
		actionType = "implement_control"
	}

	// Validate priority
	priority := s.Priority
	validPriorities := map[string]bool{"critical": true, "high": true, "medium": true, "low": true}
	if !validPriorities[priority] {
		priority = "medium"
	}

	action := &RemediationAction{
		OrganizationID:           orgID,
		PlanID:                   planID,
		ActionRef:                actionRef,
		SortOrder:                s.SortOrder,
		Title:                    s.Title,
		Description:              s.Description,
		ActionType:               actionType,
		FrameworkControlCode:     s.FrameworkControlCode,
		Priority:                 priority,
		EstimatedHours:           s.EstimatedHours,
		EstimatedCostEUR:         s.EstimatedCostEUR,
		RequiredSkills:           s.RequiredSkills,
		Status:                   "pending",
		AIImplementationGuidance: s.ImplementationGuidance,
		AIEvidenceSuggestions:    s.EvidenceSuggestions,
		AIToolRecommendations:    s.ToolRecommendations,
		AIRiskIfDeferred:         s.RiskIfDeferred,
		AICrossFrameworkBenefit:  s.CrossFrameworkBenefit,
	}

	if action.RequiredSkills == nil {
		action.RequiredSkills = []string{}
	}
	if action.AIEvidenceSuggestions == nil {
		action.AIEvidenceSuggestions = []string{}
	}
	if action.AIToolRecommendations == nil {
		action.AIToolRecommendations = []string{}
	}

	err = tx.QueryRow(ctx, `
		INSERT INTO remediation_actions (
			organization_id, plan_id, action_ref, sort_order, title, description,
			action_type, framework_control_code, priority,
			estimated_hours, estimated_cost_eur, required_skills, dependencies,
			status, evidence_paths,
			ai_implementation_guidance, ai_evidence_suggestions,
			ai_tool_recommendations, ai_risk_if_deferred, ai_cross_framework_benefit
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9,
			$10, $11, $12, '{}',
			$13, '{}',
			$14, $15,
			$16, $17, $18
		) RETURNING id, created_at, updated_at
	`,
		action.OrganizationID, action.PlanID, action.ActionRef, action.SortOrder,
		action.Title, action.Description,
		action.ActionType, action.FrameworkControlCode, action.Priority,
		action.EstimatedHours, action.EstimatedCostEUR, action.RequiredSkills,
		action.Status,
		action.AIImplementationGuidance, action.AIEvidenceSuggestions,
		action.AIToolRecommendations, action.AIRiskIfDeferred, action.AICrossFrameworkBenefit,
	).Scan(&action.ID, &action.CreatedAt, &action.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to insert remediation action: %w", err)
	}

	return action, nil
}

// GetPlan retrieves a single remediation plan by ID with its actions.
func (rp *RemediationPlanner) GetPlan(ctx context.Context, orgID, planID uuid.UUID) (*RemediationPlan, error) {
	plan := &RemediationPlan{}

	err := rp.pool.QueryRow(ctx, `
		SELECT
			id, organization_id, plan_ref, name, description, plan_type, status,
			scope_framework_ids, scope_description, priority,
			ai_generated, ai_model, ai_prompt_summary, ai_generation_date, ai_confidence_score,
			human_reviewed, human_reviewed_by, human_reviewed_at,
			target_completion_date, estimated_total_hours, estimated_total_cost_eur,
			actual_completion_date, completion_percentage,
			owner_user_id, created_by, approved_by, approved_at,
			metadata, created_at, updated_at
		FROM remediation_plans
		WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL
	`, planID, orgID).Scan(
		&plan.ID, &plan.OrganizationID, &plan.PlanRef, &plan.Name, &plan.Description,
		&plan.PlanType, &plan.Status,
		&plan.ScopeFrameworkIDs, &plan.ScopeDescription, &plan.Priority,
		&plan.AIGenerated, &plan.AIModel, &plan.AIPromptSummary, &plan.AIGenerationDate, &plan.AIConfidenceScore,
		&plan.HumanReviewed, &plan.HumanReviewedBy, &plan.HumanReviewedAt,
		&plan.TargetCompletionDate, &plan.EstimatedTotalHours, &plan.EstimatedTotalCost,
		&plan.ActualCompletionDate, &plan.CompletionPercentage,
		&plan.OwnerUserID, &plan.CreatedBy, &plan.ApprovedBy, &plan.ApprovedAt,
		&plan.Metadata, &plan.CreatedAt, &plan.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("remediation plan not found: %w", err)
	}

	// Fetch actions
	actions, err := rp.getActionsByPlanID(ctx, orgID, planID)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to fetch plan actions")
	} else {
		plan.Actions = actions
	}

	return plan, nil
}

// ListPlans returns a paginated list of remediation plans.
func (rp *RemediationPlanner) ListPlans(ctx context.Context, orgID uuid.UUID, page, pageSize int) ([]RemediationPlan, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var total int64
	err := rp.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM remediation_plans
		WHERE organization_id = $1 AND deleted_at IS NULL
	`, orgID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count plans: %w", err)
	}

	rows, err := rp.pool.Query(ctx, `
		SELECT
			id, organization_id, plan_ref, name, description, plan_type, status,
			scope_framework_ids, scope_description, priority,
			ai_generated, ai_model, ai_prompt_summary, ai_generation_date, ai_confidence_score,
			human_reviewed, human_reviewed_by, human_reviewed_at,
			target_completion_date, estimated_total_hours, estimated_total_cost_eur,
			actual_completion_date, completion_percentage,
			owner_user_id, created_by, approved_by, approved_at,
			metadata, created_at, updated_at
		FROM remediation_plans
		WHERE organization_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, orgID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list plans: %w", err)
	}
	defer rows.Close()

	plans := make([]RemediationPlan, 0)
	for rows.Next() {
		var p RemediationPlan
		err := rows.Scan(
			&p.ID, &p.OrganizationID, &p.PlanRef, &p.Name, &p.Description,
			&p.PlanType, &p.Status,
			&p.ScopeFrameworkIDs, &p.ScopeDescription, &p.Priority,
			&p.AIGenerated, &p.AIModel, &p.AIPromptSummary, &p.AIGenerationDate, &p.AIConfidenceScore,
			&p.HumanReviewed, &p.HumanReviewedBy, &p.HumanReviewedAt,
			&p.TargetCompletionDate, &p.EstimatedTotalHours, &p.EstimatedTotalCost,
			&p.ActualCompletionDate, &p.CompletionPercentage,
			&p.OwnerUserID, &p.CreatedBy, &p.ApprovedBy, &p.ApprovedAt,
			&p.Metadata, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to scan plan row")
			continue
		}
		plans = append(plans, p)
	}

	return plans, total, nil
}

// UpdatePlan updates fields on an existing remediation plan.
func (rp *RemediationPlanner) UpdatePlan(ctx context.Context, orgID, planID uuid.UUID, req UpdatePlanRequest) error {
	tx, err := rp.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_org', $1, true)", orgID.String()); err != nil {
		return fmt.Errorf("failed to set tenant context: %w", err)
	}

	// Build dynamic update query
	setClauses := make([]string, 0)
	args := make([]interface{}, 0)
	argIdx := 1

	addClause := func(field string, val interface{}) {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", field, argIdx))
		args = append(args, val)
		argIdx++
	}

	if req.Name != nil {
		addClause("name", *req.Name)
	}
	if req.Description != nil {
		addClause("description", *req.Description)
	}
	if req.Priority != nil {
		validPriorities := map[string]bool{"critical": true, "high": true, "medium": true, "low": true}
		if validPriorities[*req.Priority] {
			addClause("priority", *req.Priority)
		}
	}
	if req.Status != nil {
		validStatuses := map[string]bool{
			"draft": true, "pending_approval": true, "approved": true,
			"in_progress": true, "on_hold": true, "completed": true, "cancelled": true,
		}
		if validStatuses[*req.Status] {
			addClause("status", *req.Status)
		}
	}
	if req.OwnerUserID != nil {
		addClause("owner_user_id", *req.OwnerUserID)
	}
	if req.TargetCompletionDate != nil {
		addClause("target_completion_date", *req.TargetCompletionDate)
	}
	if req.ScopeDescription != nil {
		addClause("scope_description", *req.ScopeDescription)
	}

	if len(setClauses) == 0 {
		return fmt.Errorf("no fields to update")
	}

	query := "UPDATE remediation_plans SET "
	for i, clause := range setClauses {
		if i > 0 {
			query += ", "
		}
		query += clause
	}
	query += fmt.Sprintf(" WHERE id = $%d AND organization_id = $%d AND deleted_at IS NULL", argIdx, argIdx+1)
	args = append(args, planID, orgID)

	tag, err := tx.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update plan: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("plan not found")
	}

	return tx.Commit(ctx)
}

// ApprovePlan marks a plan as approved by the given user.
func (rp *RemediationPlanner) ApprovePlan(ctx context.Context, orgID, planID, approverID uuid.UUID) error {
	tx, err := rp.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_org', $1, true)", orgID.String()); err != nil {
		return fmt.Errorf("failed to set tenant context: %w", err)
	}

	// Check plan is in pending_approval or draft status
	var currentStatus string
	err = tx.QueryRow(ctx, `
		SELECT status FROM remediation_plans
		WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL
	`, planID, orgID).Scan(&currentStatus)
	if err != nil {
		return fmt.Errorf("plan not found: %w", err)
	}

	if currentStatus != "draft" && currentStatus != "pending_approval" {
		return fmt.Errorf("plan cannot be approved from status '%s'", currentStatus)
	}

	now := time.Now()
	tag, err := tx.Exec(ctx, `
		UPDATE remediation_plans
		SET status = 'approved', approved_by = $1, approved_at = $2
		WHERE id = $3 AND organization_id = $4 AND deleted_at IS NULL
	`, approverID, now, planID, orgID)
	if err != nil {
		return fmt.Errorf("failed to approve plan: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("plan not found")
	}

	return tx.Commit(ctx)
}

// GetTimeline produces a timeline view of a plan's actions.
func (rp *RemediationPlanner) GetTimeline(ctx context.Context, orgID, planID uuid.UUID) (*PlanTimeline, error) {
	plan, err := rp.GetPlan(ctx, orgID, planID)
	if err != nil {
		return nil, err
	}

	timeline := &PlanTimeline{
		PlanID:   plan.ID,
		PlanRef:  plan.PlanRef,
		PlanName: plan.Name,
	}

	var earliestStart, latestEnd *time.Time
	entries := make([]RemediationTimelineEntry, 0, len(plan.Actions))

	for _, action := range plan.Actions {
		entry := RemediationTimelineEntry{
			ActionID:     action.ID,
			ActionRef:    action.ActionRef,
			Title:        action.Title,
			StartDate:    action.TargetStartDate,
			EndDate:      action.TargetEndDate,
			Status:       action.Status,
			Priority:     action.Priority,
			AssignedTo:   action.AssignedTo,
			Dependencies: action.Dependencies,
		}
		entries = append(entries, entry)

		if action.TargetStartDate != nil {
			if earliestStart == nil || action.TargetStartDate.Before(*earliestStart) {
				earliestStart = action.TargetStartDate
			}
		}
		if action.TargetEndDate != nil {
			if latestEnd == nil || action.TargetEndDate.After(*latestEnd) {
				latestEnd = action.TargetEndDate
			}
		}
	}

	timeline.StartDate = earliestStart
	timeline.EndDate = latestEnd
	timeline.TimelineEntries = entries

	if earliestStart != nil && latestEnd != nil {
		weeks := int(math.Ceil(latestEnd.Sub(*earliestStart).Hours() / (24 * 7)))
		timeline.TotalWeeks = weeks
	}

	// Build critical path (high/critical priority actions that are not completed)
	criticalPath := make([]uuid.UUID, 0)
	for _, action := range plan.Actions {
		if (action.Priority == "critical" || action.Priority == "high") &&
			action.Status != "completed" && action.Status != "cancelled" {
			criticalPath = append(criticalPath, action.ID)
		}
	}
	timeline.CriticalPath = criticalPath

	// Generate milestones
	milestones := make([]TimelineMilestone, 0)
	if earliestStart != nil {
		milestones = append(milestones, TimelineMilestone{
			Date:        *earliestStart,
			Title:       "Plan Start",
			Description: "Remediation plan execution begins",
		})
	}
	if plan.TargetCompletionDate != nil {
		milestones = append(milestones, TimelineMilestone{
			Date:        *plan.TargetCompletionDate,
			Title:       "Target Completion",
			Description: "Target date for plan completion",
		})
	}
	timeline.Milestones = milestones

	return timeline, nil
}

// TrackProgress computes detailed progress metrics for a plan.
func (rp *RemediationPlanner) TrackProgress(ctx context.Context, orgID, planID uuid.UUID) (*PlanProgress, error) {
	plan, err := rp.GetPlan(ctx, orgID, planID)
	if err != nil {
		return nil, err
	}

	progress := &PlanProgress{
		PlanID:            plan.ID,
		PlanRef:           plan.PlanRef,
		TotalActions:      len(plan.Actions),
		StatusBreakdown:   make(map[string]int),
		PriorityBreakdown: make(map[string]int),
	}

	for _, action := range plan.Actions {
		progress.StatusBreakdown[action.Status]++
		progress.PriorityBreakdown[action.Priority]++

		switch action.Status {
		case "completed":
			progress.CompletedActions++
		case "in_progress", "in_review":
			progress.InProgressActions++
		case "blocked":
			progress.BlockedActions++
		case "pending", "assigned":
			progress.PendingActions++
		case "deferred":
			progress.DeferredActions++
		case "cancelled":
			progress.CancelledActions++
		}

		progress.EstimatedHoursTotal += action.EstimatedHours
		progress.EstimatedCostTotal += action.EstimatedCostEUR

		if action.ActualHours != nil {
			progress.ActualHoursTotal += *action.ActualHours
		}
		if action.ActualCostEUR != nil {
			progress.ActualCostTotal += *action.ActualCostEUR
		}

		// Critical path: non-completed high/critical actions
		if (action.Priority == "critical" || action.Priority == "high") &&
			action.Status != "completed" && action.Status != "cancelled" {
			daysLeft := 0
			if action.TargetEndDate != nil {
				daysLeft = int(time.Until(*action.TargetEndDate).Hours() / 24)
			}
			progress.CriticalPathActions = append(progress.CriticalPathActions, CriticalPathAction{
				ActionID:  action.ID,
				ActionRef: action.ActionRef,
				Title:     action.Title,
				Status:    action.Status,
				Priority:  action.Priority,
				DaysLeft:  daysLeft,
			})
		}
	}

	// Calculate completion percentage
	actionableActions := progress.TotalActions - progress.CancelledActions
	if actionableActions > 0 {
		progress.CompletionPercentage = math.Round(float64(progress.CompletedActions)/float64(actionableActions)*10000) / 100
	}

	// Update plan completion percentage in DB
	_, updateErr := rp.pool.Exec(ctx, `
		UPDATE remediation_plans SET completion_percentage = $1
		WHERE id = $2 AND organization_id = $3
	`, progress.CompletionPercentage, planID, orgID)
	if updateErr != nil {
		log.Warn().Err(updateErr).Msg("Failed to update plan completion percentage")
	}

	// On-track calculation
	if plan.TargetCompletionDate != nil {
		remaining := time.Until(*plan.TargetCompletionDate)
		progress.DaysRemaining = int(remaining.Hours() / 24)
		if progress.DaysRemaining < 0 {
			progress.DaysOverdue = -progress.DaysRemaining
			progress.DaysRemaining = 0
		}

		// On track if completion percentage is proportional to elapsed time
		if plan.CreatedAt.Before(time.Now()) {
			totalDuration := plan.TargetCompletionDate.Sub(plan.CreatedAt).Hours()
			elapsed := time.Since(plan.CreatedAt).Hours()
			if totalDuration > 0 {
				expectedCompletion := (elapsed / totalDuration) * 100
				progress.OnTrack = progress.CompletionPercentage >= expectedCompletion*0.8 // 80% of expected
			}
		}
	} else {
		progress.OnTrack = progress.BlockedActions == 0
	}

	return progress, nil
}

// UpdateAction updates fields on an existing action.
func (rp *RemediationPlanner) UpdateAction(ctx context.Context, orgID, actionID uuid.UUID, req UpdateActionRequest) error {
	tx, err := rp.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_org', $1, true)", orgID.String()); err != nil {
		return fmt.Errorf("failed to set tenant context: %w", err)
	}

	setClauses := make([]string, 0)
	args := make([]interface{}, 0)
	argIdx := 1

	addClause := func(field string, val interface{}) {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", field, argIdx))
		args = append(args, val)
		argIdx++
	}

	if req.Title != nil {
		addClause("title", *req.Title)
	}
	if req.Description != nil {
		addClause("description", *req.Description)
	}
	if req.Priority != nil {
		validPriorities := map[string]bool{"critical": true, "high": true, "medium": true, "low": true}
		if validPriorities[*req.Priority] {
			addClause("priority", *req.Priority)
		}
	}
	if req.Status != nil {
		validStatuses := map[string]bool{
			"pending": true, "assigned": true, "in_progress": true,
			"blocked": true, "in_review": true, "completed": true,
			"deferred": true, "cancelled": true,
		}
		if validStatuses[*req.Status] {
			addClause("status", *req.Status)

			// Auto-set actual_start_date when status moves to in_progress
			if *req.Status == "in_progress" {
				now := time.Now()
				addClause("actual_start_date", now)
			}
		}
	}
	if req.AssignedTo != nil {
		addClause("assigned_to", *req.AssignedTo)
	}
	if req.TargetStartDate != nil {
		addClause("target_start_date", *req.TargetStartDate)
	}
	if req.TargetEndDate != nil {
		addClause("target_end_date", *req.TargetEndDate)
	}
	if req.SortOrder != nil {
		addClause("sort_order", *req.SortOrder)
	}

	if len(setClauses) == 0 {
		return fmt.Errorf("no fields to update")
	}

	query := "UPDATE remediation_actions SET "
	for i, clause := range setClauses {
		if i > 0 {
			query += ", "
		}
		query += clause
	}
	query += fmt.Sprintf(" WHERE id = $%d AND organization_id = $%d AND deleted_at IS NULL", argIdx, argIdx+1)
	args = append(args, actionID, orgID)

	tag, err := tx.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update action: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("action not found")
	}

	return tx.Commit(ctx)
}

// CompleteAction marks an action as complete with evidence and actuals.
func (rp *RemediationPlanner) CompleteAction(ctx context.Context, orgID, actionID uuid.UUID, req CompleteActionRequest) error {
	tx, err := rp.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_org', $1, true)", orgID.String()); err != nil {
		return fmt.Errorf("failed to set tenant context: %w", err)
	}

	now := time.Now()
	evidencePaths := req.EvidencePaths
	if evidencePaths == nil {
		evidencePaths = []string{}
	}

	tag, err := tx.Exec(ctx, `
		UPDATE remediation_actions
		SET status = 'completed',
			actual_end_date = $1,
			actual_hours = $2,
			actual_cost_eur = $3,
			completion_notes = $4,
			evidence_paths = $5
		WHERE id = $6 AND organization_id = $7 AND deleted_at IS NULL
	`, now, req.ActualHours, req.ActualCostEUR, req.CompletionNotes, evidencePaths, actionID, orgID)
	if err != nil {
		return fmt.Errorf("failed to complete action: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("action not found")
	}

	// Auto-update plan completion percentage
	var planID uuid.UUID
	err = tx.QueryRow(ctx, `
		SELECT plan_id FROM remediation_actions WHERE id = $1
	`, actionID).Scan(&planID)
	if err == nil {
		_, _ = tx.Exec(ctx, `
			UPDATE remediation_plans
			SET completion_percentage = (
				SELECT COALESCE(
					ROUND(COUNT(*) FILTER (WHERE status = 'completed')::numeric /
						NULLIF(COUNT(*) FILTER (WHERE status != 'cancelled'), 0)::numeric * 100, 2),
					0
				)
				FROM remediation_actions
				WHERE plan_id = $1 AND deleted_at IS NULL
			),
			actual_completion_date = CASE
				WHEN (
					SELECT COUNT(*) FILTER (WHERE status NOT IN ('completed', 'cancelled'))
					FROM remediation_actions
					WHERE plan_id = $1 AND deleted_at IS NULL
				) = 0 THEN now()
				ELSE NULL
			END,
			status = CASE
				WHEN (
					SELECT COUNT(*) FILTER (WHERE status NOT IN ('completed', 'cancelled'))
					FROM remediation_actions
					WHERE plan_id = $1 AND deleted_at IS NULL
				) = 0 THEN 'completed'::remediation_plan_status
				ELSE status
			END
			WHERE id = $1 AND organization_id = $2
		`, planID, orgID)
	}

	return tx.Commit(ctx)
}

// getActionsByPlanID retrieves all actions for a plan.
func (rp *RemediationPlanner) getActionsByPlanID(ctx context.Context, orgID, planID uuid.UUID) ([]RemediationAction, error) {
	rows, err := rp.pool.Query(ctx, `
		SELECT
			id, organization_id, plan_id, action_ref, sort_order, title, description,
			action_type, linked_control_implementation_id, linked_finding_id,
			linked_risk_treatment_id, framework_control_code, priority,
			estimated_hours, estimated_cost_eur, required_skills, dependencies,
			assigned_to, target_start_date, target_end_date, status,
			actual_start_date, actual_end_date, actual_hours, actual_cost_eur,
			completion_notes, evidence_paths,
			ai_implementation_guidance, ai_evidence_suggestions,
			ai_tool_recommendations, ai_risk_if_deferred, ai_cross_framework_benefit,
			created_at, updated_at
		FROM remediation_actions
		WHERE plan_id = $1 AND organization_id = $2 AND deleted_at IS NULL
		ORDER BY sort_order ASC, created_at ASC
	`, planID, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to query actions: %w", err)
	}
	defer rows.Close()

	actions := make([]RemediationAction, 0)
	for rows.Next() {
		var a RemediationAction
		err := rows.Scan(
			&a.ID, &a.OrganizationID, &a.PlanID, &a.ActionRef, &a.SortOrder,
			&a.Title, &a.Description,
			&a.ActionType, &a.LinkedControlImplementationID, &a.LinkedFindingID,
			&a.LinkedRiskTreatmentID, &a.FrameworkControlCode, &a.Priority,
			&a.EstimatedHours, &a.EstimatedCostEUR, &a.RequiredSkills, &a.Dependencies,
			&a.AssignedTo, &a.TargetStartDate, &a.TargetEndDate, &a.Status,
			&a.ActualStartDate, &a.ActualEndDate, &a.ActualHours, &a.ActualCostEUR,
			&a.CompletionNotes, &a.EvidencePaths,
			&a.AIImplementationGuidance, &a.AIEvidenceSuggestions,
			&a.AIToolRecommendations, &a.AIRiskIfDeferred, &a.AICrossFrameworkBenefit,
			&a.CreatedAt, &a.UpdatedAt,
		)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to scan action row")
			continue
		}
		actions = append(actions, a)
	}

	return actions, nil
}

// ============================================================
// HELPERS
// ============================================================

// nullableString returns nil for empty strings, for nullable DB columns.
func nullableString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
