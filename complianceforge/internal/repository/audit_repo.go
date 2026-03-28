package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/complianceforge/platform/internal/models"
)

type AuditRepo struct {
	pool *pgxpool.Pool
}

func NewAuditRepo(pool *pgxpool.Pool) *AuditRepo {
	return &AuditRepo{pool: pool}
}

func (r *AuditRepo) Create(ctx context.Context, tx pgx.Tx, audit *models.Audit) error {
	var ref string
	if err := tx.QueryRow(ctx, "SELECT generate_audit_ref($1)", audit.OrganizationID).Scan(&ref); err != nil {
		return fmt.Errorf("failed to generate audit ref: %w", err)
	}
	audit.AuditRef = ref

	query := `
		INSERT INTO audits (
			organization_id, audit_ref, title, audit_type, status, description,
			scope, methodology, lead_auditor_id, audit_team_ids, linked_framework_ids,
			planned_start_date, planned_end_date, tags, metadata
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		RETURNING id, created_at, updated_at`

	return tx.QueryRow(ctx, query,
		audit.OrganizationID, audit.AuditRef, audit.Title, audit.AuditType,
		audit.Status, audit.Description, audit.Scope, audit.Methodology,
		audit.LeadAuditorID, audit.AuditTeamIDs, audit.LinkedFrameworkIDs,
		audit.PlannedStartDate, audit.PlannedEndDate, audit.Tags, audit.Metadata,
	).Scan(&audit.ID, &audit.CreatedAt, &audit.UpdatedAt)
}

func (r *AuditRepo) List(ctx context.Context, orgID uuid.UUID, params models.PaginationParams) ([]models.Audit, int64, error) {
	var total int64
	if err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM audits WHERE organization_id = $1 AND deleted_at IS NULL", orgID,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, organization_id, audit_ref, title, audit_type, status,
			description, scope, lead_auditor_id, planned_start_date, planned_end_date,
			actual_start_date, actual_end_date, total_findings, critical_findings,
			high_findings, medium_findings, low_findings, tags, created_at, updated_at
		FROM audits
		WHERE organization_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, orgID, params.PageSize, params.Offset())
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var audits []models.Audit
	for rows.Next() {
		var a models.Audit
		if err := rows.Scan(
			&a.ID, &a.OrganizationID, &a.AuditRef, &a.Title, &a.AuditType,
			&a.Status, &a.Description, &a.Scope, &a.LeadAuditorID,
			&a.PlannedStartDate, &a.PlannedEndDate, &a.ActualStartDate,
			&a.ActualEndDate, &a.TotalFindings, &a.CriticalFindings,
			&a.HighFindings, &a.MediumFindings, &a.LowFindings,
			&a.Tags, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		audits = append(audits, a)
	}
	return audits, total, nil
}

func (r *AuditRepo) GetByID(ctx context.Context, orgID, auditID uuid.UUID) (*models.Audit, error) {
	query := `
		SELECT id, organization_id, audit_ref, title, audit_type, status,
			description, scope, methodology, lead_auditor_id, audit_team_ids,
			linked_framework_ids, planned_start_date, planned_end_date,
			actual_start_date, actual_end_date, total_findings, critical_findings,
			high_findings, medium_findings, low_findings, report_file_path,
			conclusion, tags, metadata, created_at, updated_at
		FROM audits
		WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL`

	var a models.Audit
	err := r.pool.QueryRow(ctx, query, auditID, orgID).Scan(
		&a.ID, &a.OrganizationID, &a.AuditRef, &a.Title, &a.AuditType,
		&a.Status, &a.Description, &a.Scope, &a.Methodology, &a.LeadAuditorID,
		&a.AuditTeamIDs, &a.LinkedFrameworkIDs, &a.PlannedStartDate,
		&a.PlannedEndDate, &a.ActualStartDate, &a.ActualEndDate,
		&a.TotalFindings, &a.CriticalFindings, &a.HighFindings,
		&a.MediumFindings, &a.LowFindings, &a.ReportFilePath,
		&a.Conclusion, &a.Tags, &a.Metadata, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("audit not found: %w", err)
	}
	return &a, nil
}

func (r *AuditRepo) CreateFinding(ctx context.Context, tx pgx.Tx, finding *models.AuditFinding) error {
	query := `
		INSERT INTO audit_findings (
			organization_id, audit_id, finding_ref, title, description, severity,
			status, finding_type, control_id, root_cause, recommendation,
			responsible_user_id, due_date, linked_risk_id, metadata
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		RETURNING id, created_at, updated_at`

	return tx.QueryRow(ctx, query,
		finding.OrganizationID, finding.AuditID, finding.FindingRef, finding.Title,
		finding.Description, finding.Severity, finding.Status, finding.FindingType,
		finding.ControlID, finding.RootCause, finding.Recommendation,
		finding.ResponsibleUserID, finding.DueDate, finding.LinkedRiskID, finding.Metadata,
	).Scan(&finding.ID, &finding.CreatedAt, &finding.UpdatedAt)
}

func (r *AuditRepo) ListFindings(ctx context.Context, orgID, auditID uuid.UUID) ([]models.AuditFinding, error) {
	query := `
		SELECT id, organization_id, audit_id, finding_ref, title, description,
			severity, status, finding_type, control_id, root_cause, recommendation,
			management_response, responsible_user_id, due_date, closed_date,
			linked_risk_id, metadata, created_at, updated_at
		FROM audit_findings
		WHERE audit_id = $1 AND organization_id = $2 AND deleted_at IS NULL
		ORDER BY severity DESC, created_at ASC`

	rows, err := r.pool.Query(ctx, query, auditID, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var findings []models.AuditFinding
	for rows.Next() {
		var f models.AuditFinding
		if err := rows.Scan(
			&f.ID, &f.OrganizationID, &f.AuditID, &f.FindingRef, &f.Title,
			&f.Description, &f.Severity, &f.Status, &f.FindingType, &f.ControlID,
			&f.RootCause, &f.Recommendation, &f.ManagementResponse,
			&f.ResponsibleUserID, &f.DueDate, &f.ClosedDate,
			&f.LinkedRiskID, &f.Metadata, &f.CreatedAt, &f.UpdatedAt,
		); err != nil {
			return nil, err
		}
		findings = append(findings, f)
	}
	return findings, nil
}

func (r *AuditRepo) GetFindingsStats(ctx context.Context, orgID uuid.UUID) (map[string]interface{}, error) {
	var total, open, overdue, critical, high int64

	err := r.pool.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE status IN ('open', 'in_progress')),
			COUNT(*) FILTER (WHERE status = 'open' AND due_date < CURRENT_DATE),
			COUNT(*) FILTER (WHERE severity = 'critical' AND status NOT IN ('remediated', 'closed')),
			COUNT(*) FILTER (WHERE severity = 'high' AND status NOT IN ('remediated', 'closed'))
		FROM audit_findings
		WHERE organization_id = $1 AND deleted_at IS NULL`, orgID,
	).Scan(&total, &open, &overdue, &critical, &high)

	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_findings":    total,
		"open_findings":     open,
		"overdue_findings":  overdue,
		"critical_findings": critical,
		"high_findings":     high,
	}, nil
}
