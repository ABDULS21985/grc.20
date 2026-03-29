package service

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

// ============================================================
// HEX COLOUR VALIDATION TESTS
// ============================================================

func TestValidateHexColour(t *testing.T) {
	tests := []struct {
		name        string
		colour      string
		expectError bool
	}{
		{"valid black", "#000000", false},
		{"valid white", "#FFFFFF", false},
		{"valid mixed case", "#aAbBcC", false},
		{"valid primary indigo", "#4F46E5", false},
		{"valid with lowercase", "#1a2b3c", false},
		{"missing hash", "000000", true},
		{"too short", "#FFF", true},
		{"too long", "#1234567", true},
		{"invalid char g", "#GGGGGG", true},
		{"empty string", "", true},
		{"only hash", "#", true},
		{"with spaces", "# 12345", true},
		{"rgb format", "rgb(0,0,0)", true},
		{"double hash", "##123456", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateHexColour(tt.colour)
			if tt.expectError && err == nil {
				t.Errorf("Expected error for colour %q but got nil", tt.colour)
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error for colour %q but got: %v", tt.colour, err)
			}
		})
	}
}

// ============================================================
// CSS GENERATION TESTS
// ============================================================

func TestGenerateBrandingCSS(t *testing.T) {
	orgID := uuid.New()
	branding := DefaultBranding(orgID)

	css := GenerateBrandingCSS(&branding)

	// Verify it contains the :root declaration
	if !strings.Contains(css, ":root {") {
		t.Error("CSS should contain :root declaration")
	}

	// Verify CSS variables are present
	expectedVars := []string{
		"--cf-color-primary: #4F46E5",
		"--cf-color-primary-hover: #4338CA",
		"--cf-color-secondary: #7C3AED",
		"--cf-color-secondary-hover: #6D28D9",
		"--cf-color-accent: #06B6D4",
		"--cf-color-background: #F9FAFB",
		"--cf-color-surface: #FFFFFF",
		"--cf-color-text-primary: #111827",
		"--cf-color-text-secondary: #6B7280",
		"--cf-color-border: #E5E7EB",
		"--cf-color-success: #10B981",
		"--cf-color-warning: #F59E0B",
		"--cf-color-error: #EF4444",
		"--cf-color-info: #3B82F6",
		"--cf-color-sidebar-bg: #1F2937",
		"--cf-color-sidebar-text: #F9FAFB",
		"--cf-font-heading: Inter",
		"--cf-font-body: Inter",
		"--cf-font-size-base: 14px",
		"--cf-border-radius: 8px",
		"--cf-spacing-unit: 1rem",
	}

	for _, v := range expectedVars {
		if !strings.Contains(css, v) {
			t.Errorf("CSS should contain variable %q", v)
		}
	}
}

func TestGenerateBrandingCSSCustomValues(t *testing.T) {
	branding := &TenantBranding{
		ColorPrimary:        "#FF0000",
		ColorPrimaryHover:   "#CC0000",
		ColorSecondary:      "#00FF00",
		ColorSecondaryHover: "#00CC00",
		ColorAccent:         "#0000FF",
		ColorBackground:     "#FAFAFA",
		ColorSurface:        "#FEFEFE",
		ColorTextPrimary:    "#222222",
		ColorTextSecondary:  "#888888",
		ColorBorder:         "#DDDDDD",
		ColorSuccess:        "#00AA00",
		ColorWarning:        "#FFAA00",
		ColorError:          "#FF0000",
		ColorInfo:           "#0088FF",
		ColorSidebarBg:      "#333333",
		ColorSidebarText:    "#EEEEEE",
		FontFamilyHeading:   "Roboto",
		FontFamilyBody:      "Open Sans",
		FontSizeBase:        "16px",
		CornerRadius:        "large",
		Density:             "spacious",
	}

	css := GenerateBrandingCSS(branding)

	if !strings.Contains(css, "--cf-color-primary: #FF0000") {
		t.Error("CSS should contain custom primary colour")
	}
	if !strings.Contains(css, "--cf-font-heading: Roboto") {
		t.Error("CSS should contain custom heading font")
	}
	if !strings.Contains(css, "--cf-font-body: Open Sans") {
		t.Error("CSS should contain custom body font")
	}
	if !strings.Contains(css, "--cf-border-radius: 12px") {
		t.Error("CSS should map 'large' corner_radius to 12px")
	}
	if !strings.Contains(css, "--cf-spacing-unit: 1.5rem") {
		t.Error("CSS should map 'spacious' density to 1.5rem")
	}
}

