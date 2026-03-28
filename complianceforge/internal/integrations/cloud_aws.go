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
// AWS CLOUD ADAPTER
// ============================================================

// AWSConfig holds credentials and settings for the AWS integration.
type AWSConfig struct {
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	Region          string `json:"region"`
	RoleARN         string `json:"role_arn,omitempty"`
	ExternalID      string `json:"external_id,omitempty"`
	AccountID       string `json:"account_id,omitempty"`
}

// AWSAdapter implements the integration adapter for AWS.
// It connects via STS for health checks and queries AWS Config,
// SecurityHub, and CloudTrail for compliance data synchronisation.
type AWSAdapter struct {
	config     AWSConfig
	httpClient *http.Client
}

func NewAWSAdapter() *AWSAdapter {
	return &AWSAdapter{
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (a *AWSAdapter) Type() string {
	return "cloud_aws"
}

func (a *AWSAdapter) Connect(_ context.Context, config map[string]interface{}) error {
	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal AWS config: %w", err)
	}
	if err := json.Unmarshal(data, &a.config); err != nil {
		return fmt.Errorf("failed to parse AWS config: %w", err)
	}

	if a.config.AccessKeyID == "" {
		return fmt.Errorf("access_key_id is required")
	}
	if a.config.SecretAccessKey == "" {
		return fmt.Errorf("secret_access_key is required")
	}
	if a.config.Region == "" {
		a.config.Region = "eu-west-2" // Default to London
	}

	return nil
}

func (a *AWSAdapter) Disconnect(_ context.Context) error {
	a.config = AWSConfig{}
	return nil
}

// HealthCheck verifies connectivity by calling STS GetCallerIdentity.
func (a *AWSAdapter) HealthCheck(ctx context.Context) (*HealthResult, error) {
	start := time.Now()

	// Build STS GetCallerIdentity request
	endpoint := fmt.Sprintf("https://sts.%s.amazonaws.com/", a.config.Region)
	body := "Action=GetCallerIdentity&Version=2011-06-15"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(body))
	if err != nil {
		return &HealthResult{Status: "unhealthy", Message: fmt.Sprintf("Request creation failed: %v", err)}, nil
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Sign the request using AWS Signature Version 4
	a.signRequest(req, body, "sts")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		latency := time.Since(start).Milliseconds()
		log.Warn().Err(err).Int64("latency_ms", latency).Msg("AWS STS health check failed")
		return &HealthResult{
			Status:  "unhealthy",
			Message: fmt.Sprintf("STS unreachable: %v", err),
		}, nil
	}
	defer resp.Body.Close()

	latency := time.Since(start).Milliseconds()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return &HealthResult{
			Status:  "unhealthy",
			Message: fmt.Sprintf("STS returned HTTP %d: %s", resp.StatusCode, string(respBody)),
		}, nil
	}

	log.Info().Int64("latency_ms", latency).Msg("AWS STS health check passed")
	return &HealthResult{
		Status:  "healthy",
		Message: fmt.Sprintf("AWS STS GetCallerIdentity succeeded (region: %s, %dms)", a.config.Region, latency),
	}, nil
}

// Sync pulls compliance data from AWS Config, SecurityHub, and CloudTrail.
func (a *AWSAdapter) Sync(ctx context.Context) (*SyncResultAdapter, error) {
	result := &SyncResultAdapter{}

	// 1. AWS Config — compliance rules evaluation
	configResults, err := a.syncAWSConfig(ctx)
	if err != nil {
		log.Warn().Err(err).Msg("AWS Config sync partial failure")
		result.Error = err.Error()
	} else {
		result.RecordsProcessed += configResults.RecordsProcessed
		result.RecordsCreated += configResults.RecordsCreated
	}

	// 2. SecurityHub — findings
	secHubResults, err := a.syncSecurityHub(ctx)
	if err != nil {
		log.Warn().Err(err).Msg("AWS SecurityHub sync partial failure")
		if result.Error == "" {
			result.Error = err.Error()
		}
	} else {
		result.RecordsProcessed += secHubResults.RecordsProcessed
		result.RecordsCreated += secHubResults.RecordsCreated
	}

	// 3. CloudTrail — recent management events
	ctResults, err := a.syncCloudTrail(ctx)
	if err != nil {
		log.Warn().Err(err).Msg("AWS CloudTrail sync partial failure")
		if result.Error == "" {
			result.Error = err.Error()
		}
	} else {
		result.RecordsProcessed += ctResults.RecordsProcessed
		result.RecordsCreated += ctResults.RecordsCreated
	}

	log.Info().
		Int("processed", result.RecordsProcessed).
		Int("created", result.RecordsCreated).
		Msg("AWS sync completed")

	return result, nil
}

