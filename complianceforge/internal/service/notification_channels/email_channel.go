// Package notification_channels provides delivery channel implementations for
// the ComplianceForge notification engine.
package notification_channels

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/complianceforge/platform/internal/config"
	"github.com/complianceforge/platform/internal/service"
)

// EmailChannel delivers notifications via SMTP email.
type EmailChannel struct {
	cfg config.SMTPConfig
}

// NewEmailChannel creates a new SMTP email delivery channel.
func NewEmailChannel(cfg config.SMTPConfig) *EmailChannel {
	return &EmailChannel{cfg: cfg}
}

// Type returns the channel type identifier.
func (c *EmailChannel) Type() service.ChannelType {
	return service.ChannelEmail
}

// Send delivers an email notification by looking up the recipient's email address
// and sending the rendered HTML body via SMTP.
func (c *EmailChannel) Send(ctx context.Context, notification *service.Notification) error {
	// Resolve recipient email from the notification payload or metadata
	recipientEmail := ""
	if email, ok := notification.EventPayload["recipient_email"].(string); ok && email != "" {
		recipientEmail = email
	}
	if recipientEmail == "" {
		if email, ok := notification.Metadata["recipient_email"].(string); ok && email != "" {
			recipientEmail = email
		}
	}
	if recipientEmail == "" {
		// Fall back to user ID — in production, this would query the users table.
		// For now we store it in metadata when building notifications.
		return fmt.Errorf("no recipient email available for user %s", notification.RecipientUserID)
	}

	to := []string{recipientEmail}
	subject := notification.Subject
	body := notification.Body

	msg := strings.Join([]string{
		"From: " + c.cfg.FromName + " <" + c.cfg.From + ">",
		"To: " + strings.Join(to, ", "),
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=UTF-8",
		"X-Mailer: ComplianceForge/2.0",
		"X-CF-Event-Type: " + notification.EventType,
		"X-CF-Notification-ID: " + notification.ID.String(),
		"Date: " + time.Now().Format(time.RFC1123Z),
		"",
		body,
	}, "\r\n")

	addr := fmt.Sprintf("%s:%d", c.cfg.Host, c.cfg.Port)
	var auth smtp.Auth
	if c.cfg.User != "" {
		auth = smtp.PlainAuth("", c.cfg.User, c.cfg.Password, c.cfg.Host)
	}

	if err := smtp.SendMail(addr, auth, c.cfg.From, to, []byte(msg)); err != nil {
		log.Error().Err(err).
			Strs("to", to).
			Str("subject", subject).
			Str("notification_id", notification.ID.String()).
			Msg("Email notification delivery failed")
		return fmt.Errorf("smtp send failed: %w", err)
	}

	log.Info().
		Strs("to", to).
		Str("subject", subject).
		Str("notification_id", notification.ID.String()).
		Msg("Email notification sent")

	return nil
}
