package worker

import (
	"context"
	"strings"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"github.com/complianceforge/platform/internal/service"
)

// ============================================================
// CALENDAR WORKER — Background Processing for Calendar Engine
// ============================================================

// CalendarWorker manages periodic background tasks for the compliance
// calendar including reminders, overdue escalation, status updates, and
// recurring event generation.
type CalendarWorker struct {
	svc  *service.CalendarService
	pool *pgxpool.Pool
}

// NewCalendarWorker creates a new CalendarWorker.
func NewCalendarWorker(svc *service.CalendarService, pool *pgxpool.Pool) *CalendarWorker {
	return &CalendarWorker{
		svc:  svc,
		pool: pool,
	}
}

// Start begins the calendar worker. It runs three loops:
//   - ReminderScheduler: every 15 minutes — send reminders for approaching deadlines
//   - OverdueEscalator: every hour — escalate overdue items
//   - StatusUpdater: every 30 minutes — update statuses and generate recurring events
//
// Blocks until ctx is cancelled.
func (w *CalendarWorker) Start(ctx context.Context) {
	log.Info().Msg("Calendar worker started")

	go w.runLoop(ctx, "calendar_reminders", 15*time.Minute, w.ReminderScheduler)
	go w.runLoop(ctx, "calendar_overdue", 1*time.Hour, w.OverdueEscalator)
	go w.runLoop(ctx, "calendar_status", 30*time.Minute, w.StatusUpdater)

	<-ctx.Done()
	log.Info().Msg("Calendar worker shutting down")
}

// runLoop executes a job function at the specified interval.
func (w *CalendarWorker) runLoop(ctx context.Context, name string, interval time.Duration, job func(context.Context) error) {
	// Run once on startup after a small delay
	timer := time.NewTimer(5 * time.Second)
	select {
	case <-ctx.Done():
		timer.Stop()
		return
	case <-timer.C:
		log.Info().Str("job", name).Msg("Running initial calendar job")
		if err := job(ctx); err != nil {
			log.Error().Err(err).Str("job", name).Msg("Initial calendar job failed")
		}
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			log.Debug().Str("job", name).Msg("Running scheduled calendar job")
			if err := job(ctx); err != nil {
				log.Error().Err(err).Str("job", name).Msg("Calendar job failed")
			}
		}
	}
}

// ============================================================
// REMINDER SCHEDULER — Runs every 15 minutes
// ============================================================

// ReminderScheduler checks all upcoming events and sends reminders when
// the event is within the configured reminder_days_before thresholds.
func (w *CalendarWorker) ReminderScheduler(ctx context.Context) error {
	log.Info().Msg("Running calendar reminder scheduler")

	orgIDs, err := w.getActiveOrgIDs(ctx)
	if err != nil {
		return err
	}

	totalReminders := 0

	for _, orgID := range orgIDs {
		events, err := w.svc.GetEventsNeedingReminders(ctx, orgID)
		if err != nil {
			log.Error().Err(err).Str("org_id", orgID.String()).Msg("Failed to get events needing reminders")
			continue
		}

		now := time.Now()
		for _, event := range events {
			dueDate, err := time.Parse("2006-01-02", event.DueDate)
			if err != nil {
				continue
			}

			daysUntilDue := int(time.Until(dueDate).Hours() / 24)
			if daysUntilDue < 0 {
				continue
			}

			// Check if current day matches any reminder threshold
			shouldRemind := false
			for _, threshold := range event.ReminderDaysBefore {
				if daysUntilDue == threshold {
					shouldRemind = true
					break
				}
			}

			// Also remind on the due date itself
			if daysUntilDue == 0 {
				shouldRemind = true
			}

			if !shouldRemind {
				continue
			}

			// Avoid duplicate reminders within 4 hours
			if event.LastReminderSentAt != nil && now.Sub(*event.LastReminderSentAt) < 4*time.Hour {
				continue
			}

			// Log the reminder (in production, this would trigger a notification)
			log.Info().
				Str("org_id", orgID.String()).
				Str("event_ref", event.EventRef).
				Str("title", event.Title).
				Str("priority", event.Priority).
				Int("days_until_due", daysUntilDue).
				Str("category", event.Category).
				Msg("Calendar reminder triggered")

			if err := w.svc.MarkReminderSent(ctx, event.ID); err != nil {
				log.Error().Err(err).Str("event_id", event.ID.String()).Msg("Failed to mark reminder sent")
				continue
			}

			totalReminders++
		}
	}

	log.Info().
		Int("total_reminders", totalReminders).
		Int("orgs_processed", len(orgIDs)).
		Msg("Calendar reminder scheduler complete")

	return nil
}

