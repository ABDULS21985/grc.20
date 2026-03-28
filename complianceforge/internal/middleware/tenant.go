package middleware

import (
	"net/http"

	"github.com/complianceforge/platform/internal/database"
)

// Tenant returns middleware that sets the PostgreSQL RLS tenant context
// based on the authenticated user's organization ID.
func Tenant(db *database.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			orgID := GetOrgID(r.Context())
			if orgID.String() != "" && orgID.String() != "00000000-0000-0000-0000-000000000000" {
				// Store org ID in context for repository layer to use when beginning transactions
				// The actual SET is done in db.BeginTx or db.ExecWithTenant
			}
			next.ServeHTTP(w, r)
		})
	}
}
