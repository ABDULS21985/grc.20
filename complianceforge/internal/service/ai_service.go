// Package service contains the AI Service for ComplianceForge.
// It provides integration with the Anthropic Claude API for AI-assisted
// compliance remediation planning, control guidance, evidence suggestions,
// policy drafting, and risk narrative generation.
package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// ============================================================
// CONSTANTS
// ============================================================

const (
	anthropicAPIURL     = "https://api.anthropic.com/v1/messages"
	anthropicAPIVersion = "2023-06-01"
	defaultMaxTokens    = 4096
	defaultRateLimit    = 100 // calls per hour per org
	rateLimitWindow     = time.Hour

	// Cost estimates per 1K tokens (EUR) for tracking
	inputTokenCostEUR  = 0.003
	outputTokenCostEUR = 0.015
)

// ============================================================
// RATE LIMITER (in-memory, per-org token bucket)
// ============================================================

type orgRateLimiter struct {
	mu        sync.Mutex
	tokens    int
	maxTokens int
	lastReset time.Time
	window    time.Duration
}

func newOrgRateLimiter(maxTokens int, window time.Duration) *orgRateLimiter {
	return &orgRateLimiter{
		tokens:    maxTokens,
		maxTokens: maxTokens,
		lastReset: time.Now(),
		window:    window,
	}
}

func (rl *orgRateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	if now.Sub(rl.lastReset) >= rl.window {
		rl.tokens = rl.maxTokens
		rl.lastReset = now
	}

	if rl.tokens <= 0 {
		return false
	}

	rl.tokens--
	return true
}

type rateLimiterMap struct {
	mu       sync.RWMutex
	limiters map[uuid.UUID]*orgRateLimiter
	maxCalls int
	window   time.Duration
}

func newRateLimiterMap(maxCalls int, window time.Duration) *rateLimiterMap {
	return &rateLimiterMap{
		limiters: make(map[uuid.UUID]*orgRateLimiter),
		maxCalls: maxCalls,
		window:   window,
	}
}

func (m *rateLimiterMap) Allow(orgID uuid.UUID) bool {
	m.mu.RLock()
	rl, ok := m.limiters[orgID]
	m.mu.RUnlock()

	if !ok {
		m.mu.Lock()
		// Double-check after acquiring write lock
		rl, ok = m.limiters[orgID]
		if !ok {
			rl = newOrgRateLimiter(m.maxCalls, m.window)
			m.limiters[orgID] = rl
		}
		m.mu.Unlock()
	}

	return rl.Allow()
}

// ============================================================
// CLAUDE API REQUEST / RESPONSE TYPES
// ============================================================

type claudeRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	Messages  []claudeMessage `json:"messages"`
}

