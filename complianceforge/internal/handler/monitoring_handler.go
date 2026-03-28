package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/complianceforge/platform/internal/middleware"
	"github.com/complianceforge/platform/internal/models"
	"github.com/complianceforge/platform/internal/service"
)

// MonitoringHandler handles HTTP requests for continuous monitoring,
// automated evidence collection, compliance monitors, and drift detection.
type MonitoringHandler struct {
	collector *service.EvidenceCollector
	monitor   *service.ComplianceMonitor
	drift     *service.DriftDetector
}

// NewMonitoringHandler creates a new MonitoringHandler.
func NewMonitoringHandler(
	collector *service.EvidenceCollector,
	monitor *service.ComplianceMonitor,
	drift *service.DriftDetector,
) *MonitoringHandler {
	return &MonitoringHandler{
		collector: collector,
		monitor:   monitor,
		drift:     drift,
	}
}

// ============================================================
// EVIDENCE COLLECTION CONFIG ENDPOINTS
// ============================================================

// ListConfigs returns paginated evidence collection configurations.
// GET /api/v1/monitoring/configs
func (h *MonitoringHandler) ListConfigs(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	params := parsePagination(r)

	configs, total, err := h.collector.ListConfigs(r.Context(), orgID, params.PageSize, params.Offset())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list configs")
		return
	}

	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize != 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: configs,
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

// CreateConfig creates a new evidence collection configuration.
// POST /api/v1/monitoring/configs
func (h *MonitoringHandler) CreateConfig(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var req service.CollectionConfig
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELD", "Name is required")
		return
	}
	if req.CollectionMethod == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELD", "Collection method is required")
		return
	}

	req.OrganizationID = orgID
	if req.APIConfig == nil {
		req.APIConfig = json.RawMessage(`{}`)
	}
	if req.FileConfig == nil {
		req.FileConfig = json.RawMessage(`{}`)
	}
	if req.ScriptConfig == nil {
		req.ScriptConfig = json.RawMessage(`{}`)
	}
	if req.WebhookConfig == nil {
		req.WebhookConfig = json.RawMessage(`{}`)
	}
	if req.AcceptanceCriteria == nil {
		req.AcceptanceCriteria = json.RawMessage(`[]`)
	}

	if err := h.collector.CreateConfig(r.Context(), &req); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create config")
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: req})
}

// UpdateConfig updates an evidence collection configuration.
// PUT /api/v1/monitoring/configs/{id}
func (h *MonitoringHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	configID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid config ID")
		return
	}

	var req service.CollectionConfig
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	req.ID = configID
	req.OrganizationID = orgID
	if req.APIConfig == nil {
		req.APIConfig = json.RawMessage(`{}`)
	}
	if req.FileConfig == nil {
		req.FileConfig = json.RawMessage(`{}`)
	}
	if req.ScriptConfig == nil {
		req.ScriptConfig = json.RawMessage(`{}`)
	}
	if req.WebhookConfig == nil {
		req.WebhookConfig = json.RawMessage(`{}`)
	}
	if req.AcceptanceCriteria == nil {
		req.AcceptanceCriteria = json.RawMessage(`[]`)
	}

	if err := h.collector.UpdateConfig(r.Context(), &req); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update config")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]interface{}{"id": configID, "message": "Config updated"},
	})
}

// RunNow triggers an immediate evidence collection for a config.
// POST /api/v1/monitoring/configs/{id}/run-now
func (h *MonitoringHandler) RunNow(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	configID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid config ID")
		return
	}

	// Verify config exists and belongs to org
	cfg, err := h.collector.GetConfig(r.Context(), orgID, configID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Config not found")
		return
	}

	// Execute collection asynchronously by enqueuing
	if err := h.collector.EnqueueCollection(r.Context(), *cfg); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to enqueue collection")
		return
	}

	writeJSON(w, http.StatusAccepted, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"config_id": configID,
			"message":   "Evidence collection enqueued",
		},
	})
}

// GetConfigHistory returns paginated collection run history for a config.
// GET /api/v1/monitoring/configs/{id}/history
func (h *MonitoringHandler) GetConfigHistory(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	configID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid config ID")
		return
	}

	params := parsePagination(r)
	runs, total, err := h.collector.ListRuns(r.Context(), orgID, configID, params.PageSize, params.Offset())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list run history")
		return
	}

	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize != 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: runs,
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
// COMPLIANCE MONITOR ENDPOINTS
// ============================================================

// ListMonitors returns paginated compliance monitors.
// GET /api/v1/monitoring/monitors
func (h *MonitoringHandler) ListMonitors(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	params := parsePagination(r)

	monitors, total, err := h.monitor.ListMonitors(r.Context(), orgID, params.PageSize, params.Offset())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list monitors")
		return
	}

	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize != 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: monitors,
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

