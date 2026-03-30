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

// AuditProgrammeHandler handles HTTP requests for advanced audit management:
// programmes, universe, engagements, workpapers, sampling, test procedures,
// and corrective actions.
type AuditProgrammeHandler struct {
	svc *service.AuditProgrammeService
}

// NewAuditProgrammeHandler creates a new AuditProgrammeHandler.
func NewAuditProgrammeHandler(svc *service.AuditProgrammeService) *AuditProgrammeHandler {
	return &AuditProgrammeHandler{svc: svc}
}

// RegisterRoutes registers all audit programme routes on the given chi router.
// The caller is expected to wrap these under /audit.
func (h *AuditProgrammeHandler) RegisterRoutes(r chi.Router) {
	// Programmes
	r.Get("/programmes", h.ListProgrammes)
	r.Post("/programmes", h.CreateProgramme)
	r.Get("/programmes/{id}", h.GetProgramme)
	r.Post("/programmes/{id}/risk-selection", h.RiskBasedSelection)

	// Audit Universe
	r.Get("/universe", h.ListAuditUniverse)
	r.Post("/universe", h.CreateAuditableEntity)
	r.Put("/universe/{id}", h.UpdateAuditableEntity)

	// Engagements
	r.Get("/engagements", h.ListEngagements)
	r.Post("/engagements", h.CreateEngagement)
	r.Get("/engagements/{id}", h.GetEngagement)
	r.Put("/engagements/{id}/status", h.UpdateEngagementStatus)

	// Workpapers
	r.Get("/engagements/{id}/workpapers", h.ListWorkpapers)
	r.Post("/engagements/{id}/workpapers", h.CreateWorkpaper)
	r.Put("/workpapers/{id}", h.UpdateWorkpaper)
	r.Post("/workpapers/{id}/review", h.SubmitForReview)

	// Samples
	r.Post("/engagements/{id}/samples", h.GenerateSample)
	r.Get("/samples/{id}", h.GetSample)
	r.Put("/samples/{id}/items/{itemIndex}", h.RecordSampleItemResult)

	// Test Procedures
	r.Post("/engagements/{id}/test-procedures", h.CreateTestProcedure)
	r.Put("/test-procedures/{id}", h.RecordTestResult)

	// Corrective Actions
	r.Get("/corrective-actions", h.ListCorrectiveActions)
	r.Post("/corrective-actions", h.CreateCorrectiveAction)
	r.Put("/corrective-actions/{id}", h.UpdateCorrectiveAction)
	r.Post("/corrective-actions/{id}/verify", h.VerifyCorrectiveAction)
}

// ============================================================
// PROGRAMME ENDPOINTS
// ============================================================

// ListProgrammes returns a paginated list of audit programmes.
// GET /audit/programmes
func (h *AuditProgrammeHandler) ListProgrammes(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	params := parsePagination(r)

	programmes, total, err := h.svc.ListProgrammes(r.Context(), orgID, params.Page, params.PageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve audit programmes")
		return
	}

	if programmes == nil {
		programmes = []service.AuditProgramme{}
	}

	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize != 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: programmes,
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

// CreateProgramme creates a new audit programme.
// POST /audit/programmes
func (h *AuditProgrammeHandler) CreateProgramme(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var req service.AuditProgramme
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "name is required")
		return
	}

	programme, err := h.svc.CreateProgramme(r.Context(), orgID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create audit programme")
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: programme})
}

// GetProgramme returns a single audit programme.
// GET /audit/programmes/{id}
func (h *AuditProgrammeHandler) GetProgramme(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	programmeID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid programme ID")
		return
	}

	programme, err := h.svc.GetProgramme(r.Context(), orgID, programmeID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Audit programme not found")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: programme})
}

