package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/complianceforge/platform/internal/middleware"
	"github.com/complianceforge/platform/internal/models"
	"github.com/complianceforge/platform/internal/service"
)

// MobileHandler handles HTTP requests for the mobile-optimised API,
// providing condensed responses suitable for mobile bandwidth and screen sizes,
// plus push notification token and preference management.
type MobileHandler struct {
	db   *pgxpool.Pool
	push *service.PushService
}

// NewMobileHandler creates a new MobileHandler.
func NewMobileHandler(db *pgxpool.Pool, push *service.PushService) *MobileHandler {
	return &MobileHandler{db: db, push: push}
}

// RegisterRoutes mounts all mobile API routes on the router.
func (h *MobileHandler) RegisterRoutes(r chi.Router) {
	// Mobile dashboard & data endpoints (condensed payloads)
	r.Get("/dashboard", h.Dashboard)
	r.Get("/approvals", h.ListApprovals)
	r.Post("/approvals/{id}/approve", h.ApproveRequest)
	r.Post("/approvals/{id}/reject", h.RejectRequest)
	r.Get("/incidents/active", h.ActiveIncidents)
	r.Get("/deadlines", h.UpcomingDeadlines)
	r.Get("/activity", h.RecentActivity)

	// Push notification management
	r.Post("/push/register", h.RegisterPushToken)
	r.Delete("/push/unregister", h.UnregisterPushToken)
	r.Get("/push/preferences", h.GetPushPreferences)
	r.Put("/push/preferences", h.UpdatePushPreferences)
}

// ============================================================
// MOBILE DASHBOARD (condensed)
// ============================================================

// Dashboard returns a condensed mobile dashboard with key metrics.
// GET /mobile/dashboard
func (h *MobileHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())

	type riskSummary struct {
		Critical int `json:"critical"`
		High     int `json:"high"`
		Medium   int `json:"medium"`
		Low      int `json:"low"`
	}

	type mobileDashboard struct {
		ComplianceScore  float64     `json:"compliance_score"`
		RiskCounts       riskSummary `json:"risk_counts"`
		OpenIncidents    int         `json:"open_incidents"`
		PendingApprovals int         `json:"pending_approvals"`
		OverdueDeadlines int         `json:"overdue_deadlines"`
		UnreadNotifs     int         `json:"unread_notifications"`
		LastUpdated      string      `json:"last_updated"`
	}

	dash := mobileDashboard{
		LastUpdated: time.Now().UTC().Format(time.RFC3339),
	}

	// Compliance score
	h.db.QueryRow(r.Context(), `
		SELECT COALESCE(AVG(score), 0)
		FROM compliance_scores
		WHERE organization_id = $1`, orgID,
	).Scan(&dash.ComplianceScore)

	// Risk counts by level
	rows, err := h.db.Query(r.Context(), `
		SELECT COALESCE(residual_risk_level, inherent_risk_level, 'low'), COUNT(*)
		FROM risks
		WHERE organization_id = $1 AND status NOT IN ('closed', 'accepted')
			AND deleted_at IS NULL
		GROUP BY COALESCE(residual_risk_level, inherent_risk_level, 'low')`, orgID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var level string
			var count int
			if rows.Scan(&level, &count) == nil {
				switch level {
				case "critical":
					dash.RiskCounts.Critical = count
				case "high":
					dash.RiskCounts.High = count
				case "medium":
					dash.RiskCounts.Medium = count
				case "low":
					dash.RiskCounts.Low = count
				}
			}
		}
	}

	// Open incidents
	h.db.QueryRow(r.Context(), `
		SELECT COUNT(*) FROM incidents
		WHERE organization_id = $1 AND status IN ('open', 'investigating', 'contained')
			AND deleted_at IS NULL`, orgID,
	).Scan(&dash.OpenIncidents)

	// Pending approvals for this user
	h.db.QueryRow(r.Context(), `
		SELECT COUNT(*) FROM workflow_step_executions wse
		JOIN workflow_instances wi ON wse.instance_id = wi.id
		WHERE wi.organization_id = $1
			AND wse.assigned_to = $2
			AND wse.status = 'pending'`, orgID, userID,
	).Scan(&dash.PendingApprovals)

	// Overdue deadlines (policies, findings, assessments)
	h.db.QueryRow(r.Context(), `
		SELECT COUNT(*) FROM (
			SELECT id FROM policies
			WHERE organization_id = $1 AND next_review_date < NOW() AND deleted_at IS NULL
			UNION ALL
			SELECT id FROM audit_findings
			WHERE organization_id = $1 AND remediation_deadline < NOW()
				AND status NOT IN ('closed', 'accepted') AND deleted_at IS NULL
		) overdue`, orgID,
	).Scan(&dash.OverdueDeadlines)

	// Unread notifications
	h.db.QueryRow(r.Context(), `
		SELECT COUNT(*) FROM notifications
		WHERE recipient_user_id = $1 AND organization_id = $2
			AND channel_type = 'in_app' AND read_at IS NULL`,
		userID, orgID,
	).Scan(&dash.UnreadNotifs)

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    dash,
	})
}

