package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/complianceforge/platform/internal/middleware"
	"github.com/complianceforge/platform/internal/models"
	"github.com/complianceforge/platform/internal/service"
)

// ============================================================
// REPORT HANDLER V2 — Advanced Reporting Engine endpoints
// Extends the basic ReportHandler without modifying report_handler.go
// ============================================================

// ReportHandlerV2 handles advanced report generation, scheduling, and management.
type ReportHandlerV2 struct {
	engine *service.ReportEngine
}

// NewReportHandlerV2 creates a new advanced report handler.
func NewReportHandlerV2(engine *service.ReportEngine) *ReportHandlerV2 {
	return &ReportHandlerV2{engine: engine}
}

// ============================================================
// AD-HOC REPORT GENERATION
// ============================================================

// GenerateReport initiates ad-hoc report generation and returns a job ID.
// POST /api/v1/reports/generate
func (h *ReportHandlerV2) GenerateReport(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())

	var req service.GenerateReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.ReportType == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "report_type is required")
		return
	}
	if req.Format == "" {
		req.Format = "pdf"
	}

	// Create an ad-hoc definition for this generation request
	name := req.Name
	if name == "" {
		name = fmt.Sprintf("Ad-hoc %s report", req.ReportType)
	}
	classification := req.Classification
	if classification == "" {
		classification = "internal"
	}
	filters := req.Filters
	if filters == nil {
		filters = json.RawMessage(`{}`)
	}

	defReq := service.CreateReportDefinitionRequest{
		Name:                    name,
		Description:             "Ad-hoc generated report",
		ReportType:              req.ReportType,
		Format:                  req.Format,
		Filters:                 filters,
		Classification:          classification,
		IncludeExecutiveSummary: req.IncludeExecutiveSummary,
	}

	def, err := h.engine.CreateDefinition(r.Context(), orgID, userID, defReq)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create report definition")
		return
	}

	run, err := h.engine.GenerateReport(r.Context(), orgID, def.ID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "GENERATION_ERROR", "Report generation failed")
		return
	}

	writeJSON(w, http.StatusAccepted, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"run_id":        run.ID,
			"definition_id": def.ID,
			"status":        run.Status,
			"format":        run.Format,
			"message":       "Report generation initiated",
		},
	})
}

// ============================================================
// REPORT RUN STATUS & DOWNLOAD
// ============================================================

// GetReportStatus checks the status of a report generation job.
// GET /api/v1/reports/status/{id}
func (h *ReportHandlerV2) GetReportStatus(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	runID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid report run ID")
		return
	}

	run, err := h.engine.GetRun(r.Context(), orgID, runID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Report run not found")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: run})
}

// DownloadReport downloads a generated report file.
// GET /api/v1/reports/download/{id}
func (h *ReportHandlerV2) DownloadReport(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	runID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid report run ID")
		return
	}

	run, err := h.engine.GetRun(r.Context(), orgID, runID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Report run not found")
		return
	}

	if run.Status != "completed" {
		writeError(w, http.StatusConflict, "NOT_READY",
			"Report is not yet completed. Current status: "+run.Status)
		return
	}

	if run.FilePath == "" {
		writeError(w, http.StatusNotFound, "NO_FILE", "Report file not available")
		return
	}

	// Retrieve from storage
	reader, err := h.engine.GetFileReader(r.Context(), run.FilePath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "STORAGE_ERROR", "Failed to retrieve report file")
		return
	}
	defer reader.Close()

	// Set appropriate headers for file download
	contentType := formatToContentType(run.Format)
	fileName := fmt.Sprintf("report_%s.%s", run.ID.String()[:8], run.Format)

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fileName))
	if run.FileSizeBytes > 0 {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", run.FileSizeBytes))
	}

	io.Copy(w, reader)
}

// ============================================================
// REPORT DEFINITIONS — CRUD
// ============================================================

