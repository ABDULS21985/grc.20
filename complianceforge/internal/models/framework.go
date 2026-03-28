package models

import (
	"time"

	"github.com/google/uuid"
)

// ============================================================
// COMPLIANCE FRAMEWORK
// ============================================================

// ComplianceFramework represents a regulatory or standards framework.
type ComplianceFramework struct {
	ID                  uuid.UUID     `json:"id" db:"id"`
	OrganizationID      *uuid.UUID    `json:"organization_id,omitempty" db:"organization_id"`
	Code                FrameworkCode `json:"code" db:"code"`
	Name                string        `json:"name" db:"name"`
	FullName            string        `json:"full_name,omitempty" db:"full_name"`
	Version             string        `json:"version" db:"version"`
	Description         string        `json:"description,omitempty" db:"description"`
	IssuingBody         string        `json:"issuing_body,omitempty" db:"issuing_body"`
	Category            string        `json:"category" db:"category"`
	ApplicableRegions   []string      `json:"applicable_regions" db:"applicable_regions"`
	ApplicableIndustries []string     `json:"applicable_industries" db:"applicable_industries"`
	IsSystemFramework   bool          `json:"is_system_framework" db:"is_system_framework"`
	IsActive            bool          `json:"is_active" db:"is_active"`
	EffectiveDate       *time.Time    `json:"effective_date,omitempty" db:"effective_date"`
	SunsetDate          *time.Time    `json:"sunset_date,omitempty" db:"sunset_date"`
	TotalControls       int           `json:"total_controls" db:"total_controls"`
	IconURL             string        `json:"icon_url,omitempty" db:"icon_url"`
	ColorHex            string        `json:"color_hex,omitempty" db:"color_hex"`
	Metadata            JSONB         `json:"metadata" db:"metadata"`
	CreatedAt           time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time     `json:"updated_at" db:"updated_at"`
	// Populated via joins
	Domains []FrameworkDomain `json:"domains,omitempty" db:"-"`
}

// ============================================================
// FRAMEWORK DOMAIN
// ============================================================

// FrameworkDomain represents a top-level grouping within a framework.
// Examples: ISO 27001 Annex A themes, NIST CSF Functions, PCI DSS Requirements.
type FrameworkDomain struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	FrameworkID    uuid.UUID  `json:"framework_id" db:"framework_id"`
	Code           string     `json:"code" db:"code"`
	Name           string     `json:"name" db:"name"`
	Description    string     `json:"description,omitempty" db:"description"`
	SortOrder      int        `json:"sort_order" db:"sort_order"`
	ParentDomainID *uuid.UUID `json:"parent_domain_id,omitempty" db:"parent_domain_id"`
	DepthLevel     int        `json:"depth_level" db:"depth_level"`
	TotalControls  int        `json:"total_controls" db:"total_controls"`
	Metadata       JSONB      `json:"metadata" db:"metadata"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
	// Populated via joins
	Controls []FrameworkControl `json:"controls,omitempty" db:"-"`
}

// ============================================================
// FRAMEWORK CONTROL
// ============================================================

// FrameworkControl represents an individual control or requirement within a framework.
type FrameworkControl struct {
	ID                   uuid.UUID          `json:"id" db:"id"`
	FrameworkID          uuid.UUID          `json:"framework_id" db:"framework_id"`
	DomainID             uuid.UUID          `json:"domain_id" db:"domain_id"`
	Code                 string             `json:"code" db:"code"`
	Title                string             `json:"title" db:"title"`
	Description          string             `json:"description,omitempty" db:"description"`
	Guidance             string             `json:"guidance,omitempty" db:"guidance"`
	Objective            string             `json:"objective,omitempty" db:"objective"`
	ControlType          ControlType        `json:"control_type" db:"control_type"`
	ImplementationType   ImplementationType `json:"implementation_type" db:"implementation_type"`
	IsMandatory          bool               `json:"is_mandatory" db:"is_mandatory"`
	Priority             string             `json:"priority" db:"priority"`
	SortOrder            int                `json:"sort_order" db:"sort_order"`
	ParentControlID      *uuid.UUID         `json:"parent_control_id,omitempty" db:"parent_control_id"`
	DepthLevel           int                `json:"depth_level" db:"depth_level"`
	EvidenceRequirements JSONB              `json:"evidence_requirements" db:"evidence_requirements"`
	TestProcedures       JSONB              `json:"test_procedures" db:"test_procedures"`
	References           JSONB              `json:"references" db:"references"`
	Keywords             []string           `json:"keywords" db:"keywords"`
	Metadata             JSONB              `json:"metadata" db:"metadata"`
	CreatedAt            time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time          `json:"updated_at" db:"updated_at"`
}

// ============================================================
// CROSS-FRAMEWORK MAPPING
// ============================================================

// FrameworkControlMapping maps controls between different frameworks.
type FrameworkControlMapping struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	SourceControlID uuid.UUID  `json:"source_control_id" db:"source_control_id"`
	TargetControlID uuid.UUID  `json:"target_control_id" db:"target_control_id"`
	MappingType     string     `json:"mapping_type" db:"mapping_type"` // equivalent, partial, related
	MappingStrength float64    `json:"mapping_strength" db:"mapping_strength"` // 0.00-1.00
	Notes           string     `json:"notes,omitempty" db:"notes"`
	IsVerified      bool       `json:"is_verified" db:"is_verified"`
	VerifiedBy      *uuid.UUID `json:"verified_by,omitempty" db:"verified_by"`
	VerifiedAt      *time.Time `json:"verified_at,omitempty" db:"verified_at"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

