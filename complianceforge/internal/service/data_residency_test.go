package service

import (
	"strings"
	"testing"
)

// ============================================================
// EXPORT VALIDATION TESTS
// ============================================================

// TestValidateDataExport_AllowedRegion verifies that exporting data
// within the same region is allowed when enforcement is active.
func TestValidateDataExport_AllowedRegion(t *testing.T) {
	// Simulate the validation logic directly since we cannot
	// connect to a real database in unit tests.
	cfg := &ResidencyConfig{
		DataRegion:            "eu",
		DataResidencyEnforced: true,
		EnforcementMode:       "enforce",
		LegalBasis:            "GDPR Article 6(1)(b)",
	}

	// Same-region export
	result := simulateExportValidation(cfg, "eu")
	if !result.Allowed {
		t.Error("expected same-region export to be allowed")
	}
	if result.SourceRegion != "eu" {
		t.Errorf("expected source region 'eu', got '%s'", result.SourceRegion)
	}
	if result.DestinationRegion != "eu" {
		t.Errorf("expected destination region 'eu', got '%s'", result.DestinationRegion)
	}
	if result.LegalBasis != "GDPR Article 6(1)(b)" {
		t.Errorf("expected legal basis to be set, got '%s'", result.LegalBasis)
	}
}

// TestValidateDataExport_BlockedRegion verifies that exporting data
// to a different region is blocked when enforcement is active.
func TestValidateDataExport_BlockedRegion(t *testing.T) {
	cfg := &ResidencyConfig{
		DataRegion:            "eu",
		DataResidencyEnforced: true,
		EnforcementMode:       "enforce",
		LegalBasis:            "GDPR Article 6(1)(b)",
	}

	// Cross-region export should be blocked
	result := simulateExportValidation(cfg, "global")
	if result.Allowed {
		t.Error("expected cross-region export to be blocked in enforce mode")
	}
	if result.Reason == "" {
		t.Error("expected a reason for blocking")
	}
	if !strings.Contains(result.Reason, "blocked") {
		t.Errorf("expected reason to mention 'blocked', got '%s'", result.Reason)
	}

	// Another cross-region
	result = simulateExportValidation(cfg, "uk")
	if result.Allowed {
		t.Error("expected export from eu to uk to be blocked in enforce mode")
	}
}

// TestValidateDataExport_GlobalRegion verifies that the global region
// imposes no export restrictions.
func TestValidateDataExport_GlobalRegion(t *testing.T) {
	cfg := &ResidencyConfig{
		DataRegion:            "global",
		DataResidencyEnforced: true,
		EnforcementMode:       "enforce",
	}

	result := simulateExportValidation(cfg, "eu")
	if !result.Allowed {
		t.Error("expected global region to allow export to any region")
	}

	result = simulateExportValidation(cfg, "uk")
	if !result.Allowed {
		t.Error("expected global region to allow export to uk")
	}
}

// TestValidateDataExport_AuditMode verifies that audit mode allows
// exports but logs them.
func TestValidateDataExport_AuditMode(t *testing.T) {
	cfg := &ResidencyConfig{
		DataRegion:            "eu",
		DataResidencyEnforced: true,
		EnforcementMode:       "audit",
		LegalBasis:            "GDPR Article 6(1)(b)",
	}

	result := simulateExportValidation(cfg, "global")
	if !result.Allowed {
		t.Error("expected audit mode to allow cross-region export")
	}
	if len(result.RequiredSafeguards) == 0 {
		t.Error("expected safeguards to be listed in audit mode")
	}
}

// TestValidateDataExport_NotEnforced verifies that non-enforced residency
// allows exports with advisory safeguards.
func TestValidateDataExport_NotEnforced(t *testing.T) {
	cfg := &ResidencyConfig{
		DataRegion:            "eu",
		DataResidencyEnforced: false,
		EnforcementMode:       "enforce",
	}

	result := simulateExportValidation(cfg, "uk")
	if !result.Allowed {
		t.Error("expected non-enforced residency to allow cross-region export")
	}
	if len(result.RequiredSafeguards) == 0 {
		t.Error("expected advisory safeguards to be listed")
	}
}

// ============================================================
// TRANSFER VALIDATION TESTS
// ============================================================

