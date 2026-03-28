package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// OnboardingService handles the complete workflow for setting up a new organisation.
// This includes creating the organisation, admin user, default risk matrix,
// adopting selected frameworks, and initialising control implementations.
type OnboardingService struct {
	pool *pgxpool.Pool
}

func NewOnboardingService(pool *pgxpool.Pool) *OnboardingService {
	return &OnboardingService{pool: pool}
}

// OnboardingRequest contains all data needed to onboard a new organisation.
type OnboardingRequest struct {
	// Organisation
	OrgName            string   `json:"org_name"`
	LegalName          string   `json:"legal_name"`
	Industry           string   `json:"industry"`
	CountryCode        string   `json:"country_code"`
	Timezone           string   `json:"timezone"`
	EmployeeCountRange string   `json:"employee_count_range"`
	// Admin user
	AdminEmail     string `json:"admin_email"`
	AdminPassword  string `json:"admin_password"`
	AdminFirstName string `json:"admin_first_name"`
	AdminLastName  string `json:"admin_last_name"`
	AdminJobTitle  string `json:"admin_job_title"`
	// Frameworks to adopt
	FrameworkIDs []uuid.UUID `json:"framework_ids"`
	// Subscription
	PlanName string `json:"plan_name"` // starter, professional, enterprise
}

// OnboardingResult contains the IDs and details of everything created.
type OnboardingResult struct {
	OrganizationID uuid.UUID   `json:"organization_id"`
	AdminUserID    uuid.UUID   `json:"admin_user_id"`
	RiskMatrixID   uuid.UUID   `json:"risk_matrix_id"`
	AdoptedFrameworks []AdoptedFrameworkResult `json:"adopted_frameworks"`
	ControlsInitialised int   `json:"controls_initialised"`
	CreatedAt      time.Time   `json:"created_at"`
}

type AdoptedFrameworkResult struct {
	OrgFrameworkID uuid.UUID `json:"org_framework_id"`
	FrameworkCode  string    `json:"framework_code"`
	FrameworkName  string    `json:"framework_name"`
	TotalControls  int       `json:"total_controls"`
}

