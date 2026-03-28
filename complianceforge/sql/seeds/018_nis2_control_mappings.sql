-- ============================================================
-- ComplianceForge -- Seed: NIS2 to ISO 27001:2022 Control Mappings
-- Cross-maps each NIS2 Article 21(2) measure to the relevant
-- ISO/IEC 27001:2022 Annex A controls that satisfy it.
-- ============================================================

-- ============================================================
-- (a) Risk analysis and information system security policies
-- Maps to: A.5.1 (Policies), A.5.2 (Roles), A.5.35 (Independent review)
-- ============================================================

INSERT INTO nis2_control_mappings (nis2_measure_code, iso_control_code, mapping_strength, notes) VALUES
('NIS2-Art21-a', 'A.5.1',  0.95, 'ISO A.5.1 requires establishment of information security policies, directly satisfying NIS2 Art 21(2)(a)'),
('NIS2-Art21-a', 'A.5.2',  0.85, 'Roles and responsibilities for information security support the risk governance structure'),
('NIS2-Art21-a', 'A.5.35', 0.80, 'Independent review of information security ensures policy effectiveness'),
('NIS2-Art21-a', 'A.5.36', 0.75, 'Compliance with policies, rules and standards supports risk management discipline');

-- ============================================================
-- (b) Incident handling
-- Maps to: A.5.24-A.5.28 (Incident management cluster), A.6.8
-- ============================================================

INSERT INTO nis2_control_mappings (nis2_measure_code, iso_control_code, mapping_strength, notes) VALUES
('NIS2-Art21-b', 'A.5.24', 0.95, 'Incident management planning and preparation directly addresses NIS2 incident handling'),
('NIS2-Art21-b', 'A.5.25', 0.90, 'Assessment and decision on information security events supports incident triage'),
('NIS2-Art21-b', 'A.5.26', 0.95, 'Response to information security incidents is the core of incident handling'),
('NIS2-Art21-b', 'A.5.27', 0.85, 'Learning from incidents supports continuous improvement of incident handling'),
('NIS2-Art21-b', 'A.5.28', 0.80, 'Collection of evidence supports forensic and regulatory reporting requirements'),
('NIS2-Art21-b', 'A.6.8',  0.80, 'Information security event reporting enables incident detection pipeline');

-- ============================================================
-- (c) Business continuity, backup, disaster recovery, crisis management
-- Maps to: A.5.29 (Disruption), A.5.30 (BCM), A.8.13 (Backup), A.8.14 (Redundancy)
-- ============================================================

INSERT INTO nis2_control_mappings (nis2_measure_code, iso_control_code, mapping_strength, notes) VALUES
('NIS2-Art21-c', 'A.5.29', 0.90, 'Information security during disruption directly addresses crisis management'),
('NIS2-Art21-c', 'A.5.30', 0.95, 'ICT readiness for business continuity is the primary ISO control for NIS2 BCP requirements'),
('NIS2-Art21-c', 'A.8.13', 0.90, 'Information backup ensures data recovery capability as required by NIS2'),
('NIS2-Art21-c', 'A.8.14', 0.85, 'Redundancy of information processing facilities supports availability and disaster recovery');

-- ============================================================
-- (d) Supply chain security
-- Maps to: A.5.19-A.5.23 (Supplier relationship cluster)
-- ============================================================

INSERT INTO nis2_control_mappings (nis2_measure_code, iso_control_code, mapping_strength, notes) VALUES
('NIS2-Art21-d', 'A.5.19', 0.95, 'Information security in supplier relationships directly addresses supply chain risk'),
('NIS2-Art21-d', 'A.5.20', 0.90, 'Addressing security within supplier agreements ensures contractual security requirements'),
('NIS2-Art21-d', 'A.5.21', 0.95, 'Managing security in the ICT supply chain is the primary control for NIS2 supply chain security'),
('NIS2-Art21-d', 'A.5.22', 0.85, 'Monitoring and review of supplier services ensures ongoing supply chain security'),
('NIS2-Art21-d', 'A.5.23', 0.80, 'Cloud services security addresses a significant component of modern supply chains');

-- ============================================================
-- (e) Security in acquisition, development, and maintenance
-- Maps to: A.5.8 (Project security), A.8.25-A.8.31 (Development controls)
-- ============================================================

INSERT INTO nis2_control_mappings (nis2_measure_code, iso_control_code, mapping_strength, notes) VALUES
('NIS2-Art21-e', 'A.5.8',  0.85, 'Information security in project management covers acquisition and development lifecycle'),
('NIS2-Art21-e', 'A.8.8',  0.85, 'Management of technical vulnerabilities addresses vulnerability handling and disclosure'),
('NIS2-Art21-e', 'A.8.9',  0.80, 'Configuration management supports secure system maintenance'),
('NIS2-Art21-e', 'A.8.25', 0.90, 'Secure development lifecycle directly supports NIS2 development security'),
('NIS2-Art21-e', 'A.8.26', 0.85, 'Application security requirements cover security in acquisition'),
('NIS2-Art21-e', 'A.8.27', 0.85, 'Secure system architecture and engineering principles support secure development'),
('NIS2-Art21-e', 'A.8.28', 0.80, 'Secure coding practices ensure development security'),
('NIS2-Art21-e', 'A.8.29', 0.85, 'Security testing in development and acceptance verifies security of developed systems'),
('NIS2-Art21-e', 'A.8.31', 0.80, 'Separation of development, test and production environments supports secure maintenance');

