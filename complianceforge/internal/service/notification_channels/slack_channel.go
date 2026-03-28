package notification_channels

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/complianceforge/platform/internal/service"
)

// SlackChannel delivers notifications via Slack Incoming Webhooks using Block Kit.
type SlackChannel struct {
	client *http.Client
}

// NewSlackChannel creates a new Slack webhook delivery channel.
func NewSlackChannel() *SlackChannel {
	return &SlackChannel{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Type returns the channel type identifier.
func (c *SlackChannel) Type() service.ChannelType {
	return service.ChannelSlack
}

// Send delivers a notification to a Slack channel via an Incoming Webhook.
// The notification body is expected to be a JSON string representing Slack Block Kit blocks.
func (c *SlackChannel) Send(ctx context.Context, notification *service.Notification) error {
	// Extract Slack webhook URL from metadata
	webhookURL, _ := notification.Metadata["slack_webhook_url"].(string)
	if webhookURL == "" {
		return fmt.Errorf("no slack_webhook_url configured for notification %s", notification.ID)
	}

	// Extract optional channel override
	channel, _ := notification.Metadata["slack_channel"].(string)

	// Build Slack message payload
	slackPayload := buildSlackPayload(notification, channel)

	payloadBytes, err := json.Marshal(slackPayload)
	if err != nil {
		return fmt.Errorf("marshal slack payload: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(payloadBytes))
	if err != nil {
		return fmt.Errorf("create slack request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "ComplianceForge-Slack/2.0")

	// Send request
	resp, err := c.client.Do(req)
	if err != nil {
		log.Error().Err(err).
			Str("notification_id", notification.ID.String()).
			Msg("Slack notification delivery failed")
		return fmt.Errorf("slack request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))

	if resp.StatusCode != http.StatusOK {
		log.Error().
			Int("status_code", resp.StatusCode).
			Str("response", string(respBody)).
			Str("notification_id", notification.ID.String()).
			Msg("Slack webhook returned non-200 status")
		return fmt.Errorf("slack returned status %d: %s", resp.StatusCode, string(respBody))
	}

	log.Info().
		Str("notification_id", notification.ID.String()).
		Str("event_type", notification.EventType).
		Msg("Slack notification delivered")

	return nil
}

// buildSlackPayload constructs a Slack Block Kit message from the notification.
func buildSlackPayload(notification *service.Notification, channel string) map[string]interface{} {
	payload := make(map[string]interface{})

	// Try to parse the notification body as Block Kit JSON
	var blocks []interface{}
	if err := json.Unmarshal([]byte(notification.Body), &blocks); err == nil {
		payload["blocks"] = blocks
	} else {
		// Try parsing as a complete Slack payload (with "blocks" key)
		var fullPayload map[string]interface{}
		if err := json.Unmarshal([]byte(notification.Body), &fullPayload); err == nil {
			if b, ok := fullPayload["blocks"]; ok {
				payload["blocks"] = b
			} else {
				// Fall back to simple text message
				payload["text"] = notification.Subject + "\n" + notification.Body
			}
		} else {
			// Plain text fallback
			payload["text"] = notification.Subject + "\n" + notification.Body
		}
	}

	// Set fallback text (required by Slack for accessibility)
	if _, ok := payload["text"]; !ok {
		payload["text"] = notification.Subject
	}

	// Optional channel override
	if channel != "" {
		payload["channel"] = channel
	}

	// Add context block with metadata
	contextBlock := map[string]interface{}{
		"type": "context",
		"elements": []interface{}{
			map[string]interface{}{
				"type": "mrkdwn",
				"text": fmt.Sprintf("ComplianceForge | %s | %s",
					notification.EventType,
					notification.CreatedAt.Format(time.RFC3339)),
			},
		},
	}

	if blocks, ok := payload["blocks"].([]interface{}); ok {
		payload["blocks"] = append(blocks, contextBlock)
	}

	return payload
}
