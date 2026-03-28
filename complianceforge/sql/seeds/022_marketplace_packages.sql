-- 022_marketplace_packages.sql
-- Seed data: Official ComplianceForge marketplace publisher and three starter packs.

BEGIN;

-- ============================================================
-- OFFICIAL PUBLISHER: ComplianceForge
-- ============================================================

INSERT INTO marketplace_publishers (
    id, organization_id, publisher_name, publisher_slug, description,
    website, logo_url, is_verified, verification_date, is_official,
    total_packages, total_downloads, rating_avg, rating_count, contact_email
) VALUES (
    'a0000000-0000-4000-8000-000000000001',
    '00000000-0000-4000-8000-000000000001',  -- system org
    'ComplianceForge',
    'complianceforge',
    'Official compliance control packs, framework templates, and policy bundles curated by the ComplianceForge team. All official packages are free, peer-reviewed, and mapped to internationally recognised standards.',
    'https://complianceforge.io',
    '/assets/publishers/complianceforge-logo.svg',
    true,
    '2025-01-15',
    true,
    3,
    0,
    4.80,
    0,
    'marketplace@complianceforge.io'
);

-- ============================================================
-- PACKAGE 1: UK Financial Services Compliance Pack
-- ============================================================

INSERT INTO marketplace_packages (
    id, publisher_id, package_slug, name, description, long_description,
    package_type, category, applicable_frameworks, applicable_regions,
    applicable_industries, tags, current_version, min_platform_version,
    pricing_model, price_eur, download_count, install_count,
    rating_avg, rating_count, featured, contents_summary,
    status, published_at, license
) VALUES (
    'b0000000-0000-4000-8000-000000000001',
    'a0000000-0000-4000-8000-000000000001',
    'uk-financial-services-compliance',
    'UK Financial Services Compliance Pack',
    'Comprehensive control library for UK-regulated financial services firms covering FCA, PRA, and Bank of England requirements with full ISO 27001 and NIST CSF mappings.',
    E'## UK Financial Services Compliance Pack\n\nThis pack provides 25 production-ready controls specifically designed for UK financial services organisations subject to FCA, PRA, and Bank of England regulations.\n\n### What''s Included\n- 25 fully documented controls with implementation guidance\n- Cross-framework mappings to ISO 27001, NIST CSF 2.0, and PCI DSS 4.0\n- FCA Senior Managers & Certification Regime (SM&CR) alignment\n- Operational resilience controls aligned to PS21/3\n- Third-party risk management controls per SS2/21\n- Pre-built evidence templates and test procedures\n\n### Who This Is For\n- UK banks and building societies\n- Payment service providers\n- Insurance firms regulated by the FCA\n- Wealth management and investment firms\n- FinTech companies seeking FCA authorisation',
    'control_pack',
    'financial_services',
    ARRAY['ISO27001', 'NIST_CSF_2', 'PCI_DSS_4'],
    ARRAY['UK', 'GB'],
    ARRAY['financial_services', 'banking', 'insurance', 'fintech'],
    ARRAY['fca', 'pra', 'smcr', 'operational-resilience', 'uk-finance', 'ps21/3'],
    '1.0.0',
    '1.0.0',
    'free',
    0.00,
    0,
    0,
    0.00,
    0,
    true,
    '{
        "total_controls": 25,
        "control_categories": [
            {"name": "Operational Resilience", "count": 5},
            {"name": "Third-Party Risk Management", "count": 4},
            {"name": "Data Protection & Privacy", "count": 4},
            {"name": "Cyber Security", "count": 5},
            {"name": "SM&CR Governance", "count": 4},
            {"name": "Financial Crime Prevention", "count": 3}
        ],
        "framework_mappings": 3,
        "evidence_templates": 12,
        "test_procedures": 25
    }'::jsonb,
    'published',
    now(),
    'CC-BY-4.0'
);

