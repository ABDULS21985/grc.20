// Package service contains the Enterprise Notification Engine for ComplianceForge.
// It provides event-driven, multi-channel notification delivery with rule evaluation,
// template rendering, rate limiting, and delivery tracking.
package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"text/template"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// ============================================================
// EVENT TYPES
// ============================================================

const (
	// Incident events
	EventIncidentCreated  = "incident.created"
	EventIncidentUpdated  = "incident.updated"
	EventIncidentResolved = "incident.resolved"
	EventIncidentClosed   = "incident.closed"

	// Breach events (GDPR Article 33)
	EventBreachDeadlineApproaching = "breach.deadline_approaching"
	EventBreachDeadlineExpired     = "breach.deadline_expired"
	EventBreachDPANotified         = "breach.dpa_notified"

	// NIS2 events (Article 23)
	EventNIS2EarlyWarningDue    = "nis2.early_warning_due"
	EventNIS2NotificationDue    = "nis2.notification_due"
	EventNIS2EarlyWarningFiled  = "nis2.early_warning_filed"

	// Control events
	EventControlStatusChanged   = "control.status_changed"
	EventControlTestFailed      = "control.test_failed"
	EventControlTestPassed      = "control.test_passed"

	// Policy events
	EventPolicyReviewDue        = "policy.review_due"
	EventPolicyReviewOverdue    = "policy.review_overdue"
	EventPolicyPublished        = "policy.published"

	// Attestation events
	EventAttestationRequired    = "attestation.required"
	EventAttestationOverdue     = "attestation.overdue"
	EventAttestationCompleted   = "attestation.completed"

	// Audit & Finding events
	EventFindingCreated           = "finding.created"
	EventFindingRemediationOverdue = "finding.remediation_overdue"
	EventAuditCompleted           = "audit.completed"

	// Vendor events
	EventVendorAssessmentDue    = "vendor.assessment_due"
	EventVendorMissingDPA       = "vendor.missing_dpa"
	EventVendorRiskChanged      = "vendor.risk_changed"

	// Risk events
	EventRiskThresholdExceeded  = "risk.threshold_exceeded"
	EventRiskReviewDue          = "risk.review_due"

	// Compliance events
	EventComplianceScoreDropped = "compliance.score_dropped"

	// User events
	EventUserWelcome            = "user.welcome"
	EventUserPasswordReset      = "user.password_reset"
)

// ============================================================
// EVENT
// ============================================================

// Event represents a domain event that may trigger notifications.
type Event struct {
	Type      string                 `json:"type"`
	OrgID     uuid.UUID              `json:"org_id"`
	Payload   map[string]interface{} `json:"payload"`
	Severity  string                 `json:"severity"`
	Timestamp time.Time              `json:"timestamp"`
}

// ============================================================
// EVENT BUS
// ============================================================

// EventBus provides an in-process channel-based pub/sub mechanism.
type EventBus struct {
	subscribers []chan Event
	mu          sync.RWMutex
	bufferSize  int
}

// NewEventBus creates a new EventBus with a default channel buffer.
func NewEventBus(bufferSize int) *EventBus {
	if bufferSize <= 0 {
		bufferSize = 256
	}
	return &EventBus{
		bufferSize: bufferSize,
	}
}

// Subscribe returns a new channel that receives events.
func (bus *EventBus) Subscribe() chan Event {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	ch := make(chan Event, bus.bufferSize)
	bus.subscribers = append(bus.subscribers, ch)
	return ch
}

// Publish sends an event to all subscribers. Non-blocking: drops if buffer is full.
func (bus *EventBus) Publish(event Event) {
	bus.mu.RLock()
	defer bus.mu.RUnlock()
	for _, ch := range bus.subscribers {
		select {
		case ch <- event:
		default:
			log.Warn().Str("event_type", event.Type).Msg("Event bus subscriber channel full, dropping event")
		}
	}
}

