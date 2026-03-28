package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/complianceforge/platform/internal/middleware"
	"github.com/complianceforge/platform/internal/models"
	"github.com/complianceforge/platform/internal/repository"
	"github.com/complianceforge/platform/internal/service"
)

type SettingsHandler struct {
	userRepo *repository.UserRepo
	orgRepo  *repository.OrganizationRepo
	authSvc  *service.AuthService
}

func NewSettingsHandler(userRepo *repository.UserRepo, orgRepo *repository.OrganizationRepo, authSvc *service.AuthService) *SettingsHandler {
	return &SettingsHandler{userRepo: userRepo, orgRepo: orgRepo, authSvc: authSvc}
}

// GetOrganization returns the current organization settings.
// GET /api/v1/settings
func (h *SettingsHandler) GetOrganization(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	org, err := h.orgRepo.GetByID(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Organization not found")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: org})
}

// UpdateOrganization updates organization settings.
// PUT /api/v1/settings
func (h *SettingsHandler) UpdateOrganization(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var req models.UpdateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if err := h.orgRepo.Update(r.Context(), orgID, req); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update organization")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Organization updated successfully"},
	})
}

// ListUsers returns paginated users for the organization.
// GET /api/v1/settings/users?search=&page=1&page_size=20
func (h *SettingsHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	params := parsePagination(r)

	users, total, err := h.userRepo.List(r.Context(), orgID, params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve users")
		return
	}

	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize != 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: users,
		Pagination: models.Pagination{
			Page: params.Page, PageSize: params.PageSize,
			TotalItems: total, TotalPages: totalPages,
			HasNext: params.Page < totalPages, HasPrev: params.Page > 1,
		},
	})
}

// CreateUser creates a new user in the organization.
// POST /api/v1/settings/users
func (h *SettingsHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "email, password, first_name, last_name required")
		return
	}

	if len(req.Password) < 12 {
		writeError(w, http.StatusBadRequest, "WEAK_PASSWORD", "Password must be at least 12 characters")
		return
	}

	user, err := h.authSvc.Register(r.Context(), orgID, req)
	if err != nil {
		writeError(w, http.StatusConflict, "CREATE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: user})
}

// GetUser returns a single user with roles.
// GET /api/v1/settings/users/{id}
func (h *SettingsHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid user ID")
		return
	}

	user, err := h.userRepo.GetByID(r.Context(), orgID, userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "User not found")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: user})
}

// UpdateUser updates a user's profile.
// PUT /api/v1/settings/users/{id}
func (h *SettingsHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid user ID")
		return
	}

	var req models.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if err := h.userRepo.Update(r.Context(), orgID, userID, req); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update user")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "User updated successfully"},
	})
}

// DeactivateUser soft-deletes a user.
// DELETE /api/v1/settings/users/{id}
func (h *SettingsHandler) DeactivateUser(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid user ID")
		return
	}

	currentUserID := middleware.GetUserID(r.Context())
	if userID == currentUserID {
		writeError(w, http.StatusConflict, "SELF_DELETE", "Cannot deactivate your own account")
		return
	}

	if err := h.userRepo.Deactivate(r.Context(), orgID, userID); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to deactivate user")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "User deactivated"},
	})
}

// AssignRole assigns a role to a user.
// POST /api/v1/settings/users/{id}/roles
func (h *SettingsHandler) AssignRole(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	currentUserID := middleware.GetUserID(r.Context())
	userID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid user ID")
		return
	}

	var req struct {
		RoleID uuid.UUID `json:"role_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if err := h.userRepo.AssignRole(r.Context(), orgID, userID, req.RoleID, currentUserID); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to assign role")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Role assigned successfully"},
	})
}

// ListRoles returns all available roles.
// GET /api/v1/settings/roles
func (h *SettingsHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	roles, err := h.userRepo.ListRoles(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve roles")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: roles})
}

// GetAuditLog returns paginated audit trail.
// GET /api/v1/settings/audit-log
func (h *SettingsHandler) GetAuditLog(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	params := parsePagination(r)

	logs, total, err := h.userRepo.GetAuditLog(r.Context(), orgID, params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve audit log")
		return
	}

	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize != 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: logs,
		Pagination: models.Pagination{
			Page: params.Page, PageSize: params.PageSize,
			TotalItems: total, TotalPages: totalPages,
			HasNext: params.Page < totalPages, HasPrev: params.Page > 1,
		},
	})
}
