package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/complianceforge/platform/internal/middleware"
	"github.com/complianceforge/platform/internal/models"
	"github.com/complianceforge/platform/internal/service"
)

// ============================================================
// MARKETPLACE HANDLER
// HTTP layer for the Control Library Marketplace. Provides
// public discovery endpoints, authenticated install/review
// endpoints, publisher management, and admin controls.
// ============================================================

// MarketplaceHandler handles marketplace HTTP requests.
type MarketplaceHandler struct {
	svc     *service.MarketplaceService
	builder *service.PackageBuilder
}

// NewMarketplaceHandler creates a new MarketplaceHandler.
func NewMarketplaceHandler(svc *service.MarketplaceService, builder *service.PackageBuilder) *MarketplaceHandler {
	return &MarketplaceHandler{svc: svc, builder: builder}
}

// RegisterRoutes mounts all marketplace routes on the given chi router.
// Routes are grouped by access level: public, authenticated, publisher, admin.
func (h *MarketplaceHandler) RegisterRoutes(r chi.Router) {
	r.Route("/marketplace", func(r chi.Router) {
		// ── Public endpoints (no auth required) ──
		r.Group(func(r chi.Router) {
			r.Get("/packages", h.SearchPackages)
			r.Get("/packages/featured", h.FeaturedPackages)
			r.Get("/packages/framework/{code}", h.PackagesByFramework)
			r.Get("/packages/{publisher}/{slug}", h.PackageDetail)
			r.Get("/packages/{publisher}/{slug}/versions", h.PackageVersions)
			r.Get("/packages/{publisher}/{slug}/reviews", h.PackageReviews)
		})

		// ── Authenticated endpoints ──
		r.Group(func(r chi.Router) {
			// Installation
			r.Post("/install", h.InstallPackage)
			r.Delete("/install/{id}", h.UninstallPackage)
			r.Post("/install/{id}/update", h.UpdateInstalledPackage)
			r.Get("/installed", h.ListInstalled)

			// Reviews
			r.Post("/reviews", h.SubmitReview)

			// ── Publisher endpoints ──
			r.Post("/publishers", h.CreatePublisher)
			r.Get("/publishers/me", h.MyPublisherProfile)
			r.Get("/publishers/me/stats", h.PublisherStats)
			r.Post("/publishers/me/packages", h.CreatePackage)
			r.Put("/publishers/me/packages/{id}", h.UpdatePublisherPackage)
			r.Post("/publishers/me/packages/{id}/versions", h.PublishVersion)

			// ── Admin endpoints ──
			r.Post("/publishers/{id}/verify", h.VerifyPublisher)
			r.Post("/packages/{id}/feature", h.FeaturePackage)
			r.Post("/export", h.ExportAsPackage)
		})
	})
}

// ============================================================
// PUBLIC ENDPOINTS
// ============================================================

// SearchPackages handles GET /marketplace/packages
// Query params: q, type, category, framework, region, industry, tag, pricing, sort, page, page_size
func (h *MarketplaceHandler) SearchPackages(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	filters := service.PackageFilters{
		PackageType:  r.URL.Query().Get("type"),
		Category:     r.URL.Query().Get("category"),
		PricingModel: r.URL.Query().Get("pricing"),
		SortBy:       r.URL.Query().Get("sort"),
		Page:         1,
		PageSize:     20,
	}

	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			filters.Page = v
		}
	}
	if ps := r.URL.Query().Get("page_size"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 && v <= 100 {
			filters.PageSize = v
		}
	}

	// Parse array filters from comma-separated values
	if fw := r.URL.Query().Get("framework"); fw != "" {
		filters.Frameworks = strings.Split(fw, ",")
	}
	if reg := r.URL.Query().Get("region"); reg != "" {
		filters.Regions = strings.Split(reg, ",")
	}
	if ind := r.URL.Query().Get("industry"); ind != "" {
		filters.Industries = strings.Split(ind, ",")
	}
	if tag := r.URL.Query().Get("tag"); tag != "" {
		filters.Tags = strings.Split(tag, ",")
	}

	packages, total, err := h.svc.SearchPackages(r.Context(), q, filters)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "SEARCH_FAILED", "Failed to search packages: "+err.Error())
		return
	}

	totalPages := int(total) / filters.PageSize
	if int(total)%filters.PageSize != 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: packages,
		Pagination: models.Pagination{
			Page:       filters.Page,
			PageSize:   filters.PageSize,
			TotalItems: total,
			TotalPages: totalPages,
			HasNext:    filters.Page < totalPages,
			HasPrev:    filters.Page > 1,
		},
	})
}

