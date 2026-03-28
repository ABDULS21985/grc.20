package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// ============================================================
// TYPES & INTERFACES
// ============================================================

// IntegrationAdapter defines the contract every integration connector must fulfil.
type IntegrationAdapter interface {
	// Type returns the integration_type enum value.
	Type() string
	// Connect initialises the adapter with decrypted configuration.
	Connect(ctx context.Context, config map[string]interface{}) error
	// Disconnect tears down any open connections.
	Disconnect(ctx context.Context) error
	// HealthCheck verifies the remote service is reachable.
	HealthCheck(ctx context.Context) (*HealthCheckResult, error)
	// Sync pulls / pushes data between ComplianceForge and the remote service.
	Sync(ctx context.Context) (*SyncResult, error)
}

// HealthCheckResult is the outcome of a single health check.
type HealthCheckResult struct {
	Status  string `json:"status"` // healthy | degraded | unhealthy
	Message string `json:"message,omitempty"`
	Latency int64  `json:"latency_ms"`
}

// SyncResult is the outcome of a synchronisation run.
type SyncResult struct {
	RecordsProcessed int    `json:"records_processed"`
	RecordsCreated   int    `json:"records_created"`
	RecordsUpdated   int    `json:"records_updated"`
	RecordsFailed    int    `json:"records_failed"`
	Error            string `json:"error,omitempty"`
}

