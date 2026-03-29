-- 000025_board_reporting.up.sql
-- Executive Board Reporting Portal & Governance Dashboards

-- ============================================================
-- ENUMS
-- ============================================================

CREATE TYPE board_member_type AS ENUM (
    'executive_director',
    'non_executive_director',
    'independent_director',
    'committee_chair',
    'observer',
    'secretary'
);

CREATE TYPE board_meeting_type AS ENUM (
    'full_board',
    'audit_committee',
    'risk_committee',
    'remuneration_committee',
    'nomination_committee',
    'special',
    'agm',
    'egm'
);

CREATE TYPE board_meeting_status AS ENUM (
    'planned',
    'agenda_set',
    'in_progress',
    'completed',
    'minutes_approved'
);

CREATE TYPE board_decision_type AS ENUM (
    'policy_approval',
    'risk_acceptance',
    'budget_approval',
    'strategy_approval',
    'compliance_action',
    'incident_response',
    'vendor_approval',
    'regulatory_response',
    'general'
);

CREATE TYPE board_decision_outcome AS ENUM (
    'approved',
    'rejected',
    'deferred',
    'conditional_approval'
);

CREATE TYPE board_action_status AS ENUM (
    'pending',
    'in_progress',
    'completed',
    'overdue',
    'cancelled'
);

CREATE TYPE board_report_type AS ENUM (
    'board_pack',
    'compliance_summary',
    'risk_dashboard',
    'incident_report',
    'regulatory_update',
    'nis2_governance',
    'quarterly_review',
    'annual_review',
    'custom'
);

CREATE TYPE board_report_format AS ENUM (
    'pdf',
    'html',
    'docx',
    'xlsx'
);

-- ============================================================
-- TABLE: board_members
-- ============================================================

