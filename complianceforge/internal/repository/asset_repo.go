package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/complianceforge/platform/internal/models"
)

type AssetRepo struct {
	pool *pgxpool.Pool
}

func NewAssetRepo(pool *pgxpool.Pool) *AssetRepo {
	return &AssetRepo{pool: pool}
}

func (r *AssetRepo) Create(ctx context.Context, asset *models.Asset) error {
	// Generate ref
	var nextNum int
	r.pool.QueryRow(ctx,
		"SELECT COALESCE(MAX(CAST(SUBSTRING(asset_ref FROM 5) AS INT)), 0) + 1 FROM assets WHERE organization_id = $1",
		asset.OrganizationID).Scan(&nextNum)
	asset.AssetRef = fmt.Sprintf("AST-%04d", nextNum)

	query := `
		INSERT INTO assets (
			organization_id, asset_ref, name, asset_type, category, description,
			status, criticality, owner_user_id, custodian_user_id, location,
			ip_address, classification, processes_personal_data,
			linked_vendor_id, tags, metadata
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12::INET,$13,$14,$15,$16,$17)
		RETURNING id, created_at, updated_at`

	return r.pool.QueryRow(ctx, query,
		asset.OrganizationID, asset.AssetRef, asset.Name, asset.AssetType,
		asset.Category, asset.Description, asset.Status, asset.Criticality,
		asset.OwnerUserID, asset.CustodianUserID, asset.Location,
		asset.IPAddress, asset.Classification, asset.ProcessesPersonalData,
		asset.LinkedVendorID, asset.Tags, asset.Metadata,
	).Scan(&asset.ID, &asset.CreatedAt, &asset.UpdatedAt)
}

func (r *AssetRepo) List(ctx context.Context, orgID uuid.UUID, params models.PaginationParams) ([]models.Asset, int64, error) {
	var total int64
	r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM assets WHERE organization_id = $1 AND deleted_at IS NULL", orgID).Scan(&total)

	query := `
		SELECT id, organization_id, asset_ref, name, asset_type, category,
			description, status, criticality, owner_user_id, location,
			classification, processes_personal_data, linked_vendor_id,
			tags, created_at, updated_at
		FROM assets
		WHERE organization_id = $1 AND deleted_at IS NULL
		ORDER BY criticality DESC, name ASC
		LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, orgID, params.PageSize, params.Offset())
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var assets []models.Asset
	for rows.Next() {
		var a models.Asset
		if err := rows.Scan(
			&a.ID, &a.OrganizationID, &a.AssetRef, &a.Name, &a.AssetType,
			&a.Category, &a.Description, &a.Status, &a.Criticality,
			&a.OwnerUserID, &a.Location, &a.Classification,
			&a.ProcessesPersonalData, &a.LinkedVendorID,
			&a.Tags, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		assets = append(assets, a)
	}
	return assets, total, nil
}

func (r *AssetRepo) GetByID(ctx context.Context, orgID, assetID uuid.UUID) (*models.Asset, error) {
	query := `
		SELECT id, organization_id, asset_ref, name, asset_type, category,
			description, status, criticality, owner_user_id, custodian_user_id,
			location, ip_address, classification, processes_personal_data,
			linked_vendor_id, tags, metadata, created_at, updated_at
		FROM assets
		WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL`

	var a models.Asset
	err := r.pool.QueryRow(ctx, query, assetID, orgID).Scan(
		&a.ID, &a.OrganizationID, &a.AssetRef, &a.Name, &a.AssetType,
		&a.Category, &a.Description, &a.Status, &a.Criticality,
		&a.OwnerUserID, &a.CustodianUserID, &a.Location, &a.IPAddress,
		&a.Classification, &a.ProcessesPersonalData, &a.LinkedVendorID,
		&a.Tags, &a.Metadata, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("asset not found: %w", err)
	}
	return &a, nil
}

func (r *AssetRepo) GetStats(ctx context.Context, orgID uuid.UUID) (map[string]interface{}, error) {
	var total, critical, withPersonalData int64
	byType := make(map[string]int)

	r.pool.QueryRow(ctx, `
		SELECT COUNT(*),
			COUNT(*) FILTER (WHERE criticality = 'critical'),
			COUNT(*) FILTER (WHERE processes_personal_data = true)
		FROM assets WHERE organization_id = $1 AND deleted_at IS NULL`, orgID,
	).Scan(&total, &critical, &withPersonalData)

	typeRows, _ := r.pool.Query(ctx, `
		SELECT asset_type, COUNT(*) FROM assets
		WHERE organization_id = $1 AND deleted_at IS NULL
		GROUP BY asset_type`, orgID)
	if typeRows != nil {
		defer typeRows.Close()
		for typeRows.Next() {
			var t string
			var c int
			typeRows.Scan(&t, &c)
			byType[t] = c
		}
	}

	return map[string]interface{}{
		"total_assets":      total,
		"critical_assets":   critical,
		"personal_data":     withPersonalData,
		"by_type":           byType,
	}, nil
}
