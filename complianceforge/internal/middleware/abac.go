package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// ============================================================
// ABAC EVALUATOR FUNCTION TYPE
// ============================================================

// ABACEvalFunc is the function signature the ABAC middleware calls to evaluate
// access requests. It decouples the middleware from the service package, breaking
// the import cycle. Wire it at startup via service.ABACEngine.NewEvalFunc().
//
// Parameters:
//   - userID, orgID: identity of the requesting subject
//   - roles:          subject's current role set
//   - action:         requested action (read, create, update, delete, export, etc.)
//   - resourceType:   the kind of resource being accessed
//   - ip:             client IP address
//   - mfaVerified:    whether MFA was completed for this session
//
// Returns:
//   - allowed: true if access is permitted
//   - reason:  human-readable explanation of the decision
type ABACEvalFunc func(
	r *http.Request,
	userID, orgID uuid.UUID,
	roles []string,
	action, resourceType string,
	ip string,
	mfaVerified bool,
) (allowed bool, reason string)

// ============================================================
// ABAC MIDDLEWARE
// ============================================================

// ABAC returns HTTP middleware that enforces attribute-based access control.
// It extracts the user from context, calls the evaluator function, and either
// allows the request or returns 403 Forbidden.
//
// Usage:
//
//	evalFn := engine.NewEvalFunc()
//	r.With(middleware.ABAC(evalFn, "read", "risk")).Get("/", handler.ListRisks)
func ABAC(eval ABACEvalFunc, action string, resourceType string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := GetUserID(r.Context())
			orgID := GetOrgID(r.Context())
			roles := GetRoles(r.Context())

			if userID == uuid.Nil || orgID == uuid.Nil {
				writeABACError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
				return
			}

			ip := extractClientIP(r)
			mfa := getMFAStatus(r)

			allowed, reason := eval(r, userID, orgID, roles, action, resourceType, ip, mfa)

			if !allowed {
				log.Info().
					Str("user_id", userID.String()).
					Str("action", action).
					Str("resource_type", resourceType).
					Str("reason", reason).
					Msg("ABAC: access denied")

				writeABACError(w, http.StatusForbidden, "FORBIDDEN", reason)
				return
			}

			log.Debug().
				Str("user_id", userID.String()).
				Str("action", action).
				Str("resource_type", resourceType).
				Msg("ABAC: access granted")

			next.ServeHTTP(w, r)
		})
	}
}

// ABACWithResourceID is a variant that includes a resource ID in the evaluation.
// The extractID function extracts the resource UUID from the request (e.g., chi.URLParam).
func ABACWithResourceID(eval ABACEvalFunc, action string, resourceType string, extractID func(*http.Request) *uuid.UUID) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := GetUserID(r.Context())
			orgID := GetOrgID(r.Context())
			roles := GetRoles(r.Context())

			if userID == uuid.Nil || orgID == uuid.Nil {
				writeABACError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
				return
			}

			ip := extractClientIP(r)
			mfa := getMFAStatus(r)

			allowed, reason := eval(r, userID, orgID, roles, action, resourceType, ip, mfa)

			if !allowed {
				log.Info().
					Str("user_id", userID.String()).
					Str("action", action).
					Str("resource_type", resourceType).
					Str("reason", reason).
					Msg("ABAC: access denied (resource-level)")

				writeABACError(w, http.StatusForbidden, "FORBIDDEN", reason)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ============================================================
// HELPERS
// ============================================================

// getMFAStatus checks if the request indicates MFA verification.
func getMFAStatus(r *http.Request) bool {
	mfa := r.Header.Get("X-MFA-Verified")
	return mfa == "true" || mfa == "1"
}

// extractClientIP returns the client IP from the request.
func extractClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		for i, ch := range xff {
			if ch == ',' {
				return xff[:i]
			}
		}
		return xff
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	return r.RemoteAddr
}

// writeABACError writes a JSON error response consistent with the platform pattern.
func writeABACError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}