func TestGenerateBrandingCSSWithCustomCSS(t *testing.T) {
	branding := DefaultBranding(uuid.New())
	branding.CustomCSS = ".custom-header { color: red; }"

	css := GenerateBrandingCSS(&branding)

	if !strings.Contains(css, "/* Custom CSS */") {
		t.Error("CSS should contain custom CSS section marker")
	}
	if !strings.Contains(css, ".custom-header { color: red; }") {
		t.Error("CSS should contain the custom CSS content")
	}
}

func TestGenerateBrandingCSSCornerRadiusMapping(t *testing.T) {
	tests := []struct {
		radius   string
		expected string
	}{
		{"none", "0px"},
		{"small", "4px"},
		{"medium", "8px"},
		{"large", "12px"},
		{"full", "9999px"},
		{"unknown", "8px"}, // default fallback
	}

	for _, tt := range tests {
		t.Run(tt.radius, func(t *testing.T) {
			branding := DefaultBranding(uuid.New())
			branding.CornerRadius = tt.radius
			css := GenerateBrandingCSS(&branding)

			expectedVar := "--cf-border-radius: " + tt.expected
			if !strings.Contains(css, expectedVar) {
				t.Errorf("Expected %q in CSS for corner_radius %q", expectedVar, tt.radius)
			}
		})
	}
}

func TestGenerateBrandingCSSDensityMapping(t *testing.T) {
	tests := []struct {
		density  string
		expected string
	}{
		{"compact", "0.75rem"},
		{"comfortable", "1rem"},
		{"spacious", "1.5rem"},
		{"unknown", "1rem"}, // default fallback
	}

	for _, tt := range tests {
		t.Run(tt.density, func(t *testing.T) {
			branding := DefaultBranding(uuid.New())
			branding.Density = tt.density
			css := GenerateBrandingCSS(&branding)

			expectedVar := "--cf-spacing-unit: " + tt.expected
			if !strings.Contains(css, expectedVar) {
				t.Errorf("Expected %q in CSS for density %q", expectedVar, tt.density)
			}
		})
	}
}

// ============================================================
// CSS SANITIZATION TESTS
// ============================================================

func TestSanitizeCSSAllowsSafeProperties(t *testing.T) {
	safeCSS := `.header {
  color: #333;
  background-color: #fff;
  font-size: 16px;
  border: 1px solid #ddd;
  margin: 10px;
  padding: 20px;
  border-radius: 8px;
}`

	result := SanitizeCSS(safeCSS)

	// All properties should be preserved
	if !strings.Contains(result, "color: #333") {
		t.Error("color property should be allowed")
	}
	if !strings.Contains(result, "background-color: #fff") {
		t.Error("background-color property should be allowed")
	}
	if !strings.Contains(result, "font-size: 16px") {
		t.Error("font-size property should be allowed")
	}
	if !strings.Contains(result, "border: 1px solid #ddd") {
		t.Error("border property should be allowed")
	}
	if !strings.Contains(result, "margin: 10px") {
		t.Error("margin property should be allowed")
	}
	if !strings.Contains(result, "padding: 20px") {
		t.Error("padding property should be allowed")
	}
}

func TestSanitizeCSSBlocksJavaScript(t *testing.T) {
	maliciousCSS := `.evil {
  background: javascript:alert(1);
}`

	result := SanitizeCSS(maliciousCSS)

	if strings.Contains(result, "javascript:") {
		t.Error("javascript: should be blocked")
	}
	if !strings.Contains(result, "/* blocked */") {
		t.Error("Should contain blocked comment")
	}
}

