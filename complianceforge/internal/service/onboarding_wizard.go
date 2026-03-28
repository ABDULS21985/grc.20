package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// ============================================================
// ONBOARDING WIZARD SERVICE
// Manages the 7-step self-service onboarding wizard for new
// organisations in ComplianceForge.
// ============================================================

// OnboardingWizard manages the multi-step onboarding process.
type OnboardingWizard struct {
	pool *pgxpool.Pool
}

// NewOnboardingWizard creates a new OnboardingWizard service.
func NewOnboardingWizard(pool *pgxpool.Pool) *OnboardingWizard {
	return &OnboardingWizard{pool: pool}
}

// ============================================================
// DATA TYPES
// ============================================================

// OnboardingProgress represents the current state of an organisation's onboarding.
type OnboardingProgress struct {
	ID                     uuid.UUID              `json:"id"`
	OrganizationID         uuid.UUID              `json:"organization_id"`
	CurrentStep            int                    `json:"current_step"`
	TotalSteps             int                    `json:"total_steps"`
	CompletedSteps         []int                  `json:"completed_steps"`
	IsCompleted            bool                   `json:"is_completed"`
	CompletedAt            *time.Time             `json:"completed_at,omitempty"`
	SkippedSteps           []int                  `json:"skipped_steps"`
	OrgProfileData         map[string]interface{} `json:"org_profile_data"`
	IndustryAssessmentData map[string]interface{} `json:"industry_assessment_data"`
	SelectedFrameworkIDs   []uuid.UUID            `json:"selected_framework_ids"`
	TeamInvitations        []TeamInvitation       `json:"team_invitations"`
	RiskAppetiteData       map[string]interface{} `json:"risk_appetite_data"`
	QuickAssessmentData    map[string]interface{} `json:"quick_assessment_data"`
	CreatedAt              time.Time              `json:"created_at"`
	UpdatedAt              time.Time              `json:"updated_at"`
}

// OrgProfileInput captures organisation profile data in step 1.
type OrgProfileInput struct {
	DisplayName        string `json:"display_name"`
	LegalName          string `json:"legal_name"`
	Industry           string `json:"industry"`
	SubIndustry        string `json:"sub_industry"`
	CountryCode        string `json:"country_code"`
	Timezone           string `json:"timezone"`
	EmployeeCountRange string `json:"employee_count_range"`
	Website            string `json:"website"`
	PrimaryContact     string `json:"primary_contact"`
	Phone              string `json:"phone"`
	Address            string `json:"address"`
}

// IndustryAssessmentInput captures answers for the industry assessment questionnaire.
type IndustryAssessmentInput struct {
	ProcessPaymentCards  bool `json:"process_payment_cards"`
	HandleEUPersonalData bool `json:"handle_eu_personal_data"`
	EssentialServices    bool `json:"essential_services"`
	UKPublicSector       bool `json:"uk_public_sector"`
	ISOCertification     bool `json:"iso_certification"`
	USFederalContracts   bool `json:"us_federal_contracts"`
	CyberMaturity        bool `json:"cyber_maturity"`
	ITILRequirements     bool `json:"itil_requirements"`
	BoardITGovernance    bool `json:"board_it_governance"`
	DataResidencyEU      bool `json:"data_residency_eu"`
	CloudFirst           bool `json:"cloud_first"`
	RegulatedIndustry    bool `json:"regulated_industry"`
}

// FrameworkRecommendation represents a recommended framework with reason.
type FrameworkRecommendation struct {
	FrameworkID   uuid.UUID `json:"framework_id"`
	FrameworkCode string    `json:"framework_code"`
	FrameworkName string    `json:"framework_name"`
	Reason        string    `json:"reason"`
	Priority      string    `json:"priority"` // "required", "recommended", "optional"
	Description   string    `json:"description"`
}

// TeamInvitation represents an invitation to send to a team member.
type TeamInvitation struct {
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Role      string `json:"role"`
	JobTitle  string `json:"job_title"`
}

// RiskAppetiteInput captures the organisation's risk appetite configuration.
type RiskAppetiteInput struct {
	OverallAppetite     string                 `json:"overall_appetite"`     // conservative, moderate, aggressive
	RiskTolerance       map[string]interface{} `json:"risk_tolerance"`       // per-category tolerances
	AcceptableRiskLevel string                 `json:"acceptable_risk_level"` // low, medium, high
	ReviewFrequency     string                 `json:"review_frequency"`     // monthly, quarterly, annually
	EscalationThreshold string                 `json:"escalation_threshold"`
	Notes               string                 `json:"notes"`
}