type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type claudeAPIResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Model        string `json:"model"`
	StopReason   string `json:"stop_reason"`
	StopSequence string `json:"stop_sequence"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

type claudeResponse struct {
	Text         string
	Model        string
	InputTokens  int
	OutputTokens int
	LatencyMs    int
	CostEUR      float64
}

// ============================================================
// AI SERVICE REQUEST / RESPONSE TYPES
// ============================================================

// RemediationPlanRequest is the input for AI-generated remediation plan creation.
type RemediationPlanRequest struct {
	PlanType          string            `json:"plan_type"`
	FrameworkIDs      []uuid.UUID       `json:"framework_ids"`
	ScopeDescription  string            `json:"scope_description"`
	Gaps              []ComplianceGap   `json:"gaps"`
	RiskAppetite      string            `json:"risk_appetite"`
	BudgetEUR         float64           `json:"budget_eur"`
	TimelineMonths    int               `json:"timeline_months"`
	AvailableSkills   []string          `json:"available_skills"`
	ExistingControls  []string          `json:"existing_controls"`
	IndustryContext   string            `json:"industry_context"`
	RegulatoryContext string            `json:"regulatory_context"`
}

// ComplianceGap represents a gap identified in a compliance assessment.
type ComplianceGap struct {
	ControlCode   string `json:"control_code"`
	ControlTitle  string `json:"control_title"`
	FrameworkCode string `json:"framework_code"`
	CurrentStatus string `json:"current_status"`
	TargetStatus  string `json:"target_status"`
	GapSeverity   string `json:"gap_severity"`
	Description   string `json:"description"`
}

// RemediationPlanResponse is the AI-generated remediation plan structure.
type RemediationPlanResponse struct {
	InteractionID   uuid.UUID                `json:"interaction_id"`
	PlanName        string                   `json:"plan_name"`
	PlanDescription string                   `json:"plan_description"`
	Priority        string                   `json:"priority"`
	ConfidenceScore float64                  `json:"confidence_score"`
	Actions         []RemediationActionSuggestion `json:"actions"`
	EstimatedHours  float64                  `json:"estimated_total_hours"`
	EstimatedCost   float64                  `json:"estimated_total_cost_eur"`
	TimelineWeeks   int                      `json:"timeline_weeks"`
	RiskSummary     string                   `json:"risk_summary"`
	Assumptions     []string                 `json:"assumptions"`
}

// RemediationActionSuggestion is a single action suggested by the AI.
type RemediationActionSuggestion struct {
	Title                    string   `json:"title"`
	Description              string   `json:"description"`
	ActionType               string   `json:"action_type"`
	FrameworkControlCode     string   `json:"framework_control_code"`
	Priority                 string   `json:"priority"`
	EstimatedHours           float64  `json:"estimated_hours"`
	EstimatedCostEUR         float64  `json:"estimated_cost_eur"`
	RequiredSkills           []string `json:"required_skills"`
	ImplementationGuidance   string   `json:"implementation_guidance"`
	EvidenceSuggestions      []string `json:"evidence_suggestions"`
	ToolRecommendations      []string `json:"tool_recommendations"`
	RiskIfDeferred           string   `json:"risk_if_deferred"`
	CrossFrameworkBenefit    string   `json:"cross_framework_benefit"`
	DependsOn                []int    `json:"depends_on"`
	SortOrder                int      `json:"sort_order"`
}

// ControlGuidanceRequest is the input for AI control implementation guidance.
type ControlGuidanceRequest struct {
	ControlCode       string `json:"control_code"`
	ControlTitle      string `json:"control_title"`
	FrameworkCode     string `json:"framework_code"`
	CurrentStatus     string `json:"current_status"`
	OrganizationSize  string `json:"organization_size"`
	IndustryContext   string `json:"industry_context"`
}

// ControlGuidance is the AI response for control implementation guidance.
type ControlGuidance struct {
	InteractionID         uuid.UUID `json:"interaction_id"`
	ControlCode           string    `json:"control_code"`
	ImplementationSteps   []string  `json:"implementation_steps"`
	TechnicalMeasures     []string  `json:"technical_measures"`
	OrganizationalMeasures []string `json:"organizational_measures"`
	EvidenceRequired      []string  `json:"evidence_required"`
	CommonPitfalls        []string  `json:"common_pitfalls"`
	MaturityIndicators    map[string]string `json:"maturity_indicators"`
	EstimatedEffort       string    `json:"estimated_effort"`
	RelatedControls       []string  `json:"related_controls"`
}

// GapImpactAnalysis is the AI response for gap impact analysis.
type GapImpactAnalysis struct {
	InteractionID    uuid.UUID            `json:"interaction_id"`
	OverallRiskLevel string               `json:"overall_risk_level"`
	Summary          string               `json:"summary"`
	GapAssessments   []GapAssessmentItem  `json:"gap_assessments"`
	PrioritizedOrder []string             `json:"prioritized_order"`
	QuickWins        []string             `json:"quick_wins"`
	StrategicItems   []string             `json:"strategic_items"`
}

// GapAssessmentItem is a single gap assessment from the AI.
type GapAssessmentItem struct {
	ControlCode     string `json:"control_code"`
	RiskLevel       string `json:"risk_level"`
	BusinessImpact  string `json:"business_impact"`
	RegulatoryRisk  string `json:"regulatory_risk"`
	RemediationCost string `json:"remediation_cost"`
	Recommendation  string `json:"recommendation"`
}

// EvidenceTemplate is the AI response for evidence template suggestions.
type EvidenceTemplate struct {
	InteractionID    uuid.UUID `json:"interaction_id"`
	ControlCode      string    `json:"control_code"`
	ControlTitle     string    `json:"control_title"`
	EvidenceTypes    []EvidenceTypeItem `json:"evidence_types"`
	CollectionTips   []string  `json:"collection_tips"`
	ReviewFrequency  string    `json:"review_frequency"`
}

// EvidenceTypeItem describes a type of evidence the AI suggests.
type EvidenceTypeItem struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Format      string `json:"format"`
	Example     string `json:"example"`
}

// PolicyDraftRequest is the input for AI policy section drafting.
type PolicyDraftRequest struct {
	PolicyType       string   `json:"policy_type"`
	SectionTitle     string   `json:"section_title"`
	FrameworkCodes   []string `json:"framework_codes"`
	IndustryContext  string   `json:"industry_context"`
	ExistingPolicies []string `json:"existing_policies"`
	TonePreference   string   `json:"tone_preference"`
}

// PolicyDraft is the AI response for policy section drafting.
type PolicyDraft struct {
	InteractionID  uuid.UUID `json:"interaction_id"`
	SectionTitle   string    `json:"section_title"`
	Content        string    `json:"content"`
	KeyPoints      []string  `json:"key_points"`
	Definitions    map[string]string `json:"definitions"`
	RelatedClauses []string  `json:"related_clauses"`
	ReviewNotes    []string  `json:"review_notes"`
}

// RiskNarrativeRequest is the input for AI risk narrative generation.
type RiskNarrativeRequest struct {
	RiskTitle       string  `json:"risk_title"`
	RiskDescription string  `json:"risk_description"`
	RiskCategory    string  `json:"risk_category"`
	Likelihood      int     `json:"likelihood"`
	Impact          int     `json:"impact"`
	ExistingControls []string `json:"existing_controls"`
	IndustryContext string   `json:"industry_context"`
}

// RiskNarrative is the AI response for risk narrative generation.
type RiskNarrative struct {
	InteractionID      uuid.UUID `json:"interaction_id"`
	ExecutiveSummary   string    `json:"executive_summary"`
	DetailedAnalysis   string    `json:"detailed_analysis"`
	ThreatScenarios    []string  `json:"threat_scenarios"`
	MitigationOptions  []string  `json:"mitigation_options"`
	ResidualRiskNote   string    `json:"residual_risk_note"`
	MonitoringAdvice   string    `json:"monitoring_advice"`
}

// AIUsageStats holds usage statistics for an organization.
type AIUsageStats struct {
	TotalInteractions  int     `json:"total_interactions"`
	TotalInputTokens   int     `json:"total_input_tokens"`
	TotalOutputTokens  int     `json:"total_output_tokens"`
	TotalCostEUR       float64 `json:"total_cost_eur"`
	AvgLatencyMs       int     `json:"avg_latency_ms"`
	AvgRating          float64 `json:"avg_rating"`
	RatedInteractions  int     `json:"rated_interactions"`
	InteractionsByType map[string]int `json:"interactions_by_type"`
	Last30DaysCost     float64 `json:"last_30_days_cost_eur"`
}

// AIFeedbackRequest is the input for rating an AI interaction.
type AIFeedbackRequest struct {
	InteractionID uuid.UUID `json:"interaction_id"`
	Rating        int       `json:"rating"`
	Feedback      string    `json:"feedback"`
}

// ============================================================
// STATIC KNOWLEDGE BASE (fallback when API is unavailable)
// ============================================================

var staticControlGuidance = map[string]ControlGuidance{
	"ISO27001-A.5.1": {
		ControlCode:         "A.5.1",
		ImplementationSteps: []string{"Define information security policy scope", "Draft policy with management commitment", "Obtain senior management approval", "Communicate to all relevant parties", "Establish regular review cycle"},
		TechnicalMeasures:   []string{"Policy management system", "Document version control", "Electronic acknowledgement tracking"},
		OrganizationalMeasures: []string{"Policy review board", "Annual policy review schedule", "Employee awareness programme"},
		EvidenceRequired:    []string{"Approved policy document", "Management sign-off records", "Distribution and acknowledgement logs", "Review meeting minutes"},
		CommonPitfalls:      []string{"Policy too generic to be actionable", "Lack of management buy-in", "No review cycle established"},
		MaturityIndicators:  map[string]string{"initial": "Policy exists but not distributed", "managed": "Policy distributed and acknowledged", "optimized": "Policy regularly reviewed and updated based on threat landscape"},
		EstimatedEffort:     "2-4 weeks for initial development",
		RelatedControls:     []string{"A.5.2", "A.5.3", "A.6.1"},
	},
	"ISO27001-A.5.2": {
		ControlCode:         "A.5.2",
		ImplementationSteps: []string{"Identify all relevant roles and responsibilities", "Document ISMS roles in organizational chart", "Assign information security responsibilities", "Communicate responsibilities to all parties", "Review role assignments regularly"},
		TechnicalMeasures:   []string{"Identity and access management system", "Role-based access control matrix"},
		OrganizationalMeasures: []string{"RACI matrix for security responsibilities", "Job descriptions with security duties", "Security committee charter"},
		EvidenceRequired:    []string{"RACI matrix document", "Updated job descriptions", "Security committee meeting minutes", "Role assignment records"},
		CommonPitfalls:      []string{"Overlapping responsibilities without clarity", "Missing accountability for key functions", "Responsibilities not aligned with authority"},
		MaturityIndicators:  map[string]string{"initial": "Roles informally assigned", "managed": "Formal RACI with documented assignments", "optimized": "Dynamic role management with regular skills assessment"},
		EstimatedEffort:     "1-2 weeks",
		RelatedControls:     []string{"A.5.1", "A.5.3", "A.5.4"},
	},
	"ISO27001-A.6.1": {
		ControlCode:         "A.6.1",
		ImplementationSteps: []string{"Perform background verification checks", "Define screening procedures per role sensitivity", "Include security clauses in employment contracts", "Verify qualifications and references", "Document screening results securely"},
		TechnicalMeasures:   []string{"HR information system integration", "Automated background check services", "Secure document storage"},
		OrganizationalMeasures: []string{"Pre-employment screening policy", "Contractor screening requirements", "Confidentiality agreements"},
		EvidenceRequired:    []string{"Screening policy document", "Completed screening records", "Signed confidentiality agreements", "Employment contracts with security clauses"},
		CommonPitfalls:      []string{"Inconsistent screening across roles", "No contractor screening process", "Screening records not securely stored"},
		MaturityIndicators:  map[string]string{"initial": "Basic reference checks only", "managed": "Risk-based screening for all roles", "optimized": "Continuous monitoring and periodic re-screening"},
		EstimatedEffort:     "2-3 weeks to establish process",
		RelatedControls:     []string{"A.6.2", "A.6.3", "A.5.2"},
	},
	"ISO27001-A.8.1": {
		ControlCode:         "A.8.1",
		ImplementationSteps: []string{"Develop acceptable use policy for user devices", "Define device enrollment procedures", "Implement endpoint protection on all devices", "Establish BYOD policy if applicable", "Configure device management platform"},
		TechnicalMeasures:   []string{"Mobile device management (MDM)", "Endpoint detection and response (EDR)", "Full disk encryption", "Remote wipe capability"},
		OrganizationalMeasures: []string{"User device policy", "BYOD agreement forms", "Device inventory management", "Regular compliance audits"},
		EvidenceRequired:    []string{"Device policy document", "MDM enrollment reports", "Encryption status reports", "Device inventory register"},
		CommonPitfalls:      []string{"Incomplete device inventory", "BYOD devices not covered", "No remote wipe procedures tested"},
		MaturityIndicators:  map[string]string{"initial": "Basic antivirus on corporate devices", "managed": "MDM with enforced policies on all devices", "optimized": "Zero-trust endpoint verification with continuous compliance"},
		EstimatedEffort:     "4-6 weeks for full implementation",
		RelatedControls:     []string{"A.8.2", "A.8.3", "A.8.5"},
	},
	"ISO27001-A.8.5": {
		ControlCode:         "A.8.5",
		ImplementationSteps: []string{"Define authentication policy", "Implement multi-factor authentication for critical systems", "Establish password complexity requirements", "Configure single sign-on where appropriate", "Implement privileged access management"},
		TechnicalMeasures:   []string{"Multi-factor authentication (MFA)", "Single sign-on (SSO)", "Privileged access management (PAM)", "Password managers"},
		OrganizationalMeasures: []string{"Authentication policy", "Password management guidelines", "Access review procedures", "Service account management process"},
		EvidenceRequired:    []string{"Authentication policy", "MFA enrollment reports", "Access review records", "PAM configuration evidence"},
		CommonPitfalls:      []string{"MFA not enforced for all admin access", "Service accounts with static passwords", "No regular access reviews"},
		MaturityIndicators:  map[string]string{"initial": "Password-only authentication", "managed": "MFA for critical systems", "optimized": "Adaptive authentication with continuous risk assessment"},
		EstimatedEffort:     "3-5 weeks",
		RelatedControls:     []string{"A.8.2", "A.8.3", "A.8.4", "A.5.15"},
	},
	"ISO27001-A.5.15": {
		ControlCode:         "A.5.15",
		ImplementationSteps: []string{"Define access control policy aligned to business needs", "Implement role-based access control (RBAC)", "Establish access provisioning workflow", "Configure least privilege access", "Schedule regular access reviews"},
		TechnicalMeasures:   []string{"Identity governance platform", "RBAC implementation", "Just-in-time access provisioning", "Automated access certification"},
		OrganizationalMeasures: []string{"Access control policy", "Access request and approval process", "Quarterly access review schedule", "Segregation of duties matrix"},
		EvidenceRequired:    []string{"Access control policy", "RBAC matrix", "Access review reports", "Provisioning and de-provisioning logs"},
		CommonPitfalls:      []string{"Excessive permissions accumulated over time", "No regular recertification", "Segregation of duties not enforced"},
		MaturityIndicators:  map[string]string{"initial": "Ad-hoc access management", "managed": "Formal RBAC with periodic reviews", "optimized": "Automated identity governance with continuous certification"},
		EstimatedEffort:     "4-8 weeks",
		RelatedControls:     []string{"A.5.16", "A.5.17", "A.5.18", "A.8.5"},
	},
	"ISO27001-A.8.9": {
		ControlCode:         "A.8.9",
		ImplementationSteps: []string{"Inventory all configuration items", "Define secure configuration baselines", "Implement configuration management tooling", "Establish change control procedures for configurations", "Automate configuration compliance checking"},
		TechnicalMeasures:   []string{"Configuration management database (CMDB)", "Infrastructure as code", "Automated hardening tools", "Configuration drift detection"},
		OrganizationalMeasures: []string{"Configuration management policy", "Baseline approval process", "Change advisory board for config changes", "Regular configuration audits"},
		EvidenceRequired:    []string{"Configuration baselines", "CMDB records", "Configuration audit reports", "Change records for configuration items"},
		CommonPitfalls:      []string{"Configuration baselines not maintained", "No automated drift detection", "Change process bypassed for urgent fixes"},
		MaturityIndicators:  map[string]string{"initial": "Manual configuration with no baselines", "managed": "Documented baselines with periodic audits", "optimized": "Automated configuration enforcement with continuous compliance"},
		EstimatedEffort:     "6-10 weeks",
		RelatedControls:     []string{"A.8.8", "A.8.10", "A.8.25"},
	},
	"ISO27001-A.8.15": {
		ControlCode:         "A.8.15",
		ImplementationSteps: []string{"Define logging requirements for all systems", "Implement centralized log management", "Configure log retention policies", "Set up log monitoring and alerting", "Protect log integrity"},
		TechnicalMeasures:   []string{"SIEM platform", "Centralized log aggregation", "Log integrity protection (hashing)", "Automated alerting rules"},
		OrganizationalMeasures: []string{"Logging policy", "Log review procedures", "Incident response integration", "Retention schedule"},
		EvidenceRequired:    []string{"Logging policy", "SIEM configuration records", "Log review reports", "Alert response records"},
		CommonPitfalls:      []string{"Insufficient log sources covered", "Alert fatigue from too many false positives", "Logs not retained long enough for compliance"},
		MaturityIndicators:  map[string]string{"initial": "Basic system logging only", "managed": "Centralized logging with defined retention", "optimized": "AI-driven log analytics with automated threat detection"},
		EstimatedEffort:     "4-8 weeks",
		RelatedControls:     []string{"A.8.16", "A.8.17", "A.5.28"},
	},
	"ISO27001-A.5.28": {
		ControlCode:         "A.5.28",
		ImplementationSteps: []string{"Develop evidence collection procedures", "Identify digital forensics requirements", "Train incident responders on evidence handling", "Establish chain of custody procedures", "Test evidence collection capabilities"},
		TechnicalMeasures:   []string{"Forensic imaging tools", "Write blockers", "Secure evidence storage", "Chain of custody tracking system"},
		OrganizationalMeasures: []string{"Evidence collection policy", "Forensic readiness plan", "Legal liaison procedures", "Training programme for responders"},
		EvidenceRequired:    []string{"Evidence collection procedures", "Chain of custody forms", "Forensic readiness plan", "Training records"},
		CommonPitfalls:      []string{"Evidence contamination due to poor handling", "No chain of custody documentation", "Cloud evidence collection not addressed"},
		MaturityIndicators:  map[string]string{"initial": "No formal evidence procedures", "managed": "Documented procedures with trained staff", "optimized": "Automated forensic readiness with regular testing"},
		EstimatedEffort:     "3-5 weeks",
		RelatedControls:     []string{"A.5.24", "A.5.25", "A.5.26", "A.8.15"},
	},
	"ISO27001-A.5.30": {
		ControlCode:         "A.5.30",
		ImplementationSteps: []string{"Conduct business impact analysis (BIA)", "Define recovery time and point objectives", "Develop business continuity plans", "Implement technical recovery capabilities", "Test plans regularly with tabletop and full exercises"},
		TechnicalMeasures:   []string{"Backup and recovery systems", "Disaster recovery site", "High availability architecture", "Automated failover mechanisms"},
		OrganizationalMeasures: []string{"Business continuity policy", "BIA documentation", "Crisis communication plan", "Regular testing schedule"},
		EvidenceRequired:    []string{"BIA report", "Business continuity plans", "Test exercise reports", "Recovery capability evidence"},
		CommonPitfalls:      []string{"Plans not tested regularly", "BIA not updated after changes", "Recovery capabilities not validated"},
		MaturityIndicators:  map[string]string{"initial": "Basic backup procedures only", "managed": "Documented BC plans with annual testing", "optimized": "Resilient architecture with automated failover and continuous testing"},
		EstimatedEffort:     "8-12 weeks",
		RelatedControls:     []string{"A.5.29", "A.8.13", "A.8.14"},
	},
}

// ============================================================
// AI SERVICE
// ============================================================

// AIService provides AI-assisted compliance capabilities using the Anthropic Claude API.
type AIService struct {
	pool        *pgxpool.Pool
	apiKey      string
	model       string
	maxTokens   int
	httpClient  *http.Client
	rateLimiter *rateLimiterMap
}

// NewAIService creates a new AIService instance.
func NewAIService(pool *pgxpool.Pool, apiKey, model string) *AIService {
	if model == "" {
		model = "claude-sonnet-4-20250514"
	}

	return &AIService{
		pool:      pool,
		apiKey:    apiKey,
		model:     model,
		maxTokens: defaultMaxTokens,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
		rateLimiter: newRateLimiterMap(defaultRateLimit, rateLimitWindow),
	}
}

// isAvailable checks if the AI service is properly configured.
func (s *AIService) isAvailable() bool {
	return s.apiKey != ""
}

// ============================================================
// PUBLIC METHODS
// ============================================================

// GenerateRemediationPlan calls the Claude API to generate a full remediation plan.
func (s *AIService) GenerateRemediationPlan(ctx context.Context, orgID uuid.UUID, req RemediationPlanRequest) (*RemediationPlanResponse, error) {
	prompt := buildRemediationPlanPrompt(req)

	resp, err := s.callClaude(ctx, orgID, "remediation_plan", prompt)
	if err != nil {
		log.Warn().Err(err).Str("org_id", orgID.String()).Msg("AI remediation plan generation failed, using fallback")
		return s.fallbackRemediationPlan(req)
	}

	var result RemediationPlanResponse
	if err := extractJSONFromResponse(resp.Text, &result); err != nil {
		log.Warn().Err(err).Msg("Failed to parse AI response for remediation plan, using fallback")
		return s.fallbackRemediationPlan(req)
	}

	result.InteractionID, err = s.getLastInteractionID(ctx, orgID, "remediation_plan")
	if err != nil {
		log.Warn().Err(err).Msg("Failed to fetch interaction ID")
	}

	return &result, nil
}

// GenerateControlGuidance generates implementation guidance for a specific control.
func (s *AIService) GenerateControlGuidance(ctx context.Context, orgID uuid.UUID, req ControlGuidanceRequest) (*ControlGuidance, error) {
	prompt := buildControlGuidancePrompt(req)

	resp, err := s.callClaude(ctx, orgID, "control_guidance", prompt)
	if err != nil {
		log.Warn().Err(err).Msg("AI control guidance failed, using fallback")
		return s.fallbackControlGuidance(req)
	}

	var result ControlGuidance
	if err := extractJSONFromResponse(resp.Text, &result); err != nil {
		log.Warn().Err(err).Msg("Failed to parse AI response for control guidance, using fallback")
		return s.fallbackControlGuidance(req)
	}

	result.InteractionID, err = s.getLastInteractionID(ctx, orgID, "control_guidance")
	if err != nil {
		log.Warn().Err(err).Msg("Failed to fetch interaction ID")
	}

	return &result, nil
}

// AnalyseGapImpact analyses the impact of compliance gaps.
func (s *AIService) AnalyseGapImpact(ctx context.Context, orgID uuid.UUID, gaps []ComplianceGap) (*GapImpactAnalysis, error) {
	prompt := buildGapImpactPrompt(gaps)

	resp, err := s.callClaude(ctx, orgID, "gap_impact_analysis", prompt)
	if err != nil {
		log.Warn().Err(err).Msg("AI gap impact analysis failed, returning basic analysis")
		return s.fallbackGapAnalysis(gaps), nil
	}

	var result GapImpactAnalysis
	if err := extractJSONFromResponse(resp.Text, &result); err != nil {
		log.Warn().Err(err).Msg("Failed to parse AI gap analysis response")
		return s.fallbackGapAnalysis(gaps), nil
	}

	result.InteractionID, err = s.getLastInteractionID(ctx, orgID, "gap_impact_analysis")
	if err != nil {
		log.Warn().Err(err).Msg("Failed to fetch interaction ID")
	}

	return &result, nil
}

// SuggestEvidenceTemplate suggests evidence collection templates for a control.
func (s *AIService) SuggestEvidenceTemplate(ctx context.Context, orgID uuid.UUID, controlCode, controlTitle string) (*EvidenceTemplate, error) {
	prompt := fmt.Sprintf(`You are a compliance expert. Suggest evidence collection templates for the following control.

