package service

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// ============================================================
// MARKETPLACE SERVICE
// Manages the Control Library Marketplace including publisher
// registration, package publishing, discovery, installation,
// and community reviews.
// ============================================================

// Maximum allowed package_data size: 10 MB.
const maxPackageDataSize = 10 * 1024 * 1024

// MarketplaceService handles marketplace business logic.
type MarketplaceService struct {
	pool *pgxpool.Pool
}

// NewMarketplaceService creates a new MarketplaceService.
func NewMarketplaceService(pool *pgxpool.Pool) *MarketplaceService {
	return &MarketplaceService{pool: pool}
}

// Pool returns the underlying connection pool for advanced queries.
func (s *MarketplaceService) Pool() *pgxpool.Pool {
	return s.pool
}

// ============================================================
// DATA TYPES — Publishers
// ============================================================

// Publisher represents a marketplace publisher.
type Publisher struct {
	ID               uuid.UUID  `json:"id"`
	OrganizationID   uuid.UUID  `json:"organization_id"`
	PublisherName    string     `json:"publisher_name"`
	PublisherSlug    string     `json:"publisher_slug"`
	Description      string     `json:"description"`
	Website          string     `json:"website"`
	LogoURL          string     `json:"logo_url"`
	IsVerified       bool       `json:"is_verified"`
	VerificationDate *time.Time `json:"verification_date,omitempty"`
	IsOfficial       bool       `json:"is_official"`
	TotalPackages    int        `json:"total_packages"`
	TotalDownloads   int        `json:"total_downloads"`
	RatingAvg        float64    `json:"rating_avg"`
	RatingCount      int        `json:"rating_count"`
	ContactEmail     string     `json:"contact_email"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// CreatePublisherReq is the request payload for creating a publisher.
type CreatePublisherReq struct {
	PublisherName string `json:"publisher_name"`
	PublisherSlug string `json:"publisher_slug"`
	Description   string `json:"description"`
	Website       string `json:"website"`
	LogoURL       string `json:"logo_url"`
	ContactEmail  string `json:"contact_email"`
}

// ============================================================
// DATA TYPES — Packages
// ============================================================

// Package represents a marketplace package.
type Package struct {
	ID                   uuid.UUID   `json:"id"`
	PublisherID          uuid.UUID   `json:"publisher_id"`
	PackageSlug          string      `json:"package_slug"`
	Name                 string      `json:"name"`
	Description          string      `json:"description"`
	LongDescription      string      `json:"long_description,omitempty"`
	PackageType          string      `json:"package_type"`
	Category             string      `json:"category"`
	ApplicableFrameworks []string    `json:"applicable_frameworks"`
	ApplicableRegions    []string    `json:"applicable_regions"`
	ApplicableIndustries []string    `json:"applicable_industries"`
	Tags                 []string    `json:"tags"`
	CurrentVersion       string      `json:"current_version"`
	MinPlatformVersion   string      `json:"min_platform_version"`
	PricingModel         string      `json:"pricing_model"`
	PriceEUR             float64     `json:"price_eur"`
	DownloadCount        int         `json:"download_count"`
	InstallCount         int         `json:"install_count"`
	RatingAvg            float64     `json:"rating_avg"`
	RatingCount          int         `json:"rating_count"`
	Featured             bool        `json:"featured"`
	ContentsSummary      json.RawMessage `json:"contents_summary"`
	Status               string      `json:"status"`
	PublishedAt          *time.Time  `json:"published_at,omitempty"`
	DeprecatedAt         *time.Time  `json:"deprecated_at,omitempty"`
	DeprecationMessage   string      `json:"deprecation_message,omitempty"`
	License              string      `json:"license"`
	CreatedAt            time.Time   `json:"created_at"`
	UpdatedAt            time.Time   `json:"updated_at"`
	// Joined fields
	PublisherName string `json:"publisher_name,omitempty"`
	PublisherSlug string `json:"publisher_slug,omitempty"`
	IsVerified    bool   `json:"is_verified,omitempty"`
}

// CreatePackageReq is the request payload for creating a package.
type CreatePackageReq struct {
	PackageSlug          string   `json:"package_slug"`
	Name                 string   `json:"name"`
	Description          string   `json:"description"`
	LongDescription      string   `json:"long_description"`
	PackageType          string   `json:"package_type"`
	Category             string   `json:"category"`
	ApplicableFrameworks []string `json:"applicable_frameworks"`
	ApplicableRegions    []string `json:"applicable_regions"`
	ApplicableIndustries []string `json:"applicable_industries"`
	Tags                 []string `json:"tags"`
	MinPlatformVersion   string   `json:"min_platform_version"`
	PricingModel         string   `json:"pricing_model"`
	PriceEUR             float64  `json:"price_eur"`
	License              string   `json:"license"`
}

// PackageDetail includes the package plus versions, publisher, and stats.
type PackageDetail struct {
	Package  Package          `json:"package"`
	Versions []PackageVersion `json:"versions"`
	Stats    PackageStats     `json:"stats"`
}

// PackageStats holds computed statistics for a package.
type PackageStats struct {
	TotalDownloads    int     `json:"total_downloads"`
	TotalInstalls     int     `json:"total_installs"`
	AverageRating     float64 `json:"average_rating"`
	TotalReviews      int     `json:"total_reviews"`
	VersionCount      int     `json:"version_count"`
	LatestVersion     string  `json:"latest_version"`
	FirstPublished    *time.Time `json:"first_published,omitempty"`
}

// PackageFilters holds filters for package search.
type PackageFilters struct {
	PackageType string   `json:"package_type"`
	Category    string   `json:"category"`
	Frameworks  []string `json:"frameworks"`
	Regions     []string `json:"regions"`
	Industries  []string `json:"industries"`
	Tags        []string `json:"tags"`
	PricingModel string  `json:"pricing_model"`
	SortBy      string   `json:"sort_by"`   // relevance, downloads, rating, newest
	Page        int      `json:"page"`
	PageSize    int      `json:"page_size"`
}

// ============================================================
// DATA TYPES — Versions
// ============================================================

// PackageVersion represents a specific release of a package.
type PackageVersion struct {
	ID              uuid.UUID       `json:"id"`
	PackageID       uuid.UUID       `json:"package_id"`
	Version         string          `json:"version"`
	ReleaseNotes    string          `json:"release_notes"`
	PackageData     json.RawMessage `json:"package_data,omitempty"`
	PackageHash     string          `json:"package_hash"`
	FileSizeBytes   int64           `json:"file_size_bytes"`
	IsBreakingChange bool           `json:"is_breaking_change"`
	MigrationNotes  string          `json:"migration_notes,omitempty"`
	PublishedAt     time.Time       `json:"published_at"`
	CreatedAt       time.Time       `json:"created_at"`
}

// PublishVersionReq is the request payload for publishing a new version.
type PublishVersionReq struct {
	Version         string          `json:"version"`
	ReleaseNotes    string          `json:"release_notes"`
	PackageData     json.RawMessage `json:"package_data"`
	IsBreakingChange bool           `json:"is_breaking_change"`
	MigrationNotes  string          `json:"migration_notes"`
}

// ============================================================
// DATA TYPES — Installations
// ============================================================

// Installation represents a package installation for an organisation.
type Installation struct {
	ID               uuid.UUID       `json:"id"`
	OrganizationID   uuid.UUID       `json:"organization_id"`
	PackageID        uuid.UUID       `json:"package_id"`
	VersionID        uuid.UUID       `json:"version_id"`
	InstalledVersion string          `json:"installed_version"`
	Status           string          `json:"status"`
	InstalledAt      time.Time       `json:"installed_at"`
	InstalledBy      uuid.UUID       `json:"installed_by"`
	UpdatedAt        time.Time       `json:"updated_at"`
	UninstalledAt    *time.Time      `json:"uninstalled_at,omitempty"`
	Configuration    json.RawMessage `json:"configuration"`
	ImportSummary    json.RawMessage `json:"import_summary"`
	CreatedAt        time.Time       `json:"created_at"`
	// Joined fields
	PackageName    string `json:"package_name,omitempty"`
	PackageSlug    string `json:"package_slug,omitempty"`
	PublisherName  string `json:"publisher_name,omitempty"`
	CurrentVersion string `json:"current_version,omitempty"`
}

// InstallReq is the request payload for installing a package.
type InstallReq struct {
	PackageID     uuid.UUID       `json:"package_id"`
	VersionID     uuid.UUID       `json:"version_id"`
	Configuration json.RawMessage `json:"configuration"`
}

// ============================================================
// DATA TYPES — Reviews
// ============================================================

// Review represents a user review for a package.
type Review struct {
	ID                uuid.UUID `json:"id"`
	PackageID         uuid.UUID `json:"package_id"`
	OrganizationID    uuid.UUID `json:"organization_id"`
	UserID            uuid.UUID `json:"user_id"`
	Rating            int       `json:"rating"`
	Title             string    `json:"title"`
	ReviewText        string    `json:"review_text"`
	HelpfulCount      int       `json:"helpful_count"`
	IsVerifiedInstall bool      `json:"is_verified_install"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// ReviewReq is the request payload for submitting a review.
type ReviewReq struct {
	Rating     int    `json:"rating"`
	Title      string `json:"title"`
	ReviewText string `json:"review_text"`
}

// ============================================================
// PUBLISHING — Publisher Management
// ============================================================

// CreatePublisher registers a new marketplace publisher for an organisation.
func (s *MarketplaceService) CreatePublisher(ctx context.Context, orgID uuid.UUID, req CreatePublisherReq) (*Publisher, error) {
	if req.PublisherName == "" || req.PublisherSlug == "" {
		return nil, fmt.Errorf("publisher_name and publisher_slug are required")
	}

	// Normalise slug
	slug := strings.ToLower(strings.TrimSpace(req.PublisherSlug))

	var pub Publisher
	err := s.pool.QueryRow(ctx, `
		INSERT INTO marketplace_publishers (
			organization_id, publisher_name, publisher_slug, description,
			website, logo_url, contact_email
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, organization_id, publisher_name, publisher_slug, description,
			website, logo_url, is_verified, verification_date, is_official,
			total_packages, total_downloads, rating_avg, rating_count,
			contact_email, created_at, updated_at`,
		orgID, req.PublisherName, slug, req.Description,
		req.Website, req.LogoURL, req.ContactEmail,
	).Scan(
		&pub.ID, &pub.OrganizationID, &pub.PublisherName, &pub.PublisherSlug,
		&pub.Description, &pub.Website, &pub.LogoURL, &pub.IsVerified,
		&pub.VerificationDate, &pub.IsOfficial, &pub.TotalPackages,
		&pub.TotalDownloads, &pub.RatingAvg, &pub.RatingCount,
		&pub.ContactEmail, &pub.CreatedAt, &pub.UpdatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return nil, fmt.Errorf("publisher slug '%s' is already taken", slug)
		}
		return nil, fmt.Errorf("failed to create publisher: %w", err)
	}

	log.Info().Str("publisher_id", pub.ID.String()).Str("slug", slug).Msg("marketplace publisher created")
	return &pub, nil
}

