package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"github.com/complianceforge/platform/internal/pkg/pdf"
	"github.com/complianceforge/platform/internal/pkg/storage"
	xlsxpkg "github.com/complianceforge/platform/internal/pkg/xlsx"
)

// ============================================================
// REPORT ENGINE — Advanced multi-format report generation
// ============================================================

// ReportEngine generates, renders, and stores compliance reports.
type ReportEngine struct {
	pool         *pgxpool.Pool
	pdfRenderer  *pdf.ReportRenderer
	xlsxRenderer *xlsxpkg.XLSXRenderer
	storage      storage.Storage
}

// NewReportEngine creates a new ReportEngine with all dependencies.
func NewReportEngine(pool *pgxpool.Pool, store storage.Storage) *ReportEngine {
	return &ReportEngine{
		pool:         pool,
		pdfRenderer:  pdf.NewReportRenderer(),
		xlsxRenderer: xlsxpkg.NewXLSXRenderer(),
		storage:      store,
	}
}

// ============================================================
// REPORT DEFINITION & RUN MODELS
// ============================================================

// ReportDefinition represents a saved report configuration.
type ReportDefinition struct {
	ID                      uuid.UUID       `json:"id" db:"id"`
	OrganizationID          uuid.UUID       `json:"organization_id" db:"organization_id"`
	Name                    string          `json:"name" db:"name"`
	Description             string          `json:"description" db:"description"`
	ReportType              string          `json:"report_type" db:"report_type"`
	Format                  string          `json:"format" db:"format"`
	Filters                 json.RawMessage `json:"filters" db:"filters"`
	Sections                json.RawMessage `json:"sections" db:"sections"`
	Classification          string          `json:"classification" db:"classification"`
	IncludeExecutiveSummary bool            `json:"include_executive_summary" db:"include_executive_summary"`
	IncludeAppendices       bool            `json:"include_appendices" db:"include_appendices"`
	Branding                json.RawMessage `json:"branding" db:"branding"`
	CreatedBy               uuid.UUID       `json:"created_by" db:"created_by"`
	IsTemplate              bool            `json:"is_template" db:"is_template"`
	CreatedAt               time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt               time.Time       `json:"updated_at" db:"updated_at"`
}

// ReportSchedule represents a scheduled report configuration.
type ReportSchedule struct {
	ID                 uuid.UUID   `json:"id" db:"id"`
	OrganizationID     uuid.UUID   `json:"organization_id" db:"organization_id"`
	ReportDefinitionID uuid.UUID   `json:"report_definition_id" db:"report_definition_id"`
	Name               string      `json:"name" db:"name"`
	Frequency          string      `json:"frequency" db:"frequency"`
	DayOfWeek          *int        `json:"day_of_week" db:"day_of_week"`
	DayOfMonth         *int        `json:"day_of_month" db:"day_of_month"`
	TimeOfDay          string      `json:"time_of_day" db:"time_of_day"`
	Timezone           string      `json:"timezone" db:"timezone"`
	RecipientUserIDs   []uuid.UUID `json:"recipient_user_ids" db:"recipient_user_ids"`
	RecipientEmails    []string    `json:"recipient_emails" db:"recipient_emails"`
	DeliveryChannel    string      `json:"delivery_channel" db:"delivery_channel"`
	IsActive           bool        `json:"is_active" db:"is_active"`
	LastRunAt          *time.Time  `json:"last_run_at" db:"last_run_at"`
	NextRunAt          *time.Time  `json:"next_run_at" db:"next_run_at"`
	CreatedAt          time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time   `json:"updated_at" db:"updated_at"`
}

// ReportRun tracks a single report generation execution.
type ReportRun struct {
	ID                 uuid.UUID       `json:"id" db:"id"`
	OrganizationID     uuid.UUID       `json:"organization_id" db:"organization_id"`
	ReportDefinitionID uuid.UUID       `json:"report_definition_id" db:"report_definition_id"`
	ScheduleID         *uuid.UUID      `json:"schedule_id" db:"schedule_id"`
	Status             string          `json:"status" db:"status"`
	Format             string          `json:"format" db:"format"`
	FilePath           string          `json:"file_path" db:"file_path"`
	FileSizeBytes      int64           `json:"file_size_bytes" db:"file_size_bytes"`
	PageCount          int             `json:"page_count" db:"page_count"`
	GenerationTimeMs   int             `json:"generation_time_ms" db:"generation_time_ms"`
	Parameters         json.RawMessage `json:"parameters" db:"parameters"`
	GeneratedBy        uuid.UUID       `json:"generated_by" db:"generated_by"`
	ErrorMessage       string          `json:"error_message" db:"error_message"`
	CreatedAt          time.Time       `json:"created_at" db:"created_at"`
	CompletedAt        *time.Time      `json:"completed_at" db:"completed_at"`
}

// ReportFilters contains common filter parameters for report queries.
type ReportFilters struct {
	FrameworkID *uuid.UUID `json:"framework_id,omitempty"`
	AuditID     *uuid.UUID `json:"audit_id,omitempty"`
	DateFrom    *time.Time `json:"date_from,omitempty"`
	DateTo      *time.Time `json:"date_to,omitempty"`
	Severity    string     `json:"severity,omitempty"`
	Status      string     `json:"status,omitempty"`
	RiskLevel   string     `json:"risk_level,omitempty"`
	IncludeAll  bool       `json:"include_all,omitempty"`
}

// ============================================================
// API REQUEST/RESPONSE TYPES
// ============================================================

// GenerateReportRequest is the payload for ad-hoc report generation.
type GenerateReportRequest struct {
	ReportType              string          `json:"report_type" validate:"required"`
	Format                  string          `json:"format" validate:"required,oneof=pdf xlsx csv json"`
	Name                    string          `json:"name,omitempty"`
	Filters                 json.RawMessage `json:"filters,omitempty"`
	Classification          string          `json:"classification,omitempty"`
	IncludeExecutiveSummary bool            `json:"include_executive_summary,omitempty"`
}

// CreateReportDefinitionRequest is the payload for saving a report definition.
type CreateReportDefinitionRequest struct {
	Name                    string          `json:"name" validate:"required,min=3,max=500"`
	Description             string          `json:"description,omitempty"`
	ReportType              string          `json:"report_type" validate:"required"`
	Format                  string          `json:"format" validate:"required,oneof=pdf xlsx csv json"`
	Filters                 json.RawMessage `json:"filters,omitempty"`
	Sections                json.RawMessage `json:"sections,omitempty"`
	Classification          string          `json:"classification,omitempty"`
	IncludeExecutiveSummary bool            `json:"include_executive_summary,omitempty"`
	IncludeAppendices       bool            `json:"include_appendices,omitempty"`
	Branding                json.RawMessage `json:"branding,omitempty"`
	IsTemplate              bool            `json:"is_template,omitempty"`
}

// UpdateReportDefinitionRequest is the payload for updating a report definition.
type UpdateReportDefinitionRequest struct {
	Name                    *string          `json:"name,omitempty" validate:"omitempty,min=3,max=500"`
	Description             *string          `json:"description,omitempty"`
	Format                  *string          `json:"format,omitempty" validate:"omitempty,oneof=pdf xlsx csv json"`
	Filters                 *json.RawMessage `json:"filters,omitempty"`
	Sections                *json.RawMessage `json:"sections,omitempty"`
	Classification          *string          `json:"classification,omitempty"`
	IncludeExecutiveSummary *bool            `json:"include_executive_summary,omitempty"`
	IncludeAppendices       *bool            `json:"include_appendices,omitempty"`
	Branding                *json.RawMessage `json:"branding,omitempty"`
}

// CreateScheduleRequest is the payload for creating a report schedule.
type CreateScheduleRequest struct {
	ReportDefinitionID uuid.UUID   `json:"report_definition_id" validate:"required"`
	Name               string      `json:"name" validate:"required,min=3,max=500"`
	Frequency          string      `json:"frequency" validate:"required,oneof=daily weekly monthly quarterly annually"`
	DayOfWeek          *int        `json:"day_of_week,omitempty" validate:"omitempty,min=0,max=6"`
	DayOfMonth         *int        `json:"day_of_month,omitempty" validate:"omitempty,min=1,max=31"`
	TimeOfDay          string      `json:"time_of_day,omitempty"`
	Timezone           string      `json:"timezone,omitempty"`
	RecipientUserIDs   []uuid.UUID `json:"recipient_user_ids,omitempty"`
	RecipientEmails    []string    `json:"recipient_emails,omitempty"`
	DeliveryChannel    string      `json:"delivery_channel,omitempty" validate:"omitempty,oneof=email storage both"`
}

// UpdateScheduleRequest is the payload for updating a report schedule.
type UpdateScheduleRequest struct {
	Name             *string      `json:"name,omitempty" validate:"omitempty,min=3,max=500"`
	Frequency        *string      `json:"frequency,omitempty" validate:"omitempty,oneof=daily weekly monthly quarterly annually"`
	DayOfWeek        *int         `json:"day_of_week,omitempty" validate:"omitempty,min=0,max=6"`
	DayOfMonth       *int         `json:"day_of_month,omitempty" validate:"omitempty,min=1,max=31"`
	TimeOfDay        *string      `json:"time_of_day,omitempty"`
	Timezone         *string      `json:"timezone,omitempty"`
	RecipientUserIDs *[]uuid.UUID `json:"recipient_user_ids,omitempty"`
	RecipientEmails  *[]string    `json:"recipient_emails,omitempty"`
	DeliveryChannel  *string      `json:"delivery_channel,omitempty" validate:"omitempty,oneof=email storage both"`
	IsActive         *bool        `json:"is_active,omitempty"`
}

// ============================================================
// REPORT DATA STRUCTURES — used to pass data to renderers
// ============================================================

// ComplianceReportData holds all data for a compliance status report.
type ComplianceReportData struct {
	Metadata             pdf.ReportMetadata
	OverallScore         float64
	FrameworkSummaries   []FrameworkReportSummary
	TopGaps              []TopGapEntry
	MaturityDistribution map[string]int
	ControlsByStatus     map[string]int
}

// RiskReportData holds all data for a risk register report.
type RiskReportData struct {
	Metadata          pdf.ReportMetadata
	TotalRisks        int
	RisksByLevel      map[string]int
	RisksByCategory   []CategoryRiskCount
	TopRisks          []TopRiskEntry
	HeatmapData       []HeatmapCell
	TreatmentProgress TreatmentSummary
	AverageRiskScore  float64
}

// HeatmapCell represents a single cell in a risk heatmap.
type HeatmapCell struct {
	Likelihood int      `json:"likelihood"`
	Impact     int      `json:"impact"`
	Count      int      `json:"count"`
	RiskRefs   []string `json:"risk_refs"`
}

// AuditReportData holds data for an audit report.
type AuditReportData struct {
	Metadata           pdf.ReportMetadata
	AuditRef           string
	AuditTitle         string
	AuditType          string
	AuditStatus        string
	Scope              string
	LeadAuditor        string
	PlannedStart       string
	PlannedEnd         string
	ActualStart        string
	ActualEnd          string
	Conclusion         string
	TotalFindings      int
	FindingsBySeverity map[string]int
	Findings           []AuditFindingRow
}

// AuditFindingRow represents a finding in the audit report.
type AuditFindingRow struct {
	FindingRef     string `json:"finding_ref"`
	Title          string `json:"title"`
	Severity       string `json:"severity"`
	Status         string `json:"status"`
	FindingType    string `json:"finding_type"`
	ControlCode    string `json:"control_code"`
	Responsible    string `json:"responsible"`
	DueDate        string `json:"due_date"`
	RootCause      string `json:"root_cause"`
	Recommendation string `json:"recommendation"`
}

