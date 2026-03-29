-- ============================================================
-- ComplianceForge GRC Platform
-- Seed 026: TPRM Questionnaire Templates
-- ============================================================
-- Creates four system questionnaire templates:
--   1. ComplianceForge Standard Security Assessment (~80 questions)
--   2. GDPR Article 28 Processor Assessment (~40 questions)
--   3. NIS2 Supply Chain Security Assessment (~30 questions)
--   4. Quick Security Assessment (~20 questions)
-- ============================================================

BEGIN;

-- ============================================================
-- TEMPLATE 1: ComplianceForge Standard Security Assessment
-- ============================================================

INSERT INTO assessment_questionnaires (
    id, organization_id, name, description, questionnaire_type, version, status,
    total_questions, total_sections, estimated_completion_minutes,
    scoring_method, pass_threshold, risk_tier_thresholds,
    applicable_vendor_tiers, is_system, is_template
) VALUES (
    'a0000001-0000-0000-0000-000000000001', NULL,
    'ComplianceForge Standard Security Assessment',
    'Comprehensive security assessment covering governance, access control, data protection, incident response, business continuity, vulnerability management, physical security, supply chain, cloud security, privacy, and compliance. Suitable for critical and high-risk vendors.',
    'security', 1, 'active',
    80, 11, 120,
    'weighted_average', 70.00,
    '{"critical": 40, "high": 55, "medium": 70, "low": 85}',
    ARRAY['critical', 'high', 'medium'],
    true, true
);

-- Section 1: Security Governance & Organisation
INSERT INTO questionnaire_sections (id, questionnaire_id, name, description, sort_order, weight, framework_domain_code)
VALUES ('b0000001-0001-0000-0000-000000000001', 'a0000001-0000-0000-0000-000000000001',
    'Security Governance & Organisation', 'Policies, leadership, roles, risk management processes', 1, 1.20, 'ISO27001-A5');

INSERT INTO questionnaire_questions (section_id, question_text, question_type, options, is_required, weight, risk_impact, guidance_text, evidence_required, sort_order, tags) VALUES
('b0000001-0001-0000-0000-000000000001', 'Does your organisation have a formally documented Information Security Policy that is approved by senior management?', 'single_choice', '[{"value":"yes_current","label":"Yes, reviewed within 12 months","score":100},{"value":"yes_outdated","label":"Yes, but not recently reviewed","score":60},{"value":"in_progress","label":"In development","score":30},{"value":"no","label":"No","score":0}]', true, 1.50, 'high', 'Look for a policy signed by a C-level executive and reviewed at least annually.', true, 1, ARRAY['governance','policy']),
('b0000001-0001-0000-0000-000000000001', 'Is there a designated Chief Information Security Officer (CISO) or equivalent role?', 'single_choice', '[{"value":"dedicated","label":"Dedicated CISO","score":100},{"value":"combined","label":"Combined role (e.g. CTO/CISO)","score":70},{"value":"outsourced","label":"Outsourced / virtual CISO","score":50},{"value":"none","label":"No designated role","score":0}]', true, 1.20, 'high', 'The CISO should report to the board or CEO, not solely to IT.', false, 2, ARRAY['governance','leadership']),
('b0000001-0001-0000-0000-000000000001', 'Do you conduct regular information security risk assessments?', 'single_choice', '[{"value":"quarterly","label":"Quarterly or more frequently","score":100},{"value":"annually","label":"Annually","score":80},{"value":"adhoc","label":"Ad-hoc only","score":40},{"value":"never","label":"Never","score":0}]', true, 1.30, 'critical', 'Risk assessments should use a recognised methodology (ISO 27005, NIST 800-30, etc.).', true, 3, ARRAY['governance','risk']),
('b0000001-0001-0000-0000-000000000001', 'Does your organisation maintain an information asset register?', 'single_choice', '[{"value":"yes_classified","label":"Yes, with data classification","score":100},{"value":"yes_basic","label":"Yes, basic inventory","score":60},{"value":"partial","label":"Partial coverage","score":30},{"value":"no","label":"No","score":0}]', true, 1.00, 'medium', 'The register should cover all systems processing client data.', false, 4, ARRAY['governance','asset-mgmt']),
('b0000001-0001-0000-0000-000000000001', 'Is there a formal security awareness training programme for all employees?', 'single_choice', '[{"value":"regular","label":"Yes, at least annually with phishing simulations","score":100},{"value":"annual","label":"Yes, annual training only","score":70},{"value":"onboarding","label":"Onboarding only","score":40},{"value":"none","label":"No formal programme","score":0}]', true, 1.10, 'high', 'Training should cover phishing, social engineering, data handling, and incident reporting.', true, 5, ARRAY['governance','training']),
('b0000001-0001-0000-0000-000000000001', 'Do you have documented information security roles and responsibilities?', 'yes_no', '[{"value":"yes","label":"Yes","score":100},{"value":"no","label":"No","score":0}]', true, 0.80, 'medium', 'Roles should be clearly defined in job descriptions and the security policy.', false, 6, ARRAY['governance','roles']),
('b0000001-0001-0000-0000-000000000001', 'Is there a management review process for information security at least annually?', 'yes_no', '[{"value":"yes","label":"Yes","score":100},{"value":"no","label":"No","score":0}]', true, 0.90, 'medium', 'Review should cover risk posture, audit findings, incident trends, and improvement plans.', false, 7, ARRAY['governance','review']);

-- Section 2: Access Control & Identity Management
INSERT INTO questionnaire_sections (id, questionnaire_id, name, description, sort_order, weight, framework_domain_code)
VALUES ('b0000001-0002-0000-0000-000000000001', 'a0000001-0000-0000-0000-000000000001',
    'Access Control & Identity Management', 'Authentication, authorisation, privileged access, and user lifecycle', 2, 1.30, 'ISO27001-A9');

INSERT INTO questionnaire_questions (section_id, question_text, question_type, options, is_required, weight, risk_impact, guidance_text, evidence_required, sort_order, tags) VALUES
('b0000001-0002-0000-0000-000000000001', 'Is multi-factor authentication (MFA) enforced for all users accessing systems that process our data?', 'single_choice', '[{"value":"all","label":"Yes, for all users","score":100},{"value":"privileged","label":"For privileged accounts only","score":60},{"value":"optional","label":"Available but optional","score":30},{"value":"none","label":"No MFA","score":0}]', true, 1.50, 'critical', 'MFA should be enforced at minimum for admin and remote access.', true, 1, ARRAY['access-control','mfa']),
('b0000001-0002-0000-0000-000000000001', 'Do you follow the principle of least privilege for access rights?', 'single_choice', '[{"value":"enforced","label":"Yes, enforced with regular reviews","score":100},{"value":"policy","label":"Yes, policy exists but not always enforced","score":60},{"value":"partial","label":"Partially implemented","score":30},{"value":"no","label":"No","score":0}]', true, 1.20, 'high', 'Users should only have access necessary for their role.', false, 2, ARRAY['access-control','least-privilege']),
('b0000001-0002-0000-0000-000000000001', 'How often are user access rights reviewed?', 'single_choice', '[{"value":"quarterly","label":"Quarterly","score":100},{"value":"semi","label":"Semi-annually","score":80},{"value":"annually","label":"Annually","score":60},{"value":"never","label":"Never or ad-hoc","score":0}]', true, 1.10, 'high', 'Reviews should include verification of role appropriateness and removal of orphaned accounts.', true, 3, ARRAY['access-control','review']),
('b0000001-0002-0000-0000-000000000001', 'Is there a formal process for granting, modifying, and revoking user access?', 'yes_no', '[{"value":"yes","label":"Yes","score":100},{"value":"no","label":"No","score":0}]', true, 1.00, 'high', 'Joiners/movers/leavers process should be documented and auditable.', false, 4, ARRAY['access-control','lifecycle']),
('b0000001-0002-0000-0000-000000000001', 'Do you use a centralised identity management system (e.g. Active Directory, Okta)?', 'yes_no', '[{"value":"yes","label":"Yes","score":100},{"value":"no","label":"No","score":0}]', true, 0.90, 'medium', 'Centralised identity reduces the risk of orphaned accounts.', false, 5, ARRAY['access-control','identity']),
('b0000001-0002-0000-0000-000000000001', 'Are privileged accounts (admin, root) managed with a PAM solution or equivalent controls?', 'single_choice', '[{"value":"pam","label":"Yes, dedicated PAM solution","score":100},{"value":"manual","label":"Manual controls with logging","score":60},{"value":"shared","label":"Shared admin accounts","score":20},{"value":"none","label":"No specific controls","score":0}]', true, 1.40, 'critical', 'Privileged access should be time-limited, logged, and subject to break-glass procedures.', true, 6, ARRAY['access-control','pam']),
('b0000001-0002-0000-0000-000000000001', 'What is your password policy minimum length requirement?', 'single_choice', '[{"value":"14plus","label":"14+ characters","score":100},{"value":"12","label":"12-13 characters","score":80},{"value":"8","label":"8-11 characters","score":50},{"value":"less","label":"Less than 8 characters","score":0}]', true, 0.80, 'medium', 'NIST 800-63B recommends at least 8 characters; best practice is 12+.', false, 7, ARRAY['access-control','password']),
('b0000001-0002-0000-0000-000000000001', 'Are service accounts and API keys rotated regularly?', 'single_choice', '[{"value":"automated","label":"Yes, automated rotation","score":100},{"value":"manual_regular","label":"Yes, manual rotation on schedule","score":70},{"value":"adhoc","label":"Ad-hoc only","score":30},{"value":"never","label":"Never rotated","score":0}]', true, 1.00, 'high', 'Service credentials should be rotated at least every 90 days.', false, 8, ARRAY['access-control','credentials']);

-- Section 3: Data Protection & Encryption
INSERT INTO questionnaire_sections (id, questionnaire_id, name, description, sort_order, weight, framework_domain_code)
VALUES ('b0000001-0003-0000-0000-000000000001', 'a0000001-0000-0000-0000-000000000001',
    'Data Protection & Encryption', 'Encryption at rest, in transit, key management, data classification', 3, 1.30, 'ISO27001-A10');

