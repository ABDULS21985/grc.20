package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// ============================================================
// DEVELOPER PORTAL SERVICE
// Manages API key lifecycle, usage analytics, sandbox
// environments, and developer documentation metadata.
// ============================================================

// DeveloperPortalService handles developer portal business logic.
type DeveloperPortalService struct {
	pool *pgxpool.Pool
}

// NewDeveloperPortalService creates a new DeveloperPortalService.
func NewDeveloperPortalService(pool *pgxpool.Pool) *DeveloperPortalService {
	return &DeveloperPortalService{pool: pool}
}

// ============================================================
// DATA TYPES — API Keys
// ============================================================

// APIKeyResponse is returned when generating a new API key.
// The plaintext key is shown ONLY on creation.
type APIKeyResponse struct {
	ID               uuid.UUID  `json:"id"`
	OrganizationID   uuid.UUID  `json:"organization_id"`
	Name             string     `json:"name"`
	KeyPrefix        string     `json:"key_prefix"`
	Key              string     `json:"key,omitempty"` // only on creation
	Scope            []string   `json:"scope"`
	Tier             string     `json:"tier"`
	RateLimitPerMin  int        `json:"rate_limit_per_minute"`
	RateLimitPerDay  int        `json:"rate_limit_per_day"`
	AllowedIPRanges  []string   `json:"allowed_ip_ranges"`
	AllowedOrigins   []string   `json:"allowed_origins"`
	ExpiresAt        *time.Time `json:"expires_at,omitempty"`
	IsActive         bool       `json:"is_active"`
	CreatedAt        time.Time  `json:"created_at"`
}

// APIKeyInfo represents an API key in list/detail views (no plaintext).
type APIKeyInfo struct {
	ID               uuid.UUID       `json:"id"`
	OrganizationID   uuid.UUID       `json:"organization_id"`
	Name             string          `json:"name"`
	KeyPrefix        string          `json:"key_prefix"`
	Scope            []string        `json:"scope"`
	Tier             string          `json:"tier"`
	RateLimitPerMin  int             `json:"rate_limit_per_minute"`
	RateLimitPerDay  int             `json:"rate_limit_per_day"`
	AllowedIPRanges  []string        `json:"allowed_ip_ranges"`
	AllowedOrigins   []string        `json:"allowed_origins"`
	Metadata         json.RawMessage `json:"metadata"`
	ExpiresAt        *time.Time      `json:"expires_at,omitempty"`
	LastUsedAt       *time.Time      `json:"last_used_at,omitempty"`
	LastUsedIP       *string         `json:"last_used_ip,omitempty"`
	IsActive         bool            `json:"is_active"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

// GenerateAPIKeyReq holds the request payload for generating a new API key.
type GenerateAPIKeyReq struct {
	Name            string   `json:"name"`
	Scope           []string `json:"scope"`
	Tier            string   `json:"tier"`
	RateLimitPerMin int      `json:"rate_limit_per_minute"`
	RateLimitPerDay int      `json:"rate_limit_per_day"`
	AllowedIPRanges []string `json:"allowed_ip_ranges"`
	AllowedOrigins  []string `json:"allowed_origins"`
	ExpiresInDays   int      `json:"expires_in_days"`
}

// UpdateAPIKeyReq holds the request payload for updating an API key.
type UpdateAPIKeyReq struct {
	Name            *string  `json:"name,omitempty"`
	Scope           []string `json:"scope,omitempty"`
	RateLimitPerMin *int     `json:"rate_limit_per_minute,omitempty"`
	RateLimitPerDay *int     `json:"rate_limit_per_day,omitempty"`
	AllowedIPRanges []string `json:"allowed_ip_ranges,omitempty"`
	AllowedOrigins  []string `json:"allowed_origins,omitempty"`
	IsActive        *bool    `json:"is_active,omitempty"`
}

// ============================================================
// DATA TYPES — Usage Stats
// ============================================================

// UsageStats holds aggregated API usage statistics for a key.
type UsageStats struct {
	KeyID            uuid.UUID      `json:"key_id"`
	Period           string         `json:"period"`
	TotalRequests    int64          `json:"total_requests"`
	SuccessCount     int64          `json:"success_count"`
	ErrorCount       int64          `json:"error_count"`
	AvgDurationMs    float64        `json:"avg_duration_ms"`
	TopEndpoints     []EndpointStat `json:"top_endpoints"`
	RequestsByDay    []DailyStat    `json:"requests_by_day"`
	ErrorsByCode     map[string]int `json:"errors_by_code"`
}

// EndpointStat holds request count per endpoint.
type EndpointStat struct {
	Method string `json:"method"`
	Path   string `json:"path"`
	Count  int64  `json:"count"`
}

// DailyStat holds request counts per day.
type DailyStat struct {
	Date    string `json:"date"`
	Count   int64  `json:"count"`
	Errors  int64  `json:"errors"`
}

// ============================================================
// DATA TYPES — Sandbox
// ============================================================

// SandboxEnvironment represents an API sandbox.
type SandboxEnvironment struct {
	ID             uuid.UUID       `json:"id"`
	OrganizationID uuid.UUID      `json:"organization_id"`
	Name           string          `json:"name"`
	Status         string          `json:"status"`
	APIBaseURL     string          `json:"api_base_url"`
	SandboxAPIKey  *string         `json:"sandbox_api_key,omitempty"`
	SeedDataLoaded bool            `json:"seed_data_loaded"`
	ExpiresAt      time.Time       `json:"expires_at"`
	LastAccessedAt *time.Time      `json:"last_accessed_at,omitempty"`
	Metadata       json.RawMessage `json:"metadata"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// ============================================================
// DATA TYPES — Documentation Metadata
// ============================================================

// EventTypeInfo describes a webhook event type for documentation.
type EventTypeInfo struct {
	EventType   string `json:"event_type"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Version     string `json:"version"`
}

// ScopeInfo describes an API scope for documentation.
type ScopeInfo struct {
	Scope       string `json:"scope"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Access      string `json:"access"` // read, write, delete
}

// ============================================================
// API KEY GENERATION — cf_live_ prefix, SHA-256 hash
// ============================================================

// GenerateAPIKeyString creates a cryptographically random API key with cf_live_ prefix.
func GenerateAPIKeyString() (fullKey, prefix, hash string, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	keyBody := hex.EncodeToString(b)
	fullKey = "cf_live_" + keyBody
	prefix = fullKey[:16] // "cf_live_" + first 8 hex chars

	h := sha256.Sum256([]byte(fullKey))
	hash = hex.EncodeToString(h[:])

	return fullKey, prefix, hash, nil
}

// HashAPIKey computes the SHA-256 hash of an API key string.
func HashAPIKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}

