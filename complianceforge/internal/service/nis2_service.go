package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// NIS2Service implements business logic for NIS2 Directive compliance automation,
// covering entity classification, 3-phase incident reporting per Article 23,
// security measures tracking per Article 21, and management accountability per Article 20.
type NIS2Service struct {
	pool *pgxpool.Pool
}

// NewNIS2Service creates a new NIS2Service with the given database pool.
func NewNIS2Service(pool *pgxpool.Pool) *NIS2Service {
	return &NIS2Service{pool: pool}
}

// ============================================================
// DATA TYPES
// ============================================================

// NIS2EntityAssessment represents an entity categorisation under NIS2.
type NIS2EntityAssessment struct {
	ID                  uuid.UUID       `json:"id"`
	OrganizationID      uuid.UUID       `json:"organization_id"`
	EntityType          string          `json:"entity_type"`
	Sector              string          `json:"sector"`
	SubSector           string          `json:"sub_sector"`
	AssessmentCriteria  json.RawMessage `json:"assessment_criteria"`
	EmployeeCount       int             `json:"employee_count"`
	AnnualTurnoverEUR   float64         `json:"annual_turnover_eur"`
	AssessmentDate      time.Time       `json:"assessment_date"`
	AssessedBy          *uuid.UUID      `json:"assessed_by"`
	IsInScope           bool            `json:"is_in_scope"`
	MemberState         string          `json:"member_state"`
	CompetentAuthority  string          `json:"competent_authority"`
	CSIRTName           string          `json:"csirt_name"`
	CSIRTContactEmail   string          `json:"csirt_contact_email"`
	CSIRTReportingURL   string          `json:"csirt_reporting_url"`
	Notes               string          `json:"notes"`
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`
}

// NIS2IncidentReport represents a 3-phase NIS2 incident report.
type NIS2IncidentReport struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	IncidentID     uuid.UUID `json:"incident_id"`
	ReportRef      string    `json:"report_ref"`

	// Phase 1
	EarlyWarningStatus        string          `json:"early_warning_status"`
	EarlyWarningDeadline      *time.Time      `json:"early_warning_deadline"`
	EarlyWarningSubmittedAt   *time.Time      `json:"early_warning_submitted_at"`
	EarlyWarningSubmittedBy   *uuid.UUID      `json:"early_warning_submitted_by"`
	EarlyWarningContent       json.RawMessage `json:"early_warning_content"`
	EarlyWarningCSIRTRef      string          `json:"early_warning_csirt_reference"`

	// Phase 2
	NotificationStatus        string          `json:"notification_status"`
	NotificationDeadline      *time.Time      `json:"notification_deadline"`
	NotificationSubmittedAt   *time.Time      `json:"notification_submitted_at"`
	NotificationSubmittedBy   *uuid.UUID      `json:"notification_submitted_by"`
	NotificationContent       json.RawMessage `json:"notification_content"`
	NotificationCSIRTRef      string          `json:"notification_csirt_reference"`

	// Phase 3
	FinalReportStatus        string          `json:"final_report_status"`
	FinalReportDeadline      *time.Time      `json:"final_report_deadline"`
	FinalReportSubmittedAt   *time.Time      `json:"final_report_submitted_at"`
	FinalReportSubmittedBy   *uuid.UUID      `json:"final_report_submitted_by"`
	FinalReportContent       json.RawMessage `json:"final_report_content"`
	FinalReportDocumentPath  string          `json:"final_report_document_path"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NIS2SecurityMeasure represents a single Article 21 security measure.
type NIS2SecurityMeasure struct {
	ID                   uuid.UUID  `json:"id"`
	OrganizationID       uuid.UUID  `json:"organization_id"`
	MeasureCode          string     `json:"measure_code"`
	MeasureTitle         string     `json:"measure_title"`
	MeasureDescription   string     `json:"measure_description"`
	ArticleReference     string     `json:"article_reference"`
	ImplementationStatus string     `json:"implementation_status"`
	OwnerUserID          *uuid.UUID `json:"owner_user_id"`
	EvidenceDescription  string     `json:"evidence_description"`
	LastAssessedAt       *time.Time `json:"last_assessed_at"`
	NextAssessmentDate   *time.Time `json:"next_assessment_date"`
	LinkedControlIDs     []uuid.UUID `json:"linked_control_ids"`
	Notes                string     `json:"notes"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

// NIS2ManagementRecord represents a management accountability record.
type NIS2ManagementRecord struct {
	ID                      uuid.UUID  `json:"id"`
	OrganizationID          uuid.UUID  `json:"organization_id"`
	BoardMemberName         string     `json:"board_member_name"`
	BoardMemberRole         string     `json:"board_member_role"`
	TrainingCompleted       bool       `json:"training_completed"`
	TrainingDate            *time.Time `json:"training_date"`
	TrainingProvider        string     `json:"training_provider"`
	TrainingCertificatePath string     `json:"training_certificate_path"`
	RiskMeasuresApproved    bool       `json:"risk_measures_approved"`
	ApprovalDate            *time.Time `json:"approval_date"`
	ApprovalDocumentPath    string     `json:"approval_document_path"`
	NextTrainingDue         *time.Time `json:"next_training_due"`
	Notes                   string     `json:"notes"`
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at"`
}

// NIS2Dashboard aggregates NIS2 compliance metrics.
type NIS2Dashboard struct {
	EntityType             string                    `json:"entity_type"`
	IsInScope              bool                      `json:"is_in_scope"`
	MeasuresTotal          int                       `json:"measures_total"`
	MeasuresImplemented    int                       `json:"measures_implemented"`
	MeasuresVerified       int                       `json:"measures_verified"`
	MeasuresInProgress     int                       `json:"measures_in_progress"`
	MeasuresNotStarted     int                       `json:"measures_not_started"`
	CompliancePercentage   float64                   `json:"compliance_percentage"`
	IncidentReportsTotal   int                       `json:"incident_reports_total"`
	OverdueReports         int                       `json:"overdue_reports"`
	PendingEarlyWarnings   int                       `json:"pending_early_warnings"`
	PendingNotifications   int                       `json:"pending_notifications"`
	PendingFinalReports    int                       `json:"pending_final_reports"`
	ManagementTrained      int                       `json:"management_trained"`
	ManagementTotal        int                       `json:"management_total"`
	TrainingOverdue        int                       `json:"training_overdue"`
	MeasuresBreakdown      []MeasureStatusEntry      `json:"measures_breakdown"`
}

// MeasureStatusEntry provides per-measure status for the dashboard.
type MeasureStatusEntry struct {
	MeasureCode          string `json:"measure_code"`
	MeasureTitle         string `json:"measure_title"`
	ImplementationStatus string `json:"implementation_status"`
}

// ============================================================
// NIS2 ESSENTIAL/IMPORTANT ENTITY SECTOR LISTS
// ============================================================

// essentialSectors are sectors under Annex I of NIS2 (essential entities).
var essentialSectors = map[string]bool{
	"energy":                   true,
	"transport":                true,
	"banking":                  true,
	"financial_market":         true,
	"health":                   true,
	"drinking_water":           true,
	"waste_water":              true,
	"digital_infrastructure":   true,
	"ict_service_management":   true,
	"public_administration":    true,
	"space":                    true,
}

// importantSectors are sectors under Annex II of NIS2 (important entities).
var importantSectors = map[string]bool{
	"postal_courier":           true,
	"waste_management":         true,
	"chemicals":                true,
	"food":                     true,
	"manufacturing":            true,
	"digital_providers":        true,
	"research":                 true,
}

// ============================================================
// ENTITY ASSESSMENT
// ============================================================

// AssessEntityTypeInput holds the input parameters for entity classification.
type AssessEntityTypeInput struct {
	Sector              string          `json:"sector"`
	SubSector           string          `json:"sub_sector"`
	EmployeeCount       int             `json:"employee_count"`
	AnnualTurnoverEUR   float64         `json:"annual_turnover_eur"`
	MemberState         string          `json:"member_state"`
	CompetentAuthority  string          `json:"competent_authority"`
	CSIRTName           string          `json:"csirt_name"`
	CSIRTContactEmail   string          `json:"csirt_contact_email"`
	CSIRTReportingURL   string          `json:"csirt_reporting_url"`
	Notes               string          `json:"notes"`
}

// DetermineEntityType classifies an entity as essential, important, or not applicable
// based on NIS2 Directive criteria: sector, employee count, and annual turnover.
//
// Rules per NIS2 Article 3:
//   - Essential: Annex I sectors AND (employees >= 250 OR turnover >= 50M EUR)
//   - Important: Annex I sectors (medium, 50-249 employees, 10-50M turnover)
//                OR Annex II sectors (medium or large)
//   - Not applicable: below thresholds or not in listed sectors
func DetermineEntityType(sector string, employeeCount int, annualTurnoverEUR float64) string {
	sectorLower := strings.ToLower(strings.TrimSpace(sector))

	isEssentialSector := essentialSectors[sectorLower]
	isImportantSector := importantSectors[sectorLower]

	isLarge := employeeCount >= 250 || annualTurnoverEUR >= 50_000_000
	isMedium := (employeeCount >= 50 || annualTurnoverEUR >= 10_000_000) && !isLarge

	switch {
	case isEssentialSector && isLarge:
		return "essential"
	case isEssentialSector && isMedium:
		return "important"
	case isImportantSector && (isLarge || isMedium):
		return "important"
	default:
		return "not_applicable"
	}
}

// AssessEntityType creates or updates an entity assessment for the organisation.
func (s *NIS2Service) AssessEntityType(ctx context.Context, orgID uuid.UUID, assessedBy uuid.UUID, input AssessEntityTypeInput) (*NIS2EntityAssessment, error) {
	entityType := DetermineEntityType(input.Sector, input.EmployeeCount, input.AnnualTurnoverEUR)
	isInScope := entityType != "not_applicable"

	criteria := map[string]interface{}{
		"sector":             input.Sector,
		"sub_sector":         input.SubSector,
		"employee_count":     input.EmployeeCount,
		"annual_turnover_eur": input.AnnualTurnoverEUR,
		"determined_type":    entityType,
		"is_essential_sector": essentialSectors[strings.ToLower(input.Sector)],
		"is_important_sector": importantSectors[strings.ToLower(input.Sector)],
		"is_large_entity":   input.EmployeeCount >= 250 || input.AnnualTurnoverEUR >= 50_000_000,
		"is_medium_entity":  (input.EmployeeCount >= 50 || input.AnnualTurnoverEUR >= 10_000_000) && !(input.EmployeeCount >= 250 || input.AnnualTurnoverEUR >= 50_000_000),
		"assessed_at":       time.Now().UTC().Format(time.RFC3339),
	}
	criteriaJSON, err := json.Marshal(criteria)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal assessment criteria: %w", err)
	}

	assessment := &NIS2EntityAssessment{}
	err = s.pool.QueryRow(ctx, `
		INSERT INTO nis2_entity_assessment (
			organization_id, entity_type, sector, sub_sector, assessment_criteria,
			employee_count, annual_turnover_eur, assessment_date, assessed_by,
			is_in_scope, member_state, competent_authority,
			csirt_name, csirt_contact_email, csirt_reporting_url, notes
		) VALUES ($1, $2::nis2_entity_type, $3, $4, $5::JSONB, $6, $7, CURRENT_DATE, $8,
			$9, $10, $11, $12, $13, $14, $15)
		RETURNING id, organization_id, entity_type::TEXT, sector, sub_sector,
			assessment_criteria, employee_count, annual_turnover_eur,
			assessment_date, assessed_by, is_in_scope, member_state,
			competent_authority, csirt_name, csirt_contact_email,
			csirt_reporting_url, notes, created_at, updated_at`,
		orgID, entityType, input.Sector, input.SubSector, criteriaJSON,
		input.EmployeeCount, input.AnnualTurnoverEUR, assessedBy,
		isInScope, input.MemberState, input.CompetentAuthority,
		input.CSIRTName, input.CSIRTContactEmail, input.CSIRTReportingURL, input.Notes,
	).Scan(
		&assessment.ID, &assessment.OrganizationID, &assessment.EntityType,
		&assessment.Sector, &assessment.SubSector, &assessment.AssessmentCriteria,
		&assessment.EmployeeCount, &assessment.AnnualTurnoverEUR,
		&assessment.AssessmentDate, &assessment.AssessedBy, &assessment.IsInScope,
		&assessment.MemberState, &assessment.CompetentAuthority,
		&assessment.CSIRTName, &assessment.CSIRTContactEmail,
		&assessment.CSIRTReportingURL, &assessment.Notes,
		&assessment.CreatedAt, &assessment.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create entity assessment: %w", err)
	}

	log.Info().
		Str("org_id", orgID.String()).
		Str("entity_type", entityType).
		Str("sector", input.Sector).
		Bool("in_scope", isInScope).
		Msg("NIS2 entity assessment completed")

	return assessment, nil
}

