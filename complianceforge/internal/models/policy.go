package models

import (
	"time"

	"github.com/google/uuid"
)

// ============================================================
// POLICY
// ============================================================

type Policy struct {
	BaseModel
	PolicyRef                string       `json:"policy_ref" db:"policy_ref"`
	Title                    string       `json:"title" db:"title"`
	CategoryID               *uuid.UUID   `json:"category_id,omitempty" db:"category_id"`
	Status                   PolicyStatus `json:"status" db:"status"`
	Classification           string       `json:"classification" db:"classification"`
	OwnerUserID              *uuid.UUID   `json:"owner_user_id,omitempty" db:"owner_user_id"`
	AuthorUserID             *uuid.UUID   `json:"author_user_id,omitempty" db:"author_user_id"`
	ApproverUserID           *uuid.UUID   `json:"approver_user_id,omitempty" db:"approver_user_id"`
	CurrentVersion           int          `json:"current_version" db:"current_version"`
	CurrentVersionID         *uuid.UUID   `json:"current_version_id,omitempty" db:"current_version_id"`
	ReviewFrequencyMonths    int          `json:"review_frequency_months" db:"review_frequency_months"`
	LastReviewDate           *time.Time   `json:"last_review_date,omitempty" db:"last_review_date"`
	NextReviewDate           *time.Time   `json:"next_review_date,omitempty" db:"next_review_date"`
	ReviewStatus             string       `json:"review_status" db:"review_status"`
	AppliesToAll             bool         `json:"applies_to_all" db:"applies_to_all"`
	ApplicableDepartments    []uuid.UUID  `json:"applicable_departments" db:"applicable_departments"`
	ApplicableRoles          []string     `json:"applicable_roles" db:"applicable_roles"`
	ApplicableLocations      []string     `json:"applicable_locations" db:"applicable_locations"`
	LinkedFrameworkIDs       []uuid.UUID  `json:"linked_framework_ids" db:"linked_framework_ids"`
	LinkedControlIDs         []uuid.UUID  `json:"linked_control_ids" db:"linked_control_ids"`
	LinkedRiskIDs            []uuid.UUID  `json:"linked_risk_ids" db:"linked_risk_ids"`
	ParentPolicyID           *uuid.UUID   `json:"parent_policy_id,omitempty" db:"parent_policy_id"`
	SupersedesPolicyID       *uuid.UUID   `json:"supersedes_policy_id,omitempty" db:"supersedes_policy_id"`
	EffectiveDate            *time.Time   `json:"effective_date,omitempty" db:"effective_date"`
	ExpiryDate               *time.Time   `json:"expiry_date,omitempty" db:"expiry_date"`
	Tags                     []string     `json:"tags" db:"tags"`
	IsMandatory              bool         `json:"is_mandatory" db:"is_mandatory"`
	RequiresAttestation      bool         `json:"requires_attestation" db:"requires_attestation"`
	AttestationFrequencyMonths int        `json:"attestation_frequency_months" db:"attestation_frequency_months"`
	Metadata                 JSONB        `json:"metadata" db:"metadata"`
}