func TestSanitizeCSSBlocksExpression(t *testing.T) {
	maliciousCSS := `.evil {
  width: expression(document.body.clientWidth);
}`

	result := SanitizeCSS(maliciousCSS)

	if strings.Contains(result, "expression(") {
		t.Error("expression() should be blocked")
	}
}

func TestSanitizeCSSBlocksDataURIs(t *testing.T) {
	maliciousCSS := `.evil {
  background-image: url(data:text/html,<script>alert(1)</script>);
}`

	result := SanitizeCSS(maliciousCSS)

	if strings.Contains(result, "data:text/html") {
		t.Error("data: URIs should be blocked")
	}
}

func TestSanitizeCSSBlocksImport(t *testing.T) {
	maliciousCSS := `@import url('https://evil.com/steal.css');
.normal {
  color: red;
}`

	result := SanitizeCSS(maliciousCSS)

	if strings.Contains(result, "@import") {
		t.Error("@import should be blocked")
	}
}

func TestSanitizeCSSBlocksBehavior(t *testing.T) {
	maliciousCSS := `.evil {
  behavior: url(evil.htc);
}`

	result := SanitizeCSS(maliciousCSS)

	if strings.Contains(result, "behavior:") {
		t.Error("behavior: should be blocked")
	}
}

func TestSanitizeCSSBlocksMozBinding(t *testing.T) {
	maliciousCSS := `.evil {
  -moz-binding: url("evil.xml#xbl");
}`

	result := SanitizeCSS(maliciousCSS)

	if strings.Contains(result, "-moz-binding") {
		t.Error("-moz-binding should be blocked")
	}
}

func TestSanitizeCSSBlocksScriptTags(t *testing.T) {
	maliciousCSS := `</style><script>alert(1)</script><style>
.normal { color: red; }`

	result := SanitizeCSS(maliciousCSS)

	if strings.Contains(result, "<script>") {
		t.Error("<script> tags should be blocked")
	}
}

func TestSanitizeCSSAllowsCSSCustomProperties(t *testing.T) {
	customPropCSS := `:root {
  --cf-custom-color: #FF0000;
  --my-spacing: 16px;
}`

	result := SanitizeCSS(customPropCSS)

	if !strings.Contains(result, "--cf-custom-color: #FF0000") {
		t.Error("CSS custom properties (--*) should be allowed")
	}
	if !strings.Contains(result, "--my-spacing: 16px") {
		t.Error("CSS custom properties (--*) should be allowed")
	}
}

func TestSanitizeCSSBlocksUnsafeProperties(t *testing.T) {
	unsafeCSS := `.element {
  position: fixed;
  z-index: 99999;
  top: 0;
  left: 0;
}`

	result := SanitizeCSS(unsafeCSS)

	// position, z-index, top, left are NOT in the safe list
	if strings.Contains(result, "position: fixed") {
		t.Error("position property should be blocked")
	}
	if strings.Contains(result, "z-index: 99999") {
		t.Error("z-index property should be blocked")
	}
}

func TestSanitizeCSSPreservesSelectorsAndBraces(t *testing.T) {
	css := `.header {
  color: #333;
}

#sidebar {
  background: #f0f0f0;
}`

	result := SanitizeCSS(css)

	if !strings.Contains(result, ".header {") {
		t.Error("Class selectors should be preserved")
	}
	if !strings.Contains(result, "#sidebar {") {
		t.Error("ID selectors should be preserved")
	}
	if !strings.Contains(result, "}") {
		t.Error("Closing braces should be preserved")
	}
}

func TestSanitizeCSSPreservesMediaQueries(t *testing.T) {
	css := `@media (max-width: 768px) {
  .sidebar {
    display: none;
  }
}`

	result := SanitizeCSS(css)

	if !strings.Contains(result, "@media") {
		t.Error("@media queries should be preserved")
	}
}