// RiskBasedSelection generates a risk-prioritised audit schedule.
// POST /audit/programmes/{id}/risk-selection
func (h *AuditProgrammeHandler) RiskBasedSelection(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	programmeID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid programme ID")
		return
	}

	var req struct {
		TotalDays int `json:"total_days"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.TotalDays <= 0 {
		writeError(w, http.StatusBadRequest, "INVALID_PARAM", "total_days must be positive")
		return
	}

	schedule, err := h.svc.RiskBasedSelection(r.Context(), orgID, programmeID, req.TotalDays)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate risk-based selection")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: schedule})
}

// ============================================================
// AUDIT UNIVERSE ENDPOINTS
// ============================================================

// ListAuditUniverse returns a paginated list of auditable entities.
// GET /audit/universe
func (h *AuditProgrammeHandler) ListAuditUniverse(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	params := parsePagination(r)

	entities, total, err := h.svc.ListAuditUniverse(r.Context(), orgID, params.Page, params.PageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve audit universe")
		return
	}

	if entities == nil {
		entities = []service.AuditableEntity{}
	}

	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize != 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: entities,
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

// CreateAuditableEntity adds a new entity to the audit universe.
// POST /audit/universe
func (h *AuditProgrammeHandler) CreateAuditableEntity(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var req service.AuditableEntity
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Name == "" || req.EntityType == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "name and entity_type are required")
		return
	}

	entity, err := h.svc.CreateAuditableEntity(r.Context(), orgID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create auditable entity")
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: entity})
}

// UpdateAuditableEntity updates an auditable entity.
// PUT /audit/universe/{id}
func (h *AuditProgrammeHandler) UpdateAuditableEntity(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	entityID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid entity ID")
		return
	}

	var req service.AuditableEntity
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if err := h.svc.UpdateAuditableEntity(r.Context(), orgID, entityID, req); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update auditable entity")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Auditable entity updated successfully"},
	})
}

// ============================================================
// ENGAGEMENT ENDPOINTS
// ============================================================

// ListEngagements returns a paginated list of audit engagements.
// GET /audit/engagements?programme_id=...
func (h *AuditProgrammeHandler) ListEngagements(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	params := parsePagination(r)

	var programmeID *uuid.UUID
	if pid := r.URL.Query().Get("programme_id"); pid != "" {
		parsed, err := uuid.Parse(pid)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_PARAM", "Invalid programme_id")
			return
		}
		programmeID = &parsed
	}

	engagements, total, err := h.svc.ListEngagements(r.Context(), orgID, programmeID, params.Page, params.PageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve engagements")
		return
	}

	if engagements == nil {
		engagements = []service.AuditEngagement{}
	}

	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize != 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: engagements,
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

// CreateEngagement creates a new audit engagement.
// POST /audit/engagements
func (h *AuditProgrammeHandler) CreateEngagement(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var req service.AuditEngagement
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Name == "" || req.ProgrammeID == uuid.Nil {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "name and programme_id are required")
		return
	}

	engagement, err := h.svc.CreateEngagement(r.Context(), orgID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create engagement")
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: engagement})
}

// GetEngagement returns a single audit engagement.
// GET /audit/engagements/{id}
func (h *AuditProgrammeHandler) GetEngagement(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	engagementID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid engagement ID")
		return
	}

	engagement, err := h.svc.GetEngagement(r.Context(), orgID, engagementID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Engagement not found")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: engagement})
}

// UpdateEngagementStatus transitions the engagement status.
// PUT /audit/engagements/{id}/status
func (h *AuditProgrammeHandler) UpdateEngagementStatus(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	engagementID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid engagement ID")
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Status == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "status is required")
		return
	}

	if err := h.svc.UpdateEngagementStatus(r.Context(), orgID, engagementID, req.Status); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_TRANSITION", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Engagement status updated"},
	})
}

// ============================================================
// WORKPAPER ENDPOINTS
// ============================================================

// ListWorkpapers returns all workpapers for an engagement.
// GET /audit/engagements/{id}/workpapers
func (h *AuditProgrammeHandler) ListWorkpapers(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	engagementID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid engagement ID")
		return
	}

	workpapers, err := h.svc.ListWorkpapers(r.Context(), orgID, engagementID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve workpapers")
		return
	}

	if workpapers == nil {
		workpapers = []service.AuditWorkpaper{}
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: workpapers})
}

// CreateWorkpaper creates a new workpaper.
// POST /audit/engagements/{id}/workpapers
func (h *AuditProgrammeHandler) CreateWorkpaper(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	engagementID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid engagement ID")
		return
	}

	userID := middleware.GetUserID(r.Context())

	var req service.AuditWorkpaper
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "title is required")
		return
	}

	req.EngagementID = engagementID
	req.PreparedBy = userID

	workpaper, err := h.svc.CreateWorkpaper(r.Context(), orgID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create workpaper")
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: workpaper})
}

// UpdateWorkpaper updates an existing workpaper.
// PUT /audit/workpapers/{id}
func (h *AuditProgrammeHandler) UpdateWorkpaper(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	workpaperID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid workpaper ID")
		return
	}

	var req service.AuditWorkpaper
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if err := h.svc.UpdateWorkpaper(r.Context(), orgID, workpaperID, req); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update workpaper")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Workpaper updated successfully"},
	})
}

// SubmitForReview submits a workpaper for review (four-eyes principle).
// POST /audit/workpapers/{id}/review
func (h *AuditProgrammeHandler) SubmitForReview(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	workpaperID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid workpaper ID")
		return
	}

	var req struct {
		ReviewerID uuid.UUID `json:"reviewer_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.ReviewerID == uuid.Nil {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "reviewer_id is required")
		return
	}

	if err := h.svc.SubmitForReview(r.Context(), orgID, workpaperID, req.ReviewerID); err != nil {
		writeError(w, http.StatusBadRequest, "FOUR_EYES_VIOLATION", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Workpaper submitted for review"},
	})
}

// ============================================================
// SAMPLE ENDPOINTS
// ============================================================

