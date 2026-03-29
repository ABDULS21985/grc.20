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

// ExceptionHandler handles HTTP requests for the Exception Management
// & Compensating Controls module.
type ExceptionHandler struct {
	svc *service.ExceptionService
}

// NewExceptionHandler creates a new ExceptionHandler.
func NewExceptionHandler(svc *service.ExceptionService) *ExceptionHandler {
	return &ExceptionHandler{svc: svc}
}

// RegisterRoutes mounts all exception management routes on the router.
func (h *ExceptionHandler) RegisterRoutes(r chi.Router) {
	r.Get("/", h.ListExceptions)
	r.Post("/", h.CreateException)
	r.Get("/dashboard", h.GetExceptionDashboard)
	r.Get("/expiring", h.GetExpiringExceptions)
	r.Get("/impact/{id}", h.CalculateComplianceImpact)
	r.Get("/{id}", h.GetException)
	r.Put("/{id}", h.UpdateException)
	r.Post("/{id}/submit", h.SubmitForApproval)
	r.Post("/{id}/approve", h.ApproveException)
	r.Post("/{id}/reject", h.RejectException)
	r.Post("/{id}/revoke", h.RevokeException)
	r.Post("/{id}/renew", h.RenewException)
	r.Post("/{id}/review", h.ReviewException)
	r.Get("/{id}/reviews", h.GetExceptionReviews)
	r.Get("/{id}/audit-trail", h.GetExceptionAuditTrail)
}

// ============================================================
// LIST
// ============================================================

// ListExceptions returns a filtered, paginated list of exceptions.
// GET /
func (h *ExceptionHandler) ListExceptions(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	q := r.URL.Query()

	filter := service.ExceptionFilter{
		Status:        q.Get("status"),
		ExceptionType: q.Get("exception_type"),
		Priority:      q.Get("priority"),
		ScopeType:     q.Get("scope_type"),
		RiskLevel:     q.Get("risk_level"),
		Search:        q.Get("search"),
		Page:          1,
		PageSize:      20,
	}

	if p, err := strconv.Atoi(q.Get("page")); err == nil && p > 0 {
		filter.Page = p
	}
	if ps, err := strconv.Atoi(q.Get("page_size")); err == nil && ps > 0 && ps <= 100 {
		filter.PageSize = ps
	}

	result, err := h.svc.ListExceptions(r.Context(), orgID, filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list exceptions: "+err.Error())
		return
	}

	totalPages := int(result.Total) / filter.PageSize
	if int(result.Total)%filter.PageSize != 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: result.Exceptions,
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

// ============================================================
// CREATE
// ============================================================

// CreateException creates a new compliance exception.
// POST /
func (h *ExceptionHandler) CreateException(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var req service.CreateExceptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "title is required")
		return
	}
	if req.RiskJustification == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "risk_justification is required")
		return
	}

	exc, err := h.svc.CreateException(r.Context(), orgID, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "CREATE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: exc})
}

// ============================================================
// GET
// ============================================================

// GetException returns a single exception.
// GET /{id}
func (h *ExceptionHandler) GetException(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	exceptionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid exception ID")
		return
	}

	exc, err := h.svc.GetException(r.Context(), orgID, exceptionID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Exception not found")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: exc})
}

// ============================================================
// UPDATE
// ============================================================

// UpdateException updates an exception (only in draft status).
// PUT /{id}
func (h *ExceptionHandler) UpdateException(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	exceptionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid exception ID")
		return
	}

	var req service.UpdateExceptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	exc, err := h.svc.UpdateException(r.Context(), orgID, exceptionID, req)
	if err != nil {
		if err.Error() == "exception not found" {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Exception not found")
			return
		}
		writeError(w, http.StatusBadRequest, "UPDATE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: exc})
}

// ============================================================
// SUBMIT FOR APPROVAL
// ============================================================

// SubmitForApproval submits a draft exception for approval.
// POST /{id}/submit
func (h *ExceptionHandler) SubmitForApproval(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	exceptionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid exception ID")
		return
	}

	if err := h.svc.SubmitForApproval(r.Context(), orgID, exceptionID); err != nil {
		if err.Error() == "exception not found" {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Exception not found")
			return
		}
		writeError(w, http.StatusBadRequest, "SUBMIT_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Exception submitted for approval"},
	})
}

// ============================================================
// APPROVE
// ============================================================

// ApproveException approves an exception.
// POST /{id}/approve
func (h *ExceptionHandler) ApproveException(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	exceptionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid exception ID")
		return
	}

	var req service.ApproveExceptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if err := h.svc.ApproveException(r.Context(), orgID, exceptionID, userID, req.Comments); err != nil {
		if err.Error() == "exception not found" {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Exception not found")
			return
		}
		writeError(w, http.StatusBadRequest, "APPROVE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Exception approved"},
	})
}

