package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

const (
	// ContextKeyAPIKeyID is stored in context when API key auth succeeds.
	ContextKeyAPIKeyID contextKey = "api_key_id"
	// ContextKeyAPIKeyPerms holds the permissions slice from the API key.
	ContextKeyAPIKeyPerms contextKey = "api_key_permissions"
)

// ValidatedAPIKey is the result returned by the APIKeyValidator interface.
// It mirrors the relevant fields of service.APIKey without creating a
// circular import between middleware and service.
type ValidatedAPIKey struct {
	ID              uuid.UUID
	OrganizationID  uuid.UUID
	Name            string
	KeyPrefix       string
	Permissions     []string
	RateLimitPerMin int
	ExpiresAt       *time.Time
	IsActive        bool
	CreatedBy       *uuid.UUID
}

// APIKeyValidator is the interface that the middleware uses to look up
// and validate API keys. It is implemented by service.IntegrationService.
type APIKeyValidator interface {
	ValidateAPIKey(ctx context.Context, rawKey string, clientIP string) (*ValidatedAPIKey, error)
}

// apiKeyRateTracker is a simple in-memory rate counter per key ID.
// Production should use Redis (like the existing RateLimit middleware).
type apiKeyRateTracker struct {
	counts map[string]*rateBucket
}

type rateBucket struct {
	Count   int
	ResetAt time.Time
}

var apiKeyRates = &apiKeyRateTracker{
	counts: make(map[string]*rateBucket),
}

func (t *apiKeyRateTracker) check(keyID string, limit int) (bool, int) {
	now := time.Now()
	bucket, ok := t.counts[keyID]
	if !ok || now.After(bucket.ResetAt) {
		t.counts[keyID] = &rateBucket{Count: 1, ResetAt: now.Add(time.Minute)}
		return true, limit - 1
	}
	bucket.Count++
	remaining := limit - bucket.Count
	if remaining < 0 {
		remaining = 0
	}
	return bucket.Count <= limit, remaining
}

// APIKeyAuth returns middleware that authenticates requests via X-API-Key header.
// It validates the key, checks permissions, enforces per-key rate limits,
// and sets org/user context so downstream handlers work as normal.
func APIKeyAuth(validator APIKeyValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get("X-API-Key")
			if apiKey == "" {
				// No API key — fall through to normal JWT auth.
				next.ServeHTTP(w, r)
				return
			}

			// Extract client IP
			clientIP := r.RemoteAddr
			if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
				parts := strings.SplitN(fwd, ",", 2)
				clientIP = strings.TrimSpace(parts[0])
			}

			// Validate the key
			key, err := validator.ValidateAPIKey(r.Context(), apiKey, clientIP)
			if err != nil {
				log.Warn().Err(err).Str("ip", clientIP).Msg("Invalid API key")
				writeAPIKeyError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired API key")
				return
			}

			// Rate limit per key
			allowed, remaining := apiKeyRates.check(key.ID.String(), key.RateLimitPerMin)
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(key.RateLimitPerMin))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))

			if !allowed {
				w.Header().Set("Retry-After", "60")
				log.Warn().Str("key_id", key.ID.String()).Msg("API key rate limited")
				writeAPIKeyError(w, http.StatusTooManyRequests, "RATE_LIMITED",
					"API key rate limit exceeded. Try again later.")
				return
			}

			// Check permissions against the endpoint
			if !checkAPIKeyPermission(key.Permissions, r.Method, r.URL.Path) {
				log.Warn().
					Str("key_id", key.ID.String()).
					Str("method", r.Method).
					Str("path", r.URL.Path).
					Msg("API key insufficient permissions")
				writeAPIKeyError(w, http.StatusForbidden, "FORBIDDEN",
					"API key does not have permission for this endpoint")
				return
			}

			// Set context values so downstream handlers see org_id
			ctx := context.WithValue(r.Context(), ContextKeyOrgID, key.OrganizationID)
			ctx = context.WithValue(ctx, ContextKeyAPIKeyID, key.ID)
			ctx = context.WithValue(ctx, ContextKeyAPIKeyPerms, key.Permissions)

			// If the key has a created_by, use that as the user ID
			if key.CreatedBy != nil {
				ctx = context.WithValue(ctx, ContextKeyUserID, *key.CreatedBy)
			} else {
				ctx = context.WithValue(ctx, ContextKeyUserID, uuid.Nil)
			}

			log.Debug().
				Str("key_id", key.ID.String()).
				Str("org_id", key.OrganizationID.String()).
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Msg("API key authenticated")

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// checkAPIKeyPermission verifies that the key's permissions cover the requested endpoint.
// Permissions use a colon-separated format: "resource:action"
// Examples: "integrations:read", "integrations:write", "risks:read", "*:read", "*:*"
func checkAPIKeyPermission(permissions []string, method, path string) bool {
	if len(permissions) == 0 {
		return false
	}

	// Determine required action from HTTP method
	var action string
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		action = "read"
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		action = "write"
	default:
		action = "read"
	}

	// Determine resource from path
	resource := extractResource(path)

	for _, perm := range permissions {
		parts := strings.SplitN(perm, ":", 2)
		if len(parts) != 2 {
			// If permission is just a resource name, assume full access
			if parts[0] == "*" || parts[0] == resource {
				return true
			}
			continue
		}

		permResource := parts[0]
		permAction := parts[1]

		// Wildcard checks
		if permResource == "*" && permAction == "*" {
			return true
		}
		if permResource == "*" && permAction == action {
			return true
		}
		if permResource == resource && permAction == "*" {
			return true
		}
		if permResource == resource && permAction == action {
			return true
		}
	}

	return false
}