INSERT INTO marketplace_package_versions (
    id, package_id, version, release_notes, package_data, package_hash, file_size_bytes,
    is_breaking_change, migration_notes, published_at
) VALUES (
    'c0000000-0000-4000-8000-000000000001',
    'b0000000-0000-4000-8000-000000000001',
    '1.0.0',
    'Initial release of the UK Financial Services Compliance Pack with 25 controls, cross-framework mappings, and evidence templates.',
    '{
        "schema_version": "1.0",
        "package_type": "control_pack",
        "controls": [
            {
                "code": "UKFS-OR-001",
                "title": "Important Business Service Identification",
                "description": "Identify and document all important business services as defined under PS21/3, including impact tolerances for each service.",
                "category": "Operational Resilience",
                "control_type": "directive",
                "implementation_type": "administrative",
                "guidance": "Map all customer-facing and internal services. For each important business service, define maximum tolerable disruption periods and test scenarios. Review quarterly and after significant change.",
                "test_procedure": "Verify the IBS register is complete, impact tolerances are documented, and scenario testing has been conducted within the last 12 months.",
                "evidence_requirements": ["IBS register", "Impact tolerance documentation", "Scenario test results"],
                "mappings": {"ISO27001": ["A.17.1.1", "A.17.1.2"], "NIST_CSF_2": ["ID.BE-4", "ID.BE-5"], "PCI_DSS_4": []}
            },
            {
                "code": "UKFS-OR-002",
                "title": "Operational Resilience Self-Assessment",
                "description": "Conduct annual self-assessment of operational resilience capabilities against PS21/3 requirements.",
                "category": "Operational Resilience",
                "control_type": "detective",
                "implementation_type": "administrative",
                "guidance": "Perform a structured self-assessment covering: mapping of important business services, setting impact tolerances, scenario testing, communication strategy, and lessons learned from incidents.",
                "test_procedure": "Review the most recent self-assessment report. Confirm it covers all PS21/3 requirements and has board sign-off.",
                "evidence_requirements": ["Self-assessment report", "Board minutes", "Action plan"],
                "mappings": {"ISO27001": ["A.17.1.3"], "NIST_CSF_2": ["ID.GV-4"], "PCI_DSS_4": []}
            },
            {
                "code": "UKFS-OR-003",
                "title": "Third-Party Dependency Mapping",
                "description": "Map all third-party dependencies that support important business services and assess concentration risk.",
                "category": "Operational Resilience",
                "control_type": "preventive",
                "implementation_type": "administrative",
                "guidance": "Maintain a register of all third-party providers critical to IBS delivery. Assess single points of failure and concentration risk. Develop contingency arrangements for critical dependencies.",
                "test_procedure": "Review third-party register completeness. Verify concentration risk assessment is current. Test contingency arrangements for top-5 critical providers.",
                "evidence_requirements": ["Third-party register", "Concentration risk assessment", "Contingency test results"],
                "mappings": {"ISO27001": ["A.15.1.1", "A.15.1.2"], "NIST_CSF_2": ["ID.SC-2", "ID.SC-4"], "PCI_DSS_4": ["12.8.1"]}
            },
            {
                "code": "UKFS-OR-004",
                "title": "Scenario Testing Programme",
                "description": "Establish and execute a scenario testing programme that covers severe but plausible disruption events.",
                "category": "Operational Resilience",
                "control_type": "detective",
                "implementation_type": "technical",
                "guidance": "Design test scenarios covering cyber attacks, third-party failures, infrastructure outages, and pandemic situations. Execute tests at least annually and after major changes. Document outcomes and remediation actions.",
                "test_procedure": "Review scenario test schedule and results. Confirm coverage of all IBS and critical scenarios. Verify remediation actions are tracked to completion.",
                "evidence_requirements": ["Test schedule", "Scenario descriptions", "Test results", "Remediation tracker"],
                "mappings": {"ISO27001": ["A.17.1.3", "A.17.2.1"], "NIST_CSF_2": ["PR.IP-10", "RC.RP-1"], "PCI_DSS_4": ["11.6.1"]}
            },
            {
                "code": "UKFS-OR-005",
                "title": "Communications Strategy for Disruptions",
                "description": "Maintain a communications strategy for stakeholder notification during operational disruptions.",
                "category": "Operational Resilience",
                "control_type": "corrective",
                "implementation_type": "administrative",
                "guidance": "Define communication protocols for regulators, customers, and internal stakeholders during disruptions. Include escalation thresholds, templates, and channel redundancy. Test communications at least annually.",
                "test_procedure": "Review communications strategy documentation. Verify templates exist for all stakeholder groups. Confirm annual testing has been completed.",
                "evidence_requirements": ["Communications strategy", "Notification templates", "Test records"],
                "mappings": {"ISO27001": ["A.16.1.1"], "NIST_CSF_2": ["RS.CO-1", "RS.CO-2"], "PCI_DSS_4": []}
            },
            {
                "code": "UKFS-TPRM-001",
                "title": "Third-Party Due Diligence Framework",
                "description": "Implement a risk-based due diligence framework for onboarding and ongoing assessment of third-party providers.",
                "category": "Third-Party Risk Management",
                "control_type": "preventive",
                "implementation_type": "administrative",
                "guidance": "Establish a tiered due diligence process based on criticality and risk. Tier 1 (critical) providers require enhanced due diligence including on-site assessments. Review annually or upon material change.",
                "test_procedure": "Sample 10 third-party files. Verify due diligence was conducted proportionate to risk tier. Check renewal dates and outstanding actions.",
                "evidence_requirements": ["Due diligence questionnaires", "Risk tier classification", "Assessment reports"],
                "mappings": {"ISO27001": ["A.15.1.1", "A.15.2.1"], "NIST_CSF_2": ["ID.SC-1", "ID.SC-2"], "PCI_DSS_4": ["12.8.2"]}
            },
            {
                "code": "UKFS-TPRM-002",
                "title": "Outsourcing Notification to Regulators",
                "description": "Notify the FCA/PRA of material outsourcing arrangements and maintain a register per SS2/21.",
                "category": "Third-Party Risk Management",
                "control_type": "directive",
                "implementation_type": "administrative",
                "guidance": "Maintain a complete outsourcing register. Notify regulators of material outsourcing arrangements before entering into agreements. Include sub-outsourcing chains in the register.",
                "test_procedure": "Review outsourcing register for completeness. Verify FCA/PRA notifications were submitted for all material arrangements. Check sub-outsourcing documentation.",
                "evidence_requirements": ["Outsourcing register", "FCA notification copies", "Sub-outsourcing documentation"],
                "mappings": {"ISO27001": ["A.15.1.2"], "NIST_CSF_2": ["ID.SC-3"], "PCI_DSS_4": ["12.8.5"]}
            },
            {
                "code": "UKFS-TPRM-003",
                "title": "Cloud Outsourcing Risk Assessment",
                "description": "Conduct specific risk assessments for cloud service providers including data residency, exit planning, and audit rights.",
                "category": "Third-Party Risk Management",
                "control_type": "preventive",
                "implementation_type": "technical",
                "guidance": "Assess cloud providers against FCA/PRA expectations. Ensure contractual audit rights, data residency controls, exit strategy feasibility, and business continuity provisions. Review at least annually.",
                "test_procedure": "Review cloud provider risk assessments. Verify contractual terms include required provisions. Test exit plan feasibility.",
                "evidence_requirements": ["Cloud risk assessment", "Contractual review", "Exit plan documentation"],
                "mappings": {"ISO27001": ["A.15.1.2", "A.15.2.2"], "NIST_CSF_2": ["ID.SC-2"], "PCI_DSS_4": ["12.8.2"]}
            },
            {
                "code": "UKFS-TPRM-004",
                "title": "Intragroup Outsourcing Governance",
                "description": "Apply proportionate governance to intragroup outsourcing arrangements.",
                "category": "Third-Party Risk Management",
                "control_type": "directive",
                "implementation_type": "administrative",
                "guidance": "Intragroup arrangements must have formal SLAs, performance monitoring, and documented escalation paths. Treat with same rigour as external outsourcing where material.",
                "test_procedure": "Review intragroup SLAs. Verify performance reporting is in place. Check escalation procedures are documented and tested.",
                "evidence_requirements": ["Intragroup SLAs", "Performance reports", "Escalation records"],
                "mappings": {"ISO27001": ["A.15.1.1"], "NIST_CSF_2": ["ID.SC-1"], "PCI_DSS_4": []}
            },
            {
                "code": "UKFS-DP-001",
                "title": "Customer Data Classification",
                "description": "Classify all customer data according to sensitivity and regulatory requirements, including UK GDPR special categories.",
                "category": "Data Protection & Privacy",
                "control_type": "preventive",
                "implementation_type": "administrative",
                "guidance": "Implement a data classification scheme with at least four tiers: Public, Internal, Confidential, Restricted. Map all customer data repositories. Identify special category data under UK GDPR Article 9.",
                "test_procedure": "Review data classification policy. Sample customer data repositories and verify classification labels are applied. Check special category data has enhanced protections.",
                "evidence_requirements": ["Data classification policy", "Data inventory", "Classification labels evidence"],
                "mappings": {"ISO27001": ["A.8.2.1", "A.8.2.2"], "NIST_CSF_2": ["ID.AM-5", "PR.DS-1"], "PCI_DSS_4": ["3.1.1"]}
            },
            {
                "code": "UKFS-DP-002",
                "title": "Data Subject Rights Fulfilment",
                "description": "Implement processes to fulfil data subject rights under UK GDPR within statutory timeframes.",
                "category": "Data Protection & Privacy",
                "control_type": "corrective",
                "implementation_type": "administrative",
                "guidance": "Establish procedures for handling SARs, erasure requests, data portability, and objections. Track all requests with timestamps. Respond within 30 days (or extended 60 days with notification). Maintain an audit trail.",
                "test_procedure": "Review DSR log for the past quarter. Verify all requests were responded to within statutory deadlines. Check audit trail completeness.",
                "evidence_requirements": ["DSR procedure", "Request log", "Response audit trail"],
                "mappings": {"ISO27001": ["A.18.1.4"], "NIST_CSF_2": ["PR.IP-6"], "PCI_DSS_4": []}
            },
            {
                "code": "UKFS-DP-003",
                "title": "Cross-Border Data Transfer Controls",
                "description": "Implement appropriate safeguards for international data transfers post-Brexit, including adequacy decisions and SCCs.",
                "category": "Data Protection & Privacy",
                "control_type": "preventive",
                "implementation_type": "administrative",
                "guidance": "Map all cross-border data flows. Verify adequacy status of destination countries under UK GDPR. Where no adequacy decision exists, implement International Data Transfer Agreements (IDTAs) or approved SCCs with UK addendum. Conduct Transfer Impact Assessments.",
                "test_procedure": "Review data transfer register. Verify appropriate mechanisms (adequacy, IDTA, SCCs) are in place for each flow. Check TIA documentation.",
                "evidence_requirements": ["Data transfer register", "IDTAs/SCCs", "Transfer Impact Assessments"],
                "mappings": {"ISO27001": ["A.18.1.4", "A.18.1.1"], "NIST_CSF_2": ["PR.DS-5"], "PCI_DSS_4": []}
            },
            {
                "code": "UKFS-DP-004",
                "title": "Data Breach Notification Process",
                "description": "Maintain a data breach notification process aligned to both UK GDPR (ICO) and FCA reporting requirements.",
                "category": "Data Protection & Privacy",
                "control_type": "corrective",
                "implementation_type": "administrative",
                "guidance": "Define breach classification criteria. ICO notification required within 72 hours for reportable breaches. FCA notification required for material cyber incidents. Maintain breach register with root cause analysis.",
                "test_procedure": "Review breach notification procedure. Verify escalation criteria cover both ICO and FCA requirements. Conduct tabletop exercise. Review breach register.",
                "evidence_requirements": ["Breach notification procedure", "Breach register", "Tabletop exercise results"],
                "mappings": {"ISO27001": ["A.16.1.2", "A.16.1.5"], "NIST_CSF_2": ["RS.CO-2", "RS.CO-3"], "PCI_DSS_4": ["12.10.1"]}
            },
            {
                "code": "UKFS-CS-001",
                "title": "Cyber Threat Intelligence Programme",
                "description": "Establish a cyber threat intelligence capability relevant to the financial services threat landscape.",
                "category": "Cyber Security",
                "control_type": "detective",
                "implementation_type": "technical",
                "guidance": "Subscribe to FS-ISAC and NCSC threat feeds. Conduct quarterly threat landscape assessments. Feed threat intelligence into vulnerability management and security operations. Share intelligence with sector peers where appropriate.",
                "test_procedure": "Verify threat intelligence sources are active. Review quarterly threat assessments. Check integration with SOC processes.",
                "evidence_requirements": ["Threat intelligence subscriptions", "Quarterly assessments", "SOC integration evidence"],
                "mappings": {"ISO27001": ["A.6.1.4"], "NIST_CSF_2": ["ID.RA-2", "DE.AE-2"], "PCI_DSS_4": ["6.3.1"]}
            },
            {
                "code": "UKFS-CS-002",
                "title": "CBEST/STAR-FS Threat-Led Penetration Testing",
                "description": "Participate in threat-led penetration testing programmes as required by PRA/FCA supervision.",
                "category": "Cyber Security",
                "control_type": "detective",
                "implementation_type": "technical",
                "guidance": "Engage CREST-accredited providers for CBEST or STAR-FS exercises. Scope critical functions and important business services. Share results with regulators as required. Track remediation actions.",
                "test_procedure": "Review most recent CBEST/STAR-FS report. Verify remediation actions are tracked. Confirm regulatory submissions were made.",
                "evidence_requirements": ["CBEST/STAR-FS report", "Remediation tracker", "Regulatory correspondence"],
                "mappings": {"ISO27001": ["A.12.6.1", "A.18.2.3"], "NIST_CSF_2": ["DE.CM-8", "PR.IP-12"], "PCI_DSS_4": ["11.4.1"]}
            },
            {
                "code": "UKFS-CS-003",
                "title": "Privileged Access Management",
                "description": "Implement controls for managing, monitoring, and auditing privileged access across all critical systems.",
                "category": "Cyber Security",
                "control_type": "preventive",
                "implementation_type": "technical",
                "guidance": "Deploy PAM solution for all critical infrastructure. Enforce just-in-time access, session recording, and multi-factor authentication. Review privileged accounts quarterly. Remove dormant accounts after 30 days.",
                "test_procedure": "Review PAM solution coverage. Verify JIT access is enforced. Sample session recordings. Check quarterly review evidence.",
                "evidence_requirements": ["PAM deployment report", "Access review records", "Session recording samples"],
                "mappings": {"ISO27001": ["A.9.2.3", "A.9.4.1"], "NIST_CSF_2": ["PR.AC-4", "PR.AC-6"], "PCI_DSS_4": ["7.2.1", "8.6.1"]}
            },
            {
                "code": "UKFS-CS-004",
                "title": "Security Operations Centre (SOC)",
                "description": "Operate or contract a 24/7 security operations centre with financial services-specific detection rules.",
                "category": "Cyber Security",
                "control_type": "detective",
                "implementation_type": "technical",
                "guidance": "SOC must provide 24/7/365 monitoring with financial services-specific use cases (e.g., SWIFT fraud detection, payment anomalies). Integrate with SIEM, EDR, and network monitoring. Conduct purple team exercises quarterly.",
                "test_procedure": "Review SOC operational metrics (MTTD, MTTR). Verify FS-specific detection rules are deployed. Review most recent purple team exercise results.",
                "evidence_requirements": ["SOC operational dashboard", "Detection rule inventory", "Purple team report"],
                "mappings": {"ISO27001": ["A.12.4.1", "A.16.1.2"], "NIST_CSF_2": ["DE.CM-1", "DE.CM-7"], "PCI_DSS_4": ["10.4.1", "10.7.1"]}
            },
            {
                "code": "UKFS-CS-005",
                "title": "API Security for Open Banking",
                "description": "Implement security controls for Open Banking APIs and PSD2-compliant payment interfaces.",
                "category": "Cyber Security",
                "control_type": "preventive",
                "implementation_type": "technical",
                "guidance": "Secure all Open Banking APIs per OBIE standards. Implement strong customer authentication (SCA). Deploy API gateway with rate limiting, input validation, and mutual TLS. Monitor for API abuse patterns.",
                "test_procedure": "Review API security configuration. Verify SCA implementation. Test rate limiting and input validation. Review API monitoring dashboards.",
                "evidence_requirements": ["API security config", "SCA compliance evidence", "API monitoring logs"],
                "mappings": {"ISO27001": ["A.14.1.2", "A.14.1.3"], "NIST_CSF_2": ["PR.DS-2", "PR.AC-7"], "PCI_DSS_4": ["6.2.1"]}
            },
            {
                "code": "UKFS-GOV-001",
                "title": "SM&CR Responsibilities Mapping",
                "description": "Map all Senior Manager Functions (SMFs) to compliance and operational resilience responsibilities.",
                "category": "SM&CR Governance",
                "control_type": "directive",
                "implementation_type": "administrative",
                "guidance": "Document all prescribed responsibilities under SM&CR. Map each SMF to specific compliance areas including cyber security, operational resilience, and data protection. Maintain Statements of Responsibilities (SoR) and Management Responsibilities Maps (MRM).",
                "test_procedure": "Review SoRs for all senior managers. Verify MRM is current. Check that compliance responsibilities are explicitly assigned.",
                "evidence_requirements": ["Statements of Responsibilities", "Management Responsibilities Map", "FCA Directory entries"],
                "mappings": {"ISO27001": ["A.6.1.1"], "NIST_CSF_2": ["ID.GV-1", "ID.GV-2"], "PCI_DSS_4": []}
            },
            {
                "code": "UKFS-GOV-002",
                "title": "Board Cyber Risk Reporting",
                "description": "Provide regular cyber risk and operational resilience reporting to the Board and relevant sub-committees.",
                "category": "SM&CR Governance",
                "control_type": "directive",
                "implementation_type": "administrative",
                "guidance": "Produce quarterly board-level cyber risk reports covering: threat landscape, vulnerability posture, incident metrics, control effectiveness, and regulatory developments. Include RAG-rated risk appetite tracking.",
                "test_procedure": "Review last four quarterly board reports. Verify coverage of required topics. Check board minutes for discussion evidence.",
                "evidence_requirements": ["Board reports", "Board/committee minutes", "Risk appetite dashboards"],
                "mappings": {"ISO27001": ["A.6.1.1", "A.18.2.1"], "NIST_CSF_2": ["ID.GV-4", "ID.RM-1"], "PCI_DSS_4": []}
            },
            {
                "code": "UKFS-GOV-003",
                "title": "Conduct Risk Assessment",
                "description": "Integrate conduct risk considerations into operational processes and control frameworks.",
                "category": "SM&CR Governance",
                "control_type": "preventive",
                "implementation_type": "administrative",
                "guidance": "Identify conduct risk indicators across customer-facing processes. Implement monitoring controls for potential customer detriment. Report conduct risk metrics to the Board. Align with FCA Consumer Duty requirements.",
                "test_procedure": "Review conduct risk framework. Verify monitoring controls are operational. Check alignment with Consumer Duty requirements.",
                "evidence_requirements": ["Conduct risk framework", "Monitoring reports", "Consumer Duty assessment"],
                "mappings": {"ISO27001": ["A.18.1.1"], "NIST_CSF_2": ["ID.GV-3"], "PCI_DSS_4": []}
            },
            {
                "code": "UKFS-GOV-004",
                "title": "Compliance Monitoring Programme",
                "description": "Operate a risk-based compliance monitoring programme covering all applicable regulations.",
                "category": "SM&CR Governance",
                "control_type": "detective",
                "implementation_type": "administrative",
                "guidance": "Design an annual compliance monitoring plan based on regulatory risk assessment. Cover FCA/PRA rules, UK GDPR, AML/CTF, and sector-specific regulations. Report findings to compliance committee quarterly.",
                "test_procedure": "Review annual monitoring plan. Verify execution against plan. Check findings are reported and remediated.",
                "evidence_requirements": ["Annual monitoring plan", "Monitoring reports", "Findings tracker"],
                "mappings": {"ISO27001": ["A.18.2.2", "A.18.2.3"], "NIST_CSF_2": ["DE.CM-1"], "PCI_DSS_4": ["12.4.1"]}
            },
            {
                "code": "UKFS-FC-001",
                "title": "Transaction Monitoring Controls",
                "description": "Implement automated transaction monitoring for anti-money laundering, fraud detection, and sanctions screening.",
                "category": "Financial Crime Prevention",
                "control_type": "detective",
                "implementation_type": "technical",
                "guidance": "Deploy automated transaction monitoring with risk-based rules and scenarios. Cover AML typologies relevant to your business. Screen against OFSI, EU, and UN sanctions lists in real-time. Conduct annual calibration of monitoring rules.",
                "test_procedure": "Review monitoring rule coverage. Verify sanctions screening is real-time. Check calibration records. Sample alert investigation quality.",
                "evidence_requirements": ["Rule inventory", "Sanctions screening config", "Calibration records", "Alert samples"],
                "mappings": {"ISO27001": ["A.18.1.1"], "NIST_CSF_2": ["DE.CM-1", "DE.CM-7"], "PCI_DSS_4": []}
            },
            {
                "code": "UKFS-FC-002",
                "title": "Customer Due Diligence (CDD/EDD)",
                "description": "Implement risk-based customer due diligence procedures including enhanced due diligence for high-risk customers.",
                "category": "Financial Crime Prevention",
                "control_type": "preventive",
                "implementation_type": "administrative",
                "guidance": "Apply CDD at onboarding and ongoing. Implement EDD for PEPs, high-risk jurisdictions, and complex structures. Use electronic identity verification. Conduct periodic reviews proportionate to risk classification.",
                "test_procedure": "Sample CDD files across risk tiers. Verify EDD is applied where required. Check periodic review schedule adherence.",
                "evidence_requirements": ["CDD records", "EDD case files", "Periodic review schedule"],
                "mappings": {"ISO27001": ["A.18.1.1"], "NIST_CSF_2": ["ID.GV-3"], "PCI_DSS_4": []}
            },
            {
                "code": "UKFS-FC-003",
                "title": "Suspicious Activity Reporting (SAR)",
                "description": "Maintain an effective SAR regime with clear escalation from front-line staff to the MLRO.",
                "category": "Financial Crime Prevention",
                "control_type": "corrective",
                "implementation_type": "administrative",
                "guidance": "Train all staff on recognising and reporting suspicious activity. Implement a secure internal reporting mechanism. MLRO must assess and submit to NCA via SAR Online within required timeframes. Maintain confidentiality of tipping-off obligations.",
                "test_procedure": "Review SAR policy and training records. Verify internal reporting mechanism. Sample MLRO assessments. Check NCA submission records.",
                "evidence_requirements": ["SAR policy", "Training records", "MLRO assessment log", "NCA submission records"],
                "mappings": {"ISO27001": ["A.16.1.2"], "NIST_CSF_2": ["RS.CO-2"], "PCI_DSS_4": []}
            }
        ]
    }'::jsonb,
    'b6d7e8a9c0d1e2f3a4b5c6d7e8a9c0d1e2f3a4b5c6d7e8a9c0d1e2f3a4b5c6d7e8a9c0d1e2f3a4b5c6d7e8a9c0d1e2f3a4b5c6d7e8a9c0d1e2f3a4b5',
    245760,
    false,
    NULL,
    now()
);

