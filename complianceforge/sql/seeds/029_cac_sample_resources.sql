-- ============================================================
-- ComplianceForge GRC Platform
-- Seed 029: Compliance-as-Code Sample Resource Definitions
-- Sample YAML resource definitions stored as seed data to
-- demonstrate the CaC engine capabilities. These are inserted
-- into the knowledge base as reference examples.
-- ============================================================

-- ============================================================
-- SAMPLE CAC YAML RESOURCE DEFINITIONS
-- Stored as knowledge articles for reference and onboarding.
-- ============================================================

INSERT INTO knowledge_articles (
    organization_id, article_type, title, slug, content_markdown, summary,
    applicable_frameworks, applicable_control_codes, tags,
    difficulty, reading_time_minutes, is_system, is_published
) VALUES

-- 1. ControlImplementation Example
(NULL, 'implementation_guide',
 'Compliance-as-Code: ControlImplementation Resource Reference',
 'cac-control-implementation-reference',
 '## ControlImplementation Resource

A `ControlImplementation` resource defines how a specific compliance control is implemented within your organisation.

### Example YAML

```yaml
apiVersion: complianceforge.io/v1
kind: ControlImplementation
metadata:
  name: access-control-mfa-enforcement
  uid: controlimplementation/access-control-mfa-enforcement
  labels:
    framework: ISO27001
    domain: access-control
    priority: high
  annotations:
    owner: security-team
    last-review: "2026-01-15"
spec:
  control_code: A.9.4.2
  framework: ISO27001
  status: implemented
  implementation_type: technical
  description: |
    Multi-factor authentication is enforced for all user accounts
    via Azure AD Conditional Access policies. Hardware tokens are
    required for privileged accounts.
  evidence:
    - type: screenshot
      path: evidence/aad-conditional-access.png
      description: Azure AD MFA policy configuration
    - type: log_export
      path: evidence/mfa-audit-log.csv
      description: Monthly MFA compliance audit log
    - type: api_check
      endpoint: /api/v1/monitoring/checks/mfa-compliance
      frequency: daily
  owner: security-team@example.com
  review_frequency: quarterly
  next_review: "2026-04-15"
  compensating_controls: []
  risk_exceptions: []
```

### Required Fields

| Field | Description |
|-------|-------------|
| `spec.control_code` | The control identifier (e.g., A.9.4.2) |
| `spec.status` | Implementation status: not_implemented, planned, partial, implemented, effective |

### Optional Fields

| Field | Description |
|-------|-------------|
| `spec.framework` | Framework the control belongs to |
| `spec.implementation_type` | technical, administrative, physical, management |
| `spec.evidence` | Array of evidence references |
| `spec.owner` | Responsible team or individual |
| `spec.review_frequency` | How often the control is reviewed |
',
 'Reference guide for the ControlImplementation CaC resource kind, including YAML schema and examples.',
 ARRAY['ISO27001', 'NIST_CSF_2', 'NIST_800_53', 'PCI_DSS_4'],
 ARRAY['A.9.4.2', 'A.5.1'],
 ARRAY['compliance-as-code', 'control-implementation', 'yaml', 'gitops'],
 'intermediate', 8, true, true),

-- 2. Policy Example
(NULL, 'implementation_guide',
 'Compliance-as-Code: Policy Resource Reference',
 'cac-policy-reference',
 '## Policy Resource

A `Policy` resource represents a compliance or security policy managed as code.

### Example YAML

```yaml
apiVersion: complianceforge.io/v1
kind: Policy
metadata:
  name: information-security-policy
  uid: policy/information-security-policy
  labels:
    category: governance
    classification: internal
  annotations:
    approver: ciso@example.com
    legal-review: completed
spec:
  title: Information Security Policy
  version: "3.2"
  status: published
  owner: ciso@example.com
  effective_date: "2026-01-01"
  review_date: "2026-06-15"
  classification: internal
  applicable_frameworks:
    - ISO27001
    - NIST_CSF_2
    - CYBER_ESSENTIALS
  sections:
    - title: Purpose
      content: |
        This policy establishes the framework for managing
        information security across the organisation.
    - title: Scope
      content: |
        Applies to all employees, contractors, and third
        parties with access to organisational information assets.
    - title: Responsibilities
      content: |
        The CISO is responsible for overall policy governance.
        Department heads are responsible for implementation.
    - title: Policy Statements
      content: |
        1. All information assets must be classified.
        2. Access must be granted on least-privilege basis.
        3. All security incidents must be reported within 24h.
  approval_chain:
    - role: policy_owner
      status: approved
      date: "2025-12-10"
    - role: ciso
      status: approved
      date: "2025-12-15"
    - role: board
      status: approved
      date: "2025-12-20"
```

### Required Fields

| Field | Description |
|-------|-------------|
| `spec.title` | The policy title |

### Optional Fields

| Field | Description |
|-------|-------------|
| `spec.version` | Policy version string |
| `spec.status` | draft, under_review, pending_approval, approved, published, archived |
| `spec.sections` | Array of policy sections with title and content |
| `spec.approval_chain` | Approval workflow records |
',
 'Reference guide for the Policy CaC resource kind, including YAML schema and examples.',
 ARRAY['ISO27001', 'NIST_CSF_2', 'CYBER_ESSENTIALS'],
 ARRAY['A.5.1', 'A.5.2'],
 ARRAY['compliance-as-code', 'policy', 'yaml', 'gitops'],
 'intermediate', 6, true, true),

