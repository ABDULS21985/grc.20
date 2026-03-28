package handler

import (
	"encoding/json"
	"net/http"

	"github.com/complianceforge/platform/internal/middleware"
	"github.com/complianceforge/platform/internal/models"
	"github.com/complianceforge/platform/internal/service"
)

// ============================================================
// SUBSCRIPTION HANDLER
// Handles HTTP endpoints for subscription management, plan
// changes, and usage tracking.
// ============================================================

// SubscriptionHandler handles subscription-related HTTP requests.
type SubscriptionHandler struct {
	subscriptionSvc *service.SubscriptionService
}

// NewSubscriptionHandler creates a new SubscriptionHandler.
func NewSubscriptionHandler(subscriptionSvc *service.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{subscriptionSvc: subscriptionSvc}
}

// ============================================================
// GET /subscription — Current subscription + usage
// ============================================================

// GetSubscription returns the current subscription for the authenticated organisation.
func (h *SubscriptionHandler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID.String() == "00000000-0000-0000-0000-000000000000" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Organisation context required")
		return
	}

	sub, err := h.subscriptionSvc.GetSubscription(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "No subscription found: "+err.Error())
		return
	}

	// Also get usage summary
	usage, err := h.subscriptionSvc.GetUsageSummary(r.Context(), orgID)
	if err != nil {
		// Return subscription without usage if usage fails
		writeJSON(w, http.StatusOK, models.APIResponse{
			Success: true,
			Data:    sub,
		})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"subscription": sub,
			"usage":        usage,
		},
	})
}

// ============================================================
// PUT /subscription/plan — Change plan
// ============================================================

// changePlanRequest is the request body for changing subscription plan.
type changePlanRequest struct {
	PlanSlug string `json:"plan_slug"`
}

// ChangePlan changes the subscription plan for the authenticated organisation.
func (h *SubscriptionHandler) ChangePlan(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID.String() == "00000000-0000-0000-0000-000000000000" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Organisation context required")
		return
	}

	var req changePlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.PlanSlug == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "plan_slug is required")
		return
	}

	sub, err := h.subscriptionSvc.ChangePlan(r.Context(), orgID, req.PlanSlug)
	if err != nil {
		writeError(w, http.StatusBadRequest, "PLAN_CHANGE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    sub,
	})
}

// ============================================================
// POST /subscription/cancel — Cancel subscription
// ============================================================

// cancelSubscriptionRequest is the request body for cancelling a subscription.
type cancelSubscriptionRequest struct {
	Reason string `json:"reason"`
}

// CancelSubscription cancels the subscription for the authenticated organisation.
func (h *SubscriptionHandler) CancelSubscription(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID.String() == "00000000-0000-0000-0000-000000000000" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Organisation context required")
		return
	}

	var req cancelSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	sub, err := h.subscriptionSvc.CancelSubscription(r.Context(), orgID, req.Reason)
	if err != nil {
		writeError(w, http.StatusBadRequest, "CANCEL_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    sub,
	})
}

// ============================================================
// POST /subscription/pause — Pause subscription
// ============================================================

// PauseSubscription pauses the subscription for the authenticated organisation.
func (h *SubscriptionHandler) PauseSubscription(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID.String() == "00000000-0000-0000-0000-000000000000" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Organisation context required")
		return
	}

	sub, err := h.subscriptionSvc.PauseSubscription(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "PAUSE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    sub,
	})
}

// ============================================================
// POST /subscription/resume — Resume subscription
// ============================================================

// ResumeSubscription resumes the subscription for the authenticated organisation.
func (h *SubscriptionHandler) ResumeSubscription(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID.String() == "00000000-0000-0000-0000-000000000000" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Organisation context required")
		return
	}

	sub, err := h.subscriptionSvc.ResumeSubscription(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "RESUME_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    sub,
	})
}

// ============================================================
// GET /subscription/plans — Available plans
// ============================================================

// ListPlans returns all available subscription plans.
func (h *SubscriptionHandler) ListPlans(w http.ResponseWriter, r *http.Request) {
	plans, err := h.subscriptionSvc.ListPlans(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve plans")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    plans,
	})
}

// ============================================================
// GET /subscription/usage — Detailed usage
// ============================================================

// GetUsage returns detailed usage data for the authenticated organisation.
func (h *SubscriptionHandler) GetUsage(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID.String() == "00000000-0000-0000-0000-000000000000" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Organisation context required")
		return
	}

	usage, err := h.subscriptionSvc.GetUsageSummary(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve usage data")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    usage,
	})
}

// ============================================================
// GET /subscription/limits/{resource} — Check specific limit
// ============================================================

// CheckLimit checks the plan limit for a specific resource type.
func (h *SubscriptionHandler) CheckLimit(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID.String() == "00000000-0000-0000-0000-000000000000" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Organisation context required")
		return
	}

	resource := r.URL.Query().Get("resource")
	if resource == "" {
		writeError(w, http.StatusBadRequest, "MISSING_PARAM", "resource query parameter is required")
		return
	}

	check, err := h.subscriptionSvc.CheckLimits(r.Context(), orgID, resource)
	if err != nil {
		writeError(w, http.StatusBadRequest, "CHECK_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    check,
	})
}

// ============================================================
// POST /subscription/create — Create subscription (for onboarding)
// ============================================================

// createSubscriptionRequest is the request body for creating a subscription.
type createSubscriptionRequest struct {
	PlanSlug     string `json:"plan_slug"`
	BillingCycle string `json:"billing_cycle"`
}

// CreateSubscription creates a new subscription for the authenticated organisation.
func (h *SubscriptionHandler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID.String() == "00000000-0000-0000-0000-000000000000" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Organisation context required")
		return
	}

	var req createSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.PlanSlug == "" {
		req.PlanSlug = "starter"
	}
	if req.BillingCycle == "" {
		req.BillingCycle = "monthly"
	}

	sub, err := h.subscriptionSvc.CreateSubscription(r.Context(), orgID, req.PlanSlug, req.BillingCycle)
	if err != nil {
		writeError(w, http.StatusBadRequest, "CREATE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{
		Success: true,
		Data:    sub,
	})
}
