package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/complianceforge/platform/internal/middleware"
	"github.com/complianceforge/platform/internal/models"
	"github.com/complianceforge/platform/internal/service"
)

// NotificationHandler handles HTTP requests for the notification system.
// It wraps the NotificationEngine for event-driven operations and a database
// pool for direct queries on notifications, preferences, rules, and channels.
type NotificationHandler struct {
	engine *service.NotificationEngine
	db     *pgxpool.Pool
}

// NewNotificationHandler creates a new handler for notification endpoints.
func NewNotificationHandler(engine *service.NotificationEngine, db *pgxpool.Pool) *NotificationHandler {
	return &NotificationHandler{engine: engine, db: db}
}

// ============================================================
// USER NOTIFICATION ENDPOINTS
// ============================================================

// ListNotifications returns the current user's in-app notifications with pagination.
// GET /api/v1/notifications
func (h *NotificationHandler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	params := parsePagination(r)

	// Get paginated notifications
	rows, err := h.db.Query(r.Context(), `
		SELECT id, organization_id, rule_id, event_type, event_payload,
			   recipient_user_id, channel_type, channel_id, subject, body,
			   status, sent_at, delivered_at, read_at, acknowledged_at,
			   metadata, created_at
		FROM notifications
		WHERE recipient_user_id = $1 AND organization_id = $2 AND channel_type = 'in_app'
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`,
		userID, orgID, params.PageSize, params.Offset(),
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve notifications")
		return
	}
	defer rows.Close()

	type notificationItem struct {
		ID             uuid.UUID       `json:"id"`
		EventType      string          `json:"event_type"`
		EventPayload   json.RawMessage `json:"event_payload"`
		Subject        string          `json:"subject"`
		Body           string          `json:"body"`
		Status         string          `json:"status"`
		ReadAt         *time.Time      `json:"read_at,omitempty"`
		AcknowledgedAt *time.Time      `json:"acknowledged_at,omitempty"`
		Metadata       json.RawMessage `json:"metadata"`
		CreatedAt      time.Time       `json:"created_at"`
	}

	var notifications []notificationItem
	for rows.Next() {
		var n notificationItem
		var orgIDScan, ruleID, recipientID, channelID *uuid.UUID
		var channelType, status string
		var sentAt, deliveredAt *time.Time

		if err := rows.Scan(
			&n.ID, &orgIDScan, &ruleID, &n.EventType, &n.EventPayload,
			&recipientID, &channelType, &channelID, &n.Subject, &n.Body,
			&status, &sentAt, &deliveredAt, &n.ReadAt, &n.AcknowledgedAt,
			&n.Metadata, &n.CreatedAt,
		); err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to scan notification")
			return
		}
		n.Status = status
		notifications = append(notifications, n)
	}

	if notifications == nil {
		notifications = []notificationItem{}
	}

	// Get total count
	var total int64
	h.db.QueryRow(r.Context(), `
		SELECT COUNT(*) FROM notifications
		WHERE recipient_user_id = $1 AND organization_id = $2 AND channel_type = 'in_app'`,
		userID, orgID,
	).Scan(&total)

	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize != 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: notifications,
		Pagination: models.Pagination{
			Page:       params.Page,
			PageSize:   params.PageSize,
			TotalItems: total,
			TotalPages: totalPages,
			HasNext:    params.Page < totalPages,
			HasPrev:    params.Page > 1,
		},
	})
}

// GetUnreadCount returns the count of unread in-app notifications for the bell badge.
// GET /api/v1/notifications/unread-count
func (h *NotificationHandler) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())

	var count int64
	err := h.db.QueryRow(r.Context(), `
		SELECT COUNT(*) FROM notifications
		WHERE recipient_user_id = $1 AND organization_id = $2
		  AND channel_type = 'in_app' AND read_at IS NULL`,
		userID, orgID,
	).Scan(&count)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get unread count")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]interface{}{"count": count},
	})
}

