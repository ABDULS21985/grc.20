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

// BrandingHandler handles HTTP requests for the Multi-Tenant White-Labelling,
// Custom Branding & Theming Engine module.
type BrandingHandler struct {
	svc *service.BrandingService
}

// NewBrandingHandler creates a new BrandingHandler.
func NewBrandingHandler(svc *service.BrandingService) *BrandingHandler {
	return &BrandingHandler{svc: svc}
}

// RegisterRoutes mounts all branding routes on the router.
func (h *BrandingHandler) RegisterRoutes(r chi.Router) {
	// Organisation branding endpoints
	r.Get("/branding", h.GetBranding)
	r.Get("/branding/css", h.GetBrandingCSS)
	r.Put("/branding", h.UpdateBranding)
	r.Post("/branding/logo", h.UploadLogo)
	r.Delete("/branding/logo/{type}", h.DeleteLogo)
	r.Post("/branding/domain/verify", h.VerifyCustomDomain)
	r.Get("/branding/domain/status", h.GetDomainStatus)
	r.Post("/branding/preview", h.PreviewBranding)

	// White-label partner endpoints (super admin)
	r.Get("/admin/partners", h.ListPartners)
	r.Post("/admin/partners", h.CreatePartner)
	r.Put("/admin/partners/{id}", h.UpdatePartner)
	r.Get("/admin/partners/{id}/tenants", h.GetPartnerTenants)
}

// ============================================================
// GET BRANDING
// ============================================================

// GetBranding returns the current organisation's branding configuration.
// GET /branding
func (h *BrandingHandler) GetBranding(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	branding, err := h.svc.GetBranding(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get branding: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: branding})
}

// ============================================================
// GET BRANDING CSS
// ============================================================

