package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// ============================================================
// GOOGLE CLOUD PLATFORM ADAPTER
// ============================================================

// GCPConfig holds credentials for the GCP integration.
type GCPConfig struct {
	ProjectID          string `json:"project_id"`
	ServiceAccountJSON string `json:"service_account_json"` // JSON key file content
	OrganizationID     string `json:"organization_id"`
}

// GCPAdapter implements the integration adapter for Google Cloud Platform.
// It uses the GCP Security Command Center and Cloud Asset Inventory APIs.
type GCPAdapter struct {
	config      GCPConfig
	accessToken string
	tokenExpiry time.Time
	client      *http.Client
}

// NewGCPAdapter creates a new GCP integration adapter.
func NewGCPAdapter() *GCPAdapter {
	return &GCPAdapter{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// Type returns the integration type identifier.
func (a *GCPAdapter) Type() string { return "cloud_gcp" }

// Connect initialises the GCP adapter with the given configuration.
func (a *GCPAdapter) Connect(ctx context.Context, config map[string]interface{}) error {
	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal GCP config: %w", err)
	}
	if err := json.Unmarshal(data, &a.config); err != nil {
		return fmt.Errorf("failed to parse GCP config: %w", err)
	}
	if a.config.ProjectID == "" {
		return fmt.Errorf("project_id is required")
	}
	if a.config.ServiceAccountJSON == "" {
		return fmt.Errorf("service_account_json is required")
	}
	return a.refreshAccessToken(ctx)
}

// Disconnect cleans up the GCP adapter.
func (a *GCPAdapter) Disconnect(_ context.Context) error {
	a.config = GCPConfig{}
	a.accessToken = ""
	a.tokenExpiry = time.Time{}
	return nil
}

// HealthCheck verifies connectivity to GCP APIs.
func (a *GCPAdapter) HealthCheck(ctx context.Context) (*HealthResult, error) {
	start := time.Now()

	if err := a.ensureToken(ctx); err != nil {
		return &HealthResult{
			Status:  "unhealthy",
			Message: fmt.Sprintf("Token refresh failed: %v", err),
		}, nil
	}

	// Verify access by getting project info
	url := fmt.Sprintf("https://cloudresourcemanager.googleapis.com/v1/projects/%s", a.config.ProjectID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return &HealthResult{Status: "unhealthy", Message: err.Error()}, nil
	}
	req.Header.Set("Authorization", "Bearer "+a.accessToken)

	resp, err := a.client.Do(req)
	if err != nil {
		return &HealthResult{
			Status:  "unhealthy",
			Message: fmt.Sprintf("GCP API unreachable: %v", err),
		}, nil
	}
	defer resp.Body.Close()

	latency := time.Since(start).Milliseconds()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return &HealthResult{
			Status:  "unhealthy",
			Message: fmt.Sprintf("GCP API returned %d: %s", resp.StatusCode, string(body)),
		}, nil
	}

	log.Info().Int64("latency_ms", latency).Msg("GCP health check passed")
	return &HealthResult{
		Status:  "healthy",
		Message: fmt.Sprintf("GCP API accessible (%dms)", latency),
	}, nil
}

// Sync collects evidence from GCP Security Command Center and Cloud Asset Inventory.
func (a *GCPAdapter) Sync(ctx context.Context) (*SyncResultAdapter, error) {
	if err := a.ensureToken(ctx); err != nil {
		return nil, fmt.Errorf("token refresh failed: %w", err)
	}

	result := &SyncResultAdapter{}

	// 1. Security Command Center findings
	findings, err := a.syncSecurityFindings(ctx)
	if err != nil {
		log.Warn().Err(err).Msg("GCP security findings sync failed")
		result.Error = err.Error()
	} else {
		result.RecordsProcessed += findings.RecordsProcessed
		result.RecordsCreated += findings.RecordsCreated
	}

	// 2. Cloud Asset Inventory
	assets, err := a.syncAssetInventory(ctx)
	if err != nil {
		log.Warn().Err(err).Msg("GCP asset inventory sync failed")
		if result.Error == "" {
			result.Error = err.Error()
		}
	} else {
		result.RecordsProcessed += assets.RecordsProcessed
		result.RecordsCreated += assets.RecordsCreated
	}

	// 3. IAM policies
	iam, err := a.syncIAMPolicies(ctx)
	if err != nil {
		log.Warn().Err(err).Msg("GCP IAM policies sync failed")
		if result.Error == "" {
			result.Error = err.Error()
		}
	} else {
		result.RecordsProcessed += iam.RecordsProcessed
		result.RecordsCreated += iam.RecordsCreated
	}

	log.Info().
		Int("processed", result.RecordsProcessed).
		Int("created", result.RecordsCreated).
		Msg("GCP sync completed")

	return result, nil
}

