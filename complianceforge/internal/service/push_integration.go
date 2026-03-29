// Package service provides the Push Integration layer for ComplianceForge.
// It connects the NotificationEngine's event bus to the PushService,
// translating domain events into push notifications with proper preference
// enforcement, quiet hours handling, and critical breach alert overrides.
package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// ============================================================
// EVENT-TO-PUSH MAPPING
// ============================================================

// pushMapping defines how a notification engine event maps to a push notification.
type pushMapping struct {
	PushType   string
	TitleFmt   string
	BodyFmt    string
	Priority   string
	IsCritical bool // Critical notifications bypass quiet hours
}

// eventPushMappings maps notification engine event types to push notification parameters.
var eventPushMappings = map[string]pushMapping{
	// Breach alerts — ALWAYS sent, bypass all preferences and quiet hours
	EventBreachDeadlineApproaching: {
		PushType:   PushTypeBreachAlert,
		TitleFmt:   "URGENT: Data Breach Deadline Approaching",
		BodyFmt:    "A data breach is approaching the 72-hour GDPR notification deadline. Immediate action required.",
		Priority:   "high",
		IsCritical: true,
	},
	EventBreachDeadlineExpired: {
		PushType:   PushTypeBreachAlert,
		TitleFmt:   "CRITICAL: Data Breach Deadline EXPIRED",
		BodyFmt:    "A data breach has exceeded the 72-hour GDPR notification deadline. Escalate immediately.",
		Priority:   "high",
		IsCritical: true,
	},
	EventBreachDPANotified: {
		PushType:   PushTypeBreachAlert,
		TitleFmt:   "Breach: DPA Notified",
		BodyFmt:    "The Data Protection Authority has been notified of a data breach.",
		Priority:   "high",
		IsCritical: true,
	},

	// NIS2 critical events
	EventNIS2EarlyWarningDue: {
		PushType:   PushTypeBreachAlert,
		TitleFmt:   "NIS2: Early Warning Due Within 24 Hours",
		BodyFmt:    "A significant incident requires NIS2 early warning notification to CSIRT within 24 hours.",
		Priority:   "high",
		IsCritical: true,
	},
	EventNIS2NotificationDue: {
		PushType:   PushTypeBreachAlert,
		TitleFmt:   "NIS2: Incident Notification Due Within 72 Hours",
		BodyFmt:    "A significant incident requires NIS2 incident notification within 72 hours.",
		Priority:   "high",
		IsCritical: true,
	},

	// Incident events
	EventIncidentCreated: {
		PushType:   PushTypeIncidentCreated,
		TitleFmt:   "New Incident Reported",
		BodyFmt:    "A new security incident has been reported and needs attention.",
		Priority:   "high",
		IsCritical: false,
	},
	EventIncidentUpdated: {
		PushType:   PushTypeIncidentCreated,
		TitleFmt:   "Incident Updated",
		BodyFmt:    "A security incident you are tracking has been updated.",
		Priority:   "normal",
		IsCritical: false,
	},

	// Policy events
	EventPolicyReviewDue: {
		PushType:   PushTypeDeadlineReminder,
		TitleFmt:   "Policy Review Due",
		BodyFmt:    "A policy is due for review. Please review and update as needed.",
		Priority:   "normal",
		IsCritical: false,
	},
	EventPolicyReviewOverdue: {
		PushType:   PushTypeDeadlineReminder,
		TitleFmt:   "Policy Review Overdue",
		BodyFmt:    "A policy review is overdue. Please take action immediately.",
		Priority:   "high",
		IsCritical: false,
	},

	// Finding events
	EventFindingCreated: {
		PushType:   PushTypeIncidentCreated,
		TitleFmt:   "New Audit Finding",
		BodyFmt:    "A new audit finding has been created and requires remediation.",
		Priority:   "normal",
		IsCritical: false,
	},
	EventFindingRemediationOverdue: {
		PushType:   PushTypeDeadlineReminder,
		TitleFmt:   "Finding Remediation Overdue",
		BodyFmt:    "An audit finding remediation deadline has passed. Please escalate.",
		Priority:   "high",
		IsCritical: false,
	},

	// Attestation events
	EventAttestationRequired: {
		PushType:   PushTypeApprovalRequest,
		TitleFmt:   "Attestation Required",
		BodyFmt:    "You have a new policy attestation to complete.",
		Priority:   "normal",
		IsCritical: false,
	},
	EventAttestationOverdue: {
		PushType:   PushTypeDeadlineReminder,
		TitleFmt:   "Attestation Overdue",
		BodyFmt:    "Your policy attestation is overdue. Please complete it urgently.",
		Priority:   "high",
		IsCritical: false,
	},

	// Risk events
	EventRiskThresholdExceeded: {
		PushType:   PushTypeIncidentCreated,
		TitleFmt:   "Risk Threshold Exceeded",
		BodyFmt:    "A risk has exceeded the acceptable threshold and requires immediate treatment.",
		Priority:   "high",
		IsCritical: false,
	},
	EventRiskReviewDue: {
		PushType:   PushTypeDeadlineReminder,
		TitleFmt:   "Risk Review Due",
		BodyFmt:    "A risk is due for periodic review.",
		Priority:   "normal",
		IsCritical: false,
	},

	// Compliance events
	EventComplianceScoreDropped: {
		PushType:   PushTypeIncidentCreated,
		TitleFmt:   "Compliance Score Dropped",
		BodyFmt:    "Your organisation's compliance score has dropped significantly.",
		Priority:   "high",
		IsCritical: false,
	},

	// Vendor events
	EventVendorAssessmentDue: {
		PushType:   PushTypeDeadlineReminder,
		TitleFmt:   "Vendor Assessment Due",
		BodyFmt:    "A third-party vendor assessment is due for renewal.",
		Priority:   "normal",
		IsCritical: false,
	},
	EventVendorRiskChanged: {
		PushType:   PushTypeIncidentCreated,
		TitleFmt:   "Vendor Risk Level Changed",
		BodyFmt:    "A vendor's risk classification has changed. Review required.",
		Priority:   "normal",
		IsCritical: false,
	},

	// Control events
	EventControlTestFailed: {
		PushType:   PushTypeIncidentCreated,
		TitleFmt:   "Control Test Failed",
		BodyFmt:    "A control effectiveness test has failed. Remediation may be needed.",
		Priority:   "high",
		IsCritical: false,
	},
}