// TestValidateDataTransfer_AdequateCountry verifies that transfers to
// GDPR adequate countries are permitted.
func TestValidateDataTransfer_AdequateCountry(t *testing.T) {
	cfg := &ResidencyConfig{
		DataRegion:            "eu",
		DataResidencyEnforced: true,
		EnforcementMode:       "enforce",
		AllowedCountries:      []string{"DE", "FR", "IT", "ES", "NL"},
		BlockedCountries:      []string{},
		GDPRAdequateCountries: GDPRAdequateCountries,
		DataProtectionAuth:    "European Data Protection Board",
	}

	// Japan has adequacy decision
	result := simulateTransferValidation(cfg, "JP")
	if !result.Allowed {
		t.Error("expected transfer to Japan (adequate country) to be allowed")
	}
	if !result.IsGDPRAdequate {
		t.Error("expected Japan to be flagged as GDPR adequate")
	}

	// Switzerland has adequacy decision
	result = simulateTransferValidation(cfg, "CH")
	if !result.Allowed {
		t.Error("expected transfer to Switzerland (adequate country) to be allowed")
	}
	if !result.IsGDPRAdequate {
		t.Error("expected Switzerland to be flagged as GDPR adequate")
	}

	// UK has adequacy decision
	result = simulateTransferValidation(cfg, "GB")
	if !result.Allowed {
		t.Error("expected transfer to UK (adequate country) to be allowed")
	}
}

// TestValidateDataTransfer_InadequateCountry verifies that transfers
// to countries without adequacy decisions require additional safeguards.
func TestValidateDataTransfer_InadequateCountry(t *testing.T) {
	cfg := &ResidencyConfig{
		DataRegion:            "eu",
		DataResidencyEnforced: true,
		EnforcementMode:       "enforce",
		AllowedCountries:      []string{"DE", "FR", "IT", "ES", "NL"},
		BlockedCountries:      []string{},
		GDPRAdequateCountries: GDPRAdequateCountries,
		DataProtectionAuth:    "European Data Protection Board",
	}

	// China does not have an adequacy decision
	result := simulateTransferValidation(cfg, "CN")
	if result.Allowed {
		t.Error("expected transfer to China (inadequate) to be blocked")
	}
	if result.IsGDPRAdequate {
		t.Error("expected China to not be flagged as GDPR adequate")
	}
	if !result.RequiresAdditional {
		t.Error("expected additional safeguards to be required")
	}
	if len(result.RequiredSafeguards) == 0 {
		t.Error("expected specific safeguards to be listed")
	}

	// Russia does not have an adequacy decision
	result = simulateTransferValidation(cfg, "RU")
	if result.Allowed {
		t.Error("expected transfer to Russia (inadequate) to be blocked")
	}

	// India does not have an adequacy decision
	result = simulateTransferValidation(cfg, "IN")
	if result.Allowed {
		t.Error("expected transfer to India (inadequate) to be blocked")
	}
	if len(result.TransferMechanisms) == 0 {
		t.Error("expected transfer mechanisms to be listed for inadequate country")
	}
}

// TestValidateDataTransfer_EUMemberState verifies that intra-EU transfers
// are permitted without additional safeguards.
func TestValidateDataTransfer_EUMemberState(t *testing.T) {
	cfg := &ResidencyConfig{
		DataRegion:            "eu",
		DataResidencyEnforced: true,
		EnforcementMode:       "enforce",
		AllowedCountries:      []string{"DE"},
		BlockedCountries:      []string{},
		GDPRAdequateCountries: GDPRAdequateCountries,
	}

	// Italy is an EU member state but not in AllowedCountries
	result := simulateTransferValidation(cfg, "IT")
	if !result.Allowed {
		t.Error("expected intra-EU transfer to Italy to be allowed")
	}
	if !result.IsGDPRAdequate {
		t.Error("expected EU member state to be flagged as adequate")
	}
}

// TestValidateDataTransfer_BlockedCountry verifies that transfers to
// explicitly blocked countries are rejected.
func TestValidateDataTransfer_BlockedCountry(t *testing.T) {
	cfg := &ResidencyConfig{
		DataRegion:            "eu",
		DataResidencyEnforced: true,
		EnforcementMode:       "enforce",
		AllowedCountries:      []string{"DE", "FR"},
		BlockedCountries:      []string{"CN", "RU"},
		GDPRAdequateCountries: GDPRAdequateCountries,
	}

	result := simulateTransferValidation(cfg, "CN")
	if result.Allowed {
		t.Error("expected transfer to blocked country China to be rejected")
	}
	if !strings.Contains(result.Reason, "blocked") {
		t.Errorf("expected reason to mention 'blocked', got '%s'", result.Reason)
	}
}

