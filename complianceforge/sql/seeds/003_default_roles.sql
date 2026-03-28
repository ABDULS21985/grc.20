-- ============================================================
-- ComplianceForge — Seed: Default Roles & Permissions
-- ============================================================

-- System permissions
INSERT INTO permissions (id, resource, action, description) VALUES
-- Frameworks
('b0000001-0000-0000-0000-000000000001', 'frameworks', 'read', 'View compliance frameworks'),
('b0000001-0000-0000-0000-000000000002', 'frameworks', 'create', 'Create custom frameworks'),
('b0000001-0000-0000-0000-000000000003', 'frameworks', 'update', 'Update framework implementations'),
('b0000001-0000-0000-0000-000000000004', 'frameworks', 'delete', 'Remove framework adoptions'),
-- Controls
('b0000001-0000-0000-0000-000000000011', 'controls', 'read', 'View control implementations'),
('b0000001-0000-0000-0000-000000000012', 'controls', 'create', 'Create control implementations'),
('b0000001-0000-0000-0000-000000000013', 'controls', 'update', 'Update control status and evidence'),
('b0000001-0000-0000-0000-000000000014', 'controls', 'approve', 'Approve control effectiveness'),
-- Risks
('b0000001-0000-0000-0000-000000000021', 'risks', 'read', 'View risk register'),
('b0000001-0000-0000-0000-000000000022', 'risks', 'create', 'Create new risks'),
('b0000001-0000-0000-0000-000000000023', 'risks', 'update', 'Update risk assessments'),
('b0000001-0000-0000-0000-000000000024', 'risks', 'delete', 'Close/delete risks'),
('b0000001-0000-0000-0000-000000000025', 'risks', 'approve', 'Approve risk treatments'),
-- Policies
('b0000001-0000-0000-0000-000000000031', 'policies', 'read', 'View policies'),
('b0000001-0000-0000-0000-000000000032', 'policies', 'create', 'Draft new policies'),
('b0000001-0000-0000-0000-000000000033', 'policies', 'update', 'Edit policies'),
('b0000001-0000-0000-0000-000000000034', 'policies', 'approve', 'Approve and publish policies'),
('b0000001-0000-0000-0000-000000000035', 'policies', 'delete', 'Retire policies'),
-- Audits
('b0000001-0000-0000-0000-000000000041', 'audits', 'read', 'View audit plans and findings'),
('b0000001-0000-0000-0000-000000000042', 'audits', 'create', 'Plan new audits'),
('b0000001-0000-0000-0000-000000000043', 'audits', 'update', 'Execute audits and log findings'),
('b0000001-0000-0000-0000-000000000044', 'audits', 'approve', 'Approve audit reports'),
-- Incidents
('b0000001-0000-0000-0000-000000000051', 'incidents', 'read', 'View incidents'),
('b0000001-0000-0000-0000-000000000052', 'incidents', 'create', 'Report new incidents'),
('b0000001-0000-0000-0000-000000000053', 'incidents', 'update', 'Investigate and update incidents'),
('b0000001-0000-0000-0000-000000000054', 'incidents', 'approve', 'Close incidents'),
-- Vendors
('b0000001-0000-0000-0000-000000000061', 'vendors', 'read', 'View vendor risk register'),
('b0000001-0000-0000-0000-000000000062', 'vendors', 'create', 'Onboard new vendors'),
('b0000001-0000-0000-0000-000000000063', 'vendors', 'update', 'Update vendor assessments'),
('b0000001-0000-0000-0000-000000000064', 'vendors', 'approve', 'Approve vendor risk ratings'),
-- Reports & Dashboards
('b0000001-0000-0000-0000-000000000071', 'reports', 'read', 'View dashboards and reports'),
('b0000001-0000-0000-0000-000000000072', 'reports', 'export', 'Export reports'),
-- Settings & Users
('b0000001-0000-0000-0000-000000000081', 'settings', 'read', 'View organization settings'),
('b0000001-0000-0000-0000-000000000082', 'settings', 'configure', 'Configure organization settings'),
('b0000001-0000-0000-0000-000000000083', 'users', 'read', 'View users'),
('b0000001-0000-0000-0000-000000000084', 'users', 'create', 'Create users'),
('b0000001-0000-0000-0000-000000000085', 'users', 'update', 'Update users'),
('b0000001-0000-0000-0000-000000000086', 'users', 'delete', 'Deactivate users');