Control Code: %s
Control Title: %s

Respond with JSON only:
{
  "control_code": "%s",
  "control_title": "%s",
  "evidence_types": [
    {"type": "...", "description": "...", "format": "...", "example": "..."}
  ],
  "collection_tips": ["..."],
  "review_frequency": "..."
}`, controlCode, controlTitle, controlCode, controlTitle)

	resp, err := s.callClaude(ctx, orgID, "evidence_suggestion", prompt)
	if err != nil {
		log.Warn().Err(err).Msg("AI evidence suggestion failed, returning default template")
		return s.fallbackEvidenceTemplate(controlCode, controlTitle), nil
	}

	var result EvidenceTemplate
	if err := extractJSONFromResponse(resp.Text, &result); err != nil {
		log.Warn().Err(err).Msg("Failed to parse AI evidence template response")
		return s.fallbackEvidenceTemplate(controlCode, controlTitle), nil
	}

	result.InteractionID, err = s.getLastInteractionID(ctx, orgID, "evidence_suggestion")
	if err != nil {
		log.Warn().Err(err).Msg("Failed to fetch interaction ID")
	}

	return &result, nil
}

// DraftPolicySection generates a draft policy section using AI.
func (s *AIService) DraftPolicySection(ctx context.Context, orgID uuid.UUID, req PolicyDraftRequest) (*PolicyDraft, error) {
	prompt := buildPolicyDraftPrompt(req)

	resp, err := s.callClaude(ctx, orgID, "policy_draft", prompt)
	if err != nil {
		log.Warn().Err(err).Msg("AI policy draft failed")
		return nil, fmt.Errorf("AI policy drafting unavailable: %w", err)
	}

	var result PolicyDraft
	if err := extractJSONFromResponse(resp.Text, &result); err != nil {
		return nil, fmt.Errorf("failed to parse AI policy draft response: %w", err)
	}

	result.InteractionID, err = s.getLastInteractionID(ctx, orgID, "policy_draft")
	if err != nil {
		log.Warn().Err(err).Msg("Failed to fetch interaction ID")
	}

	return &result, nil
}

// AssessRiskNarrative generates a risk narrative using AI.
func (s *AIService) AssessRiskNarrative(ctx context.Context, orgID uuid.UUID, req RiskNarrativeRequest) (*RiskNarrative, error) {
	prompt := buildRiskNarrativePrompt(req)

	resp, err := s.callClaude(ctx, orgID, "risk_narrative", prompt)
	if err != nil {
		log.Warn().Err(err).Msg("AI risk narrative failed")
		return nil, fmt.Errorf("AI risk narrative unavailable: %w", err)
	}

	var result RiskNarrative
	if err := extractJSONFromResponse(resp.Text, &result); err != nil {
		return nil, fmt.Errorf("failed to parse AI risk narrative response: %w", err)
	}

	result.InteractionID, err = s.getLastInteractionID(ctx, orgID, "risk_narrative")
	if err != nil {
		log.Warn().Err(err).Msg("Failed to fetch interaction ID")
	}

	return &result, nil
}

// GetUsageStats returns AI usage statistics for an organization.
func (s *AIService) GetUsageStats(ctx context.Context, orgID uuid.UUID) (*AIUsageStats, error) {
	stats := &AIUsageStats{
		InteractionsByType: make(map[string]int),
	}

	err := s.pool.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COALESCE(SUM(input_tokens), 0),
			COALESCE(SUM(output_tokens), 0),
			COALESCE(SUM(cost_eur), 0),
			COALESCE(AVG(latency_ms), 0)::int
		FROM ai_interaction_logs
		WHERE organization_id = $1
	`, orgID).Scan(
		&stats.TotalInteractions,
		&stats.TotalInputTokens,
		&stats.TotalOutputTokens,
		&stats.TotalCostEUR,
		&stats.AvgLatencyMs,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch AI usage stats: %w", err)
	}

	err = s.pool.QueryRow(ctx, `
		SELECT
			COALESCE(AVG(rating), 0),
			COUNT(rating)
		FROM ai_interaction_logs
		WHERE organization_id = $1 AND rating IS NOT NULL
	`, orgID).Scan(&stats.AvgRating, &stats.RatedInteractions)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch AI rating stats: %w", err)
	}

	rows, err := s.pool.Query(ctx, `
		SELECT interaction_type, COUNT(*)
		FROM ai_interaction_logs
		WHERE organization_id = $1
		GROUP BY interaction_type
	`, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch interaction type stats: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var iType string
		var count int
		if err := rows.Scan(&iType, &count); err != nil {
			continue
		}
		stats.InteractionsByType[iType] = count
	}

	err = s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(cost_eur), 0)
		FROM ai_interaction_logs
		WHERE organization_id = $1 AND created_at >= now() - interval '30 days'
	`, orgID).Scan(&stats.Last30DaysCost)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch 30-day cost: %w", err)
	}

	return stats, nil
}

// RateInteraction records a user's rating and feedback for an AI interaction.
func (s *AIService) RateInteraction(ctx context.Context, orgID, interactionID uuid.UUID, rating int, feedback string) error {
	if rating < 1 || rating > 5 {
		return fmt.Errorf("rating must be between 1 and 5")
	}

	tag, err := s.pool.Exec(ctx, `
		UPDATE ai_interaction_logs
		SET rating = $1, feedback = $2
		WHERE id = $3 AND organization_id = $4
	`, rating, feedback, interactionID, orgID)
	if err != nil {
		return fmt.Errorf("failed to update AI interaction rating: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("interaction not found")
	}

	return nil
}

// ============================================================
// INTERNAL: Claude API call with logging and rate limiting
// ============================================================

func (s *AIService) callClaude(ctx context.Context, orgID uuid.UUID, interactionType, prompt string) (*claudeResponse, error) {
	if !s.isAvailable() {
		return nil, fmt.Errorf("AI service not configured: API key missing")
	}

	if !s.rateLimiter.Allow(orgID) {
		return nil, fmt.Errorf("AI rate limit exceeded for organization %s", orgID)
	}

	reqBody := claudeRequest{
		Model:     s.model,
		MaxTokens: s.maxTokens,
		Messages: []claudeMessage{
			{Role: "user", Content: prompt},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Claude request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, anthropicAPIURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create Claude request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", s.apiKey)
	httpReq.Header.Set("anthropic-version", anthropicAPIVersion)

	start := time.Now()
	httpResp, err := s.httpClient.Do(httpReq)
	latencyMs := int(time.Since(start).Milliseconds())

	if err != nil {
		s.logInteraction(ctx, orgID, interactionType, prompt, "", s.model, 0, 0, latencyMs, 0)
		return nil, fmt.Errorf("Claude API call failed: %w", err)
	}
	defer httpResp.Body.Close()

	respBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		s.logInteraction(ctx, orgID, interactionType, prompt, "", s.model, 0, 0, latencyMs, 0)
		return nil, fmt.Errorf("failed to read Claude response: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		s.logInteraction(ctx, orgID, interactionType, prompt, string(respBytes), s.model, 0, 0, latencyMs, 0)
		return nil, fmt.Errorf("Claude API returned status %d: %s", httpResp.StatusCode, string(respBytes))
	}

	var apiResp claudeAPIResponse
	if err := json.Unmarshal(respBytes, &apiResp); err != nil {
		s.logInteraction(ctx, orgID, interactionType, prompt, string(respBytes), s.model, 0, 0, latencyMs, 0)
		return nil, fmt.Errorf("failed to parse Claude response: %w", err)
	}

	var text string
	for _, block := range apiResp.Content {
		if block.Type == "text" {
			text += block.Text
		}
	}

	costEUR := (float64(apiResp.Usage.InputTokens)/1000)*inputTokenCostEUR +
		(float64(apiResp.Usage.OutputTokens)/1000)*outputTokenCostEUR

	s.logInteraction(ctx, orgID, interactionType, prompt, text, apiResp.Model, apiResp.Usage.InputTokens, apiResp.Usage.OutputTokens, latencyMs, costEUR)

	return &claudeResponse{
		Text:         text,
		Model:        apiResp.Model,
		InputTokens:  apiResp.Usage.InputTokens,
		OutputTokens: apiResp.Usage.OutputTokens,
		LatencyMs:    latencyMs,
		CostEUR:      costEUR,
	}, nil
}

// logInteraction persists an AI interaction to the database for audit and cost tracking.
func (s *AIService) logInteraction(ctx context.Context, orgID uuid.UUID, interactionType, prompt, response, model string, inputTokens, outputTokens, latencyMs int, costEUR float64) {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO ai_interaction_logs (
			organization_id, interaction_type, prompt_text, response_text,
			model, input_tokens, output_tokens, latency_ms, cost_eur
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, orgID, interactionType, prompt, response, model, inputTokens, outputTokens, latencyMs, costEUR)
	if err != nil {
		log.Error().Err(err).Str("interaction_type", interactionType).Msg("Failed to log AI interaction")
	}
}

// getLastInteractionID retrieves the most recent interaction ID for the given type.
func (s *AIService) getLastInteractionID(ctx context.Context, orgID uuid.UUID, interactionType string) (uuid.UUID, error) {
	var id uuid.UUID
	err := s.pool.QueryRow(ctx, `
		SELECT id FROM ai_interaction_logs
		WHERE organization_id = $1 AND interaction_type = $2
		ORDER BY created_at DESC
		LIMIT 1
	`, orgID, interactionType).Scan(&id)
	return id, err
}

// ============================================================
// PROMPT BUILDERS
// ============================================================

func buildRemediationPlanPrompt(req RemediationPlanRequest) string {
	var sb strings.Builder
	sb.WriteString(`You are a senior GRC (Governance, Risk, and Compliance) consultant with deep expertise in ISO 27001, NIST CSF, PCI DSS, GDPR, and other major compliance frameworks.

