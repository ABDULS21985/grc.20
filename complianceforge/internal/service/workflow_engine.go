// Package service contains the Compliance Workflow Engine for ComplianceForge.
// It provides a full-featured workflow orchestration system supporting conditional routing,
// parallel approval gates, delegation rules, SLA tracking, and automatic step advancement.
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
// DOMAIN TYPES
// ============================================================

// WorkflowDefinition represents a reusable workflow template.
type WorkflowDefinition struct {
	ID                uuid.UUID       `json:"id"`
	OrganizationID    *uuid.UUID      `json:"organization_id,omitempty"`
	Name              string          `json:"name"`
	Description       string          `json:"description"`
	WorkflowType      string          `json:"workflow_type"`
	EntityType        string          `json:"entity_type"`
	Version           int             `json:"version"`
	Status            string          `json:"status"`
	TriggerConditions json.RawMessage `json:"trigger_conditions"`
	SLAConfig         json.RawMessage `json:"sla_config"`
	Metadata          json.RawMessage `json:"metadata"`
	CreatedBy         *uuid.UUID      `json:"created_by,omitempty"`
	IsSystem          bool            `json:"is_system"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
	Steps             []WorkflowStep  `json:"steps,omitempty"`
}

// WorkflowStep represents a single step within a workflow definition.
type WorkflowStep struct {
	ID                     uuid.UUID       `json:"id"`
	WorkflowDefinitionID   uuid.UUID       `json:"workflow_definition_id"`
	OrganizationID         *uuid.UUID      `json:"organization_id,omitempty"`
	StepOrder              int             `json:"step_order"`
	Name                   string          `json:"name"`
	Description            string          `json:"description"`
	StepType               string          `json:"step_type"`
	ApproverType           *string         `json:"approver_type,omitempty"`
	ApproverIDs            []uuid.UUID     `json:"approver_ids,omitempty"`
	ApprovalMode           string          `json:"approval_mode"`
	MinimumApprovals       int             `json:"minimum_approvals"`
	TaskDescription        string          `json:"task_description,omitempty"`
	TaskAssigneeType       *string         `json:"task_assignee_type,omitempty"`
	TaskAssigneeIDs        []uuid.UUID     `json:"task_assignee_ids,omitempty"`
	ConditionExpression    json.RawMessage `json:"condition_expression,omitempty"`
	ConditionTrueStepID    *uuid.UUID      `json:"condition_true_step_id,omitempty"`
	ConditionFalseStepID   *uuid.UUID      `json:"condition_false_step_id,omitempty"`
	AutoAction             json.RawMessage `json:"auto_action,omitempty"`
	TimerHours             *int            `json:"timer_hours,omitempty"`
	TimerBusinessHoursOnly bool            `json:"timer_business_hours_only"`
	SLAHours               *int            `json:"sla_hours,omitempty"`
	EscalationUserIDs      []uuid.UUID     `json:"escalation_user_ids,omitempty"`
	IsOptional             bool            `json:"is_optional"`
	CanDelegate            bool            `json:"can_delegate"`
	CreatedAt              time.Time       `json:"created_at"`
	UpdatedAt              time.Time       `json:"updated_at"`
}

// WorkflowInstance represents a running instance of a workflow attached to an entity.
type WorkflowInstance struct {
	ID                   uuid.UUID       `json:"id"`
	OrganizationID       uuid.UUID       `json:"organization_id"`
	WorkflowDefinitionID uuid.UUID       `json:"workflow_definition_id"`
	EntityType           string          `json:"entity_type"`
	EntityID             uuid.UUID       `json:"entity_id"`
	EntityRef            string          `json:"entity_ref"`
	Status               string          `json:"status"`
	CurrentStepID        *uuid.UUID      `json:"current_step_id,omitempty"`
	CurrentStepOrder     *int            `json:"current_step_order,omitempty"`
	StartedAt            time.Time       `json:"started_at"`
	StartedBy            *uuid.UUID      `json:"started_by,omitempty"`
	CompletedAt          *time.Time      `json:"completed_at,omitempty"`
	CompletionOutcome    *string         `json:"completion_outcome,omitempty"`
	TotalDurationHours   *float64        `json:"total_duration_hours,omitempty"`
	SLAStatus            string          `json:"sla_status"`
	SLADeadline          *time.Time      `json:"sla_deadline,omitempty"`
	Metadata             json.RawMessage `json:"metadata"`
	CreatedAt            time.Time       `json:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at"`
	DefinitionName       string          `json:"definition_name,omitempty"`
	CurrentStepName      string          `json:"current_step_name,omitempty"`
	Executions           []StepExecution `json:"executions,omitempty"`
}

