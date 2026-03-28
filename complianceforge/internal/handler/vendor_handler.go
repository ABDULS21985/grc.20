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

type VendorHandler struct {
	repo *repository.VendorRepo
	db   *database.DB
}

func NewVendorHandler(repo *repository.VendorRepo, db *database.DB) *VendorHandler {
	return &VendorHandler{repo: repo, db: db}
}

// ListVendors returns paginated vendors.
// GET /api/v1/vendors
func (h *VendorHandler) ListVendors(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	params := parsePagination(r)

	vendors, total, err := h.repo.List(r.Context(), orgID, params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve vendors")
		return
	}

	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize != 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: vendors,
		Pagination: models.Pagination{
			Page: params.Page, PageSize: params.PageSize,
			TotalItems: total, TotalPages: totalPages,
			HasNext: params.Page < totalPages, HasPrev: params.Page > 1,
		},
	})
}

// GetVendor returns full vendor detail.
// GET /api/v1/vendors/{id}
func (h *VendorHandler) GetVendor(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	vendorID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid vendor ID")
		return
	}

	vendor, err := h.repo.GetByID(r.Context(), orgID, vendorID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Vendor not found")
		return
	}

	// Add compliance warnings
	warnings := []string{}
	if vendor.DataProcessing && !vendor.DPAInPlace {
		warnings = append(warnings, "GDPR: Data Processing Agreement (DPA) required but not in place")
	}
	if vendor.NextAssessmentDate != nil {
		// Assessment overdue check handled by dashboard
	}

	resp := map[string]interface{}{
		"vendor":   vendor,
		"warnings": warnings,
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: resp})
}

// OnboardVendor creates a new vendor.
// POST /api/v1/vendors
func (h *VendorHandler) OnboardVendor(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var req struct {
		Name                string    `json:"name"`
		LegalName           string    `json:"legal_name"`
		Website             string    `json:"website"`
		Industry            string    `json:"industry"`
		CountryCode         string    `json:"country_code"`
		ContactName         string    `json:"contact_name"`
		ContactEmail        string    `json:"contact_email"`
		RiskTier            string    `json:"risk_tier"`
		ServiceDescription  string    `json:"service_description"`
		DataProcessing      bool      `json:"data_processing"`
		DataCategories      []string  `json:"data_categories"`
		Certifications      []string  `json:"certifications"`
		AssessmentFrequency string    `json:"assessment_frequency"`
		OwnerUserID         *uuid.UUID `json:"owner_user_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "Vendor name is required")
		return
	}

	tx, err := h.db.BeginTx(r.Context(), orgID.String())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Transaction failed")
		return
	}
	defer tx.Rollback(r.Context())

	assessFreq := req.AssessmentFrequency
	if assessFreq == "" {
		assessFreq = "annually"
	}

	vendor := &models.Vendor{
		BaseModel:           models.BaseModel{OrganizationID: orgID},
		Name:                req.Name,
		LegalName:           req.LegalName,
		Website:             req.Website,
		Industry:            req.Industry,
		CountryCode:         req.CountryCode,
		ContactName:         req.ContactName,
		ContactEmail:        req.ContactEmail,
		Status:              "pending",
		RiskTier:            req.RiskTier,
		ServiceDescription:  req.ServiceDescription,
		DataProcessing:      req.DataProcessing,
		DataCategories:      req.DataCategories,
		Certifications:      req.Certifications,
		AssessmentFrequency: assessFreq,
		OwnerUserID:         req.OwnerUserID,
		Metadata:            models.JSONB("{}"),
	}

	if err := h.repo.Create(r.Context(), tx, vendor); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to onboard vendor")
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Commit failed")
		return
	}

	// Build response with compliance requirements
	resp := map[string]interface{}{
		"vendor": vendor,
	}
	if vendor.DataProcessing {
		resp["gdpr_requirements"] = map[string]interface{}{
			"dpa_required":              true,
			"dpa_status":                "not_in_place",
			"sub_processor_management":  "required per GDPR Article 28",
			"data_categories":           vendor.DataCategories,
			"next_step":                 "Execute Data Processing Agreement before data sharing",
		}
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: resp})
}

// GetVendorStats returns vendor risk dashboard metrics.
// GET /api/v1/vendors/stats
func (h *VendorHandler) GetVendorStats(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	stats, err := h.repo.GetDashboardStats(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get vendor stats")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: stats})
}
