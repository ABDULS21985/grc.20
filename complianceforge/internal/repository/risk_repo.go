package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/complianceforge/platform/internal/models"
)

// RiskRepo handles database operations for the risk register.
type RiskRepo struct {
	pool *pgxpool.Pool
}

func NewRiskRepo(pool *pgxpool.Pool) *RiskRepo {
	return &RiskRepo{pool: pool}
}

// Create inserts a new risk into the register.
func (r *RiskRepo) Create(ctx context.Context, tx pgx.Tx, risk *models.Risk) error {
	// Generate risk reference
	var ref string
	err := tx.QueryRow(ctx, "SELECT generate_risk_ref($1)", risk.OrganizationID).Scan(&ref)
	if err != nil {
		return fmt.Errorf("failed to generate risk ref: %w", err)
	}
	risk.RiskRef = ref

	query := `
		INSERT INTO risks (
			organization_id, risk_ref, title, description, risk_category_id, risk_source,
			risk_type, status, owner_user_id, inherent_likelihood, inherent_impact,
			residual_likelihood, residual_impact, target_likelihood, target_impact,
			financial_impact_eur, impact_description, risk_velocity, risk_proximity,
			review_frequency, tags, is_emerging, metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20, $21, $22, $23
		) RETURNING id, created_at, updated_at,
			inherent_risk_score, inherent_risk_level,
			residual_risk_score, residual_risk_level,
			target_risk_score, target_risk_level`

	return tx.QueryRow(ctx, query,
		risk.OrganizationID, risk.RiskRef, risk.Title, risk.Description,
		risk.RiskCategoryID, risk.RiskSource, risk.RiskType, risk.Status,
		risk.OwnerUserID, risk.InherentLikelihood, risk.InherentImpact,
		risk.ResidualLikelihood, risk.ResidualImpact, risk.TargetLikelihood,
		risk.TargetImpact, risk.FinancialImpactEUR, risk.ImpactDescription,
		risk.RiskVelocity, risk.RiskProximity, risk.ReviewFrequency,
		risk.Tags, risk.IsEmerging, risk.Metadata,
	).Scan(
		&risk.ID, &risk.CreatedAt, &risk.UpdatedAt,
		&risk.InherentRiskScore, &risk.InherentRiskLevel,
		&risk.ResidualRiskScore, &risk.ResidualRiskLevel,
		&risk.TargetRiskScore, &risk.TargetRiskLevel,
	)
}

