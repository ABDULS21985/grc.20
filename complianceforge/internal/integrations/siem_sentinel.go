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
// MICROSOFT SENTINEL SIEM ADAPTER
// ============================================================

// SentinelConfig holds credentials for the Microsoft Sentinel API.
type SentinelConfig struct {
	TenantID       string `json:"tenant_id"`
	ClientID       string `json:"client_id"`
	ClientSecret   string `json:"client_secret"`
	SubscriptionID string `json:"subscription_id"`
	ResourceGroup  string `json:"resource_group"`
	WorkspaceName  string `json:"workspace_name"`
	WorkspaceID    string `json:"workspace_id"` // Log Analytics workspace ID
}

// SentinelAdapter implements the integration adapter for Microsoft Sentinel SIEM.
type SentinelAdapter struct {
	config      SentinelConfig
	httpClient  *http.Client
	accessToken string
	tokenExpiry time.Time
}

func NewSentinelAdapter() *SentinelAdapter {
	return &SentinelAdapter{
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

func (s *SentinelAdapter) Type() string {
	return "siem_sentinel"
}

func (s *SentinelAdapter) Connect(_ context.Context, config map[string]interface{}) error {
	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal Sentinel config: %w", err)
	}
	if err := json.Unmarshal(data, &s.config); err != nil {
		return fmt.Errorf("failed to parse Sentinel config: %w", err)
	}

	if s.config.TenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if s.config.ClientID == "" {
		return fmt.Errorf("client_id is required")
	}
	if s.config.ClientSecret == "" {
		return fmt.Errorf("client_secret is required")
	}
	if s.config.SubscriptionID == "" {
		return fmt.Errorf("subscription_id is required")
	}
	if s.config.ResourceGroup == "" {
		return fmt.Errorf("resource_group is required")
	}
	if s.config.WorkspaceName == "" {
		return fmt.Errorf("workspace_name is required")
	}

	return nil
}

func (s *SentinelAdapter) Disconnect(_ context.Context) error {
	s.config = SentinelConfig{}
	s.accessToken = ""
	s.tokenExpiry = time.Time{}
	return nil
}

// HealthCheck verifies connectivity to the Sentinel API.
func (s *SentinelAdapter) HealthCheck(ctx context.Context) (*HealthResult, error) {
	start := time.Now()

	if err := s.ensureToken(ctx); err != nil {
		return &HealthResult{
			Status:  "unhealthy",
			Message: fmt.Sprintf("Token acquisition failed: %v", err),
		}, nil
	}

	// Verify by listing incidents
	incidentsURL := fmt.Sprintf(
		"https://management.azure.com/subscriptions/%s/resourceGroups/%s/providers/Microsoft.OperationalInsights/workspaces/%s/providers/Microsoft.SecurityInsights/incidents?api-version=2023-11-01&$top=1",
		s.config.SubscriptionID,
		s.config.ResourceGroup,
		s.config.WorkspaceName,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, incidentsURL, nil)
	if err != nil {
		return &HealthResult{Status: "unhealthy", Message: err.Error()}, nil
	}
	req.Header.Set("Authorization", "Bearer "+s.accessToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return &HealthResult{
			Status:  "unhealthy",
			Message: fmt.Sprintf("Sentinel API unreachable: %v", err),
		}, nil
	}
	defer resp.Body.Close()

	latency := time.Since(start).Milliseconds()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return &HealthResult{
			Status:  "unhealthy",
			Message: "Authentication failed — check client credentials",
		}, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return &HealthResult{
			Status:  "unhealthy",
			Message: fmt.Sprintf("Sentinel returned HTTP %d: %s", resp.StatusCode, string(body)),
		}, nil
	}

	log.Info().Int64("latency_ms", latency).Msg("Sentinel health check passed")
	return &HealthResult{
		Status:  "healthy",
		Message: fmt.Sprintf("Sentinel API accessible (%dms)", latency),
	}, nil
}

