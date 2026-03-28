package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/complianceforge/platform/internal/models"
)

type IncidentRepo struct {
	pool *pgxpool.Pool
}

func NewIncidentRepo(pool *pgxpool.Pool) *IncidentRepo {
	return &IncidentRepo{pool: pool}
}

func (r *IncidentRepo) Create(ctx context.Context, tx pgx.Tx, inc *models.Incident) error {
	var ref string
	if err := tx.QueryRow(ctx, "SELECT generate_incident_ref($1)", inc.OrganizationID).Scan(&ref); err != nil {
		return fmt.Errorf("failed to generate incident ref: %w", err)
	}
	inc.IncidentRef = ref

	query := `
		INSERT INTO incidents (
			organization_id, incident_ref, title, description, incident_type, severity,
			status, reported_by, reported_at, assigned_to,
			is_data_breach, data_subjects_affected, data_categories_affected,
			is_nis2_reportable, impact_description, financial_impact_eur,
			linked_control_ids, linked_risk_ids, tags, metadata
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20)
		RETURNING id, notification_deadline, created_at, updated_at`

	return tx.QueryRow(ctx, query,
		inc.OrganizationID, inc.IncidentRef, inc.Title, inc.Description,
		inc.IncidentType, inc.Severity, inc.Status, inc.ReportedBy, inc.ReportedAt,
		inc.AssignedTo, inc.IsDataBreach, inc.DataSubjectsAffected,
		inc.DataCategoriesAffected, inc.IsNIS2Reportable, inc.ImpactDescription,
		inc.FinancialImpactEUR, inc.LinkedControlIDs, inc.LinkedRiskIDs,
		inc.Tags, inc.Metadata,
	).Scan(&inc.ID, &inc.NotificationDeadline, &inc.CreatedAt, &inc.UpdatedAt)
}

func (r *IncidentRepo) List(ctx context.Context, orgID uuid.UUID, params models.PaginationParams) ([]models.Incident, int64, error) {
	var total int64
	if err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM incidents WHERE organization_id = $1 AND deleted_at IS NULL", orgID,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, organization_id, incident_ref, title, description, incident_type,
			severity, status, reported_by, reported_at, assigned_to,
			is_data_breach, data_subjects_affected, notification_required,
			notification_deadline, dpa_notified_at,
			is_nis2_reportable, nis2_early_warning_at,
			financial_impact_eur, contained_at, resolved_at, closed_at,
			tags, created_at, updated_at
		FROM incidents
		WHERE organization_id = $1 AND deleted_at IS NULL
		ORDER BY reported_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, orgID, params.PageSize, params.Offset())
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var incidents []models.Incident
	for rows.Next() {
		var inc models.Incident
		if err := rows.Scan(
			&inc.ID, &inc.OrganizationID, &inc.IncidentRef, &inc.Title, &inc.Description,
			&inc.IncidentType, &inc.Severity, &inc.Status, &inc.ReportedBy, &inc.ReportedAt,
			&inc.AssignedTo, &inc.IsDataBreach, &inc.DataSubjectsAffected,
			&inc.NotificationRequired, &inc.NotificationDeadline, &inc.DPANotifiedAt,
			&inc.IsNIS2Reportable, &inc.NIS2EarlyWarningAt,
			&inc.FinancialImpactEUR, &inc.ContainedAt, &inc.ResolvedAt, &inc.ClosedAt,
			&inc.Tags, &inc.CreatedAt, &inc.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		incidents = append(incidents, inc)
	}
	return incidents, total, nil
}

func (r *IncidentRepo) GetByID(ctx context.Context, orgID, incidentID uuid.UUID) (*models.Incident, error) {
	query := `
		SELECT id, organization_id, incident_ref, title, description, incident_type,
			severity, status, reported_by, reported_at, assigned_to,
			is_data_breach, data_subjects_affected, data_categories_affected,
			notification_required, dpa_notified_at, data_subjects_notified_at,
			notification_deadline,
			is_nis2_reportable, nis2_early_warning_at, nis2_notification_at, nis2_final_report_at,
			impact_description, financial_impact_eur, root_cause,
			containment_actions, remediation_actions, lessons_learned,
			contained_at, resolved_at, closed_at,
			linked_control_ids, linked_risk_ids, tags, metadata,
			created_at, updated_at
		FROM incidents
		WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL`

	var inc models.Incident
	err := r.pool.QueryRow(ctx, query, incidentID, orgID).Scan(
		&inc.ID, &inc.OrganizationID, &inc.IncidentRef, &inc.Title, &inc.Description,
		&inc.IncidentType, &inc.Severity, &inc.Status, &inc.ReportedBy, &inc.ReportedAt,
		&inc.AssignedTo, &inc.IsDataBreach, &inc.DataSubjectsAffected,
		&inc.DataCategoriesAffected, &inc.NotificationRequired, &inc.DPANotifiedAt,
		&inc.DataSubjectsNotifiedAt, &inc.NotificationDeadline,
		&inc.IsNIS2Reportable, &inc.NIS2EarlyWarningAt, &inc.NIS2NotificationAt,
		&inc.NIS2FinalReportAt, &inc.ImpactDescription, &inc.FinancialImpactEUR,
		&inc.RootCause, &inc.ContainmentActions, &inc.RemediationActions,
		&inc.LessonsLearned, &inc.ContainedAt, &inc.ResolvedAt, &inc.ClosedAt,
		&inc.LinkedControlIDs, &inc.LinkedRiskIDs, &inc.Tags, &inc.Metadata,
		&inc.CreatedAt, &inc.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("incident not found: %w", err)
	}
	return &inc, nil
}