Generate a detailed compliance remediation plan based on the following inputs. Prioritize actions by risk impact, regulatory urgency, and cross-framework benefit.

`)
	sb.WriteString(fmt.Sprintf("Plan Type: %s\n", req.PlanType))
	sb.WriteString(fmt.Sprintf("Scope: %s\n", req.ScopeDescription))
	sb.WriteString(fmt.Sprintf("Risk Appetite: %s\n", req.RiskAppetite))
	sb.WriteString(fmt.Sprintf("Budget: EUR %.2f\n", req.BudgetEUR))
	sb.WriteString(fmt.Sprintf("Timeline: %d months\n", req.TimelineMonths))
	sb.WriteString(fmt.Sprintf("Industry: %s\n", req.IndustryContext))
	sb.WriteString(fmt.Sprintf("Regulatory Context: %s\n", req.RegulatoryContext))

	if len(req.AvailableSkills) > 0 {
		sb.WriteString(fmt.Sprintf("Available Skills: %s\n", strings.Join(req.AvailableSkills, ", ")))
	}
	if len(req.ExistingControls) > 0 {
		sb.WriteString(fmt.Sprintf("Existing Controls: %s\n", strings.Join(req.ExistingControls, ", ")))
	}

	if len(req.Gaps) > 0 {
		sb.WriteString("\nCompliance Gaps:\n")
		for i, gap := range req.Gaps {
			sb.WriteString(fmt.Sprintf("  %d. [%s] %s (%s) - Current: %s, Target: %s, Severity: %s\n",
				i+1, gap.ControlCode, gap.ControlTitle, gap.FrameworkCode,
				gap.CurrentStatus, gap.TargetStatus, gap.GapSeverity))
			if gap.Description != "" {
				sb.WriteString(fmt.Sprintf("     Description: %s\n", gap.Description))
			}
		}
	}

	sb.WriteString(`