// GenerateSample creates a statistical audit sample.
// POST /audit/engagements/{id}/samples
func (h *AuditProgrammeHandler) GenerateSample(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	engagementID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid engagement ID")
		return
	}

	userID := middleware.GetUserID(r.Context())

	var config service.SampleConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if config.PopulationSize <= 0 {
		writeError(w, http.StatusBadRequest, "INVALID_PARAM", "population_size must be positive")
		return
	}

	sample, err := h.svc.GenerateSample(r.Context(), orgID, engagementID, userID, config)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate sample")
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: sample})
}

// GetSample returns a single audit sample with items.
// GET /audit/samples/{id}
func (h *AuditProgrammeHandler) GetSample(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	sampleID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid sample ID")
		return
	}

	sample, err := h.svc.GetSample(r.Context(), orgID, sampleID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Sample not found")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: sample})
}

// RecordSampleItemResult updates a single sample item's test result.
// PUT /audit/samples/{id}/items/{itemIndex}
func (h *AuditProgrammeHandler) RecordSampleItemResult(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	sampleID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid sample ID")
		return
	}

	itemIndex, err := strconv.Atoi(chi.URLParam(r, "itemIndex"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_PARAM", "Invalid item index")
		return
	}

	var req struct {
		Result string `json:"result"`
		Notes  string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Result == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "result is required")
		return
	}

	if err := h.svc.RecordSampleItemResult(r.Context(), orgID, sampleID, itemIndex, req.Result, req.Notes); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_PARAM", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Sample item result recorded"},
	})
}

// ============================================================
// TEST PROCEDURE ENDPOINTS
// ============================================================

// CreateTestProcedure creates a new test procedure.
// POST /audit/engagements/{id}/test-procedures
func (h *AuditProgrammeHandler) CreateTestProcedure(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	engagementID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid engagement ID")
		return
	}

	var req service.AuditTestProcedure
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "title is required")
		return
	}

	req.EngagementID = engagementID

	procedure, err := h.svc.CreateTestProcedure(r.Context(), orgID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create test procedure")
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: procedure})
}

// RecordTestResult records the result of a test procedure.
// PUT /audit/test-procedures/{id}
func (h *AuditProgrammeHandler) RecordTestResult(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	procedureID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid procedure ID")
		return
	}

	var req struct {
		ActualResult string `json:"actual_result"`
		Result       string `json:"result"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Result == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "result is required")
		return
	}

	if err := h.svc.RecordTestResult(r.Context(), orgID, procedureID, req.ActualResult, req.Result, userID); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to record test result")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Test result recorded"},
	})
}

// ============================================================
// CORRECTIVE ACTION ENDPOINTS
// ============================================================

// ListCorrectiveActions returns a paginated list of corrective actions.
// GET /audit/corrective-actions?status=open&priority=high
func (h *AuditProgrammeHandler) ListCorrectiveActions(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	params := parsePagination(r)

	status := r.URL.Query().Get("status")
	priority := r.URL.Query().Get("priority")

	actions, total, err := h.svc.ListCorrectiveActions(r.Context(), orgID, status, priority, params.Page, params.PageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve corrective actions")
		return
	}

	if actions == nil {
		actions = []service.AuditCorrectiveAction{}
	}

	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize != 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: actions,
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

// CreateCorrectiveAction creates a new corrective action.
// POST /audit/corrective-actions
func (h *AuditProgrammeHandler) CreateCorrectiveAction(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var req service.AuditCorrectiveAction
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Title == "" || req.PlannedAction == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "title and planned_action are required")
		return
	}

	action, err := h.svc.CreateCorrectiveAction(r.Context(), orgID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create corrective action")
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: action})
}

// UpdateCorrectiveAction updates a corrective action.
// PUT /audit/corrective-actions/{id}
func (h *AuditProgrammeHandler) UpdateCorrectiveAction(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	actionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid action ID")
		return
	}

	var req service.AuditCorrectiveAction
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if err := h.svc.UpdateCorrectiveAction(r.Context(), orgID, actionID, req); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update corrective action")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Corrective action updated successfully"},
	})
}

// VerifyCorrectiveAction verifies a corrective action (four-eyes principle).
// POST /audit/corrective-actions/{id}/verify
func (h *AuditProgrammeHandler) VerifyCorrectiveAction(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	actionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid action ID")
		return
	}

	var req struct {
		VerifierID uuid.UUID `json:"verifier_id"`
		Notes      string    `json:"notes"`
		Status     string    `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.VerifierID == uuid.Nil {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "verifier_id is required")
		return
	}

	if req.Status == "" {
		req.Status = "verified"
	}

	if err := h.svc.VerifyCorrectiveAction(r.Context(), orgID, actionID, req.VerifierID, req.Notes, req.Status); err != nil {
		writeError(w, http.StatusBadRequest, "FOUR_EYES_VIOLATION", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Corrective action verified"},
	})
}
