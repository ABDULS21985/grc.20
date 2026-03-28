package notification_channels

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/complianceforge/platform/internal/service"
)

// WebhookChannel delivers notifications via HTTP POST with HMAC-SHA256 signature.
type WebhookChannel struct {
	client *http.Client
}

// NewWebhookChannel creates a new webhook delivery channel.
func NewWebhookChannel() *WebhookChannel {
	return &WebhookChannel{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Type returns the channel type identifier.
func (c *WebhookChannel) Type() service.ChannelType {
	return service.ChannelWebhook
}

// Send delivers a notification via HTTP POST to the configured webhook URL.
// The request includes an HMAC-SHA256 signature in the X-CF-Signature header.
func (c *WebhookChannel) Send(ctx context.Context, notification *service.Notification) error {
	// Build webhook payload
	payload := map[string]interface{}{
		"notification_id": notification.ID.String(),
		"event_type":      notification.EventType,
		"organization_id": notification.OrganizationID.String(),
		"subject":         notification.Subject,
		"body":            notification.Body,
		"event_payload":   notification.EventPayload,
		"timestamp":       notification.CreatedAt.UTC().Format(time.RFC3339),
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal webhook payload: %w", err)
	}

	// Extract webhook URL and secret from metadata (set during channel configuration)
	webhookURL, _ := notification.Metadata["webhook_url"].(string)
	webhookSecret, _ := notification.Metadata["webhook_secret"].(string)

	if webhookURL == "" {
		return fmt.Errorf("no webhook_url configured for notification %s", notification.ID)
	}

	// Compute HMAC-SHA256 signature
	signature := computeHMAC(payloadBytes, webhookSecret)

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(payloadBytes))
	if err != nil {
		return fmt.Errorf("create webhook request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "ComplianceForge-Webhook/2.0")
	req.Header.Set("X-CF-Signature", "sha256="+signature)
	req.Header.Set("X-CF-Event-Type", notification.EventType)
	req.Header.Set("X-CF-Notification-ID", notification.ID.String())
	req.Header.Set("X-CF-Timestamp", fmt.Sprintf("%d", time.Now().Unix()))

	// Send request
	resp, err := c.client.Do(req)
	if err != nil {
		log.Error().Err(err).
			Str("url", webhookURL).
			Str("notification_id", notification.ID.String()).
			Msg("Webhook delivery failed")
		return fmt.Errorf("webhook request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body (limited to 1KB for logging)
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Error().
			Int("status_code", resp.StatusCode).
			Str("url", webhookURL).
			Str("response", string(respBody)).
			Str("notification_id", notification.ID.String()).
			Msg("Webhook endpoint returned non-2xx status")
		return fmt.Errorf("webhook returned status %d: %s", resp.StatusCode, string(respBody))
	}

	log.Info().
		Str("url", webhookURL).
		Int("status_code", resp.StatusCode).
		Str("notification_id", notification.ID.String()).
		Msg("Webhook notification delivered")

	return nil
}

// computeHMAC generates an HMAC-SHA256 hex digest for webhook signature verification.
func computeHMAC(payload []byte, secret string) string {
	if secret == "" {
		secret = "unsigned"
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}
