package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// ============================================================
// ROPAService
// ============================================================

// ROPAService implements business logic for GDPR Article 30 ROPA,
// processing activities, data flow mapping, DPIA triggers,
// and ROPA export management.
type ROPAService struct {
	pool *pgxpool.Pool
}

// NewROPAService creates a new ROPAService.
func NewROPAService(pool *pgxpool.Pool) *ROPAService {
	return &ROPAService{pool: pool}
}

// ============================================================
// DATA TYPES — Processing Activities
// ============================================================

// ProcessingActivity represents a GDPR Article 30 processing activity record.
type ProcessingActivity struct {
	ID                            uuid.UUID       `json:"id"`
	OrganizationID                uuid.UUID       `json:"organization_id"`
	ActivityRef                   string          `json:"activity_ref"`
	Name                          string          `json:"name"`
	Description                   string          `json:"description"`
	Purpose                       string          `json:"purpose"`
	LegalBasis                    string          `json:"legal_basis"`
	LegalBasisDetail              string          `json:"legal_basis_detail"`
	Status                        string          `json:"status"`
	Role                          string          `json:"role"`
	JointControllerDetails        string          `json:"joint_controller_details"`
	DataSubjectCategories         []string        `json:"data_subject_categories"`
	EstimatedDataSubjectsCount    *int            `json:"estimated_data_subjects_count"`
	DataCategoryIDs               []uuid.UUID     `json:"data_category_ids"`
	SpecialCategoriesProcessed    bool            `json:"special_categories_processed"`
	SpecialCategoriesLegalBasis   string          `json:"special_categories_legal_basis"`
	RecipientCategories           []string        `json:"recipient_categories"`
	RecipientVendorIDs            []uuid.UUID     `json:"recipient_vendor_ids"`
	InvolvesInternationalTransfer bool            `json:"involves_international_transfer"`
	TransferCountries             []string        `json:"transfer_countries"`
	TransferSafeguards            string          `json:"transfer_safeguards"`
	TransferSafeguardsDetail      string          `json:"transfer_safeguards_detail"`
	TIAConducted                  bool            `json:"tia_conducted"`
	TIADate                       *time.Time      `json:"tia_date"`
	TIADocumentPath               string          `json:"tia_document_path"`
	RetentionPeriodMonths         *int            `json:"retention_period_months"`
	RetentionJustification        string          `json:"retention_justification"`
	DeletionMethod                string          `json:"deletion_method"`
	DeletionResponsibleUserID     *uuid.UUID      `json:"deletion_responsible_user_id"`
	SystemIDs                     []uuid.UUID     `json:"system_ids"`
	StorageLocations              []string        `json:"storage_locations"`
	DPIARequired                  bool            `json:"dpia_required"`
	DPIAStatus                    string          `json:"dpia_status"`
	DPIADocumentPath              string          `json:"dpia_document_path"`
	DPIAConductedDate             *time.Time      `json:"dpia_conducted_date"`
	SecurityMeasures              string          `json:"security_measures"`
	LinkedControlCodes            []string        `json:"linked_control_codes"`
	RiskLevel                     string          `json:"risk_level"`
	DataStewardUserID             *uuid.UUID      `json:"data_steward_user_id"`
	Department                    string          `json:"department"`
	ProcessOwnerUserID            *uuid.UUID      `json:"process_owner_user_id"`
	LastReviewDate                *time.Time      `json:"last_review_date"`
	NextReviewDate                *time.Time      `json:"next_review_date"`
	ReviewFrequencyMonths         int             `json:"review_frequency_months"`
	Metadata                      json.RawMessage `json:"metadata"`
	CreatedAt                     time.Time       `json:"created_at"`
	UpdatedAt                     time.Time       `json:"updated_at"`
}

// DataFlowMap represents a data flow within a processing activity.
type DataFlowMap struct {
	ID                     uuid.UUID   `json:"id"`
	OrganizationID         uuid.UUID   `json:"organization_id"`
	ProcessingActivityID   uuid.UUID   `json:"processing_activity_id"`
	Name                   string      `json:"name"`
	FlowType               string      `json:"flow_type"`
	SourceType             string      `json:"source_type"`
	SourceName             string      `json:"source_name"`
	SourceEntityID         *uuid.UUID  `json:"source_entity_id"`
	DestinationType        string      `json:"destination_type"`
	DestinationName        string      `json:"destination_name"`
	DestinationEntityID    *uuid.UUID  `json:"destination_entity_id"`
	DestinationCountry     string      `json:"destination_country"`
	DataCategoryIDs        []uuid.UUID `json:"data_category_ids"`
	TransferMethod         string      `json:"transfer_method"`
	EncryptionInTransit    bool        `json:"encryption_in_transit"`
	EncryptionAtRest       bool        `json:"encryption_at_rest"`
	VolumeDescription      string      `json:"volume_description"`
	Frequency              string      `json:"frequency"`
	LegalBasis             string      `json:"legal_basis"`
	Notes                  string      `json:"notes"`
	SortOrder              int         `json:"sort_order"`
	CreatedAt              time.Time   `json:"created_at"`
	UpdatedAt              time.Time   `json:"updated_at"`
}

// ROPAExport represents a ROPA export record.
type ROPAExport struct {
	ID                 uuid.UUID `json:"id"`
	OrganizationID     uuid.UUID `json:"organization_id"`
	ExportRef          string    `json:"export_ref"`
	ExportDate         time.Time `json:"export_date"`
	Format             string    `json:"format"`
	FilePath           string    `json:"file_path"`
	ActivitiesIncluded int       `json:"activities_included"`
	ExportedBy         *uuid.UUID `json:"exported_by"`
	ExportReason       string    `json:"export_reason"`
	Notes              string    `json:"notes"`
	CreatedAt          time.Time `json:"created_at"`
}

// ============================================================
// DASHBOARD & ANALYSIS TYPES
// ============================================================

// ROPADashboard holds aggregated ROPA metrics.
type ROPADashboard struct {
	TotalActivities            int            `json:"total_activities"`
	ActiveActivities           int            `json:"active_activities"`
	DraftActivities            int            `json:"draft_activities"`
	ByLegalBasis               map[string]int `json:"by_legal_basis"`
	ByDepartment               map[string]int `json:"by_department"`
	SpecialCategoriesCount     int            `json:"special_categories_count"`
	InternationalTransfers     int            `json:"international_transfers"`
	DPIARequired               int            `json:"dpia_required"`
	DPIACompleted              int            `json:"dpia_completed"`
	DPIAInProgress             int            `json:"dpia_in_progress"`
	DPIAPending                int            `json:"dpia_pending"`
	OverdueReviews             int            `json:"overdue_reviews"`
	UpcomingReviews30Days      int            `json:"upcoming_reviews_30_days"`
	TotalDataFlows             int            `json:"total_data_flows"`
	HighRiskActivities         int            `json:"high_risk_activities"`
	ByRole                     map[string]int `json:"by_role"`
	LastExportDate             *time.Time     `json:"last_export_date"`
}