INSERT INTO questionnaire_questions (section_id, question_text, question_type, options, is_required, weight, risk_impact, guidance_text, evidence_required, sort_order, tags) VALUES
('b0000001-0003-0000-0000-000000000001', 'Is all data at rest encrypted using AES-256 or equivalent?', 'single_choice', '[{"value":"aes256","label":"Yes, AES-256 or equivalent","score":100},{"value":"aes128","label":"AES-128 or equivalent","score":80},{"value":"partial","label":"Partial encryption","score":40},{"value":"none","label":"No encryption at rest","score":0}]', true, 1.40, 'critical', 'All databases, file stores, and backups containing client data should be encrypted.', true, 1, ARRAY['data-protection','encryption']),
('b0000001-0003-0000-0000-000000000001', 'Is all data in transit encrypted using TLS 1.2 or higher?', 'single_choice', '[{"value":"tls13","label":"TLS 1.3","score":100},{"value":"tls12","label":"TLS 1.2","score":90},{"value":"mixed","label":"Mixed TLS versions","score":50},{"value":"none","label":"Unencrypted connections allowed","score":0}]', true, 1.40, 'critical', 'Deprecated protocols (SSL, TLS 1.0, TLS 1.1) should be disabled.', true, 2, ARRAY['data-protection','tls']),
('b0000001-0003-0000-0000-000000000001', 'Do you have a formal data classification scheme?', 'single_choice', '[{"value":"full","label":"Yes, with 4+ levels and handling rules","score":100},{"value":"basic","label":"Yes, basic classification","score":60},{"value":"informal","label":"Informal only","score":30},{"value":"none","label":"No classification scheme","score":0}]', true, 1.00, 'high', 'Classification should include public, internal, confidential, and restricted levels at minimum.', false, 3, ARRAY['data-protection','classification']),
('b0000001-0003-0000-0000-000000000001', 'How are encryption keys managed?', 'single_choice', '[{"value":"hsm","label":"HSM or cloud KMS with separation of duties","score":100},{"value":"kms","label":"Software KMS","score":70},{"value":"manual","label":"Manual key management","score":30},{"value":"none","label":"No formal key management","score":0}]', true, 1.20, 'critical', 'Keys should be stored separately from data with audit logging of all access.', true, 4, ARRAY['data-protection','key-mgmt']),
('b0000001-0003-0000-0000-000000000001', 'Do you have a data retention and secure deletion policy?', 'single_choice', '[{"value":"automated","label":"Yes, with automated enforcement","score":100},{"value":"manual","label":"Yes, with manual enforcement","score":70},{"value":"policy_only","label":"Policy exists but not enforced","score":30},{"value":"none","label":"No retention policy","score":0}]', true, 1.00, 'high', 'Data should be securely deleted when no longer needed or when a client requests deletion.', false, 5, ARRAY['data-protection','retention']),
('b0000001-0003-0000-0000-000000000001', 'Are backups encrypted and tested for recoverability?', 'single_choice', '[{"value":"both","label":"Yes, encrypted and regularly tested","score":100},{"value":"encrypted","label":"Encrypted but not tested","score":60},{"value":"tested","label":"Tested but not encrypted","score":40},{"value":"neither","label":"Neither","score":0}]', true, 1.10, 'high', 'Backup restoration should be tested at least quarterly.', true, 6, ARRAY['data-protection','backup']),
('b0000001-0003-0000-0000-000000000001', 'Do you prevent data exfiltration through DLP controls?', 'single_choice', '[{"value":"comprehensive","label":"Yes, comprehensive DLP","score":100},{"value":"partial","label":"Partial DLP controls","score":60},{"value":"monitoring","label":"Monitoring only","score":30},{"value":"none","label":"No DLP controls","score":0}]', true, 1.00, 'high', 'DLP should cover email, web uploads, USB, and cloud storage.', false, 7, ARRAY['data-protection','dlp']);

-- Section 4: Network Security
INSERT INTO questionnaire_sections (id, questionnaire_id, name, description, sort_order, weight, framework_domain_code)
VALUES ('b0000001-0004-0000-0000-000000000001', 'a0000001-0000-0000-0000-000000000001',
    'Network Security', 'Firewalls, segmentation, IDS/IPS, monitoring', 4, 1.10, 'ISO27001-A13');

INSERT INTO questionnaire_questions (section_id, question_text, question_type, options, is_required, weight, risk_impact, guidance_text, evidence_required, sort_order, tags) VALUES
('b0000001-0004-0000-0000-000000000001', 'Is your network segmented to isolate sensitive data processing environments?', 'single_choice', '[{"value":"micro","label":"Yes, micro-segmentation","score":100},{"value":"vlan","label":"Yes, VLAN segmentation","score":80},{"value":"basic","label":"Basic segmentation","score":50},{"value":"flat","label":"Flat network","score":0}]', true, 1.30, 'critical', 'Production, development, and management networks should be logically separated.', true, 1, ARRAY['network','segmentation']),
('b0000001-0004-0000-0000-000000000001', 'Do you have a web application firewall (WAF) protecting internet-facing applications?', 'yes_no', '[{"value":"yes","label":"Yes","score":100},{"value":"no","label":"No","score":0}]', true, 1.10, 'high', 'WAF should be configured to prevent OWASP Top 10 attacks.', false, 2, ARRAY['network','waf']),
('b0000001-0004-0000-0000-000000000001', 'Do you deploy intrusion detection/prevention systems (IDS/IPS)?', 'single_choice', '[{"value":"ips","label":"IPS (active blocking)","score":100},{"value":"ids","label":"IDS (monitoring only)","score":70},{"value":"partial","label":"Partial coverage","score":40},{"value":"none","label":"No IDS/IPS","score":0}]', true, 1.10, 'high', 'Systems should be monitored 24/7 with alert escalation procedures.', false, 3, ARRAY['network','ids']),
('b0000001-0004-0000-0000-000000000001', 'Are all network devices (firewalls, switches, routers) hardened according to a baseline?', 'yes_no', '[{"value":"yes","label":"Yes","score":100},{"value":"no","label":"No","score":0}]', true, 0.90, 'medium', 'Hardening should follow CIS benchmarks or vendor best practices.', false, 4, ARRAY['network','hardening']),
('b0000001-0004-0000-0000-000000000001', 'Do you perform regular network penetration testing?', 'single_choice', '[{"value":"quarterly","label":"Quarterly or more","score":100},{"value":"annually","label":"Annually","score":80},{"value":"adhoc","label":"Ad-hoc","score":40},{"value":"never","label":"Never","score":0}]', true, 1.20, 'high', 'Penetration tests should be performed by an independent third party.', true, 5, ARRAY['network','pentest']),
('b0000001-0004-0000-0000-000000000001', 'Is network traffic logged and monitored centrally?', 'single_choice', '[{"value":"siem","label":"Yes, SIEM with 24/7 SOC","score":100},{"value":"siem_no_soc","label":"SIEM without 24/7 SOC","score":70},{"value":"basic","label":"Basic logging only","score":40},{"value":"none","label":"No centralised logging","score":0}]', true, 1.20, 'high', 'Logs should be retained for at least 12 months and tamper-protected.', false, 6, ARRAY['network','monitoring']);

-- Section 5: Vulnerability Management
INSERT INTO questionnaire_sections (id, questionnaire_id, name, description, sort_order, weight, framework_domain_code)
VALUES ('b0000001-0005-0000-0000-000000000001', 'a0000001-0000-0000-0000-000000000001',
    'Vulnerability Management', 'Scanning, patching, software development security', 5, 1.10, 'ISO27001-A12');

INSERT INTO questionnaire_questions (section_id, question_text, question_type, options, is_required, weight, risk_impact, guidance_text, evidence_required, sort_order, tags) VALUES
('b0000001-0005-0000-0000-000000000001', 'How frequently do you conduct vulnerability scans on production systems?', 'single_choice', '[{"value":"continuous","label":"Continuous / daily","score":100},{"value":"weekly","label":"Weekly","score":90},{"value":"monthly","label":"Monthly","score":70},{"value":"quarterly","label":"Quarterly or less","score":40},{"value":"never","label":"Never","score":0}]', true, 1.30, 'critical', 'Scans should cover all internet-facing and internal systems processing client data.', true, 1, ARRAY['vuln-mgmt','scanning']),
('b0000001-0005-0000-0000-000000000001', 'What is your SLA for patching critical vulnerabilities?', 'single_choice', '[{"value":"24h","label":"Within 24 hours","score":100},{"value":"72h","label":"Within 72 hours","score":85},{"value":"7d","label":"Within 7 days","score":60},{"value":"30d","label":"Within 30 days","score":30},{"value":"none","label":"No defined SLA","score":0}]', true, 1.40, 'critical', 'CISA recommends critical CVEs be patched within 15 days of discovery.', true, 2, ARRAY['vuln-mgmt','patching']),
('b0000001-0005-0000-0000-000000000001', 'Do you have a secure software development lifecycle (SSDLC)?', 'single_choice', '[{"value":"mature","label":"Yes, with SAST/DAST/SCA integrated","score":100},{"value":"basic","label":"Yes, basic SSDLC","score":70},{"value":"informal","label":"Informal practices","score":30},{"value":"none","label":"No SSDLC","score":0}]', true, 1.10, 'high', 'SSDLC should include threat modelling, code review, and security testing.', false, 3, ARRAY['vuln-mgmt','ssdlc']),
('b0000001-0005-0000-0000-000000000001', 'Do you maintain a software bill of materials (SBOM) for your applications?', 'yes_no', '[{"value":"yes","label":"Yes","score":100},{"value":"no","label":"No","score":0}]', true, 0.80, 'medium', 'SBOM helps track third-party library vulnerabilities (e.g. Log4Shell).', false, 4, ARRAY['vuln-mgmt','sbom']),
('b0000001-0005-0000-0000-000000000001', 'How do you manage end-of-life (EOL) software?', 'single_choice', '[{"value":"no_eol","label":"No EOL software in production","score":100},{"value":"compensating","label":"EOL exists with compensating controls","score":60},{"value":"some_eol","label":"Some EOL without mitigations","score":20},{"value":"unknown","label":"Unknown","score":0}]', true, 1.00, 'high', 'EOL software no longer receives security patches and poses significant risk.', false, 5, ARRAY['vuln-mgmt','eol']),
('b0000001-0005-0000-0000-000000000001', 'Are application dependencies automatically scanned for known vulnerabilities?', 'yes_no', '[{"value":"yes","label":"Yes","score":100},{"value":"no","label":"No","score":0}]', true, 0.90, 'medium', 'Tools like Dependabot, Snyk, or Renovate should be integrated into CI/CD.', false, 6, ARRAY['vuln-mgmt','dependency-scan']);

-- Section 6: Incident Response
INSERT INTO questionnaire_sections (id, questionnaire_id, name, description, sort_order, weight, framework_domain_code)
VALUES ('b0000001-0006-0000-0000-000000000001', 'a0000001-0000-0000-0000-000000000001',
    'Incident Response', 'Response plans, communication, breach notification', 6, 1.20, 'ISO27001-A16');

