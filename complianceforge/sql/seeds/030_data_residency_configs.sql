-- 030_data_residency_configs.sql
-- Seed data for Data Residency region configurations.
-- Provides default configurations for all supported deployment regions.

INSERT INTO data_residency_configs (
    region, display_name, description,
    allowed_countries, blocked_countries,
    allowed_cloud_regions,
    primary_cloud_region, failover_cloud_region,
    storage_config,
    compliance_frameworks, legal_basis,
    data_protection_authority, dpa_contact_url,
    gdpr_adequate_countries,
    enforcement_mode,
    allow_cross_region_search, allow_cross_region_backup,
    is_active
) VALUES

-- ============================================================
-- EU (European Union / EEA)
-- ============================================================
(
    'eu',
    'European Union',
    'Data stored exclusively within EU/EEA member states. Full GDPR compliance with strict cross-border transfer controls.',
    ARRAY['AT','BE','BG','HR','CY','CZ','DK','EE','FI','FR','DE','GR','HU','IE','IT','LV','LT','LU','MT','NL','PL','PT','RO','SK','SI','ES','SE','IS','LI','NO'],
    '{}',
    '{
        "aws": ["eu-west-1", "eu-west-2", "eu-central-1", "eu-north-1", "eu-south-1"],
        "azure": ["westeurope", "northeurope", "germanywestcentral", "swedencentral", "francecentral"],
        "gcp": ["europe-west1", "europe-west3", "europe-west4", "europe-north1"]
    }'::jsonb,
    'eu-west-1',
    'eu-central-1',
    '{
        "database": {"provider": "AWS RDS", "region": "eu-west-1", "encryption": "AES-256", "backup_region": "eu-central-1"},
        "files": {"provider": "AWS S3", "region": "eu-west-1", "encryption": "AES-256-GCM"},
        "cache": {"provider": "AWS ElastiCache", "region": "eu-west-1", "encryption": "in-transit-and-at-rest"},
        "search": {"provider": "AWS OpenSearch", "region": "eu-west-1", "encryption": "AES-256"}
    }'::jsonb,
    ARRAY['GDPR', 'ePrivacy Directive', 'NIS2', 'DORA', 'EU AI Act'],
    'GDPR Article 6(1)(b) - Processing necessary for the performance of a contract',
    'European Data Protection Board (EDPB)',
    'https://edpb.europa.eu/about-edpb/about-edpb/members_en',
    ARRAY['AD','AR','CA','FO','GG','IL','IM','JP','JE','NZ','CH','UY','KR','GB','US'],
    'enforce',
    false,
    true,
    true
),

-- ============================================================
-- UK (United Kingdom)
-- ============================================================
(
    'uk',
    'United Kingdom',
    'Data stored within the United Kingdom. Compliant with UK GDPR and Data Protection Act 2018.',
    ARRAY['GB'],
    '{}',
    '{
        "aws": ["eu-west-2"],
        "azure": ["uksouth", "ukwest"],
        "gcp": ["europe-west2"]
    }'::jsonb,
    'eu-west-2',
    'eu-west-1',
    '{
        "database": {"provider": "AWS RDS", "region": "eu-west-2", "encryption": "AES-256", "backup_region": "eu-west-1"},
        "files": {"provider": "AWS S3", "region": "eu-west-2", "encryption": "AES-256-GCM"},
        "cache": {"provider": "AWS ElastiCache", "region": "eu-west-2", "encryption": "in-transit-and-at-rest"},
        "search": {"provider": "AWS OpenSearch", "region": "eu-west-2", "encryption": "AES-256"}
    }'::jsonb,
    ARRAY['UK GDPR', 'Data Protection Act 2018', 'NIS Regulations 2018', 'Cyber Essentials', 'PECR'],
    'UK GDPR Article 6(1)(b) - Processing necessary for the performance of a contract',
    'Information Commissioner''s Office (ICO)',
    'https://ico.org.uk/make-a-complaint/',
    ARRAY['AD','AR','CA','FO','GG','IL','IM','JP','JE','NZ','CH','UY','KR','AT','BE','BG','HR','CY','CZ','DK','EE','FI','FR','DE','GR','HU','IE','IT','LV','LT','LU','MT','NL','PL','PT','RO','SK','SI','ES','SE','US'],
    'enforce',
    false,
    true,
    true
),

