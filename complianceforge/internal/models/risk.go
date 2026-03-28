package models

import (
	"time"

	"github.com/google/uuid"
)

// ============================================================
// RISK
// ============================================================

// Risk represents an entry in the enterprise risk register.
type Risk struct {
	BaseModel
	RiskRef          string     `json:"risk_ref" db:"risk_ref"`
	Title            string     `json:"title" db:"title"`
	Description      string     `json:"description,omitempty" db:"description"`
	RiskCategoryID   uuid.UUID  `json:"risk_category_id" db:"risk_category_id"`
	RiskSource       string     `json:"risk_source" db:"risk_source"`
	RiskType         string     `json:"risk_type" db:"risk_type"`
	Status           RiskStatus `json:"status" db:"status"`
	OwnerUserID      *uuid.UUID `json:"owner_user_id,omitempty" db:"owner_user_id"`
	DelegateUserID   *uuid.UUID `json:"delegate_user_id,omitempty" db:"delegate_user_id"`
	BusinessUnitID   *uuid.UUID `json:"business_unit_id,omitempty" db:"business_unit_id"`
	RiskMatrixID     *uuid.UUID `json:"risk_matrix_id,omitempty" db:"risk_matrix_id"`
	// Inherent Risk (before controls)
	InherentLikelihood int       `json:"inherent_likelihood" db:"inherent_likelihood"`
	InherentImpact     int       `json:"inherent_impact" db:"inherent_impact"`
	InherentRiskScore  float64   `json:"inherent_risk_score" db:"inherent_risk_score"`
	InherentRiskLevel  RiskLevel `json:"inherent_risk_level" db:"inherent_risk_level"`
	// Residual Risk (after controls)
	ResidualLikelihood int       `json:"residual_likelihood" db:"residual_likelihood"`
	ResidualImpact     int       `json:"residual_impact" db:"residual_impact"`
	ResidualRiskScore  float64   `json:"residual_risk_score" db:"residual_risk_score"`
	ResidualRiskLevel  RiskLevel `json:"residual_risk_level" db:"residual_risk_level"`
	// Target Risk
	TargetLikelihood int       `json:"target_likelihood" db:"target_likelihood"`
	TargetImpact     int       `json:"target_impact" db:"target_impact"`
	TargetRiskScore  float64   `json:"target_risk_score" db:"target_risk_score"`
	TargetRiskLevel  RiskLevel `json:"target_risk_level" db:"target_risk_level"`
	// Financial & Impact
	FinancialImpactEUR float64  `json:"financial_impact_eur" db:"financial_impact_eur"`
	ImpactDescription  string   `json:"impact_description,omitempty" db:"impact_description"`
	ImpactCategories   JSONB    `json:"impact_categories" db:"impact_categories"`
	RiskVelocity       string   `json:"risk_velocity" db:"risk_velocity"`
	RiskProximity      string   `json:"risk_proximity" db:"risk_proximity"`
	// Dates
	IdentifiedDate  *time.Time `json:"identified_date,omitempty" db:"identified_date"`
	LastAssessedDate *time.Time `json:"last_assessed_date,omitempty" db:"last_assessed_date"`
	NextReviewDate  *time.Time `json:"next_review_date,omitempty" db:"next_review_date"`
	ReviewFrequency string     `json:"review_frequency" db:"review_frequency"`
	// Linkages
	LinkedRegulations []string    `json:"linked_regulations" db:"linked_regulations"`
	LinkedControlIDs  []uuid.UUID `json:"linked_control_ids" db:"linked_control_ids"`
	Tags              []string    `json:"tags" db:"tags"`
	IsEmerging        bool        `json:"is_emerging" db:"is_emerging"`
	Metadata          JSONB       `json:"metadata" db:"metadata"`
	// Populated via joins
	Category   *RiskCategory    `json:"category,omitempty" db:"-"`
	Owner      *User            `json:"owner,omitempty" db:"-"`
	Treatments []RiskTreatment  `json:"treatments,omitempty" db:"-"`
}

// CreateRiskRequest is the payload for creating a new risk.
type CreateRiskRequest struct {
	Title              string    `json:"title" validate:"required,min=5,max=500"`
	Description        string    `json:"description,omitempty"`
	RiskCategoryID     uuid.UUID `json:"risk_category_id" validate:"required"`
	RiskSource         string    `json:"risk_source" validate:"required,oneof=internal external third_party regulatory"`
	OwnerUserID        uuid.UUID `json:"owner_user_id" validate:"required"`
	InherentLikelihood int       `json:"inherent_likelihood" validate:"required,min=1,max=5"`
	InherentImpact     int       `json:"inherent_impact" validate:"required,min=1,max=5"`
	ResidualLikelihood int       `json:"residual_likelihood" validate:"min=1,max=5"`
	ResidualImpact     int       `json:"residual_impact" validate:"min=1,max=5"`
	FinancialImpactEUR float64   `json:"financial_impact_eur,omitempty"`
	RiskVelocity       string    `json:"risk_velocity,omitempty" validate:"omitempty,oneof=immediate fast moderate slow"`
	ReviewFrequency    string    `json:"review_frequency" validate:"required,oneof=monthly quarterly semi_annually annually"`
	Tags               []string  `json:"tags,omitempty"`
}