-- System roles (organization_id IS NULL = global)
INSERT INTO roles (id, organization_id, name, slug, description, is_system_role) VALUES
('c0000001-0000-0000-0000-000000000001', NULL, 'Organization Admin', 'org_admin', 'Full access to all organization features and settings.', true),
('c0000001-0000-0000-0000-000000000002', NULL, 'Compliance Manager', 'compliance_manager', 'Manages frameworks, controls, and compliance programs.', true),
('c0000001-0000-0000-0000-000000000003', NULL, 'Risk Manager', 'risk_manager', 'Manages the risk register, assessments, and treatments.', true),
('c0000001-0000-0000-0000-000000000004', NULL, 'Auditor', 'auditor', 'Plans and conducts audits, logs findings.', true),
('c0000001-0000-0000-0000-000000000005', NULL, 'Policy Owner', 'policy_owner', 'Drafts, manages, and publishes organizational policies.', true),
('c0000001-0000-0000-0000-000000000006', NULL, 'DPO', 'dpo', 'Data Protection Officer — manages privacy compliance and DSARs.', true),
('c0000001-0000-0000-0000-000000000007', NULL, 'CISO', 'ciso', 'Chief Information Security Officer — oversees security compliance.', true),
('c0000001-0000-0000-0000-000000000008', NULL, 'Viewer', 'viewer', 'Read-only access to dashboards and reports.', true),
('c0000001-0000-0000-0000-000000000009', NULL, 'External Auditor', 'external_auditor', 'Limited read-only access for external audit engagements.', true);

-- Org Admin gets ALL permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT 'c0000001-0000-0000-0000-000000000001', id FROM permissions;

-- Compliance Manager permissions
INSERT INTO role_permissions (role_id, permission_id) VALUES
('c0000001-0000-0000-0000-000000000002', 'b0000001-0000-0000-0000-000000000001'),
('c0000001-0000-0000-0000-000000000002', 'b0000001-0000-0000-0000-000000000002'),
('c0000001-0000-0000-0000-000000000002', 'b0000001-0000-0000-0000-000000000003'),
('c0000001-0000-0000-0000-000000000002', 'b0000001-0000-0000-0000-000000000011'),
('c0000001-0000-0000-0000-000000000002', 'b0000001-0000-0000-0000-000000000012'),
('c0000001-0000-0000-0000-000000000002', 'b0000001-0000-0000-0000-000000000013'),
('c0000001-0000-0000-0000-000000000002', 'b0000001-0000-0000-0000-000000000014'),
('c0000001-0000-0000-0000-000000000002', 'b0000001-0000-0000-0000-000000000021'),
('c0000001-0000-0000-0000-000000000002', 'b0000001-0000-0000-0000-000000000031'),
('c0000001-0000-0000-0000-000000000002', 'b0000001-0000-0000-0000-000000000041'),
('c0000001-0000-0000-0000-000000000002', 'b0000001-0000-0000-0000-000000000051'),
('c0000001-0000-0000-0000-000000000002', 'b0000001-0000-0000-0000-000000000071'),
('c0000001-0000-0000-0000-000000000002', 'b0000001-0000-0000-0000-000000000072');

-- Viewer gets read-only on everything
INSERT INTO role_permissions (role_id, permission_id)
SELECT 'c0000001-0000-0000-0000-000000000008', id FROM permissions WHERE action = 'read';

