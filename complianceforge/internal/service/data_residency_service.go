package service

import (
	"context"
	"fmt"
	"math/big"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// ============================================================
// MODEL TYPES
// ============================================================

// DataResidencyRegion represents a configured data residency region
// with its allowed cloud regions, compliance frameworks, and settings.
type DataResidencyRegion struct {
	ID                     uuid.UUID `json:"id"`
	Region                 string    `json:"region"`
	DisplayName            string    `json:"display_name"`
	Description            string    `json:"description"`
	AllowedCountries       []string  `json:"allowed_countries"`
	BlockedCountries       []string  `json:"blocked_countries"`
	AllowedCloudRegions    map[string]interface{} `json:"allowed_cloud_regions"`
	PrimaryCloudRegion     string    `json:"primary_cloud_region"`
	FailoverCloudRegion    string    `json:"failover_cloud_region"`
	ComplianceFrameworks   []string  `json:"compliance_frameworks"`
	LegalBasis             string    `json:"legal_basis"`
	DataProtectionAuth     string    `json:"data_protection_authority"`
	DPAContactURL          string    `json:"dpa_contact_url"`
	GDPRAdequateCountries  []string  `json:"gdpr_adequate_countries"`
	EnforcementMode        string    `json:"enforcement_mode"`
	AllowCrossRegionSearch bool      `json:"allow_cross_region_search"`
	AllowCrossRegionBackup bool      `json:"allow_cross_region_backup"`
	IsActive               bool      `json:"is_active"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

// ResidencyConfig holds the data residency configuration for a specific
// organisation, combining the organisation's region setting with the
// region-level configuration.
type ResidencyConfig struct {
	OrganizationID         uuid.UUID              `json:"organization_id"`
	DataRegion             string                 `json:"data_region"`
	DataResidencyEnforced  bool                   `json:"data_residency_enforced"`
	DisplayName            string                 `json:"display_name"`
	Description            string                 `json:"description"`
	AllowedCountries       []string               `json:"allowed_countries"`
	BlockedCountries       []string               `json:"blocked_countries"`
	AllowedCloudRegions    map[string]interface{} `json:"allowed_cloud_regions"`
	PrimaryCloudRegion     string                 `json:"primary_cloud_region"`
	FailoverCloudRegion    string                 `json:"failover_cloud_region"`
	ComplianceFrameworks   []string               `json:"compliance_frameworks"`
	LegalBasis             string                 `json:"legal_basis"`
	DataProtectionAuth     string                 `json:"data_protection_authority"`
	DPAContactURL          string                 `json:"dpa_contact_url"`
	GDPRAdequateCountries  []string               `json:"gdpr_adequate_countries"`
	EnforcementMode        string                 `json:"enforcement_mode"`
	AllowCrossRegionSearch bool                   `json:"allow_cross_region_search"`
	AllowCrossRegionBackup bool                   `json:"allow_cross_region_backup"`
}

// ResidencyDashboard provides an at-a-glance view of data residency
// status for an organisation.
type ResidencyDashboard struct {
	CurrentRegion        string                 `json:"current_region"`
	RegionDisplayName    string                 `json:"region_display_name"`
	Enforced             bool                   `json:"enforced"`
	EnforcementMode      string                 `json:"enforcement_mode"`
	ComplianceStatus     string                 `json:"compliance_status"`
	ComplianceFrameworks []string               `json:"compliance_frameworks"`
	StorageLocations     []StorageLocation      `json:"storage_locations"`
	RecentBlocked        []ResidencyAuditEntry  `json:"recent_blocked"`
	TotalBlockedCount    int                    `json:"total_blocked_count"`
	LastAuditEntry       *time.Time             `json:"last_audit_entry"`
	AllowedCloudRegions  map[string]interface{} `json:"allowed_cloud_regions"`
}

// StorageLocation represents where a specific data type is stored.
type StorageLocation struct {
	Type             string `json:"type"`
	Label            string `json:"label"`
	Region           string `json:"region"`
	Provider         string `json:"provider"`
	Status           string `json:"status"`
	WithinJurisdiction bool `json:"within_jurisdiction"`
}

// ResidencyAuditEntry represents a single entry in the data residency audit log.
type ResidencyAuditEntry struct {
	ID                 uuid.UUID  `json:"id"`
	OrganizationID     uuid.UUID  `json:"organization_id"`
	Action             string     `json:"action"`
	UserID             *uuid.UUID `json:"user_id,omitempty"`
	UserEmail          string     `json:"user_email,omitempty"`
	IPAddress          string     `json:"ip_address,omitempty"`
	SourceRegion       string     `json:"source_region,omitempty"`
	DestinationRegion  string     `json:"destination_region,omitempty"`
	SourceCountry      string     `json:"source_country,omitempty"`
	DestinationCountry string     `json:"destination_country,omitempty"`
	ResourceType       string     `json:"resource_type,omitempty"`
	ResourceID         *uuid.UUID `json:"resource_id,omitempty"`
	Allowed            bool       `json:"allowed"`
	BlockedReason      string     `json:"blocked_reason,omitempty"`
	VendorID           *uuid.UUID `json:"vendor_id,omitempty"`
	VendorName         string     `json:"vendor_name,omitempty"`
	TransferMechanism  string     `json:"transfer_mechanism,omitempty"`
	Details            map[string]interface{} `json:"details,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
}

// GeoLocation represents a resolved IP address location.
type GeoLocation struct {
	IPAddress   string `json:"ip_address"`
	CountryCode string `json:"country_code"`
	CountryName string `json:"country_name"`
	Region      string `json:"region"`
	City        string `json:"city"`
	IsEU        bool   `json:"is_eu"`
}

// ExportValidationResult holds the outcome of a data export validation check.
type ExportValidationResult struct {
	Allowed            bool     `json:"allowed"`
	SourceRegion       string   `json:"source_region"`
	DestinationRegion  string   `json:"destination_region"`
	Reason             string   `json:"reason"`
	RequiredSafeguards []string `json:"required_safeguards,omitempty"`
	LegalBasis         string   `json:"legal_basis,omitempty"`
}

// TransferValidation holds the outcome of a vendor data transfer validation.
type TransferValidation struct {
	Allowed              bool     `json:"allowed"`
	DestinationCountry   string   `json:"destination_country"`
	IsGDPRAdequate       bool     `json:"is_gdpr_adequate"`
	RequiresAdditional   bool     `json:"requires_additional_safeguards"`
	TransferMechanisms   []string `json:"transfer_mechanisms,omitempty"`
	RequiredSafeguards   []string `json:"required_safeguards,omitempty"`
	Reason               string   `json:"reason"`
	DataProtectionAuth   string   `json:"data_protection_authority,omitempty"`
}

// AuditLogFilter holds filtering parameters for querying the audit log.
type AuditLogFilter struct {
	Action   string     `json:"action"`
	Allowed  *bool      `json:"allowed"`
	DateFrom *time.Time `json:"date_from"`
	DateTo   *time.Time `json:"date_to"`
	Page     int        `json:"page"`
	PageSize int        `json:"page_size"`
}

// ============================================================
// GDPR ADEQUACY DECISIONS
// Reference list of countries deemed adequate under GDPR.
// Updated based on European Commission decisions.
// ============================================================

// GDPRAdequateCountries is the canonical list of countries with an
// EU GDPR adequacy decision (as of 2025). This is used for transfer
// validation when no per-region override is configured.
var GDPRAdequateCountries = []string{
	"AD", // Andorra
	"AR", // Argentina
	"CA", // Canada (commercial organisations)
	"FO", // Faroe Islands
	"GG", // Guernsey
	"IL", // Israel
	"IM", // Isle of Man
	"JP", // Japan
	"JE", // Jersey
	"NZ", // New Zealand
	"CH", // Switzerland
	"UY", // Uruguay
	"KR", // South Korea
	"GB", // United Kingdom
	"US", // United States (EU-US Data Privacy Framework)
}

// EUMemberStates is the list of EU/EEA member state country codes.
var EUMemberStates = []string{
	"AT", "BE", "BG", "HR", "CY", "CZ", "DK", "EE", "FI", "FR",
	"DE", "GR", "HU", "IE", "IT", "LV", "LT", "LU", "MT", "NL",
	"PL", "PT", "RO", "SK", "SI", "ES", "SE",
	// EEA
	"IS", "LI", "NO",
}

// regionToCountryMap maps residency region codes to the primary
// country codes they encompass.
var regionToCountryMap = map[string][]string{
	"eu":      EUMemberStates,
	"uk":      {"GB"},
	"dach":    {"DE", "AT", "CH"},
	"nordics": {"DK", "FI", "IS", "NO", "SE"},
	"france":  {"FR"},
	"global":  {}, // no restriction
}

// ipCountryRanges provides a minimal built-in IP-to-country lookup
// for well-known IP blocks. A production deployment would use a full
// GeoIP database (MaxMind, IP2Location, etc.).
var ipCountryRanges = []struct {
	Network     string
	CountryCode string
	CountryName string
	IsEU        bool
}{
	{"1.0.0.0/8", "AU", "Australia", false},
	{"2.0.0.0/8", "FR", "France", true},
	{"5.0.0.0/8", "DE", "Germany", true},
	{"8.8.0.0/16", "US", "United States", false},
	{"9.0.0.0/8", "US", "United States", false},
	{"10.0.0.0/8", "XX", "Private Network", false},
	{"17.0.0.0/8", "US", "United States", false},
	{"31.0.0.0/8", "NL", "Netherlands", true},
	{"34.0.0.0/8", "US", "United States", false},
	{"35.0.0.0/8", "US", "United States", false},
	{"46.0.0.0/8", "RU", "Russia", false},
	{"51.0.0.0/8", "GB", "United Kingdom", false},
	{"52.0.0.0/8", "US", "United States", false},
	{"77.0.0.0/8", "DE", "Germany", true},
	{"78.0.0.0/8", "GB", "United Kingdom", false},
	{"80.0.0.0/8", "DE", "Germany", true},
	{"81.0.0.0/8", "GB", "United Kingdom", false},
	{"82.0.0.0/8", "FR", "France", true},
	{"83.0.0.0/8", "ES", "Spain", true},
	{"84.0.0.0/8", "DE", "Germany", true},
	{"85.0.0.0/8", "SE", "Sweden", true},
	{"86.0.0.0/8", "FR", "France", true},
	{"87.0.0.0/8", "IE", "Ireland", true},
	{"88.0.0.0/8", "IT", "Italy", true},
	{"89.0.0.0/8", "NL", "Netherlands", true},
	{"90.0.0.0/8", "FR", "France", true},
	{"91.0.0.0/8", "ES", "Spain", true},
	{"92.0.0.0/8", "FR", "France", true},
	{"93.0.0.0/8", "SE", "Sweden", true},
	{"94.0.0.0/8", "FI", "Finland", true},
	{"95.0.0.0/8", "NO", "Norway", true},
	{"100.0.0.0/8", "XX", "Private/CGNAT", false},
	{"104.0.0.0/8", "US", "United States", false},
	{"127.0.0.0/8", "XX", "Loopback", false},
	{"142.0.0.0/8", "CA", "Canada", false},
	{"151.0.0.0/8", "JP", "Japan", false},
	{"157.0.0.0/8", "US", "United States", false},
	{"169.254.0.0/16", "XX", "Link-Local", false},
	{"172.16.0.0/12", "XX", "Private Network", false},
	{"176.0.0.0/8", "DE", "Germany", true},
	{"185.0.0.0/8", "NL", "Netherlands", true},
	{"192.168.0.0/16", "XX", "Private Network", false},
	{"193.0.0.0/8", "NL", "Netherlands", true},
	{"194.0.0.0/8", "SE", "Sweden", true},
	{"195.0.0.0/8", "DE", "Germany", true},
	{"203.0.0.0/8", "AU", "Australia", false},
	{"212.0.0.0/8", "GB", "United Kingdom", false},
	{"213.0.0.0/8", "FR", "France", true},
}

// ============================================================
// SERVICE
// ============================================================

// DataResidencyService provides business logic for multi-region
// deployment and data residency management, including region
// configuration, compliance validation, and audit logging.
type DataResidencyService struct {
	pool *pgxpool.Pool
}

// NewDataResidencyService creates a new DataResidencyService.
func NewDataResidencyService(pool *pgxpool.Pool) *DataResidencyService {
	return &DataResidencyService{pool: pool}
}

// ============================================================
// REGION CONFIG
// ============================================================

// GetRegionConfig returns the data residency configuration for the
// given organisation, combining the org's region setting with the
// region-level configuration from data_residency_configs.
func (s *DataResidencyService) GetRegionConfig(ctx context.Context, orgID uuid.UUID) (*ResidencyConfig, error) {
	var cfg ResidencyConfig
	err := s.pool.QueryRow(ctx, `
		SELECT o.id, o.data_region::text, o.data_residency_enforced,
		       COALESCE(drc.display_name, o.data_region::text),
		       COALESCE(drc.description, ''),
		       COALESCE(drc.allowed_countries, '{}'),
		       COALESCE(drc.blocked_countries, '{}'),
		       COALESCE(drc.allowed_cloud_regions, '{}'),
		       COALESCE(drc.primary_cloud_region, ''),
		       COALESCE(drc.failover_cloud_region, ''),
		       COALESCE(drc.compliance_frameworks, '{}'),
		       COALESCE(drc.legal_basis, ''),
		       COALESCE(drc.data_protection_authority, ''),
		       COALESCE(drc.dpa_contact_url, ''),
		       COALESCE(drc.gdpr_adequate_countries, '{}'),
		       COALESCE(drc.enforcement_mode::text, 'enforce'),
		       COALESCE(drc.allow_cross_region_search, false),
		       COALESCE(drc.allow_cross_region_backup, false)
		FROM organizations o
		LEFT JOIN data_residency_configs drc ON drc.region = o.data_region
		WHERE o.id = $1`, orgID).Scan(
		&cfg.OrganizationID, &cfg.DataRegion, &cfg.DataResidencyEnforced,
		&cfg.DisplayName, &cfg.Description,
		&cfg.AllowedCountries, &cfg.BlockedCountries,
		&cfg.AllowedCloudRegions,
		&cfg.PrimaryCloudRegion, &cfg.FailoverCloudRegion,
		&cfg.ComplianceFrameworks, &cfg.LegalBasis,
		&cfg.DataProtectionAuth, &cfg.DPAContactURL,
		&cfg.GDPRAdequateCountries, &cfg.EnforcementMode,
		&cfg.AllowCrossRegionSearch, &cfg.AllowCrossRegionBackup,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("organization not found")
		}
		return nil, fmt.Errorf("get region config: %w", err)
	}
	return &cfg, nil
}