// ListDefinitions returns saved report definitions with pagination.
// GET /api/v1/reports/definitions
func (h *ReportHandlerV2) ListDefinitions(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	params := parsePagination(r)

	defs, total, err := h.engine.ListDefinitions(r.Context(), orgID, params.Page, params.PageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list report definitions")
		return
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

// CreateDefinition saves a new report definition.
// POST /api/v1/reports/definitions
func (h *ReportHandlerV2) CreateDefinition(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())

	var req service.CreateReportDefinitionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "name is required")
		return
	}
	if req.ReportType == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "report_type is required")
		return
	}
	if req.Format == "" {
		req.Format = "pdf"
	}

	def, err := h.engine.CreateDefinition(r.Context(), orgID, userID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create report definition")
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: def})
}

// UpdateDefinition updates an existing report definition.
// PUT /api/v1/reports/definitions/{id}
func (h *ReportHandlerV2) UpdateDefinition(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	defID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid definition ID")
		return
	}

	var req service.UpdateReportDefinitionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	def, err := h.engine.UpdateDefinition(r.Context(), orgID, defID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update report definition")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: def})
}

// DeleteDefinition removes a report definition.
// DELETE /api/v1/reports/definitions/{id}
func (h *ReportHandlerV2) DeleteDefinition(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	defID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid definition ID")
		return
	}

	if err := h.engine.DeleteDefinition(r.Context(), orgID, defID); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete report definition")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Report definition deleted successfully"},
	})
}

// GenerateFromDefinition triggers report generation from a saved definition.
// POST /api/v1/reports/definitions/{id}/generate
func (h *ReportHandlerV2) GenerateFromDefinition(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())

	defID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid definition ID")
		return
	}

	run, err := h.engine.GenerateReport(r.Context(), orgID, defID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "GENERATION_ERROR", "Report generation failed")
		return
	}

	writeJSON(w, http.StatusAccepted, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"run_id":        run.ID,
			"definition_id": defID,
			"status":        run.Status,
			"format":        run.Format,
			"message":       "Report generation initiated",
		},
	})
}

// ============================================================
// REPORT SCHEDULES — CRUD
// ============================================================

// ListSchedules returns all report schedules for the organisation.
// GET /api/v1/reports/schedules
func (h *ReportHandlerV2) ListSchedules(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	schedules, err := h.engine.ListSchedules(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list report schedules")
		return
	}

	if schedules == nil {
		schedules = []service.ReportSchedule{}
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: schedules})
}

// CreateSchedule creates a new report schedule.
// POST /api/v1/reports/schedules
func (h *ReportHandlerV2) CreateSchedule(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var req service.CreateScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.ReportDefinitionID == uuid.Nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "report_definition_id is required")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "name is required")
		return
	}
	if req.Frequency == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "frequency is required")
		return
	}

	schedule, err := h.engine.CreateSchedule(r.Context(), orgID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create report schedule")
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: schedule})
}

// UpdateSchedule updates an existing report schedule.
// PUT /api/v1/reports/schedules/{id}
func (h *ReportHandlerV2) UpdateSchedule(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	schedID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid schedule ID")
		return
	}

	var req service.UpdateScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	schedule, err := h.engine.UpdateSchedule(r.Context(), orgID, schedID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update report schedule")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: schedule})
}

// DeleteSchedule removes a report schedule.
// DELETE /api/v1/reports/schedules/{id}
func (h *ReportHandlerV2) DeleteSchedule(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	schedID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid schedule ID")
		return
	}

	if err := h.engine.DeleteSchedule(r.Context(), orgID, schedID); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete report schedule")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Report schedule deleted successfully"},
	})
}

// ============================================================
// REPORT HISTORY
// ============================================================

// ListHistory returns past report runs with pagination.
// GET /api/v1/reports/history
func (h *ReportHandlerV2) ListHistory(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	params := parsePagination(r)

	runs, total, err := h.engine.ListRuns(r.Context(), orgID, params.Page, params.PageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list report history")
		return
	}

	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize != 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: runs,
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

// ============================================================
// HELPERS
// ============================================================

func formatToContentType(format string) string {
	switch format {
	case "pdf":
		return "application/pdf"
	case "xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case "csv":
		return "text/csv"
	case "json":
		return "application/json"
	default:
		return "application/octet-stream"
	}
}
