package service

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// ============================================================
// AuditProgrammeService
// ============================================================

// AuditProgrammeService implements business logic for advanced audit
// management: programmes, universe, engagements, workpapers, ISA 530
// sampling, test procedures, and corrective actions.
type AuditProgrammeService struct {
	pool *pgxpool.Pool
}

// NewAuditProgrammeService creates a new AuditProgrammeService with the given database pool.
func NewAuditProgrammeService(pool *pgxpool.Pool) *AuditProgrammeService {
	return &AuditProgrammeService{pool: pool}
}

// ============================================================
// DATA TYPES
// ============================================================

// AuditProgramme represents an audit programme (typically annual).
type AuditProgramme struct {
	ID              uuid.UUID  `json:"id"`
	OrganizationID  uuid.UUID  `json:"organization_id"`
	ProgrammeRef    string     `json:"programme_ref"`
	Name            string     `json:"name"`
	Description     string     `json:"description"`
	Status          string     `json:"status"`
	ProgrammeType   string     `json:"programme_type"`
	PeriodStart     time.Time  `json:"period_start"`
	PeriodEnd       time.Time  `json:"period_end"`
	TotalBudgetDays int        `json:"total_budget_days"`
	UsedBudgetDays  int        `json:"used_budget_days"`
	Objectives      string     `json:"objectives"`
	RiskAppetite    string     `json:"risk_appetite"`
	Methodology     string     `json:"methodology"`
	ApprovedBy      *uuid.UUID `json:"approved_by,omitempty"`
	ApprovedAt      *time.Time `json:"approved_at,omitempty"`
	OwnerUserID     *uuid.UUID `json:"owner_user_id,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// AuditableEntity represents an entry in the audit universe.
type AuditableEntity struct {
	ID                   uuid.UUID   `json:"id"`
	OrganizationID       uuid.UUID   `json:"organization_id"`
	EntityRef            string      `json:"entity_ref"`
	Name                 string      `json:"name"`
	EntityType           string      `json:"entity_type"`
	Description          string      `json:"description"`
	RiskRating           string      `json:"risk_rating"`
	RiskScore            float64     `json:"risk_score"`
	BusinessOwnerID      *uuid.UUID  `json:"business_owner_id,omitempty"`
	Department           string      `json:"department"`
	Location             string      `json:"location"`
	RegulatoryRelevance  []string    `json:"regulatory_relevance"`
	LastAuditDate        *time.Time  `json:"last_audit_date,omitempty"`
	NextAuditDue         *time.Time  `json:"next_audit_due,omitempty"`
	AuditFrequencyMonths int         `json:"audit_frequency_months"`
	Status               string      `json:"status"`
	LinkedFrameworkIDs   []uuid.UUID `json:"linked_framework_ids"`
	LinkedRiskIDs        []uuid.UUID `json:"linked_risk_ids"`
	CreatedAt            time.Time   `json:"created_at"`
	UpdatedAt            time.Time   `json:"updated_at"`
}

// AuditEngagement represents an engagement within a programme.
type AuditEngagement struct {
	ID                uuid.UUID   `json:"id"`
	OrganizationID    uuid.UUID   `json:"organization_id"`
	ProgrammeID       uuid.UUID   `json:"programme_id"`
	AuditID           *uuid.UUID  `json:"audit_id,omitempty"`
	AuditableEntityID *uuid.UUID  `json:"auditable_entity_id,omitempty"`
	EngagementRef     string      `json:"engagement_ref"`
	Name              string      `json:"name"`
	EngagementType    string      `json:"engagement_type"`
	Status            string      `json:"status"`
	Priority          string      `json:"priority"`
	RiskRating        string      `json:"risk_rating"`
	Scope             string      `json:"scope"`
	Objectives        string      `json:"objectives"`
	Methodology       string      `json:"methodology"`
	LeadAuditorID     *uuid.UUID  `json:"lead_auditor_id,omitempty"`
	AuditTeamIDs      []uuid.UUID `json:"audit_team_ids"`
	PlannedStartDate  *time.Time  `json:"planned_start_date,omitempty"`
	PlannedEndDate    *time.Time  `json:"planned_end_date,omitempty"`
	ActualStartDate   *time.Time  `json:"actual_start_date,omitempty"`
	ActualEndDate     *time.Time  `json:"actual_end_date,omitempty"`
	BudgetDays        int         `json:"budget_days"`
	ActualDays        int         `json:"actual_days"`
	FieldworkComplete bool        `json:"fieldwork_complete"`
	ReportIssued      bool        `json:"report_issued"`
	ReportIssuedDate  *time.Time  `json:"report_issued_date,omitempty"`
	OverallOpinion    string      `json:"overall_opinion"`
	CreatedAt         time.Time   `json:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at"`
}