// ============================================================
// ORGANIZATION-LEVEL FRAMEWORK ADOPTION
// ============================================================

// OrganizationFramework tracks which frameworks an organization has adopted.
type OrganizationFramework struct {
	ID                   uuid.UUID  `json:"id" db:"id"`
	OrganizationID       uuid.UUID  `json:"organization_id" db:"organization_id"`
	FrameworkID          uuid.UUID  `json:"framework_id" db:"framework_id"`
	Status               string     `json:"status" db:"status"`
	AdoptionDate         *time.Time `json:"adoption_date,omitempty" db:"adoption_date"`
	TargetCompletionDate *time.Time `json:"target_completion_date,omitempty" db:"target_completion_date"`
	CertificationDate    *time.Time `json:"certification_date,omitempty" db:"certification_date"`
	CertificationExpiry  *time.Time `json:"certification_expiry,omitempty" db:"certification_expiry"`
	CertifyingBody       string     `json:"certifying_body,omitempty" db:"certifying_body"`
	CertificateNumber    string     `json:"certificate_number,omitempty" db:"certificate_number"`
	ScopeDescription     string     `json:"scope_description,omitempty" db:"scope_description"`
	ComplianceScore      float64    `json:"compliance_score" db:"compliance_score"`
	LastAssessmentDate   *time.Time `json:"last_assessment_date,omitempty" db:"last_assessment_date"`
	AssessmentFrequency  string     `json:"assessment_frequency" db:"assessment_frequency"`
	ResponsibleUserID    *uuid.UUID `json:"responsible_user_id,omitempty" db:"responsible_user_id"`
	Metadata             JSONB      `json:"metadata" db:"metadata"`
	CreatedAt            time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at" db:"updated_at"`
	// Populated via joins
	Framework *ComplianceFramework `json:"framework,omitempty" db:"-"`
}

// AdoptFrameworkRequest is the payload to adopt a framework.
type AdoptFrameworkRequest struct {
	FrameworkID          uuid.UUID `json:"framework_id" validate:"required"`
	TargetCompletionDate string    `json:"target_completion_date,omitempty"` // YYYY-MM-DD
	ScopeDescription     string    `json:"scope_description,omitempty"`
	AssessmentFrequency  string    `json:"assessment_frequency" validate:"required,oneof=monthly quarterly semi_annually annually"`
	ResponsibleUserID    uuid.UUID `json:"responsible_user_id" validate:"required"`
}

// ============================================================
// CONTROL IMPLEMENTATION
// ============================================================

// ControlImplementation tracks how an organization implements each control.
type ControlImplementation struct {
	BaseModel
	FrameworkControlID            uuid.UUID     `json:"framework_control_id" db:"framework_control_id"`
	OrgFrameworkID                uuid.UUID     `json:"org_framework_id" db:"org_framework_id"`
	Status                        ControlStatus `json:"status" db:"status"`
	ImplementationStatus          string        `json:"implementation_status" db:"implementation_status"`
	MaturityLevel                 int           `json:"maturity_level" db:"maturity_level"` // 0-5 CMMI
	OwnerUserID                   *uuid.UUID    `json:"owner_user_id,omitempty" db:"owner_user_id"`
	ReviewerUserID                *uuid.UUID    `json:"reviewer_user_id,omitempty" db:"reviewer_user_id"`
	ImplementationDescription     string        `json:"implementation_description,omitempty" db:"implementation_description"`
	ImplementationNotes           string        `json:"implementation_notes,omitempty" db:"implementation_notes"`
	CompensatingControlDescription string      `json:"compensating_control_description,omitempty" db:"compensating_control_description"`
	GapDescription                string        `json:"gap_description,omitempty" db:"gap_description"`
	RemediationPlan               string        `json:"remediation_plan,omitempty" db:"remediation_plan"`
	RemediationDueDate            *time.Time    `json:"remediation_due_date,omitempty" db:"remediation_due_date"`
	TestFrequency                 string        `json:"test_frequency" db:"test_frequency"`
	LastTestedAt                  *time.Time    `json:"last_tested_at,omitempty" db:"last_tested_at"`
	LastTestedBy                  *uuid.UUID    `json:"last_tested_by,omitempty" db:"last_tested_by"`
	LastTestResult                string        `json:"last_test_result,omitempty" db:"last_test_result"`
	EffectivenessScore            float64       `json:"effectiveness_score" db:"effectiveness_score"`
	RiskIfNotImplemented          string        `json:"risk_if_not_implemented" db:"risk_if_not_implemented"`
	AutomationLevel               string        `json:"automation_level" db:"automation_level"`
	Tags                          []string      `json:"tags" db:"tags"`
	Metadata                      JSONB         `json:"metadata" db:"metadata"`
	// Populated via joins
	Control  *FrameworkControl `json:"control,omitempty" db:"-"`
	Evidence []ControlEvidence `json:"evidence,omitempty" db:"-"`
}