// ============================================================
// API KEY CRUD
// ============================================================

// GenerateAPIKey creates a new API key for the organisation.
func (s *DeveloperPortalService) GenerateAPIKey(ctx context.Context, orgID uuid.UUID, req GenerateAPIKeyReq) (*APIKeyResponse, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("API key name is required")
	}

	fullKey, prefix, keyHash, err := GenerateAPIKeyString()
	if err != nil {
		return nil, err
	}

	tier := req.Tier
	if tier == "" {
		tier = "standard"
	}

	ratePerMin := req.RateLimitPerMin
	if ratePerMin <= 0 {
		ratePerMin = 60
	}

	ratePerDay := req.RateLimitPerDay
	if ratePerDay <= 0 {
		ratePerDay = 10000
	}

	scope := req.Scope
	if scope == nil {
		scope = []string{}
	}

	ipRanges := req.AllowedIPRanges
	if ipRanges == nil {
		ipRanges = []string{}
	}

	origins := req.AllowedOrigins
	if origins == nil {
		origins = []string{}
	}

	var expiresAt *time.Time
	if req.ExpiresInDays > 0 {
		t := time.Now().AddDate(0, 0, req.ExpiresInDays)
		expiresAt = &t
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return nil, fmt.Errorf("failed to set org context: %w", err)
	}

	var id uuid.UUID
	var createdAt time.Time
	err = tx.QueryRow(ctx, `
		INSERT INTO api_keys (
			organization_id, name, key_prefix, key_hash, permissions,
			scope, tier, rate_limit_per_minute, rate_limit_per_day,
			allowed_ip_ranges, allowed_origins, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at
	`,
		orgID, req.Name, prefix, keyHash, scope,
		scope, tier, ratePerMin, ratePerDay,
		ipRanges, origins, expiresAt,
	).Scan(&id, &createdAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create API key: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}

	log.Info().
		Str("key_id", id.String()).
		Str("org_id", orgID.String()).
		Str("prefix", prefix).
		Str("tier", tier).
		Msg("API key generated")

	return &APIKeyResponse{
		ID:              id,
		OrganizationID:  orgID,
		Name:            req.Name,
		KeyPrefix:       prefix,
		Key:             fullKey, // only shown on creation
		Scope:           scope,
		Tier:            tier,
		RateLimitPerMin: ratePerMin,
		RateLimitPerDay: ratePerDay,
		AllowedIPRanges: ipRanges,
		AllowedOrigins:  origins,
		ExpiresAt:       expiresAt,
		IsActive:        true,
		CreatedAt:       createdAt,
	}, nil
}