// ============================================================
// APPROVALS (condensed)
// ============================================================

// ListApprovals returns condensed pending approvals for the current user.
// GET /mobile/approvals
func (h *MobileHandler) ListApprovals(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	params := parsePagination(r)

	type approvalItem struct {
		ID           uuid.UUID `json:"id"`
		InstanceID   uuid.UUID `json:"instance_id"`
		StepName     string    `json:"step_name"`
		WorkflowName string    `json:"workflow_name"`
		EntityType   string    `json:"entity_type"`
		EntityName   string    `json:"entity_name"`
		RequestedBy  string    `json:"requested_by"`
		RequestedAt  time.Time `json:"requested_at"`
		Priority     string    `json:"priority"`
	}

	rows, err := h.db.Query(r.Context(), `
		SELECT wse.id, wse.instance_id, wse.step_name,
			COALESCE(wd.name, 'Workflow'), wi.entity_type,
			COALESCE(wi.entity_name, ''),
			COALESCE(u.full_name, u.email, 'Unknown'),
			wse.created_at,
			COALESCE(wse.priority, 'normal')
		FROM workflow_step_executions wse
		JOIN workflow_instances wi ON wse.instance_id = wi.id
		LEFT JOIN workflow_definitions wd ON wi.definition_id = wd.id
		LEFT JOIN users u ON wi.initiated_by = u.id
		WHERE wi.organization_id = $1
			AND wse.assigned_to = $2
			AND wse.status = 'pending'
		ORDER BY wse.created_at DESC
		LIMIT $3 OFFSET $4`,
		orgID, userID, params.PageSize, params.Offset(),
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list approvals")
		return
	}
	defer rows.Close()

	var approvals []approvalItem
	for rows.Next() {
		var a approvalItem
		if err := rows.Scan(
			&a.ID, &a.InstanceID, &a.StepName, &a.WorkflowName,
			&a.EntityType, &a.EntityName, &a.RequestedBy,
			&a.RequestedAt, &a.Priority,
		); err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to scan approval")
			return
		}
		approvals = append(approvals, a)
	}

	if approvals == nil {
		approvals = []approvalItem{}
	}

	var total int64
	h.db.QueryRow(r.Context(), `
		SELECT COUNT(*) FROM workflow_step_executions wse
		JOIN workflow_instances wi ON wse.instance_id = wi.id
		WHERE wi.organization_id = $1
			AND wse.assigned_to = $2
			AND wse.status = 'pending'`,
		orgID, userID,
	).Scan(&total)

	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize != 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: approvals,
		Pagination: models.Pagination{
			Page:       params.Page,
			PageSize:   params.PageSize,
			TotalItems: total,
			TotalPages: totalPages,
			HasNext:    params.Page < totalPages,
			HasPrev:    params.Page > 1,
		},
	})
}

