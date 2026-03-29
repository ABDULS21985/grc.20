package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ============================================================
// QUERY TYPES
// ============================================================

// CalendarViewFilter holds filter options for the calendar view.
type CalendarViewFilter struct {
	StartDate  string   `json:"start_date"`
	EndDate    string   `json:"end_date"`
	Categories []string `json:"categories"`
	Priorities []string `json:"priorities"`
	Statuses   []string `json:"statuses"`
	EventTypes []string `json:"event_types"`
	AssignedTo *uuid.UUID `json:"assigned_to"`
	Search     string   `json:"search"`
	Page       int      `json:"page"`
	PageSize   int      `json:"page_size"`
}

// CalendarViewResult is the paginated result for the calendar view.
type CalendarViewResult struct {
	Events     []CalendarEvent `json:"events"`
	Total      int64           `json:"total"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	TotalPages int             `json:"total_pages"`
}

// UpcomingDeadline represents a deadline in the upcoming deadlines list.
type UpcomingDeadline struct {
	ID               uuid.UUID  `json:"id"`
	EventRef         string     `json:"event_ref"`
	EventType        string     `json:"event_type"`
	Category         string     `json:"category"`
	Priority         string     `json:"priority"`
	Status           string     `json:"status"`
	Title            string     `json:"title"`
	DueDate          string     `json:"due_date"`
	DaysLeft         int        `json:"days_left"`
	SourceEntityType string     `json:"source_entity_type"`
	SourceEntityID   uuid.UUID  `json:"source_entity_id"`
	SourceEntityRef  string     `json:"source_entity_ref"`
	AssignedToUserID *uuid.UUID `json:"assigned_to_user_id"`
	AssignedToName   string     `json:"assigned_to_name"`
}

// OverdueItem represents an overdue calendar event.
type OverdueItem struct {
	ID               uuid.UUID  `json:"id"`
	EventRef         string     `json:"event_ref"`
	EventType        string     `json:"event_type"`
	Category         string     `json:"category"`
	Priority         string     `json:"priority"`
	Title            string     `json:"title"`
	DueDate          string     `json:"due_date"`
	DaysOverdue      int        `json:"days_overdue"`
	SourceEntityType string     `json:"source_entity_type"`
	SourceEntityID   uuid.UUID  `json:"source_entity_id"`
	SourceEntityRef  string     `json:"source_entity_ref"`
	AssignedToUserID *uuid.UUID `json:"assigned_to_user_id"`
	AssignedToName   string     `json:"assigned_to_name"`
	Escalated        bool       `json:"escalated"`
}

// CalendarSummary holds monthly compliance calendar statistics.
type CalendarSummary struct {
	Month                string                      `json:"month"`
	TotalEvents          int64                       `json:"total_events"`
	UpcomingCount        int64                       `json:"upcoming_count"`
	DueSoonCount         int64                       `json:"due_soon_count"`
	OverdueCount         int64                       `json:"overdue_count"`
	CompletedCount       int64                       `json:"completed_count"`
	CancelledCount       int64                       `json:"cancelled_count"`
	ByCategory           map[string]int64            `json:"by_category"`
	ByPriority           map[string]int64            `json:"by_priority"`
	CompletionRate       float64                     `json:"completion_rate"`
	CriticalDeadlines    []UpcomingDeadline          `json:"critical_deadlines"`
	WeeklySummary        []WeekSummary               `json:"weekly_summary"`
}

// WeekSummary holds per-week event counts.
type WeekSummary struct {
	WeekStart string `json:"week_start"`
	WeekEnd   string `json:"week_end"`
	Total     int64  `json:"total"`
	Completed int64  `json:"completed"`
	Overdue   int64  `json:"overdue"`
}

// ============================================================
// CALENDAR VIEW QUERY
// ============================================================

// GetCalendarView returns a filtered view of calendar events.
func (s *CalendarService) GetCalendarView(ctx context.Context, orgID, userID uuid.UUID, filter CalendarViewFilter) (*CalendarViewResult, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return nil, err
	}

	// Build dynamic WHERE clause
	where := []string{"ce.organization_id = $1"}
	args := []interface{}{orgID}
	argIdx := 2

	if filter.StartDate != "" {
		where = append(where, fmt.Sprintf("ce.due_date >= $%d::DATE", argIdx))
		args = append(args, filter.StartDate)
		argIdx++
	}
	if filter.EndDate != "" {
		where = append(where, fmt.Sprintf("ce.due_date <= $%d::DATE", argIdx))
		args = append(args, filter.EndDate)
		argIdx++
	}
	if len(filter.Categories) > 0 {
		where = append(where, fmt.Sprintf("ce.category::TEXT = ANY($%d)", argIdx))
		args = append(args, filter.Categories)
		argIdx++
	}
	if len(filter.Priorities) > 0 {
		where = append(where, fmt.Sprintf("ce.priority::TEXT = ANY($%d)", argIdx))
		args = append(args, filter.Priorities)
		argIdx++
	}
	if len(filter.Statuses) > 0 {
		where = append(where, fmt.Sprintf("ce.status::TEXT = ANY($%d)", argIdx))
		args = append(args, filter.Statuses)
		argIdx++
	}
	if len(filter.EventTypes) > 0 {
		where = append(where, fmt.Sprintf("ce.event_type::TEXT = ANY($%d)", argIdx))
		args = append(args, filter.EventTypes)
		argIdx++
	}
	if filter.AssignedTo != nil {
		where = append(where, fmt.Sprintf("ce.assigned_to_user_id = $%d", argIdx))
		args = append(args, *filter.AssignedTo)
		argIdx++
	}
	if filter.Search != "" {
		where = append(where, fmt.Sprintf("(ce.title ILIKE '%%' || $%d || '%%' OR ce.description ILIKE '%%' || $%d || '%%')", argIdx, argIdx))
		args = append(args, filter.Search)
		argIdx++
	}

	whereClause := strings.Join(where, " AND ")

	// Count
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM calendar_events ce WHERE %s", whereClause)
	if err := tx.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count events: %w", err)
	}

	// Paginate
	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}
	offset := (page - 1) * pageSize

	query := fmt.Sprintf(`
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
		WHERE %s
		ORDER BY ce.due_date ASC, ce.priority DESC
		LIMIT $%d OFFSET $%d`,
		whereClause, argIdx, argIdx+1)

	args = append(args, pageSize, offset)

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query events: %w", err)
	}
	defer rows.Close()

	var events []CalendarEvent
	for rows.Next() {
		e, err := s.scanEventFromRows(rows)
		if err != nil {
			continue
		}
		events = append(events, *e)
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		totalPages++
	}

	tx.Commit(ctx)
	return &CalendarViewResult{
		Events:     events,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// ============================================================
// UPCOMING DEADLINES
// ============================================================

// GetUpcomingDeadlines returns events due within the specified number of days.
func (s *CalendarService) GetUpcomingDeadlines(ctx context.Context, orgID, userID uuid.UUID, withinDays, limit int) ([]UpcomingDeadline, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return nil, err
	}

	if withinDays <= 0 {
		withinDays = 30
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	rows, err := tx.Query(ctx, `
		SELECT ce.id, ce.event_ref, ce.event_type, ce.category, ce.priority, ce.status,
			   ce.title, ce.due_date::TEXT,
			   (ce.due_date - CURRENT_DATE)::INT AS days_left,
			   ce.source_entity_type, ce.source_entity_id, ce.source_entity_ref,
			   ce.assigned_to_user_id,
			   COALESCE(u.first_name || ' ' || u.last_name, '')
		FROM calendar_events ce
		LEFT JOIN users u ON u.id = ce.assigned_to_user_id
		WHERE ce.organization_id = $1
		  AND ce.status NOT IN ('completed', 'cancelled')
		  AND ce.due_date >= CURRENT_DATE
		  AND ce.due_date <= CURRENT_DATE + ($2 || ' days')::INTERVAL
		ORDER BY ce.due_date ASC, ce.priority DESC
		LIMIT $3`, orgID, withinDays, limit)
	if err != nil {
		return nil, fmt.Errorf("query deadlines: %w", err)
	}
	defer rows.Close()

	var deadlines []UpcomingDeadline
	for rows.Next() {
		var d UpcomingDeadline
		if err := rows.Scan(
			&d.ID, &d.EventRef, &d.EventType, &d.Category, &d.Priority, &d.Status,
			&d.Title, &d.DueDate,
			&d.DaysLeft,
			&d.SourceEntityType, &d.SourceEntityID, &d.SourceEntityRef,
			&d.AssignedToUserID, &d.AssignedToName,
		); err != nil {
			continue
		}
		deadlines = append(deadlines, d)
	}

	tx.Commit(ctx)
	return deadlines, nil
}

// ============================================================
// OVERDUE ITEMS
// ============================================================

// GetOverdueItems returns all overdue calendar events for an org.
func (s *CalendarService) GetOverdueItems(ctx context.Context, orgID uuid.UUID) ([]OverdueItem, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return nil, err
	}

	rows, err := tx.Query(ctx, `
		SELECT ce.id, ce.event_ref, ce.event_type, ce.category, ce.priority,
			   ce.title, ce.due_date::TEXT,
			   (CURRENT_DATE - ce.due_date)::INT AS days_overdue,
			   ce.source_entity_type, ce.source_entity_id, ce.source_entity_ref,
			   ce.assigned_to_user_id,
			   COALESCE(u.first_name || ' ' || u.last_name, ''),
			   ce.escalated
		FROM calendar_events ce
		LEFT JOIN users u ON u.id = ce.assigned_to_user_id
		WHERE ce.organization_id = $1
		  AND ce.status NOT IN ('completed', 'cancelled')
		  AND ce.due_date < CURRENT_DATE
		ORDER BY ce.priority DESC, ce.due_date ASC`, orgID)
	if err != nil {
		return nil, fmt.Errorf("query overdue: %w", err)
	}
	defer rows.Close()

	var items []OverdueItem
	for rows.Next() {
		var item OverdueItem
		if err := rows.Scan(
			&item.ID, &item.EventRef, &item.EventType, &item.Category, &item.Priority,
			&item.Title, &item.DueDate,
			&item.DaysOverdue,
			&item.SourceEntityType, &item.SourceEntityID, &item.SourceEntityRef,
			&item.AssignedToUserID, &item.AssignedToName,
			&item.Escalated,
		); err != nil {
			continue
		}
		items = append(items, item)
	}

	tx.Commit(ctx)
	return items, nil
}

// ============================================================
// COMPLIANCE CALENDAR SUMMARY
// ============================================================

// GetComplianceCalendarSummary returns aggregated metrics for a given month.
func (s *CalendarService) GetComplianceCalendarSummary(ctx context.Context, orgID uuid.UUID, month string) (*CalendarSummary, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return nil, err
	}

	// Parse the month (YYYY-MM)
	monthStart, err := time.Parse("2006-01", month)
	if err != nil {
		return nil, fmt.Errorf("invalid month format, use YYYY-MM: %w", err)
	}
	monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Nanosecond)

	summary := &CalendarSummary{
		Month:      month,
		ByCategory: make(map[string]int64),
		ByPriority: make(map[string]int64),
	}

	// Totals by status
	err = tx.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE status = 'upcoming'),
			COUNT(*) FILTER (WHERE status = 'due_soon'),
			COUNT(*) FILTER (WHERE status = 'overdue'),
			COUNT(*) FILTER (WHERE status = 'completed'),
			COUNT(*) FILTER (WHERE status = 'cancelled')
		FROM calendar_events
		WHERE organization_id = $1
		  AND due_date >= $2::DATE
		  AND due_date <= $3::DATE`,
		orgID, monthStart.Format("2006-01-02"), monthEnd.Format("2006-01-02"),
	).Scan(
		&summary.TotalEvents, &summary.UpcomingCount, &summary.DueSoonCount,
		&summary.OverdueCount, &summary.CompletedCount, &summary.CancelledCount,
	)
	if err != nil {
		return nil, fmt.Errorf("query monthly totals: %w", err)
	}

	// Completion rate
	if summary.TotalEvents > 0 {
		summary.CompletionRate = float64(summary.CompletedCount) / float64(summary.TotalEvents) * 100
	}

	// By category
	catRows, err := tx.Query(ctx, `
		SELECT category::TEXT, COUNT(*)
		FROM calendar_events
		WHERE organization_id = $1
		  AND due_date >= $2::DATE AND due_date <= $3::DATE
		GROUP BY category`,
		orgID, monthStart.Format("2006-01-02"), monthEnd.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer catRows.Close()

	for catRows.Next() {
		var cat string
		var count int64
		if err := catRows.Scan(&cat, &count); err != nil {
			continue
		}
		summary.ByCategory[cat] = count
	}

	// By priority
	priRows, err := tx.Query(ctx, `
		SELECT priority::TEXT, COUNT(*)
		FROM calendar_events
		WHERE organization_id = $1
		  AND due_date >= $2::DATE AND due_date <= $3::DATE
		GROUP BY priority`,
		orgID, monthStart.Format("2006-01-02"), monthEnd.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer priRows.Close()

	for priRows.Next() {
		var pri string
		var count int64
		if err := priRows.Scan(&pri, &count); err != nil {
			continue
		}
		summary.ByPriority[pri] = count
	}

	// Critical deadlines in the month
	critRows, err := tx.Query(ctx, `
		SELECT ce.id, ce.event_ref, ce.event_type, ce.category, ce.priority, ce.status,
			   ce.title, ce.due_date::TEXT,
			   GREATEST((ce.due_date - CURRENT_DATE)::INT, 0) AS days_left,
			   ce.source_entity_type, ce.source_entity_id, ce.source_entity_ref,
			   ce.assigned_to_user_id,
			   COALESCE(u.first_name || ' ' || u.last_name, '')
		FROM calendar_events ce
		LEFT JOIN users u ON u.id = ce.assigned_to_user_id
		WHERE ce.organization_id = $1
		  AND ce.due_date >= $2::DATE AND ce.due_date <= $3::DATE
		  AND ce.priority IN ('critical', 'high')
		  AND ce.status NOT IN ('completed', 'cancelled')
		ORDER BY ce.due_date ASC
		LIMIT 20`,
		orgID, monthStart.Format("2006-01-02"), monthEnd.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer critRows.Close()

	for critRows.Next() {
		var d UpcomingDeadline
		if err := critRows.Scan(
			&d.ID, &d.EventRef, &d.EventType, &d.Category, &d.Priority, &d.Status,
			&d.Title, &d.DueDate, &d.DaysLeft,
			&d.SourceEntityType, &d.SourceEntityID, &d.SourceEntityRef,
			&d.AssignedToUserID, &d.AssignedToName,
		); err != nil {
			continue
		}
		summary.CriticalDeadlines = append(summary.CriticalDeadlines, d)
	}

	// Weekly summary
	weekRows, err := tx.Query(ctx, `
		SELECT
			DATE_TRUNC('week', due_date)::DATE::TEXT AS week_start,
			(DATE_TRUNC('week', due_date) + INTERVAL '6 days')::DATE::TEXT AS week_end,
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE status = 'completed') AS completed,
			COUNT(*) FILTER (WHERE status = 'overdue') AS overdue
		FROM calendar_events
		WHERE organization_id = $1
		  AND due_date >= $2::DATE AND due_date <= $3::DATE
		GROUP BY DATE_TRUNC('week', due_date)
		ORDER BY week_start`,
		orgID, monthStart.Format("2006-01-02"), monthEnd.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer weekRows.Close()

	for weekRows.Next() {
		var ws WeekSummary
		if err := weekRows.Scan(&ws.WeekStart, &ws.WeekEnd, &ws.Total, &ws.Completed, &ws.Overdue); err != nil {
			continue
		}
		summary.WeeklySummary = append(summary.WeeklySummary, ws)
	}

	tx.Commit(ctx)
	return summary, nil
}

// ============================================================
// iCAL EXPORT
// ============================================================

// ExportICalFeed generates an iCal (RFC 5545) feed for a user based on their token.
func (s *CalendarService) ExportICalFeed(ctx context.Context, token string) (string, error) {
	// Look up subscription by token
	var orgID, userID uuid.UUID
	var categories []string
	var priorities []string

	err := s.pool.QueryRow(ctx, `
		SELECT cs.organization_id, cs.user_id, cs.subscribed_categories, cs.subscribed_priorities
		FROM calendar_subscriptions cs
		WHERE cs.ical_token = $1
		  AND cs.ical_export_enabled = true
		  AND (cs.ical_token_expires_at IS NULL OR cs.ical_token_expires_at > NOW())`, token,
	).Scan(&orgID, &userID, &categories, &priorities)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", fmt.Errorf("invalid or expired iCal token")
		}
		return "", fmt.Errorf("lookup token: %w", err)
	}

	// Fetch events for the user's subscribed categories
	rows, err := s.pool.Query(ctx, `
		SELECT ce.id, ce.event_ref, ce.title, ce.description,
			   ce.due_date, ce.due_time, ce.start_date, ce.end_date,
			   ce.all_day, ce.status, ce.priority, ce.category,
			   ce.created_at, ce.updated_at, ce.location
		FROM calendar_events ce
		WHERE ce.organization_id = $1
		  AND ce.category::TEXT = ANY($2)
		  AND ce.priority::TEXT = ANY($3)
		  AND ce.due_date >= CURRENT_DATE - INTERVAL '30 days'
		  AND ce.due_date <= CURRENT_DATE + INTERVAL '365 days'
		ORDER BY ce.due_date ASC`, orgID, categories, priorities)
	if err != nil {
		return "", fmt.Errorf("query events for ical: %w", err)
	}
	defer rows.Close()

	// Build iCal content
	var b strings.Builder
	b.WriteString("BEGIN:VCALENDAR\r\n")
	b.WriteString("VERSION:2.0\r\n")
	b.WriteString("PRODID:-//ComplianceForge//Calendar//EN\r\n")
	b.WriteString("CALSCALE:GREGORIAN\r\n")
	b.WriteString("METHOD:PUBLISH\r\n")
	b.WriteString("X-WR-CALNAME:ComplianceForge Calendar\r\n")

	for rows.Next() {
		var id uuid.UUID
		var ref, title, description string
		var dueDate string
		var dueTime, startDateStr, endDateStr, location *string
		var allDay bool
		var status, priority, category string
		var createdAt, updatedAt time.Time

		if err := rows.Scan(
			&id, &ref, &title, &description,
			&dueDate, &dueTime, &startDateStr, &endDateStr,
			&allDay, &status, &priority, &category,
			&createdAt, &updatedAt, &location,
		); err != nil {
			continue
		}

		b.WriteString("BEGIN:VEVENT\r\n")
		b.WriteString(fmt.Sprintf("UID:%s@complianceforge.io\r\n", id.String()))
		b.WriteString(fmt.Sprintf("DTSTAMP:%s\r\n", updatedAt.UTC().Format("20060102T150405Z")))

		if allDay {
			b.WriteString(fmt.Sprintf("DTSTART;VALUE=DATE:%s\r\n", strings.ReplaceAll(dueDate, "-", "")))
		} else {
			dtStart := strings.ReplaceAll(dueDate, "-", "") + "T090000Z"
			b.WriteString(fmt.Sprintf("DTSTART:%s\r\n", dtStart))
		}

		b.WriteString(fmt.Sprintf("SUMMARY:[%s] %s\r\n", strings.ToUpper(priority), icalEscape(title)))
		if description != "" {
			b.WriteString(fmt.Sprintf("DESCRIPTION:%s\r\n", icalEscape(description)))
		}
		b.WriteString(fmt.Sprintf("CATEGORIES:%s\r\n", category))
		b.WriteString(fmt.Sprintf("X-CF-REF:%s\r\n", ref))
		b.WriteString(fmt.Sprintf("X-CF-STATUS:%s\r\n", status))

		// Priority mapping: 1=high, 5=normal, 9=low
		switch priority {
		case "critical":
			b.WriteString("PRIORITY:1\r\n")
		case "high":
			b.WriteString("PRIORITY:2\r\n")
		case "medium":
			b.WriteString("PRIORITY:5\r\n")
		default:
			b.WriteString("PRIORITY:9\r\n")
		}

		b.WriteString("END:VEVENT\r\n")
	}

	b.WriteString("END:VCALENDAR\r\n")

	return b.String(), nil
}

// ============================================================
// REMINDER & STATUS QUERIES (used by worker)
// ============================================================

// GetEventsNeedingReminders returns events that need a reminder sent.
func (s *CalendarService) GetEventsNeedingReminders(ctx context.Context, orgID uuid.UUID) ([]CalendarEvent, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT ce.id, ce.organization_id, ce.event_ref, ce.event_type, ce.category,
			   ce.priority, ce.status, ce.title, ce.description,
			   ce.source_entity_type, ce.source_entity_id, ce.source_entity_ref,
			   ce.due_date::TEXT, ce.reminder_days_before, ce.last_reminder_sent_at,
			   ce.reminder_count, ce.assigned_to_user_id, ce.owner_user_id
		FROM calendar_events ce
		WHERE ce.organization_id = $1
		  AND ce.status NOT IN ('completed', 'cancelled')
		  AND ce.due_date >= CURRENT_DATE
		  AND ce.due_date <= CURRENT_DATE + INTERVAL '30 days'
		  AND (ce.last_reminder_sent_at IS NULL OR ce.last_reminder_sent_at < NOW() - INTERVAL '4 hours')
		ORDER BY ce.due_date ASC`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []CalendarEvent
	for rows.Next() {
		var e CalendarEvent
		if err := rows.Scan(
			&e.ID, &e.OrganizationID, &e.EventRef, &e.EventType, &e.Category,
			&e.Priority, &e.Status, &e.Title, &e.Description,
			&e.SourceEntityType, &e.SourceEntityID, &e.SourceEntityRef,
			&e.DueDate, &e.ReminderDaysBefore, &e.LastReminderSentAt,
			&e.ReminderCount, &e.AssignedToUserID, &e.OwnerUserID,
		); err != nil {
			continue
		}
		events = append(events, e)
	}
	return events, nil
}

// MarkReminderSent records that a reminder was sent for an event.
func (s *CalendarService) MarkReminderSent(ctx context.Context, eventID uuid.UUID) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE calendar_events
		SET last_reminder_sent_at = NOW(), reminder_count = reminder_count + 1
		WHERE id = $1`, eventID)
	return err
}