// IncidentReportData holds data for an incident summary report.
type IncidentReportData struct {
	Metadata            pdf.ReportMetadata
	TotalIncidents      int
	OpenIncidents       int
	IncidentsBySeverity map[string]int
	IncidentsByType     map[string]int
	IncidentsByStatus   map[string]int
	DataBreaches        []BreachEntry
	RecentIncidents     []IncidentRow
	AvgResolutionHours  float64
}

// BreachEntry represents a data breach in the report.
type BreachEntry struct {
	IncidentRef          string `json:"incident_ref"`
	Title                string `json:"title"`
	Severity             string `json:"severity"`
	DataSubjectsAffected int    `json:"data_subjects_affected"`
	DPANotified          bool   `json:"dpa_notified"`
	NotificationDeadline string `json:"notification_deadline"`
	Status               string `json:"status"`
}

// IncidentRow represents an incident entry in the report.
type IncidentRow struct {
	IncidentRef     string  `json:"incident_ref"`
	Title           string  `json:"title"`
	Type            string  `json:"type"`
	Severity        string  `json:"severity"`
	Status          string  `json:"status"`
	ReportedAt      string  `json:"reported_at"`
	AssignedTo      string  `json:"assigned_to"`
	FinancialImpact float64 `json:"financial_impact"`
}

// VendorReportData holds data for a vendor risk report.
type VendorReportData struct {
	Metadata           pdf.ReportMetadata
	TotalVendors       int
	VendorsByTier      map[string]int
	VendorsByStatus    map[string]int
	CriticalVendors    []VendorRow
	OverdueAssessments int
	AvgRiskScore       float64
}

// VendorRow represents a vendor entry in the report.
type VendorRow struct {
	VendorRef      string  `json:"vendor_ref"`
	Name           string  `json:"name"`
	RiskTier       string  `json:"risk_tier"`
	RiskScore      float64 `json:"risk_score"`
	Status         string  `json:"status"`
	DataProcessing bool    `json:"data_processing"`
	DPAInPlace     bool    `json:"dpa_in_place"`
	LastAssessment string  `json:"last_assessment"`
	NextAssessment string  `json:"next_assessment"`
	Certifications string  `json:"certifications"`
}

// PolicyReportData holds data for a policy status report.
type PolicyReportData struct {
	Metadata             pdf.ReportMetadata
	TotalPolicies        int
	PoliciesByStatus     map[string]int
	OverdueReviews       int
	AttestationRate      float64
	PoliciesDueForReview []PolicyRow
	AttestationSummary   AttestationStats
}

// PolicyRow represents a policy entry in the report.
type PolicyRow struct {
	PolicyRef       string `json:"policy_ref"`
	Title           string `json:"title"`
	Status          string `json:"status"`
	Classification  string `json:"classification"`
	Owner           string `json:"owner"`
	CurrentVersion  int    `json:"current_version"`
	NextReviewDate  string `json:"next_review_date"`
	AttestationRate string `json:"attestation_rate"`
}

// AttestationStats holds attestation completion statistics.
type AttestationStats struct {
	TotalRequired int     `json:"total_required"`
	Completed     int     `json:"completed"`
	Pending       int     `json:"pending"`
	Overdue       int     `json:"overdue"`
	Rate          float64 `json:"rate"`
}

// ExecutiveSummaryData holds data for a board-level executive summary.
type ExecutiveSummaryData struct {
	Metadata               pdf.ReportMetadata
	OverallComplianceScore float64
	OverallRiskScore       float64
	CriticalRiskCount      int
	HighRiskCount          int
	OpenIncidents          int
	DataBreachCount        int
	AuditFindingsOpen      int
	PolicyComplianceRate   float64
	TopRisks               []TopRiskEntry
	ComplianceByFramework  []FrameworkReportSummary
	KeyMetrics             []KeyMetric
	TreatmentProgress      TreatmentSummary
}

// KeyMetric represents a single KPI for the executive summary.
type KeyMetric struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Trend  string `json:"trend"`  // up, down, stable
	Status string `json:"status"` // green, amber, red
}

// GapAnalysisData holds data for a gap analysis report.
type GapAnalysisData struct {
	Metadata         pdf.ReportMetadata
	FrameworkCode    string
	FrameworkName    string
	TotalControls    int
	ImplementedCount int
	GapCount         int
	ComplianceScore  float64
	GapsByDomain     []DomainGapSummary
	RemediationPlan  []RemediationItem
}

// DomainGapSummary holds gap analysis data for a framework domain.
type DomainGapSummary struct {
	DomainCode      string  `json:"domain_code"`
	DomainName      string  `json:"domain_name"`
	TotalControls   int     `json:"total_controls"`
	Implemented     int     `json:"implemented"`
	Gaps            int     `json:"gaps"`
	ComplianceScore float64 `json:"compliance_score"`
}

// RemediationItem represents a remediation action in the gap analysis.
type RemediationItem struct {
	ControlCode    string `json:"control_code"`
	ControlTitle   string `json:"control_title"`
	GapDescription string `json:"gap_description"`
	RiskLevel      string `json:"risk_level"`
	Owner          string `json:"owner"`
	DueDate        string `json:"due_date"`
	Priority       int    `json:"priority"`
}

// CrossFrameworkData holds data for a cross-framework mapping report.
type CrossFrameworkData struct {
	Metadata           pdf.ReportMetadata
	Frameworks         []FrameworkCoverageSummary
	MappingMatrix      []MappingRow
	SharedControls     int
	UniqueControls     int
	CoveragePercentage float64
}

// FrameworkCoverageSummary holds coverage data for a single framework.
type FrameworkCoverageSummary struct {
	Code            string  `json:"code"`
	Name            string  `json:"name"`
	TotalControls   int     `json:"total_controls"`
	MappedControls  int     `json:"mapped_controls"`
	CoveragePercent float64 `json:"coverage_percent"`
}

// MappingRow represents a cross-framework mapping entry.
type MappingRow struct {
	SourceFramework string  `json:"source_framework"`
	SourceControl   string  `json:"source_control"`
	TargetFramework string  `json:"target_framework"`
	TargetControl   string  `json:"target_control"`
	MappingType     string  `json:"mapping_type"`
	Strength        float64 `json:"strength"`
}

// ============================================================
// CORE GENERATION — dispatches to specific report generators
// ============================================================

// GenerateReport creates a report run, generates the report, stores the file, and updates the run record.
func (e *ReportEngine) GenerateReport(ctx context.Context, orgID, definitionID, userID uuid.UUID) (*ReportRun, error) {
	// Load the report definition
	def, err := e.GetDefinition(ctx, orgID, definitionID)
	if err != nil {
		return nil, fmt.Errorf("failed to load report definition: %w", err)
	}

	// Create a pending report run
	run := &ReportRun{
		OrganizationID:     orgID,
		ReportDefinitionID: definitionID,
		Status:             "pending",
		Format:             def.Format,
		GeneratedBy:        userID,
	}

	err = e.pool.QueryRow(ctx, `
		INSERT INTO report_runs (organization_id, report_definition_id, status, format, generated_by, parameters)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`,
		run.OrganizationID, run.ReportDefinitionID, run.Status, run.Format, run.GeneratedBy, def.Filters,
	).Scan(&run.ID, &run.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create report run: %w", err)
	}

	// Mark as generating
	_, _ = e.pool.Exec(ctx, `UPDATE report_runs SET status = 'generating' WHERE id = $1`, run.ID)
	run.Status = "generating"

	startTime := time.Now()

	// Parse filters
	var filters ReportFilters
	if len(def.Filters) > 0 {
		_ = json.Unmarshal(def.Filters, &filters)
	}

	// Look up org name for metadata
	orgName := e.getOrgName(ctx, orgID)
	userName := e.getUserName(ctx, userID)

	baseMeta := pdf.ReportMetadata{
		OrganizationName: orgName,
		GeneratedBy:      userName,
		GeneratedAt:      time.Now(),
		Classification:   def.Classification,
		ReportPeriod:     time.Now().Format("January 2006"),
		PageSize:         "A4",
	}

	// Dispatch to the correct generator
	var reportBytes []byte
	var genErr error

	switch def.ReportType {
	case "compliance_status":
		reportBytes, genErr = e.generateComplianceOutput(ctx, orgID, def.Format, filters, baseMeta)
	case "risk_register", "risk_heatmap":
		reportBytes, genErr = e.generateRiskOutput(ctx, orgID, def.Format, filters, baseMeta)
	case "audit_summary", "audit_findings":
		reportBytes, genErr = e.generateAuditOutput(ctx, orgID, def.Format, filters, baseMeta)
	case "incident_summary", "breach_register":
		reportBytes, genErr = e.generateIncidentOutput(ctx, orgID, def.Format, filters, baseMeta)
	case "vendor_risk":
		reportBytes, genErr = e.generateVendorOutput(ctx, orgID, def.Format, filters, baseMeta)
	case "policy_status", "attestation_report":
		reportBytes, genErr = e.generatePolicyOutput(ctx, orgID, def.Format, filters, baseMeta)
	case "executive_summary":
		reportBytes, genErr = e.generateExecutiveOutput(ctx, orgID, def.Format, baseMeta)
	case "gap_analysis":
		reportBytes, genErr = e.generateGapAnalysisOutput(ctx, orgID, def.Format, filters, baseMeta)
	case "cross_framework_mapping":
		reportBytes, genErr = e.generateCrossFrameworkOutput(ctx, orgID, def.Format, baseMeta)
	default:
		genErr = fmt.Errorf("unsupported report type: %s", def.ReportType)
	}

	generationTimeMs := int(time.Since(startTime).Milliseconds())

	if genErr != nil {
		log.Error().Err(genErr).Str("run_id", run.ID.String()).Msg("Report generation failed")
		_, _ = e.pool.Exec(ctx, `
			UPDATE report_runs SET status = 'failed', error_message = $2, generation_time_ms = $3, completed_at = NOW()
			WHERE id = $1`, run.ID, genErr.Error(), generationTimeMs)
		run.Status = "failed"
		run.ErrorMessage = genErr.Error()
		return run, genErr
	}

	// Store the generated file
	ext := def.Format
	fileName := fmt.Sprintf("%s_%s_%s.%s", def.ReportType, orgID.String()[:8], time.Now().Format("20060102_150405"), ext)

	storedFile, storeErr := e.storage.Store(orgID.String(), "reports", fileName, bytes.NewReader(reportBytes))
	if storeErr != nil {
		log.Error().Err(storeErr).Msg("Failed to store report file")
		_, _ = e.pool.Exec(ctx, `
			UPDATE report_runs SET status = 'failed', error_message = $2, generation_time_ms = $3, completed_at = NOW()
			WHERE id = $1`, run.ID, "Failed to store report file: "+storeErr.Error(), generationTimeMs)
		run.Status = "failed"
		run.ErrorMessage = storeErr.Error()
		return run, storeErr
	}

	// Mark as completed
	now := time.Now()
	_, _ = e.pool.Exec(ctx, `
		UPDATE report_runs
		SET status = 'completed', file_path = $2, file_size_bytes = $3,
		    generation_time_ms = $4, completed_at = NOW()
		WHERE id = $1`,
		run.ID, storedFile.Path, storedFile.Size, generationTimeMs)

	run.Status = "completed"
	run.FilePath = storedFile.Path
	run.FileSizeBytes = storedFile.Size
	run.GenerationTimeMs = generationTimeMs
	run.CompletedAt = &now

	log.Info().
		Str("run_id", run.ID.String()).
		Str("type", def.ReportType).
		Str("format", def.Format).
		Int("generation_ms", generationTimeMs).
		Int64("file_size", storedFile.Size).
		Msg("Report generated successfully")

	return run, nil
}