func TestSanitizeCSSEmptyInput(t *testing.T) {
	result := SanitizeCSS("")
	if result != "" {
		t.Errorf("Empty input should produce empty output, got: %q", result)
	}
}

func TestSanitizeCSSBlocksEventHandlers(t *testing.T) {
	maliciousCSS := `div[onclick=alert(1)] {
  color: red;
}`

	result := SanitizeCSS(maliciousCSS)

	if strings.Contains(result, "onclick=") {
		t.Error("event handler attributes should be blocked")
	}
}

// ============================================================
// MODEL SERIALIZATION TESTS
// ============================================================

func TestTenantBrandingSerialization(t *testing.T) {
	now := time.Now()
	domainVerifiedAt := now.Add(-24 * time.Hour)

	branding := TenantBranding{
		ID:                       uuid.New(),
		OrganizationID:           uuid.New(),
		ProductName:              "MyGRC Platform",
		Tagline:                  "Enterprise Risk Management",
		CompanyName:              "Acme Corp",
		SupportEmail:             "support@acme.com",
		SupportURL:               "https://help.acme.com",
		PrivacyPolicyURL:         "https://acme.com/privacy",
		TermsOfServiceURL:        "https://acme.com/terms",
		LogoFullURL:              "/storage/orgs/123/branding/full.svg",
		LogoIconURL:              "/storage/orgs/123/branding/icon.png",
		ColorPrimary:             "#FF6600",
		ColorPrimaryHover:        "#E65C00",
		ColorSecondary:           "#0066FF",
		ColorSecondaryHover:      "#005CE6",
		ColorAccent:              "#00CCFF",
		ColorBackground:          "#FAFAFA",
		ColorSurface:             "#FFFFFF",
		ColorTextPrimary:         "#1A1A1A",
		ColorTextSecondary:       "#666666",
		ColorBorder:              "#DDDDDD",
		ColorSuccess:             "#00CC00",
		ColorWarning:             "#FFAA00",
		ColorError:               "#FF0000",
		ColorInfo:                "#0088FF",
		ColorSidebarBg:           "#2D2D2D",
		ColorSidebarText:         "#EFEFEF",
		FontFamilyHeading:        "Montserrat",
		FontFamilyBody:           "Lato",
		FontSizeBase:             "15px",
		SidebarStyle:             "branded",
		CornerRadius:             "large",
		Density:                  "spacious",
		CustomDomain:             "grc.acme.com",
		DomainVerificationToken:  "abc123def456",
		DomainVerificationStatus: "verified",
		DomainVerifiedAt:         &domainVerifiedAt,
		SSLStatus:                "active",
		CustomCSS:                ".header { color: red; }",
		ShowPoweredBy:            false,
		ShowHelpWidget:           true,
		ShowMarketplace:          false,
		ShowKnowledgeBase:        true,
		CreatedAt:                now,
		UpdatedAt:                now,
	}

	data, err := json.Marshal(branding)
	if err != nil {
		t.Fatalf("Failed to marshal TenantBranding: %v", err)
	}

	var decoded TenantBranding
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal TenantBranding: %v", err)
	}

	if decoded.ProductName != "MyGRC Platform" {
		t.Errorf("Expected product_name 'MyGRC Platform', got '%s'", decoded.ProductName)
	}
	if decoded.CompanyName != "Acme Corp" {
		t.Errorf("Expected company_name 'Acme Corp', got '%s'", decoded.CompanyName)
	}
	if decoded.ColorPrimary != "#FF6600" {
		t.Errorf("Expected color_primary '#FF6600', got '%s'", decoded.ColorPrimary)
	}
	if decoded.FontFamilyHeading != "Montserrat" {
		t.Errorf("Expected font_family_heading 'Montserrat', got '%s'", decoded.FontFamilyHeading)
	}
	if decoded.SidebarStyle != "branded" {
		t.Errorf("Expected sidebar_style 'branded', got '%s'", decoded.SidebarStyle)
	}
	if decoded.CornerRadius != "large" {
		t.Errorf("Expected corner_radius 'large', got '%s'", decoded.CornerRadius)
	}
	if decoded.CustomDomain != "grc.acme.com" {
		t.Errorf("Expected custom_domain 'grc.acme.com', got '%s'", decoded.CustomDomain)
	}
	if decoded.DomainVerificationStatus != "verified" {
		t.Errorf("Expected domain_verification_status 'verified', got '%s'", decoded.DomainVerificationStatus)
	}
	if decoded.ShowPoweredBy {
		t.Error("Expected show_powered_by to be false")
	}
	if !decoded.ShowHelpWidget {
		t.Error("Expected show_help_widget to be true")
	}
	if decoded.ShowMarketplace {
		t.Error("Expected show_marketplace to be false")
	}
}