// ListAPIKeys returns all API keys for an organisation (no plaintext).
func (s *DeveloperPortalService) ListAPIKeys(ctx context.Context, orgID uuid.UUID) ([]APIKeyInfo, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return nil, fmt.Errorf("failed to set org context: %w", err)
	}

	rows, err := tx.Query(ctx, `
		SELECT id, organization_id, name, key_prefix, scope, tier,
		       rate_limit_per_minute, rate_limit_per_day, allowed_ip_ranges,
		       allowed_origins, metadata, expires_at, last_used_at,
		       last_used_ip, is_active, created_at, updated_at
		FROM api_keys
		WHERE organization_id = $1
		ORDER BY created_at DESC
	`, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list API keys: %w", err)
	}
	defer rows.Close()

	var keys []APIKeyInfo
	for rows.Next() {
		var k APIKeyInfo
		if err := rows.Scan(
			&k.ID, &k.OrganizationID, &k.Name, &k.KeyPrefix, &k.Scope,
			&k.Tier, &k.RateLimitPerMin, &k.RateLimitPerDay, &k.AllowedIPRanges,
			&k.AllowedOrigins, &k.Metadata, &k.ExpiresAt, &k.LastUsedAt,
			&k.LastUsedIP, &k.IsActive, &k.CreatedAt, &k.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan API key: %w", err)
		}
		keys = append(keys, k)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}

	if keys == nil {
		keys = []APIKeyInfo{}
	}
	return keys, nil
}

