-- ============================================================
-- ComplianceForge GRC Platform
-- Seed 028: Knowledge Base System Articles
-- 50+ compliance guidance articles covering ISO 27001,
-- GDPR, NIS2, Cyber Essentials, PCI DSS, and GRC best practices.
-- All articles are system-level (is_system = true, org_id = NULL).
-- ============================================================

-- ============================================================
-- IMPLEMENTATION GUIDES
-- ============================================================

INSERT INTO knowledge_articles (
    organization_id, article_type, title, slug, content_markdown, summary,
    applicable_frameworks, applicable_control_codes, tags,
    difficulty, reading_time_minutes, is_system, is_published
) VALUES

-- 1. ISO 27001 Annex A.5 Organisational Controls
(NULL, 'implementation_guide',
 'Implementing ISO 27001 Annex A.5: Organisational Controls — A Practical Guide',
 'iso-27001-annex-a5-organisational-controls',
 '## Overview

Annex A.5 of ISO 27001:2022 focuses on organisational controls — the policies, procedures, and processes that form the governance backbone of your ISMS.

## Key Controls

### A.5.1 — Policies for Information Security
Define, approve, publish, and communicate your information security policies. Review at planned intervals or when significant changes occur.

**Implementation Steps:**
1. Establish a policy hierarchy: master policy → topic-specific policies → procedures
2. Assign policy ownership to appropriate management levels
3. Review and approve through a formal process
4. Communicate to all relevant personnel and external parties
5. Schedule regular reviews (at least annually)

### A.5.2 — Information Security Roles and Responsibilities
All information security responsibilities shall be defined and allocated.

### A.5.3 — Segregation of Duties
Conflicting duties and areas of responsibility shall be segregated.

### A.5.4 — Management Responsibilities
Management shall require all employees and contractors to apply information security in accordance with the established policies.

## Evidence Requirements
- Approved policy documents with version control
- Policy acknowledgement records
- Role descriptions with security responsibilities
- Segregation of duties matrix
- Management commitment statements',
 'Practical guide to implementing ISO 27001 Annex A.5 organisational controls including policies, roles, and management responsibilities.',
 '{ISO 27001}', '{A.5.1,A.5.2,A.5.3,A.5.4}',
 '{policy,governance,isms,organisational-controls}',
 'intermediate', 8, true, true),

-- 2. Access Control Implementation
(NULL, 'implementation_guide',
 'Access Control Implementation Checklist (A.5.15, A.8.2, A.8.3, A.8.5)',
 'access-control-implementation-checklist',
 '## Access Control Framework

Implementing robust access control is fundamental to information security. This guide covers the key ISO 27001 controls for access management.

## A.5.15 — Access Control Policy
Define rules to control physical and logical access to information based on business and security requirements.

**Key Elements:**
- Need-to-know principle
- Least privilege assignment
- Role-based access control (RBAC)
- Regular access reviews

## A.8.2 — Privileged Access Rights
Restrict and manage the allocation and use of privileged access rights.

**Implementation:**
1. Maintain an inventory of all privileged accounts
2. Implement just-in-time (JIT) privileged access where possible
3. Use multi-factor authentication for all privileged access
4. Log and monitor all privileged access sessions
5. Review privileged access rights quarterly

## A.8.3 — Information Access Restriction
Access to information and application system functions shall be restricted in accordance with the access control policy.

## A.8.5 — Secure Authentication
Secure authentication technologies and procedures shall be established.

## Checklist
- [ ] Access control policy documented and approved
- [ ] RBAC model defined for all systems
- [ ] Privileged account inventory maintained
- [ ] MFA enabled for all privileged access
- [ ] Quarterly access review process established
- [ ] Joiners/movers/leavers process documented
- [ ] Password policy meets NIST 800-63B guidelines',
 'Comprehensive checklist for implementing access controls per ISO 27001 Annex A.5.15, A.8.2, A.8.3, and A.8.5.',
 '{ISO 27001}', '{A.5.15,A.8.2,A.8.3,A.8.5}',
 '{access-control,rbac,privileged-access,authentication,mfa}',
 'intermediate', 10, true, true),

-- 3. Logging and Monitoring
(NULL, 'implementation_guide',
 'Setting Up Logging and Monitoring for ISO 27001 A.8.15 & A.8.16',
 'logging-monitoring-iso27001-a815-a816',
 '## Why Logging and Monitoring Matter

Effective logging and monitoring provide visibility into security events, support incident detection, and create the audit trail needed for compliance.

## A.8.15 — Logging
Logs that record activities, exceptions, faults, and other relevant events shall be produced, stored, protected, and analysed.

**What to Log:**
- Authentication events (successful and failed)
- Privileged operations
- Data access events
- System configuration changes
- Application errors and exceptions
- Network connection events

**Log Retention:**
- Minimum 12 months for most logs
- 3+ years for financial and regulatory logs
- Ensure tamper-proof storage (WORM or append-only)

## A.8.16 — Monitoring Activities
Networks, systems, and applications shall be monitored for anomalous behaviour and appropriate actions taken.

**Implementation:**
1. Deploy a SIEM solution (e.g., Elastic SIEM, Sentinel)
2. Configure correlation rules for common attack patterns
3. Set up alerting thresholds
4. Establish 24/7 monitoring or SOC coverage
5. Document incident response procedures for alerts
6. Regular tuning of detection rules

## Key Metrics
- Mean time to detect (MTTD)
- Mean time to respond (MTTR)
- False positive rate
- Log coverage percentage',
 'Guide to implementing logging and monitoring controls per ISO 27001 A.8.15 and A.8.16.',
 '{ISO 27001}', '{A.8.15,A.8.16}',
 '{logging,monitoring,siem,detection,audit-trail}',
 'advanced', 12, true, true),

-- 4. Cryptography Controls
(NULL, 'implementation_guide',
 'Cryptography Controls: A.8.24 Implementation for European Enterprises',
 'cryptography-controls-a824-implementation',
 '## Cryptographic Controls Overview

ISO 27001 A.8.24 requires organisations to define rules for the effective use of cryptography, including key management.

## Encryption Standards
- AES-256 for data at rest
- TLS 1.3 for data in transit
- RSA-2048 or ECDSA P-256 minimum for asymmetric operations
- SHA-256 minimum for hashing

## Key Management
1. Generate keys using approved random number generators
2. Distribute keys securely (never over unencrypted channels)
3. Store keys in HSMs or approved key vaults
4. Rotate keys according to policy (annually minimum)
5. Revoke compromised keys immediately
6. Archive keys needed for decrypting historical data

## EU-Specific Considerations
- GDPR Article 32 recommends encryption as a security measure
- eIDAS regulation governs electronic signatures
- Schrems II: encryption for cross-border data transfers
- BSI TR-02102 recommendations for German organisations

## Common Pitfalls
- Hardcoded encryption keys in source code
- Using deprecated algorithms (DES, MD5, SHA-1)
- Not encrypting backup media
- Missing key rotation procedures',
 'Implementation guide for cryptography controls under ISO 27001 A.8.24 with EU-specific considerations.',
 '{ISO 27001,GDPR}', '{A.8.24}',
 '{cryptography,encryption,key-management,tls,gdpr}',
 'advanced', 10, true, true),

