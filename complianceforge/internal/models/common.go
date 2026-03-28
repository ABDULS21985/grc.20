// Package models defines all domain entities for the ComplianceForge GRC platform.
package models

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ============================================================
// COMMON TYPES & ENUMS
// ============================================================

// NullTime is a nullable time for soft deletes and optional timestamps.
type NullTime = sql.NullTime

// JSONB represents a PostgreSQL JSONB column.
type JSONB json.RawMessage

// Scan implements the sql.Scanner interface for JSONB.
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = JSONB([]byte("{}"))
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	*j = JSONB(bytes)
	return nil
}

// MarshalJSON implements json.Marshaler.
func (j JSONB) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("{}"), nil
	}
	return []byte(j), nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *JSONB) UnmarshalJSON(data []byte) error {
	*j = JSONB(data)
	return nil
}

// ============================================================
// PAGINATION
// ============================================================

// PaginationParams holds pagination query parameters.
type PaginationParams struct {
	Page     int    `json:"page" validate:"min=1"`
	PageSize int    `json:"page_size" validate:"min=1,max=100"`
	SortBy   string `json:"sort_by"`
	SortDir  string `json:"sort_dir" validate:"oneof=asc desc"`
	Search   string `json:"search"`
}

// DefaultPagination returns default pagination settings.
func DefaultPagination() PaginationParams {
	return PaginationParams{
		Page:     1,
		PageSize: 20,
		SortBy:   "created_at",
		SortDir:  "desc",
	}
}

// Offset calculates the SQL offset from page and page size.
func (p PaginationParams) Offset() int {
	return (p.Page - 1) * p.PageSize
}

// PaginatedResponse wraps a list response with pagination metadata.
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

// Pagination holds pagination metadata for responses.
type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// ============================================================
// API RESPONSE TYPES
// ============================================================

// APIResponse is the standard envelope for all API responses.
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
}

// APIError represents a structured error response.
type APIError struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

// ============================================================
// ENUMS — Organization
// ============================================================

type OrgStatus string

const (
	OrgStatusActive      OrgStatus = "active"
	OrgStatusSuspended   OrgStatus = "suspended"
	OrgStatusTrial       OrgStatus = "trial"
	OrgStatusDeactivated OrgStatus = "deactivated"
)

type OrgTier string

const (
	OrgTierStarter      OrgTier = "starter"
	OrgTierProfessional OrgTier = "professional"
	OrgTierEnterprise   OrgTier = "enterprise"
	OrgTierUnlimited    OrgTier = "unlimited"
)

// ============================================================
// ENUMS — User
// ============================================================

type UserStatus string

const (
	UserStatusActive             UserStatus = "active"
	UserStatusInactive           UserStatus = "inactive"
	UserStatusLocked             UserStatus = "locked"
	UserStatusPendingVerification UserStatus = "pending_verification"
)

// ============================================================
// ENUMS — Compliance
// ============================================================

type FrameworkCode string

const (
	FrameworkISO27001        FrameworkCode = "ISO27001"
	FrameworkUKGDPR          FrameworkCode = "UK_GDPR"
	FrameworkNCSCCAF         FrameworkCode = "NCSC_CAF"
	FrameworkCyberEssentials FrameworkCode = "CYBER_ESSENTIALS"
	FrameworkNIST80053       FrameworkCode = "NIST_800_53"
	FrameworkNISTCSF2        FrameworkCode = "NIST_CSF_2"
	FrameworkPCIDSS          FrameworkCode = "PCI_DSS_4"
	FrameworkITIL4           FrameworkCode = "ITIL_4"
	FrameworkCOBIT2019       FrameworkCode = "COBIT_2019"
)

type ControlStatus string

const (
	ControlStatusNotApplicable  ControlStatus = "not_applicable"
	ControlStatusNotImplemented ControlStatus = "not_implemented"
	ControlStatusPlanned        ControlStatus = "planned"
	ControlStatusPartial        ControlStatus = "partial"
	ControlStatusImplemented    ControlStatus = "implemented"
	ControlStatusEffective      ControlStatus = "effective"
)

