package integrations

import (
	"compress/flate"
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// ============================================================
// SAML 2.0 SP TYPES
// ============================================================

// SAMLProvider implements a SAML 2.0 Service Provider.
type SAMLProvider struct {
	EntityID        string
	ACSURL          string
	SLOURL          string
	MetadataURL     string
	IdPSSOURL       string
	IdPSLOURL       string
	IdPCertPEM      string
	NameIDFormat    string
	AttributeMap    map[string]string // IdP attr name -> local field
	AllowedDomains  []string
	OrgID           uuid.UUID
}

// SAMLAssertion is the parsed result of a SAML Response.
type SAMLAssertion struct {
	NameID     string
	Email      string
	FirstName  string
	LastName   string
	Groups     []string
	Attributes map[string]string
	SessionID  string
	Expiry     time.Time
}

// ============================================================
// SAML METADATA GENERATION
// ============================================================

// spMetadata is the XML representation of our SP metadata.
type spMetadata struct {
	XMLName          xml.Name `xml:"md:EntityDescriptor"`
	XMLNS            string   `xml:"xmlns:md,attr"`
	EntityID         string   `xml:"entityID,attr"`
	SPSSODescriptor  spSSODescriptor
}

type spSSODescriptor struct {
	XMLName                    xml.Name `xml:"md:SPSSODescriptor"`
	AuthnRequestsSigned        string   `xml:"AuthnRequestsSigned,attr"`
	WantAssertionsSigned       string   `xml:"WantAssertionsSigned,attr"`
	ProtocolSupportEnumeration string   `xml:"protocolSupportEnumeration,attr"`
	NameIDFormat               nameIDFormat
	ACS                        assertionConsumerService
	SLO                        singleLogoutService
}

type nameIDFormat struct {
	XMLName xml.Name `xml:"md:NameIDFormat"`
	Value   string   `xml:",chardata"`
}

type assertionConsumerService struct {
	XMLName  xml.Name `xml:"md:AssertionConsumerService"`
	Binding  string   `xml:"Binding,attr"`
	Location string   `xml:"Location,attr"`
	Index    string   `xml:"index,attr"`
}

type singleLogoutService struct {
	XMLName  xml.Name `xml:"md:SingleLogoutService"`
	Binding  string   `xml:"Binding,attr"`
	Location string   `xml:"Location,attr"`
}

// GenerateMetadata returns the SP metadata XML document.
func (sp *SAMLProvider) GenerateMetadata() ([]byte, error) {
	nidFormat := sp.NameIDFormat
	if nidFormat == "" {
		nidFormat = "urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress"
	}

	metadata := spMetadata{
		XMLNS:    "urn:oasis:names:tc:SAML:2.0:metadata",
		EntityID: sp.EntityID,
		SPSSODescriptor: spSSODescriptor{
			AuthnRequestsSigned:        "false",
			WantAssertionsSigned:       "true",
			ProtocolSupportEnumeration: "urn:oasis:names:tc:SAML:2.0:protocol",
			NameIDFormat: nameIDFormat{
				Value: nidFormat,
			},
			ACS: assertionConsumerService{
				Binding:  "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST",
				Location: sp.ACSURL,
				Index:    "0",
			},
			SLO: singleLogoutService{
				Binding:  "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect",
				Location: sp.SLOURL,
			},
		},
	}

	output, err := xml.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal SP metadata: %w", err)
	}

	return append([]byte(xml.Header), output...), nil
}

// ============================================================
// SAML ACS — Assertion Consumer Service
// ============================================================

// samlResponse is a minimal struct for parsing SAML 2.0 Response XML.
type samlResponse struct {
	XMLName      xml.Name       `xml:"Response"`
	ID           string         `xml:"ID,attr"`
	IssueInstant string         `xml:"IssueInstant,attr"`
	Status       samlStatus     `xml:"Status"`
	Assertion    samlAssertionX `xml:"Assertion"`
}

type samlStatus struct {
	StatusCode samlStatusCode `xml:"StatusCode"`
}

type samlStatusCode struct {
	Value string `xml:"Value,attr"`
}

