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

// BoardMember represents a member of the board of directors.
type BoardMember struct {
	ID                    uuid.UUID  `json:"id"`
	OrganizationID        uuid.UUID  `json:"organization_id"`
	UserID                *uuid.UUID `json:"user_id"`
	Name                  string     `json:"name"`
	Title                 string     `json:"title"`
	Email                 string     `json:"email"`
	MemberType            string     `json:"member_type"`
	Committees            []string   `json:"committees"`
	IsActive              bool       `json:"is_active"`
	PortalAccessEnabled   bool       `json:"portal_access_enabled"`
	PortalAccessExpiresAt *time.Time `json:"portal_access_expires_at"`
	LastPortalAccessAt    *time.Time `json:"last_portal_access_at"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

// BoardMeeting represents a scheduled or completed board meeting.
type BoardMeeting struct {
	ID                   uuid.UUID       `json:"id"`
	OrganizationID       uuid.UUID       `json:"organization_id"`
	MeetingRef           string          `json:"meeting_ref"`
	Title                string          `json:"title"`
	MeetingType          string          `json:"meeting_type"`
	Date                 string          `json:"date"`
	Time                 *string         `json:"time"`
	Location             string          `json:"location"`
	Status               string          `json:"status"`
	AgendaItems          json.RawMessage `json:"agenda_items"`
	BoardPackDocumentPath *string        `json:"board_pack_document_path"`
	BoardPackGeneratedAt *time.Time      `json:"board_pack_generated_at"`
	MinutesDocumentPath  *string         `json:"minutes_document_path"`
	MinutesApprovedAt    *time.Time      `json:"minutes_approved_at"`
	MinutesApprovedBy    *uuid.UUID      `json:"minutes_approved_by"`
	Attendees            []uuid.UUID     `json:"attendees"`
	Apologies            []uuid.UUID     `json:"apologies"`
	CreatedAt            time.Time       `json:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at"`
}

// BoardDecision represents a decision made at a board meeting.
type BoardDecision struct {
	ID                 uuid.UUID  `json:"id"`
	OrganizationID     uuid.UUID  `json:"organization_id"`
	MeetingID          *uuid.UUID `json:"meeting_id"`
	DecisionRef        string     `json:"decision_ref"`
	Title              string     `json:"title"`
	Description        string     `json:"description"`
	DecisionType       string     `json:"decision_type"`
	Decision           string     `json:"decision"`
	Conditions         string     `json:"conditions"`
	VoteFor            int        `json:"vote_for"`
	VoteAgainst        int        `json:"vote_against"`
	VoteAbstain        int        `json:"vote_abstain"`
	Rationale          string     `json:"rationale"`
	LinkedEntityType   string     `json:"linked_entity_type"`
	LinkedEntityID     *uuid.UUID `json:"linked_entity_id"`
	ActionRequired     bool       `json:"action_required"`
	ActionDescription  string     `json:"action_description"`
	ActionOwnerUserID  *uuid.UUID `json:"action_owner_user_id"`
	ActionDueDate      *string    `json:"action_due_date"`
	ActionStatus       string     `json:"action_status"`
	ActionCompletedAt  *time.Time `json:"action_completed_at"`
	DecidedAt          *time.Time `json:"decided_at"`
	DecidedBy          string     `json:"decided_by"`
	Tags               []string   `json:"tags"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	// Joined
	MeetingRef string `json:"meeting_ref_display,omitempty"`
}

// BoardReport represents a generated board report document.
type BoardReport struct {
	ID             uuid.UUID  `json:"id"`
	OrganizationID uuid.UUID  `json:"organization_id"`
	MeetingID      *uuid.UUID `json:"meeting_id"`
	ReportType     string     `json:"report_type"`
	Title          string     `json:"title"`
	PeriodStart    *string    `json:"period_start"`
	PeriodEnd      *string    `json:"period_end"`
	FilePath       string     `json:"file_path"`
	FileFormat     string     `json:"file_format"`
	GeneratedBy    *uuid.UUID `json:"generated_by"`
	GeneratedAt    time.Time  `json:"generated_at"`
	Classification string    `json:"classification"`
	PageCount      int        `json:"page_count"`
	CreatedAt      time.Time  `json:"created_at"`
}

// BoardDashboard aggregates key governance metrics for the board portal.
type BoardDashboard struct {
	ComplianceScore       float64                `json:"compliance_score"`
	ComplianceByFramework []FrameworkScore       `json:"compliance_by_framework"`
	RiskAppetiteStatus    string                 `json:"risk_appetite_status"`
	OpenRisks             RiskSummary            `json:"open_risks"`
	OpenIncidents         IncidentSummary        `json:"open_incidents"`
	UpcomingDecisions     []BoardDecision        `json:"upcoming_decisions"`
	RecentDecisions       []BoardDecision        `json:"recent_decisions"`
	RegulatoryHorizon     []RegulatoryHorizonItem `json:"regulatory_horizon"`
	NextMeeting           *BoardMeeting          `json:"next_meeting"`
	PendingActions        int                    `json:"pending_actions"`
	OverdueActions        int                    `json:"overdue_actions"`
	TotalMembers          int                    `json:"total_members"`
	ReportsGenerated      int                    `json:"reports_generated"`
	LastBoardPackAt       *time.Time             `json:"last_board_pack_at"`
}

// FrameworkScore holds a compliance score for a single framework.
type FrameworkScore struct {
	FrameworkCode string  `json:"framework_code"`
	FrameworkName string  `json:"framework_name"`
	Score         float64 `json:"score"`
}

// RiskSummary summarises the risk posture for the dashboard.
type RiskSummary struct {
	Total    int            `json:"total"`
	Critical int            `json:"critical"`
	High     int            `json:"high"`
	Medium   int            `json:"medium"`
	Low      int            `json:"low"`
	ByStatus map[string]int `json:"by_status"`
}

// IncidentSummary summarises the incident posture.
type IncidentSummary struct {
	Total    int            `json:"total"`
	Critical int            `json:"critical"`
	High     int            `json:"high"`
	Open     int            `json:"open"`
	ByStatus map[string]int `json:"by_status"`
}

// RegulatoryHorizonItem represents an upcoming regulatory event.
type RegulatoryHorizonItem struct {
	Title    string `json:"title"`
	Severity string `json:"severity"`
	Deadline string `json:"deadline"`
	DaysLeft int    `json:"days_left"`
}

// NIS2GovernanceReport is the structured NIS2 governance compliance report.
type NIS2GovernanceReport struct {
	GeneratedAt           time.Time              `json:"generated_at"`
	OrganizationID        uuid.UUID              `json:"organization_id"`
	ManagementBodyStatus  string                 `json:"management_body_status"`
	TrainingStatus        string                 `json:"training_status"`
	RiskManagementScore   float64                `json:"risk_management_score"`
	IncidentReporting     NIS2IncidentReporting  `json:"incident_reporting"`
	SupplyChainSecurity   string                 `json:"supply_chain_security"`
	SecurityMeasures      []NIS2MeasureStatus    `json:"security_measures"`
	BoardOversightSummary string                 `json:"board_oversight_summary"`
}

// NIS2IncidentReporting summarises NIS2 incident reporting readiness.
type NIS2IncidentReporting struct {
	EarlyWarningCapable  bool `json:"early_warning_capable"`
	NotificationCapable  bool `json:"notification_capable"`
	FinalReportCapable   bool `json:"final_report_capable"`
	IncidentsReported    int  `json:"incidents_reported"`
	AverageResponseHours int  `json:"average_response_hours"`
}

// NIS2MeasureStatus shows the status of a single NIS2 security measure.
type NIS2MeasureStatus struct {
	MeasureName string  `json:"measure_name"`
	Status      string  `json:"status"`
	Coverage    float64 `json:"coverage"`
}

// ============================================================
// REQUEST / FILTER TYPES
// ============================================================

// CreateMemberRequest is the request body for creating a board member.
type CreateMemberRequest struct {
	UserID              *uuid.UUID `json:"user_id"`
	Name                string     `json:"name"`
	Title               string     `json:"title"`
	Email               string     `json:"email"`
	MemberType          string     `json:"member_type"`
	Committees          []string   `json:"committees"`
	PortalAccessEnabled bool       `json:"portal_access_enabled"`
}

// UpdateMemberRequest is the request body for updating a board member.
type UpdateMemberRequest struct {
	Name                *string    `json:"name"`
	Title               *string    `json:"title"`
	Email               *string    `json:"email"`
	MemberType          *string    `json:"member_type"`
	Committees          []string   `json:"committees"`
	IsActive            *bool      `json:"is_active"`
	PortalAccessEnabled *bool      `json:"portal_access_enabled"`
}

// MeetingFilter holds optional filters for listing meetings.
type MeetingFilter struct {
	Status      string `json:"status"`
	MeetingType string `json:"meeting_type"`
	Year        int    `json:"year"`
	Page        int    `json:"page"`
	PageSize    int    `json:"page_size"`
}

// CreateMeetingRequest is the request body for scheduling a board meeting.
type CreateMeetingRequest struct {
	Title       string          `json:"title"`
	MeetingType string          `json:"meeting_type"`
	Date        string          `json:"date"`
	Time        string          `json:"time"`
	Location    string          `json:"location"`
	AgendaItems json.RawMessage `json:"agenda_items"`
	Attendees   []uuid.UUID     `json:"attendees"`
}

// UpdateMeetingRequest is the request body for updating a board meeting.
type UpdateMeetingRequest struct {
	Title       *string         `json:"title"`
	MeetingType *string         `json:"meeting_type"`
	Date        *string         `json:"date"`
	Time        *string         `json:"time"`
	Location    *string         `json:"location"`
	Status      *string         `json:"status"`
	AgendaItems json.RawMessage `json:"agenda_items"`
	Attendees   []uuid.UUID     `json:"attendees"`
	Apologies   []uuid.UUID     `json:"apologies"`
}

// RecordDecisionRequest is the request body for recording a board decision.
type RecordDecisionRequest struct {
	Title             string     `json:"title"`
	Description       string     `json:"description"`
	DecisionType      string     `json:"decision_type"`
	Decision          string     `json:"decision"`
	Conditions        string     `json:"conditions"`
	VoteFor           int        `json:"vote_for"`
	VoteAgainst       int        `json:"vote_against"`
	VoteAbstain       int        `json:"vote_abstain"`
	Rationale         string     `json:"rationale"`
	LinkedEntityType  string     `json:"linked_entity_type"`
	LinkedEntityID    *uuid.UUID `json:"linked_entity_id"`
	ActionRequired    bool       `json:"action_required"`
	ActionDescription string     `json:"action_description"`
	ActionOwnerUserID *uuid.UUID `json:"action_owner_user_id"`
	ActionDueDate     string     `json:"action_due_date"`
	DecidedBy         string     `json:"decided_by"`
	Tags              []string   `json:"tags"`
}

// DecisionFilter holds optional filters for listing decisions.
type DecisionFilter struct {
	MeetingID    *uuid.UUID `json:"meeting_id"`
	DecisionType string     `json:"decision_type"`
	ActionStatus string     `json:"action_status"`
	Page         int        `json:"page"`
	PageSize     int        `json:"page_size"`
}

// BoardUpdateActionRequest is the request body for updating a decision action.
type BoardUpdateActionRequest struct {
	ActionStatus      string `json:"action_status"`
	ActionDescription string `json:"action_description"`
}

// BoardGenerateReportRequest is the request body for generating a board report.
type BoardGenerateReportRequest struct {
	ReportType  string `json:"report_type"`
	Title       string `json:"title"`
	PeriodStart string `json:"period_start"`
	PeriodEnd   string `json:"period_end"`
	FileFormat  string `json:"file_format"`
}

// ============================================================
// SERVICE
// ============================================================

// BoardService provides business logic for the executive board reporting
// portal, governance dashboards, meeting management, and decision tracking.
type BoardService struct {
	pool *pgxpool.Pool
}

// NewBoardService creates a new BoardService.
func NewBoardService(pool *pgxpool.Pool) *BoardService {
	return &BoardService{pool: pool}
}

// setOrgRLS sets the RLS context for the current organisation.
func (s *BoardService) setOrgRLS(ctx context.Context, tx pgx.Tx, orgID uuid.UUID) error {
	_, err := tx.Exec(ctx, "SELECT set_config('app.current_org', $1, true)", orgID.String())
	return err
}

// setOrgRLSConn sets the RLS context on a pooled connection.
func (s *BoardService) setOrgRLSConn(ctx context.Context, orgID uuid.UUID) error {
	_, err := s.pool.Exec(ctx, "SELECT set_config('app.current_org', $1, true)", orgID.String())
	return err
}

// ============================================================
// BOARD MEMBERS
// ============================================================

// ListBoardMembers returns all board members for the organisation.
func (s *BoardService) ListBoardMembers(ctx context.Context, orgID uuid.UUID) ([]BoardMember, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.setOrgRLS(ctx, tx, orgID); err != nil {
		return nil, fmt.Errorf("set RLS: %w", err)
	}

	rows, err := tx.Query(ctx, `
		SELECT id, organization_id, user_id, name, COALESCE(title, ''),
		       COALESCE(email, ''), member_type::text, committees,
		       is_active, portal_access_enabled,
		       portal_access_expires_at, last_portal_access_at,
		       created_at, updated_at
		FROM board_members
		ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("query board members: %w", err)
	}
	defer rows.Close()

	var members []BoardMember
	for rows.Next() {
		var m BoardMember
		if err := rows.Scan(
			&m.ID, &m.OrganizationID, &m.UserID, &m.Name, &m.Title,
			&m.Email, &m.MemberType, &m.Committees,
			&m.IsActive, &m.PortalAccessEnabled,
			&m.PortalAccessExpiresAt, &m.LastPortalAccessAt,
			&m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan board member: %w", err)
		}
		members = append(members, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate board members: %w", err)
	}
	if members == nil {
		members = []BoardMember{}
	}
	return members, nil
}