// GetPublisherByOrgID returns the publisher profile for an organisation.
func (s *MarketplaceService) GetPublisherByOrgID(ctx context.Context, orgID uuid.UUID) (*Publisher, error) {
	var pub Publisher
	err := s.pool.QueryRow(ctx, `
		SELECT id, organization_id, publisher_name, publisher_slug, description,
			website, logo_url, is_verified, verification_date, is_official,
			total_packages, total_downloads, rating_avg, rating_count,
			contact_email, created_at, updated_at
		FROM marketplace_publishers WHERE organization_id = $1`, orgID,
	).Scan(
		&pub.ID, &pub.OrganizationID, &pub.PublisherName, &pub.PublisherSlug,
		&pub.Description, &pub.Website, &pub.LogoURL, &pub.IsVerified,
		&pub.VerificationDate, &pub.IsOfficial, &pub.TotalPackages,
		&pub.TotalDownloads, &pub.RatingAvg, &pub.RatingCount,
		&pub.ContactEmail, &pub.CreatedAt, &pub.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("publisher not found: %w", err)
	}
	return &pub, nil
}

// VerifyPublisher marks a publisher as verified (admin only).
func (s *MarketplaceService) VerifyPublisher(ctx context.Context, publisherID uuid.UUID) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE marketplace_publishers
		SET is_verified = true, verification_date = CURRENT_DATE
		WHERE id = $1`, publisherID)
	if err != nil {
		return fmt.Errorf("failed to verify publisher: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("publisher not found")
	}
	return nil
}

// GetPublisherStats returns aggregate stats for a publisher's packages.
func (s *MarketplaceService) GetPublisherStats(ctx context.Context, publisherID uuid.UUID) (map[string]interface{}, error) {
	var totalPackages, totalDownloads, totalInstalls, totalReviews int
	var avgRating float64
	err := s.pool.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COALESCE(SUM(download_count), 0),
			COALESCE(SUM(install_count), 0),
			COALESCE(SUM(rating_count), 0),
			COALESCE(AVG(NULLIF(rating_avg, 0)), 0)
		FROM marketplace_packages
		WHERE publisher_id = $1 AND status = 'published'`, publisherID,
	).Scan(&totalPackages, &totalDownloads, &totalInstalls, &totalReviews, &avgRating)
	if err != nil {
		return nil, fmt.Errorf("failed to load publisher stats: %w", err)
	}

	return map[string]interface{}{
		"total_packages":  totalPackages,
		"total_downloads": totalDownloads,
		"total_installs":  totalInstalls,
		"total_reviews":   totalReviews,
		"average_rating":  avgRating,
	}, nil
}

