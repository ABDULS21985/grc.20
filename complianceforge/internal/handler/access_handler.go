package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/complianceforge/platform/internal/middleware"
	"github.com/complianceforge/platform/internal/models"
	"github.com/complianceforge/platform/internal/service"
)

// ============================================================
// ACCESS HANDLER — ABAC policy management endpoints
// ============================================================

// AccessHandler provides HTTP endpoints for managing ABAC access policies,
// assignments, field permissions, and the access audit log.
type AccessHandler struct {
	engine  *service.ABACEngine
	masker  *service.FieldMasker
}

// NewAccessHandler creates a new AccessHandler.
func NewAccessHandler(engine *service.ABACEngine, masker *service.FieldMasker) *AccessHandler {
	return &AccessHandler{engine: engine, masker: masker}
}

// ============================================================
// REQUEST / RESPONSE TYPES
// ============================================================

// CreatePolicyRequest is the JSON body for creating an access policy.
type CreatePolicyRequest struct {
	Name                  string             `json:"name"`
	Description           string             `json:"description"`
	Priority              int                `json:"priority"`
	Effect                string             `json:"effect"`
	IsActive              bool               `json:"is_active"`
	SubjectConditions     []service.Condition `json:"subject_conditions"`
	ResourceType          string             `json:"resource_type"`
	ResourceConditions    []service.Condition `json:"resource_conditions"`
	Actions               []string           `json:"actions"`
	EnvironmentConditions []service.Condition `json:"environment_conditions"`
	ValidFrom             *time.Time         `json:"valid_from,omitempty"`
	ValidUntil            *time.Time         `json:"valid_until,omitempty"`
}

// UpdatePolicyRequest is the JSON body for updating an access policy.
type UpdatePolicyRequest struct {
	Name                  string             `json:"name"`
	Description           string             `json:"description"`
	Priority              int                `json:"priority"`
	Effect                string             `json:"effect"`
	IsActive              *bool              `json:"is_active"`
	SubjectConditions     []service.Condition `json:"subject_conditions"`
	ResourceType          string             `json:"resource_type"`
	ResourceConditions    []service.Condition `json:"resource_conditions"`
	Actions               []string           `json:"actions"`
	EnvironmentConditions []service.Condition `json:"environment_conditions"`
	ValidFrom             *time.Time         `json:"valid_from,omitempty"`
	ValidUntil            *time.Time         `json:"valid_until,omitempty"`
}

// CreateAssignmentRequest is the JSON body for creating a policy assignment.
type CreateAssignmentRequest struct {
	AssigneeType string     `json:"assignee_type"`
	AssigneeID   *uuid.UUID `json:"assignee_id,omitempty"`
	ValidFrom    *time.Time `json:"valid_from,omitempty"`
	ValidUntil   *time.Time `json:"valid_until,omitempty"`
}

// EvaluateRequest is the JSON body for the diagnostic evaluate endpoint.
type EvaluateRequest struct {
	SubjectID    uuid.UUID  `json:"subject_id"`
	Action       string     `json:"action"`
	ResourceType string     `json:"resource_type"`
	ResourceID   *uuid.UUID `json:"resource_id,omitempty"`
	Roles        []string   `json:"roles"`
	MFAVerified  bool       `json:"mfa_verified"`
	IP           string     `json:"ip"`
}

// ============================================================
// ENDPOINT: GET /access/policies — list access policies
// ============================================================

func (h *AccessHandler) ListPolicies(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	params := parsePagination(r)

	policies, total, err := h.engine.ListPolicies(r.Context(), orgID, params.Page, params.PageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve access policies")
		return
	}

	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize != 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: policies,
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

// ============================================================
// ENDPOINT: POST /access/policies — create access policy
// ============================================================

func (h *AccessHandler) CreatePolicy(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())

	var req CreatePolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "Policy name is required")
		return
	}
	if req.Effect != "allow" && req.Effect != "deny" {
		writeError(w, http.StatusBadRequest, "INVALID_EFFECT", "Effect must be 'allow' or 'deny'")
		return
	}
	if req.ResourceType == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "Resource type is required")
		return
	}
	if len(req.Actions) == 0 {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "At least one action is required")
		return
	}

	policy := &service.PolicyRecord{
		OrgID:                 orgID,
		Name:                  req.Name,
		Description:           req.Description,
		Priority:              req.Priority,
		Effect:                req.Effect,
		IsActive:              req.IsActive,
		SubjectConditions:     req.SubjectConditions,
		ResourceType:          req.ResourceType,
		ResourceConditions:    req.ResourceConditions,
		Actions:               req.Actions,
		EnvironmentConditions: req.EnvironmentConditions,
		ValidFrom:             req.ValidFrom,
		ValidUntil:            req.ValidUntil,
		CreatedBy:             &userID,
	}

	if policy.Priority == 0 {
		policy.Priority = 100
	}

	if err := h.engine.CreatePolicy(r.Context(), policy); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create access policy")
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: policy})
}

// ============================================================
// ENDPOINT: PUT /access/policies/{id} — update access policy
// ============================================================

