-- ============================================================
-- ComplianceForge GRC Platform
-- Seed 023: Regulatory Sources for Horizon Scanning
-- 22 regulatory sources across UK, EU, Germany, France, and
-- international standards / industry bodies.
-- ============================================================

-- ============================================================
-- UK SOURCES
-- ============================================================

INSERT INTO regulatory_sources (id, name, source_type, country_code, region, url, rss_feed_url, relevance_frameworks, scan_frequency) VALUES

-- UK: Information Commissioner's Office (ICO)
('a0000000-0000-0000-0023-000000000001'::uuid,
 'Information Commissioner''s Office (ICO)',
 'supervisory_authority', 'GB', 'Europe',
 'https://ico.org.uk',
 'https://ico.org.uk/about-the-ico/media-centre/news-and-blogs/rss/',
 ARRAY['UK_GDPR','ISO27001','CYBER_ESSENTIALS'],
 'daily'),

-- UK: National Cyber Security Centre (NCSC)
('a0000000-0000-0000-0023-000000000002'::uuid,
 'National Cyber Security Centre (NCSC)',
 'government', 'GB', 'Europe',
 'https://www.ncsc.gov.uk',
 'https://www.ncsc.gov.uk/api/1/services/v1/all-rss-feed.xml',
 ARRAY['NCSC_CAF','CYBER_ESSENTIALS','NIST_CSF_2','ISO27001'],
 'daily'),

-- UK: Financial Conduct Authority (FCA)
('a0000000-0000-0000-0023-000000000003'::uuid,
 'Financial Conduct Authority (FCA)',
 'supervisory_authority', 'GB', 'Europe',
 'https://www.fca.org.uk',
 'https://www.fca.org.uk/news/rss.xml',
 ARRAY['ISO27001','PCI_DSS_4','UK_GDPR'],
 'daily'),

-- UK: Prudential Regulation Authority (PRA)
('a0000000-0000-0000-0023-000000000004'::uuid,
 'Prudential Regulation Authority (PRA)',
 'supervisory_authority', 'GB', 'Europe',
 'https://www.bankofengland.co.uk/prudential-regulation',
 'https://www.bankofengland.co.uk/rss/publications',
 ARRAY['ISO27001','NIST_800_53'],
 'daily'),

-- UK: Bank of England
('a0000000-0000-0000-0023-000000000005'::uuid,
 'Bank of England',
 'supervisory_authority', 'GB', 'Europe',
 'https://www.bankofengland.co.uk',
 'https://www.bankofengland.co.uk/rss/news',
 ARRAY['ISO27001','NIST_800_53'],
 'weekly');

-- ============================================================
-- EU SOURCES
-- ============================================================

INSERT INTO regulatory_sources (id, name, source_type, country_code, region, url, rss_feed_url, relevance_frameworks, scan_frequency) VALUES

-- EU: ENISA (EU Agency for Cybersecurity)
('a0000000-0000-0000-0023-000000000006'::uuid,
 'European Union Agency for Cybersecurity (ENISA)',
 'government', 'EU', 'Europe',
 'https://www.enisa.europa.eu',
 'https://www.enisa.europa.eu/rss.xml',
 ARRAY['ISO27001','NIST_CSF_2','NCSC_CAF','NIST_800_53'],
 'daily'),

-- EU: European Data Protection Board (EDPB)
('a0000000-0000-0000-0023-000000000007'::uuid,
 'European Data Protection Board (EDPB)',
 'supervisory_authority', 'EU', 'Europe',
 'https://edpb.europa.eu',
 'https://edpb.europa.eu/rss_en',
 ARRAY['UK_GDPR','ISO27001'],
 'daily'),

-- EU: European Commission (Digital Policy)
('a0000000-0000-0000-0023-000000000008'::uuid,
 'European Commission — Digital Policy',
 'government', 'EU', 'Europe',
 'https://digital-strategy.ec.europa.eu',
 'https://digital-strategy.ec.europa.eu/en/rss.xml',
 ARRAY['ISO27001','UK_GDPR','NIST_CSF_2'],
 'weekly');

-- ============================================================
-- GERMANY SOURCES
-- ============================================================

INSERT INTO regulatory_sources (id, name, source_type, country_code, region, url, rss_feed_url, relevance_frameworks, scan_frequency) VALUES

-- Germany: BSI (Federal Office for Information Security)
('a0000000-0000-0000-0023-000000000009'::uuid,
 'Bundesamt fur Sicherheit in der Informationstechnik (BSI)',
 'government', 'DE', 'Europe',
 'https://www.bsi.bund.de',
 'https://www.bsi.bund.de/SiteGlobals/Functions/RSSFeed/RSSNewsfeed/RSSNewsfeed.xml',
 ARRAY['ISO27001','NIST_CSF_2','NCSC_CAF','CYBER_ESSENTIALS'],
 'daily'),

-- Germany: BfDI (Federal Commissioner for Data Protection)
('a0000000-0000-0000-0023-000000000010'::uuid,
 'Bundesbeauftragter fur den Datenschutz (BfDI)',
 'supervisory_authority', 'DE', 'Europe',
 'https://www.bfdi.bund.de',
 'https://www.bfdi.bund.de/SiteGlobals/Functions/RSSFeed/RSSNewsfeed/RSSNewsfeed.xml',
 ARRAY['UK_GDPR','ISO27001'],
 'daily');

-- ============================================================
-- FRANCE SOURCES
-- ============================================================

