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
// BIAService
// ============================================================

// BIAService implements business logic for Business Impact Analysis,
// covering process CRUD, dependency mapping, SPoF detection,
// and BIA report generation.
type BIAService struct {
	pool *pgxpool.Pool
}

// NewBIAService creates a new BIAService with the given database pool.
func NewBIAService(pool *pgxpool.Pool) *BIAService {
	return &BIAService{pool: pool}
}

// ============================================================
// DATA TYPES
// ============================================================

// BusinessProcess represents a business process subject to BIA.
type BusinessProcess struct {
	ID                       uuid.UUID       `json:"id"`
	OrganizationID           uuid.UUID       `json:"organization_id"`
	ProcessRef               string          `json:"process_ref"`
	Name                     string          `json:"name"`
	Description              string          `json:"description"`
	ProcessOwnerUserID       *uuid.UUID      `json:"process_owner_user_id"`
	Department               string          `json:"department"`
	Category                 string          `json:"category"`
	Criticality              string          `json:"criticality"`
	Status                   string          `json:"status"`
	FinancialImpactPerHour   *float64        `json:"financial_impact_per_hour_eur"`
	FinancialImpactPerDay    *float64        `json:"financial_impact_per_day_eur"`
	RegulatoryImpact         string          `json:"regulatory_impact"`
	ReputationalImpact       *string         `json:"reputational_impact"`
	LegalImpact              *string         `json:"legal_impact"`
	OperationalImpact        *string         `json:"operational_impact"`
	SafetyImpact             *string         `json:"safety_impact"`
	RTOHours                 *float64        `json:"rto_hours"`
	RPOHours                 *float64        `json:"rpo_hours"`
	MTPDHours                *float64        `json:"mtpd_hours"`
	MinimumServiceLevel      string          `json:"minimum_service_level"`
	DependentAssetIDs        []uuid.UUID     `json:"dependent_asset_ids"`
	DependentVendorIDs       []uuid.UUID     `json:"dependent_vendor_ids"`
	DependentProcessIDs      []uuid.UUID     `json:"dependent_process_ids"`
	KeyPersonnelUserIDs      []uuid.UUID     `json:"key_personnel_user_ids"`
	DataClassification       string          `json:"data_classification"`
	PeakPeriods              []string        `json:"peak_periods"`
	LastBIADate              *time.Time      `json:"last_bia_date"`
	NextBIADue               *time.Time      `json:"next_bia_due"`
	BIAFrequencyMonths       int             `json:"bia_frequency_months"`
	Notes                    string          `json:"notes"`
	Metadata                 json.RawMessage `json:"metadata"`
	CreatedAt                time.Time       `json:"created_at"`
	UpdatedAt                time.Time       `json:"updated_at"`
}