// ListRegions returns all configured data residency regions.
func (s *DataResidencyService) ListRegions(ctx context.Context) ([]DataResidencyRegion, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, region::text, display_name, COALESCE(description, ''),
		       allowed_countries, blocked_countries,
		       allowed_cloud_regions,
		       COALESCE(primary_cloud_region, ''),
		       COALESCE(failover_cloud_region, ''),
		       compliance_frameworks,
		       COALESCE(legal_basis, ''),
		       COALESCE(data_protection_authority, ''),
		       COALESCE(dpa_contact_url, ''),
		       gdpr_adequate_countries,
		       enforcement_mode::text,
		       allow_cross_region_search,
		       allow_cross_region_backup,
		       is_active,
		       created_at, updated_at
		FROM data_residency_configs
		WHERE is_active = true
		ORDER BY display_name`)
	if err != nil {
		return nil, fmt.Errorf("query regions: %w", err)
	}
	defer rows.Close()

	var regions []DataResidencyRegion
	for rows.Next() {
		var r DataResidencyRegion
		if err := rows.Scan(
			&r.ID, &r.Region, &r.DisplayName, &r.Description,
			&r.AllowedCountries, &r.BlockedCountries,
			&r.AllowedCloudRegions,
			&r.PrimaryCloudRegion, &r.FailoverCloudRegion,
			&r.ComplianceFrameworks, &r.LegalBasis,
			&r.DataProtectionAuth, &r.DPAContactURL,
			&r.GDPRAdequateCountries,
			&r.EnforcementMode,
			&r.AllowCrossRegionSearch, &r.AllowCrossRegionBackup,
			&r.IsActive,
			&r.CreatedAt, &r.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan region: %w", err)
		}
		regions = append(regions, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate regions: %w", err)
	}
	if regions == nil {
		regions = []DataResidencyRegion{}
	}
	return regions, nil
}

// ============================================================
// DASHBOARD
// ============================================================

// GetResidencyDashboard returns the data residency dashboard for an
// organisation, including current region, storage locations, compliance
// status, and recent blocked attempts.
func (s *DataResidencyService) GetResidencyDashboard(ctx context.Context, orgID uuid.UUID) (*ResidencyDashboard, error) {
	cfg, err := s.GetRegionConfig(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("get config for dashboard: %w", err)
	}

	dash := &ResidencyDashboard{
		CurrentRegion:        cfg.DataRegion,
		RegionDisplayName:    cfg.DisplayName,
		Enforced:             cfg.DataResidencyEnforced,
		EnforcementMode:      cfg.EnforcementMode,
		ComplianceFrameworks: cfg.ComplianceFrameworks,
		AllowedCloudRegions:  cfg.AllowedCloudRegions,
	}

	if cfg.ComplianceFrameworks == nil {
		dash.ComplianceFrameworks = []string{}
	}

	// Determine compliance status
	if cfg.DataResidencyEnforced && cfg.EnforcementMode == "enforce" {
		dash.ComplianceStatus = "compliant"
	} else if cfg.EnforcementMode == "audit" {
		dash.ComplianceStatus = "monitoring"
	} else {
		dash.ComplianceStatus = "not_enforced"
	}

	// Build storage locations based on the region config
	dash.StorageLocations = buildStorageLocations(cfg)

	// Get recent blocked attempts (last 10)
	blockedRows, err := s.pool.Query(ctx, `
		SELECT id, organization_id, action::text,
		       user_id, COALESCE(user_email, ''),
		       COALESCE(host(ip_address), ''),
		       COALESCE(source_region::text, ''), COALESCE(destination_region, ''),
		       COALESCE(source_country, ''), COALESCE(destination_country, ''),
		       COALESCE(resource_type, ''), resource_id,
		       allowed, COALESCE(blocked_reason, ''),
		       vendor_id, COALESCE(vendor_name, ''),
		       COALESCE(transfer_mechanism, ''),
		       created_at
		FROM data_residency_audit_log
		WHERE organization_id = $1 AND allowed = false
		ORDER BY created_at DESC
		LIMIT 10`, orgID)
	if err == nil {
		defer blockedRows.Close()
		for blockedRows.Next() {
			var e ResidencyAuditEntry
			if blockedRows.Scan(
				&e.ID, &e.OrganizationID, &e.Action,
				&e.UserID, &e.UserEmail, &e.IPAddress,
				&e.SourceRegion, &e.DestinationRegion,
				&e.SourceCountry, &e.DestinationCountry,
				&e.ResourceType, &e.ResourceID,
				&e.Allowed, &e.BlockedReason,
				&e.VendorID, &e.VendorName,
				&e.TransferMechanism,
				&e.CreatedAt,
			) == nil {
				dash.RecentBlocked = append(dash.RecentBlocked, e)
			}
		}
	}
	if dash.RecentBlocked == nil {
		dash.RecentBlocked = []ResidencyAuditEntry{}
	}

	// Count total blocked
	_ = s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM data_residency_audit_log
		WHERE organization_id = $1 AND allowed = false`, orgID).Scan(&dash.TotalBlockedCount)

	// Get most recent audit entry timestamp
	var lastEntry *time.Time
	err = s.pool.QueryRow(ctx, `
		SELECT MAX(created_at) FROM data_residency_audit_log
		WHERE organization_id = $1`, orgID).Scan(&lastEntry)
	if err == nil && lastEntry != nil {
		dash.LastAuditEntry = lastEntry
	}

	return dash, nil
}