// HighRiskActivity identifies a processing activity that triggers DPIA.
type HighRiskActivity struct {
	ActivityID     uuid.UUID `json:"activity_id"`
	ActivityRef    string    `json:"activity_ref"`
	Name           string    `json:"name"`
	Department     string    `json:"department"`
	RiskLevel      string    `json:"risk_level"`
	Reasons        []string  `json:"reasons"`
	DPIAStatus     string    `json:"dpia_status"`
	DPIARequired   bool      `json:"dpia_required"`
}

// SubjectImpactMap describes how a particular data subject category
// is affected across all processing activities.
type SubjectImpactMap struct {
	SubjectCategory    string                `json:"subject_category"`
	TotalActivities    int                   `json:"total_activities"`
	Activities         []SubjectActivityLink `json:"activities"`
	LegalBases         map[string]int        `json:"legal_bases"`
	DataCategories     []string              `json:"data_categories"`
	InternationalFlows int                   `json:"international_flows"`
}

// SubjectActivityLink links a data subject to a processing activity.
type SubjectActivityLink struct {
	ActivityID   uuid.UUID `json:"activity_id"`
	ActivityRef  string    `json:"activity_ref"`
	Name         string    `json:"name"`
	Purpose      string    `json:"purpose"`
	LegalBasis   string    `json:"legal_basis"`
	Department   string    `json:"department"`
}

// TransferEntry represents a single international transfer for the transfer register.
type TransferEntry struct {
	ActivityID           uuid.UUID `json:"activity_id"`
	ActivityRef          string    `json:"activity_ref"`
	ActivityName         string    `json:"activity_name"`
	TransferCountries    []string  `json:"transfer_countries"`
	Safeguards           string    `json:"safeguards"`
	SafeguardsDetail     string    `json:"safeguards_detail"`
	TIAConducted         bool      `json:"tia_conducted"`
	TIADate              *time.Time `json:"tia_date"`
	LegalBasis           string    `json:"legal_basis"`
	Department           string    `json:"department"`
	Status               string    `json:"status"`
}

// ActivityListResult holds a paginated list of processing activities.
type ActivityListResult struct {
	Activities []ProcessingActivity `json:"activities"`
	Total      int64                `json:"total"`
}

// ============================================================
// REQUEST TYPES
// ============================================================

// CreateActivityRequest is the request body for creating a processing activity.
type CreateActivityRequest struct {
	Name                          string      `json:"name"`
	Description                   string      `json:"description"`
	Purpose                       string      `json:"purpose"`
	LegalBasis                    string      `json:"legal_basis"`
	LegalBasisDetail              string      `json:"legal_basis_detail"`
	Role                          string      `json:"role"`
	JointControllerDetails        string      `json:"joint_controller_details"`
	DataSubjectCategories         []string    `json:"data_subject_categories"`
	EstimatedDataSubjectsCount    *int        `json:"estimated_data_subjects_count"`
	DataCategoryIDs               []uuid.UUID `json:"data_category_ids"`
	SpecialCategoriesProcessed    bool        `json:"special_categories_processed"`
	SpecialCategoriesLegalBasis   string      `json:"special_categories_legal_basis"`
	RecipientCategories           []string    `json:"recipient_categories"`
	RecipientVendorIDs            []uuid.UUID `json:"recipient_vendor_ids"`
	InvolvesInternationalTransfer bool        `json:"involves_international_transfer"`
	TransferCountries             []string    `json:"transfer_countries"`
	TransferSafeguards            string      `json:"transfer_safeguards"`
	TransferSafeguardsDetail      string      `json:"transfer_safeguards_detail"`
	RetentionPeriodMonths         *int        `json:"retention_period_months"`
	RetentionJustification        string      `json:"retention_justification"`
	DeletionMethod                string      `json:"deletion_method"`
	DeletionResponsibleUserID     *uuid.UUID  `json:"deletion_responsible_user_id"`
	SystemIDs                     []uuid.UUID `json:"system_ids"`
	StorageLocations              []string    `json:"storage_locations"`
	SecurityMeasures              string      `json:"security_measures"`
	LinkedControlCodes            []string    `json:"linked_control_codes"`
	DataStewardUserID             *uuid.UUID  `json:"data_steward_user_id"`
	Department                    string      `json:"department"`
	ProcessOwnerUserID            *uuid.UUID  `json:"process_owner_user_id"`
	ReviewFrequencyMonths         int         `json:"review_frequency_months"`
}

// UpdateActivityRequest is the request body for updating a processing activity.
type UpdateActivityRequest struct {
	Name                          *string     `json:"name"`
	Description                   *string     `json:"description"`
	Purpose                       *string     `json:"purpose"`
	LegalBasis                    *string     `json:"legal_basis"`
	LegalBasisDetail              *string     `json:"legal_basis_detail"`
	Status                        *string     `json:"status"`
	Role                          *string     `json:"role"`
	JointControllerDetails        *string     `json:"joint_controller_details"`
	DataSubjectCategories         []string    `json:"data_subject_categories"`
	EstimatedDataSubjectsCount    *int        `json:"estimated_data_subjects_count"`
	DataCategoryIDs               []uuid.UUID `json:"data_category_ids"`
	SpecialCategoriesProcessed    *bool       `json:"special_categories_processed"`
	SpecialCategoriesLegalBasis   *string     `json:"special_categories_legal_basis"`
	RecipientCategories           []string    `json:"recipient_categories"`
	RecipientVendorIDs            []uuid.UUID `json:"recipient_vendor_ids"`
	InvolvesInternationalTransfer *bool       `json:"involves_international_transfer"`
	TransferCountries             []string    `json:"transfer_countries"`
	TransferSafeguards            *string     `json:"transfer_safeguards"`
	TransferSafeguardsDetail      *string     `json:"transfer_safeguards_detail"`
	TIAConducted                  *bool       `json:"tia_conducted"`
	TIADate                       *time.Time  `json:"tia_date"`
	TIADocumentPath               *string     `json:"tia_document_path"`
	RetentionPeriodMonths         *int        `json:"retention_period_months"`
	RetentionJustification        *string     `json:"retention_justification"`
	DeletionMethod                *string     `json:"deletion_method"`
	DeletionResponsibleUserID     *uuid.UUID  `json:"deletion_responsible_user_id"`
	SystemIDs                     []uuid.UUID `json:"system_ids"`
	StorageLocations              []string    `json:"storage_locations"`
	DPIARequired                  *bool       `json:"dpia_required"`
	DPIAStatus                    *string     `json:"dpia_status"`
	DPIADocumentPath              *string     `json:"dpia_document_path"`
	DPIAConductedDate             *time.Time  `json:"dpia_conducted_date"`
	SecurityMeasures              *string     `json:"security_measures"`
	LinkedControlCodes            []string    `json:"linked_control_codes"`
	RiskLevel                     *string     `json:"risk_level"`
	DataStewardUserID             *uuid.UUID  `json:"data_steward_user_id"`
	Department                    *string     `json:"department"`
	ProcessOwnerUserID            *uuid.UUID  `json:"process_owner_user_id"`
	LastReviewDate                *time.Time  `json:"last_review_date"`
	NextReviewDate                *time.Time  `json:"next_review_date"`
	ReviewFrequencyMonths         *int        `json:"review_frequency_months"`
}