// ============================================================
// PUBLISHING — Package Management
// ============================================================

// CreatePackage creates a new marketplace package under a publisher.
func (s *MarketplaceService) CreatePackage(ctx context.Context, publisherID uuid.UUID, req CreatePackageReq) (*Package, error) {
	if req.Name == "" || req.PackageSlug == "" || req.PackageType == "" {
		return nil, fmt.Errorf("name, package_slug, and package_type are required")
	}

	slug := strings.ToLower(strings.TrimSpace(req.PackageSlug))
	license := req.License
	if license == "" {
		license = "CC-BY-4.0"
	}
	pricingModel := req.PricingModel
	if pricingModel == "" {
		pricingModel = "free"
	}

	var pkg Package
	err := s.pool.QueryRow(ctx, `
		INSERT INTO marketplace_packages (
			publisher_id, package_slug, name, description, long_description,
			package_type, category, applicable_frameworks, applicable_regions,
			applicable_industries, tags, min_platform_version, pricing_model,
			price_eur, license
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id, publisher_id, package_slug, name, description, long_description,
			package_type, category, applicable_frameworks, applicable_regions,
			applicable_industries, tags, current_version, min_platform_version,
			pricing_model, price_eur, download_count, install_count,
			rating_avg, rating_count, featured, contents_summary, status,
			published_at, deprecated_at, deprecation_message, license,
			created_at, updated_at`,
		publisherID, slug, req.Name, req.Description, req.LongDescription,
		req.PackageType, req.Category, req.ApplicableFrameworks, req.ApplicableRegions,
		req.ApplicableIndustries, req.Tags, req.MinPlatformVersion, pricingModel,
		req.PriceEUR, license,
	).Scan(
		&pkg.ID, &pkg.PublisherID, &pkg.PackageSlug, &pkg.Name,
		&pkg.Description, &pkg.LongDescription, &pkg.PackageType,
		&pkg.Category, &pkg.ApplicableFrameworks, &pkg.ApplicableRegions,
		&pkg.ApplicableIndustries, &pkg.Tags, &pkg.CurrentVersion,
		&pkg.MinPlatformVersion, &pkg.PricingModel, &pkg.PriceEUR,
		&pkg.DownloadCount, &pkg.InstallCount, &pkg.RatingAvg,
		&pkg.RatingCount, &pkg.Featured, &pkg.ContentsSummary,
		&pkg.Status, &pkg.PublishedAt, &pkg.DeprecatedAt,
		&pkg.DeprecationMessage, &pkg.License, &pkg.CreatedAt, &pkg.UpdatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return nil, fmt.Errorf("package slug '%s' already exists for this publisher", slug)
		}
		return nil, fmt.Errorf("failed to create package: %w", err)
	}

	// Increment publisher package count
	s.pool.Exec(ctx, `UPDATE marketplace_publishers SET total_packages = total_packages + 1 WHERE id = $1`, publisherID)

	log.Info().Str("package_id", pkg.ID.String()).Str("slug", slug).Msg("marketplace package created")
	return &pkg, nil
}