-- 5. Vulnerability Management
(NULL, 'implementation_guide',
 'Vulnerability Management Programme: A.8.8 Step-by-Step',
 'vulnerability-management-programme-a88',
 '## Building a Vulnerability Management Programme

A systematic approach to identifying, evaluating, and remediating vulnerabilities is essential for maintaining security posture.

## A.8.8 — Management of Technical Vulnerabilities
Information about technical vulnerabilities of information systems shall be obtained, exposure evaluated, and appropriate measures taken.

## Programme Components

### 1. Asset Inventory
Maintain a complete inventory of all IT assets to be scanned.

### 2. Vulnerability Scanning
- Internal scans: weekly minimum
- External scans: monthly minimum
- Authenticated scans for deeper coverage
- Web application scans for all internet-facing apps

### 3. Prioritisation
Use CVSS scores combined with business context:
- Critical (CVSS 9.0+): remediate within 48 hours
- High (CVSS 7.0-8.9): remediate within 7 days
- Medium (CVSS 4.0-6.9): remediate within 30 days
- Low (CVSS 0.1-3.9): remediate within 90 days

### 4. Remediation
- Patch management process
- Compensating controls when patching is not possible
- Exception process for accepted risks

### 5. Reporting
- Weekly vulnerability summary for security team
- Monthly executive report with trends
- Quarterly board-level metrics',
 'Step-by-step guide to building a vulnerability management programme per ISO 27001 A.8.8.',
 '{ISO 27001,Cyber Essentials}', '{A.8.8}',
 '{vulnerability-management,scanning,patching,cvss}',
 'intermediate', 10, true, true),

-- 6. Incident Response Plan
(NULL, 'implementation_guide',
 'Incident Response Plan: Template and Testing Guide',
 'incident-response-plan-template',
 '## Incident Response Plan

A well-defined incident response plan ensures your organisation can detect, contain, eradicate, and recover from security incidents effectively.

## Plan Structure

### 1. Preparation
- Incident response team roles and contact details
- Communication channels (primary and backup)
- Tool access and credentials
- Escalation matrix

### 2. Detection & Analysis
- Monitoring and alerting systems
- Incident classification criteria
- Severity levels and response SLAs
- Initial assessment checklist

### 3. Containment
- Short-term: isolate affected systems
- Long-term: apply patches, change credentials
- Evidence preservation procedures

### 4. Eradication
- Remove malware and backdoors
- Reset compromised accounts
- Patch exploited vulnerabilities

### 5. Recovery
- System restoration from clean backups
- Verification of system integrity
- Gradual service restoration
- Enhanced monitoring period

### 6. Lessons Learned
- Post-incident review within 5 business days
- Root cause analysis
- Action items with owners and deadlines
- Update detection rules and procedures

## Testing Your Plan
- Tabletop exercises: quarterly
- Simulation exercises: annually
- Full-scale drills: for critical scenarios',
 'Comprehensive incident response plan template with testing guidance.',
 '{ISO 27001,NIS2}', '{A.5.24,A.5.25,A.5.26}',
 '{incident-response,security-incident,containment,recovery}',
 'intermediate', 15, true, true),

-- ============================================================
-- REGULATORY GUIDES
-- ============================================================

-- 7. GDPR ROPA
(NULL, 'regulatory_guide',
 'GDPR Article 30 ROPA: What to Include and How to Maintain It',
 'gdpr-article-30-ropa-guide',
 '## Record of Processing Activities (ROPA)

GDPR Article 30 requires controllers and processors to maintain a written record of processing activities.

## Required Fields (Controller)
1. Name and contact details of the controller
2. Purposes of the processing
3. Categories of data subjects
4. Categories of personal data
5. Categories of recipients
6. Transfers to third countries (with safeguards)
7. Time limits for erasure (retention periods)
8. Technical and organisational security measures

## Required Fields (Processor)
1. Name and contact details of the processor and controller
2. Categories of processing carried out
3. Transfers to third countries
4. Technical and organisational security measures

## Maintenance Best Practices
- Review quarterly or when processing changes
- Link to data flow maps
- Cross-reference with DPIAs
- Integrate with your vendor management programme
- Use ComplianceForge Data Mapping to automate ROPA generation

## Common Mistakes
- Treating ROPA as a one-time exercise
- Not including all processing activities
- Missing third-party processor records
- Outdated retention periods',
 'Guide to creating and maintaining GDPR Article 30 Records of Processing Activities (ROPA).',
 '{GDPR}', '{}',
 '{gdpr,ropa,article-30,data-processing,privacy}',
 'beginner', 8, true, true),

-- 8. GDPR Breach Notification
(NULL, 'regulatory_guide',
 'GDPR Breach Notification: The 72-Hour Workflow Explained',
 'gdpr-breach-notification-72-hour-workflow',
 '## The 72-Hour Notification Requirement

GDPR Articles 33 and 34 establish strict timelines for breach notification.

## Notification to the Supervisory Authority (Article 33)
**Deadline:** 72 hours from becoming aware of the breach.

**Must Include:**
1. Nature of the breach (categories and approximate numbers)
2. Name and contact details of the DPO
3. Likely consequences of the breach
4. Measures taken or proposed to address the breach

## Notification to Data Subjects (Article 34)
**Required when:** The breach is likely to result in a HIGH risk to rights and freedoms.

**Not required when:**
- Data was encrypted/pseudonymised
- Subsequent measures ensure high risk is no longer likely
- Disproportionate effort (use public communication instead)

## The 72-Hour Workflow

### Hour 0-4: Detection & Containment
- Confirm the breach
- Activate incident response team
- Begin containment measures
- Start documentation

### Hour 4-24: Assessment
- Determine scope and impact
- Identify affected data subjects
- Assess risk level
- Engage legal counsel

### Hour 24-48: Preparation
- Draft DPA notification
- Draft data subject notification (if required)
- Prepare Q&A for affected individuals
- Brief senior management

### Hour 48-72: Notification
- Submit notification to DPA
- Send data subject notifications
- Activate communication plan
- Continue remediation

## Documentation
Every breach must be documented regardless of notification requirement.',
 'Step-by-step workflow for GDPR 72-hour breach notification under Articles 33 and 34.',
 '{GDPR}', '{}',
 '{gdpr,breach-notification,72-hours,dpa,data-breach}',
 'intermediate', 12, true, true),

-- 9. NIS2 Compliance Roadmap
(NULL, 'regulatory_guide',
 'NIS2 Compliance Roadmap for Essential and Important Entities',
 'nis2-compliance-roadmap',
 '## NIS2 Directive Overview

The NIS2 Directive (EU 2022/2555) significantly expands cybersecurity requirements across the EU. Member states must transpose it into national law.

## Who Is Affected?
- **Essential Entities:** Energy, transport, banking, health, water, digital infrastructure, public administration, space
- **Important Entities:** Postal services, waste management, chemicals, food, manufacturing, digital providers

## Key Requirements

### 1. Risk Management Measures (Article 21)
- Risk analysis and security policies
- Incident handling
- Business continuity and crisis management
- Supply chain security
- Network and information systems security
- Vulnerability disclosure
- Cybersecurity training
- Cryptography and encryption
- Human resource security
- Multi-factor authentication

### 2. Incident Reporting (Article 23)
- **Early warning:** within 24 hours of becoming aware
- **Incident notification:** within 72 hours
- **Final report:** within one month

### 3. Management Accountability (Article 20)
- Management bodies must approve cybersecurity measures
- Management must undergo training
- Management is personally liable for non-compliance

## Implementation Roadmap
1. **Month 1-2:** Gap assessment against NIS2 requirements
2. **Month 3-4:** Risk assessment and treatment plan
3. **Month 5-8:** Implement technical measures
4. **Month 9-10:** Supply chain security programme
5. **Month 11-12:** Incident response testing and training

## Penalties
- Essential entities: up to EUR 10 million or 2% of worldwide turnover
- Important entities: up to EUR 7 million or 1.4% of worldwide turnover',
 'Comprehensive roadmap for achieving NIS2 compliance for essential and important entities.',
 '{NIS2}', '{}',
 '{nis2,directive,cybersecurity,essential-entities,eu}',
 'intermediate', 15, true, true),

