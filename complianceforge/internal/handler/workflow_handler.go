package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/complianceforge/platform/internal/middleware"
	"github.com/complianceforge/platform/internal/models"
	"github.com/complianceforge/platform/internal/service"
)

// WorkflowHandler handles HTTP requests for the Compliance Workflow Engine.
type WorkflowHandler struct {
	engine *service.WorkflowEngine
}

// NewWorkflowHandler creates a new WorkflowHandler.
func NewWorkflowHandler(engine *service.WorkflowEngine) *WorkflowHandler {
	return &WorkflowHandler{engine: engine}
}

// ============================================================
// DEFINITION ENDPOINTS
// ============================================================

// ListDefinitions returns paginated workflow definitions.
// GET /api/v1/workflows/definitions
func (h *WorkflowHandler) ListDefinitions(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	params := parsePagination(r)

	defs, total, err := h.engine.ListDefinitions(r.Context(), orgID, params.PageSize, params.Offset())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list workflow definitions")
		return
	}

	if defs == nil {
		defs = []service.WorkflowDefinition{}
	}

	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize != 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: defs,
		Pagination: models.Pagination{
			Page:       params.Page,
			PageSize:   params.PageSize,
			TotalItems: total,
			TotalPages: totalPages,
			HasNext:    params.Page < totalPages,
			HasPrev:    params.Page > 1,
		},
	})
}

// CreateDefinition creates a new workflow definition.
// POST /api/v1/workflows/definitions
func (h *WorkflowHandler) CreateDefinition(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var input service.CreateDefinitionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if input.Name == "" || input.WorkflowType == "" || input.EntityType == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "Name, workflow_type, and entity_type are required")
		return
	}

	def, err := h.engine.CreateDefinition(r.Context(), orgID, input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: def})
}

// UpdateDefinition updates an existing workflow definition.
// PUT /api/v1/workflows/definitions/{id}
func (h *WorkflowHandler) UpdateDefinition(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	defID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid workflow definition ID")
		return
	}

	var input service.UpdateDefinitionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	def, err := h.engine.UpdateDefinition(r.Context(), orgID, defID, input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: def})
}

// ActivateDefinition activates a draft workflow definition.
// POST /api/v1/workflows/definitions/{id}/activate
func (h *WorkflowHandler) ActivateDefinition(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	defID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid workflow definition ID")
		return
	}

	if err := h.engine.ActivateDefinition(r.Context(), orgID, defID); err != nil {
		writeError(w, http.StatusBadRequest, "ACTIVATION_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: map[string]string{
		"message": "Workflow definition activated successfully",
	}})
}

// ============================================================
// STEP ENDPOINTS
// ============================================================

// GetDefinitionSteps returns all steps for a workflow definition.
// GET /api/v1/workflows/definitions/{id}/steps
func (h *WorkflowHandler) GetDefinitionSteps(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	defID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid workflow definition ID")
		return
	}

	steps, err := h.engine.GetDefinitionSteps(r.Context(), orgID, defID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list workflow steps")
		return
	}

	if steps == nil {
		steps = []service.WorkflowStep{}
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: steps})
}

// AddStep adds a new step to a workflow definition.
// POST /api/v1/workflows/definitions/{id}/steps
func (h *WorkflowHandler) AddStep(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	defID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid workflow definition ID")
		return
	}

	var input service.StepInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if input.Name == "" || input.StepType == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "Name and step_type are required")
		return
	}

	step, err := h.engine.AddStep(r.Context(), orgID, defID, input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: step})
}

// UpdateStep updates an existing workflow step.
// PUT /api/v1/workflows/definitions/{id}/steps/{stepId}
func (h *WorkflowHandler) UpdateStep(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	defID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid workflow definition ID")
		return
	}
	stepID, err := uuid.Parse(chi.URLParam(r, "stepId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid step ID")
		return
	}

	var input service.StepInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	step, err := h.engine.UpdateStep(r.Context(), orgID, defID, stepID, input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: step})
}