-- ============================================================
-- PACKAGE 2: Healthcare GDPR Data Protection Pack
-- ============================================================

INSERT INTO marketplace_packages (
    id, publisher_id, package_slug, name, description, long_description,
    package_type, category, applicable_frameworks, applicable_regions,
    applicable_industries, tags, current_version, min_platform_version,
    pricing_model, price_eur, download_count, install_count,
    rating_avg, rating_count, featured, contents_summary,
    status, published_at, license
) VALUES (
    'b0000000-0000-4000-8000-000000000002',
    'a0000000-0000-4000-8000-000000000001',
    'healthcare-gdpr-data-protection',
    'Healthcare GDPR Data Protection Pack',
    'Tailored data protection controls for healthcare organisations processing patient data under EU/UK GDPR with specific guidance for Article 9 special category health data.',
    E'## Healthcare GDPR Data Protection Pack\n\nA comprehensive set of 20 data protection controls designed specifically for healthcare organisations handling sensitive patient data.\n\n### What''s Included\n- 20 healthcare-specific data protection controls\n- Article 9 special category data handling guidance\n- Patient consent management controls\n- Cross-border health data transfer procedures\n- Data Protection Impact Assessment (DPIA) templates\n- Mappings to ISO 27001, UK GDPR, and NHS DSPT\n\n### Who This Is For\n- NHS Trusts and Foundation Trusts\n- Private hospitals and clinics\n- Healthcare SaaS providers\n- Clinical research organisations\n- Health data processors',
    'control_pack',
    'healthcare',
    ARRAY['ISO27001', 'UK_GDPR'],
    ARRAY['EU', 'UK', 'EEA'],
    ARRAY['healthcare', 'nhs', 'clinical_research', 'health_tech'],
    ARRAY['gdpr', 'article-9', 'health-data', 'patient-privacy', 'dpia', 'nhs-dspt'],
    '1.0.0',
    '1.0.0',
    'free',
    0.00,
    0,
    0,
    0.00,
    0,
    true,
    '{
        "total_controls": 20,
        "control_categories": [
            {"name": "Special Category Data Handling", "count": 4},
            {"name": "Patient Consent Management", "count": 3},
            {"name": "Health Data Security", "count": 5},
            {"name": "Cross-Border Transfers", "count": 3},
            {"name": "Data Subject Rights (Healthcare)", "count": 3},
            {"name": "DPIA & Risk Assessment", "count": 2}
        ],
        "framework_mappings": 2,
        "evidence_templates": 10,
        "test_procedures": 20
    }'::jsonb,
    'published',
    now(),
    'CC-BY-4.0'
);