Respond with JSON only (no markdown fences):
{
  "plan_name": "...",
  "plan_description": "...",
  "priority": "critical|high|medium|low",
  "confidence_score": 0.0-1.0,
  "actions": [
    {
      "title": "...",
      "description": "...",
      "action_type": "implement_control|update_policy|deploy_technical|conduct_training|perform_assessment|gather_evidence|review_process|third_party_engagement|documentation|monitoring_setup",
      "framework_control_code": "...",
      "priority": "critical|high|medium|low",
      "estimated_hours": 0.0,
      "estimated_cost_eur": 0.0,
      "required_skills": ["..."],
      "implementation_guidance": "...",
      "evidence_suggestions": ["..."],
      "tool_recommendations": ["..."],
      "risk_if_deferred": "...",
      "cross_framework_benefit": "...",
      "depends_on": [],
      "sort_order": 1
    }
  ],
  "estimated_total_hours": 0.0,
  "estimated_total_cost_eur": 0.0,
  "timeline_weeks": 0,
  "risk_summary": "...",
  "assumptions": ["..."]
}`)
	return sb.String()
}

func buildControlGuidancePrompt(req ControlGuidanceRequest) string {
	return fmt.Sprintf(`You are a senior GRC consultant. Provide detailed implementation guidance for the following control.