// DataFlowInput is the request body for creating a data flow map entry.
type DataFlowInput struct {
	Name                string      `json:"name"`
	FlowType            string      `json:"flow_type"`
	SourceType          string      `json:"source_type"`
	SourceName          string      `json:"source_name"`
	SourceEntityID      *uuid.UUID  `json:"source_entity_id"`
	DestinationType     string      `json:"destination_type"`
	DestinationName     string      `json:"destination_name"`
	DestinationEntityID *uuid.UUID  `json:"destination_entity_id"`
	DestinationCountry  string      `json:"destination_country"`
	DataCategoryIDs     []uuid.UUID `json:"data_category_ids"`
	TransferMethod      string      `json:"transfer_method"`
	EncryptionInTransit bool        `json:"encryption_in_transit"`
	EncryptionAtRest    bool        `json:"encryption_at_rest"`
	VolumeDescription   string      `json:"volume_description"`
	Frequency           string      `json:"frequency"`
	LegalBasis          string      `json:"legal_basis"`
	Notes               string      `json:"notes"`
	SortOrder           int         `json:"sort_order"`
}

// ActivityFilter holds filter parameters for listing processing activities.
type ActivityFilter struct {
	Status     string `json:"status"`
	LegalBasis string `json:"legal_basis"`
	Department string `json:"department"`
	Search     string `json:"search"`
	Page       int    `json:"page"`
	PageSize   int    `json:"page_size"`
}

// ============================================================
// PROCESSING ACTIVITY CRUD
// ============================================================

// CreateProcessingActivity creates a new processing activity with auto-ref PA-NNN.
// It also auto-assesses whether DPIA is required based on high-risk indicators.
func (s *ROPAService) CreateProcessingActivity(ctx context.Context, orgID uuid.UUID, req CreateActivityRequest) (*ProcessingActivity, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("activity name is required")
	}
	if req.Role == "" {
		req.Role = "controller"
	}
	if req.ReviewFrequencyMonths <= 0 {
		req.ReviewFrequencyMonths = 12
	}

	// Auto-assess DPIA requirement
	dpiaRequired := assessDPIARequired(req)
	dpiaStatus := "not_required"
	riskLevel := "medium"
	if dpiaRequired {
		dpiaStatus = "required"
		riskLevel = "high"
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_org', $1, true)", orgID.String()); err != nil {
		return nil, fmt.Errorf("failed to set RLS context: %w", err)
	}

	now := time.Now()
	nextReview := now.AddDate(0, req.ReviewFrequencyMonths, 0)

	pa := &ProcessingActivity{}
	err = tx.QueryRow(ctx, `
		INSERT INTO processing_activities (
			organization_id, activity_ref, name, description, purpose,
			legal_basis, legal_basis_detail, status, role, joint_controller_details,
			data_subject_categories, estimated_data_subjects_count,
			data_category_ids, special_categories_processed, special_categories_legal_basis,
			recipient_categories, recipient_vendor_ids,
			involves_international_transfer, transfer_countries,
			transfer_safeguards, transfer_safeguards_detail,
			retention_period_months, retention_justification,
			deletion_method, deletion_responsible_user_id,
			system_ids, storage_locations,
			dpia_required, dpia_status, security_measures,
			linked_control_codes, risk_level,
			data_steward_user_id, department, process_owner_user_id,
			last_review_date, next_review_date, review_frequency_months
		) VALUES (
			$1, generate_pa_ref($1), $2, $3, $4,
			$5::processing_legal_basis, $6, 'draft'::processing_status, $7::processing_role, $8,
			$9, $10,
			$11, $12, $13,
			$14, $15,
			$16, $17,
			$18::transfer_safeguard_type, $19,
			$20, $21,
			$22, $23,
			$24, $25,
			$26, $27::dpia_status_type, $28,
			$29, $30,
			$31, $32, $33,
			$34, $35, $36
		)
		RETURNING id, organization_id, activity_ref, name,
			COALESCE(description, ''), COALESCE(purpose, ''),
			COALESCE(legal_basis::TEXT, ''), COALESCE(legal_basis_detail, ''),
			status::TEXT, role::TEXT, COALESCE(joint_controller_details, ''),
			COALESCE(data_subject_categories, '{}'), estimated_data_subjects_count,
			COALESCE(data_category_ids, '{}'), special_categories_processed,
			COALESCE(special_categories_legal_basis, ''),
			COALESCE(recipient_categories, '{}'), COALESCE(recipient_vendor_ids, '{}'),
			involves_international_transfer, COALESCE(transfer_countries, '{}'),
			COALESCE(transfer_safeguards::TEXT, ''), COALESCE(transfer_safeguards_detail, ''),
			tia_conducted, tia_date, COALESCE(tia_document_path, ''),
			retention_period_months, COALESCE(retention_justification, ''),
			COALESCE(deletion_method, ''), deletion_responsible_user_id,
			COALESCE(system_ids, '{}'), COALESCE(storage_locations, '{}'),
			dpia_required, COALESCE(dpia_status::TEXT, ''),
			COALESCE(dpia_document_path, ''), dpia_conducted_date,
			COALESCE(security_measures, ''), COALESCE(linked_control_codes, '{}'),
			COALESCE(risk_level, ''),
			data_steward_user_id, COALESCE(department, ''), process_owner_user_id,
			last_review_date, next_review_date, review_frequency_months,
			COALESCE(metadata, '{}'::jsonb),
			created_at, updated_at`,
		orgID, req.Name, req.Description, req.Purpose,
		nilIfEmpty(req.LegalBasis), req.LegalBasisDetail, req.Role, req.JointControllerDetails,
		req.DataSubjectCategories, req.EstimatedDataSubjectsCount,
		req.DataCategoryIDs, req.SpecialCategoriesProcessed, req.SpecialCategoriesLegalBasis,
		req.RecipientCategories, req.RecipientVendorIDs,
		req.InvolvesInternationalTransfer, req.TransferCountries,
		nilIfEmpty(req.TransferSafeguards), req.TransferSafeguardsDetail,
		req.RetentionPeriodMonths, req.RetentionJustification,
		req.DeletionMethod, req.DeletionResponsibleUserID,
		req.SystemIDs, req.StorageLocations,
		dpiaRequired, dpiaStatus, req.SecurityMeasures,
		req.LinkedControlCodes, riskLevel,
		req.DataStewardUserID, req.Department, req.ProcessOwnerUserID,
		now, nextReview, req.ReviewFrequencyMonths,
	).Scan(
		&pa.ID, &pa.OrganizationID, &pa.ActivityRef, &pa.Name,
		&pa.Description, &pa.Purpose,
		&pa.LegalBasis, &pa.LegalBasisDetail,
		&pa.Status, &pa.Role, &pa.JointControllerDetails,
		&pa.DataSubjectCategories, &pa.EstimatedDataSubjectsCount,
		&pa.DataCategoryIDs, &pa.SpecialCategoriesProcessed,
		&pa.SpecialCategoriesLegalBasis,
		&pa.RecipientCategories, &pa.RecipientVendorIDs,
		&pa.InvolvesInternationalTransfer, &pa.TransferCountries,
		&pa.TransferSafeguards, &pa.TransferSafeguardsDetail,
		&pa.TIAConducted, &pa.TIADate, &pa.TIADocumentPath,
		&pa.RetentionPeriodMonths, &pa.RetentionJustification,
		&pa.DeletionMethod, &pa.DeletionResponsibleUserID,
		&pa.SystemIDs, &pa.StorageLocations,
		&pa.DPIARequired, &pa.DPIAStatus,
		&pa.DPIADocumentPath, &pa.DPIAConductedDate,
		&pa.SecurityMeasures, &pa.LinkedControlCodes,
		&pa.RiskLevel,
		&pa.DataStewardUserID, &pa.Department, &pa.ProcessOwnerUserID,
		&pa.LastReviewDate, &pa.NextReviewDate, &pa.ReviewFrequencyMonths,
		&pa.Metadata,
		&pa.CreatedAt, &pa.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create processing activity: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}

	log.Info().
		Str("org_id", orgID.String()).
		Str("activity_ref", pa.ActivityRef).
		Str("name", pa.Name).
		Bool("dpia_required", dpiaRequired).
		Msg("Processing activity created")

	return pa, nil
}

