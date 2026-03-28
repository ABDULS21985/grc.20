-- ============================================================
-- ComplianceForge GRC Platform
-- Seed 015: System Notification Templates
-- organization_id IS NULL and is_system=true for all system templates
-- Uses Go text/template syntax for variable interpolation
-- ============================================================

-- 1. GDPR Breach 72h — 48 hours remaining
INSERT INTO notification_templates (id, organization_id, name, event_type, subject_template, body_html_template, body_text_template, in_app_title_template, in_app_body_template, slack_template, webhook_payload_template, variables, is_system)
VALUES (
    'a0000001-0001-4000-8000-000000000001',
    NULL,
    'GDPR Breach — 48h Remaining',
    'breach.deadline_approaching',
    '[URGENT] GDPR Breach Deadline — {{.IncidentRef}} — 48 hours remaining',
    '<!DOCTYPE html><html><body><h2>GDPR Data Breach — Notification Deadline Approaching</h2><div style="border-left:4px solid #d32f2f;background:#ffebee;padding:16px;margin:16px 0"><strong>48 HOURS REMAINING</strong></div><p>Incident <strong>{{.IncidentRef}}</strong> requires notification to the supervisory authority.</p><p><strong>Deadline:</strong> {{.Deadline}}</p><p><strong>Data Subjects Affected:</strong> {{.DataSubjectsAffected}}</p><p><strong>Severity:</strong> {{.Severity}}</p><p>Per GDPR Article 33, you must notify the DPA within 72 hours of becoming aware of the breach.</p><p><a href="{{.DashboardURL}}" style="background:#1a237e;color:white;padding:12px 24px;text-decoration:none;border-radius:4px">Open Incident</a></p></body></html>',
    'GDPR Breach Deadline — 48h Remaining\n\nIncident: {{.IncidentRef}}\nDeadline: {{.Deadline}}\nData Subjects Affected: {{.DataSubjectsAffected}}\n\nNotify the DPA within 72 hours per GDPR Article 33.\n\nOpen incident: {{.DashboardURL}}',
    'GDPR Breach: 48h Remaining — {{.IncidentRef}}',
    'Incident {{.IncidentRef}} must be reported to the DPA. Deadline: {{.Deadline}}. {{.DataSubjectsAffected}} data subjects affected.',
    '{"blocks":[{"type":"header","text":{"type":"plain_text","text":"GDPR Breach — 48h Remaining"}},{"type":"section","fields":[{"type":"mrkdwn","text":"*Incident:* {{.IncidentRef}}"},{"type":"mrkdwn","text":"*Deadline:* {{.Deadline}}"},{"type":"mrkdwn","text":"*Severity:* {{.Severity}}"},{"type":"mrkdwn","text":"*Data Subjects:* {{.DataSubjectsAffected}}"}]},{"type":"actions","elements":[{"type":"button","text":{"type":"plain_text","text":"Open Incident"},"url":"{{.DashboardURL}}","style":"danger"}]}]}',
    '{"event":"breach.deadline_approaching","hours_remaining":48,"incident_ref":"{{.IncidentRef}}","deadline":"{{.Deadline}}","severity":"{{.Severity}}","data_subjects_affected":"{{.DataSubjectsAffected}}"}',
    ARRAY['IncidentRef','Deadline','DataSubjectsAffected','Severity','DashboardURL','HoursRemaining'],
    true
);

-- 2. GDPR Breach 72h — 12 hours remaining
INSERT INTO notification_templates (id, organization_id, name, event_type, subject_template, body_html_template, body_text_template, in_app_title_template, in_app_body_template, slack_template, webhook_payload_template, variables, is_system)
VALUES (
    'a0000001-0002-4000-8000-000000000001',
    NULL,
    'GDPR Breach — 12h Remaining',
    'breach.deadline_approaching',
    '[CRITICAL] GDPR Breach Deadline — {{.IncidentRef}} — 12 hours remaining',
    '<!DOCTYPE html><html><body><h2>GDPR Data Breach — CRITICAL DEADLINE</h2><div style="border-left:4px solid #d32f2f;background:#ffebee;padding:16px;margin:16px 0"><strong>12 HOURS REMAINING — IMMEDIATE ACTION REQUIRED</strong></div><p>Incident <strong>{{.IncidentRef}}</strong> has only 12 hours until the GDPR 72-hour notification deadline expires.</p><p><strong>Deadline:</strong> {{.Deadline}}</p><p><strong>Data Subjects Affected:</strong> {{.DataSubjectsAffected}}</p><p><a href="{{.DashboardURL}}" style="background:#d32f2f;color:white;padding:12px 24px;text-decoration:none;border-radius:4px">Take Action Now</a></p></body></html>',
    'CRITICAL: GDPR Breach Deadline — 12h Remaining\n\nIncident: {{.IncidentRef}}\nDeadline: {{.Deadline}}\nData Subjects Affected: {{.DataSubjectsAffected}}\n\nIMMEDIATE ACTION REQUIRED.\n\nOpen incident: {{.DashboardURL}}',
    'CRITICAL: 12h to GDPR Deadline — {{.IncidentRef}}',
    'Only 12 hours remain to notify the DPA for incident {{.IncidentRef}}. Deadline: {{.Deadline}}.',
    '{"blocks":[{"type":"header","text":{"type":"plain_text","text":"CRITICAL: GDPR Breach — 12h Remaining"}},{"type":"section","fields":[{"type":"mrkdwn","text":"*Incident:* {{.IncidentRef}}"},{"type":"mrkdwn","text":"*Deadline:* {{.Deadline}}"}]},{"type":"actions","elements":[{"type":"button","text":{"type":"plain_text","text":"Take Action Now"},"url":"{{.DashboardURL}}","style":"danger"}]}]}',
    '{"event":"breach.deadline_approaching","hours_remaining":12,"incident_ref":"{{.IncidentRef}}","deadline":"{{.Deadline}}","severity":"{{.Severity}}"}',
    ARRAY['IncidentRef','Deadline','DataSubjectsAffected','Severity','DashboardURL','HoursRemaining'],
    true
);

