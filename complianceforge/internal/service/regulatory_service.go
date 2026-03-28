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
// MODEL TYPES
// ============================================================

// RegulatorySource represents a regulatory body, standards organisation,
// or other entity whose publications are monitored for changes.
type RegulatorySource struct {
	ID                  uuid.UUID  `json:"id"`
	Name                string     `json:"name"`
	SourceType          string     `json:"source_type"`
	CountryCode         string     `json:"country_code"`
	Region              string     `json:"region"`
	URL                 string     `json:"url"`
	RSSFeedURL          string     `json:"rss_feed_url"`
	APIURL              string     `json:"api_url"`
	RelevanceFrameworks []string   `json:"relevance_frameworks"`
	ScanFrequency       string     `json:"scan_frequency"`
	LastScannedAt       *time.Time `json:"last_scanned_at"`
	IsActive            bool       `json:"is_active"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// RegulatoryChange represents a single detected change in the regulatory
// landscape — a new law, amendment, guidance update, enforcement action, etc.
type RegulatoryChange struct {
	ID                    uuid.UUID  `json:"id"`
	SourceID              *uuid.UUID `json:"source_id"`
	ChangeRef             string     `json:"change_ref"`
	Title                 string     `json:"title"`
	Summary               string     `json:"summary"`
	FullTextURL           string     `json:"full_text_url"`
	PublishedDate         *time.Time `json:"published_date"`
	EffectiveDate         *time.Time `json:"effective_date"`
	ChangeType            string     `json:"change_type"`
	Severity              string     `json:"severity"`
	Status                string     `json:"status"`
	AffectedFrameworks    []string   `json:"affected_frameworks"`
	AffectedRegions       []string   `json:"affected_regions"`
	AffectedIndustries    []string   `json:"affected_industries"`
	AffectedControlCodes  []string   `json:"affected_control_codes"`
	ImpactAssessment      string     `json:"impact_assessment"`
	ImpactLevel           *string    `json:"impact_level"`
	ComplianceGapCreated  bool       `json:"compliance_gap_created"`
	RequiredActions       string     `json:"required_actions"`
	Deadline              *time.Time `json:"deadline"`
	AssessedBy            *uuid.UUID `json:"assessed_by"`
	AssessedAt            *time.Time `json:"assessed_at"`
	ResponsePlanID        *uuid.UUID `json:"response_plan_id"`
	AssignedTo            *uuid.UUID `json:"assigned_to"`
	Notes                 string     `json:"notes"`
	Tags                  []string   `json:"tags"`
	Metadata              json.RawMessage `json:"metadata"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
	// Joined fields
	SourceName            string     `json:"source_name,omitempty"`
}

// Subscription represents an organisation's subscription to a regulatory source.
type Subscription struct {
	ID                       uuid.UUID  `json:"id"`
	OrganizationID           uuid.UUID  `json:"organization_id"`
	SourceID                 uuid.UUID  `json:"source_id"`
	IsActive                 bool       `json:"is_active"`
	NotificationOnNew        bool       `json:"notification_on_new"`
	NotificationSeverityFilter []string `json:"notification_severity_filter"`
	AutoAssess               bool       `json:"auto_assess"`
	CreatedAt                time.Time  `json:"created_at"`
	// Joined fields
	SourceName               string     `json:"source_name,omitempty"`
	SourceType               string     `json:"source_type,omitempty"`
}

// ImpactAssessment represents a per-organisation assessment of a regulatory change.
type ImpactAssessment struct {
	ID                   uuid.UUID       `json:"id"`
	OrganizationID       uuid.UUID       `json:"organization_id"`
	ChangeID             uuid.UUID       `json:"change_id"`
	Status               string          `json:"status"`
	ImpactOnFrameworks   json.RawMessage `json:"impact_on_frameworks"`
	GapAnalysis          json.RawMessage `json:"gap_analysis"`
	ExistingCoverage     float64         `json:"existing_coverage"`
	EstimatedEffortHours float64         `json:"estimated_effort_hours"`
	EstimatedCostEUR     float64         `json:"estimated_cost_eur"`
	AIAssessment         string          `json:"ai_assessment"`
	HumanAssessment      string          `json:"human_assessment"`
	AssessedBy           *uuid.UUID      `json:"assessed_by"`
	AssessedAt           *time.Time      `json:"assessed_at"`
	RemediationPlanID    *uuid.UUID      `json:"remediation_plan_id"`
	CreatedAt            time.Time       `json:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at"`
}

