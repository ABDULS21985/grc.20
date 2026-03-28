// Package pdf generates PDF compliance reports for ComplianceForge.
// Supports compliance status reports, risk register exports,
// audit finding summaries, and executive dashboards.
package pdf

import (
	"bytes"
	"fmt"
	"time"
)

// ReportMetadata contains common metadata for all PDF reports.
type ReportMetadata struct {
	Title            string
	Subtitle         string
	OrganizationName string
	GeneratedBy      string
	GeneratedAt      time.Time
	Classification   string // PUBLIC, INTERNAL, CONFIDENTIAL, RESTRICTED
	ReportPeriod     string
	PageSize         string // A4, Letter
}

// ComplianceReportData holds data for the compliance status PDF report.
type ComplianceReportData struct {
	Metadata         ReportMetadata
	OverallScore     float64
	FrameworkScores  []FrameworkScoreRow
	TopGaps          []GapRow
	MaturityBreakdown map[string]int
	ControlsByStatus map[string]int
}

type FrameworkScoreRow struct {
	Code        string
	Name        string
	Score       float64
	Total       int
	Implemented int
	Gaps        int
	Maturity    float64
}

type GapRow struct {
	Framework   string
	ControlCode string
	Title       string
	RiskLevel   string
	Owner       string
	DueDate     string
}

// RiskReportData holds data for the risk register PDF report.
type RiskReportData struct {
	Metadata        ReportMetadata
	TotalRisks      int
	RisksByLevel    map[string]int
	TopRisks        []RiskRow
	TreatmentStats  TreatmentStats
	AverageScore    float64
}

type RiskRow struct {
	Ref           string
	Title         string
	Category      string
	ResidualScore float64
	Level         string
	Owner         string
	Treatments    int
}

type TreatmentStats struct {
	Total      int
	Completed  int
	InProgress int
	Overdue    int
}

// Generator generates PDF reports.
type Generator struct {
	// In production, use a library like gofpdf, maroto, or wkhtmltopdf
	// For now, we generate a structured text report that can be piped to a PDF renderer
}

// NewGenerator creates a new PDF generator.
func NewGenerator() *Generator {
	return &Generator{}
}

// GenerateComplianceReport creates a compliance status PDF report.
// Returns the PDF content as bytes.
func (g *Generator) GenerateComplianceReport(data ComplianceReportData) ([]byte, error) {
	var buf bytes.Buffer

	// Header
	g.writeHeader(&buf, data.Metadata)

	// Executive Summary
	buf.WriteString("\n══════════════════════════════════════════════════════\n")
	buf.WriteString("  EXECUTIVE SUMMARY\n")
	buf.WriteString("══════════════════════════════════════════════════════\n\n")
	buf.WriteString(fmt.Sprintf("  Overall Compliance Score: %.1f%%\n", data.OverallScore))
	buf.WriteString(fmt.Sprintf("  Frameworks Assessed: %d\n", len(data.FrameworkScores)))
	buf.WriteString(fmt.Sprintf("  Critical Gaps: %d\n\n", len(data.TopGaps)))

	// Framework Scores Table
	buf.WriteString("══════════════════════════════════════════════════════\n")
	buf.WriteString("  COMPLIANCE SCORES BY FRAMEWORK\n")
	buf.WriteString("══════════════════════════════════════════════════════\n\n")
	buf.WriteString(fmt.Sprintf("  %-12s %-30s %7s %6s %6s %5s %8s\n",
		"Code", "Framework", "Score", "Total", "Done", "Gaps", "Maturity"))
	buf.WriteString("  " + repeatStr("-", 85) + "\n")

	for _, f := range data.FrameworkScores {
		buf.WriteString(fmt.Sprintf("  %-12s %-30s %6.1f%% %6d %6d %5d %7.1f\n",
			f.Code, truncate(f.Name, 30), f.Score, f.Total, f.Implemented, f.Gaps, f.Maturity))
	}

	// Controls by Status
	buf.WriteString("\n══════════════════════════════════════════════════════\n")
	buf.WriteString("  CONTROL IMPLEMENTATION STATUS\n")
	buf.WriteString("══════════════════════════════════════════════════════\n\n")
	for status, count := range data.ControlsByStatus {
		buf.WriteString(fmt.Sprintf("  %-25s %d\n", status, count))
	}

	// Maturity Distribution
	buf.WriteString("\n══════════════════════════════════════════════════════\n")
	buf.WriteString("  MATURITY LEVEL DISTRIBUTION\n")
	buf.WriteString("══════════════════════════════════════════════════════\n\n")
	for level, count := range data.MaturityBreakdown {
		bar := repeatStr("█", count/2)
		buf.WriteString(fmt.Sprintf("  %-35s %4d %s\n", level, count, bar))
	}

	// Top Gaps
	if len(data.TopGaps) > 0 {
		buf.WriteString("\n══════════════════════════════════════════════════════\n")
		buf.WriteString("  TOP COMPLIANCE GAPS (Priority Remediation)\n")
		buf.WriteString("══════════════════════════════════════════════════════\n\n")
		for i, gap := range data.TopGaps {
			buf.WriteString(fmt.Sprintf("  %d. [%s] %s — %s\n", i+1, gap.RiskLevel, gap.ControlCode, gap.Title))
			buf.WriteString(fmt.Sprintf("     Framework: %s | Owner: %s | Due: %s\n\n", gap.Framework, gap.Owner, gap.DueDate))
		}
	}

	g.writeFooter(&buf, data.Metadata)

	return buf.Bytes(), nil
}