INSERT INTO regulatory_sources (id, name, source_type, country_code, region, url, rss_feed_url, relevance_frameworks, scan_frequency) VALUES

-- France: ANSSI (National Cybersecurity Agency)
('a0000000-0000-0000-0023-000000000011'::uuid,
 'Agence nationale de la securite des systemes d''information (ANSSI)',
 'government', 'FR', 'Europe',
 'https://www.ssi.gouv.fr',
 'https://www.cert.ssi.gouv.fr/feed/',
 ARRAY['ISO27001','NIST_CSF_2','NCSC_CAF'],
 'daily'),

-- France: CNIL (Data Protection Authority)
('a0000000-0000-0000-0023-000000000012'::uuid,
 'Commission nationale de l''informatique et des libertes (CNIL)',
 'supervisory_authority', 'FR', 'Europe',
 'https://www.cnil.fr',
 'https://www.cnil.fr/fr/rss.xml',
 ARRAY['UK_GDPR','ISO27001'],
 'daily');

-- ============================================================
-- INTERNATIONAL STANDARDS BODIES
-- ============================================================

INSERT INTO regulatory_sources (id, name, source_type, country_code, region, url, rss_feed_url, relevance_frameworks, scan_frequency) VALUES

-- ISO (International Organization for Standardization)
('a0000000-0000-0000-0023-000000000013'::uuid,
 'International Organization for Standardization (ISO)',
 'standards_body', 'INT', 'Global',
 'https://www.iso.org',
 'https://www.iso.org/cms/render/live/en/sites/isoorg/home/news_index.xml',
 ARRAY['ISO27001','COBIT_2019'],
 'weekly'),

-- NIST (National Institute of Standards and Technology)
('a0000000-0000-0000-0023-000000000014'::uuid,
 'National Institute of Standards and Technology (NIST)',
 'standards_body', 'US', 'North America',
 'https://www.nist.gov',
 'https://www.nist.gov/news-events/news/rss.xml',
 ARRAY['NIST_CSF_2','NIST_800_53','ISO27001'],
 'daily'),

-- PCI Security Standards Council (PCI SSC)
('a0000000-0000-0000-0023-000000000015'::uuid,
 'PCI Security Standards Council (PCI SSC)',
 'standards_body', 'INT', 'Global',
 'https://www.pcisecuritystandards.org',
 'https://blog.pcisecuritystandards.org/rss.xml',
 ARRAY['PCI_DSS_4','ISO27001','NIST_800_53'],
 'daily'),

-- ISACA
('a0000000-0000-0000-0023-000000000016'::uuid,
 'Information Systems Audit and Control Association (ISACA)',
 'industry_body', 'INT', 'Global',
 'https://www.isaca.org',
 'https://www.isaca.org/resources/news-and-trends/rss',
 ARRAY['COBIT_2019','ISO27001','NIST_CSF_2'],
 'weekly'),

-- AXELOS (ITIL)
('a0000000-0000-0000-0023-000000000017'::uuid,
 'AXELOS (ITIL / PRINCE2)',
 'industry_body', 'GB', 'Global',
 'https://www.axelos.com',
 'https://www.axelos.com/rss/news',
 ARRAY['ITIL_4','COBIT_2019','ISO27001'],
 'weekly');

-- ============================================================
-- ADDITIONAL EUROPEAN & INDUSTRY SOURCES
-- ============================================================

INSERT INTO regulatory_sources (id, name, source_type, country_code, region, url, rss_feed_url, relevance_frameworks, scan_frequency) VALUES

-- EBA (European Banking Authority)
('a0000000-0000-0000-0023-000000000018'::uuid,
 'European Banking Authority (EBA)',
 'supervisory_authority', 'EU', 'Europe',
 'https://www.eba.europa.eu',
 'https://www.eba.europa.eu/rss',
 ARRAY['ISO27001','NIST_800_53'],
 'daily'),

-- EIOPA (Insurance and Pensions)
('a0000000-0000-0000-0023-000000000019'::uuid,
 'European Insurance and Occupational Pensions Authority (EIOPA)',
 'supervisory_authority', 'EU', 'Europe',
 'https://www.eiopa.europa.eu',
 'https://www.eiopa.europa.eu/rss_en',
 ARRAY['ISO27001'],
 'weekly'),

-- ESMA (Securities and Markets)
('a0000000-0000-0000-0023-000000000020'::uuid,
 'European Securities and Markets Authority (ESMA)',
 'supervisory_authority', 'EU', 'Europe',
 'https://www.esma.europa.eu',
 'https://www.esma.europa.eu/rss',
 ARRAY['ISO27001','NIST_800_53'],
 'daily'),

-- CSA (Cloud Security Alliance)
('a0000000-0000-0000-0023-000000000021'::uuid,
 'Cloud Security Alliance (CSA)',
 'industry_body', 'INT', 'Global',
 'https://cloudsecurityalliance.org',
 'https://cloudsecurityalliance.org/feed/',
 ARRAY['ISO27001','NIST_CSF_2','NIST_800_53'],
 'weekly'),

-- OWASP Foundation
('a0000000-0000-0000-0023-000000000022'::uuid,
 'OWASP Foundation',
 'industry_body', 'INT', 'Global',
 'https://owasp.org',
 'https://owasp.org/feed.xml',
 ARRAY['ISO27001','NIST_CSF_2','PCI_DSS_4'],
 'weekly');