// ProcessDependency represents a single dependency entry for a business process.
type ProcessDependency struct {
	ID                   uuid.UUID  `json:"id"`
	OrganizationID       uuid.UUID  `json:"organization_id"`
	ProcessID            uuid.UUID  `json:"process_id"`
	DependencyType       string     `json:"dependency_type"`
	DependencyEntityType string     `json:"dependency_entity_type"`
	DependencyEntityID   *uuid.UUID `json:"dependency_entity_id"`
	DependencyName       string     `json:"dependency_name"`
	IsCritical           bool       `json:"is_critical"`
	AlternativeAvailable bool       `json:"alternative_available"`
	AlternativeDesc      string     `json:"alternative_description"`
	RecoverySequence     *int       `json:"recovery_sequence"`
	Notes                string     `json:"notes"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

// DependencyGraph is the full dependency graph for visualization.
type DependencyGraph struct {
	Nodes          []GraphNode        `json:"nodes"`
	Edges          []GraphEdge        `json:"edges"`
	CircularPaths  [][]uuid.UUID      `json:"circular_paths,omitempty"`
}

// GraphNode represents a node in the dependency graph.
type GraphNode struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	NodeType    string    `json:"node_type"` // "process" or "dependency"
	Criticality string    `json:"criticality,omitempty"`
	IsCritical  bool      `json:"is_critical,omitempty"`
}

// GraphEdge represents an edge in the dependency graph.
type GraphEdge struct {
	Source     uuid.UUID `json:"source"`
	Target     uuid.UUID `json:"target"`
	Label      string    `json:"label"`
	IsCritical bool      `json:"is_critical"`
}

// SinglePointOfFailure identifies an entity that multiple critical processes depend upon.
type SinglePointOfFailure struct {
	EntityID             *uuid.UUID `json:"entity_id"`
	EntityName           string     `json:"entity_name"`
	EntityType           string     `json:"entity_type"`
	DependencyType       string     `json:"dependency_type"`
	AffectedProcessCount int        `json:"affected_process_count"`
	AffectedProcesses    []SPOFProcess `json:"affected_processes"`
	IsCritical           bool       `json:"is_critical"`
	AlternativeAvailable bool       `json:"alternative_available"`
}

// SPOFProcess holds summary info for a process affected by a single point of failure.
type SPOFProcess struct {
	ProcessID   uuid.UUID `json:"process_id"`
	ProcessRef  string    `json:"process_ref"`
	ProcessName string    `json:"process_name"`
	Criticality string    `json:"criticality"`
	IsDirect    bool      `json:"is_direct"`
}

// BIAReport aggregates the full BIA analysis for an organisation.
type BIAReport struct {
	GeneratedAt             time.Time              `json:"generated_at"`
	TotalProcesses          int                    `json:"total_processes"`
	ProcessesByCriticality  map[string]int         `json:"processes_by_criticality"`
	ProcessesByCategory     map[string]int         `json:"processes_by_category"`
	CriticalProcesses       []ProcessSummary       `json:"critical_processes"`
	RTORPOSummary           RTORPOSummary          `json:"rto_rpo_summary"`
	FinancialImpact         FinancialImpactSummary `json:"financial_impact"`
	SinglePointsOfFailure   []SinglePointOfFailure `json:"single_points_of_failure"`
	ProcessesWithoutBIA     int                    `json:"processes_without_bia"`
	OverdueBIAs             int                    `json:"overdue_bias"`
}

// ProcessSummary provides a minimal view of a process for reports.
type ProcessSummary struct {
	ID                     uuid.UUID `json:"id"`
	ProcessRef             string    `json:"process_ref"`
	Name                   string    `json:"name"`
	Criticality            string    `json:"criticality"`
	Category               string    `json:"category"`
	RTOHours               *float64  `json:"rto_hours"`
	RPOHours               *float64  `json:"rpo_hours"`
	FinancialImpactPerDay  *float64  `json:"financial_impact_per_day_eur"`
}

// RTORPOSummary aggregates RTO/RPO statistics.
type RTORPOSummary struct {
	ProcessesWithRTO int      `json:"processes_with_rto"`
	ProcessesWithRPO int      `json:"processes_with_rpo"`
	MinRTOHours      *float64 `json:"min_rto_hours"`
	MaxRTOHours      *float64 `json:"max_rto_hours"`
	AvgRTOHours      *float64 `json:"avg_rto_hours"`
	MinRPOHours      *float64 `json:"min_rpo_hours"`
	MaxRPOHours      *float64 `json:"max_rpo_hours"`
	AvgRPOHours      *float64 `json:"avg_rpo_hours"`
}

// FinancialImpactSummary aggregates financial impact projections.
type FinancialImpactSummary struct {
	TotalHourlyImpactEUR float64 `json:"total_hourly_impact_eur"`
	TotalDailyImpactEUR  float64 `json:"total_daily_impact_eur"`
	TotalWeeklyImpactEUR float64 `json:"total_weekly_impact_eur"`
}

// ============================================================
// REQUEST TYPES
// ============================================================

// CreateProcessReq is the request body for creating a business process.
type CreateProcessReq struct {
	Name                     string      `json:"name"`
	Description              string      `json:"description"`
	ProcessOwnerUserID       *uuid.UUID  `json:"process_owner_user_id"`
	Department               string      `json:"department"`
	Category                 string      `json:"category"`
	Criticality              string      `json:"criticality"`
	DataClassification       string      `json:"data_classification"`
	PeakPeriods              []string    `json:"peak_periods"`
	BIAFrequencyMonths       int         `json:"bia_frequency_months"`
	Notes                    string      `json:"notes"`
}

// UpdateProcessReq is the request body for updating a business process.
type UpdateProcessReq struct {
	Name                     *string     `json:"name"`
	Description              *string     `json:"description"`
	ProcessOwnerUserID       *uuid.UUID  `json:"process_owner_user_id"`
	Department               *string     `json:"department"`
	Category                 *string     `json:"category"`
	Criticality              *string     `json:"criticality"`
	Status                   *string     `json:"status"`
	DataClassification       *string     `json:"data_classification"`
	PeakPeriods              []string    `json:"peak_periods"`
	BIAFrequencyMonths       *int        `json:"bia_frequency_months"`
	Notes                    *string     `json:"notes"`
	DependentAssetIDs        []uuid.UUID `json:"dependent_asset_ids"`
	DependentVendorIDs       []uuid.UUID `json:"dependent_vendor_ids"`
	DependentProcessIDs      []uuid.UUID `json:"dependent_process_ids"`
	KeyPersonnelUserIDs      []uuid.UUID `json:"key_personnel_user_ids"`
}

// AssessProcessReq is the request body for assessing a process's impact and recovery targets.
type AssessProcessReq struct {
	FinancialImpactPerHour *float64 `json:"financial_impact_per_hour_eur"`
	FinancialImpactPerDay  *float64 `json:"financial_impact_per_day_eur"`
	RegulatoryImpact       string   `json:"regulatory_impact"`
	ReputationalImpact     *string  `json:"reputational_impact"`
	LegalImpact            *string  `json:"legal_impact"`
	OperationalImpact      *string  `json:"operational_impact"`
	SafetyImpact           *string  `json:"safety_impact"`
	RTOHours               *float64 `json:"rto_hours"`
	RPOHours               *float64 `json:"rpo_hours"`
	MTPDHours              *float64 `json:"mtpd_hours"`
	MinimumServiceLevel    string   `json:"minimum_service_level"`
}

// DependencyReq is a single dependency entry in a MapDependencies request.
type DependencyReq struct {
	DependencyType       string     `json:"dependency_type"`
	DependencyEntityType string     `json:"dependency_entity_type"`
	DependencyEntityID   *uuid.UUID `json:"dependency_entity_id"`
	DependencyName       string     `json:"dependency_name"`
	IsCritical           bool       `json:"is_critical"`
	AlternativeAvailable bool       `json:"alternative_available"`
	AlternativeDesc      string     `json:"alternative_description"`
	RecoverySequence     *int       `json:"recovery_sequence"`
	Notes                string     `json:"notes"`
}

// ============================================================
// PROCESS CRUD
// ============================================================

// CreateProcess creates a new business process with auto-generated BP-NNN ref.
func (s *BIAService) CreateProcess(ctx context.Context, orgID uuid.UUID, req CreateProcessReq) (*BusinessProcess, error) {
	if req.Category == "" {
		req.Category = "operational_support"
	}
	if req.Criticality == "" {
		req.Criticality = "important"
	}
	if req.BIAFrequencyMonths <= 0 {
		req.BIAFrequencyMonths = 12
	}

	p := &BusinessProcess{}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO business_processes (
			organization_id, process_ref, name, description,
			process_owner_user_id, department, category, criticality,
			data_classification, peak_periods, bia_frequency_months, notes
		) VALUES (
			$1, generate_bp_ref($1), $2, $3,
			$4, $5, $6::bp_category, $7::bp_criticality,
			$8, $9, $10, $11
		)
		RETURNING id, organization_id, process_ref, name,
			COALESCE(description, ''), process_owner_user_id,
			COALESCE(department, ''), category::TEXT, criticality::TEXT, status::TEXT,
			financial_impact_per_hour_eur, financial_impact_per_day_eur,
			COALESCE(regulatory_impact, ''),
			reputational_impact::TEXT, legal_impact::TEXT,
			operational_impact::TEXT, safety_impact::TEXT,
			rto_hours, rpo_hours, mtpd_hours,
			COALESCE(minimum_service_level, ''),
			COALESCE(dependent_asset_ids, '{}'), COALESCE(dependent_vendor_ids, '{}'),
			COALESCE(dependent_process_ids, '{}'), COALESCE(key_personnel_user_ids, '{}'),
			COALESCE(data_classification, ''), COALESCE(peak_periods, '{}'),
			last_bia_date, next_bia_due, bia_frequency_months,
			COALESCE(notes, ''), COALESCE(metadata, '{}'::jsonb),
			created_at, updated_at`,
		orgID, req.Name, req.Description,
		req.ProcessOwnerUserID, req.Department, req.Category, req.Criticality,
		req.DataClassification, req.PeakPeriods, req.BIAFrequencyMonths, req.Notes,
	).Scan(
		&p.ID, &p.OrganizationID, &p.ProcessRef, &p.Name,
		&p.Description, &p.ProcessOwnerUserID,
		&p.Department, &p.Category, &p.Criticality, &p.Status,
		&p.FinancialImpactPerHour, &p.FinancialImpactPerDay,
		&p.RegulatoryImpact,
		&p.ReputationalImpact, &p.LegalImpact,
		&p.OperationalImpact, &p.SafetyImpact,
		&p.RTOHours, &p.RPOHours, &p.MTPDHours,
		&p.MinimumServiceLevel,
		&p.DependentAssetIDs, &p.DependentVendorIDs,
		&p.DependentProcessIDs, &p.KeyPersonnelUserIDs,
		&p.DataClassification, &p.PeakPeriods,
		&p.LastBIADate, &p.NextBIADue, &p.BIAFrequencyMonths,
		&p.Notes, &p.Metadata,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create business process: %w", err)
	}

	log.Info().
		Str("org_id", orgID.String()).
		Str("process_ref", p.ProcessRef).
		Str("name", p.Name).
		Msg("Business process created")

	return p, nil
}