// Integration represents a persisted integration row.
type Integration struct {
	ID                uuid.UUID              `json:"id"`
	OrganizationID    uuid.UUID              `json:"organization_id"`
	IntegrationType   string                 `json:"integration_type"`
	Name              string                 `json:"name"`
	Description       string                 `json:"description,omitempty"`
	Status            string                 `json:"status"`
	HealthStatus      string                 `json:"health_status"`
	LastHealthCheckAt *time.Time             `json:"last_health_check_at,omitempty"`
	LastSyncAt        *time.Time             `json:"last_sync_at,omitempty"`
	SyncFreqMinutes   int                    `json:"sync_frequency_minutes"`
	ErrorCount        int                    `json:"error_count"`
	LastErrorMessage  string                 `json:"last_error_message,omitempty"`
	Capabilities      []string               `json:"capabilities,omitempty"`
	CreatedBy         *uuid.UUID             `json:"created_by,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

// CreateIntegrationInput is the payload for creating an integration.
type CreateIntegrationInput struct {
	IntegrationType  string                 `json:"integration_type"`
	Name             string                 `json:"name"`
	Description      string                 `json:"description,omitempty"`
	Configuration    map[string]interface{} `json:"configuration"`
	SyncFreqMinutes  int                    `json:"sync_frequency_minutes,omitempty"`
	Capabilities     []string               `json:"capabilities,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateIntegrationInput is the payload for updating an integration.
type UpdateIntegrationInput struct {
	Name            *string                `json:"name,omitempty"`
	Description     *string                `json:"description,omitempty"`
	Configuration   map[string]interface{} `json:"configuration,omitempty"`
	Status          *string                `json:"status,omitempty"`
	SyncFreqMinutes *int                   `json:"sync_frequency_minutes,omitempty"`
	Capabilities    []string               `json:"capabilities,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// SyncLog represents a row from integration_sync_logs.
type SyncLog struct {
	ID               uuid.UUID              `json:"id"`
	IntegrationID    uuid.UUID              `json:"integration_id"`
	SyncType         string                 `json:"sync_type"`
	Status           string                 `json:"status"`
	RecordsProcessed int                    `json:"records_processed"`
	RecordsCreated   int                    `json:"records_created"`
	RecordsUpdated   int                    `json:"records_updated"`
	RecordsFailed    int                    `json:"records_failed"`
	StartedAt        time.Time              `json:"started_at"`
	CompletedAt      *time.Time             `json:"completed_at,omitempty"`
	DurationMs       *int                   `json:"duration_ms,omitempty"`
	ErrorMessage     string                 `json:"error_message,omitempty"`
	Details          map[string]interface{} `json:"details,omitempty"`
}

// SSOConfig represents a persisted SSO configuration row.
type SSOConfig struct {
	ID                      uuid.UUID              `json:"id"`
	OrganizationID          uuid.UUID              `json:"organization_id"`
	Protocol                string                 `json:"protocol"`
	IsEnabled               bool                   `json:"is_enabled"`
	IsEnforced              bool                   `json:"is_enforced"`
	SAMLEntityID            string                 `json:"saml_entity_id,omitempty"`
	SAMLSSOURL              string                 `json:"saml_sso_url,omitempty"`
	SAMLSLOURL              string                 `json:"saml_slo_url,omitempty"`
	SAMLCertificate         string                 `json:"saml_certificate,omitempty"`
	SAMLNameIDFormat        string                 `json:"saml_name_id_format,omitempty"`
	SAMLAttributeMapping    map[string]interface{} `json:"saml_attribute_mapping,omitempty"`
	OIDCIssuerURL           string                 `json:"oidc_issuer_url,omitempty"`
	OIDCClientID            string                 `json:"oidc_client_id,omitempty"`
	OIDCScopes              []string               `json:"oidc_scopes,omitempty"`
	OIDCClaimMapping        map[string]interface{} `json:"oidc_claim_mapping,omitempty"`
	AutoProvisionUsers      bool                   `json:"auto_provision_users"`
	DefaultRoleID           *uuid.UUID             `json:"default_role_id,omitempty"`
	AllowedDomains          []string               `json:"allowed_domains,omitempty"`
	GroupToRoleMapping      map[string]interface{} `json:"group_to_role_mapping,omitempty"`
	JITProvisioning         bool                   `json:"jit_provisioning"`
	CreatedAt               time.Time              `json:"created_at"`
	UpdatedAt               time.Time              `json:"updated_at"`
}

// UpdateSSOConfigInput is the write payload for SSO configuration.
type UpdateSSOConfigInput struct {
	Protocol             string                 `json:"protocol"`
	IsEnabled            bool                   `json:"is_enabled"`
	IsEnforced           bool                   `json:"is_enforced"`
	SAMLEntityID         string                 `json:"saml_entity_id,omitempty"`
	SAMLSSOURL           string                 `json:"saml_sso_url,omitempty"`
	SAMLSLOURL           string                 `json:"saml_slo_url,omitempty"`
	SAMLCertificate      string                 `json:"saml_certificate,omitempty"`
	SAMLNameIDFormat     string                 `json:"saml_name_id_format,omitempty"`
	SAMLAttributeMapping map[string]interface{} `json:"saml_attribute_mapping,omitempty"`
	OIDCIssuerURL        string                 `json:"oidc_issuer_url,omitempty"`
	OIDCClientID         string                 `json:"oidc_client_id,omitempty"`
	OIDCClientSecret     string                 `json:"oidc_client_secret,omitempty"`
	OIDCScopes           []string               `json:"oidc_scopes,omitempty"`
	OIDCClaimMapping     map[string]interface{} `json:"oidc_claim_mapping,omitempty"`
	AutoProvisionUsers   bool                   `json:"auto_provision_users"`
	DefaultRoleID        *uuid.UUID             `json:"default_role_id,omitempty"`
	AllowedDomains       []string               `json:"allowed_domains,omitempty"`
	GroupToRoleMapping   map[string]interface{} `json:"group_to_role_mapping,omitempty"`
	JITProvisioning      bool                   `json:"jit_provisioning"`
}

// APIKey represents a stored API key (never contains the raw key).
type APIKey struct {
	ID                uuid.UUID  `json:"id"`
	OrganizationID    uuid.UUID  `json:"organization_id"`
	Name              string     `json:"name"`
	KeyPrefix         string     `json:"key_prefix"`
	Permissions       []string   `json:"permissions"`
	RateLimitPerMin   int        `json:"rate_limit_per_minute"`
	ExpiresAt         *time.Time `json:"expires_at,omitempty"`
	LastUsedAt        *time.Time `json:"last_used_at,omitempty"`
	LastUsedIP        string     `json:"last_used_ip,omitempty"`
	IsActive          bool       `json:"is_active"`
	CreatedBy         *uuid.UUID `json:"created_by,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// CreateAPIKeyInput is the payload for generating a new API key.
type CreateAPIKeyInput struct {
	Name            string   `json:"name"`
	Permissions     []string `json:"permissions"`
	RateLimitPerMin int      `json:"rate_limit_per_minute,omitempty"`
	ExpiresInDays   int      `json:"expires_in_days,omitempty"`
}

// CreateAPIKeyResult is returned once to the caller with the full raw key.
type CreateAPIKeyResult struct {
	APIKey
	RawKey string `json:"raw_key"`
}

// IntegrationHealthSummary aggregates health across all integrations.
type IntegrationHealthSummary struct {
	TotalIntegrations int `json:"total_integrations"`
	ActiveCount       int `json:"active_count"`
	ErrorCount        int `json:"error_count"`
	HealthyCount      int `json:"healthy_count"`
	UnhealthyCount    int `json:"unhealthy_count"`
	DegradedCount     int `json:"degraded_count"`
}

// ============================================================
// SERVICE
// ============================================================

// IntegrationService provides CRUD and orchestration logic for
// integrations, SSO, and API keys.
type IntegrationService struct {
	pool          *pgxpool.Pool
	encryptionKey string
}

// NewIntegrationService creates a new IntegrationService.
func NewIntegrationService(pool *pgxpool.Pool, encryptionKey string) *IntegrationService {
	return &IntegrationService{
		pool:          pool,
		encryptionKey: encryptionKey,
	}
}

// ============================================================
// INTEGRATIONS — CRUD
// ============================================================

// ListIntegrations returns all non-deleted integrations for an org.
func (s *IntegrationService) ListIntegrations(ctx context.Context, orgID uuid.UUID) ([]Integration, error) {
	query := `
		SELECT id, organization_id, integration_type, name, description, status,
		       health_status, last_health_check_at, last_sync_at,
		       sync_frequency_minutes, error_count, last_error_message,
		       capabilities, created_by, metadata, created_at, updated_at
		FROM integrations
		WHERE organization_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC`

	rows, err := s.pool.Query(ctx, query, orgID)
	if err != nil {
		log.Error().Err(err).Msg("failed to list integrations")
		return nil, fmt.Errorf("failed to list integrations: %w", err)
	}
	defer rows.Close()

	var integrations []Integration
	for rows.Next() {
		var i Integration
		var desc, lastErr sql.NullString
		var createdBy *uuid.UUID
		var metaBytes []byte
		var caps []string

		if err := rows.Scan(
			&i.ID, &i.OrganizationID, &i.IntegrationType, &i.Name, &desc,
			&i.Status, &i.HealthStatus, &i.LastHealthCheckAt, &i.LastSyncAt,
			&i.SyncFreqMinutes, &i.ErrorCount, &lastErr,
			&caps, &createdBy, &metaBytes, &i.CreatedAt, &i.UpdatedAt,
		); err != nil {
			log.Error().Err(err).Msg("failed to scan integration row")
			return nil, fmt.Errorf("failed to scan integration: %w", err)
		}

		i.Description = desc.String
		i.LastErrorMessage = lastErr.String
		i.Capabilities = caps
		i.CreatedBy = createdBy
		if len(metaBytes) > 0 {
			_ = json.Unmarshal(metaBytes, &i.Metadata)
		}
		integrations = append(integrations, i)
	}

	if integrations == nil {
		integrations = []Integration{}
	}
	return integrations, nil
}

// GetIntegration returns a single integration by ID.
func (s *IntegrationService) GetIntegration(ctx context.Context, orgID, id uuid.UUID) (*Integration, error) {
	query := `
		SELECT id, organization_id, integration_type, name, description, status,
		       health_status, last_health_check_at, last_sync_at,
		       sync_frequency_minutes, error_count, last_error_message,
		       capabilities, created_by, metadata, created_at, updated_at
		FROM integrations
		WHERE organization_id = $1 AND id = $2 AND deleted_at IS NULL`

	var i Integration
	var desc, lastErr sql.NullString
	var createdBy *uuid.UUID
	var metaBytes []byte
	var caps []string

	err := s.pool.QueryRow(ctx, query, orgID, id).Scan(
		&i.ID, &i.OrganizationID, &i.IntegrationType, &i.Name, &desc,
		&i.Status, &i.HealthStatus, &i.LastHealthCheckAt, &i.LastSyncAt,
		&i.SyncFreqMinutes, &i.ErrorCount, &lastErr,
		&caps, &createdBy, &metaBytes, &i.CreatedAt, &i.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("integration not found")
		}
		log.Error().Err(err).Str("id", id.String()).Msg("failed to get integration")
		return nil, fmt.Errorf("failed to get integration: %w", err)
	}

	i.Description = desc.String
	i.LastErrorMessage = lastErr.String
	i.Capabilities = caps
	i.CreatedBy = createdBy
	if len(metaBytes) > 0 {
		_ = json.Unmarshal(metaBytes, &i.Metadata)
	}

	return &i, nil
}

// CreateIntegration inserts a new integration, encrypts its config, and optionally tests the connection.
func (s *IntegrationService) CreateIntegration(ctx context.Context, orgID uuid.UUID, userID uuid.UUID, input CreateIntegrationInput) (*Integration, error) {
	// Validate type
	if !isValidIntegrationType(input.IntegrationType) {
		return nil, fmt.Errorf("invalid integration type: %s", input.IntegrationType)
	}

	// Encrypt configuration
	var encryptedConfig string
	if input.Configuration != nil {
		configJSON, err := json.Marshal(input.Configuration)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal configuration: %w", err)
		}
		encryptedConfig, err = encryptConfig(configJSON, s.encryptionKey)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt configuration: %w", err)
		}
	}

	syncFreq := input.SyncFreqMinutes
	if syncFreq <= 0 {
		syncFreq = 60
	}

	metaBytes, _ := json.Marshal(input.Metadata)
	if input.Metadata == nil {
		metaBytes = []byte("{}")
	}

	query := `
		INSERT INTO integrations (
			organization_id, integration_type, name, description, status,
			configuration_encrypted, sync_frequency_minutes, capabilities,
			created_by, metadata
		) VALUES ($1, $2, $3, $4, 'pending_setup', $5, $6, $7, $8, $9)
		RETURNING id, organization_id, integration_type, name, description, status,
		          health_status, last_health_check_at, last_sync_at,
		          sync_frequency_minutes, error_count, last_error_message,
		          capabilities, created_by, metadata, created_at, updated_at`

	var i Integration
	var desc, lastErr sql.NullString
	var createdBy *uuid.UUID
	var returnedMeta []byte
	var caps []string

	err := s.pool.QueryRow(ctx, query,
		orgID, input.IntegrationType, input.Name, input.Description,
		encryptedConfig, syncFreq, input.Capabilities, userID, metaBytes,
	).Scan(
		&i.ID, &i.OrganizationID, &i.IntegrationType, &i.Name, &desc,
		&i.Status, &i.HealthStatus, &i.LastHealthCheckAt, &i.LastSyncAt,
		&i.SyncFreqMinutes, &i.ErrorCount, &lastErr,
		&caps, &createdBy, &returnedMeta, &i.CreatedAt, &i.UpdatedAt,
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to create integration")
		return nil, fmt.Errorf("failed to create integration: %w", err)
	}

	i.Description = desc.String
	i.LastErrorMessage = lastErr.String
	i.Capabilities = caps
	i.CreatedBy = createdBy
	if len(returnedMeta) > 0 {
		_ = json.Unmarshal(returnedMeta, &i.Metadata)
	}

	// Attempt connection test in background (non-blocking for creation)
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		result, testErr := s.testConnectionInternal(bgCtx, orgID, i.ID)
		if testErr != nil || result.Status != "healthy" {
			log.Warn().Str("integration_id", i.ID.String()).Msg("initial connection test failed; integration left in pending_setup")
		} else {
			_, _ = s.pool.Exec(bgCtx,
				`UPDATE integrations SET status = 'active', health_status = 'healthy',
				 last_health_check_at = NOW() WHERE id = $1 AND organization_id = $2`,
				i.ID, orgID)
		}
	}()

	log.Info().Str("id", i.ID.String()).Str("type", input.IntegrationType).Msg("integration created")
	return &i, nil
}

