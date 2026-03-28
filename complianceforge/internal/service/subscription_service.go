package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// ============================================================
// SUBSCRIPTION SERVICE
// Manages subscription lifecycle, plan changes, usage limits,
// and billing for the ComplianceForge GRC platform.
// ============================================================

// SubscriptionService manages organisation subscriptions and plan limits.
type SubscriptionService struct {
	pool *pgxpool.Pool
}

// NewSubscriptionService creates a new SubscriptionService.
func NewSubscriptionService(pool *pgxpool.Pool) *SubscriptionService {
	return &SubscriptionService{pool: pool}
}

// ============================================================
// DATA TYPES
// ============================================================

// SubscriptionPlan represents an available subscription tier.
type SubscriptionPlan struct {
	ID               uuid.UUID              `json:"id"`
	Name             string                 `json:"name"`
	Slug             string                 `json:"slug"`
	Description      string                 `json:"description"`
	Tier             string                 `json:"tier"`
	PricingMonthly   float64                `json:"pricing_eur_monthly"`
	PricingAnnual    float64                `json:"pricing_eur_annual"`
	MaxUsers         int                    `json:"max_users"`
	MaxFrameworks    int                    `json:"max_frameworks"`
	MaxRisks         int                    `json:"max_risks"`
	MaxVendors       int                    `json:"max_vendors"`
	MaxStorageGB     int                    `json:"max_storage_gb"`
	Features         map[string]interface{} `json:"features"`
	IsActive         bool                   `json:"is_active"`
	SortOrder        int                    `json:"sort_order"`
	MonthlySavings   float64                `json:"monthly_savings,omitempty"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

// Subscription represents an organisation's current subscription.
type Subscription struct {
	ID                   uuid.UUID              `json:"id"`
	OrganizationID       uuid.UUID              `json:"organization_id"`
	PlanID               *uuid.UUID             `json:"plan_id,omitempty"`
	PlanName             string                 `json:"plan_name"`
	Status               string                 `json:"status"`
	BillingCycle         string                 `json:"billing_cycle"`
	CurrentPeriodStart   *time.Time             `json:"current_period_start,omitempty"`
	CurrentPeriodEnd     *time.Time             `json:"current_period_end,omitempty"`
	TrialEndsAt          *time.Time             `json:"trial_ends_at,omitempty"`
	CancelledAt          *time.Time             `json:"cancelled_at,omitempty"`
	CancelReason         string                 `json:"cancel_reason,omitempty"`
	StripeCustomerID     string                 `json:"stripe_customer_id,omitempty"`
	StripeSubscriptionID string                 `json:"stripe_subscription_id,omitempty"`
	MaxUsers             int                    `json:"max_users"`
	MaxFrameworks        int                    `json:"max_frameworks"`
	UsageSnapshot        map[string]interface{} `json:"usage_snapshot,omitempty"`
	Plan                 *SubscriptionPlan      `json:"plan,omitempty"`
	CreatedAt            time.Time              `json:"created_at"`
	UpdatedAt            time.Time              `json:"updated_at"`
}

// LimitCheck represents the result of checking a plan limit.
type LimitCheck struct {
	Resource  string `json:"resource"`
	Current   int    `json:"current"`
	Max       int    `json:"max"`
	CanCreate bool   `json:"can_create"`
	Remaining int    `json:"remaining"`
}

// UsageSummary represents aggregated usage data for an organisation.
type UsageSummary struct {
	OrganizationID uuid.UUID       `json:"organization_id"`
	Users          ResourceUsage   `json:"users"`
	Frameworks     ResourceUsage   `json:"frameworks"`
	Risks          ResourceUsage   `json:"risks"`
	Vendors        ResourceUsage   `json:"vendors"`
	StorageGB      ResourceUsageGB `json:"storage_gb"`
	RecentEvents   []UsageEvent    `json:"recent_events"`
	PeriodStart    *time.Time      `json:"period_start,omitempty"`
	PeriodEnd      *time.Time      `json:"period_end,omitempty"`
}

// ResourceUsage represents usage for a countable resource.
type ResourceUsage struct {
	Current   int  `json:"current"`
	Max       int  `json:"max"`
	Remaining int  `json:"remaining"`
	AtLimit   bool `json:"at_limit"`
}

// ResourceUsageGB represents storage usage.
type ResourceUsageGB struct {
	CurrentGB float64 `json:"current_gb"`
	MaxGB     int     `json:"max_gb"`
	UsedPct   float64 `json:"used_pct"`
	AtLimit   bool    `json:"at_limit"`
}

// UsageEvent represents a single usage event.
type UsageEvent struct {
	ID             uuid.UUID              `json:"id"`
	OrganizationID uuid.UUID              `json:"organization_id"`
	EventType      string                 `json:"event_type"`
	Quantity       int                    `json:"quantity"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
}

