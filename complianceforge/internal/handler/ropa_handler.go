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

// ROPAHandler handles HTTP requests for Data Classification, Data Mapping,
// and ROPA (Record of Processing Activities) management.
type ROPAHandler struct {
	classSvc *service.DataClassificationService
	ropaSvc  *service.ROPAService
}

// NewROPAHandler creates a new ROPAHandler.
func NewROPAHandler(classSvc *service.DataClassificationService, ropaSvc *service.ROPAService) *ROPAHandler {
	return &ROPAHandler{classSvc: classSvc, ropaSvc: ropaSvc}
}

// RegisterRoutes mounts all data governance routes on the given chi router.
func (h *ROPAHandler) RegisterRoutes(r chi.Router) {
	// Classifications
	r.Get("/classifications", h.ListClassifications)
	r.Post("/classifications", h.CreateClassification)
	r.Put("/classifications/{id}", h.UpdateClassification)
	r.Delete("/classifications/{id}", h.DeleteClassification)

	// Data Categories
	r.Get("/categories", h.ListDataCategories)
	r.Post("/categories", h.CreateDataCategory)
	r.Put("/categories/{id}", h.UpdateDataCategory)
	r.Delete("/categories/{id}", h.DeleteDataCategory)

	// Processing Activities
	r.Get("/processing-activities", h.ListProcessingActivities)
	r.Post("/processing-activities", h.CreateProcessingActivity)
	r.Get("/processing-activities/{id}", h.GetProcessingActivity)
	r.Put("/processing-activities/{id}", h.UpdateProcessingActivity)

	// Data Flows
	r.Post("/processing-activities/{id}/flows", h.MapDataFlow)
	r.Get("/processing-activities/{id}/flows", h.ListDataFlows)

	// ROPA Export
	r.Post("/ropa/export", h.GenerateROPA)
	r.Get("/ropa/exports", h.ListROPAExports)
	r.Get("/ropa/dashboard", h.GetROPADashboard)

	// Analysis
	r.Get("/high-risk", h.IdentifyHighRiskProcessing)
	r.Get("/subject-map/{category}", h.DataSubjectImpactMap)
	r.Get("/transfers", h.GetTransferRegister)
}

// ============================================================
// CLASSIFICATION ENDPOINTS
// ============================================================

// ListClassifications returns all data classification levels for the organization.
// GET /classifications
func (h *ROPAHandler) ListClassifications(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	classifications, err := h.classSvc.ListClassifications(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list classifications")
		return
	}

	if classifications == nil {
		classifications = []service.DataClassification{}
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: classifications})
}

// CreateClassification creates a new data classification level.
// POST /classifications
func (h *ROPAHandler) CreateClassification(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var req service.CreateClassificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "name is required")
		return
	}

	classification, err := h.classSvc.CreateClassification(r.Context(), orgID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create classification: "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: classification})
}

// UpdateClassification updates a data classification level.
// PUT /classifications/{id}
func (h *ROPAHandler) UpdateClassification(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	classID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid classification ID")
		return
	}

	var req service.UpdateClassificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if err := h.classSvc.UpdateClassification(r.Context(), orgID, classID, req); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update classification: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Classification updated successfully"},
	})
}

// DeleteClassification deletes a data classification level.
// DELETE /classifications/{id}
func (h *ROPAHandler) DeleteClassification(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	classID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid classification ID")
		return
	}

	if err := h.classSvc.DeleteClassification(r.Context(), orgID, classID); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete classification: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Classification deleted successfully"},
	})
}

// ============================================================
// DATA CATEGORY ENDPOINTS
// ============================================================

// ListDataCategories returns data categories for the organization.
// GET /categories?type=personal_data&special_only=true&search=email
func (h *ROPAHandler) ListDataCategories(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	q := r.URL.Query()

	filter := service.CategoryFilter{
		CategoryType: q.Get("type"),
		SpecialOnly:  q.Get("special_only") == "true",
		Search:       q.Get("search"),
	}

	categories, err := h.classSvc.ListDataCategories(r.Context(), orgID, filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list data categories")
		return
	}

	if categories == nil {
		categories = []service.DataCategory{}
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: categories})
}

// CreateDataCategory creates a new data category.
// POST /categories
func (h *ROPAHandler) CreateDataCategory(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var req service.CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "name is required")
		return
	}

	category, err := h.classSvc.CreateDataCategory(r.Context(), orgID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create data category: "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: category})
}

// UpdateDataCategory updates a data category.
// PUT /categories/{id}
func (h *ROPAHandler) UpdateDataCategory(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	catID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid category ID")
		return
	}

	var req service.UpdateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if err := h.classSvc.UpdateDataCategory(r.Context(), orgID, catID, req); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update data category: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Data category updated successfully"},
	})
}

// DeleteDataCategory deletes a data category.
// DELETE /categories/{id}
func (h *ROPAHandler) DeleteDataCategory(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	catID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid category ID")
		return
	}

	if err := h.classSvc.DeleteDataCategory(r.Context(), orgID, catID); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete data category: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Data category deleted successfully"},
	})
}