// GetEntityAssessment returns the most recent entity assessment for the organisation.
func (s *NIS2Service) GetEntityAssessment(ctx context.Context, orgID uuid.UUID) (*NIS2EntityAssessment, error) {
	a := &NIS2EntityAssessment{}
	err := s.pool.QueryRow(ctx, `
		SELECT id, organization_id, entity_type::TEXT, sector, sub_sector,
			assessment_criteria, employee_count, annual_turnover_eur,
			assessment_date, assessed_by, is_in_scope, member_state,
			competent_authority, csirt_name, csirt_contact_email,
			csirt_reporting_url, COALESCE(notes, ''), created_at, updated_at
		FROM nis2_entity_assessment
		WHERE organization_id = $1
		ORDER BY assessment_date DESC, created_at DESC
		LIMIT 1`, orgID,
	).Scan(
		&a.ID, &a.OrganizationID, &a.EntityType, &a.Sector, &a.SubSector,
		&a.AssessmentCriteria, &a.EmployeeCount, &a.AnnualTurnoverEUR,
		&a.AssessmentDate, &a.AssessedBy, &a.IsInScope, &a.MemberState,
		&a.CompetentAuthority, &a.CSIRTName, &a.CSIRTContactEmail,
		&a.CSIRTReportingURL, &a.Notes, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get entity assessment: %w", err)
	}
	return a, nil
}

