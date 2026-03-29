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

// BIAHandler handles HTTP requests for Business Impact Analysis
// and Business Continuity management.
type BIAHandler struct {
	biaSvc *service.BIAService
	bcSvc  *service.ContinuityService
}

// NewBIAHandler creates a new BIAHandler with the given services.
func NewBIAHandler(biaSvc *service.BIAService, bcSvc *service.ContinuityService) *BIAHandler {
	return &BIAHandler{biaSvc: biaSvc, bcSvc: bcSvc}
}

// RegisterRoutes registers all BIA and BC routes on the given chi router.
func (h *BIAHandler) RegisterRoutes(r chi.Router) {
	// BIA endpoints
	r.Route("/bia", func(r chi.Router) {
		// Processes
		r.Get("/processes", h.ListProcesses)
		r.Post("/processes", h.CreateProcess)
		r.Get("/processes/{id}", h.GetProcess)
		r.Put("/processes/{id}", h.UpdateProcess)
		r.Post("/processes/{id}/assess", h.AssessProcess)
		r.Post("/processes/{id}/dependencies", h.MapDependencies)
		r.Get("/processes/{id}/dependencies", h.GetDependencies)

		// Graph & analysis
		r.Get("/dependency-graph", h.GetDependencyGraph)
		r.Get("/single-points-of-failure", h.IdentifySPoF)
		r.Get("/report", h.BIAReport)
	})

	// BC endpoints
	r.Route("/bc", func(r chi.Router) {
		// Scenarios
		r.Get("/scenarios", h.ListScenarios)
		r.Post("/scenarios", h.CreateScenario)
		r.Get("/scenarios/{id}", h.GetScenario)
		r.Put("/scenarios/{id}", h.UpdateScenario)

		// Plans
		r.Get("/plans", h.ListPlans)
		r.Post("/plans", h.CreatePlan)
		r.Get("/plans/{id}", h.GetPlan)
		r.Put("/plans/{id}", h.UpdatePlan)
		r.Post("/plans/{id}/approve", h.ApprovePlan)

		// Exercises
		r.Get("/exercises", h.ListExercises)
		r.Post("/exercises", h.ScheduleExercise)
		r.Put("/exercises/{id}/complete", h.CompleteExercise)

		// Dashboard
		r.Get("/dashboard", h.BCDashboard)
	})
}

// ============================================================
// BIA — PROCESS ENDPOINTS
// ============================================================

// ListProcesses returns a paginated list of business processes.
// GET /bia/processes?page=1&page_size=20
func (h *BIAHandler) ListProcesses(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	params := parsePagination(r)

	processes, total, err := h.biaSvc.ListProcesses(r.Context(), orgID, params.Page, params.PageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve business processes")
		return
	}

	if processes == nil {
		processes = []service.BusinessProcess{}
	}

	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize != 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: processes,
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

// CreateProcess creates a new business process.
// POST /bia/processes
func (h *BIAHandler) CreateProcess(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var req service.CreateProcessReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "name is required")
		return
	}

	process, err := h.biaSvc.CreateProcess(r.Context(), orgID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create business process")
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: process})
}

// GetProcess returns a single business process.
// GET /bia/processes/{id}
func (h *BIAHandler) GetProcess(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	processID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid process ID")
		return
	}

	process, err := h.biaSvc.GetProcess(r.Context(), orgID, processID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Business process not found")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: process})
}

// UpdateProcess updates a business process.
// PUT /bia/processes/{id}
func (h *BIAHandler) UpdateProcess(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	processID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid process ID")
		return
	}

	var req service.UpdateProcessReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if err := h.biaSvc.UpdateProcess(r.Context(), orgID, processID, req); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update business process")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: map[string]string{"message": "Process updated successfully"}})
}

// AssessProcess records impact assessment and recovery objectives for a process.
// POST /bia/processes/{id}/assess
func (h *BIAHandler) AssessProcess(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	processID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid process ID")
		return
	}

	var req service.AssessProcessReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if err := h.biaSvc.AssessProcess(r.Context(), orgID, processID, req); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to assess business process")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Process impact assessment recorded"},
	})
}

// MapDependencies replaces dependencies for a process.
// POST /bia/processes/{id}/dependencies
func (h *BIAHandler) MapDependencies(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	processID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid process ID")
		return
	}

	var deps []service.DependencyReq
	if err := json.NewDecoder(r.Body).Decode(&deps); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body. Expected array of dependency objects.")
		return
	}

	if err := h.biaSvc.MapDependencies(r.Context(), orgID, processID, deps); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to map dependencies")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]interface{}{"message": "Dependencies mapped", "count": len(deps)},
	})
}

// GetDependencies returns all dependencies for a specific process.
// GET /bia/processes/{id}/dependencies
func (h *BIAHandler) GetDependencies(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	processID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid process ID")
		return
	}

	deps, err := h.biaSvc.GetDependencies(r.Context(), orgID, processID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve dependencies")
		return
	}

	if deps == nil {
		deps = []service.ProcessDependency{}
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: deps})
}

// ============================================================
// BIA — ANALYSIS ENDPOINTS
// ============================================================

// GetDependencyGraph returns the full dependency graph for visualization.
// GET /bia/dependency-graph
func (h *BIAHandler) GetDependencyGraph(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	graph, err := h.biaSvc.GetDependencyGraph(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate dependency graph")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: graph})
}

// IdentifySPoF returns single points of failure affecting critical processes.
// GET /bia/single-points-of-failure
func (h *BIAHandler) IdentifySPoF(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	spofs, err := h.biaSvc.IdentifySinglePointsOfFailure(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to identify single points of failure")
		return
	}

	if spofs == nil {
		spofs = []service.SinglePointOfFailure{}
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: spofs})
}

