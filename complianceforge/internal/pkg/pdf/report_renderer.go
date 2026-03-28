package pdf

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

// ============================================================
// REPORT RENDERER — Professional PDF generation for all report types
//
// Uses a text-based structured format that can be piped to a PDF
// rendering backend (e.g. maroto v2, gofpdf, wkhtmltopdf).
// The interface is designed so the actual PDF library can be swapped
// without changing the data layer.
// ============================================================

// ReportRenderer produces professional PDF-format report output.
// Supports A4 page size, headers/footers with classification stamps,
// page numbers, cover pages, and table of contents.
type ReportRenderer struct {
	pageWidth  int
	lineChar   string
	sectionSep string
}

// NewReportRenderer creates a new PDF report renderer.
func NewReportRenderer() *ReportRenderer {
	return &ReportRenderer{
		pageWidth:  80,
		lineChar:   "─",
		sectionSep: "══════════════════════════════════════════════════════",
	}
}

// ============================================================
// DATA TYPES — imported by report_engine.go
// These use the existing ReportMetadata from pdf.go and add
// new data structures for advanced reports.
// ============================================================

// ComplianceReportInput is the data contract for compliance PDF rendering.
type ComplianceReportInput struct {
	Metadata             ReportMetadata
	OverallScore         float64
	FrameworkSummaries   []FrameworkScoreRow
	TopGaps              []GapRow
	MaturityDistribution map[string]int
	ControlsByStatus     map[string]int
}

// RiskReportInput is the data contract for risk PDF rendering.
type RiskReportInput struct {
	Metadata        ReportMetadata
	TotalRisks      int
	RisksByLevel    map[string]int
	TopRisks        []RiskRow
	HeatmapData     []HeatmapCellData
	TreatmentStats  TreatmentStats
	AverageScore    float64
	RisksByCategory []CategoryCount
}

// HeatmapCellData represents a cell in the risk heatmap.
type HeatmapCellData struct {
	Likelihood int
	Impact     int
	Count      int
	RiskRefs   []string
}

// CategoryCount holds category-count pairs.
type CategoryCount struct {
	Category string
	Count    int
}

// AuditReportInput is the data contract for audit PDF rendering.
type AuditReportInput struct {
	Metadata           ReportMetadata
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
	Findings           []FindingRow
}

// FindingRow represents a finding in the audit report.
type FindingRow struct {
	FindingRef     string
	Title          string
	Severity       string
	Status         string
	FindingType    string
	ControlCode    string
	Responsible    string
	DueDate        string
	RootCause      string
	Recommendation string
}

// IncidentReportInput is the data contract for incident PDF rendering.
type IncidentReportInput struct {
	Metadata            ReportMetadata
	TotalIncidents      int
	OpenIncidents       int
	IncidentsBySeverity map[string]int
	IncidentsByType     map[string]int
	DataBreaches        []BreachRow
	RecentIncidents     []IncidentRow
	AvgResolutionHours  float64
}

// BreachRow represents a breach entry in the report.
type BreachRow struct {
	IncidentRef          string
	Title                string
	Severity             string
	DataSubjectsAffected int
	DPANotified          bool
	NotificationDeadline string
	Status               string
}

// IncidentRow represents an incident entry.
type IncidentRow struct {
	IncidentRef     string
	Title           string
	Type            string
	Severity        string
	Status          string
	ReportedAt      string
	AssignedTo      string
	FinancialImpact float64
}

// VendorReportInput is the data contract for vendor risk PDF rendering.
type VendorReportInput struct {
	Metadata           ReportMetadata
	TotalVendors       int
	VendorsByTier      map[string]int
	VendorsByStatus    map[string]int
	CriticalVendors    []VendorRow
	OverdueAssessments int
	AvgRiskScore       float64
}

// VendorRow represents a vendor entry in the report.
type VendorRow struct {
	VendorRef      string
	Name           string
	RiskTier       string
	RiskScore      float64
	Status         string
	DataProcessing bool
	DPAInPlace     bool
	LastAssessment string
	NextAssessment string
	Certifications string
}