// ============================================================
// INCIDENT REPORTING (3-PHASE)
// ============================================================

// CalculateDeadlines computes the 3-phase NIS2 reporting deadlines from a detection time.
//   - Phase 1 Early Warning: detection + 24 hours
//   - Phase 2 Notification: detection + 72 hours
//   - Phase 3 Final Report: notification deadline + 1 month
func CalculateDeadlines(detectionTime time.Time) (earlyWarning, notification, finalReport time.Time) {
	earlyWarning = detectionTime.Add(24 * time.Hour)
	notification = detectionTime.Add(72 * time.Hour)
	finalReport = notification.AddDate(0, 1, 0)
	return
}

// CreateIncidentReport creates a new NIS2 3-phase incident report linked to an existing incident.
// Deadlines are automatically calculated based on the incident's reported_at timestamp.
func (s *NIS2Service) CreateIncidentReport(ctx context.Context, orgID, incidentID uuid.UUID) (*NIS2IncidentReport, error) {
	// Get the incident's reported_at time for deadline calculation
	var reportedAt time.Time
	err := s.pool.QueryRow(ctx, `
		SELECT reported_at FROM incidents
		WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL`,
		incidentID, orgID,
	).Scan(&reportedAt)
	if err != nil {
		return nil, fmt.Errorf("incident not found: %w", err)
	}

	ewDeadline, notifDeadline, finalDeadline := CalculateDeadlines(reportedAt)

	report := &NIS2IncidentReport{}
	err = s.pool.QueryRow(ctx, `
		INSERT INTO nis2_incident_reports (
			organization_id, incident_id, report_ref,
			early_warning_status, early_warning_deadline,
			notification_status, notification_deadline,
			final_report_status, final_report_deadline
		) VALUES (
			$1, $2, generate_nis2_report_ref($1),
			'pending', $3, 'pending', $4, 'pending', $5
		)
		RETURNING id, organization_id, incident_id, report_ref,
			early_warning_status::TEXT, early_warning_deadline,
			early_warning_submitted_at, early_warning_submitted_by,
			early_warning_content, COALESCE(early_warning_csirt_reference, ''),
			notification_status::TEXT, notification_deadline,
			notification_submitted_at, notification_submitted_by,
			notification_content, COALESCE(notification_csirt_reference, ''),
			final_report_status::TEXT, final_report_deadline,
			final_report_submitted_at, final_report_submitted_by,
			final_report_content, COALESCE(final_report_document_path, ''),
			created_at, updated_at`,
		orgID, incidentID, ewDeadline, notifDeadline, finalDeadline,
	).Scan(
		&report.ID, &report.OrganizationID, &report.IncidentID, &report.ReportRef,
		&report.EarlyWarningStatus, &report.EarlyWarningDeadline,
		&report.EarlyWarningSubmittedAt, &report.EarlyWarningSubmittedBy,
		&report.EarlyWarningContent, &report.EarlyWarningCSIRTRef,
		&report.NotificationStatus, &report.NotificationDeadline,
		&report.NotificationSubmittedAt, &report.NotificationSubmittedBy,
		&report.NotificationContent, &report.NotificationCSIRTRef,
		&report.FinalReportStatus, &report.FinalReportDeadline,
		&report.FinalReportSubmittedAt, &report.FinalReportSubmittedBy,
		&report.FinalReportContent, &report.FinalReportDocumentPath,
		&report.CreatedAt, &report.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create NIS2 incident report: %w", err)
	}

	log.Info().
		Str("org_id", orgID.String()).
		Str("incident_id", incidentID.String()).
		Str("report_ref", report.ReportRef).
		Time("ew_deadline", ewDeadline).
		Time("notif_deadline", notifDeadline).
		Time("final_deadline", finalDeadline).
		Msg("NIS2 incident report created with 3-phase deadlines")

	return report, nil
}