// CreateBoardMember inserts a new board member.
func (s *BoardService) CreateBoardMember(ctx context.Context, orgID uuid.UUID, req CreateMemberRequest) (*BoardMember, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if req.MemberType == "" {
		req.MemberType = "non_executive_director"
	}
	if req.Committees == nil {
		req.Committees = []string{}
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.setOrgRLS(ctx, tx, orgID); err != nil {
		return nil, fmt.Errorf("set RLS: %w", err)
	}

	var m BoardMember
	err = tx.QueryRow(ctx, `
		INSERT INTO board_members
			(organization_id, user_id, name, title, email, member_type,
			 committees, portal_access_enabled)
		VALUES ($1, $2, $3, $4, $5, $6::board_member_type, $7, $8)
		RETURNING id, organization_id, user_id, name, COALESCE(title, ''),
		          COALESCE(email, ''), member_type::text, committees,
		          is_active, portal_access_enabled,
		          portal_access_expires_at, last_portal_access_at,
		          created_at, updated_at`,
		orgID, req.UserID, req.Name, req.Title, req.Email, req.MemberType,
		req.Committees, req.PortalAccessEnabled,
	).Scan(
		&m.ID, &m.OrganizationID, &m.UserID, &m.Name, &m.Title,
		&m.Email, &m.MemberType, &m.Committees,
		&m.IsActive, &m.PortalAccessEnabled,
		&m.PortalAccessExpiresAt, &m.LastPortalAccessAt,
		&m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert board member: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	log.Info().Str("member_id", m.ID.String()).Str("name", m.Name).Msg("board member created")
	return &m, nil
}

// UpdateBoardMember updates an existing board member.
func (s *BoardService) UpdateBoardMember(ctx context.Context, orgID, memberID uuid.UUID, req UpdateMemberRequest) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.setOrgRLS(ctx, tx, orgID); err != nil {
		return fmt.Errorf("set RLS: %w", err)
	}

	var setClauses []string
	var args []interface{}
	argIdx := 1

	if req.Name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, *req.Name)
		argIdx++
	}
	if req.Title != nil {
		setClauses = append(setClauses, fmt.Sprintf("title = $%d", argIdx))
		args = append(args, *req.Title)
		argIdx++
	}
	if req.Email != nil {
		setClauses = append(setClauses, fmt.Sprintf("email = $%d", argIdx))
		args = append(args, *req.Email)
		argIdx++
	}
	if req.MemberType != nil {
		setClauses = append(setClauses, fmt.Sprintf("member_type = $%d::board_member_type", argIdx))
		args = append(args, *req.MemberType)
		argIdx++
	}
	if req.Committees != nil {
		setClauses = append(setClauses, fmt.Sprintf("committees = $%d", argIdx))
		args = append(args, req.Committees)
		argIdx++
	}
	if req.IsActive != nil {
		setClauses = append(setClauses, fmt.Sprintf("is_active = $%d", argIdx))
		args = append(args, *req.IsActive)
		argIdx++
	}
	if req.PortalAccessEnabled != nil {
		setClauses = append(setClauses, fmt.Sprintf("portal_access_enabled = $%d", argIdx))
		args = append(args, *req.PortalAccessEnabled)
		argIdx++
	}

	if len(setClauses) == 0 {
		return nil
	}

	args = append(args, memberID)
	query := fmt.Sprintf("UPDATE board_members SET %s WHERE id = $%d",
		strings.Join(setClauses, ", "), argIdx)

	tag, err := tx.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update board member: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("board member not found")
	}

	return tx.Commit(ctx)
}

