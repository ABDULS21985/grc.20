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

// CaCHandler handles HTTP requests for the Compliance-as-Code engine,
// including repository management, sync runs, drift detection, and
// resource mapping operations.
type CaCHandler struct {
	engine *service.CaCEngine
}

// NewCaCHandler creates a new CaCHandler with the given CaC engine.
func NewCaCHandler(engine *service.CaCEngine) *CaCHandler {
	return &CaCHandler{engine: engine}
}

// RegisterRoutes mounts all Compliance-as-Code routes on the router.
// The caller is expected to wrap these under /cac.
func (h *CaCHandler) RegisterRoutes(r chi.Router) {
	// Repository CRUD
	r.Get("/repositories", h.ListRepositories)
	r.Post("/repositories", h.CreateRepository)
	r.Put("/repositories/{id}", h.UpdateRepository)
	r.Delete("/repositories/{id}", h.DeleteRepository)
	r.Post("/repositories/{id}/sync", h.TriggerSync)
	r.Get("/repositories/{id}/status", h.GetRepoStatus)

	// Sync runs
	r.Get("/sync-runs", h.ListSyncRuns)
	r.Get("/sync-runs/{id}", h.GetSyncRun)
	r.Post("/sync-runs/{id}/approve", h.ApproveSyncRun)
	r.Post("/sync-runs/{id}/reject", h.RejectSyncRun)

	// Drift events
	r.Get("/drift", h.ListDrift)
	r.Post("/drift/{id}/resolve", h.ResolveDrift)

	// Resource mappings
	r.Get("/resource-mappings", h.ListResourceMappings)

	// YAML operations
	r.Post("/validate", h.ValidateYAML)
	r.Post("/plan", h.GeneratePlan)
	r.Post("/apply", h.ApplyChanges)
	r.Post("/export", h.ExportToYAML)
}

// ============================================================
// REPOSITORIES
// ============================================================

// ListRepositories returns a paginated list of CaC repositories.
// GET /repositories
func (h *CaCHandler) ListRepositories(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	q := r.URL.Query()

	filters := service.CaCRepoFilters{
		Provider: q.Get("provider"),
		Search:   q.Get("search"),
		Page:     1,
		PageSize: 20,
	}
	if p, err := strconv.Atoi(q.Get("page")); err == nil && p > 0 {
		filters.Page = p
	}
	if ps, err := strconv.Atoi(q.Get("page_size")); err == nil && ps > 0 && ps <= 100 {
		filters.PageSize = ps
	}
	if active := q.Get("is_active"); active == "true" {
		v := true
		filters.IsActive = &v
	} else if active == "false" {
		v := false
		filters.IsActive = &v
	}

	repos, total, err := h.engine.ListRepositories(r.Context(), orgID, filters)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list repositories")
		return
	}

	totalPages := total / filters.PageSize
	if total%filters.PageSize != 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: repos,
		Pagination: models.Pagination{
			Page:       filters.Page,
			PageSize:   filters.PageSize,
			TotalItems: int64(total),
			TotalPages: totalPages,
			HasNext:    filters.Page < totalPages,
			HasPrev:    filters.Page > 1,
		},
	})
}

// CreateRepository connects a new Git repository for compliance-as-code.
// POST /repositories
func (h *CaCHandler) CreateRepository(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())

	var req service.CreateRepoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "name is required")
		return
	}
	if req.RepoURL == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "repo_url is required")
		return
	}

	repo, err := h.engine.CreateRepository(r.Context(), orgID, userID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create repository: "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: repo})
}

// UpdateRepository updates an existing CaC repository configuration.
// PUT /repositories/{id}
func (h *CaCHandler) UpdateRepository(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	repoID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid repository ID")
		return
	}

	var req service.UpdateRepoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	repo, err := h.engine.UpdateRepository(r.Context(), orgID, repoID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update repository: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: repo})
}

// DeleteRepository deactivates a CaC repository.
// DELETE /repositories/{id}
func (h *CaCHandler) DeleteRepository(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	repoID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid repository ID")
		return
	}

	if err := h.engine.DeleteRepository(r.Context(), orgID, repoID); err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Repository not found")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Repository deactivated successfully"},
	})
}

// TriggerSync initiates a manual synchronisation for a repository.
// POST /repositories/{id}/sync
func (h *CaCHandler) TriggerSync(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	repoID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid repository ID")
		return
	}

	var req service.TriggerSyncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Body is optional for manual triggers
		req = service.TriggerSyncRequest{}
	}

	run, err := h.engine.TriggerSync(r.Context(), orgID, repoID, userID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to trigger sync: "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: run})
}