// StepExecution represents the execution state of a single workflow step.
type StepExecution struct {
	ID                 uuid.UUID  `json:"id"`
	OrganizationID     uuid.UUID  `json:"organization_id"`
	WorkflowInstanceID uuid.UUID  `json:"workflow_instance_id"`
	WorkflowStepID     uuid.UUID  `json:"workflow_step_id"`
	StepOrder          int        `json:"step_order"`
	Status             string     `json:"status"`
	AssignedTo         *uuid.UUID `json:"assigned_to,omitempty"`
	DelegatedTo        *uuid.UUID `json:"delegated_to,omitempty"`
	DelegatedBy        *uuid.UUID `json:"delegated_by,omitempty"`
	DelegatedAt        *time.Time `json:"delegated_at,omitempty"`
	ActionTakenBy      *uuid.UUID `json:"action_taken_by,omitempty"`
	ActionTakenAt      *time.Time `json:"action_taken_at,omitempty"`
	Action             *string    `json:"action,omitempty"`
	Comments           string     `json:"comments"`
	DecisionReason     string     `json:"decision_reason"`
	AttachmentsPaths   []string   `json:"attachments_paths,omitempty"`
	SLADeadline        *time.Time `json:"sla_deadline,omitempty"`
	SLAStatus          string     `json:"sla_status"`
	EscalatedAt        *time.Time `json:"escalated_at,omitempty"`
	EscalatedTo        *uuid.UUID `json:"escalated_to,omitempty"`
	StartedAt          *time.Time `json:"started_at,omitempty"`
	CompletedAt        *time.Time `json:"completed_at,omitempty"`
	DurationHours      *float64   `json:"duration_hours,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	StepName           string     `json:"step_name,omitempty"`
	StepType           string     `json:"step_type,omitempty"`
	EntityRef          string     `json:"entity_ref,omitempty"`
	EntityType         string     `json:"entity_type,omitempty"`
	InstanceStatus     string     `json:"instance_status,omitempty"`
}

// DelegationRule defines an out-of-office or delegation rule for workflow approvals.
type DelegationRule struct {
	ID              uuid.UUID  `json:"id"`
	OrganizationID  uuid.UUID  `json:"organization_id"`
	DelegatorUserID uuid.UUID  `json:"delegator_user_id"`
	DelegateUserID  uuid.UUID  `json:"delegate_user_id"`
	WorkflowTypes   []string   `json:"workflow_types"`
	ValidFrom       time.Time  `json:"valid_from"`
	ValidUntil      *time.Time `json:"valid_until,omitempty"`
	Reason          string     `json:"reason"`
	IsActive        bool       `json:"is_active"`
	CreatedBy       *uuid.UUID `json:"created_by,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// ============================================================
// INPUT TYPES
// ============================================================

// CreateDefinitionInput holds the fields for creating a new workflow definition.
type CreateDefinitionInput struct {
	Name              string          `json:"name"`
	Description       string          `json:"description"`
	WorkflowType      string          `json:"workflow_type"`
	EntityType        string          `json:"entity_type"`
	TriggerConditions json.RawMessage `json:"trigger_conditions"`
	SLAConfig         json.RawMessage `json:"sla_config"`
	Metadata          json.RawMessage `json:"metadata"`
}

// UpdateDefinitionInput holds the fields for updating an existing workflow definition.
type UpdateDefinitionInput struct {
	Name              *string          `json:"name,omitempty"`
	Description       *string          `json:"description,omitempty"`
	TriggerConditions *json.RawMessage `json:"trigger_conditions,omitempty"`
	SLAConfig         *json.RawMessage `json:"sla_config,omitempty"`
	Metadata          *json.RawMessage `json:"metadata,omitempty"`
}

// StepInput holds the fields for creating or updating a workflow step.
type StepInput struct {
	StepOrder              int             `json:"step_order"`
	Name                   string          `json:"name"`
	Description            string          `json:"description"`
	StepType               string          `json:"step_type"`
	ApproverType           *string         `json:"approver_type,omitempty"`
	ApproverIDs            []uuid.UUID     `json:"approver_ids,omitempty"`
	ApprovalMode           string          `json:"approval_mode"`
	MinimumApprovals       int             `json:"minimum_approvals"`
	TaskDescription        string          `json:"task_description,omitempty"`
	TaskAssigneeType       *string         `json:"task_assignee_type,omitempty"`
	TaskAssigneeIDs        []uuid.UUID     `json:"task_assignee_ids,omitempty"`
	ConditionExpression    json.RawMessage `json:"condition_expression,omitempty"`
	ConditionTrueStepID    *uuid.UUID      `json:"condition_true_step_id,omitempty"`
	ConditionFalseStepID   *uuid.UUID      `json:"condition_false_step_id,omitempty"`
	AutoAction             json.RawMessage `json:"auto_action,omitempty"`
	TimerHours             *int            `json:"timer_hours,omitempty"`
	TimerBusinessHoursOnly bool            `json:"timer_business_hours_only"`
	SLAHours               *int            `json:"sla_hours,omitempty"`
	EscalationUserIDs      []uuid.UUID     `json:"escalation_user_ids,omitempty"`
	IsOptional             bool            `json:"is_optional"`
	CanDelegate            bool            `json:"can_delegate"`
}

// CreateDelegationInput holds the fields for creating a delegation rule.
type CreateDelegationInput struct {
	DelegatorUserID uuid.UUID  `json:"delegator_user_id"`
	DelegateUserID  uuid.UUID  `json:"delegate_user_id"`
	WorkflowTypes   []string   `json:"workflow_types"`
	ValidFrom       time.Time  `json:"valid_from"`
	ValidUntil      *time.Time `json:"valid_until,omitempty"`
	Reason          string     `json:"reason"`
}

// InstanceFilters holds optional filters for listing workflow instances.
type InstanceFilters struct {
	Status     string `json:"status"`
	EntityType string `json:"entity_type"`
	SLAStatus  string `json:"sla_status"`
}

// ConditionExpression represents a parsed conditional expression for routing.
type ConditionExpression struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

// SLAConfig holds the SLA configuration parsed from a workflow definition.
type SLAConfig struct {
	TotalSLAHours       int `json:"total_sla_hours"`
	StepDefaultSLAHours int `json:"step_default_sla_hours"`
}

// ============================================================
// WORKFLOW ENGINE
// ============================================================

// WorkflowEngine orchestrates compliance workflow execution including
// step resolution, conditional routing, parallel gates, delegation, and SLA tracking.
type WorkflowEngine struct {
	pool *pgxpool.Pool
}

// NewWorkflowEngine creates a new WorkflowEngine.
func NewWorkflowEngine(pool *pgxpool.Pool) *WorkflowEngine {
	return &WorkflowEngine{pool: pool}
}

// Pool exposes the connection pool for background workers.
func (e *WorkflowEngine) Pool() *pgxpool.Pool {
	return e.pool
}

// ============================================================
// DEFINITION CRUD
// ============================================================

// ListDefinitions returns paginated workflow definitions for an organization, including system definitions.
func (e *WorkflowEngine) ListDefinitions(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]WorkflowDefinition, int64, error) {
	var total int64
	err := e.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM workflow_definitions
		WHERE (organization_id = $1 OR organization_id IS NULL)`,
		orgID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count workflow definitions: %w", err)
	}

	rows, err := e.pool.Query(ctx, `
		SELECT id, organization_id, name, description, workflow_type::text, entity_type,
			   version, status::text, trigger_conditions, sla_config, metadata,
			   created_by, is_system, created_at, updated_at
		FROM workflow_definitions
		WHERE (organization_id = $1 OR organization_id IS NULL)
		ORDER BY is_system DESC, name ASC
		LIMIT $2 OFFSET $3`,
		orgID, limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list workflow definitions: %w", err)
	}
	defer rows.Close()

	var defs []WorkflowDefinition
	for rows.Next() {
		var d WorkflowDefinition
		if err := rows.Scan(
			&d.ID, &d.OrganizationID, &d.Name, &d.Description, &d.WorkflowType, &d.EntityType,
			&d.Version, &d.Status, &d.TriggerConditions, &d.SLAConfig, &d.Metadata,
			&d.CreatedBy, &d.IsSystem, &d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan workflow definition: %w", err)
		}
		defs = append(defs, d)
	}

	return defs, total, nil
}

// GetDefinition returns a single workflow definition by ID.
func (e *WorkflowEngine) GetDefinition(ctx context.Context, orgID, defID uuid.UUID) (*WorkflowDefinition, error) {
	var d WorkflowDefinition
	err := e.pool.QueryRow(ctx, `
		SELECT id, organization_id, name, description, workflow_type::text, entity_type,
			   version, status::text, trigger_conditions, sla_config, metadata,
			   created_by, is_system, created_at, updated_at
		FROM workflow_definitions
		WHERE id = $1 AND (organization_id = $2 OR organization_id IS NULL)`,
		defID, orgID,
	).Scan(
		&d.ID, &d.OrganizationID, &d.Name, &d.Description, &d.WorkflowType, &d.EntityType,
		&d.Version, &d.Status, &d.TriggerConditions, &d.SLAConfig, &d.Metadata,
		&d.CreatedBy, &d.IsSystem, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("workflow definition not found: %w", err)
	}
	return &d, nil
}

// CreateDefinition creates a new workflow definition.
func (e *WorkflowEngine) CreateDefinition(ctx context.Context, orgID uuid.UUID, input CreateDefinitionInput) (*WorkflowDefinition, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("workflow definition name is required")
	}
	if input.WorkflowType == "" {
		return nil, fmt.Errorf("workflow type is required")
	}
	if input.EntityType == "" {
		return nil, fmt.Errorf("entity type is required")
	}

	triggerCond := defaultJSONB(input.TriggerConditions)
	slaCfg := defaultJSONB(input.SLAConfig)
	meta := defaultJSONB(input.Metadata)

	var d WorkflowDefinition
	err := e.pool.QueryRow(ctx, `
		INSERT INTO workflow_definitions
			(organization_id, name, description, workflow_type, entity_type, version, status,
			 trigger_conditions, sla_config, metadata, is_system)
		VALUES ($1, $2, $3, $4::workflow_type, $5, 1, 'draft', $6, $7, $8, false)
		RETURNING id, organization_id, name, description, workflow_type::text, entity_type,
				  version, status::text, trigger_conditions, sla_config, metadata,
				  created_by, is_system, created_at, updated_at`,
		orgID, input.Name, input.Description, input.WorkflowType, input.EntityType,
		triggerCond, slaCfg, meta,
	).Scan(
		&d.ID, &d.OrganizationID, &d.Name, &d.Description, &d.WorkflowType, &d.EntityType,
		&d.Version, &d.Status, &d.TriggerConditions, &d.SLAConfig, &d.Metadata,
		&d.CreatedBy, &d.IsSystem, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create workflow definition: %w", err)
	}

	log.Info().Str("id", d.ID.String()).Str("name", d.Name).Msg("Workflow definition created")
	return &d, nil
}

// UpdateDefinition updates an existing workflow definition. System definitions cannot be modified.
func (e *WorkflowEngine) UpdateDefinition(ctx context.Context, orgID, defID uuid.UUID, input UpdateDefinitionInput) (*WorkflowDefinition, error) {
	// Check system definition
	var isSystem bool
	err := e.pool.QueryRow(ctx,
		`SELECT is_system FROM workflow_definitions WHERE id = $1 AND (organization_id = $2 OR organization_id IS NULL)`,
		defID, orgID,
	).Scan(&isSystem)
	if err != nil {
		return nil, fmt.Errorf("workflow definition not found: %w", err)
	}
	if isSystem {
		return nil, fmt.Errorf("system workflow definitions cannot be modified")
	}

	// Build dynamic SET clause
	setClauses := []string{}
	args := []interface{}{defID, orgID}
	argIdx := 3

	if input.Name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, *input.Name)
		argIdx++
	}
	if input.Description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", argIdx))
		args = append(args, *input.Description)
		argIdx++
	}
	if input.TriggerConditions != nil {
		setClauses = append(setClauses, fmt.Sprintf("trigger_conditions = $%d", argIdx))
		args = append(args, *input.TriggerConditions)
		argIdx++
	}
	if input.SLAConfig != nil {
		setClauses = append(setClauses, fmt.Sprintf("sla_config = $%d", argIdx))
		args = append(args, *input.SLAConfig)
		argIdx++
	}
	if input.Metadata != nil {
		setClauses = append(setClauses, fmt.Sprintf("metadata = $%d", argIdx))
		args = append(args, *input.Metadata)
		argIdx++
	}

	if len(setClauses) == 0 {
		return e.GetDefinition(ctx, orgID, defID)
	}

	query := fmt.Sprintf(`
		UPDATE workflow_definitions SET %s, version = version + 1
		WHERE id = $1 AND organization_id = $2
		RETURNING id, organization_id, name, description, workflow_type::text, entity_type,
				  version, status::text, trigger_conditions, sla_config, metadata,
				  created_by, is_system, created_at, updated_at`,
		strings.Join(setClauses, ", "),
	)

	var d WorkflowDefinition
	err = e.pool.QueryRow(ctx, query, args...).Scan(
		&d.ID, &d.OrganizationID, &d.Name, &d.Description, &d.WorkflowType, &d.EntityType,
		&d.Version, &d.Status, &d.TriggerConditions, &d.SLAConfig, &d.Metadata,
		&d.CreatedBy, &d.IsSystem, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update workflow definition: %w", err)
	}

	return &d, nil
}

// ActivateDefinition sets a workflow definition's status to 'active', validating it has at least one step.
func (e *WorkflowEngine) ActivateDefinition(ctx context.Context, orgID, defID uuid.UUID) error {
	// Validate the definition has steps
	var stepCount int
	err := e.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM workflow_steps WHERE workflow_definition_id = $1`,
		defID,
	).Scan(&stepCount)
	if err != nil {
		return fmt.Errorf("failed to count steps: %w", err)
	}
	if stepCount == 0 {
		return fmt.Errorf("workflow definition must have at least one step before activation")
	}

	tag, err := e.pool.Exec(ctx, `
		UPDATE workflow_definitions SET status = 'active'
		WHERE id = $1 AND (organization_id = $2 OR organization_id IS NULL)
			AND status = 'draft'`,
		defID, orgID,
	)
	if err != nil {
		return fmt.Errorf("failed to activate workflow definition: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("definition not found, already active, or not in draft status")
	}

	log.Info().Str("definition_id", defID.String()).Msg("Workflow definition activated")
	return nil
}

// ============================================================
// STEP CRUD
// ============================================================

