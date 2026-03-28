package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// ============================================================
// AZURE CLOUD ADAPTER
// ============================================================

// AzureConfig holds credentials and settings for the Azure integration.
type AzureConfig struct {
	TenantID       string `json:"tenant_id"`
	ClientID       string `json:"client_id"`
	ClientSecret   string `json:"client_secret"`
	SubscriptionID string `json:"subscription_id"`
}

// AzureAdapter implements the integration adapter for Microsoft Azure.
// It uses the MS Graph API and Azure Security Center / Policy API.
type AzureAdapter struct {
	config      AzureConfig
	httpClient  *http.Client
	accessToken string
	tokenExpiry time.Time
}

func NewAzureAdapter() *AzureAdapter {
	return &AzureAdapter{
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (a *AzureAdapter) Type() string {
	return "cloud_azure"
}

func (a *AzureAdapter) Connect(_ context.Context, config map[string]interface{}) error {
	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal Azure config: %w", err)
	}
	if err := json.Unmarshal(data, &a.config); err != nil {
		return fmt.Errorf("failed to parse Azure config: %w", err)
	}

	if a.config.TenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if a.config.ClientID == "" {
		return fmt.Errorf("client_id is required")
	}
	if a.config.ClientSecret == "" {
		return fmt.Errorf("client_secret is required")
	}

	return nil
}

func (a *AzureAdapter) Disconnect(_ context.Context) error {
	a.config = AzureConfig{}
	a.accessToken = ""
	a.tokenExpiry = time.Time{}
	return nil
}

// HealthCheck verifies connectivity via MS Graph API /me endpoint.
func (a *AzureAdapter) HealthCheck(ctx context.Context) (*HealthResult, error) {
	start := time.Now()

	// Acquire token
	if err := a.ensureToken(ctx); err != nil {
		return &HealthResult{
			Status:  "unhealthy",
			Message: fmt.Sprintf("Token acquisition failed: %v", err),
		}, nil
	}

	// Call Graph API organization endpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://graph.microsoft.com/v1.0/organization", nil)
	if err != nil {
		return &HealthResult{Status: "unhealthy", Message: err.Error()}, nil
	}
	req.Header.Set("Authorization", "Bearer "+a.accessToken)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return &HealthResult{
			Status:  "unhealthy",
			Message: fmt.Sprintf("Graph API unreachable: %v", err),
		}, nil
	}
	defer resp.Body.Close()

	latency := time.Since(start).Milliseconds()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return &HealthResult{
			Status:  "unhealthy",
			Message: fmt.Sprintf("Graph API returned HTTP %d: %s", resp.StatusCode, string(body)),
		}, nil
	}

	log.Info().Int64("latency_ms", latency).Msg("Azure Graph API health check passed")
	return &HealthResult{
		Status:  "healthy",
		Message: fmt.Sprintf("Azure Graph API accessible (%dms)", latency),
	}, nil
}

// Sync pulls compliance data from Azure Security Center and Policy.
func (a *AzureAdapter) Sync(ctx context.Context) (*SyncResultAdapter, error) {
	if err := a.ensureToken(ctx); err != nil {
		return nil, fmt.Errorf("failed to acquire Azure token: %w", err)
	}

	result := &SyncResultAdapter{}

	// 1. Security Center — security assessments
	assessments, err := a.syncSecurityAssessments(ctx)
	if err != nil {
		log.Warn().Err(err).Msg("Azure Security Center sync partial failure")
		result.Error = err.Error()
	} else {
		result.RecordsProcessed += assessments.RecordsProcessed
		result.RecordsCreated += assessments.RecordsCreated
	}

	// 2. Azure Policy — compliance states
	policyResults, err := a.syncPolicyCompliance(ctx)
	if err != nil {
		log.Warn().Err(err).Msg("Azure Policy sync partial failure")
		if result.Error == "" {
			result.Error = err.Error()
		}
	} else {
		result.RecordsProcessed += policyResults.RecordsProcessed
		result.RecordsCreated += policyResults.RecordsCreated
	}

	log.Info().
		Int("processed", result.RecordsProcessed).
		Int("created", result.RecordsCreated).
		Msg("Azure sync completed")

	return result, nil
}

// ensureToken acquires or refreshes the OAuth2 client_credentials token.
func (a *AzureAdapter) ensureToken(ctx context.Context) error {
	if a.accessToken != "" && time.Now().Before(a.tokenExpiry) {
		return nil // Token still valid
	}

	tokenURL := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", a.config.TenantID)
	data := url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {a.config.ClientID},
		"client_secret": {a.config.ClientSecret},
		"scope":         {"https://graph.microsoft.com/.default"},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token endpoint returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		TokenType   string `json:"token_type"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return fmt.Errorf("failed to parse token response: %w", err)
	}

	a.accessToken = tokenResp.AccessToken
	a.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn-60) * time.Second)

	return nil
}

// syncSecurityAssessments fetches Azure Security Center assessments.
func (a *AzureAdapter) syncSecurityAssessments(ctx context.Context) (*SyncResultAdapter, error) {
	if a.config.SubscriptionID == "" {
		return &SyncResultAdapter{}, nil
	}

	assessmentURL := fmt.Sprintf(
		"https://management.azure.com/subscriptions/%s/providers/Microsoft.Security/assessments?api-version=2021-06-01",
		a.config.SubscriptionID,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, assessmentURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+a.accessToken)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("security assessments request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	result := &SyncResultAdapter{}

	if resp.StatusCode == http.StatusOK {
		var response struct {
			Value []interface{} `json:"value"`
		}
		if err := json.Unmarshal(body, &response); err == nil {
			result.RecordsProcessed = len(response.Value)
			result.RecordsCreated = len(response.Value)
		}
	}

	return result, nil
}

// syncPolicyCompliance fetches Azure Policy compliance states.
func (a *AzureAdapter) syncPolicyCompliance(ctx context.Context) (*SyncResultAdapter, error) {
	if a.config.SubscriptionID == "" {
		return &SyncResultAdapter{}, nil
	}

	policyURL := fmt.Sprintf(
		"https://management.azure.com/subscriptions/%s/providers/Microsoft.PolicyInsights/policyStates/latest/summarize?api-version=2019-10-01",
		a.config.SubscriptionID,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, policyURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+a.accessToken)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("policy compliance request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	result := &SyncResultAdapter{}

	if resp.StatusCode == http.StatusOK {
		var response struct {
			Value []struct {
				Results struct {
					NonCompliantResources int `json:"nonCompliantResources"`
					NonCompliantPolicies  int `json:"nonCompliantPolicies"`
				} `json:"results"`
			} `json:"value"`
		}
		if err := json.Unmarshal(body, &response); err == nil {
			for _, v := range response.Value {
				result.RecordsProcessed += v.Results.NonCompliantResources + v.Results.NonCompliantPolicies
			}
			result.RecordsCreated = result.RecordsProcessed
		}
	}

	return result, nil
}