// nilIfEmpty returns nil if the string is empty, otherwise returns a pointer.
func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// assessDPIARequired evaluates whether a processing activity triggers DPIA
// based on GDPR Article 35 high-risk indicators.
func assessDPIARequired(req CreateActivityRequest) bool {
	riskIndicators := 0

	// Special categories or criminal data
	if req.SpecialCategoriesProcessed {
		riskIndicators++
	}

	// Large-scale processing
	if req.EstimatedDataSubjectsCount != nil && *req.EstimatedDataSubjectsCount > 10000 {
		riskIndicators++
	}

	// Systematic monitoring
	for _, cat := range req.DataSubjectCategories {
		if cat == "employees" || cat == "children" || cat == "vulnerable_persons" {
			riskIndicators++
			break
		}
	}

	// International transfers without adequacy
	if req.InvolvesInternationalTransfer {
		riskIndicators++
	}

	// Innovative technology or new processing
	if len(req.DataCategoryIDs) > 5 {
		riskIndicators++
	}

	// GDPR Art.35(3): DPIA required when 2+ high-risk indicators present
	return riskIndicators >= 2
}

// GetProcessingActivity retrieves a single processing activity by ID.
func (s *ROPAService) GetProcessingActivity(ctx context.Context, orgID, actID uuid.UUID) (*ProcessingActivity, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_org', $1, true)", orgID.String()); err != nil {
		return nil, fmt.Errorf("failed to set RLS context: %w", err)
	}

	pa := &ProcessingActivity{}
	err = tx.QueryRow(ctx, `
		SELECT id, organization_id, activity_ref, name,
			COALESCE(description, ''), COALESCE(purpose, ''),
			COALESCE(legal_basis::TEXT, ''), COALESCE(legal_basis_detail, ''),
			status::TEXT, role::TEXT, COALESCE(joint_controller_details, ''),
			COALESCE(data_subject_categories, '{}'), estimated_data_subjects_count,
			COALESCE(data_category_ids, '{}'), special_categories_processed,
			COALESCE(special_categories_legal_basis, ''),
			COALESCE(recipient_categories, '{}'), COALESCE(recipient_vendor_ids, '{}'),
			involves_international_transfer, COALESCE(transfer_countries, '{}'),
			COALESCE(transfer_safeguards::TEXT, ''), COALESCE(transfer_safeguards_detail, ''),
			tia_conducted, tia_date, COALESCE(tia_document_path, ''),
			retention_period_months, COALESCE(retention_justification, ''),
			COALESCE(deletion_method, ''), deletion_responsible_user_id,
			COALESCE(system_ids, '{}'), COALESCE(storage_locations, '{}'),
			dpia_required, COALESCE(dpia_status::TEXT, ''),
			COALESCE(dpia_document_path, ''), dpia_conducted_date,
			COALESCE(security_measures, ''), COALESCE(linked_control_codes, '{}'),
			COALESCE(risk_level, ''),
			data_steward_user_id, COALESCE(department, ''), process_owner_user_id,
			last_review_date, next_review_date, review_frequency_months,
			COALESCE(metadata, '{}'::jsonb),
			created_at, updated_at
		FROM processing_activities
		WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL`,
		actID, orgID,
	).Scan(
		&pa.ID, &pa.OrganizationID, &pa.ActivityRef, &pa.Name,
		&pa.Description, &pa.Purpose,
		&pa.LegalBasis, &pa.LegalBasisDetail,
		&pa.Status, &pa.Role, &pa.JointControllerDetails,
		&pa.DataSubjectCategories, &pa.EstimatedDataSubjectsCount,
		&pa.DataCategoryIDs, &pa.SpecialCategoriesProcessed,
		&pa.SpecialCategoriesLegalBasis,
		&pa.RecipientCategories, &pa.RecipientVendorIDs,
		&pa.InvolvesInternationalTransfer, &pa.TransferCountries,
		&pa.TransferSafeguards, &pa.TransferSafeguardsDetail,
		&pa.TIAConducted, &pa.TIADate, &pa.TIADocumentPath,
		&pa.RetentionPeriodMonths, &pa.RetentionJustification,
		&pa.DeletionMethod, &pa.DeletionResponsibleUserID,
		&pa.SystemIDs, &pa.StorageLocations,
		&pa.DPIARequired, &pa.DPIAStatus,
		&pa.DPIADocumentPath, &pa.DPIAConductedDate,
		&pa.SecurityMeasures, &pa.LinkedControlCodes,
		&pa.RiskLevel,
		&pa.DataStewardUserID, &pa.Department, &pa.ProcessOwnerUserID,
		&pa.LastReviewDate, &pa.NextReviewDate, &pa.ReviewFrequencyMonths,
		&pa.Metadata,
		&pa.CreatedAt, &pa.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("processing activity not found: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}

	return pa, nil
}

// ListProcessingActivities returns a filtered, paginated list of processing activities.
func (s *ROPAService) ListProcessingActivities(ctx context.Context, orgID uuid.UUID, filter ActivityFilter) (*ActivityListResult, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 20
	}
	offset := (filter.Page - 1) * filter.PageSize

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_org', $1, true)", orgID.String()); err != nil {
		return nil, fmt.Errorf("failed to set RLS context: %w", err)
	}

	// Build dynamic query
	baseWhere := "organization_id = $1 AND deleted_at IS NULL"
	args := []interface{}{orgID}
	argIdx := 2

	if filter.Status != "" {
		baseWhere += fmt.Sprintf(" AND status = $%d::processing_status", argIdx)
		args = append(args, filter.Status)
		argIdx++
	}
	if filter.LegalBasis != "" {
		baseWhere += fmt.Sprintf(" AND legal_basis = $%d::processing_legal_basis", argIdx)
		args = append(args, filter.LegalBasis)
		argIdx++
	}
	if filter.Department != "" {
		baseWhere += fmt.Sprintf(" AND department = $%d", argIdx)
		args = append(args, filter.Department)
		argIdx++
	}
	if filter.Search != "" {
		baseWhere += fmt.Sprintf(" AND (name ILIKE '%%' || $%d || '%%' OR activity_ref ILIKE '%%' || $%d || '%%' OR description ILIKE '%%' || $%d || '%%')", argIdx, argIdx, argIdx)
		args = append(args, filter.Search)
		argIdx++
	}

	// Count
	var total int64
	err = tx.QueryRow(ctx, "SELECT COUNT(*) FROM processing_activities WHERE "+baseWhere, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count activities: %w", err)
	}

	// Query with pagination
	paginatedArgs := append(args, filter.PageSize, offset)
	limitIdx := argIdx
	query := fmt.Sprintf(`
		SELECT id, organization_id, activity_ref, name,
			COALESCE(description, ''), COALESCE(purpose, ''),
			COALESCE(legal_basis::TEXT, ''), COALESCE(legal_basis_detail, ''),
			status::TEXT, role::TEXT, COALESCE(joint_controller_details, ''),
			COALESCE(data_subject_categories, '{}'), estimated_data_subjects_count,
			COALESCE(data_category_ids, '{}'), special_categories_processed,
			COALESCE(special_categories_legal_basis, ''),
			COALESCE(recipient_categories, '{}'), COALESCE(recipient_vendor_ids, '{}'),
			involves_international_transfer, COALESCE(transfer_countries, '{}'),
			COALESCE(transfer_safeguards::TEXT, ''), COALESCE(transfer_safeguards_detail, ''),
			tia_conducted, tia_date, COALESCE(tia_document_path, ''),
			retention_period_months, COALESCE(retention_justification, ''),
			COALESCE(deletion_method, ''), deletion_responsible_user_id,
			COALESCE(system_ids, '{}'), COALESCE(storage_locations, '{}'),
			dpia_required, COALESCE(dpia_status::TEXT, ''),
			COALESCE(dpia_document_path, ''), dpia_conducted_date,
			COALESCE(security_measures, ''), COALESCE(linked_control_codes, '{}'),
			COALESCE(risk_level, ''),
			data_steward_user_id, COALESCE(department, ''), process_owner_user_id,
			last_review_date, next_review_date, review_frequency_months,
			COALESCE(metadata, '{}'::jsonb),
			created_at, updated_at
		FROM processing_activities
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`,
		baseWhere, limitIdx, limitIdx+1,
	)

	rows, err := tx.Query(ctx, query, paginatedArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to list activities: %w", err)
	}
	defer rows.Close()

	var activities []ProcessingActivity
	for rows.Next() {
		var pa ProcessingActivity
		if err := rows.Scan(
			&pa.ID, &pa.OrganizationID, &pa.ActivityRef, &pa.Name,
			&pa.Description, &pa.Purpose,
			&pa.LegalBasis, &pa.LegalBasisDetail,
			&pa.Status, &pa.Role, &pa.JointControllerDetails,
			&pa.DataSubjectCategories, &pa.EstimatedDataSubjectsCount,
			&pa.DataCategoryIDs, &pa.SpecialCategoriesProcessed,
			&pa.SpecialCategoriesLegalBasis,
			&pa.RecipientCategories, &pa.RecipientVendorIDs,
			&pa.InvolvesInternationalTransfer, &pa.TransferCountries,
			&pa.TransferSafeguards, &pa.TransferSafeguardsDetail,
			&pa.TIAConducted, &pa.TIADate, &pa.TIADocumentPath,
			&pa.RetentionPeriodMonths, &pa.RetentionJustification,
			&pa.DeletionMethod, &pa.DeletionResponsibleUserID,
			&pa.SystemIDs, &pa.StorageLocations,
			&pa.DPIARequired, &pa.DPIAStatus,
			&pa.DPIADocumentPath, &pa.DPIAConductedDate,
			&pa.SecurityMeasures, &pa.LinkedControlCodes,
			&pa.RiskLevel,
			&pa.DataStewardUserID, &pa.Department, &pa.ProcessOwnerUserID,
			&pa.LastReviewDate, &pa.NextReviewDate, &pa.ReviewFrequencyMonths,
			&pa.Metadata,
			&pa.CreatedAt, &pa.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan activity: %w", err)
		}
		activities = append(activities, pa)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}

	return &ActivityListResult{Activities: activities, Total: total}, nil
}

