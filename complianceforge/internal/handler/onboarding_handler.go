package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/complianceforge/platform/internal/middleware"
	"github.com/complianceforge/platform/internal/models"
	"github.com/complianceforge/platform/internal/service"
)

// OnboardingHandler handles HTTP requests for organisation onboarding.
// It supports both the legacy single-shot onboarding (POST /onboard)
// and the new multi-step wizard endpoints.
type OnboardingHandler struct {
	onboardingSvc *service.OnboardingService
	wizardSvc     *service.OnboardingWizard
}

// NewOnboardingHandler creates a new OnboardingHandler with legacy service only.
func NewOnboardingHandler(onboardingSvc *service.OnboardingService) *OnboardingHandler {
	return &OnboardingHandler{onboardingSvc: onboardingSvc}
}

// NewOnboardingHandlerWithWizard creates a new OnboardingHandler with both legacy and wizard services.
func NewOnboardingHandlerWithWizard(onboardingSvc *service.OnboardingService, wizardSvc *service.OnboardingWizard) *OnboardingHandler {
	return &OnboardingHandler{
		onboardingSvc: onboardingSvc,
		wizardSvc:     wizardSvc,
	}
}

// Onboard creates a complete new organisation with admin user, frameworks, and controls.
// POST /api/v1/onboard
func (h *OnboardingHandler) Onboard(w http.ResponseWriter, r *http.Request) {
	var req service.OnboardingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	// Validate required fields
	if req.OrgName == "" || req.AdminEmail == "" || req.AdminFirstName == "" || req.AdminLastName == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS",
			"Required: org_name, admin_email, admin_first_name, admin_last_name")
		return
	}

	if req.CountryCode == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "country_code is required")
		return
	}

	if req.Industry == "" {
		req.Industry = "technology"
	}
	if req.Timezone == "" {
		req.Timezone = "Europe/London"
	}
	if req.PlanName == "" {
		req.PlanName = "starter"
	}

	result, err := h.onboardingSvc.Onboard(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ONBOARDING_FAILED", "Failed to onboard organisation: "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{
		Success: true,
		Data:    result,
	})
}

// ============================================================
// WIZARD ENDPOINTS
// ============================================================

// GetProgress returns the current onboarding wizard progress.
// GET /api/v1/onboard/progress
func (h *OnboardingHandler) GetProgress(w http.ResponseWriter, r *http.Request) {
	if h.wizardSvc == nil {
		writeError(w, http.StatusNotImplemented, "NOT_IMPLEMENTED", "Onboarding wizard not configured")
		return
	}

	orgID := middleware.GetOrgID(r.Context())
	if orgID.String() == "00000000-0000-0000-0000-000000000000" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Organisation context required")
		return
	}

	progress, err := h.wizardSvc.GetProgress(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get onboarding progress: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    progress,
	})
}

// SaveStep saves data for a specific onboarding wizard step.
// PUT /api/v1/onboard/step/{n}
func (h *OnboardingHandler) SaveStep(w http.ResponseWriter, r *http.Request) {
	if h.wizardSvc == nil {
		writeError(w, http.StatusNotImplemented, "NOT_IMPLEMENTED", "Onboarding wizard not configured")
		return
	}

	orgID := middleware.GetOrgID(r.Context())
	if orgID.String() == "00000000-0000-0000-0000-000000000000" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Organisation context required")
		return
	}

	stepStr := chi.URLParam(r, "n")
	step, err := strconv.Atoi(stepStr)
	if err != nil || step < 1 || step > 7 {
		writeError(w, http.StatusBadRequest, "INVALID_STEP", "Step must be a number between 1 and 7")
		return
	}

	var progress *service.OnboardingProgress

	switch step {
	case 1:
		var data service.OrgProfileInput
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid org profile data")
			return
		}
		progress, err = h.wizardSvc.SaveOrgProfile(r.Context(), orgID, data)

	case 2:
		var data service.IndustryAssessmentInput
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid industry assessment data")
			return
		}
		progress, err = h.wizardSvc.SaveIndustryAssessment(r.Context(), orgID, data)

	case 3:
		var data struct {
			FrameworkIDs []string `json:"framework_ids"`
		}
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid framework selection data")
			return
		}
		var fwIDs []uuid.UUID
		for _, idStr := range data.FrameworkIDs {
			id, parseErr := uuid.Parse(idStr)
			if parseErr != nil {
				writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid framework ID: "+idStr)
				return
			}
			fwIDs = append(fwIDs, id)
		}
		progress, err = h.wizardSvc.SaveFrameworkSelection(r.Context(), orgID, fwIDs)

	case 4:
		var data struct {
			Invitations []service.TeamInvitation `json:"invitations"`
		}
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid team invitation data")
			return
		}
		progress, err = h.wizardSvc.SaveTeamInvitations(r.Context(), orgID, data.Invitations)

	case 5:
		var data service.RiskAppetiteInput
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid risk appetite data")
			return
		}
		progress, err = h.wizardSvc.SaveRiskAppetite(r.Context(), orgID, data)

	case 6:
		var data service.QuickAssessmentInput
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid quick assessment data")
			return
		}
		progress, err = h.wizardSvc.SaveQuickAssessment(r.Context(), orgID, data)

	case 7:
		// Step 7 is completion — use the POST /onboard/complete endpoint
		writeError(w, http.StatusBadRequest, "USE_COMPLETE_ENDPOINT", "Use POST /onboard/complete for the final step")
		return

	default:
		writeError(w, http.StatusBadRequest, "INVALID_STEP", "Invalid step number")
		return
	}

	if err != nil {
		writeError(w, http.StatusInternalServerError, "STEP_SAVE_FAILED", "Failed to save step "+stepStr+": "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    progress,
	})
}

