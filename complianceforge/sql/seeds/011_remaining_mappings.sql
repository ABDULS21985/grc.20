-- ============================================================
-- ComplianceForge — Seed: Additional Cross-Framework Mappings
-- UK GDPR ↔ ISO 27001, COBIT ↔ ISO 27001, NCSC CAF ↔ ISO 27001
-- NIST 800-53 ↔ ISO 27001, ITIL 4 ↔ ISO 27001
-- ============================================================

-- ============================================================
-- UK GDPR → ISO 27001:2022
-- ============================================================

-- Art.5(1)(f) Integrity & confidentiality → A.5.1 Policies
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'related'::mapping_type, 0.70, 'GDPR security principle relies on ISO security policies', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'Art.5(1)(f)' AND s.framework_id = 'a0000001-0000-0000-0000-000000000002'
  AND t.code = 'A.5.1' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- Art.25 Privacy by design → A.5.8 Info sec in project management
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'related'::mapping_type, 0.75, 'Both require security/privacy integrated into design', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'Art.25' AND s.framework_id = 'a0000001-0000-0000-0000-000000000002'
  AND t.code = 'A.5.8' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- Art.28 Processor obligations → A.5.19 Supplier relationships
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'partial'::mapping_type, 0.80, 'GDPR processor requirements align with supplier security', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'Art.28' AND s.framework_id = 'a0000001-0000-0000-0000-000000000002'
  AND t.code = 'A.5.19' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- Art.30 ROPA → A.5.9 Asset inventory
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'related'::mapping_type, 0.65, 'ROPA maps data assets; ISO maps all information assets', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'Art.30' AND s.framework_id = 'a0000001-0000-0000-0000-000000000002'
  AND t.code = 'A.5.9' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- Art.32 Security of processing → A.8.24 Use of cryptography
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'partial'::mapping_type, 0.70, 'GDPR Art.32 mentions encryption; ISO A.8.24 covers crypto broadly', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'Art.32' AND s.framework_id = 'a0000001-0000-0000-0000-000000000002'
  AND t.code = 'A.8.24' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- Art.33 Breach notification → A.5.24 Incident management planning
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'related'::mapping_type, 0.80, 'GDPR breach notification requires incident management processes', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'Art.33' AND s.framework_id = 'a0000001-0000-0000-0000-000000000002'
  AND t.code = 'A.5.24' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- Art.35 DPIA → A.5.34 Privacy and protection of PII
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.85, 'Both address privacy impact assessment and PII protection', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'Art.35' AND s.framework_id = 'a0000001-0000-0000-0000-000000000002'
  AND t.code = 'A.5.34' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- Art.15 Right of access → A.5.12 Classification of information
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'related'::mapping_type, 0.55, 'Data subject access requires knowing what data is held (classification)', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'Art.15' AND s.framework_id = 'a0000001-0000-0000-0000-000000000002'
  AND t.code = 'A.5.12' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- ============================================================
-- NCSC CAF → ISO 27001:2022
-- ============================================================

-- A1 Governance → A.5.1 Policies
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.85, 'Both require governance and policy framework', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'A1' AND s.framework_id = 'a0000001-0000-0000-0000-000000000003'
  AND t.code = 'A.5.1' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- A3 Asset Management → A.5.9 Asset inventory
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.90, 'Both require asset inventory management', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'A3' AND s.framework_id = 'a0000001-0000-0000-0000-000000000003'
  AND t.code = 'A.5.9' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- A4 Supply Chain → A.5.21 ICT supply chain security
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.85, 'Both address supply chain security management', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'A4' AND s.framework_id = 'a0000001-0000-0000-0000-000000000003'
  AND t.code = 'A.5.21' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- B2 Identity and Access Control → A.5.15 Access control
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.90, 'Both cover identity and access management', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'B2' AND s.framework_id = 'a0000001-0000-0000-0000-000000000003'
  AND t.code = 'A.5.15' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- B3 Data Security → A.8.24 Use of cryptography
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'partial'::mapping_type, 0.70, 'CAF data security broader; ISO A.8.24 focuses on cryptography', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'B3' AND s.framework_id = 'a0000001-0000-0000-0000-000000000003'
  AND t.code = 'A.8.24' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- B6 Staff Awareness → A.6.3 Training
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.90, 'Both require security awareness and training', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'B6' AND s.framework_id = 'a0000001-0000-0000-0000-000000000003'
  AND t.code = 'A.6.3' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- C1 Security Monitoring → A.8.16 Monitoring activities
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.90, 'Both require security monitoring capabilities', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'C1' AND s.framework_id = 'a0000001-0000-0000-0000-000000000003'
  AND t.code = 'A.8.16' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- D1 Response and Recovery → A.5.26 Response to incidents
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.85, 'Both require incident response and recovery', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'D1' AND s.framework_id = 'a0000001-0000-0000-0000-000000000003'
  AND t.code = 'A.5.26' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- ============================================================
-- COBIT 2019 → ISO 27001:2022
-- ============================================================

-- APO13 Managed Security → A.5.1 Policies
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.85, 'Both address information security management', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'APO13' AND s.framework_id = 'a0000001-0000-0000-0000-000000000009'
  AND t.code = 'A.5.1' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- APO12 Managed Risk → A.5.7 Threat intelligence
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'related'::mapping_type, 0.65, 'COBIT risk management broader; ISO threat intelligence specific', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'APO12' AND s.framework_id = 'a0000001-0000-0000-0000-000000000009'
  AND t.code = 'A.5.7' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- APO10 Managed Vendors → A.5.19 Supplier relationships
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.80, 'Both address vendor/supplier management', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'APO10' AND s.framework_id = 'a0000001-0000-0000-0000-000000000009'
  AND t.code = 'A.5.19' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- DSS02 Managed Incidents → A.5.24 Incident management
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.85, 'Both address incident management processes', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'DSS02' AND s.framework_id = 'a0000001-0000-0000-0000-000000000009'
  AND t.code = 'A.5.24' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- DSS04 Managed Continuity → A.5.29 Info sec during disruption
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.85, 'Both address business continuity', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'DSS04' AND s.framework_id = 'a0000001-0000-0000-0000-000000000009'
  AND t.code = 'A.5.29' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- DSS05 Managed Security Services → A.8.7 Malware protection
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'partial'::mapping_type, 0.65, 'COBIT security services broader; ISO specific to malware', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'DSS05' AND s.framework_id = 'a0000001-0000-0000-0000-000000000009'
  AND t.code = 'A.8.7' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- BAI06 Managed IT Changes → A.8.32 Change management
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.90, 'Both address IT change management', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'BAI06' AND s.framework_id = 'a0000001-0000-0000-0000-000000000009'
  AND t.code = 'A.8.32' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- BAI09 Managed Assets → A.5.9 Asset inventory
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.85, 'Both require asset management', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'BAI09' AND s.framework_id = 'a0000001-0000-0000-0000-000000000009'
  AND t.code = 'A.5.9' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- MEA03 Managed Compliance → A.5.31 Legal/regulatory requirements
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.85, 'Both address regulatory compliance', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'MEA03' AND s.framework_id = 'a0000001-0000-0000-0000-000000000009'
  AND t.code = 'A.5.31' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';
