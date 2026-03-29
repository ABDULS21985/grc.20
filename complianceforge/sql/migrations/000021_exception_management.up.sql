-- 000021_exception_management.up.sql
-- Exception Management & Compensating Controls

-- ============================================================
-- ENUM TYPES
-- ============================================================

CREATE TYPE exception_type AS ENUM ('permanent', 'temporary', 'conditional');

CREATE TYPE exception_status AS ENUM (
    'draft',
    'pending_risk_assessment',
    'pending_approval',
    'approved',
    'rejected',
    'expired',
    'revoked',
    'renewal_pending'
);

CREATE TYPE exception_scope_type AS ENUM (
    'control_implementation',
    'framework_control',
    'policy',
    'process',
    'system',
    'organization_wide'
);

CREATE TYPE exception_review_type AS ENUM (
    'periodic',
    'risk_reassessment',
    'incident_triggered',
    'audit_triggered',
    'renewal',
    'ad_hoc'
);

CREATE TYPE exception_review_outcome AS ENUM (
    'continue',
    'modify',
    'escalate',
    'revoke',
    'renew'
);

CREATE TYPE compensating_effectiveness AS ENUM (
    'fully_effective',
    'mostly_effective',
    'partially_effective',
    'minimally_effective',
    'not_effective',
    'not_assessed'
);

-- ============================================================
-- TABLE: compliance_exceptions
-- ============================================================

CREATE TABLE compliance_exceptions (
    id                              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id                 UUID NOT NULL REFERENCES organizations(id),
    exception_ref                   VARCHAR(20) NOT NULL UNIQUE,
    title                           VARCHAR(500) NOT NULL,
    description                     TEXT NOT NULL,

    -- Classification
    exception_type                  exception_type NOT NULL DEFAULT 'temporary',
    status                          exception_status NOT NULL DEFAULT 'draft',
    priority                        VARCHAR(20) NOT NULL DEFAULT 'medium'
                                        CHECK (priority IN ('critical', 'high', 'medium', 'low')),

    -- Scope
    scope_type                      exception_scope_type NOT NULL DEFAULT 'control_implementation',
    control_implementation_ids      UUID[],
    framework_control_codes         TEXT[],
    policy_id                       UUID,
    scope_description               TEXT,

    -- Risk
    risk_justification              TEXT NOT NULL,
    residual_risk_description       TEXT,
    residual_risk_level             VARCHAR(20) DEFAULT 'medium'
                                        CHECK (residual_risk_level IN ('critical', 'high', 'medium', 'low', 'very_low')),
    risk_assessment_id              UUID,
    risk_accepted_by                UUID,
    risk_accepted_at                TIMESTAMPTZ,

    -- Compensating Controls
    has_compensating_controls       BOOLEAN NOT NULL DEFAULT false,
    compensating_controls_description TEXT,
    compensating_control_ids        UUID[],
    compensating_effectiveness      compensating_effectiveness DEFAULT 'not_assessed',

    -- Lifecycle / Approval
    requested_by                    UUID NOT NULL,
    requested_at                    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    approved_by                     UUID,
    approved_at                     TIMESTAMPTZ,
    approval_comments               TEXT,
    rejection_reason                TEXT,
    workflow_instance_id            UUID,

    -- Validity & Review
    effective_date                  DATE NOT NULL,
    expiry_date                     DATE,
    review_frequency_months         INT NOT NULL DEFAULT 12,
    next_review_date                DATE,
    last_review_date                DATE,
    last_reviewed_by                UUID,
    renewal_count                   INT NOT NULL DEFAULT 0,

    -- Conditions & Audit
    conditions                      TEXT,
    business_impact_if_implemented  TEXT,
    regulatory_notification_required BOOLEAN NOT NULL DEFAULT false,
    regulator_notified_at           TIMESTAMPTZ,
    audit_evidence_path             TEXT,

    -- Metadata
    tags                            TEXT[],
    metadata                        JSONB DEFAULT '{}',
    created_at                      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at                      TIMESTAMPTZ
);

-- ============================================================
-- TABLE: exception_reviews
-- ============================================================

