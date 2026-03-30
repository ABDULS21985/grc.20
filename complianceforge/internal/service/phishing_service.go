package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// ============================================================
// PhishingService
// ============================================================

// PhishingService implements business logic for phishing simulation
// campaigns including creation, launch, interaction tracking,
// completion, and trend analysis.
type PhishingService struct {
	pool *pgxpool.Pool
}

// NewPhishingService creates a new PhishingService with the given database pool.
func NewPhishingService(pool *pgxpool.Pool) *PhishingService {
	return &PhishingService{pool: pool}
}

// ============================================================
// DATA TYPES
// ============================================================

// PhishingSimulation represents a phishing simulation campaign.
type PhishingSimulation struct {
	ID                   uuid.UUID       `json:"id"`
	OrganizationID       uuid.UUID       `json:"organization_id"`
	Name                 string          `json:"name"`
	Description          string          `json:"description"`
	Difficulty           string          `json:"difficulty"`
	EmailTemplateSubject string          `json:"email_template_subject"`
	EmailTemplateBody    string          `json:"email_template_body"`
	LandingPageURL       string          `json:"landing_page_url"`
	TargetUserIDs        []uuid.UUID     `json:"target_user_ids"`
	TargetDepartment     string          `json:"target_department"`
	TargetCount          int             `json:"target_count"`
	Status               string          `json:"status"`
	ScheduledAt          *time.Time      `json:"scheduled_at"`
	LaunchedAt           *time.Time      `json:"launched_at"`
	CompletedAt          *time.Time      `json:"completed_at"`
	TotalSent            int             `json:"total_sent"`
	TotalOpened          int             `json:"total_opened"`
	TotalClicked         int             `json:"total_clicked"`
	TotalSubmitted       int             `json:"total_submitted"`
	TotalReported        int             `json:"total_reported"`
	OpenRate             float64         `json:"open_rate"`
	ClickRate            float64         `json:"click_rate"`
	SubmitRate           float64         `json:"submit_rate"`
	ReportRate           float64         `json:"report_rate"`
	CreatedBy            *uuid.UUID      `json:"created_by"`
	Metadata             json.RawMessage `json:"metadata"`
	CreatedAt            time.Time       `json:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at"`
}

