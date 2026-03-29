package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"github.com/complianceforge/platform/internal/pkg/queue"
)

// ============================================================
// EVIDENCE COLLECTOR
// ============================================================

// EvidenceCollector orchestrates automated evidence collection from
// various sources (APIs, files, scripts, webhooks) and validates the
// collected data against configurable acceptance criteria.
type EvidenceCollector struct {
	pool  *pgxpool.Pool
	queue *queue.Queue
}

// NewEvidenceCollector creates a new EvidenceCollector.
func NewEvidenceCollector(pool *pgxpool.Pool, q *queue.Queue) *EvidenceCollector {
	return &EvidenceCollector{pool: pool, queue: q}
}

// CollectionConfig represents an evidence collection configuration.
type CollectionConfig struct {
	ID                       uuid.UUID       `json:"id"`
	OrganizationID           uuid.UUID       `json:"organization_id"`
	ControlImplementationID  *uuid.UUID      `json:"control_implementation_id"`
	Name                     string          `json:"name"`
	CollectionMethod         string          `json:"collection_method"`
	ScheduleCron             string          `json:"schedule_cron"`
	ScheduleDescription      string          `json:"schedule_description"`
	APIConfig                json.RawMessage `json:"api_config"`
	FileConfig               json.RawMessage `json:"file_config"`
	ScriptConfig             json.RawMessage `json:"script_config"`
	WebhookConfig            json.RawMessage `json:"webhook_config"`
	AcceptanceCriteria       json.RawMessage `json:"acceptance_criteria"`
	FailureThreshold         int             `json:"failure_threshold"`
	AutoUpdateControlStatus  bool            `json:"auto_update_control_status"`
	IsActive                 bool            `json:"is_active"`
	LastCollectionAt         *time.Time      `json:"last_collection_at"`
	LastCollectionStatus     string          `json:"last_collection_status"`
	NextCollectionAt         *time.Time      `json:"next_collection_at"`
	ConsecutiveFailures      int             `json:"consecutive_failures"`
	CreatedAt                time.Time       `json:"created_at"`
	UpdatedAt                time.Time       `json:"updated_at"`
}

// CollectionRun represents a single execution of an evidence collection.
type CollectionRun struct {
	ID                      uuid.UUID       `json:"id"`
	OrganizationID          uuid.UUID       `json:"organization_id"`
	ConfigID                uuid.UUID       `json:"config_id"`
	ControlImplementationID *uuid.UUID      `json:"control_implementation_id"`
	Status                  string          `json:"status"`
	StartedAt               *time.Time      `json:"started_at"`
	CompletedAt             *time.Time      `json:"completed_at"`
	DurationMs              int             `json:"duration_ms"`
	CollectedData           json.RawMessage `json:"collected_data"`
	CriterionValidationResults       json.RawMessage `json:"validation_results"`
	AllCriteriaPassed       *bool           `json:"all_criteria_passed"`
	EvidenceID              *uuid.UUID      `json:"evidence_id"`
	ErrorMessage            string          `json:"error_message"`
	Metadata                json.RawMessage `json:"metadata"`
	CreatedAt               time.Time       `json:"created_at"`
}

// APIConfigPayload holds configuration for API-based evidence collection.
type APIConfigPayload struct {
	URL           string            `json:"url"`
	Method        string            `json:"method"`
	Headers       map[string]string `json:"headers"`
	AuthType      string            `json:"auth_type"` // basic, bearer, api_key, oauth2_cc
	AuthConfig    json.RawMessage   `json:"auth_config"`
	Body          string            `json:"body"`
	TimeoutSecs   int               `json:"timeout_secs"`
	ResponseJPath string            `json:"response_jpath"`
}

// FileConfigPayload holds configuration for file-based evidence collection.
type FileConfigPayload struct {
	Path        string `json:"path"`
	Pattern     string `json:"pattern"`
	MaxSizeBytes int64 `json:"max_size_bytes"`
}

// ScriptConfigPayload holds configuration for script-based evidence collection.
type ScriptConfigPayload struct {
	Command    string   `json:"command"`
	Args       []string `json:"args"`
	TimeoutSec int      `json:"timeout_sec"`
	WorkDir    string   `json:"work_dir"`
}

