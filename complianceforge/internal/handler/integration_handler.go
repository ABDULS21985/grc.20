package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/complianceforge/platform/internal/integrations"
	"github.com/complianceforge/platform/internal/middleware"
	"github.com/complianceforge/platform/internal/models"
	"github.com/complianceforge/platform/internal/service"
)

// ============================================================
// INTEGRATION HANDLER
// ============================================================

// IntegrationHandler handles HTTP requests for the Integration Hub.
type IntegrationHandler struct {
	svc    *service.IntegrationService
	appURL string // Base URL for generating SSO callback URLs
}

// NewIntegrationHandler creates a new IntegrationHandler.
func NewIntegrationHandler(svc *service.IntegrationService, appURL string) *IntegrationHandler {
	return &IntegrationHandler{svc: svc, appURL: appURL}
}

// RegisterRoutes registers all integration hub routes on the given router.
func (h *IntegrationHandler) RegisterRoutes(r chi.Router) {
	// Integration CRUD
	r.Get("/integrations", h.ListIntegrations)
	r.Post("/integrations", h.CreateIntegration)
	r.Get("/integrations/{id}", h.GetIntegration)
	r.Put("/integrations/{id}", h.UpdateIntegration)
	r.Delete("/integrations/{id}", h.DeleteIntegration)

	// Integration operations
	r.Post("/integrations/{id}/test", h.TestConnection)
	r.Post("/integrations/{id}/sync", h.TriggerSync)
	r.Get("/integrations/{id}/logs", h.GetSyncLogs)
	r.Get("/integrations/{id}/health", h.GetIntegrationHealth)

	// SSO configuration
	r.Get("/settings/sso", h.GetSSOConfig)
	r.Put("/settings/sso", h.UpdateSSOConfig)

	// SAML endpoints (no auth required for metadata; ACS is IdP-initiated)
	r.Get("/auth/saml/metadata", h.SAMLMetadata)
	r.Post("/auth/saml/acs", h.SAMLACS)

	// OIDC endpoints
	r.Get("/auth/oidc/login", h.OIDCLogin)
	r.Get("/auth/oidc/callback", h.OIDCCallback)

	// API key management
	r.Get("/settings/api-keys", h.ListAPIKeys)
	r.Post("/settings/api-keys", h.CreateAPIKey)
	r.Delete("/settings/api-keys/{id}", h.RevokeAPIKey)

	// Health summary
	r.Get("/integrations/health/summary", h.GetHealthSummary)
}

// ============================================================
// INTEGRATIONS — CRUD
// ============================================================

// ListIntegrations returns all integrations for the organisation.
// GET /api/v1/integrations
func (h *IntegrationHandler) ListIntegrations(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	integrationsList, err := h.svc.ListIntegrations(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list integrations")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: integrationsList})
}

// CreateIntegration creates a new third-party integration.
// POST /api/v1/integrations
func (h *IntegrationHandler) CreateIntegration(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())

	var input service.CreateIntegrationInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if input.Name == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "name is required")
		return
	}
	if input.IntegrationType == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "integration_type is required")
		return
	}

	integration, err := h.svc.CreateIntegration(r.Context(), orgID, userID, input)
	if err != nil {
		log.Error().Err(err).Msg("failed to create integration")
		writeError(w, http.StatusBadRequest, "CREATE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: integration})
}

// GetIntegration returns a single integration by ID.
// GET /api/v1/integrations/{id}
func (h *IntegrationHandler) GetIntegration(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid integration ID")
		return
	}

	integration, err := h.svc.GetIntegration(r.Context(), orgID, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Integration not found")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: integration})
}

// UpdateIntegration updates an existing integration.
// PUT /api/v1/integrations/{id}
func (h *IntegrationHandler) UpdateIntegration(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid integration ID")
		return
	}

	var input service.UpdateIntegrationInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	integration, err := h.svc.UpdateIntegration(r.Context(), orgID, id, input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "UPDATE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: integration})
}

// DeleteIntegration soft-deletes an integration.
// DELETE /api/v1/integrations/{id}
func (h *IntegrationHandler) DeleteIntegration(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid integration ID")
		return
	}

	if err := h.svc.DeleteIntegration(r.Context(), orgID, id); err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Integration deleted successfully"},
	})
}

// ============================================================
// INTEGRATIONS — OPERATIONS
// ============================================================

// TestConnection tests connectivity for an integration.
// POST /api/v1/integrations/{id}/test
func (h *IntegrationHandler) TestConnection(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid integration ID")
		return
	}

	result, err := h.svc.TestConnection(r.Context(), orgID, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: result})
}