// PhishingResult represents a single user's result in a phishing simulation.
type PhishingResult struct {
	ID              uuid.UUID       `json:"id"`
	OrganizationID  uuid.UUID       `json:"organization_id"`
	SimulationID    uuid.UUID       `json:"simulation_id"`
	UserID          uuid.UUID       `json:"user_id"`
	EmailSentAt     *time.Time      `json:"email_sent_at"`
	EmailOpenedAt   *time.Time      `json:"email_opened_at"`
	LinkClickedAt   *time.Time      `json:"link_clicked_at"`
	DataSubmittedAt *time.Time      `json:"data_submitted_at"`
	ReportedAt      *time.Time      `json:"reported_at"`
	FinalAction     string          `json:"final_action"`
	UserAgent       string          `json:"user_agent"`
	IPAddress       string          `json:"ip_address"`
	Metadata        json.RawMessage `json:"metadata"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	UserEmail       string          `json:"user_email,omitempty"`
	UserFullName    string          `json:"user_full_name,omitempty"`
}

// PhishingTrendPoint represents a single data point in the phishing click-rate trend.
type PhishingTrendPoint struct {
	SimulationID   uuid.UUID  `json:"simulation_id"`
	SimulationName string     `json:"simulation_name"`
	CompletedAt    time.Time  `json:"completed_at"`
	Difficulty     string     `json:"difficulty"`
	TotalSent      int        `json:"total_sent"`
	ClickRate      float64    `json:"click_rate"`
	ReportRate     float64    `json:"report_rate"`
	SubmitRate     float64    `json:"submit_rate"`
}

// ============================================================
// REQUEST TYPES
// ============================================================

// CreateSimulationReq is the request body for creating a phishing simulation.
type CreateSimulationReq struct {
	Name                 string      `json:"name"`
	Description          string      `json:"description"`
	Difficulty           string      `json:"difficulty"`
	EmailTemplateSubject string      `json:"email_template_subject"`
	EmailTemplateBody    string      `json:"email_template_body"`
	LandingPageURL       string      `json:"landing_page_url"`
	TargetUserIDs        []uuid.UUID `json:"target_user_ids"`
	TargetDepartment     string      `json:"target_department"`
	ScheduledAt          *time.Time  `json:"scheduled_at"`
}

// ============================================================
// SIMULATION CRUD
// ============================================================

// ListSimulations returns a paginated list of phishing simulations.
func (s *PhishingService) ListSimulations(ctx context.Context, orgID uuid.UUID, page, pageSize int) ([]PhishingSimulation, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var total int64
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM phishing_simulations
		WHERE organization_id = $1`, orgID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count phishing simulations: %w", err)
	}

	rows, err := s.pool.Query(ctx, `
		SELECT
			id, organization_id, name, COALESCE(description, ''),
			difficulty::TEXT, COALESCE(email_template_subject, ''),
			COALESCE(email_template_body, ''), COALESCE(landing_page_url, ''),
			COALESCE(target_user_ids, '{}'), COALESCE(target_department, ''),
			COALESCE(target_count, 0),
			status::TEXT, scheduled_at, launched_at, completed_at,
			COALESCE(total_sent, 0), COALESCE(total_opened, 0),
			COALESCE(total_clicked, 0), COALESCE(total_submitted, 0),
			COALESCE(total_reported, 0),
			COALESCE(open_rate, 0), COALESCE(click_rate, 0),
			COALESCE(submit_rate, 0), COALESCE(report_rate, 0),
			created_by, COALESCE(metadata, '{}'::jsonb),
			created_at, updated_at
		FROM phishing_simulations
		WHERE organization_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`,
		orgID, pageSize, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list phishing simulations: %w", err)
	}
	defer rows.Close()

	var sims []PhishingSimulation
	for rows.Next() {
		var sim PhishingSimulation
		if err := rows.Scan(
			&sim.ID, &sim.OrganizationID, &sim.Name, &sim.Description,
			&sim.Difficulty, &sim.EmailTemplateSubject,
			&sim.EmailTemplateBody, &sim.LandingPageURL,
			&sim.TargetUserIDs, &sim.TargetDepartment,
			&sim.TargetCount,
			&sim.Status, &sim.ScheduledAt, &sim.LaunchedAt, &sim.CompletedAt,
			&sim.TotalSent, &sim.TotalOpened,
			&sim.TotalClicked, &sim.TotalSubmitted,
			&sim.TotalReported,
			&sim.OpenRate, &sim.ClickRate,
			&sim.SubmitRate, &sim.ReportRate,
			&sim.CreatedBy, &sim.Metadata,
			&sim.CreatedAt, &sim.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan phishing simulation: %w", err)
		}
		sims = append(sims, sim)
	}
	return sims, total, nil
}

