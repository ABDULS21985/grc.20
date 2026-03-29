package service

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
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

// TenantBranding represents the full branding configuration for an organisation.
type TenantBranding struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`

	// Identity
	ProductName       string `json:"product_name"`
	Tagline           string `json:"tagline"`
	CompanyName       string `json:"company_name"`
	SupportEmail      string `json:"support_email"`
	SupportURL        string `json:"support_url"`
	PrivacyPolicyURL  string `json:"privacy_policy_url"`
	TermsOfServiceURL string `json:"terms_of_service_url"`

	// Logos
	LogoFullURL  string `json:"logo_full_url"`
	LogoIconURL  string `json:"logo_icon_url"`
	LogoDarkURL  string `json:"logo_dark_url"`
	LogoLightURL string `json:"logo_light_url"`
	FaviconURL   string `json:"favicon_url"`
	EmailLogoURL string `json:"email_logo_url"`

	// Colours
	ColorPrimary        string `json:"color_primary"`
	ColorPrimaryHover   string `json:"color_primary_hover"`
	ColorSecondary      string `json:"color_secondary"`
	ColorSecondaryHover string `json:"color_secondary_hover"`
	ColorAccent         string `json:"color_accent"`
	ColorBackground     string `json:"color_background"`
	ColorSurface        string `json:"color_surface"`
	ColorTextPrimary    string `json:"color_text_primary"`
	ColorTextSecondary  string `json:"color_text_secondary"`
	ColorBorder         string `json:"color_border"`
	ColorSuccess        string `json:"color_success"`
	ColorWarning        string `json:"color_warning"`
	ColorError          string `json:"color_error"`
	ColorInfo           string `json:"color_info"`
	ColorSidebarBg      string `json:"color_sidebar_bg"`
	ColorSidebarText    string `json:"color_sidebar_text"`

	// Typography
	FontFamilyHeading string `json:"font_family_heading"`
	FontFamilyBody    string `json:"font_family_body"`
	FontSizeBase      string `json:"font_size_base"`

	// Layout
	SidebarStyle string `json:"sidebar_style"`
	CornerRadius string `json:"corner_radius"`
	Density      string `json:"density"`

	// Custom Domain
	CustomDomain             string     `json:"custom_domain"`
	DomainVerificationToken  string     `json:"domain_verification_token"`
	DomainVerificationStatus string     `json:"domain_verification_status"`
	DomainVerifiedAt         *time.Time `json:"domain_verified_at"`
	SSLStatus                string     `json:"ssl_status"`
	SSLProvisionedAt         *time.Time `json:"ssl_provisioned_at"`
	SSLExpiresAt             *time.Time `json:"ssl_expires_at"`

	// Custom CSS
	CustomCSS string `json:"custom_css"`

	// Feature Flags
	ShowPoweredBy     bool `json:"show_powered_by"`
	ShowHelpWidget    bool `json:"show_help_widget"`
	ShowMarketplace   bool `json:"show_marketplace"`
	ShowKnowledgeBase bool `json:"show_knowledge_base"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// WhiteLabelPartner represents a reseller partner.
type WhiteLabelPartner struct {
	ID                  uuid.UUID  `json:"id"`
	PartnerName         string     `json:"partner_name"`
	PartnerSlug         string     `json:"partner_slug"`
	ContactEmail        string     `json:"contact_email"`
	DefaultBrandingID   *uuid.UUID `json:"default_branding_id"`
	RevenueSharePercent float64    `json:"revenue_share_percent"`
	MaxTenants          int        `json:"max_tenants"`
	IsActive            bool       `json:"is_active"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// PartnerTenantMapping links a partner to a managed organisation.
type PartnerTenantMapping struct {
	ID             uuid.UUID `json:"id"`
	PartnerID      uuid.UUID `json:"partner_id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	OnboardedAt    time.Time `json:"onboarded_at"`
	CreatedAt      time.Time `json:"created_at"`
}

// ============================================================
// REQUEST / RESPONSE TYPES
// ============================================================

// UpdateBrandingRequest is the payload for updating branding settings.
type UpdateBrandingRequest struct {
	ProductName       *string `json:"product_name"`
	Tagline           *string `json:"tagline"`
	CompanyName       *string `json:"company_name"`
	SupportEmail      *string `json:"support_email"`
	SupportURL        *string `json:"support_url"`
	PrivacyPolicyURL  *string `json:"privacy_policy_url"`
	TermsOfServiceURL *string `json:"terms_of_service_url"`

	ColorPrimary        *string `json:"color_primary"`
	ColorPrimaryHover   *string `json:"color_primary_hover"`
	ColorSecondary      *string `json:"color_secondary"`
	ColorSecondaryHover *string `json:"color_secondary_hover"`
	ColorAccent         *string `json:"color_accent"`
	ColorBackground     *string `json:"color_background"`
	ColorSurface        *string `json:"color_surface"`
	ColorTextPrimary    *string `json:"color_text_primary"`
	ColorTextSecondary  *string `json:"color_text_secondary"`
	ColorBorder         *string `json:"color_border"`
	ColorSuccess        *string `json:"color_success"`
	ColorWarning        *string `json:"color_warning"`
	ColorError          *string `json:"color_error"`
	ColorInfo           *string `json:"color_info"`
	ColorSidebarBg      *string `json:"color_sidebar_bg"`
	ColorSidebarText    *string `json:"color_sidebar_text"`

	FontFamilyHeading *string `json:"font_family_heading"`
	FontFamilyBody    *string `json:"font_family_body"`
	FontSizeBase      *string `json:"font_size_base"`

	SidebarStyle *string `json:"sidebar_style"`
	CornerRadius *string `json:"corner_radius"`
	Density      *string `json:"density"`

	CustomCSS *string `json:"custom_css"`

	ShowPoweredBy     *bool `json:"show_powered_by"`
	ShowHelpWidget    *bool `json:"show_help_widget"`
	ShowMarketplace   *bool `json:"show_marketplace"`
	ShowKnowledgeBase *bool `json:"show_knowledge_base"`
}

// DomainVerificationResult is the result of a custom domain verification check.
type DomainVerificationResult struct {
	Domain     string `json:"domain"`
	Status     string `json:"status"`
	CNAMEValid bool   `json:"cname_valid"`
	TXTValid   bool   `json:"txt_valid"`
	Token      string `json:"verification_token"`
	SSLStatus  string `json:"ssl_status"`
	Message    string `json:"message"`
}