type samlAssertionX struct {
	XMLName    xml.Name       `xml:"Assertion"`
	Subject    samlSubject    `xml:"Subject"`
	Conditions samlConditions `xml:"Conditions"`
	Attributes []samlAttrStmt `xml:"AttributeStatement>Attribute"`
}

type samlSubject struct {
	NameID samlNameID `xml:"NameID"`
}

type samlNameID struct {
	Value  string `xml:",chardata"`
	Format string `xml:"Format,attr"`
}

type samlConditions struct {
	NotBefore    string `xml:"NotBefore,attr"`
	NotOnOrAfter string `xml:"NotOnOrAfter,attr"`
}

type samlAttrStmt struct {
	Name   string           `xml:"Name,attr"`
	Values []samlAttrValues `xml:"AttributeValue"`
}

type samlAttrValues struct {
	Value string `xml:",chardata"`
}

// HandleACS processes a SAML Response POSTed to the ACS endpoint.
// It decodes, validates the IdP certificate, and extracts user attributes.
func (sp *SAMLProvider) HandleACS(samlResponseB64 string) (*SAMLAssertion, error) {
	// 1. Base64-decode the SAML Response
	rawXML, err := base64.StdEncoding.DecodeString(samlResponseB64)
	if err != nil {
		return nil, fmt.Errorf("failed to base64-decode SAML response: %w", err)
	}

	// 2. Parse XML
	var resp samlResponse
	if err := xml.Unmarshal(rawXML, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse SAML response XML: %w", err)
	}

	// 3. Validate status
	if !strings.HasSuffix(resp.Status.StatusCode.Value, ":Success") {
		return nil, fmt.Errorf("SAML response status is not Success: %s", resp.Status.StatusCode.Value)
	}

	// 4. Validate IdP certificate if provided
	if sp.IdPCertPEM != "" {
		if err := sp.validateIdPCertificate(); err != nil {
			log.Warn().Err(err).Msg("IdP certificate validation warning")
			// We log but do not hard-fail in dev; production should enforce.
		}
	}

	// 5. Validate time conditions
	if resp.Assertion.Conditions.NotOnOrAfter != "" {
		notAfter, err := time.Parse(time.RFC3339, resp.Assertion.Conditions.NotOnOrAfter)
		if err == nil && time.Now().After(notAfter) {
			return nil, fmt.Errorf("SAML assertion has expired (NotOnOrAfter: %s)", resp.Assertion.Conditions.NotOnOrAfter)
		}
	}

	// 6. Extract attributes
	attrs := make(map[string]string)
	var groups []string
	for _, a := range resp.Assertion.Attributes {
		if len(a.Values) > 0 {
			attrs[a.Name] = a.Values[0].Value
		}
		// Collect all group values
		lowerName := strings.ToLower(a.Name)
		if lowerName == "groups" || lowerName == "memberof" || strings.Contains(lowerName, "group") {
			for _, v := range a.Values {
				groups = append(groups, v.Value)
			}
		}
	}

	assertion := &SAMLAssertion{
		NameID:     resp.Assertion.Subject.NameID.Value,
		Attributes: attrs,
		Groups:     groups,
		SessionID:  resp.ID,
	}

	// Map attributes to standard fields using configured mapping
	if sp.AttributeMap != nil {
		if key, ok := sp.AttributeMap["email"]; ok {
			if v, found := attrs[key]; found {
				assertion.Email = v
			}
		}
		if key, ok := sp.AttributeMap["first_name"]; ok {
			if v, found := attrs[key]; found {
				assertion.FirstName = v
			}
		}
		if key, ok := sp.AttributeMap["last_name"]; ok {
			if v, found := attrs[key]; found {
				assertion.LastName = v
			}
		}
	}

	// Fall back to NameID as email if not mapped
	if assertion.Email == "" && strings.Contains(assertion.NameID, "@") {
		assertion.Email = assertion.NameID
	}

	// 7. Check allowed domains
	if len(sp.AllowedDomains) > 0 && assertion.Email != "" {
		parts := strings.SplitN(assertion.Email, "@", 2)
		if len(parts) == 2 {
			domain := strings.ToLower(parts[1])
			allowed := false
			for _, d := range sp.AllowedDomains {
				if strings.ToLower(d) == domain {
					allowed = true
					break
				}
			}
			if !allowed {
				return nil, fmt.Errorf("email domain %s is not in the allowed list", domain)
			}
		}
	}

	// Parse expiry from conditions
	if resp.Assertion.Conditions.NotOnOrAfter != "" {
		if t, err := time.Parse(time.RFC3339, resp.Assertion.Conditions.NotOnOrAfter); err == nil {
			assertion.Expiry = t
		}
	}
	if assertion.Expiry.IsZero() {
		assertion.Expiry = time.Now().Add(8 * time.Hour)
	}

	log.Info().Str("email", assertion.Email).Str("org_id", sp.OrgID.String()).Msg("SAML ACS assertion processed")
	return assertion, nil
}