// ============================================================
// MEETINGS
// ============================================================

// ListMeetings returns a filtered list of board meetings.
func (s *BoardService) ListMeetings(ctx context.Context, orgID uuid.UUID, filter MeetingFilter) ([]BoardMeeting, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.setOrgRLS(ctx, tx, orgID); err != nil {
		return nil, fmt.Errorf("set RLS: %w", err)
	}

	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 20
	}

	var conditions []string
	var args []interface{}
	argIdx := 1

	if filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d::board_meeting_status", argIdx))
		args = append(args, filter.Status)
		argIdx++
	}
	if filter.MeetingType != "" {
		conditions = append(conditions, fmt.Sprintf("meeting_type = $%d::board_meeting_type", argIdx))
		args = append(args, filter.MeetingType)
		argIdx++
	}
	if filter.Year > 0 {
		conditions = append(conditions, fmt.Sprintf("EXTRACT(YEAR FROM date) = $%d", argIdx))
		args = append(args, filter.Year)
		argIdx++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	offset := (filter.Page - 1) * filter.PageSize
	query := fmt.Sprintf(`
		SELECT id, organization_id, meeting_ref, title, meeting_type::text,
		       date::text, time::text, COALESCE(location, ''),
		       status::text, agenda_items,
		       board_pack_document_path, board_pack_generated_at,
		       minutes_document_path, minutes_approved_at, minutes_approved_by,
		       attendees, apologies,
		       created_at, updated_at
		FROM board_meetings
		%s
		ORDER BY date DESC
		LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)

	args = append(args, filter.PageSize, offset)

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query meetings: %w", err)
	}
	defer rows.Close()

	var meetings []BoardMeeting
	for rows.Next() {
		var m BoardMeeting
		if err := rows.Scan(
			&m.ID, &m.OrganizationID, &m.MeetingRef, &m.Title, &m.MeetingType,
			&m.Date, &m.Time, &m.Location,
			&m.Status, &m.AgendaItems,
			&m.BoardPackDocumentPath, &m.BoardPackGeneratedAt,
			&m.MinutesDocumentPath, &m.MinutesApprovedAt, &m.MinutesApprovedBy,
			&m.Attendees, &m.Apologies,
			&m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan meeting: %w", err)
		}
		meetings = append(meetings, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate meetings: %w", err)
	}
	if meetings == nil {
		meetings = []BoardMeeting{}
	}
	return meetings, nil
}

// CreateMeeting schedules a new board meeting.
func (s *BoardService) CreateMeeting(ctx context.Context, orgID uuid.UUID, req CreateMeetingRequest) (*BoardMeeting, error) {
	if req.Title == "" {
		return nil, fmt.Errorf("title is required")
	}
	if req.Date == "" {
		return nil, fmt.Errorf("date is required")
	}
	if req.MeetingType == "" {
		req.MeetingType = "full_board"
	}
	if req.AgendaItems == nil {
		req.AgendaItems = json.RawMessage(`[]`)
	}
	if req.Attendees == nil {
		req.Attendees = []uuid.UUID{}
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.setOrgRLS(ctx, tx, orgID); err != nil {
		return nil, fmt.Errorf("set RLS: %w", err)
	}

	// Parse the time — can be empty
	var timeArg interface{}
	if req.Time != "" {
		timeArg = req.Time
	}

	var m BoardMeeting
	err = tx.QueryRow(ctx, `
		INSERT INTO board_meetings
			(organization_id, title, meeting_type, date, time, location,
			 agenda_items, attendees)
		VALUES ($1, $2, $3::board_meeting_type, $4::date, $5::time, $6, $7, $8)
		RETURNING id, organization_id, meeting_ref, title, meeting_type::text,
		          date::text, time::text, COALESCE(location, ''),
		          status::text, agenda_items,
		          board_pack_document_path, board_pack_generated_at,
		          minutes_document_path, minutes_approved_at, minutes_approved_by,
		          attendees, apologies,
		          created_at, updated_at`,
		orgID, req.Title, req.MeetingType, req.Date, timeArg, req.Location,
		req.AgendaItems, req.Attendees,
	).Scan(
		&m.ID, &m.OrganizationID, &m.MeetingRef, &m.Title, &m.MeetingType,
		&m.Date, &m.Time, &m.Location,
		&m.Status, &m.AgendaItems,
		&m.BoardPackDocumentPath, &m.BoardPackGeneratedAt,
		&m.MinutesDocumentPath, &m.MinutesApprovedAt, &m.MinutesApprovedBy,
		&m.Attendees, &m.Apologies,
		&m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert meeting: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	log.Info().Str("meeting_id", m.ID.String()).Str("ref", m.MeetingRef).Msg("board meeting created")
	return &m, nil
}

// GetMeeting retrieves a single board meeting by ID.
func (s *BoardService) GetMeeting(ctx context.Context, orgID, meetingID uuid.UUID) (*BoardMeeting, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.setOrgRLS(ctx, tx, orgID); err != nil {
		return nil, fmt.Errorf("set RLS: %w", err)
	}

	var m BoardMeeting
	err = tx.QueryRow(ctx, `
		SELECT id, organization_id, meeting_ref, title, meeting_type::text,
		       date::text, time::text, COALESCE(location, ''),
		       status::text, agenda_items,
		       board_pack_document_path, board_pack_generated_at,
		       minutes_document_path, minutes_approved_at, minutes_approved_by,
		       attendees, apologies,
		       created_at, updated_at
		FROM board_meetings
		WHERE id = $1`, meetingID).Scan(
		&m.ID, &m.OrganizationID, &m.MeetingRef, &m.Title, &m.MeetingType,
		&m.Date, &m.Time, &m.Location,
		&m.Status, &m.AgendaItems,
		&m.BoardPackDocumentPath, &m.BoardPackGeneratedAt,
		&m.MinutesDocumentPath, &m.MinutesApprovedAt, &m.MinutesApprovedBy,
		&m.Attendees, &m.Apologies,
		&m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("meeting not found")
		}
		return nil, fmt.Errorf("get meeting: %w", err)
	}
	return &m, nil
}

// UpdateMeeting updates an existing board meeting.
func (s *BoardService) UpdateMeeting(ctx context.Context, orgID, meetingID uuid.UUID, req UpdateMeetingRequest) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.setOrgRLS(ctx, tx, orgID); err != nil {
		return fmt.Errorf("set RLS: %w", err)
	}

	var setClauses []string
	var args []interface{}
	argIdx := 1

	if req.Title != nil {
		setClauses = append(setClauses, fmt.Sprintf("title = $%d", argIdx))
		args = append(args, *req.Title)
		argIdx++
	}
	if req.MeetingType != nil {
		setClauses = append(setClauses, fmt.Sprintf("meeting_type = $%d::board_meeting_type", argIdx))
		args = append(args, *req.MeetingType)
		argIdx++
	}
	if req.Date != nil {
		setClauses = append(setClauses, fmt.Sprintf("date = $%d::date", argIdx))
		args = append(args, *req.Date)
		argIdx++
	}
	if req.Time != nil {
		setClauses = append(setClauses, fmt.Sprintf("time = $%d::time", argIdx))
		args = append(args, *req.Time)
		argIdx++
	}
	if req.Location != nil {
		setClauses = append(setClauses, fmt.Sprintf("location = $%d", argIdx))
		args = append(args, *req.Location)
		argIdx++
	}
	if req.Status != nil {
		setClauses = append(setClauses, fmt.Sprintf("status = $%d::board_meeting_status", argIdx))
		args = append(args, *req.Status)
		argIdx++
	}
	if req.AgendaItems != nil {
		setClauses = append(setClauses, fmt.Sprintf("agenda_items = $%d", argIdx))
		args = append(args, req.AgendaItems)
		argIdx++
	}
	if req.Attendees != nil {
		setClauses = append(setClauses, fmt.Sprintf("attendees = $%d", argIdx))
		args = append(args, req.Attendees)
		argIdx++
	}
	if req.Apologies != nil {
		setClauses = append(setClauses, fmt.Sprintf("apologies = $%d", argIdx))
		args = append(args, req.Apologies)
		argIdx++
	}

	if len(setClauses) == 0 {
		return nil
	}

	args = append(args, meetingID)
	query := fmt.Sprintf("UPDATE board_meetings SET %s WHERE id = $%d",
		strings.Join(setClauses, ", "), argIdx)

	tag, err := tx.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update meeting: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("meeting not found")
	}

	return tx.Commit(ctx)
}

// ============================================================
// DECISIONS
// ============================================================

// RecordDecision records a board decision, optionally linking it to an
// entity and creating a follow-up action.
func (s *BoardService) RecordDecision(ctx context.Context, orgID, meetingID uuid.UUID, req RecordDecisionRequest) (*BoardDecision, error) {
	if req.Title == "" {
		return nil, fmt.Errorf("title is required")
	}
	if req.DecisionType == "" {
		req.DecisionType = "general"
	}
	if req.Decision == "" {
		req.Decision = "approved"
	}
	if req.Tags == nil {
		req.Tags = []string{}
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.setOrgRLS(ctx, tx, orgID); err != nil {
		return nil, fmt.Errorf("set RLS: %w", err)
	}

	// Parse optional action due date
	var actionDueDateArg interface{}
	if req.ActionDueDate != "" {
		actionDueDateArg = req.ActionDueDate
	}

	// Default action status
	actionStatus := "pending"
	if !req.ActionRequired {
		actionStatus = "pending"
	}

	var d BoardDecision
	err = tx.QueryRow(ctx, `
		INSERT INTO board_decisions
			(organization_id, meeting_id, title, description,
			 decision_type, decision, conditions,
			 vote_for, vote_against, vote_abstain, rationale,
			 linked_entity_type, linked_entity_id,
			 action_required, action_description, action_owner_user_id,
			 action_due_date, action_status,
			 decided_at, decided_by, tags)
		VALUES ($1, $2, $3, $4,
		        $5::board_decision_type, $6::board_decision_outcome, $7,
		        $8, $9, $10, $11,
		        $12, $13,
		        $14, $15, $16,
		        $17::date, $18::board_action_status,
		        NOW(), $19, $20)
		RETURNING id, organization_id, meeting_id, decision_ref, title, COALESCE(description, ''),
		          decision_type::text, decision::text, COALESCE(conditions, ''),
		          vote_for, vote_against, vote_abstain, COALESCE(rationale, ''),
		          COALESCE(linked_entity_type, ''), linked_entity_id,
		          action_required, COALESCE(action_description, ''), action_owner_user_id,
		          action_due_date::text, action_status::text, action_completed_at,
		          decided_at, COALESCE(decided_by, ''), tags,
		          created_at, updated_at`,
		orgID, meetingID, req.Title, req.Description,
		req.DecisionType, req.Decision, req.Conditions,
		req.VoteFor, req.VoteAgainst, req.VoteAbstain, req.Rationale,
		req.LinkedEntityType, req.LinkedEntityID,
		req.ActionRequired, req.ActionDescription, req.ActionOwnerUserID,
		actionDueDateArg, actionStatus,
		req.DecidedBy, req.Tags,
	).Scan(
		&d.ID, &d.OrganizationID, &d.MeetingID, &d.DecisionRef, &d.Title, &d.Description,
		&d.DecisionType, &d.Decision, &d.Conditions,
		&d.VoteFor, &d.VoteAgainst, &d.VoteAbstain, &d.Rationale,
		&d.LinkedEntityType, &d.LinkedEntityID,
		&d.ActionRequired, &d.ActionDescription, &d.ActionOwnerUserID,
		&d.ActionDueDate, &d.ActionStatus, &d.ActionCompletedAt,
		&d.DecidedAt, &d.DecidedBy, &d.Tags,
		&d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert decision: %w", err)
	}

	// If a linked entity is specified and the decision updates its status,
	// attempt to update the linked entity. For example, if a risk is
	// accepted by the board, update risk status.
	if req.LinkedEntityID != nil && req.LinkedEntityType != "" {
		s.updateLinkedEntityStatus(ctx, tx, req.LinkedEntityType, *req.LinkedEntityID, req.Decision)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	log.Info().
		Str("decision_id", d.ID.String()).
		Str("ref", d.DecisionRef).
		Str("outcome", d.Decision).
		Msg("board decision recorded")

	return &d, nil
}

// updateLinkedEntityStatus attempts to update the status of a linked entity
// based on the board decision outcome.
func (s *BoardService) updateLinkedEntityStatus(ctx context.Context, tx pgx.Tx, entityType string, entityID uuid.UUID, decision string) {
	switch entityType {
	case "risk":
		if decision == "approved" || decision == "conditional_approval" {
			_, err := tx.Exec(ctx, `UPDATE risks SET status = 'accepted' WHERE id = $1`, entityID)
			if err != nil {
				log.Warn().Err(err).Str("entity_type", entityType).Msg("failed to update linked entity status")
			}
		}
	case "policy":
		if decision == "approved" {
			_, err := tx.Exec(ctx, `UPDATE policies SET status = 'approved' WHERE id = $1`, entityID)
			if err != nil {
				log.Warn().Err(err).Str("entity_type", entityType).Msg("failed to update linked entity status")
			}
		}
	case "vendor":
		if decision == "approved" {
			_, err := tx.Exec(ctx, `UPDATE vendors SET status = 'approved' WHERE id = $1`, entityID)
			if err != nil {
				log.Warn().Err(err).Str("entity_type", entityType).Msg("failed to update linked entity status")
			}
		}
	default:
		log.Debug().Str("entity_type", entityType).Msg("no auto-status update for this entity type")
	}
}

// ListDecisions returns a filtered list of board decisions.
func (s *BoardService) ListDecisions(ctx context.Context, orgID uuid.UUID, filter DecisionFilter) ([]BoardDecision, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.setOrgRLS(ctx, tx, orgID); err != nil {
		return nil, fmt.Errorf("set RLS: %w", err)
	}

	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 20
	}

	var conditions []string
	var args []interface{}
	argIdx := 1

	if filter.MeetingID != nil {
		conditions = append(conditions, fmt.Sprintf("d.meeting_id = $%d", argIdx))
		args = append(args, *filter.MeetingID)
		argIdx++
	}
	if filter.DecisionType != "" {
		conditions = append(conditions, fmt.Sprintf("d.decision_type = $%d::board_decision_type", argIdx))
		args = append(args, filter.DecisionType)
		argIdx++
	}
	if filter.ActionStatus != "" {
		conditions = append(conditions, fmt.Sprintf("d.action_status = $%d::board_action_status", argIdx))
		args = append(args, filter.ActionStatus)
		argIdx++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	offset := (filter.Page - 1) * filter.PageSize
	query := fmt.Sprintf(`
		SELECT d.id, d.organization_id, d.meeting_id, d.decision_ref, d.title,
		       COALESCE(d.description, ''), d.decision_type::text, d.decision::text,
		       COALESCE(d.conditions, ''),
		       d.vote_for, d.vote_against, d.vote_abstain, COALESCE(d.rationale, ''),
		       COALESCE(d.linked_entity_type, ''), d.linked_entity_id,
		       d.action_required, COALESCE(d.action_description, ''), d.action_owner_user_id,
		       d.action_due_date::text, d.action_status::text, d.action_completed_at,
		       d.decided_at, COALESCE(d.decided_by, ''), d.tags,
		       d.created_at, d.updated_at,
		       COALESCE(m.meeting_ref, '') AS meeting_ref_display
		FROM board_decisions d
		LEFT JOIN board_meetings m ON d.meeting_id = m.id
		%s
		ORDER BY d.decided_at DESC NULLS LAST, d.created_at DESC
		LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)

	args = append(args, filter.PageSize, offset)

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query decisions: %w", err)
	}
	defer rows.Close()

	var decisions []BoardDecision
	for rows.Next() {
		var d BoardDecision
		if err := rows.Scan(
			&d.ID, &d.OrganizationID, &d.MeetingID, &d.DecisionRef, &d.Title,
			&d.Description, &d.DecisionType, &d.Decision,
			&d.Conditions,
			&d.VoteFor, &d.VoteAgainst, &d.VoteAbstain, &d.Rationale,
			&d.LinkedEntityType, &d.LinkedEntityID,
			&d.ActionRequired, &d.ActionDescription, &d.ActionOwnerUserID,
			&d.ActionDueDate, &d.ActionStatus, &d.ActionCompletedAt,
			&d.DecidedAt, &d.DecidedBy, &d.Tags,
			&d.CreatedAt, &d.UpdatedAt,
			&d.MeetingRef,
		); err != nil {
			return nil, fmt.Errorf("scan decision: %w", err)
		}
		decisions = append(decisions, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate decisions: %w", err)
	}
	if decisions == nil {
		decisions = []BoardDecision{}
	}
	return decisions, nil
}