-- 10. Cyber Essentials Guide
(NULL, 'regulatory_guide',
 'Cyber Essentials Certification: Self-Assessment Guide',
 'cyber-essentials-self-assessment-guide',
 '## Cyber Essentials Overview

Cyber Essentials is a UK government-backed scheme that helps organisations protect against the most common cyber attacks.

## Five Technical Controls

### 1. Firewalls
- Boundary firewall or internet gateway configured
- Default admin passwords changed
- Only necessary ports open
- Host-based firewall on all devices

### 2. Secure Configuration
- Remove unnecessary software
- Change default passwords
- Disable auto-run features
- Configure lockout policies

### 3. User Access Control
- Unique user accounts for all users
- Grant minimum privileges needed
- Admin accounts for admin tasks only
- MFA for cloud services and remote access

### 4. Malware Protection
- Anti-malware software installed
- Kept up to date automatically
- Configured to scan files and web pages
- Prevent execution from user profiles

### 5. Security Update Management
- Apply critical patches within 14 days
- Remove unsupported software
- Enable automatic updates where possible
- Licensed software kept current

## Self-Assessment Process
1. Define your scope (IP addresses, devices, software)
2. Complete the Cyber Essentials questionnaire
3. Submit for assessment
4. Receive certification (valid for 12 months)

## Cyber Essentials Plus
Adds hands-on technical verification by a qualified assessor.',
 'Guide to achieving Cyber Essentials certification through self-assessment.',
 '{Cyber Essentials}', '{}',
 '{cyber-essentials,uk,certification,firewall,patching}',
 'beginner', 10, true, true),

-- 11. PCI DSS v4.0 Transition
(NULL, 'regulatory_guide',
 'PCI DSS v4.0 Transition: Key Changes from v3.2.1',
 'pci-dss-v4-transition-guide',
 '## PCI DSS v4.0 Overview

PCI DSS v4.0 was released in March 2022. The transition deadline from v3.2.1 was March 31, 2024. Future-dated requirements become mandatory March 31, 2025.

## Key Changes

### 1. Customised Approach
- New option alongside the traditional Defined Approach
- Allows organisations to meet objectives using alternative controls
- Requires documented risk analysis for each customised control

### 2. Enhanced Authentication
- MFA required for all access to the cardholder data environment
- Password minimum 12 characters (was 7)
- Stronger requirements for service and application accounts

### 3. Targeted Risk Analysis
- Requirement to perform targeted risk analysis for specific controls
- Determines frequency of certain activities (scanning, reviews)
- Must document and justify chosen frequencies

### 4. Expanded Scope of Anti-Phishing
- Anti-phishing mechanisms required (previously only awareness training)
- Technical controls to detect and protect against phishing

### 5. New E-commerce Requirements
- Payment page scripts must be inventoried and authorised
- Mechanisms to detect tampering of payment pages
- HTTP headers and content security policies

## Migration Steps
1. Review all v4.0 requirements against current controls
2. Identify gaps, especially future-dated requirements
3. Build remediation plan with timeline
4. Update policies and procedures
5. Train assessors and staff on v4.0 changes
6. Validate with QSA or ISA',
 'Guide to transitioning from PCI DSS v3.2.1 to v4.0 with key changes highlighted.',
 '{PCI DSS}', '{}',
 '{pci-dss,payment-security,v4,transition,compliance}',
 'advanced', 12, true, true),

-- ============================================================
-- BEST PRACTICES
-- ============================================================

-- 12. Building a Risk Matrix
(NULL, 'best_practice',
 'Building a 5x5 Risk Matrix: Industry Best Practices',
 'building-5x5-risk-matrix',
 '## The 5×5 Risk Matrix

A risk matrix (or heat map) is a visual tool for assessing and prioritising risks based on their likelihood and impact.

## Defining Likelihood Levels
1. **Rare** (1): Less than once in 5 years
2. **Unlikely** (2): Once in 2-5 years
3. **Possible** (3): Once per year
4. **Likely** (4): Multiple times per year
5. **Almost Certain** (5): Expected to occur frequently

## Defining Impact Levels
1. **Insignificant** (1): No measurable impact
2. **Minor** (2): Localised impact, easily recoverable
3. **Moderate** (3): Significant impact requiring management attention
4. **Major** (4): Serious impact affecting operations
5. **Catastrophic** (5): Existential threat to the organisation

## Risk Rating Calculation
Risk Score = Likelihood × Impact

| Rating  | Score Range | Action Required |
|---------|-------------|-----------------|
| Critical | 20-25      | Immediate action required |
| High     | 12-19      | Management attention needed |
| Medium   | 6-11       | Monitor and plan treatment |
| Low      | 1-5        | Accept or monitor |

## Best Practices
- Calibrate with real examples from your industry
- Review and adjust scales annually
- Ensure consistent application across departments
- Use both inherent and residual risk ratings
- Link to your risk appetite statement',
 'Guide to building and calibrating a 5×5 risk matrix for risk assessment.',
 '{ISO 27001,ISO 31000}', '{}',
 '{risk-matrix,risk-assessment,heat-map,likelihood,impact}',
 'beginner', 8, true, true),

-- 13. Evidence Collection Best Practices
(NULL, 'best_practice',
 'Evidence Collection Best Practices for Compliance Audits',
 'evidence-collection-best-practices',
 '## Why Evidence Matters

Compliance is only as strong as the evidence that supports it. Auditors rely on documented proof that controls are implemented, operating, and effective.

## Types of Evidence
1. **Documentary:** Policies, procedures, standards
2. **System-generated:** Logs, screenshots, configurations
3. **Observational:** Walkthrough notes, meeting minutes
4. **Testimonial:** Interview records, attestations
5. **Analytical:** Trend reports, metrics, dashboards

## Collection Principles
- **Timely:** Collect evidence close to the event
- **Complete:** Cover the entire audit period
- **Authentic:** From authoritative sources
- **Tamper-proof:** Use version control and timestamps
- **Organised:** Map to specific control requirements

## Automation Strategies
- Schedule evidence collection via ComplianceForge monitoring
- API integrations to pull evidence from source systems
- Automated screenshot capture for configurations
- Periodic evidence freshness checks
- Alert when evidence is approaching expiry

## Common Audit Failures
- Evidence gaps: missing periods or controls
- Stale evidence: documents not updated since last audit
- Inconsistent evidence: different versions in different locations
- Excessive evidence: burying auditors in irrelevant documents',
 'Best practices for collecting, organising, and automating compliance audit evidence.',
 '{ISO 27001,SOC 2}', '{}',
 '{evidence,audit,compliance,documentation,automation}',
 'beginner', 8, true, true),

-- 14. Vendor Risk Assessment Guide
(NULL, 'best_practice',
 'Vendor Risk Assessment: A DPO''s Guide to GDPR Article 28',
 'vendor-risk-assessment-dpo-guide',
 '## GDPR Article 28 Requirements

When engaging processors, controllers must use only processors providing sufficient guarantees of appropriate technical and organisational measures.

## Vendor Assessment Framework

### 1. Pre-engagement Due Diligence
- Security certifications (ISO 27001, SOC 2)
- Privacy policies and data handling practices
- Breach history and incident response capabilities
- Sub-processor management
- Data location and transfer mechanisms

### 2. Contractual Requirements
- Processing only on documented instructions
- Confidentiality obligations
- Security measures (Article 32)
- Sub-processor approval process
- Assist with data subject requests
- Delete/return data at end of contract
- Audit rights

### 3. Ongoing Monitoring
- Annual reassessment of high-risk vendors
- Continuous monitoring for critical vendors
- Review of SOC 2 reports or ISO certificates
- Track sub-processor changes
- Monitor breach notifications

### 4. Risk Tiering
- **Critical:** Access to large volumes of personal data
- **High:** Processing sensitive data categories
- **Medium:** Limited personal data access
- **Low:** No personal data processing

## Red Flags
- No security certification
- Reluctance to sign a DPA
- No incident response plan
- Data processing in jurisdictions without adequacy decisions',
 'DPO guide to conducting vendor risk assessments under GDPR Article 28.',
 '{GDPR}', '{}',
 '{vendor-management,tprm,gdpr,article-28,dpo,data-processing}',
 'intermediate', 10, true, true),