// UpdateOverdueStatuses transitions events past their due date to overdue.
func (s *CalendarService) UpdateOverdueStatuses(ctx context.Context, orgID uuid.UUID) (int64, error) {
	tag, err := s.pool.Exec(ctx, `
		UPDATE calendar_events
		SET status = 'overdue'
		WHERE organization_id = $1
		  AND status IN ('upcoming', 'due_soon')
		  AND due_date < CURRENT_DATE`, orgID)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

// UpdateDueSoonStatuses transitions events approaching due date.
func (s *CalendarService) UpdateDueSoonStatuses(ctx context.Context, orgID uuid.UUID) (int64, error) {
	tag, err := s.pool.Exec(ctx, `
		UPDATE calendar_events
		SET status = 'due_soon'
		WHERE organization_id = $1
		  AND status = 'upcoming'
		  AND due_date <= CURRENT_DATE + INTERVAL '3 days'
		  AND due_date >= CURRENT_DATE`, orgID)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

// EscalateOverdueEvents marks events as escalated if they have been overdue
// longer than their escalation_after_days threshold.
func (s *CalendarService) EscalateOverdueEvents(ctx context.Context, orgID uuid.UUID) (int64, error) {
	tag, err := s.pool.Exec(ctx, `
		UPDATE calendar_events
		SET escalated = true, escalated_at = NOW(),
			escalated_to_user_id = owner_user_id
		WHERE organization_id = $1
		  AND status = 'overdue'
		  AND escalated = false
		  AND (CURRENT_DATE - due_date)::INT >= escalation_after_days`, orgID)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

// ============================================================
// SCAN HELPERS
// ============================================================

func (s *CalendarService) scanEventFromRows(rows pgx.Rows) (*CalendarEvent, error) {
	var e CalendarEvent
	err := rows.Scan(
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

// icalEscape escapes special characters for iCal format.
func icalEscape(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, ";", "\\;")
	s = strings.ReplaceAll(s, ",", "\\,")
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
}