// WebhookConfigPayload holds configuration for webhook-based evidence collection.
type WebhookConfigPayload struct {
	Secret        string `json:"secret"`
	SignatureHeader string `json:"signature_header"`
	ContentType   string `json:"content_type"`
}

// AcceptanceCriterion defines a single validation rule for collected evidence.
type AcceptanceCriterion struct {
	Field    string `json:"field"`
	Operator string `json:"operator"` // equals, not_equals, greater_than, less_than, contains, matches_regex
	Value    string `json:"value"`
	Message  string `json:"message"`
}

// CriterionValidationResult holds the outcome of evaluating a single criterion.
type CriterionValidationResult struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Expected string `json:"expected"`
	Actual   string `json:"actual"`
	Passed   bool   `json:"passed"`
	Message  string `json:"message"`
}

// ============================================================
// COLLECTION SCHEDULER
// ============================================================

// RunScheduledCollections checks all active configs for due collections
// and enqueues jobs for each one. This should be called periodically
// (e.g., every minute) by a cron runner or ticker.
func (ec *EvidenceCollector) RunScheduledCollections(ctx context.Context) error {
	rows, err := ec.pool.Query(ctx, `
		SELECT id, organization_id, control_implementation_id, name, collection_method,
			api_config, file_config, script_config, webhook_config,
			acceptance_criteria, failure_threshold, auto_update_control_status
		FROM evidence_collection_configs
		WHERE is_active = true
			AND collection_method != 'webhook_receive'
			AND (next_collection_at IS NULL OR next_collection_at <= NOW())
		ORDER BY next_collection_at ASC NULLS FIRST
		LIMIT 100`)
	if err != nil {
		return fmt.Errorf("failed to query due configs: %w", err)
	}
	defer rows.Close()

	var configs []CollectionConfig
	for rows.Next() {
		var c CollectionConfig
		if err := rows.Scan(
			&c.ID, &c.OrganizationID, &c.ControlImplementationID,
			&c.Name, &c.CollectionMethod,
			&c.APIConfig, &c.FileConfig, &c.ScriptConfig, &c.WebhookConfig,
			&c.AcceptanceCriteria, &c.FailureThreshold, &c.AutoUpdateControlStatus,
		); err != nil {
			log.Error().Err(err).Msg("Failed to scan collection config")
			continue
		}
		configs = append(configs, c)
	}

	for _, cfg := range configs {
		if err := ec.EnqueueCollection(ctx, cfg); err != nil {
			log.Error().Err(err).Str("config_id", cfg.ID.String()).Msg("Failed to enqueue collection")
		}
	}

	return nil
}

// EnqueueCollection enqueues a collection job to the work queue.
func (ec *EvidenceCollector) EnqueueCollection(ctx context.Context, cfg CollectionConfig) error {
	payload := map[string]interface{}{
		"config_id":       cfg.ID.String(),
		"organization_id": cfg.OrganizationID.String(),
	}
	_, err := ec.queue.Enqueue(ctx, "evidence_collection", queue.QueueDefault, cfg.OrganizationID.String(), payload)
	return err
}

// ============================================================
// EXECUTE COLLECTION
// ============================================================