-- 3. GDPR Breach 72h — 6 hours remaining
INSERT INTO notification_templates (id, organization_id, name, event_type, subject_template, body_html_template, body_text_template, in_app_title_template, in_app_body_template, slack_template, webhook_payload_template, variables, is_system)
VALUES (
    'a0000001-0003-4000-8000-000000000001',
    NULL,
    'GDPR Breach — 6h Remaining',
    'breach.deadline_approaching',
    '[CRITICAL] GDPR Breach Deadline — {{.IncidentRef}} — 6 hours remaining',
    '<!DOCTYPE html><html><body><h2>GDPR Data Breach — URGENT</h2><div style="border-left:4px solid #d32f2f;background:#ffebee;padding:16px;margin:16px 0"><strong>6 HOURS REMAINING — DPA NOTIFICATION OVERDUE RISK</strong></div><p>Incident <strong>{{.IncidentRef}}</strong> has 6 hours left.</p><p><strong>Deadline:</strong> {{.Deadline}}</p><p><a href="{{.DashboardURL}}" style="background:#d32f2f;color:white;padding:12px 24px;text-decoration:none;border-radius:4px">Notify DPA Now</a></p></body></html>',
    'CRITICAL: GDPR Breach Deadline — 6h Remaining\nIncident: {{.IncidentRef}}\nDeadline: {{.Deadline}}\n\nNotify the DPA immediately: {{.DashboardURL}}',
    'URGENT: 6h to GDPR Deadline — {{.IncidentRef}}',
    '6 hours to GDPR deadline for {{.IncidentRef}}. Submit DPA notification immediately.',
    '{"blocks":[{"type":"header","text":{"type":"plain_text","text":"URGENT: 6h to GDPR Deadline"}},{"type":"section","text":{"type":"mrkdwn","text":"*{{.IncidentRef}}* — Deadline: {{.Deadline}}"}}]}',
    '{"event":"breach.deadline_approaching","hours_remaining":6,"incident_ref":"{{.IncidentRef}}","deadline":"{{.Deadline}}"}',
    ARRAY['IncidentRef','Deadline','DataSubjectsAffected','Severity','DashboardURL','HoursRemaining'],
    true
);

-- 4. GDPR Breach 72h — 1 hour remaining
INSERT INTO notification_templates (id, organization_id, name, event_type, subject_template, body_html_template, body_text_template, in_app_title_template, in_app_body_template, slack_template, webhook_payload_template, variables, is_system)
VALUES (
    'a0000001-0004-4000-8000-000000000001',
    NULL,
    'GDPR Breach — 1h Remaining',
    'breach.deadline_approaching',
    '[FINAL WARNING] GDPR Breach Deadline — {{.IncidentRef}} — 1 hour remaining',
    '<!DOCTYPE html><html><body><h2 style="color:#d32f2f">FINAL WARNING: GDPR Breach Notification Deadline</h2><div style="border-left:4px solid #d32f2f;background:#ffebee;padding:16px;margin:16px 0"><strong>1 HOUR REMAINING — REGULATORY NON-COMPLIANCE IMMINENT</strong></div><p>Incident <strong>{{.IncidentRef}}</strong> deadline expires in 1 hour.</p><p><strong>Deadline:</strong> {{.Deadline}}</p><p>Failure to notify may result in fines up to 10M EUR or 2% of annual turnover under GDPR Article 83.</p><p><a href="{{.DashboardURL}}" style="background:#d32f2f;color:white;padding:12px 24px;text-decoration:none;border-radius:4px">Submit DPA Notification</a></p></body></html>',
    'FINAL WARNING: GDPR Deadline in 1 Hour\nIncident: {{.IncidentRef}}\nDeadline: {{.Deadline}}\nSubmit notification NOW: {{.DashboardURL}}',
    'FINAL: 1h to GDPR Deadline — {{.IncidentRef}}',
    'FINAL WARNING: 1 hour to GDPR deadline for {{.IncidentRef}}. Regulatory fines imminent.',
    '{"blocks":[{"type":"header","text":{"type":"plain_text","text":"FINAL WARNING: 1h to GDPR Deadline"}},{"type":"section","text":{"type":"mrkdwn","text":"Submit DPA notification for *{{.IncidentRef}}* immediately.\\nDeadline: {{.Deadline}}"}}]}',
    '{"event":"breach.deadline_approaching","hours_remaining":1,"incident_ref":"{{.IncidentRef}}","deadline":"{{.Deadline}}"}',
    ARRAY['IncidentRef','Deadline','DataSubjectsAffected','Severity','DashboardURL','HoursRemaining'],
    true
);