// GetBrandingCSS returns CSS custom properties for the organisation's branding.
// GET /branding/css
func (h *BrandingHandler) GetBrandingCSS(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	css, err := h.svc.GetBrandingCSS(r.Context(), orgID)
	if err != nil {
		http.Error(w, "/* error generating CSS */", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=300")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(css))
}

// ============================================================
// UPDATE BRANDING
// ============================================================

// UpdateBranding updates the organisation's branding configuration.
// PUT /branding
func (h *BrandingHandler) UpdateBranding(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var req service.UpdateBrandingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	branding, err := h.svc.UpdateBranding(r.Context(), orgID, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "UPDATE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: branding})
}

// ============================================================
// UPLOAD LOGO
// ============================================================

// UploadLogo handles logo file upload for the organisation.
// POST /branding/logo
// Multipart form: logo_type (string), file (binary)
func (h *BrandingHandler) UploadLogo(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	// Max 10 MB for the whole form
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "PARSE_ERROR", "Failed to parse multipart form: max 10MB")
		return
	}

	logoType := r.FormValue("logo_type")
	if logoType == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "logo_type is required (full, icon, dark, light, favicon, email)")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "MISSING_FILE", "file is required")
		return
	}
	defer file.Close()

	branding, err := h.svc.UploadLogo(r.Context(), orgID, logoType, header.Filename, file)
	if err != nil {
		writeError(w, http.StatusBadRequest, "UPLOAD_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: branding})
}

// ============================================================
// DELETE LOGO
// ============================================================

// DeleteLogo removes a logo for the organisation.
// DELETE /branding/logo/{type}
func (h *BrandingHandler) DeleteLogo(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	logoType := chi.URLParam(r, "type")

	branding, err := h.svc.DeleteLogo(r.Context(), orgID, logoType)
	if err != nil {
		writeError(w, http.StatusBadRequest, "DELETE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: branding})
}

// ============================================================
// VERIFY CUSTOM DOMAIN
// ============================================================

// VerifyCustomDomain initiates or checks custom domain verification.
// POST /branding/domain/verify
func (h *BrandingHandler) VerifyCustomDomain(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var req struct {
		Domain string `json:"domain"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Domain == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "domain is required")
		return
	}

	result, err := h.svc.VerifyCustomDomain(r.Context(), orgID, req.Domain)
	if err != nil {
		writeError(w, http.StatusBadRequest, "VERIFY_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: result})
}

// ============================================================
// GET DOMAIN STATUS
// ============================================================

// GetDomainStatus returns the current custom domain and SSL status.
// GET /branding/domain/status
func (h *BrandingHandler) GetDomainStatus(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	status, err := h.svc.GetDomainStatus(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get domain status")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: status})
}

// ============================================================
// PREVIEW BRANDING
// ============================================================

// PreviewBranding returns a CSS preview without saving changes.
// POST /branding/preview
func (h *BrandingHandler) PreviewBranding(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var req service.UpdateBrandingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	// Get current branding and merge preview changes
	current, err := h.svc.GetBranding(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get current branding")
		return
	}

	// Apply preview overrides
	if req.ColorPrimary != nil {
		current.ColorPrimary = *req.ColorPrimary
	}
	if req.ColorPrimaryHover != nil {
		current.ColorPrimaryHover = *req.ColorPrimaryHover
	}
	if req.ColorSecondary != nil {
		current.ColorSecondary = *req.ColorSecondary
	}
	if req.ColorSecondaryHover != nil {
		current.ColorSecondaryHover = *req.ColorSecondaryHover
	}
	if req.ColorAccent != nil {
		current.ColorAccent = *req.ColorAccent
	}
	if req.ColorBackground != nil {
		current.ColorBackground = *req.ColorBackground
	}
	if req.ColorSurface != nil {
		current.ColorSurface = *req.ColorSurface
	}
	if req.ColorTextPrimary != nil {
		current.ColorTextPrimary = *req.ColorTextPrimary
	}
	if req.ColorTextSecondary != nil {
		current.ColorTextSecondary = *req.ColorTextSecondary
	}
	if req.ColorBorder != nil {
		current.ColorBorder = *req.ColorBorder
	}
	if req.ColorSuccess != nil {
		current.ColorSuccess = *req.ColorSuccess
	}
	if req.ColorWarning != nil {
		current.ColorWarning = *req.ColorWarning
	}
	if req.ColorError != nil {
		current.ColorError = *req.ColorError
	}
	if req.ColorInfo != nil {
		current.ColorInfo = *req.ColorInfo
	}
	if req.ColorSidebarBg != nil {
		current.ColorSidebarBg = *req.ColorSidebarBg
	}
	if req.ColorSidebarText != nil {
		current.ColorSidebarText = *req.ColorSidebarText
	}
	if req.FontFamilyHeading != nil {
		current.FontFamilyHeading = *req.FontFamilyHeading
	}
	if req.FontFamilyBody != nil {
		current.FontFamilyBody = *req.FontFamilyBody
	}
	if req.FontSizeBase != nil {
		current.FontSizeBase = *req.FontSizeBase
	}
	if req.SidebarStyle != nil {
		current.SidebarStyle = *req.SidebarStyle
	}
	if req.CornerRadius != nil {
		current.CornerRadius = *req.CornerRadius
	}
	if req.Density != nil {
		current.Density = *req.Density
	}
	if req.CustomCSS != nil {
		current.CustomCSS = service.SanitizeCSS(*req.CustomCSS)
	}
	if req.ProductName != nil {
		current.ProductName = *req.ProductName
	}
	if req.Tagline != nil {
		current.Tagline = *req.Tagline
	}

	css := service.GenerateBrandingCSS(current)

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"css":      css,
			"branding": current,
		},
	})
}

// ============================================================
// LIST PARTNERS
// ============================================================

// ListPartners returns all white-label partners (super admin only).
// GET /admin/partners
func (h *BrandingHandler) ListPartners(w http.ResponseWriter, r *http.Request) {
	partners, err := h.svc.ListPartners(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list partners: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: partners})
}

// ============================================================
// CREATE PARTNER
// ============================================================

// CreatePartner creates a new white-label partner (super admin only).
// POST /admin/partners
func (h *BrandingHandler) CreatePartner(w http.ResponseWriter, r *http.Request) {
	var req service.CreatePartnerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.PartnerName == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "partner_name is required")
		return
	}
	if req.PartnerSlug == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "partner_slug is required")
		return
	}
	if req.ContactEmail == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "contact_email is required")
		return
	}

	partner, err := h.svc.CreatePartner(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "CREATE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: partner})
}

// ============================================================
// UPDATE PARTNER
// ============================================================

// UpdatePartner updates a white-label partner (super admin only).
// PUT /admin/partners/{id}
func (h *BrandingHandler) UpdatePartner(w http.ResponseWriter, r *http.Request) {
	partnerID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid partner ID")
		return
	}

	var req service.UpdatePartnerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	partner, err := h.svc.UpdatePartner(r.Context(), partnerID, req)
	if err != nil {
		if err.Error() == "partner not found" {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Partner not found")
			return
		}
		writeError(w, http.StatusBadRequest, "UPDATE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: partner})
}

// ============================================================
// GET PARTNER TENANTS
// ============================================================

// GetPartnerTenants returns all tenants mapped to a partner (super admin only).
// GET /admin/partners/{id}/tenants
func (h *BrandingHandler) GetPartnerTenants(w http.ResponseWriter, r *http.Request) {
	partnerID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid partner ID")
		return
	}

	tenants, err := h.svc.GetPartnerTenants(r.Context(), partnerID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get partner tenants: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: tenants})
}