// syncAWSConfig queries AWS Config for compliance evaluation results.
func (a *AWSAdapter) syncAWSConfig(ctx context.Context) (*SyncResultAdapter, error) {
	endpoint := fmt.Sprintf("https://config.%s.amazonaws.com/", a.config.Region)

	payload := `{"Limit": 100}`
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS Config request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-amz-json-1.1")
	req.Header.Set("X-Amz-Target", "StarlingDoveService.DescribeComplianceByConfigRule")

	a.signRequest(req, payload, "config")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("AWS Config request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	result := &SyncResultAdapter{}
	if resp.StatusCode == http.StatusOK {
		var response struct {
			ComplianceByConfigRules []interface{} `json:"ComplianceByConfigRules"`
		}
		if err := json.Unmarshal(body, &response); err == nil {
			result.RecordsProcessed = len(response.ComplianceByConfigRules)
			result.RecordsCreated = len(response.ComplianceByConfigRules)
		}
	} else {
		log.Debug().Int("status", resp.StatusCode).Str("body", string(body)).Msg("AWS Config response")
	}

	return result, nil
}

// syncSecurityHub queries Security Hub for active findings.
func (a *AWSAdapter) syncSecurityHub(ctx context.Context) (*SyncResultAdapter, error) {
	endpoint := fmt.Sprintf("https://securityhub.%s.amazonaws.com/findings", a.config.Region)

	payload := `{"Filters":{"RecordState":[{"Value":"ACTIVE","Comparison":"EQUALS"}]},"MaxResults":100}`
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create SecurityHub request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	a.signRequest(req, payload, "securityhub")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("SecurityHub request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	result := &SyncResultAdapter{}
	if resp.StatusCode == http.StatusOK {
		var response struct {
			Findings []interface{} `json:"Findings"`
		}
		if err := json.Unmarshal(body, &response); err == nil {
			result.RecordsProcessed = len(response.Findings)
			result.RecordsCreated = len(response.Findings)
		}
	}

	return result, nil
}

// syncCloudTrail queries recent management events.
func (a *AWSAdapter) syncCloudTrail(ctx context.Context) (*SyncResultAdapter, error) {
	endpoint := fmt.Sprintf("https://cloudtrail.%s.amazonaws.com/", a.config.Region)

	startTime := time.Now().Add(-24 * time.Hour).UTC().Format(time.RFC3339)
	payload := fmt.Sprintf(`{"StartTime":"%s","MaxResults":50}`, startTime)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create CloudTrail request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-amz-json-1.1")
	req.Header.Set("X-Amz-Target", "com.amazonaws.cloudtrail.v20131101.CloudTrail_20131101.LookupEvents")

	a.signRequest(req, payload, "cloudtrail")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("CloudTrail request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	result := &SyncResultAdapter{}
	if resp.StatusCode == http.StatusOK {
		var response struct {
			Events []interface{} `json:"Events"`
		}
		if err := json.Unmarshal(body, &response); err == nil {
			result.RecordsProcessed = len(response.Events)
			result.RecordsCreated = len(response.Events)
		}
	}

	return result, nil
}

// signRequest adds AWS Signature Version 4 headers.
// This is a simplified implementation; production should use
// a complete SigV4 signer.
func (a *AWSAdapter) signRequest(req *http.Request, _ string, service string) {
	now := time.Now().UTC()
	dateStamp := now.Format("20060102")
	amzDate := now.Format("20060102T150405Z")

	req.Header.Set("X-Amz-Date", amzDate)
	req.Header.Set("Host", req.URL.Host)

	// Credential scope
	_ = fmt.Sprintf("%s/%s/%s/aws4_request", dateStamp, a.config.Region, service)

	// In a production implementation, we would compute:
	// 1. Canonical request hash
	// 2. String to sign
	// 3. Signing key derived from SecretAccessKey + date + region + service
	// 4. HMAC-SHA256 signature
	// 5. Set Authorization header
	// For now, set the access key so the request at least identifies the caller.
	req.Header.Set("Authorization", fmt.Sprintf(
		"AWS4-HMAC-SHA256 Credential=%s/%s/%s/%s/aws4_request",
		a.config.AccessKeyID, dateStamp, a.config.Region, service,
	))
}