-- 5. GDPR Breach 72h — Expired
INSERT INTO notification_templates (id, organization_id, name, event_type, subject_template, body_html_template, body_text_template, in_app_title_template, in_app_body_template, slack_template, webhook_payload_template, variables, is_system)
VALUES (
    'a0000001-0005-4000-8000-000000000001',
    NULL,
    'GDPR Breach — Deadline Expired',
    'breach.deadline_expired',
    '[OVERDUE] GDPR Breach Notification Deadline EXPIRED — {{.IncidentRef}}',
    '<!DOCTYPE html><html><body><h2 style="color:#d32f2f">GDPR BREACH NOTIFICATION DEADLINE EXPIRED</h2><div style="border-left:4px solid #d32f2f;background:#ffebee;padding:16px;margin:16px 0"><strong>DEADLINE EXPIRED — REGULATORY NON-COMPLIANCE</strong></div><p>The 72-hour GDPR notification deadline for incident <strong>{{.IncidentRef}}</strong> has passed.</p><p><strong>Original Deadline:</strong> {{.Deadline}}</p><p>Notify the DPA immediately and document the reason for delay per GDPR Article 33(1).</p><p><a href="{{.DashboardURL}}" style="background:#d32f2f;color:white;padding:12px 24px;text-decoration:none;border-radius:4px">Submit Late Notification</a></p></body></html>',
    'OVERDUE: GDPR Breach Deadline EXPIRED\nIncident: {{.IncidentRef}}\nOriginal Deadline: {{.Deadline}}\nSubmit late notification: {{.DashboardURL}}',
    'OVERDUE: GDPR Deadline Expired — {{.IncidentRef}}',
    'GDPR 72h deadline expired for {{.IncidentRef}}. Submit late DPA notification immediately.',
    '{"blocks":[{"type":"header","text":{"type":"plain_text","text":"OVERDUE: GDPR Deadline Expired"}},{"type":"section","text":{"type":"mrkdwn","text":"*{{.IncidentRef}}* deadline has passed.\\nSubmit late notification."}}]}',
    '{"event":"breach.deadline_expired","incident_ref":"{{.IncidentRef}}","deadline":"{{.Deadline}}"}',
    ARRAY['IncidentRef','Deadline','DataSubjectsAffected','Severity','DashboardURL'],
    true
);

-- 6. NIS2 24h Early Warning
INSERT INTO notification_templates (id, organization_id, name, event_type, subject_template, body_html_template, body_text_template, in_app_title_template, in_app_body_template, slack_template, webhook_payload_template, variables, is_system)
VALUES (
    'a0000001-0006-4000-8000-000000000001',
    NULL,
    'NIS2 Early Warning Deadline',
    'nis2.early_warning_due',
    '[NIS2] Early Warning Required — {{.IncidentRef}} — {{.HoursRemaining}}h remaining',
    '<!DOCTYPE html><html><body><h2>NIS2 Early Warning Required</h2><div style="border-left:4px solid #f57c00;background:#fff3e0;padding:16px;margin:16px 0"><strong>NIS2 Article 23: 24-hour Early Warning Deadline</strong></div><p>Incident <strong>{{.IncidentRef}}</strong> is classified as NIS2 reportable.</p><p><strong>Early Warning Deadline:</strong> {{.Deadline}}</p><p><strong>Hours Remaining:</strong> {{.HoursRemaining}}</p><p>Submit an early warning to CSIRT per NIS2 Article 23(4)(a).</p><p><a href="{{.DashboardURL}}" style="background:#1a237e;color:white;padding:12px 24px;text-decoration:none;border-radius:4px">Submit Early Warning</a></p></body></html>',
    'NIS2 Early Warning Required\nIncident: {{.IncidentRef}}\nDeadline: {{.Deadline}}\nHours Remaining: {{.HoursRemaining}}\n\nSubmit at: {{.DashboardURL}}',
    'NIS2 Early Warning Due — {{.IncidentRef}}',
    'NIS2 early warning required for {{.IncidentRef}}. Deadline: {{.Deadline}}.',
    '{"blocks":[{"type":"header","text":{"type":"plain_text","text":"NIS2 Early Warning Required"}},{"type":"section","fields":[{"type":"mrkdwn","text":"*Incident:* {{.IncidentRef}}"},{"type":"mrkdwn","text":"*Deadline:* {{.Deadline}}"}]}]}',
    '{"event":"nis2.early_warning_due","incident_ref":"{{.IncidentRef}}","deadline":"{{.Deadline}}","hours_remaining":"{{.HoursRemaining}}"}',
    ARRAY['IncidentRef','Deadline','HoursRemaining','Severity','DashboardURL'],
    true
);