// DomainStatusResponse returns the current domain + SSL status.
type DomainStatusResponse struct {
	CustomDomain             string     `json:"custom_domain"`
	DomainVerificationStatus string     `json:"domain_verification_status"`
	DomainVerifiedAt         *time.Time `json:"domain_verified_at"`
	SSLStatus                string     `json:"ssl_status"`
	SSLProvisionedAt         *time.Time `json:"ssl_provisioned_at"`
	SSLExpiresAt             *time.Time `json:"ssl_expires_at"`
	VerificationToken        string     `json:"verification_token"`
}

// PDFBrandingConfig holds branding settings injected into PDF reports.
type PDFBrandingConfig struct {
	ProductName      string `json:"product_name"`
	CompanyName      string `json:"company_name"`
	LogoURL          string `json:"logo_url"`
	ColorPrimary     string `json:"color_primary"`
	ColorSecondary   string `json:"color_secondary"`
	FontFamily       string `json:"font_family"`
	ShowPoweredBy    bool   `json:"show_powered_by"`
	PrivacyPolicyURL string `json:"privacy_policy_url"`
}

// CreatePartnerRequest is the payload for creating a white-label partner.
type CreatePartnerRequest struct {
	PartnerName         string  `json:"partner_name"`
	PartnerSlug         string  `json:"partner_slug"`
	ContactEmail        string  `json:"contact_email"`
	RevenueSharePercent float64 `json:"revenue_share_percent"`
	MaxTenants          int     `json:"max_tenants"`
}

// UpdatePartnerRequest is the payload for updating a white-label partner.
type UpdatePartnerRequest struct {
	PartnerName         *string  `json:"partner_name"`
	ContactEmail        *string  `json:"contact_email"`
	RevenueSharePercent *float64 `json:"revenue_share_percent"`
	MaxTenants          *int     `json:"max_tenants"`
	IsActive            *bool    `json:"is_active"`
	DefaultBrandingID   *string  `json:"default_branding_id"`
}

// ============================================================
// SERVICE
// ============================================================

// BrandingService provides branding, theming, and white-label partner management.
type BrandingService struct {
	pool        *pgxpool.Pool
	storagePath string // base path for file storage
}

// NewBrandingService creates a new BrandingService.
func NewBrandingService(pool *pgxpool.Pool) *BrandingService {
	storagePath := os.Getenv("STORAGE_PATH")
	if storagePath == "" {
		storagePath = "./storage"
	}
	return &BrandingService{
		pool:        pool,
		storagePath: storagePath,
	}
}

// ============================================================
// DEFAULTS
// ============================================================

// DefaultBranding returns the ComplianceForge factory-default branding.
func DefaultBranding(orgID uuid.UUID) TenantBranding {
	return TenantBranding{
		OrganizationID:           orgID,
		ProductName:              "ComplianceForge",
		Tagline:                  "Enterprise GRC Platform",
		CompanyName:              "ComplianceForge",
		SupportEmail:             "support@complianceforge.io",
		SupportURL:               "https://help.complianceforge.io",
		ColorPrimary:             "#4F46E5",
		ColorPrimaryHover:        "#4338CA",
		ColorSecondary:           "#7C3AED",
		ColorSecondaryHover:      "#6D28D9",
		ColorAccent:              "#06B6D4",
		ColorBackground:          "#F9FAFB",
		ColorSurface:             "#FFFFFF",
		ColorTextPrimary:         "#111827",
		ColorTextSecondary:       "#6B7280",
		ColorBorder:              "#E5E7EB",
		ColorSuccess:             "#10B981",
		ColorWarning:             "#F59E0B",
		ColorError:               "#EF4444",
		ColorInfo:                "#3B82F6",
		ColorSidebarBg:           "#1F2937",
		ColorSidebarText:         "#F9FAFB",
		FontFamilyHeading:        "Inter",
		FontFamilyBody:           "Inter",
		FontSizeBase:             "14px",
		SidebarStyle:             "dark",
		CornerRadius:             "medium",
		Density:                  "comfortable",
		DomainVerificationStatus: "pending",
		SSLStatus:                "pending",
		ShowPoweredBy:            true,
		ShowHelpWidget:           true,
		ShowMarketplace:          true,
		ShowKnowledgeBase:        true,
	}
}

// ============================================================
// GET BRANDING
// Falls back: org custom -> partner default -> ComplianceForge defaults
// ============================================================

// GetBranding returns the branding configuration for an organisation,
// falling back to partner defaults then ComplianceForge defaults.
func (s *BrandingService) GetBranding(ctx context.Context, orgID uuid.UUID) (*TenantBranding, error) {
	// 1. Try org-specific branding
	branding, err := s.getBrandingByOrg(ctx, orgID)
	if err == nil {
		return branding, nil
	}

	// 2. Try partner default branding
	branding, err = s.getPartnerDefaultBranding(ctx, orgID)
	if err == nil {
		return branding, nil
	}

	// 3. Return ComplianceForge defaults
	defaults := DefaultBranding(orgID)
	return &defaults, nil
}