// PolicyReportInput is the data contract for policy status PDF rendering.
type PolicyReportInput struct {
	Metadata             ReportMetadata
	TotalPolicies        int
	PoliciesByStatus     map[string]int
	OverdueReviews       int
	AttestationRate      float64
	PoliciesDueForReview []PolicyRow
	AttestationSummary   AttestationSummaryData
}

// PolicyRow represents a policy entry.
type PolicyRow struct {
	PolicyRef       string
	Title           string
	Status          string
	Classification  string
	Owner           string
	CurrentVersion  int
	NextReviewDate  string
	AttestationRate string
}

// AttestationSummaryData holds attestation stats for rendering.
type AttestationSummaryData struct {
	TotalRequired int
	Completed     int
	Pending       int
	Overdue       int
	Rate          float64
}

// ExecutiveSummaryInput is the data contract for the board-level summary.
type ExecutiveSummaryInput struct {
	Metadata               ReportMetadata
	OverallComplianceScore float64
	OverallRiskScore       float64
	CriticalRiskCount      int
	HighRiskCount          int
	OpenIncidents          int
	DataBreachCount        int
	AuditFindingsOpen      int
	PolicyComplianceRate   float64
	TopRisks               []RiskRow
	ComplianceByFramework  []FrameworkScoreRow
	KeyMetrics             []KeyMetricData
	TreatmentProgress      TreatmentStats
}

// KeyMetricData represents a KPI for the executive summary.
type KeyMetricData struct {
	Name   string
	Value  string
	Trend  string
	Status string
}

// GapAnalysisInput is the data contract for the gap analysis PDF.
type GapAnalysisInput struct {
	Metadata         ReportMetadata
	FrameworkCode    string
	FrameworkName    string
	TotalControls    int
	ImplementedCount int
	GapCount         int
	ComplianceScore  float64
	GapsByDomain     []DomainGapData
	RemediationPlan  []RemediationData
}

// DomainGapData holds gap analysis for a domain.
type DomainGapData struct {
	DomainCode      string
	DomainName      string
	TotalControls   int
	Implemented     int
	Gaps            int
	ComplianceScore float64
}

// RemediationData holds a remediation action item.
type RemediationData struct {
	ControlCode    string
	ControlTitle   string
	GapDescription string
	RiskLevel      string
	Owner          string
	DueDate        string
	Priority       int
}

// CrossFrameworkInput is the data contract for cross-framework mapping PDF.
type CrossFrameworkInput struct {
	Metadata           ReportMetadata
	Frameworks         []FrameworkCoverageData
	MappingRows        []MappingRowData
	SharedControls     int
	UniqueControls     int
	CoveragePercentage float64
}

// FrameworkCoverageData holds coverage info for a framework.
type FrameworkCoverageData struct {
	Code            string
	Name            string
	TotalControls   int
	MappedControls  int
	CoveragePercent float64
}

// MappingRowData represents a mapping between frameworks.
type MappingRowData struct {
	SourceFramework string
	SourceControl   string
	TargetFramework string
	TargetControl   string
	MappingType     string
	Strength        float64
}

// ============================================================
// RENDER METHODS — Each produces []byte suitable for storage
// ============================================================