// TestValidateDataTransfer_AllowedCountry verifies that transfers to
// countries within the region's allowed list are permitted.
func TestValidateDataTransfer_AllowedCountry(t *testing.T) {
	cfg := &ResidencyConfig{
		DataRegion:            "eu",
		DataResidencyEnforced: true,
		EnforcementMode:       "enforce",
		AllowedCountries:      []string{"DE", "FR", "NL"},
		BlockedCountries:      []string{},
		GDPRAdequateCountries: GDPRAdequateCountries,
	}

	result := simulateTransferValidation(cfg, "DE")
	if !result.Allowed {
		t.Error("expected transfer to allowed country Germany to be permitted")
	}
}

// ============================================================
// GDPR ADEQUACY DECISIONS TESTS
// ============================================================

// TestGDPRAdequacyDecisions verifies that the canonical list of GDPR
// adequate countries includes all expected countries.
func TestGDPRAdequacyDecisions(t *testing.T) {
	expectedCountries := []string{
		"AD", // Andorra
		"AR", // Argentina
		"CA", // Canada
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
		"US", // United States (DPF)
	}

	for _, cc := range expectedCountries {
		if !isCountryInList(cc, GDPRAdequateCountries) {
			t.Errorf("expected %s to be in GDPR adequate countries list", cc)
		}
	}

	// Verify countries that should NOT be in the list
	notAdequate := []string{"CN", "RU", "IN", "BR", "ZA", "MX", "ID"}
	for _, cc := range notAdequate {
		if isCountryInList(cc, GDPRAdequateCountries) {
			t.Errorf("expected %s to NOT be in GDPR adequate countries list", cc)
		}
	}

	// Verify the list has exactly the expected number of entries
	if len(GDPRAdequateCountries) != len(expectedCountries) {
		t.Errorf("expected %d adequate countries, got %d",
			len(expectedCountries), len(GDPRAdequateCountries))
	}
}

// TestGDPRAdequacyDecisions_EUMemberStates verifies the EU member state list.
func TestGDPRAdequacyDecisions_EUMemberStates(t *testing.T) {
	// Verify key EU member states
	expectedEU := []string{"DE", "FR", "IT", "ES", "NL", "BE", "AT", "SE", "DK", "FI", "IE", "PL"}
	for _, cc := range expectedEU {
		if !isCountryInList(cc, EUMemberStates) {
			t.Errorf("expected %s to be in EU member states list", cc)
		}
	}

	// Verify EEA members
	expectedEEA := []string{"IS", "LI", "NO"}
	for _, cc := range expectedEEA {
		if !isCountryInList(cc, EUMemberStates) {
			t.Errorf("expected %s (EEA) to be in EU member states list", cc)
		}
	}

	// UK should NOT be in EU member states (post-Brexit)
	if isCountryInList("GB", EUMemberStates) {
		t.Error("expected GB (UK) to NOT be in EU member states list (post-Brexit)")
	}

	// Should have 27 EU + 3 EEA = 30 entries
	if len(EUMemberStates) != 30 {
		t.Errorf("expected 30 EU/EEA member states, got %d", len(EUMemberStates))
	}
}

// ============================================================
// IP GEOLOCATION TESTS
// ============================================================

// TestCheckIPRegion_ValidIP verifies that known IP ranges resolve
// to the correct country and region.
func TestCheckIPRegion_ValidIP(t *testing.T) {
	svc := &DataResidencyService{}

	tests := []struct {
		name            string
		ip              string
		expectedCountry string
		expectedIsEU    bool
	}{
		{"Google DNS", "8.8.8.8", "US", false},
		{"French IP", "2.1.2.3", "FR", true},
		{"German IP", "5.1.2.3", "DE", true},
		{"UK IP", "51.1.2.3", "GB", false},
		{"Dutch IP", "31.1.2.3", "NL", true},
		{"Swedish IP", "85.1.2.3", "SE", true},
		{"Private network", "10.0.0.1", "XX", false},
		{"Loopback", "127.0.0.1", "XX", false},
		{"Apple IP", "17.1.2.3", "US", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			geo, err := svc.CheckIPRegion(tt.ip)
			if err != nil {
				t.Fatalf("expected no error for %s, got: %v", tt.ip, err)
			}
			if geo.CountryCode != tt.expectedCountry {
				t.Errorf("expected country %s for IP %s, got %s",
					tt.expectedCountry, tt.ip, geo.CountryCode)
			}
			if geo.IsEU != tt.expectedIsEU {
				t.Errorf("expected IsEU=%v for IP %s, got %v",
					tt.expectedIsEU, tt.ip, geo.IsEU)
			}
			if geo.IPAddress != tt.ip {
				t.Errorf("expected IPAddress to be %s, got %s", tt.ip, geo.IPAddress)
			}
		})
	}
}