// AuditWorkpaper represents an audit workpaper.
type AuditWorkpaper struct {
	ID              uuid.UUID   `json:"id"`
	OrganizationID  uuid.UUID   `json:"organization_id"`
	EngagementID    uuid.UUID   `json:"engagement_id"`
	WorkpaperRef    string      `json:"workpaper_ref"`
	Title           string      `json:"title"`
	Description     string      `json:"description"`
	WorkpaperType   string      `json:"workpaper_type"`
	Status          string      `json:"status"`
	Content         string      `json:"content"`
	PreparedBy      uuid.UUID   `json:"prepared_by"`
	PreparedDate    time.Time   `json:"prepared_date"`
	ReviewedBy      *uuid.UUID  `json:"reviewed_by,omitempty"`
	ReviewedDate    *time.Time  `json:"reviewed_date,omitempty"`
	ReviewComments  string      `json:"review_comments"`
	LinkedControlIDs []uuid.UUID `json:"linked_control_ids"`
	LinkedRiskIDs   []uuid.UUID `json:"linked_risk_ids"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

// AuditSample represents a statistical sample for audit testing.
type AuditSample struct {
	ID                  uuid.UUID    `json:"id"`
	OrganizationID      uuid.UUID    `json:"organization_id"`
	EngagementID        uuid.UUID    `json:"engagement_id"`
	SampleRef           string       `json:"sample_ref"`
	Name                string       `json:"name"`
	Description         string       `json:"description"`
	SamplingMethod      string       `json:"sampling_method"`
	PopulationSize      int          `json:"population_size"`
	SampleSize          int          `json:"sample_size"`
	ConfidenceLevel     float64      `json:"confidence_level"`
	TolerableErrorRate  float64      `json:"tolerable_error_rate"`
	ExpectedErrorRate   float64      `json:"expected_error_rate"`
	ZScore              float64      `json:"z_score"`
	SelectionMethod     string       `json:"selection_method"`
	SelectedItems       []SampleItem `json:"selected_items"`
	ItemsTested         int          `json:"items_tested"`
	ItemsPassed         int          `json:"items_passed"`
	ItemsFailed         int          `json:"items_failed"`
	ItemsInconclusive   int          `json:"items_inconclusive"`
	ActualErrorRate     *float64     `json:"actual_error_rate,omitempty"`
	Conclusion          string       `json:"conclusion"`
	Status              string       `json:"status"`
	CreatedBy           uuid.UUID    `json:"created_by"`
	CreatedAt           time.Time    `json:"created_at"`
	UpdatedAt           time.Time    `json:"updated_at"`
}

// SampleItem represents an individual item within an audit sample.
type SampleItem struct {
	Index       int    `json:"index"`
	ItemID      string `json:"item_id"`
	Description string `json:"description"`
	Result      string `json:"result"`
	Notes       string `json:"notes"`
	TestedAt    string `json:"tested_at,omitempty"`
}

// SampleConfig holds parameters for generating a statistical sample.
type SampleConfig struct {
	Name               string  `json:"name"`
	Description        string  `json:"description"`
	PopulationSize     int     `json:"population_size"`
	ConfidenceLevel    float64 `json:"confidence_level"`
	TolerableErrorRate float64 `json:"tolerable_error_rate"`
	ExpectedErrorRate  float64 `json:"expected_error_rate"`
}

// AuditTestProcedure represents a test procedure within an engagement.
type AuditTestProcedure struct {
	ID             uuid.UUID  `json:"id"`
	OrganizationID uuid.UUID  `json:"organization_id"`
	EngagementID   uuid.UUID  `json:"engagement_id"`
	ProcedureRef   string     `json:"procedure_ref"`
	Title          string     `json:"title"`
	Description    string     `json:"description"`
	TestType       string     `json:"test_type"`
	ControlID      *uuid.UUID `json:"control_id,omitempty"`
	ControlRef     string     `json:"control_ref"`
	ExpectedResult string     `json:"expected_result"`
	ActualResult   string     `json:"actual_result"`
	Result         string     `json:"result"`
	TestedBy       *uuid.UUID `json:"tested_by,omitempty"`
	TestedDate     *time.Time `json:"tested_date,omitempty"`
	WorkpaperID    *uuid.UUID `json:"workpaper_id,omitempty"`
	SampleID       *uuid.UUID `json:"sample_id,omitempty"`
	FindingID      *uuid.UUID `json:"finding_id,omitempty"`
	EvidenceRefs   []string   `json:"evidence_refs"`
	Notes          string     `json:"notes"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// AuditCorrectiveAction represents a corrective action arising from an audit finding.
type AuditCorrectiveAction struct {
	ID                  uuid.UUID  `json:"id"`
	OrganizationID      uuid.UUID  `json:"organization_id"`
	FindingID           *uuid.UUID `json:"finding_id,omitempty"`
	EngagementID        *uuid.UUID `json:"engagement_id,omitempty"`
	ActionRef           string     `json:"action_ref"`
	Title               string     `json:"title"`
	Description         string     `json:"description"`
	ActionType          string     `json:"action_type"`
	Priority            string     `json:"priority"`
	Status              string     `json:"status"`
	RootCause           string     `json:"root_cause"`
	PlannedAction       string     `json:"planned_action"`
	ActualAction        string     `json:"actual_action"`
	ResponsibleUserID   *uuid.UUID `json:"responsible_user_id,omitempty"`
	ImplementerUserID   *uuid.UUID `json:"implementer_user_id,omitempty"`
	DueDate             *time.Time `json:"due_date,omitempty"`
	CompletedDate       *time.Time `json:"completed_date,omitempty"`
	VerifiedBy          *uuid.UUID `json:"verified_by,omitempty"`
	VerifiedDate        *time.Time `json:"verified_date,omitempty"`
	VerificationNotes   string     `json:"verification_notes"`
	VerificationStatus  string     `json:"verification_status"`
	EvidenceRefs        []string   `json:"evidence_refs"`
	CostEstimate        *float64   `json:"cost_estimate,omitempty"`
	ActualCost          *float64   `json:"actual_cost,omitempty"`
	EffectivenessRating string     `json:"effectiveness_rating"`
	FollowUpRequired    bool       `json:"follow_up_required"`
	FollowUpDate        *time.Time `json:"follow_up_date,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// AuditSchedule is the result of risk-based selection, containing
// prioritised entities for audit within available budget days.
type AuditSchedule struct {
	ProgrammeID       uuid.UUID              `json:"programme_id"`
	TotalDaysAvailable int                   `json:"total_days_available"`
	TotalDaysAllocated int                   `json:"total_days_allocated"`
	ScheduledEntities  []ScheduledAuditEntity `json:"scheduled_entities"`
	DeferredEntities   []DeferredAuditEntity  `json:"deferred_entities"`
	GeneratedAt        time.Time              `json:"generated_at"`
}

// ScheduledAuditEntity is an entity scheduled for audit.
type ScheduledAuditEntity struct {
	EntityID        uuid.UUID `json:"entity_id"`
	EntityRef       string    `json:"entity_ref"`
	Name            string    `json:"name"`
	RiskRating      string    `json:"risk_rating"`
	RiskScore       float64   `json:"risk_score"`
	PriorityScore   float64   `json:"priority_score"`
	DaysSinceAudit  int       `json:"days_since_audit"`
	EstimatedDays   int       `json:"estimated_days"`
	SuggestedStart  time.Time `json:"suggested_start"`
}

// DeferredAuditEntity is an entity that could not be scheduled.
type DeferredAuditEntity struct {
	EntityID   uuid.UUID `json:"entity_id"`
	EntityRef  string    `json:"entity_ref"`
	Name       string    `json:"name"`
	RiskRating string    `json:"risk_rating"`
	Reason     string    `json:"reason"`
}

// ============================================================
// VALID ENGAGEMENT STATUS TRANSITIONS
// ============================================================

// validEngagementTransitions defines the allowed status transitions for engagements.
var validEngagementTransitions = map[string][]string{
	"planning":    {"fieldwork", "cancelled"},
	"fieldwork":   {"review", "cancelled"},
	"review":      {"reporting", "fieldwork"},
	"reporting":   {"completed", "review"},
	"completed":   {},
	"cancelled":   {},
}

// IsValidEngagementTransition checks if a status transition is allowed.
func IsValidEngagementTransition(from, to string) bool {
	allowed, ok := validEngagementTransitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}

// ============================================================
// SAMPLE SIZE CALCULATION — ISA 530
// ============================================================

// ConfidenceToZScore maps common confidence levels to their Z-scores.
var ConfidenceToZScore = map[float64]float64{
	99.0: 2.576,
	95.0: 1.960,
	90.0: 1.645,
	85.0: 1.440,
	80.0: 1.282,
}

// CalculateSampleSize computes sample size using the ISA 530 statistical formula:
//   n = (Z^2 * p * (1-p)) / E^2
// where Z is the Z-score for the confidence level, p is the expected error rate,
// and E is the tolerable error rate (margin of error).
// For finite populations, applies the finite population correction:
//   n_adj = n / (1 + (n-1)/N)
// The minimum returned sample size is 1.
func CalculateSampleSize(populationSize int, confidenceLevel, tolerableErrorRate, expectedErrorRate float64) int {
	// Look up Z-score; default to 95% confidence if not found.
	z, ok := ConfidenceToZScore[confidenceLevel]
	if !ok {
		z = 1.960
	}

	p := expectedErrorRate
	if p <= 0 {
		p = 0.01
	}
	if p >= 1 {
		p = 0.99
	}

	e := tolerableErrorRate
	if e <= 0 {
		e = 0.05
	}

	// ISA 530 formula: n = (Z^2 * p * (1-p)) / E^2
	n := (z * z * p * (1 - p)) / (e * e)

	// Apply finite population correction if population is known and finite.
	if populationSize > 0 {
		n = n / (1.0 + (n-1.0)/float64(populationSize))
	}

	result := int(math.Ceil(n))
	if result < 1 {
		result = 1
	}

	// Sample size cannot exceed population
	if populationSize > 0 && result > populationSize {
		result = populationSize
	}

	return result
}

// GenerateRandomSample selects sampleSize unique random indices from [0, populationSize)
// using crypto/rand for unbiased selection.
func GenerateRandomSample(populationSize, sampleSize int) ([]int, error) {
	if sampleSize <= 0 {
		return nil, fmt.Errorf("sample size must be positive")
	}
	if populationSize <= 0 {
		return nil, fmt.Errorf("population size must be positive")
	}
	if sampleSize > populationSize {
		sampleSize = populationSize
	}

	// Fisher-Yates partial shuffle using crypto/rand
	indices := make([]int, populationSize)
	for i := range indices {
		indices[i] = i
	}

	selected := make([]int, 0, sampleSize)
	for i := 0; i < sampleSize; i++ {
		remaining := populationSize - i
		randIdx, err := rand.Int(rand.Reader, big.NewInt(int64(remaining)))
		if err != nil {
			return nil, fmt.Errorf("crypto/rand failed: %w", err)
		}
		j := int(randIdx.Int64()) + i
		indices[i], indices[j] = indices[j], indices[i]
		selected = append(selected, indices[i])
	}

	sort.Ints(selected)
	return selected, nil
}

// ============================================================
// RISK-BASED SELECTION LOGIC
// ============================================================

// riskRatingWeight returns a numeric weight for risk-based prioritisation.
func riskRatingWeight(rating string) float64 {
	switch rating {
	case "critical":
		return 5.0
	case "high":
		return 4.0
	case "medium":
		return 3.0
	case "low":
		return 2.0
	case "very_low":
		return 1.0
	default:
		return 3.0
	}
}

// estimateAuditDays returns estimated audit effort based on risk rating.
func estimateAuditDays(riskRating string) int {
	switch riskRating {
	case "critical":
		return 10
	case "high":
		return 7
	case "medium":
		return 5
	case "low":
		return 3
	case "very_low":
		return 2
	default:
		return 5
	}
}

// ============================================================
// PROGRAMME CRUD
// ============================================================

// ListProgrammes returns a paginated list of audit programmes for an organisation.
func (s *AuditProgrammeService) ListProgrammes(ctx context.Context, orgID uuid.UUID, page, pageSize int) ([]AuditProgramme, int64, error) {
	offset := (page - 1) * pageSize

	var total int64
	err := s.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM audit_programmes WHERE organization_id = $1`, orgID,
	).Scan(&total)
	if err != nil {
		log.Error().Err(err).Msg("audit_programme: count failed")
		return nil, 0, err
	}

	rows, err := s.pool.Query(ctx,
		`SELECT id, organization_id, programme_ref, name, description, status,
		        programme_type, period_start, period_end, total_budget_days,
		        used_budget_days, objectives, risk_appetite, methodology,
		        approved_by, approved_at, owner_user_id, created_at, updated_at
		 FROM audit_programmes
		 WHERE organization_id = $1
		 ORDER BY period_start DESC
		 LIMIT $2 OFFSET $3`, orgID, pageSize, offset)
	if err != nil {
		log.Error().Err(err).Msg("audit_programme: list failed")
		return nil, 0, err
	}
	defer rows.Close()

	var programmes []AuditProgramme
	for rows.Next() {
		var p AuditProgramme
		if err := rows.Scan(
			&p.ID, &p.OrganizationID, &p.ProgrammeRef, &p.Name, &p.Description,
			&p.Status, &p.ProgrammeType, &p.PeriodStart, &p.PeriodEnd,
			&p.TotalBudgetDays, &p.UsedBudgetDays, &p.Objectives, &p.RiskAppetite,
			&p.Methodology, &p.ApprovedBy, &p.ApprovedAt, &p.OwnerUserID,
			&p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			log.Error().Err(err).Msg("audit_programme: scan failed")
			return nil, 0, err
		}
		programmes = append(programmes, p)
	}

	return programmes, total, nil
}