// RegDashboard provides an at-a-glance summary of an organisation's
// regulatory change posture.
type RegDashboard struct {
	NewChangesCount         int              `json:"new_changes_count"`
	PendingAssessments      int              `json:"pending_assessments"`
	ActionRequired          int              `json:"action_required"`
	UpcomingDeadlines       []DeadlineEntry  `json:"upcoming_deadlines"`
	RecentChanges           []RegulatoryChange `json:"recent_changes"`
	SeverityBreakdown       map[string]int   `json:"severity_breakdown"`
	StatusBreakdown         map[string]int   `json:"status_breakdown"`
	SubscribedSourcesCount  int              `json:"subscribed_sources_count"`
}

// DeadlineEntry represents an upcoming regulatory compliance deadline.
type DeadlineEntry struct {
	ChangeID   uuid.UUID `json:"change_id"`
	ChangeRef  string    `json:"change_ref"`
	Title      string    `json:"title"`
	Deadline   time.Time `json:"deadline"`
	Severity   string    `json:"severity"`
	DaysLeft   int       `json:"days_left"`
}

// TimelineEntry represents an item on the regulatory change timeline.
type TimelineEntry struct {
	Date        string `json:"date"`
	ChangeRef   string `json:"change_ref"`
	Title       string `json:"title"`
	ChangeType  string `json:"change_type"`
	Severity    string `json:"severity"`
	Status      string `json:"status"`
	EventType   string `json:"event_type"` // "published", "effective", "deadline"
}

// ============================================================
// REQUEST TYPES
// ============================================================

// ChangeFilters holds filtering parameters for listing regulatory changes.
type ChangeFilters struct {
	SourceID    *uuid.UUID `json:"source_id"`
	Severity    string     `json:"severity"`
	Framework   string     `json:"framework"`
	Region      string     `json:"region"`
	Status      string     `json:"status"`
	ChangeType  string     `json:"change_type"`
	DateFrom    *time.Time `json:"date_from"`
	DateTo      *time.Time `json:"date_to"`
	Search      string     `json:"search"`
	Page        int        `json:"page"`
	PageSize    int        `json:"page_size"`
}

// CreateSourceReq is the request body for creating a new regulatory source.
type CreateSourceReq struct {
	Name                string   `json:"name"`
	SourceType          string   `json:"source_type"`
	CountryCode         string   `json:"country_code"`
	Region              string   `json:"region"`
	URL                 string   `json:"url"`
	RSSFeedURL          string   `json:"rss_feed_url"`
	APIURL              string   `json:"api_url"`
	RelevanceFrameworks []string `json:"relevance_frameworks"`
	ScanFrequency       string   `json:"scan_frequency"`
}

// SubscribeReq is the request body for subscribing to a regulatory source.
type SubscribeReq struct {
	SourceID                   uuid.UUID `json:"source_id"`
	NotificationOnNew          bool      `json:"notification_on_new"`
	NotificationSeverityFilter []string  `json:"notification_severity_filter"`
	AutoAssess                 bool      `json:"auto_assess"`
}

// AssessReq is the request body for creating/updating an impact assessment.
type AssessReq struct {
	Status               string          `json:"status"`
	ImpactOnFrameworks   json.RawMessage `json:"impact_on_frameworks"`
	GapAnalysis          json.RawMessage `json:"gap_analysis"`
	ExistingCoverage     float64         `json:"existing_coverage"`
	EstimatedEffortHours float64         `json:"estimated_effort_hours"`
	EstimatedCostEUR     float64         `json:"estimated_cost_eur"`
	HumanAssessment      string          `json:"human_assessment"`
}

// ============================================================
// SERVICE
// ============================================================

// RegulatoryService provides business logic for regulatory change management,
// including browsing changes, managing subscriptions, and impact assessment.
type RegulatoryService struct {
	pool    *pgxpool.Pool
	scanner *RegulatoryScanner
}

// NewRegulatoryService creates a new service.
func NewRegulatoryService(pool *pgxpool.Pool, scanner *RegulatoryScanner) *RegulatoryService {
	return &RegulatoryService{pool: pool, scanner: scanner}
}

