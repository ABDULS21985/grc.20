-- ============================================================
-- ComplianceForge GRC Platform
-- Migration 001: Extensions & Enum Types
-- ============================================================

-- Extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "btree_gist";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- ============================================================
-- ENUM TYPES
-- ============================================================

-- Organization
CREATE TYPE org_status AS ENUM ('active', 'suspended', 'trial', 'deactivated');
CREATE TYPE org_tier AS ENUM ('starter', 'professional', 'enterprise', 'unlimited');

-- User
CREATE TYPE user_status AS ENUM ('active', 'inactive', 'locked', 'pending_verification');
CREATE TYPE mfa_method AS ENUM ('totp', 'sms', 'email', 'hardware_key');

-- Compliance
CREATE TYPE control_status AS ENUM ('not_applicable', 'not_implemented', 'planned', 'partial', 'implemented', 'effective');
CREATE TYPE control_type AS ENUM ('preventive', 'detective', 'corrective', 'directive', 'compensating');
CREATE TYPE implementation_type AS ENUM ('technical', 'administrative', 'physical', 'management');
CREATE TYPE mapping_type AS ENUM ('equivalent', 'partial', 'related', 'superset', 'subset');

-- Risk
CREATE TYPE risk_level AS ENUM ('critical', 'high', 'medium', 'low', 'very_low');
CREATE TYPE risk_status AS ENUM ('identified', 'assessed', 'treated', 'accepted', 'closed', 'monitoring');
CREATE TYPE treatment_type AS ENUM ('mitigate', 'transfer', 'avoid', 'accept');

-- Policy
CREATE TYPE policy_status AS ENUM ('draft', 'under_review', 'pending_approval', 'approved', 'published', 'archived', 'retired', 'superseded');
CREATE TYPE classification_level AS ENUM ('public', 'internal', 'confidential', 'restricted');

-- Audit
CREATE TYPE audit_status AS ENUM ('planned', 'in_progress', 'completed', 'cancelled');
CREATE TYPE finding_severity AS ENUM ('critical', 'high', 'medium', 'low', 'informational');

-- Incident
CREATE TYPE incident_severity AS ENUM ('critical', 'high', 'medium', 'low');
CREATE TYPE incident_status AS ENUM ('open', 'investigating', 'contained', 'resolved', 'closed');
CREATE TYPE incident_type AS ENUM ('data_breach', 'security', 'operational', 'compliance', 'whistleblower', 'fraud', 'health_safety', 'environmental');

-- Vendor
CREATE TYPE vendor_status AS ENUM ('pending', 'active', 'under_review', 'approved', 'rejected', 'suspended', 'offboarded');
CREATE TYPE vendor_risk_tier AS ENUM ('critical', 'high', 'medium', 'low');

-- ============================================================
-- HELPER FUNCTIONS FOR RLS
-- ============================================================

-- Set current tenant for Row-Level Security
CREATE OR REPLACE FUNCTION set_current_tenant(tenant_id UUID)
RETURNS VOID AS $$
BEGIN
    PERFORM set_config('app.current_tenant', tenant_id::TEXT, true);
END;
$$ LANGUAGE plpgsql;

-- Get current tenant for RLS policies
CREATE OR REPLACE FUNCTION get_current_tenant()
RETURNS UUID AS $$
BEGIN
    RETURN NULLIF(current_setting('app.current_tenant', true), '')::UUID;
EXCEPTION
    WHEN OTHERS THEN RETURN NULL;
END;
$$ LANGUAGE plpgsql STABLE;

-- ============================================================
-- TIMESTAMP TRIGGER FUNCTION
-- ============================================================

CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