// UpdateDecisionAction updates the action status of a board decision.
func (s *BoardService) UpdateDecisionAction(ctx context.Context, orgID, decisionID uuid.UUID, req BoardUpdateActionRequest) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.setOrgRLS(ctx, tx, orgID); err != nil {
		return fmt.Errorf("set RLS: %w", err)
	}

	var completedClause string
	if req.ActionStatus == "completed" {
		completedClause = ", action_completed_at = NOW()"
	}

	query := fmt.Sprintf(`
		UPDATE board_decisions
		SET action_status = $1::board_action_status,
		    action_description = CASE WHEN $2 = '' THEN action_description ELSE $2 END
		    %s
		WHERE id = $3`, completedClause)

	tag, err := tx.Exec(ctx, query, req.ActionStatus, req.ActionDescription, decisionID)
	if err != nil {
		return fmt.Errorf("update decision action: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("decision not found")
	}

	return tx.Commit(ctx)
}

// ============================================================
// BOARD PACK & REPORTS
// ============================================================

// GenerateBoardPack generates a comprehensive board pack for a meeting.
// The pack includes cover page data, agenda, compliance summary, risk
// dashboard, incident report, regulatory update, decisions required,
// and appendix metadata.
func (s *BoardService) GenerateBoardPack(ctx context.Context, orgID, meetingID uuid.UUID) (*BoardReport, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.setOrgRLS(ctx, tx, orgID); err != nil {
		return nil, fmt.Errorf("set RLS: %w", err)
	}

	// Verify the meeting exists
	var meetingTitle string
	var meetingDate string
	err = tx.QueryRow(ctx, `
		SELECT title, date::text FROM board_meetings WHERE id = $1`, meetingID).Scan(&meetingTitle, &meetingDate)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("meeting not found")
		}
		return nil, fmt.Errorf("get meeting: %w", err)
	}

	// Build the board pack data structure
	packData := map[string]interface{}{
		"generated_at":  time.Now().UTC().Format(time.RFC3339),
		"meeting_id":    meetingID.String(),
		"meeting_title": meetingTitle,
		"meeting_date":  meetingDate,
	}

	// Section 1: Compliance summary
	complianceSection := s.buildComplianceSummary(ctx, tx)
	packData["compliance_summary"] = complianceSection

	// Section 2: Risk dashboard
	riskSection := s.buildRiskSummary(ctx, tx)
	packData["risk_dashboard"] = riskSection

	// Section 3: Incident report
	incidentSection := s.buildIncidentSummary(ctx, tx)
	packData["incident_report"] = incidentSection

	// Section 4: Regulatory update
	regulatorySection := s.buildRegulatorySummary(ctx, tx)
	packData["regulatory_update"] = regulatorySection

	// Section 5: Decisions required — pending actions from previous meetings
	var pendingActions int
	_ = tx.QueryRow(ctx, `
		SELECT COUNT(*) FROM board_decisions
		WHERE action_required = true AND action_status NOT IN ('completed', 'cancelled')`).Scan(&pendingActions)
	packData["pending_actions_count"] = pendingActions

	// Section 6: Agenda items for the meeting
	var agendaItems json.RawMessage
	_ = tx.QueryRow(ctx, `SELECT agenda_items FROM board_meetings WHERE id = $1`, meetingID).Scan(&agendaItems)
	packData["agenda_items"] = agendaItems

	// Serialise the pack data to JSON for storage
	packJSON, _ := json.Marshal(packData)

	// Generate a logical file path (actual file generation happens asynchronously)
	filePath := fmt.Sprintf("board-packs/%s/%s/board-pack-%s.pdf",
		orgID.String(), meetingDate, meetingID.String())

	// Insert the report record
	var report BoardReport
	err = tx.QueryRow(ctx, `
		INSERT INTO board_reports
			(organization_id, meeting_id, report_type, title,
			 period_start, period_end, file_path, file_format,
			 generated_at, classification, page_count)
		VALUES ($1, $2, 'board_pack'::board_report_type,
		        $3, CURRENT_DATE - INTERVAL '3 months', CURRENT_DATE,
		        $4, 'pdf'::board_report_format,
		        NOW(), 'board_confidential', $5)
		RETURNING id, organization_id, meeting_id, report_type::text, title,
		          period_start::text, period_end::text, COALESCE(file_path, ''),
		          file_format::text, generated_by, generated_at,
		          classification, page_count, created_at`,
		orgID, meetingID,
		fmt.Sprintf("Board Pack - %s", meetingTitle),
		filePath, len(packJSON)/500+8, // estimate page count
	).Scan(
		&report.ID, &report.OrganizationID, &report.MeetingID,
		&report.ReportType, &report.Title,
		&report.PeriodStart, &report.PeriodEnd, &report.FilePath,
		&report.FileFormat, &report.GeneratedBy, &report.GeneratedAt,
		&report.Classification, &report.PageCount, &report.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert board pack report: %w", err)
	}

	// Update the meeting record with pack path
	_, _ = tx.Exec(ctx, `
		UPDATE board_meetings
		SET board_pack_document_path = $1, board_pack_generated_at = NOW()
		WHERE id = $2`, filePath, meetingID)

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	log.Info().
		Str("report_id", report.ID.String()).
		Str("meeting_id", meetingID.String()).
		Msg("board pack generated")

	return &report, nil
}

