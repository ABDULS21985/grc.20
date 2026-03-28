package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/complianceforge/platform/internal/middleware"
	"github.com/complianceforge/platform/internal/models"
	"github.com/complianceforge/platform/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Login authenticates a user and returns JWT tokens.
// POST /api/v1/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "Email and password are required")
		return
	}

	resp, err := h.authService.Login(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			writeError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid email or password")
		case errors.Is(err, service.ErrAccountLocked):
			writeError(w, http.StatusForbidden, "ACCOUNT_LOCKED", "Account is locked due to too many failed attempts")
		case errors.Is(err, service.ErrAccountInactive):
			writeError(w, http.StatusForbidden, "ACCOUNT_INACTIVE", "Account is inactive")
		default:
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Authentication failed")
		}
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: resp})
}

// Register creates a new user (org admin only).
// POST /api/v1/auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "Required fields: email, password, first_name, last_name")
		return
	}

	if len(req.Password) < 12 {
		writeError(w, http.StatusBadRequest, "WEAK_PASSWORD", "Password must be at least 12 characters")
		return
	}

	user, err := h.authService.Register(r.Context(), orgID, req)
	if err != nil {
		if errors.Is(err, service.ErrEmailExists) {
			writeError(w, http.StatusConflict, "EMAIL_EXISTS", "A user with this email already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create user")
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: user})
}

// Me returns the current authenticated user's profile.
// GET /api/v1/auth/me
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	orgID := middleware.GetOrgID(r.Context())
	email := r.Context().Value(middleware.ContextKeyEmail)
	roles := middleware.GetRoles(r.Context())

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"user_id":         userID,
			"organization_id": orgID,
			"email":           email,
			"roles":           roles,
		},
	})
}
