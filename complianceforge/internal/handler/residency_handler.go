package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/complianceforge/platform/internal/middleware"
	"github.com/complianceforge/platform/internal/models"
	"github.com/complianceforge/platform/internal/service"
)

// ResidencyHandler handles HTTP requests for the Multi-Region Deployment
// & Data Residency module, including region configuration, compliance
// validation, and audit logging.
type ResidencyHandler struct {
	svc *service.DataResidencyService
}

// NewResidencyHandler creates a new ResidencyHandler.
func NewResidencyHandler(svc *service.DataResidencyService) *ResidencyHandler {
	return &ResidencyHandler{svc: svc}
}

// RegisterRoutes mounts all data residency routes on the router.
// The caller is expected to wrap these routes under /residency.
func (h *ResidencyHandler) RegisterRoutes(r chi.Router) {
	r.Get("/config", h.GetConfig)
	r.Get("/status", h.GetStatus)
	r.Get("/audit-log", h.GetAuditLog)
	r.Post("/validate-export", h.ValidateExport)
	r.Post("/validate-transfer", h.ValidateTransfer)
	r.Get("/regions", h.ListRegions)
}

// ============================================================
// GET CONFIG
// ============================================================

// GetConfig returns the data residency configuration for the current
// organisation, including region, allowed cloud regions, and compliance
// frameworks.
// GET /config
func (h *ResidencyHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	cfg, err := h.svc.GetRegionConfig(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get residency config")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: cfg})
}

// ============================================================
// GET STATUS (DASHBOARD)
// ============================================================

// GetStatus returns the data residency dashboard for the current
// organisation, including storage locations, compliance status, and
// recent blocked cross-region attempts.
// GET /status
func (h *ResidencyHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	dash, err := h.svc.GetResidencyDashboard(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get residency dashboard")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: dash})
}

// ============================================================
// GET AUDIT LOG
// ============================================================

// GetAuditLog returns a filtered, paginated audit log of data residency
// events for the current organisation.
// GET /audit-log
func (h *ResidencyHandler) GetAuditLog(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	q := r.URL.Query()

	filter := service.AuditLogFilter{
		Action:   q.Get("action"),
		Page:     1,
		PageSize: 20,
	}

	if p, err := strconv.Atoi(q.Get("page")); err == nil && p > 0 {
		filter.Page = p
	}
	if ps, err := strconv.Atoi(q.Get("page_size")); err == nil && ps > 0 && ps <= 100 {
		filter.PageSize = ps
	}

	if allowed := q.Get("allowed"); allowed == "true" {
		b := true
		filter.Allowed = &b
	} else if allowed == "false" {
		b := false
		filter.Allowed = &b
	}

	if df := q.Get("date_from"); df != "" {
		if t, err := time.Parse("2006-01-02", df); err == nil {
			filter.DateFrom = &t
		}
	}
	if dt := q.Get("date_to"); dt != "" {
		if t, err := time.Parse("2006-01-02", dt); err == nil {
			filter.DateTo = &t
		}
	}

	entries, total, err := h.svc.GetAuditLog(r.Context(), orgID, filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get audit log")
		return
	}

	totalPages := total / filter.PageSize
	if total%filter.PageSize != 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: entries,
		Pagination: models.Pagination{
			Page:       filter.Page,
			PageSize:   filter.PageSize,
			TotalItems: int64(total),
			TotalPages: totalPages,
			HasNext:    filter.Page < totalPages,
			HasPrev:    filter.Page > 1,
		},
	})
}

// ============================================================
// VALIDATE EXPORT
// ============================================================

// validateExportRequest is the request body for export validation.
type validateExportRequest struct {
	DestinationRegion string `json:"destination_region"`
}

// ValidateExport checks whether a data export to the specified destination
// region is permitted under the organisation's data residency rules.
// POST /validate-export
func (h *ResidencyHandler) ValidateExport(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var req validateExportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.DestinationRegion == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "destination_region is required")
		return
	}

	result, err := h.svc.ValidateDataExport(r.Context(), orgID, req.DestinationRegion)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to validate export: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: result})
}

// ============================================================
// VALIDATE TRANSFER
// ============================================================

// validateTransferRequest is the request body for vendor data transfer validation.
type validateTransferRequest struct {
	VendorID           uuid.UUID `json:"vendor_id"`
	DestinationCountry string    `json:"destination_country"`
}

// ValidateTransfer checks whether a vendor data transfer to the specified
// destination country is permissible under GDPR adequacy decisions and
// the organisation's residency configuration.
// POST /validate-transfer
func (h *ResidencyHandler) ValidateTransfer(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var req validateTransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.DestinationCountry == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "destination_country is required")
		return
	}

	result, err := h.svc.ValidateDataTransfer(r.Context(), orgID, req.VendorID, req.DestinationCountry)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to validate transfer: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: result})
}

// ============================================================
// LIST REGIONS
// ============================================================

// ListRegions returns all available data residency regions.
// GET /regions
func (h *ResidencyHandler) ListRegions(w http.ResponseWriter, r *http.Request) {
	regions, err := h.svc.ListRegions(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list regions")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: regions})
}