// ============================================================
// OVERDUE ESCALATOR — Runs every hour
// ============================================================

// OverdueEscalator checks for overdue events and escalates them
// when they have exceeded their escalation threshold.
func (w *CalendarWorker) OverdueEscalator(ctx context.Context) error {
	log.Info().Msg("Running calendar overdue escalator")

	orgIDs, err := w.getActiveOrgIDs(ctx)
	if err != nil {
		return err
	}

	totalEscalated := int64(0)

	for _, orgID := range orgIDs {
		// First, update overdue statuses
		overdueCount, err := w.svc.UpdateOverdueStatuses(ctx, orgID)
		if err != nil {
			log.Error().Err(err).Str("org_id", orgID.String()).Msg("Failed to update overdue statuses")
			continue
		}

		// Then, update due-soon statuses
		dueSoonCount, err := w.svc.UpdateDueSoonStatuses(ctx, orgID)
		if err != nil {
			log.Error().Err(err).Str("org_id", orgID.String()).Msg("Failed to update due-soon statuses")
			continue
		}

		// Escalate events that have been overdue past their threshold
		escalated, err := w.svc.EscalateOverdueEvents(ctx, orgID)
		if err != nil {
			log.Error().Err(err).Str("org_id", orgID.String()).Msg("Failed to escalate overdue events")
			continue
		}

		if overdueCount > 0 || dueSoonCount > 0 || escalated > 0 {
			log.Info().
				Str("org_id", orgID.String()).
				Int64("new_overdue", overdueCount).
				Int64("new_due_soon", dueSoonCount).
				Int64("escalated", escalated).
				Msg("Calendar status updates applied")
		}

		totalEscalated += escalated
	}

	log.Info().
		Int64("total_escalated", totalEscalated).
		Int("orgs_processed", len(orgIDs)).
		Msg("Calendar overdue escalator complete")

	return nil
}

// ============================================================
// STATUS UPDATER — Runs every 30 minutes
// ============================================================

// StatusUpdater handles recurring events via RRULE parsing and ensures
// event statuses are consistent. It generates the next occurrence of
// recurring events when the current occurrence is completed.
func (w *CalendarWorker) StatusUpdater(ctx context.Context) error {
	log.Info().Msg("Running calendar status updater")

	orgIDs, err := w.getActiveOrgIDs(ctx)
	if err != nil {
		return err
	}

	totalGenerated := 0

	for _, orgID := range orgIDs {
		generated, err := w.processRecurringEvents(ctx, orgID)
		if err != nil {
			log.Error().Err(err).Str("org_id", orgID.String()).Msg("Failed to process recurring events")
			continue
		}
		totalGenerated += generated
	}

	log.Info().
		Int("total_recurring_generated", totalGenerated).
		Int("orgs_processed", len(orgIDs)).
		Msg("Calendar status updater complete")

	return nil
}