-- ============================================================
-- (f) Assessing effectiveness of cybersecurity measures
-- Maps to: A.5.35 (Independent review), A.5.36 (Compliance), A.8.8 (Vuln management)
-- ============================================================

INSERT INTO nis2_control_mappings (nis2_measure_code, iso_control_code, mapping_strength, notes) VALUES
('NIS2-Art21-f', 'A.5.35', 0.95, 'Independent review of information security is the primary mechanism for assessing effectiveness'),
('NIS2-Art21-f', 'A.5.36', 0.90, 'Compliance assessment with policies, rules and standards measures control effectiveness'),
('NIS2-Art21-f', 'A.8.8',  0.80, 'Technical vulnerability management provides measurable assessment of security posture'),
('NIS2-Art21-f', 'A.8.34', 0.85, 'Protection of information systems during audit testing supports effective assessments');

-- ============================================================
-- (g) Basic cyber hygiene and cybersecurity training
-- Maps to: A.6.3 (Training), A.6.8 (Event reporting), A.7.7 (Clear desk/screen)
-- ============================================================

INSERT INTO nis2_control_mappings (nis2_measure_code, iso_control_code, mapping_strength, notes) VALUES
('NIS2-Art21-g', 'A.6.3',  0.95, 'Information security awareness, education and training is the primary control for NIS2 training requirement'),
('NIS2-Art21-g', 'A.5.4',  0.80, 'Management responsibilities include ensuring staff are aware of security obligations'),
('NIS2-Art21-g', 'A.6.8',  0.75, 'Event reporting culture is a fundamental aspect of cyber hygiene'),
('NIS2-Art21-g', 'A.7.7',  0.70, 'Clear desk and clear screen are basic cyber hygiene practices'),
('NIS2-Art21-g', 'A.8.1',  0.75, 'User endpoint devices control supports basic cyber hygiene through device management');

-- ============================================================
-- (h) Cryptography and encryption
-- Maps to: A.8.24 (Cryptography), A.5.14 (Information transfer)
-- ============================================================

INSERT INTO nis2_control_mappings (nis2_measure_code, iso_control_code, mapping_strength, notes) VALUES
('NIS2-Art21-h', 'A.8.24', 0.95, 'Use of cryptography is the primary ISO control covering NIS2 encryption requirements'),
('NIS2-Art21-h', 'A.5.14', 0.80, 'Information transfer controls often require cryptographic protection'),
('NIS2-Art21-h', 'A.5.17', 0.75, 'Authentication information management involves cryptographic controls for credential protection');

-- ============================================================
-- (i) Human resources security, access control, asset management
-- Maps to: A.5.9-A.5.18 (Access/Asset cluster), A.6.1-A.6.7 (People controls)
-- ============================================================

INSERT INTO nis2_control_mappings (nis2_measure_code, iso_control_code, mapping_strength, notes) VALUES
('NIS2-Art21-i', 'A.5.9',  0.85, 'Asset inventory is foundational to asset management'),
('NIS2-Art21-i', 'A.5.10', 0.80, 'Acceptable use policies govern asset usage'),
('NIS2-Art21-i', 'A.5.11', 0.75, 'Return of assets supports HR security for leavers'),
('NIS2-Art21-i', 'A.5.15', 0.90, 'Access control is a primary component of this NIS2 measure'),
('NIS2-Art21-i', 'A.5.16', 0.90, 'Identity management directly supports access control requirements'),
('NIS2-Art21-i', 'A.5.18', 0.85, 'Access rights management ensures least privilege'),
('NIS2-Art21-i', 'A.6.1',  0.85, 'Personnel screening is a core HR security control'),
('NIS2-Art21-i', 'A.6.2',  0.80, 'Terms and conditions of employment establish security obligations'),
('NIS2-Art21-i', 'A.6.5',  0.80, 'Responsibilities after termination ensure secure offboarding'),
('NIS2-Art21-i', 'A.8.2',  0.80, 'Privileged access rights management is critical for access control'),
('NIS2-Art21-i', 'A.8.3',  0.80, 'Information access restriction supports least-privilege access control');

-- ============================================================
-- (j) Multi-factor authentication and secured communications
-- Maps to: A.5.14 (Transfer), A.5.17 (Auth), A.8.5 (Secure auth), A.8.20-A.8.21 (Network)
-- ============================================================

INSERT INTO nis2_control_mappings (nis2_measure_code, iso_control_code, mapping_strength, notes) VALUES
('NIS2-Art21-j', 'A.8.5',  0.95, 'Secure authentication directly addresses MFA requirements'),
('NIS2-Art21-j', 'A.5.14', 0.85, 'Information transfer controls cover secured communications'),
('NIS2-Art21-j', 'A.5.17', 0.80, 'Authentication information management supports MFA implementation'),
('NIS2-Art21-j', 'A.8.20', 0.85, 'Network security supports secured communications infrastructure'),
('NIS2-Art21-j', 'A.8.21', 0.85, 'Security of network services ensures communication channel security'),
('NIS2-Art21-j', 'A.8.24', 0.80, 'Cryptography supports encryption of voice/video/text communications')
ON CONFLICT DO NOTHING;