// buildStorageLocations creates storage location entries based on
// the organisation's residency config.
func buildStorageLocations(cfg *ResidencyConfig) []StorageLocation {
	primaryRegion := cfg.PrimaryCloudRegion
	if primaryRegion == "" {
		primaryRegion = cfg.DataRegion
	}

	provider := "AWS"
	if strings.Contains(primaryRegion, "azure") || strings.Contains(primaryRegion, "westeurope") || strings.Contains(primaryRegion, "northeurope") || strings.Contains(primaryRegion, "uksouth") {
		provider = "Azure"
	} else if strings.Contains(primaryRegion, "europe-") || strings.Contains(primaryRegion, "us-") {
		provider = "GCP"
	}

	locations := []StorageLocation{
		{
			Type:               "database",
			Label:              "Primary Database (PostgreSQL)",
			Region:             primaryRegion,
			Provider:           provider,
			Status:             "active",
			WithinJurisdiction: true,
		},
		{
			Type:               "files",
			Label:              "File Storage (Object Storage)",
			Region:             primaryRegion,
			Provider:           provider,
			Status:             "active",
			WithinJurisdiction: true,
		},
		{
			Type:               "cache",
			Label:              "Cache Layer (Redis)",
			Region:             primaryRegion,
			Provider:           provider,
			Status:             "active",
			WithinJurisdiction: true,
		},
		{
			Type:               "search",
			Label:              "Search Index (Elasticsearch)",
			Region:             primaryRegion,
			Provider:           provider,
			Status:             "active",
			WithinJurisdiction: true,
		},
	}

	// Add failover if configured
	if cfg.FailoverCloudRegion != "" {
		locations = append(locations, StorageLocation{
			Type:               "backup",
			Label:              "Disaster Recovery Replica",
			Region:             cfg.FailoverCloudRegion,
			Provider:           provider,
			Status:             "standby",
			WithinJurisdiction: true,
		})
	}

	return locations
}