func (s *BrandingService) getBrandingByOrg(ctx context.Context, orgID uuid.UUID) (*TenantBranding, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return nil, fmt.Errorf("set org: %w", err)
	}

	var b TenantBranding
	err = tx.QueryRow(ctx, `
		SELECT id, organization_id,
			product_name, COALESCE(tagline,''), COALESCE(company_name,''),
			COALESCE(support_email,''), COALESCE(support_url,''),
			COALESCE(privacy_policy_url,''), COALESCE(terms_of_service_url,''),
			COALESCE(logo_full_url,''), COALESCE(logo_icon_url,''),
			COALESCE(logo_dark_url,''), COALESCE(logo_light_url,''),
			COALESCE(favicon_url,''), COALESCE(email_logo_url,''),
			color_primary, color_primary_hover, color_secondary, color_secondary_hover,
			color_accent, color_background, color_surface,
			color_text_primary, color_text_secondary, color_border,
			color_success, color_warning, color_error, color_info,
			color_sidebar_bg, color_sidebar_text,
			font_family_heading, font_family_body, font_size_base,
			sidebar_style, corner_radius, density,
			COALESCE(custom_domain,''), COALESCE(domain_verification_token,''),
			domain_verification_status, domain_verified_at,
			ssl_status, ssl_provisioned_at, ssl_expires_at,
			COALESCE(custom_css,''),
			show_powered_by, show_help_widget, show_marketplace, show_knowledge_base,
			created_at, updated_at
		FROM tenant_branding
		WHERE organization_id = $1
	`, orgID).Scan(
		&b.ID, &b.OrganizationID,
		&b.ProductName, &b.Tagline, &b.CompanyName,
		&b.SupportEmail, &b.SupportURL,
		&b.PrivacyPolicyURL, &b.TermsOfServiceURL,
		&b.LogoFullURL, &b.LogoIconURL,
		&b.LogoDarkURL, &b.LogoLightURL,
		&b.FaviconURL, &b.EmailLogoURL,
		&b.ColorPrimary, &b.ColorPrimaryHover, &b.ColorSecondary, &b.ColorSecondaryHover,
		&b.ColorAccent, &b.ColorBackground, &b.ColorSurface,
		&b.ColorTextPrimary, &b.ColorTextSecondary, &b.ColorBorder,
		&b.ColorSuccess, &b.ColorWarning, &b.ColorError, &b.ColorInfo,
		&b.ColorSidebarBg, &b.ColorSidebarText,
		&b.FontFamilyHeading, &b.FontFamilyBody, &b.FontSizeBase,
		&b.SidebarStyle, &b.CornerRadius, &b.Density,
		&b.CustomDomain, &b.DomainVerificationToken,
		&b.DomainVerificationStatus, &b.DomainVerifiedAt,
		&b.SSLStatus, &b.SSLProvisionedAt, &b.SSLExpiresAt,
		&b.CustomCSS,
		&b.ShowPoweredBy, &b.ShowHelpWidget, &b.ShowMarketplace, &b.ShowKnowledgeBase,
		&b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &b, tx.Commit(ctx)
}