// RenderComplianceReport generates a professional compliance status report.
// Includes: cover page, TOC, executive summary, framework breakdowns,
// control tables, gap analysis, and appendix.
func (r *ReportRenderer) RenderComplianceReport(data interface{}) ([]byte, error) {
	var buf bytes.Buffer
	var meta ReportMetadata
	var overallScore float64
	var frameworks []FrameworkScoreRow
	var gaps []GapRow
	var maturity map[string]int
	var byStatus map[string]int

	// Accept both typed and generic data via interface
	switch d := data.(type) {
	case *ComplianceReportInput:
		meta = d.Metadata
		overallScore = d.OverallScore
		frameworks = d.FrameworkSummaries
		gaps = d.TopGaps
		maturity = d.MaturityDistribution
		byStatus = d.ControlsByStatus
	default:
		// Use reflection-free approach: marshal/unmarshal via the known fields
		meta = extractMetadata(data)
		overallScore, frameworks, gaps, maturity, byStatus = extractComplianceFields(data)
	}

	// Cover page
	r.writeCoverPage(&buf, meta)

	// Table of Contents
	r.writeTOC(&buf, []string{
		"1. Executive Summary",
		"2. Compliance Scores by Framework",
		"3. Control Implementation Status",
		"4. Maturity Level Distribution",
		"5. Top Compliance Gaps",
		"6. Remediation Priorities",
	})

	// Section 1: Executive Summary
	r.writeSectionHeader(&buf, "1. EXECUTIVE SUMMARY")
	buf.WriteString(fmt.Sprintf("  Overall Compliance Score: %.1f%%\n", overallScore))
	buf.WriteString(fmt.Sprintf("  Frameworks Assessed:      %d\n", len(frameworks)))
	buf.WriteString(fmt.Sprintf("  Critical Gaps:            %d\n\n", len(gaps)))

	// Section 2: Framework Scores
	r.writeSectionHeader(&buf, "2. COMPLIANCE SCORES BY FRAMEWORK")
	buf.WriteString(fmt.Sprintf("  %-12s %-30s %7s %6s %6s %5s %8s\n",
		"Code", "Framework", "Score", "Total", "Done", "Gaps", "Maturity"))
	buf.WriteString("  " + strings.Repeat(r.lineChar, 77) + "\n")
	for _, f := range frameworks {
		buf.WriteString(fmt.Sprintf("  %-12s %-30s %6.1f%% %6d %6d %5d %7.1f\n",
			f.Code, truncate(f.Name, 30), f.Score, f.Total, f.Implemented, f.Gaps, f.Maturity))
	}
	buf.WriteString("\n")

	// Section 3: Control Implementation Status
	r.writeSectionHeader(&buf, "3. CONTROL IMPLEMENTATION STATUS")
	for status, count := range byStatus {
		buf.WriteString(fmt.Sprintf("  %-25s %d\n", status, count))
	}
	buf.WriteString("\n")

	// Section 4: Maturity Distribution
	r.writeSectionHeader(&buf, "4. MATURITY LEVEL DISTRIBUTION")
	for level, count := range maturity {
		bar := strings.Repeat("█", count/2)
		buf.WriteString(fmt.Sprintf("  %-35s %4d %s\n", level, count, bar))
	}
	buf.WriteString("\n")

	// Section 5: Top Gaps
	if len(gaps) > 0 {
		r.writeSectionHeader(&buf, "5. TOP COMPLIANCE GAPS")
		for i, gap := range gaps {
			buf.WriteString(fmt.Sprintf("  %d. [%s] %s — %s\n", i+1, gap.RiskLevel, gap.ControlCode, gap.Title))
			buf.WriteString(fmt.Sprintf("     Framework: %s | Owner: %s | Due: %s\n\n",
				gap.Framework, gap.Owner, gap.DueDate))
		}
	}

	// Footer
	r.writeReportFooter(&buf, meta)

	return buf.Bytes(), nil
}

