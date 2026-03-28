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

// ── Notifications (Prompt 11) ────────────────────────
export type NotificationChannelType = 'email' | 'in_app' | 'webhook' | 'slack' | 'teams';
export type NotificationStatus = 'pending' | 'sent' | 'delivered' | 'failed' | 'bounced';
export type DigestFrequency = 'immediate' | 'hourly' | 'daily' | 'weekly';

export interface NotificationChannel {
  id: UUID;
  channel_type: NotificationChannelType;
  name: string;
  configuration: Record<string, unknown>;
  is_active: boolean;
  is_default: boolean;
  created_at: string;
}

export interface NotificationRule {
  id: UUID;
  name: string;
  event_type: string;
  severity_filter: string[] | null;
  conditions: Record<string, unknown>;
  channel_ids: UUID[];
  recipient_type: 'role' | 'user' | 'owner' | 'assignee' | 'dpo' | 'ciso' | 'custom';
  recipient_ids: UUID[];
  template_id: UUID;
  is_active: boolean;
  cooldown_minutes: number;
  escalation_after_minutes: number | null;
  created_at: string;
}

export interface Notification {
  id: UUID;
  event_type: string;
  subject: string;
  body: string;
  channel_type: NotificationChannelType;
  status: NotificationStatus;
  read_at: string | null;
  acknowledged_at: string | null;
  created_at: string;
}

export interface NotificationPreference {
  id: UUID;
  event_type: string;
  email_enabled: boolean;
  in_app_enabled: boolean;
  slack_enabled: boolean;
  digest_frequency: DigestFrequency;
  quiet_hours_start: string | null;
  quiet_hours_end: string | null;
  quiet_hours_timezone: string | null;
}

// ── Reports (Prompt 12) ─────────────────────────────
export type ReportType = 'compliance_status' | 'risk_register' | 'risk_heatmap' | 'audit_summary' | 'audit_findings' | 'incident_summary' | 'breach_register' | 'vendor_risk' | 'policy_status' | 'attestation_report' | 'gap_analysis' | 'cross_framework_mapping' | 'executive_summary' | 'kri_dashboard' | 'treatment_progress' | 'custom';
export type ReportFormat = 'pdf' | 'xlsx' | 'csv' | 'json';
export type ReportRunStatus = 'pending' | 'generating' | 'completed' | 'failed';
export type ScheduleFrequency = 'daily' | 'weekly' | 'monthly' | 'quarterly' | 'annually';

export interface ReportDefinition {
  id: UUID;
  name: string;
  description: string;
  report_type: ReportType;
  format: ReportFormat;
  filters: Record<string, unknown>;
  sections: Record<string, unknown>[];
  classification: string;
  include_executive_summary: boolean;
  include_appendices: boolean;
  branding: Record<string, unknown>;
  is_template: boolean;
  created_at: string;
}

export interface ReportSchedule {
  id: UUID;
  report_definition_id: UUID;
  name: string;
  frequency: ScheduleFrequency;
  day_of_week: number | null;
  day_of_month: number | null;
  time_of_day: string;
  timezone: string;
  recipient_user_ids: UUID[];
  recipient_emails: string[];
  delivery_channel: 'email' | 'storage' | 'both';
  is_active: boolean;
  last_run_at: string | null;
  next_run_at: string | null;
  created_at: string;
}

export interface ReportRun {
  id: UUID;
  report_definition_id: UUID;
  schedule_id: UUID | null;
  status: ReportRunStatus;
  format: ReportFormat;
  file_path: string | null;
  file_size_bytes: number | null;
  page_count: number | null;
  generation_time_ms: number | null;
  parameters: Record<string, unknown>;
  generated_by: UUID | null;
  error_message: string | null;
  created_at: string;
  completed_at: string | null;
}

// ── GDPR DSR (Prompt 13) ────────────────────────────
export type DSRRequestType = 'access' | 'erasure' | 'rectification' | 'portability' | 'restriction' | 'objection' | 'automated_decision';
export type DSRStatus = 'received' | 'identity_verification' | 'in_progress' | 'extended' | 'completed' | 'rejected' | 'withdrawn';
export type DSRPriority = 'standard' | 'urgent' | 'complex';
export type DSRSLAStatus = 'on_track' | 'at_risk' | 'overdue';

export interface DSRRequest {
  id: UUID;
  request_ref: string;
  request_type: DSRRequestType;
  status: DSRStatus;
  priority: DSRPriority;
  data_subject_name: string;
  data_subject_email: string;
  data_subject_id_verified: boolean;
  request_description: string;
  request_source: string;
  received_date: string;
  acknowledged_at: string | null;
  response_deadline: string;
  extended_deadline: string | null;
  extension_reason: string | null;
  assigned_to: UUID | null;
  data_systems_affected: string[];
  data_categories_affected: string[];
  third_parties_notified: string[];
  completed_at: string | null;
  sla_status: DSRSLAStatus;
  days_remaining: number;
  was_extended: boolean;
  was_completed_on_time: boolean | null;
  created_at: string;
}

export interface DSRTask {
  id: UUID;
  dsr_request_id: UUID;
  task_type: string;
  description: string;
  system_name: string | null;
  assigned_to: UUID | null;
  status: 'pending' | 'in_progress' | 'completed' | 'blocked' | 'not_applicable';
  due_date: string | null;
  completed_at: string | null;
  notes: string | null;
  sort_order: number;
}

export interface DSRAuditEntry {
  id: UUID;
  action: string;
  performed_by: UUID;
  description: string;
  created_at: string;
}