// ============================================================
// LIST PLANS
// ============================================================

// ListPlans returns all active subscription plans ordered by sort_order.
func (s *SubscriptionService) ListPlans(ctx context.Context) ([]SubscriptionPlan, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, slug, description, tier,
		       pricing_eur_monthly, pricing_eur_annual,
		       max_users, max_frameworks, max_risks, max_vendors, max_storage_gb,
		       features, is_active, sort_order, created_at, updated_at
		FROM subscription_plans
		WHERE is_active = true
		ORDER BY sort_order ASC`)
	if err != nil {
		return nil, fmt.Errorf("failed to list plans: %w", err)
	}
	defer rows.Close()

	var plans []SubscriptionPlan
	for rows.Next() {
		var p SubscriptionPlan
		var featuresJSON []byte
		err := rows.Scan(
			&p.ID, &p.Name, &p.Slug, &p.Description, &p.Tier,
			&p.PricingMonthly, &p.PricingAnnual,
			&p.MaxUsers, &p.MaxFrameworks, &p.MaxRisks, &p.MaxVendors, &p.MaxStorageGB,
			&featuresJSON, &p.IsActive, &p.SortOrder, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan plan: %w", err)
		}
		if len(featuresJSON) > 0 {
			json.Unmarshal(featuresJSON, &p.Features)
		}
		// Calculate monthly savings for annual billing
		if p.PricingMonthly > 0 {
			annualMonthly := p.PricingAnnual / 12
			p.MonthlySavings = p.PricingMonthly - annualMonthly
		}
		plans = append(plans, p)
	}

	return plans, nil
}

// ============================================================
// GET SUBSCRIPTION
// ============================================================

// GetSubscription returns the current subscription for an organisation.
func (s *SubscriptionService) GetSubscription(ctx context.Context, orgID uuid.UUID) (*Subscription, error) {
	sub := &Subscription{}
	var usageJSON []byte

	err := s.pool.QueryRow(ctx, `
		SELECT os.id, os.organization_id, os.plan_id, os.plan_name, os.status,
		       os.billing_cycle, os.current_period_start, os.current_period_end,
		       os.trial_ends_at, os.cancelled_at, os.cancel_reason,
		       os.stripe_customer_id, os.stripe_subscription_id,
		       os.max_users, os.max_frameworks, os.usage_snapshot,
		       os.created_at, os.updated_at
		FROM organization_subscriptions os
		WHERE os.organization_id = $1`,
		orgID,
	).Scan(
		&sub.ID, &sub.OrganizationID, &sub.PlanID, &sub.PlanName, &sub.Status,
		&sub.BillingCycle, &sub.CurrentPeriodStart, &sub.CurrentPeriodEnd,
		&sub.TrialEndsAt, &sub.CancelledAt, &sub.CancelReason,
		&sub.StripeCustomerID, &sub.StripeSubscriptionID,
		&sub.MaxUsers, &sub.MaxFrameworks, &usageJSON,
		&sub.CreatedAt, &sub.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("no subscription found for organisation")
		}
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	if len(usageJSON) > 0 {
		json.Unmarshal(usageJSON, &sub.UsageSnapshot)
	}

	// Load plan details if plan_id is set
	if sub.PlanID != nil {
		plan, err := s.getPlanByID(ctx, *sub.PlanID)
		if err == nil {
			sub.Plan = plan
		}
	}

	return sub, nil
}

// ============================================================
// CREATE SUBSCRIPTION
// ============================================================

// CreateSubscription creates a new subscription for an organisation.
func (s *SubscriptionService) CreateSubscription(ctx context.Context, orgID uuid.UUID, planSlug, billingCycle string) (*Subscription, error) {
	log.Info().Str("org_id", orgID.String()).Str("plan", planSlug).Str("cycle", billingCycle).Msg("Creating subscription")

	// Validate billing cycle
	if billingCycle != "monthly" && billingCycle != "annual" {
		return nil, fmt.Errorf("invalid billing cycle: %s (must be monthly or annual)", billingCycle)
	}

	// Look up the plan
	plan, err := s.getPlanBySlug(ctx, planSlug)
	if err != nil {
		return nil, fmt.Errorf("plan not found: %w", err)
	}

	now := time.Now()
	var periodEnd time.Time
	if billingCycle == "annual" {
		periodEnd = now.AddDate(1, 0, 0)
	} else {
		periodEnd = now.AddDate(0, 1, 0)
	}

	// Trial period: 14 days
	trialEnd := now.AddDate(0, 0, 14)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Check for existing subscription
	var existingCount int
	err = tx.QueryRow(ctx, `
		SELECT COUNT(*) FROM organization_subscriptions WHERE organization_id = $1`, orgID,
	).Scan(&existingCount)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing subscription: %w", err)
	}

	var subID uuid.UUID
	if existingCount > 0 {
		// Update existing subscription
		err = tx.QueryRow(ctx, `
			UPDATE organization_subscriptions
			SET plan_id = $1, plan_name = $2, status = 'trialing',
			    billing_cycle = $3, max_users = $4, max_frameworks = $5,
			    current_period_start = $6, current_period_end = $7,
			    trial_ends_at = $8, features_enabled = $9,
			    updated_at = NOW()
			WHERE organization_id = $10
			RETURNING id`,
			plan.ID, plan.Slug, billingCycle, plan.MaxUsers, plan.MaxFrameworks,
			now, periodEnd, trialEnd, plan.Features, orgID,
		).Scan(&subID)
	} else {
		// Create new subscription
		err = tx.QueryRow(ctx, `
			INSERT INTO organization_subscriptions (
				organization_id, plan_id, plan_name, status, billing_cycle,
				max_users, max_frameworks, features_enabled,
				current_period_start, current_period_end, trial_ends_at
			) VALUES ($1, $2, $3, 'trialing', $4, $5, $6, $7, $8, $9, $10)
			RETURNING id`,
			orgID, plan.ID, plan.Slug, billingCycle,
			plan.MaxUsers, plan.MaxFrameworks, plan.Features,
			now, periodEnd, trialEnd,
		).Scan(&subID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create/update subscription: %w", err)
	}

	// Update org tier
	_, err = tx.Exec(ctx, `
		UPDATE organizations SET tier = $1, updated_at = NOW() WHERE id = $2`,
		plan.Tier, orgID,
	)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to update organization tier")
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit subscription creation: %w", err)
	}

	log.Info().
		Str("org_id", orgID.String()).
		Str("plan", planSlug).
		Str("sub_id", subID.String()).
		Msg("Subscription created successfully")

	return s.GetSubscription(ctx, orgID)
}

// ============================================================
// CHANGE PLAN
// ============================================================

// ChangePlan changes an organisation's subscription plan.
func (s *SubscriptionService) ChangePlan(ctx context.Context, orgID uuid.UUID, newPlanSlug string) (*Subscription, error) {
	log.Info().Str("org_id", orgID.String()).Str("new_plan", newPlanSlug).Msg("Changing subscription plan")

	// Look up the new plan
	newPlan, err := s.getPlanBySlug(ctx, newPlanSlug)
	if err != nil {
		return nil, fmt.Errorf("new plan not found: %w", err)
	}

	// Get current subscription
	currentSub, err := s.GetSubscription(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("no current subscription found: %w", err)
	}

	if currentSub.Status == "cancelled" {
		return nil, fmt.Errorf("cannot change plan on a cancelled subscription")
	}

	if currentSub.PlanName == newPlanSlug {
		return nil, fmt.Errorf("already on the %s plan", newPlanSlug)
	}

	// Validate downgrade feasibility: check current usage against new limits
	usageCheck, err := s.validateDowngrade(ctx, orgID, newPlan)
	if err != nil {
		return nil, err
	}
	if usageCheck != "" {
		return nil, fmt.Errorf("cannot downgrade: %s", usageCheck)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		UPDATE organization_subscriptions
		SET plan_id = $1, plan_name = $2,
		    max_users = $3, max_frameworks = $4,
		    features_enabled = $5,
		    updated_at = NOW()
		WHERE organization_id = $6`,
		newPlan.ID, newPlan.Slug,
		newPlan.MaxUsers, newPlan.MaxFrameworks,
		newPlan.Features, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update subscription plan: %w", err)
	}

	// Update org tier
	_, err = tx.Exec(ctx, `
		UPDATE organizations SET tier = $1, updated_at = NOW() WHERE id = $2`,
		newPlan.Tier, orgID,
	)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to update organization tier")
	}

	// Record usage event
	_, err = tx.Exec(ctx, `
		INSERT INTO usage_events (organization_id, event_type, quantity, metadata)
		VALUES ($1, 'plan_change', 1, $2)`,
		orgID, fmt.Sprintf(`{"from":"%s","to":"%s"}`, currentSub.PlanName, newPlanSlug),
	)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to record plan change event")
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit plan change: %w", err)
	}

	log.Info().
		Str("org_id", orgID.String()).
		Str("from", currentSub.PlanName).
		Str("to", newPlanSlug).
		Msg("Plan changed successfully")

	return s.GetSubscription(ctx, orgID)
}