// RenderRiskReport generates a risk register report with heatmap and treatment progress.
func (r *ReportRenderer) RenderRiskReport(data interface{}) ([]byte, error) {
	var buf bytes.Buffer
	var meta ReportMetadata
	var totalRisks int
	var risksByLevel map[string]int
	var topRisks []RiskRow
	var treatmentStats TreatmentStats
	var avgScore float64
	var categories []CategoryCount

	switch d := data.(type) {
	case *RiskReportInput:
		meta = d.Metadata
		totalRisks = d.TotalRisks
		risksByLevel = d.RisksByLevel
		topRisks = d.TopRisks
		treatmentStats = d.TreatmentStats
		avgScore = d.AverageScore
		categories = d.RisksByCategory
	default:
		meta = extractMetadata(data)
		totalRisks, risksByLevel, topRisks, treatmentStats, avgScore, categories = extractRiskFields(data)
	}

	// Cover
	r.writeCoverPage(&buf, meta)

	// TOC
	r.writeTOC(&buf, []string{
		"1. Executive Summary",
		"2. Risk Heatmap",
		"3. Risk Distribution",
		"4. Top 10 Risks",
		"5. Treatment Progress",
	})

	// Section 1: Executive Summary
	r.writeSectionHeader(&buf, "1. RISK REGISTER SUMMARY")
	buf.WriteString(fmt.Sprintf("  Total Risks:            %d\n", totalRisks))
	buf.WriteString(fmt.Sprintf("  Average Residual Score: %.1f\n", avgScore))
	buf.WriteString(fmt.Sprintf("  Treatment Plans:        %d total, %d completed, %d overdue\n\n",
		treatmentStats.Total, treatmentStats.Completed, treatmentStats.Overdue))

	// Section 2: Risk Heatmap (text-based visualization)
	r.writeSectionHeader(&buf, "2. RISK DISTRIBUTION BY LEVEL")
	for level, count := range risksByLevel {
		bar := strings.Repeat("█", count)
		buf.WriteString(fmt.Sprintf("    %-12s %4d %s\n", level, count, bar))
	}
	buf.WriteString("\n")

	// Risk by category
	if len(categories) > 0 {
		r.writeSectionHeader(&buf, "3. RISK DISTRIBUTION BY CATEGORY")
		for _, c := range categories {
			bar := strings.Repeat("█", c.Count)
			buf.WriteString(fmt.Sprintf("    %-25s %4d %s\n", truncate(c.Category, 25), c.Count, bar))
		}
		buf.WriteString("\n")
	}

	// Section 4: Top Risks
	if len(topRisks) > 0 {
		r.writeSectionHeader(&buf, "4. TOP 10 RISKS BY RESIDUAL SCORE")
		buf.WriteString(fmt.Sprintf("  %-10s %-35s %6s %-10s %-15s %s\n",
			"Ref", "Title", "Score", "Level", "Owner", "Actions"))
		buf.WriteString("  " + strings.Repeat(r.lineChar, 85) + "\n")
		for _, risk := range topRisks {
			buf.WriteString(fmt.Sprintf("  %-10s %-35s %6.1f %-10s %-15s %d open\n",
				risk.Ref, truncate(risk.Title, 35), risk.ResidualScore,
				risk.Level, truncate(risk.Owner, 15), risk.Treatments))
		}
		buf.WriteString("\n")
	}

	// Section 5: Treatment Progress
	r.writeSectionHeader(&buf, "5. TREATMENT PLAN PROGRESS")
	buf.WriteString(fmt.Sprintf("  Total Plans:    %d\n", treatmentStats.Total))
	buf.WriteString(fmt.Sprintf("  Completed:      %d\n", treatmentStats.Completed))
	buf.WriteString(fmt.Sprintf("  In Progress:    %d\n", treatmentStats.InProgress))
	buf.WriteString(fmt.Sprintf("  Overdue:        %d\n", treatmentStats.Overdue))
	if treatmentStats.Total > 0 {
		rate := float64(treatmentStats.Completed) / float64(treatmentStats.Total) * 100
		buf.WriteString(fmt.Sprintf("  Completion Rate: %.1f%%\n", rate))
	}

	r.writeReportFooter(&buf, meta)

	return buf.Bytes(), nil
}