INSERT INTO marketplace_package_versions (
    id, package_id, version, release_notes, package_data, package_hash, file_size_bytes,
    is_breaking_change, migration_notes, published_at
) VALUES (
    'c0000000-0000-4000-8000-000000000002',
    'b0000000-0000-4000-8000-000000000002',
    '1.0.0',
    'Initial release of the Healthcare GDPR Data Protection Pack with 20 controls and DPIA templates.',
    '{
        "schema_version": "1.0",
        "package_type": "control_pack",
        "controls": [
            {
                "code": "HC-SCD-001",
                "title": "Article 9 Legal Basis Documentation",
                "description": "Document and maintain a lawful basis register for all processing of Article 9 special category health data.",
                "category": "Special Category Data Handling",
                "control_type": "directive",
                "implementation_type": "administrative",
                "guidance": "For each processing activity involving health data, identify and document the specific Article 9(2) condition relied upon (e.g., explicit consent, healthcare provision, public health). Maintain a register linked to your ROPA.",
                "test_procedure": "Review the Article 9 legal basis register. Cross-reference with ROPA. Verify each processing activity has a documented lawful basis.",
                "evidence_requirements": ["Legal basis register", "ROPA entries", "Legal advice records"],
                "mappings": {"ISO27001": ["A.18.1.4"], "UK_GDPR": ["Art.6", "Art.9"]}
            },
            {
                "code": "HC-SCD-002",
                "title": "Health Data Pseudonymisation",
                "description": "Apply pseudonymisation techniques to health data used for secondary purposes such as research and analytics.",
                "category": "Special Category Data Handling",
                "control_type": "preventive",
                "implementation_type": "technical",
                "guidance": "Implement tokenisation or key-coding for patient identifiers. Store linkage keys separately with strict access controls. Use de-identification risk assessment frameworks (e.g., k-anonymity, l-diversity) for statistical outputs.",
                "test_procedure": "Review pseudonymisation procedures. Verify key management controls. Assess re-identification risk for sample datasets.",
                "evidence_requirements": ["Pseudonymisation procedure", "Key management evidence", "Re-identification risk assessment"],
                "mappings": {"ISO27001": ["A.10.1.1", "A.8.2.3"], "UK_GDPR": ["Art.25", "Art.89"]}
            },
            {
                "code": "HC-SCD-003",
                "title": "Genetic Data Protection Controls",
                "description": "Implement enhanced protections for genetic and genomic data processing.",
                "category": "Special Category Data Handling",
                "control_type": "preventive",
                "implementation_type": "technical",
                "guidance": "Apply encryption at rest and in transit for genetic data. Restrict access to minimum necessary personnel. Implement data segregation from other health records. Consider long-term implications of genetic data retention.",
                "test_procedure": "Review encryption configuration for genetic data stores. Verify access controls. Check data segregation architecture.",
                "evidence_requirements": ["Encryption evidence", "Access control lists", "Architecture documentation"],
                "mappings": {"ISO27001": ["A.10.1.1", "A.9.1.1"], "UK_GDPR": ["Art.9"]}
            },
            {
                "code": "HC-SCD-004",
                "title": "Mental Health Data Sensitivity Controls",
                "description": "Apply enhanced confidentiality controls for mental health records beyond standard health data protections.",
                "category": "Special Category Data Handling",
                "control_type": "preventive",
                "implementation_type": "administrative",
                "guidance": "Implement break-glass access procedures for mental health records. Restrict default access to treating clinicians only. Apply additional audit logging for all access events. Consider patient-level access restrictions.",
                "test_procedure": "Review access policies for mental health records. Verify break-glass procedures. Audit access logs for appropriateness.",
                "evidence_requirements": ["Access policy", "Break-glass procedures", "Access audit logs"],
                "mappings": {"ISO27001": ["A.9.1.1", "A.12.4.1"], "UK_GDPR": ["Art.9"]}
            },
            {
                "code": "HC-PCM-001",
                "title": "Granular Patient Consent Framework",
                "description": "Implement a granular consent management system allowing patients to control specific uses of their health data.",
                "category": "Patient Consent Management",
                "control_type": "preventive",
                "implementation_type": "technical",
                "guidance": "Provide patients with granular consent options covering: direct care, secondary uses, research participation, data sharing with third parties. Record consent decisions with timestamps and version tracking. Support consent withdrawal with downstream propagation.",
                "test_procedure": "Review consent system functionality. Verify granularity of consent options. Test withdrawal propagation. Check audit trail.",
                "evidence_requirements": ["Consent system documentation", "Consent form templates", "Withdrawal test results"],
                "mappings": {"ISO27001": ["A.18.1.4"], "UK_GDPR": ["Art.7", "Art.9"]}
            },
            {
                "code": "HC-PCM-002",
                "title": "National Data Opt-Out Compliance",
                "description": "Implement the NHS National Data Opt-Out in accordance with NHS Digital requirements.",
                "category": "Patient Consent Management",
                "control_type": "directive",
                "implementation_type": "technical",
                "guidance": "Integrate with the National Data Opt-Out service. Apply opt-out preferences before disclosing confidential patient information for purposes beyond direct care. Conduct regular compliance checks against the opt-out register.",
                "test_procedure": "Verify integration with National Data Opt-Out service. Test opt-out application logic. Review compliance check records.",
                "evidence_requirements": ["Integration documentation", "Opt-out test results", "Compliance check records"],
                "mappings": {"ISO27001": ["A.18.1.4"], "UK_GDPR": ["Art.21"]}
            },
            {
                "code": "HC-PCM-003",
                "title": "Clinical Trial Consent Management",
                "description": "Manage consent specific to clinical trial data processing, including re-consent for secondary use of trial data.",
                "category": "Patient Consent Management",
                "control_type": "directive",
                "implementation_type": "administrative",
                "guidance": "Maintain separate consent processes for clinical trial participation and data processing. Implement re-consent workflows for secondary use. Track consent against specific protocol versions. Support broad consent models with ethics committee approval.",
                "test_procedure": "Review trial consent procedures. Verify consent-to-protocol version tracking. Check re-consent workflows.",
                "evidence_requirements": ["Trial consent forms", "Ethics committee approvals", "Consent tracking records"],
                "mappings": {"ISO27001": ["A.18.1.4"], "UK_GDPR": ["Art.7", "Art.89"]}
            },
            {
                "code": "HC-HDS-001",
                "title": "Electronic Health Record Encryption",
                "description": "Implement encryption for electronic health records at rest, in transit, and during processing where feasible.",
                "category": "Health Data Security",
                "control_type": "preventive",
                "implementation_type": "technical",
                "guidance": "Apply AES-256 encryption at rest for all EHR databases. Enforce TLS 1.3 for data in transit. Consider application-level encryption for highly sensitive fields. Implement envelope encryption with HSM-backed key management.",
                "test_procedure": "Verify encryption configuration for EHR databases. Test TLS configuration. Review key management procedures.",
                "evidence_requirements": ["Encryption configuration", "TLS scan results", "Key management documentation"],
                "mappings": {"ISO27001": ["A.10.1.1", "A.10.1.2"], "UK_GDPR": ["Art.32"]}
            },
            {
                "code": "HC-HDS-002",
                "title": "Clinical System Access Controls",
                "description": "Implement role-based access controls for clinical information systems with minimum necessary access principles.",
                "category": "Health Data Security",
                "control_type": "preventive",
                "implementation_type": "technical",
                "guidance": "Define clinical roles with minimum necessary access to patient records. Implement Legitimate Relationship checks where available. Enforce context-based access (ward, department, treating team). Log all access events.",
                "test_procedure": "Review role definitions and access matrices. Verify Legitimate Relationship checks. Sample access logs for appropriateness.",
                "evidence_requirements": ["Role definitions", "Access matrices", "Access audit samples"],
                "mappings": {"ISO27001": ["A.9.1.1", "A.9.2.1"], "UK_GDPR": ["Art.25", "Art.32"]}
            },
            {
                "code": "HC-HDS-003",
                "title": "Medical Device Data Security",
                "description": "Secure data flows from connected medical devices, including IoT health monitors and diagnostic equipment.",
                "category": "Health Data Security",
                "control_type": "preventive",
                "implementation_type": "technical",
                "guidance": "Segment medical devices on dedicated network VLANs. Implement device authentication and encrypted communications. Monitor device firmware for vulnerabilities. Maintain a medical device asset inventory with data flow documentation.",
                "test_procedure": "Review network segmentation for medical devices. Verify device authentication. Check firmware patching status. Review device inventory.",
                "evidence_requirements": ["Network architecture diagrams", "Device inventory", "Firmware status report"],
                "mappings": {"ISO27001": ["A.13.1.3", "A.12.6.1"], "UK_GDPR": ["Art.32"]}
            },
            {
                "code": "HC-HDS-004",
                "title": "Health Data Backup and Recovery",
                "description": "Implement backup and disaster recovery procedures specific to health data availability requirements.",
                "category": "Health Data Security",
                "control_type": "corrective",
                "implementation_type": "technical",
                "guidance": "Maintain RPO under 1 hour for critical clinical systems. Test restoration procedures quarterly. Ensure backups are encrypted and stored in geographically separate locations. Validate data integrity post-restoration.",
                "test_procedure": "Review backup configuration and RPO/RTO targets. Verify quarterly restoration tests. Check backup encryption and geographic separation.",
                "evidence_requirements": ["Backup policy", "Restoration test records", "RPO/RTO evidence"],
                "mappings": {"ISO27001": ["A.12.3.1", "A.17.1.1"], "UK_GDPR": ["Art.32"]}
            },
            {
                "code": "HC-HDS-005",
                "title": "Health Information Exchange Security",
                "description": "Secure health data exchanges with external organisations including GP surgeries, pharmacies, and social care.",
                "category": "Health Data Security",
                "control_type": "preventive",
                "implementation_type": "technical",
                "guidance": "Use approved messaging standards (FHIR, HL7v2) with TLS. Validate organisation identity via NHS MESH or equivalent. Implement data quality checks on inbound messages. Log all exchange events.",
                "test_procedure": "Review exchange interface security. Verify messaging standard compliance. Check organisation validation. Review exchange logs.",
                "evidence_requirements": ["Interface documentation", "Security configuration", "Exchange audit logs"],
                "mappings": {"ISO27001": ["A.13.2.1", "A.13.2.2"], "UK_GDPR": ["Art.28", "Art.32"]}
            },
            {
                "code": "HC-CBT-001",
                "title": "International Health Data Transfer Assessment",
                "description": "Conduct transfer impact assessments for health data sent outside the UK/EEA.",
                "category": "Cross-Border Transfers",
                "control_type": "preventive",
                "implementation_type": "administrative",
                "guidance": "Identify all international health data transfers. Assess destination country legal framework for health data protection. Consider additional technical measures (encryption, pseudonymisation) to supplement SCCs. Document assessment and obtain DPO sign-off.",
                "test_procedure": "Review transfer impact assessments. Verify all international transfers are covered. Check DPO sign-off records.",
                "evidence_requirements": ["Transfer impact assessments", "DPO approval records", "Additional safeguard documentation"],
                "mappings": {"ISO27001": ["A.18.1.4"], "UK_GDPR": ["Art.44", "Art.46"]}
            },
            {
                "code": "HC-CBT-002",
                "title": "Cloud Hosting Data Residency",
                "description": "Ensure health data hosted in cloud environments complies with data residency requirements.",
                "category": "Cross-Border Transfers",
                "control_type": "preventive",
                "implementation_type": "technical",
                "guidance": "Configure cloud resources to restrict data storage to approved regions (UK/EEA). Monitor for configuration drift. Implement controls to prevent accidental data transfer via backup or replication. Obtain contractual data residency guarantees.",
                "test_procedure": "Review cloud resource region configuration. Verify contractual data residency terms. Test for configuration drift.",
                "evidence_requirements": ["Cloud configuration", "Contractual terms", "Drift monitoring reports"],
                "mappings": {"ISO27001": ["A.18.1.4", "A.15.1.2"], "UK_GDPR": ["Art.44"]}
            },
            {
                "code": "HC-CBT-003",
                "title": "Multi-National Clinical Research Data Sharing",
                "description": "Implement governance for sharing patient data across borders for clinical research purposes.",
                "category": "Cross-Border Transfers",
                "control_type": "directive",
                "implementation_type": "administrative",
                "guidance": "Establish data sharing agreements with all research partners. Apply de-identification before transfer where possible. Use approved transfer mechanisms. Obtain ethics committee and Caldicott Guardian approval.",
                "test_procedure": "Review data sharing agreements. Verify de-identification procedures. Check ethics approvals and Caldicott Guardian sign-off.",
                "evidence_requirements": ["Data sharing agreements", "De-identification evidence", "Ethics approvals"],
                "mappings": {"ISO27001": ["A.18.1.4"], "UK_GDPR": ["Art.46", "Art.89"]}
            },
            {
                "code": "HC-DSR-001",
                "title": "Patient Medical Records Access",
                "description": "Facilitate patient access to their complete medical records in portable, understandable formats.",
                "category": "Data Subject Rights (Healthcare)",
                "control_type": "corrective",
                "implementation_type": "technical",
                "guidance": "Provide patients access to records via patient portal. Support machine-readable export formats (FHIR). Explain clinical terminology in lay terms where possible. Process requests within 30 days.",
                "test_procedure": "Test patient portal access. Verify export format options. Review request processing times.",
                "evidence_requirements": ["Portal documentation", "Export format samples", "Request processing metrics"],
                "mappings": {"ISO27001": ["A.18.1.4"], "UK_GDPR": ["Art.15", "Art.20"]}
            },
            {
                "code": "HC-DSR-002",
                "title": "Health Data Erasure Assessment",
                "description": "Assess erasure requests against clinical record retention obligations before actioning.",
                "category": "Data Subject Rights (Healthcare)",
                "control_type": "corrective",
                "implementation_type": "administrative",
                "guidance": "Evaluate each erasure request against NHS Records Management Code retention schedules. Inform the data subject of any applicable exemptions. Where erasure is permissible, cascade to all systems including backups. Document all decisions.",
                "test_procedure": "Review erasure request handling process. Verify retention schedule checks. Sample decision records.",
                "evidence_requirements": ["Erasure procedure", "Retention schedule references", "Decision records"],
                "mappings": {"ISO27001": ["A.18.1.3"], "UK_GDPR": ["Art.17"]}
            },
            {
                "code": "HC-DSR-003",
                "title": "Third-Party Data Subject Requests",
                "description": "Handle data subject requests made on behalf of patients by authorised representatives.",
                "category": "Data Subject Rights (Healthcare)",
                "control_type": "corrective",
                "implementation_type": "administrative",
                "guidance": "Verify representative authority (Power of Attorney, parental responsibility, written authorisation). Apply Caldicott Principles when assessing disclosure. Consider patient capacity (Mental Capacity Act). Document verification steps.",
                "test_procedure": "Review representative verification procedures. Sample third-party request files. Verify Caldicott Principle application.",
                "evidence_requirements": ["Verification procedure", "Representative authority evidence", "Decision records"],
                "mappings": {"ISO27001": ["A.18.1.4"], "UK_GDPR": ["Art.12"]}
            },
            {
                "code": "HC-DPIA-001",
                "title": "Mandatory DPIA for Health Data Processing",
                "description": "Conduct Data Protection Impact Assessments for all new or significantly changed health data processing activities.",
                "category": "DPIA & Risk Assessment",
                "control_type": "preventive",
                "implementation_type": "administrative",
                "guidance": "Trigger a DPIA for any processing involving health data at scale, new technology, or systematic monitoring. Use the ICO DPIA template with healthcare-specific risk factors. Obtain DPO and Caldicott Guardian review before proceeding.",
                "test_procedure": "Review DPIA register. Verify DPIAs were conducted for all qualifying activities. Check DPO and Caldicott sign-off.",
                "evidence_requirements": ["DPIA register", "Completed DPIAs", "DPO/Caldicott sign-off"],
                "mappings": {"ISO27001": ["A.18.1.4"], "UK_GDPR": ["Art.35"]}
            },
            {
                "code": "HC-DPIA-002",
                "title": "AI/ML Health Data Processing Assessment",
                "description": "Conduct enhanced impact assessments for artificial intelligence and machine learning processing of health data.",
                "category": "DPIA & Risk Assessment",
                "control_type": "preventive",
                "implementation_type": "administrative",
                "guidance": "Supplement standard DPIA with AI-specific assessments covering: algorithmic bias in health data, explainability of clinical decisions, training data representativeness, and continuous monitoring of model performance. Engage clinical experts in review.",
                "test_procedure": "Review AI-specific DPIA components. Verify bias assessments. Check clinical expert involvement. Review model monitoring procedures.",
                "evidence_requirements": ["AI DPIA", "Bias assessment", "Clinical review records", "Model monitoring evidence"],
                "mappings": {"ISO27001": ["A.18.1.4"], "UK_GDPR": ["Art.22", "Art.35"]}
            }
        ]
    }'::jsonb,
    'a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0',
    198400,
    false,
    NULL,
    now()
);