// TriggerSync initiates a manual sync for an integration.
// POST /api/v1/integrations/{id}/sync
func (h *IntegrationHandler) TriggerSync(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid integration ID")
		return
	}

	logEntry, err := h.svc.TriggerSync(r.Context(), orgID, id)
	if err != nil {
		writeError(w, http.StatusBadRequest, "SYNC_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusAccepted, models.APIResponse{Success: true, Data: logEntry})
}

// GetSyncLogs returns paginated sync logs for an integration.
// GET /api/v1/integrations/{id}/logs?page=1&page_size=20
func (h *IntegrationHandler) GetSyncLogs(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid integration ID")
		return
	}

	params := parsePagination(r)
	logs, total, err := h.svc.GetSyncLogs(r.Context(), orgID, id, params.PageSize, params.Offset())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve sync logs")
		return
	}

	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize != 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: logs,
		Pagination: models.Pagination{
			Page:       params.Page,
			PageSize:   params.PageSize,
			TotalItems: total,
			TotalPages: totalPages,
			HasNext:    params.Page < totalPages,
			HasPrev:    params.Page > 1,
		},
	})
}

// GetIntegrationHealth returns the health status of an integration.
// GET /api/v1/integrations/{id}/health
func (h *IntegrationHandler) GetIntegrationHealth(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid integration ID")
		return
	}

	result, err := h.svc.GetHealth(r.Context(), orgID, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: result})
}

// GetHealthSummary returns aggregate health for all integrations.
// GET /api/v1/integrations/health/summary
func (h *IntegrationHandler) GetHealthSummary(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	summary, err := h.svc.GetHealthSummary(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve health summary")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: summary})
}

// ============================================================
// SSO CONFIGURATION
// ============================================================

// GetSSOConfig returns the SSO configuration for the organisation.
// GET /api/v1/settings/sso
func (h *IntegrationHandler) GetSSOConfig(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	config, err := h.svc.GetSSOConfig(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve SSO configuration")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: config})
}

// UpdateSSOConfig updates the SSO configuration.
// PUT /api/v1/settings/sso
func (h *IntegrationHandler) UpdateSSOConfig(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var input service.UpdateSSOConfigInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if input.Protocol != "saml2" && input.Protocol != "oidc" {
		writeError(w, http.StatusBadRequest, "INVALID_PROTOCOL", "Protocol must be 'saml2' or 'oidc'")
		return
	}

	config, err := h.svc.UpdateSSOConfig(r.Context(), orgID, input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "UPDATE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: config})
}

// ============================================================
// SAML ENDPOINTS
// ============================================================

// SAMLMetadata returns the SP metadata XML document.
// GET /api/v1/auth/saml/metadata
func (h *IntegrationHandler) SAMLMetadata(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	ssoConfig, err := h.svc.GetSSOConfig(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to load SSO config")
		return
	}

	entityID := h.appURL + "/api/v1/auth/saml/metadata"
	acsURL := h.appURL + "/api/v1/auth/saml/acs"
	sloURL := h.appURL + "/api/v1/auth/saml/slo"

	if ssoConfig.SAMLEntityID != "" {
		entityID = ssoConfig.SAMLEntityID
	}

	provider := &integrations.SAMLProvider{
		EntityID:     entityID,
		ACSURL:       acsURL,
		SLOURL:       sloURL,
		NameIDFormat: ssoConfig.SAMLNameIDFormat,
		OrgID:        orgID,
	}

	metadata, err := provider.GenerateMetadata()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "METADATA_FAILED", "Failed to generate SP metadata")
		return
	}

	w.Header().Set("Content-Type", "application/xml")
	w.Header().Set("Content-Disposition", "inline; filename=\"sp-metadata.xml\"")
	w.WriteHeader(http.StatusOK)
	w.Write(metadata)
}

// SAMLACS handles the SAML Assertion Consumer Service POST from the IdP.
// POST /api/v1/auth/saml/acs
func (h *IntegrationHandler) SAMLACS(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	// Parse the form data
	if err := r.ParseForm(); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_FORM", "Failed to parse form data")
		return
	}

	samlResponse := r.FormValue("SAMLResponse")
	if samlResponse == "" {
		writeError(w, http.StatusBadRequest, "MISSING_RESPONSE", "SAMLResponse is required")
		return
	}

	// Load SSO config
	ssoConfig, err := h.svc.GetSSOConfig(r.Context(), orgID)
	if err != nil || !ssoConfig.IsEnabled {
		writeError(w, http.StatusForbidden, "SSO_DISABLED", "SSO is not enabled for this organisation")
		return
	}

	// Build attribute mapping
	attrMap := make(map[string]string)
	if ssoConfig.SAMLAttributeMapping != nil {
		for k, v := range ssoConfig.SAMLAttributeMapping {
			if s, ok := v.(string); ok {
				attrMap[k] = s
			}
		}
	}

	provider := &integrations.SAMLProvider{
		EntityID:     ssoConfig.SAMLEntityID,
		ACSURL:       h.appURL + "/api/v1/auth/saml/acs",
		IdPSSOURL:    ssoConfig.SAMLSSOURL,
		IdPCertPEM:   ssoConfig.SAMLCertificate,
		AttributeMap: attrMap,
		AllowedDomains: ssoConfig.AllowedDomains,
		OrgID:        orgID,
	}

	assertion, err := provider.HandleACS(samlResponse)
	if err != nil {
		log.Error().Err(err).Msg("SAML ACS processing failed")
		writeError(w, http.StatusUnauthorized, "SAML_FAILED", err.Error())
		return
	}

	// Return the parsed assertion data
	// In production, this would trigger user creation/lookup and issue a JWT.
	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"email":      assertion.Email,
			"first_name": assertion.FirstName,
			"last_name":  assertion.LastName,
			"groups":     assertion.Groups,
			"session_id": assertion.SessionID,
			"expires_at": assertion.Expiry,
		},
	})
}

