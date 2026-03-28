-- ============================================================
-- ComplianceForge GRC Platform
-- Seed 016: Default DSR Response Templates
-- Uses Go text/template syntax for dynamic field substitution
-- ============================================================

-- ============================================================
-- ACCESS REQUEST — Acknowledgment
-- ============================================================
INSERT INTO dsr_response_templates (id, organization_id, request_type, name, subject, body_html, body_text, is_system, language)
VALUES (
    gen_random_uuid(),
    NULL,
    'access',
    'Access Request — Acknowledgment',
    'Your Data Access Request Has Been Received (Ref: {{.RequestRef}})',
    '<html><body>
<h2>Data Subject Access Request — Acknowledgment</h2>
<p>Dear {{.DataSubjectName}},</p>
<p>We acknowledge receipt of your data subject access request submitted on <strong>{{.ReceivedDate}}</strong>.</p>
<p><strong>Reference Number:</strong> {{.RequestRef}}</p>
<p>Under the General Data Protection Regulation (GDPR), we are required to respond to your request within <strong>30 calendar days</strong> of receipt. Your response deadline is <strong>{{.ResponseDeadline}}</strong>.</p>
<p>Before we can process your request, we may need to verify your identity. If additional verification is required, we will contact you promptly.</p>
<h3>What Happens Next</h3>
<ol>
<li>Identity verification (if required)</li>
<li>Identification and collection of your personal data across our systems</li>
<li>Review of the data to ensure no third-party data is disclosed</li>
<li>Compilation and delivery of your data in a readable format</li>
</ol>
<p>If you have any questions about your request, please contact our Data Protection Officer at <strong>{{.DPOEmail}}</strong>.</p>
<p>Kind regards,<br/>Data Protection Team<br/>{{.OrganizationName}}</p>
</body></html>',
    'Data Subject Access Request — Acknowledgment

Dear {{.DataSubjectName}},

We acknowledge receipt of your data subject access request submitted on {{.ReceivedDate}}.

Reference Number: {{.RequestRef}}

Under the GDPR, we are required to respond within 30 calendar days. Your response deadline is {{.ResponseDeadline}}.

Before processing, we may need to verify your identity.

What Happens Next:
1. Identity verification (if required)
2. Identification and collection of your personal data
3. Review to ensure no third-party data is disclosed
4. Compilation and delivery of your data

Questions? Contact our DPO at {{.DPOEmail}}.

Kind regards,
Data Protection Team
{{.OrganizationName}}',
    true,
    'en'
);

-- ============================================================
-- ERASURE REQUEST — Confirmation
-- ============================================================
INSERT INTO dsr_response_templates (id, organization_id, request_type, name, subject, body_html, body_text, is_system, language)
VALUES (
    gen_random_uuid(),
    NULL,
    'erasure',
    'Erasure Request — Confirmation',
    'Your Data Erasure Request Has Been Completed (Ref: {{.RequestRef}})',
    '<html><body>
<h2>Data Erasure Request — Confirmation</h2>
<p>Dear {{.DataSubjectName}},</p>
<p>We are writing to confirm that your request for erasure of personal data (Reference: <strong>{{.RequestRef}}</strong>) has been completed.</p>
<p><strong>Erasure Completed On:</strong> {{.CompletedDate}}</p>
<h3>Actions Taken</h3>
<ul>
<li>Your personal data has been erased from our active systems</li>
<li>Backup copies will be overwritten in accordance with our retention schedule</li>
{{if .ThirdPartiesNotified}}<li>The following third parties have been notified of the erasure: {{.ThirdPartiesNotified}}</li>{{end}}
</ul>
<h3>Data Retained Under Legal Exemption</h3>
{{if .RetainedData}}
<p>The following data has been retained as permitted under GDPR Article 17(3):</p>
<p>{{.RetainedData}}</p>
<p><strong>Legal Basis:</strong> {{.RetentionBasis}}</p>
{{else}}
<p>No personal data has been retained.</p>
{{end}}
<p>If you have any questions, please contact our Data Protection Officer at <strong>{{.DPOEmail}}</strong>.</p>
<p>Kind regards,<br/>Data Protection Team<br/>{{.OrganizationName}}</p>
</body></html>',
    'Data Erasure Request — Confirmation

Dear {{.DataSubjectName}},

We confirm that your erasure request (Ref: {{.RequestRef}}) has been completed.

Erasure Completed On: {{.CompletedDate}}

Actions Taken:
- Personal data erased from active systems
- Backup copies will be overwritten per retention schedule
{{if .ThirdPartiesNotified}}- Third parties notified: {{.ThirdPartiesNotified}}{{end}}

{{if .RetainedData}}Data retained under legal exemption:
{{.RetainedData}}
Legal Basis: {{.RetentionBasis}}
{{else}}No personal data has been retained.{{end}}

Contact our DPO at {{.DPOEmail}} with questions.

Kind regards,
Data Protection Team
{{.OrganizationName}}',
    true,
    'en'
);

