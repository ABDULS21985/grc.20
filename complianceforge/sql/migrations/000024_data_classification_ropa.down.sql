-- 000024_data_classification_ropa.down.sql
-- Rollback: Data Classification, Data Mapping & ROPA

BEGIN;

-- Drop triggers
DROP TRIGGER IF EXISTS set_data_flow_maps_updated_at ON data_flow_maps;
DROP TRIGGER IF EXISTS set_processing_activities_updated_at ON processing_activities;
DROP TRIGGER IF EXISTS set_data_categories_updated_at ON data_categories;
DROP TRIGGER IF EXISTS set_data_classifications_updated_at ON data_classifications;

-- Drop functions
DROP FUNCTION IF EXISTS generate_ropa_export_ref(UUID);
DROP FUNCTION IF EXISTS generate_pa_ref(UUID);

-- Drop tables (order matters due to FK constraints)
DROP TABLE IF EXISTS ropa_exports CASCADE;
DROP TABLE IF EXISTS data_flow_maps CASCADE;
DROP TABLE IF EXISTS processing_activities CASCADE;
DROP TABLE IF EXISTS data_categories CASCADE;
DROP TABLE IF EXISTS data_classifications CASCADE;

-- Drop enum types
DROP TYPE IF EXISTS ropa_export_reason;
DROP TYPE IF EXISTS ropa_export_format;
DROP TYPE IF EXISTS data_flow_dest_type;
DROP TYPE IF EXISTS data_flow_source_type;
DROP TYPE IF EXISTS data_flow_type;
DROP TYPE IF EXISTS dpia_status_type;
DROP TYPE IF EXISTS transfer_safeguard_type;
DROP TYPE IF EXISTS processing_role;
DROP TYPE IF EXISTS processing_status;
DROP TYPE IF EXISTS processing_legal_basis;
DROP TYPE IF EXISTS data_category_type;

COMMIT;