// ============================================================
// OIDC ENDPOINTS
// ============================================================

// OIDCLogin initiates the OIDC authorization code flow.
// GET /api/v1/auth/oidc/login
func (h *IntegrationHandler) OIDCLogin(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	ssoConfig, err := h.svc.GetSSOConfig(r.Context(), orgID)
	if err != nil || !ssoConfig.IsEnabled {
		writeError(w, http.StatusForbidden, "SSO_DISABLED", "SSO is not enabled for this organisation")
		return
	}
	if ssoConfig.Protocol != "oidc" {
		writeError(w, http.StatusBadRequest, "WRONG_PROTOCOL", "SSO is configured for SAML, not OIDC")
		return
	}

	// Build claim mapping
	claimMap := make(map[string]string)
	if ssoConfig.OIDCClaimMapping != nil {
		for k, v := range ssoConfig.OIDCClaimMapping {
			if s, ok := v.(string); ok {
				claimMap[k] = s
			}
		}
	}

	provider := &integrations.OIDCProvider{
		IssuerURL:      ssoConfig.OIDCIssuerURL,
		ClientID:       ssoConfig.OIDCClientID,
		RedirectURI:    h.appURL + "/api/v1/auth/oidc/callback",
		Scopes:         ssoConfig.OIDCScopes,
		ClaimMapping:   claimMap,
		AllowedDomains: ssoConfig.AllowedDomains,
		OrgID:          orgID,
	}

	// Generate state parameter (should be stored in session in production)
	state := orgID.String()

	authURL, err := provider.InitiateLogin(r.Context(), state)
	if err != nil {
		log.Error().Err(err).Msg("OIDC login initiation failed")
		writeError(w, http.StatusInternalServerError, "OIDC_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]string{
			"redirect_url": authURL,
		},
	})
}