// QuickAssessmentInput captures a rapid initial self-assessment.
type QuickAssessmentInput struct {
	SecurityMaturity       int                    `json:"security_maturity"`        // 1-5
	PolicyDocumentation    int                    `json:"policy_documentation"`     // 1-5
	IncidentResponseReady  int                    `json:"incident_response_ready"`  // 1-5
	DataProtectionLevel    int                    `json:"data_protection_level"`    // 1-5
	AccessControlMaturity  int                    `json:"access_control_maturity"`  // 1-5
	ThirdPartyRiskMgmt     int                    `json:"third_party_risk_mgmt"`    // 1-5
	BusinessContinuity     int                    `json:"business_continuity"`      // 1-5
	SecurityAwareness      int                    `json:"security_awareness"`       // 1-5
	AdditionalNotes        string                 `json:"additional_notes"`
	PriorityAreas          []string               `json:"priority_areas"`
	ExistingCertifications []string               `json:"existing_certifications"`
	CustomAnswers          map[string]interface{} `json:"custom_answers"`
}

// OnboardingCompletionResult contains everything created during final onboarding.
type OnboardingCompletionResult struct {
	OrganizationID    uuid.UUID                `json:"organization_id"`
	FrameworksAdopted []AdoptedFrameworkResult  `json:"frameworks_adopted"`
	InvitationsSent   int                      `json:"invitations_sent"`
	RiskMatrixCreated bool                     `json:"risk_matrix_created"`
	ControlsCreated   int                      `json:"controls_created"`
	CompletedAt       time.Time                `json:"completed_at"`
}

// ============================================================
// STEP 1: ORGANISATION PROFILE
// ============================================================

// SaveOrgProfile saves or updates the organisation profile data (Step 1).
func (w *OnboardingWizard) SaveOrgProfile(ctx context.Context, orgID uuid.UUID, data OrgProfileInput) (*OnboardingProgress, error) {
	log.Info().Str("org_id", orgID.String()).Msg("Onboarding wizard: saving org profile (step 1)")

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal org profile data: %w", err)
	}

	progress, err := w.ensureProgress(ctx, orgID)
	if err != nil {
		return nil, err
	}

	_, err = w.pool.Exec(ctx, `
		UPDATE onboarding_progress
		SET org_profile_data = $1,
		    current_step = GREATEST(current_step, 2),
		    completed_steps = (
		        SELECT jsonb_agg(DISTINCT val)
		        FROM jsonb_array_elements(completed_steps || '1'::jsonb) AS val
		    ),
		    updated_at = NOW()
		WHERE organization_id = $2`,
		dataJSON, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to save org profile: %w", err)
	}

	// Also update the organisations table with the profile data
	_, err = w.pool.Exec(ctx, `
		UPDATE organizations
		SET name = COALESCE(NULLIF($1, ''), name),
		    legal_name = COALESCE(NULLIF($2, ''), legal_name),
		    industry = COALESCE(NULLIF($3, ''), industry),
		    country_code = COALESCE(NULLIF($4, ''), country_code),
		    timezone = COALESCE(NULLIF($5, ''), timezone),
		    employee_count_range = COALESCE(NULLIF($6, ''), employee_count_range),
		    updated_at = NOW()
		WHERE id = $7`,
		data.DisplayName, data.LegalName, data.Industry,
		data.CountryCode, data.Timezone, data.EmployeeCountRange, orgID,
	)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to update organization with profile data")
	}

	_ = progress // used for ensureProgress side effect
	return w.GetProgress(ctx, orgID)
}

// ============================================================
// STEP 2: INDUSTRY ASSESSMENT
// ============================================================

// SaveIndustryAssessment saves the industry assessment answers (Step 2).
func (w *OnboardingWizard) SaveIndustryAssessment(ctx context.Context, orgID uuid.UUID, data IndustryAssessmentInput) (*OnboardingProgress, error) {
	log.Info().Str("org_id", orgID.String()).Msg("Onboarding wizard: saving industry assessment (step 2)")

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal assessment data: %w", err)
	}

	_, err = w.ensureProgress(ctx, orgID)
	if err != nil {
		return nil, err
	}

	_, err = w.pool.Exec(ctx, `
		UPDATE onboarding_progress
		SET industry_assessment_data = $1,
		    current_step = GREATEST(current_step, 3),
		    completed_steps = (
		        SELECT jsonb_agg(DISTINCT val)
		        FROM jsonb_array_elements(completed_steps || '2'::jsonb) AS val
		    ),
		    updated_at = NOW()
		WHERE organization_id = $2`,
		dataJSON, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to save industry assessment: %w", err)
	}

	return w.GetProgress(ctx, orgID)
}

