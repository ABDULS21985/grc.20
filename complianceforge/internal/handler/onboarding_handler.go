package handler

import (
	"encoding/json"
	"net/http"

	"github.com/complianceforge/platform/internal/models"
	"github.com/complianceforge/platform/internal/service"
)

type OnboardingHandler struct {
	onboardingSvc *service.OnboardingService
}

func NewOnboardingHandler(onboardingSvc *service.OnboardingService) *OnboardingHandler {
	return &OnboardingHandler{onboardingSvc: onboardingSvc}
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