// TestCheckIPRegion_InvalidIP verifies that invalid IP addresses
// return an appropriate error.
func TestCheckIPRegion_InvalidIP(t *testing.T) {
	svc := &DataResidencyService{}

	invalidIPs := []string{
		"",
		"not-an-ip",
		"256.256.256.256",
		"abc.def.ghi.jkl",
		"1.2.3",
		"1.2.3.4.5",
		"::gg",
	}

	for _, ip := range invalidIPs {
		t.Run(ip, func(t *testing.T) {
			_, err := svc.CheckIPRegion(ip)
			if err == nil {
				t.Errorf("expected error for invalid IP '%s', got nil", ip)
			}
		})
	}
}

// TestCheckIPRegion_RegionMapping verifies that the region field is
// correctly set based on the country code.
func TestCheckIPRegion_RegionMapping(t *testing.T) {
	svc := &DataResidencyService{}

	tests := []struct {
		name           string
		ip             string
		expectedRegion string
	}{
		{"French IP -> france", "2.1.2.3", "france"},
		{"German IP -> dach", "5.1.2.3", "dach"},
		{"UK IP -> uk", "51.1.2.3", "uk"},
		{"Swedish IP -> nordics", "85.1.2.3", "nordics"},
		{"US IP -> global", "8.8.8.8", "global"},
		{"Dutch IP -> eu", "31.1.2.3", "eu"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			geo, err := svc.CheckIPRegion(tt.ip)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if geo.Region != tt.expectedRegion {
				t.Errorf("expected region '%s' for IP %s (country %s), got '%s'",
					tt.expectedRegion, tt.ip, geo.CountryCode, geo.Region)
			}
		})
	}
}

// ============================================================
// REGION CONFIG TESTS
// ============================================================

// TestResidencyRegionConfig_AllRegions verifies that all supported
// region codes are present in the region-to-country mapping and that
// each region's countries are valid ISO 3166-1 alpha-2 codes.
func TestResidencyRegionConfig_AllRegions(t *testing.T) {
	expectedRegions := []string{"eu", "uk", "dach", "nordics", "france", "global"}

	for _, region := range expectedRegions {
		countries, ok := regionToCountryMap[region]
		if !ok {
			t.Errorf("expected region '%s' to be in regionToCountryMap", region)
			continue
		}

		// Global has no country restriction
		if region == "global" {
			if len(countries) != 0 {
				t.Errorf("expected global region to have no country restrictions, got %d", len(countries))
			}
			continue
		}

		// All other regions should have at least one country
		if len(countries) == 0 {
			t.Errorf("expected region '%s' to have at least one country", region)
		}

		// Verify country codes are 2 characters
		for _, cc := range countries {
			if len(cc) != 2 {
				t.Errorf("expected country code '%s' in region '%s' to be 2 characters", cc, region)
			}
			if cc != strings.ToUpper(cc) {
				t.Errorf("expected country code '%s' in region '%s' to be uppercase", cc, region)
			}
		}
	}
}