// GetIncidentReport retrieves a single NIS2 incident report by ID.
func (s *NIS2Service) GetIncidentReport(ctx context.Context, orgID, reportID uuid.UUID) (*NIS2IncidentReport, error) {
	report := &NIS2IncidentReport{}
	err := s.pool.QueryRow(ctx, `
		SELECT id, organization_id, incident_id, report_ref,
			early_warning_status::TEXT, early_warning_deadline,
			early_warning_submitted_at, early_warning_submitted_by,
			early_warning_content, COALESCE(early_warning_csirt_reference, ''),
			notification_status::TEXT, notification_deadline,
			notification_submitted_at, notification_submitted_by,
			notification_content, COALESCE(notification_csirt_reference, ''),
			final_report_status::TEXT, final_report_deadline,
			final_report_submitted_at, final_report_submitted_by,
			final_report_content, COALESCE(final_report_document_path, ''),
			created_at, updated_at
		FROM nis2_incident_reports
		WHERE id = $1 AND organization_id = $2`, reportID, orgID,
	).Scan(
		&report.ID, &report.OrganizationID, &report.IncidentID, &report.ReportRef,
		&report.EarlyWarningStatus, &report.EarlyWarningDeadline,
		&report.EarlyWarningSubmittedAt, &report.EarlyWarningSubmittedBy,
		&report.EarlyWarningContent, &report.EarlyWarningCSIRTRef,
		&report.NotificationStatus, &report.NotificationDeadline,
		&report.NotificationSubmittedAt, &report.NotificationSubmittedBy,
		&report.NotificationContent, &report.NotificationCSIRTRef,
		&report.FinalReportStatus, &report.FinalReportDeadline,
		&report.FinalReportSubmittedAt, &report.FinalReportSubmittedBy,
		&report.FinalReportContent, &report.FinalReportDocumentPath,
		&report.CreatedAt, &report.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("NIS2 incident report not found: %w", err)
	}
	return report, nil
}

