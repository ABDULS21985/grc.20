package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/complianceforge/platform/internal/models"
)

// FrameworkRepo handles all database operations for compliance frameworks.
type FrameworkRepo struct {
	pool *pgxpool.Pool
}

// NewFrameworkRepo creates a new FrameworkRepo.
func NewFrameworkRepo(pool *pgxpool.Pool) *FrameworkRepo {
	return &FrameworkRepo{pool: pool}
}

// ListSystemFrameworks returns all platform-provided (system) frameworks.
func (r *FrameworkRepo) ListSystemFrameworks(ctx context.Context) ([]models.ComplianceFramework, error) {
	query := `
		SELECT id, organization_id, code, name, full_name, version, description, issuing_body,
			category, applicable_regions, applicable_industries, is_system_framework, is_active,
			effective_date, sunset_date, total_controls, icon_url, color_hex, metadata,
			created_at, updated_at
		FROM compliance_frameworks
		WHERE is_system_framework = true AND deleted_at IS NULL AND is_active = true
		ORDER BY name`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list frameworks: %w", err)
	}
	defer rows.Close()

	var frameworks []models.ComplianceFramework
	for rows.Next() {
		var f models.ComplianceFramework
		err := rows.Scan(
			&f.ID, &f.OrganizationID, &f.Code, &f.Name, &f.FullName, &f.Version, &f.Description,
			&f.IssuingBody, &f.Category, &f.ApplicableRegions, &f.ApplicableIndustries,
			&f.IsSystemFramework, &f.IsActive, &f.EffectiveDate, &f.SunsetDate,
			&f.TotalControls, &f.IconURL, &f.ColorHex, &f.Metadata, &f.CreatedAt, &f.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan framework: %w", err)
		}
		frameworks = append(frameworks, f)
	}
	return frameworks, nil
}

// GetByID returns a single framework by its ID.
func (r *FrameworkRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.ComplianceFramework, error) {
	query := `
		SELECT id, organization_id, code, name, full_name, version, description, issuing_body,
			category, applicable_regions, applicable_industries, is_system_framework, is_active,
			effective_date, sunset_date, total_controls, icon_url, color_hex, metadata,
			created_at, updated_at
		FROM compliance_frameworks
		WHERE id = $1 AND deleted_at IS NULL`

	var f models.ComplianceFramework
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&f.ID, &f.OrganizationID, &f.Code, &f.Name, &f.FullName, &f.Version, &f.Description,
		&f.IssuingBody, &f.Category, &f.ApplicableRegions, &f.ApplicableIndustries,
		&f.IsSystemFramework, &f.IsActive, &f.EffectiveDate, &f.SunsetDate,
		&f.TotalControls, &f.IconURL, &f.ColorHex, &f.Metadata, &f.CreatedAt, &f.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("framework not found: %w", err)
	}
	return &f, nil
}

// GetDomainsByFrameworkID returns all domains for a framework.
func (r *FrameworkRepo) GetDomainsByFrameworkID(ctx context.Context, frameworkID uuid.UUID) ([]models.FrameworkDomain, error) {
	query := `
		SELECT id, framework_id, code, name, description, sort_order, parent_domain_id,
			depth_level, total_controls, metadata, created_at, updated_at
		FROM framework_domains
		WHERE framework_id = $1
		ORDER BY sort_order`

	rows, err := r.pool.Query(ctx, query, frameworkID)
	if err != nil {
		return nil, fmt.Errorf("failed to list domains: %w", err)
	}
	defer rows.Close()

	var domains []models.FrameworkDomain
	for rows.Next() {
		var d models.FrameworkDomain
		err := rows.Scan(
			&d.ID, &d.FrameworkID, &d.Code, &d.Name, &d.Description, &d.SortOrder,
			&d.ParentDomainID, &d.DepthLevel, &d.TotalControls, &d.Metadata,
			&d.CreatedAt, &d.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan domain: %w", err)
		}
		domains = append(domains, d)
	}
	return domains, nil
}