-- ============================================================
-- PACKAGE 3: Cloud Security Controls Pack
-- ============================================================

INSERT INTO marketplace_packages (
    id, publisher_id, package_slug, name, description, long_description,
    package_type, category, applicable_frameworks, applicable_regions,
    applicable_industries, tags, current_version, min_platform_version,
    pricing_model, price_eur, download_count, install_count,
    rating_avg, rating_count, featured, contents_summary,
    status, published_at, license
) VALUES (
    'b0000000-0000-4000-8000-000000000003',
    'a0000000-0000-4000-8000-000000000001',
    'cloud-security-controls',
    'Cloud Security Controls Pack',
    'Platform-agnostic cloud security controls covering IaaS, PaaS, and SaaS workloads with mappings to CSA CCM, ISO 27017, and NIST 800-53 cloud-specific controls.',
    E'## Cloud Security Controls Pack\n\nA comprehensive set of 30 cloud security controls applicable to AWS, Azure, GCP, and multi-cloud environments.\n\n### What''s Included\n- 30 platform-agnostic cloud security controls\n- Shared responsibility model documentation templates\n- Cloud architecture security review checklists\n- Mappings to CSA CCM v4, ISO 27017, ISO 27001, and NIST 800-53\n- Infrastructure-as-Code security scanning guidance\n- Cloud-native incident response procedures\n\n### Who This Is For\n- Cloud-native organisations\n- Organisations migrating to cloud\n- Multi-cloud environments\n- SaaS providers\n- DevOps and Platform Engineering teams',
    'control_pack',
    'technology',
    ARRAY['ISO27001', 'NIST_800_53', 'NIST_CSF_2'],
    ARRAY['Global'],
    ARRAY['technology', 'saas', 'cloud_services', 'fintech', 'healthcare'],
    ARRAY['cloud-security', 'aws', 'azure', 'gcp', 'iaas', 'paas', 'saas', 'devops', 'iac', 'ccm'],
    '1.0.0',
    '1.0.0',
    'free',
    0.00,
    0,
    0,
    0.00,
    0,
    true,
    '{
        "total_controls": 30,
        "control_categories": [
            {"name": "Identity & Access Management", "count": 5},
            {"name": "Data Protection in Cloud", "count": 5},
            {"name": "Network Security", "count": 4},
            {"name": "Compute & Container Security", "count": 5},
            {"name": "Logging & Monitoring", "count": 4},
            {"name": "DevSecOps & IaC", "count": 4},
            {"name": "Incident Response (Cloud)", "count": 3}
        ],
        "framework_mappings": 3,
        "evidence_templates": 15,
        "test_procedures": 30
    }'::jsonb,
    'published',
    now(),
    'CC-BY-4.0'
);