func TestWhiteLabelPartnerSerialization(t *testing.T) {
	now := time.Now()
	brandingID := uuid.New()

	partner := WhiteLabelPartner{
		ID:                  uuid.New(),
		PartnerName:         "SecureCo Partners",
		PartnerSlug:         "secureco",
		ContactEmail:        "partners@secureco.com",
		DefaultBrandingID:   &brandingID,
		RevenueSharePercent: 30.50,
		MaxTenants:          250,
		IsActive:            true,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	data, err := json.Marshal(partner)
	if err != nil {
		t.Fatalf("Failed to marshal WhiteLabelPartner: %v", err)
	}

	var decoded WhiteLabelPartner
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal WhiteLabelPartner: %v", err)
	}

	if decoded.PartnerName != "SecureCo Partners" {
		t.Errorf("Expected partner_name 'SecureCo Partners', got '%s'", decoded.PartnerName)
	}
	if decoded.PartnerSlug != "secureco" {
		t.Errorf("Expected partner_slug 'secureco', got '%s'", decoded.PartnerSlug)
	}
	if decoded.RevenueSharePercent != 30.50 {
		t.Errorf("Expected revenue_share_percent 30.50, got %v", decoded.RevenueSharePercent)
	}
	if decoded.MaxTenants != 250 {
		t.Errorf("Expected max_tenants 250, got %d", decoded.MaxTenants)
	}
	if !decoded.IsActive {
		t.Error("Expected is_active to be true")
	}
}

func TestPartnerTenantMappingSerialization(t *testing.T) {
	now := time.Now()

	mapping := PartnerTenantMapping{
		ID:             uuid.New(),
		PartnerID:      uuid.New(),
		OrganizationID: uuid.New(),
		OnboardedAt:    now,
		CreatedAt:      now,
	}

	data, err := json.Marshal(mapping)
	if err != nil {
		t.Fatalf("Failed to marshal PartnerTenantMapping: %v", err)
	}

	var decoded PartnerTenantMapping
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal PartnerTenantMapping: %v", err)
	}

	if decoded.PartnerID != mapping.PartnerID {
		t.Errorf("Expected partner_id %s, got %s", mapping.PartnerID, decoded.PartnerID)
	}
	if decoded.OrganizationID != mapping.OrganizationID {
		t.Errorf("Expected organization_id %s, got %s", mapping.OrganizationID, decoded.OrganizationID)
	}
}

// ============================================================
// DEFAULT BRANDING TESTS
// ============================================================

