package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/complianceforge/platform/internal/middleware"
	"github.com/complianceforge/platform/internal/models"
	"github.com/complianceforge/platform/internal/repository"
)

// FrameworkHandler handles HTTP requests for compliance frameworks.
type FrameworkHandler struct {
	repo *repository.FrameworkRepo
}

func NewFrameworkHandler(repo *repository.FrameworkRepo) *FrameworkHandler {
	return &FrameworkHandler{repo: repo}
}

// ListFrameworks returns all available system frameworks.
// GET /api/v1/frameworks
func (h *FrameworkHandler) ListFrameworks(w http.ResponseWriter, r *http.Request) {
	frameworks, err := h.repo.ListSystemFrameworks(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve frameworks")
		return
	}
	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: frameworks})
}

// GetFramework returns a single framework by ID with its domains.
// GET /api/v1/frameworks/{id}
func (h *FrameworkHandler) GetFramework(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid framework ID")
		return
	}

	framework, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Framework not found")
		return
	}

	// Load domains
	domains, err := h.repo.GetDomainsByFrameworkID(r.Context(), id)
	if err == nil {
		framework.Domains = domains
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: framework})
}

// GetFrameworkControls returns controls for a framework with pagination.
// GET /api/v1/frameworks/{id}/controls?page=1&page_size=20
func (h *FrameworkHandler) GetFrameworkControls(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid framework ID")
		return
	}

	params := parsePagination(r)
	controls, total, err := h.repo.GetControlsByFrameworkID(r.Context(), id, params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve controls")
		return
	}

	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize != 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: controls,
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

// SearchControls performs full-text search across all framework controls.
// GET /api/v1/frameworks/controls/search?q=encryption&limit=20
func (h *FrameworkHandler) SearchControls(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		writeError(w, http.StatusBadRequest, "MISSING_QUERY", "Search query 'q' is required")
		return
	}

	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	controls, err := h.repo.SearchControls(r.Context(), q, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Search failed")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: controls})
}

// GetComplianceScores returns compliance scores per adopted framework.
// GET /api/v1/compliance/scores
func (h *FrameworkHandler) GetComplianceScores(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	scores, err := h.repo.GetComplianceScores(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve scores")
		return
	}
	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: scores})
}

// ============================================================
// HELPER FUNCTIONS
// ============================================================

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, models.APIResponse{
		Success: false,
		Error: &models.APIError{
			Code:    code,
			Message: message,
		},
	})
}

func parsePagination(r *http.Request) models.PaginationParams {
	params := models.DefaultPagination()
	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			params.Page = v
		}
	}
	if ps := r.URL.Query().Get("page_size"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 && v <= 100 {
			params.PageSize = v
		}
	}
	if sb := r.URL.Query().Get("sort_by"); sb != "" {
		params.SortBy = sb
	}
	if sd := r.URL.Query().Get("sort_dir"); sd == "asc" || sd == "desc" {
		params.SortDir = sd
	}
	if s := r.URL.Query().Get("search"); s != "" {
		params.Search = s
	}
	return params
}