// PublishVersion publishes a new version for a package. It validates the
// package_data schema, computes a SHA-256 integrity hash, enforces size
// limits, and updates the package's current_version.
func (s *MarketplaceService) PublishVersion(ctx context.Context, packageID uuid.UUID, req PublishVersionReq) (*PackageVersion, error) {
	if req.Version == "" {
		return nil, fmt.Errorf("version is required")
	}
	if len(req.PackageData) == 0 {
		return nil, fmt.Errorf("package_data is required")
	}

	// Enforce 10 MB size limit
	dataBytes := []byte(req.PackageData)
	if len(dataBytes) > maxPackageDataSize {
		return nil, fmt.Errorf("package_data exceeds maximum size of 10 MB (%d bytes)", len(dataBytes))
	}

	// Validate package_data is valid JSON with a schema_version field
	var dataMap map[string]interface{}
	if err := json.Unmarshal(dataBytes, &dataMap); err != nil {
		return nil, fmt.Errorf("package_data must be valid JSON: %w", err)
	}
	if _, ok := dataMap["schema_version"]; !ok {
		return nil, fmt.Errorf("package_data must contain a 'schema_version' field")
	}
	if _, ok := dataMap["controls"]; !ok {
		if _, ok2 := dataMap["policies"]; !ok2 {
			if _, ok3 := dataMap["templates"]; !ok3 {
				return nil, fmt.Errorf("package_data must contain 'controls', 'policies', or 'templates'")
			}
		}
	}

	// Calculate SHA-256 hash
	hash := sha256.Sum256(dataBytes)
	hashStr := fmt.Sprintf("%x", hash)

	fileSizeBytes := int64(len(dataBytes))

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Insert version
	var ver PackageVersion
	err = tx.QueryRow(ctx, `
		INSERT INTO marketplace_package_versions (
			package_id, version, release_notes, package_data,
			package_hash, file_size_bytes, is_breaking_change, migration_notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, package_id, version, release_notes, package_hash,
			file_size_bytes, is_breaking_change, migration_notes, published_at, created_at`,
		packageID, req.Version, req.ReleaseNotes, req.PackageData,
		hashStr, fileSizeBytes, req.IsBreakingChange, req.MigrationNotes,
	).Scan(
		&ver.ID, &ver.PackageID, &ver.Version, &ver.ReleaseNotes,
		&ver.PackageHash, &ver.FileSizeBytes, &ver.IsBreakingChange,
		&ver.MigrationNotes, &ver.PublishedAt, &ver.CreatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return nil, fmt.Errorf("version '%s' already exists for this package", req.Version)
		}
		return nil, fmt.Errorf("failed to insert version: %w", err)
	}

	// Build contents_summary from package_data
	contentsSummary := buildContentsSummary(dataMap)
	summaryJSON, _ := json.Marshal(contentsSummary)

	// Update package: current_version, status, published_at, contents_summary
	_, err = tx.Exec(ctx, `
		UPDATE marketplace_packages
		SET current_version = $1,
			status = 'published',
			published_at = COALESCE(published_at, now()),
			contents_summary = $2
		WHERE id = $3`,
		req.Version, summaryJSON, packageID)
	if err != nil {
		return nil, fmt.Errorf("failed to update package: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}

	log.Info().Str("package_id", packageID.String()).Str("version", req.Version).Str("hash", hashStr).Msg("package version published")
	return &ver, nil
}

// buildContentsSummary extracts summary statistics from package data.
func buildContentsSummary(data map[string]interface{}) map[string]interface{} {
	summary := map[string]interface{}{}

	if controls, ok := data["controls"].([]interface{}); ok {
		summary["total_controls"] = len(controls)
		categories := map[string]int{}
		for _, c := range controls {
			if ctrl, ok := c.(map[string]interface{}); ok {
				if cat, ok := ctrl["category"].(string); ok {
					categories[cat]++
				}
			}
		}
		catList := []map[string]interface{}{}
		for name, count := range categories {
			catList = append(catList, map[string]interface{}{"name": name, "count": count})
		}
		summary["control_categories"] = catList
	}

	if policies, ok := data["policies"].([]interface{}); ok {
		summary["total_policies"] = len(policies)
	}
	if templates, ok := data["templates"].([]interface{}); ok {
		summary["total_templates"] = len(templates)
	}

	return summary
}

// DeprecatePackage marks a package as deprecated with a message.
func (s *MarketplaceService) DeprecatePackage(ctx context.Context, packageID uuid.UUID, message string) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE marketplace_packages
		SET status = 'deprecated', deprecated_at = now(), deprecation_message = $1
		WHERE id = $2 AND status = 'published'`,
		message, packageID)
	if err != nil {
		return fmt.Errorf("failed to deprecate package: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("package not found or not in published state")
	}
	log.Info().Str("package_id", packageID.String()).Msg("package deprecated")
	return nil
}

// FeaturePackage toggles the featured flag for a package (admin only).
func (s *MarketplaceService) FeaturePackage(ctx context.Context, packageID uuid.UUID, featured bool) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE marketplace_packages SET featured = $1 WHERE id = $2`,
		featured, packageID)
	if err != nil {
		return fmt.Errorf("failed to feature package: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("package not found")
	}
	return nil
}

// ============================================================
// DISCOVERY — Package Search & Browse
// ============================================================