// CreateSimulation creates a new phishing simulation campaign.
func (s *PhishingService) CreateSimulation(ctx context.Context, orgID uuid.UUID, req CreateSimulationReq) (*PhishingSimulation, error) {
	if req.Difficulty == "" {
		req.Difficulty = "medium"
	}

	targetCount := len(req.TargetUserIDs)
	if targetCount == 0 && req.TargetDepartment != "" {
		// Count users in the target department
		err := s.pool.QueryRow(ctx, `
			SELECT COUNT(*) FROM users
			WHERE organization_id = $1 AND status = 'active' AND department = $2`,
			orgID, req.TargetDepartment,
		).Scan(&targetCount)
		if err != nil {
			targetCount = 0
		}
	}

	sim := &PhishingSimulation{}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO phishing_simulations (
			organization_id, name, description, difficulty,
			email_template_subject, email_template_body, landing_page_url,
			target_user_ids, target_department, target_count,
			scheduled_at
		) VALUES (
			$1, $2, $3, $4::phishing_difficulty,
			$5, $6, $7,
			$8, $9, $10,
			$11
		)
		RETURNING id, organization_id, name, COALESCE(description, ''),
			difficulty::TEXT, COALESCE(email_template_subject, ''),
			COALESCE(email_template_body, ''), COALESCE(landing_page_url, ''),
			COALESCE(target_user_ids, '{}'), COALESCE(target_department, ''),
			COALESCE(target_count, 0),
			status::TEXT, scheduled_at, launched_at, completed_at,
			COALESCE(total_sent, 0), COALESCE(total_opened, 0),
			COALESCE(total_clicked, 0), COALESCE(total_submitted, 0),
			COALESCE(total_reported, 0),
			COALESCE(open_rate, 0), COALESCE(click_rate, 0),
			COALESCE(submit_rate, 0), COALESCE(report_rate, 0),
			created_by, COALESCE(metadata, '{}'::jsonb),
			created_at, updated_at`,
		orgID, req.Name, req.Description, req.Difficulty,
		req.EmailTemplateSubject, req.EmailTemplateBody, req.LandingPageURL,
		req.TargetUserIDs, req.TargetDepartment, targetCount,
		req.ScheduledAt,
	).Scan(
		&sim.ID, &sim.OrganizationID, &sim.Name, &sim.Description,
		&sim.Difficulty, &sim.EmailTemplateSubject,
		&sim.EmailTemplateBody, &sim.LandingPageURL,
		&sim.TargetUserIDs, &sim.TargetDepartment,
		&sim.TargetCount,
		&sim.Status, &sim.ScheduledAt, &sim.LaunchedAt, &sim.CompletedAt,
		&sim.TotalSent, &sim.TotalOpened,
		&sim.TotalClicked, &sim.TotalSubmitted,
		&sim.TotalReported,
		&sim.OpenRate, &sim.ClickRate,
		&sim.SubmitRate, &sim.ReportRate,
		&sim.CreatedBy, &sim.Metadata,
		&sim.CreatedAt, &sim.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create phishing simulation: %w", err)
	}

	log.Info().
		Str("org_id", orgID.String()).
		Str("simulation_id", sim.ID.String()).
		Str("name", sim.Name).
		Msg("Phishing simulation created")

	return sim, nil
}

// LaunchSimulation transitions a simulation to active status and creates
// result entries for each target user.
func (s *PhishingService) LaunchSimulation(ctx context.Context, orgID, simulationID uuid.UUID) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Set RLS context
	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return fmt.Errorf("failed to set org context: %w", err)
	}

	// Fetch simulation
	var status string
	var targetUserIDs []uuid.UUID
	var targetDept string
	err = tx.QueryRow(ctx, `
		SELECT status::TEXT, COALESCE(target_user_ids, '{}'), COALESCE(target_department, '')
		FROM phishing_simulations
		WHERE id = $1 AND organization_id = $2`,
		simulationID, orgID,
	).Scan(&status, &targetUserIDs, &targetDept)
	if err != nil {
		return fmt.Errorf("simulation not found: %w", err)
	}

	if status != "draft" && status != "scheduled" {
		return fmt.Errorf("simulation cannot be launched from status %s", status)
	}

	// Determine target users
	var userIDs []uuid.UUID
	if len(targetUserIDs) > 0 {
		userIDs = targetUserIDs
	} else if targetDept != "" {
		rows, err := tx.Query(ctx, `
			SELECT id FROM users
			WHERE organization_id = $1 AND status = 'active' AND department = $2`,
			orgID, targetDept,
		)
		if err != nil {
			return fmt.Errorf("failed to query target department users: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			var uid uuid.UUID
			if err := rows.Scan(&uid); err != nil {
				return fmt.Errorf("failed to scan user ID: %w", err)
			}
			userIDs = append(userIDs, uid)
		}
	} else {
		// Default: all active users
		rows, err := tx.Query(ctx, `
			SELECT id FROM users WHERE organization_id = $1 AND status = 'active'`,
			orgID,
		)
		if err != nil {
			return fmt.Errorf("failed to query all users: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			var uid uuid.UUID
			if err := rows.Scan(&uid); err != nil {
				return fmt.Errorf("failed to scan user ID: %w", err)
			}
			userIDs = append(userIDs, uid)
		}
	}

	// Create result entries for each user
	now := time.Now()
	for _, uid := range userIDs {
		_, err = tx.Exec(ctx, `
			INSERT INTO phishing_simulation_results (
				organization_id, simulation_id, user_id,
				email_sent_at, final_action
			) VALUES ($1, $2, $3, $4, 'delivered')
			ON CONFLICT (simulation_id, user_id) DO NOTHING`,
			orgID, simulationID, uid, now,
		)
		if err != nil {
			return fmt.Errorf("failed to create result for user %s: %w", uid, err)
		}
	}

	// Update simulation status
	_, err = tx.Exec(ctx, `
		UPDATE phishing_simulations SET
			status = 'active',
			launched_at = $3,
			total_sent = $4,
			target_count = $4
		WHERE id = $1 AND organization_id = $2`,
		simulationID, orgID, now, len(userIDs),
	)
	if err != nil {
		return fmt.Errorf("failed to update simulation status: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit launch: %w", err)
	}

	log.Info().
		Str("simulation_id", simulationID.String()).
		Int("target_users", len(userIDs)).
		Msg("Phishing simulation launched")

	return nil
}

// GetSimulationResults returns all individual results for a simulation.
func (s *PhishingService) GetSimulationResults(ctx context.Context, orgID, simulationID uuid.UUID) ([]PhishingResult, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT
			psr.id, psr.organization_id, psr.simulation_id, psr.user_id,
			psr.email_sent_at, psr.email_opened_at, psr.link_clicked_at,
			psr.data_submitted_at, psr.reported_at,
			psr.final_action::TEXT,
			COALESCE(psr.user_agent, ''), COALESCE(psr.ip_address, ''),
			COALESCE(psr.metadata, '{}'::jsonb),
			psr.created_at, psr.updated_at,
			u.email, COALESCE(u.full_name, u.email)
		FROM phishing_simulation_results psr
		JOIN users u ON u.id = psr.user_id
		WHERE psr.simulation_id = $1 AND psr.organization_id = $2
		ORDER BY psr.final_action DESC, u.full_name`,
		simulationID, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get simulation results: %w", err)
	}
	defer rows.Close()

	var results []PhishingResult
	for rows.Next() {
		var r PhishingResult
		if err := rows.Scan(
			&r.ID, &r.OrganizationID, &r.SimulationID, &r.UserID,
			&r.EmailSentAt, &r.EmailOpenedAt, &r.LinkClickedAt,
			&r.DataSubmittedAt, &r.ReportedAt,
			&r.FinalAction,
			&r.UserAgent, &r.IPAddress,
			&r.Metadata,
			&r.CreatedAt, &r.UpdatedAt,
			&r.UserEmail, &r.UserFullName,
		); err != nil {
			return nil, fmt.Errorf("failed to scan phishing result: %w", err)
		}
		results = append(results, r)
	}
	return results, nil
}