-- ============================================================
-- RECTIFICATION REQUEST — Confirmation
-- ============================================================
INSERT INTO dsr_response_templates (id, organization_id, request_type, name, subject, body_html, body_text, is_system, language)
VALUES (
    gen_random_uuid(),
    NULL,
    'rectification',
    'Rectification Request — Confirmation',
    'Your Data Rectification Request Has Been Completed (Ref: {{.RequestRef}})',
    '<html><body>
<h2>Data Rectification Request — Confirmation</h2>
<p>Dear {{.DataSubjectName}},</p>
<p>We are pleased to confirm that your request for rectification of personal data (Reference: <strong>{{.RequestRef}}</strong>) has been completed.</p>
<p><strong>Rectification Completed On:</strong> {{.CompletedDate}}</p>
<h3>Changes Made</h3>
<p>The following corrections have been applied to your personal data:</p>
<p>{{.RectificationDetails}}</p>
{{if .ThirdPartiesNotified}}
<h3>Third-Party Notification</h3>
<p>In accordance with GDPR Article 19, we have notified the following recipients of the rectification: {{.ThirdPartiesNotified}}</p>
{{end}}
<p>If you believe any data remains inaccurate, please contact our Data Protection Officer at <strong>{{.DPOEmail}}</strong>.</p>
<p>Kind regards,<br/>Data Protection Team<br/>{{.OrganizationName}}</p>
</body></html>',
    'Data Rectification Request — Confirmation

Dear {{.DataSubjectName}},

Your rectification request (Ref: {{.RequestRef}}) has been completed.

Rectification Completed On: {{.CompletedDate}}

Changes Made:
{{.RectificationDetails}}

{{if .ThirdPartiesNotified}}Third-party notification per GDPR Article 19: {{.ThirdPartiesNotified}}{{end}}

Contact our DPO at {{.DPOEmail}} if data remains inaccurate.

Kind regards,
Data Protection Team
{{.OrganizationName}}',
    true,
    'en'
);

-- ============================================================
-- PORTABILITY REQUEST — Response
-- ============================================================
INSERT INTO dsr_response_templates (id, organization_id, request_type, name, subject, body_html, body_text, is_system, language)
VALUES (
    gen_random_uuid(),
    NULL,
    'portability',
    'Data Portability — Response',
    'Your Data Portability Request (Ref: {{.RequestRef}})',
    '<html><body>
<h2>Data Portability Request — Response</h2>
<p>Dear {{.DataSubjectName}},</p>
<p>In response to your data portability request (Reference: <strong>{{.RequestRef}}</strong>), please find enclosed your personal data in a structured, commonly used, and machine-readable format as required by GDPR Article 20.</p>
<h3>Data Provided</h3>
<ul>
<li><strong>Format:</strong> {{.DataFormat}}</li>
<li><strong>Categories Included:</strong> {{.DataCategories}}</li>
<li><strong>Systems Covered:</strong> {{.SystemsCovered}}</li>
</ul>
{{if .DownloadLink}}
<p><strong>Download Link:</strong> <a href="{{.DownloadLink}}">{{.DownloadLink}}</a></p>
<p>This link will expire on <strong>{{.LinkExpiry}}</strong>.</p>
{{end}}
<p>If you wish to have this data transmitted directly to another controller, please provide the details and we will facilitate the transfer where technically feasible.</p>
<p>Questions? Contact our DPO at <strong>{{.DPOEmail}}</strong>.</p>
<p>Kind regards,<br/>Data Protection Team<br/>{{.OrganizationName}}</p>
</body></html>',
    'Data Portability Request — Response

Dear {{.DataSubjectName}},

Re: Data Portability Request (Ref: {{.RequestRef}})

Your personal data is provided in a structured, machine-readable format per GDPR Article 20.

Data Provided:
- Format: {{.DataFormat}}
- Categories: {{.DataCategories}}
- Systems: {{.SystemsCovered}}

{{if .DownloadLink}}Download: {{.DownloadLink}} (expires {{.LinkExpiry}}){{end}}

To transfer data to another controller, provide their details.

Contact our DPO at {{.DPOEmail}} with questions.

Kind regards,
Data Protection Team
{{.OrganizationName}}',
    true,
    'en'
);

