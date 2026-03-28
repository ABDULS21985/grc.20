-- ============================================================
-- Migration 014 DOWN: Onboarding Wizard & Subscription Management
-- ============================================================

-- Drop triggers
DROP TRIGGER IF EXISTS trg_onboarding_progress_updated_at ON onboarding_progress;
DROP TRIGGER IF EXISTS trg_subscription_plans_updated_at ON subscription_plans;

-- Drop RLS policies
DROP POLICY IF EXISTS usage_events_tenant_select ON usage_events;
DROP POLICY IF EXISTS usage_events_tenant_insert ON usage_events;
DROP POLICY IF EXISTS usage_events_tenant_isolation ON usage_events;
DROP POLICY IF EXISTS onboarding_progress_tenant_delete ON onboarding_progress;
DROP POLICY IF EXISTS onboarding_progress_tenant_update ON onboarding_progress;
DROP POLICY IF EXISTS onboarding_progress_tenant_insert ON onboarding_progress;
DROP POLICY IF EXISTS onboarding_progress_tenant_isolation ON onboarding_progress;

-- Drop usage_events table
DROP TABLE IF EXISTS usage_events;

-- Drop onboarding_progress table
DROP TABLE IF EXISTS onboarding_progress;

-- Remove added columns from organization_subscriptions
ALTER TABLE organization_subscriptions
    DROP COLUMN IF EXISTS usage_snapshot,
    DROP COLUMN IF EXISTS stripe_subscription_id,
    DROP COLUMN IF EXISTS stripe_customer_id,
    DROP COLUMN IF EXISTS cancel_reason,
    DROP COLUMN IF EXISTS cancelled_at,
    DROP COLUMN IF EXISTS trial_ends_at,
    DROP COLUMN IF EXISTS plan_id;

-- Restore original index on organization_subscriptions
DROP INDEX IF EXISTS idx_org_subscriptions_period_end;
DROP INDEX IF EXISTS idx_org_subscriptions_stripe;
DROP INDEX IF EXISTS idx_org_subscriptions_status;
DROP INDEX IF EXISTS idx_org_subscriptions_plan;
DROP INDEX IF EXISTS idx_org_subscriptions_org_unique;
CREATE INDEX IF NOT EXISTS idx_org_subscriptions_org ON organization_subscriptions(organization_id);

-- Drop subscription_plans table
DROP TABLE IF EXISTS subscription_plans;

-- Drop enums
DROP TYPE IF EXISTS billing_cycle_type;
DROP TYPE IF EXISTS subscription_status;
DROP TYPE IF EXISTS subscription_tier;