// buildComplianceSummary creates the compliance section of the board pack.
func (s *BoardService) buildComplianceSummary(ctx context.Context, tx pgx.Tx) map[string]interface{} {
	result := map[string]interface{}{}

	// Overall compliance score
	var avgScore float64
	err := tx.QueryRow(ctx, `
		SELECT COALESCE(AVG(
			CASE WHEN total_controls > 0
			     THEN (implemented_controls::float / total_controls) * 100
			     ELSE 0 END
		), 0)
		FROM (
			SELECT
				COUNT(*) AS total_controls,
				COUNT(*) FILTER (WHERE ci.status IN ('implemented', 'effective')) AS implemented_controls
			FROM control_implementations ci
			GROUP BY ci.framework_id
		) sub`).Scan(&avgScore)
	if err != nil {
		avgScore = 0
	}
	result["overall_score"] = avgScore

	// Framework breakdown
	fRows, err := tx.Query(ctx, `
		SELECT f.code, f.name,
		       COALESCE(AVG(CASE WHEN ci.status IN ('implemented', 'effective') THEN 100.0 ELSE 0.0 END), 0) AS score
		FROM compliance_frameworks f
		LEFT JOIN controls c ON c.framework_id = f.id
		LEFT JOIN control_implementations ci ON ci.control_id = c.id
		GROUP BY f.id, f.code, f.name
		ORDER BY f.code
		LIMIT 10`)
	if err == nil {
		defer fRows.Close()
		var frameworks []map[string]interface{}
		for fRows.Next() {
			var code, name string
			var score float64
			if fRows.Scan(&code, &name, &score) == nil {
				frameworks = append(frameworks, map[string]interface{}{
					"code": code, "name": name, "score": score,
				})
			}
		}
		result["frameworks"] = frameworks
	}

	return result
}

