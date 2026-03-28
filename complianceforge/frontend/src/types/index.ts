// ComplianceForge — TypeScript Types
// These mirror the Go backend models exactly

export type UUID = string;

// ── API Response Envelope ─────────────────────────────
export interface APIResponse<T = unknown> {
  success: boolean;
  data?: T;
  error?: APIError;
  meta?: unknown;
}

export interface APIError {
  code: string;
  message: string;
  details?: Record<string, string>;
}

export interface PaginatedResponse<T> {
  data: T[];
  pagination: Pagination;
}

export interface Pagination {
  page: number;
  page_size: number;
  total_items: number;
  total_pages: number;
  has_next: boolean;
  has_prev: boolean;
}

// ── Auth ──────────────────────────────────────────────
export interface LoginResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  token_type: string;
  user: User;
}

export interface User {
  id: UUID;
  organization_id: UUID;
  email: string;
  first_name: string;
  last_name: string;
  job_title: string;
  department: string;
  status: 'active' | 'inactive' | 'locked' | 'pending_verification';
  is_super_admin: boolean;
  timezone: string;
  language: string;
  last_login_at?: string;
  roles: Role[];
  created_at: string;
  updated_at: string;
}

export interface Role {
  id: UUID;
  name: string;
  slug: string;
  description: string;
  is_system_role: boolean;
}

// ── Framework ─────────────────────────────────────────
export interface ComplianceFramework {
  id: UUID;
  code: string;
  name: string;
  full_name: string;
  version: string;
  description: string;
  issuing_body: string;
  category: 'security' | 'privacy' | 'governance' | 'risk' | 'operational';
  applicable_regions: string[];
  applicable_industries: string[];
  is_system_framework: boolean;
  is_active: boolean;
  total_controls: number;
  icon_url: string;
  color_hex: string;
}

export interface FrameworkDomain {
  id: UUID;
  framework_id: UUID;
  code: string;
  name: string;
  description: string;
  sort_order: number;
  total_controls: number;
}

export interface FrameworkControl {
  id: UUID;
  framework_id: UUID;
  domain_id: UUID;
  code: string;
  title: string;
  description: string;
  guidance: string;
  control_type: 'preventive' | 'detective' | 'corrective' | 'directive' | 'compensating';
  implementation_type: 'technical' | 'administrative' | 'physical' | 'management';
  is_mandatory: boolean;
  priority: string;
}

export interface ComplianceScore {
  framework_id: UUID;
  framework_code: string;
  framework_name: string;
  total_controls: number;
  implemented: number;
  partially_implemented: number;
  not_implemented: number;
  not_applicable: number;
  compliance_score: number;
  maturity_avg: number;
}

// ── Risk ──────────────────────────────────────────────
export type RiskLevel = 'critical' | 'high' | 'medium' | 'low' | 'very_low';
export type RiskStatus = 'identified' | 'assessed' | 'treated' | 'accepted' | 'closed' | 'monitoring';

export interface Risk {
  id: UUID;
  risk_ref: string;
  title: string;
  description: string;
  risk_category_id: UUID;
  risk_source: string;
  status: RiskStatus;
  owner_user_id: UUID;
  inherent_likelihood: number;
  inherent_impact: number;
  inherent_risk_score: number;
  inherent_risk_level: RiskLevel;
  residual_likelihood: number;
  residual_impact: number;
  residual_risk_score: number;
  residual_risk_level: RiskLevel;
  financial_impact_eur: number;
  risk_velocity: string;
  next_review_date: string;
  tags: string[];
  created_at: string;
}

export interface RiskHeatmapEntry {
  risk_id: UUID;
  risk_ref: string;
  title: string;
  category_name: string;
  inherent_likelihood: number;
  inherent_impact: number;
  residual_likelihood: number;
  residual_impact: number;
  residual_risk_level: RiskLevel;
  owner_name: string;
  status: string;
}

// ── Policy ────────────────────────────────────────────
export type PolicyStatus = 'draft' | 'under_review' | 'pending_approval' | 'approved' | 'published' | 'archived' | 'retired';

export interface Policy {
  id: UUID;
  policy_ref: string;
  title: string;
  status: PolicyStatus;
  classification: string;
  current_version: number;
  review_frequency_months: number;
  next_review_date: string;
  review_status: string;
  is_mandatory: boolean;
  requires_attestation: boolean;
  tags: string[];
  created_at: string;
}

// ── Audit ─────────────────────────────────────────────
export type AuditStatus = 'planned' | 'in_progress' | 'completed' | 'cancelled';
export type FindingSeverity = 'critical' | 'high' | 'medium' | 'low' | 'informational';

export interface Audit {
  id: UUID;
  audit_ref: string;
  title: string;
  audit_type: string;
  status: AuditStatus;
  lead_auditor_id: UUID;
  planned_start_date: string;
  planned_end_date: string;
  total_findings: number;
  critical_findings: number;
  high_findings: number;
}

export interface AuditFinding {
  id: UUID;
  finding_ref: string;
  title: string;
  description: string;
  severity: FindingSeverity;
  status: string;
  finding_type: string;
  root_cause: string;
  recommendation: string;
  due_date: string;
}

// ── Incident ──────────────────────────────────────────
export type IncidentSeverity = 'critical' | 'high' | 'medium' | 'low';
export type IncidentStatus = 'open' | 'investigating' | 'contained' | 'resolved' | 'closed';

export interface Incident {
  id: UUID;
  incident_ref: string;
  title: string;
  description: string;
  incident_type: string;
  severity: IncidentSeverity;
  status: IncidentStatus;
  reported_at: string;
  is_data_breach: boolean;
  data_subjects_affected: number;
  notification_required: boolean;
  notification_deadline: string;
  dpa_notified_at?: string;
  is_nis2_reportable: boolean;
  financial_impact_eur: number;
}

// ── Vendor ────────────────────────────────────────────
export type VendorRiskTier = 'critical' | 'high' | 'medium' | 'low';

export interface Vendor {
  id: UUID;
  vendor_ref: string;
  name: string;
  status: string;
  risk_tier: VendorRiskTier;
  risk_score: number;
  data_processing: boolean;
  dpa_in_place: boolean;
  country_code: string;
  certifications: string[];
  contract_value: number;
  next_assessment_date: string;
}

// ── Dashboard ─────────────────────────────────────────
export interface DashboardSummary {
  overall_compliance_score: number;
  framework_scores: FrameworkScoreEntry[];
  risk_summary: Record<string, number>;
  open_incidents: number;
  open_findings: number;
  policies_due_for_review: number;
  vendors_high_risk: number;
  breaches_near_deadline: number;
}

export interface FrameworkScoreEntry {
  code: string;
  name: string;
  score: number;
}
