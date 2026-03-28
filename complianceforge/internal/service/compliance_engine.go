package service

import (
	"context"
	"fmt"
	"math"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ComplianceEngine handles compliance scoring, gap analysis, and cross-framework mapping.
type ComplianceEngine struct {
	pool *pgxpool.Pool
}

func NewComplianceEngine(pool *pgxpool.Pool) *ComplianceEngine {
	return &ComplianceEngine{pool: pool}
}

// CalculateFrameworkScore computes the compliance score for an organization's adopted framework.
// Score = (Implemented + Effective) / (Total - Not Applicable) * 100
func (e *ComplianceEngine) CalculateFrameworkScore(ctx context.Context, orgID, orgFrameworkID uuid.UUID) (float64, error) {
	var total, implemented, notApplicable int64

	err := e.pool.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE status IN ('implemented', 'effective')),
			COUNT(*) FILTER (WHERE status = 'not_applicable')
		FROM control_implementations
		WHERE organization_id = $1 AND org_framework_id = $2 AND deleted_at IS NULL`,
		orgID, orgFrameworkID,
	).Scan(&total, &implemented, &notApplicable)

	if err != nil {
		return 0, fmt.Errorf("failed to calculate score: %w", err)
	}

	applicable := total - notApplicable
	if applicable == 0 {
		return 0, nil
	}

	score := float64(implemented) / float64(applicable) * 100
	score = math.Round(score*100) / 100 // round to 2 decimal places

	// Update the stored score
	_, err = e.pool.Exec(ctx, `
		UPDATE organization_frameworks SET compliance_score = $1, last_assessment_date = NOW()
		WHERE id = $2 AND organization_id = $3`,
		score, orgFrameworkID, orgID,
	)

	return score, err
}

// CalculateMaturityScore computes the average CMMI maturity level across a framework's controls.
func (e *ComplianceEngine) CalculateMaturityScore(ctx context.Context, orgID, orgFrameworkID uuid.UUID) (float64, error) {
	var avgMaturity float64
	err := e.pool.QueryRow(ctx, `
		SELECT COALESCE(AVG(maturity_level), 0)
		FROM control_implementations
		WHERE organization_id = $1 AND org_framework_id = $2
			AND status != 'not_applicable' AND deleted_at IS NULL`,
		orgID, orgFrameworkID,
	).Scan(&avgMaturity)

	return math.Round(avgMaturity*100) / 100, err
}

// GapAnalysis returns all controls that are not fully implemented or effective.
type GapEntry struct {
	FrameworkCode   string  `json:"framework_code"`
	ControlCode     string  `json:"control_code"`
	ControlTitle    string  `json:"control_title"`
	Status          string  `json:"status"`
	MaturityLevel   int     `json:"maturity_level"`
	GapDescription  string  `json:"gap_description"`
	RemediationPlan string  `json:"remediation_plan"`
	RemediationDue  *string `json:"remediation_due_date"`
	RiskLevel       string  `json:"risk_if_not_implemented"`
	OwnerName       string  `json:"owner_name"`
}

func (e *ComplianceEngine) GapAnalysis(ctx context.Context, orgID uuid.UUID, frameworkID *uuid.UUID) ([]GapEntry, error) {
	query := `
		SELECT
			cf.code AS framework_code,
			fc.code AS control_code,
			fc.title AS control_title,
			ci.status::TEXT,
			ci.maturity_level,
			COALESCE(ci.gap_description, ''),
			COALESCE(ci.remediation_plan, ''),
			ci.remediation_due_date::TEXT,
			ci.risk_if_not_implemented,
			COALESCE(u.first_name || ' ' || u.last_name, 'Unassigned')
		FROM control_implementations ci
		JOIN framework_controls fc ON ci.framework_control_id = fc.id
		JOIN organization_frameworks of2 ON ci.org_framework_id = of2.id
		JOIN compliance_frameworks cf ON of2.framework_id = cf.id
		LEFT JOIN users u ON ci.owner_user_id = u.id
		WHERE ci.organization_id = $1
			AND ci.status NOT IN ('implemented', 'effective', 'not_applicable')
			AND ci.deleted_at IS NULL`

	args := []interface{}{orgID}
	if frameworkID != nil {
		query += " AND of2.framework_id = $2"
		args = append(args, *frameworkID)
	}
	query += " ORDER BY ci.risk_if_not_implemented DESC, fc.sort_order ASC"

	rows, err := e.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var gaps []GapEntry
	for rows.Next() {
		var g GapEntry
		if err := rows.Scan(
			&g.FrameworkCode, &g.ControlCode, &g.ControlTitle, &g.Status,
			&g.MaturityLevel, &g.GapDescription, &g.RemediationPlan,
			&g.RemediationDue, &g.RiskLevel, &g.OwnerName,
		); err != nil {
			return nil, err
		}
		gaps = append(gaps, g)
	}
	return gaps, nil
}

// CrossFrameworkCoverage shows how implementing one framework covers controls in another.
type CoverageEntry struct {
	SourceFramework string  `json:"source_framework"`
	TargetFramework string  `json:"target_framework"`
	SourceControl   string  `json:"source_control"`
	TargetControl   string  `json:"target_control"`
	MappingType     string  `json:"mapping_type"`
	MappingStrength float64 `json:"mapping_strength"`
	SourceStatus    string  `json:"source_status"`
}

func (e *ComplianceEngine) CrossFrameworkCoverage(ctx context.Context, orgID uuid.UUID) ([]CoverageEntry, error) {
	query := `
		SELECT
			sf.code AS source_framework,
			tf.code AS target_framework,
			sc.code AS source_control,
			tc.code AS target_control,
			m.mapping_type::TEXT,
			m.mapping_strength,
			COALESCE(ci.status::TEXT, 'not_implemented')
		FROM framework_control_mappings m
		JOIN framework_controls sc ON m.source_control_id = sc.id
		JOIN framework_controls tc ON m.target_control_id = tc.id
		JOIN compliance_frameworks sf ON sc.framework_id = sf.id
		JOIN compliance_frameworks tf ON tc.framework_id = tf.id
		LEFT JOIN control_implementations ci ON ci.framework_control_id = sc.id AND ci.organization_id = $1
		WHERE sf.id IN (SELECT framework_id FROM organization_frameworks WHERE organization_id = $1)
		ORDER BY sf.code, sc.code`

	rows, err := e.pool.Query(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []CoverageEntry
	for rows.Next() {
		var c CoverageEntry
		if err := rows.Scan(&c.SourceFramework, &c.TargetFramework, &c.SourceControl,
			&c.TargetControl, &c.MappingType, &c.MappingStrength, &c.SourceStatus); err != nil {
			return nil, err
		}
		entries = append(entries, c)
	}
	return entries, nil
}

// DashboardSummary aggregates all key metrics for the executive dashboard.
type DashboardSummary struct {
	OverallComplianceScore float64                `json:"overall_compliance_score"`
	FrameworkScores        []FrameworkScoreEntry   `json:"framework_scores"`
	RiskSummary            map[string]int64        `json:"risk_summary"`
	OpenIncidents          int64                   `json:"open_incidents"`
	OpenFindings           int64                   `json:"open_findings"`
	PoliciesDueForReview   int64                   `json:"policies_due_for_review"`
	VendorsHighRisk        int64                   `json:"vendors_high_risk"`
	BreachesNearDeadline   int64                   `json:"breaches_near_deadline"`
}

type FrameworkScoreEntry struct {
	Code  string  `json:"code"`
	Name  string  `json:"name"`
	Score float64 `json:"score"`
}

func (e *ComplianceEngine) DashboardSummary(ctx context.Context, orgID uuid.UUID) (*DashboardSummary, error) {
	summary := &DashboardSummary{}

	// Framework scores
	rows, err := e.pool.Query(ctx, `
		SELECT framework_code, framework_name, COALESCE(compliance_score, 0)
		FROM v_compliance_score_by_framework WHERE organization_id = $1`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var totalScore float64
	var count int
	for rows.Next() {
		var entry FrameworkScoreEntry
		if err := rows.Scan(&entry.Code, &entry.Name, &entry.Score); err != nil {
			return nil, err
		}
		summary.FrameworkScores = append(summary.FrameworkScores, entry)
		totalScore += entry.Score
		count++
	}
	if count > 0 {
		summary.OverallComplianceScore = math.Round(totalScore/float64(count)*100) / 100
	}

	// Risk summary
	summary.RiskSummary = make(map[string]int64)
	var critCount, highCount, medCount, lowCount int64
	e.pool.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE residual_risk_level = 'critical'),
			COUNT(*) FILTER (WHERE residual_risk_level = 'high'),
			COUNT(*) FILTER (WHERE residual_risk_level = 'medium'),
			COUNT(*) FILTER (WHERE residual_risk_level = 'low')
		FROM risks WHERE organization_id = $1 AND deleted_at IS NULL`, orgID,
	).Scan(&critCount, &highCount, &medCount, &lowCount)
	summary.RiskSummary["critical"] = critCount
	summary.RiskSummary["high"] = highCount
	summary.RiskSummary["medium"] = medCount
	summary.RiskSummary["low"] = lowCount

	// Open incidents
	e.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM incidents
		WHERE organization_id = $1 AND status NOT IN ('resolved','closed') AND deleted_at IS NULL`, orgID,
	).Scan(&summary.OpenIncidents)

	// Open findings
	e.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM audit_findings
		WHERE organization_id = $1 AND status IN ('open','in_progress') AND deleted_at IS NULL`, orgID,
	).Scan(&summary.OpenFindings)

	// Policies due for review
	e.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM policies
		WHERE organization_id = $1 AND next_review_date <= CURRENT_DATE AND deleted_at IS NULL`, orgID,
	).Scan(&summary.PoliciesDueForReview)

	// High/critical risk vendors
	e.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM vendors
		WHERE organization_id = $1 AND risk_tier IN ('critical','high') AND status = 'active' AND deleted_at IS NULL`, orgID,
	).Scan(&summary.VendorsHighRisk)

	// Breaches near deadline
	e.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM incidents
		WHERE organization_id = $1 AND is_data_breach = true AND dpa_notified_at IS NULL
			AND notification_required = true AND deleted_at IS NULL`, orgID,
	).Scan(&summary.BreachesNearDeadline)

	return summary, nil
}