// UpdateProcessingActivity updates a processing activity.
func (s *ROPAService) UpdateProcessingActivity(ctx context.Context, orgID, actID uuid.UUID, req UpdateActivityRequest) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_org', $1, true)", orgID.String()); err != nil {
		return fmt.Errorf("failed to set RLS context: %w", err)
	}

	ct, err := tx.Exec(ctx, `
		UPDATE processing_activities SET
			name = COALESCE($3, name),
			description = COALESCE($4, description),
			purpose = COALESCE($5, purpose),
			legal_basis = COALESCE($6::processing_legal_basis, legal_basis),
			legal_basis_detail = COALESCE($7, legal_basis_detail),
			status = COALESCE($8::processing_status, status),
			role = COALESCE($9::processing_role, role),
			joint_controller_details = COALESCE($10, joint_controller_details),
			data_subject_categories = COALESCE($11, data_subject_categories),
			estimated_data_subjects_count = COALESCE($12, estimated_data_subjects_count),
			data_category_ids = COALESCE($13, data_category_ids),
			special_categories_processed = COALESCE($14, special_categories_processed),
			special_categories_legal_basis = COALESCE($15, special_categories_legal_basis),
			recipient_categories = COALESCE($16, recipient_categories),
			recipient_vendor_ids = COALESCE($17, recipient_vendor_ids),
			involves_international_transfer = COALESCE($18, involves_international_transfer),
			transfer_countries = COALESCE($19, transfer_countries),
			transfer_safeguards = COALESCE($20::transfer_safeguard_type, transfer_safeguards),
			transfer_safeguards_detail = COALESCE($21, transfer_safeguards_detail),
			tia_conducted = COALESCE($22, tia_conducted),
			tia_date = COALESCE($23, tia_date),
			tia_document_path = COALESCE($24, tia_document_path),
			retention_period_months = COALESCE($25, retention_period_months),
			retention_justification = COALESCE($26, retention_justification),
			deletion_method = COALESCE($27, deletion_method),
			deletion_responsible_user_id = COALESCE($28, deletion_responsible_user_id),
			system_ids = COALESCE($29, system_ids),
			storage_locations = COALESCE($30, storage_locations),
			dpia_required = COALESCE($31, dpia_required),
			dpia_status = COALESCE($32::dpia_status_type, dpia_status),
			dpia_document_path = COALESCE($33, dpia_document_path),
			dpia_conducted_date = COALESCE($34, dpia_conducted_date),
			security_measures = COALESCE($35, security_measures),
			linked_control_codes = COALESCE($36, linked_control_codes),
			risk_level = COALESCE($37, risk_level),
			data_steward_user_id = COALESCE($38, data_steward_user_id),
			department = COALESCE($39, department),
			process_owner_user_id = COALESCE($40, process_owner_user_id),
			last_review_date = COALESCE($41, last_review_date),
			next_review_date = COALESCE($42, next_review_date),
			review_frequency_months = COALESCE($43, review_frequency_months)
		WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL`,
		actID, orgID,
		req.Name, req.Description, req.Purpose,
		req.LegalBasis, req.LegalBasisDetail,
		req.Status, req.Role, req.JointControllerDetails,
		req.DataSubjectCategories, req.EstimatedDataSubjectsCount,
		req.DataCategoryIDs, req.SpecialCategoriesProcessed,
		req.SpecialCategoriesLegalBasis,
		req.RecipientCategories, req.RecipientVendorIDs,
		req.InvolvesInternationalTransfer, req.TransferCountries,
		req.TransferSafeguards, req.TransferSafeguardsDetail,
		req.TIAConducted, req.TIADate, req.TIADocumentPath,
		req.RetentionPeriodMonths, req.RetentionJustification,
		req.DeletionMethod, req.DeletionResponsibleUserID,
		req.SystemIDs, req.StorageLocations,
		req.DPIARequired, req.DPIAStatus,
		req.DPIADocumentPath, req.DPIAConductedDate,
		req.SecurityMeasures, req.LinkedControlCodes,
		req.RiskLevel,
		req.DataStewardUserID, req.Department, req.ProcessOwnerUserID,
		req.LastReviewDate, req.NextReviewDate, req.ReviewFrequencyMonths,
	)
	if err != nil {
		return fmt.Errorf("failed to update processing activity: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("processing activity not found")
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	log.Info().
		Str("org_id", orgID.String()).
		Str("activity_id", actID.String()).
		Msg("Processing activity updated")

	return nil
}

// ============================================================
// DATA FLOW MAPPING
// ============================================================

// MapDataFlow creates a new data flow entry for a processing activity.
func (s *ROPAService) MapDataFlow(ctx context.Context, orgID, actID uuid.UUID, flow DataFlowInput) (*DataFlowMap, error) {
	if flow.Name == "" {
		return nil, fmt.Errorf("data flow name is required")
	}
	if flow.FlowType == "" {
		flow.FlowType = "processing"
	}
	if flow.SourceType == "" {
		flow.SourceType = "internal_system"
	}
	if flow.DestinationType == "" {
		flow.DestinationType = "internal_system"
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_org', $1, true)", orgID.String()); err != nil {
		return nil, fmt.Errorf("failed to set RLS context: %w", err)
	}

	// Verify the activity exists
	var exists bool
	err = tx.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM processing_activities
			WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL
		)`, actID, orgID,
	).Scan(&exists)
	if err != nil || !exists {
		return nil, fmt.Errorf("processing activity not found")
	}

	dfm := &DataFlowMap{}
	err = tx.QueryRow(ctx, `
		INSERT INTO data_flow_maps (
			organization_id, processing_activity_id, name, flow_type,
			source_type, source_name, source_entity_id,
			destination_type, destination_name, destination_entity_id,
			destination_country, data_category_ids, transfer_method,
			encryption_in_transit, encryption_at_rest,
			volume_description, frequency, legal_basis,
			notes, sort_order
		) VALUES (
			$1, $2, $3, $4::data_flow_type,
			$5::data_flow_source_type, $6, $7,
			$8::data_flow_dest_type, $9, $10,
			$11, $12, $13,
			$14, $15,
			$16, $17, $18,
			$19, $20
		)
		RETURNING id, organization_id, processing_activity_id, name,
			flow_type::TEXT, source_type::TEXT, source_name, source_entity_id,
			destination_type::TEXT, destination_name, destination_entity_id,
			COALESCE(destination_country, ''),
			COALESCE(data_category_ids, '{}'), COALESCE(transfer_method, ''),
			encryption_in_transit, encryption_at_rest,
			COALESCE(volume_description, ''), COALESCE(frequency, ''),
			COALESCE(legal_basis, ''),
			COALESCE(notes, ''), sort_order,
			created_at, updated_at`,
		orgID, actID, flow.Name, flow.FlowType,
		flow.SourceType, flow.SourceName, flow.SourceEntityID,
		flow.DestinationType, flow.DestinationName, flow.DestinationEntityID,
		nilIfEmpty(flow.DestinationCountry), flow.DataCategoryIDs, flow.TransferMethod,
		flow.EncryptionInTransit, flow.EncryptionAtRest,
		flow.VolumeDescription, flow.Frequency, flow.LegalBasis,
		flow.Notes, flow.SortOrder,
	).Scan(
		&dfm.ID, &dfm.OrganizationID, &dfm.ProcessingActivityID, &dfm.Name,
		&dfm.FlowType, &dfm.SourceType, &dfm.SourceName, &dfm.SourceEntityID,
		&dfm.DestinationType, &dfm.DestinationName, &dfm.DestinationEntityID,
		&dfm.DestinationCountry,
		&dfm.DataCategoryIDs, &dfm.TransferMethod,
		&dfm.EncryptionInTransit, &dfm.EncryptionAtRest,
		&dfm.VolumeDescription, &dfm.Frequency,
		&dfm.LegalBasis,
		&dfm.Notes, &dfm.SortOrder,
		&dfm.CreatedAt, &dfm.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create data flow: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}

	log.Info().
		Str("org_id", orgID.String()).
		Str("activity_id", actID.String()).
		Str("flow_name", dfm.Name).
		Msg("Data flow mapped")

	return dfm, nil
}

// ListDataFlows returns all data flows for a processing activity.
func (s *ROPAService) ListDataFlows(ctx context.Context, orgID, actID uuid.UUID) ([]DataFlowMap, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_org', $1, true)", orgID.String()); err != nil {
		return nil, fmt.Errorf("failed to set RLS context: %w", err)
	}

	rows, err := tx.Query(ctx, `
		SELECT id, organization_id, processing_activity_id, name,
			flow_type::TEXT, source_type::TEXT, source_name, source_entity_id,
			destination_type::TEXT, destination_name, destination_entity_id,
			COALESCE(destination_country, ''),
			COALESCE(data_category_ids, '{}'), COALESCE(transfer_method, ''),
			encryption_in_transit, encryption_at_rest,
			COALESCE(volume_description, ''), COALESCE(frequency, ''),
			COALESCE(legal_basis, ''),
			COALESCE(notes, ''), sort_order,
			created_at, updated_at
		FROM data_flow_maps
		WHERE processing_activity_id = $1 AND organization_id = $2
		ORDER BY sort_order, created_at`,
		actID, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list data flows: %w", err)
	}
	defer rows.Close()

	var flows []DataFlowMap
	for rows.Next() {
		var f DataFlowMap
		if err := rows.Scan(
			&f.ID, &f.OrganizationID, &f.ProcessingActivityID, &f.Name,
			&f.FlowType, &f.SourceType, &f.SourceName, &f.SourceEntityID,
			&f.DestinationType, &f.DestinationName, &f.DestinationEntityID,
			&f.DestinationCountry,
			&f.DataCategoryIDs, &f.TransferMethod,
			&f.EncryptionInTransit, &f.EncryptionAtRest,
			&f.VolumeDescription, &f.Frequency,
			&f.LegalBasis,
			&f.Notes, &f.SortOrder,
			&f.CreatedAt, &f.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan data flow: %w", err)
		}
		flows = append(flows, f)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}

	return flows, nil
}

// DeleteDataFlow deletes a data flow map entry.
func (s *ROPAService) DeleteDataFlow(ctx context.Context, orgID, flowID uuid.UUID) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_org', $1, true)", orgID.String()); err != nil {
		return fmt.Errorf("failed to set RLS context: %w", err)
	}

	ct, err := tx.Exec(ctx, `
		DELETE FROM data_flow_maps
		WHERE id = $1 AND organization_id = $2`,
		flowID, orgID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete data flow: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("data flow not found")
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	log.Info().
		Str("org_id", orgID.String()).
		Str("flow_id", flowID.String()).
		Msg("Data flow deleted")

	return nil
}

// ============================================================
// ROPA EXPORT
// ============================================================

// GenerateROPA creates a ROPA export record. In production this would
// also render PDF/XLSX; here it records the export metadata.
func (s *ROPAService) GenerateROPA(ctx context.Context, orgID uuid.UUID, format string, reason string) (*ROPAExport, error) {
	if format == "" {
		format = "pdf"
	}
	if reason == "" {
		reason = "ad_hoc"
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_org', $1, true)", orgID.String()); err != nil {
		return nil, fmt.Errorf("failed to set RLS context: %w", err)
	}

	// Count active processing activities to include
	var activityCount int
	err = tx.QueryRow(ctx, `
		SELECT COUNT(*) FROM processing_activities
		WHERE organization_id = $1 AND deleted_at IS NULL
			AND status IN ('active', 'under_review')`, orgID,
	).Scan(&activityCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count activities: %w", err)
	}

	filePath := fmt.Sprintf("/exports/ropa/%s/ropa_%s.%s",
		orgID.String(), time.Now().Format("20060102_150405"), format)

	export := &ROPAExport{}
	err = tx.QueryRow(ctx, `
		INSERT INTO ropa_exports (
			organization_id, export_ref, export_date,
			format, file_path, activities_included,
			export_reason
		) VALUES (
			$1, generate_ropa_export_ref($1), now(),
			$2::ropa_export_format, $3, $4,
			$5::ropa_export_reason
		)
		RETURNING id, organization_id, export_ref, export_date,
			format::TEXT, COALESCE(file_path, ''), activities_included,
			exported_by, export_reason::TEXT, COALESCE(notes, ''),
			created_at`,
		orgID, format, filePath, activityCount, reason,
	).Scan(
		&export.ID, &export.OrganizationID, &export.ExportRef, &export.ExportDate,
		&export.Format, &export.FilePath, &export.ActivitiesIncluded,
		&export.ExportedBy, &export.ExportReason, &export.Notes,
		&export.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create ROPA export: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}

	log.Info().
		Str("org_id", orgID.String()).
		Str("export_ref", export.ExportRef).
		Str("format", format).
		Int("activities_included", activityCount).
		Msg("ROPA export generated")

	return export, nil
}

// ListROPAExports returns all ROPA export records for an organization.
func (s *ROPAService) ListROPAExports(ctx context.Context, orgID uuid.UUID) ([]ROPAExport, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_org', $1, true)", orgID.String()); err != nil {
		return nil, fmt.Errorf("failed to set RLS context: %w", err)
	}

	rows, err := tx.Query(ctx, `
		SELECT id, organization_id, export_ref, export_date,
			format::TEXT, COALESCE(file_path, ''), activities_included,
			exported_by, export_reason::TEXT, COALESCE(notes, ''),
			created_at
		FROM ropa_exports
		WHERE organization_id = $1
		ORDER BY export_date DESC`, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list ROPA exports: %w", err)
	}
	defer rows.Close()

	var exports []ROPAExport
	for rows.Next() {
		var e ROPAExport
		if err := rows.Scan(
			&e.ID, &e.OrganizationID, &e.ExportRef, &e.ExportDate,
			&e.Format, &e.FilePath, &e.ActivitiesIncluded,
			&e.ExportedBy, &e.ExportReason, &e.Notes,
			&e.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan ROPA export: %w", err)
		}
		exports = append(exports, e)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}

	return exports, nil
}

// ============================================================
// DASHBOARD & ANALYTICS
// ============================================================

// GetROPADashboard returns aggregated ROPA metrics for an organization.
func (s *ROPAService) GetROPADashboard(ctx context.Context, orgID uuid.UUID) (*ROPADashboard, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_org', $1, true)", orgID.String()); err != nil {
		return nil, fmt.Errorf("failed to set RLS context: %w", err)
	}

	dash := &ROPADashboard{
		ByLegalBasis: make(map[string]int),
		ByDepartment: make(map[string]int),
		ByRole:       make(map[string]int),
	}

	// Total and status counts
	err = tx.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE status = 'active'),
			COUNT(*) FILTER (WHERE status = 'draft'),
			COUNT(*) FILTER (WHERE special_categories_processed = true),
			COUNT(*) FILTER (WHERE involves_international_transfer = true),
			COUNT(*) FILTER (WHERE dpia_required = true),
			COUNT(*) FILTER (WHERE dpia_status = 'completed'),
			COUNT(*) FILTER (WHERE dpia_status = 'in_progress'),
			COUNT(*) FILTER (WHERE dpia_required = true AND dpia_status IN ('required', 'not_required')),
			COUNT(*) FILTER (WHERE next_review_date < CURRENT_DATE),
			COUNT(*) FILTER (WHERE next_review_date BETWEEN CURRENT_DATE AND CURRENT_DATE + INTERVAL '30 days'),
			COUNT(*) FILTER (WHERE risk_level IN ('high', 'critical'))
		FROM processing_activities
		WHERE organization_id = $1 AND deleted_at IS NULL`, orgID,
	).Scan(
		&dash.TotalActivities,
		&dash.ActiveActivities,
		&dash.DraftActivities,
		&dash.SpecialCategoriesCount,
		&dash.InternationalTransfers,
		&dash.DPIARequired,
		&dash.DPIACompleted,
		&dash.DPIAInProgress,
		&dash.DPIAPending,
		&dash.OverdueReviews,
		&dash.UpcomingReviews30Days,
		&dash.HighRiskActivities,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get dashboard counts: %w", err)
	}

	// By legal basis
	lbRows, err := tx.Query(ctx, `
		SELECT COALESCE(legal_basis::TEXT, 'unset'), COUNT(*)
		FROM processing_activities
		WHERE organization_id = $1 AND deleted_at IS NULL
		GROUP BY legal_basis`, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get legal basis breakdown: %w", err)
	}
	defer lbRows.Close()
	for lbRows.Next() {
		var basis string
		var count int
		if err := lbRows.Scan(&basis, &count); err != nil {
			return nil, fmt.Errorf("failed to scan legal basis: %w", err)
		}
		dash.ByLegalBasis[basis] = count
	}

	// By department
	deptRows, err := tx.Query(ctx, `
		SELECT COALESCE(department, 'Unassigned'), COUNT(*)
		FROM processing_activities
		WHERE organization_id = $1 AND deleted_at IS NULL
		GROUP BY department`, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get department breakdown: %w", err)
	}
	defer deptRows.Close()
	for deptRows.Next() {
		var dept string
		var count int
		if err := deptRows.Scan(&dept, &count); err != nil {
			return nil, fmt.Errorf("failed to scan department: %w", err)
		}
		dash.ByDepartment[dept] = count
	}

	// By role
	roleRows, err := tx.Query(ctx, `
		SELECT role::TEXT, COUNT(*)
		FROM processing_activities
		WHERE organization_id = $1 AND deleted_at IS NULL
		GROUP BY role`, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get role breakdown: %w", err)
	}
	defer roleRows.Close()
	for roleRows.Next() {
		var role string
		var count int
		if err := roleRows.Scan(&role, &count); err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}
		dash.ByRole[role] = count
	}

	// Total data flows
	err = tx.QueryRow(ctx, `
		SELECT COUNT(*) FROM data_flow_maps
		WHERE organization_id = $1`, orgID,
	).Scan(&dash.TotalDataFlows)
	if err != nil {
		return nil, fmt.Errorf("failed to count data flows: %w", err)
	}

	// Last export date
	err = tx.QueryRow(ctx, `
		SELECT MAX(export_date) FROM ropa_exports
		WHERE organization_id = $1`, orgID,
	).Scan(&dash.LastExportDate)
	if err != nil {
		// Non-fatal: just leave nil
		dash.LastExportDate = nil
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}

	return dash, nil
}