-- ============================================================
-- DACH (Germany, Austria, Switzerland)
-- ============================================================
(
    'dach',
    'DACH Region',
    'Data stored within Germany, Austria, or Switzerland. Enhanced privacy protections meeting the strictest EU requirements plus Swiss FADP.',
    ARRAY['DE','AT','CH'],
    '{}',
    '{
        "aws": ["eu-central-1"],
        "azure": ["germanywestcentral", "switzerlandnorth"],
        "gcp": ["europe-west3"]
    }'::jsonb,
    'eu-central-1',
    'eu-west-1',
    '{
        "database": {"provider": "AWS RDS", "region": "eu-central-1", "encryption": "AES-256", "backup_region": "eu-west-1"},
        "files": {"provider": "AWS S3", "region": "eu-central-1", "encryption": "AES-256-GCM"},
        "cache": {"provider": "AWS ElastiCache", "region": "eu-central-1", "encryption": "in-transit-and-at-rest"},
        "search": {"provider": "AWS OpenSearch", "region": "eu-central-1", "encryption": "AES-256"}
    }'::jsonb,
    ARRAY['GDPR', 'BDSG (Germany)', 'DSG (Austria)', 'FADP/nDSG (Switzerland)', 'C5 (BSI)', 'TISAX', 'NIS2'],
    'GDPR Article 6(1)(b) with enhanced BDSG requirements',
    'BfDI (Germany), DSB (Austria), FDPIC (Switzerland)',
    'https://www.bfdi.bund.de/EN/Home/home_node.html',
    ARRAY['AD','AR','CA','FO','GG','IL','IM','JP','JE','NZ','UY','KR','GB','US','AT','BE','BG','HR','CY','CZ','DK','EE','FI','FR','GR','HU','IE','IT','LV','LT','LU','MT','NL','PL','PT','RO','SK','SI','ES','SE'],
    'enforce',
    false,
    true,
    true
),

-- ============================================================
-- NORDICS (Denmark, Finland, Iceland, Norway, Sweden)
-- ============================================================
(
    'nordics',
    'Nordic Region',
    'Data stored within the Nordic countries. Compliant with GDPR plus local Nordic data protection legislation.',
    ARRAY['DK','FI','IS','NO','SE'],
    '{}',
    '{
        "aws": ["eu-north-1"],
        "azure": ["swedencentral", "norwayeast"],
        "gcp": ["europe-north1"]
    }'::jsonb,
    'eu-north-1',
    'eu-west-1',
    '{
        "database": {"provider": "AWS RDS", "region": "eu-north-1", "encryption": "AES-256", "backup_region": "eu-west-1"},
        "files": {"provider": "AWS S3", "region": "eu-north-1", "encryption": "AES-256-GCM"},
        "cache": {"provider": "AWS ElastiCache", "region": "eu-north-1", "encryption": "in-transit-and-at-rest"},
        "search": {"provider": "AWS OpenSearch", "region": "eu-north-1", "encryption": "AES-256"}
    }'::jsonb,
    ARRAY['GDPR', 'Databeskyttelsesloven (Denmark)', 'Tietosuojalaki (Finland)', 'Personuppgiftslagen (Sweden)', 'Personopplysningsloven (Norway)', 'NIS2'],
    'GDPR Article 6(1)(b) - Processing necessary for the performance of a contract',
    'Nordic DPAs Collective',
    'https://www.datatilsynet.dk/',
    ARRAY['AD','AR','CA','FO','GG','IL','IM','JP','JE','NZ','CH','UY','KR','GB','US','AT','BE','BG','HR','CY','CZ','EE','FR','DE','GR','HU','IE','IT','LV','LT','LU','MT','NL','PL','PT','RO','SK','SI','ES'],
    'enforce',
    false,
    true,
    true
),