// SearchPackages performs full-text search with faceted filtering.
func (s *MarketplaceService) SearchPackages(ctx context.Context, query string, filters PackageFilters) ([]Package, int64, error) {
	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.PageSize < 1 || filters.PageSize > 100 {
		filters.PageSize = 20
	}

	// Build the WHERE clause dynamically
	conditions := []string{"p.status = 'published'"}
	args := []interface{}{}
	argIdx := 1

	if query != "" {
		conditions = append(conditions, fmt.Sprintf(
			"to_tsvector('english', coalesce(p.name, '') || ' ' || coalesce(p.description, '')) @@ plainto_tsquery('english', $%d)", argIdx))
		args = append(args, query)
		argIdx++
	}

	if filters.PackageType != "" {
		conditions = append(conditions, fmt.Sprintf("p.package_type = $%d", argIdx))
		args = append(args, filters.PackageType)
		argIdx++
	}
	if filters.Category != "" {
		conditions = append(conditions, fmt.Sprintf("p.category = $%d", argIdx))
		args = append(args, filters.Category)
		argIdx++
	}
	if filters.PricingModel != "" {
		conditions = append(conditions, fmt.Sprintf("p.pricing_model = $%d", argIdx))
		args = append(args, filters.PricingModel)
		argIdx++
	}
	if len(filters.Frameworks) > 0 {
		conditions = append(conditions, fmt.Sprintf("p.applicable_frameworks && $%d", argIdx))
		args = append(args, filters.Frameworks)
		argIdx++
	}
	if len(filters.Regions) > 0 {
		conditions = append(conditions, fmt.Sprintf("p.applicable_regions && $%d", argIdx))
		args = append(args, filters.Regions)
		argIdx++
	}
	if len(filters.Industries) > 0 {
		conditions = append(conditions, fmt.Sprintf("p.applicable_industries && $%d", argIdx))
		args = append(args, filters.Industries)
		argIdx++
	}
	if len(filters.Tags) > 0 {
		conditions = append(conditions, fmt.Sprintf("p.tags && $%d", argIdx))
		args = append(args, filters.Tags)
		argIdx++
	}

	whereClause := strings.Join(conditions, " AND ")

	// Determine sort order
	orderBy := "p.download_count DESC, p.rating_avg DESC"
	switch filters.SortBy {
	case "downloads":
		orderBy = "p.download_count DESC"
	case "rating":
		orderBy = "p.rating_avg DESC, p.rating_count DESC"
	case "newest":
		orderBy = "p.published_at DESC NULLS LAST"
	case "relevance":
		if query != "" {
			orderBy = fmt.Sprintf(
				"ts_rank(to_tsvector('english', coalesce(p.name, '') || ' ' || coalesce(p.description, '')), plainto_tsquery('english', $%d)) DESC",
				1) // query is always $1 when present
		}
	}

	// Count total matching rows
	var total int64
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM marketplace_packages p WHERE %s`, whereClause)
	err := s.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count query failed: %w", err)
	}

	// Fetch page
	offset := (filters.Page - 1) * filters.PageSize
	dataQuery := fmt.Sprintf(`
		SELECT p.id, p.publisher_id, p.package_slug, p.name, p.description,
			p.package_type, p.category, p.applicable_frameworks, p.applicable_regions,
			p.applicable_industries, p.tags, p.current_version, p.pricing_model,
			p.price_eur, p.download_count, p.install_count, p.rating_avg,
			p.rating_count, p.featured, p.contents_summary, p.status,
			p.published_at, p.license, p.created_at, p.updated_at,
			pub.publisher_name, pub.publisher_slug, pub.is_verified
		FROM marketplace_packages p
		JOIN marketplace_publishers pub ON p.publisher_id = pub.id
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d`,
		whereClause, orderBy, argIdx, argIdx+1)
	args = append(args, filters.PageSize, offset)

	rows, err := s.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("search query failed: %w", err)
	}
	defer rows.Close()

	var packages []Package
	for rows.Next() {
		var pkg Package
		if err := rows.Scan(
			&pkg.ID, &pkg.PublisherID, &pkg.PackageSlug, &pkg.Name, &pkg.Description,
			&pkg.PackageType, &pkg.Category, &pkg.ApplicableFrameworks, &pkg.ApplicableRegions,
			&pkg.ApplicableIndustries, &pkg.Tags, &pkg.CurrentVersion, &pkg.PricingModel,
			&pkg.PriceEUR, &pkg.DownloadCount, &pkg.InstallCount, &pkg.RatingAvg,
			&pkg.RatingCount, &pkg.Featured, &pkg.ContentsSummary, &pkg.Status,
			&pkg.PublishedAt, &pkg.License, &pkg.CreatedAt, &pkg.UpdatedAt,
			&pkg.PublisherName, &pkg.PublisherSlug, &pkg.IsVerified,
		); err != nil {
			return nil, 0, fmt.Errorf("scan failed: %w", err)
		}
		packages = append(packages, pkg)
	}

	return packages, total, nil
}

// GetFeaturedPackages returns all packages marked as featured.
func (s *MarketplaceService) GetFeaturedPackages(ctx context.Context) ([]Package, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT p.id, p.publisher_id, p.package_slug, p.name, p.description,
			p.package_type, p.category, p.applicable_frameworks, p.applicable_regions,
			p.applicable_industries, p.tags, p.current_version, p.pricing_model,
			p.price_eur, p.download_count, p.install_count, p.rating_avg,
			p.rating_count, p.featured, p.contents_summary, p.status,
			p.published_at, p.license, p.created_at, p.updated_at,
			pub.publisher_name, pub.publisher_slug, pub.is_verified
		FROM marketplace_packages p
		JOIN marketplace_publishers pub ON p.publisher_id = pub.id
		WHERE p.featured = true AND p.status = 'published'
		ORDER BY p.download_count DESC
		LIMIT 12`)
	if err != nil {
		return nil, fmt.Errorf("failed to query featured packages: %w", err)
	}
	defer rows.Close()

	return scanPackageRows(rows)
}

// GetPackageDetail returns a package with its versions and stats.
func (s *MarketplaceService) GetPackageDetail(ctx context.Context, publisherSlug, packageSlug string) (*PackageDetail, error) {
	var pkg Package
	err := s.pool.QueryRow(ctx, `
		SELECT p.id, p.publisher_id, p.package_slug, p.name, p.description,
			p.long_description, p.package_type, p.category,
			p.applicable_frameworks, p.applicable_regions, p.applicable_industries,
			p.tags, p.current_version, p.min_platform_version, p.pricing_model,
			p.price_eur, p.download_count, p.install_count, p.rating_avg,
			p.rating_count, p.featured, p.contents_summary, p.status,
			p.published_at, p.deprecated_at, p.deprecation_message,
			p.license, p.created_at, p.updated_at,
			pub.publisher_name, pub.publisher_slug, pub.is_verified
		FROM marketplace_packages p
		JOIN marketplace_publishers pub ON p.publisher_id = pub.id
		WHERE pub.publisher_slug = $1 AND p.package_slug = $2`,
		publisherSlug, packageSlug,
	).Scan(
		&pkg.ID, &pkg.PublisherID, &pkg.PackageSlug, &pkg.Name, &pkg.Description,
		&pkg.LongDescription, &pkg.PackageType, &pkg.Category,
		&pkg.ApplicableFrameworks, &pkg.ApplicableRegions, &pkg.ApplicableIndustries,
		&pkg.Tags, &pkg.CurrentVersion, &pkg.MinPlatformVersion, &pkg.PricingModel,
		&pkg.PriceEUR, &pkg.DownloadCount, &pkg.InstallCount, &pkg.RatingAvg,
		&pkg.RatingCount, &pkg.Featured, &pkg.ContentsSummary, &pkg.Status,
		&pkg.PublishedAt, &pkg.DeprecatedAt, &pkg.DeprecationMessage,
		&pkg.License, &pkg.CreatedAt, &pkg.UpdatedAt,
		&pkg.PublisherName, &pkg.PublisherSlug, &pkg.IsVerified,
	)
	if err != nil {
		return nil, fmt.Errorf("package not found: %w", err)
	}

	// Load versions (without full package_data for listing)
	verRows, err := s.pool.Query(ctx, `
		SELECT id, package_id, version, release_notes, package_hash,
			file_size_bytes, is_breaking_change, migration_notes, published_at, created_at
		FROM marketplace_package_versions
		WHERE package_id = $1
		ORDER BY published_at DESC`, pkg.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load versions: %w", err)
	}
	defer verRows.Close()

	var versions []PackageVersion
	for verRows.Next() {
		var v PackageVersion
		if err := verRows.Scan(
			&v.ID, &v.PackageID, &v.Version, &v.ReleaseNotes, &v.PackageHash,
			&v.FileSizeBytes, &v.IsBreakingChange, &v.MigrationNotes,
			&v.PublishedAt, &v.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan version: %w", err)
		}
		versions = append(versions, v)
	}

	stats := PackageStats{
		TotalDownloads: pkg.DownloadCount,
		TotalInstalls:  pkg.InstallCount,
		AverageRating:  pkg.RatingAvg,
		TotalReviews:   pkg.RatingCount,
		VersionCount:   len(versions),
		LatestVersion:  pkg.CurrentVersion,
		FirstPublished: pkg.PublishedAt,
	}

	return &PackageDetail{
		Package:  pkg,
		Versions: versions,
		Stats:    stats,
	}, nil
}