// ============================================================
// CANCEL SUBSCRIPTION
// ============================================================

// CancelSubscription cancels an organisation's subscription.
func (s *SubscriptionService) CancelSubscription(ctx context.Context, orgID uuid.UUID, reason string) (*Subscription, error) {
	log.Info().Str("org_id", orgID.String()).Str("reason", reason).Msg("Cancelling subscription")

	now := time.Now()
	_, err := s.pool.Exec(ctx, `
		UPDATE organization_subscriptions
		SET status = 'cancelled',
		    cancelled_at = $1,
		    cancel_reason = $2,
		    updated_at = NOW()
		WHERE organization_id = $3 AND status NOT IN ('cancelled')`,
		now, reason, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel subscription: %w", err)
	}

	// Record event
	s.RecordUsage(ctx, orgID, "subscription_cancelled", 1)

	return s.GetSubscription(ctx, orgID)
}

// ============================================================
// PAUSE / RESUME SUBSCRIPTION
// ============================================================

// PauseSubscription pauses an active subscription.
func (s *SubscriptionService) PauseSubscription(ctx context.Context, orgID uuid.UUID) (*Subscription, error) {
	log.Info().Str("org_id", orgID.String()).Msg("Pausing subscription")

	result, err := s.pool.Exec(ctx, `
		UPDATE organization_subscriptions
		SET status = 'paused', updated_at = NOW()
		WHERE organization_id = $1 AND status = 'active'`,
		orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to pause subscription: %w", err)
	}
	if result.RowsAffected() == 0 {
		return nil, fmt.Errorf("subscription is not active, cannot pause")
	}

	s.RecordUsage(ctx, orgID, "subscription_paused", 1)
	return s.GetSubscription(ctx, orgID)
}