// UpdateAPIKey updates an API key's settings.
func (s *DeveloperPortalService) UpdateAPIKey(ctx context.Context, orgID, keyID uuid.UUID, req UpdateAPIKeyReq) (*APIKeyInfo, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return nil, fmt.Errorf("failed to set org context: %w", err)
	}

	sets := []string{"updated_at = NOW()"}
	args := []interface{}{}
	argIdx := 1

	if req.Name != nil {
		sets = append(sets, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, *req.Name)
		argIdx++
	}
	if req.Scope != nil {
		sets = append(sets, fmt.Sprintf("scope = $%d", argIdx))
		args = append(args, req.Scope)
		argIdx++
		// Also update permissions to keep them in sync
		sets = append(sets, fmt.Sprintf("permissions = $%d", argIdx))
		args = append(args, req.Scope)
		argIdx++
	}
	if req.RateLimitPerMin != nil {
		sets = append(sets, fmt.Sprintf("rate_limit_per_minute = $%d", argIdx))
		args = append(args, *req.RateLimitPerMin)
		argIdx++
	}
	if req.RateLimitPerDay != nil {
		sets = append(sets, fmt.Sprintf("rate_limit_per_day = $%d", argIdx))
		args = append(args, *req.RateLimitPerDay)
		argIdx++
	}
	if req.AllowedIPRanges != nil {
		sets = append(sets, fmt.Sprintf("allowed_ip_ranges = $%d", argIdx))
		args = append(args, req.AllowedIPRanges)
		argIdx++
	}
	if req.AllowedOrigins != nil {
		sets = append(sets, fmt.Sprintf("allowed_origins = $%d", argIdx))
		args = append(args, req.AllowedOrigins)
		argIdx++
	}
	if req.IsActive != nil {
		sets = append(sets, fmt.Sprintf("is_active = $%d", argIdx))
		args = append(args, *req.IsActive)
		argIdx++
	}

	query := fmt.Sprintf(
		`UPDATE api_keys SET %s WHERE id = $%d AND organization_id = $%d
		 RETURNING id, organization_id, name, key_prefix, scope, tier,
		           rate_limit_per_minute, rate_limit_per_day, allowed_ip_ranges,
		           allowed_origins, metadata, expires_at, last_used_at,
		           last_used_ip, is_active, created_at, updated_at`,
		joinStrings(sets, ", "), argIdx, argIdx+1,
	)
	args = append(args, keyID, orgID)

	var k APIKeyInfo
	err = tx.QueryRow(ctx, query, args...).Scan(
		&k.ID, &k.OrganizationID, &k.Name, &k.KeyPrefix, &k.Scope,
		&k.Tier, &k.RateLimitPerMin, &k.RateLimitPerDay, &k.AllowedIPRanges,
		&k.AllowedOrigins, &k.Metadata, &k.ExpiresAt, &k.LastUsedAt,
		&k.LastUsedIP, &k.IsActive, &k.CreatedAt, &k.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("API key not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update API key: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}

	return &k, nil
}

// RevokeAPIKey deactivates an API key (soft delete).
func (s *DeveloperPortalService) RevokeAPIKey(ctx context.Context, orgID, keyID uuid.UUID) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return fmt.Errorf("failed to set org context: %w", err)
	}

	tag, err := tx.Exec(ctx, `
		UPDATE api_keys SET is_active = false, updated_at = NOW()
		WHERE id = $1 AND organization_id = $2
	`, keyID, orgID)
	if err != nil {
		return fmt.Errorf("failed to revoke API key: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("API key not found")
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	log.Info().
		Str("key_id", keyID.String()).
		Str("org_id", orgID.String()).
		Msg("API key revoked")

	return nil
}

// ============================================================
// API USAGE STATS
// ============================================================

// GetAPIUsageStats returns aggregated usage statistics for an API key.
func (s *DeveloperPortalService) GetAPIUsageStats(ctx context.Context, orgID, keyID uuid.UUID, period string) (*UsageStats, error) {
	if period == "" {
		period = "7d"
	}

	var interval string
	switch period {
	case "24h":
		interval = "24 hours"
	case "7d":
		interval = "7 days"
	case "30d":
		interval = "30 days"
	case "90d":
		interval = "90 days"
	default:
		interval = "7 days"
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return nil, fmt.Errorf("failed to set org context: %w", err)
	}

	stats := &UsageStats{
		KeyID:  keyID,
		Period: period,
	}

	// Total requests, success, error counts, avg duration
	err = tx.QueryRow(ctx, fmt.Sprintf(`
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE status_code >= 200 AND status_code < 400),
			COUNT(*) FILTER (WHERE status_code >= 400),
			COALESCE(AVG(duration_ms), 0)
		FROM api_usage_logs
		WHERE api_key_id = $1
		  AND organization_id = $2
		  AND created_at >= NOW() - INTERVAL '%s'
	`, interval), keyID, orgID).Scan(
		&stats.TotalRequests, &stats.SuccessCount, &stats.ErrorCount, &stats.AvgDurationMs,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage totals: %w", err)
	}

	// Top endpoints
	rows, err := tx.Query(ctx, fmt.Sprintf(`
		SELECT method, path, COUNT(*) as cnt
		FROM api_usage_logs
		WHERE api_key_id = $1
		  AND organization_id = $2
		  AND created_at >= NOW() - INTERVAL '%s'
		GROUP BY method, path
		ORDER BY cnt DESC
		LIMIT 10
	`, interval), keyID, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get top endpoints: %w", err)
	}
	defer rows.Close()

	stats.TopEndpoints = []EndpointStat{}
	for rows.Next() {
		var ep EndpointStat
		if err := rows.Scan(&ep.Method, &ep.Path, &ep.Count); err != nil {
			return nil, fmt.Errorf("failed to scan endpoint stat: %w", err)
		}
		stats.TopEndpoints = append(stats.TopEndpoints, ep)
	}

	// Requests by day
	dayRows, err := tx.Query(ctx, fmt.Sprintf(`
		SELECT
			DATE(created_at) as day,
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status_code >= 400) as errors
		FROM api_usage_logs
		WHERE api_key_id = $1
		  AND organization_id = $2
		  AND created_at >= NOW() - INTERVAL '%s'
		GROUP BY day
		ORDER BY day ASC
	`, interval), keyID, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily stats: %w", err)
	}
	defer dayRows.Close()

	stats.RequestsByDay = []DailyStat{}
	for dayRows.Next() {
		var ds DailyStat
		var day time.Time
		if err := dayRows.Scan(&day, &ds.Count, &ds.Errors); err != nil {
			return nil, fmt.Errorf("failed to scan daily stat: %w", err)
		}
		ds.Date = day.Format("2006-01-02")
		stats.RequestsByDay = append(stats.RequestsByDay, ds)
	}

	// Errors by code
	errRows, err := tx.Query(ctx, fmt.Sprintf(`
		SELECT COALESCE(error_code, 'UNKNOWN'), COUNT(*)
		FROM api_usage_logs
		WHERE api_key_id = $1
		  AND organization_id = $2
		  AND status_code >= 400
		  AND created_at >= NOW() - INTERVAL '%s'
		GROUP BY error_code
		ORDER BY COUNT(*) DESC
		LIMIT 20
	`, interval), keyID, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get error codes: %w", err)
	}
	defer errRows.Close()

	stats.ErrorsByCode = map[string]int{}
	for errRows.Next() {
		var code string
		var count int
		if err := errRows.Scan(&code, &count); err != nil {
			return nil, fmt.Errorf("failed to scan error code: %w", err)
		}
		stats.ErrorsByCode[code] = count
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}

	return stats, nil
}

// ============================================================
// SANDBOX MANAGEMENT
// ============================================================

// CreateSandbox provisions a new sandbox environment for the organisation.
func (s *DeveloperPortalService) CreateSandbox(ctx context.Context, orgID uuid.UUID) (*SandboxEnvironment, error) {
	// Generate a sandbox API key
	sandboxKey := "cf_sandbox_" + uuid.New().String()[:16]
	apiBaseURL := fmt.Sprintf("https://sandbox.complianceforge.io/api/v1/%s", orgID.String()[:8])

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return nil, fmt.Errorf("failed to set org context: %w", err)
	}

	var sb SandboxEnvironment
	err = tx.QueryRow(ctx, `
		INSERT INTO sandbox_environments (
			organization_id, name, status, api_base_url, sandbox_api_key, seed_data_loaded
		) VALUES ($1, 'Default Sandbox', 'active', $2, $3, true)
		ON CONFLICT (organization_id)
		DO UPDATE SET
			status = 'active',
			api_base_url = EXCLUDED.api_base_url,
			sandbox_api_key = EXCLUDED.sandbox_api_key,
			seed_data_loaded = true,
			expires_at = NOW() + INTERVAL '30 days',
			updated_at = NOW()
		RETURNING id, organization_id, name, status, api_base_url, sandbox_api_key,
		          seed_data_loaded, expires_at, last_accessed_at, metadata,
		          created_at, updated_at
	`, orgID, apiBaseURL, sandboxKey).Scan(
		&sb.ID, &sb.OrganizationID, &sb.Name, &sb.Status, &sb.APIBaseURL,
		&sb.SandboxAPIKey, &sb.SeedDataLoaded, &sb.ExpiresAt, &sb.LastAccessedAt,
		&sb.Metadata, &sb.CreatedAt, &sb.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create sandbox: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}

	log.Info().
		Str("sandbox_id", sb.ID.String()).
		Str("org_id", orgID.String()).
		Msg("Sandbox environment created")

	return &sb, nil
}