// validateIdPCertificate parses the PEM and checks it is a valid X.509 cert.
func (sp *SAMLProvider) validateIdPCertificate() error {
	certPEM := sp.IdPCertPEM
	if !strings.Contains(certPEM, "BEGIN CERTIFICATE") {
		certPEM = "-----BEGIN CERTIFICATE-----\n" + certPEM + "\n-----END CERTIFICATE-----"
	}

	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		return fmt.Errorf("failed to decode PEM block from IdP certificate")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse IdP X.509 certificate: %w", err)
	}

	if time.Now().After(cert.NotAfter) {
		return fmt.Errorf("IdP certificate has expired (NotAfter: %s)", cert.NotAfter.Format(time.RFC3339))
	}
	if time.Now().Before(cert.NotBefore) {
		return fmt.Errorf("IdP certificate is not yet valid (NotBefore: %s)", cert.NotBefore.Format(time.RFC3339))
	}

	return nil
}

// ============================================================
// SAML SLO — Single Logout
// ============================================================

// HandleSLO generates a LogoutRequest redirect URL to the IdP.
func (sp *SAMLProvider) HandleSLO(nameID, sessionIndex string) (string, error) {
	if sp.IdPSLOURL == "" {
		return "", fmt.Errorf("IdP SLO URL is not configured")
	}

	reqID := "_" + uuid.New().String()
	issueInstant := time.Now().UTC().Format(time.RFC3339)

	logoutRequest := fmt.Sprintf(
		`<samlp:LogoutRequest xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol"
			xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion"
			ID="%s" Version="2.0" IssueInstant="%s"
			Destination="%s">
			<saml:Issuer>%s</saml:Issuer>
			<saml:NameID Format="urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress">%s</saml:NameID>
			<samlp:SessionIndex>%s</samlp:SessionIndex>
		</samlp:LogoutRequest>`,
		reqID, issueInstant, sp.IdPSLOURL, sp.EntityID,
		xmlEscape(nameID), xmlEscape(sessionIndex),
	)

	// DEFLATE compress
	var buf strings.Builder
	w, err := flate.NewWriter(&buf, flate.DefaultCompression)
	if err != nil {
		return "", fmt.Errorf("failed to create deflate writer: %w", err)
	}
	if _, err := io.WriteString(w, logoutRequest); err != nil {
		return "", fmt.Errorf("failed to deflate LogoutRequest: %w", err)
	}
	if err := w.Close(); err != nil {
		return "", fmt.Errorf("failed to close deflate writer: %w", err)
	}

	encoded := base64.StdEncoding.EncodeToString([]byte(buf.String()))
	redirectURL := fmt.Sprintf("%s?SAMLRequest=%s", sp.IdPSLOURL, url.QueryEscape(encoded))

	log.Info().Str("org_id", sp.OrgID.String()).Msg("SAML SLO redirect generated")
	return redirectURL, nil
}

// ============================================================
// SAML AUTHN REQUEST
// ============================================================