// extractResource parses the API path to determine the top-level resource.
// e.g. /api/v1/risks/123 -> "risks", /api/v1/integrations/456/sync -> "integrations"
func extractResource(path string) string {
	// Strip the /api/v1/ prefix if present
	path = strings.TrimPrefix(path, "/api/v1/")
	path = strings.TrimPrefix(path, "/api/v1")
	path = strings.TrimPrefix(path, "/")

	// Take the first segment
	parts := strings.SplitN(path, "/", 2)
	if len(parts) > 0 && parts[0] != "" {
		return parts[0]
	}
	return "unknown"
}

// GetAPIKeyID returns the API key ID from context, if present.
func GetAPIKeyID(ctx context.Context) uuid.UUID {
	if id, ok := ctx.Value(ContextKeyAPIKeyID).(uuid.UUID); ok {
		return id
	}
	return uuid.Nil
}

// GetAPIKeyPermissions returns the API key permissions from context.
func GetAPIKeyPermissions(ctx context.Context) []string {
	if perms, ok := ctx.Value(ContextKeyAPIKeyPerms).([]string); ok {
		return perms
	}
	return nil
}

// writeAPIKeyError writes a JSON error response for API key auth failures.
func writeAPIKeyError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	resp := map[string]interface{}{
		"success": false,
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	}
	json.NewEncoder(w).Encode(resp)
}

// RequireAPIKeyPermission returns middleware that enforces a specific permission
// on requests authenticated via API key.
func RequireAPIKeyPermission(resource, action string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only check if this is an API key request
			keyID := GetAPIKeyID(r.Context())
			if keyID == uuid.Nil {
				// Not an API key request — let normal auth handle it
				next.ServeHTTP(w, r)
				return
			}

			perms := GetAPIKeyPermissions(r.Context())
			requiredPerm := fmt.Sprintf("%s:%s", resource, action)

			hasPermission := false
			for _, p := range perms {
				if p == "*:*" || p == requiredPerm || p == resource+":*" || p == "*:"+action {
					hasPermission = true
					break
				}
			}

			if !hasPermission {
				writeAPIKeyError(w, http.StatusForbidden, "FORBIDDEN",
					fmt.Sprintf("API key requires permission: %s", requiredPerm))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