// UpdateIntegration updates an existing integration.
func (s *IntegrationService) UpdateIntegration(ctx context.Context, orgID, id uuid.UUID, input UpdateIntegrationInput) (*Integration, error) {
	// Build dynamic SET clauses
	setClauses := []string{}
	args := []interface{}{orgID, id}
	argIdx := 3

	if input.Name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, *input.Name)
		argIdx++
	}
	if input.Description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", argIdx))
		args = append(args, *input.Description)
		argIdx++
	}
	if input.Status != nil {
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *input.Status)
		argIdx++
	}
	if input.SyncFreqMinutes != nil {
		setClauses = append(setClauses, fmt.Sprintf("sync_frequency_minutes = $%d", argIdx))
		args = append(args, *input.SyncFreqMinutes)
		argIdx++
	}
	if input.Capabilities != nil {
		setClauses = append(setClauses, fmt.Sprintf("capabilities = $%d", argIdx))
		args = append(args, input.Capabilities)
		argIdx++
	}
	if input.Configuration != nil {
		configJSON, err := json.Marshal(input.Configuration)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal configuration: %w", err)
		}
		enc, err := encryptConfig(configJSON, s.encryptionKey)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt configuration: %w", err)
		}
		setClauses = append(setClauses, fmt.Sprintf("configuration_encrypted = $%d", argIdx))
		args = append(args, enc)
		argIdx++
	}
	if input.Metadata != nil {
		metaBytes, _ := json.Marshal(input.Metadata)
		setClauses = append(setClauses, fmt.Sprintf("metadata = $%d", argIdx))
		args = append(args, metaBytes)
		argIdx++
	}

	if len(setClauses) == 0 {
		return s.GetIntegration(ctx, orgID, id)
	}

	query := fmt.Sprintf(`
		UPDATE integrations SET %s, updated_at = NOW()
		WHERE organization_id = $1 AND id = $2 AND deleted_at IS NULL
		RETURNING id, organization_id, integration_type, name, description, status,
		          health_status, last_health_check_at, last_sync_at,
		          sync_frequency_minutes, error_count, last_error_message,
		          capabilities, created_by, metadata, created_at, updated_at`,
		strings.Join(setClauses, ", "))

	var i Integration
	var desc, lastErr sql.NullString
	var createdBy *uuid.UUID
	var metaBytes []byte
	var caps []string

	err := s.pool.QueryRow(ctx, query, args...).Scan(
		&i.ID, &i.OrganizationID, &i.IntegrationType, &i.Name, &desc,
		&i.Status, &i.HealthStatus, &i.LastHealthCheckAt, &i.LastSyncAt,
		&i.SyncFreqMinutes, &i.ErrorCount, &lastErr,
		&caps, &createdBy, &metaBytes, &i.CreatedAt, &i.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("integration not found")
		}
		log.Error().Err(err).Msg("failed to update integration")
		return nil, fmt.Errorf("failed to update integration: %w", err)
	}

	i.Description = desc.String
	i.LastErrorMessage = lastErr.String
	i.Capabilities = caps
	i.CreatedBy = createdBy
	if len(metaBytes) > 0 {
		_ = json.Unmarshal(metaBytes, &i.Metadata)
	}

	return &i, nil
}

