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
// SPLUNK SIEM ADAPTER
// ============================================================

// SplunkConfig holds credentials for the Splunk REST API.
type SplunkConfig struct {
	BaseURL   string `json:"base_url"`   // e.g. https://splunk.corp.com:8089
	Token     string `json:"token"`      // Bearer token or session key
	Username  string `json:"username"`
	Password  string `json:"password"`
	Index     string `json:"index"`      // Default index to search
	VerifyTLS bool   `json:"verify_tls"` // Whether to verify TLS certs
}

// SplunkAdapter implements the integration adapter for Splunk SIEM.
type SplunkAdapter struct {
	config     SplunkConfig
	httpClient *http.Client
	sessionKey string
}

func NewSplunkAdapter() *SplunkAdapter {
	return &SplunkAdapter{
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

func (s *SplunkAdapter) Type() string {
	return "siem_splunk"
}

func (s *SplunkAdapter) Connect(_ context.Context, config map[string]interface{}) error {
	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal Splunk config: %w", err)
	}
	if err := json.Unmarshal(data, &s.config); err != nil {
		return fmt.Errorf("failed to parse Splunk config: %w", err)
	}

	if s.config.BaseURL == "" {
		return fmt.Errorf("base_url is required")
	}
	s.config.BaseURL = strings.TrimRight(s.config.BaseURL, "/")

	if s.config.Index == "" {
		s.config.Index = "main"
	}

	return nil
}

func (s *SplunkAdapter) Disconnect(_ context.Context) error {
	s.config = SplunkConfig{}
	s.sessionKey = ""
	return nil
}

// HealthCheck verifies connectivity to the Splunk REST API.
func (s *SplunkAdapter) HealthCheck(ctx context.Context) (*HealthResult, error) {
	start := time.Now()

	// Authenticate if needed
	if err := s.authenticate(ctx); err != nil {
		return &HealthResult{
			Status:  "unhealthy",
			Message: fmt.Sprintf("Authentication failed: %v", err),
		}, nil
	}

	// Check server info endpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		s.config.BaseURL+"/services/server/info?output_mode=json", nil)
	if err != nil {
		return &HealthResult{Status: "unhealthy", Message: err.Error()}, nil
	}
	s.setAuth(req)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return &HealthResult{
			Status:  "unhealthy",
			Message: fmt.Sprintf("Splunk unreachable: %v", err),
		}, nil
	}
	defer resp.Body.Close()

	latency := time.Since(start).Milliseconds()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return &HealthResult{
			Status:  "unhealthy",
			Message: fmt.Sprintf("Splunk returned HTTP %d: %s", resp.StatusCode, string(body)),
		}, nil
	}

	// Parse server info
	var serverInfo struct {
		Entry []struct {
			Content struct {
				Version    string `json:"version"`
				ServerName string `json:"serverName"`
			} `json:"content"`
		} `json:"entry"`
	}
	body, _ := io.ReadAll(resp.Body)
	_ = json.Unmarshal(body, &serverInfo)

	version := "unknown"
	if len(serverInfo.Entry) > 0 {
		version = serverInfo.Entry[0].Content.Version
	}

	log.Info().Str("version", version).Int64("latency_ms", latency).Msg("Splunk health check passed")
	return &HealthResult{
		Status:  "healthy",
		Message: fmt.Sprintf("Splunk v%s reachable (%dms)", version, latency),
	}, nil
}

// Sync queries Splunk for security events and notable events.
func (s *SplunkAdapter) Sync(ctx context.Context) (*SyncResultAdapter, error) {
	if err := s.authenticate(ctx); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	result := &SyncResultAdapter{}

	// 1. Search for security events
	securityEvents, err := s.searchSecurityEvents(ctx)
	if err != nil {
		log.Warn().Err(err).Msg("Splunk security events search failed")
		result.Error = err.Error()
	} else {
		result.RecordsProcessed += securityEvents.RecordsProcessed
		result.RecordsCreated += securityEvents.RecordsCreated
	}

	// 2. Search for notable events (incidents)
	notableEvents, err := s.searchNotableEvents(ctx)
	if err != nil {
		log.Warn().Err(err).Msg("Splunk notable events search failed")
		if result.Error == "" {
			result.Error = err.Error()
		}
	} else {
		result.RecordsProcessed += notableEvents.RecordsProcessed
		result.RecordsCreated += notableEvents.RecordsCreated
	}

	log.Info().
		Int("processed", result.RecordsProcessed).
		Int("created", result.RecordsCreated).
		Msg("Splunk sync completed")

	return result, nil
}

// authenticate obtains a session key from Splunk.
func (s *SplunkAdapter) authenticate(ctx context.Context) error {
	// If we already have a token, use that
	if s.config.Token != "" {
		s.sessionKey = s.config.Token
		return nil
	}

	if s.sessionKey != "" {
		return nil // Reuse existing session
	}

	if s.config.Username == "" || s.config.Password == "" {
		return fmt.Errorf("no token or username/password provided")
	}

	body := fmt.Sprintf("username=%s&password=%s&output_mode=json",
		s.config.Username, s.config.Password)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		s.config.BaseURL+"/services/auth/login", strings.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("Splunk auth request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Splunk auth failed (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	var authResp struct {
		SessionKey string `json:"sessionKey"`
	}
	if err := json.Unmarshal(respBody, &authResp); err != nil {
		return fmt.Errorf("failed to parse auth response: %w", err)
	}

	s.sessionKey = authResp.SessionKey
	return nil
}

// setAuth sets the appropriate authentication header on a request.
func (s *SplunkAdapter) setAuth(req *http.Request) {
	if s.config.Token != "" {
		req.Header.Set("Authorization", "Bearer "+s.config.Token)
	} else if s.sessionKey != "" {
		req.Header.Set("Authorization", "Splunk "+s.sessionKey)
	}
}

// searchSecurityEvents runs a search for recent security events.
func (s *SplunkAdapter) searchSecurityEvents(ctx context.Context) (*SyncResultAdapter, error) {
	searchQuery := fmt.Sprintf(
		`search index=%s sourcetype=*security* earliest=-24h | stats count by source, sourcetype, host | head 100`,
		s.config.Index,
	)
	return s.executeSearch(ctx, searchQuery)
}

// searchNotableEvents runs a search for notable events (ES).
func (s *SplunkAdapter) searchNotableEvents(ctx context.Context) (*SyncResultAdapter, error) {
	searchQuery := `| search index=notable status!="closed" earliest=-7d | head 100`
	return s.executeSearch(ctx, searchQuery)
}

// executeSearch creates a oneshot search job and reads results.
func (s *SplunkAdapter) executeSearch(ctx context.Context, query string) (*SyncResultAdapter, error) {
	body := fmt.Sprintf("search=%s&output_mode=json&exec_mode=oneshot&count=100",
		query)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		s.config.BaseURL+"/services/search/jobs/export", strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	s.setAuth(req)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	result := &SyncResultAdapter{}

	if resp.StatusCode == http.StatusOK {
		var searchResp struct {
			Results []interface{} `json:"results"`
		}
		if err := json.Unmarshal(respBody, &searchResp); err == nil {
			result.RecordsProcessed = len(searchResp.Results)
			result.RecordsCreated = len(searchResp.Results)
		}
	}

	return result, nil
}