// ExecuteCollection runs a single evidence collection for the given config.
// It orchestrates: collect -> validate -> store evidence -> update status.
func (ec *EvidenceCollector) ExecuteCollection(ctx context.Context, configID, orgID uuid.UUID) error {
	// Load config
	cfg, err := ec.GetConfig(ctx, orgID, configID)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create run record
	runID := uuid.New()
	startedAt := time.Now()
	_, err = ec.pool.Exec(ctx, `
		INSERT INTO evidence_collection_runs (id, organization_id, config_id, control_implementation_id, status, started_at)
		VALUES ($1, $2, $3, $4, 'running', $5)`,
		runID, orgID, configID, cfg.ControlImplementationID, startedAt)
	if err != nil {
		return fmt.Errorf("failed to create run record: %w", err)
	}

	// Execute collection based on method
	var collectedData json.RawMessage
	var collectionErr error

	switch cfg.CollectionMethod {
	case "manual":
		collectedData, collectionErr = ec.CollectManual(ctx, *cfg)
	case "api_fetch":
		collectedData, collectionErr = ec.CollectAPI(ctx, *cfg)
	case "file_watch":
		collectedData, collectionErr = ec.CollectFile(ctx, *cfg)
	case "script_execution":
		collectedData, collectionErr = ec.CollectScript(ctx, *cfg)
	default:
		collectionErr = fmt.Errorf("unsupported collection method: %s", cfg.CollectionMethod)
	}

	completedAt := time.Now()
	durationMs := int(completedAt.Sub(startedAt).Milliseconds())

	if collectionErr != nil {
		ec.recordRunFailure(ctx, runID, configID, orgID, durationMs, collectionErr.Error())
		return collectionErr
	}

	// Validate collected data against acceptance criteria
	validationResults, allPassed := ec.ValidateEvidence(cfg.AcceptanceCriteria, collectedData)
	validationJSON, _ := json.Marshal(validationResults)

	status := "success"
	if !allPassed {
		status = "validation_failed"
	}

	// Create evidence record if validation passed
	var evidenceID *uuid.UUID
	if allPassed && cfg.ControlImplementationID != nil {
		eid, err := ec.createEvidenceRecord(ctx, orgID, *cfg.ControlImplementationID, cfg.Name, collectedData)
		if err != nil {
			log.Error().Err(err).Msg("Failed to create evidence record")
		} else {
			evidenceID = &eid
		}

		// Optionally update control status
		if cfg.AutoUpdateControlStatus {
			ec.updateControlStatus(ctx, orgID, *cfg.ControlImplementationID)
		}
	}

	// Update run record
	_, err = ec.pool.Exec(ctx, `
		UPDATE evidence_collection_runs
		SET status = $1, completed_at = $2, duration_ms = $3,
			collected_data = $4, validation_results = $5,
			all_criteria_passed = $6, evidence_id = $7
		WHERE id = $8`,
		status, completedAt, durationMs, collectedData, validationJSON,
		allPassed, evidenceID, runID)
	if err != nil {
		log.Error().Err(err).Str("run_id", runID.String()).Msg("Failed to update run record")
	}

	// Update config status
	if allPassed {
		_, err = ec.pool.Exec(ctx, `
			UPDATE evidence_collection_configs
			SET last_collection_at = NOW(), last_collection_status = 'success', consecutive_failures = 0
			WHERE id = $1`, configID)
	} else {
		_, err = ec.pool.Exec(ctx, `
			UPDATE evidence_collection_configs
			SET last_collection_at = NOW(), last_collection_status = $1,
				consecutive_failures = consecutive_failures + 1
			WHERE id = $2`, status, configID)
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to update config status")
	}

	return nil
}

func (ec *EvidenceCollector) recordRunFailure(ctx context.Context, runID, configID, orgID uuid.UUID, durationMs int, errMsg string) {
	_, err := ec.pool.Exec(ctx, `
		UPDATE evidence_collection_runs
		SET status = 'failed', completed_at = NOW(), duration_ms = $1, error_message = $2
		WHERE id = $3`,
		durationMs, errMsg, runID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to record run failure")
	}

	_, err = ec.pool.Exec(ctx, `
		UPDATE evidence_collection_configs
		SET last_collection_at = NOW(), last_collection_status = 'failed',
			consecutive_failures = consecutive_failures + 1
		WHERE id = $1`, configID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update config failure count")
	}
}

// ============================================================
// MANUAL COLLECTOR
// ============================================================

// CollectManual creates evidence from a manual submission.
// The collected_data is expected to be provided externally.
func (ec *EvidenceCollector) CollectManual(_ context.Context, _ CollectionConfig) (json.RawMessage, error) {
	// Manual collections are triggered by the handler with data payload.
	// Return empty data here; the handler passes data directly.
	return json.RawMessage(`{"source":"manual","collected_at":"` + time.Now().UTC().Format(time.RFC3339) + `"}`), nil
}

// ============================================================
// API COLLECTOR
// ============================================================