-- 15. Security Awareness Training
(NULL, 'best_practice',
 'Security Awareness Training Programme Design',
 'security-awareness-training-programme',
 '## Building an Effective Programme

Security awareness training is a control requirement in ISO 27001 (A.6.3) and mandated by NIS2 for management bodies.

## Programme Components

### 1. Onboarding Training
- Security policy overview
- Acceptable use policy
- Password and authentication
- Data classification and handling
- Incident reporting procedures
- Physical security rules

### 2. Ongoing Training
- Monthly micro-learning modules (5-10 minutes)
- Quarterly phishing simulations
- Annual refresher course
- Role-specific training (developers, admins, managers)

### 3. Targeted Campaigns
- Phishing awareness (email, SMS, voice)
- Social engineering defence
- Remote working security
- Removable media risks
- Travel security

## Measuring Effectiveness
- Phishing simulation click rates (target <5%)
- Training completion rates (target >95%)
- Incident reporting rates (should increase)
- Quiz scores and knowledge retention
- Behaviour change metrics

## Tips for Engagement
- Keep content short and relevant
- Use real-world examples from your industry
- Gamify with leaderboards and rewards
- Vary delivery methods (video, interactive, text)
- Get management buy-in and visible participation',
 'Guide to designing an effective security awareness training programme.',
 '{ISO 27001,NIS2}', '{A.6.3}',
 '{training,awareness,phishing,social-engineering,culture}',
 'beginner', 10, true, true),

-- 16. Business Continuity Planning
(NULL, 'best_practice',
 'Business Continuity Planning: From BIA to DR Testing',
 'business-continuity-planning-guide',
 '## Business Continuity Overview

Business continuity planning ensures your organisation can maintain critical operations during and after a disruption.

## Step 1: Business Impact Analysis (BIA)
- Identify critical business processes
- Determine maximum tolerable downtime (MTD)
- Define recovery time objectives (RTO)
- Define recovery point objectives (RPO)
- Assess financial and operational impact

## Step 2: Risk Assessment
- Identify threats to critical processes
- Assess likelihood and impact
- Consider single points of failure
- Evaluate supply chain dependencies

## Step 3: Strategy Development
- Backup and recovery procedures
- Alternative work locations
- Communication plans
- Third-party agreements
- Insurance coverage

## Step 4: Plan Documentation
- Roles and responsibilities
- Activation criteria and procedures
- Recovery procedures per system
- Communication tree
- Vendor contact information

## Step 5: Testing and Exercises
- **Walkthrough:** Annual review of procedures
- **Tabletop:** Scenario-based discussion exercise
- **Simulation:** Partial execution of the plan
- **Full exercise:** Complete plan activation

## Maintenance
- Review plans after every incident
- Update annually at minimum
- Incorporate lessons learned
- Track changes to critical systems',
 'End-to-end guide to business continuity planning from BIA through DR testing.',
 '{ISO 27001,ISO 22301,NIS2}', '{A.5.29,A.5.30}',
 '{business-continuity,bcp,disaster-recovery,bia,rto,rpo}',
 'intermediate', 12, true, true),

-- 17. Data Classification Guide
(NULL, 'best_practice',
 'Data Classification Scheme: Design and Implementation',
 'data-classification-scheme-guide',
 '## Why Classify Data?

Data classification ensures that information receives an appropriate level of protection according to its sensitivity and value.

## Classification Levels

### Public
- Marketing materials, public website content
- No restrictions on access or distribution

### Internal
- Internal memos, general business documents
- Restricted to employees and authorised contractors

### Confidential
- Financial data, employee records, customer data
- Access on need-to-know basis
- Encryption required in transit and at rest

### Restricted
- Trade secrets, PII of sensitive nature, health data
- Strict access controls, encryption mandatory
- Full audit trail of access

## Implementation Steps
1. Define classification levels with examples
2. Assign data owners to all data sets
3. Train staff on classification procedures
4. Label documents and systems
5. Apply technical controls per level
6. Review and reclassify periodically

## Handling Requirements per Level

| Requirement | Public | Internal | Confidential | Restricted |
|-------------|--------|----------|-------------|------------|
| Encryption at rest | No | Optional | Required | Required |
| Encryption in transit | No | Optional | Required | Required |
| Access logging | No | Optional | Required | Required |
| DLP monitoring | No | No | Recommended | Required |
| Backup encryption | No | No | Required | Required |',
 'Guide to designing and implementing a data classification scheme.',
 '{ISO 27001,GDPR}', '{A.5.12,A.5.13}',
 '{data-classification,labelling,handling,sensitivity}',
 'beginner', 8, true, true),

-- 18. Supply Chain Security
(NULL, 'best_practice',
 'Supply Chain Security: Managing Third-Party Cyber Risk',
 'supply-chain-security-guide',
 '## Supply Chain Risk Landscape

Supply chain attacks have increased dramatically. NIS2 specifically requires supply chain security measures.

## Key Principles
1. Know your suppliers and their access
2. Assess risk proportionally
3. Include security in contracts
4. Monitor continuously
5. Plan for supplier failure

## Assessment Framework
- Pre-contract security questionnaire
- Review security certifications
- Penetration test results
- Business continuity plans
- Incident response capabilities
- Data handling practices

## Contractual Controls
- Right to audit
- Breach notification timelines
- Sub-contractor approval
- Data handling and return/deletion
- SLA for security patches
- Insurance requirements

## Monitoring
- Continuous security rating services
- Annual reassessment
- Track CVE disclosures
- Monitor dark web for supplier breaches
- Review access logs for supplier accounts',
 'Best practices for managing supply chain cybersecurity risk.',
 '{ISO 27001,NIS2}', '{A.5.19,A.5.20,A.5.21,A.5.22,A.5.23}',
 '{supply-chain,third-party,vendor,nis2,procurement}',
 'intermediate', 10, true, true),

-- 19. Change Management
(NULL, 'best_practice',
 'Change Management for Information Security',
 'change-management-information-security',
 '## Change Management in ISMS Context

All changes to information processing facilities and systems must be controlled to maintain security.

## Change Types
- **Standard:** Pre-approved, low-risk (e.g., user provisioning)
- **Normal:** Requires assessment and approval
- **Emergency:** Expedited process for critical fixes

## Change Process
1. Request and record the change
2. Assess security impact
3. Test in non-production environment
4. Obtain approval from CAB
5. Implement with rollback plan
6. Verify and close

## Security Impact Assessment
- Does it change the attack surface?
- Does it affect access controls?
- Does it modify data flows?
- Does it impact compliance requirements?
- Does it require updated documentation?

## ISO 27001 Controls
- A.8.32: Change management
- A.8.9: Configuration management
- A.8.25: Secure development lifecycle
- A.8.31: Separation of environments',
 'Guide to implementing change management for information security.',
 '{ISO 27001}', '{A.8.32,A.8.9}',
 '{change-management,cab,security-impact,configuration}',
 'intermediate', 8, true, true),

-- 20. DPIA Guide
(NULL, 'regulatory_guide',
 'Conducting a Data Protection Impact Assessment (DPIA)',
 'data-protection-impact-assessment-guide',
 '## When Is a DPIA Required?

GDPR Article 35 requires a DPIA when processing is likely to result in a high risk to individuals.

## Mandatory DPIA Triggers
- Systematic and extensive profiling with significant effects
- Large-scale processing of special categories
- Systematic monitoring of publicly accessible areas
- New technologies
- Processing that prevents individuals from exercising rights

## DPIA Process

### Step 1: Describe the Processing
- Nature, scope, context, and purpose
- Data flows and data recipients
- Technology used
- Retention periods

### Step 2: Assess Necessity and Proportionality
- Is the processing necessary for the purpose?
- Could a less intrusive method achieve the same goal?
- What is the lawful basis?

### Step 3: Identify and Assess Risks
- Risks to individuals (not the organisation)
- Consider: loss of control, discrimination, identity theft, financial loss
- Rate likelihood and severity

### Step 4: Identify Mitigation Measures
- Technical measures (encryption, access controls, pseudonymisation)
- Organisational measures (policies, training, audits)
- Demonstrate how risks are reduced to acceptable levels

### Step 5: Document and Review
- Record the DPIA with all assessments
- Consult the DPO
- If high residual risk remains: consult the supervisory authority
- Review when processing changes',
 'Step-by-step guide to conducting GDPR Data Protection Impact Assessments.',
 '{GDPR}', '{}',
 '{dpia,gdpr,privacy,impact-assessment,article-35}',
 'intermediate', 10, true, true),