// processRecurringEvents finds completed recurring events and generates the
// next occurrence based on their recurrence settings.
func (w *CalendarWorker) processRecurringEvents(ctx context.Context, orgID uuid.UUID) (int, error) {
	rows, err := w.pool.Query(ctx, `
		SELECT ce.id, ce.event_ref, ce.event_type, ce.category, ce.priority,
			   ce.title, ce.description, ce.source_entity_type, ce.source_entity_id,
			   ce.source_entity_ref, ce.due_date, ce.recurrence_type, ce.rrule,
			   ce.recurrence_end_date, ce.assigned_to_user_id, ce.owner_user_id,
			   ce.reminder_days_before, ce.tags
		FROM calendar_events ce
		WHERE ce.organization_id = $1
		  AND ce.status = 'completed'
		  AND ce.recurrence_type != 'none'
		  AND (ce.recurrence_end_date IS NULL OR ce.recurrence_end_date > CURRENT_DATE)
		  AND NOT EXISTS (
			SELECT 1 FROM calendar_events child
			WHERE child.parent_event_id = ce.id
			  AND child.status NOT IN ('completed', 'cancelled')
		  )`, orgID)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	generated := 0
	for rows.Next() {
		var id uuid.UUID
		var ref, eventType, category, priority, title, description string
		var sourceType string
		var sourceID uuid.UUID
		var sourceRef string
		var dueDate time.Time
		var recType, rrule string
		var recEndDate *time.Time
		var assignedTo, ownerID *uuid.UUID
		var reminderDays []int
		var tags []string

		if err := rows.Scan(
			&id, &ref, &eventType, &category, &priority,
			&title, &description, &sourceType, &sourceID,
			&sourceRef, &dueDate, &recType, &rrule,
			&recEndDate, &assignedTo, &ownerID,
			&reminderDays, &tags,
		); err != nil {
			continue
		}

		nextDate := computeNextOccurrence(dueDate, recType, rrule)
		if nextDate.IsZero() {
			continue
		}

		// Check recurrence end date
		if recEndDate != nil && nextDate.After(*recEndDate) {
			continue
		}

		// Create next occurrence
		_, err := w.pool.Exec(ctx, `
			INSERT INTO calendar_events (
				organization_id, event_type, category, priority, status,
				title, description, source_entity_type, source_entity_id, source_entity_ref,
				due_date, all_day, recurrence_type, rrule, recurrence_end_date,
				parent_event_id, occurrence_date,
				assigned_to_user_id, owner_user_id, reminder_days_before, tags
			) VALUES (
				$1, $2::calendar_event_type, $3::calendar_event_category, $4::calendar_event_priority,
				CASE
					WHEN $10::DATE < CURRENT_DATE THEN 'overdue'::calendar_event_status
					WHEN $10::DATE <= CURRENT_DATE + INTERVAL '3 days' THEN 'due_soon'::calendar_event_status
					ELSE 'upcoming'::calendar_event_status
				END,
				$5, $6, $7, $8, $9,
				$10::DATE, true, $11::calendar_recurrence_type, $12, $13,
				$14, $10::DATE,
				$15, $16, $17, $18
			)`,
			orgID, eventType, category, priority,
			title, description, sourceType, sourceID, sourceRef,
			nextDate.Format("2006-01-02"), recType, rrule, recEndDate,
			id, assignedTo, ownerID, reminderDays, tags,
		)
		if err != nil {
			log.Error().Err(err).
				Str("parent_ref", ref).
				Str("next_date", nextDate.Format("2006-01-02")).
				Msg("Failed to generate recurring event")
			continue
		}

		generated++
		log.Debug().
			Str("parent_ref", ref).
			Str("next_date", nextDate.Format("2006-01-02")).
			Str("recurrence_type", recType).
			Msg("Generated recurring calendar event")
	}

	return generated, nil
}

// ============================================================
// RRULE PARSING
// ============================================================

// computeNextOccurrence calculates the next occurrence date based on recurrence.
// Supports both simple recurrence types and RFC 5545 RRULE strings.
func computeNextOccurrence(currentDate time.Time, recType, rrule string) time.Time {
	// Simple recurrence types
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

// parseRRule parses a simplified RFC 5545 RRULE string and computes the next date.
// Supports: FREQ=DAILY|WEEKLY|MONTHLY|YEARLY and INTERVAL=N
// Example: "FREQ=MONTHLY;INTERVAL=3" produces quarterly recurrence.
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

// ============================================================
// HELPERS
// ============================================================

// getActiveOrgIDs returns all active organization IDs.
func (w *CalendarWorker) getActiveOrgIDs(ctx context.Context) ([]uuid.UUID, error) {
	rows, err := w.pool.Query(ctx, `
		SELECT id FROM organizations
		WHERE status = 'active'
		ORDER BY id
		LIMIT 500`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			continue
		}
		ids = append(ids, id)
	}
	return ids, nil
}