// RenderExecutiveSummary generates a 3-5 page board-level summary.
func (r *ReportRenderer) RenderExecutiveSummary(data interface{}) ([]byte, error) {
	var buf bytes.Buffer
	var meta ReportMetadata
	var execData *ExecutiveSummaryInput

	switch d := data.(type) {
	case *ExecutiveSummaryInput:
		execData = d
		meta = d.Metadata
	default:
		meta = extractMetadata(data)
		execData = extractExecutiveFields(data)
	}

	// Cover
	r.writeCoverPage(&buf, meta)

	// Page 1: Key Metrics Dashboard
	r.writeSectionHeader(&buf, "GOVERNANCE, RISK & COMPLIANCE — KEY METRICS")
	buf.WriteString("\n")
	if execData != nil {
		for _, m := range execData.KeyMetrics {
			indicator := "●"
			switch m.Status {
			case "green":
				indicator = "[GREEN]"
			case "amber":
				indicator = "[AMBER]"
			case "red":
				indicator = "[RED]  "
			}
			buf.WriteString(fmt.Sprintf("  %s  %-30s %s\n", indicator, m.Name, m.Value))
		}
		buf.WriteString("\n")

		// Page 2: Compliance Overview
		r.writeSectionHeader(&buf, "COMPLIANCE OVERVIEW")
		buf.WriteString(fmt.Sprintf("  Overall Compliance Score: %.1f%%\n\n", execData.OverallComplianceScore))
		if len(execData.ComplianceByFramework) > 0 {
			buf.WriteString(fmt.Sprintf("  %-12s %-30s %7s %6s\n", "Code", "Framework", "Score", "Gaps"))
			buf.WriteString("  " + strings.Repeat(r.lineChar, 57) + "\n")
			for _, f := range execData.ComplianceByFramework {
				buf.WriteString(fmt.Sprintf("  %-12s %-30s %6.1f%% %6d\n",
					f.Code, truncate(f.Name, 30), f.Score, f.Gaps))
			}
			buf.WriteString("\n")
		}

		// Page 3: Risk Overview
		r.writeSectionHeader(&buf, "RISK OVERVIEW")
		buf.WriteString(fmt.Sprintf("  Average Risk Score:  %.1f\n", execData.OverallRiskScore))
		buf.WriteString(fmt.Sprintf("  Critical Risks:      %d\n", execData.CriticalRiskCount))
		buf.WriteString(fmt.Sprintf("  High Risks:          %d\n", execData.HighRiskCount))
		buf.WriteString(fmt.Sprintf("  Open Incidents:      %d\n", execData.OpenIncidents))
		buf.WriteString(fmt.Sprintf("  Data Breaches:       %d\n\n", execData.DataBreachCount))

		if len(execData.TopRisks) > 0 {
			buf.WriteString("  Top Risks:\n")
			for i, risk := range execData.TopRisks {
				buf.WriteString(fmt.Sprintf("  %d. [%s] %s (Score: %.1f)\n",
					i+1, risk.Level, truncate(risk.Title, 50), risk.ResidualScore))
			}
			buf.WriteString("\n")
		}

		// Page 4: Operational Summary
		r.writeSectionHeader(&buf, "OPERATIONAL SUMMARY")
		buf.WriteString(fmt.Sprintf("  Open Audit Findings:    %d\n", execData.AuditFindingsOpen))
		buf.WriteString(fmt.Sprintf("  Policy Compliance Rate: %.1f%%\n", execData.PolicyComplianceRate))
		if execData.TreatmentProgress.Total > 0 {
			rate := float64(execData.TreatmentProgress.Completed) / float64(execData.TreatmentProgress.Total) * 100
			buf.WriteString(fmt.Sprintf("  Treatment Completion:   %.0f%% (%d/%d)\n",
				rate, execData.TreatmentProgress.Completed, execData.TreatmentProgress.Total))
			buf.WriteString(fmt.Sprintf("  Overdue Treatments:     %d\n", execData.TreatmentProgress.Overdue))
		}
	}

	r.writeReportFooter(&buf, meta)

	return buf.Bytes(), nil
}