-- ============================================================
-- Default Risk Categories
-- ============================================================
INSERT INTO risk_categories (id, organization_id, name, code, description, color_hex, icon, sort_order, is_system_default) VALUES
('e0000001-0000-0000-0000-000000000001', NULL, 'Strategic Risk', 'STRATEGIC', 'Risks affecting the organisation''s long-term goals and strategic direction.', '#1E88E5', 'target', 1, true),
('e0000001-0000-0000-0000-000000000002', NULL, 'Operational Risk', 'OPERATIONAL', 'Risks arising from internal processes, people, and systems.', '#FB8C00', 'settings', 2, true),
('e0000001-0000-0000-0000-000000000003', NULL, 'Financial Risk', 'FINANCIAL', 'Risks related to financial loss, fraud, or market conditions.', '#43A047', 'pound-sterling', 3, true),
('e0000001-0000-0000-0000-000000000004', NULL, 'Compliance Risk', 'COMPLIANCE', 'Risks of regulatory non-compliance, fines, or sanctions.', '#E53935', 'shield', 4, true),
('e0000001-0000-0000-0000-000000000005', NULL, 'Cybersecurity Risk', 'CYBERSECURITY', 'Risks from cyber threats, data breaches, and system vulnerabilities.', '#8E24AA', 'lock', 5, true),
('e0000001-0000-0000-0000-000000000006', NULL, 'Reputational Risk', 'REPUTATIONAL', 'Risks affecting brand image, trust, and stakeholder confidence.', '#F4511E', 'eye', 6, true),
('e0000001-0000-0000-0000-000000000007', NULL, 'Third-Party Risk', 'THIRD_PARTY', 'Risks from vendors, suppliers, and external service providers.', '#00897B', 'users', 7, true),
('e0000001-0000-0000-0000-000000000008', NULL, 'Environmental / ESG Risk', 'ESG', 'Risks related to environmental, social, and governance factors.', '#2E7D32', 'leaf', 8, true),
('e0000001-0000-0000-0000-000000000009', NULL, 'Legal Risk', 'LEGAL', 'Risks from litigation, contract disputes, and regulatory actions.', '#5D4037', 'gavel', 9, true),
('e0000001-0000-0000-0000-000000000010', NULL, 'Technology Risk', 'TECHNOLOGY', 'Risks from technology failures, obsolescence, and digital transformation.', '#546E7A', 'cpu', 10, true),
('e0000001-0000-0000-0000-000000000011', NULL, 'People / HR Risk', 'PEOPLE', 'Risks related to workforce, talent, culture, and human resources.', '#D81B60', 'user-check', 11, true),
('e0000001-0000-0000-0000-000000000012', NULL, 'Geopolitical Risk', 'GEOPOLITICAL', 'Risks from political instability, sanctions, and trade disruptions.', '#37474F', 'globe', 12, true);

-- ============================================================
-- Default Policy Categories
-- ============================================================
INSERT INTO policy_categories (id, organization_id, name, code, description, sort_order, is_system_default) VALUES
('f0000001-0000-0000-0000-000000000001', NULL, 'Information Security', 'INFOSEC', 'Policies governing the protection of information assets.', 1, true),
('f0000001-0000-0000-0000-000000000002', NULL, 'Data Protection & Privacy', 'PRIVACY', 'Policies for GDPR/UK GDPR compliance and data handling.', 2, true),
('f0000001-0000-0000-0000-000000000003', NULL, 'Acceptable Use', 'ACCEPTABLE_USE', 'Policies defining acceptable use of IT systems and data.', 3, true),
('f0000001-0000-0000-0000-000000000004', NULL, 'Access Control', 'ACCESS_CONTROL', 'Policies governing access to systems, data, and facilities.', 4, true),
('f0000001-0000-0000-0000-000000000005', NULL, 'Business Continuity', 'BCP', 'Policies for business continuity and disaster recovery.', 5, true),
('f0000001-0000-0000-0000-000000000006', NULL, 'Incident Management', 'INCIDENT_MGMT', 'Policies for managing security and operational incidents.', 6, true),
('f0000001-0000-0000-0000-000000000007', NULL, 'Risk Management', 'RISK_MGMT', 'Policies defining the organisation''s risk management framework.', 7, true),
('f0000001-0000-0000-0000-000000000008', NULL, 'Third-Party & Vendor', 'VENDOR', 'Policies for managing third-party and vendor relationships.', 8, true),
('f0000001-0000-0000-0000-000000000009', NULL, 'HR & Employment', 'HR', 'Policies covering employment terms, conduct, and welfare.', 9, true),
('f0000001-0000-0000-0000-000000000010', NULL, 'Physical Security', 'PHYSICAL', 'Policies for physical access control and environmental security.', 10, true),
('f0000001-0000-0000-0000-000000000011', NULL, 'Change Management', 'CHANGE_MGMT', 'Policies governing changes to systems and processes.', 11, true),
('f0000001-0000-0000-0000-000000000012', NULL, 'Compliance', 'COMPLIANCE', 'General compliance policies and regulatory alignment.', 12, true),
('f0000001-0000-0000-0000-000000000013', NULL, 'Code of Conduct', 'CODE_CONDUCT', 'Employee code of conduct and ethical guidelines.', 13, true),
('f0000001-0000-0000-0000-000000000014', NULL, 'Anti-Bribery & Corruption', 'ABC', 'Policies for preventing bribery, corruption, and financial crime.', 14, true),
('f0000001-0000-0000-0000-000000000015', NULL, 'Whistleblowing', 'WHISTLEBLOWING', 'Policies for protected disclosures and whistleblowing procedures.', 15, true),
('f0000001-0000-0000-0000-000000000016', NULL, 'Environmental & Sustainability', 'ESG', 'Policies for environmental protection and sustainability commitments.', 16, true);