-- ============================================================
-- RESTRICTION REQUEST — Confirmation
-- ============================================================
INSERT INTO dsr_response_templates (id, organization_id, request_type, name, subject, body_html, body_text, is_system, language)
VALUES (
    gen_random_uuid(),
    NULL,
    'restriction',
    'Restriction of Processing — Confirmation',
    'Restriction of Processing Confirmed (Ref: {{.RequestRef}})',
    '<html><body>
<h2>Restriction of Processing — Confirmation</h2>
<p>Dear {{.DataSubjectName}},</p>
<p>We confirm that processing of your personal data has been restricted as requested (Reference: <strong>{{.RequestRef}}</strong>), in accordance with GDPR Article 18.</p>
<p><strong>Restriction Applied On:</strong> {{.CompletedDate}}</p>
<h3>Scope of Restriction</h3>
<p>The restriction applies to the following processing activities:</p>
<p>{{.RestrictionScope}}</p>
<h3>What This Means</h3>
<ul>
<li>Your data will be stored but not further processed (except with your consent, for legal claims, for the protection of another person''s rights, or for important public interest reasons)</li>
<li>We will inform you before the restriction is lifted</li>
{{if .ThirdPartiesNotified}}<li>The following recipients have been notified: {{.ThirdPartiesNotified}}</li>{{end}}
</ul>
<p>Contact our DPO at <strong>{{.DPOEmail}}</strong> with any questions.</p>
<p>Kind regards,<br/>Data Protection Team<br/>{{.OrganizationName}}</p>
</body></html>',
    'Restriction of Processing — Confirmation

Dear {{.DataSubjectName}},

Processing of your personal data has been restricted (Ref: {{.RequestRef}}) per GDPR Article 18.

Restriction Applied On: {{.CompletedDate}}

Scope: {{.RestrictionScope}}

Your data will be stored but not further processed except as permitted by law.
We will inform you before any restriction is lifted.
{{if .ThirdPartiesNotified}}Recipients notified: {{.ThirdPartiesNotified}}{{end}}

Contact our DPO at {{.DPOEmail}} with questions.

Kind regards,
Data Protection Team
{{.OrganizationName}}',
    true,
    'en'
);

-- ============================================================
-- OBJECTION — Response
-- ============================================================
INSERT INTO dsr_response_templates (id, organization_id, request_type, name, subject, body_html, body_text, is_system, language)
VALUES (
    gen_random_uuid(),
    NULL,
    'objection',
    'Objection to Processing — Response',
    'Response to Your Objection to Processing (Ref: {{.RequestRef}})',
    '<html><body>
<h2>Objection to Processing — Response</h2>
<p>Dear {{.DataSubjectName}},</p>
<p>We are writing in response to your objection to the processing of your personal data (Reference: <strong>{{.RequestRef}}</strong>), submitted under GDPR Article 21.</p>
{{if .ObjectionUpheld}}
<h3>Objection Upheld</h3>
<p>We have reviewed your objection and confirm that we have ceased the processing activities you objected to, effective <strong>{{.CompletedDate}}</strong>.</p>
<p>The following processing activities have been stopped:</p>
<p>{{.ProcessingActivitiesStopped}}</p>
{{else}}
<h3>Objection Not Upheld</h3>
<p>After careful consideration, we have determined that compelling legitimate grounds exist for continuing the processing that override your interests, rights, and freedoms.</p>
<p><strong>Grounds for Continued Processing:</strong></p>
<p>{{.ContinuedProcessingGrounds}}</p>
<p>You have the right to lodge a complaint with the supervisory authority if you disagree with this decision.</p>
{{end}}
<p>Contact our DPO at <strong>{{.DPOEmail}}</strong> with any questions.</p>
<p>Kind regards,<br/>Data Protection Team<br/>{{.OrganizationName}}</p>
</body></html>',
    'Objection to Processing — Response

Dear {{.DataSubjectName}},

Re: Objection to Processing (Ref: {{.RequestRef}}) under GDPR Article 21.

{{if .ObjectionUpheld}}Objection Upheld
We have ceased the processing activities effective {{.CompletedDate}}.
Stopped: {{.ProcessingActivitiesStopped}}
{{else}}Objection Not Upheld
Compelling legitimate grounds exist for continued processing.
Grounds: {{.ContinuedProcessingGrounds}}
You may lodge a complaint with the supervisory authority.
{{end}}

Contact our DPO at {{.DPOEmail}} with questions.

Kind regards,
Data Protection Team
{{.OrganizationName}}',
    true,
    'en'
);