// GetFrameworkRecommendations returns recommended frameworks based on assessment answers.
func (w *OnboardingWizard) GetFrameworkRecommendations(ctx context.Context, answers IndustryAssessmentInput) ([]FrameworkRecommendation, error) {
	log.Info().Msg("Onboarding wizard: generating framework recommendations")

	// Build the list of framework codes to recommend based on answers
	type recommendation struct {
		Code     string
		Reason   string
		Priority string
	}
	var recs []recommendation

	if answers.ProcessPaymentCards {
		recs = append(recs, recommendation{
			Code:     "PCI_DSS_4",
			Reason:   "Your organisation processes payment card data, requiring PCI DSS compliance.",
			Priority: "required",
		})
	}

	if answers.HandleEUPersonalData {
		recs = append(recs, recommendation{
			Code:     "UK_GDPR",
			Reason:   "Your organisation handles EU/UK personal data, requiring GDPR compliance.",
			Priority: "required",
		})
	}

	if answers.EssentialServices {
		recs = append(recs, recommendation{
			Code:     "NCSC_CAF",
			Reason:   "As an essential services provider, NCSC Cyber Assessment Framework is required.",
			Priority: "required",
		})
		recs = append(recs, recommendation{
			Code:     "NIS2",
			Reason:   "Essential services fall under NIS2 Directive requirements.",
			Priority: "required",
		})
	}

	if answers.UKPublicSector {
		recs = append(recs, recommendation{
			Code:     "CYBER_ESSENTIALS",
			Reason:   "UK public sector organisations must achieve Cyber Essentials certification.",
			Priority: "required",
		})
	}

	if answers.ISOCertification {
		recs = append(recs, recommendation{
			Code:     "ISO27001",
			Reason:   "You indicated interest in ISO certification. ISO 27001 is the gold standard for information security management.",
			Priority: "recommended",
		})
	}

	if answers.USFederalContracts {
		recs = append(recs, recommendation{
			Code:     "NIST_800_53",
			Reason:   "US federal contracts require NIST 800-53 compliance.",
			Priority: "required",
		})
	}

	if answers.CyberMaturity {
		recs = append(recs, recommendation{
			Code:     "NIST_CSF_2",
			Reason:   "NIST Cybersecurity Framework 2.0 provides an excellent maturity model for cybersecurity programmes.",
			Priority: "recommended",
		})
	}

	if answers.ITILRequirements {
		recs = append(recs, recommendation{
			Code:     "ITIL_4",
			Reason:   "ITIL 4 practices align with your IT service management requirements.",
			Priority: "recommended",
		})
	}

	if answers.BoardITGovernance {
		recs = append(recs, recommendation{
			Code:     "COBIT_2019",
			Reason:   "COBIT 2019 provides comprehensive IT governance for board-level oversight.",
			Priority: "recommended",
		})
	}

	// If no answers produced recommendations, suggest defaults
	if len(recs) == 0 {
		recs = append(recs, recommendation{
			Code:     "NIST_CSF_2",
			Reason:   "NIST CSF 2.0 is a great starting point for building your cybersecurity programme.",
			Priority: "recommended",
		})
		recs = append(recs, recommendation{
			Code:     "ISO27001",
			Reason:   "ISO 27001 provides a comprehensive framework for information security management.",
			Priority: "recommended",
		})
	}

	// Look up framework details from the database
	var results []FrameworkRecommendation

	// Deduplicate by code
	seen := make(map[string]bool)
	var uniqueRecs []recommendation
	for _, r := range recs {
		if !seen[r.Code] {
			seen[r.Code] = true
			uniqueRecs = append(uniqueRecs, r)
		}
	}

	// Gather codes for query
	codes := make([]string, len(uniqueRecs))
	recMap := make(map[string]recommendation)
	for i, r := range uniqueRecs {
		codes[i] = r.Code
		recMap[r.Code] = r
	}

	rows, err := w.pool.Query(ctx, `
		SELECT id, code, name, description
		FROM compliance_frameworks
		WHERE code = ANY($1)
		ORDER BY name`,
		codes,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query frameworks: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var fw struct {
			ID          uuid.UUID
			Code        string
			Name        string
			Description string
		}
		if err := rows.Scan(&fw.ID, &fw.Code, &fw.Name, &fw.Description); err != nil {
			continue
		}
		rec := recMap[fw.Code]
		results = append(results, FrameworkRecommendation{
			FrameworkID:   fw.ID,
			FrameworkCode: fw.Code,
			FrameworkName: fw.Name,
			Reason:        rec.Reason,
			Priority:      rec.Priority,
			Description:   fw.Description,
		})
	}

	// Add any codes that were not in the DB as "informational" placeholders
	foundCodes := make(map[string]bool)
	for _, r := range results {
		foundCodes[r.FrameworkCode] = true
	}
	for _, r := range uniqueRecs {
		if !foundCodes[r.Code] {
			results = append(results, FrameworkRecommendation{
				FrameworkCode: r.Code,
				FrameworkName: formatFrameworkName(r.Code),
				Reason:        r.Reason,
				Priority:      r.Priority,
				Description:   "Framework not yet loaded in database.",
			})
		}
	}

	log.Info().Int("recommendations", len(results)).Msg("Framework recommendations generated")
	return results, nil
}

