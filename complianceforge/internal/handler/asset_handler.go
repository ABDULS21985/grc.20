package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/complianceforge/platform/internal/middleware"
	"github.com/complianceforge/platform/internal/models"
	"github.com/complianceforge/platform/internal/repository"
)

type AssetHandler struct {
	repo *repository.AssetRepo
}

func NewAssetHandler(repo *repository.AssetRepo) *AssetHandler {
	return &AssetHandler{repo: repo}
}

// ListAssets returns paginated assets.
// GET /api/v1/assets
func (h *AssetHandler) ListAssets(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	params := parsePagination(r)

	assets, total, err := h.repo.List(r.Context(), orgID, params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve assets")
		return
	}

	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize != 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: assets,
		Pagination: models.Pagination{
			Page: params.Page, PageSize: params.PageSize,
			TotalItems: total, TotalPages: totalPages,
			HasNext: params.Page < totalPages, HasPrev: params.Page > 1,
		},
	})
}

// GetAsset returns a single asset.
// GET /api/v1/assets/{id}
func (h *AssetHandler) GetAsset(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	assetID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid asset ID")
		return
	}

	asset, err := h.repo.GetByID(r.Context(), orgID, assetID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Asset not found")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: asset})
}

// RegisterAsset creates a new asset in the inventory.
// POST /api/v1/assets
func (h *AssetHandler) RegisterAsset(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var req struct {
		Name                string     `json:"name"`
		AssetType           string     `json:"asset_type"`
		Category            string     `json:"category"`
		Description         string     `json:"description"`
		Criticality         string     `json:"criticality"`
		OwnerUserID         *uuid.UUID `json:"owner_user_id"`
		Location            string     `json:"location"`
		IPAddress           string     `json:"ip_address"`
		Classification      string     `json:"classification"`
		ProcessesPersonalData bool     `json:"processes_personal_data"`
		LinkedVendorID      *uuid.UUID `json:"linked_vendor_id"`
		Tags                []string   `json:"tags"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Name == "" || req.AssetType == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "name and asset_type are required")
		return
	}

	criticality := req.Criticality
	if criticality == "" {
		criticality = "medium"
	}
	classification := req.Classification
	if classification == "" {
		classification = "internal"
	}

	asset := &models.Asset{
		BaseModel:           models.BaseModel{OrganizationID: orgID},
		Name:                req.Name,
		AssetType:           req.AssetType,
		Category:            req.Category,
		Description:         req.Description,
		Status:              "active",
		Criticality:         criticality,
		OwnerUserID:         req.OwnerUserID,
		Location:            req.Location,
		IPAddress:           req.IPAddress,
		Classification:      classification,
		ProcessesPersonalData: req.ProcessesPersonalData,
		LinkedVendorID:      req.LinkedVendorID,
		Tags:                req.Tags,
		Metadata:            models.JSONB("{}"),
	}

	if err := h.repo.Create(r.Context(), asset); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to register asset")
		return
	}

	resp := map[string]interface{}{
		"asset": asset,
	}
	if asset.ProcessesPersonalData {
		resp["gdpr_notice"] = "This asset processes personal data. Ensure it is included in your ROPA (Record of Processing Activities) per GDPR Article 30."
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: resp})
}

// GetAssetStats returns asset inventory statistics.
// GET /api/v1/assets/stats
func (h *AssetHandler) GetAssetStats(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	stats, err := h.repo.GetStats(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get asset stats")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: stats})
}