-- ============================================================
-- EXTENSION NOTICE
-- ============================================================
INSERT INTO dsr_response_templates (id, organization_id, request_type, name, subject, body_html, body_text, is_system, language)
VALUES (
    gen_random_uuid(),
    NULL,
    'access',
    'Deadline Extension Notice',
    'Extension of Response Deadline for Your Request (Ref: {{.RequestRef}})',
    '<html><body>
<h2>Response Deadline Extension Notice</h2>
<p>Dear {{.DataSubjectName}},</p>
<p>We are writing to inform you that we require additional time to respond to your data subject request (Reference: <strong>{{.RequestRef}}</strong>).</p>
<p>In accordance with GDPR Article 12(3), we are permitted to extend the response period by a further <strong>two months</strong> where requests are complex or numerous.</p>
<h3>Extension Details</h3>
<ul>
<li><strong>Original Deadline:</strong> {{.OriginalDeadline}}</li>
<li><strong>Extended Deadline:</strong> {{.ExtendedDeadline}}</li>
<li><strong>Reason for Extension:</strong> {{.ExtensionReason}}</li>
</ul>
<p>We apologise for the delay and assure you that we are working diligently to fulfil your request.</p>
<p>You have the right to lodge a complaint with the supervisory authority if you are dissatisfied with this extension.</p>
<p>Contact our DPO at <strong>{{.DPOEmail}}</strong> with any questions.</p>
<p>Kind regards,<br/>Data Protection Team<br/>{{.OrganizationName}}</p>
</body></html>',
    'Response Deadline Extension Notice

Dear {{.DataSubjectName}},

We require additional time for your request (Ref: {{.RequestRef}}).

Per GDPR Article 12(3), we are extending the response period by two months.

Original Deadline: {{.OriginalDeadline}}
Extended Deadline: {{.ExtendedDeadline}}
Reason: {{.ExtensionReason}}

You have the right to lodge a complaint with the supervisory authority.

Contact our DPO at {{.DPOEmail}} with questions.

Kind regards,
Data Protection Team
{{.OrganizationName}}',
    true,
    'en'
);

-- ============================================================
-- REJECTION LETTER
-- ============================================================
INSERT INTO dsr_response_templates (id, organization_id, request_type, name, subject, body_html, body_text, is_system, language)
VALUES (
    gen_random_uuid(),
    NULL,
    'access',
    'Request Rejection Letter',
    'Regarding Your Data Subject Request (Ref: {{.RequestRef}})',
    '<html><body>
<h2>Data Subject Request — Unable to Fulfil</h2>
<p>Dear {{.DataSubjectName}},</p>
<p>We are writing regarding your data subject request (Reference: <strong>{{.RequestRef}}</strong>) received on <strong>{{.ReceivedDate}}</strong>.</p>
<p>After careful review, we are unable to fulfil your request for the following reason(s):</p>
<h3>Reason for Rejection</h3>
<p>{{.RejectionReason}}</p>
<h3>Legal Basis</h3>
<p>{{.RejectionLegalBasis}}</p>
<h3>Your Rights</h3>
<p>You have the right to:</p>
<ul>
<li>Lodge a complaint with the supervisory authority ({{.SupervisoryAuthority}})</li>
<li>Seek a judicial remedy</li>
<li>Submit a new request if your circumstances change</li>
</ul>
<p>If you believe this decision is incorrect, please contact our Data Protection Officer at <strong>{{.DPOEmail}}</strong> to discuss further.</p>
<p>Kind regards,<br/>Data Protection Team<br/>{{.OrganizationName}}</p>
</body></html>',
    'Data Subject Request — Unable to Fulfil

Dear {{.DataSubjectName}},

Re: Request {{.RequestRef}} received on {{.ReceivedDate}}.

We are unable to fulfil your request.

Reason: {{.RejectionReason}}
Legal Basis: {{.RejectionLegalBasis}}

Your Rights:
- Lodge a complaint with {{.SupervisoryAuthority}}
- Seek a judicial remedy
- Submit a new request if circumstances change

Contact our DPO at {{.DPOEmail}} to discuss.

Kind regards,
Data Protection Team
{{.OrganizationName}}',
    true,
    'en'
);
