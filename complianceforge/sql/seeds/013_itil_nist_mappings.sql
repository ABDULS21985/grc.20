-- ============================================================
-- ComplianceForge — Seed: ITIL4/NIST800-53 Cross-Framework Mappings
-- ============================================================

-- ============================================================
-- ITIL 4 → ISO 27001:2022
-- ============================================================

-- ITIL GP03 Info Security Mgmt → ISO A.5.1 Policies
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.85, 'Both establish information security management', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'ITIL-GP03' AND s.framework_id = 'a0000001-0000-0000-0000-000000000008'
  AND t.code = 'A.5.1' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- ITIL GP10 Risk Mgmt → ISO A.5.7 Threat intelligence (related)
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'related'::mapping_type, 0.60, 'ITIL risk mgmt broader; ISO specific to threat intel', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'ITIL-GP10' AND s.framework_id = 'a0000001-0000-0000-0000-000000000008'
  AND t.code = 'A.5.7' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- ITIL GP13 Supplier Mgmt → ISO A.5.19 Supplier relationships
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.80, 'Both address supplier/vendor management', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'ITIL-GP13' AND s.framework_id = 'a0000001-0000-0000-0000-000000000008'
  AND t.code = 'A.5.19' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- ITIL SM04 Change Enablement → ISO A.8.32 Change management
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.90, 'Both address change management procedures', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'ITIL-SM04' AND s.framework_id = 'a0000001-0000-0000-0000-000000000008'
  AND t.code = 'A.8.32' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- ITIL SM05 Incident Mgmt → ISO A.5.24 Incident planning
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.85, 'Both address incident management', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'ITIL-SM05' AND s.framework_id = 'a0000001-0000-0000-0000-000000000008'
  AND t.code = 'A.5.24' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- ITIL SM06 IT Asset Mgmt → ISO A.5.9 Asset inventory
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.85, 'Both require asset inventory management', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'ITIL-SM06' AND s.framework_id = 'a0000001-0000-0000-0000-000000000008'
  AND t.code = 'A.5.9' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- ITIL SM07 Monitoring → ISO A.8.16 Monitoring activities
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.85, 'Both address monitoring and event management', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'ITIL-SM07' AND s.framework_id = 'a0000001-0000-0000-0000-000000000008'
  AND t.code = 'A.8.16' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- ITIL SM11 Service Configuration → ISO A.8.9 Configuration management
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.85, 'Both address configuration management', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'ITIL-SM11' AND s.framework_id = 'a0000001-0000-0000-0000-000000000008'
  AND t.code = 'A.8.9' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- ITIL SM12 Service Continuity → ISO A.5.29 Info sec during disruption
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.85, 'Both address business/service continuity', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'ITIL-SM12' AND s.framework_id = 'a0000001-0000-0000-0000-000000000008'
  AND t.code = 'A.5.29' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- ITIL TM03 Software Dev → ISO A.8.25 Secure development
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'partial'::mapping_type, 0.70, 'ITIL covers SW dev broadly; ISO specific to secure dev', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'ITIL-TM03' AND s.framework_id = 'a0000001-0000-0000-0000-000000000008'
  AND t.code = 'A.8.25' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- ITIL GP14 Workforce Mgmt → ISO A.6.1 Screening
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'partial'::mapping_type, 0.60, 'ITIL broader workforce; ISO specific to screening', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'ITIL-GP14' AND s.framework_id = 'a0000001-0000-0000-0000-000000000008'
  AND t.code = 'A.6.1' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- ITIL GP02 Continual Improvement → ISO A.5.36 Compliance with policies
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'related'::mapping_type, 0.55, 'ITIL continual improvement relates to ISO compliance review', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'ITIL-GP02' AND s.framework_id = 'a0000001-0000-0000-0000-000000000008'
  AND t.code = 'A.5.36' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- ============================================================
-- NIST 800-53 → ISO 27001:2022
-- ============================================================