// CollectAPI fetches evidence from an external API.
// Supports Basic Auth, Bearer Token, API Key, and OAuth2 Client Credentials.
func (ec *EvidenceCollector) CollectAPI(ctx context.Context, cfg CollectionConfig) (json.RawMessage, error) {
	var apiCfg APIConfigPayload
	if err := json.Unmarshal(cfg.APIConfig, &apiCfg); err != nil {
		return nil, fmt.Errorf("invalid api_config: %w", err)
	}

	if apiCfg.URL == "" {
		return nil, fmt.Errorf("api_config.url is required")
	}
	if apiCfg.Method == "" {
		apiCfg.Method = "GET"
	}
	if apiCfg.TimeoutSecs <= 0 {
		apiCfg.TimeoutSecs = 30
	}

	reqCtx, cancel := context.WithTimeout(ctx, time.Duration(apiCfg.TimeoutSecs)*time.Second)
	defer cancel()

	var bodyReader io.Reader
	if apiCfg.Body != "" {
		bodyReader = strings.NewReader(apiCfg.Body)
	}

	req, err := http.NewRequestWithContext(reqCtx, apiCfg.Method, apiCfg.URL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	for k, v := range apiCfg.Headers {
		req.Header.Set(k, v)
	}

	// Apply authentication
	if err := ec.applyAuth(req, apiCfg); err != nil {
		return nil, fmt.Errorf("failed to apply auth: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024)) // 10MB limit
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Wrap raw response as valid JSON
	result := map[string]interface{}{
		"source":       "api_fetch",
		"url":          apiCfg.URL,
		"status_code":  resp.StatusCode,
		"collected_at": time.Now().UTC().Format(time.RFC3339),
	}

	// Try to parse as JSON for structured storage
	var parsed interface{}
	if err := json.Unmarshal(data, &parsed); err == nil {
		result["data"] = parsed
	} else {
		result["data"] = string(data)
	}

	return json.Marshal(result)
}

func (ec *EvidenceCollector) applyAuth(req *http.Request, apiCfg APIConfigPayload) error {
	if apiCfg.AuthType == "" {
		return nil
	}

	var authConfig map[string]string
	if err := json.Unmarshal(apiCfg.AuthConfig, &authConfig); err != nil {
		return fmt.Errorf("invalid auth_config: %w", err)
	}

	switch apiCfg.AuthType {
	case "basic":
		req.SetBasicAuth(authConfig["username"], authConfig["password"])
	case "bearer":
		req.Header.Set("Authorization", "Bearer "+authConfig["token"])
	case "api_key":
		headerName := authConfig["header"]
		if headerName == "" {
			headerName = "X-API-Key"
		}
		req.Header.Set(headerName, authConfig["key"])
	case "oauth2_cc":
		token, err := ec.fetchOAuth2Token(req.Context(), authConfig)
		if err != nil {
			return fmt.Errorf("oauth2 token exchange failed: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
	default:
		return fmt.Errorf("unsupported auth type: %s", apiCfg.AuthType)
	}

	return nil
}

func (ec *EvidenceCollector) fetchOAuth2Token(ctx context.Context, authConfig map[string]string) (string, error) {
	tokenURL := authConfig["token_url"]
	clientID := authConfig["client_id"]
	clientSecret := authConfig["client_secret"]
	scope := authConfig["scope"]

	body := fmt.Sprintf("grant_type=client_credentials&client_id=%s&client_secret=%s",
		clientID, clientSecret)
	if scope != "" {
		body += "&scope=" + scope
	}

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token endpoint returned status %d", resp.StatusCode)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", err
	}

	return tokenResp.AccessToken, nil
}

// ============================================================
// FILE COLLECTOR
// ============================================================

// CollectFile reads evidence from a file at the configured path.
func (ec *EvidenceCollector) CollectFile(_ context.Context, cfg CollectionConfig) (json.RawMessage, error) {
	var fileCfg FileConfigPayload
	if err := json.Unmarshal(cfg.FileConfig, &fileCfg); err != nil {
		return nil, fmt.Errorf("invalid file_config: %w", err)
	}

	if fileCfg.Path == "" {
		return nil, fmt.Errorf("file_config.path is required")
	}
	if fileCfg.MaxSizeBytes <= 0 {
		fileCfg.MaxSizeBytes = 50 * 1024 * 1024 // 50MB default
	}

	info, err := os.Stat(fileCfg.Path)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}

	if info.Size() > fileCfg.MaxSizeBytes {
		return nil, fmt.Errorf("file size %d exceeds limit %d", info.Size(), fileCfg.MaxSizeBytes)
	}

	data, err := os.ReadFile(fileCfg.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	result := map[string]interface{}{
		"source":       "file_watch",
		"path":         fileCfg.Path,
		"file_size":    info.Size(),
		"modified_at":  info.ModTime().UTC().Format(time.RFC3339),
		"collected_at": time.Now().UTC().Format(time.RFC3339),
	}

	// Try to parse as JSON
	var parsed interface{}
	if err := json.Unmarshal(data, &parsed); err == nil {
		result["data"] = parsed
	} else {
		result["data"] = string(data)
	}

	return json.Marshal(result)
}

// ============================================================
// SCRIPT COLLECTOR
// ============================================================

// CollectScript executes a script/command and captures its output as evidence.
func (ec *EvidenceCollector) CollectScript(ctx context.Context, cfg CollectionConfig) (json.RawMessage, error) {
	var scriptCfg ScriptConfigPayload
	if err := json.Unmarshal(cfg.ScriptConfig, &scriptCfg); err != nil {
		return nil, fmt.Errorf("invalid script_config: %w", err)
	}

	if scriptCfg.Command == "" {
		return nil, fmt.Errorf("script_config.command is required")
	}
	if scriptCfg.TimeoutSec <= 0 {
		scriptCfg.TimeoutSec = 60
	}

	execCtx, cancel := context.WithTimeout(ctx, time.Duration(scriptCfg.TimeoutSec)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(execCtx, scriptCfg.Command, scriptCfg.Args...)
	if scriptCfg.WorkDir != "" {
		cmd.Dir = scriptCfg.WorkDir
	}

	output, err := cmd.CombinedOutput()

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else if execCtx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("script execution timed out after %d seconds", scriptCfg.TimeoutSec)
		} else {
			return nil, fmt.Errorf("script execution failed: %w", err)
		}
	}

	result := map[string]interface{}{
		"source":       "script_execution",
		"command":      scriptCfg.Command,
		"exit_code":    exitCode,
		"collected_at": time.Now().UTC().Format(time.RFC3339),
	}

	// Try to parse output as JSON
	var parsed interface{}
	if err := json.Unmarshal(output, &parsed); err == nil {
		result["data"] = parsed
	} else {
		result["data"] = string(output)
	}

	return json.Marshal(result)
}

// ============================================================
// WEBHOOK RECEIVER
// ============================================================

// ValidateWebhookSignature validates the HMAC-SHA256 signature of an incoming webhook.
func (ec *EvidenceCollector) ValidateWebhookSignature(payload []byte, signature, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))

	// Support "sha256=..." prefix format
	signature = strings.TrimPrefix(signature, "sha256=")

	return hmac.Equal([]byte(expected), []byte(signature))
}

// ProcessWebhookPayload processes an incoming webhook payload for a given config.
func (ec *EvidenceCollector) ProcessWebhookPayload(ctx context.Context, configID, orgID uuid.UUID, payload json.RawMessage) error {
	cfg, err := ec.GetConfig(ctx, orgID, configID)
	if err != nil {
		return fmt.Errorf("config not found: %w", err)
	}

	if cfg.CollectionMethod != "webhook_receive" {
		return fmt.Errorf("config %s is not a webhook receiver", configID)
	}

	runID := uuid.New()
	startedAt := time.Now()

	_, err = ec.pool.Exec(ctx, `
		INSERT INTO evidence_collection_runs (id, organization_id, config_id, control_implementation_id, status, started_at)
		VALUES ($1, $2, $3, $4, 'running', $5)`,
		runID, orgID, configID, cfg.ControlImplementationID, startedAt)
	if err != nil {
		return fmt.Errorf("failed to create run: %w", err)
	}

	// Wrap the payload
	collectedData, _ := json.Marshal(map[string]interface{}{
		"source":       "webhook_receive",
		"collected_at": time.Now().UTC().Format(time.RFC3339),
		"data":         json.RawMessage(payload),
	})

	// Validate
	validationResults, allPassed := ec.ValidateEvidence(cfg.AcceptanceCriteria, collectedData)
	validationJSON, _ := json.Marshal(validationResults)

	completedAt := time.Now()
	durationMs := int(completedAt.Sub(startedAt).Milliseconds())

	status := "success"
	if !allPassed {
		status = "validation_failed"
	}

	var evidenceID *uuid.UUID
	if allPassed && cfg.ControlImplementationID != nil {
		eid, err := ec.createEvidenceRecord(ctx, orgID, *cfg.ControlImplementationID, cfg.Name, collectedData)
		if err != nil {
			log.Error().Err(err).Msg("Failed to create evidence from webhook")
		} else {
			evidenceID = &eid
		}
	}

	_, err = ec.pool.Exec(ctx, `
		UPDATE evidence_collection_runs
		SET status = $1, completed_at = $2, duration_ms = $3,
			collected_data = $4, validation_results = $5,
			all_criteria_passed = $6, evidence_id = $7
		WHERE id = $8`,
		status, completedAt, durationMs, collectedData, validationJSON,
		allPassed, evidenceID, runID)

	return err
}

// ============================================================
// EVIDENCE VALIDATOR
// ============================================================

// ValidateEvidence evaluates collected data against a set of acceptance criteria.
// Returns individual results and whether all criteria passed.
func (ec *EvidenceCollector) ValidateEvidence(criteriaJSON json.RawMessage, collectedData json.RawMessage) ([]CriterionValidationResult, bool) {
	var criteria []AcceptanceCriterion
	if err := json.Unmarshal(criteriaJSON, &criteria); err != nil || len(criteria) == 0 {
		// No criteria means auto-pass
		return nil, true
	}

	// Parse collected data into a flat map for field access
	var dataMap map[string]interface{}
	if err := json.Unmarshal(collectedData, &dataMap); err != nil {
		return []CriterionValidationResult{{
			Field:   "_root",
			Message: "collected data is not valid JSON object",
			Passed:  false,
		}}, false
	}

	results := make([]CriterionValidationResult, 0, len(criteria))
	allPassed := true

	for _, c := range criteria {
		result := evaluateCriterion(c, dataMap)
		results = append(results, result)
		if !result.Passed {
			allPassed = false
		}
	}

	return results, allPassed
}

// evaluateCriterion evaluates a single acceptance criterion against the data.
func evaluateCriterion(c AcceptanceCriterion, data map[string]interface{}) CriterionValidationResult {
	result := CriterionValidationResult{
		Field:    c.Field,
		Operator: c.Operator,
		Expected: c.Value,
		Message:  c.Message,
	}

	// Navigate the field path (supports dot notation for nested fields)
	actual := resolveField(c.Field, data)
	actualStr := fmt.Sprintf("%v", actual)
	result.Actual = actualStr

	if actual == nil {
		result.Passed = false
		if result.Message == "" {
			result.Message = fmt.Sprintf("field '%s' not found in collected data", c.Field)
		}
		return result
	}

	switch c.Operator {
	case "equals":
		result.Passed = actualStr == c.Value
	case "not_equals":
		result.Passed = actualStr != c.Value
	case "greater_than":
		result.Passed = compareNumeric(actual, c.Value, ">")
	case "less_than":
		result.Passed = compareNumeric(actual, c.Value, "<")
	case "contains":
		result.Passed = strings.Contains(actualStr, c.Value)
	case "matches_regex":
		re, err := regexp.Compile(c.Value)
		if err != nil {
			result.Passed = false
			result.Message = fmt.Sprintf("invalid regex pattern: %s", c.Value)
		} else {
			result.Passed = re.MatchString(actualStr)
		}
	default:
		result.Passed = false
		result.Message = fmt.Sprintf("unsupported operator: %s", c.Operator)
	}

	if !result.Passed && result.Message == "" {
		result.Message = fmt.Sprintf("'%s' %s '%s' failed (actual: '%s')", c.Field, c.Operator, c.Value, actualStr)
	}

	return result
}

// resolveField traverses a dot-notated field path in the data map.
func resolveField(path string, data map[string]interface{}) interface{} {
	parts := strings.Split(path, ".")
	var current interface{} = data

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			current = v[part]
		default:
			return nil
		}
	}

	return current
}