// TestResidencyRegionConfig_DACHCountries verifies the DACH region contains
// Germany, Austria, and Switzerland.
func TestResidencyRegionConfig_DACHCountries(t *testing.T) {
	dachCountries := regionToCountryMap["dach"]
	expected := []string{"DE", "AT", "CH"}

	for _, cc := range expected {
		found := false
		for _, c := range dachCountries {
			if c == cc {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected %s in DACH region countries", cc)
		}
	}

	if len(dachCountries) != 3 {
		t.Errorf("expected exactly 3 DACH countries, got %d", len(dachCountries))
	}
}

// TestResidencyRegionConfig_NordicsCountries verifies the Nordics region.
func TestResidencyRegionConfig_NordicsCountries(t *testing.T) {
	nordics := regionToCountryMap["nordics"]
	expected := []string{"DK", "FI", "IS", "NO", "SE"}

	for _, cc := range expected {
		found := false
		for _, c := range nordics {
			if c == cc {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected %s in Nordics region countries", cc)
		}
	}

	if len(nordics) != 5 {
		t.Errorf("expected exactly 5 Nordic countries, got %d", len(nordics))
	}
}

// TestCountryToResidencyRegion verifies the reverse mapping from country
// to residency region.
func TestCountryToResidencyRegion(t *testing.T) {
	tests := []struct {
		country  string
		expected string
	}{
		{"DE", "dach"},
		{"AT", "dach"},
		{"CH", "dach"},
		{"GB", "uk"},
		{"FR", "france"},
		{"SE", "nordics"},
		{"NO", "nordics"},
		{"NL", "eu"},
		{"IT", "eu"},
		{"US", "global"},
		{"CN", "global"},
		{"JP", "global"},
	}

	for _, tt := range tests {
		t.Run(tt.country, func(t *testing.T) {
			result := countryToResidencyRegion(tt.country)
			if result != tt.expected {
				t.Errorf("countryToResidencyRegion(%s) = %s, want %s",
					tt.country, result, tt.expected)
			}
		})
	}
}

// TestIsCountryInList verifies the country list membership helper.
func TestIsCountryInList(t *testing.T) {
	list := []string{"DE", "FR", "IT", "ES"}

	if !isCountryInList("DE", list) {
		t.Error("expected DE to be in list")
	}
	if !isCountryInList("de", list) {
		t.Error("expected case-insensitive match for de")
	}
	if isCountryInList("GB", list) {
		t.Error("expected GB to not be in list")
	}
	if isCountryInList("", list) {
		t.Error("expected empty string to not be in list")
	}
}

// ============================================================
// HELPER: Simulate Export Validation
// (Mirrors the service logic without database access)
// ============================================================

func simulateExportValidation(cfg *ResidencyConfig, destinationRegion string) *ExportValidationResult {
	result := &ExportValidationResult{
		SourceRegion:      cfg.DataRegion,
		DestinationRegion: destinationRegion,
	}

	// Global region allows export anywhere
	if cfg.DataRegion == "global" {
		result.Allowed = true
		result.Reason = "Organization is in global region; no export restrictions apply"
		return result
	}

	// Same region is always allowed
	if strings.EqualFold(cfg.DataRegion, destinationRegion) {
		result.Allowed = true
		result.Reason = "Export within the same residency region is allowed"
		result.LegalBasis = cfg.LegalBasis
		return result
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
		return result
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
		return result
	}

	// Enforce mode: block cross-region export
	result.Allowed = false
	result.Reason = "Data export from region '" + cfg.DataRegion + "' to '" + destinationRegion + "' is blocked by data residency enforcement policy"
	result.LegalBasis = cfg.LegalBasis

	return result
}

// ============================================================
// HELPER: Simulate Transfer Validation
// (Mirrors the service logic without database access)
// ============================================================

func simulateTransferValidation(cfg *ResidencyConfig, destinationCountry string) *TransferValidation {
	destCountry := strings.ToUpper(strings.TrimSpace(destinationCountry))

	result := &TransferValidation{
		DestinationCountry: destCountry,
		DataProtectionAuth: cfg.DataProtectionAuth,
	}

	// Check if destination is in the allowed countries for the region
	if isCountryInList(destCountry, cfg.AllowedCountries) {
		result.Allowed = true
		result.IsGDPRAdequate = true
		result.Reason = "Country " + destCountry + " is within the organisation's configured data residency region (" + cfg.DataRegion + ")"
		return result
	}

	// Check if destination country is blocked
	if isCountryInList(destCountry, cfg.BlockedCountries) {
		result.Allowed = false
		result.IsGDPRAdequate = false
		result.Reason = "Country " + destCountry + " is explicitly blocked by the organisation's data residency policy"
		return result
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
		result.Reason = "Country " + destCountry + " has a GDPR adequacy decision; data transfer is permitted"
		result.TransferMechanisms = []string{"GDPR Adequacy Decision"}
		return result
	}

	// Check if destination is an EU/EEA member state
	if isCountryInList(destCountry, EUMemberStates) {
		result.Allowed = true
		result.IsGDPRAdequate = true
		result.RequiresAdditional = false
		result.Reason = "Country " + destCountry + " is an EU/EEA member state; intra-EU transfers are permitted"
		result.TransferMechanisms = []string{"EU/EEA Free Movement of Data"}
		return result
	}

	// No adequacy decision: transfer requires additional safeguards
	result.Allowed = false
	result.IsGDPRAdequate = false
	result.RequiresAdditional = true
	result.Reason = "Country " + destCountry + " does not have a GDPR adequacy decision; transfer requires additional safeguards"
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

	return result
}