// MarkRead marks a single notification as read.
// PUT /api/v1/notifications/{id}/read
func (h *NotificationHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	orgID := middleware.GetOrgID(r.Context())
	notifID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid notification ID")
		return
	}

	result, err := h.db.Exec(r.Context(), `
		UPDATE notifications SET read_at = NOW()
		WHERE id = $1 AND recipient_user_id = $2 AND organization_id = $3 AND read_at IS NULL`,
		notifID, userID, orgID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to mark as read")
		return
	}

	if result.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Notification not found or already read")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]interface{}{"id": notifID, "read_at": time.Now()},
	})
}

// MarkAllRead marks all of the user's unread notifications as read.
// PUT /api/v1/notifications/read-all
func (h *NotificationHandler) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())

	result, err := h.db.Exec(r.Context(), `
		UPDATE notifications SET read_at = NOW()
		WHERE recipient_user_id = $1 AND organization_id = $2
		  AND channel_type = 'in_app' AND read_at IS NULL`,
		userID, orgID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to mark all as read")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"marked_count": result.RowsAffected(),
			"read_at":      time.Now(),
		},
	})
}

// ============================================================
// USER PREFERENCES
// ============================================================

// GetPreferences returns the current user's notification preferences.
// GET /api/v1/notifications/preferences
func (h *NotificationHandler) GetPreferences(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	orgID := middleware.GetOrgID(r.Context())

	rows, err := h.db.Query(r.Context(), `
		SELECT id, event_type, email_enabled, in_app_enabled, slack_enabled,
			   digest_frequency, quiet_hours_start, quiet_hours_end, quiet_hours_timezone
		FROM notification_preferences
		WHERE user_id = $1 AND organization_id = $2
		ORDER BY event_type`,
		userID, orgID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get preferences")
		return
	}
	defer rows.Close()

	type prefItem struct {
		ID                 uuid.UUID `json:"id"`
		EventType          string    `json:"event_type"`
		EmailEnabled       bool      `json:"email_enabled"`
		InAppEnabled       bool      `json:"in_app_enabled"`
		SlackEnabled       bool      `json:"slack_enabled"`
		DigestFrequency    string    `json:"digest_frequency"`
		QuietHoursStart    *string   `json:"quiet_hours_start,omitempty"`
		QuietHoursEnd      *string   `json:"quiet_hours_end,omitempty"`
		QuietHoursTimezone string    `json:"quiet_hours_timezone"`
	}

	var prefs []prefItem
	for rows.Next() {
		var p prefItem
		if err := rows.Scan(
			&p.ID, &p.EventType, &p.EmailEnabled, &p.InAppEnabled, &p.SlackEnabled,
			&p.DigestFrequency, &p.QuietHoursStart, &p.QuietHoursEnd, &p.QuietHoursTimezone,
		); err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to scan preferences")
			return
		}
		prefs = append(prefs, p)
	}

	if prefs == nil {
		prefs = []prefItem{}
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: prefs})
}