// RecordInteraction records a user's interaction with a phishing simulation.
// Actions: opened, clicked, submitted_data, reported.
func (s *PhishingService) RecordInteraction(ctx context.Context, simulationID, userID uuid.UUID, action string) error {
	now := time.Now()

	// Determine which column to update and the action ordering
	var updateClause string
	switch action {
	case "opened":
		updateClause = "email_opened_at = COALESCE(email_opened_at, $3)"
	case "clicked":
		updateClause = "email_opened_at = COALESCE(email_opened_at, $3), link_clicked_at = COALESCE(link_clicked_at, $3)"
	case "submitted_data":
		updateClause = "email_opened_at = COALESCE(email_opened_at, $3), link_clicked_at = COALESCE(link_clicked_at, $3), data_submitted_at = COALESCE(data_submitted_at, $3)"
	case "reported":
		updateClause = "reported_at = COALESCE(reported_at, $3)"
	default:
		return fmt.Errorf("invalid phishing action: %s", action)
	}

	// Update the result with the worst action
	actionOrder := map[string]int{
		"delivered":      0,
		"opened":         1,
		"clicked":        2,
		"submitted_data": 3,
		"reported":       0, // reported is a positive action, does not escalate
	}

	query := fmt.Sprintf(`
		UPDATE phishing_simulation_results SET
			%s,
			final_action = CASE
				WHEN $4 > 0 AND $4 > (
					CASE final_action
						WHEN 'delivered' THEN 0
						WHEN 'opened' THEN 1
						WHEN 'clicked' THEN 2
						WHEN 'submitted_data' THEN 3
						WHEN 'reported' THEN 0
					END
				) THEN $5::phishing_action
				ELSE final_action
			END
		WHERE simulation_id = $1 AND user_id = $2`, updateClause)

	ct, err := s.pool.Exec(ctx, query,
		simulationID, userID, now,
		actionOrder[action], action,
	)
	if err != nil {
		return fmt.Errorf("failed to record phishing interaction: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("phishing result not found for user")
	}

	log.Debug().
		Str("simulation_id", simulationID.String()).
		Str("user_id", userID.String()).
		Str("action", action).
		Msg("Phishing interaction recorded")

	return nil
}

// CompleteSimulation finalises a simulation, computing aggregate rates.
func (s *PhishingService) CompleteSimulation(ctx context.Context, orgID, simulationID uuid.UUID) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Aggregate results
	var totalSent, totalOpened, totalClicked, totalSubmitted, totalReported int
	err = tx.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE email_opened_at IS NOT NULL),
			COUNT(*) FILTER (WHERE link_clicked_at IS NOT NULL),
			COUNT(*) FILTER (WHERE data_submitted_at IS NOT NULL),
			COUNT(*) FILTER (WHERE reported_at IS NOT NULL)
		FROM phishing_simulation_results
		WHERE simulation_id = $1 AND organization_id = $2`,
		simulationID, orgID,
	).Scan(&totalSent, &totalOpened, &totalClicked, &totalSubmitted, &totalReported)
	if err != nil {
		return fmt.Errorf("failed to aggregate results: %w", err)
	}

	// Calculate rates
	var openRate, clickRate, submitRate, reportRate float64
	if totalSent > 0 {
		openRate = float64(totalOpened) / float64(totalSent) * 100
		clickRate = float64(totalClicked) / float64(totalSent) * 100
		submitRate = float64(totalSubmitted) / float64(totalSent) * 100
		reportRate = float64(totalReported) / float64(totalSent) * 100
	}

	now := time.Now()
	_, err = tx.Exec(ctx, `
		UPDATE phishing_simulations SET
			status = 'completed',
			completed_at = $3,
			total_sent = $4,
			total_opened = $5,
			total_clicked = $6,
			total_submitted = $7,
			total_reported = $8,
			open_rate = $9,
			click_rate = $10,
			submit_rate = $11,
			report_rate = $12
		WHERE id = $1 AND organization_id = $2`,
		simulationID, orgID, now,
		totalSent, totalOpened, totalClicked, totalSubmitted, totalReported,
		openRate, clickRate, submitRate, reportRate,
	)
	if err != nil {
		return fmt.Errorf("failed to update simulation: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit completion: %w", err)
	}

	log.Info().
		Str("simulation_id", simulationID.String()).
		Float64("click_rate", clickRate).
		Float64("report_rate", reportRate).
		Msg("Phishing simulation completed")

	return nil
}

// GetPhishingTrend returns the click rate over time across completed simulations.
func (s *PhishingService) GetPhishingTrend(ctx context.Context, orgID uuid.UUID) ([]PhishingTrendPoint, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT
			id, name, completed_at, difficulty::TEXT,
			COALESCE(total_sent, 0),
			COALESCE(click_rate, 0),
			COALESCE(report_rate, 0),
			COALESCE(submit_rate, 0)
		FROM phishing_simulations
		WHERE organization_id = $1 AND status = 'completed' AND completed_at IS NOT NULL
		ORDER BY completed_at ASC`,
		orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get phishing trend: %w", err)
	}
	defer rows.Close()

	var points []PhishingTrendPoint
	for rows.Next() {
		var p PhishingTrendPoint
		if err := rows.Scan(
			&p.SimulationID, &p.SimulationName, &p.CompletedAt, &p.Difficulty,
			&p.TotalSent, &p.ClickRate, &p.ReportRate, &p.SubmitRate,
		); err != nil {
			return nil, fmt.Errorf("failed to scan trend point: %w", err)
		}
		points = append(points, p)
	}
	return points, nil
}