// GetProcess retrieves a single business process by ID.
func (s *BIAService) GetProcess(ctx context.Context, orgID, processID uuid.UUID) (*BusinessProcess, error) {
	p := &BusinessProcess{}
	err := s.pool.QueryRow(ctx, `
		SELECT id, organization_id, process_ref, name,
			COALESCE(description, ''), process_owner_user_id,
			COALESCE(department, ''), category::TEXT, criticality::TEXT, status::TEXT,
			financial_impact_per_hour_eur, financial_impact_per_day_eur,
			COALESCE(regulatory_impact, ''),
			reputational_impact::TEXT, legal_impact::TEXT,
			operational_impact::TEXT, safety_impact::TEXT,
			rto_hours, rpo_hours, mtpd_hours,
			COALESCE(minimum_service_level, ''),
			COALESCE(dependent_asset_ids, '{}'), COALESCE(dependent_vendor_ids, '{}'),
			COALESCE(dependent_process_ids, '{}'), COALESCE(key_personnel_user_ids, '{}'),
			COALESCE(data_classification, ''), COALESCE(peak_periods, '{}'),
			last_bia_date, next_bia_due, bia_frequency_months,
			COALESCE(notes, ''), COALESCE(metadata, '{}'::jsonb),
			created_at, updated_at
		FROM business_processes
		WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL`,
		processID, orgID,
	).Scan(
		&p.ID, &p.OrganizationID, &p.ProcessRef, &p.Name,
		&p.Description, &p.ProcessOwnerUserID,
		&p.Department, &p.Category, &p.Criticality, &p.Status,
		&p.FinancialImpactPerHour, &p.FinancialImpactPerDay,
		&p.RegulatoryImpact,
		&p.ReputationalImpact, &p.LegalImpact,
		&p.OperationalImpact, &p.SafetyImpact,
		&p.RTOHours, &p.RPOHours, &p.MTPDHours,
		&p.MinimumServiceLevel,
		&p.DependentAssetIDs, &p.DependentVendorIDs,
		&p.DependentProcessIDs, &p.KeyPersonnelUserIDs,
		&p.DataClassification, &p.PeakPeriods,
		&p.LastBIADate, &p.NextBIADue, &p.BIAFrequencyMonths,
		&p.Notes, &p.Metadata,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("business process not found: %w", err)
	}
	return p, nil
}

