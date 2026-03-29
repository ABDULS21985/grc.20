package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/complianceforge/platform/internal/models"
	"github.com/complianceforge/platform/internal/service"
)

// VendorPortalHandler handles PUBLIC (unauthenticated) endpoints for the
// vendor self-service assessment portal. Vendors access these via a
// unique token link sent by email.
type VendorPortalHandler struct {
	vaSvc *service.VendorAssessmentService
	qSvc  *service.QuestionnaireService
}

// NewVendorPortalHandler creates a new VendorPortalHandler.
func NewVendorPortalHandler(vaSvc *service.VendorAssessmentService, qSvc *service.QuestionnaireService) *VendorPortalHandler {
	return &VendorPortalHandler{vaSvc: vaSvc, qSvc: qSvc}
}

// RegisterRoutes mounts all vendor portal routes on the given router.
// These routes are public — no auth middleware should be applied.
func (h *VendorPortalHandler) RegisterRoutes(r chi.Router) {
	r.Get("/{token}", h.GetAssessmentByToken)
	r.Put("/{token}/responses", h.SaveResponses)
	r.Post("/{token}/submit", h.SubmitAssessment)
	r.Get("/{token}/progress", h.GetProgress)
}

// ============================================================
// GET ASSESSMENT BY TOKEN
// ============================================================

// GetAssessmentByToken returns the assessment details and questions for the
// vendor to complete. The token is validated by computing its SHA-256 hash
// and matching against the stored hash.
//
// GET /{token}
func (h *VendorPortalHandler) GetAssessmentByToken(w http.ResponseWriter, r *http.Request) {
	rawToken := chi.URLParam(r, "token")
	if rawToken == "" || len(rawToken) < 32 {
		writeError(w, http.StatusBadRequest, "INVALID_TOKEN", "Invalid or missing access token")
		return
	}

	tokenHash := service.HashToken(rawToken)

	assessment, err := h.vaSvc.GetAssessmentByToken(r.Context(), tokenHash)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "NOT_FOUND",
				"Assessment not found. The link may have expired or already been completed.")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve assessment")
		return
	}

	// Fetch questionnaire with sections and questions
	questionnaire, err := h.qSvc.GetQuestionnaire(r.Context(), assessment.OrganizationID, assessment.QuestionnaireID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to load questionnaire")
		return
	}

	// Get existing responses
	progress, err := h.vaSvc.GetProgress(r.Context(), tokenHash)
	if err != nil {
		// Non-fatal: continue without progress
		progress = &service.ProgressResult{}
	}

	// Build response
	resp := map[string]interface{}{
		"assessment": map[string]interface{}{
			"id":               assessment.ID,
			"assessment_ref":   assessment.AssessmentRef,
			"status":           assessment.Status,
			"vendor_name":      assessment.VendorName,
			"questionnaire_name": assessment.QuestionnaireName,
			"sent_to_email":    assessment.SentToEmail,
			"due_date":         assessment.DueDate,
			"submitted_at":     assessment.SubmittedAt,
		},
		"questionnaire": map[string]interface{}{
			"id":                          questionnaire.ID,
			"name":                        questionnaire.Name,
			"description":                 questionnaire.Description,
			"total_questions":             questionnaire.TotalQuestions,
			"total_sections":              questionnaire.TotalSections,
			"estimated_completion_minutes": questionnaire.EstimatedCompletionMinutes,
			"sections":                    questionnaire.Sections,
		},
		"progress": progress,
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: resp})
}

// ============================================================
// SAVE RESPONSES
// ============================================================

// SaveResponses saves or updates vendor responses for questions.
// Responses can be saved incrementally (auto-save).
//
// PUT /{token}/responses
func (h *VendorPortalHandler) SaveResponses(w http.ResponseWriter, r *http.Request) {
	rawToken := chi.URLParam(r, "token")
	if rawToken == "" || len(rawToken) < 32 {
		writeError(w, http.StatusBadRequest, "INVALID_TOKEN", "Invalid or missing access token")
		return
	}

	tokenHash := service.HashToken(rawToken)

	var req struct {
		Responses []service.ResponseInput `json:"responses"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}
	if len(req.Responses) == 0 {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "At least one response is required")
		return
	}

	if err := h.vaSvc.SaveResponses(r.Context(), tokenHash, req.Responses); err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "not accessible") {
			writeError(w, http.StatusNotFound, "NOT_FOUND",
				"Assessment not found or no longer accepting responses.")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to save responses: "+err.Error())
		return
	}

	// Return updated progress
	progress, err := h.vaSvc.GetProgress(r.Context(), tokenHash)
	if err != nil {
		progress = &service.ProgressResult{}
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"message":  "Responses saved successfully",
			"saved":    len(req.Responses),
			"progress": progress,
		},
	})
}

// ============================================================
// SUBMIT ASSESSMENT
// ============================================================

// SubmitAssessment marks the assessment as completed by the vendor.
// All required questions must be answered before submission.
//
// POST /{token}/submit
func (h *VendorPortalHandler) SubmitAssessment(w http.ResponseWriter, r *http.Request) {
	rawToken := chi.URLParam(r, "token")
	if rawToken == "" || len(rawToken) < 32 {
		writeError(w, http.StatusBadRequest, "INVALID_TOKEN", "Invalid or missing access token")
		return
	}

	tokenHash := service.HashToken(rawToken)

	if err := h.vaSvc.SubmitAssessment(r.Context(), tokenHash); err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "already submitted") {
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
			return
		}
		if strings.Contains(err.Error(), "required questions") {
			writeError(w, http.StatusUnprocessableEntity, "INCOMPLETE",
				err.Error()+". Please complete all required questions before submitting.")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to submit assessment")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"message": "Assessment submitted successfully. Thank you for completing the assessment. The requesting organisation will review your responses.",
			"status":  "submitted",
		},
	})
}

// ============================================================
// GET PROGRESS
// ============================================================

// GetProgress returns the completion progress of the assessment.
//
// GET /{token}/progress
func (h *VendorPortalHandler) GetProgress(w http.ResponseWriter, r *http.Request) {
	rawToken := chi.URLParam(r, "token")
	if rawToken == "" || len(rawToken) < 32 {
		writeError(w, http.StatusBadRequest, "INVALID_TOKEN", "Invalid or missing access token")
		return
	}

	tokenHash := service.HashToken(rawToken)

	progress, err := h.vaSvc.GetProgress(r.Context(), tokenHash)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Assessment not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get progress")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: progress})
}