// Close closes all subscriber channels.
func (bus *EventBus) Close() {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	for _, ch := range bus.subscribers {
		close(ch)
	}
	bus.subscribers = nil
}

// ============================================================
// NOTIFICATION CHANNEL INTERFACE
// ============================================================

// ChannelType identifies a notification delivery channel.
type ChannelType string

const (
	ChannelEmail   ChannelType = "email"
	ChannelInApp   ChannelType = "in_app"
	ChannelSlack   ChannelType = "slack"
	ChannelWebhook ChannelType = "webhook"
)

// NotificationChannel is the interface all delivery channels must implement.
type NotificationChannel interface {
	// Type returns the channel type identifier.
	Type() ChannelType
	// Send delivers a notification. Returns an error if delivery fails.
	Send(ctx context.Context, notification *Notification) error
}

// ============================================================
// NOTIFICATION MODELS (in-memory, mirroring DB)
// ============================================================

// NotificationRule matches events to notification actions.
type NotificationRule struct {
	ID                     uuid.UUID   `json:"id"`
	OrganizationID         uuid.UUID   `json:"organization_id"`
	Name                   string      `json:"name"`
	EventType              string      `json:"event_type"`
	SeverityFilter         []string    `json:"severity_filter"`
	Conditions             map[string]interface{} `json:"conditions"`
	ChannelIDs             []uuid.UUID `json:"channel_ids"`
	RecipientType          string      `json:"recipient_type"`
	RecipientIDs           []uuid.UUID `json:"recipient_ids"`
	TemplateID             *uuid.UUID  `json:"template_id"`
	IsActive               bool        `json:"is_active"`
	CooldownMinutes        int         `json:"cooldown_minutes"`
	EscalationAfterMinutes int         `json:"escalation_after_minutes"`
	EscalationChannelIDs   []uuid.UUID `json:"escalation_channel_ids"`
}

// NotificationTemplate holds a renderable notification template.
type NotificationTemplate struct {
	ID                     uuid.UUID              `json:"id"`
	OrganizationID         *uuid.UUID             `json:"organization_id"`
	Name                   string                 `json:"name"`
	EventType              string                 `json:"event_type"`
	SubjectTemplate        string                 `json:"subject_template"`
	BodyHTMLTemplate       string                 `json:"body_html_template"`
	BodyTextTemplate       string                 `json:"body_text_template"`
	InAppTitleTemplate     string                 `json:"in_app_title_template"`
	InAppBodyTemplate      string                 `json:"in_app_body_template"`
	SlackTemplate          map[string]interface{} `json:"slack_template"`
	WebhookPayloadTemplate map[string]interface{} `json:"webhook_payload_template"`
	Variables              []string               `json:"variables"`
	IsSystem               bool                   `json:"is_system"`
}

// Notification represents a single notification to be dispatched.
type Notification struct {
	ID              uuid.UUID              `json:"id"`
	OrganizationID  uuid.UUID              `json:"organization_id"`
	RuleID          *uuid.UUID             `json:"rule_id"`
	EventType       string                 `json:"event_type"`
	EventPayload    map[string]interface{} `json:"event_payload"`
	RecipientUserID uuid.UUID              `json:"recipient_user_id"`
	ChannelType     ChannelType            `json:"channel_type"`
	ChannelID       *uuid.UUID             `json:"channel_id"`
	Subject         string                 `json:"subject"`
	Body            string                 `json:"body"`
	Status          string                 `json:"status"`
	SentAt          *time.Time             `json:"sent_at,omitempty"`
	DeliveredAt     *time.Time             `json:"delivered_at,omitempty"`
	ReadAt          *time.Time             `json:"read_at,omitempty"`
	ErrorMessage    string                 `json:"error_message,omitempty"`
	RetryCount      int                    `json:"retry_count"`
	MaxRetries      int                    `json:"max_retries"`
	NextRetryAt     *time.Time             `json:"next_retry_at,omitempty"`
	Metadata        map[string]interface{} `json:"metadata"`
	CreatedAt       time.Time              `json:"created_at"`
}