// ============================================================
// REJECT
// ============================================================

// RejectException rejects an exception.
// POST /{id}/reject
func (h *ExceptionHandler) RejectException(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	exceptionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid exception ID")
		return
	}

	var req service.RejectExceptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Reason == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "reason is required")
		return
	}

	if err := h.svc.RejectException(r.Context(), orgID, exceptionID, userID, req.Reason); err != nil {
		if err.Error() == "exception not found" {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Exception not found")
			return
		}
		writeError(w, http.StatusBadRequest, "REJECT_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Exception rejected"},
	})
}

// ============================================================
// REVOKE
// ============================================================

// RevokeException revokes an approved exception.
// POST /{id}/revoke
func (h *ExceptionHandler) RevokeException(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	exceptionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid exception ID")
		return
	}

	var req service.RevokeExceptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Reason == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "reason is required")
		return
	}

	if err := h.svc.RevokeException(r.Context(), orgID, exceptionID, userID, req.Reason); err != nil {
		if err.Error() == "exception not found" {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Exception not found")
			return
		}
		writeError(w, http.StatusBadRequest, "REVOKE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Exception revoked"},
	})
}

// ============================================================
// RENEW
// ============================================================

// RenewException renews an exception with a new expiry date.
// POST /{id}/renew
func (h *ExceptionHandler) RenewException(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	exceptionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid exception ID")
		return
	}

	var req service.RenewExceptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.NewExpiry == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "new_expiry is required")
		return
	}
	if req.Justification == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "justification is required")
		return
	}

	newExpiry, err := time.Parse("2006-01-02", req.NewExpiry)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_DATE", "Invalid new_expiry format; use YYYY-MM-DD")
		return
	}

	if err := h.svc.RenewException(r.Context(), orgID, exceptionID, newExpiry, req.Justification); err != nil {
		if err.Error() == "exception not found" {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Exception not found")
			return
		}
		writeError(w, http.StatusBadRequest, "RENEW_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Exception renewed"},
	})
}

// ============================================================
// REVIEW
// ============================================================

// ReviewException records a review of an exception.
// POST /{id}/review
func (h *ExceptionHandler) ReviewException(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	exceptionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid exception ID")
		return
	}

	var req service.ExceptionReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Outcome == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "outcome is required")
		return
	}

	if err := h.svc.ReviewException(r.Context(), orgID, exceptionID, req); err != nil {
		if err.Error() == "exception not found" {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Exception not found")
			return
		}
		writeError(w, http.StatusBadRequest, "REVIEW_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Review recorded"},
	})
}

// ============================================================
// GET REVIEWS
// ============================================================

// GetExceptionReviews returns all reviews for an exception.
// GET /{id}/reviews
func (h *ExceptionHandler) GetExceptionReviews(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	exceptionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid exception ID")
		return
	}

	reviews, err := h.svc.GetExceptionReviews(r.Context(), orgID, exceptionID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get reviews")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: reviews})
}

// ============================================================
// GET AUDIT TRAIL
// ============================================================

// GetExceptionAuditTrail returns the audit trail for an exception.
// GET /{id}/audit-trail
func (h *ExceptionHandler) GetExceptionAuditTrail(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	exceptionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid exception ID")
		return
	}

	trail, err := h.svc.GetExceptionAuditTrail(r.Context(), orgID, exceptionID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get audit trail")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: trail})
}

// ============================================================
// DASHBOARD
// ============================================================

// GetExceptionDashboard returns aggregated exception metrics.
// GET /dashboard
func (h *ExceptionHandler) GetExceptionDashboard(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	dash, err := h.svc.GetExceptionDashboard(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate exception dashboard")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: dash})
}

// ============================================================
// EXPIRING
// ============================================================

// GetExpiringExceptions returns exceptions expiring within N days.
// GET /expiring?days=30
func (h *ExceptionHandler) GetExpiringExceptions(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	days := 30
	if d, err := strconv.Atoi(r.URL.Query().Get("days")); err == nil && d > 0 && d <= 365 {
		days = d
	}

	exceptions, err := h.svc.GetExpiringExceptions(r.Context(), orgID, days)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get expiring exceptions")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: exceptions})
}

// ============================================================
// COMPLIANCE IMPACT
// ============================================================

// CalculateComplianceImpact returns the compliance impact of an exception.
// GET /impact/{id}
func (h *ExceptionHandler) CalculateComplianceImpact(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	exceptionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid exception ID")
		return
	}

	impact, err := h.svc.CalculateComplianceImpact(r.Context(), orgID, exceptionID)
	if err != nil {
		if err.Error() == "exception not found" {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Exception not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to calculate compliance impact")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: impact})
}