// GenerateFromSchedule creates a report run triggered by a schedule.
func (e *ReportEngine) GenerateFromSchedule(ctx context.Context, schedule *ReportSchedule, userID uuid.UUID) (*ReportRun, error) {
	run, err := e.GenerateReport(ctx, schedule.OrganizationID, schedule.ReportDefinitionID, userID)
	if err != nil {
		return run, err
	}

	// Link the run to the schedule
	if run != nil {
		_, _ = e.pool.Exec(ctx, `UPDATE report_runs SET schedule_id = $2 WHERE id = $1`, run.ID, schedule.ID)
		run.ScheduleID = &schedule.ID
	}

	return run, nil
}

// ============================================================
// OUTPUT GENERATION — renders data in the requested format
// ============================================================

func (e *ReportEngine) generateComplianceOutput(ctx context.Context, orgID uuid.UUID, format string, filters ReportFilters, meta pdf.ReportMetadata) ([]byte, error) {
	data, err := e.GenerateComplianceReport(ctx, orgID, filters)
	if err != nil {
		return nil, err
	}

	meta.Title = "Compliance Status Report"
	meta.Subtitle = fmt.Sprintf("Overall Score: %.1f%%", data.OverallScore)
	data.Metadata = meta

	switch format {
	case "pdf":
		return e.pdfRenderer.RenderComplianceReport(data)
	case "xlsx":
		return e.xlsxRenderer.RenderComplianceReport(data)
	case "csv":
		return e.renderComplianceCSV(data)
	case "json":
		return json.MarshalIndent(data, "", "  ")
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

func (e *ReportEngine) generateRiskOutput(ctx context.Context, orgID uuid.UUID, format string, filters ReportFilters, meta pdf.ReportMetadata) ([]byte, error) {
	data, err := e.GenerateRiskReport(ctx, orgID, filters)
	if err != nil {
		return nil, err
	}

	meta.Title = "Risk Register Report"
	meta.Subtitle = fmt.Sprintf("Total Risks: %d | Average Score: %.1f", data.TotalRisks, data.AverageRiskScore)
	data.Metadata = meta

	switch format {
	case "pdf":
		return e.pdfRenderer.RenderRiskReport(data)
	case "xlsx":
		return e.xlsxRenderer.RenderRiskReport(data)
	case "csv":
		return e.renderRiskCSV(data)
	case "json":
		return json.MarshalIndent(data, "", "  ")
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

func (e *ReportEngine) generateAuditOutput(ctx context.Context, orgID uuid.UUID, format string, filters ReportFilters, meta pdf.ReportMetadata) ([]byte, error) {
	if filters.AuditID != nil {
		data, err := e.GenerateAuditReport(ctx, orgID, *filters.AuditID)
		if err != nil {
			return nil, err
		}
		meta.Title = fmt.Sprintf("Audit Report — %s", data.AuditRef)
		meta.Subtitle = data.AuditTitle
		data.Metadata = meta

		switch format {
		case "json":
			return json.MarshalIndent(data, "", "  ")
		default:
			return e.pdfRenderer.RenderAuditReport(data)
		}
	}

	// Summary of all audits
	data, err := e.generateAuditSummary(ctx, orgID, filters)
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(data, "", "  ")
}

func (e *ReportEngine) generateIncidentOutput(ctx context.Context, orgID uuid.UUID, format string, filters ReportFilters, meta pdf.ReportMetadata) ([]byte, error) {
	data, err := e.GenerateIncidentReport(ctx, orgID, filters)
	if err != nil {
		return nil, err
	}

	meta.Title = "Incident Summary Report"
	meta.Subtitle = fmt.Sprintf("Total: %d | Open: %d", data.TotalIncidents, data.OpenIncidents)
	data.Metadata = meta

	switch format {
	case "json":
		return json.MarshalIndent(data, "", "  ")
	default:
		return e.pdfRenderer.RenderIncidentReport(data)
	}
}

func (e *ReportEngine) generateVendorOutput(ctx context.Context, orgID uuid.UUID, format string, filters ReportFilters, meta pdf.ReportMetadata) ([]byte, error) {
	data, err := e.GenerateVendorReport(ctx, orgID, filters)
	if err != nil {
		return nil, err
	}

	meta.Title = "Vendor Risk Assessment Report"
	meta.Subtitle = fmt.Sprintf("Vendors: %d | Avg Score: %.1f", data.TotalVendors, data.AvgRiskScore)
	data.Metadata = meta

	switch format {
	case "json":
		return json.MarshalIndent(data, "", "  ")
	default:
		return e.pdfRenderer.RenderVendorReport(data)
	}
}

func (e *ReportEngine) generatePolicyOutput(ctx context.Context, orgID uuid.UUID, format string, filters ReportFilters, meta pdf.ReportMetadata) ([]byte, error) {
	data, err := e.GeneratePolicyReport(ctx, orgID, filters)
	if err != nil {
		return nil, err
	}

	meta.Title = "Policy Status Report"
	meta.Subtitle = fmt.Sprintf("Policies: %d | Attestation Rate: %.1f%%", data.TotalPolicies, data.AttestationRate)
	data.Metadata = meta

	switch format {
	case "json":
		return json.MarshalIndent(data, "", "  ")
	default:
		return e.pdfRenderer.RenderPolicyReport(data)
	}
}

func (e *ReportEngine) generateExecutiveOutput(ctx context.Context, orgID uuid.UUID, format string, meta pdf.ReportMetadata) ([]byte, error) {
	data, err := e.GenerateExecutiveSummary(ctx, orgID)
	if err != nil {
		return nil, err
	}

	meta.Title = "Executive Summary"
	meta.Subtitle = "Board-Level GRC Overview"
	data.Metadata = meta

	switch format {
	case "pdf":
		return e.pdfRenderer.RenderExecutiveSummary(data)
	case "json":
		return json.MarshalIndent(data, "", "  ")
	default:
		return e.pdfRenderer.RenderExecutiveSummary(data)
	}
}

func (e *ReportEngine) generateGapAnalysisOutput(ctx context.Context, orgID uuid.UUID, format string, filters ReportFilters, meta pdf.ReportMetadata) ([]byte, error) {
	if filters.FrameworkID == nil {
		return nil, fmt.Errorf("framework_id filter is required for gap analysis report")
	}

	data, err := e.GenerateGapAnalysis(ctx, orgID, *filters.FrameworkID)
	if err != nil {
		return nil, err
	}

	meta.Title = fmt.Sprintf("Gap Analysis — %s", data.FrameworkName)
	meta.Subtitle = fmt.Sprintf("Score: %.1f%% | Gaps: %d", data.ComplianceScore, data.GapCount)
	data.Metadata = meta

	switch format {
	case "json":
		return json.MarshalIndent(data, "", "  ")
	default:
		return e.pdfRenderer.RenderGapAnalysis(data)
	}
}

func (e *ReportEngine) generateCrossFrameworkOutput(ctx context.Context, orgID uuid.UUID, format string, meta pdf.ReportMetadata) ([]byte, error) {
	data, err := e.GenerateCrossFrameworkReport(ctx, orgID)
	if err != nil {
		return nil, err
	}

	meta.Title = "Cross-Framework Mapping Report"
	meta.Subtitle = fmt.Sprintf("Coverage: %.1f%%", data.CoveragePercentage)
	data.Metadata = meta

	switch format {
	case "json":
		return json.MarshalIndent(data, "", "  ")
	default:
		return e.pdfRenderer.RenderCrossFrameworkReport(data)
	}
}

// ============================================================
// DATA GENERATION — queries database for report data
// ============================================================

// GenerateComplianceReport queries frameworks, controls, and scores.
func (e *ReportEngine) GenerateComplianceReport(ctx context.Context, orgID uuid.UUID, filters ReportFilters) (*ComplianceReportData, error) {
	data := &ComplianceReportData{
		MaturityDistribution: make(map[string]int),
		ControlsByStatus:     make(map[string]int),
	}

	// Framework summaries from the compliance score view
	rows, err := e.pool.Query(ctx, `
		SELECT framework_code, framework_name, COALESCE(compliance_score, 0),
			total_controls, implemented,
			(total_controls - implemented - not_applicable) AS gaps,
			COALESCE(maturity_avg, 0)
		FROM v_compliance_score_by_framework WHERE organization_id = $1`, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to query framework scores: %w", err)
	}
	defer rows.Close()

	var totalScore float64
	var count int
	for rows.Next() {
		var f FrameworkReportSummary
		if err := rows.Scan(&f.Code, &f.Name, &f.Score, &f.TotalControls,
			&f.Implemented, &f.Gaps, &f.MaturityAvg); err != nil {
			return nil, err
		}
		data.FrameworkSummaries = append(data.FrameworkSummaries, f)
		totalScore += f.Score
		count++
	}
	if count > 0 {
		data.OverallScore = totalScore / float64(count)
	}

	// Top gaps ordered by risk level
	gapRows, err := e.pool.Query(ctx, `
		SELECT cf.code, fc.code, fc.title, ci.risk_if_not_implemented,
			COALESCE(u.first_name || ' ' || u.last_name, 'Unassigned'),
			COALESCE(ci.remediation_due_date::TEXT, 'Not set')
		FROM control_implementations ci
		JOIN framework_controls fc ON ci.framework_control_id = fc.id
		JOIN organization_frameworks of2 ON ci.org_framework_id = of2.id
		JOIN compliance_frameworks cf ON of2.framework_id = cf.id
		LEFT JOIN users u ON ci.owner_user_id = u.id
		WHERE ci.organization_id = $1
			AND ci.status NOT IN ('implemented', 'effective', 'not_applicable')
			AND ci.deleted_at IS NULL
		ORDER BY CASE ci.risk_if_not_implemented
			WHEN 'critical' THEN 1 WHEN 'high' THEN 2 WHEN 'medium' THEN 3 ELSE 4 END
		LIMIT 20`, orgID)
	if err == nil {
		defer gapRows.Close()
		for gapRows.Next() {
			var g TopGapEntry
			_ = gapRows.Scan(&g.Framework, &g.ControlCode, &g.ControlTitle,
				&g.RiskLevel, &g.Owner, &g.RemediationDue)
			data.TopGaps = append(data.TopGaps, g)
		}
	}

	// Maturity distribution
	matRows, err := e.pool.Query(ctx, `
		SELECT maturity_level, COUNT(*)
		FROM control_implementations
		WHERE organization_id = $1 AND status != 'not_applicable' AND deleted_at IS NULL
		GROUP BY maturity_level ORDER BY maturity_level`, orgID)
	if err == nil {
		defer matRows.Close()
		levels := []string{"Non-existent", "Initial", "Managed", "Defined", "Quantitatively Managed", "Optimizing"}
		for matRows.Next() {
			var level, cnt int
			_ = matRows.Scan(&level, &cnt)
			if level >= 0 && level < len(levels) {
				data.MaturityDistribution[fmt.Sprintf("Level %d - %s", level, levels[level])] = cnt
			}
		}
	}

	// Controls by status
	statusRows, err := e.pool.Query(ctx, `
		SELECT status::TEXT, COUNT(*)
		FROM control_implementations
		WHERE organization_id = $1 AND deleted_at IS NULL
		GROUP BY status`, orgID)
	if err == nil {
		defer statusRows.Close()
		for statusRows.Next() {
			var status string
			var cnt int
			_ = statusRows.Scan(&status, &cnt)
			data.ControlsByStatus[status] = cnt
		}
	}

	return data, nil
}

// GenerateRiskReport queries risks, heatmap data, treatment progress.
func (e *ReportEngine) GenerateRiskReport(ctx context.Context, orgID uuid.UUID, filters ReportFilters) (*RiskReportData, error) {
	data := &RiskReportData{
		RisksByLevel: make(map[string]int),
	}

	// Total risks
	_ = e.pool.QueryRow(ctx, `SELECT COUNT(*) FROM risks WHERE organization_id = $1 AND deleted_at IS NULL`, orgID).Scan(&data.TotalRisks)

	// Risks by level
	levelRows, _ := e.pool.Query(ctx, `
		SELECT residual_risk_level::TEXT, COUNT(*)
		FROM risks WHERE organization_id = $1 AND deleted_at IS NULL
		GROUP BY residual_risk_level`, orgID)
	if levelRows != nil {
		defer levelRows.Close()
		for levelRows.Next() {
			var level string
			var cnt int
			_ = levelRows.Scan(&level, &cnt)
			data.RisksByLevel[level] = cnt
		}
	}

	// Risks by category
	catRows, _ := e.pool.Query(ctx, `
		SELECT rc.name, COUNT(*)
		FROM risks r JOIN risk_categories rc ON r.risk_category_id = rc.id
		WHERE r.organization_id = $1 AND r.deleted_at IS NULL
		GROUP BY rc.name ORDER BY COUNT(*) DESC`, orgID)
	if catRows != nil {
		defer catRows.Close()
		for catRows.Next() {
			var c CategoryRiskCount
			_ = catRows.Scan(&c.Category, &c.Count)
			data.RisksByCategory = append(data.RisksByCategory, c)
		}
	}

	// Top 10 risks
	topRows, _ := e.pool.Query(ctx, `
		SELECT r.risk_ref, r.title, r.residual_risk_score, r.residual_risk_level::TEXT,
			COALESCE(u.first_name || ' ' || u.last_name, 'Unassigned'),
			(SELECT COUNT(*) FROM risk_treatments rt WHERE rt.risk_id = r.id AND rt.status NOT IN ('completed','cancelled'))
		FROM risks r LEFT JOIN users u ON r.owner_user_id = u.id
		WHERE r.organization_id = $1 AND r.deleted_at IS NULL
		ORDER BY r.residual_risk_score DESC LIMIT 10`, orgID)
	if topRows != nil {
		defer topRows.Close()
		for topRows.Next() {
			var t TopRiskEntry
			_ = topRows.Scan(&t.Ref, &t.Title, &t.ResidualScore, &t.Level, &t.Owner, &t.TreatmentCount)
			data.TopRisks = append(data.TopRisks, t)
		}
	}

	// Heatmap data — aggregate risks by likelihood/impact
	heatRows, _ := e.pool.Query(ctx, `
		SELECT residual_likelihood, residual_impact, COUNT(*), ARRAY_AGG(risk_ref)
		FROM risks WHERE organization_id = $1 AND deleted_at IS NULL
		GROUP BY residual_likelihood, residual_impact`, orgID)
	if heatRows != nil {
		defer heatRows.Close()
		for heatRows.Next() {
			var cell HeatmapCell
			_ = heatRows.Scan(&cell.Likelihood, &cell.Impact, &cell.Count, &cell.RiskRefs)
			data.HeatmapData = append(data.HeatmapData, cell)
		}
	}

	// Treatment progress
	_ = e.pool.QueryRow(ctx, `
		SELECT COUNT(*),
			COUNT(*) FILTER (WHERE status = 'completed'),
			COUNT(*) FILTER (WHERE status = 'in_progress'),
			COUNT(*) FILTER (WHERE status != 'completed' AND target_date < CURRENT_DATE)
		FROM risk_treatments WHERE organization_id = $1`, orgID,
	).Scan(&data.TreatmentProgress.Total, &data.TreatmentProgress.Completed,
		&data.TreatmentProgress.InProgress, &data.TreatmentProgress.Overdue)

	// Average residual score
	_ = e.pool.QueryRow(ctx, `
		SELECT COALESCE(AVG(residual_risk_score), 0)
		FROM risks WHERE organization_id = $1 AND deleted_at IS NULL`, orgID,
	).Scan(&data.AverageRiskScore)

	return data, nil
}

// GenerateAuditReport produces a detailed audit report for a specific audit.
func (e *ReportEngine) GenerateAuditReport(ctx context.Context, orgID, auditID uuid.UUID) (*AuditReportData, error) {
	data := &AuditReportData{
		FindingsBySeverity: make(map[string]int),
	}

	// Audit header
	err := e.pool.QueryRow(ctx, `
		SELECT a.audit_ref, a.title, a.audit_type, a.status::TEXT, COALESCE(a.scope, ''),
			COALESCE(u.first_name || ' ' || u.last_name, 'Unassigned'),
			COALESCE(a.planned_start_date::TEXT, ''), COALESCE(a.planned_end_date::TEXT, ''),
			COALESCE(a.actual_start_date::TEXT, ''), COALESCE(a.actual_end_date::TEXT, ''),
			COALESCE(a.conclusion, ''), a.total_findings
		FROM audits a
		LEFT JOIN users u ON a.lead_auditor_id = u.id
		WHERE a.id = $1 AND a.organization_id = $2 AND a.deleted_at IS NULL`,
		auditID, orgID,
	).Scan(
		&data.AuditRef, &data.AuditTitle, &data.AuditType, &data.AuditStatus,
		&data.Scope, &data.LeadAuditor, &data.PlannedStart, &data.PlannedEnd,
		&data.ActualStart, &data.ActualEnd, &data.Conclusion, &data.TotalFindings,
	)
	if err != nil {
		return nil, fmt.Errorf("audit not found: %w", err)
	}

	// Findings
	findingRows, err := e.pool.Query(ctx, `
		SELECT af.finding_ref, af.title, af.severity::TEXT, af.status, af.finding_type,
			COALESCE(fc.code, ''), COALESCE(u.first_name || ' ' || u.last_name, 'Unassigned'),
			COALESCE(af.due_date::TEXT, 'Not set'),
			COALESCE(af.root_cause, ''), COALESCE(af.recommendation, '')
		FROM audit_findings af
		LEFT JOIN framework_controls fc ON af.control_id = fc.id
		LEFT JOIN users u ON af.responsible_user_id = u.id
		WHERE af.audit_id = $1 AND af.organization_id = $2 AND af.deleted_at IS NULL
		ORDER BY CASE af.severity
			WHEN 'critical' THEN 1 WHEN 'high' THEN 2 WHEN 'medium' THEN 3
			WHEN 'low' THEN 4 ELSE 5 END`,
		auditID, orgID)
	if err == nil {
		defer findingRows.Close()
		for findingRows.Next() {
			var f AuditFindingRow
			_ = findingRows.Scan(&f.FindingRef, &f.Title, &f.Severity, &f.Status,
				&f.FindingType, &f.ControlCode, &f.Responsible, &f.DueDate,
				&f.RootCause, &f.Recommendation)
			data.Findings = append(data.Findings, f)
			data.FindingsBySeverity[f.Severity]++
		}
	}

	return data, nil
}

// GenerateIncidentReport queries incidents and breach register data.
func (e *ReportEngine) GenerateIncidentReport(ctx context.Context, orgID uuid.UUID, filters ReportFilters) (*IncidentReportData, error) {
	data := &IncidentReportData{
		IncidentsBySeverity: make(map[string]int),
		IncidentsByType:     make(map[string]int),
		IncidentsByStatus:   make(map[string]int),
	}

	// Totals
	_ = e.pool.QueryRow(ctx, `
		SELECT COUNT(*), COUNT(*) FILTER (WHERE status NOT IN ('resolved', 'closed'))
		FROM incidents WHERE organization_id = $1 AND deleted_at IS NULL`, orgID,
	).Scan(&data.TotalIncidents, &data.OpenIncidents)

	// By severity
	sevRows, _ := e.pool.Query(ctx, `
		SELECT severity::TEXT, COUNT(*)
		FROM incidents WHERE organization_id = $1 AND deleted_at IS NULL
		GROUP BY severity`, orgID)
	if sevRows != nil {
		defer sevRows.Close()
		for sevRows.Next() {
			var sev string
			var cnt int
			_ = sevRows.Scan(&sev, &cnt)
			data.IncidentsBySeverity[sev] = cnt
		}
	}

	// By type
	typeRows, _ := e.pool.Query(ctx, `
		SELECT incident_type::TEXT, COUNT(*)
		FROM incidents WHERE organization_id = $1 AND deleted_at IS NULL
		GROUP BY incident_type`, orgID)
	if typeRows != nil {
		defer typeRows.Close()
		for typeRows.Next() {
			var typ string
			var cnt int
			_ = typeRows.Scan(&typ, &cnt)
			data.IncidentsByType[typ] = cnt
		}
	}

	// By status
	statRows, _ := e.pool.Query(ctx, `
		SELECT status::TEXT, COUNT(*)
		FROM incidents WHERE organization_id = $1 AND deleted_at IS NULL
		GROUP BY status`, orgID)
	if statRows != nil {
		defer statRows.Close()
		for statRows.Next() {
			var stat string
			var cnt int
			_ = statRows.Scan(&stat, &cnt)
			data.IncidentsByStatus[stat] = cnt
		}
	}

	// Data breaches
	breachRows, _ := e.pool.Query(ctx, `
		SELECT incident_ref, title, severity::TEXT, data_subjects_affected,
			(dpa_notified_at IS NOT NULL), COALESCE(notification_deadline::TEXT, 'N/A'), status::TEXT
		FROM incidents
		WHERE organization_id = $1 AND is_data_breach = true AND deleted_at IS NULL
		ORDER BY reported_at DESC`, orgID)
	if breachRows != nil {
		defer breachRows.Close()
		for breachRows.Next() {
			var b BreachEntry
			_ = breachRows.Scan(&b.IncidentRef, &b.Title, &b.Severity,
				&b.DataSubjectsAffected, &b.DPANotified, &b.NotificationDeadline, &b.Status)
			data.DataBreaches = append(data.DataBreaches, b)
		}
	}

	// Recent incidents
	recentRows, _ := e.pool.Query(ctx, `
		SELECT i.incident_ref, i.title, i.incident_type::TEXT, i.severity::TEXT, i.status::TEXT,
			i.reported_at::TEXT, COALESCE(u.first_name || ' ' || u.last_name, 'Unassigned'),
			i.financial_impact_eur
		FROM incidents i LEFT JOIN users u ON i.assigned_to = u.id
		WHERE i.organization_id = $1 AND i.deleted_at IS NULL
		ORDER BY i.reported_at DESC LIMIT 20`, orgID)
	if recentRows != nil {
		defer recentRows.Close()
		for recentRows.Next() {
			var row IncidentRow
			_ = recentRows.Scan(&row.IncidentRef, &row.Title, &row.Type, &row.Severity,
				&row.Status, &row.ReportedAt, &row.AssignedTo, &row.FinancialImpact)
			data.RecentIncidents = append(data.RecentIncidents, row)
		}
	}

	// Average resolution time
	_ = e.pool.QueryRow(ctx, `
		SELECT COALESCE(AVG(EXTRACT(EPOCH FROM (resolved_at - reported_at))/3600), 0)
		FROM incidents
		WHERE organization_id = $1 AND resolved_at IS NOT NULL AND deleted_at IS NULL`, orgID,
	).Scan(&data.AvgResolutionHours)

	return data, nil
}

// GenerateVendorReport queries vendor risk assessment data.
func (e *ReportEngine) GenerateVendorReport(ctx context.Context, orgID uuid.UUID, filters ReportFilters) (*VendorReportData, error) {
	data := &VendorReportData{
		VendorsByTier:   make(map[string]int),
		VendorsByStatus: make(map[string]int),
	}

	// Total
	_ = e.pool.QueryRow(ctx, `SELECT COUNT(*) FROM vendors WHERE organization_id = $1 AND deleted_at IS NULL`, orgID).Scan(&data.TotalVendors)

	// By tier
	tierRows, _ := e.pool.Query(ctx, `
		SELECT risk_tier::TEXT, COUNT(*)
		FROM vendors WHERE organization_id = $1 AND deleted_at IS NULL
		GROUP BY risk_tier`, orgID)
	if tierRows != nil {
		defer tierRows.Close()
		for tierRows.Next() {
			var tier string
			var cnt int
			_ = tierRows.Scan(&tier, &cnt)
			data.VendorsByTier[tier] = cnt
		}
	}

	// By status
	statusRows, _ := e.pool.Query(ctx, `
		SELECT status::TEXT, COUNT(*)
		FROM vendors WHERE organization_id = $1 AND deleted_at IS NULL
		GROUP BY status`, orgID)
	if statusRows != nil {
		defer statusRows.Close()
		for statusRows.Next() {
			var stat string
			var cnt int
			_ = statusRows.Scan(&stat, &cnt)
			data.VendorsByStatus[stat] = cnt
		}
	}

	// Critical/high-tier vendors
	vendorRows, _ := e.pool.Query(ctx, `
		SELECT vendor_ref, name, risk_tier::TEXT, risk_score, status::TEXT,
			data_processing, dpa_in_place,
			COALESCE(last_assessment_date::TEXT, 'Never'),
			COALESCE(next_assessment_date::TEXT, 'Not set'),
			COALESCE(array_to_string(certifications, ', '), 'None')
		FROM vendors
		WHERE organization_id = $1 AND deleted_at IS NULL
		ORDER BY risk_score DESC LIMIT 20`, orgID)
	if vendorRows != nil {
		defer vendorRows.Close()
		for vendorRows.Next() {
			var v VendorRow
			_ = vendorRows.Scan(&v.VendorRef, &v.Name, &v.RiskTier, &v.RiskScore,
				&v.Status, &v.DataProcessing, &v.DPAInPlace,
				&v.LastAssessment, &v.NextAssessment, &v.Certifications)
			data.CriticalVendors = append(data.CriticalVendors, v)
		}
	}

	// Overdue assessments
	_ = e.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM vendors
		WHERE organization_id = $1 AND next_assessment_date < CURRENT_DATE AND deleted_at IS NULL`, orgID,
	).Scan(&data.OverdueAssessments)

	// Average risk score
	_ = e.pool.QueryRow(ctx, `
		SELECT COALESCE(AVG(risk_score), 0)
		FROM vendors WHERE organization_id = $1 AND deleted_at IS NULL`, orgID,
	).Scan(&data.AvgRiskScore)

	return data, nil
}

// GeneratePolicyReport queries policy status and attestation data.
func (e *ReportEngine) GeneratePolicyReport(ctx context.Context, orgID uuid.UUID, filters ReportFilters) (*PolicyReportData, error) {
	data := &PolicyReportData{
		PoliciesByStatus: make(map[string]int),
	}

	// Total
	_ = e.pool.QueryRow(ctx, `SELECT COUNT(*) FROM policies WHERE organization_id = $1 AND deleted_at IS NULL`, orgID).Scan(&data.TotalPolicies)

	// By status
	statusRows, _ := e.pool.Query(ctx, `
		SELECT status::TEXT, COUNT(*)
		FROM policies WHERE organization_id = $1 AND deleted_at IS NULL
		GROUP BY status`, orgID)
	if statusRows != nil {
		defer statusRows.Close()
		for statusRows.Next() {
			var stat string
			var cnt int
			_ = statusRows.Scan(&stat, &cnt)
			data.PoliciesByStatus[stat] = cnt
		}
	}

	// Overdue reviews
	_ = e.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM policies
		WHERE organization_id = $1 AND next_review_date < CURRENT_DATE
			AND status NOT IN ('archived', 'retired') AND deleted_at IS NULL`, orgID,
	).Scan(&data.OverdueReviews)

	// Policies due for review
	reviewRows, _ := e.pool.Query(ctx, `
		SELECT p.policy_ref, p.title, p.status::TEXT, p.classification::TEXT,
			COALESCE(u.first_name || ' ' || u.last_name, 'Unassigned'),
			p.current_version, COALESCE(p.next_review_date::TEXT, 'Not set')
		FROM policies p LEFT JOIN users u ON p.owner_user_id = u.id
		WHERE p.organization_id = $1 AND p.deleted_at IS NULL
			AND p.status NOT IN ('archived', 'retired')
		ORDER BY p.next_review_date ASC NULLS LAST LIMIT 20`, orgID)
	if reviewRows != nil {
		defer reviewRows.Close()
		for reviewRows.Next() {
			var row PolicyRow
			_ = reviewRows.Scan(&row.PolicyRef, &row.Title, &row.Status, &row.Classification,
				&row.Owner, &row.CurrentVersion, &row.NextReviewDate)
			data.PoliciesDueForReview = append(data.PoliciesDueForReview, row)
		}
	}

	// Attestation stats
	_ = e.pool.QueryRow(ctx, `
		SELECT COUNT(*),
			COUNT(*) FILTER (WHERE status = 'attested'),
			COUNT(*) FILTER (WHERE status = 'pending'),
			COUNT(*) FILTER (WHERE status = 'pending' AND due_date < CURRENT_DATE)
		FROM policy_attestations
		WHERE organization_id = $1`, orgID,
	).Scan(&data.AttestationSummary.TotalRequired, &data.AttestationSummary.Completed,
		&data.AttestationSummary.Pending, &data.AttestationSummary.Overdue)

	if data.AttestationSummary.TotalRequired > 0 {
		data.AttestationSummary.Rate = float64(data.AttestationSummary.Completed) / float64(data.AttestationSummary.TotalRequired) * 100
		data.AttestationRate = data.AttestationSummary.Rate
	}

	return data, nil
}

// GenerateExecutiveSummary creates a board-level 3-5 page overview.
func (e *ReportEngine) GenerateExecutiveSummary(ctx context.Context, orgID uuid.UUID) (*ExecutiveSummaryData, error) {
	data := &ExecutiveSummaryData{}

	// Overall compliance score
	_ = e.pool.QueryRow(ctx, `
		SELECT COALESCE(AVG(compliance_score), 0)
		FROM v_compliance_score_by_framework
		WHERE organization_id = $1`, orgID,
	).Scan(&data.OverallComplianceScore)

	// Overall risk score
	_ = e.pool.QueryRow(ctx, `
		SELECT COALESCE(AVG(residual_risk_score), 0)
		FROM risks WHERE organization_id = $1 AND deleted_at IS NULL`, orgID,
	).Scan(&data.OverallRiskScore)

	// Critical and high risk counts
	_ = e.pool.QueryRow(ctx, `
		SELECT COUNT(*) FILTER (WHERE residual_risk_level = 'critical'),
			COUNT(*) FILTER (WHERE residual_risk_level = 'high')
		FROM risks WHERE organization_id = $1 AND deleted_at IS NULL`, orgID,
	).Scan(&data.CriticalRiskCount, &data.HighRiskCount)

	// Open incidents
	_ = e.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM incidents
		WHERE organization_id = $1 AND status NOT IN ('resolved', 'closed') AND deleted_at IS NULL`, orgID,
	).Scan(&data.OpenIncidents)

	// Data breach count
	_ = e.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM incidents
		WHERE organization_id = $1 AND is_data_breach = true AND deleted_at IS NULL`, orgID,
	).Scan(&data.DataBreachCount)

	// Open audit findings
	_ = e.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM audit_findings
		WHERE organization_id = $1 AND status NOT IN ('remediated', 'closed', 'accepted') AND deleted_at IS NULL`, orgID,
	).Scan(&data.AuditFindingsOpen)

	// Policy compliance rate (% published or approved)
	var totalPolicies, compliantPolicies int
	_ = e.pool.QueryRow(ctx, `
		SELECT COUNT(*),
			COUNT(*) FILTER (WHERE status IN ('published', 'approved'))
		FROM policies WHERE organization_id = $1 AND deleted_at IS NULL`, orgID,
	).Scan(&totalPolicies, &compliantPolicies)
	if totalPolicies > 0 {
		data.PolicyComplianceRate = float64(compliantPolicies) / float64(totalPolicies) * 100
	}

	// Top risks
	topRows, _ := e.pool.Query(ctx, `
		SELECT r.risk_ref, r.title, r.residual_risk_score, r.residual_risk_level::TEXT,
			COALESCE(u.first_name || ' ' || u.last_name, 'Unassigned'),
			(SELECT COUNT(*) FROM risk_treatments rt WHERE rt.risk_id = r.id AND rt.status NOT IN ('completed','cancelled'))
		FROM risks r LEFT JOIN users u ON r.owner_user_id = u.id
		WHERE r.organization_id = $1 AND r.deleted_at IS NULL
		ORDER BY r.residual_risk_score DESC LIMIT 5`, orgID)
	if topRows != nil {
		defer topRows.Close()
		for topRows.Next() {
			var t TopRiskEntry
			_ = topRows.Scan(&t.Ref, &t.Title, &t.ResidualScore, &t.Level, &t.Owner, &t.TreatmentCount)
			data.TopRisks = append(data.TopRisks, t)
		}
	}

	// Compliance by framework
	fwRows, _ := e.pool.Query(ctx, `
		SELECT framework_code, framework_name, COALESCE(compliance_score, 0),
			total_controls, implemented,
			(total_controls - implemented - not_applicable) AS gaps,
			COALESCE(maturity_avg, 0)
		FROM v_compliance_score_by_framework WHERE organization_id = $1`, orgID)
	if fwRows != nil {
		defer fwRows.Close()
		for fwRows.Next() {
			var f FrameworkReportSummary
			_ = fwRows.Scan(&f.Code, &f.Name, &f.Score, &f.TotalControls,
				&f.Implemented, &f.Gaps, &f.MaturityAvg)
			data.ComplianceByFramework = append(data.ComplianceByFramework, f)
		}
	}

	// Treatment progress
	_ = e.pool.QueryRow(ctx, `
		SELECT COUNT(*),
			COUNT(*) FILTER (WHERE status = 'completed'),
			COUNT(*) FILTER (WHERE status = 'in_progress'),
			COUNT(*) FILTER (WHERE status != 'completed' AND target_date < CURRENT_DATE)
		FROM risk_treatments WHERE organization_id = $1`, orgID,
	).Scan(&data.TreatmentProgress.Total, &data.TreatmentProgress.Completed,
		&data.TreatmentProgress.InProgress, &data.TreatmentProgress.Overdue)

	// Build key metrics
	data.KeyMetrics = e.buildKeyMetrics(data)

	return data, nil
}

// GenerateGapAnalysis produces a gap analysis with remediation roadmap for a framework.
func (e *ReportEngine) GenerateGapAnalysis(ctx context.Context, orgID, frameworkID uuid.UUID) (*GapAnalysisData, error) {
	data := &GapAnalysisData{}

	// Framework info
	err := e.pool.QueryRow(ctx, `
		SELECT cf.code, cf.name
		FROM organization_frameworks of2
		JOIN compliance_frameworks cf ON of2.framework_id = cf.id
		WHERE of2.organization_id = $1 AND cf.id = $2`, orgID, frameworkID,
	).Scan(&data.FrameworkCode, &data.FrameworkName)
	if err != nil {
		return nil, fmt.Errorf("framework not found or not adopted: %w", err)
	}

	// Overall gap stats
	_ = e.pool.QueryRow(ctx, `
		SELECT COUNT(*),
			COUNT(*) FILTER (WHERE ci.status IN ('implemented', 'effective')),
			COUNT(*) FILTER (WHERE ci.status NOT IN ('implemented', 'effective', 'not_applicable'))
		FROM control_implementations ci
		JOIN organization_frameworks of2 ON ci.org_framework_id = of2.id
		WHERE ci.organization_id = $1 AND of2.framework_id = $2 AND ci.deleted_at IS NULL`,
		orgID, frameworkID,
	).Scan(&data.TotalControls, &data.ImplementedCount, &data.GapCount)

	if data.TotalControls > 0 {
		data.ComplianceScore = float64(data.ImplementedCount) / float64(data.TotalControls) * 100
	}

	// Gaps by domain
	domainRows, _ := e.pool.Query(ctx, `
		SELECT fd.code, fd.name,
			COUNT(*),
			COUNT(*) FILTER (WHERE ci.status IN ('implemented', 'effective')),
			COUNT(*) FILTER (WHERE ci.status NOT IN ('implemented', 'effective', 'not_applicable'))
		FROM control_implementations ci
		JOIN framework_controls fc ON ci.framework_control_id = fc.id
		JOIN framework_domains fd ON fc.domain_id = fd.id
		JOIN organization_frameworks of2 ON ci.org_framework_id = of2.id
		WHERE ci.organization_id = $1 AND of2.framework_id = $2 AND ci.deleted_at IS NULL
		GROUP BY fd.code, fd.name ORDER BY fd.code`, orgID, frameworkID)
	if domainRows != nil {
		defer domainRows.Close()
		for domainRows.Next() {
			var d DomainGapSummary
			_ = domainRows.Scan(&d.DomainCode, &d.DomainName, &d.TotalControls, &d.Implemented, &d.Gaps)
			if d.TotalControls > 0 {
				d.ComplianceScore = float64(d.Implemented) / float64(d.TotalControls) * 100
			}
			data.GapsByDomain = append(data.GapsByDomain, d)
		}
	}

	// Remediation roadmap
	remRows, _ := e.pool.Query(ctx, `
		SELECT fc.code, fc.title, COALESCE(ci.gap_description, ''),
			COALESCE(ci.risk_if_not_implemented, 'medium'),
			COALESCE(u.first_name || ' ' || u.last_name, 'Unassigned'),
			COALESCE(ci.remediation_due_date::TEXT, 'Not set')
		FROM control_implementations ci
		JOIN framework_controls fc ON ci.framework_control_id = fc.id
		JOIN organization_frameworks of2 ON ci.org_framework_id = of2.id
		LEFT JOIN users u ON ci.owner_user_id = u.id
		WHERE ci.organization_id = $1 AND of2.framework_id = $2
			AND ci.status NOT IN ('implemented', 'effective', 'not_applicable')
			AND ci.deleted_at IS NULL
		ORDER BY CASE ci.risk_if_not_implemented
			WHEN 'critical' THEN 1 WHEN 'high' THEN 2 WHEN 'medium' THEN 3 ELSE 4 END`,
		orgID, frameworkID)
	if remRows != nil {
		defer remRows.Close()
		priority := 1
		for remRows.Next() {
			var r RemediationItem
			_ = remRows.Scan(&r.ControlCode, &r.ControlTitle, &r.GapDescription,
				&r.RiskLevel, &r.Owner, &r.DueDate)
			r.Priority = priority
			priority++
			data.RemediationPlan = append(data.RemediationPlan, r)
		}
	}

	return data, nil
}

// GenerateCrossFrameworkReport produces a cross-framework mapping coverage report.
func (e *ReportEngine) GenerateCrossFrameworkReport(ctx context.Context, orgID uuid.UUID) (*CrossFrameworkData, error) {
	data := &CrossFrameworkData{}

	// Adopted frameworks with mapping stats
	fwRows, _ := e.pool.Query(ctx, `
		SELECT cf.code, cf.name, cf.total_controls,
			(SELECT COUNT(DISTINCT fcm.source_control_id) FROM framework_control_mappings fcm
			 JOIN framework_controls fc2 ON fcm.source_control_id = fc2.id
			 WHERE fc2.framework_id = cf.id)
		FROM organization_frameworks of2
		JOIN compliance_frameworks cf ON of2.framework_id = cf.id
		WHERE of2.organization_id = $1
		ORDER BY cf.code`, orgID)
	if fwRows != nil {
		defer fwRows.Close()
		for fwRows.Next() {
			var f FrameworkCoverageSummary
			_ = fwRows.Scan(&f.Code, &f.Name, &f.TotalControls, &f.MappedControls)
			if f.TotalControls > 0 {
				f.CoveragePercent = float64(f.MappedControls) / float64(f.TotalControls) * 100
			}
			data.Frameworks = append(data.Frameworks, f)
		}
	}

	// Mapping rows
	mapRows, _ := e.pool.Query(ctx, `
		SELECT scf.code, sfc.code, tcf.code, tfc.code, fcm.mapping_type::TEXT, fcm.mapping_strength
		FROM framework_control_mappings fcm
		JOIN framework_controls sfc ON fcm.source_control_id = sfc.id
		JOIN compliance_frameworks scf ON sfc.framework_id = scf.id
		JOIN framework_controls tfc ON fcm.target_control_id = tfc.id
		JOIN compliance_frameworks tcf ON tfc.framework_id = tcf.id
		WHERE scf.id IN (
			SELECT cf.id FROM organization_frameworks of2
			JOIN compliance_frameworks cf ON of2.framework_id = cf.id
			WHERE of2.organization_id = $1
		)
		ORDER BY scf.code, sfc.code LIMIT 500`, orgID)
	if mapRows != nil {
		defer mapRows.Close()
		for mapRows.Next() {
			var m MappingRow
			_ = mapRows.Scan(&m.SourceFramework, &m.SourceControl, &m.TargetFramework,
				&m.TargetControl, &m.MappingType, &m.Strength)
			data.MappingMatrix = append(data.MappingMatrix, m)
		}
	}

	// Calculate aggregate stats
	data.SharedControls = len(data.MappingMatrix)
	var totalControls, totalMapped int
	for _, f := range data.Frameworks {
		totalControls += f.TotalControls
		totalMapped += f.MappedControls
	}
	data.UniqueControls = totalControls - data.SharedControls
	if data.UniqueControls < 0 {
		data.UniqueControls = 0
	}
	if totalControls > 0 {
		data.CoveragePercentage = float64(totalMapped) / float64(totalControls) * 100
	}

	return data, nil
}

// ============================================================
// DEFINITION CRUD
// ============================================================

// GetDefinition retrieves a report definition by ID.
func (e *ReportEngine) GetDefinition(ctx context.Context, orgID, defID uuid.UUID) (*ReportDefinition, error) {
	var def ReportDefinition
	err := e.pool.QueryRow(ctx, `
		SELECT id, organization_id, name, description, report_type::TEXT, format::TEXT,
			filters, sections, classification, include_executive_summary, include_appendices,
			branding, created_by, is_template, created_at, updated_at
		FROM report_definitions
		WHERE id = $1 AND organization_id = $2`, defID, orgID,
	).Scan(
		&def.ID, &def.OrganizationID, &def.Name, &def.Description,
		&def.ReportType, &def.Format, &def.Filters, &def.Sections,
		&def.Classification, &def.IncludeExecutiveSummary, &def.IncludeAppendices,
		&def.Branding, &def.CreatedBy, &def.IsTemplate, &def.CreatedAt, &def.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("report definition not found: %w", err)
	}
	return &def, nil
}

// ListDefinitions returns all report definitions for an organization.
func (e *ReportEngine) ListDefinitions(ctx context.Context, orgID uuid.UUID, page, pageSize int) ([]ReportDefinition, int64, error) {
	var total int64
	_ = e.pool.QueryRow(ctx, `SELECT COUNT(*) FROM report_definitions WHERE organization_id = $1`, orgID).Scan(&total)

	offset := (page - 1) * pageSize
	rows, err := e.pool.Query(ctx, `
		SELECT id, organization_id, name, description, report_type::TEXT, format::TEXT,
			filters, sections, classification, include_executive_summary, include_appendices,
			branding, created_by, is_template, created_at, updated_at
		FROM report_definitions
		WHERE organization_id = $1
		ORDER BY updated_at DESC
		LIMIT $2 OFFSET $3`, orgID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var defs []ReportDefinition
	for rows.Next() {
		var def ReportDefinition
		if err := rows.Scan(
			&def.ID, &def.OrganizationID, &def.Name, &def.Description,
			&def.ReportType, &def.Format, &def.Filters, &def.Sections,
			&def.Classification, &def.IncludeExecutiveSummary, &def.IncludeAppendices,
			&def.Branding, &def.CreatedBy, &def.IsTemplate, &def.CreatedAt, &def.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		defs = append(defs, def)
	}
	return defs, total, nil
}

// CreateDefinition saves a new report definition.
func (e *ReportEngine) CreateDefinition(ctx context.Context, orgID, userID uuid.UUID, req CreateReportDefinitionRequest) (*ReportDefinition, error) {
	classification := req.Classification
	if classification == "" {
		classification = "internal"
	}
	filters := req.Filters
	if filters == nil {
		filters = json.RawMessage(`{}`)
	}
	sections := req.Sections
	if sections == nil {
		sections = json.RawMessage(`[]`)
	}
	branding := req.Branding
	if branding == nil {
		branding = json.RawMessage(`{}`)
	}

	var def ReportDefinition
	err := e.pool.QueryRow(ctx, `
		INSERT INTO report_definitions (organization_id, name, description, report_type, format,
			filters, sections, classification, include_executive_summary, include_appendices,
			branding, created_by, is_template)
		VALUES ($1, $2, $3, $4::report_type, $5::report_format, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, organization_id, name, description, report_type::TEXT, format::TEXT,
			filters, sections, classification, include_executive_summary, include_appendices,
			branding, created_by, is_template, created_at, updated_at`,
		orgID, req.Name, req.Description, req.ReportType, req.Format,
		filters, sections, classification,
		req.IncludeExecutiveSummary, req.IncludeAppendices, branding, userID, req.IsTemplate,
	).Scan(
		&def.ID, &def.OrganizationID, &def.Name, &def.Description,
		&def.ReportType, &def.Format, &def.Filters, &def.Sections,
		&def.Classification, &def.IncludeExecutiveSummary, &def.IncludeAppendices,
		&def.Branding, &def.CreatedBy, &def.IsTemplate, &def.CreatedAt, &def.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create report definition: %w", err)
	}
	return &def, nil
}

// UpdateDefinition updates an existing report definition.
func (e *ReportEngine) UpdateDefinition(ctx context.Context, orgID, defID uuid.UUID, req UpdateReportDefinitionRequest) (*ReportDefinition, error) {
	// Build dynamic update
	setClauses := []string{}
	args := []interface{}{defID, orgID}
	argIdx := 3

	if req.Name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, *req.Name)
		argIdx++
	}
	if req.Description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", argIdx))
		args = append(args, *req.Description)
		argIdx++
	}
	if req.Format != nil {
		setClauses = append(setClauses, fmt.Sprintf("format = $%d::report_format", argIdx))
		args = append(args, *req.Format)
		argIdx++
	}
	if req.Filters != nil {
		setClauses = append(setClauses, fmt.Sprintf("filters = $%d", argIdx))
		args = append(args, *req.Filters)
		argIdx++
	}
	if req.Sections != nil {
		setClauses = append(setClauses, fmt.Sprintf("sections = $%d", argIdx))
		args = append(args, *req.Sections)
		argIdx++
	}
	if req.Classification != nil {
		setClauses = append(setClauses, fmt.Sprintf("classification = $%d", argIdx))
		args = append(args, *req.Classification)
		argIdx++
	}
	if req.IncludeExecutiveSummary != nil {
		setClauses = append(setClauses, fmt.Sprintf("include_executive_summary = $%d", argIdx))
		args = append(args, *req.IncludeExecutiveSummary)
		argIdx++
	}
	if req.IncludeAppendices != nil {
		setClauses = append(setClauses, fmt.Sprintf("include_appendices = $%d", argIdx))
		args = append(args, *req.IncludeAppendices)
		argIdx++
	}
	if req.Branding != nil {
		setClauses = append(setClauses, fmt.Sprintf("branding = $%d", argIdx))
		args = append(args, *req.Branding)
		argIdx++
	}

	if len(setClauses) == 0 {
		return e.GetDefinition(ctx, orgID, defID)
	}

	query := "UPDATE report_definitions SET "
	for i, clause := range setClauses {
		if i > 0 {
			query += ", "
		}
		query += clause
	}
	query += " WHERE id = $1 AND organization_id = $2"

	_, err := e.pool.Exec(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update report definition: %w", err)
	}

	return e.GetDefinition(ctx, orgID, defID)
}

// DeleteDefinition removes a report definition.
func (e *ReportEngine) DeleteDefinition(ctx context.Context, orgID, defID uuid.UUID) error {
	tag, err := e.pool.Exec(ctx, `DELETE FROM report_definitions WHERE id = $1 AND organization_id = $2`, defID, orgID)
	if err != nil {
		return fmt.Errorf("failed to delete report definition: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("report definition not found")
	}
	return nil
}

// ============================================================
// SCHEDULE CRUD
// ============================================================

// ListSchedules returns all report schedules for an organization.
func (e *ReportEngine) ListSchedules(ctx context.Context, orgID uuid.UUID) ([]ReportSchedule, error) {
	rows, err := e.pool.Query(ctx, `
		SELECT id, organization_id, report_definition_id, name, frequency::TEXT,
			day_of_week, day_of_month, time_of_day::TEXT, timezone,
			recipient_user_ids, recipient_emails, delivery_channel::TEXT,
			is_active, last_run_at, next_run_at, created_at, updated_at
		FROM report_schedules
		WHERE organization_id = $1
		ORDER BY created_at DESC`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []ReportSchedule
	for rows.Next() {
		var s ReportSchedule
		if err := rows.Scan(
			&s.ID, &s.OrganizationID, &s.ReportDefinitionID, &s.Name, &s.Frequency,
			&s.DayOfWeek, &s.DayOfMonth, &s.TimeOfDay, &s.Timezone,
			&s.RecipientUserIDs, &s.RecipientEmails, &s.DeliveryChannel,
			&s.IsActive, &s.LastRunAt, &s.NextRunAt, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		schedules = append(schedules, s)
	}
	return schedules, nil
}

// CreateSchedule creates a new report schedule.
func (e *ReportEngine) CreateSchedule(ctx context.Context, orgID uuid.UUID, req CreateScheduleRequest) (*ReportSchedule, error) {
	// Validate definition exists
	_, err := e.GetDefinition(ctx, orgID, req.ReportDefinitionID)
	if err != nil {
		return nil, fmt.Errorf("report definition not found: %w", err)
	}

	timeOfDay := req.TimeOfDay
	if timeOfDay == "" {
		timeOfDay = "08:00"
	}
	tz := req.Timezone
	if tz == "" {
		tz = "Europe/London"
	}
	channel := req.DeliveryChannel
	if channel == "" {
		channel = "both"
	}

	nextRun := CalculateNextRun(req.Frequency, req.DayOfWeek, req.DayOfMonth, timeOfDay, tz)

	var s ReportSchedule
	err = e.pool.QueryRow(ctx, `
		INSERT INTO report_schedules (organization_id, report_definition_id, name, frequency,
			day_of_week, day_of_month, time_of_day, timezone,
			recipient_user_ids, recipient_emails, delivery_channel, is_active, next_run_at)
		VALUES ($1, $2, $3, $4::report_frequency, $5, $6, $7::TIME, $8, $9, $10, $11::report_delivery_channel, true, $12)
		RETURNING id, organization_id, report_definition_id, name, frequency::TEXT,
			day_of_week, day_of_month, time_of_day::TEXT, timezone,
			recipient_user_ids, recipient_emails, delivery_channel::TEXT,
			is_active, last_run_at, next_run_at, created_at, updated_at`,
		orgID, req.ReportDefinitionID, req.Name, req.Frequency,
		req.DayOfWeek, req.DayOfMonth, timeOfDay, tz,
		req.RecipientUserIDs, req.RecipientEmails, channel, nextRun,
	).Scan(
		&s.ID, &s.OrganizationID, &s.ReportDefinitionID, &s.Name, &s.Frequency,
		&s.DayOfWeek, &s.DayOfMonth, &s.TimeOfDay, &s.Timezone,
		&s.RecipientUserIDs, &s.RecipientEmails, &s.DeliveryChannel,
		&s.IsActive, &s.LastRunAt, &s.NextRunAt, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create report schedule: %w", err)
	}
	return &s, nil
}

// UpdateSchedule updates an existing report schedule.
func (e *ReportEngine) UpdateSchedule(ctx context.Context, orgID, schedID uuid.UUID, req UpdateScheduleRequest) (*ReportSchedule, error) {
	setClauses := []string{}
	args := []interface{}{schedID, orgID}
	argIdx := 3

	if req.Name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, *req.Name)
		argIdx++
	}
	if req.Frequency != nil {
		setClauses = append(setClauses, fmt.Sprintf("frequency = $%d::report_frequency", argIdx))
		args = append(args, *req.Frequency)
		argIdx++
	}
	if req.DayOfWeek != nil {
		setClauses = append(setClauses, fmt.Sprintf("day_of_week = $%d", argIdx))
		args = append(args, *req.DayOfWeek)
		argIdx++
	}
	if req.DayOfMonth != nil {
		setClauses = append(setClauses, fmt.Sprintf("day_of_month = $%d", argIdx))
		args = append(args, *req.DayOfMonth)
		argIdx++
	}
	if req.TimeOfDay != nil {
		setClauses = append(setClauses, fmt.Sprintf("time_of_day = $%d::TIME", argIdx))
		args = append(args, *req.TimeOfDay)
		argIdx++
	}
	if req.Timezone != nil {
		setClauses = append(setClauses, fmt.Sprintf("timezone = $%d", argIdx))
		args = append(args, *req.Timezone)
		argIdx++
	}
	if req.RecipientUserIDs != nil {
		setClauses = append(setClauses, fmt.Sprintf("recipient_user_ids = $%d", argIdx))
		args = append(args, *req.RecipientUserIDs)
		argIdx++
	}
	if req.RecipientEmails != nil {
		setClauses = append(setClauses, fmt.Sprintf("recipient_emails = $%d", argIdx))
		args = append(args, *req.RecipientEmails)
		argIdx++
	}
	if req.DeliveryChannel != nil {
		setClauses = append(setClauses, fmt.Sprintf("delivery_channel = $%d::report_delivery_channel", argIdx))
		args = append(args, *req.DeliveryChannel)
		argIdx++
	}
	if req.IsActive != nil {
		setClauses = append(setClauses, fmt.Sprintf("is_active = $%d", argIdx))
		args = append(args, *req.IsActive)
		argIdx++
	}

	if len(setClauses) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}

	query := "UPDATE report_schedules SET "
	for i, clause := range setClauses {
		if i > 0 {
			query += ", "
		}
		query += clause
	}
	query += " WHERE id = $1 AND organization_id = $2"

	_, err := e.pool.Exec(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update schedule: %w", err)
	}

	// Fetch updated schedule
	var s ReportSchedule
	err = e.pool.QueryRow(ctx, `
		SELECT id, organization_id, report_definition_id, name, frequency::TEXT,
			day_of_week, day_of_month, time_of_day::TEXT, timezone,
			recipient_user_ids, recipient_emails, delivery_channel::TEXT,
			is_active, last_run_at, next_run_at, created_at, updated_at
		FROM report_schedules WHERE id = $1 AND organization_id = $2`, schedID, orgID,
	).Scan(
		&s.ID, &s.OrganizationID, &s.ReportDefinitionID, &s.Name, &s.Frequency,
		&s.DayOfWeek, &s.DayOfMonth, &s.TimeOfDay, &s.Timezone,
		&s.RecipientUserIDs, &s.RecipientEmails, &s.DeliveryChannel,
		&s.IsActive, &s.LastRunAt, &s.NextRunAt, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("schedule not found: %w", err)
	}
	return &s, nil
}

// DeleteSchedule removes a report schedule.
func (e *ReportEngine) DeleteSchedule(ctx context.Context, orgID, schedID uuid.UUID) error {
	tag, err := e.pool.Exec(ctx, `DELETE FROM report_schedules WHERE id = $1 AND organization_id = $2`, schedID, orgID)
	if err != nil {
		return fmt.Errorf("failed to delete schedule: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("schedule not found")
	}
	return nil
}

// ============================================================
// REPORT RUNS
// ============================================================

// GetRun retrieves a report run by ID.
func (e *ReportEngine) GetRun(ctx context.Context, orgID, runID uuid.UUID) (*ReportRun, error) {
	var run ReportRun
	err := e.pool.QueryRow(ctx, `
		SELECT id, organization_id, report_definition_id, schedule_id, status::TEXT,
			format::TEXT, COALESCE(file_path, ''), COALESCE(file_size_bytes, 0),
			COALESCE(page_count, 0), COALESCE(generation_time_ms, 0),
			COALESCE(parameters, '{}'), generated_by, COALESCE(error_message, ''),
			created_at, completed_at
		FROM report_runs WHERE id = $1 AND organization_id = $2`, runID, orgID,
	).Scan(
		&run.ID, &run.OrganizationID, &run.ReportDefinitionID, &run.ScheduleID,
		&run.Status, &run.Format, &run.FilePath, &run.FileSizeBytes,
		&run.PageCount, &run.GenerationTimeMs, &run.Parameters,
		&run.GeneratedBy, &run.ErrorMessage, &run.CreatedAt, &run.CompletedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("report run not found")
		}
		return nil, err
	}
	return &run, nil
}

// ListRuns returns paginated report runs for an organization.
func (e *ReportEngine) ListRuns(ctx context.Context, orgID uuid.UUID, page, pageSize int) ([]ReportRun, int64, error) {
	var total int64
	_ = e.pool.QueryRow(ctx, `SELECT COUNT(*) FROM report_runs WHERE organization_id = $1`, orgID).Scan(&total)

	offset := (page - 1) * pageSize
	rows, err := e.pool.Query(ctx, `
		SELECT id, organization_id, report_definition_id, schedule_id, status::TEXT,
			format::TEXT, COALESCE(file_path, ''), COALESCE(file_size_bytes, 0),
			COALESCE(page_count, 0), COALESCE(generation_time_ms, 0),
			COALESCE(parameters, '{}'), generated_by, COALESCE(error_message, ''),
			created_at, completed_at
		FROM report_runs WHERE organization_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`, orgID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var runs []ReportRun
	for rows.Next() {
		var run ReportRun
		if err := rows.Scan(
			&run.ID, &run.OrganizationID, &run.ReportDefinitionID, &run.ScheduleID,
			&run.Status, &run.Format, &run.FilePath, &run.FileSizeBytes,
			&run.PageCount, &run.GenerationTimeMs, &run.Parameters,
			&run.GeneratedBy, &run.ErrorMessage, &run.CreatedAt, &run.CompletedAt,
		); err != nil {
			return nil, 0, err
		}
		runs = append(runs, run)
	}
	return runs, total, nil
}

// ============================================================
// HELPERS
// ============================================================

// GetFileReader returns an io.ReadCloser for a stored report file.
func (e *ReportEngine) GetFileReader(_ context.Context, filePath string) (io.ReadCloser, error) {
	return e.storage.Retrieve(filePath)
}

func (e *ReportEngine) getOrgName(ctx context.Context, orgID uuid.UUID) string {
	var name string
	_ = e.pool.QueryRow(ctx, `SELECT name FROM organizations WHERE id = $1`, orgID).Scan(&name)
	if name == "" {
		return "Unknown Organisation"
	}
	return name
}

func (e *ReportEngine) getUserName(ctx context.Context, userID uuid.UUID) string {
	var name string
	_ = e.pool.QueryRow(ctx, `SELECT first_name || ' ' || last_name FROM users WHERE id = $1`, userID).Scan(&name)
	if name == "" {
		return "System"
	}
	return name
}

func (e *ReportEngine) generateAuditSummary(ctx context.Context, orgID uuid.UUID, filters ReportFilters) (interface{}, error) {
	type AuditSummaryRow struct {
		AuditRef    string `json:"audit_ref"`
		Title       string `json:"title"`
		AuditType   string `json:"audit_type"`
		Status      string `json:"status"`
		Findings    int    `json:"total_findings"`
		Critical    int    `json:"critical_findings"`
		High        int    `json:"high_findings"`
		LeadAuditor string `json:"lead_auditor"`
	}

	rows, err := e.pool.Query(ctx, `
		SELECT a.audit_ref, a.title, a.audit_type, a.status::TEXT,
			a.total_findings, a.critical_findings, a.high_findings,
			COALESCE(u.first_name || ' ' || u.last_name, 'Unassigned')
		FROM audits a LEFT JOIN users u ON a.lead_auditor_id = u.id
		WHERE a.organization_id = $1 AND a.deleted_at IS NULL
		ORDER BY a.created_at DESC LIMIT 50`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var audits []AuditSummaryRow
	for rows.Next() {
		var a AuditSummaryRow
		_ = rows.Scan(&a.AuditRef, &a.Title, &a.AuditType, &a.Status,
			&a.Findings, &a.Critical, &a.High, &a.LeadAuditor)
		audits = append(audits, a)
	}
	return audits, nil
}

func (e *ReportEngine) buildKeyMetrics(data *ExecutiveSummaryData) []KeyMetric {
	metrics := []KeyMetric{
		{
			Name:   "Overall Compliance",
			Value:  fmt.Sprintf("%.1f%%", data.OverallComplianceScore),
			Trend:  "stable",
			Status: complianceRAG(data.OverallComplianceScore),
		},
		{
			Name:   "Average Risk Score",
			Value:  fmt.Sprintf("%.1f", data.OverallRiskScore),
			Trend:  "stable",
			Status: riskScoreRAG(data.OverallRiskScore),
		},
		{
			Name:   "Critical Risks",
			Value:  fmt.Sprintf("%d", data.CriticalRiskCount),
			Trend:  "stable",
			Status: countRAG(data.CriticalRiskCount, 0, 1, 3),
		},
		{
			Name:   "Open Incidents",
			Value:  fmt.Sprintf("%d", data.OpenIncidents),
			Trend:  "stable",
			Status: countRAG(data.OpenIncidents, 0, 3, 5),
		},
		{
			Name:   "Open Audit Findings",
			Value:  fmt.Sprintf("%d", data.AuditFindingsOpen),
			Trend:  "stable",
			Status: countRAG(data.AuditFindingsOpen, 0, 5, 10),
		},
		{
			Name:   "Policy Compliance",
			Value:  fmt.Sprintf("%.1f%%", data.PolicyComplianceRate),
			Trend:  "stable",
			Status: complianceRAG(data.PolicyComplianceRate),
		},
	}

	if data.TreatmentProgress.Total > 0 {
		rate := float64(data.TreatmentProgress.Completed) / float64(data.TreatmentProgress.Total) * 100
		metrics = append(metrics, KeyMetric{
			Name:   "Treatment Completion",
			Value:  fmt.Sprintf("%.0f%%", rate),
			Trend:  "stable",
			Status: complianceRAG(rate),
		})
	}

	return metrics
}

func complianceRAG(score float64) string {
	if score >= 80 {
		return "green"
	}
	if score >= 60 {
		return "amber"
	}
	return "red"
}

func riskScoreRAG(score float64) string {
	if score <= 6 {
		return "green"
	}
	if score <= 12 {
		return "amber"
	}
	return "red"
}

func countRAG(count, greenMax, amberMax, redThreshold int) string {
	if count <= greenMax {
		return "green"
	}
	if count <= amberMax {
		return "amber"
	}
	_ = redThreshold
	return "red"
}

// renderComplianceCSV produces a CSV representation of compliance data.
func (e *ReportEngine) renderComplianceCSV(data *ComplianceReportData) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString("Framework Code,Framework Name,Compliance Score %,Total Controls,Implemented,Gaps,Maturity Avg\n")
	for _, f := range data.FrameworkSummaries {
		buf.WriteString(fmt.Sprintf("%s,%s,%.1f,%d,%d,%d,%.1f\n",
			csvEscape(f.Code), csvEscape(f.Name), f.Score,
			f.TotalControls, f.Implemented, f.Gaps, f.MaturityAvg))
	}
	return buf.Bytes(), nil
}

// renderRiskCSV produces a CSV representation of risk data.
func (e *ReportEngine) renderRiskCSV(data *RiskReportData) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString("Risk Ref,Title,Residual Score,Risk Level,Owner,Open Treatments\n")
	for _, r := range data.TopRisks {
		buf.WriteString(fmt.Sprintf("%s,%s,%.1f,%s,%s,%d\n",
			csvEscape(r.Ref), csvEscape(r.Title), r.ResidualScore,
			csvEscape(r.Level), csvEscape(r.Owner), r.TreatmentCount))
	}
	return buf.Bytes(), nil
}

func csvEscape(s string) string {
	// Wrap in quotes if it contains commas, quotes, or newlines
	for _, c := range s {
		if c == ',' || c == '"' || c == '\n' {
			escaped := ""
			for _, ch := range s {
				if ch == '"' {
					escaped += `""`
				} else {
					escaped += string(ch)
				}
			}
			return `"` + escaped + `"`
		}
	}
	return s
}

// CalculateNextRun determines the next execution time for a schedule.
func CalculateNextRun(frequency string, dayOfWeek, dayOfMonth *int, timeOfDay, timezone string) time.Time {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}

	now := time.Now().In(loc)

	// Parse time of day
	hour, minute := 8, 0
	if len(timeOfDay) >= 5 {
		fmt.Sscanf(timeOfDay, "%d:%d", &hour, &minute)
	}

	// Calculate base next run
	var next time.Time
	switch frequency {
	case "daily":
		next = time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, loc)
		if next.Before(now) {
			next = next.AddDate(0, 0, 1)
		}

	case "weekly":
		dow := 1 // Monday default
		if dayOfWeek != nil {
			dow = *dayOfWeek
		}
		next = time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, loc)
		for int(next.Weekday()) != dow || next.Before(now) {
			next = next.AddDate(0, 0, 1)
		}

	case "monthly":
		dom := 1
		if dayOfMonth != nil {
			dom = *dayOfMonth
		}
		next = time.Date(now.Year(), now.Month(), dom, hour, minute, 0, 0, loc)
		if next.Before(now) {
			next = next.AddDate(0, 1, 0)
		}

	case "quarterly":
		dom := 1
		if dayOfMonth != nil {
			dom = *dayOfMonth
		}
		// Next quarter start month
		currentQ := (int(now.Month()) - 1) / 3
		nextQMonth := time.Month((currentQ+1)*3 + 1)
		nextQYear := now.Year()
		if nextQMonth > 12 {
			nextQMonth -= 12
			nextQYear++
		}
		next = time.Date(nextQYear, nextQMonth, dom, hour, minute, 0, 0, loc)
		if next.Before(now) {
			next = next.AddDate(0, 3, 0)
		}

	case "annually":
		dom := 1
		if dayOfMonth != nil {
			dom = *dayOfMonth
		}
		next = time.Date(now.Year(), time.January, dom, hour, minute, 0, 0, loc)
		if next.Before(now) {
			next = next.AddDate(1, 0, 0)
		}

	default:
		next = now.AddDate(0, 0, 1)
	}

	return next.UTC()
}

// GetDueSchedules returns all active schedules that are due for execution.
func (e *ReportEngine) GetDueSchedules(ctx context.Context) ([]ReportSchedule, error) {
	rows, err := e.pool.Query(ctx, `
		SELECT id, organization_id, report_definition_id, name, frequency::TEXT,
			day_of_week, day_of_month, time_of_day::TEXT, timezone,
			recipient_user_ids, recipient_emails, delivery_channel::TEXT,
			is_active, last_run_at, next_run_at, created_at, updated_at
		FROM report_schedules
		WHERE is_active = true AND next_run_at <= NOW()
		ORDER BY next_run_at ASC
		LIMIT 50`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []ReportSchedule
	for rows.Next() {
		var s ReportSchedule
		if err := rows.Scan(
			&s.ID, &s.OrganizationID, &s.ReportDefinitionID, &s.Name, &s.Frequency,
			&s.DayOfWeek, &s.DayOfMonth, &s.TimeOfDay, &s.Timezone,
			&s.RecipientUserIDs, &s.RecipientEmails, &s.DeliveryChannel,
			&s.IsActive, &s.LastRunAt, &s.NextRunAt, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		schedules = append(schedules, s)
	}
	return schedules, nil
}

// AdvanceSchedule updates last_run_at and calculates the next_run_at.
func (e *ReportEngine) AdvanceSchedule(ctx context.Context, schedule *ReportSchedule) error {
	nextRun := CalculateNextRun(schedule.Frequency, schedule.DayOfWeek, schedule.DayOfMonth,
		schedule.TimeOfDay, schedule.Timezone)

	_, err := e.pool.Exec(ctx, `
		UPDATE report_schedules
		SET last_run_at = NOW(), next_run_at = $2
		WHERE id = $1`, schedule.ID, nextRun)
	return err
}