// BIAReport generates a comprehensive BIA report.
// GET /bia/report
func (h *BIAHandler) BIAReport(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	report, err := h.biaSvc.GenerateBIAReport(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate BIA report")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: report})
}

// ============================================================
// BC — SCENARIO ENDPOINTS
// ============================================================

// ListScenarios returns a paginated list of BIA scenarios.
// GET /bc/scenarios?page=1&page_size=20
func (h *BIAHandler) ListScenarios(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	params := parsePagination(r)

	scenarios, total, err := h.bcSvc.ListScenarios(r.Context(), orgID, params.Page, params.PageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve scenarios")
		return
	}

	if scenarios == nil {
		scenarios = []service.BIAScenario{}
	}

	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize != 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: scenarios,
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

// CreateScenario creates a new BIA scenario.
// POST /bc/scenarios
func (h *BIAHandler) CreateScenario(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var req service.CreateScenarioReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Name == "" || req.ScenarioType == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "name and scenario_type are required")
		return
	}

	scenario, err := h.bcSvc.CreateScenario(r.Context(), orgID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create scenario")
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: scenario})
}

// GetScenario returns a single BIA scenario.
// GET /bc/scenarios/{id}
func (h *BIAHandler) GetScenario(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	scenarioID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid scenario ID")
		return
	}

	scenario, err := h.bcSvc.GetScenario(r.Context(), orgID, scenarioID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Scenario not found")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: scenario})
}

// UpdateScenario updates a BIA scenario.
// PUT /bc/scenarios/{id}
func (h *BIAHandler) UpdateScenario(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	scenarioID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid scenario ID")
		return
	}

	var req service.UpdateScenarioReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if err := h.bcSvc.UpdateScenario(r.Context(), orgID, scenarioID, req); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update scenario")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Scenario updated successfully"},
	})
}

// ============================================================
// BC — PLAN ENDPOINTS
// ============================================================

// ListPlans returns a paginated list of continuity plans.
// GET /bc/plans?page=1&page_size=20
func (h *BIAHandler) ListPlans(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	params := parsePagination(r)

	plans, total, err := h.bcSvc.ListPlans(r.Context(), orgID, params.Page, params.PageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve continuity plans")
		return
	}

	if plans == nil {
		plans = []service.ContinuityPlan{}
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

// CreatePlan creates a new continuity plan.
// POST /bc/plans
func (h *BIAHandler) CreatePlan(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var req service.CreatePlanReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Name == "" || req.PlanType == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "name and plan_type are required")
		return
	}

	plan, err := h.bcSvc.CreatePlan(r.Context(), orgID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create continuity plan")
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: plan})
}

// GetPlan returns a single continuity plan.
// GET /bc/plans/{id}
func (h *BIAHandler) GetPlan(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	planID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid plan ID")
		return
	}

	plan, err := h.bcSvc.GetPlan(r.Context(), orgID, planID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Continuity plan not found")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: plan})
}

// UpdatePlan updates a continuity plan.
// PUT /bc/plans/{id}
func (h *BIAHandler) UpdatePlan(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	planID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid plan ID")
		return
	}

	var req service.UpdatePlanReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if err := h.bcSvc.UpdatePlan(r.Context(), orgID, planID, req); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update continuity plan")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Plan updated successfully"},
	})
}

// ApprovePlan marks a plan as approved.
// POST /bc/plans/{id}/approve
func (h *BIAHandler) ApprovePlan(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	planID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid plan ID")
		return
	}

	if err := h.bcSvc.ApprovePlan(r.Context(), orgID, planID, userID); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to approve continuity plan")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Plan approved successfully"},
	})
}

// ============================================================
// BC — EXERCISE ENDPOINTS
// ============================================================

// ListExercises returns a paginated list of BC exercises.
// GET /bc/exercises?page=1&page_size=20
func (h *BIAHandler) ListExercises(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	params := parsePagination(r)

	exercises, total, err := h.bcSvc.ListExercises(r.Context(), orgID, params.Page, params.PageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve exercises")
		return
	}

	if exercises == nil {
		exercises = []service.BCExercise{}
	}

	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize != 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: exercises,
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

// ScheduleExercise creates a new BC exercise.
// POST /bc/exercises
func (h *BIAHandler) ScheduleExercise(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var req service.CreateExerciseReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Name == "" || req.ExerciseType == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "name and exercise_type are required")
		return
	}

	exercise, err := h.bcSvc.ScheduleExercise(r.Context(), orgID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to schedule exercise")
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: exercise})
}

// CompleteExercise records the results of a BC exercise.
// PUT /bc/exercises/{id}/complete
func (h *BIAHandler) CompleteExercise(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	exerciseID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid exercise ID")
		return
	}

	var req service.CompleteExerciseReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if err := h.bcSvc.CompleteExercise(r.Context(), orgID, exerciseID, req); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to complete exercise")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Exercise completed and results recorded"},
	})
}

// ============================================================
// BC — DASHBOARD
// ============================================================

// BCDashboard returns aggregated business continuity metrics.
// GET /bc/dashboard
func (h *BIAHandler) BCDashboard(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	dashboard, err := h.bcSvc.GetBCDashboard(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate BC dashboard")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: dashboard})
}