// GetControlsByFrameworkID returns all controls for a framework with pagination.
func (r *FrameworkRepo) GetControlsByFrameworkID(ctx context.Context, frameworkID uuid.UUID, params models.PaginationParams) ([]models.FrameworkControl, int64, error) {
	countQuery := `SELECT COUNT(*) FROM framework_controls WHERE framework_id = $1`
	var total int64
	if err := r.pool.QueryRow(ctx, countQuery, frameworkID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, framework_id, domain_id, code, title, description, guidance, objective,
			control_type, implementation_type, is_mandatory, priority, sort_order,
			parent_control_id, depth_level, evidence_requirements, test_procedures,
			references, keywords, metadata, created_at, updated_at
		FROM framework_controls
		WHERE framework_id = $1
		ORDER BY sort_order
		LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, frameworkID, params.PageSize, params.Offset())
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list controls: %w", err)
	}
	defer rows.Close()

	var controls []models.FrameworkControl
	for rows.Next() {
		var c models.FrameworkControl
		err := rows.Scan(
			&c.ID, &c.FrameworkID, &c.DomainID, &c.Code, &c.Title, &c.Description,
			&c.Guidance, &c.Objective, &c.ControlType, &c.ImplementationType,
			&c.IsMandatory, &c.Priority, &c.SortOrder, &c.ParentControlID, &c.DepthLevel,
			&c.EvidenceRequirements, &c.TestProcedures, &c.References, &c.Keywords,
			&c.Metadata, &c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan control: %w", err)
		}
		controls = append(controls, c)
	}
	return controls, total, nil
}

// SearchControls performs full-text search across framework controls.
func (r *FrameworkRepo) SearchControls(ctx context.Context, searchQuery string, limit int) ([]models.FrameworkControl, error) {
	query := `
		SELECT id, framework_id, domain_id, code, title, description, guidance, objective,
			control_type, implementation_type, is_mandatory, priority, sort_order,
			parent_control_id, depth_level, evidence_requirements, test_procedures,
			references, keywords, metadata, created_at, updated_at
		FROM framework_controls
		WHERE search_vector @@ plainto_tsquery('english', $1)
		ORDER BY ts_rank(search_vector, plainto_tsquery('english', $1)) DESC
		LIMIT $2`

	rows, err := r.pool.Query(ctx, query, searchQuery, limit)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}
	defer rows.Close()

	var controls []models.FrameworkControl
	for rows.Next() {
		var c models.FrameworkControl
		err := rows.Scan(
			&c.ID, &c.FrameworkID, &c.DomainID, &c.Code, &c.Title, &c.Description,
			&c.Guidance, &c.Objective, &c.ControlType, &c.ImplementationType,
			&c.IsMandatory, &c.Priority, &c.SortOrder, &c.ParentControlID, &c.DepthLevel,
			&c.EvidenceRequirements, &c.TestProcedures, &c.References, &c.Keywords,
			&c.Metadata, &c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan search result: %w", err)
		}
		controls = append(controls, c)
	}
	return controls, nil
}

// GetComplianceScores returns compliance scores per framework for an organization.
func (r *FrameworkRepo) GetComplianceScores(ctx context.Context, orgID uuid.UUID) ([]models.ComplianceScoreSummary, error) {
	query := `
		SELECT framework_id, framework_code, framework_name, total_controls,
			implemented, partially_implemented, not_implemented, not_applicable,
			COALESCE(compliance_score, 0) AS compliance_score,
			COALESCE(maturity_avg, 0) AS maturity_avg
		FROM v_compliance_score_by_framework
		WHERE organization_id = $1`

	rows, err := r.pool.Query(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get compliance scores: %w", err)
	}
	defer rows.Close()

	var scores []models.ComplianceScoreSummary
	for rows.Next() {
		var s models.ComplianceScoreSummary
		err := rows.Scan(
			&s.FrameworkID, &s.FrameworkCode, &s.FrameworkName, &s.TotalControls,
			&s.Implemented, &s.PartiallyImpl, &s.NotImplemented, &s.NotApplicable,
			&s.ComplianceScore, &s.MaturityAvg,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan score: %w", err)
		}
		scores = append(scores, s)
	}
	return scores, nil
}