// GetRepoStatus returns the current status of a repository including
// recent sync information and resource counts.
// GET /repositories/{id}/status
func (h *CaCHandler) GetRepoStatus(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	repoID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid repository ID")
		return
	}

	repo, err := h.engine.GetRepository(r.Context(), orgID, repoID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Repository not found")
		return
	}

	// Get recent sync runs
	runs, _, err := h.engine.GetSyncRuns(r.Context(), orgID, &repoID, 1, 5)
	if err != nil {
		runs = []service.CaCSyncRun{}
	}

	status := map[string]interface{}{
		"repository":       repo,
		"recent_sync_runs": runs,
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: status})
}

// ============================================================
// SYNC RUNS
// ============================================================

// ListSyncRuns returns a paginated list of sync runs.
// GET /sync-runs?repository_id=xxx&page=1&page_size=20
func (h *CaCHandler) ListSyncRuns(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	q := r.URL.Query()

	page := 1
	pageSize := 20
	if p, err := strconv.Atoi(q.Get("page")); err == nil && p > 0 {
		page = p
	}
	if ps, err := strconv.Atoi(q.Get("page_size")); err == nil && ps > 0 && ps <= 100 {
		pageSize = ps
	}

	var repoID *uuid.UUID
	if rid := q.Get("repository_id"); rid != "" {
		if parsed, err := uuid.Parse(rid); err == nil {
			repoID = &parsed
		}
	}

	runs, total, err := h.engine.GetSyncRuns(r.Context(), orgID, repoID, page, pageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list sync runs")
		return
	}

	totalPages := total / pageSize
	if total%pageSize != 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: runs,
		Pagination: models.Pagination{
			Page:       page,
			PageSize:   pageSize,
			TotalItems: int64(total),
			TotalPages: totalPages,
			HasNext:    page < totalPages,
			HasPrev:    page > 1,
		},
	})
}

// GetSyncRun returns details of a single sync run.
// GET /sync-runs/{id}
func (h *CaCHandler) GetSyncRun(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	syncRunID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid sync run ID")
		return
	}

	run, err := h.engine.GetSyncRun(r.Context(), orgID, syncRunID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Sync run not found")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: run})
}

// ApproveSyncRun approves a sync run that is awaiting approval.
// POST /sync-runs/{id}/approve
func (h *CaCHandler) ApproveSyncRun(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	syncRunID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid sync run ID")
		return
	}

	run, err := h.engine.ApproveSyncRun(r.Context(), orgID, syncRunID, userID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "APPROVAL_FAILED", "Failed to approve sync run: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: run})
}

// RejectSyncRun rejects a sync run that is awaiting approval.
// POST /sync-runs/{id}/reject
func (h *CaCHandler) RejectSyncRun(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	syncRunID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid sync run ID")
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.Reason = ""
	}

	run, err := h.engine.RejectSyncRun(r.Context(), orgID, syncRunID, userID, req.Reason)
	if err != nil {
		writeError(w, http.StatusBadRequest, "REJECTION_FAILED", "Failed to reject sync run: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: run})
}

// ============================================================
// DRIFT
// ============================================================

// ListDrift returns a paginated list of drift events.
// GET /drift?repository_id=xxx&status=open&direction=repo_ahead&kind=Policy
func (h *CaCHandler) ListDrift(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	q := r.URL.Query()

	filters := service.CaCDriftFilters{
		Direction: q.Get("direction"),
		Status:    q.Get("status"),
		Kind:      q.Get("kind"),
		Page:      1,
		PageSize:  20,
	}
	if p, err := strconv.Atoi(q.Get("page")); err == nil && p > 0 {
		filters.Page = p
	}
	if ps, err := strconv.Atoi(q.Get("page_size")); err == nil && ps > 0 && ps <= 100 {
		filters.PageSize = ps
	}
	if rid := q.Get("repository_id"); rid != "" {
		if parsed, err := uuid.Parse(rid); err == nil {
			filters.RepositoryID = &parsed
		}
	}

	events, total, err := h.engine.ListDriftEvents(r.Context(), orgID, filters)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list drift events")
		return
	}

	totalPages := total / filters.PageSize
	if total%filters.PageSize != 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: events,
		Pagination: models.Pagination{
			Page:       filters.Page,
			PageSize:   filters.PageSize,
			TotalItems: int64(total),
			TotalPages: totalPages,
			HasNext:    filters.Page < totalPages,
			HasPrev:    filters.Page > 1,
		},
	})
}