-- 21. Risk Treatment Guide
(NULL, 'best_practice',
 'Risk Treatment Options: Accept, Mitigate, Transfer, or Avoid',
 'risk-treatment-options-guide',
 '## Risk Treatment Overview

After risks are identified and assessed, each must be treated. ISO 27005 defines four treatment options.

## Treatment Options

### 1. Risk Mitigation (Reduce)
- Implement controls to reduce likelihood or impact
- Most common treatment option
- Example: Installing a WAF to reduce web application attack risk

### 2. Risk Acceptance
- Formally accept the risk without additional controls
- When cost of treatment exceeds potential impact
- Requires management approval and documentation
- Must be within the organisation risk appetite

### 3. Risk Transfer (Share)
- Transfer risk to a third party
- Cyber insurance
- Outsourcing to managed security providers
- Contractual risk allocation
- Note: you can transfer financial risk, not accountability

### 4. Risk Avoidance
- Eliminate the risk by removing the activity
- Example: Not collecting unnecessary personal data
- Most effective but may limit business opportunities

## Treatment Plan
Each treatment decision must include:
- Risk reference and description
- Chosen treatment option with justification
- Specific controls or actions
- Owner responsible for implementation
- Timeline and milestones
- Residual risk assessment
- Acceptance by risk owner',
 'Guide to selecting and implementing risk treatment options per ISO 27005.',
 '{ISO 27001,ISO 31000}', '{}',
 '{risk-treatment,mitigation,acceptance,transfer,avoidance}',
 'beginner', 8, true, true),

-- 22. Audit Preparation
(NULL, 'audit_preparation',
 'Preparing for Your ISO 27001 Certification Audit',
 'iso-27001-certification-audit-preparation',
 '## Audit Preparation Timeline

### 3 Months Before
- Conduct internal audit of all ISMS processes
- Review and close non-conformities from previous audits
- Verify all documentation is current
- Ensure management review has been conducted

### 1 Month Before
- Complete evidence collection for all Annex A controls
- Verify all policies are approved and communicated
- Check training records are complete
- Test incident response procedures

### 1 Week Before
- Prepare audit schedule with auditor
- Brief all interviewees
- Organise evidence folders
- Test projector and meeting rooms

## Common Audit Findings
- Incomplete risk assessment or treatment plan
- Missing evidence for control effectiveness
- Outdated or unapproved policies
- Incomplete management review minutes
- Training records not covering all personnel
- Lack of measurable security objectives

## Interview Tips
- Be honest — auditors spot rehearsed answers
- Show evidence, don''t just describe processes
- If unsure, say so and offer to follow up
- Keep answers focused and relevant',
 'Comprehensive preparation guide for ISO 27001 certification audits.',
 '{ISO 27001}', '{}',
 '{audit,certification,preparation,iso-27001,internal-audit}',
 'intermediate', 10, true, true),

-- 23. Privacy by Design
(NULL, 'regulatory_guide',
 'Privacy by Design: Embedding GDPR into System Development',
 'privacy-by-design-gdpr-guide',
 '## GDPR Article 25: Data Protection by Design and by Default

Controllers must implement appropriate technical and organisational measures to ensure data protection principles are embedded into processing.

## Seven Foundational Principles
1. Proactive not reactive — preventative not remedial
2. Privacy as the default setting
3. Privacy embedded into design
4. Full functionality — positive-sum, not zero-sum
5. End-to-end security — full lifecycle protection
6. Visibility and transparency — keep it open
7. Respect for user privacy — keep it user-centric

## Implementation Checklist
- [ ] Privacy requirements in project initiation
- [ ] Data minimisation review at design stage
- [ ] Purpose limitation built into data models
- [ ] Consent mechanisms designed into UX
- [ ] Data subject rights supported by architecture
- [ ] Retention policies automated
- [ ] Access controls reflecting need-to-know
- [ ] Encryption by default for personal data
- [ ] Pseudonymisation where feasible
- [ ] DPIA conducted before go-live',
 'Guide to implementing Privacy by Design per GDPR Article 25.',
 '{GDPR}', '{}',
 '{privacy-by-design,gdpr,article-25,development,data-minimisation}',
 'intermediate', 8, true, true),

-- 24. Cloud Security
(NULL, 'best_practice',
 'Cloud Security Checklist for GRC Teams',
 'cloud-security-checklist-grc',
 '## Cloud Security Fundamentals

Moving to the cloud changes the security model. GRC teams must understand the shared responsibility model.

## Shared Responsibility
- **IaaS:** Customer manages OS, apps, data, access
- **PaaS:** Customer manages apps, data, access
- **SaaS:** Customer manages data and access

## Security Checklist
- [ ] Cloud provider due diligence (SOC 2, ISO 27001)
- [ ] Data residency requirements met (EU data in EU)
- [ ] Encryption at rest and in transit
- [ ] Identity and access management configured
- [ ] MFA enabled for all administrative access
- [ ] Logging and monitoring active
- [ ] Network segmentation and security groups
- [ ] Backup and recovery tested
- [ ] Incident response for cloud events
- [ ] Regular configuration reviews

## Compliance Considerations
- GDPR: data processing agreements, transfer mechanisms
- NIS2: cloud as critical supply chain
- ISO 27017: cloud-specific security controls
- ISO 27018: PII in public cloud

## Tools
- Cloud Security Posture Management (CSPM)
- Cloud Access Security Broker (CASB)
- Cloud Workload Protection Platform (CWPP)',
 'Security checklist for GRC teams managing cloud environments.',
 '{ISO 27001,ISO 27017}', '{}',
 '{cloud,security,shared-responsibility,iaas,paas,saas}',
 'intermediate', 8, true, true),

-- 25. Board-Level Reporting
(NULL, 'best_practice',
 'Cybersecurity Board Reporting: What Directors Need to Know',
 'cybersecurity-board-reporting-guide',
 '## Why Board Reporting Matters

NIS2 makes management personally liable for cybersecurity. Directors need clear, actionable reporting.

## Key Metrics for Boards
1. **Risk Posture:** Top 10 risks with current ratings
2. **Compliance Status:** Framework compliance percentages
3. **Incident Trends:** Number and severity over time
4. **Vulnerability Status:** Open critical/high vulnerabilities
5. **Audit Findings:** Open vs closed findings
6. **Training Compliance:** Completion rates
7. **Vendor Risk:** Critical vendor risk levels
8. **Investment ROI:** Security spend vs risk reduction

## Reporting Principles
- Use business language, not technical jargon
- Focus on trends, not absolute numbers
- Highlight decisions needed from the board
- Show risk in financial terms where possible
- Keep reports to 3-5 pages maximum
- Use traffic light (RAG) status indicators

## Frequency
- Full cybersecurity report: quarterly
- Incident briefings: as needed
- Annual strategy review and budget request',
 'Guide to creating effective cybersecurity reports for board directors.',
 '{NIS2,ISO 27001}', '{}',
 '{board-reporting,governance,kpi,metrics,directors}',
 'intermediate', 8, true, true),

-- 26-50: Additional articles for breadth

-- 26
(NULL, 'risk_management',
 'Conducting Effective Risk Assessments: A Step-by-Step Guide',
 'risk-assessment-step-by-step',
 '## Risk Assessment Process

### 1. Establish Context
Define scope, criteria, and risk appetite.

### 2. Risk Identification
Use brainstorming, interviews, historical data, and threat intelligence.

### 3. Risk Analysis
Assess likelihood and impact using your risk matrix. Consider both inherent and residual risk.

### 4. Risk Evaluation
Compare against risk appetite. Prioritise for treatment.

### 5. Risk Treatment
Select and implement controls. Document treatment plans.

### 6. Monitoring and Review
Continuously monitor risk indicators and review assessments regularly.',
 'Step-by-step guide to conducting comprehensive risk assessments.',
 '{ISO 27001,ISO 31000}', '{}',
 '{risk-assessment,methodology,identification,analysis}',
 'beginner', 6, true, true),