// UpdatePreferences updates the current user's notification preferences.
// PUT /api/v1/notifications/preferences
func (h *NotificationHandler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	orgID := middleware.GetOrgID(r.Context())

	var req struct {
		Preferences []struct {
			EventType          string  `json:"event_type"`
			EmailEnabled       bool    `json:"email_enabled"`
			InAppEnabled       bool    `json:"in_app_enabled"`
			SlackEnabled       bool    `json:"slack_enabled"`
			DigestFrequency    string  `json:"digest_frequency"`
			QuietHoursStart    *string `json:"quiet_hours_start"`
			QuietHoursEnd      *string `json:"quiet_hours_end"`
			QuietHoursTimezone string  `json:"quiet_hours_timezone"`
		} `json:"preferences"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	for _, pref := range req.Preferences {
		if pref.EventType == "" {
			continue
		}
		if pref.QuietHoursTimezone == "" {
			pref.QuietHoursTimezone = "UTC"
		}
		if pref.DigestFrequency == "" {
			pref.DigestFrequency = "realtime"
		}

		_, err := h.db.Exec(r.Context(), `
			INSERT INTO notification_preferences (
				user_id, organization_id, event_type, email_enabled, in_app_enabled,
				slack_enabled, digest_frequency, quiet_hours_start, quiet_hours_end,
				quiet_hours_timezone
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			ON CONFLICT (user_id, event_type) DO UPDATE SET
				email_enabled = EXCLUDED.email_enabled,
				in_app_enabled = EXCLUDED.in_app_enabled,
				slack_enabled = EXCLUDED.slack_enabled,
				digest_frequency = EXCLUDED.digest_frequency,
				quiet_hours_start = EXCLUDED.quiet_hours_start,
				quiet_hours_end = EXCLUDED.quiet_hours_end,
				quiet_hours_timezone = EXCLUDED.quiet_hours_timezone`,
			userID, orgID, pref.EventType, pref.EmailEnabled, pref.InAppEnabled,
			pref.SlackEnabled, pref.DigestFrequency, pref.QuietHoursStart, pref.QuietHoursEnd,
			pref.QuietHoursTimezone,
		)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update preferences")
			return
		}
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Preferences updated successfully"},
	})
}

// ============================================================
// ADMIN: NOTIFICATION RULES
// ============================================================

// ListRules returns the organization's notification rules with pagination.
// GET /api/v1/notifications/rules
func (h *NotificationHandler) ListRules(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	params := parsePagination(r)

	rows, err := h.db.Query(r.Context(), `
		SELECT id, name, event_type, severity_filter, conditions, channel_ids,
			   recipient_type, recipient_ids, template_id, is_active,
			   cooldown_minutes, escalation_after_minutes, escalation_channel_ids,
			   created_at, updated_at
		FROM notification_rules
		WHERE organization_id = $1
		ORDER BY event_type, name
		LIMIT $2 OFFSET $3`,
		orgID, params.PageSize, params.Offset(),
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve rules")
		return
	}
	defer rows.Close()

	type ruleItem struct {
		ID                     uuid.UUID       `json:"id"`
		Name                   string          `json:"name"`
		EventType              string          `json:"event_type"`
		SeverityFilter         []string        `json:"severity_filter"`
		Conditions             json.RawMessage `json:"conditions"`
		ChannelIDs             []uuid.UUID     `json:"channel_ids"`
		RecipientType          string          `json:"recipient_type"`
		RecipientIDs           []uuid.UUID     `json:"recipient_ids"`
		TemplateID             *uuid.UUID      `json:"template_id,omitempty"`
		IsActive               bool            `json:"is_active"`
		CooldownMinutes        int             `json:"cooldown_minutes"`
		EscalationAfterMinutes int             `json:"escalation_after_minutes"`
		EscalationChannelIDs   []uuid.UUID     `json:"escalation_channel_ids"`
		CreatedAt              time.Time       `json:"created_at"`
		UpdatedAt              time.Time       `json:"updated_at"`
	}

	var rules []ruleItem
	for rows.Next() {
		var rule ruleItem
		if err := rows.Scan(
			&rule.ID, &rule.Name, &rule.EventType, &rule.SeverityFilter,
			&rule.Conditions, &rule.ChannelIDs, &rule.RecipientType,
			&rule.RecipientIDs, &rule.TemplateID, &rule.IsActive,
			&rule.CooldownMinutes, &rule.EscalationAfterMinutes, &rule.EscalationChannelIDs,
			&rule.CreatedAt, &rule.UpdatedAt,
		); err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to scan rule")
			return
		}
		rules = append(rules, rule)
	}

	if rules == nil {
		rules = []ruleItem{}
	}

	// Get total count
	var total int64
	h.db.QueryRow(r.Context(), `
		SELECT COUNT(*) FROM notification_rules WHERE organization_id = $1`,
		orgID,
	).Scan(&total)

	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize != 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: rules,
		Pagination: models.Pagination{
			Page:       params.Page,
			PageSize:   params.PageSize,
			TotalItems: total,
			TotalPages: totalPages,
			HasNext:    params.Page < totalPages,
			HasPrev:    params.Page > 1,
		},
	})
}

// CreateRule creates a new notification rule.
// POST /api/v1/notifications/rules
func (h *NotificationHandler) CreateRule(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var req struct {
		Name                   string          `json:"name"`
		EventType              string          `json:"event_type"`
		SeverityFilter         []string        `json:"severity_filter"`
		Conditions             json.RawMessage `json:"conditions"`
		ChannelIDs             []uuid.UUID     `json:"channel_ids"`
		RecipientType          string          `json:"recipient_type"`
		RecipientIDs           []uuid.UUID     `json:"recipient_ids"`
		TemplateID             *uuid.UUID      `json:"template_id"`
		IsActive               bool            `json:"is_active"`
		CooldownMinutes        int             `json:"cooldown_minutes"`
		EscalationAfterMinutes int             `json:"escalation_after_minutes"`
		EscalationChannelIDs   []uuid.UUID     `json:"escalation_channel_ids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Name == "" || req.EventType == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "name and event_type are required")
		return
	}

	if req.Conditions == nil {
		req.Conditions = json.RawMessage("{}")
	}

	id := uuid.New()
	_, err := h.db.Exec(r.Context(), `
		INSERT INTO notification_rules (
			id, organization_id, name, event_type, severity_filter, conditions,
			channel_ids, recipient_type, recipient_ids, template_id, is_active,
			cooldown_minutes, escalation_after_minutes, escalation_channel_ids
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`,
		id, orgID, req.Name, req.EventType, req.SeverityFilter, req.Conditions,
		req.ChannelIDs, req.RecipientType, req.RecipientIDs, req.TemplateID,
		req.IsActive, req.CooldownMinutes, req.EscalationAfterMinutes, req.EscalationChannelIDs,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create rule")
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{
		Success: true,
		Data:    map[string]interface{}{"id": id, "message": "Notification rule created"},
	})
}

// UpdateRule updates an existing notification rule.
// PUT /api/v1/notifications/rules/{id}
func (h *NotificationHandler) UpdateRule(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	ruleID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid rule ID")
		return
	}

	var req struct {
		Name                   *string          `json:"name"`
		EventType              *string          `json:"event_type"`
		SeverityFilter         *[]string        `json:"severity_filter"`
		Conditions             *json.RawMessage `json:"conditions"`
		ChannelIDs             *[]uuid.UUID     `json:"channel_ids"`
		RecipientType          *string          `json:"recipient_type"`
		RecipientIDs           *[]uuid.UUID     `json:"recipient_ids"`
		TemplateID             *uuid.UUID       `json:"template_id"`
		IsActive               *bool            `json:"is_active"`
		CooldownMinutes        *int             `json:"cooldown_minutes"`
		EscalationAfterMinutes *int             `json:"escalation_after_minutes"`
		EscalationChannelIDs   *[]uuid.UUID     `json:"escalation_channel_ids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	result, err := h.db.Exec(r.Context(), `
		UPDATE notification_rules SET
			name = COALESCE($3, name),
			event_type = COALESCE($4, event_type),
			severity_filter = COALESCE($5, severity_filter),
			conditions = COALESCE($6, conditions),
			channel_ids = COALESCE($7, channel_ids),
			recipient_type = COALESCE($8, recipient_type),
			recipient_ids = COALESCE($9, recipient_ids),
			template_id = $10,
			is_active = COALESCE($11, is_active),
			cooldown_minutes = COALESCE($12, cooldown_minutes),
			escalation_after_minutes = COALESCE($13, escalation_after_minutes),
			escalation_channel_ids = COALESCE($14, escalation_channel_ids),
			updated_at = NOW()
		WHERE id = $1 AND organization_id = $2`,
		ruleID, orgID,
		req.Name, req.EventType, req.SeverityFilter, req.Conditions,
		req.ChannelIDs, req.RecipientType, req.RecipientIDs, req.TemplateID,
		req.IsActive, req.CooldownMinutes, req.EscalationAfterMinutes, req.EscalationChannelIDs,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update rule")
		return
	}

	if result.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Rule not found")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Notification rule updated"},
	})
}

// DeleteRule removes a notification rule.
// DELETE /api/v1/notifications/rules/{id}
func (h *NotificationHandler) DeleteRule(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	ruleID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid rule ID")
		return
	}

	result, err := h.db.Exec(r.Context(), `
		DELETE FROM notification_rules WHERE id = $1 AND organization_id = $2`,
		ruleID, orgID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete rule")
		return
	}

	if result.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Rule not found")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Notification rule deleted"},
	})
}

// ============================================================
// ADMIN: NOTIFICATION CHANNELS
// ============================================================

// ListChannels returns the organization's configured notification channels.
// GET /api/v1/notifications/channels
func (h *NotificationHandler) ListChannels(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	rows, err := h.db.Query(r.Context(), `
		SELECT id, channel_type, name, configuration, is_active, is_default, created_at, updated_at
		FROM notification_channels
		WHERE organization_id = $1 AND deleted_at IS NULL
		ORDER BY channel_type, name`,
		orgID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve channels")
		return
	}
	defer rows.Close()

	type channelItem struct {
		ID            uuid.UUID       `json:"id"`
		ChannelType   string          `json:"channel_type"`
		Name          string          `json:"name"`
		Configuration json.RawMessage `json:"configuration"`
		IsActive      bool            `json:"is_active"`
		IsDefault     bool            `json:"is_default"`
		CreatedAt     time.Time       `json:"created_at"`
		UpdatedAt     time.Time       `json:"updated_at"`
	}

	var channels []channelItem
	for rows.Next() {
		var ch channelItem
		if err := rows.Scan(
			&ch.ID, &ch.ChannelType, &ch.Name, &ch.Configuration,
			&ch.IsActive, &ch.IsDefault, &ch.CreatedAt, &ch.UpdatedAt,
		); err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to scan channel")
			return
		}
		channels = append(channels, ch)
	}

	if channels == nil {
		channels = []channelItem{}
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: channels})
}

// CreateChannel creates a new notification channel configuration.
// POST /api/v1/notifications/channels
func (h *NotificationHandler) CreateChannel(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var req struct {
		ChannelType   string          `json:"channel_type"`
		Name          string          `json:"name"`
		Configuration json.RawMessage `json:"configuration"`
		IsActive      bool            `json:"is_active"`
		IsDefault     bool            `json:"is_default"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Name == "" || req.ChannelType == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "name and channel_type are required")
		return
	}

	validTypes := map[string]bool{
		"email": true, "in_app": true, "slack": true, "webhook": true, "sms": true,
	}
	if !validTypes[req.ChannelType] {
		writeError(w, http.StatusBadRequest, "INVALID_TYPE", "channel_type must be email, in_app, slack, webhook, or sms")
		return
	}

	if req.Configuration == nil {
		req.Configuration = json.RawMessage("{}")
	}

	id := uuid.New()
	_, err := h.db.Exec(r.Context(), `
		INSERT INTO notification_channels (id, organization_id, channel_type, name, configuration, is_active, is_default)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		id, orgID, req.ChannelType, req.Name, req.Configuration, req.IsActive, req.IsDefault,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create channel")
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{
		Success: true,
		Data:    map[string]interface{}{"id": id, "message": "Notification channel created"},
	})
}

// TestChannel sends a test notification through a channel.
// POST /api/v1/notifications/channels/{id}/test
func (h *NotificationHandler) TestChannel(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	channelID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid channel ID")
		return
	}

	// Verify channel exists and belongs to org
	var channelType, channelName string
	err = h.db.QueryRow(r.Context(), `
		SELECT channel_type, name FROM notification_channels
		WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL`,
		channelID, orgID,
	).Scan(&channelType, &channelName)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Channel not found")
		return
	}

	// Create a test notification in the database
	testID := uuid.New()
	_, err = h.db.Exec(r.Context(), `
		INSERT INTO notifications (
			id, organization_id, event_type, event_payload,
			recipient_user_id, channel_type, channel_id,
			subject, body, status, metadata
		) VALUES ($1, $2, 'system.test', '{"test": true}', $3, $4, $5,
			'ComplianceForge Test Notification',
			'This is a test notification from ComplianceForge to verify your notification channel configuration.',
			'sent', '{"is_test": true}')`,
		testID, orgID, userID, channelType, channelID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create test notification")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"notification_id": testID,
			"channel_id":      channelID,
			"channel_type":    channelType,
			"channel_name":    channelName,
			"message":         "Test notification sent",
		},
	})
}