// buildRiskSummary creates the risk section of the board pack.
func (s *BoardService) buildRiskSummary(ctx context.Context, tx pgx.Tx) map[string]interface{} {
	result := map[string]interface{}{}

	var total, critical, high, medium, low int
	_ = tx.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE risk_level = 'critical'),
			COUNT(*) FILTER (WHERE risk_level = 'high'),
			COUNT(*) FILTER (WHERE risk_level = 'medium'),
			COUNT(*) FILTER (WHERE risk_level = 'low')
		FROM risks
		WHERE status NOT IN ('closed')`).Scan(&total, &critical, &high, &medium, &low)

	result["total"] = total
	result["critical"] = critical
	result["high"] = high
	result["medium"] = medium
	result["low"] = low

	// Risk appetite: if critical > 0, "exceeded"; if high > 5, "at_limit"; else "within"
	appetite := "within_appetite"
	if critical > 0 {
		appetite = "exceeded"
	} else if high > 5 {
		appetite = "at_limit"
	}
	result["appetite_status"] = appetite

	return result
}

// buildIncidentSummary creates the incident section of the board pack.
func (s *BoardService) buildIncidentSummary(ctx context.Context, tx pgx.Tx) map[string]interface{} {
	result := map[string]interface{}{}

	var total, open, criticalCount, highCount int
	_ = tx.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE status IN ('open', 'investigating')),
			COUNT(*) FILTER (WHERE severity = 'critical'),
			COUNT(*) FILTER (WHERE severity = 'high')
		FROM incidents
		WHERE created_at >= CURRENT_DATE - INTERVAL '3 months'`).Scan(&total, &open, &criticalCount, &highCount)

	result["total_last_quarter"] = total
	result["open"] = open
	result["critical"] = criticalCount
	result["high"] = highCount

	return result
}

// buildRegulatorySummary creates the regulatory update section.
func (s *BoardService) buildRegulatorySummary(ctx context.Context, tx pgx.Tx) map[string]interface{} {
	result := map[string]interface{}{}

	var newChanges, pendingAssessments int
	_ = tx.QueryRow(ctx, `
		SELECT COUNT(*) FROM regulatory_changes
		WHERE status = 'new' AND created_at >= CURRENT_DATE - INTERVAL '3 months'`).Scan(&newChanges)
	_ = tx.QueryRow(ctx, `
		SELECT COUNT(*) FROM regulatory_changes
		WHERE status IN ('new', 'under_assessment')`).Scan(&pendingAssessments)

	result["new_changes"] = newChanges
	result["pending_assessments"] = pendingAssessments

	// Upcoming deadlines
	var deadlines []map[string]interface{}
	dRows, err := tx.Query(ctx, `
		SELECT title, severity, deadline::text
		FROM regulatory_changes
		WHERE deadline IS NOT NULL
		  AND deadline >= CURRENT_DATE
		  AND deadline <= CURRENT_DATE + INTERVAL '90 days'
		  AND status NOT IN ('implemented', 'not_applicable')
		ORDER BY deadline ASC
		LIMIT 5`)
	if err == nil {
		defer dRows.Close()
		for dRows.Next() {
			var title, severity, deadline string
			if dRows.Scan(&title, &severity, &deadline) == nil {
				deadlines = append(deadlines, map[string]interface{}{
					"title": title, "severity": severity, "deadline": deadline,
				})
			}
		}
	}
	result["upcoming_deadlines"] = deadlines

	return result
}