// ListIncidentReports returns all NIS2 incident reports for an organisation.
func (s *NIS2Service) ListIncidentReports(ctx context.Context, orgID uuid.UUID) ([]NIS2IncidentReport, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, organization_id, incident_id, report_ref,
			early_warning_status::TEXT, early_warning_deadline,
			early_warning_submitted_at, early_warning_submitted_by,
			early_warning_content, COALESCE(early_warning_csirt_reference, ''),
			notification_status::TEXT, notification_deadline,
			notification_submitted_at, notification_submitted_by,
			notification_content, COALESCE(notification_csirt_reference, ''),
			final_report_status::TEXT, final_report_deadline,
			final_report_submitted_at, final_report_submitted_by,
			final_report_content, COALESCE(final_report_document_path, ''),
			created_at, updated_at
		FROM nis2_incident_reports
		WHERE organization_id = $1
		ORDER BY created_at DESC`, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list NIS2 incident reports: %w", err)
	}
	defer rows.Close()

	var reports []NIS2IncidentReport
	for rows.Next() {
		var r NIS2IncidentReport
		if err := rows.Scan(
			&r.ID, &r.OrganizationID, &r.IncidentID, &r.ReportRef,
			&r.EarlyWarningStatus, &r.EarlyWarningDeadline,
			&r.EarlyWarningSubmittedAt, &r.EarlyWarningSubmittedBy,
			&r.EarlyWarningContent, &r.EarlyWarningCSIRTRef,
			&r.NotificationStatus, &r.NotificationDeadline,
			&r.NotificationSubmittedAt, &r.NotificationSubmittedBy,
			&r.NotificationContent, &r.NotificationCSIRTRef,
			&r.FinalReportStatus, &r.FinalReportDeadline,
			&r.FinalReportSubmittedAt, &r.FinalReportSubmittedBy,
			&r.FinalReportContent, &r.FinalReportDocumentPath,
			&r.CreatedAt, &r.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan NIS2 incident report: %w", err)
		}
		reports = append(reports, r)
	}
	return reports, nil
}

// EarlyWarningInput holds the data for an early warning submission.
type EarlyWarningInput struct {
	Content       json.RawMessage `json:"content"`
	CSIRTReference string         `json:"csirt_reference"`
}

// SubmitEarlyWarning records Phase 1 (24-hour early warning) submission.
func (s *NIS2Service) SubmitEarlyWarning(ctx context.Context, orgID, reportID, userID uuid.UUID, input EarlyWarningInput) (*NIS2IncidentReport, error) {
	now := time.Now()
	_, err := s.pool.Exec(ctx, `
		UPDATE nis2_incident_reports
		SET early_warning_status = 'submitted',
			early_warning_submitted_at = $1,
			early_warning_submitted_by = $2,
			early_warning_content = $3::JSONB,
			early_warning_csirt_reference = $4
		WHERE id = $5 AND organization_id = $6`,
		now, userID, input.Content, input.CSIRTReference, reportID, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to submit early warning: %w", err)
	}

	log.Info().
		Str("report_id", reportID.String()).
		Str("org_id", orgID.String()).
		Msg("NIS2 Phase 1 Early Warning submitted")

	return s.GetIncidentReport(ctx, orgID, reportID)
}

// NotificationInput holds the data for a notification submission.
type NotificationInput struct {
	Content       json.RawMessage `json:"content"`
	CSIRTReference string         `json:"csirt_reference"`
}

// SubmitNotification records Phase 2 (72-hour notification) submission.
func (s *NIS2Service) SubmitNotification(ctx context.Context, orgID, reportID, userID uuid.UUID, input NotificationInput) (*NIS2IncidentReport, error) {
	now := time.Now()
	_, err := s.pool.Exec(ctx, `
		UPDATE nis2_incident_reports
		SET notification_status = 'submitted',
			notification_submitted_at = $1,
			notification_submitted_by = $2,
			notification_content = $3::JSONB,
			notification_csirt_reference = $4
		WHERE id = $5 AND organization_id = $6`,
		now, userID, input.Content, input.CSIRTReference, reportID, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to submit notification: %w", err)
	}

	log.Info().
		Str("report_id", reportID.String()).
		Str("org_id", orgID.String()).
		Msg("NIS2 Phase 2 Incident Notification submitted")

	return s.GetIncidentReport(ctx, orgID, reportID)
}

// FinalReportInput holds the data for a final report submission.
type FinalReportInput struct {
	Content       json.RawMessage `json:"content"`
	DocumentPath  string          `json:"document_path"`
}

// SubmitFinalReport records Phase 3 (1-month final report) submission.
func (s *NIS2Service) SubmitFinalReport(ctx context.Context, orgID, reportID, userID uuid.UUID, input FinalReportInput) (*NIS2IncidentReport, error) {
	now := time.Now()
	_, err := s.pool.Exec(ctx, `
		UPDATE nis2_incident_reports
		SET final_report_status = 'submitted',
			final_report_submitted_at = $1,
			final_report_submitted_by = $2,
			final_report_content = $3::JSONB,
			final_report_document_path = $4
		WHERE id = $5 AND organization_id = $6`,
		now, userID, input.Content, input.DocumentPath, reportID, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to submit final report: %w", err)
	}

	log.Info().
		Str("report_id", reportID.String()).
		Str("org_id", orgID.String()).
		Msg("NIS2 Phase 3 Final Report submitted")

	return s.GetIncidentReport(ctx, orgID, reportID)
}

// ============================================================
// SECURITY MEASURES (Article 21)
// ============================================================

// GetSecurityMeasuresStatus returns all 10 NIS2 Article 21 measures for the organisation.
func (s *NIS2Service) GetSecurityMeasuresStatus(ctx context.Context, orgID uuid.UUID) ([]NIS2SecurityMeasure, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, organization_id, measure_code, measure_title,
			COALESCE(measure_description, ''), COALESCE(article_reference, ''),
			implementation_status::TEXT, owner_user_id,
			COALESCE(evidence_description, ''), last_assessed_at,
			next_assessment_date, COALESCE(linked_control_ids, '{}'),
			COALESCE(notes, ''), created_at, updated_at
		FROM nis2_security_measures
		WHERE organization_id = $1
		ORDER BY measure_code ASC`, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get security measures: %w", err)
	}
	defer rows.Close()

	var measures []NIS2SecurityMeasure
	for rows.Next() {
		var m NIS2SecurityMeasure
		if err := rows.Scan(
			&m.ID, &m.OrganizationID, &m.MeasureCode, &m.MeasureTitle,
			&m.MeasureDescription, &m.ArticleReference,
			&m.ImplementationStatus, &m.OwnerUserID,
			&m.EvidenceDescription, &m.LastAssessedAt,
			&m.NextAssessmentDate, &m.LinkedControlIDs,
			&m.Notes, &m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan security measure: %w", err)
		}
		measures = append(measures, m)
	}
	return measures, nil
}