// ============================================================
// CHANGES
// ============================================================

// ListChanges returns a filtered, paginated list of regulatory changes.
func (s *RegulatoryService) ListChanges(ctx context.Context, filters ChangeFilters) ([]RegulatoryChange, int64, error) {
	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.PageSize < 1 || filters.PageSize > 100 {
		filters.PageSize = 20
	}
	offset := (filters.Page - 1) * filters.PageSize

	var conditions []string
	var args []interface{}
	argIdx := 1

	if filters.SourceID != nil {
		conditions = append(conditions, fmt.Sprintf("rc.source_id = $%d", argIdx))
		args = append(args, *filters.SourceID)
		argIdx++
	}
	if filters.Severity != "" {
		conditions = append(conditions, fmt.Sprintf("rc.severity = $%d", argIdx))
		args = append(args, filters.Severity)
		argIdx++
	}
	if filters.Status != "" {
		conditions = append(conditions, fmt.Sprintf("rc.status = $%d", argIdx))
		args = append(args, filters.Status)
		argIdx++
	}
	if filters.ChangeType != "" {
		conditions = append(conditions, fmt.Sprintf("rc.change_type = $%d", argIdx))
		args = append(args, filters.ChangeType)
		argIdx++
	}
	if filters.Framework != "" {
		conditions = append(conditions, fmt.Sprintf("$%d = ANY(rc.affected_frameworks)", argIdx))
		args = append(args, filters.Framework)
		argIdx++
	}
	if filters.Region != "" {
		conditions = append(conditions, fmt.Sprintf("$%d = ANY(rc.affected_regions)", argIdx))
		args = append(args, filters.Region)
		argIdx++
	}
	if filters.DateFrom != nil {
		conditions = append(conditions, fmt.Sprintf("rc.published_date >= $%d", argIdx))
		args = append(args, *filters.DateFrom)
		argIdx++
	}
	if filters.DateTo != nil {
		conditions = append(conditions, fmt.Sprintf("rc.published_date <= $%d", argIdx))
		args = append(args, *filters.DateTo)
		argIdx++
	}
	if filters.Search != "" {
		conditions = append(conditions, fmt.Sprintf(
			"(rc.title ILIKE '%%' || $%d || '%%' OR rc.summary ILIKE '%%' || $%d || '%%')", argIdx, argIdx))
		args = append(args, filters.Search)
		argIdx++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count query
	countSQL := fmt.Sprintf(`SELECT COUNT(*) FROM regulatory_changes rc %s`, where)
	var total int64
	if err := s.pool.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count changes: %w", err)
	}

	// Data query
	dataSQL := fmt.Sprintf(`
		SELECT rc.id, rc.source_id, rc.change_ref, rc.title, rc.summary,
		       rc.full_text_url, rc.published_date, rc.effective_date,
		       rc.change_type, rc.severity, rc.status,
		       rc.affected_frameworks, rc.affected_regions,
		       rc.affected_industries, rc.affected_control_codes,
		       rc.impact_level, rc.compliance_gap_created,
		       rc.deadline, rc.tags, rc.metadata,
		       rc.created_at, rc.updated_at,
		       COALESCE(rs.name, '') AS source_name
		FROM regulatory_changes rc
		LEFT JOIN regulatory_sources rs ON rc.source_id = rs.id
		%s
		ORDER BY rc.published_date DESC NULLS LAST, rc.created_at DESC
		LIMIT $%d OFFSET $%d`,
		where, argIdx, argIdx+1)

	args = append(args, filters.PageSize, offset)

	rows, err := s.pool.Query(ctx, dataSQL, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query changes: %w", err)
	}
	defer rows.Close()

	var changes []RegulatoryChange
	for rows.Next() {
		var c RegulatoryChange
		if err := rows.Scan(
			&c.ID, &c.SourceID, &c.ChangeRef, &c.Title, &c.Summary,
			&c.FullTextURL, &c.PublishedDate, &c.EffectiveDate,
			&c.ChangeType, &c.Severity, &c.Status,
			&c.AffectedFrameworks, &c.AffectedRegions,
			&c.AffectedIndustries, &c.AffectedControlCodes,
			&c.ImpactLevel, &c.ComplianceGapCreated,
			&c.Deadline, &c.Tags, &c.Metadata,
			&c.CreatedAt, &c.UpdatedAt,
			&c.SourceName,
		); err != nil {
			return nil, 0, fmt.Errorf("scan change: %w", err)
		}
		changes = append(changes, c)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate changes: %w", err)
	}

	if changes == nil {
		changes = []RegulatoryChange{}
	}
	return changes, total, nil
}