// ============================================================
// RATE LIMITER
// ============================================================

// rateLimiter tracks notification counts per user to enforce rate limits.
type rateLimiter struct {
	mu      sync.Mutex
	counts  map[uuid.UUID]*userCounter
	maxRate int // max notifications per user per window
	window  time.Duration
}

type userCounter struct {
	count   int
	resetAt time.Time
}

func newRateLimiter(maxRate int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		counts:  make(map[uuid.UUID]*userCounter),
		maxRate: maxRate,
		window:  window,
	}
}

// Allow returns true if the user has not exceeded the rate limit.
func (rl *rateLimiter) Allow(userID uuid.UUID) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	counter, exists := rl.counts[userID]
	if !exists || now.After(counter.resetAt) {
		rl.counts[userID] = &userCounter{count: 1, resetAt: now.Add(rl.window)}
		return true
	}

	if counter.count >= rl.maxRate {
		return false
	}

	counter.count++
	return true
}

// ============================================================
// NOTIFICATION ENGINE
// ============================================================

// NotificationEngine is the core orchestrator that listens for events,
// evaluates rules, renders templates, and dispatches notifications.
type NotificationEngine struct {
	db          *pgxpool.Pool
	bus         *EventBus
	channels    map[ChannelType]NotificationChannel
	rateLimiter *rateLimiter
	mu          sync.RWMutex
}

// NewNotificationEngine creates a new engine with the given database pool and event bus.
func NewNotificationEngine(db *pgxpool.Pool, bus *EventBus) *NotificationEngine {
	return &NotificationEngine{
		db:          db,
		bus:         bus,
		channels:    make(map[ChannelType]NotificationChannel),
		rateLimiter: newRateLimiter(100, 1*time.Hour),
	}
}

// RegisterChannel registers a delivery channel with the engine.
func (engine *NotificationEngine) RegisterChannel(ch NotificationChannel) {
	engine.mu.Lock()
	defer engine.mu.Unlock()
	engine.channels[ch.Type()] = ch
	log.Info().Str("channel", string(ch.Type())).Msg("Notification channel registered")
}

// Emit publishes an event to the event bus.
func (engine *NotificationEngine) Emit(event Event) {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	engine.bus.Publish(event)
	log.Debug().
		Str("event_type", event.Type).
		Str("org_id", event.OrgID.String()).
		Str("severity", event.Severity).
		Msg("Event emitted")
}

// Start begins the event listener loop. It blocks until the context is cancelled.
func (engine *NotificationEngine) Start(ctx context.Context) {
	events := engine.bus.Subscribe()
	log.Info().Msg("Notification engine started")

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Notification engine shutting down")
			return
		case event, ok := <-events:
			if !ok {
				log.Info().Msg("Event bus channel closed")
				return
			}
			engine.handleEvent(ctx, event)
		}
	}
}

// handleEvent processes a single event: evaluate rules, render, dispatch.
func (engine *NotificationEngine) handleEvent(ctx context.Context, event Event) {
	logger := log.With().
		Str("event_type", event.Type).
		Str("org_id", event.OrgID.String()).
		Logger()

	logger.Debug().Msg("Processing event")

	rules, err := engine.evaluateRules(ctx, event)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to evaluate notification rules")
		return
	}

	if len(rules) == 0 {
		logger.Debug().Msg("No matching rules for event")
		return
	}

	for _, rule := range rules {
		if err := engine.processRule(ctx, event, &rule); err != nil {
			logger.Error().Err(err).Str("rule_id", rule.ID.String()).Msg("Failed to process rule")
		}
	}
}

