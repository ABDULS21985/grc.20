package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/complianceforge/platform/internal/middleware"
	"github.com/complianceforge/platform/internal/models"
	"github.com/complianceforge/platform/internal/service"
)

// QuestionnaireHandler handles HTTP requests for TPRM questionnaire
// and vendor assessment management.
type QuestionnaireHandler struct {
	qSvc  *service.QuestionnaireService
	vaSvc *service.VendorAssessmentService
}

// NewQuestionnaireHandler creates a new QuestionnaireHandler.
func NewQuestionnaireHandler(qSvc *service.QuestionnaireService, vaSvc *service.VendorAssessmentService) *QuestionnaireHandler {
	return &QuestionnaireHandler{qSvc: qSvc, vaSvc: vaSvc}
}

// RegisterRoutes mounts all TPRM questionnaire and vendor assessment routes.
func (h *QuestionnaireHandler) RegisterRoutes(r chi.Router) {
	// Questionnaire template management
	r.Get("/questionnaires", h.ListQuestionnaires)
	r.Post("/questionnaires", h.CreateQuestionnaire)
	r.Get("/questionnaires/{id}", h.GetQuestionnaire)
	r.Put("/questionnaires/{id}", h.UpdateQuestionnaire)
	r.Post("/questionnaires/{id}/clone", h.CloneTemplate)

	// Vendor assessment management
	r.Get("/vendor-assessments", h.ListAssessments)
	r.Post("/vendor-assessments", h.SendAssessment)
	r.Get("/vendor-assessments/dashboard", h.GetAssessmentDashboard)
	r.Get("/vendor-assessments/compare", h.CompareVendors)
	r.Get("/vendor-assessments/{id}", h.GetAssessment)
	r.Post("/vendor-assessments/{id}/review", h.ReviewAssessment)
	r.Post("/vendor-assessments/{id}/reminder", h.SendReminder)
}

// ============================================================
// QUESTIONNAIRE ENDPOINTS
// ============================================================

// ListQuestionnaires returns a filtered, paginated list of questionnaire templates.
// GET /questionnaires
func (h *QuestionnaireHandler) ListQuestionnaires(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	q := r.URL.Query()

	filter := service.QuestionnaireFilter{
		QuestionnaireType: q.Get("type"),
		Status:            q.Get("status"),
		Search:            q.Get("search"),
		Page:              1,
		PageSize:          20,
	}
	if p, err := strconv.Atoi(q.Get("page")); err == nil && p > 0 {
		filter.Page = p
	}
	if ps, err := strconv.Atoi(q.Get("page_size")); err == nil && ps > 0 && ps <= 100 {
		filter.PageSize = ps
	}
	if q.Get("is_template") == "true" {
		filter.IsTemplate = true
	}

	list, total, err := h.qSvc.ListQuestionnaires(r.Context(), orgID, filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list questionnaires")
		return
	}

	totalPages := int(total) / filter.PageSize
	if int(total)%filter.PageSize != 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: list,
		Pagination: models.Pagination{
			Page:       filter.Page,
			PageSize:   filter.PageSize,
			TotalItems: total,
			TotalPages: totalPages,
			HasNext:    filter.Page < totalPages,
			HasPrev:    filter.Page > 1,
		},
	})
}

// CreateQuestionnaire creates a new questionnaire template.
// POST /questionnaires
func (h *QuestionnaireHandler) CreateQuestionnaire(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var req service.CreateQuestionnaireRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "Questionnaire name is required")
		return
	}

	q, err := h.qSvc.CreateQuestionnaire(r.Context(), orgID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create questionnaire: "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: q})
}

// GetQuestionnaire returns a questionnaire with sections and questions.
// GET /questionnaires/{id}
func (h *QuestionnaireHandler) GetQuestionnaire(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	qID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid questionnaire ID")
		return
	}

	q, err := h.qSvc.GetQuestionnaire(r.Context(), orgID, qID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Questionnaire not found")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: q})
}

// UpdateQuestionnaire updates a questionnaire template's metadata.
// PUT /questionnaires/{id}
func (h *QuestionnaireHandler) UpdateQuestionnaire(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	qID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid questionnaire ID")
		return
	}

	var req service.UpdateQuestionnaireRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if err := h.qSvc.UpdateQuestionnaire(r.Context(), orgID, qID, req); err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
			return
		}
		if strings.Contains(err.Error(), "cannot modify") {
			writeError(w, http.StatusForbidden, "FORBIDDEN", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update questionnaire")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Questionnaire updated successfully"},
	})
}

// CloneTemplate creates an editable copy of a questionnaire template.
// POST /questionnaires/{id}/clone
func (h *QuestionnaireHandler) CloneTemplate(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	templateID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid template ID")
		return
	}

	clone, err := h.qSvc.CloneTemplate(r.Context(), orgID, templateID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to clone template: "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: clone})
}

// ============================================================
// VENDOR ASSESSMENT ENDPOINTS
// ============================================================

