package handler

import (
	"context"
	"encoding/json"
	"io"
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
func NewMonitoringHandler(ec *service.EvidenceCollector, cm *service.ComplianceMonitor, dd *service.DriftDetector) *MonitoringHandler {
	return &MonitoringHandler{collector: ec, monitor: cm, drift: dd}
}

// ============================================================
// EVIDENCE COLLECTION
// ============================================================

// ListCollectionConfigs returns paginated evidence collection configurations.
// GET /monitoring/evidence
func (h *MonitoringHandler) ListCollectionConfigs(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	params := parsePagination(r)

	configs, total, err := h.collector.ListConfigs(r.Context(), orgID, params.PageSize, params.Offset())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list collection configs")
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

// CreateCollectionConfig creates a new evidence collection configuration.
// POST /monitoring/evidence
func (h *MonitoringHandler) CreateCollectionConfig(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var req service.CollectionConfig
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	req.OrganizationID = orgID

	if err := h.collector.CreateConfig(r.Context(), &req); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create collection config")
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: req})
}

// UpdateCollectionConfig updates an evidence collection configuration.
// PUT /monitoring/evidence/{id}
func (h *MonitoringHandler) UpdateCollectionConfig(w http.ResponseWriter, r *http.Request) {
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

	if err := h.collector.UpdateConfig(r.Context(), &req); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update collection config")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: req})
}

// RunCollectionNow triggers an immediate evidence collection for a configuration.
// POST /monitoring/evidence/{id}/run
func (h *MonitoringHandler) RunCollectionNow(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	configID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid config ID")
		return
	}

	go h.collector.ExecuteCollection(context.Background(), configID, orgID)

	writeJSON(w, http.StatusAccepted, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"config_id": configID,
			"message":   "Evidence collection started",
		},
	})
}

// ListCollectionRuns returns paginated collection run history for a configuration.
// GET /monitoring/evidence/{id}/runs
func (h *MonitoringHandler) ListCollectionRuns(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	configID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid config ID")
		return
	}

	params := parsePagination(r)
	runs, total, err := h.collector.ListRuns(r.Context(), orgID, configID, params.PageSize, params.Offset())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list collection runs")
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
// WEBHOOKS
// ============================================================

// HandleWebhook processes incoming webhook payloads for evidence collection.
// POST /monitoring/webhooks/{id}
func (h *MonitoringHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	configID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid config ID")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "READ_ERROR", "Failed to read request body")
		return
	}
	defer r.Body.Close()

	// Look up the config to get the webhook secret
	cfg, err := h.collector.GetConfig(r.Context(), orgID, configID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Config not found")
		return
	}

	var webhookCfg service.WebhookConfigPayload
	if err := json.Unmarshal(cfg.WebhookConfig, &webhookCfg); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to parse webhook config")
		return
	}

	signature := r.Header.Get("X-Webhook-Signature")
	if !h.collector.ValidateWebhookSignature(body, signature, webhookCfg.Secret) {
		writeError(w, http.StatusUnauthorized, "INVALID_SIGNATURE", "Webhook signature validation failed")
		return
	}

	if err := h.collector.ProcessWebhookPayload(r.Context(), configID, orgID, json.RawMessage(body)); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to process webhook payload")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: map[string]interface{}{
		"message": "Webhook processed",
	}})
}

// ============================================================
// COMPLIANCE MONITORS
// ============================================================

// ListMonitors returns paginated compliance monitors.
// GET /monitoring/monitors
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
// POST /monitoring/monitors
func (h *MonitoringHandler) CreateMonitor(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var req service.MonitorConfig
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	req.OrganizationID = orgID

	if err := h.monitor.CreateMonitor(r.Context(), &req); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create monitor")
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: req})
}

// UpdateMonitor updates a compliance monitor.
// PUT /monitoring/monitors/{id}
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

	if err := h.monitor.UpdateMonitor(r.Context(), &req); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update monitor")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: req})
}

// GetMonitorResults returns paginated check results for a compliance monitor.
// GET /monitoring/monitors/{id}/results
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
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list monitor results")
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
// DRIFT EVENTS
// ============================================================

// ListDriftEvents returns paginated drift events.
// GET /monitoring/drift
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
// POST /monitoring/drift/{id}/acknowledge
func (h *MonitoringHandler) AcknowledgeDrift(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	eventID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid drift event ID")
		return
	}

	if err := h.drift.AcknowledgeDrift(r.Context(), orgID, eventID, userID); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to acknowledge drift event")
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
// POST /monitoring/drift/{id}/resolve
func (h *MonitoringHandler) ResolveDrift(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	eventID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid drift event ID")
		return
	}

	var req struct {
		ResolutionNotes string `json:"resolution_notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if err := h.drift.ResolveDrift(r.Context(), orgID, eventID, userID, req.ResolutionNotes); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to resolve drift event")
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
// DASHBOARD
// ============================================================

// GetDashboard returns an aggregated monitoring dashboard with drift summary,
// evidence collection stats, and compliance monitor stats.
// GET /monitoring/dashboard
func (h *MonitoringHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orgID := middleware.GetOrgID(ctx)

	// Get drift summary
	driftSummary, err := h.drift.GetDriftSummary(ctx, orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to build dashboard")
		return
	}

	// Get evidence collection config stats
	var totalConfigs, activeConfigs, failingConfigs int64
	h.collector.Pool().QueryRow(ctx,
		`SELECT COUNT(*), COUNT(*) FILTER (WHERE is_active), COUNT(*) FILTER (WHERE consecutive_failures > 0) FROM evidence_collection_configs WHERE organization_id = $1`,
		orgID,
	).Scan(&totalConfigs, &activeConfigs, &failingConfigs)

	// Get compliance monitor stats
	var totalMonitors, passingMonitors, failingMonitors int64
	h.monitor.Pool().QueryRow(ctx,
		`SELECT COUNT(*), COUNT(*) FILTER (WHERE last_check_status = 'passing'), COUNT(*) FILTER (WHERE last_check_status = 'failing') FROM compliance_monitors WHERE organization_id = $1`,
		orgID,
	).Scan(&totalMonitors, &passingMonitors, &failingMonitors)

	dashboard := map[string]interface{}{
		"drift_summary": driftSummary,
		"evidence_collection": map[string]interface{}{
			"total_configs":   totalConfigs,
			"active_configs":  activeConfigs,
			"failing_configs": failingConfigs,
		},
		"compliance_monitors": map[string]interface{}{
			"total_monitors":   totalMonitors,
			"passing_monitors": passingMonitors,
			"failing_monitors": failingMonitors,
		},
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: dashboard})
}