INSERT INTO marketplace_package_versions (
    id, package_id, version, release_notes, package_data, package_hash, file_size_bytes,
    is_breaking_change, migration_notes, published_at
) VALUES (
    'c0000000-0000-4000-8000-000000000003',
    'b0000000-0000-4000-8000-000000000003',
    '1.0.0',
    'Initial release of the Cloud Security Controls Pack with 30 controls, CSA CCM mappings, and IaC scanning guidance.',
    '{
        "schema_version": "1.0",
        "package_type": "control_pack",
        "controls": [
            {
                "code": "CSC-IAM-001",
                "title": "Cloud Identity Federation",
                "description": "Implement identity federation between corporate IdP and all cloud provider accounts using SAML 2.0 or OIDC.",
                "category": "Identity & Access Management",
                "control_type": "preventive",
                "implementation_type": "technical",
                "guidance": "Configure SAML 2.0 or OIDC federation with your corporate IdP (Azure AD, Okta, etc.) for all cloud accounts. Disable local cloud account passwords for human users. Enforce MFA at the IdP level. Implement session duration limits (max 12 hours).",
                "test_procedure": "Verify federation configuration for all accounts. Test that local passwords are disabled. Verify MFA enforcement. Check session timeout settings.",
                "evidence_requirements": ["Federation configuration", "MFA policy", "Session timeout settings"],
                "mappings": {"ISO27001": ["A.9.2.1", "A.9.4.2"], "NIST_800_53": ["IA-2", "IA-5"], "NIST_CSF_2": ["PR.AC-1"]}
            },
            {
                "code": "CSC-IAM-002",
                "title": "Cloud Least-Privilege Access",
                "description": "Implement least-privilege access for all cloud IAM roles and policies using just-in-time elevation.",
                "category": "Identity & Access Management",
                "control_type": "preventive",
                "implementation_type": "technical",
                "guidance": "Audit all IAM policies for overly permissive access. Remove wildcard permissions. Implement JIT privilege elevation for administrative actions. Review access quarterly using cloud-native access analysers (AWS IAM Access Analyzer, Azure PIM).",
                "test_procedure": "Run IAM access analyser. Review policies for wildcard permissions. Verify JIT elevation is enforced. Check quarterly review records.",
                "evidence_requirements": ["Access analyser report", "JIT configuration", "Quarterly review records"],
                "mappings": {"ISO27001": ["A.9.1.2", "A.9.2.3"], "NIST_800_53": ["AC-6", "AC-2"], "NIST_CSF_2": ["PR.AC-4"]}
            },
            {
                "code": "CSC-IAM-003",
                "title": "Service Account Governance",
                "description": "Govern service accounts and machine identities with automated rotation, scope limitation, and monitoring.",
                "category": "Identity & Access Management",
                "control_type": "preventive",
                "implementation_type": "technical",
                "guidance": "Inventory all service accounts across cloud providers. Enforce credential rotation (max 90 days). Use workload identity federation where available. Monitor for anomalous service account activity. Disable unused service accounts after 30 days.",
                "test_procedure": "Review service account inventory. Verify rotation schedules. Check for unused accounts. Review monitoring alerts.",
                "evidence_requirements": ["Service account inventory", "Rotation evidence", "Monitoring dashboard"],
                "mappings": {"ISO27001": ["A.9.2.5", "A.9.4.3"], "NIST_800_53": ["IA-5", "AC-2"], "NIST_CSF_2": ["PR.AC-1"]}
            },
            {
                "code": "CSC-IAM-004",
                "title": "Cross-Account Access Controls",
                "description": "Implement secure cross-account access patterns for multi-account cloud architectures.",
                "category": "Identity & Access Management",
                "control_type": "preventive",
                "implementation_type": "technical",
                "guidance": "Use role assumption (AssumeRole) rather than static credentials for cross-account access. Implement external ID requirements. Restrict trust relationships to known accounts. Monitor cross-account API calls.",
                "test_procedure": "Review cross-account trust policies. Verify external ID requirements. Check for static credential usage. Review CloudTrail/Activity Log for cross-account events.",
                "evidence_requirements": ["Trust policy configurations", "Cross-account access logs"],
                "mappings": {"ISO27001": ["A.9.1.2"], "NIST_800_53": ["AC-3", "AC-17"], "NIST_CSF_2": ["PR.AC-3"]}
            },
            {
                "code": "CSC-IAM-005",
                "title": "Emergency Break-Glass Procedures",
                "description": "Maintain break-glass accounts and procedures for cloud access during IdP or MFA outages.",
                "category": "Identity & Access Management",
                "control_type": "corrective",
                "implementation_type": "administrative",
                "guidance": "Maintain 2-3 break-glass accounts per cloud provider with strong static credentials stored in physical safes. Alert on any break-glass usage. Test procedures quarterly. Rotate credentials after each use.",
                "test_procedure": "Verify break-glass accounts exist. Check alerting configuration. Review quarterly test records. Verify post-use rotation.",
                "evidence_requirements": ["Break-glass procedure", "Alerting configuration", "Test records"],
                "mappings": {"ISO27001": ["A.9.2.3", "A.11.1.1"], "NIST_800_53": ["AC-2", "CP-2"], "NIST_CSF_2": ["PR.AC-4"]}
            },
            {
                "code": "CSC-DP-001",
                "title": "Cloud Storage Encryption",
                "description": "Enforce encryption at rest for all cloud storage services using customer-managed keys where data is classified Confidential or above.",
                "category": "Data Protection in Cloud",
                "control_type": "preventive",
                "implementation_type": "technical",
                "guidance": "Enable default encryption for all storage services (S3, Blob, GCS). Use customer-managed KMS keys (CMK) for Confidential and Restricted data. Implement key rotation annually. Prevent unencrypted storage via SCP/Policy.",
                "test_procedure": "Scan all storage resources for encryption status. Verify CMK usage for classified data. Check key rotation schedules. Verify preventive policies.",
                "evidence_requirements": ["Encryption scan report", "KMS key inventory", "Policy configuration"],
                "mappings": {"ISO27001": ["A.10.1.1", "A.8.2.3"], "NIST_800_53": ["SC-28", "SC-12"], "NIST_CSF_2": ["PR.DS-1"]}
            },
            {
                "code": "CSC-DP-002",
                "title": "Cloud Data Classification Tagging",
                "description": "Implement mandatory resource tagging that maps to the organisation data classification policy.",
                "category": "Data Protection in Cloud",
                "control_type": "preventive",
                "implementation_type": "technical",
                "guidance": "Define mandatory tags: DataClassification, DataOwner, CostCentre, Environment. Enforce tagging via cloud policies (AWS SCP, Azure Policy, GCP Org Policies). Block resource creation without required tags. Audit compliance weekly.",
                "test_procedure": "Review tagging policy configuration. Verify enforcement is active. Run compliance report for untagged resources.",
                "evidence_requirements": ["Tagging policy", "Enforcement configuration", "Compliance report"],
                "mappings": {"ISO27001": ["A.8.2.1", "A.8.1.1"], "NIST_800_53": ["RA-2", "SC-16"], "NIST_CSF_2": ["ID.AM-1"]}
            },
            {
                "code": "CSC-DP-003",
                "title": "Cloud Data Loss Prevention",
                "description": "Deploy cloud-native DLP services to detect and prevent sensitive data exposure in storage and transit.",
                "category": "Data Protection in Cloud",
                "control_type": "detective",
                "implementation_type": "technical",
                "guidance": "Enable cloud DLP scanning (AWS Macie, Azure Information Protection, GCP DLP API) for all storage containing customer data. Configure custom detectors for industry-specific data patterns. Alert on and quarantine exposed sensitive data.",
                "test_procedure": "Verify DLP service coverage. Review custom detector configurations. Test detection with sample sensitive data. Check alert routing.",
                "evidence_requirements": ["DLP configuration", "Detection rule inventory", "Alert samples"],
                "mappings": {"ISO27001": ["A.13.2.1", "A.8.2.3"], "NIST_800_53": ["SC-7", "SI-4"], "NIST_CSF_2": ["PR.DS-5"]}
            },
            {
                "code": "CSC-DP-004",
                "title": "Cloud Backup Encryption and Isolation",
                "description": "Encrypt cloud backups and store in separate accounts/regions with immutability protections.",
                "category": "Data Protection in Cloud",
                "control_type": "corrective",
                "implementation_type": "technical",
                "guidance": "Store backups in dedicated backup accounts. Enable vault lock/immutability to prevent deletion (ransomware protection). Encrypt with separate KMS keys from production. Test restore procedures quarterly. Implement cross-region replication for critical data.",
                "test_procedure": "Verify backup account isolation. Check immutability configuration. Verify separate key usage. Review restore test records.",
                "evidence_requirements": ["Backup architecture", "Immutability configuration", "Restore test records"],
                "mappings": {"ISO27001": ["A.12.3.1", "A.17.1.2"], "NIST_800_53": ["CP-9", "CP-6"], "NIST_CSF_2": ["PR.IP-4"]}
            },
            {
                "code": "CSC-DP-005",
                "title": "Secrets Management",
                "description": "Centralise secrets management using cloud-native vaults with automated rotation and access auditing.",
                "category": "Data Protection in Cloud",
                "control_type": "preventive",
                "implementation_type": "technical",
                "guidance": "Use cloud secrets managers (AWS Secrets Manager, Azure Key Vault, GCP Secret Manager) for all application secrets. Enable automated rotation. Prevent secrets in source code via pre-commit hooks. Audit all secret access events.",
                "test_procedure": "Review secrets inventory. Verify automated rotation configuration. Scan repositories for embedded secrets. Review access audit logs.",
                "evidence_requirements": ["Secrets inventory", "Rotation configuration", "Repository scan results", "Audit logs"],
                "mappings": {"ISO27001": ["A.10.1.2", "A.9.4.3"], "NIST_800_53": ["IA-5", "SC-12"], "NIST_CSF_2": ["PR.DS-1"]}
            },
            {
                "code": "CSC-NET-001",
                "title": "Cloud Network Segmentation",
                "description": "Implement network segmentation using VPCs/VNets with strict inter-segment access controls.",
                "category": "Network Security",
                "control_type": "preventive",
                "implementation_type": "technical",
                "guidance": "Create separate VPCs for production, staging, and development. Implement private subnets for backend services. Use security groups and NACLs for defence in depth. Restrict inter-VPC communication to approved flows via transit gateway.",
                "test_procedure": "Review VPC architecture. Verify security group rules. Test inter-VPC connectivity restrictions. Map approved vs actual traffic flows.",
                "evidence_requirements": ["Network architecture diagram", "Security group rules", "Traffic flow analysis"],
                "mappings": {"ISO27001": ["A.13.1.1", "A.13.1.3"], "NIST_800_53": ["SC-7", "AC-4"], "NIST_CSF_2": ["PR.AC-5"]}
            },
            {
                "code": "CSC-NET-002",
                "title": "Public Exposure Prevention",
                "description": "Prevent accidental public exposure of cloud resources through preventive policies and continuous scanning.",
                "category": "Network Security",
                "control_type": "preventive",
                "implementation_type": "technical",
                "guidance": "Block public S3 bucket creation via account-level settings. Prevent public IP assignment to instances via SCP. Deploy continuous scanning for public-facing resources. Alert on any new public exposure. Implement exception approval workflow.",
                "test_procedure": "Verify account-level public access blocks. Run public exposure scan. Review exception approvals. Check alerting configuration.",
                "evidence_requirements": ["Account settings", "Scan results", "Exception register"],
                "mappings": {"ISO27001": ["A.13.1.1", "A.14.1.2"], "NIST_800_53": ["SC-7", "CM-7"], "NIST_CSF_2": ["PR.AC-5"]}
            },
            {
                "code": "CSC-NET-003",
                "title": "WAF and DDoS Protection",
                "description": "Deploy Web Application Firewall and DDoS protection for all internet-facing cloud workloads.",
                "category": "Network Security",
                "control_type": "preventive",
                "implementation_type": "technical",
                "guidance": "Deploy managed WAF (AWS WAF, Azure WAF, Cloud Armor) with OWASP core rule set for all public endpoints. Enable DDoS protection services. Implement rate limiting and geo-blocking where appropriate. Review and tune WAF rules monthly.",
                "test_procedure": "Verify WAF deployment coverage. Review rule sets. Test DDoS protection configuration. Check monthly tuning records.",
                "evidence_requirements": ["WAF configuration", "Rule set inventory", "DDoS protection status", "Tuning records"],
                "mappings": {"ISO27001": ["A.13.1.1", "A.14.2.5"], "NIST_800_53": ["SC-5", "SC-7"], "NIST_CSF_2": ["PR.PT-4"]}
            },
            {
                "code": "CSC-NET-004",
                "title": "Private Connectivity",
                "description": "Use private endpoints and private links to keep data-plane traffic off the public internet.",
                "category": "Network Security",
                "control_type": "preventive",
                "implementation_type": "technical",
                "guidance": "Implement VPC endpoints / Private Link for all supported cloud services. Route database, storage, and API traffic through private endpoints. Block public endpoints for services where private connectivity is available. Monitor for traffic leaking to public paths.",
                "test_procedure": "Inventory private endpoint configurations. Verify public endpoint blocks. Monitor traffic paths for public leaks.",
                "evidence_requirements": ["Private endpoint inventory", "Configuration evidence", "Traffic analysis"],
                "mappings": {"ISO27001": ["A.13.1.1"], "NIST_800_53": ["SC-7", "SC-8"], "NIST_CSF_2": ["PR.DS-2"]}
            },
            {
                "code": "CSC-CCS-001",
                "title": "Container Image Security",
                "description": "Scan container images for vulnerabilities and enforce signed image policies before deployment.",
                "category": "Compute & Container Security",
                "control_type": "preventive",
                "implementation_type": "technical",
                "guidance": "Integrate image scanning into CI/CD pipelines. Block deployment of images with Critical/High CVEs. Implement image signing and verification. Use minimal base images (distroless/scratch). Scan running containers for drift.",
                "test_procedure": "Review CI/CD scanning integration. Test deployment blocking for vulnerable images. Verify image signing. Check runtime scanning.",
                "evidence_requirements": ["Pipeline configuration", "Scan results", "Signing policy", "Runtime scan reports"],
                "mappings": {"ISO27001": ["A.12.6.1", "A.14.2.5"], "NIST_800_53": ["SI-2", "SA-11"], "NIST_CSF_2": ["PR.IP-12"]}
            },
            {
                "code": "CSC-CCS-002",
                "title": "Kubernetes Security Hardening",
                "description": "Harden Kubernetes clusters according to CIS benchmarks with network policies, RBAC, and pod security.",
                "category": "Compute & Container Security",
                "control_type": "preventive",
                "implementation_type": "technical",
                "guidance": "Apply CIS Kubernetes Benchmark. Enforce Pod Security Standards (Restricted profile). Implement network policies for pod-to-pod communication. Enable audit logging. Disable anonymous authentication. Rotate certificates and tokens.",
                "test_procedure": "Run CIS benchmark assessment. Review pod security policies. Verify network policies. Check audit log configuration.",
                "evidence_requirements": ["CIS benchmark report", "Pod security configuration", "Network policies", "Audit logs"],
                "mappings": {"ISO27001": ["A.12.1.2", "A.14.2.5"], "NIST_800_53": ["CM-6", "AC-3"], "NIST_CSF_2": ["PR.IP-1"]}
            },
            {
                "code": "CSC-CCS-003",
                "title": "Serverless Security Controls",
                "description": "Implement security controls for serverless functions including least-privilege execution roles and dependency scanning.",
                "category": "Compute & Container Security",
                "control_type": "preventive",
                "implementation_type": "technical",
                "guidance": "Assign minimum-privilege IAM roles to each function. Scan function dependencies for vulnerabilities. Set memory and timeout limits. Implement function-level network restrictions (VPC placement where needed). Monitor for cold-start anomalies.",
                "test_procedure": "Review function IAM roles for over-permission. Run dependency scans. Verify resource limits. Check network configuration.",
                "evidence_requirements": ["IAM role review", "Dependency scan results", "Resource configuration"],
                "mappings": {"ISO27001": ["A.9.1.2", "A.12.6.1"], "NIST_800_53": ["AC-6", "SI-2"], "NIST_CSF_2": ["PR.AC-4"]}
            },
            {
                "code": "CSC-CCS-004",
                "title": "VM Image Hardening Pipeline",
                "description": "Automate the creation of hardened VM images using CIS benchmarks with continuous validation.",
                "category": "Compute & Container Security",
                "control_type": "preventive",
                "implementation_type": "technical",
                "guidance": "Build golden images via automated pipeline (Packer, EC2 Image Builder). Apply CIS Level 2 hardening. Scan images before publishing to shared gallery. Rebuild images monthly for patch incorporation. Prevent deployment of non-approved images.",
                "test_procedure": "Review image pipeline configuration. Run CIS benchmark on latest golden image. Verify deployment controls. Check rebuild schedule.",
                "evidence_requirements": ["Pipeline configuration", "CIS scan results", "Deployment policy", "Rebuild records"],
                "mappings": {"ISO27001": ["A.12.1.2", "A.12.6.1"], "NIST_800_53": ["CM-2", "CM-6"], "NIST_CSF_2": ["PR.IP-1"]}
            },
            {
                "code": "CSC-CCS-005",
                "title": "Immutable Infrastructure Enforcement",
                "description": "Enforce immutable infrastructure patterns to prevent configuration drift and unauthorised changes.",
                "category": "Compute & Container Security",
                "control_type": "preventive",
                "implementation_type": "technical",
                "guidance": "Disable SSH/RDP access to production instances by default. Deploy changes via CI/CD pipeline only (blue-green/canary). Monitor for out-of-band changes. Alert on and auto-remediate drift from desired state.",
                "test_procedure": "Verify remote access is disabled. Review deployment pipeline. Check drift detection configuration. Review auto-remediation logs.",
                "evidence_requirements": ["Remote access policy", "Pipeline evidence", "Drift reports", "Remediation logs"],
                "mappings": {"ISO27001": ["A.12.1.2", "A.12.5.1"], "NIST_800_53": ["CM-3", "CM-5"], "NIST_CSF_2": ["PR.IP-3"]}
            },
            {
                "code": "CSC-LM-001",
                "title": "Centralised Cloud Logging",
                "description": "Aggregate all cloud provider logs into a centralised, immutable logging platform.",
                "category": "Logging & Monitoring",
                "control_type": "detective",
                "implementation_type": "technical",
                "guidance": "Enable CloudTrail, Activity Log, or Audit Logs in all accounts/projects. Forward to centralised SIEM with immutable retention. Retain for minimum 12 months online, 7 years archive. Protect logs from deletion with separate account controls.",
                "test_procedure": "Verify logging is enabled in all accounts. Check centralisation configuration. Verify immutability. Review retention settings.",
                "evidence_requirements": ["Logging configuration", "Centralisation evidence", "Retention policy"],
                "mappings": {"ISO27001": ["A.12.4.1", "A.12.4.3"], "NIST_800_53": ["AU-2", "AU-6"], "NIST_CSF_2": ["DE.AE-3"]}
            },
            {
                "code": "CSC-LM-002",
                "title": "Cloud Security Posture Management",
                "description": "Deploy CSPM tooling to continuously assess cloud configuration against security benchmarks.",
                "category": "Logging & Monitoring",
                "control_type": "detective",
                "implementation_type": "technical",
                "guidance": "Deploy CSPM across all cloud accounts (AWS Security Hub, Azure Defender for Cloud, GCP SCC). Benchmark against CIS Foundations and custom policies. Prioritise findings by risk. Integrate with ticketing for remediation tracking.",
                "test_procedure": "Verify CSPM coverage across all accounts. Review benchmark scores. Check ticketing integration. Review remediation SLA compliance.",
                "evidence_requirements": ["CSPM dashboard", "Benchmark scores", "Remediation metrics"],
                "mappings": {"ISO27001": ["A.18.2.3", "A.12.6.1"], "NIST_800_53": ["CA-7", "RA-5"], "NIST_CSF_2": ["DE.CM-8"]}
            },
            {
                "code": "CSC-LM-003",
                "title": "Cloud Cost Anomaly Detection",
                "description": "Monitor cloud spending for anomalies that may indicate security incidents (crypto-mining, data exfiltration).",
                "category": "Logging & Monitoring",
                "control_type": "detective",
                "implementation_type": "technical",
                "guidance": "Enable cloud cost anomaly detection services. Set budget alerts with security team notification. Monitor for unexpected compute spin-up, data transfer spikes, and new region usage. Investigate all anomalies within 4 hours.",
                "test_procedure": "Review anomaly detection configuration. Verify alerting. Check investigation SLA for recent anomalies.",
                "evidence_requirements": ["Anomaly detection config", "Alert configuration", "Investigation records"],
                "mappings": {"ISO27001": ["A.12.4.1"], "NIST_800_53": ["SI-4", "AU-6"], "NIST_CSF_2": ["DE.AE-2"]}
            },
            {
                "code": "CSC-LM-004",
                "title": "Runtime Threat Detection",
                "description": "Deploy cloud-native runtime threat detection for compute, container, and serverless workloads.",
                "category": "Logging & Monitoring",
                "control_type": "detective",
                "implementation_type": "technical",
                "guidance": "Enable GuardDuty, Defender for Cloud, or SCC Threat Detection across all accounts. Configure custom threat models for your workload patterns. Integrate detections with SIEM and incident response workflow. Tune to reduce false positives quarterly.",
                "test_procedure": "Verify runtime detection coverage. Review custom detections. Check SIEM integration. Review false positive tuning records.",
                "evidence_requirements": ["Detection service configuration", "Custom rules", "SIEM integration", "Tuning records"],
                "mappings": {"ISO27001": ["A.12.4.1", "A.16.1.2"], "NIST_800_53": ["SI-4", "IR-4"], "NIST_CSF_2": ["DE.CM-1"]}
            },
            {
                "code": "CSC-DSO-001",
                "title": "Infrastructure as Code Security Scanning",
                "description": "Integrate IaC security scanning into CI/CD pipelines to detect misconfigurations before deployment.",
                "category": "DevSecOps & IaC",
                "control_type": "preventive",
                "implementation_type": "technical",
                "guidance": "Integrate tools like Checkov, tfsec, or cfn-lint into pull request pipelines. Block merges with Critical/High findings. Maintain custom rules for organisation-specific policies. Report scan metrics to security team weekly.",
                "test_procedure": "Verify pipeline integration. Test merge blocking. Review custom rule sets. Check weekly reporting.",
                "evidence_requirements": ["Pipeline configuration", "Scan results sample", "Custom rules", "Weekly reports"],
                "mappings": {"ISO27001": ["A.14.2.1", "A.14.2.5"], "NIST_800_53": ["SA-11", "CM-6"], "NIST_CSF_2": ["PR.IP-12"]}
            },
            {
                "code": "CSC-DSO-002",
                "title": "GitOps Deployment Controls",
                "description": "Enforce GitOps workflows where all infrastructure and application changes flow through version-controlled, reviewed code.",
                "category": "DevSecOps & IaC",
                "control_type": "preventive",
                "implementation_type": "technical",
                "guidance": "All changes to production must go through Git with mandatory code review. Enforce branch protection with minimum 1 approval. Implement automated rollback on failure. Maintain audit trail of all deployments via Git history.",
                "test_procedure": "Verify branch protection rules. Review approval requirements. Test automated rollback. Verify deployment audit trail.",
                "evidence_requirements": ["Branch protection config", "Approval records", "Rollback test results", "Deployment history"],
                "mappings": {"ISO27001": ["A.12.1.2", "A.14.2.2"], "NIST_800_53": ["CM-3", "CM-5"], "NIST_CSF_2": ["PR.IP-3"]}
            },
            {
                "code": "CSC-DSO-003",
                "title": "Software Supply Chain Security",
                "description": "Secure the software supply chain for cloud deployments including dependency scanning and SBOM generation.",
                "category": "DevSecOps & IaC",
                "control_type": "preventive",
                "implementation_type": "technical",
                "guidance": "Generate SBOMs for all deployed applications. Scan dependencies for known vulnerabilities in CI/CD. Use lock files to pin dependency versions. Verify package integrity via checksums. Monitor for newly disclosed vulnerabilities in deployed dependencies.",
                "test_procedure": "Verify SBOM generation. Review dependency scan results. Check lock file usage. Verify integrity checks. Review CVE monitoring.",
                "evidence_requirements": ["SBOM samples", "Dependency scan results", "Lock files", "CVE monitoring dashboard"],
                "mappings": {"ISO27001": ["A.14.2.7", "A.14.2.5"], "NIST_800_53": ["SA-12", "SI-7"], "NIST_CSF_2": ["PR.IP-12"]}
            },
            {
                "code": "CSC-DSO-004",
                "title": "Environment Promotion Controls",
                "description": "Implement controlled promotion of changes through development, staging, and production environments.",
                "category": "DevSecOps & IaC",
                "control_type": "preventive",
                "implementation_type": "administrative",
                "guidance": "Enforce sequential promotion: dev -> staging -> production. Require all tests to pass in staging before production promotion. Implement change freezes during high-risk periods. Maintain separate credentials and configurations per environment.",
                "test_procedure": "Review promotion workflow. Verify test gate requirements. Check change freeze policy. Verify environment separation.",
                "evidence_requirements": ["Promotion workflow", "Test gate configuration", "Change freeze records", "Environment configuration"],
                "mappings": {"ISO27001": ["A.12.1.4", "A.14.2.6"], "NIST_800_53": ["CM-3", "SA-10"], "NIST_CSF_2": ["PR.IP-3"]}
            },
            {
                "code": "CSC-IR-001",
                "title": "Cloud Incident Response Playbook",
                "description": "Maintain cloud-specific incident response playbooks covering credential compromise, data exposure, and resource hijacking.",
                "category": "Incident Response (Cloud)",
                "control_type": "corrective",
                "implementation_type": "administrative",
                "guidance": "Develop playbooks for: compromised IAM credentials, public S3/Blob exposure, crypto-mining detection, insider threat in cloud console, and supply chain compromise. Include automated containment actions. Test quarterly via tabletop exercises.",
                "test_procedure": "Review playbook inventory. Verify automated containment steps. Check tabletop exercise records. Verify playbook update dates.",
                "evidence_requirements": ["Playbook documents", "Automation configuration", "Tabletop exercise records"],
                "mappings": {"ISO27001": ["A.16.1.5", "A.16.1.1"], "NIST_800_53": ["IR-8", "IR-4"], "NIST_CSF_2": ["RS.RP-1"]}
            },
            {
                "code": "CSC-IR-002",
                "title": "Cloud Forensics Capability",
                "description": "Maintain forensic investigation capabilities for cloud environments including snapshot capture and log preservation.",
                "category": "Incident Response (Cloud)",
                "control_type": "corrective",
                "implementation_type": "technical",
                "guidance": "Pre-configure forensic tooling: EBS snapshot automation, memory capture capability, network traffic mirroring. Maintain a dedicated forensic account. Train incident response team on cloud-specific forensic techniques. Document chain of custody procedures for cloud evidence.",
                "test_procedure": "Verify forensic tooling is pre-configured. Test snapshot capture. Review forensic account access. Check training records.",
                "evidence_requirements": ["Forensic tooling config", "Snapshot test results", "Training records", "Chain of custody procedure"],
                "mappings": {"ISO27001": ["A.16.1.7", "A.16.1.5"], "NIST_800_53": ["IR-4", "AU-7"], "NIST_CSF_2": ["RS.AN-3"]}
            },
            {
                "code": "CSC-IR-003",
                "title": "Automated Containment Actions",
                "description": "Implement automated containment actions that can isolate compromised cloud resources within minutes.",
                "category": "Incident Response (Cloud)",
                "control_type": "corrective",
                "implementation_type": "technical",
                "guidance": "Pre-build Lambda/Functions for: revoking IAM sessions, isolating EC2/VM instances via security group swap, disabling access keys, blocking S3 bucket public access, and quarantining containers. Integrate with SOAR platform.",
                "test_procedure": "Review automated containment functions. Test each function in non-production. Verify SOAR integration. Check execution logs.",
                "evidence_requirements": ["Containment function code", "Test results", "SOAR integration", "Execution logs"],
                "mappings": {"ISO27001": ["A.16.1.5", "A.16.1.2"], "NIST_800_53": ["IR-4", "IR-6"], "NIST_CSF_2": ["RS.MI-1"]}
            }
        ]
    }'::jsonb,
    'd4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3',
    312000,
    false,
    NULL,
    now()
);

COMMIT;
