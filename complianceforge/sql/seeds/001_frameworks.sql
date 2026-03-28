-- ============================================================
-- ComplianceForge — Seed: Compliance Frameworks
-- All 9 supported frameworks with domains
-- ============================================================

-- ISO/IEC 27001:2022
INSERT INTO compliance_frameworks (id, code, name, full_name, version, description, issuing_body, category, applicable_regions, applicable_industries, is_system_framework, is_active, total_controls, color_hex)
VALUES
('a0000001-0000-0000-0000-000000000001', 'ISO27001', 'ISO 27001', 'ISO/IEC 27001:2022 Information Security Management Systems', '2022', 'International standard for information security management systems (ISMS). Specifies requirements for establishing, implementing, maintaining and continually improving an ISMS.', 'ISO/IEC', 'security', '{global}', '{all}', true, true, 93, '#1E88E5'),

-- UK GDPR
('a0000001-0000-0000-0000-000000000002', 'UK_GDPR', 'UK GDPR', 'United Kingdom General Data Protection Regulation', '2021', 'UK-specific data protection regulation derived from EU GDPR, governing the processing of personal data of UK data subjects.', 'UK Government / ICO', 'privacy', '{UK}', '{all}', true, true, 99, '#7B1FA2'),

-- NCSC CAF
('a0000001-0000-0000-0000-000000000003', 'NCSC_CAF', 'NCSC CAF', 'NCSC Cyber Assessment Framework (NIS-aligned)', '3.2', 'UK National Cyber Security Centre framework for assessing cyber resilience of essential services and critical national infrastructure.', 'NCSC (UK)', 'security', '{UK}', '{critical_infrastructure,energy,transport,health,water,digital_services}', true, true, 39, '#FF6F00'),

-- Cyber Essentials
('a0000001-0000-0000-0000-000000000004', 'CYBER_ESSENTIALS', 'Cyber Essentials', 'Cyber Essentials Certification Scheme', '3.1', 'UK government-backed scheme helping organisations guard against the most common cyber threats with 5 technical controls.', 'NCSC (UK)', 'security', '{UK}', '{all}', true, true, 5, '#43A047'),

-- NIST SP 800-53 Rev 5
('a0000001-0000-0000-0000-000000000005', 'NIST_800_53', 'NIST 800-53', 'NIST Special Publication 800-53 Revision 5', 'Rev5', 'Security and privacy controls for information systems and organizations from the US National Institute of Standards and Technology.', 'NIST', 'security', '{US,global}', '{government,defense,all}', true, true, 1189, '#E53935'),

-- NIST CSF 2.0
('a0000001-0000-0000-0000-000000000006', 'NIST_CSF_2', 'NIST CSF 2.0', 'NIST Cybersecurity Framework 2.0', '2.0', 'Voluntary framework providing guidance for managing cybersecurity risks. Version 2.0 adds Govern function.', 'NIST', 'security', '{US,global}', '{all}', true, true, 106, '#F4511E'),

-- PCI DSS v4.0
('a0000001-0000-0000-0000-000000000007', 'PCI_DSS_4', 'PCI DSS', 'Payment Card Industry Data Security Standard v4.0', 'v4.0', 'Global security standard for organizations that handle cardholder data. Mandated by payment card brands.', 'PCI SSC', 'security', '{global}', '{financial_services,retail,hospitality,ecommerce}', true, true, 252, '#FF8F00'),

-- ITIL 4
('a0000001-0000-0000-0000-000000000008', 'ITIL_4', 'ITIL 4', 'Information Technology Infrastructure Library 4', '4', 'Best practice framework for IT service management supporting organizations in digital transformation.', 'AXELOS / PeopleCert', 'operational', '{global}', '{all}', true, true, 34, '#00897B'),

-- COBIT 2019
('a0000001-0000-0000-0000-000000000009', 'COBIT_2019', 'COBIT 2019', 'Control Objectives for Information and Related Technologies 2019', '2019', 'Enterprise IT governance framework providing 40 governance and management objectives across 5 domains.', 'ISACA', 'governance', '{global}', '{all}', true, true, 40, '#5E35B1');

-- ============================================================
-- FRAMEWORK DOMAINS
-- ============================================================