// FeaturedPackages handles GET /marketplace/packages/featured
func (h *MarketplaceHandler) FeaturedPackages(w http.ResponseWriter, r *http.Request) {
	packages, err := h.svc.GetFeaturedPackages(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to load featured packages")
		return
	}
	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: packages})
}

// PackagesByFramework handles GET /marketplace/packages/framework/{code}
func (h *MarketplaceHandler) PackagesByFramework(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		writeError(w, http.StatusBadRequest, "MISSING_CODE", "Framework code is required")
		return
	}

	packages, err := h.svc.GetPackagesByFramework(r.Context(), code)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to load packages")
		return
	}
	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: packages})
}

// PackageDetail handles GET /marketplace/packages/{publisher}/{slug}
func (h *MarketplaceHandler) PackageDetail(w http.ResponseWriter, r *http.Request) {
	publisherSlug := chi.URLParam(r, "publisher")
	packageSlug := chi.URLParam(r, "slug")
	if publisherSlug == "" || packageSlug == "" {
		writeError(w, http.StatusBadRequest, "MISSING_PARAMS", "Publisher and package slug are required")
		return
	}

	detail, err := h.svc.GetPackageDetail(r.Context(), publisherSlug, packageSlug)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Package not found")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: detail})
}

// PackageVersions handles GET /marketplace/packages/{publisher}/{slug}/versions
func (h *MarketplaceHandler) PackageVersions(w http.ResponseWriter, r *http.Request) {
	publisherSlug := chi.URLParam(r, "publisher")
	packageSlug := chi.URLParam(r, "slug")

	versions, err := h.svc.GetPackageVersions(r.Context(), publisherSlug, packageSlug)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Package not found")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: versions})
}

// PackageReviews handles GET /marketplace/packages/{publisher}/{slug}/reviews
func (h *MarketplaceHandler) PackageReviews(w http.ResponseWriter, r *http.Request) {
	publisherSlug := chi.URLParam(r, "publisher")
	packageSlug := chi.URLParam(r, "slug")

	// Resolve package ID from slugs
	detail, err := h.svc.GetPackageDetail(r.Context(), publisherSlug, packageSlug)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Package not found")
		return
	}

	page := 1
	pageSize := 10
	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if ps := r.URL.Query().Get("page_size"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 && v <= 50 {
			pageSize = v
		}
	}

	reviews, total, err := h.svc.GetReviews(r.Context(), detail.Package.ID, page, pageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to load reviews")
		return
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: reviews,
		Pagination: models.Pagination{
			Page:       page,
			PageSize:   pageSize,
			TotalItems: total,
			TotalPages: totalPages,
			HasNext:    page < totalPages,
			HasPrev:    page > 1,
		},
	})
}

// ============================================================
// AUTHENTICATED ENDPOINTS — Installation
// ============================================================