-- 7. Incident Created (Critical)
INSERT INTO notification_templates (id, organization_id, name, event_type, subject_template, body_html_template, body_text_template, in_app_title_template, in_app_body_template, slack_template, webhook_payload_template, variables, is_system)
VALUES (
    'a0000001-0007-4000-8000-000000000001',
    NULL,
    'Incident Created',
    'incident.created',
    '[{{.Severity}}] New Incident — {{.IncidentRef}}: {{.Title}}',
    '<!DOCTYPE html><html><body><h2>New Incident Reported</h2><div style="border-left:4px solid #d32f2f;background:#ffebee;padding:16px;margin:16px 0"><strong>Severity: {{.Severity}}</strong></div><p><strong>Ref:</strong> {{.IncidentRef}}</p><p><strong>Title:</strong> {{.Title}}</p><p><strong>Type:</strong> {{.IncidentType}}</p><p><strong>Reported By:</strong> {{.ReportedBy}}</p><p><strong>Description:</strong> {{.Description}}</p><p><a href="{{.DashboardURL}}" style="background:#1a237e;color:white;padding:12px 24px;text-decoration:none;border-radius:4px">View Incident</a></p></body></html>',
    'New Incident: {{.IncidentRef}} — {{.Title}}\nSeverity: {{.Severity}}\nType: {{.IncidentType}}\nReported By: {{.ReportedBy}}\n\nView: {{.DashboardURL}}',
    'New Incident: {{.IncidentRef}}',
    '{{.Severity}} incident "{{.Title}}" reported by {{.ReportedBy}}.',
    '{"blocks":[{"type":"header","text":{"type":"plain_text","text":"New Incident: {{.IncidentRef}}"}},{"type":"section","fields":[{"type":"mrkdwn","text":"*Severity:* {{.Severity}}"},{"type":"mrkdwn","text":"*Type:* {{.IncidentType}}"},{"type":"mrkdwn","text":"*Title:* {{.Title}}"},{"type":"mrkdwn","text":"*Reported By:* {{.ReportedBy}}"}]},{"type":"actions","elements":[{"type":"button","text":{"type":"plain_text","text":"View Incident"},"url":"{{.DashboardURL}}"}]}]}',
    '{"event":"incident.created","incident_ref":"{{.IncidentRef}}","title":"{{.Title}}","severity":"{{.Severity}}","incident_type":"{{.IncidentType}}"}',
    ARRAY['IncidentRef','Title','Severity','IncidentType','ReportedBy','Description','DashboardURL'],
    true
);

-- 8. Control Status Changed
INSERT INTO notification_templates (id, organization_id, name, event_type, subject_template, body_html_template, body_text_template, in_app_title_template, in_app_body_template, slack_template, webhook_payload_template, variables, is_system)
VALUES (
    'a0000001-0008-4000-8000-000000000001',
    NULL,
    'Control Status Changed',
    'control.status_changed',
    'Control Status Update — {{.ControlCode}}: {{.OldStatus}} to {{.NewStatus}}',
    '<!DOCTYPE html><html><body><h2>Control Status Changed</h2><p>A control implementation status has been updated.</p><p><strong>Framework:</strong> {{.FrameworkCode}}</p><p><strong>Control:</strong> {{.ControlCode}} — {{.ControlTitle}}</p><p><strong>Previous Status:</strong> {{.OldStatus}}</p><p><strong>New Status:</strong> {{.NewStatus}}</p><p><strong>Changed By:</strong> {{.ChangedBy}}</p><p><a href="{{.DashboardURL}}" style="background:#1a237e;color:white;padding:12px 24px;text-decoration:none;border-radius:4px">View Control</a></p></body></html>',
    'Control Status Changed\nControl: {{.ControlCode}} — {{.ControlTitle}}\nFramework: {{.FrameworkCode}}\nOld: {{.OldStatus}} -> New: {{.NewStatus}}\nChanged By: {{.ChangedBy}}',
    'Control Updated: {{.ControlCode}}',
    '{{.ControlCode}} status changed from {{.OldStatus}} to {{.NewStatus}}.',
    '{"blocks":[{"type":"section","text":{"type":"mrkdwn","text":"*Control Status Changed*\\n{{.ControlCode}}: {{.OldStatus}} -> {{.NewStatus}}"}}]}',
    '{"event":"control.status_changed","control_code":"{{.ControlCode}}","old_status":"{{.OldStatus}}","new_status":"{{.NewStatus}}"}',
    ARRAY['ControlCode','ControlTitle','FrameworkCode','OldStatus','NewStatus','ChangedBy','DashboardURL'],
    true
);

-- 9. Policy Review Due
INSERT INTO notification_templates (id, organization_id, name, event_type, subject_template, body_html_template, body_text_template, in_app_title_template, in_app_body_template, slack_template, webhook_payload_template, variables, is_system)
VALUES (
    'a0000001-0009-4000-8000-000000000001',
    NULL,
    'Policy Review Due',
    'policy.review_due',
    'Policy Review Due — {{.PolicyRef}}: {{.Title}}',
    '<!DOCTYPE html><html><body><h2>Policy Review Due</h2><div style="border-left:4px solid #f57c00;background:#fff3e0;padding:16px;margin:16px 0"><strong>Review Due: {{.DueDate}}</strong></div><p><strong>Policy:</strong> {{.PolicyRef}} — {{.Title}}</p><p><strong>Owner:</strong> {{.OwnerName}}</p><p><strong>Last Reviewed:</strong> {{.LastReviewDate}}</p><p><a href="{{.DashboardURL}}" style="background:#1a237e;color:white;padding:12px 24px;text-decoration:none;border-radius:4px">Review Policy</a></p></body></html>',
    'Policy Review Due\nPolicy: {{.PolicyRef}} — {{.Title}}\nDue: {{.DueDate}}\nOwner: {{.OwnerName}}\n\nReview at: {{.DashboardURL}}',
    'Policy Review Due — {{.PolicyRef}}',
    'Policy "{{.Title}}" is due for review on {{.DueDate}}.',
    '{"blocks":[{"type":"section","text":{"type":"mrkdwn","text":"*Policy Review Due*\\n{{.PolicyRef}} — {{.Title}}\\nDue: {{.DueDate}}"}}]}',
    '{"event":"policy.review_due","policy_ref":"{{.PolicyRef}}","title":"{{.Title}}","due_date":"{{.DueDate}}"}',
    ARRAY['PolicyRef','Title','DueDate','OwnerName','LastReviewDate','DashboardURL'],
    true
);

