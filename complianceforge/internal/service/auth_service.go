// Package service implements the business logic layer for ComplianceForge.
package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"

	"github.com/complianceforge/platform/internal/config"
	"github.com/complianceforge/platform/internal/middleware"
	"github.com/complianceforge/platform/internal/models"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrAccountLocked      = errors.New("account is locked")
	ErrAccountInactive    = errors.New("account is inactive")
	ErrEmailExists        = errors.New("email already registered")
)

// AuthService handles authentication, JWT generation, and user management.
type AuthService struct {
	pool *pgxpool.Pool
	cfg  config.JWTConfig
}

// NewAuthService creates a new AuthService.
func NewAuthService(pool *pgxpool.Pool, cfg config.JWTConfig) *AuthService {
	return &AuthService{pool: pool, cfg: cfg}
}

// Register creates a new user account with hashed password.
func (s *AuthService) Register(ctx context.Context, orgID uuid.UUID, req models.CreateUserRequest) (*models.User, error) {
	// Check if email already exists in this org
	var exists bool
	err := s.pool.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM users WHERE organization_id = $1 AND email = $2 AND deleted_at IS NULL)",
		orgID, req.Email,
	).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check email: %w", err)
	}
	if exists {
		return nil, ErrEmailExists
	}

	// Hash password with bcrypt (cost 12 for production security)
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User{}
	now := time.Now()

	err = s.pool.QueryRow(ctx, `
		INSERT INTO users (organization_id, email, password_hash, first_name, last_name,
			job_title, department, phone, status, password_changed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'active', $9)
		RETURNING id, organization_id, email, first_name, last_name, job_title,
			department, phone, status, is_super_admin, timezone, language,
			created_at, updated_at`,
		orgID, req.Email, string(hash), req.FirstName, req.LastName,
		req.JobTitle, req.Department, req.Phone, now,
	).Scan(
		&user.ID, &user.OrganizationID, &user.Email, &user.FirstName, &user.LastName,
		&user.JobTitle, &user.Department, &user.Phone, &user.Status, &user.IsSuperAdmin,
		&user.Timezone, &user.Language, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Assign roles
	for _, roleID := range req.RoleIDs {
		_, err = s.pool.Exec(ctx,
			"INSERT INTO user_roles (user_id, role_id, organization_id) VALUES ($1, $2, $3)",
			user.ID, roleID, orgID,
		)
		if err != nil {
			log.Warn().Err(err).Str("role_id", roleID.String()).Msg("Failed to assign role")
		}
	}

	return user, nil
}

// Login authenticates a user and returns JWT tokens.
func (s *AuthService) Login(ctx context.Context, req models.LoginRequest) (*models.LoginResponse, error) {
	var user models.User
	var passwordHash string

	err := s.pool.QueryRow(ctx, `
		SELECT id, organization_id, email, password_hash, first_name, last_name,
			job_title, department, status, is_super_admin, timezone, language,
			failed_login_attempts, locked_until, created_at, updated_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
		LIMIT 1`,
		req.Email,
	).Scan(
		&user.ID, &user.OrganizationID, &user.Email, &passwordHash,
		&user.FirstName, &user.LastName, &user.JobTitle, &user.Department,
		&user.Status, &user.IsSuperAdmin, &user.Timezone, &user.Language,
		&user.FailedLoginAttempts, &user.LockedUntil,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Check account status
	if user.Status == models.UserStatusLocked || user.IsLocked() {
		return nil, ErrAccountLocked
	}
	if user.Status == models.UserStatusInactive {
		return nil, ErrAccountInactive
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		// Increment failed attempts
		s.incrementFailedAttempts(ctx, user.ID)
		return nil, ErrInvalidCredentials
	}

	// Reset failed attempts and update last login
	s.recordSuccessfulLogin(ctx, user.ID, "")

	// Load roles
	roles, err := s.getUserRoles(ctx, user.ID, user.OrganizationID)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to load user roles")
	}
	user.Roles = roles

	roleSlugs := make([]string, len(roles))
	for i, r := range roles {
		roleSlugs[i] = r.Slug
	}

	// Generate tokens
	accessToken, err := s.generateAccessToken(user.ID, user.OrganizationID, user.Email, roleSlugs)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.generateRefreshToken(user.ID, user.OrganizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &models.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.cfg.AccessTokenExpiry.Seconds()),
		TokenType:    "Bearer",
		User:         &user,
	}, nil
}

// generateAccessToken creates a signed JWT access token.
func (s *AuthService) generateAccessToken(userID, orgID uuid.UUID, email string, roles []string) (string, error) {
	now := time.Now()
	claims := middleware.Claims{
		UserID:         userID,
		OrganizationID: orgID,
		Email:          email,
		Roles:          roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.cfg.AccessTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "complianceforge",
			Subject:   userID.String(),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.Secret))
}

// generateRefreshToken creates a longer-lived refresh token.
func (s *AuthService) generateRefreshToken(userID, orgID uuid.UUID) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(now.Add(s.cfg.RefreshTokenExpiry)),
		IssuedAt:  jwt.NewNumericDate(now),
		Issuer:    "complianceforge",
		Subject:   userID.String(),
		ID:        uuid.New().String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.Secret))
}

// getUserRoles returns all roles assigned to a user in an organization.
func (s *AuthService) getUserRoles(ctx context.Context, userID, orgID uuid.UUID) ([]models.Role, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT r.id, r.name, r.slug, r.description, r.is_system_role
		FROM roles r
		JOIN user_roles ur ON ur.role_id = r.id
		WHERE ur.user_id = $1 AND ur.organization_id = $2 AND r.deleted_at IS NULL`,
		userID, orgID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []models.Role
	for rows.Next() {
		var r models.Role
		if err := rows.Scan(&r.ID, &r.Name, &r.Slug, &r.Description, &r.IsSystemRole); err != nil {
			return nil, err
		}
		roles = append(roles, r)
	}
	return roles, nil
}

func (s *AuthService) incrementFailedAttempts(ctx context.Context, userID uuid.UUID) {
	// Lock after 5 failed attempts for 30 minutes
	_, _ = s.pool.Exec(ctx, `
		UPDATE users SET
			failed_login_attempts = failed_login_attempts + 1,
			locked_until = CASE
				WHEN failed_login_attempts >= 4 THEN NOW() + INTERVAL '30 minutes'
				ELSE locked_until
			END,
			status = CASE
				WHEN failed_login_attempts >= 4 THEN 'locked'::user_status
				ELSE status
			END
		WHERE id = $1`, userID)
}

func (s *AuthService) recordSuccessfulLogin(ctx context.Context, userID uuid.UUID, ip string) {
	_, _ = s.pool.Exec(ctx, `
		UPDATE users SET
			failed_login_attempts = 0,
			locked_until = NULL,
			last_login_at = NOW(),
			last_login_ip = NULLIF($2, '')::INET
		WHERE id = $1`, userID, ip)
}