// GetSandbox returns the current sandbox environment for an organisation.
func (s *DeveloperPortalService) GetSandbox(ctx context.Context, orgID uuid.UUID) (*SandboxEnvironment, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return nil, fmt.Errorf("failed to set org context: %w", err)
	}

	var sb SandboxEnvironment
	err = tx.QueryRow(ctx, `
		SELECT id, organization_id, name, status, api_base_url, sandbox_api_key,
		       seed_data_loaded, expires_at, last_accessed_at, metadata,
		       created_at, updated_at
		FROM sandbox_environments
		WHERE organization_id = $1
	`, orgID).Scan(
		&sb.ID, &sb.OrganizationID, &sb.Name, &sb.Status, &sb.APIBaseURL,
		&sb.SandboxAPIKey, &sb.SeedDataLoaded, &sb.ExpiresAt, &sb.LastAccessedAt,
		&sb.Metadata, &sb.CreatedAt, &sb.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("no sandbox environment found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get sandbox: %w", err)
	}

	// Update last accessed
	_, _ = tx.Exec(ctx, `
		UPDATE sandbox_environments SET last_accessed_at = NOW()
		WHERE id = $1
	`, sb.ID)

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}

	return &sb, nil
}

// DestroySandbox tears down the sandbox environment.
func (s *DeveloperPortalService) DestroySandbox(ctx context.Context, orgID uuid.UUID) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return fmt.Errorf("failed to set org context: %w", err)
	}

	tag, err := tx.Exec(ctx, `
		UPDATE sandbox_environments SET status = 'destroyed', updated_at = NOW()
		WHERE organization_id = $1 AND status = 'active'
	`, orgID)
	if err != nil {
		return fmt.Errorf("failed to destroy sandbox: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("no active sandbox found")
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	log.Info().Str("org_id", orgID.String()).Msg("Sandbox environment destroyed")
	return nil
}

// ============================================================
// DOCUMENTATION METADATA
// ============================================================

// ListWebhookEventTypes returns all available webhook event types.
func (s *DeveloperPortalService) ListWebhookEventTypes() []EventTypeInfo {
	return []EventTypeInfo{
		// Risk Management
		{EventType: "risk.created", Category: "Risk Management", Description: "A new risk has been created", Version: "2024-01-01"},
		{EventType: "risk.updated", Category: "Risk Management", Description: "A risk has been updated", Version: "2024-01-01"},
		{EventType: "risk.deleted", Category: "Risk Management", Description: "A risk has been deleted", Version: "2024-01-01"},
		{EventType: "risk.status_changed", Category: "Risk Management", Description: "A risk status has changed", Version: "2024-01-01"},
		{EventType: "risk.review_due", Category: "Risk Management", Description: "A risk review is due", Version: "2024-01-01"},
		{EventType: "risk.score_changed", Category: "Risk Management", Description: "A risk score has changed", Version: "2024-01-01"},
		// Control Management
		{EventType: "control.created", Category: "Control Management", Description: "A new control has been created", Version: "2024-01-01"},
		{EventType: "control.updated", Category: "Control Management", Description: "A control has been updated", Version: "2024-01-01"},
		{EventType: "control.status_changed", Category: "Control Management", Description: "A control implementation status has changed", Version: "2024-01-01"},
		{EventType: "control.evidence_attached", Category: "Control Management", Description: "Evidence has been attached to a control", Version: "2024-01-01"},
		{EventType: "control.test_completed", Category: "Control Management", Description: "A control test has been completed", Version: "2024-01-01"},
		// Policy Management
		{EventType: "policy.created", Category: "Policy Management", Description: "A new policy has been created", Version: "2024-01-01"},
		{EventType: "policy.updated", Category: "Policy Management", Description: "A policy has been updated", Version: "2024-01-01"},
		{EventType: "policy.published", Category: "Policy Management", Description: "A policy has been published", Version: "2024-01-01"},
		{EventType: "policy.approved", Category: "Policy Management", Description: "A policy has been approved", Version: "2024-01-01"},
		{EventType: "policy.review_due", Category: "Policy Management", Description: "A policy review is coming due", Version: "2024-01-01"},
		{EventType: "policy.acknowledged", Category: "Policy Management", Description: "A policy has been acknowledged by a user", Version: "2024-01-01"},
		// Audit Management
		{EventType: "audit.created", Category: "Audit Management", Description: "A new audit has been created", Version: "2024-01-01"},
		{EventType: "audit.completed", Category: "Audit Management", Description: "An audit has been completed", Version: "2024-01-01"},
		{EventType: "audit.finding_created", Category: "Audit Management", Description: "A new audit finding has been created", Version: "2024-01-01"},
		{EventType: "audit.finding_resolved", Category: "Audit Management", Description: "An audit finding has been resolved", Version: "2024-01-01"},
		// Incident Management
		{EventType: "incident.created", Category: "Incident Management", Description: "A new incident has been reported", Version: "2024-01-01"},
		{EventType: "incident.updated", Category: "Incident Management", Description: "An incident has been updated", Version: "2024-01-01"},
		{EventType: "incident.resolved", Category: "Incident Management", Description: "An incident has been resolved", Version: "2024-01-01"},
		{EventType: "incident.escalated", Category: "Incident Management", Description: "An incident has been escalated", Version: "2024-01-01"},
		{EventType: "incident.breach_detected", Category: "Incident Management", Description: "A data breach has been detected", Version: "2024-01-01"},
		// Compliance
		{EventType: "compliance.score_changed", Category: "Compliance", Description: "Overall compliance score has changed", Version: "2024-01-01"},
		{EventType: "compliance.framework_mapped", Category: "Compliance", Description: "A framework mapping has been updated", Version: "2024-01-01"},
		{EventType: "compliance.gap_identified", Category: "Compliance", Description: "A compliance gap has been identified", Version: "2024-01-01"},
		{EventType: "compliance.deadline_approaching", Category: "Compliance", Description: "A compliance deadline is approaching", Version: "2024-01-01"},
		// Vendor/TPRM
		{EventType: "vendor.created", Category: "Vendor Management", Description: "A new vendor has been added", Version: "2024-01-01"},
		{EventType: "vendor.risk_changed", Category: "Vendor Management", Description: "Vendor risk level has changed", Version: "2024-01-01"},
		{EventType: "vendor.assessment_completed", Category: "Vendor Management", Description: "A vendor assessment has been completed", Version: "2024-01-01"},
		{EventType: "vendor.contract_expiring", Category: "Vendor Management", Description: "A vendor contract is expiring", Version: "2024-01-01"},
		// Evidence
		{EventType: "evidence.uploaded", Category: "Evidence", Description: "New evidence has been uploaded", Version: "2024-01-01"},
		{EventType: "evidence.expired", Category: "Evidence", Description: "Evidence has expired", Version: "2024-01-01"},
		{EventType: "evidence.review_required", Category: "Evidence", Description: "Evidence requires review", Version: "2024-01-01"},
		// Regulatory
		{EventType: "regulatory.change_detected", Category: "Regulatory", Description: "A regulatory change has been detected", Version: "2024-01-01"},
		{EventType: "regulatory.impact_assessed", Category: "Regulatory", Description: "Impact of a regulatory change has been assessed", Version: "2024-01-01"},
		{EventType: "regulatory.deadline_approaching", Category: "Regulatory", Description: "A regulatory deadline is approaching", Version: "2024-01-01"},
		// User & Organisation
		{EventType: "user.created", Category: "User Management", Description: "A new user has been created", Version: "2024-01-01"},
		{EventType: "user.deactivated", Category: "User Management", Description: "A user has been deactivated", Version: "2024-01-01"},
		{EventType: "user.role_changed", Category: "User Management", Description: "A user role has been changed", Version: "2024-01-01"},
		// Workflow
		{EventType: "workflow.started", Category: "Workflow", Description: "A workflow has been started", Version: "2024-01-01"},
		{EventType: "workflow.completed", Category: "Workflow", Description: "A workflow has been completed", Version: "2024-01-01"},
		{EventType: "workflow.step_completed", Category: "Workflow", Description: "A workflow step has been completed", Version: "2024-01-01"},
		{EventType: "workflow.approval_required", Category: "Workflow", Description: "A workflow step requires approval", Version: "2024-01-01"},
		// System
		{EventType: "ping", Category: "System", Description: "Connectivity test event", Version: "2024-01-01"},
		{EventType: "api_key.created", Category: "System", Description: "A new API key has been created", Version: "2024-01-01"},
		{EventType: "api_key.revoked", Category: "System", Description: "An API key has been revoked", Version: "2024-01-01"},
	}
}

// ListAPIScopes returns all available API scopes.
func (s *DeveloperPortalService) ListAPIScopes() []ScopeInfo {
	return []ScopeInfo{
		// Risk
		{Scope: "risks:read", Category: "Risk Management", Description: "Read risk records", Access: "read"},
		{Scope: "risks:write", Category: "Risk Management", Description: "Create and update risk records", Access: "write"},
		{Scope: "risks:delete", Category: "Risk Management", Description: "Delete risk records", Access: "delete"},
		// Controls
		{Scope: "controls:read", Category: "Control Management", Description: "Read control records", Access: "read"},
		{Scope: "controls:write", Category: "Control Management", Description: "Create and update control records", Access: "write"},
		{Scope: "controls:delete", Category: "Control Management", Description: "Delete control records", Access: "delete"},
		// Policies
		{Scope: "policies:read", Category: "Policy Management", Description: "Read policy documents", Access: "read"},
		{Scope: "policies:write", Category: "Policy Management", Description: "Create and update policy documents", Access: "write"},
		{Scope: "policies:delete", Category: "Policy Management", Description: "Delete policy documents", Access: "delete"},
		// Audits
		{Scope: "audits:read", Category: "Audit Management", Description: "Read audit records and findings", Access: "read"},
		{Scope: "audits:write", Category: "Audit Management", Description: "Create and update audits", Access: "write"},
		{Scope: "audits:delete", Category: "Audit Management", Description: "Delete audit records", Access: "delete"},
		// Incidents
		{Scope: "incidents:read", Category: "Incident Management", Description: "Read incident reports", Access: "read"},
		{Scope: "incidents:write", Category: "Incident Management", Description: "Create and update incidents", Access: "write"},
		{Scope: "incidents:delete", Category: "Incident Management", Description: "Delete incident reports", Access: "delete"},
		// Vendors
		{Scope: "vendors:read", Category: "Vendor Management", Description: "Read vendor information", Access: "read"},
		{Scope: "vendors:write", Category: "Vendor Management", Description: "Create and update vendor records", Access: "write"},
		{Scope: "vendors:delete", Category: "Vendor Management", Description: "Delete vendor records", Access: "delete"},
		// Evidence
		{Scope: "evidence:read", Category: "Evidence", Description: "Read evidence records", Access: "read"},
		{Scope: "evidence:write", Category: "Evidence", Description: "Upload and update evidence", Access: "write"},
		{Scope: "evidence:delete", Category: "Evidence", Description: "Delete evidence records", Access: "delete"},
		// Compliance
		{Scope: "compliance:read", Category: "Compliance", Description: "Read compliance data and scores", Access: "read"},
		{Scope: "compliance:write", Category: "Compliance", Description: "Update compliance mappings", Access: "write"},
		// Frameworks
		{Scope: "frameworks:read", Category: "Frameworks", Description: "Read framework definitions", Access: "read"},
		{Scope: "frameworks:write", Category: "Frameworks", Description: "Create custom frameworks", Access: "write"},
		// Users
		{Scope: "users:read", Category: "User Management", Description: "Read user information", Access: "read"},
		{Scope: "users:write", Category: "User Management", Description: "Create and update users", Access: "write"},
		{Scope: "users:delete", Category: "User Management", Description: "Deactivate users", Access: "delete"},
		// Reports
		{Scope: "reports:read", Category: "Reports", Description: "Generate and read reports", Access: "read"},
		{Scope: "reports:write", Category: "Reports", Description: "Create report templates", Access: "write"},
		// Webhooks
		{Scope: "webhooks:read", Category: "Webhooks", Description: "Read webhook subscriptions", Access: "read"},
		{Scope: "webhooks:write", Category: "Webhooks", Description: "Create and update webhooks", Access: "write"},
		{Scope: "webhooks:delete", Category: "Webhooks", Description: "Delete webhook subscriptions", Access: "delete"},
		// Analytics
		{Scope: "analytics:read", Category: "Analytics", Description: "Read analytics and dashboards", Access: "read"},
		// Settings
		{Scope: "settings:read", Category: "Settings", Description: "Read organisation settings", Access: "read"},
		{Scope: "settings:write", Category: "Settings", Description: "Update organisation settings", Access: "write"},
	}
}

// ============================================================
// HELPERS
// ============================================================

// joinStrings joins strings with a separator (avoids importing strings in this file
// since it is already a large file and we want to keep imports minimal).
func joinStrings(parts []string, sep string) string {
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += sep
		}
		result += p
	}
	return result
}