Control Code: %s
Control Title: %s
Framework: %s
Current Status: %s
Organization Size: %s
Industry: %s

Respond with JSON only (no markdown fences):
{
  "control_code": "%s",
  "implementation_steps": ["..."],
  "technical_measures": ["..."],
  "organizational_measures": ["..."],
  "evidence_required": ["..."],
  "common_pitfalls": ["..."],
  "maturity_indicators": {"initial": "...", "managed": "...", "optimized": "..."},
  "estimated_effort": "...",
  "related_controls": ["..."]
}`,
		req.ControlCode, req.ControlTitle, req.FrameworkCode,
		req.CurrentStatus, req.OrganizationSize, req.IndustryContext,
		req.ControlCode)
}

func buildGapImpactPrompt(gaps []ComplianceGap) string {
	var sb strings.Builder
	sb.WriteString(`You are a senior GRC consultant. Analyse the impact of the following compliance gaps and provide a prioritized remediation order.

Compliance Gaps:
`)
	for i, gap := range gaps {
		sb.WriteString(fmt.Sprintf("  %d. [%s] %s (%s) - Current: %s, Target: %s, Severity: %s\n",
			i+1, gap.ControlCode, gap.ControlTitle, gap.FrameworkCode,
			gap.CurrentStatus, gap.TargetStatus, gap.GapSeverity))
	}

	sb.WriteString(`
