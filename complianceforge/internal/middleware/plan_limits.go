package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// ============================================================
// PLAN LIMITS MIDDLEWARE
// Enforces subscription plan resource limits before allowing
// resource creation. Returns 402 Payment Required if the
// organisation has reached its plan limit for the resource.
//
// Uses interfaces to avoid circular imports with the service
// package. The service.SubscriptionService implements these
// interfaces.
// ============================================================

// LimitCheckResult represents the result of checking a plan limit.
type LimitCheckResult struct {
	Resource  string `json:"resource"`
	Current   int    `json:"current"`
	Max       int    `json:"max"`
	CanCreate bool   `json:"can_create"`
	Remaining int    `json:"remaining"`
}

// LimitChecker is implemented by SubscriptionService.CheckLimits.
type LimitChecker interface {
	CheckLimits(ctx context.Context, orgID uuid.UUID, resource string) (*LimitCheckResult, error)
}

// SubscriptionInfo holds subscription status for middleware checks.
type SubscriptionInfo struct {
	Status   string
	Features map[string]interface{}
}

// SubscriptionChecker is implemented by SubscriptionService for status checks.
type SubscriptionChecker interface {
	GetSubscriptionStatus(ctx context.Context, orgID uuid.UUID) (*SubscriptionInfo, error)
}

// PlanLimits returns middleware that checks subscription plan limits
// before allowing resource creation. It uses the LimitChecker interface
// to check whether the organisation can create more of the given resource.
//
// Usage:
//
//	r.With(middleware.PlanLimits(subSvc, "users")).Post("/users", handler.CreateUser)
//	r.With(middleware.PlanLimits(subSvc, "frameworks")).Post("/frameworks/adopt", handler.AdoptFramework)
//	r.With(middleware.PlanLimits(subSvc, "risks")).Post("/risks", handler.CreateRisk)
//	r.With(middleware.PlanLimits(subSvc, "vendors")).Post("/vendors", handler.CreateVendor)
func PlanLimits(checker LimitChecker, resource string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only enforce on creation requests (POST, PUT for adopt-style)
			if r.Method != http.MethodPost && r.Method != http.MethodPut {
				next.ServeHTTP(w, r)
				return
			}

			orgID := GetOrgID(r.Context())
			if orgID.String() == "00000000-0000-0000-0000-000000000000" {
				// No org context, skip limit check (e.g., public endpoints)
				next.ServeHTTP(w, r)
				return
			}

			check, err := checker.CheckLimits(r.Context(), orgID, resource)
			if err != nil {
				log.Warn().
					Err(err).
					Str("org_id", orgID.String()).
					Str("resource", resource).
					Msg("Failed to check plan limits, allowing request")
				// On error checking limits, allow the request through
				// rather than blocking legitimate operations
				next.ServeHTTP(w, r)
				return
			}

			if !check.CanCreate {
				log.Info().
					Str("org_id", orgID.String()).
					Str("resource", resource).
					Int("current", check.Current).
					Int("max", check.Max).
					Msg("Plan limit reached, blocking resource creation")

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusPaymentRequired)
				json.NewEncoder(w).Encode(PlanLimitResponse{
					Success: false,
					Error: PlanLimitError{
						Code:    "PLAN_LIMIT_REACHED",
						Message: "You have reached the maximum number of " + resource + " for your current plan. Please upgrade to continue.",
					},
					Limit: PlanLimitDetail{
						Resource:  resource,
						Current:   check.Current,
						Maximum:   check.Max,
						Remaining: check.Remaining,
					},
				})
				return
			}

			// Limit not reached, proceed with the request
			next.ServeHTTP(w, r)
		})
	}
}

// PlanLimitResponse is the response returned when a plan limit is reached.
type PlanLimitResponse struct {
	Success bool            `json:"success"`
	Error   PlanLimitError  `json:"error"`
	Limit   PlanLimitDetail `json:"limit"`
}

// PlanLimitError describes the limit violation.
type PlanLimitError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// PlanLimitDetail provides specifics about the limit.
type PlanLimitDetail struct {
	Resource  string `json:"resource"`
	Current   int    `json:"current"`
	Maximum   int    `json:"maximum"`
	Remaining int    `json:"remaining"`
}

// RequireActiveSubscription returns middleware that ensures the organisation
// has an active (non-cancelled, non-paused) subscription before allowing access.
func RequireActiveSubscription(checker SubscriptionChecker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			orgID := GetOrgID(r.Context())
			if orgID.String() == "00000000-0000-0000-0000-000000000000" {
				next.ServeHTTP(w, r)
				return
			}

			sub, err := checker.GetSubscriptionStatus(r.Context(), orgID)
			if err != nil {
				log.Warn().Err(err).Str("org_id", orgID.String()).Msg("No subscription found, allowing request")
				next.ServeHTTP(w, r)
				return
			}

			switch sub.Status {
			case "active", "trialing":
				// All good, proceed
				next.ServeHTTP(w, r)
			case "past_due":
				// Allow access but warn
				w.Header().Set("X-Subscription-Warning", "payment_past_due")
				next.ServeHTTP(w, r)
			case "paused":
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusPaymentRequired)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"error": map[string]string{
						"code":    "SUBSCRIPTION_PAUSED",
						"message": "Your subscription is paused. Please resume your subscription to continue using the platform.",
					},
				})
			case "cancelled":
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusPaymentRequired)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"error": map[string]string{
						"code":    "SUBSCRIPTION_CANCELLED",
						"message": "Your subscription has been cancelled. Please resubscribe to continue using the platform.",
					},
				})
			default:
				next.ServeHTTP(w, r)
			}
		})
	}
}

// RequireFeature returns middleware that checks whether a specific feature
// is enabled for the organisation's current subscription plan.
func RequireFeature(checker SubscriptionChecker, feature string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			orgID := GetOrgID(r.Context())
			if orgID.String() == "00000000-0000-0000-0000-000000000000" {
				next.ServeHTTP(w, r)
				return
			}

			sub, err := checker.GetSubscriptionStatus(r.Context(), orgID)
			if err != nil {
				// No subscription, allow through with default features
				next.ServeHTTP(w, r)
				return
			}

			// Check if feature is enabled in plan
			if sub.Features != nil {
				if enabled, ok := sub.Features[feature]; ok {
					if b, ok := enabled.(bool); ok && b {
						next.ServeHTTP(w, r)
						return
					}
				}
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusPaymentRequired)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error": map[string]string{
					"code":    "FEATURE_NOT_AVAILABLE",
					"message": "The " + feature + " feature is not included in your current plan. Please upgrade to access this feature.",
				},
			})
		})
	}
}
