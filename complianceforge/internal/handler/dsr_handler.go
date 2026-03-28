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

// DSRHandler handles HTTP requests for GDPR Data Subject Request management.
type DSRHandler struct {
	svc *service.DSRService
}

// NewDSRHandler creates a new DSRHandler.
func NewDSRHandler(svc *service.DSRService) *DSRHandler {
	return &DSRHandler{svc: svc}
}

// ============================================================
// LIST / GET REQUESTS
// ============================================================

// ListRequests returns paginated DSR requests with optional filters.
// GET /api/v1/dsr?page=1&page_size=20&status=in_progress&request_type=access&sla_status=at_risk
func (h *DSRHandler) ListRequests(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	params := parsePagination(r)

	status := r.URL.Query().Get("status")
	requestType := r.URL.Query().Get("request_type")
	slaStatus := r.URL.Query().Get("sla_status")

	dsrs, total, err := h.svc.ListRequests(r.Context(), orgID, params.Page, params.PageSize, status, requestType, slaStatus)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve DSR requests")
		return
	}

	if dsrs == nil {
		dsrs = []service.DSRRequest{}
	}

	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize != 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: dsrs,
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

// GetRequest returns a single DSR by ID with tasks and audit trail.
// GET /api/v1/dsr/{id}
func (h *DSRHandler) GetRequest(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	dsrID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid DSR request ID")
		return
	}

	dsr, err := h.svc.GetRequest(r.Context(), orgID, dsrID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "DSR request not found")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: dsr})
}

// ============================================================
// CREATE / UPDATE REQUEST
// ============================================================

// CreateRequest creates a new DSR with auto-generated task checklist.
// POST /api/v1/dsr
func (h *DSRHandler) CreateRequest(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())

	var input service.CreateDSRRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if input.RequestType == "" || input.DataSubjectName == "" || input.DataSubjectEmail == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "request_type, data_subject_name, and data_subject_email are required")
		return
	}

	dsr, err := h.svc.CreateRequest(r.Context(), orgID, userID, input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create DSR request")
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: dsr})
}

// UpdateRequest updates mutable fields of a DSR.
// PUT /api/v1/dsr/{id}
func (h *DSRHandler) UpdateRequest(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	dsrID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid DSR request ID")
		return
	}

	var input service.UpdateDSRRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if err := h.svc.UpdateRequest(r.Context(), orgID, dsrID, input, userID); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update DSR request")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "DSR request updated successfully"},
	})
}

// ============================================================
// IDENTITY VERIFICATION
// ============================================================

// VerifyIdentity records that the data subject's identity has been verified.
// POST /api/v1/dsr/{id}/verify-identity
func (h *DSRHandler) VerifyIdentity(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	dsrID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid DSR request ID")
		return
	}

	var req struct {
		Method string `json:"method"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Method == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "method is required")
		return
	}

	if err := h.svc.VerifyIdentity(r.Context(), orgID, dsrID, req.Method, userID); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to verify identity")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Data subject identity verified"},
	})
}

// ============================================================
// ASSIGNMENT
// ============================================================

// AssignRequest assigns a DSR to a user.
// POST /api/v1/dsr/{id}/assign
func (h *DSRHandler) AssignRequest(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	dsrID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid DSR request ID")
		return
	}

	var req struct {
		AssigneeID uuid.UUID `json:"assignee_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.AssigneeID == uuid.Nil {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "assignee_id is required")
		return
	}

	if err := h.svc.AssignRequest(r.Context(), orgID, dsrID, req.AssigneeID, userID); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to assign request")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"assigned_to": req.AssigneeID,
			"message":     "DSR request assigned successfully",
		},
	})
}

// ============================================================
// EXTEND DEADLINE
// ============================================================

// ExtendDeadline extends the DSR response deadline by 60 days per GDPR Article 12(3).
// POST /api/v1/dsr/{id}/extend
func (h *DSRHandler) ExtendDeadline(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	dsrID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid DSR request ID")
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Reason == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "reason is required for deadline extension")
		return
	}

	if err := h.svc.ExtendDeadline(r.Context(), orgID, dsrID, req.Reason, userID); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to extend deadline")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"extended": true,
			"reason":   req.Reason,
			"message":  "DSR response deadline extended by 60 days",
		},
	})
}

