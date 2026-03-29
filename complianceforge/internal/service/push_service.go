// Package service provides the Push Notification Service for ComplianceForge.
// It manages device tokens, enforces notification preferences and quiet hours,
// and dispatches push notifications across iOS (APNs), Android (FCM), and Web platforms.
package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
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
// NOTIFICATION TYPES
// ============================================================

const (
	PushTypeBreachAlert      = "breach_alert"
	PushTypeApprovalRequest  = "approval_request"
	PushTypeIncidentCreated  = "incident_created"
	PushTypeDeadlineReminder = "deadline_reminder"
	PushTypeMention          = "mention"
	PushTypeComment          = "comment"
)

// MaxTokensPerUser is the maximum number of active push tokens per user.
const MaxTokensPerUser = 5

// ============================================================
// MODELS
// ============================================================

// PushToken represents a registered device push token.
type PushToken struct {
	ID             uuid.UUID `json:"id"`
	UserID         uuid.UUID `json:"user_id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	Platform       string    `json:"platform"`
	Token          string    `json:"token"`
	TokenHash      string    `json:"token_hash"`
	DeviceName     string    `json:"device_name"`
	DeviceModel    string    `json:"device_model"`
	OSVersion      string    `json:"os_version"`
	AppVersion     string    `json:"app_version"`
	IsActive       bool      `json:"is_active"`
	LastUsedAt     time.Time `json:"last_used_at"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// DeviceInfo contains device metadata provided during token registration.
type DeviceInfo struct {
	DeviceName  string `json:"device_name"`
	DeviceModel string `json:"device_model"`
	OSVersion   string `json:"os_version"`
	AppVersion  string `json:"app_version"`
}

// PushNotification is the payload to send via push.
type PushNotification struct {
	Type     string                 `json:"type"`
	Title    string                 `json:"title"`
	Body     string                 `json:"body"`
	Data     map[string]interface{} `json:"data,omitempty"`
	Priority string                 `json:"priority,omitempty"` // "high" or "normal"
	Badge    *int                   `json:"badge,omitempty"`
	Sound    string                 `json:"sound,omitempty"`
}

// MobilePreferences holds a user's push notification preferences.
type MobilePreferences struct {
	ID                    uuid.UUID `json:"id"`
	UserID                uuid.UUID `json:"user_id"`
	OrganizationID        uuid.UUID `json:"organization_id"`
	PushEnabled           bool      `json:"push_enabled"`
	PushBreachAlerts      bool      `json:"push_breach_alerts"`
	PushApprovalRequests  bool      `json:"push_approval_requests"`
	PushIncidentAlerts    bool      `json:"push_incident_alerts"`
	PushDeadlineReminders bool      `json:"push_deadline_reminders"`
	PushMentions          bool      `json:"push_mentions"`
	PushComments          bool      `json:"push_comments"`
	QuietHoursEnabled     bool      `json:"quiet_hours_enabled"`
	QuietHoursStart       string    `json:"quiet_hours_start"`
	QuietHoursEnd         string    `json:"quiet_hours_end"`
	QuietHoursTimezone    string    `json:"quiet_hours_timezone"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

// PushLogEntry represents a logged push notification delivery attempt.
type PushLogEntry struct {
	ID               uuid.UUID              `json:"id"`
	UserID           uuid.UUID              `json:"user_id"`
	TokenID          *uuid.UUID             `json:"token_id,omitempty"`
	NotificationType string                 `json:"notification_type"`
	Title            string                 `json:"title"`
	Body             string                 `json:"body"`
	Data             map[string]interface{} `json:"data,omitempty"`
	Status           string                 `json:"status"`
	Platform         string                 `json:"platform"`
	SentAt           time.Time              `json:"sent_at"`
	ErrorMessage     string                 `json:"error_message,omitempty"`
}

// ============================================================
// PUSH PROVIDER INTERFACE (abstracted for FCM/APNs/Web Push)
// ============================================================

// PushProvider is the interface for platform-specific push delivery.
// Implementations are injected for FCM, APNs, and Web Push.
type PushProvider interface {
	// Send delivers a push notification to a single device token.
	// Returns an error if delivery fails. If the token is invalid,
	// the error message should contain "invalid" or "unregistered".
	Send(ctx context.Context, platform, token string, notification *PushNotification) error
}

// StubPushProvider is a no-op provider used during development.
// Replace with FCMProvider / APNsProvider / WebPushProvider in production.
type StubPushProvider struct{}

// Send logs the push attempt but does not actually deliver.
func (s *StubPushProvider) Send(ctx context.Context, platform, token string, notification *PushNotification) error {
	log.Debug().
		Str("platform", platform).
		Str("token_prefix", truncateToken(token)).
		Str("type", notification.Type).
		Str("title", notification.Title).
		Msg("Stub push provider: notification sent (no-op)")
	return nil
}

// ============================================================
// PUSH SERVICE
// ============================================================

// PushService manages push notification tokens, preferences, and delivery.
type PushService struct {
	db       *pgxpool.Pool
	provider PushProvider
}

// NewPushService creates a new PushService with the given database pool and push provider.
func NewPushService(db *pgxpool.Pool, provider PushProvider) *PushService {
	if provider == nil {
		provider = &StubPushProvider{}
	}
	return &PushService{
		db:       db,
		provider: provider,
	}
}

// ============================================================
// TOKEN MANAGEMENT
// ============================================================

// RegisterToken registers or updates a push notification token for a user.
// If the token already exists (by hash), it updates metadata and reactivates it.
// If the user has more than MaxTokensPerUser active tokens, the oldest are deactivated.
func (s *PushService) RegisterToken(ctx context.Context, orgID, userID uuid.UUID, platform, token string, info DeviceInfo) (*PushToken, error) {
	if token == "" {
		return nil, fmt.Errorf("token cannot be empty")
	}
	if platform != "ios" && platform != "android" && platform != "web" {
		return nil, fmt.Errorf("invalid platform: %s (must be ios, android, or web)", platform)
	}

	tokenHash := hashToken(token)

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Set RLS context
	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return nil, fmt.Errorf("set org context: %w", err)
	}

	// Upsert token
	var result PushToken
	err = tx.QueryRow(ctx, `
		INSERT INTO push_notification_tokens (
			user_id, organization_id, platform, token, token_hash,
			device_name, device_model, os_version, app_version,
			is_active, last_used_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, true, NOW())
		ON CONFLICT (token_hash) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			organization_id = EXCLUDED.organization_id,
			platform = EXCLUDED.platform,
			token = EXCLUDED.token,
			device_name = EXCLUDED.device_name,
			device_model = EXCLUDED.device_model,
			os_version = EXCLUDED.os_version,
			app_version = EXCLUDED.app_version,
			is_active = true,
			last_used_at = NOW(),
			updated_at = NOW()
		RETURNING id, user_id, organization_id, platform, token, token_hash,
			device_name, device_model, os_version, app_version,
			is_active, last_used_at, created_at, updated_at`,
		userID, orgID, platform, token, tokenHash,
		info.DeviceName, info.DeviceModel, info.OSVersion, info.AppVersion,
	).Scan(
		&result.ID, &result.UserID, &result.OrganizationID, &result.Platform,
		&result.Token, &result.TokenHash, &result.DeviceName, &result.DeviceModel,
		&result.OSVersion, &result.AppVersion, &result.IsActive, &result.LastUsedAt,
		&result.CreatedAt, &result.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("upsert token: %w", err)
	}

	// Count active tokens for user and deactivate oldest if > MaxTokensPerUser
	var activeCount int
	err = tx.QueryRow(ctx, `
		SELECT COUNT(*) FROM push_notification_tokens
		WHERE user_id = $1 AND is_active = true`,
		userID,
	).Scan(&activeCount)
	if err != nil {
		return nil, fmt.Errorf("count active tokens: %w", err)
	}

	if activeCount > MaxTokensPerUser {
		excess := activeCount - MaxTokensPerUser
		_, err = tx.Exec(ctx, `
			UPDATE push_notification_tokens SET is_active = false, updated_at = NOW()
			WHERE id IN (
				SELECT id FROM push_notification_tokens
				WHERE user_id = $1 AND is_active = true
				ORDER BY last_used_at ASC
				LIMIT $2
			)`,
			userID, excess,
		)
		if err != nil {
			return nil, fmt.Errorf("deactivate excess tokens: %w", err)
		}
		log.Info().
			Str("user_id", userID.String()).
			Int("deactivated", excess).
			Msg("Deactivated oldest push tokens (exceeded max per user)")
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	log.Info().
		Str("user_id", userID.String()).
		Str("platform", platform).
		Str("token_hash", tokenHash[:12]).
		Msg("Push token registered")

	return &result, nil
}

// UnregisterToken deactivates a push token by its hash.
func (s *PushService) UnregisterToken(ctx context.Context, orgID, userID uuid.UUID, tokenHash string) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return fmt.Errorf("set org context: %w", err)
	}

	result, err := tx.Exec(ctx, `
		UPDATE push_notification_tokens
		SET is_active = false, updated_at = NOW()
		WHERE user_id = $1 AND token_hash = $2 AND organization_id = $3`,
		userID, tokenHash, orgID,
	)
	if err != nil {
		return fmt.Errorf("unregister token: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("token not found")
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	log.Info().
		Str("user_id", userID.String()).
		Str("token_hash", tokenHash[:min(12, len(tokenHash))]).
		Msg("Push token unregistered")

	return nil
}

// GetActiveTokens returns all active push tokens for a user.
func (s *PushService) GetActiveTokens(ctx context.Context, orgID, userID uuid.UUID) ([]PushToken, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, user_id, organization_id, platform, token, token_hash,
			device_name, device_model, os_version, app_version,
			is_active, last_used_at, created_at, updated_at
		FROM push_notification_tokens
		WHERE user_id = $1 AND organization_id = $2 AND is_active = true
		ORDER BY last_used_at DESC`,
		userID, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("query active tokens: %w", err)
	}
	defer rows.Close()

	var tokens []PushToken
	for rows.Next() {
		var t PushToken
		if err := rows.Scan(
			&t.ID, &t.UserID, &t.OrganizationID, &t.Platform, &t.Token, &t.TokenHash,
			&t.DeviceName, &t.DeviceModel, &t.OSVersion, &t.AppVersion,
			&t.IsActive, &t.LastUsedAt, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan token: %w", err)
		}
		tokens = append(tokens, t)
	}

	return tokens, nil
}