-- 3. RiskAcceptance Example
(NULL, 'implementation_guide',
 'Compliance-as-Code: RiskAcceptance Resource Reference',
 'cac-risk-acceptance-reference',
 '## RiskAcceptance Resource

A `RiskAcceptance` resource documents a formal risk acceptance decision, providing an auditable record of accepted risks and their compensating controls.

### Example YAML

```yaml
apiVersion: complianceforge.io/v1
kind: RiskAcceptance
metadata:
  name: legacy-erp-encryption-risk
  uid: riskacceptance/legacy-erp-encryption-risk
  labels:
    risk-level: medium
    department: finance
  annotations:
    jira-ticket: SEC-1234
spec:
  risk_id: RISK-2024-047
  risk_title: Legacy ERP system lacks modern encryption at rest
  risk_level: medium
  inherent_score: 12
  residual_score: 6
  accepted_by: cto@example.com
  accepted_date: "2026-01-15"
  expiry_date: "2026-07-15"
  justification: |
    Migration to cloud ERP with full encryption is scheduled
    for Q3 2026. Current system is isolated within a dedicated
    VLAN with strict access controls.
  compensating_controls:
    - Network segmentation isolating the legacy ERP
    - Enhanced IDS/IPS monitoring on the ERP VLAN
    - Restricted access to 12 named individuals only
    - Monthly access reviews with department head sign-off
    - Database-level column encryption for PII fields
  conditions:
    - Must be reviewed monthly by security team
    - Migration project must remain on schedule
    - Any new vulnerability in ERP triggers re-assessment
    - Access list must not exceed 15 individuals
  affected_controls:
    - A.10.1.1
    - A.10.1.2
```

### Required Fields

| Field | Description |
|-------|-------------|
| `spec.risk_id` | Unique risk identifier |
| `spec.accepted_by` | Person who accepted the risk |

### Optional Fields

| Field | Description |
|-------|-------------|
| `spec.risk_level` | critical, high, medium, low |
| `spec.expiry_date` | When the acceptance expires |
| `spec.compensating_controls` | Array of compensating control descriptions |
| `spec.conditions` | Conditions that must be maintained |
',
 'Reference guide for the RiskAcceptance CaC resource kind, including YAML schema and examples.',
 ARRAY['ISO27001', 'NIST_800_53'],
 ARRAY['A.10.1.1', 'A.10.1.2'],
 ARRAY['compliance-as-code', 'risk-acceptance', 'yaml', 'gitops'],
 'intermediate', 5, true, true),