// CreateProgramme creates a new audit programme.
func (s *AuditProgrammeService) CreateProgramme(ctx context.Context, orgID uuid.UUID, req AuditProgramme) (*AuditProgramme, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return nil, fmt.Errorf("set tenant: %w", err)
	}

	// Generate programme ref
	var nextNum int
	err = tx.QueryRow(ctx,
		`SELECT COALESCE(MAX(CAST(SUBSTRING(programme_ref FROM 5) AS INT)), 0) + 1
		 FROM audit_programmes WHERE organization_id = $1`, orgID,
	).Scan(&nextNum)
	if err != nil {
		return nil, fmt.Errorf("generate ref: %w", err)
	}
	ref := fmt.Sprintf("APR-%04d", nextNum)

	var p AuditProgramme
	err = tx.QueryRow(ctx,
		`INSERT INTO audit_programmes
		 (organization_id, programme_ref, name, description, status, programme_type,
		  period_start, period_end, total_budget_days, objectives, risk_appetite,
		  methodology, owner_user_id)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		 RETURNING id, organization_id, programme_ref, name, description, status,
		           programme_type, period_start, period_end, total_budget_days,
		           used_budget_days, objectives, risk_appetite, methodology,
		           approved_by, approved_at, owner_user_id, created_at, updated_at`,
		orgID, ref, req.Name, req.Description, "draft", req.ProgrammeType,
		req.PeriodStart, req.PeriodEnd, req.TotalBudgetDays, req.Objectives,
		req.RiskAppetite, req.Methodology, req.OwnerUserID,
	).Scan(
		&p.ID, &p.OrganizationID, &p.ProgrammeRef, &p.Name, &p.Description,
		&p.Status, &p.ProgrammeType, &p.PeriodStart, &p.PeriodEnd,
		&p.TotalBudgetDays, &p.UsedBudgetDays, &p.Objectives, &p.RiskAppetite,
		&p.Methodology, &p.ApprovedBy, &p.ApprovedAt, &p.OwnerUserID,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		log.Error().Err(err).Msg("audit_programme: create failed")
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return &p, nil
}