// IdentifyHighRiskProcessing identifies processing activities that trigger
// DPIA based on GDPR Article 35 high-risk indicators.
func (s *ROPAService) IdentifyHighRiskProcessing(ctx context.Context, orgID uuid.UUID) ([]HighRiskActivity, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_org', $1, true)", orgID.String()); err != nil {
		return nil, fmt.Errorf("failed to set RLS context: %w", err)
	}

	rows, err := tx.Query(ctx, `
		SELECT id, activity_ref, name, COALESCE(department, ''),
			COALESCE(risk_level, 'medium'),
			special_categories_processed,
			involves_international_transfer,
			estimated_data_subjects_count,
			COALESCE(data_subject_categories, '{}'),
			COALESCE(data_category_ids, '{}'),
			dpia_required, COALESCE(dpia_status::TEXT, 'not_required')
		FROM processing_activities
		WHERE organization_id = $1 AND deleted_at IS NULL
		ORDER BY
			CASE risk_level
				WHEN 'critical' THEN 1
				WHEN 'high' THEN 2
				WHEN 'medium' THEN 3
				WHEN 'low' THEN 4
				ELSE 5
			END,
			name ASC`, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query activities: %w", err)
	}
	defer rows.Close()

	var highRisk []HighRiskActivity
	for rows.Next() {
		var id uuid.UUID
		var ref, name, department, riskLevel string
		var specialCategories, intlTransfer, dpiaRequired bool
		var estimatedSubjects *int
		var subjectCategories []string
		var dataCategoryIDs []uuid.UUID
		var dpiaStatus string

		if err := rows.Scan(
			&id, &ref, &name, &department, &riskLevel,
			&specialCategories, &intlTransfer,
			&estimatedSubjects, &subjectCategories, &dataCategoryIDs,
			&dpiaRequired, &dpiaStatus,
		); err != nil {
			return nil, fmt.Errorf("failed to scan activity: %w", err)
		}

		// Evaluate risk reasons
		var reasons []string
		if specialCategories {
			reasons = append(reasons, "Processes special category data (Art. 9)")
		}
		if intlTransfer {
			reasons = append(reasons, "Involves international data transfers")
		}
		if estimatedSubjects != nil && *estimatedSubjects > 10000 {
			reasons = append(reasons, "Large-scale processing (>10,000 data subjects)")
		}
		for _, cat := range subjectCategories {
			if cat == "children" {
				reasons = append(reasons, "Processes children's data")
				break
			}
			if cat == "vulnerable_persons" {
				reasons = append(reasons, "Processes vulnerable persons' data")
				break
			}
		}
		if len(dataCategoryIDs) > 5 {
			reasons = append(reasons, "Combines multiple data categories")
		}

		// Only include activities with 2+ risk indicators
		if len(reasons) >= 2 {
			highRisk = append(highRisk, HighRiskActivity{
				ActivityID:   id,
				ActivityRef:  ref,
				Name:         name,
				Department:   department,
				RiskLevel:    riskLevel,
				Reasons:      reasons,
				DPIAStatus:   dpiaStatus,
				DPIARequired: dpiaRequired,
			})
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}

	return highRisk, nil
}