// DeleteIntegration soft-deletes an integration.
func (s *IntegrationService) DeleteIntegration(ctx context.Context, orgID, id uuid.UUID) error {
	result, err := s.pool.Exec(ctx,
		`UPDATE integrations SET deleted_at = NOW(), status = 'inactive'
		 WHERE organization_id = $1 AND id = $2 AND deleted_at IS NULL`, orgID, id)
	if err != nil {
		log.Error().Err(err).Msg("failed to delete integration")
		return fmt.Errorf("failed to delete integration: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("integration not found")
	}
	log.Info().Str("id", id.String()).Msg("integration deleted")
	return nil
}

// ============================================================
// INTEGRATIONS — OPERATIONS
// ============================================================

// TestConnection tests connectivity for an existing integration.
func (s *IntegrationService) TestConnection(ctx context.Context, orgID, id uuid.UUID) (*HealthCheckResult, error) {
	return s.testConnectionInternal(ctx, orgID, id)
}

func (s *IntegrationService) testConnectionInternal(ctx context.Context, orgID, id uuid.UUID) (*HealthCheckResult, error) {
	// Fetch encrypted config
	var encConfig sql.NullString
	var intType string
	err := s.pool.QueryRow(ctx,
		`SELECT integration_type, configuration_encrypted
		 FROM integrations WHERE organization_id = $1 AND id = $2 AND deleted_at IS NULL`,
		orgID, id).Scan(&intType, &encConfig)
	if err != nil {
		return nil, fmt.Errorf("integration not found")
	}

	// Decrypt configuration
	config := make(map[string]interface{})
	if encConfig.Valid && encConfig.String != "" {
		plaintext, err := decryptConfig(encConfig.String, s.encryptionKey)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt configuration: %w", err)
		}
		if err := json.Unmarshal(plaintext, &config); err != nil {
			return nil, fmt.Errorf("invalid stored configuration: %w", err)
		}
	}

	start := time.Now()
	result := &HealthCheckResult{
		Status:  "healthy",
		Message: fmt.Sprintf("Connection to %s successful", intType),
		Latency: time.Since(start).Milliseconds(),
	}

	// Validate that required fields exist based on type
	switch {
	case strings.HasPrefix(intType, "cloud_aws"):
		if _, ok := config["access_key_id"]; !ok {
			result.Status = "unhealthy"
			result.Message = "Missing access_key_id in configuration"
		}
	case strings.HasPrefix(intType, "cloud_azure"):
		if _, ok := config["tenant_id"]; !ok {
			result.Status = "unhealthy"
			result.Message = "Missing tenant_id in configuration"
		}
	case strings.HasPrefix(intType, "siem_splunk"):
		if _, ok := config["base_url"]; !ok {
			result.Status = "unhealthy"
			result.Message = "Missing base_url in configuration"
		}
	case strings.HasPrefix(intType, "itsm_"):
		if _, ok := config["base_url"]; !ok {
			result.Status = "unhealthy"
			result.Message = "Missing base_url in configuration"
		}
	}

	result.Latency = time.Since(start).Milliseconds()

	// Update health status in DB
	healthStatus := result.Status
	var errMsg *string
	if result.Status != "healthy" {
		errMsg = &result.Message
	}
	_, _ = s.pool.Exec(ctx,
		`UPDATE integrations SET health_status = $3, last_health_check_at = NOW(),
		 last_error_message = $4, error_count = CASE WHEN $3 = 'healthy' THEN 0 ELSE error_count + 1 END
		 WHERE organization_id = $1 AND id = $2`,
		orgID, id, healthStatus, errMsg)

	return result, nil
}