func TestDefaultBranding(t *testing.T) {
	orgID := uuid.New()
	d := DefaultBranding(orgID)

	if d.OrganizationID != orgID {
		t.Errorf("Expected organization_id %s, got %s", orgID, d.OrganizationID)
	}
	if d.ProductName != "ComplianceForge" {
		t.Errorf("Expected product_name 'ComplianceForge', got '%s'", d.ProductName)
	}
	if d.ColorPrimary != "#4F46E5" {
		t.Errorf("Expected default color_primary '#4F46E5', got '%s'", d.ColorPrimary)
	}
	if d.FontFamilyHeading != "Inter" {
		t.Errorf("Expected default font_family_heading 'Inter', got '%s'", d.FontFamilyHeading)
	}
	if d.SidebarStyle != "dark" {
		t.Errorf("Expected default sidebar_style 'dark', got '%s'", d.SidebarStyle)
	}
	if d.CornerRadius != "medium" {
		t.Errorf("Expected default corner_radius 'medium', got '%s'", d.CornerRadius)
	}
	if d.Density != "comfortable" {
		t.Errorf("Expected default density 'comfortable', got '%s'", d.Density)
	}
	if !d.ShowPoweredBy {
		t.Error("Expected show_powered_by to default to true")
	}
	if !d.ShowHelpWidget {
		t.Error("Expected show_help_widget to default to true")
	}
	if !d.ShowMarketplace {
		t.Error("Expected show_marketplace to default to true")
	}
	if !d.ShowKnowledgeBase {
		t.Error("Expected show_knowledge_base to default to true")
	}
}

// ============================================================
// ENUM VALIDATION TESTS
// ============================================================

func TestIsValidEnum(t *testing.T) {
	tests := []struct {
		name    string
		val     string
		allowed []string
		valid   bool
	}{
		{"valid sidebar dark", "dark", []string{"light", "dark", "branded"}, true},
		{"valid sidebar light", "light", []string{"light", "dark", "branded"}, true},
		{"invalid sidebar", "blue", []string{"light", "dark", "branded"}, false},
		{"valid radius none", "none", []string{"none", "small", "medium", "large", "full"}, true},
		{"invalid radius", "tiny", []string{"none", "small", "medium", "large", "full"}, false},
		{"valid density compact", "compact", []string{"compact", "comfortable", "spacious"}, true},
		{"invalid density", "tight", []string{"compact", "comfortable", "spacious"}, false},
		{"empty value", "", []string{"a", "b"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidEnum(tt.val, tt.allowed)
			if result != tt.valid {
				t.Errorf("isValidEnum(%q, %v) = %v, want %v", tt.val, tt.allowed, result, tt.valid)
			}
		})
	}
}

// ============================================================
// URL VALIDATION TESTS
// ============================================================

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		expectError bool
	}{
		{"valid https", "https://example.com", false},
		{"valid http", "http://example.com", false},
		{"valid with path", "https://example.com/privacy", false},
		{"valid with port", "https://example.com:8080/api", false},
		{"ftp scheme", "ftp://example.com", true},
		{"no scheme", "example.com", true},
		{"javascript scheme", "javascript:alert(1)", true},
		{"data scheme", "data:text/html,test", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateURL(tt.url)
			if tt.expectError && err == nil {
				t.Errorf("Expected error for URL %q but got nil", tt.url)
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error for URL %q but got: %v", tt.url, err)
			}
		})
	}
}

// ============================================================
// ALLOWED LOGO TYPES TESTS
// ============================================================

func TestAllowedLogoTypes(t *testing.T) {
	allowed := []string{"full", "icon", "dark", "light", "favicon", "email"}
	for _, lt := range allowed {
		if !AllowedLogoTypes[lt] {
			t.Errorf("Logo type %q should be allowed", lt)
		}
	}

	disallowed := []string{"banner", "splash", "thumbnail", "watermark", ""}
	for _, lt := range disallowed {
		if AllowedLogoTypes[lt] {
			t.Errorf("Logo type %q should not be allowed", lt)
		}
	}
}

// ============================================================
// PDFBrandingConfig TESTS
// ============================================================

func TestPDFBrandingConfigSerialization(t *testing.T) {
	config := PDFBrandingConfig{
		ProductName:      "MyGRC",
		CompanyName:      "Acme Corp",
		LogoURL:          "/storage/orgs/123/branding/full.svg",
		ColorPrimary:     "#FF6600",
		ColorSecondary:   "#0066FF",
		FontFamily:       "Montserrat",
		ShowPoweredBy:    true,
		PrivacyPolicyURL: "https://acme.com/privacy",
	}

	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal PDFBrandingConfig: %v", err)
	}

	var decoded PDFBrandingConfig
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal PDFBrandingConfig: %v", err)
	}

	if decoded.ProductName != "MyGRC" {
		t.Errorf("Expected product_name 'MyGRC', got '%s'", decoded.ProductName)
	}
	if decoded.FontFamily != "Montserrat" {
		t.Errorf("Expected font_family 'Montserrat', got '%s'", decoded.FontFamily)
	}
	if !decoded.ShowPoweredBy {
		t.Error("Expected show_powered_by to be true")
	}
}