// RecordDPANotification marks a GDPR breach as notified to the supervisory authority.
func (r *IncidentRepo) RecordDPANotification(ctx context.Context, orgID, incidentID uuid.UUID) error {
	now := time.Now()
	_, err := r.pool.Exec(ctx, `
		UPDATE incidents SET dpa_notified_at = $1
		WHERE id = $2 AND organization_id = $3 AND is_data_breach = true AND deleted_at IS NULL`,
		now, incidentID, orgID,
	)
	return err
}

// RecordNIS2EarlyWarning marks a NIS2 early warning (24-hour) as submitted.
func (r *IncidentRepo) RecordNIS2EarlyWarning(ctx context.Context, orgID, incidentID uuid.UUID) error {
	now := time.Now()
	_, err := r.pool.Exec(ctx, `
		UPDATE incidents SET nis2_early_warning_at = $1
		WHERE id = $2 AND organization_id = $3 AND is_nis2_reportable = true AND deleted_at IS NULL`,
		now, incidentID, orgID,
	)
	return err
}

// GetBreachesNearingDeadline returns data breaches approaching the 72-hour notification deadline.
func (r *IncidentRepo) GetBreachesNearingDeadline(ctx context.Context, orgID uuid.UUID) ([]models.Incident, error) {
	query := `
		SELECT id, incident_ref, title, severity, reported_at, notification_deadline,
			dpa_notified_at, data_subjects_affected
		FROM incidents
		WHERE organization_id = $1
			AND is_data_breach = true
			AND notification_required = true
			AND dpa_notified_at IS NULL
			AND deleted_at IS NULL
		ORDER BY notification_deadline ASC`

	rows, err := r.pool.Query(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var incidents []models.Incident
	for rows.Next() {
		var inc models.Incident
		if err := rows.Scan(&inc.ID, &inc.IncidentRef, &inc.Title, &inc.Severity,
			&inc.ReportedAt, &inc.NotificationDeadline, &inc.DPANotifiedAt,
			&inc.DataSubjectsAffected,
		); err != nil {
			return nil, err
		}
		incidents = append(incidents, inc)
	}
	return incidents, nil
}

// GetDashboardStats returns incident metrics for the dashboard.
func (r *IncidentRepo) GetDashboardStats(ctx context.Context, orgID uuid.UUID) (map[string]interface{}, error) {
	var totalOpen, totalClosed, criticalOpen, dataBreaches, nis2Reportable int64
	var avgResolutionHours float64

	err := r.pool.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE status NOT IN ('resolved', 'closed')),
			COUNT(*) FILTER (WHERE status IN ('resolved', 'closed')),
			COUNT(*) FILTER (WHERE severity = 'critical' AND status NOT IN ('resolved', 'closed')),
			COUNT(*) FILTER (WHERE is_data_breach = true),
			COUNT(*) FILTER (WHERE is_nis2_reportable = true),
			COALESCE(AVG(EXTRACT(EPOCH FROM (resolved_at - reported_at)) / 3600)
				FILTER (WHERE resolved_at IS NOT NULL), 0)
		FROM incidents
		WHERE organization_id = $1 AND deleted_at IS NULL`, orgID,
	).Scan(&totalOpen, &totalClosed, &criticalOpen, &dataBreaches, &nis2Reportable, &avgResolutionHours)

	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"open_incidents":         totalOpen,
		"closed_incidents":       totalClosed,
		"critical_open":          criticalOpen,
		"data_breaches":          dataBreaches,
		"nis2_reportable":        nis2Reportable,
		"avg_resolution_hours":   avgResolutionHours,
	}, nil
}
