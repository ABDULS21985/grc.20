package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/complianceforge/platform/internal/models"
)

type VendorRepo struct {
	pool *pgxpool.Pool
}

func NewVendorRepo(pool *pgxpool.Pool) *VendorRepo {
	return &VendorRepo{pool: pool}
}

func (r *VendorRepo) Create(ctx context.Context, tx pgx.Tx, vendor *models.Vendor) error {
	var ref string
	if err := tx.QueryRow(ctx, "SELECT generate_vendor_ref($1)", vendor.OrganizationID).Scan(&ref); err != nil {
		return fmt.Errorf("failed to generate vendor ref: %w", err)
	}
	vendor.VendorRef = ref

	query := `
		INSERT INTO vendors (
			organization_id, vendor_ref, name, legal_name, website, industry,
			country_code, contact_name, contact_email, contact_phone, status,
			risk_tier, service_description, data_processing, data_categories,
			contract_start_date, contract_end_date, contract_value,
			assessment_frequency, certifications, dpa_in_place, dpa_signed_date,
			owner_user_id, tags, metadata
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25)
		RETURNING id, created_at, updated_at`

	return tx.QueryRow(ctx, query,
		vendor.OrganizationID, vendor.VendorRef, vendor.Name, vendor.LegalName,
		vendor.Website, vendor.Industry, vendor.CountryCode, vendor.ContactName,
		vendor.ContactEmail, vendor.ContactPhone, vendor.Status, vendor.RiskTier,
		vendor.ServiceDescription, vendor.DataProcessing, vendor.DataCategories,
		vendor.ContractStartDate, vendor.ContractEndDate, vendor.ContractValue,
		vendor.AssessmentFrequency, vendor.Certifications, vendor.DPAInPlace,
		vendor.DPASignedDate, vendor.OwnerUserID, vendor.Tags, vendor.Metadata,
	).Scan(&vendor.ID, &vendor.CreatedAt, &vendor.UpdatedAt)
}

func (r *VendorRepo) List(ctx context.Context, orgID uuid.UUID, params models.PaginationParams) ([]models.Vendor, int64, error) {
	var total int64
	if err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM vendors WHERE organization_id = $1 AND deleted_at IS NULL", orgID,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, organization_id, vendor_ref, name, legal_name, website, industry,
			country_code, contact_name, contact_email, status, risk_tier, risk_score,
			data_processing, contract_start_date, contract_end_date, contract_value,
			last_assessment_date, next_assessment_date, certifications,
			dpa_in_place, owner_user_id, tags, created_at, updated_at
		FROM vendors
		WHERE organization_id = $1 AND deleted_at IS NULL
		ORDER BY risk_score DESC, name ASC
		LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, orgID, params.PageSize, params.Offset())
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var vendors []models.Vendor
	for rows.Next() {
		var v models.Vendor
		if err := rows.Scan(
			&v.ID, &v.OrganizationID, &v.VendorRef, &v.Name, &v.LegalName,
			&v.Website, &v.Industry, &v.CountryCode, &v.ContactName,
			&v.ContactEmail, &v.Status, &v.RiskTier, &v.RiskScore,
			&v.DataProcessing, &v.ContractStartDate, &v.ContractEndDate,
			&v.ContractValue, &v.LastAssessmentDate, &v.NextAssessmentDate,
			&v.Certifications, &v.DPAInPlace, &v.OwnerUserID,
			&v.Tags, &v.CreatedAt, &v.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		vendors = append(vendors, v)
	}
	return vendors, total, nil
}

func (r *VendorRepo) GetByID(ctx context.Context, orgID, vendorID uuid.UUID) (*models.Vendor, error) {
	query := `
		SELECT id, organization_id, vendor_ref, name, legal_name, website, industry,
			country_code, contact_name, contact_email, contact_phone, status,
			risk_tier, risk_score, service_description, data_processing,
			data_categories, contract_start_date, contract_end_date, contract_value,
			last_assessment_date, next_assessment_date, assessment_frequency,
			certifications, dpa_in_place, dpa_signed_date, owner_user_id,
			tags, metadata, created_at, updated_at
		FROM vendors
		WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL`

	var v models.Vendor
	err := r.pool.QueryRow(ctx, query, vendorID, orgID).Scan(
		&v.ID, &v.OrganizationID, &v.VendorRef, &v.Name, &v.LegalName,
		&v.Website, &v.Industry, &v.CountryCode, &v.ContactName,
		&v.ContactEmail, &v.ContactPhone, &v.Status, &v.RiskTier, &v.RiskScore,
		&v.ServiceDescription, &v.DataProcessing, &v.DataCategories,
		&v.ContractStartDate, &v.ContractEndDate, &v.ContractValue,
		&v.LastAssessmentDate, &v.NextAssessmentDate, &v.AssessmentFrequency,
		&v.Certifications, &v.DPAInPlace, &v.DPASignedDate, &v.OwnerUserID,
		&v.Tags, &v.Metadata, &v.CreatedAt, &v.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("vendor not found: %w", err)
	}
	return &v, nil
}

func (r *VendorRepo) GetDashboardStats(ctx context.Context, orgID uuid.UUID) (map[string]interface{}, error) {
	var total, critical, high, medium, low, withoutDPA, assessmentOverdue int64

	err := r.pool.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE risk_tier = 'critical'),
			COUNT(*) FILTER (WHERE risk_tier = 'high'),
			COUNT(*) FILTER (WHERE risk_tier = 'medium'),
			COUNT(*) FILTER (WHERE risk_tier = 'low'),
			COUNT(*) FILTER (WHERE data_processing = true AND dpa_in_place = false),
			COUNT(*) FILTER (WHERE next_assessment_date < CURRENT_DATE)
		FROM vendors
		WHERE organization_id = $1 AND status = 'active' AND deleted_at IS NULL`, orgID,
	).Scan(&total, &critical, &high, &medium, &low, &withoutDPA, &assessmentOverdue)

	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_vendors":       total,
		"critical_risk":       critical,
		"high_risk":           high,
		"medium_risk":         medium,
		"low_risk":            low,
		"without_dpa":         withoutDPA,
		"assessment_overdue":  assessmentOverdue,
	}, nil
}
