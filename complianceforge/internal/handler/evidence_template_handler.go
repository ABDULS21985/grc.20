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

// EvidenceTemplateHandler handles HTTP requests for the Evidence Template
// Library and Automated Evidence Testing module.
type EvidenceTemplateHandler struct {
	tmplSvc *service.EvidenceTemplateService
	runner  *service.EvidenceTestRunner
}

// NewEvidenceTemplateHandler creates a new EvidenceTemplateHandler.
func NewEvidenceTemplateHandler(tmplSvc *service.EvidenceTemplateService, runner *service.EvidenceTestRunner) *EvidenceTemplateHandler {
	return &EvidenceTemplateHandler{tmplSvc: tmplSvc, runner: runner}
}

// RegisterRoutes mounts all evidence template and testing routes on the router.
func (h *EvidenceTemplateHandler) RegisterRoutes(r chi.Router) {
	// Template endpoints
	r.Get("/templates", h.ListTemplates)
	r.Get("/templates/{id}", h.GetTemplate)
	r.Post("/templates", h.CreateTemplate)

	// Requirement endpoints
	r.Get("/requirements", h.ListRequirements)
	r.Post("/requirements/generate", h.GenerateEvidenceRequirements)
	r.Put("/requirements/{id}", h.UpdateRequirement)
	r.Post("/requirements/{id}/validate", h.ValidateEvidence)

	// Gap analysis and schedule
	r.Get("/gaps", h.GetEvidenceGaps)
	r.Get("/schedule", h.GetCollectionSchedule)

	// Test suite endpoints
	r.Get("/test-suites", h.ListTestSuites)
	r.Post("/test-suites", h.CreateTestSuite)
	r.Post("/test-suites/{id}/run", h.RunTestSuite)
	r.Get("/test-suites/{id}/results", h.GetTestRunResults)

	// Pre-audit check
	r.Post("/pre-audit-check", h.RunPreAuditChecks)
	r.Get("/pre-audit-check/{id}/report", h.GetPreAuditReport)
}

// ============================================================
// TEMPLATE ENDPOINTS
// ============================================================

// ListTemplates returns a filtered, paginated list of evidence templates.
// GET /templates?framework_code=ISO27001&category=document&page=1&page_size=20
func (h *EvidenceTemplateHandler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	q := r.URL.Query()

	filter := service.TemplateFilter{
		FrameworkCode:   q.Get("framework_code"),
		ControlCode:     q.Get("control_code"),
		Category:        q.Get("category"),
		Difficulty:      q.Get("difficulty"),
		AuditorPriority: q.Get("auditor_priority"),
		Search:          q.Get("search"),
		Page:            1,
		PageSize:        20,
	}

	if p, err := strconv.Atoi(q.Get("page")); err == nil && p > 0 {
		filter.Page = p
	}
	if ps, err := strconv.Atoi(q.Get("page_size")); err == nil && ps > 0 && ps <= 100 {
		filter.PageSize = ps
	}
	if is := q.Get("is_system"); is != "" {
		b := is == "true" || is == "1"
		filter.IsSystem = &b
	}

	result, err := h.tmplSvc.ListTemplates(r.Context(), orgID, filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list evidence templates")
		return
	}

	totalPages := int(result.Total) / filter.PageSize
	if int(result.Total)%filter.PageSize != 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: result.Templates,
		Pagination: models.Pagination{
			Page:       filter.Page,
			PageSize:   filter.PageSize,
			TotalItems: result.Total,
			TotalPages: totalPages,
			HasNext:    filter.Page < totalPages,
			HasPrev:    filter.Page > 1,
		},
	})
}

// GetTemplate returns a single evidence template by ID.
// GET /templates/{id}
func (h *EvidenceTemplateHandler) GetTemplate(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	templateID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid template ID")
		return
	}

	template, err := h.tmplSvc.GetTemplate(r.Context(), orgID, templateID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Evidence template not found")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: template})
}

// CreateTemplate creates a new custom evidence template.
// POST /templates
func (h *EvidenceTemplateHandler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var req service.CreateTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "name is required")
		return
	}
	if req.FrameworkControlCode == "" || req.FrameworkCode == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "framework_control_code and framework_code are required")
		return
	}

	template, err := h.tmplSvc.CreateTemplate(r.Context(), orgID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create evidence template: "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: template})
}

// ============================================================
// REQUIREMENT ENDPOINTS
// ============================================================

// ListRequirements returns a filtered, paginated list of evidence requirements.
// GET /requirements?status=pending&framework_code=ISO27001&page=1
func (h *EvidenceTemplateHandler) ListRequirements(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	q := r.URL.Query()

	filter := service.RequirementFilter{
		Status:           q.Get("status"),
		ValidationStatus: q.Get("validation_status"),
		FrameworkCode:    q.Get("framework_code"),
		Page:             1,
		PageSize:         20,
	}

	if p, err := strconv.Atoi(q.Get("page")); err == nil && p > 0 {
		filter.Page = p
	}
	if ps, err := strconv.Atoi(q.Get("page_size")); err == nil && ps > 0 && ps <= 100 {
		filter.PageSize = ps
	}
	if assignedTo := q.Get("assigned_to"); assignedTo != "" {
		if parsed, err := uuid.Parse(assignedTo); err == nil {
			filter.AssignedTo = &parsed
		}
	}
	if mandatory := q.Get("is_mandatory"); mandatory != "" {
		b := mandatory == "true" || mandatory == "1"
		filter.IsMandatory = &b
	}

	result, err := h.tmplSvc.ListRequirements(r.Context(), orgID, filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list evidence requirements")
		return
	}

	totalPages := int(result.Total) / filter.PageSize
	if int(result.Total)%filter.PageSize != 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: result.Requirements,
		Pagination: models.Pagination{
			Page:       filter.Page,
			PageSize:   filter.PageSize,
			TotalItems: result.Total,
			TotalPages: totalPages,
			HasNext:    filter.Page < totalPages,
			HasPrev:    filter.Page > 1,
		},
	})
}

