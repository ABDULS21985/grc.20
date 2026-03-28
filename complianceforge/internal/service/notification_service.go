package service

import (
	"fmt"
	"net/smtp"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/complianceforge/platform/internal/config"
)

// NotificationService sends email notifications.
type NotificationService struct {
	cfg config.SMTPConfig
}

func NewNotificationService(cfg config.SMTPConfig) *NotificationService {
	return &NotificationService{cfg: cfg}
}

// SendEmail sends a plain-text email.
func (s *NotificationService) SendEmail(to []string, subject, body string) error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
	from := s.cfg.From

	msg := strings.Join([]string{
		"From: " + s.cfg.FromName + " <" + from + ">",
		"To: " + strings.Join(to, ", "),
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=UTF-8",
		"",
		body,
	}, "\r\n")

	var auth smtp.Auth
	if s.cfg.User != "" {
		auth = smtp.PlainAuth("", s.cfg.User, s.cfg.Password, s.cfg.Host)
	}

	if err := smtp.SendMail(addr, auth, from, to, []byte(msg)); err != nil {
		log.Error().Err(err).Strs("to", to).Str("subject", subject).Msg("Failed to send email")
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Info().Strs("to", to).Str("subject", subject).Msg("Email sent")
	return nil
}

// SendBreachNotificationReminder sends an urgent reminder about GDPR breach notification deadline.
func (s *NotificationService) SendBreachNotificationReminder(to []string, incidentRef, deadline string) error {
	subject := fmt.Sprintf("[URGENT] GDPR Breach Notification Deadline — %s", incidentRef)
	body := fmt.Sprintf(`
		<h2>GDPR Data Breach — Notification Deadline Approaching</h2>
		<p>Incident <strong>%s</strong> requires notification to the supervisory authority.</p>
		<p><strong>Deadline: %s</strong> (72 hours from discovery)</p>
		<p>Please ensure the DPA notification is submitted before the deadline to avoid regulatory penalties.</p>
		<p>Log into ComplianceForge to submit the notification.</p>
	`, incidentRef, deadline)

	return s.SendEmail(to, subject, body)
}

// SendPolicyReviewReminder sends a reminder about an upcoming policy review.
func (s *NotificationService) SendPolicyReviewReminder(to []string, policyRef, title, dueDate string) error {
	subject := fmt.Sprintf("Policy Review Due — %s: %s", policyRef, title)
	body := fmt.Sprintf(`
		<h2>Policy Review Reminder</h2>
		<p>Policy <strong>%s — %s</strong> is due for review.</p>
		<p><strong>Review Due: %s</strong></p>
		<p>Please log into ComplianceForge to complete the review.</p>
	`, policyRef, title, dueDate)

	return s.SendEmail(to, subject, body)
}

// SendAuditFindingEscalation sends an escalation for overdue audit findings.
func (s *NotificationService) SendAuditFindingEscalation(to []string, findingRef, severity, dueDate string) error {
	subject := fmt.Sprintf("[OVERDUE] Audit Finding %s — %s Severity", findingRef, severity)
	body := fmt.Sprintf(`
		<h2>Overdue Audit Finding</h2>
		<p>Finding <strong>%s</strong> (Severity: %s) has exceeded its remediation deadline.</p>
		<p><strong>Original Due Date: %s</strong></p>
		<p>Please take immediate action to address this finding.</p>
	`, findingRef, severity, dueDate)

	return s.SendEmail(to, subject, body)
}