// ResumeSubscription resumes a paused subscription.
func (s *SubscriptionService) ResumeSubscription(ctx context.Context, orgID uuid.UUID) (*Subscription, error) {
	log.Info().Str("org_id", orgID.String()).Msg("Resuming subscription")

	result, err := s.pool.Exec(ctx, `
		UPDATE organization_subscriptions
		SET status = 'active', updated_at = NOW()
		WHERE organization_id = $1 AND status = 'paused'`,
		orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to resume subscription: %w", err)
	}
	if result.RowsAffected() == 0 {
		return nil, fmt.Errorf("subscription is not paused, cannot resume")
	}

	s.RecordUsage(ctx, orgID, "subscription_resumed", 1)
	return s.GetSubscription(ctx, orgID)
}

// ============================================================
// PLAN LIMITS
// ============================================================

// CheckLimits checks whether an organisation can create more of a given resource.
func (s *SubscriptionService) CheckLimits(ctx context.Context, orgID uuid.UUID, resource string) (*LimitCheck, error) {
	// Get current subscription limits
	sub, err := s.GetSubscription(ctx, orgID)
	if err != nil {
		// If no subscription, use starter defaults
		sub = &Subscription{
			MaxUsers:      5,
			MaxFrameworks: 3,
		}
	}

	// Get plan-level limits if available
	maxRisks := 100
	maxVendors := 20
	if sub.Plan != nil {
		maxRisks = sub.Plan.MaxRisks
		maxVendors = sub.Plan.MaxVendors
	}

	check := &LimitCheck{Resource: resource}

	switch resource {
	case "users":
		var count int
		err := s.pool.QueryRow(ctx, `
			SELECT COUNT(*) FROM users
			WHERE organization_id = $1 AND status != 'inactive'`, orgID,
		).Scan(&count)
		if err != nil {
			return nil, fmt.Errorf("failed to count users: %w", err)
		}
		check.Current = count
		check.Max = sub.MaxUsers

	case "frameworks":
		var count int
		err := s.pool.QueryRow(ctx, `
			SELECT COUNT(*) FROM organization_frameworks
			WHERE organization_id = $1`, orgID,
		).Scan(&count)
		if err != nil {
			return nil, fmt.Errorf("failed to count frameworks: %w", err)
		}
		check.Current = count
		check.Max = sub.MaxFrameworks

	case "risks":
		var count int
		err := s.pool.QueryRow(ctx, `
			SELECT COUNT(*) FROM risks
			WHERE organization_id = $1 AND status != 'closed'`, orgID,
		).Scan(&count)
		if err != nil {
			return nil, fmt.Errorf("failed to count risks: %w", err)
		}
		check.Current = count
		check.Max = maxRisks

	case "vendors":
		var count int
		err := s.pool.QueryRow(ctx, `
			SELECT COUNT(*) FROM vendors
			WHERE organization_id = $1 AND status != 'offboarded'`, orgID,
		).Scan(&count)
		if err != nil {
			return nil, fmt.Errorf("failed to count vendors: %w", err)
		}
		check.Current = count
		check.Max = maxVendors

	default:
		return nil, fmt.Errorf("unknown resource type: %s", resource)
	}

	check.Remaining = check.Max - check.Current
	if check.Remaining < 0 {
		check.Remaining = 0
	}
	check.CanCreate = check.Current < check.Max

	return check, nil
}