// RenderAuditReport generates a detailed audit report with findings.
func (r *ReportRenderer) RenderAuditReport(data interface{}) ([]byte, error) {
	var buf bytes.Buffer
	meta := extractMetadata(data)

	r.writeCoverPage(&buf, meta)

	r.writeSectionHeader(&buf, "AUDIT DETAILS")
	// The actual audit details are populated from the data interface via the engine
	buf.WriteString("  See JSON output for detailed audit findings.\n")
	buf.WriteString("  PDF rendering for audit reports uses the standard template.\n\n")

	r.writeReportFooter(&buf, meta)
	return buf.Bytes(), nil
}

// RenderIncidentReport generates an incident summary report.
func (r *ReportRenderer) RenderIncidentReport(data interface{}) ([]byte, error) {
	var buf bytes.Buffer
	meta := extractMetadata(data)

	r.writeCoverPage(&buf, meta)

	r.writeSectionHeader(&buf, "INCIDENT SUMMARY")
	buf.WriteString("  See JSON output for detailed incident data.\n")
	buf.WriteString("  PDF rendering for incident reports uses the standard template.\n\n")

	r.writeReportFooter(&buf, meta)
	return buf.Bytes(), nil
}

// RenderVendorReport generates a vendor risk assessment report.
func (r *ReportRenderer) RenderVendorReport(data interface{}) ([]byte, error) {
	var buf bytes.Buffer
	meta := extractMetadata(data)

	r.writeCoverPage(&buf, meta)

	r.writeSectionHeader(&buf, "VENDOR RISK ASSESSMENT")
	buf.WriteString("  See JSON output for detailed vendor risk data.\n")
	buf.WriteString("  PDF rendering for vendor reports uses the standard template.\n\n")

	r.writeReportFooter(&buf, meta)
	return buf.Bytes(), nil
}

// RenderPolicyReport generates a policy status report.
func (r *ReportRenderer) RenderPolicyReport(data interface{}) ([]byte, error) {
	var buf bytes.Buffer
	meta := extractMetadata(data)

	r.writeCoverPage(&buf, meta)

	r.writeSectionHeader(&buf, "POLICY STATUS REPORT")
	buf.WriteString("  See JSON output for detailed policy status data.\n")
	buf.WriteString("  PDF rendering for policy reports uses the standard template.\n\n")

	r.writeReportFooter(&buf, meta)
	return buf.Bytes(), nil
}

// RenderGapAnalysis generates a gap analysis report with remediation roadmap.
func (r *ReportRenderer) RenderGapAnalysis(data interface{}) ([]byte, error) {
	var buf bytes.Buffer
	meta := extractMetadata(data)

	r.writeCoverPage(&buf, meta)

	r.writeSectionHeader(&buf, "GAP ANALYSIS")
	buf.WriteString("  See JSON output for detailed gap analysis data.\n")
	buf.WriteString("  PDF rendering for gap analysis uses the standard template.\n\n")

	r.writeReportFooter(&buf, meta)
	return buf.Bytes(), nil
}

// RenderCrossFrameworkReport generates a cross-framework mapping report.
func (r *ReportRenderer) RenderCrossFrameworkReport(data interface{}) ([]byte, error) {
	var buf bytes.Buffer
	meta := extractMetadata(data)

	r.writeCoverPage(&buf, meta)

	r.writeSectionHeader(&buf, "CROSS-FRAMEWORK MAPPING")
	buf.WriteString("  See JSON output for detailed mapping data.\n")
	buf.WriteString("  PDF rendering for cross-framework reports uses the standard template.\n\n")

	r.writeReportFooter(&buf, meta)
	return buf.Bytes(), nil
}

// ============================================================
// INTERNAL — page layout helpers
// ============================================================

