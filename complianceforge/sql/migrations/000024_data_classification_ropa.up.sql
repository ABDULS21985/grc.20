-- 000024_data_classification_ropa.up.sql
-- Data Classification, Data Mapping & ROPA (Record of Processing Activities)
-- Implements GDPR Art.30 ROPA, data classification levels, data categories,
-- processing activities, data flow mapping, and ROPA export management.

BEGIN;

-- ============================================================
-- ENUM TYPES
-- ============================================================

CREATE TYPE data_category_type AS ENUM (
    'personal_data',
    'special_category',
    'financial',
    'technical',
    'business',
    'public',
    'proprietary'
);

CREATE TYPE processing_legal_basis AS ENUM (
    'consent',
    'contract',
    'legal_obligation',
    'vital_interest',
    'public_task',
    'legitimate_interest'
);

CREATE TYPE processing_status AS ENUM (
    'draft',
    'active',
    'under_review',
    'suspended',
    'retired'
);

CREATE TYPE processing_role AS ENUM (
    'controller',
    'joint_controller',
    'processor'
);

CREATE TYPE transfer_safeguard_type AS ENUM (
    'adequacy_decision',
    'standard_contractual_clauses',
    'binding_corporate_rules',
    'approved_code_of_conduct',
    'approved_certification',
    'derogation_art49',
    'none'
);

CREATE TYPE dpia_status_type AS ENUM (
    'not_required',
    'required',
    'in_progress',
    'completed',
    'review_needed'
);

CREATE TYPE data_flow_type AS ENUM (
    'collection',
    'storage',
    'processing',
    'transfer',
    'sharing',
    'deletion'
);

CREATE TYPE data_flow_source_type AS ENUM (
    'data_subject',
    'internal_system',
    'external_system',
    'vendor',
    'public_source',
    'partner'
);

CREATE TYPE data_flow_dest_type AS ENUM (
    'internal_system',
    'external_system',
    'vendor',
    'regulator',
    'data_subject',
    'partner',
    'archive',
    'deletion'
);

CREATE TYPE ropa_export_format AS ENUM (
    'pdf',
    'xlsx',
    'csv',
    'json'
);

CREATE TYPE ropa_export_reason AS ENUM (
    'regulatory_request',
    'internal_audit',
    'dpa_request',
    'annual_review',
    'ad_hoc'
);

-- ============================================================
-- TABLE: data_classifications
-- ============================================================

CREATE TABLE data_classifications (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id             UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name                        VARCHAR(100) NOT NULL,
    level                       INT NOT NULL DEFAULT 0,
    description                 TEXT,
    handling_requirements       TEXT,
    encryption_required         BOOLEAN NOT NULL DEFAULT false,
    access_restriction_required BOOLEAN NOT NULL DEFAULT false,
    data_masking_required       BOOLEAN NOT NULL DEFAULT false,
    retention_policy            TEXT,
    disposal_method             TEXT,
    color_hex                   VARCHAR(7),
    is_system                   BOOLEAN NOT NULL DEFAULT false,
    sort_order                  INT NOT NULL DEFAULT 0,
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_classification_org_name UNIQUE (organization_id, name),
    CONSTRAINT uq_classification_org_level UNIQUE (organization_id, level)
);

-- ============================================================
-- TABLE: data_categories
-- ============================================================

CREATE TABLE data_categories (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id         UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name                    VARCHAR(200) NOT NULL,
    category_type           data_category_type NOT NULL DEFAULT 'personal_data',
    gdpr_special_category   BOOLEAN NOT NULL DEFAULT false,
    gdpr_article_9_basis    TEXT,
    description             TEXT,
    examples                TEXT[],
    classification_id       UUID REFERENCES data_classifications(id) ON DELETE SET NULL,
    retention_period_months INT,
    is_system               BOOLEAN NOT NULL DEFAULT false,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_category_org_name UNIQUE (organization_id, name)
);

-- ============================================================
-- TABLE: processing_activities
-- ============================================================