-- 10. Policy Review Overdue
INSERT INTO notification_templates (id, organization_id, name, event_type, subject_template, body_html_template, body_text_template, in_app_title_template, in_app_body_template, slack_template, webhook_payload_template, variables, is_system)
VALUES (
    'a0000001-0010-4000-8000-000000000001',
    NULL,
    'Policy Review Overdue',
    'policy.review_overdue',
    '[OVERDUE] Policy Review — {{.PolicyRef}}: {{.Title}}',
    '<!DOCTYPE html><html><body><h2>Policy Review Overdue</h2><div style="border-left:4px solid #d32f2f;background:#ffebee;padding:16px;margin:16px 0"><strong>OVERDUE — Original Due Date: {{.DueDate}}</strong></div><p><strong>Policy:</strong> {{.PolicyRef}} — {{.Title}}</p><p><strong>Owner:</strong> {{.OwnerName}}</p><p><strong>Days Overdue:</strong> {{.DaysOverdue}}</p><p><a href="{{.DashboardURL}}" style="background:#d32f2f;color:white;padding:12px 24px;text-decoration:none;border-radius:4px">Review Now</a></p></body></html>',
    'OVERDUE: Policy Review\nPolicy: {{.PolicyRef}} — {{.Title}}\nDue: {{.DueDate}}\nDays Overdue: {{.DaysOverdue}}\n\nReview at: {{.DashboardURL}}',
    'OVERDUE: Policy Review — {{.PolicyRef}}',
    'Policy "{{.Title}}" review is {{.DaysOverdue}} days overdue.',
    '{"blocks":[{"type":"section","text":{"type":"mrkdwn","text":"*OVERDUE: Policy Review*\\n{{.PolicyRef}} — {{.Title}}\\n{{.DaysOverdue}} days overdue"}}]}',
    '{"event":"policy.review_overdue","policy_ref":"{{.PolicyRef}}","title":"{{.Title}}","days_overdue":"{{.DaysOverdue}}"}',
    ARRAY['PolicyRef','Title','DueDate','OwnerName','DaysOverdue','LastReviewDate','DashboardURL'],
    true
);

-- 11. Attestation Required
INSERT INTO notification_templates (id, organization_id, name, event_type, subject_template, body_html_template, body_text_template, in_app_title_template, in_app_body_template, slack_template, webhook_payload_template, variables, is_system)
VALUES (
    'a0000001-0011-4000-8000-000000000001',
    NULL,
    'Policy Attestation Required',
    'attestation.required',
    'Policy Attestation Required — {{.PolicyTitle}}',
    '<!DOCTYPE html><html><body><h2>Policy Attestation Required</h2><div style="border-left:4px solid #1976d2;background:#e3f2fd;padding:16px;margin:16px 0"><strong>Please acknowledge the following policy</strong></div><p><strong>Policy:</strong> {{.PolicyTitle}}</p><p><strong>Version:</strong> {{.VersionLabel}}</p><p><strong>Due By:</strong> {{.DueDate}}</p><p><a href="{{.AttestURL}}" style="background:#1a237e;color:white;padding:12px 24px;text-decoration:none;border-radius:4px">Read and Acknowledge</a></p></body></html>',
    'Policy Attestation Required\nPolicy: {{.PolicyTitle}}\nVersion: {{.VersionLabel}}\nDue: {{.DueDate}}\n\nAcknowledge at: {{.AttestURL}}',
    'Attestation Required — {{.PolicyTitle}}',
    'Please acknowledge policy "{{.PolicyTitle}}" by {{.DueDate}}.',
    '{"blocks":[{"type":"section","text":{"type":"mrkdwn","text":"*Policy Attestation Required*\\n{{.PolicyTitle}} v{{.VersionLabel}}\\nDue: {{.DueDate}}"}}]}',
    '{"event":"attestation.required","policy_title":"{{.PolicyTitle}}","due_date":"{{.DueDate}}"}',
    ARRAY['PolicyTitle','VersionLabel','DueDate','AttestURL','DashboardURL'],
    true
);

