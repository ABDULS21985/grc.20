// Package email provides templated email sending for ComplianceForge.
// All notifications comply with GDPR Article 33 (breach notification),
// NIS2 Article 23 (incident reporting), and ISO 27001 A.5.24 (incident management).
package email

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/complianceforge/platform/internal/config"
)

// Sender handles email delivery.
type Sender struct {
	cfg       config.SMTPConfig
	templates map[string]*template.Template
}

// NewSender creates a new email sender with pre-compiled templates.
func NewSender(cfg config.SMTPConfig) *Sender {
	s := &Sender{
		cfg:       cfg,
		templates: make(map[string]*template.Template),
	}
	s.loadTemplates()
	return s
}

func (s *Sender) loadTemplates() {
	s.templates["breach_alert"] = template.Must(template.New("breach_alert").Parse(breachAlertTemplate))
	s.templates["policy_review"] = template.Must(template.New("policy_review").Parse(policyReviewTemplate))
	s.templates["finding_escalation"] = template.Must(template.New("finding_escalation").Parse(findingEscalationTemplate))
	s.templates["attestation_reminder"] = template.Must(template.New("attestation_reminder").Parse(attestationReminderTemplate))
	s.templates["vendor_assessment_due"] = template.Must(template.New("vendor_assessment_due").Parse(vendorAssessmentTemplate))
	s.templates["welcome"] = template.Must(template.New("welcome").Parse(welcomeTemplate))
}

// SendTemplated renders a template and sends the email.
func (s *Sender) SendTemplated(to []string, subject, templateName string, data interface{}) error {
	tmpl, ok := s.templates[templateName]
	if !ok {
		return fmt.Errorf("template not found: %s", templateName)
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("template render error: %w", err)
	}

	return s.Send(to, subject, body.String())
}

// Send delivers a raw HTML email.
func (s *Sender) Send(to []string, subject, htmlBody string) error {
	msg := strings.Join([]string{
		"From: " + s.cfg.FromName + " <" + s.cfg.From + ">",
		"To: " + strings.Join(to, ", "),
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=UTF-8",
		"X-Mailer: ComplianceForge/1.0",
		"Date: " + time.Now().Format(time.RFC1123Z),
		"",
		htmlBody,
	}, "\r\n")

	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
	var auth smtp.Auth
	if s.cfg.User != "" {
		auth = smtp.PlainAuth("", s.cfg.User, s.cfg.Password, s.cfg.Host)
	}

	if err := smtp.SendMail(addr, auth, s.cfg.From, to, []byte(msg)); err != nil {
		log.Error().Err(err).Strs("to", to).Str("subject", subject).Msg("Email delivery failed")
		return err
	}

	log.Info().Strs("to", to).Str("subject", subject).Msg("Email sent")
	return nil
}

// ============================================================
// EMAIL TEMPLATES
// ============================================================

const baseStyle = `
<style>
  body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Arial, sans-serif; color: #333; line-height: 1.6; }
  .container { max-width: 600px; margin: 0 auto; padding: 20px; }
  .header { background: #1a237e; color: white; padding: 20px; border-radius: 8px 8px 0 0; }
  .header h1 { margin: 0; font-size: 20px; }
  .content { background: #f8f9fa; padding: 24px; border: 1px solid #e0e0e0; }
  .alert-critical { border-left: 4px solid #d32f2f; background: #ffebee; padding: 16px; margin: 16px 0; }
  .alert-warning { border-left: 4px solid #f57c00; background: #fff3e0; padding: 16px; margin: 16px 0; }
  .alert-info { border-left: 4px solid #1976d2; background: #e3f2fd; padding: 16px; margin: 16px 0; }
  .btn { display: inline-block; background: #1a237e; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px; margin-top: 16px; }
  .footer { text-align: center; color: #999; font-size: 12px; padding: 16px; }
  .deadline { font-size: 24px; font-weight: bold; color: #d32f2f; }
</style>`

const breachAlertTemplate = `<!DOCTYPE html><html><head>` + baseStyle + `</head><body>
<div class="container">
  <div class="header"><h1>⚠️ GDPR Data Breach — Notification Required</h1></div>
  <div class="content">
    <div class="alert-critical">
      <strong>URGENT: Supervisory Authority Notification Required</strong><br>
      Per GDPR Article 33, you must notify the ICO/DPA within 72 hours.
    </div>
    <p><strong>Incident:</strong> {{.IncidentRef}}</p>
    <p><strong>Title:</strong> {{.Title}}</p>
    <p><strong>Severity:</strong> {{.Severity}}</p>
    <p><strong>Data Subjects Affected:</strong> {{.DataSubjectsAffected}}</p>
    <p class="deadline">Deadline: {{.Deadline}}</p>
    <p><strong>Hours Remaining:</strong> {{.HoursRemaining}}</p>
    <a href="{{.DashboardURL}}" class="btn">Open Incident in ComplianceForge</a>
  </div>
  <div class="footer">ComplianceForge GRC Platform — Automated Breach Alert</div>
</div></body></html>`