// UpdateMeasureInput contains the fields to update on a security measure.
type UpdateMeasureInput struct {
	ImplementationStatus string     `json:"implementation_status"`
	OwnerUserID          *uuid.UUID `json:"owner_user_id"`
	EvidenceDescription  string     `json:"evidence_description"`
	Notes                string     `json:"notes"`
}

// UpdateSecurityMeasure updates a single NIS2 security measure.
func (s *NIS2Service) UpdateSecurityMeasure(ctx context.Context, orgID, measureID uuid.UUID, input UpdateMeasureInput) (*NIS2SecurityMeasure, error) {
	_, err := s.pool.Exec(ctx, `
		UPDATE nis2_security_measures
		SET implementation_status = $1::nis2_measure_status,
			owner_user_id = $2,
			evidence_description = $3,
			notes = $4,
			last_assessed_at = NOW()
		WHERE id = $5 AND organization_id = $6`,
		input.ImplementationStatus, input.OwnerUserID,
		input.EvidenceDescription, input.Notes,
		measureID, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update security measure: %w", err)
	}

	// Return the updated measure
	m := &NIS2SecurityMeasure{}
	err = s.pool.QueryRow(ctx, `
		SELECT id, organization_id, measure_code, measure_title,
			COALESCE(measure_description, ''), COALESCE(article_reference, ''),
			implementation_status::TEXT, owner_user_id,
			COALESCE(evidence_description, ''), last_assessed_at,
			next_assessment_date, COALESCE(linked_control_ids, '{}'),
			COALESCE(notes, ''), created_at, updated_at
		FROM nis2_security_measures
		WHERE id = $1 AND organization_id = $2`, measureID, orgID,
	).Scan(
		&m.ID, &m.OrganizationID, &m.MeasureCode, &m.MeasureTitle,
		&m.MeasureDescription, &m.ArticleReference,
		&m.ImplementationStatus, &m.OwnerUserID,
		&m.EvidenceDescription, &m.LastAssessedAt,
		&m.NextAssessmentDate, &m.LinkedControlIDs,
		&m.Notes, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve updated measure: %w", err)
	}
	return m, nil
}

// InitSecurityMeasures copies the 10 Article 21 template measures into the organisation.
// This is typically called after entity assessment confirms NIS2 scope.
func (s *NIS2Service) InitSecurityMeasures(ctx context.Context, orgID uuid.UUID) (int, error) {
	result, err := s.pool.Exec(ctx, `
		INSERT INTO nis2_security_measures (
			organization_id, measure_code, measure_title,
			measure_description, article_reference, implementation_status, notes
		)
		SELECT $1, measure_code, measure_title, measure_description,
			article_reference, 'not_started'::nis2_measure_status, notes
		FROM nis2_security_measures
		WHERE organization_id = '00000000-0000-0000-0000-000000000000'
		ON CONFLICT (organization_id, measure_code) DO NOTHING`,
		orgID,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to initialise security measures: %w", err)
	}
	return int(result.RowsAffected()), nil
}