// DeleteStep removes a step from a workflow definition.
// DELETE /api/v1/workflows/definitions/{id}/steps/{stepId}
func (h *WorkflowHandler) DeleteStep(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	defID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid workflow definition ID")
		return
	}
	stepID, err := uuid.Parse(chi.URLParam(r, "stepId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid step ID")
		return
	}

	if err := h.engine.DeleteStep(r.Context(), orgID, defID, stepID); err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: map[string]string{
		"message": "Step deleted successfully",
	}})
}

// ============================================================
// INSTANCE ENDPOINTS
// ============================================================

// ListInstances returns paginated workflow instances with optional filters.
// GET /api/v1/workflows/instances?status=active&entity_type=policy&sla_status=on_track
func (h *WorkflowHandler) ListInstances(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	params := parsePagination(r)

	filters := service.InstanceFilters{
		Status:     r.URL.Query().Get("status"),
		EntityType: r.URL.Query().Get("entity_type"),
		SLAStatus:  r.URL.Query().Get("sla_status"),
	}

	instances, total, err := h.engine.ListInstances(r.Context(), orgID, filters, params.PageSize, params.Offset())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list workflow instances")
		return
	}

	if instances == nil {
		instances = []service.WorkflowInstance{}
	}

	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize != 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: instances,
		Pagination: models.Pagination{
			Page:       params.Page,
			PageSize:   params.PageSize,
			TotalItems: total,
			TotalPages: totalPages,
			HasNext:    params.Page < totalPages,
			HasPrev:    params.Page > 1,
		},
	})
}

// GetInstance returns a single workflow instance with its step executions.
// GET /api/v1/workflows/instances/{id}
func (h *WorkflowHandler) GetInstance(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	instanceID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid workflow instance ID")
		return
	}

	instance, err := h.engine.GetInstance(r.Context(), orgID, instanceID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Workflow instance not found")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: instance})
}

// StartWorkflow starts a new workflow instance for an entity.
// POST /api/v1/workflows/start
func (h *WorkflowHandler) StartWorkflow(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())

	var req struct {
		WorkflowType string    `json:"workflow_type"`
		EntityType   string    `json:"entity_type"`
		EntityID     uuid.UUID `json:"entity_id"`
		EntityRef    string    `json:"entity_ref"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.WorkflowType == "" || req.EntityType == "" || req.EntityID == uuid.Nil {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "workflow_type, entity_type, and entity_id are required")
		return
	}

	instance, err := h.engine.StartWorkflow(r.Context(), orgID, req.WorkflowType, req.EntityType, req.EntityID, req.EntityRef, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "WORKFLOW_START_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: instance})
}

// CancelWorkflow cancels an active workflow instance.
// POST /api/v1/workflows/instances/{id}/cancel
func (h *WorkflowHandler) CancelWorkflow(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	instanceID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid workflow instance ID")
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if err := h.engine.CancelWorkflow(r.Context(), orgID, instanceID, userID, req.Reason); err != nil {
		writeError(w, http.StatusInternalServerError, "CANCEL_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: map[string]string{
		"message": "Workflow cancelled successfully",
	}})
}

// ============================================================
// APPROVAL / EXECUTION ENDPOINTS
// ============================================================

// GetMyApprovals returns all pending approvals for the current user.
// GET /api/v1/workflows/my-approvals
func (h *WorkflowHandler) GetMyApprovals(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())

	approvals, err := h.engine.GetPendingApprovals(r.Context(), orgID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve pending approvals")
		return
	}

	if approvals == nil {
		approvals = []service.StepExecution{}
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: approvals})
}

// ApproveExecution approves a step execution.
// POST /api/v1/workflows/executions/{id}/approve
func (h *WorkflowHandler) ApproveExecution(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	execID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid execution ID")
		return
	}

	var req struct {
		Comments string `json:"comments"`
		Reason   string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Allow empty body for simple approvals
		req.Comments = ""
		req.Reason = ""
	}

	exec, err := h.engine.ProcessStep(r.Context(), orgID, execID, "approve", userID, req.Comments, req.Reason)
	if err != nil {
		writeError(w, http.StatusBadRequest, "APPROVAL_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: exec})
}

// RejectExecution rejects a step execution.
// POST /api/v1/workflows/executions/{id}/reject
func (h *WorkflowHandler) RejectExecution(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	execID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid execution ID")
		return
	}

	var req struct {
		Comments string `json:"comments"`
		Reason   string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Reason == "" {
		writeError(w, http.StatusBadRequest, "MISSING_REASON", "Reason is required when rejecting")
		return
	}

	exec, err := h.engine.ProcessStep(r.Context(), orgID, execID, "reject", userID, req.Comments, req.Reason)
	if err != nil {
		writeError(w, http.StatusBadRequest, "REJECTION_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: exec})
}

// DelegateExecution delegates a step execution to another user.
// POST /api/v1/workflows/executions/{id}/delegate
func (h *WorkflowHandler) DelegateExecution(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	execID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid execution ID")
		return
	}

	var req struct {
		DelegateUserID uuid.UUID `json:"delegate_user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.DelegateUserID == uuid.Nil {
		writeError(w, http.StatusBadRequest, "MISSING_DELEGATE", "delegate_user_id is required")
		return
	}

	if err := h.engine.DelegateStep(r.Context(), orgID, execID, userID, req.DelegateUserID); err != nil {
		writeError(w, http.StatusBadRequest, "DELEGATION_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: map[string]string{
		"message": "Step delegated successfully",
	}})
}