// ============================================================
// STEP 3: FRAMEWORK SELECTION
// ============================================================

// SaveFrameworkSelection saves the selected framework IDs (Step 3).
func (w *OnboardingWizard) SaveFrameworkSelection(ctx context.Context, orgID uuid.UUID, frameworkIDs []uuid.UUID) (*OnboardingProgress, error) {
	log.Info().Str("org_id", orgID.String()).Int("count", len(frameworkIDs)).Msg("Onboarding wizard: saving framework selection (step 3)")

	_, err := w.ensureProgress(ctx, orgID)
	if err != nil {
		return nil, err
	}

	_, err = w.pool.Exec(ctx, `
		UPDATE onboarding_progress
		SET selected_framework_ids = $1,
		    current_step = GREATEST(current_step, 4),
		    completed_steps = (
		        SELECT jsonb_agg(DISTINCT val)
		        FROM jsonb_array_elements(completed_steps || '3'::jsonb) AS val
		    ),
		    updated_at = NOW()
		WHERE organization_id = $2`,
		frameworkIDs, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to save framework selection: %w", err)
	}

	return w.GetProgress(ctx, orgID)
}

// ============================================================
// STEP 4: TEAM SETUP
// ============================================================

// SaveTeamInvitations saves team invitation data (Step 4).
func (w *OnboardingWizard) SaveTeamInvitations(ctx context.Context, orgID uuid.UUID, invitations []TeamInvitation) (*OnboardingProgress, error) {
	log.Info().Str("org_id", orgID.String()).Int("invitations", len(invitations)).Msg("Onboarding wizard: saving team invitations (step 4)")

	invJSON, err := json.Marshal(invitations)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal invitations: %w", err)
	}

	_, err = w.ensureProgress(ctx, orgID)
	if err != nil {
		return nil, err
	}

	_, err = w.pool.Exec(ctx, `
		UPDATE onboarding_progress
		SET team_invitations = $1,
		    current_step = GREATEST(current_step, 5),
		    completed_steps = (
		        SELECT jsonb_agg(DISTINCT val)
		        FROM jsonb_array_elements(completed_steps || '4'::jsonb) AS val
		    ),
		    updated_at = NOW()
		WHERE organization_id = $2`,
		invJSON, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to save team invitations: %w", err)
	}

	return w.GetProgress(ctx, orgID)
}

// ============================================================
// STEP 5: RISK APPETITE
// ============================================================

// SaveRiskAppetite saves the risk appetite configuration (Step 5).
func (w *OnboardingWizard) SaveRiskAppetite(ctx context.Context, orgID uuid.UUID, data RiskAppetiteInput) (*OnboardingProgress, error) {
	log.Info().Str("org_id", orgID.String()).Msg("Onboarding wizard: saving risk appetite (step 5)")

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal risk appetite data: %w", err)
	}

	_, err = w.ensureProgress(ctx, orgID)
	if err != nil {
		return nil, err
	}

	_, err = w.pool.Exec(ctx, `
		UPDATE onboarding_progress
		SET risk_appetite_data = $1,
		    current_step = GREATEST(current_step, 6),
		    completed_steps = (
		        SELECT jsonb_agg(DISTINCT val)
		        FROM jsonb_array_elements(completed_steps || '5'::jsonb) AS val
		    ),
		    updated_at = NOW()
		WHERE organization_id = $2`,
		dataJSON, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to save risk appetite: %w", err)
	}

	return w.GetProgress(ctx, orgID)
}

// ============================================================
// STEP 6: QUICK ASSESSMENT
// ============================================================