// ============================================================
// MANAGEMENT ACCOUNTABILITY (Article 20)
// ============================================================

// ManagementTrainingInput holds input for recording management training.
type ManagementTrainingInput struct {
	BoardMemberName         string `json:"board_member_name"`
	BoardMemberRole         string `json:"board_member_role"`
	TrainingDate            string `json:"training_date"`
	TrainingProvider        string `json:"training_provider"`
	TrainingCertificatePath string `json:"training_certificate_path"`
	Notes                   string `json:"notes"`
}

// RecordManagementTraining records a board member's cybersecurity training completion.
func (s *NIS2Service) RecordManagementTraining(ctx context.Context, orgID uuid.UUID, input ManagementTrainingInput) (*NIS2ManagementRecord, error) {
	var trainingDate *time.Time
	if input.TrainingDate != "" {
		t, err := time.Parse("2006-01-02", input.TrainingDate)
		if err != nil {
			return nil, fmt.Errorf("invalid training_date format, expected YYYY-MM-DD: %w", err)
		}
		trainingDate = &t
	}

	// Next training due in 12 months
	var nextDue *time.Time
	if trainingDate != nil {
		nd := trainingDate.AddDate(1, 0, 0)
		nextDue = &nd
	}

	record := &NIS2ManagementRecord{}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO nis2_management_accountability (
			organization_id, board_member_name, board_member_role,
			training_completed, training_date, training_provider,
			training_certificate_path, next_training_due, notes
		) VALUES ($1, $2, $3, true, $4, $5, $6, $7, $8)
		RETURNING id, organization_id, board_member_name, board_member_role,
			training_completed, training_date, training_provider,
			training_certificate_path, risk_measures_approved, approval_date,
			COALESCE(approval_document_path, ''), next_training_due,
			COALESCE(notes, ''), created_at, updated_at`,
		orgID, input.BoardMemberName, input.BoardMemberRole,
		trainingDate, input.TrainingProvider, input.TrainingCertificatePath,
		nextDue, input.Notes,
	).Scan(
		&record.ID, &record.OrganizationID, &record.BoardMemberName,
		&record.BoardMemberRole, &record.TrainingCompleted, &record.TrainingDate,
		&record.TrainingProvider, &record.TrainingCertificatePath,
		&record.RiskMeasuresApproved, &record.ApprovalDate,
		&record.ApprovalDocumentPath, &record.NextTrainingDue,
		&record.Notes, &record.CreatedAt, &record.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to record management training: %w", err)
	}

	log.Info().
		Str("org_id", orgID.String()).
		Str("member", input.BoardMemberName).
		Msg("NIS2 management cybersecurity training recorded")

	return record, nil
}

// RiskApprovalInput holds input for recording risk measures approval.
type RiskApprovalInput struct {
	ApprovalDate         string `json:"approval_date"`
	ApprovalDocumentPath string `json:"approval_document_path"`
}

// RecordRiskMeasuresApproval records a board member's approval of cybersecurity risk measures.
func (s *NIS2Service) RecordRiskMeasuresApproval(ctx context.Context, orgID, memberID uuid.UUID, input RiskApprovalInput) (*NIS2ManagementRecord, error) {
	var approvalDate *time.Time
	if input.ApprovalDate != "" {
		t, err := time.Parse("2006-01-02", input.ApprovalDate)
		if err != nil {
			return nil, fmt.Errorf("invalid approval_date format, expected YYYY-MM-DD: %w", err)
		}
		approvalDate = &t
	}

	_, err := s.pool.Exec(ctx, `
		UPDATE nis2_management_accountability
		SET risk_measures_approved = true,
			approval_date = $1,
			approval_document_path = $2
		WHERE id = $3 AND organization_id = $4`,
		approvalDate, input.ApprovalDocumentPath, memberID, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to record risk approval: %w", err)
	}

	record := &NIS2ManagementRecord{}
	err = s.pool.QueryRow(ctx, `
		SELECT id, organization_id, board_member_name, board_member_role,
			training_completed, training_date, training_provider,
			COALESCE(training_certificate_path, ''), risk_measures_approved,
			approval_date, COALESCE(approval_document_path, ''),
			next_training_due, COALESCE(notes, ''), created_at, updated_at
		FROM nis2_management_accountability
		WHERE id = $1 AND organization_id = $2`, memberID, orgID,
	).Scan(
		&record.ID, &record.OrganizationID, &record.BoardMemberName,
		&record.BoardMemberRole, &record.TrainingCompleted, &record.TrainingDate,
		&record.TrainingProvider, &record.TrainingCertificatePath,
		&record.RiskMeasuresApproved, &record.ApprovalDate,
		&record.ApprovalDocumentPath, &record.NextTrainingDue,
		&record.Notes, &record.CreatedAt, &record.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve updated record: %w", err)
	}
	return record, nil
}

// ListManagementRecords returns all management accountability records for an organisation.
func (s *NIS2Service) ListManagementRecords(ctx context.Context, orgID uuid.UUID) ([]NIS2ManagementRecord, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, organization_id, board_member_name, board_member_role,
			training_completed, training_date, training_provider,
			COALESCE(training_certificate_path, ''), risk_measures_approved,
			approval_date, COALESCE(approval_document_path, ''),
			next_training_due, COALESCE(notes, ''), created_at, updated_at
		FROM nis2_management_accountability
		WHERE organization_id = $1
		ORDER BY board_member_name ASC`, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list management records: %w", err)
	}
	defer rows.Close()

	var records []NIS2ManagementRecord
	for rows.Next() {
		var r NIS2ManagementRecord
		if err := rows.Scan(
			&r.ID, &r.OrganizationID, &r.BoardMemberName,
			&r.BoardMemberRole, &r.TrainingCompleted, &r.TrainingDate,
			&r.TrainingProvider, &r.TrainingCertificatePath,
			&r.RiskMeasuresApproved, &r.ApprovalDate,
			&r.ApprovalDocumentPath, &r.NextTrainingDue,
			&r.Notes, &r.CreatedAt, &r.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan management record: %w", err)
		}
		records = append(records, r)
	}
	return records, nil
}

