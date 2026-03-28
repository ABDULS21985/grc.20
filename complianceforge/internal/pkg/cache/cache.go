// Package cache provides a Redis-backed caching layer for ComplianceForge.
// It caches expensive computations like compliance scores, dashboard summaries,
// and framework data to reduce database load.
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"

	"github.com/complianceforge/platform/internal/config"
)

// DefaultTTL is the default cache expiration time.
const DefaultTTL = 5 * time.Minute

// LongTTL is for data that changes infrequently (frameworks, controls).
const LongTTL = 1 * time.Hour

// ShortTTL is for frequently changing data (dashboards, scores).
const ShortTTL = 2 * time.Minute

// Key prefixes for cache namespacing.
const (
	PrefixFramework       = "fw:"
	PrefixFrameworkList   = "fw:list"
	PrefixControls        = "ctrl:"
	PrefixComplianceScore = "score:"
	PrefixDashboard       = "dash:"
	PrefixRiskHeatmap     = "risk:heatmap:"
	PrefixGapAnalysis     = "gap:"
	PrefixReport          = "report:"
	PrefixUser            = "user:"
)

// Client wraps the Redis client with GRC-specific caching methods.
type Client struct {
	rdb *redis.Client
}

// New creates a new Redis cache client.
func New(cfg config.RedisConfig) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr(),
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     20,
		MinIdleConns: 5,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Info().Str("addr", cfg.Addr()).Msg("Connected to Redis")

	return &Client{rdb: rdb}, nil
}

// Close closes the Redis connection.
func (c *Client) Close() error {
	return c.rdb.Close()
}

// Health checks if Redis is reachable.
func (c *Client) Health(ctx context.Context) error {
	return c.rdb.Ping(ctx).Err()
}

// ============================================================
// GENERIC GET / SET / DELETE
// ============================================================

// Get retrieves a cached value and unmarshals it into dest.
// Returns false if the key doesn't exist.
func (c *Client) Get(ctx context.Context, key string, dest interface{}) (bool, error) {
	val, err := c.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("cache get error: %w", err)
	}

	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return false, fmt.Errorf("cache unmarshal error: %w", err)
	}

	return true, nil
}

// Set stores a value in the cache with the given TTL.
func (c *Client) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("cache marshal error: %w", err)
	}

	if err := c.rdb.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("cache set error: %w", err)
	}

	return nil
}

// Delete removes a key from the cache.
func (c *Client) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return c.rdb.Del(ctx, keys...).Err()
}

// DeletePattern removes all keys matching a glob pattern.
func (c *Client) DeletePattern(ctx context.Context, pattern string) error {
	iter := c.rdb.Scan(ctx, 0, pattern, 100).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return err
	}
	if len(keys) > 0 {
		return c.rdb.Del(ctx, keys...).Err()
	}
	return nil
}

// ============================================================
// GRC-SPECIFIC CACHE METHODS
// ============================================================

// FrameworkListKey returns the cache key for the list of all system frameworks.
func FrameworkListKey() string {
	return PrefixFrameworkList
}

// FrameworkKey returns the cache key for a single framework.
func FrameworkKey(frameworkID string) string {
	return PrefixFramework + frameworkID
}

// ControlsKey returns the cache key for a framework's controls.
func ControlsKey(frameworkID string, page, pageSize int) string {
	return fmt.Sprintf("%s%s:p%d:s%d", PrefixControls, frameworkID, page, pageSize)
}

// ComplianceScoreKey returns the cache key for an org's compliance scores.
func ComplianceScoreKey(orgID string) string {
	return PrefixComplianceScore + orgID
}

// DashboardKey returns the cache key for an org's dashboard summary.
func DashboardKey(orgID string) string {
	return PrefixDashboard + orgID
}

// RiskHeatmapKey returns the cache key for an org's risk heatmap.
func RiskHeatmapKey(orgID string) string {
	return PrefixRiskHeatmap + orgID
}

// GapAnalysisKey returns the cache key for an org's gap analysis.
func GapAnalysisKey(orgID string, frameworkID string) string {
	if frameworkID != "" {
		return PrefixGapAnalysis + orgID + ":" + frameworkID
	}
	return PrefixGapAnalysis + orgID + ":all"
}

// ReportKey returns the cache key for a generated report.
func ReportKey(orgID, reportType string) string {
	return PrefixReport + orgID + ":" + reportType
}

// ============================================================
// INVALIDATION HELPERS
// ============================================================

// InvalidateOrgCache clears all cached data for an organization.
// Call this when any compliance data changes (control status, risk, policy, etc.)
func (c *Client) InvalidateOrgCache(ctx context.Context, orgID string) error {
	patterns := []string{
		PrefixComplianceScore + orgID,
		PrefixDashboard + orgID,
		PrefixRiskHeatmap + orgID,
		PrefixGapAnalysis + orgID + ":*",
		PrefixReport + orgID + ":*",
	}

	for _, pattern := range patterns {
		if err := c.DeletePattern(ctx, pattern); err != nil {
			log.Warn().Err(err).Str("pattern", pattern).Msg("Failed to invalidate cache")
		}
	}

	return nil
}

// InvalidateFrameworkCache clears cached framework data.
// Call when frameworks or controls are updated.
func (c *Client) InvalidateFrameworkCache(ctx context.Context) error {
	patterns := []string{
		PrefixFramework + "*",
		PrefixFrameworkList,
		PrefixControls + "*",
	}

	for _, pattern := range patterns {
		if err := c.DeletePattern(ctx, pattern); err != nil {
			log.Warn().Err(err).Str("pattern", pattern).Msg("Failed to invalidate framework cache")
		}
	}

	return nil
}

// ============================================================
// GET-OR-SET PATTERN
// ============================================================

// GetOrSet tries to get a cached value; if not found, calls fn to compute it,
// caches the result, and returns it. This is the primary caching pattern.
func (c *Client) GetOrSet(ctx context.Context, key string, dest interface{}, ttl time.Duration, fn func() (interface{}, error)) error {
	// Try cache first
	found, err := c.Get(ctx, key, dest)
	if err != nil {
		log.Warn().Err(err).Str("key", key).Msg("Cache read error, falling through to source")
	}
	if found {
		return nil
	}

	// Cache miss — compute the value
	result, err := fn()
	if err != nil {
		return err
	}

	// Store in cache (non-blocking, best-effort)
	go func() {
		cacheCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if setErr := c.Set(cacheCtx, key, result, ttl); setErr != nil {
			log.Warn().Err(setErr).Str("key", key).Msg("Failed to cache result")
		}
	}()

	// Unmarshal result into dest
	data, err := json.Marshal(result)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

// ============================================================
// RATE LIMITING (Token Bucket via Redis)
// ============================================================

// CheckRateLimit returns true if the request is within rate limits.
// Uses a sliding window counter per key.
func (c *Client) CheckRateLimit(ctx context.Context, key string, limit int, window time.Duration) (bool, int, error) {
	now := time.Now().UnixMilli()
	windowStart := now - window.Milliseconds()

	pipe := c.rdb.Pipeline()

	// Remove old entries outside the window
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))

	// Add current request
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: fmt.Sprintf("%d", now)})

	// Count requests in window
	countCmd := pipe.ZCard(ctx, key)

	// Set expiry on the key
	pipe.Expire(ctx, key, window)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return true, 0, err // Allow on error
	}

	count := int(countCmd.Val())
	remaining := limit - count
	if remaining < 0 {
		remaining = 0
	}

	return count <= int64ToInt(int64(limit)), remaining, nil
}

func int64ToInt(v int64) int {
	return int(v)
}