INSERT INTO questionnaire_questions (section_id, question_text, question_type, options, is_required, weight, risk_impact, guidance_text, evidence_required, sort_order, tags) VALUES
('b0000001-0006-0000-0000-000000000001', 'Do you have a documented incident response plan?', 'single_choice', '[{"value":"tested","label":"Yes, documented and tested annually","score":100},{"value":"documented","label":"Yes, documented but not tested","score":60},{"value":"informal","label":"Informal process only","score":30},{"value":"none","label":"No incident response plan","score":0}]', true, 1.40, 'critical', 'Plans should include roles, escalation procedures, and communication templates.', true, 1, ARRAY['incident-response','plan']),
('b0000001-0006-0000-0000-000000000001', 'Do you have the ability to notify affected clients within 72 hours of a data breach?', 'single_choice', '[{"value":"under_24h","label":"Yes, within 24 hours","score":100},{"value":"under_72h","label":"Yes, within 72 hours","score":80},{"value":"best_effort","label":"Best effort, no guarantee","score":40},{"value":"no","label":"No defined process","score":0}]', true, 1.30, 'critical', 'GDPR requires supervisory authority notification within 72 hours.', false, 2, ARRAY['incident-response','notification']),
('b0000001-0006-0000-0000-000000000001', 'Do you conduct regular incident response tabletop exercises?', 'single_choice', '[{"value":"quarterly","label":"Quarterly or more","score":100},{"value":"annually","label":"Annually","score":70},{"value":"adhoc","label":"Ad-hoc","score":30},{"value":"never","label":"Never","score":0}]', true, 1.10, 'high', 'Exercises should simulate realistic scenarios relevant to your threat landscape.', true, 3, ARRAY['incident-response','exercise']),
('b0000001-0006-0000-0000-000000000001', 'Do you have cyber insurance coverage?', 'yes_no', '[{"value":"yes","label":"Yes","score":100},{"value":"no","label":"No","score":0}]', false, 0.60, 'low', 'Cyber insurance provides financial risk transfer for breach costs.', false, 4, ARRAY['incident-response','insurance']),
('b0000001-0006-0000-0000-000000000001', 'Do you retain forensic investigation capabilities (internal or third-party retainer)?', 'single_choice', '[{"value":"internal","label":"Internal forensics team","score":100},{"value":"retainer","label":"Third-party retainer","score":90},{"value":"adhoc","label":"Engaged ad-hoc","score":50},{"value":"none","label":"No capability","score":0}]', true, 1.00, 'high', 'Pre-arranged retainers ensure rapid response in a crisis.', false, 5, ARRAY['incident-response','forensics']),
('b0000001-0006-0000-0000-000000000001', 'Is there a post-incident review (lessons learned) process?', 'yes_no', '[{"value":"yes","label":"Yes","score":100},{"value":"no","label":"No","score":0}]', true, 0.80, 'medium', 'Reviews should identify root causes and track remediation actions.', false, 6, ARRAY['incident-response','review']),
('b0000001-0006-0000-0000-000000000001', 'Do you maintain an incident log with classification and trending analysis?', 'yes_no', '[{"value":"yes","label":"Yes","score":100},{"value":"no","label":"No","score":0}]', true, 0.80, 'medium', 'Historical data enables trend analysis and threat intelligence enrichment.', false, 7, ARRAY['incident-response','logging']);

-- Section 7: Business Continuity & Disaster Recovery
INSERT INTO questionnaire_sections (id, questionnaire_id, name, description, sort_order, weight, framework_domain_code)
VALUES ('b0000001-0007-0000-0000-000000000001', 'a0000001-0000-0000-0000-000000000001',
    'Business Continuity & Disaster Recovery', 'BCP/DRP, RTO/RPO, failover, and resilience', 7, 1.00, 'ISO27001-A17');

INSERT INTO questionnaire_questions (section_id, question_text, question_type, options, is_required, weight, risk_impact, guidance_text, evidence_required, sort_order, tags) VALUES
('b0000001-0007-0000-0000-000000000001', 'Do you have a documented Business Continuity Plan (BCP)?', 'single_choice', '[{"value":"tested","label":"Yes, tested within 12 months","score":100},{"value":"documented","label":"Yes, but not recently tested","score":60},{"value":"in_progress","label":"In development","score":30},{"value":"none","label":"No BCP","score":0}]', true, 1.20, 'high', 'BCP should cover all critical business processes and be tested annually.', true, 1, ARRAY['bcdr','bcp']),
('b0000001-0007-0000-0000-000000000001', 'What is your Recovery Time Objective (RTO) for services provided to us?', 'single_choice', '[{"value":"under_1h","label":"Under 1 hour","score":100},{"value":"under_4h","label":"1-4 hours","score":80},{"value":"under_24h","label":"4-24 hours","score":60},{"value":"over_24h","label":"Over 24 hours","score":30},{"value":"undefined","label":"Not defined","score":0}]', true, 1.10, 'high', 'RTO should be contractually aligned with your SLA commitments.', false, 2, ARRAY['bcdr','rto']),
('b0000001-0007-0000-0000-000000000001', 'What is your Recovery Point Objective (RPO)?', 'single_choice', '[{"value":"zero","label":"Near-zero (synchronous replication)","score":100},{"value":"under_1h","label":"Under 1 hour","score":85},{"value":"under_24h","label":"1-24 hours","score":60},{"value":"over_24h","label":"Over 24 hours","score":30},{"value":"undefined","label":"Not defined","score":0}]', true, 1.00, 'high', 'RPO determines maximum acceptable data loss measured in time.', false, 3, ARRAY['bcdr','rpo']),
('b0000001-0007-0000-0000-000000000001', 'Do you operate from geographically diverse data centres?', 'yes_no', '[{"value":"yes","label":"Yes","score":100},{"value":"no","label":"No","score":0}]', true, 0.90, 'medium', 'Geographic diversity protects against regional disasters.', false, 4, ARRAY['bcdr','geo-diversity']),
('b0000001-0007-0000-0000-000000000001', 'How often are disaster recovery procedures tested?', 'single_choice', '[{"value":"quarterly","label":"Quarterly","score":100},{"value":"semi","label":"Semi-annually","score":80},{"value":"annually","label":"Annually","score":60},{"value":"never","label":"Never","score":0}]', true, 1.10, 'high', 'Tests should include full failover exercises, not just plan walkthroughs.', true, 5, ARRAY['bcdr','testing']),
('b0000001-0007-0000-0000-000000000001', 'Do you have automated failover for critical systems?', 'yes_no', '[{"value":"yes","label":"Yes","score":100},{"value":"no","label":"No","score":0}]', true, 1.00, 'high', 'Automated failover reduces RTO and human error during a crisis.', false, 6, ARRAY['bcdr','failover']);

-- Section 8: Physical Security
INSERT INTO questionnaire_sections (id, questionnaire_id, name, description, sort_order, weight, framework_domain_code)
VALUES ('b0000001-0008-0000-0000-000000000001', 'a0000001-0000-0000-0000-000000000001',
    'Physical Security', 'Facility access, environmental controls, media handling', 8, 0.70, 'ISO27001-A11');

INSERT INTO questionnaire_questions (section_id, question_text, question_type, options, is_required, weight, risk_impact, guidance_text, evidence_required, sort_order, tags) VALUES
('b0000001-0008-0000-0000-000000000001', 'Are data centres protected with multi-layered physical access controls?', 'single_choice', '[{"value":"multi","label":"Yes, biometric + card + escort","score":100},{"value":"card","label":"Card access with logging","score":70},{"value":"basic","label":"Lock and key","score":30},{"value":"none","label":"No specific controls","score":0}]', true, 1.10, 'high', 'Controls should include CCTV, mantrap doors, and visitor logs.', false, 1, ARRAY['physical','access']),
('b0000001-0008-0000-0000-000000000001', 'Do you have environmental controls (fire suppression, UPS, HVAC) in data centres?', 'yes_no', '[{"value":"yes","label":"Yes","score":100},{"value":"no","label":"No","score":0}]', true, 0.90, 'medium', 'Controls should be regularly tested and maintained.', false, 2, ARRAY['physical','environmental']),
('b0000001-0008-0000-0000-000000000001', 'Is there a clean desk policy for areas handling sensitive information?', 'yes_no', '[{"value":"yes","label":"Yes","score":100},{"value":"no","label":"No","score":0}]', false, 0.50, 'low', 'Sensitive documents should not be left unattended.', false, 3, ARRAY['physical','clean-desk']),
('b0000001-0008-0000-0000-000000000001', 'Is media (hard drives, tapes, USB) securely destroyed when decommissioned?', 'single_choice', '[{"value":"certified","label":"Yes, certified destruction","score":100},{"value":"internal","label":"Yes, internal destruction","score":70},{"value":"wiped","label":"Wiped but not destroyed","score":40},{"value":"no","label":"No secure disposal","score":0}]', true, 0.90, 'high', 'Destruction certificates should be retained for audit purposes.', true, 4, ARRAY['physical','media-destruction']);

-- Section 9: Third-Party / Supply Chain
INSERT INTO questionnaire_sections (id, questionnaire_id, name, description, sort_order, weight, framework_domain_code)
VALUES ('b0000001-0009-0000-0000-000000000001', 'a0000001-0000-0000-0000-000000000001',
    'Third-Party / Supply Chain Security', 'Sub-processor management, vendor risk assessment', 9, 1.00, 'ISO27001-A15');

INSERT INTO questionnaire_questions (section_id, question_text, question_type, options, is_required, weight, risk_impact, guidance_text, evidence_required, sort_order, tags) VALUES
('b0000001-0009-0000-0000-000000000001', 'Do you maintain a register of all sub-processors who may access our data?', 'single_choice', '[{"value":"yes_notified","label":"Yes, with change notification process","score":100},{"value":"yes","label":"Yes, maintained","score":70},{"value":"partial","label":"Partial register","score":30},{"value":"no","label":"No","score":0}]', true, 1.30, 'critical', 'GDPR Article 28 requires disclosure and prior consent for sub-processor changes.', true, 1, ARRAY['supply-chain','sub-processors']),
('b0000001-0009-0000-0000-000000000001', 'Do you assess your own third-party vendors for security?', 'single_choice', '[{"value":"formal","label":"Yes, formal assessment programme","score":100},{"value":"basic","label":"Yes, basic due diligence","score":60},{"value":"adhoc","label":"Ad-hoc only","score":30},{"value":"no","label":"No vendor assessment","score":0}]', true, 1.10, 'high', 'Your supply chain risk directly impacts our data security.', false, 2, ARRAY['supply-chain','assessment']),
('b0000001-0009-0000-0000-000000000001', 'Do contracts with your sub-processors include security and data protection requirements?', 'yes_no', '[{"value":"yes","label":"Yes","score":100},{"value":"no","label":"No","score":0}]', true, 1.00, 'high', 'Contracts should flow down equivalent security obligations.', false, 3, ARRAY['supply-chain','contracts']),
('b0000001-0009-0000-0000-000000000001', 'Can you provide a list of countries where our data may be processed or stored?', 'yes_no', '[{"value":"yes","label":"Yes","score":100},{"value":"no","label":"No","score":0}]', true, 1.10, 'high', 'Data residency is critical for GDPR and other jurisdictional requirements.', false, 4, ARRAY['supply-chain','data-residency']),
('b0000001-0009-0000-0000-000000000001', 'Do you have contractual rights to audit your sub-processors?', 'yes_no', '[{"value":"yes","label":"Yes","score":100},{"value":"no","label":"No","score":0}]', true, 0.80, 'medium', 'Audit rights ensure ongoing oversight of the supply chain.', false, 5, ARRAY['supply-chain','audit-rights']);