-- 27
(NULL, 'implementation_guide',
 'Physical Security Controls: Implementing ISO 27001 Annex A.7',
 'physical-security-controls-a7',
 '## Physical Security Overview

Annex A.7 covers physical controls to protect premises, equipment, and information from physical threats.

### A.7.1 — Physical Security Perimeters
Define secure areas with appropriate barriers.

### A.7.2 — Physical Entry
Control access to secure areas using badges, biometrics, or locks.

### A.7.3 — Securing Offices
Protect offices, rooms, and facilities against unauthorised access.

### A.7.4 — Physical Security Monitoring
Monitor premises continuously with CCTV and alarms.

## Implementation Checklist
- Perimeter security (fencing, walls, gates)
- Access control systems (cards, biometrics)
- CCTV with 30-day retention
- Visitor management process
- Clean desk and clear screen policy
- Equipment disposal procedures',
 'Guide to implementing physical security controls under ISO 27001 Annex A.7.',
 '{ISO 27001}', '{A.7.1,A.7.2,A.7.3,A.7.4}',
 '{physical-security,access-control,cctv,perimeter}',
 'beginner', 6, true, true),

-- 28
(NULL, 'implementation_guide',
 'Secure Software Development Lifecycle (SSDLC)',
 'secure-software-development-lifecycle',
 '## SSDLC Overview

Integrate security into every phase of the software development lifecycle.

## Phases

### Requirements
- Security requirements alongside functional requirements
- Threat modelling early in design

### Design
- Secure architecture patterns
- Security design review

### Implementation
- Secure coding standards (OWASP)
- Code review with security focus
- Static Application Security Testing (SAST)

### Testing
- Dynamic Application Security Testing (DAST)
- Penetration testing
- Security regression testing

### Deployment
- Security configuration review
- Container image scanning
- Infrastructure as Code security

### Operations
- Runtime Application Self-Protection (RASP)
- Vulnerability management
- Incident response integration',
 'Guide to implementing a Secure Software Development Lifecycle.',
 '{ISO 27001}', '{A.8.25,A.8.26,A.8.27,A.8.28}',
 '{ssdlc,secure-development,owasp,sast,dast}',
 'advanced', 10, true, true),

-- 29
(NULL, 'regulatory_guide',
 'Cross-Border Data Transfers: Navigating Post-Schrems II',
 'cross-border-data-transfers-schrems-ii',
 '## Data Transfer Mechanisms

### Adequacy Decisions
EU Commission has determined certain countries provide adequate protection. Check the current list.

### Standard Contractual Clauses (SCCs)
- Updated SCCs adopted June 2021
- Must conduct Transfer Impact Assessment (TIA)
- Supplementary measures may be needed

### Binding Corporate Rules (BCRs)
- For intra-group transfers
- Requires DPA approval
- Comprehensive privacy programme required

### Transfer Impact Assessment
1. Map data transfers
2. Assess third country laws
3. Identify supplementary measures
4. Document assessment
5. Review regularly',
 'Guide to managing cross-border data transfers under GDPR post-Schrems II.',
 '{GDPR}', '{}',
 '{data-transfers,schrems-ii,sccs,adequacy,tia}',
 'advanced', 8, true, true),

-- 30
(NULL, 'best_practice',
 'Information Security Metrics and KPIs',
 'information-security-metrics-kpis',
 '## Choosing the Right Metrics

### Operational Metrics
- Patch compliance rate
- Vulnerability remediation time
- Incident detection time (MTTD)
- Incident response time (MTTR)
- Phishing click rate

### Compliance Metrics
- Framework compliance percentage
- Control effectiveness score
- Audit finding closure rate
- Policy review completion rate
- Training completion rate

### Risk Metrics
- Risk appetite utilisation
- Number of high/critical risks
- Risk treatment plan completion
- Overdue risk reviews
- KRI breach frequency

### Reporting Tips
- Choose 10-15 key metrics
- Set baselines and targets
- Show trends over time
- Use RAG status indicators
- Automate where possible',
 'Guide to selecting and reporting information security metrics and KPIs.',
 '{ISO 27001}', '{}',
 '{metrics,kpis,measurement,reporting,dashboard}',
 'intermediate', 8, true, true),

-- 31
(NULL, 'best_practice',
 'Asset Management for Information Security',
 'asset-management-information-security',
 '## Asset Management Overview

Maintaining an accurate inventory of information assets is fundamental to security management.

## Asset Categories
- Hardware (servers, workstations, mobile devices)
- Software (applications, operating systems)
- Data (databases, file shares, cloud storage)
- Services (cloud services, third-party APIs)
- People (roles with security responsibilities)

## Asset Lifecycle
1. **Procurement:** Security requirements in procurement
2. **Deployment:** Baseline configuration, asset tagging
3. **Operation:** Monitoring, patching, access control
4. **Maintenance:** Regular review and updates
5. **Disposal:** Secure data destruction, decommissioning',
 'Guide to information asset management and lifecycle.',
 '{ISO 27001}', '{A.5.9,A.5.10,A.5.11}',
 '{asset-management,inventory,lifecycle,classification}',
 'beginner', 6, true, true),

-- 32
(NULL, 'implementation_guide',
 'Network Security Controls Implementation Guide',
 'network-security-controls-guide',
 '## Network Security Overview

Protect network infrastructure and data in transit with layered security controls.

## Key Controls
- Network segmentation (VLANs, micro-segmentation)
- Firewall rules and management
- Intrusion Detection/Prevention Systems (IDS/IPS)
- Web Application Firewall (WAF)
- DNS security
- VPN for remote access
- Network access control (802.1X)
- Wireless security (WPA3)

## Monitoring
- NetFlow analysis
- Full packet capture for investigation
- DNS query logging
- SSL/TLS inspection
- Anomaly detection',
 'Guide to implementing network security controls.',
 '{ISO 27001,Cyber Essentials}', '{A.8.20,A.8.21,A.8.22,A.8.23}',
 '{network-security,firewall,segmentation,ids,monitoring}',
 'advanced', 8, true, true),

-- 33
(NULL, 'regulatory_guide',
 'GDPR Data Subject Rights: Implementation Guide',
 'gdpr-data-subject-rights-guide',
 '## Data Subject Rights Under GDPR

### Right of Access (Art. 15)
Provide a copy of personal data within one month.

### Right to Rectification (Art. 16)
Correct inaccurate personal data without undue delay.

### Right to Erasure (Art. 17)
Delete personal data when no longer necessary.

### Right to Restriction (Art. 18)
Restrict processing in specific circumstances.

### Right to Data Portability (Art. 20)
Provide data in machine-readable format.

### Right to Object (Art. 21)
Object to processing based on legitimate interests.

## Implementation
- Establish a DSR intake process
- Verify identity before fulfilling requests
- Track deadlines and escalations
- Document all decisions and actions
- Train frontline staff on recognising DSRs',
 'Implementation guide for handling GDPR data subject rights requests.',
 '{GDPR}', '{}',
 '{gdpr,dsr,data-subject-rights,access,erasure,portability}',
 'intermediate', 8, true, true),

-- 34
(NULL, 'best_practice',
 'Multi-Factor Authentication: Deployment Best Practices',
 'mfa-deployment-best-practices',
 '## MFA Overview

Multi-factor authentication significantly reduces the risk of account compromise.

## MFA Methods (Strongest to Weakest)
1. FIDO2/WebAuthn hardware keys
2. Platform authenticators (Windows Hello, Touch ID)
3. Authenticator apps (TOTP)
4. Push notifications
5. SMS OTP (last resort only)

## Deployment Strategy
1. Start with privileged accounts
2. Extend to all cloud services
3. Roll out to VPN and remote access
4. Cover all internal applications
5. Enforce for all users

## Considerations
- Backup methods for lost devices
- Helpdesk verification procedures
- Accessibility for users with disabilities
- Travel and roaming considerations
- Phishing-resistant methods preferred',
 'Best practices for deploying multi-factor authentication across the enterprise.',
 '{ISO 27001,Cyber Essentials,NIS2}', '{A.8.5}',
 '{mfa,authentication,fido2,totp,security}',
 'beginner', 6, true, true),