-- 4. EvidenceConfig Example
(NULL, 'implementation_guide',
 'Compliance-as-Code: EvidenceConfig Resource Reference',
 'cac-evidence-config-reference',
 '## EvidenceConfig Resource

An `EvidenceConfig` resource defines automated or semi-automated evidence collection configurations for compliance controls.

### Example YAML

```yaml
apiVersion: complianceforge.io/v1
kind: EvidenceConfig
metadata:
  name: aws-cloudtrail-audit-evidence
  uid: evidenceconfig/aws-cloudtrail-audit-evidence
  labels:
    provider: aws
    evidence-type: automated
  annotations:
    integration: aws-cloudtrail
spec:
  evidence_type: automated_log
  collection_method: api_pull
  source_system: AWS CloudTrail
  frequency: daily
  retention_days: 365
  controls:
    - A.12.4.1
    - A.12.4.3
    - A.12.4.4
  format: json
  storage:
    bucket: compliance-evidence-prod
    prefix: cloudtrail/
    encryption: AES-256
  validation_rules:
    - field: eventSource
      required: true
    - field: eventTime
      required: true
    - field: userIdentity.arn
      required: true
  alerting:
    on_failure: true
    on_gap: true
    gap_threshold_hours: 36
    notify:
      - security-team@example.com
      - compliance@example.com
```

### Required Fields

| Field | Description |
|-------|-------------|
| `spec.evidence_type` | automated_log, screenshot, document, api_check, manual |
| `spec.collection_method` | api_pull, webhook, manual, scheduled_export |

### Optional Fields

| Field | Description |
|-------|-------------|
| `spec.frequency` | Collection frequency: real_time, hourly, daily, weekly, monthly |
| `spec.controls` | Array of control codes this evidence supports |
| `spec.validation_rules` | Rules to validate collected evidence |
| `spec.alerting` | Alert configuration for collection failures |
',
 'Reference guide for the EvidenceConfig CaC resource kind, including YAML schema and examples.',
 ARRAY['ISO27001', 'NIST_800_53', 'PCI_DSS_4'],
 ARRAY['A.12.4.1', 'A.12.4.3'],
 ARRAY['compliance-as-code', 'evidence', 'yaml', 'gitops', 'automation'],
 'advanced', 7, true, true),

-- 5. Getting Started Guide
(NULL, 'implementation_guide',
 'Getting Started with Compliance-as-Code in ComplianceForge',
 'cac-getting-started',
 '## Getting Started with Compliance-as-Code

Compliance-as-Code (CaC) lets you manage your compliance posture using Git-based workflows. Define controls, policies, risk acceptances, and evidence configurations as YAML files in a Git repository, and synchronise them with the ComplianceForge platform.

### Step 1: Create Your Repository Structure

```
compliance-repo/
  policies/
    information-security-policy.yaml
    acceptable-use-policy.yaml
  controls/
    iso27001/
      a5-organisational-controls.yaml
      a6-people-controls.yaml
      a8-technology-controls.yaml
  risks/
    accepted/
      legacy-erp-risk.yaml
  evidence/
    configs/
      aws-cloudtrail.yaml
      azure-sentinel.yaml
```

### Step 2: Connect Your Repository

1. Navigate to **Compliance as Code** in the sidebar
2. Click **Connect Repository**
3. Enter your repository URL and credentials
4. Select the sync direction (Pull, Push, or Bidirectional)
5. Choose whether to require approval for sync runs

### Step 3: Define Resources

Each YAML file follows the CaC resource format:

```yaml
apiVersion: complianceforge.io/v1
kind: <ResourceKind>
metadata:
  name: <unique-name>
  uid: <kind/name>
  labels:
    key: value
spec:
  # Kind-specific fields
```

### Supported Resource Kinds

| Kind | Description |
|------|-------------|
| `ControlImplementation` | Control implementation details and evidence |
| `Policy` | Compliance and security policies |
| `RiskAcceptance` | Formal risk acceptance records |
| `EvidenceConfig` | Evidence collection configurations |
| `Framework` | Custom framework definitions |
| `RiskTreatment` | Risk treatment plans |
| `AssetClassification` | Asset classification rules |
| `AuditSchedule` | Audit scheduling configurations |
| `IncidentPlaybook` | Incident response playbooks |
| `VendorAssessment` | Vendor assessment templates |

### Step 4: Sync and Monitor

- **Manual Sync**: Click "Sync Now" on any repository
- **Auto Sync**: Enable webhook-based syncing for real-time updates
- **Drift Detection**: Monitor for configuration drift between repo and platform
- **Approval Workflow**: Require manager approval before changes are applied

### Best Practices

1. **Use branches**: Develop changes on feature branches, merge to main
2. **Review before sync**: Enable require_approval for production repos
3. **Version everything**: Keep all compliance artefacts in version control
4. **Automate evidence**: Use EvidenceConfig to automate evidence collection
5. **Monitor drift**: Regularly check drift dashboard for discrepancies
',
 'Step-by-step guide to setting up and using Compliance-as-Code with ComplianceForge.',
 ARRAY['ISO27001', 'NIST_CSF_2', 'NIST_800_53', 'PCI_DSS_4', 'CYBER_ESSENTIALS'],
 ARRAY[],
 ARRAY['compliance-as-code', 'getting-started', 'gitops', 'yaml'],
 'beginner', 10, true, true);