// processRule handles a single matched rule: resolve recipients, render template, dispatch.
func (engine *NotificationEngine) processRule(ctx context.Context, event Event, rule *NotificationRule) error {
	// Load template
	tmpl, err := engine.loadTemplate(ctx, rule, event.Type, event.OrgID)
	if err != nil {
		return fmt.Errorf("load template: %w", err)
	}

	// Resolve recipient user IDs
	recipients, err := engine.resolveRecipients(ctx, rule, event.OrgID)
	if err != nil {
		return fmt.Errorf("resolve recipients: %w", err)
	}

	for _, userID := range recipients {
		// Rate limit check
		if !engine.rateLimiter.Allow(userID) {
			log.Warn().
				Str("user_id", userID.String()).
				Str("event_type", event.Type).
				Msg("Rate limit exceeded for user, skipping notification")
			continue
		}

		// Check user preferences
		prefs, _ := engine.getUserPreferences(ctx, userID, event.Type)

		// Determine which channels to use
		channelTypes := engine.resolveChannelTypes(rule, prefs)

		for _, chType := range channelTypes {
			notification, err := engine.buildNotification(event, rule, tmpl, userID, chType)
			if err != nil {
				log.Error().Err(err).
					Str("channel", string(chType)).
					Str("user_id", userID.String()).
					Msg("Failed to build notification")
				continue
			}

			// Persist to DB
			if err := engine.persistNotification(ctx, notification); err != nil {
				log.Error().Err(err).Msg("Failed to persist notification")
				continue
			}

			// Dispatch
			if err := engine.dispatch(ctx, notification); err != nil {
				log.Error().Err(err).Msg("Failed to dispatch notification")
				engine.trackDelivery(ctx, notification.ID, "failed", err.Error())
				continue
			}

			engine.trackDelivery(ctx, notification.ID, "sent", "")
		}
	}

	return nil
}