// CreateMonitor creates a new compliance monitor.
// POST /api/v1/monitoring/monitors
func (h *MonitoringHandler) CreateMonitor(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var req service.MonitorConfig
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELD", "Name is required")
		return
	}
	if req.MonitorType == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELD", "Monitor type is required")
		return
	}

	req.OrganizationID = orgID
	if req.Conditions == nil {
		req.Conditions = json.RawMessage(`{}`)
	}

	if err := h.monitor.CreateMonitor(r.Context(), &req); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create monitor")
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: req})
}

// UpdateMonitor updates a compliance monitor.
// PUT /api/v1/monitoring/monitors/{id}
func (h *MonitoringHandler) UpdateMonitor(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	monitorID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid monitor ID")
		return
	}

	var req service.MonitorConfig
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	req.ID = monitorID
	req.OrganizationID = orgID
	if req.Conditions == nil {
		req.Conditions = json.RawMessage(`{}`)
	}

	if err := h.monitor.UpdateMonitor(r.Context(), &req); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update monitor")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]interface{}{"id": monitorID, "message": "Monitor updated"},
	})
}

// GetMonitorResults returns paginated check history for a monitor.
// GET /api/v1/monitoring/monitors/{id}/results
func (h *MonitoringHandler) GetMonitorResults(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	monitorID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid monitor ID")
		return
	}

	params := parsePagination(r)
	results, total, err := h.monitor.ListMonitorResults(r.Context(), orgID, monitorID, params.PageSize, params.Offset())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list results")
		return
	}

	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize != 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: results,
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
// DRIFT EVENT ENDPOINTS
// ============================================================

// ListDriftEvents returns paginated active drift events.
// GET /api/v1/monitoring/drift
func (h *MonitoringHandler) ListDriftEvents(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	params := parsePagination(r)

	events, total, err := h.drift.ListDriftEvents(r.Context(), orgID, params.PageSize, params.Offset())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list drift events")
		return
	}

	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize != 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: events,
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

// AcknowledgeDrift marks a drift event as acknowledged.
// PUT /api/v1/monitoring/drift/{id}/acknowledge
func (h *MonitoringHandler) AcknowledgeDrift(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	eventID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid drift event ID")
		return
	}

	if err := h.drift.AcknowledgeDrift(r.Context(), orgID, eventID, userID); err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"id":      eventID,
			"message": "Drift event acknowledged",
		},
	})
}

// ResolveDrift marks a drift event as resolved.
// PUT /api/v1/monitoring/drift/{id}/resolve
func (h *MonitoringHandler) ResolveDrift(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	eventID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid drift event ID")
		return
	}

	var req struct {
		Notes string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if err := h.drift.ResolveDrift(r.Context(), orgID, eventID, userID, req.Notes); err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"id":      eventID,
			"message": "Drift event resolved",
		},
	})
}

// ============================================================
// DASHBOARD ENDPOINT
// ============================================================

// GetMonitoringDashboard returns an aggregated monitoring dashboard.
// GET /api/v1/monitoring/dashboard
func (h *MonitoringHandler) GetMonitoringDashboard(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	// Get drift summary
	driftSummary, err := h.drift.GetDriftSummary(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to build dashboard")
		return
	}

	// Get active collection config stats
	var activeConfigs, failingConfigs int64
	h.collector.Pool().QueryRow(r.Context(), `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE last_collection_status = 'failed' OR last_collection_status = 'validation_failed')
		FROM evidence_collection_configs
		WHERE organization_id = $1 AND is_active = true`, orgID,
	).Scan(&activeConfigs, &failingConfigs)

	// Get active monitor stats
	var activeMonitors, failingMonitors int64
	h.monitor.Pool().QueryRow(r.Context(), `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE last_check_status = 'failing')
		FROM compliance_monitors
		WHERE organization_id = $1 AND is_active = true`, orgID,
	).Scan(&activeMonitors, &failingMonitors)

	// Get recent collection runs
	var recentRunsSuccess, recentRunsFailed int64
	h.collector.Pool().QueryRow(r.Context(), `
		SELECT
			COUNT(*) FILTER (WHERE status = 'success'),
			COUNT(*) FILTER (WHERE status IN ('failed','timeout','validation_failed'))
		FROM evidence_collection_runs
		WHERE organization_id = $1 AND created_at > NOW() - INTERVAL '24 hours'`, orgID,
	).Scan(&recentRunsSuccess, &recentRunsFailed)

	dashboard := map[string]interface{}{
		"drift_summary": driftSummary,
		"evidence_collection": map[string]interface{}{
			"active_configs":  activeConfigs,
			"failing_configs": failingConfigs,
			"runs_24h": map[string]interface{}{
				"success": recentRunsSuccess,
				"failed":  recentRunsFailed,
			},
		},
		"compliance_monitors": map[string]interface{}{
			"active_monitors":  activeMonitors,
			"failing_monitors": failingMonitors,
		},
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: dashboard})
}