// ============================================================
// RISK CATEGORY
// ============================================================

type RiskCategory struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	OrganizationID   *uuid.UUID `json:"organization_id,omitempty" db:"organization_id"`
	Name             string     `json:"name" db:"name"`
	Code             string     `json:"code" db:"code"`
	Description      string     `json:"description,omitempty" db:"description"`
	ParentCategoryID *uuid.UUID `json:"parent_category_id,omitempty" db:"parent_category_id"`
	ColorHex         string     `json:"color_hex,omitempty" db:"color_hex"`
	Icon             string     `json:"icon,omitempty" db:"icon"`
	SortOrder        int        `json:"sort_order" db:"sort_order"`
	IsSystemDefault  bool       `json:"is_system_default" db:"is_system_default"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

// ============================================================
// RISK TREATMENT
// ============================================================

type RiskTreatment struct {
	BaseModel
	RiskID                uuid.UUID     `json:"risk_id" db:"risk_id"`
	TreatmentType         TreatmentType `json:"treatment_type" db:"treatment_type"`
	Title                 string        `json:"title" db:"title"`
	Description           string        `json:"description,omitempty" db:"description"`
	Status                string        `json:"status" db:"status"`
	Priority              string        `json:"priority" db:"priority"`
	OwnerUserID           *uuid.UUID    `json:"owner_user_id,omitempty" db:"owner_user_id"`
	StartDate             *time.Time    `json:"start_date,omitempty" db:"start_date"`
	TargetDate            *time.Time    `json:"target_date,omitempty" db:"target_date"`
	CompletedDate         *time.Time    `json:"completed_date,omitempty" db:"completed_date"`
	EstimatedCostEUR      float64       `json:"estimated_cost_eur" db:"estimated_cost_eur"`
	ActualCostEUR         float64       `json:"actual_cost_eur" db:"actual_cost_eur"`
	ExpectedRiskReduction float64       `json:"expected_risk_reduction" db:"expected_risk_reduction"`
	ProgressPercentage    int           `json:"progress_percentage" db:"progress_percentage"`
	LinkedControlIDs      []uuid.UUID   `json:"linked_control_ids" db:"linked_control_ids"`
	Notes                 string        `json:"notes,omitempty" db:"notes"`
}

// ============================================================
// RISK INDICATOR (KRI)
// ============================================================

type RiskIndicator struct {
	BaseModel
	RiskID              *uuid.UUID `json:"risk_id,omitempty" db:"risk_id"`
	Name                string     `json:"name" db:"name"`
	Description         string     `json:"description,omitempty" db:"description"`
	MetricType          string     `json:"metric_type" db:"metric_type"`
	MeasurementUnit     string     `json:"measurement_unit,omitempty" db:"measurement_unit"`
	CollectionFrequency string     `json:"collection_frequency" db:"collection_frequency"`
	DataSource          string     `json:"data_source,omitempty" db:"data_source"`
	ThresholdGreen      float64    `json:"threshold_green" db:"threshold_green"`
	ThresholdAmber      float64    `json:"threshold_amber" db:"threshold_amber"`
	ThresholdRed        float64    `json:"threshold_red" db:"threshold_red"`
	CurrentValue        float64    `json:"current_value" db:"current_value"`
	Trend               string     `json:"trend" db:"trend"`
	OwnerUserID         *uuid.UUID `json:"owner_user_id,omitempty" db:"owner_user_id"`
	LastUpdatedAt       *time.Time `json:"last_updated_at,omitempty" db:"last_updated_at"`
	IsAutomated         bool       `json:"is_automated" db:"is_automated"`
	AutomationConfig    JSONB      `json:"automation_config" db:"automation_config"`
}

// RiskHeatmapEntry is a read model for risk heatmap visualization.
type RiskHeatmapEntry struct {
	RiskID             uuid.UUID `json:"risk_id"`
	RiskRef            string    `json:"risk_ref"`
	Title              string    `json:"title"`
	CategoryName       string    `json:"category_name"`
	InherentLikelihood int       `json:"inherent_likelihood"`
	InherentImpact     int       `json:"inherent_impact"`
	ResidualLikelihood int       `json:"residual_likelihood"`
	ResidualImpact     int       `json:"residual_impact"`
	ResidualRiskLevel  RiskLevel `json:"residual_risk_level"`
	OwnerName          string    `json:"owner_name"`
	Status             string    `json:"status"`
}