// GetProgramme returns a single audit programme by ID.
func (s *AuditProgrammeService) GetProgramme(ctx context.Context, orgID, programmeID uuid.UUID) (*AuditProgramme, error) {
	var p AuditProgramme
	err := s.pool.QueryRow(ctx,
		`SELECT id, organization_id, programme_ref, name, description, status,
		        programme_type, period_start, period_end, total_budget_days,
		        used_budget_days, objectives, risk_appetite, methodology,
		        approved_by, approved_at, owner_user_id, created_at, updated_at
		 FROM audit_programmes
		 WHERE id = $1 AND organization_id = $2`, programmeID, orgID,
	).Scan(
		&p.ID, &p.OrganizationID, &p.ProgrammeRef, &p.Name, &p.Description,
		&p.Status, &p.ProgrammeType, &p.PeriodStart, &p.PeriodEnd,
		&p.TotalBudgetDays, &p.UsedBudgetDays, &p.Objectives, &p.RiskAppetite,
		&p.Methodology, &p.ApprovedBy, &p.ApprovedAt, &p.OwnerUserID,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// ============================================================
// RISK-BASED SELECTION
// ============================================================

// RiskBasedSelection generates a risk-prioritised audit schedule for a programme.
// It prioritises entities by risk rating weight and time since last audit, then
// schedules as many as fit within the available budget days.
func (s *AuditProgrammeService) RiskBasedSelection(ctx context.Context, orgID, programmeID uuid.UUID, totalDays int) (*AuditSchedule, error) {
	// Fetch all active entities from the audit universe.
	rows, err := s.pool.Query(ctx,
		`SELECT id, entity_ref, name, risk_rating, risk_score, last_audit_date,
		        audit_frequency_months
		 FROM audit_universe
		 WHERE organization_id = $1 AND status = 'active'
		 ORDER BY risk_score DESC, name ASC`, orgID)
	if err != nil {
		return nil, fmt.Errorf("query universe: %w", err)
	}
	defer rows.Close()

	type entityRow struct {
		ID                   uuid.UUID
		EntityRef            string
		Name                 string
		RiskRating           string
		RiskScore            float64
		LastAuditDate        *time.Time
		AuditFrequencyMonths int
	}

	var entities []entityRow
	for rows.Next() {
		var e entityRow
		if err := rows.Scan(&e.ID, &e.EntityRef, &e.Name, &e.RiskRating,
			&e.RiskScore, &e.LastAuditDate, &e.AuditFrequencyMonths); err != nil {
			return nil, fmt.Errorf("scan entity: %w", err)
		}
		entities = append(entities, e)
	}

	now := time.Now()
	type scored struct {
		entityRow
		priorityScore  float64
		daysSinceAudit int
		estimatedDays  int
	}

	var scoredEntities []scored
	for _, e := range entities {
		riskWeight := riskRatingWeight(e.RiskRating)

		daysSince := 365 * 2 // Default: assume 2 years if never audited
		if e.LastAuditDate != nil {
			daysSince = int(now.Sub(*e.LastAuditDate).Hours() / 24)
		}

		// Overdue bonus: if past the audit frequency, multiply time factor.
		overdueFactor := 1.0
		if e.AuditFrequencyMonths > 0 {
			dueAfterDays := e.AuditFrequencyMonths * 30
			if daysSince > dueAfterDays {
				overdueFactor = 1.5
			}
		}

		// Priority score = risk_weight * overdue_factor * (daysSinceAudit / 365)
		timeFactor := float64(daysSince) / 365.0
		priority := riskWeight * overdueFactor * timeFactor

		scoredEntities = append(scoredEntities, scored{
			entityRow:      e,
			priorityScore:  priority,
			daysSinceAudit: daysSince,
			estimatedDays:  estimateAuditDays(e.RiskRating),
		})
	}

	// Sort by priority score descending
	sort.Slice(scoredEntities, func(i, j int) bool {
		return scoredEntities[i].priorityScore > scoredEntities[j].priorityScore
	})

	schedule := &AuditSchedule{
		ProgrammeID:        programmeID,
		TotalDaysAvailable: totalDays,
		GeneratedAt:        now,
	}

	daysAllocated := 0
	// Fetch programme period for scheduling.
	prog, err := s.GetProgramme(ctx, orgID, programmeID)
	if err != nil {
		return nil, fmt.Errorf("get programme: %w", err)
	}
	currentStart := prog.PeriodStart

	for _, e := range scoredEntities {
		if daysAllocated+e.estimatedDays <= totalDays {
			suggestedStart := currentStart.AddDate(0, 0, daysAllocated)
			schedule.ScheduledEntities = append(schedule.ScheduledEntities, ScheduledAuditEntity{
				EntityID:       e.ID,
				EntityRef:      e.EntityRef,
				Name:           e.Name,
				RiskRating:     e.RiskRating,
				RiskScore:      e.RiskScore,
				PriorityScore:  e.priorityScore,
				DaysSinceAudit: e.daysSinceAudit,
				EstimatedDays:  e.estimatedDays,
				SuggestedStart: suggestedStart,
			})
			daysAllocated += e.estimatedDays
		} else {
			schedule.DeferredEntities = append(schedule.DeferredEntities, DeferredAuditEntity{
				EntityID:   e.ID,
				EntityRef:  e.EntityRef,
				Name:       e.Name,
				RiskRating: e.RiskRating,
				Reason:     "Insufficient budget days remaining",
			})
		}
	}
	schedule.TotalDaysAllocated = daysAllocated

	return schedule, nil
}

// ============================================================
// AUDIT UNIVERSE CRUD
// ============================================================

// ListAuditUniverse returns a paginated list of auditable entities.
func (s *AuditProgrammeService) ListAuditUniverse(ctx context.Context, orgID uuid.UUID, page, pageSize int) ([]AuditableEntity, int64, error) {
	offset := (page - 1) * pageSize

	var total int64
	err := s.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM audit_universe WHERE organization_id = $1`, orgID,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := s.pool.Query(ctx,
		`SELECT id, organization_id, entity_ref, name, entity_type, description,
		        risk_rating, risk_score, business_owner_id, department, location,
		        regulatory_relevance, last_audit_date, next_audit_due,
		        audit_frequency_months, status, linked_framework_ids,
		        linked_risk_ids, created_at, updated_at
		 FROM audit_universe
		 WHERE organization_id = $1
		 ORDER BY risk_score DESC, name ASC
		 LIMIT $2 OFFSET $3`, orgID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var entities []AuditableEntity
	for rows.Next() {
		var e AuditableEntity
		if err := rows.Scan(
			&e.ID, &e.OrganizationID, &e.EntityRef, &e.Name, &e.EntityType,
			&e.Description, &e.RiskRating, &e.RiskScore, &e.BusinessOwnerID,
			&e.Department, &e.Location, &e.RegulatoryRelevance, &e.LastAuditDate,
			&e.NextAuditDue, &e.AuditFrequencyMonths, &e.Status,
			&e.LinkedFrameworkIDs, &e.LinkedRiskIDs, &e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		entities = append(entities, e)
	}

	return entities, total, nil
}

// CreateAuditableEntity adds a new entity to the audit universe.
func (s *AuditProgrammeService) CreateAuditableEntity(ctx context.Context, orgID uuid.UUID, req AuditableEntity) (*AuditableEntity, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return nil, fmt.Errorf("set tenant: %w", err)
	}

	var nextNum int
	err = tx.QueryRow(ctx,
		`SELECT COALESCE(MAX(CAST(SUBSTRING(entity_ref FROM 4) AS INT)), 0) + 1
		 FROM audit_universe WHERE organization_id = $1`, orgID,
	).Scan(&nextNum)
	if err != nil {
		return nil, fmt.Errorf("generate ref: %w", err)
	}
	ref := fmt.Sprintf("AU-%04d", nextNum)

	var e AuditableEntity
	err = tx.QueryRow(ctx,
		`INSERT INTO audit_universe
		 (organization_id, entity_ref, name, entity_type, description, risk_rating,
		  risk_score, business_owner_id, department, location, regulatory_relevance,
		  last_audit_date, next_audit_due, audit_frequency_months, status,
		  linked_framework_ids, linked_risk_ids)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		 RETURNING id, organization_id, entity_ref, name, entity_type, description,
		           risk_rating, risk_score, business_owner_id, department, location,
		           regulatory_relevance, last_audit_date, next_audit_due,
		           audit_frequency_months, status, linked_framework_ids,
		           linked_risk_ids, created_at, updated_at`,
		orgID, ref, req.Name, req.EntityType, req.Description, req.RiskRating,
		req.RiskScore, req.BusinessOwnerID, req.Department, req.Location,
		req.RegulatoryRelevance, req.LastAuditDate, req.NextAuditDue,
		req.AuditFrequencyMonths, "active", req.LinkedFrameworkIDs, req.LinkedRiskIDs,
	).Scan(
		&e.ID, &e.OrganizationID, &e.EntityRef, &e.Name, &e.EntityType,
		&e.Description, &e.RiskRating, &e.RiskScore, &e.BusinessOwnerID,
		&e.Department, &e.Location, &e.RegulatoryRelevance, &e.LastAuditDate,
		&e.NextAuditDue, &e.AuditFrequencyMonths, &e.Status,
		&e.LinkedFrameworkIDs, &e.LinkedRiskIDs, &e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		log.Error().Err(err).Msg("audit_universe: create failed")
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return &e, nil
}

// UpdateAuditableEntity updates an existing auditable entity.
func (s *AuditProgrammeService) UpdateAuditableEntity(ctx context.Context, orgID, entityID uuid.UUID, req AuditableEntity) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return fmt.Errorf("set tenant: %w", err)
	}

	_, err = tx.Exec(ctx,
		`UPDATE audit_universe SET
		    name = $3, entity_type = $4, description = $5, risk_rating = $6,
		    risk_score = $7, business_owner_id = $8, department = $9, location = $10,
		    regulatory_relevance = $11, last_audit_date = $12, next_audit_due = $13,
		    audit_frequency_months = $14, status = $15
		 WHERE id = $1 AND organization_id = $2`,
		entityID, orgID, req.Name, req.EntityType, req.Description, req.RiskRating,
		req.RiskScore, req.BusinessOwnerID, req.Department, req.Location,
		req.RegulatoryRelevance, req.LastAuditDate, req.NextAuditDue,
		req.AuditFrequencyMonths, req.Status,
	)
	if err != nil {
		return fmt.Errorf("update entity: %w", err)
	}

	return tx.Commit(ctx)
}

// ============================================================
// ENGAGEMENT CRUD
// ============================================================

// ListEngagements returns engagements, optionally filtered by programme ID.
func (s *AuditProgrammeService) ListEngagements(ctx context.Context, orgID uuid.UUID, programmeID *uuid.UUID, page, pageSize int) ([]AuditEngagement, int64, error) {
	offset := (page - 1) * pageSize

	var total int64
	var rows pgx.Rows
	var err error

	if programmeID != nil {
		err = s.pool.QueryRow(ctx,
			`SELECT COUNT(*) FROM audit_engagements WHERE organization_id = $1 AND programme_id = $2`,
			orgID, *programmeID,
		).Scan(&total)
		if err != nil {
			return nil, 0, err
		}
		rows, err = s.pool.Query(ctx,
			`SELECT id, organization_id, programme_id, audit_id, auditable_entity_id,
			        engagement_ref, name, engagement_type, status, priority, risk_rating,
			        scope, objectives, methodology, lead_auditor_id, audit_team_ids,
			        planned_start_date, planned_end_date, actual_start_date, actual_end_date,
			        budget_days, actual_days, fieldwork_complete, report_issued,
			        report_issued_date, overall_opinion, created_at, updated_at
			 FROM audit_engagements
			 WHERE organization_id = $1 AND programme_id = $2
			 ORDER BY created_at DESC
			 LIMIT $3 OFFSET $4`, orgID, *programmeID, pageSize, offset)
	} else {
		err = s.pool.QueryRow(ctx,
			`SELECT COUNT(*) FROM audit_engagements WHERE organization_id = $1`, orgID,
		).Scan(&total)
		if err != nil {
			return nil, 0, err
		}
		rows, err = s.pool.Query(ctx,
			`SELECT id, organization_id, programme_id, audit_id, auditable_entity_id,
			        engagement_ref, name, engagement_type, status, priority, risk_rating,
			        scope, objectives, methodology, lead_auditor_id, audit_team_ids,
			        planned_start_date, planned_end_date, actual_start_date, actual_end_date,
			        budget_days, actual_days, fieldwork_complete, report_issued,
			        report_issued_date, overall_opinion, created_at, updated_at
			 FROM audit_engagements
			 WHERE organization_id = $1
			 ORDER BY created_at DESC
			 LIMIT $2 OFFSET $3`, orgID, pageSize, offset)
	}
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var engagements []AuditEngagement
	for rows.Next() {
		var e AuditEngagement
		if err := rows.Scan(
			&e.ID, &e.OrganizationID, &e.ProgrammeID, &e.AuditID, &e.AuditableEntityID,
			&e.EngagementRef, &e.Name, &e.EngagementType, &e.Status, &e.Priority,
			&e.RiskRating, &e.Scope, &e.Objectives, &e.Methodology, &e.LeadAuditorID,
			&e.AuditTeamIDs, &e.PlannedStartDate, &e.PlannedEndDate,
			&e.ActualStartDate, &e.ActualEndDate, &e.BudgetDays, &e.ActualDays,
			&e.FieldworkComplete, &e.ReportIssued, &e.ReportIssuedDate,
			&e.OverallOpinion, &e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		engagements = append(engagements, e)
	}

	return engagements, total, nil
}

// CreateEngagement creates a new audit engagement within a programme.
func (s *AuditProgrammeService) CreateEngagement(ctx context.Context, orgID uuid.UUID, req AuditEngagement) (*AuditEngagement, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return nil, fmt.Errorf("set tenant: %w", err)
	}

	var nextNum int
	err = tx.QueryRow(ctx,
		`SELECT COALESCE(MAX(CAST(SUBSTRING(engagement_ref FROM 5) AS INT)), 0) + 1
		 FROM audit_engagements WHERE organization_id = $1`, orgID,
	).Scan(&nextNum)
	if err != nil {
		return nil, fmt.Errorf("generate ref: %w", err)
	}
	ref := fmt.Sprintf("AEN-%04d", nextNum)

	var e AuditEngagement
	err = tx.QueryRow(ctx,
		`INSERT INTO audit_engagements
		 (organization_id, programme_id, audit_id, auditable_entity_id,
		  engagement_ref, name, engagement_type, status, priority, risk_rating,
		  scope, objectives, methodology, lead_auditor_id, audit_team_ids,
		  planned_start_date, planned_end_date, budget_days)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)
		 RETURNING id, organization_id, programme_id, audit_id, auditable_entity_id,
		           engagement_ref, name, engagement_type, status, priority, risk_rating,
		           scope, objectives, methodology, lead_auditor_id, audit_team_ids,
		           planned_start_date, planned_end_date, actual_start_date, actual_end_date,
		           budget_days, actual_days, fieldwork_complete, report_issued,
		           report_issued_date, overall_opinion, created_at, updated_at`,
		orgID, req.ProgrammeID, req.AuditID, req.AuditableEntityID,
		ref, req.Name, req.EngagementType, "planning", req.Priority, req.RiskRating,
		req.Scope, req.Objectives, req.Methodology, req.LeadAuditorID,
		req.AuditTeamIDs, req.PlannedStartDate, req.PlannedEndDate, req.BudgetDays,
	).Scan(
		&e.ID, &e.OrganizationID, &e.ProgrammeID, &e.AuditID, &e.AuditableEntityID,
		&e.EngagementRef, &e.Name, &e.EngagementType, &e.Status, &e.Priority,
		&e.RiskRating, &e.Scope, &e.Objectives, &e.Methodology, &e.LeadAuditorID,
		&e.AuditTeamIDs, &e.PlannedStartDate, &e.PlannedEndDate,
		&e.ActualStartDate, &e.ActualEndDate, &e.BudgetDays, &e.ActualDays,
		&e.FieldworkComplete, &e.ReportIssued, &e.ReportIssuedDate,
		&e.OverallOpinion, &e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		log.Error().Err(err).Msg("audit_engagement: create failed")
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return &e, nil
}

// GetEngagement returns a single audit engagement.
func (s *AuditProgrammeService) GetEngagement(ctx context.Context, orgID, engagementID uuid.UUID) (*AuditEngagement, error) {
	var e AuditEngagement
	err := s.pool.QueryRow(ctx,
		`SELECT id, organization_id, programme_id, audit_id, auditable_entity_id,
		        engagement_ref, name, engagement_type, status, priority, risk_rating,
		        scope, objectives, methodology, lead_auditor_id, audit_team_ids,
		        planned_start_date, planned_end_date, actual_start_date, actual_end_date,
		        budget_days, actual_days, fieldwork_complete, report_issued,
		        report_issued_date, overall_opinion, created_at, updated_at
		 FROM audit_engagements
		 WHERE id = $1 AND organization_id = $2`, engagementID, orgID,
	).Scan(
		&e.ID, &e.OrganizationID, &e.ProgrammeID, &e.AuditID, &e.AuditableEntityID,
		&e.EngagementRef, &e.Name, &e.EngagementType, &e.Status, &e.Priority,
		&e.RiskRating, &e.Scope, &e.Objectives, &e.Methodology, &e.LeadAuditorID,
		&e.AuditTeamIDs, &e.PlannedStartDate, &e.PlannedEndDate,
		&e.ActualStartDate, &e.ActualEndDate, &e.BudgetDays, &e.ActualDays,
		&e.FieldworkComplete, &e.ReportIssued, &e.ReportIssuedDate,
		&e.OverallOpinion, &e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// UpdateEngagementStatus transitions the engagement to a new status if valid.
func (s *AuditProgrammeService) UpdateEngagementStatus(ctx context.Context, orgID, engagementID uuid.UUID, newStatus string) error {
	eng, err := s.GetEngagement(ctx, orgID, engagementID)
	if err != nil {
		return fmt.Errorf("get engagement: %w", err)
	}

	if !IsValidEngagementTransition(eng.Status, newStatus) {
		return fmt.Errorf("invalid status transition from %q to %q", eng.Status, newStatus)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return fmt.Errorf("set tenant: %w", err)
	}

	now := time.Now()
	switch newStatus {
	case "fieldwork":
		_, err = tx.Exec(ctx,
			`UPDATE audit_engagements SET status = $3, actual_start_date = $4
			 WHERE id = $1 AND organization_id = $2`,
			engagementID, orgID, newStatus, now)
	case "completed":
		_, err = tx.Exec(ctx,
			`UPDATE audit_engagements SET status = $3, actual_end_date = $4, fieldwork_complete = true
			 WHERE id = $1 AND organization_id = $2`,
			engagementID, orgID, newStatus, now)
	default:
		_, err = tx.Exec(ctx,
			`UPDATE audit_engagements SET status = $3
			 WHERE id = $1 AND organization_id = $2`,
			engagementID, orgID, newStatus)
	}
	if err != nil {
		return fmt.Errorf("update status: %w", err)
	}

	return tx.Commit(ctx)
}

// ============================================================
// WORKPAPERS
// ============================================================

// ListWorkpapers returns all workpapers for a given engagement.
func (s *AuditProgrammeService) ListWorkpapers(ctx context.Context, orgID, engagementID uuid.UUID) ([]AuditWorkpaper, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, organization_id, engagement_id, workpaper_ref, title,
		        description, workpaper_type, status, content, prepared_by,
		        prepared_date, reviewed_by, reviewed_date, review_comments,
		        linked_control_ids, linked_risk_ids, created_at, updated_at
		 FROM audit_workpapers
		 WHERE engagement_id = $1 AND organization_id = $2
		 ORDER BY workpaper_ref ASC`, engagementID, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workpapers []AuditWorkpaper
	for rows.Next() {
		var w AuditWorkpaper
		if err := rows.Scan(
			&w.ID, &w.OrganizationID, &w.EngagementID, &w.WorkpaperRef, &w.Title,
			&w.Description, &w.WorkpaperType, &w.Status, &w.Content, &w.PreparedBy,
			&w.PreparedDate, &w.ReviewedBy, &w.ReviewedDate, &w.ReviewComments,
			&w.LinkedControlIDs, &w.LinkedRiskIDs, &w.CreatedAt, &w.UpdatedAt,
		); err != nil {
			return nil, err
		}
		workpapers = append(workpapers, w)
	}

	return workpapers, nil
}

// CreateWorkpaper adds a new workpaper to an engagement.
func (s *AuditProgrammeService) CreateWorkpaper(ctx context.Context, orgID uuid.UUID, req AuditWorkpaper) (*AuditWorkpaper, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return nil, fmt.Errorf("set tenant: %w", err)
	}

	var nextNum int
	err = tx.QueryRow(ctx,
		`SELECT COALESCE(MAX(CAST(SUBSTRING(workpaper_ref FROM 4) AS INT)), 0) + 1
		 FROM audit_workpapers WHERE organization_id = $1`, orgID,
	).Scan(&nextNum)
	if err != nil {
		return nil, fmt.Errorf("generate ref: %w", err)
	}
	ref := fmt.Sprintf("WP-%04d", nextNum)

	var w AuditWorkpaper
	err = tx.QueryRow(ctx,
		`INSERT INTO audit_workpapers
		 (organization_id, engagement_id, workpaper_ref, title, description,
		  workpaper_type, status, content, prepared_by, linked_control_ids, linked_risk_ids)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		 RETURNING id, organization_id, engagement_id, workpaper_ref, title,
		           description, workpaper_type, status, content, prepared_by,
		           prepared_date, reviewed_by, reviewed_date, review_comments,
		           linked_control_ids, linked_risk_ids, created_at, updated_at`,
		orgID, req.EngagementID, ref, req.Title, req.Description,
		req.WorkpaperType, "draft", req.Content, req.PreparedBy,
		req.LinkedControlIDs, req.LinkedRiskIDs,
	).Scan(
		&w.ID, &w.OrganizationID, &w.EngagementID, &w.WorkpaperRef, &w.Title,
		&w.Description, &w.WorkpaperType, &w.Status, &w.Content, &w.PreparedBy,
		&w.PreparedDate, &w.ReviewedBy, &w.ReviewedDate, &w.ReviewComments,
		&w.LinkedControlIDs, &w.LinkedRiskIDs, &w.CreatedAt, &w.UpdatedAt,
	)
	if err != nil {
		log.Error().Err(err).Msg("audit_workpaper: create failed")
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return &w, nil
}

// UpdateWorkpaper updates an existing workpaper.
func (s *AuditProgrammeService) UpdateWorkpaper(ctx context.Context, orgID, workpaperID uuid.UUID, req AuditWorkpaper) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE audit_workpapers SET
		    title = $3, description = $4, workpaper_type = $5, content = $6,
		    linked_control_ids = $7, linked_risk_ids = $8
		 WHERE id = $1 AND organization_id = $2`,
		workpaperID, orgID, req.Title, req.Description, req.WorkpaperType,
		req.Content, req.LinkedControlIDs, req.LinkedRiskIDs,
	)
	return err
}

// SubmitForReview transitions a workpaper to review status, enforcing four-eyes:
// the reviewer must be different from the preparer.
func (s *AuditProgrammeService) SubmitForReview(ctx context.Context, orgID, workpaperID, reviewerID uuid.UUID) error {
	// Fetch the workpaper to check preparer.
	var preparedBy uuid.UUID
	err := s.pool.QueryRow(ctx,
		`SELECT prepared_by FROM audit_workpapers WHERE id = $1 AND organization_id = $2`,
		workpaperID, orgID,
	).Scan(&preparedBy)
	if err != nil {
		return fmt.Errorf("workpaper not found: %w", err)
	}

	// Four-eyes principle: reviewer must differ from preparer.
	if reviewerID == preparedBy {
		return fmt.Errorf("four-eyes violation: reviewer cannot be the same as the preparer")
	}

	now := time.Now()
	_, err = s.pool.Exec(ctx,
		`UPDATE audit_workpapers SET status = 'under_review', reviewed_by = $3, reviewed_date = $4
		 WHERE id = $1 AND organization_id = $2`,
		workpaperID, orgID, reviewerID, now,
	)
	return err
}

// ============================================================
// SAMPLING — ISA 530
// ============================================================

// GenerateSample creates a statistical audit sample using ISA 530 methodology.
func (s *AuditProgrammeService) GenerateSample(ctx context.Context, orgID, engagementID uuid.UUID, userID uuid.UUID, config SampleConfig) (*AuditSample, error) {
	sampleSize := CalculateSampleSize(
		config.PopulationSize,
		config.ConfidenceLevel,
		config.TolerableErrorRate,
		config.ExpectedErrorRate,
	)

	indices, err := GenerateRandomSample(config.PopulationSize, sampleSize)
	if err != nil {
		return nil, fmt.Errorf("random selection: %w", err)
	}

	items := make([]SampleItem, len(indices))
	for i, idx := range indices {
		items[i] = SampleItem{
			Index:  idx,
			ItemID: fmt.Sprintf("ITEM-%04d", idx+1),
			Result: "pending",
		}
	}

	itemsJSON, err := json.Marshal(items)
	if err != nil {
		return nil, fmt.Errorf("marshal items: %w", err)
	}

	z := ConfidenceToZScore[config.ConfidenceLevel]
	if z == 0 {
		z = 1.960
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return nil, fmt.Errorf("set tenant: %w", err)
	}

	var nextNum int
	err = tx.QueryRow(ctx,
		`SELECT COALESCE(MAX(CAST(SUBSTRING(sample_ref FROM 5) AS INT)), 0) + 1
		 FROM audit_samples WHERE organization_id = $1`, orgID,
	).Scan(&nextNum)
	if err != nil {
		return nil, fmt.Errorf("generate ref: %w", err)
	}
	ref := fmt.Sprintf("SMP-%04d", nextNum)

	var sample AuditSample
	err = tx.QueryRow(ctx,
		`INSERT INTO audit_samples
		 (organization_id, engagement_id, sample_ref, name, description,
		  sampling_method, population_size, sample_size, confidence_level,
		  tolerable_error_rate, expected_error_rate, z_score, selection_method,
		  selected_items, status, created_by)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
		 RETURNING id, organization_id, engagement_id, sample_ref, name, description,
		           sampling_method, population_size, sample_size, confidence_level,
		           tolerable_error_rate, expected_error_rate, z_score, selection_method,
		           items_tested, items_passed, items_failed, items_inconclusive,
		           actual_error_rate, conclusion, status, created_by, created_at, updated_at`,
		orgID, engagementID, ref, config.Name, config.Description,
		"statistical", config.PopulationSize, sampleSize, config.ConfidenceLevel,
		config.TolerableErrorRate, config.ExpectedErrorRate, z, "random",
		itemsJSON, "pending", userID,
	).Scan(
		&sample.ID, &sample.OrganizationID, &sample.EngagementID, &sample.SampleRef,
		&sample.Name, &sample.Description, &sample.SamplingMethod, &sample.PopulationSize,
		&sample.SampleSize, &sample.ConfidenceLevel, &sample.TolerableErrorRate,
		&sample.ExpectedErrorRate, &sample.ZScore, &sample.SelectionMethod,
		&sample.ItemsTested, &sample.ItemsPassed, &sample.ItemsFailed,
		&sample.ItemsInconclusive, &sample.ActualErrorRate, &sample.Conclusion,
		&sample.Status, &sample.CreatedBy, &sample.CreatedAt, &sample.UpdatedAt,
	)
	if err != nil {
		log.Error().Err(err).Msg("audit_sample: create failed")
		return nil, err
	}

	sample.SelectedItems = items

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return &sample, nil
}

// GetSample returns a single audit sample by ID with its items.
func (s *AuditProgrammeService) GetSample(ctx context.Context, orgID, sampleID uuid.UUID) (*AuditSample, error) {
	var sample AuditSample
	var itemsJSON []byte

	err := s.pool.QueryRow(ctx,
		`SELECT id, organization_id, engagement_id, sample_ref, name, description,
		        sampling_method, population_size, sample_size, confidence_level,
		        tolerable_error_rate, expected_error_rate, z_score, selection_method,
		        selected_items, items_tested, items_passed, items_failed,
		        items_inconclusive, actual_error_rate, conclusion, status,
		        created_by, created_at, updated_at
		 FROM audit_samples
		 WHERE id = $1 AND organization_id = $2`, sampleID, orgID,
	).Scan(
		&sample.ID, &sample.OrganizationID, &sample.EngagementID, &sample.SampleRef,
		&sample.Name, &sample.Description, &sample.SamplingMethod, &sample.PopulationSize,
		&sample.SampleSize, &sample.ConfidenceLevel, &sample.TolerableErrorRate,
		&sample.ExpectedErrorRate, &sample.ZScore, &sample.SelectionMethod,
		&itemsJSON, &sample.ItemsTested, &sample.ItemsPassed, &sample.ItemsFailed,
		&sample.ItemsInconclusive, &sample.ActualErrorRate, &sample.Conclusion,
		&sample.Status, &sample.CreatedBy, &sample.CreatedAt, &sample.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(itemsJSON, &sample.SelectedItems); err != nil {
		log.Warn().Err(err).Msg("audit_sample: unmarshal items failed")
		sample.SelectedItems = []SampleItem{}
	}

	return &sample, nil
}

// RecordSampleItemResult updates a single item's test result within a sample.
func (s *AuditProgrammeService) RecordSampleItemResult(ctx context.Context, orgID, sampleID uuid.UUID, itemIndex int, result, notes string) error {
	sample, err := s.GetSample(ctx, orgID, sampleID)
	if err != nil {
		return fmt.Errorf("get sample: %w", err)
	}

	if itemIndex < 0 || itemIndex >= len(sample.SelectedItems) {
		return fmt.Errorf("item index %d out of range [0, %d)", itemIndex, len(sample.SelectedItems))
	}

	// Update the item result
	prevResult := sample.SelectedItems[itemIndex].Result
	sample.SelectedItems[itemIndex].Result = result
	sample.SelectedItems[itemIndex].Notes = notes
	sample.SelectedItems[itemIndex].TestedAt = time.Now().Format(time.RFC3339)

	// Recount
	tested, passed, failed, inconclusive := 0, 0, 0, 0
	for _, item := range sample.SelectedItems {
		switch item.Result {
		case "pass":
			tested++
			passed++
		case "fail":
			tested++
			failed++
		case "inconclusive":
			tested++
			inconclusive++
		}
	}

	itemsJSON, err := json.Marshal(sample.SelectedItems)
	if err != nil {
		return fmt.Errorf("marshal items: %w", err)
	}

	var actualErrorRate *float64
	if tested > 0 {
		rate := float64(failed) / float64(tested)
		actualErrorRate = &rate
	}

	status := "in_progress"
	if tested == len(sample.SelectedItems) {
		status = "completed"
	}

	_, err = s.pool.Exec(ctx,
		`UPDATE audit_samples SET
		    selected_items = $3, items_tested = $4, items_passed = $5,
		    items_failed = $6, items_inconclusive = $7, actual_error_rate = $8,
		    status = $9
		 WHERE id = $1 AND organization_id = $2`,
		sampleID, orgID, itemsJSON, tested, passed, failed, inconclusive,
		actualErrorRate, status,
	)
	if err != nil {
		return fmt.Errorf("update sample: %w", err)
	}

	_ = prevResult // consumed
	return nil
}

// ============================================================
// TEST PROCEDURES
// ============================================================

// CreateTestProcedure creates a new test procedure within an engagement.
func (s *AuditProgrammeService) CreateTestProcedure(ctx context.Context, orgID uuid.UUID, req AuditTestProcedure) (*AuditTestProcedure, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return nil, fmt.Errorf("set tenant: %w", err)
	}

	var nextNum int
	err = tx.QueryRow(ctx,
		`SELECT COALESCE(MAX(CAST(SUBSTRING(procedure_ref FROM 4) AS INT)), 0) + 1
		 FROM audit_test_procedures WHERE organization_id = $1`, orgID,
	).Scan(&nextNum)
	if err != nil {
		return nil, fmt.Errorf("generate ref: %w", err)
	}
	ref := fmt.Sprintf("TP-%04d", nextNum)

	var tp AuditTestProcedure
	err = tx.QueryRow(ctx,
		`INSERT INTO audit_test_procedures
		 (organization_id, engagement_id, procedure_ref, title, description,
		  test_type, control_id, control_ref, expected_result, evidence_refs, notes)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		 RETURNING id, organization_id, engagement_id, procedure_ref, title,
		           description, test_type, control_id, control_ref, expected_result,
		           actual_result, result, tested_by, tested_date, workpaper_id,
		           sample_id, finding_id, evidence_refs, notes, created_at, updated_at`,
		orgID, req.EngagementID, ref, req.Title, req.Description,
		req.TestType, req.ControlID, req.ControlRef, req.ExpectedResult,
		req.EvidenceRefs, req.Notes,
	).Scan(
		&tp.ID, &tp.OrganizationID, &tp.EngagementID, &tp.ProcedureRef, &tp.Title,
		&tp.Description, &tp.TestType, &tp.ControlID, &tp.ControlRef,
		&tp.ExpectedResult, &tp.ActualResult, &tp.Result, &tp.TestedBy,
		&tp.TestedDate, &tp.WorkpaperID, &tp.SampleID, &tp.FindingID,
		&tp.EvidenceRefs, &tp.Notes, &tp.CreatedAt, &tp.UpdatedAt,
	)
	if err != nil {
		log.Error().Err(err).Msg("audit_test_procedure: create failed")
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return &tp, nil
}

// RecordTestResult records the actual result of a test procedure.
func (s *AuditProgrammeService) RecordTestResult(ctx context.Context, orgID, procedureID uuid.UUID, actualResult, result string, testedBy uuid.UUID) error {
	now := time.Now()
	_, err := s.pool.Exec(ctx,
		`UPDATE audit_test_procedures SET
		    actual_result = $3, result = $4, tested_by = $5, tested_date = $6
		 WHERE id = $1 AND organization_id = $2`,
		procedureID, orgID, actualResult, result, testedBy, now,
	)
	return err
}

// ============================================================
// CORRECTIVE ACTIONS
// ============================================================

// ListCorrectiveActions returns a paginated list of corrective actions for the org.
func (s *AuditProgrammeService) ListCorrectiveActions(ctx context.Context, orgID uuid.UUID, status, priority string, page, pageSize int) ([]AuditCorrectiveAction, int64, error) {
	offset := (page - 1) * pageSize

	query := `FROM audit_corrective_actions WHERE organization_id = $1`
	args := []interface{}{orgID}
	argIdx := 2

	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}
	if priority != "" {
		query += fmt.Sprintf(" AND priority = $%d", argIdx)
		args = append(args, priority)
		argIdx++
	}

	var total int64
	countQuery := "SELECT COUNT(*) " + query
	err := s.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	selectQuery := fmt.Sprintf(
		`SELECT id, organization_id, finding_id, engagement_id, action_ref, title,
		        description, action_type, priority, status, root_cause, planned_action,
		        actual_action, responsible_user_id, implementer_user_id, due_date,
		        completed_date, verified_by, verified_date, verification_notes,
		        verification_status, evidence_refs, cost_estimate, actual_cost,
		        effectiveness_rating, follow_up_required, follow_up_date,
		        created_at, updated_at
		 %s ORDER BY due_date ASC NULLS LAST LIMIT $%d OFFSET $%d`,
		query, argIdx, argIdx+1)
	args = append(args, pageSize, offset)

	rows, err := s.pool.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var actions []AuditCorrectiveAction
	for rows.Next() {
		var a AuditCorrectiveAction
		if err := rows.Scan(
			&a.ID, &a.OrganizationID, &a.FindingID, &a.EngagementID, &a.ActionRef,
			&a.Title, &a.Description, &a.ActionType, &a.Priority, &a.Status,
			&a.RootCause, &a.PlannedAction, &a.ActualAction, &a.ResponsibleUserID,
			&a.ImplementerUserID, &a.DueDate, &a.CompletedDate, &a.VerifiedBy,
			&a.VerifiedDate, &a.VerificationNotes, &a.VerificationStatus,
			&a.EvidenceRefs, &a.CostEstimate, &a.ActualCost,
			&a.EffectivenessRating, &a.FollowUpRequired, &a.FollowUpDate,
			&a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		actions = append(actions, a)
	}

	return actions, total, nil
}

// CreateCorrectiveAction creates a new corrective action.
func (s *AuditProgrammeService) CreateCorrectiveAction(ctx context.Context, orgID uuid.UUID, req AuditCorrectiveAction) (*AuditCorrectiveAction, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return nil, fmt.Errorf("set tenant: %w", err)
	}

	var nextNum int
	err = tx.QueryRow(ctx,
		`SELECT COALESCE(MAX(CAST(SUBSTRING(action_ref FROM 4) AS INT)), 0) + 1
		 FROM audit_corrective_actions WHERE organization_id = $1`, orgID,
	).Scan(&nextNum)
	if err != nil {
		return nil, fmt.Errorf("generate ref: %w", err)
	}
	ref := fmt.Sprintf("CA-%04d", nextNum)

	var a AuditCorrectiveAction
	err = tx.QueryRow(ctx,
		`INSERT INTO audit_corrective_actions
		 (organization_id, finding_id, engagement_id, action_ref, title,
		  description, action_type, priority, status, root_cause, planned_action,
		  responsible_user_id, implementer_user_id, due_date, evidence_refs,
		  cost_estimate)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
		 RETURNING id, organization_id, finding_id, engagement_id, action_ref, title,
		           description, action_type, priority, status, root_cause, planned_action,
		           actual_action, responsible_user_id, implementer_user_id, due_date,
		           completed_date, verified_by, verified_date, verification_notes,
		           verification_status, evidence_refs, cost_estimate, actual_cost,
		           effectiveness_rating, follow_up_required, follow_up_date,
		           created_at, updated_at`,
		orgID, req.FindingID, req.EngagementID, ref, req.Title,
		req.Description, req.ActionType, req.Priority, "open", req.RootCause,
		req.PlannedAction, req.ResponsibleUserID, req.ImplementerUserID,
		req.DueDate, req.EvidenceRefs, req.CostEstimate,
	).Scan(
		&a.ID, &a.OrganizationID, &a.FindingID, &a.EngagementID, &a.ActionRef,
		&a.Title, &a.Description, &a.ActionType, &a.Priority, &a.Status,
		&a.RootCause, &a.PlannedAction, &a.ActualAction, &a.ResponsibleUserID,
		&a.ImplementerUserID, &a.DueDate, &a.CompletedDate, &a.VerifiedBy,
		&a.VerifiedDate, &a.VerificationNotes, &a.VerificationStatus,
		&a.EvidenceRefs, &a.CostEstimate, &a.ActualCost,
		&a.EffectivenessRating, &a.FollowUpRequired, &a.FollowUpDate,
		&a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		log.Error().Err(err).Msg("audit_corrective_action: create failed")
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return &a, nil
}

// UpdateCorrectiveAction updates an existing corrective action.
func (s *AuditProgrammeService) UpdateCorrectiveAction(ctx context.Context, orgID, actionID uuid.UUID, req AuditCorrectiveAction) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE audit_corrective_actions SET
		    title = $3, description = $4, action_type = $5, priority = $6,
		    status = $7, root_cause = $8, planned_action = $9, actual_action = $10,
		    responsible_user_id = $11, implementer_user_id = $12, due_date = $13,
		    completed_date = $14, evidence_refs = $15, cost_estimate = $16,
		    actual_cost = $17, effectiveness_rating = $18, follow_up_required = $19,
		    follow_up_date = $20
		 WHERE id = $1 AND organization_id = $2`,
		actionID, orgID, req.Title, req.Description, req.ActionType, req.Priority,
		req.Status, req.RootCause, req.PlannedAction, req.ActualAction,
		req.ResponsibleUserID, req.ImplementerUserID, req.DueDate,
		req.CompletedDate, req.EvidenceRefs, req.CostEstimate,
		req.ActualCost, req.EffectivenessRating, req.FollowUpRequired,
		req.FollowUpDate,
	)
	return err
}

// VerifyCorrectiveAction marks a corrective action as verified, enforcing that
// the verifier is different from the implementer (four-eyes principle).
func (s *AuditProgrammeService) VerifyCorrectiveAction(ctx context.Context, orgID, actionID, verifierID uuid.UUID, notes, status string) error {
	// Fetch the current action to enforce four-eyes.
	var implementerID *uuid.UUID
	err := s.pool.QueryRow(ctx,
		`SELECT implementer_user_id FROM audit_corrective_actions
		 WHERE id = $1 AND organization_id = $2`, actionID, orgID,
	).Scan(&implementerID)
	if err != nil {
		return fmt.Errorf("action not found: %w", err)
	}

	if implementerID != nil && *implementerID == verifierID {
		return fmt.Errorf("four-eyes violation: verifier cannot be the same as the implementer")
	}

	now := time.Now()
	_, err = s.pool.Exec(ctx,
		`UPDATE audit_corrective_actions SET
		    verified_by = $3, verified_date = $4, verification_notes = $5,
		    verification_status = $6, status = CASE WHEN $6 = 'verified' THEN 'closed' ELSE status END
		 WHERE id = $1 AND organization_id = $2`,
		actionID, orgID, verifierID, now, notes, status,
	)
	return err
}