// SaveQuickAssessment saves the quick self-assessment data (Step 6).
func (w *OnboardingWizard) SaveQuickAssessment(ctx context.Context, orgID uuid.UUID, data QuickAssessmentInput) (*OnboardingProgress, error) {
	log.Info().Str("org_id", orgID.String()).Msg("Onboarding wizard: saving quick assessment (step 6)")

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal assessment data: %w", err)
	}

	_, err = w.ensureProgress(ctx, orgID)
	if err != nil {
		return nil, err
	}

	_, err = w.pool.Exec(ctx, `
		UPDATE onboarding_progress
		SET quick_assessment_data = $1,
		    current_step = GREATEST(current_step, 7),
		    completed_steps = (
		        SELECT jsonb_agg(DISTINCT val)
		        FROM jsonb_array_elements(completed_steps || '6'::jsonb) AS val
		    ),
		    updated_at = NOW()
		WHERE organization_id = $2`,
		dataJSON, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to save quick assessment: %w", err)
	}

	return w.GetProgress(ctx, orgID)
}

// ============================================================
// STEP 7: COMPLETE ONBOARDING
// ============================================================

// CompleteOnboarding finalises the onboarding process in an atomic transaction.
// It adopts selected frameworks, sends team invitations, and creates
// initial resources based on all collected wizard data.
func (w *OnboardingWizard) CompleteOnboarding(ctx context.Context, orgID uuid.UUID) (*OnboardingCompletionResult, error) {
	log.Info().Str("org_id", orgID.String()).Msg("Onboarding wizard: completing onboarding (step 7)")

	// Fetch current progress
	progress, err := w.GetProgress(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get onboarding progress: %w", err)
	}

	if progress.IsCompleted {
		return nil, fmt.Errorf("onboarding already completed for this organisation")
	}

	// Start atomic transaction
	tx, err := w.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	result := &OnboardingCompletionResult{
		OrganizationID: orgID,
		CompletedAt:    time.Now(),
	}

	// ── Adopt selected frameworks ──────────────────────
	if len(progress.SelectedFrameworkIDs) > 0 {
		// Get admin user for this org
		var adminUserID uuid.UUID
		err = tx.QueryRow(ctx, `
			SELECT u.id FROM users u
			JOIN user_roles ur ON ur.user_id = u.id
			JOIN roles r ON r.id = ur.role_id
			WHERE u.organization_id = $1 AND r.slug = 'org_admin'
			ORDER BY u.created_at ASC
			LIMIT 1`,
			orgID,
		).Scan(&adminUserID)
		if err != nil {
			log.Warn().Err(err).Msg("No admin user found, using nil UUID for framework adoption")
			adminUserID = uuid.Nil
		}

		totalControls := 0
		for _, fwID := range progress.SelectedFrameworkIDs {
			// Check if already adopted
			var existingCount int
			err = tx.QueryRow(ctx, `
				SELECT COUNT(*) FROM organization_frameworks
				WHERE organization_id = $1 AND framework_id = $2`,
				orgID, fwID,
			).Scan(&existingCount)
			if err == nil && existingCount > 0 {
				continue // Skip already adopted frameworks
			}

			adopted, err := w.adoptFrameworkTx(ctx, tx, orgID, adminUserID, fwID)
			if err != nil {
				log.Warn().Err(err).Str("framework_id", fwID.String()).Msg("Failed to adopt framework during onboarding completion")
				continue
			}
			result.FrameworksAdopted = append(result.FrameworksAdopted, *adopted)
			totalControls += adopted.TotalControls
		}
		result.ControlsCreated = totalControls
	}

	// ── Create risk matrix if risk appetite data provided ──
	if progress.RiskAppetiteData != nil && len(progress.RiskAppetiteData) > 0 {
		var matrixExists int
		err = tx.QueryRow(ctx, `
			SELECT COUNT(*) FROM risk_matrices
			WHERE organization_id = $1 AND is_default = true`, orgID,
		).Scan(&matrixExists)
		if err == nil && matrixExists == 0 {
			_, err = w.createDefaultRiskMatrix(ctx, tx, orgID)
			if err != nil {
				log.Warn().Err(err).Msg("Failed to create default risk matrix during onboarding")
			} else {
				result.RiskMatrixCreated = true
			}
		} else if matrixExists > 0 {
			result.RiskMatrixCreated = true // Already exists
		}
	}

	// ── Process team invitations ──────────────────────
	if len(progress.TeamInvitations) > 0 {
		for _, inv := range progress.TeamInvitations {
			err = w.createPendingInvitation(ctx, tx, orgID, inv)
			if err != nil {
				log.Warn().Err(err).Str("email", inv.Email).Msg("Failed to create invitation")
				continue
			}
			result.InvitationsSent++
		}
	}

	// ── Mark onboarding as complete ──────────────────
	now := time.Now()
	_, err = tx.Exec(ctx, `
		UPDATE onboarding_progress
		SET is_completed = true,
		    completed_at = $1,
		    current_step = total_steps,
		    completed_steps = (
		        SELECT jsonb_agg(DISTINCT val)
		        FROM jsonb_array_elements(completed_steps || '7'::jsonb) AS val
		    ),
		    updated_at = $1
		WHERE organization_id = $2`,
		now, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to mark onboarding as complete: %w", err)
	}

	// ── Update org status to active if still in trial ─
	_, err = tx.Exec(ctx, `
		UPDATE organizations
		SET status = 'active', updated_at = NOW()
		WHERE id = $1 AND status = 'trial'`, orgID,
	)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to update org status to active")
	}

	// ── Commit transaction ──────────────────────────
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit onboarding completion: %w", err)
	}

	log.Info().
		Str("org_id", orgID.String()).
		Int("frameworks", len(result.FrameworksAdopted)).
		Int("controls", result.ControlsCreated).
		Int("invitations", result.InvitationsSent).
		Msg("Onboarding wizard completed successfully")

	return result, nil
}