-- AC-2 Account Management → A.5.16 Identity management
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.90, 'Both address user account/identity management', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'AC-2' AND s.framework_id = 'a0000001-0000-0000-0000-000000000005'
  AND t.code = 'A.5.16' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- AC-3 Access Enforcement → A.8.3 Information access restriction
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.90, 'Both enforce access control', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'AC-3' AND s.framework_id = 'a0000001-0000-0000-0000-000000000005'
  AND t.code = 'A.8.3' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- AC-6 Least Privilege → A.8.2 Privileged access rights
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.90, 'Both address principle of least privilege', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'AC-6' AND s.framework_id = 'a0000001-0000-0000-0000-000000000005'
  AND t.code = 'A.8.2' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- AT-2 Awareness Training → A.6.3 Awareness training
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.95, 'Both require security awareness training', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'AT-2' AND s.framework_id = 'a0000001-0000-0000-0000-000000000005'
  AND t.code = 'A.6.3' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- AU-2 Event Logging → A.8.15 Logging
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.90, 'Both require audit event logging', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'AU-2' AND s.framework_id = 'a0000001-0000-0000-0000-000000000005'
  AND t.code = 'A.8.15' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- CM-2 Baseline Configuration → A.8.9 Configuration management
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.85, 'Both address secure baseline configuration', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'CM-2' AND s.framework_id = 'a0000001-0000-0000-0000-000000000005'
  AND t.code = 'A.8.9' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- CP-9 System Backup → A.8.13 Information backup
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.90, 'Both require system/data backup', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'CP-9' AND s.framework_id = 'a0000001-0000-0000-0000-000000000005'
  AND t.code = 'A.8.13' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- IA-2 User I&A → A.8.5 Secure authentication
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.90, 'Both address user authentication', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'IA-2' AND s.framework_id = 'a0000001-0000-0000-0000-000000000005'
  AND t.code = 'A.8.5' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- IR-4 Incident Handling → A.5.26 Response to incidents
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.90, 'Both address incident response handling', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'IR-4' AND s.framework_id = 'a0000001-0000-0000-0000-000000000005'
  AND t.code = 'A.5.26' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- PE-3 Physical Access Control → A.7.2 Physical entry
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.90, 'Both address physical access control', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'PE-3' AND s.framework_id = 'a0000001-0000-0000-0000-000000000005'
  AND t.code = 'A.7.2' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- RA-5 Vulnerability Scanning → A.8.8 Vulnerability management
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.90, 'Both address vulnerability scanning', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'RA-5' AND s.framework_id = 'a0000001-0000-0000-0000-000000000005'
  AND t.code = 'A.8.8' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- SC-7 Boundary Protection → A.8.20 Networks security
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.85, 'Both address network boundary protection', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'SC-7' AND s.framework_id = 'a0000001-0000-0000-0000-000000000005'
  AND t.code = 'A.8.20' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- SC-28 Protection at Rest → A.8.24 Use of cryptography
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'partial'::mapping_type, 0.75, 'NIST specific to data at rest; ISO broader crypto', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'SC-28' AND s.framework_id = 'a0000001-0000-0000-0000-000000000005'
  AND t.code = 'A.8.24' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- SI-2 Flaw Remediation → A.8.8 Vulnerability management
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.85, 'Both address flaw/vulnerability remediation', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'SI-2' AND s.framework_id = 'a0000001-0000-0000-0000-000000000005'
  AND t.code = 'A.8.8' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- SI-3 Malicious Code Protection → A.8.7 Malware protection
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.95, 'Both require malware/malicious code protection', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'SI-3' AND s.framework_id = 'a0000001-0000-0000-0000-000000000005'
  AND t.code = 'A.8.7' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- SI-4 System Monitoring → A.8.16 Monitoring activities
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.90, 'Both require system monitoring', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'SI-4' AND s.framework_id = 'a0000001-0000-0000-0000-000000000005'
  AND t.code = 'A.8.16' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- SR-2 Supply Chain Risk Mgmt Plan → A.5.21 ICT supply chain
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.85, 'Both address supply chain risk management', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'SR-2' AND s.framework_id = 'a0000001-0000-0000-0000-000000000005'
  AND t.code = 'A.5.21' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- PS-3 Personnel Screening → A.6.1 Screening
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.95, 'Both address pre-employment screening', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'PS-3' AND s.framework_id = 'a0000001-0000-0000-0000-000000000005'
  AND t.code = 'A.6.1' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- CA-8 Penetration Testing → A.8.29 Security testing
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.90, 'Both address penetration/security testing', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'CA-8' AND s.framework_id = 'a0000001-0000-0000-0000-000000000005'
  AND t.code = 'A.8.29' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- MP-6 Media Sanitisation → A.7.14 Secure disposal
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.90, 'Both address secure media disposal', true
FROM framework_controls s, framework_controls t
WHERE s.code = 'MP-6' AND s.framework_id = 'a0000001-0000-0000-0000-000000000005'
  AND t.code = 'A.7.14' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';