// GetChange retrieves a single regulatory change by ID.
func (s *RegulatoryService) GetChange(ctx context.Context, changeID uuid.UUID) (*RegulatoryChange, error) {
	var c RegulatoryChange
	err := s.pool.QueryRow(ctx, `
		SELECT rc.id, rc.source_id, rc.change_ref, rc.title, rc.summary,
		       rc.full_text_url, rc.published_date, rc.effective_date,
		       rc.change_type, rc.severity, rc.status,
		       rc.affected_frameworks, rc.affected_regions,
		       rc.affected_industries, rc.affected_control_codes,
		       rc.impact_assessment, rc.impact_level, rc.compliance_gap_created,
		       rc.required_actions, rc.deadline,
		       rc.assessed_by, rc.assessed_at,
		       rc.response_plan_id, rc.assigned_to,
		       rc.notes, rc.tags, rc.metadata,
		       rc.created_at, rc.updated_at,
		       COALESCE(rs.name, '') AS source_name
		FROM regulatory_changes rc
		LEFT JOIN regulatory_sources rs ON rc.source_id = rs.id
		WHERE rc.id = $1`, changeID).Scan(
		&c.ID, &c.SourceID, &c.ChangeRef, &c.Title, &c.Summary,
		&c.FullTextURL, &c.PublishedDate, &c.EffectiveDate,
		&c.ChangeType, &c.Severity, &c.Status,
		&c.AffectedFrameworks, &c.AffectedRegions,
		&c.AffectedIndustries, &c.AffectedControlCodes,
		&c.ImpactAssessment, &c.ImpactLevel, &c.ComplianceGapCreated,
		&c.RequiredActions, &c.Deadline,
		&c.AssessedBy, &c.AssessedAt,
		&c.ResponsePlanID, &c.AssignedTo,
		&c.Notes, &c.Tags, &c.Metadata,
		&c.CreatedAt, &c.UpdatedAt,
		&c.SourceName,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("regulatory change not found")
		}
		return nil, fmt.Errorf("get change: %w", err)
	}
	return &c, nil
}

// ============================================================
// SOURCES
// ============================================================

// ListSources returns all regulatory sources, both active and inactive.
func (s *RegulatoryService) ListSources(ctx context.Context) ([]RegulatorySource, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, source_type, COALESCE(country_code, ''), COALESCE(region, ''),
		       COALESCE(url, ''), COALESCE(rss_feed_url, ''), COALESCE(api_url, ''),
		       relevance_frameworks, scan_frequency, last_scanned_at,
		       is_active, created_at, updated_at
		FROM regulatory_sources
		ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("query sources: %w", err)
	}
	defer rows.Close()

	var sources []RegulatorySource
	for rows.Next() {
		var src RegulatorySource
		if err := rows.Scan(
			&src.ID, &src.Name, &src.SourceType, &src.CountryCode, &src.Region,
			&src.URL, &src.RSSFeedURL, &src.APIURL,
			&src.RelevanceFrameworks, &src.ScanFrequency, &src.LastScannedAt,
			&src.IsActive, &src.CreatedAt, &src.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan source: %w", err)
		}
		sources = append(sources, src)
	}
	if sources == nil {
		sources = []RegulatorySource{}
	}
	return sources, nil
}