// Sync pulls security incidents and alerts from Microsoft Sentinel.
func (s *SentinelAdapter) Sync(ctx context.Context) (*SyncResultAdapter, error) {
	if err := s.ensureToken(ctx); err != nil {
		return nil, fmt.Errorf("token acquisition failed: %w", err)
	}

	result := &SyncResultAdapter{}

	// 1. Security incidents
	incidents, err := s.syncIncidents(ctx)
	if err != nil {
		log.Warn().Err(err).Msg("Sentinel incidents sync failed")
		result.Error = err.Error()
	} else {
		result.RecordsProcessed += incidents.RecordsProcessed
		result.RecordsCreated += incidents.RecordsCreated
	}

	// 2. Security alerts
	alerts, err := s.syncAlerts(ctx)
	if err != nil {
		log.Warn().Err(err).Msg("Sentinel alerts sync failed")
		if result.Error == "" {
			result.Error = err.Error()
		}
	} else {
		result.RecordsProcessed += alerts.RecordsProcessed
		result.RecordsCreated += alerts.RecordsCreated
	}

	log.Info().
		Int("processed", result.RecordsProcessed).
		Int("created", result.RecordsCreated).
		Msg("Sentinel sync completed")

	return result, nil
}

// syncIncidents fetches active Sentinel incidents.
func (s *SentinelAdapter) syncIncidents(ctx context.Context) (*SyncResultAdapter, error) {
	incidentsURL := fmt.Sprintf(
		"https://management.azure.com/subscriptions/%s/resourceGroups/%s/providers/Microsoft.OperationalInsights/workspaces/%s/providers/Microsoft.SecurityInsights/incidents?api-version=2023-11-01&$top=100&$filter=properties/status ne 'Closed'&$orderby=properties/createdTimeUtc desc",
		s.config.SubscriptionID,
		s.config.ResourceGroup,
		s.config.WorkspaceName,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, incidentsURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+s.accessToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("incidents request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Sentinel returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	var incidentsResp struct {
		Value []struct {
			ID         string `json:"id"`
			Name       string `json:"name"`
			Properties struct {
				Title    string `json:"title"`
				Severity string `json:"severity"`
				Status   string `json:"status"`
			} `json:"properties"`
		} `json:"value"`
	}
	if err := json.Unmarshal(body, &incidentsResp); err != nil {
		return nil, fmt.Errorf("failed to parse incidents response: %w", err)
	}

	return &SyncResultAdapter{
		RecordsProcessed: len(incidentsResp.Value),
		RecordsCreated:   len(incidentsResp.Value),
	}, nil
}

// syncAlerts fetches Sentinel security alerts.
func (s *SentinelAdapter) syncAlerts(ctx context.Context) (*SyncResultAdapter, error) {
	alertsURL := fmt.Sprintf(
		"https://management.azure.com/subscriptions/%s/resourceGroups/%s/providers/Microsoft.OperationalInsights/workspaces/%s/providers/Microsoft.SecurityInsights/alertRules?api-version=2023-11-01",
		s.config.SubscriptionID,
		s.config.ResourceGroup,
		s.config.WorkspaceName,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, alertsURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+s.accessToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("alert rules request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Sentinel returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	var alertsResp struct {
		Value []interface{} `json:"value"`
	}
	if err := json.Unmarshal(body, &alertsResp); err != nil {
		return nil, fmt.Errorf("failed to parse alerts response: %w", err)
	}

	return &SyncResultAdapter{
		RecordsProcessed: len(alertsResp.Value),
		RecordsCreated:   len(alertsResp.Value),
	}, nil
}

// ensureToken acquires or refreshes the OAuth2 client_credentials token.
func (s *SentinelAdapter) ensureToken(ctx context.Context) error {
	if s.accessToken != "" && time.Now().Before(s.tokenExpiry) {
		return nil
	}

	tokenURL := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", s.config.TenantID)
	data := url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {s.config.ClientID},
		"client_secret": {s.config.ClientSecret},
		"scope":         {"https://management.azure.com/.default"},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
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
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return fmt.Errorf("failed to parse token response: %w", err)
	}

	s.accessToken = tokenResp.AccessToken
	s.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn-60) * time.Second)

	return nil
}