// ============================================================
// COMPLEX CSS SANITIZATION TESTS
// ============================================================

func TestSanitizeCSSMultipleThreats(t *testing.T) {
	maliciousCSS := `
@import url('https://evil.com/steal.css');
.header {
  color: #333;
  background: javascript:alert(1);
  font-size: 14px;
  position: absolute;
  behavior: url(evil.htc);
  -moz-binding: url("evil.xml#xbl");
  padding: 10px;
  z-index: 99999;
}
.body {
  background-image: url(data:text/html,<script>alert(1)</script>);
  margin: 20px;
}
</style><script>alert(1)</script><style>
`

	result := SanitizeCSS(maliciousCSS)

	// Safe properties should survive
	if !strings.Contains(result, "color: #333") {
		t.Error("Safe 'color' property should survive")
	}
	if !strings.Contains(result, "font-size: 14px") {
		t.Error("Safe 'font-size' property should survive")
	}
	if !strings.Contains(result, "padding: 10px") {
		t.Error("Safe 'padding' property should survive")
	}
	if !strings.Contains(result, "margin: 20px") {
		t.Error("Safe 'margin' property should survive")
	}

	// Dangerous content should be blocked
	if strings.Contains(result, "javascript:") {
		t.Error("javascript: should be blocked")
	}
	if strings.Contains(result, "@import") {
		t.Error("@import should be blocked")
	}
	if strings.Contains(result, "behavior:") {
		t.Error("behavior: should be blocked")
	}
	if strings.Contains(result, "-moz-binding") {
		t.Error("-moz-binding should be blocked")
	}
	if strings.Contains(result, "<script>") {
		t.Error("<script> should be blocked")
	}
	if strings.Contains(result, "data:text/html") {
		t.Error("data: URIs should be blocked")
	}
}

func TestSanitizeCSSPreservesFlexboxProperties(t *testing.T) {
	css := `.container {
  display: flex;
  flex-direction: row;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
}`

	result := SanitizeCSS(css)

	if !strings.Contains(result, "display: flex") {
		t.Error("display should be allowed")
	}
	if !strings.Contains(result, "flex-direction: row") {
		t.Error("flex-direction should be allowed")
	}
	if !strings.Contains(result, "justify-content: space-between") {
		t.Error("justify-content should be allowed")
	}
	if !strings.Contains(result, "align-items: center") {
		t.Error("align-items should be allowed")
	}
	if !strings.Contains(result, "gap: 16px") {
		t.Error("gap should be allowed")
	}
}

// ============================================================
// DOMAIN VERIFICATION RESULT TESTS
// ============================================================

func TestDomainVerificationResultSerialization(t *testing.T) {
	result := DomainVerificationResult{
		Domain:     "grc.acme.com",
		Status:     "verified",
		CNAMEValid: true,
		TXTValid:   false,
		Token:      "abc123def456",
		SSLStatus:  "provisioning",
		Message:    "Domain verified successfully.",
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to marshal DomainVerificationResult: %v", err)
	}

	var decoded DomainVerificationResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal DomainVerificationResult: %v", err)
	}

	if decoded.Domain != "grc.acme.com" {
		t.Errorf("Expected domain 'grc.acme.com', got '%s'", decoded.Domain)
	}
	if decoded.Status != "verified" {
		t.Errorf("Expected status 'verified', got '%s'", decoded.Status)
	}
	if !decoded.CNAMEValid {
		t.Error("Expected cname_valid to be true")
	}
	if decoded.TXTValid {
		t.Error("Expected txt_valid to be false")
	}
}
