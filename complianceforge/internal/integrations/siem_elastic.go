package integrations

import (
	"bytes"
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
// ELASTIC / ELK STACK SIEM ADAPTER
// ============================================================

// ElasticConfig holds credentials for the Elasticsearch / Elastic SIEM API.
type ElasticConfig struct {
	BaseURL   string `json:"base_url"`   // e.g. https://elastic.corp.com:9200
	Username  string `json:"username"`
	Password  string `json:"password"`
	APIKey    string `json:"api_key"`    // Alternative to username/password
	CloudID   string `json:"cloud_id"`   // Elastic Cloud deployment ID
	Index     string `json:"index"`      // Default index pattern (e.g. ".siem-signals-*")
	VerifyTLS bool   `json:"verify_tls"`
}

// ElasticAdapter implements the integration adapter for Elastic SIEM (ELK Stack).
type ElasticAdapter struct {
	config     ElasticConfig
	httpClient *http.Client
}

func NewElasticAdapter() *ElasticAdapter {
	return &ElasticAdapter{
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

func (e *ElasticAdapter) Type() string {
	return "siem_elastic"
}

func (e *ElasticAdapter) Connect(_ context.Context, config map[string]interface{}) error {
	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal Elastic config: %w", err)
	}
	if err := json.Unmarshal(data, &e.config); err != nil {
		return fmt.Errorf("failed to parse Elastic config: %w", err)
	}

	if e.config.BaseURL == "" {
		return fmt.Errorf("base_url is required")
	}
	e.config.BaseURL = strings.TrimRight(e.config.BaseURL, "/")

	if e.config.Index == "" {
		e.config.Index = ".siem-signals-*"
	}

	return nil
}

func (e *ElasticAdapter) Disconnect(_ context.Context) error {
	e.config = ElasticConfig{}
	return nil
}

// HealthCheck verifies connectivity to the Elasticsearch cluster.
func (e *ElasticAdapter) HealthCheck(ctx context.Context) (*HealthResult, error) {
	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		e.config.BaseURL+"/_cluster/health", nil)
	if err != nil {
		return &HealthResult{Status: "unhealthy", Message: err.Error()}, nil
	}
	e.setAuth(req)

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return &HealthResult{
			Status:  "unhealthy",
			Message: fmt.Sprintf("Elasticsearch unreachable: %v", err),
		}, nil
	}
	defer resp.Body.Close()

	latency := time.Since(start).Milliseconds()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return &HealthResult{
			Status:  "unhealthy",
			Message: "Authentication failed — check credentials or API key",
		}, nil
	}

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return &HealthResult{
			Status:  "unhealthy",
			Message: fmt.Sprintf("Elasticsearch returned HTTP %d: %s", resp.StatusCode, string(body)),
		}, nil
	}

	var clusterHealth struct {
		ClusterName string `json:"cluster_name"`
		Status      string `json:"status"`
		NumNodes    int    `json:"number_of_nodes"`
	}
	_ = json.Unmarshal(body, &clusterHealth)

	esStatus := "healthy"
	if clusterHealth.Status == "red" {
		esStatus = "degraded"
	}

	log.Info().
		Str("cluster", clusterHealth.ClusterName).
		Str("es_status", clusterHealth.Status).
		Int("nodes", clusterHealth.NumNodes).
		Int64("latency_ms", latency).
		Msg("Elastic health check passed")

	return &HealthResult{
		Status:  esStatus,
		Message: fmt.Sprintf("Cluster %q status=%s nodes=%d (%dms)", clusterHealth.ClusterName, clusterHealth.Status, clusterHealth.NumNodes, latency),
	}, nil
}

// Sync queries Elastic SIEM for security signals and detection alerts.
func (e *ElasticAdapter) Sync(ctx context.Context) (*SyncResultAdapter, error) {
	result := &SyncResultAdapter{}

	// 1. SIEM detection signals
	signals, err := e.searchSIEMSignals(ctx)
	if err != nil {
		log.Warn().Err(err).Msg("Elastic SIEM signals search failed")
		result.Error = err.Error()
	} else {
		result.RecordsProcessed += signals.RecordsProcessed
		result.RecordsCreated += signals.RecordsCreated
	}

	// 2. Security detection alerts
	alerts, err := e.searchDetectionAlerts(ctx)
	if err != nil {
		log.Warn().Err(err).Msg("Elastic detection alerts search failed")
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
		Msg("Elastic sync completed")

	return result, nil
}

// searchSIEMSignals queries the SIEM signals index for open alerts.
func (e *ElasticAdapter) searchSIEMSignals(ctx context.Context) (*SyncResultAdapter, error) {
	query := map[string]interface{}{
		"size": 100,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{
						"range": map[string]interface{}{
							"@timestamp": map[string]string{
								"gte": "now-24h",
							},
						},
					},
				},
				"must_not": []map[string]interface{}{
					{
						"term": map[string]interface{}{
							"signal.status": "closed",
						},
					},
				},
			},
		},
		"sort": []map[string]interface{}{
			{"@timestamp": map[string]string{"order": "desc"}},
		},
	}

	return e.executeSearch(ctx, e.config.Index, query)
}

// searchDetectionAlerts queries for Elastic Security detection rule alerts.
func (e *ElasticAdapter) searchDetectionAlerts(ctx context.Context) (*SyncResultAdapter, error) {
	query := map[string]interface{}{
		"size": 100,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"filter": []map[string]interface{}{
					{
						"range": map[string]interface{}{
							"@timestamp": map[string]string{
								"gte": "now-7d",
							},
						},
					},
					{
						"term": map[string]interface{}{
							"event.kind": "signal",
						},
					},
				},
			},
		},
		"sort": []map[string]interface{}{
			{"@timestamp": map[string]string{"order": "desc"}},
		},
	}

	return e.executeSearch(ctx, ".alerts-security.alerts-*", query)
}

// executeSearch runs an Elasticsearch _search query.
func (e *ElasticAdapter) executeSearch(ctx context.Context, index string, query interface{}) (*SyncResultAdapter, error) {
	payload, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search query: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/%s/_search", e.config.BaseURL, index),
		bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	e.setAuth(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	result := &SyncResultAdapter{}

	if resp.StatusCode == http.StatusOK {
		var searchResp struct {
			Hits struct {
				Total struct {
					Value int `json:"value"`
				} `json:"total"`
				Hits []interface{} `json:"hits"`
			} `json:"hits"`
		}
		if err := json.Unmarshal(body, &searchResp); err == nil {
			result.RecordsProcessed = len(searchResp.Hits.Hits)
			result.RecordsCreated = len(searchResp.Hits.Hits)
		}
	} else {
		return nil, fmt.Errorf("Elasticsearch returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	return result, nil
}

// setAuth sets the appropriate authentication header.
func (e *ElasticAdapter) setAuth(req *http.Request) {
	if e.config.APIKey != "" {
		req.Header.Set("Authorization", "ApiKey "+e.config.APIKey)
	} else if e.config.Username != "" {
		req.SetBasicAuth(e.config.Username, e.config.Password)
	}
	req.Header.Set("Accept", "application/json")
}