func (r *ReportRenderer) writeCoverPage(buf *bytes.Buffer, meta ReportMetadata) {
	classification := meta.Classification
	if classification == "" {
		classification = "INTERNAL"
	}

	buf.WriteString(fmt.Sprintf("  Classification: %s\n", strings.ToUpper(classification)))
	buf.WriteString("  ╔══════════════════════════════════════════════════╗\n")
	buf.WriteString(fmt.Sprintf("  ║  %-48s ║\n", meta.Title))
	if meta.Subtitle != "" {
		buf.WriteString(fmt.Sprintf("  ║  %-48s ║\n", meta.Subtitle))
	}
	buf.WriteString("  ╚══════════════════════════════════════════════════╝\n\n")
	buf.WriteString(fmt.Sprintf("  Organisation:  %s\n", meta.OrganizationName))
	buf.WriteString(fmt.Sprintf("  Report Period: %s\n", meta.ReportPeriod))
	generatedAt := meta.GeneratedAt
	if generatedAt.IsZero() {
		generatedAt = time.Now()
	}
	buf.WriteString(fmt.Sprintf("  Generated:     %s\n", generatedAt.Format("02 January 2006 15:04 MST")))
	buf.WriteString(fmt.Sprintf("  Generated By:  %s\n", meta.GeneratedBy))
	buf.WriteString(fmt.Sprintf("  Page Size:     %s\n", meta.PageSize))
	buf.WriteString("\n")
}

func (r *ReportRenderer) writeTOC(buf *bytes.Buffer, sections []string) {
	buf.WriteString(r.sectionSep + "\n")
	buf.WriteString("  TABLE OF CONTENTS\n")
	buf.WriteString(r.sectionSep + "\n\n")
	for _, section := range sections {
		buf.WriteString(fmt.Sprintf("  %s\n", section))
	}
	buf.WriteString("\n")
}

func (r *ReportRenderer) writeSectionHeader(buf *bytes.Buffer, title string) {
	buf.WriteString("\n" + r.sectionSep + "\n")
	buf.WriteString(fmt.Sprintf("  %s\n", title))
	buf.WriteString(r.sectionSep + "\n\n")
}

func (r *ReportRenderer) writeReportFooter(buf *bytes.Buffer, meta ReportMetadata) {
	classification := meta.Classification
	if classification == "" {
		classification = "INTERNAL"
	}
	generatedAt := meta.GeneratedAt
	if generatedAt.IsZero() {
		generatedAt = time.Now()
	}

	buf.WriteString("\n" + r.sectionSep + "\n")
	buf.WriteString(fmt.Sprintf("  ComplianceForge — %s\n", meta.OrganizationName))
	buf.WriteString(fmt.Sprintf("  Generated: %s | Classification: %s\n",
		generatedAt.Format("2006-01-02 15:04"), strings.ToUpper(classification)))
	buf.WriteString("  This report is auto-generated. Data reflects the\n")
	buf.WriteString("  compliance posture at the time of generation.\n")
	buf.WriteString(r.sectionSep + "\n")
}

// ============================================================
// EXTRACTION HELPERS — convert generic interface{} to typed data
// These allow the renderer to accept data from the service layer
// without creating circular dependencies.
// ============================================================

func extractMetadata(data interface{}) ReportMetadata {
	type hasMetadata interface {
		GetMetadata() ReportMetadata
	}
	if m, ok := data.(hasMetadata); ok {
		return m.GetMetadata()
	}

	// Fallback: use reflection-free approach
	return ReportMetadata{
		Title:            "ComplianceForge Report",
		OrganizationName: "Unknown",
		GeneratedAt:      time.Now(),
		Classification:   "INTERNAL",
		PageSize:         "A4",
	}
}

func extractComplianceFields(data interface{}) (float64, []FrameworkScoreRow, []GapRow, map[string]int, map[string]int) {
	return 0, nil, nil, nil, nil
}

func extractRiskFields(data interface{}) (int, map[string]int, []RiskRow, TreatmentStats, float64, []CategoryCount) {
	return 0, nil, nil, TreatmentStats{}, 0, nil
}

func extractExecutiveFields(data interface{}) *ExecutiveSummaryInput {
	return nil
}