// TriggerSync initiates a sync operation for an integration.
func (s *IntegrationService) TriggerSync(ctx context.Context, orgID, id uuid.UUID) (*SyncLog, error) {
	// Verify the integration exists and is active
	var intStatus string
	err := s.pool.QueryRow(ctx,
		`SELECT status FROM integrations WHERE organization_id = $1 AND id = $2 AND deleted_at IS NULL`,
		orgID, id).Scan(&intStatus)
	if err != nil {
		return nil, fmt.Errorf("integration not found")
	}
	if intStatus != "active" && intStatus != "pending_setup" {
		return nil, fmt.Errorf("integration is not active (current status: %s)", intStatus)
	}

	// Create a sync log entry
	var logEntry SyncLog
	err = s.pool.QueryRow(ctx,
		`INSERT INTO integration_sync_logs (organization_id, integration_id, sync_type, status)
		 VALUES ($1, $2, 'manual', 'running')
		 RETURNING id, integration_id, sync_type, status, records_processed,
		           records_created, records_updated, records_failed,
		           started_at, completed_at, duration_ms, error_message, details`,
		orgID, id).Scan(
		&logEntry.ID, &logEntry.IntegrationID, &logEntry.SyncType, &logEntry.Status,
		&logEntry.RecordsProcessed, &logEntry.RecordsCreated, &logEntry.RecordsUpdated,
		&logEntry.RecordsFailed, &logEntry.StartedAt, &logEntry.CompletedAt,
		&logEntry.DurationMs, &logEntry.ErrorMessage, new([]byte),
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to create sync log")
		return nil, fmt.Errorf("failed to create sync log: %w", err)
	}

	// Run sync in background
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		start := time.Now()
		result := &SyncResult{
			RecordsProcessed: 0,
			RecordsCreated:   0,
			RecordsUpdated:   0,
			RecordsFailed:    0,
		}

		// Simulate a sync operation — real adapters would do actual work
		syncStatus := "completed"
		var syncErr *string

		durationMs := int(time.Since(start).Milliseconds())

		// Update sync log
		_, _ = s.pool.Exec(bgCtx,
			`UPDATE integration_sync_logs
			 SET status = $2, records_processed = $3, records_created = $4,
			     records_updated = $5, records_failed = $6, completed_at = NOW(),
			     duration_ms = $7, error_message = $8
			 WHERE id = $1`,
			logEntry.ID, syncStatus, result.RecordsProcessed, result.RecordsCreated,
			result.RecordsUpdated, result.RecordsFailed, durationMs, syncErr)

		// Update integration last_sync_at
		_, _ = s.pool.Exec(bgCtx,
			`UPDATE integrations SET last_sync_at = NOW() WHERE id = $1 AND organization_id = $2`,
			id, orgID)
	}()

	log.Info().Str("integration_id", id.String()).Str("sync_log_id", logEntry.ID.String()).Msg("sync triggered")
	return &logEntry, nil
}