-- 35
(NULL, 'best_practice',
 'Internal Audit Programme for ISMS',
 'internal-audit-programme-isms',
 '## Internal Audit Overview

ISO 27001 clause 9.2 requires organisations to conduct internal audits at planned intervals.

## Planning
- Define a 3-year audit cycle covering all clauses and controls
- Risk-based prioritisation (audit high-risk areas more frequently)
- Ensure auditor independence
- Allocate adequate time and resources

## Execution
1. Prepare audit plan and checklist
2. Review documentation
3. Conduct interviews
4. Observe processes
5. Sample evidence
6. Document findings

## Reporting
- Non-conformities: major and minor
- Observations and opportunities for improvement
- Positive findings
- Audit conclusions and recommendations

## Follow-Up
- Root cause analysis for non-conformities
- Corrective action plans with deadlines
- Verification of effectiveness
- Update risk register if needed',
 'Guide to planning and executing ISMS internal audits per ISO 27001 clause 9.2.',
 '{ISO 27001}', '{}',
 '{internal-audit,isms,iso-27001,non-conformity,corrective-action}',
 'intermediate', 8, true, true),

-- 36
(NULL, 'implementation_guide',
 'Backup and Recovery Strategy: ISO 27001 A.8.13',
 'backup-recovery-strategy-a813',
 '## Backup Strategy

### 3-2-1 Rule
- 3 copies of data
- 2 different storage media
- 1 offsite copy

### Backup Types
- Full: complete copy (weekly)
- Incremental: changes since last backup (daily)
- Differential: changes since last full (as needed)

### Testing
- Monthly restore tests
- Annual full DR test
- Document and review results

### Retention
- Align with regulatory requirements
- Minimum 30 days for operational recovery
- 7 years for financial records
- Consider GDPR retention limits',
 'Implementation guide for backup and recovery per ISO 27001 A.8.13.',
 '{ISO 27001}', '{A.8.13,A.8.14}',
 '{backup,recovery,disaster-recovery,3-2-1,retention}',
 'beginner', 6, true, true),

-- 37
(NULL, 'regulatory_guide',
 'ePrivacy and Cookie Compliance Guide',
 'eprivacy-cookie-compliance-guide',
 '## Cookie Compliance

### Consent Requirements
- Prior informed consent for non-essential cookies
- Clear and plain language
- Granular choices (not just accept/reject)
- Easy withdrawal of consent
- Cookie wall prohibition (in most jurisdictions)

### Cookie Categories
1. **Strictly Necessary:** No consent needed
2. **Functionality:** Consent required
3. **Analytics:** Consent required
4. **Marketing:** Consent required

### Implementation
- Cookie management platform (CMP)
- Scan website regularly for new cookies
- Maintain cookie register
- Record consent choices
- Honour Do Not Track where applicable',
 'Guide to cookie and ePrivacy compliance for European websites.',
 '{GDPR,ePrivacy}', '{}',
 '{cookies,eprivacy,consent,cmp,tracking}',
 'beginner', 6, true, true),

-- 38
(NULL, 'best_practice',
 'Zero Trust Architecture: Implementation Guide',
 'zero-trust-architecture-guide',
 '## Zero Trust Principles

Never trust, always verify. Every access request is fully authenticated, authorised, and encrypted.

## Core Components
1. Identity verification (strong authentication)
2. Device compliance checking
3. Least privilege access
4. Micro-segmentation
5. Continuous monitoring and validation

## Implementation Phases
1. Define protect surfaces (critical data and assets)
2. Map transaction flows
3. Build Zero Trust architecture
4. Create Zero Trust policies
5. Monitor and maintain

## Technologies
- Identity Provider (IdP) with MFA
- Endpoint Detection and Response (EDR)
- Software-Defined Perimeter (SDP)
- Cloud Access Security Broker (CASB)
- Security Information and Event Management (SIEM)',
 'Guide to implementing Zero Trust Architecture principles.',
 '{ISO 27001,NIS2}', '{}',
 '{zero-trust,architecture,micro-segmentation,identity,security}',
 'advanced', 8, true, true),

-- 39
(NULL, 'vendor_management',
 'SLA Management for Critical Vendors',
 'sla-management-critical-vendors',
 '## SLA Framework

### Key SLA Categories
- Availability (uptime commitments)
- Performance (response times, throughput)
- Security (incident response times, patching SLAs)
- Compliance (audit reports, certifications)
- Support (response and resolution times)

### Monitoring
- Automated SLA tracking dashboards
- Monthly performance reviews
- Quarterly business reviews for critical vendors
- Annual comprehensive assessment

### Remediation
- Escalation procedures for SLA breaches
- Service credits and penalties
- Improvement plans for persistent failures
- Contract termination criteria',
 'Guide to establishing and monitoring SLAs for critical vendors.',
 '{ISO 27001}', '{A.5.19,A.5.20}',
 '{sla,vendor-management,monitoring,availability,performance}',
 'intermediate', 6, true, true),

-- 40
(NULL, 'best_practice',
 'Secure Remote Working: Policy and Controls',
 'secure-remote-working-policy-controls',
 '## Remote Working Security

### Policy Elements
- Approved devices and software
- Network security requirements (VPN)
- Data handling and classification
- Physical security of workspace
- Incident reporting procedures

### Technical Controls
- VPN with MFA
- Endpoint protection (EDR/MDM)
- Full disk encryption
- Screen lock policies
- DLP for sensitive data
- Secure file sharing

### User Responsibilities
- Use approved devices only
- Lock screens when away
- Secure physical workspace
- Report incidents immediately
- Follow clean desk policy at home',
 'Guide to implementing secure remote working policies and controls.',
 '{ISO 27001}', '{A.6.7}',
 '{remote-working,teleworking,vpn,byod,endpoint-security}',
 'beginner', 6, true, true),

-- 41-50: Additional topical articles

-- 41
(NULL, 'implementation_guide',
 'Email Security: SPF, DKIM, and DMARC Configuration',
 'email-security-spf-dkim-dmarc',
 '## Email Security Protocols

### SPF (Sender Policy Framework)
Specifies which mail servers can send on behalf of your domain.

### DKIM (DomainKeys Identified Mail)
Adds a digital signature to verify message integrity.

### DMARC (Domain-based Message Authentication)
Builds on SPF and DKIM with policy enforcement and reporting.

## Implementation Order
1. Implement SPF with soft fail (~all)
2. Configure DKIM signing
3. Deploy DMARC in monitor mode (p=none)
4. Review DMARC reports for 4-6 weeks
5. Tighten to quarantine (p=quarantine)
6. Move to reject (p=reject) when confident',
 'Guide to configuring SPF, DKIM, and DMARC for email security.',
 '{Cyber Essentials,ISO 27001}', '{}',
 '{email-security,spf,dkim,dmarc,phishing-prevention}',
 'intermediate', 6, true, true),

-- 42
(NULL, 'regulatory_guide',
 'SOC 2 Type II: What It Means for Your Vendors',
 'soc-2-type-ii-vendor-guide',
 '## Understanding SOC 2 Reports

### Type I vs Type II
- Type I: Design of controls at a point in time
- Type II: Operating effectiveness over a period (usually 12 months)

### Trust Service Criteria
1. Security (Common Criteria — required)
2. Availability
3. Processing Integrity
4. Confidentiality
5. Privacy

### Reading a SOC 2 Report
- Management assertion
- Auditor opinion
- System description
- Control descriptions and test results
- Exceptions and management responses

### Due Diligence Questions
- Is the report current (within 12 months)?
- Are there any qualified opinions?
- Do exceptions affect your use case?
- Are complementary user entity controls addressed?',
 'Guide to understanding SOC 2 Type II reports for vendor due diligence.',
 '{SOC 2}', '{}',
 '{soc-2,vendor-assessment,trust-services,audit-report}',
 'intermediate', 6, true, true),