// DataSubjectImpactMap returns how a particular data subject category
// is impacted across all processing activities.
func (s *ROPAService) DataSubjectImpactMap(ctx context.Context, orgID uuid.UUID, subjectCategory string) (*SubjectImpactMap, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_org', $1, true)", orgID.String()); err != nil {
		return nil, fmt.Errorf("failed to set RLS context: %w", err)
	}

	result := &SubjectImpactMap{
		SubjectCategory: subjectCategory,
		LegalBases:      make(map[string]int),
	}

	rows, err := tx.Query(ctx, `
		SELECT id, activity_ref, name, COALESCE(purpose, ''),
			COALESCE(legal_basis::TEXT, ''), COALESCE(department, ''),
			involves_international_transfer
		FROM processing_activities
		WHERE organization_id = $1
			AND deleted_at IS NULL
			AND $2 = ANY(data_subject_categories)
		ORDER BY name`, orgID, subjectCategory,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query subject impact: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var link SubjectActivityLink
		var intlTransfer bool
		if err := rows.Scan(
			&link.ActivityID, &link.ActivityRef, &link.Name,
			&link.Purpose, &link.LegalBasis, &link.Department,
			&intlTransfer,
		); err != nil {
			return nil, fmt.Errorf("failed to scan subject activity: %w", err)
		}
		result.Activities = append(result.Activities, link)
		result.TotalActivities++
		if link.LegalBasis != "" {
			result.LegalBases[link.LegalBasis]++
		}
		if intlTransfer {
			result.InternationalFlows++
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}

	return result, nil
}