-- 12. Audit Finding Created (Critical/High)
INSERT INTO notification_templates (id, organization_id, name, event_type, subject_template, body_html_template, body_text_template, in_app_title_template, in_app_body_template, slack_template, webhook_payload_template, variables, is_system)
VALUES (
    'a0000001-0012-4000-8000-000000000001',
    NULL,
    'Audit Finding Created',
    'finding.created',
    '[{{.Severity}}] New Audit Finding — {{.FindingRef}}: {{.Title}}',
    '<!DOCTYPE html><html><body><h2>New Audit Finding</h2><div style="border-left:4px solid #d32f2f;background:#ffebee;padding:16px;margin:16px 0"><strong>Severity: {{.Severity}}</strong></div><p><strong>Finding:</strong> {{.FindingRef}} — {{.Title}}</p><p><strong>Audit:</strong> {{.AuditRef}} — {{.AuditTitle}}</p><p><strong>Responsible:</strong> {{.ResponsibleName}}</p><p><strong>Due Date:</strong> {{.DueDate}}</p><p><a href="{{.DashboardURL}}" style="background:#1a237e;color:white;padding:12px 24px;text-decoration:none;border-radius:4px">View Finding</a></p></body></html>',
    'New Audit Finding: {{.FindingRef}}\nSeverity: {{.Severity}}\nAudit: {{.AuditRef}}\nDue: {{.DueDate}}\n\nView: {{.DashboardURL}}',
    'New Finding: {{.FindingRef}} ({{.Severity}})',
    '{{.Severity}} finding "{{.Title}}" from audit {{.AuditRef}}. Due: {{.DueDate}}.',
    '{"blocks":[{"type":"section","text":{"type":"mrkdwn","text":"*New Audit Finding*\\n{{.FindingRef}} — {{.Severity}}\\n{{.Title}}"}}]}',
    '{"event":"finding.created","finding_ref":"{{.FindingRef}}","severity":"{{.Severity}}","audit_ref":"{{.AuditRef}}"}',
    ARRAY['FindingRef','Title','Severity','AuditRef','AuditTitle','ResponsibleName','DueDate','DashboardURL'],
    true
);

-- 13. Finding Remediation Overdue
INSERT INTO notification_templates (id, organization_id, name, event_type, subject_template, body_html_template, body_text_template, in_app_title_template, in_app_body_template, slack_template, webhook_payload_template, variables, is_system)
VALUES (
    'a0000001-0013-4000-8000-000000000001',
    NULL,
    'Finding Remediation Overdue',
    'finding.remediation_overdue',
    '[OVERDUE] Audit Finding {{.FindingRef}} — {{.Severity}} — {{.DaysOverdue}} days overdue',
    '<!DOCTYPE html><html><body><h2>Overdue Audit Finding</h2><div style="border-left:4px solid #d32f2f;background:#ffebee;padding:16px;margin:16px 0"><strong>{{.DaysOverdue}} DAYS OVERDUE</strong></div><p><strong>Finding:</strong> {{.FindingRef}} — {{.Title}}</p><p><strong>Severity:</strong> {{.Severity}}</p><p><strong>Original Due Date:</strong> {{.DueDate}}</p><p><strong>Responsible:</strong> {{.ResponsibleName}}</p><p><a href="{{.DashboardURL}}" style="background:#d32f2f;color:white;padding:12px 24px;text-decoration:none;border-radius:4px">Address Finding</a></p></body></html>',
    'OVERDUE: Audit Finding {{.FindingRef}}\nSeverity: {{.Severity}}\nDays Overdue: {{.DaysOverdue}}\nDue: {{.DueDate}}\n\nView: {{.DashboardURL}}',
    'OVERDUE: Finding {{.FindingRef}}',
    'Finding {{.FindingRef}} ({{.Severity}}) is {{.DaysOverdue}} days overdue.',
    '{"blocks":[{"type":"section","text":{"type":"mrkdwn","text":"*OVERDUE Finding*\\n{{.FindingRef}} — {{.Severity}}\\n{{.DaysOverdue}} days overdue"}}]}',
    '{"event":"finding.remediation_overdue","finding_ref":"{{.FindingRef}}","severity":"{{.Severity}}","days_overdue":"{{.DaysOverdue}}"}',
    ARRAY['FindingRef','Title','Severity','DueDate','DaysOverdue','ResponsibleName','AuditRef','DashboardURL'],
    true
);

-- 14. Vendor Assessment Due
INSERT INTO notification_templates (id, organization_id, name, event_type, subject_template, body_html_template, body_text_template, in_app_title_template, in_app_body_template, slack_template, webhook_payload_template, variables, is_system)
VALUES (
    'a0000001-0014-4000-8000-000000000001',
    NULL,
    'Vendor Assessment Due',
    'vendor.assessment_due',
    'Vendor Assessment Due — {{.VendorRef}}: {{.VendorName}}',
    '<!DOCTYPE html><html><body><h2>Vendor Assessment Due</h2><div style="border-left:4px solid #f57c00;background:#fff3e0;padding:16px;margin:16px 0"><strong>Risk Tier: {{.RiskTier}}</strong></div><p><strong>Vendor:</strong> {{.VendorRef}} — {{.VendorName}}</p><p><strong>Assessment Due:</strong> {{.DueDate}}</p><p><strong>Relationship Owner:</strong> {{.OwnerName}}</p><p><a href="{{.DashboardURL}}" style="background:#1a237e;color:white;padding:12px 24px;text-decoration:none;border-radius:4px">Start Assessment</a></p></body></html>',
    'Vendor Assessment Due\nVendor: {{.VendorRef}} — {{.VendorName}}\nRisk Tier: {{.RiskTier}}\nDue: {{.DueDate}}\n\nView: {{.DashboardURL}}',
    'Vendor Assessment Due — {{.VendorName}}',
    'Vendor "{{.VendorName}}" ({{.RiskTier}}) assessment due on {{.DueDate}}.',
    '{"blocks":[{"type":"section","text":{"type":"mrkdwn","text":"*Vendor Assessment Due*\\n{{.VendorRef}} — {{.VendorName}}\\nRisk Tier: {{.RiskTier}}\\nDue: {{.DueDate}}"}}]}',
    '{"event":"vendor.assessment_due","vendor_ref":"{{.VendorRef}}","vendor_name":"{{.VendorName}}","risk_tier":"{{.RiskTier}}"}',
    ARRAY['VendorRef','VendorName','RiskTier','DueDate','OwnerName','DashboardURL'],
    true
);