// ListAssessments returns a filtered, paginated list of vendor assessments.
// GET /vendor-assessments
func (h *QuestionnaireHandler) ListAssessments(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	q := r.URL.Query()

	filter := service.AssessmentFilter{
		Status:     q.Get("status"),
		PassFail:   q.Get("pass_fail"),
		RiskRating: q.Get("risk_rating"),
		Search:     q.Get("search"),
		Page:       1,
		PageSize:   20,
	}
	if p, err := strconv.Atoi(q.Get("page")); err == nil && p > 0 {
		filter.Page = p
	}
	if ps, err := strconv.Atoi(q.Get("page_size")); err == nil && ps > 0 && ps <= 100 {
		filter.PageSize = ps
	}
	if vid := q.Get("vendor_id"); vid != "" {
		if parsed, err := uuid.Parse(vid); err == nil {
			filter.VendorID = &parsed
		}
	}
	if qid := q.Get("questionnaire_id"); qid != "" {
		if parsed, err := uuid.Parse(qid); err == nil {
			filter.QuestionnaireID = &parsed
		}
	}

	list, total, err := h.vaSvc.ListAssessments(r.Context(), orgID, filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list assessments")
		return
	}

	totalPages := int(total) / filter.PageSize
	if int(total)%filter.PageSize != 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: list,
		Pagination: models.Pagination{
			Page:       filter.Page,
			PageSize:   filter.PageSize,
			TotalItems: total,
			TotalPages: totalPages,
			HasNext:    filter.Page < totalPages,
			HasPrev:    filter.Page > 1,
		},
	})
}

// SendAssessment creates and sends a new assessment to a vendor.
// POST /vendor-assessments
func (h *QuestionnaireHandler) SendAssessment(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var req struct {
		VendorID        uuid.UUID `json:"vendor_id"`
		QuestionnaireID uuid.UUID `json:"questionnaire_id"`
		ContactEmail    string    `json:"contact_email"`
		DueDate         string    `json:"due_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.VendorID == uuid.Nil {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "vendor_id is required")
		return
	}
	if req.QuestionnaireID == uuid.Nil {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "questionnaire_id is required")
		return
	}
	if req.ContactEmail == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "contact_email is required")
		return
	}

	dueDate := time.Now().AddDate(0, 0, 30) // Default: 30 days from now
	if req.DueDate != "" {
		if parsed, err := time.Parse("2006-01-02", req.DueDate); err == nil {
			dueDate = parsed
		}
	}

	va, err := h.vaSvc.SendAssessment(r.Context(), orgID, req.VendorID, req.QuestionnaireID, dueDate, req.ContactEmail)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to send assessment: "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: va})
}

// GetAssessment returns a single assessment with responses.
// GET /vendor-assessments/{id}
func (h *QuestionnaireHandler) GetAssessment(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	assessmentID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid assessment ID")
		return
	}

	va, err := h.vaSvc.GetAssessment(r.Context(), orgID, assessmentID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Assessment not found")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: va})
}

// ReviewAssessment records reviewer feedback on a completed assessment.
// POST /vendor-assessments/{id}/review
func (h *QuestionnaireHandler) ReviewAssessment(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	assessmentID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid assessment ID")
		return
	}

	var req service.ReviewInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if err := h.vaSvc.ReviewAssessment(r.Context(), orgID, assessmentID, req); err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "not in") {
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to review assessment: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Assessment review completed"},
	})
}

// SendReminder sends a reminder to the vendor for a pending assessment.
// POST /vendor-assessments/{id}/reminder
func (h *QuestionnaireHandler) SendReminder(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	assessmentID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid assessment ID")
		return
	}

	if err := h.vaSvc.SendReminder(r.Context(), orgID, assessmentID); err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "not in") {
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to send reminder")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Reminder sent successfully"},
	})
}

// CompareVendors returns a side-by-side comparison of multiple vendor assessments.
// GET /vendor-assessments/compare?ids=uuid1,uuid2,uuid3
func (h *QuestionnaireHandler) CompareVendors(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	idsStr := r.URL.Query().Get("ids")
	if idsStr == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "ids query parameter is required")
		return
	}

	parts := strings.Split(idsStr, ",")
	var assessmentIDs []uuid.UUID
	for _, p := range parts {
		id, err := uuid.Parse(strings.TrimSpace(p))
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid assessment ID: "+p)
			return
		}
		assessmentIDs = append(assessmentIDs, id)
	}

	comparison, err := h.qSvc.CompareVendors(r.Context(), orgID, assessmentIDs)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to compare vendors: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: comparison})
}

// GetAssessmentDashboard returns aggregated TPRM metrics.
// GET /vendor-assessments/dashboard
func (h *QuestionnaireHandler) GetAssessmentDashboard(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	dash, err := h.vaSvc.GetAssessmentDashboard(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate assessment dashboard")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: dash})
}
