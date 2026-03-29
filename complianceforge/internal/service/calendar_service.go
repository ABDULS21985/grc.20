package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
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

// CalendarEvent represents a single compliance calendar event.
type CalendarEvent struct {
	ID                  uuid.UUID       `json:"id"`
	OrganizationID      uuid.UUID       `json:"organization_id"`
	EventRef            string          `json:"event_ref"`
	EventType           string          `json:"event_type"`
	Category            string          `json:"category"`
	Priority            string          `json:"priority"`
	Status              string          `json:"status"`
	Title               string          `json:"title"`
	Description         string          `json:"description"`
	SourceEntityType    string          `json:"source_entity_type"`
	SourceEntityID      uuid.UUID       `json:"source_entity_id"`
	SourceEntityRef     string          `json:"source_entity_ref"`
	DueDate             string          `json:"due_date"`
	DueTime             *string         `json:"due_time"`
	StartDate           *string         `json:"start_date"`
	EndDate             *string         `json:"end_date"`
	AllDay              bool            `json:"all_day"`
	Timezone            string          `json:"timezone"`
	RecurrenceType      string          `json:"recurrence_type"`
	Rrule               string          `json:"rrule"`
	RecurrenceEndDate   *string         `json:"recurrence_end_date"`
	ParentEventID       *uuid.UUID      `json:"parent_event_id"`
	OccurrenceDate      *string         `json:"occurrence_date"`
	AssignedToUserID    *uuid.UUID      `json:"assigned_to_user_id"`
	AssignedToRole      string          `json:"assigned_to_role"`
	OwnerUserID         *uuid.UUID      `json:"owner_user_id"`
	ReminderDaysBefore  []int           `json:"reminder_days_before"`
	LastReminderSentAt  *time.Time      `json:"last_reminder_sent_at"`
	ReminderCount       int             `json:"reminder_count"`
	EscalationAfterDays int             `json:"escalation_after_days"`
	Escalated           bool            `json:"escalated"`
	EscalatedToUserID   *uuid.UUID      `json:"escalated_to_user_id"`
	EscalatedAt         *time.Time      `json:"escalated_at"`
	CompletedAt         *time.Time      `json:"completed_at"`
	CompletedBy         *uuid.UUID      `json:"completed_by"`
	CompletionNotes     string          `json:"completion_notes"`
	OriginalDueDate     *string         `json:"original_due_date"`
	RescheduledCount    int             `json:"rescheduled_count"`
	RescheduleReason    string          `json:"reschedule_reason"`
	URL                 string          `json:"url"`
	Color               string          `json:"color"`
	Tags                []string        `json:"tags"`
	Metadata            json.RawMessage `json:"metadata"`
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`
	// Joined fields
	AssignedToName string `json:"assigned_to_name,omitempty"`
	OwnerName      string `json:"owner_name,omitempty"`
}

// CalendarSubscription represents per-user calendar preferences.
type CalendarSubscription struct {
	ID                   uuid.UUID  `json:"id"`
	OrganizationID       uuid.UUID  `json:"organization_id"`
	UserID               uuid.UUID  `json:"user_id"`
	SubscribedCategories []string   `json:"subscribed_categories"`
	SubscribedPriorities []string   `json:"subscribed_priorities"`
	EmailReminders       bool       `json:"email_reminders"`
	InAppReminders       bool       `json:"in_app_reminders"`
	DailyDigest          bool       `json:"daily_digest"`
	WeeklyDigest         bool       `json:"weekly_digest"`
	ICalExportEnabled    bool       `json:"ical_export_enabled"`
	ICalToken            string     `json:"ical_token,omitempty"`
	ICalTokenExpiresAt   *time.Time `json:"ical_token_expires_at"`
	ReminderDaysOverride []int      `json:"reminder_days_override"`
	QuietHoursStart      *string    `json:"quiet_hours_start"`
	QuietHoursEnd        *string    `json:"quiet_hours_end"`
	Timezone             string     `json:"timezone"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

// CalendarSyncConfig represents per-module sync settings.
type CalendarSyncConfig struct {
	ID                    uuid.UUID  `json:"id"`
	OrganizationID        uuid.UUID  `json:"organization_id"`
	ModuleName            string     `json:"module_name"`
	IsEnabled             bool       `json:"is_enabled"`
	SyncFrequencyMinutes  int        `json:"sync_frequency_minutes"`
	LastSyncAt            *time.Time `json:"last_sync_at"`
	LastSyncStatus        string     `json:"last_sync_status"`
	LastSyncEventsCreated int        `json:"last_sync_events_created"`
	LastSyncEventsUpdated int        `json:"last_sync_events_updated"`
	LastSyncError         string     `json:"last_sync_error"`
	AutoCreateEvents      bool       `json:"auto_create_events"`
	AutoCompleteEvents    bool       `json:"auto_complete_events"`
	DefaultReminderDays   []int      `json:"default_reminder_days"`
	DefaultPriority       string     `json:"default_priority"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

// CalendarSyncResult holds the result of a sync operation.
type CalendarSyncResult struct {
	Module        string `json:"module"`
	EventsCreated int    `json:"events_created"`
	EventsUpdated int    `json:"events_updated"`
	EventsSkipped int    `json:"events_skipped"`
	Errors        int    `json:"errors"`
	Duration      string `json:"duration"`
}

// ============================================================
// REQUEST TYPES
// ============================================================

// CreateCalendarEventRequest is the request body for creating a calendar event.
type CreateCalendarEventRequest struct {
	EventType        string     `json:"event_type"`
	Category         string     `json:"category"`
	Priority         string     `json:"priority"`
	Title            string     `json:"title"`
	Description      string     `json:"description"`
	SourceEntityType string     `json:"source_entity_type"`
	SourceEntityID   uuid.UUID  `json:"source_entity_id"`
	SourceEntityRef  string     `json:"source_entity_ref"`
	DueDate          string     `json:"due_date"`
	AllDay           bool       `json:"all_day"`
	RecurrenceType   string     `json:"recurrence_type"`
	Rrule            string     `json:"rrule"`
	AssignedToUserID *uuid.UUID `json:"assigned_to_user_id"`
	OwnerUserID      *uuid.UUID `json:"owner_user_id"`
	Tags             []string   `json:"tags"`
}

// CompleteEventRequest is the request body for completing a calendar event.
type CompleteEventRequest struct {
	CompletionNotes string `json:"completion_notes"`
}

// RescheduleEventRequest is the request body for rescheduling a calendar event.
type RescheduleEventRequest struct {
	NewDueDate string `json:"new_due_date"`
	Reason     string `json:"reason"`
}

// AssignEventRequest is the request body for assigning a calendar event.
type AssignEventRequest struct {
	AssignedToUserID uuid.UUID `json:"assigned_to_user_id"`
}

// UpdateSubscriptionRequest is the request body for updating calendar preferences.
type UpdateSubscriptionRequest struct {
	SubscribedCategories []string `json:"subscribed_categories"`
	SubscribedPriorities []string `json:"subscribed_priorities"`
	EmailReminders       *bool    `json:"email_reminders"`
	InAppReminders       *bool    `json:"in_app_reminders"`
	DailyDigest          *bool    `json:"daily_digest"`
	WeeklyDigest         *bool    `json:"weekly_digest"`
	ICalExportEnabled    *bool    `json:"ical_export_enabled"`
	ReminderDaysOverride []int    `json:"reminder_days_override"`
	Timezone             string   `json:"timezone"`
}

// UpdateSyncConfigRequest is the request body for updating sync configuration.
type UpdateSyncConfigRequest struct {
	Configs []SyncConfigEntry `json:"configs"`
}

// SyncConfigEntry represents a single sync config update.
type SyncConfigEntry struct {
	ModuleName           string `json:"module_name"`
	IsEnabled            *bool  `json:"is_enabled"`
	SyncFrequencyMinutes *int   `json:"sync_frequency_minutes"`
	AutoCreateEvents     *bool  `json:"auto_create_events"`
	AutoCompleteEvents   *bool  `json:"auto_complete_events"`
	DefaultPriority      string `json:"default_priority"`
}

// SyncStatus represents the overall sync status across modules.
type SyncStatus struct {
	Modules       []CalendarSyncConfig `json:"modules"`
	LastFullSync  *time.Time           `json:"last_full_sync"`
	TotalEvents   int64                `json:"total_events"`
	OverdueEvents int64                `json:"overdue_events"`
}

// ============================================================
// SERVICE
// ============================================================

// CalendarService manages compliance calendar events and sync.
type CalendarService struct {
	pool *pgxpool.Pool
}

// NewCalendarService creates a new CalendarService.
func NewCalendarService(pool *pgxpool.Pool) *CalendarService {
	return &CalendarService{pool: pool}
}

// ============================================================
// CRUD OPERATIONS
// ============================================================

// CreateEvent creates a new calendar event.
func (s *CalendarService) CreateEvent(ctx context.Context, orgID uuid.UUID, req CreateCalendarEventRequest) (*CalendarEvent, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return nil, fmt.Errorf("set org: %w", err)
	}

	recurrenceType := req.RecurrenceType
	if recurrenceType == "" {
		recurrenceType = "none"
	}
	priority := req.Priority
	if priority == "" {
		priority = "medium"
	}

	var event CalendarEvent
	err = tx.QueryRow(ctx, `
		INSERT INTO calendar_events (
			organization_id, event_type, category, priority, status, title, description,
			source_entity_type, source_entity_id, source_entity_ref,
			due_date, all_day, recurrence_type, rrule,
			assigned_to_user_id, owner_user_id, tags
		) VALUES (
			$1, $2, $3, $4, 'upcoming', $5, $6,
			$7, $8, $9,
			$10::DATE, $11, $12, $13,
			$14, $15, $16
		)
		RETURNING id, event_ref, created_at, updated_at`,
		orgID, req.EventType, req.Category, priority, req.Title, req.Description,
		req.SourceEntityType, req.SourceEntityID, req.SourceEntityRef,
		req.DueDate, req.AllDay, recurrenceType, req.Rrule,
		req.AssignedToUserID, req.OwnerUserID, req.Tags,
	).Scan(&event.ID, &event.EventRef, &event.CreatedAt, &event.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert event: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	event.OrganizationID = orgID
	event.EventType = req.EventType
	event.Category = req.Category
	event.Priority = priority
	event.Status = "upcoming"
	event.Title = req.Title
	event.Description = req.Description
	event.SourceEntityType = req.SourceEntityType
	event.SourceEntityID = req.SourceEntityID
	event.SourceEntityRef = req.SourceEntityRef
	event.DueDate = req.DueDate
	event.AllDay = req.AllDay
	event.RecurrenceType = recurrenceType
	event.Rrule = req.Rrule
	event.AssignedToUserID = req.AssignedToUserID
	event.OwnerUserID = req.OwnerUserID
	event.Tags = req.Tags

	return &event, nil
}

// GetEvent returns a single calendar event by ID.
func (s *CalendarService) GetEvent(ctx context.Context, orgID, eventID uuid.UUID) (*CalendarEvent, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return nil, err
	}

	event, err := s.scanEvent(tx.QueryRow(ctx, `
		SELECT
			ce.id, ce.organization_id, ce.event_ref, ce.event_type, ce.category,
			ce.priority, ce.status, ce.title, ce.description,
			ce.source_entity_type, ce.source_entity_id, ce.source_entity_ref,
			ce.due_date, ce.due_time, ce.start_date, ce.end_date,
			ce.all_day, ce.timezone, ce.recurrence_type, ce.rrule,
			ce.recurrence_end_date, ce.parent_event_id, ce.occurrence_date,
			ce.assigned_to_user_id, ce.assigned_to_role, ce.owner_user_id,
			ce.reminder_days_before, ce.last_reminder_sent_at, ce.reminder_count,
			ce.escalation_after_days, ce.escalated, ce.escalated_to_user_id, ce.escalated_at,
			ce.completed_at, ce.completed_by, ce.completion_notes,
			ce.original_due_date, ce.rescheduled_count, ce.reschedule_reason,
			ce.url, ce.color, ce.tags, ce.metadata,
			ce.created_at, ce.updated_at,
			COALESCE(ua.first_name || ' ' || ua.last_name, ''),
			COALESCE(uo.first_name || ' ' || uo.last_name, '')
		FROM calendar_events ce
		LEFT JOIN users ua ON ua.id = ce.assigned_to_user_id
		LEFT JOIN users uo ON uo.id = ce.owner_user_id
		WHERE ce.id = $1 AND ce.organization_id = $2`, eventID, orgID))
	if err != nil {
		return nil, err
	}

	tx.Commit(ctx)
	return event, nil
}

// CompleteEvent marks a calendar event as completed.
func (s *CalendarService) CompleteEvent(ctx context.Context, orgID, eventID, userID uuid.UUID, req CompleteEventRequest) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return err
	}

	tag, err := tx.Exec(ctx, `
		UPDATE calendar_events
		SET status = 'completed', completed_at = NOW(), completed_by = $3, completion_notes = $4
		WHERE id = $1 AND organization_id = $2 AND status NOT IN ('completed', 'cancelled')`,
		eventID, orgID, userID, req.CompletionNotes)
	if err != nil {
		return fmt.Errorf("update event: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("event not found or already completed")
	}

	return tx.Commit(ctx)
}

// RescheduleEvent reschedules a calendar event to a new due date.
func (s *CalendarService) RescheduleEvent(ctx context.Context, orgID, eventID uuid.UUID, req RescheduleEventRequest) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return err
	}

	tag, err := tx.Exec(ctx, `
		UPDATE calendar_events
		SET due_date = $3::DATE,
			original_due_date = COALESCE(original_due_date, due_date),
			rescheduled_count = rescheduled_count + 1,
			reschedule_reason = $4,
			status = 'upcoming'
		WHERE id = $1 AND organization_id = $2 AND status NOT IN ('completed', 'cancelled')`,
		eventID, orgID, req.NewDueDate, req.Reason)
	if err != nil {
		return fmt.Errorf("reschedule event: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("event not found or already completed/cancelled")
	}

	return tx.Commit(ctx)
}

// AssignEvent assigns a calendar event to a user.
func (s *CalendarService) AssignEvent(ctx context.Context, orgID, eventID uuid.UUID, req AssignEventRequest) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return err
	}

	tag, err := tx.Exec(ctx, `
		UPDATE calendar_events
		SET assigned_to_user_id = $3
		WHERE id = $1 AND organization_id = $2`,
		eventID, orgID, req.AssignedToUserID)
	if err != nil {
		return fmt.Errorf("assign event: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("event not found")
	}

	return tx.Commit(ctx)
}

// ============================================================
// SUBSCRIPTION MANAGEMENT
// ============================================================

// GetSubscription returns the calendar subscription for a user.
func (s *CalendarService) GetSubscription(ctx context.Context, orgID, userID uuid.UUID) (*CalendarSubscription, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return nil, err
	}

	var sub CalendarSubscription
	err = tx.QueryRow(ctx, `
		SELECT id, organization_id, user_id, subscribed_categories, subscribed_priorities,
			   email_reminders, in_app_reminders, daily_digest, weekly_digest,
			   ical_export_enabled, COALESCE(ical_token, ''), ical_token_expires_at,
			   reminder_days_override, quiet_hours_start::TEXT, quiet_hours_end::TEXT,
			   timezone, created_at, updated_at
		FROM calendar_subscriptions
		WHERE organization_id = $1 AND user_id = $2`, orgID, userID,
	).Scan(
		&sub.ID, &sub.OrganizationID, &sub.UserID,
		&sub.SubscribedCategories, &sub.SubscribedPriorities,
		&sub.EmailReminders, &sub.InAppReminders, &sub.DailyDigest, &sub.WeeklyDigest,
		&sub.ICalExportEnabled, &sub.ICalToken, &sub.ICalTokenExpiresAt,
		&sub.ReminderDaysOverride, &sub.QuietHoursStart, &sub.QuietHoursEnd,
		&sub.Timezone, &sub.CreatedAt, &sub.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			// Return default subscription
			return s.createDefaultSubscription(ctx, tx, orgID, userID)
		}
		return nil, err
	}

	tx.Commit(ctx)
	return &sub, nil
}

// createDefaultSubscription creates and returns a default subscription.
func (s *CalendarService) createDefaultSubscription(ctx context.Context, tx pgx.Tx, orgID, userID uuid.UUID) (*CalendarSubscription, error) {
	var sub CalendarSubscription
	err := tx.QueryRow(ctx, `
		INSERT INTO calendar_subscriptions (organization_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT (organization_id, user_id) DO NOTHING
		RETURNING id, organization_id, user_id, subscribed_categories, subscribed_priorities,
			email_reminders, in_app_reminders, daily_digest, weekly_digest,
			ical_export_enabled, COALESCE(ical_token, ''), ical_token_expires_at,
			reminder_days_override, quiet_hours_start::TEXT, quiet_hours_end::TEXT,
			timezone, created_at, updated_at`,
		orgID, userID,
	).Scan(
		&sub.ID, &sub.OrganizationID, &sub.UserID,
		&sub.SubscribedCategories, &sub.SubscribedPriorities,
		&sub.EmailReminders, &sub.InAppReminders, &sub.DailyDigest, &sub.WeeklyDigest,
		&sub.ICalExportEnabled, &sub.ICalToken, &sub.ICalTokenExpiresAt,
		&sub.ReminderDaysOverride, &sub.QuietHoursStart, &sub.QuietHoursEnd,
		&sub.Timezone, &sub.CreatedAt, &sub.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create default subscription: %w", err)
	}

	tx.Commit(ctx)
	return &sub, nil
}

// UpdateSubscription updates a user's calendar subscription.
func (s *CalendarService) UpdateSubscription(ctx context.Context, orgID, userID uuid.UUID, req UpdateSubscriptionRequest) (*CalendarSubscription, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return nil, err
	}

	// Generate iCal token if enabling export
	var icalToken *string
	var icalTokenExpires *time.Time
	if req.ICalExportEnabled != nil && *req.ICalExportEnabled {
		token, err := generateICalToken()
		if err != nil {
			return nil, fmt.Errorf("generate ical token: %w", err)
		}
		icalToken = &token
		expires := time.Now().Add(365 * 24 * time.Hour)
		icalTokenExpires = &expires
	}

	tz := req.Timezone
	if tz == "" {
		tz = "UTC"
	}

	var sub CalendarSubscription
	err = tx.QueryRow(ctx, `
		INSERT INTO calendar_subscriptions (
			organization_id, user_id,
			subscribed_categories, subscribed_priorities,
			email_reminders, in_app_reminders, daily_digest, weekly_digest,
			ical_export_enabled, ical_token, ical_token_expires_at,
			reminder_days_override, timezone
		) VALUES (
			$1, $2,
			COALESCE($3::calendar_event_category[], '{policy,risk,vendor,audit,evidence,exception,dsr,incident,regulatory,business_continuity,board,custom}'::calendar_event_category[]),
			COALESCE($4::calendar_event_priority[], '{critical,high,medium,low}'::calendar_event_priority[]),
			COALESCE($5, true), COALESCE($6, true), COALESCE($7, false), COALESCE($8, true),
			COALESCE($9, false), $10, $11,
			$12, $13
		)
		ON CONFLICT (organization_id, user_id) DO UPDATE SET
			subscribed_categories = COALESCE(EXCLUDED.subscribed_categories, calendar_subscriptions.subscribed_categories),
			subscribed_priorities = COALESCE(EXCLUDED.subscribed_priorities, calendar_subscriptions.subscribed_priorities),
			email_reminders = COALESCE(EXCLUDED.email_reminders, calendar_subscriptions.email_reminders),
			in_app_reminders = COALESCE(EXCLUDED.in_app_reminders, calendar_subscriptions.in_app_reminders),
			daily_digest = COALESCE(EXCLUDED.daily_digest, calendar_subscriptions.daily_digest),
			weekly_digest = COALESCE(EXCLUDED.weekly_digest, calendar_subscriptions.weekly_digest),
			ical_export_enabled = COALESCE(EXCLUDED.ical_export_enabled, calendar_subscriptions.ical_export_enabled),
			ical_token = COALESCE(EXCLUDED.ical_token, calendar_subscriptions.ical_token),
			ical_token_expires_at = COALESCE(EXCLUDED.ical_token_expires_at, calendar_subscriptions.ical_token_expires_at),
			reminder_days_override = COALESCE(EXCLUDED.reminder_days_override, calendar_subscriptions.reminder_days_override),
			timezone = COALESCE(EXCLUDED.timezone, calendar_subscriptions.timezone)
		RETURNING id, organization_id, user_id, subscribed_categories, subscribed_priorities,
			email_reminders, in_app_reminders, daily_digest, weekly_digest,
			ical_export_enabled, COALESCE(ical_token, ''), ical_token_expires_at,
			reminder_days_override, quiet_hours_start::TEXT, quiet_hours_end::TEXT,
			timezone, created_at, updated_at`,
		orgID, userID,
		req.SubscribedCategories, req.SubscribedPriorities,
		req.EmailReminders, req.InAppReminders, req.DailyDigest, req.WeeklyDigest,
		req.ICalExportEnabled, icalToken, icalTokenExpires,
		req.ReminderDaysOverride, tz,
	).Scan(
		&sub.ID, &sub.OrganizationID, &sub.UserID,
		&sub.SubscribedCategories, &sub.SubscribedPriorities,
		&sub.EmailReminders, &sub.InAppReminders, &sub.DailyDigest, &sub.WeeklyDigest,
		&sub.ICalExportEnabled, &sub.ICalToken, &sub.ICalTokenExpiresAt,
		&sub.ReminderDaysOverride, &sub.QuietHoursStart, &sub.QuietHoursEnd,
		&sub.Timezone, &sub.CreatedAt, &sub.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("upsert subscription: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &sub, nil
}

// ============================================================
// SYNC ENGINE
// ============================================================

// SyncAllModules performs a full sync across all enabled modules.
func (s *CalendarService) SyncAllModules(ctx context.Context, orgID uuid.UUID) ([]CalendarSyncResult, error) {
	var results []CalendarSyncResult

	syncFuncs := map[string]func(context.Context, uuid.UUID) (*CalendarSyncResult, error){
		"policies":            s.SyncPolicyEvents,
		"risks":               s.SyncRiskEvents,
		"vendors":             s.SyncVendorEvents,
		"audits":              s.SyncAuditEvents,
		"evidence":            s.SyncEvidenceEvents,
		"exceptions":          s.SyncExceptionEvents,
		"dsr":                 s.SyncDSREvents,
		"incidents":           s.SyncIncidentEvents,
		"regulatory":          s.SyncRegulatoryEvents,
		"business_continuity": s.SyncBCEvents,
		"board":               s.SyncBoardEvents,
	}

	for module, syncFn := range syncFuncs {
		enabled, err := s.isModuleEnabled(ctx, orgID, module)
		if err != nil {
			log.Error().Err(err).Str("module", module).Msg("Failed to check module status")
			continue
		}
		if !enabled {
			continue
		}

		result, err := syncFn(ctx, orgID)
		if err != nil {
			log.Error().Err(err).Str("module", module).Msg("Sync failed")
			s.updateSyncStatus(ctx, orgID, module, "error", 0, 0, err.Error())
			results = append(results, CalendarSyncResult{Module: module, Errors: 1})
			continue
		}
		if result != nil {
			results = append(results, *result)
			s.updateSyncStatus(ctx, orgID, module, "success", result.EventsCreated, result.EventsUpdated, "")
		}
	}

	return results, nil
}

// SyncPolicyEvents syncs policy review and expiry deadlines to the calendar.
func (s *CalendarService) SyncPolicyEvents(ctx context.Context, orgID uuid.UUID) (*CalendarSyncResult, error) {
	start := time.Now()
	result := &CalendarSyncResult{Module: "policies"}

	rows, err := s.pool.Query(ctx, `
		SELECT p.id, p.policy_ref, p.title, p.next_review_date, p.expiry_date,
			   p.owner_user_id, p.status
		FROM policies p
		WHERE p.organization_id = $1
		  AND p.deleted_at IS NULL
		  AND p.status NOT IN ('retired', 'archived')
		  AND (p.next_review_date IS NOT NULL OR p.expiry_date IS NOT NULL)`, orgID)
	if err != nil {
		return nil, fmt.Errorf("query policies: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id uuid.UUID
		var ref, title, status string
		var nextReview, expiry *time.Time
		var ownerID *uuid.UUID

		if err := rows.Scan(&id, &ref, &title, &nextReview, &expiry, &ownerID, &status); err != nil {
			log.Error().Err(err).Msg("scan policy row")
			result.Errors++
			continue
		}

		if nextReview != nil {
			created, updated, err := s.upsertSyncedEvent(ctx, orgID, UpsertEventParams{
				SourceEntityType: "policy",
				SourceEntityID:   id,
				SourceEntityRef:  ref,
				EventType:        "policy_review_due",
				Category:         "policy",
				Priority:         "medium",
				Title:            fmt.Sprintf("Policy review due: %s", title),
				Description:      fmt.Sprintf("Periodic review required for policy %s (%s)", ref, title),
				DueDate:          nextReview.Format("2006-01-02"),
				AssignedTo:       ownerID,
				RecurrenceType:   "annually",
			})
			if err != nil {
				log.Error().Err(err).Str("policy", ref).Msg("upsert policy review event")
				result.Errors++
			} else if created {
				result.EventsCreated++
			} else if updated {
				result.EventsUpdated++
			} else {
				result.EventsSkipped++
			}
		}

		if expiry != nil {
			created, updated, err := s.upsertSyncedEvent(ctx, orgID, UpsertEventParams{
				SourceEntityType: "policy",
				SourceEntityID:   id,
				SourceEntityRef:  ref,
				EventType:        "policy_expiry",
				Category:         "policy",
				Priority:         "high",
				Title:            fmt.Sprintf("Policy expiring: %s", title),
				Description:      fmt.Sprintf("Policy %s (%s) expires on this date", ref, title),
				DueDate:          expiry.Format("2006-01-02"),
				AssignedTo:       ownerID,
			})
			if err != nil {
				log.Error().Err(err).Str("policy", ref).Msg("upsert policy expiry event")
				result.Errors++
			} else if created {
				result.EventsCreated++
			} else if updated {
				result.EventsUpdated++
			} else {
				result.EventsSkipped++
			}
		}
	}

	result.Duration = time.Since(start).String()
	return result, nil
}

// SyncRiskEvents syncs risk review and treatment deadlines to the calendar.
func (s *CalendarService) SyncRiskEvents(ctx context.Context, orgID uuid.UUID) (*CalendarSyncResult, error) {
	start := time.Now()
	result := &CalendarSyncResult{Module: "risks"}

	rows, err := s.pool.Query(ctx, `
		SELECT r.id, r.risk_ref, r.title, r.next_review_date, r.treatment_due_date,
			   r.owner_user_id, r.risk_level
		FROM risks r
		WHERE r.organization_id = $1
		  AND r.deleted_at IS NULL
		  AND r.status NOT IN ('closed')
		  AND (r.next_review_date IS NOT NULL OR r.treatment_due_date IS NOT NULL)`, orgID)
	if err != nil {
		return nil, fmt.Errorf("query risks: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id uuid.UUID
		var ref, title, riskLevel string
		var nextReview, treatmentDue *time.Time
		var ownerID *uuid.UUID

		if err := rows.Scan(&id, &ref, &title, &nextReview, &treatmentDue, &ownerID, &riskLevel); err != nil {
			log.Error().Err(err).Msg("scan risk row")
			result.Errors++
			continue
		}

		priority := riskLevelToPriority(riskLevel)

		if nextReview != nil {
			created, updated, err := s.upsertSyncedEvent(ctx, orgID, UpsertEventParams{
				SourceEntityType: "risk",
				SourceEntityID:   id,
				SourceEntityRef:  ref,
				EventType:        "risk_review_due",
				Category:         "risk",
				Priority:         priority,
				Title:            fmt.Sprintf("Risk review due: %s", title),
				Description:      fmt.Sprintf("Periodic review for risk %s (%s) - Level: %s", ref, title, riskLevel),
				DueDate:          nextReview.Format("2006-01-02"),
				AssignedTo:       ownerID,
			})
			if err != nil {
				result.Errors++
			} else if created {
				result.EventsCreated++
			} else if updated {
				result.EventsUpdated++
			} else {
				result.EventsSkipped++
			}
		}

		if treatmentDue != nil {
			created, updated, err := s.upsertSyncedEvent(ctx, orgID, UpsertEventParams{
				SourceEntityType: "risk",
				SourceEntityID:   id,
				SourceEntityRef:  ref,
				EventType:        "risk_treatment_due",
				Category:         "risk",
				Priority:         priority,
				Title:            fmt.Sprintf("Risk treatment due: %s", title),
				Description:      fmt.Sprintf("Treatment plan deadline for risk %s (%s)", ref, title),
				DueDate:          treatmentDue.Format("2006-01-02"),
				AssignedTo:       ownerID,
			})
			if err != nil {
				result.Errors++
			} else if created {
				result.EventsCreated++
			} else if updated {
				result.EventsUpdated++
			} else {
				result.EventsSkipped++
			}
		}
	}

	result.Duration = time.Since(start).String()
	return result, nil
}

// SyncVendorEvents syncs vendor assessment and contract deadlines.
func (s *CalendarService) SyncVendorEvents(ctx context.Context, orgID uuid.UUID) (*CalendarSyncResult, error) {
	start := time.Now()
	result := &CalendarSyncResult{Module: "vendors"}

	rows, err := s.pool.Query(ctx, `
		SELECT v.id, v.vendor_ref, v.name, v.next_assessment_date,
			   v.contract_end_date, v.dpa_signed_date, v.owner_user_id, v.risk_tier
		FROM vendors v
		WHERE v.organization_id = $1
		  AND v.deleted_at IS NULL
		  AND v.status NOT IN ('offboarded', 'rejected')
		  AND (v.next_assessment_date IS NOT NULL OR v.contract_end_date IS NOT NULL)`, orgID)
	if err != nil {
		return nil, fmt.Errorf("query vendors: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id uuid.UUID
		var ref, name, riskTier string
		var nextAssessment, contractEnd, dpaSigned *time.Time
		var ownerID *uuid.UUID

		if err := rows.Scan(&id, &ref, &name, &nextAssessment, &contractEnd, &dpaSigned, &ownerID, &riskTier); err != nil {
			result.Errors++
			continue
		}

		priority := riskLevelToPriority(riskTier)

		if nextAssessment != nil {
			created, updated, err := s.upsertSyncedEvent(ctx, orgID, UpsertEventParams{
				SourceEntityType: "vendor",
				SourceEntityID:   id,
				SourceEntityRef:  ref,
				EventType:        "vendor_assessment_due",
				Category:         "vendor",
				Priority:         priority,
				Title:            fmt.Sprintf("Vendor assessment due: %s", name),
				Description:      fmt.Sprintf("Scheduled vendor risk assessment for %s (%s)", ref, name),
				DueDate:          nextAssessment.Format("2006-01-02"),
				AssignedTo:       ownerID,
				RecurrenceType:   "annually",
			})
			if err != nil {
				result.Errors++
			} else if created {
				result.EventsCreated++
			} else if updated {
				result.EventsUpdated++
			} else {
				result.EventsSkipped++
			}
		}

		if contractEnd != nil {
			created, updated, err := s.upsertSyncedEvent(ctx, orgID, UpsertEventParams{
				SourceEntityType: "vendor",
				SourceEntityID:   id,
				SourceEntityRef:  ref,
				EventType:        "vendor_contract_expiry",
				Category:         "vendor",
				Priority:         "high",
				Title:            fmt.Sprintf("Vendor contract expiry: %s", name),
				Description:      fmt.Sprintf("Contract with vendor %s (%s) expires", ref, name),
				DueDate:          contractEnd.Format("2006-01-02"),
				AssignedTo:       ownerID,
			})
			if err != nil {
				result.Errors++
			} else if created {
				result.EventsCreated++
			} else if updated {
				result.EventsUpdated++
			} else {
				result.EventsSkipped++
			}
		}
	}

	result.Duration = time.Since(start).String()
	return result, nil
}

// SyncAuditEvents syncs audit schedule and finding deadlines.
func (s *CalendarService) SyncAuditEvents(ctx context.Context, orgID uuid.UUID) (*CalendarSyncResult, error) {
	start := time.Now()
	result := &CalendarSyncResult{Module: "audits"}

	// Audit schedule
	rows, err := s.pool.Query(ctx, `
		SELECT a.id, a.audit_ref, a.title, a.planned_start_date, a.planned_end_date,
			   a.lead_auditor_id, a.status
		FROM audits a
		WHERE a.organization_id = $1
		  AND a.deleted_at IS NULL
		  AND a.status IN ('planned', 'in_progress')
		  AND (a.planned_start_date IS NOT NULL OR a.planned_end_date IS NOT NULL)`, orgID)
	if err != nil {
		return nil, fmt.Errorf("query audits: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id uuid.UUID
		var ref, title, status string
		var startDate, endDate *time.Time
		var leadID *uuid.UUID

		if err := rows.Scan(&id, &ref, &title, &startDate, &endDate, &leadID, &status); err != nil {
			result.Errors++
			continue
		}

		if startDate != nil {
			created, updated, err := s.upsertSyncedEvent(ctx, orgID, UpsertEventParams{
				SourceEntityType: "audit",
				SourceEntityID:   id,
				SourceEntityRef:  ref,
				EventType:        "audit_planned_start",
				Category:         "audit",
				Priority:         "high",
				Title:            fmt.Sprintf("Audit start: %s", title),
				Description:      fmt.Sprintf("Planned start date for audit %s (%s)", ref, title),
				DueDate:          startDate.Format("2006-01-02"),
				AssignedTo:       leadID,
			})
			if err != nil {
				result.Errors++
			} else if created {
				result.EventsCreated++
			} else if updated {
				result.EventsUpdated++
			} else {
				result.EventsSkipped++
			}
		}

		if endDate != nil {
			created, updated, err := s.upsertSyncedEvent(ctx, orgID, UpsertEventParams{
				SourceEntityType: "audit",
				SourceEntityID:   id,
				SourceEntityRef:  ref,
				EventType:        "audit_planned_end",
				Category:         "audit",
				Priority:         "high",
				Title:            fmt.Sprintf("Audit deadline: %s", title),
				Description:      fmt.Sprintf("Planned end date for audit %s (%s)", ref, title),
				DueDate:          endDate.Format("2006-01-02"),
				AssignedTo:       leadID,
			})
			if err != nil {
				result.Errors++
			} else if created {
				result.EventsCreated++
			} else if updated {
				result.EventsUpdated++
			} else {
				result.EventsSkipped++
			}
		}
	}

	// Audit findings
	fRows, err := s.pool.Query(ctx, `
		SELECT af.id, af.finding_ref, af.title, af.due_date,
			   af.responsible_user_id, af.severity
		FROM audit_findings af
		WHERE af.organization_id = $1
		  AND af.deleted_at IS NULL
		  AND af.status NOT IN ('closed', 'verified')
		  AND af.due_date IS NOT NULL`, orgID)
	if err != nil {
		return nil, fmt.Errorf("query audit findings: %w", err)
	}
	defer fRows.Close()

	for fRows.Next() {
		var id uuid.UUID
		var ref, title, severity string
		var dueDate *time.Time
		var responsibleID *uuid.UUID

		if err := fRows.Scan(&id, &ref, &title, &dueDate, &responsibleID, &severity); err != nil {
			result.Errors++
			continue
		}

		priority := "medium"
		switch severity {
		case "critical":
			priority = "critical"
		case "high":
			priority = "high"
		}

		if dueDate != nil {
			created, updated, err := s.upsertSyncedEvent(ctx, orgID, UpsertEventParams{
				SourceEntityType: "audit_finding",
				SourceEntityID:   id,
				SourceEntityRef:  ref,
				EventType:        "audit_finding_due",
				Category:         "audit",
				Priority:         priority,
				Title:            fmt.Sprintf("Finding remediation due: %s", title),
				Description:      fmt.Sprintf("Audit finding %s (%s) - Severity: %s", ref, title, severity),
				DueDate:          dueDate.Format("2006-01-02"),
				AssignedTo:       responsibleID,
			})
			if err != nil {
				result.Errors++
			} else if created {
				result.EventsCreated++
			} else if updated {
				result.EventsUpdated++
			} else {
				result.EventsSkipped++
			}
		}
	}

	result.Duration = time.Since(start).String()
	return result, nil
}

// SyncEvidenceEvents syncs evidence collection deadlines.
func (s *CalendarService) SyncEvidenceEvents(ctx context.Context, orgID uuid.UUID) (*CalendarSyncResult, error) {
	start := time.Now()
	result := &CalendarSyncResult{Module: "evidence"}

	rows, err := s.pool.Query(ctx, `
		SELECT ec.id, ec.collection_ref, ec.title, ec.next_collection_date,
			   ec.assigned_to_user_id
		FROM evidence_collections ec
		WHERE ec.organization_id = $1
		  AND ec.status NOT IN ('completed', 'cancelled')
		  AND ec.next_collection_date IS NOT NULL`, orgID)
	if err != nil {
		// Table may not exist — that is ok; skip silently
		log.Debug().Err(err).Msg("evidence collections query (table may not exist)")
		result.Duration = time.Since(start).String()
		return result, nil
	}
	defer rows.Close()

	for rows.Next() {
		var id uuid.UUID
		var ref, title string
		var nextCollection *time.Time
		var assignedID *uuid.UUID

		if err := rows.Scan(&id, &ref, &title, &nextCollection, &assignedID); err != nil {
			result.Errors++
			continue
		}

		if nextCollection != nil {
			created, updated, err := s.upsertSyncedEvent(ctx, orgID, UpsertEventParams{
				SourceEntityType: "evidence_collection",
				SourceEntityID:   id,
				SourceEntityRef:  ref,
				EventType:        "evidence_collection_due",
				Category:         "evidence",
				Priority:         "medium",
				Title:            fmt.Sprintf("Evidence collection due: %s", title),
				Description:      fmt.Sprintf("Evidence collection %s (%s) is due", ref, title),
				DueDate:          nextCollection.Format("2006-01-02"),
				AssignedTo:       assignedID,
			})
			if err != nil {
				result.Errors++
			} else if created {
				result.EventsCreated++
			} else if updated {
				result.EventsUpdated++
			} else {
				result.EventsSkipped++
			}
		}
	}

	result.Duration = time.Since(start).String()
	return result, nil
}

// SyncExceptionEvents syncs exception expiry and review deadlines.
func (s *CalendarService) SyncExceptionEvents(ctx context.Context, orgID uuid.UUID) (*CalendarSyncResult, error) {
	start := time.Now()
	result := &CalendarSyncResult{Module: "exceptions"}

	rows, err := s.pool.Query(ctx, `
		SELECT ce.id, ce.exception_ref, ce.title, ce.expiry_date,
			   ce.next_review_date, ce.requested_by, ce.priority
		FROM compliance_exceptions ce
		WHERE ce.organization_id = $1
		  AND ce.deleted_at IS NULL
		  AND ce.status IN ('approved', 'renewal_pending')
		  AND (ce.expiry_date IS NOT NULL OR ce.next_review_date IS NOT NULL)`, orgID)
	if err != nil {
		log.Debug().Err(err).Msg("exceptions query (table may not exist)")
		result.Duration = time.Since(start).String()
		return result, nil
	}
	defer rows.Close()

	for rows.Next() {
		var id uuid.UUID
		var ref, title, priority string
		var expiryDate, reviewDate *time.Time
		var requestedBy uuid.UUID

		if err := rows.Scan(&id, &ref, &title, &expiryDate, &reviewDate, &requestedBy, &priority); err != nil {
			result.Errors++
			continue
		}

		if expiryDate != nil {
			created, updated, err := s.upsertSyncedEvent(ctx, orgID, UpsertEventParams{
				SourceEntityType: "exception",
				SourceEntityID:   id,
				SourceEntityRef:  ref,
				EventType:        "exception_expiry",
				Category:         "exception",
				Priority:         priorityOrDefault(priority, "high"),
				Title:            fmt.Sprintf("Exception expiring: %s", title),
				Description:      fmt.Sprintf("Compliance exception %s (%s) expires", ref, title),
				DueDate:          expiryDate.Format("2006-01-02"),
				AssignedTo:       &requestedBy,
			})
			if err != nil {
				result.Errors++
			} else if created {
				result.EventsCreated++
			} else if updated {
				result.EventsUpdated++
			} else {
				result.EventsSkipped++
			}
		}

		if reviewDate != nil {
			created, updated, err := s.upsertSyncedEvent(ctx, orgID, UpsertEventParams{
				SourceEntityType: "exception",
				SourceEntityID:   id,
				SourceEntityRef:  ref,
				EventType:        "exception_review_due",
				Category:         "exception",
				Priority:         "medium",
				Title:            fmt.Sprintf("Exception review due: %s", title),
				Description:      fmt.Sprintf("Review required for exception %s (%s)", ref, title),
				DueDate:          reviewDate.Format("2006-01-02"),
				AssignedTo:       &requestedBy,
			})
			if err != nil {
				result.Errors++
			} else if created {
				result.EventsCreated++
			} else if updated {
				result.EventsUpdated++
			} else {
				result.EventsSkipped++
			}
		}
	}

	result.Duration = time.Since(start).String()
	return result, nil
}

// SyncDSREvents syncs DSR deadlines to the calendar.
func (s *CalendarService) SyncDSREvents(ctx context.Context, orgID uuid.UUID) (*CalendarSyncResult, error) {
	start := time.Now()
	result := &CalendarSyncResult{Module: "dsr"}

	rows, err := s.pool.Query(ctx, `
		SELECT dr.id, dr.request_ref, dr.subject_name, dr.statutory_deadline,
			   dr.extension_deadline, dr.assigned_to, dr.priority
		FROM dsr_requests dr
		WHERE dr.organization_id = $1
		  AND dr.status NOT IN ('completed', 'rejected', 'withdrawn')
		  AND dr.statutory_deadline IS NOT NULL`, orgID)
	if err != nil {
		log.Debug().Err(err).Msg("dsr_requests query (table may not exist)")
		result.Duration = time.Since(start).String()
		return result, nil
	}
	defer rows.Close()

	for rows.Next() {
		var id uuid.UUID
		var ref, subjectName, priority string
		var deadline, extDeadline *time.Time
		var assignedTo *uuid.UUID

		if err := rows.Scan(&id, &ref, &subjectName, &deadline, &extDeadline, &assignedTo, &priority); err != nil {
			result.Errors++
			continue
		}

		if deadline != nil {
			created, updated, err := s.upsertSyncedEvent(ctx, orgID, UpsertEventParams{
				SourceEntityType: "dsr",
				SourceEntityID:   id,
				SourceEntityRef:  ref,
				EventType:        "dsr_deadline",
				Category:         "dsr",
				Priority:         priorityOrDefault(priority, "high"),
				Title:            fmt.Sprintf("DSR response deadline: %s", ref),
				Description:      fmt.Sprintf("Data subject request %s for %s - statutory deadline", ref, subjectName),
				DueDate:          deadline.Format("2006-01-02"),
				AssignedTo:       assignedTo,
			})
			if err != nil {
				result.Errors++
			} else if created {
				result.EventsCreated++
			} else if updated {
				result.EventsUpdated++
			} else {
				result.EventsSkipped++
			}
		}

		if extDeadline != nil {
			created, updated, err := s.upsertSyncedEvent(ctx, orgID, UpsertEventParams{
				SourceEntityType: "dsr",
				SourceEntityID:   id,
				SourceEntityRef:  ref,
				EventType:        "dsr_extension_deadline",
				Category:         "dsr",
				Priority:         "high",
				Title:            fmt.Sprintf("DSR extension deadline: %s", ref),
				Description:      fmt.Sprintf("Extended deadline for DSR %s", ref),
				DueDate:          extDeadline.Format("2006-01-02"),
				AssignedTo:       assignedTo,
			})
			if err != nil {
				result.Errors++
			} else if created {
				result.EventsCreated++
			} else if updated {
				result.EventsUpdated++
			} else {
				result.EventsSkipped++
			}
		}
	}

	result.Duration = time.Since(start).String()
	return result, nil
}

// SyncIncidentEvents syncs incident notification deadlines.
func (s *CalendarService) SyncIncidentEvents(ctx context.Context, orgID uuid.UUID) (*CalendarSyncResult, error) {
	start := time.Now()
	result := &CalendarSyncResult{Module: "incidents"}

	rows, err := s.pool.Query(ctx, `
		SELECT i.id, i.incident_ref, i.title, i.notification_deadline,
			   i.is_nis2_reportable, i.reported_at, i.assigned_to, i.severity
		FROM incidents i
		WHERE i.organization_id = $1
		  AND i.deleted_at IS NULL
		  AND i.status NOT IN ('resolved', 'closed')
		  AND (i.notification_deadline IS NOT NULL OR i.is_nis2_reportable = true)`, orgID)
	if err != nil {
		return nil, fmt.Errorf("query incidents: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id uuid.UUID
		var ref, title, severity string
		var notifDeadline, reportedAt *time.Time
		var isNIS2 bool
		var assignedTo *uuid.UUID

		if err := rows.Scan(&id, &ref, &title, &notifDeadline, &isNIS2, &reportedAt, &assignedTo, &severity); err != nil {
			result.Errors++
			continue
		}

		if notifDeadline != nil {
			created, updated, err := s.upsertSyncedEvent(ctx, orgID, UpsertEventParams{
				SourceEntityType: "incident",
				SourceEntityID:   id,
				SourceEntityRef:  ref,
				EventType:        "incident_notification_deadline",
				Category:         "incident",
				Priority:         "critical",
				Title:            fmt.Sprintf("Incident notification deadline: %s", title),
				Description:      fmt.Sprintf("GDPR 72h notification deadline for incident %s (%s)", ref, title),
				DueDate:          notifDeadline.Format("2006-01-02"),
				AssignedTo:       assignedTo,
			})
			if err != nil {
				result.Errors++
			} else if created {
				result.EventsCreated++
			} else if updated {
				result.EventsUpdated++
			} else {
				result.EventsSkipped++
			}
		}

		if isNIS2 && reportedAt != nil {
			earlyWarning := reportedAt.Add(24 * time.Hour)
			created, updated, err := s.upsertSyncedEvent(ctx, orgID, UpsertEventParams{
				SourceEntityType: "incident",
				SourceEntityID:   id,
				SourceEntityRef:  ref,
				EventType:        "incident_nis2_early_warning",
				Category:         "incident",
				Priority:         "critical",
				Title:            fmt.Sprintf("NIS2 early warning deadline: %s", title),
				Description:      fmt.Sprintf("NIS2 24h early warning deadline for incident %s", ref),
				DueDate:          earlyWarning.Format("2006-01-02"),
				AssignedTo:       assignedTo,
			})
			if err != nil {
				result.Errors++
			} else if created {
				result.EventsCreated++
			} else if updated {
				result.EventsUpdated++
			} else {
				result.EventsSkipped++
			}
		}
	}

	result.Duration = time.Since(start).String()
	return result, nil
}

// SyncRegulatoryEvents syncs regulatory change deadlines.
func (s *CalendarService) SyncRegulatoryEvents(ctx context.Context, orgID uuid.UUID) (*CalendarSyncResult, error) {
	start := time.Now()
	result := &CalendarSyncResult{Module: "regulatory"}

	rows, err := s.pool.Query(ctx, `
		SELECT rc.id, rc.change_ref, rc.title, rc.effective_date, rc.deadline,
			   rc.assigned_to, rc.severity
		FROM regulatory_changes rc
		JOIN regulatory_subscriptions rs ON rs.source_id = rc.source_id AND rs.organization_id = $1
		WHERE rc.status NOT IN ('implemented', 'not_applicable')
		  AND (rc.effective_date IS NOT NULL OR rc.deadline IS NOT NULL)`, orgID)
	if err != nil {
		log.Debug().Err(err).Msg("regulatory_changes query (table may not exist)")
		result.Duration = time.Since(start).String()
		return result, nil
	}
	defer rows.Close()

	for rows.Next() {
		var id uuid.UUID
		var ref, title, severity string
		var effectiveDate, deadline *time.Time
		var assignedTo *uuid.UUID

		if err := rows.Scan(&id, &ref, &title, &effectiveDate, &deadline, &assignedTo, &severity); err != nil {
			result.Errors++
			continue
		}

		if effectiveDate != nil {
			created, updated, err := s.upsertSyncedEvent(ctx, orgID, UpsertEventParams{
				SourceEntityType: "regulatory_change",
				SourceEntityID:   id,
				SourceEntityRef:  ref,
				EventType:        "regulatory_effective_date",
				Category:         "regulatory",
				Priority:         severityToPriority(severity),
				Title:            fmt.Sprintf("Regulation effective: %s", title),
				Description:      fmt.Sprintf("Regulatory change %s (%s) takes effect", ref, title),
				DueDate:          effectiveDate.Format("2006-01-02"),
				AssignedTo:       assignedTo,
			})
			if err != nil {
				result.Errors++
			} else if created {
				result.EventsCreated++
			} else if updated {
				result.EventsUpdated++
			} else {
				result.EventsSkipped++
			}
		}

		if deadline != nil {
			created, updated, err := s.upsertSyncedEvent(ctx, orgID, UpsertEventParams{
				SourceEntityType: "regulatory_change",
				SourceEntityID:   id,
				SourceEntityRef:  ref,
				EventType:        "regulatory_response_deadline",
				Category:         "regulatory",
				Priority:         severityToPriority(severity),
				Title:            fmt.Sprintf("Regulatory response deadline: %s", title),
				Description:      fmt.Sprintf("Response deadline for regulatory change %s (%s)", ref, title),
				DueDate:          deadline.Format("2006-01-02"),
				AssignedTo:       assignedTo,
			})
			if err != nil {
				result.Errors++
			} else if created {
				result.EventsCreated++
			} else if updated {
				result.EventsUpdated++
			} else {
				result.EventsSkipped++
			}
		}
	}

	result.Duration = time.Since(start).String()
	return result, nil
}

// SyncBCEvents syncs business continuity plan reviews and exercises.
func (s *CalendarService) SyncBCEvents(ctx context.Context, orgID uuid.UUID) (*CalendarSyncResult, error) {
	start := time.Now()
	result := &CalendarSyncResult{Module: "business_continuity"}

	// BC plan reviews
	rows, err := s.pool.Query(ctx, `
		SELECT bp.id, bp.plan_ref, bp.title, bp.next_review_date,
			   bp.owner_user_id
		FROM bc_plans bp
		WHERE bp.organization_id = $1
		  AND bp.status NOT IN ('retired', 'archived')
		  AND bp.next_review_date IS NOT NULL`, orgID)
	if err != nil {
		log.Debug().Err(err).Msg("bc_plans query (table may not exist)")
		result.Duration = time.Since(start).String()
		return result, nil
	}
	defer rows.Close()

	for rows.Next() {
		var id uuid.UUID
		var ref, title string
		var nextReview *time.Time
		var ownerID *uuid.UUID

		if err := rows.Scan(&id, &ref, &title, &nextReview, &ownerID); err != nil {
			result.Errors++
			continue
		}

		if nextReview != nil {
			created, updated, err := s.upsertSyncedEvent(ctx, orgID, UpsertEventParams{
				SourceEntityType: "bc_plan",
				SourceEntityID:   id,
				SourceEntityRef:  ref,
				EventType:        "bc_plan_review_due",
				Category:         "business_continuity",
				Priority:         "medium",
				Title:            fmt.Sprintf("BC plan review due: %s", title),
				Description:      fmt.Sprintf("Business continuity plan %s (%s) is due for review", ref, title),
				DueDate:          nextReview.Format("2006-01-02"),
				AssignedTo:       ownerID,
				RecurrenceType:   "annually",
			})
			if err != nil {
				result.Errors++
			} else if created {
				result.EventsCreated++
			} else if updated {
				result.EventsUpdated++
			} else {
				result.EventsSkipped++
			}
		}
	}

	// BC exercises
	eRows, err := s.pool.Query(ctx, `
		SELECT be.id, be.exercise_ref, be.title, be.scheduled_date,
			   be.lead_user_id
		FROM bc_exercises be
		WHERE be.organization_id = $1
		  AND be.status NOT IN ('completed', 'cancelled')
		  AND be.scheduled_date IS NOT NULL`, orgID)
	if err != nil {
		log.Debug().Err(err).Msg("bc_exercises query (table may not exist)")
		result.Duration = time.Since(start).String()
		return result, nil
	}
	defer eRows.Close()

	for eRows.Next() {
		var id uuid.UUID
		var ref, title string
		var scheduledDate *time.Time
		var leadID *uuid.UUID

		if err := eRows.Scan(&id, &ref, &title, &scheduledDate, &leadID); err != nil {
			result.Errors++
			continue
		}

		if scheduledDate != nil {
			created, updated, err := s.upsertSyncedEvent(ctx, orgID, UpsertEventParams{
				SourceEntityType: "bc_exercise",
				SourceEntityID:   id,
				SourceEntityRef:  ref,
				EventType:        "bc_exercise_scheduled",
				Category:         "business_continuity",
				Priority:         "high",
				Title:            fmt.Sprintf("BC exercise: %s", title),
				Description:      fmt.Sprintf("Business continuity exercise %s (%s)", ref, title),
				DueDate:          scheduledDate.Format("2006-01-02"),
				AssignedTo:       leadID,
			})
			if err != nil {
				result.Errors++
			} else if created {
				result.EventsCreated++
			} else if updated {
				result.EventsUpdated++
			} else {
				result.EventsSkipped++
			}
		}
	}

	result.Duration = time.Since(start).String()
	return result, nil
}

// SyncBoardEvents syncs board meetings and decision action deadlines.
func (s *CalendarService) SyncBoardEvents(ctx context.Context, orgID uuid.UUID) (*CalendarSyncResult, error) {
	start := time.Now()
	result := &CalendarSyncResult{Module: "board"}

	// Board meetings
	rows, err := s.pool.Query(ctx, `
		SELECT bm.id, bm.meeting_ref, bm.title, bm.date, bm.status
		FROM board_meetings bm
		WHERE bm.organization_id = $1
		  AND bm.status NOT IN ('completed', 'minutes_approved')`, orgID)
	if err != nil {
		log.Debug().Err(err).Msg("board_meetings query (table may not exist)")
		result.Duration = time.Since(start).String()
		return result, nil
	}
	defer rows.Close()

	for rows.Next() {
		var id uuid.UUID
		var ref, title, status string
		var meetingDate time.Time

		if err := rows.Scan(&id, &ref, &title, &meetingDate, &status); err != nil {
			result.Errors++
			continue
		}

		created, updated, err := s.upsertSyncedEvent(ctx, orgID, UpsertEventParams{
			SourceEntityType: "board_meeting",
			SourceEntityID:   id,
			SourceEntityRef:  ref,
			EventType:        "board_meeting_scheduled",
			Category:         "board",
			Priority:         "high",
			Title:            fmt.Sprintf("Board meeting: %s", title),
			Description:      fmt.Sprintf("Board meeting %s (%s) - Status: %s", ref, title, status),
			DueDate:          meetingDate.Format("2006-01-02"),
		})
		if err != nil {
			result.Errors++
		} else if created {
			result.EventsCreated++
		} else if updated {
			result.EventsUpdated++
		} else {
			result.EventsSkipped++
		}
	}

	// Board decision actions
	dRows, err := s.pool.Query(ctx, `
		SELECT bd.id, bd.decision_ref, bd.title, bd.action_due_date,
			   bd.action_owner_user_id, bd.action_status
		FROM board_decisions bd
		WHERE bd.organization_id = $1
		  AND bd.action_required = true
		  AND bd.action_status NOT IN ('completed', 'cancelled')
		  AND bd.action_due_date IS NOT NULL`, orgID)
	if err != nil {
		log.Debug().Err(err).Msg("board_decisions query")
		result.Duration = time.Since(start).String()
		return result, nil
	}
	defer dRows.Close()

	for dRows.Next() {
		var id uuid.UUID
		var ref, title, actionStatus string
		var dueDate time.Time
		var ownerID *uuid.UUID

		if err := dRows.Scan(&id, &ref, &title, &dueDate, &ownerID, &actionStatus); err != nil {
			result.Errors++
			continue
		}

		created, updated, err := s.upsertSyncedEvent(ctx, orgID, UpsertEventParams{
			SourceEntityType: "board_decision",
			SourceEntityID:   id,
			SourceEntityRef:  ref,
			EventType:        "board_decision_action_due",
			Category:         "board",
			Priority:         "high",
			Title:            fmt.Sprintf("Board action due: %s", title),
			Description:      fmt.Sprintf("Action required for board decision %s - Status: %s", ref, actionStatus),
			DueDate:          dueDate.Format("2006-01-02"),
			AssignedTo:       ownerID,
		})
		if err != nil {
			result.Errors++
		} else if created {
			result.EventsCreated++
		} else if updated {
			result.EventsUpdated++
		} else {
			result.EventsSkipped++
		}
	}

	result.Duration = time.Since(start).String()
	return result, nil
}

// ============================================================
// IDEMPOTENT UPSERT ENGINE
// ============================================================

// UpsertEventParams holds parameters for the idempotent upsert.
type UpsertEventParams struct {
	SourceEntityType string
	SourceEntityID   uuid.UUID
	SourceEntityRef  string
	EventType        string
	Category         string
	Priority         string
	Title            string
	Description      string
	DueDate          string
	AssignedTo       *uuid.UUID
	RecurrenceType   string
}

// upsertSyncedEvent creates or updates a calendar event based on source entity.
// Returns (created, updated, error).
func (s *CalendarService) upsertSyncedEvent(ctx context.Context, orgID uuid.UUID, p UpsertEventParams) (bool, bool, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return false, false, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return false, false, err
	}

	recurrence := p.RecurrenceType
	if recurrence == "" {
		recurrence = "none"
	}

	// Check for existing event with same source+type
	var existingID uuid.UUID
	var existingDueDate, existingTitle, existingPriority, existingStatus string
	err = tx.QueryRow(ctx, `
		SELECT id, due_date::TEXT, title, priority::TEXT, status::TEXT
		FROM calendar_events
		WHERE organization_id = $1
		  AND source_entity_type = $2
		  AND source_entity_id = $3
		  AND event_type = $4::calendar_event_type
		  AND COALESCE(occurrence_date, '1970-01-01'::DATE) = '1970-01-01'::DATE
		LIMIT 1`,
		orgID, p.SourceEntityType, p.SourceEntityID, p.EventType,
	).Scan(&existingID, &existingDueDate, &existingTitle, &existingPriority, &existingStatus)

	if err == nil {
		// Event exists — check if update is needed
		if existingDueDate == p.DueDate && existingTitle == p.Title && existingPriority == p.Priority {
			tx.Commit(ctx)
			return false, false, nil // no change
		}

		// Do not update completed/cancelled events
		if existingStatus == "completed" || existingStatus == "cancelled" {
			tx.Commit(ctx)
			return false, false, nil
		}

		_, err := tx.Exec(ctx, `
			UPDATE calendar_events
			SET due_date = $2::DATE, title = $3, description = $4,
				priority = $5::calendar_event_priority,
				source_entity_ref = $6,
				assigned_to_user_id = $7,
				status = CASE
					WHEN $2::DATE < CURRENT_DATE THEN 'overdue'::calendar_event_status
					WHEN $2::DATE <= CURRENT_DATE + INTERVAL '3 days' THEN 'due_soon'::calendar_event_status
					ELSE 'upcoming'::calendar_event_status
				END
			WHERE id = $1`,
			existingID, p.DueDate, p.Title, p.Description, p.Priority, p.SourceEntityRef, p.AssignedTo)
		if err != nil {
			return false, false, fmt.Errorf("update event: %w", err)
		}

		if err := tx.Commit(ctx); err != nil {
			return false, false, err
		}
		return false, true, nil
	}

	if err != pgx.ErrNoRows {
		return false, false, fmt.Errorf("check existing: %w", err)
	}

	// Create new event
	_, err = tx.Exec(ctx, `
		INSERT INTO calendar_events (
			organization_id, event_type, category, priority,
			status, title, description,
			source_entity_type, source_entity_id, source_entity_ref,
			due_date, all_day, recurrence_type,
			assigned_to_user_id,
			reminder_days_before
		) VALUES (
			$1, $2::calendar_event_type, $3::calendar_event_category, $4::calendar_event_priority,
			CASE
				WHEN $8::DATE < CURRENT_DATE THEN 'overdue'::calendar_event_status
				WHEN $8::DATE <= CURRENT_DATE + INTERVAL '3 days' THEN 'due_soon'::calendar_event_status
				ELSE 'upcoming'::calendar_event_status
			END,
			$5, $6,
			$7, $9, $10,
			$8::DATE, true, $11::calendar_recurrence_type,
			$12,
			'{7,3,1}'
		)`,
		orgID, p.EventType, p.Category, p.Priority,
		p.Title, p.Description,
		p.SourceEntityType, p.DueDate, p.SourceEntityID, p.SourceEntityRef,
		recurrence, p.AssignedTo)
	if err != nil {
		return false, false, fmt.Errorf("insert event: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return false, false, err
	}
	return true, false, nil
}

// ============================================================
// SYNC CONFIG HELPERS
// ============================================================

// isModuleEnabled checks if a module sync is enabled for an org.
func (s *CalendarService) isModuleEnabled(ctx context.Context, orgID uuid.UUID, module string) (bool, error) {
	var enabled bool
	err := s.pool.QueryRow(ctx, `
		SELECT is_enabled FROM calendar_sync_configs
		WHERE organization_id = $1 AND module_name = $2`, orgID, module,
	).Scan(&enabled)
	if err == pgx.ErrNoRows {
		return true, nil // default: enabled
	}
	return enabled, err
}

// updateSyncStatus records the result of a module sync.
func (s *CalendarService) updateSyncStatus(ctx context.Context, orgID uuid.UUID, module, status string, created, updated int, errMsg string) {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO calendar_sync_configs (organization_id, module_name, last_sync_at, last_sync_status, last_sync_events_created, last_sync_events_updated, last_sync_error)
		VALUES ($1, $2, NOW(), $3, $4, $5, $6)
		ON CONFLICT (organization_id, module_name) DO UPDATE SET
			last_sync_at = NOW(),
			last_sync_status = EXCLUDED.last_sync_status,
			last_sync_events_created = EXCLUDED.last_sync_events_created,
			last_sync_events_updated = EXCLUDED.last_sync_events_updated,
			last_sync_error = EXCLUDED.last_sync_error`,
		orgID, module, status, created, updated, errMsg)
	if err != nil {
		log.Error().Err(err).Str("module", module).Msg("Failed to update sync status")
	}
}

// GetSyncStatus returns the current sync status for all modules.
func (s *CalendarService) GetSyncStatus(ctx context.Context, orgID uuid.UUID) (*SyncStatus, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return nil, err
	}

	rows, err := tx.Query(ctx, `
		SELECT id, organization_id, module_name, is_enabled, sync_frequency_minutes,
			   last_sync_at, last_sync_status, last_sync_events_created, last_sync_events_updated,
			   COALESCE(last_sync_error, ''), auto_create_events, auto_complete_events,
			   default_reminder_days, default_priority::TEXT, created_at, updated_at
		FROM calendar_sync_configs
		WHERE organization_id = $1
		ORDER BY module_name`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	status := &SyncStatus{}
	for rows.Next() {
		var cfg CalendarSyncConfig
		if err := rows.Scan(
			&cfg.ID, &cfg.OrganizationID, &cfg.ModuleName, &cfg.IsEnabled,
			&cfg.SyncFrequencyMinutes, &cfg.LastSyncAt, &cfg.LastSyncStatus,
			&cfg.LastSyncEventsCreated, &cfg.LastSyncEventsUpdated, &cfg.LastSyncError,
			&cfg.AutoCreateEvents, &cfg.AutoCompleteEvents, &cfg.DefaultReminderDays,
			&cfg.DefaultPriority, &cfg.CreatedAt, &cfg.UpdatedAt,
		); err != nil {
			continue
		}
		status.Modules = append(status.Modules, cfg)
		if cfg.LastSyncAt != nil {
			if status.LastFullSync == nil || cfg.LastSyncAt.After(*status.LastFullSync) {
				status.LastFullSync = cfg.LastSyncAt
			}
		}
	}

	// Total and overdue event counts
	err = tx.QueryRow(ctx, `
		SELECT COUNT(*), COUNT(*) FILTER (WHERE status = 'overdue')
		FROM calendar_events
		WHERE organization_id = $1`, orgID,
	).Scan(&status.TotalEvents, &status.OverdueEvents)
	if err != nil {
		return nil, err
	}

	tx.Commit(ctx)
	return status, nil
}

// UpdateSyncConfigs updates sync configurations for an org.
func (s *CalendarService) UpdateSyncConfigs(ctx context.Context, orgID uuid.UUID, req UpdateSyncConfigRequest) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return err
	}

	for _, c := range req.Configs {
		if c.ModuleName == "" {
			continue
		}
		priority := c.DefaultPriority
		if priority == "" {
			priority = "medium"
		}

		_, err := tx.Exec(ctx, `
			INSERT INTO calendar_sync_configs (organization_id, module_name, is_enabled, sync_frequency_minutes, auto_create_events, auto_complete_events, default_priority)
			VALUES ($1, $2, COALESCE($3, true), COALESCE($4, 30), COALESCE($5, true), COALESCE($6, true), $7::calendar_event_priority)
			ON CONFLICT (organization_id, module_name) DO UPDATE SET
				is_enabled = COALESCE(EXCLUDED.is_enabled, calendar_sync_configs.is_enabled),
				sync_frequency_minutes = COALESCE(EXCLUDED.sync_frequency_minutes, calendar_sync_configs.sync_frequency_minutes),
				auto_create_events = COALESCE(EXCLUDED.auto_create_events, calendar_sync_configs.auto_create_events),
				auto_complete_events = COALESCE(EXCLUDED.auto_complete_events, calendar_sync_configs.auto_complete_events),
				default_priority = COALESCE(EXCLUDED.default_priority, calendar_sync_configs.default_priority)`,
			orgID, c.ModuleName, c.IsEnabled, c.SyncFrequencyMinutes, c.AutoCreateEvents, c.AutoCompleteEvents, priority)
		if err != nil {
			return fmt.Errorf("update config %s: %w", c.ModuleName, err)
		}
	}

	return tx.Commit(ctx)
}