// ============================================================
// PROCESSING ACTIVITY ENDPOINTS
// ============================================================

// ListProcessingActivities returns a filtered, paginated list of processing activities.
// GET /processing-activities?status=active&legal_basis=consent&department=IT&search=marketing&page=1&page_size=20
func (h *ROPAHandler) ListProcessingActivities(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	q := r.URL.Query()

	filter := service.ActivityFilter{
		Status:     q.Get("status"),
		LegalBasis: q.Get("legal_basis"),
		Department: q.Get("department"),
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

	result, err := h.ropaSvc.ListProcessingActivities(r.Context(), orgID, filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list processing activities")
		return
	}

	activities := result.Activities
	if activities == nil {
		activities = []service.ProcessingActivity{}
	}

	totalPages := int(result.Total) / filter.PageSize
	if int(result.Total)%filter.PageSize != 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: activities,
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

// CreateProcessingActivity creates a new processing activity with auto-ref.
// POST /processing-activities
func (h *ROPAHandler) CreateProcessingActivity(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var req service.CreateActivityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "name is required")
		return
	}

	activity, err := h.ropaSvc.CreateProcessingActivity(r.Context(), orgID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create processing activity: "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: activity})
}

// GetProcessingActivity returns a single processing activity.
// GET /processing-activities/{id}
func (h *ROPAHandler) GetProcessingActivity(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	actID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid processing activity ID")
		return
	}

	activity, err := h.ropaSvc.GetProcessingActivity(r.Context(), orgID, actID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Processing activity not found")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: activity})
}

// UpdateProcessingActivity updates a processing activity.
// PUT /processing-activities/{id}
func (h *ROPAHandler) UpdateProcessingActivity(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	actID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid processing activity ID")
		return
	}

	var req service.UpdateActivityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if err := h.ropaSvc.UpdateProcessingActivity(r.Context(), orgID, actID, req); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update processing activity: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Processing activity updated successfully"},
	})
}

// ============================================================
// DATA FLOW ENDPOINTS
// ============================================================

// MapDataFlow creates a new data flow entry for a processing activity.
// POST /processing-activities/{id}/flows
func (h *ROPAHandler) MapDataFlow(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	actID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid processing activity ID")
		return
	}

	var req service.DataFlowInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Name == "" || req.SourceName == "" || req.DestinationName == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "name, source_name, and destination_name are required")
		return
	}

	flow, err := h.ropaSvc.MapDataFlow(r.Context(), orgID, actID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to map data flow: "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: flow})
}

// ListDataFlows returns all data flows for a processing activity.
// GET /processing-activities/{id}/flows
func (h *ROPAHandler) ListDataFlows(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	actID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid processing activity ID")
		return
	}

	flows, err := h.ropaSvc.ListDataFlows(r.Context(), orgID, actID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list data flows")
		return
	}

	if flows == nil {
		flows = []service.DataFlowMap{}
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: flows})
}

// ============================================================
// ROPA EXPORT ENDPOINTS
// ============================================================

// GenerateROPA creates a ROPA export.
// POST /ropa/export
func (h *ROPAHandler) GenerateROPA(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var req struct {
		Format string `json:"format"`
		Reason string `json:"reason"`
		Notes  string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	export, err := h.ropaSvc.GenerateROPA(r.Context(), orgID, req.Format, req.Reason)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate ROPA: "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: export})
}

// ListROPAExports returns all ROPA export records.
// GET /ropa/exports
func (h *ROPAHandler) ListROPAExports(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	exports, err := h.ropaSvc.ListROPAExports(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list ROPA exports")
		return
	}

	if exports == nil {
		exports = []service.ROPAExport{}
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: exports})
}

// GetROPADashboard returns aggregated ROPA metrics.
// GET /ropa/dashboard
func (h *ROPAHandler) GetROPADashboard(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	dashboard, err := h.ropaSvc.GetROPADashboard(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate ROPA dashboard")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: dashboard})
}

// ============================================================
// ANALYSIS ENDPOINTS
// ============================================================

// IdentifyHighRiskProcessing returns processing activities that trigger DPIA.
// GET /high-risk
func (h *ROPAHandler) IdentifyHighRiskProcessing(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	highRisk, err := h.ropaSvc.IdentifyHighRiskProcessing(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to identify high-risk processing")
		return
	}

	if highRisk == nil {
		highRisk = []service.HighRiskActivity{}
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: highRisk})
}

// DataSubjectImpactMap returns how a data subject category is affected across activities.
// GET /subject-map/{category}
func (h *ROPAHandler) DataSubjectImpactMap(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	category := chi.URLParam(r, "category")
	if category == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "category is required")
		return
	}

	impactMap, err := h.ropaSvc.DataSubjectImpactMap(r.Context(), orgID, category)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate subject impact map")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: impactMap})
}

// GetTransferRegister returns all international transfer entries.
// GET /transfers
func (h *ROPAHandler) GetTransferRegister(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	transfers, err := h.ropaSvc.GetTransferRegister(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get transfer register")
		return
	}

	if transfers == nil {
		transfers = []service.TransferEntry{}
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: transfers})
}