// UpdateControlImplementationRequest is the payload for updating a control implementation.
type UpdateControlImplementationRequest struct {
	Status                         *ControlStatus `json:"status,omitempty"`
	MaturityLevel                  *int           `json:"maturity_level,omitempty" validate:"omitempty,min=0,max=5"`
	OwnerUserID                    *uuid.UUID     `json:"owner_user_id,omitempty"`
	ImplementationDescription      *string        `json:"implementation_description,omitempty"`
	GapDescription                 *string        `json:"gap_description,omitempty"`
	RemediationPlan                *string        `json:"remediation_plan,omitempty"`
	RemediationDueDate             *string        `json:"remediation_due_date,omitempty"` // YYYY-MM-DD
	CompensatingControlDescription *string        `json:"compensating_control_description,omitempty"`
	AutomationLevel                *string        `json:"automation_level,omitempty" validate:"omitempty,oneof=fully_automated semi_automated manual"`
}

// ============================================================
// CONTROL EVIDENCE
// ============================================================

// ControlEvidence represents evidence linked to a control implementation.
type ControlEvidence struct {
	BaseModel
	ControlImplementationID uuid.UUID  `json:"control_implementation_id" db:"control_implementation_id"`
	Title                   string     `json:"title" db:"title"`
	Description             string     `json:"description,omitempty" db:"description"`
	EvidenceType            string     `json:"evidence_type" db:"evidence_type"`
	FilePath                string     `json:"file_path,omitempty" db:"file_path"`
	FileName                string     `json:"file_name,omitempty" db:"file_name"`
	FileSizeBytes           int64      `json:"file_size_bytes" db:"file_size_bytes"`
	MimeType                string     `json:"mime_type,omitempty" db:"mime_type"`
	FileHash                string     `json:"file_hash,omitempty" db:"file_hash"`
	CollectionMethod        string     `json:"collection_method" db:"collection_method"`
	CollectedAt             time.Time  `json:"collected_at" db:"collected_at"`
	CollectedBy             uuid.UUID  `json:"collected_by" db:"collected_by"`
	ValidFrom               *time.Time `json:"valid_from,omitempty" db:"valid_from"`
	ValidUntil              *time.Time `json:"valid_until,omitempty" db:"valid_until"`
	IsCurrent               bool       `json:"is_current" db:"is_current"`
	ReviewStatus            string     `json:"review_status" db:"review_status"`
	ReviewedBy              *uuid.UUID `json:"reviewed_by,omitempty" db:"reviewed_by"`
	ReviewedAt              *time.Time `json:"reviewed_at,omitempty" db:"reviewed_at"`
	ReviewNotes             string     `json:"review_notes,omitempty" db:"review_notes"`
	Metadata                JSONB      `json:"metadata" db:"metadata"`
}

// ============================================================
// COMPLIANCE SCORE (Aggregated View)
// ============================================================

// ComplianceScoreSummary is a read model for dashboard display.
type ComplianceScoreSummary struct {
	FrameworkID     uuid.UUID `json:"framework_id"`
	FrameworkCode   string    `json:"framework_code"`
	FrameworkName   string    `json:"framework_name"`
	TotalControls   int       `json:"total_controls"`
	Implemented     int       `json:"implemented"`
	PartiallyImpl   int       `json:"partially_implemented"`
	NotImplemented  int       `json:"not_implemented"`
	NotApplicable   int       `json:"not_applicable"`
	ComplianceScore float64   `json:"compliance_score"` // percentage
	MaturityAvg     float64   `json:"maturity_avg"`     // average maturity 0-5
}