type ControlType string

const (
	ControlTypePreventive   ControlType = "preventive"
	ControlTypeDetective    ControlType = "detective"
	ControlTypeCorrective   ControlType = "corrective"
	ControlTypeDirective    ControlType = "directive"
	ControlTypeCompensating ControlType = "compensating"
)

type ImplementationType string

const (
	ImplTypeTechnical      ImplementationType = "technical"
	ImplTypeAdministrative ImplementationType = "administrative"
	ImplTypePhysical       ImplementationType = "physical"
	ImplTypeManagement     ImplementationType = "management"
)

// ============================================================
// ENUMS — Risk
// ============================================================

type RiskLevel string

const (
	RiskLevelCritical RiskLevel = "critical"
	RiskLevelHigh     RiskLevel = "high"
	RiskLevelMedium   RiskLevel = "medium"
	RiskLevelLow      RiskLevel = "low"
	RiskLevelVeryLow  RiskLevel = "very_low"
)

type RiskStatus string

const (
	RiskStatusIdentified RiskStatus = "identified"
	RiskStatusAssessed   RiskStatus = "assessed"
	RiskStatusTreated    RiskStatus = "treated"
	RiskStatusAccepted   RiskStatus = "accepted"
	RiskStatusClosed     RiskStatus = "closed"
	RiskStatusMonitoring RiskStatus = "monitoring"
)

type TreatmentType string

const (
	TreatmentMitigate TreatmentType = "mitigate"
	TreatmentTransfer TreatmentType = "transfer"
	TreatmentAvoid    TreatmentType = "avoid"
	TreatmentAccept   TreatmentType = "accept"
)

// ============================================================
// ENUMS — Policy
// ============================================================

type PolicyStatus string

const (
	PolicyStatusDraft           PolicyStatus = "draft"
	PolicyStatusUnderReview     PolicyStatus = "under_review"
	PolicyStatusPendingApproval PolicyStatus = "pending_approval"
	PolicyStatusApproved        PolicyStatus = "approved"
	PolicyStatusPublished       PolicyStatus = "published"
	PolicyStatusArchived        PolicyStatus = "archived"
	PolicyStatusRetired         PolicyStatus = "retired"
)

// ============================================================
// ENUMS — Audit
// ============================================================

type AuditStatus string

const (
	AuditStatusPlanned    AuditStatus = "planned"
	AuditStatusInProgress AuditStatus = "in_progress"
	AuditStatusCompleted  AuditStatus = "completed"
	AuditStatusCancelled  AuditStatus = "cancelled"
)

type FindingSeverity string

const (
	FindingSeverityCritical FindingSeverity = "critical"
	FindingSeverityHigh     FindingSeverity = "high"
	FindingSeverityMedium   FindingSeverity = "medium"
	FindingSeverityLow      FindingSeverity = "low"
	FindingSeverityInfo     FindingSeverity = "informational"
)

// ============================================================
// ENUMS — Incidents
// ============================================================

type IncidentSeverity string

const (
	IncidentSeverityCritical IncidentSeverity = "critical"
	IncidentSeverityHigh     IncidentSeverity = "high"
	IncidentSeverityMedium   IncidentSeverity = "medium"
	IncidentSeverityLow      IncidentSeverity = "low"
)

type IncidentStatus string

const (
	IncidentStatusOpen         IncidentStatus = "open"
	IncidentStatusInvestigating IncidentStatus = "investigating"
	IncidentStatusContained     IncidentStatus = "contained"
	IncidentStatusResolved      IncidentStatus = "resolved"
	IncidentStatusClosed        IncidentStatus = "closed"
)

// ============================================================
// BASE MODEL
// ============================================================

// BaseModel contains fields common to all entities.
type BaseModel struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID  `json:"organization_id" db:"organization_id"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}