func (s *BrandingService) getPartnerDefaultBranding(ctx context.Context, orgID uuid.UUID) (*TenantBranding, error) {
	var brandingID uuid.UUID

	err := s.pool.QueryRow(ctx, `
		SELECT wlp.default_branding_id
		FROM partner_tenant_mappings ptm
		JOIN white_label_partners wlp ON wlp.id = ptm.partner_id
		WHERE ptm.organization_id = $1
			AND wlp.is_active = true
			AND wlp.default_branding_id IS NOT NULL
		LIMIT 1
	`, orgID).Scan(&brandingID)
	if err != nil {
		return nil, err
	}

	var b TenantBranding
	err = s.pool.QueryRow(ctx, `
		SELECT id, organization_id,
			product_name, COALESCE(tagline,''), COALESCE(company_name,''),
			COALESCE(support_email,''), COALESCE(support_url,''),
			COALESCE(privacy_policy_url,''), COALESCE(terms_of_service_url,''),
			COALESCE(logo_full_url,''), COALESCE(logo_icon_url,''),
			COALESCE(logo_dark_url,''), COALESCE(logo_light_url,''),
			COALESCE(favicon_url,''), COALESCE(email_logo_url,''),
			color_primary, color_primary_hover, color_secondary, color_secondary_hover,
			color_accent, color_background, color_surface,
			color_text_primary, color_text_secondary, color_border,
			color_success, color_warning, color_error, color_info,
			color_sidebar_bg, color_sidebar_text,
			font_family_heading, font_family_body, font_size_base,
			sidebar_style, corner_radius, density,
			COALESCE(custom_domain,''), COALESCE(domain_verification_token,''),
			domain_verification_status, domain_verified_at,
			ssl_status, ssl_provisioned_at, ssl_expires_at,
			COALESCE(custom_css,''),
			show_powered_by, show_help_widget, show_marketplace, show_knowledge_base,
			created_at, updated_at
		FROM tenant_branding
		WHERE id = $1
	`, brandingID).Scan(
		&b.ID, &b.OrganizationID,
		&b.ProductName, &b.Tagline, &b.CompanyName,
		&b.SupportEmail, &b.SupportURL,
		&b.PrivacyPolicyURL, &b.TermsOfServiceURL,
		&b.LogoFullURL, &b.LogoIconURL,
		&b.LogoDarkURL, &b.LogoLightURL,
		&b.FaviconURL, &b.EmailLogoURL,
		&b.ColorPrimary, &b.ColorPrimaryHover, &b.ColorSecondary, &b.ColorSecondaryHover,
		&b.ColorAccent, &b.ColorBackground, &b.ColorSurface,
		&b.ColorTextPrimary, &b.ColorTextSecondary, &b.ColorBorder,
		&b.ColorSuccess, &b.ColorWarning, &b.ColorError, &b.ColorInfo,
		&b.ColorSidebarBg, &b.ColorSidebarText,
		&b.FontFamilyHeading, &b.FontFamilyBody, &b.FontSizeBase,
		&b.SidebarStyle, &b.CornerRadius, &b.Density,
		&b.CustomDomain, &b.DomainVerificationToken,
		&b.DomainVerificationStatus, &b.DomainVerifiedAt,
		&b.SSLStatus, &b.SSLProvisionedAt, &b.SSLExpiresAt,
		&b.CustomCSS,
		&b.ShowPoweredBy, &b.ShowHelpWidget, &b.ShowMarketplace, &b.ShowKnowledgeBase,
		&b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &b, nil
}

// ============================================================
// UPDATE BRANDING
// ============================================================

// UpdateBranding updates the branding configuration for an organisation.
// It validates hex colours, URLs, sanitizes custom CSS, and upserts the row.
func (s *BrandingService) UpdateBranding(ctx context.Context, orgID uuid.UUID, req UpdateBrandingRequest) (*TenantBranding, error) {
	// Validate hex colour fields
	colourFields := map[string]*string{
		"color_primary":         req.ColorPrimary,
		"color_primary_hover":   req.ColorPrimaryHover,
		"color_secondary":       req.ColorSecondary,
		"color_secondary_hover": req.ColorSecondaryHover,
		"color_accent":          req.ColorAccent,
		"color_background":      req.ColorBackground,
		"color_surface":         req.ColorSurface,
		"color_text_primary":    req.ColorTextPrimary,
		"color_text_secondary":  req.ColorTextSecondary,
		"color_border":          req.ColorBorder,
		"color_success":         req.ColorSuccess,
		"color_warning":         req.ColorWarning,
		"color_error":           req.ColorError,
		"color_info":            req.ColorInfo,
		"color_sidebar_bg":      req.ColorSidebarBg,
		"color_sidebar_text":    req.ColorSidebarText,
	}
	for name, val := range colourFields {
		if val != nil {
			if err := ValidateHexColour(*val); err != nil {
				return nil, fmt.Errorf("invalid %s: %w", name, err)
			}
		}
	}

	// Validate URL fields
	urlFields := map[string]*string{
		"support_url":          req.SupportURL,
		"privacy_policy_url":   req.PrivacyPolicyURL,
		"terms_of_service_url": req.TermsOfServiceURL,
	}
	for name, val := range urlFields {
		if val != nil && *val != "" {
			if err := validateURL(*val); err != nil {
				return nil, fmt.Errorf("invalid %s: %w", name, err)
			}
		}
	}

	// Validate enum fields
	if req.SidebarStyle != nil {
		if !isValidEnum(*req.SidebarStyle, []string{"light", "dark", "branded"}) {
			return nil, fmt.Errorf("invalid sidebar_style: must be light, dark, or branded")
		}
	}
	if req.CornerRadius != nil {
		if !isValidEnum(*req.CornerRadius, []string{"none", "small", "medium", "large", "full"}) {
			return nil, fmt.Errorf("invalid corner_radius: must be none, small, medium, large, or full")
		}
	}
	if req.Density != nil {
		if !isValidEnum(*req.Density, []string{"compact", "comfortable", "spacious"}) {
			return nil, fmt.Errorf("invalid density: must be compact, comfortable, or spacious")
		}
	}

	// Sanitize custom CSS
	if req.CustomCSS != nil {
		sanitized := SanitizeCSS(*req.CustomCSS)
		req.CustomCSS = &sanitized
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return nil, fmt.Errorf("set org: %w", err)
	}

	// Upsert: insert default then update
	_, err = tx.Exec(ctx, `
		INSERT INTO tenant_branding (organization_id)
		VALUES ($1)
		ON CONFLICT (organization_id) DO NOTHING
	`, orgID)
	if err != nil {
		return nil, fmt.Errorf("upsert branding: %w", err)
	}

	// Build dynamic UPDATE
	setClauses := []string{}
	args := []interface{}{orgID}
	argIdx := 2

	addField := func(col string, val interface{}) {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", col, argIdx))
		args = append(args, val)
		argIdx++
	}

	if req.ProductName != nil {
		addField("product_name", *req.ProductName)
	}
	if req.Tagline != nil {
		addField("tagline", *req.Tagline)
	}
	if req.CompanyName != nil {
		addField("company_name", *req.CompanyName)
	}
	if req.SupportEmail != nil {
		addField("support_email", *req.SupportEmail)
	}
	if req.SupportURL != nil {
		addField("support_url", *req.SupportURL)
	}
	if req.PrivacyPolicyURL != nil {
		addField("privacy_policy_url", *req.PrivacyPolicyURL)
	}
	if req.TermsOfServiceURL != nil {
		addField("terms_of_service_url", *req.TermsOfServiceURL)
	}
	if req.ColorPrimary != nil {
		addField("color_primary", *req.ColorPrimary)
	}
	if req.ColorPrimaryHover != nil {
		addField("color_primary_hover", *req.ColorPrimaryHover)
	}
	if req.ColorSecondary != nil {
		addField("color_secondary", *req.ColorSecondary)
	}
	if req.ColorSecondaryHover != nil {
		addField("color_secondary_hover", *req.ColorSecondaryHover)
	}
	if req.ColorAccent != nil {
		addField("color_accent", *req.ColorAccent)
	}
	if req.ColorBackground != nil {
		addField("color_background", *req.ColorBackground)
	}
	if req.ColorSurface != nil {
		addField("color_surface", *req.ColorSurface)
	}
	if req.ColorTextPrimary != nil {
		addField("color_text_primary", *req.ColorTextPrimary)
	}
	if req.ColorTextSecondary != nil {
		addField("color_text_secondary", *req.ColorTextSecondary)
	}
	if req.ColorBorder != nil {
		addField("color_border", *req.ColorBorder)
	}
	if req.ColorSuccess != nil {
		addField("color_success", *req.ColorSuccess)
	}
	if req.ColorWarning != nil {
		addField("color_warning", *req.ColorWarning)
	}
	if req.ColorError != nil {
		addField("color_error", *req.ColorError)
	}
	if req.ColorInfo != nil {
		addField("color_info", *req.ColorInfo)
	}
	if req.ColorSidebarBg != nil {
		addField("color_sidebar_bg", *req.ColorSidebarBg)
	}
	if req.ColorSidebarText != nil {
		addField("color_sidebar_text", *req.ColorSidebarText)
	}
	if req.FontFamilyHeading != nil {
		addField("font_family_heading", *req.FontFamilyHeading)
	}
	if req.FontFamilyBody != nil {
		addField("font_family_body", *req.FontFamilyBody)
	}
	if req.FontSizeBase != nil {
		addField("font_size_base", *req.FontSizeBase)
	}
	if req.SidebarStyle != nil {
		addField("sidebar_style", *req.SidebarStyle)
	}
	if req.CornerRadius != nil {
		addField("corner_radius", *req.CornerRadius)
	}
	if req.Density != nil {
		addField("density", *req.Density)
	}
	if req.CustomCSS != nil {
		addField("custom_css", *req.CustomCSS)
	}
	if req.ShowPoweredBy != nil {
		addField("show_powered_by", *req.ShowPoweredBy)
	}
	if req.ShowHelpWidget != nil {
		addField("show_help_widget", *req.ShowHelpWidget)
	}
	if req.ShowMarketplace != nil {
		addField("show_marketplace", *req.ShowMarketplace)
	}
	if req.ShowKnowledgeBase != nil {
		addField("show_knowledge_base", *req.ShowKnowledgeBase)
	}

	if len(setClauses) == 0 {
		// Nothing to update, just return current branding
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}
		return s.GetBranding(ctx, orgID)
	}

	query := fmt.Sprintf("UPDATE tenant_branding SET %s WHERE organization_id = $1",
		strings.Join(setClauses, ", "))

	if _, err := tx.Exec(ctx, query, args...); err != nil {
		return nil, fmt.Errorf("update branding: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	log.Info().Str("org_id", orgID.String()).Msg("branding updated")
	return s.GetBranding(ctx, orgID)
}

// ============================================================
// UPLOAD LOGO
// ============================================================

// AllowedLogoTypes lists the valid logo_type values.
var AllowedLogoTypes = map[string]bool{
	"full": true, "icon": true, "dark": true,
	"light": true, "favicon": true, "email": true,
}

// UploadLogo validates and stores a logo file.
func (s *BrandingService) UploadLogo(ctx context.Context, orgID uuid.UUID, logoType string, filename string, fileData io.Reader) (*TenantBranding, error) {
	if !AllowedLogoTypes[logoType] {
		return nil, fmt.Errorf("invalid logo type: %s; allowed: full, icon, dark, light, favicon, email", logoType)
	}

	ext := strings.ToLower(filepath.Ext(filename))
	if ext != ".svg" && ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
		return nil, fmt.Errorf("invalid file type: %s; allowed: .svg, .png, .jpg, .jpeg", ext)
	}

	// Storage path: orgs/{orgID}/branding/{logo_type}.{ext}
	dir := filepath.Join(s.storagePath, "orgs", orgID.String(), "branding")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create storage dir: %w", err)
	}

	storedFilename := fmt.Sprintf("%s%s", logoType, ext)
	fullPath := filepath.Join(dir, storedFilename)

	outFile, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("create file: %w", err)
	}
	defer outFile.Close()

	written, err := io.Copy(outFile, fileData)
	if err != nil {
		return nil, fmt.Errorf("write file: %w", err)
	}

	// 5 MB max
	if written > 5*1024*1024 {
		os.Remove(fullPath)
		return nil, fmt.Errorf("file too large: max 5MB")
	}

	logoURL := fmt.Sprintf("/storage/orgs/%s/branding/%s", orgID.String(), storedFilename)

	// Map logo type to DB column
	colMap := map[string]string{
		"full":    "logo_full_url",
		"icon":    "logo_icon_url",
		"dark":    "logo_dark_url",
		"light":   "logo_light_url",
		"favicon": "favicon_url",
		"email":   "email_logo_url",
	}

	col := colMap[logoType]

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return nil, fmt.Errorf("set org: %w", err)
	}

	// Upsert branding row
	_, err = tx.Exec(ctx, `
		INSERT INTO tenant_branding (organization_id)
		VALUES ($1)
		ON CONFLICT (organization_id) DO NOTHING
	`, orgID)
	if err != nil {
		return nil, fmt.Errorf("upsert branding: %w", err)
	}

	query := fmt.Sprintf("UPDATE tenant_branding SET %s = $2 WHERE organization_id = $1", col)
	if _, err := tx.Exec(ctx, query, orgID, logoURL); err != nil {
		return nil, fmt.Errorf("update logo: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	log.Info().Str("org_id", orgID.String()).Str("logo_type", logoType).Msg("logo uploaded")
	return s.GetBranding(ctx, orgID)
}

// DeleteLogo removes a logo file and clears the DB reference.
func (s *BrandingService) DeleteLogo(ctx context.Context, orgID uuid.UUID, logoType string) (*TenantBranding, error) {
	if !AllowedLogoTypes[logoType] {
		return nil, fmt.Errorf("invalid logo type: %s", logoType)
	}

	colMap := map[string]string{
		"full":    "logo_full_url",
		"icon":    "logo_icon_url",
		"dark":    "logo_dark_url",
		"light":   "logo_light_url",
		"favicon": "favicon_url",
		"email":   "email_logo_url",
	}
	col := colMap[logoType]

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return nil, fmt.Errorf("set org: %w", err)
	}

	// Get current URL to delete file
	var currentURL *string
	err = tx.QueryRow(ctx,
		fmt.Sprintf("SELECT %s FROM tenant_branding WHERE organization_id = $1", col),
		orgID,
	).Scan(&currentURL)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("get current logo: %w", err)
	}

	// Clear the DB column
	query := fmt.Sprintf("UPDATE tenant_branding SET %s = NULL WHERE organization_id = $1", col)
	if _, err := tx.Exec(ctx, query, orgID); err != nil {
		return nil, fmt.Errorf("clear logo: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	// Delete file from storage (best effort)
	if currentURL != nil && *currentURL != "" {
		relPath := strings.TrimPrefix(*currentURL, "/storage/")
		fullPath := filepath.Join(s.storagePath, relPath)
		os.Remove(fullPath)
	}

	log.Info().Str("org_id", orgID.String()).Str("logo_type", logoType).Msg("logo deleted")
	return s.GetBranding(ctx, orgID)
}

// ============================================================
// CUSTOM DOMAIN VERIFICATION
// ============================================================

// VerifyCustomDomain performs DNS lookup to verify a custom domain.
func (s *BrandingService) VerifyCustomDomain(ctx context.Context, orgID uuid.UUID, domain string) (*DomainVerificationResult, error) {
	if domain == "" {
		return nil, fmt.Errorf("domain is required")
	}

	// Basic domain validation
	domain = strings.ToLower(strings.TrimSpace(domain))
	if len(domain) > 253 {
		return nil, fmt.Errorf("domain too long (max 253 characters)")
	}

	domainRegex := regexp.MustCompile(`^([a-z0-9]([a-z0-9-]*[a-z0-9])?\.)+[a-z]{2,}$`)
	if !domainRegex.MatchString(domain) {
		return nil, fmt.Errorf("invalid domain format")
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return nil, fmt.Errorf("set org: %w", err)
	}

	// Upsert and generate verification token
	token := uuid.New().String()[:16]
	_, err = tx.Exec(ctx, `
		INSERT INTO tenant_branding (organization_id, custom_domain, domain_verification_token, domain_verification_status)
		VALUES ($1, $2, $3, 'pending')
		ON CONFLICT (organization_id)
		DO UPDATE SET
			custom_domain = $2,
			domain_verification_token = $3,
			domain_verification_status = 'pending',
			domain_verified_at = NULL,
			ssl_status = 'pending'
	`, orgID, domain, token)
	if err != nil {
		return nil, fmt.Errorf("set domain: %w", err)
	}

	result := &DomainVerificationResult{
		Domain: domain,
		Token:  token,
		Status: "pending",
	}

	// Verify CNAME record
	cnames, err := net.LookupCNAME(domain)
	if err == nil && strings.Contains(cnames, "complianceforge") {
		result.CNAMEValid = true
	}

	// Verify TXT record
	txtRecords, err := net.LookupTXT(domain)
	if err == nil {
		for _, txt := range txtRecords {
			if strings.Contains(txt, token) {
				result.TXTValid = true
				break
			}
		}
	}

	if result.CNAMEValid || result.TXTValid {
		now := time.Now()
		result.Status = "verified"
		_, err = tx.Exec(ctx, `
			UPDATE tenant_branding
			SET domain_verification_status = 'verified',
				domain_verified_at = $2,
				ssl_status = 'provisioning'
			WHERE organization_id = $1
		`, orgID, now)
		if err != nil {
			return nil, fmt.Errorf("update verified status: %w", err)
		}
		result.SSLStatus = "provisioning"
		result.Message = "Domain verified successfully. SSL certificate is being provisioned."
	} else {
		_, err = tx.Exec(ctx, `
			UPDATE tenant_branding
			SET domain_verification_status = 'dns_configured'
			WHERE organization_id = $1
		`, orgID)
		if err != nil {
			return nil, fmt.Errorf("update dns status: %w", err)
		}
		result.Status = "dns_configured"
		result.SSLStatus = "pending"
		result.Message = fmt.Sprintf(
			"Add a CNAME record pointing %s to app.complianceforge.io OR add a TXT record with value: cf-verify=%s",
			domain, token,
		)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return result, nil
}

// GetDomainStatus returns the current custom domain and SSL status.
func (s *BrandingService) GetDomainStatus(ctx context.Context, orgID uuid.UUID) (*DomainStatusResponse, error) {
	branding, err := s.GetBranding(ctx, orgID)
	if err != nil {
		return nil, err
	}

	return &DomainStatusResponse{
		CustomDomain:             branding.CustomDomain,
		DomainVerificationStatus: branding.DomainVerificationStatus,
		DomainVerifiedAt:         branding.DomainVerifiedAt,
		SSLStatus:                branding.SSLStatus,
		SSLProvisionedAt:         branding.SSLProvisionedAt,
		SSLExpiresAt:             branding.SSLExpiresAt,
		VerificationToken:        branding.DomainVerificationToken,
	}, nil
}

// ============================================================
// CSS GENERATION
// ============================================================

// GetBrandingCSS generates CSS custom properties from the branding configuration.
func (s *BrandingService) GetBrandingCSS(ctx context.Context, orgID uuid.UUID) (string, error) {
	branding, err := s.GetBranding(ctx, orgID)
	if err != nil {
		return "", err
	}

	return GenerateBrandingCSS(branding), nil
}

// GenerateBrandingCSS builds a CSS string from a TenantBranding config.
func GenerateBrandingCSS(b *TenantBranding) string {
	var sb strings.Builder

	sb.WriteString("/* ComplianceForge Branding — Auto-generated */\n")
	sb.WriteString(":root {\n")

	// Colours
	writeVar(&sb, "--cf-color-primary", b.ColorPrimary)
	writeVar(&sb, "--cf-color-primary-hover", b.ColorPrimaryHover)
	writeVar(&sb, "--cf-color-secondary", b.ColorSecondary)
	writeVar(&sb, "--cf-color-secondary-hover", b.ColorSecondaryHover)
	writeVar(&sb, "--cf-color-accent", b.ColorAccent)
	writeVar(&sb, "--cf-color-background", b.ColorBackground)
	writeVar(&sb, "--cf-color-surface", b.ColorSurface)
	writeVar(&sb, "--cf-color-text-primary", b.ColorTextPrimary)
	writeVar(&sb, "--cf-color-text-secondary", b.ColorTextSecondary)
	writeVar(&sb, "--cf-color-border", b.ColorBorder)
	writeVar(&sb, "--cf-color-success", b.ColorSuccess)
	writeVar(&sb, "--cf-color-warning", b.ColorWarning)
	writeVar(&sb, "--cf-color-error", b.ColorError)
	writeVar(&sb, "--cf-color-info", b.ColorInfo)
	writeVar(&sb, "--cf-color-sidebar-bg", b.ColorSidebarBg)
	writeVar(&sb, "--cf-color-sidebar-text", b.ColorSidebarText)

	// Typography
	writeVar(&sb, "--cf-font-heading", b.FontFamilyHeading)
	writeVar(&sb, "--cf-font-body", b.FontFamilyBody)
	writeVar(&sb, "--cf-font-size-base", b.FontSizeBase)

	// Corner radius mapping
	radiusMap := map[string]string{
		"none":   "0px",
		"small":  "4px",
		"medium": "8px",
		"large":  "12px",
		"full":   "9999px",
	}
	radius := radiusMap[b.CornerRadius]
	if radius == "" {
		radius = "8px"
	}
	writeVar(&sb, "--cf-border-radius", radius)

	// Density spacing
	densityMap := map[string]string{
		"compact":     "0.75rem",
		"comfortable": "1rem",
		"spacious":    "1.5rem",
	}
	spacing := densityMap[b.Density]
	if spacing == "" {
		spacing = "1rem"
	}
	writeVar(&sb, "--cf-spacing-unit", spacing)

	sb.WriteString("}\n")

	// Append sanitized custom CSS
	if b.CustomCSS != "" {
		sb.WriteString("\n/* Custom CSS */\n")
		sb.WriteString(b.CustomCSS)
		sb.WriteString("\n")
	}

	return sb.String()
}

func writeVar(sb *strings.Builder, name, value string) {
	if value != "" {
		sb.WriteString(fmt.Sprintf("  %s: %s;\n", name, value))
	}
}

// ============================================================
// EMAIL BRANDING
// ============================================================

// ApplyBrandingToEmail replaces branding placeholders in an email HTML template.
func (s *BrandingService) ApplyBrandingToEmail(ctx context.Context, orgID uuid.UUID, emailHTML string) (string, error) {
	branding, err := s.GetBranding(ctx, orgID)
	if err != nil {
		return emailHTML, err
	}

	replacer := strings.NewReplacer(
		"{{product_name}}", branding.ProductName,
		"{{company_name}}", branding.CompanyName,
		"{{support_email}}", branding.SupportEmail,
		"{{support_url}}", branding.SupportURL,
		"{{privacy_policy_url}}", branding.PrivacyPolicyURL,
		"{{terms_of_service_url}}", branding.TermsOfServiceURL,
		"{{logo_url}}", branding.EmailLogoURL,
		"{{logo_full_url}}", branding.LogoFullURL,
		"{{color_primary}}", branding.ColorPrimary,
		"{{color_secondary}}", branding.ColorSecondary,
		"{{color_background}}", branding.ColorBackground,
		"{{color_text_primary}}", branding.ColorTextPrimary,
	)

	result := replacer.Replace(emailHTML)

	// Add powered-by footer if required
	if branding.ShowPoweredBy {
		poweredBy := `<p style="text-align:center;color:#999;font-size:11px;margin-top:20px;">Powered by ComplianceForge</p>`
		result = strings.Replace(result, "{{powered_by}}", poweredBy, 1)
	} else {
		result = strings.Replace(result, "{{powered_by}}", "", 1)
	}

	return result, nil
}

// ============================================================
// PDF BRANDING
// ============================================================

// ApplyBrandingToPDF returns a PDF branding configuration for report generation.
func (s *BrandingService) ApplyBrandingToPDF(ctx context.Context, orgID uuid.UUID) (*PDFBrandingConfig, error) {
	branding, err := s.GetBranding(ctx, orgID)
	if err != nil {
		return nil, err
	}

	logoURL := branding.LogoFullURL
	if logoURL == "" {
		logoURL = branding.LogoDarkURL
	}

	return &PDFBrandingConfig{
		ProductName:      branding.ProductName,
		CompanyName:      branding.CompanyName,
		LogoURL:          logoURL,
		ColorPrimary:     branding.ColorPrimary,
		ColorSecondary:   branding.ColorSecondary,
		FontFamily:       branding.FontFamilyBody,
		ShowPoweredBy:    branding.ShowPoweredBy,
		PrivacyPolicyURL: branding.PrivacyPolicyURL,
	}, nil
}

// ============================================================
// PARTNER MANAGEMENT
// ============================================================

// ListPartners returns all white-label partners (super admin).
func (s *BrandingService) ListPartners(ctx context.Context) ([]WhiteLabelPartner, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, partner_name, partner_slug, contact_email,
			default_branding_id, revenue_share_percent, max_tenants, is_active,
			created_at, updated_at
		FROM white_label_partners
		ORDER BY partner_name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list partners: %w", err)
	}
	defer rows.Close()

	var partners []WhiteLabelPartner
	for rows.Next() {
		var p WhiteLabelPartner
		if err := rows.Scan(
			&p.ID, &p.PartnerName, &p.PartnerSlug, &p.ContactEmail,
			&p.DefaultBrandingID, &p.RevenueSharePercent, &p.MaxTenants, &p.IsActive,
			&p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan partner: %w", err)
		}
		partners = append(partners, p)
	}

	if partners == nil {
		partners = []WhiteLabelPartner{}
	}
	return partners, nil
}

// CreatePartner creates a new white-label partner.
func (s *BrandingService) CreatePartner(ctx context.Context, req CreatePartnerRequest) (*WhiteLabelPartner, error) {
	if req.PartnerName == "" {
		return nil, fmt.Errorf("partner_name is required")
	}
	if req.PartnerSlug == "" {
		return nil, fmt.Errorf("partner_slug is required")
	}
	if req.ContactEmail == "" {
		return nil, fmt.Errorf("contact_email is required")
	}

	slugRegex := regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$`)
	if !slugRegex.MatchString(req.PartnerSlug) {
		return nil, fmt.Errorf("partner_slug must be lowercase alphanumeric with hyphens")
	}

	if req.RevenueSharePercent < 0 || req.RevenueSharePercent > 100 {
		return nil, fmt.Errorf("revenue_share_percent must be between 0 and 100")
	}

	if req.MaxTenants <= 0 {
		req.MaxTenants = 100
	}

	var p WhiteLabelPartner
	err := s.pool.QueryRow(ctx, `
		INSERT INTO white_label_partners (partner_name, partner_slug, contact_email, revenue_share_percent, max_tenants)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, partner_name, partner_slug, contact_email,
			default_branding_id, revenue_share_percent, max_tenants, is_active,
			created_at, updated_at
	`, req.PartnerName, req.PartnerSlug, req.ContactEmail, req.RevenueSharePercent, req.MaxTenants).Scan(
		&p.ID, &p.PartnerName, &p.PartnerSlug, &p.ContactEmail,
		&p.DefaultBrandingID, &p.RevenueSharePercent, &p.MaxTenants, &p.IsActive,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return nil, fmt.Errorf("partner_slug already exists")
		}
		return nil, fmt.Errorf("create partner: %w", err)
	}

	log.Info().Str("partner_id", p.ID.String()).Str("slug", p.PartnerSlug).Msg("partner created")
	return &p, nil
}

// UpdatePartner updates a white-label partner.
func (s *BrandingService) UpdatePartner(ctx context.Context, partnerID uuid.UUID, req UpdatePartnerRequest) (*WhiteLabelPartner, error) {
	setClauses := []string{}
	args := []interface{}{partnerID}
	argIdx := 2

	addField := func(col string, val interface{}) {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", col, argIdx))
		args = append(args, val)
		argIdx++
	}

	if req.PartnerName != nil {
		addField("partner_name", *req.PartnerName)
	}
	if req.ContactEmail != nil {
		addField("contact_email", *req.ContactEmail)
	}
	if req.RevenueSharePercent != nil {
		if *req.RevenueSharePercent < 0 || *req.RevenueSharePercent > 100 {
			return nil, fmt.Errorf("revenue_share_percent must be between 0 and 100")
		}
		addField("revenue_share_percent", *req.RevenueSharePercent)
	}
	if req.MaxTenants != nil {
		addField("max_tenants", *req.MaxTenants)
	}
	if req.IsActive != nil {
		addField("is_active", *req.IsActive)
	}
	if req.DefaultBrandingID != nil {
		if *req.DefaultBrandingID == "" {
			addField("default_branding_id", nil)
		} else {
			bid, err := uuid.Parse(*req.DefaultBrandingID)
			if err != nil {
				return nil, fmt.Errorf("invalid default_branding_id")
			}
			addField("default_branding_id", bid)
		}
	}

	if len(setClauses) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}

	query := fmt.Sprintf("UPDATE white_label_partners SET %s WHERE id = $1 RETURNING id, partner_name, partner_slug, contact_email, default_branding_id, revenue_share_percent, max_tenants, is_active, created_at, updated_at",
		strings.Join(setClauses, ", "))

	var p WhiteLabelPartner
	err := s.pool.QueryRow(ctx, query, args...).Scan(
		&p.ID, &p.PartnerName, &p.PartnerSlug, &p.ContactEmail,
		&p.DefaultBrandingID, &p.RevenueSharePercent, &p.MaxTenants, &p.IsActive,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("partner not found")
		}
		return nil, fmt.Errorf("update partner: %w", err)
	}

	log.Info().Str("partner_id", p.ID.String()).Msg("partner updated")
	return &p, nil
}

// GetPartnerTenants returns all tenant mappings for a partner.
func (s *BrandingService) GetPartnerTenants(ctx context.Context, partnerID uuid.UUID) ([]PartnerTenantMapping, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT ptm.id, ptm.partner_id, ptm.organization_id, ptm.onboarded_at, ptm.created_at
		FROM partner_tenant_mappings ptm
		WHERE ptm.partner_id = $1
		ORDER BY ptm.onboarded_at DESC
	`, partnerID)
	if err != nil {
		return nil, fmt.Errorf("list partner tenants: %w", err)
	}
	defer rows.Close()

	var mappings []PartnerTenantMapping
	for rows.Next() {
		var m PartnerTenantMapping
		if err := rows.Scan(&m.ID, &m.PartnerID, &m.OrganizationID, &m.OnboardedAt, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan mapping: %w", err)
		}
		mappings = append(mappings, m)
	}

	if mappings == nil {
		mappings = []PartnerTenantMapping{}
	}
	return mappings, nil
}