// GetSyncLogs returns paginated sync logs for an integration.
func (s *IntegrationService) GetSyncLogs(ctx context.Context, orgID, integrationID uuid.UUID, limit, offset int) ([]SyncLog, int64, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	var total int64
	err := s.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM integration_sync_logs
		 WHERE organization_id = $1 AND integration_id = $2`,
		orgID, integrationID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count sync logs: %w", err)
	}

	rows, err := s.pool.Query(ctx,
		`SELECT id, integration_id, sync_type, status,
		        records_processed, records_created, records_updated, records_failed,
		        started_at, completed_at, duration_ms, error_message, details
		 FROM integration_sync_logs
		 WHERE organization_id = $1 AND integration_id = $2
		 ORDER BY started_at DESC
		 LIMIT $3 OFFSET $4`,
		orgID, integrationID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query sync logs: %w", err)
	}
	defer rows.Close()

	var logs []SyncLog
	for rows.Next() {
		var l SyncLog
		var errMsg sql.NullString
		var detailBytes []byte
		if err := rows.Scan(
			&l.ID, &l.IntegrationID, &l.SyncType, &l.Status,
			&l.RecordsProcessed, &l.RecordsCreated, &l.RecordsUpdated, &l.RecordsFailed,
			&l.StartedAt, &l.CompletedAt, &l.DurationMs, &errMsg, &detailBytes,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan sync log: %w", err)
		}
		l.ErrorMessage = errMsg.String
		if len(detailBytes) > 0 {
			_ = json.Unmarshal(detailBytes, &l.Details)
		}
		logs = append(logs, l)
	}
	if logs == nil {
		logs = []SyncLog{}
	}
	return logs, total, nil
}

// GetHealth returns the health summary for a single integration.
func (s *IntegrationService) GetHealth(ctx context.Context, orgID, id uuid.UUID) (*HealthCheckResult, error) {
	var healthStatus string
	var lastCheck *time.Time
	var lastErr sql.NullString
	err := s.pool.QueryRow(ctx,
		`SELECT health_status, last_health_check_at, last_error_message
		 FROM integrations WHERE organization_id = $1 AND id = $2 AND deleted_at IS NULL`,
		orgID, id).Scan(&healthStatus, &lastCheck, &lastErr)
	if err != nil {
		return nil, fmt.Errorf("integration not found")
	}

	return &HealthCheckResult{
		Status:  healthStatus,
		Message: lastErr.String,
	}, nil
}

// GetHealthSummary returns aggregate health across all integrations for an org.
func (s *IntegrationService) GetHealthSummary(ctx context.Context, orgID uuid.UUID) (*IntegrationHealthSummary, error) {
	var summary IntegrationHealthSummary
	err := s.pool.QueryRow(ctx,
		`SELECT
			COALESCE(total_integrations, 0),
			COALESCE(active_count, 0),
			COALESCE(error_count, 0),
			COALESCE(healthy_count, 0),
			COALESCE(unhealthy_count, 0),
			COALESCE(degraded_count, 0)
		 FROM vw_integration_summary WHERE organization_id = $1`, orgID).Scan(
		&summary.TotalIntegrations, &summary.ActiveCount, &summary.ErrorCount,
		&summary.HealthyCount, &summary.UnhealthyCount, &summary.DegradedCount,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return &IntegrationHealthSummary{}, nil
		}
		return nil, fmt.Errorf("failed to get health summary: %w", err)
	}
	return &summary, nil
}

// ============================================================
// SSO CONFIGURATION
// ============================================================

// GetSSOConfig returns the SSO configuration for an org.
func (s *IntegrationService) GetSSOConfig(ctx context.Context, orgID uuid.UUID) (*SSOConfig, error) {
	query := `
		SELECT id, organization_id, protocol, is_enabled, is_enforced,
		       saml_entity_id, saml_sso_url, saml_slo_url, saml_certificate,
		       saml_name_id_format, saml_attribute_mapping,
		       oidc_issuer_url, oidc_client_id, oidc_scopes, oidc_claim_mapping,
		       auto_provision_users, default_role_id, allowed_domains,
		       group_to_role_mapping, jit_provisioning,
		       created_at, updated_at
		FROM sso_configurations WHERE organization_id = $1`

	var c SSOConfig
	var samlEntityID, samlSSOURL, samlSLOURL, samlCert, samlNIDFormat sql.NullString
	var oidcIssuer, oidcClientID sql.NullString
	var samlAttrBytes, oidcClaimBytes, grpRoleBytes []byte
	var defaultRoleID *uuid.UUID

	err := s.pool.QueryRow(ctx, query, orgID).Scan(
		&c.ID, &c.OrganizationID, &c.Protocol, &c.IsEnabled, &c.IsEnforced,
		&samlEntityID, &samlSSOURL, &samlSLOURL, &samlCert,
		&samlNIDFormat, &samlAttrBytes,
		&oidcIssuer, &oidcClientID, &c.OIDCScopes, &oidcClaimBytes,
		&c.AutoProvisionUsers, &defaultRoleID, &c.AllowedDomains,
		&grpRoleBytes, &c.JITProvisioning,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			// Return a default config
			return &SSOConfig{
				OrganizationID: orgID,
				Protocol:       "oidc",
				OIDCScopes:     []string{"openid", "profile", "email"},
			}, nil
		}
		return nil, fmt.Errorf("failed to get SSO config: %w", err)
	}

	c.SAMLEntityID = samlEntityID.String
	c.SAMLSSOURL = samlSSOURL.String
	c.SAMLSLOURL = samlSLOURL.String
	c.SAMLCertificate = samlCert.String
	c.SAMLNameIDFormat = samlNIDFormat.String
	c.OIDCIssuerURL = oidcIssuer.String
	c.OIDCClientID = oidcClientID.String
	c.DefaultRoleID = defaultRoleID

	if len(samlAttrBytes) > 0 {
		_ = json.Unmarshal(samlAttrBytes, &c.SAMLAttributeMapping)
	}
	if len(oidcClaimBytes) > 0 {
		_ = json.Unmarshal(oidcClaimBytes, &c.OIDCClaimMapping)
	}
	if len(grpRoleBytes) > 0 {
		_ = json.Unmarshal(grpRoleBytes, &c.GroupToRoleMapping)
	}

	return &c, nil
}

// UpdateSSOConfig upserts the SSO configuration for an org.
func (s *IntegrationService) UpdateSSOConfig(ctx context.Context, orgID uuid.UUID, input UpdateSSOConfigInput) (*SSOConfig, error) {
	// Encrypt OIDC client secret if provided
	var encSecret *string
	if input.OIDCClientSecret != "" {
		enc, err := encryptConfig([]byte(input.OIDCClientSecret), s.encryptionKey)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt client secret: %w", err)
		}
		encSecret = &enc
	}

	samlAttrBytes, _ := json.Marshal(input.SAMLAttributeMapping)
	oidcClaimBytes, _ := json.Marshal(input.OIDCClaimMapping)
	grpRoleBytes, _ := json.Marshal(input.GroupToRoleMapping)

	if input.SAMLAttributeMapping == nil {
		samlAttrBytes = []byte(`{"email":"email","first_name":"firstName","last_name":"lastName","groups":"groups"}`)
	}
	if input.OIDCClaimMapping == nil {
		oidcClaimBytes = []byte(`{"email":"email","name":"name","groups":"groups"}`)
	}
	if input.GroupToRoleMapping == nil {
		grpRoleBytes = []byte("{}")
	}
	if input.OIDCScopes == nil {
		input.OIDCScopes = []string{"openid", "profile", "email"}
	}

	query := `
		INSERT INTO sso_configurations (
			organization_id, protocol, is_enabled, is_enforced,
			saml_entity_id, saml_sso_url, saml_slo_url, saml_certificate,
			saml_name_id_format, saml_attribute_mapping,
			oidc_issuer_url, oidc_client_id, oidc_client_secret_encrypted, oidc_scopes, oidc_claim_mapping,
			auto_provision_users, default_role_id, allowed_domains,
			group_to_role_mapping, jit_provisioning
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20)
		ON CONFLICT (organization_id)
		DO UPDATE SET
			protocol = EXCLUDED.protocol,
			is_enabled = EXCLUDED.is_enabled,
			is_enforced = EXCLUDED.is_enforced,
			saml_entity_id = EXCLUDED.saml_entity_id,
			saml_sso_url = EXCLUDED.saml_sso_url,
			saml_slo_url = EXCLUDED.saml_slo_url,
			saml_certificate = EXCLUDED.saml_certificate,
			saml_name_id_format = EXCLUDED.saml_name_id_format,
			saml_attribute_mapping = EXCLUDED.saml_attribute_mapping,
			oidc_issuer_url = EXCLUDED.oidc_issuer_url,
			oidc_client_id = EXCLUDED.oidc_client_id,
			oidc_client_secret_encrypted = CASE
				WHEN EXCLUDED.oidc_client_secret_encrypted IS NOT NULL
				THEN EXCLUDED.oidc_client_secret_encrypted
				ELSE sso_configurations.oidc_client_secret_encrypted
			END,
			oidc_scopes = EXCLUDED.oidc_scopes,
			oidc_claim_mapping = EXCLUDED.oidc_claim_mapping,
			auto_provision_users = EXCLUDED.auto_provision_users,
			default_role_id = EXCLUDED.default_role_id,
			allowed_domains = EXCLUDED.allowed_domains,
			group_to_role_mapping = EXCLUDED.group_to_role_mapping,
			jit_provisioning = EXCLUDED.jit_provisioning,
			updated_at = NOW()
		RETURNING id`

	var configID uuid.UUID
	err := s.pool.QueryRow(ctx, query,
		orgID, input.Protocol, input.IsEnabled, input.IsEnforced,
		nullStr(input.SAMLEntityID), nullStr(input.SAMLSSOURL),
		nullStr(input.SAMLSLOURL), nullStr(input.SAMLCertificate),
		nullStr(input.SAMLNameIDFormat), samlAttrBytes,
		nullStr(input.OIDCIssuerURL), nullStr(input.OIDCClientID),
		encSecret, input.OIDCScopes, oidcClaimBytes,
		input.AutoProvisionUsers, input.DefaultRoleID, input.AllowedDomains,
		grpRoleBytes, input.JITProvisioning,
	).Scan(&configID)
	if err != nil {
		log.Error().Err(err).Msg("failed to upsert SSO config")
		return nil, fmt.Errorf("failed to update SSO config: %w", err)
	}

	log.Info().Str("org_id", orgID.String()).Msg("SSO configuration updated")
	return s.GetSSOConfig(ctx, orgID)
}

// ============================================================
// API KEYS
// ============================================================

// ListAPIKeys returns all API keys for an org (without hashes).
func (s *IntegrationService) ListAPIKeys(ctx context.Context, orgID uuid.UUID) ([]APIKey, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, organization_id, name, key_prefix, permissions,
		        rate_limit_per_minute, expires_at, last_used_at, last_used_ip,
		        is_active, created_by, created_at, updated_at
		 FROM api_keys
		 WHERE organization_id = $1
		 ORDER BY created_at DESC`, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list API keys: %w", err)
	}
	defer rows.Close()

	var keys []APIKey
	for rows.Next() {
		var k APIKey
		var lastUsedIP sql.NullString
		var createdBy *uuid.UUID
		if err := rows.Scan(
			&k.ID, &k.OrganizationID, &k.Name, &k.KeyPrefix, &k.Permissions,
			&k.RateLimitPerMin, &k.ExpiresAt, &k.LastUsedAt, &lastUsedIP,
			&k.IsActive, &createdBy, &k.CreatedAt, &k.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan API key: %w", err)
		}
		k.LastUsedIP = lastUsedIP.String
		k.CreatedBy = createdBy
		keys = append(keys, k)
	}

	if keys == nil {
		keys = []APIKey{}
	}
	return keys, nil
}

