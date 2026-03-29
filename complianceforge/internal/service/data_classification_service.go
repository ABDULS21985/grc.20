package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// ============================================================
// DataClassificationService
// ============================================================

// DataClassificationService implements business logic for data classification
// levels and data categories.
type DataClassificationService struct {
	pool *pgxpool.Pool
}

// NewDataClassificationService creates a new DataClassificationService.
func NewDataClassificationService(pool *pgxpool.Pool) *DataClassificationService {
	return &DataClassificationService{pool: pool}
}

// ============================================================
// DATA TYPES — Classifications
// ============================================================

// DataClassification represents a data sensitivity classification level.
type DataClassification struct {
	ID                      uuid.UUID `json:"id"`
	OrganizationID          uuid.UUID `json:"organization_id"`
	Name                    string    `json:"name"`
	Level                   int       `json:"level"`
	Description             string    `json:"description"`
	HandlingRequirements    string    `json:"handling_requirements"`
	EncryptionRequired      bool      `json:"encryption_required"`
	AccessRestrictionReqd   bool      `json:"access_restriction_required"`
	DataMaskingRequired     bool      `json:"data_masking_required"`
	RetentionPolicy         string    `json:"retention_policy"`
	DisposalMethod          string    `json:"disposal_method"`
	ColorHex                string    `json:"color_hex"`
	IsSystem                bool      `json:"is_system"`
	SortOrder               int       `json:"sort_order"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
}

// DataCategory represents a category of data (e.g., "Health Data", "Email Address").
type DataCategory struct {
	ID                    uuid.UUID `json:"id"`
	OrganizationID        uuid.UUID `json:"organization_id"`
	Name                  string    `json:"name"`
	CategoryType          string    `json:"category_type"`
	GDPRSpecialCategory   bool      `json:"gdpr_special_category"`
	GDPRArticle9Basis     string    `json:"gdpr_article_9_basis"`
	Description           string    `json:"description"`
	Examples              []string  `json:"examples"`
	ClassificationID      *uuid.UUID `json:"classification_id"`
	RetentionPeriodMonths *int       `json:"retention_period_months"`
	IsSystem              bool      `json:"is_system"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

// ============================================================
// REQUEST TYPES — Classifications
// ============================================================

// CreateClassificationRequest is the request body for creating a classification.
type CreateClassificationRequest struct {
	Name                    string `json:"name"`
	Level                   int    `json:"level"`
	Description             string `json:"description"`
	HandlingRequirements    string `json:"handling_requirements"`
	EncryptionRequired      bool   `json:"encryption_required"`
	AccessRestrictionReqd   bool   `json:"access_restriction_required"`
	DataMaskingRequired     bool   `json:"data_masking_required"`
	RetentionPolicy         string `json:"retention_policy"`
	DisposalMethod          string `json:"disposal_method"`
	ColorHex                string `json:"color_hex"`
}

// UpdateClassificationRequest is the request body for updating a classification.
type UpdateClassificationRequest struct {
	Name                    *string `json:"name"`
	Level                   *int    `json:"level"`
	Description             *string `json:"description"`
	HandlingRequirements    *string `json:"handling_requirements"`
	EncryptionRequired      *bool   `json:"encryption_required"`
	AccessRestrictionReqd   *bool   `json:"access_restriction_required"`
	DataMaskingRequired     *bool   `json:"data_masking_required"`
	RetentionPolicy         *string `json:"retention_policy"`
	DisposalMethod          *string `json:"disposal_method"`
	ColorHex                *string `json:"color_hex"`
}

// ============================================================
// REQUEST TYPES — Categories
// ============================================================

// CategoryFilter holds filter parameters for listing data categories.
type CategoryFilter struct {
	CategoryType string `json:"category_type"`
	SpecialOnly  bool   `json:"special_only"`
	Search       string `json:"search"`
}

// CreateCategoryRequest is the request body for creating a data category.
type CreateCategoryRequest struct {
	Name                  string     `json:"name"`
	CategoryType          string     `json:"category_type"`
	GDPRSpecialCategory   bool       `json:"gdpr_special_category"`
	GDPRArticle9Basis     string     `json:"gdpr_article_9_basis"`
	Description           string     `json:"description"`
	Examples              []string   `json:"examples"`
	ClassificationID      *uuid.UUID `json:"classification_id"`
	RetentionPeriodMonths *int       `json:"retention_period_months"`
}

// UpdateCategoryRequest is the request body for updating a data category.
type UpdateCategoryRequest struct {
	Name                  *string    `json:"name"`
	CategoryType          *string    `json:"category_type"`
	GDPRSpecialCategory   *bool      `json:"gdpr_special_category"`
	GDPRArticle9Basis     *string    `json:"gdpr_article_9_basis"`
	Description           *string    `json:"description"`
	Examples              []string   `json:"examples"`
	ClassificationID      *uuid.UUID `json:"classification_id"`
	RetentionPeriodMonths *int       `json:"retention_period_months"`
}

// ============================================================
// CLASSIFICATION CRUD
// ============================================================

// setOrgRLS sets the RLS context for the current transaction.
func (s *DataClassificationService) setOrgRLS(ctx context.Context, orgID uuid.UUID) error {
	_, err := s.pool.Exec(ctx, "SELECT set_config('app.current_org', $1, true)", orgID.String())
	return err
}

// ListClassifications returns all data classification levels for an organization,
// ordered by level ascending.
func (s *DataClassificationService) ListClassifications(ctx context.Context, orgID uuid.UUID) ([]DataClassification, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_org', $1, true)", orgID.String()); err != nil {
		return nil, fmt.Errorf("failed to set RLS context: %w", err)
	}

	rows, err := tx.Query(ctx, `
		SELECT id, organization_id, name, level,
			COALESCE(description, ''), COALESCE(handling_requirements, ''),
			encryption_required, access_restriction_required, data_masking_required,
			COALESCE(retention_policy, ''), COALESCE(disposal_method, ''),
			COALESCE(color_hex, ''), is_system, sort_order,
			created_at, updated_at
		FROM data_classifications
		WHERE organization_id = $1
		ORDER BY level ASC`, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list classifications: %w", err)
	}
	defer rows.Close()

	var results []DataClassification
	for rows.Next() {
		var c DataClassification
		if err := rows.Scan(
			&c.ID, &c.OrganizationID, &c.Name, &c.Level,
			&c.Description, &c.HandlingRequirements,
			&c.EncryptionRequired, &c.AccessRestrictionReqd, &c.DataMaskingRequired,
			&c.RetentionPolicy, &c.DisposalMethod,
			&c.ColorHex, &c.IsSystem, &c.SortOrder,
			&c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan classification: %w", err)
		}
		results = append(results, c)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}

	return results, nil
}

