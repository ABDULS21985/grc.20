package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/complianceforge/platform/internal/models"
)

type OrganizationRepo struct {
	pool *pgxpool.Pool
}

func NewOrganizationRepo(pool *pgxpool.Pool) *OrganizationRepo {
	return &OrganizationRepo{pool: pool}
}

func (r *OrganizationRepo) Create(ctx context.Context, org *models.Organization) error {
	// Generate slug from name
	slug := strings.ToLower(strings.ReplaceAll(org.Name, " ", "-"))
	slug = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return -1
	}, slug)

	query := `
		INSERT INTO organizations (
			name, slug, legal_name, industry, country_code, status, tier,
			timezone, default_language, employee_count_range, metadata
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING id, created_at, updated_at`

	return r.pool.QueryRow(ctx, query,
		org.Name, slug, org.LegalName, org.Industry, org.CountryCode,
		org.Status, org.Tier, org.Timezone, org.DefaultLanguage,
		org.EmployeeCountRange, org.Metadata,
	).Scan(&org.ID, &org.CreatedAt, &org.UpdatedAt)
}

func (r *OrganizationRepo) GetByID(ctx context.Context, orgID uuid.UUID) (*models.Organization, error) {
	query := `
		SELECT id, name, slug, legal_name, registration_number, tax_id,
			industry, sector, country_code, headquarters_address, status, tier,
			settings, branding, timezone, default_language, supported_languages,
			employee_count_range, annual_revenue_range, parent_organization_id,
			metadata, created_at, updated_at
		FROM organizations
		WHERE id = $1 AND deleted_at IS NULL`

	var org models.Organization
	err := r.pool.QueryRow(ctx, query, orgID).Scan(
		&org.ID, &org.Name, &org.Slug, &org.LegalName, &org.RegistrationNumber,
		&org.TaxID, &org.Industry, &org.Sector, &org.CountryCode,
		&org.HeadquartersAddress, &org.Status, &org.Tier, &org.Settings,
		&org.Branding, &org.Timezone, &org.DefaultLanguage, &org.SupportedLanguages,
		&org.EmployeeCountRange, &org.AnnualRevenueRange,
		&org.ParentOrganizationID, &org.Metadata, &org.CreatedAt, &org.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}
	return &org, nil
}

func (r *OrganizationRepo) Update(ctx context.Context, orgID uuid.UUID, req models.UpdateOrganizationRequest) error {
	sets := []string{}
	args := []interface{}{}
	argIdx := 1

	if req.Name != nil {
		sets = append(sets, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, *req.Name)
		argIdx++
	}
	if req.LegalName != nil {
		sets = append(sets, fmt.Sprintf("legal_name = $%d", argIdx))
		args = append(args, *req.LegalName)
		argIdx++
	}
	if req.Industry != nil {
		sets = append(sets, fmt.Sprintf("industry = $%d", argIdx))
		args = append(args, *req.Industry)
		argIdx++
	}
	if req.CountryCode != nil {
		sets = append(sets, fmt.Sprintf("country_code = $%d", argIdx))
		args = append(args, *req.CountryCode)
		argIdx++
	}
	if req.Branding != nil {
		sets = append(sets, fmt.Sprintf("branding = $%d", argIdx))
		args = append(args, *req.Branding)
		argIdx++
	}
	if req.Settings != nil {
		sets = append(sets, fmt.Sprintf("settings = $%d", argIdx))
		args = append(args, *req.Settings)
		argIdx++
	}

	if len(sets) == 0 {
		return nil
	}

	query := fmt.Sprintf("UPDATE organizations SET %s WHERE id = $%d AND deleted_at IS NULL",
		strings.Join(sets, ", "), argIdx)
	args = append(args, orgID)

	_, err := r.pool.Exec(ctx, query, args...)
	return err
}

// CreateAuditLog records an action in the immutable audit trail.
func (r *OrganizationRepo) CreateAuditLog(ctx context.Context, log *models.AuditLog) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO audit_logs (organization_id, user_id, action, entity_type, entity_id,
			changes, ip_address, user_agent, metadata)
		VALUES ($1,$2,$3,$4,$5,$6,$7::INET,$8,$9)`,
		log.OrganizationID, log.UserID, log.Action, log.EntityType, log.EntityID,
		log.Changes, log.IPAddress, log.UserAgent, log.Metadata,
	)
	return err
}