// ============================================================
// PROGRESS MANAGEMENT
// ============================================================

// GetProgress returns the current onboarding progress for an organisation.
func (w *OnboardingWizard) GetProgress(ctx context.Context, orgID uuid.UUID) (*OnboardingProgress, error) {
	progress := &OnboardingProgress{}

	var completedStepsJSON []byte
	var invitationsJSON []byte
	var orgProfileJSON []byte
	var assessmentJSON []byte
	var riskAppetiteJSON []byte
	var quickAssessmentJSON []byte

	err := w.pool.QueryRow(ctx, `
		SELECT id, organization_id, current_step, total_steps,
		       completed_steps, is_completed, completed_at,
		       skipped_steps, org_profile_data, industry_assessment_data,
		       selected_framework_ids, team_invitations,
		       risk_appetite_data, quick_assessment_data,
		       created_at, updated_at
		FROM onboarding_progress
		WHERE organization_id = $1`,
		orgID,
	).Scan(
		&progress.ID, &progress.OrganizationID,
		&progress.CurrentStep, &progress.TotalSteps,
		&completedStepsJSON, &progress.IsCompleted, &progress.CompletedAt,
		&progress.SkippedSteps, &orgProfileJSON, &assessmentJSON,
		&progress.SelectedFrameworkIDs, &invitationsJSON,
		&riskAppetiteJSON, &quickAssessmentJSON,
		&progress.CreatedAt, &progress.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			// Return empty progress for org that hasn't started
			return &OnboardingProgress{
				OrganizationID: orgID,
				CurrentStep:    1,
				TotalSteps:     7,
				CompletedSteps: []int{},
				SkippedSteps:   []int{},
			}, nil
		}
		return nil, fmt.Errorf("failed to get onboarding progress: %w", err)
	}

	// Parse JSON fields
	if len(completedStepsJSON) > 0 {
		json.Unmarshal(completedStepsJSON, &progress.CompletedSteps)
	}
	if progress.CompletedSteps == nil {
		progress.CompletedSteps = []int{}
	}

	if len(invitationsJSON) > 0 {
		json.Unmarshal(invitationsJSON, &progress.TeamInvitations)
	}

	if len(orgProfileJSON) > 0 {
		json.Unmarshal(orgProfileJSON, &progress.OrgProfileData)
	}
	if len(assessmentJSON) > 0 {
		json.Unmarshal(assessmentJSON, &progress.IndustryAssessmentData)
	}
	if len(riskAppetiteJSON) > 0 {
		json.Unmarshal(riskAppetiteJSON, &progress.RiskAppetiteData)
	}
	if len(quickAssessmentJSON) > 0 {
		json.Unmarshal(quickAssessmentJSON, &progress.QuickAssessmentData)
	}

	if progress.SkippedSteps == nil {
		progress.SkippedSteps = []int{}
	}
	if progress.SelectedFrameworkIDs == nil {
		progress.SelectedFrameworkIDs = []uuid.UUID{}
	}

	return progress, nil
}

