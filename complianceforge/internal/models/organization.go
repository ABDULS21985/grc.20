package models

import (
	"time"

	"github.com/google/uuid"
)

// ============================================================
// ORGANIZATION
// ============================================================

// Organization represents a tenant in the multi-tenant platform.
type Organization struct {
	BaseModel
	Name                 string    `json:"name" db:"name"`
	Slug                 string    `json:"slug" db:"slug"`
	LegalName            string    `json:"legal_name,omitempty" db:"legal_name"`
	RegistrationNumber   string    `json:"registration_number,omitempty" db:"registration_number"`
	TaxID                string    `json:"tax_id,omitempty" db:"tax_id"`
	Industry             string    `json:"industry" db:"industry"`
	Sector               string    `json:"sector,omitempty" db:"sector"`
	CountryCode          string    `json:"country_code" db:"country_code"`
	HeadquartersAddress  JSONB     `json:"headquarters_address" db:"headquarters_address"`
	Status               OrgStatus `json:"status" db:"status"`
	Tier                 OrgTier   `json:"tier" db:"tier"`
	Settings             JSONB     `json:"settings" db:"settings"`
	Branding             JSONB     `json:"branding" db:"branding"`
	Timezone             string    `json:"timezone" db:"timezone"`
	DefaultLanguage      string    `json:"default_language" db:"default_language"`
	SupportedLanguages   []string  `json:"supported_languages" db:"supported_languages"`
	EmployeeCountRange   string    `json:"employee_count_range,omitempty" db:"employee_count_range"`
	AnnualRevenueRange   string    `json:"annual_revenue_range,omitempty" db:"annual_revenue_range"`
	ParentOrganizationID *uuid.UUID `json:"parent_organization_id,omitempty" db:"parent_organization_id"`
	Metadata             JSONB     `json:"metadata" db:"metadata"`
}

// CreateOrganizationRequest is the payload for creating a new organization.
type CreateOrganizationRequest struct {
	Name               string `json:"name" validate:"required,min=2,max=255"`
	LegalName          string `json:"legal_name,omitempty" validate:"max=500"`
	Industry           string `json:"industry" validate:"required"`
	CountryCode        string `json:"country_code" validate:"required,len=2"`
	Timezone           string `json:"timezone" validate:"required"`
	DefaultLanguage    string `json:"default_language" validate:"required,len=2"`
	EmployeeCountRange string `json:"employee_count_range,omitempty"`
}

// UpdateOrganizationRequest is the payload for updating an organization.
type UpdateOrganizationRequest struct {
	Name               *string `json:"name,omitempty" validate:"omitempty,min=2,max=255"`
	LegalName          *string `json:"legal_name,omitempty"`
	Industry           *string `json:"industry,omitempty"`
	CountryCode        *string `json:"country_code,omitempty" validate:"omitempty,len=2"`
	Timezone           *string `json:"timezone,omitempty"`
	EmployeeCountRange *string `json:"employee_count_range,omitempty"`
	Branding           *JSONB  `json:"branding,omitempty"`
	Settings           *JSONB  `json:"settings,omitempty"`
}

// ============================================================
// USER
// ============================================================

// User represents a platform user belonging to an organization.
type User struct {
	BaseModel
	Email                  string     `json:"email" db:"email"`
	PasswordHash           string     `json:"-" db:"password_hash"` // Never expose in JSON
	FirstName              string     `json:"first_name" db:"first_name"`
	LastName               string     `json:"last_name" db:"last_name"`
	JobTitle               string     `json:"job_title,omitempty" db:"job_title"`
	Department             string     `json:"department,omitempty" db:"department"`
	Phone                  string     `json:"phone,omitempty" db:"phone"`
	AvatarURL              string     `json:"avatar_url,omitempty" db:"avatar_url"`
	Status                 UserStatus `json:"status" db:"status"`
	IsSuperAdmin           bool       `json:"is_super_admin" db:"is_super_admin"`
	Timezone               string     `json:"timezone" db:"timezone"`
	Language               string     `json:"language" db:"language"`
	LastLoginAt            *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
	LastLoginIP            string     `json:"last_login_ip,omitempty" db:"last_login_ip"`
	PasswordChangedAt      *time.Time `json:"password_changed_at,omitempty" db:"password_changed_at"`
	FailedLoginAttempts    int        `json:"failed_login_attempts" db:"failed_login_attempts"`
	LockedUntil            *time.Time `json:"locked_until,omitempty" db:"locked_until"`
	NotificationPreferences JSONB     `json:"notification_preferences" db:"notification_preferences"`
	Metadata               JSONB      `json:"metadata" db:"metadata"`
	// Loaded via joins
	Roles []Role `json:"roles,omitempty" db:"-"`
}