// evaluateRules finds active rules that match the given event.
func (engine *NotificationEngine) evaluateRules(ctx context.Context, event Event) ([]NotificationRule, error) {
	rows, err := engine.db.Query(ctx, `
		SELECT id, organization_id, name, event_type, severity_filter, conditions,
			   channel_ids, recipient_type, recipient_ids, template_id, is_active,
			   cooldown_minutes, escalation_after_minutes, escalation_channel_ids
		FROM notification_rules
		WHERE organization_id = $1
		  AND event_type = $2
		  AND is_active = true`,
		event.OrgID, event.Type,
	)
	if err != nil {
		return nil, fmt.Errorf("query rules: %w", err)
	}
	defer rows.Close()

	var rules []NotificationRule
	for rows.Next() {
		var r NotificationRule
		var conditionsJSON []byte
		var severityFilter []string
		var channelIDs, recipientIDs, escalationChannelIDs []uuid.UUID
		var templateID *uuid.UUID

		if err := rows.Scan(
			&r.ID, &r.OrganizationID, &r.Name, &r.EventType,
			&severityFilter, &conditionsJSON,
			&channelIDs, &r.RecipientType, &recipientIDs,
			&templateID, &r.IsActive,
			&r.CooldownMinutes, &r.EscalationAfterMinutes, &escalationChannelIDs,
		); err != nil {
			return nil, fmt.Errorf("scan rule: %w", err)
		}

		r.SeverityFilter = severityFilter
		r.ChannelIDs = channelIDs
		r.RecipientIDs = recipientIDs
		r.TemplateID = templateID
		r.EscalationChannelIDs = escalationChannelIDs

		if len(conditionsJSON) > 0 {
			json.Unmarshal(conditionsJSON, &r.Conditions)
		}

		// Severity filter: if specified, event severity must match
		if len(r.SeverityFilter) > 0 && event.Severity != "" {
			matched := false
			for _, s := range r.SeverityFilter {
				if s == event.Severity {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}

		// Cooldown check: ensure we haven't sent this recently
		if r.CooldownMinutes > 0 {
			var count int64
			engine.db.QueryRow(ctx, `
				SELECT COUNT(*) FROM notifications
				WHERE organization_id = $1 AND rule_id = $2
				  AND created_at > NOW() - ($3 || ' minutes')::interval`,
				event.OrgID, r.ID, fmt.Sprintf("%d", r.CooldownMinutes),
			).Scan(&count)
			if count > 0 {
				continue
			}
		}

		rules = append(rules, r)
	}

	return rules, nil
}

// loadTemplate retrieves the template for a rule, falling back to system templates.
func (engine *NotificationEngine) loadTemplate(ctx context.Context, rule *NotificationRule, eventType string, orgID uuid.UUID) (*NotificationTemplate, error) {
	var tmpl NotificationTemplate
	var slackJSON, webhookJSON []byte

	// Try rule-specific template first
	if rule.TemplateID != nil {
		err := engine.db.QueryRow(ctx, `
			SELECT id, organization_id, name, event_type, subject_template,
				   body_html_template, body_text_template, in_app_title_template,
				   in_app_body_template, slack_template, webhook_payload_template,
				   variables, is_system
			FROM notification_templates
			WHERE id = $1`,
			*rule.TemplateID,
		).Scan(
			&tmpl.ID, &tmpl.OrganizationID, &tmpl.Name, &tmpl.EventType,
			&tmpl.SubjectTemplate, &tmpl.BodyHTMLTemplate, &tmpl.BodyTextTemplate,
			&tmpl.InAppTitleTemplate, &tmpl.InAppBodyTemplate,
			&slackJSON, &webhookJSON,
			&tmpl.Variables, &tmpl.IsSystem,
		)
		if err == nil {
			json.Unmarshal(slackJSON, &tmpl.SlackTemplate)
			json.Unmarshal(webhookJSON, &tmpl.WebhookPayloadTemplate)
			return &tmpl, nil
		}
	}

	// Fall back to system template for this event type
	err := engine.db.QueryRow(ctx, `
		SELECT id, organization_id, name, event_type, subject_template,
			   body_html_template, body_text_template, in_app_title_template,
			   in_app_body_template, slack_template, webhook_payload_template,
			   variables, is_system
		FROM notification_templates
		WHERE event_type = $1 AND is_system = true
		ORDER BY created_at ASC
		LIMIT 1`,
		eventType,
	).Scan(
		&tmpl.ID, &tmpl.OrganizationID, &tmpl.Name, &tmpl.EventType,
		&tmpl.SubjectTemplate, &tmpl.BodyHTMLTemplate, &tmpl.BodyTextTemplate,
		&tmpl.InAppTitleTemplate, &tmpl.InAppBodyTemplate,
		&slackJSON, &webhookJSON,
		&tmpl.Variables, &tmpl.IsSystem,
	)
	if err != nil {
		return nil, fmt.Errorf("no template found for event %s: %w", eventType, err)
	}

	json.Unmarshal(slackJSON, &tmpl.SlackTemplate)
	json.Unmarshal(webhookJSON, &tmpl.WebhookPayloadTemplate)
	return &tmpl, nil
}

// resolveRecipients determines the user IDs that should receive the notification.
func (engine *NotificationEngine) resolveRecipients(ctx context.Context, rule *NotificationRule, orgID uuid.UUID) ([]uuid.UUID, error) {
	switch rule.RecipientType {
	case "user":
		return rule.RecipientIDs, nil

	case "role":
		var userIDs []uuid.UUID
		rows, err := engine.db.Query(ctx, `
			SELECT DISTINCT ur.user_id FROM user_roles ur
			JOIN roles r ON ur.role_id = r.id
			WHERE r.id = ANY($1) AND ur.user_id IN (
				SELECT id FROM users WHERE organization_id = $2 AND status = 'active'
			)`, rule.RecipientIDs, orgID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			var id uuid.UUID
			if err := rows.Scan(&id); err == nil {
				userIDs = append(userIDs, id)
			}
		}
		return userIDs, nil

	case "all_admins":
		var userIDs []uuid.UUID
		rows, err := engine.db.Query(ctx, `
			SELECT DISTINCT ur.user_id FROM user_roles ur
			JOIN roles r ON ur.role_id = r.id
			WHERE r.slug = 'org_admin' AND ur.user_id IN (
				SELECT id FROM users WHERE organization_id = $1 AND status = 'active'
			)`, orgID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			var id uuid.UUID
			if err := rows.Scan(&id); err == nil {
				userIDs = append(userIDs, id)
			}
		}
		return userIDs, nil

	case "all_users":
		var userIDs []uuid.UUID
		rows, err := engine.db.Query(ctx, `
			SELECT id FROM users
			WHERE organization_id = $1 AND status = 'active'`, orgID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			var id uuid.UUID
			if err := rows.Scan(&id); err == nil {
				userIDs = append(userIDs, id)
			}
		}
		return userIDs, nil

	case "custom":
		return rule.RecipientIDs, nil

	default:
		return rule.RecipientIDs, nil
	}
}

// getUserPreferences loads user notification preferences for a specific event type.
func (engine *NotificationEngine) getUserPreferences(ctx context.Context, userID uuid.UUID, eventType string) (*UserNotificationPreferences, error) {
	prefs := &UserNotificationPreferences{
		EmailEnabled: true,
		InAppEnabled: true,
		SlackEnabled: false,
	}

	err := engine.db.QueryRow(ctx, `
		SELECT email_enabled, in_app_enabled, slack_enabled, digest_frequency,
			   quiet_hours_start, quiet_hours_end, quiet_hours_timezone
		FROM notification_preferences
		WHERE user_id = $1 AND event_type = $2`,
		userID, eventType,
	).Scan(
		&prefs.EmailEnabled, &prefs.InAppEnabled, &prefs.SlackEnabled,
		&prefs.DigestFrequency, &prefs.QuietHoursStart, &prefs.QuietHoursEnd,
		&prefs.QuietHoursTimezone,
	)
	if err != nil {
		// Return defaults if no preferences set
		return prefs, nil
	}

	return prefs, nil
}

// UserNotificationPreferences holds a user's notification preferences.
type UserNotificationPreferences struct {
	EmailEnabled       bool    `json:"email_enabled"`
	InAppEnabled       bool    `json:"in_app_enabled"`
	SlackEnabled       bool    `json:"slack_enabled"`
	DigestFrequency    *string `json:"digest_frequency"`
	QuietHoursStart    *string `json:"quiet_hours_start"`
	QuietHoursEnd      *string `json:"quiet_hours_end"`
	QuietHoursTimezone *string `json:"quiet_hours_timezone"`
}

// resolveChannelTypes determines which channel types to use based on the rule and user prefs.
func (engine *NotificationEngine) resolveChannelTypes(rule *NotificationRule, prefs *UserNotificationPreferences) []ChannelType {
	var types []ChannelType

	// Always include in_app unless user disabled it
	if prefs == nil || prefs.InAppEnabled {
		types = append(types, ChannelInApp)
	}

	// Email unless disabled
	if prefs == nil || prefs.EmailEnabled {
		types = append(types, ChannelEmail)
	}

	// Slack if enabled
	if prefs != nil && prefs.SlackEnabled {
		types = append(types, ChannelSlack)
	}

	return types
}

// buildNotification renders templates and creates a notification record.
func (engine *NotificationEngine) buildNotification(
	event Event,
	rule *NotificationRule,
	tmpl *NotificationTemplate,
	userID uuid.UUID,
	chType ChannelType,
) (*Notification, error) {
	var subject, body string
	var err error

	switch chType {
	case ChannelEmail:
		subject, err = RenderTemplate(tmpl.SubjectTemplate, event.Payload)
		if err != nil {
			return nil, fmt.Errorf("render subject: %w", err)
		}
		body, err = RenderTemplate(tmpl.BodyHTMLTemplate, event.Payload)
		if err != nil {
			return nil, fmt.Errorf("render body: %w", err)
		}
	case ChannelInApp:
		subject, err = RenderTemplate(tmpl.InAppTitleTemplate, event.Payload)
		if err != nil {
			return nil, fmt.Errorf("render in-app title: %w", err)
		}
		body, err = RenderTemplate(tmpl.InAppBodyTemplate, event.Payload)
		if err != nil {
			return nil, fmt.Errorf("render in-app body: %w", err)
		}
	case ChannelSlack:
		subject = event.Type
		slackJSON, _ := json.Marshal(tmpl.SlackTemplate)
		body, err = RenderTemplate(string(slackJSON), event.Payload)
		if err != nil {
			return nil, fmt.Errorf("render slack: %w", err)
		}
	case ChannelWebhook:
		subject = event.Type
		whJSON, _ := json.Marshal(tmpl.WebhookPayloadTemplate)
		body, err = RenderTemplate(string(whJSON), event.Payload)
		if err != nil {
			return nil, fmt.Errorf("render webhook: %w", err)
		}
	}

	n := &Notification{
		ID:              uuid.New(),
		OrganizationID:  event.OrgID,
		RuleID:          &rule.ID,
		EventType:       event.Type,
		EventPayload:    event.Payload,
		RecipientUserID: userID,
		ChannelType:     chType,
		Subject:         subject,
		Body:            body,
		Status:          "pending",
		RetryCount:      0,
		MaxRetries:      3,
		Metadata:        make(map[string]interface{}),
		CreatedAt:       time.Now(),
	}

	return n, nil
}

// RenderTemplate renders a Go text/template string with the given data.
// Exported for use by channels and tests.
func RenderTemplate(tmplStr string, data interface{}) (string, error) {
	if tmplStr == "" {
		return "", nil
	}

	t, err := template.New("notification").Option("missingkey=zero").Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	return buf.String(), nil
}

// persistNotification stores a notification record in the database.
func (engine *NotificationEngine) persistNotification(ctx context.Context, n *Notification) error {
	payloadJSON, _ := json.Marshal(n.EventPayload)
	metadataJSON, _ := json.Marshal(n.Metadata)

	_, err := engine.db.Exec(ctx, `
		INSERT INTO notifications (
			id, organization_id, rule_id, event_type, event_payload,
			recipient_user_id, channel_type, channel_id, subject, body,
			status, retry_count, max_retries, metadata, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`,
		n.ID, n.OrganizationID, n.RuleID, n.EventType, payloadJSON,
		n.RecipientUserID, string(n.ChannelType), n.ChannelID, n.Subject, n.Body,
		n.Status, n.RetryCount, n.MaxRetries, metadataJSON, n.CreatedAt,
	)
	return err
}

// dispatch sends a notification via the appropriate channel.
func (engine *NotificationEngine) dispatch(ctx context.Context, n *Notification) error {
	engine.mu.RLock()
	ch, exists := engine.channels[n.ChannelType]
	engine.mu.RUnlock()

	if !exists {
		return fmt.Errorf("no channel registered for type: %s", n.ChannelType)
	}

	return ch.Send(ctx, n)
}

// trackDelivery updates the notification status in the database.
func (engine *NotificationEngine) trackDelivery(ctx context.Context, notificationID uuid.UUID, status, errMsg string) {
	now := time.Now()

	switch status {
	case "sent":
		engine.db.Exec(ctx, `
			UPDATE notifications SET status = 'sent', sent_at = $1 WHERE id = $2`,
			now, notificationID)
	case "delivered":
		engine.db.Exec(ctx, `
			UPDATE notifications SET status = 'delivered', delivered_at = $1 WHERE id = $2`,
			now, notificationID)
	case "failed":
		engine.db.Exec(ctx, `
			UPDATE notifications
			SET status = 'failed', error_message = $1, retry_count = retry_count + 1,
				next_retry_at = NOW() + (retry_count * interval '5 minutes')
			WHERE id = $2`,
			errMsg, notificationID)
	case "bounced":
		engine.db.Exec(ctx, `
			UPDATE notifications SET status = 'bounced', error_message = $1 WHERE id = $2`,
			errMsg, notificationID)
	}
}