-- ISO 27001:2022 — 4 Annex A Themes
INSERT INTO framework_domains (id, framework_id, code, name, description, sort_order, depth_level, total_controls) VALUES
('d0000001-0001-0000-0000-000000000001', 'a0000001-0000-0000-0000-000000000001', 'A.5', 'Organisational Controls', 'Controls related to organisational policies, roles, responsibilities, and management direction.', 1, 0, 37),
('d0000001-0001-0000-0000-000000000002', 'a0000001-0000-0000-0000-000000000001', 'A.6', 'People Controls', 'Controls related to personnel security throughout their employment lifecycle.', 2, 0, 8),
('d0000001-0001-0000-0000-000000000003', 'a0000001-0000-0000-0000-000000000001', 'A.7', 'Physical Controls', 'Controls related to physical security of premises, equipment and information.', 3, 0, 14),
('d0000001-0001-0000-0000-000000000004', 'a0000001-0000-0000-0000-000000000001', 'A.8', 'Technological Controls', 'Controls related to technology and technical security measures.', 4, 0, 34);

-- NIST CSF 2.0 — 6 Functions
INSERT INTO framework_domains (id, framework_id, code, name, description, sort_order, depth_level, total_controls) VALUES
('d0000006-0001-0000-0000-000000000001', 'a0000001-0000-0000-0000-000000000006', 'GV', 'Govern', 'Establish and monitor the organization''s cybersecurity risk management strategy, expectations, and policy.', 1, 0, 16),
('d0000006-0001-0000-0000-000000000002', 'a0000001-0000-0000-0000-000000000006', 'ID', 'Identify', 'Understand the organisation''s current cybersecurity risks.', 2, 0, 17),
('d0000006-0001-0000-0000-000000000003', 'a0000001-0000-0000-0000-000000000006', 'PR', 'Protect', 'Use safeguards to manage the organisation''s cybersecurity risks.', 3, 0, 24),
('d0000006-0001-0000-0000-000000000004', 'a0000001-0000-0000-0000-000000000006', 'DE', 'Detect', 'Find and analyse possible cybersecurity attacks and compromises.', 4, 0, 8),
('d0000006-0001-0000-0000-000000000005', 'a0000001-0000-0000-0000-000000000006', 'RS', 'Respond', 'Take action regarding a detected cybersecurity incident.', 5, 0, 18),
('d0000006-0001-0000-0000-000000000006', 'a0000001-0000-0000-0000-000000000006', 'RC', 'Recover', 'Restore assets and operations affected by a cybersecurity incident.', 6, 0, 6);

-- PCI DSS v4.0 — 12 Requirements
INSERT INTO framework_domains (id, framework_id, code, name, description, sort_order, depth_level, total_controls) VALUES
('d0000007-0001-0000-0000-000000000001', 'a0000001-0000-0000-0000-000000000007', 'Req-1', 'Install and Maintain Network Security Controls', '', 1, 0, 15),
('d0000007-0001-0000-0000-000000000002', 'a0000001-0000-0000-0000-000000000007', 'Req-2', 'Apply Secure Configurations', '', 2, 0, 8),
('d0000007-0001-0000-0000-000000000003', 'a0000001-0000-0000-0000-000000000007', 'Req-3', 'Protect Stored Account Data', '', 3, 0, 25),
('d0000007-0001-0000-0000-000000000004', 'a0000001-0000-0000-0000-000000000007', 'Req-4', 'Protect Cardholder Data in Transit', '', 4, 0, 8),
('d0000007-0001-0000-0000-000000000005', 'a0000001-0000-0000-0000-000000000007', 'Req-5', 'Protect All Systems Against Malware', '', 5, 0, 15),
('d0000007-0001-0000-0000-000000000006', 'a0000001-0000-0000-0000-000000000007', 'Req-6', 'Develop and Maintain Secure Systems and Software', '', 6, 0, 30),
('d0000007-0001-0000-0000-000000000007', 'a0000001-0000-0000-0000-000000000007', 'Req-7', 'Restrict Access by Business Need to Know', '', 7, 0, 12),
('d0000007-0001-0000-0000-000000000008', 'a0000001-0000-0000-0000-000000000007', 'Req-8', 'Identify Users and Authenticate Access', '', 8, 0, 30),
('d0000007-0001-0000-0000-000000000009', 'a0000001-0000-0000-0000-000000000007', 'Req-9', 'Restrict Physical Access', '', 9, 0, 25),
('d0000007-0001-0000-0000-000000000010', 'a0000001-0000-0000-0000-000000000007', 'Req-10', 'Log and Monitor All Access', '', 10, 0, 30),
('d0000007-0001-0000-0000-000000000011', 'a0000001-0000-0000-0000-000000000007', 'Req-11', 'Test Security Regularly', '', 11, 0, 30),
('d0000007-0001-0000-0000-000000000012', 'a0000001-0000-0000-0000-000000000007', 'Req-12', 'Support InfoSec with Policies and Programs', '', 12, 0, 24);