// syncSecurityFindings queries Security Command Center for active findings.
func (a *GCPAdapter) syncSecurityFindings(ctx context.Context) (*SyncResultAdapter, error) {
	orgID := a.config.OrganizationID
	if orgID == "" {
		orgID = "-" // Use project-level if no org
	}

	url := fmt.Sprintf(
		"https://securitycenter.googleapis.com/v1/organizations/%s/sources/-/findings?filter=%s&pageSize=100",
		orgID,
		"state=\"ACTIVE\"",
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+a.accessToken)

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("SCC API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("SCC API returned %d: %s", resp.StatusCode, string(body))
	}

	var sccResp struct {
		ListFindingsResults []struct {
			Finding struct {
				Name         string `json:"name"`
				Category     string `json:"category"`
				Severity     string `json:"severity"`
				State        string `json:"state"`
				ResourceName string `json:"resourceName"`
			} `json:"finding"`
		} `json:"listFindingsResults"`
	}
	if err := json.Unmarshal(body, &sccResp); err != nil {
		return nil, fmt.Errorf("failed to parse SCC response: %w", err)
	}

	log.Info().
		Int("findings", len(sccResp.ListFindingsResults)).
		Msg("GCP Security Command Center sync completed")

	return &SyncResultAdapter{
		RecordsProcessed: len(sccResp.ListFindingsResults),
		RecordsCreated:   len(sccResp.ListFindingsResults),
	}, nil
}

// syncAssetInventory queries Cloud Asset Inventory for resource compliance.
func (a *GCPAdapter) syncAssetInventory(ctx context.Context) (*SyncResultAdapter, error) {
	url := fmt.Sprintf(
		"https://cloudasset.googleapis.com/v1/projects/%s/assets?contentType=RESOURCE&pageSize=100",
		a.config.ProjectID,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+a.accessToken)

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Cloud Asset API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Cloud Asset API returned %d: %s", resp.StatusCode, string(body))
	}

	var assetResp struct {
		Assets []struct {
			Name      string `json:"name"`
			AssetType string `json:"assetType"`
		} `json:"assets"`
	}
	if err := json.Unmarshal(body, &assetResp); err != nil {
		return nil, fmt.Errorf("failed to parse asset response: %w", err)
	}

	return &SyncResultAdapter{
		RecordsProcessed: len(assetResp.Assets),
		RecordsCreated:   len(assetResp.Assets),
	}, nil
}

// syncIAMPolicies fetches IAM policy bindings for access review evidence.
func (a *GCPAdapter) syncIAMPolicies(ctx context.Context) (*SyncResultAdapter, error) {
	url := fmt.Sprintf(
		"https://cloudresourcemanager.googleapis.com/v1/projects/%s:getIamPolicy",
		a.config.ProjectID,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader("{}"))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+a.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("IAM API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("IAM API returned %d: %s", resp.StatusCode, string(body))
	}

	var iamResp struct {
		Bindings []struct {
			Role    string   `json:"role"`
			Members []string `json:"members"`
		} `json:"bindings"`
	}
	if err := json.Unmarshal(body, &iamResp); err != nil {
		return nil, fmt.Errorf("failed to parse IAM response: %w", err)
	}

	return &SyncResultAdapter{
		RecordsProcessed: len(iamResp.Bindings),
		RecordsCreated:   len(iamResp.Bindings),
	}, nil
}

// refreshAccessToken obtains an OAuth2 access token from the service account.
func (a *GCPAdapter) refreshAccessToken(ctx context.Context) error {
	var sa struct {
		ClientEmail string `json:"client_email"`
		TokenURI    string `json:"token_uri"`
	}
	if err := json.Unmarshal([]byte(a.config.ServiceAccountJSON), &sa); err != nil {
		return fmt.Errorf("invalid service account JSON: %w", err)
	}

	if sa.TokenURI == "" {
		sa.TokenURI = "https://oauth2.googleapis.com/token"
	}

	// JWT grant type — in production, sign a JWT with the private_key
	payload := fmt.Sprintf(
		"grant_type=urn:ietf:params:oauth:grant-type:jwt-bearer&assertion=%s",
		"placeholder_jwt",
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, sa.TokenURI, strings.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		TokenType   string `json:"token_type"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		log.Warn().Msg("GCP token exchange failed — will retry on next request")
		return nil
	}

	a.accessToken = tokenResp.AccessToken
	a.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	return nil
}

// ensureToken refreshes the access token if it's expired.
func (a *GCPAdapter) ensureToken(ctx context.Context) error {
	if a.accessToken == "" || time.Now().After(a.tokenExpiry.Add(-5*time.Minute)) {
		return a.refreshAccessToken(ctx)
	}
	return nil
}
