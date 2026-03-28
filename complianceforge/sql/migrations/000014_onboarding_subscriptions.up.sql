-- ============================================================
-- Migration 014: Onboarding Wizard & Subscription Management
-- Adds subscription plans, enhanced organization subscriptions,
-- onboarding progress tracking, and usage event recording.
-- ============================================================

-- ============================================================
-- ENUMS
-- ============================================================

CREATE TYPE subscription_tier AS ENUM (
    'starter',
    'professional',
    'enterprise',
    'unlimited'
);

CREATE TYPE subscription_status AS ENUM (
    'trialing',
    'active',
    'past_due',
    'cancelled',
    'paused'
);

CREATE TYPE billing_cycle_type AS ENUM (
    'monthly',
    'annual'
);

-- ============================================================
-- TABLE: subscription_plans
-- Defines available subscription tiers and their limits.
-- ============================================================

CREATE TABLE subscription_plans (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name                VARCHAR(100) NOT NULL,
    slug                VARCHAR(100) NOT NULL UNIQUE,
    description         TEXT,
    tier                subscription_tier NOT NULL,
    pricing_eur_monthly NUMERIC(10,2) NOT NULL DEFAULT 0,
    pricing_eur_annual  NUMERIC(10,2) NOT NULL DEFAULT 0,
    max_users           INT NOT NULL DEFAULT 5,
    max_frameworks      INT NOT NULL DEFAULT 3,
    max_risks           INT NOT NULL DEFAULT 100,
    max_vendors         INT NOT NULL DEFAULT 20,
    max_storage_gb      INT NOT NULL DEFAULT 5,
    features            JSONB NOT NULL DEFAULT '{}',
    is_active           BOOLEAN NOT NULL DEFAULT true,
    sort_order          INT NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_subscription_plans_slug ON subscription_plans(slug);
CREATE INDEX idx_subscription_plans_tier ON subscription_plans(tier);
CREATE INDEX idx_subscription_plans_active ON subscription_plans(is_active) WHERE is_active = true;

-- ============================================================
-- ALTER: organization_subscriptions
-- Extend the existing table from migration 002 with new columns
-- for plan reference, Stripe integration, and usage tracking.
-- ============================================================

-- Add plan_id reference to subscription_plans
ALTER TABLE organization_subscriptions
    ADD COLUMN IF NOT EXISTS plan_id UUID REFERENCES subscription_plans(id),
    ADD COLUMN IF NOT EXISTS trial_ends_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS cancelled_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS cancel_reason TEXT,
    ADD COLUMN IF NOT EXISTS stripe_customer_id VARCHAR(255),
    ADD COLUMN IF NOT EXISTS stripe_subscription_id VARCHAR(255),
    ADD COLUMN IF NOT EXISTS usage_snapshot JSONB DEFAULT '{}';

-- Make organization_id unique (one active subscription per org)
-- First drop the existing index, then create a unique one
DROP INDEX IF EXISTS idx_org_subscriptions_org;
CREATE UNIQUE INDEX idx_org_subscriptions_org_unique ON organization_subscriptions(organization_id);

CREATE INDEX idx_org_subscriptions_plan ON organization_subscriptions(plan_id);
CREATE INDEX idx_org_subscriptions_status ON organization_subscriptions(status);
CREATE INDEX idx_org_subscriptions_stripe ON organization_subscriptions(stripe_customer_id) WHERE stripe_customer_id IS NOT NULL;
CREATE INDEX idx_org_subscriptions_period_end ON organization_subscriptions(current_period_end);

-- ============================================================
-- TABLE: onboarding_progress
-- Tracks multi-step onboarding wizard state per organisation.
-- ============================================================

CREATE TABLE onboarding_progress (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id         UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    current_step            INT NOT NULL DEFAULT 1,
    total_steps             INT NOT NULL DEFAULT 7,
    completed_steps         JSONB NOT NULL DEFAULT '[]',
    is_completed            BOOLEAN NOT NULL DEFAULT false,
    completed_at            TIMESTAMPTZ,
    skipped_steps           INT[] DEFAULT '{}',
    org_profile_data        JSONB DEFAULT '{}',
    industry_assessment_data JSONB DEFAULT '{}',
    selected_framework_ids  UUID[] DEFAULT '{}',
    team_invitations        JSONB DEFAULT '[]',
    risk_appetite_data      JSONB DEFAULT '{}',
    quick_assessment_data   JSONB DEFAULT '{}',
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_onboarding_progress_org UNIQUE (organization_id)
);

CREATE INDEX idx_onboarding_progress_org ON onboarding_progress(organization_id);
CREATE INDEX idx_onboarding_progress_completed ON onboarding_progress(is_completed);

-- ============================================================
-- TABLE: usage_events
-- Records resource consumption events for billing and limits.
-- ============================================================

CREATE TABLE usage_events (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id     UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    event_type          VARCHAR(100) NOT NULL,
    quantity            INT NOT NULL DEFAULT 1,
    metadata            JSONB DEFAULT '{}',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_usage_events_org ON usage_events(organization_id);
CREATE INDEX idx_usage_events_type ON usage_events(organization_id, event_type);
CREATE INDEX idx_usage_events_created ON usage_events(created_at);
CREATE INDEX idx_usage_events_org_type_created ON usage_events(organization_id, event_type, created_at DESC);

-- ============================================================
-- ROW-LEVEL SECURITY
-- ============================================================

-- Enable RLS on onboarding_progress
ALTER TABLE onboarding_progress ENABLE ROW LEVEL SECURITY;

CREATE POLICY onboarding_progress_tenant_isolation ON onboarding_progress
    USING (organization_id = current_setting('app.current_organization_id')::UUID);

CREATE POLICY onboarding_progress_tenant_insert ON onboarding_progress
    FOR INSERT WITH CHECK (organization_id = current_setting('app.current_organization_id')::UUID);

CREATE POLICY onboarding_progress_tenant_update ON onboarding_progress
    FOR UPDATE USING (organization_id = current_setting('app.current_organization_id')::UUID);

CREATE POLICY onboarding_progress_tenant_delete ON onboarding_progress
    FOR DELETE USING (organization_id = current_setting('app.current_organization_id')::UUID);

-- Enable RLS on usage_events
ALTER TABLE usage_events ENABLE ROW LEVEL SECURITY;

CREATE POLICY usage_events_tenant_isolation ON usage_events
    USING (organization_id = current_setting('app.current_organization_id')::UUID);

CREATE POLICY usage_events_tenant_insert ON usage_events
    FOR INSERT WITH CHECK (organization_id = current_setting('app.current_organization_id')::UUID);

CREATE POLICY usage_events_tenant_select ON usage_events
    FOR SELECT USING (organization_id = current_setting('app.current_organization_id')::UUID);

-- ============================================================
-- UPDATED_AT TRIGGERS
-- ============================================================

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_subscription_plans_updated_at
    BEFORE UPDATE ON subscription_plans
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_onboarding_progress_updated_at
    BEFORE UPDATE ON onboarding_progress
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