// SkipStep marks a step as skipped without saving data.
// POST /api/v1/onboard/step/{n}/skip
func (h *OnboardingHandler) SkipStep(w http.ResponseWriter, r *http.Request) {
	if h.wizardSvc == nil {
		writeError(w, http.StatusNotImplemented, "NOT_IMPLEMENTED", "Onboarding wizard not configured")
		return
	}

	orgID := middleware.GetOrgID(r.Context())
	if orgID.String() == "00000000-0000-0000-0000-000000000000" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Organisation context required")
		return
	}

	stepStr := chi.URLParam(r, "n")
	step, err := strconv.Atoi(stepStr)
	if err != nil || step < 1 || step > 7 {
		writeError(w, http.StatusBadRequest, "INVALID_STEP", "Step must be a number between 1 and 7")
		return
	}

	progress, err := h.wizardSvc.SkipStep(r.Context(), orgID, step)
	if err != nil {
		writeError(w, http.StatusBadRequest, "SKIP_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    progress,
	})
}

// CompleteOnboarding finalises the onboarding wizard.
// POST /api/v1/onboard/complete
func (h *OnboardingHandler) CompleteOnboarding(w http.ResponseWriter, r *http.Request) {
	if h.wizardSvc == nil {
		writeError(w, http.StatusNotImplemented, "NOT_IMPLEMENTED", "Onboarding wizard not configured")
		return
	}

	orgID := middleware.GetOrgID(r.Context())
	if orgID.String() == "00000000-0000-0000-0000-000000000000" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Organisation context required")
		return
	}

	result, err := h.wizardSvc.CompleteOnboarding(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "COMPLETION_FAILED", "Failed to complete onboarding: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    result,
	})
}

// GetRecommendations returns framework recommendations based on industry assessment.
// GET /api/v1/onboard/recommendations
// Also accepts POST with IndustryAssessmentInput body for ad-hoc queries.
func (h *OnboardingHandler) GetRecommendations(w http.ResponseWriter, r *http.Request) {
	if h.wizardSvc == nil {
		writeError(w, http.StatusNotImplemented, "NOT_IMPLEMENTED", "Onboarding wizard not configured")
		return
	}

	orgID := middleware.GetOrgID(r.Context())

	var answers service.IndustryAssessmentInput

	if r.Method == http.MethodPost {
		// Accept answers from request body
		if err := json.NewDecoder(r.Body).Decode(&answers); err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid assessment answers")
			return
		}
	} else {
		// GET: try to load from saved progress
		if orgID.String() == "00000000-0000-0000-0000-000000000000" {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Organisation context required for GET request")
			return
		}
		progress, err := h.wizardSvc.GetProgress(r.Context(), orgID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get progress: "+err.Error())
			return
		}
		// Convert saved data to IndustryAssessmentInput
		if progress.IndustryAssessmentData != nil {
			dataBytes, err := json.Marshal(progress.IndustryAssessmentData)
			if err == nil {
				json.Unmarshal(dataBytes, &answers)
			}
		}
	}

	recommendations, err := h.wizardSvc.GetFrameworkRecommendations(r.Context(), answers)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "RECOMMENDATION_FAILED", "Failed to generate recommendations: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    recommendations,
	})
}