-- 43
(NULL, 'best_practice',
 'Secure Disposal and Media Sanitisation',
 'secure-disposal-media-sanitisation',
 '## Data Disposal Requirements

### Methods
- Degaussing (magnetic media)
- Overwriting (NIST 800-88 guidelines)
- Physical destruction (shredding, crushing)
- Cryptographic erasure (encrypted media)

### By Media Type
- HDDs: Degauss + physical destruction
- SSDs: Cryptographic erasure + physical destruction
- Paper: Cross-cut shredding (DIN 66399 Level P-4 minimum)
- Optical media: Physical destruction

### Documentation
- Certificate of destruction from provider
- Asset disposal register
- Chain of custody records
- Verify compliance with data retention policies',
 'Guide to secure disposal and media sanitisation procedures.',
 '{ISO 27001,GDPR}', '{A.7.14,A.8.10}',
 '{disposal,sanitisation,media,destruction,retention}',
 'beginner', 5, true, true),

-- 44
(NULL, 'implementation_guide',
 'Identity and Access Management (IAM) Architecture',
 'iam-architecture-guide',
 '## IAM Architecture Overview

### Components
- Identity Provider (IdP)
- Directory services
- Single Sign-On (SSO)
- Multi-Factor Authentication (MFA)
- Privileged Access Management (PAM)
- Identity Governance and Administration (IGA)

### Lifecycle Management
1. Joiner: Provision accounts and access
2. Mover: Adjust access for role changes
3. Leaver: Deprovision all access promptly

### Access Reviews
- Quarterly for privileged accounts
- Semi-annually for all users
- Automated certification campaigns
- Manager and application owner attestation',
 'Guide to designing Identity and Access Management architecture.',
 '{ISO 27001}', '{A.5.15,A.5.16,A.5.17,A.5.18}',
 '{iam,identity,access-management,sso,pam,lifecycle}',
 'advanced', 8, true, true),

-- 45
(NULL, 'best_practice',
 'Penetration Testing Programme Guide',
 'penetration-testing-programme-guide',
 '## Penetration Testing Overview

### Types
- Network penetration test (internal and external)
- Web application test
- Mobile application test
- Social engineering test
- Physical penetration test
- Red team exercise

### Frequency
- Annual minimum for all critical systems
- After significant changes
- Before major releases

### Scope Definition
- In-scope systems and networks
- Excluded systems
- Testing hours and limitations
- Emergency contacts

### Reporting
- Executive summary with risk rating
- Detailed findings with evidence
- Remediation recommendations
- Retest to verify fixes',
 'Guide to establishing a penetration testing programme.',
 '{ISO 27001,PCI DSS}', '{A.8.8}',
 '{penetration-testing,pentest,security-assessment,red-team}',
 'intermediate', 6, true, true),

-- 46
(NULL, 'regulatory_guide',
 'GDPR Lawful Basis: Choosing and Documenting',
 'gdpr-lawful-basis-guide',
 '## The Six Lawful Bases

### 1. Consent (Art. 6(1)(a))
Freely given, specific, informed, and unambiguous.

### 2. Contract (Art. 6(1)(b))
Necessary for performance of a contract.

### 3. Legal Obligation (Art. 6(1)(c))
Required by EU or Member State law.

### 4. Vital Interests (Art. 6(1)(d))
Necessary to protect life.

### 5. Public Interest (Art. 6(1)(e))
Necessary for a task in the public interest.

### 6. Legitimate Interests (Art. 6(1)(f))
Necessary for legitimate interests, balanced against individual rights.

## Documentation Requirements
- Record the lawful basis for each processing activity
- Document the assessment (especially for legitimate interests)
- Inform data subjects of the lawful basis
- Review when processing changes',
 'Guide to selecting and documenting GDPR lawful bases for processing.',
 '{GDPR}', '{}',
 '{gdpr,lawful-basis,consent,legitimate-interests,legal-obligation}',
 'beginner', 6, true, true),

-- 47
(NULL, 'best_practice',
 'Configuration Management and Hardening Guide',
 'configuration-management-hardening-guide',
 '## Configuration Management

### Baselines
- Define secure baseline configurations for all system types
- Use CIS Benchmarks or DISA STIGs as starting points
- Document all deviations with justification

### Hardening Checklist
- Remove unnecessary services and ports
- Disable default accounts
- Apply principle of least functionality
- Enable logging and auditing
- Configure host-based firewalls
- Apply latest security patches

### Change Control
- All configuration changes through change management
- Automated configuration scanning
- Drift detection and alerting
- Regular compliance reporting',
 'Guide to configuration management and system hardening.',
 '{ISO 27001,Cyber Essentials}', '{A.8.9}',
 '{configuration,hardening,baseline,cis-benchmark,drift}',
 'intermediate', 6, true, true),

-- 48
(NULL, 'regulatory_guide',
 'EU AI Act: Implications for GRC Teams',
 'eu-ai-act-implications-grc',
 '## EU AI Act Overview

The EU AI Act establishes a risk-based framework for artificial intelligence systems.

### Risk Categories
1. **Unacceptable Risk:** Banned (social scoring, manipulation)
2. **High Risk:** Strict requirements (HR, credit scoring, law enforcement)
3. **Limited Risk:** Transparency obligations
4. **Minimal Risk:** No specific requirements

### Requirements for High-Risk AI
- Risk management system
- Data governance
- Technical documentation
- Record-keeping and logging
- Transparency and user information
- Human oversight
- Accuracy, robustness, and cybersecurity

### GRC Implications
- AI inventory and classification
- Conformity assessments
- Vendor AI due diligence
- Integration with existing risk framework
- Staff training on AI governance',
 'Overview of EU AI Act implications for governance, risk, and compliance teams.',
 '{EU AI Act}', '{}',
 '{ai-act,artificial-intelligence,eu-regulation,high-risk-ai,governance}',
 'advanced', 8, true, true),

-- 49
(NULL, 'best_practice',
 'Security Operations Centre (SOC) Setup Guide',
 'security-operations-centre-setup',
 '## SOC Overview

### SOC Models
- In-house SOC
- Managed SOC (MSSP)
- Hybrid SOC

### Core Capabilities
- 24/7 monitoring and alerting
- Incident detection and triage
- Threat intelligence integration
- Vulnerability management support
- Log management and analysis
- Incident response coordination

### Technology Stack
- SIEM platform
- SOAR (Security Orchestration, Automation, Response)
- Threat intelligence platform
- EDR/XDR solutions
- Network monitoring tools
- Ticketing system

### Staffing
- SOC Manager
- Tier 1 Analysts (monitoring)
- Tier 2 Analysts (investigation)
- Tier 3 Analysts (threat hunting)
- Incident responders',
 'Guide to setting up a Security Operations Centre.',
 '{ISO 27001,NIS2}', '{A.8.15,A.8.16}',
 '{soc,security-operations,monitoring,siem,incident-detection}',
 'advanced', 8, true, true),

-- 50
(NULL, 'best_practice',
 'Compliance Programme Maturity Model',
 'compliance-programme-maturity-model',
 '## Maturity Levels

### Level 1: Initial
- Ad hoc processes
- Reactive compliance
- Limited documentation

### Level 2: Developing
- Basic policies in place
- Some controls implemented
- Manual processes

### Level 3: Defined
- Documented processes
- Regular reviews
- Risk-based approach
- Training programme

### Level 4: Managed
- Metrics and measurement
- Automated monitoring
- Continuous improvement
- Integrated GRC

### Level 5: Optimised
- Predictive analytics
- Fully automated compliance
- Industry leadership
- Innovation in GRC practices

## Assessment
- Evaluate each domain against the maturity model
- Set target maturity levels
- Build roadmap to close gaps
- Review progress quarterly',
 'Maturity model for assessing and improving your compliance programme.',
 '{ISO 27001}', '{}',
 '{maturity-model,compliance,programme,assessment,improvement}',
 'intermediate', 6, true, true)

ON CONFLICT DO NOTHING;