// compareNumeric compares two values numerically.
func compareNumeric(actual interface{}, expected string, op string) bool {
	var actualFloat float64
	switch v := actual.(type) {
	case float64:
		actualFloat = v
	case int:
		actualFloat = float64(v)
	case int64:
		actualFloat = float64(v)
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return false
		}
		actualFloat = f
	case json.Number:
		f, err := v.Float64()
		if err != nil {
			return false
		}
		actualFloat = f
	default:
		s := fmt.Sprintf("%v", v)
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return false
		}
		actualFloat = f
	}

	expectedFloat, err := strconv.ParseFloat(expected, 64)
	if err != nil {
		return false
	}

	switch op {
	case ">":
		return actualFloat > expectedFloat
	case "<":
		return actualFloat < expectedFloat
	default:
		return false
	}
}

// ============================================================
// HELPER METHODS
// ============================================================

func (ec *EvidenceCollector) createEvidenceRecord(ctx context.Context, orgID, controlImplID uuid.UUID, title string, data json.RawMessage) (uuid.UUID, error) {
	var evidenceID uuid.UUID
	err := ec.pool.QueryRow(ctx, `
		INSERT INTO control_evidence (
			organization_id, control_implementation_id, title, description,
			evidence_type, collection_method, collected_by, is_current,
			review_status, metadata
		) VALUES ($1, $2, $3, 'Automated evidence collection', 'automated',
			'automated_collection', '00000000-0000-0000-0000-000000000000', true, 'pending', $4)
		RETURNING id`,
		orgID, controlImplID, title, data,
	).Scan(&evidenceID)

	return evidenceID, err
}