// ============================================================
// PREFERENCES
// ============================================================

// GetPreferences returns the user's mobile push preferences, creating defaults if none exist.
func (s *PushService) GetPreferences(ctx context.Context, orgID, userID uuid.UUID) (*MobilePreferences, error) {
	var prefs MobilePreferences
	err := s.db.QueryRow(ctx, `
		SELECT id, user_id, organization_id, push_enabled, push_breach_alerts,
			push_approval_requests, push_incident_alerts, push_deadline_reminders,
			push_mentions, push_comments, quiet_hours_enabled,
			quiet_hours_start, quiet_hours_end, quiet_hours_timezone,
			created_at, updated_at
		FROM user_mobile_preferences
		WHERE user_id = $1 AND organization_id = $2`,
		userID, orgID,
	).Scan(
		&prefs.ID, &prefs.UserID, &prefs.OrganizationID,
		&prefs.PushEnabled, &prefs.PushBreachAlerts,
		&prefs.PushApprovalRequests, &prefs.PushIncidentAlerts,
		&prefs.PushDeadlineReminders, &prefs.PushMentions, &prefs.PushComments,
		&prefs.QuietHoursEnabled, &prefs.QuietHoursStart, &prefs.QuietHoursEnd,
		&prefs.QuietHoursTimezone, &prefs.CreatedAt, &prefs.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		// Return defaults
		return &MobilePreferences{
			UserID:                userID,
			OrganizationID:        orgID,
			PushEnabled:           true,
			PushBreachAlerts:      true,
			PushApprovalRequests:  true,
			PushIncidentAlerts:    true,
			PushDeadlineReminders: true,
			PushMentions:          true,
			PushComments:          false,
			QuietHoursEnabled:     false,
			QuietHoursStart:       "22:00",
			QuietHoursEnd:         "08:00",
			QuietHoursTimezone:    "UTC",
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get preferences: %w", err)
	}
	return &prefs, nil
}

// UpdatePreferences creates or updates mobile push preferences.
// push_breach_alerts is always forced to true.
func (s *PushService) UpdatePreferences(ctx context.Context, orgID, userID uuid.UUID, prefs *MobilePreferences) (*MobilePreferences, error) {
	// Breach alerts cannot be disabled
	prefs.PushBreachAlerts = true

	if prefs.QuietHoursTimezone == "" {
		prefs.QuietHoursTimezone = "UTC"
	}
	if prefs.QuietHoursStart == "" {
		prefs.QuietHoursStart = "22:00"
	}
	if prefs.QuietHoursEnd == "" {
		prefs.QuietHoursEnd = "08:00"
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return nil, fmt.Errorf("set org context: %w", err)
	}

	var result MobilePreferences
	err = tx.QueryRow(ctx, `
		INSERT INTO user_mobile_preferences (
			user_id, organization_id, push_enabled, push_breach_alerts,
			push_approval_requests, push_incident_alerts, push_deadline_reminders,
			push_mentions, push_comments, quiet_hours_enabled,
			quiet_hours_start, quiet_hours_end, quiet_hours_timezone
		) VALUES ($1, $2, $3, true, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (user_id) DO UPDATE SET
			push_enabled = EXCLUDED.push_enabled,
			push_breach_alerts = true,
			push_approval_requests = EXCLUDED.push_approval_requests,
			push_incident_alerts = EXCLUDED.push_incident_alerts,
			push_deadline_reminders = EXCLUDED.push_deadline_reminders,
			push_mentions = EXCLUDED.push_mentions,
			push_comments = EXCLUDED.push_comments,
			quiet_hours_enabled = EXCLUDED.quiet_hours_enabled,
			quiet_hours_start = EXCLUDED.quiet_hours_start,
			quiet_hours_end = EXCLUDED.quiet_hours_end,
			quiet_hours_timezone = EXCLUDED.quiet_hours_timezone,
			updated_at = NOW()
		RETURNING id, user_id, organization_id, push_enabled, push_breach_alerts,
			push_approval_requests, push_incident_alerts, push_deadline_reminders,
			push_mentions, push_comments, quiet_hours_enabled,
			quiet_hours_start, quiet_hours_end, quiet_hours_timezone,
			created_at, updated_at`,
		userID, orgID, prefs.PushEnabled,
		prefs.PushApprovalRequests, prefs.PushIncidentAlerts,
		prefs.PushDeadlineReminders, prefs.PushMentions, prefs.PushComments,
		prefs.QuietHoursEnabled, prefs.QuietHoursStart, prefs.QuietHoursEnd,
		prefs.QuietHoursTimezone,
	).Scan(
		&result.ID, &result.UserID, &result.OrganizationID,
		&result.PushEnabled, &result.PushBreachAlerts,
		&result.PushApprovalRequests, &result.PushIncidentAlerts,
		&result.PushDeadlineReminders, &result.PushMentions, &result.PushComments,
		&result.QuietHoursEnabled, &result.QuietHoursStart, &result.QuietHoursEnd,
		&result.QuietHoursTimezone, &result.CreatedAt, &result.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("upsert preferences: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &result, nil
}

// ============================================================
// SEND PUSH NOTIFICATIONS
// ============================================================

// SendPush sends a push notification to all active tokens for a user,
// respecting their notification preferences and quiet hours settings.
// Breach alerts ALWAYS bypass preferences and quiet hours.
func (s *PushService) SendPush(ctx context.Context, orgID, userID uuid.UUID, notification *PushNotification) error {
	if notification == nil {
		return fmt.Errorf("notification cannot be nil")
	}

	isBreachAlert := notification.Type == PushTypeBreachAlert

	// Check preferences (breach alerts bypass all preference checks)
	if !isBreachAlert {
		prefs, err := s.GetPreferences(ctx, orgID, userID)
		if err != nil {
			log.Warn().Err(err).Str("user_id", userID.String()).Msg("Failed to get push preferences, proceeding with defaults")
		} else {
			// Global push disabled
			if !prefs.PushEnabled {
				log.Debug().Str("user_id", userID.String()).Msg("Push notifications disabled for user, skipping")
				return nil
			}

			// Check type-specific preference
			if !s.isTypeEnabled(prefs, notification.Type) {
				log.Debug().
					Str("user_id", userID.String()).
					Str("type", notification.Type).
					Msg("Push type disabled by user preference, skipping")
				return nil
			}

			// Check quiet hours
			if s.IsInQuietHours(prefs) {
				log.Debug().
					Str("user_id", userID.String()).
					Str("timezone", prefs.QuietHoursTimezone).
					Msg("User is in quiet hours, skipping push")
				return nil
			}
		}
	}

	// Get active tokens
	tokens, err := s.GetActiveTokens(ctx, orgID, userID)
	if err != nil {
		return fmt.Errorf("get active tokens: %w", err)
	}
	if len(tokens) == 0 {
		log.Debug().Str("user_id", userID.String()).Msg("No active push tokens for user")
		return nil
	}

	// Set priority for breach alerts
	if isBreachAlert && notification.Priority == "" {
		notification.Priority = "high"
	}

	// Send to all active tokens
	for i := range tokens {
		token := &tokens[i]
		err := s.provider.Send(ctx, token.Platform, token.Token, notification)

		status := "sent"
		errMsg := ""
		if err != nil {
			errMsg = err.Error()
			if isInvalidTokenError(err) {
				status = "invalid_token"
				// Deactivate invalid token
				s.deactivateToken(ctx, token.ID)
			} else {
				status = "failed"
			}
		}

		// Log the delivery attempt
		s.logPushDelivery(ctx, orgID, userID, token, notification, status, errMsg)

		// Update last_used_at for successful sends
		if status == "sent" {
			s.db.Exec(ctx, `
				UPDATE push_notification_tokens SET last_used_at = NOW()
				WHERE id = $1`, token.ID)
		}
	}

	return nil
}

// SendBulkPush sends a push notification to multiple users.
func (s *PushService) SendBulkPush(ctx context.Context, orgID uuid.UUID, userIDs []uuid.UUID, notification *PushNotification) error {
	if len(userIDs) == 0 {
		return nil
	}

	var lastErr error
	sent := 0
	failed := 0

	for _, userID := range userIDs {
		if err := s.SendPush(ctx, orgID, userID, notification); err != nil {
			log.Warn().
				Err(err).
				Str("user_id", userID.String()).
				Str("type", notification.Type).
				Msg("Failed to send bulk push to user")
			lastErr = err
			failed++
		} else {
			sent++
		}
	}

	log.Info().
		Int("sent", sent).
		Int("failed", failed).
		Int("total_users", len(userIDs)).
		Str("type", notification.Type).
		Msg("Bulk push notification completed")

	if failed > 0 && sent == 0 {
		return fmt.Errorf("all bulk push sends failed, last error: %w", lastErr)
	}

	return nil
}

// ============================================================
// QUIET HOURS
// ============================================================

// IsInQuietHours checks if the current time falls within the user's quiet hours,
// using the user's configured timezone.
func (s *PushService) IsInQuietHours(prefs *MobilePreferences) bool {
	if prefs == nil || !prefs.QuietHoursEnabled {
		return false
	}

	loc, err := time.LoadLocation(prefs.QuietHoursTimezone)
	if err != nil {
		log.Warn().
			Str("timezone", prefs.QuietHoursTimezone).
			Err(err).
			Msg("Invalid timezone in quiet hours preferences, defaulting to UTC")
		loc = time.UTC
	}

	now := time.Now().In(loc)
	currentTime := now.Format("15:04")

	start := prefs.QuietHoursStart
	end := prefs.QuietHoursEnd

	return isTimeInRange(currentTime, start, end)
}

// isTimeInRange checks if currentTime is between start and end.
// Handles overnight ranges (e.g. 22:00 to 08:00).
func isTimeInRange(current, start, end string) bool {
	if start <= end {
		// Same-day range: e.g. 09:00 - 17:00
		return current >= start && current < end
	}
	// Overnight range: e.g. 22:00 - 08:00
	return current >= start || current < end
}

// ============================================================
// NOTIFICATION LOG
// ============================================================

// GetPushLog returns recent push notification log entries for a user.
func (s *PushService) GetPushLog(ctx context.Context, orgID, userID uuid.UUID, limit int) ([]PushLogEntry, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	rows, err := s.db.Query(ctx, `
		SELECT id, user_id, token_id, notification_type, title, body, data,
			status, platform, sent_at, error_message
		FROM push_notification_log
		WHERE user_id = $1 AND organization_id = $2
		ORDER BY sent_at DESC
		LIMIT $3`,
		userID, orgID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("query push log: %w", err)
	}
	defer rows.Close()

	var entries []PushLogEntry
	for rows.Next() {
		var e PushLogEntry
		var dataJSON []byte
		var errMsg *string
		if err := rows.Scan(
			&e.ID, &e.UserID, &e.TokenID, &e.NotificationType,
			&e.Title, &e.Body, &dataJSON,
			&e.Status, &e.Platform, &e.SentAt, &errMsg,
		); err != nil {
			return nil, fmt.Errorf("scan push log entry: %w", err)
		}
		if len(dataJSON) > 0 {
			json.Unmarshal(dataJSON, &e.Data)
		}
		if errMsg != nil {
			e.ErrorMessage = *errMsg
		}
		entries = append(entries, e)
	}

	return entries, nil
}

// ============================================================
// INTERNAL HELPERS
// ============================================================

// isTypeEnabled checks if a specific notification type is enabled in preferences.
func (s *PushService) isTypeEnabled(prefs *MobilePreferences, notifType string) bool {
	if prefs == nil {
		return true
	}
	switch notifType {
	case PushTypeBreachAlert:
		return true // Always enabled, enforced by DB constraint too
	case PushTypeApprovalRequest:
		return prefs.PushApprovalRequests
	case PushTypeIncidentCreated:
		return prefs.PushIncidentAlerts
	case PushTypeDeadlineReminder:
		return prefs.PushDeadlineReminders
	case PushTypeMention:
		return prefs.PushMentions
	case PushTypeComment:
		return prefs.PushComments
	default:
		return prefs.PushEnabled
	}
}

// logPushDelivery persists a push notification delivery attempt.
func (s *PushService) logPushDelivery(ctx context.Context, orgID, userID uuid.UUID, token *PushToken, notification *PushNotification, status, errMsg string) {
	dataJSON, _ := json.Marshal(notification.Data)
	if dataJSON == nil {
		dataJSON = []byte("{}")
	}

	var errMsgPtr *string
	if errMsg != "" {
		errMsgPtr = &errMsg
	}

	_, err := s.db.Exec(ctx, `
		INSERT INTO push_notification_log (
			user_id, organization_id, token_id, notification_type,
			title, body, data, status, platform, sent_at, error_message
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8::push_delivery_status, $9::push_platform, NOW(), $10)`,
		userID, orgID, token.ID, notification.Type,
		notification.Title, notification.Body, dataJSON,
		status, token.Platform, errMsgPtr,
	)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("Failed to log push delivery")
	}
}

// deactivateToken marks a push token as inactive (e.g. when it's invalid).
func (s *PushService) deactivateToken(ctx context.Context, tokenID uuid.UUID) {
	_, err := s.db.Exec(ctx, `
		UPDATE push_notification_tokens SET is_active = false, updated_at = NOW()
		WHERE id = $1`, tokenID)
	if err != nil {
		log.Error().Err(err).Str("token_id", tokenID.String()).Msg("Failed to deactivate invalid push token")
	} else {
		log.Info().Str("token_id", tokenID.String()).Msg("Deactivated invalid push token")
	}
}

// hashToken produces a SHA-256 hex hash of the raw push token.
func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// HashPushToken is the exported version for use by handlers.
func HashPushToken(token string) string {
	return hashToken(token)
}

// isInvalidTokenError checks if a push provider error indicates an invalid/expired token.
func isInvalidTokenError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "invalid") ||
		strings.Contains(msg, "unregistered") ||
		strings.Contains(msg, "not registered") ||
		strings.Contains(msg, "expired token")
}

// truncateToken returns the first 12 characters of a token for logging.
func truncateToken(token string) string {
	if len(token) <= 12 {
		return token
	}
	return token[:12] + "..."
}