// ============================================================
// PUSH INTEGRATION SERVICE
// ============================================================

// PushIntegration bridges the NotificationEngine event bus with the PushService.
// It subscribes to events and dispatches push notifications to affected users.
type PushIntegration struct {
	db   *pgxpool.Pool
	push *PushService
	bus  *EventBus
}

// NewPushIntegration creates a new PushIntegration.
func NewPushIntegration(db *pgxpool.Pool, push *PushService, bus *EventBus) *PushIntegration {
	return &PushIntegration{
		db:   db,
		push: push,
		bus:  bus,
	}
}

// Start begins listening for events and dispatching push notifications.
// It blocks until the context is cancelled.
func (pi *PushIntegration) Start(ctx context.Context) {
	events := pi.bus.Subscribe()
	log.Info().Msg("Push integration service started")

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Push integration service shutting down")
			return
		case event, ok := <-events:
			if !ok {
				log.Info().Msg("Push integration: event bus channel closed")
				return
			}
			pi.handleEvent(ctx, event)
		}
	}
}

// handleEvent processes a single event and sends push notifications.
func (pi *PushIntegration) handleEvent(ctx context.Context, event Event) {
	mapping, exists := eventPushMappings[event.Type]
	if !exists {
		// No push mapping for this event type
		return
	}

	logger := log.With().
		Str("event_type", event.Type).
		Str("push_type", mapping.PushType).
		Str("org_id", event.OrgID.String()).
		Logger()

	logger.Debug().Msg("Processing event for push notification")

	// Build the push notification
	notification := &PushNotification{
		Type:     mapping.PushType,
		Title:    enrichTitle(mapping.TitleFmt, event),
		Body:     enrichBody(mapping.BodyFmt, event),
		Priority: mapping.Priority,
		Data: map[string]interface{}{
			"event_type": event.Type,
			"org_id":     event.OrgID.String(),
		},
	}

	// Copy relevant payload fields into push data
	if event.Payload != nil {
		for _, key := range []string{"incident_id", "policy_id", "risk_id", "finding_id", "vendor_id", "control_id", "reference", "severity"} {
			if val, ok := event.Payload[key]; ok {
				notification.Data[key] = val
			}
		}
	}

	// Add sound for critical alerts
	if mapping.IsCritical {
		notification.Sound = "critical_alert"
	}

	// Resolve recipients for this event
	recipients, err := pi.resolveEventRecipients(ctx, event)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to resolve push notification recipients")
		return
	}

	if len(recipients) == 0 {
		logger.Debug().Msg("No push recipients for event")
		return
	}

	// For critical alerts (breach), we call SendPush directly for each user
	// because SendPush will handle the breach alert bypass internally.
	// For non-critical, we use bulk send.
	if mapping.IsCritical {
		for _, userID := range recipients {
			if err := pi.push.SendPush(ctx, event.OrgID, userID, notification); err != nil {
				logger.Warn().
					Err(err).
					Str("user_id", userID.String()).
					Msg("Failed to send critical push notification")
			}
		}
		logger.Info().
			Int("recipients", len(recipients)).
			Msg("Critical push notifications dispatched")
	} else {
		if err := pi.push.SendBulkPush(ctx, event.OrgID, recipients, notification); err != nil {
			logger.Warn().Err(err).Msg("Bulk push send had failures")
		}
		logger.Info().
			Int("recipients", len(recipients)).
			Msg("Push notifications dispatched")
	}
}

