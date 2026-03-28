// ComplianceForge — Zod Validation Schemas
// Client-side validation for all entity creation/mutation forms.
// These mirror the backend validation rules.

import { z } from 'zod';

// ── Auth ─────────────────────────────────────────────────
export const loginSchema = z.object({
  email: z.string().email('Please enter a valid email address'),
  password: z.string().min(12, 'Password must be at least 12 characters'),
});

export type LoginInput = z.infer<typeof loginSchema>;

// ── Risk ─────────────────────────────────────────────────
export const createRiskSchema = z.object({
  title: z.string().min(5, 'Title must be at least 5 characters'),
  description: z.string().min(1, 'Description is required'),
  risk_category_id: z.string().uuid('Invalid risk category'),
  risk_source: z.string().min(1, 'Risk source is required'),
  owner_user_id: z.string().uuid('Invalid owner'),
  inherent_likelihood: z.number().int().min(1).max(5),
  inherent_impact: z.number().int().min(1).max(5),
  residual_likelihood: z.number().int().min(1).max(5),
  residual_impact: z.number().int().min(1).max(5),
  financial_impact_eur: z.number().min(0).optional(),
  tags: z.array(z.string()).optional(),
});

export type CreateRiskInput = z.infer<typeof createRiskSchema>;

// ── Policy ───────────────────────────────────────────────
export const createPolicySchema = z.object({
  title: z.string().min(3, 'Title must be at least 3 characters'),
  classification: z.string().min(1, 'Classification is required'),
  content: z.string().min(1, 'Policy content is required'),
  summary: z.string().optional(),
  review_frequency_months: z.number().int().min(1).max(36),
  is_mandatory: z.boolean(),
  requires_attestation: z.boolean(),
});

export type CreatePolicyInput = z.infer<typeof createPolicySchema>;

// ── Audit ────────────────────────────────────────────────
export const createAuditSchema = z.object({
  title: z.string().min(1, 'Title is required'),
  audit_type: z.string().min(1, 'Audit type is required'),
  lead_auditor_id: z.string().uuid('Invalid auditor'),
  planned_start_date: z.string().min(1, 'Start date is required'),
  planned_end_date: z.string().min(1, 'End date is required'),
});

export type CreateAuditInput = z.infer<typeof createAuditSchema>;

// ── Audit Finding ────────────────────────────────────────
export const createFindingSchema = z.object({
  title: z.string().min(1, 'Title is required'),
  description: z.string().min(1, 'Description is required'),
  severity: z.enum(['critical', 'high', 'medium', 'low', 'informational']),
  finding_type: z.string().min(1, 'Finding type is required'),
  root_cause: z.string().optional(),
  recommendation: z.string().min(1, 'Recommendation is required'),
  responsible_user_id: z.string().uuid('Invalid user').optional(),
  due_date: z.string().min(1, 'Due date is required'),
});

export type CreateFindingInput = z.infer<typeof createFindingSchema>;

// ── Incident ─────────────────────────────────────────────
export const reportIncidentSchema = z.object({
  title: z.string().min(5, 'Title must be at least 5 characters'),
  description: z.string().min(1, 'Description is required'),
  incident_type: z.string().min(1, 'Incident type is required'),
  severity: z.enum(['critical', 'high', 'medium', 'low']),
  is_data_breach: z.boolean(),
  data_subjects_affected: z.number().int().min(0).optional(),
  data_categories: z.array(z.string()).optional(),
});

export type ReportIncidentInput = z.infer<typeof reportIncidentSchema>;

// ── Vendor ───────────────────────────────────────────────
export const onboardVendorSchema = z.object({
  name: z.string().min(2, 'Name must be at least 2 characters'),
  country_code: z.string().min(1, 'Country is required'),
  risk_tier: z.enum(['critical', 'high', 'medium', 'low']),
  data_processing: z.boolean(),
  certifications: z.array(z.string()).optional(),
});

export type OnboardVendorInput = z.infer<typeof onboardVendorSchema>;

// ── Asset ────────────────────────────────────────────────
export const registerAssetSchema = z.object({
  name: z.string().min(1, 'Name is required'),
  asset_type: z.string().min(1, 'Asset type is required'),
  criticality: z.string().min(1, 'Criticality is required'),
  owner_user_id: z.string().uuid('Invalid owner').optional(),
  processes_personal_data: z.boolean(),
});

export type RegisterAssetInput = z.infer<typeof registerAssetSchema>;

// ── User ─────────────────────────────────────────────────
export const createUserSchema = z.object({
  email: z.string().email('Please enter a valid email address'),
  password: z.string().min(12, 'Password must be at least 12 characters'),
  first_name: z.string().min(1, 'First name is required'),
  last_name: z.string().min(1, 'Last name is required'),
  job_title: z.string().optional(),
  department: z.string().optional(),
  role_ids: z.array(z.string().uuid()).min(1, 'At least one role is required'),
});

export type CreateUserInput = z.infer<typeof createUserSchema>;