// CreateSource inserts a new regulatory source.
func (s *RegulatoryService) CreateSource(ctx context.Context, req CreateSourceReq) (*RegulatorySource, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("source name is required")
	}
	if req.SourceType == "" {
		req.SourceType = "custom"
	}
	if req.ScanFrequency == "" {
		req.ScanFrequency = "daily"
	}
	if req.RelevanceFrameworks == nil {
		req.RelevanceFrameworks = []string{}
	}

	var src RegulatorySource
	err := s.pool.QueryRow(ctx, `
		INSERT INTO regulatory_sources
			(name, source_type, country_code, region, url, rss_feed_url, api_url,
			 relevance_frameworks, scan_frequency)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, name, source_type, country_code, region, url, rss_feed_url, api_url,
		          relevance_frameworks, scan_frequency, last_scanned_at, is_active, created_at, updated_at`,
		req.Name, req.SourceType, req.CountryCode, req.Region,
		req.URL, req.RSSFeedURL, req.APIURL,
		req.RelevanceFrameworks, req.ScanFrequency,
	).Scan(
		&src.ID, &src.Name, &src.SourceType, &src.CountryCode, &src.Region,
		&src.URL, &src.RSSFeedURL, &src.APIURL,
		&src.RelevanceFrameworks, &src.ScanFrequency, &src.LastScannedAt,
		&src.IsActive, &src.CreatedAt, &src.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert source: %w", err)
	}
	return &src, nil
}

// ============================================================
// SUBSCRIPTIONS
// ============================================================