// resolveEventRecipients determines which users should receive push notifications for an event.
func (pi *PushIntegration) resolveEventRecipients(ctx context.Context, event Event) ([]uuid.UUID, error) {
	// If the event payload specifies target users, use those
	if userIDStr, ok := event.Payload["user_id"].(string); ok {
		if uid, err := uuid.Parse(userIDStr); err == nil {
			return []uuid.UUID{uid}, nil
		}
	}

	if userIDs, ok := event.Payload["user_ids"].([]interface{}); ok {
		var ids []uuid.UUID
		for _, raw := range userIDs {
			if uidStr, ok := raw.(string); ok {
				if uid, err := uuid.Parse(uidStr); err == nil {
					ids = append(ids, uid)
				}
			}
		}
		if len(ids) > 0 {
			return ids, nil
		}
	}

	// For breach alerts and critical events, notify all org admins
	mapping, exists := eventPushMappings[event.Type]
	if exists && mapping.IsCritical {
		return pi.getOrgAdmins(ctx, event.OrgID)
	}

	// For incident events, notify the assigned user and org admins
	if strings.HasPrefix(event.Type, "incident.") {
		var recipients []uuid.UUID
		admins, _ := pi.getOrgAdmins(ctx, event.OrgID)
		recipients = append(recipients, admins...)

		if assignedStr, ok := event.Payload["assigned_to"].(string); ok {
			if uid, err := uuid.Parse(assignedStr); err == nil {
				recipients = appendUniqueUUID(recipients, uid)
			}
		}
		return recipients, nil
	}

	// Default: notify all users with push tokens in the org who have
	// active tokens (the PushService will handle preference filtering)
	return pi.getUsersWithActiveTokens(ctx, event.OrgID)
}

// getOrgAdmins returns user IDs of all active admins in the organisation.
func (pi *PushIntegration) getOrgAdmins(ctx context.Context, orgID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := pi.db.Query(ctx, `
		SELECT DISTINCT ur.user_id FROM user_roles ur
		JOIN roles r ON ur.role_id = r.id
		WHERE r.slug = 'org_admin' AND ur.user_id IN (
			SELECT id FROM users WHERE organization_id = $1 AND status = 'active'
		)`, orgID)
	if err != nil {
		return nil, fmt.Errorf("query org admins: %w", err)
	}
	defer rows.Close()

	var userIDs []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err == nil {
			userIDs = append(userIDs, id)
		}
	}
	return userIDs, nil
}

// getUsersWithActiveTokens returns user IDs of all users with active push tokens in the org.
func (pi *PushIntegration) getUsersWithActiveTokens(ctx context.Context, orgID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := pi.db.Query(ctx, `
		SELECT DISTINCT user_id FROM push_notification_tokens
		WHERE organization_id = $1 AND is_active = true`, orgID)
	if err != nil {
		return nil, fmt.Errorf("query users with tokens: %w", err)
	}
	defer rows.Close()

	var userIDs []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err == nil {
			userIDs = append(userIDs, id)
		}
	}
	return userIDs, nil
}

// ============================================================
// HELPERS
// ============================================================

// enrichTitle adds context from event payload to the notification title.
func enrichTitle(titleFmt string, event Event) string {
	title := titleFmt
	if ref, ok := event.Payload["reference"].(string); ok && ref != "" {
		title = fmt.Sprintf("%s [%s]", title, ref)
	}
	if severity := event.Severity; severity != "" {
		title = fmt.Sprintf("[%s] %s", strings.ToUpper(severity), title)
	}
	return title
}

// enrichBody adds specific details from event payload to the notification body.
func enrichBody(bodyFmt string, event Event) string {
	body := bodyFmt
	if titleVal, ok := event.Payload["title"].(string); ok && titleVal != "" {
		body = fmt.Sprintf("%s — %s", body, titleVal)
	}
	return body
}

// appendUniqueUUID appends a UUID to a slice only if it is not already present.
func appendUniqueUUID(slice []uuid.UUID, id uuid.UUID) []uuid.UUID {
	for _, existing := range slice {
		if existing == id {
			return slice
		}
	}
	return append(slice, id)
}