// GetPackagesByFramework returns packages applicable to a given framework code.
func (s *MarketplaceService) GetPackagesByFramework(ctx context.Context, frameworkCode string) ([]Package, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT p.id, p.publisher_id, p.package_slug, p.name, p.description,
			p.package_type, p.category, p.applicable_frameworks, p.applicable_regions,
			p.applicable_industries, p.tags, p.current_version, p.pricing_model,
			p.price_eur, p.download_count, p.install_count, p.rating_avg,
			p.rating_count, p.featured, p.contents_summary, p.status,
			p.published_at, p.license, p.created_at, p.updated_at,
			pub.publisher_name, pub.publisher_slug, pub.is_verified
		FROM marketplace_packages p
		JOIN marketplace_publishers pub ON p.publisher_id = pub.id
		WHERE p.status = 'published' AND $1 = ANY(p.applicable_frameworks)
		ORDER BY p.download_count DESC`, frameworkCode)
	if err != nil {
		return nil, fmt.Errorf("failed to query packages by framework: %w", err)
	}
	defer rows.Close()

	return scanPackageRows(rows)
}

// GetPackageVersions returns all versions for a package.
func (s *MarketplaceService) GetPackageVersions(ctx context.Context, publisherSlug, packageSlug string) ([]PackageVersion, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT v.id, v.package_id, v.version, v.release_notes, v.package_hash,
			v.file_size_bytes, v.is_breaking_change, v.migration_notes,
			v.published_at, v.created_at
		FROM marketplace_package_versions v
		JOIN marketplace_packages p ON v.package_id = p.id
		JOIN marketplace_publishers pub ON p.publisher_id = pub.id
		WHERE pub.publisher_slug = $1 AND p.package_slug = $2
		ORDER BY v.published_at DESC`,
		publisherSlug, packageSlug)
	if err != nil {
		return nil, fmt.Errorf("failed to query versions: %w", err)
	}
	defer rows.Close()

	var versions []PackageVersion
	for rows.Next() {
		var v PackageVersion
		if err := rows.Scan(
			&v.ID, &v.PackageID, &v.Version, &v.ReleaseNotes, &v.PackageHash,
			&v.FileSizeBytes, &v.IsBreakingChange, &v.MigrationNotes,
			&v.PublishedAt, &v.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan version: %w", err)
		}
		versions = append(versions, v)
	}

	return versions, nil
}

// scanPackageRows scans rows into a []Package slice (shared helper).
func scanPackageRows(rows pgx.Rows) ([]Package, error) {
	var packages []Package
	for rows.Next() {
		var pkg Package
		if err := rows.Scan(
			&pkg.ID, &pkg.PublisherID, &pkg.PackageSlug, &pkg.Name, &pkg.Description,
			&pkg.PackageType, &pkg.Category, &pkg.ApplicableFrameworks, &pkg.ApplicableRegions,
			&pkg.ApplicableIndustries, &pkg.Tags, &pkg.CurrentVersion, &pkg.PricingModel,
			&pkg.PriceEUR, &pkg.DownloadCount, &pkg.InstallCount, &pkg.RatingAvg,
			&pkg.RatingCount, &pkg.Featured, &pkg.ContentsSummary, &pkg.Status,
			&pkg.PublishedAt, &pkg.License, &pkg.CreatedAt, &pkg.UpdatedAt,
			&pkg.PublisherName, &pkg.PublisherSlug, &pkg.IsVerified,
		); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		packages = append(packages, pkg)
	}
	return packages, nil
}

// ============================================================
// INSTALLATION — Install / Uninstall / Update
// ============================================================

