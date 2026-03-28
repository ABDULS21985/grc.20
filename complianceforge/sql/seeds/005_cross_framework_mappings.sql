-- ============================================================
-- ComplianceForge — Seed: Cross-Framework Control Mappings
-- Maps ISO 27001:2022 controls to NIST CSF 2.0 and Cyber Essentials
-- NOTE: These use subqueries to resolve UUIDs from control codes
-- ============================================================

-- ============================================================
-- ISO 27001 → NIST CSF 2.0 Mappings
-- ============================================================

-- A.5.1 Policies → GV.PO-01 Policy established
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.95, 'Both require establishment of security policies', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'A.5.1' AND s.framework_id = 'a0000001-0000-0000-0000-000000000001'
  AND t.code = 'GV.PO-01' AND t.framework_id = 'a0000001-0000-0000-0000-000000000006';

-- A.5.2 Roles → GV.RR-02 Roles established
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.90, 'Both require defining security roles and responsibilities', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'A.5.2' AND s.framework_id = 'a0000001-0000-0000-0000-000000000001'
  AND t.code = 'GV.RR-02' AND t.framework_id = 'a0000001-0000-0000-0000-000000000006';

-- A.5.7 Threat intelligence → ID.RA-02 Threat intelligence received
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.90, 'Both address threat intelligence gathering', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'A.5.7' AND s.framework_id = 'a0000001-0000-0000-0000-000000000001'
  AND t.code = 'ID.RA-02' AND t.framework_id = 'a0000001-0000-0000-0000-000000000006';

-- A.5.9 Asset inventory → ID.AM-01 Physical devices inventoried
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'partial'::mapping_type, 0.80, 'ISO covers all assets; NIST CSF focuses on physical devices', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'A.5.9' AND s.framework_id = 'a0000001-0000-0000-0000-000000000001'
  AND t.code = 'ID.AM-01' AND t.framework_id = 'a0000001-0000-0000-0000-000000000006';

-- A.5.15 Access control → PR.AA-05 Access permissions managed
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.90, 'Both address access control management', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'A.5.15' AND s.framework_id = 'a0000001-0000-0000-0000-000000000001'
  AND t.code = 'PR.AA-05' AND t.framework_id = 'a0000001-0000-0000-0000-000000000006';

-- A.5.24 Incident planning → RS.MA-01 Incident response plan
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.90, 'Both require incident response planning', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'A.5.24' AND s.framework_id = 'a0000001-0000-0000-0000-000000000001'
  AND t.code = 'RS.MA-01' AND t.framework_id = 'a0000001-0000-0000-0000-000000000006';

-- A.5.29 Business continuity → RC.RP-01 Recovery plan executed
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'related'::mapping_type, 0.75, 'ISO covers BCP; NIST CSF covers recovery execution', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'A.5.29' AND s.framework_id = 'a0000001-0000-0000-0000-000000000001'
  AND t.code = 'RC.RP-01' AND t.framework_id = 'a0000001-0000-0000-0000-000000000006';

-- A.6.3 Training → PR.AT-01 Security awareness training
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.95, 'Both require security awareness and training', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'A.6.3' AND s.framework_id = 'a0000001-0000-0000-0000-000000000001'
  AND t.code = 'PR.AT-01' AND t.framework_id = 'a0000001-0000-0000-0000-000000000006';

-- A.8.2 Privileged access → PR.AA-05 Access permissions managed
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'partial'::mapping_type, 0.80, 'ISO specific to privileged access; NIST broader', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'A.8.2' AND s.framework_id = 'a0000001-0000-0000-0000-000000000001'
  AND t.code = 'PR.AA-05' AND t.framework_id = 'a0000001-0000-0000-0000-000000000006';

-- A.8.5 Secure auth → PR.AA-03 Authentication
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.90, 'Both address secure authentication', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'A.8.5' AND s.framework_id = 'a0000001-0000-0000-0000-000000000001'
  AND t.code = 'PR.AA-03' AND t.framework_id = 'a0000001-0000-0000-0000-000000000006';

-- A.8.7 Malware protection → PR.PS-05 Software installation restricted
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'related'::mapping_type, 0.70, 'ISO covers malware; NIST CSF covers software restriction', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'A.8.7' AND s.framework_id = 'a0000001-0000-0000-0000-000000000001'
  AND t.code = 'PR.PS-05' AND t.framework_id = 'a0000001-0000-0000-0000-000000000006';