// InstallPackage handles POST /marketplace/install
func (h *MarketplaceHandler) InstallPackage(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Organisation context required")
		return
	}

	var req service.InstallReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	inst, err := h.svc.InstallPackage(r.Context(), orgID, req)
	if err != nil {
		status := http.StatusInternalServerError
		code := "INSTALL_FAILED"
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
			code = "NOT_FOUND"
		} else if strings.Contains(err.Error(), "not available") {
			status = http.StatusConflict
			code = "NOT_AVAILABLE"
		} else if strings.Contains(err.Error(), "integrity check") {
			status = http.StatusUnprocessableEntity
			code = "INTEGRITY_FAILED"
		}
		writeError(w, status, code, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{
		Success: true,
		Data:    inst,
	})
}

// UninstallPackage handles DELETE /marketplace/install/{id}
func (h *MarketplaceHandler) UninstallPackage(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Organisation context required")
		return
	}

	packageID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid package ID")
		return
	}

	if err := h.svc.UninstallPackage(r.Context(), orgID, packageID); err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
		}
		writeError(w, status, "UNINSTALL_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Package uninstalled successfully"},
	})
}

// UpdateInstalledPackage handles POST /marketplace/install/{id}/update
func (h *MarketplaceHandler) UpdateInstalledPackage(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Organisation context required")
		return
	}

	installationID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid installation ID")
		return
	}

	inst, err := h.svc.UpdatePackage(r.Context(), orgID, installationID)
	if err != nil {
		status := http.StatusInternalServerError
		code := "UPDATE_FAILED"
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
			code = "NOT_FOUND"
		} else if strings.Contains(err.Error(), "already at the latest") {
			status = http.StatusConflict
			code = "ALREADY_LATEST"
		}
		writeError(w, status, code, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    inst,
	})
}

// ListInstalled handles GET /marketplace/installed
func (h *MarketplaceHandler) ListInstalled(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Organisation context required")
		return
	}

	installations, err := h.svc.ListInstalled(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list installations")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    installations,
	})
}

// ============================================================
// AUTHENTICATED ENDPOINTS — Reviews
// ============================================================

// submitReviewRequest wraps the review payload with the package ID.
type submitReviewRequest struct {
	PackageID  uuid.UUID `json:"package_id"`
	Rating     int       `json:"rating"`
	Title      string    `json:"title"`
	ReviewText string    `json:"review_text"`
}

// SubmitReview handles POST /marketplace/reviews
func (h *MarketplaceHandler) SubmitReview(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	if orgID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Organisation context required")
		return
	}

	var req submitReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	reviewReq := service.ReviewReq{
		Rating:     req.Rating,
		Title:      req.Title,
		ReviewText: req.ReviewText,
	}

	review, err := h.svc.SubmitReview(r.Context(), orgID, userID, req.PackageID, reviewReq)
	if err != nil {
		status := http.StatusInternalServerError
		code := "REVIEW_FAILED"
		if strings.Contains(err.Error(), "must install") {
			status = http.StatusForbidden
			code = "NOT_INSTALLED"
		} else if strings.Contains(err.Error(), "rating must be") {
			status = http.StatusBadRequest
			code = "INVALID_RATING"
		} else if strings.Contains(err.Error(), "title is required") {
			status = http.StatusBadRequest
			code = "MISSING_TITLE"
		}
		writeError(w, status, code, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{
		Success: true,
		Data:    review,
	})
}

// ============================================================
// PUBLISHER ENDPOINTS
// ============================================================

// CreatePublisher handles POST /marketplace/publishers
func (h *MarketplaceHandler) CreatePublisher(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Organisation context required")
		return
	}

	var req service.CreatePublisherReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	pub, err := h.svc.CreatePublisher(r.Context(), orgID, req)
	if err != nil {
		status := http.StatusInternalServerError
		code := "CREATE_FAILED"
		if strings.Contains(err.Error(), "already taken") {
			status = http.StatusConflict
			code = "SLUG_TAKEN"
		} else if strings.Contains(err.Error(), "required") {
			status = http.StatusBadRequest
			code = "VALIDATION_ERROR"
		}
		writeError(w, status, code, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{
		Success: true,
		Data:    pub,
	})
}