-- ============================================================
-- FRANCE
-- ============================================================
(
    'france',
    'France',
    'Data stored exclusively within France. Compliant with GDPR and French data protection law (Loi Informatique et Libertes).',
    ARRAY['FR'],
    '{}',
    '{
        "aws": ["eu-west-3"],
        "azure": ["francecentral", "francesouth"],
        "gcp": ["europe-west1", "europe-west9"]
    }'::jsonb,
    'eu-west-3',
    'eu-central-1',
    '{
        "database": {"provider": "AWS RDS", "region": "eu-west-3", "encryption": "AES-256", "backup_region": "eu-central-1"},
        "files": {"provider": "AWS S3", "region": "eu-west-3", "encryption": "AES-256-GCM"},
        "cache": {"provider": "AWS ElastiCache", "region": "eu-west-3", "encryption": "in-transit-and-at-rest"},
        "search": {"provider": "AWS OpenSearch", "region": "eu-west-3", "encryption": "AES-256"}
    }'::jsonb,
    ARRAY['GDPR', 'Loi Informatique et Libertes', 'SecNumCloud', 'HDS (Health Data Hosting)', 'NIS2', 'ANSSI Guidelines'],
    'GDPR Article 6(1)(b) with French Loi Informatique et Libertes requirements',
    'Commission Nationale de l''Informatique et des Libertes (CNIL)',
    'https://www.cnil.fr/en/home',
    ARRAY['AD','AR','CA','FO','GG','IL','IM','JP','JE','NZ','CH','UY','KR','GB','US','AT','BE','BG','HR','CY','CZ','DK','EE','FI','DE','GR','HU','IE','IT','LV','LT','LU','MT','NL','PL','PT','RO','SK','SI','ES','SE'],
    'enforce',
    false,
    true,
    true
),

-- ============================================================
-- GLOBAL (No geographic restriction)
-- ============================================================
(
    'global',
    'Global',
    'No geographic data residency restrictions. Data may be stored in any available region. Suitable for organisations without specific data localisation requirements.',
    '{}',
    '{}',
    '{
        "aws": ["us-east-1", "us-west-2", "eu-west-1", "eu-central-1", "ap-southeast-1", "ap-northeast-1"],
        "azure": ["eastus", "westus2", "westeurope", "southeastasia", "japaneast"],
        "gcp": ["us-central1", "us-east1", "europe-west1", "asia-southeast1", "asia-northeast1"]
    }'::jsonb,
    'eu-west-1',
    'us-east-1',
    '{
        "database": {"provider": "AWS RDS", "region": "eu-west-1", "encryption": "AES-256", "backup_region": "us-east-1"},
        "files": {"provider": "AWS S3", "region": "eu-west-1", "encryption": "AES-256-GCM"},
        "cache": {"provider": "AWS ElastiCache", "region": "eu-west-1", "encryption": "in-transit-and-at-rest"},
        "search": {"provider": "AWS OpenSearch", "region": "eu-west-1", "encryption": "AES-256"}
    }'::jsonb,
    ARRAY['ISO 27001', 'SOC 2 Type II'],
    'Contractual agreement - Data Processing Agreement (DPA)',
    'Varies by jurisdiction',
    '',
    ARRAY['AD','AR','CA','FO','GG','IL','IM','JP','JE','NZ','CH','UY','KR','GB','US'],
    'disabled',
    true,
    true,
    true
)

ON CONFLICT (region) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    description = EXCLUDED.description,
    allowed_countries = EXCLUDED.allowed_countries,
    blocked_countries = EXCLUDED.blocked_countries,
    allowed_cloud_regions = EXCLUDED.allowed_cloud_regions,
    primary_cloud_region = EXCLUDED.primary_cloud_region,
    failover_cloud_region = EXCLUDED.failover_cloud_region,
    storage_config = EXCLUDED.storage_config,
    compliance_frameworks = EXCLUDED.compliance_frameworks,
    legal_basis = EXCLUDED.legal_basis,
    data_protection_authority = EXCLUDED.data_protection_authority,
    dpa_contact_url = EXCLUDED.dpa_contact_url,
    gdpr_adequate_countries = EXCLUDED.gdpr_adequate_countries,
    enforcement_mode = EXCLUDED.enforcement_mode,
    allow_cross_region_search = EXCLUDED.allow_cross_region_search,
    allow_cross_region_backup = EXCLUDED.allow_cross_region_backup,
    is_active = EXCLUDED.is_active,
    updated_at = NOW();