// ============================================================
// VALIDATION HELPERS (exported for testing)
// ============================================================

// hexColourRegex matches a 7-character hex colour code (#XXXXXX).
var hexColourRegex = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

// ValidateHexColour checks if a string is a valid 7-character hex colour code.
func ValidateHexColour(colour string) error {
	if !hexColourRegex.MatchString(colour) {
		return fmt.Errorf("must be a 7-character hex colour code (e.g. #1A2B3C), got: %s", colour)
	}
	return nil
}

// validateURL checks if a string is a valid HTTP/HTTPS URL.
func validateURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("URL must use http or https scheme")
	}
	if u.Host == "" {
		return fmt.Errorf("URL must have a host")
	}
	return nil
}

// isValidEnum checks if a value is in an allowed list.
func isValidEnum(val string, allowed []string) bool {
	for _, a := range allowed {
		if val == a {
			return true
		}
	}
	return false
}

// ============================================================
// CSS SANITIZATION
// ============================================================

// safeProperties is the whitelist of CSS properties allowed in custom CSS.
var safeProperties = map[string]bool{
	"color": true, "background": true, "background-color": true,
	"background-image": true, "background-size": true, "background-position": true,
	"background-repeat": true,
	"font":              true, "font-family": true, "font-size": true, "font-weight": true,
	"font-style": true, "font-variant": true,
	"line-height": true, "letter-spacing": true, "word-spacing": true,
	"text-align": true, "text-decoration": true, "text-transform": true,
	"text-indent": true, "text-shadow": true,
	"border": true, "border-color": true, "border-width": true, "border-style": true,
	"border-radius": true,
	"border-top":    true, "border-right": true, "border-bottom": true, "border-left": true,
	"border-top-color": true, "border-right-color": true, "border-bottom-color": true, "border-left-color": true,
	"border-top-width": true, "border-right-width": true, "border-bottom-width": true, "border-left-width": true,
	"border-top-left-radius": true, "border-top-right-radius": true,
	"border-bottom-left-radius": true, "border-bottom-right-radius": true,
	"margin": true, "margin-top": true, "margin-right": true, "margin-bottom": true, "margin-left": true,
	"padding": true, "padding-top": true, "padding-right": true, "padding-bottom": true, "padding-left": true,
	"width": true, "max-width": true, "min-width": true,
	"height": true, "max-height": true, "min-height": true,
	"display": true, "visibility": true, "opacity": true,
	"overflow": true, "overflow-x": true, "overflow-y": true,
	"box-shadow": true, "outline": true, "outline-color": true,
	"cursor": true, "white-space": true, "list-style": true,
	"gap": true, "row-gap": true, "column-gap": true,
	"flex": true, "flex-direction": true, "flex-wrap": true,
	"align-items": true, "justify-content": true,
	"grid-template-columns": true, "grid-template-rows": true,
	"transition": true, "transform": true,
}