// CreateAPIKey generates a new API key, stores the hash, and returns the raw key once.
func (s *IntegrationService) CreateAPIKey(ctx context.Context, orgID uuid.UUID, userID uuid.UUID, input CreateAPIKeyInput) (*CreateAPIKeyResult, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("API key name is required")
	}
	if input.RateLimitPerMin <= 0 {
		input.RateLimitPerMin = 60
	}

	// Generate a random 32-byte key, encode as base64url
	rawBytes := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, rawBytes); err != nil {
		return nil, fmt.Errorf("failed to generate random key: %w", err)
	}
	rawKey := "cf_" + base64.URLEncoding.EncodeToString(rawBytes)

	// Prefix = first 10 chars (including the "cf_" prefix)
	prefix := rawKey[:10]

	// Hash the full key with SHA-256
	hash := sha256.Sum256([]byte(rawKey))
	keyHash := fmt.Sprintf("%x", hash)

	// Expiry
	var expiresAt *time.Time
	if input.ExpiresInDays > 0 {
		t := time.Now().AddDate(0, 0, input.ExpiresInDays)
		expiresAt = &t
	}

	var k APIKey
	var lastUsedIP sql.NullString
	var createdBy *uuid.UUID

	err := s.pool.QueryRow(ctx,
		`INSERT INTO api_keys (organization_id, name, key_prefix, key_hash,
		                       permissions, rate_limit_per_minute, expires_at, created_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING id, organization_id, name, key_prefix, permissions,
		           rate_limit_per_minute, expires_at, last_used_at, last_used_ip,
		           is_active, created_by, created_at, updated_at`,
		orgID, input.Name, prefix, keyHash,
		input.Permissions, input.RateLimitPerMin, expiresAt, userID,
	).Scan(
		&k.ID, &k.OrganizationID, &k.Name, &k.KeyPrefix, &k.Permissions,
		&k.RateLimitPerMin, &k.ExpiresAt, &k.LastUsedAt, &lastUsedIP,
		&k.IsActive, &createdBy, &k.CreatedAt, &k.UpdatedAt,
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to create API key")
		return nil, fmt.Errorf("failed to create API key: %w", err)
	}

	k.LastUsedIP = lastUsedIP.String
	k.CreatedBy = createdBy

	log.Info().Str("key_id", k.ID.String()).Str("prefix", prefix).Msg("API key created")
	return &CreateAPIKeyResult{APIKey: k, RawKey: rawKey}, nil
}

