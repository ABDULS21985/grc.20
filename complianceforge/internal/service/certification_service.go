package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// ============================================================
// CertificationService
// ============================================================

// CertificationService implements business logic for tracking
// professional certifications held by team members, including
// expiry monitoring and certification matrix reporting.
type CertificationService struct {
	pool *pgxpool.Pool
}

// NewCertificationService creates a new CertificationService with the given database pool.
func NewCertificationService(pool *pgxpool.Pool) *CertificationService {
	return &CertificationService{pool: pool}
}

// ============================================================
// DATA TYPES
// ============================================================

// ProfessionalCertification represents a professional certification held by a user.
type ProfessionalCertification struct {
	ID                  uuid.UUID       `json:"id"`
	OrganizationID      uuid.UUID       `json:"organization_id"`
	UserID              uuid.UUID       `json:"user_id"`
	CertificationName   string          `json:"certification_name"`
	IssuingBody         string          `json:"issuing_body"`
	CredentialID        string          `json:"credential_id"`
	CertificationURL    string          `json:"certification_url"`
	DateObtained        time.Time       `json:"date_obtained"`
	ExpiryDate          *time.Time      `json:"expiry_date"`
	RenewalDate         *time.Time      `json:"renewal_date"`
	Status              string          `json:"status"`
	Notes               string          `json:"notes"`
	EvidenceDocumentURL string          `json:"evidence_document_url"`
	Metadata            json.RawMessage `json:"metadata"`
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`
	UserEmail           string          `json:"user_email,omitempty"`
	UserFullName        string          `json:"user_full_name,omitempty"`
	DaysUntilExpiry     *int            `json:"days_until_expiry,omitempty"`
}

// CertificationMatrix shows which certifications each team member holds.
type CertificationMatrix struct {
	CertificationNames []string                     `json:"certification_names"`
	Users              []CertMatrixUser             `json:"users"`
	Cells              [][]CertMatrixCell           `json:"cells"`
	Summary            CertificationMatrixSummary   `json:"summary"`
}

// CertMatrixUser is a row header in the certification matrix.
type CertMatrixUser struct {
	ID       uuid.UUID `json:"id"`
	FullName string    `json:"full_name"`
	Email    string    `json:"email"`
}

// CertMatrixCell holds the status of one user x certification intersection.
type CertMatrixCell struct {
	Held        bool       `json:"held"`
	Status      string     `json:"status"`
	ExpiryDate  *time.Time `json:"expiry_date,omitempty"`
	CredentialID string    `json:"credential_id,omitempty"`
}

// CertificationMatrixSummary provides aggregate data for the matrix.
type CertificationMatrixSummary struct {
	TotalCertifications int            `json:"total_certifications"`
	ActiveCount         int            `json:"active_count"`
	ExpiringCount       int            `json:"expiring_count"`
	ExpiredCount        int            `json:"expired_count"`
	ByCertification     map[string]int `json:"by_certification"`
}

// ============================================================
// REQUEST TYPES
// ============================================================

// AddCertificationReq is the request body for adding a professional certification.
type AddCertificationReq struct {
	UserID              uuid.UUID  `json:"user_id"`
	CertificationName   string     `json:"certification_name"`
	IssuingBody         string     `json:"issuing_body"`
	CredentialID        string     `json:"credential_id"`
	CertificationURL    string     `json:"certification_url"`
	DateObtained        time.Time  `json:"date_obtained"`
	ExpiryDate          *time.Time `json:"expiry_date"`
	RenewalDate         *time.Time `json:"renewal_date"`
	Notes               string     `json:"notes"`
	EvidenceDocumentURL string     `json:"evidence_document_url"`
}

// UpdateCertificationReq is the request body for updating a certification.
type UpdateCertificationReq struct {
	CertificationName   *string    `json:"certification_name"`
	IssuingBody         *string    `json:"issuing_body"`
	CredentialID        *string    `json:"credential_id"`
	CertificationURL    *string    `json:"certification_url"`
	ExpiryDate          *time.Time `json:"expiry_date"`
	RenewalDate         *time.Time `json:"renewal_date"`
	Status              *string    `json:"status"`
	Notes               *string    `json:"notes"`
	EvidenceDocumentURL *string    `json:"evidence_document_url"`
}

// ============================================================
// CERTIFICATION CRUD
// ============================================================

// ListCertifications returns a paginated list of professional certifications.
func (s *CertificationService) ListCertifications(ctx context.Context, orgID uuid.UUID, page, pageSize int) ([]ProfessionalCertification, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var total int64
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM professional_certifications
		WHERE organization_id = $1 AND deleted_at IS NULL`, orgID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count certifications: %w", err)
	}

	rows, err := s.pool.Query(ctx, `
		SELECT
			pc.id, pc.organization_id, pc.user_id,
			pc.certification_name, pc.issuing_body,
			COALESCE(pc.credential_id, ''), COALESCE(pc.certification_url, ''),
			pc.date_obtained, pc.expiry_date, pc.renewal_date,
			pc.status::TEXT, COALESCE(pc.notes, ''),
			COALESCE(pc.evidence_document_url, ''),
			COALESCE(pc.metadata, '{}'::jsonb),
			pc.created_at, pc.updated_at,
			u.email, COALESCE(u.full_name, u.email),
			CASE
				WHEN pc.expiry_date IS NOT NULL
				THEN EXTRACT(DAY FROM pc.expiry_date - now())::INTEGER
				ELSE NULL
			END AS days_until_expiry
		FROM professional_certifications pc
		JOIN users u ON u.id = pc.user_id
		WHERE pc.organization_id = $1 AND pc.deleted_at IS NULL
		ORDER BY pc.expiry_date ASC NULLS LAST, pc.certification_name
		LIMIT $2 OFFSET $3`,
		orgID, pageSize, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list certifications: %w", err)
	}
	defer rows.Close()

	var certs []ProfessionalCertification
	for rows.Next() {
		var c ProfessionalCertification
		if err := rows.Scan(
			&c.ID, &c.OrganizationID, &c.UserID,
			&c.CertificationName, &c.IssuingBody,
			&c.CredentialID, &c.CertificationURL,
			&c.DateObtained, &c.ExpiryDate, &c.RenewalDate,
			&c.Status, &c.Notes,
			&c.EvidenceDocumentURL,
			&c.Metadata,
			&c.CreatedAt, &c.UpdatedAt,
			&c.UserEmail, &c.UserFullName,
			&c.DaysUntilExpiry,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan certification: %w", err)
		}
		certs = append(certs, c)
	}
	return certs, total, nil
}