-- NCSC CAF — 4 Objectives, 14 Principles
INSERT INTO framework_domains (id, framework_id, code, name, description, sort_order, depth_level, total_controls) VALUES
('d0000003-0001-0000-0000-000000000001', 'a0000001-0000-0000-0000-000000000003', 'A', 'Managing Security Risk', 'Appropriate governance structures, policies and processes to understand, assess and manage security risks.', 1, 0, 11),
('d0000003-0001-0000-0000-000000000002', 'a0000001-0000-0000-0000-000000000003', 'B', 'Protecting Against Cyber Attack', 'Proportionate security measures to protect systems and data from cyber attack.', 2, 0, 13),
('d0000003-0001-0000-0000-000000000003', 'a0000001-0000-0000-0000-000000000003', 'C', 'Detecting Cyber Security Events', 'Capabilities to detect cyber security events affecting systems and services.', 3, 0, 6),
('d0000003-0001-0000-0000-000000000004', 'a0000001-0000-0000-0000-000000000003', 'D', 'Minimising Impact of Cyber Security Incidents', 'Capabilities to minimise the impact of cyber security incidents.', 4, 0, 9);

-- Cyber Essentials — 5 Technical Controls
INSERT INTO framework_domains (id, framework_id, code, name, description, sort_order, depth_level, total_controls) VALUES
('d0000004-0001-0000-0000-000000000001', 'a0000001-0000-0000-0000-000000000004', 'CE-1', 'Firewalls', 'Boundary firewalls and internet gateways to protect against unauthorised access.', 1, 0, 1),
('d0000004-0001-0000-0000-000000000002', 'a0000001-0000-0000-0000-000000000004', 'CE-2', 'Secure Configuration', 'Computers and network devices configured to reduce vulnerabilities.', 2, 0, 1),
('d0000004-0001-0000-0000-000000000003', 'a0000001-0000-0000-0000-000000000004', 'CE-3', 'Security Update Management', 'Software and devices kept up to date to fix known vulnerabilities.', 3, 0, 1),
('d0000004-0001-0000-0000-000000000004', 'a0000001-0000-0000-0000-000000000004', 'CE-4', 'User Access Control', 'Accounts with special access privileges controlled and managed.', 4, 0, 1),
('d0000004-0001-0000-0000-000000000005', 'a0000001-0000-0000-0000-000000000004', 'CE-5', 'Malware Protection', 'Protection against malware using anti-malware software or application whitelisting.', 5, 0, 1);

-- COBIT 2019 — 5 Domains
INSERT INTO framework_domains (id, framework_id, code, name, description, sort_order, depth_level, total_controls) VALUES
('d0000009-0001-0000-0000-000000000001', 'a0000001-0000-0000-0000-000000000009', 'EDM', 'Evaluate, Direct and Monitor', 'Governance domain ensuring stakeholder-driven goal setting, direction, and monitoring.', 1, 0, 5),
('d0000009-0001-0000-0000-000000000002', 'a0000001-0000-0000-0000-000000000009', 'APO', 'Align, Plan and Organize', 'Management objectives for IT strategy, architecture, innovation and portfolio management.', 2, 0, 14),
('d0000009-0001-0000-0000-000000000003', 'a0000001-0000-0000-0000-000000000009', 'BAI', 'Build, Acquire and Implement', 'Management objectives for solutions development, acquisition and change management.', 3, 0, 11),
('d0000009-0001-0000-0000-000000000004', 'a0000001-0000-0000-0000-000000000009', 'DSS', 'Deliver, Service and Support', 'Management objectives for IT operations, service requests and incidents.', 4, 0, 6),
('d0000009-0001-0000-0000-000000000005', 'a0000001-0000-0000-0000-000000000009', 'MEA', 'Monitor, Evaluate and Assess', 'Management objectives for performance monitoring, internal control and compliance.', 5, 0, 4);