// GenerateRiskReport creates a risk register PDF report.
func (g *Generator) GenerateRiskReport(data RiskReportData) ([]byte, error) {
	var buf bytes.Buffer

	g.writeHeader(&buf, data.Metadata)

	// Executive Summary
	buf.WriteString("\n══════════════════════════════════════════════════════\n")
	buf.WriteString("  RISK REGISTER SUMMARY\n")
	buf.WriteString("══════════════════════════════════════════════════════\n\n")
	buf.WriteString(fmt.Sprintf("  Total Risks: %d\n", data.TotalRisks))
	buf.WriteString(fmt.Sprintf("  Average Residual Score: %.1f\n", data.AverageScore))
	buf.WriteString(fmt.Sprintf("  Treatment Plans: %d total, %d completed, %d overdue\n\n",
		data.TreatmentStats.Total, data.TreatmentStats.Completed, data.TreatmentStats.Overdue))

	// Risk Distribution
	buf.WriteString("  Risk Distribution by Level:\n")
	for level, count := range data.RisksByLevel {
		bar := repeatStr("█", count)
		buf.WriteString(fmt.Sprintf("    %-12s %4d %s\n", level, count, bar))
	}

	// Top 10 Risks
	if len(data.TopRisks) > 0 {
		buf.WriteString("\n══════════════════════════════════════════════════════\n")
		buf.WriteString("  TOP 10 RISKS BY RESIDUAL SCORE\n")
		buf.WriteString("══════════════════════════════════════════════════════\n\n")
		buf.WriteString(fmt.Sprintf("  %-10s %-35s %6s %-10s %-15s %s\n",
			"Ref", "Title", "Score", "Level", "Owner", "Actions"))
		buf.WriteString("  " + repeatStr("-", 95) + "\n")

		for _, r := range data.TopRisks {
			buf.WriteString(fmt.Sprintf("  %-10s %-35s %6.1f %-10s %-15s %d open\n",
				r.Ref, truncate(r.Title, 35), r.ResidualScore, r.Level, truncate(r.Owner, 15), r.Treatments))
		}
	}

	// Treatment Progress
	buf.WriteString("\n══════════════════════════════════════════════════════\n")
	buf.WriteString("  TREATMENT PLAN PROGRESS\n")
	buf.WriteString("══════════════════════════════════════════════════════\n\n")
	buf.WriteString(fmt.Sprintf("  Total Plans:    %d\n", data.TreatmentStats.Total))
	buf.WriteString(fmt.Sprintf("  Completed:      %d\n", data.TreatmentStats.Completed))
	buf.WriteString(fmt.Sprintf("  In Progress:    %d\n", data.TreatmentStats.InProgress))
	buf.WriteString(fmt.Sprintf("  Overdue:        %d\n", data.TreatmentStats.Overdue))

	if data.TreatmentStats.Total > 0 {
		completionRate := float64(data.TreatmentStats.Completed) / float64(data.TreatmentStats.Total) * 100
		buf.WriteString(fmt.Sprintf("  Completion Rate: %.1f%%\n", completionRate))
	}

	g.writeFooter(&buf, data.Metadata)

	return buf.Bytes(), nil
}

func (g *Generator) writeHeader(buf *bytes.Buffer, meta ReportMetadata) {
	classification := meta.Classification
	if classification == "" {
		classification = "INTERNAL"
	}

	buf.WriteString(fmt.Sprintf("  Classification: %s\n", classification))
	buf.WriteString("  ╔══════════════════════════════════════════════════╗\n")
	buf.WriteString(fmt.Sprintf("  ║  %-48s ║\n", meta.Title))
	if meta.Subtitle != "" {
		buf.WriteString(fmt.Sprintf("  ║  %-48s ║\n", meta.Subtitle))
	}
	buf.WriteString("  ╚══════════════════════════════════════════════════╝\n\n")
	buf.WriteString(fmt.Sprintf("  Organisation: %s\n", meta.OrganizationName))
	buf.WriteString(fmt.Sprintf("  Report Period: %s\n", meta.ReportPeriod))
	buf.WriteString(fmt.Sprintf("  Generated: %s\n", meta.GeneratedAt.Format("02 January 2006 15:04 MST")))
	buf.WriteString(fmt.Sprintf("  Generated By: %s\n", meta.GeneratedBy))
}

func (g *Generator) writeFooter(buf *bytes.Buffer, meta ReportMetadata) {
	buf.WriteString("\n══════════════════════════════════════════════════════\n")
	buf.WriteString(fmt.Sprintf("  ComplianceForge — %s\n", meta.OrganizationName))
	buf.WriteString(fmt.Sprintf("  Generated: %s | Classification: %s\n",
		meta.GeneratedAt.Format("2006-01-02 15:04"), meta.Classification))
	buf.WriteString("  This report is auto-generated. Data reflects the\n")
	buf.WriteString("  compliance posture at the time of generation.\n")
	buf.WriteString("══════════════════════════════════════════════════════\n")
}

func repeatStr(s string, n int) string {
	if n <= 0 {
		return ""
	}
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
