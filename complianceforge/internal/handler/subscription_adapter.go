package handler

import (
	"context"

	"github.com/google/uuid"

	"github.com/complianceforge/platform/internal/middleware"
	"github.com/complianceforge/platform/internal/service"
)

// ============================================================
// SUBSCRIPTION MIDDLEWARE ADAPTERS
// These adapters bridge the service.SubscriptionService to the
// middleware interfaces (LimitChecker, SubscriptionChecker),
// avoiding circular imports between service and middleware.
// ============================================================

// SubscriptionLimitAdapter adapts service.SubscriptionService to
// satisfy the middleware.LimitChecker interface.
type SubscriptionLimitAdapter struct {
	svc *service.SubscriptionService
}

// NewSubscriptionLimitAdapter creates a new adapter.
func NewSubscriptionLimitAdapter(svc *service.SubscriptionService) *SubscriptionLimitAdapter {
	return &SubscriptionLimitAdapter{svc: svc}
}

// CheckLimits satisfies middleware.LimitChecker by delegating to
// service.SubscriptionService.CheckLimits and converting the result.
func (a *SubscriptionLimitAdapter) CheckLimits(ctx context.Context, orgID uuid.UUID, resource string) (*middleware.LimitCheckResult, error) {
	check, err := a.svc.CheckLimits(ctx, orgID, resource)
	if err != nil {
		return nil, err
	}
	return &middleware.LimitCheckResult{
		Resource:  check.Resource,
		Current:   check.Current,
		Max:       check.Max,
		CanCreate: check.CanCreate,
		Remaining: check.Remaining,
	}, nil
}

// SubscriptionStatusAdapter adapts service.SubscriptionService to
// satisfy the middleware.SubscriptionChecker interface.
type SubscriptionStatusAdapter struct {
	svc *service.SubscriptionService
}

// NewSubscriptionStatusAdapter creates a new adapter.
func NewSubscriptionStatusAdapter(svc *service.SubscriptionService) *SubscriptionStatusAdapter {
	return &SubscriptionStatusAdapter{svc: svc}
}

// GetSubscriptionStatus satisfies middleware.SubscriptionChecker by
// delegating to service.SubscriptionService.GetSubscriptionStatus.
func (a *SubscriptionStatusAdapter) GetSubscriptionStatus(ctx context.Context, orgID uuid.UUID) (*middleware.SubscriptionInfo, error) {
	status, features, err := a.svc.GetSubscriptionStatus(ctx, orgID)
	if err != nil {
		return nil, err
	}
	return &middleware.SubscriptionInfo{
		Status:   status,
		Features: features,
	}, nil
}