-- Section 10: Cloud Security
INSERT INTO questionnaire_sections (id, questionnaire_id, name, description, sort_order, weight, framework_domain_code)
VALUES ('b0000001-0010-0000-0000-000000000001', 'a0000001-0000-0000-0000-000000000001',
    'Cloud Security', 'Cloud architecture, shared responsibility, configuration management', 10, 1.00, 'NIST_CSF_2-PR');

INSERT INTO questionnaire_questions (section_id, question_text, question_type, options, is_required, weight, risk_impact, guidance_text, evidence_required, sort_order, tags) VALUES
('b0000001-0010-0000-0000-000000000001', 'Which cloud providers do you use for processing our data?', 'text', '[]', true, 0.50, 'informational', 'We need to understand the shared responsibility model that applies.', false, 1, ARRAY['cloud','providers']),
('b0000001-0010-0000-0000-000000000001', 'Do you use infrastructure-as-code (IaC) with security scanning?', 'single_choice', '[{"value":"iac_scanned","label":"Yes, IaC with security scanning","score":100},{"value":"iac_only","label":"IaC without scanning","score":60},{"value":"manual","label":"Manual configuration","score":20},{"value":"unknown","label":"Unknown","score":0}]', true, 0.90, 'medium', 'IaC ensures reproducible, auditable infrastructure deployments.', false, 2, ARRAY['cloud','iac']),
('b0000001-0010-0000-0000-000000000001', 'Are cloud configurations continuously monitored for drift and misconfigurations?', 'single_choice', '[{"value":"cspm","label":"Yes, CSPM tool deployed","score":100},{"value":"manual","label":"Manual reviews","score":50},{"value":"none","label":"No monitoring","score":0}]', true, 1.10, 'high', 'Tools like AWS Config, Azure Policy, or third-party CSPM should be in place.', false, 3, ARRAY['cloud','cspm']),
('b0000001-0010-0000-0000-000000000001', 'Do you enforce principle of least privilege for cloud IAM roles?', 'yes_no', '[{"value":"yes","label":"Yes","score":100},{"value":"no","label":"No","score":0}]', true, 1.10, 'high', 'Over-provisioned IAM roles are a leading cause of cloud breaches.', false, 4, ARRAY['cloud','iam']),
('b0000001-0010-0000-0000-000000000001', 'Is logging enabled for all cloud services (CloudTrail, Activity Log, Audit Log)?', 'yes_no', '[{"value":"yes","label":"Yes","score":100},{"value":"no","label":"No","score":0}]', true, 1.00, 'high', 'Logs should be sent to a centralised SIEM and retained for at least 12 months.', false, 5, ARRAY['cloud','logging']),
('b0000001-0010-0000-0000-000000000001', 'Do you use container security scanning in your CI/CD pipeline?', 'single_choice', '[{"value":"integrated","label":"Yes, integrated in CI/CD","score":100},{"value":"manual","label":"Manual scanning","score":50},{"value":"none","label":"No container scanning","score":0}]', false, 0.80, 'medium', 'Container images should be scanned before deployment and in registries.', false, 6, ARRAY['cloud','containers']);

-- Section 11: Compliance & Certifications
INSERT INTO questionnaire_sections (id, questionnaire_id, name, description, sort_order, weight, framework_domain_code)
VALUES ('b0000001-0011-0000-0000-000000000001', 'a0000001-0000-0000-0000-000000000001',
    'Compliance & Certifications', 'Industry certifications, audits, and regulatory compliance', 11, 0.80, 'ISO27001-A18');

INSERT INTO questionnaire_questions (section_id, question_text, question_type, options, is_required, weight, risk_impact, guidance_text, evidence_required, sort_order, tags) VALUES
('b0000001-0011-0000-0000-000000000001', 'Do you hold ISO 27001 certification?', 'single_choice', '[{"value":"certified","label":"Yes, currently certified","score":100},{"value":"in_progress","label":"Certification in progress","score":50},{"value":"aligned","label":"Aligned but not certified","score":30},{"value":"no","label":"No","score":0}]', true, 1.20, 'high', 'ISO 27001 provides assurance of a systematic approach to information security.', true, 1, ARRAY['compliance','iso27001']),
('b0000001-0011-0000-0000-000000000001', 'Do you hold SOC 2 Type II certification?', 'single_choice', '[{"value":"type2","label":"SOC 2 Type II","score":100},{"value":"type1","label":"SOC 2 Type I","score":60},{"value":"in_progress","label":"In progress","score":30},{"value":"no","label":"No","score":0}]', true, 1.10, 'high', 'SOC 2 Type II provides evidence of controls operating effectively over time.', true, 2, ARRAY['compliance','soc2']),
('b0000001-0011-0000-0000-000000000001', 'Have you undergone an independent security audit in the past 12 months?', 'yes_no', '[{"value":"yes","label":"Yes","score":100},{"value":"no","label":"No","score":0}]', true, 1.00, 'high', 'Audit reports should be available upon request under NDA.', true, 3, ARRAY['compliance','audit']),
('b0000001-0011-0000-0000-000000000001', 'Are you compliant with PCI DSS (if processing payment data)?', 'single_choice', '[{"value":"certified","label":"PCI DSS certified","score":100},{"value":"saq","label":"SAQ completed","score":70},{"value":"na","label":"Not applicable","score":100},{"value":"no","label":"Not compliant","score":0}]', false, 0.80, 'medium', 'Required if any payment card data is in scope.', false, 4, ARRAY['compliance','pci']),
('b0000001-0011-0000-0000-000000000001', 'Do you hold any additional certifications relevant to our engagement?', 'text', '[]', false, 0.50, 'informational', 'Examples: Cyber Essentials Plus, NCSC CAF, HITRUST, FedRAMP, etc.', false, 5, ARRAY['compliance','other']),
('b0000001-0011-0000-0000-000000000001', 'Are you willing to provide your latest penetration test executive summary under NDA?', 'yes_no', '[{"value":"yes","label":"Yes","score":100},{"value":"no","label":"No","score":0}]', false, 0.70, 'medium', 'This helps us understand your current vulnerability posture.', false, 6, ARRAY['compliance','pentest-report']),
('b0000001-0011-0000-0000-000000000001', 'Do you have a legal/regulatory compliance monitoring programme?', 'yes_no', '[{"value":"yes","label":"Yes","score":100},{"value":"no","label":"No","score":0}]', true, 0.80, 'medium', 'Monitoring should cover applicable laws, regulations, and contractual obligations.', false, 7, ARRAY['compliance','monitoring']);


-- ============================================================
-- TEMPLATE 2: GDPR Article 28 Processor Assessment
-- ============================================================

INSERT INTO assessment_questionnaires (
    id, organization_id, name, description, questionnaire_type, version, status,
    total_questions, total_sections, estimated_completion_minutes,
    scoring_method, pass_threshold, risk_tier_thresholds,
    applicable_vendor_tiers, is_system, is_template
) VALUES (
    'a0000002-0000-0000-0000-000000000001', NULL,
    'GDPR Article 28 Processor Assessment',
    'Assessment specifically designed to evaluate compliance with GDPR Article 28 requirements for data processors. Covers lawful processing, data subject rights, security measures, sub-processor management, international transfers, DPO, and breach notification.',
    'privacy', 1, 'active',
    40, 7, 60,
    'weighted_average', 75.00,
    '{"critical": 50, "high": 60, "medium": 75, "low": 90}',
    ARRAY['critical', 'high', 'medium'],
    true, true
);

-- GDPR Section 1: Lawful Processing Basis
INSERT INTO questionnaire_sections (id, questionnaire_id, name, description, sort_order, weight)
VALUES ('b0000002-0001-0000-0000-000000000001', 'a0000002-0000-0000-0000-000000000001',
    'Lawful Processing Basis', 'Ensuring processing is only on documented instructions', 1, 1.20);

INSERT INTO questionnaire_questions (section_id, question_text, question_type, options, is_required, weight, risk_impact, guidance_text, evidence_required, sort_order, tags) VALUES
('b0000002-0001-0000-0000-000000000001', 'Do you process personal data only on the documented instructions of the controller?', 'single_choice', '[{"value":"always","label":"Always, with documented procedures","score":100},{"value":"mostly","label":"Generally, with some exceptions","score":60},{"value":"unclear","label":"Not clearly defined","score":20},{"value":"no","label":"No","score":0}]', true, 1.50, 'critical', 'Article 28(3)(a): Processor shall process only on documented instructions.', true, 1, ARRAY['gdpr','art28','instructions']),
('b0000002-0001-0000-0000-000000000001', 'Do you maintain a Record of Processing Activities (ROPA) as required by Article 30?', 'single_choice', '[{"value":"complete","label":"Yes, complete and current","score":100},{"value":"partial","label":"Partial ROPA","score":50},{"value":"no","label":"No ROPA maintained","score":0}]', true, 1.30, 'critical', 'Article 30(2) requires processors to maintain records of processing activities.', true, 2, ARRAY['gdpr','art30','ropa']),
('b0000002-0001-0000-0000-000000000001', 'Is there a Data Processing Agreement (DPA) in place that meets Article 28(3) requirements?', 'yes_no', '[{"value":"yes","label":"Yes","score":100},{"value":"no","label":"No","score":0}]', true, 1.50, 'critical', 'The DPA must cover all elements listed in Article 28(3).', true, 3, ARRAY['gdpr','art28','dpa']),
('b0000002-0001-0000-0000-000000000001', 'Do you have processes to identify when processing requires a Data Protection Impact Assessment?', 'yes_no', '[{"value":"yes","label":"Yes","score":100},{"value":"no","label":"No","score":0}]', true, 1.00, 'high', 'Article 35 requires DPIAs for high-risk processing.', false, 4, ARRAY['gdpr','art35','dpia']),
('b0000002-0001-0000-0000-000000000001', 'Do you implement data protection by design and by default?', 'single_choice', '[{"value":"systematic","label":"Yes, systematic approach","score":100},{"value":"adhoc","label":"Ad-hoc consideration","score":50},{"value":"no","label":"No","score":0}]', true, 1.20, 'high', 'Article 25 requires appropriate technical and organisational measures.', false, 5, ARRAY['gdpr','art25','by-design']),
('b0000002-0001-0000-0000-000000000001', 'Can you demonstrate the lawful basis for each category of personal data you process on our behalf?', 'yes_no', '[{"value":"yes","label":"Yes","score":100},{"value":"no","label":"No","score":0}]', true, 1.10, 'critical', 'The controller must establish lawful basis; the processor must respect it.', false, 6, ARRAY['gdpr','lawful-basis']);

-- GDPR Section 2: Data Subject Rights
INSERT INTO questionnaire_sections (id, questionnaire_id, name, description, sort_order, weight)
VALUES ('b0000002-0002-0000-0000-000000000001', 'a0000002-0000-0000-0000-000000000001',
    'Data Subject Rights', 'Support for fulfilling data subject requests', 2, 1.10);

