-- ============================================================
-- ComplianceForge — Seed: NIST CSF 2.0 ↔ NIST SP 800-53 Rev 5
-- Direct mappings between the two NIST frameworks
-- ============================================================

-- GV.PO-01 Policy established → AC-1 Access Control Policy
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'related'::mapping_type, 0.70, 'CSF policy establishment relates to 800-53 AC policy', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'GV.PO-01' AND s.framework_id = 'a0000001-0000-0000-0000-000000000006'
  AND t.code = 'AC-1' AND t.framework_id = 'a0000001-0000-0000-0000-000000000005';

-- GV.RM-01 Risk mgmt objectives → RA-1 Risk Assessment Policy
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.85, 'Both establish risk management policy', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'GV.RM-01' AND s.framework_id = 'a0000001-0000-0000-0000-000000000006'
  AND t.code = 'RA-1' AND t.framework_id = 'a0000001-0000-0000-0000-000000000005';

-- ID.AM-01 Physical devices inventoried → CM-8 Component Inventory
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.90, 'Both require hardware inventory', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'ID.AM-01' AND s.framework_id = 'a0000001-0000-0000-0000-000000000006'
  AND t.code = 'CM-8' AND t.framework_id = 'a0000001-0000-0000-0000-000000000005';

-- ID.RA-01 Vulnerabilities identified → RA-5 Vulnerability Scanning
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.90, 'Both address vulnerability identification', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'ID.RA-01' AND s.framework_id = 'a0000001-0000-0000-0000-000000000006'
  AND t.code = 'RA-5' AND t.framework_id = 'a0000001-0000-0000-0000-000000000005';

-- ID.RA-05 Risk determination → RA-3 Risk Assessment
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.90, 'Both require risk assessment', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'ID.RA-05' AND s.framework_id = 'a0000001-0000-0000-0000-000000000006'
  AND t.code = 'RA-3' AND t.framework_id = 'a0000001-0000-0000-0000-000000000005';

-- PR.AA-01 Identity management → IA-4 Identifier Management
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.90, 'Both address identity management', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'PR.AA-01' AND s.framework_id = 'a0000001-0000-0000-0000-000000000006'
  AND t.code = 'IA-4' AND t.framework_id = 'a0000001-0000-0000-0000-000000000005';

-- PR.AA-03 Authentication → IA-2 User Identification and Authentication
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.95, 'Both require user authentication', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'PR.AA-03' AND s.framework_id = 'a0000001-0000-0000-0000-000000000006'
  AND t.code = 'IA-2' AND t.framework_id = 'a0000001-0000-0000-0000-000000000005';

-- PR.AA-05 Access permissions → AC-3 Access Enforcement
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.90, 'Both enforce access control', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'PR.AA-05' AND s.framework_id = 'a0000001-0000-0000-0000-000000000006'
  AND t.code = 'AC-3' AND t.framework_id = 'a0000001-0000-0000-0000-000000000005';

-- PR.AT-01 Security training → AT-2 Awareness Training
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.95, 'Both require security awareness training', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'PR.AT-01' AND s.framework_id = 'a0000001-0000-0000-0000-000000000006'
  AND t.code = 'AT-2' AND t.framework_id = 'a0000001-0000-0000-0000-000000000005';

-- PR.DS-01 Data at rest protected → SC-28 Protection of Information at Rest
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.95, 'Both address data at rest protection', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'PR.DS-01' AND s.framework_id = 'a0000001-0000-0000-0000-000000000006'
  AND t.code = 'SC-28' AND t.framework_id = 'a0000001-0000-0000-0000-000000000005';

-- PR.DS-02 Data in transit → SC-8 Transmission Confidentiality
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.95, 'Both address data in transit protection', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'PR.DS-02' AND s.framework_id = 'a0000001-0000-0000-0000-000000000006'
  AND t.code = 'SC-8' AND t.framework_id = 'a0000001-0000-0000-0000-000000000005';

-- PR.DS-11 Backups → CP-9 System Backup
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.95, 'Both require backup management', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'PR.DS-11' AND s.framework_id = 'a0000001-0000-0000-0000-000000000006'
  AND t.code = 'CP-9' AND t.framework_id = 'a0000001-0000-0000-0000-000000000005';

-- PR.PS-01 Configuration management → CM-2 Baseline Configuration
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.85, 'Both address configuration management', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'PR.PS-01' AND s.framework_id = 'a0000001-0000-0000-0000-000000000006'
  AND t.code = 'CM-2' AND t.framework_id = 'a0000001-0000-0000-0000-000000000005';

-- PR.PS-04 Log records → AU-2 Event Logging
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.90, 'Both require logging', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'PR.PS-04' AND s.framework_id = 'a0000001-0000-0000-0000-000000000006'
  AND t.code = 'AU-2' AND t.framework_id = 'a0000001-0000-0000-0000-000000000005';

-- PR.PS-06 Secure dev practices → SA-3 System Development Life Cycle
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.85, 'Both address secure SDLC', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'PR.PS-06' AND s.framework_id = 'a0000001-0000-0000-0000-000000000006'
  AND t.code = 'SA-3' AND t.framework_id = 'a0000001-0000-0000-0000-000000000005';

-- PR.IR-01 Network access protection → SC-7 Boundary Protection
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.90, 'Both address network boundary protection', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'PR.IR-01' AND s.framework_id = 'a0000001-0000-0000-0000-000000000006'
  AND t.code = 'SC-7' AND t.framework_id = 'a0000001-0000-0000-0000-000000000005';

-- DE.CM-01 Network monitoring → SI-4 System Monitoring
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.90, 'Both address monitoring', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'DE.CM-01' AND s.framework_id = 'a0000001-0000-0000-0000-000000000006'
  AND t.code = 'SI-4' AND t.framework_id = 'a0000001-0000-0000-0000-000000000005';

-- RS.MA-01 Incident response plan → IR-8 Incident Response Plan
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.95, 'Both require incident response planning', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'RS.MA-01' AND s.framework_id = 'a0000001-0000-0000-0000-000000000006'
  AND t.code = 'IR-8' AND t.framework_id = 'a0000001-0000-0000-0000-000000000005';

-- RS.MI-01 Incidents contained → IR-4 Incident Handling
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.85, 'Both address incident containment', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'RS.MI-01' AND s.framework_id = 'a0000001-0000-0000-0000-000000000006'
  AND t.code = 'IR-4' AND t.framework_id = 'a0000001-0000-0000-0000-000000000005';

-- RC.RP-01 Recovery plan → CP-10 System Recovery
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.90, 'Both address recovery operations', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'RC.RP-01' AND s.framework_id = 'a0000001-0000-0000-0000-000000000006'
  AND t.code = 'CP-10' AND t.framework_id = 'a0000001-0000-0000-0000-000000000005';

-- GV.SC-01 Supply chain risk → SR-2 Supply Chain Risk Mgmt Plan
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.90, 'Both address supply chain risk management', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'GV.SC-01' AND s.framework_id = 'a0000001-0000-0000-0000-000000000006'
  AND t.code = 'SR-2' AND t.framework_id = 'a0000001-0000-0000-0000-000000000005';