CREATE TABLE board_members (
    id                       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id          UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id                  UUID REFERENCES users(id) ON DELETE SET NULL,
    name                     VARCHAR(200) NOT NULL,
    title                    VARCHAR(200),
    email                    VARCHAR(300),
    member_type              board_member_type NOT NULL DEFAULT 'non_executive_director',
    committees               TEXT[] NOT NULL DEFAULT '{}',
    is_active                BOOLEAN NOT NULL DEFAULT true,
    portal_access_enabled    BOOLEAN NOT NULL DEFAULT false,
    portal_access_token_hash VARCHAR(128),
    portal_access_expires_at TIMESTAMPTZ,
    last_portal_access_at    TIMESTAMPTZ,
    created_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- TABLE: board_meetings
-- ============================================================

CREATE TABLE board_meetings (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id         UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    meeting_ref             VARCHAR(20) NOT NULL,
    title                   VARCHAR(300) NOT NULL,
    meeting_type            board_meeting_type NOT NULL DEFAULT 'full_board',
    date                    DATE NOT NULL,
    time                    TIME,
    location                VARCHAR(300),
    status                  board_meeting_status NOT NULL DEFAULT 'planned',
    agenda_items            JSONB NOT NULL DEFAULT '[]',
    board_pack_document_path TEXT,
    board_pack_generated_at TIMESTAMPTZ,
    minutes_document_path   TEXT,
    minutes_approved_at     TIMESTAMPTZ,
    minutes_approved_by     UUID REFERENCES users(id) ON DELETE SET NULL,
    attendees               UUID[] NOT NULL DEFAULT '{}',
    apologies               UUID[] NOT NULL DEFAULT '{}',
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_board_meetings_ref UNIQUE (organization_id, meeting_ref)
);

-- ============================================================
-- TABLE: board_decisions
-- ============================================================

CREATE TABLE board_decisions (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id       UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    meeting_id            UUID REFERENCES board_meetings(id) ON DELETE SET NULL,
    decision_ref          VARCHAR(20) NOT NULL,
    title                 VARCHAR(500) NOT NULL,
    description           TEXT,
    decision_type         board_decision_type NOT NULL DEFAULT 'general',
    decision              board_decision_outcome NOT NULL DEFAULT 'approved',
    conditions            TEXT,
    vote_for              INT NOT NULL DEFAULT 0,
    vote_against          INT NOT NULL DEFAULT 0,
    vote_abstain          INT NOT NULL DEFAULT 0,
    rationale             TEXT,
    linked_entity_type    VARCHAR(50),
    linked_entity_id      UUID,
    action_required       BOOLEAN NOT NULL DEFAULT false,
    action_description    TEXT,
    action_owner_user_id  UUID REFERENCES users(id) ON DELETE SET NULL,
    action_due_date       DATE,
    action_status         board_action_status NOT NULL DEFAULT 'pending',
    action_completed_at   TIMESTAMPTZ,
    decided_at            TIMESTAMPTZ,
    decided_by            VARCHAR(200),
    tags                  TEXT[] NOT NULL DEFAULT '{}',
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_board_decisions_ref UNIQUE (organization_id, decision_ref)
);

-- ============================================================
-- TABLE: board_reports
-- ============================================================

CREATE TABLE board_reports (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id   UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    meeting_id        UUID REFERENCES board_meetings(id) ON DELETE SET NULL,
    report_type       board_report_type NOT NULL DEFAULT 'board_pack',
    title             VARCHAR(300) NOT NULL,
    period_start      DATE,
    period_end        DATE,
    file_path         TEXT,
    file_format       board_report_format NOT NULL DEFAULT 'pdf',
    generated_by      UUID REFERENCES users(id) ON DELETE SET NULL,
    generated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    classification    VARCHAR(50) NOT NULL DEFAULT 'board_confidential',
    page_count        INT NOT NULL DEFAULT 0,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- ROW-LEVEL SECURITY
-- ============================================================

ALTER TABLE board_members ENABLE ROW LEVEL SECURITY;
ALTER TABLE board_meetings ENABLE ROW LEVEL SECURITY;
ALTER TABLE board_decisions ENABLE ROW LEVEL SECURITY;
ALTER TABLE board_reports ENABLE ROW LEVEL SECURITY;

CREATE POLICY board_members_org_isolation ON board_members
    USING (organization_id = current_setting('app.current_org', true)::uuid);

CREATE POLICY board_meetings_org_isolation ON board_meetings
    USING (organization_id = current_setting('app.current_org', true)::uuid);

CREATE POLICY board_decisions_org_isolation ON board_decisions
    USING (organization_id = current_setting('app.current_org', true)::uuid);

CREATE POLICY board_reports_org_isolation ON board_reports
    USING (organization_id = current_setting('app.current_org', true)::uuid);

-- ============================================================
-- AUTO-REF FUNCTIONS
-- ============================================================

-- Meeting ref: BM-YYYY-QN (e.g., BM-2026-Q1)
CREATE OR REPLACE FUNCTION generate_board_meeting_ref()
RETURNS TRIGGER AS $$
DECLARE
    yr   TEXT;
    qtr  TEXT;
    seq  INT;
    ref  TEXT;
BEGIN
    yr := EXTRACT(YEAR FROM NEW.date)::TEXT;
    qtr := 'Q' || CEIL(EXTRACT(MONTH FROM NEW.date) / 3.0)::TEXT;

    SELECT COALESCE(MAX(
        CASE WHEN meeting_ref ~ ('^BM-' || yr || '-' || qtr || '-\d+$')
             THEN CAST(SUBSTRING(meeting_ref FROM '\d+$') AS INT)
             ELSE 0 END
    ), 0) + 1
    INTO seq
    FROM board_meetings
    WHERE organization_id = NEW.organization_id;

    ref := 'BM-' || yr || '-' || qtr || '-' || LPAD(seq::TEXT, 2, '0');
    NEW.meeting_ref := ref;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_board_meeting_ref
    BEFORE INSERT ON board_meetings
    FOR EACH ROW
    WHEN (NEW.meeting_ref IS NULL OR NEW.meeting_ref = '')
    EXECUTE FUNCTION generate_board_meeting_ref();

-- Decision ref: BD-YYYY-NNN
CREATE OR REPLACE FUNCTION generate_board_decision_ref()
RETURNS TRIGGER AS $$
DECLARE
    yr  TEXT;
    seq INT;
BEGIN
    yr := EXTRACT(YEAR FROM NOW())::TEXT;

    SELECT COALESCE(MAX(
        CASE WHEN decision_ref ~ ('^BD-' || yr || '-\d+$')
             THEN CAST(SUBSTRING(decision_ref FROM '\d+$') AS INT)
             ELSE 0 END
    ), 0) + 1
    INTO seq
    FROM board_decisions
    WHERE organization_id = NEW.organization_id;

    NEW.decision_ref := 'BD-' || yr || '-' || LPAD(seq::TEXT, 3, '0');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_board_decision_ref
    BEFORE INSERT ON board_decisions
    FOR EACH ROW
    WHEN (NEW.decision_ref IS NULL OR NEW.decision_ref = '')
    EXECUTE FUNCTION generate_board_decision_ref();

-- ============================================================
-- AUTO-UPDATE updated_at TRIGGERS
-- ============================================================

CREATE TRIGGER set_board_members_updated_at
    BEFORE UPDATE ON board_members
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER set_board_meetings_updated_at
    BEFORE UPDATE ON board_meetings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER set_board_decisions_updated_at
    BEFORE UPDATE ON board_decisions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================
-- INDEXES
-- ============================================================

CREATE INDEX idx_board_members_org ON board_members(organization_id);
CREATE INDEX idx_board_members_user ON board_members(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX idx_board_members_active ON board_members(organization_id) WHERE is_active = true;
CREATE INDEX idx_board_members_token ON board_members(portal_access_token_hash) WHERE portal_access_token_hash IS NOT NULL;

CREATE INDEX idx_board_meetings_org ON board_meetings(organization_id);
CREATE INDEX idx_board_meetings_date ON board_meetings(organization_id, date DESC);
CREATE INDEX idx_board_meetings_status ON board_meetings(organization_id, status);

CREATE INDEX idx_board_decisions_org ON board_decisions(organization_id);
CREATE INDEX idx_board_decisions_meeting ON board_decisions(meeting_id) WHERE meeting_id IS NOT NULL;
CREATE INDEX idx_board_decisions_action_status ON board_decisions(organization_id, action_status) WHERE action_required = true;
CREATE INDEX idx_board_decisions_action_due ON board_decisions(action_due_date) WHERE action_required = true AND action_status NOT IN ('completed', 'cancelled');

CREATE INDEX idx_board_reports_org ON board_reports(organization_id);
CREATE INDEX idx_board_reports_meeting ON board_reports(meeting_id) WHERE meeting_id IS NOT NULL;
CREATE INDEX idx_board_reports_type ON board_reports(organization_id, report_type);