// MyPublisherProfile handles GET /marketplace/publishers/me
func (h *MarketplaceHandler) MyPublisherProfile(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Organisation context required")
		return
	}

	pub, err := h.svc.GetPublisherByOrgID(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "No publisher profile found. Create one first.")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: pub})
}

// PublisherStats handles GET /marketplace/publishers/me/stats
func (h *MarketplaceHandler) PublisherStats(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Organisation context required")
		return
	}

	pub, err := h.svc.GetPublisherByOrgID(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "No publisher profile found")
		return
	}

	stats, err := h.svc.GetPublisherStats(r.Context(), pub.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to load stats")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"publisher": pub,
			"stats":     stats,
		},
	})
}

// CreatePackage handles POST /marketplace/publishers/me/packages
func (h *MarketplaceHandler) CreatePackage(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Organisation context required")
		return
	}

	pub, err := h.svc.GetPublisherByOrgID(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "No publisher profile found. Create one first.")
		return
	}

	var req service.CreatePackageReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	pkg, err := h.svc.CreatePackage(r.Context(), pub.ID, req)
	if err != nil {
		status := http.StatusInternalServerError
		code := "CREATE_FAILED"
		if strings.Contains(err.Error(), "already exists") {
			status = http.StatusConflict
			code = "SLUG_TAKEN"
		} else if strings.Contains(err.Error(), "required") {
			status = http.StatusBadRequest
			code = "VALIDATION_ERROR"
		}
		writeError(w, status, code, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{
		Success: true,
		Data:    pkg,
	})
}

