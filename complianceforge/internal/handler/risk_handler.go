package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/complianceforge/platform/internal/database"
	"github.com/complianceforge/platform/internal/middleware"
	"github.com/complianceforge/platform/internal/models"
	"github.com/complianceforge/platform/internal/repository"
)

// RiskHandler handles HTTP requests for risk management.
type RiskHandler struct {
	repo *repository.RiskRepo
	db   *database.DB
}

func NewRiskHandler(repo *repository.RiskRepo, db *database.DB) *RiskHandler {
	return &RiskHandler{repo: repo, db: db}
}

// ListRisks returns paginated risks for the organization.
// GET /api/v1/risks?page=1&page_size=20&sort_by=residual_risk_score&sort_dir=desc
func (h *RiskHandler) ListRisks(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	params := parsePagination(r)

	risks, total, err := h.repo.List(r.Context(), orgID, params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve risks")
		return
	}

	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize != 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: risks,
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

// GetRisk returns a single risk by ID.
// GET /api/v1/risks/{id}
func (h *RiskHandler) GetRisk(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	riskID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid risk ID")
		return
	}

	risk, err := h.repo.GetByID(r.Context(), orgID, riskID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Risk not found")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: risk})
}

// CreateRisk creates a new risk in the register.
// POST /api/v1/risks
func (h *RiskHandler) CreateRisk(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var req models.CreateRiskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	risk := &models.Risk{
		BaseModel:          models.BaseModel{OrganizationID: orgID},
		Title:              req.Title,
		Description:        req.Description,
		RiskCategoryID:     req.RiskCategoryID,
		RiskSource:         req.RiskSource,
		Status:             models.RiskStatusIdentified,
		OwnerUserID:        &req.OwnerUserID,
		InherentLikelihood: req.InherentLikelihood,
		InherentImpact:     req.InherentImpact,
		ResidualLikelihood: req.ResidualLikelihood,
		ResidualImpact:     req.ResidualImpact,
		FinancialImpactEUR: req.FinancialImpactEUR,
		RiskVelocity:       req.RiskVelocity,
		ReviewFrequency:    req.ReviewFrequency,
		Tags:               req.Tags,
		Metadata:           models.JSONB("{}"),
	}

	err := h.db.ExecWithTenant(r.Context(), orgID.String(), func(tx interface{ QueryRow(ctx interface{}, sql string, args ...interface{}) interface{} }) error {
		return nil // placeholder — actual impl uses db.BeginTx
	})
	_ = err

	// Simplified: direct creation via pool transaction
	tx, err := h.db.BeginTx(r.Context(), orgID.String())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to start transaction")
		return
	}
	defer tx.Rollback(r.Context())

	if err := h.repo.Create(r.Context(), tx, risk); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create risk")
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to commit transaction")
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: risk})
}

// GetHeatmap returns risk heatmap data for visualization.
// GET /api/v1/risks/heatmap
func (h *RiskHandler) GetHeatmap(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	entries, err := h.repo.GetHeatmapData(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate heatmap")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: entries})
}