// SkipStep marks a step as skipped without saving data.
func (w *OnboardingWizard) SkipStep(ctx context.Context, orgID uuid.UUID, step int) (*OnboardingProgress, error) {
	if step < 1 || step > 7 {
		return nil, fmt.Errorf("invalid step number: %d (must be 1-7)", step)
	}
	if step == 7 {
		return nil, fmt.Errorf("cannot skip the final completion step")
	}

	log.Info().Str("org_id", orgID.String()).Int("step", step).Msg("Onboarding wizard: skipping step")

	_, err := w.ensureProgress(ctx, orgID)
	if err != nil {
		return nil, err
	}

	_, err = w.pool.Exec(ctx, `
		UPDATE onboarding_progress
		SET current_step = GREATEST(current_step, $1 + 1),
		    skipped_steps = array_append(
		        CASE WHEN $1 = ANY(skipped_steps) THEN skipped_steps
		             ELSE skipped_steps
		        END,
		        CASE WHEN $1 = ANY(skipped_steps) THEN NULL
		             ELSE $1
		        END
		    ),
		    completed_steps = (
		        SELECT jsonb_agg(DISTINCT val)
		        FROM jsonb_array_elements(completed_steps || to_jsonb($1)) AS val
		    ),
		    updated_at = NOW()
		WHERE organization_id = $2`,
		step, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to skip step: %w", err)
	}

	return w.GetProgress(ctx, orgID)
}

// ============================================================
// INTERNAL HELPERS
// ============================================================

// ensureProgress creates the onboarding_progress record if it doesn't exist.
func (w *OnboardingWizard) ensureProgress(ctx context.Context, orgID uuid.UUID) (*OnboardingProgress, error) {
	_, err := w.pool.Exec(ctx, `
		INSERT INTO onboarding_progress (organization_id, current_step, total_steps, completed_steps, skipped_steps)
		VALUES ($1, 1, 7, '[]', '{}')
		ON CONFLICT (organization_id) DO NOTHING`,
		orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure onboarding progress: %w", err)
	}
	return w.GetProgress(ctx, orgID)
}

// adoptFrameworkTx adopts a single framework within a transaction.
func (w *OnboardingWizard) adoptFrameworkTx(ctx context.Context, tx pgx.Tx, orgID, adminID, frameworkID uuid.UUID) (*AdoptedFrameworkResult, error) {
	var code, name string
	var controlCount int
	err := tx.QueryRow(ctx, `
		SELECT code, name, total_controls FROM compliance_frameworks WHERE id = $1`, frameworkID,
	).Scan(&code, &name, &controlCount)
	if err != nil {
		return nil, fmt.Errorf("framework not found: %w", err)
	}

	var orgFrameworkID uuid.UUID
	err = tx.QueryRow(ctx, `
		INSERT INTO organization_frameworks (organization_id, framework_id, status,
			adoption_date, assessment_frequency, responsible_user_id, metadata)
		VALUES ($1, $2, 'in_progress', CURRENT_DATE, 'quarterly', $3, '{}')
		RETURNING id`,
		orgID, frameworkID, adminID,
	).Scan(&orgFrameworkID)
	if err != nil {
		return nil, fmt.Errorf("failed to create organization framework: %w", err)
	}

	result, err := tx.Exec(ctx, `
		INSERT INTO control_implementations (
			organization_id, framework_control_id, org_framework_id,
			status, implementation_status, maturity_level,
			test_frequency, risk_if_not_implemented, automation_level, metadata
		)
		SELECT $1, fc.id, $2,
			'not_implemented', 'not_started', 0,
			'quarterly',
			CASE fc.priority
				WHEN 'critical' THEN 'critical'
				WHEN 'high' THEN 'high'
				ELSE 'medium'
			END,
			'manual', '{}'
		FROM framework_controls fc
		WHERE fc.framework_id = $3`,
		orgID, orgFrameworkID, frameworkID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create control implementations: %w", err)
	}

	insertedCount := int(result.RowsAffected())

	log.Info().
		Str("framework", code).
		Int("controls", insertedCount).
		Msg("Framework adopted during onboarding wizard completion")

	return &AdoptedFrameworkResult{
		OrgFrameworkID: orgFrameworkID,
		FrameworkCode:  code,
		FrameworkName:  name,
		TotalControls:  insertedCount,
	}, nil
}