// CreateClassification creates a new data classification level.
func (s *DataClassificationService) CreateClassification(ctx context.Context, orgID uuid.UUID, req CreateClassificationRequest) (*DataClassification, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("classification name is required")
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_org', $1, true)", orgID.String()); err != nil {
		return nil, fmt.Errorf("failed to set RLS context: %w", err)
	}

	c := &DataClassification{}
	err = tx.QueryRow(ctx, `
		INSERT INTO data_classifications (
			organization_id, name, level, description, handling_requirements,
			encryption_required, access_restriction_required, data_masking_required,
			retention_policy, disposal_method, color_hex, is_system, sort_order
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8,
			$9, $10, $11, false, $3
		)
		RETURNING id, organization_id, name, level,
			COALESCE(description, ''), COALESCE(handling_requirements, ''),
			encryption_required, access_restriction_required, data_masking_required,
			COALESCE(retention_policy, ''), COALESCE(disposal_method, ''),
			COALESCE(color_hex, ''), is_system, sort_order,
			created_at, updated_at`,
		orgID, req.Name, req.Level, req.Description, req.HandlingRequirements,
		req.EncryptionRequired, req.AccessRestrictionReqd, req.DataMaskingRequired,
		req.RetentionPolicy, req.DisposalMethod, req.ColorHex,
	).Scan(
		&c.ID, &c.OrganizationID, &c.Name, &c.Level,
		&c.Description, &c.HandlingRequirements,
		&c.EncryptionRequired, &c.AccessRestrictionReqd, &c.DataMaskingRequired,
		&c.RetentionPolicy, &c.DisposalMethod,
		&c.ColorHex, &c.IsSystem, &c.SortOrder,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create classification: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}

	log.Info().
		Str("org_id", orgID.String()).
		Str("name", c.Name).
		Int("level", c.Level).
		Msg("Data classification created")

	return c, nil
}