Respond with JSON only (no markdown fences):
{
  "overall_risk_level": "critical|high|medium|low",
  "summary": "...",
  "gap_assessments": [
    {
      "control_code": "...",
      "risk_level": "critical|high|medium|low",
      "business_impact": "...",
      "regulatory_risk": "...",
      "remediation_cost": "...",
      "recommendation": "..."
    }
  ],
  "prioritized_order": ["control_code_1", "control_code_2"],
  "quick_wins": ["..."],
  "strategic_items": ["..."]
}`)
	return sb.String()
}

func buildPolicyDraftPrompt(req PolicyDraftRequest) string {
	return fmt.Sprintf(`You are a senior GRC consultant specializing in policy development. Draft a policy section with the following parameters.

Policy Type: %s
Section Title: %s
Applicable Frameworks: %s
Industry Context: %s
Tone: %s

Respond with JSON only (no markdown fences):
{
  "section_title": "%s",
  "content": "...",
  "key_points": ["..."],
  "definitions": {"term": "definition"},
  "related_clauses": ["..."],
  "review_notes": ["..."]
}`,
		req.PolicyType, req.SectionTitle, strings.Join(req.FrameworkCodes, ", "),
		req.IndustryContext, req.TonePreference, req.SectionTitle)
}

func buildRiskNarrativePrompt(req RiskNarrativeRequest) string {
	controls := "None specified"
	if len(req.ExistingControls) > 0 {
		controls = strings.Join(req.ExistingControls, ", ")
	}

	return fmt.Sprintf(`You are a senior risk management consultant. Generate a comprehensive risk narrative for the following risk.

Risk Title: %s
Description: %s
Category: %s
Likelihood: %d/5
Impact: %d/5
Existing Controls: %s
Industry: %s

Respond with JSON only (no markdown fences):
{
  "executive_summary": "...",
  "detailed_analysis": "...",
  "threat_scenarios": ["..."],
  "mitigation_options": ["..."],
  "residual_risk_note": "...",
  "monitoring_advice": "..."
}`,
		req.RiskTitle, req.RiskDescription, req.RiskCategory,
		req.Likelihood, req.Impact, controls, req.IndustryContext)
}

// ============================================================
// RESPONSE PARSER HELPER
// ============================================================

// extractJSONFromResponse attempts to extract JSON from the AI response text.
// It handles responses that might include markdown code fences.
func extractJSONFromResponse(text string, v interface{}) error {
	text = strings.TrimSpace(text)

	// Try direct parse first
	if err := json.Unmarshal([]byte(text), v); err == nil {
		return nil
	}

	// Try extracting from markdown code fences
	if idx := strings.Index(text, "```json"); idx >= 0 {
		text = text[idx+7:]
		if endIdx := strings.Index(text, "```"); endIdx >= 0 {
			text = text[:endIdx]
		}
	} else if idx := strings.Index(text, "```"); idx >= 0 {
		text = text[idx+3:]
		if endIdx := strings.Index(text, "```"); endIdx >= 0 {
			text = text[:endIdx]
		}
	}

	text = strings.TrimSpace(text)

	// Try to find JSON object boundaries
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start >= 0 && end > start {
		text = text[start : end+1]
	}

	if err := json.Unmarshal([]byte(text), v); err != nil {
		return fmt.Errorf("failed to extract JSON from AI response: %w", err)
	}

	return nil
}

// ============================================================
// FALLBACK METHODS (when AI is unavailable)
// ============================================================