-- NIST 800-53 — Top-level families (20)
INSERT INTO framework_domains (id, framework_id, code, name, description, sort_order, depth_level, total_controls) VALUES
('d0000005-0001-0000-0000-000000000001', 'a0000001-0000-0000-0000-000000000005', 'AC', 'Access Control', '', 1, 0, 25),
('d0000005-0001-0000-0000-000000000002', 'a0000001-0000-0000-0000-000000000005', 'AT', 'Awareness and Training', '', 2, 0, 6),
('d0000005-0001-0000-0000-000000000003', 'a0000001-0000-0000-0000-000000000005', 'AU', 'Audit and Accountability', '', 3, 0, 16),
('d0000005-0001-0000-0000-000000000004', 'a0000001-0000-0000-0000-000000000005', 'CA', 'Assessment Authorization and Monitoring', '', 4, 0, 9),
('d0000005-0001-0000-0000-000000000005', 'a0000001-0000-0000-0000-000000000005', 'CM', 'Configuration Management', '', 5, 0, 14),
('d0000005-0001-0000-0000-000000000006', 'a0000001-0000-0000-0000-000000000005', 'CP', 'Contingency Planning', '', 6, 0, 13),
('d0000005-0001-0000-0000-000000000007', 'a0000001-0000-0000-0000-000000000005', 'IA', 'Identification and Authentication', '', 7, 0, 12),
('d0000005-0001-0000-0000-000000000008', 'a0000001-0000-0000-0000-000000000005', 'IR', 'Incident Response', '', 8, 0, 10),
('d0000005-0001-0000-0000-000000000009', 'a0000001-0000-0000-0000-000000000005', 'MA', 'Maintenance', '', 9, 0, 7),
('d0000005-0001-0000-0000-000000000010', 'a0000001-0000-0000-0000-000000000005', 'MP', 'Media Protection', '', 10, 0, 8),
('d0000005-0001-0000-0000-000000000011', 'a0000001-0000-0000-0000-000000000005', 'PE', 'Physical and Environmental Protection', '', 11, 0, 23),
('d0000005-0001-0000-0000-000000000012', 'a0000001-0000-0000-0000-000000000005', 'PL', 'Planning', '', 12, 0, 11),
('d0000005-0001-0000-0000-000000000013', 'a0000001-0000-0000-0000-000000000005', 'PM', 'Program Management', '', 13, 0, 32),
('d0000005-0001-0000-0000-000000000014', 'a0000001-0000-0000-0000-000000000005', 'PS', 'Personnel Security', '', 14, 0, 9),
('d0000005-0001-0000-0000-000000000015', 'a0000001-0000-0000-0000-000000000005', 'PT', 'Personally Identifiable Information Processing', '', 15, 0, 8),
('d0000005-0001-0000-0000-000000000016', 'a0000001-0000-0000-0000-000000000005', 'RA', 'Risk Assessment', '', 16, 0, 10),
('d0000005-0001-0000-0000-000000000017', 'a0000001-0000-0000-0000-000000000005', 'SA', 'System and Services Acquisition', '', 17, 0, 23),
('d0000005-0001-0000-0000-000000000018', 'a0000001-0000-0000-0000-000000000005', 'SC', 'System and Communications Protection', '', 18, 0, 51),
('d0000005-0001-0000-0000-000000000019', 'a0000001-0000-0000-0000-000000000005', 'SI', 'System and Information Integrity', '', 19, 0, 23),
('d0000005-0001-0000-0000-000000000020', 'a0000001-0000-0000-0000-000000000005', 'SR', 'Supply Chain Risk Management', '', 20, 0, 12);

-- ITIL 4 — Service Management Practices
INSERT INTO framework_domains (id, framework_id, code, name, description, sort_order, depth_level, total_controls) VALUES
('d0000008-0001-0000-0000-000000000001', 'a0000001-0000-0000-0000-000000000008', 'GP', 'General Management Practices', '', 1, 0, 14),
('d0000008-0001-0000-0000-000000000002', 'a0000001-0000-0000-0000-000000000008', 'SM', 'Service Management Practices', '', 2, 0, 17),
('d0000008-0001-0000-0000-000000000003', 'a0000001-0000-0000-0000-000000000008', 'TM', 'Technical Management Practices', '', 3, 0, 3);

-- UK GDPR — Chapter groupings
INSERT INTO framework_domains (id, framework_id, code, name, description, sort_order, depth_level, total_controls) VALUES
('d0000002-0001-0000-0000-000000000001', 'a0000001-0000-0000-0000-000000000002', 'CH2', 'Principles', 'Core data protection principles.', 1, 0, 7),
('d0000002-0001-0000-0000-000000000002', 'a0000001-0000-0000-0000-000000000002', 'CH3', 'Rights of the Data Subject', 'Rights granted to individuals regarding their personal data.', 2, 0, 12),
('d0000002-0001-0000-0000-000000000003', 'a0000001-0000-0000-0000-000000000002', 'CH4', 'Controller and Processor', 'Obligations of data controllers and processors.', 3, 0, 30),
('d0000002-0001-0000-0000-000000000004', 'a0000001-0000-0000-0000-000000000002', 'CH5', 'Transfers of Personal Data', 'Rules for transferring personal data internationally.', 4, 0, 15),
('d0000002-0001-0000-0000-000000000005', 'a0000001-0000-0000-0000-000000000002', 'CH8', 'Remedies, Liability and Penalties', 'Enforcement mechanisms and penalties.', 5, 0, 10);