// ============================================================
// AUDIT LOG
// ============================================================

// GetAuditLog returns a filtered, paginated audit log for an organisation.
func (s *DataResidencyService) GetAuditLog(ctx context.Context, orgID uuid.UUID, filter AuditLogFilter) ([]ResidencyAuditEntry, int, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 20
	}
	offset := (filter.Page - 1) * filter.PageSize

	var conditions []string
	var args []interface{}
	argIdx := 1

	conditions = append(conditions, fmt.Sprintf("organization_id = $%d", argIdx))
	args = append(args, orgID)
	argIdx++

	if filter.Action != "" {
		conditions = append(conditions, fmt.Sprintf("action::text = $%d", argIdx))
		args = append(args, filter.Action)
		argIdx++
	}
	if filter.Allowed != nil {
		conditions = append(conditions, fmt.Sprintf("allowed = $%d", argIdx))
		args = append(args, *filter.Allowed)
		argIdx++
	}
	if filter.DateFrom != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argIdx))
		args = append(args, *filter.DateFrom)
		argIdx++
	}
	if filter.DateTo != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argIdx))
		args = append(args, *filter.DateTo)
		argIdx++
	}

	where := "WHERE " + strings.Join(conditions, " AND ")

	// Count
	countSQL := fmt.Sprintf(`SELECT COUNT(*) FROM data_residency_audit_log %s`, where)
	var total int
	if err := s.pool.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count audit entries: %w", err)
	}

	// Data
	dataSQL := fmt.Sprintf(`
		SELECT id, organization_id, action::text,
		       user_id, COALESCE(user_email, ''),
		       COALESCE(host(ip_address), ''),
		       COALESCE(source_region::text, ''), COALESCE(destination_region, ''),
		       COALESCE(source_country, ''), COALESCE(destination_country, ''),
		       COALESCE(resource_type, ''), resource_id,
		       allowed, COALESCE(blocked_reason, ''),
		       vendor_id, COALESCE(vendor_name, ''),
		       COALESCE(transfer_mechanism, ''),
		       created_at
		FROM data_residency_audit_log
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)

	args = append(args, filter.PageSize, offset)

	rows, err := s.pool.Query(ctx, dataSQL, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query audit log: %w", err)
	}
	defer rows.Close()

	var entries []ResidencyAuditEntry
	for rows.Next() {
		var e ResidencyAuditEntry
		if err := rows.Scan(
			&e.ID, &e.OrganizationID, &e.Action,
			&e.UserID, &e.UserEmail, &e.IPAddress,
			&e.SourceRegion, &e.DestinationRegion,
			&e.SourceCountry, &e.DestinationCountry,
			&e.ResourceType, &e.ResourceID,
			&e.Allowed, &e.BlockedReason,
			&e.VendorID, &e.VendorName,
			&e.TransferMechanism,
			&e.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan audit entry: %w", err)
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate audit entries: %w", err)
	}
	if entries == nil {
		entries = []ResidencyAuditEntry{}
	}
	return entries, total, nil
}

// LogDataAccess records a data access or transfer event in the audit log.
func (s *DataResidencyService) LogDataAccess(ctx context.Context, orgID uuid.UUID, entry ResidencyAuditEntry) error {
	var ipAddr interface{} = nil
	if entry.IPAddress != "" {
		ipAddr = entry.IPAddress
	}

	var srcRegion interface{} = nil
	if entry.SourceRegion != "" {
		srcRegion = entry.SourceRegion
	}

	var srcCountry interface{} = nil
	if entry.SourceCountry != "" {
		srcCountry = entry.SourceCountry
	}

	var dstCountry interface{} = nil
	if entry.DestinationCountry != "" {
		dstCountry = entry.DestinationCountry
	}

	var dstRegion interface{} = nil
	if entry.DestinationRegion != "" {
		dstRegion = entry.DestinationRegion
	}

	_, err := s.pool.Exec(ctx, `
		INSERT INTO data_residency_audit_log
			(organization_id, action, user_id, user_email, ip_address,
			 source_region, destination_region, source_country, destination_country,
			 resource_type, resource_id,
			 allowed, blocked_reason,
			 vendor_id, vendor_name, transfer_mechanism,
			 details)
		VALUES ($1, $2::audit_action_type, $3, $4, $5::inet,
		        $6::data_residency_region, $7, $8, $9,
		        $10, $11,
		        $12, $13,
		        $14, $15, $16,
		        $17)`,
		orgID, entry.Action, entry.UserID, entry.UserEmail, ipAddr,
		srcRegion, dstRegion, srcCountry, dstCountry,
		entry.ResourceType, entry.ResourceID,
		entry.Allowed, entry.BlockedReason,
		entry.VendorID, entry.VendorName, entry.TransferMechanism,
		"{}",
	)
	if err != nil {
		log.Error().Err(err).Str("org_id", orgID.String()).Str("action", entry.Action).
			Msg("failed to log data access audit entry")
		return fmt.Errorf("log data access: %w", err)
	}
	return nil
}

// ============================================================
// VALIDATION
// ============================================================

// ValidateDataExport checks whether exporting data from an organisation's
// region to the given destination region is allowed under the configured
// residency rules.
func (s *DataResidencyService) ValidateDataExport(ctx context.Context, orgID uuid.UUID, destinationRegion string) (*ExportValidationResult, error) {
	cfg, err := s.GetRegionConfig(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("get config for export validation: %w", err)
	}

	result := &ExportValidationResult{
		SourceRegion:      cfg.DataRegion,
		DestinationRegion: destinationRegion,
	}

	// Global region allows export anywhere
	if cfg.DataRegion == "global" {
		result.Allowed = true
		result.Reason = "Organization is in global region; no export restrictions apply"
		return result, nil
	}

	// Same region is always allowed
	if strings.EqualFold(cfg.DataRegion, destinationRegion) {
		result.Allowed = true
		result.Reason = "Export within the same residency region is allowed"
		result.LegalBasis = cfg.LegalBasis
		return result, nil
	}

	// If residency is not enforced, allow with advisory
	if !cfg.DataResidencyEnforced || cfg.EnforcementMode == "disabled" {
		result.Allowed = true
		result.Reason = "Data residency is not enforced; cross-region export is allowed but may require additional safeguards"
		result.RequiredSafeguards = []string{
			"Standard Contractual Clauses (SCCs)",
			"Data Processing Agreement (DPA)",
			"Transfer Impact Assessment (TIA)",
		}
		return result, nil
	}

	// Audit mode: allow but flag
	if cfg.EnforcementMode == "audit" {
		result.Allowed = true
		result.Reason = "Cross-region export detected (audit mode); export allowed but logged for review"
		result.RequiredSafeguards = []string{
			"Standard Contractual Clauses (SCCs)",
			"Binding Corporate Rules (BCRs)",
			"Transfer Impact Assessment (TIA)",
		}
		result.LegalBasis = cfg.LegalBasis
		return result, nil
	}

	// Enforce mode: block cross-region export
	result.Allowed = false
	result.Reason = fmt.Sprintf(
		"Data export from region '%s' to '%s' is blocked by data residency enforcement policy",
		cfg.DataRegion, destinationRegion,
	)
	result.LegalBasis = cfg.LegalBasis

	return result, nil
}

// ValidateDataTransfer checks whether a data transfer to a vendor in the
// specified destination country is permissible under GDPR adequacy decisions
// and the organisation's residency configuration.
func (s *DataResidencyService) ValidateDataTransfer(ctx context.Context, orgID, vendorID uuid.UUID, destinationCountry string) (*TransferValidation, error) {
	cfg, err := s.GetRegionConfig(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("get config for transfer validation: %w", err)
	}

	destCountry := strings.ToUpper(strings.TrimSpace(destinationCountry))

	result := &TransferValidation{
		DestinationCountry:  destCountry,
		DataProtectionAuth:  cfg.DataProtectionAuth,
	}

	// Check if destination is in the allowed countries for the region
	if isCountryInList(destCountry, cfg.AllowedCountries) {
		result.Allowed = true
		result.IsGDPRAdequate = true
		result.Reason = fmt.Sprintf(
			"Country %s is within the organisation's configured data residency region (%s)",
			destCountry, cfg.DataRegion,
		)
		return result, nil
	}

	// Check if destination country is blocked
	if isCountryInList(destCountry, cfg.BlockedCountries) {
		result.Allowed = false
		result.IsGDPRAdequate = false
		result.Reason = fmt.Sprintf(
			"Country %s is explicitly blocked by the organisation's data residency policy",
			destCountry,
		)
		return result, nil
	}

	// Check GDPR adequacy
	adequateCountries := cfg.GDPRAdequateCountries
	if len(adequateCountries) == 0 {
		adequateCountries = GDPRAdequateCountries
	}

	if isCountryInList(destCountry, adequateCountries) {
		result.Allowed = true
		result.IsGDPRAdequate = true
		result.RequiresAdditional = false
		result.Reason = fmt.Sprintf(
			"Country %s has a GDPR adequacy decision; data transfer is permitted",
			destCountry,
		)
		result.TransferMechanisms = []string{"GDPR Adequacy Decision"}
		return result, nil
	}

	// Check if destination is an EU/EEA member state
	if isCountryInList(destCountry, EUMemberStates) {
		result.Allowed = true
		result.IsGDPRAdequate = true
		result.RequiresAdditional = false
		result.Reason = fmt.Sprintf(
			"Country %s is an EU/EEA member state; intra-EU transfers are permitted",
			destCountry,
		)
		result.TransferMechanisms = []string{"EU/EEA Free Movement of Data"}
		return result, nil
	}

	// No adequacy decision: transfer requires additional safeguards
	result.Allowed = false
	result.IsGDPRAdequate = false
	result.RequiresAdditional = true
	result.Reason = fmt.Sprintf(
		"Country %s does not have a GDPR adequacy decision; transfer requires additional safeguards",
		destCountry,
	)
	result.TransferMechanisms = []string{
		"Standard Contractual Clauses (SCCs)",
		"Binding Corporate Rules (BCRs)",
		"Explicit Consent (Article 49 GDPR)",
	}
	result.RequiredSafeguards = []string{
		"Transfer Impact Assessment (TIA)",
		"Supplementary Measures Assessment",
		"Data Processing Agreement with vendor",
		"Record in Article 30 Records of Processing",
	}

	return result, nil
}

// ============================================================
// IP GEOLOCATION
// ============================================================

// CheckIPRegion resolves an IP address to a country/region using the
// built-in IP range database. In production, this would delegate to a
// full GeoIP service (MaxMind, IP2Location, etc.).
func (s *DataResidencyService) CheckIPRegion(ipAddress string) (*GeoLocation, error) {
	ip := net.ParseIP(strings.TrimSpace(ipAddress))
	if ip == nil {
		return nil, fmt.Errorf("invalid IP address: %s", ipAddress)
	}

	geo := &GeoLocation{
		IPAddress: ipAddress,
	}

	for _, entry := range ipCountryRanges {
		_, network, err := net.ParseCIDR(entry.Network)
		if err != nil {
			continue
		}
		if network.Contains(ip) {
			geo.CountryCode = entry.CountryCode
			geo.CountryName = entry.CountryName
			geo.IsEU = entry.IsEU
			geo.Region = countryToResidencyRegion(entry.CountryCode)
			return geo, nil
		}
	}

	// IP not found in known ranges - use heuristic based on first octet
	geo.CountryCode = "XX"
	geo.CountryName = "Unknown"
	geo.Region = "global"
	geo.IsEU = false

	// Try to resolve via reverse DNS as a hint
	ipInt := ipToInt(ip)
	if ipInt != nil {
		// European IP space is generally in the 2.x-95.x and 176.x-213.x ranges
		firstOctet := ip[len(ip)-4]
		if (firstOctet >= 2 && firstOctet <= 95) || (firstOctet >= 176 && firstOctet <= 213) {
			geo.Region = "eu"
		}
	}

	return geo, nil
}

// ipToInt converts an IP address to a big.Int for comparison purposes.
func ipToInt(ip net.IP) *big.Int {
	ip = ip.To4()
	if ip == nil {
		return nil
	}
	return new(big.Int).SetBytes(ip)
}

// countryToResidencyRegion maps a country code to the most specific
// data residency region. More specific regions (dach, nordics, france,
// uk) are checked before the broader "eu" region to ensure that a
// country like DE resolves to "dach" rather than "eu".
func countryToResidencyRegion(countryCode string) string {
	// Check specific sub-regions first, then broader regions.
	// Order matters: most specific regions must come before "eu".
	orderedRegions := []string{"france", "uk", "dach", "nordics", "eu"}
	for _, region := range orderedRegions {
		countries := regionToCountryMap[region]
		for _, cc := range countries {
			if strings.EqualFold(cc, countryCode) {
				return region
			}
		}
	}
	return "global"
}

// isCountryInList checks if a country code is in the given list.
func isCountryInList(country string, list []string) bool {
	for _, c := range list {
		if strings.EqualFold(c, country) {
			return true
		}
	}
	return false
}
