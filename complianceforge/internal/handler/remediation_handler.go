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

// RemediationHandler handles HTTP requests for the AI-assisted compliance
// remediation planner, including plan CRUD, AI-powered generation, action
// management, and AI utility endpoints.
type RemediationHandler struct {
	planner   *service.RemediationPlanner
	aiService *service.AIService
}

// NewRemediationHandler creates a new RemediationHandler.
func NewRemediationHandler(planner *service.RemediationPlanner, aiService *service.AIService) *RemediationHandler {
	return &RemediationHandler{
		planner:   planner,
		aiService: aiService,
	}
}

// RegisterRoutes mounts all remediation and AI endpoints on the given router.
func (h *RemediationHandler) RegisterRoutes(r chi.Router) {
	// Plan endpoints
	r.Get("/plans", h.ListPlans)
	r.Post("/plans", h.CreatePlan)
	r.Post("/plans/generate", h.GeneratePlan)
	r.Get("/plans/{id}", h.GetPlan)
	r.Put("/plans/{id}", h.UpdatePlan)
	r.Post("/plans/{id}/approve", h.ApprovePlan)
	r.Get("/plans/{id}/timeline", h.GetTimeline)
	r.Get("/plans/{id}/progress", h.GetProgress)

	// Action endpoints
	r.Put("/actions/{id}", h.UpdateAction)
	r.Post("/actions/{id}/complete", h.CompleteAction)

	// AI utility endpoints
	r.Post("/ai/control-guidance", h.ControlGuidance)
	r.Post("/ai/evidence-suggestion", h.EvidenceSuggestion)
	r.Post("/ai/policy-draft", h.PolicyDraft)
	r.Post("/ai/risk-narrative", h.RiskNarrative)
	r.Get("/ai/usage", h.AIUsage)
	r.Post("/ai/feedback", h.AIFeedback)
}

// ============================================================
// PLAN ENDPOINTS
// ============================================================

// ListPlans returns a paginated list of remediation plans.
// GET /plans
func (h *RemediationHandler) ListPlans(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	params := parsePagination(r)

	plans, total, err := h.planner.ListPlans(r.Context(), orgID, params.Page, params.PageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list remediation plans")
		return
	}

	if plans == nil {
		plans = []service.RemediationPlan{}
	}

	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize != 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: plans,
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

// CreatePlan creates a new remediation plan without AI generation.
// POST /plans
func (h *RemediationHandler) CreatePlan(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var req service.GeneratePlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "Plan name is required")
		return
	}

	if req.PlanType == "" {
		req.PlanType = "gap_remediation"
	}
	if req.Priority == "" {
		req.Priority = "medium"
	}

	// Disable AI for manual creation
	req.UseAI = false

	plan, err := h.planner.GeneratePlan(r.Context(), orgID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create remediation plan")
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: plan})
}

// GeneratePlan creates a new AI-generated remediation plan.
// POST /plans/generate
func (h *RemediationHandler) GeneratePlan(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var req service.GeneratePlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Name == "" {
		req.Name = "AI-Generated Remediation Plan"
	}
	if req.PlanType == "" {
		req.PlanType = "gap_remediation"
	}
	if req.Priority == "" {
		req.Priority = "medium"
	}

	req.UseAI = true
	if req.AIRequest == nil {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "AI request configuration is required for AI-generated plans")
		return
	}

	plan, err := h.planner.GeneratePlan(r.Context(), orgID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "AI_GENERATION_ERROR", "Failed to generate AI remediation plan")
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: plan})
}

// GetPlan retrieves a single remediation plan by ID.
// GET /plans/{id}
func (h *RemediationHandler) GetPlan(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	planID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid plan ID")
		return
	}

	plan, err := h.planner.GetPlan(r.Context(), orgID, planID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Remediation plan not found")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: plan})
}

// UpdatePlan updates an existing remediation plan.
// PUT /plans/{id}
func (h *RemediationHandler) UpdatePlan(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	planID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid plan ID")
		return
	}

	var req service.UpdatePlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if err := h.planner.UpdatePlan(r.Context(), orgID, planID, req); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update remediation plan")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: map[string]string{"message": "Plan updated successfully"}})
}