const policyReviewTemplate = `<!DOCTYPE html><html><head>` + baseStyle + `</head><body>
<div class="container">
  <div class="header"><h1>📋 Policy Review Due</h1></div>
  <div class="content">
    <div class="alert-warning">
      <strong>Policy review is {{.Status}}</strong>
    </div>
    <p><strong>Policy:</strong> {{.PolicyRef}} — {{.Title}}</p>
    <p><strong>Owner:</strong> {{.OwnerName}}</p>
    <p><strong>Review Due:</strong> {{.DueDate}}</p>
    <p><strong>Last Reviewed:</strong> {{.LastReviewDate}}</p>
    <a href="{{.DashboardURL}}" class="btn">Review Policy</a>
  </div>
  <div class="footer">ComplianceForge GRC Platform — Policy Management</div>
</div></body></html>`

const findingEscalationTemplate = `<!DOCTYPE html><html><head>` + baseStyle + `</head><body>
<div class="container">
  <div class="header"><h1>🔍 Overdue Audit Finding</h1></div>
  <div class="content">
    <div class="alert-critical">
      <strong>Finding {{.FindingRef}} has exceeded its remediation deadline</strong>
    </div>
    <p><strong>Audit:</strong> {{.AuditRef}} — {{.AuditTitle}}</p>
    <p><strong>Finding:</strong> {{.FindingTitle}}</p>
    <p><strong>Severity:</strong> {{.Severity}}</p>
    <p><strong>Original Due Date:</strong> {{.DueDate}}</p>
    <p><strong>Responsible:</strong> {{.ResponsibleName}}</p>
    <p><strong>Days Overdue:</strong> {{.DaysOverdue}}</p>
    <a href="{{.DashboardURL}}" class="btn">View Finding</a>
  </div>
  <div class="footer">ComplianceForge GRC Platform — Audit Management</div>
</div></body></html>`

const attestationReminderTemplate = `<!DOCTYPE html><html><head>` + baseStyle + `</head><body>
<div class="container">
  <div class="header"><h1>📝 Policy Attestation Required</h1></div>
  <div class="content">
    <div class="alert-info">
      <strong>Please acknowledge the following policy</strong>
    </div>
    <p><strong>Policy:</strong> {{.PolicyTitle}}</p>
    <p><strong>Version:</strong> {{.VersionLabel}}</p>
    <p><strong>Due By:</strong> {{.DueDate}}</p>
    <p>Please read the policy and confirm your understanding by the deadline.</p>
    <a href="{{.AttestURL}}" class="btn">Read & Acknowledge Policy</a>
  </div>
  <div class="footer">ComplianceForge GRC Platform — Policy Compliance</div>
</div></body></html>`

const vendorAssessmentTemplate = `<!DOCTYPE html><html><head>` + baseStyle + `</head><body>
<div class="container">
  <div class="header"><h1>🏢 Vendor Assessment Due</h1></div>
  <div class="content">
    <div class="alert-warning">
      <strong>Vendor risk assessment is due for review</strong>
    </div>
    <p><strong>Vendor:</strong> {{.VendorRef}} — {{.VendorName}}</p>
    <p><strong>Risk Tier:</strong> {{.RiskTier}}</p>
    <p><strong>Assessment Due:</strong> {{.DueDate}}</p>
    <p><strong>Relationship Owner:</strong> {{.OwnerName}}</p>
    {{if .DPAMissing}}
    <div class="alert-critical">
      <strong>GDPR Alert:</strong> Data Processing Agreement (DPA) is not in place for this vendor.
    </div>
    {{end}}
    <a href="{{.DashboardURL}}" class="btn">Start Assessment</a>
  </div>
  <div class="footer">ComplianceForge GRC Platform — Vendor Management</div>
</div></body></html>`

const welcomeTemplate = `<!DOCTYPE html><html><head>` + baseStyle + `</head><body>
<div class="container">
  <div class="header"><h1>Welcome to ComplianceForge</h1></div>
  <div class="content">
    <p>Hello {{.FirstName}},</p>
    <p>Your account has been created for <strong>{{.OrganizationName}}</strong>.</p>
    <p><strong>Email:</strong> {{.Email}}</p>
    <p><strong>Role:</strong> {{.RoleName}}</p>
    <p>Please set your password using the link below:</p>
    <a href="{{.SetPasswordURL}}" class="btn">Set Your Password</a>
    <p style="margin-top:16px; color:#666;">This link expires in 24 hours.</p>
  </div>
  <div class="footer">ComplianceForge GRC Platform</div>
</div></body></html>`