// OIDCCallback handles the OIDC authorization code callback.
// GET /api/v1/auth/oidc/callback?code=...&state=...
func (h *IntegrationHandler) OIDCCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if code == "" {
		// Check for error response
		oidcError := r.URL.Query().Get("error")
		errorDesc := r.URL.Query().Get("error_description")
		if oidcError != "" {
			writeError(w, http.StatusBadRequest, "OIDC_ERROR",
				"OIDC error: "+oidcError+" — "+errorDesc)
			return
		}
		writeError(w, http.StatusBadRequest, "MISSING_CODE", "Authorization code is required")
		return
	}

	// Parse org ID from state
	orgID, err := uuid.Parse(state)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_STATE", "Invalid state parameter")
		return
	}

	ssoConfig, err := h.svc.GetSSOConfig(r.Context(), orgID)
	if err != nil || !ssoConfig.IsEnabled {
		writeError(w, http.StatusForbidden, "SSO_DISABLED", "SSO is not enabled")
		return
	}

	claimMap := make(map[string]string)
	if ssoConfig.OIDCClaimMapping != nil {
		for k, v := range ssoConfig.OIDCClaimMapping {
			if s, ok := v.(string); ok {
				claimMap[k] = s
			}
		}
	}

	provider := &integrations.OIDCProvider{
		IssuerURL:      ssoConfig.OIDCIssuerURL,
		ClientID:       ssoConfig.OIDCClientID,
		RedirectURI:    h.appURL + "/api/v1/auth/oidc/callback",
		Scopes:         ssoConfig.OIDCScopes,
		ClaimMapping:   claimMap,
		AllowedDomains: ssoConfig.AllowedDomains,
		OrgID:          orgID,
	}

	// Note: in production, we would decrypt the OIDC client secret from ssoConfig
	// and set it on the provider before calling HandleCallback.

	claims, tokens, err := provider.HandleCallback(r.Context(), code)
	if err != nil {
		log.Error().Err(err).Msg("OIDC callback processing failed")
		writeError(w, http.StatusUnauthorized, "OIDC_FAILED", err.Error())
		return
	}

	// Return claims and token data
	// In production, this would trigger user creation/lookup and issue a platform JWT.
	responseData := map[string]interface{}{
		"email":      claims.Email,
		"name":       claims.Name,
		"first_name": claims.FirstName,
		"last_name":  claims.LastName,
		"sub":        claims.Sub,
		"groups":     claims.Groups,
	}
	if tokens != nil {
		responseData["access_token"] = tokens.AccessToken
		responseData["expires_in"] = tokens.ExpiresIn
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: responseData})
}

// ============================================================
// API KEY MANAGEMENT
// ============================================================

// ListAPIKeys returns all API keys for the organisation.
// GET /api/v1/settings/api-keys
func (h *IntegrationHandler) ListAPIKeys(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	keys, err := h.svc.ListAPIKeys(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list API keys")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: keys})
}

// CreateAPIKey generates a new API key. The raw key is returned only once.
// POST /api/v1/settings/api-keys
func (h *IntegrationHandler) CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())

	var input service.CreateAPIKeyInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if input.Name == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "name is required")
		return
	}

	result, err := h.svc.CreateAPIKey(r.Context(), orgID, userID, input)
	if err != nil {
		writeError(w, http.StatusBadRequest, "CREATE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: result})
}

// RevokeAPIKey deactivates an API key.
// DELETE /api/v1/settings/api-keys/{id}
func (h *IntegrationHandler) RevokeAPIKey(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid API key ID")
		return
	}

	if err := h.svc.RevokeAPIKey(r.Context(), orgID, id); err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "API key revoked successfully"},
	})
}

// ============================================================
// INTEGRATION TYPE CATALOG
// ============================================================