// ApprovePlan marks a plan as approved.
// POST /plans/{id}/approve
func (h *RemediationHandler) ApprovePlan(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	planID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid plan ID")
		return
	}

	if err := h.planner.ApprovePlan(r.Context(), orgID, planID, userID); err != nil {
		writeError(w, http.StatusBadRequest, "APPROVAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: map[string]string{"message": "Plan approved successfully"}})
}

// GetTimeline returns the timeline view of a plan.
// GET /plans/{id}/timeline
func (h *RemediationHandler) GetTimeline(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	planID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid plan ID")
		return
	}

	timeline, err := h.planner.GetTimeline(r.Context(), orgID, planID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Plan not found")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: timeline})
}

// GetProgress returns progress tracking data for a plan.
// GET /plans/{id}/progress
func (h *RemediationHandler) GetProgress(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	planID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid plan ID")
		return
	}

	progress, err := h.planner.TrackProgress(r.Context(), orgID, planID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Plan not found")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: progress})
}

// ============================================================
// ACTION ENDPOINTS
// ============================================================

// UpdateAction updates an existing remediation action.
// PUT /actions/{id}
func (h *RemediationHandler) UpdateAction(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	actionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid action ID")
		return
	}

	var req service.UpdateActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if err := h.planner.UpdateAction(r.Context(), orgID, actionID, req); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update action")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: map[string]string{"message": "Action updated successfully"}})
}

// CompleteAction marks an action as complete with evidence and actuals.
// POST /actions/{id}/complete
func (h *RemediationHandler) CompleteAction(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	actionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid action ID")
		return
	}

	var req service.CompleteActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if err := h.planner.CompleteAction(r.Context(), orgID, actionID, req); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to complete action")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: map[string]string{"message": "Action completed successfully"}})
}

// ============================================================
// AI UTILITY ENDPOINTS
// ============================================================

// ControlGuidance generates AI implementation guidance for a control.
// POST /ai/control-guidance
func (h *RemediationHandler) ControlGuidance(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var req service.ControlGuidanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.ControlCode == "" || req.FrameworkCode == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "control_code and framework_code are required")
		return
	}

	guidance, err := h.aiService.GenerateControlGuidance(r.Context(), orgID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "AI_ERROR", "Failed to generate control guidance")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: guidance})
}

// EvidenceSuggestion generates AI evidence collection suggestions for a control.
// POST /ai/evidence-suggestion
func (h *RemediationHandler) EvidenceSuggestion(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var req struct {
		ControlCode  string `json:"control_code"`
		ControlTitle string `json:"control_title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.ControlCode == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "control_code is required")
		return
	}

	template, err := h.aiService.SuggestEvidenceTemplate(r.Context(), orgID, req.ControlCode, req.ControlTitle)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "AI_ERROR", "Failed to generate evidence suggestions")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: template})
}

// PolicyDraft generates an AI-drafted policy section.
// POST /ai/policy-draft
func (h *RemediationHandler) PolicyDraft(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var req service.PolicyDraftRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.PolicyType == "" || req.SectionTitle == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "policy_type and section_title are required")
		return
	}

	draft, err := h.aiService.DraftPolicySection(r.Context(), orgID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "AI_ERROR", "Failed to draft policy section")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: draft})
}

// RiskNarrative generates an AI risk narrative.
// POST /ai/risk-narrative
func (h *RemediationHandler) RiskNarrative(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var req service.RiskNarrativeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.RiskTitle == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "risk_title is required")
		return
	}

	narrative, err := h.aiService.AssessRiskNarrative(r.Context(), orgID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "AI_ERROR", "Failed to generate risk narrative")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: narrative})
}

// AIUsage returns AI usage statistics for the organization.
// GET /ai/usage
func (h *RemediationHandler) AIUsage(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	stats, err := h.aiService.GetUsageStats(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch AI usage statistics")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: stats})
}

// AIFeedback records user feedback for an AI interaction.
// POST /ai/feedback
func (h *RemediationHandler) AIFeedback(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var req service.AIFeedbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.InteractionID == uuid.Nil {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "interaction_id is required")
		return
	}
	if req.Rating < 1 || req.Rating > 5 {
		writeError(w, http.StatusBadRequest, "INVALID_RATING", "Rating must be between 1 and 5")
		return
	}

	if err := h.aiService.RateInteraction(r.Context(), orgID, req.InteractionID, req.Rating, req.Feedback); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to record feedback")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: map[string]string{"message": "Feedback recorded successfully"}})
}