// ============================================================
// COMPLETE / REJECT REQUEST
// ============================================================

// CompleteRequest marks a DSR as completed.
// POST /api/v1/dsr/{id}/complete
func (h *DSRHandler) CompleteRequest(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	dsrID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid DSR request ID")
		return
	}

	var req struct {
		ResponseMethod string `json:"response_method"`
		DocumentPath   string `json:"document_path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}
	if req.ResponseMethod == "" {
		req.ResponseMethod = "email"
	}

	if err := h.svc.CompleteRequest(r.Context(), orgID, dsrID, req.ResponseMethod, req.DocumentPath, userID); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to complete request")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "DSR request completed successfully"},
	})
}

// RejectRequest rejects a DSR with a reason and legal basis.
// POST /api/v1/dsr/{id}/reject
func (h *DSRHandler) RejectRequest(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	dsrID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid DSR request ID")
		return
	}

	var req struct {
		Reason     string `json:"reason"`
		LegalBasis string `json:"legal_basis"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Reason == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "reason is required for rejection")
		return
	}

	if err := h.svc.RejectRequest(r.Context(), orgID, dsrID, req.Reason, req.LegalBasis, userID); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to reject request")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"rejected":    true,
			"reason":      req.Reason,
			"legal_basis": req.LegalBasis,
			"message":     "DSR request rejected",
		},
	})
}

// ============================================================
// TASKS
// ============================================================

// GetTasks returns tasks for a specific DSR.
// GET /api/v1/dsr/{id}/tasks
func (h *DSRHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	dsrID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid DSR request ID")
		return
	}

	tasks, err := h.svc.GetTasks(r.Context(), orgID, dsrID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve DSR tasks")
		return
	}

	if tasks == nil {
		tasks = []service.DSRTask{}
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: tasks})
}

// UpdateTask updates a DSR task.
// PUT /api/v1/dsr/{id}/tasks/{taskId}
func (h *DSRHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	taskID, err := uuid.Parse(chi.URLParam(r, "taskId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid task ID")
		return
	}

	var req struct {
		Status       string     `json:"status"`
		Notes        string     `json:"notes"`
		EvidencePath string     `json:"evidence_path"`
		AssignedTo   *uuid.UUID `json:"assigned_to"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Status == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "status is required")
		return
	}

	if err := h.svc.UpdateTask(r.Context(), orgID, taskID, req.Status, req.Notes, req.EvidencePath, req.AssignedTo, userID); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update task")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"task_id": taskID,
			"status":  req.Status,
			"message": "Task updated successfully",
		},
	})
}

// ============================================================
// AUDIT TRAIL
// ============================================================

// GetAuditTrail returns the audit trail for a specific DSR.
// GET /api/v1/dsr/{id}/audit-trail
func (h *DSRHandler) GetAuditTrail(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	dsrID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid DSR request ID")
		return
	}

	entries, err := h.svc.GetAuditTrail(r.Context(), orgID, dsrID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve audit trail")
		return
	}

	if entries == nil {
		entries = []service.DSRAuditEntry{}
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: entries})
}

// ============================================================
// DASHBOARD, OVERDUE & TEMPLATES
// ============================================================

// GetDashboard returns aggregated DSR compliance metrics.
// GET /api/v1/dsr/dashboard
func (h *DSRHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	dashboard, err := h.svc.GetDSRDashboard(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get DSR dashboard")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: dashboard})
}

// GetOverdue returns DSR requests that are past their deadline.
// GET /api/v1/dsr/overdue
func (h *DSRHandler) GetOverdue(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	overdue, err := h.svc.GetOverdueRequests(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get overdue requests")
		return
	}

	if overdue == nil {
		overdue = []service.DSRRequest{}
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: overdue})
}

// ListTemplates returns DSR response templates (system + org-specific).
// GET /api/v1/dsr/templates
func (h *DSRHandler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	templates, err := h.svc.GetResponseTemplates(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get response templates")
		return
	}

	if templates == nil {
		templates = []service.DSRResponseTemplate{}
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: templates})
}