INSERT INTO questionnaire_questions (section_id, question_text, question_type, options, is_required, weight, risk_impact, guidance_text, evidence_required, sort_order, tags) VALUES
('b0000002-0002-0000-0000-000000000001', 'Can you assist the controller in responding to data subject access requests (DSARs)?', 'single_choice', '[{"value":"automated","label":"Yes, automated extraction capability","score":100},{"value":"manual","label":"Yes, manual process in place","score":70},{"value":"limited","label":"Limited capability","score":30},{"value":"no","label":"No capability","score":0}]', true, 1.30, 'critical', 'Article 28(3)(e): Processor must assist with data subject right obligations.', false, 1, ARRAY['gdpr','dsar','art28']),
('b0000002-0002-0000-0000-000000000001', 'Can you facilitate the right to erasure (right to be forgotten)?', 'single_choice', '[{"value":"full","label":"Yes, full erasure capability","score":100},{"value":"partial","label":"Partial (some data cannot be deleted)","score":50},{"value":"no","label":"No erasure capability","score":0}]', true, 1.30, 'critical', 'Article 17: Data subjects have the right to erasure in certain circumstances.', false, 2, ARRAY['gdpr','erasure','art17']),
('b0000002-0002-0000-0000-000000000001', 'Can you provide data in a structured, machine-readable format for portability requests?', 'single_choice', '[{"value":"api","label":"Yes, via API/export","score":100},{"value":"csv","label":"Yes, CSV/JSON export","score":80},{"value":"manual","label":"Manual extraction only","score":40},{"value":"no","label":"No portability capability","score":0}]', true, 1.00, 'high', 'Article 20: Right to data portability requires structured, machine-readable format.', false, 3, ARRAY['gdpr','portability','art20']),
('b0000002-0002-0000-0000-000000000001', 'What is your typical response time for data subject requests forwarded by the controller?', 'single_choice', '[{"value":"48h","label":"Within 48 hours","score":100},{"value":"1w","label":"Within 1 week","score":70},{"value":"1m","label":"Within 1 month","score":40},{"value":"undefined","label":"No defined SLA","score":0}]', true, 1.10, 'high', 'Controllers must respond within 1 month; processors should respond promptly.', false, 4, ARRAY['gdpr','dsar','response-time']),
('b0000002-0002-0000-0000-000000000001', 'Can you restrict processing of specific data when requested by the controller?', 'yes_no', '[{"value":"yes","label":"Yes","score":100},{"value":"no","label":"No","score":0}]', true, 1.00, 'high', 'Article 18: Right to restriction of processing must be supported.', false, 5, ARRAY['gdpr','restriction','art18']),
('b0000002-0002-0000-0000-000000000001', 'Do you have processes to handle objections to automated decision-making and profiling?', 'single_choice', '[{"value":"yes","label":"Yes, documented process","score":100},{"value":"na","label":"Not applicable — no automated decisions","score":100},{"value":"no","label":"No process","score":0}]', false, 0.80, 'medium', 'Article 22: Right not to be subject to solely automated decisions.', false, 6, ARRAY['gdpr','automated-decisions','art22']);

-- GDPR Section 3: Security Measures (Article 32)
INSERT INTO questionnaire_sections (id, questionnaire_id, name, description, sort_order, weight)
VALUES ('b0000002-0003-0000-0000-000000000001', 'a0000002-0000-0000-0000-000000000001',
    'Security Measures (Article 32)', 'Technical and organisational measures for data security', 3, 1.30);

INSERT INTO questionnaire_questions (section_id, question_text, question_type, options, is_required, weight, risk_impact, guidance_text, evidence_required, sort_order, tags) VALUES
('b0000002-0003-0000-0000-000000000001', 'Do you pseudonymise personal data where possible?', 'single_choice', '[{"value":"systematic","label":"Yes, systematically applied","score":100},{"value":"selective","label":"Selectively applied","score":60},{"value":"no","label":"No pseudonymisation","score":0}]', true, 1.10, 'high', 'Article 32(1)(a): Pseudonymisation as a security measure.', false, 1, ARRAY['gdpr','art32','pseudonymisation']),
('b0000002-0003-0000-0000-000000000001', 'Do you encrypt personal data at rest and in transit?', 'single_choice', '[{"value":"both","label":"Yes, both at rest and in transit","score":100},{"value":"transit","label":"In transit only","score":50},{"value":"rest","label":"At rest only","score":50},{"value":"no","label":"No encryption","score":0}]', true, 1.40, 'critical', 'Article 32(1)(a): Encryption is an explicit security measure.', true, 2, ARRAY['gdpr','art32','encryption']),
('b0000002-0003-0000-0000-000000000001', 'Can you ensure ongoing confidentiality, integrity, availability, and resilience of processing systems?', 'single_choice', '[{"value":"comprehensive","label":"Yes, comprehensive controls","score":100},{"value":"partial","label":"Partial controls","score":50},{"value":"no","label":"No specific measures","score":0}]', true, 1.30, 'critical', 'Article 32(1)(b): Ensuring CIA triad and resilience.', false, 3, ARRAY['gdpr','art32','cia']),
('b0000002-0003-0000-0000-000000000001', 'Do you have the ability to restore access to personal data in a timely manner after an incident?', 'single_choice', '[{"value":"automated","label":"Yes, automated recovery","score":100},{"value":"manual","label":"Yes, manual recovery tested","score":70},{"value":"untested","label":"Backup exists but untested","score":30},{"value":"no","label":"No recovery capability","score":0}]', true, 1.20, 'critical', 'Article 32(1)(c): Ability to restore availability and access.', true, 4, ARRAY['gdpr','art32','recovery']),
('b0000002-0003-0000-0000-000000000001', 'Do you regularly test and evaluate the effectiveness of security measures?', 'single_choice', '[{"value":"regular","label":"Yes, regularly with documented results","score":100},{"value":"annual","label":"Annually","score":70},{"value":"adhoc","label":"Ad-hoc","score":30},{"value":"no","label":"No testing","score":0}]', true, 1.20, 'high', 'Article 32(1)(d): Regular testing and evaluation of measures.', true, 5, ARRAY['gdpr','art32','testing']),
('b0000002-0003-0000-0000-000000000001', 'Do your employees and contractors with access to personal data sign confidentiality agreements?', 'yes_no', '[{"value":"yes","label":"Yes","score":100},{"value":"no","label":"No","score":0}]', true, 1.10, 'high', 'Article 28(3)(b): Persons authorised to process must commit to confidentiality.', true, 6, ARRAY['gdpr','art28','confidentiality']);

-- GDPR Section 4: Sub-Processor Management
INSERT INTO questionnaire_sections (id, questionnaire_id, name, description, sort_order, weight)
VALUES ('b0000002-0004-0000-0000-000000000001', 'a0000002-0000-0000-0000-000000000001',
    'Sub-Processor Management', 'Management of sub-processors per Article 28(2)', 4, 1.20);

INSERT INTO questionnaire_questions (section_id, question_text, question_type, options, is_required, weight, risk_impact, guidance_text, evidence_required, sort_order, tags) VALUES
('b0000002-0004-0000-0000-000000000001', 'Do you obtain prior specific or general written authorisation before engaging sub-processors?', 'single_choice', '[{"value":"specific","label":"Prior specific authorisation","score":100},{"value":"general","label":"General authorisation with notification","score":80},{"value":"informal","label":"Informal notification","score":30},{"value":"none","label":"No authorisation sought","score":0}]', true, 1.40, 'critical', 'Article 28(2): Prior authorisation required for sub-processors.', true, 1, ARRAY['gdpr','art28','sub-processor']),
('b0000002-0004-0000-0000-000000000001', 'Do you impose the same data protection obligations on sub-processors via contract?', 'yes_no', '[{"value":"yes","label":"Yes","score":100},{"value":"no","label":"No","score":0}]', true, 1.40, 'critical', 'Article 28(4): Same obligations must be imposed on sub-processors.', true, 2, ARRAY['gdpr','art28','flow-down']),
('b0000002-0004-0000-0000-000000000001', 'Can you provide a complete, current list of sub-processors?', 'yes_no', '[{"value":"yes","label":"Yes","score":100},{"value":"no","label":"No","score":0}]', true, 1.20, 'critical', 'Transparency about sub-processor chain is essential.', true, 3, ARRAY['gdpr','sub-processor','list']),
('b0000002-0004-0000-0000-000000000001', 'Do you have a mechanism to notify us of intended sub-processor changes?', 'single_choice', '[{"value":"30d_advance","label":"30+ days advance notice","score":100},{"value":"14d_advance","label":"14 days advance notice","score":80},{"value":"after_change","label":"Notification after change","score":30},{"value":"no","label":"No notification process","score":0}]', true, 1.10, 'high', 'Controller must have opportunity to object to sub-processor changes.', false, 4, ARRAY['gdpr','sub-processor','notification']),
('b0000002-0004-0000-0000-000000000001', 'Do you remain fully liable for the acts of your sub-processors?', 'yes_no', '[{"value":"yes","label":"Yes","score":100},{"value":"no","label":"No","score":0}]', true, 1.00, 'high', 'Article 28(4): The initial processor remains liable.', false, 5, ARRAY['gdpr','art28','liability']);

-- GDPR Section 5: International Transfers
INSERT INTO questionnaire_sections (id, questionnaire_id, name, description, sort_order, weight)
VALUES ('b0000002-0005-0000-0000-000000000001', 'a0000002-0000-0000-0000-000000000001',
    'International Transfers', 'Transfer mechanisms and adequacy for cross-border data flows', 5, 1.10);

INSERT INTO questionnaire_questions (section_id, question_text, question_type, options, is_required, weight, risk_impact, guidance_text, evidence_required, sort_order, tags) VALUES
('b0000002-0005-0000-0000-000000000001', 'Is personal data transferred outside the EEA/UK?', 'single_choice', '[{"value":"no","label":"No transfers outside EEA/UK","score":100},{"value":"adequacy","label":"Yes, to adequacy decision countries only","score":90},{"value":"sccs","label":"Yes, with Standard Contractual Clauses","score":70},{"value":"no_mechanism","label":"Yes, without appropriate safeguards","score":0}]', true, 1.40, 'critical', 'Chapter V GDPR: Transfer mechanisms required for international transfers.', true, 1, ARRAY['gdpr','transfers','chapter5']),
('b0000002-0005-0000-0000-000000000001', 'Have you conducted a Transfer Impact Assessment (TIA) for non-EEA transfers?', 'single_choice', '[{"value":"yes","label":"Yes, documented TIA","score":100},{"value":"partial","label":"Partial assessment","score":50},{"value":"na","label":"Not applicable — no non-EEA transfers","score":100},{"value":"no","label":"No TIA conducted","score":0}]', true, 1.20, 'critical', 'Schrems II ruling requires supplementary measures assessment.', true, 2, ARRAY['gdpr','tia','schrems']),
('b0000002-0005-0000-0000-000000000001', 'Do you use the EU Commission approved Standard Contractual Clauses (2021)?', 'single_choice', '[{"value":"2021","label":"Yes, 2021 SCCs","score":100},{"value":"old","label":"Using older SCCs","score":40},{"value":"na","label":"Not applicable","score":100},{"value":"no","label":"No SCCs in place","score":0}]', true, 1.10, 'high', 'Updated 2021 SCCs are required; old versions expired in December 2022.', true, 3, ARRAY['gdpr','sccs']),
('b0000002-0005-0000-0000-000000000001', 'What supplementary measures have you implemented for international transfers?', 'text', '[]', false, 0.80, 'medium', 'Examples: additional encryption, pseudonymisation, access controls, etc.', false, 4, ARRAY['gdpr','supplementary-measures']),
('b0000002-0005-0000-0000-000000000001', 'Can you guarantee that data will not be disclosed to third-country authorities without lawful basis?', 'yes_no', '[{"value":"yes","label":"Yes","score":100},{"value":"no","label":"No","score":0}]', true, 1.00, 'critical', 'Article 48: Transfers based on third-country judgments are prohibited without international agreements.', false, 5, ARRAY['gdpr','art48','disclosure']);

