package service

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
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
// WEBHOOK SERVICE
// Manages webhook subscriptions, event delivery, signature
// generation/verification, and delivery replay.
// ============================================================

// WebhookService handles webhook business logic.
type WebhookService struct {
	pool *pgxpool.Pool
}

// NewWebhookService creates a new WebhookService.
func NewWebhookService(pool *pgxpool.Pool) *WebhookService {
	return &WebhookService{pool: pool}
}

// ============================================================
// DATA TYPES
// ============================================================

// WebhookSubscription represents a webhook endpoint registration.
type WebhookSubscription struct {
	ID                uuid.UUID  `json:"id"`
	OrganizationID    uuid.UUID  `json:"organization_id"`
	URL               string     `json:"url"`
	Description       string     `json:"description"`
	Secret            string     `json:"secret,omitempty"`
	Events            []string   `json:"events"`
	Status            string     `json:"status"`
	Version           string     `json:"version"`
	Headers           json.RawMessage `json:"headers"`
	FailureCount      int        `json:"failure_count"`
	MaxRetries        int        `json:"max_retries"`
	LastTriggeredAt   *time.Time `json:"last_triggered_at,omitempty"`
	LastSuccessAt     *time.Time `json:"last_success_at,omitempty"`
	LastFailureAt     *time.Time `json:"last_failure_at,omitempty"`
	LastFailureReason *string    `json:"last_failure_reason,omitempty"`
	CreatedBy         *uuid.UUID `json:"created_by,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// WebhookDelivery represents a single delivery attempt for a webhook event.
type WebhookDelivery struct {
	ID              uuid.UUID       `json:"id"`
	SubscriptionID  uuid.UUID       `json:"subscription_id"`
	OrganizationID  uuid.UUID       `json:"organization_id"`
	EventType       string          `json:"event_type"`
	Payload         json.RawMessage `json:"payload"`
	RequestHeaders  json.RawMessage `json:"request_headers"`
	ResponseStatus  *int            `json:"response_status,omitempty"`
	ResponseBody    *string         `json:"response_body,omitempty"`
	ResponseHeaders json.RawMessage `json:"response_headers"`
	Status          string          `json:"status"`
	AttemptCount    int             `json:"attempt_count"`
	MaxAttempts     int             `json:"max_attempts"`
	NextRetryAt     *time.Time      `json:"next_retry_at,omitempty"`
	DurationMs      *int            `json:"duration_ms,omitempty"`
	ErrorMessage    *string         `json:"error_message,omitempty"`
	IdempotencyKey  string          `json:"idempotency_key"`
	CreatedAt       time.Time       `json:"created_at"`
	CompletedAt     *time.Time      `json:"completed_at,omitempty"`
}

// WebhookEvent represents an event to be dispatched to subscribers.
type WebhookEvent struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	OrgID     uuid.UUID   `json:"org_id"`
	Data      interface{} `json:"data"`
}

// CreateSubscriptionReq holds the payload for creating a webhook subscription.
type CreateSubscriptionReq struct {
	URL         string   `json:"url"`
	Description string   `json:"description"`
	Events      []string `json:"events"`
	Version     string   `json:"version"`
	MaxRetries  int      `json:"max_retries"`
}

// UpdateSubscriptionReq holds the payload for updating a webhook subscription.
type UpdateSubscriptionReq struct {
	URL         *string  `json:"url,omitempty"`
	Description *string  `json:"description,omitempty"`
	Events      []string `json:"events,omitempty"`
	Status      *string  `json:"status,omitempty"`
	MaxRetries  *int     `json:"max_retries,omitempty"`
}

// ============================================================
// SIGNATURE HELPERS — HMAC-SHA256
// ============================================================

// GenerateSignature creates an HMAC-SHA256 signature for the given payload.
func (s *WebhookService) GenerateSignature(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

// VerifySignature verifies an HMAC-SHA256 signature against the expected value.
func (s *WebhookService) VerifySignature(payload []byte, secret, signature string) bool {
	if signature == "" || secret == "" {
		return false
	}
	expected := s.GenerateSignature(payload, secret)
	return hmac.Equal([]byte(expected), []byte(signature))
}

// ============================================================
// SECRET GENERATION
// ============================================================

// generateWebhookSecret creates a cryptographically random 32-byte hex secret.
func generateWebhookSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate webhook secret: %w", err)
	}
	return "whsec_" + hex.EncodeToString(b), nil
}

// ============================================================
// SUBSCRIPTION CRUD
// ============================================================

// Subscribe creates a new webhook subscription after validating the URL.
func (s *WebhookService) Subscribe(ctx context.Context, orgID uuid.UUID, req CreateSubscriptionReq) (*WebhookSubscription, error) {
	// Validate HTTPS requirement
	if !strings.HasPrefix(req.URL, "https://") {
		return nil, fmt.Errorf("webhook URL must use HTTPS")
	}

	if len(req.Events) == 0 {
		return nil, fmt.Errorf("at least one event type is required")
	}

	// Generate a cryptographic secret for signing deliveries
	secret, err := generateWebhookSecret()
	if err != nil {
		return nil, err
	}

	version := req.Version
	if version == "" {
		version = "2024-01-01"
	}

	maxRetries := req.MaxRetries
	if maxRetries <= 0 || maxRetries > 10 {
		maxRetries = 5
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return nil, fmt.Errorf("failed to set org context: %w", err)
	}

	var sub WebhookSubscription
	err = tx.QueryRow(ctx, `
		INSERT INTO webhook_subscriptions (
			organization_id, url, description, secret, events, version, max_retries, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, organization_id, url, description, secret, events, status, version,
		          headers, failure_count, max_retries, last_triggered_at, last_success_at,
		          last_failure_at, last_failure_reason, created_by, created_at, updated_at
	`,
		orgID, req.URL, req.Description, secret, req.Events, version, maxRetries, nil,
	).Scan(
		&sub.ID, &sub.OrganizationID, &sub.URL, &sub.Description, &sub.Secret,
		&sub.Events, &sub.Status, &sub.Version, &sub.Headers, &sub.FailureCount,
		&sub.MaxRetries, &sub.LastTriggeredAt, &sub.LastSuccessAt, &sub.LastFailureAt,
		&sub.LastFailureReason, &sub.CreatedBy, &sub.CreatedAt, &sub.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create webhook subscription: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Info().
		Str("subscription_id", sub.ID.String()).
		Str("org_id", orgID.String()).
		Str("url", req.URL).
		Int("event_count", len(req.Events)).
		Msg("Webhook subscription created")

	return &sub, nil
}

// ListSubscriptions returns all webhook subscriptions for an organisation.
func (s *WebhookService) ListSubscriptions(ctx context.Context, orgID uuid.UUID) ([]WebhookSubscription, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return nil, fmt.Errorf("failed to set org context: %w", err)
	}

	rows, err := tx.Query(ctx, `
		SELECT id, organization_id, url, description, events, status, version,
		       headers, failure_count, max_retries, last_triggered_at, last_success_at,
		       last_failure_at, last_failure_reason, created_by, created_at, updated_at
		FROM webhook_subscriptions
		WHERE organization_id = $1
		ORDER BY created_at DESC
	`, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list subscriptions: %w", err)
	}
	defer rows.Close()

	var subs []WebhookSubscription
	for rows.Next() {
		var sub WebhookSubscription
		if err := rows.Scan(
			&sub.ID, &sub.OrganizationID, &sub.URL, &sub.Description, &sub.Events,
			&sub.Status, &sub.Version, &sub.Headers, &sub.FailureCount, &sub.MaxRetries,
			&sub.LastTriggeredAt, &sub.LastSuccessAt, &sub.LastFailureAt,
			&sub.LastFailureReason, &sub.CreatedBy, &sub.CreatedAt, &sub.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan subscription: %w", err)
		}
		// Do not expose the secret in list responses
		sub.Secret = ""
		subs = append(subs, sub)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}

	if subs == nil {
		subs = []WebhookSubscription{}
	}
	return subs, nil
}

// GetSubscription returns a single webhook subscription by ID.
func (s *WebhookService) GetSubscription(ctx context.Context, orgID, subscriptionID uuid.UUID) (*WebhookSubscription, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return nil, fmt.Errorf("failed to set org context: %w", err)
	}

	var sub WebhookSubscription
	err = tx.QueryRow(ctx, `
		SELECT id, organization_id, url, description, events, status, version,
		       headers, failure_count, max_retries, last_triggered_at, last_success_at,
		       last_failure_at, last_failure_reason, created_by, created_at, updated_at
		FROM webhook_subscriptions
		WHERE id = $1 AND organization_id = $2
	`, subscriptionID, orgID).Scan(
		&sub.ID, &sub.OrganizationID, &sub.URL, &sub.Description, &sub.Events,
		&sub.Status, &sub.Version, &sub.Headers, &sub.FailureCount, &sub.MaxRetries,
		&sub.LastTriggeredAt, &sub.LastSuccessAt, &sub.LastFailureAt,
		&sub.LastFailureReason, &sub.CreatedBy, &sub.CreatedAt, &sub.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("webhook subscription not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}

	// Do not expose the secret
	sub.Secret = ""
	return &sub, nil
}

// UpdateSubscription updates a webhook subscription.
func (s *WebhookService) UpdateSubscription(ctx context.Context, orgID, subscriptionID uuid.UUID, req UpdateSubscriptionReq) (*WebhookSubscription, error) {
	if req.URL != nil && !strings.HasPrefix(*req.URL, "https://") {
		return nil, fmt.Errorf("webhook URL must use HTTPS")
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return nil, fmt.Errorf("failed to set org context: %w", err)
	}

	// Build dynamic update
	sets := []string{"updated_at = NOW()"}
	args := []interface{}{}
	argIdx := 1

	if req.URL != nil {
		sets = append(sets, fmt.Sprintf("url = $%d", argIdx))
		args = append(args, *req.URL)
		argIdx++
	}
	if req.Description != nil {
		sets = append(sets, fmt.Sprintf("description = $%d", argIdx))
		args = append(args, *req.Description)
		argIdx++
	}
	if req.Events != nil {
		sets = append(sets, fmt.Sprintf("events = $%d", argIdx))
		args = append(args, req.Events)
		argIdx++
	}
	if req.Status != nil {
		sets = append(sets, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *req.Status)
		argIdx++
	}
	if req.MaxRetries != nil {
		sets = append(sets, fmt.Sprintf("max_retries = $%d", argIdx))
		args = append(args, *req.MaxRetries)
		argIdx++
	}

	query := fmt.Sprintf(
		"UPDATE webhook_subscriptions SET %s WHERE id = $%d AND organization_id = $%d RETURNING id, organization_id, url, description, events, status, version, headers, failure_count, max_retries, last_triggered_at, last_success_at, last_failure_at, last_failure_reason, created_by, created_at, updated_at",
		strings.Join(sets, ", "), argIdx, argIdx+1,
	)
	args = append(args, subscriptionID, orgID)

	var sub WebhookSubscription
	err = tx.QueryRow(ctx, query, args...).Scan(
		&sub.ID, &sub.OrganizationID, &sub.URL, &sub.Description, &sub.Events,
		&sub.Status, &sub.Version, &sub.Headers, &sub.FailureCount, &sub.MaxRetries,
		&sub.LastTriggeredAt, &sub.LastSuccessAt, &sub.LastFailureAt,
		&sub.LastFailureReason, &sub.CreatedBy, &sub.CreatedAt, &sub.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("webhook subscription not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update subscription: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}

	sub.Secret = ""
	return &sub, nil
}

// DeleteSubscription removes a webhook subscription.
func (s *WebhookService) DeleteSubscription(ctx context.Context, orgID, subscriptionID uuid.UUID) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return fmt.Errorf("failed to set org context: %w", err)
	}

	tag, err := tx.Exec(ctx, `
		DELETE FROM webhook_subscriptions
		WHERE id = $1 AND organization_id = $2
	`, subscriptionID, orgID)
	if err != nil {
		return fmt.Errorf("failed to delete subscription: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("webhook subscription not found")
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	log.Info().
		Str("subscription_id", subscriptionID.String()).
		Str("org_id", orgID.String()).
		Msg("Webhook subscription deleted")

	return nil
}

// ============================================================
// EVENT DELIVERY
// ============================================================

// DeliverEvent dispatches an event to all matching active subscriptions.
func (s *WebhookService) DeliverEvent(ctx context.Context, orgID uuid.UUID, eventType string, payload interface{}) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal event payload: %w", err)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return fmt.Errorf("failed to set org context: %w", err)
	}

	// Find all active subscriptions that listen for this event type
	rows, err := tx.Query(ctx, `
		SELECT id, url, secret, max_retries
		FROM webhook_subscriptions
		WHERE organization_id = $1
		  AND status = 'active'
		  AND ($2 = ANY(events) OR '*' = ANY(events))
	`, orgID, eventType)
	if err != nil {
		return fmt.Errorf("failed to query matching subscriptions: %w", err)
	}
	defer rows.Close()

	type subInfo struct {
		id         uuid.UUID
		url        string
		secret     string
		maxRetries int
	}

	var matchingSubs []subInfo
	for rows.Next() {
		var si subInfo
		if err := rows.Scan(&si.id, &si.url, &si.secret, &si.maxRetries); err != nil {
			return fmt.Errorf("failed to scan subscription: %w", err)
		}
		matchingSubs = append(matchingSubs, si)
	}

	if len(matchingSubs) == 0 {
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit: %w", err)
		}
		return nil
	}

	// Create delivery records for each matching subscription
	for _, sub := range matchingSubs {
		event := WebhookEvent{
			ID:        uuid.New().String(),
			Type:      eventType,
			Timestamp: time.Now().UTC(),
			OrgID:     orgID,
			Data:      json.RawMessage(payloadBytes),
		}

		eventBytes, err := json.Marshal(event)
		if err != nil {
			log.Error().Err(err).Str("subscription_id", sub.id.String()).Msg("Failed to marshal webhook event")
			continue
		}

		signature := s.GenerateSignature(eventBytes, sub.secret)
		reqHeaders, _ := json.Marshal(map[string]string{
			"Content-Type":           "application/json",
			"X-Webhook-Signature":    signature,
			"X-Webhook-Event":        eventType,
			"X-Webhook-Delivery-ID":  uuid.New().String(),
			"X-Webhook-Timestamp":    event.Timestamp.Format(time.RFC3339),
		})

		_, err = tx.Exec(ctx, `
			INSERT INTO webhook_deliveries (
				subscription_id, organization_id, event_type, payload,
				request_headers, status, max_attempts
			) VALUES ($1, $2, $3, $4, $5, 'pending', $6)
		`, sub.id, orgID, eventType, eventBytes, reqHeaders, sub.maxRetries)
		if err != nil {
			log.Error().Err(err).Str("subscription_id", sub.id.String()).Msg("Failed to create webhook delivery")
			continue
		}

		// Update last_triggered_at on the subscription
		_, _ = tx.Exec(ctx, `
			UPDATE webhook_subscriptions SET last_triggered_at = NOW()
			WHERE id = $1
		`, sub.id)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit deliveries: %w", err)
	}

	log.Info().
		Str("org_id", orgID.String()).
		Str("event_type", eventType).
		Int("subscriptions_matched", len(matchingSubs)).
		Msg("Webhook event queued for delivery")

	return nil
}

// CreateDelivery creates a single delivery record for a specific subscription.
func (s *WebhookService) CreateDelivery(ctx context.Context, subscriptionID uuid.UUID, eventType string, payload interface{}) (*WebhookDelivery, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	var delivery WebhookDelivery
	err = s.pool.QueryRow(ctx, `
		INSERT INTO webhook_deliveries (
			subscription_id, organization_id, event_type, payload, status
		)
		SELECT $1, ws.organization_id, $2, $3, 'pending'
		FROM webhook_subscriptions ws WHERE ws.id = $1
		RETURNING id, subscription_id, organization_id, event_type, payload,
		          request_headers, response_status, response_body, response_headers,
		          status, attempt_count, max_attempts, next_retry_at, duration_ms,
		          error_message, idempotency_key, created_at, completed_at
	`, subscriptionID, eventType, payloadBytes).Scan(
		&delivery.ID, &delivery.SubscriptionID, &delivery.OrganizationID,
		&delivery.EventType, &delivery.Payload, &delivery.RequestHeaders,
		&delivery.ResponseStatus, &delivery.ResponseBody, &delivery.ResponseHeaders,
		&delivery.Status, &delivery.AttemptCount, &delivery.MaxAttempts,
		&delivery.NextRetryAt, &delivery.DurationMs, &delivery.ErrorMessage,
		&delivery.IdempotencyKey, &delivery.CreatedAt, &delivery.CompletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create delivery: %w", err)
	}

	return &delivery, nil
}

// GetDeliveryHistory returns paginated delivery history for a subscription.
func (s *WebhookService) GetDeliveryHistory(ctx context.Context, orgID, subscriptionID uuid.UUID, page, pageSize int) ([]WebhookDelivery, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return nil, 0, fmt.Errorf("failed to set org context: %w", err)
	}

	var total int
	err = tx.QueryRow(ctx, `
		SELECT COUNT(*) FROM webhook_deliveries
		WHERE subscription_id = $1 AND organization_id = $2
	`, subscriptionID, orgID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count deliveries: %w", err)
	}

	rows, err := tx.Query(ctx, `
		SELECT id, subscription_id, organization_id, event_type, payload,
		       request_headers, response_status, response_body, response_headers,
		       status, attempt_count, max_attempts, next_retry_at, duration_ms,
		       error_message, idempotency_key, created_at, completed_at
		FROM webhook_deliveries
		WHERE subscription_id = $1 AND organization_id = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`, subscriptionID, orgID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query deliveries: %w", err)
	}
	defer rows.Close()

	var deliveries []WebhookDelivery
	for rows.Next() {
		var d WebhookDelivery
		if err := rows.Scan(
			&d.ID, &d.SubscriptionID, &d.OrganizationID, &d.EventType, &d.Payload,
			&d.RequestHeaders, &d.ResponseStatus, &d.ResponseBody, &d.ResponseHeaders,
			&d.Status, &d.AttemptCount, &d.MaxAttempts, &d.NextRetryAt, &d.DurationMs,
			&d.ErrorMessage, &d.IdempotencyKey, &d.CreatedAt, &d.CompletedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan delivery: %w", err)
		}
		deliveries = append(deliveries, d)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, 0, fmt.Errorf("failed to commit: %w", err)
	}

	if deliveries == nil {
		deliveries = []WebhookDelivery{}
	}
	return deliveries, total, nil
}

// ReplayDelivery re-queues a failed delivery for retry.
func (s *WebhookService) ReplayDelivery(ctx context.Context, orgID, deliveryID uuid.UUID) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return fmt.Errorf("failed to set org context: %w", err)
	}

	// Fetch the original delivery
	var subID uuid.UUID
	var eventType string
	var payload json.RawMessage
	err = tx.QueryRow(ctx, `
		SELECT subscription_id, event_type, payload
		FROM webhook_deliveries
		WHERE id = $1 AND organization_id = $2
	`, deliveryID, orgID).Scan(&subID, &eventType, &payload)
	if err == pgx.ErrNoRows {
		return fmt.Errorf("delivery not found")
	}
	if err != nil {
		return fmt.Errorf("failed to get delivery: %w", err)
	}

	// Create a new delivery record for the replay
	_, err = tx.Exec(ctx, `
		INSERT INTO webhook_deliveries (
			subscription_id, organization_id, event_type, payload, status
		) VALUES ($1, $2, $3, $4, 'pending')
	`, subID, orgID, eventType, payload)
	if err != nil {
		return fmt.Errorf("failed to create replay delivery: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	log.Info().
		Str("original_delivery_id", deliveryID.String()).
		Str("org_id", orgID.String()).
		Msg("Webhook delivery replayed")

	return nil
}

// PingWebhook sends a test ping event to a specific subscription.
func (s *WebhookService) PingWebhook(ctx context.Context, orgID, subscriptionID uuid.UUID) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return fmt.Errorf("failed to set org context: %w", err)
	}

	// Verify the subscription exists and belongs to this org
	var subURL, subSecret string
	err = tx.QueryRow(ctx, `
		SELECT url, secret FROM webhook_subscriptions
		WHERE id = $1 AND organization_id = $2
	`, subscriptionID, orgID).Scan(&subURL, &subSecret)
	if err == pgx.ErrNoRows {
		return fmt.Errorf("webhook subscription not found")
	}
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	// Create a ping delivery
	pingPayload := map[string]interface{}{
		"event":     "ping",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"message":   "Webhook connectivity test from ComplianceForge",
	}
	payloadBytes, _ := json.Marshal(pingPayload)
	signature := s.GenerateSignature(payloadBytes, subSecret)
	reqHeaders, _ := json.Marshal(map[string]string{
		"Content-Type":        "application/json",
		"X-Webhook-Signature": signature,
		"X-Webhook-Event":     "ping",
	})

	_, err = tx.Exec(ctx, `
		INSERT INTO webhook_deliveries (
			subscription_id, organization_id, event_type, payload,
			request_headers, status, max_attempts
		) VALUES ($1, $2, 'ping', $3, $4, 'pending', 1)
	`, subscriptionID, orgID, payloadBytes, reqHeaders)
	if err != nil {
		return fmt.Errorf("failed to create ping delivery: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	log.Info().
		Str("subscription_id", subscriptionID.String()).
		Str("url", subURL).
		Msg("Webhook ping queued")

	return nil
}
