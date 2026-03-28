package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/complianceforge/platform/internal/models"
)

type PolicyRepo struct {
	pool *pgxpool.Pool
}

func NewPolicyRepo(pool *pgxpool.Pool) *PolicyRepo {
	return &PolicyRepo{pool: pool}
}

func (r *PolicyRepo) Create(ctx context.Context, tx pgx.Tx, policy *models.Policy) error {
	var ref string
	if err := tx.QueryRow(ctx, "SELECT generate_policy_ref($1)", policy.OrganizationID).Scan(&ref); err != nil {
		return fmt.Errorf("failed to generate policy ref: %w", err)
	}
	policy.PolicyRef = ref

	query := `
		INSERT INTO policies (
			organization_id, policy_ref, title, category_id, status, classification,
			owner_user_id, author_user_id, approver_user_id, review_frequency_months,
			applies_to_all, is_mandatory, requires_attestation, attestation_frequency_months,
			tags, metadata
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
		RETURNING id, created_at, updated_at`

	return tx.QueryRow(ctx, query,
		policy.OrganizationID, policy.PolicyRef, policy.Title, policy.CategoryID,
		policy.Status, policy.Classification, policy.OwnerUserID, policy.AuthorUserID,
		policy.ApproverUserID, policy.ReviewFrequencyMonths, policy.AppliesToAll,
		policy.IsMandatory, policy.RequiresAttestation, policy.AttestationFrequencyMonths,
		policy.Tags, policy.Metadata,
	).Scan(&policy.ID, &policy.CreatedAt, &policy.UpdatedAt)
}

func (r *PolicyRepo) List(ctx context.Context, orgID uuid.UUID, params models.PaginationParams) ([]models.Policy, int64, error) {
	var total int64
	if err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM policies WHERE organization_id = $1 AND deleted_at IS NULL", orgID,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT p.id, p.organization_id, p.policy_ref, p.title, p.category_id,
			p.status, p.classification, p.owner_user_id, p.author_user_id,
			p.approver_user_id, p.current_version, p.review_frequency_months,
			p.last_review_date, p.next_review_date, p.review_status,
			p.applies_to_all, p.effective_date, p.is_mandatory,
			p.requires_attestation, p.tags, p.created_at, p.updated_at
		FROM policies p
		WHERE p.organization_id = $1 AND p.deleted_at IS NULL
		ORDER BY p.%s %s
		LIMIT $2 OFFSET $3`,
		sanitizePolicySortField(params.SortBy), sanitizeSortDir(params.SortDir))

	rows, err := r.pool.Query(ctx, query, orgID, params.PageSize, params.Offset())
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var policies []models.Policy
	for rows.Next() {
		var p models.Policy
		if err := rows.Scan(
			&p.ID, &p.OrganizationID, &p.PolicyRef, &p.Title, &p.CategoryID,
			&p.Status, &p.Classification, &p.OwnerUserID, &p.AuthorUserID,
			&p.ApproverUserID, &p.CurrentVersion, &p.ReviewFrequencyMonths,
			&p.LastReviewDate, &p.NextReviewDate, &p.ReviewStatus,
			&p.AppliesToAll, &p.EffectiveDate, &p.IsMandatory,
			&p.RequiresAttestation, &p.Tags, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		policies = append(policies, p)
	}
	return policies, total, nil
}

func (r *PolicyRepo) GetByID(ctx context.Context, orgID, policyID uuid.UUID) (*models.Policy, error) {
	query := `
		SELECT id, organization_id, policy_ref, title, category_id, status, classification,
			owner_user_id, author_user_id, approver_user_id, current_version,
			current_version_id, review_frequency_months, last_review_date, next_review_date,
			review_status, applies_to_all, linked_framework_ids, linked_control_ids,
			linked_risk_ids, parent_policy_id, effective_date, expiry_date, tags,
			is_mandatory, requires_attestation, attestation_frequency_months,
			metadata, created_at, updated_at
		FROM policies
		WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL`

	var p models.Policy
	err := r.pool.QueryRow(ctx, query, policyID, orgID).Scan(
		&p.ID, &p.OrganizationID, &p.PolicyRef, &p.Title, &p.CategoryID,
		&p.Status, &p.Classification, &p.OwnerUserID, &p.AuthorUserID,
		&p.ApproverUserID, &p.CurrentVersion, &p.CurrentVersionID,
		&p.ReviewFrequencyMonths, &p.LastReviewDate, &p.NextReviewDate,
		&p.ReviewStatus, &p.AppliesToAll, &p.LinkedFrameworkIDs,
		&p.LinkedControlIDs, &p.LinkedRiskIDs, &p.ParentPolicyID,
		&p.EffectiveDate, &p.ExpiryDate, &p.Tags, &p.IsMandatory,
		&p.RequiresAttestation, &p.AttestationFrequencyMonths,
		&p.Metadata, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("policy not found: %w", err)
	}
	return &p, nil
}

func (r *PolicyRepo) CreateVersion(ctx context.Context, tx pgx.Tx, ver *models.PolicyVersion) error {
	query := `
		INSERT INTO policy_versions (
			policy_id, organization_id, version_number, version_label, title,
			content_html, content_text, summary, change_description, change_type,
			language, word_count, status, created_by, metadata
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		RETURNING id, created_at, updated_at`

	return tx.QueryRow(ctx, query,
		ver.PolicyID, ver.OrganizationID, ver.VersionNumber, ver.VersionLabel,
		ver.Title, ver.ContentHTML, ver.ContentText, ver.Summary,
		ver.ChangeDescription, ver.ChangeType, ver.Language, ver.WordCount,
		ver.Status, ver.CreatedBy, ver.Metadata,
	).Scan(&ver.ID, &ver.CreatedAt, &ver.UpdatedAt)
}

func (r *PolicyRepo) UpdateStatus(ctx context.Context, orgID, policyID uuid.UUID, status models.PolicyStatus) error {
	_, err := r.pool.Exec(ctx,
		"UPDATE policies SET status = $1 WHERE id = $2 AND organization_id = $3 AND deleted_at IS NULL",
		status, policyID, orgID,
	)
	return err
}

func (r *PolicyRepo) GetAttestationStats(ctx context.Context, orgID uuid.UUID) (map[string]interface{}, error) {
	var total, attested, pending, overdue int64

	err := r.pool.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE status = 'attested'),
			COUNT(*) FILTER (WHERE status = 'pending'),
			COUNT(*) FILTER (WHERE status = 'overdue')
		FROM policy_attestations
		WHERE organization_id = $1`, orgID,
	).Scan(&total, &attested, &pending, &overdue)

	if err != nil {
		return nil, err
	}

	rate := float64(0)
	if total > 0 {
		rate = float64(attested) / float64(total) * 100
	}

	return map[string]interface{}{
		"total":           total,
		"attested":        attested,
		"pending":         pending,
		"overdue":         overdue,
		"completion_rate": rate,
	}, nil
}

func sanitizePolicySortField(field string) string {
	allowed := map[string]bool{
		"created_at": true, "updated_at": true, "title": true,
		"policy_ref": true, "status": true, "next_review_date": true,
	}
	if allowed[field] {
		return field
	}
	return "created_at"
}