type PolicyVersion struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	PolicyID        uuid.UUID  `json:"policy_id" db:"policy_id"`
	OrganizationID  uuid.UUID  `json:"organization_id" db:"organization_id"`
	VersionNumber   int        `json:"version_number" db:"version_number"`
	VersionLabel    string     `json:"version_label" db:"version_label"`
	Title           string     `json:"title" db:"title"`
	ContentHTML     string     `json:"content_html" db:"content_html"`
	ContentText     string     `json:"content_text,omitempty" db:"content_text"`
	Summary         string     `json:"summary,omitempty" db:"summary"`
	ChangeDescription string  `json:"change_description,omitempty" db:"change_description"`
	ChangeType      string     `json:"change_type" db:"change_type"`
	Language        string     `json:"language" db:"language"`
	WordCount       int        `json:"word_count" db:"word_count"`
	Status          string     `json:"status" db:"status"`
	CreatedBy       uuid.UUID  `json:"created_by" db:"created_by"`
	PublishedAt     *time.Time `json:"published_at,omitempty" db:"published_at"`
	PublishedBy     *uuid.UUID `json:"published_by,omitempty" db:"published_by"`
	FilePath        string     `json:"file_path,omitempty" db:"file_path"`
	FileHash        string     `json:"file_hash,omitempty" db:"file_hash"`
	Metadata        JSONB      `json:"metadata" db:"metadata"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

type PolicyAttestation struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	PolicyID         uuid.UUID  `json:"policy_id" db:"policy_id"`
	PolicyVersionID  uuid.UUID  `json:"policy_version_id" db:"policy_version_id"`
	OrganizationID   uuid.UUID  `json:"organization_id" db:"organization_id"`
	UserID           uuid.UUID  `json:"user_id" db:"user_id"`
	CampaignID       *uuid.UUID `json:"campaign_id,omitempty" db:"campaign_id"`
	Status           string     `json:"status" db:"status"`
	AttestedAt       *time.Time `json:"attested_at,omitempty" db:"attested_at"`
	AttestedFromIP   string     `json:"attested_from_ip,omitempty" db:"attested_from_ip"`
	AttestationMethod string    `json:"attestation_method" db:"attestation_method"`
	DueDate          *time.Time `json:"due_date,omitempty" db:"due_date"`
	ReminderCount    int        `json:"reminder_count" db:"reminder_count"`
	ExpiresAt        *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
}

type CreatePolicyRequest struct {
	Title                 string    `json:"title" validate:"required,min=5,max=500"`
	CategoryID            uuid.UUID `json:"category_id" validate:"required"`
	Classification        string    `json:"classification" validate:"required,oneof=public internal confidential restricted"`
	ContentHTML           string    `json:"content_html" validate:"required"`
	Summary               string    `json:"summary,omitempty"`
	OwnerUserID           uuid.UUID `json:"owner_user_id" validate:"required"`
	ApproverUserID        uuid.UUID `json:"approver_user_id" validate:"required"`
	ReviewFrequencyMonths int       `json:"review_frequency_months" validate:"required,min=1,max=36"`
	IsMandatory           bool      `json:"is_mandatory"`
	RequiresAttestation   bool      `json:"requires_attestation"`
	Tags                  []string  `json:"tags,omitempty"`
}

// ============================================================
// AUDIT
// ============================================================

type Audit struct {
	BaseModel
	AuditRef            string      `json:"audit_ref" db:"audit_ref"`
	Title               string      `json:"title" db:"title"`
	AuditType           string      `json:"audit_type" db:"audit_type"` // internal, external, regulatory, compliance
	Status              AuditStatus `json:"status" db:"status"`
	Description         string      `json:"description,omitempty" db:"description"`
	Scope               string      `json:"scope,omitempty" db:"scope"`
	Methodology         string      `json:"methodology,omitempty" db:"methodology"`
	LeadAuditorID       *uuid.UUID  `json:"lead_auditor_id,omitempty" db:"lead_auditor_id"`
	AuditTeamIDs        []uuid.UUID `json:"audit_team_ids" db:"audit_team_ids"`
	LinkedFrameworkIDs  []uuid.UUID `json:"linked_framework_ids" db:"linked_framework_ids"`
	PlannedStartDate    *time.Time  `json:"planned_start_date,omitempty" db:"planned_start_date"`
	PlannedEndDate      *time.Time  `json:"planned_end_date,omitempty" db:"planned_end_date"`
	ActualStartDate     *time.Time  `json:"actual_start_date,omitempty" db:"actual_start_date"`
	ActualEndDate       *time.Time  `json:"actual_end_date,omitempty" db:"actual_end_date"`
	TotalFindings       int         `json:"total_findings" db:"total_findings"`
	CriticalFindings    int         `json:"critical_findings" db:"critical_findings"`
	HighFindings        int         `json:"high_findings" db:"high_findings"`
	MediumFindings      int         `json:"medium_findings" db:"medium_findings"`
	LowFindings         int         `json:"low_findings" db:"low_findings"`
	ReportFilePath      string      `json:"report_file_path,omitempty" db:"report_file_path"`
	Conclusion          string      `json:"conclusion,omitempty" db:"conclusion"`
	Tags                []string    `json:"tags" db:"tags"`
	Metadata            JSONB       `json:"metadata" db:"metadata"`
	Findings            []AuditFinding `json:"findings,omitempty" db:"-"`
}

type AuditFinding struct {
	BaseModel
	AuditID             uuid.UUID       `json:"audit_id" db:"audit_id"`
	FindingRef          string          `json:"finding_ref" db:"finding_ref"`
	Title               string          `json:"title" db:"title"`
	Description         string          `json:"description" db:"description"`
	Severity            FindingSeverity `json:"severity" db:"severity"`
	Status              string          `json:"status" db:"status"` // open, in_progress, remediated, closed, accepted
	FindingType         string          `json:"finding_type" db:"finding_type"` // non_conformity, observation, opportunity
	ControlID           *uuid.UUID      `json:"control_id,omitempty" db:"control_id"`
	RootCause           string          `json:"root_cause,omitempty" db:"root_cause"`
	Recommendation      string          `json:"recommendation,omitempty" db:"recommendation"`
	ManagementResponse  string          `json:"management_response,omitempty" db:"management_response"`
	ResponsibleUserID   *uuid.UUID      `json:"responsible_user_id,omitempty" db:"responsible_user_id"`
	DueDate             *time.Time      `json:"due_date,omitempty" db:"due_date"`
	ClosedDate          *time.Time      `json:"closed_date,omitempty" db:"closed_date"`
	EvidenceIDs         []uuid.UUID     `json:"evidence_ids" db:"evidence_ids"`
	LinkedRiskID        *uuid.UUID      `json:"linked_risk_id,omitempty" db:"linked_risk_id"`
	Metadata            JSONB           `json:"metadata" db:"metadata"`
}

// ============================================================
// INCIDENT
// ============================================================

type Incident struct {
	BaseModel
	IncidentRef          string           `json:"incident_ref" db:"incident_ref"`
	Title                string           `json:"title" db:"title"`
	Description          string           `json:"description" db:"description"`
	IncidentType         string           `json:"incident_type" db:"incident_type"` // data_breach, security, operational, compliance, whistleblower
	Severity             IncidentSeverity `json:"severity" db:"severity"`
	Status               IncidentStatus   `json:"status" db:"status"`
	ReportedBy           uuid.UUID        `json:"reported_by" db:"reported_by"`
	ReportedAt           time.Time        `json:"reported_at" db:"reported_at"`
	AssignedTo           *uuid.UUID       `json:"assigned_to,omitempty" db:"assigned_to"`
	// GDPR Breach specific (72-hour notification)
	IsDataBreach              bool       `json:"is_data_breach" db:"is_data_breach"`
	DataSubjectsAffected      int        `json:"data_subjects_affected" db:"data_subjects_affected"`
	DataCategoriesAffected    []string   `json:"data_categories_affected" db:"data_categories_affected"`
	NotificationRequired      bool       `json:"notification_required" db:"notification_required"`
	DPANotifiedAt             *time.Time `json:"dpa_notified_at,omitempty" db:"dpa_notified_at"`
	DataSubjectsNotifiedAt    *time.Time `json:"data_subjects_notified_at,omitempty" db:"data_subjects_notified_at"`
	NotificationDeadline      *time.Time `json:"notification_deadline,omitempty" db:"notification_deadline"`
	// NIS2 Incident specific
	IsNIS2Reportable          bool       `json:"is_nis2_reportable" db:"is_nis2_reportable"`
	NIS2EarlyWarningAt        *time.Time `json:"nis2_early_warning_at,omitempty" db:"nis2_early_warning_at"`
	NIS2NotificationAt        *time.Time `json:"nis2_notification_at,omitempty" db:"nis2_notification_at"`
	NIS2FinalReportAt         *time.Time `json:"nis2_final_report_at,omitempty" db:"nis2_final_report_at"`
	// Impact & Resolution
	ImpactDescription         string     `json:"impact_description,omitempty" db:"impact_description"`
	FinancialImpactEUR        float64    `json:"financial_impact_eur" db:"financial_impact_eur"`
	RootCause                 string     `json:"root_cause,omitempty" db:"root_cause"`
	ContainmentActions        string     `json:"containment_actions,omitempty" db:"containment_actions"`
	RemediationActions        string     `json:"remediation_actions,omitempty" db:"remediation_actions"`
	LessonsLearned            string     `json:"lessons_learned,omitempty" db:"lessons_learned"`
	ContainedAt               *time.Time `json:"contained_at,omitempty" db:"contained_at"`
	ResolvedAt                *time.Time `json:"resolved_at,omitempty" db:"resolved_at"`
	ClosedAt                  *time.Time `json:"closed_at,omitempty" db:"closed_at"`
	LinkedControlIDs          []uuid.UUID `json:"linked_control_ids" db:"linked_control_ids"`
	LinkedRiskIDs             []uuid.UUID `json:"linked_risk_ids" db:"linked_risk_ids"`
	Tags                      []string   `json:"tags" db:"tags"`
	Metadata                  JSONB      `json:"metadata" db:"metadata"`
}

// ============================================================
// VENDOR / THIRD-PARTY
// ============================================================

type Vendor struct {
	BaseModel
	VendorRef            string     `json:"vendor_ref" db:"vendor_ref"`
	Name                 string     `json:"name" db:"name"`
	LegalName            string     `json:"legal_name,omitempty" db:"legal_name"`
	Website              string     `json:"website,omitempty" db:"website"`
	Industry             string     `json:"industry,omitempty" db:"industry"`
	CountryCode          string     `json:"country_code" db:"country_code"`
	ContactName          string     `json:"contact_name,omitempty" db:"contact_name"`
	ContactEmail         string     `json:"contact_email,omitempty" db:"contact_email"`
	ContactPhone         string     `json:"contact_phone,omitempty" db:"contact_phone"`
	Status               string     `json:"status" db:"status"` // active, under_review, approved, rejected, offboarded
	RiskTier             string     `json:"risk_tier" db:"risk_tier"` // critical, high, medium, low
	RiskScore            float64    `json:"risk_score" db:"risk_score"`
	ServiceDescription   string     `json:"service_description,omitempty" db:"service_description"`
	DataProcessing       bool       `json:"data_processing" db:"data_processing"` // processes personal data
	DataCategories       []string   `json:"data_categories" db:"data_categories"`
	ContractStartDate    *time.Time `json:"contract_start_date,omitempty" db:"contract_start_date"`
	ContractEndDate      *time.Time `json:"contract_end_date,omitempty" db:"contract_end_date"`
	ContractValue        float64    `json:"contract_value" db:"contract_value"`
	LastAssessmentDate   *time.Time `json:"last_assessment_date,omitempty" db:"last_assessment_date"`
	NextAssessmentDate   *time.Time `json:"next_assessment_date,omitempty" db:"next_assessment_date"`
	AssessmentFrequency  string     `json:"assessment_frequency" db:"assessment_frequency"`
	Certifications       []string   `json:"certifications" db:"certifications"` // ISO27001, SOC2, PCI DSS
	DPAInPlace           bool       `json:"dpa_in_place" db:"dpa_in_place"`
	DPASignedDate        *time.Time `json:"dpa_signed_date,omitempty" db:"dpa_signed_date"`
	OwnerUserID          *uuid.UUID `json:"owner_user_id,omitempty" db:"owner_user_id"`
	Tags                 []string   `json:"tags" db:"tags"`
	Metadata             JSONB      `json:"metadata" db:"metadata"`
}

// ============================================================
// ASSET
// ============================================================

type Asset struct {
	BaseModel
	AssetRef         string     `json:"asset_ref" db:"asset_ref"`
	Name             string     `json:"name" db:"name"`
	AssetType        string     `json:"asset_type" db:"asset_type"` // hardware, software, data, service, people, facility
	Category         string     `json:"category" db:"category"`
	Description      string     `json:"description,omitempty" db:"description"`
	Status           string     `json:"status" db:"status"` // active, decommissioned, planned
	Criticality      string     `json:"criticality" db:"criticality"` // critical, high, medium, low
	OwnerUserID      *uuid.UUID `json:"owner_user_id,omitempty" db:"owner_user_id"`
	CustodianUserID  *uuid.UUID `json:"custodian_user_id,omitempty" db:"custodian_user_id"`
	Location         string     `json:"location,omitempty" db:"location"`
	IPAddress        string     `json:"ip_address,omitempty" db:"ip_address"`
	Classification   string     `json:"classification" db:"classification"` // public, internal, confidential, restricted
	ProcessesPersonalData bool  `json:"processes_personal_data" db:"processes_personal_data"`
	LinkedVendorID   *uuid.UUID `json:"linked_vendor_id,omitempty" db:"linked_vendor_id"`
	Tags             []string   `json:"tags" db:"tags"`
	Metadata         JSONB      `json:"metadata" db:"metadata"`
}
