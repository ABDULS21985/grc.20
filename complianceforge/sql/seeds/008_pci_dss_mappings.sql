-- ============================================================
-- ComplianceForge — Seed: PCI DSS ↔ ISO 27001 & NIST CSF Mappings
-- ============================================================

-- PCI DSS 1.2.1 (network traffic) → ISO A.8.20 (network security)
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'partial'::mapping_type, 0.80, 'Both address network traffic restrictions', true
FROM framework_controls s, framework_controls t
WHERE s.code = '1.2.1' AND s.framework_id = 'a0000001-0000-0000-0000-000000000007'
  AND t.code = 'A.8.20' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- PCI DSS 2.2.1 (config standards) → ISO A.8.9 (configuration management)
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.85, 'Both require configuration standards', true
FROM framework_controls s, framework_controls t
WHERE s.code = '2.2.1' AND s.framework_id = 'a0000001-0000-0000-0000-000000000007'
  AND t.code = 'A.8.9' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- PCI DSS 3.5.1 (PAN unreadable) → ISO A.8.24 (cryptography)
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'subset'::mapping_type, 0.70, 'PCI specific to payment data; ISO covers all crypto', true
FROM framework_controls s, framework_controls t
WHERE s.code = '3.5.1' AND s.framework_id = 'a0000001-0000-0000-0000-000000000007'
  AND t.code = 'A.8.24' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- PCI DSS 4.2.1 (data in transit) → ISO A.5.14 (information transfer)
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'partial'::mapping_type, 0.75, 'PCI specific to cardholder data; ISO covers all data', true
FROM framework_controls s, framework_controls t
WHERE s.code = '4.2.1' AND s.framework_id = 'a0000001-0000-0000-0000-000000000007'
  AND t.code = 'A.5.14' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- PCI DSS 5.2.1 (anti-malware) → ISO A.8.7 (malware protection)
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.90, 'Both require anti-malware deployment', true
FROM framework_controls s, framework_controls t
WHERE s.code = '5.2.1' AND s.framework_id = 'a0000001-0000-0000-0000-000000000007'
  AND t.code = 'A.8.7' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- PCI DSS 6.2.1 (secure development) → ISO A.8.25 (SDLC)
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.85, 'Both require secure software development', true
FROM framework_controls s, framework_controls t
WHERE s.code = '6.2.1' AND s.framework_id = 'a0000001-0000-0000-0000-000000000007'
  AND t.code = 'A.8.25' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- PCI DSS 6.3.3 (patching) → ISO A.8.8 (vulnerability management)
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'partial'::mapping_type, 0.75, 'PCI requires patching within 1 month; ISO broader', true
FROM framework_controls s, framework_controls t
WHERE s.code = '6.3.3' AND s.framework_id = 'a0000001-0000-0000-0000-000000000007'
  AND t.code = 'A.8.8' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- PCI DSS 7.2.1 (access model) → ISO A.5.15 (access control)
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.85, 'Both require access control policies', true
FROM framework_controls s, framework_controls t
WHERE s.code = '7.2.1' AND s.framework_id = 'a0000001-0000-0000-0000-000000000007'
  AND t.code = 'A.5.15' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- PCI DSS 8.4.1 (MFA) → ISO A.8.5 (secure authentication)
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'partial'::mapping_type, 0.80, 'PCI requires MFA for CDE; ISO broader auth requirements', true
FROM framework_controls s, framework_controls t
WHERE s.code = '8.4.1' AND s.framework_id = 'a0000001-0000-0000-0000-000000000007'
  AND t.code = 'A.8.5' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- PCI DSS 9.2.1 (physical access) → ISO A.7.2 (physical entry)
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.85, 'Both require physical access controls', true
FROM framework_controls s, framework_controls t
WHERE s.code = '9.2.1' AND s.framework_id = 'a0000001-0000-0000-0000-000000000007'
  AND t.code = 'A.7.2' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- PCI DSS 10.2.1 (audit logs) → ISO A.8.15 (logging)
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.90, 'Both require comprehensive audit logging', true
FROM framework_controls s, framework_controls t
WHERE s.code = '10.2.1' AND s.framework_id = 'a0000001-0000-0000-0000-000000000007'
  AND t.code = 'A.8.15' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- PCI DSS 11.4.1 (pen testing) → ISO A.8.29 (security testing)
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.85, 'Both require security testing', true
FROM framework_controls s, framework_controls t
WHERE s.code = '11.4.1' AND s.framework_id = 'a0000001-0000-0000-0000-000000000007'
  AND t.code = 'A.8.29' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- PCI DSS 11.5.1 (IDS/IPS) → NIST CSF DE.CM-01 (network monitoring)
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.85, 'Both require intrusion detection on networks', true
FROM framework_controls s, framework_controls t
WHERE s.code = '11.5.1' AND s.framework_id = 'a0000001-0000-0000-0000-000000000007'
  AND t.code = 'DE.CM-01' AND t.framework_id = 'a0000001-0000-0000-0000-000000000006';

-- PCI DSS 12.1.1 (security policy) → ISO A.5.1 (policies)
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.90, 'Both require information security policy', true
FROM framework_controls s, framework_controls t
WHERE s.code = '12.1.1' AND s.framework_id = 'a0000001-0000-0000-0000-000000000007'
  AND t.code = 'A.5.1' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- PCI DSS 12.6.1 (awareness program) → ISO A.6.3 (training)
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.90, 'Both require security awareness programs', true
FROM framework_controls s, framework_controls t
WHERE s.code = '12.6.1' AND s.framework_id = 'a0000001-0000-0000-0000-000000000007'
  AND t.code = 'A.6.3' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- PCI DSS 12.8.1 (TPSP list) → ISO A.5.19 (supplier relationships)
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'partial'::mapping_type, 0.75, 'PCI specific to payment TPSPs; ISO covers all suppliers', true
FROM framework_controls s, framework_controls t
WHERE s.code = '12.8.1' AND s.framework_id = 'a0000001-0000-0000-0000-000000000007'
  AND t.code = 'A.5.19' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- PCI DSS 12.10.1 (incident response) → ISO A.5.24 (incident planning)
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.90, 'Both require incident response plans', true
FROM framework_controls s, framework_controls t
WHERE s.code = '12.10.1' AND s.framework_id = 'a0000001-0000-0000-0000-000000000007'
  AND t.code = 'A.5.24' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';

-- PCI DSS 12.10.1 (incident response) → NIST CSF RS.MA-01 (IR plan executed)
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.85, 'Both require incident response execution', true
FROM framework_controls s, framework_controls t
WHERE s.code = '12.10.1' AND s.framework_id = 'a0000001-0000-0000-0000-000000000007'
  AND t.code = 'RS.MA-01' AND t.framework_id = 'a0000001-0000-0000-0000-000000000006';

-- PCI DSS 6.5.1 (change control) → ISO A.8.32 (change management)
INSERT INTO framework_control_mappings (source_control_id, target_control_id, mapping_type, mapping_strength, notes, is_verified)
SELECT s.id, t.id, 'equivalent'::mapping_type, 0.85, 'Both require change management procedures', true
FROM framework_controls s, framework_controls t
WHERE s.code = '6.5.1' AND s.framework_id = 'a0000001-0000-0000-0000-000000000007'
  AND t.code = 'A.8.32' AND t.framework_id = 'a0000001-0000-0000-0000-000000000001';