// ResolveDrift marks a drift event as resolved.
// POST /drift/{id}/resolve
func (h *CaCHandler) ResolveDrift(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	driftID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid drift event ID")
		return
	}

	var req service.ResolveDriftRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if err := h.engine.ResolveDrift(r.Context(), orgID, driftID, userID, req); err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Drift event not found or already resolved")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Drift event resolved successfully"},
	})
}

// ============================================================
// RESOURCE MAPPINGS
// ============================================================

// ListResourceMappings returns a paginated list of resource mappings.
// GET /resource-mappings?repository_id=xxx&kind=Policy&status=synced
func (h *CaCHandler) ListResourceMappings(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	q := r.URL.Query()

	filters := service.CaCMappingFilters{
		Kind:   q.Get("kind"),
		Status: q.Get("status"),
		Search: q.Get("search"),
		Page:   1,
		PageSize: 20,
	}
	if p, err := strconv.Atoi(q.Get("page")); err == nil && p > 0 {
		filters.Page = p
	}
	if ps, err := strconv.Atoi(q.Get("page_size")); err == nil && ps > 0 && ps <= 100 {
		filters.PageSize = ps
	}
	if rid := q.Get("repository_id"); rid != "" {
		if parsed, err := uuid.Parse(rid); err == nil {
			filters.RepositoryID = &parsed
		}
	}

	mappings, total, err := h.engine.ListResourceMappings(r.Context(), orgID, filters)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list resource mappings")
		return
	}

	totalPages := total / filters.PageSize
	if total%filters.PageSize != 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: mappings,
		Pagination: models.Pagination{
			Page:       filters.Page,
			PageSize:   filters.PageSize,
			TotalItems: int64(total),
			TotalPages: totalPages,
			HasNext:    filters.Page < totalPages,
			HasPrev:    filters.Page > 1,
		},
	})
}

// ============================================================
// YAML OPERATIONS
// ============================================================

// ValidateYAML validates a YAML document against the CaC schema.
// POST /validate
func (h *CaCHandler) ValidateYAML(w http.ResponseWriter, r *http.Request) {
	var req service.ValidateYAMLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Content == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "content is required")
		return
	}

	resources, parseErrs := h.engine.ParseMultiYAML([]byte(req.Content))

	// Validate cross-resource constraints
	validationErrs := h.engine.ValidateResources(resources)

	// Combine parse errors
	var allErrors []service.ValidationError
	for _, pe := range parseErrs {
		allErrors = append(allErrors, service.ValidationError{
			Field:    "document",
			Message:  pe.Error(),
			Severity: "error",
		})
	}
	allErrors = append(allErrors, validationErrs...)

	result := map[string]interface{}{
		"valid":            len(allErrors) == 0,
		"resources_parsed": len(resources),
		"errors":           allErrors,
		"resources":        resources,
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: result})
}

// GeneratePlan creates a diff plan comparing YAML content against platform state.
// POST /plan
func (h *CaCHandler) GeneratePlan(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var req service.PlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.RepositoryID == uuid.Nil {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "repository_id is required")
		return
	}
	if req.Content == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "content is required")
		return
	}

	resources, parseErrs := h.engine.ParseMultiYAML([]byte(req.Content))
	if len(parseErrs) > 0 && len(resources) == 0 {
		writeError(w, http.StatusBadRequest, "PARSE_ERROR", "Failed to parse any YAML resources")
		return
	}

	plan, err := h.engine.DiffWithPlatform(r.Context(), orgID, req.RepositoryID, resources)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate plan: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: plan})
}

// ApplyChanges applies a previously approved sync run's diff plan.
// POST /apply
func (h *CaCHandler) ApplyChanges(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var req service.ApplyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.SyncRunID == uuid.Nil {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "sync_run_id is required")
		return
	}

	// Fetch the sync run to get the diff plan
	run, err := h.engine.GetSyncRun(r.Context(), orgID, req.SyncRunID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Sync run not found")
		return
	}

	if run.Status != "approved" && run.Status != "pending" {
		writeError(w, http.StatusBadRequest, "INVALID_STATUS", "Sync run must be approved or pending to apply")
		return
	}

	// Deserialise the diff plan from the sync run
	var plan service.DiffPlan
	if err := json.Unmarshal(run.DiffPlanJSON, &plan); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to deserialise diff plan")
		return
	}

	result, err := h.engine.ApplyChanges(r.Context(), orgID, req.SyncRunID, &plan)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to apply changes: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: result})
}

// ExportToYAML exports all platform resources as compliance-as-code YAML.
// POST /export
func (h *CaCHandler) ExportToYAML(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	resources, err := h.engine.ExportAsYAML(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to export resources")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"resource_count": len(resources),
			"resources":      resources,
		},
	})
}