-- GDPR Section 6: Data Protection Officer
INSERT INTO questionnaire_sections (id, questionnaire_id, name, description, sort_order, weight)
VALUES ('b0000002-0006-0000-0000-000000000001', 'a0000002-0000-0000-0000-000000000001',
    'Data Protection Officer', 'DPO appointment and independence', 6, 0.90);

INSERT INTO questionnaire_questions (section_id, question_text, question_type, options, is_required, weight, risk_impact, guidance_text, evidence_required, sort_order, tags) VALUES
('b0000002-0006-0000-0000-000000000001', 'Have you appointed a Data Protection Officer (DPO)?', 'single_choice', '[{"value":"internal","label":"Yes, internal DPO","score":100},{"value":"external","label":"Yes, external DPO","score":90},{"value":"not_required","label":"Not required under Article 37","score":80},{"value":"no","label":"No, but required","score":0}]', true, 1.10, 'high', 'Article 37-39: DPO required for public bodies and large-scale processing.', false, 1, ARRAY['gdpr','dpo','art37']),
('b0000002-0006-0000-0000-000000000001', 'Does the DPO report directly to the highest management level?', 'single_choice', '[{"value":"yes","label":"Yes, reports to board/CEO","score":100},{"value":"indirect","label":"Indirect reporting line","score":60},{"value":"na","label":"No DPO appointed","score":0}]', false, 0.80, 'medium', 'Article 38(3): DPO must report to highest management level.', false, 2, ARRAY['gdpr','dpo','independence']),
('b0000002-0006-0000-0000-000000000001', 'Can you provide DPO contact details?', 'text', '[]', true, 0.50, 'informational', 'DPO should be accessible to controllers and data subjects.', false, 3, ARRAY['gdpr','dpo','contact']);

-- GDPR Section 7: Breach Notification
INSERT INTO questionnaire_sections (id, questionnaire_id, name, description, sort_order, weight)
VALUES ('b0000002-0007-0000-0000-000000000001', 'a0000002-0000-0000-0000-000000000001',
    'Breach Notification', 'Personal data breach detection and notification', 7, 1.30);

INSERT INTO questionnaire_questions (section_id, question_text, question_type, options, is_required, weight, risk_impact, guidance_text, evidence_required, sort_order, tags) VALUES
('b0000002-0007-0000-0000-000000000001', 'Can you notify the controller without undue delay upon becoming aware of a personal data breach?', 'single_choice', '[{"value":"under_24h","label":"Yes, within 24 hours","score":100},{"value":"under_48h","label":"Within 48 hours","score":80},{"value":"under_72h","label":"Within 72 hours","score":60},{"value":"no_guarantee","label":"No guaranteed timeframe","score":20}]', true, 1.50, 'critical', 'Article 33(2): Processor must notify controller without undue delay.', false, 1, ARRAY['gdpr','art33','breach-notification']),
('b0000002-0007-0000-0000-000000000001', 'Does your breach notification include all information required by Article 33(3)?', 'single_choice', '[{"value":"complete","label":"Yes, complete information","score":100},{"value":"partial","label":"Partial — additional info provided later","score":60},{"value":"basic","label":"Basic notification only","score":30},{"value":"no","label":"No structured notification","score":0}]', true, 1.20, 'critical', 'Must include nature of breach, DPO contact, likely consequences, and measures taken.', true, 2, ARRAY['gdpr','art33','content']),
('b0000002-0007-0000-0000-000000000001', 'Do you maintain a breach register documenting all personal data breaches?', 'yes_no', '[{"value":"yes","label":"Yes","score":100},{"value":"no","label":"No","score":0}]', true, 1.00, 'high', 'Article 33(5): Processor should document all breaches regardless of severity.', false, 3, ARRAY['gdpr','art33','register']),
('b0000002-0007-0000-0000-000000000001', 'Do you have an incident response team trained specifically on GDPR breach procedures?', 'yes_no', '[{"value":"yes","label":"Yes","score":100},{"value":"no","label":"No","score":0}]', true, 1.10, 'high', 'Team should understand the 72-hour supervisory authority notification requirement.', false, 4, ARRAY['gdpr','incident-response','team']),
('b0000002-0007-0000-0000-000000000001', 'Can you assist the controller in communicating a breach to affected data subjects?', 'yes_no', '[{"value":"yes","label":"Yes","score":100},{"value":"no","label":"No","score":0}]', true, 0.90, 'high', 'Article 34: High-risk breaches require data subject communication.', false, 5, ARRAY['gdpr','art34','communication']);


-- ============================================================
-- TEMPLATE 3: NIS2 Supply Chain Security Assessment
-- ============================================================

INSERT INTO assessment_questionnaires (
    id, organization_id, name, description, questionnaire_type, version, status,
    total_questions, total_sections, estimated_completion_minutes,
    scoring_method, pass_threshold, risk_tier_thresholds,
    applicable_vendor_tiers, is_system, is_template
) VALUES (
    'a0000003-0000-0000-0000-000000000001', NULL,
    'NIS2 Supply Chain Security Assessment',
    'Assessment for suppliers of essential and important entities under the NIS2 Directive (EU 2022/2555). Covers supply chain risk management, incident reporting, security measures, governance, and resilience.',
    'compliance', 1, 'active',
    30, 5, 45,
    'weighted_average', 70.00,
    '{"critical": 40, "high": 55, "medium": 70, "low": 85}',
    ARRAY['critical', 'high'],
    true, true
);

-- NIS2 Section 1: Governance & Risk Management
INSERT INTO questionnaire_sections (id, questionnaire_id, name, description, sort_order, weight)
VALUES ('b0000003-0001-0000-0000-000000000001', 'a0000003-0000-0000-0000-000000000001',
    'Governance & Risk Management', 'Security governance, risk management processes', 1, 1.20);

INSERT INTO questionnaire_questions (section_id, question_text, question_type, options, is_required, weight, risk_impact, guidance_text, evidence_required, sort_order, tags) VALUES
('b0000003-0001-0000-0000-000000000001', 'Does your management body approve and oversee cybersecurity risk management measures?', 'single_choice', '[{"value":"board","label":"Yes, board-level oversight","score":100},{"value":"exec","label":"Executive committee oversight","score":80},{"value":"it_only","label":"IT management only","score":40},{"value":"none","label":"No management oversight","score":0}]', true, 1.40, 'critical', 'NIS2 Article 20(1): Management bodies must approve cybersecurity measures.', true, 1, ARRAY['nis2','governance','art20']),
('b0000003-0001-0000-0000-000000000001', 'Do members of management undergo regular cybersecurity training?', 'single_choice', '[{"value":"regular","label":"Yes, at least annually","score":100},{"value":"onboarding","label":"Onboarding only","score":40},{"value":"no","label":"No management training","score":0}]', true, 1.20, 'high', 'NIS2 Article 20(2): Management training is mandatory.', false, 2, ARRAY['nis2','training','art20']),
('b0000003-0001-0000-0000-000000000001', 'Do you have a formal cybersecurity risk management framework?', 'single_choice', '[{"value":"iso","label":"Yes, ISO 27001 or equivalent","score":100},{"value":"nist","label":"Yes, NIST CSF or similar","score":90},{"value":"custom","label":"Custom framework","score":60},{"value":"none","label":"No framework","score":0}]', true, 1.30, 'critical', 'NIS2 Article 21: Appropriate and proportionate measures required.', true, 3, ARRAY['nis2','risk-mgmt','art21']),
('b0000003-0001-0000-0000-000000000001', 'Are cybersecurity policies reviewed and updated at least annually?', 'single_choice', '[{"value":"yes","label":"Yes, annually or more","score":100},{"value":"adhoc","label":"Ad-hoc reviews only","score":40},{"value":"no","label":"No regular reviews","score":0}]', true, 1.00, 'high', 'Policies should reflect current threat landscape and organisational changes.', false, 4, ARRAY['nis2','policy','review']),
('b0000003-0001-0000-0000-000000000001', 'Do you conduct regular risk assessments of your network and information systems?', 'single_choice', '[{"value":"continuous","label":"Continuous monitoring","score":100},{"value":"quarterly","label":"Quarterly","score":80},{"value":"annually","label":"Annually","score":60},{"value":"never","label":"Never","score":0}]', true, 1.20, 'critical', 'NIS2 Article 21(2)(a): Policies on risk analysis and information system security.', false, 5, ARRAY['nis2','risk-assessment','art21']),
('b0000003-0001-0000-0000-000000000001', 'Do you have a designated CISO or equivalent cybersecurity leadership role?', 'yes_no', '[{"value":"yes","label":"Yes","score":100},{"value":"no","label":"No","score":0}]', true, 1.00, 'high', 'Clear cybersecurity leadership is essential for NIS2 compliance.', false, 6, ARRAY['nis2','ciso','leadership']);

-- NIS2 Section 2: Incident Handling
INSERT INTO questionnaire_sections (id, questionnaire_id, name, description, sort_order, weight)
VALUES ('b0000003-0002-0000-0000-000000000001', 'a0000003-0000-0000-0000-000000000001',
    'Incident Handling', 'Detection, response, reporting, and recovery', 2, 1.30);