// GenerateEvidenceRequirements automatically generates evidence requirements
// for a framework based on system templates.
// POST /requirements/generate
func (h *EvidenceTemplateHandler) GenerateEvidenceRequirements(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var req struct {
		FrameworkID uuid.UUID `json:"framework_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.FrameworkID == uuid.Nil {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "framework_id is required")
		return
	}

	result, err := h.tmplSvc.GenerateEvidenceRequirements(r.Context(), orgID, req.FrameworkID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate evidence requirements: "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: result})
}

// UpdateRequirement updates an evidence requirement.
// PUT /requirements/{id}
func (h *EvidenceTemplateHandler) UpdateRequirement(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	reqID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid requirement ID")
		return
	}

	var req service.UpdateRequirementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if err := h.tmplSvc.UpdateRequirement(r.Context(), orgID, reqID, req); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update evidence requirement: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Requirement updated successfully"},
	})
}

// ValidateEvidence validates evidence against the requirement's template rules.
// POST /requirements/{id}/validate
func (h *EvidenceTemplateHandler) ValidateEvidence(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	requirementID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid requirement ID")
		return
	}

	var req struct {
		EvidenceID uuid.UUID `json:"evidence_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.EvidenceID == uuid.Nil {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "evidence_id is required")
		return
	}

	result, err := h.tmplSvc.ValidateEvidence(r.Context(), orgID, requirementID, req.EvidenceID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to validate evidence: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: result})
}

// ============================================================
// GAP ANALYSIS & SCHEDULE
// ============================================================

// GetEvidenceGaps returns evidence collection gap analysis.
// GET /gaps
func (h *EvidenceTemplateHandler) GetEvidenceGaps(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	gaps, err := h.tmplSvc.GetEvidenceGaps(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to analyze evidence gaps")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: gaps})
}

// GetCollectionSchedule returns the upcoming evidence collection schedule.
// GET /schedule
func (h *EvidenceTemplateHandler) GetCollectionSchedule(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	schedule, err := h.tmplSvc.GetCollectionSchedule(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get collection schedule")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: schedule})
}

// ============================================================
// TEST SUITE ENDPOINTS
// ============================================================

// ListTestSuites returns all test suites for the organization.
// GET /test-suites
func (h *EvidenceTemplateHandler) ListTestSuites(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	suites, err := h.runner.ListTestSuites(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list test suites")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: suites})
}

// CreateTestSuite creates a new evidence test suite.
// POST /test-suites
func (h *EvidenceTemplateHandler) CreateTestSuite(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var req service.CreateTestSuiteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "name is required")
		return
	}

	suite, err := h.runner.CreateTestSuite(r.Context(), orgID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create test suite: "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: suite})
}

// RunTestSuite executes all test cases in a test suite.
// POST /test-suites/{id}/run
func (h *EvidenceTemplateHandler) RunTestSuite(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	suiteID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid test suite ID")
		return
	}

	run, err := h.runner.RunTestSuite(r.Context(), orgID, suiteID, "manual")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to run test suite: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: run})
}

// GetTestRunResults returns test run history for a test suite.
// GET /test-suites/{id}/results
func (h *EvidenceTemplateHandler) GetTestRunResults(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	suiteID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid test suite ID")
		return
	}

	runs, err := h.runner.GetTestRunResults(r.Context(), orgID, suiteID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get test results")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: runs})
}

// ============================================================
// PRE-AUDIT CHECK
// ============================================================

// RunPreAuditChecks performs a comprehensive pre-audit readiness check.
// POST /pre-audit-check
func (h *EvidenceTemplateHandler) RunPreAuditChecks(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var req struct {
		FrameworkID uuid.UUID `json:"framework_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.FrameworkID == uuid.Nil {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "framework_id is required")
		return
	}

	report, err := h.runner.RunPreAuditChecks(r.Context(), orgID, req.FrameworkID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to run pre-audit checks: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: report})
}

// GetPreAuditReport returns a previously generated pre-audit report.
// GET /pre-audit-check/{id}/report
// Note: Currently re-generates the report. A production system would cache reports.
func (h *EvidenceTemplateHandler) GetPreAuditReport(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	reportID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid report ID")
		return
	}

	// For now, return a message indicating the report ID.
	// A production implementation would retrieve cached reports from a table.
	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"report_id":      reportID,
			"organization_id": orgID,
			"message":        "Use POST /pre-audit-check to generate a fresh report. Report caching is a future enhancement.",
		},
	})
}