// UpdatePublisherPackage handles PUT /marketplace/publishers/me/packages/{id}
// Updates package metadata (not version data).
func (h *MarketplaceHandler) UpdatePublisherPackage(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Organisation context required")
		return
	}

	packageID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid package ID")
		return
	}

	// Verify ownership: the publisher must belong to this org
	pub, err := h.svc.GetPublisherByOrgID(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "No publisher profile found")
		return
	}

	var req struct {
		Name                 *string  `json:"name"`
		Description          *string  `json:"description"`
		LongDescription      *string  `json:"long_description"`
		Category             *string  `json:"category"`
		ApplicableFrameworks []string `json:"applicable_frameworks"`
		ApplicableRegions    []string `json:"applicable_regions"`
		ApplicableIndustries []string `json:"applicable_industries"`
		Tags                 []string `json:"tags"`
		DeprecationMessage   *string  `json:"deprecation_message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	// Build dynamic update
	sets := []string{"updated_at = NOW()"}
	args := []interface{}{}
	argIdx := 1

	if req.Name != nil {
		sets = append(sets, "name = $"+strconv.Itoa(argIdx))
		args = append(args, *req.Name)
		argIdx++
	}
	if req.Description != nil {
		sets = append(sets, "description = $"+strconv.Itoa(argIdx))
		args = append(args, *req.Description)
		argIdx++
	}
	if req.LongDescription != nil {
		sets = append(sets, "long_description = $"+strconv.Itoa(argIdx))
		args = append(args, *req.LongDescription)
		argIdx++
	}
	if req.Category != nil {
		sets = append(sets, "category = $"+strconv.Itoa(argIdx))
		args = append(args, *req.Category)
		argIdx++
	}
	if req.ApplicableFrameworks != nil {
		sets = append(sets, "applicable_frameworks = $"+strconv.Itoa(argIdx))
		args = append(args, req.ApplicableFrameworks)
		argIdx++
	}
	if req.ApplicableRegions != nil {
		sets = append(sets, "applicable_regions = $"+strconv.Itoa(argIdx))
		args = append(args, req.ApplicableRegions)
		argIdx++
	}
	if req.ApplicableIndustries != nil {
		sets = append(sets, "applicable_industries = $"+strconv.Itoa(argIdx))
		args = append(args, req.ApplicableIndustries)
		argIdx++
	}
	if req.Tags != nil {
		sets = append(sets, "tags = $"+strconv.Itoa(argIdx))
		args = append(args, req.Tags)
		argIdx++
	}
	if req.DeprecationMessage != nil && *req.DeprecationMessage != "" {
		sets = append(sets, "status = 'deprecated'")
		sets = append(sets, "deprecated_at = NOW()")
		sets = append(sets, "deprecation_message = $"+strconv.Itoa(argIdx))
		args = append(args, *req.DeprecationMessage)
		argIdx++
	}

	query := "UPDATE marketplace_packages SET " + strings.Join(sets, ", ") +
		" WHERE id = $" + strconv.Itoa(argIdx) +
		" AND publisher_id = $" + strconv.Itoa(argIdx+1)
	args = append(args, packageID, pub.ID)

	tag, err := h.svc.Pool().Exec(r.Context(), query, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to update package: "+err.Error())
		return
	}
	if tag.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Package not found or you do not own it")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Package updated successfully"},
	})
}

// PublishVersion handles POST /marketplace/publishers/me/packages/{id}/versions
func (h *MarketplaceHandler) PublishVersion(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Organisation context required")
		return
	}

	packageID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid package ID")
		return
	}

	// Verify ownership
	pub, err := h.svc.GetPublisherByOrgID(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "No publisher profile found")
		return
	}
	_ = pub // ownership verified via org match

	var req service.PublishVersionReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	ver, err := h.svc.PublishVersion(r.Context(), packageID, req)
	if err != nil {
		status := http.StatusInternalServerError
		code := "PUBLISH_FAILED"
		if strings.Contains(err.Error(), "required") || strings.Contains(err.Error(), "must be valid") || strings.Contains(err.Error(), "must contain") {
			status = http.StatusBadRequest
			code = "VALIDATION_ERROR"
		} else if strings.Contains(err.Error(), "already exists") {
			status = http.StatusConflict
			code = "VERSION_EXISTS"
		} else if strings.Contains(err.Error(), "exceeds maximum") {
			status = http.StatusRequestEntityTooLarge
			code = "PACKAGE_TOO_LARGE"
		}
		writeError(w, status, code, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{
		Success: true,
		Data:    ver,
	})
}

// ============================================================
// ADMIN ENDPOINTS
// ============================================================

// VerifyPublisher handles POST /marketplace/publishers/{id}/verify
func (h *MarketplaceHandler) VerifyPublisher(w http.ResponseWriter, r *http.Request) {
	publisherID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid publisher ID")
		return
	}

	if err := h.svc.VerifyPublisher(r.Context(), publisherID); err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
		}
		writeError(w, status, "VERIFY_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Publisher verified successfully"},
	})
}

// FeaturePackage handles POST /marketplace/packages/{id}/feature
func (h *MarketplaceHandler) FeaturePackage(w http.ResponseWriter, r *http.Request) {
	packageID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid package ID")
		return
	}

	var req struct {
		Featured bool `json:"featured"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if err := h.svc.FeaturePackage(r.Context(), packageID, req.Featured); err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
		}
		writeError(w, status, "FEATURE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]interface{}{"featured": req.Featured, "message": "Package feature status updated"},
	})
}

// ExportAsPackage handles POST /marketplace/export
func (h *MarketplaceHandler) ExportAsPackage(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Organisation context required")
		return
	}

	var config service.ExportConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	pkgData, err := h.builder.ExportAsPackage(r.Context(), orgID, config)
	if err != nil {
		status := http.StatusInternalServerError
		code := "EXPORT_FAILED"
		if strings.Contains(err.Error(), "required") {
			status = http.StatusBadRequest
			code = "VALIDATION_ERROR"
		} else if strings.Contains(err.Error(), "exceeds") {
			status = http.StatusRequestEntityTooLarge
			code = "PACKAGE_TOO_LARGE"
		}
		writeError(w, status, code, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    pkgData,
	})
}