// GetSubscriptions returns all subscriptions for an organisation.
func (s *RegulatoryService) GetSubscriptions(ctx context.Context, orgID uuid.UUID) ([]Subscription, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT sub.id, sub.organization_id, sub.source_id, sub.is_active,
		       sub.notification_on_new, sub.notification_severity_filter,
		       sub.auto_assess, sub.created_at,
		       COALESCE(rs.name, '') AS source_name,
		       COALESCE(rs.source_type::text, '') AS source_type
		FROM regulatory_subscriptions sub
		LEFT JOIN regulatory_sources rs ON sub.source_id = rs.id
		WHERE sub.organization_id = $1
		ORDER BY rs.name`, orgID)
	if err != nil {
		return nil, fmt.Errorf("query subscriptions: %w", err)
	}
	defer rows.Close()

	var subs []Subscription
	for rows.Next() {
		var sub Subscription
		if err := rows.Scan(
			&sub.ID, &sub.OrganizationID, &sub.SourceID, &sub.IsActive,
			&sub.NotificationOnNew, &sub.NotificationSeverityFilter,
			&sub.AutoAssess, &sub.CreatedAt,
			&sub.SourceName, &sub.SourceType,
		); err != nil {
			return nil, fmt.Errorf("scan subscription: %w", err)
		}
		subs = append(subs, sub)
	}
	if subs == nil {
		subs = []Subscription{}
	}
	return subs, nil
}

// Subscribe creates a new regulatory source subscription for an organisation.
func (s *RegulatoryService) Subscribe(ctx context.Context, orgID uuid.UUID, req SubscribeReq) (*Subscription, error) {
	if req.SourceID == uuid.Nil {
		return nil, fmt.Errorf("source_id is required")
	}
	if req.NotificationSeverityFilter == nil {
		req.NotificationSeverityFilter = []string{}
	}

	var sub Subscription
	err := s.pool.QueryRow(ctx, `
		INSERT INTO regulatory_subscriptions
			(organization_id, source_id, notification_on_new, notification_severity_filter, auto_assess)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (organization_id, source_id) DO UPDATE SET
			is_active = true,
			notification_on_new = EXCLUDED.notification_on_new,
			notification_severity_filter = EXCLUDED.notification_severity_filter,
			auto_assess = EXCLUDED.auto_assess
		RETURNING id, organization_id, source_id, is_active, notification_on_new,
		          notification_severity_filter, auto_assess, created_at`,
		orgID, req.SourceID, req.NotificationOnNew,
		req.NotificationSeverityFilter, req.AutoAssess,
	).Scan(
		&sub.ID, &sub.OrganizationID, &sub.SourceID, &sub.IsActive,
		&sub.NotificationOnNew, &sub.NotificationSeverityFilter,
		&sub.AutoAssess, &sub.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("subscribe: %w", err)
	}
	return &sub, nil
}

// Unsubscribe deactivates a subscription.
func (s *RegulatoryService) Unsubscribe(ctx context.Context, orgID, subID uuid.UUID) error {
	tag, err := s.pool.Exec(ctx, `
		DELETE FROM regulatory_subscriptions
		WHERE id = $1 AND organization_id = $2`, subID, orgID)
	if err != nil {
		return fmt.Errorf("unsubscribe: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("subscription not found")
	}
	return nil
}

// ============================================================
// IMPACT ASSESSMENT
// ============================================================

// AssessImpact creates or updates an impact assessment for a regulatory change.
func (s *RegulatoryService) AssessImpact(ctx context.Context, orgID, changeID uuid.UUID, req AssessReq) (*ImpactAssessment, error) {
	userID := contextUserID(ctx)

	if req.Status == "" {
		req.Status = "in_progress"
	}
	if req.ImpactOnFrameworks == nil {
		req.ImpactOnFrameworks = json.RawMessage(`{}`)
	}
	if req.GapAnalysis == nil {
		req.GapAnalysis = json.RawMessage(`{}`)
	}

	var ia ImpactAssessment
	err := s.pool.QueryRow(ctx, `
		INSERT INTO regulatory_impact_assessments
			(organization_id, change_id, status, impact_on_frameworks, gap_analysis,
			 existing_coverage, estimated_effort_hours, estimated_cost_eur,
			 human_assessment, assessed_by, assessed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
		ON CONFLICT (organization_id, change_id) DO UPDATE SET
			status = EXCLUDED.status,
			impact_on_frameworks = EXCLUDED.impact_on_frameworks,
			gap_analysis = EXCLUDED.gap_analysis,
			existing_coverage = EXCLUDED.existing_coverage,
			estimated_effort_hours = EXCLUDED.estimated_effort_hours,
			estimated_cost_eur = EXCLUDED.estimated_cost_eur,
			human_assessment = EXCLUDED.human_assessment,
			assessed_by = EXCLUDED.assessed_by,
			assessed_at = NOW(),
			updated_at = NOW()
		RETURNING id, organization_id, change_id, status, impact_on_frameworks,
		          gap_analysis, existing_coverage, estimated_effort_hours,
		          estimated_cost_eur, ai_assessment, human_assessment,
		          assessed_by, assessed_at, remediation_plan_id,
		          created_at, updated_at`,
		orgID, changeID, req.Status, req.ImpactOnFrameworks, req.GapAnalysis,
		req.ExistingCoverage, req.EstimatedEffortHours, req.EstimatedCostEUR,
		req.HumanAssessment, userID,
	).Scan(
		&ia.ID, &ia.OrganizationID, &ia.ChangeID, &ia.Status,
		&ia.ImpactOnFrameworks, &ia.GapAnalysis,
		&ia.ExistingCoverage, &ia.EstimatedEffortHours,
		&ia.EstimatedCostEUR, &ia.AIAssessment, &ia.HumanAssessment,
		&ia.AssessedBy, &ia.AssessedAt, &ia.RemediationPlanID,
		&ia.CreatedAt, &ia.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("upsert impact assessment: %w", err)
	}

	// If assessment is completed, update the change status
	if req.Status == "completed" {
		_, _ = s.pool.Exec(ctx, `
			UPDATE regulatory_changes SET status = 'assessed', assessed_at = NOW(), assessed_by = $1
			WHERE id = $2 AND status IN ('new', 'under_assessment')`,
			userID, changeID)
	} else {
		_, _ = s.pool.Exec(ctx, `
			UPDATE regulatory_changes SET status = 'under_assessment'
			WHERE id = $1 AND status = 'new'`, changeID)
	}

	return &ia, nil
}

// GetAssessment retrieves the impact assessment for a change and organisation.
func (s *RegulatoryService) GetAssessment(ctx context.Context, orgID, changeID uuid.UUID) (*ImpactAssessment, error) {
	var ia ImpactAssessment
	err := s.pool.QueryRow(ctx, `
		SELECT id, organization_id, change_id, status, impact_on_frameworks,
		       gap_analysis, existing_coverage, estimated_effort_hours,
		       estimated_cost_eur, ai_assessment, human_assessment,
		       assessed_by, assessed_at, remediation_plan_id,
		       created_at, updated_at
		FROM regulatory_impact_assessments
		WHERE organization_id = $1 AND change_id = $2`, orgID, changeID).Scan(
		&ia.ID, &ia.OrganizationID, &ia.ChangeID, &ia.Status,
		&ia.ImpactOnFrameworks, &ia.GapAnalysis,
		&ia.ExistingCoverage, &ia.EstimatedEffortHours,
		&ia.EstimatedCostEUR, &ia.AIAssessment, &ia.HumanAssessment,
		&ia.AssessedBy, &ia.AssessedAt, &ia.RemediationPlanID,
		&ia.CreatedAt, &ia.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("impact assessment not found")
		}
		return nil, fmt.Errorf("get assessment: %w", err)
	}
	return &ia, nil
}

// ============================================================
// RESPONSE PLAN
// ============================================================

// CreateResponsePlan creates a placeholder response plan for a regulatory change
// and links it to the change record. Returns the plan UUID.
func (s *RegulatoryService) CreateResponsePlan(ctx context.Context, orgID, changeID uuid.UUID) (uuid.UUID, error) {
	planID := uuid.New()

	tag, err := s.pool.Exec(ctx, `
		UPDATE regulatory_changes SET response_plan_id = $1
		WHERE id = $2`, planID, changeID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("set response plan: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return uuid.Nil, fmt.Errorf("regulatory change not found")
	}

	// Also update the impact assessment if one exists
	_, _ = s.pool.Exec(ctx, `
		UPDATE regulatory_impact_assessments SET remediation_plan_id = $1
		WHERE organization_id = $2 AND change_id = $3`,
		planID, orgID, changeID)

	log.Info().
		Str("plan_id", planID.String()).
		Str("change_id", changeID.String()).
		Msg("response plan created (placeholder)")

	return planID, nil
}

// ============================================================
// DASHBOARD
// ============================================================

// GetDashboard returns aggregated regulatory change metrics for an organisation.
func (s *RegulatoryService) GetDashboard(ctx context.Context, orgID uuid.UUID) (*RegDashboard, error) {
	dash := &RegDashboard{
		SeverityBreakdown: make(map[string]int),
		StatusBreakdown:   make(map[string]int),
	}

	// Count of changes in 'new' status
	_ = s.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM regulatory_changes WHERE status = 'new'`).Scan(&dash.NewChangesCount)

	// Count pending assessments for this org
	_ = s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM regulatory_impact_assessments
		WHERE organization_id = $1 AND status = 'pending'`, orgID).Scan(&dash.PendingAssessments)

	// Count changes requiring action
	_ = s.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM regulatory_changes WHERE status = 'action_required'`).Scan(&dash.ActionRequired)

	// Subscribed sources count
	_ = s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM regulatory_subscriptions
		WHERE organization_id = $1 AND is_active = true`, orgID).Scan(&dash.SubscribedSourcesCount)

	// Severity breakdown
	sevRows, err := s.pool.Query(ctx, `
		SELECT severity::text, COUNT(*) FROM regulatory_changes
		WHERE status NOT IN ('implemented', 'not_applicable')
		GROUP BY severity`)
	if err == nil {
		defer sevRows.Close()
		for sevRows.Next() {
			var sev string
			var cnt int
			if sevRows.Scan(&sev, &cnt) == nil {
				dash.SeverityBreakdown[sev] = cnt
			}
		}
	}

	// Status breakdown
	statRows, err := s.pool.Query(ctx, `
		SELECT status::text, COUNT(*) FROM regulatory_changes GROUP BY status`)
	if err == nil {
		defer statRows.Close()
		for statRows.Next() {
			var st string
			var cnt int
			if statRows.Scan(&st, &cnt) == nil {
				dash.StatusBreakdown[st] = cnt
			}
		}
	}

	// Upcoming deadlines (next 90 days)
	deadRows, err := s.pool.Query(ctx, `
		SELECT id, change_ref, title, deadline, severity
		FROM regulatory_changes
		WHERE deadline IS NOT NULL
		  AND deadline >= CURRENT_DATE
		  AND deadline <= CURRENT_DATE + INTERVAL '90 days'
		  AND status NOT IN ('implemented', 'not_applicable')
		ORDER BY deadline ASC
		LIMIT 10`)
	if err == nil {
		defer deadRows.Close()
		for deadRows.Next() {
			var d DeadlineEntry
			if deadRows.Scan(&d.ChangeID, &d.ChangeRef, &d.Title, &d.Deadline, &d.Severity) == nil {
				d.DaysLeft = int(time.Until(d.Deadline).Hours() / 24)
				dash.UpcomingDeadlines = append(dash.UpcomingDeadlines, d)
			}
		}
	}
	if dash.UpcomingDeadlines == nil {
		dash.UpcomingDeadlines = []DeadlineEntry{}
	}

	// Recent changes (last 10)
	recRows, err := s.pool.Query(ctx, `
		SELECT rc.id, rc.change_ref, rc.title, rc.severity, rc.status,
		       rc.change_type, rc.published_date, rc.created_at,
		       COALESCE(rs.name, '') AS source_name
		FROM regulatory_changes rc
		LEFT JOIN regulatory_sources rs ON rc.source_id = rs.id
		ORDER BY rc.created_at DESC
		LIMIT 10`)
	if err == nil {
		defer recRows.Close()
		for recRows.Next() {
			var c RegulatoryChange
			if recRows.Scan(
				&c.ID, &c.ChangeRef, &c.Title, &c.Severity, &c.Status,
				&c.ChangeType, &c.PublishedDate, &c.CreatedAt, &c.SourceName,
			) == nil {
				dash.RecentChanges = append(dash.RecentChanges, c)
			}
		}
	}
	if dash.RecentChanges == nil {
		dash.RecentChanges = []RegulatoryChange{}
	}

	return dash, nil
}

// ============================================================
// TIMELINE
// ============================================================

// GetTimeline returns a chronological list of regulatory events over the
// specified number of months, including published dates, effective dates,
// and compliance deadlines.
func (s *RegulatoryService) GetTimeline(ctx context.Context, orgID uuid.UUID, months int) ([]TimelineEntry, error) {
	if months < 1 {
		months = 6
	}
	if months > 24 {
		months = 24
	}

	rows, err := s.pool.Query(ctx, `
		SELECT change_ref, title, change_type, severity, status,
		       published_date, effective_date, deadline
		FROM regulatory_changes
		WHERE (
			(published_date >= CURRENT_DATE - ($1 || ' months')::interval AND published_date <= CURRENT_DATE + ($1 || ' months')::interval)
			OR (effective_date >= CURRENT_DATE - ($1 || ' months')::interval AND effective_date <= CURRENT_DATE + ($1 || ' months')::interval)
			OR (deadline >= CURRENT_DATE - ($1 || ' months')::interval AND deadline <= CURRENT_DATE + ($1 || ' months')::interval)
		)
		ORDER BY COALESCE(published_date, effective_date, deadline) ASC
		LIMIT 200`, months)
	if err != nil {
		return nil, fmt.Errorf("query timeline: %w", err)
	}
	defer rows.Close()

	var entries []TimelineEntry
	for rows.Next() {
		var (
			changeRef     string
			title         string
			changeType    *string
			severity      string
			status        string
			publishedDate *time.Time
			effectiveDate *time.Time
			deadline      *time.Time
		)

		if err := rows.Scan(&changeRef, &title, &changeType, &severity, &status,
			&publishedDate, &effectiveDate, &deadline); err != nil {
			continue
		}

		ct := ""
		if changeType != nil {
			ct = *changeType
		}

		if publishedDate != nil {
			entries = append(entries, TimelineEntry{
				Date:       publishedDate.Format("2006-01-02"),
				ChangeRef:  changeRef,
				Title:      title,
				ChangeType: ct,
				Severity:   severity,
				Status:     status,
				EventType:  "published",
			})
		}
		if effectiveDate != nil {
			entries = append(entries, TimelineEntry{
				Date:       effectiveDate.Format("2006-01-02"),
				ChangeRef:  changeRef,
				Title:      title,
				ChangeType: ct,
				Severity:   severity,
				Status:     status,
				EventType:  "effective",
			})
		}
		if deadline != nil {
			entries = append(entries, TimelineEntry{
				Date:       deadline.Format("2006-01-02"),
				ChangeRef:  changeRef,
				Title:      title,
				ChangeType: ct,
				Severity:   severity,
				Status:     status,
				EventType:  "deadline",
			})
		}
	}

	if entries == nil {
		entries = []TimelineEntry{}
	}
	return entries, nil
}

// ============================================================
// HELPERS
// ============================================================

// contextUserID extracts the user ID from context if available.
// Returns a nil UUID pointer if not present.
func contextUserID(ctx context.Context) *uuid.UUID {
	if v := ctx.Value("user_id"); v != nil {
		if uid, ok := v.(uuid.UUID); ok {
			return &uid
		}
	}
	return nil
}