func (ec *EvidenceCollector) updateControlStatus(ctx context.Context, orgID, controlImplID uuid.UUID) {
	// Check if we have recent passing evidence
	var hasPassingEvidence bool
	err := ec.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM control_evidence
			WHERE organization_id = $1
				AND control_implementation_id = $2
				AND is_current = true
				AND review_status IN ('pending', 'accepted')
				AND collected_at > NOW() - INTERVAL '90 days'
				AND deleted_at IS NULL
		)`, orgID, controlImplID).Scan(&hasPassingEvidence)

	if err != nil {
		log.Error().Err(err).Msg("Failed to check evidence status")
		return
	}

	if hasPassingEvidence {
		_, err = ec.pool.Exec(ctx, `
			UPDATE control_implementations
			SET status = 'effective', updated_at = NOW()
			WHERE id = $1 AND organization_id = $2
				AND status NOT IN ('effective', 'not_applicable')
				AND deleted_at IS NULL`,
			controlImplID, orgID)
		if err != nil {
			log.Error().Err(err).Msg("Failed to update control status")
		}
	}
}

// GetConfig retrieves a single evidence collection config by ID.
func (ec *EvidenceCollector) GetConfig(ctx context.Context, orgID, configID uuid.UUID) (*CollectionConfig, error) {
	var c CollectionConfig
	err := ec.pool.QueryRow(ctx, `
		SELECT id, organization_id, control_implementation_id, name, collection_method,
			schedule_cron, schedule_description,
			api_config, file_config, script_config, webhook_config,
			acceptance_criteria, failure_threshold, auto_update_control_status,
			is_active, last_collection_at, last_collection_status,
			next_collection_at, consecutive_failures, created_at, updated_at
		FROM evidence_collection_configs
		WHERE id = $1 AND organization_id = $2`,
		configID, orgID,
	).Scan(
		&c.ID, &c.OrganizationID, &c.ControlImplementationID,
		&c.Name, &c.CollectionMethod,
		&c.ScheduleCron, &c.ScheduleDescription,
		&c.APIConfig, &c.FileConfig, &c.ScriptConfig, &c.WebhookConfig,
		&c.AcceptanceCriteria, &c.FailureThreshold, &c.AutoUpdateControlStatus,
		&c.IsActive, &c.LastCollectionAt, &c.LastCollectionStatus,
		&c.NextCollectionAt, &c.ConsecutiveFailures, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("config not found: %w", err)
	}
	return &c, nil
}

// ListConfigs returns paginated evidence collection configs for an organization.
func (ec *EvidenceCollector) ListConfigs(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]CollectionConfig, int64, error) {
	var total int64
	err := ec.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM evidence_collection_configs WHERE organization_id = $1`, orgID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := ec.pool.Query(ctx, `
		SELECT id, organization_id, control_implementation_id, name, collection_method,
			schedule_cron, schedule_description,
			api_config, file_config, script_config, webhook_config,
			acceptance_criteria, failure_threshold, auto_update_control_status,
			is_active, last_collection_at, last_collection_status,
			next_collection_at, consecutive_failures, created_at, updated_at
		FROM evidence_collection_configs
		WHERE organization_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`,
		orgID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var configs []CollectionConfig
	for rows.Next() {
		var c CollectionConfig
		if err := rows.Scan(
			&c.ID, &c.OrganizationID, &c.ControlImplementationID,
			&c.Name, &c.CollectionMethod,
			&c.ScheduleCron, &c.ScheduleDescription,
			&c.APIConfig, &c.FileConfig, &c.ScriptConfig, &c.WebhookConfig,
			&c.AcceptanceCriteria, &c.FailureThreshold, &c.AutoUpdateControlStatus,
			&c.IsActive, &c.LastCollectionAt, &c.LastCollectionStatus,
			&c.NextCollectionAt, &c.ConsecutiveFailures, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		configs = append(configs, c)
	}

	return configs, total, nil
}

// ListRuns returns paginated collection runs for a config.
func (ec *EvidenceCollector) ListRuns(ctx context.Context, orgID, configID uuid.UUID, limit, offset int) ([]CollectionRun, int64, error) {
	var total int64
	err := ec.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM evidence_collection_runs
		WHERE organization_id = $1 AND config_id = $2`, orgID, configID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := ec.pool.Query(ctx, `
		SELECT id, organization_id, config_id, control_implementation_id,
			status, started_at, completed_at, duration_ms,
			collected_data, validation_results, all_criteria_passed,
			evidence_id, error_message, metadata, created_at
		FROM evidence_collection_runs
		WHERE organization_id = $1 AND config_id = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`,
		orgID, configID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var runs []CollectionRun
	for rows.Next() {
		var r CollectionRun
		if err := rows.Scan(
			&r.ID, &r.OrganizationID, &r.ConfigID, &r.ControlImplementationID,
			&r.Status, &r.StartedAt, &r.CompletedAt, &r.DurationMs,
			&r.CollectedData, &r.CriterionValidationResults, &r.AllCriteriaPassed,
			&r.EvidenceID, &r.ErrorMessage, &r.Metadata, &r.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		runs = append(runs, r)
	}

	return runs, total, nil
}

// CreateConfig creates a new evidence collection configuration.
func (ec *EvidenceCollector) CreateConfig(ctx context.Context, cfg *CollectionConfig) error {
	return ec.pool.QueryRow(ctx, `
		INSERT INTO evidence_collection_configs (
			organization_id, control_implementation_id, name, collection_method,
			schedule_cron, schedule_description,
			api_config, file_config, script_config, webhook_config,
			acceptance_criteria, failure_threshold, auto_update_control_status,
			is_active
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
		RETURNING id, created_at, updated_at`,
		cfg.OrganizationID, cfg.ControlImplementationID, cfg.Name, cfg.CollectionMethod,
		cfg.ScheduleCron, cfg.ScheduleDescription,
		cfg.APIConfig, cfg.FileConfig, cfg.ScriptConfig, cfg.WebhookConfig,
		cfg.AcceptanceCriteria, cfg.FailureThreshold, cfg.AutoUpdateControlStatus,
		cfg.IsActive,
	).Scan(&cfg.ID, &cfg.CreatedAt, &cfg.UpdatedAt)
}

// Pool returns the database pool for direct queries by the handler layer.
func (ec *EvidenceCollector) Pool() *pgxpool.Pool {
	return ec.pool
}

// UpdateConfig updates an existing evidence collection configuration.
func (ec *EvidenceCollector) UpdateConfig(ctx context.Context, cfg *CollectionConfig) error {
	_, err := ec.pool.Exec(ctx, `
		UPDATE evidence_collection_configs SET
			name = $1, collection_method = $2,
			schedule_cron = $3, schedule_description = $4,
			api_config = $5, file_config = $6, script_config = $7, webhook_config = $8,
			acceptance_criteria = $9, failure_threshold = $10,
			auto_update_control_status = $11, is_active = $12,
			control_implementation_id = $13
		WHERE id = $14 AND organization_id = $15`,
		cfg.Name, cfg.CollectionMethod,
		cfg.ScheduleCron, cfg.ScheduleDescription,
		cfg.APIConfig, cfg.FileConfig, cfg.ScriptConfig, cfg.WebhookConfig,
		cfg.AcceptanceCriteria, cfg.FailureThreshold,
		cfg.AutoUpdateControlStatus, cfg.IsActive,
		cfg.ControlImplementationID,
		cfg.ID, cfg.OrganizationID,
	)
	return err
}