// InstallPackage installs a marketplace package for an organisation.
func (s *MarketplaceService) InstallPackage(ctx context.Context, orgID uuid.UUID, req InstallReq) (*Installation, error) {
	if req.PackageID == uuid.Nil || req.VersionID == uuid.Nil {
		return nil, fmt.Errorf("package_id and version_id are required")
	}

	// Verify the package and version exist
	var packageStatus string
	var versionStr string
	err := s.pool.QueryRow(ctx, `
		SELECT p.status, v.version
		FROM marketplace_packages p
		JOIN marketplace_package_versions v ON v.package_id = p.id
		WHERE p.id = $1 AND v.id = $2`,
		req.PackageID, req.VersionID,
	).Scan(&packageStatus, &versionStr)
	if err != nil {
		return nil, fmt.Errorf("package or version not found: %w", err)
	}
	if packageStatus != "published" && packageStatus != "deprecated" {
		return nil, fmt.Errorf("package is not available for installation (status: %s)", packageStatus)
	}

	// Download package_data for validation
	var pkgData json.RawMessage
	var pkgHash string
	err = s.pool.QueryRow(ctx, `
		SELECT package_data, package_hash
		FROM marketplace_package_versions WHERE id = $1`, req.VersionID,
	).Scan(&pkgData, &pkgHash)
	if err != nil {
		return nil, fmt.Errorf("failed to download package data: %w", err)
	}

	// Verify integrity
	computedHash := sha256.Sum256([]byte(pkgData))
	computedHashStr := fmt.Sprintf("%x", computedHash)
	if computedHashStr != pkgHash {
		return nil, fmt.Errorf("package integrity check failed: hash mismatch")
	}

	// Build import summary from package_data
	var dataMap map[string]interface{}
	json.Unmarshal([]byte(pkgData), &dataMap)
	importSummary := buildContentsSummary(dataMap)
	importSummary["installed_at"] = time.Now().UTC().Format(time.RFC3339)
	importSummary["status"] = "success"
	importJSON, _ := json.Marshal(importSummary)

	cfg := req.Configuration
	if len(cfg) == 0 {
		cfg = json.RawMessage(`{}`)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Insert installation record
	var inst Installation
	err = tx.QueryRow(ctx, `
		INSERT INTO marketplace_installations (
			organization_id, package_id, version_id, installed_version,
			status, installed_by, configuration, import_summary
		) VALUES ($1, $2, $3, $4, 'installed', $5, $6, $7)
		ON CONFLICT (organization_id, package_id) DO UPDATE SET
			version_id = EXCLUDED.version_id,
			installed_version = EXCLUDED.installed_version,
			status = 'installed',
			installed_by = EXCLUDED.installed_by,
			configuration = EXCLUDED.configuration,
			import_summary = EXCLUDED.import_summary,
			uninstalled_at = NULL
		RETURNING id, organization_id, package_id, version_id, installed_version,
			status, installed_at, installed_by, updated_at, uninstalled_at,
			configuration, import_summary, created_at`,
		orgID, req.PackageID, req.VersionID, versionStr,
		uuid.Nil, cfg, importJSON,
	).Scan(
		&inst.ID, &inst.OrganizationID, &inst.PackageID, &inst.VersionID,
		&inst.InstalledVersion, &inst.Status, &inst.InstalledAt, &inst.InstalledBy,
		&inst.UpdatedAt, &inst.UninstalledAt, &inst.Configuration,
		&inst.ImportSummary, &inst.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to record installation: %w", err)
	}

	// Increment package download and install counts
	_, err = tx.Exec(ctx, `
		UPDATE marketplace_packages
		SET download_count = download_count + 1, install_count = install_count + 1
		WHERE id = $1`, req.PackageID)
	if err != nil {
		return nil, fmt.Errorf("failed to update counts: %w", err)
	}

	// Increment publisher download count
	_, err = tx.Exec(ctx, `
		UPDATE marketplace_publishers
		SET total_downloads = total_downloads + 1
		WHERE id = (SELECT publisher_id FROM marketplace_packages WHERE id = $1)`, req.PackageID)
	if err != nil {
		return nil, fmt.Errorf("failed to update publisher counts: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}

	log.Info().
		Str("org_id", orgID.String()).
		Str("package_id", req.PackageID.String()).
		Str("version", versionStr).
		Msg("package installed")

	return &inst, nil
}

// UninstallPackage marks a package installation as uninstalled.
func (s *MarketplaceService) UninstallPackage(ctx context.Context, orgID, packageID uuid.UUID) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE marketplace_installations
		SET status = 'uninstalled', uninstalled_at = now()
		WHERE organization_id = $1 AND package_id = $2 AND status != 'uninstalled'`,
		orgID, packageID)
	if err != nil {
		return fmt.Errorf("failed to uninstall: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("installation not found or already uninstalled")
	}

	// Decrement install count
	s.pool.Exec(ctx, `
		UPDATE marketplace_packages
		SET install_count = GREATEST(install_count - 1, 0)
		WHERE id = $1`, packageID)

	log.Info().Str("org_id", orgID.String()).Str("package_id", packageID.String()).Msg("package uninstalled")
	return nil
}

// UpdatePackage updates an installation to the latest version.
func (s *MarketplaceService) UpdatePackage(ctx context.Context, orgID, installationID uuid.UUID) (*Installation, error) {
	// Get current installation
	var packageID uuid.UUID
	var currentVersion string
	err := s.pool.QueryRow(ctx, `
		SELECT package_id, installed_version
		FROM marketplace_installations
		WHERE id = $1 AND organization_id = $2 AND status = 'installed'`,
		installationID, orgID,
	).Scan(&packageID, &currentVersion)
	if err != nil {
		return nil, fmt.Errorf("installation not found: %w", err)
	}

	// Get the latest version
	var latestVersionID uuid.UUID
	var latestVersion string
	err = s.pool.QueryRow(ctx, `
		SELECT id, version
		FROM marketplace_package_versions
		WHERE package_id = $1
		ORDER BY published_at DESC
		LIMIT 1`, packageID,
	).Scan(&latestVersionID, &latestVersion)
	if err != nil {
		return nil, fmt.Errorf("no versions found: %w", err)
	}

	if latestVersion == currentVersion {
		return nil, fmt.Errorf("package is already at the latest version (%s)", currentVersion)
	}

	// Update installation
	var inst Installation
	err = s.pool.QueryRow(ctx, `
		UPDATE marketplace_installations
		SET version_id = $1, installed_version = $2, status = 'installed'
		WHERE id = $3 AND organization_id = $4
		RETURNING id, organization_id, package_id, version_id, installed_version,
			status, installed_at, installed_by, updated_at, uninstalled_at,
			configuration, import_summary, created_at`,
		latestVersionID, latestVersion, installationID, orgID,
	).Scan(
		&inst.ID, &inst.OrganizationID, &inst.PackageID, &inst.VersionID,
		&inst.InstalledVersion, &inst.Status, &inst.InstalledAt, &inst.InstalledBy,
		&inst.UpdatedAt, &inst.UninstalledAt, &inst.Configuration,
		&inst.ImportSummary, &inst.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update installation: %w", err)
	}

	log.Info().
		Str("installation_id", installationID.String()).
		Str("from_version", currentVersion).
		Str("to_version", latestVersion).
		Msg("package updated")

	return &inst, nil
}

// ListInstalled returns all active installations for an organisation.
func (s *MarketplaceService) ListInstalled(ctx context.Context, orgID uuid.UUID) ([]Installation, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT i.id, i.organization_id, i.package_id, i.version_id,
			i.installed_version, i.status, i.installed_at, i.installed_by,
			i.updated_at, i.uninstalled_at, i.configuration, i.import_summary,
			i.created_at,
			p.name, p.package_slug, pub.publisher_name, p.current_version
		FROM marketplace_installations i
		JOIN marketplace_packages p ON i.package_id = p.id
		JOIN marketplace_publishers pub ON p.publisher_id = pub.id
		WHERE i.organization_id = $1 AND i.status != 'uninstalled'
		ORDER BY i.installed_at DESC`, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to query installations: %w", err)
	}
	defer rows.Close()

	var installations []Installation
	for rows.Next() {
		var inst Installation
		if err := rows.Scan(
			&inst.ID, &inst.OrganizationID, &inst.PackageID, &inst.VersionID,
			&inst.InstalledVersion, &inst.Status, &inst.InstalledAt, &inst.InstalledBy,
			&inst.UpdatedAt, &inst.UninstalledAt, &inst.Configuration,
			&inst.ImportSummary, &inst.CreatedAt,
			&inst.PackageName, &inst.PackageSlug, &inst.PublisherName, &inst.CurrentVersion,
		); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		installations = append(installations, inst)
	}

	return installations, nil
}

// ============================================================
// REVIEWS — Submit & List
// ============================================================

// SubmitReview submits or updates a review for a package. The reviewer
// must have the package installed to qualify as a verified install.
func (s *MarketplaceService) SubmitReview(ctx context.Context, orgID, userID, packageID uuid.UUID, req ReviewReq) (*Review, error) {
	if req.Rating < 1 || req.Rating > 5 {
		return nil, fmt.Errorf("rating must be between 1 and 5")
	}
	if req.Title == "" {
		return nil, fmt.Errorf("title is required")
	}

	// Verify that the organisation has installed this package
	var installCount int
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM marketplace_installations
		WHERE organization_id = $1 AND package_id = $2 AND status = 'installed'`,
		orgID, packageID,
	).Scan(&installCount)
	if err != nil || installCount == 0 {
		return nil, fmt.Errorf("you must install this package before reviewing it")
	}

	// Upsert review
	var review Review
	err = s.pool.QueryRow(ctx, `
		INSERT INTO marketplace_reviews (
			package_id, organization_id, user_id, rating, title,
			review_text, is_verified_install
		) VALUES ($1, $2, $3, $4, $5, $6, true)
		ON CONFLICT (package_id, organization_id) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			rating = EXCLUDED.rating,
			title = EXCLUDED.title,
			review_text = EXCLUDED.review_text,
			is_verified_install = EXCLUDED.is_verified_install
		RETURNING id, package_id, organization_id, user_id, rating, title,
			review_text, helpful_count, is_verified_install, created_at, updated_at`,
		packageID, orgID, userID, req.Rating, req.Title, req.ReviewText,
	).Scan(
		&review.ID, &review.PackageID, &review.OrganizationID, &review.UserID,
		&review.Rating, &review.Title, &review.ReviewText, &review.HelpfulCount,
		&review.IsVerifiedInstall, &review.CreatedAt, &review.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to submit review: %w", err)
	}

	// Recalculate package rating
	s.pool.Exec(ctx, `
		UPDATE marketplace_packages SET
			rating_avg = (SELECT COALESCE(AVG(rating), 0) FROM marketplace_reviews WHERE package_id = $1),
			rating_count = (SELECT COUNT(*) FROM marketplace_reviews WHERE package_id = $1)
		WHERE id = $1`, packageID)

	// Recalculate publisher aggregate rating
	s.pool.Exec(ctx, `
		UPDATE marketplace_publishers SET
			rating_avg = sub.avg_rating,
			rating_count = sub.total_count
		FROM (
			SELECT publisher_id,
				COALESCE(AVG(NULLIF(rating_avg, 0)), 0) AS avg_rating,
				COALESCE(SUM(rating_count), 0)::int AS total_count
			FROM marketplace_packages
			WHERE publisher_id = (SELECT publisher_id FROM marketplace_packages WHERE id = $1)
			GROUP BY publisher_id
		) sub
		WHERE marketplace_publishers.id = sub.publisher_id`, packageID)

	log.Info().
		Str("package_id", packageID.String()).
		Str("org_id", orgID.String()).
		Int("rating", req.Rating).
		Msg("review submitted")

	return &review, nil
}

// GetReviews returns paginated reviews for a package.
func (s *MarketplaceService) GetReviews(ctx context.Context, packageID uuid.UUID, page, pageSize int) ([]Review, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	var total int64
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM marketplace_reviews WHERE package_id = $1`, packageID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count failed: %w", err)
	}

	offset := (page - 1) * pageSize
	rows, err := s.pool.Query(ctx, `
		SELECT id, package_id, organization_id, user_id, rating, title,
			review_text, helpful_count, is_verified_install, created_at, updated_at
		FROM marketplace_reviews
		WHERE package_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`, packageID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var reviews []Review
	for rows.Next() {
		var r Review
		if err := rows.Scan(
			&r.ID, &r.PackageID, &r.OrganizationID, &r.UserID, &r.Rating,
			&r.Title, &r.ReviewText, &r.HelpfulCount, &r.IsVerifiedInstall,
			&r.CreatedAt, &r.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan failed: %w", err)
		}
		reviews = append(reviews, r)
	}

	return reviews, total, nil
}