func (h *AccessHandler) UpdatePolicy(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	policyID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid policy ID")
		return
	}

	var req UpdatePolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	existing, err := h.engine.GetPolicyByID(r.Context(), orgID, policyID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Access policy not found")
		return
	}

	// Apply updates
	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.Description != "" {
		existing.Description = req.Description
	}
	if req.Priority != 0 {
		existing.Priority = req.Priority
	}
	if req.Effect == "allow" || req.Effect == "deny" {
		existing.Effect = req.Effect
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}
	if req.SubjectConditions != nil {
		existing.SubjectConditions = req.SubjectConditions
	}
	if req.ResourceType != "" {
		existing.ResourceType = req.ResourceType
	}
	if req.ResourceConditions != nil {
		existing.ResourceConditions = req.ResourceConditions
	}
	if req.Actions != nil {
		existing.Actions = req.Actions
	}
	if req.EnvironmentConditions != nil {
		existing.EnvironmentConditions = req.EnvironmentConditions
	}
	if req.ValidFrom != nil {
		existing.ValidFrom = req.ValidFrom
	}
	if req.ValidUntil != nil {
		existing.ValidUntil = req.ValidUntil
	}

	if err := h.engine.UpdatePolicy(r.Context(), existing); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update access policy")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: existing})
}

// ============================================================
// ENDPOINT: DELETE /access/policies/{id} — delete access policy
// ============================================================

func (h *AccessHandler) DeletePolicy(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	policyID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid policy ID")
		return
	}

	if err := h.engine.DeletePolicy(r.Context(), orgID, policyID); err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Access policy not found")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"id":      policyID,
			"deleted": true,
		},
	})
}

// ============================================================
// ENDPOINT: POST /access/policies/{id}/assignments — create assignment
// ============================================================

func (h *AccessHandler) CreateAssignment(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	policyID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid policy ID")
		return
	}

	// Verify policy exists
	if _, err := h.engine.GetPolicyByID(r.Context(), orgID, policyID); err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Access policy not found")
		return
	}

	var req CreateAssignmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	validTypes := map[string]bool{"user": true, "role": true, "group": true, "all_users": true}
	if !validTypes[req.AssigneeType] {
		writeError(w, http.StatusBadRequest, "INVALID_TYPE", "Assignee type must be: user, role, group, or all_users")
		return
	}

	assignment := &service.PolicyAssignment{
		OrgID:          orgID,
		AccessPolicyID: policyID,
		AssigneeType:   req.AssigneeType,
		AssigneeID:     req.AssigneeID,
		ValidFrom:      req.ValidFrom,
		ValidUntil:     req.ValidUntil,
		CreatedBy:      &userID,
	}

	if err := h.engine.CreateAssignment(r.Context(), assignment); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create assignment")
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: assignment})
}

// ============================================================
// ENDPOINT: DELETE /access/policies/{id}/assignments/{assignmentId}
// ============================================================

func (h *AccessHandler) DeleteAssignment(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	assignmentID, err := uuid.Parse(chi.URLParam(r, "assignmentId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid assignment ID")
		return
	}

	if err := h.engine.DeleteAssignment(r.Context(), orgID, assignmentID); err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Assignment not found")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"id":      assignmentID,
			"deleted": true,
		},
	})
}

// ============================================================
// ENDPOINT: POST /access/evaluate — diagnostic evaluation
// ============================================================

func (h *AccessHandler) EvaluateAccess(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var req EvaluateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.SubjectID == uuid.Nil {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "Subject ID is required")
		return
	}
	if req.Action == "" || req.ResourceType == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "Action and resource_type are required")
		return
	}

	accessReq := service.AccessRequest{
		SubjectID:    req.SubjectID,
		OrgID:        orgID,
		Action:       req.Action,
		ResourceType: req.ResourceType,
		ResourceID:   req.ResourceID,
	}

	subject := service.SubjectAttributes{
		UserID:      req.SubjectID,
		OrgID:       orgID,
		Roles:       req.Roles,
		MFAVerified: req.MFAVerified,
	}

	env := service.EnvironmentAttributes{
		IP:        req.IP,
		Time:      time.Now().UTC(),
		MFAStatus: req.MFAVerified,
	}

	decision := h.engine.Evaluate(r.Context(), accessReq, subject, env)

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"decision":         decision.Effect,
			"policy_id":        decision.PolicyID,
			"policy_name":      decision.PolicyName,
			"reason":           decision.Reason,
			"evaluation_time_us": decision.EvaluationTimeUS,
		},
	})
}

// ============================================================
// ENDPOINT: GET /access/audit-log — access decision log
// ============================================================

func (h *AccessHandler) ListAuditLog(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	params := parsePagination(r)

	entries, total, err := h.engine.ListAuditLog(r.Context(), orgID, params.Page, params.PageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve audit log")
		return
	}

	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize != 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: entries,
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

// ============================================================
// ENDPOINT: GET /access/my-permissions — current user permissions
// ============================================================

func (h *AccessHandler) GetMyPermissions(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	roles := middleware.GetRoles(r.Context())

	perms, err := h.engine.GetMyPermissions(r.Context(), orgID, userID, roles)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve permissions")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"user_id":     userID,
			"roles":       roles,
			"permissions": perms,
		},
	})
}

// ============================================================
// ENDPOINT: GET /access/field-permissions?resource_type=risk
// ============================================================

func (h *AccessHandler) GetFieldPermissions(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	roles := middleware.GetRoles(r.Context())
	resourceType := r.URL.Query().Get("resource_type")

	if resourceType == "" {
		writeError(w, http.StatusBadRequest, "MISSING_PARAM", "resource_type query parameter is required")
		return
	}

	perms, err := h.masker.GetFieldPermissionsDetailed(r.Context(), orgID, userID, roles, resourceType)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve field permissions")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"resource_type": resourceType,
			"fields":        perms,
		},
	})
}