// createDefaultRiskMatrix creates a default 5x5 risk matrix for the organisation.
func (w *OnboardingWizard) createDefaultRiskMatrix(ctx context.Context, tx pgx.Tx, orgID uuid.UUID) (uuid.UUID, error) {
	likelihoodScale := `[
		{"level":1,"label":"Rare","description":"May occur only in exceptional circumstances","range":"0-5%"},
		{"level":2,"label":"Unlikely","description":"Could occur at some time","range":"5-25%"},
		{"level":3,"label":"Possible","description":"Might occur at some time","range":"25-50%"},
		{"level":4,"label":"Likely","description":"Will probably occur","range":"50-80%"},
		{"level":5,"label":"Almost Certain","description":"Expected to occur","range":"80-100%"}
	]`

	impactScale := `[
		{"level":1,"label":"Insignificant","financial":"<EUR10K","operational":"Minimal disruption"},
		{"level":2,"label":"Minor","financial":"EUR10K-EUR50K","operational":"Some disruption, quickly contained"},
		{"level":3,"label":"Moderate","financial":"EUR50K-EUR250K","operational":"Significant disruption"},
		{"level":4,"label":"Major","financial":"EUR250K-EUR1M","operational":"Major disruption to operations"},
		{"level":5,"label":"Catastrophic","financial":">EUR1M","operational":"Complete loss of critical service"}
	]`

	riskLevels := `[
		{"min_score":1,"max_score":3,"label":"Very Low","color":"#4CAF50"},
		{"min_score":4,"max_score":5,"label":"Low","color":"#8BC34A"},
		{"min_score":6,"max_score":11,"label":"Medium","color":"#FFC107"},
		{"min_score":12,"max_score":19,"label":"High","color":"#FF9800"},
		{"min_score":20,"max_score":25,"label":"Critical","color":"#F44336"}
	]`

	var matrixID uuid.UUID
	err := tx.QueryRow(ctx, `
		INSERT INTO risk_matrices (organization_id, name, description, likelihood_scale,
			impact_scale, risk_levels, matrix_size, is_default)
		VALUES ($1, 'Standard 5x5 Risk Matrix', 'Default risk assessment matrix aligned to ISO 31000',
			$2::JSONB, $3::JSONB, $4::JSONB, 5, true)
		RETURNING id`,
		orgID, likelihoodScale, impactScale, riskLevels,
	).Scan(&matrixID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create risk matrix: %w", err)
	}

	return matrixID, nil
}

// createPendingInvitation creates a pending team invitation record.
func (w *OnboardingWizard) createPendingInvitation(ctx context.Context, tx pgx.Tx, orgID uuid.UUID, inv TeamInvitation) error {
	// Check if user already exists
	var existingCount int
	err := tx.QueryRow(ctx, `
		SELECT COUNT(*) FROM users
		WHERE organization_id = $1 AND email = $2`,
		orgID, inv.Email,
	).Scan(&existingCount)
	if err == nil && existingCount > 0 {
		return nil // User already exists, skip
	}

	// Create pending user
	_, err = tx.Exec(ctx, `
		INSERT INTO users (organization_id, email, first_name, last_name,
			job_title, status, timezone, language, password_hash)
		VALUES ($1, $2, $3, $4, $5, 'pending_verification', 'Europe/London', 'en',
			'$2a$12$pending.invitation.hash.placeholder')
		ON CONFLICT DO NOTHING`,
		orgID, inv.Email, inv.FirstName, inv.LastName, inv.JobTitle,
	)
	if err != nil {
		return fmt.Errorf("failed to create pending user: %w", err)
	}

	// Assign role if specified
	if inv.Role != "" {
		_, err = tx.Exec(ctx, `
			INSERT INTO user_roles (user_id, role_id, organization_id, assigned_by)
			SELECT u.id, r.id, $1, (SELECT id FROM users WHERE organization_id = $1 ORDER BY created_at ASC LIMIT 1)
			FROM users u, roles r
			WHERE u.email = $2 AND u.organization_id = $1
			  AND r.slug = $3
			ON CONFLICT DO NOTHING`,
			orgID, inv.Email, inv.Role,
		)
		if err != nil {
			log.Warn().Err(err).Str("email", inv.Email).Str("role", inv.Role).Msg("Failed to assign role to invited user")
		}
	}

	return nil
}

// formatFrameworkName converts a framework code to a human-readable name.
func formatFrameworkName(code string) string {
	names := map[string]string{
		"ISO27001":          "ISO 27001",
		"UK_GDPR":           "UK GDPR",
		"NCSC_CAF":          "NCSC CAF",
		"NIS2":              "NIS2 Directive",
		"CYBER_ESSENTIALS":  "Cyber Essentials",
		"NIST_800_53":       "NIST 800-53",
		"NIST_CSF_2":        "NIST CSF 2.0",
		"PCI_DSS_4":         "PCI DSS v4.0",
		"ITIL_4":            "ITIL 4",
		"COBIT_2019":        "COBIT 2019",
	}
	if name, ok := names[code]; ok {
		return name
	}
	return strings.ReplaceAll(code, "_", " ")
}