// GetDefinitionSteps returns all steps for a workflow definition, ordered by step_order.
func (e *WorkflowEngine) GetDefinitionSteps(ctx context.Context, orgID, defID uuid.UUID) ([]WorkflowStep, error) {
	rows, err := e.pool.Query(ctx, `
		SELECT id, workflow_definition_id, organization_id, step_order, name, description,
			   step_type::text, approver_type::text, approver_ids, approval_mode::text,
			   minimum_approvals, task_description, task_assignee_type::text, task_assignee_ids,
			   condition_expression, condition_true_step_id, condition_false_step_id,
			   auto_action, timer_hours, timer_business_hours_only,
			   sla_hours, escalation_user_ids, is_optional, can_delegate,
			   created_at, updated_at
		FROM workflow_steps
		WHERE workflow_definition_id = $1
			AND (organization_id = $2 OR organization_id IS NULL)
		ORDER BY step_order ASC`,
		defID, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list workflow steps: %w", err)
	}
	defer rows.Close()

	var steps []WorkflowStep
	for rows.Next() {
		s, err := scanWorkflowStep(rows)
		if err != nil {
			return nil, err
		}
		steps = append(steps, s)
	}

	return steps, nil
}

// AddStep adds a new step to a workflow definition.
func (e *WorkflowEngine) AddStep(ctx context.Context, orgID, defID uuid.UUID, input StepInput) (*WorkflowStep, error) {
	// Verify definition exists and is editable
	var status string
	var isSystem bool
	err := e.pool.QueryRow(ctx,
		`SELECT status::text, is_system FROM workflow_definitions WHERE id = $1 AND (organization_id = $2 OR organization_id IS NULL)`,
		defID, orgID,
	).Scan(&status, &isSystem)
	if err != nil {
		return nil, fmt.Errorf("workflow definition not found: %w", err)
	}
	if isSystem {
		return nil, fmt.Errorf("cannot add steps to system workflow definitions")
	}

	condExpr := defaultJSONB(input.ConditionExpression)
	autoAct := defaultJSONB(input.AutoAction)

	var s WorkflowStep
	err = e.pool.QueryRow(ctx, `
		INSERT INTO workflow_steps
			(workflow_definition_id, organization_id, step_order, name, description, step_type,
			 approver_type, approver_ids, approval_mode, minimum_approvals,
			 task_description, task_assignee_type, task_assignee_ids,
			 condition_expression, condition_true_step_id, condition_false_step_id,
			 auto_action, timer_hours, timer_business_hours_only,
			 sla_hours, escalation_user_ids, is_optional, can_delegate)
		VALUES ($1, $2, $3, $4, $5, $6::workflow_step_type,
				$7::workflow_approver_type, $8, $9::workflow_approval_mode, $10,
				$11, $12::workflow_task_assignee_type, $13,
				$14, $15, $16,
				$17, $18, $19,
				$20, $21, $22, $23)
		RETURNING id, workflow_definition_id, organization_id, step_order, name, description,
				  step_type::text, approver_type::text, approver_ids, approval_mode::text,
				  minimum_approvals, task_description, task_assignee_type::text, task_assignee_ids,
				  condition_expression, condition_true_step_id, condition_false_step_id,
				  auto_action, timer_hours, timer_business_hours_only,
				  sla_hours, escalation_user_ids, is_optional, can_delegate,
				  created_at, updated_at`,
		defID, orgID, input.StepOrder, input.Name, input.Description, input.StepType,
		input.ApproverType, input.ApproverIDs, nullableApprovalMode(input.ApprovalMode), input.MinimumApprovals,
		input.TaskDescription, input.TaskAssigneeType, input.TaskAssigneeIDs,
		condExpr, input.ConditionTrueStepID, input.ConditionFalseStepID,
		autoAct, input.TimerHours, input.TimerBusinessHoursOnly,
		input.SLAHours, input.EscalationUserIDs, input.IsOptional, input.CanDelegate,
	).Scan(
		&s.ID, &s.WorkflowDefinitionID, &s.OrganizationID, &s.StepOrder, &s.Name, &s.Description,
		&s.StepType, &s.ApproverType, &s.ApproverIDs, &s.ApprovalMode,
		&s.MinimumApprovals, &s.TaskDescription, &s.TaskAssigneeType, &s.TaskAssigneeIDs,
		&s.ConditionExpression, &s.ConditionTrueStepID, &s.ConditionFalseStepID,
		&s.AutoAction, &s.TimerHours, &s.TimerBusinessHoursOnly,
		&s.SLAHours, &s.EscalationUserIDs, &s.IsOptional, &s.CanDelegate,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add workflow step: %w", err)
	}

	return &s, nil
}

// UpdateStep updates an existing workflow step.
func (e *WorkflowEngine) UpdateStep(ctx context.Context, orgID, defID, stepID uuid.UUID, input StepInput) (*WorkflowStep, error) {
	condExpr := defaultJSONB(input.ConditionExpression)
	autoAct := defaultJSONB(input.AutoAction)

	var s WorkflowStep
	err := e.pool.QueryRow(ctx, `
		UPDATE workflow_steps SET
			step_order = $4, name = $5, description = $6, step_type = $7::workflow_step_type,
			approver_type = $8::workflow_approver_type, approver_ids = $9,
			approval_mode = $10::workflow_approval_mode, minimum_approvals = $11,
			task_description = $12, task_assignee_type = $13::workflow_task_assignee_type, task_assignee_ids = $14,
			condition_expression = $15, condition_true_step_id = $16, condition_false_step_id = $17,
			auto_action = $18, timer_hours = $19, timer_business_hours_only = $20,
			sla_hours = $21, escalation_user_ids = $22, is_optional = $23, can_delegate = $24
		WHERE id = $1 AND workflow_definition_id = $2
			AND (organization_id = $3 OR organization_id IS NULL)
		RETURNING id, workflow_definition_id, organization_id, step_order, name, description,
				  step_type::text, approver_type::text, approver_ids, approval_mode::text,
				  minimum_approvals, task_description, task_assignee_type::text, task_assignee_ids,
				  condition_expression, condition_true_step_id, condition_false_step_id,
				  auto_action, timer_hours, timer_business_hours_only,
				  sla_hours, escalation_user_ids, is_optional, can_delegate,
				  created_at, updated_at`,
		stepID, defID, orgID,
		input.StepOrder, input.Name, input.Description, input.StepType,
		input.ApproverType, input.ApproverIDs,
		nullableApprovalMode(input.ApprovalMode), input.MinimumApprovals,
		input.TaskDescription, input.TaskAssigneeType, input.TaskAssigneeIDs,
		condExpr, input.ConditionTrueStepID, input.ConditionFalseStepID,
		autoAct, input.TimerHours, input.TimerBusinessHoursOnly,
		input.SLAHours, input.EscalationUserIDs, input.IsOptional, input.CanDelegate,
	).Scan(
		&s.ID, &s.WorkflowDefinitionID, &s.OrganizationID, &s.StepOrder, &s.Name, &s.Description,
		&s.StepType, &s.ApproverType, &s.ApproverIDs, &s.ApprovalMode,
		&s.MinimumApprovals, &s.TaskDescription, &s.TaskAssigneeType, &s.TaskAssigneeIDs,
		&s.ConditionExpression, &s.ConditionTrueStepID, &s.ConditionFalseStepID,
		&s.AutoAction, &s.TimerHours, &s.TimerBusinessHoursOnly,
		&s.SLAHours, &s.EscalationUserIDs, &s.IsOptional, &s.CanDelegate,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update workflow step: %w", err)
	}

	return &s, nil
}

// DeleteStep removes a step from a workflow definition.
func (e *WorkflowEngine) DeleteStep(ctx context.Context, orgID, defID, stepID uuid.UUID) error {
	tag, err := e.pool.Exec(ctx, `
		DELETE FROM workflow_steps
		WHERE id = $1 AND workflow_definition_id = $2
			AND (organization_id = $3 OR organization_id IS NULL)`,
		stepID, defID, orgID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete workflow step: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("workflow step not found")
	}
	return nil
}

// ============================================================
// WORKFLOW LIFECYCLE
// ============================================================