INSERT INTO questionnaire_questions (section_id, question_text, question_type, options, is_required, weight, risk_impact, guidance_text, evidence_required, sort_order, tags) VALUES
('b0000003-0002-0000-0000-000000000001', 'Do you have incident handling and response procedures aligned with NIS2 requirements?', 'single_choice', '[{"value":"nis2_aligned","label":"Yes, NIS2-aligned procedures","score":100},{"value":"generic","label":"Generic incident response plan","score":60},{"value":"none","label":"No formal procedures","score":0}]', true, 1.40, 'critical', 'NIS2 Article 21(2)(b): Incident handling procedures required.', true, 1, ARRAY['nis2','incident','art21']),
('b0000003-0002-0000-0000-000000000001', 'Can you provide an early warning to affected clients within 24 hours of a significant incident?', 'single_choice', '[{"value":"yes","label":"Yes, within 24 hours","score":100},{"value":"48h","label":"Within 48 hours","score":60},{"value":"no_guarantee","label":"No guaranteed timeframe","score":0}]', true, 1.50, 'critical', 'NIS2 Article 23(4)(a): Early warning within 24 hours.', false, 2, ARRAY['nis2','early-warning','art23']),
('b0000003-0002-0000-0000-000000000001', 'Can you provide an incident notification with impact assessment within 72 hours?', 'single_choice', '[{"value":"yes","label":"Yes, within 72 hours","score":100},{"value":"best_effort","label":"Best effort","score":50},{"value":"no","label":"No","score":0}]', true, 1.40, 'critical', 'NIS2 Article 23(4)(b): Incident notification within 72 hours.', false, 3, ARRAY['nis2','notification','art23']),
('b0000003-0002-0000-0000-000000000001', 'Do you have 24/7 security monitoring capabilities?', 'single_choice', '[{"value":"soc_247","label":"Yes, 24/7 SOC","score":100},{"value":"business_hours","label":"Business hours only","score":50},{"value":"none","label":"No active monitoring","score":0}]', true, 1.20, 'high', 'Continuous monitoring enables rapid incident detection.', false, 4, ARRAY['nis2','monitoring','soc']),
('b0000003-0002-0000-0000-000000000001', 'Can you provide a final incident report within one month of notification?', 'yes_no', '[{"value":"yes","label":"Yes","score":100},{"value":"no","label":"No","score":0}]', true, 1.10, 'high', 'NIS2 Article 23(4)(d): Final report within one month.', false, 5, ARRAY['nis2','final-report','art23']),
('b0000003-0002-0000-0000-000000000001', 'Do you participate in threat intelligence sharing with industry peers or CERTs?', 'single_choice', '[{"value":"active","label":"Yes, active participation","score":100},{"value":"passive","label":"Passive consumption only","score":50},{"value":"no","label":"No participation","score":0}]', true, 0.80, 'medium', 'NIS2 Article 29: Voluntary cyber threat intelligence sharing.', false, 6, ARRAY['nis2','threat-intel','art29']);

-- NIS2 Section 3: Supply Chain Security
INSERT INTO questionnaire_sections (id, questionnaire_id, name, description, sort_order, weight)
VALUES ('b0000003-0003-0000-0000-000000000001', 'a0000003-0000-0000-0000-000000000001',
    'Supply Chain Security', 'Management of your own supply chain and sub-contractors', 3, 1.20);

INSERT INTO questionnaire_questions (section_id, question_text, question_type, options, is_required, weight, risk_impact, guidance_text, evidence_required, sort_order, tags) VALUES
('b0000003-0003-0000-0000-000000000001', 'Do you assess cybersecurity risks in your own supply chain?', 'single_choice', '[{"value":"formal","label":"Yes, formal programme","score":100},{"value":"basic","label":"Basic due diligence","score":60},{"value":"none","label":"No supply chain assessment","score":0}]', true, 1.30, 'critical', 'NIS2 Article 21(2)(d): Supply chain security required.', true, 1, ARRAY['nis2','supply-chain','art21']),
('b0000003-0003-0000-0000-000000000001', 'Do contracts with your suppliers include cybersecurity requirements?', 'yes_no', '[{"value":"yes","label":"Yes","score":100},{"value":"no","label":"No","score":0}]', true, 1.10, 'high', 'Contractual security obligations should be flowed down.', false, 2, ARRAY['nis2','contracts','supply-chain']),
('b0000003-0003-0000-0000-000000000001', 'Do you verify the cybersecurity posture of your critical suppliers?', 'single_choice', '[{"value":"regular","label":"Yes, regularly assessed","score":100},{"value":"initial","label":"Initial assessment only","score":50},{"value":"no","label":"No verification","score":0}]', true, 1.20, 'high', 'Ongoing verification ensures suppliers maintain security standards.', false, 3, ARRAY['nis2','supplier-verification']),
('b0000003-0003-0000-0000-000000000001', 'Do you have a software/hardware bill of materials for products/services provided?', 'single_choice', '[{"value":"sbom","label":"Yes, SBOM maintained","score":100},{"value":"partial","label":"Partial inventory","score":50},{"value":"no","label":"No BOM","score":0}]', true, 1.00, 'high', 'SBOM transparency helps identify supply chain vulnerabilities.', false, 4, ARRAY['nis2','sbom','supply-chain']),
('b0000003-0003-0000-0000-000000000001', 'Can you notify us of supply chain incidents that may affect our security?', 'yes_no', '[{"value":"yes","label":"Yes","score":100},{"value":"no","label":"No","score":0}]', true, 1.10, 'critical', 'Cascading supply chain incidents require prompt notification.', false, 5, ARRAY['nis2','supply-chain','notification']),
('b0000003-0003-0000-0000-000000000001', 'Do you have processes to manage coordinated vulnerability disclosure?', 'yes_no', '[{"value":"yes","label":"Yes","score":100},{"value":"no","label":"No","score":0}]', true, 0.90, 'medium', 'NIS2 Article 12: Coordinated vulnerability disclosure framework.', false, 6, ARRAY['nis2','art12','vulnerability-disclosure']);

-- NIS2 Section 4: Technical Security Measures
INSERT INTO questionnaire_sections (id, questionnaire_id, name, description, sort_order, weight)
VALUES ('b0000003-0004-0000-0000-000000000001', 'a0000003-0000-0000-0000-000000000001',
    'Technical Security Measures', 'Specific technical controls required by NIS2', 4, 1.10);

INSERT INTO questionnaire_questions (section_id, question_text, question_type, options, is_required, weight, risk_impact, guidance_text, evidence_required, sort_order, tags) VALUES
('b0000003-0004-0000-0000-000000000001', 'Do you employ multi-factor authentication for all critical system access?', 'single_choice', '[{"value":"all","label":"Yes, all critical systems","score":100},{"value":"some","label":"Some critical systems","score":50},{"value":"none","label":"No MFA","score":0}]', true, 1.30, 'critical', 'NIS2 Article 21(2)(j): Multi-factor authentication required.', false, 1, ARRAY['nis2','mfa','art21']),
('b0000003-0004-0000-0000-000000000001', 'Do you use cryptography and encryption as appropriate?', 'single_choice', '[{"value":"comprehensive","label":"Yes, comprehensive use","score":100},{"value":"partial","label":"Partial implementation","score":50},{"value":"none","label":"No cryptographic controls","score":0}]', true, 1.20, 'critical', 'NIS2 Article 21(2)(h): Policies on the use of cryptography and encryption.', false, 2, ARRAY['nis2','cryptography','art21']),
('b0000003-0004-0000-0000-000000000001', 'Do you implement network segmentation for critical services?', 'yes_no', '[{"value":"yes","label":"Yes","score":100},{"value":"no","label":"No","score":0}]', true, 1.10, 'high', 'Segmentation limits lateral movement in case of compromise.', false, 3, ARRAY['nis2','segmentation']),
('b0000003-0004-0000-0000-000000000001', 'Do you have vulnerability management and patching processes?', 'single_choice', '[{"value":"automated","label":"Yes, automated scanning and patching","score":100},{"value":"manual","label":"Manual process","score":60},{"value":"adhoc","label":"Ad-hoc","score":20},{"value":"none","label":"No formal process","score":0}]', true, 1.20, 'critical', 'NIS2 Article 21(2)(e): Vulnerability handling and disclosure.', true, 4, ARRAY['nis2','vulnerability-mgmt','art21']),
('b0000003-0004-0000-0000-000000000001', 'Do you have secure communications channels for emergency and crisis management?', 'yes_no', '[{"value":"yes","label":"Yes","score":100},{"value":"no","label":"No","score":0}]', true, 0.90, 'high', 'NIS2 Article 21(2)(j): Secured voice, video, and text communications.', false, 5, ARRAY['nis2','secure-comms','art21']),
('b0000003-0004-0000-0000-000000000001', 'Do you conduct regular security awareness training for all staff?', 'single_choice', '[{"value":"regular","label":"Yes, regular training with testing","score":100},{"value":"annual","label":"Annual training","score":70},{"value":"onboarding","label":"Onboarding only","score":30},{"value":"none","label":"No training","score":0}]', true, 1.00, 'high', 'NIS2 Article 21(2)(g): Basic cyber hygiene practices and training.', false, 6, ARRAY['nis2','training','art21']);

-- NIS2 Section 5: Business Continuity & Resilience
INSERT INTO questionnaire_sections (id, questionnaire_id, name, description, sort_order, weight)
VALUES ('b0000003-0005-0000-0000-000000000001', 'a0000003-0000-0000-0000-000000000001',
    'Business Continuity & Resilience', 'Continuity planning and crisis management', 5, 1.00);

INSERT INTO questionnaire_questions (section_id, question_text, question_type, options, is_required, weight, risk_impact, guidance_text, evidence_required, sort_order, tags) VALUES
('b0000003-0005-0000-0000-000000000001', 'Do you have a business continuity plan that covers the services provided to us?', 'single_choice', '[{"value":"tested","label":"Yes, tested within 12 months","score":100},{"value":"documented","label":"Yes, documented","score":60},{"value":"none","label":"No BCP","score":0}]', true, 1.20, 'high', 'NIS2 Article 21(2)(c): Business continuity and crisis management.', true, 1, ARRAY['nis2','bcp','art21']),
('b0000003-0005-0000-0000-000000000001', 'Do you have backup management and disaster recovery procedures?', 'single_choice', '[{"value":"tested","label":"Yes, tested regularly","score":100},{"value":"documented","label":"Documented but not tested","score":50},{"value":"none","label":"No DR procedures","score":0}]', true, 1.10, 'high', 'NIS2 Article 21(2)(c): Backup management and disaster recovery required.', true, 2, ARRAY['nis2','dr','backup']),
('b0000003-0005-0000-0000-000000000001', 'Do you have a crisis management framework with clear escalation procedures?', 'yes_no', '[{"value":"yes","label":"Yes","score":100},{"value":"no","label":"No","score":0}]', true, 1.00, 'high', 'Crisis management should include communication plans for all stakeholders.', false, 3, ARRAY['nis2','crisis-mgmt']),
('b0000003-0005-0000-0000-000000000001', 'What is your target Recovery Time Objective (RTO) for services provided?', 'single_choice', '[{"value":"under_4h","label":"Under 4 hours","score":100},{"value":"under_24h","label":"4-24 hours","score":70},{"value":"over_24h","label":"Over 24 hours","score":40},{"value":"undefined","label":"Not defined","score":0}]', true, 1.00, 'high', 'RTO should be proportionate to the criticality of the service.', false, 4, ARRAY['nis2','rto']),
('b0000003-0005-0000-0000-000000000001', 'Do you conduct regular exercises and drills for business continuity scenarios?', 'single_choice', '[{"value":"annually","label":"At least annually","score":100},{"value":"adhoc","label":"Ad-hoc","score":40},{"value":"never","label":"Never","score":0}]', true, 1.00, 'high', 'Regular testing validates the effectiveness of continuity plans.', true, 5, ARRAY['nis2','exercises','testing']),
('b0000003-0005-0000-0000-000000000001', 'Have you identified and documented single points of failure in your infrastructure?', 'yes_no', '[{"value":"yes","label":"Yes","score":100},{"value":"no","label":"No","score":0}]', true, 0.90, 'medium', 'Understanding SPoFs is essential for resilience planning.', false, 6, ARRAY['nis2','spof','resilience']);