// ApproveRequest approves a pending workflow step execution.
// POST /mobile/approvals/{id}/approve
func (h *MobileHandler) ApproveRequest(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	execID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid execution ID")
		return
	}

	var req struct {
		Comment string `json:"comment"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	result, err := h.db.Exec(r.Context(), `
		UPDATE workflow_step_executions SET
			status = 'approved',
			completed_by = $1,
			completed_at = NOW(),
			comment = $2,
			updated_at = NOW()
		WHERE id = $3 AND assigned_to = $1 AND status = 'pending'
			AND instance_id IN (SELECT id FROM workflow_instances WHERE organization_id = $4)`,
		userID, req.Comment, execID, orgID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to approve request")
		return
	}

	if result.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Approval not found or not assigned to you")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"id":          execID,
			"status":      "approved",
			"approved_at": time.Now(),
		},
	})
}

// RejectRequest rejects a pending workflow step execution.
// POST /mobile/approvals/{id}/reject
func (h *MobileHandler) RejectRequest(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	execID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid execution ID")
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}
	if req.Reason == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "reason is required for rejection")
		return
	}

	result, err := h.db.Exec(r.Context(), `
		UPDATE workflow_step_executions SET
			status = 'rejected',
			completed_by = $1,
			completed_at = NOW(),
			comment = $2,
			updated_at = NOW()
		WHERE id = $3 AND assigned_to = $1 AND status = 'pending'
			AND instance_id IN (SELECT id FROM workflow_instances WHERE organization_id = $4)`,
		userID, req.Reason, execID, orgID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to reject request")
		return
	}

	if result.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Approval not found or not assigned to you")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"id":          execID,
			"status":      "rejected",
			"rejected_at": time.Now(),
		},
	})
}

// ============================================================
// INCIDENTS (condensed)
// ============================================================

// ActiveIncidents returns condensed active incidents.
// GET /mobile/incidents/active
func (h *MobileHandler) ActiveIncidents(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	params := parsePagination(r)

	type incidentItem struct {
		ID           uuid.UUID `json:"id"`
		Reference    string    `json:"reference"`
		Title        string    `json:"title"`
		Severity     string    `json:"severity"`
		Status       string    `json:"status"`
		Category     string    `json:"category"`
		AssignedTo   string    `json:"assigned_to"`
		ReportedAt   time.Time `json:"reported_at"`
		IsBreachable bool      `json:"is_breachable"`
	}

	rows, err := h.db.Query(r.Context(), `
		SELECT i.id, i.reference, i.title, i.severity, i.status,
			COALESCE(i.category, ''),
			COALESCE(u.full_name, u.email, 'Unassigned'),
			i.created_at,
			COALESCE(i.is_data_breach, false)
		FROM incidents i
		LEFT JOIN users u ON i.assigned_to = u.id
		WHERE i.organization_id = $1
			AND i.status IN ('open', 'investigating', 'contained')
			AND i.deleted_at IS NULL
		ORDER BY
			CASE i.severity
				WHEN 'critical' THEN 1
				WHEN 'high' THEN 2
				WHEN 'medium' THEN 3
				WHEN 'low' THEN 4
				ELSE 5
			END,
			i.created_at DESC
		LIMIT $2 OFFSET $3`,
		orgID, params.PageSize, params.Offset(),
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list active incidents")
		return
	}
	defer rows.Close()

	var incidents []incidentItem
	for rows.Next() {
		var inc incidentItem
		if err := rows.Scan(
			&inc.ID, &inc.Reference, &inc.Title, &inc.Severity,
			&inc.Status, &inc.Category, &inc.AssignedTo,
			&inc.ReportedAt, &inc.IsBreachable,
		); err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to scan incident")
			return
		}
		incidents = append(incidents, inc)
	}

	if incidents == nil {
		incidents = []incidentItem{}
	}

	var total int64
	h.db.QueryRow(r.Context(), `
		SELECT COUNT(*) FROM incidents
		WHERE organization_id = $1
			AND status IN ('open', 'investigating', 'contained')
			AND deleted_at IS NULL`, orgID,
	).Scan(&total)

	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize != 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: incidents,
		Pagination: models.Pagination{
			Page:       params.Page,
			PageSize:   params.PageSize,
			TotalItems: total,
			TotalPages: totalPages,
			HasNext:    params.Page < totalPages,
			HasPrev:    params.Page > 1,
		},
	})
}

// ============================================================
// DEADLINES (condensed)
// ============================================================

// UpcomingDeadlines returns upcoming deadlines within a configurable window.
// GET /mobile/deadlines?days=7
func (h *MobileHandler) UpcomingDeadlines(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	days := 7
	if d, err := strconv.Atoi(r.URL.Query().Get("days")); err == nil && d > 0 && d <= 90 {
		days = d
	}

	type deadlineItem struct {
		ID       uuid.UUID `json:"id"`
		Type     string    `json:"type"`
		Title    string    `json:"title"`
		DueDate  time.Time `json:"due_date"`
		DaysLeft int       `json:"days_left"`
		Priority string    `json:"priority"`
		Overdue  bool      `json:"overdue"`
	}

	rows, err := h.db.Query(r.Context(), `
		SELECT id, type, title, due_date FROM (
			-- Policy reviews
			SELECT id, 'policy_review' AS type, title,
				next_review_date AS due_date
			FROM policies
			WHERE organization_id = $1 AND deleted_at IS NULL
				AND next_review_date IS NOT NULL
				AND next_review_date <= NOW() + ($2 || ' days')::interval

			UNION ALL

			-- Finding remediations
			SELECT id, 'finding_remediation' AS type, title,
				remediation_deadline AS due_date
			FROM audit_findings
			WHERE organization_id = $1 AND deleted_at IS NULL
				AND status NOT IN ('closed', 'accepted')
				AND remediation_deadline IS NOT NULL
				AND remediation_deadline <= NOW() + ($2 || ' days')::interval

			UNION ALL

			-- Risk reviews
			SELECT id, 'risk_review' AS type, title,
				next_review_date AS due_date
			FROM risks
			WHERE organization_id = $1 AND deleted_at IS NULL
				AND status NOT IN ('closed')
				AND next_review_date IS NOT NULL
				AND next_review_date <= NOW() + ($2 || ' days')::interval
		) deadlines
		ORDER BY due_date ASC
		LIMIT 50`,
		orgID, strconv.Itoa(days),
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve deadlines")
		return
	}
	defer rows.Close()

	now := time.Now()
	var deadlines []deadlineItem
	for rows.Next() {
		var d deadlineItem
		if err := rows.Scan(&d.ID, &d.Type, &d.Title, &d.DueDate); err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to scan deadline")
			return
		}
		daysLeft := int(d.DueDate.Sub(now).Hours() / 24)
		d.DaysLeft = daysLeft
		d.Overdue = daysLeft < 0

		if daysLeft < 0 {
			d.Priority = "critical"
		} else if daysLeft <= 2 {
			d.Priority = "high"
		} else if daysLeft <= 7 {
			d.Priority = "medium"
		} else {
			d.Priority = "low"
		}

		deadlines = append(deadlines, d)
	}

	if deadlines == nil {
		deadlines = []deadlineItem{}
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"deadlines":  deadlines,
			"total":      len(deadlines),
			"days_ahead": days,
		},
	})
}

// ============================================================
// ACTIVITY FEED (condensed)
// ============================================================

// RecentActivity returns condensed recent activity for the org.
// GET /mobile/activity?limit=20
func (h *MobileHandler) RecentActivity(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	limit := 20
	if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && l > 0 && l <= 50 {
		limit = l
	}

	type activityItem struct {
		ID        uuid.UUID `json:"id"`
		Type      string    `json:"type"`
		Title     string    `json:"title"`
		Actor     string    `json:"actor"`
		Timestamp time.Time `json:"timestamp"`
	}

	rows, err := h.db.Query(r.Context(), `
		SELECT id, type, title, actor, timestamp FROM (
			-- Recent incidents
			SELECT i.id, 'incident' AS type,
				i.title,
				COALESCE(u.full_name, u.email, 'System') AS actor,
				i.created_at AS timestamp
			FROM incidents i
			LEFT JOIN users u ON i.reported_by = u.id
			WHERE i.organization_id = $1 AND i.deleted_at IS NULL

			UNION ALL

			-- Recent policy changes
			SELECT p.id, 'policy' AS type,
				p.title,
				COALESCE(u.full_name, u.email, 'System') AS actor,
				p.updated_at AS timestamp
			FROM policies p
			LEFT JOIN users u ON p.owner_id = u.id
			WHERE p.organization_id = $1 AND p.deleted_at IS NULL
				AND p.updated_at > NOW() - interval '30 days'

			UNION ALL

			-- Recent risk updates
			SELECT r.id, 'risk' AS type,
				r.title,
				COALESCE(u.full_name, u.email, 'System') AS actor,
				r.updated_at AS timestamp
			FROM risks r
			LEFT JOIN users u ON r.owner_id = u.id
			WHERE r.organization_id = $1 AND r.deleted_at IS NULL
				AND r.updated_at > NOW() - interval '30 days'
		) activity
		ORDER BY timestamp DESC
		LIMIT $2`,
		orgID, limit,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve recent activity")
		return
	}
	defer rows.Close()

	var activities []activityItem
	for rows.Next() {
		var a activityItem
		if err := rows.Scan(&a.ID, &a.Type, &a.Title, &a.Actor, &a.Timestamp); err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to scan activity")
			return
		}
		activities = append(activities, a)
	}

	if activities == nil {
		activities = []activityItem{}
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"activities": activities,
			"total":      len(activities),
		},
	})
}

// ============================================================
// PUSH NOTIFICATION TOKEN MANAGEMENT
// ============================================================

// RegisterPushToken registers a device push notification token.
// POST /mobile/push/register
func (h *MobileHandler) RegisterPushToken(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())

	var req struct {
		Platform    string `json:"platform"`
		Token       string `json:"token"`
		DeviceName  string `json:"device_name"`
		DeviceModel string `json:"device_model"`
		OSVersion   string `json:"os_version"`
		AppVersion  string `json:"app_version"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Token == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "token is required")
		return
	}
	if req.Platform == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "platform is required (ios, android, or web)")
		return
	}

	info := service.DeviceInfo{
		DeviceName:  req.DeviceName,
		DeviceModel: req.DeviceModel,
		OSVersion:   req.OSVersion,
		AppVersion:  req.AppVersion,
	}

	token, err := h.push.RegisterToken(r.Context(), orgID, userID, req.Platform, req.Token, info)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to register push token: "+err.Error())
		return
	}

	// Return condensed token info (don't expose raw token back)
	writeJSON(w, http.StatusCreated, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"id":            token.ID,
			"platform":      token.Platform,
			"device_name":   token.DeviceName,
			"device_model":  token.DeviceModel,
			"token_hash":    token.TokenHash,
			"is_active":     token.IsActive,
			"registered_at": token.CreatedAt,
		},
	})
}