CREATE TABLE exception_reviews (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id     UUID NOT NULL REFERENCES organizations(id),
    exception_id        UUID NOT NULL REFERENCES compliance_exceptions(id) ON DELETE CASCADE,
    review_type         exception_review_type NOT NULL DEFAULT 'periodic',
    reviewer_id         UUID NOT NULL,
    review_date         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    outcome             exception_review_outcome NOT NULL,
    risk_level_at_review VARCHAR(20) CHECK (risk_level_at_review IN ('critical', 'high', 'medium', 'low', 'very_low')),
    compensating_effective BOOLEAN,
    findings            TEXT,
    recommendations     TEXT,
    next_review_date    DATE,
    attachments         TEXT[],
    metadata            JSONB DEFAULT '{}',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- TABLE: exception_audit_trail
-- ============================================================

CREATE TABLE exception_audit_trail (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id     UUID NOT NULL REFERENCES organizations(id),
    exception_id        UUID NOT NULL REFERENCES compliance_exceptions(id) ON DELETE CASCADE,
    action              VARCHAR(100) NOT NULL,
    actor_id            UUID NOT NULL,
    actor_email         VARCHAR(255),
    old_status          VARCHAR(50),
    new_status          VARCHAR(50),
    details             TEXT,
    ip_address          VARCHAR(45),
    user_agent          TEXT,
    metadata            JSONB DEFAULT '{}',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- ROW LEVEL SECURITY
-- ============================================================

ALTER TABLE compliance_exceptions ENABLE ROW LEVEL SECURITY;
CREATE POLICY compliance_exceptions_org_isolation ON compliance_exceptions
    USING (organization_id = current_setting('app.current_org')::uuid);

ALTER TABLE exception_reviews ENABLE ROW LEVEL SECURITY;
CREATE POLICY exception_reviews_org_isolation ON exception_reviews
    USING (organization_id = current_setting('app.current_org')::uuid);

ALTER TABLE exception_audit_trail ENABLE ROW LEVEL SECURITY;
CREATE POLICY exception_audit_trail_org_isolation ON exception_audit_trail
    USING (organization_id = current_setting('app.current_org')::uuid);

-- ============================================================
-- AUTO-REF FUNCTION
-- ============================================================

CREATE OR REPLACE FUNCTION generate_exception_ref()
RETURNS TRIGGER AS $$
DECLARE
    next_seq INT;
    ref_year TEXT;
BEGIN
    ref_year := TO_CHAR(NOW(), 'YYYY');
    SELECT COALESCE(MAX(
        CAST(SUBSTRING(exception_ref FROM 'EXC-' || ref_year || '-(\d+)') AS INT)
    ), 0) + 1
    INTO next_seq
    FROM compliance_exceptions
    WHERE exception_ref LIKE 'EXC-' || ref_year || '-%';

    NEW.exception_ref := 'EXC-' || ref_year || '-' || LPAD(next_seq::TEXT, 4, '0');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_exception_ref
    BEFORE INSERT ON compliance_exceptions
    FOR EACH ROW
    WHEN (NEW.exception_ref IS NULL OR NEW.exception_ref = '')
    EXECUTE FUNCTION generate_exception_ref();

-- ============================================================
-- AUTO-UPDATE updated_at
-- ============================================================

CREATE OR REPLACE FUNCTION update_exception_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_exception_updated_at
    BEFORE UPDATE ON compliance_exceptions
    FOR EACH ROW
    EXECUTE FUNCTION update_exception_updated_at();

-- ============================================================
-- INDEXES
-- ============================================================

CREATE INDEX idx_compliance_exceptions_org_id ON compliance_exceptions(organization_id);
CREATE INDEX idx_compliance_exceptions_status ON compliance_exceptions(status);
CREATE INDEX idx_compliance_exceptions_expiry ON compliance_exceptions(expiry_date) WHERE expiry_date IS NOT NULL;
CREATE INDEX idx_compliance_exceptions_next_review ON compliance_exceptions(next_review_date) WHERE next_review_date IS NOT NULL;
CREATE INDEX idx_compliance_exceptions_org_status ON compliance_exceptions(organization_id, status) WHERE deleted_at IS NULL;
CREATE INDEX idx_compliance_exceptions_requested_by ON compliance_exceptions(requested_by);
CREATE INDEX idx_compliance_exceptions_priority ON compliance_exceptions(priority);

CREATE INDEX idx_exception_reviews_exception_id ON exception_reviews(exception_id);
CREATE INDEX idx_exception_reviews_org_id ON exception_reviews(organization_id);

CREATE INDEX idx_exception_audit_trail_exception_id ON exception_audit_trail(exception_id);
CREATE INDEX idx_exception_audit_trail_org_id ON exception_audit_trail(organization_id);
CREATE INDEX idx_exception_audit_trail_created ON exception_audit_trail(created_at DESC);