// ListProcesses returns a paginated list of business processes for an organisation.
func (s *BIAService) ListProcesses(ctx context.Context, orgID uuid.UUID, page, pageSize int) ([]BusinessProcess, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var total int64
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM business_processes
		WHERE organization_id = $1 AND deleted_at IS NULL`, orgID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count business processes: %w", err)
	}

	rows, err := s.pool.Query(ctx, `
		SELECT id, organization_id, process_ref, name,
			COALESCE(description, ''), process_owner_user_id,
			COALESCE(department, ''), category::TEXT, criticality::TEXT, status::TEXT,
			financial_impact_per_hour_eur, financial_impact_per_day_eur,
			COALESCE(regulatory_impact, ''),
			reputational_impact::TEXT, legal_impact::TEXT,
			operational_impact::TEXT, safety_impact::TEXT,
			rto_hours, rpo_hours, mtpd_hours,
			COALESCE(minimum_service_level, ''),
			COALESCE(dependent_asset_ids, '{}'), COALESCE(dependent_vendor_ids, '{}'),
			COALESCE(dependent_process_ids, '{}'), COALESCE(key_personnel_user_ids, '{}'),
			COALESCE(data_classification, ''), COALESCE(peak_periods, '{}'),
			last_bia_date, next_bia_due, bia_frequency_months,
			COALESCE(notes, ''), COALESCE(metadata, '{}'::jsonb),
			created_at, updated_at
		FROM business_processes
		WHERE organization_id = $1 AND deleted_at IS NULL
		ORDER BY
			CASE criticality
				WHEN 'mission_critical' THEN 1
				WHEN 'business_critical' THEN 2
				WHEN 'important' THEN 3
				WHEN 'minor' THEN 4
				WHEN 'non_essential' THEN 5
			END,
			name ASC
		LIMIT $2 OFFSET $3`,
		orgID, pageSize, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list business processes: %w", err)
	}
	defer rows.Close()

	var processes []BusinessProcess
	for rows.Next() {
		var p BusinessProcess
		if err := rows.Scan(
			&p.ID, &p.OrganizationID, &p.ProcessRef, &p.Name,
			&p.Description, &p.ProcessOwnerUserID,
			&p.Department, &p.Category, &p.Criticality, &p.Status,
			&p.FinancialImpactPerHour, &p.FinancialImpactPerDay,
			&p.RegulatoryImpact,
			&p.ReputationalImpact, &p.LegalImpact,
			&p.OperationalImpact, &p.SafetyImpact,
			&p.RTOHours, &p.RPOHours, &p.MTPDHours,
			&p.MinimumServiceLevel,
			&p.DependentAssetIDs, &p.DependentVendorIDs,
			&p.DependentProcessIDs, &p.KeyPersonnelUserIDs,
			&p.DataClassification, &p.PeakPeriods,
			&p.LastBIADate, &p.NextBIADue, &p.BIAFrequencyMonths,
			&p.Notes, &p.Metadata,
			&p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan business process: %w", err)
		}
		processes = append(processes, p)
	}
	return processes, total, nil
}

// UpdateProcess updates mutable fields on a business process.
func (s *BIAService) UpdateProcess(ctx context.Context, orgID, processID uuid.UUID, req UpdateProcessReq) error {
	ct, err := s.pool.Exec(ctx, `
		UPDATE business_processes SET
			name = COALESCE($3, name),
			description = COALESCE($4, description),
			process_owner_user_id = COALESCE($5, process_owner_user_id),
			department = COALESCE($6, department),
			category = COALESCE($7::bp_category, category),
			criticality = COALESCE($8::bp_criticality, criticality),
			status = COALESCE($9::bp_status, status),
			data_classification = COALESCE($10, data_classification),
			peak_periods = COALESCE($11, peak_periods),
			bia_frequency_months = COALESCE($12, bia_frequency_months),
			notes = COALESCE($13, notes),
			dependent_asset_ids = COALESCE($14, dependent_asset_ids),
			dependent_vendor_ids = COALESCE($15, dependent_vendor_ids),
			dependent_process_ids = COALESCE($16, dependent_process_ids),
			key_personnel_user_ids = COALESCE($17, key_personnel_user_ids)
		WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL`,
		processID, orgID,
		req.Name, req.Description, req.ProcessOwnerUserID,
		req.Department, req.Category, req.Criticality, req.Status,
		req.DataClassification, req.PeakPeriods,
		req.BIAFrequencyMonths, req.Notes,
		req.DependentAssetIDs, req.DependentVendorIDs,
		req.DependentProcessIDs, req.KeyPersonnelUserIDs,
	)
	if err != nil {
		return fmt.Errorf("failed to update business process: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("business process not found")
	}
	return nil
}

// AssessProcess updates the impact fields and recovery objectives for a process,
// and records the BIA date and next due date.
func (s *BIAService) AssessProcess(ctx context.Context, orgID, processID uuid.UUID, req AssessProcessReq) error {
	// Fetch the process to get bia_frequency_months
	var freqMonths int
	err := s.pool.QueryRow(ctx, `
		SELECT bia_frequency_months FROM business_processes
		WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL`,
		processID, orgID,
	).Scan(&freqMonths)
	if err != nil {
		return fmt.Errorf("business process not found: %w", err)
	}

	now := time.Now()
	nextDue := now.AddDate(0, freqMonths, 0)

	ct, err := s.pool.Exec(ctx, `
		UPDATE business_processes SET
			financial_impact_per_hour_eur = $3,
			financial_impact_per_day_eur = $4,
			regulatory_impact = $5,
			reputational_impact = $6::impact_level,
			legal_impact = $7::impact_level,
			operational_impact = $8::impact_level,
			safety_impact = $9::impact_level,
			rto_hours = $10,
			rpo_hours = $11,
			mtpd_hours = $12,
			minimum_service_level = $13,
			last_bia_date = $14,
			next_bia_due = $15
		WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL`,
		processID, orgID,
		req.FinancialImpactPerHour, req.FinancialImpactPerDay,
		req.RegulatoryImpact,
		req.ReputationalImpact, req.LegalImpact,
		req.OperationalImpact, req.SafetyImpact,
		req.RTOHours, req.RPOHours, req.MTPDHours,
		req.MinimumServiceLevel,
		now, nextDue,
	)
	if err != nil {
		return fmt.Errorf("failed to assess business process: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("business process not found")
	}

	log.Info().
		Str("org_id", orgID.String()).
		Str("process_id", processID.String()).
		Msg("Business process impact assessment recorded")

	return nil
}

// ============================================================
// DEPENDENCIES
// ============================================================

// MapDependencies replaces all dependency entries for a process.
func (s *BIAService) MapDependencies(ctx context.Context, orgID, processID uuid.UUID, deps []DependencyReq) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Verify the process exists
	var exists bool
	err = tx.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM business_processes
			WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL
		)`, processID, orgID,
	).Scan(&exists)
	if err != nil || !exists {
		return fmt.Errorf("business process not found")
	}

	// Remove existing dependencies
	_, err = tx.Exec(ctx, `
		DELETE FROM process_dependencies_map
		WHERE process_id = $1 AND organization_id = $2`,
		processID, orgID,
	)
	if err != nil {
		return fmt.Errorf("failed to clear existing dependencies: %w", err)
	}

	// Insert new dependencies
	for _, d := range deps {
		_, err = tx.Exec(ctx, `
			INSERT INTO process_dependencies_map (
				organization_id, process_id, dependency_type,
				dependency_entity_type, dependency_entity_id,
				dependency_name, is_critical, alternative_available,
				alternative_description, recovery_sequence, notes
			) VALUES (
				$1, $2, $3::dependency_type, $4, $5, $6, $7, $8, $9, $10, $11
			)`,
			orgID, processID, d.DependencyType,
			d.DependencyEntityType, d.DependencyEntityID,
			d.DependencyName, d.IsCritical, d.AlternativeAvailable,
			d.AlternativeDesc, d.RecoverySequence, d.Notes,
		)
		if err != nil {
			return fmt.Errorf("failed to insert dependency %q: %w", d.DependencyName, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit dependencies: %w", err)
	}

	log.Info().
		Str("org_id", orgID.String()).
		Str("process_id", processID.String()).
		Int("count", len(deps)).
		Msg("Process dependencies mapped")

	return nil
}

// GetDependencies returns all dependencies for a specific process.
func (s *BIAService) GetDependencies(ctx context.Context, orgID, processID uuid.UUID) ([]ProcessDependency, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, organization_id, process_id, dependency_type::TEXT,
			COALESCE(dependency_entity_type, ''), dependency_entity_id,
			dependency_name, is_critical, alternative_available,
			COALESCE(alternative_description, ''), recovery_sequence,
			COALESCE(notes, ''), created_at, updated_at
		FROM process_dependencies_map
		WHERE process_id = $1 AND organization_id = $2
		ORDER BY COALESCE(recovery_sequence, 999999), dependency_name`,
		processID, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get dependencies: %w", err)
	}
	defer rows.Close()

	var deps []ProcessDependency
	for rows.Next() {
		var d ProcessDependency
		if err := rows.Scan(
			&d.ID, &d.OrganizationID, &d.ProcessID, &d.DependencyType,
			&d.DependencyEntityType, &d.DependencyEntityID,
			&d.DependencyName, &d.IsCritical, &d.AlternativeAvailable,
			&d.AlternativeDesc, &d.RecoverySequence,
			&d.Notes, &d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan dependency: %w", err)
		}
		deps = append(deps, d)
	}
	return deps, nil
}

// GetDependencyGraph builds the full dependency graph for visualization,
// including circular dependency detection using DFS.
func (s *BIAService) GetDependencyGraph(ctx context.Context, orgID uuid.UUID) (*DependencyGraph, error) {
	graph := &DependencyGraph{}

	// Load all processes as nodes
	procRows, err := s.pool.Query(ctx, `
		SELECT id, name, criticality::TEXT
		FROM business_processes
		WHERE organization_id = $1 AND deleted_at IS NULL
		ORDER BY name`, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load processes for graph: %w", err)
	}
	defer procRows.Close()

	processIDs := make(map[uuid.UUID]bool)
	for procRows.Next() {
		var node GraphNode
		if err := procRows.Scan(&node.ID, &node.Name, &node.Criticality); err != nil {
			return nil, fmt.Errorf("failed to scan process node: %w", err)
		}
		node.NodeType = "process"
		graph.Nodes = append(graph.Nodes, node)
		processIDs[node.ID] = true
	}

	// Load all dependencies as edges and dependency nodes
	depRows, err := s.pool.Query(ctx, `
		SELECT process_id, dependency_type::TEXT, dependency_entity_type,
			dependency_entity_id, dependency_name, is_critical
		FROM process_dependencies_map
		WHERE organization_id = $1
		ORDER BY process_id, dependency_name`, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load dependencies for graph: %w", err)
	}
	defer depRows.Close()

	depNodeMap := make(map[string]uuid.UUID) // key -> node ID for deduplication
	// adjacency list for cycle detection: process -> list of process dependencies
	adjList := make(map[uuid.UUID][]uuid.UUID)

	for depRows.Next() {
		var procID uuid.UUID
		var depType, depEntityType string
		var depEntityID *uuid.UUID
		var depName string
		var isCritical bool

		if err := depRows.Scan(&procID, &depType, &depEntityType, &depEntityID, &depName, &isCritical); err != nil {
			return nil, fmt.Errorf("failed to scan dependency edge: %w", err)
		}

		// Determine target node
		var targetID uuid.UUID
		if depType == "process" && depEntityID != nil && processIDs[*depEntityID] {
			// Process-to-process dependency: use the actual process node
			targetID = *depEntityID
			adjList[procID] = append(adjList[procID], targetID)
		} else {
			// Non-process dependency: create or reuse a dependency node
			nodeKey := depType + ":" + depName
			if depEntityID != nil {
				nodeKey = depType + ":" + depEntityID.String()
			}
			if existingID, ok := depNodeMap[nodeKey]; ok {
				targetID = existingID
			} else {
				targetID = uuid.New()
				depNodeMap[nodeKey] = targetID
				graph.Nodes = append(graph.Nodes, GraphNode{
					ID:       targetID,
					Name:     depName,
					NodeType: "dependency",
					IsCritical: isCritical,
				})
			}
		}

		graph.Edges = append(graph.Edges, GraphEdge{
			Source:     procID,
			Target:     targetID,
			Label:      depType,
			IsCritical: isCritical,
		})
	}

	// Detect circular dependencies using DFS
	graph.CircularPaths = detectCycles(adjList)

	return graph, nil
}

// detectCycles uses DFS with visited/in-stack tracking to find circular dependencies.
func detectCycles(adjList map[uuid.UUID][]uuid.UUID) [][]uuid.UUID {
	var cycles [][]uuid.UUID
	visited := make(map[uuid.UUID]bool)
	inStack := make(map[uuid.UUID]bool)
	path := make([]uuid.UUID, 0)

	var dfs func(node uuid.UUID)
	dfs = func(node uuid.UUID) {
		visited[node] = true
		inStack[node] = true
		path = append(path, node)

		for _, neighbor := range adjList[node] {
			if !visited[neighbor] {
				dfs(neighbor)
			} else if inStack[neighbor] {
				// Found a cycle: extract the cycle path
				cycleStart := -1
				for i, n := range path {
					if n == neighbor {
						cycleStart = i
						break
					}
				}
				if cycleStart >= 0 {
					cycle := make([]uuid.UUID, len(path)-cycleStart)
					copy(cycle, path[cycleStart:])
					cycle = append(cycle, neighbor) // close the cycle
					cycles = append(cycles, cycle)
				}
			}
		}

		path = path[:len(path)-1]
		inStack[node] = false
	}

	for node := range adjList {
		if !visited[node] {
			dfs(node)
		}
	}

	return cycles
}

// IdentifySinglePointsOfFailure finds entities that multiple critical processes
// depend upon, including transitive dependencies.
//
// Algorithm:
//  1. Load all processes with criticality mission_critical or business_critical.
//  2. For each critical process, collect all direct dependencies from process_dependencies_map.
//  3. For process-type dependencies, transitively follow through to gather the full
//     dependency chain (if Process A depends on Process B which depends on Asset X,
//     then Asset X is a dependency for both A and B).
//  4. Aggregate by dependency entity, count how many distinct critical processes
//     depend on each entity.
//  5. Return entities depended upon by 2+ critical processes.
func (s *BIAService) IdentifySinglePointsOfFailure(ctx context.Context, orgID uuid.UUID) ([]SinglePointOfFailure, error) {
	// Step 1: Load all critical processes
	critRows, err := s.pool.Query(ctx, `
		SELECT id, process_ref, name, criticality::TEXT
		FROM business_processes
		WHERE organization_id = $1
			AND deleted_at IS NULL
			AND criticality IN ('mission_critical', 'business_critical')
		ORDER BY criticality, name`, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load critical processes: %w", err)
	}
	defer critRows.Close()

	type critProcess struct {
		ID          uuid.UUID
		Ref         string
		Name        string
		Criticality string
	}
	var critProcesses []critProcess
	critProcMap := make(map[uuid.UUID]critProcess)
	for critRows.Next() {
		var cp critProcess
		if err := critRows.Scan(&cp.ID, &cp.Ref, &cp.Name, &cp.Criticality); err != nil {
			return nil, fmt.Errorf("failed to scan critical process: %w", err)
		}
		critProcesses = append(critProcesses, cp)
		critProcMap[cp.ID] = cp
	}

	if len(critProcesses) == 0 {
		return []SinglePointOfFailure{}, nil
	}

	// Step 2: Load all dependencies for the org
	depRows, err := s.pool.Query(ctx, `
		SELECT process_id, dependency_type::TEXT, dependency_entity_type,
			dependency_entity_id, dependency_name, is_critical, alternative_available
		FROM process_dependencies_map
		WHERE organization_id = $1`, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load dependencies: %w", err)
	}
	defer depRows.Close()

	type depEntry struct {
		ProcessID            uuid.UUID
		DependencyType       string
		DependencyEntityType string
		DependencyEntityID   *uuid.UUID
		DependencyName       string
		IsCritical           bool
		AlternativeAvailable bool
	}

	// Build adjacency: processID -> list of depEntry
	procDeps := make(map[uuid.UUID][]depEntry)
	for depRows.Next() {
		var d depEntry
		if err := depRows.Scan(
			&d.ProcessID, &d.DependencyType, &d.DependencyEntityType,
			&d.DependencyEntityID, &d.DependencyName, &d.IsCritical,
			&d.AlternativeAvailable,
		); err != nil {
			return nil, fmt.Errorf("failed to scan dependency: %w", err)
		}
		procDeps[d.ProcessID] = append(procDeps[d.ProcessID], d)
	}

	// Step 3: For each critical process, resolve transitive dependencies
	// depKey -> set of critical process IDs that depend on it
	type depKey struct {
		Name       string
		EntityID   string // uuid string or ""
		DepType    string
	}
	type depInfo struct {
		EntityID             *uuid.UUID
		EntityName           string
		EntityType           string
		DepType              string
		IsCritical           bool
		AlternativeAvailable bool
		AffectedProcesses    map[uuid.UUID]bool // processID -> isDirect
		DirectMap            map[uuid.UUID]bool
	}

	spofMap := make(map[depKey]*depInfo)

	for _, cp := range critProcesses {
		// BFS/DFS to collect all dependencies transitively
		visited := make(map[uuid.UUID]bool)
		queue := []struct {
			procID   uuid.UUID
			isDirect bool
		}{{cp.ID, true}}

		for len(queue) > 0 {
			current := queue[0]
			queue = queue[1:]

			if visited[current.procID] {
				continue
			}
			visited[current.procID] = true

			for _, d := range procDeps[current.procID] {
				// Skip process-type deps for spof aggregation (we follow them instead)
				if d.DependencyType == "process" && d.DependencyEntityID != nil {
					// Follow transitive: add the sub-process to the queue
					queue = append(queue, struct {
						procID   uuid.UUID
						isDirect bool
					}{*d.DependencyEntityID, false})
					continue
				}

				// Non-process dependency: aggregate
				entityIDStr := ""
				if d.DependencyEntityID != nil {
					entityIDStr = d.DependencyEntityID.String()
				}
				key := depKey{
					Name:     d.DependencyName,
					EntityID: entityIDStr,
					DepType:  d.DependencyType,
				}

				if _, ok := spofMap[key]; !ok {
					spofMap[key] = &depInfo{
						EntityID:             d.DependencyEntityID,
						EntityName:           d.DependencyName,
						EntityType:           d.DependencyEntityType,
						DepType:              d.DependencyType,
						IsCritical:           d.IsCritical,
						AlternativeAvailable: d.AlternativeAvailable,
						AffectedProcesses:    make(map[uuid.UUID]bool),
						DirectMap:            make(map[uuid.UUID]bool),
					}
				}
				info := spofMap[key]
				info.AffectedProcesses[cp.ID] = true
				if current.isDirect {
					info.DirectMap[cp.ID] = true
				}
				// Accumulate worst-case: if any link is critical, mark as critical
				if d.IsCritical {
					info.IsCritical = true
				}
				// Only mark alternative available if ALL links have alternatives
				if !d.AlternativeAvailable {
					info.AlternativeAvailable = false
				}
			}
		}
	}

	// Step 5: Filter to entities with 2+ critical processes
	var spofs []SinglePointOfFailure
	for _, info := range spofMap {
		if len(info.AffectedProcesses) < 2 {
			continue
		}
		spof := SinglePointOfFailure{
			EntityID:             info.EntityID,
			EntityName:           info.EntityName,
			EntityType:           info.EntityType,
			DependencyType:       info.DepType,
			AffectedProcessCount: len(info.AffectedProcesses),
			IsCritical:           info.IsCritical,
			AlternativeAvailable: info.AlternativeAvailable,
		}
		for pid := range info.AffectedProcesses {
			if cp, ok := critProcMap[pid]; ok {
				_, isDirect := info.DirectMap[pid]
				spof.AffectedProcesses = append(spof.AffectedProcesses, SPOFProcess{
					ProcessID:   cp.ID,
					ProcessRef:  cp.Ref,
					ProcessName: cp.Name,
					Criticality: cp.Criticality,
					IsDirect:    isDirect,
				})
			}
		}
		spofs = append(spofs, spof)
	}

	return spofs, nil
}

// ============================================================
// REPORTS
// ============================================================

// GenerateBIAReport produces a comprehensive BIA report for the organisation,
// including all processes ranked by criticality, SPoFs, RTO/RPO summary,
// and financial impact projections.
func (s *BIAService) GenerateBIAReport(ctx context.Context, orgID uuid.UUID) (*BIAReport, error) {
	report := &BIAReport{
		GeneratedAt:            time.Now().UTC(),
		ProcessesByCriticality: make(map[string]int),
		ProcessesByCategory:    make(map[string]int),
	}

	// Load all active processes
	rows, err := s.pool.Query(ctx, `
		SELECT id, process_ref, name, criticality::TEXT, category::TEXT,
			rto_hours, rpo_hours,
			financial_impact_per_hour_eur, financial_impact_per_day_eur,
			last_bia_date, next_bia_due
		FROM business_processes
		WHERE organization_id = $1 AND deleted_at IS NULL AND status = 'active'
		ORDER BY
			CASE criticality
				WHEN 'mission_critical' THEN 1
				WHEN 'business_critical' THEN 2
				WHEN 'important' THEN 3
				WHEN 'minor' THEN 4
				WHEN 'non_essential' THEN 5
			END,
			name ASC`, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load processes for report: %w", err)
	}
	defer rows.Close()

	now := time.Now()
	var totalHourly, totalDaily float64
	var rtoCount, rpoCount int
	var rtoMin, rtoMax, rtoSum float64
	var rpoMin, rpoMax, rpoSum float64
	rtoFirst := true
	rpoFirst := true

	for rows.Next() {
		var ps ProcessSummary
		var lastBIA, nextBIA *time.Time
		var hourly, daily *float64

		if err := rows.Scan(
			&ps.ID, &ps.ProcessRef, &ps.Name, &ps.Criticality, &ps.Category,
			&ps.RTOHours, &ps.RPOHours,
			&hourly, &daily,
			&lastBIA, &nextBIA,
		); err != nil {
			return nil, fmt.Errorf("failed to scan process for report: %w", err)
		}

		ps.FinancialImpactPerDay = daily

		report.TotalProcesses++
		report.ProcessesByCriticality[ps.Criticality]++
		report.ProcessesByCategory[ps.Category]++
		report.CriticalProcesses = append(report.CriticalProcesses, ps)

		// Financial impact
		if hourly != nil {
			totalHourly += *hourly
		}
		if daily != nil {
			totalDaily += *daily
		}

		// RTO stats
		if ps.RTOHours != nil {
			rtoCount++
			v := *ps.RTOHours
			rtoSum += v
			if rtoFirst || v < rtoMin {
				rtoMin = v
			}
			if rtoFirst || v > rtoMax {
				rtoMax = v
			}
			rtoFirst = false
		}

		// RPO stats
		if ps.RPOHours != nil {
			rpoCount++
			v := *ps.RPOHours
			rpoSum += v
			if rpoFirst || v < rpoMin {
				rpoMin = v
			}
			if rpoFirst || v > rpoMax {
				rpoMax = v
			}
			rpoFirst = false
		}

		// BIA tracking
		if lastBIA == nil {
			report.ProcessesWithoutBIA++
		}
		if nextBIA != nil && nextBIA.Before(now) {
			report.OverdueBIAs++
		}
	}

	// Financial impact summary
	report.FinancialImpact = FinancialImpactSummary{
		TotalHourlyImpactEUR: totalHourly,
		TotalDailyImpactEUR:  totalDaily,
		TotalWeeklyImpactEUR: totalDaily * 5, // 5 business days
	}

	// RTO/RPO summary
	report.RTORPOSummary.ProcessesWithRTO = rtoCount
	report.RTORPOSummary.ProcessesWithRPO = rpoCount
	if rtoCount > 0 {
		avg := rtoSum / float64(rtoCount)
		report.RTORPOSummary.MinRTOHours = &rtoMin
		report.RTORPOSummary.MaxRTOHours = &rtoMax
		report.RTORPOSummary.AvgRTOHours = &avg
	}
	if rpoCount > 0 {
		avg := rpoSum / float64(rpoCount)
		report.RTORPOSummary.MinRPOHours = &rpoMin
		report.RTORPOSummary.MaxRPOHours = &rpoMax
		report.RTORPOSummary.AvgRPOHours = &avg
	}

	// SPoFs
	spofs, err := s.IdentifySinglePointsOfFailure(ctx, orgID)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to identify SPoFs for BIA report")
		spofs = []SinglePointOfFailure{}
	}
	report.SinglePointsOfFailure = spofs

	return report, nil
}