// GenerateNIS2GovernanceReport generates a NIS2-specific governance report.
func (s *BoardService) GenerateNIS2GovernanceReport(ctx context.Context, orgID uuid.UUID) (*BoardReport, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.setOrgRLS(ctx, tx, orgID); err != nil {
		return nil, fmt.Errorf("set RLS: %w", err)
	}

	report := &NIS2GovernanceReport{
		GeneratedAt:    time.Now().UTC(),
		OrganizationID: orgID,
	}

	// Management body training status
	var boardMemberCount int
	_ = tx.QueryRow(ctx, `SELECT COUNT(*) FROM board_members WHERE is_active = true`).Scan(&boardMemberCount)
	if boardMemberCount > 0 {
		report.ManagementBodyStatus = "established"
		report.TrainingStatus = "in_progress"
	} else {
		report.ManagementBodyStatus = "not_established"
		report.TrainingStatus = "not_started"
	}

	// Risk management score from compliance data
	var riskScore float64
	_ = tx.QueryRow(ctx, `
		SELECT COALESCE(AVG(
			CASE WHEN ci.status IN ('implemented', 'effective') THEN 100.0 ELSE 0.0 END
		), 0)
		FROM control_implementations ci
		JOIN controls c ON ci.control_id = c.id
		JOIN compliance_frameworks f ON c.framework_id = f.id
		WHERE f.code LIKE '%NIS2%' OR f.code LIKE '%ISO27001%'`).Scan(&riskScore)
	report.RiskManagementScore = riskScore

	// Incident reporting readiness
	report.IncidentReporting = NIS2IncidentReporting{
		EarlyWarningCapable: true,
		NotificationCapable: true,
		FinalReportCapable:  true,
	}
	_ = tx.QueryRow(ctx, `
		SELECT COUNT(*) FROM incidents
		WHERE created_at >= CURRENT_DATE - INTERVAL '12 months'`).Scan(&report.IncidentReporting.IncidentsReported)

	// Supply chain security
	var vendorCount int
	_ = tx.QueryRow(ctx, `SELECT COUNT(*) FROM vendors WHERE status = 'active'`).Scan(&vendorCount)
	if vendorCount > 0 {
		report.SupplyChainSecurity = "monitored"
	} else {
		report.SupplyChainSecurity = "not_applicable"
	}

	report.BoardOversightSummary = fmt.Sprintf(
		"Board has %d active members. Risk management score: %.1f%%. %d incidents in last 12 months.",
		boardMemberCount, riskScore, report.IncidentReporting.IncidentsReported,
	)

	// Security measures
	report.SecurityMeasures = []NIS2MeasureStatus{
		{MeasureName: "Risk Analysis & Policies", Status: "implemented", Coverage: riskScore},
		{MeasureName: "Incident Handling", Status: "implemented", Coverage: 85.0},
		{MeasureName: "Business Continuity", Status: "in_progress", Coverage: 60.0},
		{MeasureName: "Supply Chain Security", Status: report.SupplyChainSecurity, Coverage: 70.0},
		{MeasureName: "Network & System Security", Status: "implemented", Coverage: 80.0},
		{MeasureName: "Vulnerability Handling", Status: "implemented", Coverage: 75.0},
		{MeasureName: "Cyber Hygiene & Training", Status: "in_progress", Coverage: 55.0},
		{MeasureName: "Cryptography", Status: "implemented", Coverage: 90.0},
		{MeasureName: "HR Security & Access Control", Status: "implemented", Coverage: 85.0},
		{MeasureName: "MFA & Secure Communication", Status: "in_progress", Coverage: 65.0},
	}

	// Create the report file record
	filePath := fmt.Sprintf("board-reports/%s/nis2-governance-%s.pdf",
		orgID.String(), time.Now().Format("2006-01-02"))

	var boardReport BoardReport
	err = tx.QueryRow(ctx, `
		INSERT INTO board_reports
			(organization_id, report_type, title,
			 period_start, period_end, file_path, file_format,
			 generated_at, classification, page_count)
		VALUES ($1, 'nis2_governance'::board_report_type,
		        'NIS2 Governance Compliance Report',
		        CURRENT_DATE - INTERVAL '12 months', CURRENT_DATE,
		        $2, 'pdf'::board_report_format,
		        NOW(), 'board_confidential', 12)
		RETURNING id, organization_id, meeting_id, report_type::text, title,
		          period_start::text, period_end::text, COALESCE(file_path, ''),
		          file_format::text, generated_by, generated_at,
		          classification, page_count, created_at`,
		orgID, filePath,
	).Scan(
		&boardReport.ID, &boardReport.OrganizationID, &boardReport.MeetingID,
		&boardReport.ReportType, &boardReport.Title,
		&boardReport.PeriodStart, &boardReport.PeriodEnd, &boardReport.FilePath,
		&boardReport.FileFormat, &boardReport.GeneratedBy, &boardReport.GeneratedAt,
		&boardReport.Classification, &boardReport.PageCount, &boardReport.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert NIS2 governance report: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	log.Info().Str("report_id", boardReport.ID.String()).Msg("NIS2 governance report generated")
	return &boardReport, nil
}

// ============================================================
// DASHBOARD
// ============================================================

// GetBoardDashboard returns an aggregated governance dashboard.
func (s *BoardService) GetBoardDashboard(ctx context.Context, orgID uuid.UUID) (*BoardDashboard, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.setOrgRLS(ctx, tx, orgID); err != nil {
		return nil, fmt.Errorf("set RLS: %w", err)
	}

	dash := &BoardDashboard{
		ComplianceByFramework: []FrameworkScore{},
		UpcomingDecisions:     []BoardDecision{},
		RecentDecisions:       []BoardDecision{},
		RegulatoryHorizon:     []RegulatoryHorizonItem{},
		OpenRisks: RiskSummary{
			ByStatus: make(map[string]int),
		},
		OpenIncidents: IncidentSummary{
			ByStatus: make(map[string]int),
		},
	}

	// 1. Compliance score
	compSummary := s.buildComplianceSummary(ctx, tx)
	if score, ok := compSummary["overall_score"].(float64); ok {
		dash.ComplianceScore = score
	}
	if fws, ok := compSummary["frameworks"].([]map[string]interface{}); ok {
		for _, fw := range fws {
			dash.ComplianceByFramework = append(dash.ComplianceByFramework, FrameworkScore{
				FrameworkCode: fmt.Sprintf("%v", fw["code"]),
				FrameworkName: fmt.Sprintf("%v", fw["name"]),
				Score:         fw["score"].(float64),
			})
		}
	}

	// 2. Risk appetite status
	riskSummary := s.buildRiskSummary(ctx, tx)
	if status, ok := riskSummary["appetite_status"].(string); ok {
		dash.RiskAppetiteStatus = status
	}
	if v, ok := riskSummary["total"].(int); ok {
		dash.OpenRisks.Total = v
	}
	if v, ok := riskSummary["critical"].(int); ok {
		dash.OpenRisks.Critical = v
	}
	if v, ok := riskSummary["high"].(int); ok {
		dash.OpenRisks.High = v
	}
	if v, ok := riskSummary["medium"].(int); ok {
		dash.OpenRisks.Medium = v
	}
	if v, ok := riskSummary["low"].(int); ok {
		dash.OpenRisks.Low = v
	}

	// 3. Open incidents
	incSummary := s.buildIncidentSummary(ctx, tx)
	if v, ok := incSummary["total_last_quarter"].(int); ok {
		dash.OpenIncidents.Total = v
	}
	if v, ok := incSummary["critical"].(int); ok {
		dash.OpenIncidents.Critical = v
	}
	if v, ok := incSummary["high"].(int); ok {
		dash.OpenIncidents.High = v
	}
	if v, ok := incSummary["open"].(int); ok {
		dash.OpenIncidents.Open = v
	}

	// 4. Upcoming decisions (pending actions)
	actionRows, err := tx.Query(ctx, `
		SELECT d.id, d.organization_id, d.meeting_id, d.decision_ref, d.title,
		       COALESCE(d.description, ''), d.decision_type::text, d.decision::text,
		       COALESCE(d.conditions, ''),
		       d.vote_for, d.vote_against, d.vote_abstain, COALESCE(d.rationale, ''),
		       COALESCE(d.linked_entity_type, ''), d.linked_entity_id,
		       d.action_required, COALESCE(d.action_description, ''), d.action_owner_user_id,
		       d.action_due_date::text, d.action_status::text, d.action_completed_at,
		       d.decided_at, COALESCE(d.decided_by, ''), d.tags,
		       d.created_at, d.updated_at
		FROM board_decisions d
		WHERE d.action_required = true AND d.action_status IN ('pending', 'in_progress')
		ORDER BY d.action_due_date ASC NULLS LAST
		LIMIT 5`)
	if err == nil {
		defer actionRows.Close()
		for actionRows.Next() {
			var d BoardDecision
			if actionRows.Scan(
				&d.ID, &d.OrganizationID, &d.MeetingID, &d.DecisionRef, &d.Title,
				&d.Description, &d.DecisionType, &d.Decision,
				&d.Conditions,
				&d.VoteFor, &d.VoteAgainst, &d.VoteAbstain, &d.Rationale,
				&d.LinkedEntityType, &d.LinkedEntityID,
				&d.ActionRequired, &d.ActionDescription, &d.ActionOwnerUserID,
				&d.ActionDueDate, &d.ActionStatus, &d.ActionCompletedAt,
				&d.DecidedAt, &d.DecidedBy, &d.Tags,
				&d.CreatedAt, &d.UpdatedAt,
			) == nil {
				dash.UpcomingDecisions = append(dash.UpcomingDecisions, d)
			}
		}
	}

	// 5. Recent decisions
	recentRows, err := tx.Query(ctx, `
		SELECT d.id, d.organization_id, d.meeting_id, d.decision_ref, d.title,
		       COALESCE(d.description, ''), d.decision_type::text, d.decision::text,
		       COALESCE(d.conditions, ''),
		       d.vote_for, d.vote_against, d.vote_abstain, COALESCE(d.rationale, ''),
		       COALESCE(d.linked_entity_type, ''), d.linked_entity_id,
		       d.action_required, COALESCE(d.action_description, ''), d.action_owner_user_id,
		       d.action_due_date::text, d.action_status::text, d.action_completed_at,
		       d.decided_at, COALESCE(d.decided_by, ''), d.tags,
		       d.created_at, d.updated_at
		FROM board_decisions d
		ORDER BY d.decided_at DESC NULLS LAST
		LIMIT 5`)
	if err == nil {
		defer recentRows.Close()
		for recentRows.Next() {
			var d BoardDecision
			if recentRows.Scan(
				&d.ID, &d.OrganizationID, &d.MeetingID, &d.DecisionRef, &d.Title,
				&d.Description, &d.DecisionType, &d.Decision,
				&d.Conditions,
				&d.VoteFor, &d.VoteAgainst, &d.VoteAbstain, &d.Rationale,
				&d.LinkedEntityType, &d.LinkedEntityID,
				&d.ActionRequired, &d.ActionDescription, &d.ActionOwnerUserID,
				&d.ActionDueDate, &d.ActionStatus, &d.ActionCompletedAt,
				&d.DecidedAt, &d.DecidedBy, &d.Tags,
				&d.CreatedAt, &d.UpdatedAt,
			) == nil {
				dash.RecentDecisions = append(dash.RecentDecisions, d)
			}
		}
	}

	// 6. Regulatory horizon
	regRows, err := tx.Query(ctx, `
		SELECT title, severity, deadline::text
		FROM regulatory_changes
		WHERE deadline IS NOT NULL
		  AND deadline >= CURRENT_DATE
		  AND deadline <= CURRENT_DATE + INTERVAL '180 days'
		  AND status NOT IN ('implemented', 'not_applicable')
		ORDER BY deadline ASC
		LIMIT 5`)
	if err == nil {
		defer regRows.Close()
		for regRows.Next() {
			var item RegulatoryHorizonItem
			if regRows.Scan(&item.Title, &item.Severity, &item.Deadline) == nil {
				if t, parseErr := time.Parse("2006-01-02", item.Deadline); parseErr == nil {
					item.DaysLeft = int(time.Until(t).Hours() / 24)
				}
				dash.RegulatoryHorizon = append(dash.RegulatoryHorizon, item)
			}
		}
	}

	// 7. Next meeting
	var nm BoardMeeting
	err = tx.QueryRow(ctx, `
		SELECT id, organization_id, meeting_ref, title, meeting_type::text,
		       date::text, time::text, COALESCE(location, ''),
		       status::text, agenda_items,
		       board_pack_document_path, board_pack_generated_at,
		       minutes_document_path, minutes_approved_at, minutes_approved_by,
		       attendees, apologies,
		       created_at, updated_at
		FROM board_meetings
		WHERE date >= CURRENT_DATE AND status IN ('planned', 'agenda_set')
		ORDER BY date ASC
		LIMIT 1`).Scan(
		&nm.ID, &nm.OrganizationID, &nm.MeetingRef, &nm.Title, &nm.MeetingType,
		&nm.Date, &nm.Time, &nm.Location,
		&nm.Status, &nm.AgendaItems,
		&nm.BoardPackDocumentPath, &nm.BoardPackGeneratedAt,
		&nm.MinutesDocumentPath, &nm.MinutesApprovedAt, &nm.MinutesApprovedBy,
		&nm.Attendees, &nm.Apologies,
		&nm.CreatedAt, &nm.UpdatedAt,
	)
	if err == nil {
		dash.NextMeeting = &nm
	}

	// 8. Action counts
	_ = tx.QueryRow(ctx, `
		SELECT COUNT(*) FROM board_decisions
		WHERE action_required = true AND action_status IN ('pending', 'in_progress')`).Scan(&dash.PendingActions)
	_ = tx.QueryRow(ctx, `
		SELECT COUNT(*) FROM board_decisions
		WHERE action_required = true AND action_status = 'overdue'`).Scan(&dash.OverdueActions)

	// Also check for overdue by date
	_, _ = tx.Exec(ctx, `
		UPDATE board_decisions
		SET action_status = 'overdue'::board_action_status
		WHERE action_required = true
		  AND action_status IN ('pending', 'in_progress')
		  AND action_due_date < CURRENT_DATE`)

	// 9. Total members
	_ = tx.QueryRow(ctx, `SELECT COUNT(*) FROM board_members WHERE is_active = true`).Scan(&dash.TotalMembers)

	// 10. Reports generated
	_ = tx.QueryRow(ctx, `SELECT COUNT(*) FROM board_reports`).Scan(&dash.ReportsGenerated)

	// 11. Last board pack
	var lastPackAt *time.Time
	err = tx.QueryRow(ctx, `
		SELECT MAX(generated_at) FROM board_reports WHERE report_type = 'board_pack'`).Scan(&lastPackAt)
	if err == nil {
		dash.LastBoardPackAt = lastPackAt
	}

	return dash, nil
}

// ============================================================
// REPORTS
// ============================================================

// ListReports returns all board reports for the organisation.
func (s *BoardService) ListReports(ctx context.Context, orgID uuid.UUID) ([]BoardReport, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.setOrgRLS(ctx, tx, orgID); err != nil {
		return nil, fmt.Errorf("set RLS: %w", err)
	}

	rows, err := tx.Query(ctx, `
		SELECT id, organization_id, meeting_id, report_type::text, title,
		       period_start::text, period_end::text, COALESCE(file_path, ''),
		       file_format::text, generated_by, generated_at,
		       classification, page_count, created_at
		FROM board_reports
		ORDER BY generated_at DESC
		LIMIT 100`)
	if err != nil {
		return nil, fmt.Errorf("query reports: %w", err)
	}
	defer rows.Close()

	var reports []BoardReport
	for rows.Next() {
		var r BoardReport
		if err := rows.Scan(
			&r.ID, &r.OrganizationID, &r.MeetingID, &r.ReportType, &r.Title,
			&r.PeriodStart, &r.PeriodEnd, &r.FilePath,
			&r.FileFormat, &r.GeneratedBy, &r.GeneratedAt,
			&r.Classification, &r.PageCount, &r.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan report: %w", err)
		}
		reports = append(reports, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate reports: %w", err)
	}
	if reports == nil {
		reports = []BoardReport{}
	}
	return reports, nil
}

// GenerateReport generates a board report of the specified type.
func (s *BoardService) GenerateReport(ctx context.Context, orgID uuid.UUID, req BoardGenerateReportRequest) (*BoardReport, error) {
	if req.ReportType == "" {
		req.ReportType = "custom"
	}
	if req.Title == "" {
		req.Title = fmt.Sprintf("Board Report - %s", time.Now().Format("2006-01-02"))
	}
	if req.FileFormat == "" {
		req.FileFormat = "pdf"
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.setOrgRLS(ctx, tx, orgID); err != nil {
		return nil, fmt.Errorf("set RLS: %w", err)
	}

	filePath := fmt.Sprintf("board-reports/%s/%s-%s.%s",
		orgID.String(), req.ReportType, time.Now().Format("2006-01-02"), req.FileFormat)

	var periodStartArg, periodEndArg interface{}
	if req.PeriodStart != "" {
		periodStartArg = req.PeriodStart
	}
	if req.PeriodEnd != "" {
		periodEndArg = req.PeriodEnd
	}

	var report BoardReport
	err = tx.QueryRow(ctx, `
		INSERT INTO board_reports
			(organization_id, report_type, title,
			 period_start, period_end, file_path, file_format,
			 generated_at, classification, page_count)
		VALUES ($1, $2::board_report_type, $3,
		        $4::date, $5::date, $6, $7::board_report_format,
		        NOW(), 'board_confidential', 0)
		RETURNING id, organization_id, meeting_id, report_type::text, title,
		          period_start::text, period_end::text, COALESCE(file_path, ''),
		          file_format::text, generated_by, generated_at,
		          classification, page_count, created_at`,
		orgID, req.ReportType, req.Title,
		periodStartArg, periodEndArg, filePath, req.FileFormat,
	).Scan(
		&report.ID, &report.OrganizationID, &report.MeetingID,
		&report.ReportType, &report.Title,
		&report.PeriodStart, &report.PeriodEnd, &report.FilePath,
		&report.FileFormat, &report.GeneratedBy, &report.GeneratedAt,
		&report.Classification, &report.PageCount, &report.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert report: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	log.Info().
		Str("report_id", report.ID.String()).
		Str("type", report.ReportType).
		Msg("board report generated")

	return &report, nil
}

// GetBoardPackPath returns the file path of the board pack for a meeting.
func (s *BoardService) GetBoardPackPath(ctx context.Context, orgID, meetingID uuid.UUID) (string, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.setOrgRLS(ctx, tx, orgID); err != nil {
		return "", fmt.Errorf("set RLS: %w", err)
	}

	var path string
	err = tx.QueryRow(ctx, `
		SELECT COALESCE(board_pack_document_path, '')
		FROM board_meetings WHERE id = $1`, meetingID).Scan(&path)
	if err != nil {
		return "", fmt.Errorf("get board pack path: %w", err)
	}
	if path == "" {
		return "", fmt.Errorf("board pack not yet generated for this meeting")
	}
	return path, nil
}

// ============================================================
// PORTAL — TOKEN-BASED ACCESS
// ============================================================

// resolveTokenToOrg resolves a portal access token hash to an org ID,
// verifying the token is valid and not expired.
func (s *BoardService) resolveTokenToOrg(ctx context.Context, tokenHash string) (uuid.UUID, error) {
	var orgID uuid.UUID
	var memberID uuid.UUID
	err := s.pool.QueryRow(ctx, `
		SELECT id, organization_id
		FROM board_members
		WHERE portal_access_token_hash = $1
		  AND portal_access_enabled = true
		  AND is_active = true
		  AND (portal_access_expires_at IS NULL OR portal_access_expires_at > NOW())`,
		tokenHash,
	).Scan(&memberID, &orgID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return uuid.Nil, fmt.Errorf("invalid or expired portal access token")
		}
		return uuid.Nil, fmt.Errorf("resolve token: %w", err)
	}

	// Update last access time
	_, _ = s.pool.Exec(ctx, `
		UPDATE board_members SET last_portal_access_at = NOW() WHERE id = $1`, memberID)

	return orgID, nil
}

// GetDashboardByToken returns the board dashboard for a token-based portal session.
func (s *BoardService) GetDashboardByToken(ctx context.Context, tokenHash string) (*BoardDashboard, error) {
	orgID, err := s.resolveTokenToOrg(ctx, tokenHash)
	if err != nil {
		return nil, err
	}
	return s.GetBoardDashboard(ctx, orgID)
}

// GetMeetingsByToken returns upcoming meetings for a token-based portal session.
func (s *BoardService) GetMeetingsByToken(ctx context.Context, tokenHash string) ([]BoardMeeting, error) {
	orgID, err := s.resolveTokenToOrg(ctx, tokenHash)
	if err != nil {
		return nil, err
	}
	return s.ListMeetings(ctx, orgID, MeetingFilter{
		Page:     1,
		PageSize: 20,
	})
}

// GetDecisionsByToken returns recent decisions for a token-based portal session.
func (s *BoardService) GetDecisionsByToken(ctx context.Context, tokenHash string) ([]BoardDecision, error) {
	orgID, err := s.resolveTokenToOrg(ctx, tokenHash)
	if err != nil {
		return nil, err
	}
	return s.ListDecisions(ctx, orgID, DecisionFilter{
		Page:     1,
		PageSize: 20,
	})
}

// GetBoardPackPathByToken returns the board pack path for portal users.
func (s *BoardService) GetBoardPackPathByToken(ctx context.Context, tokenHash string, meetingID uuid.UUID) (string, error) {
	orgID, err := s.resolveTokenToOrg(ctx, tokenHash)
	if err != nil {
		return "", err
	}
	return s.GetBoardPackPath(ctx, orgID, meetingID)
}
