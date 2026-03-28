package notification_channels

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"github.com/complianceforge/platform/internal/service"
)

// InAppChannel stores notifications in the database for retrieval via the API.
// These appear in the in-app notification bell/inbox.
type InAppChannel struct {
	db *pgxpool.Pool
}

// NewInAppChannel creates a new in-app notification channel.
func NewInAppChannel(db *pgxpool.Pool) *InAppChannel {
	return &InAppChannel{db: db}
}

// Type returns the channel type identifier.
func (c *InAppChannel) Type() service.ChannelType {
	return service.ChannelInApp
}

// Send persists the notification to the database as an in-app notification.
// The notification is already persisted by the engine, so this channel marks
// it as "delivered" — it is immediately available for the user to read.
func (c *InAppChannel) Send(ctx context.Context, notification *service.Notification) error {
	payloadJSON, _ := json.Marshal(notification.EventPayload)
	metadataJSON, _ := json.Marshal(notification.Metadata)

	// Upsert: if the engine already inserted the record, update it to delivered.
	// Otherwise insert a new record specifically for the in-app channel.
	_, err := c.db.Exec(ctx, `
		INSERT INTO notifications (
			id, organization_id, rule_id, event_type, event_payload,
			recipient_user_id, channel_type, channel_id, subject, body,
			status, sent_at, delivered_at, retry_count, max_retries, metadata, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, 'in_app', $7, $8, $9, 'delivered', NOW(), NOW(), 0, 0, $10, $11)
		ON CONFLICT (id) DO UPDATE SET
			status = 'delivered',
			sent_at = NOW(),
			delivered_at = NOW()`,
		notification.ID, notification.OrganizationID, notification.RuleID,
		notification.EventType, payloadJSON,
		notification.RecipientUserID, notification.ChannelID,
		notification.Subject, notification.Body,
		metadataJSON, notification.CreatedAt,
	)
	if err != nil {
		log.Error().Err(err).
			Str("notification_id", notification.ID.String()).
			Str("user_id", notification.RecipientUserID.String()).
			Msg("Failed to store in-app notification")
		return err
	}

	log.Debug().
		Str("notification_id", notification.ID.String()).
		Str("user_id", notification.RecipientUserID.String()).
		Msg("In-app notification stored")

	return nil
}