CREATE TABLE processing_activities (
    id                              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id                 UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    activity_ref                    VARCHAR(20) NOT NULL,
    name                            VARCHAR(500) NOT NULL,
    description                     TEXT,
    purpose                         TEXT,
    legal_basis                     processing_legal_basis,
    legal_basis_detail              TEXT,
    status                          processing_status NOT NULL DEFAULT 'draft',
    role                            processing_role NOT NULL DEFAULT 'controller',
    joint_controller_details        TEXT,
    data_subject_categories         TEXT[],
    estimated_data_subjects_count   INT,
    data_category_ids               UUID[],
    special_categories_processed    BOOLEAN NOT NULL DEFAULT false,
    special_categories_legal_basis  TEXT,
    recipient_categories            TEXT[],
    recipient_vendor_ids            UUID[],
    involves_international_transfer BOOLEAN NOT NULL DEFAULT false,
    transfer_countries              TEXT[],
    transfer_safeguards             transfer_safeguard_type,
    transfer_safeguards_detail      TEXT,
    tia_conducted                   BOOLEAN NOT NULL DEFAULT false,
    tia_date                        DATE,
    tia_document_path               TEXT,
    retention_period_months         INT,
    retention_justification         TEXT,
    deletion_method                 TEXT,
    deletion_responsible_user_id    UUID,
    system_ids                      UUID[],
    storage_locations               TEXT[],
    dpia_required                   BOOLEAN NOT NULL DEFAULT false,
    dpia_status                     dpia_status_type DEFAULT 'not_required',
    dpia_document_path              TEXT,
    dpia_conducted_date             DATE,
    security_measures               TEXT,
    linked_control_codes            TEXT[],
    risk_level                      VARCHAR(20) DEFAULT 'medium',
    data_steward_user_id            UUID,
    department                      VARCHAR(200),
    process_owner_user_id           UUID,
    last_review_date                DATE,
    next_review_date                DATE,
    review_frequency_months         INT DEFAULT 12,
    metadata                        JSONB DEFAULT '{}'::jsonb,
    created_at                      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at                      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at                      TIMESTAMPTZ,

    CONSTRAINT uq_activity_org_ref UNIQUE (organization_id, activity_ref)
);

-- ============================================================
-- TABLE: data_flow_maps
-- ============================================================