-- ============================================================
-- TEMPLATE 4: Quick Security Assessment
-- ============================================================

INSERT INTO assessment_questionnaires (
    id, organization_id, name, description, questionnaire_type, version, status,
    total_questions, total_sections, estimated_completion_minutes,
    scoring_method, pass_threshold, risk_tier_thresholds,
    applicable_vendor_tiers, is_system, is_template
) VALUES (
    'a0000004-0000-0000-0000-000000000001', NULL,
    'Quick Security Assessment',
    'A streamlined 20-question assessment for low-risk vendors or preliminary screening. Covers essential security controls, data protection basics, incident response readiness, and compliance status.',
    'security', 1, 'active',
    20, 4, 20,
    'simple_average', 65.00,
    '{"critical": 30, "high": 45, "medium": 65, "low": 80}',
    ARRAY['low', 'medium'],
    true, true
);

-- Quick Section 1: Security Basics
INSERT INTO questionnaire_sections (id, questionnaire_id, name, description, sort_order, weight)
VALUES ('b0000004-0001-0000-0000-000000000001', 'a0000004-0000-0000-0000-000000000001',
    'Security Basics', 'Fundamental security controls and governance', 1, 1.00);

INSERT INTO questionnaire_questions (section_id, question_text, question_type, options, is_required, weight, risk_impact, guidance_text, evidence_required, sort_order, tags) VALUES
('b0000004-0001-0000-0000-000000000001', 'Do you have a documented Information Security Policy?', 'yes_no', '[{"value":"yes","label":"Yes","score":100},{"value":"no","label":"No","score":0}]', true, 1.00, 'high', 'A security policy is the foundation of any security programme.', false, 1, ARRAY['quick','policy']),
('b0000004-0001-0000-0000-000000000001', 'Is multi-factor authentication (MFA) enforced for remote access?', 'yes_no', '[{"value":"yes","label":"Yes","score":100},{"value":"no","label":"No","score":0}]', true, 1.20, 'critical', 'MFA significantly reduces the risk of credential-based attacks.', false, 2, ARRAY['quick','mfa']),
('b0000004-0001-0000-0000-000000000001', 'Do you conduct security awareness training for employees?', 'single_choice', '[{"value":"regular","label":"Yes, at least annually","score":100},{"value":"onboarding","label":"Onboarding only","score":50},{"value":"no","label":"No","score":0}]', true, 1.00, 'high', 'Training is essential to prevent social engineering and phishing.', false, 3, ARRAY['quick','training']),
('b0000004-0001-0000-0000-000000000001', 'Are user access rights reviewed at least annually?', 'yes_no', '[{"value":"yes","label":"Yes","score":100},{"value":"no","label":"No","score":0}]', true, 1.00, 'high', 'Regular reviews prevent access creep and orphaned accounts.', false, 4, ARRAY['quick','access-review']),
('b0000004-0001-0000-0000-000000000001', 'Do you have an endpoint protection (antivirus/EDR) solution deployed?', 'yes_no', '[{"value":"yes","label":"Yes","score":100},{"value":"no","label":"No","score":0}]', true, 1.00, 'high', 'Endpoint protection is a fundamental security control.', false, 5, ARRAY['quick','endpoint']);

-- Quick Section 2: Data Protection
INSERT INTO questionnaire_sections (id, questionnaire_id, name, description, sort_order, weight)
VALUES ('b0000004-0002-0000-0000-000000000001', 'a0000004-0000-0000-0000-000000000001',
    'Data Protection', 'Data handling, encryption, and backup', 2, 1.00);

INSERT INTO questionnaire_questions (section_id, question_text, question_type, options, is_required, weight, risk_impact, guidance_text, evidence_required, sort_order, tags) VALUES
('b0000004-0002-0000-0000-000000000001', 'Is data encrypted in transit (TLS 1.2+)?', 'yes_no', '[{"value":"yes","label":"Yes","score":100},{"value":"no","label":"No","score":0}]', true, 1.20, 'critical', 'All data exchanged with clients should be encrypted in transit.', false, 1, ARRAY['quick','encryption','transit']),
('b0000004-0002-0000-0000-000000000001', 'Is data encrypted at rest?', 'yes_no', '[{"value":"yes","label":"Yes","score":100},{"value":"no","label":"No","score":0}]', true, 1.20, 'critical', 'Databases and storage containing client data should be encrypted.', false, 2, ARRAY['quick','encryption','rest']),
('b0000004-0002-0000-0000-000000000001', 'Are backups performed regularly and tested for recoverability?', 'single_choice', '[{"value":"regular_tested","label":"Regular backups, tested","score":100},{"value":"regular","label":"Regular but untested","score":50},{"value":"irregular","label":"Irregular backups","score":20},{"value":"none","label":"No backups","score":0}]', true, 1.10, 'high', 'Backups are critical for data recovery after incidents.', false, 3, ARRAY['quick','backup']),
('b0000004-0002-0000-0000-000000000001', 'Do you have a data retention and secure deletion policy?', 'yes_no', '[{"value":"yes","label":"Yes","score":100},{"value":"no","label":"No","score":0}]', true, 0.90, 'medium', 'Data should be securely deleted when no longer needed.', false, 4, ARRAY['quick','retention']),
('b0000004-0002-0000-0000-000000000001', 'In which countries is our data processed and stored?', 'text', '[]', true, 0.50, 'informational', 'Data residency affects regulatory compliance requirements.', false, 5, ARRAY['quick','data-residency']);

-- Quick Section 3: Incident Response
INSERT INTO questionnaire_sections (id, questionnaire_id, name, description, sort_order, weight)
VALUES ('b0000004-0003-0000-0000-000000000001', 'a0000004-0000-0000-0000-000000000001',
    'Incident Response', 'Breach detection and notification capabilities', 3, 1.00);

INSERT INTO questionnaire_questions (section_id, question_text, question_type, options, is_required, weight, risk_impact, guidance_text, evidence_required, sort_order, tags) VALUES
('b0000004-0003-0000-0000-000000000001', 'Do you have a documented incident response plan?', 'yes_no', '[{"value":"yes","label":"Yes","score":100},{"value":"no","label":"No","score":0}]', true, 1.20, 'critical', 'An incident response plan is essential for managing security incidents.', false, 1, ARRAY['quick','incident','plan']),
('b0000004-0003-0000-0000-000000000001', 'Can you notify us of a security breach within 72 hours?', 'yes_no', '[{"value":"yes","label":"Yes","score":100},{"value":"no","label":"No","score":0}]', true, 1.30, 'critical', 'Timely notification enables us to meet our regulatory obligations.', false, 2, ARRAY['quick','breach','notification']),
('b0000004-0003-0000-0000-000000000001', 'Do you have security logging and monitoring in place?', 'single_choice', '[{"value":"siem","label":"Yes, centralised SIEM","score":100},{"value":"basic","label":"Basic logging","score":50},{"value":"none","label":"No logging","score":0}]', true, 1.10, 'high', 'Logging enables incident detection and forensic investigation.', false, 3, ARRAY['quick','logging']),
('b0000004-0003-0000-0000-000000000001', 'Do you conduct vulnerability scanning on your systems?', 'single_choice', '[{"value":"regular","label":"Yes, regular scanning","score":100},{"value":"adhoc","label":"Ad-hoc scanning","score":50},{"value":"never","label":"Never","score":0}]', true, 1.10, 'high', 'Regular scanning identifies vulnerabilities before they are exploited.', false, 4, ARRAY['quick','scanning']),
('b0000004-0003-0000-0000-000000000001', 'What is your patching SLA for critical vulnerabilities?', 'single_choice', '[{"value":"7d","label":"Within 7 days","score":100},{"value":"30d","label":"Within 30 days","score":60},{"value":"none","label":"No defined SLA","score":0}]', true, 1.10, 'high', 'Timely patching is one of the most effective security controls.', false, 5, ARRAY['quick','patching']);

-- Quick Section 4: Compliance & Certifications
INSERT INTO questionnaire_sections (id, questionnaire_id, name, description, sort_order, weight)
VALUES ('b0000004-0004-0000-0000-000000000001', 'a0000004-0000-0000-0000-000000000001',
    'Compliance & Certifications', 'Industry certifications and regulatory compliance', 4, 1.00);

INSERT INTO questionnaire_questions (section_id, question_text, question_type, options, is_required, weight, risk_impact, guidance_text, evidence_required, sort_order, tags) VALUES
('b0000004-0004-0000-0000-000000000001', 'Do you hold ISO 27001 or SOC 2 certification?', 'single_choice', '[{"value":"both","label":"Both ISO 27001 and SOC 2","score":100},{"value":"iso","label":"ISO 27001 only","score":90},{"value":"soc2","label":"SOC 2 only","score":85},{"value":"in_progress","label":"In progress","score":40},{"value":"none","label":"Neither","score":0}]', true, 1.20, 'high', 'Industry certifications provide independent assurance.', true, 1, ARRAY['quick','certifications']),
('b0000004-0004-0000-0000-000000000001', 'Have you had an independent security audit in the past 12 months?', 'yes_no', '[{"value":"yes","label":"Yes","score":100},{"value":"no","label":"No","score":0}]', false, 1.00, 'medium', 'External audits identify gaps that internal reviews may miss.', false, 2, ARRAY['quick','audit']),
('b0000004-0004-0000-0000-000000000001', 'Do you have a privacy policy that complies with applicable data protection laws?', 'yes_no', '[{"value":"yes","label":"Yes","score":100},{"value":"no","label":"No","score":0}]', true, 1.00, 'high', 'Privacy compliance is essential for data processing arrangements.', false, 3, ARRAY['quick','privacy']),
('b0000004-0004-0000-0000-000000000001', 'Do you use sub-processors, and if so, do you maintain a current list?', 'single_choice', '[{"value":"yes_list","label":"Yes, with maintained list","score":100},{"value":"yes_no_list","label":"Yes, but no maintained list","score":30},{"value":"no_subprocessors","label":"No sub-processors used","score":100}]', true, 1.00, 'high', 'Sub-processor transparency is required under GDPR.', false, 4, ARRAY['quick','sub-processors']),
('b0000004-0004-0000-0000-000000000001', 'Is there any pending litigation or regulatory enforcement action related to data protection?', 'single_choice', '[{"value":"none","label":"No pending actions","score":100},{"value":"minor","label":"Minor/non-material actions","score":50},{"value":"major","label":"Major pending actions","score":0}]', true, 0.80, 'high', 'Pending enforcement may indicate systemic compliance issues.', false, 5, ARRAY['quick','litigation']);

COMMIT;