// List returns paginated risks for an organization.
func (r *RiskRepo) List(ctx context.Context, orgID uuid.UUID, params models.PaginationParams) ([]models.Risk, int64, error) {
	countQuery := `SELECT COUNT(*) FROM risks WHERE organization_id = $1 AND deleted_at IS NULL`
	var total int64
	if err := r.pool.QueryRow(ctx, countQuery, orgID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT r.id, r.organization_id, r.risk_ref, r.title, r.description,
			r.risk_category_id, r.risk_source, r.risk_type, r.status,
			r.owner_user_id, r.inherent_likelihood, r.inherent_impact,
			r.inherent_risk_score, r.inherent_risk_level,
			r.residual_likelihood, r.residual_impact,
			r.residual_risk_score, r.residual_risk_level,
			r.financial_impact_eur, r.risk_velocity, r.risk_proximity,
			r.next_review_date, r.review_frequency, r.tags, r.is_emerging,
			r.created_at, r.updated_at
		FROM risks r
		WHERE r.organization_id = $1 AND r.deleted_at IS NULL
		ORDER BY r.%s %s
		LIMIT $2 OFFSET $3`, sanitizeSortField(params.SortBy, "residual_risk_score"), sanitizeSortDir(params.SortDir))

	rows, err := r.pool.Query(ctx, query, orgID, params.PageSize, params.Offset())
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list risks: %w", err)
	}
	defer rows.Close()

	var risks []models.Risk
	for rows.Next() {
		var risk models.Risk
		err := rows.Scan(
			&risk.ID, &risk.OrganizationID, &risk.RiskRef, &risk.Title, &risk.Description,
			&risk.RiskCategoryID, &risk.RiskSource, &risk.RiskType, &risk.Status,
			&risk.OwnerUserID, &risk.InherentLikelihood, &risk.InherentImpact,
			&risk.InherentRiskScore, &risk.InherentRiskLevel,
			&risk.ResidualLikelihood, &risk.ResidualImpact,
			&risk.ResidualRiskScore, &risk.ResidualRiskLevel,
			&risk.FinancialImpactEUR, &risk.RiskVelocity, &risk.RiskProximity,
			&risk.NextReviewDate, &risk.ReviewFrequency, &risk.Tags, &risk.IsEmerging,
			&risk.CreatedAt, &risk.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan risk: %w", err)
		}
		risks = append(risks, risk)
	}
	return risks, total, nil
}

// GetByID returns a single risk by ID.
func (r *RiskRepo) GetByID(ctx context.Context, orgID, riskID uuid.UUID) (*models.Risk, error) {
	query := `
		SELECT id, organization_id, risk_ref, title, description,
			risk_category_id, risk_source, risk_type, status, owner_user_id,
			inherent_likelihood, inherent_impact, inherent_risk_score, inherent_risk_level,
			residual_likelihood, residual_impact, residual_risk_score, residual_risk_level,
			target_likelihood, target_impact, target_risk_score, target_risk_level,
			financial_impact_eur, impact_description, risk_velocity, risk_proximity,
			identified_date, last_assessed_date, next_review_date, review_frequency,
			linked_regulations, tags, is_emerging, metadata, created_at, updated_at
		FROM risks
		WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL`

	var risk models.Risk
	err := r.pool.QueryRow(ctx, query, riskID, orgID).Scan(
		&risk.ID, &risk.OrganizationID, &risk.RiskRef, &risk.Title, &risk.Description,
		&risk.RiskCategoryID, &risk.RiskSource, &risk.RiskType, &risk.Status, &risk.OwnerUserID,
		&risk.InherentLikelihood, &risk.InherentImpact, &risk.InherentRiskScore, &risk.InherentRiskLevel,
		&risk.ResidualLikelihood, &risk.ResidualImpact, &risk.ResidualRiskScore, &risk.ResidualRiskLevel,
		&risk.TargetLikelihood, &risk.TargetImpact, &risk.TargetRiskScore, &risk.TargetRiskLevel,
		&risk.FinancialImpactEUR, &risk.ImpactDescription, &risk.RiskVelocity, &risk.RiskProximity,
		&risk.IdentifiedDate, &risk.LastAssessedDate, &risk.NextReviewDate, &risk.ReviewFrequency,
		&risk.LinkedRegulations, &risk.Tags, &risk.IsEmerging, &risk.Metadata,
		&risk.CreatedAt, &risk.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("risk not found: %w", err)
	}
	return &risk, nil
}

// GetHeatmapData returns risk heatmap data for an organization.
func (r *RiskRepo) GetHeatmapData(ctx context.Context, orgID uuid.UUID) ([]models.RiskHeatmapEntry, error) {
	query := `SELECT risk_id, risk_ref, title, category_name,
		inherent_likelihood, inherent_impact, residual_likelihood, residual_impact,
		residual_risk_level, owner_name, status
		FROM v_risk_heatmap WHERE organization_id = $1`

	rows, err := r.pool.Query(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []models.RiskHeatmapEntry
	for rows.Next() {
		var e models.RiskHeatmapEntry
		if err := rows.Scan(&e.RiskID, &e.RiskRef, &e.Title, &e.CategoryName,
			&e.InherentLikelihood, &e.InherentImpact, &e.ResidualLikelihood, &e.ResidualImpact,
			&e.ResidualRiskLevel, &e.OwnerName, &e.Status); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, nil
}

// Helper to prevent SQL injection in ORDER BY
func sanitizeSortField(field, defaultField string) string {
	allowed := map[string]bool{
		"created_at": true, "updated_at": true, "title": true,
		"residual_risk_score": true, "inherent_risk_score": true,
		"risk_ref": true, "status": true, "financial_impact_eur": true,
	}
	if allowed[field] {
		return field
	}
	return defaultField
}

func sanitizeSortDir(dir string) string {
	if dir == "asc" {
		return "ASC"
	}
	return "DESC"
}