CREATE TABLE data_flow_maps (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id             UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    processing_activity_id      UUID NOT NULL REFERENCES processing_activities(id) ON DELETE CASCADE,
    name                        VARCHAR(300) NOT NULL,
    flow_type                   data_flow_type NOT NULL DEFAULT 'processing',
    source_type                 data_flow_source_type NOT NULL DEFAULT 'internal_system',
    source_name                 VARCHAR(300) NOT NULL,
    source_entity_id            UUID,
    destination_type            data_flow_dest_type NOT NULL DEFAULT 'internal_system',
    destination_name            VARCHAR(300) NOT NULL,
    destination_entity_id       UUID,
    destination_country         VARCHAR(5),
    data_category_ids           UUID[],
    transfer_method             VARCHAR(200),
    encryption_in_transit       BOOLEAN NOT NULL DEFAULT false,
    encryption_at_rest          BOOLEAN NOT NULL DEFAULT false,
    volume_description          VARCHAR(200),
    frequency                   VARCHAR(100),
    legal_basis                 VARCHAR(100),
    notes                       TEXT,
    sort_order                  INT NOT NULL DEFAULT 0,
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ============================================================
-- TABLE: ropa_exports
-- ============================================================

CREATE TABLE ropa_exports (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id         UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    export_ref              VARCHAR(20) NOT NULL,
    export_date             TIMESTAMPTZ NOT NULL DEFAULT now(),
    format                  ropa_export_format NOT NULL DEFAULT 'pdf',
    file_path               TEXT,
    activities_included     INT NOT NULL DEFAULT 0,
    exported_by             UUID,
    export_reason           ropa_export_reason NOT NULL DEFAULT 'ad_hoc',
    notes                   TEXT,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_ropa_export_org_ref UNIQUE (organization_id, export_ref)
);

-- ============================================================
-- AUTO-REFERENCE FUNCTION: generate_pa_ref
-- Generates PA-001, PA-002, etc. per organization
-- ============================================================

CREATE OR REPLACE FUNCTION generate_pa_ref(org_id UUID)
RETURNS VARCHAR(20) AS $$
DECLARE
    next_num INT;
    ref_val  VARCHAR(20);
BEGIN
    SELECT COALESCE(MAX(
        CAST(SUBSTRING(activity_ref FROM 4) AS INT)
    ), 0) + 1
    INTO next_num
    FROM processing_activities
    WHERE organization_id = org_id;

    ref_val := 'PA-' || LPAD(next_num::TEXT, 3, '0');
    RETURN ref_val;
END;
$$ LANGUAGE plpgsql;

-- ============================================================
-- AUTO-REFERENCE FUNCTION: generate_ropa_export_ref
-- Generates ROPA-001, ROPA-002, etc. per organization
-- ============================================================

CREATE OR REPLACE FUNCTION generate_ropa_export_ref(org_id UUID)
RETURNS VARCHAR(20) AS $$
DECLARE
    next_num INT;
    ref_val  VARCHAR(20);
BEGIN
    SELECT COALESCE(MAX(
        CAST(SUBSTRING(export_ref FROM 6) AS INT)
    ), 0) + 1
    INTO next_num
    FROM ropa_exports
    WHERE organization_id = org_id;

    ref_val := 'ROPA-' || LPAD(next_num::TEXT, 3, '0');
    RETURN ref_val;
END;
$$ LANGUAGE plpgsql;

-- ============================================================
-- ROW LEVEL SECURITY
-- ============================================================

ALTER TABLE data_classifications ENABLE ROW LEVEL SECURITY;
ALTER TABLE data_categories ENABLE ROW LEVEL SECURITY;
ALTER TABLE processing_activities ENABLE ROW LEVEL SECURITY;
ALTER TABLE data_flow_maps ENABLE ROW LEVEL SECURITY;
ALTER TABLE ropa_exports ENABLE ROW LEVEL SECURITY;

CREATE POLICY data_classifications_org_isolation ON data_classifications
    USING (organization_id = current_setting('app.current_org', true)::UUID);

CREATE POLICY data_categories_org_isolation ON data_categories
    USING (organization_id = current_setting('app.current_org', true)::UUID);

CREATE POLICY processing_activities_org_isolation ON processing_activities
    USING (organization_id = current_setting('app.current_org', true)::UUID);

CREATE POLICY data_flow_maps_org_isolation ON data_flow_maps
    USING (organization_id = current_setting('app.current_org', true)::UUID);

CREATE POLICY ropa_exports_org_isolation ON ropa_exports
    USING (organization_id = current_setting('app.current_org', true)::UUID);

-- ============================================================
-- INDEXES
-- ============================================================

-- data_classifications
CREATE INDEX idx_data_classifications_org ON data_classifications(organization_id);
CREATE INDEX idx_data_classifications_level ON data_classifications(organization_id, level);

-- data_categories
CREATE INDEX idx_data_categories_org ON data_categories(organization_id);
CREATE INDEX idx_data_categories_type ON data_categories(organization_id, category_type);
CREATE INDEX idx_data_categories_classification ON data_categories(classification_id);
CREATE INDEX idx_data_categories_special ON data_categories(organization_id) WHERE gdpr_special_category = true;

-- processing_activities
CREATE INDEX idx_processing_activities_org ON processing_activities(organization_id);
CREATE INDEX idx_processing_activities_status ON processing_activities(organization_id, status);
CREATE INDEX idx_processing_activities_legal_basis ON processing_activities(organization_id, legal_basis);
CREATE INDEX idx_processing_activities_ref ON processing_activities(organization_id, activity_ref);
CREATE INDEX idx_processing_activities_dept ON processing_activities(organization_id, department);
CREATE INDEX idx_processing_activities_dpia ON processing_activities(organization_id, dpia_status);
CREATE INDEX idx_processing_activities_transfer ON processing_activities(organization_id)
    WHERE involves_international_transfer = true;
CREATE INDEX idx_processing_activities_special ON processing_activities(organization_id)
    WHERE special_categories_processed = true;
CREATE INDEX idx_processing_activities_review ON processing_activities(organization_id, next_review_date);
CREATE INDEX idx_processing_activities_deleted ON processing_activities(organization_id)
    WHERE deleted_at IS NULL;

-- data_flow_maps
CREATE INDEX idx_data_flow_maps_org ON data_flow_maps(organization_id);
CREATE INDEX idx_data_flow_maps_activity ON data_flow_maps(processing_activity_id);
CREATE INDEX idx_data_flow_maps_dest_country ON data_flow_maps(organization_id, destination_country)
    WHERE destination_country IS NOT NULL;

-- ropa_exports
CREATE INDEX idx_ropa_exports_org ON ropa_exports(organization_id);
CREATE INDEX idx_ropa_exports_date ON ropa_exports(organization_id, export_date DESC);

-- ============================================================
-- UPDATED_AT TRIGGERS
-- ============================================================

CREATE TRIGGER set_data_classifications_updated_at
    BEFORE UPDATE ON data_classifications
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER set_data_categories_updated_at
    BEFORE UPDATE ON data_categories
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER set_processing_activities_updated_at
    BEFORE UPDATE ON processing_activities
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER set_data_flow_maps_updated_at
    BEFORE UPDATE ON data_flow_maps
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMIT;