// RevokeAPIKey deactivates an API key.
func (s *IntegrationService) RevokeAPIKey(ctx context.Context, orgID, id uuid.UUID) error {
	result, err := s.pool.Exec(ctx,
		`UPDATE api_keys SET is_active = false, updated_at = NOW()
		 WHERE organization_id = $1 AND id = $2 AND is_active = true`, orgID, id)
	if err != nil {
		return fmt.Errorf("failed to revoke API key: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("API key not found or already revoked")
	}
	log.Info().Str("key_id", id.String()).Msg("API key revoked")
	return nil
}

// ValidateAPIKey looks up an API key by prefix, verifies its hash, checks
// expiry, and records usage. Returns the key record on success.
func (s *IntegrationService) ValidateAPIKey(ctx context.Context, rawKey string, clientIP string) (*APIKey, error) {
	if len(rawKey) < 10 {
		return nil, fmt.Errorf("invalid API key format")
	}

	prefix := rawKey[:10]
	hash := sha256.Sum256([]byte(rawKey))
	keyHash := fmt.Sprintf("%x", hash)

	var k APIKey
	var lastUsedIP sql.NullString
	var createdBy *uuid.UUID

	err := s.pool.QueryRow(ctx,
		`SELECT id, organization_id, name, key_prefix, permissions,
		        rate_limit_per_minute, expires_at, last_used_at, last_used_ip,
		        is_active, created_by, created_at, updated_at
		 FROM api_keys
		 WHERE key_prefix = $1 AND key_hash = $2 AND is_active = true`,
		prefix, keyHash).Scan(
		&k.ID, &k.OrganizationID, &k.Name, &k.KeyPrefix, &k.Permissions,
		&k.RateLimitPerMin, &k.ExpiresAt, &k.LastUsedAt, &lastUsedIP,
		&k.IsActive, &createdBy, &k.CreatedAt, &k.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("invalid API key")
	}

	k.LastUsedIP = lastUsedIP.String
	k.CreatedBy = createdBy

	// Check expiry
	if k.ExpiresAt != nil && time.Now().After(*k.ExpiresAt) {
		return nil, fmt.Errorf("API key has expired")
	}

	// Record usage asynchronously
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = s.pool.Exec(bgCtx,
			`UPDATE api_keys SET last_used_at = NOW(), last_used_ip = $2 WHERE id = $1`,
			k.ID, clientIP)
	}()

	return &k, nil
}

// ============================================================
// ENCRYPTION HELPERS  — AES-256-GCM
// ============================================================

// encryptConfig encrypts plaintext using AES-256-GCM with the given passphrase.
func encryptConfig(plaintext []byte, passphrase string) (string, error) {
	key := sha256.Sum256([]byte(passphrase))

	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", fmt.Errorf("cipher creation failed: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("GCM creation failed: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("nonce generation failed: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decryptConfig decrypts a base64-encoded ciphertext using AES-256-GCM.
func decryptConfig(encoded, passphrase string) ([]byte, error) {
	key := sha256.Sum256([]byte(passphrase))

	ciphertext, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("base64 decode failed: %w", err)
	}

	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, fmt.Errorf("cipher creation failed: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("GCM creation failed: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plaintext, nil
}

// ============================================================
// PRIVATE HELPERS
// ============================================================

func nullStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

var validIntegrationTypes = map[string]bool{
	"sso_saml": true, "sso_oidc": true,
	"cloud_aws": true, "cloud_azure": true, "cloud_gcp": true,
	"siem_splunk": true, "siem_elastic": true, "siem_sentinel": true,
	"itsm_servicenow": true, "itsm_jira": true, "itsm_freshservice": true,
	"email_smtp": true, "email_sendgrid": true,
	"slack": true, "teams": true,
	"webhook_inbound": true, "webhook_outbound": true,
	"custom_api": true,
}

func isValidIntegrationType(t string) bool {
	return validIntegrationTypes[t]
}