// UnregisterPushToken deactivates a device push notification token.
// DELETE /mobile/push/unregister
func (h *MobileHandler) UnregisterPushToken(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())

	var req struct {
		TokenHash string `json:"token_hash"`
		Token     string `json:"token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	// Accept either token_hash directly or raw token (we hash it)
	tokenHash := req.TokenHash
	if tokenHash == "" && req.Token != "" {
		tokenHash = service.HashPushToken(req.Token)
	}
	if tokenHash == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "token_hash or token is required")
		return
	}

	if err := h.push.UnregisterToken(r.Context(), orgID, userID, tokenHash); err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Push token not found")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Push token unregistered"},
	})
}

// ============================================================
// PUSH NOTIFICATION PREFERENCES
// ============================================================

// GetPushPreferences returns the user's mobile push notification preferences.
// GET /mobile/push/preferences
func (h *MobileHandler) GetPushPreferences(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())

	prefs, err := h.push.GetPreferences(r.Context(), orgID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get push preferences")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    prefs,
	})
}

// UpdatePushPreferences updates the user's mobile push notification preferences.
// PUT /mobile/push/preferences
func (h *MobileHandler) UpdatePushPreferences(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())

	var req struct {
		PushEnabled           *bool   `json:"push_enabled"`
		PushApprovalRequests  *bool   `json:"push_approval_requests"`
		PushIncidentAlerts    *bool   `json:"push_incident_alerts"`
		PushDeadlineReminders *bool   `json:"push_deadline_reminders"`
		PushMentions          *bool   `json:"push_mentions"`
		PushComments          *bool   `json:"push_comments"`
		QuietHoursEnabled     *bool   `json:"quiet_hours_enabled"`
		QuietHoursStart       *string `json:"quiet_hours_start"`
		QuietHoursEnd         *string `json:"quiet_hours_end"`
		QuietHoursTimezone    *string `json:"quiet_hours_timezone"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	// Load current preferences to apply partial updates
	current, err := h.push.GetPreferences(r.Context(), orgID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to load current preferences")
		return
	}

	// Apply non-nil updates
	if req.PushEnabled != nil {
		current.PushEnabled = *req.PushEnabled
	}
	if req.PushApprovalRequests != nil {
		current.PushApprovalRequests = *req.PushApprovalRequests
	}
	if req.PushIncidentAlerts != nil {
		current.PushIncidentAlerts = *req.PushIncidentAlerts
	}
	if req.PushDeadlineReminders != nil {
		current.PushDeadlineReminders = *req.PushDeadlineReminders
	}
	if req.PushMentions != nil {
		current.PushMentions = *req.PushMentions
	}
	if req.PushComments != nil {
		current.PushComments = *req.PushComments
	}
	if req.QuietHoursEnabled != nil {
		current.QuietHoursEnabled = *req.QuietHoursEnabled
	}
	if req.QuietHoursStart != nil {
		current.QuietHoursStart = *req.QuietHoursStart
	}
	if req.QuietHoursEnd != nil {
		current.QuietHoursEnd = *req.QuietHoursEnd
	}
	if req.QuietHoursTimezone != nil {
		current.QuietHoursTimezone = *req.QuietHoursTimezone
	}

	// Breach alerts always stay on
	current.PushBreachAlerts = true

	updated, err := h.push.UpdatePreferences(r.Context(), orgID, userID, current)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update push preferences")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    updated,
	})
}