// StartWorkflow creates a new workflow instance for a given entity, resolves the first step,
// and creates the initial step execution record.
func (e *WorkflowEngine) StartWorkflow(
	ctx context.Context,
	orgID uuid.UUID,
	workflowType, entityType string,
	entityID uuid.UUID,
	entityRef string,
	startedBy uuid.UUID,
) (*WorkflowInstance, error) {
	// Find the active definition for this workflow type and entity type
	var defID uuid.UUID
	var slaCfgJSON json.RawMessage
	err := e.pool.QueryRow(ctx, `
		SELECT id, sla_config
		FROM workflow_definitions
		WHERE workflow_type = $1::workflow_type
			AND entity_type = $2
			AND status = 'active'
			AND (organization_id = $3 OR organization_id IS NULL)
		ORDER BY organization_id DESC NULLS LAST
		LIMIT 1`,
		workflowType, entityType, orgID,
	).Scan(&defID, &slaCfgJSON)
	if err != nil {
		return nil, fmt.Errorf("no active workflow definition found for type %s/%s: %w", workflowType, entityType, err)
	}

	// Parse SLA config
	var slaCfg SLAConfig
	if len(slaCfgJSON) > 0 {
		_ = json.Unmarshal(slaCfgJSON, &slaCfg)
	}

	// Get the first step
	firstStep, err := e.getFirstStep(ctx, defID)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve first step: %w", err)
	}

	now := time.Now()
	var slaDeadline *time.Time
	if slaCfg.TotalSLAHours > 0 {
		d := now.Add(time.Duration(slaCfg.TotalSLAHours) * time.Hour)
		slaDeadline = &d
	}

	// Begin transaction
	tx, err := e.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Create workflow instance
	var inst WorkflowInstance
	err = tx.QueryRow(ctx, `
		INSERT INTO workflow_instances
			(organization_id, workflow_definition_id, entity_type, entity_id, entity_ref,
			 status, current_step_id, current_step_order, started_at, started_by,
			 sla_status, sla_deadline, metadata)
		VALUES ($1, $2, $3, $4, $5, 'active', $6, $7, $8, $9, 'on_track', $10, '{}')
		RETURNING id, organization_id, workflow_definition_id, entity_type, entity_id,
				  entity_ref, status::text, current_step_id, current_step_order,
				  started_at, started_by, completed_at, completion_outcome::text,
				  total_duration_hours, sla_status::text, sla_deadline,
				  metadata, created_at, updated_at`,
		orgID, defID, entityType, entityID, entityRef,
		firstStep.ID, firstStep.StepOrder, now, startedBy, slaDeadline,
	).Scan(
		&inst.ID, &inst.OrganizationID, &inst.WorkflowDefinitionID,
		&inst.EntityType, &inst.EntityID, &inst.EntityRef,
		&inst.Status, &inst.CurrentStepID, &inst.CurrentStepOrder,
		&inst.StartedAt, &inst.StartedBy, &inst.CompletedAt, &inst.CompletionOutcome,
		&inst.TotalDurationHours, &inst.SLAStatus, &inst.SLADeadline,
		&inst.Metadata, &inst.CreatedAt, &inst.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create workflow instance: %w", err)
	}

	// Create the first step execution
	err = e.createStepExecution(ctx, tx, orgID, inst.ID, firstStep, slaCfg.StepDefaultSLAHours)
	if err != nil {
		return nil, fmt.Errorf("failed to create initial step execution: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Info().
		Str("instance_id", inst.ID.String()).
		Str("workflow_type", workflowType).
		Str("entity", entityRef).
		Msg("Workflow started")

	return &inst, nil
}

// ProcessStep handles a user action on a step execution (approve/reject/complete).
// It validates the actor, records the action, and advances the workflow to the next step.
func (e *WorkflowEngine) ProcessStep(
	ctx context.Context,
	orgID uuid.UUID,
	executionID uuid.UUID,
	action string,
	actorID uuid.UUID,
	comments, reason string,
) (*StepExecution, error) {
	tx, err := e.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Lock and fetch the execution
	var exec StepExecution
	var instanceID uuid.UUID
	var stepType string
	var approvalMode string
	var minimumApprovals int
	err = tx.QueryRow(ctx, `
		SELECT
			e.id, e.organization_id, e.workflow_instance_id, e.workflow_step_id,
			e.step_order, e.status::text, e.assigned_to, e.delegated_to,
			e.sla_deadline, e.sla_status::text, e.started_at, e.created_at,
			s.step_type::text, s.approval_mode::text, s.minimum_approvals
		FROM workflow_step_executions e
		JOIN workflow_steps s ON s.id = e.workflow_step_id
		WHERE e.id = $1 AND e.organization_id = $2
		FOR UPDATE OF e`,
		executionID, orgID,
	).Scan(
		&exec.ID, &exec.OrganizationID, &exec.WorkflowInstanceID, &exec.WorkflowStepID,
		&exec.StepOrder, &exec.Status, &exec.AssignedTo, &exec.DelegatedTo,
		&exec.SLADeadline, &exec.SLAStatus, &exec.StartedAt, &exec.CreatedAt,
		&stepType, &approvalMode, &minimumApprovals,
	)
	if err != nil {
		return nil, fmt.Errorf("step execution not found: %w", err)
	}
	instanceID = exec.WorkflowInstanceID

	// Validate that the execution is actionable
	if exec.Status != "pending" && exec.Status != "in_progress" {
		return nil, fmt.Errorf("step execution is not in an actionable state: %s", exec.Status)
	}

	// Validate action is valid for this step type
	if !isValidAction(stepType, action) {
		return nil, fmt.Errorf("action '%s' is not valid for step type '%s'", action, stepType)
	}

	// Resolve the execution status based on the action
	execStatus := resolveExecutionStatus(action)
	now := time.Now()
	var durationHours *float64
	if exec.StartedAt != nil {
		d := now.Sub(*exec.StartedAt).Hours()
		durationHours = &d
	}

	// Update the execution
	_, err = tx.Exec(ctx, `
		UPDATE workflow_step_executions SET
			status = $3::workflow_step_exec_status,
			action_taken_by = $4,
			action_taken_at = $5,
			action = $6::workflow_step_action,
			comments = $7,
			decision_reason = $8,
			completed_at = $5,
			duration_hours = $9
		WHERE id = $1 AND organization_id = $2`,
		executionID, orgID, execStatus, actorID, now, action, comments, reason, durationHours,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update step execution: %w", err)
	}

	// For parallel_gate steps, check if all required approvals are met
	if stepType == "parallel_gate" || approvalMode == "all_required" || approvalMode == "majority" {
		allMet, err := e.checkParallelCompletion(ctx, tx, instanceID, exec.WorkflowStepID, approvalMode, minimumApprovals)
		if err != nil {
			return nil, fmt.Errorf("failed to check parallel completion: %w", err)
		}
		if !allMet {
			// Not all required approvals met yet; commit and return
			if err := tx.Commit(ctx); err != nil {
				return nil, fmt.Errorf("failed to commit: %w", err)
			}
			exec.Status = execStatus
			exec.Action = &action
			exec.ActionTakenBy = &actorID
			exec.ActionTakenAt = &now
			exec.Comments = comments
			exec.DecisionReason = reason
			return &exec, nil
		}
	}

	// Handle rejection: complete the workflow as rejected
	if action == "reject" {
		err = e.completeWorkflow(ctx, tx, instanceID, orgID, "rejected")
		if err != nil {
			return nil, fmt.Errorf("failed to complete workflow on rejection: %w", err)
		}
	} else {
		// Advance to the next step
		err = e.advanceWorkflow(ctx, tx, orgID, instanceID, exec.WorkflowStepID, exec.StepOrder)
		if err != nil {
			return nil, fmt.Errorf("failed to advance workflow: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	exec.Status = execStatus
	exec.Action = &action
	exec.ActionTakenBy = &actorID
	exec.ActionTakenAt = &now
	exec.Comments = comments
	exec.DecisionReason = reason
	exec.CompletedAt = &now
	exec.DurationHours = durationHours

	log.Info().
		Str("execution_id", executionID.String()).
		Str("action", action).
		Str("actor", actorID.String()).
		Msg("Workflow step processed")

	return &exec, nil
}

// DelegateStep delegates a step execution to another user.
func (e *WorkflowEngine) DelegateStep(ctx context.Context, orgID, executionID, delegatorID, delegateID uuid.UUID) error {
	now := time.Now()

	tag, err := e.pool.Exec(ctx, `
		UPDATE workflow_step_executions SET
			status = 'delegated'::workflow_step_exec_status,
			delegated_to = $3,
			delegated_by = $4,
			delegated_at = $5,
			assigned_to = $3
		WHERE id = $1 AND organization_id = $2
			AND status IN ('pending', 'in_progress')`,
		executionID, orgID, delegateID, delegatorID, now,
	)
	if err != nil {
		return fmt.Errorf("failed to delegate step: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("step execution not found or not in a delegatable state")
	}

	// Reset status to pending for the new assignee
	_, err = e.pool.Exec(ctx, `
		UPDATE workflow_step_executions SET status = 'pending'::workflow_step_exec_status
		WHERE id = $1 AND organization_id = $2`,
		executionID, orgID,
	)
	if err != nil {
		return fmt.Errorf("failed to reset delegated step status: %w", err)
	}

	log.Info().
		Str("execution_id", executionID.String()).
		Str("delegator", delegatorID.String()).
		Str("delegate", delegateID.String()).
		Msg("Workflow step delegated")

	return nil
}

// EscalateStep escalates an overdue step execution to configured escalation users.
func (e *WorkflowEngine) EscalateStep(ctx context.Context, orgID, executionID uuid.UUID) error {
	now := time.Now()

	// Fetch escalation user IDs from the step definition
	var escalationUserIDs []uuid.UUID
	err := e.pool.QueryRow(ctx, `
		SELECT s.escalation_user_ids
		FROM workflow_step_executions e
		JOIN workflow_steps s ON s.id = e.workflow_step_id
		WHERE e.id = $1 AND e.organization_id = $2`,
		executionID, orgID,
	).Scan(&escalationUserIDs)
	if err != nil {
		return fmt.Errorf("failed to fetch escalation config: %w", err)
	}

	var escalatedTo *uuid.UUID
	if len(escalationUserIDs) > 0 {
		escalatedTo = &escalationUserIDs[0]
	}

	_, err = e.pool.Exec(ctx, `
		UPDATE workflow_step_executions SET
			status = 'escalated'::workflow_step_exec_status,
			escalated_at = $3,
			escalated_to = $4,
			sla_status = 'breached'::workflow_sla_status
		WHERE id = $1 AND organization_id = $2
			AND status IN ('pending', 'in_progress')`,
		executionID, orgID, now, escalatedTo,
	)
	if err != nil {
		return fmt.Errorf("failed to escalate step: %w", err)
	}

	// If escalated to someone, also update assigned_to and reset to in_progress
	if escalatedTo != nil {
		_, _ = e.pool.Exec(ctx, `
			UPDATE workflow_step_executions SET
				assigned_to = $3,
				status = 'in_progress'::workflow_step_exec_status
			WHERE id = $1 AND organization_id = $2`,
			executionID, orgID, *escalatedTo,
		)
	}

	log.Warn().
		Str("execution_id", executionID.String()).
		Msg("Workflow step escalated")

	return nil
}

// CancelWorkflow cancels an active workflow instance and all its pending step executions.
func (e *WorkflowEngine) CancelWorkflow(ctx context.Context, orgID, instanceID, cancelledBy uuid.UUID, reason string) error {
	tx, err := e.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Cancel all pending/in_progress step executions
	_, err = tx.Exec(ctx, `
		UPDATE workflow_step_executions SET
			status = 'skipped'::workflow_step_exec_status,
			comments = $3,
			completed_at = NOW()
		WHERE workflow_instance_id = $1 AND organization_id = $2
			AND status IN ('pending', 'in_progress')`,
		instanceID, orgID, "Workflow cancelled: "+reason,
	)
	if err != nil {
		return fmt.Errorf("failed to cancel step executions: %w", err)
	}

	err = e.completeWorkflow(ctx, tx, instanceID, orgID, "cancelled")
	if err != nil {
		return fmt.Errorf("failed to cancel workflow: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	log.Info().
		Str("instance_id", instanceID.String()).
		Str("cancelled_by", cancelledBy.String()).
		Str("reason", reason).
		Msg("Workflow cancelled")

	return nil
}

// ============================================================
// QUERIES
// ============================================================

// GetPendingApprovals returns all step executions pending action by a specific user.
func (e *WorkflowEngine) GetPendingApprovals(ctx context.Context, orgID, userID uuid.UUID) ([]StepExecution, error) {
	rows, err := e.pool.Query(ctx, `
		SELECT
			e.id, e.organization_id, e.workflow_instance_id, e.workflow_step_id,
			e.step_order, e.status::text, e.assigned_to, e.delegated_to,
			e.delegated_by, e.delegated_at, e.action_taken_by, e.action_taken_at,
			e.action::text, e.comments, e.decision_reason, e.attachments_paths,
			e.sla_deadline, e.sla_status::text, e.escalated_at, e.escalated_to,
			e.started_at, e.completed_at, e.duration_hours, e.created_at,
			s.name, s.step_type::text,
			i.entity_ref, i.entity_type, i.status::text
		FROM workflow_step_executions e
		JOIN workflow_steps s ON s.id = e.workflow_step_id
		JOIN workflow_instances i ON i.id = e.workflow_instance_id
		WHERE e.organization_id = $1
			AND (e.assigned_to = $2 OR e.delegated_to = $2)
			AND e.status IN ('pending', 'in_progress')
			AND i.status = 'active'
		ORDER BY e.sla_deadline ASC NULLS LAST, e.created_at ASC`,
		orgID, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending approvals: %w", err)
	}
	defer rows.Close()

	var executions []StepExecution
	for rows.Next() {
		var exec StepExecution
		if err := rows.Scan(
			&exec.ID, &exec.OrganizationID, &exec.WorkflowInstanceID, &exec.WorkflowStepID,
			&exec.StepOrder, &exec.Status, &exec.AssignedTo, &exec.DelegatedTo,
			&exec.DelegatedBy, &exec.DelegatedAt, &exec.ActionTakenBy, &exec.ActionTakenAt,
			&exec.Action, &exec.Comments, &exec.DecisionReason, &exec.AttachmentsPaths,
			&exec.SLADeadline, &exec.SLAStatus, &exec.EscalatedAt, &exec.EscalatedTo,
			&exec.StartedAt, &exec.CompletedAt, &exec.DurationHours, &exec.CreatedAt,
			&exec.StepName, &exec.StepType,
			&exec.EntityRef, &exec.EntityType, &exec.InstanceStatus,
		); err != nil {
			return nil, fmt.Errorf("failed to scan pending approval: %w", err)
		}
		executions = append(executions, exec)
	}

	return executions, nil
}

// GetWorkflowHistory returns all workflow instances and their executions for a given entity.
func (e *WorkflowEngine) GetWorkflowHistory(ctx context.Context, orgID uuid.UUID, entityType string, entityID uuid.UUID) ([]WorkflowInstance, error) {
	rows, err := e.pool.Query(ctx, `
		SELECT
			i.id, i.organization_id, i.workflow_definition_id, i.entity_type, i.entity_id,
			i.entity_ref, i.status::text, i.current_step_id, i.current_step_order,
			i.started_at, i.started_by, i.completed_at, i.completion_outcome::text,
			i.total_duration_hours, i.sla_status::text, i.sla_deadline,
			i.metadata, i.created_at, i.updated_at,
			d.name
		FROM workflow_instances i
		JOIN workflow_definitions d ON d.id = i.workflow_definition_id
		WHERE i.organization_id = $1 AND i.entity_type = $2 AND i.entity_id = $3
		ORDER BY i.started_at DESC`,
		orgID, entityType, entityID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query workflow history: %w", err)
	}
	defer rows.Close()

	var instances []WorkflowInstance
	for rows.Next() {
		var inst WorkflowInstance
		if err := rows.Scan(
			&inst.ID, &inst.OrganizationID, &inst.WorkflowDefinitionID,
			&inst.EntityType, &inst.EntityID, &inst.EntityRef,
			&inst.Status, &inst.CurrentStepID, &inst.CurrentStepOrder,
			&inst.StartedAt, &inst.StartedBy, &inst.CompletedAt, &inst.CompletionOutcome,
			&inst.TotalDurationHours, &inst.SLAStatus, &inst.SLADeadline,
			&inst.Metadata, &inst.CreatedAt, &inst.UpdatedAt,
			&inst.DefinitionName,
		); err != nil {
			return nil, fmt.Errorf("failed to scan workflow instance: %w", err)
		}
		instances = append(instances, inst)
	}

	// Load executions for each instance
	for i := range instances {
		execs, err := e.getInstanceExecutions(ctx, orgID, instances[i].ID)
		if err != nil {
			log.Error().Err(err).Str("instance_id", instances[i].ID.String()).Msg("Failed to load instance executions")
			continue
		}
		instances[i].Executions = execs
	}

	return instances, nil
}

// ListInstances returns paginated workflow instances with optional filters.
func (e *WorkflowEngine) ListInstances(ctx context.Context, orgID uuid.UUID, filters InstanceFilters, limit, offset int) ([]WorkflowInstance, int64, error) {
	// Build WHERE clause
	where := "i.organization_id = $1"
	args := []interface{}{orgID}
	argIdx := 2

	if filters.Status != "" {
		where += fmt.Sprintf(" AND i.status = $%d::workflow_instance_status", argIdx)
		args = append(args, filters.Status)
		argIdx++
	}
	if filters.EntityType != "" {
		where += fmt.Sprintf(" AND i.entity_type = $%d", argIdx)
		args = append(args, filters.EntityType)
		argIdx++
	}
	if filters.SLAStatus != "" {
		where += fmt.Sprintf(" AND i.sla_status = $%d::workflow_sla_status", argIdx)
		args = append(args, filters.SLAStatus)
		argIdx++
	}

	// Count
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM workflow_instances i WHERE %s", where)
	if err := e.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count workflow instances: %w", err)
	}

	// Query
	listArgs := append(args, limit, offset)
	query := fmt.Sprintf(`
		SELECT
			i.id, i.organization_id, i.workflow_definition_id, i.entity_type, i.entity_id,
			i.entity_ref, i.status::text, i.current_step_id, i.current_step_order,
			i.started_at, i.started_by, i.completed_at, i.completion_outcome::text,
			i.total_duration_hours, i.sla_status::text, i.sla_deadline,
			i.metadata, i.created_at, i.updated_at,
			d.name,
			COALESCE(cs.name, '')
		FROM workflow_instances i
		JOIN workflow_definitions d ON d.id = i.workflow_definition_id
		LEFT JOIN workflow_steps cs ON cs.id = i.current_step_id
		WHERE %s
		ORDER BY i.created_at DESC
		LIMIT $%d OFFSET $%d`,
		where, argIdx, argIdx+1,
	)

	rows, err := e.pool.Query(ctx, query, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list workflow instances: %w", err)
	}
	defer rows.Close()

	var instances []WorkflowInstance
	for rows.Next() {
		var inst WorkflowInstance
		if err := rows.Scan(
			&inst.ID, &inst.OrganizationID, &inst.WorkflowDefinitionID,
			&inst.EntityType, &inst.EntityID, &inst.EntityRef,
			&inst.Status, &inst.CurrentStepID, &inst.CurrentStepOrder,
			&inst.StartedAt, &inst.StartedBy, &inst.CompletedAt, &inst.CompletionOutcome,
			&inst.TotalDurationHours, &inst.SLAStatus, &inst.SLADeadline,
			&inst.Metadata, &inst.CreatedAt, &inst.UpdatedAt,
			&inst.DefinitionName,
			&inst.CurrentStepName,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan workflow instance: %w", err)
		}
		instances = append(instances, inst)
	}

	return instances, total, nil
}

// GetInstance returns a single workflow instance by ID with its executions.
func (e *WorkflowEngine) GetInstance(ctx context.Context, orgID, instanceID uuid.UUID) (*WorkflowInstance, error) {
	var inst WorkflowInstance
	err := e.pool.QueryRow(ctx, `
		SELECT
			i.id, i.organization_id, i.workflow_definition_id, i.entity_type, i.entity_id,
			i.entity_ref, i.status::text, i.current_step_id, i.current_step_order,
			i.started_at, i.started_by, i.completed_at, i.completion_outcome::text,
			i.total_duration_hours, i.sla_status::text, i.sla_deadline,
			i.metadata, i.created_at, i.updated_at,
			d.name,
			COALESCE(cs.name, '')
		FROM workflow_instances i
		JOIN workflow_definitions d ON d.id = i.workflow_definition_id
		LEFT JOIN workflow_steps cs ON cs.id = i.current_step_id
		WHERE i.id = $1 AND i.organization_id = $2`,
		instanceID, orgID,
	).Scan(
		&inst.ID, &inst.OrganizationID, &inst.WorkflowDefinitionID,
		&inst.EntityType, &inst.EntityID, &inst.EntityRef,
		&inst.Status, &inst.CurrentStepID, &inst.CurrentStepOrder,
		&inst.StartedAt, &inst.StartedBy, &inst.CompletedAt, &inst.CompletionOutcome,
		&inst.TotalDurationHours, &inst.SLAStatus, &inst.SLADeadline,
		&inst.Metadata, &inst.CreatedAt, &inst.UpdatedAt,
		&inst.DefinitionName,
		&inst.CurrentStepName,
	)
	if err != nil {
		return nil, fmt.Errorf("workflow instance not found: %w", err)
	}

	execs, err := e.getInstanceExecutions(ctx, orgID, instanceID)
	if err == nil {
		inst.Executions = execs
	}

	return &inst, nil
}

// ============================================================
// DELEGATION RULES
// ============================================================

// ListDelegations returns all delegation rules for an organization.
func (e *WorkflowEngine) ListDelegations(ctx context.Context, orgID uuid.UUID) ([]DelegationRule, error) {
	rows, err := e.pool.Query(ctx, `
		SELECT id, organization_id, delegator_user_id, delegate_user_id,
			   workflow_types, valid_from, valid_until, reason,
			   is_active, created_by, created_at, updated_at
		FROM workflow_delegation_rules
		WHERE organization_id = $1
		ORDER BY created_at DESC`,
		orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list delegation rules: %w", err)
	}
	defer rows.Close()

	var rules []DelegationRule
	for rows.Next() {
		var r DelegationRule
		if err := rows.Scan(
			&r.ID, &r.OrganizationID, &r.DelegatorUserID, &r.DelegateUserID,
			&r.WorkflowTypes, &r.ValidFrom, &r.ValidUntil, &r.Reason,
			&r.IsActive, &r.CreatedBy, &r.CreatedAt, &r.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan delegation rule: %w", err)
		}
		rules = append(rules, r)
	}

	return rules, nil
}

// CreateDelegation creates a new delegation rule.
func (e *WorkflowEngine) CreateDelegation(ctx context.Context, orgID uuid.UUID, input CreateDelegationInput) (*DelegationRule, error) {
	if input.DelegatorUserID == uuid.Nil || input.DelegateUserID == uuid.Nil {
		return nil, fmt.Errorf("delegator and delegate user IDs are required")
	}
	if input.DelegatorUserID == input.DelegateUserID {
		return nil, fmt.Errorf("cannot delegate to yourself")
	}

	var r DelegationRule
	err := e.pool.QueryRow(ctx, `
		INSERT INTO workflow_delegation_rules
			(organization_id, delegator_user_id, delegate_user_id, workflow_types,
			 valid_from, valid_until, reason, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, true)
		RETURNING id, organization_id, delegator_user_id, delegate_user_id,
				  workflow_types, valid_from, valid_until, reason,
				  is_active, created_by, created_at, updated_at`,
		orgID, input.DelegatorUserID, input.DelegateUserID, input.WorkflowTypes,
		input.ValidFrom, input.ValidUntil, input.Reason,
	).Scan(
		&r.ID, &r.OrganizationID, &r.DelegatorUserID, &r.DelegateUserID,
		&r.WorkflowTypes, &r.ValidFrom, &r.ValidUntil, &r.Reason,
		&r.IsActive, &r.CreatedBy, &r.CreatedAt, &r.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create delegation rule: %w", err)
	}

	log.Info().
		Str("delegator", input.DelegatorUserID.String()).
		Str("delegate", input.DelegateUserID.String()).
		Msg("Delegation rule created")

	return &r, nil
}

// ============================================================
// SLA MANAGEMENT (exposed for scheduler)
// ============================================================

// CheckSLABreaches checks all active step executions for SLA breaches and returns execution IDs that are breached.
func (e *WorkflowEngine) CheckSLABreaches(ctx context.Context) ([]uuid.UUID, error) {
	now := time.Now()

	rows, err := e.pool.Query(ctx, `
		SELECT e.id, e.organization_id
		FROM workflow_step_executions e
		JOIN workflow_instances i ON i.id = e.workflow_instance_id
		WHERE e.status IN ('pending', 'in_progress')
			AND e.sla_deadline IS NOT NULL
			AND e.sla_deadline < $1
			AND e.sla_status != 'breached'
			AND i.status = 'active'`,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check SLA breaches: %w", err)
	}
	defer rows.Close()

	var breachedIDs []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		var orgID uuid.UUID
		if err := rows.Scan(&id, &orgID); err != nil {
			continue
		}

		// Update SLA status
		_, _ = e.pool.Exec(ctx, `
			UPDATE workflow_step_executions SET sla_status = 'breached'::workflow_sla_status
			WHERE id = $1 AND organization_id = $2`,
			id, orgID,
		)

		breachedIDs = append(breachedIDs, id)
	}

	// Also update at-risk: within 20% of deadline
	_, _ = e.pool.Exec(ctx, `
		UPDATE workflow_step_executions SET sla_status = 'at_risk'::workflow_sla_status
		WHERE status IN ('pending', 'in_progress')
			AND sla_deadline IS NOT NULL
			AND sla_status = 'on_track'
			AND sla_deadline < $1 + INTERVAL '1 hour' * (
				EXTRACT(EPOCH FROM (sla_deadline - created_at)) / 3600 * 0.2
			)`,
		now,
	)

	// Update instance SLA status if any step is breached
	_, _ = e.pool.Exec(ctx, `
		UPDATE workflow_instances SET sla_status = 'breached'::workflow_sla_status
		WHERE status = 'active'
			AND sla_deadline IS NOT NULL
			AND sla_deadline < $1
			AND sla_status != 'breached'`,
		now,
	)

	return breachedIDs, nil
}

// GetOverdueTimerSteps returns step executions of type 'timer' whose timer has expired.
func (e *WorkflowEngine) GetOverdueTimerSteps(ctx context.Context) ([]StepExecution, error) {
	rows, err := e.pool.Query(ctx, `
		SELECT e.id, e.organization_id, e.workflow_instance_id, e.workflow_step_id,
			   e.step_order, e.status::text, e.created_at
		FROM workflow_step_executions e
		JOIN workflow_steps s ON s.id = e.workflow_step_id
		JOIN workflow_instances i ON i.id = e.workflow_instance_id
		WHERE s.step_type = 'timer'
			AND e.status IN ('pending', 'in_progress')
			AND i.status = 'active'
			AND e.created_at + INTERVAL '1 hour' * COALESCE(s.timer_hours, 0) <= NOW()`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query overdue timer steps: %w", err)
	}
	defer rows.Close()

	var execs []StepExecution
	for rows.Next() {
		var exec StepExecution
		if err := rows.Scan(
			&exec.ID, &exec.OrganizationID, &exec.WorkflowInstanceID,
			&exec.WorkflowStepID, &exec.StepOrder, &exec.Status, &exec.CreatedAt,
		); err != nil {
			continue
		}
		execs = append(execs, exec)
	}

	return execs, nil
}

// ProcessTimerStep completes a timer step and advances the workflow.
func (e *WorkflowEngine) ProcessTimerStep(ctx context.Context, exec StepExecution) error {
	tx, err := e.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	now := time.Now()
	_, err = tx.Exec(ctx, `
		UPDATE workflow_step_executions SET
			status = 'completed'::workflow_step_exec_status,
			action = 'auto_complete'::workflow_step_action,
			completed_at = $3,
			comments = 'Timer expired - auto-advanced'
		WHERE id = $1 AND organization_id = $2`,
		exec.ID, exec.OrganizationID, now,
	)
	if err != nil {
		return fmt.Errorf("failed to complete timer step: %w", err)
	}

	err = e.advanceWorkflow(ctx, tx, exec.OrganizationID, exec.WorkflowInstanceID, exec.WorkflowStepID, exec.StepOrder)
	if err != nil {
		return fmt.Errorf("failed to advance workflow after timer: %w", err)
	}

	return tx.Commit(ctx)
}

// ============================================================
// INTERNAL HELPERS
// ============================================================

// getFirstStep returns the step with the lowest step_order in a workflow definition.
func (e *WorkflowEngine) getFirstStep(ctx context.Context, defID uuid.UUID) (*WorkflowStep, error) {
	rows, err := e.pool.Query(ctx, `
		SELECT id, workflow_definition_id, organization_id, step_order, name, description,
			   step_type::text, approver_type::text, approver_ids, approval_mode::text,
			   minimum_approvals, task_description, task_assignee_type::text, task_assignee_ids,
			   condition_expression, condition_true_step_id, condition_false_step_id,
			   auto_action, timer_hours, timer_business_hours_only,
			   sla_hours, escalation_user_ids, is_optional, can_delegate,
			   created_at, updated_at
		FROM workflow_steps
		WHERE workflow_definition_id = $1
		ORDER BY step_order ASC
		LIMIT 1`,
		defID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query first step: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("workflow definition has no steps")
	}

	s, err := scanWorkflowStep(rows)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// getNextStep returns the step with the next higher step_order after currentOrder,
// or nil if the current step is the last.
func (e *WorkflowEngine) getNextStep(ctx context.Context, defID uuid.UUID, currentOrder int) (*WorkflowStep, error) {
	rows, err := e.pool.Query(ctx, `
		SELECT id, workflow_definition_id, organization_id, step_order, name, description,
			   step_type::text, approver_type::text, approver_ids, approval_mode::text,
			   minimum_approvals, task_description, task_assignee_type::text, task_assignee_ids,
			   condition_expression, condition_true_step_id, condition_false_step_id,
			   auto_action, timer_hours, timer_business_hours_only,
			   sla_hours, escalation_user_ids, is_optional, can_delegate,
			   created_at, updated_at
		FROM workflow_steps
		WHERE workflow_definition_id = $1 AND step_order > $2
		ORDER BY step_order ASC
		LIMIT 1`,
		defID, currentOrder,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query next step: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil // No next step; workflow is complete
	}

	s, err := scanWorkflowStep(rows)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// getStepByID returns a workflow step by its primary key.
func (e *WorkflowEngine) getStepByID(ctx context.Context, stepID uuid.UUID) (*WorkflowStep, error) {
	rows, err := e.pool.Query(ctx, `
		SELECT id, workflow_definition_id, organization_id, step_order, name, description,
			   step_type::text, approver_type::text, approver_ids, approval_mode::text,
			   minimum_approvals, task_description, task_assignee_type::text, task_assignee_ids,
			   condition_expression, condition_true_step_id, condition_false_step_id,
			   auto_action, timer_hours, timer_business_hours_only,
			   sla_hours, escalation_user_ids, is_optional, can_delegate,
			   created_at, updated_at
		FROM workflow_steps
		WHERE id = $1`,
		stepID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query step by ID: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("step not found: %s", stepID)
	}

	s, err := scanWorkflowStep(rows)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// createStepExecution creates a new step execution record for a given step.
func (e *WorkflowEngine) createStepExecution(ctx context.Context, tx pgx.Tx, orgID uuid.UUID, instanceID uuid.UUID, step *WorkflowStep, defaultSLAHours int) error {
	now := time.Now()

	// Calculate SLA deadline
	var slaDeadline *time.Time
	slaHours := defaultSLAHours
	if step.SLAHours != nil && *step.SLAHours > 0 {
		slaHours = *step.SLAHours
	}
	if slaHours > 0 {
		d := now.Add(time.Duration(slaHours) * time.Hour)
		slaDeadline = &d
	}

	// Determine initial status based on step type
	initialStatus := "pending"
	if step.StepType == "auto_action" || step.StepType == "notification" || step.StepType == "condition" {
		initialStatus = "in_progress"
	}

	_, err := tx.Exec(ctx, `
		INSERT INTO workflow_step_executions
			(organization_id, workflow_instance_id, workflow_step_id, step_order,
			 status, sla_deadline, sla_status, started_at)
		VALUES ($1, $2, $3, $4, $5::workflow_step_exec_status, $6, 'on_track', $7)`,
		orgID, instanceID, step.ID, step.StepOrder,
		initialStatus, slaDeadline, now,
	)
	if err != nil {
		return fmt.Errorf("failed to insert step execution: %w", err)
	}

	// For auto-advancing step types, immediately process them
	if step.StepType == "auto_action" || step.StepType == "notification" {
		err = e.autoAdvanceStep(ctx, tx, orgID, instanceID, step)
		if err != nil {
			return fmt.Errorf("failed to auto-advance step: %w", err)
		}
	} else if step.StepType == "condition" {
		err = e.processConditionStep(ctx, tx, orgID, instanceID, step)
		if err != nil {
			return fmt.Errorf("failed to process condition step: %w", err)
		}
	}

	return nil
}

// autoAdvanceStep completes auto_action and notification steps and moves to the next step.
func (e *WorkflowEngine) autoAdvanceStep(ctx context.Context, tx pgx.Tx, orgID, instanceID uuid.UUID, step *WorkflowStep) error {
	now := time.Now()

	// Mark the execution as completed
	_, err := tx.Exec(ctx, `
		UPDATE workflow_step_executions SET
			status = 'completed'::workflow_step_exec_status,
			action = 'auto_complete'::workflow_step_action,
			completed_at = $3,
			comments = 'Auto-completed'
		WHERE workflow_instance_id = $1 AND workflow_step_id = $2
			AND organization_id = $4`,
		instanceID, step.ID, now, orgID,
	)
	if err != nil {
		return fmt.Errorf("failed to auto-complete step execution: %w", err)
	}

	// Get the workflow definition ID
	var defID uuid.UUID
	err = tx.QueryRow(ctx,
		`SELECT workflow_definition_id FROM workflow_instances WHERE id = $1`,
		instanceID,
	).Scan(&defID)
	if err != nil {
		return fmt.Errorf("failed to get definition ID: %w", err)
	}

	// Get SLA config
	var slaCfgJSON json.RawMessage
	_ = tx.QueryRow(ctx,
		`SELECT sla_config FROM workflow_definitions WHERE id = $1`,
		defID,
	).Scan(&slaCfgJSON)
	var slaCfg SLAConfig
	if len(slaCfgJSON) > 0 {
		_ = json.Unmarshal(slaCfgJSON, &slaCfg)
	}

	// Find the next step
	nextStep, err := e.getNextStep(ctx, defID, step.StepOrder)
	if err != nil {
		return fmt.Errorf("failed to find next step: %w", err)
	}

	if nextStep == nil {
		// Workflow is complete
		return e.completeWorkflow(ctx, tx, instanceID, orgID, "completed")
	}

	// Update instance to point to next step
	_, err = tx.Exec(ctx, `
		UPDATE workflow_instances SET current_step_id = $2, current_step_order = $3
		WHERE id = $1`,
		instanceID, nextStep.ID, nextStep.StepOrder,
	)
	if err != nil {
		return fmt.Errorf("failed to update instance current step: %w", err)
	}

	// Create execution for the next step
	return e.createStepExecution(ctx, tx, orgID, instanceID, nextStep, slaCfg.StepDefaultSLAHours)
}

// processConditionStep evaluates a condition expression and routes to the appropriate branch.
func (e *WorkflowEngine) processConditionStep(ctx context.Context, tx pgx.Tx, orgID, instanceID uuid.UUID, step *WorkflowStep) error {
	now := time.Now()

	// Mark condition step as completed
	_, err := tx.Exec(ctx, `
		UPDATE workflow_step_executions SET
			status = 'completed'::workflow_step_exec_status,
			action = 'auto_complete'::workflow_step_action,
			completed_at = $3,
			comments = 'Condition evaluated'
		WHERE workflow_instance_id = $1 AND workflow_step_id = $2
			AND organization_id = $4`,
		instanceID, step.ID, now, orgID,
	)
	if err != nil {
		return fmt.Errorf("failed to complete condition step: %w", err)
	}

	// Parse the condition expression
	var condExpr ConditionExpression
	if len(step.ConditionExpression) > 0 {
		_ = json.Unmarshal(step.ConditionExpression, &condExpr)
	}

	// Evaluate the condition against the instance's entity metadata
	var instanceMetadata json.RawMessage
	err = tx.QueryRow(ctx,
		`SELECT metadata FROM workflow_instances WHERE id = $1`,
		instanceID,
	).Scan(&instanceMetadata)
	if err != nil {
		return fmt.Errorf("failed to get instance metadata: %w", err)
	}

	result := evaluateCondition(condExpr, instanceMetadata)

	// Route to the appropriate step
	var targetStepID *uuid.UUID
	if result {
		targetStepID = step.ConditionTrueStepID
	} else {
		targetStepID = step.ConditionFalseStepID
	}

	var targetStep *WorkflowStep
	if targetStepID != nil {
		targetStep, err = e.getStepByID(ctx, *targetStepID)
		if err != nil {
			return fmt.Errorf("failed to get target step: %w", err)
		}
	}

	if targetStep == nil {
		// No target; try next sequential step
		var defID uuid.UUID
		_ = tx.QueryRow(ctx,
			`SELECT workflow_definition_id FROM workflow_instances WHERE id = $1`,
			instanceID,
		).Scan(&defID)

		targetStep, err = e.getNextStep(ctx, defID, step.StepOrder)
		if err != nil || targetStep == nil {
			return e.completeWorkflow(ctx, tx, instanceID, orgID, "completed")
		}
	}

	// Get SLA config
	var defID uuid.UUID
	_ = tx.QueryRow(ctx,
		`SELECT workflow_definition_id FROM workflow_instances WHERE id = $1`,
		instanceID,
	).Scan(&defID)
	var slaCfgJSON json.RawMessage
	_ = tx.QueryRow(ctx,
		`SELECT sla_config FROM workflow_definitions WHERE id = $1`,
		defID,
	).Scan(&slaCfgJSON)
	var slaCfg SLAConfig
	if len(slaCfgJSON) > 0 {
		_ = json.Unmarshal(slaCfgJSON, &slaCfg)
	}

	// Update instance pointer
	_, err = tx.Exec(ctx, `
		UPDATE workflow_instances SET current_step_id = $2, current_step_order = $3
		WHERE id = $1`,
		instanceID, targetStep.ID, targetStep.StepOrder,
	)
	if err != nil {
		return fmt.Errorf("failed to update instance to target step: %w", err)
	}

	return e.createStepExecution(ctx, tx, orgID, instanceID, targetStep, slaCfg.StepDefaultSLAHours)
}

// advanceWorkflow moves the workflow to the next step after the current one completes.
func (e *WorkflowEngine) advanceWorkflow(ctx context.Context, tx pgx.Tx, orgID, instanceID, currentStepID uuid.UUID, currentOrder int) error {
	// Get definition ID
	var defID uuid.UUID
	err := tx.QueryRow(ctx,
		`SELECT workflow_definition_id FROM workflow_instances WHERE id = $1`,
		instanceID,
	).Scan(&defID)
	if err != nil {
		return fmt.Errorf("failed to get definition ID: %w", err)
	}

	// Get the current step to check type
	currentStep, err := e.getStepByID(ctx, currentStepID)
	if err != nil {
		return fmt.Errorf("failed to get current step: %w", err)
	}

	// For condition steps, the routing is handled by processConditionStep
	// For other types, just get the next sequential step
	var nextStep *WorkflowStep

	if currentStep.StepType == "condition" {
		// Already handled in processConditionStep
		return nil
	}

	nextStep, err = e.getNextStep(ctx, defID, currentOrder)
	if err != nil {
		return fmt.Errorf("failed to find next step: %w", err)
	}

	if nextStep == nil {
		// Workflow is complete
		return e.completeWorkflow(ctx, tx, instanceID, orgID, "approved")
	}

	// Get SLA config
	var slaCfgJSON json.RawMessage
	_ = tx.QueryRow(ctx,
		`SELECT sla_config FROM workflow_definitions WHERE id = $1`,
		defID,
	).Scan(&slaCfgJSON)
	var slaCfg SLAConfig
	if len(slaCfgJSON) > 0 {
		_ = json.Unmarshal(slaCfgJSON, &slaCfg)
	}

	// Update instance to point to next step
	_, err = tx.Exec(ctx, `
		UPDATE workflow_instances SET current_step_id = $2, current_step_order = $3
		WHERE id = $1`,
		instanceID, nextStep.ID, nextStep.StepOrder,
	)
	if err != nil {
		return fmt.Errorf("failed to update instance step: %w", err)
	}

	// Create execution for next step
	return e.createStepExecution(ctx, tx, orgID, instanceID, nextStep, slaCfg.StepDefaultSLAHours)
}

// completeWorkflow marks a workflow instance as completed with the given outcome.
func (e *WorkflowEngine) completeWorkflow(ctx context.Context, tx pgx.Tx, instanceID, orgID uuid.UUID, outcome string) error {
	now := time.Now()

	// Calculate total duration
	var startedAt time.Time
	err := tx.QueryRow(ctx,
		`SELECT started_at FROM workflow_instances WHERE id = $1`,
		instanceID,
	).Scan(&startedAt)
	if err != nil {
		return fmt.Errorf("failed to get instance start time: %w", err)
	}

	durationHours := now.Sub(startedAt).Hours()
	durationHours = math.Round(durationHours*100) / 100

	_, err = tx.Exec(ctx, `
		UPDATE workflow_instances SET
			status = 'completed'::workflow_instance_status,
			completed_at = $3,
			completion_outcome = $4::workflow_completion_outcome,
			total_duration_hours = $5,
			current_step_id = NULL
		WHERE id = $1 AND organization_id = $2`,
		instanceID, orgID, now, outcome, durationHours,
	)
	if err != nil {
		return fmt.Errorf("failed to complete workflow: %w", err)
	}

	log.Info().
		Str("instance_id", instanceID.String()).
		Str("outcome", outcome).
		Float64("duration_hours", durationHours).
		Msg("Workflow completed")

	return nil
}

// checkParallelCompletion checks if all required parallel approvals for a step have been met.
func (e *WorkflowEngine) checkParallelCompletion(ctx context.Context, tx pgx.Tx, instanceID, stepID uuid.UUID, approvalMode string, minimumApprovals int) (bool, error) {
	var totalExecs, completedExecs, approvedExecs int
	err := tx.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE status IN ('approved', 'completed')),
			COUNT(*) FILTER (WHERE status = 'approved')
		FROM workflow_step_executions
		WHERE workflow_instance_id = $1 AND workflow_step_id = $2`,
		instanceID, stepID,
	).Scan(&totalExecs, &completedExecs, &approvedExecs)
	if err != nil {
		return false, fmt.Errorf("failed to check parallel completion: %w", err)
	}

	switch approvalMode {
	case "all_required":
		return completedExecs >= totalExecs && totalExecs > 0, nil
	case "majority":
		needed := (totalExecs / 2) + 1
		return approvedExecs >= needed, nil
	case "any_one":
		return approvedExecs >= 1, nil
	default:
		if minimumApprovals > 0 {
			return approvedExecs >= minimumApprovals, nil
		}
		return completedExecs >= totalExecs && totalExecs > 0, nil
	}
}

// getInstanceExecutions returns all executions for a given workflow instance.
func (e *WorkflowEngine) getInstanceExecutions(ctx context.Context, orgID, instanceID uuid.UUID) ([]StepExecution, error) {
	rows, err := e.pool.Query(ctx, `
		SELECT
			e.id, e.organization_id, e.workflow_instance_id, e.workflow_step_id,
			e.step_order, e.status::text, e.assigned_to, e.delegated_to,
			e.delegated_by, e.delegated_at, e.action_taken_by, e.action_taken_at,
			e.action::text, e.comments, e.decision_reason, e.attachments_paths,
			e.sla_deadline, e.sla_status::text, e.escalated_at, e.escalated_to,
			e.started_at, e.completed_at, e.duration_hours, e.created_at,
			s.name, s.step_type::text
		FROM workflow_step_executions e
		JOIN workflow_steps s ON s.id = e.workflow_step_id
		WHERE e.workflow_instance_id = $1 AND e.organization_id = $2
		ORDER BY e.step_order ASC, e.created_at ASC`,
		instanceID, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query executions: %w", err)
	}
	defer rows.Close()

	var execs []StepExecution
	for rows.Next() {
		var exec StepExecution
		if err := rows.Scan(
			&exec.ID, &exec.OrganizationID, &exec.WorkflowInstanceID, &exec.WorkflowStepID,
			&exec.StepOrder, &exec.Status, &exec.AssignedTo, &exec.DelegatedTo,
			&exec.DelegatedBy, &exec.DelegatedAt, &exec.ActionTakenBy, &exec.ActionTakenAt,
			&exec.Action, &exec.Comments, &exec.DecisionReason, &exec.AttachmentsPaths,
			&exec.SLADeadline, &exec.SLAStatus, &exec.EscalatedAt, &exec.EscalatedTo,
			&exec.StartedAt, &exec.CompletedAt, &exec.DurationHours, &exec.CreatedAt,
			&exec.StepName, &exec.StepType,
		); err != nil {
			return nil, fmt.Errorf("failed to scan execution: %w", err)
		}
		execs = append(execs, exec)
	}

	return execs, nil
}

// findDelegateForUser checks if there is an active delegation rule for the given user and workflow type.
func (e *WorkflowEngine) findDelegateForUser(ctx context.Context, orgID, userID uuid.UUID, workflowType string) *uuid.UUID {
	now := time.Now()
	var delegateID uuid.UUID

	err := e.pool.QueryRow(ctx, `
		SELECT delegate_user_id FROM workflow_delegation_rules
		WHERE organization_id = $1
			AND delegator_user_id = $2
			AND is_active = true
			AND valid_from <= $3
			AND (valid_until IS NULL OR valid_until >= $3)
			AND ($4 = '' OR workflow_types = '{}' OR $4 = ANY(workflow_types))
		ORDER BY created_at DESC
		LIMIT 1`,
		orgID, userID, now, workflowType,
	).Scan(&delegateID)

	if err != nil {
		return nil
	}
	return &delegateID
}

// ============================================================
// UTILITY FUNCTIONS
// ============================================================

// evaluateCondition evaluates a simple condition expression against JSON metadata.
func evaluateCondition(expr ConditionExpression, metadata json.RawMessage) bool {
	if expr.Field == "" {
		return false
	}

	// Parse metadata into a map
	var meta map[string]interface{}
	if err := json.Unmarshal(metadata, &meta); err != nil {
		return false
	}

	fieldValue, ok := meta[expr.Field]
	if !ok {
		return false
	}

	switch expr.Operator {
	case "eq", "equals", "==":
		return fmt.Sprintf("%v", fieldValue) == fmt.Sprintf("%v", expr.Value)
	case "neq", "not_equals", "!=":
		return fmt.Sprintf("%v", fieldValue) != fmt.Sprintf("%v", expr.Value)
	case "in":
		valueList, ok := expr.Value.([]interface{})
		if !ok {
			return false
		}
		fieldStr := fmt.Sprintf("%v", fieldValue)
		for _, v := range valueList {
			if fmt.Sprintf("%v", v) == fieldStr {
				return true
			}
		}
		return false
	case "not_in":
		valueList, ok := expr.Value.([]interface{})
		if !ok {
			return true
		}
		fieldStr := fmt.Sprintf("%v", fieldValue)
		for _, v := range valueList {
			if fmt.Sprintf("%v", v) == fieldStr {
				return false
			}
		}
		return true
	case "gt", ">":
		return toFloat(fieldValue) > toFloat(expr.Value)
	case "gte", ">=":
		return toFloat(fieldValue) >= toFloat(expr.Value)
	case "lt", "<":
		return toFloat(fieldValue) < toFloat(expr.Value)
	case "lte", "<=":
		return toFloat(fieldValue) <= toFloat(expr.Value)
	case "contains":
		return strings.Contains(fmt.Sprintf("%v", fieldValue), fmt.Sprintf("%v", expr.Value))
	default:
		return false
	}
}

// toFloat attempts to convert an interface value to float64 for numeric comparisons.
func toFloat(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case json.Number:
		f, _ := val.Float64()
		return f
	default:
		return 0
	}
}

// isValidAction returns true if the action is valid for the given step type.
func isValidAction(stepType, action string) bool {
	switch stepType {
	case "approval":
		return action == "approve" || action == "reject"
	case "review":
		return action == "approve" || action == "reject" || action == "complete"
	case "task":
		return action == "complete" || action == "reject"
	case "parallel_gate":
		return action == "approve" || action == "reject" || action == "complete"
	default:
		return action == "complete"
	}
}

// resolveExecutionStatus maps an action string to the resulting step execution status.
func resolveExecutionStatus(action string) string {
	switch action {
	case "approve":
		return "approved"
	case "reject":
		return "rejected"
	case "complete":
		return "completed"
	case "skip":
		return "skipped"
	case "delegate":
		return "delegated"
	case "escalate":
		return "escalated"
	default:
		return "completed"
	}
}

// defaultJSONB returns the input if non-nil/non-empty, otherwise returns "{}".
func defaultJSONB(data json.RawMessage) json.RawMessage {
	if len(data) == 0 {
		return json.RawMessage(`{}`)
	}
	return data
}

// nullableApprovalMode returns the input if non-empty, otherwise returns "any_one" as a default.
func nullableApprovalMode(mode string) string {
	if mode == "" {
		return "any_one"
	}
	return mode
}

// scanWorkflowStep scans a workflow step from a pgx.Rows result.
func scanWorkflowStep(rows pgx.Rows) (WorkflowStep, error) {
	var s WorkflowStep
	if err := rows.Scan(
		&s.ID, &s.WorkflowDefinitionID, &s.OrganizationID, &s.StepOrder, &s.Name, &s.Description,
		&s.StepType, &s.ApproverType, &s.ApproverIDs, &s.ApprovalMode,
		&s.MinimumApprovals, &s.TaskDescription, &s.TaskAssigneeType, &s.TaskAssigneeIDs,
		&s.ConditionExpression, &s.ConditionTrueStepID, &s.ConditionFalseStepID,
		&s.AutoAction, &s.TimerHours, &s.TimerBusinessHoursOnly,
		&s.SLAHours, &s.EscalationUserIDs, &s.IsOptional, &s.CanDelegate,
		&s.CreatedAt, &s.UpdatedAt,
	); err != nil {
		return s, fmt.Errorf("failed to scan workflow step: %w", err)
	}
	return s, nil
}