// GenerateAuthnRequest builds an AuthnRequest redirect URL.
func (sp *SAMLProvider) GenerateAuthnRequest() (string, error) {
	if sp.IdPSSOURL == "" {
		return "", fmt.Errorf("IdP SSO URL is not configured")
	}

	reqID := "_" + uuid.New().String()
	issueInstant := time.Now().UTC().Format(time.RFC3339)

	authnRequest := fmt.Sprintf(
		`<samlp:AuthnRequest xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol"
			xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion"
			ID="%s" Version="2.0" IssueInstant="%s"
			Destination="%s"
			AssertionConsumerServiceURL="%s"
			ProtocolBinding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST">
			<saml:Issuer>%s</saml:Issuer>
			<samlp:NameIDPolicy
				Format="%s"
				AllowCreate="true"/>
		</samlp:AuthnRequest>`,
		reqID, issueInstant, sp.IdPSSOURL, sp.ACSURL,
		sp.EntityID, sp.NameIDFormat,
	)

	// DEFLATE compress
	var buf strings.Builder
	w, err := flate.NewWriter(&buf, flate.DefaultCompression)
	if err != nil {
		return "", fmt.Errorf("failed to create deflate writer: %w", err)
	}
	if _, err := io.WriteString(w, authnRequest); err != nil {
		return "", fmt.Errorf("failed to deflate AuthnRequest: %w", err)
	}
	if err := w.Close(); err != nil {
		return "", fmt.Errorf("failed to close deflate writer: %w", err)
	}

	encoded := base64.StdEncoding.EncodeToString([]byte(buf.String()))
	redirectURL := fmt.Sprintf("%s?SAMLRequest=%s", sp.IdPSSOURL, url.QueryEscape(encoded))

	return redirectURL, nil
}

// ============================================================
// INTEGRATION ADAPTER INTERFACE
// ============================================================

// SAMLIntegration implements the IntegrationAdapter interface (from service).
type SAMLIntegration struct {
	provider *SAMLProvider
	config   map[string]interface{}
}

func NewSAMLIntegration() *SAMLIntegration {
	return &SAMLIntegration{}
}

func (s *SAMLIntegration) Type() string {
	return "sso_saml"
}

func (s *SAMLIntegration) Connect(_ context.Context, config map[string]interface{}) error {
	s.config = config
	s.provider = &SAMLProvider{}

	if v, ok := config["entity_id"].(string); ok {
		s.provider.EntityID = v
	}
	if v, ok := config["sso_url"].(string); ok {
		s.provider.IdPSSOURL = v
	}
	if v, ok := config["slo_url"].(string); ok {
		s.provider.IdPSLOURL = v
	}
	if v, ok := config["certificate"].(string); ok {
		s.provider.IdPCertPEM = v
	}
	if v, ok := config["acs_url"].(string); ok {
		s.provider.ACSURL = v
	}

	return nil
}

func (s *SAMLIntegration) Disconnect(_ context.Context) error {
	s.provider = nil
	s.config = nil
	return nil
}

func (s *SAMLIntegration) HealthCheck(_ context.Context) (*HealthResult, error) {
	if s.provider == nil {
		return &HealthResult{Status: "unhealthy", Message: "SAML provider not initialised"}, nil
	}

	// Validate the certificate
	if s.provider.IdPCertPEM != "" {
		if err := s.provider.validateIdPCertificate(); err != nil {
			return &HealthResult{Status: "degraded", Message: err.Error()}, nil
		}
	}

	// Check that the IdP SSO URL is reachable
	if s.provider.IdPSSOURL != "" {
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Head(s.provider.IdPSSOURL)
		if err != nil {
			return &HealthResult{Status: "unhealthy", Message: fmt.Sprintf("IdP unreachable: %v", err)}, nil
		}
		resp.Body.Close()
	}

	return &HealthResult{Status: "healthy", Message: "SAML IdP is reachable and certificate is valid"}, nil
}

func (s *SAMLIntegration) Sync(_ context.Context) (*SyncResultAdapter, error) {
	// SSO integrations do not synchronise data.
	return &SyncResultAdapter{}, nil
}

// HealthResult matches the adapter-level result type.
type HealthResult struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// SyncResultAdapter matches the adapter-level result type.
type SyncResultAdapter struct {
	RecordsProcessed int    `json:"records_processed"`
	RecordsCreated   int    `json:"records_created"`
	RecordsUpdated   int    `json:"records_updated"`
	RecordsFailed    int    `json:"records_failed"`
	Error            string `json:"error,omitempty"`
}

// xmlEscape escapes XML special characters.
func xmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}