// AddCertification records a new professional certification.
func (s *CertificationService) AddCertification(ctx context.Context, orgID uuid.UUID, req AddCertificationReq) (*ProfessionalCertification, error) {
	// Determine initial status based on expiry date
	status := "active"
	if req.ExpiryDate != nil {
		now := time.Now()
		daysUntil := int(req.ExpiryDate.Sub(now).Hours() / 24)
		if daysUntil < 0 {
			status = "expired"
		} else if daysUntil <= 90 {
			status = "expiring_soon"
		}
	}

	c := &ProfessionalCertification{}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO professional_certifications (
			organization_id, user_id,
			certification_name, issuing_body, credential_id,
			certification_url, date_obtained, expiry_date, renewal_date,
			status, notes, evidence_document_url
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10::certification_status, $11, $12
		)
		RETURNING id, organization_id, user_id,
			certification_name, issuing_body,
			COALESCE(credential_id, ''), COALESCE(certification_url, ''),
			date_obtained, expiry_date, renewal_date,
			status::TEXT, COALESCE(notes, ''),
			COALESCE(evidence_document_url, ''),
			COALESCE(metadata, '{}'::jsonb),
			created_at, updated_at`,
		orgID, req.UserID,
		req.CertificationName, req.IssuingBody, req.CredentialID,
		req.CertificationURL, req.DateObtained, req.ExpiryDate, req.RenewalDate,
		status, req.Notes, req.EvidenceDocumentURL,
	).Scan(
		&c.ID, &c.OrganizationID, &c.UserID,
		&c.CertificationName, &c.IssuingBody,
		&c.CredentialID, &c.CertificationURL,
		&c.DateObtained, &c.ExpiryDate, &c.RenewalDate,
		&c.Status, &c.Notes,
		&c.EvidenceDocumentURL,
		&c.Metadata,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add certification: %w", err)
	}

	log.Info().
		Str("org_id", orgID.String()).
		Str("user_id", req.UserID.String()).
		Str("cert", req.CertificationName).
		Msg("Professional certification added")

	return c, nil
}

// UpdateCertification updates mutable fields on a professional certification.
func (s *CertificationService) UpdateCertification(ctx context.Context, orgID, certID uuid.UUID, req UpdateCertificationReq) error {
	ct, err := s.pool.Exec(ctx, `
		UPDATE professional_certifications SET
			certification_name = COALESCE($3, certification_name),
			issuing_body = COALESCE($4, issuing_body),
			credential_id = COALESCE($5, credential_id),
			certification_url = COALESCE($6, certification_url),
			expiry_date = COALESCE($7, expiry_date),
			renewal_date = COALESCE($8, renewal_date),
			status = COALESCE($9::certification_status, status),
			notes = COALESCE($10, notes),
			evidence_document_url = COALESCE($11, evidence_document_url)
		WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL`,
		certID, orgID,
		req.CertificationName, req.IssuingBody,
		req.CredentialID, req.CertificationURL,
		req.ExpiryDate, req.RenewalDate,
		req.Status, req.Notes, req.EvidenceDocumentURL,
	)
	if err != nil {
		return fmt.Errorf("failed to update certification: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("certification not found")
	}
	return nil
}

// ============================================================
// EXPIRY MONITORING
// ============================================================

// GetExpiringCertifications returns certifications expiring within the specified number of days.
func (s *CertificationService) GetExpiringCertifications(ctx context.Context, orgID uuid.UUID, withinDays int) ([]ProfessionalCertification, error) {
	if withinDays <= 0 {
		withinDays = 90
	}

	rows, err := s.pool.Query(ctx, `
		SELECT
			pc.id, pc.organization_id, pc.user_id,
			pc.certification_name, pc.issuing_body,
			COALESCE(pc.credential_id, ''), COALESCE(pc.certification_url, ''),
			pc.date_obtained, pc.expiry_date, pc.renewal_date,
			pc.status::TEXT, COALESCE(pc.notes, ''),
			COALESCE(pc.evidence_document_url, ''),
			COALESCE(pc.metadata, '{}'::jsonb),
			pc.created_at, pc.updated_at,
			u.email, COALESCE(u.full_name, u.email),
			EXTRACT(DAY FROM pc.expiry_date - now())::INTEGER AS days_until_expiry
		FROM professional_certifications pc
		JOIN users u ON u.id = pc.user_id
		WHERE pc.organization_id = $1
			AND pc.deleted_at IS NULL
			AND pc.expiry_date IS NOT NULL
			AND pc.expiry_date <= now() + ($2 || ' days')::INTERVAL
			AND pc.status NOT IN ('revoked')
		ORDER BY pc.expiry_date ASC`,
		orgID, fmt.Sprintf("%d", withinDays),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get expiring certifications: %w", err)
	}
	defer rows.Close()

	var certs []ProfessionalCertification
	for rows.Next() {
		var c ProfessionalCertification
		if err := rows.Scan(
			&c.ID, &c.OrganizationID, &c.UserID,
			&c.CertificationName, &c.IssuingBody,
			&c.CredentialID, &c.CertificationURL,
			&c.DateObtained, &c.ExpiryDate, &c.RenewalDate,
			&c.Status, &c.Notes,
			&c.EvidenceDocumentURL,
			&c.Metadata,
			&c.CreatedAt, &c.UpdatedAt,
			&c.UserEmail, &c.UserFullName,
			&c.DaysUntilExpiry,
		); err != nil {
			return nil, fmt.Errorf("failed to scan expiring certification: %w", err)
		}
		certs = append(certs, c)
	}
	return certs, nil
}

// ============================================================
// CERTIFICATION MATRIX
// ============================================================

// GetCertificationMatrix builds a matrix showing which certifications each team member holds.
func (s *CertificationService) GetCertificationMatrix(ctx context.Context, orgID uuid.UUID) (*CertificationMatrix, error) {
	matrix := &CertificationMatrix{
		Summary: CertificationMatrixSummary{
			ByCertification: make(map[string]int),
		},
	}

	// Get distinct certification names
	nameRows, err := s.pool.Query(ctx, `
		SELECT DISTINCT certification_name
		FROM professional_certifications
		WHERE organization_id = $1 AND deleted_at IS NULL
		ORDER BY certification_name`, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query certification names: %w", err)
	}
	defer nameRows.Close()

	for nameRows.Next() {
		var name string
		if err := nameRows.Scan(&name); err != nil {
			return nil, fmt.Errorf("failed to scan cert name: %w", err)
		}
		matrix.CertificationNames = append(matrix.CertificationNames, name)
	}

	if len(matrix.CertificationNames) == 0 {
		return matrix, nil
	}

	// Get users who have certifications
	userRows, err := s.pool.Query(ctx, `
		SELECT DISTINCT u.id, COALESCE(u.full_name, u.email), u.email
		FROM professional_certifications pc
		JOIN users u ON u.id = pc.user_id
		WHERE pc.organization_id = $1 AND pc.deleted_at IS NULL
		ORDER BY COALESCE(u.full_name, u.email)`, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query users for cert matrix: %w", err)
	}
	defer userRows.Close()

	for userRows.Next() {
		var mu CertMatrixUser
		if err := userRows.Scan(&mu.ID, &mu.FullName, &mu.Email); err != nil {
			return nil, fmt.Errorf("failed to scan cert matrix user: %w", err)
		}
		matrix.Users = append(matrix.Users, mu)
	}

	// Build lookup: userID -> certName -> cell
	certLookup := make(map[string]CertMatrixCell)
	allRows, err := s.pool.Query(ctx, `
		SELECT user_id, certification_name, status::TEXT,
			expiry_date, COALESCE(credential_id, '')
		FROM professional_certifications
		WHERE organization_id = $1 AND deleted_at IS NULL`,
		orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query certs for matrix: %w", err)
	}
	defer allRows.Close()

	for allRows.Next() {
		var userID uuid.UUID
		var certName, status, credID string
		var expiry *time.Time
		if err := allRows.Scan(&userID, &certName, &status, &expiry, &credID); err != nil {
			return nil, fmt.Errorf("failed to scan cert matrix cell: %w", err)
		}
		key := userID.String() + ":" + certName
		certLookup[key] = CertMatrixCell{
			Held:         true,
			Status:       status,
			ExpiryDate:   expiry,
			CredentialID: credID,
		}

		// Update summary
		matrix.Summary.TotalCertifications++
		matrix.Summary.ByCertification[certName]++
		switch status {
		case "active":
			matrix.Summary.ActiveCount++
		case "expiring_soon", "pending_renewal":
			matrix.Summary.ExpiringCount++
		case "expired":
			matrix.Summary.ExpiredCount++
		}
	}

	// Build cells grid
	matrix.Cells = make([][]CertMatrixCell, len(matrix.Users))
	for i, u := range matrix.Users {
		matrix.Cells[i] = make([]CertMatrixCell, len(matrix.CertificationNames))
		for j, certName := range matrix.CertificationNames {
			key := u.ID.String() + ":" + certName
			if cell, ok := certLookup[key]; ok {
				matrix.Cells[i][j] = cell
			} else {
				matrix.Cells[i][j] = CertMatrixCell{
					Held:   false,
					Status: "not_held",
				}
			}
		}
	}

	return matrix, nil
}