-- A.8.8 Vulnerability mgmt → ID.RA-01 Vulnerabilities identified
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.90, 'Both address vulnerability management', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'A.8.8' AND s.framework_id = 'a0000001-0000-0000-0000-000000000001'
  AND t.code = 'ID.RA-01' AND t.framework_id = 'a0000001-0000-0000-0000-000000000006';

-- A.8.9 Configuration mgmt → PR.PS-01 Configuration management
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.95, 'Both address configuration management', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'A.8.9' AND s.framework_id = 'a0000001-0000-0000-0000-000000000001'
  AND t.code = 'PR.PS-01' AND t.framework_id = 'a0000001-0000-0000-0000-000000000006';

-- A.8.13 Backup → PR.DS-11 Backups maintained
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.95, 'Both address data backup', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'A.8.13' AND s.framework_id = 'a0000001-0000-0000-0000-000000000001'
  AND t.code = 'PR.DS-11' AND t.framework_id = 'a0000001-0000-0000-0000-000000000006';

-- A.8.15 Logging → PR.PS-04 Log records generated
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.90, 'Both address logging requirements', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'A.8.15' AND s.framework_id = 'a0000001-0000-0000-0000-000000000001'
  AND t.code = 'PR.PS-04' AND t.framework_id = 'a0000001-0000-0000-0000-000000000006';

-- A.8.16 Monitoring → DE.CM-01 Network monitoring
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.85, 'Both address monitoring activities', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'A.8.16' AND s.framework_id = 'a0000001-0000-0000-0000-000000000001'
  AND t.code = 'DE.CM-01' AND t.framework_id = 'a0000001-0000-0000-0000-000000000006';

-- A.8.20 Network security → PR.IR-01 Network access protection
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.90, 'Both address network security', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'A.8.20' AND s.framework_id = 'a0000001-0000-0000-0000-000000000001'
  AND t.code = 'PR.IR-01' AND t.framework_id = 'a0000001-0000-0000-0000-000000000006';

-- A.8.24 Cryptography → PR.DS-01 Data at rest protected
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'partial'::mapping_type, 0.75, 'ISO covers all crypto; NIST CSF here focuses on data at rest', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'A.8.24' AND s.framework_id = 'a0000001-0000-0000-0000-000000000001'
  AND t.code = 'PR.DS-01' AND t.framework_id = 'a0000001-0000-0000-0000-000000000006';

-- A.8.25 SDLC → PR.PS-06 Secure development practices
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.90, 'Both address secure development lifecycle', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'A.8.25' AND s.framework_id = 'a0000001-0000-0000-0000-000000000001'
  AND t.code = 'PR.PS-06' AND t.framework_id = 'a0000001-0000-0000-0000-000000000006';

-- ============================================================
-- ISO 27001 → Cyber Essentials Mappings
-- ============================================================

-- A.8.20 Network security → CE-FW-01 Firewalls
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'superset'::mapping_type, 0.70, 'ISO network security is broader; CE focuses on firewalls', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'A.8.20' AND s.framework_id = 'a0000001-0000-0000-0000-000000000001'
  AND t.code = 'CE-FW-01' AND t.framework_id = 'a0000001-0000-0000-0000-000000000004';

-- A.8.9 Config mgmt → CE-SC-01 Secure Configuration
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'superset'::mapping_type, 0.75, 'ISO config mgmt is broader; CE is specific to hardening', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'A.8.9' AND s.framework_id = 'a0000001-0000-0000-0000-000000000001'
  AND t.code = 'CE-SC-01' AND t.framework_id = 'a0000001-0000-0000-0000-000000000004';

-- A.8.8 Vulnerability mgmt → CE-SU-01 Security Update Management
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'superset'::mapping_type, 0.70, 'ISO vuln mgmt is broader; CE focuses on patching', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'A.8.8' AND s.framework_id = 'a0000001-0000-0000-0000-000000000001'
  AND t.code = 'CE-SU-01' AND t.framework_id = 'a0000001-0000-0000-0000-000000000004';

-- A.5.15 Access control → CE-AC-01 User Access Control
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'superset'::mapping_type, 0.75, 'ISO access control is broader; CE focuses on user accounts', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'A.5.15' AND s.framework_id = 'a0000001-0000-0000-0000-000000000001'
  AND t.code = 'CE-AC-01' AND t.framework_id = 'a0000001-0000-0000-0000-000000000004';

-- A.8.7 Malware → CE-MP-01 Malware Protection
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.85, 'Both address malware protection', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'A.8.7' AND s.framework_id = 'a0000001-0000-0000-0000-000000000001'
  AND t.code = 'CE-MP-01' AND t.framework_id = 'a0000001-0000-0000-0000-000000000004';