// Onboard performs the complete onboarding workflow in a single transaction.
func (s *OnboardingService) Onboard(ctx context.Context, req OnboardingRequest) (*OnboardingResult, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	result := &OnboardingResult{CreatedAt: time.Now()}

	// ── Step 1: Create Organisation ─────────────────────
	log.Info().Str("org", req.OrgName).Msg("Onboarding: Creating organisation")

	slug := generateSlug(req.OrgName)
	tier := "starter"
	if req.PlanName != "" {
		tier = req.PlanName
	}

	err = tx.QueryRow(ctx, `
		INSERT INTO organizations (name, slug, legal_name, industry, country_code, 
			status, tier, timezone, default_language, employee_count_range, metadata)
		VALUES ($1, $2, $3, $4, $5, 'active', $6, $7, 'en', $8, '{}')
		RETURNING id`,
		req.OrgName, slug, req.LegalName, req.Industry, req.CountryCode,
		tier, req.Timezone, req.EmployeeCountRange,
	).Scan(&result.OrganizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to create organisation: %w", err)
	}

	// ── Step 2: Create Subscription ─────────────────────
	maxUsers := 5
	maxFrameworks := 3
	switch tier {
	case "professional":
		maxUsers = 25
		maxFrameworks = 5
	case "enterprise":
		maxUsers = 100
		maxFrameworks = 9
	case "unlimited":
		maxUsers = 9999
		maxFrameworks = 9
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO organization_subscriptions (organization_id, plan_name, status, 
			max_users, max_frameworks, billing_cycle, current_period_start, current_period_end)
		VALUES ($1, $2, 'active', $3, $4, 'annual', NOW(), NOW() + INTERVAL '1 year')`,
		result.OrganizationID, tier, maxUsers, maxFrameworks,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	// ── Step 3: Create Admin User ───────────────────────
	log.Info().Str("email", req.AdminEmail).Msg("Onboarding: Creating admin user")

	// In production, use bcrypt. Here we store a placeholder.
	err = tx.QueryRow(ctx, `
		INSERT INTO users (organization_id, email, password_hash, first_name, last_name, 
			job_title, status, timezone, language)
		VALUES ($1, $2, '$2a$12$placeholder.hash.will.be.replaced', $3, $4, $5, 'active', $6, 'en')
		RETURNING id`,
		result.OrganizationID, req.AdminEmail, req.AdminFirstName, 
		req.AdminLastName, req.AdminJobTitle, req.Timezone,
	).Scan(&result.AdminUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to create admin user: %w", err)
	}

	// Assign org_admin role
	_, err = tx.Exec(ctx, `
		INSERT INTO user_roles (user_id, role_id, organization_id, assigned_by)
		SELECT $1, id, $2, $1 FROM roles WHERE slug = 'org_admin' AND is_system_role = true`,
		result.AdminUserID, result.OrganizationID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to assign admin role: %w", err)
	}

	// ── Step 4: Create Default Risk Matrix ──────────────
	log.Info().Msg("Onboarding: Creating default 5x5 risk matrix")

	likelihoodScale := `[
		{"level":1,"label":"Rare","description":"May occur only in exceptional circumstances","range":"0-5%"},
		{"level":2,"label":"Unlikely","description":"Could occur at some time","range":"5-25%"},
		{"level":3,"label":"Possible","description":"Might occur at some time","range":"25-50%"},
		{"level":4,"label":"Likely","description":"Will probably occur","range":"50-80%"},
		{"level":5,"label":"Almost Certain","description":"Expected to occur","range":"80-100%"}
	]`

	impactScale := `[
		{"level":1,"label":"Insignificant","financial":"<€10K","operational":"Minimal disruption"},
		{"level":2,"label":"Minor","financial":"€10K-€50K","operational":"Some disruption, quickly contained"},
		{"level":3,"label":"Moderate","financial":"€50K-€250K","operational":"Significant disruption"},
		{"level":4,"label":"Major","financial":"€250K-€1M","operational":"Major disruption to operations"},
		{"level":5,"label":"Catastrophic","financial":">€1M","operational":"Complete loss of critical service"}
	]`

	riskLevels := `[
		{"min_score":1,"max_score":3,"label":"Very Low","color":"#4CAF50"},
		{"min_score":4,"max_score":5,"label":"Low","color":"#8BC34A"},
		{"min_score":6,"max_score":11,"label":"Medium","color":"#FFC107"},
		{"min_score":12,"max_score":19,"label":"High","color":"#FF9800"},
		{"min_score":20,"max_score":25,"label":"Critical","color":"#F44336"}
	]`

	err = tx.QueryRow(ctx, `
		INSERT INTO risk_matrices (organization_id, name, description, likelihood_scale, 
			impact_scale, risk_levels, matrix_size, is_default)
		VALUES ($1, 'Standard 5×5 Risk Matrix', 'Default risk assessment matrix aligned to ISO 31000', 
			$2::JSONB, $3::JSONB, $4::JSONB, 5, true)
		RETURNING id`,
		result.OrganizationID, likelihoodScale, impactScale, riskLevels,
	).Scan(&result.RiskMatrixID)
	if err != nil {
		return nil, fmt.Errorf("failed to create risk matrix: %w", err)
	}

	// ── Step 5: Adopt Frameworks & Initialise Controls ──
	totalControls := 0
	for _, fwID := range req.FrameworkIDs {
		adopted, err := s.adoptFramework(ctx, tx, result.OrganizationID, result.AdminUserID, fwID)
		if err != nil {
			log.Warn().Err(err).Str("framework_id", fwID.String()).Msg("Failed to adopt framework")
			continue
		}
		result.AdoptedFrameworks = append(result.AdoptedFrameworks, *adopted)
		totalControls += adopted.TotalControls
	}
	result.ControlsInitialised = totalControls

	// ── Commit ──────────────────────────────────────────
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit onboarding transaction: %w", err)
	}

	log.Info().
		Str("org_id", result.OrganizationID.String()).
		Str("admin_id", result.AdminUserID.String()).
		Int("frameworks", len(result.AdoptedFrameworks)).
		Int("controls", result.ControlsInitialised).
		Msg("Onboarding complete")

	return result, nil
}

// adoptFramework adopts a single framework and creates control_implementations for all its controls.
func (s *OnboardingService) adoptFramework(ctx context.Context, tx pgx.Tx, orgID, adminID, frameworkID uuid.UUID) (*AdoptedFrameworkResult, error) {
	// Get framework info
	var code, name string
	var controlCount int
	err := tx.QueryRow(ctx, `
		SELECT code, name, total_controls FROM compliance_frameworks WHERE id = $1`, frameworkID,
	).Scan(&code, &name, &controlCount)
	if err != nil {
		return nil, err
	}

	// Create organization_frameworks record
	var orgFrameworkID uuid.UUID
	err = tx.QueryRow(ctx, `
		INSERT INTO organization_frameworks (organization_id, framework_id, status, 
			adoption_date, assessment_frequency, responsible_user_id, metadata)
		VALUES ($1, $2, 'in_progress', CURRENT_DATE, 'quarterly', $3, '{}')
		RETURNING id`,
		orgID, frameworkID, adminID,
	).Scan(&orgFrameworkID)
	if err != nil {
		return nil, err
	}

	// Create control_implementations for every control in this framework
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
		return nil, err
	}

	insertedCount := int(result.RowsAffected())

	log.Info().
		Str("framework", code).
		Int("controls", insertedCount).
		Msg("Framework adopted with control implementations")

	return &AdoptedFrameworkResult{
		OrgFrameworkID: orgFrameworkID,
		FrameworkCode:  code,
		FrameworkName:  name,
		TotalControls:  insertedCount,
	}, nil
}

func generateSlug(name string) string {
	slug := make([]byte, 0, len(name))
	for _, c := range []byte(name) {
		switch {
		case c >= 'a' && c <= 'z':
			slug = append(slug, c)
		case c >= 'A' && c <= 'Z':
			slug = append(slug, c+32) // lowercase
		case c >= '0' && c <= '9':
			slug = append(slug, c)
		case c == ' ' || c == '-':
			if len(slug) > 0 && slug[len(slug)-1] != '-' {
				slug = append(slug, '-')
			}
		}
	}
	if len(slug) > 100 {
		slug = slug[:100]
	}
	return string(slug)
}
