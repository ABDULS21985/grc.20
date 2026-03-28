package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ReportingService generates compliance, risk, audit, and executive reports.
type ReportingService struct {
	pool *pgxpool.Pool
}

func NewReportingService(pool *pgxpool.Pool) *ReportingService {
	return &ReportingService{pool: pool}
}

// ComplianceReport contains comprehensive compliance status data.
type ComplianceReport struct {
	GeneratedAt        time.Time                `json:"generated_at"`
	OrganizationID     uuid.UUID                `json:"organization_id"`
	ReportPeriod       string                   `json:"report_period"`
	OverallScore       float64                  `json:"overall_compliance_score"`
	FrameworkSummaries []FrameworkReportSummary  `json:"framework_summaries"`
	TopGaps            []TopGapEntry            `json:"top_gaps"`
	MaturityDistribution map[string]int          `json:"maturity_distribution"`
	ControlsByStatus   map[string]int            `json:"controls_by_status"`
}

type FrameworkReportSummary struct {
	Code            string  `json:"code"`
	Name            string  `json:"name"`
	Score           float64 `json:"compliance_score"`
	TotalControls   int     `json:"total_controls"`
	Implemented     int     `json:"implemented"`
	Gaps            int     `json:"gaps"`
	MaturityAvg     float64 `json:"maturity_avg"`
}

type TopGapEntry struct {
	Framework       string `json:"framework"`
	ControlCode     string `json:"control_code"`
	ControlTitle    string `json:"control_title"`
	RiskLevel       string `json:"risk_if_not_implemented"`
	Owner           string `json:"owner"`
	RemediationDue  string `json:"remediation_due_date"`
}

func (s *ReportingService) GenerateComplianceReport(ctx context.Context, orgID uuid.UUID) (*ComplianceReport, error) {
	report := &ComplianceReport{
		GeneratedAt:    time.Now(),
		OrganizationID: orgID,
		ReportPeriod:   time.Now().Format("January 2006"),
	}

	// Framework summaries
	rows, err := s.pool.Query(ctx, `
		SELECT framework_code, framework_name, COALESCE(compliance_score, 0),
			total_controls, implemented,
			(total_controls - implemented - not_applicable) AS gaps,
			COALESCE(maturity_avg, 0)
		FROM v_compliance_score_by_framework WHERE organization_id = $1`, orgID)
	if err != nil {
		return nil, err
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
		report.FrameworkSummaries = append(report.FrameworkSummaries, f)
		totalScore += f.Score
		count++
	}
	if count > 0 {
		report.OverallScore = totalScore / float64(count)
	}

	// Top 10 critical gaps
	gapRows, err := s.pool.Query(ctx, `
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
		LIMIT 10`, orgID)
	if err == nil {
		defer gapRows.Close()
		for gapRows.Next() {
			var g TopGapEntry
			gapRows.Scan(&g.Framework, &g.ControlCode, &g.ControlTitle,
				&g.RiskLevel, &g.Owner, &g.RemediationDue)
			report.TopGaps = append(report.TopGaps, g)
		}
	}

	// Maturity distribution
	report.MaturityDistribution = make(map[string]int)
	matRows, err := s.pool.Query(ctx, `
		SELECT maturity_level, COUNT(*)
		FROM control_implementations
		WHERE organization_id = $1 AND status != 'not_applicable' AND deleted_at IS NULL
		GROUP BY maturity_level ORDER BY maturity_level`, orgID)
	if err == nil {
		defer matRows.Close()
		levels := []string{"Non-existent", "Initial", "Managed", "Defined", "Quantitatively Managed", "Optimizing"}
		for matRows.Next() {
			var level, count int
			matRows.Scan(&level, &count)
			if level >= 0 && level < len(levels) {
				report.MaturityDistribution[fmt.Sprintf("Level %d - %s", level, levels[level])] = count
			}
		}
	}

	// Controls by status
	report.ControlsByStatus = make(map[string]int)
	statusRows, err := s.pool.Query(ctx, `
		SELECT status::TEXT, COUNT(*)
		FROM control_implementations
		WHERE organization_id = $1 AND deleted_at IS NULL
		GROUP BY status`, orgID)
	if err == nil {
		defer statusRows.Close()
		for statusRows.Next() {
			var status string
			var count int
			statusRows.Scan(&status, &count)
			report.ControlsByStatus[status] = count
		}
	}

	return report, nil
}