-- 15. Vendor Missing DPA
INSERT INTO notification_templates (id, organization_id, name, event_type, subject_template, body_html_template, body_text_template, in_app_title_template, in_app_body_template, slack_template, webhook_payload_template, variables, is_system)
VALUES (
    'a0000001-0015-4000-8000-000000000001',
    NULL,
    'Vendor Missing DPA',
    'vendor.missing_dpa',
    '[GDPR] Missing Data Processing Agreement — {{.VendorName}}',
    '<!DOCTYPE html><html><body><h2>Missing Data Processing Agreement</h2><div style="border-left:4px solid #d32f2f;background:#ffebee;padding:16px;margin:16px 0"><strong>GDPR Compliance Issue: No DPA in place</strong></div><p>Vendor <strong>{{.VendorRef}} — {{.VendorName}}</strong> processes personal data but does not have a Data Processing Agreement (DPA) in place.</p><p><strong>Risk Tier:</strong> {{.RiskTier}}</p><p><strong>Data Categories:</strong> {{.DataCategories}}</p><p>Per GDPR Article 28, a DPA must be in place before any personal data processing.</p><p><a href="{{.DashboardURL}}" style="background:#d32f2f;color:white;padding:12px 24px;text-decoration:none;border-radius:4px">Resolve Issue</a></p></body></html>',
    'GDPR: Missing DPA for {{.VendorName}}\nVendor: {{.VendorRef}}\nRisk Tier: {{.RiskTier}}\n\nAddress at: {{.DashboardURL}}',
    'Missing DPA — {{.VendorName}}',
    'Vendor "{{.VendorName}}" processes personal data without a DPA. GDPR compliance issue.',
    '{"blocks":[{"type":"section","text":{"type":"mrkdwn","text":"*Missing DPA: {{.VendorName}}*\\nProcesses personal data without DPA.\\nGDPR Article 28 compliance issue."}}]}',
    '{"event":"vendor.missing_dpa","vendor_ref":"{{.VendorRef}}","vendor_name":"{{.VendorName}}","risk_tier":"{{.RiskTier}}"}',
    ARRAY['VendorRef','VendorName','RiskTier','DataCategories','DashboardURL'],
    true
);

-- 16. Risk Threshold Exceeded
INSERT INTO notification_templates (id, organization_id, name, event_type, subject_template, body_html_template, body_text_template, in_app_title_template, in_app_body_template, slack_template, webhook_payload_template, variables, is_system)
VALUES (
    'a0000001-0016-4000-8000-000000000001',
    NULL,
    'Risk Threshold Exceeded',
    'risk.threshold_exceeded',
    '[RISK ALERT] Risk Threshold Exceeded — {{.RiskRef}}: {{.Title}}',
    '<!DOCTYPE html><html><body><h2>Risk Threshold Exceeded</h2><div style="border-left:4px solid #d32f2f;background:#ffebee;padding:16px;margin:16px 0"><strong>Risk Level: {{.RiskLevel}}</strong></div><p><strong>Risk:</strong> {{.RiskRef}} — {{.Title}}</p><p><strong>Category:</strong> {{.Category}}</p><p><strong>Residual Risk Score:</strong> {{.ResidualScore}}</p><p><strong>Threshold:</strong> {{.Threshold}}</p><p><strong>Owner:</strong> {{.OwnerName}}</p><p><a href="{{.DashboardURL}}" style="background:#1a237e;color:white;padding:12px 24px;text-decoration:none;border-radius:4px">View Risk</a></p></body></html>',
    'Risk Threshold Exceeded\nRisk: {{.RiskRef}} — {{.Title}}\nLevel: {{.RiskLevel}}\nScore: {{.ResidualScore}} (Threshold: {{.Threshold}})\n\nView: {{.DashboardURL}}',
    'Risk Threshold: {{.RiskRef}}',
    'Risk "{{.Title}}" exceeded threshold. Level: {{.RiskLevel}}, Score: {{.ResidualScore}}.',
    '{"blocks":[{"type":"section","text":{"type":"mrkdwn","text":"*Risk Threshold Exceeded*\\n{{.RiskRef}} — {{.Title}}\\nLevel: {{.RiskLevel}} | Score: {{.ResidualScore}}"}}]}',
    '{"event":"risk.threshold_exceeded","risk_ref":"{{.RiskRef}}","title":"{{.Title}}","risk_level":"{{.RiskLevel}}","score":"{{.ResidualScore}}"}',
    ARRAY['RiskRef','Title','RiskLevel','Category','ResidualScore','Threshold','OwnerName','DashboardURL'],
    true
);