func (s *AIService) fallbackRemediationPlan(req RemediationPlanRequest) (*RemediationPlanResponse, error) {
	actions := make([]RemediationActionSuggestion, 0, len(req.Gaps))

	for i, gap := range req.Gaps {
		priority := "medium"
		hours := 16.0
		cost := 2000.0
		switch gap.GapSeverity {
		case "critical":
			priority = "critical"
			hours = 40.0
			cost = 5000.0
		case "high":
			priority = "high"
			hours = 24.0
			cost = 3000.0
		case "low":
			priority = "low"
			hours = 8.0
			cost = 1000.0
		}

		guidance := ""
		key := gap.FrameworkCode + "-" + gap.ControlCode
		if g, ok := staticControlGuidance[key]; ok {
			guidance = strings.Join(g.ImplementationSteps, "; ")
		} else {
			guidance = fmt.Sprintf("Review control %s requirements and implement necessary measures to move from %s to %s status.",
				gap.ControlCode, gap.CurrentStatus, gap.TargetStatus)
		}

		actions = append(actions, RemediationActionSuggestion{
			Title:                  fmt.Sprintf("Remediate %s: %s", gap.ControlCode, gap.ControlTitle),
			Description:            fmt.Sprintf("Address gap in %s - move from %s to %s", gap.ControlCode, gap.CurrentStatus, gap.TargetStatus),
			ActionType:             "implement_control",
			FrameworkControlCode:   gap.ControlCode,
			Priority:               priority,
			EstimatedHours:         hours,
			EstimatedCostEUR:       cost,
			RequiredSkills:         []string{"compliance", "information security"},
			ImplementationGuidance: guidance,
			EvidenceSuggestions:    []string{"Implementation evidence", "Test results", "Configuration screenshots"},
			ToolRecommendations:    []string{"GRC platform", "Document management system"},
			RiskIfDeferred:         fmt.Sprintf("Continued non-compliance with %s requirements for %s", gap.FrameworkCode, gap.ControlCode),
			CrossFrameworkBenefit:  "May address related requirements across multiple frameworks",
			SortOrder:              i + 1,
		})
	}

	var totalHours, totalCost float64
	for _, a := range actions {
		totalHours += a.EstimatedHours
		totalCost += a.EstimatedCostEUR
	}

	overallPriority := "medium"
	for _, gap := range req.Gaps {
		if gap.GapSeverity == "critical" {
			overallPriority = "critical"
			break
		}
		if gap.GapSeverity == "high" {
			overallPriority = "high"
		}
	}

	return &RemediationPlanResponse{
		PlanName:        fmt.Sprintf("Remediation Plan - %s", req.PlanType),
		PlanDescription: fmt.Sprintf("Rule-based remediation plan addressing %d compliance gaps. AI-enhanced analysis was unavailable; this plan uses standard prioritization.", len(req.Gaps)),
		Priority:        overallPriority,
		ConfidenceScore: 0.60,
		Actions:         actions,
		EstimatedHours:  totalHours,
		EstimatedCost:   totalCost,
		TimelineWeeks:   req.TimelineMonths * 4,
		RiskSummary:     fmt.Sprintf("Plan addresses %d identified gaps. Manual review recommended for prioritization accuracy.", len(req.Gaps)),
		Assumptions:     []string{"Standard effort estimates used", "AI analysis unavailable - rule-based fallback applied", "Manual review strongly recommended"},
	}, nil
}

func (s *AIService) fallbackControlGuidance(req ControlGuidanceRequest) (*ControlGuidance, error) {
	key := req.FrameworkCode + "-" + req.ControlCode
	if g, ok := staticControlGuidance[key]; ok {
		return &g, nil
	}

	return &ControlGuidance{
		ControlCode:         req.ControlCode,
		ImplementationSteps: []string{"Review control requirements", "Assess current state", "Develop implementation plan", "Implement controls", "Validate effectiveness"},
		TechnicalMeasures:   []string{"Implement appropriate technical controls as required by the framework"},
		OrganizationalMeasures: []string{"Document processes and procedures", "Assign responsibilities", "Establish review cycles"},
		EvidenceRequired:    []string{"Policy documents", "Configuration evidence", "Test results", "Review records"},
		CommonPitfalls:      []string{"Insufficient documentation", "Lack of regular reviews", "No evidence of effectiveness"},
		MaturityIndicators:  map[string]string{"initial": "Control exists informally", "managed": "Control documented and monitored", "optimized": "Control continuously improved"},
		EstimatedEffort:     "2-6 weeks depending on complexity",
		RelatedControls:     []string{},
	}, nil
}

func (s *AIService) fallbackGapAnalysis(gaps []ComplianceGap) *GapImpactAnalysis {
	assessments := make([]GapAssessmentItem, 0, len(gaps))
	order := make([]string, 0, len(gaps))
	var quickWins, strategic []string

	overallRisk := "medium"
	for _, gap := range gaps {
		risk := gap.GapSeverity
		if risk == "critical" {
			overallRisk = "critical"
		} else if risk == "high" && overallRisk != "critical" {
			overallRisk = "high"
		}

		assessments = append(assessments, GapAssessmentItem{
			ControlCode:     gap.ControlCode,
			RiskLevel:       gap.GapSeverity,
			BusinessImpact:  fmt.Sprintf("Gap in %s may expose the organization to compliance risk", gap.ControlCode),
			RegulatoryRisk:  fmt.Sprintf("Non-compliance with %s framework requirements", gap.FrameworkCode),
			RemediationCost: "Moderate - requires assessment",
			Recommendation:  fmt.Sprintf("Implement %s to move from %s to %s", gap.ControlTitle, gap.CurrentStatus, gap.TargetStatus),
		})
		order = append(order, gap.ControlCode)

		if gap.GapSeverity == "low" || gap.CurrentStatus == "partial" {
			quickWins = append(quickWins, gap.ControlCode)
		} else if gap.GapSeverity == "critical" || gap.GapSeverity == "high" {
			strategic = append(strategic, gap.ControlCode)
		}
	}

	return &GapImpactAnalysis{
		OverallRiskLevel: overallRisk,
		Summary:          fmt.Sprintf("Analysis of %d compliance gaps using rule-based assessment. AI-enhanced analysis was unavailable.", len(gaps)),
		GapAssessments:   assessments,
		PrioritizedOrder: order,
		QuickWins:        quickWins,
		StrategicItems:   strategic,
	}
}

func (s *AIService) fallbackEvidenceTemplate(controlCode, controlTitle string) *EvidenceTemplate {
	return &EvidenceTemplate{
		ControlCode:  controlCode,
		ControlTitle: controlTitle,
		EvidenceTypes: []EvidenceTypeItem{
			{Type: "Policy Document", Description: "Documented policy covering this control", Format: "PDF/DOCX", Example: "Information Security Policy v2.0"},
			{Type: "Configuration Evidence", Description: "Screenshots or exports showing control configuration", Format: "PNG/PDF", Example: "System hardening configuration export"},
			{Type: "Test Results", Description: "Results from control effectiveness testing", Format: "PDF/CSV", Example: "Penetration test report or audit findings"},
			{Type: "Review Records", Description: "Meeting minutes or sign-off records", Format: "PDF/DOCX", Example: "Quarterly review meeting minutes"},
		},
		CollectionTips: []string{
			"Collect evidence at the time of implementation",
			"Ensure evidence is dated and attributable",
			"Store evidence in a centralized, access-controlled location",
			"Maintain a clear chain of custody",
		},
		ReviewFrequency: "Quarterly or as defined by organizational policy",
	}
}