// RiskReport contains risk register summary data.
type RiskReport struct {
	GeneratedAt        time.Time            `json:"generated_at"`
	TotalRisks         int                  `json:"total_risks"`
	RisksByLevel       map[string]int       `json:"risks_by_level"`
	RisksByCategory    []CategoryRiskCount  `json:"risks_by_category"`
	TopRisks           []TopRiskEntry       `json:"top_10_risks"`
	TreatmentProgress  TreatmentSummary     `json:"treatment_progress"`
	AverageRiskScore   float64              `json:"average_residual_score"`
}

type CategoryRiskCount struct {
	Category string `json:"category"`
	Count    int    `json:"count"`
}

type TopRiskEntry struct {
	Ref            string  `json:"risk_ref"`
	Title          string  `json:"title"`
	ResidualScore  float64 `json:"residual_score"`
	Level          string  `json:"level"`
	Owner          string  `json:"owner"`
	TreatmentCount int     `json:"open_treatments"`
}

type TreatmentSummary struct {
	Total     int `json:"total"`
	Completed int `json:"completed"`
	InProgress int `json:"in_progress"`
	Overdue   int `json:"overdue"`
}

func (s *ReportingService) GenerateRiskReport(ctx context.Context, orgID uuid.UUID) (*RiskReport, error) {
	report := &RiskReport{
		GeneratedAt:  time.Now(),
		RisksByLevel: make(map[string]int),
	}

	// Total and by level
	s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM risks WHERE organization_id = $1 AND deleted_at IS NULL`, orgID).Scan(&report.TotalRisks)

	levelRows, _ := s.pool.Query(ctx, `
		SELECT residual_risk_level::TEXT, COUNT(*)
		FROM risks WHERE organization_id = $1 AND deleted_at IS NULL
		GROUP BY residual_risk_level`, orgID)
	if levelRows != nil {
		defer levelRows.Close()
		for levelRows.Next() {
			var level string
			var count int
			levelRows.Scan(&level, &count)
			report.RisksByLevel[level] = count
		}
	}

	// By category
	catRows, _ := s.pool.Query(ctx, `
		SELECT rc.name, COUNT(*)
		FROM risks r JOIN risk_categories rc ON r.risk_category_id = rc.id
		WHERE r.organization_id = $1 AND r.deleted_at IS NULL
		GROUP BY rc.name ORDER BY COUNT(*) DESC`, orgID)
	if catRows != nil {
		defer catRows.Close()
		for catRows.Next() {
			var c CategoryRiskCount
			catRows.Scan(&c.Category, &c.Count)
			report.RisksByCategory = append(report.RisksByCategory, c)
		}
	}

	// Top 10
	topRows, _ := s.pool.Query(ctx, `
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
			topRows.Scan(&t.Ref, &t.Title, &t.ResidualScore, &t.Level, &t.Owner, &t.TreatmentCount)
			report.TopRisks = append(report.TopRisks, t)
		}
	}

	// Treatment progress
	s.pool.QueryRow(ctx, `
		SELECT COUNT(*),
			COUNT(*) FILTER (WHERE status = 'completed'),
			COUNT(*) FILTER (WHERE status = 'in_progress'),
			COUNT(*) FILTER (WHERE status != 'completed' AND target_date < CURRENT_DATE)
		FROM risk_treatments WHERE organization_id = $1`, orgID,
	).Scan(&report.TreatmentProgress.Total, &report.TreatmentProgress.Completed,
		&report.TreatmentProgress.InProgress, &report.TreatmentProgress.Overdue)

	// Average score
	s.pool.QueryRow(ctx, `
		SELECT COALESCE(AVG(residual_risk_score), 0)
		FROM risks WHERE organization_id = $1 AND deleted_at IS NULL`, orgID,
	).Scan(&report.AverageRiskScore)

	return report, nil
}
