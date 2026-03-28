package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/complianceforge/platform/internal/pkg/cache"
)

// RateLimit returns middleware that enforces per-user or per-IP rate limiting via Redis.
func RateLimit(cacheClient *cache.Client, maxRequests int, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Use user ID if authenticated, otherwise IP
			key := r.RemoteAddr
			if userID := GetUserID(r.Context()); userID.String() != "00000000-0000-0000-0000-000000000000" {
				key = userID.String()
			}
			rateLimitKey := fmt.Sprintf("ratelimit:%s", key)

			allowed, remaining, err := cacheClient.CheckRateLimit(r.Context(), rateLimitKey, maxRequests, window)
			if err != nil {
				// On Redis error, allow the request through
				next.ServeHTTP(w, r)
				return
			}

			// Set rate limit headers
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(maxRequests))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(window).Unix(), 10))

			if !allowed {
				w.Header().Set("Retry-After", strconv.Itoa(int(window.Seconds())))
				http.Error(w, `{"error":{"code":"RATE_LIMITED","message":"Too many requests. Please try again later."}}`, http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
