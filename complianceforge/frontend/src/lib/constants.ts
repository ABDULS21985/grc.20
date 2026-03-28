// ComplianceForge — Constants & Configuration
// Centralized color mappings, navigation structure, and category definitions.

// ── Severity Colors ──────────────────────────────────────
// Used for findings, incidents, vendor risk tiers, and general severity badges.
export const SEVERITY_COLORS: Record<string, string> = {
  critical: 'badge-critical',
  high: 'badge-high',
  medium: 'badge-medium',
  low: 'badge-low',
  informational: 'badge-info',
};

// ── Risk Level Colors ────────────────────────────────────
// Maps risk levels to Tailwind-based badge/bg classes.
export const RISK_LEVEL_COLORS: Record<string, string> = {
  critical: 'bg-red-600 text-white',
  high: 'bg-orange-500 text-white',
  medium: 'bg-yellow-500 text-black',
  low: 'bg-green-500 text-white',
  very_low: 'bg-green-300 text-black',
};

// ── Status Colors ────────────────────────────────────────
// Per-entity status color mappings for badges and indicators.

export const RISK_STATUS_COLORS: Record<string, string> = {
  identified: 'bg-blue-100 text-blue-800',
  assessed: 'bg-indigo-100 text-indigo-800',
  treated: 'bg-green-100 text-green-800',
  accepted: 'bg-yellow-100 text-yellow-800',
  closed: 'bg-gray-100 text-gray-800',
  monitoring: 'bg-purple-100 text-purple-800',
};

export const POLICY_STATUS_COLORS: Record<string, string> = {
  draft: 'bg-gray-100 text-gray-800',
  under_review: 'bg-blue-100 text-blue-800',
  pending_approval: 'bg-yellow-100 text-yellow-800',
  approved: 'bg-indigo-100 text-indigo-800',
  published: 'bg-green-100 text-green-800',
  archived: 'bg-gray-100 text-gray-600',
  retired: 'bg-red-100 text-red-800',
};

export const AUDIT_STATUS_COLORS: Record<string, string> = {
  planned: 'bg-blue-100 text-blue-800',
  in_progress: 'bg-yellow-100 text-yellow-800',
  completed: 'bg-green-100 text-green-800',
  cancelled: 'bg-red-100 text-red-800',
};

export const INCIDENT_STATUS_COLORS: Record<string, string> = {
  open: 'bg-red-100 text-red-800',
  investigating: 'bg-orange-100 text-orange-800',
  contained: 'bg-yellow-100 text-yellow-800',
  resolved: 'bg-green-100 text-green-800',
  closed: 'bg-gray-100 text-gray-800',
};

export const VENDOR_STATUS_COLORS: Record<string, string> = {
  active: 'bg-green-100 text-green-800',
  onboarding: 'bg-blue-100 text-blue-800',
  under_review: 'bg-yellow-100 text-yellow-800',
  suspended: 'bg-red-100 text-red-800',
  offboarded: 'bg-gray-100 text-gray-800',
};

// Convenience combined status map for generic lookups.
export const STATUS_COLORS: Record<string, string> = {
  // Risk statuses
  identified: RISK_STATUS_COLORS.identified,
  assessed: RISK_STATUS_COLORS.assessed,
  treated: RISK_STATUS_COLORS.treated,
  accepted: RISK_STATUS_COLORS.accepted,
  monitoring: RISK_STATUS_COLORS.monitoring,
  // Policy statuses
  draft: POLICY_STATUS_COLORS.draft,
  under_review: POLICY_STATUS_COLORS.under_review,
  pending_approval: POLICY_STATUS_COLORS.pending_approval,
  approved: POLICY_STATUS_COLORS.approved,
  published: POLICY_STATUS_COLORS.published,
  archived: POLICY_STATUS_COLORS.archived,
  retired: POLICY_STATUS_COLORS.retired,
  // Audit statuses
  planned: AUDIT_STATUS_COLORS.planned,
  in_progress: AUDIT_STATUS_COLORS.in_progress,
  completed: AUDIT_STATUS_COLORS.completed,
  cancelled: AUDIT_STATUS_COLORS.cancelled,
  // Incident statuses
  open: INCIDENT_STATUS_COLORS.open,
  investigating: INCIDENT_STATUS_COLORS.investigating,
  contained: INCIDENT_STATUS_COLORS.contained,
  resolved: INCIDENT_STATUS_COLORS.resolved,
  closed: INCIDENT_STATUS_COLORS.closed,
  // General
  active: VENDOR_STATUS_COLORS.active,
  onboarding: VENDOR_STATUS_COLORS.onboarding,
  suspended: VENDOR_STATUS_COLORS.suspended,
  offboarded: VENDOR_STATUS_COLORS.offboarded,
};

// ── Framework Categories ─────────────────────────────────
export const FRAMEWORK_CATEGORIES: Record<string, { label: string; color: string }> = {
  security: { label: 'Security', color: 'bg-blue-100 text-blue-800' },
  privacy: { label: 'Privacy', color: 'bg-purple-100 text-purple-800' },
  governance: { label: 'Governance', color: 'bg-indigo-100 text-indigo-800' },
  risk: { label: 'Risk', color: 'bg-orange-100 text-orange-800' },
  operational: { label: 'Operational', color: 'bg-teal-100 text-teal-800' },
};

// ── Navigation Items ─────────────────────────────────────
// Icon names reference lucide-react icon component names.
export interface NavItem {
  href: string;
  label: string;
  icon: string;
}

export const NAV_ITEMS: NavItem[] = [
  { href: '/dashboard', label: 'Dashboard', icon: 'LayoutDashboard' },
  { href: '/frameworks', label: 'Frameworks', icon: 'Shield' },
  { href: '/risks', label: 'Risk Register', icon: 'AlertTriangle' },
  { href: '/policies', label: 'Policies', icon: 'FileText' },
  { href: '/audits', label: 'Audits', icon: 'ClipboardCheck' },
  { href: '/incidents', label: 'Incidents', icon: 'AlertOctagon' },
  { href: '/vendors', label: 'Vendors', icon: 'Building2' },
  { href: '/assets', label: 'Assets', icon: 'Server' },
  { href: '/settings', label: 'Settings', icon: 'Settings' },
];

// ── Miscellaneous ────────────────────────────────────────
export const DEFAULT_PAGE_SIZE = 20;

export const RISK_LIKELIHOOD_LABELS: Record<number, string> = {
  1: 'Very Low',
  2: 'Low',
  3: 'Medium',
  4: 'High',
  5: 'Very High',
};

export const RISK_IMPACT_LABELS: Record<number, string> = {
  1: 'Negligible',
  2: 'Minor',
  3: 'Moderate',
  4: 'Major',
  5: 'Catastrophic',
};