// GetTransferRegister returns all processing activities involving international
// transfers, forming the transfer register required by GDPR.
func (s *ROPAService) GetTransferRegister(ctx context.Context, orgID uuid.UUID) ([]TransferEntry, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_org', $1, true)", orgID.String()); err != nil {
		return nil, fmt.Errorf("failed to set RLS context: %w", err)
	}

	rows, err := tx.Query(ctx, `
		SELECT id, activity_ref, name,
			COALESCE(transfer_countries, '{}'),
			COALESCE(transfer_safeguards::TEXT, ''),
			COALESCE(transfer_safeguards_detail, ''),
			tia_conducted, tia_date,
			COALESCE(legal_basis::TEXT, ''),
			COALESCE(department, ''),
			status::TEXT
		FROM processing_activities
		WHERE organization_id = $1
			AND deleted_at IS NULL
			AND involves_international_transfer = true
		ORDER BY activity_ref`, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query transfer register: %w", err)
	}
	defer rows.Close()

	var entries []TransferEntry
	for rows.Next() {
		var e TransferEntry
		if err := rows.Scan(
			&e.ActivityID, &e.ActivityRef, &e.ActivityName,
			&e.TransferCountries,
			&e.Safeguards, &e.SafeguardsDetail,
			&e.TIAConducted, &e.TIADate,
			&e.LegalBasis, &e.Department, &e.Status,
		); err != nil {
			return nil, fmt.Errorf("failed to scan transfer entry: %w", err)
		}
		entries = append(entries, e)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}

	return entries, nil
}