// ============================================================
// USAGE TRACKING
// ============================================================

// RecordUsage records a usage event for an organisation.
func (s *SubscriptionService) RecordUsage(ctx context.Context, orgID uuid.UUID, eventType string, quantity int) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO usage_events (organization_id, event_type, quantity)
		VALUES ($1, $2, $3)`,
		orgID, eventType, quantity,
	)
	if err != nil {
		log.Warn().Err(err).Str("org_id", orgID.String()).Str("event", eventType).Msg("Failed to record usage event")
		return fmt.Errorf("failed to record usage: %w", err)
	}
	return nil
}

// RecordUsageWithMetadata records a usage event with additional metadata.
func (s *SubscriptionService) RecordUsageWithMetadata(ctx context.Context, orgID uuid.UUID, eventType string, quantity int, metadata map[string]interface{}) error {
	metaJSON, err := json.Marshal(metadata)
	if err != nil {
		metaJSON = []byte("{}")
	}

	_, err = s.pool.Exec(ctx, `
		INSERT INTO usage_events (organization_id, event_type, quantity, metadata)
		VALUES ($1, $2, $3, $4)`,
		orgID, eventType, quantity, metaJSON,
	)
	if err != nil {
		log.Warn().Err(err).Str("org_id", orgID.String()).Str("event", eventType).Msg("Failed to record usage event with metadata")
		return fmt.Errorf("failed to record usage: %w", err)
	}
	return nil
}

// GetUsageSummary returns a comprehensive usage summary for an organisation.
func (s *SubscriptionService) GetUsageSummary(ctx context.Context, orgID uuid.UUID) (*UsageSummary, error) {
	summary := &UsageSummary{OrganizationID: orgID}

	// Get subscription for limits
	sub, err := s.GetSubscription(ctx, orgID)
	if err != nil {
		sub = &Subscription{MaxUsers: 5, MaxFrameworks: 3}
	}

	maxRisks := 100
	maxVendors := 20
	maxStorageGB := 5
	if sub.Plan != nil {
		maxRisks = sub.Plan.MaxRisks
		maxVendors = sub.Plan.MaxVendors
		maxStorageGB = sub.Plan.MaxStorageGB
	}

	summary.PeriodStart = sub.CurrentPeriodStart
	summary.PeriodEnd = sub.CurrentPeriodEnd

	// Count users
	var userCount int
	s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM users WHERE organization_id = $1 AND status != 'inactive'`, orgID,
	).Scan(&userCount)
	summary.Users = ResourceUsage{
		Current:   userCount,
		Max:       sub.MaxUsers,
		Remaining: max(0, sub.MaxUsers-userCount),
		AtLimit:   userCount >= sub.MaxUsers,
	}

	// Count frameworks
	var fwCount int
	s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM organization_frameworks WHERE organization_id = $1`, orgID,
	).Scan(&fwCount)
	summary.Frameworks = ResourceUsage{
		Current:   fwCount,
		Max:       sub.MaxFrameworks,
		Remaining: max(0, sub.MaxFrameworks-fwCount),
		AtLimit:   fwCount >= sub.MaxFrameworks,
	}

	// Count risks
	var riskCount int
	s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM risks WHERE organization_id = $1 AND status != 'closed'`, orgID,
	).Scan(&riskCount)
	summary.Risks = ResourceUsage{
		Current:   riskCount,
		Max:       maxRisks,
		Remaining: max(0, maxRisks-riskCount),
		AtLimit:   riskCount >= maxRisks,
	}

	// Count vendors
	var vendorCount int
	s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM vendors WHERE organization_id = $1 AND status != 'offboarded'`, orgID,
	).Scan(&vendorCount)
	summary.Vendors = ResourceUsage{
		Current:   vendorCount,
		Max:       maxVendors,
		Remaining: max(0, maxVendors-vendorCount),
		AtLimit:   vendorCount >= maxVendors,
	}

	// Storage (estimated from evidence/documents)
	summary.StorageGB = ResourceUsageGB{
		CurrentGB: 0,
		MaxGB:     maxStorageGB,
		UsedPct:   0,
		AtLimit:   false,
	}

	// Recent usage events
	rows, err := s.pool.Query(ctx, `
		SELECT id, organization_id, event_type, quantity, metadata, created_at
		FROM usage_events
		WHERE organization_id = $1
		ORDER BY created_at DESC
		LIMIT 20`,
		orgID,
	)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var evt UsageEvent
			var metaJSON []byte
			if err := rows.Scan(&evt.ID, &evt.OrganizationID, &evt.EventType, &evt.Quantity, &metaJSON, &evt.CreatedAt); err != nil {
				continue
			}
			if len(metaJSON) > 0 {
				json.Unmarshal(metaJSON, &evt.Metadata)
			}
			summary.RecentEvents = append(summary.RecentEvents, evt)
		}
	}

	if summary.RecentEvents == nil {
		summary.RecentEvents = []UsageEvent{}
	}

	return summary, nil
}

// ============================================================
// INTERNAL HELPERS
// ============================================================

// getPlanBySlug fetches a subscription plan by its slug.
func (s *SubscriptionService) getPlanBySlug(ctx context.Context, slug string) (*SubscriptionPlan, error) {
	p := &SubscriptionPlan{}
	var featuresJSON []byte

	err := s.pool.QueryRow(ctx, `
		SELECT id, name, slug, description, tier,
		       pricing_eur_monthly, pricing_eur_annual,
		       max_users, max_frameworks, max_risks, max_vendors, max_storage_gb,
		       features, is_active, sort_order, created_at, updated_at
		FROM subscription_plans
		WHERE slug = $1 AND is_active = true`, slug,
	).Scan(
		&p.ID, &p.Name, &p.Slug, &p.Description, &p.Tier,
		&p.PricingMonthly, &p.PricingAnnual,
		&p.MaxUsers, &p.MaxFrameworks, &p.MaxRisks, &p.MaxVendors, &p.MaxStorageGB,
		&featuresJSON, &p.IsActive, &p.SortOrder, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("plan %q not found", slug)
		}
		return nil, fmt.Errorf("failed to get plan: %w", err)
	}
	if len(featuresJSON) > 0 {
		json.Unmarshal(featuresJSON, &p.Features)
	}
	return p, nil
}

// getPlanByID fetches a subscription plan by its ID.
func (s *SubscriptionService) getPlanByID(ctx context.Context, id uuid.UUID) (*SubscriptionPlan, error) {
	p := &SubscriptionPlan{}
	var featuresJSON []byte

	err := s.pool.QueryRow(ctx, `
		SELECT id, name, slug, description, tier,
		       pricing_eur_monthly, pricing_eur_annual,
		       max_users, max_frameworks, max_risks, max_vendors, max_storage_gb,
		       features, is_active, sort_order, created_at, updated_at
		FROM subscription_plans
		WHERE id = $1`, id,
	).Scan(
		&p.ID, &p.Name, &p.Slug, &p.Description, &p.Tier,
		&p.PricingMonthly, &p.PricingAnnual,
		&p.MaxUsers, &p.MaxFrameworks, &p.MaxRisks, &p.MaxVendors, &p.MaxStorageGB,
		&featuresJSON, &p.IsActive, &p.SortOrder, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get plan by ID: %w", err)
	}
	if len(featuresJSON) > 0 {
		json.Unmarshal(featuresJSON, &p.Features)
	}
	return p, nil
}

// validateDowngrade checks if a downgrade to a smaller plan is feasible.
func (s *SubscriptionService) validateDowngrade(ctx context.Context, orgID uuid.UUID, newPlan *SubscriptionPlan) (string, error) {
	// Check user count
	var userCount int
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM users
		WHERE organization_id = $1 AND status != 'inactive'`, orgID,
	).Scan(&userCount)
	if err != nil {
		return "", fmt.Errorf("failed to count users: %w", err)
	}
	if userCount > newPlan.MaxUsers {
		return fmt.Sprintf("current user count (%d) exceeds new plan limit (%d)", userCount, newPlan.MaxUsers), nil
	}

	// Check framework count
	var fwCount int
	err = s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM organization_frameworks
		WHERE organization_id = $1`, orgID,
	).Scan(&fwCount)
	if err != nil {
		return "", fmt.Errorf("failed to count frameworks: %w", err)
	}
	if fwCount > newPlan.MaxFrameworks {
		return fmt.Sprintf("current framework count (%d) exceeds new plan limit (%d)", fwCount, newPlan.MaxFrameworks), nil
	}

	// Check risk count
	var riskCount int
	err = s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM risks
		WHERE organization_id = $1 AND status != 'closed'`, orgID,
	).Scan(&riskCount)
	if err != nil {
		return "", fmt.Errorf("failed to count risks: %w", err)
	}
	if riskCount > newPlan.MaxRisks {
		return fmt.Sprintf("current risk count (%d) exceeds new plan limit (%d)", riskCount, newPlan.MaxRisks), nil
	}

	// Check vendor count
	var vendorCount int
	err = s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM vendors
		WHERE organization_id = $1 AND status != 'offboarded'`, orgID,
	).Scan(&vendorCount)
	if err != nil {
		return "", fmt.Errorf("failed to count vendors: %w", err)
	}
	if vendorCount > newPlan.MaxVendors {
		return fmt.Sprintf("current vendor count (%d) exceeds new plan limit (%d)", vendorCount, newPlan.MaxVendors), nil
	}

	return "", nil // OK to downgrade
}

// ============================================================
// MIDDLEWARE INTERFACE ADAPTERS
// These methods allow SubscriptionService to be wrapped by an
// adapter in the wiring layer so it satisfies the interfaces
// defined in the middleware package (LimitChecker,
// SubscriptionChecker). The service itself does not import
// middleware to avoid import cycles.
// ============================================================

// GetSubscriptionStatus returns the subscription status and features for an org.
// This is used by the adapter layer to satisfy middleware.SubscriptionChecker.
func (s *SubscriptionService) GetSubscriptionStatus(ctx context.Context, orgID uuid.UUID) (status string, features map[string]interface{}, err error) {
	sub, err := s.GetSubscription(ctx, orgID)
	if err != nil {
		return "", nil, err
	}

	feat := make(map[string]interface{})
	if sub.Plan != nil && sub.Plan.Features != nil {
		feat = sub.Plan.Features
	}

	return sub.Status, feat, nil
}

// max returns the larger of two ints. Needed for Go < 1.21 compat.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