// dangerousPatterns are regex patterns that indicate malicious CSS.
var dangerousPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)javascript\s*:`),
	regexp.MustCompile(`(?i)expression\s*\(`),
	regexp.MustCompile(`(?i)url\s*\(\s*["']?\s*data\s*:`),
	regexp.MustCompile(`(?i)@import`),
	regexp.MustCompile(`(?i)behavior\s*:`),
	regexp.MustCompile(`(?i)-moz-binding`),
	regexp.MustCompile(`(?i)\\00`),
	regexp.MustCompile(`(?i)<\s*script`),
	regexp.MustCompile(`(?i)<\s*/\s*script`),
	regexp.MustCompile(`(?i)on\w+\s*=`),
}

// SanitizeCSS removes unsafe CSS properties and patterns from a custom CSS string.
// It whitelists safe properties and strips anything potentially dangerous.
func SanitizeCSS(css string) string {
	// First: reject any dangerous patterns entirely
	for _, pattern := range dangerousPatterns {
		css = pattern.ReplaceAllString(css, "/* blocked */")
	}

	// Block url() with data: URIs but allow safe url() usage
	dataURIRegex := regexp.MustCompile(`(?i)url\s*\(\s*["']?\s*data\s*:`)
	css = dataURIRegex.ReplaceAllString(css, "/* blocked-url */")

	// Parse CSS rules and filter by property whitelist
	var result strings.Builder
	lines := strings.Split(css, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Pass through selectors, braces, comments
		if trimmed == "" || trimmed == "}" || strings.HasPrefix(trimmed, "/*") ||
			strings.HasSuffix(trimmed, "{") || strings.HasPrefix(trimmed, ":root") ||
			strings.HasPrefix(trimmed, ".") || strings.HasPrefix(trimmed, "#") ||
			strings.HasPrefix(trimmed, "@media") || strings.HasPrefix(trimmed, "@keyframes") {
			result.WriteString(line)
			result.WriteString("\n")
			continue
		}

		// Check if this is a property declaration
		if strings.Contains(trimmed, ":") && !strings.HasSuffix(trimmed, "{") {
			propParts := strings.SplitN(trimmed, ":", 2)
			prop := strings.TrimSpace(propParts[0])
			prop = strings.TrimLeft(prop, " \t")

			// CSS custom properties (--cf-*) are always allowed
			if strings.HasPrefix(prop, "--") {
				result.WriteString(line)
				result.WriteString("\n")
				continue
			}

			if safeProperties[prop] {
				result.WriteString(line)
				result.WriteString("\n")
			} else {
				result.WriteString(fmt.Sprintf("  /* blocked: %s */\n", prop))
			}
		} else {
			result.WriteString(line)
			result.WriteString("\n")
		}
	}

	return strings.TrimSpace(result.String())
}