export interface DSRDashboard {
  total_requests: number;
  by_type: Record<string, number>;
  by_status: Record<string, number>;
  overdue_count: number;
  at_risk_count: number;
  avg_completion_days: number;
  sla_compliance_rate: number;
}

// ── NIS2 (Prompt 14) ────────────────────────────────
export type NIS2EntityType = 'essential' | 'important' | 'not_applicable';
export type NIS2PhaseStatus = 'not_required' | 'pending' | 'submitted' | 'overdue';
export type NIS2MeasureStatus = 'not_started' | 'in_progress' | 'implemented' | 'verified';

export interface NIS2EntityAssessment {
  id: UUID;
  entity_type: NIS2EntityType;
  sector: string;
  sub_sector: string;
  employee_count: number;
  annual_turnover_eur: number;
  assessment_date: string;
  is_in_scope: boolean;
  member_state: string;
  competent_authority: string;
  csirt_name: string;
  csirt_contact_email: string;
  csirt_reporting_url: string;
  created_at: string;
}

export interface NIS2IncidentReport {
  id: UUID;
  incident_id: UUID;
  report_ref: string;
  early_warning_status: NIS2PhaseStatus;
  early_warning_deadline: string;
  early_warning_submitted_at: string | null;
  notification_status: NIS2PhaseStatus;
  notification_deadline: string;
  notification_submitted_at: string | null;
  final_report_status: NIS2PhaseStatus;
  final_report_deadline: string;
  final_report_submitted_at: string | null;
  created_at: string;
}

export interface NIS2SecurityMeasure {
  id: UUID;
  measure_code: string;
  measure_title: string;
  measure_description: string;
  article_reference: string;
  implementation_status: NIS2MeasureStatus;
  owner_user_id: UUID | null;
  evidence_description: string | null;
  last_assessed_at: string | null;
  next_assessment_date: string | null;
  linked_control_ids: UUID[];
}

export interface NIS2ManagementRecord {
  id: UUID;
  board_member_name: string;
  board_member_role: string;
  training_completed: boolean;
  training_date: string | null;
  training_provider: string | null;
  risk_measures_approved: boolean;
  approval_date: string | null;
  next_training_due: string | null;
}

export interface NIS2Dashboard {
  entity_type: NIS2EntityType;
  is_in_scope: boolean;
  measures_total: number;
  measures_implemented: number;
  measures_verified: number;
  open_incidents: number;
  overdue_reports: number;
  management_trained: number;
  management_total: number;
}

// ── Continuous Monitoring (Prompt 15) ────────────────
export type CollectionMethod = 'manual' | 'api_fetch' | 'file_watch' | 'script_execution' | 'email_parse' | 'webhook_receive';
export type MonitorType = 'control_effectiveness' | 'evidence_freshness' | 'kri_threshold' | 'policy_attestation' | 'vendor_assessment' | 'training_completion';
export type MonitorStatus = 'passing' | 'failing' | 'unknown';
export type DriftType = 'control_degraded' | 'evidence_expired' | 'kri_breached' | 'policy_unattested' | 'vendor_overdue' | 'training_expired' | 'score_dropped';

export interface EvidenceCollectionConfig {
  id: UUID;
  control_implementation_id: UUID;
  name: string;
  collection_method: CollectionMethod;
  schedule_cron: string;
  schedule_description: string;
  acceptance_criteria: Record<string, unknown>[];
  failure_threshold: number;
  auto_update_control_status: boolean;
  is_active: boolean;
  last_collection_at: string | null;
  last_collection_status: string | null;
  next_collection_at: string | null;
  consecutive_failures: number;
  created_at: string;
}

export interface EvidenceCollectionRun {
  id: UUID;
  config_id: UUID;
  status: 'scheduled' | 'running' | 'success' | 'failed' | 'timeout' | 'validation_failed';
  started_at: string;
  completed_at: string | null;
  duration_ms: number | null;
  all_criteria_passed: boolean | null;
  error_message: string | null;
  created_at: string;
}

export interface ComplianceMonitor {
  id: UUID;
  name: string;
  monitor_type: MonitorType;
  target_entity_type: string;
  target_entity_id: UUID;
  check_frequency_cron: string;
  conditions: Record<string, unknown>;
  alert_on_failure: boolean;
  alert_severity: string;
  is_active: boolean;
  last_check_at: string | null;
  last_check_status: MonitorStatus;
  consecutive_failures: number;
  failure_since: string | null;
  created_at: string;
}

export interface ComplianceMonitorResult {
  id: UUID;
  monitor_id: UUID;
  status: 'passing' | 'failing';
  check_time: string;
  message: string | null;
  created_at: string;
}

export interface DriftEvent {
  id: UUID;
  drift_type: DriftType;
  severity: 'critical' | 'high' | 'medium' | 'low';
  entity_type: string;
  entity_id: UUID;
  entity_ref: string;
  description: string;
  previous_state: string;
  current_state: string;
  detected_at: string;
  acknowledged_at: string | null;
  acknowledged_by: UUID | null;
  resolved_at: string | null;
  resolved_by: UUID | null;
  resolution_notes: string | null;
  created_at: string;
}

export interface MonitoringDashboard {
  health_status: 'healthy' | 'degraded' | 'critical';
  active_drift_events: number;
  drift_by_severity: Record<string, number>;
  evidence_success_rate_24h: number;
  evidence_success_rate_7d: number;
  monitors_passing: number;
  monitors_failing: number;
  monitors_total: number;
}