// ============================================================
// DASHBOARD
// ============================================================

// GetComplianceDashboard aggregates all NIS2 compliance metrics for the organisation.
func (s *NIS2Service) GetComplianceDashboard(ctx context.Context, orgID uuid.UUID) (*NIS2Dashboard, error) {
	dash := &NIS2Dashboard{}

	// Entity assessment
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(entity_type::TEXT, 'not_applicable'), COALESCE(is_in_scope, false)
		FROM nis2_entity_assessment
		WHERE organization_id = $1
		ORDER BY assessment_date DESC, created_at DESC
		LIMIT 1`, orgID,
	).Scan(&dash.EntityType, &dash.IsInScope)
	if err != nil {
		if err == pgx.ErrNoRows {
			dash.EntityType = "not_assessed"
			dash.IsInScope = false
		} else {
			return nil, fmt.Errorf("failed to get entity status: %w", err)
		}
	}

	// Security measures summary
	err = s.pool.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE implementation_status = 'implemented'),
			COUNT(*) FILTER (WHERE implementation_status = 'verified'),
			COUNT(*) FILTER (WHERE implementation_status = 'in_progress'),
			COUNT(*) FILTER (WHERE implementation_status = 'not_started')
		FROM nis2_security_measures
		WHERE organization_id = $1`, orgID,
	).Scan(
		&dash.MeasuresTotal, &dash.MeasuresImplemented,
		&dash.MeasuresVerified, &dash.MeasuresInProgress,
		&dash.MeasuresNotStarted,
	)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("failed to get measures summary: %w", err)
	}

	if dash.MeasuresTotal > 0 {
		completed := dash.MeasuresImplemented + dash.MeasuresVerified
		dash.CompliancePercentage = float64(completed) / float64(dash.MeasuresTotal) * 100
	}

	// Incident reports summary
	err = s.pool.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE early_warning_status = 'overdue'
				OR notification_status = 'overdue'
				OR final_report_status = 'overdue'),
			COUNT(*) FILTER (WHERE early_warning_status = 'pending'),
			COUNT(*) FILTER (WHERE notification_status = 'pending'),
			COUNT(*) FILTER (WHERE final_report_status = 'pending')
		FROM nis2_incident_reports
		WHERE organization_id = $1`, orgID,
	).Scan(
		&dash.IncidentReportsTotal, &dash.OverdueReports,
		&dash.PendingEarlyWarnings, &dash.PendingNotifications,
		&dash.PendingFinalReports,
	)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("failed to get incident report summary: %w", err)
	}

	// Management accountability
	err = s.pool.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE training_completed = true),
			COUNT(*) FILTER (WHERE next_training_due < CURRENT_DATE AND training_completed = true)
		FROM nis2_management_accountability
		WHERE organization_id = $1`, orgID,
	).Scan(
		&dash.ManagementTotal, &dash.ManagementTrained,
		&dash.TrainingOverdue,
	)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("failed to get management summary: %w", err)
	}

	// Per-measure breakdown
	rows, err := s.pool.Query(ctx, `
		SELECT measure_code, measure_title, implementation_status::TEXT
		FROM nis2_security_measures
		WHERE organization_id = $1
		ORDER BY measure_code ASC`, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get measures breakdown: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var entry MeasureStatusEntry
		if err := rows.Scan(&entry.MeasureCode, &entry.MeasureTitle, &entry.ImplementationStatus); err != nil {
			return nil, err
		}
		dash.MeasuresBreakdown = append(dash.MeasuresBreakdown, entry)
	}

	return dash, nil
}