// ============================================================
// DELEGATION ENDPOINTS
// ============================================================

// ListDelegations returns all delegation rules for the organization.
// GET /api/v1/workflows/delegations
func (h *WorkflowHandler) ListDelegations(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	rules, err := h.engine.ListDelegations(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list delegation rules")
		return
	}

	if rules == nil {
		rules = []service.DelegationRule{}
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: rules})
}

// CreateDelegation creates a new delegation rule.
// POST /api/v1/workflows/delegations
func (h *WorkflowHandler) CreateDelegation(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var input service.CreateDelegationInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if input.DelegatorUserID == uuid.Nil || input.DelegateUserID == uuid.Nil {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "delegator_user_id and delegate_user_id are required")
		return
	}

	rule, err := h.engine.CreateDelegation(r.Context(), orgID, input)
	if err != nil {
		writeError(w, http.StatusBadRequest, "DELEGATION_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: rule})
}

// ============================================================
// ROUTE REGISTRATION
// ============================================================

// RegisterRoutes registers all workflow handler routes on a chi router.
func (h *WorkflowHandler) RegisterRoutes(r chi.Router) {
	r.Route("/workflows", func(r chi.Router) {
		// Definitions
		r.Get("/definitions", h.ListDefinitions)
		r.Post("/definitions", h.CreateDefinition)
		r.Put("/definitions/{id}", h.UpdateDefinition)
		r.Post("/definitions/{id}/activate", h.ActivateDefinition)
		r.Get("/definitions/{id}/steps", h.GetDefinitionSteps)
		r.Post("/definitions/{id}/steps", h.AddStep)
		r.Put("/definitions/{id}/steps/{stepId}", h.UpdateStep)
		r.Delete("/definitions/{id}/steps/{stepId}", h.DeleteStep)

		// Instances
		r.Get("/instances", h.ListInstances)
		r.Get("/instances/{id}", h.GetInstance)
		r.Post("/start", h.StartWorkflow)
		r.Post("/instances/{id}/cancel", h.CancelWorkflow)

		// Approvals
		r.Get("/my-approvals", h.GetMyApprovals)
		r.Post("/executions/{id}/approve", h.ApproveExecution)
		r.Post("/executions/{id}/reject", h.RejectExecution)
		r.Post("/executions/{id}/delegate", h.DelegateExecution)

		// Delegations
		r.Get("/delegations", h.ListDelegations)
		r.Post("/delegations", h.CreateDelegation)
	})
}