-- 17. Compliance Score Dropped
INSERT INTO notification_templates (id, organization_id, name, event_type, subject_template, body_html_template, body_text_template, in_app_title_template, in_app_body_template, slack_template, webhook_payload_template, variables, is_system)
VALUES (
    'a0000001-0017-4000-8000-000000000001',
    NULL,
    'Compliance Score Dropped',
    'compliance.score_dropped',
    'Compliance Score Alert — {{.FrameworkCode}} dropped to {{.CurrentScore}}%',
    '<!DOCTYPE html><html><body><h2>Compliance Score Alert</h2><div style="border-left:4px solid #f57c00;background:#fff3e0;padding:16px;margin:16px 0"><strong>Score decreased by {{.DropAmount}} percentage points</strong></div><p><strong>Framework:</strong> {{.FrameworkCode}} — {{.FrameworkName}}</p><p><strong>Previous Score:</strong> {{.PreviousScore}}%</p><p><strong>Current Score:</strong> {{.CurrentScore}}%</p><p><strong>Drop:</strong> -{{.DropAmount}}%</p><p><a href="{{.DashboardURL}}" style="background:#1a237e;color:white;padding:12px 24px;text-decoration:none;border-radius:4px">View Dashboard</a></p></body></html>',
    'Compliance Score Alert\nFramework: {{.FrameworkCode}}\nPrevious: {{.PreviousScore}}%\nCurrent: {{.CurrentScore}}%\nDrop: -{{.DropAmount}}%\n\nView: {{.DashboardURL}}',
    'Score Drop: {{.FrameworkCode}} at {{.CurrentScore}}%',
    '{{.FrameworkCode}} compliance score dropped from {{.PreviousScore}}% to {{.CurrentScore}}%.',
    '{"blocks":[{"type":"section","text":{"type":"mrkdwn","text":"*Compliance Score Alert*\\n{{.FrameworkCode}}: {{.PreviousScore}}% -> {{.CurrentScore}}% (-{{.DropAmount}}%)"}}]}',
    '{"event":"compliance.score_dropped","framework":"{{.FrameworkCode}}","previous_score":"{{.PreviousScore}}","current_score":"{{.CurrentScore}}"}',
    ARRAY['FrameworkCode','FrameworkName','PreviousScore','CurrentScore','DropAmount','DashboardURL'],
    true
);

-- 18. Welcome Email
INSERT INTO notification_templates (id, organization_id, name, event_type, subject_template, body_html_template, body_text_template, in_app_title_template, in_app_body_template, slack_template, webhook_payload_template, variables, is_system)
VALUES (
    'a0000001-0018-4000-8000-000000000001',
    NULL,
    'Welcome Email',
    'user.welcome',
    'Welcome to ComplianceForge — {{.OrganizationName}}',
    '<!DOCTYPE html><html><body><h2>Welcome to ComplianceForge</h2><p>Hello {{.FirstName}},</p><p>Your account has been created for <strong>{{.OrganizationName}}</strong>.</p><p><strong>Email:</strong> {{.Email}}</p><p><strong>Role:</strong> {{.RoleName}}</p><p>Please set your password using the link below:</p><p><a href="{{.SetPasswordURL}}" style="background:#1a237e;color:white;padding:12px 24px;text-decoration:none;border-radius:4px">Set Your Password</a></p><p style="color:#666">This link expires in 24 hours.</p></body></html>',
    'Welcome to ComplianceForge\n\nHello {{.FirstName}},\n\nYour account for {{.OrganizationName}} has been created.\nEmail: {{.Email}}\nRole: {{.RoleName}}\n\nSet your password: {{.SetPasswordURL}}\n\nThis link expires in 24 hours.',
    'Welcome to ComplianceForge',
    'Your account for {{.OrganizationName}} is ready. Set your password to get started.',
    '{}',
    '{"event":"user.welcome","email":"{{.Email}}","organization":"{{.OrganizationName}}"}',
    ARRAY['FirstName','LastName','Email','OrganizationName','RoleName','SetPasswordURL','DashboardURL'],
    true
);

-- 19. Password Reset
INSERT INTO notification_templates (id, organization_id, name, event_type, subject_template, body_html_template, body_text_template, in_app_title_template, in_app_body_template, slack_template, webhook_payload_template, variables, is_system)
VALUES (
    'a0000001-0019-4000-8000-000000000001',
    NULL,
    'Password Reset',
    'user.password_reset',
    'Password Reset Request — ComplianceForge',
    '<!DOCTYPE html><html><body><h2>Password Reset</h2><p>Hello {{.FirstName}},</p><p>We received a request to reset your password for ComplianceForge.</p><p><a href="{{.ResetURL}}" style="background:#1a237e;color:white;padding:12px 24px;text-decoration:none;border-radius:4px">Reset Password</a></p><p style="color:#666">This link expires in 1 hour. If you did not request this, please ignore this email.</p><p style="color:#666">IP Address: {{.IPAddress}}</p></body></html>',
    'Password Reset Request\n\nHello {{.FirstName}},\nReset your password: {{.ResetURL}}\n\nExpires in 1 hour.\nIP: {{.IPAddress}}',
    'Password Reset Requested',
    'A password reset was requested for your account.',
    '{}',
    '{"event":"user.password_reset","email":"{{.Email}}"}',
    ARRAY['FirstName','Email','ResetURL','IPAddress','DashboardURL'],
    true
);