// IntegrationCatalogEntry describes an available integration type.
type IntegrationCatalogEntry struct {
	Type         string   `json:"type"`
	Category     string   `json:"category"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Icon         string   `json:"icon"`
	Capabilities []string `json:"capabilities"`
	SetupFields  []string `json:"setup_fields"`
}

// GetIntegrationCatalog returns the list of available integration types.
// This is a static catalog embedded in the handler — no DB query needed.
func GetIntegrationCatalog() []IntegrationCatalogEntry {
	return []IntegrationCatalogEntry{
		{
			Type: "sso_saml", Category: "identity", Name: "SAML 2.0 SSO",
			Description:  "Single Sign-On via SAML 2.0 (Azure AD, Okta, ADFS, Google Workspace)",
			Icon:         "shield-check", Capabilities: []string{"sso", "user_provisioning"},
			SetupFields:  []string{"entity_id", "sso_url", "slo_url", "certificate"},
		},
		{
			Type: "sso_oidc", Category: "identity", Name: "OpenID Connect SSO",
			Description:  "Single Sign-On via OIDC (Azure AD, Okta, Auth0, Google)",
			Icon:         "key", Capabilities: []string{"sso", "user_provisioning"},
			SetupFields:  []string{"issuer_url", "client_id", "client_secret"},
		},
		{
			Type: "cloud_aws", Category: "cloud", Name: "Amazon Web Services",
			Description:  "Import findings from AWS Security Hub, Config, and CloudTrail",
			Icon:         "cloud", Capabilities: []string{"compliance_sync", "finding_import", "asset_discovery"},
			SetupFields:  []string{"access_key_id", "secret_access_key", "region", "role_arn"},
		},
		{
			Type: "cloud_azure", Category: "cloud", Name: "Microsoft Azure",
			Description:  "Import compliance data from Azure Security Center and Policy",
			Icon:         "cloud", Capabilities: []string{"compliance_sync", "finding_import", "asset_discovery"},
			SetupFields:  []string{"tenant_id", "client_id", "client_secret", "subscription_id"},
		},
		{
			Type: "cloud_gcp", Category: "cloud", Name: "Google Cloud Platform",
			Description:  "Import findings from Security Command Center and Cloud Asset Inventory",
			Icon:         "cloud", Capabilities: []string{"compliance_sync", "finding_import"},
			SetupFields:  []string{"project_id", "service_account_key"},
		},
		{
			Type: "siem_splunk", Category: "siem", Name: "Splunk",
			Description:  "Query security events and create incidents from Splunk notable events",
			Icon:         "search", Capabilities: []string{"event_ingestion", "incident_creation"},
			SetupFields:  []string{"base_url", "token", "index"},
		},
		{
			Type: "siem_elastic", Category: "siem", Name: "Elastic Security",
			Description:  "Ingest security alerts and detection rules from Elastic SIEM",
			Icon:         "search", Capabilities: []string{"event_ingestion", "alert_sync"},
			SetupFields:  []string{"base_url", "api_key", "cloud_id"},
		},
		{
			Type: "siem_sentinel", Category: "siem", Name: "Microsoft Sentinel",
			Description:  "Import security incidents and alerts from Azure Sentinel",
			Icon:         "search", Capabilities: []string{"event_ingestion", "incident_creation"},
			SetupFields:  []string{"tenant_id", "client_id", "client_secret", "workspace_id"},
		},
		{
			Type: "itsm_servicenow", Category: "itsm", Name: "ServiceNow",
			Description:  "Create and sync incidents bidirectionally with ServiceNow ITSM",
			Icon:         "ticket", Capabilities: []string{"incident_sync", "ticket_creation", "status_sync"},
			SetupFields:  []string{"instance_url", "username", "password"},
		},
		{
			Type: "itsm_jira", Category: "itsm", Name: "Jira",
			Description:  "Create Jira issues from audit findings and sync statuses",
			Icon:         "ticket", Capabilities: []string{"issue_creation", "status_sync"},
			SetupFields:  []string{"base_url", "email", "api_token", "project"},
		},
		{
			Type: "itsm_freshservice", Category: "itsm", Name: "Freshservice",
			Description:  "Create and track tickets in Freshservice ITSM",
			Icon:         "ticket", Capabilities: []string{"ticket_creation", "status_sync"},
			SetupFields:  []string{"domain", "api_key"},
		},
		{
			Type: "email_smtp", Category: "communication", Name: "Custom SMTP",
			Description:  "Send notifications via your own SMTP server",
			Icon:         "mail", Capabilities: []string{"email_notifications"},
			SetupFields:  []string{"host", "port", "username", "password", "from_address"},
		},
		{
			Type: "email_sendgrid", Category: "communication", Name: "SendGrid",
			Description:  "Send transactional emails via SendGrid API",
			Icon:         "mail", Capabilities: []string{"email_notifications"},
			SetupFields:  []string{"api_key", "from_address", "from_name"},
		},
		{
			Type: "slack", Category: "communication", Name: "Slack",
			Description:  "Send alerts and notifications to Slack channels",
			Icon:         "message-square", Capabilities: []string{"notifications", "alerts"},
			SetupFields:  []string{"webhook_url", "channel", "bot_token"},
		},
		{
			Type: "teams", Category: "communication", Name: "Microsoft Teams",
			Description:  "Send alerts and notifications to Microsoft Teams channels",
			Icon:         "message-square", Capabilities: []string{"notifications", "alerts"},
			SetupFields:  []string{"webhook_url", "channel"},
		},
		{
			Type: "webhook_inbound", Category: "automation", Name: "Inbound Webhook",
			Description:  "Receive data from external systems via webhook",
			Icon:         "arrow-down-circle", Capabilities: []string{"data_ingestion"},
			SetupFields:  []string{"secret"},
		},
		{
			Type: "webhook_outbound", Category: "automation", Name: "Outbound Webhook",
			Description:  "Send events to external systems when actions occur in ComplianceForge",
			Icon:         "arrow-up-circle", Capabilities: []string{"event_dispatch"},
			SetupFields:  []string{"url", "secret", "events"},
		},
		{
			Type: "custom_api", Category: "automation", Name: "Custom API",
			Description:  "Connect to any REST API with custom configuration",
			Icon:         "code", Capabilities: []string{"custom"},
			SetupFields:  []string{"base_url", "auth_type", "api_key_or_token"},
		},
	}
}

// handleParseLimit is a helper that parses a "limit" query parameter.
func handleParseLimit(r *http.Request, defaultLimit int) int {
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 100 {
			return v
		}
	}
	return defaultLimit
}