// ============================================================
// SCAN HELPER
// ============================================================

func (s *CalendarService) scanEvent(row pgx.Row) (*CalendarEvent, error) {
	var e CalendarEvent
	err := row.Scan(
		&e.ID, &e.OrganizationID, &e.EventRef, &e.EventType, &e.Category,
		&e.Priority, &e.Status, &e.Title, &e.Description,
		&e.SourceEntityType, &e.SourceEntityID, &e.SourceEntityRef,
		&e.DueDate, &e.DueTime, &e.StartDate, &e.EndDate,
		&e.AllDay, &e.Timezone, &e.RecurrenceType, &e.Rrule,
		&e.RecurrenceEndDate, &e.ParentEventID, &e.OccurrenceDate,
		&e.AssignedToUserID, &e.AssignedToRole, &e.OwnerUserID,
		&e.ReminderDaysBefore, &e.LastReminderSentAt, &e.ReminderCount,
		&e.EscalationAfterDays, &e.Escalated, &e.EscalatedToUserID, &e.EscalatedAt,
		&e.CompletedAt, &e.CompletedBy, &e.CompletionNotes,
		&e.OriginalDueDate, &e.RescheduledCount, &e.RescheduleReason,
		&e.URL, &e.Color, &e.Tags, &e.Metadata,
		&e.CreatedAt, &e.UpdatedAt,
		&e.AssignedToName, &e.OwnerName,
	)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// ============================================================
// UTILITY HELPERS
// ============================================================

func generateICalToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func riskLevelToPriority(level string) string {
	switch strings.ToLower(level) {
	case "critical":
		return "critical"
	case "high":
		return "high"
	case "medium":
		return "medium"
	default:
		return "low"
	}
}

func severityToPriority(severity string) string {
	switch strings.ToLower(severity) {
	case "critical":
		return "critical"
	case "high":
		return "high"
	case "medium":
		return "medium"
	default:
		return "low"
	}
}

func priorityOrDefault(priority, def string) string {
	switch strings.ToLower(priority) {
	case "critical", "high", "medium", "low":
		return strings.ToLower(priority)
	default:
		return def
	}
}

// ============================================================
// RRULE / RECURRENCE PARSING
// ============================================================

// ComputeNextOccurrence calculates the next occurrence date based on recurrence.
// Supports both simple recurrence types and RFC 5545 RRULE strings.
func ComputeNextOccurrence(currentDate time.Time, recType, rrule string) time.Time {
	return computeNextOccurrence(currentDate, recType, rrule)
}

func computeNextOccurrence(currentDate time.Time, recType, rrule string) time.Time {
	switch recType {
	case "daily":
		return currentDate.AddDate(0, 0, 1)
	case "weekly":
		return currentDate.AddDate(0, 0, 7)
	case "monthly":
		return currentDate.AddDate(0, 1, 0)
	case "quarterly":
		return currentDate.AddDate(0, 3, 0)
	case "semi_annually":
		return currentDate.AddDate(0, 6, 0)
	case "annually":
		return currentDate.AddDate(1, 0, 0)
	case "custom_rrule":
		return parseRRule(currentDate, rrule)
	}
	return time.Time{}
}

// ParseRRule parses a simplified RFC 5545 RRULE string and computes the next date.
// Supports: FREQ=DAILY|WEEKLY|MONTHLY|YEARLY and INTERVAL=N
// Example: "FREQ=MONTHLY;INTERVAL=3" produces quarterly recurrence.
func ParseRRule(currentDate time.Time, rrule string) time.Time {
	return parseRRule(currentDate, rrule)
}

func parseRRule(currentDate time.Time, rrule string) time.Time {
	if rrule == "" {
		return time.Time{}
	}

	parts := strings.Split(rrule, ";")
	freq := ""
	interval := 1

	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.TrimSpace(strings.ToUpper(kv[0]))
		val := strings.TrimSpace(strings.ToUpper(kv[1]))

		switch key {
		case "FREQ":
			freq = val
		case "INTERVAL":
			if n, err := strconv.Atoi(val); err == nil && n > 0 {
				interval = n
			}
		}
	}

	switch freq {
	case "DAILY":
		return currentDate.AddDate(0, 0, interval)
	case "WEEKLY":
		return currentDate.AddDate(0, 0, 7*interval)
	case "MONTHLY":
		return currentDate.AddDate(0, interval, 0)
	case "YEARLY":
		return currentDate.AddDate(interval, 0, 0)
	}

	return time.Time{}
}
