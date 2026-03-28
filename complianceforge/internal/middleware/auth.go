package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/complianceforge/platform/internal/config"
)

type contextKey string

const (
	ContextKeyUserID   contextKey = "user_id"
	ContextKeyOrgID    contextKey = "organization_id"
	ContextKeyEmail    contextKey = "email"
	ContextKeyRoles    contextKey = "roles"
)

// Claims represents the JWT token claims.
type Claims struct {
	UserID         uuid.UUID `json:"user_id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	Email          string    `json:"email"`
	Roles          []string  `json:"roles"`
	jwt.RegisteredClaims
}

// Auth returns middleware that validates JWT tokens and sets user context.
func Auth(cfg config.JWTConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error":{"code":"UNAUTHORIZED","message":"Missing authorization header"}}`, http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				http.Error(w, `{"error":{"code":"UNAUTHORIZED","message":"Invalid authorization format"}}`, http.StatusUnauthorized)
				return
			}

			tokenStr := parts[1]
			claims := &Claims{}

			token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(cfg.Secret), nil
			})

			if err != nil || !token.Valid {
				log.Warn().Err(err).Str("ip", r.RemoteAddr).Msg("Invalid JWT token")
				http.Error(w, `{"error":{"code":"UNAUTHORIZED","message":"Invalid or expired token"}}`, http.StatusUnauthorized)
				return
			}

			// Set user context for downstream handlers
			ctx := context.WithValue(r.Context(), ContextKeyUserID, claims.UserID)
			ctx = context.WithValue(ctx, ContextKeyOrgID, claims.OrganizationID)
			ctx = context.WithValue(ctx, ContextKeyEmail, claims.Email)
			ctx = context.WithValue(ctx, ContextKeyRoles, claims.Roles)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserID extracts the user ID from request context.
func GetUserID(ctx context.Context) uuid.UUID {
	if id, ok := ctx.Value(ContextKeyUserID).(uuid.UUID); ok {
		return id
	}
	return uuid.Nil
}

// GetOrgID extracts the organization ID from request context.
func GetOrgID(ctx context.Context) uuid.UUID {
	if id, ok := ctx.Value(ContextKeyOrgID).(uuid.UUID); ok {
		return id
	}
	return uuid.Nil
}

// GetRoles extracts the user's roles from request context.
func GetRoles(ctx context.Context) []string {
	if roles, ok := ctx.Value(ContextKeyRoles).([]string); ok {
		return roles
	}
	return nil
}

// RequireRole returns middleware that enforces role-based access.
func RequireRole(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRoles := GetRoles(r.Context())
			for _, allowed := range allowedRoles {
				for _, userRole := range userRoles {
					if userRole == allowed || userRole == "org_admin" {
						next.ServeHTTP(w, r)
						return
					}
				}
			}
			http.Error(w, `{"error":{"code":"FORBIDDEN","message":"Insufficient permissions"}}`, http.StatusForbidden)
		})
	}
}