// UpdateClassification updates an existing data classification.
func (s *DataClassificationService) UpdateClassification(ctx context.Context, orgID, classID uuid.UUID, req UpdateClassificationRequest) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_org', $1, true)", orgID.String()); err != nil {
		return fmt.Errorf("failed to set RLS context: %w", err)
	}

	// Prevent modifying system classifications
	var isSystem bool
	err = tx.QueryRow(ctx, `
		SELECT is_system FROM data_classifications
		WHERE id = $1 AND organization_id = $2`, classID, orgID,
	).Scan(&isSystem)
	if err != nil {
		return fmt.Errorf("classification not found: %w", err)
	}
	if isSystem {
		return fmt.Errorf("system classifications cannot be modified")
	}

	ct, err := tx.Exec(ctx, `
		UPDATE data_classifications SET
			name = COALESCE($3, name),
			level = COALESCE($4, level),
			description = COALESCE($5, description),
			handling_requirements = COALESCE($6, handling_requirements),
			encryption_required = COALESCE($7, encryption_required),
			access_restriction_required = COALESCE($8, access_restriction_required),
			data_masking_required = COALESCE($9, data_masking_required),
			retention_policy = COALESCE($10, retention_policy),
			disposal_method = COALESCE($11, disposal_method),
			color_hex = COALESCE($12, color_hex)
		WHERE id = $1 AND organization_id = $2`,
		classID, orgID,
		req.Name, req.Level, req.Description, req.HandlingRequirements,
		req.EncryptionRequired, req.AccessRestrictionReqd, req.DataMaskingRequired,
		req.RetentionPolicy, req.DisposalMethod, req.ColorHex,
	)
	if err != nil {
		return fmt.Errorf("failed to update classification: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("classification not found")
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	log.Info().
		Str("org_id", orgID.String()).
		Str("classification_id", classID.String()).
		Msg("Data classification updated")

	return nil
}

// DeleteClassification deletes a data classification level.
// System classifications cannot be deleted.
func (s *DataClassificationService) DeleteClassification(ctx context.Context, orgID, classID uuid.UUID) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_org', $1, true)", orgID.String()); err != nil {
		return fmt.Errorf("failed to set RLS context: %w", err)
	}

	var isSystem bool
	err = tx.QueryRow(ctx, `
		SELECT is_system FROM data_classifications
		WHERE id = $1 AND organization_id = $2`, classID, orgID,
	).Scan(&isSystem)
	if err != nil {
		return fmt.Errorf("classification not found: %w", err)
	}
	if isSystem {
		return fmt.Errorf("system classifications cannot be deleted")
	}

	// Nullify references in data_categories
	_, err = tx.Exec(ctx, `
		UPDATE data_categories SET classification_id = NULL
		WHERE classification_id = $1 AND organization_id = $2`,
		classID, orgID,
	)
	if err != nil {
		return fmt.Errorf("failed to unlink categories: %w", err)
	}

	ct, err := tx.Exec(ctx, `
		DELETE FROM data_classifications
		WHERE id = $1 AND organization_id = $2 AND is_system = false`,
		classID, orgID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete classification: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("classification not found or is a system classification")
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	log.Info().
		Str("org_id", orgID.String()).
		Str("classification_id", classID.String()).
		Msg("Data classification deleted")

	return nil
}

// ============================================================
// CATEGORY CRUD
// ============================================================

// ListDataCategories returns data categories for an organization, with optional filters.
func (s *DataClassificationService) ListDataCategories(ctx context.Context, orgID uuid.UUID, filter CategoryFilter) ([]DataCategory, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_org', $1, true)", orgID.String()); err != nil {
		return nil, fmt.Errorf("failed to set RLS context: %w", err)
	}

	query := `
		SELECT id, organization_id, name, category_type::TEXT,
			gdpr_special_category, COALESCE(gdpr_article_9_basis, ''),
			COALESCE(description, ''), COALESCE(examples, '{}'),
			classification_id, retention_period_months,
			is_system, created_at, updated_at
		FROM data_categories
		WHERE organization_id = $1`

	args := []interface{}{orgID}
	argIdx := 2

	if filter.CategoryType != "" {
		query += fmt.Sprintf(" AND category_type = $%d::data_category_type", argIdx)
		args = append(args, filter.CategoryType)
		argIdx++
	}
	if filter.SpecialOnly {
		query += " AND gdpr_special_category = true"
	}
	if filter.Search != "" {
		query += fmt.Sprintf(" AND (name ILIKE '%%' || $%d || '%%' OR description ILIKE '%%' || $%d || '%%')", argIdx, argIdx)
		args = append(args, filter.Search)
		argIdx++
	}

	query += " ORDER BY category_type, name ASC"

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list data categories: %w", err)
	}
	defer rows.Close()

	var results []DataCategory
	for rows.Next() {
		var c DataCategory
		if err := rows.Scan(
			&c.ID, &c.OrganizationID, &c.Name, &c.CategoryType,
			&c.GDPRSpecialCategory, &c.GDPRArticle9Basis,
			&c.Description, &c.Examples,
			&c.ClassificationID, &c.RetentionPeriodMonths,
			&c.IsSystem, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan data category: %w", err)
		}
		results = append(results, c)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}

	return results, nil
}

