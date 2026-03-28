package integrations

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

// ============================================================
// SERVICENOW ITSM ADAPTER
// ============================================================

// ServiceNowConfig holds credentials for ServiceNow REST API.
type ServiceNowConfig struct {
	InstanceURL  string `json:"instance_url"`  // e.g. https://mycompany.service-now.com
	Username     string `json:"username"`
	Password     string `json:"password"`
	ClientID     string `json:"client_id,omitempty"`
	ClientSecret string `json:"client_secret,omitempty"`
	Table        string `json:"table,omitempty"` // Default table (e.g. "incident")
}

// ServiceNowIncident represents a ServiceNow incident record.
type ServiceNowIncident struct {
	SysID            string `json:"sys_id,omitempty"`
	Number           string `json:"number,omitempty"`
	ShortDescription string `json:"short_description"`
	Description      string `json:"description,omitempty"`
	Priority         string `json:"priority,omitempty"`
	State            string `json:"state,omitempty"`
	Category         string `json:"category,omitempty"`
	AssignmentGroup  string `json:"assignment_group,omitempty"`
	AssignedTo       string `json:"assigned_to,omitempty"`
	Impact           string `json:"impact,omitempty"`
	Urgency          string `json:"urgency,omitempty"`
	CallerID         string `json:"caller_id,omitempty"`
	OpenedAt         string `json:"opened_at,omitempty"`
	ResolvedAt       string `json:"resolved_at,omitempty"`
	ClosedAt         string `json:"closed_at,omitempty"`
}

// ServiceNowAdapter implements the integration adapter for ServiceNow ITSM.
type ServiceNowAdapter struct {
	config     ServiceNowConfig
	httpClient *http.Client
	oauthToken string
	tokenExpiry time.Time
}

func NewServiceNowAdapter() *ServiceNowAdapter {
	return &ServiceNowAdapter{
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (s *ServiceNowAdapter) Type() string {
	return "itsm_servicenow"
}

func (s *ServiceNowAdapter) Connect(_ context.Context, config map[string]interface{}) error {
	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal ServiceNow config: %w", err)
	}
	if err := json.Unmarshal(data, &s.config); err != nil {
		return fmt.Errorf("failed to parse ServiceNow config: %w", err)
	}

	if s.config.InstanceURL == "" {
		return fmt.Errorf("instance_url is required")
	}
	if s.config.Table == "" {
		s.config.Table = "incident"
	}

	return nil
}

func (s *ServiceNowAdapter) Disconnect(_ context.Context) error {
	s.config = ServiceNowConfig{}
	s.oauthToken = ""
	return nil
}

// HealthCheck verifies connectivity to ServiceNow.
func (s *ServiceNowAdapter) HealthCheck(ctx context.Context) (*HealthResult, error) {
	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		s.config.InstanceURL+"/api/now/table/sys_properties?sysparm_limit=1", nil)
	if err != nil {
		return &HealthResult{Status: "unhealthy", Message: err.Error()}, nil
	}
	s.setAuth(req)
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return &HealthResult{
			Status:  "unhealthy",
			Message: fmt.Sprintf("ServiceNow unreachable: %v", err),
		}, nil
	}
	defer resp.Body.Close()

	latency := time.Since(start).Milliseconds()

	if resp.StatusCode == http.StatusUnauthorized {
		return &HealthResult{
			Status:  "unhealthy",
			Message: "Authentication failed — check credentials",
		}, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return &HealthResult{
			Status:  "degraded",
			Message: fmt.Sprintf("ServiceNow returned HTTP %d: %s", resp.StatusCode, string(body)),
		}, nil
	}

	log.Info().Int64("latency_ms", latency).Msg("ServiceNow health check passed")
	return &HealthResult{
		Status:  "healthy",
		Message: fmt.Sprintf("ServiceNow instance reachable (%dms)", latency),
	}, nil
}

// Sync fetches open incidents from ServiceNow and syncs statuses.
func (s *ServiceNowAdapter) Sync(ctx context.Context) (*SyncResultAdapter, error) {
	result := &SyncResultAdapter{}

	// Fetch open incidents
	incidents, err := s.ListIncidents(ctx, "state!=7^stateNOT IN8", 200)
	if err != nil {
		return nil, fmt.Errorf("failed to list incidents: %w", err)
	}

	result.RecordsProcessed = len(incidents)
	result.RecordsCreated = len(incidents)

	log.Info().
		Int("incidents_found", len(incidents)).
		Msg("ServiceNow sync completed")

	return result, nil
}

// ============================================================
// CRUD OPERATIONS
// ============================================================

// CreateIncident creates a new incident in ServiceNow.
func (s *ServiceNowAdapter) CreateIncident(ctx context.Context, incident ServiceNowIncident) (*ServiceNowIncident, error) {
	payload, err := json.Marshal(incident)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal incident: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		s.config.InstanceURL+"/api/now/table/"+s.config.Table, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	s.setAuth(req)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("create incident request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ServiceNow returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Result ServiceNowIncident `json:"result"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	log.Info().Str("number", result.Result.Number).Msg("ServiceNow incident created")
	return &result.Result, nil
}

// UpdateIncident updates an existing incident.
func (s *ServiceNowAdapter) UpdateIncident(ctx context.Context, sysID string, update ServiceNowIncident) (*ServiceNowIncident, error) {
	payload, err := json.Marshal(update)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal update: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch,
		fmt.Sprintf("%s/api/now/table/%s/%s", s.config.InstanceURL, s.config.Table, sysID),
		bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	s.setAuth(req)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("update incident request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ServiceNow returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Result ServiceNowIncident `json:"result"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result.Result, nil
}

// ListIncidents queries incidents matching a filter query.
func (s *ServiceNowAdapter) ListIncidents(ctx context.Context, query string, limit int) ([]ServiceNowIncident, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	url := fmt.Sprintf("%s/api/now/table/%s?sysparm_query=%s&sysparm_limit=%d&sysparm_display_value=true",
		s.config.InstanceURL, s.config.Table, query, limit)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	s.setAuth(req)
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("list incidents request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ServiceNow returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Result []ServiceNowIncident `json:"result"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse incidents: %w", err)
	}

	return result.Result, nil
}

// GetIncident fetches a single incident by sys_id.
func (s *ServiceNowAdapter) GetIncident(ctx context.Context, sysID string) (*ServiceNowIncident, error) {
	url := fmt.Sprintf("%s/api/now/table/%s/%s", s.config.InstanceURL, s.config.Table, sysID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	s.setAuth(req)
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get incident request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ServiceNow returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Result ServiceNowIncident `json:"result"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse incident: %w", err)
	}

	return &result.Result, nil
}

// setAuth sets basic auth or OAuth headers on a request.
func (s *ServiceNowAdapter) setAuth(req *http.Request) {
	if s.oauthToken != "" {
		req.Header.Set("Authorization", "Bearer "+s.oauthToken)
	} else if s.config.Username != "" && s.config.Password != "" {
		req.SetBasicAuth(s.config.Username, s.config.Password)
	}
}
