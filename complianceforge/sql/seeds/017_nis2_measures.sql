-- ============================================================
-- ComplianceForge -- Seed: NIS2 Article 21 Security Measures
-- The 10 mandatory cybersecurity risk-management measures
-- required under NIS2 Directive Article 21(2)(a)-(j).
-- Uses placeholder organization_id for template data.
-- ============================================================

INSERT INTO nis2_security_measures (
    organization_id, measure_code, measure_title, measure_description, article_reference,
    implementation_status, notes
) VALUES

-- (a) Risk analysis and information system security policies
(
    '00000000-0000-0000-0000-000000000000',
    'NIS2-Art21-a',
    'Policies on risk analysis and information system security',
    'Establish and maintain policies for risk analysis and information system security. '
    'This includes defining an overall risk management approach, identifying and assessing '
    'cybersecurity risks, and ensuring that appropriate security policies are in place to '
    'govern the protection of network and information systems.',
    'Article 21(2)(a)',
    'not_started',
    'Foundation measure: requires documented risk analysis methodology and information security policy framework.'
),

-- (b) Incident handling
(
    '00000000-0000-0000-0000-000000000000',
    'NIS2-Art21-b',
    'Incident handling',
    'Implement processes and procedures for incident handling, including prevention, detection, '
    'analysis, containment, response, and recovery from cybersecurity incidents. This measure '
    'must integrate with the NIS2 Article 23 incident reporting obligations (early warning '
    'within 24 hours, notification within 72 hours, final report within 1 month).',
    'Article 21(2)(b)',
    'not_started',
    'Closely linked to Article 23 reporting obligations. Must include CSIRT notification workflows.'
),

-- (c) Business continuity and crisis management
(
    '00000000-0000-0000-0000-000000000000',
    'NIS2-Art21-c',
    'Business continuity, backup management, disaster recovery, and crisis management',
    'Ensure business continuity including backup management, disaster recovery, and crisis '
    'management. Organisations must have tested plans to maintain or rapidly restore critical '
    'services following a cybersecurity incident, including regular backup and recovery testing.',
    'Article 21(2)(c)',
    'not_started',
    'Requires documented BCP/DRP, regular backup testing, and crisis management procedures.'
),

-- (d) Supply chain security
(
    '00000000-0000-0000-0000-000000000000',
    'NIS2-Art21-d',
    'Supply chain security',
    'Address supply chain security including security-related aspects concerning the relationships '
    'between each entity and its direct suppliers or service providers. This includes assessing '
    'the overall quality and resilience of products and cybersecurity practices of suppliers, '
    'and incorporating security requirements into contractual arrangements.',
    'Article 21(2)(d)',
    'not_started',
    'Requires supplier risk assessment, contractual security clauses, and ongoing monitoring of supplier security posture.'
),

-- (e) Security in acquisition, development, and maintenance
(
    '00000000-0000-0000-0000-000000000000',
    'NIS2-Art21-e',
    'Security in network and information systems acquisition, development, and maintenance',
    'Ensure security throughout the lifecycle of network and information systems, including '
    'acquisition, development, and maintenance phases. This covers vulnerability handling and '
    'disclosure, secure development practices, and security testing throughout the system lifecycle.',
    'Article 21(2)(e)',
    'not_started',
    'Covers SDLC security, vulnerability management, patch management, and secure procurement.'
),

-- (f) Assessing effectiveness of cybersecurity measures
(
    '00000000-0000-0000-0000-000000000000',
    'NIS2-Art21-f',
    'Policies and procedures for assessing the effectiveness of cybersecurity risk-management measures',
    'Establish policies and procedures to assess the effectiveness of cybersecurity risk-management '
    'measures. This includes regular security audits, penetration testing, vulnerability assessments, '
    'and metrics-driven evaluation of the cybersecurity posture.',
    'Article 21(2)(f)',
    'not_started',
    'Requires regular security assessments, KPI tracking, and continuous improvement processes.'
),

-- (g) Basic cyber hygiene and training
(
    '00000000-0000-0000-0000-000000000000',
    'NIS2-Art21-g',
    'Basic cyber hygiene practices and cybersecurity training',
    'Implement basic cyber hygiene practices and provide regular cybersecurity training to all '
    'staff. This includes awareness programmes covering phishing, password management, secure '
    'use of devices, social engineering, and role-specific technical training for IT and security staff.',
    'Article 21(2)(g)',
    'not_started',
    'NIS2 Article 20(2) specifically requires management body members to undergo cybersecurity training.'
),

-- (h) Cryptography and encryption
(
    '00000000-0000-0000-0000-000000000000',
    'NIS2-Art21-h',
    'Policies and procedures regarding the use of cryptography and, where appropriate, encryption',
    'Define and implement policies and procedures for the use of cryptography and encryption '
    'to protect the confidentiality, integrity, and authenticity of data. This includes key '
    'management, encryption standards for data at rest and in transit, and cryptographic controls '
    'for communications.',
    'Article 21(2)(h)',
    'not_started',
    'Covers encryption standards, key management lifecycle, TLS/mTLS requirements, and data classification-driven encryption.'
),

-- (i) Human resources security, access control, asset management
(
    '00000000-0000-0000-0000-000000000000',
    'NIS2-Art21-i',
    'Human resources security, access control policies, and asset management',
    'Implement human resources security measures, access control policies, and asset management '
    'practices. This includes personnel screening, role-based access control (RBAC), least '
    'privilege principles, privileged access management, and comprehensive asset inventory '
    'and lifecycle management.',
    'Article 21(2)(i)',
    'not_started',
    'Covers HR security (joiners/movers/leavers), identity and access management, PAM, and CMDB/asset register.'
),

-- (j) Multi-factor authentication and secured communications
(
    '00000000-0000-0000-0000-000000000000',
    'NIS2-Art21-j',
    'Use of multi-factor authentication, secured voice/video/text communications, and secured emergency communication systems',
    'Deploy multi-factor authentication (MFA) or continuous authentication solutions, ensure '
    'secured voice, video, and text communications within the entity, and establish secured '
    'emergency communication systems that can operate during incidents affecting primary channels.',
    'Article 21(2)(j)',
    'not_started',
    'Requires MFA for all privileged and remote access, encrypted communications, and out-of-band emergency comms.'
)
ON CONFLICT DO NOTHING;