// CreateDataCategory creates a new data category.
func (s *DataClassificationService) CreateDataCategory(ctx context.Context, orgID uuid.UUID, req CreateCategoryRequest) (*DataCategory, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("category name is required")
	}
	if req.CategoryType == "" {
		req.CategoryType = "personal_data"
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_org', $1, true)", orgID.String()); err != nil {
		return nil, fmt.Errorf("failed to set RLS context: %w", err)
	}

	c := &DataCategory{}
	err = tx.QueryRow(ctx, `
		INSERT INTO data_categories (
			organization_id, name, category_type, gdpr_special_category,
			gdpr_article_9_basis, description, examples,
			classification_id, retention_period_months, is_system
		) VALUES (
			$1, $2, $3::data_category_type, $4,
			$5, $6, $7,
			$8, $9, false
		)
		RETURNING id, organization_id, name, category_type::TEXT,
			gdpr_special_category, COALESCE(gdpr_article_9_basis, ''),
			COALESCE(description, ''), COALESCE(examples, '{}'),
			classification_id, retention_period_months,
			is_system, created_at, updated_at`,
		orgID, req.Name, req.CategoryType, req.GDPRSpecialCategory,
		req.GDPRArticle9Basis, req.Description, req.Examples,
		req.ClassificationID, req.RetentionPeriodMonths,
	).Scan(
		&c.ID, &c.OrganizationID, &c.Name, &c.CategoryType,
		&c.GDPRSpecialCategory, &c.GDPRArticle9Basis,
		&c.Description, &c.Examples,
		&c.ClassificationID, &c.RetentionPeriodMonths,
		&c.IsSystem, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create data category: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}

	log.Info().
		Str("org_id", orgID.String()).
		Str("name", c.Name).
		Str("type", c.CategoryType).
		Msg("Data category created")

	return c, nil
}

// UpdateDataCategory updates an existing data category.
func (s *DataClassificationService) UpdateDataCategory(ctx context.Context, orgID, catID uuid.UUID, req UpdateCategoryRequest) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_org', $1, true)", orgID.String()); err != nil {
		return fmt.Errorf("failed to set RLS context: %w", err)
	}

	ct, err := tx.Exec(ctx, `
		UPDATE data_categories SET
			name = COALESCE($3, name),
			category_type = COALESCE($4::data_category_type, category_type),
			gdpr_special_category = COALESCE($5, gdpr_special_category),
			gdpr_article_9_basis = COALESCE($6, gdpr_article_9_basis),
			description = COALESCE($7, description),
			examples = COALESCE($8, examples),
			classification_id = COALESCE($9, classification_id),
			retention_period_months = COALESCE($10, retention_period_months)
		WHERE id = $1 AND organization_id = $2`,
		catID, orgID,
		req.Name, req.CategoryType, req.GDPRSpecialCategory,
		req.GDPRArticle9Basis, req.Description, req.Examples,
		req.ClassificationID, req.RetentionPeriodMonths,
	)
	if err != nil {
		return fmt.Errorf("failed to update data category: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("data category not found")
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	log.Info().
		Str("org_id", orgID.String()).
		Str("category_id", catID.String()).
		Msg("Data category updated")

	return nil
}

// DeleteDataCategory deletes a data category. System categories cannot be deleted.
func (s *DataClassificationService) DeleteDataCategory(ctx context.Context, orgID, catID uuid.UUID) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_org', $1, true)", orgID.String()); err != nil {
		return fmt.Errorf("failed to set RLS context: %w", err)
	}

	var isSystem bool
	err = tx.QueryRow(ctx, `
		SELECT is_system FROM data_categories
		WHERE id = $1 AND organization_id = $2`, catID, orgID,
	).Scan(&isSystem)
	if err != nil {
		return fmt.Errorf("data category not found: %w", err)
	}
	if isSystem {
		return fmt.Errorf("system data categories cannot be deleted")
	}

	ct, err := tx.Exec(ctx, `
		DELETE FROM data_categories
		WHERE id = $1 AND organization_id = $2 AND is_system = false`,
		catID, orgID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete data category: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("data category not found or is a system category")
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	log.Info().
		Str("org_id", orgID.String()).
		Str("category_id", catID.String()).
		Msg("Data category deleted")

	return nil
}