// FullName returns the user's full name.
func (u *User) FullName() string {
	return u.FirstName + " " + u.LastName
}

// IsLocked checks if the user account is locked.
func (u *User) IsLocked() bool {
	if u.LockedUntil == nil {
		return false
	}
	return time.Now().Before(*u.LockedUntil)
}

// CreateUserRequest is the payload for creating a new user.
type CreateUserRequest struct {
	Email      string `json:"email" validate:"required,email"`
	Password   string `json:"password" validate:"required,min=12,max=128"`
	FirstName  string `json:"first_name" validate:"required,min=1,max=100"`
	LastName   string `json:"last_name" validate:"required,min=1,max=100"`
	JobTitle   string `json:"job_title,omitempty" validate:"max=200"`
	Department string `json:"department,omitempty" validate:"max=200"`
	Phone      string `json:"phone,omitempty" validate:"max=50"`
	RoleIDs    []uuid.UUID `json:"role_ids" validate:"required,min=1"`
}

// UpdateUserRequest is the payload for updating a user.
type UpdateUserRequest struct {
	FirstName  *string `json:"first_name,omitempty" validate:"omitempty,min=1,max=100"`
	LastName   *string `json:"last_name,omitempty" validate:"omitempty,min=1,max=100"`
	JobTitle   *string `json:"job_title,omitempty"`
	Department *string `json:"department,omitempty"`
	Phone      *string `json:"phone,omitempty"`
	Language   *string `json:"language,omitempty"`
	Timezone   *string `json:"timezone,omitempty"`
}

// LoginRequest is the payload for user authentication.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse contains tokens returned after successful authentication.
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // seconds
	TokenType    string `json:"token_type"`
	User         *User  `json:"user"`
}

// ============================================================
// ROLES & PERMISSIONS
// ============================================================

// Role represents a role in the RBAC system.
type Role struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	OrganizationID *uuid.UUID `json:"organization_id,omitempty" db:"organization_id"`
	Name           string     `json:"name" db:"name"`
	Slug           string     `json:"slug" db:"slug"`
	Description    string     `json:"description,omitempty" db:"description"`
	IsSystemRole   bool       `json:"is_system_role" db:"is_system_role"`
	IsCustom       bool       `json:"is_custom" db:"is_custom"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
	Permissions    []Permission `json:"permissions,omitempty" db:"-"`
}

// Permission represents a granular permission in the RBAC system.
type Permission struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Resource    string    `json:"resource" db:"resource"`
	Action      string    `json:"action" db:"action"`
	Description string    `json:"description,omitempty" db:"description"`
}

// AuditLog represents an entry in the immutable audit trail.
type AuditLog struct {
	ID             uuid.UUID `json:"id" db:"id"`
	OrganizationID uuid.UUID `json:"organization_id" db:"organization_id"`
	UserID         uuid.UUID `json:"user_id" db:"user_id"`
	Action         string    `json:"action" db:"action"`
	EntityType     string    `json:"entity_type" db:"entity_type"`
	EntityID       uuid.UUID `json:"entity_id" db:"entity_id"`
	Changes        JSONB     `json:"changes" db:"changes"`
	IPAddress      string    `json:"ip_address" db:"ip_address"`
	UserAgent      string    `json:"user_agent" db:"user_agent"`
	Metadata       JSONB     `json:"metadata" db:"metadata"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}
